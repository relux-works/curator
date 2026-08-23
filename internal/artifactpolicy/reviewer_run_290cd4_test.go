package artifactpolicy

import (
	"archive/tar"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
)

func TestManagerSelectedToolchainModeMutationStopsAdapterExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose portable executable permission bits")
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte(runtime.Version()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	toolBytes := conformance.GNUDynamicPIE()
	toolName := "go"
	if runtime.GOOS == "windows" {
		toolName = "go.exe"
	}
	if err := os.WriteFile(filepath.Join(root, "bin", toolName), toolBytes, 0o700); err != nil {
		t.Fatal(err)
	}
	previousRoot := centrallySelectedGoRoot
	centrallySelectedGoRoot = root
	t.Cleanup(func() { centrallySelectedGoRoot = previousRoot })

	service := NewService()
	source := []byte("package selected\n")
	dependencyResult, err := service.AdmitDependency(t.Context(), dependencyRequest("selected.go", source, ProfileGoV1))
	if err != nil {
		t.Fatal(err)
	}
	dependencies := []*Admission{dependencyResult.Admission}
	selection, err := service.SelectExternalToolchain(t.Context(), ToolchainSelectorRuntimeGoV1, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	toolchain, err := service.AdmitSelectedToolchain(
		t.Context(), selection, dependencies, fixtureDescriptor(toolBytes, ProfileCommonV1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "bin", toolName), 0o600); err != nil {
		t.Fatal(err)
	}
	adapterStarts := 0
	if _, err := service.AuthorizeSelectedAdapterExecution(
		t.Context(), selection, dependencies, toolchain.Admission,
	); err == nil {
		adapterStarts++
	}
	if adapterStarts != 0 {
		t.Fatalf("drifted selected toolchain started %d adapter actions", adapterStarts)
	}
}

func TestMetadataFullPathBudgetsPrecedeSparsePayloadReads(t *testing.T) {
	const declaredSize = int64(64 << 20)
	containerPath := pathWithLength(t, DefaultLimits().MaxPathBytes-2-16)
	tests := []struct {
		name     string
		typeflag byte
		prefix   func(int64) []byte
		walk     func(*inspector, blob) *Diagnostic
	}{
		{
			name: "PAX local path", typeflag: tar.TypeXHeader,
			prefix: func(size int64) []byte {
				prefix := fmt.Sprintf("%d path=", size)
				if int64(len(prefix)+1) >= size {
					panic("invalid PAX sparse fixture")
				}
				return []byte(prefix)
			},
		},
		{
			name: "PAX global link", typeflag: tar.TypeXGlobalHeader,
			prefix: func(size int64) []byte {
				prefix := fmt.Sprintf("%d linkpath=", size)
				if int64(len(prefix)+1) >= size {
					panic("invalid PAX sparse fixture")
				}
				return []byte(prefix)
			},
		},
		{
			name: "GNU long name", typeflag: tar.TypeGNULongName,
			prefix: func(int64) []byte { return nil },
		},
		{
			name: "GNU long link", typeflag: tar.TypeGNULongLink,
			prefix: func(int64) []byte { return nil },
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			item := sparseTarMetadataBlob(t, testCase.typeflag, declaredSize, testCase.prefix(declaredSize))
			limits := DefaultLimits()
			account, err := newLimitAccountant(limits, item.size)
			if err != nil {
				t.Fatal(err)
			}
			worker := &inspector{ctx: t.Context(), limits: limits, account: account}
			runtime.GC()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			_, diagnostic := worker.scanTarPhysical(item, containerPath, nil)
			runtime.ReadMemStats(&after)
			if diagnostic == nil || diagnostic.Code != CodeArchiveUnsafePath ||
				diagnostic.LimitName != "max_path_bytes" || diagnostic.Observed <= limits.MaxPathBytes {
				t.Fatalf("sparse metadata path diagnostic = %+v", diagnostic)
			}
			if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 16<<20 {
				t.Fatalf("sparse metadata path preflight allocated %d bytes", allocated)
			}
		})
	}

	t.Run("BSD extended name", func(t *testing.T) {
		item := sparseARBlob(t, "#1/"+strconv.FormatInt(declaredSize, 10), declaredSize, nil)
		diagnostic, allocated := walkSparseAR(t, item, containerPath)
		if diagnostic.Code != CodeArchiveUnsafePath || diagnostic.LimitName != "max_path_bytes" ||
			diagnostic.Observed <= DefaultLimits().MaxPathBytes {
			t.Fatalf("BSD sparse-name diagnostic = %+v", diagnostic)
		}
		if allocated > 16<<20 {
			t.Fatalf("BSD sparse-name preflight allocated %d bytes", allocated)
		}
	})

	t.Run("GNU string-table name", func(t *testing.T) {
		item := sparseARBlob(t, "//", declaredSize, bytes.Repeat([]byte{'x'}, 64))
		diagnostic, allocated := walkSparseAR(t, item, containerPath)
		if diagnostic.Code != CodeArchiveUnsafePath || diagnostic.LimitName != "max_path_bytes" ||
			diagnostic.Observed <= DefaultLimits().MaxPathBytes {
			t.Fatalf("GNU sparse-name diagnostic = %+v", diagnostic)
		}
		if allocated > 16<<20 {
			t.Fatalf("GNU sparse-name preflight allocated %d bytes", allocated)
		}
	})
}

func TestMetadataPathPreflightProducesCanonicalFailureEvidence(t *testing.T) {
	containerPath := pathWithLength(t, DefaultLimits().MaxPathBytes-2-16)
	metadata := paxRecord(t, "path", strings.Repeat("p", 17))
	payload := rawTarArchive(rawTarEntry("PaxHeaders/source", tar.TypeXHeader, metadata, false))
	result, err := NewService().AdmitDependency(t.Context(), DependencyRequest{
		Descriptor: fixtureDescriptor(payload, ProfileCommonV1),
		Payload:    Payload{Path: containerPath, Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
	})
	requireCode(t, err, CodeArchiveUnsafePath)
	if len(result.CanonicalBytes) == 0 || result.Manifest.Diagnostics[0].LimitName != "max_path_bytes" ||
		result.Manifest.Diagnostics[0].Observed <= DefaultLimits().MaxPathBytes {
		t.Fatalf("canonical metadata path failure = %+v", result.Manifest.Diagnostics)
	}
	if _, decodeErr := DecodeManifest(result.CanonicalBytes); decodeErr != nil {
		t.Fatalf("decode canonical metadata path failure: %v", decodeErr)
	}
}

func TestGNUARStringTableValidatesEveryUnreferencedNameOrderIndependently(t *testing.T) {
	for _, table := range [][]byte{
		[]byte("safe.go/\n../escape/\n"),
		[]byte("../escape/\nsafe.go/\n"),
	} {
		payload := rawARArchive(rawARMember("//", table))
		item, store := captureMetadataLimitBlob(t, payload, DefaultLimits())
		account, err := newLimitAccountant(DefaultLimits(), item.size)
		if err != nil {
			t.Fatal(err)
		}
		worker := &inspector{
			ctx: t.Context(), limits: DefaultLimits(), account: account, store: store,
			nodes: []ManifestNode{{
				Path: "libcase.a", Kind: NodeArchive, Class: ClassNativeLibraryStatic,
				Decision: DecisionDescend, InspectionComplete: true,
			}},
		}
		if !worker.walkAR(0, item, 1) {
			t.Fatal("unsafe unreferenced GNU string-table name was not rejected")
		}
		diagnostic := worker.findings.recorded()[0]
		if diagnostic.Code != CodeArchiveUnsafePath || diagnostic.Path != "libcase.a!/$ar-metadata/string-table-001" {
			t.Fatalf("unreferenced GNU name diagnostic = %+v", diagnostic)
		}
	}
}

func TestBSDARNameAccountingAndManifestEvidenceAreExact(t *testing.T) {
	name := "nested/source.go"
	content := []byte("package fixture\n")
	payload := rawARArchive(rawBSDARMember(name, content))
	descriptor := fixtureDescriptor(payload, ProfileCommonV1)
	result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
		Descriptor:    descriptor,
		Payload:       Payload{Path: "libsource.a", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
		Authorization: validToolchainAuthorization(t, "libsource.a", payload),
	})
	if err != nil {
		t.Fatal(err)
	}
	requireDecision(t, result, DecisionAllowToolchain)
	node := requireNode(t, result, "libsource.a!/"+name)
	physicalSize := int64(len(name) + len(content))
	if result.Manifest.Accounting.TotalEmittedBytes != physicalSize ||
		result.Manifest.Accounting.ManifestedEmittedBytes != physicalSize ||
		result.Manifest.Accounting.UnmanifestedEmittedBytes != 0 ||
		result.Manifest.Accounting.MaxObservedLeafBytes != physicalSize ||
		result.Manifest.Accounting.MaxManifestedLeafBytes != physicalSize {
		t.Fatalf("BSD name accounting = %+v, want physical size %d", result.Manifest.Accounting, physicalSize)
	}
	if factValue(t, node, "archive-ar-v1", "declared_size") != strconv.FormatInt(physicalSize, 10) ||
		factValue(t, node, "archive-ar-v1", "member_size") != strconv.Itoa(len(content)) ||
		factValue(t, node, "archive-ar-v1", "extended_name_size") != strconv.Itoa(len(name)) ||
		factValue(t, node, "archive-ar-v1", "extended_name_sha256") != digestBytes([]byte(name)) {
		t.Fatalf("BSD extended-name facts = %#v", node.Observations)
	}
	original, decodeErr := base64.StdEncoding.DecodeString(node.OriginalNameBase64)
	if decodeErr != nil || string(original) != name {
		t.Fatalf("BSD original name = %q, %v", original, decodeErr)
	}
	encoded, encodeErr := EncodeManifest(result.Manifest)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if _, decodeErr := DecodeManifest(encoded); decodeErr != nil {
		t.Fatalf("decode exact BSD manifest: %v", decodeErr)
	}

	mutations := []struct {
		name   string
		mutate func(*ManifestNode)
	}{
		{
			name: "extended-name size",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "extended_name_size", strconv.Itoa(len(name)+1))
			},
		},
		{
			name: "extended-name digest",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "extended_name_sha256", digestBytes([]byte("other.go")))
			},
		},
		{
			name: "declared size",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "declared_size", strconv.FormatInt(physicalSize+1, 10))
			},
		},
		{
			name: "member size",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "member_size", strconv.Itoa(len(content)+1))
			},
		},
		{
			name: "original name",
			mutate: func(node *ManifestNode) {
				node.OriginalNameBase64 = base64.StdEncoding.EncodeToString([]byte("forged.go"))
			},
		},
		{
			name: "canonical logical path",
			mutate: func(node *ManifestNode) {
				node.Path = "libsource.a!/nested/forged.go"
			},
		},
	}
	for _, testCase := range mutations {
		t.Run(testCase.name, func(t *testing.T) {
			forged := cloneManifest(t, result.Manifest)
			testCase.mutate(requireManifestNode(t, &forged, "libsource.a!/"+name))
			if err := decodeRehashedManifest(forged); err == nil {
				t.Fatal("self-rehashed BSD native-archive evidence was accepted")
			}
		})
	}
	t.Run("manifest accounting", func(t *testing.T) {
		forged := cloneManifest(t, result.Manifest)
		forgeManifestedAccounting(&forged)
		if err := decodeRehashedManifest(forged); err == nil {
			t.Fatal("self-rehashed BSD native-archive accounting was accepted")
		}
	})
}

func TestBSDExtendedSymbolTableNameRemainsExactMetadata(t *testing.T) {
	name := "__.SYMDEF"
	payload := rawARArchive(rawBSDARMember(name, make([]byte, 8)))
	result, err := NewService().AdmitToolchain(t.Context(), ToolchainRequest{
		Descriptor:    fixtureDescriptor(payload, ProfileCommonV1),
		Payload:       Payload{Path: "libmetadata.a", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
		Authorization: validToolchainAuthorization(t, "libmetadata.a", payload),
	})
	if err != nil {
		t.Fatal(err)
	}
	node := requireNode(t, result, "libmetadata.a!/$ar-metadata/bsd-symbol-table-001")
	if node.Variant != "ar.metadata.bsd_symbol_table_001" ||
		factValue(t, node, "archive-ar-v1", "extended_name_sha256") != digestBytes([]byte(name)) {
		t.Fatalf("BSD symbol-table metadata = %+v", node)
	}
	if _, err := DecodeManifest(result.CanonicalBytes); err != nil {
		t.Fatalf("decode BSD symbol-table manifest: %v", err)
	}

	mutations := []struct {
		name   string
		mutate func(*ManifestNode)
	}{
		{
			name: "extended-name size",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "extended_name_size", strconv.Itoa(len(name)+1))
			},
		},
		{
			name: "extended-name digest",
			mutate: func(node *ManifestNode) {
				setNodeObservationFact(node, "archive-ar-v1", "extended_name_sha256", digestBytes([]byte("__.SYMDEF SORTED")))
			},
		},
		{
			name: "original name",
			mutate: func(node *ManifestNode) {
				node.OriginalNameBase64 = base64.StdEncoding.EncodeToString([]byte("__.SYMDEF SORTED"))
			},
		},
		{
			name: "canonical logical path",
			mutate: func(node *ManifestNode) {
				node.Path = "libmetadata.a!/$ar-metadata/bsd-symbol-table-002"
			},
		},
	}
	for _, testCase := range mutations {
		t.Run(testCase.name, func(t *testing.T) {
			forged := cloneManifest(t, result.Manifest)
			testCase.mutate(requireManifestNode(t, &forged, "libmetadata.a!/$ar-metadata/bsd-symbol-table-001"))
			if err := decodeRehashedManifest(forged); err == nil {
				t.Fatal("self-rehashed BSD symbol-table evidence was accepted")
			}
		})
	}
	t.Run("manifest accounting", func(t *testing.T) {
		forged := cloneManifest(t, result.Manifest)
		forgeManifestedAccounting(&forged)
		if err := decodeRehashedManifest(forged); err == nil {
			t.Fatal("self-rehashed BSD symbol-table accounting was accepted")
		}
	})
}

func sparseTarMetadataBlob(t *testing.T, typeflag byte, declaredSize int64, prefix []byte) blob {
	t.Helper()
	header := rawTarDeclaredHeader("metadata", typeflag, declaredSize, typeflag == tar.TypeGNULongName || typeflag == tar.TypeGNULongLink)
	file, err := os.CreateTemp(t.TempDir(), "sparse-metadata-*.tar")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if _, err := file.WriteAt(header, 0); err != nil {
		t.Fatal(err)
	}
	if len(prefix) > 0 {
		if _, err := file.WriteAt(prefix, 512); err != nil {
			t.Fatal(err)
		}
	}
	padded, ok := roundTarBlock(declaredSize)
	if !ok {
		t.Fatal("tar sparse fixture size overflow")
	}
	end := int64(512) + padded
	if _, err := file.WriteAt(make([]byte, 1024), end); err != nil {
		t.Fatal(err)
	}
	return blob{file: file, size: end + 1024}
}

func sparseARBlob(t *testing.T, name string, declaredSize int64, prefix []byte) blob {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "sparse-metadata-*.a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	header := []byte(fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "100644", declaredSize))
	if len(header) != 60 {
		t.Fatalf("ar header length = %d", len(header))
	}
	if _, err := file.WriteAt(append([]byte("!<arch>\n"), header...), 0); err != nil {
		t.Fatal(err)
	}
	if len(prefix) > 0 {
		if _, err := file.WriteAt(prefix, 68); err != nil {
			t.Fatal(err)
		}
	}
	end := int64(68) + declaredSize + declaredSize%2
	if _, err := file.WriteAt([]byte{0}, end-1); err != nil {
		t.Fatal(err)
	}
	return blob{file: file, size: end}
}

func walkSparseAR(t *testing.T, item blob, containerPath string) (Diagnostic, uint64) {
	t.Helper()
	limits := DefaultLimits()
	account, err := newLimitAccountant(limits, item.size)
	if err != nil {
		t.Fatal(err)
	}
	worker := &inspector{
		ctx: t.Context(), limits: limits, account: account,
		nodes: []ManifestNode{{
			Path: containerPath, Kind: NodeArchive, Class: ClassNativeLibraryStatic,
			Decision: DecisionDescend, InspectionComplete: true,
		}},
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if !worker.walkAR(0, item, 1) {
		t.Fatal("sparse native-archive metadata was not rejected")
	}
	runtime.ReadMemStats(&after)
	if worker.findings == nil || worker.findings.total != 1 {
		t.Fatalf("native-archive findings = %+v", worker.findings)
	}
	return worker.findings.recorded()[0], after.TotalAlloc - before.TotalAlloc
}

func pathWithLength(t *testing.T, target int64) string {
	t.Helper()
	if target <= 0 {
		t.Fatal("invalid target path length")
	}
	components := make([]string, 0, target/201+1)
	remaining := target
	for remaining > 0 {
		if len(components) > 0 {
			remaining--
		}
		length := remaining
		if length > 200 {
			length = 200
		}
		components = append(components, strings.Repeat("a", int(length)))
		remaining -= length
	}
	result := strings.Join(components, "/")
	if int64(len(result)) != target {
		t.Fatalf("path length = %d, want %d", len(result), target)
	}
	if _, err := ValidateVirtualPath(result); err != nil {
		t.Fatalf("generated path is invalid: %v", err)
	}
	return result
}

func rawARArchive(members ...[]byte) []byte {
	payload := []byte("!<arch>\n")
	for _, member := range members {
		payload = append(payload, member...)
	}
	return payload
}

func rawARMember(name string, data []byte) []byte {
	header := []byte(fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "100644", len(data)))
	payload := append(header, data...)
	if len(data)%2 != 0 {
		payload = append(payload, '\n')
	}
	return payload
}

func rawBSDARMember(name string, data []byte) []byte {
	physical := append([]byte(name), data...)
	return rawARMember("#1/"+strconv.Itoa(len(name)), physical)
}

func setNodeObservationFact(node *ManifestNode, detector, key, value string) {
	for observationIndex := range node.Observations {
		observation := &node.Observations[observationIndex]
		if observation.DetectorID != detector {
			continue
		}
		for factIndex := range observation.Facts {
			if observation.Facts[factIndex].Key == key {
				observation.Facts[factIndex].Value = value
				return
			}
		}
	}
}

func forgeManifestedAccounting(manifest *Manifest) {
	manifest.Accounting.TotalEmittedBytes++
	manifest.Accounting.ManifestedEmittedBytes++
	manifest.Accounting.MaxObservedLeafBytes++
	manifest.Accounting.MaxManifestedLeafBytes++
	manifest.Accounting.AggregateExpansionRatio = ceilingRatio(
		manifest.Accounting.TotalEmittedBytes,
		manifest.Accounting.RawPayloadBytes,
	)
}
