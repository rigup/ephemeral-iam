package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rigup/ephemeral-iam/cmd/eiam"
	"github.com/rigup/ephemeral-iam/pkg/options"
	testutil "github.com/rigup/ephemeral-iam/test"
)

var (
	gcloudCommand *cobra.Command
	commonArgs    []string
)

func TestIntegration_All(t *testing.T) {
	if testutil.RunIntegrationTests == "" {
		t.Skipf("$%s is not set. Skipping integration tests for gcloud command", testutil.RunIntegrationTestsEnv)
	}
	if testutil.ServiceAccountEmail == "" || testutil.Reason == "" {
		t.Skipf("Both $%s and $%s must be set to run integration tests", testutil.TestServiceAccountEmailEnv, testutil.TestReasonEnv)
	}
	gcloudCommand = eiam.NewCmdGcloud()

	options.AddPersistentFlags(gcloudCommand.PersistentFlags())

	testGcloudCommandNoArgs(t)
	testGcloudCommandInvalidSA(t)
	testGcloudCommandNoAccessToSA(t)
	testGcloudInfo(t)
	testGcloudComputeInstancesList(t)
}

func testGcloudCommandNoArgs(t *testing.T) {
	expectedOutput := "(gcloud) Command name argument expected."
	output, err := testutil.ExecuteCommand(gcloudCommand, commonArgs...)
	if err == nil {
		t.Errorf("expected error caused by passing no input arguments\nOUTPUT:\n%s", output)
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: \"%s\"\nACTUAL: %s", expectedOutput, output)
	}
}

func testGcloudCommandInvalidSA(t *testing.T) {
	expectedOutput := "Requested entity was not found."
	output, err := testutil.ExecuteCommand(gcloudCommand, "version", "-s", "wrong@email.com", "-R", testutil.Reason, "--yes")
	if err == nil {
		t.Errorf("expected error caused by invalid service account email\nOUTPUT:\n%s", output)
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: \"%s\"\nACTUAL: %s", expectedOutput, output)
	}
}

func testGcloudCommandNoAccessToSA(t *testing.T) {
	expectedOutput := fmt.Sprintf("cannot impersonate %s", testutil.ServiceAccountEmailNoAccess)
	output, err := testutil.ExecuteCommand(gcloudCommand, "version", "-s", testutil.ServiceAccountEmailNoAccess, "-R", testutil.Reason, "--yes")
	if err == nil {
		t.Errorf("expected error caused by no access to service account\nOUTPUT:\n%s", output)
	}
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("unexpected output:\nEXPECTED TO FIND: \"%s\"\nACTUAL: %s", expectedOutput, output)
	}
}

func testGcloudInfo(t *testing.T) {
	expectedOutput := regexp.MustCompile(`Google Cloud SDK \d+\.\d+\.\d+`)
	output, err := testutil.ExecuteCommand(gcloudCommand, append(commonArgs, "version")...)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	if !expectedOutput.MatchString(output) {
		t.Errorf("unexpected output:\n%s", output)
	}
}

func testGcloudComputeInstancesList(t *testing.T) {
	expectedOutput := regexp.MustCompile(`NAME\s+ZONE\s+MACHINE_TYPE\s+PREEMPTIBLE\s+INTERNAL_IP\s+EXTERNAL_IP\s+STATUS`)
	output, err := testutil.ExecuteCommand(gcloudCommand, append(commonArgs, "compute", "instances", "list")...)
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}
	if !expectedOutput.MatchString(output) {
		t.Errorf("unexpected output:\n%s", output)
	}
}
