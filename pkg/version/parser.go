// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version

import (
	"regexp"
	"strings"
	"time"
)

const (
	pseudoTimeLayout = "20060102150405"
	develVersion     = "(devel)" // Main.Version sentinel for local non-module builds.
)

// Trailing "<sep>yyyymmddhhmmss-abcdef123456" of a Go pseudo-version.
// See https://go.dev/ref/mod#pseudo-versions.
var pseudoVersionRE = regexp.MustCompile(`[.-](\d{14})-([0-9a-f]{12})$`)

type MainVersionKind int

const (
	MainVersionAbsent MainVersionKind = iota // empty or "(devel)"
	MainVersionTag                           // regular module tag
	MainVersionPseudo                        // Go pseudo-version
)

type ParsedMainVersion struct {
	Time     time.Time
	Raw      string
	Tag      string
	Revision string
	Kind     MainVersionKind
}

func ParsePseudoVersion(v string) (time.Time, string, bool) {
	m := pseudoVersionRE.FindStringSubmatch(v)
	if m == nil {
		return time.Time{}, "", false
	}

	t, err := time.Parse(pseudoTimeLayout, m[1])
	if err != nil {
		return time.Time{}, "", false
	}

	return t.UTC(), m[2], true
}

func ParseMainVersion(v string) ParsedMainVersion {
	p := ParsedMainVersion{Raw: v}

	if v == "" || v == develVersion {
		return p
	}

	if t, rev, ok := ParsePseudoVersion(v); ok {
		p.Kind = MainVersionPseudo
		p.Time = t
		p.Revision = rev

		return p
	}

	p.Kind = MainVersionTag
	p.Tag = strings.TrimPrefix(v, "v")

	return p
}
