package workflow

import (
	"bytes"
	"io"
	"strings"
)

// NewRespBodyFromString creates an io.ReadCloser from a string that
// is suitable for use as an http response body.
//
// To pass the content of an existing file as body use httpmock.File as in:
//   httpmock.NewRespBodyFromString(httpmock.File("body.txt").String())
func NewRespBodyFromString(body string) io.ReadCloser {
	return &dummyReadCloser{orig: body}
}

// NewRespBodyFromBytes creates an io.ReadCloser from a byte slice
// that is suitable for use as an http response body.
//
// To pass the content of an existing file as body use httpmock.File as in:
//   httpmock.NewRespBodyFromBytes(httpmock.File("body.txt").Bytes())
func NewRespBodyFromBytes(body []byte) io.ReadCloser {
	return &dummyReadCloser{orig: body}
}

type dummyReadCloser struct {
	orig interface{}   // string or []byte
	body io.ReadSeeker // instanciated on demand from orig
}

// setup ensures d.body is correctly initialized.
func (d *dummyReadCloser) setup() {
	if d.body == nil {
		switch body := d.orig.(type) {
		case string:
			d.body = strings.NewReader(body)
		case []byte:
			d.body = bytes.NewReader(body)
		}
	}
}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	d.setup()
	return d.body.Read(p)
}

func (d *dummyReadCloser) Close() error {
	d.setup()
	d.body.Seek(0, io.SeekEnd) // nolint: errcheck
	return nil
}
