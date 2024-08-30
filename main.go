package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "github.com/khouwdevin/uploadthing-go/types"
)

// basic types

type UTApi struct {
	config *UTApiConfig
	client *http.Client
}

type UTApiConfig struct {
	secret  string
	baseUrl string
}

// create new utapi instance

func NewUTApi(userSecret string) (*UTApi, error) {
	return &UTApi{
		config: &UTApiConfig{
			secret:  userSecret,
			baseUrl: "https://api.uploadthing.com/v6/",
		},
		client: &http.Client{},
	}, nil
}

// check api key

func (utapi *UTApi) checkAvailability() error {
	if len(utapi.config.secret) <= 0 {
		return errors.New("there is no secret attach")
	}

	return nil
}

// uploadFiles

func (utapi *UTApi) uploadFiles(files *[]multipart.FileHeader) (string, error) {
	var fileArr []FileForUpload

	for _, file := range *files {
		name := file.Filename
		size := file.Size
		fileType := file.Header["Content-Type"][0]

		fileArr = append(fileArr, FileForUpload{name, size, fileType})
	}

	fileMarshal, err := json.Marshal(UploadFilesType{fileArr, "public-read", "inline"})

	if err != nil {
		return "error", err
	}

	payload := bytes.NewBuffer(fileMarshal)

	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"uploadFiles", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "error", err
	}

	if res.StatusCode != 200 {
		return "error", errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	var uploadInfos []UploadInfo

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "error", err
	}

	err = json.Unmarshal(body, &uploadInfos)

	if err != nil {
		return "error", err
	}

	multipartBody := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartBody)

	for index, uploadInfo := range uploadInfos {
		for key, value := range uploadInfo.Fields {
			writer.WriteField(key, value)
		}
		part, err := writer.CreateFormFile("file", fileArr[index].Name)

		if err != nil {
			return "error create form file", err
		}

		file, err := os.Open(fileArr[index].Name)
		if err != nil {
			return "error open file", err
		}
		defer file.Close()

		_, err = io.Copy(part, file)
		if err != nil {
			return "error copy file", err
		}

		err = writer.Close()
		if err != nil {
			return "error close multipart", err
		}

		req, err := http.NewRequest("POST", utapi.config.baseUrl+uploadInfo.URL, multipartBody)
		if err != nil {
			return "error create request", err
		}

		req.Header.Set("Content-Type", "application/xml")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "error do http client", err
		}

		if res.StatusCode != 200 {
			return "error has occured " + strconv.Itoa(res.StatusCode), err
		}

		defer res.Body.Close()

		fmt.Println("Response status:", res.StatusCode)
	}

	return "hello", nil
}

// listFiles

func (utapi *UTApi) listFiles() ([]ListFilesType, error) {
	var filelistRequest ListFilesRequest
	err := utapi.checkAvailability()

	if err != nil {
		return []ListFilesType{}, err
	}

	payload := strings.NewReader("{}")
	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"listFiles", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return []ListFilesType{}, err
	}

	if res.StatusCode != 200 {
		return []ListFilesType{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []ListFilesType{}, err
	}

	err = json.Unmarshal([]byte(body), &filelistRequest)

	if err != nil {
		return []ListFilesType{}, err
	}

	return filelistRequest.Files, nil
}

// deleteFiles

func (utapi *UTApi) deleteFiles(files []string) error {
	err := utapi.checkAvailability()

	if err != nil {
		return err
	}

	keyArr, err := json.Marshal(FileKeys{FileKeys: files})

	if err != nil {
		return err
	}

	payload := bytes.NewBuffer(keyArr)

	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"deleteFiles", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	return nil
}

// getUsageInfo

func (utapi *UTApi) getUsageInfo() (UsageInfo, error) {
	var usageInfo UsageInfo
	err := utapi.checkAvailability()

	if err != nil {
		return UsageInfo{}, err
	}

	payload := strings.NewReader("{}")

	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"getUsageInfo", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return UsageInfo{}, err
	}

	if res.StatusCode != 200 {
		return UsageInfo{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return UsageInfo{}, err
	}

	err = json.Unmarshal(body, &usageInfo)

	if err != nil {
		return UsageInfo{}, err
	}

	return usageInfo, err
}

func main() {
	utapi, err := NewUTApi("")

	if err != nil {
		fmt.Println("error")
	}

	usage, err := utapi.getUsageInfo()

	if err != nil {
		fmt.Print(err)
	}

	fmt.Print(usage)
}
