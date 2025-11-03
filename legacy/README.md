# Legacy Python Code

This directory contains the original Python implementation of CollectiveFS, preserved for reference during the Go migration.

## Files

- `cfs.py` - Main Python orchestrator (original entry point)
- `filexfer/filexfer.py` - WebRTC file transfer implementation

## Original Behavior

The Python code was designed to:
1. Watch a directory for new files
2. Encode files using Reed-Solomon encoding (via Go encoder binary)
3. Encrypt chunks using Fernet encryption
4. Transfer chunks via WebRTC peer-to-peer connections
5. Serve a basic HTTP UI on localhost:8080

## Migration Notes

This code is being translated to Go for better performance, reliability, and cross-platform compatibility. The Go implementation maintains functional parity with this Python version.

