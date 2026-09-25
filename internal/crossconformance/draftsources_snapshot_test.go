package crossconformance

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/snapshot"
)

// TestDraftSourcesSnapshotVectors drives the 3 pinned snapshot byte
// vectors through the production capture entry: the vendored
// utf8_files are materialized as a live package tree, admitted through
// PrepareLocalAcquisition, captured, and every per-file hash,
// executable flag, and inventory digest must equal the pinned values.
func TestDraftSourcesSnapshotVectors(t *testing.T) {
	vectors := loadDraftSnapshots(t)
	seenSkill := ""
	seenSnapshots := map[string]string{}
	conformancecoverage.Run(t, "draft-sources-v1/snapshot-cases", vectors,
		func(vector draftSnapshotVector) string { return vector.ID }, func(t *testing.T, vector draftSnapshotVector) {
			base := t.TempDir()
			pkg := filepath.Join(base, "pkg")
			materializeSnapshotVector(t, pkg, vector)
			project := filepath.Dir(pkg)
			acquisition, err := snapshot.PrepareLocalAcquisition(pkg, ".", project, "s", []string{filepath.Join(project, ".agents")}, nil, map[string]bool{"s": true}, "")
			if err != nil {
				t.Fatalf("PrepareLocalAcquisition: %v", err)
			}
			t.Cleanup(func() { _ = acquisition.Close() })
			inventory, err := snapshot.Capture(acquisition)
			if err != nil {
				t.Fatalf("Capture: %v", err)
			}
			if inventory.SchemaVersion != vector.Inventory.SchemaVersion || inventory.Algorithm != vector.Inventory.Algorithm {
				t.Fatalf("inventory header = %d/%s, want %d/%s", inventory.SchemaVersion, inventory.Algorithm, vector.Inventory.SchemaVersion, vector.Inventory.Algorithm)
			}
			if len(inventory.Files) != len(vector.Inventory.Files) {
				t.Fatalf("files = %d entries, want %d", len(inventory.Files), len(vector.Inventory.Files))
			}
			for i, entry := range inventory.Files {
				want := vector.Inventory.Files[i]
				if entry.Path != want.Path || entry.SHA256 != want.SHA256 {
					t.Fatalf("file %d = %s %s, want %s %s", i, entry.Path, entry.SHA256, want.Path, want.SHA256)
				}
				if entry.Executable != want.Executable {
					t.Fatalf("%s executable = %v, want %v", entry.Path, entry.Executable, want.Executable)
				}
			}
			if inventory.Snapshot != vector.Inventory.Snapshot {
				t.Fatalf("snapshot = %s, want %s", inventory.Snapshot, vector.Inventory.Snapshot)
			}
		})
	for _, vector := range vectors {
		for _, entry := range vector.Inventory.Files {
			if entry.Path == "SKILL.md" {
				if seenSkill == "" {
					seenSkill = entry.SHA256
				} else if entry.SHA256 != seenSkill {
					t.Errorf("%s changed SKILL.md bytes %s, want the frozen %s", vector.ID, entry.SHA256, seenSkill)
				}
			}
		}
		if prior, ok := seenSnapshots[vector.Inventory.Snapshot]; ok {
			t.Errorf("snapshot collision: %s and %s share %s", prior, vector.ID, vector.Inventory.Snapshot)
		}
		seenSnapshots[vector.Inventory.Snapshot] = vector.ID
	}
	if len(seenSnapshots) != len(vectors) {
		t.Fatalf("want %d distinct package identities, got %v", len(vectors), seenSnapshots)
	}
}

func materializeSnapshotVector(t *testing.T, pkg string, vector draftSnapshotVector) {
	t.Helper()
	wantExec := map[string]bool{}
	for _, entry := range vector.Inventory.Files {
		wantExec[entry.Path] = entry.Executable
	}
	for rel, content := range vector.UTF8Files {
		full := filepath.Join(pkg, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if wantExec[rel] {
			mode = 0o755
		}
		if err := os.Chmod(full, mode); err != nil {
			t.Fatal(err)
		}
	}
	// Same platform rule as the in-package vector test: the
	// executable bit is normative where the platform reports POSIX
	// execute bits, and deferred with the declared platform-control
	// reason where Windows synthesizes permission bits.
	info, err := os.Lstat(filepath.Join(pkg, "scripts", "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		if info.Mode().Perm()&0o111 == 0 {
			t.Skip("Windows does not expose portable executable permission bits; the normative executable vectors are asserted on the unix runners")
		}
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("scripts/run.sh lost its POSIX execute bit %o; normative executable vectors must hold on unix", info.Mode().Perm())
	}
}
