package eiamutil

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/appconfig"
)

var (
	config = &appconfig.Config
	// Logger is the global logging instance
	Logger *logrus.Logger
)

func init() {
	Logger = newLogger(&config.Logging)
}

func newLogger(config *appconfig.LogConfig) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger instance: %v", err)
		os.Exit(1)
	}

	logger.Level = level
	logger.Out = os.Stderr

	switch config.Format {
	case "json":
		logger.Formatter = new(logrus.JSONFormatter)

	default:
		logger.Formatter = &logrus.TextFormatter{
			DisableLevelTruncation: config.DisableLevelTrucation,
			PadLevelText:           config.PadLevelText,
			DisableTimestamp:       true,
		}
	}

	return logger
}
