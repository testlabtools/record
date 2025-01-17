package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

// This logging code is extracted from:
// https://github.com/samber/slog-multi/blob/main/multi.go.
//
// See also: https://github.com/samber/slog-multi

var _ slog.Handler = (*FanoutHandler)(nil)

type FanoutHandler struct {
	handlers []slog.Handler
}

// Fanout distributes records to multiple slog.Handler in parallel
func Fanout(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{
		handlers: handlers,
	}
}

// Implements slog.Handler
func (h *FanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.Handler
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := tryHandler(func() error {
				return h.handlers[i].Handle(ctx, r.Clone())
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// Implements slog.Handler
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var handlers []slog.Handler
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(slices.Clone(attrs)))
	}

	return Fanout(handlers...)
}

// Implements slog.Handler
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	var handlers []slog.Handler
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}

	return Fanout(handlers...)
}

func tryHandler(callback func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("unexpected error: %+v", r)
			}
		}
	}()

	err = callback()

	return
}
