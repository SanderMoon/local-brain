# Release Process

This document describes how to create a new release of Local Brain.

## Prerequisites

1. **GoReleaser installed:**
   ```bash
   make install-goreleaser
   ```

2. **GitHub token** (for automated releases):
   ```bash
   export GITHUB_TOKEN="your_github_token"
   ```
   Get a token at: https://github.com/settings/tokens

3. **Clean working directory:**
   ```bash
   git status  # Should be clean
   ```

4. **All tests passing:**
   ```bash
   make test-all
   make test-race
   ```

## Release Steps

### 1. Update Version Information

Ensure all code is ready for release:

```bash
# Run full test suite
make check

# Verify build
make build
./brain --version
```

### 2. Create a Git Tag

Choose your version number following [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality
- **PATCH**: Backwards-compatible bug fixes

```bash
# Example: releasing version 2.0.0
VERSION="v2.0.0"

# Create and push the tag
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
```

### 3. GitHub Actions (Automated)

Once you push the tag, GitHub Actions will automatically:
1. Run all tests
2. Build binaries for all platforms
3. Create a GitHub release
4. Upload artifacts
5. Update the Homebrew tap (if configured)

Monitor the workflow at: https://github.com/sandermoonemans/local-brain/actions

### 4. Manual Release (Alternative)

If you prefer to release manually or test locally:

```bash
# Create a snapshot (test without pushing)
make snapshot

# Check the artifacts
ls -la dist/

# Create a real release (requires git tag)
make release
```

## Testing a Release Locally

Before creating an official release, test with a snapshot:

```bash
# Create snapshot release
make snapshot

# Artifacts will be in dist/
ls dist/

# Test the binary for your platform
dist/brain_darwin_arm64/brain --version
```

## Post-Release Steps

### 1. Verify the Release

Check that the release is available:
- GitHub: https://github.com/sandermoonemans/local-brain/releases
- Homebrew (if configured): `brew install sandermoonemans/tap/brain`

### 2. Test Installation Methods

Test each installation method:

```bash
# Go install
go install github.com/sandermoonemans/local-brain@latest

# Homebrew
brew install sandermoonemans/tap/brain

# Direct download
curl -LO https://github.com/sandermoonemans/local-brain/releases/latest/download/brain_Darwin_arm64.tar.gz
```

### 3. Update Documentation

Ensure documentation references the new version:
- README.md
- INSTALL.md
- Any version-specific docs

### 4. Announce the Release

Consider announcing on:
- GitHub Discussions
- Social media
- Relevant communities

## Release Checklist

- [ ] All tests passing (`make test-all`)
- [ ] Race detector clean (`make test-race`)
- [ ] Code formatted (`make fmt`)
- [ ] Linter passing (`make lint`)
- [ ] Version number decided (semantic versioning)
- [ ] CHANGELOG updated (if applicable)
- [ ] Documentation updated
- [ ] Git tag created and pushed
- [ ] GitHub Actions workflow completed successfully
- [ ] Release artifacts available on GitHub
- [ ] Installation methods tested
- [ ] Homebrew formula updated (if applicable)
- [ ] Release announced

## Versioning Strategy

### Major Version (X.0.0)
- Breaking changes to CLI interface
- Incompatible configuration format changes
- Major feature removals

### Minor Version (x.Y.0)
- New commands or features
- Backwards-compatible enhancements
- New configuration options

### Patch Version (x.y.Z)
- Bug fixes
- Performance improvements
- Documentation updates
- Dependency updates

## Rollback Procedure

If a release has critical issues:

### 1. Delete the Release

```bash
# Delete the tag locally
git tag -d $VERSION

# Delete the tag remotely
git push origin :refs/tags/$VERSION
```

### 2. Delete GitHub Release

Go to the releases page and delete the release manually:
https://github.com/sandermoonemans/local-brain/releases

### 3. Fix and Re-release

```bash
# Fix the issue
git commit -m "fix: critical issue"

# Create a new patch version
NEW_VERSION="v2.0.1"
git tag -a $NEW_VERSION -m "Release $NEW_VERSION"
git push origin $NEW_VERSION
```

## Troubleshooting

### GoReleaser fails

Check common issues:
- Working directory is dirty (uncommitted changes)
- No git tag exists
- GitHub token not set or invalid
- Tests failing

### Homebrew formula not updated

Check:
- TAP_GITHUB_TOKEN is set
- Homebrew tap repository exists
- Repository permissions are correct

### Missing artifacts

Ensure `.goreleaser.yml` includes all platforms:
- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm64
- windows/amd64

## Release Automation

The project uses GitHub Actions for automated releases. The workflow:

1. Triggered on tag push (`v*`)
2. Runs full test suite
3. Builds with GoReleaser
4. Creates GitHub release
5. Uploads binaries
6. Updates Homebrew tap

See `.github/workflows/release.yml` for details.

## Beta Releases

For testing features before official release:

```bash
# Tag as pre-release
git tag -a v2.1.0-beta.1 -m "Beta release v2.1.0-beta.1"
git push origin v2.1.0-beta.1
```

GoReleaser will mark it as a pre-release on GitHub.

## Support

For release process questions:
- Open an issue: https://github.com/sandermoonemans/local-brain/issues
- Check GoReleaser docs: https://goreleaser.com
