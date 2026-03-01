nav-widget() {
  local selected start_dir
  local current_word="${LBUFFER##* }"

  # Use the current word as starting directory if it's valid.
  if [[ -d "$current_word" ]]; then
    start_dir="$current_word"
  elif [[ -n "$current_word" && -d "${current_word:h}" ]]; then
    start_dir="${current_word:h}"
  else
    start_dir="."
  fi

  selected=$(command "{{.NavPath}}" "$start_dir")
  local ret=$?

  if [[ $ret -eq 0 && -n "$selected" ]]; then
    if [[ -n "$current_word" ]]; then
      LBUFFER="${LBUFFER:0:$(( ${#LBUFFER} - ${#current_word} ))}${(q)selected}"
    else
      LBUFFER+="${(q)selected}"
    fi
  fi

  zle reset-prompt
}
zle -N nav-widget
bindkey '^[[Z' nav-widget
