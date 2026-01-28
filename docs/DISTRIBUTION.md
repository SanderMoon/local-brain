# Distribution Strategy

This document outlines how Local Brain is distributed and the rationale behind each method.

## Distribution Methods

Local Brain follows modern Go CLI distribution best practices:

### 1. Go Install (Primary for Go Users)

```bash
go install github.com/sandermoonemans/local-brain@latest
```

**Pros:**
- Standard Go tooling
- Always builds from source
- Works on any platform with Go installed
- No pre-built binaries needed

**Cons:**
- Requires Go installation
- Slower than pre-built binaries
- Doesn't include shell completion files

**Best for:**
- Go developers
- Users who want latest commit
- Platforms without pre-built binaries

### 2. Homebrew (Primary for macOS)

```bash
brew tap sandermoonemans/tap
brew install brain
```

**Pros:**
- Native macOS experience
- Automatic updates via `brew upgrade`
- Handles dependencies
- Includes shell completion
- Trusted by macOS users

**Cons:**
- macOS and Linux only
- Requires maintaining a tap repository
- Release delay (formula updates)

**Best for:**
- macOS users (recommended)
- Linux users who use Homebrew

**Setup Requirements:**
1. Create tap repository: `github.com/sandermoonemans/homebrew-tap`
2. GoReleaser automatically updates formula on release
3. Configure TAP_GITHUB_TOKEN for automated updates

### 3. GitHub Releases (Universal)

Pre-built binaries for all platforms:
- darwin/amd64 (Intel Mac)
- darwin/arm64 (Apple Silicon)
- linux/amd64
- linux/arm64
- windows/amd64 (future)

**Pros:**
- Fast download and installation
- No build tools required
- Checksums for verification
- All platforms supported

**Cons:**
- Manual download and installation
- Manual updates
- Multiple files to choose from

**Best for:**
- CI/CD pipelines
- Users without Go or Homebrew
- Quick one-time installations

### 4. Install Script (Automated)

```bash
curl -fsSL https://raw.githubusercontent.com/sandermoonemans/local-brain/main/install.sh | bash
```

**Pros:**
- One-command installation
- Detects platform automatically
- Installs dependencies
- Sets up shell integration

**Cons:**
- Requires trusting the script
- Less control over installation
- Needs curl/wget

**Best for:**
- New users
- Quick demos
- Automated provisioning

## Build and Release Process

### GoReleaser Configuration

`.goreleaser.yml` handles:
- Multi-platform builds (darwin, linux, windows)
- Archive generation with checksums
- GitHub release creation
- Homebrew tap updates
- Shell completion packaging
- Changelog generation

### GitHub Actions Workflow

`.github/workflows/release.yml`:
- Triggered on git tags (`v*`)
- Runs full test suite
- Builds with GoReleaser
- Publishes to GitHub Releases
- Updates Homebrew tap (if configured)

### Makefile Targets

```bash
make build          # Local development build
make install        # System-wide installation (/usr/local/bin)
make dev-install    # User installation (~/.local/bin)
make snapshot       # Test release without tags
make release        # Production release (requires tag)
```

## Version Management

### Semantic Versioning

- **Major (X.0.0)**: Breaking changes
- **Minor (0.X.0)**: New features, backward compatible
- **Patch (0.0.X)**: Bug fixes, backward compatible

### Version Injection

Version information is injected at build time:

```go
// main.go
var (
    version = "dev"    // Injected: git tag
    commit  = "unknown" // Injected: git commit hash
    date    = "unknown" // Injected: build timestamp
)
```

Build flags (Makefile and GoReleaser):
```bash
-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"
```

### Pre-releases

Beta/RC versions:
```bash
git tag v2.0.0-beta.1
git push origin v2.0.0-beta.1
```

GoReleaser marks these as pre-releases on GitHub.

## Installation Locations

### System-wide Installation

- **Binary**: `/usr/local/bin/brain`
- **Library**: `/usr/local/lib/brain/brain-prompt.sh`
- **Completions**: `/usr/local/etc/bash_completion.d/`

### User-local Installation

- **Binary**: `~/.local/bin/brain` or `$(go env GOPATH)/bin/brain`
- **Config**: `~/.config/brain/`
- **Completions**: User shell completion directories

### Homebrew Installation

- **Binary**: `$(brew --prefix)/bin/brain`
- **Library**: `$(brew --prefix)/share/brain/brain-prompt.sh`
- **Completions**: Managed by Homebrew

## Shell Completion

Cobra provides built-in completion generation:

```bash
brain completion bash > completions/brain.bash
brain completion zsh > completions/_brain
brain completion fish > completions/brain.fish
```

These are:
- Generated during release
- Included in archives
- Automatically installed by Homebrew
- Available via `brain completion` command

## Dependency Management

### Build Dependencies

- **Go 1.21+**: Required for building
- **goreleaser**: Release automation (optional, for maintainers)

### Runtime Dependencies (Optional)

Not bundled, but recommended:
- **ripgrep** (rg): Fast text search
- **fzf**: Fuzzy finder
- **bat**: Syntax highlighting
- **syncthing**: Synchronization
- **jq**: JSON processing

The install script helps users install these.

## Update Strategy

### Homebrew

```bash
brew upgrade brain
```

### Go Install

```bash
go install github.com/sandermoonemans/local-brain@latest
```

### Manual

```bash
# Check current version
brain --version

# Download new version
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Darwin_arm64.tar.gz

# Replace binary
tar -xzf brain_Darwin_arm64.tar.gz
sudo mv brain /usr/local/bin/
```

### In-app Update Check (Future)

Could add version checking:
```bash
brain update-check
# You are running v2.0.0
# Latest version is v2.1.0
# Run: brew upgrade brain
```

## Security

### Checksums

GoReleaser generates `checksums.txt` for all artifacts:
```bash
# Verify download
sha256sum -c checksums.txt
```

### Signing (Future)

Could add GPG signing:
```yaml
# .goreleaser.yml
signs:
  - cmd: gpg
    args:
      - --output
      - $signature
      - --detach-sign
      - $artifact
    signature: ${artifact}.sig
```

### Supply Chain Security

- Dependabot for dependency updates
- GitHub Actions security scanning
- Minimal dependencies (only Cobra)
- Reproducible builds

## Monitoring and Analytics

### Download Metrics

GitHub provides:
- Release download counts
- Asset-specific downloads
- Per-platform statistics

### Homebrew Metrics

Homebrew analytics (opt-in):
- Installation counts
- Platform distribution
- Update frequency

## Future Considerations

### Additional Platforms

- **Windows**: Already supported in GoReleaser config
- **BSD**: Could be added if requested
- **ARM32**: For Raspberry Pi

### Package Managers

Potential additions:
- **apt**: Debian/Ubuntu native packages
- **rpm**: Fedora/RHEL native packages
- **AUR**: Arch User Repository
- **Scoop**: Windows package manager
- **Chocolatey**: Windows package manager

### Container Images

Docker image for:
- Testing
- CI/CD
- Sandboxed usage

```dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY brain /usr/local/bin/
ENTRYPOINT ["brain"]
```

### Homebrew Core

Eventually submit to homebrew-core for:
- Wider distribution
- Official verification
- No tap needed

Requirements:
- Stable release history
- Active maintenance
- Popular enough (30+ forks or 75+ stars)

## Best Practices Followed

✅ Semantic versioning
✅ Automated releases
✅ Multi-platform builds
✅ Checksums for verification
✅ Shell completion support
✅ Standard installation paths
✅ Clear upgrade path
✅ Homebrew tap for macOS
✅ GitHub Releases
✅ `go install` support
✅ Minimal dependencies
✅ Reproducible builds
✅ CI/CD integration
✅ Version information in binary

## Resources

- **GoReleaser**: https://goreleaser.com
- **Homebrew**: https://docs.brew.sh/
- **Go Modules**: https://go.dev/ref/mod
- **Semantic Versioning**: https://semver.org
- **Cobra**: https://github.com/spf13/cobra

## Maintenance

### Release Checklist

See [RELEASE.md](RELEASE.md) for detailed release process.

Quick version:
1. Ensure tests pass
2. Update version number
3. Create git tag
4. Push tag (triggers automated release)
5. Verify artifacts on GitHub
6. Test installation methods

### Deprecation

If deprecating a distribution method:
1. Announce 3 releases in advance
2. Update documentation
3. Provide migration guide
4. Keep method working during deprecation period
5. Remove only after sufficient notice

## Support

- Report issues: https://github.com/sandermoonemans/local-brain/issues
- Installation help: See [INSTALL.md](INSTALL.md)
- Migration guide: See [MIGRATION.md](MIGRATION.md)
