package eiamutil

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Logger is the global logging instance
var Logger *logrus.Logger

// NewLogger instantiates a new logging instance
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger instance: %v", err)
		os.Exit(1)
	}

	logger.Level = level
	logger.Out = os.Stderr

	switch viper.GetString("logging.format") {
	case "json":
		logger.Formatter = new(logrus.JSONFormatter)

	default:
		logger.Formatter = &logrus.TextFormatter{
			DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
			PadLevelText:           viper.GetBool("logging.padleveltext"),
			DisableTimestamp:       true,
		}
	}

	return logger
}
