package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ModifiedDirHandler handles file creation events, matching Python's ModifiedDirHandler class.
type ModifiedDirHandler struct {
	rootPath    string
	processPath string
	programPath string
	encoderPath string
	onFileReady func(fileInfo interface{}) // Callback for when encoding is finished
}

// NewModifiedDirHandler creates a new file event handler.
func NewModifiedDirHandler(rootPath, processPath, programPath, encoderPath string, onFileReady func(fileInfo interface{})) *ModifiedDirHandler {
	return &ModifiedDirHandler{
		rootPath:    rootPath,
		processPath: processPath,
		programPath: programPath,
		encoderPath: encoderPath,
		onFileReady: onFileReady,
	}
}

// OnCreated handles file creation events, matching Python's on_created method.
func (h *ModifiedDirHandler) OnCreated(filePath string) {
	// Calculate relative path
	filePathRel := strings.TrimPrefix(filePath, h.rootPath)
	filePathRel = strings.TrimPrefix(filePathRel, string(filepath.Separator))

	// Get file name
	fileName := filepath.Base(filePath)

	// Check if file is in .collective directory (already filtered by watcher, but double-check)
	if strings.Contains(filePath, ".collective") {
		return
	}

	// Process the file
	// Matching Python logic: numOfDataChunks = 64, numOfParChunks = 32
	numOfDataChunks := 64
	numOfParChunks := 32

	id := uuid.New()
	fileFolder := filepath.Join(h.processPath, filePathRel+".d")

	// Create file folder
	if err := os.MkdirAll(fileFolder, 0755); err != nil {
		log.Printf("failed to create file folder: %v", err)
		return
	}

	// Build encoder command
	// Format: encoder --data <num> --par <num> --out "<outDir>" "<inputFile>"
	encoderCmd := []string{
		h.encoderPath,
		"--data", strconv.Itoa(numOfDataChunks),
		"--par", strconv.Itoa(numOfParChunks),
		"--out", fileFolder,
		filePath,
	}

	// Execute encoder (this will be handled by encoder package)
	// For now, we'll pass the command to be executed
	// The encoder package will handle subprocess execution

	num := numOfParChunks + numOfDataChunks
	fileInfo := map[string]interface{}{
		"id":               id.String(),
		"name":             fileName,
		"folder":           fileFolder,
		"number_of_chunks": num,
		"chunks":           []interface{}{},
		"encoder_cmd":      encoderCmd, // Pass command info for encoder execution
		"file_path":        filePath,
	}

	// Call callback with file info (this will trigger encoding, then encryption)
	if h.onFileReady != nil {
		h.onFileReady(fileInfo)
	}
}

// makeFolder creates a directory if it doesn't exist (matching Python's makeFolder).
func makeFolder(path string) error {
	return os.MkdirAll(path, 0755)
}
