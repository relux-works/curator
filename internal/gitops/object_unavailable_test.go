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
	for _, loss := range []string{"none", "blob", "commit", "tree", "repository"} {
		t.Run(loss, func(t *testing.T) {
			repo, commit := directoryRepo(t, nil)
			want, err := PinnedDirectorySnapshotPaths(repo, commit, "agents/skills/demo")
			if err != nil {
				t.Fatal(err)
			}
			oid := ""
			switch loss {
			case "blob":
				oid = gitRun(t, repo, "rev-parse", commit+":agents/skills/demo/SKILL.md")
			case "commit":
				oid = commit
			case "tree":
				oid = gitRun(t, repo, "rev-parse", commit+"^{tree}")
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
			if !errors.As(err, &unavailable) || paths != nil || unavailable.ObjectType == "" {
				t.Fatalf("object loss lacked typed availability: %v", err)
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
