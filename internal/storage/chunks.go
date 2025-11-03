package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ProcessChunks encrypts chunks in a file folder and creates chunk metadata.
// This matches Python's encryptChunks function.
func ProcessChunks(fileFolder string, fileInfo map[string]interface{}, encryptChunk func(chunkPath string) error, onChunkReady func(fileInfo map[string]interface{}, chunkInfo map[string]interface{})) error {
	// Read directory
	entries, err := os.ReadDir(fileFolder)
	if err != nil {
		return fmt.Errorf("failed to read file folder: %w", err)
	}

	chunks := []map[string]interface{}{}

	// Process each file in the folder
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Extract chunk number from filename (format: filename.ext.N)
		// Matching Python: chunkNum = int(os.path.splitext(filename.name)[1][1:])
		ext := filepath.Ext(entry.Name())
		if len(ext) < 2 {
			continue
		}
		chunkNumStr := ext[1:] // Remove the dot
		chunkNum, err := strconv.Atoi(chunkNumStr)
		if err != nil {
			continue // Skip files without numeric extension
		}

		chunkPath := filepath.Join(fileFolder, entry.Name())
		chunkID := uuid.New().String()

		chunkInfo := map[string]interface{}{
			"num":    chunkNum,
			"id":     chunkID,
			"path":   chunkPath,
			"offer":  map[string]interface{}{},
			"answer": map[string]interface{}{},
		}

		// Encrypt chunk
		if err := encryptChunk(chunkPath); err != nil {
			return fmt.Errorf("failed to encrypt chunk %d: %w", chunkNum, err)
		}

		// Insert chunk at correct position (matching Python's insert behavior)
		// Since we're processing in order, append is fine, but Python uses insert
		if chunkNum >= len(chunks) {
			// Extend slice
			for len(chunks) <= chunkNum {
				chunks = append(chunks, nil)
			}
		}
		chunks[chunkNum] = chunkInfo

		// Call callback for chunk transfer (matching Python threading behavior)
		if onChunkReady != nil {
			onChunkReady(fileInfo, chunkInfo)
		}
	}

	// Update fileInfo with chunks
	fileInfo["chunks"] = chunks

	return nil
}

// ExtractChunkNumber extracts chunk number from filename.
func ExtractChunkNumber(filename string) (int, error) {
	ext := filepath.Ext(filename)
	if len(ext) < 2 {
		return 0, fmt.Errorf("invalid chunk filename: %s", filename)
	}
	chunkNumStr := strings.TrimPrefix(ext, ".")
	return strconv.Atoi(chunkNumStr)
}
