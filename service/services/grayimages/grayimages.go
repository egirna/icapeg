package grayimages

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

func (g Grayimages) Processing(b bool) (int, interface{}, map[string]string) {
	//TODO implement me
	panic("implement me")
}

func (g *Grayimages) SendFileToAPI(f *bytes.Buffer, fileType string, fileName string) (*http.Response, error) {
	var url string
	switch fileType {
	case "png":
		url = g.BaseURL + "/png"
	case "webp":
		url = g.BaseURL + "/webp"
	case "jpeg":
		url = g.BaseURL + "/jpeg"
	case "jpg":
		url = g.BaseURL + "/jpeg"
	}
	log.Println("113, url: ", url)
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part1,
		errFile1 := writer.CreateFormFile("img", fileName)
	_, errFile1 = io.Copy(part1, bytes.NewReader(f.Bytes()))
	if errFile1 != nil {
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *Grayimages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
