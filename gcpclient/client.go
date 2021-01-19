package gcpclient

import (
	"context"
	"sync"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"emperror.dev/emperror"
)

var gcpClient *credentials.IamCredentialsClient
var privilegedClient *credentials.IamCredentialsClient
var once sync.Once

// GetGCPClient gets a gcloud client using the local gcloud configuration
func GetGCPClient() *credentials.IamCredentialsClient {
	once.Do(func() {
		gcpClient = newGcpClient()
	})
	return gcpClient
}

func newGcpClient() *credentials.IamCredentialsClient {
	ctx := context.Background()
	gcpClient, err := credentials.NewIamCredentialsClient(ctx)
	emperror.Panic(err)
	return gcpClient
}
