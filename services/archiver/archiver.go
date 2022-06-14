package archiver

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Archive(archivePath, sourceDir string) error {
	err := os.MkdirAll(filepath.Dir(archivePath), 0755)
	if err != nil {
		return err
	}
	archive, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	writer := zip.NewWriter(archive)
	defer writer.Close()
	err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == archivePath {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		w, err := writer.Create(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		return err
	})
	return err
}
