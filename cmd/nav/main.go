// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/s3rj1k/nav/internal/config"
	"github.com/s3rj1k/nav/internal/navigator"
	"github.com/s3rj1k/nav/internal/shell"
	"github.com/s3rj1k/nav/internal/version"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: nav [OPTIONS] [PATH]")
		fmt.Fprintln(os.Stderr, "\nInteractive file/directory navigator.")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fmt.Fprintln(os.Stderr, "  -h, --help       show this help")
		fmt.Fprintln(os.Stderr, "  --version        show version")
		fmt.Fprintln(os.Stderr, "  --init-bash      output bash shell init script")
		fmt.Fprintln(os.Stderr, "  --init-zsh       output zsh shell init script")
		fmt.Fprintln(os.Stderr, "\nEnvironment:")
		fmt.Fprintf(os.Stderr, "  %-21s display height (0 for auto, default: %d)\n", config.EnvHeight, config.DefaultHeight)
		fmt.Fprintf(os.Stderr, "  %-21s path/directories color (default: %s)\n", config.EnvColorPath, config.CSIBlue)
		fmt.Fprintf(os.Stderr, "  %-21s dim color (default: %s)\n", config.EnvColorDim, config.CSIDim)
		fmt.Fprintf(os.Stderr, "  %-21s accent color (default: %s)\n", config.EnvColorAccent, config.CSIGreen)
		fmt.Fprintf(os.Stderr, "  %-21s selected item color (default: %s)\n", config.EnvColorHighlight, config.CSIWhite)
		fmt.Fprintf(os.Stderr, "  %-21s error message color (default: %s)\n", config.EnvColorError, config.CSIRed)
		fmt.Fprintf(os.Stderr, "  %-21s cursor prefix symbol (default: %s)\n", config.EnvSelectedPrefix, config.DefaultSelectedPrefix)
		fmt.Fprintf(os.Stderr, "  %-21s help bar separator (default: %s)\n", config.EnvHelpSeparator, config.DefaultHelpSeparator)
		fmt.Fprintln(os.Stderr, "\nShell widget:")
		fmt.Fprintln(os.Stderr, "  Shift+Tab              open navigator (after shell integration)")
		fmt.Fprintln(os.Stderr, "\nPress F1 for keybindings.")
	}
	showVersion := flag.Bool("version", false, "show version")
	initBash := flag.Bool("init-bash", false, "output bash shell init script")
	initZsh := flag.Bool("init-zsh", false, "output zsh shell init script")

	flag.Parse()

	// Handle informational flags and exit early.
	switch {
	case *showVersion:
		fmt.Fprint(os.Stderr, version.VersionInfo())
		os.Exit(0)
	case *initBash:
		script, err := shell.Init("bash")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(script)
		os.Exit(0)
	case *initZsh:
		script, err := shell.Init("zsh")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(script)
		os.Exit(0)
	}

	// Use current directory as default; override with positional argument if provided.
	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Build the initial model and load the directory contents.
	m := &navigator.Model{
		Path:          absPath,
		Filter:        config.FilterBoth,
		DisplayHeight: config.DisplayHeight,
	}
	m.LoadEntries()

	// Run the TUI program. Render to stderr so stdout is reserved for
	// the selected path, enabling shell integration via $(nav).
	p := tea.NewProgram(
		m,
		tea.WithoutSignalHandler(),
		tea.WithOutput(os.Stderr),
	)

	// Block until the TUI exits or encounters a fatal error.
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Process the final model state after the TUI exits.
	if fm, ok := finalModel.(*navigator.Model); ok {
		// Exit when the user canceled (Ctrl+C or Esc).
		if fm.Canceled {
			os.Exit(config.ExitCanceled)
		}
		// Print the selected path to stdout for shell integration.
		if fm.Selected != "" {
			fmt.Println(fm.Selected)
		}
	}
}
