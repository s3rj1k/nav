// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version

import (
	"fmt"
	"time"
)

const revPrefix = "r"

// Abbreviate returns rev truncated to at most n characters.
func Abbreviate(rev string, n uint8) string {
	if len(rev) <= int(n) {
		return rev
	}

	return rev[:n]
}

// FormatCalver produces a SemVer-2.0.0-compliant calver string from a commit
// time and an abbreviated revision.
func FormatCalver(t time.Time, abbRevision string) string {
	secondsSinceMidnight := t.Hour()*3600 + t.Minute()*60 + t.Second()

	return fmt.Sprintf("%d.%d.%d-%d.%s%s",
		t.Year()%100, int(t.Month()), t.Day(), //nolint:mnd // last two digits of calendar year.
		secondsSinceMidnight,
		revPrefix, abbRevision,
	)
}
