package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The revision-2 corpus mirrors the published schema cases
// (source-policy-v2 under conformance/draft-sources-v1): every file
// whose name starts with "valid" must load through the production
// loader entry, and every other file must fail with the policy-invalid
// class before any network I/O.
func TestDraftPolicySchema2Corpus(t *testing.T) {
	files, err := filepath.Glob("testdata/draft-sources-v1/source-policy-v2/*.json")
	if err != nil || len(files) != 13 {
		t.Fatalf("v2 corpus: %d files, %v", len(files), err)
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

// Schema 1 loads byte-identically under the revision-2 reader: same
// version stamp, no alias table, empty revision-2 endpoint fields, and
// the same effective entry bytes as before the superset.
func TestSchema1GoldenUnchanged(t *testing.T) {
	doc := `{"schema_version":1,"repositories":{"example.org/kit":{"endpoints":[` +
		`{"url":"git@example.org:kit.git","authentication":"team-ssh"},` +
		`{"url":"https://example.org/kit.git","authentication":"team-https"}],` +
		`"fallback":"availability-auth"}},"root_inputs":{"local":["SKILL.md","scripts"]}}`
	policy, err := ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err != nil {
		t.Fatalf("ParseSourcePolicy: %v", err)
	}
	if policy.SchemaVersion != SourcePolicySchemaVersion {
		t.Fatalf("SchemaVersion = %d", policy.SchemaVersion)
	}
	if policy.Aliases != nil {
		t.Fatalf("Aliases = %+v", policy.Aliases)
	}
	entry, ok := policy.Repositories["example.org/kit"]
	if !ok || len(entry.Endpoints) != 2 || entry.Fallback != FallbackAvailabilityAuth || entry.Pin != "" {
		t.Fatalf("Repositories = %+v", policy.Repositories)
	}
	for _, endpoint := range entry.Endpoints {
		if endpoint.MirrorOf != "" || endpoint.Alias != "" {
			t.Fatalf("schema-1 endpoint carries revision-2 fields: %+v", endpoint)
		}
	}
	type goldenEndpoint struct {
		URL            string `json:"url"`
		Authentication string `json:"authentication"`
	}
	type goldenEntry struct {
		Endpoints []goldenEndpoint `json:"endpoints"`
		Fallback  string           `json:"fallback"`
	}
	projected := goldenEntry{Fallback: entry.Fallback}
	for _, endpoint := range entry.Endpoints {
		projected.Endpoints = append(projected.Endpoints, goldenEndpoint{URL: endpoint.URL, Authentication: endpoint.Authentication})
	}
	encoded, err := json.Marshal(projected)
	if err != nil {
		t.Fatal(err)
	}
	const golden = `{"endpoints":[{"url":"git@example.org:kit.git","authentication":"team-ssh"},` +
		`{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"availability-auth"}`
	if string(encoded) != golden {
		t.Fatalf("golden mismatch:\n got %s\nwant %s", encoded, golden)
	}
	if len(policy.RootInputs["local"]) != 2 {
		t.Fatalf("RootInputs = %+v", policy.RootInputs)
	}
}

func v2doc(entry, aliases string) string {
	doc := `{"schema_version":2,"repositories":{"example.org/kit":` + entry + `}`
	if aliases != "" {
		doc += `,"aliases":` + aliases
	}
	return doc + `}`
}

func v2entry(endpoints, extra string) string {
	entry := `"endpoints":[` + endpoints + `],"fallback":"none"`
	if extra != "" {
		entry += "," + extra
	}
	return `{` + entry + `}`
}

func v2endpoint(url, auth, extra string) string {
	endpoint := `{"url":` + strconv.Quote(url) + `,"authentication":` + strconv.Quote(auth)
	if extra != "" {
		endpoint += "," + extra
	}
	return endpoint + `}`
}

func mustRejectWithClass(t *testing.T, doc, class string) {
	t.Helper()
	_, err := ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err == nil {
		t.Fatalf("ParseSourcePolicy accepted %s", doc)
	}
	if !strings.Contains(err.Error(), class) {
		t.Fatalf("error %q lacks class %q", err, class)
	}
}

// The four published positives load as an additive superset: ports,
// a declared mirror, alias resolution, and the bare revision-1 shape.
func TestParseSourcePolicyV2Positives(t *testing.T) {
	t.Run("port endpoints", func(t *testing.T) {
		doc := v2doc(v2entry(
			v2endpoint("ssh://git@example.org:2222/kit.git", "team-ssh", "")+","+
				v2endpoint("https://example.org:8443/kit.git", "team-https", ""), ""), "")
		policy := mustParsePolicy(t, doc)
		if policy.SchemaVersion != SourcePolicySchemaVersionV2 {
			t.Fatalf("SchemaVersion = %d", policy.SchemaVersion)
		}
		resolved, err := ResolveRepositoryEndpoints(policy, "", "example.org/kit")
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Identity != "example.org/kit" {
			t.Fatalf("Identity = %q", resolved.Identity)
		}
		if len(resolved.Attempts) != 2 {
			t.Fatalf("Attempts = %+v", resolved.Attempts)
		}
		first, second := resolved.Attempts[0], resolved.Attempts[1]
		if first.ResolvedHost != "example.org" || first.ResolvedPort != 2222 || !first.HasExplicitPort {
			t.Fatalf("first = %+v", first)
		}
		if second.ResolvedHost != "example.org" || second.ResolvedPort != 8443 || !second.HasExplicitPort {
			t.Fatalf("second = %+v", second)
		}
		if first.MirrorOf != "" || first.Alias != "" {
			t.Fatalf("port endpoint carries mirror state: %+v", first)
		}
	})

	t.Run("declared mirror", func(t *testing.T) {
		doc := v2doc(v2entry(v2endpoint("https://mirror.example.net/kit.git", "mirror-https", `"mirror_of":"example.org/kit"`), ""), "")
		policy := mustParsePolicy(t, doc)
		if got := policy.Repositories["example.org/kit"].Endpoints[0].MirrorOf; got != "example.org/kit" {
			t.Fatalf("MirrorOf = %q", got)
		}
		resolved, err := ResolveRepositoryEndpoints(policy, "", "example.org/kit")
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Identity != "example.org/kit" {
			t.Fatalf("Identity = %q, mirror host must never become identity", resolved.Identity)
		}
		attempt := resolved.Attempts[0]
		if attempt.ResolvedHost != "mirror.example.net" || attempt.ResolvedPort != 0 || attempt.HasExplicitPort {
			t.Fatalf("attempt = %+v", attempt)
		}
	})

	t.Run("alias resolution", func(t *testing.T) {
		doc := v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https",
			`"alias":"corp-mirror","mirror_of":"example.org/kit"`), ""),
			`{"corp-mirror":{"host":"mirror.corp.example","port":8443,"authentication":"team-https"}}`)
		policy := mustParsePolicy(t, doc)
		stored := policy.Repositories["example.org/kit"].Endpoints[0]
		if stored.Alias != "corp-mirror" || stored.MirrorOf != "example.org/kit" {
			t.Fatalf("stored = %+v", stored)
		}
		if policy.Aliases["corp-mirror"].Host != "mirror.corp.example" || !policy.Aliases["corp-mirror"].HasPort {
			t.Fatalf("Aliases = %+v", policy.Aliases)
		}
		resolved, err := ResolveRepositoryEndpoints(policy, "", "example.org/kit")
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Identity != "example.org/kit" {
			t.Fatalf("Identity = %q, alias target must never become identity", resolved.Identity)
		}
		attempt := resolved.Attempts[0]
		if attempt.ResolvedHost != "mirror.corp.example" || attempt.ResolvedPort != 8443 || !attempt.HasExplicitPort {
			t.Fatalf("attempt = %+v", attempt)
		}
	})

	t.Run("same-host alias needs no attestation", func(t *testing.T) {
		doc := v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"local-vip"`), ""),
			`{"local-vip":{"host":"example.org","authentication":"team-https"}}`)
		policy := mustParsePolicy(t, doc)
		resolved, err := ResolveRepositoryEndpoints(policy, "", "example.org/kit")
		if err != nil {
			t.Fatal(err)
		}
		attempt := resolved.Attempts[0]
		if resolved.Identity != "example.org/kit" || attempt.ResolvedHost != "example.org" || attempt.HasExplicitPort {
			t.Fatalf("resolved = %+v", resolved)
		}
	})

	t.Run("distinct port spellings are distinct endpoints", func(t *testing.T) {
		doc := v2doc(v2entry(
			v2endpoint("https://example.org/kit.git", "team-https", "")+","+
				v2endpoint("https://example.org:8443/kit.git", "team-https", ""),
			`"pin":"https://example.org:8443/kit.git"`), "")
		policy := mustParsePolicy(t, doc)
		resolved, err := ResolveRepositoryEndpoints(policy, "", "example.org/kit")
		if err != nil {
			t.Fatal(err)
		}
		if len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != "https://example.org:8443/kit.git" || !resolved.Attempts[0].HasExplicitPort {
			t.Fatalf("pinned = %+v", resolved)
		}
	})
}

// Every refusal row of §5 fails at the loader production entry with its
// published class: mirror_undeclared for a differing resolved host
// without attestation, alias_unknown for a dangling alias, and
// policy_invalid for every misuse.
func TestParseSourcePolicyV2Refusals(t *testing.T) {
	mirrorAliases := `{"corp-mirror":{"host":"mirror.corp.example","authentication":"team-https"}}`
	for _, tc := range []struct{ name, doc, class string }{
		{
			"undeclared mirror URL",
			v2doc(v2entry(v2endpoint("https://mirror.example.net/kit.git", "mirror-https", ""), ""), ""),
			CodeRepositoryMirrorUndeclared,
		},
		{
			"alias to another host without mirror_of",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""), mirrorAliases),
			CodeRepositoryMirrorUndeclared,
		},
		{
			"unknown alias with empty table",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"absent-alias"`), ""), `{}`),
			CodeRepositoryAliasUnknown,
		},
		{
			"unknown alias without table",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"absent-alias"`), ""), ""),
			CodeRepositoryAliasUnknown,
		},
		{
			"unknown alias beside a populated table",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"other-alias"`), ""), mirrorAliases),
			CodeRepositoryAliasUnknown,
		},
		{
			"mirror_of mismatch",
			v2doc(v2entry(v2endpoint("https://mirror.example.net/kit.git", "mirror-https", `"mirror_of":"other.example/kit"`), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"spurious mirror_of on same host",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"mirror_of":"example.org/kit"`), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"spurious mirror_of on same-host alias",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"local-vip","mirror_of":"example.org/kit"`), ""),
				`{"local-vip":{"host":"example.org","authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"mirror URL combined with alias",
			v2doc(v2entry(v2endpoint("https://mirror.example.net/kit.git", "mirror-https", `"alias":"corp-mirror","mirror_of":"example.org/kit"`), ""),
				`{"corp-mirror":{"host":"mirror.corp.example","authentication":"mirror-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"embedded alias host",
			v2doc(v2entry(v2endpoint("ssh://git@corp-mirror/kit.git", "team-ssh", ""), ""),
				`{"corp-mirror":{"host":"mirror.corp.example","authentication":"team-ssh"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"alias authentication mismatch",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"example.org","authentication":"other-provider"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"chained alias",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"first","mirror_of":"example.org/kit"`), ""),
				`{"first":{"host":"second","authentication":"team-https"},"second":{"host":"example.org","authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"double port",
			v2doc(v2entry(v2endpoint("https://example.org:8443/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"example.org","port":9443,"authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"pin differs only by port",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", ""), `"pin":"https://example.org:8443/kit.git"`), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"port zero",
			v2doc(v2entry(v2endpoint("https://example.org:0/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"port above range",
			v2doc(v2entry(v2endpoint("https://example.org:70000/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"ssh port above range",
			v2doc(v2entry(v2endpoint("ssh://git@example.org:65536/kit.git", "team-ssh", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"leading-zero port",
			v2doc(v2entry(v2endpoint("ssh://git@example.org:0222/kit.git", "team-ssh", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"empty port",
			v2doc(v2entry(v2endpoint("https://example.org:/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"non-numeric port",
			v2doc(v2entry(v2endpoint("https://example.org:https/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"scp colon segment is a path, never a port",
			v2doc(v2entry(v2endpoint("git@example.org:2222/kit.git", "team-ssh", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"mirror path must equal the key path",
			v2doc(v2entry(v2endpoint("https://mirror.example.net/other.git", "mirror-https", `"mirror_of":"example.org/kit"`), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"userinfo on https with port",
			v2doc(v2entry(v2endpoint("https://user@example.org:8443/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"uppercase alias name in table",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"Corp-Mirror"`), ""),
				`{"Corp-Mirror":{"host":"mirror.corp.example","authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"uppercase alias target host",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"Mirror.Corp.Example","authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"uppercase endpoint alias field",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"Corp-Mirror"`), ""), mirrorAliases),
			CodeRepositoryPolicyInvalid,
		},
		{
			"string alias port",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"example.org","port":"8443","authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"zero alias port",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"example.org","port":0,"authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"alias missing authentication",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"example.org"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"alias missing host",
			v2doc(v2entry(v2endpoint("https://example.org/kit.git", "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"authentication":"team-https"}}`),
			CodeRepositoryPolicyInvalid,
		},
		{
			"null mirror_of",
			v2doc(v2entry(`{"url":"https://example.org/kit.git","authentication":"team-https","mirror_of":null}`, ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"non-canonical mirror_of",
			v2doc(v2entry(v2endpoint("https://mirror.example.net/kit.git", "mirror-https", `"mirror_of":"Example.org/kit"`), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"null aliases table",
			`{"schema_version":2,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}},"aliases":null}`,
			CodeRepositoryPolicyInvalid,
		},
		{
			"unknown top-level member",
			`{"schema_version":2,"repositories":{"example.org/kit":{"endpoints":[{"url":"https://example.org/kit.git","authentication":"team-https"}],"fallback":"none"}},"unexpected":true}`,
			CodeRepositoryPolicyInvalid,
		},
		{
			"extra endpoint member",
			v2doc(v2entry(`{"url":"https://example.org/kit.git","authentication":"team-https","helper":"evil"}`, ""), ""),
			CodeRepositoryPolicyInvalid,
		},
		{
			"duplicate port-bearing endpoints",
			v2doc(v2entry(
				v2endpoint("https://example.org:8443/kit.git", "team-https", "")+","+
					v2endpoint("https://example.org:8443/kit.git", "team-https", ""), ""), ""),
			CodeRepositoryPolicyInvalid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustRejectWithClass(t, tc.doc, tc.class)
		})
	}
}

// The planning entry re-applies the fail-closed rows so hand-built
// policies cannot smuggle an unattested mirror or a dangling alias past
// the loader.
func TestResolveV2HandBuiltViolations(t *testing.T) {
	mirror := &SourcePolicy{Repositories: map[string]RepositoryPolicy{
		"example.org/kit": {Identity: "example.org/kit",
			Endpoints: []Endpoint{{URL: "https://mirror.example.net/kit.git", Authentication: "mirror-https"}},
			Fallback:  FallbackNone},
	}}
	if _, err := ResolveRepositoryEndpoints(mirror, "", "example.org/kit"); err == nil || !strings.Contains(err.Error(), CodeRepositoryMirrorUndeclared) {
		t.Fatalf("undeclared mirror err=%v", err)
	}
	dangling := &SourcePolicy{Repositories: map[string]RepositoryPolicy{
		"example.org/kit": {Identity: "example.org/kit",
			Endpoints: []Endpoint{{URL: "https://example.org/kit.git", Authentication: "team-https", Alias: "absent-alias"}},
			Fallback:  FallbackNone},
	}}
	if _, err := ResolveRepositoryEndpoints(dangling, "", "example.org/kit"); err == nil || !strings.Contains(err.Error(), CodeRepositoryAliasUnknown) {
		t.Fatalf("unknown alias err=%v", err)
	}
}
