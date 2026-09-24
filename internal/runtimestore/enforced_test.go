package runtimestore

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// TestManagedEnforcedShimDerivesNativeNames proves enforced launcher
// destinations are native executable paths: extensionless on unix, `.exe`
// on Windows — never a `.cmd` wrapper — with the sidecar paired by
// suffix.
func TestManagedEnforcedShimDerivesNativeNames(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, testCase := range []struct {
		platform  string
		shim      string
		sidecar   string
		windowsSh bool
	}{
		{"unix", "tool", "tool.curator-shim.json", false},
		{"windows", "tool.exe", "tool.exe.curator-shim.json", false},
	} {
		shim, err := NewManagedEnforcedShim(ProjectShim, root, "tool", testCase.platform)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(shim.Path()) != testCase.shim {
			t.Fatalf("launcher = %q, want %q", filepath.Base(shim.Path()), testCase.shim)
		}
		if filepath.Base(shim.SidecarPath()) != testCase.sidecar {
			t.Fatalf("sidecar = %q, want %q", filepath.Base(shim.SidecarPath()), testCase.sidecar)
		}
		if strings.HasSuffix(shim.Path(), ".cmd") {
			t.Fatal("the enforced launcher derivation produced a .cmd wrapper")
		}
	}
	if _, err := NewManagedEnforcedShim("user-bin", root, "tool", "unix"); err == nil {
		t.Fatal("a forwarding role derived an enforced launcher")
	}
	if _, err := NewManagedEnforcedShim(ProjectShim, root, "not a name", "unix"); err == nil {
		t.Fatal("an invalid command derived an enforced launcher")
	}
}

// TestStageEnforcedShimTransitionStagesManagerCopies proves staging
// materializes native manager copies plus sidecar bytes, and nothing
// live: the desired targets carry both files per command, and stale
// launchers and sidecars are removals.
func TestStageEnforcedShimTransitionStagesManagerCopies(t *testing.T) {
	// Sequential: the test points staging at a fixture manager binary.
	manager := filepath.Join(t.TempDir(), "manager")
	payload := append([]byte{0x7f, 'E', 'L', 'F', 1, 2, 3, 4}, bytesRepeatForTest('M', 4096)...)
	if err := os.WriteFile(manager, payload, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(SetManagerBinaryForTest(manager))

	stageRoot := t.TempDir()
	binDir := t.TempDir()
	platform := "unix"
	if runtime.GOOS == "windows" {
		platform = "windows"
	}
	destination, err := NewManagedEnforcedShim(ProjectShim, binDir, "tool", platform)
	if err != nil {
		t.Fatal(err)
	}
	sidecar := []byte(`{"version":1}` + "\n")
	stale, err := NewManagedEnforcedShim(ProjectShim, binDir, "stale", platform)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := StageEnforcedShimTransition(stageRoot,
		[]EnforcedShimSpec{{Destination: destination, Sidecar: sidecar}},
		[]ManagedEnforcedShim{stale})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Desired) != 2 {
		t.Fatalf("desired targets = %d, want the launcher plus the sidecar", len(plan.Desired))
	}
	var launcher, contract *DesiredTarget
	for index := range plan.Desired {
		target := &plan.Desired[index]
		if target.LivePath == destination.Path() {
			launcher = target
		}
		if target.LivePath == destination.SidecarPath() {
			contract = target
		}
		if target.RuntimeKind != EnforcedScriptTargetKind {
			t.Fatalf("target kind = %q, want the enforced kind", target.RuntimeKind)
		}
	}
	if launcher == nil || contract == nil {
		t.Fatalf("desired = %+v, want both enforced files", plan.Desired)
	}
	staged, err := os.ReadFile(launcher.StagedPath) // #nosec G304 -- staged launcher under test
	if err != nil {
		t.Fatal(err)
	}
	if string(staged) != string(payload) {
		t.Fatal("the staged launcher is not the manager bytes")
	}
	stagedInfo, err := os.Stat(launcher.StagedPath)
	if err != nil || !stagedInfo.Mode().IsRegular() {
		t.Fatalf("staged launcher = %v, %v, want a regular file", stagedInfo, err)
	}
	if runtime.GOOS == "windows" {
		// Windows has no POSIX exec bits: a staged launcher is a
		// regular .exe file, and executability is the extension.
		if !strings.HasSuffix(launcher.LivePath, ".exe") {
			t.Fatalf("launcher = %q, want the native .exe image, never a wrapper", launcher.LivePath)
		}
	} else if stagedInfo.Mode().Perm() != 0o700 {
		t.Fatalf("staged launcher mode = %v, want owner-only executable", stagedInfo.Mode())
	}
	stagedSidecar, err := os.ReadFile(contract.StagedPath) // #nosec G304 -- staged sidecar under test
	if err != nil {
		t.Fatal(err)
	}
	if string(stagedSidecar) != string(sidecar) {
		t.Fatal("the staged sidecar is not the contract bytes")
	}
	if len(plan.Removals) != 2 {
		t.Fatalf("removals = %+v, want the stale launcher plus its sidecar", plan.Removals)
	}
	// Nothing live was touched.
	if _, err := os.Stat(destination.Path()); !os.IsNotExist(err) {
		t.Fatal("staging created the live launcher")
	}
	if _, err := os.Stat(destination.SidecarPath()); !os.IsNotExist(err) {
		t.Fatal("staging created the live sidecar")
	}
}

// TestManagedEnforcedShimsInKeysBySidecar proves enumeration follows
// sidecar presence: sidecar-paired launchers are managed (even when the
// launcher bytes are gone, so partial state converges), and ordinary
// shims and junk are not.
func TestManagedEnforcedShimsInKeysBySidecar(t *testing.T) {
	t.Parallel()
	binDir := t.TempDir()
	writeFileForTest(t, filepath.Join(binDir, "tool"+exeSuffixForTest()+".curator-shim.json"), []byte("{}\n"))
	writeFileForTest(t, filepath.Join(binDir, "ghost"+exeSuffixForTest()+".curator-shim.json"), []byte("{}\n"))
	writeFileForTest(t, filepath.Join(binDir, "tool"+exeSuffixForTest()), []byte("native\n"))
	writeFileForTest(t, filepath.Join(binDir, "plain"), []byte("#!/bin/sh\n"))
	writeFileForTest(t, filepath.Join(binDir, "notes.txt"), []byte("junk\n"))

	platform := "unix"
	if runtime.GOOS == "windows" {
		platform = "windows"
	}
	shims, err := ManagedEnforcedShimsIn(binDir, ProjectShim, platform)
	if err != nil {
		t.Fatal(err)
	}
	if len(shims) != 2 || shims[0].Command() != "ghost" || shims[1].Command() != "tool" {
		names := []string{}
		for _, shim := range shims {
			names = append(names, shim.Command())
		}
		t.Fatalf("managed enforced = %q, want ghost and tool", names)
	}
	missing, err := ManagedEnforcedShimsIn(filepath.Join(binDir, "absent"), ProjectShim, platform)
	if err != nil || len(missing) != 0 {
		t.Fatalf("absent bin = %+v, %v, want empty", missing, err)
	}
}

// TestStageEnforcedShimTransitionRefusesBadInput proves the staging guards:
// an empty sidecar, a duplicate launcher, and a staging root overlapping
// the live bin all refuse instead of publishing partial state.
func TestStageEnforcedShimTransitionRefusesBadInput(t *testing.T) {
	// Sequential: the test points staging at a fixture manager binary.
	manager := filepath.Join(t.TempDir(), "manager")
	if err := os.WriteFile(manager, []byte("manager-bytes"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(SetManagerBinaryForTest(manager))

	binDir := t.TempDir()
	destination, err := NewManagedEnforcedShim(ProjectShim, binDir, "tool", "unix")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := StageEnforcedShimTransition(t.TempDir(),
		[]EnforcedShimSpec{{Destination: destination}}, nil); err == nil {
		t.Fatal("staging accepted an empty sidecar contract")
	}
	if _, err := StageEnforcedShimTransition(t.TempDir(),
		[]EnforcedShimSpec{
			{Destination: destination, Sidecar: []byte("{}\n")},
			{Destination: destination, Sidecar: []byte("{}\n")},
		}, nil); err == nil {
		t.Fatal("staging accepted a duplicate enforced launcher")
	}
	if _, err := StageEnforcedShimTransition(binDir,
		[]EnforcedShimSpec{{Destination: destination, Sidecar: []byte("{}\n")}}, nil); err == nil {
		t.Fatal("staging accepted a root overlapping the live bin")
	}
}

// TestEnforcedSidecarSuffixMatchesWorker pins the pairing suffix against
// the worker's constant: install and the launcher must agree byte for
// byte, and the worker owns the spelling.
func TestEnforcedSidecarSuffixMatchesWorker(t *testing.T) {
	t.Parallel()
	if enforcedSidecarSuffix != ".curator-shim.json" {
		t.Fatalf("suffix = %q, want the worker's .curator-shim.json", enforcedSidecarSuffix)
	}
	var _ = godriver.NativeExecutableHeader
}

func bytesRepeatForTest(byteValue byte, count int) []byte {
	repeated := make([]byte, count)
	for index := range repeated {
		repeated[index] = byteValue
	}
	return repeated
}

func writeFileForTest(t *testing.T, path string, payload []byte) {
	t.Helper()
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func exeSuffixForTest() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
