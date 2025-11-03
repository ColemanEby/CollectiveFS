# Objective
The objective of CollectiveFS is to create a public file system where users can store personal files. I draw from protocols such as BitTorrent and BitCoin to create resilant distributed networks.

# Description
The cloud is a cluster of servers owned by a single entity. Typically, their motive is to collect payments directly or from 3rd parties which creates reliability and security risks. CollectiveFS serves to be a public alternative to privately owned cloud storage. Control is distributed among entities who choose to provide disc space in exchange for having their files on the network.

# Hidden in Plain Sight
File chunks are exchanged with untrusted peers but are encrypted. Symetric keys are used since the encryptor/decryptor are the same entity.

# Similar Projects
IPFS - Aims to replace IP based HTTP websites with content addressed ones hosted by p2p clusters. They introduce the concept of pinning where you can prioritize data. On CollectiveFS, each byte is as valued as any other byte on the network and parity can be configured so users can choose their desired level of fault tolerance against data erasures. IPFS also uses version control to track file history. On CollectiveFS, there is no version control although this can be implemented at the user level.

Hadoop - A distributed file system (HDFS) for big data. Used at companies like Facebook. Hadoop must be configured from the top down by a single entity where CollectiveFS is built from the bottom up by the individual nodes.

Syncthing - Synchronizes files over many nodes using p2p. Only synchronizes between nodes you own therefor is not public.


# Technologies
WebRTC  
Symmetric Encryption (Fernet)  
Encoding (ReedSolomon)  
FUSE  

## Saving a file 
![Alt text](/images/CollectiveFS_save_file.png?raw=true "Saving files")


## Getting a file:
![Alt text](/images/CollectiveFS_get_file.png?raw=true "Saving files")



# Colemans Notes

## Questions
- Is it intentional that a pycache was committed to the repo? 
- What are the goals with using Reed Solomon Encoding over other encoding methods for 


## QuickStart

```
pip install -r requirements.txt
```

2. Build the Go encoder:
```bash
cd lib
go mod init encoder  # if needed
go mod edit -replace github.com/klauspost/reedsolomon=../reedsolomon
go mod tidy
go build -o encoder encoder.go
cd ..
```

3. Set up config:
   - The code hardcodes `/home/andy/.collective/` for config
   - Create a config file at `/home/andy/.collective/config` with the path to watch
   - Or modify line 27-28 in `cfs.py` to your path

### Running the system

```bash
# Basic run (watches directory, serves web UI)
python cfs.py

# With verbose logging
python cfs.py --verbose

# Help
python cfs.py --help
```

**What happens when you run it:**
1. Reads config from `/home/andy/.collective/config` (or hardcoded path)
2. Generates/loads encryption key from `/home/andy/.collective/key`
3. Starts HTTP server on `localhost:8080`
4. Watches the configured directory for new files
5. When a new file appears:
   - Encodes it using `lib/encoder` (64 data + 32 parity = 96 chunks)
   - Encrypts each chunk
   - Attempts to transfer chunks via WebRTC

### Using NixOS (from your `flake.nix`)

If you have Nix installed:
```bash
nix develop
# This will auto-install dependencies and build the encoder
python cfs.py
```

## Issues and problems

1. Hardcoded paths:
   - Line 27: `ConfigDir = "/home/andy/.collective/"` — change this
   - Line 218: `programPath` should handle relative paths better

2. Missing decoder:
   - No decoder implementation to reconstruct files from chunks
   - You'd need to build one using the Reed-Solomon library

3. Incomplete WebRTC signaling:
   - `filexfer.py` uses hardcoded signaling (`127.0.0.1:1234`)
   - No actual signaling server implementation visible

4. Web UI:
   - `index.html` appears to be leftover test code

5. No tests:
   - No test files or test infrastructure visible

## Quick start script (recommendation)

Create a setup script:

```bash
#!/bin/bash
# setup.sh

# Install Python deps
pip install -r requirements.txt

# Build encoder
cd lib
go mod init encoder 2>/dev/null || true
go mod edit -replace github.com/klauspost/reedsolomon=../reedsolomon
go build -o encoder encoder.go
cd ..

# Create config directory (update path!)
mkdir -p ~/.collective
echo "$HOME/test_watch_dir" > ~/.collective/config
mkdir -p "$HOME/test_watch_dir"

echo "Setup complete! Update ~/.collective/config with your watch directory"
echo "Then run: python cfs.py"
```

## Testing workflow

Since there's no formal test suite:

1. Create a test directory:
```bash
mkdir -p ~/test_collective
echo "test file" > ~/test_collective/test.txt
```

2. Update config to point to it

3. Run `cfs.py` and copy a file into the watch directory

4. Check the `.collective/proc/` directory for encoded chunks

## Summary

- Entry point: `python cfs.py`
- Config: `/home/andy/.collective/config` (or modify the hardcoded path)
- Dependencies: Install from `requirements.txt`, build Go encoder
- Workflow: Watches directory → encodes → encrypts → attempts to transfer

The project is functional but incomplete (no decoder, basic signaling, hardcoded paths). If you want help fixing these, say which to tackle first.