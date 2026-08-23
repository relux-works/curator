//go:build !windows

package artifactpolicy

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
	"golang.org/x/sys/unix"
)

func TestSelectedToolchainRecheckRejectsRealUnsafeNodesBeforeProcessStart(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "escaping symlink",
			mutate: func(t *testing.T, root string) {
				t.Helper()
				if err := os.Symlink("../outside", filepath.Join(root, "escape")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "special node",
			mutate: func(t *testing.T, root string) {
				t.Helper()
				if err := unix.Mkfifo(filepath.Join(root, "pipe"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			service, selection, dependencies, toolchain, root := selectedFixtureToolchain(t)
			testCase.mutate(t, root)

			processStarts := 0
			executable, err := service.AuthorizeSelectedAdapterExecution(
				t.Context(), selection, dependencies, toolchain,
			)
			if err == nil {
				command := exec.CommandContext(t.Context(), executable, "version")
				if startErr := command.Start(); startErr == nil {
					processStarts++
					_ = command.Process.Kill()
					_ = command.Wait()
				}
				t.Fatalf("unsafe selected toolchain was authorized: %q", executable)
			}
			if processStarts != 0 {
				t.Fatalf("unsafe selected toolchain started %d processes", processStarts)
			}
		})
	}
}

func selectedFixtureToolchain(
	t *testing.T,
) (*Service, *SelectedToolchain, []*Admission, *Admission, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte(runtime.Version()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	toolBytes := conformance.GNUDynamicPIE()
	if err := os.WriteFile(filepath.Join(root, "bin", "go"), toolBytes, 0o700); err != nil {
		t.Fatal(err)
	}
	previousRoot := centrallySelectedGoRoot
	centrallySelectedGoRoot = root
	t.Cleanup(func() { centrallySelectedGoRoot = previousRoot })

	service := NewService()
	source := []byte("package selected\n")
	dependency, err := service.AdmitDependency(t.Context(), dependencyRequest("selected.go", source, ProfileGoV1))
	if err != nil {
		t.Fatal(err)
	}
	dependencies := []*Admission{dependency.Admission}
	selection, err := service.SelectExternalToolchain(t.Context(), ToolchainSelectorRuntimeGoV1, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	toolchain, err := service.AdmitSelectedToolchain(t.Context(), selection, dependencies, ToolchainRequest{
		Descriptor: fixtureDescriptor(toolBytes, ProfileCommonV1),
		Payload: Payload{
			Path: "bin/go", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes),
		},
	}.Descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if toolchain.Admission == nil {
		t.Fatal("selected fixture toolchain has no admission")
	}
	return service, selection, dependencies, toolchain.Admission, root
}
