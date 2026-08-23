package skillspec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/moduleroots"
	"github.com/relux-works/curator/internal/verr"
)

// writeSkill lays out a snapshot directory with a csk-skill.json and any
// extra files (path -> content).
func writeSkill(t *testing.T, manifest string, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if manifest != "" {
		if err := os.WriteFile(filepath.Join(dir, "csk-skill.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mustFail(t *testing.T, dir, wantPathPrefix string) {
	t.Helper()
	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected error with path prefix %q", wantPathPrefix)
	}
	var v *verr.Error
	if !errors.As(err, &v) {
		t.Fatalf("error is not a validation error: %v", err)
	}
	if !strings.HasPrefix(v.Path, wantPathPrefix) {
		t.Fatalf("error path %q does not start with %q (message: %s)", v.Path, wantPathPrefix, v.Message)
	}
}

func TestPureContextSkill(t *testing.T) {
	dir := t.TempDir()
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if spec.SourceFile != "" || len(spec.Commands) != 0 {
		t.Fatalf("pure context skill: %+v", spec)
	}
}

func TestCanonicalManifestResolution(t *testing.T) {
	manifest := `{"schema_version":1,"commands":{}}`

	t.Run("canonical only", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, CanonicalManifestName), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		spec, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if spec.SourceFile != CanonicalManifestName {
			t.Fatalf("source file = %q", spec.SourceFile)
		}
	})

	t.Run("equal dual files select canonical", func(t *testing.T) {
		dir := t.TempDir()
		files := map[string]string{
			CanonicalManifestName: manifest,
			LegacyManifestName:    "{\n\"commands\": {}, \"schema_version\": 1\n}",
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		spec, err := Load(dir)
		if err != nil {
			t.Fatal(err)
		}
		if spec.SourceFile != CanonicalManifestName {
			t.Fatalf("source file = %q", spec.SourceFile)
		}
	})

	t.Run("conflicting dual files fail closed", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, CanonicalManifestName), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
		conflict := `{"schema_version":1,"commands":{"legacy":{"type":"system","command":"legacy"}}}`
		if err := os.WriteFile(filepath.Join(dir, LegacyManifestName), []byte(conflict), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Load(dir)
		if err == nil || !strings.Contains(err.Error(), "conflicting_skill_manifests") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("invalid peer does not fall back", func(t *testing.T) {
		for _, invalidName := range []string{CanonicalManifestName, LegacyManifestName} {
			t.Run(invalidName, func(t *testing.T) {
				dir := t.TempDir()
				for _, name := range []string{CanonicalManifestName, LegacyManifestName} {
					content := manifest
					if name == invalidName {
						content = "{"
					}
					if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := Load(dir); err == nil {
					t.Fatal("invalid peer was ignored")
				}
			})
		}
	})
}

func TestSchemaVersionValidation(t *testing.T) {
	mustFail(t, writeSkill(t, `{"schema_version": "3"}`, nil), "schema_version")
	mustFail(t, writeSkill(t, `{"schema_version": 3.5}`, nil), "schema_version")
	mustFail(t, writeSkill(t, `{"schema_version": 9}`, nil), "schema_version")
	mustFail(t, writeSkill(t, `{"schema_version": true}`, nil), "schema_version")
	mustFail(t, writeSkill(t, `{}`, nil), "schema_version")
}

func TestUnknownTopLevelFieldsRejectedFromV2(t *testing.T) {
	// v1 tolerates unknown fields.
	dir := writeSkill(t, `{"schema_version": 1, "install": "curl | sh"}`, nil)
	if _, err := Load(dir); err != nil {
		t.Fatalf("v1 unknown fields must be tolerated: %v", err)
	}
	// v2 rejects them: no install hooks (Spec 5.9).
	mustFail(t, writeSkill(t, `{"schema_version": 2, "install": "curl | sh"}`, nil), "csk-skill.json")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "capabilities": {}}`, nil), "csk-skill.json")
}

func TestCapabilitiesRequiredFromV3(t *testing.T) {
	mustFail(t, writeSkill(t, `{"schema_version": 3}`, nil), "capabilities")
	dir := writeSkill(t, `{"schema_version": 3, "capabilities": {"network": "none"}}`, nil)
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Capabilities.Filesystem.Keyword != "repo" {
		t.Fatalf("capabilities defaults: %+v", spec.Capabilities)
	}
}

func TestRuntimeRoots(t *testing.T) {
	files := map[string]string{"scripts/tool": "#!/bin/sh\n", "lib/util.py": ""}
	ok := writeSkill(t, `{"schema_version": 2, "runtime_roots": ["scripts", "lib"]}`, files)
	spec, err := Load(ok)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.RuntimeRoots) != 2 {
		t.Fatalf("runtime roots: %v", spec.RuntimeRoots)
	}

	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["missing"]}`, nil), "runtime_roots[0]")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["scripts/tool"]}`, files), "runtime_roots[0]")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["scripts", "scripts"]}`, files), "runtime_roots")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["../out"]}`, files), "runtime_roots[0]")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["a//b"]}`, files), "runtime_roots[0]")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["./scripts"]}`, files), "runtime_roots[0]")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["NUL"]}`, nil), "runtime_roots[0]")
	// nested roots are not disjoint
	nested := map[string]string{"scripts/tool": "", "scripts/inner/x": ""}
	mustFail(t, writeSkill(t, `{"schema_version": 2, "runtime_roots": ["scripts", "scripts/inner"]}`, nested), "runtime_roots")
}

func TestScriptCommands(t *testing.T) {
	files := map[string]string{"scripts/ytx": "", "scripts/ytx.cmd": ""}
	manifest := `{"schema_version": 2, "runtime_roots": ["scripts"],
		"commands": {"ytx": {"type": "script", "unix_path": "scripts/ytx", "win_path": "scripts/ytx.cmd"}}}`
	spec, err := Load(writeSkill(t, manifest, files))
	if err != nil {
		t.Fatal(err)
	}
	command := spec.Commands["ytx"]
	if command.UnixPath != "scripts/ytx" || command.WinPath != "scripts/ytx.cmd" {
		t.Fatalf("command: %+v", command)
	}

	// v2 requires at least one path
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "script"}}}`, nil), "commands.x")
	// path must exist
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "script", "unix_path": "scripts/x"}}}`, nil), "commands.x.unix_path")
	// path must sit inside a runtime root when roots are declared
	outside := map[string]string{"scripts/tool": "", "other/x": ""}
	mustFail(t, writeSkill(t,
		`{"schema_version": 2, "runtime_roots": ["scripts"], "commands": {"x": {"type": "script", "unix_path": "other/x"}}}`,
		outside), "commands.x.unix_path")
	// absolute and escaping paths rejected
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "script", "unix_path": "/etc/passwd"}}}`, nil), "commands.x.unix_path")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "script", "unix_path": "../up"}}}`, nil), "commands.x.unix_path")
	// unknown fields on the command object (v2+)
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "script", "unix_path": "s", "check": true}}}`, map[string]string{"s": ""}), "commands.x")
	// invalid command name
	mustFail(t, writeSkill(t, `{"schema_version": 1, "commands": {"-bad": {"type": "system", "command": "b"}}}`, nil), "commands.-bad")
	// unsupported type
	mustFail(t, writeSkill(t, `{"schema_version": 1, "commands": {"x": {"type": "binary"}}}`, nil), "commands.x")
}

func TestSystemCommands(t *testing.T) {
	spec, err := Load(writeSkill(t, `{"schema_version": 1,
		"commands": {"wiki": {"type": "system", "command": "wiki", "hint": "install it"}}}`, nil))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Commands["wiki"].Command != "wiki" || spec.Commands["wiki"].Hint != "install it" {
		t.Fatalf("system command: %+v", spec.Commands["wiki"])
	}
	mustFail(t, writeSkill(t, `{"schema_version": 1, "commands": {"x": {"type": "system"}}}`, nil), "commands.x")
	mustFail(t, writeSkill(t, `{"schema_version": 1, "commands": {"x": {"type": "system", "command": ""}}}`, nil), "commands.x")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "commands": {"x": {"type": "system", "command": "b", "hint": 5}}}`, nil), "commands.x.hint")
}

func TestDependenciesGating(t *testing.T) {
	mustFail(t, writeSkill(t, `{"schema_version": 1, "dependencies": {"commands": {}}}`, nil), "dependencies")
	mustFail(t, writeSkill(t, `{"schema_version": 3, "capabilities": {}, "dependencies": {"skills": {}}}`, nil), "dependencies.skills")
	mustFail(t, writeSkill(t, `{"schema_version": 4, "capabilities": {}, "dependencies": {"mcp_servers": {}}}`, nil), "dependencies.mcp_servers")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": {"extras": {}}}`, nil), "dependencies")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": []}`, nil), "dependencies")
}

func TestCommandDependencies(t *testing.T) {
	spec, err := Load(writeSkill(t, `{"schema_version": 2, "dependencies": {"commands": {
		"glab": {"type": "system", "command": "glab", "hint": "bootstrap"},
		"wk": {"type": "skill", "skill": "skill-wiki", "command": "wk"}}}}`, nil))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Dependencies["glab"].Type != "system" || spec.Dependencies["wk"].Skill != "skill-wiki" {
		t.Fatalf("dependencies: %+v", spec.Dependencies)
	}
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": {"commands": {"x": {"type": "system"}}}}`, nil), "dependencies.commands.x")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": {"commands": {"x": {"type": "skill", "skill": "s"}}}}`, nil), "dependencies.commands.x")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": {"commands": {"x": {"type": "npm"}}}}`, nil), "dependencies.commands.x")
	mustFail(t, writeSkill(t, `{"schema_version": 2, "dependencies": {"commands": {"x": {"type": "system", "command": "b", "install": "y"}}}}`, nil), "dependencies.commands.x")
}

func TestRequirements(t *testing.T) {
	manifest := `{"schema_version": 4, "capabilities": {}, "dependencies": {"skills": {
		"skill-wiki": {"git": "git@example.com:skills/skill-wiki.git",
			"ref": {"kind": "tag", "value": "v1.4.2"}, "mode": "runtime", "commands": ["wk", "wk"]}}}}`
	spec, err := Load(writeSkill(t, manifest, nil))
	if err != nil {
		t.Fatal(err)
	}
	requirement := spec.Requirements["skill-wiki"]
	if requirement.RefKind != "tag" || requirement.RefValue != "v1.4.2" || requirement.Mode != "runtime" {
		t.Fatalf("requirement: %+v", requirement)
	}
	if len(requirement.Commands) != 1 || requirement.Commands[0] != "wk" {
		t.Fatalf("commands not deduplicated: %v", requirement.Commands)
	}

	base := `{"schema_version": 4, "capabilities": {}, "dependencies": {"skills": {"x": %s}}}`
	fail := func(entry, pathPrefix string) {
		t.Helper()
		mustFail(t, writeSkill(t, strings.Replace(base, "%s", entry, 1), nil), pathPrefix)
	}
	fail(`{"git": "u", "ref": {"kind": "branch", "value": "main"}}`, "dependencies.skills.x.ref")
	fail(`{"git": "u", "ref": {"kind": "semver", "value": "1"}}`, "dependencies.skills.x.ref.kind")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": "^1.0"}}`, "dependencies.skills.x.ref.value")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": ">= 1"}}`, "dependencies.skills.x.ref.value")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": ""}}`, "dependencies.skills.x.ref.value")
	fail(`{"git": "u", "version": "1.2.3"}`, "dependencies.skills.x")
	fail(`{"ref": {"kind": "tag", "value": "v1"}}`, "dependencies.skills.x")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": "v1"}, "mode": "lazy"}`, "dependencies.skills.x.mode")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": "v1"}, "commands": ["c"]}`, "dependencies.skills.x.commands")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": "v1"}, "mode": "runtime", "commands": []}`, "dependencies.skills.x.commands")
	fail(`{"git": "u", "ref": {"kind": "tag", "value": "v1", "extra": 1}}`, "dependencies.skills.x.ref")
}

func TestMcpServers(t *testing.T) {
	manifest := `{"schema_version": 5, "capabilities": {}, "dependencies": {"mcp_servers": {
		"google-sheets": {"hint": "connect it", "transport": "http", "required_in": "all"},
		"tracker": {"hint": "connect tracker"}}}}`
	spec, err := Load(writeSkill(t, manifest, nil))
	if err != nil {
		t.Fatal(err)
	}
	if spec.McpServers["google-sheets"].RequiredIn != "all" || spec.McpServers["google-sheets"].Transport != "http" {
		t.Fatalf("mcp: %+v", spec.McpServers["google-sheets"])
	}
	if spec.McpServers["tracker"].RequiredIn != "any" {
		t.Fatalf("required_in default: %+v", spec.McpServers["tracker"])
	}

	base := `{"schema_version": 5, "capabilities": {}, "dependencies": {"mcp_servers": {"x": %s}}}`
	fail := func(entry, pathPrefix string) {
		t.Helper()
		mustFail(t, writeSkill(t, strings.Replace(base, "%s", entry, 1), nil), pathPrefix)
	}
	fail(`{}`, "dependencies.mcp_servers.x")
	fail(`{"hint": ""}`, "dependencies.mcp_servers.x")
	fail(`{"hint": "h", "transport": "grpc"}`, "dependencies.mcp_servers.x.transport")
	fail(`{"hint": "h", "required_in": "some"}`, "dependencies.mcp_servers.x.required_in")
	fail(`{"hint": "h", "port": 1}`, "dependencies.mcp_servers.x")
}

func TestLegacyRuntimeFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := `{"commands": {"ytx": "scripts/ytx", "ytx.cmd": "scripts/ytx.cmd"}}`
	if err := os.WriteFile(filepath.Join(dir, "agents", "runtime.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if spec.SourceFile != "agents/runtime.json" {
		t.Fatalf("source file: %q", spec.SourceFile)
	}
	if spec.Commands["ytx"].WinPath != "" {
		t.Fatalf("plain path must not double as win path: %+v", spec.Commands["ytx"])
	}
	if spec.Commands["ytx.cmd"].WinPath != "scripts/ytx.cmd" {
		t.Fatalf(".cmd path must double as win path: %+v", spec.Commands["ytx.cmd"])
	}
}

func TestLegacyManifestWinsOverRuntimeFallback(t *testing.T) {
	dir := writeSkill(t, `{"schema_version": 1}`, map[string]string{
		"agents/runtime.json": `{"commands": {"x": "s"}}`,
	})
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if spec.SourceFile != LegacyManifestName || len(spec.Commands) != 0 {
		t.Fatalf("precedence: %+v", spec)
	}
}

// schema8Build is a schema-8 manifest with one local go-v1 build command whose
// `modules` list is supplied by the caller.
func schema8Build(modules string) string {
	return `{"schema_version":8,"capabilities":{},"build_roots":["tools/cli"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"tools/cli/cmd/tool"` +
		modules + `}}}`
}

// buildSnapshotFiles is the minimum snapshot a schema8Build manifest needs.
func buildSnapshotFiles(extra map[string]string) map[string]string {
	files := map[string]string{
		"tools/cli/go.mod":            "module fixture\n",
		"tools/cli/cmd/tool/main.go":  "package main\n",
		"pkg/board/go.mod":            "module fixture\n",
		"pkg/remoteconfig/go.mod":     "module fixture\n",
		"pkg/board/codec/placeholder": "",
	}
	for path, content := range extra {
		files[path] = content
	}
	return files
}

func TestSchema8ModuleRootsAreParsed(t *testing.T) {
	dir := writeSkill(t, schema8Build(`,"modules":["pkg/board","pkg/remoteconfig"]`), buildSnapshotFiles(nil))
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	modules := spec.Commands["tool"].Modules
	if len(modules) != 2 || modules[0] != "pkg/board" || modules[1] != "pkg/remoteconfig" {
		t.Fatalf("modules = %v", modules)
	}
}

// TestSchema8AbsentAndEmptyModuleListsAgree proves an empty list is admitted
// and carries the same meaning as an absent one: a single-module build root.
func TestSchema8AbsentAndEmptyModuleListsAgree(t *testing.T) {
	for name, manifest := range map[string]string{
		"absent": schema8Build(""),
		"empty":  schema8Build(`,"modules":[]`),
	} {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, manifest, buildSnapshotFiles(nil))
			spec, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(spec.Commands["tool"].Modules) != 0 {
				t.Fatalf("modules = %v", spec.Commands["tool"].Modules)
			}
		})
	}
}

func TestSchema8ModuleRootRejections(t *testing.T) {
	for name, modules := range map[string]string{
		"dot":                      `,"modules":["."]`,
		"parent":                   `,"modules":["../pkg/lib"]`,
		"absolute":                 `,"modules":["/pkg/lib"]`,
		"backslash":                `,"modules":["pkg\\lib"]`,
		"windows device":           `,"modules":["pkg/CON"]`,
		"duplicate":                `,"modules":["pkg/board","pkg/board"]`,
		"absent from the snapshot": `,"modules":["pkg/missing"]`,
		"no go.mod":                `,"modules":["pkg/board/codec"]`,
		"inside the build root":    `,"modules":["tools/cli"]`,
		"nested declarations":      `,"modules":["pkg/board","pkg/board/codec"]`,
		"null":                     `,"modules":null`,
		"string":                   `,"modules":"pkg/board"`,
		"non-string entry":         `,"modules":[7]`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, schema8Build(modules), buildSnapshotFiles(nil))
			if _, err := Load(dir); err == nil {
				t.Fatal("expected the manifest to be rejected")
			}
		})
	}
}

// TestSchema8ModuleRootDiagnostics pins the closed diagnostics the manager
// profile publishes for the declaration and containment half.
func TestSchema8ModuleRootDiagnostics(t *testing.T) {
	nestedModule := map[string]string{"pkg/board/codec/go.mod": "module fixture\n"}
	for name, testCase := range map[string]struct {
		modules string
		extra   map[string]string
		code    string
	}{
		"unportable path":       {modules: `,"modules":["../pkg/lib"]`, code: moduleroots.DeclarationInvalid},
		"missing go.mod":        {modules: `,"modules":["pkg/board/codec"]`, code: moduleroots.DeclarationInvalid},
		"nested declarations":   {modules: `,"modules":["pkg/board","pkg/board/codec"]`, extra: nestedModule, code: moduleroots.ContainmentInvalid},
		"inside the build root": {modules: `,"modules":["tools/cli"]`, code: moduleroots.ContainmentInvalid},
	} {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, schema8Build(testCase.modules), buildSnapshotFiles(testCase.extra))
			_, err := Load(dir)
			if err == nil {
				t.Fatalf("expected %s", testCase.code)
			}
			if got := moduleroots.Code(err); got != testCase.code {
				t.Fatalf("diagnostic = %q (%v), want %q", got, err, testCase.code)
			}
		})
	}
}

// TestSchema8ModuleRootsRejectRuntimeRootOverlap keeps a declared module
// directory out of every runtime root.
func TestSchema8ModuleRootsRejectRuntimeRootOverlap(t *testing.T) {
	manifest := `{"schema_version":8,"capabilities":{},"runtime_roots":["pkg"],"build_roots":["tools/cli"],` +
		`"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"tools/cli/cmd/tool","modules":["pkg/board"]}}}`
	dir := writeSkill(t, manifest, buildSnapshotFiles(nil))
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected the runtime-root overlap to be rejected")
	}
	if got := moduleroots.Code(err); got != moduleroots.ContainmentInvalid {
		t.Fatalf("diagnostic = %q (%v)", got, err)
	}
}

// TestModulesRequireSchema8 proves schemas 1 through 7 do not acquire the
// field: they reject it as unknown on the command surface.
func TestModulesRequireSchema8(t *testing.T) {
	for _, schema := range []int{6, 7} {
		t.Run(fmt.Sprintf("schema %d", schema), func(t *testing.T) {
			manifest := fmt.Sprintf(`{"schema_version":%d,"capabilities":{},"build_roots":["tools/cli"],`+
				`"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"tools/cli/cmd/tool","modules":["pkg/board"]}}}`, schema)
			dir := writeSkill(t, manifest, buildSnapshotFiles(nil))
			mustFail(t, dir, "commands.tool.modules")
		})
	}
}

// TestTopLevelModulesRejected proves the field is per command: schema 8 has no
// manifest-level module declaration.
func TestTopLevelModulesRejected(t *testing.T) {
	manifest := `{"schema_version":8,"capabilities":{},"modules":["pkg/board"],"build_roots":["tools/cli"],` +
		`"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"tools/cli/cmd/tool"}}}`
	dir := writeSkill(t, manifest, buildSnapshotFiles(nil))
	mustFail(t, dir, LegacyManifestName)
}

// scriptManifest is a schema-8 manifest with one script command carrying the
// supplied execution surface.
func scriptManifest(schema int, execution string) string {
	return fmt.Sprintf(`{"schema_version":%d,"capabilities":{},"runtime_roots":["scripts"],`+
		`"commands":{"tool":{"type":"script","unix_path":"scripts/tool"%s}}}`, schema, execution)
}

func TestSchema8ScriptExecutionPolicyIsParsed(t *testing.T) {
	for _, interpreter := range []string{"node-v1", "python3-v1"} {
		t.Run(interpreter, func(t *testing.T) {
			manifest := scriptManifest(8, fmt.Sprintf(`,"execution_policy":"script-worker-v1","interpreter":%q`, interpreter))
			dir := writeSkill(t, manifest, map[string]string{"scripts/tool": "#!/bin/sh\n"})
			spec, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			command := spec.Commands["tool"]
			if command.ExecutionPolicy != ScriptExecutionPolicy || command.Interpreter != interpreter {
				t.Fatalf("execution surface = %q/%q", command.ExecutionPolicy, command.Interpreter)
			}
		})
	}
}

// TestSchema8DeclaredOnlyScriptKeepsSchema7Meaning proves the absence of
// `execution_policy` is the default and the only spelling of declared-only.
func TestSchema8DeclaredOnlyScriptKeepsSchema7Meaning(t *testing.T) {
	dir := writeSkill(t, scriptManifest(8, ""), map[string]string{"scripts/tool": "#!/bin/sh\n"})
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	command := spec.Commands["tool"]
	if command.ExecutionPolicy != "" || command.Interpreter != "" {
		t.Fatalf("declared-only command carries %q/%q", command.ExecutionPolicy, command.Interpreter)
	}
}

func TestSchema8ScriptExecutionRejections(t *testing.T) {
	for name, testCase := range map[string]struct{ execution, wantPath string }{
		"policy without interpreter": {`,"execution_policy":"script-worker-v1"`, "commands.tool.execution_policy"},
		"interpreter without policy": {`,"interpreter":"python3-v1"`, "commands.tool.interpreter"},
		"null policy":                {`,"execution_policy":null,"interpreter":"python3-v1"`, "commands.tool.execution_policy"},
		"opt-out policy":             {`,"execution_policy":"none","interpreter":"python3-v1"`, "commands.tool.execution_policy"},
		"compiled policy":            {`,"execution_policy":"manager-worker-v1","interpreter":"python3-v1"`, "commands.tool.execution_policy"},
		"hardened policy":            {`,"execution_policy":"hardened-worker-v1","interpreter":"python3-v1"`, "commands.tool.execution_policy"},
		"successor policy":           {`,"execution_policy":"script-worker-v2","interpreter":"python3-v1"`, "commands.tool.execution_policy"},
		"shell interpreter":          {`,"execution_policy":"script-worker-v1","interpreter":"bash-v1"`, "commands.tool.interpreter"},
		"null interpreter":           {`,"execution_policy":"script-worker-v1","interpreter":null`, "commands.tool.interpreter"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, scriptManifest(8, testCase.execution), map[string]string{"scripts/tool": "#!/bin/sh\n"})
			mustFail(t, dir, testCase.wantPath)
		})
	}
}

// TestScriptExecutionSurfaceRequiresSchema8 proves schemas 2 through 7 reject
// both fields as unknown rather than acquiring schema-8 meaning.
func TestScriptExecutionSurfaceRequiresSchema8(t *testing.T) {
	for _, schema := range []int{6, 7} {
		for _, execution := range []string{
			`,"execution_policy":"script-worker-v1","interpreter":"python3-v1"`,
			`,"interpreter":"python3-v1"`,
		} {
			t.Run(fmt.Sprintf("schema %d%s", schema, execution), func(t *testing.T) {
				dir := writeSkill(t, scriptManifest(schema, execution), map[string]string{"scripts/tool": "#!/bin/sh\n"})
				mustFail(t, dir, "commands.tool")
			})
		}
	}
}

// TestScriptExecutionSurfaceRejectedOnOtherCommandTypes keeps the schema-8
// script fields off the system and build surfaces, whose closed field sets do
// not admit them at any schema.
func TestScriptExecutionSurfaceRejectedOnOtherCommandTypes(t *testing.T) {
	system := `{"schema_version":8,"capabilities":{},"commands":{"tool":{"type":"system","command":"tool",` +
		`"execution_policy":"script-worker-v1","interpreter":"python3-v1"}}}`
	build := `{"schema_version":8,"capabilities":{},"build_roots":["tools/cli"],"commands":{"tool":{"type":"build",` +
		`"driver":"go-v1","source_dir":"tools/cli/cmd/tool","execution_policy":"script-worker-v1","interpreter":"python3-v1"}}}`
	for name, manifest := range map[string]string{"system": system, "build": build} {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, manifest, buildSnapshotFiles(nil))
			mustFail(t, dir, "commands.tool")
		})
	}
}

// TestSchema9IsNotAccepted keeps the supported range closed at the schema this
// build implements.
func TestSchema9IsNotAccepted(t *testing.T) {
	dir := writeSkill(t, `{"schema_version":9,"capabilities":{},"commands":{}}`, nil)
	mustFail(t, dir, "schema_version")
}

// TestSchema1DoesNotAcquireSchema8Meaning proves the deployed schema-1
// extension tolerance stays tolerance: an unknown command field is ignored,
// never read as an enforcement claim or a module declaration.
func TestSchema1DoesNotAcquireSchema8Meaning(t *testing.T) {
	manifest := `{"schema_version":1,"commands":{"tool":{"type":"script","unix_path":"scripts/tool",` +
		`"execution_policy":"script-worker-v1","interpreter":"python3-v1"}}}`
	dir := writeSkill(t, manifest, map[string]string{"scripts/tool": "#!/bin/sh\n"})
	spec, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	command := spec.Commands["tool"]
	if command.ExecutionPolicy != "" || command.Interpreter != "" {
		t.Fatalf("schema 1 acquired schema-8 meaning: %q/%q", command.ExecutionPolicy, command.Interpreter)
	}
}
