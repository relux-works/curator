package gitops

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/relux-works/curator/internal/identity"
)

// AcquisitionError contains safe provenance only. Raw Git output is used for
// classification and is deliberately not retained or exposed through Unwrap.
// repository-transport §§1–3 keeps this separate from portable lock identity.
type AcquisitionError struct{ Class, Identity, Endpoint, Detail string }

func (e *AcquisitionError) Error() string {
	return fmt.Sprintf("repository_%s: identity=%s endpoint=%s %s", e.Class, e.Identity, e.Endpoint, e.Detail)
}
func acquisitionError(endpoint string, err error) error {
	if err == nil {
		return nil
	}
	text := strings.ToLower(err.Error())
	class, detail := "transport_failed", ""
	switch {
	case strings.Contains(text, "authentication failed"), strings.Contains(text, "permission denied (publickey"), strings.Contains(text, "could not read username"), strings.Contains(text, "error: 401"), strings.Contains(text, "error: 403"):
		class = "authentication_failed"
	case strings.Contains(text, "not our ref"), strings.Contains(text, "couldn't find remote ref"), strings.Contains(text, "unadvertised object"):
		class = "ref_unavailable"
	case strings.Contains(text, "transport 'ext' not allowed"):
		class, detail = "protocol_denied", "transport 'ext' not allowed"
	case strings.Contains(text, "transport 'fd' not allowed"):
		class, detail = "protocol_denied", "transport 'fd' not allowed"
	case strings.Contains(text, "already exists"):
		class, detail = "destination_exists", "destination already exists"
	}
	safe, canonical := "<invalid>", "<unknown>"
	if parsed, parseErr := url.Parse(endpoint); parseErr == nil && parsed.Host != "" {
		parsed.User = nil
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.ForceQuery = false
		safe = parsed.String()
		if id, idErr := identity.Parse(safe); idErr == nil && id != "" {
			canonical = id
		} else {
			safe = parsed.Scheme + "://" + parsed.Host
		}
	} else if id, idErr := identity.Parse(endpoint); idErr == nil && id != "" {
		canonical = id
		safe = endpoint
		if at := strings.IndexByte(safe, '@'); at >= 0 {
			safe = safe[at+1:]
		}
	} else if strings.HasPrefix(endpoint, "file:") || !strings.Contains(endpoint, ":") {
		safe = "<local>"
	}
	return &AcquisitionError{Class: class, Identity: canonical, Endpoint: safe, Detail: detail}
}

// FetchDeclaredCommit is the no-policy exact-pin acquisition boundary. The
// policy-aware replay caller retains its existing raw-output classifier in
// FetchCommitFromURLIsolated; this API exposes only sanitized typed errors.
func FetchDeclaredCommit(repo, endpoint, commit string) error {
	return acquisitionError(endpoint, FetchCommitFromURLIsolated(repo, endpoint, commit))
}
