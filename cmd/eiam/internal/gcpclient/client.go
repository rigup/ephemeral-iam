package gcpclient

import (
	"context"
	"fmt"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"google.golang.org/api/option"
)

// ClientWithReason creates a client SDK with the provided reason field
func ClientWithReason(reason string) (*credentials.IamCredentialsClient, error) {
	ctx := context.Background()
	gcpClientWithReason, err := credentials.NewIamCredentialsClient(ctx, option.WithRequestReason(reason))
	if err != nil {
		return nil, fmt.Errorf("Failed to create a client SDK with a reason field: %v", err)
	}
	return gcpClientWithReason, nil
}
