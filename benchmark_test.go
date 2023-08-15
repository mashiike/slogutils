package slogutils

import (
	"context"
	"io"
	"log"
	"log/slog"
	"testing"

	"github.com/fatih/color"
)

var levels []slog.Level
var messages []string
var messagesWithLevel []string

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
	messagesWithLevel = []string{
		"[DEBUG] foo",
		"[INFO] bar",
		"[WARN] baz",
		"[ERROR] buzz",
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

func BenchmarkLogOutput(b *testing.B) {
	logger := log.New(io.Discard, "", log.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Output(0, messages[i%len(messages)])
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

func BenchmarkMiddlewareWithRecordTrnasformer(b *testing.B) {
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
				RecordTransformerFuncs: []RecordTransformerFunc{
					DefaultAttrs("hoge", "fuga"),
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

func BenchmarkLogOutputWithMiddleware(b *testing.B) {
	logger := slog.NewLogLogger(
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
		slog.LevelInfo,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Output(0, messages[i%len(messages)])
	}
}

func BenchmarkLogOutputWithRecordTransformer(b *testing.B) {
	logger := slog.NewLogLogger(
		NewMiddleware(
			slog.NewJSONHandler,
			MiddlewareOptions{
				ModifierFuncs: map[slog.Level]ModifierFunc{
					slog.LevelDebug: Color(color.FgBlack),
					slog.LevelInfo:  Color(color.FgBlue),
					slog.LevelWarn:  Color(color.FgYellow),
					slog.LevelError: Color(color.FgRed, color.BgBlack),
				},
				RecordTransformerFuncs: []RecordTransformerFunc{
					ConvertLegacyLevel(
						map[string]slog.Level{
							"debug": slog.LevelDebug,
							"info":  slog.LevelInfo,
							"warn":  slog.LevelWarn,
							"error": slog.LevelError,
						},
						false),
				},
				Writer: io.Discard,
				HandlerOptions: &slog.HandlerOptions{
					Level: slog.LevelWarn,
				},
			},
		),
		slog.LevelInfo,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Output(0, messages[i%len(messages)])
	}
}
