package logging

type HandlerOption func(h *Handler)

func WithAddSource(addFunc AddSourceFunc) HandlerOption {
	return func(h *Handler) {
		h.addSource = addFunc
	}
}

func WithLevel(lvl Level) HandlerOption {
	return func(h *Handler) {
		h.lvl = lvl
		h.h = createJsonSlogHandler(h.w, lvl, h.keys, h.replaceAttr)
	}
}

func WithReplaceAttr(replace ReplaceAttrFunc) HandlerOption {
	return func(h *Handler) {
		h.replaceAttr = replace
		h.h = createJsonSlogHandler(h.w, h.lvl, h.keys, replace)
	}
}

func WithAttrsFromCtx(fromCtx AttrsFromCtxFunc) HandlerOption {
	return func(h *Handler) {
		h.attrsFromCtx = fromCtx
	}
}

func WithBuiltInKeys(keys BuiltInKeys) HandlerOption {
	return func(h *Handler) {
		h.keys = keys
		h.h = createJsonSlogHandler(h.w, h.lvl, keys, h.replaceAttr)
	}
}

func WithExitOnFatal(b bool) HandlerOption {
	return func(h *Handler) {
		h.exitOnFatal = b
	}
}
