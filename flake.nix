{
  description = "CollectiveFS - A distributed file system built on WebRTC, Reed-Solomon encoding, and encryption";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        # Python (packages will be installed from requirements.txt)
        python = pkgs.python3.withPackages (ps: with ps; [
          pip       # For installing from requirements.txt
          setuptools
          wheel
        ]);

        # Go toolchain
        go = pkgs.go;
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            # Python (packages installed from requirements.txt)
            python
            
            # Go toolchain
            go
            
            # Build tools
            pkgs.gnumake      # In case there are Makefiles
            
            # Utilities
            pkgs.just         # Task runner (optional but useful)
            pkgs.direnv       # Auto-activate env
            
            # Development tools
            pkgs.go-tools     # Go development tools
            pkgs.gopls        # Go language server
          ];

          shellHook = ''
            echo "CollectiveFS Development Environment"
            echo "===================================="
            echo "Python: $(python --version)"
            echo "Go: $(go version)"
            echo ""
            
            # Ensure user site-packages are in Python path
            USER_SITE=$(python -c "import site; print(site.getusersitepackages())" 2>/dev/null)
            if [ -n "$USER_SITE" ] && [ -d "$USER_SITE" ]; then
              export PYTHONPATH="$USER_SITE:$PYTHONPATH"
            fi
            
            # Install Python dependencies from requirements.txt if it exists
            if [ -f requirements.txt ]; then
              if ! python -c "import watchdog, cryptography, simplejson" 2>/dev/null; then
                echo "Installing Python dependencies from requirements.txt..."
                pip install --user --break-system-packages -r requirements.txt 2>/dev/null || \
                pip install --user -r requirements.txt 2>/dev/null || {
                  echo "⚠ Warning: Failed to install dependencies from requirements.txt."
                  echo "   You may need to install manually: pip install -r requirements.txt"
                }
              else
                echo "✓ Python dependencies appear to be installed"
              fi
            else
              echo "⚠ Warning: requirements.txt not found. Python packages should be installed manually."
            fi
            
            # Build encoder if it doesn't exist or is outdated
            ENCODER_PATH="./lib/encoder"
            ENCODER_SRC="./lib/encoder.go"
            
            if [ ! -f "$ENCODER_PATH" ] || [ "$ENCODER_SRC" -nt "$ENCODER_PATH" ] 2>/dev/null; then
              echo "Building encoder executable..."
              cd lib
              
              # Initialize go.mod if it doesn't exist
              if [ ! -f go.mod ]; then
                ${go}/bin/go mod init encoder 2>/dev/null || true
              fi
              
              # Replace reedsolomon with local path (relative to lib directory)
              ${go}/bin/go mod edit -replace github.com/klauspost/reedsolomon=../reedsolomon 2>/dev/null || true
              
              # Get dependencies and tidy
              ${go}/bin/go mod tidy 2>/dev/null || true
              
              # Build the encoder
              echo "Compiling encoder..."
              if ${go}/bin/go build -o encoder encoder.go 2>&1; then
                echo "✓ Encoder built successfully!"
              else
                echo "⚠ Warning: Encoder build had issues. Check above for errors."
              fi
              cd ..
            else
              echo "Encoder is up to date."
            fi
            
            echo ""
            echo "Available commands:"
            echo "  python cfs.py --help                    # Run CollectiveFS"
            echo "  pyinstaller cfs.spec                   # Build executable"
            echo "  cd lib && go build -o encoder encoder.go # Rebuild encoder"
            echo ""
            echo "Environment ready!"
          '';

          # Environment variables for Go
          # Use the project directory structure
          GO111MODULE = "on";
          # Allow Go to use the local reedsolomon module
          GOWORK = "off";
        };

        # Build package for the encoder executable
        packages.encoder = pkgs.stdenv.mkDerivation {
          name = "collectivefs-encoder";
          src = ./.;
          
          nativeBuildInputs = [ go ];
          
          buildPhase = ''
            cd lib
            
            # Set up Go module
            ${go}/bin/go mod init encoder || true
            ${go}/bin/go mod edit -replace github.com/klauspost/reedsolomon=../reedsolomon
            ${go}/bin/go mod tidy
            
            # Build the encoder
            ${go}/bin/go build -o encoder encoder.go
          '';
          
          installPhase = ''
            mkdir -p $out/bin
            cp lib/encoder $out/bin/collectivefs-encoder
          '';
        };
      }
    );
}
