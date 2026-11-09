package envprofile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextmaterialize"
)

type repinFixture struct {
	home, repo, operand, a, b string
	homes                     map[string]string
	ids                       *gitIdentities
}

func repinHead(t *testing.T, repo string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func newRepinFixture(t *testing.T) *repinFixture {
	t.Helper()
	f := &repinFixture{home: gitFixtureHome(t), repo: t.TempDir(), homes: pinHomes(t), ids: newGitIdentities(t)}
	// Both directories contain the same package name; source identity must still
	// refuse a directory switch even when the name and requirement match.
	for _, dir := range []string{"profiles/one", "profiles/two"} {
		writePackage(t, filepath.Join(f.repo, filepath.FromSlash(dir)), "repin", "1.0.0", "revision A\n")
	}
	gitRun(t, f.repo, "init")
	gitRun(t, f.repo, "add", ".")
	gitRun(t, f.repo, "commit", "-m", "revision A")
	gitRun(t, f.repo, "tag", "v1.0.0")
	f.a = repinHead(t, f.repo)
	f.operand = f.ids.serve(f.repo, "https://example.com/repin")
	if _, _, _, err := Install(f.home, InstallOptions{Operand: f.operand, Directory: "profiles/one", Revision: f.a}); err != nil {
		t.Fatal(err)
	}
	writePackage(t, filepath.Join(f.repo, "profiles", "one"), "repin", "1.1.0", "revision B\n")
	f.commit(t, "v1.1.0")
	return f
}

func (f *repinFixture) commit(t *testing.T, tag string) {
	t.Helper()
	gitRun(t, f.repo, "add", ".")
	gitRun(t, f.repo, "commit", "-m", tag)
	gitRun(t, f.repo, "tag", tag)
	f.b = repinHead(t, f.repo)
}

// Snapshot every profile record (including source/requirement, current/scoped
// selections, locks and managed homes) plus native homes, bytes/modes/link
// targets. Git acquisition caches and audit caches are outside this contract.
func (f *repinFixture) snapshot(t *testing.T) map[string]map[string]string {
	t.Helper()
	result := map[string]map[string]string{"profiles": preserveSnapshot(t, ProfilesDir(f.home))}
	if _, err := os.Lstat(EnvRoot(f.home)); err == nil {
		result["managed"] = preserveSnapshot(t, EnvRoot(f.home))
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for id, dir := range f.homes {
		if _, err := os.Lstat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			t.Fatal(err)
		}
		result[id] = preserveSnapshot(t, dir)
	}
	return result
}

// Repinning a subdirectory root preserves its source and directory, updates
// machine homes, and makes same-lock repins read-only.
func TestRepinDirectoryScopesAndNoop(t *testing.T) {
	f := newRepinFixture(t)
	expectedSource, err := readSource(f.home, "repin")
	if err != nil {
		t.Fatal(err)
	}
	expectedSource.Req = Requirement{Revision: f.b}
	info, moved, err := UpdateWithOptions(f.home, "repin", UpdateOptions{RootRequirement: &Requirement{Revision: f.b}})
	if err != nil || !moved {
		t.Fatalf("repin: moved=%v err=%v", moved, err)
	}
	root, ok := info.Lock.RootMember()
	if !ok || root.Commit != f.b || root.Directory != "profiles/one" || root.Version != "1.1.0" {
		t.Fatalf("root: %+v", root)
	}
	source, err := readSource(f.home, "repin")
	if err != nil || source != expectedSource || info.Source != expectedSource {
		t.Fatalf("source: %+v %v", source, err)
	}
	for _, entry := range []struct{ id, file string }{{"claude_code", "CLAUDE.md"}, {"codex_cli", "AGENTS.md"}} {
		rendered, err := os.ReadFile(filepath.Join(f.homes[entry.id], entry.file))
		if err != nil || !strings.Contains(string(rendered), "revision B") || strings.Contains(string(rendered), "revision A") {
			t.Fatalf("%s scope: %s %v", entry.id, rendered, err)
		}
	}
	for _, req := range []Requirement{{Revision: f.b}, {Tag: "v1.1.0"}, {Range: "^1.1.0"}} {
		before := f.snapshot(t)
		info, moved, err := UpdateWithOptions(f.home, "repin", UpdateOptions{RootRequirement: &req})
		if err != nil || moved || info.Source != source {
			t.Fatalf("same-lock repin: moved=%v source=%+v err=%v", moved, info.Source, err)
		}
		if !reflect.DeepEqual(before, f.snapshot(t)) {
			t.Fatal("same-lock repin changed source, lock or homes")
		}
	}
}

// Resolve, strict audit, admission and unconfirmed system deltas preserve the
// complete mutable profile/native snapshot. Each refusal has a valid control.
func TestRepinRefusalsPreserveState(t *testing.T) {
	for _, kind := range []string{"resolve", "audit", "admission", "system-delta"} {
		t.Run(kind, func(t *testing.T) {
			f := newRepinFixture(t)
			opts := UpdateOptions{RootRequirement: &Requirement{Revision: f.b}}
			diagnostic := ""
			switch kind {
			case "resolve":
				opts.RootRequirement = &Requirement{Revision: strings.Repeat("f", 40)}
				diagnostic = "context_version_mismatch"
			case "audit":
				writeGitFile(t, f.repo, "profiles/one/context/a.md", directorySecret())
				f.commit(t, "v1.2.0")
				opts.RootRequirement = &Requirement{Revision: f.b}
				diagnostic = "blocking finding"
			case "admission":
				leaf := gitRepo(t, map[string]string{
					"agent-context.json": `{"schema_version":1,"name":"leaf","version":"1.0.0","context":{"modules":[{"path":"system.md","class":"system"}]}}`,
					"context/system.md":  "leaf system context\n",
				}, "v1.0.0")
				leafURL := f.ids.serve(leaf, "https://example.com/leaf")
				mid := gitRepo(t, map[string]string{
					"agent-context.json": fmt.Sprintf(`{"schema_version":1,"name":"mid","version":"1.0.0","requires":{"contexts":{"leaf":{"git":%q,"tag":"v1.0.0"}}}}`, leafURL),
				}, "v1.0.0")
				operand := f.ids.serve(mid, "https://example.com/mid")
				manifest := fmt.Sprintf(`{"schema_version":1,"name":"repin","version":"1.2.0","requires":{"contexts":{"mid":{"git":%q,"tag":"v1.0.0"}}}}`, operand)
				writeGitFile(t, f.repo, "profiles/one/agent-context.json", manifest)
				f.commit(t, "v1.2.0")
				opts.RootRequirement = &Requirement{Revision: f.b}
				opts.Policy.TransitiveSystemModules = contextmaterialize.TransitiveError
				diagnostic = contextmaterialize.DiagSystemModuleTransitive
			case "system-delta":
				writeGitFile(t, f.repo, "profiles/one/agent-context.json", `{"schema_version":1,"name":"repin","version":"1.2.0","context":{"modules":[{"path":"a.md","class":"system"}]}}`)
				f.commit(t, "v1.2.0")
				opts.RootRequirement = &Requirement{Revision: f.b}
				diagnostic = DiagSystemDeltaConfirmationRequired
			}
			before := f.snapshot(t)
			_, moved, err := UpdateWithOptions(f.home, "repin", opts)
			if err == nil || moved || !strings.Contains(err.Error(), diagnostic) {
				t.Fatalf("%s refusal: moved=%v err=%v, want %s", kind, moved, err, diagnostic)
			}
			if !reflect.DeepEqual(before, f.snapshot(t)) {
				t.Fatal("refusal changed source, requirement, lock or homes")
			}
			if kind == "audit" {
				writePackage(t, filepath.Join(f.repo, "profiles", "one"), "repin", "1.3.0", "safe revision C\n")
				f.commit(t, "v1.3.0")
			}
			opts.RootRequirement = &Requirement{Revision: f.b}
			opts.ConfirmSystemDelta = true
			if kind == "admission" {
				opts.Policy.SystemModuleWaivers = []SystemModuleWaiver{{Package: "leaf", Reason: "fixture control"}}
			}
			if _, moved, err := UpdateWithOptions(f.home, "repin", opts); err != nil || !moved {
				t.Fatalf("valid control: moved=%v err=%v", moved, err)
			}
		})
	}
}

// Install identity still includes both Git URL and directory; repin exposes
// neither override. Matching install still succeeds without changing the pin.
func TestRepinDoesNotBroadenInstallIdentity(t *testing.T) {
	f := newRepinFixture(t)
	other := f.ids.serve(f.repo, "https://example.com/other")
	for _, opts := range []InstallOptions{
		{Operand: other, Directory: "profiles/one", Revision: f.a, As: "repin"},
		{Operand: f.operand, Directory: "profiles/two", Revision: f.a, As: "repin"},
		{Operand: f.operand, Directory: "profiles/one", Revision: f.b, As: "repin"},
	} {
		before := f.snapshot(t)
		if _, _, _, err := Install(f.home, opts); err == nil || !strings.Contains(err.Error(), DiagNameTaken) {
			t.Fatalf("install identity refusal: %v", err)
		}
		if !reflect.DeepEqual(before, f.snapshot(t)) {
			t.Fatal("refused install changed mutable profile state")
		}
	}
	if _, _, updated, err := Install(f.home, InstallOptions{Operand: f.operand, Directory: "profiles/one", Revision: f.a, As: "repin"}); err != nil || !updated {
		t.Fatalf("same source install: updated=%v err=%v", updated, err)
	}
}

// Internal callers cannot supply empty/conflicting requirements or repin a
// path/builtin profile; those refusals preserve existing mutable state.
func TestRepinRequirementValidation(t *testing.T) {
	f := newRepinFixture(t)
	pkg := t.TempDir()
	writePackage(t, pkg, "path-profile", "1.0.0", "path\n")
	if _, _, _, err := Install(f.home, InstallOptions{Operand: pkg}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		req  Requirement
	}{
		{"repin", Requirement{}}, {"repin", Requirement{Revision: f.b, Tag: "v1.1.0"}},
		{"", Requirement{Revision: f.b}}, {"path-profile", Requirement{Revision: f.b}}, {DefaultProfile, Requirement{Revision: f.b}},
	} {
		before := f.snapshot(t)
		if _, moved, err := UpdateWithOptions(f.home, tc.name, UpdateOptions{RootRequirement: &tc.req}); err == nil || moved || !strings.Contains(err.Error(), DiagRefConflict) {
			t.Fatalf("%s validation: moved=%v err=%v", tc.name, moved, err)
		}
		if !reflect.DeepEqual(before, f.snapshot(t)) {
			t.Fatal("requirement refusal changed mutable profile state")
		}
	}
}

// A late unmanaged-home refusal compensates already planned/published scopes
// and the repin records; the same candidate succeeds with explicit takeover.
func TestRepinTakeoverRefusalPreservesState(t *testing.T) {
	f := newRepinFixture(t)
	native := claudeHome(t)
	f.homes["claude_code"] = native
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("unmanaged context\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := f.snapshot(t)
	opts := UpdateOptions{RootRequirement: &Requirement{Revision: f.b}}
	_, moved, err := UpdateWithOptions(f.home, "repin", opts)
	if err == nil || moved || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("takeover refusal: moved=%v err=%v", moved, err)
	}
	if !reflect.DeepEqual(before, f.snapshot(t)) {
		after := f.snapshot(t)
		for group, entries := range before {
			for path, value := range entries {
				if after[group][path] != value {
					t.Errorf("late refusal changed %s/%s", group, path)
				}
			}
		}
		for group, entries := range after {
			for path, value := range entries {
				if before[group][path] != value {
					t.Errorf("late refusal added/changed %s/%s", group, path)
				}
			}
		}
	}
	opts.Policy.Takeover = true
	if _, moved, err := UpdateWithOptions(f.home, "repin", opts); err != nil || !moved {
		t.Fatalf("takeover control: moved=%v err=%v", moved, err)
	}
	rendered, err := os.ReadFile(filepath.Join(native, "CLAUDE.md"))
	if err != nil || !strings.Contains(string(rendered), "revision B") {
		t.Fatalf("takeover context: %s %v", rendered, err)
	}
}

// Like plain update, repin refuses an absent exact dependency object in a stale
// cache. After fixture acquisition, it resolves that pin and configured overlays.
func TestRepinOverlayAndMissingPinRules(t *testing.T) {
	f := newRepinFixture(t)
	dep := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"dep","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`,
		"context/a.md":       "dependency A\n",
	}, "v1.0.0")
	depURL := f.ids.serve(dep, "https://example.com/dep")
	manager := newGitManager(f.home)
	identity, err := manager.Identity("context", "dep", depURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ensureRepo(identity); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, dep, "context/a.md", "dependency B\n")
	gitRun(t, dep, "add", ".")
	gitRun(t, dep, "commit", "-m", "new exact dependency pin")
	pin := repinHead(t, dep)
	overlay := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"overlay","version":"1.0.0","context":{"modules":[{"path":"o.md"}]}}`,
		"context/o.md":       "overlay context\n",
	}, "v1.0.0")
	overlayURL := f.ids.serve(overlay, "https://example.com/overlay")
	manifest := fmt.Sprintf(`{"schema_version":1,"name":"repin","version":"1.2.0","context":{"modules":[{"path":"a.md"}]},"requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, depURL, pin)
	writeGitFile(t, f.repo, "profiles/one/agent-context.json", manifest)
	f.commit(t, "v1.2.0")
	opts := UpdateOptions{
		RootRequirement: &Requirement{Revision: f.b},
		Policy:          Policy{OverlaysAllowed: true, OverlayDefaultWeight: 1000, Overlays: map[string][]OverlaySpec{"repin": {{Source: overlayURL, Tag: "v1.0.0"}}}},
	}
	before := f.snapshot(t)
	if _, moved, err := UpdateWithOptions(f.home, "repin", opts); err == nil || moved || !strings.Contains(err.Error(), DiagSourceInvalid) || !strings.Contains(err.Error(), "extract dep") {
		t.Fatalf("missing-pin refusal: moved=%v err=%v", moved, err)
	}
	if !reflect.DeepEqual(before, f.snapshot(t)) {
		t.Fatal("missing-pin refusal changed profile state")
	}
	// Explicit fixture acquisition is the positive control, not install-style
	// missing-pin recovery added to the update contract.
	if err := manager.fetch(identity); err != nil {
		t.Fatal(err)
	}
	info, moved, err := UpdateWithOptions(f.home, "repin", opts)
	if err != nil || !moved {
		t.Fatalf("repin: moved=%v err=%v", moved, err)
	}
	member, ok := info.Lock.Find("context", "dep")
	if !ok || member.Commit != pin {
		t.Fatalf("missing dependency pin: %+v", member)
	}
	if member, ok := info.Lock.Find("context", "overlay"); !ok || !member.Overlay {
		t.Fatalf("missing overlay: %+v", member)
	}
	rendered, err := os.ReadFile(filepath.Join(f.homes["codex_cli"], "AGENTS.md"))
	if err != nil || !strings.Contains(string(rendered), "dependency B") || !strings.Contains(string(rendered), "overlay context") {
		t.Fatalf("resolved scope: %s %v", rendered, err)
	}
}

// Repinning an environment-scoped profile updates that home while preserving
// the other profile's machine home and both scope selections.
func TestRepinScopedHome(t *testing.T) {
	f := newRepinFixture(t)
	other := t.TempDir()
	writePackage(t, other, "other", "1.0.0", "other context\n")
	if _, _, _, err := Install(f.home, InstallOptions{Operand: other}); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(f.home, "other", "", "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := Use(f.home, "repin", "codex_cli", "", false); err != nil {
		t.Fatal(err)
	}
	before := preserveSnapshot(t, f.homes["claude_code"])
	opts := UpdateOptions{RootRequirement: &Requirement{Tag: "v1.1.0"}}
	if _, moved, err := UpdateWithOptions(f.home, "repin", opts); err != nil || !moved {
		t.Fatalf("scoped repin: moved=%v err=%v", moved, err)
	}
	if !reflect.DeepEqual(before, preserveSnapshot(t, f.homes["claude_code"])) {
		t.Fatal("scoped repin changed unrelated machine home")
	}
	if current, err := Current(f.home); err != nil || current != "other" {
		t.Fatalf("machine current: %s %v", current, err)
	}
	scoped, err := ScopedCurrents(f.home)
	if err != nil || scoped["env:codex_cli"] != "repin" {
		t.Fatalf("scoped current: %v %v", scoped, err)
	}
	rendered, err := os.ReadFile(filepath.Join(f.homes["codex_cli"], "AGENTS.md"))
	if err != nil || !strings.Contains(string(rendered), "revision B") {
		t.Fatalf("scoped context: %s %v", rendered, err)
	}
}

// A selected-directory tag still has to agree with that package's version;
// validating an unrelated repository-root manifest must not admit the tag.
func TestRepinDirectoryTagMismatch(t *testing.T) {
	f := newRepinFixture(t)
	if _, moved, err := UpdateWithOptions(f.home, "repin", UpdateOptions{RootRequirement: &Requirement{Tag: "v1.1.0"}}); err != nil || !moved {
		t.Fatalf("valid selected tag: moved=%v err=%v", moved, err)
	}
	gitRun(t, f.repo, "tag", "v9.0.0")
	before := f.snapshot(t)
	if _, moved, err := UpdateWithOptions(f.home, "repin", UpdateOptions{RootRequirement: &Requirement{Tag: "v9.0.0"}}); err == nil || moved || !strings.Contains(err.Error(), "context_version_mismatch") {
		t.Fatalf("tag mismatch: moved=%v err=%v", moved, err)
	}
	if !reflect.DeepEqual(before, f.snapshot(t)) {
		t.Fatal("tag mismatch changed state")
	}

}
