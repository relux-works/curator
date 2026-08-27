package install

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// maxDiagnosticRunes bounds one rendered diagnostic detail. Compiler
	// output, receipts, and cache bytes are untrusted and unbounded; a report
	// line must stay a report line.
	maxDiagnosticRunes = 240
	// redactedPath replaces every absolute path in a rendered detail. Manager
	// home, cache, staging, and probe locations are private implementation
	// state and are never published to a report or a machine-readable field.
	redactedPath = "<path>"
	// truncationMarker ends a detail that was longer than the bound.
	truncationMarker = "..."
)

// pathOpeners are the characters an absolute path may directly follow. A path
// that begins after one of them is embedded — `source=/private/cache/x`,
// `file:///private/cache/x`, `error=C:\Users\name\cache` — and is redacted
// exactly like a free-standing one. Anything else before the first path
// character means the value is a relative fragment such as `assets/build-tool`
// or `linux/amd64`, which is declaration state and survives.
const pathOpeners = `=:"'` + "`" + `([{<>,;|`

// pathClosers end an absolute path run. They can never be part of the path a
// diagnostic names, so the scan stops before them and they are re-emitted.
const pathClosers = `"'` + "`" + `)]}>,;|`

// pathTrailers are trailing sentence punctuation. They are dropped from the
// captured run and re-emitted, so `read /a/b: denied` still reads as a
// sentence after `/a/b` became a placeholder.
const pathTrailers = `.,;:!?`

// RedactDiagnostic renders one untrusted diagnostic detail for an operator.
//
// Build reasons, receipt bytes, cache contents, compiler output, and the text
// of a failed build-phase error are package-, filesystem-, or
// toolchain-controlled data, so they are never printed verbatim. The rendering
// collapses the value onto one line, drops control characters so the value
// cannot drive a terminal, replaces every absolute path — free-standing,
// embedded, or in URI form — with a placeholder, and bounds the result.
// Protocol-relative paths such as a build root survive: they are declaration
// state, not private location state.
func RedactDiagnostic(detail string) string {
	cleaned := strings.ToValidUTF8(detail, "")
	// Anything that is not printable — control characters, escape sequences,
	// line breaks, and zero-width format runes — becomes plain whitespace, so
	// an untrusted detail cannot address the terminal or forge a second line.
	cleaned = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return ' '
		}
		return r
	}, cleaned)

	// Collapsing first makes the scan below operate on single spaces only, so
	// one rule decides where a path may begin regardless of the original
	// whitespace.
	return boundRunes(redactPaths(strings.Join(strings.Fields(cleaned), " ")))
}

// redactPaths replaces every absolute path anywhere in the value, not only at
// the start of a whitespace-delimited token. The scan is a single left-to-right
// pass over the already-bounded, already-printable value, so an untrusted
// detail cannot make it quadratic.
func redactPaths(value string) string {
	var out strings.Builder
	out.Grow(len(value))
	for index := 0; index < len(value); {
		length := absolutePathRun(value, index)
		if length == 0 {
			out.WriteByte(value[index])
			index++
			continue
		}
		out.WriteString(redactedPath)
		index += length
	}
	return out.String()
}

// absolutePathRun returns the length of the absolute path that starts at index,
// or zero when no path starts there. It deliberately does not consult the
// filesystem: redaction must not depend on what happens to exist on the
// reporting host.
func absolutePathRun(value string, index int) int {
	if !pathMayStartAt(value, index) || !pathStarts(value[index:]) {
		return 0
	}
	end := index
	for end < len(value) && value[end] != ' ' && strings.IndexByte(pathClosers, value[end]) < 0 {
		end++
	}
	for end > index && strings.IndexByte(pathTrailers, value[end-1]) >= 0 {
		end--
	}
	if !pathStarts(value[index:end]) {
		return 0
	}
	return end - index
}

// pathMayStartAt reports whether the character before index can precede an
// absolute path.
func pathMayStartAt(value string, index int) bool {
	if index == 0 {
		return true
	}
	previous := value[index-1]
	return previous == ' ' || strings.IndexByte(pathOpeners, previous) >= 0
}

// pathStarts reports whether a run begins an absolute Unix, URI, Windows, or
// UNC path.
func pathStarts(run string) bool {
	switch {
	case len(run) > 1 && run[0] == '/':
		return true
	case len(run) > 2 && run[0] == '\\' && run[1] == '\\':
		return true
	case len(run) > 2 && (run[2] == '\\' || run[2] == '/') && run[1] == ':' && driveLetter(run[0]):
		return true
	default:
		return false
	}
}

func driveLetter(letter byte) bool {
	return letter >= 'A' && letter <= 'Z' || letter >= 'a' && letter <= 'z'
}

func boundRunes(value string) string {
	if utf8.RuneCountInString(value) <= maxDiagnosticRunes {
		return value
	}
	keep := maxDiagnosticRunes - utf8.RuneCountInString(truncationMarker)
	runes := []rune(value)
	return strings.TrimRight(string(runes[:keep]), " ") + truncationMarker
}
