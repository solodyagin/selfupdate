package selfupdate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// HTTPSource provide a Source that will download the update from a HTTP url.
// It is expecting the signature file to be served at ${URL}.ed25519
type HTTPSource struct {
	client  *http.Client
	baseURL string
}

var _ Source = (*HTTPSource)(nil)

type platform struct {
	OS         string
	Arch       string
	Ext        string
	Executable string
}

// NewHTTPSource provide a selfupdate.Source that will fetch the specified base URL
// for update and signature using the http.Client provided. To help into providing
// cross platform application, the base is actually a Go Template string where the
// following parameter are recognized:
// {{.OS}} will be filled by the runtime OS name
// {{.Arch}} will be filled by the runtime Arch name
// {{.Ext}} will be filled by the executable expected extension for the OS
// As an example the following string `http://localhost/myapp-{{.OS}}-{{.Arch}}{{.Ext}}`
// would fetch on Windows AMD64 the following URL: `http://localhost/myapp-windows-amd64.exe`
// and on Linux AMD64: `http://localhost/myapp-linux-amd64`.
func NewHTTPSource(client *http.Client, base string) Source {
	if client == nil {
		client = http.DefaultClient
	}

	base = replaceURLTemplate(base)

	return &HTTPSource{client: client, baseURL: base}
}

// Get will return if it succeed an io.ReaderCloser to the new executable being downloaded and its length
func (h *HTTPSource) Get(v *Version) (io.ReadCloser, int64, error) {
	request, err := http.NewRequest("GET", h.baseURL, nil)
	if err != nil {
		return nil, 0, err
	}

	if v != nil && !v.Date.IsZero() {
		request.Header.Add("If-Modified-Since", v.Date.Format(http.TimeFormat))

	}

	response, err := h.client.Do(request)
	if err != nil {
		return nil, 0, err
	}

	return response.Body, response.ContentLength, nil
}

// GetSignature will return the content of  ${URL}.ed25519
func (h *HTTPSource) GetSignature() ([64]byte, error) {
	resp, err := h.client.Get(h.baseURL + ".ed25519")
	if err != nil {
		return [64]byte{}, err
	}
	defer resp.Body.Close()

	if resp.ContentLength != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", resp.ContentLength)
	}

	writer := bytes.NewBuffer(make([]byte, 0, 64))
	n, err := io.Copy(writer, resp.Body)
	if err != nil {
		return [64]byte{}, err
	}

	if n != 64 {
		return [64]byte{}, fmt.Errorf("ed25519 signature must be 64 bytes long and was %v", n)
	}

	r := [64]byte{}
	copy(r[:], writer.Bytes())

	return r, nil
}

// LatestVersion will return the URL Last-Modified time
func (h *HTTPSource) LatestVersion() (*Version, error) {
	resp, err := h.client.Head(h.baseURL)
	if err != nil {
		return nil, err
	}

	lastModified := resp.Header.Get("Last-Modified")
	if lastModified == "" {
		return nil, errors.New("no Last-Modified served")
	}

	t, err := http.ParseTime(lastModified)
	if err != nil {
		return nil, err
	}

	return &Version{Date: t}, nil
}
