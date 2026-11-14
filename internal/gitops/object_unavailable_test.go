package gitops

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The roster authenticates paths from trees without requiring blob payloads.
// Missing commits, trees and repositories are typed availability, with Git's
// exit code retained; unsupported entry metadata remains a structural error.
func TestPinnedPathsObjectAvailability(t *testing.T) {
	for _, loss := range []string{"none", "blob", "commit", "tree", "selected-tree", "ancestor-tree", "repository", "head-missing", "objects-missing"} {
		t.Run(loss, func(t *testing.T) {
			repo, commit := directoryRepo(t, nil)
			want, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo")
			if err != nil {
				t.Fatal(err)
			}
			oid := ""
			wantType := "commit"
			switch loss {
			case "blob":
				oid = gitRun(t, repo, "rev-parse", commit+":agents/skills/demo/SKILL.md")
			case "commit":
				oid = commit
			case "tree":
				wantType = "tree"
				oid = gitRun(t, repo, "rev-parse", commit+"^{tree}")
			case "selected-tree", "ancestor-tree":
				wantType = "tree"
				selected := "agents/skills/demo"
				if loss == "ancestor-tree" {
					selected = "agents"
				}
				oid = gitRun(t, repo, "rev-parse", commit+":"+selected)
			case "head-missing", "objects-missing":
				wantType = "repository"
				name := "HEAD"
				if loss == "objects-missing" {
					name = "objects"
				}
				if err := os.RemoveAll(filepath.Join(repo, ".git", name)); err != nil {
					t.Fatal(err)
				}
			case "repository":
				if err := os.RemoveAll(repo); err != nil {
					t.Fatal(err)
				}
			}
			if oid != "" {
				path := filepath.Join(repo, ".git", "objects", oid[:2], oid[2:])
				if err := os.Chmod(path, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			paths, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo")
			if loss == "none" || loss == "blob" {
				if err != nil || !reflect.DeepEqual(paths, want) {
					t.Fatalf("metadata roster required blob bytes: %v %v", paths, err)
				}
				return
			}
			var unavailable *ObjectUnavailableError
			if !errors.As(err, &unavailable) || paths != nil || unavailable.ObjectType != wantType {
				t.Fatalf("object loss lacked typed availability %s: %v", wantType, err)
			}
			if loss != "repository" && unavailable.ExitCode <= 0 {
				t.Fatalf("Git failure lost exit code: %+v", unavailable)
			}
		})
	}
	repo, commit := directoryRepo(t, map[string]string{"link": "outside"})
	_, err := PinnedDirectorySnapshotPaths(repo, commit, "")
	var unavailable *ObjectUnavailableError
	if err == nil || errors.As(err, &unavailable) {
		t.Fatalf("unsupported metadata became availability: %v", err)
	}
}

// Direct commit-tree readers classify invalid repositories before resolving the
// commit. The same fixture succeeds immediately before each metadata removal.
func TestPinnedCommitTreeRepositoryAvailability(t *testing.T) {
	for _, name := range []string{"HEAD", "objects"} {
		t.Run(name, func(t *testing.T) {
			repo, commit := directoryRepo(t, nil)
			tree, format, err := PinnedCommitTree(repo, commit)
			if err != nil || tree == "" || format != "sha1" {
				t.Fatalf("intact repository neighbor: %v", err)
			}
			if err := os.RemoveAll(filepath.Join(repo, ".git", name)); err != nil {
				t.Fatal(err)
			}
			tree, format, err = PinnedCommitTree(repo, commit)
			var unavailable *ObjectUnavailableError
			if !errors.As(err, &unavailable) || unavailable.ObjectType != "repository" || unavailable.ExitCode <= 0 || tree != "" || format != "" {
				t.Fatalf("direct repository loss lacked typed availability repository: %v", err)
			}
		})
	}
}

// Each control removes exactly one typed return and must fail at the typed
// assertion, after its intact fixture neighbor has already succeeded.
func TestPinnedObjectAvailabilityControls(t *testing.T) {
	for _, plant := range []struct{ name, file, before, after, test, witness string }{
		{"untyped-ls-tree", "snapshot_paths.go", "return nil, objectUnavailable(\"tree\", err)", "return nil, err", "TestPinnedPathsObjectAvailability/selected-tree", "object loss lacked typed availability tree"},
		{"untyped-dirtree", "directory.go", "return \"\", objectUnavailable(\"tree\", err)", "return \"\", err", "TestPinnedPathsObjectAvailability/ancestor-tree", "object loss lacked typed availability tree"},
		{"untyped-repository", "treehash.go", "return \"\", \"\", objectUnavailable(\"repository\", err)", "return \"\", \"\", err", "TestPinnedCommitTreeRepositoryAvailability/HEAD", "direct repository loss lacked typed availability repository"},
	} {
		t.Run(plant.name, func(t *testing.T) {
			runGitObjectControl(t, plant.file, plant.before, plant.after, plant.test, plant.witness)
		})
	}
}
