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

// home/basePath
func baseHome() string {
	baseOnce.Do(func() {
		home, err := home()
		if err != nil {
			log.Fatal("read home directory failed: %v", err)
		}
		basePath = filepath.Join(home, ".gogendb")

		stat, err := os.Stat(basePath)
		if err != nil {
			err = os.MkdirAll(basePath, os.ModePerm)
			if err != nil {
				log.Fatal("mkdir for basepath failed: %v", err)
			}
			return
		}
		if !stat.IsDir() {
			log.Fatal(`basepath "%s" is a file, conflict `+
				`with storage requirements`, basePath)
		}
	})
	return basePath
}

// home on any os(windows, unix, linux)
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

// home on unix(linux)
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

// home on windows
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
