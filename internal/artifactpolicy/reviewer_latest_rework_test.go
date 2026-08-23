package artifactpolicy

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
)

func TestCanonicalTreeManifestRederivesCapturedIdentity(t *testing.T) {
	root := makeSourceTree(t)
	if err := os.Chmod(filepath.Join(root, "cmd", "tool", "main.go"), 0o644); err != nil {
		t.Fatal(err)
	}
	descriptor := fixtureDescriptor(nil, ProfileGoV1)
	digest := treeDigestFromRejected(t, root, "source", descriptor)
	descriptor.Origin = OriginEvidence{
		Locator: "fixture://tree", ImmutableID: "tree-revision-1", LockRecord: "tree-lock-1",
		ChecksumSHA256: digest, Verified: true,
	}
	result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{
		Descriptor: descriptor, Root: root, VirtualRoot: "source",
	})
	if err != nil {
		t.Fatal(err)
	}

	fakeDigest := digestBytes([]byte("forged-tree-root"))
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{
			name: "root and origin digest drift",
			mutate: func(manifest *Manifest) {
				manifest.RawPayload.SHA256 = fakeDigest
				manifest.Origin.ChecksumSHA256 = fakeDigest
				for index := range manifest.RoleEvidence {
					if manifest.RoleEvidence[index].Key == "origin_checksum_sha256" {
						manifest.RoleEvidence[index].Value = fakeDigest
					}
				}
			},
		},
		{name: "total byte drift", mutate: func(manifest *Manifest) { manifest.RawPayload.Size++ }},
		{
			name: "executable bit drift",
			mutate: func(manifest *Manifest) {
				requireManifestNode(t, manifest, "source/cmd/tool/main.go").Mode |= 0o111
			},
		},
		{
			name: "entry digest drift",
			mutate: func(manifest *Manifest) {
				requireManifestNode(t, manifest, "source/LICENSE").SHA256 = digestBytes([]byte("other-license"))
			},
		},
		{
			name: "entry path drift",
			mutate: func(manifest *Manifest) {
				requireManifestNode(t, manifest, "source/LICENSE").Path = "source/NOTICE"
			},
		},
		{
			name: "entry kind drift",
			mutate: func(manifest *Manifest) {
				requireManifestNode(t, manifest, "source/LICENSE").Kind = NodeDirectory
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			manifest := cloneManifest(t, result.Manifest)
			testCase.mutate(&manifest)
			if err := decodeRehashedManifest(manifest); err == nil {
				t.Fatal("self-rehashed canonical-tree drift was accepted")
			}
		})
	}

	for _, field := range []string{"adapter", "manager", "package_name", "package_version"} {
		t.Run("empty "+field, func(t *testing.T) {
			manifest := cloneManifest(t, result.Manifest)
			switch field {
			case "adapter":
				manifest.AdapterID = ""
			case "manager":
				manifest.Manager = ""
			case "package_name":
				manifest.PackageName = ""
			case "package_version":
				manifest.PackageVersion = ""
			}
			if err := decodeRehashedManifest(manifest); err == nil {
				t.Fatal("manifest with an empty descriptor identity was accepted")
			}
		})
	}
}

func TestPinnedJVMClassResolvesCAFEBABEStructurally(t *testing.T) {
	payload := conformance.JVMClass()
	result, err := NewService().AdmitDependency(t.Context(), dependencyRequest("Fixture.class", payload, ProfileCommonV1))
	requireCode(t, err, CodeCompiledDependency)
	node := requireNode(t, result, "Fixture.class")
	if node.Class != ClassJVMBytecode || node.SelectedDetectorID != "jvm-class-v1" || !node.InspectionComplete {
		t.Fatalf("JVM classification = class %q detector %q complete %t", node.Class, node.SelectedDetectorID, node.InspectionComplete)
	}
	for _, observation := range node.Observations {
		if observation.DetectorID == "macho-v1" {
			t.Fatalf("invalid shared-magic Mach-O observation survived valid JVM resolution: %#v", observation)
		}
	}

	invalid := append([]byte(nil), payload...)
	constantPoolEnd := jvmConstantPoolEnd(t, invalid)
	// access_flags is followed by this_class. Zero is never a valid this_class
	// index and must be rejected by structural validation.
	binary.BigEndian.PutUint16(invalid[constantPoolEnd+2:constantPoolEnd+4], 0)
	invalidResult, invalidErr := NewService().AdmitDependency(t.Context(), dependencyRequest("Fixture.class", invalid, ProfileCommonV1))
	if invalidErr == nil {
		t.Fatal("JVM class with an invalid this_class index was admitted")
	}
	invalidNode := requireNode(t, invalidResult, "Fixture.class")
	if invalidNode.InspectionComplete || invalidNode.Class != ClassOpaqueUnknown {
		t.Fatalf("invalid JVM node = class %q complete %t", invalidNode.Class, invalidNode.InspectionComplete)
	}
	if !hasObservationResult(invalidNode, "jvm-class-v1", "ERROR") {
		t.Fatal("invalid JVM structure lacks the required JVM ERROR observation")
	}
}

func TestPinnedCompiledFixtureProvenanceAndIdentity(t *testing.T) {
	type fixture struct {
		payload       []byte
		sourcePath    string
		commandTokens []string
	}
	fixtures := map[string]fixture{
		"gnu-dynamic-pie": {
			payload: conformance.GNUDynamicPIE(), sourcePath: "conformance/testdata/gnu-pie.c",
			commandTokens: []string{"-fPIE", "-pie", "--dynamic-linker"},
		},
		"gnu-static-pie": {
			payload: conformance.GNUStaticPIE(), sourcePath: "conformance/testdata/gnu-pie.c",
			commandTokens: []string{"-fPIE", "-static-pie", "--no-dynamic-linker"},
		},
		"gnu-shared-object": {
			payload: conformance.GNUSharedObject(), sourcePath: "conformance/testdata/gnu-shared.c",
			commandTokens: []string{"-fPIC", "-shared", "-soname"},
		},
		"jvm-fixture-class": {
			payload: conformance.JVMClass(), sourcePath: "conformance/testdata/Fixture.java",
			commandTokens: []string{"javac", "--release 17", "-g:none"},
		},
	}
	evidence := conformance.PinnedFixtureEvidence()
	if len(evidence) != len(fixtures) {
		t.Fatalf("pinned evidence count = %d, want %d", len(evidence), len(fixtures))
	}
	seen := make(map[string]bool, len(evidence))
	for _, record := range evidence {
		fixture, ok := fixtures[record.ID]
		if !ok || seen[record.ID] {
			t.Fatalf("unexpected or duplicate pinned evidence ID %q", record.ID)
		}
		seen[record.ID] = true
		if record.Toolchain == "" {
			t.Fatalf("pinned evidence %q has no toolchain", record.ID)
		}
		if record.Size != int64(len(fixture.payload)) || record.PayloadSHA256 != digestBytes(fixture.payload) {
			t.Fatalf("pinned payload identity drift for %q", record.ID)
		}
		source, err := os.ReadFile(fixture.sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if record.SourceSHA256 != digestBytes(source) {
			t.Fatalf("pinned source identity drift for %q", record.ID)
		}
		for _, token := range fixture.commandTokens {
			if !strings.Contains(record.Command, token) {
				t.Fatalf("pinned command for %q lacks %q: %q", record.ID, token, record.Command)
			}
		}
	}
}

func TestFatMachOCAFEBABEResolvesWithoutJVMError(t *testing.T) {
	payload := conformance.FatMachO(conformance.MachO(2))
	result, err := NewService().AdmitDependency(t.Context(), dependencyRequest("universal", payload, ProfileCommonV1))
	requireCode(t, err, CodeCompiledDependency)
	node := requireNode(t, result, "universal")
	if node.Class != ClassNativeExecutable || node.SelectedDetectorID != "macho-v1" || !node.InspectionComplete {
		t.Fatalf("fat Mach-O classification = class %q detector %q complete %t", node.Class, node.SelectedDetectorID, node.InspectionComplete)
	}
	if hasObservationResult(node, "jvm-class-v1", "ERROR") {
		t.Fatalf("valid fat Mach-O retained an invalid competing JVM error: %#v", node.Observations)
	}
}

func jvmConstantPoolEnd(t *testing.T, payload []byte) int {
	t.Helper()
	if len(payload) < 10 {
		t.Fatal("short JVM fixture")
	}
	count := int(binary.BigEndian.Uint16(payload[8:10]))
	offset := 10
	for index := 1; index < count; index++ {
		if offset >= len(payload) {
			t.Fatal("truncated JVM fixture constant pool")
		}
		tag := payload[offset]
		offset++
		switch tag {
		case 1:
			if offset+2 > len(payload) {
				t.Fatal("truncated JVM UTF-8 constant")
			}
			length := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
			offset += 2 + length
		case 3, 4:
			offset += 4
		case 5, 6:
			offset += 8
			index++
		case 7, 8, 16, 19, 20:
			offset += 2
		case 9, 10, 11, 12, 17, 18:
			offset += 4
		case 15:
			offset += 3
		default:
			t.Fatalf("unsupported JVM fixture tag %d", tag)
		}
		if offset > len(payload) {
			t.Fatal("truncated JVM fixture")
		}
	}
	return offset
}

func hasObservationResult(node ManifestNode, detector, result string) bool {
	for _, observation := range node.Observations {
		if observation.DetectorID == detector && observation.Result == result {
			return true
		}
	}
	return false
}

func TestTarPhysicalMetadataIsCountedManifestedAndBound(t *testing.T) {
	longPAXName := strings.Repeat("p", 120) + ".go"
	paxPayload := append([]byte{}, paxRecord(t, "path", longPAXName)...)
	paxPayload = append(paxPayload, paxRecord(t, "mtime", "1.250")...)
	paxPayload = append(paxPayload, paxRecord(t, "uid", "42")...)
	paxArchive := rawTarArchive(
		rawTarEntry("PaxHeaders/source.go", tar.TypeXHeader, paxPayload, false),
		rawTarEntry("source.go", tar.TypeReg, []byte("package fixture\n"), false),
	)
	paxResult, paxErr := admitDependency(t, "source.tar", paxArchive, ProfileCommonV1)
	if paxErr != nil {
		t.Fatal(paxErr)
	}
	paxMetadata := requireNode(t, paxResult, "source.tar!/$tar-metadata-000001-pax-local")
	if paxMetadata.Class != ClassTextMetadata || paxMetadata.Variant != "tar.metadata.pax-local" {
		t.Fatalf("PAX metadata class = %q/%q", paxMetadata.Class, paxMetadata.Variant)
	}
	paxSource := requireNode(t, paxResult, "source.tar!/"+longPAXName+"")
	if factValue(t, paxSource, "archive-tar-v1", "pax_path_present") != "true" ||
		factValue(t, paxSource, "archive-tar-v1", "mtime_present") != "true" ||
		factValue(t, paxSource, "archive-tar-v1", "uid_present") != "true" ||
		factValue(t, paxSource, "archive-tar-v1", "metadata_members") != "$tar-metadata-000001-pax-local" {
		t.Fatalf("PAX source metadata facts = %#v", paxSource.Observations)
	}
	if paxResult.Manifest.Accounting.EntryCount != paxResult.Manifest.Accounting.ManifestedEntryCount {
		t.Fatalf("PAX entry accounting = %+v", paxResult.Manifest.Accounting)
	}
	paxForged := cloneManifest(t, paxResult.Manifest)
	setObservationFact(t, requireManifestNode(t, &paxForged, "source.tar!/"+longPAXName), "archive-tar-v1", "metadata_members", "")
	if err := decodeRehashedManifest(paxForged); err == nil {
		t.Fatal("self-rehashed manifest detached a PAX metadata member from its resolved entry")
	}

	longGNUName := strings.Repeat("g", 120) + ".go"
	gnuArchive := rawTarArchive(
		rawTarEntry("././@LongLink", tar.TypeGNULongName, append([]byte(longGNUName), 0), true),
		rawTarEntry("placeholder.go", tar.TypeReg, []byte("package fixture\n"), true),
	)
	gnuResult, gnuErr := admitDependency(t, "source.tar", gnuArchive, ProfileCommonV1)
	if gnuErr != nil {
		t.Fatal(gnuErr)
	}
	gnuMetadata := requireNode(t, gnuResult, "source.tar!/$tar-metadata-000001-gnu-long-name")
	if gnuMetadata.Class != ClassTextMetadata || gnuMetadata.Variant != "tar.metadata.gnu-long-name" {
		t.Fatalf("GNU metadata class = %q/%q", gnuMetadata.Class, gnuMetadata.Variant)
	}
	gnuSource := requireNode(t, gnuResult, "source.tar!/"+longGNUName+"")
	if factValue(t, gnuSource, "archive-tar-v1", "gnu_long_name_present") != "true" ||
		factValue(t, gnuSource, "archive-tar-v1", "metadata_members") != "$tar-metadata-000001-gnu-long-name" {
		t.Fatalf("GNU source metadata facts = %#v", gnuSource.Observations)
	}
	if gnuResult.Manifest.Accounting.EntryCount != gnuResult.Manifest.Accounting.ManifestedEntryCount {
		t.Fatalf("GNU entry accounting = %+v", gnuResult.Manifest.Accounting)
	}
	gnuForged := cloneManifest(t, gnuResult.Manifest)
	setObservationFact(t, requireManifestNode(t, &gnuForged, "source.tar!/"+longGNUName), "archive-tar-v1", "metadata_members", "")
	if err := decodeRehashedManifest(gnuForged); err == nil {
		t.Fatal("self-rehashed manifest detached a GNU metadata member from its resolved entry")
	}
}

func TestTarMetadataRejectsLimitsMalformedRepeatedAndResolutionFeatures(t *testing.T) {
	t.Run("100001 physical metadata headers", func(t *testing.T) {
		header := rawTarEntry("GlobalHead.0", tar.TypeXGlobalHeader, nil, false)
		payload := make([]byte, 0, len(header)*100_001+1024)
		for index := 0; index < 100_001; index++ {
			payload = append(payload, header...)
		}
		payload = append(payload, make([]byte, 1024)...)
		result, err := admitDependency(t, "metadata.tar", payload, ProfileCommonV1)
		requireCode(t, err, CodeInspectionLimitExceeded)
		if result.Manifest.Diagnostics[0].LimitName != "max_entry_count" ||
			result.Manifest.Diagnostics[0].Observed != 100_001 {
			t.Fatalf("entry limit diagnostic = %+v", result.Manifest.Diagnostics[0])
		}
		if result.Manifest.Accounting.EntryCount != 100_001 {
			t.Fatalf("physical entry count = %d", result.Manifest.Accounting.EntryCount)
		}
	})

	tests := []struct {
		name    string
		code    DiagnosticCode
		entries [][]byte
	}{
		{
			name: "malformed PAX record", code: CodeArchiveInvalid,
			entries: [][]byte{rawTarEntry("PaxHeaders/source.go", tar.TypeXHeader, []byte("9 bad=x\n"), false)},
		},
		{
			name: "repeated local metadata", code: CodeArchiveInvalid,
			entries: [][]byte{
				rawTarEntry("PaxHeaders/one", tar.TypeXHeader, paxRecord(t, "comment", "one"), false),
				rawTarEntry("PaxHeaders/two", tar.TypeXHeader, paxRecord(t, "comment", "two"), false),
				rawTarEntry("source.go", tar.TypeReg, []byte("package fixture\n"), false),
			},
		},
		{
			name: "xattr execution metadata", code: CodeArchiveUnsafeEntry,
			entries: [][]byte{
				rawTarEntry("PaxHeaders/source.go", tar.TypeXHeader, paxRecord(t, "SCHILY.xattr.security.capability", "forged"), false),
				rawTarEntry("source.go", tar.TypeReg, []byte("package fixture\n"), false),
			},
		},
		{
			name: "conflicting path resolution", code: CodeArchiveInvalid,
			entries: [][]byte{
				rawTarEntry("PaxHeaders/source.go", tar.TypeXHeader, paxRecord(t, "path", "pax.go"), false),
				rawTarEntry("././@LongLink", tar.TypeGNULongName, append([]byte("gnu.go"), 0), true),
				rawTarEntry("source.go", tar.TypeReg, []byte("package fixture\n"), true),
			},
		},
		{
			name: "link metadata on regular file", code: CodeArchiveUnsupported,
			entries: [][]byte{
				rawTarEntry("PaxHeaders/source.go", tar.TypeXHeader, paxRecord(t, "linkpath", "target"), false),
				rawTarEntry("source.go", tar.TypeReg, []byte("package fixture\n"), false),
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := admitDependency(t, "metadata.tar", rawTarArchive(testCase.entries...), ProfileCommonV1)
			requireCode(t, err, testCase.code)
			requireDecision(t, result, DecisionReject)
			if testCase.name == "xattr execution metadata" {
				node := requireNode(t, result, "metadata.tar!/$tar-metadata-000001-pax-local")
				if factValue(t, node, "archive-tar-v1", "xattr_present") != "true" {
					t.Fatalf("xattr presence not retained: %#v", node.Observations)
				}
			}
		})
	}
}

func factValue(t *testing.T, node ManifestNode, detector, key string) string {
	t.Helper()
	for _, observation := range node.Observations {
		if observation.DetectorID != detector {
			continue
		}
		for _, fact := range observation.Facts {
			if fact.Key == key {
				return fact.Value
			}
		}
	}
	t.Fatalf("fact %s.%s absent from %q", detector, key, node.Path)
	return ""
}

func paxRecord(t *testing.T, key, value string) []byte {
	t.Helper()
	base := len(key) + len(value) + 3
	size := base + 1
	for {
		next := base + len(fmt.Sprint(size))
		if next == size {
			return []byte(fmt.Sprintf("%d %s=%s\n", size, key, value))
		}
		size = next
	}
}

func rawTarArchive(entries ...[]byte) []byte {
	var payload []byte
	for _, entry := range entries {
		payload = append(payload, entry...)
	}
	return append(payload, make([]byte, 1024)...)
}

func rawTarEntry(name string, typeflag byte, payload []byte, gnu bool) []byte {
	header := make([]byte, 512)
	copy(header[:100], name)
	writeTarOctal(header[100:108], 0o644)
	writeTarOctal(header[108:116], 0)
	writeTarOctal(header[116:124], 0)
	writeTarOctal(header[124:136], int64(len(payload)))
	writeTarOctal(header[136:148], 0)
	for index := 148; index < 156; index++ {
		header[index] = ' '
	}
	header[156] = typeflag
	if gnu {
		copy(header[257:263], "ustar ")
		copy(header[263:265], " \x00")
	} else {
		copy(header[257:263], "ustar\x00")
		copy(header[263:265], "00")
	}
	checksum := int64(0)
	for _, value := range header {
		checksum += int64(value)
	}
	copy(header[148:156], fmt.Sprintf("%06o\x00 ", checksum))
	entry := append(header, payload...)
	if remainder := len(payload) % 512; remainder != 0 {
		entry = append(entry, make([]byte, 512-remainder)...)
	}
	return entry
}

func writeTarOctal(field []byte, value int64) {
	encoded := fmt.Sprintf("%0*o", len(field)-1, value)
	copy(field, encoded)
	field[len(field)-1] = 0
}

func TestELFDuplicateUseEdgesPrecedeSharedObjectResolution(t *testing.T) {
	for _, hasSoname := range []bool{false, true} {
		name := "without-soname"
		if hasSoname {
			name = "with-soname"
		}
		t.Run(name, func(t *testing.T) {
			payload := makeELF64(elfTypeDyn, false, false, hasSoname)
			descriptor := fixtureDescriptor(payload, ProfileCommonV1)
			descriptor.ResolvedUses = map[string][]UseEdge{
				"libcase": {
					{Kind: UseLinkOrLoad, Origin: "active_graph.edges[0]"},
					{Kind: UseLinkOrLoad, Origin: "active_graph.edges[1]"},
				},
			}
			result, err := NewService().AdmitDependency(t.Context(), DependencyRequest{
				Descriptor: descriptor,
				Payload:    Payload{Path: "libcase", Size: int64(len(payload)), Reader: bytes.NewReader(payload)},
			})
			requireCode(t, err, CodeCompiledDependency)
			node := requireNode(t, result, "libcase")
			if node.Class != ClassELFETDYNAmbiguous || node.Variant != "duplicate_use_edges" {
				t.Fatalf("duplicate-use class = %q/%q", node.Class, node.Variant)
			}

			manifest := cloneManifest(t, result.Manifest)
			forged := requireManifestNode(t, &manifest, "libcase")
			forged.Class = ClassNativeLibraryDynamic
			forged.Variant = "elf.shared_object"
			forged.Rule = "class_role_decision"
			syncClassificationDiagnostics(t, &manifest)
			if err := decodeRehashedManifest(manifest); err == nil {
				t.Fatal("self-rehashed shared-object resolution bypassed duplicate-use semantics")
			}
		})
	}
}

func requireManifestNode(t *testing.T, manifest *Manifest, path string) *ManifestNode {
	t.Helper()
	for index := range manifest.Nodes {
		if manifest.Nodes[index].Path == path {
			return &manifest.Nodes[index]
		}
	}
	t.Fatalf("manifest node %q not found", path)
	return nil
}
