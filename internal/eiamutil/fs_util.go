package eiamutil

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func MoveFile(src, dst string) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(dst)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file.
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

func DownloadAndExtract(url, tmpDir string) error {
	Logger.Infof("Downloading new version from %s", url)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	Logger.Info("Successfully downloaded the archive, now extracting its contents")
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header.Typeflag == tar.TypeDir:
			target := filepath.Join(tmpDir, filepath.Clean(header.Name))
			if sErr := os.MkdirAll(target, 0o755); sErr != nil {
				return sErr
			}
		case header.Typeflag == tar.TypeReg:
			target := filepath.Join(tmpDir, filepath.Clean(header.Name))
			var f *os.File
			f, err = os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// Limit readable amount to 2GB to prevent decompression bomb.
			maxSize := 2 << (10 * 3)
			limiter := io.LimitReader(tarReader, int64(maxSize))
			if _, err = io.Copy(f, limiter); err != nil {
				return err
			}
			// Manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		default:
			err = fmt.Errorf("unknown type %v in %s", header.Typeflag, header.Name)
			return err
		}
	}
}
