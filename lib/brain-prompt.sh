#!/usr/bin/env bash

# brain-prompt.sh - Shell prompt integration for Local Brain
# Source this file in your ~/.zshrc or ~/.bashrc to show current brain in prompt

# Get current brain name for prompt display
brain_prompt() {
  local config_file="$HOME/.config/brain/config.json"

  if [[ ! -f "$config_file" ]]; then
    return
  fi

  local current_brain

  # Try jq first, fallback to grep/sed
  if command -v jq &> /dev/null; then
    current_brain=$(jq -r '.current // empty' "$config_file" 2>/dev/null)
  else
    current_brain=$(grep '"current"' "$config_file" | sed 's/.*"current"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' | grep -v "null")
  fi

  if [[ -n "$current_brain" ]] && [[ "$current_brain" != "null" ]]; then
    echo "$current_brain"
  fi
}

# Get current brain name with formatting for prompt
brain_prompt_formatted() {
  local brain=$(brain_prompt)

  if [[ -n "$brain" ]]; then
    # Return formatted string for prompt
    # Customize colors/format as needed
    echo "[${brain}]"
  fi
}

# Export BRAIN_CURRENT environment variable
# This updates dynamically and can be used by other tools
export_brain_current() {
  export BRAIN_CURRENT=$(brain_prompt)
}

# Auto-export on prompt display (for bash/zsh)
if [[ -n "$ZSH_VERSION" ]]; then
  # Zsh: Add to precmd hook
  precmd_brain_export() {
    export_brain_current
  }

  # Add to precmd_functions if not already there
  if [[ -z "${precmd_functions[(r)precmd_brain_export]}" ]]; then
    precmd_functions+=(precmd_brain_export)
  fi
elif [[ -n "$BASH_VERSION" ]]; then
  # Bash: Add to PROMPT_COMMAND
  if [[ ! "$PROMPT_COMMAND" =~ "export_brain_current" ]]; then
    PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND; }export_brain_current"
  fi
fi

# Example usage in your prompt:
#
# For Zsh (add to ~/.zshrc):
#  source ~/.config/brain/lib/brain-prompt.sh
#  PROMPT='$(brain_prompt_formatted) %~ %# '
#
# For Bash (add to ~/.bashrc):
#  source ~/.config/brain/lib/brain-prompt.sh
#  PS1='$(brain_prompt_formatted) \w\$ '
#
# Or use the environment variable:
#  PROMPT='${BRAIN_CURRENT:+[$BRAIN_CURRENT] }%~ %# '
