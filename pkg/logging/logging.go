package logging

import (
	"context"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type TraceIDCtxKey string

const TraceIDKey TraceIDCtxKey = "trace-id"

type ErrorPriority uint

type Level = slog.Level

const (
	LevelDebug    Level = slog.LevelDebug
	LevelTrace    Level = slog.LevelDebug + 1
	LevelInfo     Level = slog.LevelInfo
	LevelWarn     Level = slog.LevelWarn
	LevelError    Level = slog.LevelError
	LevelCritical Level = slog.LevelError + 1
	LevelFatal    Level = slog.LevelError + 2
)

type Logger struct {
	h slog.Handler
}

func (l *Logger) log(ctx context.Context, lvl slog.Level, msg string, attrs ...slog.Attr) {
	// Skip if level is not enabled
	if !l.h.Enabled(ctx, lvl) {
		return
	}

	// Get caller source
	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	rec := slog.NewRecord(time.Now(), lvl, msg, pc)
	rec.AddAttrs(attrs...)

	err := l.h.Handle(ctx, rec)
	if err != nil {
		// Fallback to log.Println
		log.Println("logger error:", err.Error())
	}
}

func New(h slog.Handler) *Logger {
	return &Logger{
		h: h,
	}
}

func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	h := l.h.WithAttrs(attrs)
	return New(h)
}

func (l *Logger) WithGroup(name string) *Logger {
	h := l.h.WithGroup(name)
	return New(h)
}

func (l *Logger) Debug(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelDebug, msg, attrs...)
}

func (l *Logger) Trace(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelTrace, msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelInfo, msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelWarn, msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelError, msg, attrs...)
}

func (l *Logger) Critical(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelCritical, msg, attrs...)
}

func (l *Logger) Fatal(msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), LevelFatal, msg, attrs...)
}

func (l *Logger) Log(lvl Level, msg string, attrs ...slog.Attr) {
	l.log(context.TODO(), lvl, msg, attrs...)
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelDebug, msg, attrs...)
}

func (l *Logger) TraceCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelTrace, msg, attrs...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelInfo, msg, attrs...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelWarn, msg, attrs...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelError, msg, attrs...)
}

func (l *Logger) CriticalCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelCritical, msg, attrs...)
}

func (l *Logger) FatalCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, LevelFatal, msg, attrs...)
}

func (l *Logger) LogCtx(ctx context.Context, lvl Level, msg string, attrs ...slog.Attr) {
	l.log(ctx, lvl, msg, attrs...)
}

func CreateDefaultLogger(lvlStr string) *Logger {
	logLvl := LevelDebug
	lvlStr = strings.ToUpper(lvlStr)
	switch lvlStr {
	case "DEBUG":
		logLvl = LevelDebug
	case "TRACE":
		logLvl = LevelTrace
	case "INFO":
		logLvl = LevelInfo
	case "WARN":
		logLvl = LevelWarn
	case "ERROR":
		logLvl = LevelError
	case "CRITICAL":
		logLvl = LevelCritical
	case "FATAL":
		logLvl = LevelFatal
	}

	logh := NewHandler(os.Stderr,
		WithLevel(logLvl),
		WithAddSource(func(lvl Level) bool {
			switch lvl {
			case LevelDebug, LevelTrace,
				LevelError, LevelCritical,
				LevelFatal:
				return true
			}

			return false
		}),
		WithAttrsFromCtx(func(ctx context.Context) []slog.Attr {
			v := ctx.Value(TraceIDKey)
			if v != nil {
				traceId, ok := v.(string)
				if ok {
					attrs := []slog.Attr{slog.String("trace_id", traceId)}
					return attrs
				}
			}

			return nil
		}),
	)

	return New(logh)
}
