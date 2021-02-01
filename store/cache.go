package store

import (
	"fmt"
	"os"
	"path/filepath"
)

func WalkCache(walkFunc func(path string, info os.FileInfo) error) error {
	root := filepath.Join(baseHome(), "cache")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return walkFunc(path, info)
	})
	if os.IsNotExist(err) {
		return fmt.Errorf("no cache data!")
	}

	return err
}

func RemoveCache() error {
	path := filepath.Join(baseHome(), "cache")
	return os.RemoveAll(path)
}
