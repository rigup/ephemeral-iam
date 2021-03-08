package eiamutil

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/pflag"
)

// FormatReason formats the reason field for logging visibility
func FormatReason(reason *string) error {
	randomID, err := sessionID()
	if err != nil {
		return err
	}

	s := fmt.Sprintf("ephemeral-iam %s: %s", randomID, *reason)
	reason = &s
	return nil
}

func sessionID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("Failed to generate random log ID: %v", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Confirm asks the user for confirmation before running a command
func Confirm(vals map[string]string) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 4, '-', 0)

	for key, val := range vals {
		fmt.Fprintf(w, "%s \t %s\n", key, val)
	}

	w.Flush()
	cmdInfo := strings.Split(buf.String(), "\n")

	fmt.Println()
	for _, line := range cmdInfo {
		fmt.Println(line)
	}

	prompt := promptui.Prompt{
		Label:     "Continue",
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
		Logger.Warn("Abandoning Command...")
		os.Exit(0)
	}
}

// ExtractUnknownArgs fetches unknown args passed to a command
// Modified from https://github.com/davidovich/summon/blob/master/cmd/run.go
// see https://github.com/spf13/pflag/pull/160
// and https://github.com/spf13/cobra/issues/739
// and https://github.com/spf13/pflag/pull/199
func ExtractUnknownArgs(flags *pflag.FlagSet, args []string) []string {
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

func Contains(values []string, val string) bool {
	for _, i := range values {
		if i == val {
			return true
		}
	}
	return false
}

func Uniq(a []string) []string {
	mb := make(map[string]struct{}, len(a))
	for _, x := range a {
		mb[x] = struct{}{}
	}
	set := make([]string, 0, len(mb))
	for k := range mb {
		set = append(set, k)
	}
	sort.Strings(set)
	return set
}
