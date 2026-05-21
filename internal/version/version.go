// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2026 Serhii "s3rj1k" Ivanov.

package version

import (
	"fmt"
	"runtime/debug"
)

// VCS returns a calver-style version string derived from build VCS info.
func VCS(abbRevisionNum uint8) string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return baseVersion + "-unknown"
	}

	return VersionFromBuildInfo(bi, abbRevisionNum)
}

// VersionInfo returns a multi-line string with detailed build information.
func VersionInfo() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("version:  %s\n", baseVersion+"-unknown")
	}

	return VersionInfoFromBuildInfo(bi)
}
