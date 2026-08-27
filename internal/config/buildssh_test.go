package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// buildSSHConfig carries all three authentication shapes at three nesting
// depths, so one fixture exercises parsing, matching, and serialization.
const buildSSHConfig = `{
	"schema_version": 1, "skills_root": "/tmp/skills", "projects": {},
	"build_ssh": {
		"git.example.com": {"agent": true},
		"git.example.com/relux-works": {"identity": "~/.ssh/org", "known_hosts": "~/.ssh/known_hosts_org"},
		"git.example.com/relux-works/portals": {"agent": "/run/portals.sock", "identity": "/keys/portals"}
	}
}`

func loadBuildSSHConfig(t *testing.T) *Config {
	t.Helper()
	cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", buildSSHConfig), nil)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestParseBuildSSH(t *testing.T) {
	cfg := loadBuildSSHConfig(t)
	want := map[string]BuildSSHCredential{
		"git.example.com": {
			Scope: "git.example.com", Agent: true,
		},
		"git.example.com/relux-works": {
			Scope:    "git.example.com/relux-works",
			Identity: "~/.ssh/org", KnownHosts: "~/.ssh/known_hosts_org",
		},
		"git.example.com/relux-works/portals": {
			Scope: "git.example.com/relux-works/portals", Agent: true,
			AgentSocket: "/run/portals.sock", Identity: "/keys/portals",
		},
	}
	if !reflect.DeepEqual(cfg.BuildSSH, want) {
		t.Fatalf("build_ssh = %+v, want %+v", cfg.BuildSSH, want)
	}
	wantScopes := []string{
		"git.example.com", "git.example.com/relux-works",
		"git.example.com/relux-works/portals",
	}
	if got := cfg.BuildSSHScopes(); !reflect.DeepEqual(got, wantScopes) {
		t.Fatalf("scopes = %v, want %v", got, wantScopes)
	}
}

func TestParseBuildSSHAbsentAndEmpty(t *testing.T) {
	for name, text := range map[string]string{
		"absent": minimal,
		"null":   `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_ssh": null}`,
		"empty":  `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_ssh": {}}`,
	} {
		t.Run(name, func(t *testing.T) {
			cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(cfg.BuildSSH) != 0 {
				t.Fatalf("build_ssh = %+v, want empty", cfg.BuildSSH)
			}
			if len(cfg.BuildSSHScopes()) != 0 {
				t.Fatalf("scopes = %v, want none", cfg.BuildSSHScopes())
			}
			if _, ok := cfg.BuildSSHFor("git.example.com/relux-works/portals"); ok {
				t.Fatal("an unset build_ssh must select no credential")
			}
		})
	}
}

// TestBuildSSHSerializationRoundtrip writes the parsed credentials back out
// through BuildSSHObject and requires the reparsed config to be identical.
func TestBuildSSHSerializationRoundtrip(t *testing.T) {
	dir := t.TempDir()
	first := loadBuildSSHConfig(t)
	payload, err := json.Marshal(map[string]any{
		"schema_version": 1, "skills_root": "/tmp/skills", "projects": map[string]any{},
		"build_ssh": BuildSSHObject(first.BuildSSH),
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
		t.Fatalf("serialized build_ssh did not parse: %v", err)
	}
	if !reflect.DeepEqual(second.BuildSSH, first.BuildSSH) {
		t.Fatalf("roundtrip = %+v, want %+v", second.BuildSSH, first.BuildSSH)
	}
}

func TestBuildSSHObjectOmitsUnsetFields(t *testing.T) {
	for name, tc := range map[string]struct {
		credential BuildSSHCredential
		want       string
	}{
		"agent only": {
			BuildSSHCredential{Scope: "git.example.com", Agent: true},
			`{"git.example.com":{"agent":true}}`,
		},
		"pinned socket": {
			BuildSSHCredential{Scope: "git.example.com", Agent: true, AgentSocket: "/run/a.sock", Identity: "/keys/id"},
			`{"git.example.com":{"agent":"/run/a.sock","identity":"/keys/id"}}`,
		},
		"identity only": {
			BuildSSHCredential{Scope: "git.example.com", Identity: "~/.ssh/id", KnownHosts: "~/.ssh/kh"},
			`{"git.example.com":{"identity":"~/.ssh/id","known_hosts":"~/.ssh/kh"}}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			payload, err := json.Marshal(BuildSSHObject(map[string]BuildSSHCredential{
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
	if got := BuildSSHObject(nil); len(got) != 0 {
		t.Fatalf("BuildSSHObject(nil) = %v, want empty", got)
	}
}

func TestValidBuildSSHScope(t *testing.T) {
	cases := []struct {
		scope string
		want  bool
	}{
		{"git.example.com", true},
		{"git.example.com/relux-works", true},
		{"git.example.com/relux-works/portals", true},
		{"git.example.com/a.b/c_d/e-f", true},
		{"localhost", true},
		{"127.0.0.1/repo", true},
		// Spec §6.3 lowercases the host and preserves repository path case.
		{"git.example.com/RELUX", true},
		{"", false},
		{"Git.Example.com", false},
		{"git.example.com//portals", false},
		{"git.example.com/", false},
		{"/git.example.com", false},
		{"git.example.com/.", false},
		{"git.example.com/..", false},
		{"git.example.com/a b", false},
		{`git.example.com/a\b`, false},
		{"git.example.com/a:b", false},
		{"git.example.com/ünicode", false},
		{"git..example.com", false},
		{"-git.example.com", false},
		{"git.example.com./repo", false},
		{"git.example.com:22/repo", false},
		{"ssh://git.example.com/repo", false},
		{"git@example.com/repo", false},
		{strings.Repeat("a", 253), true},
		{strings.Repeat("a", 254), false},
		{"git.example.com/" + strings.Repeat("a", 4096), false},
	}
	for _, tc := range cases {
		if got := ValidBuildSSHScope(tc.scope); got != tc.want {
			t.Errorf("ValidBuildSSHScope(%q) = %v, want %v", tc.scope, got, tc.want)
		}
	}
}

func TestParseBuildSSHRejections(t *testing.T) {
	const prefix = `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_ssh": `
	cases := []struct {
		name string
		body string
		want string
	}{
		{"not an object", `[]`, "build_ssh: must be an object"},
		{"entry not an object", `{"git.example.com": "agent"}`, "build_ssh.git.example.com: must be an object"},
		{"uppercase host", `{"Git.Example.com": {"agent": true}}`, `scope "Git.Example.com"`},
		{"empty segment", `{"git.example.com//portals": {"agent": true}}`, `scope "git.example.com//portals"`},
		{"trailing slash", `{"git.example.com/portals/": {"agent": true}}`, `scope "git.example.com/portals/"`},
		{"dot segment", `{"git.example.com/./portals": {"agent": true}}`, `scope "git.example.com/./portals"`},
		{"dotdot segment", `{"git.example.com/../portals": {"agent": true}}`, `scope "git.example.com/../portals"`},
		{"scheme in scope", `{"ssh://git.example.com/repo": {"agent": true}}`, "scope"},
		{"port in scope", `{"git.example.com:22/repo": {"agent": true}}`, "scope"},
		{"empty scope", `{"": {"agent": true}}`, `scope ""`},
		{"unknown entry field", `{"git.example.com": {"agent": true, "port": 22}}`, `unsupported field(s): port`},
		{"unknown entry fields sorted", `{"git.example.com": {"agent": true, "user": "x", "port": 22}}`, `unsupported field(s): port, user`},
		{"no selection", `{"git.example.com": {"known_hosts": "/k"}}`, "requires 'agent', 'identity', or both"},
		{"empty entry", `{"git.example.com": {}}`, "requires 'agent', 'identity', or both"},
		{"agent false", `{"git.example.com": {"agent": false}}`, "build_ssh.git.example.com.agent"},
		{"agent false with identity", `{"git.example.com": {"agent": false, "identity": "/keys/id"}}`, "build_ssh.git.example.com.agent"},
		{"agent null", `{"git.example.com": {"agent": null}}`, "build_ssh.git.example.com.agent"},
		{"agent number", `{"git.example.com": {"agent": 22}}`, "build_ssh.git.example.com.agent"},
		{"agent empty socket", `{"git.example.com": {"agent": ""}}`, "build_ssh.git.example.com.agent"},
		{"agent relative socket", `{"git.example.com": {"agent": "run/agent.sock"}}`, "build_ssh.git.example.com.agent"},
		{"identity empty", `{"git.example.com": {"identity": ""}}`, "build_ssh.git.example.com.identity"},
		{"identity null", `{"git.example.com": {"identity": null}}`, "build_ssh.git.example.com.identity"},
		{"identity relative", `{"git.example.com": {"identity": ".ssh/id"}}`, "build_ssh.git.example.com.identity"},
		{"identity bare tilde", `{"git.example.com": {"identity": "~other/id"}}`, "build_ssh.git.example.com.identity"},
		{"identity padded", `{"git.example.com": {"identity": " /keys/id"}}`, "build_ssh.git.example.com.identity"},
		{"identity control byte", `{"git.example.com": {"identity": "/keys/\u0001id"}}`, "build_ssh.git.example.com.identity"},
		{"identity number", `{"git.example.com": {"identity": 1}}`, "build_ssh.git.example.com.identity"},
		{"known_hosts empty", `{"git.example.com": {"agent": true, "known_hosts": ""}}`, "build_ssh.git.example.com.known_hosts"},
		{"known_hosts null", `{"git.example.com": {"agent": true, "known_hosts": null}}`, "build_ssh.git.example.com.known_hosts"},
		{"known_hosts relative", `{"git.example.com": {"agent": true, "known_hosts": "known_hosts"}}`, "build_ssh.git.example.com.known_hosts"},
		{"known_hosts control byte", `{"git.example.com": {"agent": true, "known_hosts": "/etc/\u0007kh"}}`, "build_ssh.git.example.com.known_hosts"},
		{"known_hosts number", `{"git.example.com": {"agent": true, "known_hosts": 1}}`, "build_ssh.git.example.com.known_hosts"},
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

// TestParseBuildSSHReportsOneFaultDeterministically pins the scope ordering,
// so a config with several faults never reports a different one per run.
func TestParseBuildSSHReportsOneFaultDeterministically(t *testing.T) {
	const text = `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_ssh": {
		"z.example.com": {},
		"a.example.com": {}
	}}`
	for range 8 {
		_, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
		if err == nil || !strings.Contains(err.Error(), "build_ssh.a.example.com") {
			t.Fatalf("err = %v, want the lowest scope reported first", err)
		}
	}
}

func TestBuildSSHAcceptsWindowsPaths(t *testing.T) {
	const text = `{"schema_version": 1, "skills_root": "x", "projects": {}, "build_ssh": {
		"git.example.com": {"agent": "\\\\.\\pipe\\openssh-ssh-agent", "identity": "C:/keys/id"}
	}}`
	cfg, err := Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
	if err != nil {
		t.Fatal(err)
	}
	credential := cfg.BuildSSH["git.example.com"]
	if credential.AgentSocket != `\\.\pipe\openssh-ssh-agent` || credential.Identity != "C:/keys/id" {
		t.Fatalf("credential = %+v", credential)
	}
}

func TestBuildSSHLongestPrefixMatch(t *testing.T) {
	cfg := loadBuildSSHConfig(t)
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
		{"ssh://git.example.com/relux-works/portals", ""},
		{"git@git.example.com:relux-works/portals", ""},
		{"git.example.com/relux-works/portals.git", "git.example.com/relux-works"},
	}
	for _, tc := range cases {
		got, ok := cfg.BuildSSHFor(tc.identity)
		if tc.want == "" {
			if ok {
				t.Errorf("BuildSSHFor(%q) = %+v, want no match", tc.identity, got)
			}
			continue
		}
		if !ok || got.Scope != tc.want {
			t.Errorf("BuildSSHFor(%q) = %+v, %v; want scope %q", tc.identity, got, ok, tc.want)
		}
	}
}

func TestBuildSSHMatchCarriesCredential(t *testing.T) {
	cfg := loadBuildSSHConfig(t)
	credential, ok := cfg.BuildSSHFor("git.example.com/relux-works/portals/deep")
	if !ok || !credential.Agent || credential.AgentSocket != "/run/portals.sock" ||
		credential.Identity != "/keys/portals" || credential.KnownHosts != "" {
		t.Fatalf("match = %+v, %v", credential, ok)
	}
	organization, ok := cfg.BuildSSHFor("git.example.com/relux-works/other")
	if !ok || organization.Agent || organization.Identity != "~/.ssh/org" ||
		organization.KnownHosts != "~/.ssh/known_hosts_org" {
		t.Fatalf("organization match = %+v, %v", organization, ok)
	}
}

func TestBuildSSHExpandedResolvesHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	credential := BuildSSHCredential{
		Scope: "git.example.com", Agent: true, AgentSocket: "~/.ssh/agent.sock",
		Identity: "~/.ssh/org", KnownHosts: "~/.ssh/known_hosts_org",
	}
	expanded := credential.Expanded()
	for field, got := range map[string]string{
		"agent": expanded.AgentSocket, "identity": expanded.Identity,
		"known_hosts": expanded.KnownHosts,
	} {
		if !strings.HasPrefix(got, filepath.Join(home, ".ssh")+string(filepath.Separator)) {
			t.Errorf("%s = %q, want a path under %s", field, got, home)
		}
	}
	if !expanded.Agent || expanded.Scope != "git.example.com" {
		t.Errorf("expanded = %+v", expanded)
	}
	absolute := BuildSSHCredential{Identity: "/keys/id", KnownHosts: "/etc/known_hosts"}
	if got := absolute.Expanded(); got != absolute {
		t.Errorf("Expanded(%+v) = %+v, want unchanged", absolute, got)
	}
}

func TestValidateBuildSSH(t *testing.T) {
	valid := []BuildSSHCredential{
		{Scope: "git.example.com", Agent: true},
		{Scope: "git.example.com", Identity: "~/.ssh/id"},
		{Scope: "git.example.com/org", Agent: true, AgentSocket: "/run/a.sock", Identity: "/keys/id"},
		{Scope: "git.example.com", Agent: true, KnownHosts: "/etc/known_hosts"},
	}
	for _, credential := range valid {
		if err := ValidateBuildSSH(credential); err != nil {
			t.Errorf("ValidateBuildSSH(%+v) = %v, want nil", credential, err)
		}
	}
	invalid := map[string]BuildSSHCredential{
		"bad scope":            {Scope: "Git.Example.com", Agent: true},
		"no scope":             {Agent: true},
		"nothing chosen":       {Scope: "git.example.com"},
		"known_hosts only":     {Scope: "git.example.com", KnownHosts: "/etc/known_hosts"},
		"socket without agent": {Scope: "git.example.com", AgentSocket: "/run/a.sock", Identity: "/keys/id"},
		"relative identity":    {Scope: "git.example.com", Identity: "keys/id"},
		"relative known_hosts": {Scope: "git.example.com", Agent: true, KnownHosts: "kh"},
	}
	for name, credential := range invalid {
		if err := ValidateBuildSSH(credential); err == nil {
			t.Errorf("ValidateBuildSSH(%s) = nil, want an error", name)
		}
	}
}

// TestValidateBuildSSHAcceptsEveryParsedCredential keeps the exported
// validator and the parser from drifting apart: whatever the config accepts
// must also satisfy the check the CLI runs before writing an entry.
func TestValidateBuildSSHAcceptsEveryParsedCredential(t *testing.T) {
	for scope, credential := range loadBuildSSHConfig(t).BuildSSH {
		if err := ValidateBuildSSH(credential); err != nil {
			t.Errorf("ValidateBuildSSH(%q, %+v) = %v", scope, credential, err)
		}
	}
}

// TestBuildSSHIsAManagerKey covers the system-config path: build_ssh is a
// recognised top-level key, so an enforced config may supply it as a default.
func TestBuildSSHIsAManagerKey(t *testing.T) {
	if !managerKeys["build_ssh"] {
		t.Fatal("build_ssh is not a manager key")
	}
	dir := t.TempDir()
	userPath := writeConfig(t, dir, "config.json", minimal)
	systemPath := writeConfig(t, dir, "system.json", `{"schema_version": 1, "locked": [],
		"build_ssh": {"git.example.com": {"agent": true}}}`)
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	cfg, err := Load(userPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.BuildSSHFor("git.example.com/relux-works/portals"); !ok {
		t.Fatalf("system default not applied: %+v", cfg.BuildSSH)
	}
}

func TestMatchBuildSSHOverABareScopeMap(t *testing.T) {
	scopes := map[string]BuildSSHCredential{
		"git.example.com":              {Scope: "git.example.com", Agent: true},
		"git.example.com/relux-works":  {Scope: "git.example.com/relux-works", Identity: "/operator/group"},
		"other.example.com/relux-work": {Scope: "other.example.com/relux-work", Identity: "/operator/other"},
	}
	credential, matched := MatchBuildSSH(scopes, "git.example.com/relux-works/portals")
	if !matched || credential.Scope != "git.example.com/relux-works" {
		t.Fatalf("match = %+v %v, want the longest covering scope", credential, matched)
	}
	if _, matched := MatchBuildSSH(scopes, "GIT.EXAMPLE.COM/relux-works"); matched {
		t.Fatal("a non-canonical identity selected a credential")
	}
	if _, matched := MatchBuildSSH(nil, "git.example.com/relux-works"); matched {
		t.Fatal("an empty scope map selected a credential")
	}
	// The method and the function are the same rule over the same scopes.
	cfg := &Config{BuildSSH: scopes}
	viaConfig, _ := cfg.BuildSSHFor("git.example.com/relux-works/portals")
	if viaConfig != credential {
		t.Fatalf("BuildSSHFor = %+v, want %+v", viaConfig, credential)
	}
}
