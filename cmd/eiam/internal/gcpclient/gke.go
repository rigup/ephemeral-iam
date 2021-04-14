package gcpclient

import (
	"context"
	"fmt"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
	errorsutil "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/errors"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

func GetClusters(project, reason string) ([]map[string]string, error) {
	gkeService, err := container.NewService(context.Background(), option.WithRequestReason(reason))
	if err != nil {
		return []map[string]string{}, &errorsutil.SDKClientCreateError{Err: err, ResourceType: "Container"}
	}
	clustersService := container.NewProjectsLocationsClustersService(gkeService)

	projectRes := fmt.Sprintf("projects/%s/locations/-", project)
	if resp, err := clustersService.List(projectRes).Do(); err != nil {
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
