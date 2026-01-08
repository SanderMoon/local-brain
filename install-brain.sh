#!/usr/bin/env bash

# Standalone installer for Local Brain CLI
# Can be called from dotfiles or run independently
# Supports: macOS, Ubuntu/Debian, Arch Linux

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "$1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            OS="macos"
            print_info "Detected macOS"
            ;;
        Linux*)
            if [ -f /etc/arch-release ]; then
                OS="arch"
                print_info "Detected Arch Linux"
            elif [ -f /etc/debian_version ]; then
                OS="ubuntu"
                print_info "Detected Debian/Ubuntu"
            elif [ -f /etc/redhat-release ]; then
                OS="fedora"
                print_info "Detected Fedora/RHEL"
            else
                OS="linux"
                print_info "Detected generic Linux"
            fi
            ;;
        *)
            print_error "Unsupported OS: $(uname -s)"
            exit 1
            ;;
    esac
}

# Detect package manager
detect_package_manager() {
    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            PKG_MGR="brew"
            print_success "Found Homebrew"
        else
            print_error "Homebrew not found. Please install from https://brew.sh"
            exit 1
        fi
    elif [[ "$OS" == "ubuntu" ]]; then
        PKG_MGR="apt"
        print_success "Using apt"
    elif [[ "$OS" == "arch" ]]; then
        PKG_MGR="pacman"
        print_success "Using pacman"
    elif [[ "$OS" == "fedora" ]]; then
        PKG_MGR="dnf"
        print_success "Using dnf"
    else
        print_error "No supported package manager found"
        exit 1
    fi
}

# Check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Install a package if not already installed
install_package() {
    local package=$1
    local cmd=${2:-$package}

    if command_exists "$cmd"; then
        print_success "$package already installed"
        return 0
    fi

    print_info "Installing $package..."

    case "$PKG_MGR" in
        brew)
            brew install "$package" 2>&1 | grep -v "HOMEBREW_NO" || true
            if command_exists "$cmd"; then
                print_success "$package installed"
            else
                print_error "Failed to install $package"
                return 1
            fi
            ;;
        apt)
            sudo apt-get update -qq
            sudo apt-get install -y "$package" || {
                print_error "Failed to install $package"
                return 1
            }
            print_success "$package installed"
            ;;
        dnf)
            sudo dnf install -y "$package" || {
                print_error "Failed to install $package"
                return 1
            }
            print_success "$package installed"
            ;;
        pacman)
            sudo pacman -S --noconfirm "$package" || {
                print_error "Failed to install $package"
                return 1
            }
            print_success "$package installed"
            ;;
    esac
}

# Install dependencies
install_dependencies() {
    print_header "Installing Dependencies"

    case "$PKG_MGR" in
        brew)
            install_package "ripgrep" "rg"
            install_package "fzf" "fzf"
            install_package "bat" "bat"
            install_package "syncthing" "syncthing"
            install_package "jq" "jq"
            # make is usually available on macOS, but we ensure it
            if ! command_exists make; then
                print_info "Installing make (via Xcode Command Line Tools)..."
                xcode-select --install || true
            fi
            ;;
        apt)
            install_package "ripgrep" "rg"
            install_package "fzf" "fzf"
            install_package "bat" "bat"
            install_package "syncthing" "syncthing"
            install_package "jq" "jq"
            install_package "make" "make"
            ;;
        dnf)
            install_package "ripgrep" "rg"
            install_package "fzf" "fzf"
            install_package "bat" "bat"
            install_package "syncthing" "syncthing"
            install_package "jq" "jq"
            install_package "make" "make"
            ;;
        pacman)
            install_package "ripgrep" "rg"
            install_package "fzf" "fzf"
            install_package "bat" "bat"
            install_package "syncthing" "syncthing"
            install_package "jq" "jq"
            install_package "make" "make"
            ;;
    esac
}

# Initialize first brain interactively
init_first_brain() {
    print_header "Initialize Your First Brain"

    local config_dir="$HOME/.config/brain"

    # Create config directories
    mkdir -p "$config_dir"/{bin,lib,templates}
    print_success "Created config directories"

    # Check if any brains already configured
    local config_file="$config_dir/config.json"
    if [[ -f "$config_file" ]]; then
        local has_brains=$(grep -c '"brains"' "$config_file" 2>/dev/null || echo "0")
        if [[ "$has_brains" -gt 0 ]]; then
            print_success "Brain already configured"
            return 0
        fi
    fi

    # Run brain init interactively
    print_info "Let's set up your first brain..."
    echo ""

    # Check if brain command is available
    if command -v brain &> /dev/null; then
        brain init
    else
        # Fallback: brain command not in PATH yet
        "$config_dir/bin/brain" init
    fi
}

# Install brain scripts
install_scripts() {
    print_header "Installing Brain Scripts"

    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Ensure we are in the project root where Makefile exists
    if [[ ! -f "$script_dir/Makefile" ]]; then
        print_error "Makefile not found in $script_dir"
        exit 1
    fi

    print_info "Running make install..."
    
    # Run make install pointing to user's config directory
    # PREFIX is set to ~/.config/brain, so files go to:
    #   bin -> ~/.config/brain/bin
    #   lib -> ~/.config/brain/lib/brain
    if make install PREFIX="$HOME/.config/brain" -C "$script_dir"; then
        print_success "Scripts installed using Makefile"
    else
        print_error "Make install failed"
        exit 1
    fi
}

# Update shell RC file
update_shell_rc() {
    print_header "Updating Shell Configuration"

    local path_export='export PATH="$HOME/.config/brain/bin:$PATH"'
    local rc_file=""

    # Detect shell and RC file
    case "$SHELL" in
        */zsh)
            rc_file="$HOME/.zshrc"
            ;;
        */bash)
            if [[ "$OS" == "macos" ]]; then
                rc_file="$HOME/.bash_profile"
            else
                rc_file="$HOME/.bashrc"
            fi
            ;;
        *)
            print_error "Unsupported shell: $SHELL"
            print_info "Please manually add: $path_export"
            return 1
            ;;
    esac

    # Create RC file if it doesn't exist
    touch "$rc_file"

    # Check if PATH export already exists
    if grep -q 'export PATH="$HOME/.config/brain/bin:$PATH"' "$rc_file" 2>/dev/null || \
       grep -q "export PATH=\"\$HOME/.config/brain/bin:\$PATH\"" "$rc_file" 2>/dev/null; then
        print_success "PATH already configured in $rc_file"
    else
        echo "" >> "$rc_file"
        echo "# Local Brain CLI" >> "$rc_file"
        echo "$path_export" >> "$rc_file"
        print_success "Added PATH to $rc_file"
    fi

    # Add 'pg' function alias for project navigation
    if grep -q "function pg()" "$rc_file" 2>/dev/null || grep -q "pg()" "$rc_file" 2>/dev/null; then
        print_success "'pg' function already configured"
    else
        cat >> "$rc_file" << 'EOF'

# Local Brain: Project Go
# Allows jumping to project directories in current shell
pg() {
    local active_dir
    active_dir="$(brain path)/01_active"
    if [ ! -d "$active_dir" ]; then
        echo "Error: Active directory not found"
        return 1
    fi
    local project
    project=$(find "$active_dir" -mindepth 1 -maxdepth 1 -type d | fzf --height 40% --layout=reverse --border --prompt="Go to project> ")
    if [ -n "$project" ]; then
        cd "$project" || return
    fi
}
EOF
        print_success "Added 'pg' function to $rc_file"
        print_info "Run: source $rc_file (or restart your terminal)"
    fi
}

# Configure Syncthing auto-start (macOS only)
configure_syncthing_autostart() {
    if [[ "$OS" == "macos" ]] && command_exists syncthing; then
        print_header "Configuring Syncthing Auto-start"

        if brew services list | grep -q "syncthing.*started"; then
            print_success "Syncthing already configured to auto-start"
        else
            print_info "Setting up Syncthing to start on login..."
            brew services start syncthing 2>&1 | grep -v "error" || true
            print_success "Syncthing configured (will start on next login)"
        fi
    fi
}

# Main installation
main() {
    echo ""
    echo "╔══════════════════════════════════════════╗"
    echo "║   Local Brain CLI Installation           ║"
    echo "╚══════════════════════════════════════════╝"

    detect_os
    detect_package_manager
    install_dependencies
    install_scripts
    update_shell_rc
    configure_syncthing_autostart
    init_first_brain

    print_header "Installation Complete!"
    echo ""
    print_success "Local Brain has been installed successfully"
    echo ""
    print_info "Next steps:"
    echo "  1. Restart your terminal (or run: source ~/.zshrc or source ~/.bashrc)"
    echo "  2. Try: brain add \"My first task\""
    echo "  3. View tasks: brain todo"
    echo ""
    print_info "Directory structure created at: ~/brain"
    print_info "Scripts installed at: ~/.config/brain/bin"
    echo ""
}

# Run main installation
main
