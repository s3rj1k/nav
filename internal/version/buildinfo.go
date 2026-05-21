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
	abbRevisionLen = 8 // Length of the abbreviated git revision.
)

// VersionFromBuildInfo derives a calver-style version string from a
// *debug.BuildInfo.
func VersionFromBuildInfo(bi *debug.BuildInfo, abbRevisionNum uint8) string {
	var (
		vcsRevision string
		vcsModified string
		vcsTime     string
	)

	for _, el := range bi.Settings {
		switch el.Key {
		case "vcs.revision":
			vcsRevision = el.Value
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

	if vcsRevision != "" && vcsTime != "" {
		if t, err := time.Parse(time.RFC3339, vcsTime); err == nil {
			return FormatCalver(t, Abbreviate(vcsRevision, abbRevisionNum))
		}
	}

	// Fallback: Main.Version (populated by `go install pkg@version`).
	switch p := ParseMainVersion(bi.Main.Version); p.Kind {
	case MainVersionTag:
		return p.Tag
	case MainVersionPseudo:
		return FormatCalver(p.Time, Abbreviate(p.Revision, abbRevisionNum))
	case MainVersionAbsent:
	}

	return baseVersion + "-unknown"
}

// FallbackPseudoFields fills missing commit/time fields from a pseudo-version
// recorded in bi.Main.Version, leaving already-set values untouched.
func FallbackPseudoFields(bi *debug.BuildInfo, revision, commitTime string) (outRevision, outCommitTime string) {
	outRevision, outCommitTime = revision, commitTime

	if outRevision != "" && outCommitTime != "" {
		return outRevision, outCommitTime
	}

	p := ParseMainVersion(bi.Main.Version)
	if p.Kind != MainVersionPseudo {
		return outRevision, outCommitTime
	}

	if outRevision == "" {
		outRevision = p.Revision
	}

	if outCommitTime == "" {
		outCommitTime = p.Time.Format(time.RFC3339)
	}

	return outRevision, outCommitTime
}

// VersionInfoFromBuildInfo renders multi-line build information from a
// *debug.BuildInfo.
func VersionInfoFromBuildInfo(bi *debug.BuildInfo) string {
	var b strings.Builder

	fmt.Fprintf(&b, "version:  %s\n", VersionFromBuildInfo(bi, abbRevisionLen))
	fmt.Fprintf(&b, "go:       %s\n", bi.GoVersion)

	var (
		revision   string
		modified   string
		commitTime string
		arch       string
		goos       string
	)

	for _, s := range bi.Settings {
		if s.Value == "" {
			continue
		}

		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		case "vcs.time":
			commitTime = s.Value
		case "GOARCH":
			arch = s.Value
		case "GOOS":
			goos = s.Value
		}
	}

	// `go install` builds: synthesize commit and time from a pseudo-version.
	revision, commitTime = FallbackPseudoFields(bi, revision, commitTime)

	if revision != "" {
		fmt.Fprintf(&b, "commit:   %s\n", revision)
	}

	if modified != "" {
		fmt.Fprintf(&b, "modified: %s\n", modified)
	}

	if commitTime != "" {
		fmt.Fprintf(&b, "time:     %s\n", commitTime)
	}

	if arch != "" {
		fmt.Fprintf(&b, "arch:     %s\n", arch)
	}

	if goos != "" {
		fmt.Fprintf(&b, "os:       %s\n", goos)
	}

	return b.String()
}
