package skillspec

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/verr"
)

const validBuildCommand = `"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool"}`

func buildFiles() map[string]string {
	return map[string]string{
		"build/go.mod":           "module example.com/tool\n\ngo 1.23\n",
		"build/cmd/tool/main.go": "package main\n",
		"scripts/helper":         "#!/bin/sh\n",
	}
}

func mustFailExact(t *testing.T, dir, wantPath string) {
	t.Helper()
	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected validation error at %q", wantPath)
	}
	var validation *verr.Error
	if !errors.As(err, &validation) {
		t.Fatalf("error is not a validation error: %v", err)
	}
	if validation.Path != wantPath {
		t.Fatalf("error path = %q, want %q (message: %s)", validation.Path, wantPath, validation.Message)
	}
}

func TestSchemaV6BuildManifestAliasesAndMixedCommands(t *testing.T) {
	manifest := `{"schema_version":6,"runtime_roots":["scripts"],"build_roots":["build"],"capabilities":{},"commands":{` +
		validBuildCommand + `,"helper":{"type":"script","unix_path":"scripts/helper"},"git":{"type":"system","command":"git"}}}`

	for _, manifestName := range []string{CanonicalManifestName, LegacyManifestName} {
		t.Run(manifestName, func(t *testing.T) {
			dir := writeSkill(t, "", buildFiles())
			if err := os.WriteFile(filepath.Join(dir, manifestName), []byte(manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			spec, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if spec.SchemaVersion != 6 || spec.SourceFile != manifestName {
				t.Fatalf("manifest identity: %+v", spec)
			}
			if !reflect.DeepEqual(spec.BuildRoots, []string{"build"}) {
				t.Fatalf("build roots = %v", spec.BuildRoots)
			}
			command := spec.Commands["tool"]
			if command.Type != "build" || command.Driver != "go-v1" || command.SourceDir != "build/cmd/tool" {
				t.Fatalf("build command = %+v", command)
			}
			if spec.Commands["helper"].UnixPath != "scripts/helper" || spec.Commands["git"].Command != "git" {
				t.Fatalf("mixed commands = %+v", spec.Commands)
			}
		})
	}
}

func TestSchemaV6BuildCommandClosedShapeDiagnostics(t *testing.T) {
	mustFailExact(t, writeSkill(t,
		`{"schema_version":5,"capabilities":{},"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool"}}}`,
		buildFiles()), "commands.tool")
	mustFailExact(t, writeSkill(t,
		`{"schema_version":5,"capabilities":{},"build_roots":[]}`,
		buildFiles()), LegacyManifestName)

	manifest := func(roots, command string) string {
		return `{"schema_version":6,"capabilities":{},"build_roots":` + roots + `,"commands":{"tool":` + command + `}}`
	}
	cases := []struct {
		name     string
		roots    string
		command  string
		wantPath string
	}{
		{name: "missing driver", roots: `["build"]`, command: `{"type":"build","source_dir":"build/cmd/tool"}`, wantPath: "commands.tool.driver"},
		{name: "unknown driver", roots: `["build"]`, command: `{"type":"build","driver":"custom-v1","source_dir":"build/cmd/tool"}`, wantPath: "commands.tool.driver"},
		{name: "missing source dir", roots: `["build"]`, command: `{"type":"build","driver":"go-v1"}`, wantPath: "commands.tool.source_dir"},
		{name: "source dir dot", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"."}`, wantPath: "commands.tool.source_dir"},
		{name: "mixed script", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","unix_path":"scripts/tool"}`, wantPath: "commands.tool.unix_path"},
		{name: "mixed system", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","command":"tool"}`, wantPath: "commands.tool.command"},
		{name: "args", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","args":[]}`, wantPath: "commands.tool.args"},
		{name: "env", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","env":{}}`, wantPath: "commands.tool.env"},
		{name: "flags", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","flags":[]}`, wantPath: "commands.tool.flags"},
		{name: "output", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","output":"bin/tool"}`, wantPath: "commands.tool.output"},
		{name: "toolchain", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","toolchain":"go1.25"}`, wantPath: "commands.tool.toolchain"},
		{name: "hooks", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","hooks":[]}`, wantPath: "commands.tool.hooks"},
		{name: "scripts", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","scripts":[]}`, wantPath: "commands.tool.scripts"},
		{name: "tags", roots: `["build"]`, command: `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","tags":[]}`, wantPath: "commands.tool.tags"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mustFailExact(t, writeSkill(t, manifest(testCase.roots, testCase.command), buildFiles()), testCase.wantPath)
		})
	}
}

func TestLegacyRuntimeFallbackRejectsBuildObject(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := `{"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool"}}}`
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(RuntimeFallbackName)), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFailExact(t, dir, "commands.tool")
}

func TestSchemaV6BuildRootDiagnostics(t *testing.T) {
	manifest := func(runtimeRoots, buildRoots, commands string) string {
		return `{"schema_version":6,"capabilities":{},"runtime_roots":` + runtimeRoots + `,"build_roots":` + buildRoots + `,"commands":` + commands + `}`
	}
	cases := []struct {
		name         string
		runtimeRoots string
		buildRoots   string
		commands     string
		files        map[string]string
		wantPath     string
	}{
		{name: "missing root", runtimeRoots: `[]`, buildRoots: `["missing"]`, commands: `{}`, files: buildFiles(), wantPath: "build_roots[0]"},
		{name: "root is file", runtimeRoots: `[]`, buildRoots: `["build"]`, commands: `{}`, files: map[string]string{"build": "not a directory"}, wantPath: "build_roots[0]"},
		{name: "dot root", runtimeRoots: `[]`, buildRoots: `["."]`, commands: `{}`, files: buildFiles(), wantPath: "build_roots[0]"},
		{name: "escaped root", runtimeRoots: `[]`, buildRoots: `["../build"]`, commands: `{}`, files: buildFiles(), wantPath: "build_roots[0]"},
		{name: "duplicate roots", runtimeRoots: `[]`, buildRoots: `["build","build"]`, commands: `{` + validBuildCommand + `}`, files: buildFiles(), wantPath: "build_roots"},
		{name: "overlapping roots", runtimeRoots: `[]`, buildRoots: `["build","build/nested"]`, commands: `{` + validBuildCommand + `}`, files: map[string]string{"build/go.mod": "module x\n", "build/cmd/tool/main.go": "package main\n", "build/nested/go.mod": "module y\n"}, wantPath: "build_roots"},
		{name: "equal runtime overlap", runtimeRoots: `["build"]`, buildRoots: `["build"]`, commands: `{` + validBuildCommand + `}`, files: buildFiles(), wantPath: "build_roots"},
		{name: "contained runtime overlap", runtimeRoots: `["build/runtime"]`, buildRoots: `["build"]`, commands: `{` + validBuildCommand + `}`, files: map[string]string{"build/go.mod": "module x\n", "build/cmd/tool/main.go": "package main\n", "build/runtime/file": ""}, wantPath: "build_roots"},
		{name: "unused root", runtimeRoots: `[]`, buildRoots: `["build","other"]`, commands: `{` + validBuildCommand + `}`, files: map[string]string{"build/go.mod": "module x\n", "build/cmd/tool/main.go": "package main\n", "other/go.mod": "module y\n"}, wantPath: "build_roots[1]"},
		{name: "missing roots declaration", runtimeRoots: `[]`, buildRoots: `[]`, commands: `{` + validBuildCommand + `}`, files: buildFiles(), wantPath: "commands.tool.source_dir"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mustFailExact(t, writeSkill(t, manifest(testCase.runtimeRoots, testCase.buildRoots, testCase.commands), testCase.files), testCase.wantPath)
		})
	}
}

func TestSchemaV6BuildPathAndModuleDiagnostics(t *testing.T) {
	manifest := func(sourceDir string) string {
		return `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"` + sourceDir + `"}}}`
	}
	cases := []struct {
		name      string
		sourceDir string
		files     map[string]string
	}{
		{name: "escaped source", sourceDir: "../tool", files: buildFiles()},
		{name: "outside root", sourceDir: "other/tool", files: map[string]string{"build/go.mod": "module x\n", "other/tool/main.go": "package main\n"}},
		{name: "missing source", sourceDir: "build/cmd/missing", files: map[string]string{"build/go.mod": "module x\n"}},
		{name: "source is file", sourceDir: "build/cmd/tool", files: map[string]string{"build/go.mod": "module x\n", "build/cmd/tool": "package main\n"}},
		{name: "missing root go mod", sourceDir: "build/cmd/tool", files: map[string]string{"build/cmd/tool/main.go": "package main\n"}},
		{name: "intervening module", sourceDir: "build/cmd/tool", files: map[string]string{"build/go.mod": "module x\n", "build/cmd/go.mod": "module x/cmd\n", "build/cmd/tool/main.go": "package main\n"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			mustFailExact(t, writeSkill(t, manifest(testCase.sourceDir), testCase.files), "commands.tool.source_dir")
		})
	}
}

func TestSchemaV6AllowsSourceDirEqualBuildRoot(t *testing.T) {
	manifest := `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build"}}}`
	spec, err := Load(writeSkill(t, manifest, map[string]string{"build/go.mod": "module x\n", "build/main.go": "package main\n"}))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Commands["tool"].SourceDir != "build" {
		t.Fatalf("command = %+v", spec.Commands["tool"])
	}
}

func TestSchemaV6UsesEveryBuildRoot(t *testing.T) {
	manifest := `{"schema_version":6,"capabilities":{},"build_roots":["first","second"],"commands":{` +
		`"one":{"type":"build","driver":"go-v1","source_dir":"first/cmd/one"},` +
		`"two":{"type":"build","driver":"go-v1","source_dir":"second"}}}`
	files := map[string]string{
		"first/go.mod":          "module first\n",
		"first/cmd/one/main.go": "package main\n",
		"second/go.mod":         "module second\n",
		"second/main.go":        "package main\n",
	}
	spec, err := Load(writeSkill(t, manifest, files))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.BuildRoots, []string{"first", "second"}) {
		t.Fatalf("build roots = %v", spec.BuildRoots)
	}
}

func TestSchemaV6SystemCommandSchemaChecksRemainVersionGated(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		manifest string
		field    string
	}{
		{name: "non-identifier command", manifest: `{"schema_version":6,"capabilities":{},"commands":{"tool":{"type":"system","command":"bin/tool"}}}`, field: "command"},
		{name: "empty hint", manifest: `{"schema_version":6,"capabilities":{},"commands":{"tool":{"type":"system","command":"tool","hint":""}}}`, field: "hint"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			mustFailExact(t, writeSkill(t, testCase.manifest, nil), "commands.tool."+testCase.field)
		})
	}

	legacy, err := Load(writeSkill(t,
		`{"schema_version":5,"capabilities":{},"commands":{"tool":{"type":"system","command":"bin/tool","hint":""}}}`,
		nil))
	if err != nil {
		t.Fatalf("schema 5 behavior changed: %v", err)
	}
	if legacy.Commands["tool"].Command != "bin/tool" || legacy.Commands["tool"].Hint != "" {
		t.Fatalf("schema 5 command = %+v", legacy.Commands["tool"])
	}
}

func TestSchemaV6RejectsLinkedBuildPaths(t *testing.T) {
	t.Run("build root", func(t *testing.T) {
		dir := writeSkill(t, `{"schema_version":6,"capabilities":{},"build_roots":["linked"],"commands":{}}`, nil)
		target := filepath.Join(t.TempDir(), "target")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(dir, "linked")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		mustFailExact(t, dir, "build_roots[0]")
	})

	t.Run("source directory", func(t *testing.T) {
		dir := writeSkill(t, `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{`+validBuildCommand+`}}`, map[string]string{
			"build/go.mod": "module x\n",
		})
		target := filepath.Join(dir, "build", "real-tool")
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "build", "cmd"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(dir, "build", "cmd", "tool")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		mustFailExact(t, dir, "commands.tool.source_dir")
	})
}

func TestSchemaV6RejectsLinkedGoMod(t *testing.T) {
	dir := writeSkill(t, `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{`+validBuildCommand+`}}`, map[string]string{
		"build/cmd/tool/main.go": "package main\n",
	})
	target := filepath.Join(t.TempDir(), "go.mod")
	if err := os.WriteFile(target, []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "build", "go.mod")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	mustFailExact(t, dir, "commands.tool.source_dir")
}

func TestSchemaV6BuildDiagnosticsAreStable(t *testing.T) {
	manifest := `{"schema_version":6,"capabilities":{},"build_roots":["build"],"commands":{"z":{"type":"build","driver":"bad","source_dir":"build"},"a":{"type":"build","driver":"bad","source_dir":"build"}}}`
	for range 25 {
		dir := writeSkill(t, manifest, map[string]string{"build/go.mod": "module x\n"})
		_, err := Load(dir)
		var validation *verr.Error
		if !errors.As(err, &validation) || validation.Path != "commands.a.driver" {
			t.Fatalf("unstable validation error: %v", err)
		}
	}
}
