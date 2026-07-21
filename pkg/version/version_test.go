// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version //nolint:testpackage // whitebox tests cover unexported parseMainVersion and constants.

import (
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"golang.org/x/mod/semver"
)

const (
	testRev = "abcd1234"

	testCommit        = "abcdef1234567890"
	testCommitTime    = "2026-05-21T14:00:00Z"
	testGoVersion     = "go1.26.3"
	testTag           = "v26.5.21-50400.rabcdef12"
	testTagWant       = "26.5.21-50400.rabcdef12"
	testPseudoVersion = "v0.0.0-20260521140000-abcdef123456"
	testPseudoRev     = "abcdef123456"
	valArm64          = "arm64"
	valDarwin         = "darwin"
	valModifiedFalse  = "false"

	keyVCSRevision = "vcs.revision"
	keyVCSTime     = "vcs.time"
	keyVCSModified = "vcs.modified"
	keyGOARCH      = "GOARCH"
	keyGOOS        = "GOOS"

	wantVersionLine = "version:  26.5.21-50400.rabcdef12"
	wantGoLine      = "go:       go1.26.3"
	wantArchLine    = "arch:     arm64"
	wantOSLine      = "os:       darwin"
)

// Every FormatCalver output must parse as SemVer 2.0.0 once "v"-prefixed.
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
			v := FormatCalver(tc.t, tc.rev)
			tag := "v" + v

			if !semver.IsValid(tag) {
				t.Fatalf("FormatCalver(%v, %q) = %q; %q is not valid SemVer", tc.t, tc.rev, v, tag)
			}
		})
	}
}

// A zero n means no truncation, so the full revision is returned.
func TestAbbreviate_ZeroMeansFull(t *testing.T) {
	if got := Abbreviate(testCommit, 0); got != testCommit {
		t.Fatalf("got %q, want %q", got, testCommit)
	}
}

// The dirty/unknown fallback constants must also be valid SemVer.
func TestFallbackIsSemVer(t *testing.T) {
	for _, v := range []string{baseVersion + "-dirty", baseVersion + "-unknown"} {
		tag := "v" + v
		if !semver.IsValid(tag) {
			t.Fatalf("%q is not valid SemVer", tag)
		}
	}
}

func TestParseMainVersion(t *testing.T) {
	pseudoTime := time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC)

	cases := []struct {
		wantT   time.Time
		name    string
		input   string
		tag     string
		wantRev string
		kind    MainVersionKind
	}{
		{name: "empty", input: "", kind: MainVersionAbsent},
		{name: "devel", input: develVersion, kind: MainVersionAbsent},

		{
			name:    "pseudo-no-earlier-version",
			input:   testPseudoVersion,
			kind:    MainVersionPseudo,
			wantT:   pseudoTime,
			wantRev: testPseudoRev,
		},
		{
			name:    "pseudo-prerelease-base",
			input:   "v1.2.3-pre.0.20260521140000-abcdef123456",
			kind:    MainVersionPseudo,
			wantT:   pseudoTime,
			wantRev: testPseudoRev,
		},
		{
			name:    "pseudo-after-release",
			input:   "v1.2.4-0.20260521140000-abcdef123456",
			kind:    MainVersionPseudo,
			wantT:   pseudoTime,
			wantRev: testPseudoRev,
		},

		{name: "semver-tag", input: "v1.2.3", kind: MainVersionTag, tag: "1.2.3"},
		{name: "calver-tag", input: testTag, kind: MainVersionTag, tag: testTagWant},
		{name: "prerelease-tag", input: "v1.2.3-rc1", kind: MainVersionTag, tag: "1.2.3-rc1"},

		{
			name:  "uppercase-hex-not-pseudo",
			input: "v0.0.0-20260521140000-ABCDEF123456",
			kind:  MainVersionTag,
			tag:   "0.0.0-20260521140000-ABCDEF123456",
		},
		{
			name:  "short-rev-not-pseudo",
			input: "v0.0.0-20260521140000-abcd",
			kind:  MainVersionTag,
			tag:   "0.0.0-20260521140000-abcd",
		},
		{
			name:  "bad-timestamp-not-pseudo",
			input: "v0.0.0-20261321140000-abcdef123456",
			kind:  MainVersionTag,
			tag:   "0.0.0-20261321140000-abcdef123456",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := ParseMainVersion(tc.input)

			if p.Kind != tc.kind {
				t.Fatalf("Kind = %v, want %v", p.Kind, tc.kind)
			}

			if p.Raw != tc.input {
				t.Errorf("Raw = %q, want %q", p.Raw, tc.input)
			}

			switch tc.kind {
			case MainVersionTag:
				if p.Tag != tc.tag {
					t.Errorf("Tag = %q, want %q", p.Tag, tc.tag)
				}
			case MainVersionPseudo:
				if p.Revision != tc.wantRev {
					t.Errorf("Revision = %q, want %q", p.Revision, tc.wantRev)
				}
				if !p.Time.Equal(tc.wantT) {
					t.Errorf("Time = %v, want %v", p.Time, tc.wantT)
				}
			case MainVersionAbsent:
			}
		})
	}
}

func TestFromBuildInfo_LocalBuild(t *testing.T) {
	bi := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: keyVCSRevision, Value: testCommit},
			{Key: keyVCSTime, Value: testCommitTime},
			{Key: keyVCSModified, Value: valModifiedFalse},
		},
	}

	want := FormatCalver(time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "abcdef12")
	if got := FromBuildInfo(bi, AbbRevisionLen); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// A vcs.time value with an offset other than UTC must normalize to the
// same version string as its UTC equivalent.
func TestFromBuildInfo_NonUTCTime(t *testing.T) {
	bi := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: keyVCSRevision, Value: testCommit},
			{Key: keyVCSTime, Value: "2026-05-21T16:00:00+02:00"},
			{Key: keyVCSModified, Value: valModifiedFalse},
		},
	}

	want := FormatCalver(time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "abcdef12")
	if got := FromBuildInfo(bi, AbbRevisionLen); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFromBuildInfo_DirtyTree(t *testing.T) {
	bi := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: keyVCSRevision, Value: testCommit},
			{Key: keyVCSTime, Value: testCommitTime},
			{Key: keyVCSModified, Value: "true"},
		},
	}

	if got := FromBuildInfo(bi, AbbRevisionLen); got != baseVersion+"-dirty" {
		t.Fatalf("got %q, want %q", got, baseVersion+"-dirty")
	}
}

// `go install pkg@latest` on an untagged repo: no vcs.*, pseudo Main.Version.
func TestFromBuildInfo_PseudoVersion(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: testPseudoVersion},
	}

	want := FormatCalver(time.Date(2026, 5, 21, 14, 0, 0, 0, time.UTC), "abcdef12")
	if got := FromBuildInfo(bi, AbbRevisionLen); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// `go install pkg@vX.Y.Z`: Main.Version holds a real tag.
func TestFromBuildInfo_Tag(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: testTag},
	}

	want := testTagWant
	if got := FromBuildInfo(bi, AbbRevisionLen); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFromBuildInfo_Devel(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: develVersion},
	}

	if got := FromBuildInfo(bi, AbbRevisionLen); got != baseVersion+"-unknown" {
		t.Fatalf("got %q, want %q", got, baseVersion+"-unknown")
	}
}

func TestVersionInfoFromBuildInfo_LocalBuild(t *testing.T) {
	bi := &debug.BuildInfo{
		GoVersion: testGoVersion,
		Settings: []debug.BuildSetting{
			{Key: keyVCSRevision, Value: testCommit},
			{Key: keyVCSTime, Value: testCommitTime},
			{Key: keyVCSModified, Value: valModifiedFalse},
			{Key: keyGOARCH, Value: valArm64},
			{Key: keyGOOS, Value: valDarwin},
		},
	}

	out := VersionInfoFromBuildInfo(bi)

	for _, want := range []string{
		wantVersionLine,
		wantGoLine,
		"commit:   abcdef1234567890",
		"modified: false",
		"time:     2026-05-21T14:00:00Z",
		wantArchLine,
		wantOSLine,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// Commit and time are synthesized from the pseudo-version when vcs.* absent.
func TestVersionInfoFromBuildInfo_PseudoVersion(t *testing.T) {
	bi := &debug.BuildInfo{
		GoVersion: testGoVersion,
		Main:      debug.Module{Version: testPseudoVersion},
		Settings: []debug.BuildSetting{
			{Key: keyGOARCH, Value: valArm64},
			{Key: keyGOOS, Value: valDarwin},
		},
	}

	out := VersionInfoFromBuildInfo(bi)

	for _, want := range []string{
		wantVersionLine,
		wantGoLine,
		"commit:   abcdef123456",
		"time:     2026-05-21T14:00:00Z",
		wantArchLine,
		wantOSLine,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}

	if strings.Contains(out, "modified:") {
		t.Errorf("did not expect modified: line for go-install build:\n%s", out)
	}
}

// A plain tag carries no commit/time, so only the version line is rendered.
func TestVersionInfoFromBuildInfo_Tag(t *testing.T) {
	bi := &debug.BuildInfo{
		GoVersion: testGoVersion,
		Main:      debug.Module{Version: "v26.5.21-50400.rabcdef12"},
		Settings: []debug.BuildSetting{
			{Key: keyGOARCH, Value: valArm64},
			{Key: keyGOOS, Value: valDarwin},
		},
	}

	out := VersionInfoFromBuildInfo(bi)

	for _, want := range []string{
		wantVersionLine,
		wantGoLine,
		wantArchLine,
		wantOSLine,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}

	for _, unwanted := range []string{"commit:", "time:", "modified:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("did not expect %q line for tag-only build:\n%s", unwanted, out)
		}
	}
}
