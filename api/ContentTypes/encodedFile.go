package ContentTypes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type EncodedFile struct {
	data map[string]interface{}
}

// GetFileFromRequest is used For getting the file from the body which encoded,
//then we decode it to make it able to be scanned
func (e EncodedFile) GetFileFromRequest() *bytes.Buffer {
	base64Code := fmt.Sprint(e.data["Base64"])
	decodedFile, _ := base64.StdEncoding.DecodeString(base64Code)
	return bytes.NewBuffer(decodedFile)
}

// BodyAfterScanning is used for encoding the file after scanning
//and returning it to be written in the http request body
func (e EncodedFile) BodyAfterScanning(bodyByte []byte) string {
	encodedFile := base64.StdEncoding.EncodeToString(bodyByte)
	e.data["Base64"] = encodedFile
	j, _ := json.Marshal(e.data)
	return string(j)
}

// NewEncodedFile is used for returning a new instance from EncodedFile struct
func NewEncodedFile(data map[string]interface{}) EncodedFile {
	return EncodedFile{data: data}
}
