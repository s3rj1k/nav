_nav_widget() {
  local selected start_dir
  local before="${READLINE_LINE:0:$READLINE_POINT}"
  local after="${READLINE_LINE:$READLINE_POINT}"
  local current_word="${before##* }"

  # Use the current word as starting directory if it's valid.
  if [[ -d "$current_word" ]]; then
    start_dir="$current_word"
  elif [[ -n "$current_word" && -d "$(dirname "$current_word")" ]]; then
    start_dir="$(dirname "$current_word")"
  else
    start_dir="."
  fi

  selected=$(command "{{.NavPath}}" "$start_dir")
  local ret=$?

  if [[ $ret -eq 0 && -n "$selected" ]]; then
    # Quote the selected path for safe shell use.
    local quoted
    printf -v quoted '%q' "$selected"

    if [[ -n "$current_word" ]]; then
      before="${before:0:$(( ${#before} - ${#current_word} ))}${quoted}"
    else
      before+="${quoted}"
    fi

    READLINE_LINE="${before}${after}"
    READLINE_POINT=${#before}
  fi
}

bind -x '"\e[Z": _nav_widget'
