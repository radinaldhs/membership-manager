package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type BuiltInKeys struct {
	TimeKey    string
	LevelKey   string
	MessageKey string
	SourceKey  string
}

const (
	TimeKey    = slog.TimeKey
	LevelKey   = slog.LevelKey
	MessageKey = slog.MessageKey
	SourceKey  = slog.SourceKey
)

type ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr

type AddSourceFunc func(lvl Level) bool

type AttrsFromCtxFunc func(ctx context.Context) []slog.Attr

type Handler struct {
	w            io.Writer
	addSource    AddSourceFunc
	lvl          Level
	replaceAttr  ReplaceAttrFunc
	attrsFromCtx AttrsFromCtxFunc
	keys         BuiltInKeys
	exitOnFatal  bool
	h            *slog.JSONHandler
}

func createJsonSlogHandler(w io.Writer, lvl slog.Level, keys BuiltInKeys, replace ReplaceAttrFunc) *slog.JSONHandler {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       lvl,
		ReplaceAttr: replaceAttr(keys, replaceAttr(keys, replace)),
	})

	return h
}

func NewHandler(w io.Writer, opts ...HandlerOption) *Handler {
	if w == nil {
		w = os.Stderr
	}

	jh := Handler{
		w:           w,
		addSource:   func(lvl Level) bool { return false },
		lvl:         LevelDebug,
		exitOnFatal: true,
		h:           createJsonSlogHandler(w, LevelDebug, BuiltInKeys{}, nil),
	}

	for _, opt := range opts {
		opt(&jh)
	}

	return &jh
}

func (jh *Handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return jh.h.Enabled(ctx, lvl)
}

func (jh *Handler) Handle(ctx context.Context, rec slog.Record) error {
	if !jh.addSource(rec.Level) {
		rec.PC = 0
	}

	if jh.attrsFromCtx != nil {
		attrs := jh.attrsFromCtx(ctx)
		rec.AddAttrs(attrs...)
	}

	if jh.exitOnFatal {
		if rec.Level == LevelFatal {
			defer func() { os.Exit(1) }()
		}
	}

	return jh.h.Handle(ctx, rec)
}

func (jh *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sjh := jh.h.WithAttrs(attrs).(*slog.JSONHandler)
	return jh.copy(sjh)
}

func (jh *Handler) WithGroup(name string) slog.Handler {
	sjh := jh.h.WithGroup(name).(*slog.JSONHandler)
	return jh.copy(sjh)
}

func (jh *Handler) copy(sjh *slog.JSONHandler) *Handler {
	newJh := Handler{
		w:            jh.w,
		addSource:    jh.addSource,
		lvl:          jh.lvl,
		replaceAttr:  jh.replaceAttr,
		attrsFromCtx: jh.attrsFromCtx,
		keys:         jh.keys,
		exitOnFatal:  jh.exitOnFatal,
		h:            sjh,
	}

	return &newJh
}

func replaceAttr(keys BuiltInKeys, replace ReplaceAttrFunc) ReplaceAttrFunc {
	return func(groups []string, a slog.Attr) slog.Attr {
		lvlKey := LevelKey

		// Built-in keys replacement
		switch a.Key {
		case TimeKey:
			if keys.TimeKey != "" {
				a.Key = keys.TimeKey
			}

		case LevelKey:
			if keys.LevelKey != "" {
				a.Key = keys.LevelKey
				lvlKey = keys.LevelKey
			}

		case MessageKey:
			if keys.MessageKey != "" {
				a.Key = keys.MessageKey
			}

		case SourceKey:
			if keys.SourceKey != "" {
				a.Key = keys.SourceKey
			}
		}

		// Log level string for Trace, Critical, and Fatal level
		if a.Key == lvlKey {
			switch a.Value.String() {
			case "DEBUG+1":
				a.Value = slog.StringValue("TRACE")

			case "ERROR+1":
				a.Value = slog.StringValue("CRITICAL")

			case "ERROR+2":
				a.Value = slog.StringValue("FATAL")
			}
		}

		if replace != nil {
			return replace(groups, a)
		}

		return a
	}
}
