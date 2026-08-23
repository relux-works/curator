package artifactpolicy

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- DEX format validation requires its specified SHA-1 field.
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"io"
	"math"
	"strings"
	"unicode/utf8"
)

const (
	elfTypeRel   = 1
	elfTypeExec  = 2
	elfTypeDyn   = 3
	elfPTLoad    = 1
	elfPTDynamic = 2
	elfPTInterp  = 3
	elfDTNull    = 0
	elfDTSoname  = 14
	elfDTFlags1  = 0x6ffffffb
	elfDF1PIE    = 0x08000000
)

func detectELF(item blob, uses []UseEdge) detectorCandidate {
	prefix, err := item.prefix(64)
	if err != nil || len(prefix) < 4 || !bytes.Equal(prefix[:4], []byte{0x7f, 'E', 'L', 'F'}) {
		return detectorCandidate{}
	}
	details, class, variant, err := parseELF(item, uses)
	observation := Observation{DetectorID: "elf-v1", Facts: details}
	if err != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: err.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: class, variant: variant, detectorID: "elf-v1"},
		observed:  &observation,
	}
}

func parseELF(item blob, uses []UseEdge) ([]Fact, ArtifactClass, string, error) {
	ident, err := readExactAt(item, 0, 16)
	if err != nil {
		return nil, "", "", fmt.Errorf("truncated ELF identification")
	}
	class := ident[4]
	encoding := ident[5]
	if (class != 1 && class != 2) || (encoding != 1 && encoding != 2) || ident[6] != 1 {
		return nil, "", "", fmt.Errorf("unsupported ELF class, encoding, or version")
	}
	order := binary.ByteOrder(binary.LittleEndian)
	if encoding == 2 {
		order = binary.BigEndian
	}
	headerSize := int64(52)
	programHeaderSize := uint16(32)
	if class == 2 {
		headerSize = 64
		programHeaderSize = 56
	}
	header, err := readExactAt(item, 0, headerSize)
	if err != nil {
		return nil, "", "", fmt.Errorf("truncated ELF header")
	}
	eType := order.Uint16(header[16:18])
	machine := order.Uint16(header[18:20])
	if order.Uint32(header[20:24]) != 1 {
		return nil, "", "", fmt.Errorf("invalid ELF header version")
	}
	var entry, programOffset uint64
	var flags uint32
	var declaredHeaderSize, declaredProgramSize, programCount uint16
	if class == 1 {
		entry = uint64(order.Uint32(header[24:28]))
		programOffset = uint64(order.Uint32(header[28:32]))
		flags = order.Uint32(header[36:40])
		declaredHeaderSize = order.Uint16(header[40:42])
		declaredProgramSize = order.Uint16(header[42:44])
		programCount = order.Uint16(header[44:46])
	} else {
		entry = order.Uint64(header[24:32])
		programOffset = order.Uint64(header[32:40])
		flags = order.Uint32(header[48:52])
		declaredHeaderSize = order.Uint16(header[52:54])
		declaredProgramSize = order.Uint16(header[54:56])
		programCount = order.Uint16(header[56:58])
	}
	if int64(declaredHeaderSize) != headerSize {
		return nil, "", "", fmt.Errorf("invalid ELF header size")
	}
	if programCount == 0xffff {
		return nil, "", "", fmt.Errorf("extended ELF program-header count is unsupported")
	}
	if programCount > 0 && declaredProgramSize != programHeaderSize {
		return nil, "", "", fmt.Errorf("invalid ELF program-header size")
	}
	programBytes, ok := checkedUint64Mul(uint64(programCount), uint64(programHeaderSize))
	itemSize := uint64(item.size) // #nosec G115 -- blob sizes are nonnegative by construction.
	if !ok || programOffset > itemSize || programBytes > itemSize-programOffset {
		return nil, "", "", fmt.Errorf("ELF program-header table is out of bounds")
	}

	programTypes := make([]string, 0, programCount)
	interpCount := 0
	hasDynamic := false
	hasSoname := false
	var flags1 uint64
	rawFacts := make([]Fact, 0, int(programCount)*5)
	for index := uint16(0); index < programCount; index++ {
		offset := int64(programOffset + uint64(index)*uint64(programHeaderSize)) // #nosec G115 -- table bounds were proved above.
		program, readErr := readExactAt(item, offset, int64(programHeaderSize))
		if readErr != nil {
			return nil, "", "", fmt.Errorf("read ELF program header: %w", readErr)
		}
		programType := order.Uint32(program[0:4])
		var fileOffset, fileSize, memorySize uint64
		if class == 1 {
			fileOffset = uint64(order.Uint32(program[4:8]))
			fileSize = uint64(order.Uint32(program[16:20]))
			memorySize = uint64(order.Uint32(program[20:24]))
		} else {
			fileOffset = order.Uint64(program[8:16])
			fileSize = order.Uint64(program[32:40])
			memorySize = order.Uint64(program[40:48])
		}
		if fileOffset > itemSize || fileSize > itemSize-fileOffset {
			return nil, "", "", fmt.Errorf("ELF segment is out of bounds")
		}
		if programType == elfPTLoad && fileSize > memorySize {
			return nil, "", "", fmt.Errorf("ELF load segment file size exceeds memory size")
		}
		factPrefix := fmt.Sprintf("program_header.%05d.", index)
		rawFacts = append(rawFacts,
			Fact{Key: factPrefix + "file_offset", Value: fmt.Sprint(fileOffset)},
			Fact{Key: factPrefix + "file_size", Value: fmt.Sprint(fileSize)},
			Fact{Key: factPrefix + "memory_size", Value: fmt.Sprint(memorySize)},
			Fact{Key: factPrefix + "type", Value: fmt.Sprint(programType)},
		)
		programTypes = append(programTypes, fmt.Sprint(programType))
		switch programType {
		case elfPTInterp:
			interpCount++
			if interpCount > 1 || fileSize < 2 || fileSize > 4_096 {
				return nil, "", "", fmt.Errorf("invalid ELF PT_INTERP cardinality or size")
			}
			interpreter, readErr := readExactAt(item, int64(fileOffset), int64(fileSize)) // #nosec G115 -- segment bounds were proved.
			if readErr != nil || interpreter[len(interpreter)-1] != 0 || bytes.IndexByte(interpreter[:len(interpreter)-1], 0) >= 0 || !utf8.Valid(interpreter[:len(interpreter)-1]) {
				return nil, "", "", fmt.Errorf("invalid ELF PT_INTERP value")
			}
			rawFacts = append(rawFacts, Fact{Key: factPrefix + "interpreter", Value: string(interpreter[:len(interpreter)-1])})
		case elfPTDynamic:
			if hasDynamic {
				return nil, "", "", fmt.Errorf("duplicate ELF PT_DYNAMIC segment")
			}
			hasDynamic = true
			entrySize := uint64(8)
			if class == 2 {
				entrySize = 16
			}
			if fileSize%entrySize != 0 {
				return nil, "", "", fmt.Errorf("misaligned ELF dynamic table")
			}
			terminated := false
			seenFlags1 := false
			seenSoname := false
			for dynamicOffset := uint64(0); dynamicOffset < fileSize; dynamicOffset += entrySize {
				entryBytes, readErr := readExactAt(item, int64(fileOffset+dynamicOffset), int64(entrySize)) // #nosec G115 -- segment bounds were proved.
				if readErr != nil {
					return nil, "", "", fmt.Errorf("read ELF dynamic table: %w", readErr)
				}
				var tag, value uint64
				if class == 1 {
					tag = uint64(order.Uint32(entryBytes[0:4]))
					value = uint64(order.Uint32(entryBytes[4:8]))
				} else {
					tag = order.Uint64(entryBytes[0:8])
					value = order.Uint64(entryBytes[8:16])
				}
				dynamicIndex := dynamicOffset / entrySize
				dynamicPrefix := fmt.Sprintf("dynamic.%05d.", dynamicIndex)
				rawFacts = append(rawFacts,
					Fact{Key: dynamicPrefix + "tag", Value: fmt.Sprint(tag)},
					Fact{Key: dynamicPrefix + "value", Value: fmt.Sprint(value)},
				)
				if terminated && (tag != elfDTNull || value != 0) {
					return nil, "", "", fmt.Errorf("non-null ELF dynamic entry follows DT_NULL")
				}
				switch tag {
				case elfDTNull:
					terminated = true
				case elfDTFlags1:
					if seenFlags1 {
						return nil, "", "", fmt.Errorf("duplicate ELF DT_FLAGS_1")
					}
					seenFlags1 = true
					flags1 = value
				case elfDTSoname:
					if seenSoname {
						return nil, "", "", fmt.Errorf("duplicate ELF DT_SONAME")
					}
					seenSoname = true
					hasSoname = true
				}
			}
			if !terminated {
				return nil, "", "", fmt.Errorf("unterminated ELF dynamic table")
			}
		}
	}

	details := facts(map[string]any{
		"abi_version": ident[8], "class": class, "data_encoding": encoding,
		"df_1_pie": flags1&elfDF1PIE != 0, "dt_flags_1": flags1,
		"dt_soname_present": hasSoname, "e_entry": entry, "e_flags": flags,
		"e_machine": machine, "e_type": eType, "os_abi": ident[7],
		"program_header_types": strings.Join(programTypes, ","),
		"pt_dynamic":           hasDynamic, "pt_interp_count": interpCount,
	})
	details = append(details, useFacts(uses)...)
	details = append(details, rawFacts...)
	sortFacts(details)
	switch eType {
	case elfTypeRel:
		return details, ClassNativeObject, "elf.relocatable", nil
	case elfTypeExec:
		return details, ClassNativeExecutable, "elf.executable", nil
	case elfTypeDyn:
		executeEdges, linkEdges := countUses(uses)
		if flags1&elfDF1PIE != 0 {
			variant := "elf.pie.no_interpreter"
			if interpCount == 1 {
				variant = "elf.pie.interpreter"
			}
			return details, ClassNativeExecutable, variant, nil
		}
		if executeEdges > 0 && linkEdges > 0 {
			return details, ClassELFETDYNAmbiguous, "use_conflict", nil
		}
		if executeEdges > 1 || linkEdges > 1 {
			return details, ClassELFETDYNAmbiguous, "duplicate_use_edges", nil
		}
		if executeEdges == 1 && linkEdges == 0 && interpCount == 1 && !hasSoname {
			return details, ClassNativeExecutable, "elf.et_dyn.executable_by_use", nil
		}
		if executeEdges == 0 && interpCount == 0 && (hasSoname || linkEdges == 1) {
			return details, ClassNativeLibraryDynamic, "elf.shared_object", nil
		}
		reason := "insufficient_evidence"
		if interpCount > 0 && hasSoname {
			reason = "interp_soname_conflict"
		}
		return details, ClassELFETDYNAmbiguous, reason, nil
	default:
		return details, "", "", fmt.Errorf("unsupported ELF e_type %d", eType)
	}
}

func countUses(uses []UseEdge) (execute, link int) {
	for _, use := range uses {
		switch use.Kind {
		case UseExecute:
			execute++
		case UseLinkOrLoad:
			link++
		}
	}
	return execute, link
}

func detectPECOFF(item blob) detectorCandidate {
	prefix, err := item.prefix(128)
	if err != nil {
		return detectorCandidate{}
	}
	class, variant, details, matched, parseErr := parsePECOFF(item, prefix)
	if !matched {
		return detectorCandidate{}
	}
	observation := Observation{DetectorID: "pe-coff-v1", Facts: details}
	if parseErr != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: parseErr.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: class, variant: variant, detectorID: "pe-coff-v1"},
		observed:  &observation,
	}
}

func parsePECOFF(item blob, prefix []byte) (ArtifactClass, string, []Fact, bool, error) {
	itemSize := uint64(item.size) // #nosec G115 -- blob sizes are nonnegative by construction.
	if len(prefix) >= 2 && string(prefix[:2]) == "MZ" {
		if item.size < 64 {
			return "", "", nil, true, fmt.Errorf("truncated DOS header")
		}
		dos, err := readExactAt(item, 0, 64)
		if err != nil {
			return "", "", nil, true, err
		}
		peOffset := int64(binary.LittleEndian.Uint32(dos[0x3c:0x40]))
		if peOffset < 64 || peOffset > item.size-24 {
			return "", "", nil, true, fmt.Errorf("PE header offset is out of bounds")
		}
		header, err := readExactAt(item, peOffset, 24)
		if err != nil || string(header[:4]) != "PE\x00\x00" {
			return "", "", nil, true, fmt.Errorf("invalid PE signature")
		}
		machine := binary.LittleEndian.Uint16(header[4:6])
		sections := binary.LittleEndian.Uint16(header[6:8])
		optionalSize := binary.LittleEndian.Uint16(header[20:22])
		characteristics := binary.LittleEndian.Uint16(header[22:24])
		if !knownCOFFMachine(uint64(machine)) || sections == 0 || sections > 96 || optionalSize < 2 {
			return "", "", nil, true, fmt.Errorf("invalid PE COFF header")
		}
		optional, err := readExactAt(item, peOffset+24, int64(optionalSize))
		if err != nil {
			return "", "", nil, true, fmt.Errorf("truncated PE optional header")
		}
		magic := binary.LittleEndian.Uint16(optional[:2])
		if magic != 0x10b && magic != 0x20b {
			return "", "", nil, true, fmt.Errorf("unsupported PE optional-header magic")
		}
		minimumOptionalSize := 96
		directoryCountOffset := 92
		if magic == 0x20b {
			minimumOptionalSize = 112
			directoryCountOffset = 108
		}
		if len(optional) < minimumOptionalSize {
			return "", "", nil, true, fmt.Errorf("truncated PE optional header")
		}
		directoryCount := binary.LittleEndian.Uint32(optional[directoryCountOffset : directoryCountOffset+4])
		if directoryCount > 16 || uint64(minimumOptionalSize)+uint64(directoryCount)*8 > uint64(len(optional)) {
			return "", "", nil, true, fmt.Errorf("PE data-directory table is out of bounds")
		}
		sectionTable := peOffset + 24 + int64(optionalSize)
		if sectionTable > item.size || int64(sections)*40 > item.size-sectionTable {
			return "", "", nil, true, fmt.Errorf("PE section table is out of bounds")
		}
		details := facts(map[string]any{
			"characteristics": characteristics, "data_directory_count": directoryCount,
			"machine": machine, "optional_magic": magic, "sections": sections,
		})
		for index := uint16(0); index < sections; index++ {
			section, sectionErr := readExactAt(item, sectionTable+int64(index)*40, 40)
			if sectionErr != nil {
				return "", "", nil, true, fmt.Errorf("read PE section %d: %w", index, sectionErr)
			}
			rawSize := binary.LittleEndian.Uint32(section[16:20])
			rawOffset := binary.LittleEndian.Uint32(section[20:24])
			if rawSize > 0 && (uint64(rawOffset) > itemSize || uint64(rawSize) > itemSize-uint64(rawOffset)) {
				return "", "", nil, true, fmt.Errorf("PE section %d raw range is out of bounds", index)
			}
			prefix := fmt.Sprintf("section.%05d.", index)
			details = append(details,
				Fact{Key: prefix + "raw_offset", Value: fmt.Sprint(rawOffset)},
				Fact{Key: prefix + "raw_size", Value: fmt.Sprint(rawSize)},
			)
		}
		sortFacts(details)
		class := ClassNativeExecutable
		variant := "pe.image"
		if characteristics&0x2000 != 0 {
			class = ClassNativeLibraryDynamic
			variant = "pe.dll"
		}
		return class, variant, details, true, nil
	}
	if len(prefix) < 20 {
		return "", "", nil, false, nil
	}
	machine := binary.LittleEndian.Uint16(prefix[0:2])
	sections := binary.LittleEndian.Uint16(prefix[2:4])
	optionalSize := binary.LittleEndian.Uint16(prefix[16:18])
	if !knownCOFFMachine(uint64(machine)) || sections == 0 || sections > 96 || optionalSize != 0 {
		return "", "", nil, false, nil
	}
	if int64(20+sections*40) > item.size {
		return "", "", nil, true, fmt.Errorf("COFF section table is out of bounds")
	}
	details := facts(map[string]any{
		"machine": machine, "sections": sections,
	})
	for index := uint16(0); index < sections; index++ {
		section, err := readExactAt(item, 20+int64(index)*40, 40)
		if err != nil {
			return "", "", nil, true, err
		}
		rawSize := binary.LittleEndian.Uint32(section[16:20])
		rawOffset := binary.LittleEndian.Uint32(section[20:24])
		if rawSize > 0 && (uint64(rawOffset) > itemSize || uint64(rawSize) > itemSize-uint64(rawOffset)) {
			return "", "", nil, true, fmt.Errorf("COFF section %d raw range is out of bounds", index)
		}
		prefix := fmt.Sprintf("section.%05d.", index)
		details = append(details,
			Fact{Key: prefix + "raw_offset", Value: fmt.Sprint(rawOffset)},
			Fact{Key: prefix + "raw_size", Value: fmt.Sprint(rawSize)},
		)
	}
	sortFacts(details)
	return ClassNativeObject, "coff.object", details, true, nil
}

func knownCOFFMachine(machine uint64) bool {
	switch machine {
	case 0x014c, 0x0166, 0x0169, 0x01c0, 0x01c2, 0x01c4, 0x01d3,
		0x01f0, 0x01f1, 0x0200, 0x0266, 0x0284, 0x0366, 0x0466,
		0x0520, 0x0ebc, 0x8664, 0x9041, 0xaa64:
		return true
	default:
		return false
	}
}

func detectMachO(item blob) detectorCandidate {
	prefix, err := item.prefix(32)
	if err != nil || len(prefix) < 4 {
		return detectorCandidate{}
	}
	magic := binary.BigEndian.Uint32(prefix[:4])
	if !machMagic(magic) {
		return detectorCandidate{}
	}
	class, variant, details, err := parseMachO(item)
	observation := Observation{DetectorID: "macho-v1", Facts: details}
	if err != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: err.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: class, variant: variant, detectorID: "macho-v1"},
		observed:  &observation,
	}
}

func machMagic(magic uint32) bool {
	switch magic {
	case 0xfeedface, 0xcefaedfe, 0xfeedfacf, 0xcffaedfe,
		0xcafebabe, 0xbebafeca, 0xcafebabf, 0xbfbafeca:
		return true
	default:
		return false
	}
}

func parseMachO(item blob) (ArtifactClass, string, []Fact, error) {
	magicBytes, err := readExactAt(item, 0, 4)
	if err != nil {
		return "", "", nil, err
	}
	magic := binary.BigEndian.Uint32(magicBytes)
	switch magic {
	case 0xcafebabe, 0xbebafeca, 0xcafebabf, 0xbfbafeca:
		return parseFatMachO(item, magic)
	default:
		return parseThinMachO(item, magic)
	}
}

func parseThinMachO(item blob, magic uint32) (ArtifactClass, string, []Fact, error) {
	itemSize := uint64(item.size) // #nosec G115 -- blob sizes are nonnegative by construction.
	var order binary.ByteOrder
	is64 := false
	switch magic {
	case 0xfeedface:
		order = binary.BigEndian
	case 0xcefaedfe:
		order = binary.LittleEndian
	case 0xfeedfacf:
		order = binary.BigEndian
		is64 = true
	case 0xcffaedfe:
		order = binary.LittleEndian
		is64 = true
	default:
		return "", "", nil, fmt.Errorf("unsupported Mach-O magic")
	}
	headerSize := int64(28)
	if is64 {
		headerSize = 32
	}
	header, err := readExactAt(item, 0, headerSize)
	if err != nil {
		return "", "", nil, fmt.Errorf("truncated Mach-O header")
	}
	fileType := order.Uint32(header[12:16])
	commands := order.Uint32(header[16:20])
	commandBytes := order.Uint32(header[20:24])
	if uint64(headerSize)+uint64(commandBytes) > itemSize || commands > 65_535 {
		return "", "", nil, fmt.Errorf("Mach-O load commands are out of bounds")
	}
	commandData, err := readExactAt(item, headerSize, int64(commandBytes))
	if err != nil {
		return "", "", nil, fmt.Errorf("read Mach-O load commands: %w", err)
	}
	commandOffset := uint64(0)
	for index := uint32(0); index < commands; index++ {
		if uint64(len(commandData))-commandOffset < 8 {
			return "", "", nil, fmt.Errorf("Mach-O load command %d is truncated", index)
		}
		commandSize := uint64(order.Uint32(commandData[commandOffset+4 : commandOffset+8]))
		if commandSize < 8 || commandSize%4 != 0 || commandSize > uint64(len(commandData))-commandOffset {
			return "", "", nil, fmt.Errorf("Mach-O load command %d has an invalid size", index)
		}
		commandOffset += commandSize
	}
	if commandOffset != uint64(len(commandData)) {
		return "", "", nil, fmt.Errorf("Mach-O load command byte count is inconsistent")
	}
	var class ArtifactClass
	var variant string
	switch fileType {
	case 1:
		class, variant = ClassNativeObject, "macho.object"
	case 2, 3, 4, 5, 7:
		class, variant = ClassNativeExecutable, "macho.executable"
	case 6:
		class, variant = ClassNativeLibraryDynamic, "macho.dylib"
	case 8:
		class, variant = ClassNativeLibraryDynamic, "macho.bundle"
	default:
		return "", "", nil, fmt.Errorf("unsupported Mach-O file type %d", fileType)
	}
	return class, variant, facts(map[string]any{
		"file_type": fileType, "is_64": is64, "load_command_bytes": commandBytes,
		"load_commands": commands,
	}), nil
}

func parseFatMachO(item blob, magic uint32) (ArtifactClass, string, []Fact, error) {
	itemSize := uint64(item.size) // #nosec G115 -- blob sizes are nonnegative by construction.
	order := binary.ByteOrder(binary.BigEndian)
	is64 := magic == 0xcafebabf || magic == 0xbfbafeca
	if magic == 0xbebafeca || magic == 0xbfbafeca {
		order = binary.LittleEndian
	}
	header, err := readExactAt(item, 0, 8)
	if err != nil {
		return "", "", nil, fmt.Errorf("truncated fat Mach-O header")
	}
	count := order.Uint32(header[4:8])
	if count == 0 || count > 64 {
		return "", "", nil, fmt.Errorf("invalid fat Mach-O slice count")
	}
	entrySize := int64(20)
	if is64 {
		entrySize = 32
	}
	if int64(count) > (item.size-8)/entrySize {
		return "", "", nil, fmt.Errorf("fat Mach-O table is out of bounds")
	}
	selectedClass := ArtifactClass("")
	variants := make([]string, 0, count)
	ranges := make([][2]uint64, 0, count)
	sliceFacts := make([]Fact, 0, count*5)
	tableEnd := uint64(8) + uint64(count)*uint64(entrySize)
	for index := uint32(0); index < count; index++ {
		entry, readErr := readExactAt(item, 8+int64(index)*entrySize, entrySize)
		if readErr != nil {
			return "", "", nil, readErr
		}
		var offset, size uint64
		if is64 {
			offset = order.Uint64(entry[8:16])
			size = order.Uint64(entry[16:24])
		} else {
			offset = uint64(order.Uint32(entry[8:12]))
			size = uint64(order.Uint32(entry[12:16]))
		}
		alignmentPower := order.Uint32(entry[16:20])
		if is64 {
			alignmentPower = order.Uint32(entry[24:28])
		}
		if size == 0 || offset < tableEnd || offset > itemSize || size > itemSize-offset {
			return "", "", nil, fmt.Errorf("fat Mach-O slice is out of bounds")
		}
		if alignmentPower > 63 || (alignmentPower > 0 && offset%(uint64(1)<<alignmentPower) != 0) {
			return "", "", nil, fmt.Errorf("fat Mach-O slice alignment is invalid")
		}
		for _, existing := range ranges {
			if offset < existing[1] && existing[0] < offset+size {
				return "", "", nil, fmt.Errorf("fat Mach-O slices overlap")
			}
		}
		ranges = append(ranges, [2]uint64{offset, offset + size})
		slice := blob{file: item.file, offset: item.offset + int64(offset), size: int64(size)} // #nosec G115 -- bounds were proved above.
		sliceMagicBytes, readErr := readExactAt(slice, 0, 4)
		if readErr != nil {
			return "", "", nil, readErr
		}
		sliceClass, sliceVariant, _, parseErr := parseThinMachO(slice, binary.BigEndian.Uint32(sliceMagicBytes))
		if parseErr != nil {
			return "", "", nil, fmt.Errorf("invalid fat Mach-O slice %d: %w", index, parseErr)
		}
		if selectedClass == "" || machoSliceClassPriority(sliceClass) < machoSliceClassPriority(selectedClass) {
			selectedClass = sliceClass
		}
		variants = append(variants, sliceVariant)
		prefix := fmt.Sprintf("slice.%03d.", index)
		sliceFacts = append(sliceFacts,
			Fact{Key: prefix + "alignment_power", Value: fmt.Sprint(alignmentPower)},
			Fact{Key: prefix + "class", Value: string(sliceClass)},
			Fact{Key: prefix + "offset", Value: fmt.Sprint(offset)},
			Fact{Key: prefix + "size", Value: fmt.Sprint(size)},
			Fact{Key: prefix + "variant", Value: sliceVariant},
		)
	}
	details := facts(map[string]any{
		"fat_64": is64, "slice_count": count, "slice_variants": strings.Join(variants, ","),
	})
	details = append(details, sliceFacts...)
	sortFacts(details)
	return selectedClass, "macho.universal", details, nil
}

func machoSliceClassPriority(class ArtifactClass) int {
	// Universal images are one physical native artifact. Resolve mixed slices
	// with an explicit deny-dominant order instead of inheriting slice-table
	// order. In particular, any executable slice makes the universal artifact
	// a native executable as required by the closed taxonomy.
	switch class {
	case ClassNativeExecutable:
		return 0
	case ClassNativeObject:
		return 1
	case ClassNativeLibraryDynamic:
		return 2
	default:
		return 9
	}
}

func detectJVMClass(item blob) detectorCandidate {
	prefix, err := item.prefix(10)
	if err != nil || len(prefix) < 4 || !bytes.Equal(prefix[:4], []byte{0xca, 0xfe, 0xba, 0xbe}) {
		return detectorCandidate{}
	}
	details, err := parseJVMClass(item)
	observation := Observation{DetectorID: "jvm-class-v1", Facts: details}
	if err != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: err.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: ClassJVMBytecode, variant: "jvm.class", detectorID: "jvm-class-v1"},
		observed:  &observation,
	}
}

func parseJVMClass(item blob) ([]Fact, error) {
	if item.size > 256<<20 {
		return nil, fmt.Errorf("class file exceeds leaf limit")
	}
	payload, err := readExactAt(item, 0, item.size)
	if err != nil {
		return nil, err
	}
	cursor := byteCursor{payload: payload}
	if cursor.u4() != 0xcafebabe {
		return nil, fmt.Errorf("invalid class magic")
	}
	minor := cursor.u2()
	major := cursor.u2()
	constantCount := cursor.u2()
	if cursor.err != nil || major < 45 || major > 80 || constantCount < 1 {
		return nil, fmt.Errorf("invalid class version or constant-pool count")
	}
	pool := make([]jvmConstant, constantCount)
	for index := uint16(1); index < constantCount; index++ {
		tag := cursor.u1()
		entry := jvmConstant{tag: tag}
		switch tag {
		case 1:
			length := cursor.u2()
			entry.text = string(cursor.take(uint32(length)))
		case 3, 4:
			cursor.skip(4)
		case 5, 6:
			cursor.skip(8)
			if index+1 >= constantCount {
				return nil, fmt.Errorf("wide constant occupies a missing reserved slot")
			}
			if cursor.err != nil {
				return nil, fmt.Errorf("truncated class constant pool")
			}
			pool[index] = entry
			index++
			continue
		case 7, 8, 16, 19, 20:
			entry.first = cursor.u2()
		case 9, 10, 11, 12, 17, 18:
			entry.first = cursor.u2()
			entry.second = cursor.u2()
		case 15:
			entry.referenceKind = cursor.u1()
			entry.first = cursor.u2()
		default:
			return nil, fmt.Errorf("invalid constant-pool tag %d", tag)
		}
		if cursor.err != nil {
			return nil, fmt.Errorf("truncated class constant pool")
		}
		pool[index] = entry
	}
	if err := validateJVMConstantPool(pool); err != nil {
		return nil, err
	}
	accessFlags := cursor.u2()
	thisClass := cursor.u2()
	superClass := cursor.u2()
	if !jvmConstantHasTag(pool, thisClass, 7) {
		return nil, fmt.Errorf("invalid this_class constant-pool index")
	}
	thisName := jvmClassName(pool, thisClass)
	if thisName == "" {
		return nil, fmt.Errorf("this_class has an empty name")
	}
	if superClass == 0 {
		if thisName != "java/lang/Object" && accessFlags&0x8000 == 0 {
			return nil, fmt.Errorf("invalid zero super_class index")
		}
	} else if !jvmConstantHasTag(pool, superClass, 7) {
		return nil, fmt.Errorf("invalid super_class constant-pool index")
	}
	interfaces := cursor.u2()
	for index := uint16(0); index < interfaces; index++ {
		if !jvmConstantHasTag(pool, cursor.u2(), 7) {
			return nil, fmt.Errorf("invalid interface constant-pool index")
		}
	}
	fields := cursor.u2()
	for index := uint16(0); index < fields; index++ {
		if err := parseJVMMember(&cursor, pool); err != nil {
			return nil, fmt.Errorf("field %d: %w", index, err)
		}
	}
	methods := cursor.u2()
	for index := uint16(0); index < methods; index++ {
		if err := parseJVMMember(&cursor, pool); err != nil {
			return nil, fmt.Errorf("method %d: %w", index, err)
		}
	}
	attributes, err := parseJVMAttributes(&cursor, pool)
	if err != nil {
		return nil, fmt.Errorf("class attributes: %w", err)
	}
	if uint64(len(payload)) > math.MaxUint32 {
		return nil, fmt.Errorf("class structure exceeds supported size")
	}
	if cursor.err != nil || cursor.offset != uint32(len(payload)) { // #nosec G115 -- length was bounded to MaxUint32 immediately above.
		return nil, fmt.Errorf("truncated class structure or trailing data")
	}
	return facts(map[string]any{
		"access_flags": accessFlags, "attributes": attributes,
		"constant_pool_count": constantCount, "fields": fields,
		"interfaces": interfaces, "major": major, "methods": methods,
		"minor": minor, "super_class": superClass, "this_class": thisClass,
		"this_class_name": thisName,
	}), nil
}

type jvmConstant struct {
	tag           byte
	first         uint16
	second        uint16
	referenceKind byte
	text          string
}

func validateJVMConstantPool(pool []jvmConstant) error {
	for index := 1; index < len(pool); index++ {
		entry := pool[index]
		switch entry.tag {
		case 0, 1, 3, 4, 5, 6:
		case 7, 8, 16, 19, 20:
			if !jvmConstantHasTag(pool, entry.first, 1) {
				return fmt.Errorf("constant-pool entry %d has an invalid UTF-8 reference", index)
			}
		case 9, 10, 11:
			if !jvmConstantHasTag(pool, entry.first, 7) || !jvmConstantHasTag(pool, entry.second, 12) {
				return fmt.Errorf("constant-pool member reference %d has invalid indexes", index)
			}
		case 12:
			if !jvmConstantHasTag(pool, entry.first, 1) || !jvmConstantHasTag(pool, entry.second, 1) {
				return fmt.Errorf("constant-pool name-and-type %d has invalid indexes", index)
			}
		case 15:
			if !validJVMMethodHandleReference(pool, entry.referenceKind, entry.first) {
				return fmt.Errorf("constant-pool method handle %d has an invalid reference", index)
			}
		case 17, 18:
			if !jvmConstantHasTag(pool, entry.second, 12) {
				return fmt.Errorf("constant-pool dynamic entry %d has an invalid name-and-type index", index)
			}
		default:
			return fmt.Errorf("constant-pool entry %d has unsupported tag %d", index, entry.tag)
		}
	}
	return nil
}

func validJVMMethodHandleReference(pool []jvmConstant, kind byte, index uint16) bool {
	switch kind {
	case 1, 2, 3, 4:
		return jvmConstantHasTag(pool, index, 9)
	case 5, 8:
		return jvmConstantHasTag(pool, index, 10)
	case 6, 7:
		return jvmConstantHasTag(pool, index, 10, 11)
	case 9:
		return jvmConstantHasTag(pool, index, 11)
	default:
		return false
	}
}

func jvmConstantHasTag(pool []jvmConstant, index uint16, tags ...byte) bool {
	if index == 0 || int(index) >= len(pool) {
		return false
	}
	for _, tag := range tags {
		if pool[index].tag == tag {
			return true
		}
	}
	return false
}

func jvmClassName(pool []jvmConstant, classIndex uint16) string {
	if !jvmConstantHasTag(pool, classIndex, 7) {
		return ""
	}
	nameIndex := pool[classIndex].first
	if !jvmConstantHasTag(pool, nameIndex, 1) {
		return ""
	}
	return pool[nameIndex].text
}

func parseJVMMember(cursor *byteCursor, pool []jvmConstant) error {
	cursor.skip(2)
	name := cursor.u2()
	descriptor := cursor.u2()
	if !jvmConstantHasTag(pool, name, 1) || pool[name].text == "" ||
		!jvmConstantHasTag(pool, descriptor, 1) || pool[descriptor].text == "" {
		return fmt.Errorf("invalid name or descriptor constant-pool index")
	}
	_, err := parseJVMAttributes(cursor, pool)
	return err
}

func parseJVMAttributes(cursor *byteCursor, pool []jvmConstant) (uint16, error) {
	count := cursor.u2()
	for index := uint16(0); index < count; index++ {
		name := cursor.u2()
		if !jvmConstantHasTag(pool, name, 1) || pool[name].text == "" {
			return 0, fmt.Errorf("attribute %d has an invalid name index", index)
		}
		length := cursor.u4()
		cursor.skip(length)
	}
	if cursor.err != nil {
		return 0, fmt.Errorf("truncated attribute table")
	}
	return count, nil
}

type byteCursor struct {
	payload []byte
	offset  uint32
	err     error
}

func (cursor *byteCursor) u1() byte {
	if cursor.err != nil || uint64(cursor.offset)+1 > uint64(len(cursor.payload)) {
		cursor.err = io.ErrUnexpectedEOF
		return 0
	}
	value := cursor.payload[cursor.offset]
	cursor.offset++
	return value
}

func (cursor *byteCursor) u2() uint16 {
	if cursor.err != nil || uint64(cursor.offset)+2 > uint64(len(cursor.payload)) {
		cursor.err = io.ErrUnexpectedEOF
		return 0
	}
	value := binary.BigEndian.Uint16(cursor.payload[cursor.offset : cursor.offset+2])
	cursor.offset += 2
	return value
}

func (cursor *byteCursor) u4() uint32 {
	if cursor.err != nil || uint64(cursor.offset)+4 > uint64(len(cursor.payload)) {
		cursor.err = io.ErrUnexpectedEOF
		return 0
	}
	value := binary.BigEndian.Uint32(cursor.payload[cursor.offset : cursor.offset+4])
	cursor.offset += 4
	return value
}

func (cursor *byteCursor) skip(amount uint32) {
	if cursor.err != nil || uint64(cursor.offset)+uint64(amount) > uint64(len(cursor.payload)) {
		cursor.err = io.ErrUnexpectedEOF
		return
	}
	cursor.offset += amount
}

func (cursor *byteCursor) take(amount uint32) []byte {
	if cursor.err != nil || uint64(cursor.offset)+uint64(amount) > uint64(len(cursor.payload)) {
		cursor.err = io.ErrUnexpectedEOF
		return nil
	}
	value := cursor.payload[cursor.offset : cursor.offset+amount]
	cursor.offset += amount
	return value
}

func detectDEX(item blob) detectorCandidate {
	prefix, err := item.prefix(112)
	if err != nil || len(prefix) < 8 || string(prefix[:4]) != "dex\n" || prefix[7] != 0 {
		return detectorCandidate{}
	}
	details, err := parseDEX(item)
	observation := Observation{DetectorID: "dex-v1", Facts: details}
	if err != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: err.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: ClassJVMBytecode, variant: "android.dex", detectorID: "dex-v1"},
		observed:  &observation,
	}
}

func parseDEX(item blob) ([]Fact, error) {
	if item.size < 112 || item.size > 256<<20 {
		return nil, fmt.Errorf("invalid DEX size")
	}
	payload, err := readExactAt(item, 0, item.size)
	if err != nil {
		return nil, err
	}
	if binary.LittleEndian.Uint32(payload[32:36]) != uint32(item.size) ||
		binary.LittleEndian.Uint32(payload[36:40]) != 112 ||
		binary.LittleEndian.Uint32(payload[40:44]) != 0x12345678 {
		return nil, fmt.Errorf("invalid DEX header sizes or endian tag")
	}
	signature := sha1.Sum(payload[32:]) // #nosec G401 -- DEX mandates SHA-1 for its structural signature field.
	if !bytes.Equal(signature[:], payload[12:32]) || adler32.Checksum(payload[12:]) != binary.LittleEndian.Uint32(payload[8:12]) {
		return nil, fmt.Errorf("DEX checksum or signature mismatch")
	}
	return facts(map[string]any{"version": string(payload[4:7])}), nil
}

func detectWebAssembly(item blob) detectorCandidate {
	prefix, err := item.prefix(8)
	if err != nil || len(prefix) < 4 || !bytes.Equal(prefix[:4], []byte{0, 'a', 's', 'm'}) {
		return detectorCandidate{}
	}
	details, err := parseWebAssembly(item)
	observation := Observation{DetectorID: "wasm-v1", Facts: details}
	if err != nil {
		observation.Result = "ERROR"
		observation.Facts = append(observation.Facts, Fact{Key: "error", Value: err.Error()})
		sortFacts(observation.Facts)
		return detectorCandidate{observed: &observation}
	}
	observation.Result = "MATCH"
	return detectorCandidate{
		detection: &detection{class: ClassWebAssembly, variant: "wasm.core", detectorID: "wasm-v1"},
		observed:  &observation,
	}
}

func parseWebAssembly(item blob) ([]Fact, error) {
	if item.size < 8 || item.size > 256<<20 {
		return nil, fmt.Errorf("invalid WebAssembly size")
	}
	payload, err := readExactAt(item, 0, item.size)
	if err != nil {
		return nil, err
	}
	version := binary.LittleEndian.Uint32(payload[4:8])
	if version != 1 {
		return nil, fmt.Errorf("unsupported WebAssembly version %d", version)
	}
	offset := 8
	lastSection := byte(0)
	sectionCount := 0
	for offset < len(payload) {
		section := payload[offset]
		offset++
		length, consumed, ok := readVarUint32(payload[offset:])
		if !ok {
			return nil, fmt.Errorf("invalid WebAssembly section length")
		}
		offset += consumed
		if section > 12 || (section != 0 && section <= lastSection) {
			return nil, fmt.Errorf("invalid WebAssembly section order or identifier")
		}
		if section != 0 {
			lastSection = section
		}
		remaining := len(payload) - offset
		if uint64(length) > uint64(remaining) { // #nosec G115 -- remaining is nonnegative because offset is bounded by the loop.
			return nil, fmt.Errorf("WebAssembly section is out of bounds")
		}
		offset += int(length)
		sectionCount++
	}
	return facts(map[string]any{"section_count": sectionCount, "version": version}), nil
}

func readVarUint32(payload []byte) (uint32, int, bool) {
	var value uint32
	for index := 0; index < 5 && index < len(payload); index++ {
		current := payload[index]
		if index == 4 && current&0xf0 != 0 {
			return 0, 0, false
		}
		value |= uint32(current&0x7f) << (7 * index)
		if current&0x80 == 0 {
			return value, index + 1, true
		}
	}
	return 0, 0, false
}

func detectLLVMBitcode(item blob) detectorCandidate {
	itemSize := uint64(item.size) // #nosec G115 -- blob sizes are nonnegative by construction.
	prefix, err := item.prefix(20)
	if err != nil || len(prefix) < 4 {
		return detectorCandidate{}
	}
	variant := ""
	details := []Fact{}
	switch {
	case bytes.Equal(prefix[:4], []byte{'B', 'C', 0xc0, 0xde}):
		variant = "llvm.bitcode.raw"
		if err := validateLLVMBitstream(item, 0, item.size); err != nil {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: err.Error()}},
			}}
		}
	case bytes.Equal(prefix[:4], []byte{0xde, 0xc0, 0x17, 0x0b}):
		if len(prefix) < 20 {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "truncated LLVM bitcode wrapper"}},
			}}
		}
		offset := binary.LittleEndian.Uint32(prefix[8:12])
		size := binary.LittleEndian.Uint32(prefix[12:16])
		if offset < 20 || offset%4 != 0 || uint64(offset) > itemSize ||
			uint64(size) > itemSize-uint64(offset) || uint64(offset)+uint64(size) != itemSize {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "LLVM bitcode wrapper range is invalid"}},
			}}
		}
		wrapped, readErr := readExactAt(item, int64(offset), 4)
		if readErr != nil || !bytes.Equal(wrapped, []byte{'B', 'C', 0xc0, 0xde}) {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "LLVM bitcode wrapper payload is invalid"}},
			}}
		}
		padding, readErr := readExactAt(item, 20, int64(offset)-20)
		if readErr != nil || !allZero(padding) {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "LLVM bitcode wrapper padding is invalid"}},
			}}
		}
		if validationErr := validateLLVMBitstream(item, int64(offset), int64(size)); validationErr != nil {
			return detectorCandidate{observed: &Observation{
				DetectorID: "llvm-bitcode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: validationErr.Error()}},
			}}
		}
		variant = "llvm.bitcode.wrapper"
		details = facts(map[string]any{"offset": offset, "size": size})
	default:
		return detectorCandidate{}
	}
	observation := Observation{DetectorID: "llvm-bitcode-v1", Result: "MATCH", Facts: details}
	return detectorCandidate{
		detection: &detection{class: ClassCompilerSerialized, variant: variant, detectorID: "llvm-bitcode-v1"},
		observed:  &observation,
	}
}

type llvmBitCursor struct {
	item blob
	bit  int64
	end  int64
}

func validateLLVMBitstream(item blob, offset, size int64) error {
	if offset < 0 || size < 16 || offset%4 != 0 || size%4 != 0 || offset > item.size || size > item.size-offset {
		return fmt.Errorf("LLVM bitstream range is invalid")
	}
	magic, err := readExactAt(item, offset, 4)
	if err != nil || !bytes.Equal(magic, []byte{'B', 'C', 0xc0, 0xde}) {
		return fmt.Errorf("LLVM bitstream magic is invalid")
	}
	cursor := llvmBitCursor{item: item, bit: (offset + 4) * 8, end: (offset + size) * 8}
	blocks := 0
	for cursor.bit < cursor.end {
		code, readErr := cursor.readBits(2)
		if readErr != nil {
			return readErr
		}
		if code != 1 {
			return fmt.Errorf("LLVM top-level bitstream record is invalid")
		}
		if _, readErr = cursor.readVBR(8); readErr != nil {
			return readErr
		}
		codeWidth, readErr := cursor.readVBR(4)
		if readErr != nil || codeWidth < 2 || codeWidth > 32 {
			return fmt.Errorf("LLVM subblock code width is invalid")
		}
		cursor.align32()
		wordCount, readErr := cursor.readBits(32)
		remainingWords := uint64((cursor.end - cursor.bit) / 32) // #nosec G115 -- readBits guarantees cursor.bit does not exceed cursor.end.
		if readErr != nil || wordCount == 0 || wordCount > remainingWords {
			return fmt.Errorf("LLVM subblock range is invalid")
		}
		cursor.bit += int64(wordCount * 32) // #nosec G115 -- bounded by the remaining cursor range above.
		blocks++
	}
	if blocks == 0 || cursor.bit != cursor.end {
		return fmt.Errorf("LLVM bitstream has no complete top-level block")
	}
	return nil
}

func (cursor *llvmBitCursor) readBits(count uint) (uint64, error) {
	if count == 0 || count > 64 || int64(count) > cursor.end-cursor.bit {
		return 0, io.ErrUnexpectedEOF
	}
	var value uint64
	for index := uint(0); index < count; index++ {
		byteOffset := cursor.bit / 8
		payload, err := readExactAt(cursor.item, byteOffset, 1)
		if err != nil {
			return 0, err
		}
		value |= uint64((payload[0]>>byte(cursor.bit&7))&1) << index
		cursor.bit++
	}
	return value, nil
}

func (cursor *llvmBitCursor) readVBR(width uint) (uint64, error) {
	if width < 2 || width > 32 {
		return 0, fmt.Errorf("LLVM VBR width is invalid")
	}
	var value uint64
	shift := uint(0)
	for groups := 0; groups < 32; groups++ {
		chunk, err := cursor.readBits(width)
		if err != nil {
			return 0, err
		}
		payload := chunk & ((uint64(1) << (width - 1)) - 1)
		if shift >= 64 || (payload != 0 && payload > ^uint64(0)>>shift) {
			return 0, fmt.Errorf("LLVM VBR value overflows")
		}
		value |= payload << shift
		if chunk&(uint64(1)<<(width-1)) == 0 {
			return value, nil
		}
		shift += width - 1
	}
	return 0, fmt.Errorf("LLVM VBR value is unterminated")
}

func (cursor *llvmBitCursor) align32() {
	cursor.bit = (cursor.bit + 31) &^ 31
}

func detectCompilerSerialized(item blob, virtualPath string) detectorCandidate {
	lower := strings.ToLower(leafPath(virtualPath))
	claimed := strings.HasSuffix(lower, ".swiftmodule") || strings.HasSuffix(lower, ".swiftdoc") ||
		strings.HasSuffix(lower, ".pcm") || strings.HasSuffix(lower, ".pch") ||
		strings.HasSuffix(lower, ".rmeta") || strings.HasSuffix(lower, ".gch") ||
		strings.HasSuffix(lower, ".ifc")
	if !claimed {
		return detectorCandidate{}
	}
	prefix, err := item.prefix(4 << 10)
	if err != nil {
		return detectorCandidate{}
	}
	if valid, _ := validateTextStream(bytes.NewReader(prefix), GrammarPlain); valid && int64(len(prefix)) == item.size {
		return detectorCandidate{}
	}
	observation := Observation{
		DetectorID: "compiler-serialized-v1", Result: "MATCH",
		Facts: facts(map[string]any{"claimed_by_path": true, "suffix": pathExtension(lower)}),
	}
	return detectorCandidate{
		detection: &detection{
			class: ClassCompilerSerialized, variant: "compiler.serialized_by_role",
			detectorID: "compiler-serialized-v1",
		},
		observed: &observation,
	}
}

func pathExtension(value string) string {
	if index := strings.LastIndexByte(value, '.'); index >= 0 {
		return value[index:]
	}
	return ""
}

func detectPythonBytecode(item blob, virtualPath string) detectorCandidate {
	prefix, err := item.prefix(16)
	claimed := strings.HasSuffix(strings.ToLower(leafPath(virtualPath)), ".pyc")
	if err != nil || len(prefix) < 4 {
		if claimed {
			return detectorCandidate{
				detection: &detection{class: ClassPythonBytecode, variant: "python.pyc.claimed", detectorID: "python-bytecode-v1"},
				observed:  &Observation{DetectorID: "python-bytecode-v1", Result: "MATCH", Facts: []Fact{{Key: "claimed_by_path", Value: "true"}}},
			}
		}
		return detectorCandidate{}
	}
	recognized := prefix[2] == '\r' && prefix[3] == '\n'
	if !recognized && !claimed {
		return detectorCandidate{}
	}
	variant := "python.pyc.claimed"
	factMap := map[string]any{"claimed_by_path": claimed, "magic": fmt.Sprintf("%x", prefix[:4])}
	if recognized {
		variant = "python.pyc"
		if len(prefix) < 16 {
			return detectorCandidate{observed: &Observation{
				DetectorID: "python-bytecode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "truncated pyc header"}},
			}}
		}
		flags := binary.LittleEndian.Uint32(prefix[4:8])
		if flags&^uint32(3) != 0 {
			return detectorCandidate{observed: &Observation{
				DetectorID: "python-bytecode-v1", Result: "ERROR",
				Facts: []Fact{{Key: "error", Value: "unsupported pyc flags"}},
			}}
		}
		factMap["flags"] = flags
	}
	observation := Observation{DetectorID: "python-bytecode-v1", Result: "MATCH", Facts: facts(factMap)}
	return detectorCandidate{
		detection: &detection{class: ClassPythonBytecode, variant: variant, detectorID: "python-bytecode-v1"},
		observed:  &observation,
	}
}

func detectJavaScriptCache(item blob, virtualPath string) detectorCandidate {
	lower := strings.ToLower(leafPath(virtualPath))
	claimed := strings.HasSuffix(lower, ".jsc") || strings.HasSuffix(lower, ".v8cache") ||
		strings.HasSuffix(lower, ".snapshot") || strings.HasSuffix(lower, "snapshot_blob.bin")
	prefix, err := item.prefix(16)
	if err != nil {
		return detectorCandidate{}
	}
	magic := len(prefix) >= 8 && (bytes.Equal(prefix[:4], []byte("V8CS")) || bytes.Equal(prefix[:4], []byte("NODE")))
	if !claimed && !magic {
		return detectorCandidate{}
	}
	observation := Observation{DetectorID: "v8-cache-v1", Result: "MATCH", Facts: facts(map[string]any{
		"claimed_by_path": claimed, "recognized_magic": magic,
	})}
	return detectorCandidate{
		detection: &detection{class: ClassJavaScriptCodeCache, variant: "v8.serialized_code", detectorID: "v8-cache-v1"},
		observed:  &observation,
	}
}

func checkedUint64Mul(left, right uint64) (uint64, bool) {
	if left != 0 && right > math.MaxUint64/left {
		return 0, false
	}
	return left * right, true
}
