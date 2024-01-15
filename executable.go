package selfupdate

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fynelabs/selfupdate/internal/osext"
)

func lastModifiedExecutable() (time.Time, error) {
	exe, err := GetExecutableRealPath()
	if err != nil {
		return time.Time{}, err
	}

	fi, err := os.Stat(exe)
	if err != nil {
		return time.Time{}, err
	}

	return fi.ModTime(), nil
}

func GetExecutableRealPath() (string, error) {
	exe, err := osext.Executable()
	if err != nil {
		return "", err
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return exe, nil
}
