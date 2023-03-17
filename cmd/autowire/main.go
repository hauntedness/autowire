package main

import (
	"os"
	"time"

	"github.com/hauntedness/autowire/logs"
	"github.com/hauntedness/autowire/pkg"
	"golang.org/x/exp/slog"
)

func main() {
	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
			}
			return a
		},
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	logs.SetDefault(log)
	di := pkg.NewDIContext(nil)
	di.Process("pattern=.")
}
