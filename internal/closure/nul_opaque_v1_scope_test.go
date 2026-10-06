// Rework-2 regressions for revision-3 finding F1: the frozen v1 guard
// scans the full skill snapshot before filtering or hashing, including
// declared runtime/build roots and non-whitelisted paths the context
// projection omits. A mutant that moves the scan back to the projection
// returns a digest for every shape below instead of refusing.
package closure

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/skillspec"
)

// The frozen helper refuses NUL under an excluded runtime root, an
// excluded build root, and a non-whitelisted root alike: the projection
// would omit every one of these files, so only the full-snapshot scan
// can refuse. Clean twins in the same shapes still hash, and the
// refusal computes no v1 identity at all.
func TestContentHashForRefusesExcludedNUL(t *testing.T) {
	shapes := []struct {
		name  string
		rel   string
		roots func(*skillspec.Spec)
	}{
		{"runtime", "assets/runtime/deep/a.bin", func(spec *skillspec.Spec) {
			spec.RuntimeRoots = []string{"assets/runtime"}
		}},
		{"build", "assets/runtime/deep/a.bin", func(spec *skillspec.Spec) {
			spec.BuildRoots = []string{"assets/runtime"}
		}},
		{"unlisted", "runtime/deep/a.bin", func(*skillspec.Spec) {}},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			nulTree := t.TempDir()
			writeFrozenTestFiles(t, nulTree, map[string][]byte{
				"SKILL.md": []byte("# Test\n"),
				shape.rel:  []byte("x\x00y"),
			})
			spec := &skillspec.Spec{}
			shape.roots(spec)
			var digest string
			var err error
			calls := hashing.CountV1Hashes(func() {
				digest, err = ContentHashFor(nulTree, spec)
			})
			if err == nil || !strings.Contains(err.Error(), opaquescan.FindingNUL) {
				t.Fatalf("frozen v1 %s NUL = (%q, %v), want the opaque refusal before projection/hash", shape.name, digest, err)
			}
			if digest != "" {
				t.Fatalf("frozen v1 %s NUL returned digest %s alongside the refusal", shape.name, digest)
			}
			if calls != 0 {
				t.Fatalf("frozen v1 %s NUL computed %d v1 identities before refusing, want 0", shape.name, calls)
			}

			cleanTree := t.TempDir()
			writeFrozenTestFiles(t, cleanTree, map[string][]byte{
				"SKILL.md": []byte("# Test\n"),
				shape.rel:  []byte("x"),
			})
			cleanSpec := &skillspec.Spec{}
			shape.roots(cleanSpec)
			var cleanDigest string
			cleanCalls := hashing.CountV1Hashes(func() {
				var cleanErr error
				cleanDigest, cleanErr = ContentHashFor(cleanTree, cleanSpec)
				if cleanErr != nil {
					t.Fatal(cleanErr)
				}
			})
			if cleanDigest == "" {
				t.Fatalf("clean frozen %s context produced no digest", shape.name)
			}
			if cleanCalls == 0 {
				t.Fatalf("clean frozen %s context observed no v1 hash; the seam is not wired", shape.name)
			}
		})
	}
}

const excludedNULPayload = `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`

func writeExcludedNULSkill(t *testing.T, dir, shape, content string) {
	t.Helper()
	switch shape {
	case "runtime":
		writeDraftSkill(t, dir, "review", false, map[string]string{"runtime/deep/a.bin": content}, nil)
	case "build":
		writeDraftSkill(t, dir, "review", false, nil, map[string]string{"build/blob.bin": content})
	case "unlisted":
		writeDraftSkill(t, dir, "review", false, nil, nil)
		full := filepath.Join(dir, "tooling", "cache.bin")
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown excluded shape %q", shape)
	}
}

// The production resolve entry refuses a NUL file the frozen context
// projection would omit, in every excluded shape, before any frozen v1
// identity is produced. Clean twins resolve, proving the refusal is
// NUL-specific rather than shape-specific.
func TestResolveDraftRefusesExcludedNUL(t *testing.T) {
	for _, shape := range []string{"runtime", "build", "unlisted"} {
		t.Run(shape, func(t *testing.T) {
			project, home := t.TempDir(), t.TempDir()
			writeExcludedNULSkill(t, filepath.Join(project, "skills", "review"), shape, "x\x00y")
			m, raw := parseDraftManifest(t, project, excludedNULPayload)
			var resolveErr error
			calls := hashing.CountV1Hashes(func() {
				_, resolveErr = ResolveDraft(DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw})
			})
			if resolveErr == nil || !strings.Contains(resolveErr.Error(), opaquescan.FindingNUL) {
				t.Fatalf("ResolveDraft accepted NUL under excluded %s root: %v", shape, resolveErr)
			}
			if calls != 0 {
				t.Fatalf("refused ResolveDraft computed %d v1 identities, want 0", calls)
			}

			cleanProject, cleanHome := t.TempDir(), t.TempDir()
			writeExcludedNULSkill(t, filepath.Join(cleanProject, "skills", "review"), shape, "clean-bytes")
			cleanManifest, cleanRaw := parseDraftManifest(t, cleanProject, excludedNULPayload)
			var plan *DraftPlan
			cleanCalls := hashing.CountV1Hashes(func() {
				var err error
				plan, err = ResolveDraft(DraftResolveConfig{ProjectRoot: cleanProject, Home: cleanHome, Manifest: cleanManifest, ManifestPayload: cleanRaw})
				if err != nil {
					t.Fatal(err)
				}
			})
			if plan.Lock == nil || len(plan.Lock.Members) != 1 {
				t.Fatalf("clean resolve locked %+v, want one member", plan.Lock)
			}
			if cleanCalls == 0 {
				t.Fatalf("clean resolve observed no v1 hash; the seam is not wired")
			}
		})
	}
}

// A refresh that meets NUL under an excluded root fails with the opaque
// finding and preserves the prior lock and bindings byte-identically:
// the refused generation never publishes.
func TestRefreshDraftExcludedNULPreservesPriorLock(t *testing.T) {
	project, home := t.TempDir(), t.TempDir()
	skill := filepath.Join(project, "skills", "review")
	writeDraftSkill(t, skill, "review", false, map[string]string{"runtime/deep/a.bin": "clean-bytes"}, nil)
	m, raw := parseDraftManifest(t, project, excludedNULPayload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := filepath.Join(home, "bindings.json")
	if _, err := RefreshDraft(cfg, lockPath, bindingsPath); err != nil {
		t.Fatal(err)
	}
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	priorBindings, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(skill, "runtime", "deep", "a.bin"), []byte("x\x00y"), 0o644); err != nil {
		t.Fatal(err)
	}
	var refreshErr error
	calls := hashing.CountV1Hashes(func() {
		_, refreshErr = RefreshDraft(cfg, lockPath, bindingsPath)
	})
	if refreshErr == nil || !strings.Contains(refreshErr.Error(), opaquescan.FindingNUL) {
		t.Fatalf("refresh over excluded NUL = %v, want the opaque refusal", refreshErr)
	}
	if calls != 0 {
		t.Fatalf("refused refresh computed %d v1 identities, want 0", calls)
	}
	keptLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keptLock, priorLock) {
		t.Fatal("refused refresh rewrote the prior lock")
	}
	keptBindings, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keptBindings, priorBindings) {
		t.Fatal("refused refresh rewrote the prior bindings")
	}
}
