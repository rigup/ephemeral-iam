// Copyright 2021 Workrise Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appconfig

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/google/go-github/v33/github"
	"github.com/manifoldco/promptui"

	archutil "github.com/rigup/ephemeral-iam/internal/appconfig/arch_util"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

var (
	repoOwner = "rigup"
	repoName  = "ephemeral-iam"

	// Version is the currently installed eiam client version.
	// This is populated by goreleaser when a new release is built.
	Version = "v0.0.0"
)

// CheckForNewRelease checks to see if there is a new version of eiam available.
func CheckForNewRelease() {
	githubClient := github.NewClient(nil)
	releases, _, err := githubClient.Repositories.ListReleases(context.Background(), repoOwner, repoName, nil)
	if err != nil {
		util.Logger.WithError(err).Error("unable to check for new eiam releases")
	}

	newestVersion := Version
	var newestRelease *github.RepositoryRelease
	for _, release := range releases {
		if semver.Compare(release.GetTagName(), newestVersion) > 0 {
			newestVersion = release.GetTagName()
			newestRelease = release
		}
	}

	if newestVersion == Version {
		util.Logger.Debugf("Newest version of eiam (%s) is currently installed", newestVersion)
	} else {
		util.Logger.Infof("Found a new version of eiam: %s (installed version is %s)", newestVersion, Version)
		updatePrompt := fmt.Sprintf("Would you like to install eiam %s now", newestVersion)
		prompt := promptui.Prompt{
			Label:     updatePrompt,
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err == nil {
			util.Logger.Infof("Installing eiam %s", newestVersion)
			installNewVersion(newestRelease)
		}
	}
}

func installNewVersion(release *github.RepositoryRelease) {
	var downloadURL string
	currentRuntime := fmt.Sprintf("%s_%s", archutil.FormattedOS, archutil.FormattedArch)
	for _, asset := range release.Assets {
		if strings.Contains(asset.GetName(), currentRuntime) {
			downloadURL = asset.GetBrowserDownloadURL()
			break
		}
	}
	if downloadURL == "" {
		err := fmt.Errorf("failed to find a release version that matches your OS and architecture")
		util.Logger.WithError(err).Error("Skipping update, please try again later\n")
		return
	}

	if err := downloadAndExtract(downloadURL); err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later")
		return
	}

	installPath, err := CheckCommandExists("eiam")
	if err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later\n")
		return
	}

	tmpLoc := filepath.Join(os.TempDir(), "eiam")
	if err := os.Rename(tmpLoc, installPath); err != nil {
		util.Logger.Errorf("failed to move %s to %s", tmpLoc, installPath)
		return
	}
	util.Logger.Info("Update completed successfully")
	os.Exit(0)
}

func downloadAndExtract(url string) error {
	util.Logger.Infof("Downloading new version from %s", url)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to create HTTP client",
			Err: err,
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: fmt.Sprintf("Failed to download release from %s", url),
			Err: err,
		}
	}
	defer resp.Body.Close()

	util.Logger.Info("Successfully downloaded the archive, now extracting its contents")
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to create gzip reader",
			Err: err,
		}
	}
	defer gzr.Close()

	tmpDir := os.TempDir()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()

		target := filepath.Join(tmpDir, filepath.Clean(header.Name))
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: "Failed to extract release archive",
				Err: err,
			}
		case header.Typeflag == tar.TypeDir:
			if err = os.MkdirAll(target, 0o755); err != nil {
				return errorsutil.EiamError{
					Log: util.Logger.WithError(err),
					Msg: "Failed to create directory while extracting release archive",
					Err: err,
				}
			}
		case header.Typeflag == tar.TypeReg:
			var f *os.File
			f, err = os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return errorsutil.EiamError{
					Log: util.Logger.WithError(err),
					Msg: "Failed to create file while extracting release archive",
					Err: err,
				}
			}
			// Limit readable amount to 2GB to prevent decompression bomb.
			maxSize := 2 << (10 * 3)
			limiter := io.LimitReader(tarReader, int64(maxSize))
			if _, err = io.Copy(f, limiter); err != nil {
				return errorsutil.EiamError{
					Log: util.Logger.WithError(err),
					Msg: "Failed to copy file contents while extracting release archive",
					Err: err,
				}
			}
			// Manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		default:
			err = fmt.Errorf("unknown type %v in %s", header.Typeflag, header.Name)
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: "Encountered unknown type while extracting update",
				Err: err,
			}
		}
	}
}
