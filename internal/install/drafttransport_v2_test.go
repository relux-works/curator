package install

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
)

// Revision-2 bounded resolution and provenance through the production
// entry point (repository-transport §§6-7): ExternalDeps.acquireDraftNetwork
// with the draft/opt-in switch on.
//
// Every test below drives the production caller against the stand-in git
// harness from drafttransport_test.go: the fake git fails a fetch with
// fixture stderr or rewrites it to a fixture bare repository, and
// everything else is the real lane. A refusal row is covered only when
// the test asserts the published failure class with zero fetches and
// zero trace records at this entry point.

// --- revision-2 fixtures ---

const (
	draftMirrorHTTPS = "https://mirror.fixture.test/repository.git"
	draftPortHTTPS   = "https://fixture.test:8443/repository.git"
	draftPortSSH     = "ssh://git@fixture.test:2222/repository.git"
)

const draftMirrorDNSStderr = "fatal: unable to access 'https://mirror.fixture.test/repository.git': Could not resolve host: mirror.fixture.test"

// v2deps routes one fetch through the production caller with the switch
// on, capturing machine-private provenance records through the trace hook.
func v2deps(tool buildrepo.GitTool, policy, providers string, reader buildrepo.ProviderSecretReader, records *[]buildrepo.AttemptRecord) ExternalDeps {
	return ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policy,
		DraftProvidersPath: providers, DraftProviderReader: reader,
		DraftTransportTrace: func(record buildrepo.AttemptRecord) { *records = append(*records, record) },
		Audit:               func(context.Context, buildrepo.AuditSubject) error { return nil }}
}

func v2acquire(t *testing.T, deps ExternalDeps, tool buildrepo.GitTool, lock buildrepo.LockedCommit) (*buildrepo.Snapshot, error) {
	t.Helper()
	return deps.acquireDraftNetwork(context.Background(), tool,
		draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
}

func v2anonymousProvidersDoc(names ...string) string {
	var providers []string
	for _, name := range names {
		providers = append(providers, fmt.Sprintf(`%q:{"https":{"anonymous":true}}`, name))
	}
	return `{"schema_version":1,"providers":{` + strings.Join(providers, ",") + `}}`
}

func v2mirrorEndpoint(url, auth, extra string) string {
	endpoint := fmt.Sprintf(`{"url":%q,"authentication":%q`, url, auth)
	if extra != "" {
		endpoint += "," + extra
	}
	return endpoint + `}`
}

func v2policy(entry, aliases string) string {
	doc := `{"schema_version":2,"repositories":{"fixture.test/repository":` + entry + `}`
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

// TestAcquireDraftNetworkRevision2Positives proves declared mirrors are
// ordinary endpoints at the production entry: list order, pin, and the
// revision-1 availability-auth fallback apply unchanged, provenance
// names the canonical identity, and each endpoint is attempted at most
// once.
func TestAcquireDraftNetworkRevision2Positives(t *testing.T) {
	bare, commit := draftBareFixture(t)
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}

	t.Run("declared mirror admitted with canonical provenance", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https",
			`"mirror_of":"fixture.test/repository"`), ""), ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("mirror-https"))
		var records []buildrepo.AttemptRecord
		snapshot, err := v2acquire(t, v2deps(tool, policy, providers, nil, &records), tool, lock)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftMirrorHTTPS+">") {
			t.Fatalf("fetches = %q, want exactly the listed mirror endpoint", fetches)
		}
		if len(records) != 1 {
			t.Fatalf("records = %+v, want one provenance record", records)
		}
		record := records[0]
		if record.Identity != "fixture.test/repository" {
			t.Fatalf("record identity = %q, want the canonical key, never the mirror host", record.Identity)
		}
		if record.URL != draftMirrorHTTPS || record.ResolvedHost != "mirror.fixture.test" || record.ResolvedPort != 0 {
			t.Fatalf("record = %+v, want the listed URL with the resolved mirror host", record)
		}
		if record.MirrorOf != "fixture.test/repository" || record.Alias != "" {
			t.Fatalf("record = %+v, want the mirror attestation and no alias", record)
		}
		if !record.Succeeded || !record.NetworkAttempted || record.Class != "" || record.Transport != "https" {
			t.Fatalf("record = %+v, want a successful https attempt", record)
		}
	})

	t.Run("mirror first with availability fallback attempts second", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS:  {stderr: draftMirrorDNSStderr},
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(`{"endpoints":[`+
			v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"mirror_of":"fixture.test/repository"`)+`,`+
			v2mirrorEndpoint(draftHTTPSPrimary, "team-https", "")+`],"fallback":"availability-auth"}`, ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("mirror-https", "team-https"))
		var records []buildrepo.AttemptRecord
		snapshot, err := v2acquire(t, v2deps(tool, policy, providers, nil, &records), tool, lock)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 2 {
			t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if len(records) != 2 || records[0].Class != buildrepo.FailureAvailability || !records[1].Succeeded {
			t.Fatalf("records = %+v, want availability then success", records)
		}
		for _, record := range records {
			if record.Identity != "fixture.test/repository" {
				t.Fatalf("record identity = %q, want the canonical key", record.Identity)
			}
		}
	})

	t.Run("pin selects the mirror and forbids fallback", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS:  {succeed: true, fileRepo: bare},
			draftHTTPSPrimary: {stderr: draftDNSStderr},
		})
		policy := draftPolicyFile(t, v2policy(`{"endpoints":[`+
			v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"mirror_of":"fixture.test/repository"`)+`,`+
			v2mirrorEndpoint(draftHTTPSPrimary, "team-https", "")+`],"fallback":"availability-auth",`+
			`"pin":`+fmt.Sprintf(`%q`, draftMirrorHTTPS)+`}`, ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("mirror-https", "team-https"))
		var records []buildrepo.AttemptRecord
		snapshot, err := v2acquire(t, v2deps(tool, policy, providers, nil, &records), tool, lock)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftMirrorHTTPS+">") {
			t.Fatalf("fetches = %q, want exactly the pinned mirror", fetches)
		}
		if len(records) != 1 || !records[0].Succeeded {
			t.Fatalf("records = %+v, want one successful pinned attempt", records)
		}
	})

	t.Run("fail-closed first failure stops before the mirror", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {stderr: draftTLSStderr},
			draftMirrorHTTPS:  {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(`{"endpoints":[`+
			v2mirrorEndpoint(draftHTTPSPrimary, "team-https", "")+`,`+
			v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"mirror_of":"fixture.test/repository"`)+`],"fallback":"availability-auth"}`, ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("team-https", "mirror-https"))
		var records []buildrepo.AttemptRecord
		_, err := v2acquire(t, v2deps(tool, policy, providers, nil, &records), tool, lock)
		if buildrepo.ErrorCode(err) != buildrepo.CodeSourceUnavailable {
			t.Fatalf("err = %v, want the lane diagnostic unchanged", err)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if len(records) != 1 || records[0].Class != buildrepo.FailureTLS {
			t.Fatalf("records = %+v, want one fail-closed tls record", records)
		}
	})

	t.Run("fallback none stops after the first availability failure", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {stderr: draftDNSStderr},
			draftMirrorHTTPS:  {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(v2entry(
			v2mirrorEndpoint(draftHTTPSPrimary, "team-https", "")+","+
				v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"mirror_of":"fixture.test/repository"`), ""), ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("team-https", "mirror-https"))
		var records []buildrepo.AttemptRecord
		_, err := v2acquire(t, v2deps(tool, policy, providers, nil, &records), tool, lock)
		if buildrepo.ErrorCode(err) != buildrepo.CodeRepositoryEndpointUnavailable {
			t.Fatalf("err = %v, want %s", err, buildrepo.CodeRepositoryEndpointUnavailable)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if len(records) != 1 {
			t.Fatalf("records = %+v, want one attempt record", records)
		}
	})
}

// TestAcquireDraftNetworkRevision2ExhaustionSanitized proves §7 secret
// handling at the production entry: exhaustion names the canonical
// identity with closed-vocabulary classifications while broker secrets
// and endpoint URLs stay out of the error.
func TestAcquireDraftNetworkRevision2ExhaustionSanitized(t *testing.T) {
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}
	tool, logPath := draftGitTool(t, map[string]draftGitArm{
		draftMirrorHTTPS:  {stderr: draftMirrorDNSStderr},
		draftHTTPSPrimary: {stderr: draftDNSStderr},
	})
	policy := draftPolicyFile(t, v2policy(`{"endpoints":[`+
		v2mirrorEndpoint(draftMirrorHTTPS, "prov-a", `"mirror_of":"fixture.test/repository"`)+`,`+
		v2mirrorEndpoint(draftHTTPSPrimary, "prov-b", "")+`],"fallback":"availability-auth"}`, ""))
	providers := draftProvidersFile(t, `{"schema_version":1,"providers":{`+
		`"prov-a":{"https":{"username":"a"}},"prov-b":{"https":{"username":"b"}}}}`)
	const secretMarker = "zz-secret-marker-260916"
	reader := &draftProviderReader{secrets: map[string]gitcred.HostCredential{
		"prov-a": {Username: "a", Secret: secretMarker + "-a"},
		"prov-b": {Username: "b", Secret: secretMarker + "-b"},
	}}
	var records []buildrepo.AttemptRecord
	_, err := v2acquire(t, v2deps(tool, policy, providers, reader, &records), tool, lock)
	if buildrepo.ErrorCode(err) != buildrepo.CodeRepositoryEndpointUnavailable {
		t.Fatalf("err = %v, want %s", err, buildrepo.CodeRepositoryEndpointUnavailable)
	}
	if fetches := draftFetchLines(t, logPath); len(fetches) != 2 {
		t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
	}
	message := err.Error()
	for _, want := range []string{buildrepo.CodeRepositoryEndpointUnavailable, "fixture.test/repository", "source-policy.json"} {
		if !strings.Contains(message, want) {
			t.Errorf("exhaustion %q lacks %q", message, want)
		}
	}
	for _, forbidden := range []string{secretMarker, "https://", "mirror.fixture.test", "\n"} {
		if strings.Contains(message, forbidden) {
			t.Errorf("exhaustion %q contains %q", message, forbidden)
		}
	}
	if len(records) != 2 {
		t.Fatalf("records = %+v, want two provenance records", records)
	}
	dump := fmt.Sprintf("%+v", records)
	for _, forbidden := range []string{secretMarker} {
		if strings.Contains(dump, forbidden) {
			t.Errorf("provenance records contain %q", forbidden)
		}
	}
	if records[0].Identity != "fixture.test/repository" || records[1].Identity != "fixture.test/repository" {
		t.Errorf("records = %+v, want the canonical identity on every record", records)
	}
	if records[0].ResolvedHost != "mirror.fixture.test" || records[0].MirrorOf != "fixture.test/repository" {
		t.Errorf("records = %+v, want mirror provenance on the first record", records)
	}
	// Broker addressing follows the connection: the mirror attempt
	// queries provider material for the mirror host, never the key host
	// and never another attempt's provider.
	wantQueries := []string{"prov-a@mirror.fixture.test", "prov-b@fixture.test"}
	if fmt.Sprint(reader.queries) != fmt.Sprint(wantQueries) {
		t.Errorf("broker queries = %q, want %q", reader.queries, wantQueries)
	}
}

// TestAcquireDraftNetworkRevision2Refusals proves every §6/§7 refusal
// row at the production entry: the published failure class with zero
// fetches and zero provenance records. The providers file is broken on
// purpose: a refusal that consulted it first would surface a different
// class.
func TestAcquireDraftNetworkRevision2Refusals(t *testing.T) {
	_, commit := draftBareFixture(t)
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}
	brokenProviders := draftProvidersFile(t, `not json`)
	mirrorAliases := `{"corp-mirror":{"host":"mirror.corp.example","authentication":"team-https"}}`

	for _, testCase := range []struct {
		name       string
		policy     string
		wantSubstr string
		wantCode   string
	}{
		{
			"undeclared mirror",
			v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", ""), ""), ""),
			config.CodeRepositoryMirrorUndeclared, "",
		},
		{
			"pin differs only by port",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", ""),
				`"pin":"https://fixture.test:8443/repository.git"`), ""),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"unknown alias",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"absent-alias"`), ""), `{}`),
			config.CodeRepositoryAliasUnknown, "",
		},
		{
			"alias to another host without mirror_of",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"corp-mirror"`), ""), mirrorAliases),
			config.CodeRepositoryMirrorUndeclared, "",
		},
		{
			"embedded alias host",
			v2policy(v2entry(v2mirrorEndpoint("ssh://git@corp-mirror/repository.git", "team-ssh", ""), ""),
				`{"corp-mirror":{"host":"mirror.corp.example","authentication":"team-ssh"}}`),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"alias authentication mismatch",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"fixture.test","authentication":"other-provider"}}`),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"chained alias",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"first","mirror_of":"fixture.test/repository"`), ""),
				`{"first":{"host":"second","authentication":"team-https"},"second":{"host":"fixture.test","authentication":"team-https"}}`),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"double port",
			v2policy(v2entry(v2mirrorEndpoint(draftPortHTTPS, "team-https", `"alias":"corp-mirror"`), ""),
				`{"corp-mirror":{"host":"fixture.test","port":9443,"authentication":"team-https"}}`),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"spurious mirror_of on same host",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"mirror_of":"fixture.test/repository"`), ""), ""),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"mirror_of mismatch",
			v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"mirror_of":"other.test/repository"`), ""), ""),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"mirror URL combined with alias",
			v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https", `"alias":"corp-mirror","mirror_of":"fixture.test/repository"`), ""),
				`{"corp-mirror":{"host":"mirror.corp.example","authentication":"mirror-https"}}`),
			config.CodeRepositoryPolicyInvalid, "",
		},
		{
			"explicit https port refused in the strict lane",
			v2policy(v2entry(v2mirrorEndpoint(draftPortHTTPS, "team-https", ""), ""), ""),
			"explicit port or host alias", buildrepo.CodeIdentityInvalid,
		},
		{
			"explicit ssh port refused in the strict lane",
			v2policy(v2entry(v2mirrorEndpoint(draftPortSSH, "team-ssh", ""), ""), ""),
			"explicit port or host alias", buildrepo.CodeIdentityInvalid,
		},
		{
			"same-host alias refused in the strict lane",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"local-vip"`), ""),
				`{"local-vip":{"host":"fixture.test","authentication":"team-https"}}`),
			"explicit port or host alias", buildrepo.CodeIdentityInvalid,
		},
		{
			"mirror alias refused in the strict lane",
			v2policy(v2entry(v2mirrorEndpoint(draftHTTPSPrimary, "team-https", `"alias":"corp-mirror","mirror_of":"fixture.test/repository"`), ""),
				mirrorAliases),
			"explicit port or host alias", buildrepo.CodeIdentityInvalid,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tool, logPath := draftGitTool(t, map[string]draftGitArm{
				draftHTTPSPrimary: {succeed: true, fileRepo: ""},
				draftMirrorHTTPS:  {succeed: true, fileRepo: ""},
				draftPortHTTPS:    {succeed: true, fileRepo: ""},
				draftPortSSH:      {succeed: true, fileRepo: ""},
			})
			policy := draftPolicyFile(t, testCase.policy)
			var records []buildrepo.AttemptRecord
			_, err := v2acquire(t, v2deps(tool, policy, brokenProviders, nil, &records), tool, lock)
			if err == nil {
				t.Fatalf("acquireDraftNetwork succeeded, want a refusal")
			}
			if testCase.wantCode != "" {
				if buildrepo.ErrorCode(err) != testCase.wantCode || !strings.Contains(err.Error(), testCase.wantSubstr) {
					t.Fatalf("err = %v, want code %s containing %q", err, testCase.wantCode, testCase.wantSubstr)
				}
			} else if !strings.Contains(err.Error(), testCase.wantSubstr) {
				t.Fatalf("err = %v, want class %s", err, testCase.wantSubstr)
			}
			if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
				t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
			}
			if len(records) != 0 {
				t.Fatalf("records = %+v, want zero attempts", records)
			}
		})
	}
}

// TestRevision1ReaderRejectsV2Shapes proves §7 compatibility at the
// loader the production caller uses: a revision-1 document carrying
// revision-2 members fails closed instead of silently ignoring endpoint
// properties it cannot enforce.
func TestRevision1ReaderRejectsV2Shapes(t *testing.T) {
	for _, doc := range []string{
		`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[{"url":"https://fixture.test/repository.git","authentication":"team-https"}],"fallback":"none"}},"aliases":{"a":{"host":"example.org","authentication":"team-https"}}}`,
		`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[{"url":"https://fixture.test/repository.git","authentication":"team-https","mirror_of":"fixture.test/repository"}],"fallback":"none"}}}`,
		`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[{"url":"https://fixture.test:8443/repository.git","authentication":"team-https"}],"fallback":"none"}}}`,
		`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[{"url":"https://fixture.test/repository.git","authentication":"team-https","alias":"a"}],"fallback":"none"}}}`,
	} {
		_, err := config.ParseSourcePolicy([]byte(doc), "source-policy.json")
		if err == nil || !strings.Contains(err.Error(), config.CodeRepositoryPolicyInvalid) {
			t.Fatalf("doc %s: err = %v, want %s", doc, err, config.CodeRepositoryPolicyInvalid)
		}
	}
	// The revision-2 reader still accepts the bare revision-1 shape with
	// revision-1 semantics: no alias table, no endpoint properties.
	policy, err := config.ParseSourcePolicy([]byte(
		`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[{"url":"https://fixture.test/repository.git","authentication":"team-https"}],"fallback":"none"}}}`),
		"source-policy.json")
	if err != nil {
		t.Fatal(err)
	}
	if policy.SchemaVersion != config.SourcePolicySchemaVersion || policy.Aliases != nil {
		t.Fatalf("policy = %+v, want a bare schema-1 load", policy)
	}
	resolved, err := config.ResolveRepositoryEndpoints(policy, draftHTTPSPrimary, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].MirrorOf != "" || resolved.Attempts[0].Alias != "" || resolved.Attempts[0].HasExplicitPort {
		t.Fatalf("resolved = %+v, want a bare revision-1 attempt", resolved)
	}
}

// TestUserConfigIgnoredByResolvedLane proves §5 user-configuration
// isolation at the production entry: hostile process-owned Git and SSH
// configuration never redirects a fetch. The lane pins its own config
// paths and wrapper, so planting insteadOf and GIT_SSH redirection in
// the process environment must leave the declared-URL fetch literal.
func TestUserConfigIgnoredByResolvedLane(t *testing.T) {
	bare, commit := draftBareFixture(t)
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}

	hostileGitConfig := draftPolicyFile(t, "[url \"https://evil.test/\"]\n\tinsteadOf = https://fixture.test/\n")
	t.Setenv("GIT_CONFIG_GLOBAL", hostileGitConfig)
	t.Setenv("GIT_CONFIG_SYSTEM", hostileGitConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_SSH", "/nonexistent-evil-ssh")

	t.Run("https insteadOf never applied without a policy", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		missing := draftPolicyFile(t, `{"schema_version":1,"repositories":{}}`)
		// No entry for the declared identity: absent entry attempts the
		// declared URL once with existing lane policy.
		deps := ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: missing,
			Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
		snapshot, err := deps.acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftHTTPSPrimary+">") ||
			strings.Contains(fetches[0], "evil.test") {
			t.Fatalf("fetches = %q, want exactly the declared URL, literally", fetches)
		}
	})

	t.Run("ssh wrapper ignores hostile GIT_SSH without a policy", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftSSHPrimary: {succeed: true, fileRepo: bare},
		})
		ssh := draftSSHCredentials(t)
		tool.SSHCredentials = ssh
		missing := draftPolicyFile(t, `{"schema_version":1,"repositories":{}}`)
		deps := ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: missing,
			Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
		snapshot, err := deps.acquireDraftNetwork(context.Background(), tool,
			draftSSHPrimary, "ssh", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		fetches := draftFetchLines(t, logPath)
		if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftSSHPrimary+">") {
			t.Fatalf("fetches = %q, want exactly the declared URL, literally", fetches)
		}
		// The fetch keeps the tool's bound static wrapper: hostile
		// process GIT_SSH never reaches the lane, whose environment
		// carries only manager-owned values.
		if !strings.Contains(fetches[0], tool.SSHWrapper) || strings.Contains(fetches[0], "evil-ssh") {
			t.Fatalf("fetches = %q, want the bound static wrapper only", fetches)
		}
	})
}

// TestDraftTransportPlanV2Conversion proves the production caller maps
// revision-2 resolutions field-for-field: the executor's revalidation
// sees the loader's mirror_of, alias, and resolved address, never a
// mistranslation.
func TestDraftTransportPlanV2Conversion(t *testing.T) {
	t.Parallel()
	resolution := config.Resolution{Identity: "fixture.test/repository",
		Attempts: []config.Attempt{
			{URL: "https://mirror.fixture.test/repository.git", Authentication: "mirror-https",
				MirrorOf: "fixture.test/repository", ResolvedHost: "mirror.fixture.test"},
			{URL: "https://fixture.test/repository.git", Authentication: "team-https",
				ResolvedHost: "fixture.test"},
		}, Fallback: config.FallbackAvailabilityAuth}
	plan, err := draftTransportPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Identity != resolution.Identity || len(plan.Attempts) != 2 {
		t.Fatalf("plan = %+v", plan)
	}
	if plan.Attempts[0].MirrorOf != "fixture.test/repository" || plan.Attempts[0].ResolvedHost != "mirror.fixture.test" {
		t.Fatalf("plan attempts = %+v, want the carried mirror provenance", plan.Attempts)
	}
	if plan.Attempts[1].MirrorOf != "" || plan.Attempts[1].Alias != "" || plan.Attempts[1].HasExplicitPort {
		t.Fatalf("plan attempts = %+v, want a bare second attempt", plan.Attempts)
	}
	if plan.Fallback != buildrepo.TransportFallbackAvailabilityAuth {
		t.Fatalf("fallback = %q", plan.Fallback)
	}
}

// TestDraftPlanNeedsSSHRevision2 keeps SSH base discovery port-aware:
// a port-bearing ssh:// endpoint still needs the manager wrapper base
// even though the strict lane later refuses the port.
func TestDraftPlanNeedsSSHRevision2(t *testing.T) {
	t.Parallel()
	portSSH := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "ssh://git@fixture.test:2222/repository.git"}}}
	if !draftPlanNeedsSSH(portSSH) {
		t.Fatal("port-bearing ssh plan needs SSH")
	}
	portHTTPS := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "https://fixture.test:8443/repository.git"}}}
	if draftPlanNeedsSSH(portHTTPS) {
		t.Fatal("port-bearing https plan needs no SSH")
	}
}
