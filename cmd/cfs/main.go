package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/collectivefs/cfs/internal/config"
	"github.com/collectivefs/cfs/internal/encoder"
	"github.com/collectivefs/cfs/internal/encryption"
	"github.com/collectivefs/cfs/internal/logger"
	"github.com/collectivefs/cfs/internal/server"
	"github.com/collectivefs/cfs/internal/storage"
	"github.com/collectivefs/cfs/internal/transfer"
	"github.com/collectivefs/cfs/internal/watcher"
)

const (
	hostName   = "localhost"
	portNumber = 8080
)

func main() {
	// Parse command line arguments (matching Python argparse)
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	verboseShort := flag.Bool("v", false, "Enable verbose logging (short)")
	version := flag.Bool("version", false, "Show version")
	input := flag.String("input", "", "Enter source directory to watch")
	_ = flag.String("output", "", "Enter the directory to copy to") // Reserved for future use
	_ = flag.Bool("service", false, "Run continuously")             // Reserved for future use
	configDir := flag.String("config-dir", "", "Configuration directory (overrides default)")

	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Println("CollectiveFS v1.0.0 (Go)")
		os.Exit(0)
	}

	// Enable verbose if either flag is set
	if *verbose || *verboseShort {
		logger.Init(true)
	} else {
		logger.Init(false)
	}

	logger.Info("Starting CollectiveFS (Go)")

	// Load configuration
	cfg, err := config.LoadConfig(*configDir)
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// Override root path if --input is provided
	if *input != "" {
		cfg.RootPath = *input
		cfg.CollectivePath = config.GetCollectivePath(cfg.RootPath)
		cfg.ProcessPath = config.GetProcessPath(cfg.RootPath)
		cfg.CachePath = config.GetCachePath(cfg.RootPath)
		cfg.PublicPath = config.GetPublicPath(cfg.RootPath)
		cfg.TreeFilePath = config.GetTreeFilePath(cfg.RootPath)

		// Ensure directories exist
		if err := os.MkdirAll(cfg.CollectivePath, 0755); err != nil {
			logger.Fatal("Failed to create directories: %v", err)
		}
		if err := os.MkdirAll(cfg.ProcessPath, 0755); err != nil {
			logger.Fatal("Failed to create directories: %v", err)
		}
		if err := os.MkdirAll(cfg.CachePath, 0755); err != nil {
			logger.Fatal("Failed to create directories: %v", err)
		}
		if err := os.MkdirAll(cfg.PublicPath, 0755); err != nil {
			logger.Fatal("Failed to create directories: %v", err)
		}
	}

	// Load or generate encryption key
	key, err := cfg.LoadOrGenerateKey()
	if err != nil {
		logger.Fatal("Failed to load or generate key: %v", err)
	}

	if len(key) > 0 {
		if _, err := os.Stat(cfg.KeyFile); err == nil {
			logger.Info("Found key.")
		} else {
			logger.Info("Creating new key.")
		}
	}

	// Initialize encryption
	fernet, err := encryption.NewFernetEncryption(key)
	if err != nil {
		logger.Fatal("Failed to initialize encryption: %v", err)
	}

	// Initialize encoder
	enc := encoder.NewEncoder(cfg.GetEncoderPath())

	// Start HTTP server (matching Python's run_server)
	server.RunServer(hostName, portNumber, cfg.ProgramPath, cfg.ConfigDir)

	// File processing callback - matches Python's finishedEncoding -> encryptChunks workflow
	onFileReady := func(fileInfo interface{}) {
		fileInfoMap, ok := fileInfo.(map[string]interface{})
		if !ok {
			logger.Error("Invalid fileInfo type")
			return
		}

		// Extract encoder command and execute encoding
		filePath, _ := fileInfoMap["file_path"].(string)
		fileFolder, _ := fileInfoMap["folder"].(string)

		logger.Info("Processing file: %s", filePath)

		// Execute encoder (matching Python subprocess.run())
		if err := enc.EncodeFileFromInfo(fileInfoMap); err != nil {
			logger.Error("Encoder failed: %v", err)
			return
		}

		// After encoding, encrypt chunks (matching Python's encryptChunks)
		err := storage.ProcessChunks(
			fileFolder,
			fileInfoMap,
			func(chunkPath string) error {
				// Encrypt chunk (matching Python's encryptChunk)
				if err := fernet.EncryptChunk(chunkPath); err != nil {
					return fmt.Errorf("failed to encrypt chunk: %w", err)
				}
				logger.Debug("Encrypted chunk: %s", chunkPath)
				return nil
			},
			func(fileInfo, chunkInfo map[string]interface{}) {
				// Start transfer for each chunk (matching Python's transfer.start_transfer)
				// This runs in the background like Python threading
				go func() {
					if err := transfer.StartTransfer("send", fileInfo, chunkInfo); err != nil {
						logger.Error("Transfer failed: %v", err)
					}
				}()
			},
		)

		if err != nil {
			logger.Error("Failed to process chunks: %v", err)
			return
		}

		// Store file metadata for later retrieval
		if err := storage.StoreFileMetadata(cfg.CachePath, fileInfoMap); err != nil {
			logger.Error("Failed to store file metadata: %v", err)
		} else {
			logger.Info("Stored metadata for file: %s (ID: %s)", filePath, fileInfoMap["id"])
		}

		// Print file info as JSON in debug mode (matching Python's json.dumps)
		if jsonStr, err := json.MarshalIndent(fileInfoMap, "", "  "); err == nil {
			logger.Debug("File info: %s", string(jsonStr))
		}
	}

	// Create file watcher event handler
	eventHandler := watcher.NewModifiedDirHandler(
		cfg.RootPath,
		cfg.ProcessPath,
		cfg.ProgramPath,
		cfg.GetEncoderPath(),
		onFileReady,
	)

	// Create and start watcher
	fileWatcher, err := watcher.NewWatcher(cfg.RootPath, true, eventHandler)
	if err != nil {
		logger.Fatal("Failed to create watcher: %v", err)
	}
	defer fileWatcher.Close()

	fileWatcher.Start()
	logger.Info("File watcher started on: %s", cfg.RootPath)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Main loop (matching Python's while True: time.sleep(1))
	logger.Info("CollectiveFS running. Press Ctrl+C to stop.")
	for {
		select {
		case <-sigChan:
			logger.Info("Shutting down...")
			return
		default:
			time.Sleep(time.Second)
		}
	}
}
