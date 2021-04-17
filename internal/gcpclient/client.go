package gcpclient

import (
	"context"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	errorsutil "github.com/jessesomerville/ephemeral-iam/internal/errors"
	"google.golang.org/api/option"
)

// ClientWithReason creates a client SDK with the provided reason field
func ClientWithReason(reason string) (*credentials.IamCredentialsClient, error) {
	ctx := context.Background()
	gcpClientWithReason, err := credentials.NewIamCredentialsClient(ctx, option.WithRequestReason(reason))
	if err != nil {
		return nil, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Credentials"}
	}
	return gcpClientWithReason, nil
}
