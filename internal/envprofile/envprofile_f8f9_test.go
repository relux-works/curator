package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/identity"
)

// Production entry points under test: Install, Use.

// TestCanonicalIdentityUnifiesSpellings checks F8 through Install: one
// repository installed under an https spelling (with .git and an uppercase
// host), an scp-style spelling, and an ssh:// spelling resolves to one
// canonical identity, one lock_sha256, and one store entry. git insteadOf
// keeps the test offline. The produced lock and marker carry the canonical
// identity and validate.
func TestCanonicalIdentityUnifiesSpellings(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	repo := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "acme", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"context/a.md": "hello\n",
	}, "v1.0.0")
	// Every spelling routes through the shared insteadOf helper, so the
	// fixture remote always takes the git-safe file URL form.
	ids := newGitIdentities(t)
	for _, operand := range []string{
		"https://EXAMPLE.com/acme.git",
		"git@example.com:acme.git",
		"ssh://git@example.com/acme",
	} {
		ids.serve(repo, operand)
	}

	operands := []string{
		"https://EXAMPLE.com/acme.git",
		"git@example.com:acme.git",
		"ssh://git@example.com/acme",
	}
	const want = "example.com/acme"
	var hashes []string
	for index, operand := range operands {
		home := t.TempDir()
		pinHomes(t)
		info, _, _, err := Install(home, InstallOptions{Operand: operand})
		if err != nil {
			t.Fatalf("install %s: %v", operand, err)
		}
		member, ok := info.Lock.RootMember()
		if !ok {
			t.Fatalf("install %s: no root member", operand)
		}
		if member.Source != want {
			t.Fatalf("install %s: lock source %q, want %q", operand, member.Source, want)
		}
		if !identity.ValidCanonical(member.Source) || strings.HasSuffix(member.Source, ".git") {
			t.Fatalf("install %s: lock source %q is not a canonical identity", operand, member.Source)
		}
		if err := info.Lock.Validate(); err != nil {
			t.Fatalf("install %s: lock invalid: %v", operand, err)
		}
		hashes = append(hashes, info.LockHash)
		results, err := Use(home, info.Name, "", "", false)
		if err != nil {
			t.Fatalf("use %s: %v", operand, err)
		}
		for _, result := range results {
			if !result.OK {
				t.Fatalf("use %s: %s %s", operand, result.Adapter, result.Detail)
			}
		}
		marker, err := envmarker.Read(os.Getenv("CLAUDE_CONFIG_DIR"))
		if err != nil || marker == nil {
			t.Fatalf("install %s: marker %v", operand, err)
		}
		if marker.Profile.Source != want {
			t.Fatalf("install %s: marker source %q, want %q", operand, marker.Profile.Source, want)
		}
		if err := marker.Validate(); err != nil {
			t.Fatalf("install %s: marker invalid: %v", operand, err)
		}
		_ = index
	}
	for _, hash := range hashes[1:] {
		if hash != hashes[0] {
			t.Fatalf("lock hashes diverge: %q", hashes)
		}
	}
}

// TestCanonicalIdentityRejectsMalformedSource checks F8 through Install: a
// malformed network source (explicit port) is rejected with
// profile_source_invalid before any clone.
func TestCanonicalIdentityRejectsMalformedSource(t *testing.T) {
	home := t.TempDir()
	_, _, _, err := Install(home, InstallOptions{Operand: "https://example.com:8443/acme"})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("malformed source err = %v, want %s", err, DiagSourceInvalid)
	}
	// Rejected at the identity boundary (explicit port), never reaching the
	// clone: a mutant that passes the operand through must fail this clause.
	if !strings.Contains(err.Error(), "port") {
		t.Fatalf("malformed source err = %v, want the port reason", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos"))
	if readErr == nil && len(entries) != 0 {
		t.Fatalf("malformed source cloned: %v", entries)
	}
}

// TestSourceAllowlistRefusesBeforeClone checks F9 through Install: a git
// member outside allowed_sources is refused with profile_source_invalid
// before any clone (the profile-repos cache stays empty).
func TestSourceAllowlistRefusesBeforeClone(t *testing.T) {
	home := t.TempDir()
	policy := Policy{AllowedSources: []string{"github.com/relux-works"}}
	_, _, _, err := Install(home, InstallOptions{
		Operand: "https://evil.example.com/pkg/root",
		Policy:  policy,
	})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("allowlist err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "allowed sources") {
		t.Fatalf("allowlist err = %v, want the allowlist reason", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos"))
	if readErr == nil && len(entries) != 0 {
		t.Fatalf("blocked install cloned: %v", entries)
	}
}

// TestRevokedSourceIsRefused checks F9 through Install: a member whose
// canonical source matches audit.revocations is refused.
func TestRevokedSourceIsRefused(t *testing.T) {
	repo := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "pwned", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]}}` + "\n",
		"context/a.md": "hello\n",
	}, "v1.0.0")
	ids := newGitIdentities(t)
	operand := ids.serve(repo, "https://example.com/evil/pkg")
	home := t.TempDir()
	policy := Policy{Revocations: []string{"source:example.com/evil/pkg"}}
	_, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("revocation err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revocation err = %v, want the revocation reason", err)
	}
}

// TestMCPAllowlistRefusesOutsidePackage checks F9 through Install with the
// resolution production path: an mcp member outside the MCP package
// allowlist is refused with mcp_package_not_allowed.
func TestMCPAllowlistRefusesOutsidePackage(t *testing.T) {
	ids := newGitIdentities(t)
	mcp := gitRepo(t, map[string]string{
		"agent-mcp.json": `{"schema_version": 1, "name": "tool", "version": "1.0.0",` +
			`"server": {"transport": "stdio", "command": "sh", "args": []}}` + "\n",
	}, "v1.0.0")
	mcpOperand := ids.serve(mcp, "https://example.com/evil-tool")
	root := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "withmcp", "version": "1.0.0",` +
			`"requires": {"mcp": {"tool": {"git": "` + mcpOperand + `", "range": "*"}}}}` + "\n",
	}, "v1.0.0")
	home := t.TempDir()
	policy := Policy{MCPAllowlist: []string{"example.com/allowed"}}
	_, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/withmcp"), Policy: policy})
	if err == nil || !strings.Contains(err.Error(), "mcp_package_not_allowed") {
		t.Fatalf("mcp allowlist err = %v, want mcp_package_not_allowed", err)
	}
}

// TestStrictAuditCanaryPasses checks the F9 canary precondition: the static
// detector canary fires with working detectors. A mutant that empties the
// detectors must fail this test and every Install test (the canary always
// blocks profile installation).
func TestStrictAuditCanaryPasses(t *testing.T) {
	if !audit.CanaryPasses() {
		t.Fatal("audit canary must pass with working detectors")
	}
}

// TestMigratedSkillSourceIsCanonical checks F8 at the migration boundary:
// a tag-pinned global skill migrates with its canonical source identity,
// never the raw declared URL — the old trim-space/trim-slash normalizer
// passes this test only by accident of the assertion, so the assertion is
// on the canonical identity and the lock validates against
// context-lock-v1.
func TestMigratedSkillSourceIsCanonical(t *testing.T) {
	skill := gitRepo(t, map[string]string{"README.md": "skill\n"}, "v1.0.0")
	ids := newGitIdentities(t)
	operand := ids.serve(skill, "https://EXAMPLE.com/skills/hello.git")
	home := t.TempDir()
	pinHomes(t)
	if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillfile := `{"schema_version": 1, "skills": [{"name": "sk", "git": "` + operand + `", "tag": "v1.0.0"}]}` + "\n"
	if err := os.WriteFile(filepath.Join(home, "global", "Skillfile.json"), []byte(skillfile), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(home, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	var member *contextlock.Member
	for index := range lock.Members {
		if lock.Members[index].Kind == contextlock.KindSkill {
			member = &lock.Members[index]
		}
	}
	if member == nil {
		t.Fatalf("migrated lock %+v", lock.Members)
	}
	if member.Source != "example.com/skills/hello" {
		t.Fatalf("migrated skill source %q, want the canonical identity", member.Source)
	}
	if !identity.ValidCanonical(member.Source) {
		t.Fatalf("migrated skill source %q is not a canonical identity", member.Source)
	}
	if err := lock.Validate(); err != nil {
		t.Fatalf("migrated lock invalid: %v", err)
	}
}
