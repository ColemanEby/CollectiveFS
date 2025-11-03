package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StoreFileMetadata saves file metadata to a JSON file in the cache directory.
// This allows the system to track processed files and their chunk information.
func StoreFileMetadata(cachePath string, fileInfo map[string]interface{}) error {
	// Extract file ID
	fileID, ok := fileInfo["id"].(string)
	if !ok {
		return fmt.Errorf("fileInfo missing 'id' field")
	}

	// Create metadata file path
	metadataFile := filepath.Join(cachePath, fileID+".json")

	// Ensure cache directory exists
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write metadata as JSON
	data, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal file metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// LoadFileMetadata loads file metadata from the cache directory.
func LoadFileMetadata(cachePath, fileID string) (map[string]interface{}, error) {
	metadataFile := filepath.Join(cachePath, fileID+".json")

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var fileInfo map[string]interface{}
	if err := json.Unmarshal(data, &fileInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return fileInfo, nil
}

// ListFileMetadata returns all file metadata IDs in the cache directory.
func ListFileMetadata(cachePath string) ([]string, error) {
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Extract ID from filename (remove .json extension)
			id := entry.Name()[:len(entry.Name())-5]
			ids = append(ids, id)
		}
	}

	return ids, nil
}
