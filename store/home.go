package store

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fioncat/go-gendb/misc/log"
)

var (
	basePath string
	baseOnce sync.Once
)

func baseHome() string {
	baseOnce.Do(func() {
		home, err := home()
		if err != nil {
			log.Errorf("fetch home dir failed: %v,"+
				" we will save file in current directoy.",
				err)
			return
		}
		basePath = filepath.Join(home, ".gogendb")
	})
	return basePath
}

func home() (string, error) {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir, nil
	}

	if runtime.GOOS == "windows" {
		return homeWindows()
	}

	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var out bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "", errors.New("empty home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are empty")
	}

	return home, nil
}
