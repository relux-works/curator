package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

func makeStateFileUnreadable(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("this environment can read a mode-000 file: chmod refused: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
}

func makeStateDirectoryUnreadable(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("this environment can read a mode-000 directory: chmod refused: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
	if _, err := os.ReadDir(path); err == nil {
		t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
	}
}

func TestCurrentPointerAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	if current, err := Current(home); err != nil || current != "" {
		t.Fatalf("absent current pointer = (%q, %v), want empty, nil", current, err)
	}
	path := CurrentFile(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	if _, err := Current(home); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("unreadable current pointer error = %v, want typed unreadable diagnostic for %s", err, path)
	}
}

func TestScopedCurrentAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	if current, err := ScopedCurrents(home); err != nil || len(current) != 0 {
		t.Fatalf("absent scoped-current directory = (%v, %v), want empty, nil", current, err)
	}
	path := filepath.Join(ScopedDir(home), scopeFileName("claude_code"))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("default\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	if _, err := ScopedCurrents(home); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("unreadable scoped-current error = %v, want typed unreadable diagnostic for %s", err, path)
	}
}

func TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent(t *testing.T) {
	home := t.TempDir()
	if err := SetCurrent(home, "default"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ScopedDir(home), scopeFileName("env:claude_code"))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("scoped\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	profile, err := currentProfileFor(home, "claude_code", "")
	if err == nil || profile != "" || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("currentProfileFor with unreadable scoped pointer = (%q, %v), want typed read failure without machine fallback", profile, err)
	}
}

func TestProfileSourceAbsenceAndUnreadabilityDiffer(t *testing.T) {
	home := t.TempDir()
	path := sourcePath(home, "profile-a")
	if _, err := readSource(home, "profile-a"); err == nil || !strings.Contains(err.Error(), stateread.DiagAbsent) {
		t.Fatalf("absent profile source error = %v, want typed absence", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	if _, err := readSource(home, "profile-a"); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("unreadable profile source error = %v, want typed unreadable diagnostic for %s", err, path)
	}
}

func TestUnreadableDefaultProfileSourceDoesNotRecreateDefault(t *testing.T) {
	home := t.TempDir()
	path := sourcePath(home, DefaultProfile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	if err := ensureDefault(nil, home, Policy{}); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("ensureDefault with unreadable source = %v, want typed read failure for %s", err, path)
	}
}

func TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer(t *testing.T) {
	fx := writeManagedFixture(t, "profile-a")
	if _, owner, err := overlayOwner(fx.home, fx.profile); err != nil || owner {
		t.Fatalf("readable root lock overlay result = (owner=%v, err=%v), want no owner", owner, err)
	}
	path := lockPath(fx.home, fx.profile)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := overlayOwner(fx.home, fx.profile); err == nil || !strings.Contains(err.Error(), stateread.DiagAbsent) {
		t.Fatalf("absent profile lock error = %v, want typed absence", err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	makeStateFileUnreadable(t, path)
	if _, _, err := overlayOwner(fx.home, fx.profile); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("unreadable profile lock error = %v, want typed unreadable diagnostic for %s", err, path)
	}
}

func TestImportAbsentAndUnreadableRootSurfaceDiffer(t *testing.T) {
	t.Run("absent is not a loss", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		seedCurrentDefault(t, home)
		_, _, _, err := Import(home, seams.options(Policy{}))
		if err != nil {
			t.Fatalf("Import with an absent root surface: %v", err)
		}
	})

	t.Run("unreadable is a typed loss", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		path := filepath.Join(seams.native["claude_code"], "CLAUDE.md")
		writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
		makeStateFileUnreadable(t, path)
		seedCurrentDefault(t, home)
		_, _, _, err := Import(home, seams.options(Policy{}))
		if err == nil || !strings.Contains(err.Error(), DiagImportLossy) || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
			t.Fatalf("Import unreadable root error = %v; want %s and %s naming %s", err, DiagImportLossy, stateread.DiagUnreadable, path)
		}
	})
}

func TestImportAbsentAndUnreadableSkillsLedgerDiffer(t *testing.T) {
	makeFixture := func(t *testing.T) (string, importSeams, string) {
		t.Helper()
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
		ids := newGitIdentities(t)
		skills := filepath.Join(seams.native["claude_code"], "skills")
		if err := os.MkdirAll(skills, 0o755); err != nil {
			t.Fatal(err)
		}
		gitSkillCheckout(t, ids, skills, "foreign", "https://example.com/skills/foreign")
		seedCurrentDefault(t, home)
		return home, seams, filepath.Join(skills, ".csk-managed.json")
	}

	t.Run("absent ledger imports the entry", func(t *testing.T) {
		home, seams, ledger := makeFixture(t)
		if _, err := os.Lstat(ledger); !os.IsNotExist(err) {
			t.Fatalf("fixture ledger state = %v, want absent", err)
		}
		info, _, _, err := Import(home, seams.options(Policy{}))
		if err != nil {
			t.Fatalf("Import with absent ledger: %v", err)
		}
		if member, ok := lockMember(info.Lock, "foreign"); !ok || member.Kind != "skill" {
			t.Fatalf("absent ledger import lock = %+v, want foreign skill", info.Lock.Members)
		}
	})

	t.Run("unreadable ledger is a typed loss", func(t *testing.T) {
		home, seams, ledger := makeFixture(t)
		if err := os.WriteFile(ledger, []byte(`{"schema_version":1,"entries":[]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		makeStateFileUnreadable(t, ledger)
		_, _, _, err := Import(home, seams.options(Policy{}))
		if err == nil || !strings.Contains(err.Error(), DiagImportLossy) || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), ledger) {
			t.Fatalf("Import unreadable ledger error = %v; want %s and %s naming %s", err, DiagImportLossy, stateread.DiagUnreadable, ledger)
		}
	})
}

func TestImportAbsentAndUnreadableSkillsSurfaceDiffer(t *testing.T) {
	t.Run("absent skills directory is empty", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
		seedCurrentDefault(t, home)
		_, _, _, err := Import(home, seams.options(Policy{}))
		if err != nil {
			t.Fatalf("Import with absent skills directory: %v", err)
		}
	})

	t.Run("unreadable skills directory is a typed loss", func(t *testing.T) {
		home := t.TempDir()
		pinHomes(t)
		seams := pinImportSeams(t)
		writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
		skills := filepath.Join(seams.native["claude_code"], "skills")
		if err := os.MkdirAll(skills, 0o700); err != nil {
			t.Fatal(err)
		}
		makeStateDirectoryUnreadable(t, skills)
		seedCurrentDefault(t, home)
		_, _, _, err := Import(home, seams.options(Policy{}))
		if err == nil || !strings.Contains(err.Error(), DiagImportLossy) || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), skills) {
			t.Fatalf("Import unreadable skills directory error = %v; want %s and %s naming %s", err, DiagImportLossy, stateread.DiagUnreadable, skills)
		}
	})
}

func TestUpdatePathSnapshotUnreadableIsNotAbsent(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "pk", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "original\n"})
	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	pinned, ok := lockMember(info.Lock, "pk")
	if !ok || pinned.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", info.Lock.Members)
	}
	entry := contextstore.EntryDir(home, contextlock.KindContext, "pk", pinned.StateHash)
	manifest := filepath.Join(entry, "agent-context.json")
	makeStateFileUnreadable(t, manifest)
	_, _, err = UpdateWithPolicy(home, "pk", Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), "path snapshot cannot be read") {
		t.Fatalf("Update unreadable snapshot error = %v; want %s and %s", err, DiagSourceInvalid, stateread.DiagUnreadable)
	}
}

func TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer(t *testing.T) {
	t.Run("absent marker is known unprovisioned", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		if err := SetCurrent(fx.home, "acme"); err != nil {
			t.Fatal(err)
		}
		managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
		if err := os.Remove(filepath.Join(managed, envmarker.Name)); err != nil {
			t.Fatal(err)
		}
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		row := findScopeHome(status, "acme", "codex_cli")
		if row == nil || !row.ProvisionedKnown || row.Provisioned || row.Diagnostic != nil {
			t.Fatalf("absent marker scope row = %+v; want known unprovisioned", row)
		}
	})

	t.Run("unreadable marker is unknown with a typed diagnostic", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		if err := SetCurrent(fx.home, "acme"); err != nil {
			t.Fatal(err)
		}
		managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
		marker := filepath.Join(managed, envmarker.Name)
		makeStateFileUnreadable(t, marker)
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		row := findScopeHome(status, "acme", "codex_cli")
		if row == nil || row.ProvisionedKnown || row.Provisioned || row.Diagnostic == nil || row.Diagnostic.Code != envmarker.DiagMarkerUnreadable || row.Diagnostic.Path != marker {
			t.Fatalf("unreadable marker scope row = %+v; want unknown with %s", row, envmarker.DiagMarkerUnreadable)
		}
		if !status.NonCurrent {
			t.Fatal("unreadable scope marker must make status non-current")
		}
	})
}

func TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer(t *testing.T) {
	makeFixture := func(t *testing.T) (*managedFixture, string) {
		t.Helper()
		fx := writeManagedFixture(t, "acme")
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		if err := SetCurrent(fx.home, "default"); err != nil {
			t.Fatal(err)
		}
		if err := Remove(fx.home, "acme", false); err != nil {
			t.Fatal(err)
		}
		managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
		return fx, filepath.Join(managed, envmarker.Name)
	}

	t.Run("absent marker is not an orphan", func(t *testing.T) {
		fx, marker := makeFixture(t)
		if err := os.Remove(marker); err != nil {
			t.Fatal(err)
		}
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		if hasDiagnostic(status.Diagnostics, marker, envmarker.DiagMarkerUnreadable) || containsPath(status.Orphans, filepath.Dir(marker)) {
			t.Fatalf("absent orphan marker produced diagnostics=%+v or orphans=%v", status.Diagnostics, status.Orphans)
		}
	})

	t.Run("unreadable marker is reported, not treated as absence", func(t *testing.T) {
		fx, marker := makeFixture(t)
		makeStateFileUnreadable(t, marker)
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		if !hasDiagnostic(status.Diagnostics, marker, envmarker.DiagMarkerUnreadable) || !status.NonCurrent {
			t.Fatalf("unreadable orphan marker diagnostics=%+v non_current=%v", status.Diagnostics, status.NonCurrent)
		}
		if containsPath(status.Orphans, filepath.Dir(marker)) {
			t.Fatalf("unknown marker was reported as an orphan: %v", status.Orphans)
		}
	})
}

func TestStatusUnreadableOrphanInventoryIsNotEmpty(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	root := EnvRoot(fx.home)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	makeStateDirectoryUnreadable(t, root)
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	if !hasDiagnostic(status.Diagnostics, root, stateread.DiagUnreadable) || !status.NonCurrent {
		t.Fatalf("unreadable orphan inventory diagnostics=%+v non_current=%v", status.Diagnostics, status.NonCurrent)
	}
}

func TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer(t *testing.T) {
	t.Run("absent backup inventory is known empty", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		row := findHome(status, "acme", "codex_cli")
		if row == nil || !row.BackupsKnown || row.Backups != 0 {
			t.Fatalf("absent backup inventory row = %+v; want known zero", row)
		}
	})

	t.Run("unreadable backup inventory is unknown and non-current", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
		backupRoot := filepath.Join(managed, ".agent-environment-backup")
		if err := os.Mkdir(backupRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		makeStateDirectoryUnreadable(t, backupRoot)
		status, err := StatusOf(statusRequest(fx))
		if err != nil {
			t.Fatal(err)
		}
		row := findHome(status, "acme", "codex_cli")
		if row == nil || row.BackupsKnown || !containsText(row.Findings, envregistry.DiagBackupRecordUnreadable) || !status.NonCurrent {
			t.Fatalf("unreadable backup row = %+v; non_current=%v", row, status.NonCurrent)
		}
	})
}

func TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer(t *testing.T) {
	makeFixture := func(t *testing.T) (string, map[string]string, string) {
		t.Helper()
		home := t.TempDir()
		homes := pinHomes(t)
		source := t.TempDir()
		writePackage(t, source, "acme", "1.0.0", "hello\n")
		if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
			t.Fatal(err)
		}
		if _, err := Use(home, "acme", "", "", false); err != nil {
			t.Fatal(err)
		}
		if err := SetCurrent(home, "default"); err != nil {
			t.Fatal(err)
		}
		return home, homes, filepath.Join(homes["claude_code"], envmarker.Name)
	}

	t.Run("absent marker permits purge", func(t *testing.T) {
		home, _, marker := makeFixture(t)
		if err := os.Remove(marker); err != nil {
			t.Fatal(err)
		}
		if err := Remove(home, "acme", true); err != nil {
			t.Fatalf("purge with absent marker: %v", err)
		}
		if _, err := readSource(home, "acme"); err == nil {
			t.Fatal("absent-marker purge retained the profile")
		}
	})

	t.Run("unreadable marker refuses before purge writes", func(t *testing.T) {
		home, homes, marker := makeFixture(t)
		makeStateFileUnreadable(t, marker)
		if err := Remove(home, "acme", true); err == nil || !strings.Contains(err.Error(), envmarker.DiagMarkerUnreadable) || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
			t.Fatalf("purge error = %v; want typed unreadable marker refusal", err)
		}
		if _, err := readSource(home, "acme"); err != nil {
			t.Fatalf("failed purge removed the profile: %v", err)
		}
		if _, err := os.Stat(marker); err != nil {
			t.Fatalf("failed purge removed unreadable marker: %v", err)
		}
		if _, err := os.Stat(filepath.Join(homes["claude_code"], "CLAUDE.md")); err != nil {
			t.Fatalf("failed purge removed recorded surface: %v", err)
		}
	})
}

func findScopeHome(status *Status, profile, env string) *ScopeHome {
	for i := range status.Scopes {
		if status.Scopes[i].Profile == profile && status.Scopes[i].Environment == env {
			return &status.Scopes[i]
		}
	}
	return nil
}

func hasDiagnostic(diagnostics []StateDiagnostic, path, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Path == path && diagnostic.Code == code {
			return true
		}
	}
	return false
}

func containsPath(paths []string, path string) bool {
	for _, candidate := range paths {
		if candidate == path {
			return true
		}
	}
	return false
}

func containsText(values []string, text string) bool {
	for _, value := range values {
		if strings.Contains(value, text) {
			return true
		}
	}
	return false
}
