package buildrepo

// Bounded authenticated transport resolution (repository-transport-v1,
// revision 1, §2): apply a machine policy attempt plan to the strict
// external-repository lane.
//
// A TransportPlan is the validated form of one declaration's ordered
// endpoints: one or two closed-grammar URLs that canonicalize to one
// identity, an opaque operator provider reference per endpoint, and an
// effective fallback mode. The sibling policy loader produces the same
// shape as config.Resolution; this package cannot import config (config
// imports buildrepo for the lane grammar), so the wiring task converts
// field-for-field and this executor revalidates everything before any
// network I/O.
//
// Every endpoint attempt is a full strict-lane acquisition: trusted Git,
// clean configuration/environment, closed process graph, the same
// admission checks as a direct lane call, a per-attempt SSH wrapper
// policy bound from the attempt's credentials, the existing per-fetch
// HTTPS credential broker, exact-ref fetch, and raw-object proof of the
// locked content. There is
// at most one attempt per listed endpoint (maximum two total), no implicit
// retry, never a synthesized URL, and one total deadline shared by both
// attempts. The second endpoint is attempted only under
// `availability-auth` after a positively classified availability or
// authentication failure. Every other class fails closed: the lane's own
// diagnostic is returned unchanged, with no alternate network traffic and
// no cache shortcut.
//
// Errors and trace records never carry fetch stderr, secrets, or full
// endpoint URLs. Exhaustion diagnostics are built from a closed
// class/reason vocabulary; URLs reach only the machine-private trace
// callback, never portable errors.

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/gitcred"
)

// CodeRepositoryEndpointUnavailable reports exhausted bounded resolution:
// every planned endpoint failed with a fallback-eligible class and no
// attempt remains (repository-transport §2). It shares its spelling with
// the policy loader's diagnostic; both name the same published class.
const CodeRepositoryEndpointUnavailable = "repository_endpoint_unavailable"

// Effective fallback modes (repository-transport §2). Pin is already
// applied when a plan is built: a pinned plan holds one attempt with
// TransportFallbackNone, so the executor needs no pin logic.
const (
	TransportFallbackNone             = "none"
	TransportFallbackAvailabilityAuth = "availability-auth"
)

// FailureClass is one §2 failure row. Only FailureAvailability and
// FailureAuth ever permit a second endpoint attempt, and only under
// TransportFallbackAvailabilityAuth.
type FailureClass string

// Transport failure classes. The string values reuse the conformance
// vocabulary where it names the class (tls, host-key, ref-moved,
// identity, integrity, audit, unknown, http-404, policy-unreadable);
// availability and auth are the two fallback-eligible rows.
const (
	FailureAvailability FailureClass = "availability"
	FailureAuth         FailureClass = "auth"
	FailureTLS          FailureClass = "tls"
	FailureHostKey      FailureClass = "host-key"
	FailureRef          FailureClass = "ref-moved"
	FailureIdentity     FailureClass = "identity"
	FailureIntegrity    FailureClass = "integrity"
	FailureAudit        FailureClass = "audit"
	FailureUnknown      FailureClass = "unknown"
	FailureHTTP404      FailureClass = "http-404"
	FailurePolicy       FailureClass = "policy-unreadable"
)

// Reason is the closed per-class reason used in exhaustion diagnostics.
// Raw fetch output is never interpolated: the class alone determines the
// text, so a hostile or confused remote cannot inject secrets or
// misleading detail into errors. The CLI draft fetch path renders the
// same vocabulary; the strict lane keeps calling failureReason.
func (c FailureClass) Reason() string { return c.failureReason() }

// failureReason is the closed per-class reason used in exhaustion
// diagnostics. Raw fetch output is never interpolated: the class alone
// determines the text, so a hostile or confused remote cannot inject
// secrets or misleading detail into errors.
func (c FailureClass) failureReason() string {
	switch c {
	case FailureAvailability:
		return "endpoint unavailable (DNS, connection, or HTTP 502/503/504 failure)"
	case FailureAuth:
		return "endpoint authentication unavailable or rejected"
	case FailureTLS:
		return "TLS certificate validation failed"
	case FailureHostKey:
		return "SSH host-key validation failed"
	case FailureRef:
		return "missing or moved ref"
	case FailureIdentity:
		return "wrong repository identity"
	case FailureIntegrity:
		return "object or integrity mismatch"
	case FailureAudit:
		return "audit denial"
	case FailureHTTP404:
		return "ambiguous HTTP 404"
	case FailurePolicy:
		return "unreadable machine policy"
	default:
		return "unclassified failure"
	}
}

// AllowSecondAttempt is the §2 fallback gate: only `availability-auth`
// plans advance past a positively classified availability or
// authentication failure. Attempts remaining, the shared deadline, and
// plan shape are enforced by the executor; this predicate decides the
// class row only.
func AllowSecondAttempt(fallback string, class FailureClass) bool {
	if fallback != TransportFallbackAvailabilityAuth {
		return false
	}
	return class == FailureAvailability || class == FailureAuth
}

// SemanticFailureClass maps a draft-sources-v1 semantic-case first_failure
// token to the failure class the gate decides on. `dns` stands for the
// whole availability row (refused/timeout/502/503/504 classify the same);
// `auth-rejected` stands for the auth row.
func SemanticFailureClass(token string) (FailureClass, bool) {
	switch token {
	case "dns":
		return FailureAvailability, true
	case "auth-rejected":
		return FailureAuth, true
	case "tls":
		return FailureTLS, true
	case "host-key":
		return FailureHostKey, true
	case "integrity":
		return FailureIntegrity, true
	case "identity":
		return FailureIdentity, true
	case "ref-moved":
		return FailureRef, true
	case "audit":
		return FailureAudit, true
	case "unknown":
		return FailureUnknown, true
	case "policy-unreadable":
		return FailurePolicy, true
	case "http-404":
		return FailureHTTP404, true
	default:
		return FailureUnknown, false
	}
}

// Closed transport diagnostic grammar (repository-transport-v1 revision 1,
// §2): every entry below is one full stderr line git or ssh emits,
// anchored at both ends and built only from fixed text plus bounded tokens
// (a dotted RFC-1123 hostname, an optional numeric port, a three-digit HTTP
// status, a single-quoted endpoint span, an enumerated reason phrase).
// Open wildcards are absent by construction: a line that matches no entry
// is unparsed evidence and fails the whole fetch output closed, however
// many recognized lines surround it. Unknown facilities, truncated output,
// empty output, and any unconsumed tail all refuse the same way.
//
// Endpoint text can never read as a signal: patterns that carry the
// endpoint consume it inside the quoted span, and reason phrases are drawn
// from an enumerated list of known git, curl, and ssh sentences (resolver
// failures, refused or timed-out connections, unreachable routes, HTTP
// 401/403 and 502/503/504 wrappers, RPC failures, headless credential
// prompts, forge rejections, the ssh daemon method list, TLS and host-key
// sentences). Anything outside the list — a future-git sentence, a local
// filesystem failure quoting a signal as a filename, a secret echo, an
// informational warning, a progress line — matches nothing.
//
// transportHostLabel is one RFC-1123 hostname label: alphanumerics with
// interior hyphens, bounded to 63 octets. transportHost is a dotted
// sequence of labels.
var transportHostLabel = `[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?`

var transportHost = transportHostLabel + `(\.` + transportHostLabel + `)*`

// transportLineSignals is the closed line table. Each entry pairs one
// full-line pattern with the §2 class it reports. A framing entry carries
// no failure evidence — git's fixed ssh/rpc trailers name no cause — so
// the classifier skips it instead of reporting it. Alone, framing
// classifies nothing.
var transportLineSignals = []struct {
	pattern *regexp.Regexp
	class   FailureClass
	framing bool
}{
	// Fixed ssh/rpc failure framing: the "Could not read from remote
	// repository" trailer ssh-transport failures always carry (with its
	// access-rights hint), and the "remote end hung up" trailer
	// accompanying RPC failures. Skipped, never evidence.
	{regexp.MustCompile(`(?i)^fatal: the remote end hung up unexpectedly\.?$`), FailureUnknown, true},
	{regexp.MustCompile(`(?i)^fatal: Could not read from remote repository\.?$`), FailureUnknown, true},
	{regexp.MustCompile(`(?i)^Please make sure you have the correct access rights$`), FailureUnknown, true},
	{regexp.MustCompile(`(?i)^and the repository exists\.?$`), FailureUnknown, true},
	// TLS certificate failures (fail closed).
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': SSL certificate problem: (unable to get local issuer certificate|self signed certificate)$`), FailureTLS, false},
	{regexp.MustCompile(`(?i)^fatal: SSL certificate problem: (unable to get local issuer certificate|self signed certificate)$`), FailureTLS, false},
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': schannel: failed to receive handshake, SSL/TLS connection failed$`), FailureTLS, false},
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': OpenSSL SSL_connect: Connection timed out$`), FailureTLS, false},
	// SSH host-key failures (fail closed).
	{regexp.MustCompile(`(?i)^Host key verification failed\.$`), FailureHostKey, false},
	{regexp.MustCompile(`(?i)^WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!$`), FailureHostKey, false},

	// Missing or moved refs (fail closed).
	{regexp.MustCompile(`(?i)^fatal: couldn't find remote ref [A-Za-z0-9_./-]+\.?$`), FailureRef, false},
	// Ambiguous HTTP 404 shapes (fail closed: a 404 can hide an auth
	// rejection, so it never authorizes an alternate).
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': The requested URL returned error: 404$`), FailureHTTP404, false},
	{regexp.MustCompile(`(?i)^remote: Repository not found\.?$`), FailureHTTP404, false},
	// Wrong identity: redirects and remapping (fail closed).
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': Redirection from https?://[^ ]+ to https?://[^ ]+ is forbidden$`), FailureIdentity, false},
	// Object and integrity mismatches (fail closed).
	{regexp.MustCompile(`(?i)^error: object hash mismatch$`), FailureIntegrity, false},
	{regexp.MustCompile(`(?i)^fatal: fsck error: object [0-9a-f]+ is corrupted\.?$`), FailureIntegrity, false},
	{regexp.MustCompile(`(?i)^error: checksum mismatch for object [0-9a-f]+$`), FailureIntegrity, false},
	// Audit denials, revocations, canaries, and assurance failures (fail
	// closed).
	{regexp.MustCompile(`(?i)^remote: audit denied: operation timed out$`), FailureAudit, false},
	{regexp.MustCompile(`(?i)^remote: key is revoked: access denied$`), FailureAudit, false},
	{regexp.MustCompile(`(?i)^remote: canary check failed for this repository$`), FailureAudit, false},
	{regexp.MustCompile(`(?i)^remote: assurance failure: invalid credentials presented$`), FailureAudit, false},
	// Availability sentences (fallback-eligible under availability-auth).
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': Could not resolve host: ` + transportHost + `$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^ssh: Could not resolve hostname ` + transportHost + `: (Temporary failure in name resolution|Name or service not known|nodename nor servname provided, or not known)$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^ssh: connect to host ([A-Za-z0-9._-]+@)?` + transportHost + ` port [0-9]{1,5}: (Connection refused|Connection timed out|No route to host|Network is unreachable)$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': Operation timed out after [0-9]{1,20} milliseconds with [0-9]{1,20} bytes received$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': Failed to connect to ` + transportHost + `: (Network is unreachable|No route to host|Connection refused|Connection timed out)$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': The requested URL returned error: 50[234]$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^error: RPC failed; HTTP 50[234] curl [0-9]{1,3} The requested URL returned error: 50[234]$`), FailureAvailability, false},
	{regexp.MustCompile(`(?i)^fatal: connection timed out\.?$`), FailureAvailability, false},
	// Authentication sentences (fallback-eligible under availability-auth).
	{regexp.MustCompile(`(?i)^fatal: unable to access '[^']+': The requested URL returned error: 40[13]$`), FailureAuth, false},
	{regexp.MustCompile(`(?i)^fatal: Authentication failed for '[^']+'\.?$`), FailureAuth, false},
	{regexp.MustCompile(`(?i)^([A-Za-z0-9._-]+@` + transportHost + `: )?Permission denied \((publickey|password|keyboard-interactive|hostbased)(,(publickey|password|keyboard-interactive|hostbased))*\)\.?$`), FailureAuth, false},
	{regexp.MustCompile(`(?i)^remote: Invalid username or password\.?$`), FailureAuth, false},
	{regexp.MustCompile(`(?i)^remote: Permission to [A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)* denied to [A-Za-z0-9_.-]+\.?$`), FailureAuth, false},
	{regexp.MustCompile(`(?i)^fatal: could not read (Username|Password) for '[^']+': (terminal prompts disabled|No such device or address)$`), FailureAuth, false},
}

// classifyFetchLine maps one stderr line to its §2 class through the
// closed table. The second result is false for framing lines, which the
// caller skips: they carry no failure evidence either way. A line that
// matches no table entry reports (FailureUnknown, true): unparsed
// evidence the caller must fail closed on, never skip.
func classifyFetchLine(raw string) (FailureClass, bool) {
	for _, signal := range transportLineSignals {
		if signal.pattern.MatchString(raw) {
			if signal.framing {
				return FailureUnknown, false
			}
			return signal.class, true
		}
	}
	return FailureUnknown, true
}

// failClosedPrecedence orders fail-closed classes when one fetch emits
// several recognized diagnostics: the first class in this list present
// in the output is reported. The order mirrors the previous
// implementation so mixed-evidence records keep their established
// class.
var failClosedPrecedence = []FailureClass{
	FailureTLS,
	FailureHostKey,
	FailureIdentity,
	FailureRef,
	FailureHTTP404,
	FailureIntegrity,
	FailureAudit,
}

// ClassifyFetchOutput maps bounded fetch stderr to a §2 failure class
// through the closed line table. The output is fallback-eligible only
// when every non-empty line matches exactly one table entry and at least
// one matched line is an availability or authentication diagnostic.
// Git's fixed ssh/rpc framing trailers match framing entries and are
// skipped as non-evidence; every other line must match a diagnostic
// entry. Empty output and any unmatched line — an unknown facility, a
// future-git sentence, a local-filesystem failure quoting a signal as a
// filename, an HTTP status with an unconsumed tail, a bare number, a
// progress line — are FailureUnknown and fail closed. Fallback-eligible
// availability/auth is reported only when at least one line positively
// identifies that row and no line reports a fail-closed row or unparsed
// evidence, so mixed output (for example integrity details beside a
// connection note, an audit denial beside a timeout, or a recognized
// diagnostic beside an unknown line) never permits an alternate attempt.
func ClassifyFetchOutput(stderr string) FailureClass {
	var sawAvailability, sawAuth bool
	seenFailClosed := map[FailureClass]bool{}
	anyDiagnostic := false
	// Split only on real newline boundaries. A helper that does not
	// newline-terminate its final line glues content into one line
	// that matches no closed-table entry, so it fails closed as
	// unparsed evidence: the classifier never invents boundaries.
	for _, line := range strings.Split(stderr, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		class, diagnostic := classifyFetchLine(trimmed)
		if !diagnostic {
			continue
		}
		anyDiagnostic = true
		switch class {
		case FailureUnknown:
			return FailureUnknown
		case FailureAvailability:
			sawAvailability = true
		case FailureAuth:
			sawAuth = true
		default:
			seenFailClosed[class] = true
		}
	}
	if !anyDiagnostic {
		return FailureUnknown
	}
	for _, class := range failClosedPrecedence {
		if seenFailClosed[class] {
			return class
		}
	}
	if len(seenFailClosed) != 0 {
		return FailureUnknown
	}
	if sawAvailability {
		return FailureAvailability
	}
	if sawAuth {
		return FailureAuth
	}
	return FailureUnknown
}

// ClassifyAdmissionCode maps a strict-lane admission code produced after
// the fetch phase to a §2 failure class. Post-fetch lane failures are
// deterministic (the lane proved what went wrong), so the mapping is by
// code, never by stderr. CodeSourceUnavailable here means a non-fetch
// lane failure (private-state initialization); fetch failures carry
// stderr and classify through ClassifyFetchOutput instead.
func ClassifyAdmissionCode(code string) FailureClass {
	switch code {
	case CodeRefMoved:
		return FailureRef
	case CodeIncompleteSource, CodeObjectSemanticsInvalid, CodeLFSUnsupported:
		return FailureIntegrity
	case CodeIdentityInvalid, CodeLocalGitfileUnsupported, CodeLocalBareUnsupported,
		CodeLocalLinkedUnsupported, CodeLocalLayoutUnsafe, CodeLocalFormatUnsupported,
		CodeLocalObjectFormatUnsupported:
		return FailureIdentity
	case CodeSSHCredentialMissing:
		return FailureAuth
	case CodeAuditBlocked:
		return FailureAudit
	case CodeDescriptorInvalid, CodeReceiptInvalid, CodeArtifactInvalid,
		CodeProtectedBoundaryUntrusted, CodeUnverifiedOffline, CodeSignerPolicyUnsupported,
		CodePackageSigningForbidden:
		return FailureIntegrity
	default:
		return FailureUnknown
	}
}

// TransportAttempt is one planned endpoint attempt: a closed-grammar URL
// and the opaque operator provider reference that authenticates it. An
// empty Authentication names no policy provider: the attempt uses the
// existing lane credentials from the base request, which is the only
// shape a URL declaration without a policy entry produces.
//
// Revision-2 endpoints (repository-transport §§5-7) carry machine-policy
// endpoint properties beside the URL: MirrorOf attests a resolved
// connection host that differs from the plan identity host, Alias names
// the operator alias the connection substitutes, and ResolvedHost,
// ResolvedPort and HasExplicitPort record the resulting connection
// address (alias target when an alias is named, else the URL host and
// URL port, else the transport default). The executor revalidates the
// §6 predicate over these carried values before any network I/O; the
// alias-table structural checks (chaining, authentication match,
// embedded hosts) stay loader-enforced, since the plan carries no table.
type TransportAttempt struct {
	URL             string
	Authentication  string
	MirrorOf        string
	Alias           string
	ResolvedHost    string
	ResolvedPort    int
	HasExplicitPort bool
}

// carriesRevision2 reports whether the attempt carries any revision-2
// endpoint property. A bare revision-1 attempt carries none: its
// resolved address is the URL itself.
func (a TransportAttempt) carriesRevision2() bool {
	return a.MirrorOf != "" || a.Alias != "" || a.ResolvedHost != "" || a.ResolvedPort != 0 || a.HasExplicitPort
}

// TransportPlan is the ordered attempt plan for one declaration: one or
// two attempts that canonicalize to Identity, and the effective fallback
// mode after pin is applied.
type TransportPlan struct {
	Identity string
	Attempts []TransportAttempt
	Fallback string
}

// isLegacyShape reports the no-policy-entry shape: the declared URL once,
// existing lane credentials, no fallback. A legacy failure keeps the
// lane's own diagnostic exactly as a direct strict-lane call would.
func (p TransportPlan) isLegacyShape() bool {
	return len(p.Attempts) == 1 && p.Attempts[0].Authentication == "" && p.Fallback == TransportFallbackNone
}

// TransportPlanRefusalPrefix is the static class-and-diagnostic prefix of
// the strict-lane §7 transport-plan refusal below: the only
// build_repository_identity_invalid diagnostic that names a transport
// plan endpoint. The install acquisition layer preserves this refusal
// through its collapse, and the CLI keys its remediation row on this
// same prefix, so both stay scoped to the refusal they describe and
// every other identity diagnostic keeps the legacy behavior.
const TransportPlanRefusalPrefix = CodeIdentityInvalid + ": transport plan endpoint"

// ValidateTransportPlan checks plan shape before any network I/O: one or
// two attempts, distinct closed-grammar URLs that canonicalize to one
// identity, a closed fallback mode, and opaque provider references that
// are never commands or paths. Violations fail repository_policy_invalid:
// a plan always derives from machine policy, so a malformed plan is a
// malformed policy.
//
// Revision-2 attempts additionally carry the §5 endpoint properties
// (mirror_of, alias, resolved connection address), which are revalidated
// against the §6 predicate here. An attempt with an explicit port or an
// alias field then fails build_repository_identity_invalid: this
// executor runs the strict external-build lane, whose URL grammar admits
// neither, and the lane must never strip a port or ignore an alias to
// force admission (§7). A declared mirror without ports or aliases is an
// ordinary lane URL and remains admissible.
func ValidateTransportPlan(plan TransportPlan) error {
	_, err := parseTransportPlan(plan)
	return err
}

// parseTransportPlan validates the plan and returns the parsed lane
// sources in attempt order. Revision-1 attempts take the original
// checks byte-identically; revision-2 attempts additionally pass the §6
// predicate and the §7 lane-grammar refusal.
func parseTransportPlan(plan TransportPlan) ([]Source, error) {
	if len(plan.Attempts) < 1 || len(plan.Attempts) > 2 {
		return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan requires one or two attempts")
	}
	if plan.Fallback != TransportFallbackNone && plan.Fallback != TransportFallbackAvailabilityAuth {
		return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan requires fallback %q or %q", TransportFallbackNone, TransportFallbackAvailabilityAuth)
	}
	if plan.Identity == "" {
		return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan requires a canonical identity")
	}
	sources := make([]Source, 0, len(plan.Attempts))
	seen := map[string]bool{}
	for index, attempt := range plan.Attempts {
		urlHost, urlPort, hasURLPort, strippedIdentity, err := parseTransportEndpointURL(attempt.URL)
		if err != nil {
			if attempt.carriesRevision2() {
				return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d is outside the revision-2 endpoint grammar", index+1)
			}
			return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d is outside the closed repository grammar", index+1)
		}
		if !attempt.carriesRevision2() {
			// A port-bearing URL without revision-2 provenance is a
			// mistranslation, never a silently stripped endpoint.
			if hasURLPort {
				return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d carries an explicit port without revision-2 provenance", index+1)
			}
			if strippedIdentity != plan.Identity {
				return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d canonicalizes to %q, want %q", index+1, strippedIdentity, plan.Identity)
			}
		} else if err := checkTransportRevision2(plan.Identity, index, attempt, urlHost, urlPort, hasURLPort, strippedIdentity); err != nil {
			return nil, err
		}
		// The strict lane admits no explicit port and no alias
		// rewriting: refuse before any network I/O rather than strip
		// the port or fetch the unsubstituted URL (§7).
		if attempt.HasExplicitPort || attempt.Alias != "" {
			return nil, admissionError(CodeIdentityInvalid, "transport plan endpoint %d carries an explicit port or host alias outside the strict external-build lane grammar", index+1)
		}
		if seen[attempt.URL] {
			return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoints must be distinct URLs")
		}
		seen[attempt.URL] = true
		if attempt.Authentication != "" && !gitcred.ValidProvider(attempt.Authentication) {
			return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d authentication is not an opaque operator provider identifier", index+1)
		}
		// Port and alias attempts never reach here: they are refused
		// above, so the listed URL always parses under the lane's
		// closed grammar. A declared mirror is an ordinary lane URL.
		parsed, err := ParseSource(attempt.URL)
		if err != nil {
			return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d is outside the closed repository grammar", index+1)
		}
		sources = append(sources, parsed)
	}
	return sources, nil
}

// CodeRepositoryMirrorUndeclared names the published §6 class for a
// resolved connection host that differs from the plan identity host
// without a mirror_of attestation equal to the identity. Zero attempts,
// no fallback.
const CodeRepositoryMirrorUndeclared = "repository_mirror_undeclared"

// checkTransportRevision2 revalidates one revision-2 attempt's §5
// endpoint properties against the single §6 resolved-host predicate.
// The alias-table structural checks (embedded alias hosts, chained
// aliases, authentication match) stay loader-enforced: the plan carries
// the resolved address, not the table, so this check enforces
// consistency of the carried values with the URL and the identity.
func checkTransportRevision2(planIdentity string, index int, attempt TransportAttempt, urlHost string, urlPort int, hasURLPort bool, strippedIdentity string) error {
	endpoint := index + 1
	keyHost, keyPath, _ := strings.Cut(planIdentity, "/")
	_, epPath, _ := strings.Cut(strippedIdentity, "/")
	// The port-stripped path must equal the identity path; only the
	// host may differ, and only when attested.
	if epPath != keyPath {
		return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d path %q is not the plan identity path %q; only the host may differ, and only when attested", endpoint, epPath, keyPath)
	}
	resolvedHost := attempt.ResolvedHost
	if attempt.Alias == "" {
		if resolvedHost == "" {
			resolvedHost = urlHost
		}
		if resolvedHost != urlHost {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d resolved host %q does not match the endpoint URL host %q", endpoint, resolvedHost, urlHost)
		}
		if attempt.HasExplicitPort != hasURLPort || attempt.ResolvedPort != urlPort {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d resolved port does not match the endpoint URL port", endpoint)
		}
	} else {
		// A mirror URL combined with an alias is refused: the URL
		// host would be neither identity nor connection address.
		if urlHost != keyHost {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d combines a mirror URL with an alias", endpoint)
		}
		if resolvedHost == "" || !hostRE.MatchString(resolvedHost) {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d names an alias without a resolved connection host", endpoint)
		}
		// A URL port and an alias port must not both be present: the
		// resolved port differs from the URL port only through the
		// alias port, so a carried port that equals neither reading
		// (or the absence pattern it breaks) is a mistranslation.
		if hasURLPort && attempt.HasExplicitPort && attempt.ResolvedPort != 0 && attempt.ResolvedPort != urlPort {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d carries both a URL port and an alias port", endpoint)
		}
		if attempt.HasExplicitPort != (attempt.ResolvedPort != 0) || (hasURLPort && (!attempt.HasExplicitPort || attempt.ResolvedPort != urlPort)) {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d resolved port does not match the endpoint URL port", endpoint)
		}
	}
	if resolvedHost != keyHost {
		if attempt.MirrorOf == "" {
			return admissionError(CodeRepositoryMirrorUndeclared, "transport plan endpoint %d resolved host %q differs from the plan identity host %q without a mirror_of attestation", endpoint, resolvedHost, keyHost)
		}
		if attempt.MirrorOf != planIdentity {
			return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d mirror_of must equal the plan identity %q exactly", endpoint, planIdentity)
		}
	} else if attempt.MirrorOf != "" {
		return admissionError(CodeRepositoryPolicyInvalid, "transport plan endpoint %d mirror_of is forbidden when the resolved host equals the plan identity host", endpoint)
	}
	return nil
}

// parseTransportEndpointURL validates a §5 endpoint URL: the lane's
// closed grammar plus an optional explicit port on URI forms only
// (https://host[:port]/path, ssh://[user@]host[:port]/path), decimal
// 1-65535 with no leading zeros. Scp-like spellings carry no port: the
// segment after ":" is always the path. It returns the lowercased URL
// host, the explicit port (0 when absent), and the port-stripped lane
// identity the key-path check applies to. Port-free URLs behave exactly
// as ParseSource: stripped input equals raw input.
func parseTransportEndpointURL(raw string) (urlHost string, urlPort int, hasPort bool, identity string, err error) {
	stripped, port, hasExplicitPort, ok := splitTransportEndpointPort(raw)
	if !ok {
		return "", 0, false, "", fmt.Errorf("outside the revision-2 endpoint grammar")
	}
	parsed, parseErr := ParseSource(stripped)
	if parseErr != nil {
		return "", 0, false, "", parseErr
	}
	host, _, _ := strings.Cut(parsed.Identity, "/")
	return host, port, hasExplicitPort, parsed.Identity, nil
}

// splitTransportEndpointPort separates an explicit URI-form port from
// the endpoint URL. Every other spelling passes through unchanged.
func splitTransportEndpointPort(raw string) (stripped string, port int, hasPort bool, ok bool) {
	scheme := ""
	rest := ""
	if remainder, found := strings.CutPrefix(raw, "https://"); found {
		scheme, rest = "https", remainder
	} else if remainder, found := strings.CutPrefix(raw, "ssh://"); found {
		scheme, rest = "ssh", remainder
	} else {
		return raw, 0, false, true
	}
	authority := rest
	path := ""
	if i := strings.Index(rest, "/"); i >= 0 {
		authority, path = rest[:i], rest[i:]
	}
	hostport := authority
	userinfo := ""
	if i := strings.LastIndex(authority, "@"); i >= 0 {
		userinfo, hostport = authority[:i], authority[i+1:]
	}
	if scheme == "https" && userinfo != "" {
		return "", 0, false, false
	}
	host := hostport
	if i := strings.LastIndex(hostport, ":"); i >= 0 {
		value, valid := parseTransportPort(hostport[i+1:])
		if !valid {
			return "", 0, false, false
		}
		host, port, hasPort = hostport[:i], value, true
	}
	if host == "" {
		return "", 0, false, false
	}
	rebuilt := scheme + "://"
	if userinfo != "" {
		rebuilt += userinfo + "@"
	}
	return rebuilt + host + path, port, hasPort, true
}

// parseTransportPort validates an explicit endpoint port: decimal
// 1-65535 with no leading zeros.
func parseTransportPort(value string) (int, bool) {
	if value == "" || len(value) > 5 {
		return 0, false
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, false
	}
	number := 0
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		number = number*10 + int(c-'0')
	}
	if number < 1 || number > 65535 {
		return 0, false
	}
	return number, true
}

// CodeRepositoryPolicyInvalid names the published malformed-policy class
// for plan-shape violations. It shares its spelling with the policy
// loader's diagnostic; both name the same §2 class.
const CodeRepositoryPolicyInvalid = "repository_policy_invalid"

// CodeTransportResolutionUnsupportedPlatform names the typed refusal of
// the bounded resolved lane on a platform without fetch process-graph
// lifetime control. The lane's total deadline must bound the whole
// child process tree; only process-group cancellation provides that,
// and Windows has no equivalent in this lane (no Job Object control is
// implemented), so the lane refuses before any process creation rather
// than run unbounded there. This draft-scope diagnostic is outside the
// §2 failure table: it describes lane support, not endpoint evidence.
const CodeTransportResolutionUnsupportedPlatform = "transport_resolution_unsupported_platform"

// transportResolutionPlatformError reports whether the bounded resolved
// lane may run on the named GOOS. It is parameterized for testing; the
// executor passes runtime.GOOS.
func transportResolutionPlatformError(goos string) error {
	if goos == "windows" {
		return admissionError(CodeTransportResolutionUnsupportedPlatform,
			"bounded transport resolution is not supported on windows: fetch process-graph lifetime control is unavailable")
	}
	return nil
}

// AuthProvider resolves one opaque policy provider reference to fetch
// material for a single attempt. It is operator-owned: identifiers
// resolve only through operator configuration, never through package
// input, and never as commands or executable paths.
//
// ResolveHTTPS returns the credential the per-fetch broker answers with,
// ErrProviderAnonymous when the provider is explicitly anonymous (offer
// nothing), or ErrProviderUnavailable when the provider is unconfigured
// or holds no usable HTTPS material. ResolveSSH returns the operator SSH
// selection for the attempt, or ErrProviderUnavailable when the provider
// holds none; SSH has no anonymous transport in this lane, so an
// explicitly anonymous provider is unavailable for SSH. A nil error with
// unselected credentials is treated as unavailable: a named provider
// must yield material, name anonymity explicitly, or stay silent.
type AuthProvider interface {
	ResolveHTTPS(ctx context.Context, provider, host string) (HTTPSCredentials, error)
	ResolveSSH(provider string) (OperatorSSHCredentials, error)
}

// Provider resolution outcomes.
var (
	ErrProviderAnonymous   = errors.New("authentication provider is explicitly anonymous")
	ErrProviderUnavailable = errors.New("authentication provider holds no usable material")
)

// ProviderHTTPS is the operator's HTTPS configuration for one provider:
// explicit anonymity, or a broker-held secret with an optional username.
type ProviderHTTPS struct {
	Username  string
	Anonymous bool
}

// ProviderSSH is the operator's SSH configuration for one provider:
// absolute identity, agent-socket, and known-hosts paths, admitted by
// the existing SSH credential validator. Paths are validated, never
// executed.
type ProviderSSH struct {
	Identity    string
	AgentSocket string
	KnownHosts  string
}

// Provider is one operator-configured authentication provider. A nil
// transport entry means the provider offers no material for that
// transport.
type Provider struct {
	HTTPS *ProviderHTTPS
	SSH   *ProviderSSH
}

// ProviderSet is the operator-owned provider table, keyed by opaque
// provider identifier. The manager populates it from operator
// configuration; package input can neither add providers nor select
// them, since plans carry only the identifier the policy named.
type ProviderSet map[string]Provider

func (s ProviderSet) get(provider string) (Provider, bool) {
	if !gitcred.ValidProvider(provider) {
		return Provider{}, false
	}
	entry, ok := s[provider]
	return entry, ok
}

// ProviderSecretReader is the trusted-broker surface provider HTTPS
// resolution reads through. gitcred.Access is its production
// implementation: provider secrets stay in the existing broker.
type ProviderSecretReader interface {
	ReadProvider(ctx context.Context, provider, host string) (gitcred.HostCredential, bool)
}

// CredentialProviders resolves policy provider references against an
// operator provider table, reading HTTPS secrets through the existing
// trusted broker. A nil Reader or a nil Set resolves nothing: every
// named provider is unavailable rather than ambient.
type CredentialProviders struct {
	Set    ProviderSet
	Reader ProviderSecretReader
}

// ResolveHTTPS implements AuthProvider.
func (c CredentialProviders) ResolveHTTPS(ctx context.Context, provider, host string) (HTTPSCredentials, error) {
	entry, ok := c.Set.get(provider)
	if !ok || entry.HTTPS == nil {
		return HTTPSCredentials{}, ErrProviderUnavailable
	}
	if entry.HTTPS.Anonymous {
		return HTTPSCredentials{}, ErrProviderAnonymous
	}
	if c.Reader == nil {
		return HTTPSCredentials{}, ErrProviderUnavailable
	}
	material, ok := c.Reader.ReadProvider(ctx, provider, host)
	if !ok {
		return HTTPSCredentials{}, ErrProviderUnavailable
	}
	username := entry.HTTPS.Username
	if username == "" {
		username = material.Username
	}
	credentials := NewHTTPSCredentials(host, username, material.Secret)
	if !credentials.Selected() {
		return HTTPSCredentials{}, ErrProviderUnavailable
	}
	return credentials, nil
}

// ResolveSSH implements AuthProvider.
func (c CredentialProviders) ResolveSSH(provider string) (OperatorSSHCredentials, error) {
	entry, ok := c.Set.get(provider)
	if !ok || entry.SSH == nil {
		return OperatorSSHCredentials{}, ErrProviderUnavailable
	}
	admitted, err := ValidateOperatorSSHCredentials(OperatorSSHCredentials{
		Identity: entry.SSH.Identity, AgentSocket: entry.SSH.AgentSocket, KnownHosts: entry.SSH.KnownHosts,
	})
	if err != nil || !admitted.Selected() {
		return OperatorSSHCredentials{}, ErrProviderUnavailable
	}
	return admitted, nil
}

// AttemptRecord is the sanitized machine-private record of one planned
// endpoint: which URL was attempted with which provider, what class it
// failed with (empty on success), and whether any network traffic went
// out. Records never carry stderr or secrets.
//
// Revision-2 provenance (§7) rides the same record: the canonical plan
// identity (never an alias or mirror host), the listed URL with its
// port, the resolved connection host and port, and the alias and
// mirror_of properties when used. The record stays machine-private:
// identity, URLs, and hosts never enter portable artifacts (locks,
// receipts, markers, manifests), and errors carry only the closed
// class vocabulary.
type AttemptRecord struct {
	Index int
	URL   string
	// Transport is the lane transport of the attempt ("https" or "ssh").
	Transport string
	// Provider is the opaque operator provider identifier, "" when the
	// attempt used existing lane credentials.
	Provider string
	// Class is the §2 failure class, "" when the attempt succeeded.
	Class FailureClass
	// NetworkAttempted reports whether a fetch went out. A provider
	// with no usable material records its auth failure with no traffic.
	NetworkAttempted bool
	Succeeded        bool
	// Identity is the canonical repository identity the attempt proves,
	// always the plan identity: ports, mirrors, and aliases never enter
	// it.
	Identity string
	// ResolvedHost and ResolvedPort are the connection address the
	// attempt dials: the alias target when an alias is named, else the
	// endpoint URL host and port. ResolvedPort is 0 for the transport
	// default.
	ResolvedHost string
	ResolvedPort int
	// Alias names the operator alias substituted for this attempt, ""
	// when the URL was dialed directly.
	Alias string
	// MirrorOf carries the mirror attestation when the resolved host
	// differs from the identity host, "" otherwise.
	MirrorOf string
}

// TransportTrace receives one AttemptRecord per planned endpoint, in
// attempt order, including the successful attempt. The caller stores
// records in machine-private operation diagnostics, separately from
// portable identity. A nil trace records nothing.
type TransportTrace func(AttemptRecord)

// AcquireNetworkResolved acquires one locked source through a bounded
// transport plan: at most one strict-lane fetch per listed endpoint
// within the lane's single total deadline. The lane is not supported on
// Windows, where fetch process-graph lifetime control is unavailable:
// there it refuses with transport_resolution_unsupported_platform
// before any process creation, with no attempt recorded.
//
// base carries the lock, exact tag, substitution refs, limits, and — for
// attempts without a policy provider — the existing lane credentials.
// The plan's identity must equal the base source identity: the alternate
// endpoint proves the same repository, never another one. providers
// resolves named provider references; it may be nil when no attempt
// names a provider. trace receives the sanitized per-endpoint records.
//
// Every fetching attempt runs the same strict-lane admission as
// AcquireNetwork under the shared deadline, so a resolved acquisition
// refuses exactly what the lane refuses. Every SSH attempt additionally
// binds its endpoint and its selected credentials into a real wrapper
// policy; the fetch then runs behind a per-attempt wrapper pinned to
// that policy instead of the tool's static wrapper.
//
// Success returns the proved snapshot: full locked object, exact tag
// when declared, objects, and snapshot, verified independently of which
// endpoint served them. A policy-plan exhaustion returns
// repository_endpoint_unavailable with sanitized classifications and
// remediation. Any fail-closed class returns the lane's own diagnostic
// unchanged, with no further network traffic. A legacy-shape plan (one
// attempt, no provider, no fallback) behaves exactly like AcquireNetwork.
func AcquireNetworkResolved(ctx context.Context, base NetworkRequest, plan TransportPlan, providers AuthProvider, trace TransportTrace) (*Snapshot, error) {
	sources, err := parseTransportPlan(plan)
	if err != nil {
		return nil, err
	}
	if plan.Identity != base.Source.Identity {
		return nil, admissionError(CodeRepositoryPolicyInvalid, "transport plan identity %q does not match acquisition identity %q", plan.Identity, base.Source.Identity)
	}
	// Platform support is checked after plan validation (a malformed
	// policy is reported identically everywhere) and before any provider
	// binding, admission, or fetch: on a refused platform the lane
	// creates no process and records no attempt.
	if err := transportResolutionPlatformError(runtime.GOOS); err != nil {
		return nil, err
	}
	limits := normalizedLimits(base.Limits)
	ctx, cancel := context.WithTimeout(ctx, limits.Timeout)
	defer cancel()

	records := make([]AttemptRecord, 0, len(plan.Attempts))
	emit := func(record AttemptRecord) {
		records = append(records, record)
		if trace != nil {
			trace(record)
		}
	}
	// refuse records one attempt that ended before any fetch traffic —
	// a lane admission refusal or an unbindable SSH policy — and gates
	// the plan on its class.
	refuse := func(record *AttemptRecord, err error) (bool, error) {
		record.Class, record.NetworkAttempted = classifyAttemptError(ctx, err), false
		emit(*record)
		return gateTransportAttempt(ctx, plan, records, record.Class, laneDiagnostic(err))
	}
	for index, attempt := range plan.Attempts {
		// Provenance is bound before any traffic: the canonical plan
		// identity with the attempt's own connection address (§7). A
		// bare revision-1 attempt resolves to its URL host directly.
		resolvedHost, resolvedPort := attempt.ResolvedHost, attempt.ResolvedPort
		if resolvedHost == "" {
			resolvedHost, _, _ = strings.Cut(sources[index].Identity, "/")
		}
		record := AttemptRecord{Index: index + 1, URL: attempt.URL, Transport: sources[index].Transport, Provider: attempt.Authentication,
			Identity: plan.Identity, ResolvedHost: resolvedHost, ResolvedPort: resolvedPort, Alias: attempt.Alias, MirrorOf: attempt.MirrorOf}
		request := base
		request.Source = sources[index]
		request.sshPolicy = nil
		if attempt.Authentication != "" {
			bound, skip := bindProviderAttempt(ctx, request.Tool, sources[index], attempt.Authentication, providers)
			if skip != nil {
				record.Class, record.NetworkAttempted = skip.class, false
				emit(record)
				if done, terminal := gateTransportAttempt(ctx, plan, records, record.Class, nil); done {
					return nil, terminal
				}
				continue
			}
			request.Tool = bound
		}
		if err := admitNetworkRequest(ctx, request); err != nil {
			if done, terminal := refuse(&record, err); done {
				return nil, terminal
			}
			continue
		}
		if sources[index].Transport == "ssh" {
			policy, err := bindSSHWrapperPolicy(request.Tool, sources[index])
			if err != nil {
				if done, terminal := refuse(&record, err); done {
					return nil, terminal
				}
				continue
			}
			request.sshPolicy = &policy
		}
		snapshot, attemptErr := acquireNetworkFormat(ctx, request, limits)
		if attemptErr == nil {
			record.NetworkAttempted, record.Succeeded = true, true
			emit(record)
			return snapshot, nil
		}
		record.NetworkAttempted = true
		record.Class = classifyAttemptError(ctx, attemptErr)
		emit(record)
		if done, terminal := gateTransportAttempt(ctx, plan, records, record.Class, laneDiagnostic(attemptErr)); done {
			return nil, terminal
		}
	}
	// The loop always returns: a one-attempt plan terminates on its
	// attempt, and a two-attempt plan terminates on its second.
	return nil, admissionError(CodeSourceUnavailable, "exact source fetch failed")
}

// providerSkip records an endpoint whose provider yielded no usable
// material: an auth-method failure with no network traffic.
type providerSkip struct {
	class FailureClass
}

// bindProviderAttempt binds one named provider's material for one
// endpoint attempt. It returns the bound tool, or a skip when the
// provider offers nothing usable for the attempt's transport. Binding
// replaces lane credentials rather than merging: a named provider is
// the attempt's only authentication, and an explicitly anonymous HTTPS
// provider offers nothing even when the base request carries a secret.
func bindProviderAttempt(ctx context.Context, tool GitTool, source Source, provider string, providers AuthProvider) (GitTool, *providerSkip) {
	skip := &providerSkip{class: FailureAuth}
	if providers == nil {
		return GitTool{}, skip
	}
	switch source.Transport {
	case "https":
		credentials, err := providers.ResolveHTTPS(ctx, provider, sourceHost(source))
		if err != nil {
			if errors.Is(err, ErrProviderAnonymous) {
				tool.HTTPSCredentials = HTTPSCredentials{}
				return tool, nil
			}
			return GitTool{}, skip
		}
		if !credentials.Selected() || credentials.Host != sourceHost(source) {
			return GitTool{}, skip
		}
		tool.HTTPSCredentials = credentials
		return tool, nil
	case "ssh":
		credentials, err := providers.ResolveSSH(provider)
		if err != nil || !credentials.Selected() {
			return GitTool{}, skip
		}
		tool.SSHCredentials = credentials
		return tool, nil
	default:
		return GitTool{}, &providerSkip{class: FailureIdentity}
	}
}

// sourceHost is the canonical host of a parsed lane source.
func sourceHost(source Source) string {
	host, _, _ := strings.Cut(source.Identity, "/")
	return host
}

// classifyAttemptError maps one failed strict-lane attempt to its §2
// class. A fetch failure classifies from its stderr unless the retained
// output is incomplete or the shared deadline (or cancellation) fired: a
// truncated diagnostic prefix and a locally expired clock are not
// positive endpoint classifications, so both fail closed as unknown.
// Post-fetch lane failures classify deterministically by admission code.
func classifyAttemptError(ctx context.Context, attemptErr error) FailureClass {
	var fetch *fetchError
	if errors.As(attemptErr, &fetch) {
		if fetch.truncated || ctx.Err() != nil {
			return FailureUnknown
		}
		return ClassifyFetchOutput(fetch.stderr)
	}
	if code := ErrorCode(attemptErr); code != "" {
		return ClassifyAdmissionCode(code)
	}
	return FailureUnknown
}

// gateTransportAttempt decides whether a failed attempt ends the plan.
// It reports done with the terminal error when no further attempt is
// allowed: exhaustion for a policy plan whose last failure is
// fallback-eligible, otherwise the lane's own diagnostic unchanged.
func gateTransportAttempt(ctx context.Context, plan TransportPlan, records []AttemptRecord, class FailureClass, laneErr error) (bool, error) {
	if AllowSecondAttempt(plan.Fallback, class) && len(records) < len(plan.Attempts) && ctx.Err() == nil {
		return false, nil
	}
	if !plan.isLegacyShape() && (class == FailureAvailability || class == FailureAuth) {
		return true, exhaustionError(plan, records, ctx.Err() != nil)
	}
	if laneErr != nil {
		return true, laneErr
	}
	// Unreachable through the executor: only provider skips record a
	// failure without a lane error, and skips always carry the auth
	// class, which exhausts above. Fail closed with the lane's fetch
	// diagnostic rather than inventing one.
	return true, admissionError(CodeSourceUnavailable, "exact source fetch failed")
}

// exhaustionError builds the sanitized repository_endpoint_unavailable
// diagnostic: one closed-vocabulary clause per attempted endpoint, a
// deadline note when the shared clock expired before the plan ran out,
// and operator remediation. Attempt order, transports, provider
// identifiers, and the stable identity are included; stderr, secrets,
// and full URLs are not.
func exhaustionError(plan TransportPlan, records []AttemptRecord, deadlineExpired bool) error {
	clauses := make([]string, 0, len(plan.Attempts))
	for _, record := range records {
		clause := fmt.Sprintf("endpoint %d (%s", record.Index, record.Transport)
		if record.Provider != "" {
			clause += fmt.Sprintf(", provider %q", record.Provider)
		}
		clause += "): "
		if !record.NetworkAttempted {
			clause += "no fetch attempted: " + FailureAuth.failureReason()
		} else {
			clause += string(record.Class) + ": " + record.Class.failureReason()
		}
		clauses = append(clauses, clause)
	}
	for index := len(records); index < len(plan.Attempts); index++ {
		note := "not attempted"
		if deadlineExpired {
			note += " (total deadline expired)"
		}
		clauses = append(clauses, fmt.Sprintf("endpoint %d: %s", index+1, note))
	}
	detail := strings.Join(clauses, "; ")
	return admissionError(CodeRepositoryEndpointUnavailable,
		"%s: %s; verify the network path and operator authentication for the listed endpoints, then retry with machine source-policy.json", plan.Identity, detail)
}
