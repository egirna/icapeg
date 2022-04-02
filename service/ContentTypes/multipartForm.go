package ContentTypes

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
)

type FormPart struct {
	FormName string
	FileName string
	Content  []byte
}

type MultipartForm struct {
	formParts []FormPart
	theFile   FormPart
	boundary  string
}

// GetFileFromRequest is used for parsing the multipart form
func (m MultipartForm) GetFileFromRequest() *bytes.Buffer {
	return bytes.NewBuffer(m.theFile.Content)
}

// ParsingRequest is utility function to GetFileFromRequest function, and it's used for helping functions which are outside the pkg
// to initialize a new instance from MultipartForm struct
func ParsingRequest(req *http.Request) ([]FormPart, FormPart, string) {
	_, params, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
	mr := multipart.NewReader(req.Body, params["boundary"])
	boundary := params["boundary"]
	p, _ := mr.NextPart()
	var formParts []FormPart
	for p != nil {
		var thisPart FormPart
		slurp, _ := io.ReadAll(p)
		thisPart.FormName = p.FormName()
		thisPart.FileName = p.FileName()
		thisPart.Content = slurp
		formParts = append(formParts, thisPart)
		p, _ = mr.NextPart()
	}
	var theFile FormPart
	for i := 0; i < len(formParts); i++ {
		if formParts[i].FileName != "" {
			theFile = formParts[i]
			break
		}
	}
	return formParts, theFile, boundary
}

// BodyAfterScanning is used for returning the file to be written in the http request body
//by making a multipart form
func (m MultipartForm) BodyAfterScanning(bodyByte []byte) string {
	m.theFile.Content = bodyByte
	for i := 0; i < len(m.formParts); i++ {
		if m.formParts[i].FileName == m.theFile.FileName &&
			m.formParts[i].FormName == m.theFile.FormName {
			m.formParts[i] = m.theFile
			break
		}
	}
	return m.creatingMultipartForm()
}

// creatingMultipartFor is utility function to BodyAfterScanning function
func (m MultipartForm) creatingMultipartForm() string {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.SetBoundary(m.boundary)
	for i := 0; i < len(m.formParts); i++ {
		if m.formParts[i].FileName == "" {
			bodyWriter.WriteField(m.formParts[i].FormName, bytes.NewBuffer(m.formParts[i].Content).String())
		} else {
			part, _ := bodyWriter.CreateFormFile(m.formParts[i].FormName, m.formParts[i].FileName)
			io.Copy(part, bytes.NewReader(m.formParts[i].Content))
		}
	}
	bodyWriter.Close()
	return bodyBuf.String()
}

// NewMultipartForm is used for returning a new instance from MultipartForm struct
func NewMultipartForm(formParts []FormPart, theFile FormPart, boundary string) MultipartForm {
	return MultipartForm{formParts: formParts, theFile: theFile, boundary: boundary}
}
