package slogutils

import (
	"context"
	"io"
	"testing"

	"github.com/fatih/color"
	"golang.org/x/exp/slog"
)

var levels []slog.Level
var messages []string

func init() {
	levels = []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}
	messages = []string{
		"foo",
		"bar",
		"baz",
		"buzz",
	}
}

func BenchmarkSlogDefault(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log(context.Background(), levels[i%len(messages)], messages[i%len(messages)])
	}
}

func BenchmarkMiddleware(b *testing.B) {
	logger := slog.New(
		NewMiddleware(
			slog.NewJSONHandler,
			MiddlewareOptions{
				ModifierFuncs: map[slog.Level]ModifierFunc{
					slog.LevelDebug: Color(color.FgBlack),
					slog.LevelInfo:  Color(color.FgBlue),
					slog.LevelWarn:  Color(color.FgYellow),
					slog.LevelError: Color(color.FgRed, color.BgBlack),
				},
				Writer: io.Discard,
				HandlerOptions: &slog.HandlerOptions{
					Level: slog.LevelWarn,
				},
			},
		),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log(context.Background(), levels[i%len(messages)], messages[i%len(messages)])
	}
}
