// Package identity computes the canonical source identity of git artifacts
// and matches identities against the machine allowlist (Spec §8.2).
//
// The canonical identity is "host/path" with the transport removed, the host
// lowercased, a trailing ".git" stripped, and the path kept case sensitive.
// SSH and HTTPS URLs of one repository yield one identity. Local filesystem
// sources carry no identity: the allowlist gates network fetches only.
package identity

import (
	"net/url"
	"regexp"
	"strings"
)

var urlSchemes = map[string]bool{"ssh": true, "git": true, "http": true, "https": true}

// scp-style remote: [user@]host:path (no scheme). The host part must look
// like a hostname, which keeps Windows drive paths (C:\x) and plain local
// paths out.
var scpRE = regexp.MustCompile(`^(?:[^@/\s]+@)?([A-Za-z0-9][A-Za-z0-9.-]*):([^\\]+)$`)

// Canonical returns the canonical "host/path" identity, or "" for local
// sources and unrecognized forms.
func Canonical(rawURL string) string {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return ""
	}
	var host, repoPath string
	if parsed, err := url.Parse(value); err == nil && urlSchemes[strings.ToLower(parsed.Scheme)] && parsed.Host != "" {
		host = strings.ToLower(parsed.Hostname())
		repoPath = parsed.Path
	} else if strings.HasPrefix(strings.ToLower(value), "file:") ||
		strings.HasPrefix(value, "/") || strings.HasPrefix(value, "./") ||
		strings.HasPrefix(value, "../") || strings.HasPrefix(value, "~") {
		return ""
	} else {
		match := scpRE.FindStringSubmatch(value)
		if match == nil {
			return ""
		}
		host = strings.ToLower(match[1])
		repoPath = match[2]
		if len(host) == 1 {
			// A single-letter host is a Windows drive, not a hostname.
			return ""
		}
	}
	repoPath = strings.Trim(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = strings.TrimRight(repoPath, "/")
	if host == "" || repoPath == "" {
		return ""
	}
	return host + "/" + repoPath
}

// MatchesPrefix reports a segment-aware prefix match: "h/skills" matches
// "h/skills/x" and never "h/skills-evil".
func MatchesPrefix(identity, prefix string) bool {
	trimmed := strings.TrimRight(strings.TrimSpace(prefix), "/")
	if trimmed == "" {
		return false
	}
	return identity == trimmed || strings.HasPrefix(identity, trimmed+"/")
}

// Allowed checks an identity against the machine allowlist. An empty
// allowlist allows every source; an empty identity is a local source, which
// involves no network operation and passes.
func Allowed(identity string, allowedPrefixes []string) bool {
	if len(allowedPrefixes) == 0 {
		return true
	}
	if identity == "" {
		return true
	}
	for _, prefix := range allowedPrefixes {
		if MatchesPrefix(identity, prefix) {
			return true
		}
	}
	return false
}
