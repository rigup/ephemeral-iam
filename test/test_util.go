package eiamtests

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
)

const (
	RunIntegrationTestsEnv        = "EIAM_INTEGRATION_TESTS_RUN"
	TestServiceAccountEmailEnv    = "EIAM_INTEGRATION_TESTS_SA"
	TestServiceAccountNoAccessEnv = "EIAM_INTEGRATION_TESTS_SA_NOACCESS"
	TestReasonEnv                 = "EIAM_INTEGRATION_TESTS_REASON"
)

var (
	RunIntegrationTests         string
	ServiceAccountEmail         string
	ServiceAccountEmailNoAccess string
	Reason                      string
	CommonArgs                  []string
)

var once sync.Once

func InitEiam() error {
	var err error
	once.Do(func() {
		if initErr := appconfig.InitConfig(); initErr != nil {
			err = initErr
		}
		if initErr := appconfig.Setup(); initErr != nil {
			err = initErr
		}
	})
	return err
}

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	RunIntegrationTests = os.Getenv(RunIntegrationTestsEnv)
	ServiceAccountEmail = os.Getenv(TestServiceAccountEmailEnv)
	ServiceAccountEmailNoAccess = os.Getenv(TestServiceAccountNoAccessEnv)
	Reason = os.Getenv(TestReasonEnv)

	CommonArgs = []string{"-s", ServiceAccountEmail, "-R", Reason, "--yes"}
}

// https://chromium.googlesource.com/external/github.com/spf13/cobra/+/refs/heads/master/command_test.go
func ExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	root.SetArgs(args)

	rescueStdout := os.Stdout
	rescueStderr := os.Stderr

	r, w, execErr := os.Pipe()
	if execErr != nil {
		log.Fatal(execErr)
	}
	os.Stdout = w
	os.Stderr = w
	util.Logger.Out = w

	c, err = root.ExecuteC()

	w.Close()
	cmdOutBytes, execErr := ioutil.ReadAll(r)
	if execErr != nil {
		log.Fatal(execErr)
	}
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr

	output = string(cmdOutBytes)
	return c, output, err
}
