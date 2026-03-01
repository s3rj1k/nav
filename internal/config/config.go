// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package config

import (
	"cmp"
	"os"
	"strconv"
	"syscall"
)

// FilterMode controls which entry types are shown in the directory listing.
type FilterMode int

// FilterCount is the total number of FilterMode values, used for cycling.
const FilterCount = 3

const (
	FilterDirs  FilterMode = iota // Show only directories.
	FilterFiles                   // Show only files.
	FilterBoth                    // Show both files and directories.
)

const (
	// Esc is the ASCII escape character used to build ANSI sequences.
	Esc = "\x1b"

	// ANSI SGR sequences for text styling.
	StyleBold  = Esc + "[1m" // Enable bold text.
	StyleReset = Esc + "[0m" // Reset all text attributes to default.

	// Default CSI color parameter strings (without the ESC prefix).
	// Users override these via NAV_COLOR_* environment variables.
	CSIBlue  = "[94m"
	CSIDim   = "[90m"
	CSIGreen = "[92m"
	CSIWhite = "[97m"
	CSIRed   = "[31m"

	// DefaultSelectedPrefix is the string drawn before the highlighted entry.
	DefaultSelectedPrefix = "▌ "

	// DefaultHelpSeparator is the string drawn between help bar hints.
	DefaultHelpSeparator = "│"

	// ExitCanceled follows the Unix convention of 128 + signal number.
	ExitCanceled = 128 + int(syscall.SIGINT)

	// MinItems is the minimum number of entry lines rendered in the list area.
	MinItems = 3

	// DefaultHeight is the default number of TUI rows when NAV_HEIGHT is unset.
	DefaultHeight = 15
)

const (
	// Environment variable names for runtime configuration.
	EnvHeight         = "NAV_HEIGHT"
	EnvColorPath      = "NAV_COLOR_PATH"
	EnvColorDim       = "NAV_COLOR_DIM"
	EnvColorAccent    = "NAV_COLOR_ACCENT"
	EnvColorHighlight = "NAV_COLOR_HIGHLIGHT"
	EnvColorError     = "NAV_COLOR_ERROR"
	EnvSelectedPrefix = "NAV_SELECTED_PREFIX"
	EnvHelpSeparator  = "NAV_HELP_SEPARATOR"
)

// EnvInt reads an environment variable as an integer, returning fallback if unset or invalid.
func EnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if h, err := strconv.Atoi(v); err == nil {
			return h
		}
	}

	return fallback
}

// Resolved color and style variables. Each reads its NAV_COLOR_* env var at
// program startup and falls back to the default CSI value when unset.
var (
	ColorPath      = Esc + cmp.Or(os.Getenv(EnvColorPath), CSIBlue)
	ColorDim       = Esc + cmp.Or(os.Getenv(EnvColorDim), CSIDim)
	ColorAccent    = Esc + cmp.Or(os.Getenv(EnvColorAccent), CSIGreen)
	ColorHighlight = Esc + cmp.Or(os.Getenv(EnvColorHighlight), CSIWhite)
	ColorError     = Esc + cmp.Or(os.Getenv(EnvColorError), CSIRed)
	SelectedPrefix = cmp.Or(os.Getenv(EnvSelectedPrefix), DefaultSelectedPrefix)
	HelpSeparator  = cmp.Or(os.Getenv(EnvHelpSeparator), DefaultHelpSeparator)
	StyleSelected  = ColorAccent + SelectedPrefix + ColorHighlight + StyleBold

	// DisplayHeight is the number of TUI rows, parsed from NAV_HEIGHT at startup.
	DisplayHeight = EnvInt(EnvHeight, DefaultHeight)
)

// String returns the human-readable name of a FilterMode.
func (m FilterMode) String() string {
	switch m {
	case FilterDirs:
		return "directories"
	case FilterFiles:
		return "files"
	case FilterBoth:
		return "both"
	default:
		return "unknown"
	}
}
