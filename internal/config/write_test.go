package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestBootstrapAndAddProject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "ru", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := Bootstrap(path, "/other", "", nil, false); err == nil {
		t.Fatal("bootstrap must preserve an existing config without --force")
	}
	project := t.TempDir()
	if err := AddProject(path, "app", project, []string{"claude_code"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PreferredLocale != "ru" || cfg.Projects["app"].Path != project {
		t.Fatalf("config = %+v", cfg)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("config mode = %o, want private", info.Mode().Perm())
	}
}

func TestSetBuildSSHRecordsReplacesAndPreservesOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "ru", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	if err := AddProject(path, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	replaced, err := SetBuildSSH(path, BuildSSHCredential{
		Scope: "git.example.com", Agent: true, AgentSocket: "/run/agent.sock",
	})
	if err != nil || replaced {
		t.Fatalf("first set: replaced=%v err=%v", replaced, err)
	}
	if replaced, err = SetBuildSSH(path, BuildSSHCredential{
		Scope: "git.example.com/portals", Identity: "~/.ssh/portals", KnownHosts: "~/.ssh/known_hosts",
	}); err != nil || replaced {
		t.Fatalf("second set: replaced=%v err=%v", replaced, err)
	}
	// A replacement is a whole entry: a leftover agent selection from the
	// previous spelling would still authenticate.
	if replaced, err = SetBuildSSH(path, BuildSSHCredential{
		Scope: "git.example.com", Identity: "/keys/org",
	}); err != nil || !replaced {
		t.Fatalf("replacement: replaced=%v err=%v", replaced, err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]BuildSSHCredential{
		"git.example.com": {Scope: "git.example.com", Identity: "/keys/org"},
		"git.example.com/portals": {
			Scope: "git.example.com/portals", Identity: "~/.ssh/portals",
			KnownHosts: "~/.ssh/known_hosts",
		},
	}
	if !reflect.DeepEqual(cfg.BuildSSH, want) {
		t.Fatalf("build_ssh = %+v, want %+v", cfg.BuildSSH, want)
	}
	if cfg.PreferredLocale != "ru" || cfg.Projects["app"].Path != project {
		t.Fatalf("unrelated fields disturbed: %+v", cfg)
	}
}

func TestSetBuildSSHRejectsInvalidCredentialsWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "", nil, false); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for name, credential := range map[string]BuildSSHCredential{
		"malformed scope":  {Scope: "Git.Example.com", Agent: true},
		"selects nothing":  {Scope: "git.example.com"},
		"known hosts only": {Scope: "git.example.com", KnownHosts: "/etc/known_hosts"},
		"relative path":    {Scope: "git.example.com", Identity: "keys/org"},
		"socket no agent":  {Scope: "git.example.com", AgentSocket: "/run/agent.sock", Identity: "/keys/org"},
	} {
		if _, err := SetBuildSSH(path, credential); err == nil {
			t.Fatalf("%s: SetBuildSSH accepted %+v", name, credential)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a rejected credential rewrote the config:\n%s", after)
	}
}

func TestRemoveBuildSSHDropsScopesAndReportsMissingOnes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "", nil, false); err != nil {
		t.Fatal(err)
	}
	if err := RemoveBuildSSH(path, "git.example.com"); err == nil {
		t.Fatal("removing from a config with no build_ssh must fail")
	}
	for _, scope := range []string{"git.example.com", "git.example.com/portals"} {
		if _, err := SetBuildSSH(path, BuildSSHCredential{Scope: scope, Agent: true}); err != nil {
			t.Fatal(err)
		}
	}
	if err := RemoveBuildSSH(path, "git.example.com/other"); err == nil {
		t.Fatal("removing an unconfigured scope must fail")
	}
	if err := RemoveBuildSSH(path, "git.example.com"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.BuildSSH) != 1 || !cfg.BuildSSH["git.example.com/portals"].Agent {
		t.Fatalf("build_ssh = %+v", cfg.BuildSSH)
	}
	// The last scope takes the field with it, so nothing reads as "an empty
	// credential set was configured on purpose".
	if err := RemoveBuildSSH(path, "git.example.com/portals"); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if _, present := object["build_ssh"]; present {
		t.Fatalf("build_ssh survived its last scope: %s", payload)
	}
	if err := RemoveBuildSSH(path, "git.example.com/portals"); err == nil {
		t.Fatal("a second removal must fail")
	}
}

func TestSetBuildHTTPSRecordsReplacesAndPreservesOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "ru", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	if err := AddProject(path, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	replaced, err := SetBuildHTTPS(path, BuildHTTPSCredential{
		Scope: "git.example.com", Token: TokenSourceGitCredentials,
	})
	if err != nil || replaced {
		t.Fatalf("first set: replaced=%v err=%v", replaced, err)
	}
	if replaced, err = SetBuildHTTPS(path, BuildHTTPSCredential{
		Scope: "git.example.com/portals", TokenEnv: "PORTALS_TOKEN",
	}); err != nil || replaced {
		t.Fatalf("second set: replaced=%v err=%v", replaced, err)
	}
	// A replacement is a whole entry: a leftover username from the previous
	// spelling would still be sent alongside the new token.
	if replaced, err = SetBuildHTTPS(path, BuildHTTPSCredential{
		Scope: "git.example.com", Token: TokenSourceKeyring, Username: "oauth2",
	}); err != nil || !replaced {
		t.Fatalf("replacement: replaced=%v err=%v", replaced, err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]BuildHTTPSCredential{
		"git.example.com": {Scope: "git.example.com", Token: TokenSourceKeyring, Username: "oauth2"},
		"git.example.com/portals": {
			Scope: "git.example.com/portals", TokenEnv: "PORTALS_TOKEN",
		},
	}
	if !reflect.DeepEqual(cfg.BuildHTTPS, want) {
		t.Fatalf("build_https = %+v, want %+v", cfg.BuildHTTPS, want)
	}
	if cfg.PreferredLocale != "ru" || cfg.Projects["app"].Path != project {
		t.Fatalf("unrelated fields disturbed: %+v", cfg)
	}
}

func TestSetBuildHTTPSRejectsInvalidCredentialsWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "", nil, false); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for name, credential := range map[string]BuildHTTPSCredential{
		"malformed scope":    {Scope: "Git.Example.com", Token: TokenSourceGitCredentials},
		"selects nothing":    {Scope: "git.example.com"},
		"both token and env": {Scope: "git.example.com", Token: TokenSourceGitCredentials, TokenEnv: "X"},
		"literal secret":     {Scope: "git.example.com", Token: "ghp_abc123secret"},
		"invalid token_env":  {Scope: "git.example.com", TokenEnv: "1_TOKEN"},
	} {
		if _, err := SetBuildHTTPS(path, credential); err == nil {
			t.Fatalf("%s: SetBuildHTTPS accepted %+v", name, credential)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("a rejected credential rewrote the config:\n%s", after)
	}
}

func TestRemoveBuildHTTPSDropsScopesAndReportsMissingOnes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home", "config.json")
	if err := Bootstrap(path, filepath.Join(t.TempDir(), "skills"), "", nil, false); err != nil {
		t.Fatal(err)
	}
	if err := RemoveBuildHTTPS(path, "git.example.com"); err == nil {
		t.Fatal("removing from a config with no build_https must fail")
	}
	for _, scope := range []string{"git.example.com", "git.example.com/portals"} {
		if _, err := SetBuildHTTPS(path, BuildHTTPSCredential{Scope: scope, Token: TokenSourceKeyring}); err != nil {
			t.Fatal(err)
		}
	}
	if err := RemoveBuildHTTPS(path, "git.example.com/other"); err == nil {
		t.Fatal("removing an unconfigured scope must fail")
	}
	if err := RemoveBuildHTTPS(path, "git.example.com"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.BuildHTTPS) != 1 || cfg.BuildHTTPS["git.example.com/portals"].Token != TokenSourceKeyring {
		t.Fatalf("build_https = %+v", cfg.BuildHTTPS)
	}
	// The last scope takes the field with it, so nothing reads as "an empty
	// credential set was configured on purpose".
	if err := RemoveBuildHTTPS(path, "git.example.com/portals"); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if _, present := object["build_https"]; present {
		t.Fatalf("build_https survived its last scope: %s", payload)
	}
	if err := RemoveBuildHTTPS(path, "git.example.com/portals"); err == nil {
		t.Fatal("a second removal must fail")
	}
}
