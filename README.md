# slogutils
This package provides a middleware and utility functions for slog 


[![GoDoc](https://godoc.org/github.com/mashiike/slogutils?status.svg)](https://godoc.org/github.com/mashiike/slogutils)
[![Go Report Card](https://goreportcard.com/badge/github.com/mashiike/slogutils)](https://goreportcard.com/report/github.com/mashiike/slogutils)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Overview

This package provides middleware and utility functions for easy logging management. It enables color-coded display for different log levels and automatically collects attributes set in the context. This allows developers to have flexible logging recording and analysis capabilities.

## Installation

```bash
go get github.com/mashiike/slogutils
```

## Usage

```go
package main

import (
	"context"
	"os"
	"log/slog"

	"github.com/fatih/color"
	"github.com/mashiike/slogutils"
)

func main() {
	middleware := slogutils.NewMiddleware(
		slog.NewJSONHandler,
		slogutils.MiddlewareOptions{
			ModifierFuncs: map[slog.Level]slogutils.ModifierFunc{
				slog.LevelDebug: slogutils.Color(color.FgBlack),
				slog.LevelInfo:  nil,
				slog.LevelWarn:  slogutils.Color(color.FgYellow),
				slog.LevelError: slogutils.Color(color.FgRed, color.Bold),
			},
			Writer: os.Stderr,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelWarn,
			},
		},
	)
	slog.SetDefault(slog.New(middleware))
	ctx := slogutils.With(context.Background(), slog.Int64("request_id", 12))
	slog.WarnContext(ctx, "foo")
	slog.ErrorContext(ctx, "bar")
	slog.DebugContext(ctx, "baz")
	slog.WarnContext(ctx, "buzz")
}
```

## Benchmark

```bash
$ go test -bench . -benchmem         
goos: darwin
goarch: arm64
pkg: github.com/mashiike/slogutils
BenchmarkSlogDefault-8                           3304731               343.8 ns/op             0 B/op          0 allocs/op
BenchmarkLogOutput-8                            262371212                4.491 ns/op           0 B/op          0 allocs/op
BenchmarkMiddleware-8                            2929984               415.4 ns/op            48 B/op          1 allocs/op
BenchmarkMiddlewareWithRecordTrnasformer-8       1614752               741.3 ns/op           104 B/op          2 allocs/op
BenchmarkLogOutputWithMiddleware-8              13597328                92.34 ns/op            0 B/op          0 allocs/op
BenchmarkLogOutputWithRecordTransformer-8        1786766               672.7 ns/op             4 B/op          1 allocs/op
PASS
ok      github.com/mashiike/slogutils   10.145s
```

## License
This project is licensed under the MIT License - see the LICENSE(./LICENCE) file for details.

## Contribution
Contributions, bug reports, and feature requests are welcome. Pull requests are also highly appreciated. For more details, please
