# nav

An interactive terminal file and directory navigator with shell integration. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it renders a TUI on stderr and prints the selected path to stdout, enabling usage like `cd $(nav)`.

## Install

### From source

```bash
go install github.com/s3rj1k/nav/cmd/nav@latest
```

### From release binaries

Download the appropriate binary from the [Releases](https://github.com/s3rj1k/nav/releases) page.

On macOS, you need to remove the quarantine attribute before running the binary:

```bash
xattr -dr com.apple.quarantine /path/to/nav
```

## Usage

Launch the navigator by running `nav` or pressing **Shift+Tab** (after shell integration).

### Keybindings

| Key                | Action                                 |
|--------------------|----------------------------------------|
| `↑` / `↓`          | Move cursor up / down                  |
| `←`                | Go to parent directory                 |
| `→`                | Enter directory                        |
| `Enter`            | Select (confirm selection)             |
| Type any character | Fuzzy search / filter entries          |
| `Backspace`        | Clear search query                     |
| `Ctrl+F`           | Cycle filter mode: dirs → files → both |
| `F1`               | Toggle help bar                        |
| `Esc`              | Quit without selecting                 |

### Environment variables

Appearance can be customized via environment variables. Run `nav --help` for the full list.

## Shell integration

Add to your shell profile:

```bash
# bash
eval "$(nav --init-bash)"

# zsh
eval "$(nav --init-zsh)"
```

## License

This project is licensed under the [MIT License](LICENSE).
