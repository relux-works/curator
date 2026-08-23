package moduleroots

import (
	"os"
	"path/filepath"
	"testing"
)

// snapshot lays out an immutable raw skill snapshot: every directory listed,
// with go.mod written into every directory named in goMods.
func snapshot(t *testing.T, directories []string, goMods []string) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range directories {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, directory := range goMods {
		path := filepath.Join(root, filepath.FromSlash(directory), GoModName)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("module fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	if code == "" {
		if err != nil {
			t.Fatalf("unexpected failure: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected %s, got no error", code)
	}
	if got := Code(err); got != code {
		t.Fatalf("diagnostic = %q (%v), want %q", got, err, code)
	}
}

func TestValidateDeclarationAdmitsDeclaredDirectories(t *testing.T) {
	root := snapshot(t,
		[]string{"pkg/board", "pkg/remoteconfig", "scripts", "tools/cli"},
		[]string{"pkg/board", "pkg/remoteconfig", "tools/cli"})
	err := ValidateDeclaration(root, "commands.tool.modules",
		[]string{"pkg/board", "pkg/remoteconfig"}, []string{"tools/cli"}, []string{"scripts"})
	wantCode(t, err, "")
}

func TestValidateDeclarationRejections(t *testing.T) {
	cases := []struct {
		name         string
		directories  []string
		goMods       []string
		modules      []string
		buildRoots   []string
		runtimeRoots []string
		code         string
	}{
		{
			name: "dot is not a module directory", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"."}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name: "parent escape", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"../pkg/lib"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name: "absolute path", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"/pkg/lib"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name: "backslash", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{`pkg\lib`}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name: "windows device component", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"pkg/CON"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name:        "duplicate declaration",
			directories: []string{"pkg/lib", "tools/cli"}, goMods: []string{"pkg/lib", "tools/cli"},
			modules: []string{"pkg/lib", "pkg/lib"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name: "directory absent from the snapshot", directories: []string{"tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"pkg/lib"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name:        "directory without go.mod",
			directories: []string{"pkg/lib", "tools/cli"}, goMods: []string{"tools/cli"},
			modules: []string{"pkg/lib"}, buildRoots: []string{"tools/cli"}, code: DeclarationInvalid,
		},
		{
			name:        "nested declared directories",
			directories: []string{"pkg/board", "pkg/board/codec", "tools/cli"},
			goMods:      []string{"pkg/board", "pkg/board/codec", "tools/cli"},
			modules:     []string{"pkg/board", "pkg/board/codec"}, buildRoots: []string{"tools/cli"},
			code: ContainmentInvalid,
		},
		{
			name:        "declared directory inside a build root",
			directories: []string{"tools/cli", "tools/cli/pkg/lib"},
			goMods:      []string{"tools/cli", "tools/cli/pkg/lib"},
			modules:     []string{"tools/cli/pkg/lib"}, buildRoots: []string{"tools/cli"},
			code: ContainmentInvalid,
		},
		{
			name:        "declared directory inside a runtime root",
			directories: []string{"pkg", "pkg/board", "tools/cli"},
			goMods:      []string{"pkg/board", "tools/cli"},
			modules:     []string{"pkg/board"}, buildRoots: []string{"tools/cli"}, runtimeRoots: []string{"pkg"},
			code: ContainmentInvalid,
		},
		{
			name:        "declared directory equal to a build root",
			directories: []string{"pkg/board", "tools/cli"}, goMods: []string{"pkg/board", "tools/cli"},
			modules: []string{"pkg/board"}, buildRoots: []string{"tools/cli", "pkg/board"},
			code: ContainmentInvalid,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			root := snapshot(t, testCase.directories, testCase.goMods)
			err := ValidateDeclaration(root, "commands.tool.modules",
				testCase.modules, testCase.buildRoots, testCase.runtimeRoots)
			wantCode(t, err, testCase.code)
		})
	}
}

// TestValidateDeclarationRejectsPlatformPathCollision proves the disjointness
// comparison also holds under the platform path mapping, so two declarations
// that differ only by case are rejected even where only one of them can exist.
func TestValidateDeclarationRejectsPlatformPathCollision(t *testing.T) {
	root := snapshot(t,
		[]string{"pkg/Board", "pkg/board", "tools/cli"},
		[]string{"pkg/Board", "pkg/board", "tools/cli"})
	err := ValidateDeclaration(root, "commands.tool.modules",
		[]string{"pkg/Board", "pkg/board"}, []string{"tools/cli"}, nil)
	wantCode(t, err, ContainmentInvalid)
}

// TestValidateDeclarationRejectsLinkedDirectory proves a link cannot redirect
// the check outside the snapshot: the component is inspected with Lstat, so a
// symlink to a real module directory is still rejected.
func TestValidateDeclarationRejectsLinkedDirectory(t *testing.T) {
	root := snapshot(t, []string{"pkg/real", "tools/cli"}, []string{"pkg/real", "tools/cli"})
	if err := os.Symlink(filepath.Join(root, "pkg", "real"), filepath.Join(root, "pkg", "linked")); err != nil {
		t.Skipf("this host cannot create the symlink: %v", err)
	}
	err := ValidateDeclaration(root, "commands.tool.modules", []string{"pkg/linked"}, []string{"tools/cli"}, nil)
	wantCode(t, err, DeclarationInvalid)
}

// TestValidateDeclarationRejectsLinkedGoMod proves the go.mod itself must be a
// real regular file, not a link into or out of the snapshot.
func TestValidateDeclarationRejectsLinkedGoMod(t *testing.T) {
	root := snapshot(t, []string{"pkg/lib", "tools/cli"}, []string{"tools/cli"})
	if err := os.Symlink(filepath.Join(root, "tools", "cli", GoModName), filepath.Join(root, "pkg", "lib", GoModName)); err != nil {
		t.Skipf("this host cannot create the symlink: %v", err)
	}
	err := ValidateDeclaration(root, "commands.tool.modules", []string{"pkg/lib"}, []string{"tools/cli"}, nil)
	wantCode(t, err, DeclarationInvalid)
}

func TestEffectiveReplaceSet(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		want       []Directive
		code       string
		wantModule string
	}{
		{
			name: "unversioned directive with its selection annotation",
			content: "# example.com/board => ../../pkg/board\n" +
				"## explicit; go 1.23\n" +
				"example.com/board\n" +
				"# example.com/board v0.0.0 => ../../pkg/board\n",
			want: []Directive{{Module: "example.com/board", Target: "../../pkg/board"}},
		},
		{
			name:    "no replacement annotations at all",
			content: "# example.com/dep v1.0.0\n## explicit; go 1.23\nexample.com/dep\n",
			want:    nil,
		},
		{
			name:    "a line that only mentions the separator is not an annotation",
			content: "example.com/dep => ../../pkg/dep\n#no-space-prefix => ../../pkg/dep\n",
			want:    nil,
		},
		{
			name:    "module-to-module redirect",
			content: "# example.com/board => example.com/fork v1.2.3\n",
			code:    DirectiveFormUnsupported,
		},
		{
			name:    "versioned left side with no matching directive",
			content: "# example.com/board v1.2.3 => ../../pkg/board\n",
			code:    DirectiveFormUnsupported,
		},
		{
			name:    "selection annotation whose right side differs is a versioned directive",
			content: "# example.com/board => ../../pkg/board\n# example.com/board v1.2.3 => ../../pkg/other\n",
			code:    DirectiveFormUnsupported,
		},
		{
			name:    "three tokens on the left",
			content: "# example.com/board v1 v2 => ../../pkg/board\n",
			code:    DirectiveFormUnsupported,
		},
		{
			name:    "empty right side",
			content: "# example.com/board => \n",
			code:    DirectiveFormUnsupported,
		},
		{
			name:    "one module carrying two different replacements",
			content: "# example.com/board => ../../pkg/board\n# example.com/board => ../../pkg/other\n",
			code:    DirectiveFormUnsupported,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := EffectiveReplaceSet([]byte(testCase.content))
			wantCode(t, err, testCase.code)
			if testCase.code != "" {
				return
			}
			if len(got) != len(testCase.want) {
				t.Fatalf("directives = %+v, want %+v", got, testCase.want)
			}
			for index := range got {
				if got[index] != testCase.want[index] {
					t.Fatalf("directive[%d] = %+v, want %+v", index, got[index], testCase.want[index])
				}
			}
		})
	}
}

// TestEffectiveReplaceSetIgnoresCarriageReturns proves a modules.txt written
// with CRLF endings is read the same way as one written with LF endings.
func TestEffectiveReplaceSetIgnoresCarriageReturns(t *testing.T) {
	directives, err := EffectiveReplaceSet([]byte("# example.com/board => ../../pkg/board\r\n"))
	wantCode(t, err, "")
	if len(directives) != 1 || directives[0].Target != "../../pkg/board" {
		t.Fatalf("directives = %+v", directives)
	}
}

func TestValidateBijection(t *testing.T) {
	cases := []struct {
		name       string
		buildRoot  string
		modules    []string
		directives []Directive
		code       string
	}{
		{
			name: "declared and effective agree", buildRoot: "tools/cli",
			modules: []string{"pkg/board", "pkg/remoteconfig"},
			directives: []Directive{
				{Module: "example.com/board", Target: "../../pkg/board"},
				{Module: "example.com/remoteconfig", Target: "../../pkg/remoteconfig"},
			},
		},
		{
			name: "no declaration and no replacement", buildRoot: "tools/cli",
		},
		{
			name: "replacement target escapes the snapshot", buildRoot: "tools/cli",
			directives: []Directive{{Module: "example.com/escape", Target: "../../../outside"}},
			code:       DirectiveUndeclared,
		},
		{
			name: "undeclared directory replacement", buildRoot: "tools/cli",
			directives: []Directive{{Module: "example.com/extra", Target: "../../pkg/extra"}},
			code:       DirectiveUndeclared,
		},
		{
			name: "declared module without replacement", buildRoot: "tools/cli",
			modules: []string{"pkg/board"}, code: DeclarationUnused,
		},
		{
			name: "two replacements naming one declaration", buildRoot: "tools/cli",
			modules: []string{"pkg/board"},
			directives: []Directive{
				{Module: "example.com/board", Target: "../../pkg/board"},
				{Module: "example.com/alias", Target: "../../pkg/board"},
			},
			code: DirectiveUndeclared,
		},
		{
			name: "absolute replacement target", buildRoot: "tools/cli",
			modules:    []string{"pkg/board"},
			directives: []Directive{{Module: "example.com/board", Target: "/pkg/board"}},
			code:       DirectiveFormUnsupported,
		},
		{
			name: "windows drive replacement target", buildRoot: "tools/cli",
			modules:    []string{"pkg/board"},
			directives: []Directive{{Module: "example.com/board", Target: `C:\pkg\board`}},
			code:       DirectiveFormUnsupported,
		},
		{
			name: "replacement resolving onto the build root", buildRoot: "tools/cli",
			directives: []Directive{{Module: "example.com/self", Target: "."}},
			code:       DirectiveUndeclared,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			wantCode(t, ValidateBijection(testCase.buildRoot, testCase.modules, testCase.directives), testCase.code)
		})
	}
}
