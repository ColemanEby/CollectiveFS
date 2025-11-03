package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the cross-platform configuration directory.
// Uses os.UserConfigDir() if available, otherwise falls back to platform-specific defaults.
func GetConfigDir(customDir string) string {
	if customDir != "" {
		return customDir
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Last resort: current directory
			return ".collective"
		}
		configDir = homeDir
	}

	// Append CollectiveFS directory
	if runtime.GOOS == "windows" {
		return filepath.Join(configDir, "CollectiveFS")
	}
	return filepath.Join(configDir, "collective")
}

// GetConfigFile returns the path to the config file.
func GetConfigFile(configDir string) string {
	return filepath.Join(configDir, "config")
}

// GetKeyFile returns the path to the encryption key file.
func GetKeyFile(configDir string) string {
	return filepath.Join(configDir, "key")
}

// GetCollectivePath returns the path to the .collective directory inside rootPath.
func GetCollectivePath(rootPath string) string {
	return filepath.Join(rootPath, ".collective")
}

// GetProcessPath returns the path to the proc directory.
func GetProcessPath(rootPath string) string {
	return filepath.Join(rootPath, ".collective", "proc")
}

// GetCachePath returns the path to the cache directory.
func GetCachePath(rootPath string) string {
	return filepath.Join(rootPath, ".collective", "cache")
}

// GetPublicPath returns the path to the public directory.
func GetPublicPath(rootPath string) string {
	return filepath.Join(rootPath, ".collective", "public")
}

// GetTreeFilePath returns the path to the tree file.
func GetTreeFilePath(rootPath string) string {
	return filepath.Join(rootPath, ".collective", "tree")
}

