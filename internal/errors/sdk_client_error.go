package errors

import "fmt"

// SDKClientCreateError is used for errors caused when attempting to
// create a GCP SDK client/service.
type SDKClientCreateError struct {
	Err            error
	ResourceType   string
	ServiceAccount string
}

func (e *SDKClientCreateError) Error() string {
	if e.ServiceAccount != "" {
		return fmt.Sprintf("failed to create %s SDK client with service account %s: %v", e.ResourceType, e.ServiceAccount, e.Err)
	}
	return fmt.Sprintf("failed to create %s SDK client: %v", e.ResourceType, e.Err)
}
