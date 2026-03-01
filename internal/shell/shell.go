// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package shell

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed init.bash init.zsh
var scripts embed.FS

// Init returns the rendered shell init script for the given shell name.
func Init(shell string) (string, error) {
	filename := "init." + shell

	data, err := scripts.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unsupported shell %q: %w", shell, err)
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}

	tmpl, err := template.New(filename).Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, struct{ NavPath string }{NavPath: exe}); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}
