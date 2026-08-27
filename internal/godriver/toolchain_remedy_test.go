package godriver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// The exact operator-facing halves of both toolchain_executable_mismatch sites.
// They are spelled out here rather than composed from the constants the driver
// renders them from, so that changing either half of either string has to be
// done twice, deliberately: the protocol detail is what a reader matches on and
// may not drift, and the remedy is what an operator acts on.
const (
	mismatchSelectionDetail = "selected Go executable is not the regular executable under the derived GOROOT"
	mismatchProbeDetail     = "go env GOROOT does not match the selected toolchain"
	mismatchRemedy          = `put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`
)

// TestToolchainExecutableMismatchCarriesTheOperatorRemedy pins both halves of
// the diagnostic a version-manager selection produces: the protocol detail the
// boundary has always reported, byte for byte and with the remedy kept out of
// it, and the remedy that names the one host fact an operator can act on.
//
// Both sites that raise the code are covered, because they are the two ways the
// same wrapper is seen — a launcher that resolves outside the derived
// GOROOT/bin, and a launcher that answers go env with a different root than the
// one selected — and an operator who reached either has the same thing to fix.
func TestToolchainExecutableMismatchCarriesTheOperatorRemedy(t *testing.T) {
	// A version-manager layout in the small: the selected launcher sits under
	// one root and resolves to the real launcher under another, which is what a
	// goenv, asdf, or mise tree looks like to the boundary.
	t.Run("selection resolves outside the derived GOROOT", func(t *testing.T) {
		inside, outside := t.TempDir(), testToolchain(t)
		if err := os.MkdirAll(filepath.Join(inside, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(outside, "bin", platformGoName), filepath.Join(inside, "bin", platformGoName)); err != nil {
			t.Skipf("this host cannot create the symbolic link the case needs: %v", err)
		}

		_, _, _, err := selectToolchain(Config{CuratorGo: filepath.Join(inside, "bin", platformGoName)}, outside)
		assertMismatchDiagnostic(t, err, mismatchSelectionDetail)
	})

	t.Run("go env reports another root", func(t *testing.T) {
		root, otherRoot := testToolchain(t), testToolchain(t)
		host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
		valid := validProbeExecutor(t, root, host)
		executor := &recordingExecutor{run: func(index int, process Process) (Output, error) {
			output, err := valid.run(index, process)
			if err != nil || index != 2 {
				return output, err
			}
			output.Stdout = mutateJSON(t, output.Stdout, "GOROOT", otherRoot)
			return output, nil
		}}

		_, err := establish(context.Background(), Config{
			PrivateBase: t.TempDir(),
			CuratorGo:   filepath.Join(root, "bin", platformGoName),
			Executor:    executor,
		}, host)
		assertMismatchDiagnostic(t, err, mismatchProbeDetail)
	})
}

// assertMismatchDiagnostic reads the boundary verdict out of err and checks all
// three of its published forms: the code a caller branches on, the protocol
// detail a reader matches on, and the one line an operator reads. It reads the
// diagnostic rather than err.Error() so the assertion stays about what the
// boundary reported, not about whatever cleanup error a caller may have joined
// to it on the way out.
func assertMismatchDiagnostic(t *testing.T, err error, wantDetail string) {
	t.Helper()
	var failure *Diagnostic
	if !errors.As(err, &failure) {
		t.Fatalf("error = %v, want a go-v1 diagnostic", err)
	}
	if failure.Code != "toolchain_executable_mismatch" {
		t.Fatalf("code = %q, want %q", failure.Code, "toolchain_executable_mismatch")
	}
	if failure.Detail != wantDetail {
		t.Fatalf("protocol detail = %q, want %q", failure.Detail, wantDetail)
	}
	if failure.Remedy != mismatchRemedy {
		t.Fatalf("remedy = %q, want %q", failure.Remedy, mismatchRemedy)
	}
	want := "go-v1 toolchain_executable_mismatch: " + wantDetail + "; " + mismatchRemedy
	if failure.Error() != want {
		t.Fatalf("diagnostic = %q, want %q", failure.Error(), want)
	}
}

// TestDiagnosticRenderingIsUnchangedWithoutARemedy pins the rest of the
// vocabulary: a boundary that carries no remedy renders exactly what it
// rendered before there was a remedy to carry, with and without a detail.
func TestDiagnosticRenderingIsUnchangedWithoutARemedy(t *testing.T) {
	for name, testCase := range map[string]struct {
		failure Diagnostic
		want    string
	}{
		"code and detail": {
			failure: Diagnostic{Code: "toolchain_mutated", Detail: "toolchain file changed while opening"},
			want:    "go-v1 toolchain_mutated: toolchain file changed while opening",
		},
		"code only": {
			failure: Diagnostic{Code: "toolchain_mutated"},
			want:    "go-v1 toolchain_mutated",
		},
		"remedy without a detail": {
			failure: Diagnostic{Code: "toolchain_executable_mismatch", Remedy: mismatchRemedy},
			want:    "go-v1 toolchain_executable_mismatch; " + mismatchRemedy,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := testCase.failure.Error(); got != testCase.want {
				t.Fatalf("Error() = %q, want %q", got, testCase.want)
			}
		})
	}
}
