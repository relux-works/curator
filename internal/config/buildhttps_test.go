package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// buildHTTPSConfig carries all three source shapes at three nesting depths, so
// one fixture exercises parsing, matching, and serialization.
const buildHTTPSConfig = `{
	"schema_version": 1, "skills_root": "/tmp/skills", "projects": {},
	"build_https": {
		"git.example.com": {"token": "git-credentials"},
		"git.example.com/relux-works": {"token": "keyring", "username": "oauth2"},
		"git.example.com/relux-works/portals": {"token_env": "PORTALS_TOKEN"}
	}
}`

func loadBuildHTTPSConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", buildHTTPSConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestParseBuildHTTPS(t *testing.T) {
	cfg := loadBuildHTTPSConfig(t)
	want := map[string]BuildHTTPSCredential{
		"git.example.com": {
			Scope: "git.example.com", Token: TokenSourceGitCredentials,
		},
		"git.example.com/relux-works": {
			Scope: "git.example.com/relux-works", Token: TokenSourceKeyring, Username: "oauth2",
		},
		"git.example.com/relux-works/portals": {
			Scope: "git.example.com/relux-works/portals", TokenEnv: "PORTALS_TOKEN",
		},
	}
	if !reflect.DeepEqual(cfg.BuildHTTPS, want) {
		t.Fatalf("build_https = %+v, want %+v", cfg.BuildHTTPS, want)
	}
	wantScopes := []string{
		"git.example.com", "git.example.com/relux-works",
		"git.example.com/relux-works/portals",
	}
	if got := cfg.BuildHTTPSScopes(); !reflect.DeepEqual(got, wantScopes) {
		t.Fatalf("scopes = %v, want %v", got, wantScopes)
	}
}

func TestParseBuildHTTPSAbsentAndEmpty(t *testing.T) {
	for name, text := range map[string]string{
		"absent": minimal,
		"null":   `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_https": null}`,
		"empty":  `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_https": {}}`,
	} {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(cfg.BuildHTTPS) != 0 {
				t.Fatalf("build_https = %+v, want empty", cfg.BuildHTTPS)
			}
			if len(cfg.BuildHTTPSScopes()) != 0 {
				t.Fatalf("scopes = %v, want none", cfg.BuildHTTPSScopes())
			}
			if _, ok := cfg.BuildHTTPSFor("git.example.com/relux-works/portals"); ok {
				t.Fatal("an unset build_https must select no credential")
			}
		})
	}
}

// TestBuildHTTPSSerializationRoundtrip writes the parsed credentials back out
// through BuildHTTPSObject and requires the reparsed config to be identical.
func TestBuildHTTPSSerializationRoundtrip(t *testing.T) {
	dir := t.TempDir()
	first := loadBuildHTTPSConfig(t)
	payload, err := json.Marshal(map[string]any{
		"schema_version": 1, "skills_root": "/tmp/skills", "projects": map[string]any{},
		"build_https": BuildHTTPSObject(first.BuildHTTPS),
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "roundtrip.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := Load(path, nil)
	if err != nil {
		t.Fatalf("serialized build_https did not parse: %v", err)
	}
	if !reflect.DeepEqual(second.BuildHTTPS, first.BuildHTTPS) {
		t.Fatalf("roundtrip = %+v, want %+v", second.BuildHTTPS, first.BuildHTTPS)
	}
}

func TestBuildHTTPSObjectOmitsUnsetFields(t *testing.T) {
	for name, tc := range map[string]struct {
		credential BuildHTTPSCredential
		want       string
	}{
		"token only": {
			BuildHTTPSCredential{Scope: "git.example.com", Token: TokenSourceGitCredentials},
			`{"git.example.com":{"token":"git-credentials"}}`,
		},
		"token with username": {
			BuildHTTPSCredential{Scope: "git.example.com", Token: TokenSourceKeyring, Username: "oauth2"},
			`{"git.example.com":{"token":"keyring","username":"oauth2"}}`,
		},
		"token_env only": {
			BuildHTTPSCredential{Scope: "git.example.com", TokenEnv: "MY_TOKEN"},
			`{"git.example.com":{"token_env":"MY_TOKEN"}}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(BuildHTTPSObject(map[string]BuildHTTPSCredential{
				tc.credential.Scope: tc.credential,
			}))
			if err != nil {
				t.Fatal(err)
			}
			if string(payload) != tc.want {
				t.Fatalf("serialized = %s, want %s", payload, tc.want)
			}
		})
	}
	if got := BuildHTTPSObject(nil); len(got) != 0 {
		t.Fatalf("BuildHTTPSObject(nil) = %v, want empty", got)
	}
}

// TestBuildHTTPSReusesBuildSSHScopeGrammar pins that scope validity is decided
// by the shared build_ssh grammar, not a parallel implementation of it.
func TestBuildHTTPSReusesBuildSSHScopeGrammar(t *testing.T) {
	cases := []struct {
		scope string
		want  bool
	}{
		{"git.example.com", true},
		{"git.example.com/relux-works", true},
		{"git.example.com/relux-works/portals", true},
		{"", false},
		{"Git.Example.com", false},
		{"git.example.com//portals", false},
		{"git.example.com/", false},
		{"git.example.com/.", false},
		{"git.example.com/..", false},
		{"git.example.com/a b", false},
		{"git.example.com/ünicode", false},
		{"ssh://git.example.com/repo", false},
		{strings.Repeat("a", 253), true},
		{strings.Repeat("a", 254), false},
	}
	for _, tc := range cases {
		credential := BuildHTTPSCredential{Scope: tc.scope, Token: TokenSourceGitCredentials}
		if got := ValidateBuildHTTPS(credential) == nil; got != tc.want {
			t.Errorf("ValidateBuildHTTPS(scope=%q) accepted = %v, want %v", tc.scope, got, tc.want)
		}
		if got := ValidBuildSSHScope(tc.scope); got != tc.want {
			t.Errorf("ValidBuildSSHScope(%q) = %v, want %v (build_https must reuse this)", tc.scope, got, tc.want)
		}
	}
}

func TestParseBuildHTTPSRejections(t *testing.T) {
	const prefix = `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_https": `
	cases := []struct {
		name string
		body string
		want string
	}{
		{"not an object", `[]`, "build_https: must be an object"},
		{"entry not an object", `{"git.example.com": "token"}`, "build_https.git.example.com: must be an object"},
		{"uppercase host", `{"Git.Example.com": {"token": "git-credentials"}}`, `scope "Git.Example.com"`},
		{"empty segment", `{"git.example.com//portals": {"token": "git-credentials"}}`, `scope "git.example.com//portals"`},
		{"trailing slash", `{"git.example.com/portals/": {"token": "git-credentials"}}`, `scope "git.example.com/portals/"`},
		{"dot segment", `{"git.example.com/./portals": {"token": "git-credentials"}}`, `scope "git.example.com/./portals"`},
		{"dotdot segment", `{"git.example.com/../portals": {"token": "git-credentials"}}`, `scope "git.example.com/../portals"`},
		{"scheme in scope", `{"ssh://git.example.com/repo": {"token": "git-credentials"}}`, "scope"},
		{"port in scope", `{"git.example.com:22/repo": {"token": "git-credentials"}}`, "scope"},
		{"empty scope", `{"": {"token": "git-credentials"}}`, `scope ""`},
		{"unknown entry field", `{"git.example.com": {"token": "git-credentials", "password": "x"}}`, `unsupported field(s): password`},
		{"unknown entry fields sorted", `{"git.example.com": {"token": "git-credentials", "user": "x", "port": 22}}`, `unsupported field(s): port, user`},
		{"no selection", `{"git.example.com": {"username": "x"}}`, "requires exactly one of 'token' or 'token_env'"},
		{"empty entry", `{"git.example.com": {}}`, "requires exactly one of 'token' or 'token_env'"},
		{"both token and token_env", `{"git.example.com": {"token": "git-credentials", "token_env": "X"}}`, "requires exactly one of 'token' or 'token_env'"},
		// A literal secret in 'token' must be rejected, not silently stored.
		{"literal secret in token", `{"git.example.com": {"token": "ghp_abc123secret"}}`, "secrets never live in the config"},
		{"token null", `{"git.example.com": {"token": null}}`, "build_https.git.example.com.token"},
		{"token number", `{"git.example.com": {"token": 1}}`, "build_https.git.example.com.token"},
		{"token empty", `{"git.example.com": {"token": ""}}`, "build_https.git.example.com.token"},
		{"token_env empty", `{"git.example.com": {"token_env": ""}}`, "build_https.git.example.com.token_env"},
		{"token_env null", `{"git.example.com": {"token_env": null}}`, "build_https.git.example.com.token_env"},
		{"token_env number", `{"git.example.com": {"token_env": 1}}`, "build_https.git.example.com.token_env"},
		{"token_env invalid identifier", `{"git.example.com": {"token_env": "1_TOKEN"}}`, "build_https.git.example.com.token_env"},
		{"token_env with dash", `{"git.example.com": {"token_env": "MY-TOKEN"}}`, "build_https.git.example.com.token_env"},
		{"token_env with space", `{"git.example.com": {"token_env": "MY TOKEN"}}`, "build_https.git.example.com.token_env"},
		{"username empty", `{"git.example.com": {"token": "git-credentials", "username": ""}}`, "build_https.git.example.com.username"},
		{"username null", `{"git.example.com": {"token": "git-credentials", "username": null}}`, "build_https.git.example.com.username"},
		{"username number", `{"git.example.com": {"token": "git-credentials", "username": 1}}`, "build_https.git.example.com.username"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text := prefix + tc.body + "}"
			_, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want mention of %q", err, tc.want)
			}
		})
	}
}

// TestParseBuildHTTPSReportsOneFaultDeterministically pins the scope
// ordering, so a config with several faults never reports a different one
// per run.
func TestParseBuildHTTPSReportsOneFaultDeterministically(t *testing.T) {
	const text = `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_https": {
		"z.example.com": {},
		"a.example.com": {}
	}}`
	for range 8 {
		_, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
		if err == nil || !strings.Contains(err.Error(), "build_https.a.example.com") {
			t.Fatalf("err = %v, want the lowest scope reported first", err)
		}
	}
}

func TestBuildHTTPSLongestPrefixMatch(t *testing.T) {
	cfg := loadBuildHTTPSConfig(t)
	cases := []struct {
		identity string
		want     string // matched scope, "" for no match
	}{
		{"git.example.com/relux-works/portals", "git.example.com/relux-works/portals"},
		{"git.example.com/relux-works/portals/sub", "git.example.com/relux-works/portals"},
		// Boundary: a scope matches on whole segments only, so a longer
		// repository name falls back to the enclosing organization scope.
		{"git.example.com/relux-works/portals-evil", "git.example.com/relux-works"},
		{"git.example.com/relux-works/portals.git-mirror", "git.example.com/relux-works"},
		{"git.example.com/relux-works", "git.example.com/relux-works"},
		{"git.example.com/relux-works-evil/portals", "git.example.com"},
		{"git.example.com/other/repo", "git.example.com"},
		// The same boundary rule applies to the host.
		{"git.example.community/relux-works/portals", ""},
		{"other.example.com/relux-works/portals", ""},
		{"", ""},
		{"git.example.com", ""}, // a host alone is not a canonical identity
		{"GIT.EXAMPLE.COM/relux-works/portals", ""},
		{"git.example.com/relux-works/portals.git", "git.example.com/relux-works"},
	}
	for _, tc := range cases {
		got, ok := cfg.BuildHTTPSFor(tc.identity)
		if tc.want == "" {
			if ok {
				t.Errorf("BuildHTTPSFor(%q) = %+v, want no match", tc.identity, got)
			}
			continue
		}
		if !ok || got.Scope != tc.want {
			t.Errorf("BuildHTTPSFor(%q) = %+v, %v; want scope %q", tc.identity, got, ok, tc.want)
		}
	}
}

func TestBuildHTTPSMatchCarriesCredential(t *testing.T) {
	cfg := loadBuildHTTPSConfig(t)
	credential, ok := cfg.BuildHTTPSFor("git.example.com/relux-works/portals/deep")
	if !ok || credential.TokenEnv != "PORTALS_TOKEN" || credential.Token != "" {
		t.Fatalf("match = %+v, %v", credential, ok)
	}
	organization, ok := cfg.BuildHTTPSFor("git.example.com/relux-works/other")
	if !ok || organization.Token != TokenSourceKeyring || organization.Username != "oauth2" {
		t.Fatalf("organization match = %+v, %v", organization, ok)
	}
}

func TestBuildHTTPSEffectiveUsername(t *testing.T) {
	if got := (BuildHTTPSCredential{}).EffectiveUsername(); got != BuildHTTPSDefaultUsername {
		t.Fatalf("EffectiveUsername() = %q, want %q", got, BuildHTTPSDefaultUsername)
	}
	if got := (BuildHTTPSCredential{Username: "oauth2"}).EffectiveUsername(); got != "oauth2" {
		t.Fatalf("EffectiveUsername() = %q, want oauth2", got)
	}
}

func TestValidateBuildHTTPS(t *testing.T) {
	valid := []BuildHTTPSCredential{
		{Scope: "git.example.com", Token: TokenSourceGitCredentials},
		{Scope: "git.example.com", Token: TokenSourceKeyring},
		{Scope: "git.example.com/org", TokenEnv: "ORG_TOKEN"},
		{Scope: "git.example.com", Token: TokenSourceGitCredentials, Username: "oauth2"},
	}
	for _, credential := range valid {
		if err := ValidateBuildHTTPS(credential); err != nil {
			t.Errorf("ValidateBuildHTTPS(%+v) = %v, want nil", credential, err)
		}
	}
	invalid := map[string]BuildHTTPSCredential{
		"bad scope":                {Scope: "Git.Example.com", Token: TokenSourceGitCredentials},
		"no scope":                 {Token: TokenSourceGitCredentials},
		"nothing chosen":           {Scope: "git.example.com"},
		"username only":            {Scope: "git.example.com", Username: "oauth2"},
		"both token and token_env": {Scope: "git.example.com", Token: TokenSourceGitCredentials, TokenEnv: "X"},
		"literal secret":           {Scope: "git.example.com", Token: "ghp_abc123secret"},
		"bad token_env":            {Scope: "git.example.com", TokenEnv: "1_TOKEN"},
	}
	for name, credential := range invalid {
		if err := ValidateBuildHTTPS(credential); err == nil {
			t.Errorf("ValidateBuildHTTPS(%s) = nil, want an error", name)
		}
	}
}

// TestValidateBuildHTTPSAcceptsEveryParsedCredential keeps the exported
// validator and the parser from drifting apart: whatever the config accepts
// must also satisfy the check the CLI runs before writing an entry.
func TestValidateBuildHTTPSAcceptsEveryParsedCredential(t *testing.T) {
	for scope, credential := range loadBuildHTTPSConfig(t).BuildHTTPS {
		if err := ValidateBuildHTTPS(credential); err != nil {
			t.Errorf("ValidateBuildHTTPS(%q, %+v) = %v", scope, credential, err)
		}
	}
}

// TestBuildHTTPSIsAManagerKey covers the system-config path: build_https is a
// recognised top-level key, so an enforced config may supply it as a default.
func TestBuildHTTPSIsAManagerKey(t *testing.T) {
	if !managerKeys["build_https"] {
		t.Fatal("build_https is not a manager key")
	}
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", minimal)
	systemPath := writeConfig(t, dir, "system.json", `{"schema_version": 1, "locked": [],
		"build_https": {"git.example.com": {"token": "git-credentials"}}}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.BuildHTTPSFor("git.example.com/relux-works/portals"); !ok {
		t.Fatalf("system default not applied: %+v", cfg.BuildHTTPS)
	}
}

// TestBuildHTTPSIsNotLockable pins the ratified rule that build_https, like
// build_ssh, stays out of the lockable system-configuration keys: it is an
// operator credential selection, not an organization policy an admin config
// can force onto every machine.
func TestBuildHTTPSIsNotLockable(t *testing.T) {
	if LockableKeys["build_https"] {
		t.Fatal("build_https must not be a lockable key")
	}
}

func TestMatchBuildHTTPSOverABareScopeMap(t *testing.T) {
	scopes := map[string]BuildHTTPSCredential{
		"git.example.com":              {Scope: "git.example.com", Token: TokenSourceGitCredentials},
		"git.example.com/relux-works":  {Scope: "git.example.com/relux-works", Token: TokenSourceKeyring},
		"other.example.com/relux-work": {Scope: "other.example.com/relux-work", TokenEnv: "OTHER_TOKEN"},
	}
	credential, matched := MatchBuildHTTPS(scopes, "git.example.com/relux-works/portals")
	if !matched || credential.Scope != "git.example.com/relux-works" {
		t.Fatalf("match = %+v %v, want the longest covering scope", credential, matched)
	}
	if _, matched := MatchBuildHTTPS(scopes, "GIT.EXAMPLE.COM/relux-works"); matched {
		t.Fatal("a non-canonical identity selected a credential")
	}
	if _, matched := MatchBuildHTTPS(nil, "git.example.com/relux-works"); matched {
		t.Fatal("an empty scope map selected a credential")
	}
	// The method and the function are the same rule over the same scopes.
	cfg := &Config{BuildHTTPS: scopes}
	viaConfig, _ := cfg.BuildHTTPSFor("git.example.com/relux-works/portals")
	if viaConfig != credential {
		t.Fatalf("BuildHTTPSFor = %+v, want %+v", viaConfig, credential)
	}
}
