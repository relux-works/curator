package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const launchEnvFragmentV2SchemaSHA256 = "4d26b5e2a89452eb3c7fe6945f19dd16dda557382be15972ab9c57cb02d38d40"

const launchEnvFragmentV2WindowsSchemaGap = `launch-env-fragment-v2 schema gap (environments §10.2): Windows host-native paths do not match pattern ^/[^\u0000]*$; permissions assertions passed`

type ecmaRegexp regexp2.Regexp

func (re *ecmaRegexp) MatchString(value string) bool {
	matched, err := (*regexp2.Regexp)(re).MatchString(value)
	return err == nil && matched
}

func (re *ecmaRegexp) String() string {
	return (*regexp2.Regexp)(re).String()
}

func compileECMARegexp(pattern string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	return (*ecmaRegexp)(re), nil
}

func launchEnvFragmentV2FixtureRoots(t *testing.T) (string, string) {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not locate this test")
	}
	repo := filepath.Join(filepath.Dir(source), "..", "..")
	fixture := filepath.Join(repo, "internal", "envfragment", "testdata", "curator-spec-main")
	return filepath.Join(fixture, "schemas", "v1"), filepath.Join(fixture, "conformance", "v1", "schema-cases", "launch-env-fragment-v2")
}

func launchEnvFragmentV2Schema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	schemaDir, _ := launchEnvFragmentV2FixtureRoots(t)
	names := []string{
		"agent-context-v1.schema.json",
		"agent-environment-marker-v1.schema.json",
		"agent-mcp-v1.schema.json",
		"common.schema.json",
		"context-lock-v1.schema.json",
		"launch-env-fragment-v2.schema.json",
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseRegexpEngine(compileECMARegexp)
	mainID := ""
	for _, name := range names {
		path := filepath.Join(schemaDir, name)
		raw, err := os.ReadFile(path) // #nosec G304 -- the input is a task-owned, explicit schema fixture
		if err != nil {
			t.Fatalf("read schema fixture %s: %v", path, err)
		}
		if name == "launch-env-fragment-v2.schema.json" {
			hash := sha256.Sum256(raw)
			if got := hex.EncodeToString(hash[:]); got != launchEnvFragmentV2SchemaSHA256 {
				t.Fatalf("v2 schema fixture hash = %s, want copied spec-main hash %s", got, launchEnvFragmentV2SchemaSHA256)
			}
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatalf("parse schema fixture %s: %v", path, err)
		}
		id, ok := document["$id"].(string)
		if !ok || id == "" {
			t.Fatalf("schema fixture %s has no $id", path)
		}
		if err := compiler.AddResource(id, document); err != nil {
			t.Fatalf("register schema fixture %s: %v", path, err)
		}
		if name == "launch-env-fragment-v2.schema.json" {
			mainID = id
		}
	}
	if mainID == "" {
		t.Fatal("launch-env-fragment-v2 schema is missing from the fixture set")
	}
	compiled, err := compiler.Compile(mainID)
	if err != nil {
		t.Fatalf("compile copied spec-main v2 schema: %v", err)
	}
	return compiled
}

func validateLaunchEnvFragmentV2(t *testing.T, schema *jsonschema.Schema, payload []byte) {
	t.Helper()
	var instance any
	if err := json.Unmarshal(payload, &instance); err != nil {
		t.Fatalf("fragment output is not JSON: %v\n%s", err, payload)
	}
	if err := schema.Validate(instance); err != nil {
		t.Fatalf("fragment output does not validate against copied curator-spec main schema: %v\n%s", err, payload)
	}
}

func TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases(t *testing.T) {
	schema := launchEnvFragmentV2Schema(t)
	_, casesDir := launchEnvFragmentV2FixtureRoots(t)
	expected := map[string]bool{
		"valid.json":                                     true,
		"valid-permissions-yolo-unlocked.json":           true,
		"valid-permissions-native-locked.json":           true,
		"valid-permissions-native-silent.json":           true,
		"invalid.json":                                   false,
		"invalid-permissions-locked-profile-source.json": false,
		"invalid-permissions-missing-mode.json":          false,
		"invalid-permissions-unknown-field.json":         false,
		"invalid-permissions-yolo-silent.json":           false,
		"invalid-permissions-global-unlocked.json":       false,
		"invalid-permissions-unknown-source.json":        false,
		"invalid-permissions-absent.json":                false,
		"invalid-permissions-yolo-locked.json":           false,
		"invalid-permissions-unknown-mode.json":          false,
	}
	paths, err := filepath.Glob(filepath.Join(casesDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != len(expected) {
		t.Fatalf("copied spec-main schema case count = %d, want %d", len(paths), len(expected))
	}
	sort.Strings(paths)
	for _, path := range paths {
		name := filepath.Base(path)
		wantValid, known := expected[name]
		if !known {
			t.Fatalf("unexpected copied spec-main schema case %q", name)
		}
		raw, err := os.ReadFile(path) // #nosec G304 -- filenames are constrained by the explicit fixture glob
		if err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) {
			var instance any
			if err := json.Unmarshal(raw, &instance); err != nil {
				t.Fatal(err)
			}
			err := schema.Validate(instance)
			if wantValid && err != nil {
				t.Fatalf("spec-main valid case rejected: %v", err)
			}
			if !wantValid && err == nil {
				t.Fatal("spec-main invalid case accepted")
			}
		})
	}
}

func runFileConfigProfile(t *testing.T, configPath string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := run(args, fileConfigSource(configPath), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestEnvResolveEmitsPermissionsV2AtProductionEntry(t *testing.T) {
	schema := launchEnvFragmentV2Schema(t)
	cases := []struct {
		name              string
		machinePermission string
		locked            bool
		wantMode          string
		wantLocked        bool
		wantSource        string
		wantWarning       bool
	}{
		{name: "permissions-yolo-profile", machinePermission: "yolo", wantMode: "yolo", wantSource: "profile"},
		{name: "permissions-native-profile", machinePermission: "native", wantMode: "native", wantSource: "profile"},
		{name: "permissions-absent-default", wantMode: "native", wantSource: "default"},
		{name: "permissions-system-lock-global", machinePermission: "yolo", locked: true, wantMode: "native", wantLocked: true, wantSource: "global", wantWarning: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, home := profileHome(t)
			machine := map[string]any{
				"schema_version": 2,
				"skills_root":    filepath.Join(home, "skills"),
				"projects":       map[string]any{},
			}
			if tc.machinePermission != "" {
				machine["environments"] = map[string]any{
					"permissions": map[string]any{"acme": tc.machinePermission},
				}
			}
			machinePayload, err := json.Marshal(machine)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(source.path, machinePayload, 0o600); err != nil {
				t.Fatal(err)
			}
			systemPath := filepath.Join(t.TempDir(), "system.json")
			system := map[string]any{"schema_version": 2, "locked": []string{}}
			if tc.locked {
				system["locked"] = []string{"environments.permissions"}
				system["environments"] = map[string]any{"permissions": map[string]any{}}
			}
			systemPayload, err := json.Marshal(system)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(systemPath, systemPayload, 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)

			pkg := t.TempDir()
			writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
			if code, _, stderr := runFileConfigProfile(t, source.path, "profile", "install", pkg); code != exitOK {
				t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
			}
			code, stdout, stderr := runFileConfigProfile(t, source.path,
				"env", "resolve", "codex_cli", "--repair", "--format", "json")
			if code != exitOK {
				t.Fatalf("env resolve = %d\nstderr:\n%s", code, stderr)
			}
			if tc.wantWarning && (!strings.Contains(stderr, `config key "environments.permissions"`) || !strings.Contains(stderr, systemPath)) {
				t.Fatalf("locked resolve warning = %q, want manager §1 locked-key warning naming %s", stderr, systemPath)
			}
			var fragment map[string]any
			if err := json.Unmarshal([]byte(stdout), &fragment); err != nil {
				t.Fatalf("env resolve emitted malformed JSON: %v\n%s", err, stdout)
			}
			if fragment["fragment"] != "launch-env-fragment-v2" {
				t.Fatalf("fragment token = %v, want launch-env-fragment-v2", fragment["fragment"])
			}
			permission, ok := fragment["permissions"].(map[string]any)
			if !ok {
				t.Fatalf("v2 permissions member missing or malformed: %v", fragment["permissions"])
			}
			if permission["mode"] != tc.wantMode || permission["locked"] != tc.wantLocked || permission["source"] != tc.wantSource {
				t.Fatalf("permissions = %v for profile %v, want mode=%s locked=%t source=%s", permission, fragment["profile"], tc.wantMode, tc.wantLocked, tc.wantSource)
			}
			t.Run("schema-validation", func(t *testing.T) {
				if runtime.GOOS == "windows" {
					// The v2 schema's absolutePath rule accepts POSIX paths only.
					// The permissions assertions above still pass on Windows; skip
					// only this incompatible schema assertion until curator-spec
					// defines a Windows path form.
					t.Skip(launchEnvFragmentV2WindowsSchemaGap)
				}
				validateLaunchEnvFragmentV2(t, schema, []byte(stdout))
			})
		})
	}
}
