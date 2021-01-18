package google

import (
	"context"
	"sync"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"emperror.dev/emperror"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
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

// GetPrivilegedClient returns an instance of a GCP client using the provided OAuth 2 token
func GetPrivilegedClient(tokenSource oauth2.TokenSource) *credentials.IamCredentialsClient {
	once.Do(func() {
		privilegedClient = newPrivClient(tokenSource)
	})
	return privilegedClient
}

func newGcpClient() *credentials.IamCredentialsClient {
	ctx := context.Background()
	gcpClient, err := credentials.NewIamCredentialsClient(ctx)
	emperror.Panic(err)
	return gcpClient
}

func newPrivClient(tokenSource oauth2.TokenSource) *credentials.IamCredentialsClient {
	ctx := context.Background()
	privClient, err := credentials.NewIamCredentialsClient(ctx, option.WithTokenSource(tokenSource))
	emperror.Panic(err)
	return privClient
}
