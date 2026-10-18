//go:build darwin || linux

package envsnapshot

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/sys/unix"
)

func publicationFixture(t *testing.T) Publication {
	t.Helper()
	p := Publication{Manifest: read(t, "golden/manifest.json"), Launch: read(t, "golden/launch.json"), Files: map[string][]byte{}}
	for _, f := range manifest(t).Files {
		p.Files[f.Path] = read(t, "golden/"+f.Path)
	}
	return p
}

func publishControl(t *testing.T, root string, policy DestinationPolicy) {
	t.Helper()
	p := publicationFixture(t)
	d := admitDestination(t, filepath.Join(root, "control"), policy, DestinationAbsent)
	v, err := d.Publish(context.Background(), p)
	if err != nil || v == nil || v.ManifestSHA256() != Digest(p.Manifest) {
		t.Fatal("positive publication", v, err)
	}
}

func requireOnlyEntries(t *testing.T, root string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != len(names) {
		t.Fatalf("unexpected entries: %v %v", entries, err)
	}
	want := map[string]bool{}
	for _, name := range names {
		want[name] = true
	}
	for _, entry := range entries {
		if !want[entry.Name()] {
			t.Fatalf("unexpected entry %s", entry.Name())
		}
	}
}

// Real no-replace publication produces independent private files with the exact
// captured bytes and semantic modes. Repeat publication verifies a no-op;
// a different generation refuses without changing even manifest inode/mtime.
func TestPublicationPrivateAndPrior(t *testing.T) {
	root, policy := destinationSandbox(t)
	p := publicationFixture(t)
	path := filepath.Join(root, "snapshot")
	d := admitDestination(t, path, policy, DestinationAbsent)
	v, err := d.Publish(context.Background(), p)
	if err != nil || v == nil || v.ManifestSHA256() != Digest(p.Manifest) {
		t.Fatal(v, err)
	}
	verified(t, path)
	requireOnlyEntries(t, root, "protected", "snapshot")
	before, err := os.Stat(filepath.Join(path, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err = d.Publish(context.Background(), p)
	if err != nil || v == nil {
		t.Fatal("no-op rejected", err)
	}
	p.Manifest = read(t, "golden/other-source-manifest.json")
	v, err = d.Publish(context.Background(), p)
	requireDestinationError(t, err, Mismatch, "generation-mismatch")
	if v != nil {
		t.Fatal("failure returned success")
	}
	after, err := os.Stat(filepath.Join(path, "manifest.json"))
	if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("prior changed", err)
	}
	// Captured content files, metadata and directories carry no imported xattrs
	// or special/set-id modes. SNAP-ID independently checks modes and nlink=1.
	if err := filepath.WalkDir(path, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		f, e := os.Open(name)
		if e != nil {
			return e
		}
		defer func() { _ = f.Close() }()
		size, e := unix.Flistxattr(int(f.Fd()), nil)
		if e != nil {
			return e
		}
		if size != 0 {
			names := make([]byte, size)
			if _, e := unix.Flistxattr(int(f.Fd()), names); e != nil {
				return e
			}
			// macOS may stamp its own provenance on a newly created object;
			// this is not imported source metadata. No other xattr is admitted.
			if runtime.GOOS != "darwin" || string(names) != "com.apple.provenance\x00" {
				t.Fatalf("unexpected xattrs on %s: %q", entry.Name(), names)
			}
		}
		return destinationACL(f)
	}); err != nil {
		t.Fatal(err)
	}
}

// Corrupt captured bytes, missing/extra entries and invalid metadata refuse
// before staging; none can write a partial target or a scratch entry.
func TestPublicationCapturedIdentity(t *testing.T) {
	for _, name := range []string{"bytes", "missing", "extra", "metadata", "mode"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			publishControl(t, root, policy)
			p := publicationFixture(t)
			switch name {
			case "bytes":
				p.Files["home/AGENTS.md"][0] ^= 1
			case "missing":
				delete(p.Files, "home/AGENTS.md")
			case "extra":
				p.Files["../outside"] = []byte("forbidden")
			case "metadata":
				p.Manifest = append(p.Manifest, '\n')
			case "mode":
				p.Manifest = bytes.ReplaceAll(p.Manifest, []byte(`"0600"`), []byte(`"0777"`))
			}
			d := admitDestination(t, filepath.Join(root, "target"), policy, DestinationAbsent)
			v, err := d.Publish(context.Background(), p)
			requireDestinationError(t, err, Mismatch, "")
			if v != nil {
				t.Fatal("failure returned success")
			}
			requireOnlyEntries(t, root, "protected", "control")
		})
	}
}

// A real mkdir/link/file syscall at the destination AFTER final revalidation
// cannot be overwritten by publication. The positive control uses the same
// backend without competition. A competing empty directory catches rename
// implementations that silently replace empty directories.
func TestPublicationNoReplaceRace(t *testing.T) {
	for _, name := range []string{"empty", "snapshot", "file", "link"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			publishControl(t, root, policy)
			path := filepath.Join(root, "target")
			d := admitDestination(t, path, policy, DestinationAbsent)
			var before os.FileInfo
			d.publicationHook = func(phase string, _ *os.File) error {
				if phase != "before-rename" {
					return nil
				}
				var err error
				switch name {
				case "empty":
					err = os.Mkdir(path, 0700)
				case "snapshot":
					snapshotAtDestination(t, path)
				case "file":
					err = os.WriteFile(path, []byte("preserved"), 0600)
				case "link":
					err = os.Symlink(filepath.Join(root, "control"), path)
				}
				if err != nil {
					t.Fatal(err)
				}
				before, err = os.Lstat(path)
				if err != nil {
					t.Fatal(err)
				}
				return nil
			}
			v, err := d.Publish(context.Background(), publicationFixture(t))
			requireDestinationError(t, err, DestinationInvalid, "publication-no-replace")
			if v != nil {
				t.Fatal("race returned success")
			}
			after, err := os.Lstat(path)
			if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || before.Mode() != after.Mode() {
				t.Fatal("competitor replaced", err)
			}
			if name == "empty" {
				requireOnlyEntries(t, path)
			}
			if name == "file" {
				b, e := os.ReadFile(path)
				if e != nil || string(b) != "preserved" {
					t.Fatal("competitor bytes", e)
				}
			}
			verified(t, filepath.Join(root, "control"))
			requireOnlyEntries(t, root, "protected", "control", "target")
		})
	}
}

// Replacing the ancestor with a link during the final syscall window never
// redirects writes. Publication is anchored to the retained original parent;
// rechecking the chain then refuses success and leaves that object for inspection.
func TestPublicationRetainedParentRace(t *testing.T) {
	root, policy := destinationSandbox(t)
	publishControl(t, root, policy)
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	write(t, outside, "sentinel", []byte("preserved"), 0600)
	d := admitDestination(t, filepath.Join(parent, "target"), policy, DestinationAbsent)
	d.publicationHook = func(phase string, _ *os.File) error {
		if phase != "before-rename" {
			return nil
		}
		if err := os.Rename(parent, parent+"-original"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, parent); err != nil {
			t.Fatal(err)
		}
		return nil
	}
	v, err := d.Publish(context.Background(), publicationFixture(t))
	requireDestinationError(t, err, DestinationInvalid, "ancestor-changed")
	if v != nil {
		t.Fatal("changed ancestor returned success")
	}
	requireOnlyEntries(t, outside, "sentinel")
	verified(t, filepath.Join(parent+"-original", "target"))
}

// Cancellation and an injected I/O failure clean only recorded private staging
// entries; unrelated sibling directories and links (including another snapshot)
// survive. No cancellation emits success or publishes a partial target.
func TestPublicationStagingOnlyCleanup(t *testing.T) {
	for _, phase := range []string{"before-start", "stage-created", "file-staged", "before-rename", "io-failure"} {
		t.Run(phase, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			publishControl(t, root, policy)
			write(t, root, "unrelated/sentinel", []byte("preserved"), 0600)
			if err := os.Symlink(filepath.Join(root, "unrelated"), filepath.Join(root, ".curator-snapshot-foreign")); err != nil {
				t.Fatal(err)
			}
			d := admitDestination(t, filepath.Join(root, "target"), policy, DestinationAbsent)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if phase == "before-start" {
				cancel()
			}
			d.publicationHook = func(at string, _ *os.File) error {
				if phase == "io-failure" && at == "file-staged" {
					return os.ErrPermission
				}
				if phase == at {
					cancel()
				}
				return nil
			}
			v, err := d.Publish(ctx, publicationFixture(t))
			reason := "publication-interrupted"
			if phase == "io-failure" {
				reason = "staging-io"
			}
			requireDestinationError(t, err, DestinationInvalid, reason)
			if v != nil {
				t.Fatal("failure returned success")
			}
			assertAbsent(t, filepath.Join(root, "target"))
			requireOnlyEntries(t, root, "protected", "control", "unrelated", ".curator-snapshot-foreign")
			b, e := os.ReadFile(filepath.Join(root, "unrelated/sentinel"))
			if e != nil || string(b) != "preserved" {
				t.Fatal("outside cleanup", e)
			}
			verified(t, filepath.Join(root, "control"))
		})
	}
}

// A real kernel permission failure during staging is not success. Cleanup can
// remove its empty stage without walking other trees. Not qualified under root.
func TestPublicationKernelStagingFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires unprivileged kernel permission checks")
	}
	root, policy := destinationSandbox(t)
	publishControl(t, root, policy)
	d := admitDestination(t, filepath.Join(root, "target"), policy, DestinationAbsent)
	d.publicationHook = func(phase string, stage *os.File) error {
		if phase == "stage-created" {
			if err := stage.Chmod(0500); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}
	v, err := d.Publish(context.Background(), publicationFixture(t))
	requireDestinationError(t, err, DestinationInvalid, "staging-create")
	if v != nil {
		t.Fatal("failure returned success")
	}
	requireOnlyEntries(t, root, "protected", "control")
}

// Post-publication actual bytes, size, mode, extras and metadata bind success.
// Each mismatch leaves the exact published object and tampered bytes for the
// operator; cancellation after rename follows the same no-rollback rule.
func TestPublicationPostVerification(t *testing.T) {
	for _, name := range []string{"bytes", "size", "mode", "extra", "manifest", "target-link", "cancel"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			publishControl(t, root, policy)
			path := filepath.Join(root, "target")
			d := admitDestination(t, path, policy, DestinationAbsent)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var before os.FileInfo
			d.publicationHook = func(phase string, _ *os.File) error {
				if phase != "after-rename" {
					return nil
				}
				switch name {
				case "bytes":
					b := bytes.Clone(publicationFixture(t).Files["home/AGENTS.md"])
					b[0] ^= 1
					write(t, path, "home/AGENTS.md", b, 0600)
				case "size":
					write(t, path, "home/AGENTS.md", []byte("changed size"), 0600)
				case "mode":
					if err := os.Chmod(filepath.Join(path, "home/AGENTS.md"), 0700); err != nil {
						t.Fatal(err)
					}
				case "extra":
					write(t, path, "extra", []byte("preserved"), 0600)
				case "manifest":
					write(t, path, "manifest.json", read(t, "golden/other-source-manifest.json"), 0600)
				case "target-link":
					if err := os.Rename(path, path+"-original"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(filepath.Join(root, "control"), path); err != nil {
						t.Fatal(err)
					}
				case "cancel":
					cancel()
				}
				var e error
				before, e = os.Lstat(path)
				if e != nil {
					t.Fatal(e)
				}
				return nil
			}
			v, err := d.Publish(ctx, publicationFixture(t))
			code, reason := Mismatch, "file-identity"
			switch name {
			case "mode":
				reason = "file-mode"
			case "extra":
				reason = "entry-set"
			case "manifest":
				reason = "baseline-mismatch"
			case "target-link":
				reason = "published-unreadable"
			case "cancel":
				code, reason = DestinationInvalid, "publication-interrupted"
			}
			requireDestinationError(t, err, code, reason)
			if v != nil {
				t.Fatal("mismatch returned success")
			}
			after, e := os.Lstat(path)
			if e != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
				t.Fatal("destructive rollback", e)
			}
			verified(t, filepath.Join(root, "control"))
		})
	}
}

// Cleanup observes only its recorded allocation; even a substituted top-level
// stage link is refused without following it. Same-user mutation *inside* the
// private directory is outside the contract and is deliberately not simulated.
func TestPublicationCleanupSubstitutedRoot(t *testing.T) {
	root, policy := destinationSandbox(t)
	publishControl(t, root, policy)
	d := admitDestination(t, filepath.Join(root, "target"), policy, DestinationAbsent)
	d.publicationHook = func(phase string, _ *os.File) error {
		if phase != "stage-created" {
			return nil
		}
		entries, e := os.ReadDir(root)
		if e != nil {
			t.Fatal(e)
		}
		for _, entry := range entries {
			if len(entry.Name()) > len(".curator-snapshot-") && entry.Name()[:len(".curator-snapshot-")] == ".curator-snapshot-" {
				stage := filepath.Join(root, entry.Name())
				if e := os.Rename(stage, stage+"-saved"); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(filepath.Join(root, "control"), stage); e != nil {
					t.Fatal(e)
				}
			}
		}
		return os.ErrPermission
	}
	v, err := d.Publish(context.Background(), publicationFixture(t))
	requireDestinationError(t, err, DestinationInvalid, "staging-io")
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatal("cleanup substitution not reported")
	}
	if v != nil {
		t.Fatal("failure returned success")
	}
	verified(t, filepath.Join(root, "control"))
	assertAbsent(t, filepath.Join(root, "target"))
}
