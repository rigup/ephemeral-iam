package plugins

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/h2non/filetype"

	"github.com/rigup/ephemeral-iam/internal/appconfig"
	util "github.com/rigup/ephemeral-iam/internal/eiamutil"
	errorsutil "github.com/rigup/ephemeral-iam/internal/errors"
)

// InstallPlugin attempts to install an ephemeral-iam plugin from a given Github repo.
func InstallPlugin(repoOwner, repoName string) error {
	release, err := util.GetLatestRelease(repoOwner, repoName)
	if err != nil {
		return errorsutil.New("Failed to get release from repository", err)
	}
	downloadURL, err := util.GetReleaseDownloadURL(release)
	if err != nil {
		return errorsutil.New("Failed to get download URL for release", err)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "eiamplugin")
	if err != nil {
		return errorsutil.New("Failed to create temp dir for plugin", err)
	}
	if err := util.DownloadAndExtract(downloadURL, tmpDir); err != nil {
		return err
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
			if err := util.MoveFile(fp, filepath.Join(pluginDir, file.Name())); err != nil {
				return errorsutil.New("Failed to move plugin binary to plugins directory", err)
			}
		}
	}
	return nil
}
