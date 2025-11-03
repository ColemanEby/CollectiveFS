package encoder

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

// Encoder handles Reed-Solomon encoding via the lib/encoder binary.
type Encoder struct {
	encoderPath string
}

// NewEncoder creates a new encoder instance.
func NewEncoder(encoderPath string) *Encoder {
	return &Encoder{
		encoderPath: encoderPath,
	}
}

// EncodeFile encodes a file using the Go encoder binary, matching Python's subprocess.run() behavior.
func (e *Encoder) EncodeFile(filePath string, dataChunks, parChunks int, outDir string) error {
	// Build command matching Python's encoderCmd format
	cmd := exec.Command(
		e.encoderPath,
		"--data", fmt.Sprintf("%d", dataChunks),
		"--par", fmt.Sprintf("%d", parChunks),
		"--out", outDir,
		filePath,
	)

	// Execute command (matching Python's subprocess.run() - no output capture)
	log.Printf("Running encoder: %v", cmd.Args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoder failed: %w", err)
	}

	return nil
}

// EncodeFileFromInfo encodes a file using information from the fileInfo map.
// This matches the Python workflow where encoder command is built from fileInfo.
func (e *Encoder) EncodeFileFromInfo(fileInfo map[string]interface{}) error {
	// Extract information from fileInfo
	encoderCmd, ok := fileInfo["encoder_cmd"].([]string)
	if !ok {
		return fmt.Errorf("invalid encoder_cmd in fileInfo")
	}

	filePath, ok := fileInfo["file_path"].(string)
	if !ok {
		return fmt.Errorf("invalid file_path in fileInfo")
	}

	// Build command
	cmd := exec.Command(encoderCmd[0], encoderCmd[1:]...)

	// Execute (matching Python behavior - no output parsing)
	log.Printf("Encoding file: %s", filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encoder failed for %s: %w", filePath, err)
	}

	return nil
}

// GetEncoderPath returns the full path to the encoder binary.
func GetEncoderPath(programPath string) string {
	return filepath.Join(programPath, "lib", "encoder")
}
