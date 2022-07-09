package workflow

import (
	"bytes"
	"io"
)

// NewRespBodyFromBytes creates an io.ReadCloser from a byte slice
// that is suitable for use as an http response body.
func NewRespBodyFromBytes(body []byte) io.ReadCloser {
	return &dummyReadCloser{body: bytes.NewReader(body)}
}

type dummyReadCloser struct {
	body io.ReadSeeker
}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	return d.body.Read(p)
}

func (d *dummyReadCloser) Close() error {
	_, _ = d.body.Seek(0, io.SeekEnd)
	return nil
}
