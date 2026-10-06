package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

func TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled(t *testing.T) {
	// The opaque-NUL interim rule (Spec §8) applies exactly to v1
	// identities: pin the v1 writers so this matrix exercises the v1
	// reader. The v2 production entry lives in nul_opaque_v2_test.go.
	priorWriter := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = priorWriter })
	type lane struct {
		name    string
		install func(*testing.T, *env) Result
		target  func(*env) string
	}
	lanes := []lane{
		{
			name: "project",
			install: func(_ *testing.T, e *env) Result {
				e.declare("skill-a")
				return e.install(Options{})
			},
			target: func(e *env) string { return filepath.Join(e.project, ".agents", "skills", "skill-a") },
		},
		{
			name: "global",
			install: func(t *testing.T, e *env) Result {
				t.Helper()
				if _, err := GlobalInit(e.home); err != nil {
					t.Fatal(err)
				}
				if err := manifestAddGlobal(e, "skill-a"); err != nil {
					t.Fatal(err)
				}
				return Global(e.cfg, t.TempDir(), Options{Platform: installPlatform()})
			},
			target: func(e *env) string { return filepath.Join(GlobalRoot(e.home), "skills", "skill-a") },
		},
	}
	cases := []struct {
		name        string
		files       map[string]string
		wantFinding string
	}{
		{
			name: "single-file-collision",
			files: map[string]string{
				"assets/a.bin": "x\x00docs/b.md\x00y\x00z",
			},
			wantFinding: "assets/a.bin",
		},
		{
			name: "split-file-collision",
			files: map[string]string{
				"assets/a.bin": "x",
				"docs/b.md":    "y\x00z",
			},
			wantFinding: "docs/b.md",
		},
		{
			name: "deep-docs-opaque-file",
			files: map[string]string{
				"docs/deep/opaque.unsupported": "prefix\x00suffix",
			},
			wantFinding: "docs/deep/opaque.unsupported",
		},
		{name: "nul-free-control"},
	}

	first, second := t.TempDir(), t.TempDir()
	writeNULTree(t, first, cases[0].files)
	writeNULTree(t, second, cases[1].files)
	firstHash, err := hashing.ContentSHA256(first, nil)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := hashing.ContentSHA256(second, nil)
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash {
		t.Fatalf("v1 framing did not collide: %s != %s", firstHash, secondHash)
	}

	for _, installLane := range lanes {
		for _, tc := range cases {
			t.Run(installLane.name+"/"+tc.name, func(t *testing.T) {
				e := newEnv(t)
				e.skill("skill-a")
				if len(tc.files) != 0 {
					addSkillFilesAndRetag(t, e, "skill-a", tc.files)
				}
				if e.cfg.Audit.Enabled {
					t.Fatal("test requires the default disabled audit configuration")
				}

				result := installLane.install(t, e)
				joined := strings.Join(append(result.Errors, result.Messages...), "\n")
				if tc.wantFinding == "" {
					if result.Status != "ok" {
						t.Fatalf("NUL-free install status %q: %s", result.Status, joined)
					}
					return
				}
				if result.Status != "failed" ||
					!strings.Contains(joined, "critical audit.opaque.nul-byte") ||
					!strings.Contains(joined, tc.wantFinding) {
					t.Fatalf("install result %+v, want blocking opaque finding naming %s", result, tc.wantFinding)
				}
				if _, err := os.Lstat(installLane.target(e)); err == nil {
					t.Fatalf("blocked install materialized %s", installLane.target(e))
				}
			})
		}
	}
}

func addSkillFilesAndRetag(t *testing.T, e *env, skill string, files map[string]string) {
	t.Helper()
	repo := filepath.Join(e.skillsRoot, skill)
	for name, content := range files {
		e.write(repo, name, content)
	}
	e.git(repo, "add", ".")
	e.git(repo, "commit", "-qm", "add opaque snapshot files")
	e.git(repo, "tag", "-f", "v1")
}

func writeNULTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
