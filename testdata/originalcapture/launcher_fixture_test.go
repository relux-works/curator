package harnessparity

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

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/testtoolchain"
)

// Inputs reproduce every row of the six original launcher witnesses. Their
// original assertions remain in their original packages and run unchanged.
const launcherExtensions = `{"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"claude","account":"operator"}}}`

var launcherValid = []string{``, `,"extensions":{}`, `,"extensions":{"proxy":{"profile":"private"}}`, `,"extensions":{"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"s","account":"a"}}}`, `,"extensions":` + launcherExtensions}
var launcherInvalid = []string{
	`null`, `[]`, `{"HOSTILE_SECRET_KEY":true}`, `{"proxy":null}`, `{"proxy":{"profile":""}}`, `{"proxy":{"profile":"x","HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}}`,
	`{"proxy":{"profile":"x","profile":"y"}}`,
	`{"claude_auth":{"type":"token","store":{"type":"file","path":"/tmp/x"}}}`,
	`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"relative"}}}`,
	`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"/tmp/../x"}}}`,
	`{"claude_auth":{"type":"oauth","store":{"type":"file","path":"/tmp/x","service":"s"}}}`,
	`{"claude_auth":{"type":"oauth","store":{"type":"keychain","service":"s","account":"a","path":"/tmp/x"}}}`,
	`{"claude_auth":{"type":"oauth","token":"HOSTILE_SECRET_VALUE","store":{"type":"keychain","service":"s","account":"a"}}}`,
	`{"HOSTILE_SECRET_KEY":1,"HOSTILE_SECRET_KEY":2}`,
}

func launcherCases() []fixtureCase {
	var cases []fixtureCase
	add := func(state string) {
		cases = append(cases, fixtureCase{Name: "claude-launcher-" + state, ID: "claude_code", Operation: "launcher", State: state})
	}
	add("normal-edit")
	add("typed-projection")
	for i := range launcherValid {
		add(fmt.Sprintf("valid-%02d", i))
	}
	for i := range launcherInvalid {
		add(fmt.Sprintf("invalid-%02d", i))
	}
	for _, version := range []string{"1", "2", "3", "5", "6"} {
		add("old-version-" + version)
	}
	add("system-authority")
	add("cli-edits")
	add("cli-refusal")
	add("cli-arguments")
	return cases
}
func allCases() []fixtureCase { return append(fixtureCases(), launcherCases()...) }
func runCase(t testing.TB, c fixtureCase) fixtureResult {
	t.Helper()
	if c.Operation == "launcher" {
		return runLauncherFixture(t, c)
	}
	return runLegacyFixture(t, c)
}

func runLauncherFixture(t testing.TB, c fixtureCase) fixtureResult {
	t.Helper()
	root := t.TempDir()
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
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	system := filepath.Join(root, "system.json")
	mustWrite(t, system, []byte(`{"schema_version":2,"locked":[]}`), 0o600)
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	path := filepath.Join(root, "selected.json")
	output := filepath.Join(root, "output")
	if err := os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	result := fixtureResult{Case: c, Executed: true, NativePreserved: true, ProjectPreserved: true}
	artifact := func(name string, payload []byte) {
		mustWrite(t, filepath.Join(output, name), mustProjectFixture(t, rootBinding, payload), 0o600)
	}
	projection := func(cfg *config.Config) {
		t.Helper()
		payload, err := json.Marshal(cfg.LauncherProjection())
		if err != nil {
			t.Fatal(err)
		}
		artifact("projection.json", payload)
	}
	load := func(text string) (*config.Config, error) {
		mustWrite(t, path, []byte(text), 0o600)
		return config.Load(path, nil)
	}
	refuse := func(text string) {
		t.Helper()
		mustWrite(t, path, []byte(text), 0o600)
		beforePaths := fixturePaths(t, root)
		beforeInfo, errInfo := os.Stat(path)
		if errInfo != nil {
			t.Fatal(errInfo)
		}
		cfg, err := config.Load(path, nil)
		if err == nil || cfg != nil || strings.Contains(err.Error(), "HOSTILE_SECRET") {
			t.Fatal("original launcher refusal contract failed")
		}
		actual, errRead := os.ReadFile(path)
		if errRead != nil || string(actual) != text {
			t.Fatal("rejected read changed configuration")
		}
		info, errStat := os.Stat(path)
		if errStat != nil || info.Mode().Perm() != beforeInfo.Mode().Perm() || !reflect.DeepEqual(beforePaths, fixturePaths(t, root)) {
			t.Fatal("rejected read changed mode or created a partial output")
		}
		result.Refusal = string(mustProjectFixture(t, rootBinding, []byte(err.Error())))
		artifact("refusal.txt", []byte(result.Refusal))
		// Rejected source files are inputs, not emitted outputs. Assert their
		// complete unchanged bytes in-process, and export only the digest/mode.
		artifact("retained-input.sha256", []byte(hashBytes(actual)))
		artifact("retained-input-mode.txt", []byte(fmt.Sprintf("%04o\n", info.Mode().Perm())))
	}
	prefix := `{"schema_version":7,"skills_root":"x","projects":{}`
	switch {
	case c.State == "normal-edit":
		project := filepath.Join(root, "project")
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal(err)
		}
		cfg, err := load(`{"schema_version":7,"skills_root":"x","projects":{},"preferred_locale":"en","environments":{"credential_store":{"p":{"codex_cli":"file"}}},"extensions":` + launcherExtensions + `}`)
		if err != nil {
			t.Fatal(err)
		}
		if err := config.AddProject(path, "sample", project, nil); err != nil {
			t.Fatal(err)
		}
		raw, err := config.ReadRaw(path)
		if err != nil {
			t.Fatal(err)
		}
		raw["adapter_mode"] = "copy"
		if _, err := config.WriteRaw(path, raw); err != nil {
			t.Fatal(err)
		}
		after, err := config.Load(path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg.Extensions, after.Extensions) || after.Extensions.Proxy.Profile != "private" || after.Extensions.ClaudeAuth.Store.Account != "operator" || after.PreferredLocale != "en" || after.AdapterMode != "copy" || len(after.Projects) != 1 || !reflect.DeepEqual(cfg.Env.CredentialStore, after.Env.CredentialStore) {
			t.Fatal("normal edits changed asserted launcher state")
		}
		effective := after.EffectiveJSON()
		if _, ok := effective["extensions"]; !ok {
			t.Fatal("effective extensions absent")
		}
		if _, ok := effective["environments"].(map[string]any)["credential_store"]; !ok {
			t.Fatal("credential-store projection absent")
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		artifact("edited-config.json", payload)
		preserveArtifactMode(t, path, filepath.Join(output, "edited-config.json"))
		payload, err = json.Marshal(effective)
		if err != nil {
			t.Fatal(err)
		}
		artifact("effective.json", payload)
		projection(after)
	case c.State == "typed-projection":
		cfg, err := load(prefix + `,"extensions":` + launcherExtensions + `}`)
		if err != nil {
			t.Fatal(err)
		}
		p := cfg.LauncherProjection()
		payload, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(payload, &got); err != nil {
			t.Fatal(err)
		}
		if got["schema"] != "curator.launcher-config/1" || got["config_schema"] != float64(7) {
			t.Fatal("typed projection schema changed")
		}
		artifact("projection.json", payload)
		p.Extensions.Proxy.Profile = "changed"
		if cfg.Extensions.Proxy.Profile != "private" {
			t.Fatal("typed projection aliases configuration")
		}
	case strings.HasPrefix(c.State, "valid-"):
		var n int
		if _, err := fmt.Sscanf(c.State, "valid-%02d", &n); err != nil || n < 0 || n >= len(launcherValid) {
			t.Fatal("invalid case selector")
		}
		cfg, err := load(prefix + launcherValid[n] + `}`)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 && cfg.Extensions != nil {
			t.Fatal("absent selection became present")
		}
		projection(cfg)
	case strings.HasPrefix(c.State, "invalid-"):
		var n int
		if _, err := fmt.Sscanf(c.State, "invalid-%02d", &n); err != nil || n < 0 || n >= len(launcherInvalid) {
			t.Fatal("invalid case selector")
		}
		refuse(prefix + `,"extensions":` + launcherInvalid[n] + `}`)
	case strings.HasPrefix(c.State, "old-version-"):
		version := strings.TrimPrefix(c.State, "old-version-")
		refuse(`{"schema_version":` + version + `,"skills_root":"x","projects":{},"extensions":` + launcherExtensions + `}`)
	case c.State == "system-authority":
		mustWrite(t, system, []byte(`{"schema_version":2,"locked":[],"extensions":`+launcherExtensions+`}`), 0o600)
		refuse(prefix + `,"extensions":` + launcherExtensions + `}`)
		if !strings.Contains(result.Refusal, "extensions") {
			t.Fatal("system authority refusal changed")
		}
	case strings.HasPrefix(c.State, "cli-"):
		runLauncherCLI(t, c, root, path, output, artifact, rootBinding, &result)
	default:
		t.Fatal("unrecognized launcher fixture")
	}
	result.Tree = readTree(t, output, rootBinding)
	return result
}

func moduleRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && info.Mode().IsRegular() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("fixture module root unavailable")
	return ""
}

func runLauncherCLI(t testing.TB, c fixtureCase, root, path, output string, artifact func(string, []byte), rootBinding fixtureRootBinding, result *fixtureResult) {
	t.Helper()
	name := "curator-fixture"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(root, name)
	build := exec.Command(testtoolchain.Go(t), "build", "-o", binary, "./cmd/curator")
	build.Dir = moduleRoot(t)
	if payload, err := build.CombinedOutput(); err != nil {
		t.Fatalf("isolated fixture CLI build failed: %v (%d diagnostic bytes)", err, len(payload))
	}
	store := filepath.Join(root, "not-created-token")
	text := `{"schema_version":7,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"proxy":{"profile":"private"},"claude_auth":{"type":"oauth","store":{"type":"file","path":` + strconvJSON(store) + `}}}}`
	if c.State == "cli-refusal" {
		text = `{"schema_version":7,"skills_root":"x","projects":{},"extensions":{"HOSTILE_SECRET_KEY":"HOSTILE_SECRET_VALUE"}}`
	}
	mustWrite(t, path, []byte(text), 0o600)
	run := func(args ...string) (int, []byte, []byte) {
		t.Helper()
		cmd := exec.Command(binary, args...)
		cmd.Env = append(os.Environ(), "CURATOR_CONFIG="+path)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		code := 0
		if err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				code = exit.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		return code, stdout.Bytes(), stderr.Bytes()
	}
	if c.State == "cli-edits" {
		for _, args := range [][]string{{"env", "config", "set", "overlay_default_weight", "9"}, {"env", "config", "set", "credential_store.sample.codex_cli", "file"}, {"env", "config", "unset", "overlay_default_weight"}} {
			if code, _, _ := run(args...); code != 0 {
				t.Fatal("original CLI edit refused")
			}
		}
		code, stdout, stderr := run("config", "launcher")
		if code != 0 || len(stderr) != 0 {
			t.Fatal("original CLI projection refused")
		}
		var p config.LauncherConfigProjection
		if err := json.Unmarshal(stdout, &p); err != nil {
			t.Fatal(err)
		}
		if p.Schema != "curator.launcher-config/1" || p.ConfigSchema != 7 || p.ConfigPath != path || p.Extensions.Proxy.Profile != "private" || p.Extensions.ClaudeAuth.Store.Path != store {
			t.Fatal("selected CLI projection changed")
		}
		if _, err := os.Lstat(store); !os.IsNotExist(err) {
			t.Fatal("CLI accessed credential fixture")
		}
		payload, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(payload), `"preferred_locale": "en"`) {
			t.Fatal("ordinary edit lost unrelated setting")
		}
		artifact("edited-config.json", payload)
		preserveArtifactMode(t, path, filepath.Join(output, "edited-config.json"))
		artifact("projection.json", stdout)
		artifact("stderr.txt", stderr)
	} else {
		args := []string{"config", "launcher"}
		want := 1
		if c.State == "cli-arguments" {
			path = filepath.Join(root, "absent-config")
			args = append(args, "extra")
			want = 2
		}
		code, stdout, stderr := run(args...)
		if code != want || len(stdout) != 0 || strings.Contains(string(stderr), "HOSTILE_SECRET") || c.State == "cli-arguments" && !strings.Contains(string(stderr), "takes no arguments") {
			t.Fatal("original CLI refusal changed")
		}
		result.Refusal = string(mustProjectFixture(t, rootBinding, stderr))
		artifact("stdout.txt", stdout)
		artifact("refusal.txt", stderr)
	}
}
func strconvJSON(value string) string { payload, _ := json.Marshal(value); return string(payload) }

func preserveArtifactMode(t testing.TB, source, destination string) {
	t.Helper()
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(destination, info.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
}
func fixturePaths(t testing.TB, root string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return paths
}
