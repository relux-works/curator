package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Compile an inert, portable binary: no real Muse or credential store is used.
func fakeMuse(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "fake.go")
	program := `package main
import("encoding/json";"os")
func main(){if len(os.Args)==2 && os.Args[1]=="--version" {os.Stdout.WriteString("1.4.1-R4503.1\n");return}; env:=map[string]string{};for _,key:=range []string{"HOME","XDG_CONFIG_HOME","XDG_DATA_HOME","XDG_STATE_HOME","XDG_CACHE_HOME"}{env[key]=os.Getenv(key)};json.NewEncoder(os.Stdout).Encode(map[string]any{"env":env,"argv":os.Args[1:]})}`
	if err := os.WriteFile(src, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	name := "muse"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", binary, src)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake Muse: %v %s", err, output)
	}
	return binary
}

// prepareNativeMuse uses only an isolated credential fixture and inert binary.
func prepareNativeMuse(t *testing.T) {
	t.Helper()
	binary := fakeMuse(t)
	t.Setenv("PATH", filepath.Dir(binary)+string(os.PathListSeparator)+os.Getenv("PATH"))
	native := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "muse", "auth.json")
	if err := os.MkdirAll(filepath.Dir(native), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(native, []byte("synthetic-native\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

// Keep status-control fixtures current across the registered adapter set so
// each --check result measures the posture under test.
func provisionRegisteredEnvironments(t *testing.T, source stubConfigSource, profiles ...string) {
	t.Helper()
	for _, profile := range profiles {
		for _, adapter := range envregistry.Registry {
			if code, _, stderr := runProfile(t, source, "env", "resolve", adapter.ID, "--profile", profile, "--repair"); code != exitOK {
				t.Fatalf("repair %s %s stderr:\n%s", profile, adapter.ID, stderr)
			}
		}
	}
}

func TestMuseCLIFragmentAndFakeLaunch(t *testing.T) {
	source, _ := profileHome(t)
	operatorHome := t.TempDir()
	binary := fakeMuse(t)
	t.Setenv("HOME", operatorHome)
	t.Setenv("PATH", filepath.Dir(binary)+string(os.PathListSeparator)+os.Getenv("PATH"))
	native := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "muse", "auth.json")
	if err := os.MkdirAll(filepath.Dir(native), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(native, []byte("synthetic-native\n"), 0600); err != nil {
		t.Fatal(err)
	}
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	for _, seed := range []string{"settings.json", "trust.json"} {
		path := filepath.Join(pkg, "config", "muse", seed)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{\"fixture\":true}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install: %d %s", code, stderr)
	}
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair", "--format", "json")
	if code != exitOK {
		t.Fatalf("resolve: %d %s", code, stderr)
	}
	validateMuseFragmentV3(t, museFragmentV3Schema(t), []byte(stdout))
	resolved, err := envprofile.Resolve(envprofile.ResolveRequest{
		Home: source.cfg.Home(), Profile: "acme", EnvID: envregistry.Muse,
		Machine: envregistry.DefaultMachineConfig(), Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	validateMuseFragmentV3(t, museFragmentV3Schema(t), resolved.Document)
	var fragment struct {
		Fragment string            `json:"fragment"`
		Env      map[string]string `json:"env"`
	}
	if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
		t.Fatal(err)
	}
	if fragment.Fragment != "launch-env-fragment-v3" || len(fragment.Env) != 4 {
		t.Fatalf("fragment %s", stdout)
	}
	home := envprofile.ManagedHomeDir(source.cfg.Home(), "acme", "muse")
	for key, dir := range map[string]string{"XDG_CONFIG_HOME": "config", "XDG_DATA_HOME": "data", "XDG_STATE_HOME": "state", "XDG_CACHE_HOME": "cache"} {
		if fragment.Env[key] != filepath.Join(home, dir) {
			t.Fatalf("%s=%s", key, fragment.Env[key])
		}
	}
	for _, seed := range []string{"settings.json", "trust.json"} {
		payload, err := os.ReadFile(filepath.Join(home, "config", "muse", seed))
		if err != nil || string(payload) != "{\"fixture\":true}\n" {
			t.Fatalf("seed %s: %s %v", seed, payload, err)
		}
	}
	command := exec.Command(binary, "exec", "--", "synthetic prompt")
	command.Env = os.Environ()
	// Apply exactly the production resolver's fragment, preserving operator HOME.
	for key, value := range fragment.Env {
		var filtered []string
		for _, entry := range command.Env {
			if !strings.HasPrefix(entry, key+"=") {
				filtered = append(filtered, entry)
			}
		}
		command.Env = append(filtered, key+"="+value)
	}
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Env  map[string]string `json:"env"`
		Argv []string          `json:"argv"`
	}
	if err := json.Unmarshal(output, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Env["HOME"] != operatorHome {
		t.Fatalf("HOME replaced: %s", output)
	}
	for key, value := range fragment.Env {
		if observed.Env[key] != value {
			t.Fatalf("child %s=%s", key, observed.Env[key])
		}
	}
	if !reflect.DeepEqual(observed.Argv, []string{"exec", "--", "synthetic prompt"}) {
		t.Fatalf("argv %v", observed.Argv)
	}
	// Tool-owned seeds are not refreshed by repair.
	seed := filepath.Join(home, "config", "muse", "settings.json")
	if err := os.WriteFile(seed, []byte("tool-owned\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
		t.Fatal(stderr)
	}
	if payload, err := os.ReadFile(seed); err != nil || string(payload) != "tool-owned\n" {
		t.Fatalf("seed refreshed: %s %v", payload, err)
	}
	// A fork is refused through the real CLI for both resolve paths.
	link := filepath.Join(home, "config", "muse", "auth.json")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("synthetic-fork\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"env", "resolve", "muse"}, {"env", "resolve", "muse", "--repair"}} {
		code, stdout, stderr := runProfile(t, source, args...)
		if code != exitFail || stdout != "" || !strings.Contains(stderr, "environment_") {
			t.Fatalf("fork accepted: %d %s %s", code, stdout, stderr)
		}
	}
}

func TestMuseProfileSeedSecretAndCredentialRefusals(t *testing.T) {
	for _, tc := range []struct{ name, file, body string }{{"secret", "settings.json", "Bearer abcdefghijklmnopqrstuvwxyz123456"}, {"auth", "auth.json", "synthetic-credential"}} {
		t.Run(tc.name, func(t *testing.T) {
			source, _ := profileHome(t)
			pkg := t.TempDir()
			writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
			path := filepath.Join(pkg, "config", "muse", tc.file)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			code, _, stderr := runProfile(t, source, "profile", "install", pkg)
			if code == exitOK {
				t.Fatalf("unsafe profile accepted: %s", stderr)
			}
		})
	}
}

func TestMuseFragmentV3PublishedCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	digest, err := conformancecoverage.SelectedSuiteManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	if !conformancecoverage.IsMuseCandidate(digest) {
		t.Log("selected suite has no fragment-v3 family; rc.13 unchanged")
		return
	}
	schema := museFragmentV3Schema(t)
	source := museCLIProfile(t)
	raw, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var index []struct {
		Instance string `json:"instance"`
		Valid    bool   `json:"valid"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Instance string `json:"instance"`
		Valid    bool   `json:"valid"`
	}
	for _, entry := range index {
		if strings.HasPrefix(entry.Instance, "launch-env-fragment-v3/") {
			cases = append(cases, entry)
		}
	}
	conformancecoverage.Run(t, "launch-env-fragment-v3/schema-cases", cases, func(tc struct {
		Instance string `json:"instance"`
		Valid    bool   `json:"valid"`
	}) string {
		return strings.TrimPrefix(tc.Instance, "launch-env-fragment-v3/")
	}, func(t *testing.T, tc struct {
		Instance string `json:"instance"`
		Valid    bool   `json:"valid"`
	}) {
		raw, err := os.ReadFile(filepath.Join(root, "schema-cases", tc.Instance))
		if err != nil {
			t.Fatal(err)
		}
		var object any
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		fixture := object.(map[string]any)
		if tc.Valid && fixture["environment"] == envregistry.Muse {
			// These four rows describe producer-reachable permission policies.
			// Both entry points validate their actual, unmodified emitted JSON.
			driveMuseFragmentPolicy(t, schema, source, fixture)
			t.Log("production Resolve + CLI emission; schema/reader oracle also checked")
		} else {
			// Invalid wire documents and other adapters' v3 documents are
			// reader-only corpus rows. Resolve emits v2 for the other adapters;
			// malformed wire documents cannot be supplied as resolver inputs.
			t.Log("static schema/reader oracle; not a production emission row")
		}
		err = schema.Validate(object)
		if err == nil {
			var value struct {
				Environment string            `json:"environment"`
				Env         map[string]string `json:"env"`
			}
			if decodeErr := json.Unmarshal(raw, &value); decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if value.Environment == envregistry.Muse {
				adapter, lookupErr := envregistry.ByID(value.Environment)
				if lookupErr != nil {
					t.Fatal(lookupErr)
				}
				nativeRoot := filepath.Join(t.TempDir(), "environments")
				for name, path := range value.Env {
					value.Env[name] = filepath.FromSlash(strings.Replace(path, "/manager/environments", nativeRoot, 1))
				}
				err = envfragment.CheckBoundary(adapter, nativeRoot, &envfragment.Fragment{Environment: value.Environment, Env: value.Env})
			}
		}
		if (err == nil) != tc.Valid {
			t.Fatalf("valid=%v: %v", tc.Valid, err)
		}
	})
}

// The v3 schema is copied verbatim from d373078a, independently of the selected
// conformance root. This keeps the emission regression live in rc.13 CI too.
func museFragmentV3Schema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	mainDir, _ := launchEnvFragmentV2FixtureRoots(t)
	compiler := jsonschema.NewCompiler()
	compiler.UseRegexpEngine(compileECMARegexp)
	for _, name := range []string{"common", "agent-environment-marker-v1", "agent-context-v1", "agent-mcp-v1", "context-lock-v1", "launch-env-fragment-v3"} {
		path := filepath.Join(mainDir, name+".schema.json")
		if name == "launch-env-fragment-v3" {
			path = filepath.Join(mainDir, "..", "..", "..", "curator-spec-muse", "schemas", "v1", name+".schema.json")
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if name == "launch-env-fragment-v3" {
			hash := sha256.Sum256(raw)
			if hex.EncodeToString(hash[:]) != "1e05c7f86d341873d4167348297f2339567c45a79e40572544d850c34e370449" {
				t.Fatal("pinned Muse v3 schema changed")
			}
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		if err := compiler.AddResource(document["$id"].(string), document); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := compiler.Compile("https://relux-works.github.io/curator-spec/schemas/v1/launch-env-fragment-v3.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func validateMuseFragmentV3(t *testing.T, schema *jsonschema.Schema, payload []byte) map[string]any {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		// The normative schema only admits POSIX paths. Validate the emitted
		// structure using slash-rooted projections of host paths; raw Windows
		// path schema conformance remains an explicit spec bound. No members
		// are synthesized, so missing permissions or XDG parents still fail.
		t.Log("v3 Windows path schema bound: structure checked with POSIX path projection")
		if env, ok := object["env"].(map[string]any); ok {
			for name, raw := range env {
				if path, ok := raw.(string); ok {
					env[name] = filepath.ToSlash(strings.TrimPrefix(path, filepath.VolumeName(path)))
				}
			}
		}
	}
	if err := schema.Validate(object); err != nil {
		t.Fatalf("production Muse fragment violates pinned v3 schema: %v", err)
	}
	return object
}

func museCLIProfile(t *testing.T) stubConfigSource {
	t.Helper()
	source, _ := profileHome(t)
	prepareNativeMuse(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install: %d %s", code, stderr)
	}
	return source
}

func driveMuseFragmentPolicy(t *testing.T, schema *jsonschema.Schema, source stubConfigSource, fixture map[string]any) {
	t.Helper()
	permission := fixture["permissions"].(map[string]any)
	machine := envregistry.DefaultMachineConfig()
	configObject := map[string]any{"schema_version": 2, "security_posture": "permissive", "skills_root": filepath.Join(source.cfg.Home(), "skills"), "projects": map[string]any{}}
	if permission["source"] == "profile" {
		machine.Permissions["acme"] = permission["mode"].(string)
		configObject["environments"] = map[string]any{"permissions": machine.Permissions}
	}
	machine.PermissionsLocked = permission["locked"].(bool)
	system := map[string]any{"schema_version": 2, "locked": []string{}}
	if machine.PermissionsLocked {
		system["locked"] = []string{"environments.permissions"}
		system["environments"] = map[string]any{"permissions": map[string]any{}}
	}
	systemPath := filepath.Join(t.TempDir(), "system.json")
	for path, object := range map[string]any{source.path: configObject, systemPath: system} {
		raw, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
	for _, repair := range []bool{true, false} {
		req := envprofile.ResolveRequest{Home: source.cfg.Home(), Profile: "acme", EnvID: envregistry.Muse, Machine: machine, Format: "json", Repair: repair}
		result, err := envprofile.Resolve(req)
		if err != nil {
			t.Fatal(err)
		}
		args := []string{"env", "resolve", "muse", "--format", "json"}
		if repair {
			args = append(args, "--repair")
		}
		code, stdout, stderr := runFileConfigProfile(t, source.path, args...)
		if code != exitOK {
			t.Fatalf("CLI resolve: %d %s", code, stderr)
		}
		for _, document := range [][]byte{result.Document, []byte(stdout)} {
			object := validateMuseFragmentV3(t, schema, document)
			if !reflect.DeepEqual(object["permissions"], permission) {
				t.Fatalf("production permission %v, corpus wants %v", object["permissions"], permission)
			}
		}
	}
}

func TestMuseCLIAuthParentBoundary(t *testing.T) {
	source := museCLIProfile(t)
	if code, _, stderr := runProfile(t, source, "env", "resolve", "muse", "--repair"); code != exitOK {
		t.Fatalf("provision: %d %s", code, stderr)
	}
	home := envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.Muse)
	config := filepath.Join(home, "config")
	outside := filepath.Join(t.TempDir(), "preserved-config")
	if err := os.Rename(config, outside); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, config); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(outside, "muse", "auth.json")
	native := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "muse", "auth.json")
	if target, err := os.Readlink(link); err != nil || target != native {
		t.Fatalf("fixture auth link: %s %v", target, err)
	}
	preserved := filepath.Join(outside, "preserve.txt")
	if err := os.WriteFile(preserved, []byte("outside-owned\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, repair := range []bool{false, true} {
		name := "resolve"
		args := []string{"env", "resolve", "muse", "--format", "json"}
		if repair {
			name = "resolve-repair"
			args = append(args, "--repair")
		}
		t.Run(name, func(t *testing.T) {
			code, stdout, stderr := runProfile(t, source, args...)
			if code != exitFail || stdout != "" || !strings.Contains(stderr, envprofile.DiagWriteWouldFollowLink) {
				t.Fatalf("parent link refusal: exit=%d stdout=%q stderr=%s", code, stdout, stderr)
			}
			t.Logf("production CLI refusal exit=%d; no fragment emitted", code)
			for path, want := range map[string]string{native: "synthetic-native\n", preserved: "outside-owned\n"} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("preserved bytes %s: %v", path, err)
				}
			}
			if target, err := os.Readlink(link); err != nil || target != native {
				t.Fatalf("outside auth link changed: %s %v", target, err)
			}
		})
	}
}
