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
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/v33/github"
	"github.com/manifoldco/promptui"

	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
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
	release, err := util.GetLatestRelease(repoOwner, repoName)
	if err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later")
		return
	}

	newestVersion := release.GetTagName()

	if newestVersion == Version {
		util.Logger.Debugf("Newest version of eiam (%s) is currently installed", newestVersion)
		return
	}

	util.Logger.Infof("Found a new version of eiam: %s (installed version is %s)", newestVersion, Version)
	updatePrompt := fmt.Sprintf("Would you like to install eiam %s now", newestVersion)
	prompt := promptui.Prompt{
		Label:     updatePrompt,
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err == nil {
		util.Logger.Infof("Installing eiam %s", newestVersion)
		installNewVersion(release)
	}
}

func installNewVersion(release *github.RepositoryRelease) {
	downloadURL, err := util.GetReleaseDownloadURL(release)
	if err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later")
		return
	}
	tmpDir := os.TempDir()
	if err = util.DownloadAndExtract(downloadURL, tmpDir); err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later")
		return
	}

	installPath, err := CheckCommandExists("eiam")
	if err != nil {
		util.Logger.WithError(err).Error("Skipping update, please try again later")
		return
	}

	tmpLoc := filepath.Join(tmpDir, "eiam")
	if err := os.Rename(tmpLoc, installPath); err != nil {
		util.Logger.Errorf("failed to move %s to %s", tmpLoc, installPath)
		return
	}
	util.Logger.Info("Update completed successfully")
	os.Exit(0)
}
