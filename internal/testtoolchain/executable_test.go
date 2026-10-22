package testtoolchain

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Explicit Go selection wins over GOROOT, GOROOT wins over the build default,
// and missing/invalid selections never fall back to an ambient go executable.
func TestGoExecutableSelection(t *testing.T) {
	root := t.TempDir()
	makeGo := func(name string) string {
		t.Helper()
		path := filepath.Join(root, name, "bin", executableName("go"))
		writeExecutable(t, path)
		return path
	}
	explicit, selected, fallback := makeGo("explicit"), makeGo("selected"), makeGo("fallback")
	ambient := filepath.Dir(makeGo("ambient"))
	t.Setenv("PATH", ambient)
	for _, tc := range []struct {
		name, configured, goroot, defaultRoot, want string
	}{
		{"explicit", explicit, filepath.Dir(filepath.Dir(selected)), filepath.Dir(filepath.Dir(fallback)), explicit},
		{"GOROOT", "", filepath.Dir(filepath.Dir(selected)), filepath.Dir(filepath.Dir(fallback)), selected},
		{"build-default", "", "", filepath.Dir(filepath.Dir(fallback)), fallback},
		{"absent", "", "", "", ""},
		{"missing-explicit", filepath.Join(root, "missing"), filepath.Dir(filepath.Dir(selected)), "", ""},
		{"relative-explicit", "go", filepath.Dir(filepath.Dir(selected)), "", ""},
		{"directory", root, "", "", ""},
		{"missing-root", "", filepath.Join(root, "missing"), filepath.Dir(filepath.Dir(fallback)), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(name string) string {
				switch name {
				case "CURATOR_GO":
					return tc.configured
				case "GOROOT":
					return tc.goroot
				default:
					return os.Getenv(name)
				}
			}
			got, err := goExecutable(getenv, tc.defaultRoot)
			if got != tc.want || (err == nil) != (tc.want != "") {
				t.Fatalf("selection = %q, %v; want %q (ambient go must not supply absence)", got, err, tc.want)
			}
		})
	}
}

// Pinned Rust selection uses only the named installation; an ambient rustc
// cannot fill in a missing pin or an unsupported target.
func TestRustcExecutableSelection(t *testing.T) {
	home := t.TempDir()
	name := executableName("rustc")
	ambient := filepath.Join(t.TempDir(), name)
	writeExecutable(t, ambient)
	t.Setenv("PATH", filepath.Dir(ambient))
	for _, tc := range []struct{ goos, goarch, target, file string }{
		{"darwin", "arm64", "aarch64-apple-darwin", "rustc"},
		{"linux", "amd64", "x86_64-unknown-linux-gnu", "rustc"},
		{"windows", "amd64", "x86_64-pc-windows-msvc", "rustc.exe"},
	} {
		t.Run(tc.goos, func(t *testing.T) {
			want := filepath.Join(home, ".rustup", "toolchains", "fixture-"+tc.target, "bin", tc.file)
			if got, err := rustcExecutable(home, "fixture", tc.goos, tc.goarch); err == nil || got != "" {
				t.Fatalf("absent pin = %q, %v; ambient rustc must not supply it", got, err)
			}
			writeExecutable(t, want)
			if got, err := rustcExecutable(home, "fixture", tc.goos, tc.goarch); err != nil || got != want {
				t.Fatalf("present pin = %q, %v; want %q", got, err, want)
			}
		})
	}
	if got, err := rustcExecutable(home, "fixture", "unsupported", runtime.GOARCH); err == nil || got != "" {
		t.Fatalf("unsupported target = %q, %v", got, err)
	}
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("test executable\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}
