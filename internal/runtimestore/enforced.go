package runtimestore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/stateread"
)

// Enforced launcher staging for `script-worker-v1` commands (Protocol Core
// §4.1.1: "an enforced command's shim MUST NOT be a symlink, a
// POSIX-shell wrapper, a `.cmd` wrapper, or any other program that resolves
// the command through a shell or through the inherited `PATH`").
//
// An enforced launcher is a native copy of the installed manager
// executable plus a manager-published sidecar contract the launcher reads
// at invocation. The copy replays the manager role — derive the profile,
// re-execute itself in the fixed hidden worker mode — so no shell, `.cmd`,
// or symlink stands between the manager and the interpreter, and every
// path the launcher follows is fixed install state rather than PATH or
// shell resolution.

// EnforcedScriptTargetKind names staged enforced launchers and sidecars in
// the transaction plan.
const EnforcedScriptTargetKind = TargetKind("enforced-script")

// enforcedSidecarSuffix pairs a launcher with its sidecar. It matches the
// worker's suffix byte for byte; the worker owns the constant, and this
// package's tests pin the equality.
const enforcedSidecarSuffix = ".curator-shim.json"

// ManagedEnforcedShim is a typed manager-owned enforced launcher
// destination: the native executable plus its sidecar contract.
type ManagedEnforcedShim struct {
	role     ShimRole
	binDir   string
	command  string
	platform string
	path     string
	sidecar  string
}

// NewManagedEnforcedShim derives one manager-owned enforced launcher
// destination from a validated bin directory, command name, and platform.
// The launcher is extensionless on unix and `.exe` on Windows — never a
// `.cmd` wrapper, which the policy forbids.
func NewManagedEnforcedShim(role ShimRole, binDir, command, platform string) (ManagedEnforcedShim, error) {
	if role != ProjectShim && role != GlobalCanonicalShim {
		return ManagedEnforcedShim{}, fmt.Errorf("unsupported enforced shim role %q", role)
	}
	if !identifiers.Valid(command) {
		return ManagedEnforcedShim{}, fmt.Errorf("enforced shim command is not a portable identifier")
	}
	if platform != "unix" && platform != "windows" {
		return ManagedEnforcedShim{}, fmt.Errorf("unsupported shim platform %q", platform)
	}
	abs, err := cleanAbsolute(binDir)
	if err != nil {
		return ManagedEnforcedShim{}, fmt.Errorf("enforced shim bin: %w", err)
	}
	name := command
	if platform == "windows" {
		name += ".exe"
	}
	path := filepath.Join(abs, name)
	return ManagedEnforcedShim{
		role: role, binDir: abs, command: command, platform: platform,
		path: path, sidecar: path + enforcedSidecarSuffix,
	}, nil
}

// Role reports the directory class this launcher belongs to.
func (shim ManagedEnforcedShim) Role() ShimRole { return shim.role }

// BinDir is the validated directory that holds the launcher.
func (shim ManagedEnforcedShim) BinDir() string { return shim.binDir }

// Command is the portable command identifier the launcher publishes.
func (shim ManagedEnforcedShim) Command() string { return shim.command }

// Platform is the platform the launcher was derived for.
func (shim ManagedEnforcedShim) Platform() string { return shim.platform }

// Path is the launcher's own absolute path, including any platform suffix.
func (shim ManagedEnforcedShim) Path() string { return shim.path }

// SidecarPath is the absolute path of the launcher's paired sidecar.
func (shim ManagedEnforcedShim) SidecarPath() string { return shim.sidecar }

// ManagedEnforcedShimsIn enumerates the enforced launchers a manager-owned
// bin directory currently holds, keyed by sidecar presence: every
// `*.curator-shim.json` regular file with a portable command stem claims
// its paired launcher, whether or not the launcher bytes are still there,
// so partial state still converges.
func ManagedEnforcedShimsIn(binDir string, role ShimRole, platform string) ([]ManagedEnforcedShim, error) {
	listing, err := stateread.ReadDir(binDir)
	if err != nil {
		return nil, err
	}
	if listing.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if listing.Kind != stateread.KindPresent {
		return nil, stateread.UnusableError(binDir, fmt.Errorf("unknown directory read state %q", listing.Kind))
	}
	entries := listing.Entries
	var shims []ManagedEnforcedShim
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, enforcedSidecarSuffix) {
			continue
		}
		stem := strings.TrimSuffix(name, enforcedSidecarSuffix)
		if platform == "windows" {
			stem = strings.TrimSuffix(stem, ".exe")
		}
		if !identifiers.Valid(stem) {
			continue
		}
		shim, err := NewManagedEnforcedShim(role, binDir, stem, platform)
		if err != nil {
			continue
		}
		shims = append(shims, shim)
	}
	sort.Slice(shims, func(i, j int) bool { return enforcedSortKey(shims[i]) < enforcedSortKey(shims[j]) })
	return shims, nil
}

// EnforcedShimSpec describes one desired enforced launcher and its
// validated sidecar bytes.
type EnforcedShimSpec struct {
	Destination ManagedEnforcedShim
	// Sidecar is the validated marshalled sidecar contract. The caller
	// builds it through the worker's constructor; this package stages the
	// bytes without interpreting them.
	Sidecar []byte
}

// testManagerBinary overrides the manager executable enforced launchers
// are copied from. Production always copies the running manager; tests
// point it at a built production binary. No production code path sets it.
var testManagerBinary = ""

// SetManagerBinaryForTest points enforced launcher staging at path and
// returns a restore function. Tests that set it must be sequential and
// must defer the restore call.
func SetManagerBinaryForTest(path string) (restore func()) {
	testManagerBinary = path
	return func() { testManagerBinary = "" }
}

// managerBinarySource resolves the installed manager executable one
// enforced launcher is copied from.
func managerBinarySource() (string, error) {
	if testManagerBinary != "" {
		return testManagerBinary, nil
	}
	return os.Executable()
}

// StageEnforcedShimTransition materializes only operation-private enforced
// launcher files: a native manager copy plus the sidecar contract per
// desired command. Live paths are returned as transaction targets and are
// never created, replaced, or removed here.
func StageEnforcedShimTransition(stageRoot string, desired []EnforcedShimSpec, currentlyManaged []ManagedEnforcedShim) (TransitionPlan, error) {
	stageRoot, err := cleanAbsolute(stageRoot)
	if err != nil {
		return TransitionPlan{}, fmt.Errorf("enforced shim staging root: %w", err)
	}
	desired = append([]EnforcedShimSpec(nil), desired...)
	currentlyManaged = append([]ManagedEnforcedShim(nil), currentlyManaged...)
	sort.Slice(desired, func(i, j int) bool {
		return enforcedSortKey(desired[i].Destination) < enforcedSortKey(desired[j].Destination)
	})
	manager, err := managerBinarySource()
	if err != nil {
		return TransitionPlan{}, fmt.Errorf("enforced shim manager source: %w", err)
	}
	managerPayload, err := readManagerBinary(manager)
	if err != nil {
		return TransitionPlan{}, fmt.Errorf("enforced shim manager source: %w", err)
	}
	plan := TransitionPlan{}
	desiredPaths := make(map[string]bool, len(desired))
	for _, spec := range desired {
		shim := spec.Destination
		if err := validateManagedEnforcedShim(shim); err != nil {
			return TransitionPlan{}, err
		}
		if pathsOverlap(stageRoot, shim.binDir) {
			return TransitionPlan{}, fmt.Errorf("enforced shim staging root overlaps live managed bin %s", shim.binDir)
		}
		if len(spec.Sidecar) == 0 {
			return TransitionPlan{}, fmt.Errorf("enforced shim %s has no sidecar contract", shim.path)
		}
		key := platformPathKey(shim.path, shim.platform)
		if desiredPaths[key] {
			return TransitionPlan{}, fmt.Errorf("duplicate desired enforced shim %s", shim.path)
		}
		desiredPaths[key] = true
		stagedLauncher := stagedEnforcedShimPath(stageRoot, shim, filepath.Base(shim.path))
		if err := writeStagedManagerCopy(stagedLauncher, managerPayload); err != nil {
			return TransitionPlan{}, fmt.Errorf("stage enforced shim %s: %w", shim.path, err)
		}
		stagedSidecar := stagedEnforcedShimPath(stageRoot, shim, filepath.Base(shim.sidecar))
		if err := writeStagedSidecar(stagedSidecar, spec.Sidecar); err != nil {
			return TransitionPlan{}, fmt.Errorf("stage enforced shim %s: %w", shim.path, err)
		}
		plan.Desired = append(plan.Desired,
			DesiredTarget{
				Kind: ShimTarget, Role: shim.role, Command: shim.command,
				RuntimeKind: EnforcedScriptTargetKind, LivePath: shim.path, StagedPath: stagedLauncher,
			},
			DesiredTarget{
				Kind: ShimTarget, Role: shim.role, Command: shim.command,
				RuntimeKind: EnforcedScriptTargetKind, LivePath: shim.sidecar, StagedPath: stagedSidecar,
			},
		)
	}

	sort.Slice(currentlyManaged, func(i, j int) bool {
		return enforcedSortKey(currentlyManaged[i]) < enforcedSortKey(currentlyManaged[j])
	})
	seenCurrent := map[string]bool{}
	for _, shim := range currentlyManaged {
		if err := validateManagedEnforcedShim(shim); err != nil {
			return TransitionPlan{}, err
		}
		key := platformPathKey(shim.path, shim.platform)
		if seenCurrent[key] {
			continue
		}
		seenCurrent[key] = true
		if !desiredPaths[key] {
			plan.Removals = append(plan.Removals,
				RemovalTarget{Kind: ShimTarget, Role: shim.role, Command: shim.command, LivePath: shim.path},
				RemovalTarget{Kind: ShimTarget, Role: shim.role, Command: shim.command, LivePath: shim.sidecar},
			)
		}
	}
	return plan, nil
}

// readManagerBinary reads the installed manager executable enforced
// launchers are copied from. It must be a regular file of usable size.
func readManagerBinary(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 {
		return nil, fmt.Errorf("the manager executable is not a regular file")
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- resolved manager executable under staging
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) != info.Size() {
		return nil, fmt.Errorf("the manager executable changed while reading")
	}
	return payload, nil
}

// writeStagedManagerCopy stages one native launcher and verifies the
// staged bytes hash to the manager bytes, so a short write never
// publishes a corrupt launcher.
func writeStagedManagerCopy(path string, manager []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, manager, 0o700); err != nil {
		return err
	}
	staged, err := os.ReadFile(path) // #nosec G304 -- staged launcher just written
	if err != nil {
		return err
	}
	want := sha256.Sum256(manager)
	got := sha256.Sum256(staged)
	if want != got {
		return fmt.Errorf("the staged launcher does not match the manager bytes")
	}
	return nil
}

func writeStagedSidecar(path string, sidecar []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, sidecar, 0o644)
}

func validateManagedEnforcedShim(shim ManagedEnforcedShim) error {
	want, err := NewManagedEnforcedShim(shim.role, shim.binDir, shim.command, shim.platform)
	if err != nil {
		return err
	}
	if want.path != shim.path || want.sidecar != shim.sidecar {
		return fmt.Errorf("enforced shim path is not manager-derived")
	}
	return nil
}

func stagedEnforcedShimPath(stageRoot string, shim ManagedEnforcedShim, base string) string {
	digest := sha256.Sum256([]byte(shim.path))
	return filepath.Join(stageRoot, "shims", string(shim.role), hex.EncodeToString(digest[:]), base)
}

func enforcedSortKey(shim ManagedEnforcedShim) string {
	return string(shim.role) + "\x00" + shim.command + "\x00" + platformPathKey(shim.path, shim.platform)
}
