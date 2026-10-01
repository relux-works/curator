package scopes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/sourcelock"
)

const (
	refMarkerCommit  = "1111111111111111111111111111111111111111"
	refSourceCommit  = "2222222222222222222222222222222222222222"
	refProfileCommit = "3333333333333333333333333333333333333333"
	refProjectCommit = "4444444444444444444444444444444444444444"
)

func writeSourceLock(t *testing.T, root, name, commit string) {
	t.Helper()
	pkg, err := sourcelock.NetworkGitPackage("github.com/example/"+name, sourcelock.Commit{ObjectFormat: "sha1", Hex: commit}, ".")
	if err != nil {
		t.Fatal(err)
	}
	lock, err := sourcelock.New("sha256:"+strings.Repeat("a", 64), []sourcelock.Member{{
		Name: name, Directory: ".", Package: pkg, ContentSHA256: "sha256:" + strings.Repeat("b", 64),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(root), lock); err != nil {
		t.Fatal(err)
	}
}

func writeProfileLock(t *testing.T, home, profile, commit string) {
	t.Helper()
	dir := filepath.Join(home, "profiles", profile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	lock := &contextlock.Lock{Root: profile, Members: []contextlock.Member{{
		Kind: contextlock.KindContext, Name: profile, Source: "github.com/example/" + profile,
		Directory: "packages/root", Version: "1.0.0", Commit: commit, RequiredBy: []string{},
	}}}
	if _, err := contextlock.Write(filepath.Join(dir, "lock.json"), lock); err != nil {
		t.Fatal(err)
	}
}

func TestCollectSnapshotReferencesReadsMarkersAndLocks(t *testing.T) {
	home := t.TempDir()
	installMarker(t, filepath.Join(home, "global", "skills"), "skill-a", refMarkerCommit)
	writeSourceLock(t, filepath.Join(home, "global"), "skill-b", refSourceCommit)
	writeProfileLock(t, home, "root-context", refProfileCommit)
	project := t.TempDir()
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-c", strings.Repeat("5", 40))
	writeSourceLock(t, project, "skill-d", refProjectCommit)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	// A current-profile pointer is a file, not a profile, and holds no lock.
	if err := os.WriteFile(filepath.Join(home, "profiles", "current"), []byte("root-context\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	refs := CollectSnapshotReferences(home)
	if !refs.Certain() {
		t.Fatalf("reference set is uncertain: %q", refs.Uncertain)
	}
	for _, commit := range []string{refMarkerCommit, refSourceCommit, refProfileCommit, refProjectCommit, strings.Repeat("5", 40)} {
		if !refs.Commits[commit] {
			t.Errorf("commit %s is not referenced; have %v", commit, refs.Commits)
		}
	}
}

func TestCollectSnapshotReferencesIsUncertainForUntrustedRecords(t *testing.T) {
	for name, damage := range map[string]func(t *testing.T, home string){
		"invalid profile lock": func(t *testing.T, home string) {
			writeProfileLock(t, home, "root-context", refProfileCommit)
			if err := os.WriteFile(filepath.Join(home, "profiles", "root-context", "lock.json"), []byte("{"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"invalid source lock": func(t *testing.T, home string) {
			if err := os.MkdirAll(filepath.Join(home, "global"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sourcelock.PathIn(filepath.Join(home, "global")), []byte("[]"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"source lock is a directory": func(t *testing.T, home string) {
			if err := os.MkdirAll(sourcelock.PathIn(filepath.Join(home, "hybrid")), 0o755); err != nil {
				t.Fatal(err)
			}
		},
		"invalid install marker": func(t *testing.T, home string) {
			dir := filepath.Join(home, "global", "skills", "skill-a")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			installMarker(t, filepath.Join(home, "global", "skills"), "skill-a", refMarkerCommit)
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if err := os.WriteFile(filepath.Join(dir, entry.Name()), []byte("{"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		},
		"consumer registry is invalid": func(t *testing.T, home string) {
			if err := os.WriteFile(filepath.Join(home, ConsumersName), []byte("{"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			damage(t, home)
			if refs := CollectSnapshotReferences(home); refs.Certain() {
				t.Fatalf("reference set is certain over an untrusted record")
			}
		})
	}
}

func TestCollectSnapshotReferencesEmptyHomeIsCertain(t *testing.T) {
	refs := CollectSnapshotReferences(t.TempDir())
	if !refs.Certain() || len(refs.Commits) != 0 {
		t.Fatalf("empty home: certain=%v commits=%v uncertain=%q", refs.Certain(), refs.Commits, refs.Uncertain)
	}
}
