#!/usr/bin/env bash

# brain-config.sh - Configuration management library for Local Brain
# Handles multiple brain locations, switching, and configuration persistence

CONFIG_DIR="$HOME/.config/brain"
CONFIG_FILE="$CONFIG_DIR/config.json"
BRAIN_SYMLINK="$HOME/brain"

# Ensure jq is available
check_jq() {
    if ! command -v jq &> /dev/null;
    then
        echo "Error: jq is required but not found." >&2
        echo "Please install jq: brew install jq (macOS) or apt install jq (Linux)" >&2
        exit 1
    fi
}

# Ensure config directory exists
ensure_config_dir() {
    mkdir -p "$CONFIG_DIR"
}

# Initialize config file if it doesn't exist
init_config_file() {
    check_jq
    ensure_config_dir

    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo '{ "current": null, "brains": {} }' | jq '.' > "$CONFIG_FILE"
    fi
}

# Get current brain name
get_current_brain() {
    init_config_file
    jq -r '.current // empty' "$CONFIG_FILE"
}

# Get path for a specific brain
get_brain_path() {
    local brain_name="$1"
    init_config_file
    jq -r --arg name "$brain_name" '.brains[$name].path // empty' "$CONFIG_FILE"
}

# Get path to current active brain
get_current_brain_path() {
    local current_brain=$(get_current_brain)

    if [[ -z "$current_brain" ]]; then
        # Fallback to default location if no brain configured
        echo "$HOME/brain"
        return 0
    fi

    get_brain_path "$current_brain"
}

# List all configured brains
list_brains() {
    init_config_file
    jq -r '.brains | keys[]' "$CONFIG_FILE"
}

# Check if a brain exists
brain_exists() {
    local brain_name="$1"
    local path=$(get_brain_path "$brain_name")
    [[ -n "$path" ]]
}

# Add a new brain to config
add_brain() {
    local brain_name="$1"
    local brain_path="$2"

    ensure_config_dir
    init_config_file

    # Expand ~ to full path
    brain_path="${brain_path/#\~/$HOME}"

    local created=$(date +"%Y-%m-%d")

    # Use a temporary file to ensure atomic write
    local tmp_file=$(mktemp)
    
jq --arg name "$brain_name" \
       --arg path "$brain_path" \
       --arg created "$created" \
       '.brains[$name] = {"path": $path, "created": $created}' \
       "$CONFIG_FILE" > "$tmp_file" && mv "$tmp_file" "$CONFIG_FILE"
}

# Set current brain
set_current_brain() {
    local brain_name="$1"

    init_config_file

    if ! brain_exists "$brain_name"; then
        return 1
    fi

    local tmp_file=$(mktemp)
    jq --arg name "$brain_name" '.current = $name' "$CONFIG_FILE" > "$tmp_file" && mv "$tmp_file" "$CONFIG_FILE"

    # Update symlink
    update_brain_symlink "$brain_name"
}

# Update ~/brain symlink to point to current brain
update_brain_symlink() {
    local brain_name="$1"
    local brain_path=$(get_brain_path "$brain_name")

    if [[ -z "$brain_path" ]]; then
        return 1
    fi

    # Remove old symlink if it exists
    if [[ -L "$BRAIN_SYMLINK" ]]; then
        rm "$BRAIN_SYMLINK"
    elif [[ -e "$BRAIN_SYMLINK" ]] && [[ ! -d "$BRAIN_SYMLINK" ]]; then
        # If it's a file (not directory), warn and don't replace
        echo "Warning: $BRAIN_SYMLINK exists and is not a symlink" >&2
        return 1
    fi

    # Create new symlink
    ln -sf "$brain_path" "$BRAIN_SYMLINK"
}

# Set focused project for current brain
set_focused_project() {
    local project_name="$1"
    local current_brain=$(get_current_brain)

    init_config_file

    local tmp_file=$(mktemp)
    jq --arg brain "$current_brain" --arg project "$project_name" \
       '.brains[$brain].focus = $project' \
       "$CONFIG_FILE" > "$tmp_file" && mv "$tmp_file" "$CONFIG_FILE"
}

# Get focused project for current brain
get_focused_project() {
    local current_brain=$(get_current_brain)
    init_config_file
    jq -r --arg brain "$current_brain" '.brains[$brain].focus // empty' "$CONFIG_FILE"
}

# Get linked repositories for a project
# Returns list of local paths in ~/dev
get_linked_repos() {
    local project_name="$1"
    local brain_path=$(get_current_brain_path)
    local project_dir="$brain_path/01_active/$project_name"
    local repos_file="$project_dir/.repos"
    local dev_dir="$HOME/dev"

    if [[ ! -f "$repos_file" ]]; then
        return 0
    fi

    while IFS= read -r git_url; do
        # Skip empty lines and comments
        [[ -z "$git_url" ]] && continue
        [[ "$git_url" =~ ^# ]] && continue

        # Extract repo name from URL (Same logic as brain-project)
        local repo_name=""
        if [[ "$git_url" =~ /([^/]+)\.git$ ]]; then
            repo_name="${BASH_REMATCH[1]}"
        elif [[ "$git_url" =~ /([^/]+)$ ]]; then
            repo_name="${BASH_REMATCH[1]}"
        elif [[ "$git_url" =~ :([^/]+)\.git$ ]]; then
            repo_name="${BASH_REMATCH[1]}"
        fi

        if [[ -n "$repo_name" ]]; then
            echo "$dev_dir/$repo_name"
        fi
    done < "$repos_file"
}

# Export functions for use in other scripts
export -f get_current_brain
export -f get_brain_path
export -f get_current_brain_path
export -f list_brains
export -f brain_exists
export -f add_brain
export -f set_current_brain
export -f update_brain_symlink
export -f init_config_file
export -f set_focused_project
export -f get_focused_project
export -f get_linked_repos