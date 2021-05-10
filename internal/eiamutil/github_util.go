package eiamutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"

	archutil "github.com/rigup/ephemeral-iam/internal/appconfig/arch_util"
)

func GetLatestRelease(repoOwner, repoName, token string) (*github.RepositoryRelease, error) {
	httpClient := http.Client{}
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = *oauth2.NewClient(context.Background(), ts)
	}
	githubClient := github.NewClient(&httpClient)

	release, _, err := githubClient.Repositories.GetLatestRelease(context.Background(), repoOwner, repoName)
	if err != nil {
		return nil, err
	}
	return release, nil
}

func GetReleaseDownloadURL(r *github.RepositoryRelease, isPlugin bool) (string, error) {
	var downloadURL string
	currentRuntime := fmt.Sprintf("%s_%s", archutil.FormattedOS, archutil.FormattedArch)
	for _, asset := range r.Assets {
		if strings.Contains(asset.GetName(), currentRuntime) {
			if isPlugin {
				downloadURL = asset.GetURL()
			} else {
				downloadURL = asset.GetBrowserDownloadURL()
			}
			return downloadURL, nil
		}
	}
	if downloadURL == "" {
		return "", errors.New("failed to find a release version that matches your OS and architecture")
	}
	return downloadURL, nil
}
