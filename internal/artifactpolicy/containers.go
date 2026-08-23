package artifactpolicy

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type rawMember struct {
	name          string
	originalName  string
	kind          NodeKind
	mode          int64
	data          *blob
	open          func() (io.ReadCloser, error)
	declaredSize  int64
	compressed    int64
	charged       bool
	explicit      bool
	failureCode   DiagnosticCode
	failureReason string
	observations  []Observation
	preclassified bool
	class         ArtifactClass
	variant       string
	detectorID    string
	rule          string
	identity      string
}

type preparedMember struct {
	raw       rawMember
	path      VirtualPath
	logical   VirtualPath
	synthetic bool
}

type archiveFeatureError struct {
	code   DiagnosticCode
	reason string
}

func (err *archiveFeatureError) Error() string {
	return err.reason
}

func (inspector *inspector) walkContainer(nodeIndex int, item blob, format containerFormat, depth int64) bool {
	pathValue := inspector.nodes[nodeIndex].Path
	chain := inspector.nodes[nodeIndex].ContainerChain
	if err := inspector.account.addContainer(depth); err != nil {
		inspector.rejectNode(nodeIndex, diagnosticFromError(pathValue, chain, err))
		return true
	}
	var rejected bool
	switch format {
	case formatZIP:
		rejected = inspector.walkZIP(nodeIndex, item, depth)
	case formatTar:
		rejected = inspector.walkTar(nodeIndex, item, depth)
	case formatGZIP:
		rejected = inspector.walkGZIP(nodeIndex, item, depth)
	case formatAR:
		rejected = inspector.walkAR(nodeIndex, item, depth)
	default:
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveUnsupported, Path: pathValue,
			ContainerChain: append([]string(nil), chain...), Reason: "unsupported_container",
		})
		return true
	}
	if rejected {
		inspector.nodes[nodeIndex].Decision = DecisionReject
		inspector.nodes[nodeIndex].Rule = "member_rejected"
	}
	return rejected
}

func (inspector *inspector) walkZIP(nodeIndex int, item blob, depth int64) bool {
	containerPath := inspector.nodes[nodeIndex].Path
	chain := inspector.nodes[nodeIndex].ContainerChain
	entryCount, err := validateZIPEnvelope(item)
	if err != nil {
		code := CodeArchiveInvalid
		var feature *archiveFeatureError
		var limit *limitFailure
		if errorAs(err, &feature) {
			code = feature.code
		} else if errorAs(err, &limit) {
			inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, err))
			return true
		}
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: code, Path: containerPath, DetectorID: "archive-zip-v1",
			ContainerChain: append([]string(nil), chain...), Reason: err.Error(),
		})
		return true
	}
	if err := inspector.account.addEntry(entryCount); err != nil {
		inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, err))
		return true
	}
	reader, err := zip.NewReader(item.reader(), item.size)
	if err != nil {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-zip-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "zip_index:" + err.Error(),
		})
		return true
	}
	if int64(len(reader.File)) != entryCount {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-zip-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "zip_entry_count_mismatch",
		})
		return true
	}
	raw := make([]rawMember, 0, len(reader.File))
	for _, file := range reader.File {
		member := rawMember{
			name: file.Name, kind: NodeRegularFile, mode: int64(file.Mode()),
			declaredSize: int64(file.UncompressedSize64), compressed: int64(file.CompressedSize64), // #nosec G115 -- values above MaxInt64 fail the bounds check below.
			explicit: true, open: file.Open,
			observations: []Observation{{
				DetectorID: "archive-zip-v1", Result: "ENTRY",
				Facts: facts(map[string]any{
					"compressed_size": file.CompressedSize64, "crc32": file.CRC32,
					"flags": file.Flags, "method": file.Method,
					"uncompressed_size": file.UncompressedSize64,
				}),
			}},
		}
		if file.UncompressedSize64 > uint64(^uint64(0)>>1) || file.CompressedSize64 > uint64(^uint64(0)>>1) {
			member.failureCode = CodeInspectionLimitExceeded
			member.failureReason = "zip_size_outside_int64"
		}
		if file.NonUTF8 {
			member.failureCode = CodeArchiveUnsafePath
			member.failureReason = "zip_name_not_utf8"
		}
		if file.Flags&1 != 0 {
			member.failureCode = CodeArchiveEncrypted
			member.failureReason = "encrypted_zip_member"
		}
		if file.Method != zip.Store && file.Method != zip.Deflate {
			member.failureCode = CodeArchiveUnsupported
			member.failureReason = fmt.Sprintf("unsupported_zip_method_%d", file.Method)
		}
		mode := file.Mode()
		if strings.HasSuffix(file.Name, "/") {
			member.kind = NodeDirectory
			member.name = strings.TrimSuffix(file.Name, "/")
			member.open = nil
			member.declaredSize = 0
		} else if mode&fs.ModeSymlink != 0 {
			member.kind = NodeLink
			member.open = nil
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = "zip_symlink"
		} else if !mode.IsRegular() && mode.Type() != 0 {
			member.kind = NodeSpecial
			member.open = nil
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = "zip_special_entry"
		}
		raw = append(raw, member)
	}
	if diagnostic := inspector.preflightRawMembers(containerPath, chain, raw); diagnostic != nil {
		inspector.rejectNode(nodeIndex, *diagnostic)
		return true
	}
	spool, err := inspector.store.newFile()
	if err != nil {
		inspector.rejectNode(nodeIndex, unavailableDiagnostic(containerPath, chain, "create_zip_spool", err))
		return true
	}
	rejected := inspector.materializeZIPMembers(raw, spool)
	prepared, preparationRejected := inspector.prepareMembers(containerPath, chain, raw)
	rejected = rejected || preparationRejected
	if isWheelContainer(containerPath, inspector.descriptor.ProfileID) {
		for index := range prepared {
			fullPath := joinContainerPath(containerPath, prepared[index].path.Canonical)
			memberChain := append(append([]string(nil), chain...), containerPath)
			if diagnostic := inspector.materializeMember(&prepared[index], spool, "archive-zip-v1", fullPath, memberChain); diagnostic != nil {
				inspector.addDiagnostic(*diagnostic)
				rejected = true
			}
		}
		if diagnostic := validateWheelRecord(containerPath, chain, prepared); diagnostic != nil {
			inspector.addDiagnostic(*diagnostic)
			rejected = true
		}
	}
	return inspector.inspectMembers(nodeIndex, prepared, spool, depth, rejected)
}

func validateZIPEnvelope(item blob) (int64, error) {
	if item.size < 22 {
		return 0, fmt.Errorf("truncated ZIP end record")
	}
	if prefix, err := item.prefix(4); err != nil || !isZIPPrefix(prefix) {
		return 0, fmt.Errorf("ZIP has a prepended or invalid prefix")
	}
	tailSize := item.size
	if tailSize > 65_557 {
		tailSize = 65_557
	}
	tail, err := readExactAt(item, item.size-tailSize, tailSize)
	if err != nil {
		return 0, err
	}
	signature := []byte{'P', 'K', 5, 6}
	index := bytes.LastIndex(tail, signature)
	if index < 0 || index+22 > len(tail) {
		return 0, fmt.Errorf("ZIP end record is missing or truncated")
	}
	eocdOffset := item.size - tailSize + int64(index)
	eocd := tail[index:]
	commentLength := int(binary.LittleEndian.Uint16(eocd[20:22]))
	if index+22+commentLength != len(tail) {
		return 0, fmt.Errorf("ZIP has trailing data or an inconsistent comment length")
	}
	disk := binary.LittleEndian.Uint16(eocd[4:6])
	centralDisk := binary.LittleEndian.Uint16(eocd[6:8])
	entriesDisk := binary.LittleEndian.Uint16(eocd[8:10])
	entriesTotal := binary.LittleEndian.Uint16(eocd[10:12])
	centralSize32 := binary.LittleEndian.Uint32(eocd[12:16])
	centralOffset32 := binary.LittleEndian.Uint32(eocd[16:20])
	if disk != 0 || centralDisk != 0 || (entriesDisk != entriesTotal && entriesDisk != 0xffff) {
		return 0, &archiveFeatureError{code: CodeArchiveUnsupported, reason: "multi_volume_zip"}
	}
	zip64 := entriesDisk == 0xffff || entriesTotal == 0xffff || centralSize32 == 0xffffffff || centralOffset32 == 0xffffffff
	if !zip64 {
		centralOffset := int64(centralOffset32)
		centralSize := int64(centralSize32)
		if centralOffset > eocdOffset || centralSize > eocdOffset-centralOffset || centralOffset+centralSize != eocdOffset {
			return 0, fmt.Errorf("ZIP central directory range is inconsistent")
		}
		return int64(entriesTotal), nil
	}
	if eocdOffset < 20 {
		return 0, fmt.Errorf("ZIP64 locator is missing")
	}
	locator, err := readExactAt(item, eocdOffset-20, 20)
	if err != nil || !bytes.Equal(locator[:4], []byte{'P', 'K', 6, 7}) {
		return 0, fmt.Errorf("ZIP64 locator is invalid")
	}
	if binary.LittleEndian.Uint32(locator[4:8]) != 0 || binary.LittleEndian.Uint32(locator[16:20]) != 1 {
		return 0, &archiveFeatureError{code: CodeArchiveUnsupported, reason: "multi_volume_zip64"}
	}
	zip64Offset := binary.LittleEndian.Uint64(locator[8:16])
	if zip64Offset > uint64(item.size) || uint64(eocdOffset-20) < zip64Offset || uint64(eocdOffset-20)-zip64Offset < 56 {
		return 0, fmt.Errorf("ZIP64 end record is out of bounds")
	}
	header, err := readExactAt(item, int64(zip64Offset), 56) // #nosec G115 -- offset was bounded by item.size.
	if err != nil || !bytes.Equal(header[:4], []byte{'P', 'K', 6, 6}) {
		return 0, fmt.Errorf("ZIP64 end record is invalid")
	}
	recordSize := binary.LittleEndian.Uint64(header[4:12])
	if recordSize < 44 || zip64Offset+12+recordSize != uint64(eocdOffset-20) {
		return 0, fmt.Errorf("ZIP64 end record size is inconsistent")
	}
	if binary.LittleEndian.Uint32(header[16:20]) != 0 || binary.LittleEndian.Uint32(header[20:24]) != 0 {
		return 0, &archiveFeatureError{code: CodeArchiveUnsupported, reason: "multi_volume_zip64"}
	}
	centralSize := binary.LittleEndian.Uint64(header[40:48])
	centralOffset := binary.LittleEndian.Uint64(header[48:56])
	if centralOffset > zip64Offset || centralSize > zip64Offset-centralOffset || centralOffset+centralSize != zip64Offset {
		return 0, fmt.Errorf("ZIP64 central directory range is inconsistent")
	}
	entriesTotal64 := binary.LittleEndian.Uint64(header[32:40])
	if entriesTotal64 > uint64(^uint64(0)>>1) {
		return 0, &limitFailure{name: "max_entry_count", limit: DefaultLimits().MaxEntryCount, observed: int64(^uint64(0) >> 1)}
	}
	return int64(entriesTotal64), nil
}

func (inspector *inspector) walkTar(nodeIndex int, item blob, depth int64) bool {
	containerPath := inspector.nodes[nodeIndex].Path
	chain := inspector.nodes[nodeIndex].ContainerChain
	scan, scanDiagnostic := inspector.scanTarPhysical(item, containerPath, chain)
	if scanDiagnostic != nil {
		inspector.rejectNode(nodeIndex, *scanDiagnostic)
		return true
	}
	counted := &countingReader{reader: item.reader()}
	reader := tar.NewReader(counted)
	spool, err := inspector.store.newFile()
	if err != nil {
		inspector.rejectNode(nodeIndex, unavailableDiagnostic(containerPath, chain, "create_tar_spool", err))
		return true
	}
	raw := append([]rawMember(nil), scan.metadata...)
	logicalIndex := 0
	globalIndex := 0
	for {
		header, nextErr := reader.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "tar_header:" + nextErr.Error(),
			})
			return true
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			globalIndex++
			continue
		}
		if logicalIndex >= len(scan.logical) {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "tar_logical_header_count_exceeded_raw_scan",
			})
			return true
		}
		expected := scan.logical[logicalIndex]
		logicalIndex++
		if header.Name != expected.name || header.Linkname != expected.linkname ||
			header.Size != expected.size || normalizedTarType(header.Typeflag, header.Name) != expected.typeflag {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "tar_raw_and_logical_header_mismatch",
			})
			return true
		}
		entryFacts := map[string]any{
			"format": header.Format.String(), "mode": header.Mode,
			"size": header.Size, "typeflag": header.Typeflag,
		}
		physicalIndex := int64(0)
		if len(expected.metadataMembers) > 0 {
			physicalIndex = expected.physicalIndex
		}
		appendTarPresenceFacts(entryFacts, expected.presence, expected.metadataMembers, physicalIndex, expected.rawName)
		originalName := ""
		if expected.rawName != header.Name {
			originalName = expected.rawName
		}
		member := rawMember{
			name: header.Name, originalName: originalName, kind: NodeRegularFile, mode: header.Mode,
			declaredSize: header.Size, compressed: header.Size, explicit: true,
			observations: []Observation{{
				DetectorID: "archive-tar-v1", Result: "ENTRY",
				Facts: facts(entryFacts),
			}},
		}
		if sparseTarHeader(header) {
			member.kind = NodeSpecial
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = "tar_sparse_or_external_extent"
		}
		switch header.Typeflag {
		case tar.TypeReg, 0:
		case tar.TypeDir:
			member.kind = NodeDirectory
			member.name = strings.TrimSuffix(header.Name, "/")
		case tar.TypeSymlink, tar.TypeLink:
			member.kind = NodeLink
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = "tar_link"
		case tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
			member.kind = NodeSpecial
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = "tar_special_entry"
		default:
			member.kind = NodeSpecial
			member.failureCode = CodeArchiveUnsafeEntry
			member.failureReason = fmt.Sprintf("unsupported_tar_type_%d", header.Typeflag)
		}
		if _, pathErr := validateVirtualPath(member.name, inspector.limits); pathErr != nil {
			raw = append(raw, member)
			continue
		}
		if member.kind == NodeRegularFile {
			if limitErr := inspector.account.checkLeaf(header.Size); limitErr != nil {
				inspector.rejectNode(nodeIndex, diagnosticFromError(
					joinContainerPath(containerPath, member.name),
					append(append([]string(nil), chain...), containerPath), limitErr,
				))
				return true
			}
			if limitErr := inspector.account.preflightEmitted(header.Size, header.Size); limitErr != nil {
				inspector.rejectNode(nodeIndex, diagnosticFromError(
					joinContainerPath(containerPath, member.name),
					append(append([]string(nil), chain...), containerPath), limitErr,
				))
				return true
			}
			if header.Size < 0 {
				inspector.rejectNode(nodeIndex, Diagnostic{
					Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
					ContainerChain: append([]string(nil), chain...), Reason: "negative_tar_member_size",
				})
				return true
			}
			captured, captureErr := inspector.store.appendBlob(inspector.ctx, spool, reader, header.Size)
			if captureErr != nil {
				inspector.rejectNode(nodeIndex, containerReadDiagnostic(containerPath, chain, "archive-tar-v1", captureErr))
				return true
			}
			if accountErr := inspector.account.addEmitted(captured.size, header.Size); accountErr != nil {
				inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, accountErr))
				return true
			}
			member.data = &captured
			member.charged = true
		}
		raw = append(raw, member)
	}
	if logicalIndex != len(scan.logical) || globalIndex != scan.globalCount {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "tar_raw_and_logical_header_count_mismatch",
		})
		return true
	}
	if nonzero, readErr := nonzeroRemainder(counted.reader); readErr != nil || nonzero {
		reason := "tar_trailing_data"
		if readErr != nil {
			reason = "tar_trailing_read:" + readErr.Error()
		}
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-tar-v1",
			ContainerChain: append([]string(nil), chain...), Reason: reason,
		})
		return true
	}
	prepared, rejected := inspector.prepareMembers(containerPath, chain, raw)
	return inspector.inspectMembers(nodeIndex, prepared, spool, depth, rejected)
}

type tarMetadataPresence struct {
	atime, ctime, gid, gname, mtime, uid, uname     bool
	xattr, paxPath, paxLink, paxComment, paxCharset bool
	gnuLongName, gnuLongLink                        bool
}

type rawTarLogicalHeader struct {
	physicalIndex   int64
	rawName         string
	name            string
	linkname        string
	size            int64
	typeflag        byte
	presence        tarMetadataPresence
	metadataMembers []string
}

type rawTarScan struct {
	metadata    []rawMember
	logical     []rawTarLogicalHeader
	globalCount int
}

type pendingTarMetadata struct {
	pax          map[string]string
	gnuLongName  string
	gnuLongLink  string
	presence     tarMetadataPresence
	members      []string
	localPAXSeen bool
	gnuNameSeen  bool
	gnuLinkSeen  bool
}

func (inspector *inspector) scanTarPhysical(item blob, containerPath string, chain []string) (rawTarScan, *Diagnostic) {
	if diagnostic := inspector.preflightTarPhysicalHeaders(item, containerPath, chain); diagnostic != nil {
		return rawTarScan{}, diagnostic
	}
	result := rawTarScan{metadata: make([]rawMember, 0), logical: make([]rawTarLogicalHeader, 0)}
	pending := pendingTarMetadata{}
	offset := int64(0)
	physicalIndex := int64(0)
	zeroBlocks := 0
	for offset < item.size {
		if item.size-offset < 512 {
			return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "truncated_tar_header")
		}
		header, err := readExactAt(item, offset, 512)
		if err != nil {
			return rawTarScan{}, tarReadDiagnostic(containerPath, chain, err)
		}
		if allZero(header) {
			zeroBlocks++
			offset += 512
			if zeroBlocks < 2 {
				continue
			}
			if pending.localPAXSeen || pending.gnuNameSeen || pending.gnuLinkSeen {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_metadata_without_logical_member")
			}
			if nonzero, readErr := nonzeroBlobRemainder(item, offset); readErr != nil {
				return rawTarScan{}, tarReadDiagnostic(containerPath, chain, readErr)
			} else if nonzero {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_trailing_data")
			}
			return result, nil
		}
		if zeroBlocks != 0 {
			return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_single_zero_block_before_header")
		}
		if !validTarHeaderChecksum(header) {
			return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "invalid_tar_header_checksum")
		}
		physicalIndex++
		size, ok := parseTarNumber(header[124:136])
		if !ok {
			return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "invalid_tar_member_size")
		}
		payloadOffset := offset + 512
		paddedSize, ok := roundTarBlock(size)
		if !ok || payloadOffset > item.size || paddedSize > item.size-payloadOffset {
			return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_member_out_of_bounds")
		}
		payloadEnd := payloadOffset + size
		if paddedSize > size {
			padding, readErr := readExactAt(item, payloadEnd, paddedSize-size)
			if readErr != nil {
				return rawTarScan{}, tarReadDiagnostic(containerPath, chain, readErr)
			}
			if !allZero(padding) {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "nonzero_tar_padding")
			}
		}
		typeflag := header[156]
		rawName := rawTarHeaderName(header)
		rawLink := tarString(header[157:257])
		presence := tarHeaderPresence(header)
		switch typeflag {
		case tar.TypeXHeader, tar.TypeXGlobalHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
			kind := tarMetadataKind(typeflag)
			memberName := fmt.Sprintf("$tar-metadata-%06d-%s", physicalIndex, kind)
			memberPath := joinContainerPath(containerPath, memberName)
			memberChain := append(append([]string(nil), chain...), containerPath)
			if limitErr := inspector.account.preflightLeaf(size); limitErr != nil {
				_ = inspector.account.checkLeaf(size)
				diagnostic := diagnosticFromError(memberPath, memberChain, limitErr)
				diagnostic.Size = size
				return rawTarScan{}, &diagnostic
			}
			if limitErr := inspector.account.preflightEmitted(size, size); limitErr != nil {
				diagnostic := diagnosticFromError(memberPath, memberChain, limitErr)
				diagnostic.Size = size
				return rawTarScan{}, &diagnostic
			}
			parsedKind, metadataPresence, values, observed, metadataErr := parseTarMetadataBlob(
				inspector.ctx, item, payloadOffset, size, typeflag, containerPath, inspector.limits,
			)
			if metadataErr != nil {
				if observed > 0 {
					if limitErr := inspector.account.checkLeaf(observed); limitErr != nil {
						return rawTarScan{}, tarDiagnostic(containerPath, chain, CodePolicyInternalError, "tar_metadata_observed_leaf_drift")
					}
					if accountErr := inspector.account.addEmitted(observed, observed); accountErr != nil {
						diagnostic := diagnosticFromError(memberPath, memberChain, accountErr)
						return rawTarScan{}, &diagnostic
					}
				}
				diagnostic := tarDiagnostic(containerPath, chain, metadataErr.code, metadataErr.reason)
				if metadataErr.pathFailure != nil {
					diagnostic.Path = memberPath
					diagnostic.Code = CodeArchiveUnsafePath
					diagnostic.OriginalNameBase64 = originalNameBase64(rawName)
					diagnostic.ContainerChain = memberChain
					diagnostic.LimitName = metadataErr.pathFailure.limitName
					diagnostic.Limit = metadataErr.pathFailure.limit
					diagnostic.Observed = metadataErr.pathFailure.observed
				}
				return rawTarScan{}, diagnostic
			}
			if observed != size {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodePolicyInternalError, "tar_metadata_read_accounting_drift")
			}
			if limitErr := inspector.account.checkLeaf(size); limitErr != nil {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodePolicyInternalError, "tar_metadata_leaf_preflight_drift")
			}
			if accountErr := inspector.account.addEmitted(size, size); accountErr != nil {
				diagnostic := diagnosticFromError(memberPath, memberChain, accountErr)
				return rawTarScan{}, &diagnostic
			}
			if parsedKind != kind {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodePolicyInternalError, "tar_metadata_kind_drift")
			}
			mergeTarPresence(&presence, metadataPresence)
			memberBlob := blob{file: item.file, offset: item.offset + payloadOffset, size: size}
			memberDigest, readErr := hashBlob(inspector.ctx, memberBlob)
			if readErr != nil {
				return rawTarScan{}, tarReadDiagnostic(containerPath, chain, readErr)
			}
			memberBlob.sha256 = memberDigest
			entryFacts := map[string]any{
				"format": tarMetadataFormat(typeflag), "metadata_kind": kind,
				"mode": tarHeaderMode(header), "size": size, "typeflag": typeflag,
			}
			appendTarPresenceFacts(entryFacts, presence, nil, physicalIndex, rawName)
			metadataMember := rawMember{
				name: memberName, originalName: rawName, kind: NodeRegularFile,
				mode: tarHeaderMode(header), data: &memberBlob, declaredSize: size, compressed: size,
				charged: true, explicit: true, preclassified: true, class: ClassTextMetadata,
				variant: "tar.metadata." + kind, detectorID: "archive-tar-v1", rule: "tar_archive_metadata",
				observations: []Observation{{DetectorID: "archive-tar-v1", Result: "ENTRY", Facts: facts(entryFacts)}},
			}
			if metadataErr := applyTarMetadata(&pending, typeflag, values, presence, memberName); metadataErr != nil {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, metadataErr.code, metadataErr.reason)
			}
			if values.failureCode != "" {
				metadataMember.failureCode = values.failureCode
				metadataMember.failureReason = values.failureReason
			}
			result.metadata = append(result.metadata, metadataMember)
			if typeflag == tar.TypeXGlobalHeader {
				result.globalCount++
			}
		default:
			name := rawName
			linkname := rawLink
			if value := pending.pax["path"]; value != "" {
				name = value
			}
			if pending.gnuLongName != "" {
				name = pending.gnuLongName
			}
			if value := pending.pax["linkpath"]; value != "" {
				linkname = value
			}
			if pending.gnuLongLink != "" {
				linkname = pending.gnuLongLink
			}
			normalizedType := normalizedTarType(typeflag, name)
			if (pending.pax["linkpath"] != "" || pending.gnuLongLink != "") &&
				normalizedType != tar.TypeLink && normalizedType != tar.TypeSymlink {
				return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveUnsupported, "tar_link_metadata_on_nonlink_entry")
			}
			mergeTarPresence(&presence, pending.presence)
			result.logical = append(result.logical, rawTarLogicalHeader{
				physicalIndex: physicalIndex, rawName: rawName, name: name, linkname: linkname,
				size: size, typeflag: normalizedType, presence: presence,
				metadataMembers: append([]string(nil), pending.members...),
			})
			pending = pendingTarMetadata{}
		}
		offset = payloadOffset + paddedSize
	}
	return rawTarScan{}, tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_end_markers_missing")
}

// preflightTarPhysicalHeaders accounts every physical header before any PAX or
// GNU metadata can be merged by archive/tar and before member evidence is
// accumulated. Invalid headers are counted as encountered entries; the 100001st
// physical header therefore deterministically hits max_entry_count first.
func (inspector *inspector) preflightTarPhysicalHeaders(item blob, containerPath string, chain []string) *Diagnostic {
	offset := int64(0)
	zeroBlocks := 0
	for offset < item.size {
		if item.size-offset < 512 {
			return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "truncated_tar_header")
		}
		header, err := readExactAt(item, offset, 512)
		if err != nil {
			return tarReadDiagnostic(containerPath, chain, err)
		}
		if allZero(header) {
			zeroBlocks++
			offset += 512
			if zeroBlocks < 2 {
				continue
			}
			if nonzero, readErr := nonzeroBlobRemainder(item, offset); readErr != nil {
				return tarReadDiagnostic(containerPath, chain, readErr)
			} else if nonzero {
				return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_trailing_data")
			}
			return nil
		}
		if zeroBlocks != 0 {
			return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_single_zero_block_before_header")
		}
		if accountErr := inspector.account.addEntry(1); accountErr != nil {
			diagnostic := diagnosticFromError(containerPath, chain, accountErr)
			return &diagnostic
		}
		if !validTarHeaderChecksum(header) {
			return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "invalid_tar_header_checksum")
		}
		size, ok := parseTarNumber(header[124:136])
		if !ok {
			return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "invalid_tar_member_size")
		}
		paddedSize, ok := roundTarBlock(size)
		if !ok || offset+512 > item.size || paddedSize > item.size-(offset+512) {
			return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_member_out_of_bounds")
		}
		offset += 512 + paddedSize
	}
	return tarDiagnostic(containerPath, chain, CodeArchiveInvalid, "tar_end_markers_missing")
}

type tarMetadataValues struct {
	records       map[string]string
	text          string
	failureCode   DiagnosticCode
	failureReason string
}

type tarMetadataError struct {
	code        DiagnosticCode
	reason      string
	pathFailure *pathFailure
}

type tarMetadataCursor struct {
	ctx    context.Context
	item   blob
	base   int64
	size   int64
	offset int64
}

func (cursor *tarMetadataCursor) readByte() (byte, error) {
	if cursor.offset >= cursor.size {
		return 0, io.ErrUnexpectedEOF
	}
	if err := contextError(cursor.ctx); err != nil {
		return 0, err
	}
	var value [1]byte
	if _, err := cursor.item.readAt(value[:], cursor.base+cursor.offset); err != nil {
		return 0, err
	}
	cursor.offset++
	return value[0], nil
}

func (cursor *tarMetadataCursor) readBytes(size int64) ([]byte, error) {
	if size < 0 || size > cursor.size-cursor.offset {
		return nil, io.ErrUnexpectedEOF
	}
	if err := contextError(cursor.ctx); err != nil {
		return nil, err
	}
	payload := make([]byte, size)
	if size > 0 {
		if _, err := cursor.item.reader().ReadAt(payload, cursor.base+cursor.offset); err != nil && err != io.EOF {
			return nil, err
		}
	}
	cursor.offset += size
	return payload, nil
}

func (cursor *tarMetadataCursor) readUTF8(size int64, capture bool) (string, error) {
	if size < 0 || size > cursor.size-cursor.offset {
		return "", io.ErrUnexpectedEOF
	}
	var captured []byte
	if capture {
		captured = make([]byte, 0, size)
	}
	remaining := size
	carry := make([]byte, 0, utf8.UTFMax)
	buffer := make([]byte, 32<<10)
	for remaining > 0 {
		amount := int64(len(buffer))
		if amount > remaining {
			amount = remaining
		}
		chunk, err := cursor.readBytes(amount)
		if err != nil {
			return "", err
		}
		if capture {
			captured = append(captured, chunk...)
		}
		data := append([]byte(nil), carry...)
		data = append(data, chunk...)
		carry = carry[:0]
		for len(data) > 0 {
			if !utf8.FullRune(data) {
				carry = append(carry, data...)
				break
			}
			runeValue, width := utf8.DecodeRune(data)
			if runeValue == utf8.RuneError && width == 1 {
				return "", fmt.Errorf("metadata is not UTF-8")
			}
			for _, value := range data[:width] {
				if value == 0 {
					return "", fmt.Errorf("metadata contains NUL")
				}
			}
			data = data[width:]
		}
		remaining -= amount
	}
	if len(carry) != 0 {
		return "", fmt.Errorf("metadata is not UTF-8")
	}
	return string(captured), nil
}

func preflightContainerMemberPathLength(containerPath string, memberBytes int64, limits LimitVector) *pathFailure {
	prefix, ok := checkedAdd(int64(len(containerPath)), 2)
	if !ok {
		return &pathFailure{
			reason: "path_too_long", limitName: "max_path_bytes",
			limit: limits.MaxPathBytes, observed: math.MaxInt64,
		}
	}
	observed, ok := checkedAdd(prefix, memberBytes)
	if !ok {
		observed = math.MaxInt64
	}
	if observed > limits.MaxPathBytes {
		return &pathFailure{
			reason: "path_too_long", limitName: "max_path_bytes",
			limit: limits.MaxPathBytes, observed: observed,
		}
	}
	return nil
}

func parseTarMetadataBlob(
	ctx context.Context,
	item blob,
	payloadOffset int64,
	size int64,
	typeflag byte,
	containerPath string,
	limits LimitVector,
) (string, tarMetadataPresence, tarMetadataValues, int64, *tarMetadataError) {
	cursor := &tarMetadataCursor{ctx: ctx, item: item, base: payloadOffset, size: size}
	presence := tarMetadataPresence{}
	values := tarMetadataValues{}
	readFailure := func(err error) (string, tarMetadataPresence, tarMetadataValues, int64, *tarMetadataError) {
		code := CodeArchiveInvalid
		reason := "malformed_tar_metadata:" + err.Error()
		if contextErrorCode(err) {
			code = CodeInspectionUnavailable
			reason = "tar_metadata_read:" + err.Error()
		}
		return "", presence, values, cursor.offset, &tarMetadataError{code: code, reason: reason}
	}
	switch typeflag {
	case tar.TypeGNULongName, tar.TypeGNULongLink:
		if size < 2 {
			return readFailure(fmt.Errorf("GNU long value is too short"))
		}
		valueSize := size - 1
		if pathErr := preflightContainerMemberPathLength(containerPath, valueSize, limits); pathErr != nil {
			return "", presence, values, cursor.offset, &tarMetadataError{
				code: CodeArchiveUnsafePath, reason: pathErr.reason, pathFailure: pathErr,
			}
		}
		text, err := cursor.readUTF8(valueSize, true)
		if err != nil {
			return readFailure(err)
		}
		terminator, err := cursor.readByte()
		if err != nil {
			return readFailure(err)
		}
		if terminator != 0 || text == "" {
			return readFailure(fmt.Errorf("GNU long value is not exactly NUL terminated"))
		}
		values.text = text
		if typeflag == tar.TypeGNULongName {
			presence.gnuLongName = true
			if _, err := validateVirtualPath(text, limits); err != nil {
				var pathErr *pathFailure
				_ = errorAs(err, &pathErr)
				return "", presence, values, cursor.offset, &tarMetadataError{
					code: CodeArchiveUnsafePath, reason: "unsafe_tar_gnu_long_name", pathFailure: pathErr,
				}
			}
			return "gnu-long-name", presence, values, cursor.offset, nil
		}
		presence.gnuLongLink = true
		return "gnu-long-link", presence, values, cursor.offset, nil
	case tar.TypeXHeader, tar.TypeXGlobalHeader:
		records := make(map[string]string)
		seen := make(map[string]struct{})
		for cursor.offset < cursor.size {
			recordStart := cursor.offset
			lengthDigits := make([]byte, 0, 20)
			for {
				value, err := cursor.readByte()
				if err != nil {
					return readFailure(err)
				}
				if value == ' ' {
					break
				}
				if value < '0' || value > '9' || len(lengthDigits) == cap(lengthDigits) {
					return readFailure(fmt.Errorf("PAX record length is invalid"))
				}
				lengthDigits = append(lengthDigits, value)
			}
			if len(lengthDigits) == 0 {
				return readFailure(fmt.Errorf("PAX record length is absent"))
			}
			recordLength, err := strconv.ParseInt(string(lengthDigits), 10, 64)
			if err != nil || recordLength < 5 || recordLength > cursor.size-recordStart {
				return readFailure(fmt.Errorf("PAX record length is invalid"))
			}
			recordEnd, ok := checkedAdd(recordStart, recordLength)
			if !ok || recordEnd > cursor.size || recordEnd <= cursor.offset {
				return readFailure(fmt.Errorf("PAX record length is invalid"))
			}
			keyBytes := make([]byte, 0, 64)
			for {
				if cursor.offset >= recordEnd-1 {
					return readFailure(fmt.Errorf("PAX record key/value is invalid"))
				}
				value, err := cursor.readByte()
				if err != nil {
					return readFailure(err)
				}
				if value == '=' {
					break
				}
				if value == 0 || len(keyBytes) >= int(limits.MaxPathBytes) {
					return readFailure(fmt.Errorf("PAX record key/value is invalid"))
				}
				keyBytes = append(keyBytes, value)
			}
			if len(keyBytes) == 0 || !utf8.Valid(keyBytes) {
				return readFailure(fmt.Errorf("PAX record key/value is invalid"))
			}
			key := string(keyBytes)
			if _, duplicate := seen[key]; duplicate {
				return readFailure(fmt.Errorf("duplicate PAX record key"))
			}
			seen[key] = struct{}{}
			valueSize := recordEnd - cursor.offset - 1
			if valueSize < 0 {
				return readFailure(fmt.Errorf("PAX record key/value is invalid"))
			}
			if key == "path" || key == "linkpath" {
				if pathErr := preflightContainerMemberPathLength(containerPath, valueSize, limits); pathErr != nil {
					return "", presence, values, cursor.offset, &tarMetadataError{
						code: CodeArchiveUnsafePath, reason: pathErr.reason, pathFailure: pathErr,
					}
				}
			}
			capture := key == "path" || key == "linkpath" || key == "uid" || key == "gid" ||
				key == "mtime" || key == "atime" || key == "ctime"
			if capture && valueSize > limits.MaxPathBytes {
				return readFailure(fmt.Errorf("PAX scalar value is too long"))
			}
			value, err := cursor.readUTF8(valueSize, capture)
			if err != nil {
				return readFailure(err)
			}
			terminator, err := cursor.readByte()
			if err != nil {
				return readFailure(err)
			}
			if terminator != '\n' || cursor.offset != recordEnd {
				return readFailure(fmt.Errorf("PAX record terminator is invalid"))
			}
			switch key {
			case "path":
				presence.paxPath = true
				if _, err := validateVirtualPath(value, limits); err != nil {
					var pathErr *pathFailure
					_ = errorAs(err, &pathErr)
					return "", presence, values, cursor.offset, &tarMetadataError{
						code: CodeArchiveUnsafePath, reason: "unsafe_tar_pax_path", pathFailure: pathErr,
					}
				}
				records[key] = value
			case "linkpath":
				presence.paxLink = true
				records[key] = value
			case "uid":
				presence.uid = true
				if _, err := strconv.ParseInt(value, 10, 64); err != nil {
					return readFailure(fmt.Errorf("invalid PAX uid"))
				}
			case "gid":
				presence.gid = true
				if _, err := strconv.ParseInt(value, 10, 64); err != nil {
					return readFailure(fmt.Errorf("invalid PAX gid"))
				}
			case "uname":
				presence.uname = true
			case "gname":
				presence.gname = true
			case "mtime", "atime", "ctime":
				if !validPAXTime(value) {
					return readFailure(fmt.Errorf("invalid PAX %s", key))
				}
				switch key {
				case "mtime":
					presence.mtime = true
				case "atime":
					presence.atime = true
				case "ctime":
					presence.ctime = true
				}
			case "comment":
				presence.paxComment = true
			case "charset":
				presence.paxCharset = true
			case "size":
				return "", presence, values, cursor.offset, &tarMetadataError{
					code: CodeArchiveUnsupported, reason: "tar_pax_size_override_unsupported",
				}
			default:
				lowerKey := strings.ToLower(key)
				if strings.HasPrefix(key, "SCHILY.xattr.") {
					presence.xattr = true
					values.failureCode = CodeArchiveUnsafeEntry
					values.failureReason = "tar_xattr_unsupported"
				} else if strings.HasPrefix(lowerKey, "gnu.sparse") || strings.Contains(lowerKey, "sparse") {
					values.failureCode = CodeArchiveUnsafeEntry
					values.failureReason = "tar_sparse_or_external_extent"
				} else {
					values.failureCode = CodeArchiveUnsupported
					values.failureReason = "tar_pax_key_unsupported"
				}
			}
		}
		values.records = records
		kind := "pax-local"
		if typeflag == tar.TypeXGlobalHeader {
			kind = "pax-global"
			if presence.paxPath || presence.paxLink {
				return "", presence, values, cursor.offset, &tarMetadataError{
					code: CodeArchiveUnsupported, reason: "tar_global_path_resolution_unsupported",
				}
			}
		}
		return kind, presence, values, cursor.offset, nil
	default:
		panic("parseTarMetadataBlob called for non-metadata type")
	}
}

func applyTarMetadata(pending *pendingTarMetadata, typeflag byte, values tarMetadataValues, presence tarMetadataPresence, member string) *tarMetadataError {
	switch typeflag {
	case tar.TypeXGlobalHeader:
		if pending.localPAXSeen || pending.gnuNameSeen || pending.gnuLinkSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "tar_global_metadata_interrupts_local_metadata"}
		}
		return nil
	case tar.TypeXHeader:
		if pending.localPAXSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "repeated_tar_pax_local_header"}
		}
		pending.localPAXSeen = true
		pending.pax = values.records
		if values.records["path"] != "" && pending.gnuNameSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "conflicting_tar_path_resolution_metadata"}
		}
		if values.records["linkpath"] != "" && pending.gnuLinkSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "conflicting_tar_link_resolution_metadata"}
		}
	case tar.TypeGNULongName:
		if pending.gnuNameSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "repeated_tar_gnu_long_name_header"}
		}
		if pending.pax["path"] != "" {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "conflicting_tar_path_resolution_metadata"}
		}
		pending.gnuNameSeen = true
		pending.gnuLongName = values.text
	case tar.TypeGNULongLink:
		if pending.gnuLinkSeen {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "repeated_tar_gnu_long_link_header"}
		}
		if pending.pax["linkpath"] != "" {
			return &tarMetadataError{code: CodeArchiveInvalid, reason: "conflicting_tar_link_resolution_metadata"}
		}
		pending.gnuLinkSeen = true
		pending.gnuLongLink = values.text
	}
	mergeTarPresence(&pending.presence, presence)
	pending.members = append(pending.members, member)
	return nil
}

func validPAXTime(value string) bool {
	if value == "" {
		return false
	}
	index := 0
	if value[0] == '-' {
		index++
		if index == len(value) {
			return false
		}
	}
	digits := 0
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		index++
		digits++
	}
	if digits == 0 {
		return false
	}
	if index == len(value) {
		return true
	}
	if value[index] != '.' || index+1 == len(value) {
		return false
	}
	for index++; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func tarHeaderPresence(header []byte) tarMetadataPresence {
	return tarMetadataPresence{
		mtime: tarFieldPresent(header[136:148]), uid: tarFieldPresent(header[108:116]),
		gid: tarFieldPresent(header[116:124]), uname: tarFieldPresent(header[265:297]),
		gname: tarFieldPresent(header[297:329]),
	}
}

func tarFieldPresent(field []byte) bool { return len(bytes.Trim(field, " \x00")) > 0 }

func mergeTarPresence(target *tarMetadataPresence, source tarMetadataPresence) {
	target.atime = target.atime || source.atime
	target.ctime = target.ctime || source.ctime
	target.gid = target.gid || source.gid
	target.gname = target.gname || source.gname
	target.mtime = target.mtime || source.mtime
	target.uid = target.uid || source.uid
	target.uname = target.uname || source.uname
	target.xattr = target.xattr || source.xattr
	target.paxPath = target.paxPath || source.paxPath
	target.paxLink = target.paxLink || source.paxLink
	target.paxComment = target.paxComment || source.paxComment
	target.paxCharset = target.paxCharset || source.paxCharset
	target.gnuLongName = target.gnuLongName || source.gnuLongName
	target.gnuLongLink = target.gnuLongLink || source.gnuLongLink
}

func appendTarPresenceFacts(target map[string]any, presence tarMetadataPresence, members []string, physicalIndex int64, rawName string) {
	target["atime_present"] = presence.atime
	target["ctime_present"] = presence.ctime
	target["gid_present"] = presence.gid
	target["gname_present"] = presence.gname
	target["gnu_long_link_present"] = presence.gnuLongLink
	target["gnu_long_name_present"] = presence.gnuLongName
	target["metadata_members"] = strings.Join(members, ",")
	target["mtime_present"] = presence.mtime
	target["pax_charset_present"] = presence.paxCharset
	target["pax_comment_present"] = presence.paxComment
	target["pax_linkpath_present"] = presence.paxLink
	target["pax_path_present"] = presence.paxPath
	if physicalIndex > 0 {
		target["physical_header_index"] = physicalIndex
	}
	target["raw_name_base64"] = base64.StdEncoding.EncodeToString([]byte(rawName))
	target["uid_present"] = presence.uid
	target["uname_present"] = presence.uname
	target["xattr_present"] = presence.xattr
}

func rawTarHeaderName(header []byte) string {
	name := tarString(header[:100])
	if string(header[257:263]) == "ustar\x00" {
		if prefix := tarString(header[345:500]); prefix != "" {
			name = prefix + "/" + name
		}
	}
	return name
}

func tarString(field []byte) string {
	if index := bytes.IndexByte(field, 0); index >= 0 {
		field = field[:index]
	}
	return string(field)
}

func tarHeaderMode(header []byte) int64 {
	mode, ok := parseTarNumber(header[100:108])
	if !ok {
		return 0
	}
	return mode
}

func normalizedTarType(typeflag byte, name string) byte {
	if typeflag == 0 {
		if strings.HasSuffix(name, "/") {
			return tar.TypeDir
		}
		return tar.TypeReg
	}
	return typeflag
}

func tarMetadataFormat(typeflag byte) string {
	if typeflag == tar.TypeXHeader || typeflag == tar.TypeXGlobalHeader {
		return "PAX"
	}
	return "GNU"
}

func tarMetadataKind(typeflag byte) string {
	switch typeflag {
	case tar.TypeXHeader:
		return "pax-local"
	case tar.TypeXGlobalHeader:
		return "pax-global"
	case tar.TypeGNULongName:
		return "gnu-long-name"
	case tar.TypeGNULongLink:
		return "gnu-long-link"
	default:
		panic("tarMetadataKind called for non-metadata type")
	}
}

func roundTarBlock(size int64) (int64, bool) {
	if size < 0 || size > math.MaxInt64-511 {
		return 0, false
	}
	return (size + 511) &^ 511, true
}

func nonzeroBlobRemainder(item blob, offset int64) (bool, error) {
	buffer := make([]byte, 32<<10)
	for offset < item.size {
		amount := int64(len(buffer))
		if amount > item.size-offset {
			amount = item.size - offset
		}
		payload, err := readExactAt(item, offset, amount)
		if err != nil {
			return false, err
		}
		if !allZero(payload) {
			return true, nil
		}
		offset += amount
	}
	return false, nil
}

func tarDiagnostic(containerPath string, chain []string, code DiagnosticCode, reason string) *Diagnostic {
	return &Diagnostic{
		Code: code, Path: containerPath, DetectorID: "archive-tar-v1",
		ContainerChain: append([]string(nil), chain...), Reason: reason,
	}
}

func tarReadDiagnostic(containerPath string, chain []string, err error) *Diagnostic {
	diagnostic := containerReadDiagnostic(containerPath, chain, "archive-tar-v1", err)
	return &diagnostic
}

func sparseTarHeader(header *tar.Header) bool {
	for key := range header.PAXRecords {
		if strings.HasPrefix(strings.ToLower(key), "gnu.sparse") || strings.Contains(strings.ToLower(key), "sparse") {
			return true
		}
	}
	return false
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (reader *countingReader) Read(payload []byte) (int, error) {
	read, err := reader.reader.Read(payload)
	reader.count += int64(read)
	return read, err
}

func nonzeroRemainder(reader io.Reader) (bool, error) {
	buffer := make([]byte, 32<<10)
	for {
		read, err := reader.Read(buffer)
		if read > 0 && !allZero(buffer[:read]) {
			return true, nil
		}
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
}

func (inspector *inspector) walkGZIP(nodeIndex int, item blob, depth int64) bool {
	containerPath := inspector.nodes[nodeIndex].Path
	chain := inspector.nodes[nodeIndex].ContainerChain
	if err := inspector.account.addEntry(1); err != nil {
		inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, err))
		return true
	}
	compressed := newCountingByteReader(item.reader())
	reader, err := gzip.NewReader(compressed)
	if err != nil {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "compression-gzip-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "gzip_header:" + err.Error(),
		})
		return true
	}
	reader.Multistream(false)
	// A root payload may use a portable virtual directory (for example
	// npm/pkg.tgz). The emitted gzip member is still a direct child of the
	// compressed stream; retaining the root directory in childName would create
	// a synthetic directory between them and lose compressed-size accounting.
	childName := gzipChildName(path.Base(leafPath(containerPath)))
	childPath := joinContainerPath(containerPath, childName)
	memberChain := append(append([]string(nil), chain...), containerPath)
	spool, err := inspector.store.newFile()
	if err != nil {
		_ = reader.Close()
		inspector.rejectNode(nodeIndex, unavailableDiagnostic(containerPath, chain, "create_gzip_spool", err))
		return true
	}
	decoded, observed, compressedInput, limitErr, readErr := inspector.captureGZIPStream(spool, reader, compressed)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil {
		if readErr == nil {
			readErr = closeErr
		}
		if observed > 0 {
			_ = inspector.account.checkLeaf(observed)
			_ = inspector.account.addEmitted(observed, compressedInput)
		}
		inspector.rejectNode(nodeIndex, containerReadDiagnostic(childPath, memberChain, "compression-gzip-v1", readErr))
		return true
	}
	if limitErr != nil {
		_ = inspector.account.checkLeaf(observed)
		_ = inspector.account.addEmitted(observed, compressedInput)
		inspector.rejectNode(nodeIndex, diagnosticFromError(childPath, memberChain, limitErr))
		return true
	}
	if err := inspector.account.checkLeaf(decoded.size); err != nil {
		inspector.rejectNode(nodeIndex, diagnosticFromError(childPath, memberChain, err))
		return true
	}
	if err := inspector.account.addEmitted(decoded.size, compressedInput); err != nil {
		inspector.rejectNode(nodeIndex, diagnosticFromError(childPath, memberChain, err))
		return true
	}
	_, trailingErr := compressed.ReadByte()
	if trailingErr != io.EOF {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "compression-gzip-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "gzip_trailing_or_second_stream",
		})
		return true
	}
	raw := []rawMember{{
		name: childName, kind: NodeRegularFile, data: &decoded,
		declaredSize: decoded.size, compressed: compressedInput, charged: true, explicit: true,
		observations: []Observation{{
			DetectorID: "compression-gzip-v1", Result: "ENTRY",
			Facts: facts(map[string]any{
				"comment": reader.Comment, "extra_size": len(reader.Extra),
				"header_name": reader.Name, "modtime_recorded": !reader.ModTime.IsZero(),
				"os": reader.OS,
			}),
		}},
	}}
	prepared, rejected := inspector.prepareMembers(containerPath, chain, raw)
	return inspector.inspectMembers(nodeIndex, prepared, spool, depth, rejected)
}

// countingByteReader implements both io.Reader and io.ByteReader so gzip and
// flate do not read past the first stream. Bytes buffered from the underlying
// blob are charged only when delivered to the decoder; padding and subsequent
// streams therefore cannot inflate the first stream's expansion allowance.
type countingByteReader struct {
	reader io.Reader
	buffer [32 << 10]byte
	start  int
	end    int
	count  int64
}

func newCountingByteReader(reader io.Reader) *countingByteReader {
	return &countingByteReader{reader: reader}
}

func (reader *countingByteReader) Read(payload []byte) (int, error) {
	if len(payload) == 0 {
		return 0, nil
	}
	if reader.start < reader.end {
		read := copy(payload, reader.buffer[reader.start:reader.end])
		reader.start += read
		reader.count += int64(read)
		return read, nil
	}
	read, err := reader.reader.Read(payload)
	reader.count += int64(read)
	return read, err
}

func (reader *countingByteReader) ReadByte() (byte, error) {
	if reader.start == reader.end {
		read, err := reader.reader.Read(reader.buffer[:])
		if read == 0 {
			return 0, err
		}
		reader.start = 0
		reader.end = read
	}
	value := reader.buffer[reader.start]
	reader.start++
	reader.count++
	return value, nil
}

func (inspector *inspector) captureGZIPStream(
	destination *os.File,
	reader *gzip.Reader,
	compressed *countingByteReader,
) (blob, int64, int64, *limitFailure, error) {
	offset, err := destination.Seek(0, io.SeekEnd)
	if err != nil {
		return blob{}, 0, compressed.count, nil, err
	}
	digest := sha256.New()
	written := int64(0)
	buffer := make([]byte, 32<<10)
	for {
		if err := contextError(inspector.ctx); err != nil {
			return blob{}, written, compressed.count, nil, err
		}
		budget, budgetErr := inspector.account.streamBudget(compressed.count)
		if budgetErr != nil {
			return blob{}, written, compressed.count, nil, budgetErr
		}
		allowance := budget.maximum - written
		if allowance < 0 {
			allowance = 0
		}
		request := allowance + 1
		if request > int64(len(buffer)) {
			request = int64(len(buffer))
		}
		read, readErr := reader.Read(buffer[:request])
		if read > 0 {
			after, afterErr := inspector.account.streamBudget(compressed.count)
			if afterErr != nil {
				return blob{}, written, compressed.count, nil, afterErr
			}
			permitted := after.maximum - written
			if permitted < 0 {
				permitted = 0
			}
			retained := int64(read)
			if retained > permitted {
				retained = permitted
			}
			if retained > 0 {
				if _, err := destination.Write(buffer[:retained]); err != nil {
					return blob{}, written, compressed.count, nil, err
				}
				_, _ = digest.Write(buffer[:retained])
				written += retained
			}
			if retained < int64(read) {
				observed := written + 1
				return blob{
					file: destination, offset: offset, size: written,
					sha256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
				}, observed, compressed.count, after.failure(observed), nil
			}
		}
		if readErr == io.EOF {
			return blob{
				file: destination, offset: offset, size: written,
				sha256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
			}, written, compressed.count, nil, nil
		}
		if readErr != nil {
			return blob{}, written, compressed.count, nil, readErr
		}
		if read == 0 {
			return blob{}, written, compressed.count, nil, io.ErrNoProgress
		}
	}
}

func gzipChildName(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return name[:len(name)-3]
	case strings.HasSuffix(lower, ".tgz"):
		return name[:len(name)-4] + ".tar"
	case strings.HasSuffix(lower, ".gz"):
		return name[:len(name)-3]
	default:
		return "$stream"
	}
}

type arStringTable struct {
	names map[int64]string
}

type arStringTableError struct {
	code        DiagnosticCode
	reason      string
	observed    int64
	pathFailure *pathFailure
}

func parseARStringTable(
	ctx context.Context,
	item blob,
	dataOffset int64,
	dataSize int64,
	containerPath string,
	limits LimitVector,
) (*arStringTable, *arStringTableError) {
	table := &arStringTable{names: make(map[int64]string)}
	offset := int64(0)
	remainingPathBytes := limits.MaxPathBytes - int64(len(containerPath)) - 2
	if remainingPathBytes < 0 {
		remainingPathBytes = 0
	}
	for offset < dataSize {
		start := offset
		nameBytes := make([]byte, 0, minInt64(remainingPathBytes+1, 256))
		for {
			if offset >= dataSize {
				return nil, &arStringTableError{
					code: CodeArchiveInvalid, reason: "unterminated_GNU_ar_member_name", observed: offset,
				}
			}
			if err := contextError(ctx); err != nil {
				return nil, &arStringTableError{
					code: CodeInspectionUnavailable, reason: "ar_string_table_read:" + err.Error(), observed: offset,
				}
			}
			var current [1]byte
			if _, err := item.readAt(current[:], dataOffset+offset); err != nil {
				return nil, &arStringTableError{
					code: CodeInspectionUnavailable, reason: "ar_string_table_read:" + err.Error(), observed: offset,
				}
			}
			offset++
			if current[0] == '\n' || current[0] == 0 {
				break
			}
			nameBytes = append(nameBytes, current[0])
			// One extra byte is sufficient to distinguish a valid trailing GNU
			// slash from a name that cannot fit in the remaining full path.
			if int64(len(nameBytes)) > remainingPathBytes+1 {
				pathErr := preflightContainerMemberPathLength(containerPath, int64(len(nameBytes)), limits)
				return nil, &arStringTableError{
					code: CodeArchiveUnsafePath, reason: pathErr.reason, observed: offset, pathFailure: pathErr,
				}
			}
		}
		name := strings.TrimSuffix(string(nameBytes), "/")
		if name == "" || !utf8.ValidString(name) {
			return nil, &arStringTableError{
				code: CodeArchiveInvalid, reason: "invalid_GNU_ar_member_name", observed: offset,
			}
		}
		if pathErr := preflightContainerMemberPathLength(containerPath, int64(len(name)), limits); pathErr != nil {
			return nil, &arStringTableError{
				code: CodeArchiveUnsafePath, reason: pathErr.reason, observed: offset, pathFailure: pathErr,
			}
		}
		if _, err := validateVirtualPath(name, limits); err != nil {
			var pathErr *pathFailure
			_ = errorAs(err, &pathErr)
			return nil, &arStringTableError{
				code: CodeArchiveUnsafePath, reason: "unsafe_GNU_ar_string_table_name", observed: offset,
				pathFailure: pathErr,
			}
		}
		table.names[start] = name
	}
	return table, nil
}

func minInt64(left, right int64) int {
	if left < right {
		return int(left)
	}
	return int(right)
}

func (inspector *inspector) walkAR(nodeIndex int, item blob, depth int64) bool {
	containerPath := inspector.nodes[nodeIndex].Path
	chain := inspector.nodes[nodeIndex].ContainerChain
	magic, err := readExactAt(item, 0, 8)
	if err != nil {
		inspector.rejectNode(nodeIndex, containerReadDiagnostic(containerPath, chain, "archive-ar-v1", err))
		return true
	}
	if string(magic) == "!<thin>\n" {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveUnsafeEntry, Path: containerPath, DetectorID: "archive-ar-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "thin_archive_external_members",
		})
		return true
	}
	if string(magic) != "!<arch>\n" {
		inspector.rejectNode(nodeIndex, Diagnostic{
			Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
			ContainerChain: append([]string(nil), chain...), Reason: "invalid_ar_magic",
		})
		return true
	}
	offset := int64(8)
	physicalIndex := int64(0)
	var stringTable *arStringTable
	metadataCounts := make(map[string]int)
	raw := make([]rawMember, 0)
	for offset < item.size {
		if item.size-offset < 60 {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "truncated_ar_header",
			})
			return true
		}
		if accountErr := inspector.account.addEntry(1); accountErr != nil {
			inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, accountErr))
			return true
		}
		physicalIndex++
		header, readErr := readExactAt(item, offset, 60)
		if readErr != nil || string(header[58:60]) != "`\n" {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "invalid_ar_header",
			})
			return true
		}
		mode, metadataOK := validateARMetadata(header)
		if !metadataOK {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "invalid_ar_metadata",
			})
			return true
		}
		size, ok := parseDecimal(header[48:58])
		if !ok || size > item.size-(offset+60) {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "invalid_ar_member_size",
			})
			return true
		}
		nameField := strings.TrimSpace(string(header[:16]))
		dataOffset := offset + 60
		dataSize := size
		bsdExtendedName := strings.HasPrefix(nameField, "#1/")
		bsdNameLength := int64(0)
		if bsdExtendedName {
			nameLength, nameLengthErr := strconv.ParseInt(strings.TrimPrefix(nameField, "#1/"), 10, 64)
			if nameLengthErr != nil || nameLength <= 0 || nameLength > dataSize {
				inspector.rejectNode(nodeIndex, Diagnostic{
					Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
					ContainerChain: append([]string(nil), chain...), Reason: "invalid BSD ar extended name",
				})
				return true
			}
			bsdNameLength = nameLength
			metadataPath := joinContainerPath(
				containerPath, fmt.Sprintf("$ar-metadata/bsd-extended-name-%06d", physicalIndex),
			)
			metadataChain := append(append([]string(nil), chain...), containerPath)
			if limitErr := inspector.account.preflightLeaf(size); limitErr != nil {
				_ = inspector.account.checkLeaf(size)
				inspector.rejectNode(nodeIndex, diagnosticFromError(metadataPath, metadataChain, limitErr))
				return true
			}
			if limitErr := inspector.account.preflightEmitted(size, size); limitErr != nil {
				inspector.rejectNode(nodeIndex, diagnosticFromError(metadataPath, metadataChain, limitErr))
				return true
			}
			if pathErr := preflightContainerMemberPathLength(containerPath, nameLength, inspector.limits); pathErr != nil {
				diagnostic := containerPathDiagnostic(containerPath, "#1/"+strconv.FormatInt(nameLength, 10), chain, pathErr)
				diagnostic.Path = metadataPath
				inspector.rejectNode(nodeIndex, diagnostic)
				return true
			}
		}
		name, nameErr := resolveARName(nameField, stringTable, item, dataOffset, &dataSize)
		if nameErr != nil {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: nameErr.Error(),
			})
			return true
		}
		if bsdExtendedName {
			dataOffset += bsdNameLength
		}
		memberName, metadataKind, metadata := canonicalARMemberName(nameField, name, metadataCounts)
		memberBlob := blob{file: item.file, offset: item.offset + dataOffset, size: dataSize}
		if _, pathErr := validateVirtualPath(memberName, inspector.limits); pathErr != nil {
			raw = append(raw, rawMember{
				name: memberName, originalName: name, kind: NodeRegularFile, mode: mode, data: &memberBlob,
				declaredSize: dataSize, compressed: dataSize, explicit: true,
			})
		} else {
			accountedSize := dataSize
			if bsdExtendedName {
				accountedSize = size
			}
			if !bsdExtendedName {
				if limitErr := inspector.account.preflightLeaf(accountedSize); limitErr != nil {
					_ = inspector.account.checkLeaf(accountedSize)
					inspector.rejectNode(nodeIndex, diagnosticFromError(
						joinContainerPath(containerPath, memberName),
						append(append([]string(nil), chain...), containerPath), limitErr,
					))
					return true
				}
			}
			if !bsdExtendedName {
				if limitErr := inspector.account.preflightEmitted(accountedSize, accountedSize); limitErr != nil {
					inspector.rejectNode(nodeIndex, diagnosticFromError(
						joinContainerPath(containerPath, memberName),
						append(append([]string(nil), chain...), containerPath), limitErr,
					))
					return true
				}
			}
			var metadataPayload []byte
			if nameField == "//" {
				var tableErr *arStringTableError
				stringTable, tableErr = parseARStringTable(
					inspector.ctx, item, dataOffset, dataSize, containerPath, inspector.limits,
				)
				if tableErr != nil {
					if tableErr.observed > 0 {
						_ = inspector.account.checkLeaf(tableErr.observed)
						_ = inspector.account.addEmitted(tableErr.observed, tableErr.observed)
					}
					diagnostic := Diagnostic{
						Code: tableErr.code, Path: joinContainerPath(containerPath, memberName),
						OriginalNameBase64: originalNameBase64(nameField),
						DetectorID:         "archive-ar-v1",
						ContainerChain:     append(append([]string(nil), chain...), containerPath),
						Reason:             tableErr.reason,
					}
					if tableErr.pathFailure != nil {
						diagnostic.LimitName = tableErr.pathFailure.limitName
						diagnostic.Limit = tableErr.pathFailure.limit
						diagnostic.Observed = tableErr.pathFailure.observed
					}
					inspector.rejectNode(nodeIndex, diagnostic)
					return true
				}
			}
			if metadata && nameField != "//" {
				metadataPayload, readErr = readExactAt(item, dataOffset, dataSize)
				if readErr != nil {
					inspector.rejectNode(nodeIndex, containerReadDiagnostic(containerPath, chain, "archive-ar-v1", readErr))
					return true
				}
			}
			if metadata {
				if metadataErr := validateARStructuralMetadata(nameField, name, metadataPayload); metadataErr != nil {
					inspector.rejectNode(nodeIndex, Diagnostic{
						Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
						ContainerChain: append([]string(nil), chain...), Reason: metadataErr.Error(),
					})
					return true
				}
			}
			memberBlob.sha256, readErr = hashBlob(inspector.ctx, memberBlob)
			if readErr != nil {
				inspector.rejectNode(nodeIndex, containerReadDiagnostic(containerPath, chain, "archive-ar-v1", readErr))
				return true
			}
			if limitErr := inspector.account.checkLeaf(accountedSize); limitErr != nil {
				inspector.rejectNode(nodeIndex, Diagnostic{
					Code: CodePolicyInternalError, Path: joinContainerPath(containerPath, memberName),
					DetectorID: "archive-ar-v1", ContainerChain: append(append([]string(nil), chain...), containerPath),
					Reason: "ar_member_leaf_preflight_drift",
				})
				return true
			}
			if accountErr := inspector.account.addEmitted(accountedSize, accountedSize); accountErr != nil {
				inspector.rejectNode(nodeIndex, diagnosticFromError(containerPath, chain, accountErr))
				return true
			}
			entryFacts := map[string]any{
				"declared_size": size, "member_size": dataSize, "raw_name": nameField,
			}
			originalName := ""
			if bsdExtendedName {
				entryFacts["extended_name_size"] = bsdNameLength
				entryFacts["extended_name_sha256"] = digestBytes([]byte(name))
				originalName = name
			}
			if metadata {
				entryFacts["metadata_kind"] = metadataKind
			}
			raw = append(raw, rawMember{
				name: memberName, originalName: originalName, kind: NodeRegularFile, mode: mode, data: &memberBlob,
				declaredSize: dataSize, compressed: dataSize, charged: true, explicit: true,
				preclassified: metadata, class: ClassNativeLibraryStatic,
				variant: "ar.metadata." + metadataKind,
				observations: []Observation{{
					DetectorID: "archive-ar-v1", Result: "ENTRY", Facts: facts(entryFacts),
				}},
			})
		}
		next, ok := checkedAdd(offset+60, size)
		if !ok {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "ar_offset_overflow",
			})
			return true
		}
		if next%2 != 0 {
			padding, paddingErr := readExactAt(item, next, 1)
			if paddingErr != nil || padding[0] != '\n' {
				inspector.rejectNode(nodeIndex, Diagnostic{
					Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
					ContainerChain: append([]string(nil), chain...), Reason: "invalid_ar_padding",
				})
				return true
			}
			next++
		}
		if next > item.size {
			inspector.rejectNode(nodeIndex, Diagnostic{
				Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-ar-v1",
				ContainerChain: append([]string(nil), chain...), Reason: "ar_padding_out_of_bounds",
			})
			return true
		}
		offset = next
	}
	prepared, rejected := inspector.prepareMembers(containerPath, chain, raw)
	return inspector.inspectMembers(nodeIndex, prepared, nil, depth, rejected)
}

func (inspector *inspector) preflightRawMembers(
	containerPath string,
	chain []string,
	members []rawMember,
) *Diagnostic {
	var total, compressed, maxLeaf int64
	for _, member := range members {
		if member.kind != NodeRegularFile {
			continue
		}
		fullPath := containerPath
		if validated, err := validateVirtualPath(member.name, inspector.limits); err == nil {
			fullPath = joinContainerPath(containerPath, validated.Canonical)
		}
		memberChain := append(append([]string(nil), chain...), containerPath)
		if member.declaredSize > maxLeaf {
			maxLeaf = member.declaredSize
		}
		if err := inspector.account.checkLeaf(member.declaredSize); err != nil {
			diagnostic := diagnosticFromError(fullPath, memberChain, err)
			diagnostic.Size = member.declaredSize
			return &diagnostic
		}
		if member.failureCode != "" {
			continue
		}
		if member.declaredSize > 0 && member.compressed == 0 {
			diagnostic := diagnosticFromError(fullPath, memberChain, &limitFailure{
				name: "max_expansion_ratio", limit: inspector.limits.MaxExpansionRatio,
				observed: int64(^uint64(0) >> 1),
			})
			diagnostic.Size = member.declaredSize
			return &diagnostic
		}
		if member.compressed > 0 && ratioExceeded(
			member.declaredSize, member.compressed, inspector.limits.MaxExpansionRatio,
		) {
			diagnostic := diagnosticFromError(fullPath, memberChain, &limitFailure{
				name: "max_expansion_ratio", limit: inspector.limits.MaxExpansionRatio,
				observed: ceilingRatio(member.declaredSize, member.compressed),
			})
			diagnostic.Size = member.declaredSize
			return &diagnostic
		}
		var ok bool
		total, ok = checkedAdd(total, member.declaredSize)
		if !ok {
			diagnostic := diagnosticFromError(fullPath, memberChain, &limitFailure{
				name: "max_total_emitted_bytes", limit: inspector.limits.MaxTotalEmittedBytes,
				observed: int64(^uint64(0) >> 1),
			})
			diagnostic.Size = member.declaredSize
			return &diagnostic
		}
		compressed, ok = checkedAdd(compressed, member.compressed)
		if !ok {
			diagnostic := diagnosticFromError(fullPath, memberChain, &limitFailure{
				name: "max_expansion_ratio", limit: inspector.limits.MaxExpansionRatio,
				observed: int64(^uint64(0) >> 1),
			})
			diagnostic.Size = member.declaredSize
			return &diagnostic
		}
	}
	if err := inspector.account.preflightEmitted(total, compressed); err != nil {
		diagnostic := diagnosticFromError(containerPath, chain, err)
		diagnostic.Size = maxLeaf
		return &diagnostic
	}
	return nil
}

func (inspector *inspector) materializeZIPMembers(
	members []rawMember,
	spool *os.File,
) bool {
	rejected := false
	for index := range members {
		member := &members[index]
		if member.kind != NodeRegularFile || member.failureCode != "" || member.data != nil {
			continue
		}
		if member.open == nil {
			member.failureCode = CodeInspectionUnavailable
			member.failureReason = "member_reader_missing"
			rejected = true
			continue
		}
		reader, err := member.open()
		if err != nil {
			member.failureCode = CodeArchiveInvalid
			member.failureReason = "zip_member_open:" + err.Error()
			rejected = true
			continue
		}
		captured, captureErr := inspector.store.appendBlob(inspector.ctx, spool, reader, member.declaredSize)
		closeErr := reader.Close()
		if captureErr == nil {
			captureErr = closeErr
		}
		if captureErr != nil {
			member.failureCode = CodeArchiveInvalid
			if contextErrorCode(captureErr) {
				member.failureCode = CodeInspectionUnavailable
			}
			member.failureReason = "zip_member_read:" + captureErr.Error()
			rejected = true
			continue
		}
		if err := inspector.account.addEmitted(captured.size, member.compressed); err != nil {
			member.failureCode = CodeInspectionLimitExceeded
			member.failureReason = err.Error()
			rejected = true
			continue
		}
		member.data = &captured
		member.charged = true
	}
	return rejected
}

func validateARMetadata(header []byte) (int64, bool) {
	if len(header) != 60 || !validDecimalField(header[16:28]) ||
		!validDecimalField(header[28:34]) || !validDecimalField(header[34:40]) {
		return 0, false
	}
	modeText := strings.TrimSpace(string(header[40:48]))
	if modeText == "" {
		return 0, true
	}
	mode, err := strconv.ParseInt(modeText, 8, 64)
	if err != nil || mode < 0 {
		return 0, false
	}
	return mode, true
}

func validDecimalField(field []byte) bool {
	value := strings.TrimSpace(string(field))
	if value == "" {
		return true
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func resolveARName(field string, stringTable *arStringTable, item blob, dataOffset int64, dataSize *int64) (string, error) {
	switch {
	case field == "/" || field == "/SYM64/" || field == "//":
		return field, nil
	case strings.HasPrefix(field, "#1/"):
		length, err := strconv.ParseInt(strings.TrimPrefix(field, "#1/"), 10, 64)
		if err != nil || length <= 0 || length > *dataSize {
			return "", fmt.Errorf("invalid BSD ar extended name")
		}
		name, err := readExactAt(item, dataOffset, length)
		if err != nil || !utf8.Valid(name) {
			return "", fmt.Errorf("invalid BSD ar member name")
		}
		*dataSize -= length
		return string(name), nil
	case strings.HasPrefix(field, "/"):
		offset, err := strconv.ParseInt(strings.TrimPrefix(field, "/"), 10, 64)
		if err != nil || offset < 0 || stringTable == nil {
			return "", fmt.Errorf("invalid GNU ar string-table reference")
		}
		name, ok := stringTable.names[offset]
		if !ok {
			return "", fmt.Errorf("invalid GNU ar string-table reference")
		}
		return name, nil
	default:
		return strings.TrimSuffix(field, "/"), nil
	}
}

func canonicalARMemberName(field, resolved string, counts map[string]int) (string, string, bool) {
	kind := ""
	switch field {
	case "/":
		kind = "symbol_table"
	case "/SYM64/":
		kind = "symbol_table_64"
	case "//":
		kind = "string_table"
	default:
		upper := strings.ToUpper(resolved)
		if strings.HasPrefix(upper, "__.SYMDEF") {
			kind = "bsd_symbol_table"
		}
	}
	if kind == "" {
		return resolved, "", false
	}
	counts[kind]++
	indexedKind := fmt.Sprintf("%s_%03d", kind, counts[kind])
	return "$ar-metadata/" + strings.ReplaceAll(indexedKind, "_", "-"), indexedKind, true
}

func validateARStructuralMetadata(field, resolved string, payload []byte) error {
	switch {
	case field == "//":
		// The table was already streamed, fully validated, and indexed before
		// any member reference can resolve through it.
		return nil
	case field == "/SYM64/":
		if !validARBigEndianSymbolTable(payload, 8) {
			return fmt.Errorf("invalid_ar_64_symbol_table")
		}
		return nil
	case field == "/":
		if !validARBigEndianSymbolTable(payload, 4) && !validCOFFSecondLinkerMember(payload) {
			return fmt.Errorf("invalid_ar_symbol_table")
		}
		return nil
	case strings.HasPrefix(strings.ToUpper(resolved), "__.SYMDEF"):
		if len(payload) < 8 || (!validBSDSymbolTable(payload, binary.LittleEndian) &&
			!validBSDSymbolTable(payload, binary.BigEndian)) {
			return fmt.Errorf("invalid_bsd_ar_symbol_table")
		}
		return nil
	default:
		return nil
	}
}

func validARBigEndianSymbolTable(payload []byte, width int) bool {
	if (width != 4 && width != 8) || len(payload) < width {
		return false
	}
	var count uint64
	if width == 4 {
		count = uint64(binary.BigEndian.Uint32(payload[:4]))
	} else {
		count = binary.BigEndian.Uint64(payload[:8])
	}
	if count > math.MaxInt64 {
		return false
	}
	count64 := int64(count)
	width64 := int64(width)
	payloadLength := int64(len(payload))
	if count64 > payloadLength/width64 {
		return false
	}
	offsetBytes := width64 + count64*width64
	if offsetBytes > payloadLength {
		return false
	}
	names := payload[offsetBytes:]
	for index := int64(0); index < count64; index++ {
		terminator := bytes.IndexByte(names, 0)
		if terminator < 0 {
			return false
		}
		names = names[terminator+1:]
	}
	return true
}

func validCOFFSecondLinkerMember(payload []byte) bool {
	if len(payload) < 8 {
		return false
	}
	memberCount := uint64(binary.LittleEndian.Uint32(payload[:4]))
	offset := uint64(4) + memberCount*4
	if memberCount > uint64(len(payload)/4) || offset+4 > uint64(len(payload)) {
		return false
	}
	symbolCount := uint64(binary.LittleEndian.Uint32(payload[offset : offset+4]))
	offset += 4
	if symbolCount > uint64(len(payload)/2) || offset+symbolCount*2 > uint64(len(payload)) {
		return false
	}
	names := payload[offset+symbolCount*2:]
	for index := uint64(0); index < symbolCount; index++ {
		terminator := bytes.IndexByte(names, 0)
		if terminator < 0 {
			return false
		}
		names = names[terminator+1:]
	}
	return true
}

func validBSDSymbolTable(payload []byte, order binary.ByteOrder) bool {
	if len(payload) < 8 {
		return false
	}
	tableBytes := uint64(order.Uint32(payload[:4]))
	if tableBytes%8 != 0 || tableBytes+8 > uint64(len(payload)) {
		return false
	}
	stringSizeOffset := 4 + tableBytes
	stringBytes := uint64(order.Uint32(payload[stringSizeOffset : stringSizeOffset+4]))
	return stringSizeOffset+4+stringBytes == uint64(len(payload))
}

func (inspector *inspector) prepareMembers(containerPath string, chain []string, raw []rawMember) ([]preparedMember, bool) {
	groups := make(map[string][]preparedMember)
	originalPaths := make(map[string]struct{})
	rejected := false
	for index := range raw {
		member := raw[index]
		validated, err := validateVirtualPath(member.name, inspector.limits)
		if err != nil {
			inspector.addDiagnostic(containerPathDiagnostic(containerPath, member.name, chain, err))
			rejected = true
			continue
		}
		if err := validateManifestVirtualPathWithLimits(joinContainerPath(containerPath, validated.Canonical), inspector.limits); err != nil {
			inspector.addDiagnostic(containerPathDiagnostic(containerPath, member.name, chain, err))
			rejected = true
			continue
		}
		member.explicit = true
		member.identity = canonicalRawMemberIdentity(member)
		groups[validated.Canonical] = append(groups[validated.Canonical], preparedMember{
			raw: member, path: validated, logical: validated,
		})
		originalPaths[validated.Canonical] = struct{}{}
	}

	logicalPaths := make([]string, 0, len(groups))
	for logical := range groups {
		logicalPaths = append(logicalPaths, logical)
	}
	sort.Strings(logicalPaths)
	collisions := make(map[string]string, len(logicalPaths))
	for _, logical := range logicalPaths {
		group := groups[logical]
		validated := group[0].logical
		if existing, collision := collisions[validated.CollisionKey]; collision && existing != logical {
			inspector.addDiagnostic(Diagnostic{
				Code: CodeArchiveUnsafePath, Path: joinContainerPath(containerPath, logical),
				OriginalNameBase64: originalNameBase64(rawMemberOriginalName(group[0].raw)), CollisionKey: validated.CollisionKey,
				ContainerChain: append(append([]string(nil), chain...), containerPath), Reason: "portable_path_collision",
				Details: []Fact{{Key: "collides_with", Value: existing}},
			})
			rejected = true
		}
		collisions[validated.CollisionKey] = logical
	}

	usedPaths := make(map[string]struct{}, len(originalPaths)+len(raw))
	usedCollisionKeys := make(map[string]struct{}, len(originalPaths)+len(raw))
	for original := range originalPaths {
		usedPaths[original] = struct{}{}
		validated, _ := validateVirtualPath(original, inspector.limits)
		usedCollisionKeys[validated.CollisionKey] = struct{}{}
	}
	prepared := make([]preparedMember, 0, len(raw))
	for _, logical := range logicalPaths {
		group := groups[logical]
		sort.SliceStable(group, func(left, right int) bool {
			return group[left].raw.identity < group[right].raw.identity
		})
		for occurrence := range group {
			member := group[occurrence]
			if len(group) > 1 {
				appendOccurrenceEvidence(
					&member.raw, joinContainerPath(containerPath, logical), member.raw.identity, occurrence+1, len(group),
				)
			}
			if occurrence > 0 {
				member.path = uniqueDuplicatePath(
					logical, member.raw.identity, occurrence+1, usedPaths, usedCollisionKeys, inspector.limits,
				)
				inspector.addDiagnostic(Diagnostic{
					Code: CodeArchiveUnsafePath, Path: joinContainerPath(containerPath, logical),
					OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)), CollisionKey: member.logical.CollisionKey,
					ContainerChain: append(append([]string(nil), chain...), containerPath), Reason: "duplicate_path",
					Details: facts(map[string]any{
						"occurrence_identity": member.raw.identity, "occurrence_number": occurrence + 1,
					}),
				})
				rejected = true
			}
			usedPaths[member.path.Canonical] = struct{}{}
			usedCollisionKeys[member.path.CollisionKey] = struct{}{}
			prepared = append(prepared, member)
		}
	}

	sort.Slice(prepared, func(left, right int) bool {
		return prepared[left].path.Canonical < prepared[right].path.Canonical
	})
	byPath := make(map[string]int, len(prepared))
	for index := range prepared {
		byPath[prepared[index].path.Canonical] = index
	}
	initialCount := len(prepared)
	for preparedIndex := 0; preparedIndex < initialCount; preparedIndex++ {
		member := prepared[preparedIndex]
		components := member.path.Components
		for index := 1; index < len(components); index++ {
			directory := strings.Join(components[:index], "/")
			if existingIndex, ok := byPath[directory]; ok {
				existing := prepared[existingIndex]
				if existing.raw.kind != NodeDirectory {
					inspector.addDiagnostic(Diagnostic{
						Code: CodeArchiveUnsafePath, Path: joinContainerPath(containerPath, member.path.Canonical),
						CollisionKey:   existing.path.CollisionKey,
						ContainerChain: append(append([]string(nil), chain...), containerPath),
						Reason:         "parent_is_not_directory",
					})
					rejected = true
				}
				continue
			}
			validated, _ := validateVirtualPath(directory, inspector.limits)
			if _, collision := usedCollisionKeys[validated.CollisionKey]; collision {
				inspector.addDiagnostic(Diagnostic{
					Code: CodeArchiveUnsafePath, Path: joinContainerPath(containerPath, directory),
					CollisionKey:   validated.CollisionKey,
					ContainerChain: append(append([]string(nil), chain...), containerPath),
					Reason:         "synthetic_directory_collision",
				})
				rejected = true
				continue
			}
			if err := inspector.account.addEntry(1); err != nil {
				inspector.addDiagnostic(diagnosticFromError(containerPath, chain, err))
				return nil, true
			}
			usedPaths[directory] = struct{}{}
			usedCollisionKeys[validated.CollisionKey] = struct{}{}
			byPath[directory] = len(prepared)
			prepared = append(prepared, preparedMember{
				raw:  rawMember{name: directory, kind: NodeDirectory},
				path: validated, logical: validated, synthetic: true,
			})
		}
	}
	sort.Slice(prepared, func(left, right int) bool {
		return prepared[left].path.Canonical < prepared[right].path.Canonical
	})
	return prepared, rejected
}

func canonicalRawMemberIdentity(member rawMember) string {
	for index := range member.observations {
		if member.observations[index].Facts == nil {
			member.observations[index].Facts = []Fact{}
		}
		sortFacts(member.observations[index].Facts)
	}
	payloadDigest := ""
	payloadSize := int64(0)
	if member.data != nil {
		payloadDigest = member.data.sha256
		payloadSize = member.data.size
	}
	identityObservations := rawMemberIdentityObservations(member)
	identity := struct {
		Name          string
		Kind          NodeKind
		Mode          int64
		DeclaredSize  int64
		Compressed    int64
		PayloadSize   int64
		PayloadSHA256 string
		FailureCode   DiagnosticCode
		FailureReason string
		Observations  []Observation
		Preclassified bool
		Class         ArtifactClass
		Variant       string
	}{
		Name: member.name, Kind: member.kind, Mode: member.mode,
		DeclaredSize: member.declaredSize, Compressed: member.compressed,
		PayloadSize: payloadSize, PayloadSHA256: payloadDigest,
		FailureCode: member.failureCode, FailureReason: member.failureReason,
		Observations: identityObservations, Preclassified: member.preclassified,
		Class: member.class, Variant: member.variant,
	}
	var payload []byte
	var err error
	if member.originalName == "" && member.detectorID == "" && member.rule == "" {
		payload, err = marshalCanonicalStruct(identity)
	} else {
		payload, err = marshalCanonicalStruct(struct {
			Base         any
			OriginalName string
			DetectorID   string
			Rule         string
		}{
			Base: identity, OriginalName: rawMemberOriginalName(member),
			DetectorID: member.detectorID, Rule: member.rule,
		})
	}
	if err != nil {
		panic("canonical raw member identity: " + err.Error())
	}
	return digestBytes(payload)
}

func rawMemberIdentityObservations(member rawMember) []Observation {
	if member.originalName != "" || member.detectorID != "" || member.rule != "" || len(member.observations) != 1 {
		return member.observations
	}
	observation := member.observations[0]
	if observation.DetectorID != "archive-tar-v1" || observation.Result != "ENTRY" {
		return member.observations
	}
	if kind, ok := singleFactValue(observation.Facts, "metadata_kind"); ok && kind != "" {
		return member.observations
	}
	if members, ok := singleFactValue(observation.Facts, "metadata_members"); !ok || members != "" {
		return member.observations
	}
	// Preserve the accepted F14 order-independent logical-member identity.
	// Physical metadata facts are bound elsewhere, but ordinary tar member
	// ordering remains a function of the same normalized four facts as v1's
	// original tar walker.
	filtered := Observation{DetectorID: observation.DetectorID, Result: observation.Result, Facts: make([]Fact, 0, 4)}
	for _, fact := range observation.Facts {
		switch fact.Key {
		case "format", "mode", "size", "typeflag":
			filtered.Facts = append(filtered.Facts, fact)
		}
	}
	return []Observation{filtered}
}

func rawMemberOriginalName(member rawMember) string {
	if member.originalName != "" {
		return member.originalName
	}
	return member.name
}

func appendOccurrenceEvidence(member *rawMember, logical, identity string, occurrence, count int) {
	finding := []Fact{
		{Key: "logical_path", Value: logical},
		{Key: "occurrence_count", Value: strconv.Itoa(count)},
		{Key: "occurrence_identity", Value: identity},
		{Key: "occurrence_number", Value: strconv.Itoa(occurrence)},
	}
	if len(member.observations) == 0 {
		member.observations = []Observation{{DetectorID: "archive-ar-v1", Result: "ENTRY", Facts: finding}}
		return
	}
	member.observations[0].Facts = append(member.observations[0].Facts, finding...)
	sortFacts(member.observations[0].Facts)
}

func uniqueDuplicatePath(
	logical, identity string,
	occurrence int,
	usedPaths, usedCollisionKeys map[string]struct{},
	limits LimitVector,
) VirtualPath {
	shortIdentity := strings.TrimPrefix(identity, "sha256:")
	if len(shortIdentity) > 16 {
		shortIdentity = shortIdentity[:16]
	}
	directory, base := path.Split(logical)
	extension := path.Ext(base)
	stem := strings.TrimSuffix(base, extension)
	for salt := 0; ; salt++ {
		marker := fmt.Sprintf("~curator-duplicate-%s-%06d-%03d", shortIdentity, occurrence, salt)
		candidateBase := stem + marker + extension
		candidate := directory + candidateBase
		validated, err := validateVirtualPath(candidate, limits)
		if err != nil {
			candidate = "$curator-duplicate-" + shortIdentity + fmt.Sprintf("-%06d-%03d", occurrence, salt)
			validated, err = validateVirtualPath(candidate, limits)
			if err != nil {
				continue
			}
		}
		if _, exists := usedPaths[validated.Canonical]; exists {
			continue
		}
		if _, collides := usedCollisionKeys[validated.CollisionKey]; collides {
			continue
		}
		return validated
	}
}

func (inspector *inspector) inspectMembers(containerIndex int, members []preparedMember, spool *os.File, depth int64, rejected bool) bool {
	containerPath := inspector.nodes[containerIndex].Path
	containerChain := inspector.nodes[containerIndex].ContainerChain
	memberChain := append(append([]string(nil), containerChain...), containerPath)
	for _, member := range members {
		fullPath := joinContainerPath(containerPath, member.path.Canonical)
		parentPath := containerPath
		if parent := path.Dir(member.path.Canonical); parent != "." {
			parentPath = joinContainerPath(containerPath, parent)
		}
		if member.raw.failureCode != "" {
			class := ClassOpaqueUnknown
			kind := member.raw.kind
			selectedDetector := ""
			variant := ""
			sha256Value := ""
			if member.raw.data != nil {
				sha256Value = member.raw.data.sha256
			}
			if member.raw.preclassified {
				class = member.raw.class
				selectedDetector = member.raw.detectorID
				variant = member.raw.variant
			}
			switch kind {
			case NodeLink:
				class = ClassLink
			case NodeSpecial:
				class = ClassSpecial
			}
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey: portableCollisionKey(fullPath), Kind: kind,
				Parent: parentPath, ContainerChain: append([]string(nil), memberChain...),
				Size: member.raw.declaredSize, SHA256: sha256Value, Mode: member.raw.mode,
				Observations:       append([]Observation(nil), member.raw.observations...),
				SelectedDetectorID: selectedDetector, Class: class, Variant: variant,
				Decision: DecisionReject, Rule: "container_entry_rejected",
				InspectionComplete: member.raw.failureCode == CodeArchiveUnsafeEntry || member.raw.failureCode == CodeArchiveUnsafePath,
			}
			inspector.nodes = append(inspector.nodes, node)
			inspector.addDiagnostic(Diagnostic{
				Code: member.raw.failureCode, Path: fullPath,
				OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey:       member.path.CollisionKey, Class: class, Variant: variant,
				DetectorID: selectedDetector, SHA256: sha256Value,
				ContainerChain: append([]string(nil), memberChain...),
				Size:           member.raw.declaredSize, Reason: member.raw.failureReason,
			})
			rejected = true
			continue
		}
		if member.raw.preclassified {
			if member.raw.data == nil || member.raw.data.sha256 == "" {
				inspector.addDiagnostic(Diagnostic{
					Code: CodePolicyInternalError, Path: fullPath,
					ContainerChain: append([]string(nil), memberChain...), Reason: "preclassified_member_identity_missing",
				})
				rejected = true
				continue
			}
			decision := decisionForClass(inspector.role, member.raw.class)
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey: portableCollisionKey(fullPath), Kind: NodeRegularFile,
				Parent: parentPath, ContainerChain: append([]string(nil), memberChain...),
				Size: member.raw.data.size, SHA256: member.raw.data.sha256, Mode: member.raw.mode,
				Observations:       append([]Observation(nil), member.raw.observations...),
				SelectedDetectorID: member.raw.detectorID, Class: member.raw.class, Variant: member.raw.variant,
				Decision: decision, Rule: member.raw.rule, InspectionComplete: true,
			}
			if node.SelectedDetectorID == "" {
				node.SelectedDetectorID = "archive-ar-v1"
			}
			if node.Rule == "" {
				node.Rule = "native_archive_metadata"
			}
			inspector.nodes = append(inspector.nodes, node)
			if decision == DecisionReject {
				inspector.addDiagnostic(inspector.classDiagnostic(node, "native_archive_metadata_forbidden"))
				rejected = true
			}
			continue
		}
		switch member.raw.kind {
		case NodeDirectory:
			class := ClassDirectory
			decision := DecisionDescend
			rule := "descend_directory"
			selectedDetector := ""
			observations := append([]Observation(nil), member.raw.observations...)
			if bundleClass, bundle := appleBundleClass(fullPath); bundle {
				class = bundleClass
				decision = decisionForClass(inspector.role, class)
				rule = "apple_bundle_forbidden"
				selectedDetector = "apple-bundle-path-v1"
				observations = append(observations, Observation{
					DetectorID: selectedDetector, Result: "MATCH",
					Facts: []Fact{{Key: "bundle_path", Value: fullPath}},
				})
			}
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey: portableCollisionKey(fullPath), Kind: NodeDirectory,
				Parent: parentPath, ContainerChain: append([]string(nil), memberChain...),
				Mode: member.raw.mode, Observations: observations,
				SelectedDetectorID: selectedDetector,
				Class:              class, Decision: decision, Rule: rule, InspectionComplete: true,
			}
			inspector.nodes = append(inspector.nodes, node)
			if decision == DecisionReject {
				inspector.addDiagnostic(inspector.classDiagnostic(node, "apple_bundle_forbidden"))
				rejected = true
			}
		case NodeLink, NodeSpecial:
			class := ClassLink
			if member.raw.kind == NodeSpecial {
				class = ClassSpecial
			}
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey: portableCollisionKey(fullPath), Kind: member.raw.kind,
				Parent: parentPath, ContainerChain: append([]string(nil), memberChain...),
				Mode:         member.raw.mode,
				Observations: append([]Observation(nil), member.raw.observations...),
				Class:        class, Decision: DecisionReject, Rule: "unsafe_entry", InspectionComplete: true,
			}
			inspector.nodes = append(inspector.nodes, node)
			inspector.addDiagnostic(Diagnostic{
				Code: CodeArchiveUnsafeEntry, Path: fullPath,
				OriginalNameBase64: originalNameBase64(rawMemberOriginalName(member.raw)),
				CollisionKey:       member.path.CollisionKey, Class: class,
				ContainerChain: append([]string(nil), memberChain...), Reason: "unsafe_entry_kind",
			})
			rejected = true
		case NodeRegularFile:
			memberBlob := member.raw.data
			if memberBlob == nil {
				memberCopy := member
				if diagnostic := inspector.materializeMember(&memberCopy, spool, "archive-zip-v1", fullPath, memberChain); diagnostic != nil {
					inspector.addDiagnostic(*diagnostic)
					rejected = true
					continue
				}
				memberBlob = memberCopy.raw.data
			}
			if inspector.inspectBlob(
				fullPath, rawMemberOriginalName(member.raw), parentPath, memberChain, *memberBlob,
				depth+1, member.raw.mode, member.raw.observations,
			) {
				rejected = true
			}
		default:
			inspector.addDiagnostic(Diagnostic{
				Code: CodePolicyInternalError, Path: fullPath,
				ContainerChain: append([]string(nil), memberChain...), Reason: "unknown_prepared_member_kind",
			})
			rejected = true
		}
	}
	return rejected
}

func (inspector *inspector) materializeMember(
	member *preparedMember,
	spool *os.File,
	detector, fullPath string,
	chain []string,
) *Diagnostic {
	if member.raw.kind != NodeRegularFile || member.raw.data != nil || member.raw.failureCode != "" {
		return nil
	}
	if spool == nil || member.raw.open == nil {
		diagnostic := unavailableDiagnostic(fullPath, chain, "member_reader_missing", nil)
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	if err := inspector.account.checkLeaf(member.raw.declaredSize); err != nil {
		diagnostic := diagnosticFromError(fullPath, chain, err)
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	if member.raw.declaredSize < 0 || member.raw.declaredSize > inspector.limits.MaxTotalEmittedBytes {
		diagnostic := diagnosticFromError(fullPath, chain, &limitFailure{
			name: "max_total_emitted_bytes", limit: inspector.limits.MaxTotalEmittedBytes,
			observed: member.raw.declaredSize,
		})
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	reader, err := member.raw.open()
	if err != nil {
		diagnostic := containerReadDiagnostic(fullPath, chain, detector, err)
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	captured, captureErr := inspector.store.appendBlob(inspector.ctx, spool, reader, member.raw.declaredSize)
	closeErr := reader.Close()
	if captureErr == nil {
		captureErr = closeErr
	}
	if captureErr != nil {
		diagnostic := containerReadDiagnostic(fullPath, chain, detector, captureErr)
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	if err := inspector.account.addEmitted(captured.size, member.raw.compressed); err != nil {
		diagnostic := diagnosticFromError(fullPath, chain, err)
		diagnostic.Size = member.raw.declaredSize
		return &diagnostic
	}
	member.raw.data = &captured
	member.raw.charged = true
	return nil
}

func isWheelContainer(containerPath string, profile ProfileID) bool {
	return profile == ProfilePythonSourceV1 && strings.HasSuffix(strings.ToLower(leafPath(containerPath)), ".whl")
}

func validateWheelRecord(containerPath string, chain []string, members []preparedMember) *Diagnostic {
	files := make(map[string]blob)
	recordPath := ""
	for _, member := range members {
		if member.raw.kind != NodeRegularFile || member.raw.failureCode != "" || member.raw.data == nil {
			continue
		}
		files[member.path.Canonical] = *member.raw.data
		if strings.HasSuffix(member.path.Canonical, ".dist-info/RECORD") {
			if recordPath != "" {
				return wheelDiagnostic(containerPath, chain, "multiple_wheel_record_files")
			}
			recordPath = member.path.Canonical
		}
	}
	if recordPath == "" {
		return wheelDiagnostic(containerPath, chain, "wheel_record_missing")
	}
	record := files[recordPath]
	reader := csv.NewReader(record.reader())
	reader.FieldsPerRecord = -1
	seen := make(map[string]struct{})
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) != 3 {
			return wheelDiagnostic(containerPath, chain, "wheel_record_malformed")
		}
		validated, err := ValidateVirtualPath(row[0])
		if err != nil || validated.Canonical != row[0] {
			return wheelDiagnostic(containerPath, chain, "wheel_record_path_invalid")
		}
		if _, duplicate := seen[row[0]]; duplicate {
			return wheelDiagnostic(containerPath, chain, "wheel_record_duplicate_path")
		}
		seen[row[0]] = struct{}{}
		member, exists := files[row[0]]
		if !exists {
			return wheelDiagnostic(containerPath, chain, "wheel_record_references_missing_member")
		}
		if row[0] == recordPath {
			if row[1] != "" || row[2] != "" {
				return wheelDiagnostic(containerPath, chain, "wheel_record_self_hash_must_be_empty")
			}
			continue
		}
		if !strings.HasPrefix(row[1], "sha256=") {
			return wheelDiagnostic(containerPath, chain, "wheel_record_hash_algorithm_unsupported")
		}
		digestBytesValue, err := hex.DecodeString(strings.TrimPrefix(member.sha256, "sha256:"))
		if err != nil || strings.TrimPrefix(row[1], "sha256=") != base64.RawURLEncoding.EncodeToString(digestBytesValue) {
			return wheelDiagnostic(containerPath, chain, "wheel_record_hash_mismatch")
		}
		size, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil || size != member.size {
			return wheelDiagnostic(containerPath, chain, "wheel_record_size_mismatch")
		}
	}
	if len(seen) != len(files) {
		return wheelDiagnostic(containerPath, chain, "wheel_record_incomplete")
	}
	return nil
}

func wheelDiagnostic(containerPath string, chain []string, reason string) *Diagnostic {
	diagnostic := Diagnostic{
		Code: CodeArchiveInvalid, Path: containerPath, DetectorID: "archive-zip-v1",
		ContainerChain: append([]string(nil), chain...), Reason: reason,
	}
	return &diagnostic
}

func parseTarNumber(field []byte) (int64, bool) {
	if len(field) == 0 {
		return 0, false
	}
	if field[0]&0x80 != 0 {
		negative := field[0]&0x40 != 0
		if negative {
			return 0, false
		}
		value := int64(field[0] & 0x3f)
		for _, character := range field[1:] {
			if value > (int64(^uint64(0)>>1)-int64(character))/256 {
				return 0, false
			}
			value = value*256 + int64(character)
		}
		return value, true
	}
	text := strings.Trim(string(field), " \x00")
	if text == "" {
		return 0, true
	}
	value, err := strconv.ParseInt(text, 8, 64)
	return value, err == nil && value >= 0
}
