package gcpclient

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

func GetClusters(project, reason string) ([]map[string]string, error) {
	gkeClient, err := container.NewClusterManagerClient(context.Background(), option.WithRequestReason(reason))
	if err != nil {
		return []map[string]string{}, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Container"}
	}

	listClustersReq := &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", project),
	}

	if resp, err := gkeClient.ListClusters(ctx, listClustersReq); err != nil {
		util.Logger.Error("Failed to list GKE clusters")
		return []map[string]string{}, err
	} else {
		clusterNames := []map[string]string{}
		for _, cluster := range resp.Clusters {
			clusterNames = append(clusterNames, map[string]string{"name": cluster.Name, "location": cluster.Location})
		}
		return clusterNames, nil
	}
}
