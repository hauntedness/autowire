package pkg

import (
	"os"
	"time"

	"golang.org/x/exp/slog"
)

func init() {
	opts := slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
			}
			return a
		},
	}
	log := slog.New(opts.NewTextHandler(os.Stdout))
	slog.SetDefault(log)
}
