// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eiamutil

import (
	"log"
	"os"

	rt "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Logger is the global logging instance.
var Logger *logrus.Logger

// NewLogger instantiates a new logging instance.
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(viper.GetString("logging.level"))
	if err != nil {
		log.Fatalf("Failed to create logger instance: %v", err)
		os.Exit(1)
	}

	logger.Level = level
	logger.Out = os.Stderr

	switch viper.GetString("logging.format") {
	case "json":
		logger.Formatter = NewJSONFormatter()
	case "debug":
		// The 'debug' formatter will include the filename, function, and line number
		// that a log entry is written from.
		logger.Formatter = NewRuntimeFormatter()
	default:
		logger.Formatter = NewTextFormatter()
	}

	return logger
}

func NewTextFormatter() *logrus.TextFormatter {
	return &logrus.TextFormatter{
		DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
		DisableQuote:           true,
		DisableTimestamp:       true,
		PadLevelText:           viper.GetBool("logging.padleveltext"),
	}
}

func NewJSONFormatter() *logrus.JSONFormatter {
	return new(logrus.JSONFormatter)
}

func NewRuntimeFormatter() *rt.Formatter {
	return &rt.Formatter{
		ChildFormatter: &logrus.TextFormatter{
			DisableLevelTruncation: viper.GetBool("logging.disableleveltruncation"),
			DisableQuote:           true,
			DisableTimestamp:       true,
			PadLevelText:           viper.GetBool("logging.padleveltext"),
		},
		Line: true,
	}
}
