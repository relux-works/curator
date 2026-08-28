package artifactpolicy

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestTarMetadataLimitsPrecedePayloadParsingAndHashing(t *testing.T) {
	const declaredSize = int64(65)
	tests := []struct {
		name      string
		typeflag  byte
		kind      string
		limitName string
		configure func(*LimitVector)
	}{
		{
			name: "PAX local leaf", typeflag: tar.TypeXHeader, kind: "pax-local",
			limitName: "max_single_leaf_bytes",
			configure: func(limits *LimitVector) { limits.MaxSingleLeafBytes = declaredSize - 1 },
		},
		{
			name: "GNU long name leaf", typeflag: tar.TypeGNULongName, kind: "gnu-long-name",
			limitName: "max_single_leaf_bytes",
			configure: func(limits *LimitVector) { limits.MaxSingleLeafBytes = declaredSize - 1 },
		},
		{
			name: "GNU long link leaf", typeflag: tar.TypeGNULongLink, kind: "gnu-long-link",
			limitName: "max_single_leaf_bytes",
			configure: func(limits *LimitVector) { limits.MaxSingleLeafBytes = declaredSize - 1 },
		},
		{
			name: "PAX global emitted bytes", typeflag: tar.TypeXGlobalHeader, kind: "pax-global",
			limitName: "max_total_emitted_bytes",
			configure: func(limits *LimitVector) {
				limits.MaxSingleLeafBytes = declaredSize
				limits.MaxTotalEmittedBytes = declaredSize - 1
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			limits := DefaultLimits()
			testCase.configure(&limits)
			// Invalid metadata makes precedence observable: parsing these bytes
			// first would produce artifact_archive_invalid instead of the closed
			// declared-size diagnostic.
			payload := rawTarArchive(rawTarEntry(
				"metadata", testCase.typeflag, bytes.Repeat([]byte{0xff}, int(declaredSize)), true,
			))
			item, store := captureMetadataLimitBlob(t, payload, limits)
			account, err := newLimitAccountant(limits, item.size)
			if err != nil {
				t.Fatal(err)
			}
			worker := &inspector{ctx: t.Context(), limits: limits, account: account, store: store}
			_, diagnostic := worker.scanTarPhysical(item, "metadata.tar", nil)
			if diagnostic == nil || diagnostic.Code != CodeInspectionLimitExceeded ||
				diagnostic.LimitName != testCase.limitName || diagnostic.Observed != declaredSize {
				t.Fatalf("metadata preflight diagnostic = %+v", diagnostic)
			}
			wantPath := fmt.Sprintf("metadata.tar!/$tar-metadata-000001-%s", testCase.kind)
			if diagnostic.Path != wantPath {
				t.Fatalf("metadata preflight path = %q, want %q", diagnostic.Path, wantPath)
			}
			accounting := account.snapshot()
			wantObservedLeaf := declaredSize
			if testCase.limitName == "max_total_emitted_bytes" {
				wantObservedLeaf = 0
			}
			if accounting.TotalEmittedBytes != 0 || accounting.MaxObservedLeafBytes != wantObservedLeaf {
				t.Fatalf("metadata preflight accounting = %+v", accounting)
			}
		})
	}
}

func TestDefaultOversizedPAXMetadataIsBoundedBeforePayloadAllocation(t *testing.T) {
	limits := DefaultLimits()
	declaredSize := limits.MaxSingleLeafBytes + 1
	header := rawTarDeclaredHeader("PaxHeaders/oversized", tar.TypeXHeader, declaredSize, false)
	file, err := os.CreateTemp(t.TempDir(), "oversized-pax-*.tar")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if _, err := file.WriteAt(header, 0); err != nil {
		t.Fatal(err)
	}
	paddedSize, ok := roundTarBlock(declaredSize)
	if !ok {
		t.Fatal("oversized PAX size did not round")
	}
	endOffset := int64(512) + paddedSize
	if _, err := file.WriteAt(make([]byte, 1024), endOffset); err != nil {
		t.Fatal(err)
	}
	item := blob{file: file, size: endOffset + 1024}
	account, err := newLimitAccountant(limits, item.size)
	if err != nil {
		t.Fatal(err)
	}
	worker := &inspector{ctx: t.Context(), limits: limits, account: account}

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, diagnostic := worker.scanTarPhysical(item, "oversized.tar", nil)
	runtime.ReadMemStats(&after)
	if diagnostic == nil || diagnostic.Code != CodeInspectionLimitExceeded ||
		diagnostic.LimitName != "max_single_leaf_bytes" || diagnostic.Observed != declaredSize {
		t.Fatalf("default oversized PAX diagnostic = %+v", diagnostic)
	}
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 32<<20 {
		t.Fatalf("oversized PAX preflight allocated %d bytes before rejection", allocated)
	}
}

func TestDefaultOversizedARStringTableIsBoundedBeforePayloadAllocation(t *testing.T) {
	limits := DefaultLimits()
	declaredSize := limits.MaxSingleLeafBytes + 1
	header := []byte(fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", "//", 0, 0, 0, "100644", declaredSize))
	if len(header) != 60 {
		t.Fatalf("ar header length = %d", len(header))
	}
	file, err := os.CreateTemp(t.TempDir(), "oversized-string-table-*.a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	physicalHeader := append([]byte("!<arch>\n"), header...)
	if _, err := file.Write(physicalHeader); err != nil {
		t.Fatal(err)
	}
	virtualSize := int64(len(physicalHeader)) + declaredSize + declaredSize%2
	item := blob{file: file, size: virtualSize}
	account, err := newLimitAccountant(limits, virtualSize)
	if err != nil {
		t.Fatal(err)
	}
	worker := &inspector{
		ctx: t.Context(), limits: limits, account: account,
		nodes: []ManifestNode{{
			Path: "libcase.a", Kind: NodeArchive, Class: ClassNativeLibraryStatic,
			Decision: DecisionDescend, InspectionComplete: true,
		}},
	}

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	rejected := worker.walkAR(0, item, 1)
	runtime.ReadMemStats(&after)
	if !rejected || worker.findings == nil || worker.findings.total != 1 {
		t.Fatalf("oversized ar string table rejection/findings = %t/%v", rejected, worker.findings)
	}
	diagnostic := worker.findings.recorded()[0]
	if diagnostic.Code != CodeInspectionLimitExceeded || diagnostic.LimitName != "max_single_leaf_bytes" ||
		diagnostic.Observed != declaredSize || diagnostic.Path != "libcase.a!/$ar-metadata/string-table-001" {
		t.Fatalf("default oversized ar diagnostic = %+v", diagnostic)
	}
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 32<<20 {
		t.Fatalf("oversized ar string-table preflight allocated %d bytes before rejection", allocated)
	}
}

func TestDefaultOversizedBSDARNameIsBoundedBeforePayloadAllocation(t *testing.T) {
	limits := DefaultLimits()
	declaredNameLength := limits.MaxSingleLeafBytes + 1
	nameField := fmt.Sprintf("#1/%d", declaredNameLength)
	header := []byte(fmt.Sprintf(
		"%-16s%-12d%-6d%-6d%-8s%-10d`\n", nameField, 0, 0, 0, "100644", declaredNameLength,
	))
	if len(header) != 60 {
		t.Fatalf("ar header length = %d", len(header))
	}
	file, err := os.CreateTemp(t.TempDir(), "oversized-bsd-name-*.a")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	physicalHeader := append([]byte("!<arch>\n"), header...)
	if _, err := file.Write(physicalHeader); err != nil {
		t.Fatal(err)
	}
	virtualSize := int64(len(physicalHeader)) + declaredNameLength + declaredNameLength%2
	item := blob{file: file, size: virtualSize}
	account, err := newLimitAccountant(limits, virtualSize)
	if err != nil {
		t.Fatal(err)
	}
	worker := &inspector{
		ctx: t.Context(), limits: limits, account: account,
		nodes: []ManifestNode{{
			Path: "libcase.a", Kind: NodeArchive, Class: ClassNativeLibraryStatic,
			Decision: DecisionDescend, InspectionComplete: true,
		}},
	}

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	rejected := worker.walkAR(0, item, 1)
	runtime.ReadMemStats(&after)
	if !rejected || worker.findings == nil || worker.findings.total != 1 {
		t.Fatalf("oversized BSD ar name rejection/findings = %t/%v", rejected, worker.findings)
	}
	diagnostic := worker.findings.recorded()[0]
	if diagnostic.Code != CodeInspectionLimitExceeded || diagnostic.LimitName != "max_single_leaf_bytes" ||
		diagnostic.Observed != declaredNameLength ||
		diagnostic.Path != "libcase.a!/$ar-metadata/bsd-extended-name-000001" {
		t.Fatalf("default oversized BSD ar name diagnostic = %+v", diagnostic)
	}
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 32<<20 {
		t.Fatalf("oversized BSD ar name preflight allocated %d bytes before rejection", allocated)
	}
}

func captureMetadataLimitBlob(t *testing.T, payload []byte, limits LimitVector) (blob, *blobStore) {
	t.Helper()
	store, err := newBlobStore()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.close)
	item, err := store.captureRoot(t.Context(), Payload{
		Path: "metadata", Size: int64(len(payload)), Reader: bytes.NewReader(payload),
	}, limits)
	if err != nil {
		t.Fatal(err)
	}
	return item, store
}

func rawTarDeclaredHeader(name string, typeflag byte, size int64, gnu bool) []byte {
	header := make([]byte, 512)
	copy(header[:100], name)
	writeTarOctal(header[100:108], 0o644)
	writeTarOctal(header[108:116], 0)
	writeTarOctal(header[116:124], 0)
	writeTarOctal(header[124:136], size)
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
	return header
}
