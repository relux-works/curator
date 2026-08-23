package artifactpolicy

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1" // #nosec G505 -- constructs a format-correct DEX fixture.
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/artifactpolicy/conformance"
)

type zipFixtureEntry struct {
	name    string
	content []byte
	mode    fs.FileMode
	method  uint16
	nonUTF8 bool
}

type tarFixtureEntry struct {
	name     string
	content  []byte
	typeflag byte
	mode     int64
	linkname string
	pax      map[string]string
}

func fixtureDescriptor(data []byte, profile ProfileID) Descriptor {
	return Descriptor{
		AdapterID: "fixture-adapter-v1", ProfileID: profile, Manager: "fixture-manager-v1",
		PackageName: "fixture-package", PackageVersion: "1.0.0",
		Origin: OriginEvidence{
			Locator: "fixture://package/1.0.0", ImmutableID: "fixture-revision-1",
			LockRecord: "fixture-lock-record-1", ChecksumSHA256: digestBytes(data), Verified: true,
		},
	}
}

func dependencyRequest(path string, data []byte, profile ProfileID) DependencyRequest {
	return DependencyRequest{
		Descriptor: fixtureDescriptor(data, profile),
		Payload:    Payload{Path: path, Size: int64(len(data)), Reader: bytes.NewReader(data)},
	}
}

func admitDependency(t *testing.T, path string, data []byte, profile ProfileID) (Result, error) {
	t.Helper()
	return NewService().AdmitDependency(t.Context(), dependencyRequest(path, data, profile))
}

func validToolchainAuthorization(t *testing.T, path string, payload []byte) ToolchainAuthorization {
	t.Helper()
	root := "/curator-test-toolchain"
	if runtime.GOOS == "windows" {
		root = `C:\curator-test-toolchain`
	}
	fingerprint := digestBytes([]byte("trusted-toolchain-tree"))
	record := toolchainAuthorizationRecord{
		seal:                        managerAuthorizationSeal,
		policySelector:              "fixture-toolchain-policy-v1",
		resolvedRoot:                root,
		executableRelativePath:      path,
		environmentSearchResolution: filepath.Join(root, filepath.FromSlash(path)),
		version:                     "fixture toolchain 1.0",
		platform:                    "fixture-platform-v1",
		fingerprintAlgorithm:        toolchainFingerprintAlgorithm,
		checkpointFingerprintSHA256: fingerprint,
		timeOfUseFingerprintSHA256:  fingerprint,
		payloadPath:                 path,
		payloadSHA256:               digestBytes(payload),
		payloadSize:                 int64(len(payload)),
		outsideDependencyClosure:    true,
		containedLinksValidated:     true,
		ordinaryNodesValidated:      true,
	}
	if code, reason := validateToolchainRecord(record, managerAuthorizationSeal); code != "" {
		t.Fatalf("invalid test-only central-manager receipt: %s: %s", code, reason)
	}
	return sealedToolchainAuthorization{record: record}
}

func validLocalOutputAuthorization(t *testing.T, path string, payload []byte, class ArtifactClass) LocalOutputAuthorization {
	t.Helper()
	digest := digestBytes(payload)
	record := localOutputAuthorizationRecord{
		seal:                            managerAuthorizationSeal,
		sourceClosureDigest:             digestBytes([]byte("source-closure")),
		artifactManifestDigest:          digestBytes([]byte("input-artifact-manifest")),
		buildPlanDigest:                 digestBytes([]byte("build-plan")),
		declaredActionID:                "compile-main",
		executionReceiptSHA256:          digestBytes([]byte("execution-receipt")),
		protectedReceiptSHA256:          digestBytes([]byte("protected-receipt")),
		protectedStoreIdentity:          "protected-store-v1",
		stagingRootIdentity:             "staging-root-v1",
		payloadPath:                     path,
		payloadSHA256:                   digest,
		payloadSize:                     int64(len(payload)),
		expectation:                     ArtifactExpectation{Path: path, Class: class, SHA256: digest, Size: int64(len(payload))},
		stagingStartedEmpty:             true,
		observedProduction:              true,
		writeSetMatched:                 true,
		preexistingInputExcluded:        true,
		hardlinkSourceExcluded:          true,
		expectationIndependentlyDerived: true,
		completeInputMatched:            true,
		protectedPublicationValidated:   true,
	}
	if code, reason := validateLocalOutputRecord(record, managerAuthorizationSeal); code != "" {
		t.Fatalf("invalid test-only protected-executor receipt: %s: %s", code, reason)
	}
	return sealedLocalOutputAuthorization{record: record}
}

func requireCode(t *testing.T, err error, want DiagnosticCode) {
	t.Helper()
	if got := ErrorCode(err); got != want {
		t.Fatalf("error code = %q (%v), want %q", got, err, want)
	}
}

func requireDecision(t *testing.T, result Result, want Decision) {
	t.Helper()
	if result.Manifest.Decision != want {
		t.Fatalf("manifest decision = %q, want %q; diagnostics=%+v", result.Manifest.Decision, want, result.Manifest.Diagnostics)
	}
}

func requireNode(t *testing.T, result Result, path string) ManifestNode {
	t.Helper()
	for _, node := range result.Manifest.Nodes {
		if node.Path == path {
			return node
		}
	}
	t.Fatalf("manifest node %q absent; nodes=%+v", path, result.Manifest.Nodes)
	return ManifestNode{}
}

func buildZIP(t *testing.T, entries []zipFixtureEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: entry.method, NonUTF8: entry.nonUTF8}
		header.Modified = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(entry.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildTar(t *testing.T, entries []tarFixtureEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for _, entry := range entries {
		typeflag := entry.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		mode := entry.mode
		if mode == 0 {
			mode = 0o644
		}
		size := int64(len(entry.content))
		if typeflag != tar.TypeReg && typeflag != 0 && typeflag != tar.TypeGNUSparse {
			size = 0
		}
		header := &tar.Header{
			Name: entry.name, Mode: mode, Size: size, Typeflag: typeflag,
			Linkname: entry.linkname, ModTime: time.Unix(0, 0).UTC(), Format: tar.FormatPAX,
			PAXRecords: entry.pax,
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if size > 0 {
			if _, err := writer.Write(entry.content); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildGZIP(t *testing.T, payload []byte, name string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buffer, gzip.BestSpeed)
	if err != nil {
		t.Fatal(err)
	}
	writer.Name = name
	writer.ModTime = time.Unix(0, 0).UTC()
	writer.OS = 255
	if _, err := writer.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildWheel(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var record strings.Builder
	for _, path := range paths {
		digest := sha256.Sum256(files[path])
		fmt.Fprintf(&record, "%s,sha256=%s,%d\n", path, base64.RawURLEncoding.EncodeToString(digest[:]), len(files[path]))
	}
	recordPath := "fixture-1.0.0.dist-info/RECORD"
	fmt.Fprintf(&record, "%s,,\n", recordPath)
	entries := make([]zipFixtureEntry, 0, len(files)+1)
	for _, path := range paths {
		entries = append(entries, zipFixtureEntry{name: path, content: files[path], method: zip.Store})
	}
	entries = append(entries, zipFixtureEntry{name: recordPath, content: []byte(record.String()), method: zip.Store})
	return buildZIP(t, entries)
}

func buildNestedZIP(t *testing.T, depth int, leafName string, leaf []byte) []byte {
	t.Helper()
	data := leaf
	name := leafName
	for level := 0; level < depth; level++ {
		data = buildZIP(t, []zipFixtureEntry{{name: name, content: data, method: zip.Store}})
		name = fmt.Sprintf("layer-%02d.zip", level)
	}
	return data
}

func makeELF64(eType uint16, pie, interpreter, soname bool) []byte {
	programCount := 0
	if interpreter {
		programCount++
	}
	if pie || soname {
		programCount++
	}
	headerSize := 64
	programSize := 56
	offset := headerSize + programCount*programSize
	interp := []byte("/lib64/ld-linux-x86-64.so.2\x00")
	interpOffset := 0
	if interpreter {
		interpOffset = offset
		offset += len(interp)
	}
	dynamicOffset := 0
	dynamicEntries := 0
	if pie {
		dynamicEntries++
	}
	if soname {
		dynamicEntries++
	}
	if dynamicEntries > 0 {
		dynamicEntries++
		dynamicOffset = offset
		offset += dynamicEntries * 16
	}
	payload := make([]byte, offset)
	copy(payload[:16], []byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0, 0})
	binary.LittleEndian.PutUint16(payload[16:18], eType)
	binary.LittleEndian.PutUint16(payload[18:20], 62)
	binary.LittleEndian.PutUint32(payload[20:24], 1)
	binary.LittleEndian.PutUint64(payload[24:32], 0x401000)
	if programCount > 0 {
		binary.LittleEndian.PutUint64(payload[32:40], uint64(headerSize))
	}
	binary.LittleEndian.PutUint16(payload[52:54], uint16(headerSize))
	if programCount > 0 {
		binary.LittleEndian.PutUint16(payload[54:56], uint16(programSize))
	}
	binary.LittleEndian.PutUint16(payload[56:58], uint16(programCount))
	programOffset := headerSize
	if interpreter {
		program := payload[programOffset : programOffset+programSize]
		binary.LittleEndian.PutUint32(program[0:4], elfPTInterp)
		binary.LittleEndian.PutUint32(program[4:8], 4)
		binary.LittleEndian.PutUint64(program[8:16], uint64(interpOffset))
		binary.LittleEndian.PutUint64(program[32:40], uint64(len(interp)))
		binary.LittleEndian.PutUint64(program[40:48], uint64(len(interp)))
		binary.LittleEndian.PutUint64(program[48:56], 1)
		copy(payload[interpOffset:], interp)
		programOffset += programSize
	}
	if dynamicEntries > 0 {
		program := payload[programOffset : programOffset+programSize]
		binary.LittleEndian.PutUint32(program[0:4], elfPTDynamic)
		binary.LittleEndian.PutUint32(program[4:8], 4)
		binary.LittleEndian.PutUint64(program[8:16], uint64(dynamicOffset))
		binary.LittleEndian.PutUint64(program[32:40], uint64(dynamicEntries*16))
		binary.LittleEndian.PutUint64(program[40:48], uint64(dynamicEntries*16))
		binary.LittleEndian.PutUint64(program[48:56], 8)
		entryOffset := dynamicOffset
		if pie {
			binary.LittleEndian.PutUint64(payload[entryOffset:entryOffset+8], elfDTFlags1)
			binary.LittleEndian.PutUint64(payload[entryOffset+8:entryOffset+16], elfDF1PIE)
			entryOffset += 16
		}
		if soname {
			binary.LittleEndian.PutUint64(payload[entryOffset:entryOffset+8], elfDTSoname)
			binary.LittleEndian.PutUint64(payload[entryOffset+8:entryOffset+16], 1)
		}
	}
	return payload
}

func makePE(dll bool) []byte {
	const peOffset = 64
	const optionalSize = 112
	payload := make([]byte, peOffset+24+optionalSize+40)
	copy(payload[:2], "MZ")
	binary.LittleEndian.PutUint32(payload[0x3c:0x40], peOffset)
	copy(payload[peOffset:peOffset+4], "PE\x00\x00")
	header := payload[peOffset+4 : peOffset+24]
	binary.LittleEndian.PutUint16(header[0:2], 0x8664)
	binary.LittleEndian.PutUint16(header[2:4], 1)
	binary.LittleEndian.PutUint16(header[16:18], optionalSize)
	if dll {
		binary.LittleEndian.PutUint16(header[18:20], 0x2000)
	}
	binary.LittleEndian.PutUint16(payload[peOffset+24:peOffset+26], 0x20b)
	return payload
}

func makeCOFFObject() []byte {
	payload := make([]byte, 60)
	binary.LittleEndian.PutUint16(payload[0:2], 0x8664)
	binary.LittleEndian.PutUint16(payload[2:4], 1)
	return payload
}

func makeMachO(fileType uint32) []byte {
	payload := make([]byte, 32)
	binary.BigEndian.PutUint32(payload[0:4], 0xcffaedfe)
	binary.LittleEndian.PutUint32(payload[4:8], 0x01000007)
	binary.LittleEndian.PutUint32(payload[8:12], 3)
	binary.LittleEndian.PutUint32(payload[12:16], fileType)
	return payload
}

func makeFatMachO(slice []byte) []byte {
	return makeFatMachOSlices(slice)
}

func makeFatMachOSlices(slices ...[]byte) []byte {
	tableEnd := 8 + 20*len(slices)
	total := tableEnd
	for _, slice := range slices {
		total += len(slice)
	}
	payload := make([]byte, total)
	binary.BigEndian.PutUint32(payload[0:4], 0xcafebabe)
	binary.BigEndian.PutUint32(payload[4:8], uint32(len(slices))) // #nosec G115 -- test fixtures contain a bounded slice list.
	offset := tableEnd
	for index, slice := range slices {
		entry := payload[8+index*20 : 8+(index+1)*20]
		binary.BigEndian.PutUint32(entry[0:4], uint32(0x01000007+index)) // #nosec G115 -- test fixture index is bounded.
		binary.BigEndian.PutUint32(entry[4:8], 3)
		binary.BigEndian.PutUint32(entry[8:12], uint32(offset))      // #nosec G115 -- test fixture sizes are tiny.
		binary.BigEndian.PutUint32(entry[12:16], uint32(len(slice))) // #nosec G115 -- test fixture sizes are tiny.
		copy(payload[offset:], slice)
		offset += len(slice)
	}
	return payload
}

func makeJVMClass() []byte {
	return conformance.JVMClass()
}

func makeDEX() []byte {
	payload := make([]byte, 112)
	copy(payload[:8], []byte{'d', 'e', 'x', '\n', '0', '3', '5', 0})
	binary.LittleEndian.PutUint32(payload[32:36], uint32(len(payload)))
	binary.LittleEndian.PutUint32(payload[36:40], 112)
	binary.LittleEndian.PutUint32(payload[40:44], 0x12345678)
	signature := sha1.Sum(payload[32:])
	copy(payload[12:32], signature[:])
	binary.LittleEndian.PutUint32(payload[8:12], adler32.Checksum(payload[12:]))
	return payload
}

func makeWasm() []byte {
	return []byte{0, 'a', 's', 'm', 1, 0, 0, 0}
}

func makeLLVMBitcode() []byte {
	payload := make([]byte, 16)
	copy(payload[:4], []byte{'B', 'C', 0xc0, 0xde})
	// ENTER_SUBBLOCK, block ID 8, code width 2, then one 32-bit body word.
	binary.LittleEndian.PutUint32(payload[4:8], 1|(8<<2)|(2<<10))
	binary.LittleEndian.PutUint32(payload[8:12], 1)
	return payload
}

func makeLLVMBitcodeWrapper() []byte {
	bitcode := makeLLVMBitcode()
	payload := make([]byte, 20+len(bitcode))
	copy(payload[:4], []byte{0xde, 0xc0, 0x17, 0x0b})
	binary.LittleEndian.PutUint32(payload[4:8], 1)
	binary.LittleEndian.PutUint32(payload[8:12], 20)
	binary.LittleEndian.PutUint32(payload[12:16], uint32(len(bitcode)))
	binary.LittleEndian.PutUint32(payload[16:20], 0xffffffff)
	copy(payload[20:], bitcode)
	return payload
}

func buildAR(t *testing.T, members map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	buffer.WriteString("!<arch>\n")
	names := make([]string, 0, len(members))
	for name := range members {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if len(name) > 15 || strings.Contains(name, " ") {
			t.Fatalf("test ar member name %q is not short-form", name)
		}
		content := members[name]
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name+"/", 0, 0, 0, "100644", len(content))
		if len(header) != 60 {
			t.Fatalf("ar header length = %d", len(header))
		}
		buffer.WriteString(header)
		buffer.Write(content)
		if len(content)%2 != 0 {
			buffer.WriteByte('\n')
		}
	}
	return buffer.Bytes()
}

type arFixtureEntry struct {
	name    string
	content []byte
}

func buildAROrdered(t *testing.T, members []arFixtureEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	buffer.WriteString("!<arch>\n")
	for _, member := range members {
		name := member.name
		if name != "/" && name != "/SYM64/" && name != "//" &&
			!strings.HasPrefix(name, "/") && !strings.HasPrefix(name, "#1/") &&
			!strings.HasSuffix(name, "/") {
			name += "/"
		}
		if len(name) > 16 || strings.Contains(name, " ") {
			t.Fatalf("test ar member name %q does not fit its raw header", name)
		}
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "100644", len(member.content))
		if len(header) != 60 {
			t.Fatalf("ar header length = %d", len(header))
		}
		buffer.WriteString(header)
		buffer.Write(member.content)
		if len(member.content)%2 != 0 {
			buffer.WriteByte('\n')
		}
	}
	return buffer.Bytes()
}

func buildZIP64(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	if len(name) > 0xffff {
		t.Fatal("ZIP64 fixture name is too long")
	}
	var buffer bytes.Buffer
	write16 := func(value uint16) { _ = binary.Write(&buffer, binary.LittleEndian, value) }
	write32 := func(value uint32) { _ = binary.Write(&buffer, binary.LittleEndian, value) }
	write64 := func(value uint64) { _ = binary.Write(&buffer, binary.LittleEndian, value) }
	checksum := crc32.ChecksumIEEE(content)
	size := uint64(len(content))

	write32(0x04034b50)
	write16(45)
	write16(0)
	write16(zip.Store)
	write16(0)
	write16(0)
	write32(checksum)
	write32(0xffffffff)
	write32(0xffffffff)
	write16(uint16(len(name)))
	write16(20)
	buffer.WriteString(name)
	write16(0x0001)
	write16(16)
	write64(size)
	write64(size)
	buffer.Write(content)

	centralOffset := uint64(buffer.Len())
	write32(0x02014b50)
	write16(3<<8 | 45)
	write16(45)
	write16(0)
	write16(zip.Store)
	write16(0)
	write16(0)
	write32(checksum)
	write32(0xffffffff)
	write32(0xffffffff)
	write16(uint16(len(name)))
	write16(28)
	write16(0)
	write16(0)
	write16(0)
	write32(uint32(0o100644) << 16)
	write32(0xffffffff)
	buffer.WriteString(name)
	write16(0x0001)
	write16(24)
	write64(size)
	write64(size)
	write64(0)
	centralSize := uint64(buffer.Len()) - centralOffset

	zip64Offset := uint64(buffer.Len())
	write32(0x06064b50)
	write64(44)
	write16(3<<8 | 45)
	write16(45)
	write32(0)
	write32(0)
	write64(1)
	write64(1)
	write64(centralSize)
	write64(centralOffset)
	write32(0x07064b50)
	write32(0)
	write64(zip64Offset)
	write32(1)
	write32(0x06054b50)
	write16(0)
	write16(0)
	write16(0xffff)
	write16(0xffff)
	write32(0xffffffff)
	write32(0xffffffff)
	write16(0)
	return buffer.Bytes()
}

func patchZIPEncrypted(t *testing.T, payload []byte) []byte {
	t.Helper()
	result := append([]byte(nil), payload...)
	for offset := 0; offset+10 <= len(result); {
		signature := binary.LittleEndian.Uint32(result[offset : offset+4])
		switch signature {
		case 0x04034b50:
			flags := binary.LittleEndian.Uint16(result[offset+6:offset+8]) | 1
			binary.LittleEndian.PutUint16(result[offset+6:offset+8], flags)
			nameLength := int(binary.LittleEndian.Uint16(result[offset+26 : offset+28]))
			extraLength := int(binary.LittleEndian.Uint16(result[offset+28 : offset+30]))
			compressed := int(binary.LittleEndian.Uint32(result[offset+18 : offset+22]))
			offset += 30 + nameLength + extraLength + compressed
		case 0x02014b50:
			flags := binary.LittleEndian.Uint16(result[offset+8:offset+10]) | 1
			binary.LittleEndian.PutUint16(result[offset+8:offset+10], flags)
			nameLength := int(binary.LittleEndian.Uint16(result[offset+28 : offset+30]))
			extraLength := int(binary.LittleEndian.Uint16(result[offset+30 : offset+32]))
			commentLength := int(binary.LittleEndian.Uint16(result[offset+32 : offset+34]))
			offset += 46 + nameLength + extraLength + commentLength
		case 0x06054b50:
			return result
		default:
			offset++
		}
	}
	t.Fatal("ZIP fixture had no complete end record")
	return nil
}

func corruptZIPBody(t *testing.T, payload []byte) []byte {
	t.Helper()
	result := append([]byte(nil), payload...)
	if len(result) < 31 || binary.LittleEndian.Uint32(result[:4]) != 0x04034b50 {
		t.Fatal("fixture is not a local ZIP entry")
	}
	nameLength := int(binary.LittleEndian.Uint16(result[26:28]))
	extraLength := int(binary.LittleEndian.Uint16(result[28:30]))
	offset := 30 + nameLength + extraLength
	if offset >= len(result) {
		t.Fatal("ZIP fixture has no body")
	}
	result[offset] ^= 0xff
	return result
}

func repeatReaderError(prefix []byte, err error) io.Reader {
	return &errorReader{payload: append([]byte(nil), prefix...), err: err}
}

type errorReader struct {
	payload []byte
	err     error
}

func (reader *errorReader) Read(target []byte) (int, error) {
	if len(reader.payload) > 0 {
		read := copy(target, reader.payload)
		reader.payload = reader.payload[read:]
		return read, nil
	}
	return 0, reader.err
}

func treeDigestFromRejected(t *testing.T, root, virtualRoot string, descriptor Descriptor) string {
	t.Helper()
	descriptor.Origin = OriginEvidence{}
	result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{
		Descriptor: descriptor, Root: root, VirtualRoot: virtualRoot,
	})
	requireCode(t, err, CodeOriginUnverified)
	return result.Manifest.RawPayload.SHA256
}
