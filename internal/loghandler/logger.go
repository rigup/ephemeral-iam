/*
Copyright © 2021 Jesse Somerville

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package loghandler

import (
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
)

var logger *logrus.Logger
var once sync.Once

// GetLogger returns the output log instance
func GetLogger(config *appconfig.LogConfig) *logrus.Logger {
	once.Do(func() {
		logger = newLogger(config)
	})
	return logger
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
		}
	}

	return logger
}
