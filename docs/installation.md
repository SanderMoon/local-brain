# Installation Guide

Local Brain can be installed in several ways. Choose the method that works best for your setup.

## Quick Install (Recommended)

### Homebrew (macOS/Linux)

```bash
brew tap SanderMoon/tap
brew install brain
```

Dependencies (fzf and ripgrep) are required and will be suggested automatically:

```bash
brew install fzf ripgrep
```

**Verification:**

```bash
brain --version
brain --help
```

---

## Alternative Installation Methods

### Go Install (All Platforms)

If you have Go 1.21+ installed:

```bash
go install github.com/SanderMoon/local-brain@latest
```

Make sure `$(go env GOPATH)/bin` is in your PATH:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Add this to your shell config (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/SanderMoon/local-brain/releases).

#### Linux

```bash
# AMD64
curl -LO https://github.com/SanderMoon/local-brain/releases/latest/download/brain_Linux_x86_64.tar.gz
tar -xzf brain_Linux_x86_64.tar.gz
sudo mv brain /usr/local/bin/

# ARM64
curl -LO https://github.com/SanderMoon/local-brain/releases/latest/download/brain_Linux_arm64.tar.gz
tar -xzf brain_Linux_arm64.tar.gz
sudo mv brain /usr/local/bin/
```

#### macOS

```bash
# Intel Mac
curl -LO https://github.com/SanderMoon/local-brain/releases/latest/download/brain_Darwin_x86_64.tar.gz
tar -xzf brain_Darwin_x86_64.tar.gz
sudo mv brain /usr/local/bin/

# Apple Silicon
curl -LO https://github.com/SanderMoon/local-brain/releases/latest/download/brain_Darwin_arm64.tar.gz
tar -xzf brain_Darwin_arm64.tar.gz
sudo mv brain /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain
make build
sudo make install
```

For a local installation (installs to `~/.local/bin` instead of `/usr/local/bin`):

```bash
make dev-install
```

Make sure `~/.local/bin` is in your PATH:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

---

## Dependencies

Local Brain requires **fzf** and **ripgrep** for core functionality. Additional tools are optional but recommended.

| Tool | Required | Purpose | Installation |
|------|----------|---------|--------------|
| **fzf** | ✅ Yes | Fuzzy finder for interactive selection | `brew install fzf` |
| **ripgrep** | ✅ Yes | Fast text search | `brew install ripgrep` |
| **bat** | Optional | Syntax-highlighted file preview | `brew install bat` |
| **tmux** | Optional | Dev mode workspace management | `brew install tmux` |
| **jq** | Optional | JSON processing for scripts | `brew install jq` |
| **syncthing** | Optional | Cross-device synchronization | `brew install syncthing` |

### Install All Dependencies

**macOS:**

```bash
brew install fzf ripgrep bat tmux jq syncthing
```

**Ubuntu/Debian:**

```bash
sudo apt install fzf ripgrep bat tmux jq syncthing
```

**Arch Linux:**

```bash
sudo pacman -S fzf ripgrep bat tmux jq syncthing
```

**Fedora/RHEL:**

```bash
sudo dnf install fzf ripgrep bat tmux jq syncthing
```

---

## Shell Integration

### Shell Completion

Local Brain supports shell completion for bash, zsh, and fish.

#### Bash

Add to your `~/.bashrc`:

```bash
source <(brain completion bash)
```

Or install system-wide:

```bash
brain completion bash > /usr/local/etc/bash_completion.d/brain
```

#### Zsh

Add to your `~/.zshrc`:

```bash
source <(brain completion zsh)
```

Or install to zsh completions directory:

```bash
brain completion zsh > "${fpath[1]}/_brain"
```

#### Fish

```bash
brain completion fish > ~/.config/fish/completions/brain.fish
```

### Brain Prompt Helper

Display active brain in your shell prompt (optional).

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Source the prompt helper
source /usr/local/lib/brain/brain-prompt.sh

# Add to your PS1 (bash) or PROMPT (zsh)
# Example for bash:
PS1='$(brain_prompt)[\u@\h \W]\$ '

# Example for zsh:
PROMPT='$(brain_prompt)%n@%h %1~ %# '
```

This displays `[brain: work] ` when a brain is active.

---

## First Run

After installation, initialize your first brain:

```bash
brain init
```

This will:

1. Create the configuration directory at `~/.config/brain`
2. Prompt you for your brain's location (default: `~/brains/default`)
3. Set up the directory structure:
   ```
   ~/brains/
   └── default/
       ├── 00_dump.md      # Inbox for quick captures
       ├── 01_active/      # Active projects
       └── 99_archive/     # Archived projects
   ```
4. Create a symlink at `~/brain` → `~/brains/default`

### Quick Test

Verify everything works:

```bash
brain add "Set up my knowledge management system"
brain dump ls
brain --version
```

---

## Environment Variables

Customize storage locations by setting environment variables in your shell config:

```bash
# Root directory for all brains (default: ~/brains)
export BRAIN_ROOT="$HOME/Dropbox/Brains"

# Location of the active brain symlink (default: ~/brain)
export BRAIN_SYMLINK="$HOME/Desktop/ActiveBrain"

# Config directory (default: ~/.config/brain)
export BRAIN_CONFIG_DIR="$HOME/.config/brain"
```

---

## Updating

### Homebrew

```bash
brew upgrade brain
```

### Go Install

```bash
go install github.com/SanderMoon/local-brain@latest
```

### Manual Update

Download and install the latest release following the [pre-built binaries](#pre-built-binaries) instructions above.

---

## Troubleshooting

### Binary not found after installation

Make sure the installation directory is in your PATH:

- **Go install**: Add `$(go env GOPATH)/bin` to PATH
- **Local install**: Add `~/.local/bin` to PATH
- **System install**: `/usr/local/bin` should already be in PATH

Add to your shell config (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.bashrc  # or source ~/.zshrc
```

### Permission denied

If you get permission errors during system installation:

```bash
sudo make install
```

Or use `make dev-install` for a user-local installation that doesn't require sudo.

### Dependencies not found

Brain requires **fzf** and **ripgrep**. Install them:

```bash
# macOS
brew install fzf ripgrep

# Ubuntu/Debian
sudo apt install fzf ripgrep

# Arch Linux
sudo pacman -S fzf ripgrep
```

### Homebrew tap not found

If `brew install brain` fails, ensure you've tapped the repository first:

```bash
brew tap SanderMoon/tap
brew install brain
```

---

## Uninstallation

### Homebrew

```bash
brew uninstall brain
brew untap SanderMoon/tap
```

### System Installation

```bash
sudo make uninstall
```

Or manually:

```bash
sudo rm /usr/local/bin/brain
sudo rm -rf /usr/local/lib/brain
```

### Local Installation

```bash
rm ~/.local/bin/brain
```

### Remove All Data

To completely remove all brains and configuration:

```bash
rm -rf ~/.config/brain
rm -rf ~/brains
rm ~/brain  # Removes symlink
```

**Warning:** This permanently deletes all your notes and tasks. Back up first if needed.

---

## Next Steps

- Return to the [Quickstart Guide](index.md) for usage examples
- Read the [Command Reference](commands.md) for complete documentation
- Check the [Development Guide](development.md) if you want to contribute

---

## Getting Help

- **Issues**: [https://github.com/SanderMoon/local-brain/issues](https://github.com/SanderMoon/local-brain/issues)
- **Discussions**: [https://github.com/SanderMoon/local-brain/discussions](https://github.com/SanderMoon/local-brain/discussions)
- **Documentation**: [https://sandermoon.github.io/local-brain/](https://sandermoon.github.io/local-brain/)
