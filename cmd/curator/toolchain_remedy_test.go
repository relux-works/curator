package main

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// TestInstallPrintsTheToolchainRemedyAVersionManagerSelectionEarns drives the
// whole chain an operator with a version-manager toolchain takes: the driver
// refuses the selection at its trust boundary, the install boundary renders
// that refusal, and the CLI prints it. The remedy is only worth attaching if it
// survives all three, so the assertion is on what stderr actually says.
func TestInstallPrintsTheToolchainRemedyAVersionManagerSelectionEarns(t *testing.T) {
	compiledProject(t)

	// A version-manager layout in the small: the selected launcher sits in a
	// bin directory of its own and resolves to the real launcher under another
	// root, which is what a goenv, asdf, or mise wrapper looks like to the
	// driver.
	real := build.Default.GOROOT
	if real == "" {
		t.Skip("this host does not report a GOROOT to link against")
	}
	wrapperRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wrapperRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(wrapperRoot, "bin", filepath.Base(filepath.Join(real, "bin", "go")))
	if err := os.Symlink(filepath.Join(real, "bin", filepath.Base(wrapper)), wrapper); err != nil {
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
	// the operator the rule that says the manager does not read that shell.
	if !strings.Contains(stderr, "Curator never searches PATH and never downloads a toolchain") {
		t.Fatalf("the remedy displaced the closed selection rule:\n%s", stderr)
	}
}
