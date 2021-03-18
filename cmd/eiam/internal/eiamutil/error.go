package eiamutil

import (
	"fmt"
	"strings"
)

type SDKClientCreateError struct {
	Err            error
	ResourceType   string
	ServiceAccount string
}

func (e *SDKClientCreateError) Error() string {
	if e.ServiceAccount != "" {
		return fmt.Sprintf("failed to create %s SDK client with service account %s: %s", e.ResourceType, e.ServiceAccount, e.Err)
	}
	return fmt.Sprintf("failed to create %s SDK client: %s", e.ResourceType, e.Err)
}

// CheckError handles simple error handling
func CheckError(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "could not find default credentials") {
			Logger.Fatal("No Application Default Credentials were found. Please run the following command to remediate this issue: \n\n  $ gcloud auth application-default login\n\n")
		}
		fmt.Println()
		Logger.Fatalf("%s\n\n", err.Error())
	}
}
