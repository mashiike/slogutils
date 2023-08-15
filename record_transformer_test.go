package slogutils

import (
	"log/slog"
	"testing"
	"time"
)

func TestDefaultAttrs(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "TestDefaultAttrs", 0)
	r.AddAttrs(slog.String("foo", "foo"))
	r = DefaultAttrs("foo", "bar", "baz", "buzz")(r)
	exits := make(map[string]bool, 2)
	r.Attrs(func(a slog.Attr) bool {
		exits[a.Key] = true
		return true
	})
	if !exits["foo"] || !exits["baz"] {
		t.Errorf("expected attrs to contain foo and baz")
	}
}

func TestDropAttrs(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "TestDropAttrs", 0)
	r.AddAttrs(slog.String("foo", "foo"), slog.String("bar", "bar"))
	r = DropAttrs("foo")(r)
	exits := make(map[string]bool, 2)
	r.Attrs(func(a slog.Attr) bool {
		exits[a.Key] = true
		return true
	})
	if exits["foo"] || !exits["bar"] {
		t.Errorf("expected attrs to contain bar and not foo")
	}
}

func TestRenameAttrs(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "TestRenameAttrs", 0)
	r.AddAttrs(slog.String("foo", "foo"), slog.String("bar", "bar"))
	r = RenameAttrs(map[string]string{"foo": "baz"})(r)
	exits := make(map[string]bool, 2)
	r.Attrs(func(a slog.Attr) bool {
		exits[a.Key] = true
		return true
	})
	if exits["foo"] || !exits["baz"] {
		t.Errorf("expected attrs to contain baz and not foo")
	}
}

func TestUniqueAttrs(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "TestUniqueAttrs", 0)
	r.AddAttrs(slog.String("foo", "foo"), slog.String("foo", "bar"))
	r = UniqueAttrs()(r)
	actual := make(map[string]slog.Value)
	found := make(map[string]int)
	r.Attrs(func(a slog.Attr) bool {
		actual[a.Key] = a.Value
		found[a.Key]++
		return true
	})
	if len(actual) != 1 {
		t.Errorf("expected attrs to contain only one foo")
	}
	if found["foo"] != 1 {
		t.Errorf("expected attrs to contain only one foo")
	}
	if actual["foo"].String() != "bar" {
		t.Errorf("expected foo to be bar")
	}
	t.Log(actual)
}

func TestConvertLegacyLevel(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "TestConvertLegacyLeveler", 0)
	transformer := ConvertLegacyLevel(map[string]slog.Level{
		"debug":  slog.LevelDebug,
		"info":   slog.LevelInfo,
		"notice": slog.Level(2),
		"warn":   slog.LevelWarn,
		"error":  slog.LevelError,
	}, true)
	cases := []struct {
		name     string
		level    slog.Level
		msg      string
		expected slog.Level
	}{
		{"no change level is not info", slog.LevelDebug, "foo", slog.LevelDebug},
		{"no change level message not found", slog.LevelInfo, "foo", slog.LevelInfo},
		{"change level to debug", slog.LevelInfo, "[debug] foo", slog.LevelDebug},
		{"change level to notice", slog.LevelInfo, "[notice] foo", slog.Level(2)},
		{"change level to warn", slog.LevelInfo, "[warn] foo", slog.LevelWarn},
		{"change level to error", slog.LevelInfo, "[error] foo", slog.LevelError},
		{"not change level danger is not found in level map", slog.LevelInfo, "[danger] foo", slog.LevelInfo},
		{"change uppercase level to debug", slog.LevelInfo, "[DEBUG] foo", slog.LevelDebug},
		{"change uppercase level to notice", slog.LevelInfo, "[NOTICE] foo", slog.Level(2)},
		{"change uppercase level to warn", slog.LevelInfo, "[WARN] foo", slog.LevelWarn},
		{"change uppercase level to error", slog.LevelInfo, "[ERROR] foo", slog.LevelError},
		{"not change uppercase level danger is not found in level map", slog.LevelInfo, "[DANGER] foo", slog.LevelInfo},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r.Message = c.msg
			r.Level = c.level
			actual := transformer(r)
			if actual.Level != c.expected {
				t.Errorf("expected level %v, got %v", c.expected, actual.Level)
			}
		})
	}
}
