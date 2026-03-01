// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package navigator

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"

	"github.com/s3rj1k/nav/internal/config"
)

// Entry represents a single file or directory in the listing.
type Entry struct {
	Name  string
	IsDir bool
}

// Entries implements fuzzy.Source for entry slices.
type Entries []Entry

func (e Entries) String(i int) string { return e[i].Name }

func (e Entries) Len() int { return len(e) }

// Sort sorts entries with directories first, then alphabetically by name.
func (e Entries) Sort() {
	slices.SortFunc(e, func(a, b Entry) int {
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Name, b.Name)
	})
}

// Model holds the entire application state for the Bubble Tea TUI.
type Model struct {
	Path          string            // Current working directory being displayed.
	Selected      string            // Full path chosen by the user on exit.
	Query         string            // Active fuzzy search query typed by the user.
	Error         error             // Last directory-read error, shown in the view.
	Entries       Entries           // All directory entries after filter mode is applied.
	Filtered      Entries           // Subset of Entries matching the current query.
	Cursor        int               // Index of the highlighted entry in Filtered.
	Width         int               // Terminal width in columns.
	Height        int               // Terminal height in rows.
	DisplayHeight int               // Maximum TUI height (from NAV_HEIGHT or default).
	Filter        config.FilterMode // Active filter mode (dirs, files, or both).
	Canceled      bool              // True when the user exits without selecting.
	ShowHelp      bool              // True when the help bar is visible.
}

// LoadEntries reads the current directory and populates the entry list.
func (m *Model) LoadEntries() {
	// Reset state
	m.Entries = nil
	m.Error = nil

	// Read directory contents
	entries, err := os.ReadDir(m.Path)
	if err != nil {
		m.Error = err
		return
	}

	// Filter entries by mode.
	for _, e := range entries {
		name := e.Name()
		isDir := e.IsDir()

		// Apply filter mode: dirs only, files only, or both.
		switch m.Filter {
		case config.FilterDirs:
			if !isDir {
				continue
			}
		case config.FilterFiles:
			if isDir {
				continue
			}
		case config.FilterBoth:
			// show all entries
		}

		m.Entries = append(m.Entries, Entry{
			Name:  name,
			IsDir: isDir,
		})
	}

	m.Entries.Sort()

	// Prepend "." (current directory) entry for dirs and both modes
	if m.Filter != config.FilterFiles {
		m.Entries = append([]Entry{{Name: ".", IsDir: true}}, m.Entries...)
	}

	// Apply current filter query.
	m.ApplyFilter()
}

// ApplyFilter updates the Filtered list by fuzzy-matching Entries against Query.
// When Query is empty all entries are shown. The cursor is clamped to stay in bounds.
func (m *Model) ApplyFilter() {
	if m.Query == "" {
		m.Filtered = m.Entries
	} else {
		matches := fuzzy.FindFrom(m.Query, m.Entries)
		m.Filtered = make(Entries, len(matches))

		for i, match := range matches {
			m.Filtered[i] = m.Entries[match.Index]
		}
	}

	m.ClampCursor()
}

// ClampCursor ensures the cursor stays within the bounds of the filtered list.
func (m *Model) ClampCursor() {
	if m.Cursor >= len(m.Filtered) {
		m.Cursor = max(0, len(m.Filtered)-1)
	}
}

// Init requests the current terminal size so the first View has valid dimensions.
func (*Model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

// Update handles keyboard input and window resize events.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc": // quit with cancellation
			m.Canceled = true
			return m, tea.Quit

		case "f1": // toggle help bar
			m.ShowHelp = !m.ShowHelp
			return m, nil

		case "up": // move cursor up, wrap to bottom
			if m.Cursor > 0 {
				m.Cursor--
			} else if len(m.Filtered) > 0 {
				m.Cursor = len(m.Filtered) - 1
			}

		case "down": // move cursor down, wrap to top
			if m.Cursor < len(m.Filtered)-1 {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "left": // navigate to parent directory
			parent := filepath.Dir(m.Path)
			if parent != m.Path {
				m.Path = parent
				m.Query = ""
				m.LoadEntries()
			}

		case "right": // enter selected directory
			if len(m.Filtered) == 0 {
				return m, nil
			}
			selected := m.Filtered[m.Cursor]
			if selected.IsDir {
				m.Path = filepath.Join(m.Path, selected.Name)
				m.Query = ""
				m.Cursor = 0
				m.LoadEntries()
			}

		case "enter": // select current item and quit
			if len(m.Filtered) > 0 {
				selected := m.Filtered[m.Cursor]
				m.Selected = filepath.Join(m.Path, selected.Name)
				return m, tea.Quit
			}
			m.Selected = m.Path
			return m, tea.Quit

		case "backspace": // delete last character from filter query
			if m.Query != "" {
				m.Query = m.Query[:len(m.Query)-1]
				m.ApplyFilter()
			}

		case "ctrl+f": // cycle filter mode: dirs -> files -> both
			m.Filter = (m.Filter + 1) % config.FilterCount
			m.LoadEntries()

		default: // append character to filter query
			if len(msg.String()) == 1 {
				m.Query += msg.String()
				m.ApplyFilter()
				m.Cursor = 0
			}
		}
	}

	return m, nil
}
