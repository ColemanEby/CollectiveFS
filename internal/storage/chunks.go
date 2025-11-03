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

	// First pass: collect all chunk files and their numbers
	type chunkEntry struct {
		chunkNum int
		path     string
		name     string
	}
	var chunkFiles []chunkEntry

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

		chunkFiles = append(chunkFiles, chunkEntry{
			chunkNum: chunkNum,
			path:     filepath.Join(fileFolder, entry.Name()),
			name:     entry.Name(),
		})
	}

	// Sort chunks by number to process in order
	for i := 0; i < len(chunkFiles)-1; i++ {
		for j := i + 1; j < len(chunkFiles); j++ {
			if chunkFiles[i].chunkNum > chunkFiles[j].chunkNum {
				chunkFiles[i], chunkFiles[j] = chunkFiles[j], chunkFiles[i]
			}
		}
	}

	// Second pass: process chunks in order and build chunk array
	chunks := make([]map[string]interface{}, 0)
	maxChunkNum := 0
	if len(chunkFiles) > 0 {
		maxChunkNum = chunkFiles[len(chunkFiles)-1].chunkNum
	}

	for _, cf := range chunkFiles {
		chunkID := uuid.New().String()
		chunkInfo := map[string]interface{}{
			"num":    cf.chunkNum,
			"id":     chunkID,
			"path":   cf.path,
			"offer":  map[string]interface{}{},
			"answer": map[string]interface{}{},
		}

		// Encrypt chunk
		if err := encryptChunk(cf.path); err != nil {
			return fmt.Errorf("failed to encrypt chunk %d: %w", cf.chunkNum, err)
		}

		// Insert chunk at correct position (matching Python's insert behavior)
		if cf.chunkNum >= len(chunks) {
			// Extend slice to accommodate this chunk number
			for len(chunks) <= cf.chunkNum {
				chunks = append(chunks, nil)
			}
		}
		chunks[cf.chunkNum] = chunkInfo

		// Call callback for chunk transfer (matching Python threading behavior)
		if onChunkReady != nil {
			onChunkReady(fileInfo, chunkInfo)
		}
	}

	// Remove nil entries (if any chunks were skipped)
	filteredChunks := make([]map[string]interface{}, 0, len(chunks))
	for i := 0; i <= maxChunkNum; i++ {
		if i < len(chunks) && chunks[i] != nil {
			filteredChunks = append(filteredChunks, chunks[i])
		}
	}
	chunks = filteredChunks

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
