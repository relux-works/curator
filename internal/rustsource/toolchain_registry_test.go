package rustsource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSupportedRustToolchainVersionIsTheRegistrySource(t *testing.T) {
	if pinnedRustToolchainPrefix != SupportedRustToolchainVersion+"-" {
		t.Fatalf("pinnedRustToolchainPrefix = %q, want %q", pinnedRustToolchainPrefix, SupportedRustToolchainVersion+"-")
	}
	if len(approvedCargoDescriptors) == 0 {
		t.Fatal("approvedCargoDescriptors is empty")
	}
	for target, descriptor := range approvedCargoDescriptors {
		if descriptor.Version != SupportedRustToolchainVersion {
			t.Fatalf("approvedCargoDescriptors[%q].Version = %q, want %q", target, descriptor.Version, SupportedRustToolchainVersion)
		}
	}
}

func TestCargoToolchainValidateAdmitsOnlyTheSupportedVersion(t *testing.T) {
	// The path must be absolute on every GOOS: a Unix-style "/pinned/..."
	// literal is not absolute on Windows (no volume), and validate rightly
	// rejects it there. Build it from t.TempDir like the sibling cases.
	pinnedRoot := filepath.Join(t.TempDir(), "toolchains", SupportedRustToolchainVersion+"-test")
	tool := cargoToolchain{
		CargoPath: filepath.Join(pinnedRoot, "bin", "cargo"),
		Version:   SupportedRustToolchainVersion, ImplementationCommit: "ea2d97820c16195b0ca3fadb4319fe512c199a43",
		BinarySHA256: "sha256:" + strings.Repeat("a", 64),
		Fingerprint:  "sha256:" + strings.Repeat("b", 64), C0CheckpointID: "sha256:" + strings.Repeat("c", 64),
	}
	if err := tool.validate(); err != nil {
		t.Fatalf("supported toolchain rejected: %v", err)
	}
	tool.Version = "9.9.9-drift-probe"
	if err := tool.validate(); err == nil {
		t.Fatal("drifted toolchain version admitted")
	}
}

func TestCargoHostCapabilityReasonClassifiesOnlyAbsence(t *testing.T) {
	target := "aarch64-apple-darwin"
	missingReason := "pinned Cargo toolchain root or executable unavailable for native target " + target

	t.Run("descriptor absent", func(t *testing.T) {
		if got := cargoHostCapabilityReason(target, false, "", ""); got != "no operator-approved Cargo descriptor for native target "+target {
			t.Fatalf("reason = %q", got)
		}
	})

	t.Run("toolchain root absent", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "absent")
		if got := cargoHostCapabilityReason(target, true, root, filepath.Join(root, "bin", "cargo")); got != missingReason {
			t.Fatalf("reason = %q, want %q", got, missingReason)
		}
	})

	t.Run("executable absent", func(t *testing.T) {
		root := t.TempDir()
		if got := cargoHostCapabilityReason(target, true, root, filepath.Join(root, "bin", "cargo")); got != missingReason {
			t.Fatalf("reason = %q, want %q", got, missingReason)
		}
	})

	t.Run("present executable identity is not inferred", func(t *testing.T) {
		root := t.TempDir()
		executable := filepath.Join(root, "bin", "cargo")
		if err := os.MkdirAll(filepath.Dir(executable), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(executable, []byte("not an approved Cargo executable"), 0o700); err != nil {
			t.Fatal(err)
		}
		if got := cargoHostCapabilityReason(target, true, root, executable); got != "" {
			t.Fatalf("present mismatched executable was classified as unavailable: %q", got)
		}
		rootInfo, err := os.Stat(root)
		if err != nil {
			t.Fatal(err)
		}
		registration := cargoRegistration{
			root:             root,
			executable:       executable,
			rootInfo:         rootInfo,
			executableSHA256: "sha256:" + strings.Repeat("0", 64),
			descriptor: approvedCargoDescriptor{
				Version: "1.91.0", ImplementationCommit: "fixture",
				ExecutableSHA256: "sha256:" + strings.Repeat("1", 64),
			},
		}
		if _, err := registration.recheck(t.Context()); err == nil || !strings.Contains(err.Error(), "registered Cargo executable changed") {
			t.Fatalf("present mismatched executable was not fatal: %v", err)
		}
	})

	t.Run("present non-directory root is not absence", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "toolchain")
		if err := os.WriteFile(root, []byte("not a directory"), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := cargoHostCapabilityReason(target, true, root, filepath.Join(root, "bin", "cargo")); got != "" {
			t.Fatalf("malformed present root was classified as unavailable: %q", got)
		}
	})
}
