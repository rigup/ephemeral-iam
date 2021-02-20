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
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
	"github.com/jessesomerville/ephemeral-iam/internal/loghandler"
)

var config = &appconfig.Config
var logger *logrus.Logger

var (
	// Accept states whether to prompt the user for confirmation or not
	Accept              bool
	serviceAccountEmail string
	reason              string
)

func init() {
	logger = loghandler.GetLogger(&config.Logging)
}

func handleErr(err error) {
	if err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}
}

func sessionID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("Failed to generate random log ID: %v", err)
	}
	return hex.EncodeToString(bytes), nil
}

func formatReason(reason string) (string, error) {
	randomID, err := sessionID()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ephemeral-iam %s: %s", randomID, reason), nil
}

func confirm() error {
	prompt := promptui.Prompt{
		Label:     "Is this correct",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("Prompt: %v", err)
	}

	return nil
}

// Modified from https://github.com/davidovich/summon/blob/master/cmd/run.go
// see https://github.com/spf13/pflag/pull/160
// https://github.com/spf13/cobra/issues/739
// and https://github.com/spf13/pflag/pull/199
func extractUnknownArgs(flags *pflag.FlagSet, args []string) []string {
	// Ensure args were passed to command
	if len(args) < 3 {
		return []string{}
	}
	trimmed := args[2:]

	unknownArgs := []string{}
	for i := 0; i < len(trimmed); i++ {
		currArg := trimmed[i]

		var currFlag *pflag.Flag
		if currArg[0] == '-' && len(currArg) > 1 {
			if currArg[1] == '-' {
				// Arg starts with two dashes, search for full flag names
				currFlag = flags.Lookup(strings.SplitN(currArg[2:], "=", 2)[0])
			} else {
				// Arg starts with single dash, look for single char shorthand flags
				currFlag = flags.ShorthandLookup(string(currArg[1]))
			}
		}

		// If the current flag is known and it accepts an argument, skip the next loop
		if currFlag != nil {
			if currFlag.NoOptDefVal == "" {
				if i+1 < len(trimmed) && currFlag.Value.String() == trimmed[i+1] {
					i++
				}
			}
			continue
		}
		unknownArgs = append(unknownArgs, currArg)
	}
	return unknownArgs
}
