//go:build unix

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
)

// Drive run() through install, takeover, and unmanage. Seed a current profile
// first so the subsequent install does not implicitly activate over the link.
func TestEnvUnmanageTakeoverRestoresSymlink(t *testing.T) {
	for _, shape := range []string{"absolute", "relative", "dangling", "directory"} {
		t.Run(shape, func(t *testing.T) {
			source, _ := profileHome(t)
			seed := t.TempDir()
			writeContextPackage(t, seed, "seed", "1.0.0", "seed context\n")
			if code, _, stderr := runProfile(t, source, "profile", "install", seed); code != exitOK {
				t.Fatalf("seed install exit %d: %s", code, stderr)
			}
			if code, _, stderr := runProfile(t, source, "env", "unmanage", "--env", "claude_code"); code != exitOK {
				t.Fatalf("seed unmanage exit %d: %s", code, stderr)
			}
			native := os.Getenv("CLAUDE_CONFIG_DIR")
			full := filepath.Join(native, "CLAUDE.md")
			outside := t.TempDir()
			external := filepath.Join(outside, "private.md")
			original := "external operator context\n"
			if err := os.WriteFile(external, []byte(original), 0o600); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(external)
			if err != nil {
				t.Fatal(err)
			}
			target := external
			switch shape {
			case "relative":
				target, err = filepath.Rel(native, external)
				if err != nil {
					t.Fatal(err)
				}
			case "dangling":
				target = filepath.Join(outside, "absent.md")
			case "directory":
				target = outside
			}
			if err := os.Symlink(target, full); err != nil {
				t.Fatal(err)
			}
			pkg := t.TempDir()
			writeContextPackage(t, pkg, "acme", "1.0.0", "managed context\n")
			if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
				t.Fatalf("profile install exit %d: %s", code, stderr)
			}
			assertUnmanageSymlink(t, full, target)
			if code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--takeover", "--env", "claude_code"); code != exitOK {
				t.Fatalf("profile use --takeover exit %d: %s", code, stderr)
			}
			backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
			assertUnmanageSymlink(t, backup, target)
			if code, _, stderr := runProfile(t, source, "env", "unmanage", "--restore-backups", "--env", "claude_code"); code != exitOK {
				t.Fatalf("env unmanage --restore-backups exit %d: %s", code, stderr)
			}
			assertUnmanageSymlink(t, full, target)
			assertUnmanageSymlink(t, backup, target)
			if got := string(readFileForUnmanageTest(t, external)); got != original {
				t.Fatalf("external bytes = %q, want %q", got, original)
			}
			after, err := os.Stat(external)
			if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
				t.Fatalf("external file identity, mode, or modification time changed: %v", err)
			}
			if shape == "dangling" {
				if _, err := os.Lstat(target); !os.IsNotExist(err) {
					t.Fatalf("restore created dangling target: %v", err)
				}
			}
		})
	}
}

func assertUnmanageSymlink(t *testing.T, full, target string) {
	t.Helper()
	info, err := os.Lstat(full)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink: %v", filepath.Base(full), err)
	}
	got, err := os.Readlink(full)
	if err != nil || got != target {
		t.Fatalf("%s link target = %q (%v), want %q", filepath.Base(full), got, err, target)
	}
}

func TestEnvUnmanageSymlinkRestoreRefusesLinkedParent(t *testing.T) {
	for _, route := range []string{"context", "nested/context"} {
		t.Run(route, func(t *testing.T) {
			fixture := newUnmanageFixture(t, true)
			outside := t.TempDir()
			external := filepath.Join(outside, "private.md")
			if err := os.WriteFile(external, []byte("external context\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(filepath.Dir(fixture.backupFile), filepath.FromSlash(route), "private.md")
			if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, backup); err != nil {
				t.Fatal(err)
			}
			parent := filepath.Join(fixture.nativeHome, filepath.FromSlash(route))
			if err := os.MkdirAll(filepath.Dir(parent), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, parent); err != nil {
				t.Fatal(err)
			}
			beforeSurface := readFileForUnmanageTest(t, fixture.surfacePath)
			beforeMarker := readFileForUnmanageTest(t, fixture.markerPath)
			code, _, stderr := runProfile(t, fixture.source, "env", "unmanage", "--restore-backups", "--env", "claude_code")
			if code != exitFail || !strings.Contains(stderr, envprofile.DiagWriteWouldFollowLink) {
				t.Fatalf("unmanage exit %d: %s; want linked-parent refusal", code, stderr)
			}
			if string(readFileForUnmanageTest(t, fixture.surfacePath)) != string(beforeSurface) ||
				string(readFileForUnmanageTest(t, fixture.markerPath)) != string(beforeMarker) {
				t.Fatal("parent refusal changed the managed surface or marker")
			}
			assertUnmanageSymlink(t, backup, external)
			info, err := os.Lstat(external)
			if err != nil || !info.Mode().IsRegular() || string(readFileForUnmanageTest(t, external)) != "external context\n" {
				t.Fatalf("parent refusal changed the external entry: %v", err)
			}
		})
	}
}

func TestEnvUnmanageSymlinkRestoreRefusesPlannedLinkedParent(t *testing.T) {
	for _, child := range []string{"private.md", "nested/private.md"} {
		t.Run(child, func(t *testing.T) {
			fixture := newUnmanageFixture(t, true)
			outside := t.TempDir()
			external := filepath.Join(outside, filepath.FromSlash(child))
			if err := os.MkdirAll(filepath.Dir(external), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(external, []byte("external context\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			// A stale recorded child is absent in the native home. The saved parent
			// link must not make that later removal address an external child.
			marker, err := envmarker.Read(fixture.nativeHome)
			if err != nil {
				t.Fatal(err)
			}
			surface := marker.Surfaces[envmarker.SurfaceRootContext]
			surface.Paths = append(surface.Paths, "context/"+child)
			marker.Surfaces[envmarker.SurfaceRootContext] = surface
			payload, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(fixture.markerPath, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(filepath.Dir(fixture.backupFile), "context")
			if err := os.Symlink(outside, backup); err != nil {
				t.Fatal(err)
			}
			beforeSurface := readFileForUnmanageTest(t, fixture.surfacePath)
			code, _, stderr := runProfile(t, fixture.source, "env", "unmanage", "--restore-backups", "--env", "claude_code")
			if got := string(readFileForUnmanageTest(t, external)); got != "external context\n" {
				t.Fatalf("restore followed its newly restored parent link: %q", got)
			}
			if code != exitFail || !strings.Contains(stderr, envprofile.DiagWriteWouldFollowLink) {
				t.Fatalf("unmanage exit %d: %s; want planned parent-link refusal", code, stderr)
			}
			if string(readFileForUnmanageTest(t, fixture.surfacePath)) != string(beforeSurface) ||
				string(readFileForUnmanageTest(t, fixture.markerPath)) != string(payload) {
				t.Fatal("planned parent-link refusal changed the managed surface or marker")
			}
			if _, err := os.Lstat(filepath.Join(fixture.nativeHome, "context")); !os.IsNotExist(err) {
				t.Fatalf("planned parent-link refusal created a native parent: %v", err)
			}
			assertUnmanageSymlink(t, backup, outside)
		})
	}
}
