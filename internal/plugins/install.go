package plugins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-github/v33/github"
	"github.com/h2non/filetype"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

// InstallPlugin attempts to install an ephemeral-iam plugin from a given Github repo.
func InstallPlugin(repoOwner, repoName, tokenName string) error {
	token := ""
	if viper.GetBool(appconfig.GithubAuth) && tokenName != "" {
		tokenConfig := viper.GetStringMapString(appconfig.GithubTokens)
		if t, ok := tokenConfig[tokenName]; ok {
			token = t
		} else {
			util.Logger.Errorf("No Github token named %s exists. Continuing without authentication", tokenName)
		}
	}
	release, err := util.GetLatestRelease(repoOwner, repoName, token)
	if err != nil {
		if sErr, ok := err.(*github.ErrorResponse); ok {
			if sErr.Message == "Not Found" {
				return handleRepoNotFound(repoOwner, repoName)
			}
			return errorsutil.New(sErr.Message, errors.New(sErr.Response.Status))
		}
		return errorsutil.New("Failed to get release from repository", err)
	}
	downloadURL, err := util.GetReleaseDownloadURL(release, true)
	if err != nil {
		return errorsutil.New("Failed to get download URL for release", err)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "eiamplugin")
	if err != nil {
		return errorsutil.New("Failed to create temp dir for plugin", err)
	}
	defer os.Remove(tmpDir)
	if err := util.DownloadAndExtract(downloadURL, tmpDir, token); err != nil {
		return errorsutil.New("Failed to process the plugin release", err)
	}
	return installDownloadedPlugin(tmpDir)
}

func installDownloadedPlugin(tmpDir string) error {
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return errorsutil.New("Failed to list downloaded files", err)
	}

	pluginDir := filepath.Join(appconfig.GetConfigDir(), "plugins")
	for _, file := range files {
		fp := filepath.Join(tmpDir, file.Name())
		buf, err := ioutil.ReadFile(fp)
		if err != nil {
			return errorsutil.New("Failed to read file downloaded in release", err)
		}
		kind, err := filetype.Match(buf)
		if err != nil {
			return errorsutil.New("Failed to determine MIME type of file downloaded in release", err)
		}
		if kind.MIME.Value == "application/x-executable" {
			targetPath := filepath.Join(pluginDir, file.Name())
			if err := util.MoveFile(fp, targetPath); err != nil {
				return errorsutil.New("Failed to move plugin binary to plugins directory", err)
			}
			if err := os.Chmod(targetPath, 0o700); err != nil {
				if rmErr := os.Remove(targetPath); rmErr != nil {
					return errorsutil.New("Failed to update file permissions then remove binary", rmErr)
				}
				return errorsutil.New("Failed to make plugin binary executable", err)
			}
		}
	}
	return nil
}

func handleRepoNotFound(repoOwner, repoName string) error {
	util.Logger.WithFields(logrus.Fields{
		"repo": fmt.Sprintf("github.com/%s/%s", repoOwner, repoName),
	}).Errorf("Repository either doesn't exist, or it is private")

	if !viper.GetBool(appconfig.GithubAuth) {
		util.Logger.Warn("If the repo is private, add an access token with the 'plugins auth add' command")
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "Try to access repo with authentication",
		IsConfirm: true,
	}

	fmt.Println()
	if _, err := prompt.Run(); err != nil {
		return nil
	}
	fmt.Println()
	tokenConfig := viper.GetStringMapString(appconfig.GithubTokens)

	if len(tokenConfig) == 0 {
		viper.Set(appconfig.GithubAuth, false)
		if err := viper.WriteConfig(); err != nil {
			return errorsutil.New("Failed to update 'github.auth' field in config", err)
		}
		err := errors.New("no Github access tokens found")
		return errorsutil.New("Please add a Github access token using the 'plugins auth add' command", err)
	}

	var token string
	if len(tokenConfig) == 1 {
		for key := range tokenConfig {
			token = key
		}
	} else {
		tokenName, err := util.SelectToken(tokenConfig)
		if err != nil {
			return errorsutil.New("Failed to select access token", err)
		}
		token = tokenName
	}
	return InstallPlugin(repoOwner, repoName, token)
}
