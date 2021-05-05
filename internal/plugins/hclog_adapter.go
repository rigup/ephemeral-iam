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

package plugins

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

func NewHCLogAdapter(l *logrus.Logger, name string) hclog.Logger {
	return &HCLogAdapter{l, name, nil}
}

type HCLogAdapter struct {
	l    *logrus.Logger
	name string

	impliedArgs []interface{}
}

func (h HCLogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
		return
	case hclog.Trace:
		h.Trace(msg, args...)
	case hclog.Debug:
		h.Debug(msg, args...)
	case hclog.Info:
		h.Info(msg, args...)
	case hclog.Warn:
		h.Warn(msg, args...)
	case hclog.Error:
		h.Error(msg, args...)
	}
}

func (h HCLogAdapter) Trace(msg string, args ...interface{}) {
	h.l.WithFields(toMap(args)).Trace(msg)
}

func (h HCLogAdapter) Debug(msg string, args ...interface{}) {
	h.l.WithFields(toMap(args)).Debug(msg)
}

func (h HCLogAdapter) Info(msg string, args ...interface{}) {
	h.l.WithFields(toMap(args)).Info(msg)
}

func (h HCLogAdapter) Warn(msg string, args ...interface{}) {
	h.l.WithFields(toMap(args)).Warn(msg)
}

func (h HCLogAdapter) Error(msg string, args ...interface{}) {
	h.l.WithFields(toMap(args)).Error(msg)
}

func (h HCLogAdapter) IsTrace() bool {
	return h.l.GetLevel() >= logrus.TraceLevel
}

func (h HCLogAdapter) IsDebug() bool {
	return h.l.GetLevel() >= logrus.DebugLevel
}

func (h HCLogAdapter) IsInfo() bool {
	return h.l.GetLevel() >= logrus.InfoLevel
}

func (h HCLogAdapter) IsWarn() bool {
	return h.l.GetLevel() >= logrus.WarnLevel
}

func (h HCLogAdapter) IsError() bool {
	return h.l.GetLevel() >= logrus.ErrorLevel
}

func (h HCLogAdapter) ImpliedArgs() []interface{} {
	// Not supported.
	return nil
}

func (h HCLogAdapter) With(args ...interface{}) hclog.Logger {
	return &h
}

func (h HCLogAdapter) Name() string {
	return h.name
}

func (h HCLogAdapter) Named(name string) hclog.Logger {
	return NewHCLogAdapter(h.l, name)
}

func (h HCLogAdapter) ResetNamed(name string) hclog.Logger {
	return &h
}

func (h *HCLogAdapter) SetLevel(level hclog.Level) {
	h.l.SetLevel(convertLevel(level))
}

func (h HCLogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &hclog.StandardLoggerOptions{}
	}
	return log.New(h.StandardWriter(opts), "", 0)
}

func (h HCLogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}

func convertLevel(level hclog.Level) logrus.Level {
	switch level {
	case hclog.NoLevel:
		return logrus.InfoLevel
	case hclog.Trace:
		return logrus.TraceLevel
	case hclog.Debug:
		return logrus.DebugLevel
	case hclog.Info:
		return logrus.InfoLevel
	case hclog.Warn:
		return logrus.WarnLevel
	case hclog.Error:
		return logrus.ErrorLevel
	}
	return logrus.InfoLevel
}

func toMap(kvs []interface{}) map[string]interface{} {
	m := map[string]interface{}{}

	if len(kvs) == 0 {
		return m
	}

	if len(kvs)%2 == 1 {
		kvs = append(kvs, nil)
	}

	for i := 0; i < len(kvs); i += 2 {
		if kvs[i] != "timestamp" {
			merge(m, kvs[i], kvs[i+1])
		}
	}

	return m
}

func merge(dst map[string]interface{}, k, v interface{}) {
	var key string

	switch x := k.(type) {
	case string:
		key = x
	case fmt.Stringer:
		key = safeString(x)
	default:
		key = fmt.Sprint(x)
	}

	dst[key] = v
}

func safeString(str fmt.Stringer) (s string) {
	defer func() {
		if panicVal := recover(); panicVal != nil {
			if v := reflect.ValueOf(str); v.Kind() == reflect.Ptr && v.IsNil() {
				s = "NULL"
			} else {
				panic(panicVal)
			}
		}
	}()

	s = str.String()

	return
}
