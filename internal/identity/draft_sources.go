package identity

import (
	"strings"
	"unicode/utf8"
)

// DraftCanonicalKey reports whether key is an already-exact canonical
// repository identity (repository-transport-v1 §1: lowercase host,
// case-sensitive path, one terminal ".git" removed, no transport, no
// user, no port). Policy entry keys and logical `repository` declarations
// must satisfy this without normalization: a spelling that canonicalizes
// to anything but itself is refused rather than repaired, so two
// spellings of one repository never key two entries.
//
// The check reuses Parse, the core §6.1 canonicalizer, over the HTTPS
// spelling of the key. Endpoint URLs themselves are validated against
// the stricter closed lane grammar by the policy loader; equality of
// that lane identity with the key ties the two together.
func DraftCanonicalKey(key string) bool {
	if key == "" {
		return false
	}
	identity, err := Parse("https://" + key)
	return err == nil && identity == key
}

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
