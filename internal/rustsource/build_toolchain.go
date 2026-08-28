package rustsource

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/relux-works/curator/internal/closuregraph"
)

// BuildToolRole is the closed set of physical native Rust build components.
type BuildToolRole string

const (
	// BuildToolCargo identifies the selected physical Cargo executable.
	BuildToolCargo BuildToolRole = "cargo"
	// BuildToolRustc identifies the selected physical rustc executable.
	BuildToolRustc BuildToolRole = "rustc"
	// BuildToolSysroot identifies the selected Rust sysroot.
	BuildToolSysroot BuildToolRole = "sysroot"
	// BuildToolTargetStdlib identifies the selected native target standard library.
	BuildToolTargetStdlib BuildToolRole = "target_stdlib"
	// BuildToolLinker identifies the selected native linker.
	BuildToolLinker BuildToolRole = "linker"
	// BuildToolSDK identifies the selected native platform SDK.
	BuildToolSDK BuildToolRole = "sdk"
)

var requiredBuildToolRoles = []BuildToolRole{BuildToolCargo, BuildToolLinker, BuildToolRustc, BuildToolSDK, BuildToolSysroot, BuildToolTargetStdlib}

// BuildToolEvidence is the detached C0 identity used to construct C4
// toolchain_component records. PhysicalPath is operational evidence and is not
// itself a portable graph identity.
type BuildToolEvidence struct {
	Role                   BuildToolRole
	PhysicalPath           string
	ExecutableRelativePath string
	ContentFingerprint     closuregraph.ID
	VersionOutput          string
}

type rustBuildToolchain struct {
	target string
	items  map[BuildToolRole]BuildToolEvidence
}

func registerRustBuildToolchain(registration cargoRegistration) (rustBuildToolchain, error) {
	if registration.err != nil {
		return rustBuildToolchain{}, registration.err
	}
	target, ok := nativeRustTarget()
	if !ok {
		return rustBuildToolchain{}, fail(CodeTargetUnsupported, "native Rust target is unsupported", nil)
	}
	root := registration.root
	rustc := filepath.Join(root, "bin", "rustc")
	stdlib := filepath.Join(root, "lib", "rustlib", target, "lib")
	linker := "/usr/bin/cc"
	for _, path := range []string{registration.executable, rustc, stdlib, linker} {
		if info, err := os.Stat(path); err != nil || (path == stdlib && !info.IsDir()) || (path != stdlib && !info.Mode().IsRegular()) {
			return rustBuildToolchain{}, fail(CodeTargetUnsupported, "required native Rust toolchain component is unavailable", map[string]string{"path": path})
		}
	}
	sdk := root
	if runtime.GOOS == "darwin" {
		var sdkErr error
		sdk, sdkErr = registeredDarwinSDK()
		if sdkErr != nil {
			return rustBuildToolchain{}, sdkErr
		}
	}
	fileFingerprint := func(path string) (closuregraph.ID, error) {
		payload, err := os.ReadFile(path) // #nosec G304 -- manager-selected physical tool.
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(payload)
		return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:])), nil
	}
	cargoID, err := fileFingerprint(registration.executable)
	if err != nil {
		return rustBuildToolchain{}, err
	}
	rustcID, err := fileFingerprint(rustc)
	if err != nil {
		return rustBuildToolchain{}, err
	}
	linkerID, err := fileFingerprint(linker)
	if err != nil {
		return rustBuildToolchain{}, err
	}
	stdlibDigest, err := directoryDigest(stdlib)
	if err != nil {
		return rustBuildToolchain{}, err
	}
	sdkID, err := sdkFactsFingerprint(sdk)
	if err != nil {
		return rustBuildToolchain{}, err
	}
	items := map[BuildToolRole]BuildToolEvidence{
		BuildToolCargo:        {Role: BuildToolCargo, PhysicalPath: registration.executable, ExecutableRelativePath: registration.executableRelative, ContentFingerprint: cargoID, VersionOutput: contentVersion(cargoID)},
		BuildToolRustc:        {Role: BuildToolRustc, PhysicalPath: rustc, ExecutableRelativePath: "bin/rustc", ContentFingerprint: rustcID, VersionOutput: contentVersion(rustcID)},
		BuildToolSysroot:      {Role: BuildToolSysroot, PhysicalPath: root, ExecutableRelativePath: "sysroot", ContentFingerprint: closuregraph.ID(registration.rootFingerprint), VersionOutput: contentVersion(closuregraph.ID(registration.rootFingerprint))},
		BuildToolTargetStdlib: {Role: BuildToolTargetStdlib, PhysicalPath: stdlib, ExecutableRelativePath: filepath.ToSlash(filepath.Join("lib", "rustlib", target, "lib")), ContentFingerprint: closuregraph.ID(stdlibDigest), VersionOutput: contentVersion(closuregraph.ID(stdlibDigest))},
		BuildToolLinker:       {Role: BuildToolLinker, PhysicalPath: linker, ExecutableRelativePath: "usr/bin/cc", ContentFingerprint: linkerID, VersionOutput: contentVersion(linkerID)},
		BuildToolSDK:          {Role: BuildToolSDK, PhysicalPath: sdk, ExecutableRelativePath: "sdk", ContentFingerprint: sdkID, VersionOutput: contentVersion(sdkID)},
	}
	return rustBuildToolchain{target: target, items: items}, nil
}

func registeredDarwinSDK() (string, error) {
	developer, err := filepath.EvalSymlinks("/var/db/xcode_select_link")
	if err != nil {
		return "", fail(CodeTargetUnsupported, "closed native Apple developer registry is unavailable", nil)
	}
	sdk := filepath.Join(developer, "Platforms", "MacOSX.platform", "Developer", "SDKs", "MacOSX.sdk")
	sdk, err = filepath.EvalSymlinks(sdk)
	if err != nil {
		return "", fail(CodeTargetUnsupported, "selected native Apple SDK is unavailable", nil)
	}
	info, err := os.Stat(sdk)
	if err != nil || !info.IsDir() {
		return "", fail(CodeTargetUnsupported, "selected native Apple SDK is invalid", nil)
	}
	return sdk, nil
}

func contentVersion(id closuregraph.ID) string { return "content-" + string(id) }

func sdkFactsFingerprint(root string) (closuregraph.ID, error) {
	records := []any{filepath.Base(root)}
	for _, name := range []string{"SDKSettings.json", "SDKSettings.plist", "System/Library/CoreServices/SystemVersion.plist"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		payload, err := os.ReadFile(path) // #nosec G304 -- fixed SDK fact paths below selected root.
		if err == nil {
			sum := sha256.Sum256(payload)
			records = append(records, name, hex.EncodeToString(sum[:]))
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return closuregraph.DomainID("rust-sdk-facts-v1", map[string]any{"records": records})
}

func (toolchain rustBuildToolchain) evidence() []BuildToolEvidence {
	result := make([]BuildToolEvidence, 0, len(requiredBuildToolRoles))
	for _, role := range requiredBuildToolRoles {
		result = append(result, toolchain.items[role])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Role < result[j].Role })
	return result
}

func (toolchain rustBuildToolchain) recheck(registration cargoRegistration) error {
	current, err := registerRustBuildToolchain(registration)
	if err != nil {
		return err
	}
	if current.target != toolchain.target || fmt.Sprint(current.evidence()) != fmt.Sprint(toolchain.evidence()) {
		return fail(CodeToolchainIdentityChanged, "native Rust toolchain identity changed before use", map[string]string{"target": toolchain.target})
	}
	return nil
}

// BuildToolchain returns the exact manager-selected C0 components for C4
// binding construction.
func (m *Manager) BuildToolchain() ([]BuildToolEvidence, error) {
	state, err := m.authority()
	if err != nil {
		return nil, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed {
		return nil, fail(CodeConfigUntrusted, "manager is closed", nil)
	}
	return state.buildTools.evidence(), nil
}
