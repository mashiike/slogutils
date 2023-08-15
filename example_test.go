package slogutils_test

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/mashiike/slogutils"
)

func Example() {
	middleware := slogutils.NewMiddleware(
		slog.NewJSONHandler,
		slogutils.MiddlewareOptions{
			ModifierFuncs: map[slog.Level]slogutils.ModifierFunc{
				slog.LevelDebug: slogutils.Color(color.FgBlack),
				slog.LevelInfo:  nil,
				slog.LevelWarn:  slogutils.Color(color.FgYellow),
				slog.LevelError: slogutils.Color(color.FgRed, color.Bold),
			},
			RecordTransformerFuncs: []slogutils.RecordTransformerFunc{
				slogutils.ConvertLegacyLevel(
					map[string]slog.Level{
						"DEBUG": slog.LevelDebug,
						"INFO":  slog.LevelInfo,
						"WARN":  slog.LevelWarn,
						"ERROR": slog.LevelError,
					},
					false,
				),
			},
			Writer: os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelWarn,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == "time" {
						return slog.Attr{}
					}
					return a
				},
			},
		},
	)
	slog.SetDefault(slog.New(middleware))
	ctx := slogutils.With(context.Background(), slog.Int64("request_id", 12))
	slog.WarnContext(ctx, "foo")
	slog.ErrorContext(ctx, "bar")
	slog.DebugContext(ctx, "baz")
	slog.WarnContext(ctx, "buzz")
	log.Println("[DEBUG] this is not slog.")
	log.Println("[INFO] this is not slog.")
	log.Println("[WARN] this is not slog.")
	log.Println("[ERROR] this is not slog.")

	// Output:
	//{"level":"WARN","msg":"foo","request_id":12}
	//{"level":"ERROR","msg":"bar","request_id":12}
	//{"level":"WARN","msg":"buzz","request_id":12}
	//{"level":"WARN","msg":"this is not slog."}
	//{"level":"ERROR","msg":"this is not slog."}
}
