package config

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	ConfigDir      string
	ConfigFile     string
	KeyFile        string
	RootPath       string
	CollectivePath string
	ProcessPath    string
	CachePath      string
	PublicPath     string
	TreeFilePath   string
	ProgramPath    string
}

// LoadConfig loads configuration from the config directory.
// If customConfigDir is provided, it overrides the default.
func LoadConfig(customConfigDir string) (*Config, error) {
	configDir := GetConfigDir(customConfigDir)
	cfg := &Config{
		ConfigDir:  configDir,
		ConfigFile: GetConfigFile(configDir),
		KeyFile:    GetKeyFile(configDir),
	}

	// Get program path (where the executable is located)
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		execPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to determine program path: %w", err)
		}
	}
	cfg.ProgramPath = filepath.Dir(execPath)

	// Read root path from config file
	rootPath, err := cfg.readRootPath()
	if err != nil {
		return nil, fmt.Errorf("failed to read root path: %w", err)
	}
	cfg.RootPath = rootPath

	// Set up derived paths
	cfg.CollectivePath = GetCollectivePath(rootPath)
	cfg.ProcessPath = GetProcessPath(rootPath)
	cfg.CachePath = GetCachePath(rootPath)
	cfg.PublicPath = GetPublicPath(rootPath)
	cfg.TreeFilePath = GetTreeFilePath(rootPath)

	// Create directories if they don't exist
	if err := cfg.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return cfg, nil
}

// readRootPath reads the root path from the config file.
// The config file contains a single line with the path to watch.
func (c *Config) readRootPath() (string, error) {
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return "", fmt.Errorf("config file not found or unreadable: %w", err)
	}

	// Read first line, trim whitespace
	rootPath := filepath.Clean(string(data))
	if len(rootPath) == 0 {
		return "", fmt.Errorf("config file is empty")
	}

	// Verify the path exists
	if _, err := os.Stat(rootPath); err != nil {
		return "", fmt.Errorf("root path does not exist: %s: %w", rootPath, err)
	}

	return rootPath, nil
}

// ensureDirectories creates all necessary directories.
func (c *Config) ensureDirectories() error {
	dirs := []string{
		c.ConfigDir,
		c.CollectivePath,
		c.ProcessPath,
		c.CachePath,
		c.PublicPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LoadOrGenerateKey loads the encryption key from the key file, or generates a new one if it doesn't exist.
// Returns the key as a byte slice (32 bytes for Fernet).
func (c *Config) LoadOrGenerateKey() ([]byte, error) {
	// Check if key file exists
	if _, err := os.Stat(c.KeyFile); err == nil {
		// Key exists, load it
		key, err := os.ReadFile(c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("key file has invalid size: expected 32 bytes, got %d", len(key))
		}
		return key, nil
	}

	// Generate new key (Fernet uses URL-safe base64 encoded 32-byte key)
	// But we store it as raw 32 bytes to match Python implementation
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Write key to file
	if err := os.WriteFile(c.KeyFile, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}

	return key, nil
}

// GetEncoderPath returns the path to the encoder binary.
func (c *Config) GetEncoderPath() string {
	return filepath.Join(c.ProgramPath, "lib", "encoder")
}
