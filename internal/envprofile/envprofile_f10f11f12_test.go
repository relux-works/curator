package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production entry points under test: Install, List, EnsureDefault,
// EnsureDefaultWithPolicy.
//
// F10 covers the builtin default migration's machine gates (§9.1): the
// migration must enforce the effective machine configuration — the system
// overlay with its locked keys and the CURATOR_CONFIG override — never a
// weaker re-parse. These tests drive the bare entry points (which resolve
// the policy through loadMachinePolicy) with the process environment pinned
// at test files. A mutant that reads the user file but drops the system
// overlay, or that maps a read/parse failure to an empty policy, must fail
// them.

const permissiveUserConfig = `{"schema_version":1,"skills_root":"skills","projects":{}}` + "\n"

func writeTestConfig(t *testing.T, path, payload string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
}

// migrationHomeWithSkill declares one tag-pinned global skill under a fake
// network identity served from a local repository, and returns the home and
// the canonical identity.
func migrationHomeWithSkill(t *testing.T, ids *gitIdentities, identity string) (string, string) {
	t.Helper()
	skill := gitRepo(t, map[string]string{"README.md": "skill\n"}, "v1.0.0")
	operand := ids.serve(skill, "https://"+identity)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillfile := `{"schema_version": 1, "skills": [{"name": "hello", "git": "` + operand + `", "tag": "v1.0.0"}]}` + "\n"
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.json"), []byte(skillfile), 0o644); err != nil {
		t.Fatal(err)
	}
	return home, identity
}

// TestMigrationHonoursSystemLockedAllowlist checks F10 through EnsureDefault:
// a declared global skill outside the system-locked allowed_sources is
// refused, even though the user configuration permits everything. A mutant
// that loads the user file but drops the system overlay admits the skill
// and must fail this test.
func TestMigrationHonoursSystemLockedAllowlist(t *testing.T) {
	ids := newGitIdentities(t)
	home, _ := migrationHomeWithSkill(t, ids, "example.com/skills/hello")
	writeTestConfig(t, filepath.Join(home, "config.json"), permissiveUserConfig)
	t.Setenv("CURATOR_CONFIG", filepath.Join(home, "config.json"))
	system := filepath.Join(t.TempDir(), "system.json")
	writeTestConfig(t, system, `{"schema_version":1,"locked":["allowed_sources","audit"],`+
		`"allowed_sources":["github.com/relux-works"],`+
		`"audit":{"enabled":true,"mode":"strict","revocations":[]}}`+"\n")
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	err := EnsureDefault(home)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("migration err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "allowed sources") {
		t.Fatalf("migration err = %v, want the allowlist reason", err)
	}
	if _, err := readSource(home, DefaultProfile); err == nil {
		t.Fatal("refused migration must leave no default profile")
	}
	if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
		t.Fatalf("refused migration cloned: %v", entries)
	}
}

// TestMigrationHonoursSystemLockedRevocation checks F10 through EnsureDefault:
// a declared global skill matching a system-locked audit revocation is
// refused, even though the allowlist permits its identity.
func TestMigrationHonoursSystemLockedRevocation(t *testing.T) {
	ids := newGitIdentities(t)
	home, identity := migrationHomeWithSkill(t, ids, "example.com/skills/hello")
	writeTestConfig(t, filepath.Join(home, "config.json"), permissiveUserConfig)
	t.Setenv("CURATOR_CONFIG", filepath.Join(home, "config.json"))
	system := filepath.Join(t.TempDir(), "system.json")
	writeTestConfig(t, system, `{"schema_version":1,"locked":["allowed_sources","audit"],`+
		`"allowed_sources":["example.com"],`+
		`"audit":{"enabled":true,"mode":"strict","revocations":["source:`+identity+`"]}}`+"\n")
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	err := EnsureDefault(home)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("migration err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("migration err = %v, want the revocation reason", err)
	}
}

// TestMigrationSucceedsWhenSystemPolicyPermits is the mirror case: the same
// migration under a system configuration that permits the skill migrates it
// with its canonical identity.
func TestMigrationSucceedsWhenSystemPolicyPermits(t *testing.T) {
	ids := newGitIdentities(t)
	home, identity := migrationHomeWithSkill(t, ids, "example.com/skills/hello")
	writeTestConfig(t, filepath.Join(home, "config.json"), permissiveUserConfig)
	t.Setenv("CURATOR_CONFIG", filepath.Join(home, "config.json"))
	system := filepath.Join(t.TempDir(), "system.json")
	writeTestConfig(t, system, `{"schema_version":1,"locked":["allowed_sources","audit"],`+
		`"allowed_sources":["example.com"],`+
		`"audit":{"enabled":true,"mode":"strict","revocations":[]}}`+"\n")
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	if err := EnsureDefault(home); err != nil {
		t.Fatalf("permitted migration must succeed: %v", err)
	}
	lock, _, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, member := range lock.Members {
		if member.Kind == "skill" && member.Name == "hello" && member.Source == identity {
			found = true
		}
	}
	if !found {
		t.Fatalf("migrated lock %+v", lock.Members)
	}
}

// TestMigrationHonoursConfigPathOverride checks F10 through List with the
// operator configuration at a non-config.json file name: the gates in
// CURATOR_CONFIG apply even though <home>/config.json does not exist. A
// mutant that reads <home>/config.json directly sees no file, maps the
// absence to an empty policy, and must fail this test.
func TestMigrationHonoursConfigPathOverride(t *testing.T) {
	ids := newGitIdentities(t)
	home, _ := migrationHomeWithSkill(t, ids, "example.com/skills/hello")
	writeTestConfig(t, filepath.Join(home, "curator.json"),
		`{"schema_version":1,"skills_root":"skills","projects":{},`+
			`"allowed_sources":["github.com/relux-works"]}`+"\n")
	t.Setenv("CURATOR_CONFIG", filepath.Join(home, "curator.json"))
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	_, err := List(home)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("list err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "allowed sources") {
		t.Fatalf("list err = %v, want the allowlist reason", err)
	}
}

// TestMachinePolicyLoadFailureFailsMigration pins the absence-vs-unreadable
// rule: a malformed machine configuration fails the migration — a failed
// read is never an empty policy. A mutant that maps the parse failure to an
// empty policy migrates the skill and must fail this test.
func TestMachinePolicyLoadFailureFailsMigration(t *testing.T) {
	ids := newGitIdentities(t)
	home, _ := migrationHomeWithSkill(t, ids, "example.com/skills/hello")
	writeTestConfig(t, filepath.Join(home, "config.json"), `{"schema_version":`)
	t.Setenv("CURATOR_CONFIG", filepath.Join(home, "config.json"))
	t.Setenv("CURATOR_SYSTEM_CONFIG", "")
	err := EnsureDefault(home)
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("migration err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "machine configuration") {
		t.Fatalf("migration err = %v, want the configuration reason", err)
	}
}

// TestSourceAllowlistRejectsAdjacentPathSegment narrows the F9 allowlist
// gate (review F11): an allowlist entry example.com/org must not admit
// example.com/org-evil/pkg — same host, adjacent path segment. A mutant that
// matches on the host only admits the operand (the clone succeeds through
// the fixture) and must fail this test.
func TestSourceAllowlistRejectsAdjacentPathSegment(t *testing.T) {
	repo := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "evil", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"context/a.md": "hello\n",
	}, "v1.0.0")
	ids := newGitIdentities(t)
	operand := ids.serve(repo, "https://example.com/org-evil/pkg")
	home := t.TempDir()
	policy := Policy{AllowedSources: []string{"example.com/org"}}
	_, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("allowlist err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "allowed sources") {
		t.Fatalf("allowlist err = %v, want the allowlist reason", err)
	}
	if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
		t.Fatalf("blocked install cloned: %v", entries)
	}
}

// TestRevokedSkillMemberIsRefused narrows the F9 revocation gate (review
// F11): revocation applies to skill members, not only to context members. A
// mutant that revokes only KindContext installs the skill and must fail this
// test. Production entry point: Install of a clean root requiring the
// revoked skill.
func TestRevokedSkillMemberIsRefused(t *testing.T) {
	ids := newGitIdentities(t)
	skill := gitRepo(t, map[string]string{"README.md": "skill\n"}, "v1.0.0")
	skillOperand := ids.serve(skill, "https://example.com/evil-skill")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "clean", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"skills": {"bad": {"git": "` + skillOperand + `", "range": "*"}}}}` + "\n",
		"context/a.md": "clean\n",
	}, "v1.0.0")
	home := t.TempDir()
	policy := Policy{Revocations: []string{"source:example.com/evil-skill"}}
	_, _, _, err := Install(home, InstallOptions{
		Operand: ids.serve(root, "https://example.com/clean"), Policy: policy,
	})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("revocation err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revocation err = %v, want the revocation reason", err)
	}
}

// TestRevokedMCPMemberIsRefused narrows the same gate for mcp members: a
// revoked mcp package is refused. Production entry point: Install of a clean
// root requiring the revoked mcp member.
func TestRevokedMCPMemberIsRefused(t *testing.T) {
	ids := newGitIdentities(t)
	mcp := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "sh", "args": []}}` + "\n",
	}, "v1.0.0")
	mcpOperand := ids.serve(mcp, "https://example.com/evil-tool")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "clean", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"mcp": {"tool": {"git": "` + mcpOperand + `", "range": "*"}}}}` + "\n",
		"context/a.md": "clean\n",
	}, "v1.0.0")
	home := t.TempDir()
	policy := Policy{Revocations: []string{"source:example.com/evil-tool"}}
	_, _, _, err := Install(home, InstallOptions{
		Operand: ids.serve(root, "https://example.com/clean"), Policy: policy,
	})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("revocation err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revocation err = %v, want the revocation reason", err)
	}
}

// TestStrictAuditCanaryFailureBlocksInstall proves the canary's blocking role
// on the profile path (review F11, §9.1 "whose failure always blocks"): with
// the canary forced to fail, the strict-audit member path refuses the
// install. A mutant that drops the canary guard installs cleanly and must
// fail this test. Production entry point: Install of a clean path package.
func TestStrictAuditCanaryFailureBlocksInstall(t *testing.T) {
	previous := canaryPasses
	canaryPasses = func() bool { return false }
	defer func() { canaryPasses = previous }()
	home := t.TempDir()
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", "hello\n")
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("canary err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "canary") {
		t.Fatalf("canary err = %v, want the canary reason", err)
	}
	if _, err := readSource(home, "acme"); err == nil {
		t.Fatal("canary-blocked install must leave no profile record")
	}
}

// TestCanonicalGitRejectsFileRemote pins the F12 operand boundary itself:
// canonicalGit rejects every file: spelling, so no file:// remote enters
// the install, requirement, or migration paths as an identity. A mutant that
// restores the old pass-through admits the operand here and must fail this
// test (the Install-level refusal would still hold through the gateSource
// backstop, so this unit test is what ties the mutant to its layer).
func TestCanonicalGitRejectsFileRemote(t *testing.T) {
	for _, operand := range []string{
		"file:///tmp/pkg",
		"FILE:///tmp/pkg",
		"file://host/org/pkg",
	} {
		identity, err := canonicalGit(operand)
		if err == nil || identity != "" {
			t.Fatalf("operand %s: canonicalGit = %q, %v, want rejection", operand, identity, err)
		}
		if !strings.Contains(err.Error(), "no network identity") {
			t.Fatalf("operand %s: err = %v, want the identity reason", operand, err)
		}
	}
}

// TestGateSourceRejectsFileRemote pins the F12 clone boundary itself: the
// allowlist gate refuses a file:// source even with an empty allowlist. A
// mutant that restores the old bypass admits it here and must fail this
// test.
func TestGateSourceRejectsFileRemote(t *testing.T) {
	manager := newGitManager(t.TempDir())
	if err := manager.gateSource("file:///tmp/pkg"); err == nil {
		t.Fatal("gateSource must refuse a file:// source")
	} else if !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("gateSource err = %v, want %s", err, DiagSourceInvalid)
	}
}

// TestFileOperandIsRefused checks F12 through Install: a file:// git operand
// is rejected with profile_source_invalid before any clone, in any letter
// case. Such remotes carry no network identity and admit no valid lock or
// marker shape.
func TestFileOperandIsRefused(t *testing.T) {
	for _, operand := range []string{"file:///tmp/never-there-pkg", "FILE:///tmp/never-there-pkg"} {
		home := t.TempDir()
		_, _, _, err := Install(home, InstallOptions{Operand: operand})
		if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
			t.Fatalf("operand %s: err = %v, want %s", operand, err, DiagSourceInvalid)
		}
		if !strings.Contains(err.Error(), "no network identity") {
			t.Fatalf("operand %s: err = %v, want the identity reason", operand, err)
		}
		if _, err := readSource(home, "never-there-pkg"); err == nil {
			t.Fatalf("operand %s: refused install must leave no profile record", operand)
		}
		if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
			t.Fatalf("operand %s: refused install cloned: %v", operand, entries)
		}
	}
}

// TestFileRequirementSourceIsRefused checks F12 for the transitive shape: a
// requires.*.git source with a file:// remote is rejected at the Identity
// boundary with profile_source_invalid, without ever cloning the dependency.
func TestFileRequirementSourceIsRefused(t *testing.T) {
	ids := newGitIdentities(t)
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "clean", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"contexts": {"dep": {"git": "file:///tmp/never-there-dep", "range": "*"}}}}` + "\n",
		"context/a.md": "clean\n",
	}, "v1.0.0")
	home := t.TempDir()
	_, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/clean")})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("requirement err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "no network identity") {
		t.Fatalf("requirement err = %v, want the identity reason", err)
	}
}
