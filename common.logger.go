package main

import (
	"os"

	"github.com/webdevops/go-common/log/slogger"
)

var (
	logger *slogger.Logger
)

func initLogger() *slogger.Logger {
	loggerOpts := []slogger.LoggerOptionFunc{
		slogger.WithLevelText(Opts.Logger.Level),
		slogger.WithFormat(slogger.FormatMode(Opts.Logger.Format)),
		slogger.WithSourceMode(slogger.SourceMode(Opts.Logger.Source)),
		slogger.WithTime(Opts.Logger.Time),
		slogger.WithColor(slogger.ColorMode(Opts.Logger.Color)),
	}

	logger = slogger.NewCliLogger(
		os.Stderr, loggerOpts...,
	)

	return logger
}
