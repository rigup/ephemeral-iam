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

// Modified from https://github.com/davidovich/summon/master/tree/cmd/run.go
// see https://github.com/spf13/pflag/pull/160
// https://github.com/spf13/cobra/issues/739
// and https://github.com/spf13/pflag/pull/199
func extractUnknownArgs(flags *pflag.FlagSet, args []string) []string {
	unknownArgs := []string{}

	trimmed := args[2:]
	skipLoop := false
	for i := 0; i < len(trimmed); i++ {
		if skipLoop {
			skipLoop = false
			continue
		}
		var f *pflag.Flag
		a := trimmed[i]
		if a[0] == '-' {
			if a[1] == '-' {
				f = flags.Lookup(strings.SplitN(a[2:], "=", 2)[0])
			} else {
				f = flags.ShorthandLookup(string(a[1]))
			}
			if f == nil {
				unknownArgs = append(unknownArgs, a)
			} else {
				skipLoop = true
			}
		} else {
			unknownArgs = append(unknownArgs, a)
		}
	}
	args = difference(args, unknownArgs)
	return unknownArgs
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func confirm() error {
	prompt := promptui.Prompt{
		Label:     "Is this correct?",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("Prompt: %v", err)
	}

	return nil
}
