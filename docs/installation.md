# Installation Guide

## Quick Install (Recommended)

### Homebrew (macOS/Linux)

```bash
brew tap SanderMoon/tap
brew install brain fzf ripgrep
brain --version
```

---

## Alternative Installation Methods

### Go Install (All Platforms)

Requires Go 1.25+. Ensure `$(go env GOPATH)/bin` is in your PATH.

```bash
go install github.com/SanderMoon/local-brain@latest
```

### Pre-built Binaries

Download from the [releases page](https://github.com/SanderMoon/local-brain/releases), then:

```bash
# Replace <ARCHIVE> with the file for your platform
# (e.g., brain_Darwin_arm64.tar.gz, brain_Linux_x86_64.tar.gz)
tar -xzf <ARCHIVE>
sudo mv brain /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain
make build
sudo make install       # system-wide, or:
make dev-install        # installs to ~/.local/bin (no sudo)
```

---

## Dependencies

**Required:** fzf (interactive selection), ripgrep (text search).

**Optional:** bat (syntax-highlighted preview), tmux (dev mode), jq (JSON scripting), syncthing (cross-device sync).

```bash
# macOS
brew install fzf ripgrep bat tmux jq syncthing

# Ubuntu/Debian
sudo apt install fzf ripgrep bat tmux jq syncthing

# Arch
sudo pacman -S fzf ripgrep bat tmux jq syncthing
```

---

## Shell Integration

### Shell Completion

```bash
# Bash (add to ~/.bashrc)
source <(brain completion bash)

# Zsh (add to ~/.zshrc)
source <(brain completion zsh)

# Fish
brain completion fish > ~/.config/fish/completions/brain.fish
```

### Prompt Helper (optional)

Display the active brain name in your shell prompt. Add to your `~/.bashrc` or `~/.zshrc`:

```bash
source /usr/local/lib/brain/brain-prompt.sh

# Bash:
PS1='$(brain_prompt)[\u@\h \W]\$ '
# Zsh:
PROMPT='$(brain_prompt)%n@%h %1~ %# '
```

This shows `[brain: work] ` when a brain is active.

---

## First Run

```bash
brain new              # creates ~/brains/default with inbox, project folders, and templates
brain new work         # or create a named brain
```

This sets the new brain as active (symlinked to `~/brain`).

Verify it works:

```bash
brain add "Set up my knowledge management system"
brain dump ls
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

```bash
brew upgrade brain                                    # Homebrew
go install github.com/SanderMoon/local-brain@latest   # Go
```

For pre-built binaries, download and replace following the [instructions above](#pre-built-binaries).

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `brain: command not found` | Ensure the install directory is in your PATH. For `go install`: add `$(go env GOPATH)/bin`. For `make dev-install`: add `~/.local/bin`. Then reload your shell (`source ~/.zshrc`). |
| Permission denied on install | Use `sudo make install`, or `make dev-install` to avoid sudo entirely. |
| Homebrew tap not found | Run `brew tap SanderMoon/tap` before `brew install brain`. |

---

## Uninstallation

```bash
# Homebrew
brew uninstall brain && brew untap SanderMoon/tap

# System install
sudo rm /usr/local/bin/brain /usr/local/lib/brain

# Local install
rm ~/.local/bin/brain
```

To remove all data (**permanently deletes all notes and tasks**):

```bash
rm -rf ~/.config/brain ~/brains ~/brain
```

---

## AI Agent Setup

Local Brain auto-detects installed AI agents (Claude Code, Codex, Gemini CLI, OpenCode).

```bash
brain mcp install          # register the MCP server with all detected agents
brain mcp status           # check registration

brain skill install        # install all bundled workflow skills
brain skill status         # check what's installed
```

To target a specific agent, add `--agent claude` (or `codex`, `gemini`, `opencode`).

After updating Local Brain, run `brain skill upgrade` to push new skill versions to your agents. See the [Skills reference](skills.md) for the full list of bundled skills, customization, and authoring your own.

---

## Next Steps

- [Quickstart Guide](index.md) for usage examples
- [Command Reference](commands.md) for all commands
- [Development Guide](development.md) to contribute
- [Issues](https://github.com/SanderMoon/local-brain/issues) / [Discussions](https://github.com/SanderMoon/local-brain/discussions) for help
