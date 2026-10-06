package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/pathboundary"
)

// TestInstallBlocksBothV1CollidingSkillTrees drives Install through
// contextresolve.Resolve -> auditAndStore -> auditMember. Both trees have the
// same v1 hash, but each has a regular NUL-bearing file in a different path.
func TestInstallBlocksBothV1CollidingSkillTrees(t *testing.T) {
	enableV1WritersForTest(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	collision := []struct {
		name    string
		files   map[string]string
		finding string
	}{
		{
			name: "single-record",
			files: map[string]string{
				"assets/a.bin": "x\x00docs/b.md\x00y\x00z",
			},
			finding: "assets/a.bin",
		},
		{
			name: "split-records",
			files: map[string]string{
				"assets/a.bin": "x",
				"docs/b.md":    "y\x00z",
			},
			finding: "docs/b.md",
		},
	}
	left, right := t.TempDir(), t.TempDir()
	writeV1Files(t, left, collision[0].files)
	writeV1Files(t, right, collision[1].files)
	leftHash, err := hashing.ContentSHA256(left, nil)
	if err != nil {
		t.Fatal(err)
	}
	rightHash, err := hashing.ContentSHA256(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	if leftHash != rightHash {
		t.Fatalf("v1 framing did not collide: %s != %s", leftHash, rightHash)
	}

	for _, tc := range collision {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			skillRepo := gitRepo(t, tc.files, "v1.0.0")
			skillOperand := ids.serve(skillRepo, "https://example.com/"+tc.name+"-skill")
			rootRepo := gitRepo(t, skillDependencyRootFiles(tc.name, skillOperand), "v1.0.0")
			rootOperand := ids.serve(rootRepo, "https://example.com/"+tc.name+"-root")

			_, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) ||
				!strings.Contains(err.Error(), contextaudit.ClassOpaqueFile) || !strings.Contains(err.Error(), tc.finding) {
				t.Fatalf("Install error = %v, want a blocking opaque finding naming %s", err, tc.finding)
			}
			if _, err := readSource(home, tc.name); err == nil {
				t.Fatal("blocked install published a profile source")
			}
		})
	}

	t.Run("nul-free-skill-still-installs", func(t *testing.T) {
		home := t.TempDir()
		skillRepo := gitRepo(t, map[string]string{
			"assets/a.bin": "x",
			"docs/b.md":    "y",
		}, "v1.0.0")
		skillOperand := ids.serve(skillRepo, "https://example.com/nul-free-skill")
		rootRepo := gitRepo(t, skillDependencyRootFiles("clean-root", skillOperand), "v1.0.0")
		rootOperand := ids.serve(rootRepo, "https://example.com/nul-free-root")
		info, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
		if err != nil {
			t.Fatal(err)
		}
		if member, ok := lockMember(info.Lock, "sk"); !ok || member.Kind != contextlock.KindSkill {
			t.Fatalf("installed lock lacks skill member: %+v", info.Lock.Members)
		}
	})
}

func TestInstallBlocksDeepNULFileInContextPathSnapshot(t *testing.T) {
	enableV1WritersForTest(t)
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writeGitFile(t, source, "agent-context.json", `{"schema_version": 1, "name": "nul-context", "version": "1.0.0",`+
		`"context": {"modules": [{"path": "a.md"}]}}`+"\n")
	writeGitFile(t, source, "context/a.md", "clean\n")
	deep := filepath.Join(source, "assets", "deep", "image.unsupported")
	if err := os.MkdirAll(filepath.Dir(deep), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deep, []byte{1, 0, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(source); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := Install(home, InstallOptions{Operand: source})
	// The refusal fires at the store content hash, before any v1 identity
	// is computed over these bytes (Spec §8 interim rule), so it carries
	// the shared pre-hash opaque id rather than the detector class the
	// later audit layers would have reported.
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) ||
		!strings.Contains(err.Error(), opaquescan.FindingNUL) || !strings.Contains(err.Error(), "assets/deep/image.unsupported") {
		t.Fatalf("Install error = %v, want a blocking opaque finding naming the deep file", err)
	}
	if _, err := readSource(home, "nul-context"); err == nil {
		t.Fatal("blocked context install published a profile source")
	}
}

func TestUpdateBlocksDeepNULFileInNewContextMember(t *testing.T) {
	enableV1WritersForTest(t)
	home, _, oldHash := installBlockingOverlayRoot(t)
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeManifestPackage(t, overlay,
		`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{
			"a.md":                     "clean\n",
			"docs/deep/opaque.unknown": "first\x00last",
		})
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	_, _, err := UpdateWithPolicy(home, "acme", policy)
	// The store pre-hash guard refuses the overlay source at load, before
	// any member audit exists — hence profile_source_invalid rather than
	// the update-blocked member diagnostic — so the shared opaque id, not
	// the detector class, names the deep file. The update is still
	// refused and the old lock still stands, asserted below.
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) ||
		!strings.Contains(err.Error(), opaquescan.FindingNUL) || !strings.Contains(err.Error(), "docs/deep/opaque.unknown") {
		t.Fatalf("UpdateWithPolicy error = %v, want a blocking opaque finding naming the deep file", err)
	}
	_, afterHash, err := readLock(home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if afterHash != oldHash {
		t.Fatal("blocked update moved the lock")
	}
}

func skillDependencyRootFiles(name, skillOperand string) map[string]string {
	return map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "` + name + `", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"skills": {"sk": {"git": "` + skillOperand + `", "range": "*"}}}}` + "\n",
		"context/a.md": "clean\n",
	}
}

func writeV1Files(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
