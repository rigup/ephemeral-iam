package loghandler

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var once sync.Once

// GetLogger returns the output log instance
func GetLogger() *logrus.Logger {
	once.Do(func() {
		logger = newLogger()
	})
	return logger
}

func newLogger() *logrus.Logger {
	logger := logrus.New()

	logger.Level = logrus.InfoLevel

	// switch config.Format {
	// case "json":
	// 	logger.Formatter = new(logrus.JSONFormatter)

	// default:
	// 	logger.Formatter = &logrus.TextFormatter{
	// 		DisableLevelTruncation: true,
	// 		PadLevelText:           true,
	// 	}
	// }

	logger.Formatter = &logrus.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
	}

	return logger
}
