package slogutils

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/fatih/color"
)

// ModifierFunc is a function that modifies a log line.
type ModifierFunc func([]byte) []byte

// Color returns a ModifierFunc that colors the log line.
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

// MiddlewareOptions are options for creating a Middleware.
type MiddlewareOptions struct {
	// ModifierFuncs is a map of log levels to ModifierFunc.
	ModifierFuncs map[slog.Level]ModifierFunc

	// RecordTransformerFuncs is a list of RecordTransformerFunc.
	RecordTransformerFuncs []RecordTransformerFunc

	// Writer is the writer to write to.
	Writer io.Writer

	// HandlerOptions are options for the handler.
	HandlerOptions *slog.HandlerOptions
}

// Middleware is a slog.Handler that modifies log lines.
type Middleware[H slog.Handler] struct {
	mu                     sync.RWMutex
	modifierFuncs          map[slog.Level]ModifierFunc
	recordTransformerFuncs []RecordTransformerFunc
	opts                   MiddlewareOptions
	h                      slog.Handler
	w                      *modifierWriter
	f                      func(io.Writer, *slog.HandlerOptions) H
}

func NewMiddleware[H slog.Handler](f func(io.Writer, *slog.HandlerOptions) H, opts MiddlewareOptions) *Middleware[H] {
	if opts.ModifierFuncs == nil {
		opts.ModifierFuncs = map[slog.Level]ModifierFunc{}
	}
	w := &modifierWriter{w: opts.Writer}
	h := f(w, opts.HandlerOptions)
	return &Middleware[H]{
		modifierFuncs:          opts.ModifierFuncs,
		recordTransformerFuncs: opts.RecordTransformerFuncs,
		h:                      h,
		w:                      w,
		f:                      f,
		opts:                   opts,
	}
}

func (m *Middleware[H]) SetMinLevel(l slog.Leveler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.opts.HandlerOptions.Level = l
	h := m.f(m.w, m.opts.HandlerOptions)
	m.h = h
}

// Handle implements slog.Handler.
func (m *Middleware[H]) Handle(ctx context.Context, record slog.Record) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.h
	if attrs, ok := attrsFromContext(ctx); ok {
		h = h.WithAttrs(attrs)
	}
	l := record.Level
	for _, f := range m.recordTransformerFuncs {
		record = f(record)
	}
	if l != record.Level {
		if !m.Enabled(ctx, record.Level) {
			return nil
		}
	}
	m.w.Lock()
	defer m.w.Unlock()
	m.w.SetModifierFunc(m.modifierFuncs[record.Level])
	return h.Handle(ctx, record)
}

// Clone returns a new Middleware with the same Handler and modifierFuncs.
func (m *Middleware[H]) Clone() *Middleware[H] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	modifierFuncs := make(map[slog.Level]ModifierFunc, len(m.modifierFuncs))
	for k, v := range m.modifierFuncs {
		modifierFuncs[k] = v
	}
	recordTransformerFuncs := make([]RecordTransformerFunc, len(m.recordTransformerFuncs))
	copy(recordTransformerFuncs, m.recordTransformerFuncs)
	return &Middleware[H]{
		modifierFuncs:          modifierFuncs,
		recordTransformerFuncs: recordTransformerFuncs,
		h:                      m.h,
		w:                      m.w,
	}
}

func (m *Middleware[H]) Enabled(ctx context.Context, l slog.Level) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.h.Enabled(ctx, l)
}

func (m *Middleware[H]) WithAttrs(as []slog.Attr) slog.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c := m.Clone()
	c.h = c.h.WithAttrs(as)
	return c
}

func (m *Middleware[H]) WithGroup(name string) slog.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
