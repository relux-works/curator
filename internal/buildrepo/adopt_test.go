package buildrepo

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// transactionCopy recreates a staged entry the way the install transaction
// publishes it (internal/transaction/staging.go createStagingEntry +
// copyStagingFile): every directory and file is a fresh object created with
// mode 0 and chmod-ed to the source mode, then the bytes are appended. This is
// exactly the object shape a protected lookup meets after the commit — modes
// preserved on unix, DACL inherited from the parent on Windows.
func transactionCopy(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if err := os.Mkdir(target, 0); err != nil {
				return err
			}
			return os.Chmod(target, info.Mode().Perm())
		}
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0)
		if err != nil {
			return err
		}
		if err := file.Chmod(info.Mode().Perm()); err != nil {
			_ = file.Close()
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			_ = file.Close()
			return err
		}
		_, copyErr := io.Copy(file, in)
		_ = in.Close()
		if closeErr := file.Close(); copyErr == nil {
			copyErr = closeErr
		}
		return copyErr
	})
	if err != nil {
		t.Fatal(err)
	}
}

func adoptionArtifactInput(version int) map[string]any {
	build := map[string]any{"command": "tool", "target": map[string]any{"goos": runtime.GOOS}}
	if version == LegacyReceiptSchemaVersion {
		return build
	}
	return map[string]any{"schema_version": SourceAwareReceiptSchemaVersion, "package": testPackage().Object(), "build": build}
}

// stagedArtifactEntry publishes one artifact into an operation-private staging
// store and returns the store, the derived key, the input and the entry path.
func stagedArtifactEntry(t *testing.T, version int) (*DiskProtectedStore, string, map[string]any, string) {
	t.Helper()
	input := adoptionArtifactInput(version)
	key, err := cacheKey(input)
	if err != nil {
		t.Fatal(err)
	}
	staging := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "staging")}
	if err := staging.PrepareNamespaces(version); err != nil {
		t.Fatal(err)
	}
	artifact := []byte("adopted artifact " + ArtifactsDir(version))
	if _, err := staging.StoreArtifact(key, input, "tool", artifact, testExecutionReceiptBytes(t, driverInputOf(input), artifact)); err != nil {
		t.Fatal(err)
	}
	return staging, key, input, filepath.Join(staging.Root, ArtifactsDir(version), strings.TrimPrefix(key, "sha256:"))
}

// TestAdoptArtifactMakesATransactionCopyACleanHit: an artifact entry that the
// transaction recreated below a prepared final namespace is a clean exact hit
// after AdoptArtifact, in both receipt namespaces. On Windows the copy is
// refused BEFORE adoption (inherited DACL), which is the failure the hosted
// gate reported; on unix the mode-only copy already proves and adoption is
// the walk-only mode. Either way the post-adoption lookup is the proof.
func TestAdoptArtifactMakesATransactionCopyACleanHit(t *testing.T) {
	for _, version := range []int{LegacyReceiptSchemaVersion, SourceAwareReceiptSchemaVersion} {
		t.Run(ArtifactsDir(version), func(t *testing.T) {
			_, key, input, stagedEntry := stagedArtifactEntry(t, version)
			final := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "final")}
			if err := final.PrepareNamespaces(version); err != nil {
				t.Fatal(err)
			}
			finalEntry := filepath.Join(final.Root, ArtifactsDir(version), strings.TrimPrefix(key, "sha256:"))
			transactionCopy(t, stagedEntry, finalEntry)

			before, err := final.LookupArtifact(key, input, false)
			if runtime.GOOS == "windows" {
				if err == nil || !strings.Contains(err.Error(), "DACL") {
					t.Fatalf("Windows lookup of the unadopted transaction copy = %v, %v; want the inherited-DACL refusal", before, err)
				}
			} else if err != nil || before == nil {
				t.Fatalf("unix lookup of the transaction copy = %v, %v; want a hit from the preserved modes", before, err)
			}
			if err := final.AdoptArtifact(version, key); err != nil {
				t.Fatalf("adoption: %v", err)
			}
			hit, err := final.LookupArtifact(key, input, false)
			if err != nil || hit == nil || string(hit.Bytes) != "adopted artifact "+ArtifactsDir(version) {
				t.Fatalf("lookup after adoption = %v, %v", hit, err)
			}
			if err := final.AdoptArtifact(version, key); err != nil {
				t.Fatalf("second adoption is not idempotent: %v", err)
			}
			// The entry was adopted in place: no quarantine, no extra objects.
			if _, err := os.Lstat(filepath.Join(final.Root, "quarantine")); !os.IsNotExist(err) {
				t.Fatalf("adoption quarantined or created objects: %v", err)
			}
		})
	}
}

// TestAdoptArtifactRefusesWhatItCannotProve: every refusal of adoption is an
// error and never a write — no entry, namespace or parent is created and the
// evidence stays in place.
func TestAdoptArtifactRefusesWhatItCannotProve(t *testing.T) {
	version := SourceAwareReceiptSchemaVersion
	_, key, _, stagedEntry := stagedArtifactEntry(t, version)
	name := strings.TrimPrefix(key, "sha256:")
	newFinal := func(t *testing.T, prepare bool) *DiskProtectedStore {
		t.Helper()
		final := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "final")}
		if prepare {
			if err := final.PrepareNamespaces(version); err != nil {
				t.Fatal(err)
			}
		}
		return final
	}
	rows := []struct {
		name    string
		arrange func(t *testing.T) (*DiskProtectedStore, int, string)
		want    string
	}{
		{"unknown schema version", func(t *testing.T) (*DiskProtectedStore, int, string) {
			return newFinal(t, true), 4, key
		}, "unknown receipt schema version"},
		{"malformed key", func(t *testing.T) (*DiskProtectedStore, int, string) {
			return newFinal(t, true), version, "sha256:zz"
		}, "invalid protected key"},
		{"absent store root", func(t *testing.T) (*DiskProtectedStore, int, string) {
			return newFinal(t, false), version, key
		}, "store root"},
		{"unprepared namespace", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, false)
			if err := final.PrepareNamespaces(LegacyReceiptSchemaVersion); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, "namespace " + ArtifactsDir(version)},
		{"absent entry", func(t *testing.T) (*DiskProtectedStore, int, string) {
			return newFinal(t, true), version, key
		}, name},
		{"entry is a file", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			if err := os.WriteFile(filepath.Join(final.Root, ArtifactsDir(version), name), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, "not a directory"},
		{"entry copied under another key", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			other := "sha256:" + strings.Repeat("0", 64)
			transactionCopy(t, stagedEntry, filepath.Join(final.Root, ArtifactsDir(version), strings.TrimPrefix(other, "sha256:")))
			return final, version, other
		}, "does not derive the entry key"},
		{"entry copied into the other namespace", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			if err := final.PrepareNamespaces(LegacyReceiptSchemaVersion); err != nil {
				t.Fatal(err)
			}
			transactionCopy(t, stagedEntry, filepath.Join(final.Root, ArtifactsDir(LegacyReceiptSchemaVersion), name))
			return final, LegacyReceiptSchemaVersion, key
		}, "does not derive the entry key"},
		{"tampered artifact bytes", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			entry := filepath.Join(final.Root, ArtifactsDir(version), name)
			transactionCopy(t, stagedEntry, entry)
			if err := os.WriteFile(filepath.Join(entry, "artifact"), []byte("forged"), 0o700); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, CodeArtifactInvalid},
		{"tampered receipt", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			entry := filepath.Join(final.Root, ArtifactsDir(version), name)
			transactionCopy(t, stagedEntry, entry)
			if err := os.WriteFile(filepath.Join(entry, "receipt.json"), []byte(`{"input":{}}`), 0o600); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, "does not derive the entry key"},
		{"missing execution receipt", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			entry := filepath.Join(final.Root, ArtifactsDir(version), name)
			transactionCopy(t, stagedEntry, entry)
			if err := os.Remove(filepath.Join(entry, "execution-receipt.ccj.json")); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, CodeReceiptInvalid},
	}
	if runtime.GOOS != "windows" {
		rows = append(rows, struct {
			name    string
			arrange func(t *testing.T) (*DiskProtectedStore, int, string)
			want    string
		}{"symbolic link inside the entry", func(t *testing.T) (*DiskProtectedStore, int, string) {
			final := newFinal(t, true)
			entry := filepath.Join(final.Root, ArtifactsDir(version), name)
			transactionCopy(t, stagedEntry, entry)
			if err := os.Symlink(filepath.Join(entry, "artifact"), filepath.Join(entry, "alias")); err != nil {
				t.Fatal(err)
			}
			return final, version, key
		}, "non-regular entry"})
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			final, adoptVersion, adoptKey := row.arrange(t)
			err := final.AdoptArtifact(adoptVersion, adoptKey)
			if err == nil || !strings.Contains(err.Error(), row.want) {
				t.Fatalf("adoption = %v, want an error containing %q", err, row.want)
			}
			if _, statErr := os.Lstat(filepath.Join(final.Root, "quarantine")); !os.IsNotExist(statErr) {
				t.Fatalf("a refused adoption quarantined the evidence: %v", statErr)
			}
		})
	}
	// A refusal creates nothing: the absent-store row leaves no root behind.
	final := newFinal(t, false)
	if err := final.AdoptArtifact(version, key); err == nil {
		t.Fatal("adoption into an absent store succeeded")
	}
	if _, err := os.Lstat(final.Root); !os.IsNotExist(err) {
		t.Fatalf("a refused adoption created the store root: %v", err)
	}
}

// TestAdoptSnapshotMakesATransactionCopyACleanHit is the snapshot arm of the
// same property, plus its refusals: absent entry and tampered file bytes.
func TestAdoptSnapshotMakesATransactionCopyACleanHit(t *testing.T) {
	snapshot, _, effective := pipelineFixture(t)
	key, err := SnapshotKey(effective, snapshot.Digest)
	if err != nil {
		t.Fatal(err)
	}
	name := strings.TrimPrefix(key, "sha256:")
	staging := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "staging")}
	if err := staging.PrepareNamespaces(); err != nil {
		t.Fatal(err)
	}
	if err := staging.StoreSnapshot(key, snapshot); err != nil {
		t.Fatal(err)
	}
	final := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "final")}
	if err := final.AdoptSnapshot(key); err == nil {
		t.Fatal("adoption into an absent store succeeded")
	}
	if _, err := os.Lstat(final.Root); !os.IsNotExist(err) {
		t.Fatalf("a refused adoption created the store root: %v", err)
	}
	if err := final.PrepareNamespaces(); err != nil {
		t.Fatal(err)
	}
	if err := final.AdoptSnapshot(key); err == nil || !strings.Contains(err.Error(), name) {
		t.Fatalf("absent snapshot adopted: %v", err)
	}
	transactionCopy(t, filepath.Join(staging.Root, "snapshots", name), filepath.Join(final.Root, "snapshots", name))
	before, err := final.LoadSnapshot(key, false)
	if runtime.GOOS == "windows" {
		if err == nil || !strings.Contains(err.Error(), "DACL") {
			t.Fatalf("Windows load of the unadopted transaction copy = %v, %v; want the inherited-DACL refusal", before, err)
		}
	} else if err != nil || before == nil {
		t.Fatalf("unix load of the transaction copy = %v, %v", before, err)
	}
	if err := final.AdoptSnapshot(key); err != nil {
		t.Fatalf("adoption: %v", err)
	}
	loaded, err := final.LoadSnapshot(key, false)
	if err != nil || loaded == nil || loaded.Digest != snapshot.Digest || loaded.Commit != snapshot.Commit {
		t.Fatalf("load after adoption = %+v, %v", loaded, err)
	}
	// Tampered bytes are refused by the proof, and adoption never repairs or
	// quarantines them.
	files := filepath.Join(final.Root, "snapshots", name, "files")
	var first string
	_ = filepath.WalkDir(files, func(path string, entry os.DirEntry, _ error) error {
		if first == "" && !entry.IsDir() {
			first = path
		}
		return nil
	})
	if first == "" {
		t.Fatal("snapshot has no files")
	}
	if err := os.WriteFile(first, []byte("forged"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := final.AdoptSnapshot(key); err == nil || !strings.Contains(err.Error(), CodeObjectSemanticsInvalid) {
		t.Fatalf("tampered snapshot adopted: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(final.Root, "quarantine")); !os.IsNotExist(err) {
		t.Fatalf("a refused adoption quarantined the evidence: %v", err)
	}
}
