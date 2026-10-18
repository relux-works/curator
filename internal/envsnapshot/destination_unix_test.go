//go:build darwin || linux

package envsnapshot

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Admission rejects shared temporary ancestors, so these real syscall fixtures
// live in a caller-owned repository scratch allocation, outside protected roots.
func destinationSandbox(t *testing.T) (string, DestinationPolicy) {
	t.Helper()
	base, err := filepath.Abs("../../.temp")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		t.Fatal(err)
	}
	allocation, err := os.MkdirTemp(base, "destination-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(allocation); err != nil {
			t.Error(err)
		}
	})
	// The target containing a protected root also needs a private parent.
	root := filepath.Join(allocation, "fixture")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	protected := filepath.Join(root, "protected")
	if err := os.Mkdir(protected, 0700); err != nil {
		t.Fatal(err)
	}
	return root, DestinationPolicy{ProtectedRoots: []string{protected}}
}

func admitDestination(t *testing.T, path string, policy DestinationPolicy, kind DestinationKind) *Destination {
	t.Helper()
	d, err := OpenDestination(path, policy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Error(err)
		}
	})
	if d.Kind() != kind {
		t.Fatalf("kind = %s, want %s", d.Kind(), kind)
	}
	if err := d.Revalidate(); err != nil {
		t.Fatal(err)
	}
	return d
}

func requireDestinationError(t *testing.T, err error, code, reason string) {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) || e.Code != code || (reason != "" && e.Reason != reason) {
		t.Fatalf("refusal = %v, want %s/%s", err, code, reason)
	}
}

func snapshotAtDestination(t *testing.T, root string) {
	t.Helper()
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	m := manifest(t)
	for _, f := range m.Files {
		mode := os.FileMode(0600)
		if f.Mode == "0700" {
			mode = 0700
		}
		write(t, root, f.Path, read(t, "golden/"+f.Path), mode)
	}
	write(t, root, "manifest.json", read(t, "golden/manifest.json"), 0600)
	write(t, root, "launch.json", read(t, "golden/launch.json"), 0600)
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected absent destination, got %v", err)
	}
}

// OpenDestination admits absent private destinations and refuses an existing
// empty target without writes. Close is idempotent and revokes observations.
func TestDestinationReadOnlyAdmission(t *testing.T) {
	var zero Destination
	requireDestinationError(t, zero.Revalidate(), DestinationInvalid, "destination-closed")
	root, policy := destinationSandbox(t)
	absent := filepath.Join(root, "absent")
	d := admitDestination(t, absent, policy, DestinationAbsent)
	if yes, err := d.MatchPrior(read(t, "golden/manifest.json")); err != nil || yes {
		t.Fatal("absent target reported prior snapshot", yes, err)
	}
	assertAbsent(t, absent)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	admitDestination(t, relative+"/relative", policy, DestinationAbsent)
	assertAbsent(t, filepath.Join(root, "relative"))
	for _, invalid := range []string{"", "/", "\x00", root + "/.", root + "/.."} {
		_, err := OpenDestination(invalid, policy)
		requireDestinationError(t, err, DestinationInvalid, "path-invalid")
	}
	empty := filepath.Join(root, "empty")
	if err := os.Mkdir(empty, 0700); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(empty)
	if err != nil {
		t.Fatal(err)
	}
	_, err = OpenDestination(empty, policy)
	requireDestinationError(t, err, DestinationInvalid, "empty-target")
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	requireDestinationError(t, d.Revalidate(), DestinationInvalid, "destination-closed")
	after, err := os.Stat(empty)
	if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) || before.Mode() != after.Mode() {
		t.Fatal("admission changed empty target", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 2 {
		t.Fatal("admission wrote scratch entries", entries, err)
	}
}

// Every supplied component is no-follow, including a linked ancestor before
// '..'. Invalid reads and unrelated content refuse without altering sentinels.
func TestDestinationBoundary(t *testing.T) {
	for _, name := range []string{"linked-ancestor", "linked-before-parent", "linked-target", "file-target", "missing-parent", "shared-ancestor", "shared-target", "public-target", "unrelated", "missing-protected", "no-policy"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			admitDestination(t, filepath.Join(root, "control"), policy, DestinationAbsent)
			safe := filepath.Join(root, "safe")
			if err := os.Mkdir(safe, 0700); err != nil {
				t.Fatal(err)
			}
			write(t, safe, "sentinel", []byte("preserved"), 0600)
			path := filepath.Join(root, "target")
			switch name {
			case "linked-ancestor", "linked-before-parent":
				if err := os.Symlink(safe, path); err != nil {
					t.Fatal(err)
				}
				if name == "linked-before-parent" {
					path += "/../new"
				} else {
					path += "/new"
				}
			case "linked-target":
				if err := os.Symlink(safe, path); err != nil {
					t.Fatal(err)
				}
			case "file-target", "unrelated":
				if name == "unrelated" {
					write(t, path, "foreign", []byte("preserved"), 0600)
				} else if err := os.WriteFile(path, []byte("preserved"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing-parent":
				path += "/child"
			case "shared-ancestor", "shared-target", "public-target":
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
				mode := os.FileMode(0770)
				if name == "public-target" {
					mode = 0755
				}
				if err := os.Chmod(path, mode); err != nil {
					t.Fatal(err)
				}
				if name == "shared-ancestor" {
					// Separate ancestor permissions from the immediate parent's
					// additional private-mode requirement.
					private := filepath.Join(path, "private")
					if err := os.Mkdir(private, 0700); err != nil {
						t.Fatal(err)
					}
					path = filepath.Join(private, "child")
				}
			case "missing-protected":
				policy.ProtectedRoots = append(policy.ProtectedRoots, filepath.Join(root, "missing"))
			case "no-policy":
				policy.ProtectedRoots = nil
			}
			_, err := OpenDestination(path, policy)
			requireDestinationError(t, err, DestinationInvalid, "")
			b, err := os.ReadFile(filepath.Join(safe, "sentinel"))
			if err != nil || string(b) != "preserved" {
				t.Fatal("outside sentinel changed", err)
			}
			assertAbsent(t, filepath.Join(safe, "new"))
			assertAbsent(t, filepath.Join(root, "new"))
		})
	}
}

// Root-owned system parents are not caller-owned publication parents for an
// unprivileged caller. A closed retained handle is a read failure, not absence.
func TestDestinationOwnershipAndReadFailure(t *testing.T) {
	root, policy := destinationSandbox(t)
	d := admitDestination(t, filepath.Join(root, "control"), policy, DestinationAbsent)
	if os.Geteuid() == 0 {
		t.Log("foreign-owner control unrun: effective root owns system parent")
	} else {
		_, err := OpenDestination("/snapshot-owner-control", policy)
		requireDestinationError(t, err, DestinationInvalid, "directory-owner")
		assertAbsent(t, "/snapshot-owner-control")
	}
	if err := d.chain[len(d.chain)-1].file.Close(); err != nil {
		t.Fatal(err)
	}
	requireDestinationError(t, d.Revalidate(), DestinationInvalid, "directory-unreadable")
	assertAbsent(t, filepath.Join(root, "control"))
}

// Handle identity excludes equal, descendant and containing destinations for
// every protected-root class. Similar lexical prefixes remain valid siblings.
func TestDestinationProtectedIdentity(t *testing.T) {
	for _, class := range []string{"manager", "store", "managed-home", "native-home", "credentials", "repository-admin", "workspace"} {
		t.Run(class, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			protected := filepath.Join(root, class)
			// An intact snapshot avoids masking a broken protected-identity
			// gate with the unrelated/empty-target refusal.
			snapshotAtDestination(t, protected)
			policy.ProtectedRoots = append(policy.ProtectedRoots, protected)
			admitDestination(t, protected+"-sibling", policy, DestinationAbsent)
			for _, path := range []string{protected, protected + "/child", root} {
				_, err := OpenDestination(path, policy)
				requireDestinationError(t, err, DestinationInvalid, "protected-destination")
			}
			upper := filepath.Join(root, strings.ToUpper(class))
			if _, err := os.Stat(upper); err == nil {
				_, err := OpenDestination(upper, policy)
				requireDestinationError(t, err, DestinationInvalid, "protected-destination")
			} // A folding filesystem exercises a physical alias; Linux does not.
			assertAbsent(t, protected+"/child")
		})
	}
}

// Kernel directory substitutions after admission refuse through Revalidate;
// neither moved originals nor competing targets/links are cleaned or replaced.
func TestDestinationReplacementRaces(t *testing.T) {
	for _, name := range []string{"ancestor-link", "ancestor-directory", "target-link", "target-directory", "target-empty", "target-identical", "competing-target", "empty-filled", "permission-change", "protected-moved"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			parent := filepath.Join(root, "parent")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(parent, "target")
			kind := DestinationAbsent
			if name != "competing-target" {
				snapshotAtDestination(t, path)
				kind = DestinationSnapshot
			}
			d := admitDestination(t, path, policy, kind)
			sentinel := filepath.Join(root, "outside")
			write(t, sentinel, "sentinel", []byte("preserved"), 0600)
			switch name {
			case "ancestor-link", "ancestor-directory", "target-link", "target-directory", "target-empty", "target-identical":
				replaced := path
				if strings.HasPrefix(name, "ancestor") {
					replaced = parent
				}
				if err := os.Rename(replaced, replaced+"-original"); err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(name, "link") {
					if err := os.Symlink(sentinel, replaced); err != nil {
						t.Fatal(err)
					}
				} else if name == "target-identical" {
					snapshotAtDestination(t, replaced)
				} else if name == "target-empty" {
					if err := os.Mkdir(replaced, 0700); err != nil {
						t.Fatal(err)
					}
				} else {
					write(t, replaced, "competitor", []byte("preserved"), 0600)
				}
			case "competing-target", "empty-filled":
				write(t, path, "competitor", []byte("preserved"), 0600)
			case "permission-change":
				if err := os.Chmod(parent, 0770); err != nil {
					t.Fatal(err)
				}
			case "protected-moved":
				if err := os.Rename(policy.ProtectedRoots[0], policy.ProtectedRoots[0]+"-original"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(policy.ProtectedRoots[0], 0700); err != nil {
					t.Fatal(err)
				}
			}
			code := DestinationInvalid
			if name == "empty-filled" {
				code = Mismatch
			}
			requireDestinationError(t, d.Revalidate(), code, "")
			if err := d.Close(); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(sentinel, "sentinel"))
			if err != nil || string(b) != "preserved" {
				t.Fatal("outside content changed", err)
			}
			if strings.HasSuffix(name, "directory") || name == "competing-target" || name == "empty-filled" {
				p := path
				if name == "ancestor-directory" {
					p = parent
				}
				b, err := os.ReadFile(filepath.Join(p, "competitor"))
				if err != nil || string(b) != "preserved" {
					t.Fatal("competing target changed", err)
				}
			}
		})
	}
}

// A private parent admits an absent destination; making that same retained
// parent readable by others refuses both fresh admission and revalidation.
func TestDestinationPrivateParent(t *testing.T) {
	root, policy := destinationSandbox(t)
	parent := filepath.Join(root, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "target")
	d := admitDestination(t, path, policy, DestinationAbsent)
	if err := os.Chmod(parent, 0755); err != nil {
		t.Fatal(err)
	}
	requireDestinationError(t, d.Revalidate(), DestinationInvalid, "directory-permissions")
	_, err := OpenDestination(path, policy)
	requireDestinationError(t, err, DestinationInvalid, "directory-permissions")
	assertAbsent(t, path)
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatal("refusal wrote into public parent", err)
	}
}

// Intact exact generations match without writes. Actual byte/mode/entry/version
// drift refuses despite unchanged saved manifests; different source binding is
// not portable idempotence. Both initial admission and revalidation exercise it.
func TestDestinationPriorIdentity(t *testing.T) {
	for _, name := range []string{"identical", "bytes", "mode", "extra", "unreadable", "unknown-version", "new-generation"} {
		t.Run(name, func(t *testing.T) {
			root, policy := destinationSandbox(t)
			path := filepath.Join(root, "snapshot")
			snapshotAtDestination(t, path)
			d := admitDestination(t, path, policy, DestinationSnapshot)
			original := read(t, "golden/manifest.json")
			requested := bytes.Clone(original)
			reason := ""
			switch name {
			case "bytes":
				p := filepath.Join(path, "home/AGENTS.md")
				b, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				b[0] ^= 1
				write(t, path, "home/AGENTS.md", b, 0600)
				reason = "file-identity"
			case "mode", "unreadable":
				mode := os.FileMode(0700)
				if name == "unreadable" {
					mode = 0000
				}
				if err := os.Chmod(filepath.Join(path, "home/AGENTS.md"), mode); err != nil {
					t.Fatal(err)
				}
				reason = "file-mode"
				if name == "unreadable" {
					reason = "file-unreadable"
				}
			case "extra":
				write(t, path, "extra", []byte("preserved"), 0600)
				reason = "entry-set"
			case "unknown-version":
				write(t, path, "manifest.json", bytes.Replace(original, []byte(ManifestVersion), []byte("managed-home-snapshot-v0"), 1), 0600)
			case "new-generation":
				requested = read(t, "golden/other-source-manifest.json")
				reason = "generation-mismatch"
			}
			before, err := os.ReadFile(filepath.Join(path, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			beforeInfo, err := os.Stat(filepath.Join(path, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			match, err := d.MatchPrior(requested)
			if name == "identical" {
				if !match || err != nil {
					t.Fatal("intact prior rejected", err)
				}
			} else {
				requireDestinationError(t, err, Mismatch, reason)
				if match {
					t.Fatal("mismatched snapshot admitted")
				}
				if name != "new-generation" {
					_, err := OpenDestination(path, policy)
					requireDestinationError(t, err, Mismatch, reason)
				}
			}
			after, err := os.ReadFile(filepath.Join(path, "manifest.json"))
			afterInfo, statErr := os.Stat(filepath.Join(path, "manifest.json"))
			if err != nil || statErr != nil || !bytes.Equal(before, after) || !os.SameFile(beforeInfo, afterInfo) || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
				t.Fatal("prior manifest changed during admission", err, statErr)
			}
		})
	}
}

// Once admitted as a prior snapshot, lost metadata is a snapshot mismatch,
// rather than unrelated content or an absent target. No metadata is recreated.
func TestDestinationPriorMissingMetadata(t *testing.T) {
	root, policy := destinationSandbox(t)
	path := filepath.Join(root, "snapshot")
	snapshotAtDestination(t, path)
	d := admitDestination(t, path, policy, DestinationSnapshot)
	manifestPath := filepath.Join(path, "manifest.json")
	if err := os.Rename(manifestPath, manifestPath+".saved"); err != nil {
		t.Fatal(err)
	}
	_, err := d.MatchPrior(read(t, "golden/manifest.json"))
	requireDestinationError(t, err, Mismatch, "prior-unreadable")
	assertAbsent(t, manifestPath)
	b, err := os.ReadFile(manifestPath + ".saved")
	if err != nil || !bytes.Equal(b, read(t, "golden/manifest.json")) {
		t.Fatal("refusal rewrote moved metadata", err)
	}
}
