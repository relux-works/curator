package harnessparity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
)

// These are test drivers for the unchanged legacy APIs, never staged adapters.
// The same driver bytes are copied to an ignored package in the clean original
// checkout for the one-time baseline capture. Tests have no golden update mode.
type fixtureCase struct{ Name, ID, Operation, State string }
type treeNode struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Mode   uint32 `json:"mode"`
	Bytes  []byte `json:"bytes,omitempty"`
	Link   string `json:"link,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}
type fixtureResult struct {
	Case             fixtureCase `json:"case"`
	Tree             []treeNode  `json:"tree"`
	Rendered         []treeNode  `json:"rendered,omitempty"`
	Document         []byte      `json:"document,omitempty"`
	Notice           string      `json:"notice,omitempty"`
	Warnings         []string    `json:"warnings,omitempty"`
	Refusal          string      `json:"refusal,omitempty"`
	Executed         bool        `json:"executed"`
	NativePreserved  bool        `json:"native_preserved"`
	ProjectPreserved bool        `json:"project_preserved"`
	NativeBefore     []treeNode  `json:"native_before"`
	NativeAfter      []treeNode  `json:"native_after"`
	ProjectBefore    []treeNode  `json:"project_before"`
	ProjectAfter     []treeNode  `json:"project_after"`
}

func fixtureCases() []fixtureCase {
	var cases []fixtureCase
	for _, id := range []string{"claude_code", "codex_cli", "opencode", "muse", "pi"} {
		cases = append(cases, fixtureCase{Name: id + "-empty", ID: id, Operation: "managed", State: "empty"})
	}
	for _, id := range []string{"cursor", "gemini"} {
		cases = append(cases, fixtureCase{Name: id + "-project-mirror", ID: id, Operation: "project-skills", State: "empty"})
	}
	for _, id := range []string{"windsurf", "opencode"} {
		cases = append(cases, fixtureCase{Name: id + "-native-discovery", ID: id, Operation: "native-skills", State: "empty"})
	}
	for _, state := range []string{"user-settings", "current", "modified", "missing-state", "unreadable-state", "malformed-state", "native-parent"} {
		cases = append(cases, fixtureCase{Name: "codex_cli-" + state, ID: "codex_cli", Operation: "managed", State: state})
	}
	cases = append(cases, fixtureCase{Name: "claude_code-mcp", ID: "claude_code", Operation: "managed", State: "user-settings"})
	return cases
}

func mustWrite(t testing.TB, path string, payload []byte, mode fs.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}
func mustJSON(t testing.TB, path string, value any) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, append(payload, '\n'), 0o600)
}
func hashBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func fixtureGitStore(t testing.TB, manager, kind, name, identity, source string) string {
	t.Helper()
	repo := filepath.Join(manager, "profile-repos", strings.ReplaceAll(identity, "/", "_"))
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated fixture Git failed: %v", err)
		}
		return strings.TrimSpace(string(output))
	}
	run("init", "--quiet")
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		mustWrite(t, filepath.Join(repo, rel), payload, 0o600)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	run("add", "--all")
	run("-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "Fixture input")
	commit := run("rev-parse", "HEAD")
	if _, err := contextstore.EnsureGit(manager, kind, name, repo, commit); err != nil {
		t.Fatal(err)
	}
	return commit
}

type fixtureRootForm struct {
	value   string
	kind    string
	windows bool
}

type fixtureRootEvidence struct {
	Root              string
	Origin            string
	RecordedBeforeRun bool
	Ambiguous         bool
}

type fixtureRootBinding struct {
	evidence fixtureRootEvidence
	forms    []fixtureRootForm
	windows  bool
}

type fixtureRootRefusal string

func (refusal fixtureRootRefusal) Error() string { return string(refusal) }

func isFixtureRootBindingRefusal(err error) bool {
	refusal, ok := err.(fixtureRootRefusal)
	if !ok {
		return false
	}
	switch refusal {
	case "fixture-root/missing-binding", "fixture-root/ambiguous-binding", "fixture-root/untrusted-origin", "fixture-root/late-binding", "fixture-root/non-absolute", "fixture-root/trailing-separator":
		return true
	default:
		return false
	}
}

func bindFixtureRoot(evidence fixtureRootEvidence) (fixtureRootBinding, error) {
	if evidence.Root == "" {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/missing-binding")
	}
	if evidence.Ambiguous {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/ambiguous-binding")
	}
	if evidence.Origin != "driver-created" {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/untrusted-origin")
	}
	if !evidence.RecordedBeforeRun {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/late-binding")
	}
	root := evidence.Root
	windows := absoluteWindowsFixtureRoot(root)
	if !windows && !strings.HasPrefix(root, "/") {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/non-absolute")
	}
	if strings.HasSuffix(root, "/") || strings.HasSuffix(root, `\`) {
		return fixtureRootBinding{}, fixtureRootRefusal("fixture-root/trailing-separator")
	}

	binding := fixtureRootBinding{evidence: evidence, windows: windows}
	add := func(value, kind string) {
		if value == "" {
			return
		}
		for _, prior := range binding.forms {
			if prior.value == value {
				return
			}
		}
		binding.forms = append(binding.forms, fixtureRootForm{value: value, kind: kind, windows: windows})
	}
	add(root, "native")
	if windows {
		add(strings.ReplaceAll(root, `\`, "/"), "slash")
	}
	quoted, err := json.Marshal(root)
	if err != nil || len(quoted) < 2 {
		return fixtureRootBinding{}, fmt.Errorf("fixture root JSON spelling unavailable")
	}
	add(string(quoted[1:len(quoted)-1]), "json-escaped")
	sort.SliceStable(binding.forms, func(i, j int) bool {
		return len(binding.forms[i].value) > len(binding.forms[j].value)
	})
	return binding, nil
}

func absoluteWindowsFixtureRoot(root string) bool {
	if len(root) >= 3 && ((root[0] >= 'A' && root[0] <= 'Z') || (root[0] >= 'a' && root[0] <= 'z')) && root[1] == ':' && (root[2] == '\\' || root[2] == '/') {
		return true
	}
	if !strings.HasPrefix(root, `\\`) {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(root, `\\`), `\`)
	return len(parts) >= 2 && parts[0] != "" && parts[1] != ""
}

func (form fixtureRootForm) suffixBoundary(payload []byte, after int) bool {
	if after == len(payload) {
		return true
	}
	next := payload[after]
	if next == '/' {
		return true
	}
	if next == '\\' {
		return form.windows && (form.kind != "json-escaped" || after+1 < len(payload) && payload[after+1] == '\\')
	}
	switch next {
	case '"', '\'', ' ', '\t', '\n', '\r', ',', ';', ')', ']', '}':
		return true
	default:
		return false
	}
}

func fixturePathStartBoundary(payload []byte, at int) bool {
	if at == 0 {
		return true
	}
	switch payload[at-1] {
	case '"', '\'', ' ', '\t', '\n', '\r', '=', ':', '(', '[', '{', ',', ';':
		return true
	default:
		return false
	}
}

func (binding fixtureRootBinding) project(payload []byte) ([]byte, error) {
	if binding.evidence.Root == "" || len(binding.forms) == 0 {
		return nil, fixtureRootRefusal("fixture-root/missing-binding")
	}
	var result bytes.Buffer
	for at := 0; at < len(payload); {
		matched := false
		if fixturePathStartBoundary(payload, at) {
			for _, form := range binding.forms {
				if !bytes.HasPrefix(payload[at:], []byte(form.value)) {
					continue
				}
				after := at + len(form.value)
				if !form.suffixBoundary(payload, after) {
					continue
				}
				result.WriteString("${FIXTURE_ROOT}")
				at = after
				matched = true
				break
			}
		}
		if !matched {
			result.WriteByte(payload[at])
			at++
		}
	}
	return result.Bytes(), nil
}

func projectFixtureRoot(evidence fixtureRootEvidence, payload []byte) ([]byte, error) {
	binding, err := bindFixtureRoot(evidence)
	if err != nil {
		return nil, err
	}
	return binding.project(payload)
}

func mustProjectFixture(t testing.TB, binding fixtureRootBinding, payload []byte) []byte {
	t.Helper()
	projected, err := binding.project(payload)
	if err != nil {
		t.Fatal("fixture root projection refused")
	}
	return projected
}

func readTree(t testing.TB, dir string, binding fixtureRootBinding) []treeNode {
	t.Helper()
	var nodes []treeNode
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		node := treeNode{Path: filepath.ToSlash(rel), Mode: uint32(info.Mode().Perm())}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			node.Kind = "link"
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			node.Link = string(mustProjectFixture(t, binding, []byte(link)))
			node.SHA256 = hashBytes([]byte(node.Link))
		case info.IsDir():
			node.Kind = "directory"
		case info.Mode().IsRegular():
			node.Kind = "file"
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			node.Bytes = mustProjectFixture(t, binding, payload)
			node.SHA256 = hashBytes(node.Bytes)
		default:
			return fmt.Errorf("unexpected fixture node %s", rel)
		}
		nodes = append(nodes, node)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
	return nodes
}

func runLegacyFixture(t testing.TB, c fixtureCase) fixtureResult {
	result, _ := runLegacyFixtureWithRoot(t, c)
	return result
}

func runLegacyFixtureWithRoot(t testing.TB, c fixtureCase) (fixtureResult, string) {
	return runLegacyFixtureAtRoot(t, c, "", false)
}

func runLegacyFixtureAtRoot(t testing.TB, c fixtureCase, fixtureRoot string, isolateNativeParent bool) (fixtureResult, string) {
	t.Helper()
	root := fixtureRoot
	if root == "" {
		root = t.TempDir()
	} else if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal("fixture root creation failed")
	}
	physical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = physical
	rootBinding, err := bindFixtureRoot(fixtureRootEvidence{
		Root: root, Origin: "driver-created", RecordedBeforeRun: true,
	})
	if err != nil {
		t.Fatal("fixture root binding refused")
	}
	manager, native, project := filepath.Join(root, "manager"), filepath.Join(root, "native"), filepath.Join(root, "project")
	for _, dir := range []string{manager, native, project, filepath.Join(root, "operator"), filepath.Join(root, "xdg")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", filepath.Join(root, "operator"))
	t.Setenv("USERPROFILE", filepath.Join(root, "operator"))
	for _, key := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME", "SSH_AUTH_SOCK", "DBUS_SESSION_BUS_ADDRESS"} {
		t.Setenv(key, "")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_DATE", "2000-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2000-01-01T00:00:00Z")
	result := fixtureResult{Case: c}
	if c.Operation != "managed" {
		return runLegacySkills(t, c, root, manager, native, project, rootBinding), root
	}
	profile := "sample"
	contextSource := filepath.Join(root, "context-source")
	mustWrite(t, filepath.Join(contextSource, "agent-context.json"), []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"a.md"},{"path":"s.md","class":"system"}]}}`), 0o600)
	mustWrite(t, filepath.Join(contextSource, "context", "a.md"), []byte("Fixture root context.\n"), 0o600)
	mustWrite(t, filepath.Join(contextSource, "context", "s.md"), []byte("Fixture system context.\n"), 0o600)
	if err := pathboundary.ProtectTree(contextSource); err != nil {
		t.Fatal(err)
	}
	_, pin, err := contextstore.EnsureState(manager, "context", profile, contextSource)
	if err != nil {
		t.Fatal(err)
	}
	members := []contextlock.Member{{Kind: "context", Name: profile, Version: "1.0.0", StateHash: pin, Weight: 100}}
	// Nonsecret MCP declaration is data; its inert executable is never started.
	mcpSource := filepath.Join(root, "mcp-source")
	mustWrite(t, filepath.Join(mcpSource, "agent-mcp.json"), []byte(`{"schema_version":1,"name":"fixture-server","version":"1.0.0","server":{"transport":"stdio","command":"fixture-never-launched","args":["--inert"],"env_names":[],"environments":["claude_code","codex_cli","opencode"]}}`), 0o600)
	mcpIdentity := "github.com/example/fixture-mcp"
	mcpPin := fixtureGitStore(t, manager, "mcp", "fixture-server", mcpIdentity, mcpSource)
	members = append(members, contextlock.Member{Kind: "mcp", Name: "fixture-server", Version: "1.0.0", Source: mcpIdentity, Commit: mcpPin, RequiredBy: []string{profile}})
	skillSource := filepath.Join(root, "skill-source")
	mustWrite(t, filepath.Join(skillSource, "SKILL.md"), []byte("# Fixture skill\n"), 0o600)
	mustWrite(t, filepath.Join(skillSource, "agent-skill.json"), []byte(`{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}`), 0o600)
	skillIdentity := "github.com/example/fixture-skill"
	skillPin := fixtureGitStore(t, manager, "skill", "fixture-skill", skillIdentity, skillSource)
	members = append(members, contextlock.Member{Kind: "skill", Name: "fixture-skill", Source: skillIdentity, Commit: skillPin, RequiredBy: []string{profile}})
	lock := &contextlock.Lock{Root: profile, Members: members}
	lock.Sort()
	if _, err := contextlock.Write(filepath.Join(envprofile.ProfileDir(manager, profile), "lock.json"), lock); err != nil {
		t.Fatal(err)
	}
	mustJSON(t, filepath.Join(envprofile.ProfileDir(manager, profile), "source.json"), envprofile.Source{Kind: envprofile.KindPath, Path: contextSource})
	if err := envprofile.SetCurrent(manager, profile); err != nil {
		t.Fatal(err)
	}
	nativeHomes := map[string]string{}
	for _, id := range []string{"claude_code", "codex_cli", "opencode", "muse", "pi"} {
		nativeHomes[id] = filepath.Join(native, id)
		if err := os.MkdirAll(nativeHomes[id], 0o700); err != nil {
			t.Fatal(err)
		}
	}
	nativeParentHome := filepath.Join(native, c.ID+"-native-parent")
	if c.State == "native-parent" {
		if err := os.MkdirAll(nativeParentHome, 0o700); err != nil {
			t.Fatal("fixture-owned native home setup failed")
		}
	}
	// Synthetic credential targets are fixtures only; the comparator never
	// follows these links or reads their contents through a managed home.
	mustWrite(t, filepath.Join(nativeHomes["claude_code"], ".credentials.json"), []byte("fixture-only\n"), 0o600)
	mustWrite(t, filepath.Join(nativeHomes["codex_cli"], "auth.json"), []byte("fixture-only\n"), 0o600)
	mustWrite(t, filepath.Join(nativeHomes["muse"], "auth.json"), []byte("fixture-only\n"), 0o600)
	mustWrite(t, filepath.Join(nativeHomes["pi"], "agent", "auth.json"), []byte("fixture-only\n"), 0o600)
	if c.State == "user-settings" {
		if c.ID == "codex_cli" {
			mustWrite(t, filepath.Join(nativeHomes[c.ID], "config.toml"), []byte("model = \"fixture-model\"\n[projects.\"/synthetic/project\"]\ntrust_level = \"trusted\"\n[mcp_servers.fixture]\ncommand = \"never-launched\"\n"), 0o640)
		}
		if c.ID == "claude_code" {
			mustWrite(t, filepath.Join(nativeHomes[c.ID], "settings.json"), []byte("{\"fixtureUnknown\":true}\n"), 0o640)
		}
	}
	beforeNative := readTree(t, native, rootBinding)
	beforeProject := readTree(t, project, rootBinding)
	req := envprofile.ResolveRequest{Home: manager, Profile: profile, EnvID: c.ID, LaunchDir: project, Machine: envregistry.DefaultMachineConfig(), Repair: true, Detect: func(envregistry.Adapter) string { return "unknown" }, NativeHomeOf: func(id string) (string, error) { return nativeHomes[id], nil }, OperatorXDG: filepath.Join(root, "xdg")}
	managed := envprofile.ManagedHomeDir(manager, profile, c.ID)
	if c.State == "native-parent" {
		// This home belongs to the fixture and is external to the manager's
		// managed environments; never derive it from the caller's CODEX_HOME.
		req.NativeHomeOf = func(string) (string, error) { return nativeParentHome, nil }
	}
	if isolateNativeParent {
		if c.State != "native-parent" {
			t.Fatal("native-parent isolation used for another fixture")
		}
		trustedPath := os.Getenv("PATH")
		os.Clearenv()
		if err := os.Setenv("HOME", filepath.Join(root, "operator")); err != nil {
			t.Fatal("isolated fixture HOME unavailable")
		}
		if err := os.Setenv("PATH", trustedPath); err != nil {
			t.Fatal("isolated fixture PATH unavailable")
		}
		if !onlyHomeAndPathEnvironment() {
			t.Fatal("capture-driver/native-parent-environment-not-isolated")
		}
	}
	resolved, resolveErr := envprofile.Resolve(req)
	if resolveErr == nil {
		switch c.State {
		case "current":
			req.Repair = false
			resolved, resolveErr = envprofile.Resolve(req)
		case "modified":
			mustWrite(t, filepath.Join(managed, "AGENTS.md"), []byte("user-modified\n"), 0o600)
			req.Repair = false
			resolved, resolveErr = envprofile.Resolve(req)
		case "missing-state":
			if err := os.Remove(filepath.Join(managed, envmarker.Name)); err != nil {
				t.Fatal(err)
			}
			req.Repair = false
			resolved, resolveErr = envprofile.Resolve(req)
		case "unreadable-state":
			// A directory at a regular-state boundary is an objective unreadable
			// file even under privileged CI, without depending on chmod bypass.
			if err := os.Remove(filepath.Join(managed, envmarker.Name)); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(managed, envmarker.Name), 0o700); err != nil {
				t.Fatal(err)
			}
			req.Repair = false
			resolved, resolveErr = envprofile.Resolve(req)
		case "malformed-state":
			mustWrite(t, filepath.Join(managed, envmarker.Name), []byte("{invalid\n"), 0o600)
			req.Repair = false
			resolved, resolveErr = envprofile.Resolve(req)
		}
	}
	result.Executed = true
	if resolveErr != nil {
		result.Refusal = string(mustProjectFixture(t, rootBinding, []byte(resolveErr.Error())))
	}
	if resolved != nil {
		result.Document = mustProjectFixture(t, rootBinding, resolved.Document)
		result.Notice = string(mustProjectFixture(t, rootBinding, []byte(resolved.Notice)))
		for _, w := range resolved.Warnings {
			result.Warnings = append(result.Warnings, string(mustProjectFixture(t, rootBinding, []byte(w))))
		}
	}
	result.Tree = readTree(t, managed, rootBinding)
	result.Rendered = readTree(t, filepath.Join(envprofile.ProfileDir(manager, profile), "rendered", c.ID), rootBinding)
	// Legacy resolution also reconciles linked native views of the selected
	// profile. Preserve those exact effects in the original baseline.
	result.NativeBefore = beforeNative
	result.NativeAfter = readTree(t, native, rootBinding)
	result.ProjectBefore = beforeProject
	result.ProjectAfter = readTree(t, project, rootBinding)
	result.NativePreserved = reflect.DeepEqual(result.NativeBefore, result.NativeAfter)
	result.ProjectPreserved = reflect.DeepEqual(result.ProjectBefore, result.ProjectAfter)
	return result, root
}

func runLegacySkills(t testing.TB, c fixtureCase, root, manager, native, project string, rootBinding fixtureRootBinding) fixtureResult {
	t.Helper()
	source := filepath.Join(root, "skill-source")
	mustWrite(t, filepath.Join(source, "fixture-skill", "SKILL.md"), []byte("# Fixture skill\n"), 0o600)
	stage := filepath.Join(root, "stage")
	mirror, err := adapters.StageProject(stage, project, []string{c.ID}, []adapters.Group{{Root: source, Skills: []string{"fixture-skill"}}}, "copy")
	result := fixtureResult{Case: c, Executed: true, NativePreserved: true, ProjectPreserved: true}
	if err != nil {
		result.Refusal = string(mustProjectFixture(t, rootBinding, []byte(err.Error())))
		return result
	}
	plan := mirror.Plan()
	// Apply exactly the legacy staged file/tree entries into a disposable
	// project. No transaction policy is being substituted or qualified here.
	for _, target := range plan.Sorted() {
		if target.Removal() {
			t.Fatal("unexpected removal from empty projection")
		}
		if err := os.MkdirAll(filepath.Dir(target.LivePath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(target.StagedPath, target.LivePath); err != nil {
			t.Fatal(err)
		}
	}
	result.Tree = readTree(t, project, rootBinding)
	if c.Operation == "native-skills" {
		if !adapters.NativeDiscoveryAgents[c.ID] || !plan.Empty() {
			t.Fatal("native-discovery no-mirror contract changed")
		}
		result.Notice = "native discovery: no project mirror"
	}
	if c.Operation == "project-skills" && len(result.Tree) == 0 {
		t.Fatal("project mirror did not execute")
	}
	return result
}

func compareResult(expected, actual fixtureResult) error {
	if !actual.Executed {
		return fmt.Errorf("%s: adapter case not executed", actual.Case.Name)
	}
	if !reflect.DeepEqual(expected.Case, actual.Case) {
		return fmt.Errorf("fixture case mismatch")
	}
	if err := compareTrees("managed", expected.Tree, actual.Tree); err != nil {
		return err
	}
	if err := compareTrees("rendered", expected.Rendered, actual.Rendered); err != nil {
		return err
	}
	if !bytes.Equal(expected.Document, actual.Document) || expected.Notice != actual.Notice || !reflect.DeepEqual(expected.Warnings, actual.Warnings) || expected.Refusal != actual.Refusal || expected.NativePreserved != actual.NativePreserved || expected.ProjectPreserved != actual.ProjectPreserved {
		return fmt.Errorf("%s: document/notice/refusal/preservation mismatch", actual.Case.Name)
	}
	for label, pair := range map[string][2][]treeNode{"native-before": {expected.NativeBefore, actual.NativeBefore}, "native-after": {expected.NativeAfter, actual.NativeAfter}, "project-before": {expected.ProjectBefore, actual.ProjectBefore}, "project-after": {expected.ProjectAfter, actual.ProjectAfter}} {
		if err := compareTrees(label, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func compareTrees(label string, expected, actual []treeNode) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%s: path count expected=%d actual=%d", label, len(expected), len(actual))
	}
	for i, want := range expected {
		got := actual[i]
		if want.Path != got.Path {
			return fmt.Errorf("path set mismatch expected=%s actual=%s", want.Path, got.Path)
		}
		if want.Kind != got.Kind || want.Mode != got.Mode || want.Link != got.Link {
			return fmt.Errorf("%s: kind/mode/link mismatch", want.Path)
		}
		if want.SHA256 != got.SHA256 || !bytes.Equal(want.Bytes, got.Bytes) {
			return fmt.Errorf("%s: expected digest=%s actual digest=%s", want.Path, want.SHA256, got.SHA256)
		}
	}
	return nil
}

func onlyRoot(nodes []treeNode) bool {
	return len(nodes) == 1 && nodes[0].Path == "." && nodes[0].Kind == "directory"
}

func containsNode(result fixtureResult, suffix string) bool {
	for _, node := range result.Tree {
		if strings.HasSuffix(node.Path, suffix) && node.Kind == "file" && len(node.Bytes) > 0 {
			return true
		}
	}
	return false
}
