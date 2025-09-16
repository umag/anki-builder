{
  description = "Nederlands Anki Card Builder - Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Go 1.24 - using the latest available Go version
        go = pkgs.go_1_24; # Note: Go 1.24 may not be available yet, using 1.23
        
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            git
            curl
            jq
            golangci-lint
            gopls
            gotools
            go-tools
          ];

          shellHook = ''
            echo "🚀 Nederlands Anki Card Builder Development Environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go run .           - Run the application"
            echo "  go build           - Build the application"
            echo "  go mod tidy        - Clean up dependencies"
            echo "  golangci-lint run  - Run linter"
            echo ""
            echo "Make sure to:"
            echo "  1. Copy config.json.example to config.json"
            echo "  2. Fill in your Gemini API key"
            echo "  3. Ensure Anki is running with AnkiConnect"
            echo ""
            
            # Set up Go environment
            export GOROOT="${go}/share/go"
            export PATH="$GOROOT/bin:$PATH"
            
            # Create config.json if it doesn't exist
            if [ ! -f config.json ]; then
              echo "Creating config.json from example..."
              cp config.json.example config.json
              echo "⚠️  Please edit config.json with your API keys and settings"
            fi
          '';
        };

        # Optional: Add packages that can be built
        packages.default = pkgs.buildGoModule {
          pname = "anki-builder";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # Will need to be updated when you have dependencies
        };
      });
}
