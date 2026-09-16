package identity

import (
	"strings"
	"unicode/utf8"
)

// DraftSourceRefName validates draft-sources-v1's gitRefName scalar-length
// bound and Git ref grammar. It does not alter legacy acquisition validators.
func DraftSourceRefName(value string) bool {
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 255 || value == "@" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.HasSuffix(value, ".") || strings.Contains(value, "//") || strings.Contains(value, "..") || strings.Contains(value, "@{") {
		return false
	}
	for _, r := range value {
		if r <= 0x20 || r == 0x7f || strings.ContainsRune("~^:?*[\\", r) {
			return false
		}
	}
	for _, component := range strings.Split(value, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return false
		}
	}
	return true
}
