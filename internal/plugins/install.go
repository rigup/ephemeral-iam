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

func InstallPlugin(repoOwner, repoName string) error {
	release, err := util.GetLatestRelease(repoOwner, repoName)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to get release from repository",
			Err: err,
		}
	}
	downloadURL, err := util.GetReleaseDownloadURL(release)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to get download URL for release",
			Err: err,
		}
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "eiamplugin")
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to create temp dir for plugin",
			Err: err,
		}
	}
	if err := util.DownloadAndExtract(downloadURL, tmpDir); err != nil {
		return err
	}
	return installDownloadedPlugin(tmpDir)
}

func installDownloadedPlugin(tmpDir string) error {
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return errorsutil.EiamError{
			Log: util.Logger.WithError(err),
			Msg: "Failed to list downloaded files",
			Err: err,
		}
	}

	pluginDir := filepath.Join(appconfig.GetConfigDir(), "plugins")
	for _, file := range files {
		fp := filepath.Join(tmpDir, file.Name())
		buf, err := ioutil.ReadFile(fp)
		if err != nil {
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: "Failed to read file downloaded in release",
				Err: err,
			}
		}
		kind, err := filetype.Match(buf)
		if err != nil {
			return errorsutil.EiamError{
				Log: util.Logger.WithError(err),
				Msg: "Failed to determine MIME type of file downloaded in release",
				Err: err,
			}
		}
		if kind.MIME.Value == "application/x-executable" {
			if err := util.MoveFile(fp, filepath.Join(pluginDir, file.Name())); err != nil {
				return errorsutil.EiamError{
					Log: util.Logger.WithError(err),
					Msg: "Failed to move plugin binary to plugins directory",
					Err: err,
				}
			}
		}
	}
	return nil
}
