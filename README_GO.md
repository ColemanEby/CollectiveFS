# CollectiveFS - Go Implementation

This is the Go migration of CollectiveFS, a distributed file system built on WebRTC, Reed-Solomon encoding, and encryption.

## Migration Status

This is an active migration from Python to Go. The original Python code is preserved in the `legacy/` directory.

**Current Status:** Core functionality is working - file watching, encoding, encryption, and metadata storage are operational. WebRTC transfer and signaling are placeholder implementations.

## Quick Start

### First-Time Setup

1. **Create the config directory and file:**
   ```bash
   # The config directory defaults to:
   # - Linux/macOS: ~/.config/collective
   # - Windows: %APPDATA%\CollectiveFS
   
   mkdir -p ~/.config/collective
   
   # Create config file with the directory path you want to watch
   echo "/path/to/your/watch/directory" > ~/.config/collective/config
   ```

2. **Build the project:**
   ```bash
   # Download dependencies
   go mod download
   
   # Build the encoder binary (used by the main program)
   go build -o lib/encoder ./lib/encoder.go
   
   # Build CollectiveFS
   go build -o cfs ./cmd/cfs
   ```

3. **Run:**
   ```bash
   ./cfs
   ```

## Building

### Full Build Process

```bash
# Download dependencies
go mod download

# Build the encoder binary (used by the main program)
go build -o lib/encoder ./lib/encoder.go

# Build CollectiveFS
go build -o cfs ./cmd/cfs
```

Or build everything in one go:

```bash
go build -o lib/encoder ./lib/encoder.go && go build -o cfs ./cmd/cfs
```

### Encoder Binary Notes

The `lib/encoder.go` file is built as a standalone binary that the main program calls via subprocess. 

**Important:** Since `lib/encoder.go` has `//go:build ignore` tags (excluding it from normal builds), we maintain the dependency by importing reedsolomon in `internal/encoder/dummy_import.go`. This ensures `go mod tidy` doesn't remove the dependency.

## Running

### Basic Usage

```bash
# Run with default config location
./cfs

# Run with verbose logging
./cfs --verbose
# or
./cfs -v

# Override config directory
./cfs --config-dir ~/.collective

# Watch a specific directory (overrides config file)
./cfs --input /path/to/watch

# Show version
./cfs --version

# Show help
./cfs --help
```

### Configuration

CollectiveFS uses a config directory with the following structure:

**Default locations:**
- Linux/macOS: `~/.config/collective/` (uses `os.UserConfigDir()`)
- Windows: `%APPDATA%\CollectiveFS\`
- Fallback: `~/.collective/` (Unix) or current directory

**Required files:**
- `config` - Single line file containing the absolute path to the directory you want to watch
  ```bash
  echo "/home/user/files_to_watch" > ~/.config/collective/config
  ```

**Generated files:**
- `key` - Binary Fernet encryption key (32 bytes, auto-generated on first run)

**Auto-created directories:**
- `.collective/proc/` - Processed/encoded chunks (created in the watched directory)
- `.collective/cache/` - File metadata cache (JSON files keyed by file ID)
- `.collective/public/` - Public files directory

### Example: First Run

```bash
# 1. Create a directory to watch
mkdir -p ~/test_files

# 2. Set up config
mkdir -p ~/.config/collective
echo "$HOME/test_files" > ~/.config/collective/config

# 3. Build (if not already built)
go build -o lib/encoder ./lib/encoder.go && go build -o cfs ./cmd/cfs

# 4. Run
./cfs --verbose

# 5. In another terminal, add a file to test
echo "Hello CollectiveFS" > ~/test_files/test.txt

# The program will:
# - Detect the new file
# - Encode it into chunks (64 data + 32 parity = 96 chunks)
# - Encrypt each chunk
# - Store metadata in ~/test_files/.collective/cache/
```

## Architecture

### Directory Structure

```
cmd/cfs/
  main.go              # Main entry point

internal/
  config/              # Configuration and path management (cross-platform)
  watcher/             # File system watching (fsnotify)
  encoder/             # Reed-Solomon encoding integration
  encryption/          # Fernet encryption
  transfer/            # WebRTC peer-to-peer transfer (placeholder)
  server/              # HTTP server (port 8080)
  storage/             # File metadata and chunk management
  logger/              # Structured logging

lib/
  encoder.go           # Encoder binary source (built separately)
  encoder              # Built encoder binary (gitignored)

legacy/                # Original Python code (archived)
```

### Component Overview

- **File Watching**: Watches configured directory for new files using `fsnotify`
- **Encoding**: Files are encoded into chunks using Reed-Solomon encoding (64 data + 32 parity = 96 chunks)
- **Encryption**: Each chunk is encrypted using Fernet encryption (symmetric)
- **Metadata Storage**: File and chunk metadata stored in `.collective/cache/` as JSON files
- **HTTP Server**: Serves on `localhost:8080` with basic API endpoints
- **Transfer**: WebRTC peer-to-peer transfer (structure exists, signaling needs implementation)

## Dependencies

All dependencies are managed through `go.mod` and fetched from official repositories:

- `github.com/fsnotify/fsnotify` - File system watching
- `github.com/google/uuid` - UUID generation
- `github.com/klauspost/reedsolomon v1.12.5` - Reed-Solomon erasure coding
- `github.com/pion/webrtc/v3` - WebRTC implementation
- `github.com/fernet/fernet-go` - Fernet encryption
- `golang.org/x/crypto` - Cryptographic primitives

**No local source code dependencies** - everything uses official Go modules.

## Simplification & Cleanup

### What Was Simplified

1. **Removed Local Source Code Dependencies**
   - ✅ Removed local `replace` directive for reedsolomon
   - ✅ Now uses `github.com/klauspost/reedsolomon v1.12.5` from Go modules
   - ✅ Removed separate `lib/go.mod` (single `go.mod` at root)
   - ✅ Updated encoder to use current Go standard library (`os.ReadFile`/`os.WriteFile`)

2. **Removed Hardcoded Test/Debug Code**
   - ✅ Removed writing POST data to `test123456.json`
   - ✅ Removed attempting to serve `for_presen.py`
   - ✅ Added proper JSON API responses
   - ✅ Added `/api/status` endpoint for health checks

3. **Improved File Metadata Management**
   - ✅ File metadata persisted to `.collective/cache/` directory
   - ✅ Metadata stored as JSON files keyed by file ID
   - ✅ Functions for storing/loading/listing metadata

### Unused Source Code Directories

The following directories are **NOT used** in the Go build (kept for reference only):

- `reedsolomon/` - Local clone, replaced by Go module
- `cryptography/` - Python-only library, not needed for Go version

These can be safely removed but are kept for reference.

## Current Functionality

### ✅ What's Working

- Cross-platform path handling (Linux, macOS, Windows)
- Configuration management with auto-generated encryption keys
- File system watching (recursive directory watching)
- Encoder execution (calls Go binary via subprocess)
- Encryption/decryption (Fernet compatibility)
- HTTP server with basic API (`/api/status` endpoint)
- Metadata persistence (file and chunk tracking)
- Graceful shutdown (signal handling)

### ⚠️ What Needs Work

- WebRTC signaling server/client (placeholder exists)
- WebRTC peer connections (structure exists, needs implementation)
- File retrieval/decoding workflow (encoder exists, decoder needed)
- Web UI (`index.html` exists but is basic/test code)

## HTTP API

The HTTP server runs on `localhost:8080` by default:

- `GET /` - Serves `index.html` (if present)
- `GET /api/status` - Returns server status, version, and uptime
- `POST /` - Accepts JSON data (placeholder for future API functionality)

Example status check:
```bash
curl http://localhost:8080/api/status
```

## File Processing Workflow

When a new file is detected:

1. **Detection**: File watcher detects new file in watched directory
2. **Encoding**: File is encoded using Reed-Solomon into chunks (64 data + 32 parity)
   - Chunks are written to `.collective/proc/<filename>.d/`
   - Each chunk gets a numeric extension (`.0`, `.1`, `.2`, etc.)
3. **Encryption**: Each chunk is encrypted in-place using Fernet
4. **Metadata**: File metadata is stored in `.collective/cache/<file-id>.json`
   - Contains: file ID, name, folder, chunk count, chunk info (IDs, paths)
5. **Transfer**: Chunks are queued for WebRTC transfer (not yet fully implemented)

## Cross-Platform Compatibility

This implementation is designed to work on:
- Linux
- macOS  
- Windows

All paths are handled in a cross-platform manner using Go's `filepath` package. The config directory automatically adapts to the operating system.

## Troubleshooting

### "Config file not found or unreadable"

Create the config file:
```bash
mkdir -p ~/.config/collective
echo "/path/to/watch" > ~/.config/collective/config
```

### "Encoder failed"

Make sure the encoder binary is built:
```bash
go build -o lib/encoder ./lib/encoder.go
```

### "Failed to create directories"

Check that you have write permissions to the directory specified in your config file.

## Next Steps

1. Implement basic WebRTC signaling (simple HTTP or WebSocket based)
2. Test end-to-end: file → encode → encrypt → store metadata
3. Implement decoder for file reconstruction
4. Add basic file retrieval API
5. Clean up or replace `index.html` with proper UI

## IDEAS

- Ideally there would be some way of enforcing order of the file chunks from the file chunks themselves
- There should be a set of default files to be used for our testing
- Figuring out some way to test the whole workflow locally would also be very beneficial (docker?)
