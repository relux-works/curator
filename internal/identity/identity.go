// Package identity computes the canonical source identity of git artifacts
// and matches identities against the machine allowlist (Spec §8.2).
//
// The canonical identity is "host/path" with the transport removed, the host
// lowercased, a trailing ".git" stripped, and the path kept case sensitive.
// SSH and HTTPS URLs of one repository yield one identity. Local filesystem
// sources carry no identity: the allowlist gates network fetches only.
package identity

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var urlSchemes = map[string]bool{"ssh": true, "git": true, "http": true, "https": true}

// scp-style remote: [user@]host:path (no scheme). The host part must look
// like a hostname, which keeps Windows drive paths (C:\x) and plain local
// paths out.
var scpRE = regexp.MustCompile(`^(?:[^@/\s]+@)?([A-Za-z0-9][A-Za-z0-9.-]*):([^\\]+)$`)
var hostRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*$`)

// Canonical returns the canonical "host/path" identity, or "" for local
// sources and unrecognized forms.
func Canonical(rawURL string) string {
	identity, _ := Parse(rawURL)
	return identity
}

// Parse returns the canonical identity, an empty identity for a local source,
// or an error for a malformed network source. Invalid network forms must not
// silently bypass the network allowlist as if they were local.
func Parse(rawURL string) (string, error) {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		return "", nil
	}
	if strings.Contains(value, "%") {
		return "", fmt.Errorf("network source must not contain percent escapes")
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" &&
		!urlSchemes[strings.ToLower(parsed.Scheme)] && strings.Contains(value, "://") {
		if strings.EqualFold(parsed.Scheme, "file") {
			return "", nil
		}
		return "", fmt.Errorf("unsupported network source scheme %q", parsed.Scheme)
	}
	var host, repoPath string
	if parsed, err := url.Parse(value); err == nil && urlSchemes[strings.ToLower(parsed.Scheme)] && parsed.Host != "" {
		if parsed.Port() != "" {
			return "", fmt.Errorf("network source must not contain an explicit port")
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", fmt.Errorf("network source must not contain a query or fragment")
		}
		if parsed.User != nil {
			if _, hasPassword := parsed.User.Password(); hasPassword {
				return "", fmt.Errorf("network source must not contain a password")
			}
		}
		host = strings.ToLower(parsed.Hostname())
		repoPath = parsed.Path
	} else if strings.HasPrefix(strings.ToLower(value), "file:") ||
		strings.HasPrefix(value, "/") || strings.HasPrefix(value, "./") ||
		strings.HasPrefix(value, "../") || strings.HasPrefix(value, "~") {
		return "", nil
	} else {
		match := scpRE.FindStringSubmatch(value)
		if match == nil {
			if strings.Contains(value, "://") || strings.Contains(value, "@") || strings.Contains(value, ":") {
				return "", fmt.Errorf("malformed or unsupported network source")
			}
			return "", nil
		}
		host = strings.ToLower(match[1])
		repoPath = match[2]
		if len(host) == 1 {
			// A single-letter host is a Windows drive, not a hostname.
			return "", nil
		}
	}
	if !hostRE.MatchString(host) {
		return "", fmt.Errorf("network source host is not portable ASCII")
	}
	repoPath = strings.Trim(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = strings.TrimRight(repoPath, "/")
	if host == "" || repoPath == "" || !portableRepositoryPath(repoPath) {
		return "", fmt.Errorf("network source repository path is not portable")
	}
	return host + "/" + repoPath, nil
}

func portableRepositoryPath(value string) bool {
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.ContainsAny(segment, `\:`) ||
			strings.HasSuffix(segment, ".") || strings.HasSuffix(segment, " ") {
			return false
		}
		base := strings.ToUpper(strings.SplitN(segment, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" ||
			(len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return false
		}
	}
	return true
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
