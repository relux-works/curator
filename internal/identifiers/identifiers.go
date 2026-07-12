// Package identifiers validates the identifier alphabet of Spec §5.2 and the
// source path rule of Spec §6.1.
//
// Skill names, command names, and MCP server names become filesystem path
// components (shim filenames, runtime directories). Restricting them to a
// safe alphabet keeps a third-party manifest from writing outside its
// designated directories.
package identifiers

import (
	"regexp"
	"strings"
)

var identifierRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Rule is the human-readable identifier rule, phrased once for reuse in
// error messages.
const Rule = "must start with a letter or digit and contain only letters, digits, dots, underscores, or hyphens"

// SourceRule is the human-readable rule for source paths under skills_root.
const SourceRule = "must be a relative path whose segments start with a letter or digit and contain only letters, digits, dots, underscores, or hyphens"

// Valid reports whether value matches the identifier alphabet of Spec §5.2.
func Valid(value string) bool {
	return identifierRE.MatchString(value)
}

// ValidSourcePath reports whether value is a POSIX-style relative path whose
// every segment is a valid identifier (Spec §6.1). This rules out "..",
// absolute paths, backslashes, and option-like segments.
func ValidSourcePath(value string) bool {
	if value == "" {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if !Valid(segment) {
			return false
		}
	}
	return true
}
