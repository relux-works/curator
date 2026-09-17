package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/identity"
)

func TestDraftPolicyPublishedSchemaCases(t *testing.T) {
	files, err := filepath.Glob("testdata/draft-sources-v1/source-policy-v1/*.json")
	if err != nil || len(files) != 5 {
		t.Fatalf("published corpus: %d files, %v", len(files), err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			payload, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseSourcePolicy(payload, file)
			want := strings.HasPrefix(filepath.Base(file), "valid")
			if (err == nil) != want {
				t.Fatalf("valid=%v: %v", want, err)
			}
			if err != nil && !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
				t.Fatalf("error %q lacks class %q", err, CodeRepositoryPolicyInvalid)
			}
		})
	}
}

func policyDoc(repositories string) string {
	return `{"schema_version":1,"repositories":{` + repositories + `}}`
}

func policyEntry(endpoints, extra string) string {
	entry := `"endpoints":[` + endpoints + `],"fallback":"availability-auth"`
	if extra != "" {
		entry += "," + extra
	}
	return `{` + entry + `}`
}

func mustParsePolicy(t *testing.T, doc string) *SourcePolicy {
	t.Helper()
	policy, err := ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err != nil {
		t.Fatalf("ParseSourcePolicy: %v", err)
	}
	return policy
}

func mustRejectPolicy(t *testing.T, doc string) {
	t.Helper()
	_, err := ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err == nil {
		t.Fatalf("ParseSourcePolicy accepted %s", doc)
	}
	if !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("error %q lacks class %q", err, CodeRepositoryPolicyInvalid)
	}
}

func TestParseSourcePolicyDocumentShape(t *testing.T) {
	valid := policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}`)
	policy := mustParsePolicy(t, valid)
	if policy.Path != "source-policy.json" {
		t.Fatalf("Path = %q", policy.Path)
	}
	if len(policy.Repositories) != 1 || policy.Repositories["example.org/kit"].Fallback != FallbackNone {
		t.Fatalf("Repositories = %+v", policy.Repositories)
	}
	empty := mustParsePolicy(t, `{"schema_version":1,"repositories":{}}`)
	if len(empty.Repositories) != 0 {
		t.Fatalf("Repositories = %+v", empty.Repositories)
	}
	for _, doc := range []string{
		`not json`,
		`{"schema_version":1,"repositories":{}}trailing`,
		`["schema_version"]`,
		`{"schema_version":1,"repositories":{},"unexpected":true}`,
		`{"schema_version":1}`,
		`{"schema_version":1,"repositories":null}`,
		`{"schema_version":1,"repositories":[]}`,
		`{"repositories":{}}`,
		`{"schema_version":"1","repositories":{}}`,
		`{"schema_version":1.5,"repositories":{}}`,
		`{"schema_version":0,"repositories":{}}`,
		`{"schema_version":3,"repositories":{}}`,
		`{"schema_version":null,"repositories":{}}`,
	} {
		mustRejectPolicy(t, doc)
	}
}

func TestParseSourcePolicyRejectsRevision2Closed(t *testing.T) {
	_, err := ParseSourcePolicy([]byte(`{"schema_version":2,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}}}`), "source-policy.json")
	if err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) || !strings.Contains(err.Error(), "revision-2") {
		t.Fatalf("v2 rejection %v does not name the revision boundary", err)
	}
	mustRejectPolicy(t, `{"schema_version":1,"repositories":{},"aliases":{"corp-mirror":{"host":"mirror.example","authentication":"team-https"}}}`)
	mustRejectPolicy(t, policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https","alias":"corp-mirror"}],"fallback":"none"}`))
	mustRejectPolicy(t, policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https","mirror_of":"example.org/kit"}],"fallback":"none"}`))
}

func TestParseSourcePolicyEntryKeysAreExactCanonical(t *testing.T) {
	mkdoc := func(key string) string {
		return policyDoc(`"` + key + `":` + policyEntry(`{"url":"https://example.org/kit.git","authentication":"team-https"}`, ""))
	}
	for _, key := range []string{
		"example.org/kit",
		"example.org/Kit",
		"example.org/a/b/c",
		"e.co/x",
		"example.org/a_b.c-d/e",
	} {
		doc := policyDoc(`"` + key + `":` + policyEntry(`{"url":"https://`+key+`.git","authentication":"team-https"}`, ""))
		mustParsePolicy(t, doc)
	}
	for _, key := range []string{
		"Example.org/kit",
		"example.org/kit.git",
		"example.org:8443/kit",
		"user@example.org/kit",
		"example.org/../kit",
		"example.org/./kit",
		"example.org/kit/",
		"example.org",
		"",
		"example.org/kit?x",
		"example.org/ki t",
		"ssh://example.org/kit",
		"git@example.org:kit.git",
	} {
		mustRejectPolicy(t, mkdoc(key))
	}
}

func TestParseSourcePolicyEndpoints(t *testing.T) {
	one := func(url, auth string) string {
		return policyDoc(`"example.org/kit":` + policyEntry(`{"url":"`+url+`","authentication":"`+auth+`"}`, ""))
	}
	two := func(first, second string) string {
		return policyDoc(`"example.org/kit":` + policyEntry(first+","+second, ""))
	}
	ssh := `{"url":"git@example.org:kit.git","authentication":"team-ssh"}`
	https := `{"url":"https://example.org/kit.git","authentication":"team-https"}`
	policy := mustParsePolicy(t, two(ssh, https))
	entry := policy.Repositories["example.org/kit"]
	if len(entry.Endpoints) != 2 || entry.Endpoints[0].URL != "git@example.org:kit.git" || entry.Endpoints[1].Authentication != "team-https" {
		t.Fatalf("Endpoints = %+v", entry.Endpoints)
	}
	mustParsePolicy(t, one("https://example.org/kit", "team-https"))
	mustParsePolicy(t, one("ssh://git@example.org/kit.git", "team-ssh"))

	for _, doc := range []string{
		// Identity mismatch: the published endpoint-identity-mismatch refusal.
		one("https://evil.org/kit", "team-https"),
		one("https://example.org/other", "team-https"),
		one("git@example.org:kit/extra.git", "team-ssh"),
		// Repeated URLs fail even with distinct providers.
		two(https, `{"url":"https://example.org/kit.git","authentication":"other"}`),
		two(ssh, ssh),
		// Unsafe or foreign grammar.
		one("http://example.org/kit", "team-https"),
		one("https://example.org:8443/kit", "team-https"),
		one("https://example.org/kit?x=1", "team-https"),
		one("git://example.org/kit", "team-git"),
		one("not a url", "team-https"),
		one("", "team-https"),
		// Provider refs are opaque identifiers, never commands or paths.
		one("https://example.org/kit", ""),
		one("https://example.org/kit", "/bin/sh"),
		one("https://example.org/kit", "a b"),
		one("https://example.org/kit", "a/b"),
		one("https://example.org/kit", "../x"),
		one("https://example.org/kit", `C:\x`),
		one("https://example.org/kit", "a:b"),
		one("https://example.org/kit", "a@b"),
		// Shape violations.
		policyDoc(`"example.org/kit":{"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[` + ssh + `,` + https + `,{"url":"ssh://example.org/kit","authentication":"third"}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":"https://example.org/kit","fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit"}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"authentication":"team-https"}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":1,"authentication":"team-https"}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit","authentication":1}],"fallback":"none"}`),
		policyDoc(`"example.org/kit":{"endpoints":[` + ssh + `],"fallback":"none","helper":"evil"}`),
		policyDoc(`"example.org/kit":"https://example.org/kit"`),
	} {
		mustRejectPolicy(t, doc)
	}
}

func TestParseSourcePolicyPinAndFallback(t *testing.T) {
	pinned := func(pin string) string {
		return policyDoc(`"example.org/kit":{"endpoints":[{"url":"git@example.org:kit.git","authentication":"team-ssh"},{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"availability-auth","pin":"` + pin + `"}`)
	}
	if got := mustParsePolicy(t, pinned("git@example.org:kit.git")).Repositories["example.org/kit"].Pin; got != "git@example.org:kit.git" {
		t.Fatalf("Pin = %q", got)
	}
	if got := mustParsePolicy(t, pinned("https://example.org/kit.git")).Repositories["example.org/kit"].Pin; got != "https://example.org/kit.git" {
		t.Fatalf("Pin = %q", got)
	}
	for _, doc := range []string{
		// Pin must equal a listed URL exactly: same identity is not enough.
		pinned("https://example.org/kit"),
		pinned("ssh://example.org/kit.git"),
		pinned(""),
		pinned("https://evil.org/kit"),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none","pin":1}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none","pin":null}`),
		// Fallback is required and closed.
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}]}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"any-error"}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":null}`),
		policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"NONE"}`),
	} {
		mustRejectPolicy(t, doc)
	}
}

func TestParseSourcePolicyRootInputs(t *testing.T) {
	withInputs := func(inputs string) string {
		return `{"schema_version":1,"repositories":{},"root_inputs":{` + inputs + `}}`
	}
	policy := mustParsePolicy(t, withInputs(`"local":["SKILL.md","scripts"]`))
	if len(policy.RootInputs["local"]) != 2 {
		t.Fatalf("RootInputs = %+v", policy.RootInputs)
	}
	if got := mustParsePolicy(t, `{"schema_version":1,"repositories":{}}`).RootInputs; got != nil {
		t.Fatalf("RootInputs = %+v", got)
	}
	for _, doc := range []string{
		withInputs(`"local":[]`),
		withInputs(`"local":["SKILL.md","SKILL.md"]`),
		withInputs(`"local":["refs","refs/guide.md"]`),
		withInputs(`"local":["refs/guide.md","refs"]`),
		withInputs(`"local":["/abs"]`),
		withInputs(`"local":["../up"]`),
		withInputs(`"local":["a/../b"]`),
		withInputs(`"local":["a\\b"]`),
		withInputs(`"local":[""]`),
		withInputs(`"local":["trailing."]`),
		withInputs(`"local":"SKILL.md"`),
		withInputs(`"not an alias":["SKILL.md"]`),
		withInputs(`"":["SKILL.md"]`),
		`{"schema_version":1,"repositories":{},"root_inputs":[]}`,
		`{"schema_version":1,"repositories":{},"root_inputs":null}`,
	} {
		mustRejectPolicy(t, doc)
	}
}

func TestLoadSourcePolicy(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "absent", "source-policy.json")
	policy, err := LoadSourcePolicy(missing)
	if err != nil || policy != nil {
		t.Fatalf("missing file: policy=%v err=%v", policy, err)
	}
	malformed := filepath.Join(dir, "source-policy.json")
	if err := os.WriteFile(malformed, []byte(`{"schema_version":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSourcePolicy(malformed); err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("malformed file err=%v", err)
	}
	if _, err := LoadSourcePolicy(dir); err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("unreadable path err=%v", err)
	}
	valid := filepath.Join(dir, "valid.json")
	doc := policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}`)
	if err := os.WriteFile(valid, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	policy, err = LoadSourcePolicy(valid)
	if err != nil || policy == nil || policy.Path != valid {
		t.Fatalf("valid file: policy=%+v err=%v", policy, err)
	}
	if policy.Repositories["example.org/kit"].Endpoints[0].Authentication != "team-https" {
		t.Fatalf("Repositories = %+v", policy.Repositories)
	}
}

func TestLoadSourcePolicyRejectsNullOptionalFields(t *testing.T) {
	for _, tc := range []struct{ name, doc string }{
		{"null root_inputs", `{"schema_version":1,"repositories":{},"root_inputs":null}`},
		{"null pin", `{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team"}],"fallback":"none","pin":null}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "source-policy.json")
			if err := os.WriteFile(file, []byte(tc.doc), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := LoadSourcePolicy(file)
			if err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
				t.Fatalf("LoadSourcePolicy(%s) err=%v", tc.name, err)
			}
		})
	}
	// Controls: omitted optionals and valid values still load.
	for _, tc := range []struct{ name, doc string }{
		{"omitted optionals", policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team"}],"fallback":"none"}`)},
		{"valid optionals", `{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team"}],"fallback":"none","pin":"https://example.org/kit.git"}},"root_inputs":{"local":["SKILL.md"]}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "source-policy.json")
			if err := os.WriteFile(file, []byte(tc.doc), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadSourcePolicy(file); err != nil {
				t.Fatalf("LoadSourcePolicy(%s): %v", tc.name, err)
			}
		})
	}
}

func TestResolveRepositoryEndpoints(t *testing.T) {
	policy := mustParsePolicy(t, policyDoc(`"example.org/kit":{"endpoints":[{"url":"git@example.org:kit.git","authentication":"team-ssh"},{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"availability-auth"}`))

	resolved, err := ResolveRepositoryEndpoints(nil, "https://example.org/kit.git", "")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Identity != "example.org/kit" || len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != "https://example.org/kit.git" || resolved.Attempts[0].Authentication != "" || resolved.Fallback != FallbackNone {
		t.Fatalf("URL without policy: %+v", resolved)
	}

	resolved, err = ResolveRepositoryEndpoints(policy, "https://other.org/kit.git", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != "https://other.org/kit.git" || resolved.Fallback != FallbackNone {
		t.Fatalf("URL without entry: %+v", resolved)
	}

	// A URL hint for the second endpoint never reorders the machine list.
	resolved, err = ResolveRepositoryEndpoints(policy, "https://example.org/kit.git", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Attempts) != 2 || resolved.Attempts[0].URL != "git@example.org:kit.git" || resolved.Attempts[0].Authentication != "team-ssh" || resolved.Attempts[1].URL != "https://example.org/kit.git" || resolved.Fallback != FallbackAvailabilityAuth {
		t.Fatalf("URL with entry: %+v", resolved)
	}

	resolved, err = ResolveRepositoryEndpoints(policy, "", "example.org/kit")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Identity != "example.org/kit" || len(resolved.Attempts) != 2 || resolved.Fallback != FallbackAvailabilityAuth {
		t.Fatalf("logical with entry: %+v", resolved)
	}

	for _, candidate := range []*SourcePolicy{nil, policy} {
		_, err = ResolveRepositoryEndpoints(candidate, "", "example.org/missing")
		if err == nil || !strings.Contains(err.Error(), CodeRepositoryEndpointUnavailable) {
			t.Fatalf("logical without entry err=%v", err)
		}
		if !strings.Contains(err.Error(), "example.org/missing") {
			t.Fatalf("unavailable error %q names no identity", err)
		}
	}

	pinned := mustParsePolicy(t, policyDoc(`"example.org/kit":{"endpoints":[{"url":"git@example.org:kit.git","authentication":"team-ssh"},{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"availability-auth","pin":"https://example.org/kit.git"}`))
	resolved, err = ResolveRepositoryEndpoints(pinned, "", "example.org/kit")
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != "https://example.org/kit.git" || resolved.Attempts[0].Authentication != "team-https" || resolved.Fallback != FallbackNone {
		t.Fatalf("pin: %+v", resolved)
	}

	unpinned := mustParsePolicy(t, policyDoc(`"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}`))
	resolved, err = ResolveRepositoryEndpoints(unpinned, "", "example.org/kit")
	if err != nil || len(resolved.Attempts) != 1 || resolved.Fallback != FallbackNone {
		t.Fatalf("single endpoint: %+v err=%v", resolved, err)
	}

	for _, args := range [][2]string{
		{"https://example.org/kit.git", "example.org/kit"},
		{"", ""},
		{"http://example.org/kit", ""},
		{"not a url", ""},
		{"", "Example.org/kit"},
		{"", "example.org/kit.git"},
		{"", "https://example.org/kit"},
	} {
		if _, err := ResolveRepositoryEndpoints(policy, args[0], args[1]); err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("Resolve(%q,%q) err=%v", args[0], args[1], err)
		}
	}
}

func TestResolveRepositoryEndpointsRejectsHandBuiltViolations(t *testing.T) {
	badPin := &SourcePolicy{Repositories: map[string]RepositoryPolicy{
		"example.org/kit": {Identity: "example.org/kit", Endpoints: []Endpoint{{URL: "https://example.org/kit", Authentication: "team-https"}}, Pin: "https://example.org/other", Fallback: FallbackNone},
	}}
	if _, err := ResolveRepositoryEndpoints(badPin, "", "example.org/kit"); err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("bad pin err=%v", err)
	}
	empty := &SourcePolicy{Repositories: map[string]RepositoryPolicy{
		"example.org/kit": {Identity: "example.org/kit", Fallback: FallbackNone},
	}}
	if _, err := ResolveRepositoryEndpoints(empty, "", "example.org/kit"); err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("empty endpoints err=%v", err)
	}
}

// The policy loader validates entry keys through identity.DraftCanonicalKey
// and endpoint URLs through the closed lane grammar; the two must agree, or
// a valid key could never match a valid endpoint. Every accepted key is a
// fixed point of the lane canonicalizer.
func TestPolicyKeyAndEndpointIdentityAgree(t *testing.T) {
	keys := []string{
		"example.org/kit",
		"example.org/Kit",
		"example.org/a/b/c",
		"e.co/x",
		"example.org/a_b.c-d/e",
		"Example.org/kit",
		"example.org/kit.git",
		"example.org:8443/kit",
		"user@example.org/kit",
		"example.org/../kit",
		"example.org",
		"",
		"example.org/ki t",
	}
	for _, key := range keys {
		accepted := identity.DraftCanonicalKey(key)
		parsed, err := buildrepo.ParseSource("https://" + key)
		fixed := err == nil && parsed.Identity == key
		if accepted != fixed {
			t.Fatalf("key %q: identity=%v lane-fixed-point=%v", key, accepted, fixed)
		}
		if accepted && parsed.Identity != key {
			t.Fatalf("key %q: lane identity %q", key, parsed.Identity)
		}
	}
}
