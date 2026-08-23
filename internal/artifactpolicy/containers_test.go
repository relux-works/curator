package artifactpolicy

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestGZIPStreamingBudgetStopsAtFirstClosedLimit(t *testing.T) {
	t.Run("high-ratio output stops at ratio budget", func(t *testing.T) {
		payload := bytes.Repeat([]byte{'A'}, 4<<10)
		envelope := buildGZIP(t, payload, "payload.bin")
		limits := DefaultLimits()
		limits.MaxSingleLeafBytes = 16 << 10
		limits.MaxTotalEmittedBytes = 16 << 10
		limits.MaxExpansionRatio = 1
		account, err := newLimitAccountant(limits, int64(len(envelope)))
		if err != nil {
			t.Fatal(err)
		}
		budget, err := account.streamBudget(int64(len(envelope)))
		if err != nil {
			t.Fatal(err)
		}
		captured, observed := captureBoundedGZIPForTest(t, envelope, budget.maximum)
		if captured.size != budget.maximum || observed != budget.maximum+1 {
			t.Fatalf("bounded capture size/observed = %d/%d, want %d/%d", captured.size, observed, budget.maximum, budget.maximum+1)
		}
		failure := budget.failure(observed)
		if failure.name != "max_expansion_ratio" || failure.observed != 2 {
			t.Fatalf("ratio failure = %+v", failure)
		}
	})

	t.Run("nested output stops at remaining aggregate budget", func(t *testing.T) {
		payload := bytes.Repeat([]byte{'B'}, 64)
		envelope := buildGZIP(t, payload, "nested.bin")
		limits := DefaultLimits()
		limits.MaxSingleLeafBytes = 100
		limits.MaxTotalEmittedBytes = 50
		limits.MaxExpansionRatio = 200
		account, err := newLimitAccountant(limits, 1_000)
		if err != nil {
			t.Fatal(err)
		}
		if err := account.addEmitted(47, 47); err != nil {
			t.Fatal(err)
		}
		budget, err := account.streamBudget(int64(len(envelope)))
		if err != nil {
			t.Fatal(err)
		}
		if budget.maximum != 3 {
			t.Fatalf("remaining stream budget = %d, want 3", budget.maximum)
		}
		captured, observed := captureBoundedGZIPForTest(t, envelope, budget.maximum)
		if captured.size != 3 || observed != 4 {
			t.Fatalf("bounded nested capture size/observed = %d/%d, want 3/4", captured.size, observed)
		}
		failure := budget.failure(observed)
		if failure.name != "max_total_emitted_bytes" || failure.observed != 51 {
			t.Fatalf("total failure = %+v", failure)
		}
	})

	t.Run("public API rejects a gzip bomb without spooling to the leaf ceiling", func(t *testing.T) {
		envelope := buildGZIP(t, bytes.Repeat([]byte{0}, 1<<20), "bomb.bin")
		result, err := admitDependency(t, "bomb.bin.gz", envelope, ProfileCommonV1)
		requireLimitDiagnostic(t, result, err, "max_expansion_ratio", 201)
		if result.Manifest.Accounting.UnmanifestedEmittedBytes == 0 ||
			result.Manifest.Accounting.UnmanifestedEmittedBytes > int64(len(envelope))*DefaultLimits().MaxExpansionRatio+1 {
			t.Fatalf("gzip bounded work evidence = %+v", result.Manifest.Accounting)
		}
		if len(result.Manifest.Nodes) != 1 {
			t.Fatalf("gzip limit materialized %d nodes, want root only", len(result.Manifest.Nodes))
		}
	})
}

func captureBoundedGZIPForTest(t *testing.T, envelope []byte, maximum int64) (blob, int64) {
	t.Helper()
	reader, err := gzip.NewReader(bytes.NewReader(envelope))
	if err != nil {
		t.Fatal(err)
	}
	reader.Multistream(false)
	store, err := newBlobStore()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.close)
	spool, err := store.newFile()
	if err != nil {
		t.Fatal(err)
	}
	captured, observed, exceeded, err := store.appendUnknownBounded(t.Context(), spool, reader, maximum)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if !exceeded {
		t.Fatal("bounded gzip fixture did not cross its output budget")
	}
	info, err := spool.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != maximum {
		t.Fatalf("spool retained %d bytes, want exactly bounded %d", info.Size(), maximum)
	}
	return captured, observed
}

func TestEntryLimitIsEnforcedByEveryWalkerDuringEnumeration(t *testing.T) {
	const overLimit = 100_001

	t.Run("ZIP central count is refused before member accumulation", func(t *testing.T) {
		payload := buildManyZIP(t, overLimit, func(index int) string {
			return "src/file-" + strconv.Itoa(index) + ".go"
		})
		result, err := admitDependency(t, "flood.zip", payload, ProfileGoV1)
		assertEntryLimitResult(t, result, err)
		if len(result.Manifest.Nodes) != 1 {
			t.Fatalf("ZIP entry limit materialized %d nodes, want only the root", len(result.Manifest.Nodes))
		}
	})

	t.Run("tar invalid entries count before path filtering", func(t *testing.T) {
		payload := buildManyTar(t, overLimit, func(index int) string {
			return "../invalid-" + strconv.Itoa(index)
		})
		result, err := admitDependency(t, "flood.tar", payload, ProfileCommonV1)
		assertEntryLimitResult(t, result, err)
		if len(result.Manifest.Nodes) != 1 {
			t.Fatalf("tar entry limit materialized %d nodes, want only the root", len(result.Manifest.Nodes))
		}
	})

	t.Run("native archive duplicate entries count before hashing", func(t *testing.T) {
		payload := buildManyAR(t, overLimit, "same.o")
		result, err := admitDependency(t, "flood.a", payload, ProfileCommonV1)
		assertEntryLimitResult(t, result, err)
		if len(result.Manifest.Nodes) != 1 {
			t.Fatalf("ar entry limit materialized %d nodes, want only the root", len(result.Manifest.Nodes))
		}
	})
}

func TestDirectoryEntryLimitStopsTheLiveWalker(t *testing.T) {
	if testing.Short() {
		t.Skip("creates the normative 100,001-entry live-directory fixture")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot create the invalid-name flood used to keep evidence bounded")
	}
	const overLimit = 100_001
	root := t.TempDir()
	for index := 0; index < overLimit; index++ {
		name := filepath.Join(root, fmt.Sprintf("invalid:%06d", index))
		file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	descriptor := fixtureDescriptor(nil, ProfileGoV1)
	first, firstErr := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{
		Descriptor: descriptor, Root: root, VirtualRoot: "tree",
	})
	if firstErr == nil {
		t.Fatal("over-limit directory unexpectedly admitted during identity probe")
	}
	descriptor.Origin = OriginEvidence{
		Locator: "fixture://large-tree", ImmutableID: "large-tree-v1", LockRecord: "large-tree-lock-v1",
		ChecksumSHA256: first.Manifest.RawPayload.SHA256, Verified: true,
	}
	result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{
		Descriptor: descriptor, Root: root, VirtualRoot: "tree",
	})
	assertEntryLimitResult(t, result, err)
	if got := int64(len(result.Manifest.Nodes)); got > DefaultLimits().MaxEntryCount {
		t.Fatalf("directory walker materialized %d nodes after the entry limit", got)
	}
}

func TestDuplicateFloodUsesBoundedRecordedFindings(t *testing.T) {
	const entries = 5_000
	payload := buildManyZIP(t, entries, func(int) string { return "duplicate.go" })
	result, err := admitDependency(t, "duplicates.zip", payload, ProfileGoV1)
	requireCode(t, err, CodeArchiveUnsafePath)
	if result.Manifest.Accounting.EntryCount != entries {
		t.Fatalf("duplicate entry count = %d, want %d", result.Manifest.Accounting.EntryCount, entries)
	}
	if result.Manifest.Findings.Total != entries-1 {
		t.Fatalf("complete duplicate finding count = %d, want %d", result.Manifest.Findings.Total, entries-1)
	}
	if result.Manifest.Findings.Recorded != DefaultLimits().MaxRecordedFindings ||
		int64(len(result.Manifest.Diagnostics)) != DefaultLimits().MaxRecordedFindings {
		t.Fatalf("bounded findings = %+v diagnostics=%d", result.Manifest.Findings, len(result.Manifest.Diagnostics))
	}
	if result.Manifest.Findings.Algorithm != findingsDigestAlgorithm || !sha256Identity.MatchString(result.Manifest.Findings.SHA256) {
		t.Fatalf("complete finding-set identity = %+v", result.Manifest.Findings)
	}
}

func assertEntryLimitResult(t *testing.T, result Result, err error) {
	t.Helper()
	requireCode(t, err, CodeInspectionLimitExceeded)
	requireDecision(t, result, DecisionReject)
	if result.Manifest.Accounting.EntryCount != 100_001 {
		t.Fatalf("entry accounting = %d, want 100001", result.Manifest.Accounting.EntryCount)
	}
	if result.Admission != nil {
		t.Fatal("entry-limit rejection returned an admission token")
	}
	if _, decodeErr := DecodeManifest(result.CanonicalBytes); decodeErr != nil {
		t.Fatalf("decode canonical entry-limit rejection: %v", decodeErr)
	}
	diagnostic := result.Manifest.Diagnostics[0]
	if diagnostic.LimitName != "max_entry_count" || diagnostic.Observed != 100_001 {
		t.Fatalf("entry-limit diagnostic = %+v", diagnostic)
	}
}

func buildManyZIP(t *testing.T, count int, name func(int) string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for index := 0; index < count; index++ {
		header := &zip.FileHeader{Name: name(index), Method: zip.Store}
		header.SetMode(0o644)
		if _, err := writer.CreateHeader(header); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildManyTar(t *testing.T, count int, name func(int) string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for index := 0; index < count; index++ {
		if err := writer.WriteHeader(&tar.Header{
			Name: name(index), Mode: 0o644, Typeflag: tar.TypeReg, Size: 0,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildManyAR(t *testing.T, count int, name string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	buffer.WriteString("!<arch>\n")
	for index := 0; index < count; index++ {
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", name+"/", 0, 0, 0, "100644", 0)
		if len(header) != 60 {
			t.Fatalf("ar header length = %d", len(header))
		}
		buffer.WriteString(header)
	}
	return buffer.Bytes()
}

func TestFailClosedVectorsF01ThroughF06(t *testing.T) {
	t.Run("F01 unsafe portable paths", func(t *testing.T) {
		invalid := []string{
			"/absolute", "C:/drive", `\\server\share`, `back\slash`, "a/../b",
			"nul\x00byte", "control\x01byte", strings.Repeat("a", 4_097),
			strings.Repeat("b", 256), "CON", "name:stream", "trailing.", "trailing ",
		}
		for _, name := range invalid {
			if _, err := ValidateVirtualPath(name); err == nil {
				t.Errorf("path %q admitted", name)
			}
		}
		for _, name := range []string{"/absolute.go", "C:/drive.go", `back\slash.go`, "a/../b.go", "control\x01.go"} {
			payload := buildZIP(t, []zipFixtureEntry{{name: name, content: []byte("package p\n"), method: zip.Store}})
			result, err := admitDependency(t, "unsafe.zip", payload, ProfileGoV1)
			requireCode(t, err, CodeArchiveUnsafePath)
			requireDecision(t, result, DecisionReject)
		}
	})

	t.Run("F02 duplicate and portable collisions", func(t *testing.T) {
		fixtures := [][]zipFixtureEntry{
			{
				{name: "a.go", content: []byte("package a\n"), method: zip.Store},
				{name: "a.go", content: []byte("package b\n"), method: zip.Store},
			},
			{
				{name: "Readme.txt", content: []byte("one\n"), method: zip.Store},
				{name: "README.txt", content: []byte("two\n"), method: zip.Store},
			},
			{{name: "trailing./file.go", content: []byte("package p\n"), method: zip.Store}},
			{{name: "e\u0301/file.go", content: []byte("package p\n"), method: zip.Store}},
		}
		for index, entries := range fixtures {
			payload := buildZIP(t, entries)
			result, err := admitDependency(t, "collision.zip", payload, ProfileCommonV1)
			requireCode(t, err, CodeArchiveUnsafePath)
			if result.Manifest.Findings.Total < 1 {
				t.Fatalf("fixture %d has no findings", index)
			}
		}
	})

	t.Run("F03 links and special entries", func(t *testing.T) {
		symlink := buildZIP(t, []zipFixtureEntry{{
			name: "link", content: []byte("target"), mode: os.ModeSymlink | 0o777, method: zip.Store,
		}})
		result, err := admitDependency(t, "link.zip", symlink, ProfileCommonV1)
		requireCode(t, err, CodeArchiveUnsafeEntry)
		if got := requireNode(t, result, "link.zip!/link").Class; got != ClassLink {
			t.Fatalf("ZIP link class = %q", got)
		}

		for name, typeflag := range map[string]byte{
			"hardlink.tar": tar.TypeLink,
			"fifo.tar":     tar.TypeFifo,
			"char.tar":     tar.TypeChar,
			"block.tar":    tar.TypeBlock,
		} {
			payload := buildTar(t, []tarFixtureEntry{{name: "unsafe", typeflag: typeflag, linkname: "target"}})
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeArchiveUnsafeEntry)
			requireDecision(t, result, DecisionReject)
		}

		externalExtent := buildTar(t, []tarFixtureEntry{{
			name: "unsafe", content: []byte("extent"),
			pax: map[string]string{"CURATOR.sparse.external_extent": "0,6"},
		}})
		result, err = admitDependency(t, "external-extent.tar", externalExtent, ProfileCommonV1)
		requireCode(t, err, CodeArchiveUnsafeEntry)
		requireDecision(t, result, DecisionReject)

		for name, mode := range map[string]os.FileMode{
			"device.zip": os.ModeDevice,
			"fifo.zip":   os.ModeNamedPipe,
			"socket.zip": os.ModeSocket,
		} {
			payload := buildZIP(t, []zipFixtureEntry{{name: "unsafe", mode: mode | 0o600, method: zip.Store}})
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeArchiveUnsafeEntry)
			requireDecision(t, result, DecisionReject)
		}
	})

	t.Run("F04 encrypted ZIP member", func(t *testing.T) {
		payload := buildZIP(t, []zipFixtureEntry{{name: "main.go", content: []byte("package main\n"), method: zip.Store}})
		payload = patchZIPEncrypted(t, payload)
		result, err := admitDependency(t, "encrypted.zip", payload, ProfileGoV1)
		requireCode(t, err, CodeArchiveEncrypted)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("F05 unsupported containers compression and multi-volume", func(t *testing.T) {
		fixtures := map[string][]byte{
			"case.7z":  {0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c, 0, 0},
			"case.xz":  {0xfd, 0x37, 0x7a, 0x58, 0x5a, 0, 0, 0},
			"case.bz2": {'B', 'Z', 'h', '9', 0},
			"case.zst": {0x28, 0xb5, 0x2f, 0xfd, 0},
			"case.iso": func() []byte {
				payload := make([]byte, 0x8006)
				copy(payload[0x8001:], "CD001")
				return payload
			}(),
			"case.dmg": func() []byte {
				payload := make([]byte, 512)
				copy(payload, "koly")
				return payload
			}(),
		}
		for name, payload := range fixtures {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeArchiveUnsupported)
			requireDecision(t, result, DecisionReject)
		}

		ordinary := buildZIP(t, []zipFixtureEntry{{name: "main.go", content: []byte("package main\n"), method: zip.Store}})
		unsupportedMethod := patchZIPMethod(t, ordinary, 99)
		result, err := admitDependency(t, "unsupported-method.zip", unsupportedMethod, ProfileGoV1)
		requireCode(t, err, CodeArchiveUnsupported)
		requireDecision(t, result, DecisionReject)

		multiVolume := patchZIPDisk(t, ordinary, 1)
		result, err = admitDependency(t, "split.zip", multiVolume, ProfileGoV1)
		requireCode(t, err, CodeArchiveUnsupported)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("F06 malformed checksum and trailing data", func(t *testing.T) {
		ordinary := buildZIP(t, []zipFixtureEntry{{name: "main.go", content: []byte("package main\n"), method: zip.Store}})
		fixtures := map[string][]byte{
			"truncated.zip": ordinary[:len(ordinary)-7],
			"crc.zip":       corruptZIPBody(t, ordinary),
			"trailing.zip":  append(append([]byte(nil), ordinary...), []byte("trailing")...),
		}
		for name, payload := range fixtures {
			result, err := admitDependency(t, name, payload, ProfileGoV1)
			requireCode(t, err, CodeArchiveInvalid)
			requireDecision(t, result, DecisionReject)
		}

		tarPayload := buildTar(t, []tarFixtureEntry{{name: "main.go", content: []byte("package main\n")}})
		tarPayload = append(tarPayload, []byte("not-zero-trailing-data")...)
		result, err := admitDependency(t, "trailing.tar", tarPayload, ProfileGoV1)
		requireCode(t, err, CodeArchiveInvalid)
		requireDecision(t, result, DecisionReject)
	})
}

func TestFailClosedVectorsF07ThroughF14(t *testing.T) {
	t.Run("F07 archive depth nine", func(t *testing.T) {
		payload := buildNestedZIP(t, 9, "main.go", []byte("package main\n"))
		result, err := admitDependency(t, "deep.zip", payload, ProfileGoV1)
		requireCode(t, err, CodeInspectionLimitExceeded)
		requireDecision(t, result, DecisionReject)
		if result.Manifest.Diagnostics[0].LimitName != "max_archive_depth" {
			t.Fatalf("limit = %+v", result.Manifest.Diagnostics[0])
		}
	})

	t.Run("F08 closed resource limits", func(t *testing.T) {
		limits := DefaultLimits()
		if _, err := newLimitAccountant(limits, limits.MaxRawPayloadBytes+1); limitName(err) != "max_raw_payload_bytes" {
			t.Fatalf("raw limit error = %v", err)
		}
		account, err := newLimitAccountant(limits, limits.MaxRawPayloadBytes)
		if err != nil {
			t.Fatal(err)
		}
		if err := account.checkLeaf(limits.MaxSingleLeafBytes + 1); limitName(err) != "max_single_leaf_bytes" {
			t.Fatalf("leaf limit error = %v", err)
		}
		if err := account.addContainer(limits.MaxArchiveDepth + 1); limitName(err) != "max_archive_depth" {
			t.Fatalf("depth limit error = %v", err)
		}
		if err := account.addEntry(limits.MaxEntryCount + 1); limitName(err) != "max_entry_count" {
			t.Fatalf("entry limit error = %v", err)
		}
		if err := account.addEmitted(201, 1); limitName(err) != "max_expansion_ratio" {
			t.Fatalf("ratio limit error = %v", err)
		}
		account, _ = newLimitAccountant(limits, limits.MaxRawPayloadBytes)
		if err := account.addEmitted(limits.MaxTotalEmittedBytes+1, limits.MaxTotalEmittedBytes+1); limitName(err) != "max_total_emitted_bytes" {
			t.Fatalf("total limit error = %v", err)
		}

		t.Run("public service refuses raw payload above 512 MiB before reading", func(t *testing.T) {
			request := dependencyRequest("oversized.bin", nil, ProfileCommonV1)
			request.Payload.Size = limits.MaxRawPayloadBytes + 1
			result, err := NewService().AdmitDependency(t.Context(), request)
			requireLimitDiagnostic(t, result, err, "max_raw_payload_bytes", limits.MaxRawPayloadBytes+1)
		})

		t.Run("public service refuses declared 257 MiB leaf before reading it", func(t *testing.T) {
			payload := buildZIP(t, []zipFixtureEntry{{name: "large.bin", method: zip.Store}})
			payload = patchZIPDeclaredSizes(t, payload, []uint32{uint32(limits.MaxSingleLeafBytes + 1)}, []uint32{1})
			result, err := admitDependency(t, "large.zip", payload, ProfileCommonV1)
			requireLimitDiagnostic(t, result, err, "max_single_leaf_bytes", limits.MaxSingleLeafBytes+1)
		})

		t.Run("public service refuses declared total above 2 GiB before reading members", func(t *testing.T) {
			entries := make([]zipFixtureEntry, 9)
			declared := make([]uint32, len(entries))
			compressed := make([]uint32, len(entries))
			for index := range entries {
				entries[index] = zipFixtureEntry{name: fmt.Sprintf("part-%02d.bin", index), method: zip.Store}
				declared[index] = uint32(limits.MaxSingleLeafBytes)
				compressed[index] = uint32(2 << 20)
			}
			payload := patchZIPDeclaredSizes(t, buildZIP(t, entries), declared, compressed)
			result, err := admitDependency(t, "total.zip", payload, ProfileCommonV1)
			requireLimitDiagnostic(t, result, err, "max_total_emitted_bytes", 9*limits.MaxSingleLeafBytes)
		})

		t.Run("public service refuses declared stream above 200 to 1", func(t *testing.T) {
			payload := buildZIP(t, []zipFixtureEntry{{name: "ratio.bin", method: zip.Store}})
			payload = patchZIPDeclaredSizes(t, payload, []uint32{201}, []uint32{1})
			result, err := admitDependency(t, "ratio.zip", payload, ProfileCommonV1)
			requireLimitDiagnostic(t, result, err, "max_expansion_ratio", 201)
		})

		t.Run("public service refuses the 1025th container", func(t *testing.T) {
			emptyZIP := buildZIP(t, nil)
			entries := make([]zipFixtureEntry, limits.MaxContainerCount)
			for index := range entries {
				entries[index] = zipFixtureEntry{
					name: fmt.Sprintf("archive-%04d.zip", index), content: emptyZIP, method: zip.Store,
				}
			}
			result, err := admitDependency(t, "containers.zip", buildZIP(t, entries), ProfileCommonV1)
			requireLimitDiagnostic(t, result, err, "max_container_count", limits.MaxContainerCount+1)
		})
	})

	t.Run("F09 read failure and cancellation", func(t *testing.T) {
		data := []byte("package main\n")
		request := dependencyRequest("main.go", data, ProfileGoV1)
		request.Payload.Reader = repeatReaderError(data[:4], io.ErrUnexpectedEOF)
		result, err := NewService().AdmitDependency(t.Context(), request)
		requireCode(t, err, CodeInspectionUnavailable)
		requireDecision(t, result, DecisionReject)
		if result.Admission != nil {
			t.Fatal("incomplete inspection returned a token")
		}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		request = dependencyRequest("main.go", data, ProfileGoV1)
		result, err = NewService().AdmitDependency(ctx, request)
		requireCode(t, err, CodeInspectionUnavailable)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("F10 deny-indicating name with text bytes", func(t *testing.T) {
		for _, name := range []string{"library.so", "addon.node", "module.wasm"} {
			result, err := admitDependency(t, name, []byte("this is ordinary text\n"), ProfileCommonV1)
			requireCode(t, err, CodeTypeAmbiguous)
			requireDecision(t, result, DecisionReject)
		}
		container := buildZIP(t, []zipFixtureEntry{{name: "main.go", content: []byte("package main\n"), method: zip.Store}})
		result, err := admitDependency(t, "library.a", container, ProfileGoV1)
		requireCode(t, err, CodeTypeAmbiguous)
		requireDecision(t, result, DecisionReject)
	})

	t.Run("F11 compiled bytes with text name", func(t *testing.T) {
		result, err := admitDependency(t, "looks-safe.txt", makeELF64(elfTypeExec, false, false, false), ProfileCommonV1)
		requireCode(t, err, CodeCompiledDependency)
		if requireNode(t, result, "looks-safe.txt").Class != ClassNativeExecutable {
			t.Fatal("compiled detector did not dominate text suffix")
		}
	})

	t.Run("F12 undeclared text and unknown binary", func(t *testing.T) {
		for name, payload := range map[string][]byte{
			"undeclared.blob": []byte("valid UTF-8 without a declared grammar\n"),
			"unknown.bin":     {0x12, 0x34, 0x00, 0xff, 0x55},
		} {
			result, err := admitDependency(t, name, payload, ProfileCommonV1)
			requireCode(t, err, CodeOpaqueDependency)
			requireDecision(t, result, DecisionReject)
		}
	})

	t.Run("F13 missing immutable origin", func(t *testing.T) {
		data := []byte("package main\nfunc main() {}\n")
		request := dependencyRequest("main.go", data, ProfileGoV1)
		request.Descriptor.Origin = OriginEvidence{}
		result, err := NewService().AdmitDependency(t.Context(), request)
		requireCode(t, err, CodeOriginUnverified)
		requireDecision(t, result, DecisionReject)
		if requireNode(t, result, "main.go").Class != ClassSourceAuthoredText {
			t.Fatal("origin failure prevented deterministic classification")
		}
	})

	t.Run("F14 archive enumeration order is canonical", func(t *testing.T) {
		first := buildZIP(t, []zipFixtureEntry{
			{name: "z.txt", content: []byte("z\n"), method: zip.Store},
			{name: "a.txt", content: []byte("a\n"), method: zip.Store},
		})
		second := buildZIP(t, []zipFixtureEntry{
			{name: "a.txt", content: []byte("a\n"), method: zip.Store},
			{name: "z.txt", content: []byte("z\n"), method: zip.Store},
		})
		firstResult, firstErr := admitDependency(t, "order.zip", first, ProfileCommonV1)
		secondResult, secondErr := admitDependency(t, "order.zip", second, ProfileCommonV1)
		if firstErr != nil || secondErr != nil {
			t.Fatalf("order permutations rejected: first=%v second=%v", firstErr, secondErr)
		}
		if firstResult.Manifest.Decision != secondResult.Manifest.Decision {
			t.Fatal("archive order changed final decision")
		}
		firstRepeat, repeatErr := admitDependency(t, "order.zip", first, ProfileCommonV1)
		if repeatErr != nil {
			t.Fatal(repeatErr)
		}
		if !bytes.Equal(firstResult.CanonicalBytes, firstRepeat.CanonicalBytes) ||
			firstResult.Manifest.ManifestDigest != firstRepeat.Manifest.ManifestDigest {
			t.Fatal("identical F14 archive bytes did not reproduce exact canonical manifest bytes and digest")
		}
		assertF14CaptureAndOrderIdentity(t, firstResult, secondResult)
	})
}

func requireLimitDiagnostic(t *testing.T, result Result, err error, limitNameValue string, observed int64) {
	t.Helper()
	requireCode(t, err, CodeInspectionLimitExceeded)
	requireDecision(t, result, DecisionReject)
	if result.Admission != nil {
		t.Fatal("limit rejection returned an authorization token")
	}
	for _, diagnostic := range result.Manifest.Diagnostics {
		if diagnostic.Code == CodeInspectionLimitExceeded && diagnostic.LimitName == limitNameValue {
			if diagnostic.Observed != observed {
				t.Fatalf("%s observed = %d, want %d", limitNameValue, diagnostic.Observed, observed)
			}
			return
		}
	}
	t.Fatalf("%s diagnostic absent: %+v", limitNameValue, result.Manifest.Diagnostics)
}

func patchZIPDeclaredSizes(t *testing.T, payload []byte, uncompressed, compressed []uint32) []byte {
	t.Helper()
	if len(uncompressed) != len(compressed) {
		t.Fatal("ZIP declared-size vectors differ in length")
	}
	result := append([]byte(nil), payload...)
	patchHeaders := func(signature []byte, compressedOffset, uncompressedOffset int) int {
		count := 0
		for search := 0; search < len(result); {
			relative := bytes.Index(result[search:], signature)
			if relative < 0 {
				break
			}
			offset := search + relative
			if count >= len(uncompressed) || offset+uncompressedOffset+4 > len(result) {
				t.Fatal("ZIP fixture header count or bounds are invalid")
			}
			binary.LittleEndian.PutUint32(result[offset+compressedOffset:offset+compressedOffset+4], compressed[count])
			binary.LittleEndian.PutUint32(result[offset+uncompressedOffset:offset+uncompressedOffset+4], uncompressed[count])
			count++
			search = offset + len(signature)
		}
		return count
	}
	if got := patchHeaders([]byte{'P', 'K', 3, 4}, 18, 22); got != len(uncompressed) {
		t.Fatalf("local ZIP headers = %d, want %d", got, len(uncompressed))
	}
	if got := patchHeaders([]byte{'P', 'K', 1, 2}, 20, 24); got != len(uncompressed) {
		t.Fatalf("central ZIP headers = %d, want %d", got, len(uncompressed))
	}
	return result
}

func TestWheelRecordAndCompressionEnvelopeValidation(t *testing.T) {
	wrongRecord := buildZIP(t, []zipFixtureEntry{
		{name: "fixture/__init__.py", content: []byte("VALUE = 1\n"), method: zip.Store},
		{name: "fixture-1.0.dist-info/RECORD", content: []byte("fixture/__init__.py,sha256=wrong,10\nfixture-1.0.dist-info/RECORD,,\n"), method: zip.Store},
	})
	result, err := admitDependency(t, "fixture-1.0-py3-none-any.whl", wrongRecord, ProfilePythonSourceV1)
	requireCode(t, err, CodeArchiveInvalid)
	requireDecision(t, result, DecisionReject)

	stream := buildGZIP(t, []byte("package main\n"), "main.go")
	multiple := append(append([]byte(nil), stream...), stream...)
	result, err = admitDependency(t, "source.gz", multiple, ProfileGoV1)
	requireCode(t, err, CodeArchiveInvalid)
	requireDecision(t, result, DecisionReject)

	result, err = admitDependency(t, "thin.a", []byte("!<thin>\n"), ProfileCommonV1)
	requireCode(t, err, CodeArchiveUnsafeEntry)
	requireDecision(t, result, DecisionReject)
}

func TestZIP64SourceTraversal(t *testing.T) {
	payload := buildZIP64(t, "src/main.go", []byte("package main\nfunc main() {}\n"))
	result, err := admitDependency(t, "source.zip", payload, ProfileGoV1)
	if err != nil {
		t.Fatal(err)
	}
	requireDecision(t, result, DecisionAdmitInput)
	leaf := requireNode(t, result, "source.zip!/src/main.go")
	if leaf.Class != ClassSourceAuthoredText || leaf.Decision != DecisionAdmitInput {
		t.Fatalf("ZIP64 source leaf = %+v", leaf)
	}
}

func TestNativeArchiveStructuralValidation(t *testing.T) {
	ordinary := buildAR(t, map[string][]byte{"odd.o": {1}})
	badMetadata := append([]byte(nil), ordinary...)
	copy(badMetadata[8+16:8+28], "not-a-time  ")
	result, err := admitDependency(t, "bad-metadata.a", badMetadata, ProfileCommonV1)
	requireCode(t, err, CodeArchiveInvalid)
	requireDecision(t, result, DecisionReject)

	badPadding := append([]byte(nil), ordinary...)
	badPadding[len(badPadding)-1] = 0
	result, err = admitDependency(t, "bad-padding.a", badPadding, ProfileCommonV1)
	requireCode(t, err, CodeArchiveInvalid)
	requireDecision(t, result, DecisionReject)
}

func limitName(err error) string {
	var limit *limitFailure
	if errors.As(err, &limit) {
		return limit.name
	}
	return ""
}

func patchZIPMethod(t *testing.T, payload []byte, method uint16) []byte {
	t.Helper()
	result := append([]byte(nil), payload...)
	local := bytes.Index(result, []byte{'P', 'K', 3, 4})
	central := bytes.Index(result, []byte{'P', 'K', 1, 2})
	if local < 0 || central < 0 {
		t.Fatal("ZIP fixture lacks local or central header")
	}
	binary.LittleEndian.PutUint16(result[local+8:local+10], method)
	binary.LittleEndian.PutUint16(result[central+10:central+12], method)
	return result
}

func patchZIPDisk(t *testing.T, payload []byte, disk uint16) []byte {
	t.Helper()
	result := append([]byte(nil), payload...)
	eocd := bytes.LastIndex(result, []byte{'P', 'K', 5, 6})
	if eocd < 0 {
		t.Fatal("ZIP fixture lacks end record")
	}
	binary.LittleEndian.PutUint16(result[eocd+4:eocd+6], disk)
	return result
}

func TestGZIPRootVirtualDirectoryRetainsStreamAccounting(t *testing.T) {
	payload := buildGZIP(t, buildTar(t, []tarFixtureEntry{{
		name: "package/index.js", content: []byte("module.exports = 1\n"),
	}}), "package.tar")
	result, err := admitDependency(t, "npm/package.tgz", payload, ProfileNodeV1)
	if err != nil {
		t.Fatal(err)
	}
	requireDecision(t, result, DecisionAdmitInput)
	requireNode(t, result, "npm/package.tgz!/package.tar!/package/index.js")
}

func TestNodeProfileRecognizesBindingGYPAsInspectableMetadata(t *testing.T) {
	result, err := admitDependency(t, "binding.gyp", []byte("{'targets': []}\n"), ProfileNodeV1)
	if err != nil {
		t.Fatal(err)
	}
	if node := requireNode(t, result, "binding.gyp"); node.Class != ClassTextMetadata {
		t.Fatalf("binding.gyp class = %q, want %q", node.Class, ClassTextMetadata)
	}
}
