package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/testtoolchain"
)

type museRootVector struct {
	Name       string          `json:"name"`
	Context    json.RawMessage `json:"context"`
	Form       string          `json:"form"`
	Written    bool            `json:"written"`
	Body       string          `json:"body"`
	Diagnostic string          `json:"diagnostic"`
}

func museRootProfile(t *testing.T, context json.RawMessage) stubConfigSource {
	t.Helper()
	source, _ := profileHome(t)
	prepareNativeMuse(t)
	pkg := t.TempDir()
	manifest := map[string]any{"schema_version": 1, "name": "acme", "version": "1.0.0"}
	if context != nil {
		manifest["context"] = context
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(pkg, "context"), 0700); err != nil {
		t.Fatal(err)
	}
	for path, payload := range map[string][]byte{"agent-context.json": raw, "context/a.md": []byte("hello\n")} {
		if err := os.WriteFile(filepath.Join(pkg, path), payload, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := pathboundary.ProtectTree(pkg); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install: exit %d: %s", code, stderr)
	}
	return source
}

// The CLI drives cmdEnvResolve -> envprofile.Resolve -> homePlan.rootContext.
// Expected bytes are assembled independently of Monolithic, including the
// header-only case, so a correct path alone cannot satisfy the vector.
func TestMuseRootContextVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/muse-root-context.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Cases []museRootVector `json:"cases"`
	}
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Cases) != 4 {
		t.Fatalf("want 4 admission vectors, got %d", len(vectors.Cases))
	}
	for _, tc := range vectors.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			source := museRootProfile(t, tc.Context)
			source.cfg.Env.Forms = map[string]string{envregistry.Muse: tc.Form}
			if tc.Name == "no-context" {
				// Unrelated native/workspace/header-like files cannot mint policy.
				for _, path := range []string{filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "muse", "AGENTS.md"), filepath.Join(source.cfg.Home(), "AGENTS.md")} {
					if err := os.WriteFile(path, []byte("<!-- curator-root-context-v2 -->\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			code, stdout, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair", "--format", "json")
			home := envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.Muse)
			target := filepath.Join(home, "config", "muse", "AGENTS.md")
			if tc.Diagnostic != "" {
				if code != exitFail || stdout != "" || !strings.Contains(stderr, tc.Diagnostic) || strings.Contains(stderr, contextmaterialize.DiagFormUnavailable) {
					t.Fatalf("referenced must refuse without fallback: exit %d stdout=%q stderr=%q", code, stdout, stderr)
				}
				if _, err := os.Lstat(target); !os.IsNotExist(err) {
					t.Fatalf("refused form wrote root target: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(home, envmarker.Name)); !os.IsNotExist(err) {
					t.Fatalf("refused form published a marker: %v", err)
				}
				return
			}
			if code != exitOK {
				t.Fatalf("resolve: exit %d: %s", code, stderr)
			}
			var fragment map[string]any
			if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
				t.Fatal(err)
			}
			fact, present := fragment["root_context"]
			if present != tc.Written || present && !reflect.DeepEqual(fact, map[string]any{"exclusive": true}) {
				t.Fatalf("root fact disagrees with verified surface: %s", stdout)
			}
			if _, err := os.Lstat(filepath.Join(home, "AGENTS.md")); !os.IsNotExist(err) {
				t.Fatalf("root context written outside the Muse XDG target: %v", err)
			}
			marker, err := envmarker.Read(home)
			if err != nil || marker == nil {
				t.Fatalf("marker: %v", err)
			}
			surface, recorded := marker.Surfaces[envmarker.SurfaceRootContext]
			got, err := os.ReadFile(target)
			if !tc.Written {
				if !os.IsNotExist(err) || recorded {
					t.Fatalf("no-context profile wrote a root surface: recorded=%t err=%v", recorded, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			lock, hash, err := contextlock.Read(filepath.Join(envprofile.ProfileDir(source.cfg.Home(), "acme"), "lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			root, ok := lock.RootMember()
			if !ok {
				t.Fatal("missing lock root")
			}
			want := fmt.Sprintf("<!--\ncurator-root-context-v2\nroot: acme 1.0.0 %s\nmember: acme 1.0.0 %s weight 0\nprecedence: winner=higher-weight placement=winner-last\nlock: %s\ngenerated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)\nnotice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead\n-->\n", root.Pin(), root.Pin(), hash)
			if tc.Body != "" {
				want += "\n---\n\n## Context: acme 1.0.0\n\n" + tc.Body
			}
			if !bytes.Equal(got, []byte(want)) {
				t.Fatalf("root bytes differ: got %q want %q", got, want)
			}
			if !recorded || surface.Form != "monolithic" || len(surface.Paths) != 1 || surface.Paths[0] != "config/muse/AGENTS.md" {
				t.Fatalf("root surface not recorded at admitted target: %+v", surface)
			}
			if info, err := os.Lstat(target); err != nil {
				t.Fatal(err)
			} else if info.Mode()&os.ModeSymlink == 0 {
				copied := false
				if surface.Copies != nil {
					for _, copy := range *surface.Copies {
						copied = copied || copy.Path == "config/muse/AGENTS.md" && copy.Reason == envmarker.ReasonSymlinkFallback
					}
				}
				if !copied {
					t.Fatal("regular root lacks fallback record")
				}
				t.Log("verified copied root")
			}
			wantHash, err := contextmaterialize.SurfaceHashWithVersion(map[string][]byte{"config/muse/AGENTS.md": []byte(want)}, lock.ContentHashVersion())
			if err != nil || surface.ContentSHA256 != wantHash {
				t.Fatalf("root marker hash: %q want %q err=%v", surface.ContentSHA256, wantHash, err)
			}
			code, current, stderr := runProfile(t, source, "env", "resolve", "muse")
			if code != exitOK {
				t.Fatalf("unchanged root is not current: exit %d: %s", code, stderr)
			}
			var currentFragment struct {
				RootContext *struct{ Exclusive bool } `json:"root_context"`
			}
			if err := json.Unmarshal([]byte(current), &currentFragment); err != nil {
				t.Fatal(err)
			}
			if currentFragment.RootContext == nil || !currentFragment.RootContext.Exclusive {
				t.Fatal("current root lost verified fact")
			}
		})
	}
}

// Both drift classes drive the read-only CLI refusal and StatusOf before repair.
func TestMuseRootContextDrift(t *testing.T) {
	for _, action := range []string{"changed", "removed"} {
		t.Run(action, func(t *testing.T) {
			source := museRootProfile(t, json.RawMessage(`{"modules":[{"path":"a.md"}]}`))
			if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
				t.Fatalf("provision: exit %d: %s", code, stderr)
			}
			home := envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.Muse)
			target := filepath.Join(home, "config", "muse", "AGENTS.md")
			original, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			markerPath := filepath.Join(home, envmarker.Name)
			markerBefore, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}
			state := envprofile.DiagSurfaceDrift
			if action == "changed" {
				err = os.WriteFile(target, []byte("changed\n"), 0600)
			} else {
				err = os.Remove(target)
				state = envprofile.DiagSurfaceMissing
			}
			if err != nil {
				t.Fatal(err)
			}
			if code, stdout, stderr := runProfile(t, source, "env", "resolve", "muse", "--format", "json"); code != exitFail || stdout != "" || !strings.Contains(stderr, envprofile.DiagHomeStale) {
				t.Fatalf("drift must refuse fragment: exit %d stdout=%q stderr=%q", code, stdout, stderr)
			}
			status := readMuseRootStatus(t, source)
			found := false
			for _, row := range status.Homes {
				if row.Profile != "acme" || row.Environment != envregistry.Muse {
					continue
				}
				if row.Current {
					t.Fatal("drifted Muse home reported current")
				}
				for _, surface := range row.Surfaces {
					if surface.Key == envmarker.SurfaceRootContext && surface.State == state {
						found = true
					}
				}
			}
			if !found {
				t.Fatalf("missing Muse %s status", state)
			}
			if got, err := os.ReadFile(markerPath); err != nil || !bytes.Equal(got, markerBefore) {
				t.Fatal("read-only refusal/status changed marker")
			}
			got, err := os.ReadFile(target)
			if action == "changed" && (err != nil || string(got) != "changed\n") || action == "removed" && !os.IsNotExist(err) {
				t.Fatal("read-only refusal/status changed root")
			}
			if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
				t.Fatalf("repair: exit %d: %s", code, stderr)
			}
			repaired, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(repaired, original) {
				t.Fatalf("repair bytes: %q %v", repaired, err)
			}
		})
	}
}

// Permission denial is exercised on hosts that enforce chmod; a blocked
// repair uses a nonempty target directory on every platform. Neither path
// may emit a fragment or publish a new marker; the declared root is retained.
func TestMuseRootContextRefusals(t *testing.T) {
	for _, action := range []string{"denied", "failed-repair"} {
		t.Run(action, func(t *testing.T) {
			source := museRootProfile(t, json.RawMessage(`{"modules":[{"path":"a.md"}]}`))
			if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
				t.Fatal(stderr)
			}
			home := envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.Muse)
			target := filepath.Join(home, "config", "muse", "AGENTS.md")
			markerPath := filepath.Join(home, envmarker.Name)
			markerBefore, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"env", "resolve", "muse", "--format", "json"}
			want := envprofile.DiagHomeStale
			if action == "denied" {
				if runtime.GOOS == "windows" {
					t.Skip("POSIX mode-bit unreadability: chmod cannot establish Windows ACL denial")
				}
				if err := os.Chmod(target, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(target, 0600) })
				if _, err := os.ReadFile(target); !os.IsPermission(err) {
					t.Skip("host does not enforce chmod read denial")
				}
			} else {
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "preserve"), []byte("foreign\n"), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--repair")
				want = envprofile.DiagRepairFailed
			}
			if code, stdout, stderr := runProfile(t, source, args...); code != exitFail || stdout != "" || !strings.Contains(stderr, want) {
				t.Fatalf("required-root refusal: %d %q %s", code, stdout, stderr)
			}
			status := readMuseRootStatus(t, source)
			found := false
			for _, row := range status.Homes {
				if row.Environment != envregistry.Muse || row.Profile != "acme" {
					continue
				}
				if row.Current {
					t.Fatal("unusable required root is current")
				}
				for _, surface := range row.Surfaces {
					if surface.Key == envmarker.SurfaceRootContext {
						found = true
						if action == "denied" && surface.State != envprofile.DiagSurfaceUnreadable {
							t.Fatalf("denial became %s", surface.State)
						}
					}
				}
			}
			if !found {
				t.Fatal("required root silently disappeared from status")
			}
			if got, err := os.ReadFile(markerPath); err != nil || !bytes.Equal(got, markerBefore) {
				t.Fatal("refusal/status published a marker")
			}
			if action == "denied" {
				if err := os.Chmod(target, 0600); err != nil {
					t.Fatal(err)
				}
				if got, err := os.ReadFile(target); err != nil || !bytes.Equal(got, original) {
					t.Fatal("refusal changed denied root")
				}
			} else {
				if got, err := os.ReadFile(filepath.Join(target, "preserve")); err != nil || string(got) != "foreign\n" {
					t.Fatal("failed repair changed blocked target")
				}
				if err := os.Remove(filepath.Join(target, "preserve")); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			}
			// The adjacent recovery control must emit the fact, not just succeed.
			code, stdout, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair", "--format", "json")
			var fragment struct {
				RootContext *struct{ Exclusive bool } `json:"root_context"`
			}
			if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
				t.Fatal(err)
			}
			if code != exitOK || fragment.RootContext == nil || !fragment.RootContext.Exclusive {
				t.Fatalf("recovery lost fact: %d %s %s", code, stdout, stderr)
			}
		})
	}
}

// Shared and isolated status reuse the verified policy, report effective
// file/mode, preserve config bytes and never print synthetic credential bytes.
func TestMuseCredentialPolicyStatus(t *testing.T) {
	for _, mode := range []string{"shared", "isolated"} {
		t.Run(mode, func(t *testing.T) {
			source := museRootProfile(t, nil)
			source.cfg.Env.Isolation = map[string]map[string]string{"acme": {envregistry.Muse: mode}}
			if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
				t.Fatal(stderr)
			}
			configBefore, err := json.Marshal(source.cfg)
			if err != nil {
				t.Fatal(err)
			}
			status := readMuseRootStatus(t, source)
			found := false
			for _, row := range status.Homes {
				if row.Profile == "acme" && row.Environment == envregistry.Muse {
					found = row.Current && strings.Contains(strings.Join(row.Warnings, "\n"), "effective backend=file; isolation="+mode)
				}
			}
			if !found {
				t.Fatalf("status omitted verified %s file policy", mode)
			}
			if raw, _ := json.Marshal(status); strings.Contains(string(raw), "synthetic-native") {
				t.Fatal("status leaked credential bytes")
			}
			if got, err := json.Marshal(source.cfg); err != nil || !bytes.Equal(got, configBefore) {
				t.Fatal("status changed config")
			}
		})
	}
}

func readMuseRootStatus(t *testing.T, source stubConfigSource) envprofile.Status {
	t.Helper()
	code, stdout, stderr := runProfile(t, source, "env", "status", "--json")
	if code != exitOK {
		t.Fatalf("status: exit %d: %s", code, stderr)
	}
	var status envprofile.Status
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatal(err)
	}
	return status
}

// Public status separates historical loader evidence, backend evidence and
// unqualified replay/refresh; it reports no runtime qualification of its own.
func TestMuseRootContextStatus(t *testing.T) {
	source := museRootProfile(t, nil)
	notes := strings.Join(readMuseRootStatus(t, source).Notes, "\n")
	for _, text := range []string{"monolithic root context delivered at config/muse/AGENTS.md", "verified on 1.4.7", "1.4.7-R6101.1", "suppresses foreign personal rules fallback", "foreign personal skills persist", "exec requires --no-foreign-personal-context when the target is materialized", "interactive requires -c context.foreign_personal_skills=false (JSON boolean)", "serve cannot exclude foreign personal context", "required serve exclusion refuses", "workspace AGENTS.md prevailed", "1.4.8 on macOS 27: file backend", "zero Keychain access", "1.4.8 exclusion qualification unknown", "exact build", "P1–P4", "both forms", "root present/absent", "foreign rules/skills separately", "each exclusion spelling", "Curator link and regular-file snapshot", "loader observations separately from model-visible behavior", "model-visible foreign removal", "skills discovery", "system-prompt discovery", "refresh semantics", "remain unverified"} {
		if !strings.Contains(notes, text) {
			t.Fatalf("status omitted %q: %s", text, notes)
		}
	}
	if strings.Contains(notes, "root context, skills discovery") {
		t.Fatal("status still claims root context is unverified")
	}
}

// Go overlays weaken one production behavior at a time without editing the
// candidate. Each expected-red subprocess must fail its named behavioral test;
// compile failures and unrelated failures are not mutant kills.
func TestMuseRootContextNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	mutants := []struct{ name, file, before, after, test string }{
		{"configured-form-ignored", "cmd/curator/env.go", "machine.Forms = cfg.Env.Forms", "machine.Forms = cfg.Env.Forms\nif machine.Forms[envregistry.Muse] == envregistry.FormReferenced { machine.Forms = map[string]string{}; for id, form := range cfg.Env.Forms { if id != envregistry.Muse { machine.Forms[id] = form } } }", "TestMuseRootContextVectors/referenced-refused"},
		{"referenced-fallback", "internal/envregistry/envregistry.go", "func (a Adapter) ResolveForm(configured string) (string, error) {", "func (a Adapter) ResolveForm(configured string) (string, error) {\nif a.ID == Muse && configured == FormReferenced { return FormMonolithic, nil }", "TestMuseRootContextVectors/referenced-refused"},
		{"no-context-admitted", "internal/contextmaterialize/contextmaterialize.go", "if !rootPkg.HasContext {", "if !rootPkg.HasContext && environment != \"muse\" {", "TestMuseRootContextVectors/no-context"},
		{"wrong-target", "internal/envregistry/envregistry.go", "RootTarget:      \"config/muse/AGENTS.md\",", "RootTarget:      \"AGENTS.md\",", "TestMuseRootContextVectors/monolithic"},
		{"wrong-bytes", "internal/contextmaterialize/contextmaterialize.go", "parts := [][]byte{header}", "if environment == \"muse\" { header = append(header, []byte(\"unexpected\\n\")...) }; parts := [][]byte{header}", "TestMuseRootContextVectors/zero-applicable"},
		{"changed-root-admitted", "internal/envprofile/managed.go", "for _, path := range recorded.Paths {\n\t\t\tfull :=", "for _, path := range recorded.Paths {\nif plan.adapter.ID == envregistry.Muse && path == \"config/muse/AGENTS.md\" { b, e := os.ReadFile(filepath.Join(plan.homeDir, path)); if e == nil && string(b) == \"changed\\n\" { continue } }\n\t\t\tfull :=", "TestMuseRootContextDrift/changed"},
		{"removed-root-admitted", "internal/envprofile/managed.go", "for _, path := range recorded.Paths {\n\t\t\tfull :=", "for _, path := range recorded.Paths {\nif plan.adapter.ID == envregistry.Muse && path == \"config/muse/AGENTS.md\" { m, e := stateread.Lstat(filepath.Join(plan.homeDir, path)); if e == nil && m.Kind == stateread.KindAbsent { continue } }\n\t\t\tfull :=", "TestMuseRootContextDrift/removed"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			original := filepath.Join(root, filepath.FromSlash(mutant.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			// Windows checkouts may use CRLF; overlays change Go behavior,
			// independently of the checkout's line-ending convention.
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			parts := strings.Split(mutant.test, "/")
			command := exec.Command(testtoolchain.Go(t), "test", "-overlay", overlayPath, "-count=1", "-timeout=5m", "-v", "-run", "^"+parts[0]+"$/^"+parts[1]+"$", ".")
			// Reuse the CLI helper protocol: the parent already holds the
			// package GOROOT lock, and these behavioral tests do not mutate it.
			// TestMain still isolates git in the child process.
			command.Env = append(os.Environ(), cliHelperProcess+"=1")
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() == 0 || !strings.Contains(string(output), "--- FAIL: "+mutant.test) {
				t.Fatalf("mutant survived or lacked named behavioral failure %s: %v\n%s", mutant.test, err, output)
			}
			t.Logf("%s: expected-red go test exit %d; killed by %s", mutant.name, exit.ExitCode(), mutant.test)
		})
	}
}
