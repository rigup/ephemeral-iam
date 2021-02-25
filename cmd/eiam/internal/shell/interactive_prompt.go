package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/manifoldco/promptui"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

var promptTmpl = &promptui.PromptTemplates{
	Prompt:  "[{{ . }}] > ",
	Valid:   "[{{ . | green }}] > ",
	Invalid: "[{{ . | red }}] > ",
	Success: "[{{ . | bold }}] > ",
}

// CommandPrompt prompts the user for a command to execute
func CommandPrompt(sigint chan<- os.Signal) error {
	fmt.Println()
	prompt := promptui.Prompt{
		Label:     "eiam",
		Templates: promptTmpl,
	}
	input, err := prompt.Run() // TODO: directory navigation
	if err != nil {
		switch {
		case err.Error() == promptui.ErrInterrupt.Error():
			sigint <- syscall.SIGINT
			time.Sleep(100 * time.Millisecond) // Give promptui time to reset the cursor
			return nil
		default:
			util.Logger.Fatalf("Failed to run prompt: %v\n", err)
		}
	}
	args := strings.Fields(input)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		if serr, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("Failed to run command %s: %s", input, serr.Error())
		}
	}
	return nil
}
