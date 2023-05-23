package slogutils

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/exp/slog"
)

type ModifierFunc func([]byte) []byte

func Color(attr ...color.Attribute) ModifierFunc {
	c := color.New(attr...)
	buf := &bytes.Buffer{}
	return func(b []byte) []byte {
		buf.Reset()
		c.Fprint(buf, string(b))
		return buf.Bytes()
	}
}

type modifierWriter struct {
	f ModifierFunc
	w io.Writer
	sync.Mutex
}

func (w *modifierWriter) Write(b []byte) (int, error) {
	if w.f == nil {
		return w.w.Write(b)
	}
	return w.w.Write(w.f(b))
}

func (w *modifierWriter) SetModifierFunc(f ModifierFunc) {
	w.f = f
}

type MiddlewareOptions struct {
	ModifierFuncs          map[slog.Level]ModifierFunc
	RecordTransformerFuncs []RecordTransformerFunc
	Writer                 io.Writer
	HandlerOptions         *slog.HandlerOptions
}

type Middleware struct {
	modifierFuncs          map[slog.Level]ModifierFunc
	recordTransformerFuncs []RecordTransformerFunc
	h                      slog.Handler
	w                      *modifierWriter
}

func NewMiddleware[H slog.Handler](f func(io.Writer, *slog.HandlerOptions) H, opts MiddlewareOptions) *Middleware {
	if opts.ModifierFuncs == nil {
		opts.ModifierFuncs = map[slog.Level]ModifierFunc{}
	}
	w := &modifierWriter{w: opts.Writer}
	h := f(w, opts.HandlerOptions)
	return &Middleware{
		modifierFuncs:          opts.ModifierFuncs,
		recordTransformerFuncs: opts.RecordTransformerFuncs,
		h:                      h,
		w:                      w,
	}
}

// Handle implements slog.Handler.
func (m *Middleware) Handle(ctx context.Context, record slog.Record) error {
	h := m.h
	if attrs, ok := attrsFromContext(ctx); ok {
		h = h.WithAttrs(attrs)
	}
	for _, f := range m.recordTransformerFuncs {
		record = f(record)
	}
	m.w.Lock()
	defer m.w.Unlock()
	m.w.SetModifierFunc(m.modifierFuncs[record.Level])
	return h.Handle(ctx, record)
}

// Clone returns a new Middleware with the same Handler and modifierFuncs.
func (m *Middleware) Clone() *Middleware {
	modifierFuncs := make(map[slog.Level]ModifierFunc, len(m.modifierFuncs))
	for k, v := range m.modifierFuncs {
		modifierFuncs[k] = v
	}
	recordTransformerFuncs := make([]RecordTransformerFunc, len(m.recordTransformerFuncs))
	copy(recordTransformerFuncs, m.recordTransformerFuncs)
	return &Middleware{
		modifierFuncs:          modifierFuncs,
		recordTransformerFuncs: recordTransformerFuncs,
		h:                      m.h,
		w:                      m.w,
	}
}

func (m *Middleware) Enabled(ctx context.Context, l slog.Level) bool {
	return m.h.Enabled(ctx, l)
}

func (m *Middleware) WithAttrs(as []slog.Attr) slog.Handler {
	c := m.Clone()
	c.h = c.h.WithAttrs(as)
	return c
}

func (m *Middleware) WithGroup(name string) slog.Handler {
	c := m.Clone()
	c.h = c.h.WithGroup(name)
	return c
}

type contextKeyType struct{}

var contextKey contextKeyType

func With(ctx context.Context, args ...any) context.Context {
	defualtAttr, ok := attrsFromContext(ctx)
	var attrs []slog.Attr
	if ok {
		attrs = append(attrs, defualtAttr...)
	}
	attrs = append(attrs, argsToAttrs(args)...)
	return context.WithValue(ctx, contextKey, attrs)
}

func attrsFromContext(ctx context.Context) ([]slog.Attr, bool) {
	m, ok := ctx.Value(contextKey).([]slog.Attr)
	return m, ok
}

func argsToAttrs(args []any) []slog.Attr {
	var attrs []slog.Attr
	for len(args) > 0 {
		switch v := args[0].(type) {
		case slog.Attr:
			attrs = append(attrs, v)
			args = args[1:]
		case string:
			if len(args) < 2 {
				attrs = append(attrs, slog.Any("!BADKEY", v))
				args = args[1:]
			} else {
				attrs = append(attrs, slog.Any(v, args[1]))
				args = args[2:]
			}
		default:
			attrs = append(attrs, slog.Any("!BADKEY", v))
		}
	}
	return attrs
}
