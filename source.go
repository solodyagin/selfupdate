package selfupdate

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

// Source define an interface that is able to get an update
type Source interface {
	Get(*Version) (io.ReadCloser, int64, error) // Get the executable to be updated to
	GetSignature() ([64]byte, error)            // Get the signature that match the executable
	LatestVersion() (*Version, error)           // Get the latest version information to determine if we should trigger an update
}

func replaceURLTemplate(base string) string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	p := platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Ext:  ext,
	}

	exe, err := ExecutableRealPath()
	if err != nil {
		exe = filepath.Base(os.Args[0])
	} else {
		exe = filepath.Base(exe)
	}
	if runtime.GOOS == "windows" {
		p.Executable = exe[:len(exe)-len(".exe")]
	} else {
		p.Executable = exe
	}

	t, err := template.New("platform").Parse(base)
	if err != nil {
		return base
	}

	buf := &strings.Builder{}
	err = t.Execute(buf, p)
	if err != nil {
		return base
	}
	return buf.String()
}
