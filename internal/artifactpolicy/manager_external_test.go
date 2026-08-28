package artifactpolicy_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
)

func TestCentralManagerSelectsFingerprintsAndRechecksRuntimeGoToolchain(t *testing.T) {
	service := artifactpolicy.NewService()
	source := []byte("package selected\n")
	descriptor := externalDescriptor(artifactpolicy.ProfileGoV1)
	descriptor.Origin = artifactpolicy.OriginEvidence{
		Locator: "fixture://selected-root", ImmutableID: "selected-root-r1",
		LockRecord: "selected-root-lock", ChecksumSHA256: externalDigest(source), Verified: true,
	}
	descriptor.DeclaredText = map[string]artifactpolicy.TextDeclaration{
		"selected.go": {Grammar: artifactpolicy.GrammarGo, Class: artifactpolicy.ClassSourceAuthoredText},
	}
	dependency, err := service.AdmitDependency(t.Context(), artifactpolicy.DependencyRequest{
		Descriptor: descriptor,
		Payload: artifactpolicy.Payload{
			Path: "selected.go", Size: int64(len(source)), Reader: bytes.NewReader(source),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	dependencies := []*artifactpolicy.Admission{dependency.Admission}

	// Neither the caller-visible go/build default nor GOROOT environment can
	// redirect the package-internal central selector after initialization.
	callerRoot := filepath.Join(t.TempDir(), "caller-go-root")
	if err := os.MkdirAll(filepath.Join(callerRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(callerRoot, "bin", "go"), conformance.GNUDynamicPIE(), 0o700); err != nil {
		t.Fatal(err)
	}
	previousDefaultRoot := build.Default.GOROOT
	build.Default.GOROOT = callerRoot
	t.Cleanup(func() { build.Default.GOROOT = previousDefaultRoot })
	t.Setenv("GOROOT", callerRoot)

	selection, err := service.SelectExternalToolchain(
		t.Context(), artifactpolicy.ToolchainSelectorRuntimeGoV1, dependencies,
	)
	if err != nil {
		t.Fatal(err)
	}
	toolDescriptor := externalDescriptor(artifactpolicy.ProfileCommonV1)
	toolDescriptor.PackageName = "curator-runtime-go-toolchain"
	toolDescriptor.PackageVersion = runtime.Version()
	toolchain, err := service.AdmitSelectedToolchain(t.Context(), selection, dependencies, toolDescriptor)
	if err != nil {
		t.Fatal(err)
	}
	if toolchain.Manifest.Decision != artifactpolicy.DecisionAllowToolchain || toolchain.Admission == nil {
		t.Fatalf("selected toolchain decision/admission = %q/%v", toolchain.Manifest.Decision, toolchain.Admission)
	}
	if len(toolchain.Manifest.Nodes) == 0 || toolchain.Manifest.Nodes[0].Class != artifactpolicy.ClassNativeExecutable {
		t.Fatalf("selected toolchain root class = %+v", toolchain.Manifest.Nodes)
	}
	roleEvidence := make(map[string]string, len(toolchain.Manifest.RoleEvidence))
	for _, fact := range toolchain.Manifest.RoleEvidence {
		roleEvidence[fact.Key] = fact.Value
	}
	if roleEvidence["policy_selector"] != string(artifactpolicy.ToolchainSelectorRuntimeGoV1) ||
		roleEvidence["authorization"] != "manager-issued-v1" ||
		roleEvidence["checkpoint_fingerprint"] == "" ||
		roleEvidence["checkpoint_fingerprint"] != roleEvidence["time_of_use_fingerprint"] ||
		filepath.Clean(roleEvidence["resolved_root"]) == filepath.Clean(callerRoot) {
		t.Fatalf("selected toolchain role evidence = %#v", roleEvidence)
	}
	processStarts := 0
	executable, err := service.AuthorizeSelectedAdapterExecution(
		t.Context(), selection, dependencies, toolchain.Admission,
	)
	if err != nil {
		t.Fatal(err)
	}
	callerPrefix := filepath.Clean(callerRoot) + string(filepath.Separator)
	if filepath.Clean(executable) == filepath.Clean(filepath.Join(callerRoot, "bin", "go")) ||
		strings.HasPrefix(filepath.Clean(executable), callerPrefix) {
		t.Fatalf("caller root minted trusted executable %q", executable)
	}
	if info, err := os.Stat(executable); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("authorized executable %q is not a real regular file: %v", executable, err)
	}

	executionContext, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	output := &boundedCommandOutput{remaining: 4096}
	command := exec.CommandContext(executionContext, executable, "version")
	command.Stdout = output
	command.Stderr = output
	command.Env = replaceEnvironment(os.Environ(), "GOROOT", roleEvidence["resolved_root"])
	if err := command.Start(); err != nil {
		t.Fatalf("start selected Go executable: %v", err)
	}
	processStarts++
	if err := command.Wait(); err != nil {
		t.Fatalf("wait for selected Go executable: %v; output: %q", err, output.String())
	}
	if executionContext.Err() != nil {
		t.Fatalf("selected Go executable exceeded its context: %v", executionContext.Err())
	}
	wantVersion := "go version " + roleEvidence["version"] + " " + roleEvidence["platform"]
	if got := strings.TrimSpace(output.String()); got != wantVersion {
		t.Fatalf("selected Go version output = %q, want evidence-bound %q", got, wantVersion)
	}
	if processStarts != 1 {
		t.Fatalf("selected toolchain started %d processes", processStarts)
	}

	changedSource := []byte("package changed\n")
	changedDescriptor := externalDescriptor(artifactpolicy.ProfileGoV1)
	changedDescriptor.Origin = artifactpolicy.OriginEvidence{
		Locator: "fixture://selected-root", ImmutableID: "selected-root-r2",
		LockRecord: "selected-root-lock-2", ChecksumSHA256: externalDigest(changedSource), Verified: true,
	}
	changedDescriptor.DeclaredText = map[string]artifactpolicy.TextDeclaration{
		"selected.go": {Grammar: artifactpolicy.GrammarGo, Class: artifactpolicy.ClassSourceAuthoredText},
	}
	changedDependency, changedErr := service.AdmitDependency(t.Context(), artifactpolicy.DependencyRequest{
		Descriptor: changedDescriptor,
		Payload: artifactpolicy.Payload{
			Path: "selected.go", Size: int64(len(changedSource)), Reader: bytes.NewReader(changedSource),
		},
	})
	if changedErr != nil {
		t.Fatal(changedErr)
	}
	if changedDependency.Manifest.ManifestDigest == dependency.Manifest.ManifestDigest {
		t.Fatal("changed dependency bytes did not change the manifest digest")
	}
	startsBeforeRejection := processStarts
	if _, err := service.AuthorizeSelectedAdapterExecution(
		t.Context(), selection, []*artifactpolicy.Admission{changedDependency.Admission}, toolchain.Admission,
	); err == nil {
		t.Fatal("same-count changed dependency boundary was authorized")
	}
	if processStarts != startsBeforeRejection {
		t.Fatalf("changed dependency boundary started %d processes", processStarts-startsBeforeRejection)
	}
}

func TestExternalCallerCreatedRootsAndCopiedObjectsStayUntrusted(t *testing.T) {
	workspace := t.TempDir()
	callerToolchainRoot := filepath.Join(workspace, "caller-toolchain")
	if err := os.MkdirAll(filepath.Join(callerToolchainRoot, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	toolBytes := conformance.GNUDynamicPIE()
	if err := os.WriteFile(filepath.Join(callerToolchainRoot, "bin", "clang"), toolBytes, 0o700); err != nil {
		t.Fatal(err)
	}

	service := artifactpolicy.NewService()
	toolchain, toolchainErr := service.AdmitToolchain(t.Context(), artifactpolicy.ToolchainRequest{
		Descriptor: externalDescriptor(artifactpolicy.ProfileCommonV1),
		Payload: artifactpolicy.Payload{
			Path: "bin/clang", Size: int64(len(toolBytes)), Reader: bytes.NewReader(toolBytes),
		},
	})
	if artifactpolicy.ErrorCode(toolchainErr) != artifactpolicy.CodeToolchainUntrusted {
		t.Fatalf("caller-created toolchain root error = %v", toolchainErr)
	}
	if toolchain.Manifest.Decision != artifactpolicy.DecisionReject || toolchain.Admission != nil {
		t.Fatalf("caller-created toolchain decision/admission = %q/%v", toolchain.Manifest.Decision, toolchain.Admission)
	}

	source := []byte("package main\nfunc main() {}\n")
	dependencyDescriptor := externalDescriptor(artifactpolicy.ProfileGoV1)
	dependencyDescriptor.Origin = artifactpolicy.OriginEvidence{
		Locator: "fixture://dependency", ImmutableID: "dependency-revision-1",
		LockRecord: "dependency-lock-1", ChecksumSHA256: externalDigest(source), Verified: true,
	}
	dependencyDescriptor.DeclaredText = map[string]artifactpolicy.TextDeclaration{
		"main.go": {Grammar: artifactpolicy.GrammarGo, Class: artifactpolicy.ClassSourceAuthoredText},
	}
	dependency, dependencyErr := service.AdmitDependency(t.Context(), artifactpolicy.DependencyRequest{
		Descriptor: dependencyDescriptor,
		Payload: artifactpolicy.Payload{
			Path: "main.go", Size: int64(len(source)), Reader: bytes.NewReader(source),
		},
	})
	if dependencyErr != nil {
		t.Fatal(dependencyErr)
	}
	adapterStarts := 0
	if executionErr := artifactpolicy.AuthorizeAdapterExecution(
		[]*artifactpolicy.Admission{dependency.Admission}, toolchain.Admission,
	); executionErr == nil {
		adapterStarts++
	}
	if adapterStarts != 0 {
		t.Fatalf("caller-created toolchain started %d adapter actions", adapterStarts)
	}

	preexistingObject := conformance.ELF64(1, false, false, false)
	ambientPath := filepath.Join(workspace, "ambient", "main.o")
	callerStagingPath := filepath.Join(workspace, "caller-staging", "obj", "main.o")
	for _, parent := range []string{filepath.Dir(ambientPath), filepath.Dir(callerStagingPath)} {
		if err := os.MkdirAll(parent, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(ambientPath, preexistingObject, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(ambientPath, callerStagingPath); err != nil {
		t.Fatal(err)
	}
	ambientInfo, err := os.Stat(ambientPath)
	if err != nil {
		t.Fatal(err)
	}
	stagingInfo, err := os.Stat(callerStagingPath)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(ambientInfo, stagingInfo) {
		t.Fatal("pre-existing object fixture is not an actual hard link")
	}
	output, outputErr := service.AdmitLocalOutput(t.Context(), artifactpolicy.LocalOutputRequest{
		Descriptor: externalDescriptor(artifactpolicy.ProfileCommonV1),
		Payload: artifactpolicy.Payload{
			Path: "obj/main.o", Size: int64(len(preexistingObject)), Reader: bytes.NewReader(preexistingObject),
		},
	})
	if artifactpolicy.ErrorCode(outputErr) != artifactpolicy.CodeLocalOutputUnreceipted {
		t.Fatalf("copied pre-existing output error = %v", outputErr)
	}
	if output.Manifest.Decision != artifactpolicy.DecisionReject || output.Admission != nil {
		t.Fatalf("copied pre-existing output decision/admission = %q/%v", output.Manifest.Decision, output.Admission)
	}
	cachePublications := 0
	if publicationErr := artifactpolicy.AuthorizeCachePublication(output.Admission); publicationErr == nil {
		cachePublications++
	}
	if cachePublications != 0 {
		t.Fatalf("copied pre-existing output published %d cache entries", cachePublications)
	}
}

type boundedCommandOutput struct {
	buffer    bytes.Buffer
	remaining int64
}

func (output *boundedCommandOutput) Write(payload []byte) (int, error) {
	if int64(len(payload)) > output.remaining {
		return 0, fmt.Errorf("command output exceeds 4096 bytes")
	}
	count, err := output.buffer.Write(payload)
	output.remaining -= int64(count)
	return count, err
}

func (output *boundedCommandOutput) String() string { return output.buffer.String() }

func replaceEnvironment(environment []string, key, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}
	return append(result, prefix+value)
}

var _ io.Writer = (*boundedCommandOutput)(nil)

func externalDescriptor(profile artifactpolicy.ProfileID) artifactpolicy.Descriptor {
	return artifactpolicy.Descriptor{
		AdapterID: "external-fixture-adapter-v1", ProfileID: profile, Manager: "external-fixture-manager",
		PackageName: "fixture-package", PackageVersion: "1.0.0",
	}
}

func externalDigest(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}
