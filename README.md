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

	"github.com/fatih/color"
	"github.com/mashiike/slogutils"
	"golang.org/x/exp/slog"
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
	slog.WarnCtx(ctx, "foo")
	slog.ErrorCtx(ctx, "bar")
	slog.DebugCtx(ctx, "baz")
	slog.WarnCtx(ctx, "buzz")
}
```

## License
This project is licensed under the MIT License - see the LICENSE(./LICENCE) file for details.

## Contribution
Contributions, bug reports, and feature requests are welcome. Pull requests are also highly appreciated. For more details, please
