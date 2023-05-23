package slogutils

import (
	"bytes"
	"context"
	"encoding/json"
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
		if !strings.Contains(actual[i], colorPrefix[i]) {
			t.Errorf("expected %q, got %q", colorPrefix[i], actual[i])
		} else {
			var actualObj, expectedObj map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(actual[i], colorPrefix[i])), &actualObj); err != nil {
				t.Fatalf("failed to unmarshal actual %q: %s", actual[i], err)
			}
			if err := json.Unmarshal([]byte(expected[i]), &expectedObj); err != nil {
				t.Fatalf("failed to unmarshal expected %q: %s", expected[i], err)
			}
			delete(actualObj, "time")
			if !jsonEqual(actualObj, expectedObj) {
				t.Errorf("expected %q, got %q", expected[i], actual[i])
			}
		}
	}
	t.Log(result)
}

func TestMiddleware__WithRecordTransformer(t *testing.T) {
	color.NoColor = true

	buf := new(bytes.Buffer)
	middleware := NewMiddleware(
		slog.NewJSONHandler,
		MiddlewareOptions{
			RecordTransformerFuncs: []RecordTransformerFunc{
				DefaultAttrs(slog.String("log_category", "general")),
				DropAttrs("secrets"),
			},
			Writer: buf,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		},
	)
	logger := slog.New(middleware)
	ctx := With(context.Background(), slog.Int64("request_id", 12))
	logger.WarnCtx(ctx, "foo")
	logger.ErrorCtx(ctx, "bar", "secrets", "HIDDEN_VALUE")
	logger.DebugCtx(ctx, "baz")
	logger.InfoCtx(ctx, "buzz", slog.String("log_category", "special"))
	logger.WarnCtx(ctx, "buzz")
	result := buf.String()
	expected := []string{
		`{"level":"WARN","msg":"foo","request_id":12, "log_category": "general"}`,
		`{"level":"ERROR","msg":"bar","request_id":12, "log_category": "general"}`,
		`{"level":"INFO","msg":"buzz","request_id":12, "log_category": "special"}`,
		`{"level":"WARN","msg":"buzz","request_id":12, "log_category": "general"}`,
	}
	actual := strings.Split(result, "\n")
	if len(expected) != len(actual)-1 {
		t.Fatalf("expected %d lines, got %d lines", len(expected), len(actual))
	}

	for i := range expected {
		var actualObj, expectedObj map[string]interface{}
		if err := json.Unmarshal([]byte(actual[i]), &actualObj); err != nil {
			t.Fatalf("failed to unmarshal actual %q: %s", actual[i], err)
		}
		if err := json.Unmarshal([]byte(expected[i]), &expectedObj); err != nil {
			t.Fatalf("failed to unmarshal expected %q: %s", expected[i], err)
		}
		delete(actualObj, "time")
		if !jsonEqual(actualObj, expectedObj) {
			t.Errorf("expected %q, got %q", expected[i], actual[i])
		}
	}
	t.Log(result)
}

func TestMiddleware__LoggerWith(t *testing.T) {
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
	logger = logger.With(slog.Int64("request_id", 12))
	logger.Warn("foo")
	logger.Error("bar")
	logger.Debug("baz")
	logger.Warn("buzz")
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
		if !strings.Contains(actual[i], colorPrefix[i]) {
			t.Errorf("expected %q, got %q", colorPrefix[i], actual[i])
		} else {
			var actualObj, expectedObj map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(actual[i], colorPrefix[i])), &actualObj); err != nil {
				t.Fatalf("failed to unmarshal actual %q: %s", actual[i], err)
			}
			if err := json.Unmarshal([]byte(expected[i]), &expectedObj); err != nil {
				t.Fatalf("failed to unmarshal expected %q: %s", expected[i], err)
			}
			delete(actualObj, "time")
			if !jsonEqual(actualObj, expectedObj) {
				t.Errorf("expected %q, got %q", expected[i], actual[i])
			}
		}
	}
	t.Log(result)
}

func jsonEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !jsonEqualValue(v, b[k]) {
			return false
		}
	}
	return true
}

func jsonEqualValue(a, b interface{}) bool {
	switch a := a.(type) {
	case map[string]interface{}:
		b, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return jsonEqual(a, b)
	case []interface{}:
		b, ok := b.([]interface{})
		if !ok {
			return false
		}
		return jsonEqualSlice(a, b)
	default:
		return a == b
	}
}

func jsonEqualSlice(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !jsonEqualValue(a[i], b[i]) {
			return false
		}
	}
	return true
}
