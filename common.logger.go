package main

import (
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

var (
	logger  *zap.SugaredLogger
	slogger *slog.Logger
)

func initLogger() *zap.SugaredLogger {
	var config zap.Config
	if Opts.Logger.Development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// debug level
	if Opts.Logger.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	// json log format
	if Opts.Logger.Json {
		config.Encoding = "json"

		// if running in containers, logs already enriched with timestamp by the container runtime
		config.EncoderConfig.TimeKey = ""
	}

	// build logger
	log, err := config.Build()
	if err != nil {
		panic(err)
	}

	logger = log.Sugar()
	slogger = slog.New(zapslog.NewHandler(log.Core(), nil))

	return logger
}
