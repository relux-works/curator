package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
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
		t.Skipf("this environment can read a mode-000 directory: chmod refused: %v", err)
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

// TestPathReinstallBareAndNonCurrentMovesPin drives the §1 reinstall in the
// two addressing modes the immutable-snapshot test does not take: the bare
// `profile install <same path>` form with no --as (the CLI row's default
// form), and a reinstall of a profile that is not current. Production entry
// point: Install. A mutant restricting the reinstall arm to
// `options.As != ""` restores the dead reinstall through the default form,
// and a mutant restricting it to the current profile restores it for every
// non-current profile; both must fail this test.
func TestPathReinstallBareAndNonCurrentMovesPin(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	bare := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, bare,
		`{"schema_version": 1, "name": "bare", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "original\n"})
	info, _, _, err := Install(home, InstallOptions{Operand: bare})
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "bare" {
		t.Fatalf("profile %q, want the manifest name", info.Name)
	}
	before, ok := lockMember(info.Lock, "bare")
	if !ok || before.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", info.Lock.Members)
	}
	// A second profile installs without --use, so the machine stays on
	// bare and second is not current.
	other := filepath.Join(t.TempDir(), "other")
	writeManifestPackage(t, other,
		`{"schema_version": 1, "name": "second", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "second\n"})
	secondInfo, _, _, err := Install(home, InstallOptions{Operand: other, As: "second"})
	if err != nil {
		t.Fatal(err)
	}
	secondBefore, ok := lockMember(secondInfo.Lock, "second")
	if !ok || secondBefore.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", secondInfo.Lock.Members)
	}
	if machine, _ := Current(home); machine != "bare" {
		t.Fatalf("current=%q, want bare", machine)
	}
	if err := os.WriteFile(filepath.Join(bare, "context", "a.md"), []byte("EDITED BARE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "context", "a.md"), []byte("EDITED SECOND\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The bare form with no --as must move the pin.
	reinstalled, activated, isUpdated, err := Install(home, InstallOptions{Operand: bare})
	if err != nil {
		t.Fatalf("bare reinstall: %v", err)
	}
	if !isUpdated || activated {
		t.Fatalf("bare reinstall flags activated=%v updated=%v, want false true", activated, isUpdated)
	}
	afterBare, ok := lockMember(reinstalled.Lock, "bare")
	if !ok || afterBare.StateHash == "" {
		t.Fatalf("reinstalled lock members %+v carry no state pin", reinstalled.Lock.Members)
	}
	if afterBare.StateHash == before.StateHash {
		t.Fatal("bare reinstall of the edited tree pinned the same hash: the §1 reinstall is dead through the default form")
	}
	// A reinstall of the non-current profile must move its pin without
	// switching the machine.
	reinstalledSecond, activated, isUpdated, err := Install(home, InstallOptions{Operand: other, As: "second"})
	if err != nil {
		t.Fatalf("non-current reinstall: %v", err)
	}
	if !isUpdated || activated {
		t.Fatalf("non-current reinstall flags activated=%v updated=%v, want false true", activated, isUpdated)
	}
	afterSecond, ok := lockMember(reinstalledSecond.Lock, "second")
	if !ok || afterSecond.StateHash == "" {
		t.Fatalf("reinstalled lock members %+v carry no state pin", reinstalledSecond.Lock.Members)
	}
	if afterSecond.StateHash == secondBefore.StateHash {
		t.Fatal("non-current reinstall of the edited tree pinned the same hash: the §1 reinstall is dead for a non-current profile")
	}
	if machine, _ := Current(home); machine != "bare" {
		t.Fatalf("reinstall without --use switched current=%q, want bare", machine)
	}
}

// installBlockingOverlayRoot installs a clean path root and returns the
// manager home, the root source directory, and the lock hash before any
// overlay joins the closure.
func installBlockingOverlayRoot(t *testing.T) (home, root string, oldHash string) {
	t.Helper()
	home = t.TempDir()
	pinHomes(t)
	root = filepath.Join(t.TempDir(), "root")
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "root\n"})
	if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
		t.Fatal(err)
	}
	_, oldHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	return home, root, oldHash
}

// TestUpdateBlocksSecretOverlayMember drives UpdateWithPolicy onto a path
// profile whose machine overlay carries secret material: the deterministic
// detector reports a blocking finding on the new member, updateLocked
// refuses with profile_update_blocked, and the old lock stands. Production
// entry point: UpdateWithPolicy. A mutant admitting exactly a blocking
// context member (report.Blocking() but not for KindContext) must fail
// this test.
func TestUpdateBlocksSecretOverlayMember(t *testing.T) {
	home, _, oldHash := installBlockingOverlayRoot(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": directorySecret()})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	_, _, err := UpdateWithPolicy(home, "acme", policy)
	if err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
		t.Fatalf("err = %v, want %s", err, DiagUpdateBlocked)
	}
	if !strings.Contains(err.Error(), "blocking finding") {
		t.Fatalf("err = %v, want the blocking-finding reason", err)
	}
	_, afterHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if afterHash != oldHash {
		t.Fatal("a blocked update moved the lock: the old lock must stand")
	}
}

// TestReinstallBlocksSecretOverlayMember drives Install of the same path
// source with a secret-carrying machine overlay: the reinstall's new-member
// gate refuses with profile_update_blocked exactly as the update twin does.
// Production entry point: Install. It kills the reinstall copy of the
// mutant TestUpdateBlocksSecretOverlayMember kills for updateLocked.
func TestReinstallBlocksSecretOverlayMember(t *testing.T) {
	home, root, oldHash := installBlockingOverlayRoot(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": directorySecret()})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
		t.Fatalf("err = %v, want %s", err, DiagUpdateBlocked)
	}
	if !strings.Contains(err.Error(), "blocking finding") {
		t.Fatalf("err = %v, want the blocking-finding reason", err)
	}
	_, afterHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if afterHash != oldHash {
		t.Fatal("a blocked reinstall moved the lock: the old lock must stand")
	}
}

// TestUpdateBlocksRevokedOverlayMember drives UpdateWithPolicy onto a path
// profile whose machine overlay is clean but revoked by source: the
// deterministic detector passes, and the strict member audit refuses with
// profile_update_blocked. This is a blocking-but-not-strict finding —
// reachable in production through machine revocations — so the second gate
// in the new-member loop is load-bearing, not decoration. Production entry
// point: UpdateWithPolicy. A mutant ignoring the strict-audit error for a
// context member must fail this test.
func TestUpdateBlocksRevokedOverlayMember(t *testing.T) {
	home, _, oldHash := installBlockingOverlayRoot(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	// A path member carries no network identity (its lock source is
	// empty), so the revocation names the snapshot's content hash — the
	// production mechanism for pinning local sources (§9.1).
	contentHash, err := hashing.ContentSHA256(overlay, nil)
	if err != nil {
		t.Fatal(err)
	}
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
		Revocations:          []string{contentHash},
	}
	_, _, err = UpdateWithPolicy(home, "acme", policy)
	if err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
		t.Fatalf("err = %v, want %s", err, DiagUpdateBlocked)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("err = %v, want the revocation reason", err)
	}
	_, afterHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if afterHash != oldHash {
		t.Fatal("a blocked update moved the lock: the old lock must stand")
	}
}

// TestReinstallBlocksRevokedOverlayMember drives Install of the same path
// source with a clean but revoked machine overlay: the reinstall's strict
// member audit refuses with profile_update_blocked exactly as the update
// twin does. Production entry point: Install. It kills the reinstall copy
// of the mutant TestUpdateBlocksRevokedOverlayMember kills for updateLocked.
func TestReinstallBlocksRevokedOverlayMember(t *testing.T) {
	home, root, oldHash := installBlockingOverlayRoot(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "overlay\n"})
	// A path member carries no network identity (its lock source is
	// empty), so the revocation names the snapshot's content hash — the
	// production mechanism for pinning local sources (§9.1).
	contentHash, err := hashing.ContentSHA256(overlay, nil)
	if err != nil {
		t.Fatal(err)
	}
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
		Revocations:          []string{contentHash},
	}
	_, _, _, err = Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
		t.Fatalf("err = %v, want %s", err, DiagUpdateBlocked)
	}
	if !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("err = %v, want the revocation reason", err)
	}
	_, afterHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if afterHash != oldHash {
		t.Fatal("a blocked reinstall moved the lock: the old lock must stand")
	}
}
