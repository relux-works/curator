package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

// Rework-1 regressions for finding 1 at the install production entry: a
// v1-installed tree that gains NUL bytes afterwards — including a
// v1-collision-preserving edit — is refused with the opaque finding on
// reinstall and on the read-only status plan. Only the pre-hash guard can
// refuse the collision edit, because the v1 digests match by construction.
func TestV1ReinstallRefusesNULAppearingAfterInstall(t *testing.T) {
	priorWriter := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = priorWriter })

	setup := func(t *testing.T) (*env, string) {
		t.Helper()
		e := newEnv(t)
		e.skill("skill-a")
		addSkillFilesAndRetag(t, e, "skill-a", map[string]string{
			"assets/a.bin": "x",
			"assets/b.md":  "y",
		})
		e.declare("skill-a")
		if result := e.install(Options{}); result.Status != "ok" {
			t.Fatalf("clean v1 install status %q: %+v", result.Status, result)
		}
		return e, filepath.Join(e.project, ".agents", "skills", "skill-a")
	}
	assertOpaqueRefusal := func(t *testing.T, result Result, what string) {
		t.Helper()
		joined := strings.Join(append(result.Errors, result.Messages...), "\n")
		if result.Status != "failed" || !strings.Contains(joined, "audit.opaque.nul-byte") {
			t.Fatalf("%s: status %q: %s, want a failed opaque refusal", what, result.Status, joined)
		}
		if strings.Contains(strings.Join(result.Messages, "\n"), "up-to-date") {
			t.Fatalf("%s reported up-to-date over a NUL-bearing v1 tree: %+v", what, result)
		}
	}

	t.Run("collision-preserving-edit", func(t *testing.T) {
		e, installed := setup(t)
		before, err := hashing.ContentSHA256(installed, nil)
		if err != nil {
			t.Fatal(err)
		}
		// Splice the two installed records into one NUL-bearing record
		// with identical v1 stream bytes.
		e.write(installed, "assets/a.bin", "x\x00assets/b.md\x00y")
		if err := os.Remove(filepath.Join(installed, "assets", "b.md")); err != nil {
			t.Fatal(err)
		}
		after, err := hashing.ContentSHA256(installed, nil)
		if err != nil {
			t.Fatal(err)
		}
		if before != after {
			t.Fatalf("fixture is not a v1 collision: %s != %s", before, after)
		}
		assertOpaqueRefusal(t, e.install(Options{}), "reinstall after v1-collision edit")
		assertOpaqueRefusal(t, e.install(Options{Operation: OperationStatus, DryRun: true}), "status plan after v1-collision edit")
	})

	t.Run("fresh-nul-file", func(t *testing.T) {
		e, installed := setup(t)
		e.write(installed, "assets/c.bin", "prefix\x00suffix")
		assertOpaqueRefusal(t, e.install(Options{}), "reinstall after NUL appears")
		assertOpaqueRefusal(t, e.install(Options{Operation: OperationStatus, DryRun: true}), "status plan after NUL appears")
	})
}
