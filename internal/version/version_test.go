// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version //nolint:testpackage // formatCalver is unexported; whitebox tests keep it that way.

import (
	"testing"
	"time"

	"golang.org/x/mod/semver"
)

const testRev = "abcd1234"

// TestCalverIsSemVer asserts every value produced by formatCalver parses as
// SemVer 2.0.0. The release workflow prepends "v" before tagging, so we
// validate "v"+version here to mirror the actual tag form.
func TestCalverIsSemVer(t *testing.T) {
	cases := []struct {
		name string
		t    time.Time
		rev  string
	}{
		{"new-year-midnight", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), testRev},
		{"end-of-year-last-second", time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC), testRev},
		{"year-rolls-to-zero", time.Date(2100, 6, 15, 12, 30, 45, 0, time.UTC), testRev},
		{"single-digit-month-day", time.Date(2026, 3, 5, 10, 20, 30, 0, time.UTC), testRev},
		{"double-digit-month-day", time.Date(2026, 11, 25, 1, 2, 3, 0, time.UTC), testRev},
		{"short-rev", time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "a"},
		{"long-rev", time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "abcdef0123456789"},
		{"numeric-rev", time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "01234567"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := formatCalver(tc.t, tc.rev)
			tag := "v" + v

			if !semver.IsValid(tag) {
				t.Fatalf("formatCalver(%v, %q) = %q; %q is not valid SemVer", tc.t, tc.rev, v, tag)
			}
		})
	}
}

// TestFallbackIsSemVer covers the dirty/unknown branches of VCS(), which
// return constant strings without invoking formatCalver.
func TestFallbackIsSemVer(t *testing.T) {
	for _, v := range []string{baseVersion + "-dirty", baseVersion + "-unknown"} {
		tag := "v" + v
		if !semver.IsValid(tag) {
			t.Fatalf("%q is not valid SemVer", tag)
		}
	}
}
