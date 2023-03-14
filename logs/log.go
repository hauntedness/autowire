package logs

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

func Log(level slog.Level, msg string, args ...any) {
	slog.Log(level, msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, err error, args ...any) {
	slog.Error(msg, err, args...)
}

func Group(key string, as ...slog.Attr) {
	slog.Group(key, as...)
}
