package rustsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const pinnedRustToolchainPrefix = "1.91.0-"

type approvedCargoDescriptor struct {
	Version, ImplementationCommit, ExecutableSHA256 string
}

// approvedCargoDescriptors is the closed operator registry for the exact
// Cargo release used by cargo-vendor-transform-v1. A directory label never
// establishes a release claim: the selected executable bytes must match this
// independently reviewed identity before Cargo is registered at C0.
var approvedCargoDescriptors = map[string]approvedCargoDescriptor{
	"aarch64-apple-darwin": {
		Version: "1.91.0", ImplementationCommit: "ea2d97820c16195b0ca3fadb4319fe512c199a43",
		ExecutableSHA256: "sha256:0da859e1130e00a81dac84fa1e86a3dbdd968ddfccef627a8d37255fcbb39e78",
	},
}

// NativeCargoUnavailableReason reports only the two host-capability absences
// that prevent the native Cargo path from starting: no approved descriptor for
// this target, or no pinned toolchain root/executable on disk. A present
// executable is not trusted here; registerCargoAtC0 remains the sole authority
// that verifies its canonical path, bytes, descriptor, and whole-root identity.
func NativeCargoUnavailableReason() string {
	target, supported := nativeRustTarget()
	if !supported {
		return ""
	}
	_, approved := approvedCargoDescriptors[target]
	if !approved {
		return "no operator-approved Cargo descriptor for native target " + target
	}
	currentUser, err := user.Current()
	if err != nil || currentUser.HomeDir == "" {
		return ""
	}
	root := filepath.Join(currentUser.HomeDir, ".rustup", "toolchains", "1.91.0-"+target)
	executable := filepath.Join(root, "bin", cargoExecutableName())
	return cargoHostCapabilityReason(target, true, root, executable)
}

func cargoHostCapabilityReason(target string, approved bool, root, executable string) string {
	if !approved {
		return "no operator-approved Cargo descriptor for native target " + target
	}
	rootInfo, err := os.Stat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return "pinned Cargo toolchain root or executable unavailable for native target " + target
	}
	if err != nil || !rootInfo.IsDir() {
		return ""
	}
	if _, err := os.Stat(executable); errors.Is(err, fs.ErrNotExist) {
		return "pinned Cargo toolchain root or executable unavailable for native target " + target
	}
	return ""
}

// cargoRegistration is selected by NewManager while establishing C0. Adapter
// requests cannot replace it and selection never performs PATH, environment,
// shell, or rustup process lookup.
type cargoRegistration struct {
	root, executable, executableRelative string
	rootFingerprint, executableSHA256    string
	rootInfo                             fs.FileInfo
	descriptor                           approvedCargoDescriptor
	err                                  error
}

func registerCargoAtC0(ctx context.Context) cargoRegistration {
	currentUser, err := user.Current()
	if err != nil || currentUser.HomeDir == "" {
		return cargoRegistration{err: fmt.Errorf("resolve operator home for pinned Cargo: %w", err)}
	}
	target, ok := nativeRustTarget()
	if !ok {
		return cargoRegistration{err: fmt.Errorf("native Rust target is unsupported: %s/%s", runtime.GOOS, runtime.GOARCH)}
	}
	descriptor, ok := approvedCargoDescriptors[target]
	if !ok {
		return cargoRegistration{err: fmt.Errorf("no operator-approved Cargo descriptor for native target %s", target)}
	}
	root := filepath.Join(currentUser.HomeDir, ".rustup", "toolchains", "1.91.0-"+target)
	executable := filepath.Join(root, "bin", cargoExecutableName())
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return cargoRegistration{err: fmt.Errorf("canonicalize pinned Cargo executable: %w", err)}
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return cargoRegistration{err: err}
	}
	root = filepath.Dir(filepath.Dir(executable))
	if !strings.HasPrefix(filepath.Base(root), pinnedRustToolchainPrefix) {
		return cargoRegistration{err: fmt.Errorf("pinned Cargo root has unexpected identity %q", filepath.Base(root))}
	}
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return cargoRegistration{err: fmt.Errorf("pinned Cargo root is unavailable")}
	}
	relative, err := filepath.Rel(root, executable)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return cargoRegistration{err: fmt.Errorf("pinned Cargo executable escapes its tool root")}
	}
	executableBytes, err := os.ReadFile(executable) // #nosec G304 -- closed startup-selected executable.
	if err != nil {
		return cargoRegistration{err: err}
	}
	executableSum := sha256.Sum256(executableBytes)
	executableSHA := "sha256:" + hex.EncodeToString(executableSum[:])
	if executableSHA != descriptor.ExecutableSHA256 {
		return cargoRegistration{err: fmt.Errorf("selected Cargo bytes are not operator-approved for %s", target)}
	}
	fingerprint, err := fingerprintCargoRoot(ctx, root, rootInfo)
	if err != nil {
		return cargoRegistration{err: err}
	}
	return cargoRegistration{
		root: root, executable: executable, executableRelative: filepath.ToSlash(relative),
		rootFingerprint:  fingerprint,
		executableSHA256: executableSHA,
		rootInfo:         rootInfo,
		descriptor:       descriptor,
	}
}

func nativeRustTarget() (string, bool) {
	arch := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[runtime.GOARCH]
	osName := map[string]string{"darwin": "apple-darwin", "linux": "unknown-linux-gnu", "windows": "pc-windows-msvc"}[runtime.GOOS]
	if arch == "" || osName == "" {
		return "", false
	}
	return arch + "-" + osName, true
}

// cargoExecutableName is the pinned Cargo binary's file name below the
// toolchain root's bin directory on this platform.
func cargoExecutableName() string {
	if runtime.GOOS == "windows" {
		return "cargo.exe"
	}
	return "cargo"
}

func (registration cargoRegistration) recheck(ctx context.Context) (cargoToolchain, error) {
	if registration.err != nil {
		return cargoToolchain{}, fail(CodeVendorTransformUnsupported, registration.err.Error(), nil)
	}
	current, err := os.Stat(registration.root)
	if err != nil || !current.IsDir() || !os.SameFile(registration.rootInfo, current) {
		return cargoToolchain{}, fail(CodeVendorTransformUnsupported, "registered Cargo root identity changed", nil)
	}
	payload, err := os.ReadFile(registration.executable) // #nosec G304 -- sealed startup registration.
	if err != nil {
		return cargoToolchain{}, err
	}
	sum := sha256.Sum256(payload)
	executableSHA := "sha256:" + hex.EncodeToString(sum[:])
	if executableSHA != registration.executableSHA256 {
		return cargoToolchain{}, fail(CodeVendorTransformUnsupported, "registered Cargo executable changed", nil)
	}
	if registration.descriptor.ExecutableSHA256 != executableSHA || registration.descriptor.Version == "" || registration.descriptor.ImplementationCommit == "" {
		return cargoToolchain{}, fail(CodeVendorTransformUnsupported, "registered Cargo descriptor differs from approved executable bytes", nil)
	}
	fingerprint, err := fingerprintCargoRoot(ctx, registration.root, registration.rootInfo)
	if err != nil {
		return cargoToolchain{}, err
	}
	if fingerprint != registration.rootFingerprint {
		return cargoToolchain{}, fail(CodeVendorTransformUnsupported, "registered Cargo tool root changed", nil)
	}
	return cargoToolchain{
		CargoPath: registration.executable, Version: registration.descriptor.Version,
		ImplementationCommit: registration.descriptor.ImplementationCommit,
		BinarySHA256:         executableSHA, Fingerprint: fingerprint,
		C0CheckpointID: fingerprint,
	}, nil
}

func fingerprintCargoRoot(ctx context.Context, root string, expected fs.FileInfo) (string, error) {
	current, err := os.Stat(root)
	if err != nil || expected == nil || !os.SameFile(expected, current) {
		return "", fmt.Errorf("registered Cargo root identity changed")
	}
	type record struct {
		path, link string
		mode       fs.FileMode
		size       int64
	}
	records := []record{}
	err = filepath.WalkDir(root, func(currentPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(root, currentPath)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		item := record{path: filepath.ToSlash(relative), mode: info.Mode(), size: info.Size()}
		if info.Mode()&fs.ModeSymlink != 0 {
			item.link, err = os.Readlink(currentPath)
			if err != nil || filepath.IsAbs(item.link) {
				return fmt.Errorf("registered Cargo root contains an invalid link %q", item.path)
			}
		} else if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("registered Cargo root contains a special node %q", item.path)
		}
		records = append(records, item)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].path < records[j].path })
	hash := sha256.New()
	_, _ = io.WriteString(hash, "curator-rust-toolchain-root-v1\x00")
	for _, item := range records {
		_, _ = io.WriteString(hash, item.path+"\x00"+item.mode.String()+"\x00"+fmt.Sprint(item.size)+"\x00"+item.link+"\x00")
		if item.mode.IsRegular() {
			file, openErr := os.Open(filepath.Join(root, filepath.FromSlash(item.path))) // #nosec G304 -- sealed root walk.
			if openErr != nil {
				return "", openErr
			}
			_, copyErr := io.Copy(hash, file)
			closeErr := file.Close()
			if copyErr != nil {
				return "", copyErr
			}
			if closeErr != nil {
				return "", closeErr
			}
		}
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
