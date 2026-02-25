# Shelly AI - Inline Zsh Suggestions
# Source this file in your .zshrc: source ~/.shelly-ai/shelly.zsh

_shelly_suggest() {
  # Skip if buffer is empty or too short
  [[ ${#BUFFER} -lt 3 ]] && return

  local suggestion
  suggestion=$(q --suggest "$BUFFER" 2>/dev/null)
  if [[ -n "$suggestion" && "$suggestion" != "$BUFFER" ]]; then
    POSTDISPLAY="${suggestion#$BUFFER}"
  else
    POSTDISPLAY=""
  fi
}

_shelly_accept() {
  if [[ -n "$POSTDISPLAY" ]]; then
    BUFFER="$BUFFER$POSTDISPLAY"
    POSTDISPLAY=""
    CURSOR=$#BUFFER
  else
    # Fall back to normal tab behavior
    zle expand-or-complete
  fi
}

zle -N _shelly_suggest
zle -N _shelly_accept
autoload -U add-zle-hook-widget
add-zle-hook-widget line-pre-redraw _shelly_suggest
bindkey '^[s' _shelly_accept  # Alt+S to accept suggestion
