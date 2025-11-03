package storage

import (
	"encoding/json"
)

// FileInfo represents file metadata, matching Python's fileInfo dict structure.
type FileInfo struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Folder         string      `json:"folder"`
	NumberOfChunks int         `json:"number_of_chunks"`
	Chunks         []ChunkInfo `json:"chunks"`
}

// ChunkInfo represents chunk metadata, matching Python's chunkInfo dict structure.
type ChunkInfo struct {
	Num    int                    `json:"num"`
	ID     string                 `json:"id"`
	Path   string                 `json:"path"`
	Offer  map[string]interface{} `json:"offer"`
	Answer map[string]interface{} `json:"answer"`
}

// ToJSON converts FileInfo to JSON string, matching Python's json.dumps() output.
func (f *FileInfo) ToJSON() (string, error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
