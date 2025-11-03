# CollectiveFS - Go Implementation

This is the Go migration of CollectiveFS, a distributed file system built on WebRTC, Reed-Solomon encoding, and encryption.

## Migration Status

This is an active migration from Python to Go. The original Python code is preserved in the `legacy/` directory.

## Building

```bash
# Initialize dependencies
go mod download

# Build the encoder (if not already built)
cd lib
go mod init encoder 2>/dev/null || true
go mod edit -replace github.com/klauspost/reedsolomon=../reedsolomon
go mod tidy
go build -o encoder encoder.go
cd ..

# Build CollectiveFS
go build -o cfs ./cmd/cfs
```

## Running

```bash
# Basic usage
./cfs

# With verbose logging
./cfs --verbose

# Specify config directory
./cfs --config-dir ~/.collective

# Watch a specific directory
./cfs --input /path/to/watch

# Help
./cfs --help
```

## Configuration

CollectiveFS uses a config directory (defaults to `~/.collective/` on Unix or `%APPDATA%\CollectiveFS\` on Windows):

- `config` - Single line file containing the path to watch
- `key` - Binary Fernet encryption key (32 bytes)

The program will create subdirectories:
- `.collective/proc/` - Processed/encoded chunks
- `.collective/cache/` - Cache directory
- `.collective/public/` - Public files

## Architecture

- `cmd/cfs/` - Main entry point
- `internal/config/` - Configuration and path management
- `internal/watcher/` - File system watching
- `internal/encoder/` - Reed-Solomon encoding integration
- `internal/encryption/` - Fernet encryption
- `internal/transfer/` - WebRTC peer-to-peer transfer
- `internal/server/` - HTTP server
- `internal/storage/` - File metadata and chunk management
- `lib/` - Go encoder binary (Reed-Solomon)

## Cross-Platform Compatibility

This implementation is designed to work on:
- Linux
- macOS
- Windows

All paths are handled in a cross-platform manner using Go's `filepath` package.

