package buildrepo

import (
	"strings"
	"testing"
)

// Revision-2 plan validation (repository-transport §§6-7): the §6
// resolved-host predicate over carried endpoint properties, the two new
// failure classes, the attempt bound, and the strict-lane port/alias
// refusal. No network I/O occurs here.

// A declared mirror without ports or aliases is an ordinary lane URL:
// validation succeeds and the lane source names the mirror endpoint.
func TestValidateTransportPlanAdmitsDeclaredMirror(t *testing.T) {
	t.Parallel()
	plan := TransportPlan{
		Identity: "example.org/kit",
		Attempts: []TransportAttempt{{
			URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
			MirrorOf: "example.org/kit", ResolvedHost: "mirror.example.net",
		}},
		Fallback: TransportFallbackNone,
	}
	sources, err := parseTransportPlan(plan)
	if err != nil {
		t.Fatalf("parseTransportPlan(mirror) = %v", err)
	}
	if len(sources) != 1 || sources[0].Git != "https://mirror.example.net/kit.git" || sources[0].Transport != "https" {
		t.Fatalf("sources = %+v, want the listed mirror endpoint", sources)
	}
}

// Every §6/§7 refusal row fails at plan validation with its published
// class, before any network I/O.
func TestValidateTransportPlanRevision2Refusals(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		attempts []TransportAttempt
		fallback string
		code     string
	}{
		{
			"undeclared mirror",
			[]TransportAttempt{{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
				ResolvedHost: "mirror.example.net"}},
			TransportFallbackNone, CodeRepositoryMirrorUndeclared,
		},
		{
			// A mirror URL without any revision-2 provenance is not
			// an attested mirror at all: it fails the revision-1
			// identity check, still closed, still before any I/O.
			"mirror URL without revision-2 provenance",
			[]TransportAttempt{{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"mirror_of mismatch",
			[]TransportAttempt{{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
				MirrorOf: "other.example/kit", ResolvedHost: "mirror.example.net"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"spurious mirror_of on same host",
			[]TransportAttempt{{URL: "https://example.org/kit.git", Authentication: "team-https",
				MirrorOf: "example.org/kit", ResolvedHost: "example.org"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"mirror path must equal the identity path",
			[]TransportAttempt{{URL: "https://mirror.example.net/other.git", Authentication: "mirror-https",
				MirrorOf: "example.org/kit", ResolvedHost: "mirror.example.net"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"https port refused in the strict lane",
			[]TransportAttempt{{URL: "https://example.org:8443/kit.git", Authentication: "team-https",
				ResolvedHost: "example.org", ResolvedPort: 8443, HasExplicitPort: true}},
			TransportFallbackNone, CodeIdentityInvalid,
		},
		{
			"ssh port refused in the strict lane",
			[]TransportAttempt{{URL: "ssh://git@example.org:2222/kit.git", Authentication: "team-ssh",
				ResolvedHost: "example.org", ResolvedPort: 2222, HasExplicitPort: true}},
			TransportFallbackNone, CodeIdentityInvalid,
		},
		{
			"alias refused in the strict lane even on the same host",
			[]TransportAttempt{{URL: "https://example.org/kit.git", Authentication: "team-https",
				Alias: "local-vip", ResolvedHost: "example.org"}},
			TransportFallbackNone, CodeIdentityInvalid,
		},
		{
			"alias refused in the strict lane on a mirror target",
			[]TransportAttempt{{URL: "https://example.org/kit.git", Authentication: "team-https",
				Alias: "corp-mirror", MirrorOf: "example.org/kit", ResolvedHost: "mirror.corp.example", ResolvedPort: 8443, HasExplicitPort: true}},
			TransportFallbackNone, CodeIdentityInvalid,
		},
		{
			"mirror URL combined with an alias",
			[]TransportAttempt{{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
				Alias: "corp-mirror", MirrorOf: "example.org/kit", ResolvedHost: "mirror.corp.example"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"alias without a resolved host",
			[]TransportAttempt{{URL: "https://example.org/kit.git", Authentication: "team-https",
				Alias: "corp-mirror"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"double port",
			[]TransportAttempt{{URL: "https://example.org:8443/kit.git", Authentication: "team-https",
				Alias: "corp-mirror", ResolvedHost: "example.org", ResolvedPort: 9443, HasExplicitPort: true}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"resolved port mistranslation",
			[]TransportAttempt{{URL: "https://example.org:8443/kit.git", Authentication: "team-https",
				ResolvedHost: "example.org", ResolvedPort: 9443, HasExplicitPort: true}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"port without revision-2 provenance is never stripped",
			[]TransportAttempt{{URL: "https://example.org:8443/kit.git", Authentication: "team-https"}},
			TransportFallbackNone, CodeRepositoryPolicyInvalid,
		},
		{
			"three endpoints exceed the unchanged attempt bound",
			[]TransportAttempt{
				{URL: "https://example.org/kit.git"},
				{URL: "https://example.org/kit"},
				{URL: "ssh://git@example.org/kit.git"},
			},
			TransportFallbackAvailabilityAuth, CodeRepositoryPolicyInvalid,
		},
		{
			"duplicate mirror endpoints",
			[]TransportAttempt{
				{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
					MirrorOf: "example.org/kit", ResolvedHost: "mirror.example.net"},
				{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https",
					MirrorOf: "example.org/kit", ResolvedHost: "mirror.example.net"},
			},
			TransportFallbackAvailabilityAuth, CodeRepositoryPolicyInvalid,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			plan := TransportPlan{Identity: "example.org/kit", Attempts: testCase.attempts, Fallback: testCase.fallback}
			err := ValidateTransportPlan(plan)
			if err == nil {
				t.Fatalf("ValidateTransportPlan(%+v) succeeded", plan)
			}
			if ErrorCode(err) != testCase.code {
				t.Fatalf("err = %v, want code %s", err, testCase.code)
			}
		})
	}
}

// A narrowing bound on the port/alias lane refusal: it fires only for
// explicit ports and alias fields. Removing either property from an
// otherwise identical plan must admit it (mirror) or move it to a
// different class (undeclared mirror), never keep the lane refusal.
func TestPortAliasRefusalNarrows(t *testing.T) {
	t.Parallel()
	port := TransportAttempt{URL: "https://example.org:8443/kit.git", Authentication: "team-https",
		ResolvedHost: "example.org", ResolvedPort: 8443, HasExplicitPort: true}
	plan := TransportPlan{Identity: "example.org/kit", Attempts: []TransportAttempt{port}, Fallback: TransportFallbackNone}
	if err := ValidateTransportPlan(plan); ErrorCode(err) != CodeIdentityInvalid {
		t.Fatalf("port plan err = %v, want %s", err, CodeIdentityInvalid)
	}
	// Dropping the port (URL and provenance together) admits the same
	// endpoint: the refusal is the port, not the host or provider.
	unported := port
	unported.URL = "https://example.org/kit.git"
	unported.ResolvedPort, unported.HasExplicitPort = 0, false
	if err := ValidateTransportPlan(TransportPlan{Identity: "example.org/kit",
		Attempts: []TransportAttempt{unported}, Fallback: TransportFallbackNone}); err != nil {
		t.Fatalf("unported plan err = %v, want nil", err)
	}
	// Dropping only the carried provenance keeps a refusal, but as a
	// mistranslation — never a silent strip.
	stripped := port
	stripped.ResolvedHost, stripped.ResolvedPort, stripped.HasExplicitPort = "", 0, false
	if err := ValidateTransportPlan(TransportPlan{Identity: "example.org/kit",
		Attempts: []TransportAttempt{stripped}, Fallback: TransportFallbackNone}); ErrorCode(err) != CodeRepositoryPolicyInvalid {
		t.Fatalf("provenance-stripped port plan err = %v, want %s", err, CodeRepositoryPolicyInvalid)
	}
}

// Revision-1 plans validate exactly as before: no provenance is
// required, none is inferred, and the error vocabulary is unchanged.
func TestValidateTransportPlanRevision1Unchanged(t *testing.T) {
	t.Parallel()
	legacy := TransportPlan{Identity: "example.org/kit",
		Attempts: []TransportAttempt{{URL: "https://example.org/kit.git"}}, Fallback: TransportFallbackNone}
	if err := ValidateTransportPlan(legacy); err != nil {
		t.Fatalf("legacy plan err = %v, want nil", err)
	}
	mismatch := TransportPlan{Identity: "example.org/kit",
		Attempts: []TransportAttempt{{URL: "https://other.example/kit.git"}}, Fallback: TransportFallbackNone}
	err := ValidateTransportPlan(mismatch)
	if err == nil || ErrorCode(err) != CodeRepositoryPolicyInvalid ||
		!strings.Contains(err.Error(), `canonicalizes to "other.example/kit", want "example.org/kit"`) {
		t.Fatalf("mismatch err = %v, want the byte-identical revision-1 diagnostic", err)
	}
}

// The §5 endpoint URL grammar: explicit URI-form ports split off for
// the identity check; every other spelling behaves as the lane grammar.
func TestParseTransportEndpointURL(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name         string
		raw          string
		wantHost     string
		wantPort     int
		wantHasPort  bool
		wantIdentity string
		wantOK       bool
	}{
		{"https port", "https://example.org:8443/kit.git", "example.org", 8443, true, "example.org/kit", true},
		{"https default-shaped port one", "https://example.org:1/kit.git", "example.org", 1, true, "example.org/kit", true},
		{"https upper bound", "https://example.org:65535/kit.git", "example.org", 65535, true, "example.org/kit", true},
		{"ssh user and port", "ssh://git@example.org:2222/kit.git", "example.org", 2222, true, "example.org/kit", true},
		{"port-free https", "https://example.org/kit.git", "example.org", 0, false, "example.org/kit", true},
		{"port-free scp", "git@example.org:kit.git", "example.org", 0, false, "example.org/kit", true},
		{"scp colon segment is a path, never a port", "git@example.org:2222/kit.git", "example.org", 0, false, "example.org/2222/kit", true},
		{"port zero", "https://example.org:0/kit.git", "", 0, false, "", false},
		{"port above range", "https://example.org:65536/kit.git", "", 0, false, "", false},
		{"leading-zero port", "ssh://git@example.org:0222/kit.git", "", 0, false, "", false},
		{"empty port", "https://example.org:/kit.git", "", 0, false, "", false},
		{"non-numeric port", "https://example.org:https/kit.git", "", 0, false, "", false},
		{"userinfo on https with port", "https://user@example.org:8443/kit.git", "", 0, false, "", false},
		{"missing host", "https://:8443/kit.git", "", 0, false, "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			host, port, hasPort, identity, err := parseTransportEndpointURL(testCase.raw)
			if (err == nil) != testCase.wantOK {
				t.Fatalf("parseTransportEndpointURL(%q) err = %v, want ok=%v", testCase.raw, err, testCase.wantOK)
			}
			if !testCase.wantOK {
				return
			}
			if host != testCase.wantHost || port != testCase.wantPort || hasPort != testCase.wantHasPort || identity != testCase.wantIdentity {
				t.Fatalf("parseTransportEndpointURL(%q) = %q, %d, %v, %q", testCase.raw, host, port, hasPort, identity)
			}
		})
	}
}

// Exhaustion stays closed-vocabulary with revision-2 records: the
// listed URL (with port), alias, and mirror host reach the
// machine-private trace only, never the error.
func TestExhaustionV2RecordsStayClosed(t *testing.T) {
	t.Parallel()
	plan := TransportPlan{
		Identity: "example.org/kit",
		Attempts: []TransportAttempt{
			{URL: "https://example.org:8443/kit.git", Authentication: "prov-a",
				ResolvedHost: "example.org", ResolvedPort: 8443, HasExplicitPort: true},
			{URL: "https://mirror.example.net/kit.git", Authentication: "prov-b",
				MirrorOf: "example.org/kit", ResolvedHost: "mirror.example.net"},
		},
		Fallback: TransportFallbackAvailabilityAuth,
	}
	records := []AttemptRecord{
		{Index: 1, URL: plan.Attempts[0].URL, Transport: "https", Provider: "prov-a",
			Class: FailureAvailability, NetworkAttempted: true, Identity: plan.Identity,
			ResolvedHost: "example.org", ResolvedPort: 8443},
		{Index: 2, URL: plan.Attempts[1].URL, Transport: "https", Provider: "prov-b",
			Class: FailureAuth, NetworkAttempted: true, Identity: plan.Identity,
			ResolvedHost: "mirror.example.net", MirrorOf: "example.org/kit"},
	}
	message := exhaustionError(plan, records, false).Error()
	for _, want := range []string{CodeRepositoryEndpointUnavailable, "example.org/kit", "source-policy.json"} {
		if !strings.Contains(message, want) {
			t.Errorf("exhaustion %q lacks %q", message, want)
		}
	}
	for _, forbidden := range []string{"https://", "mirror.example.net", "8443", "prov-a-secret", "\n"} {
		if strings.Contains(message, forbidden) {
			t.Errorf("exhaustion %q contains %q", message, forbidden)
		}
	}
}
