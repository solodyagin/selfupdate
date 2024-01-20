package selfupdate

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlwaysFindExecutableTime(t *testing.T) {
	_, err := lastModifiedExecutable()
	assert.Nil(t, err)
}

func TestAlwaysFindExecutable(t *testing.T) {
	exe, err := ExecutableRealPath()
	ext := filepath.Ext(exe)
	assert.Nil(t, err)
	assert.NotEmpty(t, exe)
	if runtime.GOOS == "windows" {
		assert.Equal(t, ".exe", ext)
	} else {
		assert.True(t, ext == ".test" || ext == "", fmt.Sprintf("Linux extesion not correct, got '%s'", ext))
	}
}

func TestAlwaysFindOldExecutable(t *testing.T) {
	exe, err := ExecutableDefaultOldPath()
	ext := filepath.Ext(exe)
	assert.Nil(t, err)
	assert.NotEmpty(t, exe)
	assert.Equal(t, ".old", ext)
}

func TestAlwaysFindNewExecutable(t *testing.T) {
	exe, err := ExecutableNewPath()
	ext := filepath.Ext(exe)
	assert.Nil(t, err)
	assert.NotEmpty(t, exe)
	assert.Equal(t, ".new", ext)
}
