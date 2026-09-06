package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
)

// Production entry point under test: Install. Every fixture builds paths
// with filepath.Join over a platform-absolute temporary base; no POSIX
// literal feeds a filepath gate.

// TestMissingPathOperandIsPathMissing drives Install with a path operand
// naming no existing entry: the absence is profile_source_path_missing,
// never a failed read.
func TestMissingPathOperandIsPathMissing(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	absent := filepath.Join(t.TempDir(), "absent")
	_, _, _, err := Install(home, InstallOptions{Operand: absent})
	if err == nil || !strings.Contains(err.Error(), DiagPathMissing) {
		t.Fatalf("err = %v, want %s", err, DiagPathMissing)
	}
	if strings.Contains(err.Error(), DiagPathUnreadable) {
		t.Fatalf("err = %v, absence must never report unreadable", err)
	}
}

// TestMissingOverlayPathIsPathMissing drives Install with a
// machine-declared overlay naming no existing entry: the overlay failure
// carries the source's section 1.1 absence diagnostic.
func TestMissingOverlayPathIsPathMissing(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays: map[string][]OverlaySpec{
			"acme": {{Source: filepath.Join(t.TempDir(), "absent")}},
		},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagPathMissing) {
		t.Fatalf("err = %v, want %s", err, DiagPathMissing)
	}
}

// TestNonDirectoryPathIsSourceInvalid drives Install with a path operand
// naming a regular file: a path package is a directory tree, so the
// operand is profile_source_invalid.
func TestNonDirectoryPathIsSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := Install(home, InstallOptions{Operand: file})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("err = %v, want %s", err, DiagSourceInvalid)
	}
}

// TestNestedGitIsSourceInvalid drives Install with a .git entry below the
// source root: the snapshot is profile_source_invalid.
func TestNestedGitIsSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	if err := os.WriteFile(filepath.Join(source, "context", ".git"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("err = %v, want %s", err, DiagSourceInvalid)
	}
}

// TestSymlinkInPathIsSourceInvalid drives Install with a symbolic link in
// the source tree: the snapshot discipline is profile_source_invalid.
func TestSymlinkInPathIsSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	link := filepath.Join(source, "context", "link.md")
	if err := os.Symlink(filepath.Join(source, "context", "a.md"), link); err != nil {
		t.Skipf("this host cannot create symlinks: %v", err)
	}
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("err = %v, want %s", err, DiagSourceInvalid)
	}
}

// TestFoldedPathsAreSourceInvalid drives Install with two module files
// folding to one platform path: the snapshot is profile_source_invalid
// before anything is written.
func TestFoldedPathsAreSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	if err := os.WriteFile(filepath.Join(source, "context", "A.md"), []byte("fold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// On a case-folding filesystem (default macOS APFS) the second write
	// lands on the first file and no snapshot can observe two entries;
	// the Linux lanes run case-sensitively and exercise the gate there.
	if payload, err := os.ReadFile(filepath.Join(source, "context", "a.md")); err == nil && string(payload) == "fold\n" {
		t.Skip("per-directory case sensitivity folds entries on this host; the collision gate runs on case-sensitive lanes")
	}
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("err = %v, want %s", err, DiagSourceInvalid)
	}
	if _, statErr := readSource(home, "acme"); statErr == nil {
		t.Fatal("failed install must leave no profile record")
	}
}

// TestRootGitExcluded drives Install with a .git directory at the source
// root: the entry is excluded from the immutable snapshot and the install
// succeeds.
func TestRootGitExcluded(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	gitDir := filepath.Join(source, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "junk"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "acme")
	if !ok || member.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", info.Lock.Members)
	}
}

// TestUnreadablePathIsUnreadable drives Install with a source tree whose
// module directory cannot be read: the failure is
// profile_source_path_unreadable, never absence.
func TestUnreadablePathIsUnreadable(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	contextDir := filepath.Join(source, "context")
	// The probe below decides: a host that reads through mode 000
	// (superuser, or Windows ACL semantics) skips under the classified
	// host-capability reason instead of asserting untestable behaviour.
	if err := os.Chmod(contextDir, 0o000); err != nil {
		t.Skipf("this environment can read a mode-000 directory: chmod refused: %v", err)
	}
	defer func() { _ = os.Chmod(contextDir, 0o755) }()
	if payload, err := os.ReadFile(filepath.Join(contextDir, "a.md")); err == nil {
		_ = payload
		t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
	}
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.Contains(err.Error(), DiagPathUnreadable) {
		t.Fatalf("err = %v, want %s", err, DiagPathUnreadable)
	}
	if strings.Contains(err.Error(), DiagPathMissing) {
		t.Fatalf("err = %v, a failed read must never report absence", err)
	}
	// §1.1 requires the specific diagnostic to lead (§8.4: unreadable
	// evidence is reported as unreadable, never as something else). A
	// wrapper that prefixes profile_source_invalid shadows it.
	if !strings.HasPrefix(err.Error(), DiagPathUnreadable+":") {
		t.Fatalf("err = %q, want the leading diagnostic %s", err, DiagPathUnreadable)
	}
	if strings.Contains(err.Error(), DiagSourceInvalid+": "+DiagPathUnreadable) {
		t.Fatalf("err = %q carries the doubled wrapper prefix", err)
	}
}

// TestPathSnapshotImmutableAcrossUpdateSyncUse drives Install, edits the
// source directory afterwards, and asserts the state pin does not move
// across Update, Sync and Use, but does move across a same-name reinstall
// (environments §1: installation copies the tree as an immutable snapshot
// and never reads the source directory again until the operator
// reinstalls). Production entry points: Install, UpdateWithPolicy,
// SyncWithPolicy, UseWithPolicy — the same calls the CLI rows reach. A
// mutant that re-reads source.Path in updateLocked moves the pin on Update
// and must fail this test; a mutant that routes a same-source path install
// through updateLocked keeps the pin on reinstall and must fail it too.
func TestPathSnapshotImmutableAcrossUpdateSyncUse(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "pk", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "original\n"})
	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatal(err)
	}
	before, ok := lockMember(info.Lock, "pk")
	if !ok || before.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", info.Lock.Members)
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("EDITED AFTER INSTALL\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updated, moved, err := UpdateWithPolicy(home, "pk", Policy{})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if moved {
		t.Fatal("update of a path root must not move the lock: the snapshot is immutable")
	}
	afterUpdate, ok := lockMember(updated.Lock, "pk")
	if !ok || afterUpdate.StateHash != before.StateHash {
		t.Fatalf("update moved the pin %q -> %q", before.StateHash, afterUpdate.StateHash)
	}
	if _, err := SyncWithPolicy(home, Policy{}); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if _, err := UseWithPolicy(home, "pk", "", "", false, Policy{}); err != nil {
		t.Fatalf("use: %v", err)
	}
	lock, _, err := readLock(home, "pk")
	if err != nil {
		t.Fatal(err)
	}
	afterAll, ok := lockMember(lock, "pk")
	if !ok || afterAll.StateHash != before.StateHash {
		t.Fatalf("pin moved across update/sync/use: %q -> %q", before.StateHash, afterAll.StateHash)
	}
	// The §1 reinstall is the one operator path that refreshes the
	// snapshot: installing the edited tree under the same name must move
	// the pin and re-materialize the current profile's surface with the
	// new bytes.
	reinstalled, activated, isUpdated, err := Install(home, InstallOptions{Operand: source, As: "pk"})
	if err != nil {
		t.Fatalf("reinstall: %v", err)
	}
	if !isUpdated || activated {
		t.Fatalf("reinstall flags activated=%v updated=%v, want false true", activated, isUpdated)
	}
	afterReinstall, ok := lockMember(reinstalled.Lock, "pk")
	if !ok || afterReinstall.StateHash == "" {
		t.Fatalf("reinstalled lock members %+v carry no state pin", reinstalled.Lock.Members)
	}
	if afterReinstall.StateHash == before.StateHash {
		t.Fatal("reinstall of the edited tree pinned the same hash: the §1 reinstall is dead")
	}
	payload, err := os.ReadFile(filepath.Join(homes["claude_code"], "CLAUDE.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatalf("re-materialized surface: %v", err)
	}
	if !strings.Contains(string(payload), "EDITED AFTER INSTALL") {
		t.Fatalf("reinstall left the materialized surface stale:\n%s", payload)
	}
	// The edit is real even without the reinstall path: a fresh install of
	// the edited tree under another name pins a different hash, so the
	// update half of this test is not vacuous.
	second, _, _, err := Install(home, InstallOptions{Operand: source, As: "pk2"})
	if err != nil {
		t.Fatal(err)
	}
	fresh, ok := lockMember(second.Lock, "pk")
	if !ok || fresh.StateHash == "" {
		t.Fatalf("second lock members %+v carry no state pin", second.Lock.Members)
	}
	if fresh.StateHash == before.StateHash {
		t.Fatal("fresh install of the edited tree pinned the same hash: the edit left no trace")
	}
	if fresh.StateHash != afterReinstall.StateHash {
		t.Fatalf("same-tree reinstall pin %q != fresh-install pin %q", afterReinstall.StateHash, fresh.StateHash)
	}
}

// TestPathUpdateAllSucceedsWithImportedProfile drives Import (which deletes
// its §9.6 staging directory, so the recorded source.Path names no existing
// entry) followed by the per-profile update loop the `profile update --all`
// row runs: ListWithPolicy skipping default, UpdateWithPolicy each.
// Production entry points: Import, ListWithPolicy, UpdateWithPolicy. Before
// the §1 fix, the update re-read source.Path and failed the whole machine
// with profile_source_path_missing.
func TestPathUpdateAllSucceedsWithImportedProfile(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "native\n")
	seedCurrentDefault(t, home)
	info, _, _, err := Import(home, seams.options(Policy{}))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	source, err := readSource(home, info.Name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source.Path); !os.IsNotExist(err) {
		t.Fatalf("import staging %q still exists: the test does not exercise the deleted-source shape", source.Path)
	}
	profiles, err := ListWithPolicy(home, Policy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range profiles {
		if profile.Name == DefaultProfile {
			continue
		}
		if _, _, err := UpdateWithPolicy(home, profile.Name, Policy{}); err != nil {
			t.Fatalf("update %s: %v", profile.Name, err)
		}
	}
}

// TestUnreadablePathRootLeadsUnreadable drives Install with a path root
// that itself cannot be read: the canonical §1.1 condition reports
// profile_source_path_unreadable leading, never profile_source_invalid.
func TestUnreadablePathRootLeadsUnreadable(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	if err := os.Chmod(source, 0o000); err != nil {
		t.Skipf("chmod refused: %v", err)
	}
	defer func() { _ = os.Chmod(source, 0o755) }()
	if _, err := os.ReadDir(source); err == nil {
		t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
	}
	_, _, _, err := Install(home, InstallOptions{Operand: source})
	if err == nil || !strings.HasPrefix(err.Error(), DiagPathUnreadable+":") {
		t.Fatalf("err = %v, want leading %s", err, DiagPathUnreadable)
	}
}

// TestUpdatePathWithoutStatePinIsSourceInvalid drives UpdateWithPolicy on a
// path profile whose lock carries no state pin for its root: the root member
// is rewritten to a commit pin (the only validated lock shape without a
// state hash), so readLock still validates and updateLocked must refuse with
// profile_source_invalid naming the missing pin. Production entry point:
// UpdateWithPolicy. A mutant dropping the empty-pin refusal admits the lock
// and must fail this test.
func TestUpdatePathWithoutStatePinIsSourceInvalid(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "pk", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "original\n"})
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(home, "pk")
	if err != nil {
		t.Fatal(err)
	}
	for i, member := range lock.Members {
		if member.Kind == contextlock.KindContext && member.Name == "pk" {
			lock.Members[i].StateHash = ""
			lock.Members[i].Commit = strings.Repeat("ab", 20)
			lock.Members[i].Source = "https://example.com/pk"
		}
	}
	canonical, err := lock.Canonical()
	if err != nil {
		t.Fatalf("tampered lock must still validate: %v", err)
	}
	if err := os.WriteFile(lockPath(home, "pk"), canonical, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err = UpdateWithPolicy(home, "pk", Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), "carries no state pin") {
		t.Fatalf("err = %v, want %s naming the missing state pin", err, DiagSourceInvalid)
	}
}

// TestUpdatePathMissingSnapshotIsSourceInvalid drives UpdateWithPolicy on a
// path profile whose store snapshot is gone: the entry under the lock's
// state pin is removed, so the snapshot read fails and updateLocked must
// refuse with profile_source_invalid naming the unreadable snapshot — never
// a silent unchanged. Production entry point: UpdateWithPolicy. A mutant
// returning the old lock as unchanged with a nil error on the read failure
// (§8.4: a failed read turned into "nothing to do") must fail this test.
func TestUpdatePathMissingSnapshotIsSourceInvalid(t *testing.T) {
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
	if err := os.RemoveAll(entry); err != nil {
		t.Fatal(err)
	}
	_, _, err = UpdateWithPolicy(home, "pk", Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), "path snapshot cannot be read") {
		t.Fatalf("err = %v, want %s naming the unreadable snapshot", err, DiagSourceInvalid)
	}
}

// TestUpdatePathSnapshotNameMismatchIsSourceInvalid drives UpdateWithPolicy
// on a path profile whose store snapshot names a different root than the
// lock: the snapshot manifest is rewritten, so updateLocked must refuse with
// profile_source_invalid naming both names. Production entry point:
// UpdateWithPolicy. A mutant deleting the snapshot-name check admits the
// mismatched snapshot and must fail this test.
func TestUpdatePathSnapshotNameMismatchIsSourceInvalid(t *testing.T) {
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
	manifestPath := filepath.Join(entry, "agent-context.json")
	payload, err := os.ReadFile(manifestPath) // #nosec G304 -- test store entry
	if err != nil {
		t.Fatal(err)
	}
	renamed := strings.Replace(string(payload), `"pk"`, `"other"`, 1)
	if renamed == string(payload) {
		t.Fatal("snapshot manifest names no pk to rename")
	}
	if err := os.WriteFile(manifestPath, []byte(renamed), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err = UpdateWithPolicy(home, "pk", Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), "snapshot names") {
		t.Fatalf("err = %v, want %s naming the snapshot mismatch", err, DiagSourceInvalid)
	}
}
