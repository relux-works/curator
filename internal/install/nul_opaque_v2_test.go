package install

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	markerpkg "github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// NUL fixture bytes shared with nul_opaque_test.go: the single-record tree
// collides under v1 with the split-records twin. The collision itself is
// asserted there and in the audit package; these tests prove the v2
// production entry distinguishes the twins and never computes v1 over them.
var (
	nulSingleRecordFiles = map[string]string{
		"assets/a.bin": "x\x00docs/b.md\x00y\x00z",
	}
	nulSplitRecordsFiles = map[string]string{
		"assets/a.bin": "x",
		"docs/b.md":    "y\x00z",
	}
	nulDeepFile = map[string]string{
		"docs/deep/opaque.unsupported": "prefix\x00suffix",
	}
)

func pinV2Writers(t *testing.T) {
	t.Helper()
	priorWriter := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = priorWriter })
}

// A NUL-bearing skill installs, audits, and reports current under v2 with
// a v2 identity on both the project and global lanes: NUL bytes are
// ordinary v2 data (Spec §8), so the interim v1 rule stays silent.
func TestV2InstallAdmitsNULBearingSkillAtProjectAndGlobal(t *testing.T) {
	pinV2Writers(t)
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

	for _, installLane := range lanes {
		t.Run(installLane.name, func(t *testing.T) {
			e := newEnv(t)
			e.skill("skill-a")
			addSkillFilesAndRetag(t, e, "skill-a", nulSingleRecordFiles)
			addSkillFilesAndRetag(t, e, "skill-a", nulDeepFile)
			if e.cfg.Audit.Enabled {
				t.Fatal("test requires the default disabled audit configuration")
			}

			result := installLane.install(t, e)
			if result.Status != "ok" {
				t.Fatalf("v2 NUL install status %q: %s", result.Status,
					strings.Join(append(result.Errors, result.Messages...), "\n"))
			}
			installed := installLane.target(e)
			recorded := markerpkg.Read(installed)
			if recorded == nil {
				t.Fatalf("no marker recorded at %s", installed)
			}
			if recorded.SchemaVersion != markerpkg.SchemaV5 || recorded.HashVersion != hashing.VersionV2 {
				t.Fatalf("marker versions = schema:%d hash:%d, want schema:%d hash:2",
					recorded.SchemaVersion, recorded.HashVersion, markerpkg.SchemaV5)
			}
			wantV2, err := hashing.ContentSHA256WithVersion(installed, nil, hashing.VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if recorded.ContentSHA256 != wantV2 {
				t.Fatalf("installed identity = %s, want the v2 recomputation %s", recorded.ContentSHA256, wantV2)
			}
			// No v1 identity was computed or trusted for this tree: the
			// v1 digest keys no audit trust state.
			v1Installed, err := hashing.ContentSHA256(installed, nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(e.home, "audit", hashing.Normalize(v1Installed))); !os.IsNotExist(err) {
				t.Fatalf("v2 install left trust state under the v1 digest: %v", err)
			}

			again := installLane.install(t, e)
			if again.Status != "ok" {
				t.Fatalf("v2 NUL reinstall status %q: %+v", again.Status, again)
			}
			if !strings.Contains(strings.Join(again.Messages, "\n"), "up-to-date") {
				t.Fatalf("v2 NUL reinstall did not report up-to-date installations: %q", again.Messages)
			}
		})
	}
}

// The v1-colliding twins install under v2 with distinct v2 identities: the
// install never confuses the NUL tree with its clean twin.
func TestV2InstallDistinguishesNULCollisionTwins(t *testing.T) {
	pinV2Writers(t)
	first, second := t.TempDir(), t.TempDir()
	writeNULTree(t, first, nulSingleRecordFiles)
	writeNULTree(t, second, nulSplitRecordsFiles)
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

	e := newEnv(t)
	e.skill("skill-a")
	addSkillFilesAndRetag(t, e, "skill-a", nulSingleRecordFiles)
	e.skill("skill-b")
	addSkillFilesAndRetag(t, e, "skill-b", nulSplitRecordsFiles)
	e.declare("skill-a", "skill-b")

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("v2 twin install status %q: %+v", result.Status, result)
	}
	markerA := markerpkg.Read(filepath.Join(e.project, ".agents", "skills", "skill-a"))
	markerB := markerpkg.Read(filepath.Join(e.project, ".agents", "skills", "skill-b"))
	if markerA == nil || markerB == nil {
		t.Fatalf("missing markers: skill-a=%+v skill-b=%+v", markerA, markerB)
	}
	if markerA.HashVersion != hashing.VersionV2 || markerB.HashVersion != hashing.VersionV2 {
		t.Fatalf("twin marker hash versions = %d/%d, want v2/v2", markerA.HashVersion, markerB.HashVersion)
	}
	if markerA.ContentSHA256 == markerB.ContentSHA256 {
		t.Fatalf("twin v2 identities collide: %s", markerA.ContentSHA256)
	}
}

// The draft lane carries frozen v1 identities, so a NUL-bearing draft
// member is refused with the opaque finding at the explicit
// resolve/refresh entry: the frozen package context hash is never
// computed over NUL bytes, so no lock can bind such a tree and no v1
// digest of it ever exists. Refresh additionally preserves the prior
// lock on refusal.
func TestDraftLaneStillBlocksNULBearingMember(t *testing.T) {
	pinV2Writers(t)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	writeNULMember := func(t *testing.T, project string) {
		t.Helper()
		nulPath := filepath.Join(project, "skills", "review", "assets", "a.bin")
		if err := os.MkdirAll(filepath.Dir(nulPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(nulPath, []byte("x\x00docs/b.md\x00y"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	resolveErr := func(t *testing.T, project, home string) error {
		t.Helper()
		m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = closure.ResolveDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload)})
		return err
	}
	assertOpaqueRefusal := func(t *testing.T, err error, what string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "audit.opaque.nul-byte") {
			t.Fatalf("%s: err = %v, want the opaque refusal", what, err)
		}
		// The refusal fires at the frozen-context load, before any v1
		// digest exists: without the pre-hash guard this same tree
		// would hash and fail later at the audit gate instead.
		if !strings.Contains(err.Error(), "source_member_invalid") {
			t.Fatalf("%s: err = %v, want the pre-hash frozen-context refusal", what, err)
		}
	}

	t.Run("explicit-resolve-refuses", func(t *testing.T) {
		project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
		writeNULMember(t, project)
		assertOpaqueRefusal(t, resolveErr(t, project, home), "explicit resolve over a NUL member")
	})

	t.Run("refresh-refuses-and-keeps-prior-lock", func(t *testing.T) {
		project, home, _ := draftProject(t, payload, map[string]string{"skills/review": "review"})
		resolveDraftForInstall(t, project, home, payload)
		lockPath := sourcelock.PathIn(project)
		before, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatal(err)
		}
		writeNULMember(t, project)
		m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = closure.RefreshDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload)}, lockPath, "")
		assertOpaqueRefusal(t, err, "refresh after NUL appears")
		after, readErr := os.ReadFile(lockPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("refused refresh rewrote the prior lock")
		}
	})
}
