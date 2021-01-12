package store

import (
	"os"
	"path/filepath"
	"strings"
)

func WalkCache(prefix string, walkFunc func(path string, info os.FileInfo) error) error {
	err := filepath.Walk(baseHome(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if !strings.HasPrefix(filename, "cache.") {
			return nil
		}

		if prefix != "" {
			if !strings.HasPrefix(filename, "cache."+prefix) {
				return nil
			}
		}
		return walkFunc(path, info)
	})

	return err
}
