package eiamutil

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v33/github"

	archutil "github.com/rigup/ephemeral-iam/internal/appconfig/arch_util"
)

func GetLatestRelease(repoOwner, repoName string) (*github.RepositoryRelease, error) {
	listOpts := &github.ListOptions{
		PerPage: 1,
	}
	githubClient := github.NewClient(nil)
	releases, _, err := githubClient.Repositories.ListReleases(context.Background(), repoOwner, repoName, listOpts)
	if err != nil {
		return nil, err
	}
	return releases[0], nil
}

func GetReleaseDownloadURL(r *github.RepositoryRelease) (string, error) {
	var downloadURL string
	currentRuntime := fmt.Sprintf("%s_%s", archutil.FormattedOS, archutil.FormattedArch)
	for _, asset := range r.Assets {
		if strings.Contains(asset.GetName(), currentRuntime) {
			downloadURL = asset.GetBrowserDownloadURL()
			break
		}
	}
	if downloadURL == "" {
		return "", errors.New("failed to find a release version that matches your OS and architecture")
	}
	return downloadURL, nil
}
