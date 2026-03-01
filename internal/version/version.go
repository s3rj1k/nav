// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version

import (
	"fmt"
	"runtime/debug"
	"strings"

	_ "embed"
)

//go:embed version.txt
var generatedVersion string

const shortRevLen = 7 // Length of the abbreviated git revision.

// Version returns the generated calver version with optional build metadata
// (+{short_hash} or +{short_hash}.dirty) from VCS info when available.
func Version() string {
	var (
		revision string
		modified bool
	)

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
				if len(revision) > shortRevLen {
					revision = revision[:shortRevLen]
				}
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}

	version := generatedVersion
	if version == "" {
		version = "v0.0.0"
	}

	var meta []string
	if revision != "" {
		meta = append(meta, revision)
	}

	if modified {
		meta = append(meta, "dirty")
	}

	if len(meta) == 0 && generatedVersion == "" {
		meta = append(meta, "unknown")
	}

	if len(meta) > 0 {
		version += "+" + strings.Join(meta, ".")
	}

	return version
}

// VersionInfo returns a multi-line string with detailed build information.
func VersionInfo() string {
	var b strings.Builder

	fmt.Fprintf(&b, "version:  %s\n", Version())

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return b.String()
	}

	fmt.Fprintf(&b, "go:       %s\n", info.GoVersion)

	for _, s := range info.Settings {
		if s.Value == "" {
			continue
		}

		switch s.Key {
		case "vcs.revision":
			fmt.Fprintf(&b, "commit:   %s\n", s.Value)
		case "vcs.modified":
			fmt.Fprintf(&b, "modified: %s\n", s.Value)
		case "GOARCH":
			fmt.Fprintf(&b, "arch:     %s\n", s.Value)
		case "GOOS":
			fmt.Fprintf(&b, "os:       %s\n", s.Value)
		}
	}

	return b.String()
}
