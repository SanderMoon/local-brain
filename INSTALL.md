# Installation Guide

Local Brain can be installed in several ways. Choose the method that works best for your setup.

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/sandermoonemans/local-brain/main/install.sh | bash
```

Or for interactive installation:

```bash
curl -fsSL https://raw.githubusercontent.com/sandermoonemans/local-brain/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

## Installation Methods

### 1. Homebrew (macOS - Recommended)

```bash
brew tap sandermoonemans/tap
brew install brain
```

Dependencies will be automatically suggested. To install them:

```bash
brew install ripgrep fzf bat syncthing jq
```

### 2. Go Install (All Platforms)

If you have Go 1.21+ installed:

```bash
go install github.com/sandermoonemans/local-brain@latest
```

This will install the `brain` binary to `$(go env GOPATH)/bin`. Make sure this directory is in your PATH:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### 3. Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/sandermoonemans/local-brain/releases).

**Linux:**
```bash
# AMD64
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Linux_x86_64.tar.gz
tar -xzf brain_Linux_x86_64.tar.gz
sudo mv brain /usr/local/bin/

# ARM64
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Linux_arm64.tar.gz
tar -xzf brain_Linux_arm64.tar.gz
sudo mv brain /usr/local/bin/
```

**macOS:**
```bash
# AMD64 (Intel)
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Darwin_x86_64.tar.gz
tar -xzf brain_Darwin_x86_64.tar.gz
sudo mv brain /usr/local/bin/

# ARM64 (Apple Silicon)
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Darwin_arm64.tar.gz
tar -xzf brain_Darwin_arm64.tar.gz
sudo mv brain /usr/local/bin/
```

### 4. Build from Source

```bash
git clone https://github.com/sandermoonemans/local-brain.git
cd local-brain
make build
sudo make install
```

Or for a local installation:

```bash
make dev-install
```

This installs to `~/.local/bin` instead of `/usr/local/bin`.

## Dependencies

Local Brain has a few optional but recommended dependencies for full functionality:

| Tool | Purpose | Installation |
|------|---------|--------------|
| **ripgrep** | Fast text search | `brew install ripgrep` or `apt install ripgrep` |
| **fzf** | Fuzzy finder for interactive selection | `brew install fzf` or `apt install fzf` |
| **bat** | Syntax-highlighted file preview | `brew install bat` or `apt install bat` |
| **syncthing** | Cross-device synchronization | `brew install syncthing` or `apt install syncthing` |
| **jq** | JSON processing | `brew install jq` or `apt install jq` |

Install all at once:

**macOS:**
```bash
brew install ripgrep fzf bat syncthing jq
```

**Ubuntu/Debian:**
```bash
sudo apt install ripgrep fzf bat syncthing jq
```

**Arch Linux:**
```bash
sudo pacman -S ripgrep fzf bat syncthing jq
```

**Fedora/RHEL:**
```bash
sudo dnf install ripgrep fzf bat syncthing jq
```

## Shell Completion

Local Brain supports shell completion for bash, zsh, and fish.

### Bash

```bash
brain completion bash > /usr/local/etc/bash_completion.d/brain
```

Or add to your `~/.bashrc`:

```bash
source <(brain completion bash)
```

### Zsh

Add to your `~/.zshrc`:

```bash
source <(brain completion zsh)
```

Or install to zsh completions directory:

```bash
brain completion zsh > "${fpath[1]}/_brain"
```

### Fish

```bash
brain completion fish > ~/.config/fish/completions/brain.fish
```

## First Run

After installation, initialize your first brain:

```bash
brain init
```

This will:
1. Create the configuration directory at `~/.config/brain`
2. Prompt you for your brain's location (default: `~/brain`)
3. Set up the directory structure
4. Configure the initial brain

## Verification

Verify the installation:

```bash
brain --version
brain --help
```

Try creating your first task:

```bash
brain add "Set up my knowledge management system"
brain todo
```

## Updating

### Homebrew

```bash
brew upgrade brain
```

### Go Install

```bash
go install github.com/sandermoonemans/local-brain@latest
```

### Manual Update

Download and install the latest release following the [pre-built binaries](#3-pre-built-binaries) instructions above.

## Troubleshooting

### Binary not found after installation

Make sure the installation directory is in your PATH:

- For `go install`: Add `$(go env GOPATH)/bin` to PATH
- For `make dev-install`: Add `~/.local/bin` to PATH
- For system install: `/usr/local/bin` should already be in PATH

Add to your shell config (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Permission denied

If you get permission errors during system installation:

```bash
sudo make install
```

Or use `make dev-install` for a user-local installation.

### Dependencies not found

The tool will work without optional dependencies but with reduced functionality. Install dependencies as shown in the [Dependencies](#dependencies) section.

## Uninstallation

### Homebrew

```bash
brew uninstall brain
brew untap sandermoonemans/tap
```

### System Installation

```bash
sudo make uninstall
```

### Local Installation

```bash
rm ~/.local/bin/brain
```

### Remove Configuration

To completely remove all data:

```bash
rm -rf ~/.config/brain
rm -rf ~/brain  # Or wherever your brain directory is located
```

## Next Steps

- Read the [README](README.md) for usage examples
- Check the [documentation](docs/) for advanced features
- Join the community discussions

## Support

- Issues: https://github.com/sandermoonemans/local-brain/issues
- Discussions: https://github.com/sandermoonemans/local-brain/discussions
