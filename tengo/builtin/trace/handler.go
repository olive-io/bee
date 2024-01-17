// Copyright 2024 The bee Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/d5/tengo/v2"
)

// A traceHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type traceHandler struct {
	sync.RWMutex

	level    slog.Leveler
	handlers []slog.Handler
}

// newTraceHandler returns a traceHandler with the given level.
// All methods except Enabled delegate to h.
func newTraceHandler(level slog.Leveler, h slog.Handler) *traceHandler {
	// Optimization: avoid chains of LevelHandlers.
	handlers := []slog.Handler{h}
	if lh, ok := h.(*traceHandler); ok {
		handlers = lh.handlers
	}
	return &traceHandler{level: level, handlers: handlers}
}

func (h *traceHandler) AddHandler(handler slog.Handler) {
	h.Lock()
	defer h.Unlock()
	h.handlers = append(h.handlers, handler)
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *traceHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	h.RLock()
	defer h.RUnlock()
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs implements Handler.WithAttrs.
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.RLock()
	defer h.RUnlock()
	handlers := make([]slog.Handler, 0)
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}
	return &traceHandler{level: h.level, handlers: handlers}
}

// WithGroup implements Handler.WithGroup.
func (h *traceHandler) WithGroup(name string) slog.Handler {
	h.RLock()
	defer h.RUnlock()
	handlers := make([]slog.Handler, 0)
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}
	return &traceHandler{level: h.level, handlers: handlers}
}

// AddHandler adds a new slog.Handler to ImportModule
func (m *ImportModule) AddHandler() tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		numArgs := len(args)
		if numArgs <= 2 {
			return nil, errors.Wrap(tengo.ErrWrongNumArguments, "must greater than 2")
		}

		out, ok := args[0].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "writer",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		writer, err := os.OpenFile(out.Value, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 755)
		if err != nil {
			return nil, err
		}

		levelStr, ok := args[1].(*tengo.String)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "level",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}
		level, ok := parseLevel(levelStr.Value)
		if !ok {
			level = defaultLevel
		}

		attrs := make([]slog.Attr, 0)
		if len(args) > 2 {
			for _, arg := range args[2:] {
				if attr, ok := arg.(*traceField); ok {
					attrs = append(attrs, attr.Value)
				}
			}
		}

		options := &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				return a
			},
		}

		handler := slog.NewJSONHandler(writer, options)
		m.handler.AddHandler(handler.WithAttrs(attrs))

		return tengo.UndefinedValue, nil
	}
}