package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// moduleRootsSkill writes a schema-8 build skill whose single build command
// declares one first-party module root, plus the runtime root the declaration
// must stay disjoint from.
func (e *env) moduleRootsSkill(name, command string, modules []string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "src/go.mod", "module example.com/"+name+"\n")
	e.write(dir, "src/cmd/"+command+"/main.go", "package main\n\nfunc main() {}\n")
	e.write(dir, "src/vendor/modules.txt", "# example.com/board => ../pkg/board\n")
	for _, module := range modules {
		e.write(dir, module+"/go.mod", "module example.com/board\n")
	}
	e.write(dir, "scripts/run.sh", "#!/bin/sh\nexit 0\n")
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 8,
		"build_roots":    []string{"src"},
		"runtime_roots":  []string{"scripts"},
		"capabilities":   map[string]any{},
		"commands": map[string]any{
			command: map[string]any{
				"type": "build", "driver": "go-v1", "source_dir": "src/cmd/" + command, "modules": modules,
			},
		},
	}, "", "  ")
	if err != nil {
		e.t.Fatal(err)
	}
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// TestDeclaredModuleRootsReachTheBuilder proves the schema-8 declaration is
// carried all the way from the manifest to the staging boundary. The driver
// re-validates it there, so a plan that dropped it would silently fall back to
// the single-module build root and reject the very build the package declared.
func TestDeclaredModuleRootsReachTheBuilder(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.moduleRootsSkill("module-skill", "alpha", []string{"pkg/board"})
	e.declare("module-skill")
	deps, _, _, builder := newFakeDeps(t)

	var observed StageRequest
	builder.observe = func(request StageRequest) { observed = request }

	result := e.install(Options{Build: deps})
	if result.Status == "failed" {
		t.Fatalf("install failed: %v", result.Errors)
	}
	if len(builder.calls) != 1 || builder.calls[0] != "alpha" {
		t.Fatalf("builder calls = %v, want exactly [alpha]", builder.calls)
	}
	if want := []string{"pkg/board"}; !reflect.DeepEqual(observed.Modules, want) {
		t.Fatalf("staged modules = %q, want %q", observed.Modules, want)
	}
	if want := []string{"scripts"}; !reflect.DeepEqual(observed.RuntimeRoots, want) {
		t.Fatalf("staged runtime roots = %q, want %q", observed.RuntimeRoots, want)
	}
	// The command surface handed to the driver is the package's own, so the
	// driver can refuse a modules list the manager never validated.
	declared, ok := observed.CommandObject["modules"].([]string)
	if !ok || !reflect.DeepEqual(declared, []string{"pkg/board"}) {
		t.Fatalf("command object modules = %#v, want the declared list", observed.CommandObject["modules"])
	}
}

// TestASchemaSixCommandDeclaresNoModules keeps the pre-schema-8 surface exactly
// three fields wide: an absent list is not spelled as an empty one.
func TestASchemaSixCommandDeclaresNoModules(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, _, builder := newFakeDeps(t)

	var observed StageRequest
	builder.observe = func(request StageRequest) { observed = request }

	result := e.install(Options{Build: deps})
	if result.Status == "failed" {
		t.Fatalf("install failed: %v", result.Errors)
	}
	if len(observed.Modules) != 0 {
		t.Fatalf("staged modules = %q, want none", observed.Modules)
	}
	if _, declared := observed.CommandObject["modules"]; declared {
		t.Fatalf("command object = %#v, want no modules field", observed.CommandObject)
	}
}
