package slogutils

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fatih/color"
	"golang.org/x/exp/slog"
)

func TestMiddleware__WithColor(t *testing.T) {
	color.NoColor = false

	buf := new(bytes.Buffer)
	middleware := NewMiddleware(
		slog.NewJSONHandler,
		MiddlewareOptions{
			ModifierFuncs: map[slog.Level]ModifierFunc{
				slog.LevelDebug: Color(color.FgBlack),
				slog.LevelInfo:  nil,
				slog.LevelWarn:  Color(color.FgYellow),
				slog.LevelError: Color(color.FgRed, color.Bold),
			},
			Writer: buf,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelWarn,
			},
		},
	)
	logger := slog.New(middleware)
	ctx := With(context.Background(), slog.Int64("request_id", 12))
	logger.WarnCtx(ctx, "foo")
	logger.ErrorCtx(ctx, "bar")
	logger.DebugCtx(ctx, "baz")
	logger.WarnCtx(ctx, "buzz")
	result := buf.String()
	expected := []string{
		`{"level":"WARN","msg":"foo","request_id":12}`,
		`{"level":"ERROR","msg":"bar","request_id":12}`,
		`{"level":"WARN","msg":"buzz","request_id":12}`,
	}
	colorPrefix := []string{
		"\x1b[33m",
		"\x1b[31;1m",
		"\x1b[33m",
	}
	actual := strings.Split(result, "\n\x1b[0m")
	if len(expected) != len(actual)-1 {
		t.Fatalf("expected %d lines, got %d lines", len(expected), len(actual))
	}
	for i := range expected {
		// ignore time key
		if !strings.Contains(actual[i], expected[i][1:]) {
			t.Errorf("expected %q, got %q", expected[i], actual[i])
		}
		if !strings.Contains(actual[i], colorPrefix[i]) {
			t.Errorf("expected %q, got %q", colorPrefix[i], actual[i])
		}
	}
	t.Log(result)
}
