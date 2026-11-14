package envprofile

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/privatedir"
)

type staleDirectoryInfo struct {
	fs.FileInfo
}

func (staleDirectoryInfo) ModTime() time.Time { return time.Unix(1, 0) }

// Current handle metadata replaces a stale enumeration value. On each reported
// path the witness accepts reads and detects persistent bytes, a transient
// create/remove, explicit mtime changes and directory replacement. No path or
// directory timestamp is exempted. This is a fixture observer, not syscall proof.
func TestReadPurityDirectoryHandleWitness(t *testing.T) {
	for _, name := range []string{"manager/state/locks/v1", "manager/contexts/context", "profile-repos/example/.git/logs"} {
		t.Run(name, func(t *testing.T) {
			for _, plant := range []string{"none", "bytes", "transient", "mtime", "identity", "file-identity"} {
				t.Run(plant, func(t *testing.T) {
					root := t.TempDir()
					path := filepath.Join(root, filepath.FromSlash(name))
					if err := os.MkdirAll(path, 0700); err != nil {
						t.Fatal(err)
					}
					leaf := filepath.Join(path, "existing")
					if err := os.WriteFile(leaf, []byte("before"), 0600); err != nil {
						t.Fatal(err)
					}
					stamp := time.Unix(1700000000, 0)
					if err := os.Chtimes(path, stamp, stamp); err != nil {
						t.Fatal(err)
					}
					cached, err := os.Lstat(path)
					if err != nil {
						t.Fatal(err)
					}
					current, err := readWitnessInfo(path, staleDirectoryInfo{cached})
					if err != nil || !current.ModTime().Equal(stamp) {
						t.Fatalf("stale directory enumeration survived handle snapshot: %v", err)
					}
					before := readWitnesses(t, root)
					if err := checkReadWitnesses(before, readWitnesses(t, root)); err != nil {
						t.Fatalf("unchanged positive control: %v", err)
					}
					switch plant {
					case "none":
						if _, err := os.ReadDir(path); err != nil {
							t.Fatal(err)
						}
					case "bytes":
						if err := os.WriteFile(leaf, []byte("after!"), 0600); err != nil {
							t.Fatal(err)
						}
					case "transient":
						added := filepath.Join(path, "forbidden-lock")
						if err := os.WriteFile(added, []byte("write"), 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.Remove(added); err != nil {
							t.Fatal(err)
						}
					case "mtime":
						later := stamp.Add(time.Hour)
						if err := os.Chtimes(path, later, later); err != nil {
							t.Fatal(err)
						}
					case "identity":
						if err := os.Rename(path, filepath.Join(root, "old-directory")); err != nil {
							t.Fatal(err)
						}
						if err := os.Mkdir(path, 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.Chtimes(path, stamp, stamp); err != nil {
							t.Fatal(err)
						}
					case "file-identity":
						if err := os.Rename(leaf, filepath.Join(root, "old-file")); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(leaf, []byte("before"), 0600); err != nil {
							t.Fatal(err)
						}
						original := before[filepath.Join(filepath.FromSlash(name), "existing")].info.ModTime()
						if err := os.Chtimes(leaf, original, original); err != nil {
							t.Fatal(err)
						}
					}
					after := readWitnesses(t, root)
					// Isolate identity from the rename's incidental entry/parent
					// changes; the selected object retains its bytes and mtime.
					if plant == "identity" || plant == "file-identity" {
						selected := filepath.FromSlash(name)
						if plant == "file-identity" {
							selected = filepath.Join(selected, "existing")
						}
						before = map[string]readWitness{selected: before[selected]}
						after = map[string]readWitness{selected: after[selected]}
					}
					err = checkReadWitnesses(before, after)
					if plant == "none" {
						if err != nil {
							t.Fatalf("read-only positive control: %v", err)
						}
					} else if err == nil || !strings.Contains(err.Error(), "read-only witnesses changed") {
						t.Fatalf("%s plant escaped directory witness: %v", plant, err)
					}
				})
			}
			t.Logf("%s directory-handle-snapshot and write controls PASS", runtime.GOOS)
		})
	}
}

// The public keyring reader admits a private existing fixture and refuses an
// ordinary broader/inheriting directory without repairing its permissions.
func TestReadOnlyGPGKeyringPrivateBoundary(t *testing.T) {
	root := t.TempDir()
	private := filepath.Join(root, "private")
	if err := privatedir.Make(private); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(private, "pubring.gpg"), []byte("public fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GNUPGHOME", private)
	if home, ring, err := readOnlyGPGKeyring(); err != nil || home != private || ring != filepath.Join(private, "pubring.gpg") {
		t.Fatalf("private keyring positive control: %q %q %v", home, ring, err)
	}
	broader := filepath.Join(root, "broader")
	if err := os.Mkdir(broader, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(broader, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GNUPGHOME", broader)
	before := readWitnesses(t, root)
	if _, _, err := readOnlyGPGKeyring(); err == nil || !strings.Contains(err.Error(), "public keyring boundary is untrusted") {
		t.Fatalf("unsafe keyring boundary admitted: %v", err)
	}
	if err := privatedir.Validate(broader); err == nil {
		t.Fatal("read repaired the unsafe keyring boundary")
	}
	assertReadWitnesses(t, root, before)
	t.Logf("%s private/adversarial keyring boundary controls PASS", runtime.GOOS)
}
