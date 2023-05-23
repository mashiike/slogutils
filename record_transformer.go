package slogutils

import (
	"strings"

	"golang.org/x/exp/slog"
)

// RecordTransformerFunc is a function that transforms a slog.Record.
type RecordTransformerFunc func(slog.Record) slog.Record

// DefaultAttrs returns a RecordTransformerFunc that adds the given attributes to a slog.Record if they don't already exist.
func DefaultAttrs(args ...any) func(slog.Record) slog.Record {
	attrs := argsToAttrs(args)
	return func(r slog.Record) slog.Record {
		exits := make(map[string]bool, len(attrs))
		r.Attrs(func(a slog.Attr) bool {
			exits[a.Key] = true
			return true
		})
		notExits := make([]slog.Attr, 0, len(attrs))
		for _, a := range attrs {
			if !exits[a.Key] {
				notExits = append(notExits, a)
			}
		}
		r.AddAttrs(notExits...)
		return r
	}
}

// DropAttrs returns a RecordTransformerFunc that drops the given attributes from a slog.Record.
func DropAttrs(keys ...string) func(slog.Record) slog.Record {
	return func(r slog.Record) slog.Record {
		attrs := make([]slog.Attr, 0, len(keys))
		r.Attrs(func(a slog.Attr) bool {
			for _, key := range keys {
				if a.Key == key {
					return true
				}
			}
			attrs = append(attrs, a)
			return true
		})
		c := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		c.AddAttrs(attrs...)
		return c
	}
}

// RenameAttrs returns a RecordTransformerFunc that renames the given attributes from a slog.Record.
func RenameAttrs(m map[string]string) func(slog.Record) slog.Record {
	return func(r slog.Record) slog.Record {
		attrs := make([]slog.Attr, 0, len(m))
		r.Attrs(func(a slog.Attr) bool {
			if key, ok := m[a.Key]; ok {
				attrs = append(attrs, slog.Attr{Key: key, Value: a.Value})
				return true
			}
			attrs = append(attrs, a)
			return true
		})
		c := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		c.AddAttrs(attrs...)
		return c
	}
}

// UniqueAttrs returns a RecordTransformerFunc that removes duplicate attributes from a slog.Record.
func UniqueAttrs() func(slog.Record) slog.Record {
	return func(r slog.Record) slog.Record {
		attrMap := make(map[string]slog.Value)
		r.Attrs(func(a slog.Attr) bool {
			attrMap[a.Key] = a.Value
			return true
		})
		attrs := make([]slog.Attr, 0, len(attrMap))
		for k, v := range attrMap {
			attrs = append(attrs, slog.Attr{Key: k, Value: v})
		}
		c := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		c.AddAttrs(attrs...)
		return c
	}
}

// ConvertLegacyLevel returns a RecordTransformerFunc that converts legacy level to slog.Level.
// The legacy level is the first word in the message enclosed in square brackets.
// The levelMap maps the legacy level to slog.Level.
// If caseInsensitive is true, the legacy level is case insensitive.
// If the legacy level is not found in the levelMap, the slog.Record is not changed.
// If the slog.Record.Level is not slog.LevelInfo, the slog.Record is not changed.
//
// Example:
//
//	ConvertLegacyLevel(map[string]slog.Level{"debug": slog.LevelDebug}, true)
//	If the message is "[DEBUG] hello world", the slog.Record.Level is converted to slog.LevelDebug.
func ConvertLegacyLevel(levelMap map[string]slog.Level, caseInsensitive bool) func(slog.Record) slog.Record {
	if levelMap == nil {
		return func(r slog.Record) slog.Record { return r }
	}
	if caseInsensitive {
		for k, v := range levelMap {
			levelMap[strings.ToLower(k)] = v
		}
	}
	return func(r slog.Record) slog.Record {
		if r.Level != slog.LevelInfo {
			return r
		}
		x := strings.IndexByte(r.Message, '[')
		if x < 0 {
			return r
		}
		y := strings.IndexByte(r.Message[x:], ']')
		if y < 0 {
			return r
		}
		legacyLevel := r.Message[x+1 : x+y]
		msg := strings.TrimSpace(r.Message[x+y+1:])
		if caseInsensitive {
			legacyLevel = strings.ToLower(legacyLevel)
		}
		if level, ok := levelMap[legacyLevel]; ok {
			c := r.Clone()
			c.Level = level
			c.Message = msg
			return c
		}
		return r
	}
}
