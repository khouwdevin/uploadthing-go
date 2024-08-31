package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/khouwdevin/uploadthing-go/types"
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

func (utapi *UTApi) uploadFiles(files *[]multipart.FileHeader) ([]types.File, error) {
	var fileArr []types.FileForUpload

	for _, file := range *files {
		name := file.Filename
		size := file.Size
		fileType := file.Header["Content-Type"][0]

		fileArr = append(fileArr, types.FileForUpload{Name: name, Size: size, Type: fileType})
	}

	fileMarshal, err := json.Marshal(types.UploadFilesType{Files: fileArr, Acl: "public-read", ContentDisposition: "inline"})

	if err != nil {
		return []types.File{}, err
	}

	payload := bytes.NewBuffer(fileMarshal)

	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"uploadFiles", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return []types.File{}, err
	}

	if res.StatusCode != 200 {
		return []types.File{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	var uploadInfos []types.UploadInfo

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []types.File{}, err
	}

	err = json.Unmarshal(body, &uploadInfos)

	if err != nil {
		return []types.File{}, err
	}

	multipartBody := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartBody)

	var fileInfos []types.File

	for index, file := range *files {
		for key, value := range uploadInfos[index].Fields {
			writer.WriteField(key, value)
		}
		part, err := writer.CreateFormFile("file", file.Filename)

		if err != nil {
			return []types.File{}, err
		}

		fileTmp, err := file.Open()
		if err != nil {
			return []types.File{}, err
		}
		defer fileTmp.Close()

		_, err = io.Copy(part, fileTmp)
		if err != nil {
			return []types.File{}, err
		}

		err = writer.Close()
		if err != nil {
			return []types.File{}, err
		}

		req, err := http.NewRequest(http.MethodPost, utapi.config.baseUrl+uploadInfos[index].URL, multipartBody)
		if err != nil {
			return []types.File{}, err
		}

		req.Header.Set("Accept", "application/xml")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return []types.File{}, err
		}

		if res.StatusCode != 200 {
			return []types.File{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
		}

		defer res.Body.Close()

		fileInfos = append(fileInfos, types.File{
			FileName: uploadInfos[index].FileName,
			FileType: uploadInfos[index].FileType, FileUrl: uploadInfos[index].FileType})
	}

	return fileInfos, nil
}

// listFiles

func (utapi *UTApi) listFiles() ([]types.ListFilesType, error) {
	var filelistRequest types.ListFilesRequest
	err := utapi.checkAvailability()

	if err != nil {
		return []types.ListFilesType{}, err
	}

	payload := strings.NewReader("{}")
	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"listFiles", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return []types.ListFilesType{}, err
	}

	if res.StatusCode != 200 {
		return []types.ListFilesType{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []types.ListFilesType{}, err
	}

	err = json.Unmarshal([]byte(body), &filelistRequest)

	if err != nil {
		return []types.ListFilesType{}, err
	}

	return filelistRequest.Files, nil
}

// deleteFiles

func (utapi *UTApi) deleteFiles(files []string) error {
	err := utapi.checkAvailability()

	if err != nil {
		return err
	}

	keyArr, err := json.Marshal(types.FileKeys{FileKeys: files})

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

func (utapi *UTApi) getUsageInfo() (types.UsageInfo, error) {
	var usageInfo types.UsageInfo
	err := utapi.checkAvailability()

	if err != nil {
		return types.UsageInfo{}, err
	}

	payload := strings.NewReader("{}")

	req, _ := http.NewRequest(http.MethodPost, utapi.config.baseUrl+"getUsageInfo", payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-uploadthing-api-key", utapi.config.secret)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return types.UsageInfo{}, err
	}

	if res.StatusCode != 200 {
		return types.UsageInfo{}, errors.New("error has occured " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return types.UsageInfo{}, err
	}

	err = json.Unmarshal(body, &usageInfo)

	if err != nil {
		return types.UsageInfo{}, err
	}

	return usageInfo, err
}

// renameFile

func (utapi *UTApi) renameFile(key string, newName string) (string, error) {
	return "hello", nil
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
