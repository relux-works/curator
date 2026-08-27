package godriver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// The two protocol strings a toolchain_executable_mismatch reports and the
// remedy that now rides behind them, spelled out literally rather than
// composed from the constants the driver renders them from. Changing either
// half has to be done twice, on purpose: the protocol detail is what a reader
// matches on and may not drift, and the remedy is what an operator acts on.
const (
	wantSelectionMismatchDetail = "selected Go executable is not the regular executable under the derived GOROOT"
	wantProbeMismatchDetail     = "go env GOROOT does not match the selected toolchain"
	wantToolchainRemedy         = `put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"`
)

// operatorDiagnosticBound mirrors internal/install.maxDiagnosticRunes, which
// bounds one rendered diagnostic where an operator reads it. A remedy is only
// a remedy if it survives that rendering whole.
const operatorDiagnosticBound = 240

// TestToolchainExecutableMismatchKeepsItsProtocolStringAndCarriesTheRemedy
// pins both halves of the diagnostic a version-manager selection produces: the
// protocol detail the boundary has always reported, byte for byte, and the
// operator remedy behind it.
//
// Both sites that raise the code are covered, because they are the two ways
// the same wrapper is seen — a launcher that resolves outside the derived
// GOROOT/bin, and a launcher that answers go env with a different root than
// the one selected — and an operator who reached either has the same thing to
// fix.
func TestToolchainExecutableMismatchKeepsItsProtocolStringAndCarriesTheRemedy(t *testing.T) {
	// A version-manager layout in the small: the selected launcher lives under
	// one root and resolves to the real launcher under another, which is what
	// a goenv, asdf, or mise tree looks like to the boundary.
	t.Run("the selection resolves outside the derived GOROOT", func(t *testing.T) {
		inside, outside := t.TempDir(), testToolchain(t)
		if err := os.MkdirAll(filepath.Join(inside, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(outside, "bin", platformGoName), filepath.Join(inside, "bin", platformGoName)); err != nil {
			t.Skipf("this host cannot create the symbolic link the case needs: %v", err)
		}

		_, _, _, err := selectToolchain(Config{CuratorGo: filepath.Join(inside, "bin", platformGoName)}, outside)
		assertRemediedMismatch(t, err, wantSelectionMismatchDetail)
	})

	t.Run("the selected launcher answers go env with another root", func(t *testing.T) {
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
		assertRemediedMismatch(t, err, wantProbeMismatchDetail)
	})
}

// assertRemediedMismatch reads the boundary verdict out of err — rather than
// err.Error(), so the assertion stays about what the boundary reported and not
// about whatever cleanup error a caller may have joined to it on the way out —
// and holds it to the whole contract: the code, the untouched protocol detail,
// the remedy, the order the two are rendered in, and a rendering an operator
// still reads whole.
func assertRemediedMismatch(t *testing.T, err error, wantDetail string) {
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
	if failure.Remedy != wantToolchainRemedy {
		t.Fatalf("remedy = %q, want %q", failure.Remedy, wantToolchainRemedy)
	}
	want := "go-v1 toolchain_executable_mismatch: " + wantDetail + "; " + wantToolchainRemedy
	if failure.Error() != want {
		t.Fatalf("rendered diagnostic = %q, want %q", failure.Error(), want)
	}
	if runes := len([]rune(want)); runes > operatorDiagnosticBound {
		t.Fatalf("rendered diagnostic is %d runes, more than the %d an operator boundary keeps", runes, operatorDiagnosticBound)
	}
}

// TestDiagnosticRenderingLeavesRemedylessBoundariesAlone proves the remedy is
// additive: every boundary that has none renders exactly what it rendered
// before, so only the two sites above changed shape.
func TestDiagnosticRenderingLeavesRemedylessBoundariesAlone(t *testing.T) {
	for name, testCase := range map[string]struct {
		failure *Diagnostic
		want    string
	}{
		"code and detail": {
			failure: &Diagnostic{Code: "toolchain_mutated", Detail: "trusted toolchain changed"},
			want:    "go-v1 toolchain_mutated: trusted toolchain changed",
		},
		"code alone": {
			failure: &Diagnostic{Code: "toolchain_mutated"},
			want:    "go-v1 toolchain_mutated",
		},
		"a remedy without a detail still follows the code": {
			failure: &Diagnostic{Code: "toolchain_executable_mismatch", Remedy: wantToolchainRemedy},
			want:    "go-v1 toolchain_executable_mismatch; " + wantToolchainRemedy,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := testCase.failure.Error(); got != testCase.want {
				t.Fatalf("rendered diagnostic = %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestDiagnosticRemedyIsCarriedThroughAWrappedCause proves the remedy survives
// the way the boundary errors actually travel: wrapped by a caller and read
// back with errors.As, which is how the install boundary renders them.
func TestDiagnosticRemedyIsCarriedThroughAWrappedCause(t *testing.T) {
	cause := errors.New("probe read failed")
	err := diagnosticErrRemedy("toolchain_executable_mismatch", wantToolchainRemedy, cause,
		"go env GOROOT does not match the selected toolchain")

	if !errors.Is(err, cause) {
		t.Fatalf("diagnostic %v does not unwrap to its cause", err)
	}
	var failure *Diagnostic
	if !errors.As(errors.Join(errors.New("staging failed"), err), &failure) {
		t.Fatal("a joined diagnostic is no longer readable")
	}
	if failure.Remedy != wantToolchainRemedy {
		t.Fatalf("remedy = %q, want %q", failure.Remedy, wantToolchainRemedy)
	}
}
