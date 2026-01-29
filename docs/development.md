# Development Guide

Guide for contributors and developers working on Local Brain.

---

## Quick Start for Contributors

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Make** - Build automation
- **Git** - Version control

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/SanderMoon/local-brain.git
cd local-brain

# Build the binary
make build

# Run tests
make test

# Install to ~/.local/bin for testing
make dev-install
```

### Build Commands

```bash
make build              # Build binary to ./brain
make dev-install        # Install to ~/.local/bin
make install            # Install to /usr/local/bin (system-wide, requires sudo)
```

### Testing

```bash
make test               # Fast unit tests only (default)
make test-all           # All tests including integration
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-cover         # Generate coverage report (coverage.html)
make test-race          # Run with race detection

# Run single test
go test -v ./pkg/api -run TestGenerateItemID
go test -v ./pkg/config -run TestLoadConfig
```

### Code Quality

```bash
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint (if installed)
make check              # Run all checks (fmt, vet, test)
```

### Release Testing

```bash
make snapshot           # Test release build without tags (artifacts in dist/)
make release            # Create production release (requires git tag)
```

---

## Architecture Overview

### Project Structure

```
local-brain/
├── cmd/                    # Cobra CLI commands (19+ commands)
│   ├── root.go            # Base command, version handling
│   ├── add.go             # Quick capture
│   ├── todo.go            # Task management
│   ├── project.go         # Project commands
│   └── ...
├── pkg/                   # Core packages
│   ├── api/               # Business logic (JSON API layer)
│   ├── config/            # Configuration management
│   ├── fileutil/          # File operations & locking
│   ├── external/          # External tool integration
│   ├── markdown/          # Markdown parsing
│   └── testutil/          # Test utilities
├── docs/                  # Documentation
├── lib/                   # Shell integration scripts
├── Makefile               # Build automation
├── .goreleaser.yml        # Release configuration
└── go.mod                 # Go module definition
```

### Package Structure

#### `cmd/` - Cobra CLI Commands

- Each command in separate file (e.g., `add.go`, `project.go`, `todo.go`)
- All register with `rootCmd` in their `init()` functions
- `root.go` defines base command and version handling

**Key Commands:**
- `add.go` - Quick capture to dump
- `refile.go` - Interactive dump processing
- `plan.go` - Batch task planning
- `todo.go` - Task management with filters
- `project.go` - Project lifecycle
- `dump.go` - Dump item listing

#### `pkg/api/` - Core Business Logic

**Purpose:** JSON API layer, file parsing, ID generation

**Key Files:**
- `dump.go` - Parse dump file, generate stable IDs for items
- `todo.go` - Parse todo.md files, extract tasks with metadata
- `note.go` - Parse notes.md files, extract note entries
- `project.go` - List projects, extract repo URLs from `.repos` files
- `id.go` - MD5-based ID generation (**must** match bash version for compatibility)

**Design:**
- Stateless functions for parsing and manipulation
- Returns structured data (TodoItem, NoteEntry, etc.)
- Thread-safe, no shared state

#### `pkg/config/` - Configuration Management

**Key Files:**
- `config.go` - Brain config (JSON), thread-safe with mutex
- `paths.go` - Path resolution with env var overrides
- `symlink.go` - Active brain symlink management

**Thread Safety:**
- Uses `sync.RWMutex` for concurrent access
- Safe for parallel command execution

**Environment Variables:**
- `BRAIN_ROOT` - Root directory for brains (default: `~/brains`)
- `BRAIN_SYMLINK` - Active brain symlink location (default: `~/brain`)
- `BRAIN_CONFIG_DIR`, `BRAIN_CONFIG_PATH` - Config overrides

#### `pkg/fileutil/` - File Operations

**Key Files:**
- `atomic.go` - Atomic file writes with locking
- `lock.go` - Directory-based file locking (prevents race conditions)
- `platform.go` - Cross-platform utilities (symlinks, path expansion)

**Design Patterns:**
- **Atomic Writes:** Use temp file + rename for crash safety
- **File Locking:** Directory-based locks prevent concurrent modification
- **Cross-platform:** Handles symlinks, ~ expansion on all OSes

#### `pkg/external/` - External Tool Integration

**Key Files:**
- `git.go` - Git operations (clone, pull, status)
- `fzf.go` - Fuzzy finder integration for interactive selection
- `tmux.go` - Tmux session management for dev mode
- `editor.go` - Launch user's editor (vim/nvim)

**Design:**
- Check tool availability before use (`IsFZFAvailable()`)
- Graceful degradation (interactive vs non-interactive fallback)

#### `pkg/markdown/` - Markdown Parsing

**Key Files:**
- `parser.go` - Extract tasks/notes from markdown with metadata

**Supported Metadata:**
- `#captured:YYYY-MM-DD` - Capture timestamp
- `#done:YYYY-MM-DD` - Completion timestamp
- `#p:N` - Priority (1-3)
- `#due:YYYY-MM-DD` - Due date
- `#tagname` - Free-form tags

#### `pkg/testutil/` - Test Utilities

**Key Files:**
- Isolated test brain creation with `t.TempDir()`
- Fixture files in `fixtures/` directory

**Testing Strategy:**
- Unit tests isolated with `t.TempDir()` and env var overrides
- Integration tests verify multi-component workflows
- **No test affects user's actual brain data**

---

## Key Design Patterns

### Thread Safety

**Config Access:**
```go
// pkg/config/config.go uses sync.RWMutex
func (c *Config) GetCurrentBrain() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.ActiveBrain
}
```

**File Operations:**
```go
// Use atomic writes + directory locks
err := fileutil.WithLock(todoFile, func() error {
    return fileutil.AtomicWriteFile(todoFile, content)
})
```

### Backward Compatibility

**Item ID Generation:**
- IDs use same MD5 algorithm as bash version
- See `pkg/api/id.go` - **Do not change without extensive testing**

**File Formats:**
- Config: `~/.config/brain/config.json` (preserved)
- Markdown: `#captured:`, `#done:` metadata (preserved)
- Task states: `[ ]`, `[>]`, `[-]`, `[x]` (preserved)

### Testing Strategy

**Unit Tests:**
```go
func TestParseProject(t *testing.T) {
    tmpDir := t.TempDir() // Isolated test environment
    projectDir := filepath.Join(tmpDir, "test-project")
    os.MkdirAll(projectDir, 0755)

    // Test logic here

    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
}
```

**Key Principles:**
- Use `t.TempDir()` for test isolation
- Override env vars for config paths (see `pkg/testutil/testutil.go`)
- Test both success and error cases
- Follow naming: `TestFunctionName`, `TestFunctionName_ErrorCase`

---

## Data Flow Example: Refiling

Understanding how a command flows through the codebase:

**User runs:** `brain refile <id> project-name`

1. **`cmd/refile.go`** - Parses CLI args, validates inputs
2. **`pkg/api/dump.go`** - Reads `00_dump.md`, finds item by ID
3. **`pkg/config/config.go`** - Resolves project path
4. **`pkg/fileutil/atomic.go`** - Appends to `todo.md` with file locking
5. **`pkg/fileutil/atomic.go`** - Removes item from dump atomically

**Key Points:**
- Each layer has clear responsibility
- File operations are atomic and locked
- Configuration access is thread-safe

---

## Contributing Guidelines

### Code Style

- **Format:** Run `make fmt` before committing (uses `gofmt`)
- **Linting:** Run `make lint` (uses `golangci-lint`)
- **Naming:** Follow Go conventions (exported vs unexported)
- **Comments:** Document exported functions with doc comments

### Testing Requirements

**All new features and bug fixes MUST include unit tests.**

**Test Placement:**
- Place tests in `*_test.go` files alongside the code
- Example: `dump.go` → `dump_test.go`

**Test Structure:**
```go
func TestFunctionName(t *testing.T) {
    tmpDir := t.TempDir()
    // Setup test environment

    // Execute function under test
    result, err := FunctionName(...)

    // Assertions
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

**Test Isolation:**
- Never affect user data
- Use `t.TempDir()` for all test brains
- Override env vars: `BRAIN_CONFIG_DIR`, `BRAIN_ROOT`, `BRAIN_SYMLINK`
- See `pkg/testutil/testutil.go` for test brain setup helpers

### Pull Request Process

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Make your changes**
   - Write code
   - Add tests
   - Update documentation if needed
4. **Run checks**
   ```bash
   make check
   make test-all
   ```
5. **Commit with descriptive messages**
   ```
   feat: add natural language date parsing

   - Supports "tomorrow", "+3d", "next-friday"
   - Integrates with brain plan and brain todo due
   - Adds pkg/dateutil package
   ```
6. **Push and open PR**
   ```bash
   git push origin feature/my-feature
   ```
7. **Respond to review feedback**

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

<optional body>

<optional footer>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `style:` - Code style (formatting, no logic change)
- `refactor:` - Code restructuring (no behavior change)
- `test:` - Adding or updating tests
- `chore:` - Maintenance (dependencies, build config)

**Examples:**
```
feat(todo): add priority filtering to todo ls

fix(refile): handle empty dump file gracefully

docs: update installation guide with Homebrew instructions
```

---

## API Documentation

### Programmatic Access

Local Brain provides JSON output for scripting:

```bash
# Get dump items
brain dump ls --json

# List projects
brain project list --json

# List todos
brain todo ls --json
```

### Go Package API

Full API documentation available at:
- [pkg.go.dev/github.com/SanderMoon/local-brain](https://pkg.go.dev/github.com/SanderMoon/local-brain)

**Key Packages:**

#### `pkg/api`

Core business logic functions:

```go
// Parse all todos from active projects
func ParseAllTodos(activeDir string, includeCompleted bool) ([]TodoItem, error)

// Generate stable item ID
func GenerateTaskID(lineNum int, content string, mtime int64) string

// List projects with metadata
func ListProjects(activeDir string, focusedProject string) ([]Project, error)
```

#### `pkg/config`

Configuration management:

```go
// Load configuration
func Load() (*Config, error)

// Get active brain path
func (c *Config) GetCurrentBrainPath() (string, error)

// Set focused project
func (c *Config) SetFocusedProject(name string) error
```

#### `pkg/fileutil`

File operations:

```go
// Atomic file write
func AtomicWriteFile(path string, data []byte) error

// Execute function with file lock
func WithLock(path string, fn func() error) error
```

---

## Important Notes for Contributors

### ID Generation Must Match Bash

The `pkg/api/id.go` GenerateItemID function uses MD5 hashing that **must** match the original bash implementation for backward compatibility.

**DO NOT change the algorithm without:**
- Extensive testing with existing brain data
- Migration path for existing users
- Approval from maintainers

### Atomic File Operations

When modifying dump or project files, **always** use:
- `pkg/fileutil.AtomicWrite()` - Crash-safe writes
- `pkg/fileutil.WithLock()` - Prevent race conditions

**Why:**
- Users may run multiple brain commands concurrently
- File corruption must be prevented
- Data loss is unacceptable

### Linter Requirements

All code must pass `golangci-lint`:
- Check error returns (errcheck)
- No empty branches (staticcheck)
- Handle `defer func() { _ = f.Close() }()` for non-critical errors

### Version Injection

Version info is injected at build time via ldflags:

```go
// main.go
var (
    version = "dev"
    commit  = "unknown"
    date    = "unknown"
)
```

Build command:
```bash
go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123 -X main.date=2026-01-28"
```

---

## Release Process

*For maintainers only*

### Prerequisites

1. **GoReleaser installed:**
   ```bash
   make install-goreleaser
   ```

2. **GitHub token:**
   ```bash
   export GITHUB_TOKEN="your_github_token"
   ```

3. **Clean working directory:**
   ```bash
   git status  # Should be clean
   ```

4. **All tests passing:**
   ```bash
   make test-all
   make test-race
   ```

### Release Steps

#### 1. Update Version Information

```bash
# Run full test suite
make check

# Verify build
make build
./brain --version
```

#### 2. Create Git Tag

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR:** Incompatible API changes
- **MINOR:** Backwards-compatible functionality
- **PATCH:** Backwards-compatible bug fixes

```bash
VERSION="v2.0.0"

# Create and push the tag
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
```

#### 3. GitHub Actions (Automated)

Once you push the tag, GitHub Actions automatically:
1. Runs all tests
2. Builds binaries for all platforms
3. Creates a GitHub release
4. Uploads artifacts
5. Updates the Homebrew tap (if configured)

Monitor at: https://github.com/SanderMoon/local-brain/actions

#### 4. Manual Release (Alternative)

```bash
# Create a snapshot (test without pushing)
make snapshot

# Check the artifacts
ls -la dist/

# Create a real release (requires git tag)
make release
```

### Testing a Release Locally

```bash
# Create snapshot release
make snapshot

# Test the binary for your platform
dist/brain_darwin_arm64/brain --version
```

### Post-Release Steps

1. **Verify the release** at GitHub releases page
2. **Test installation methods:**
   ```bash
   # Go install
   go install github.com/SanderMoon/local-brain@latest

   # Homebrew
   brew upgrade brain
   ```
3. **Update documentation** if needed
4. **Announce the release** (GitHub Discussions, social media)

### Release Checklist

- [ ] All tests passing (`make test-all`)
- [ ] Race detector clean (`make test-race`)
- [ ] Code formatted (`make fmt`)
- [ ] Linter passing (`make lint`)
- [ ] Version number decided
- [ ] Documentation updated
- [ ] Git tag created and pushed
- [ ] GitHub Actions workflow completed
- [ ] Release artifacts available
- [ ] Installation methods tested
- [ ] Release announced

### Rollback Procedure

If a release has critical issues:

```bash
# Delete the tag locally
git tag -d $VERSION

# Delete the tag remotely
git push origin :refs/tags/$VERSION

# Fix the issue and re-release with patch version
```

---

## Distribution

### Distribution Methods

1. **Homebrew** (macOS/Linux) - Recommended
2. **Go install** - For Go developers
3. **GitHub Releases** - Pre-built binaries
4. **Build from source** - For contributors

### Build Artifacts

GoReleaser generates:
- macOS (Intel/ARM)
- Linux (x64/ARM)
- Windows (future)
- Checksums for verification
- Homebrew formula auto-update

See `.goreleaser.yml` for configuration.

---

## Resources

- **GitHub Repository:** [https://github.com/SanderMoon/local-brain](https://github.com/SanderMoon/local-brain)
- **Issues:** [https://github.com/SanderMoon/local-brain/issues](https://github.com/SanderMoon/local-brain/issues)
- **Discussions:** [https://github.com/SanderMoon/local-brain/discussions](https://github.com/SanderMoon/local-brain/discussions)
- **Go Docs:** [https://pkg.go.dev/github.com/SanderMoon/local-brain](https://pkg.go.dev/github.com/SanderMoon/local-brain)
- **GoReleaser:** [https://goreleaser.com](https://goreleaser.com)
- **Cobra CLI:** [https://github.com/spf13/cobra](https://github.com/spf13/cobra)

---

## Getting Help

- **General questions:** GitHub Discussions
- **Bug reports:** GitHub Issues
- **Feature requests:** GitHub Issues with "enhancement" label
- **Contributing questions:** Open an issue or discussion

---

## License

MIT License - See [LICENSE](https://github.com/SanderMoon/local-brain/blob/main/LICENSE) file for details.
