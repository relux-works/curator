//go:build unix

package buildcache

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"golang.org/x/sys/unix"
)

func protectTestHome(t *testing.T, home string) {
	t.Helper()
	if err := os.Chmod(home, 0o700); err != nil {
		t.Fatal(err)
	}
}

func TestUnixProtectedStateMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, store *Store, hit Result)
	}{
		{
			name: "manager home group writable",
			mutate: func(t *testing.T, store *Store, _ Result) {
				if err := os.Chmod(store.Home(), 0o770); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "entry other writable",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				entry := filepath.Dir(filepath.Dir(hit.ArtifactPath))
				if err := os.Chmod(entry, 0o707); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "receipt group writable",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				receipt := filepath.Join(filepath.Dir(filepath.Dir(hit.ArtifactPath)), ReceiptFilename)
				if err := os.Chmod(receipt, 0o660); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact not executable",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				if err := os.Chmod(hit.ArtifactPath, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact hard link",
			mutate: func(t *testing.T, store *Store, hit Result) {
				if err := os.Link(hit.ArtifactPath, filepath.Join(store.Home(), "artifact-hardlink")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "receipt hard link",
			mutate: func(t *testing.T, store *Store, hit Result) {
				receipt := filepath.Join(filepath.Dir(filepath.Dir(hit.ArtifactPath)), ReceiptFilename)
				if err := os.Link(receipt, filepath.Join(store.Home(), "receipt-hardlink")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact symlink",
			mutate: func(t *testing.T, store *Store, hit Result) {
				if err := os.Remove(hit.ArtifactPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(store.Home(), "private-builds", "missing"), hit.ArtifactPath); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact is directory",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				if err := os.Remove(hit.ArtifactPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(hit.ArtifactPath, 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "artifact directory symlink",
			mutate: func(t *testing.T, store *Store, hit Result) {
				entry := filepath.Dir(filepath.Dir(hit.ArtifactPath))
				bin := filepath.Join(entry, "bin")
				external := filepath.Join(store.Home(), "moved-bin")
				if err := os.Rename(bin, external); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(external, bin); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "receipt special file",
			mutate: func(t *testing.T, _ *Store, hit Result) {
				receipt := filepath.Join(filepath.Dir(filepath.Dir(hit.ArtifactPath)), ReceiptFilename)
				if err := os.Remove(receipt); err != nil {
					t.Fatal(err)
				}
				if err := unix.Mkfifo(receipt, 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newTestStore(t)
			publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
			if _, err := store.Publish(publication, testHomeLock{}); err != nil {
				t.Fatal(err)
			}
			hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
			if hit.Status != Hit {
				t.Fatalf("initial inspection = %+v", hit)
			}
			test.mutate(t, store, hit)
			result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
			if result.Status != UntrustedProvenance {
				t.Fatalf("protected-state violation = %+v", result)
			}
		})
	}
}

func TestUnixForgedSelfConsistentEntryIsNeverAdopted(t *testing.T) {
	store := newTestStore(t)
	publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("attacker-chosen-artifact"))
	key, err := publication.Input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	entry, base, err := store.paths(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(entry, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(store.Home(), "cache"), filepath.Join(store.Home(), "cache", "build"), base} {
		if err := os.Chmod(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(entry, 0o777); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(entry, ReceiptFilename), publication.ReceiptBytes, 0o600)
	artifactRel, err := buildmetaArtifactRel(publication)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(entry, filepath.FromSlash(artifactRel)), []byte("attacker-chosen-artifact"), 0o700)

	before := treeFingerprint(t, store.Home())
	result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
	if result.Status != UntrustedProvenance || result.DryRunOutcome() != "would-rebuild-untrusted-cache" {
		t.Fatalf("forged inspection = %+v", result)
	}
	if after := treeFingerprint(t, store.Home()); after != before {
		t.Fatal("read-only forged-cache inspection mutated persistent state")
	}

	rebuilt, err := store.Publish(publication, testHomeLock{})
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt.Status != Published {
		t.Fatalf("rebuilt publication = %+v", rebuilt)
	}
	if hit := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash}); hit.Status != Hit {
		t.Fatalf("rebuilt inspection = %+v", hit)
	}
}

func TestUnixRejectsManagerHomeSymlinkBoundary(t *testing.T) {
	realHome := t.TempDir()
	protectTestHome(t, realHome)
	linkRoot := t.TempDir()
	linkHome := filepath.Join(linkRoot, "linked-home")
	if err := os.Symlink(realHome, linkHome); err != nil {
		t.Fatal(err)
	}
	store, err := New(linkHome)
	if err != nil {
		t.Fatal(err)
	}
	publication, receiptHash := testPublication(t, realHome, testInput("tool"), []byte("artifact"))
	realStore, err := New(realHome)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := realStore.Publish(publication, testHomeLock{}); err != nil {
		t.Fatal(err)
	}
	result := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash})
	if result.Status != UntrustedProvenance || !strings.Contains(result.Reason, "links") {
		t.Fatalf("symlink boundary = %+v", result)
	}
}

func TestUnixRejectsEntrySymlinkAndUnprotectedCacheRoot(t *testing.T) {
	t.Run("entry symlink", func(t *testing.T) {
		store := newTestStore(t)
		publication, receiptHash := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
		result, err := store.Publish(publication, testHomeLock{})
		if err != nil {
			t.Fatal(err)
		}
		entry := filepath.Dir(filepath.Dir(result.ArtifactPath))
		moved := filepath.Join(store.Home(), "moved-entry")
		if err := os.Rename(entry, moved); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(moved, entry); err != nil {
			t.Fatal(err)
		}
		if got := store.Inspect(Expectation{Input: publication.Input, ReceiptHash: receiptHash}); got.Status != UntrustedProvenance {
			t.Fatalf("entry symlink = %+v", got)
		}
	})

	t.Run("unprotected cache root", func(t *testing.T) {
		store := newTestStore(t)
		cacheRoot := filepath.Join(store.Home(), "cache")
		if err := os.Mkdir(cacheRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(cacheRoot, 0o777); err != nil {
			t.Fatal(err)
		}
		publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
		if _, err := store.Publish(publication, testHomeLock{}); err == nil {
			t.Fatal("publication repaired an unprotected cache root")
		}
		info, err := os.Stat(cacheRoot)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o777 {
			t.Fatalf("unprotected cache root was permission-repaired to %o", info.Mode().Perm())
		}
	})

	t.Run("source symlink", func(t *testing.T) {
		store := newTestStore(t)
		publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
		link := filepath.Join(store.Home(), "source-link")
		if err := os.Symlink(publication.ArtifactSource, link); err != nil {
			t.Fatal(err)
		}
		publication.ArtifactSource = link
		if _, err := store.Publish(publication, testHomeLock{}); err == nil {
			t.Fatal("publication accepted a symlink source")
		}
	})

	t.Run("entry is regular file", func(t *testing.T) {
		store := newTestStore(t)
		publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
		result, err := store.Publish(publication, testHomeLock{})
		if err != nil {
			t.Fatal(err)
		}
		entry := filepath.Dir(filepath.Dir(result.ArtifactPath))
		if err := os.RemoveAll(entry); err != nil {
			t.Fatal(err)
		}
		writeFile(t, entry, []byte("not a directory"), 0o600)
		if got := store.Inspect(Expectation{Input: publication.Input}); got.Status != UntrustedProvenance {
			t.Fatalf("regular entry = %+v", got)
		}
	})

	t.Run("bin is regular file", func(t *testing.T) {
		store := newTestStore(t)
		publication, _ := testPublication(t, store.Home(), testInput("tool"), []byte("artifact"))
		result, err := store.Publish(publication, testHomeLock{})
		if err != nil {
			t.Fatal(err)
		}
		bin := filepath.Dir(result.ArtifactPath)
		if err := os.RemoveAll(bin); err != nil {
			t.Fatal(err)
		}
		writeFile(t, bin, []byte("not a directory"), 0o600)
		if got := store.Inspect(Expectation{Input: publication.Input}); got.Status != UntrustedProvenance {
			t.Fatalf("regular bin = %+v", got)
		}
	})
}

func buildmetaArtifactRel(publication Publication) (string, error) {
	receipt, err := buildmeta.DecodeReceipt(publication.ReceiptBytes)
	if err != nil {
		return "", err
	}
	return receipt.Artifact.Path, nil
}

func TestUnixProtectionHelperFailuresFailClosed(t *testing.T) {
	root := t.TempDir()
	protectTestHome(t, root)
	missing := filepath.Join(root, "missing")
	if _, err := makeProtectedTempDir(missing, "stage-"); err == nil {
		t.Fatal("temporary staging under a missing root succeeded")
	}
	existingDir := filepath.Join(root, "existing")
	if err := os.Mkdir(existingDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := createProtectedDir(existingDir); err == nil {
		t.Fatal("protected directory creation replaced an existing path")
	}
	existingFile := filepath.Join(root, "existing-file")
	writeFile(t, existingFile, []byte("x"), 0o600)
	if err := writeProtectedFile(existingFile, 0o600, strings.NewReader("x")); err == nil {
		t.Fatal("protected file creation replaced an existing path")
	}
	failingFile := filepath.Join(root, "failing-file")
	if err := writeProtectedFile(failingFile, 0o600, errorReader{}); err == nil {
		t.Fatal("staged write ignored its source error")
	}
	if _, err := os.Lstat(failingFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed staged file survived: %v", err)
	}
	if err := syncDirectory(missing); err == nil {
		t.Fatal("sync of missing directory succeeded")
	}
	file, err := os.Open(existingFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if err := validateUnixDir(int(file.Fd()), "regular file"); !errors.Is(err, errUntrusted) {
		t.Fatalf("regular file validated as directory: %v", err)
	}
	dir, err := os.Open(existingDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = dir.Close() }()
	if err := validateUnixFile(int(dir.Fd()), "directory", false); !errors.Is(err, errUntrusted) {
		t.Fatalf("directory validated as regular file: %v", err)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
