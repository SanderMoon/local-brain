#!/usr/bin/env bash

# brain-api.sh - Programmatic API utilities for Local Brain
# Provides functions for non-interactive CLI usage by AI agents

# Cross-platform MD5 computation
compute_md5() {
  if command -v md5 &>/dev/null; then
    # macOS
    md5 | awk '{print $NF}'
  else
    # Linux
    md5sum | awk '{print $1}'
  fi
}

# Generate stable ID for a dump item
# Args: line_num, content, file_mtime
# Returns: 6-char hex ID
generate_item_id() {
  local line_num="$1"
  local content="$2"
  local mtime="$3"

  local hash_input="${line_num}:${content}:${mtime}"
  echo -n "$hash_input" | compute_md5 | cut -c1-6
}

# Parse dump items to pipe-delimited format
# Args: inbox_path
# Output format: LINE_NUM|TYPE|RAW_CONTENT|ID
parse_dump_items() {
  local inbox="$1"

  [[ ! -f "$inbox" ]] && return 1

  # Get file modification time (cross-platform)
  local mtime=$(stat -f %m "$inbox" 2>/dev/null || stat -c %Y "$inbox" 2>/dev/null)
  local line_num=0

  while IFS= read -r line; do
    line_num=$((line_num + 1))

    # Skip empty lines and headers
    [[ -z "${line// /}" ]] && continue
    [[ "$line" =~ ^#+ ]] && continue

    local type=""

    # Detect type
    if [[ "$line" =~ ^-\ \[\ \] ]]; then
      type="todo"
    elif [[ "$line" =~ ^\[Note\] ]]; then
      type="note"
    else
      continue
    fi

    # Generate ID
    local id=$(generate_item_id "$line_num" "$line" "$mtime")

    # Output: LINE_NUM|TYPE|RAW|ID
    echo "${line_num}|${type}|${line}|${id}"
  done < "$inbox"
}

# Convert parsed items to JSON
# Input: parse_dump_items output via stdin
# Output: JSON array with id, content, raw, type, timestamp
dump_to_json() {
  echo "["
  local first=true

  while IFS='|' read -r line_num type raw id; do
    # Extract content (remove prefixes)
    local content="$raw"
    content=$(echo "$content" | sed 's/^- \[ \] //')
    content=$(echo "$content" | sed 's/^\[Note\] //')

    # Extract timestamp
    local timestamp=""
    local capture_pattern='#captured:([0-9-]+)'
    if [[ "$content" =~ $capture_pattern ]]; then
      timestamp="${BASH_REMATCH[1]}"
      content=$(echo "$content" | sed -E 's/ #captured:[0-9-]+$//')
    fi

    # Build JSON with jq (proper escaping)
    local json=$(jq -n \
      --arg id "$id" \
      --arg content "$content" \
      --arg raw "$raw" \
      --arg type "$type" \
      --arg ts "$timestamp" \
      '{id: $id, content: $content, raw: $raw, type: $type, timestamp: $ts}')

    # Comma separator
    [[ "$first" == "false" ]] && echo ","
    first=false

    echo -n "  $json"
  done

  echo ""
  echo "]"
}

# Match project name (exact/case-insensitive/fuzzy)
# Args: pattern, active_dir
# Returns: exact project name
# Exit codes: 0 on success, 4 on error
match_project() {
  local pattern="$1"
  local active_dir="$2"

  local projects=$(find "$active_dir" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort)

  # 1. Exact match
  while IFS= read -r proj; do
    [[ "$proj" == "$pattern" ]] && echo "$proj" && return 0
  done <<< "$projects"

  # 2. Case-insensitive match
  local lower_pattern=$(echo "$pattern" | tr '[:upper:]' '[:lower:]')
  while IFS= read -r proj; do
    local lower_proj=$(echo "$proj" | tr '[:upper:]' '[:lower:]')
    [[ "$lower_proj" == "$lower_pattern" ]] && echo "$proj" && return 0
  done <<< "$projects"

  # 3. Fuzzy prefix match
  local matches=()
  while IFS= read -r proj; do
    [[ "$proj" == "$pattern"* ]] && matches+=("$proj")
  done <<< "$projects"

  if [[ ${#matches[@]} -eq 1 ]]; then
    echo "${matches[0]}"
    return 0
  elif [[ ${#matches[@]} -gt 1 ]]; then
    echo "Error: Ambiguous project. Matches: ${matches[*]}" >&2
    return 4
  fi

  echo "Error: Project not found: $pattern" >&2
  return 4
}

# Refile item by ID (non-interactive)
# Args: id, project_name, type_override (optional)
# Exit codes: 0=success, 3=ID not found, 4=project not found
refile_by_id() {
  local target_id="$1"
  local project_name="$2"
  local type_override="${3:-}"

  local brain_path=$(get_current_brain_path)
  local inbox="$brain_path/00_dump.md"
  local active_dir="$brain_path/01_active"

  # Find item
  local found_line=""
  local found_type=""
  local found_raw=""

  while IFS='|' read -r line_num type raw id; do
    if [[ "$id" == "$target_id" ]]; then
      found_line="$line_num"
      found_type="$type"
      found_raw="$raw"
      break
    fi
  done < <(parse_dump_items "$inbox")

  if [[ -z "$found_line" ]]; then
    echo "Error: Item not found: $target_id" >&2
    return 3
  fi

  # Match project
  local matched_project=$(match_project "$project_name" "$active_dir")
  [[ $? -ne 0 ]] && return 4

  local project_dir="$active_dir/$matched_project"
  local target_type="${type_override:-$found_type}"

  # Route to file
  if [[ "$target_type" == "todo" ]]; then
    local target_file="$project_dir/todo.md"

    # Ensure exists
    if [[ ! -f "$target_file" ]]; then
      cat > "$target_file" << 'EOF'
# Tasks
## Active
## Completed
EOF
    fi

    # Normalize to task format
    local clean_item="$found_raw"
    if [[ "$clean_item" =~ ^\[Note\]\ (.+)$ ]]; then
      clean_item="- [ ] ${BASH_REMATCH[1]}"
    fi

    echo "$clean_item" >> "$target_file"

  elif [[ "$target_type" == "note" ]]; then
    local target_file="$project_dir/notes.md"

    # Ensure exists
    if [[ ! -f "$target_file" ]]; then
      cat > "$target_file" << EOF
# $matched_project
Created: $(date +"%Y-%m-%d")
## Overview
[Description]
## Notes
EOF
    fi

    # Normalize to note format
    local clean_item="$found_raw"
    if [[ "$clean_item" =~ ^-\ \[\ \]\ (.+)$ ]]; then
      clean_item="${BASH_REMATCH[1]}"
    elif [[ "$clean_item" =~ ^\[Note\]\ (.+)$ ]]; then
      clean_item="${BASH_REMATCH[1]}"
    fi

    echo "" >> "$target_file"
    echo "### $(date '+%Y-%m-%d %H:%M')" >> "$target_file"
    echo "$clean_item" >> "$target_file"
  fi

  # Remove from dump (atomic)
  local temp_inbox=$(mktemp)
  local current_line=0

  while IFS= read -r line; do
    current_line=$((current_line + 1))
    [[ "$current_line" -ne "$found_line" ]] && echo "$line" >> "$temp_inbox"
  done < "$inbox"

  mv "$temp_inbox" "$inbox"

  echo "OK: Refiled $target_id to $matched_project/$target_type.md"
  return 0
}

# Convert projects to JSON
# Args: active_dir
# Output: JSON with name, path, focused, repo_count, task_count
projects_to_json() {
  local active_dir="$1"
  local focused=$(get_focused_project)

  echo "["
  local first=true

  for proj_path in "$active_dir"/*; do
    [[ ! -d "$proj_path" ]] && continue

    local name=$(basename "$proj_path")
    local is_focused="false"
    [[ "$name" == "$focused" ]] && is_focused="true"

    # Stats
    local repo_count=0
    local task_count=0

    if [[ -f "$proj_path/.repos" ]]; then
      repo_count=$(grep -vc "^#" "$proj_path/.repos" 2>/dev/null || true)
      [[ -z "$repo_count" ]] && repo_count=0
    fi

    if [[ -f "$proj_path/todo.md" ]]; then
      task_count=$(grep -c "^\s*- \[ \]" "$proj_path/todo.md" 2>/dev/null || true)
      [[ -z "$task_count" ]] && task_count=0
    fi

    # Build JSON
    local json=$(jq -n \
      --arg name "$name" \
      --arg path "$proj_path" \
      --argjson focused "$is_focused" \
      --argjson repos "$repo_count" \
      --argjson tasks "$task_count" \
      '{name: $name, path: $path, focused: $focused, repo_count: $repos, task_count: $tasks}')

    [[ "$first" == "false" ]] && echo ","
    first=false

    echo -n "  $json"
  done

  echo ""
  echo "]"
}

# Export functions
export -f compute_md5
export -f generate_item_id
export -f parse_dump_items
export -f dump_to_json
export -f match_project
export -f refile_by_id
export -f projects_to_json
