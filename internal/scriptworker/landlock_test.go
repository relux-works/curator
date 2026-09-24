package scriptworker

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// The expected masks below are independent UAPI literals from
// include/uapi/linux/landlock.h — never the production constants under
// test, so a wrong constant fails here instead of agreeing with
// itself. The ABI introduction follows the same header: ABI 1 brings
// execute/write/read plus the entry creation/removal rights, ABI 2
// reparenting (refer), ABI 3 truncate, ABI 4 no new filesystem right,
// ABI 5 ioctl-dev.
const (
	uapiExecute    = 1 << 0
	uapiWriteFile  = 1 << 1
	uapiReadFile   = 1 << 2
	uapiReadDir    = 1 << 3
	uapiRemoveDir  = 1 << 4
	uapiRemoveFile = 1 << 5
	uapiMakeChar   = 1 << 6
	uapiMakeDir    = 1 << 7
	uapiMakeReg    = 1 << 8
	uapiMakeSock   = 1 << 9
	uapiMakeFifo   = 1 << 10
	uapiMakeBlock  = 1 << 11
	uapiMakeSym    = 1 << 12
	uapiRefer      = 1 << 13
	uapiTruncate   = 1 << 14
	uapiIoctlDev   = 1 << 15
)

const uapiMutationABI1 = uapiRemoveDir | uapiRemoveFile |
	uapiMakeChar | uapiMakeDir | uapiMakeReg | uapiMakeSock |
	uapiMakeFifo | uapiMakeBlock | uapiMakeSym

const uapiMutationABI2 = uapiMutationABI1 | uapiRefer

// TestLandlockHandledMaskFollowsABI pins the ABI-gated handled mask on
// every platform against the independent UAPI values above: write
// confinement handles write plus the entry creation/removal rights on
// ABI 1, reparenting (refer) on ABI 2, truncation on ABI 3, and device
// ioctl on ABI 5; ABI 4 adds no filesystem right and newer ABIs add
// none this build knows, so the set stays put there; an ABI below 1
// handles nothing; reads are never handled. A mask carrying a right
// the running kernel does not know fails ruleset creation with EINVAL,
// so the enforcement must never request more than the probed ABI
// provides — and every right it provides for the installable controls
// is handled, because Landlock permits unhandled actions.
func TestLandlockHandledMaskFollowsABI(t *testing.T) {
	execute := uint64(uapiExecute)
	write := uint64(uapiWriteFile)
	mutationABI1 := uint64(uapiMutationABI1)
	refer := uint64(uapiRefer)
	truncate := uint64(uapiTruncate)
	ioctl := uint64(uapiIoctlDev)
	writeABI1 := write | mutationABI1
	writeABI2 := write | mutationABI1 | refer
	writeABI3 := write | mutationABI1 | refer | truncate
	writeABI5 := write | mutationABI1 | refer | truncate | ioctl
	for _, testCase := range []struct {
		name             string
		abi              int
		execDenial       bool
		writeConfinement bool
		wantHandled      uint64
		wantWriteAccess  uint64
	}{
		{"no-abi-handles-nothing", 0, true, true, 0, 0},
		{"abi-1-write-and-mutation", 1, true, true, execute | writeABI1, writeABI1},
		{"abi-2-adds-refer", 2, true, true, execute | writeABI2, writeABI2},
		{"abi-3-adds-truncate", 3, true, true, execute | writeABI3, writeABI3},
		{"abi-4-adds-no-filesystem-right", 4, true, true, execute | writeABI3, writeABI3},
		{"abi-5-adds-ioctl-dev", 5, true, true, execute | writeABI5, writeABI5},
		{"abi-6-keeps-the-set", 6, true, true, execute | writeABI5, writeABI5},
		{"abi-9-keeps-the-set", 9, true, true, execute | writeABI5, writeABI5},
		{"exec-only", 4, true, false, execute, writeABI3},
		{"write-only", 4, false, true, writeABI3, writeABI3},
		{"no-control-handles-nothing", 4, false, false, 0, writeABI3},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := landlockHandledForABI(testCase.abi, testCase.execDenial, testCase.writeConfinement); got != testCase.wantHandled {
				t.Fatalf("handled mask = %#x, want %#x", got, testCase.wantHandled)
			}
			if got := landlockHandledForABI(testCase.abi, testCase.execDenial, testCase.writeConfinement); got&(uapiReadFile|uapiReadDir) != 0 {
				t.Fatalf("handled mask = %#x, want no read right handled", got)
			}
			if got := landlockWriteAccessForABI(testCase.abi); got != testCase.wantWriteAccess {
				t.Fatalf("write access = %#x, want %#x", got, testCase.wantWriteAccess)
			}
			if got := landlockFileAccessForABI(testCase.abi); got != testCase.wantWriteAccess&^uint64(uapiMutationABI2) {
				t.Fatalf("file access = %#x, want the write set without directory-only rights", got)
			}
		})
	}
}

// TestLandlockRuleRightsFollowObjectType pins the per-object-type rule
// rights on every platform against the independent UAPI values: a rule
// carries only handled rights, and only the rights that apply to the
// object type — directory grants keep the creation, removal, and
// reparenting rights, while file and device grants shed them (a
// directory-only rule on a non-directory returns EINVAL) and keep the
// file rights, truncation included: it governs truncate(2),
// ftruncate(2), creat(2), and open(2) with O_TRUNC on regular files.
func TestLandlockRuleRightsFollowObjectType(t *testing.T) {
	execute := uint64(uapiExecute)
	write := uint64(uapiWriteFile)
	truncate := uint64(uapiTruncate)
	ioctl := uint64(uapiIoctlDev)
	mutation := uint64(uapiMutationABI2)
	fileRights := write | truncate | ioctl
	full := execute | fileRights | mutation
	for _, testCase := range []struct {
		name    string
		isDir   bool
		access  uint64
		handled uint64
		want    uint64
	}{
		{"directory-keeps-the-full-write-set", true, write | mutation | truncate | ioctl, full, write | mutation | truncate | ioctl},
		{"file-keeps-write-and-truncate", false, write | truncate, full, write | truncate},
		{"file-keeps-ioctl-dev", false, write | ioctl, full, write | ioctl},
		{"file-sheds-directory-mutation", false, write | mutation | truncate, full, write | truncate},
		{"device-sheds-directory-mutation", false, write | mutation | truncate | ioctl, full, fileRights},
		{"interpreter-file-keeps-execute", false, execute, full, execute},
		{"farm-directory-keeps-execute", true, execute, full, execute},
		{"rule-narrows-to-handled", true, full, write | truncate, write | truncate},
		{"rule-without-handled-grants-nothing", true, write, 0, 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := landlockRuleRights(testCase.isDir, testCase.access, testCase.handled); got != testCase.want {
				t.Fatalf("rule rights = %#x, want %#x", got, testCase.want)
			}
		})
	}
}

// TestLandlockConstructionConfirmsTypedRules proves the pre-ready
// Landlock confirmation builds the exact enforcement ruleset — same ABI
// mask, same typed rules over the same grant paths — without
// restricting the caller: unit tests exercise construction only, never
// restriction. A grant set over paths that exist confirms; a grant over
// a path that does not exist fails only its own control (each control
// is confirmed from its own grants, so a missing writable never voids
// the execute denial); and a canary write outside the grants succeeds
// afterwards, proving the test process itself was never restricted.
func TestLandlockConstructionConfirmsTypedRules(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	root := mustPhysical(t, t.TempDir())
	executableDir := filepath.Join(root, "farm")
	if err := os.MkdirAll(executableDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writableDir := filepath.Join(root, "private")
	if err := os.MkdirAll(writableDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writableFile := filepath.Join(writableDir, "derived.txt")
	writeTestFile(t, writableFile, []byte("derived\n"), 0o644)
	grants := landlockGrants{
		executables: []string{executableDir},
		writables:   []string{writableDir, writableFile},
	}
	if execErr, writeErr := confirmScriptLandlock(true, true, grants); execErr != nil || writeErr != nil {
		t.Fatalf("construction over existing grants fails (%v, %v), want (nil, nil)", execErr, writeErr)
	}
	if execErr, writeErr := confirmScriptLandlock(true, false, grants); execErr != nil || writeErr != nil {
		t.Fatalf("exec-only construction fails (%v, %v), want (nil, nil)", execErr, writeErr)
	}
	if execErr, writeErr := confirmScriptLandlock(false, true, grants); execErr != nil || writeErr != nil {
		t.Fatalf("write-only construction fails (%v, %v), want (nil, nil)", execErr, writeErr)
	}
	bogusExec := landlockGrants{
		executables: []string{filepath.Join(root, "no-such-farm")},
		writables:   []string{writableDir},
	}
	if execErr, writeErr := confirmScriptLandlock(true, true, bogusExec); execErr == nil || writeErr != nil {
		t.Fatalf("construction over a missing executable fails (%v, %v), want (err, nil)", execErr, writeErr)
	}
	bogusWrite := landlockGrants{
		executables: []string{executableDir},
		writables:   []string{filepath.Join(root, "no-such-path")},
	}
	if execErr, writeErr := confirmScriptLandlock(true, true, bogusWrite); execErr != nil || writeErr == nil {
		t.Fatalf("construction over a missing writable fails (%v, %v), want (nil, err)", execErr, writeErr)
	}
	canary := filepath.Join(mustPhysical(t, t.TempDir()), "canary.txt")
	if err := os.WriteFile(canary, []byte("unrestricted\n"), 0o644); err != nil {
		t.Fatalf("a canary write outside the grants failed after construction: %v", err)
	}
}

// TestLinuxLandlockProbeFindsPresentControls requires the true Linux
// probe to find both Landlock controls present: the hosted runner
// provides Landlock (kernel 6.8+, ABI 4) and the probe child proves the
// full enforcement sequence — no_new_privs, ruleset, rule,
// self-restriction — on this host. When the probe cannot prove the
// mechanism the controls report absent and sessions proceed
// unconfined, so a silent absent here would disarm every enforcement
// row below it.
func TestLinuxLandlockProbeFindsPresentControls(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	platform, probes, err := probeScriptInventory()
	if err != nil {
		t.Fatalf("the host probe refused: %v", err)
	}
	if platform != ScriptPlatformLinux {
		t.Fatalf("platform = %q, want the Linux inventory", platform)
	}
	for _, control := range []string{ScriptControlDescendantExecDenial, ScriptControlFilesystemWriteConfinement} {
		found := false
		for _, probe := range probes {
			if probe.Name != control {
				continue
			}
			found = true
			if !probe.Present {
				t.Fatalf("control %q probes absent on a Landlock host: the probe child could not enforce the sequence (no_new_privs, ABI ruleset, rule, restrict)", control)
			}
		}
		if !found {
			t.Fatalf("the probe reports no entry for %q", control)
		}
	}
}

// TestLinuxLandlockProbeChildAnswersAtTheBoundary bounds the probe
// child entry at the process boundary: against a directory the caller
// owns the re-executed binary exits 0, and against a path that does not
// exist it exits nonzero. Both branches run in the child process, never
// in the test binary — the success branch restricts its own thread,
// which must never happen to the test process itself.
func TestLinuxLandlockProbeChildAnswersAtTheBoundary(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Landlock rights are a Linux-only control in script-worker-v1-native-control-inventory-v1")
	}
	answer := func(directory string) int {
		t.Helper()
		command := exec.Command(os.Args[0], ScriptLandlockProbeMode, directory) // #nosec G204 -- fixed test-only re-execution in the hidden probe mode
		command.Env = scriptWorkerEnvironment()
		err := command.Run()
		if err == nil {
			return 0
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		t.Fatalf("the probe child did not exit: %v", err)
		return -1
	}
	owned := mustPhysical(t, t.TempDir())
	if got := answer(owned); got != 0 {
		t.Fatalf("the probe child exits %d over an owned directory, want 0", got)
	}
	missing := filepath.Join(mustPhysical(t, t.TempDir()), "no-such-directory")
	if got := answer(missing); got == 0 {
		t.Fatal("the probe child exits 0 over a missing directory, want nonzero")
	}
}
