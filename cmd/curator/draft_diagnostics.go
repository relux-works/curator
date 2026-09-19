package main

// Draft source diagnostics (skillfile-sources §5, repository-transport §§2,
// 6): sanitized remediation for the stable source/transport error classes at
// the existing CLI entry points.
//
// This is a presentation layer over the landed lanes. The underlying error
// strings are the sibling lanes' contract and stay byte-identical;
// remediation is appended at print time, so output without a stable class —
// every frozen v1 message — is unchanged. Remediation text carries no
// secrets and no endpoint provenance: only the closed class vocabulary,
// the portable identity or selector the error already named, and the
// operator-owned files and commands that fix it. The per-attempt fetch
// rendering mirrors the accepted revision-2 exhaustion shape
// (buildrepo.AcquireNetworkResolved): endpoint order, transport, opaque
// provider identifier, closed class and reason — never raw tool output or
// full URLs.

import (
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
)

// draftRemediation maps one stable error class to its sanitized operator
// remediation. already names guidance the underlying message may already
// carry; when present the entry is left alone instead of repeated.
type draftRemediation struct {
	remediation string
	already     []string
}

// draftEndpointRemediation is the shared operator remediation for an
// exhausted endpoint plan, worded identically to the strict-lane
// exhaustion diagnostic so both lanes guide the operator the same way.
const draftEndpointRemediation = `verify the network path and operator authentication for the listed endpoints, then retry with machine source-policy.json`

// draftRemediations covers the nine skillfile-sources §5 classes, the four
// repository-transport §§2/6 classes, and the two source-audit-v1 outcome
// classes. Every value is static prose: nothing from the failing input is
// interpolated.
var draftRemediations = []struct {
	class string
	entry draftRemediation
}{
	{"source_alias_unknown", draftRemediation{
		remediation: `declare the alias under "sources" in Skillfile.json, or fix the "from" spelling, then run: curator project resolve`,
	}},
	{"source_selection_invalid", draftRemediation{
		remediation: `fix the named selector (directory, include/exclude, and ref rules in docs/cli.md), then retry the explicit attempt`,
		already:     []string{"requires the draft lane"},
	}},
	{"source_member_missing", draftRemediation{
		remediation: `add the named member directory with valid SKILL.md, or drop it from "include", then run: curator project resolve`,
	}},
	{"source_member_invalid", draftRemediation{
		remediation: `fix the named package (valid SKILL.md frontmatter and manifest identity), then run: curator project resolve`,
	}},
	{"source_name_conflict", draftRemediation{
		remediation: `give each installed skill exactly one selection (rename or drop a duplicate), then run: curator project resolve`,
	}},
	{"source_output_overlap", draftRemediation{
		remediation: `move the authored package out of managed output, or admit a root package via "root_inputs" in machine source-policy.json, then run: curator project resolve`,
	}},
	{"source_snapshot_changed", draftRemediation{
		remediation: `inputs changed during capture; retry the explicit attempt without editing mid-run`,
		already:     []string{"retry the explicit attempt"},
	}},
	{"source_snapshot_unavailable", draftRemediation{
		remediation: `run the explicit attempt first: curator project resolve`,
		already:     []string{"run explicit resolve first", "explicit attempt"},
	}},
	{"source_lock_stale", draftRemediation{
		remediation: `the Skillfile changed since the lock; run: curator project refresh`,
		already:     []string{"explicit refresh required", "explicit refresh"},
	}},
	{"repository_endpoint_unavailable", draftRemediation{
		remediation: draftEndpointRemediation,
		already:     []string{"verify the network path"},
	}},
	{"repository_policy_invalid", draftRemediation{
		remediation: `fix machine source-policy.json beside the manager configuration; an invalid policy is never treated as absent`,
	}},
	{"repository_mirror_undeclared", draftRemediation{
		remediation: `attest the mirror with "mirror_of" equal to the entry key in machine source-policy.json`,
	}},
	{"repository_alias_unknown", draftRemediation{
		remediation: `declare the alias in the "aliases" table of machine source-policy.json`,
	}},
	{"source_audit_rejected", draftRemediation{
		remediation: `re-resolve under trusted machine policy; the persisted audit report must match the locked package`,
	}},
	{"source_audit_unavailable", draftRemediation{
		remediation: `run the explicit attempt under trusted machine policy so the audit report is persisted`,
	}},
}

// withDraftRemediation appends the sanitized remediation of the first stable
// class named in message. Messages without a stable class, and messages
// that already carry their guidance, return unchanged.
func withDraftRemediation(message string) string {
	for _, row := range draftRemediations {
		if !strings.Contains(message, row.class) {
			continue
		}
		for _, marker := range row.entry.already {
			if strings.Contains(message, marker) {
				return message
			}
		}
		if strings.Contains(message, row.entry.remediation) {
			return message
		}
		return message + "; " + row.entry.remediation
	}
	return message
}

// draftAttemptTransport names the display transport of one planned endpoint
// URL for sanitized diagnostics. Only the scheme shape is reported; the
// URL itself is never printed.
func draftAttemptTransport(rawURL string) string {
	trimmed := strings.TrimSpace(rawURL)
	switch {
	case strings.HasPrefix(trimmed, "https://"):
		return "https"
	case strings.HasPrefix(trimmed, "http://"):
		return "http"
	case strings.HasPrefix(trimmed, "ssh://"):
		return "ssh"
	case strings.HasPrefix(trimmed, "file://"):
		return "file"
	case strings.HasPrefix(trimmed, "git://"):
		return "git"
	}
	// An scp-like spelling (user@host:path) carries no scheme; anything
	// else is outside the endpoint grammar and reported as unknown.
	if at := strings.Index(trimmed, "@"); at > 0 {
		if colon := strings.Index(trimmed[at:], ":"); colon > 0 && !strings.Contains(trimmed, "://") {
			return "ssh"
		}
	}
	return "unknown"
}

// draftAttemptClause renders one sanitized per-attempt clause in the
// revision-2 exhaustion shape: endpoint order, transport, opaque provider
// identifier, closed class and reason. The URL and raw tool output never
// appear.
func draftAttemptClause(index int, attempt config.Attempt, class buildrepo.FailureClass) string {
	clause := "endpoint " + strconv.Itoa(index) + " (" + draftAttemptTransport(attempt.URL)
	if attempt.Authentication != "" {
		clause += `, provider "` + attempt.Authentication + `"`
	}
	clause += "): " + string(class) + ": " + class.Reason()
	return clause
}
