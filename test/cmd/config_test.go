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

package eiam

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/lithammer/dedent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rigup/ephemeral-iam/cmd"
	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	testutil "github.com/rigup/ephemeral-iam/test"
)

var (
	configCommand     = cmd.NewCmdConfig()
	configSetCommand  = cmd.NewCmdConfigSet()
	configViewCommand = cmd.NewCmdConfigView()
)

func TestConfigNoSubCommand(t *testing.T) {
	expectedOutput := dedent.Dedent(`
        Manage configuration values

        Usage:
          config [command]

        Available Commands:
          help        Help about any command
          info        Print information about config fields
          print       Print the current configuration
          set         Set the value of a provided config item
          view        View the value of a provided config item

        Flags:
          -h, --help   help for config

        Use "config [command] --help" for more information about a command.`)

	output, err := testutil.ExecuteCommand(configCommand)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.TrimSpace(output) != strings.TrimSpace(expectedOutput) {
		t.Errorf("unexpected output:\n%s", diff.LineDiff(expectedOutput, output))
	}
}

func TestConfigViewCommand(t *testing.T) {
	for configKey, configValue := range allSettings {
		output, err := testutil.ExecuteCommand(configViewCommand, configKey)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedOutput := fmt.Sprintf("%s: %v", configKey, configValue)
		if !strings.Contains(output, expectedOutput) {
			t.Errorf("unexpected output:\nEXPECTED TO FIND: level=info msg=\"%s\"\nACTUAL: %s", expectedOutput, output)
		}
	}
}

func TestConfigSetCommandWithoutSufficientArgs(t *testing.T) {
	output, err := testutil.ExecuteCommand(configSetCommand)
	if err == nil {
		t.Errorf("expected error caused by passing no input arguments\nOUTPUT:\n%s", output)
	}
	expectedOutput := "requires both a config key and a new value"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: %s\nACTUAL: %v", expectedOutput, output)
	}
	output, err = testutil.ExecuteCommand(configSetCommand, "logging.level")
	if err == nil {
		t.Errorf("expected error caused by passing only one input argument\nOUTPUT:\n%s", output)
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: %s\nACTUAL: %s", expectedOutput, output)
	}
}

func TestConfigSetCommandInvalidKey(t *testing.T) {
	output, err := testutil.ExecuteCommand(configSetCommand, "notakey.thatexists", "some value")
	if err == nil {
		t.Errorf("expected error caused by invalid key name 'notakey.thatexists'\nOUTPUT:\n%s", output)
	}
	expectedOutput := "invalid config key notakey.thatexists"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: %s\nACTUAL: %s", expectedOutput, output)
	}
}

func setLoggingLevel(t *testing.T, currLogLevel, newLogLevel string) {
	output, err := testutil.ExecuteCommand(configSetCommand, "logging.level", newLogLevel)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOutput := fmt.Sprintf("Updated logging.level from %s to %s", currLogLevel, newLogLevel)
	if !util.Contains([]string{"trace", "debug", "info"}, newLogLevel) {
		expectedOutput = ""
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: level=info msg=\"%s\"\nACTUAL: %s", expectedOutput, output)
	}
}

func TestConfigSetCommandSetLoggingLevel(t *testing.T) {
	initialLogLevel := viper.GetString(appconfig.LoggingLevel)

	// Try setting the log level to an invalid value
	output, err := testutil.ExecuteCommand(configSetCommand, "logging.level", "invalid")
	if err == nil {
		t.Errorf("expected error caused by invalid logging level value `invalid`\nOUTPUT:\n%s", output)
	}

	currLogLevel := initialLogLevel
	setLoggingLevel(t, currLogLevel, "trace")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "trace" {
		t.Error("Unexpected failure: The `eiam config set logging.level trace` command did not properly update the config")
	}
	currLogLevel = "trace"

	setLoggingLevel(t, currLogLevel, "debug")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "debug" {
		t.Error("Unexpected failure: The `eiam config set logging.level debug` command did not properly update the config")
	}
	currLogLevel = "debug"

	setLoggingLevel(t, currLogLevel, "info")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "info" {
		t.Error("Unexpected failure: The `eiam config set logging.level info` command did not properly update the config")
	}
	currLogLevel = "info"

	setLoggingLevel(t, currLogLevel, "warn")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "warn" {
		t.Error("Unexpected failure: The `eiam config set logging.level warn` command did not properly update the config")
	}
	currLogLevel = "warn"

	setLoggingLevel(t, currLogLevel, "error")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "error" {
		t.Error("Unexpected failure: The `eiam config set logging.level error` command did not properly update the config")
	}
	currLogLevel = "error"

	setLoggingLevel(t, currLogLevel, "fatal")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "fatal" {
		t.Error("Unexpected failure: The `eiam config set logging.level fatal` command did not properly update the config")
	}
	currLogLevel = "fatal"

	setLoggingLevel(t, currLogLevel, "panic")
	if logLevel := viper.GetString(appconfig.LoggingLevel); logLevel != "panic" {
		t.Error("Unexpected failure: The `eiam config set logging.level panic` command did not properly update the config")
	}
	currLogLevel = "panic"

	setLoggingLevel(t, currLogLevel, initialLogLevel)
}

func TestConfigSetCommandSetLoggingFormat(t *testing.T) {
	initialLogFormat := viper.GetString(appconfig.LoggingFormat)

	output, err := testutil.ExecuteCommand(configSetCommand, "logging.format", "invalid")
	if err == nil {
		t.Errorf("expected error caused by invalid logging format value `invalid`\nOUTPUT:\n%s", output)
	}

	output, err = testutil.ExecuteCommand(configSetCommand, "logging.format", "text")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOutput := "New value is the same as the current one"
	if initialLogFormat != "text" {
		expectedOutput = fmt.Sprintf("Updated logging.format from %s to text", initialLogFormat)
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: level=info msg=\"%s\"\nACTUAL: %s", expectedOutput, output)
	}
	if logFormat := viper.GetString(appconfig.LoggingFormat); logFormat != "text" {
		t.Error("Unexpected failure: The `eiam config set logging.format text` command did not properly update the config")
	}

	output, err = testutil.ExecuteCommand(configSetCommand, "logging.format", "json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedOutput = "Updated logging.format from text to json"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: level=info msg=\"%s\"\nACTUAL: %s", expectedOutput, output)
	}
	if logFormat := viper.GetString(appconfig.LoggingFormat); logFormat != "json" {
		t.Error("Unexpected failure: The `eiam config set logging.format json` command did not properly update the config")
	}

	output, err = testutil.ExecuteCommand(configViewCommand, "logging.format")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var jsonLogEntry logrus.Entry
	if err := json.Unmarshal([]byte(output), &jsonLogEntry); err != nil {
		t.Errorf("failed to parse JSON logging output: %v\nOUTPUT: %s", err, output)
	}

	_, err = testutil.ExecuteCommand(configSetCommand, "logging.format", initialLogFormat)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
