#!/bin/bash

# QueryBuilder CLI Installation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="querybuilder"
INSTALL_DIR="${GOPATH:-$HOME/go}/bin"
REPO_URL="https://github.com/dchlong/querybuilder"

# Functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

check_go() {
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go first: https://golang.org/doc/install"
    fi
    
    local go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    local major=$(echo $go_version | cut -d. -f1)
    local minor=$(echo $go_version | cut -d. -f2)
    
    if [ "$major" -lt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -lt 21 ]); then
        warn "Go version $go_version detected. Recommended: Go 1.21 or higher"
    else
        log "Go version $go_version detected ✓"
    fi
}

install_from_source() {
    log "Installing $BINARY_NAME from source..."
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Clone repository
    log "Cloning repository..."
    git clone "$REPO_URL" . || error "Failed to clone repository"
    
    # Build and install
    log "Building and installing..."
    go install ./cmd/querybuilder || error "Failed to install $BINARY_NAME"
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$temp_dir"
    
    success "$BINARY_NAME installed successfully!"
}

install_direct() {
    log "Installing $BINARY_NAME directly from Go modules..."
    go install github.com/dchlong/querybuilder/cmd/querybuilder@latest || error "Failed to install $BINARY_NAME"
    success "$BINARY_NAME installed successfully!"
}

verify_installation() {
    log "Verifying installation..."
    
    if ! command -v $BINARY_NAME &> /dev/null; then
        warn "$BINARY_NAME not found in PATH"
        warn "Make sure $INSTALL_DIR is in your PATH"
        warn "Add this to your shell profile: export PATH=\"$INSTALL_DIR:\$PATH\""
        return 1
    fi
    
    local version=$($BINARY_NAME -version 2>/dev/null || echo "unknown")
    success "$BINARY_NAME is installed and working!"
    log "Version: $version"
    log "Location: $(which $BINARY_NAME)"
    
    return 0
}

show_usage() {
    cat << EOF
QueryBuilder CLI Installation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -s, --source     Install from source (clone and build)
    -d, --direct     Install directly via go install (default)
    -h, --help       Show this help message

EXAMPLES:
    $0               # Install via go install
    $0 --source      # Install from source
    $0 --direct      # Install via go install (explicit)

REQUIREMENTS:
    - Go 1.21+ installed
    - Git (for --source option)
    - GOPATH/bin or GOBIN in PATH

POST-INSTALLATION:
    Run '$BINARY_NAME --help' to get started!

EOF
}

main() {
    local install_method="direct"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--source)
                install_method="source"
                shift
                ;;
            -d|--direct)
                install_method="direct"
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    log "QueryBuilder CLI Installation"
    log "Install method: $install_method"
    log "Install directory: $INSTALL_DIR"
    echo
    
    # Check prerequisites
    check_go
    
    if [ "$install_method" = "source" ]; then
        if ! command -v git &> /dev/null; then
            error "Git is required for source installation"
        fi
        install_from_source
    else
        install_direct
    fi
    
    echo
    verify_installation
    
    echo
    log "🎉 Installation complete!"
    log "Try running: $BINARY_NAME --help"
    log "Or: $BINARY_NAME -types"
    echo
    log "Example usage:"
    log "  $BINARY_NAME models.go                 # Generate query builder"
    log "  $BINARY_NAME -dir ./internal/models    # Process directory"
    log "  $BINARY_NAME -verbose models.go        # Verbose output"
}

# Run main function
main "$@"