package logger

import (
	"context"
	"log/slog"
	"sync"
)

type replayHandler struct {
	buf *replayBuffer
	ops []func(slog.Handler) slog.Handler
}

type replayBuffer struct {
	mu      sync.Mutex
	entries []replayEntry
}

type replayEntry struct {
	record slog.Record
	ops    []func(slog.Handler) slog.Handler
}

func newReplayHandler() *replayHandler {
	return &replayHandler{buf: &replayBuffer{}}
}

func (h *replayHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *replayHandler) Handle(_ context.Context, r slog.Record) error {
	h.buf.mu.Lock()
	defer h.buf.mu.Unlock()
	h.buf.entries = append(h.buf.entries, replayEntry{record: r.Clone(), ops: h.ops})
	return nil
}

func (h *replayHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.with(func(t slog.Handler) slog.Handler { return t.WithAttrs(attrs) })
}

func (h *replayHandler) WithGroup(name string) slog.Handler {
	return h.with(func(t slog.Handler) slog.Handler { return t.WithGroup(name) })
}

func (h *replayHandler) with(op func(slog.Handler) slog.Handler) *replayHandler {
	ops := make([]func(slog.Handler) slog.Handler, len(h.ops), len(h.ops)+1)
	copy(ops, h.ops)
	return &replayHandler{buf: h.buf, ops: append(ops, op)}
}

func (h *replayHandler) replay(target slog.Handler) {
	h.buf.mu.Lock()
	entries := h.buf.entries
	h.buf.entries = nil
	h.buf.mu.Unlock()

	ctx := context.Background()
	for _, e := range entries {
		t := target
		for _, op := range e.ops {
			t = op(t)
		}
		if t.Enabled(ctx, e.record.Level) {
			_ = t.Handle(ctx, e.record)
		}
	}
}
