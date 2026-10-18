package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
)

// Project must refuse every unowned stale sibling before publication, preserving
// exact trees, the adapter ledger, and a foreign link; removing the conflicts
// must still permit removal of a genuinely owned stale node.
func TestInstallPreservesUnownedStaleStoreEntries(t *testing.T) {
	project, home := setupDraftSweep(t)
	if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
		t.Fatalf("initial: %+v", r)
	}
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	mustOwnershipWrite(t, filepath.Join(project, "Skillfile.json"), payload)
	refreshDraftLocked(t, project, home, payload, nil)
	store := filepath.Join(project, ".agents", "skills")
	names := []string{"unmanaged-skill", "unmanaged-notes", "loose.txt"}
	mustOwnershipWrite(t, filepath.Join(store, names[0], "SKILL.md"), "manual skill\n")
	mustOwnershipWrite(t, filepath.Join(store, names[1], "notes.md"), "manual notes\n")
	mustOwnershipWrite(t, filepath.Join(store, names[2]), "loose bytes\n")
	link := filepath.Join(project, ".claude", "skills", "manual-link")
	if err := os.Symlink(filepath.Join(store, names[0]), link); err != nil {
		t.Fatal(err)
	}
	before := snapshotDraftState(t, home, project)
	r := draftSweepInstall(t, home, project, Options{})
	if r.Status != "failed" {
		t.Fatalf("unowned stale entries admitted: %+v", r)
	}
	for _, name := range names {
		if !strings.Contains(strings.Join(r.Errors, ";"), filepath.Join(store, name)) {
			t.Errorf("missing conflict %s: %v", name, r.Errors)
		}
	}
	assertOwnershipUnchanged(t, before, home, project)
	if _, err := os.Stat(filepath.Join(link, "SKILL.md")); err != nil {
		t.Fatalf("foreign link broken: %v", err)
	}
	for _, name := range names {
		if err := os.RemoveAll(filepath.Join(store, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
		t.Fatalf("owned cleanup: %+v", r)
	}
	if _, err := os.Lstat(filepath.Join(store, "scan")); !os.IsNotExist(err) {
		t.Fatalf("owned stale node retained: %v", err)
	}
	if marker.Read(filepath.Join(store, "review")) == nil {
		t.Fatal("selected owned node lost")
	}
}

// Project must refuse existing unproven canonical destinations, including
// notes-only entries and marker/canonical links, without partial publication.
// Directory-as-marker exercises an actual read failure without chmod/root assumptions.
func TestInstallRefusesUnownedSelectedCanonicalDestination(t *testing.T) {
	for _, variant := range []string{"manual", "notes-only", "invalid", "unreadable", "wrong-name", "wrong-name-partial", "marker-link", "canonical-link", "loose-file"} {
		t.Run(variant, func(t *testing.T) {
			project, home := setupDraftSweep(t)
			// Capture a valid preceding marker for identity and link-neighbour cases.
			if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
				t.Fatalf("initial: %+v", r)
			}
			live := filepath.Join(project, ".agents", "skills", "review")
			saved, err := os.ReadFile(filepath.Join(live, marker.Name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.RemoveAll(live); err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "loose-file":
				mustOwnershipWrite(t, live, "manual file")
			case "canonical-link":
				target := filepath.Join(project, "foreign")
				mustOwnershipWrite(t, filepath.Join(target, "SKILL.md"), "foreign")
				mustOwnershipWrite(t, filepath.Join(target, marker.Name), string(saved))
				if err := os.Symlink(target, live); err != nil {
					t.Fatal(err)
				}
			default:
				mustOwnershipWrite(t, filepath.Join(live, "note.md"), "keep exact notes\n")
				if variant != "notes-only" {
					mustOwnershipWrite(t, filepath.Join(live, "SKILL.md"), "manual skill\n")
				}
				switch variant {
				case "invalid":
					mustOwnershipWrite(t, filepath.Join(live, marker.Name), "{}\n")
				case "unreadable":
					if err := os.Mkdir(filepath.Join(live, marker.Name), 0o755); err != nil {
						t.Fatal(err)
					}
				case "wrong-name-partial":
					mustOwnershipWrite(t, filepath.Join(live, marker.Name), `{"schema_version":5,"name":"other"}`)
				case "wrong-name":
					mustOwnershipWrite(t, filepath.Join(live, marker.Name), strings.Replace(string(saved), `"name": "review"`, `"name": "other"`, 1))
				case "marker-link":
					target := filepath.Join(project, "foreign-marker.json")
					mustOwnershipWrite(t, target, string(saved))
					if err := os.Symlink(target, filepath.Join(live, marker.Name)); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := snapshotDraftState(t, home, project)
			r := draftSweepInstall(t, home, project, Options{})
			if r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, ";"), live) {
				t.Fatalf("want named canonical refusal: %+v", r)
			}
			assertOwnershipUnchanged(t, before, home, project)
		})
	}
}

// Invalid/read-failing markers and links in stale positions cannot authorize
// deletion; each refusal leaves both managed and unmanaged state intact.
func TestInstallRefusesUnprovenStaleOwnership(t *testing.T) {
	for _, variant := range []string{"invalid", "unreadable", "wrong-name", "marker-link", "canonical-link"} {
		t.Run(variant, func(t *testing.T) {
			project, home := setupDraftSweep(t)
			if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
				t.Fatalf("initial: %+v", r)
			}
			store := filepath.Join(project, ".agents", "skills")
			live := filepath.Join(store, "scan")
			path := filepath.Join(live, marker.Name)
			saved, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "invalid":
				mustOwnershipWrite(t, path, "{}")
			case "unreadable":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			case "wrong-name":
				mustOwnershipWrite(t, path, strings.Replace(string(saved), `"name": "scan"`, `"name": "other"`, 1))
			case "marker-link":
				target := filepath.Join(project, "saved-marker")
				mustOwnershipWrite(t, target, string(saved))
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			case "canonical-link":
				target := filepath.Join(project, "saved-skill")
				if err := os.Rename(live, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, live); err != nil {
					t.Fatal(err)
				}
			}
			payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
			mustOwnershipWrite(t, filepath.Join(project, "Skillfile.json"), payload)
			refreshDraftLocked(t, project, home, payload, nil)
			before := snapshotDraftState(t, home, project)
			r := draftSweepInstall(t, home, project, Options{})
			if r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, ";"), live) {
				t.Fatalf("want named stale refusal: %+v", r)
			}
			assertOwnershipUnchanged(t, before, home, project)
		})
	}
}

// Absent destinations install; owned drift updates; clean reinstall succeeds;
// unrelated authored instructions survive all three production operations.
func TestInstallOwnedCanonicalPositiveControls(t *testing.T) {
	project, home := setupDraftSweep(t)
	unrelated := filepath.Join(project, ".agents", ".instructions", "hand.md")
	mustOwnershipWrite(t, unrelated, "manual instructions")
	for i := 0; i < 3; i++ {
		if i == 1 {
			mustOwnershipWrite(t, filepath.Join(project, ".agents", "skills", "review", "SKILL.md"), "drift")
		}
		if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
			t.Fatalf("install %d: %+v", i, r)
		}
		content, err := os.ReadFile(filepath.Join(project, ".agents", "skills", "review", "SKILL.md"))
		if err != nil || string(content) == "drift" {
			t.Fatalf("owned content not repaired: %q %v", content, err)
		}
		content, err = os.ReadFile(unrelated)
		if err != nil || string(content) != "manual instructions" {
			t.Fatalf("unrelated instructions: %q %v", content, err)
		}
	}
}

func mustOwnershipWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
func assertOwnershipUnchanged(t *testing.T, before draftSharedState, home, project string) {
	t.Helper()
	if changed := before.diff(snapshotDraftState(t, home, project)); len(changed) > 0 {
		t.Fatalf("refusal changed state: %v", changed)
	}
	assertNoDraftJournalRemains(t, home)
}

// The public Global entry shares canonical collision and stale-removal gates;
// refusal preserves global bytes and native mirrors, then owned repair succeeds.
func TestGlobalCanonicalOwnershipPreservation(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}
	userHome := t.TempDir()
	run := func() Result { return Global(e.cfg, userHome, Options{Platform: installPlatform()}) }
	live := filepath.Join(e.home, "global", "skills", "skill-g")
	mustOwnershipWrite(t, filepath.Join(live, "note.md"), "manual global")
	before := draftEntryDigest(filepath.Join(e.home, "global"))
	if r := run(); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, ";"), live) {
		t.Fatalf("selected global: %+v", r)
	}
	if got := draftEntryDigest(filepath.Join(e.home, "global")); got != before {
		t.Fatal("global collision changed state")
	}
	if _, err := os.Lstat(filepath.Join(userHome, ".claude", "skills", "skill-g")); !os.IsNotExist(err) {
		t.Fatalf("global mirror published: %v", err)
	}
	if err := os.RemoveAll(live); err != nil {
		t.Fatal(err)
	}
	if r := run(); r.Status != "ok" {
		t.Fatalf("absent global: %+v", r)
	}
	foreign := filepath.Join(e.home, "global", "skills", "foreign.txt")
	mustOwnershipWrite(t, foreign, "foreign")
	before = draftEntryDigest(filepath.Join(e.home, "global"))
	mirrors := draftEntryDigest(userHome)
	if r := run(); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, ";"), foreign) {
		t.Fatalf("stale global: %+v", r)
	}
	if draftEntryDigest(filepath.Join(e.home, "global")) != before || draftEntryDigest(userHome) != mirrors {
		t.Fatal("stale global refusal changed state")
	}
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	mustOwnershipWrite(t, filepath.Join(live, "SKILL.md"), "drift")
	if r := run(); r.Status != "ok" {
		t.Fatalf("owned global update: %+v", r)
	}
}

// Hybrid selected nodes share ownership checks; unrelated shared siblings are
// retained because hybrid cleanup belongs to consumer-aware GC, not project pruning.
func TestHybridCanonicalOwnershipPreservation(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-h")
	e.declare()
	e.write(e.home, "hybrid/Skillfile.json", `{"schema_version":1,"skills":[{"name":"skill-h","tag":"v1","targets":["test"]}]}`)
	live := filepath.Join(e.home, "hybrid", "skills", "skill-h")
	mustOwnershipWrite(t, filepath.Join(live, "note.md"), "manual hybrid")
	before := snapshotState(t, e)
	if r := e.install(Options{}); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, ";"), live) {
		t.Fatalf("hybrid collision: %+v", r)
	}
	if changed := before.diff(snapshotState(t, e)); len(changed) > 0 {
		t.Fatalf("hybrid collision changed state: %v", changed)
	}
	if err := os.RemoveAll(live); err != nil {
		t.Fatal(err)
	}
	if r := e.install(Options{}); r.Status != "ok" {
		t.Fatalf("absent hybrid: %+v", r)
	}
	foreign := filepath.Join(e.home, "hybrid", "skills", "foreign.txt")
	mustOwnershipWrite(t, foreign, "shared sibling")
	mustOwnershipWrite(t, filepath.Join(live, "SKILL.md"), "drift")
	if r := e.install(Options{}); r.Status != "ok" {
		t.Fatalf("owned hybrid: %+v", r)
	}
	if content, err := os.ReadFile(foreign); err != nil || string(content) != "shared sibling" {
		t.Fatalf("hybrid sibling lost: %q %v", content, err)
	}
	if content, err := os.ReadFile(filepath.Join(live, "SKILL.md")); err != nil || string(content) == "drift" {
		t.Fatalf("hybrid repair failed: %q %v", content, err)
	}
}
