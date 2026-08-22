package main

import (
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// TestInstallPrintsTheRemedyAVersionManagerSelectionEarns drives the whole
// chain an operator with a version-manager toolchain takes: the driver refuses
// the selection at its trust boundary, the install boundary renders that
// refusal through its bounded, path-redacting rendering, and the command prints
// it. A remedy is only worth attaching if it survives all three, so the
// assertion is on what stderr actually says.
func TestInstallPrintsTheRemedyAVersionManagerSelectionEarns(t *testing.T) {
	compiledProject(t)

	realRoot := build.Default.GOROOT
	if realRoot == "" {
		t.Skip("this host reports no GOROOT to select a real launcher from")
	}
	launcher := "go"
	if runtime.GOOS == "windows" {
		launcher = "go.exe"
	}
	// A version-manager layout in the small: the selected launcher sits in a
	// bin directory of its own and resolves to the real launcher under another
	// root, which is what a goenv, asdf, or mise wrapper looks like here.
	wrapperRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wrapperRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(wrapperRoot, "bin", launcher)
	if err := os.Symlink(filepath.Join(realRoot, "bin", launcher), wrapper); err != nil {
		t.Skipf("this host cannot create the symbolic link the case needs: %v", err)
	}
	t.Setenv(godriver.SelectionCuratorGo, wrapper)

	code, _, stderr := capture(t, "install", "app")
	if code != exitFail {
		t.Fatalf("install with a version-manager selection = %d, want %d\n%s", code, exitFail, stderr)
	}

	// The protocol string and the remedy reach the operator on one line, the
	// remedy behind the string a reader matches on.
	const want = "go-v1 toolchain_executable_mismatch: " +
		"selected Go executable is not the regular executable under the derived GOROOT" +
		`; put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`
	if !strings.Contains(stderr, want) {
		t.Fatalf("install did not print the remedied diagnostic:\nwant to contain: %s\ngot:\n%s", want, stderr)
	}
	// The remedy is advice about the operator's own shell, so it must not cost
	// the operator the rule that says the manager never reads that shell to
	// choose a toolchain.
	if !strings.Contains(stderr, "Curator never searches PATH and never downloads a toolchain") {
		t.Fatalf("the remedy displaced the closed selection rule:\n%s", stderr)
	}
}
