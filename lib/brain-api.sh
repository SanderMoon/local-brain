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
# Handles both one-liner tasks and multi-line indented notes
# Args: inbox_path
# Output format: START_LINE|END_LINE|TYPE|TITLE|ID
parse_dump_items() {
  local inbox="$1"

  [[ ! -f "$inbox" ]] && return 1

  # Get file modification time (cross-platform)
  local mtime=$(stat -f %m "$inbox" 2>/dev/null || stat -c %Y "$inbox" 2>/dev/null)
  local line_num=0
  local in_note=false
  local note_start=0
  local note_title=""

  while IFS= read -r line || [[ -n "$line" ]]; do
    line_num=$((line_num + 1))

    # Check if line is indented (part of note content)
    if [[ "$line" =~ ^[[:space:]]{4} ]] && [[ "$in_note" == "true" ]]; then
      continue
    fi

    # If we were in a note and hit non-indented line, close the note
    if [[ "$in_note" == "true" ]]; then
      local note_id=$(echo -n "${note_start}:${note_title}:${mtime}" | compute_md5 | cut -c1-6)
      echo "${note_start}|$((line_num - 1))|note|${note_title}|${note_id}"
      in_note=false
    fi

    # Skip empty lines and markdown headers
    [[ -z "${line// /}" ]] && continue
    [[ "$line" =~ ^#+ ]] && continue

    # Detect task
    if [[ "$line" =~ ^-\ \[\ \]\ (.+)$ ]]; then
      local task_content="${BASH_REMATCH[1]}"
      local task_id=$(echo -n "${line_num}:${line}:${mtime}" | compute_md5 | cut -c1-6)
      echo "${line_num}|${line_num}|todo|${task_content}|${task_id}"

    # Detect note header
    elif [[ "$line" =~ ^\[Note\]\ (.+)$ ]]; then
      in_note=true
      note_start=$line_num
      note_title="${BASH_REMATCH[1]}"
    fi
  done < "$inbox"

  # Close any remaining note at end of file
  if [[ "$in_note" == "true" ]]; then
    local note_id=$(echo -n "${note_start}:${note_title}:${mtime}" | compute_md5 | cut -c1-6)
    echo "${note_start}|${line_num}|note|${note_title}|${note_id}"
  fi
}

# Convert parsed dump items to JSON
# Input: parse_dump_items output via stdin
# Output: JSON array
dump_to_json() {
  while IFS='|' read -r start_line end_line type title id; do
    # Extract timestamp from title
    local content="$title"
    local timestamp=""
    if [[ "$content" =~ \#captured:([0-9-]+) ]]; then
      timestamp="${BASH_REMATCH[1]}"
      content=$(echo "$content" | sed -E 's/ #captured:[0-9-]+$//')
    fi

    jq -n \
      --arg id "$id" \
      --arg content "$content" \
      --arg type "$type" \
      --arg ts "$timestamp" \
      --argjson start "$start_line" \
      --argjson end "$end_line" \
      '{id: $id, content: $content, type: $type, timestamp: $ts, start_line: $start, end_line: $end}'
  done | jq -s '.'
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

# Convert notes directory to JSON
# Args: notes_dir
# Output: JSON array with name, title, path, created
notes_to_json() {
  local notes_dir="$1"

  echo "["
  local first=true

  for f in "$notes_dir"/*.md; do
    [[ ! -f "$f" ]] && continue

    local name=$(basename "$f" .md)
    local title=$(head -1 "$f" | sed 's/^# //')
    local created=""

    # Extract date from filename or Created: line
    if [[ "$name" =~ ^([0-9]{4}-[0-9]{2}-[0-9]{2})- ]]; then
      created="${BASH_REMATCH[1]}"
    else
      created=$(grep -m1 "^Created:" "$f" | sed 's/Created: //' || true)
    fi

    local json=$(jq -n \
      --arg name "$name" \
      --arg title "$title" \
      --arg path "$f" \
      --arg created "$created" \
      '{name: $name, title: $title, path: $path, created: $created}')

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
export -f projects_to_json
export -f notes_to_json
