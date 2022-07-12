package types

type StorageRequest struct {
	Requirements []string                          `json:"requirements"`
	Providers    map[string]map[string]interface{} `json:"providers"`
	Hash         string                            `json:"hash"`
	FilePointer  string                            `json:"file_pointer"`
}

type UploadStatus int

const (
	Pending UploadStatus = iota
	InProgress
	Success
	Failed
)
