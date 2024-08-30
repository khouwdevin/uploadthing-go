package types

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
