package ContentTypes

import (
	"bytes"
	"encoding/base64"
	"regexp"
)

type RegularFile struct {
	buf     *bytes.Buffer
	encoded bool
}

// GetFileFromRequest is used For getting the file from the body which amy be encoded,
//so we check first if it's encoded or not, then we decode it if it's encoded or send it directly if it's not encoded
func (r RegularFile) GetFileFromRequest() *bytes.Buffer {
	r.encoded, _ = regexp.Match("^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$", []byte(r.buf.String()))
	if r.encoded {
		decodedFile, _ := base64.StdEncoding.DecodeString(r.buf.String())
		r.buf = bytes.NewBuffer(decodedFile)
	}
	return r.buf
}

// BodyAfterScanning is used for returning the file to be written in the http request body
//it encode the file before returning in case it came encoded from the request
func (r RegularFile) BodyAfterScanning(bodyByte []byte) string {
	if r.encoded {
		encodedFile := base64.StdEncoding.EncodeToString(bodyByte)
		return encodedFile
	}
	return string(bodyByte)
}

// NewRegularFile is used for returning a new instance from RegularFile struct
func NewRegularFile(buf *bytes.Buffer, encoded bool) RegularFile {
	return RegularFile{buf: buf, encoded: encoded}
}
