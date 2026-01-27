#!/usr/bin/env bash

# Release helper script for Local Brain
# Automates the release process with safety checks

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_info() { echo -e "${BLUE}→${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }

print_header() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "$1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Check if we're in the right directory
check_directory() {
  if [[ ! -f "go.mod" ]] || [[ ! -f ".goreleaser.yml" ]]; then
    print_error "Not in the local-brain project root"
    exit 1
  fi
  print_success "In project root"
}

# Check git status
check_git_status() {
  if [[ -n $(git status --porcelain) ]]; then
    print_error "Working directory is not clean"
    git status --short
    exit 1
  fi
  print_success "Working directory is clean"
}

# Check if on main branch
check_branch() {
  BRANCH=$(git rev-parse --abbrev-ref HEAD)
  if [[ "$BRANCH" != "main" ]] && [[ "$BRANCH" != "go-rewrite" ]]; then
    print_warning "Not on main or go-rewrite branch (on: $BRANCH)"
    read -rp "Continue anyway? [y/N] " response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
      exit 1
    fi
  else
    print_success "On $BRANCH branch"
  fi
}

# Run tests
run_tests() {
  print_header "Running Tests"

  print_info "Running unit tests..."
  if ! make test-unit; then
    print_error "Unit tests failed"
    exit 1
  fi
  print_success "Unit tests passed"

  print_info "Running race detector..."
  if ! make test-race; then
    print_error "Race detector found issues"
    exit 1
  fi
  print_success "Race detector clean"

  print_info "Running all tests..."
  if ! make test-all; then
    print_error "Tests failed"
    exit 1
  fi
  print_success "All tests passed"
}

# Validate version number
validate_version() {
  local version=$1

  # Check semantic versioning format
  if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    print_error "Invalid version format: $version"
    echo "Expected format: vX.Y.Z or vX.Y.Z-beta.N"
    exit 1
  fi

  # Check if tag already exists
  if git rev-parse "$version" >/dev/null 2>&1; then
    print_error "Tag $version already exists"
    exit 1
  fi

  print_success "Version format valid: $version"
}

# Get version input
get_version() {
  echo ""
  print_info "Current git tags:"
  git tag | tail -5 | sed 's/^/  /'

  echo ""
  read -rp "Enter new version (e.g., v2.0.0): " VERSION

  if [[ -z "$VERSION" ]]; then
    print_error "Version cannot be empty"
    exit 1
  fi

  validate_version "$VERSION"
}

# Generate changelog entry (placeholder)
generate_changelog() {
  print_header "Generating Changelog"

  # Get commits since last tag
  LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

  if [[ -n "$LAST_TAG" ]]; then
    print_info "Changes since $LAST_TAG:"
    git log "$LAST_TAG"..HEAD --oneline --no-merges | sed 's/^/  /'
  else
    print_info "First release - showing recent commits:"
    git log --oneline --no-merges -10 | sed 's/^/  /'
  fi

  echo ""
}

# Confirm release
confirm_release() {
  print_header "Release Summary"
  echo "  Version: $VERSION"
  echo "  Branch:  $(git rev-parse --abbrev-ref HEAD)"
  echo "  Commit:  $(git rev-parse --short HEAD)"
  echo ""

  read -rp "Proceed with release? [y/N] " response
  if [[ ! "$response" =~ ^[Yy]$ ]]; then
    print_info "Release cancelled"
    exit 0
  fi
}

# Create and push tag
create_tag() {
  print_header "Creating Git Tag"

  print_info "Creating tag $VERSION..."
  git tag -a "$VERSION" -m "Release $VERSION"
  print_success "Tag created"

  print_info "Pushing tag to remote..."
  git push origin "$VERSION"
  print_success "Tag pushed"
}

# Monitor release
monitor_release() {
  print_header "Release in Progress"

  REPO_URL=$(git config --get remote.origin.url | sed 's/\.git$//' | sed 's/git@github.com:/https:\/\/github.com\//')

  echo ""
  print_success "Release initiated!"
  echo ""
  print_info "Monitor the release at:"
  echo "  Actions:  $REPO_URL/actions"
  echo "  Releases: $REPO_URL/releases"
  echo ""
  print_info "The release will be ready in a few minutes."
  echo ""
}

# Create snapshot for testing
create_snapshot() {
  print_header "Creating Snapshot Release"

  print_info "Building snapshot (no tags required)..."
  if ! make snapshot; then
    print_error "Snapshot build failed"
    exit 1
  fi

  print_success "Snapshot created in dist/"

  echo ""
  print_info "Test the binary:"
  echo "  dist/brain_darwin_arm64/brain --version"
  echo "  dist/brain_linux_amd64/brain --version"
  echo ""
}

# Main menu
show_menu() {
  print_header "Local Brain Release Tool"

  echo "What would you like to do?"
  echo ""
  echo "  1) Create a release (full process)"
  echo "  2) Create a snapshot (test build)"
  echo "  3) Run tests only"
  echo "  4) Check release readiness"
  echo "  0) Exit"
  echo ""

  read -rp "Select option [0-4]: " choice

  case "$choice" in
    1)
      do_release
      ;;
    2)
      check_directory
      create_snapshot
      ;;
    3)
      check_directory
      run_tests
      ;;
    4)
      check_directory
      check_git_status
      check_branch
      run_tests
      print_success "Ready for release!"
      ;;
    0)
      print_info "Goodbye!"
      exit 0
      ;;
    *)
      print_error "Invalid option"
      exit 1
      ;;
  esac
}

# Full release process
do_release() {
  check_directory
  check_git_status
  check_branch
  run_tests
  get_version
  generate_changelog
  confirm_release
  create_tag
  monitor_release
}

# Run main menu
show_menu

