package gcpclient

import (
	"context"
	"fmt"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"

	util "github.com/jessesomerville/ephemeral-iam/cmd/eiam/internal/eiamutil"
)

func GetClusters(project, reason string) ([]map[string]string, error) {
	gkeService, err := container.NewService(context.Background(), option.WithRequestReason(reason))
	if err != nil {
		return []map[string]string{}, &util.SDKClientCreateError{Err: err, ResourceType: "Container"}
	}
	clustersService := container.NewProjectsLocationsClustersService(gkeService)

	projectRes := fmt.Sprintf("projects/%s/locations/-", project)
	if resp, err := clustersService.List(projectRes).Do(); err != nil {
		return []map[string]string{}, fmt.Errorf("failed to list clusters in %s: %v", project, err)
	} else {
		clusterNames := []map[string]string{}
		for _, cluster := range resp.Clusters {
			clusterNames = append(clusterNames, map[string]string{"name": cluster.Name, "location": cluster.Location})
		}
		return clusterNames, nil
	}
}
