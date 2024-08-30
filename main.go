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
)

type UTApi struct {
	config *UTApiConfig
	client *http.Client
}

type UTApiConfig struct {
	secret  string
	baseUrl string
}

type FileKeys struct {
	FileKeys []string `json:"fileKeys"`
}

type ListFilesType struct {
	Id       string `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	CustomId string `json:"customId"`
	Status   string `json:"status"`
}

type ListFilesRequest struct {
	HasMore bool            `json:"hasMore"`
	Files   []ListFilesType `json:"files"`
}

type UsageInfo struct {
	TotalBytes    int `json:"totalBytes"`
	AppTotalBytes int `json:"appTotalBytes"`
	FilesUploaded int `json:"filesUploaded"`
	LimitBytes    int `json:"limitBytes"`
}

type UploadFilesType struct {
	Files              []FileForUpload `json:"files"`
	Acl                string          `json:"acl"`
	ContentDisposition string          `json:"contentDisposition"`
}

type FileForUpload struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

type UploadInfo struct {
	URL                string            `json:"url"`
	Fields             map[string]string `json:"fields"`
	Key                string            `json:"key"`
	ContentDisposition string            `json:"contentDisposition"`
	FileUrl            string            `json:"fileUrl"`
	AppUrl             string            `json:"appUrl"`
	FileName           string            `json:"fileName"`
	PollingUrl         string            `json:"pollingUrl"`
	PollingJwt         string            `json:"pollingJwt"`
	FileType           string            `json:"fileType"`
	CustomId           interface{}       `json:"customId"`
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
		writer.CreateFormFile("file", fileArr[index].Name)

		part, err := writer.CreateFormFile("file", "your_file.txt")
		if err != nil {
			fmt.Println("Error creating form file:", err)
		}

		file, err := os.Open(files[index])
		if err != nil {
			fmt.Println("Error opening file:", err)
		}
		defer file.Close()

		_, err = io.Copy(part, files[index])
		if err != nil {
			fmt.Println("Error writing file to form data:", err)
		}

		err = writer.Close()
		if err != nil {
			fmt.Println("Error closing multipart writer:", err)
		}

		req, err := http.NewRequest("POST", utapi.config.baseUrl+uploadInfo.URL, multipartBody)
		if err != nil {
			fmt.Println("Error creating HTTP request:", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error sending HTTP request:", err)
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
