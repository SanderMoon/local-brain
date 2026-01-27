#!/usr/bin/env bash

# Modern installer for Local Brain CLI (Go version)
# This script provides multiple installation methods following Go best practices

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="sandermoonemans/local-brain"
BINARY_NAME="brain"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.config/brain"

# Print functions
print_success() {
  echo -e "${GREEN}✓${NC} $1"
}

print_info() {
  echo -e "${BLUE}→${NC} $1"
}

print_error() {
  echo -e "${RED}✗${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}⚠${NC} $1"
}

print_header() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "$1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Darwin*)
      OS="darwin"
      print_info "Detected macOS"
      ;;
    Linux*)
      OS="linux"
      print_info "Detected Linux"
      ;;
    *)
      print_error "Unsupported OS: $OS"
      exit 1
      ;;
  esac

  case "$ARCH" in
    x86_64)
      ARCH="amd64"
      ;;
    arm64|aarch64)
      ARCH="arm64"
      ;;
    *)
      print_error "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac

  print_success "Platform: $OS/$ARCH"
}

# Check if a command exists
command_exists() {
  command -v "$1" &> /dev/null
}

# Check for Go installation
check_go() {
  if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION installed"
    return 0
  else
    print_warning "Go not installed"
    return 1
  fi
}

# Check for Homebrew (macOS only)
check_brew() {
  if [[ "$OS" == "darwin" ]] && command_exists brew; then
    print_success "Homebrew installed"
    return 0
  else
    return 1
  fi
}

# Installation method 1: Homebrew (macOS, recommended)
install_via_homebrew() {
  print_header "Installing via Homebrew"

  if ! check_brew; then
    print_error "Homebrew not found"
    return 1
  fi

  print_info "Adding tap sandermoonemans/tap..."
  brew tap sandermoonemans/tap || {
    print_warning "Tap not available yet - falling back to other methods"
    return 1
  }

  print_info "Installing brain..."
  brew install brain

  print_success "Installed via Homebrew"
  return 0
}

# Installation method 2: go install (requires Go)
install_via_go_install() {
  print_header "Installing via 'go install'"

  if ! check_go; then
    print_error "Go is required for this installation method"
    return 1
  fi

  print_info "Installing from source..."
  go install github.com/$REPO@latest || {
    print_error "Failed to install via go install"
    return 1
  }

  # Check where it was installed
  GOBIN="${GOBIN:-$(go env GOPATH)/bin}"

  if [[ -f "$GOBIN/$BINARY_NAME" ]]; then
    print_success "Installed to $GOBIN/$BINARY_NAME"

    # Check if GOBIN is in PATH
    if ! echo "$PATH" | grep -q "$GOBIN"; then
      print_warning "Add $GOBIN to your PATH:"
      echo "  export PATH=\"$GOBIN:\$PATH\""
    fi

    return 0
  else
    print_error "Installation failed"
    return 1
  fi
}

# Installation method 3: Download pre-built binary
install_via_binary_download() {
  print_header "Downloading Pre-built Binary"

  print_info "Fetching latest release info..."

  # Get latest release info from GitHub API
  if command_exists curl; then
    RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
  elif command_exists wget; then
    RELEASE_INFO=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest")
  else
    print_error "Neither curl nor wget found"
    return 1
  fi

  # Extract download URL for our platform
  DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*${BINARY_NAME}_.*${OS}_${ARCH}" | cut -d '"' -f 4 | head -n 1)

  if [[ -z "$DOWNLOAD_URL" ]]; then
    print_error "No binary found for $OS/$ARCH"
    return 1
  fi

  VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name":' | cut -d '"' -f 4)
  print_info "Downloading version $VERSION..."

  # Create temporary directory
  TMP_DIR=$(mktemp -d)
  trap "rm -rf $TMP_DIR" EXIT

  # Download archive
  if command_exists curl; then
    curl -L "$DOWNLOAD_URL" -o "$TMP_DIR/brain.tar.gz"
  else
    wget -O "$TMP_DIR/brain.tar.gz" "$DOWNLOAD_URL"
  fi

  # Extract binary
  print_info "Extracting..."
  tar -xzf "$TMP_DIR/brain.tar.gz" -C "$TMP_DIR"

  # Install binary
  mkdir -p "$INSTALL_DIR"
  cp "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"

  print_success "Installed to $INSTALL_DIR/$BINARY_NAME"

  # Check if install dir is in PATH
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    print_warning "Add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  fi

  return 0
}

# Install dependencies
install_dependencies() {
  print_header "Checking Dependencies"

  local deps=("ripgrep:rg" "fzf:fzf" "bat:bat" "syncthing:syncthing" "jq:jq")
  local missing=()

  for dep in "${deps[@]}"; do
    IFS=':' read -r name cmd <<< "$dep"
    if command_exists "$cmd"; then
      print_success "$name installed"
    else
      print_warning "$name not installed (optional but recommended)"
      missing+=("$name")
    fi
  done

  if [[ ${#missing[@]} -gt 0 ]]; then
    echo ""
    print_info "To install missing dependencies:"

    if check_brew; then
      echo "  brew install ${missing[*]}"
    elif [[ "$OS" == "linux" ]]; then
      if command_exists apt; then
        echo "  sudo apt install ${missing[*]}"
      elif command_exists dnf; then
        echo "  sudo dnf install ${missing[*]}"
      elif command_exists pacman; then
        echo "  sudo pacman -S ${missing[*]}"
      fi
    fi
  fi
}

# Initialize first brain
init_brain() {
  print_header "Initialize Your First Brain"

  if [[ ! -f "$CONFIG_DIR/config.json" ]]; then
    print_info "Let's set up your first brain..."
    echo ""

    if command_exists brain; then
      brain init
    elif [[ -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
      "$INSTALL_DIR/$BINARY_NAME" init
    elif [[ -f "$(go env GOPATH)/bin/$BINARY_NAME" ]]; then
      "$(go env GOPATH)/bin/$BINARY_NAME" init
    else
      print_warning "Brain binary not found in PATH"
      print_info "Please run 'brain init' after adding it to your PATH"
    fi
  else
    print_success "Brain already configured"
  fi
}

# Setup shell integration
setup_shell_integration() {
  print_header "Shell Integration"

  local rc_file=""

  case "$SHELL" in
    */zsh)
      rc_file="$HOME/.zshrc"
      ;;
    */bash)
      if [[ "$OS" == "darwin" ]]; then
        rc_file="$HOME/.bash_profile"
      else
        rc_file="$HOME/.bashrc"
      fi
      ;;
    *)
      print_warning "Unsupported shell: $SHELL"
      return
      ;;
  esac

  touch "$rc_file"

  # Add PATH if needed
  local path_export="export PATH=\"$INSTALL_DIR:\$PATH\""
  if ! grep -q "$INSTALL_DIR" "$rc_file" 2>/dev/null; then
    echo "" >> "$rc_file"
    echo "# Local Brain CLI" >> "$rc_file"
    echo "$path_export" >> "$rc_file"
    print_success "Added $INSTALL_DIR to PATH in $rc_file"
  else
    print_success "PATH already configured"
  fi

  print_info "Restart your terminal or run: source $rc_file"
}

# Show installation menu
show_menu() {
  print_header "Local Brain CLI Installer"

  echo "Choose installation method:"
  echo ""

  local methods=()
  local i=1

  if check_brew; then
    echo "  $i) Homebrew (recommended for macOS)"
    methods+=("homebrew")
    ((i++))
  fi

  if check_go; then
    echo "  $i) go install (build from source)"
    methods+=("go_install")
    ((i++))
  fi

  echo "  $i) Download pre-built binary"
  methods+=("binary")
  ((i++))

  echo "  0) Exit"
  echo ""

  read -rp "Select option [0-$((i-1))]: " choice

  if [[ "$choice" == "0" ]]; then
    print_info "Installation cancelled"
    exit 0
  fi

  if [[ "$choice" -ge 1 ]] && [[ "$choice" -lt "$i" ]]; then
    selected_method="${methods[$((choice-1))]}"
    echo ""

    case "$selected_method" in
      homebrew)
        install_via_homebrew || {
          print_warning "Homebrew installation failed, trying alternative..."
          install_via_binary_download
        }
        ;;
      go_install)
        install_via_go_install || {
          print_warning "go install failed, trying alternative..."
          install_via_binary_download
        }
        ;;
      binary)
        install_via_binary_download
        ;;
    esac
  else
    print_error "Invalid option"
    exit 1
  fi
}

# Main installation flow
main() {
  detect_platform

  # If --auto flag is provided, try methods in order
  if [[ "${1:-}" == "--auto" ]]; then
    print_header "Automatic Installation"

    if check_brew && install_via_homebrew; then
      :
    elif check_go && install_via_go_install; then
      :
    elif install_via_binary_download; then
      :
    else
      print_error "All installation methods failed"
      exit 1
    fi
  else
    show_menu
  fi

  echo ""
  install_dependencies
  setup_shell_integration
  init_brain

  print_header "Installation Complete!"
  echo ""
  print_success "Local Brain has been installed successfully"
  echo ""
  print_info "Quick start:"
  echo "  1. Restart your terminal (or source your shell config)"
  echo "  2. Run: brain --help"
  echo "  3. Try: brain add \"My first task\""
  echo ""
  print_info "For more information:"
  echo "  https://github.com/$REPO"
  echo ""
}

# Run main installation
main "$@"
