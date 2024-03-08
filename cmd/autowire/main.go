package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/hauntedness/autowire/pkg"
)

func main() {
	initLogger()
	di := pkg.NewDIContext(nil)
	di.Process("pattern=.")
}

func initLogger() {
	verbose := flag.Bool("v", false, "verbose output")
	level := slog.LevelWarn
	if *verbose {
		level = slog.LevelDebug
	}
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	handler := slog.NewTextHandler(os.Stdout, &opts)
	slog.SetDefault(slog.New(handler))
}
