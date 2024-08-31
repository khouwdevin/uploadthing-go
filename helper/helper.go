package helper

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/khouwdevin/uploadthing-go/types"
)

func uploadSinglePart(file *multipart.FileHeader, uploadInfo *types.UploadInfo) (types.File, error) {
	multipartBody := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartBody)

	for key, value := range uploadInfo.Fields {
		writer.WriteField(key, value)
	}

	err := addFormFile(writer, file)
	if err != nil {
		return types.File{}, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.uploadthing.com/"+uploadInfo.URL, multipartBody)
	if err != nil {
		return types.File{}, err
	}

	req.Header.Set("Accept", "application/xml")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.File{}, err
	}

	if res.StatusCode != 200 {
		return types.File{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	return types.File{
		FileName: uploadInfo.FileName,
		FileType: uploadInfo.FileType, FileUrl: uploadInfo.FileType}, err
}

func createRequest(path string, secret string, body io.Reader) (req *http.Request) {
	req, _ = http.NewRequest(http.MethodPost, path, body)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", secret)

	return req
}

func addFormFile(writer *multipart.Writer, file *multipart.FileHeader) error {
	part, err := writer.CreateFormFile("file", file.Filename)

	if err != nil {
		return err
	}

	fileTmp, err := file.Open()
	if err != nil {
		return err
	}
	defer fileTmp.Close()

	_, err = io.Copy(part, fileTmp)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}
