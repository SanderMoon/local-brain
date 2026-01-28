# Developer Documentation

This directory contains documentation for maintainers and contributors.

## Contents

- **[DISTRIBUTION.md](DISTRIBUTION.md)** - Distribution strategy and methods
- **[RELEASE.md](RELEASE.md)** - Release process and procedures

## User Documentation

For user-facing documentation, see:
- [README.md](../README.md) - Main documentation
- [INSTALL.md](../INSTALL.md) - Installation guide

## Contributing

Contributions are welcome! Please:
1. Read the existing code and documentation
2. Follow Go best practices
3. Add tests for new features
4. Run `make test-all` before submitting
5. Keep commits focused and well-described

## Running Locally

```bash
# Build
make build

# Run tests
make test-all

# Install locally
make dev-install
```

## Release Process

See [RELEASE.md](RELEASE.md) for the complete release process.
