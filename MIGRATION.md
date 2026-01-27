# Migration Guide: Bash to Go

This guide helps you migrate from the bash version of Local Brain to the modern Go implementation.

## Why Migrate?

The Go rewrite provides several advantages:

- **Performance**: ~10x faster startup time
- **Single Binary**: No dependency on specific bash versions or shell utilities
- **Cross-platform**: Better Windows support (future)
- **Easier Installation**: Standard Go installation methods
- **Better Testing**: Comprehensive test coverage with race detection
- **Professional Distribution**: GitHub releases, Homebrew, `go install`

## Backward Compatibility

The Go version is **100% backward compatible** with the bash version:

✅ Same configuration format (JSON)
✅ Same directory structure
✅ Same task ID generation (MD5-based)
✅ Same file formats (Markdown)
✅ Same CLI interface
✅ Same behavior

You can switch between bash and Go versions without any data migration.

## Installation

### Uninstall Bash Version (Optional)

If you installed the bash version system-wide:

```bash
# Remove bash scripts
sudo rm -f /usr/local/bin/brain*
sudo rm -rf /usr/local/lib/brain

# Or if using the old install script locations
rm -rf ~/.config/brain/bin
```

**Note**: This won't affect your brain data or configuration!

### Install Go Version

Choose your preferred method:

#### Homebrew (Recommended for macOS)

```bash
brew tap sandermoonemans/tap
brew install brain
```

#### Go Install

```bash
go install github.com/sandermoonemans/local-brain@latest
```

#### Pre-built Binary

See [INSTALL.md](INSTALL.md) for detailed instructions.

## Verification

Verify the Go version is installed:

```bash
brain --version
# Should show: vX.Y.Z (commit: abc1234) (built: 2024-01-01)

# Test basic functionality
brain todo
brain add "Test the new version"
```

## Key Differences

While functionally identical, there are some subtle differences:

### 1. Single Binary vs Multiple Scripts

**Bash version:**
```bash
$ ls /usr/local/bin/brain*
brain-add
brain-todo
brain-project
... (18 separate scripts)
```

**Go version:**
```bash
$ ls /usr/local/bin/brain
brain  # Single binary with subcommands
```

Usage is identical - both support `brain add`, `brain todo`, etc.

### 2. Error Messages

The Go version may have slightly different error formatting, but all errors are handled gracefully.

### 3. Performance

The Go version is significantly faster, especially for operations like:
- `brain todo` - Parsing and filtering tasks
- `brain project list` - Repository scanning
- Shell completion

### 4. JSON API

Both versions support JSON output with `--json`, but the Go version has more consistent formatting.

## Shell Integration

Update your shell configuration to use the new installation:

### If you used the old install script

**Old (~/.zshrc):**
```bash
export PATH="$HOME/.config/brain/bin:$PATH"
source "$HOME/.config/brain/lib/brain-prompt.sh"
```

**New (~/.zshrc):**
```bash
# If installed via Homebrew
source "$(brew --prefix)/share/brain/brain-prompt.sh"

# Or if installed via go install
export PATH="$(go env GOPATH)/bin:$PATH"

# Or if using make dev-install
export PATH="$HOME/.local/bin:$PATH"
```

### Shell Completion

The Go version has built-in completion generation:

```bash
# Bash
brain completion bash > /usr/local/etc/bash_completion.d/brain

# Zsh
brain completion zsh > "${fpath[1]}/_brain"

# Fish
brain completion fish > ~/.config/fish/completions/brain.fish
```

## Troubleshooting

### "brain: command not found"

Ensure the installation directory is in your PATH:

```bash
# Check where brain is installed
which brain

# If using go install, add to PATH
export PATH="$(go env GOPATH)/bin:$PATH"

# If using dev-install, add to PATH
export PATH="$HOME/.local/bin:$PATH"
```

### "config.json not found"

Your existing configuration should work automatically. If not:

```bash
# Check config location
ls ~/.config/brain/config.json

# Re-initialize if needed (won't lose data)
brain init
```

### Different behavior

If you notice any behavioral differences:

1. Check versions:
   ```bash
   brain --version
   ```

2. Verify configuration:
   ```bash
   brain path
   cat ~/.config/brain/config.json
   ```

3. Report issues:
   https://github.com/sandermoonemans/local-brain/issues

### Performance issues

If the Go version seems slower:

1. Check for proper installation (not running bash version):
   ```bash
   file $(which brain)
   # Should show: Mach-O executable (not shell script)
   ```

2. Clear any stale locks:
   ```bash
   rm ~/.config/brain/.lock
   ```

## Reverting to Bash Version

If you need to temporarily revert:

```bash
# The bash version is still in the bin/ directory
cd /path/to/local-brain-repo
bin/brain-add "Task added with bash version"
```

Your data remains compatible with both versions.

## Development

For developers working on the project:

### Building from Source

```bash
git clone https://github.com/sandermoonemans/local-brain.git
cd local-brain
make build
make dev-install
```

### Running Tests

```bash
# Unit tests
make test-unit

# All tests
make test-all

# With coverage
make test-cover
```

### Contributing

The Go version is the actively maintained version. New features and bug fixes should target the Go implementation in `cmd/` and `pkg/`.

The bash version in `bin/` is kept for reference and backward compatibility testing.

## FAQ

### Q: Will the bash version continue to be supported?

A: The bash version will remain in the repository for reference, but active development focuses on the Go version.

### Q: Do I need to migrate my data?

A: No! The data format is identical. Just install the Go version and continue using your existing brain.

### Q: What about the dependencies (ripgrep, fzf, bat)?

A: These are still recommended for the Go version. Install them the same way:

```bash
# macOS
brew install ripgrep fzf bat syncthing jq

# Linux
sudo apt install ripgrep fzf bat syncthing jq
```

### Q: Can I use both versions simultaneously?

A: Yes! They're fully compatible. However, we recommend using only the Go version to avoid confusion.

### Q: What about brain-prompt.sh?

A: The prompt script still exists and works with the Go version. Install it using:

```bash
# Homebrew
source "$(brew --prefix)/share/brain/brain-prompt.sh"

# Manual installation
cp lib/brain-prompt.sh ~/.local/share/brain/
source ~/.local/share/brain/brain-prompt.sh
```

## Getting Help

- Issues: https://github.com/sandermoonemans/local-brain/issues
- Discussions: https://github.com/sandermoonemans/local-brain/discussions
- Comparison Document: [COMPARISON.md](COMPARISON.md)

## Next Steps

1. Install the Go version using your preferred method
2. Verify it works with your existing brain
3. Update your shell configuration
4. Set up shell completion
5. Enjoy the improved performance!

Welcome to the modern Go version of Local Brain! 🎉
