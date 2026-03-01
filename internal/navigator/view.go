// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package navigator

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/s3rj1k/nav/internal/config"
)

// View renders the complete TUI frame: status bar, optional help bar,
// and the entry list. Bubble Tea's inline renderer handles cursor
// positioning and screen management.
func (m *Model) View() tea.View {
	// Wait for the first WindowSizeMsg before rendering; without terminal
	// dimensions the layout calculation cannot determine item count.
	if m.Height == 0 {
		return tea.NewView("")
	}

	var b strings.Builder

	itemLines, scrollOffset := m.Layout()
	m.RenderStatusBar(&b)
	m.RenderHelpBar(&b)
	m.RenderItems(&b, scrollOffset, itemLines)

	return tea.NewView(b.String())
}

// HelpBarHeight returns 1 when the help bar is visible, 0 otherwise.
func (m *Model) HelpBarHeight() int {
	if m.ShowHelp {
		return 1
	}

	return 0
}

// EffectiveHeight returns the display height clamped to the terminal size.
func (m *Model) EffectiveHeight() int {
	dh := m.DisplayHeight
	if dh <= 0 {
		dh = config.DefaultHeight
	}

	// Leave at least one row for the shell prompt.
	if m.Height > 0 && dh > m.Height-1 {
		dh = m.Height - 1
	}

	return dh
}

// Layout calculates how many entry lines fit on screen and the scroll offset
// needed to keep the cursor visible.
func (m *Model) Layout() (itemLines, scrollOffset int) {
	// Reserve rows for the status bar and optional help bar.
	itemLines = m.EffectiveHeight() - 1 - m.HelpBarHeight()
	if itemLines < config.MinItems {
		itemLines = config.MinItems + 2 //nolint:mnd // ensure a usable minimum
	}

	// Compute scroll offset so the cursor stays within the visible window.
	scrollOffset = m.ScrollOffset(itemLines)

	return itemLines, scrollOffset
}

// ScrollOffset returns the first visible index so that the cursor row
// is always within the rendered page.
func (m *Model) ScrollOffset(itemLines int) int {
	if len(m.Filtered) <= itemLines {
		return 0
	}

	offset := 0
	if m.Cursor >= itemLines {
		offset = m.Cursor - itemLines + 1
	}

	// Clamp so we never scroll past the end.
	if offset+itemLines > len(m.Filtered) {
		offset = len(m.Filtered) - itemLines
	}

	return offset
}

// RenderStatusBar writes the top line: current path, cursor position counter,
// active query highlight, and filter mode indicator.
func (m *Model) RenderStatusBar(b *strings.Builder) {
	// Cursor position (1-based for display).
	cursorNum := 0
	if len(m.Filtered) > 0 {
		cursorNum = m.Cursor + 1
	}

	// Path and counter.
	b.WriteString(config.ColorPath)
	b.WriteString(m.Path)
	b.WriteString(config.ColorDim)
	fmt.Fprintf(b, " < %d/%d", cursorNum, len(m.Filtered))

	// Query indicator: shown only while a search is active.
	if m.Query != "" {
		b.WriteString(" [")
		b.WriteString(config.ColorAccent)
		b.WriteString(m.Query)
		b.WriteString(config.ColorDim + "]")
	}

	// Filter mode indicator: hidden when showing both files and dirs.
	if m.Filter != config.FilterBoth {
		b.WriteString(" {")
		b.WriteString(config.ColorAccent)
		b.WriteString(m.Filter.String())
		b.WriteString(config.ColorDim)
		b.WriteString("}")
	}

	b.WriteString(config.StyleReset + "\n")
}

// RenderHelpBar writes a single line of keybinding hints.
// Does nothing when help is toggled off.
func (m *Model) RenderHelpBar(b *strings.Builder) {
	if !m.ShowHelp {
		return
	}

	type hint struct {
		key   string
		label string
	}

	hints := []hint{
		{"↑", "up"},
		{"↓", "down"},
		{"←", "parent"},
		{"→", "enter"},
		{"⏎", "select"},
		{"type", "search"},
		{"⌫", "clear"},
		{"ctrl+f", "filter"},
		{"f1", "help"},
		{"esc", "quit"},
	}

	var helpLen int

	for i, h := range hints {
		// Key in highlight color, label in dim.
		b.WriteString(config.ColorHighlight + " ")
		b.WriteString(h.key)
		b.WriteString(config.ColorDim + " ")
		b.WriteString(h.label)
		helpLen += 2 + utf8.RuneCountInString(h.key) + 1 + utf8.RuneCountInString(h.label)

		// Separator between hints, but not after the last one.
		if i < len(hints)-1 {
			b.WriteString(config.ColorDim + config.HelpSeparator)
			helpLen += utf8.RuneCountInString(config.HelpSeparator)
		}
	}

	// Pad the rest of the line so the background color fills the row.
	for range max(0, m.Width-helpLen) {
		b.WriteString(" ")
	}

	b.WriteString(config.StyleReset + "\n")
}

// RenderItems writes the visible portion of the entry list, or a
// placeholder message for error / empty / no-match states.
func (m *Model) RenderItems(b *strings.Builder, scrollOffset, itemLines int) {
	switch {
	case m.Error != nil:
		m.RenderError(b)
	case len(m.Filtered) == 0:
		m.RenderEmpty(b)
	default:
		m.RenderEntryList(b, scrollOffset, itemLines)
	}
}

// RenderError writes a red error message line.
func (m *Model) RenderError(b *strings.Builder) {
	b.WriteString(config.ColorError + "Error: ")
	b.WriteString(m.Error.Error())
	b.WriteString(config.StyleReset + "\n")
}

// RenderEmpty writes a placeholder when there are no entries to show.
func (m *Model) RenderEmpty(b *strings.Builder) {
	if m.Query != "" {
		b.WriteString(config.ColorDim + "No matches for " + config.ColorHighlight)
		b.WriteString(m.Query)
	} else {
		b.WriteString(config.ColorDim + "(empty)")
	}

	b.WriteString(config.StyleReset + "\n")
}

// RenderEntryList writes the paginated file/directory rows with the
// cursor-selected entry highlighted.
func (m *Model) RenderEntryList(b *strings.Builder, scrollOffset, itemLines int) {
	end := min(scrollOffset+itemLines, len(m.Filtered))

	for i := scrollOffset; i < end; i++ {
		m.RenderEntry(b, i)
	}
}

// RenderEntry writes a single entry line with appropriate styling.
func (m *Model) RenderEntry(b *strings.Builder, index int) {
	e := m.Filtered[index]
	selected := m.Cursor == index

	// Cursor prefix or indentation.
	if selected {
		b.WriteString(config.StyleSelected)
	} else {
		b.WriteString(config.ColorDim + "  ")
	}

	// Entry name; directories get a trailing slash and path color.
	if e.IsDir {
		if !selected {
			b.WriteString(config.ColorPath)
		}

		b.WriteString(e.Name)
		b.WriteString("/")
	} else {
		b.WriteString(e.Name)
	}

	b.WriteString(config.StyleReset + "\n")
}
