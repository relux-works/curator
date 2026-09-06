package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
