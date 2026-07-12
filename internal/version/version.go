// Package version resolves the Curator build version.
package version

import "runtime/debug"

// value is set at build time through -ldflags "-X ...version.value=v0.1.0".
// When unset, the module build info (go install version) is used, and "dev"
// is the final fallback for plain source builds.
var value = ""

// String returns the version to print for --version.
func String() string {
	if value != "" {
		return value
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
