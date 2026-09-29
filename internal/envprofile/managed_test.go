// Production entry points under test: Resolve (read-only verify,
// stale refusal, repair, provision) and the marker records behind it.
package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

// managedFixture installs a hand-built profile: a context root with a
// root and a system module, one MCP server, and one skill. Store entries
// are installed through the product store helpers; the lock is canonical.
type managedFixture struct {
	home     string
	profile  string
	native   map[string]string
	xdg      string
	launch   string
	lockHash string
}

func writeManagedFixture(t *testing.T, profile string) *managedFixture {
	t.Helper()
	home := t.TempDir()
	native := map[string]string{}
	for _, id := range []string{"claude_code", "codex_cli", "opencode-native", "pi"} {
		native[id] = t.TempDir()
	}
	fx := &managedFixture{home: home, profile: profile, native: native, xdg: t.TempDir(), launch: t.TempDir()}
	sourceDir := filepath.Join(t.TempDir(), profile+"-source")
	if err := os.Mkdir(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, sourceDir, "agent-context.json", `{"schema_version": 1, "name": "`+profile+`", "version": "1.0.0", "context": {"modules": [{"path": "a.md"}, {"path": "s.md", "class": "system"}]}}`)
	writeGitFile(t, sourceDir, "context/a.md", "hello\n")
	writeGitFile(t, sourceDir, "context/s.md", "Be terse.\n")
	if err := pathboundary.ProtectTree(sourceDir); err != nil {
		t.Fatalf("protect path source fixture: %v", err)
	}
	_, ctxPin, err := contextstore.EnsureState(home, "context", profile, sourceDir)
	if err != nil {
		t.Fatalf("create context snapshot fixture: %v", err)
	}
	mcpSource := "github.com/example/mcp-figma"
	mcpPin := writeGitStoreFixture(t, home, "mcp", "figma-devmode", mcpSource, map[string]string{
		"agent-mcp.json":      `{"schema_version": 1, "name": "figma-devmode", "version": "1.2.0", "server": {"transport": "stdio", "command": "npx", "args": ["-y", "figma-developer-mcp", "--stdio"], "env_names": ["FIGMA_API_KEY"], "environments": ["claude_code", "codex_cli", "opencode"]}}`,
		"unused/pin-only.txt": "not surfaced\n",
	})
	skillSource := "github.com/example/skill-my"
	skillPin := writeGitStoreFixture(t, home, "skill", "myskill", skillSource, map[string]string{
		"SKILL.md": "# myskill\n",
	})
	lock := &contextlock.Lock{
		Root: profile,
		Members: []contextlock.Member{
			{Kind: "context", Name: profile, Version: "1.0.0", StateHash: ctxPin, Weight: 100},
			{Kind: "mcp", Name: "figma-devmode", Source: mcpSource, Version: "1.2.0", Commit: mcpPin, RequiredBy: []string{profile}},
			{Kind: "skill", Name: "myskill", Source: skillSource, Commit: skillPin, RequiredBy: []string{profile}},
		},
	}
	lock.Sort()
	hash, err := contextlock.Write(lockPath(home, profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	fx.lockHash = hash
	source, err := json.Marshal(Source{Kind: KindPath, Path: sourceDir})
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, '\n')
	if err := privatedir.MakeAll(ProfileDir(home, profile)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath(home, profile), source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetCurrent(home, profile); err != nil {
		t.Fatal(err)
	}
	return fx
}

func writeGitStoreFixture(t *testing.T, home, kind, name, source string, files map[string]string) string {
	t.Helper()
	repo := newGitManager(home).repoDir(source)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "init", "--quiet")
	for path, content := range files {
		writeGitFile(t, repo, path, content)
	}
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "--quiet", "-m", "fixture")
	commit := fixtureGitOutput(t, repo, "rev-parse", "HEAD")
	if _, err := contextstore.EnsureGit(home, kind, name, repo, commit); err != nil {
		t.Fatalf("create %s/%s Git snapshot fixture: %v", kind, name, err)
	}
	return commit
}

func fixtureGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// seedLiveNativeCredentials writes live native credential targets for
// every file-linked adapter — codex auth.json, the pi agent-root
// auth.json, and the Linux claude_code credentials file — so tests about
// other behaviors provision homes without the detached-pending finding.
// The writes are harmless where an adapter links nothing on the platform.
func seedLiveNativeCredentials(t *testing.T, fx *managedFixture) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	agentAuth := filepath.Join(fx.native["pi"], "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fx.native["claude_code"], ".credentials.json"), []byte("{\"t\":\"operator-claude\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func (fx *managedFixture) request(envID string) ResolveRequest {
	return ResolveRequest{
		Home:      fx.home,
		Profile:   fx.profile,
		EnvID:     envID,
		LaunchDir: fx.launch,
		Machine:   envregistry.DefaultMachineConfig(),
		Detect:    func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) {
			if id == "opencode" {
				return fx.native["opencode-native"], nil
			}
			return fx.native[id], nil
		},
		OperatorXDG: fx.xdg,
	}
}

func readManagedMarker(t *testing.T, fx *managedFixture, envID string) *envmarker.Marker {
	t.Helper()
	marker, err := envmarker.Read(ManagedHomeDir(fx.home, fx.profile, envID))
	if err != nil {
		t.Fatal(err)
	}
	if marker == nil {
		t.Fatal("managed home carries no marker")
	}
	return marker
}

func TestResolveVerifiesGitStoreEntryAgainstPinnedTree(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision := fx.request("codex_cli")
	provision.Repair = true
	if _, err := Resolve(provision); err != nil {
		t.Fatalf("provision intact fixture: %v", err)
	}

	result, err := Resolve(fx.request("codex_cli"))
	if err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("Resolve with matching Git tree = (%v, %v), want a fragment", result, err)
	}

	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
	if !ok {
		t.Fatal("fixture lock has no Git MCP member")
	}
	entry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	if err := os.WriteFile(filepath.Join(entry, "unused", "pin-only.txt"), []byte("swapped sidecar\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err = Resolve(fx.request("codex_cli"))
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) ||
		!strings.Contains(err.Error(), "mcp/figma-devmode") || !strings.Contains(err.Error(), "pin_hash") {
		t.Fatalf("Resolve after swapping Git entry emitted fragment=%t, error=%v; want named pin failure",
			result != nil && len(result.Document) != 0, err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatalf("untrusted Git store emitted a fragment: %s", result.Document)
	}
}

func TestResolveRejectsGitPinMissingFromLocalObjectDatabase(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(contextlock.KindMCP, "figma-devmode")
	if !ok {
		t.Fatal("fixture lock has no Git MCP member")
	}
	oldEntry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	alternate := filepath.Join(t.TempDir(), "alternate-object-database")
	if err := os.Mkdir(alternate, 0o700); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, alternate, "agent-mcp.json", `{"schema_version": 1, "name": "figma-devmode", "version": "1.2.0", "server": {"transport": "stdio", "command": "npx", "args": ["-y", "figma-developer-mcp", "--stdio"], "env_names": ["FIGMA_API_KEY"], "environments": ["claude_code", "codex_cli", "opencode"]}}`)
	writeGitFile(t, alternate, "unused/pin-only.txt", "not surfaced\n")
	gitRun(t, alternate, "init", "--quiet")
	gitRun(t, alternate, "add", "--all")
	gitRun(t, alternate, "commit", "--quiet", "-m", "alternate object database")
	member.Commit = fixtureGitOutput(t, alternate, "rev-parse", "HEAD")
	newEntry := contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
	if err := os.Rename(oldEntry, newEntry); err != nil {
		t.Fatalf("move fixture entry to missing pin: %v", err)
	}
	for i := range lock.Members {
		if lock.Members[i].Kind == member.Kind && lock.Members[i].Name == member.Name {
			lock.Members[i] = member
		}
	}
	lock.Sort()
	if _, err := contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
		t.Fatal(err)
	}

	before := hashTreeForTest(t, fx.home)
	req := fx.request("codex_cli")
	req.Repair = true
	t.Setenv("GIT_DIR", filepath.Join(alternate, ".git"))
	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) ||
		!strings.Contains(err.Error(), "mcp/figma-devmode") || !strings.Contains(err.Error(), "pinned commit object unavailable") {
		t.Fatalf("Resolve without local pinned object emitted fragment=%t, error=%v; want fail-closed missing-object diagnostic",
			result != nil && len(result.Document) != 0, err)
	}
	if strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("Resolve attempted repair without the pinned object: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatalf("missing pinned object emitted a fragment: %s", result.Document)
	}
	if after := hashTreeForTest(t, fx.home); after != before {
		t.Fatal("resolve changed manager state while the pinned object was unavailable")
	}
}

// TestResolveProvisionRepair drives the production Resolve from
// provisioning through a current bare resolve: the first --repair
// provisions, records the marker, prints the notice, and emits a fragment
// whose env names the managed home; the bare resolve that follows is
// current and emits identical bytes without repair.
func TestResolveProvisionRepair(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("codex_cli")
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Provisioned || result.Notice == "" {
		t.Fatal("the first repair provisions and prints the first-resolve notice")
	}
	if !strings.Contains(result.Notice, ManagedHomeDir(fx.home, "acme", "codex_cli")) {
		t.Fatalf("the notice names the managed-home path: %q", result.Notice)
	}
	var fragment map[string]any
	if err := json.Unmarshal(result.Document, &fragment); err != nil {
		t.Fatalf("fragment is not JSON: %v", err)
	}
	env := fragment["env"].(map[string]any)
	if env["CODEX_HOME"] != ManagedHomeDir(fx.home, "acme", "codex_cli") {
		t.Fatalf("fragment env %v names the wrong home", env)
	}
	if fragment["profile"].(map[string]any)["lock_sha256"] != strings.TrimPrefix(fx.lockHash, "sha256:") {
		t.Fatal("fragment carries the wrong lock hash")
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	managedHome := ManagedHomeDir(fx.home, "acme", "codex_cli")
	for _, dir := range []string{EnvRoot(fx.home), managedHome} {
		if err := privatedir.Validate(dir); err != nil {
			t.Fatalf("managed directory %s is not owner-only: %v", dir, err)
		}
	}
	markerPath := filepath.Join(managedHome, envmarker.Name)
	if err := privatedir.ValidateFile(markerPath); err != nil {
		t.Fatalf("published marker is not owner-only: %v", err)
	}
	if marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != "acme" {
		t.Fatalf("marker mode/profile %+v", marker)
	}
	if _, ok := marker.Surfaces["root-context"]; !ok {
		t.Fatal("marker records no root-context surface")
	}
	if _, ok := marker.Surfaces["system-prompt"]; !ok {
		t.Fatal("marker records no system-prompt surface for a chain with system modules")
	}
	if _, ok := marker.Surfaces["mcp"]; !ok {
		t.Fatal("marker records no mcp surface for a non-empty set")
	}
	if _, ok := marker.Surfaces["skills"]; !ok {
		t.Fatal("marker records no skills surface")
	}
	bare, err := Resolve(fx.request("codex_cli"))
	if err != nil {
		t.Fatalf("a provisioned home resolves bare: %v", err)
	}
	if bare.Provisioned || string(bare.Document) != string(result.Document) {
		t.Fatal("a current bare resolve emits identical bytes without provisioning")
	}
}

// TestResolveReportsRecordedContextSurfaceHashDrift proves that Resolve
// compares the marker's root-context and system-prompt hashes with the
// surfaces produced from the verified lock, even when their link targets
// still have the expected paths.
func TestResolveReportsRecordedContextSurfaceHashDrift(t *testing.T) {
	for _, surface := range []string{envmarker.SurfaceRootContext, envmarker.SurfaceSystemPrompt} {
		t.Run(surface, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			seedLiveNativeCredentials(t, fx)
			provision := fx.request("codex_cli")
			provision.Repair = true
			if _, err := Resolve(provision); err != nil {
				t.Fatalf("provision managed home: %v", err)
			}
			marker := readManagedMarker(t, fx, "codex_cli")
			recorded, ok := marker.Surfaces[surface]
			if !ok {
				t.Fatalf("fixture marker has no %s surface", surface)
			}
			recorded.ContentSHA256 = "sha256:" + strings.Repeat("0", 64)
			marker.Surfaces[surface] = recorded
			payload, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envmarker.Name)
			if err := os.WriteFile(markerPath, payload, 0o600); err != nil {
				t.Fatalf("change recorded surface hash: %v", err)
			}

			result, err := Resolve(fx.request("codex_cli"))
			if err == nil || !strings.Contains(err.Error(), DiagHomeStale) || !strings.Contains(err.Error(), "surface "+surface+" record does not match") {
				t.Fatalf("Resolve emitted fragment=%t, error=%v; want stale surface-hash drift for %s",
					result != nil && len(result.Document) != 0, err, surface)
			}
			if result == nil || len(result.Document) != 0 {
				t.Fatalf("drifted resolve emitted a fragment: %+v", result)
			}

			status, err := StatusOf(statusRequest(fx))
			if err != nil {
				t.Fatalf("StatusOf: %v", err)
			}
			var row *HomeState
			for i := range status.Homes {
				if status.Homes[i].Profile == fx.profile && status.Homes[i].Environment == "codex_cli" {
					row = &status.Homes[i]
					break
				}
			}
			if row == nil || row.Current {
				t.Fatalf("status row = %+v, want non-current drift", row)
			}
			foundDrift := false
			for _, state := range row.Surfaces {
				if state.Key == surface && state.State == DiagSurfaceDrift {
					foundDrift = true
				}
			}
			if !foundDrift {
				t.Fatalf("status row does not report %s as %s: %+v", surface, DiagSurfaceDrift, row.Surfaces)
			}
		})
	}
}

// TestResolveStaleUnprovisioned narrows the fail-closed gate: without
// --repair a stale home reports environment_home_stale with reasons and
// emits no fragment. Weakening the gate to emit anyway must fail this
// test.
func TestResolveStaleUnprovisioned(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	result, err := Resolve(fx.request("codex_cli"))
	if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("an unprovisioned bare resolve must fail with %s, got %v", DiagHomeStale, err)
	}
	if result == nil || len(result.StaleReasons) == 0 {
		t.Fatal("staleness carries reasons")
	}
	if result.Document != nil {
		t.Fatal("a stale resolve emits no fragment")
	}
}

// TestResolveUnknowns covers the operand gates through Resolve.
func TestResolveUnknowns(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("cursor")
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagUnknown) {
		t.Fatalf("an unregistered environment must fail, got %v", err)
	}
	req = fx.request("codex_cli")
	req.Profile = "ghost"
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), DiagProfileUnknown) {
		t.Fatalf("an uninstalled profile must fail with %s, got %v", DiagProfileUnknown, err)
	}
}

// TestResolveDefaultProfile resolves the machine current when no profile
// is named.
func TestResolveDefaultProfile(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("codex_cli")
	req.Profile = ""
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("an unnamed resolve uses the current profile: %v", err)
	}
	if result.Document == nil {
		t.Fatal("no fragment emitted")
	}
}

// TestResolveDriftRepair dirties a copied surface and a link surface, and
// proves repair restores managed bytes from the store without adopting
// candidate bytes.
func TestResolveDriftRepair(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("claude_code")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	homeDir := ManagedHomeDir(fx.home, "acme", "claude_code")
	if err := os.WriteFile(filepath.Join(homeDir, "CLAUDE.md"), []byte("operator bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(fx.request("claude_code")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("a drifted copy must be stale, got %v", err)
	}
	repaired, err := Resolve(req)
	if err != nil {
		t.Fatalf("repair restores: %v", err)
	}
	if repaired.Provisioned {
		t.Fatal("repair is not provisioning")
	}
	payload, err := os.ReadFile(filepath.Join(homeDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "operator bytes") {
		t.Fatal("repair adopted candidate bytes from the home")
	}
	if _, err := Resolve(fx.request("claude_code")); err != nil {
		t.Fatalf("the repaired home is current: %v", err)
	}
	// A replaced store link is drift of the same class.
	codex := fx.request("codex_cli")
	codex.Repair = true
	if _, err := Resolve(codex); err != nil {
		t.Fatal(err)
	}
	codexHome := ManagedHomeDir(fx.home, "acme", "codex_cli")
	_ = os.Remove(filepath.Join(codexHome, "AGENTS.md"))
	if err := os.WriteFile(filepath.Join(codexHome, "AGENTS.md"), []byte("candidate\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(fx.request("codex_cli")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("a replaced link must be stale, got %v", err)
	}
	// A write through an intact link into a manager-authored rendered
	// document is the same §8.4 class ("a target whose bytes fail the
	// recorded hash"): the link is unchanged, the bytes are not. Cover
	// every adapter and every rendered surface in both supported forms;
	// links into the immutable store (skills trees, referenced module
	// files) keep the link-identity fast path and are excluded here.
	for _, tc := range []struct {
		env  string
		form string
	}{
		{"claude_code", ""},
		{"claude_code", "referenced"},
		{"codex_cli", ""},
		{"opencode", ""},
		{"opencode", "referenced"},
		{"pi", ""},
	} {
		probe := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, probe)
		provision := probe.request(tc.env)
		if tc.form != "" {
			provision.Machine.Forms = map[string]string{tc.env: tc.form}
		}
		provision.Repair = true
		if _, err := Resolve(provision); err != nil {
			t.Fatalf("%s %s provision: %v", tc.env, tc.form, err)
		}
		marker := readManagedMarker(t, probe, tc.env)
		homeDir := ManagedHomeDir(probe.home, "acme", tc.env)
		rendered := 0
		for _, surface := range marker.Surfaces {
			for _, rel := range surface.Paths {
				full := filepath.Join(homeDir, filepath.FromSlash(rel))
				info, err := os.Lstat(full)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					continue
				}
				target, err := os.Readlink(full)
				if err != nil || !strings.Contains(filepath.ToSlash(target), "/rendered/") {
					continue
				}
				rendered++
				// Write through the intact link: the link bytes are
				// unchanged, the target bytes drift.
				if err := os.WriteFile(full, []byte("tampered through intact link\n"), 0o644); err != nil {
					t.Fatalf("%s %s %s write through link: %v", tc.env, tc.form, rel, err)
				}
				check := probe.request(tc.env)
				if tc.form != "" {
					check.Machine.Forms = map[string]string{tc.env: tc.form}
				}
				if _, err := Resolve(check); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
					t.Fatalf("%s %s %s byte drift through an intact link must be stale, got %v", tc.env, tc.form, rel, err)
				}
				repair := probe.request(tc.env)
				if tc.form != "" {
					repair.Machine.Forms = map[string]string{tc.env: tc.form}
				}
				repair.Repair = true
				if _, err := Resolve(repair); err != nil {
					t.Fatalf("%s %s %s repair restores: %v", tc.env, tc.form, rel, err)
				}
				if _, err := Resolve(check); err != nil {
					t.Fatalf("%s %s %s repaired home is current: %v", tc.env, tc.form, rel, err)
				}
			}
		}
		if rendered == 0 {
			t.Fatalf("%s %s matched no rendered symlink surface", tc.env, tc.form)
		}
	}
}

// TestResolvePassthroughLiveness removes a file-link passthrough entry and
// proves the liveness row goes stale and --repair re-links the absent
// link. A severed link replaced by a regular file is the separate
// environment_credential_conflict refusal (Decision 0017, §7.4/§10.1):
// repair never removes or replaces the file. That refusal half moved to
// TestCredentialLinkRegularFileRefuses with the 0017 contract; this test
// keeps the absent-link re-link half.
func TestResolvePassthroughLiveness(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte(`{"openai_api_key":null}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	if target, err := os.Readlink(link); err != nil || target != filepath.Join(fx.native["codex_cli"], "auth.json") {
		t.Fatalf("passthrough link targets %q (%v)", target, err)
	}
	_ = os.Remove(link)
	if _, err := Resolve(fx.request("codex_cli")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("a detached passthrough must be stale, got %v", err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("repair re-links: %v", err)
	}
	if target, err := os.Readlink(link); err != nil || target != filepath.Join(fx.native["codex_cli"], "auth.json") {
		t.Fatalf("repair left %q (%v)", target, err)
	}
}

// TestClaudeRootAlwaysCopied narrows the §8.1 gate: the claude_code
// root-context surface is a copied regular file whatever the form. A
// mutant that links CLAUDE.md under the referenced form must fail this
// test.
func TestClaudeRootAlwaysCopied(t *testing.T) {
	for _, form := range []string{"", "referenced"} {
		label := form
		if label == "" {
			label = "monolithic"
		}
		fx := writeManagedFixture(t, "acme")
		req := fx.request("claude_code")
		if form != "" {
			req.Machine.Forms = map[string]string{"claude_code": form}
		}
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatalf("%s provision: %v", label, err)
		}
		homeDir := ManagedHomeDir(fx.home, "acme", "claude_code")
		info, err := os.Lstat(filepath.Join(homeDir, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("%s CLAUDE.md missing: %v", label, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("%s CLAUDE.md is a symlink, want a copied regular file", label)
		}
		marker := readManagedMarker(t, fx, "claude_code")
		surface := marker.Surfaces["root-context"]
		if surface.Copies == nil {
			t.Fatalf("%s root-context records no copies", label)
		}
		found := false
		for _, copy := range *surface.Copies {
			if copy.Path == "CLAUDE.md" && copy.Reason == envmarker.ReasonClaudeCodeRootContext {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s root-context records no claude-code-root-context copy for CLAUDE.md: %+v", label, surface.Copies)
		}
	}
}

// TestResolveClaudeProjectEntry proves the referenced-form project-entry
// rule: a launch directory without its entry is stale for that directory,
// and --repair adds the entry with the external-includes key.
func TestResolveClaudeProjectEntry(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("claude_code")
	req.Machine.Forms = map[string]string{"claude_code": "referenced"}
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	marker := readManagedMarker(t, fx, "claude_code")
	if surface := marker.Surfaces["root-context"]; surface.Form != "referenced" {
		t.Fatalf("marker form %q", surface.Form)
	}
	found := false
	for _, name := range marker.SeededProjects {
		if name == fx.launch {
			found = true
		}
	}
	if !found {
		t.Fatal("the launch directory is not in seeded_projects")
	}
	other := fx.request("claude_code")
	other.Machine.Forms = map[string]string{"claude_code": "referenced"}
	other.LaunchDir = t.TempDir()
	if _, err := Resolve(other); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("a launch directory without its entry must be stale, got %v", err)
	}
	other.Repair = true
	if _, err := Resolve(other); err != nil {
		t.Fatalf("repair adds the entry: %v", err)
	}
	entries, err := claudeProjects(ManagedHomeDir(fx.home, "acme", "claude_code"))
	if err != nil || !entries[other.LaunchDir] {
		t.Fatal("the repaired entry is missing")
	}
	// The tool owns .claude.json after provisioning (§7.4): dropping only
	// the approval key while keeping the project entry must still go
	// stale, otherwise every @-include is silently discarded (§5.3).
	homeDir := ManagedHomeDir(fx.home, "acme", "claude_code")
	payload, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	projects, ok := object["projects"].(map[string]any)
	if !ok {
		t.Fatal("managed .claude.json carries no projects")
	}
	entry, ok := projects[fx.launch].(map[string]any)
	if !ok {
		t.Fatal("the launch directory entry is missing")
	}
	delete(entry, "hasClaudeMdExternalIncludesApproved")
	projects[fx.launch] = entry
	object["projects"] = projects
	rewritten, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	rewritten = append(rewritten, '\n')
	if err := os.WriteFile(filepath.Join(homeDir, ".claude.json"), rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
	dropped := fx.request("claude_code")
	dropped.Machine.Forms = map[string]string{"claude_code": "referenced"}
	if _, err := Resolve(dropped); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("a referenced home whose approval key the tool dropped must be stale, got %v", err)
	}
}

// TestCodexKeyringAmbient drives the keyring-preferred strategy through
// Resolve: a keyring store keeps the credential ambient with no link, any
// other store links auth.json, and an absent config means the default file
// store. The file-store subtest is the narrowing killer for a mutant that
// keeps searching the cli_auth_credentials_store token but treats any
// value as keyring.
func TestCodexKeyringAmbient(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		linked bool
	}{
		{"absent", "", true},
		{"file", "cli_auth_credentials_store = \"file\"\n", true},
		{"auto", "cli_auth_credentials_store = \"auto\"\n", true},
		{"keyring", "cli_auth_credentials_store = \"keyring\"\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			if tc.config != "" {
				if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte(tc.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			// The linked stores get a live native target so the
			// detached-pending finding stays out of this test, which
			// is about ambient-versus-linked, not dangling.
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			req := fx.request("codex_cli")
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			marker := readManagedMarker(t, fx, "codex_cli")
			links := 0
			if marker.Passthrough != nil {
				for _, entry := range *marker.Passthrough {
					if entry.Path != "" {
						links++
					}
				}
			}
			if tc.linked && links != 1 {
				t.Fatalf("a %s store links auth.json, records %d", tc.name, links)
			}
			if !tc.linked && links != 0 {
				t.Fatalf("a keyring store is ambient, records %d", links)
			}
			if tc.name == "keyring" {
				if marker.Passthrough == nil || len(*marker.Passthrough) != 1 || (*marker.Passthrough)[0].Path != "" || (*marker.Passthrough)[0].Backend != "ambient" {
					t.Fatalf("a keyring store records one pathless ambient credential: %+v", marker.Passthrough)
				}
			}
			if _, err := Resolve(fx.request("codex_cli")); err != nil {
				t.Fatalf("the home is current: %v", err)
			}
		})
	}
}

// TestCodexUnreadableConfigFails proves an unreadable native config.toml
// is never treated as the file credential store (§8.4): provisioning
// fails instead of linking, and an existing home goes stale. A directory
// at the config path fails the read deterministically, even as root.
func TestCodexUnreadableConfigFails(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	if err := os.Mkdir(filepath.Join(fx.native["codex_cli"], "config.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err == nil {
		t.Fatal("an unreadable native config.toml must fail provisioning, not default to file")
	}
	if _, err := os.Stat(ManagedHomeDir(fx.home, "acme", "codex_cli")); !os.IsNotExist(err) {
		t.Fatal("a failed provisioning must not leave a managed home behind")
	}
	// An existing home whose native config becomes unreadable goes stale
	// through the same gate.
	fy := writeManagedFixture(t, "acme")
	provision := fy.request("codex_cli")
	provision.Repair = true
	if _, err := Resolve(provision); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(fy.native["codex_cli"], "config.toml")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(fy.native["codex_cli"], "config.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(fy.request("codex_cli")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("an unreadable native config must be stale, got %v", err)
	}
}

// TestRepairNeverRefreshesSeeds proves tool-owned seed state survives
// repair byte for byte: the managed copy keeps its bytes and its record
// even when the native source changed.
func TestRepairNeverRefreshesSeeds(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	nativeSeed := []byte("operator = true\n")
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), nativeSeed, 0o644); err != nil {
		t.Fatal(err)
	}
	// The linked credential gets a live native target so the
	// detached-pending finding stays out of this seed test.
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	homeDir := ManagedHomeDir(fx.home, "acme", "codex_cli")
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("changed = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(homeDir, "AGENTS.md"))
	if _, err := Resolve(req); err != nil {
		t.Fatalf("repair restores: %v", err)
	}
	if payload, _ := os.ReadFile(filepath.Join(homeDir, "config.toml")); string(payload) != string(nativeSeed) {
		t.Fatal("repair refreshed a tool-owned seed")
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Seeds == nil || len(*marker.Seeds) != 1 || (*marker.Seeds)[0] != "config.toml" {
		t.Fatalf("seed record %v", marker.Seeds)
	}
}

// TestResolveIsolatedOpencode narrows the isolation gate through Resolve:
// configuring isolated for opencode fails, never silently shares.
func TestResolveIsolatedOpencode(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("opencode")
	req.Machine.Isolation = map[string]map[string]string{"acme": {"opencode": "isolated"}}
	req.Repair = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagIsolatedUnsupported) {
		t.Fatalf("opencode isolated must fail, got %v", err)
	}
}

// TestResolveFormUnsupported narrows the form gate through Resolve.
func TestResolveFormUnsupported(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request("codex_cli")
	req.Machine.Forms = map[string]string{"codex_cli": "referenced"}
	req.Repair = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagFormUnsupported) {
		t.Fatalf("codex referenced must fail, got %v", err)
	}
}

// TestResolveLockContention proves the --repair lock wait is bounded and
// reports the distinct lock-acquisition diagnostic, not repair_failed.
func TestResolveLockContention(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	held, err := manager.AcquireHomeOnly(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Close() }()
	previous := lockTimeout
	lockTimeout = 0
	defer func() { lockTimeout = previous }()
	req := fx.request("codex_cli")
	req.Repair = true
	_, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagLockUnavailable) {
		t.Fatalf("a contended repair must fail with %s, got %v", DiagLockUnavailable, err)
	}
	if err != nil && strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("lock contention must not masquerade as %s: %v", DiagRepairFailed, err)
	}
}

// TestResolveFormats renders env and shell through Resolve.
func TestResolveFormats(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request("pi")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	req.Repair = false
	req.Format = "env"
	bare, err := Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	want := "PI_CODING_AGENT_DIR=" + ManagedHomeDir(fx.home, "acme", "pi") + "\n"
	if string(bare.Document) != want {
		t.Fatalf("env format %q, want %q", bare.Document, want)
	}
	req.Format = "shell"
	bare, err = Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := "export PI_CODING_AGENT_DIR='" + ManagedHomeDir(fx.home, "acme", "pi") + "'\n"; string(bare.Document) != want {
		t.Fatalf("shell format %q, want %q", bare.Document, want)
	}
	req.Format = "pwsh"
	if _, err := Resolve(req); err == nil {
		t.Fatal("an unknown format must fail")
	}
}
