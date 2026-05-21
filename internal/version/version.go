// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

const (
	baseVersion    = "0.0.0"
	revPrefix      = "r"
	abbRevisionLen = 8 // Length of the abbreviated git revision.
)

// VCS returns a calver-style version string derived from build VCS info.
func VCS(abbRevisionNum uint8) string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return baseVersion + "-unknown"
	}

	var (
		vcsRevision []rune
		vcsModified string
		vcsTime     string
	)

	for _, el := range buildInfo.Settings {
		switch el.Key {
		case "vcs.revision":
			vcsRevision = []rune(el.Value)
		case "vcs.modified":
			vcsModified = el.Value
		case "vcs.time":
			vcsTime = el.Value
		default:
			continue
		}
	}

	if strings.EqualFold(vcsModified, "true") {
		return baseVersion + "-dirty"
	}

	if len(vcsRevision) == 0 || vcsTime == "" {
		return baseVersion + "-unknown"
	}

	t, err := time.Parse(time.RFC3339, vcsTime)
	if err != nil {
		return baseVersion + "-unknown"
	}

	var abbRevision string
	if len(vcsRevision) <= int(abbRevisionNum) {
		abbRevision = string(vcsRevision)
	} else {
		abbRevision = string(vcsRevision[:abbRevisionNum])
	}

	return formatCalver(t, abbRevision)
}

// formatCalver produces the SemVer-2.0.0-compliant calver string used by VCS.
func formatCalver(t time.Time, abbRevision string) string {
	secondsSinceMidnight := t.Hour()*3600 + t.Minute()*60 + t.Second()

	return fmt.Sprintf("%d.%d.%d-%d.%s%s",
		t.Year()%100, int(t.Month()), t.Day(), //nolint:mnd // last two digits of calendar year.
		secondsSinceMidnight,
		revPrefix, abbRevision,
	)
}

// VersionInfo returns a multi-line string with detailed build information.
func VersionInfo() string {
	var b strings.Builder

	fmt.Fprintf(&b, "version:  %s\n", VCS(abbRevisionLen))

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
		case "vcs.time":
			fmt.Fprintf(&b, "time:     %s\n", s.Value)
		case "GOARCH":
			fmt.Fprintf(&b, "arch:     %s\n", s.Value)
		case "GOOS":
			fmt.Fprintf(&b, "os:       %s\n", s.Value)
		}
	}

	return b.String()
}
