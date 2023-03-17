package logs

import (
	"golang.org/x/exp/slog"
)

func SetDefault(log *slog.Logger) {
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
