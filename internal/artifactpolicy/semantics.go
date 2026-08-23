package artifactpolicy

import (
	"encoding/base64"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type semanticClassification struct {
	class      ArtifactClass
	variant    string
	detectorID string
}

// deriveNodeClassification independently resolves the normalized detector
// evidence carried by a manifest. DecodeManifest uses this instead of trusting
// the selected class, variant, or detector fields written by the producer.
func deriveNodeClassification(node ManifestNode) (semanticClassification, error) {
	if err := validateObservationSet(node); err != nil {
		return semanticClassification{}, err
	}
	if observation, ok := findObservation(node.Observations, "apple-bundle-path-v1", "MATCH"); ok {
		class, bundle := appleBundleClass(node.Path)
		if !bundle || !factsExactly(observation.Facts, map[string]string{"bundle_path": node.Path}) {
			return semanticClassification{}, fmt.Errorf("bundle classification lacks exact path evidence")
		}
		return semanticClassification{class: class, detectorID: "apple-bundle-path-v1"}, nil
	}

	switch node.Kind {
	case NodeDirectory:
		if _, bundle := appleBundleClass(node.Path); bundle {
			return semanticClassification{}, fmt.Errorf("bundle path lacks bundle detector evidence")
		}
		return semanticClassification{class: ClassDirectory}, nil
	case NodeLink:
		return semanticClassification{class: ClassLink}, nil
	case NodeSpecial:
		return semanticClassification{class: ClassSpecial}, nil
	}

	if metadata, ok, err := deriveARMetadataClassification(node); err != nil {
		return semanticClassification{}, err
	} else if ok {
		return metadata, nil
	}
	if metadata, ok, err := deriveTarMetadataClassification(node); err != nil {
		return semanticClassification{}, err
	} else if ok {
		return metadata, nil
	}

	candidates := make([]semanticClassification, 0, 2)
	errorDetectors := make([]string, 0, 1)
	var sourceFallback *semanticClassification
	for _, observation := range node.Observations {
		switch observation.Result {
		case "ERROR":
			errorDetectors = append(errorDetectors, observation.DetectorID)
		case "NO_MATCH", "NO_PROFILE_MATCH":
			if observation.DetectorID != "source-text-v1" {
				return semanticClassification{}, fmt.Errorf("detector %q cannot emit %q", observation.DetectorID, observation.Result)
			}
			if err := validateSourceFallbackObservation(node, observation); err != nil {
				return semanticClassification{}, err
			}
			fallback := semanticClassification{class: ClassOpaqueUnknown, detectorID: "source-text-v1"}
			sourceFallback = &fallback
		case "MATCH":
			candidate, err := classificationFromMatch(node, observation)
			if err != nil {
				return semanticClassification{}, err
			}
			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		classification := semanticClassification{class: ClassOpaqueUnknown}
		if sourceFallback != nil {
			classification = *sourceFallback
		}
		if len(errorDetectors) > 0 {
			sort.Strings(errorDetectors)
			classification.detectorID = errorDetectors[0]
		}
		return classification, nil
	}

	sort.SliceStable(candidates, func(left, right int) bool {
		leftPriority := semanticClassPriority(candidates[left])
		rightPriority := semanticClassPriority(candidates[right])
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return detectorSemanticPriority(candidates[left].detectorID) < detectorSemanticPriority(candidates[right].detectorID)
	})
	selected := candidates[0]
	selected.class = specializeNativeExtension(selected.class, node.Path)
	if len(errorDetectors) > 0 {
		sort.Strings(errorDetectors)
		selected.detectorID = errorDetectors[0]
	}
	return selected, nil
}

func validateObservationSet(node ManifestNode) error {
	seen := make(map[string]struct{}, len(node.Observations))
	for _, observation := range node.Observations {
		identity := observation.DetectorID + "\x00" + observation.Result
		if _, duplicate := seen[identity]; duplicate {
			return fmt.Errorf("duplicate detector result %q/%q", observation.DetectorID, observation.Result)
		}
		seen[identity] = struct{}{}
		if observation.Result == "ENTRY" {
			switch observation.DetectorID {
			case "archive-ar-v1", "archive-tar-v1", "archive-zip-v1", "compression-gzip-v1":
			default:
				return fmt.Errorf("detector %q cannot emit entry evidence", observation.DetectorID)
			}
			if err := validateEntryObservation(node, observation); err != nil {
				return fmt.Errorf("%s entry evidence: %w", observation.DetectorID, err)
			}
		}
		if observation.Result == "ERROR" && !hasNonemptyFact(observation.Facts) {
			return fmt.Errorf("detector %q error lacks evidence", observation.DetectorID)
		}
	}
	return nil
}

func validateEntryObservation(node ManifestNode, observation Observation) error {
	base := map[string]struct{}{}
	switch observation.DetectorID {
	case "archive-zip-v1":
		base = closedSet("compressed_size", "crc32", "flags", "method", "uncompressed_size")
		if _, ok := requiredDecimalFact(observation.Facts, "compressed_size", 0, ^uint64(0)>>1); !ok {
			return fmt.Errorf("invalid compressed size")
		}
		if size, ok := singleFactValue(observation.Facts, "uncompressed_size"); !ok ||
			size != strconv.FormatInt(node.Size, 10) || node.Size < 0 {
			return fmt.Errorf("uncompressed size does not match the node")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "crc32", 0, 1<<32-1); !ok {
			return fmt.Errorf("invalid CRC-32")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "flags", 0, 1<<16-1); !ok {
			return fmt.Errorf("invalid ZIP flags")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "method", 0, 1<<16-1); !ok {
			return fmt.Errorf("invalid ZIP method")
		}
	case "archive-tar-v1":
		base = closedSet(
			"atime_present", "ctime_present", "format", "gid_present", "gname_present",
			"gnu_long_link_present", "gnu_long_name_present", "metadata_kind", "metadata_members",
			"mode", "mtime_present", "pax_charset_present", "pax_comment_present",
			"pax_linkpath_present", "pax_path_present", "physical_header_index",
			"raw_name_base64", "size", "typeflag", "uid_present", "uname_present", "xattr_present",
		)
		format, formatOK := singleFactValue(observation.Facts, "format")
		if !formatOK || format == "" {
			return fmt.Errorf("tar format is absent")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "mode", 0, ^uint64(0)>>1); !ok {
			return fmt.Errorf("invalid tar mode")
		}
		if size, ok := singleFactValue(observation.Facts, "size"); !ok ||
			size != strconv.FormatInt(node.Size, 10) || node.Size < 0 {
			return fmt.Errorf("tar size does not match the node")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "typeflag", 0, 255); !ok {
			return fmt.Errorf("invalid tar type flag")
		}
		metadataKind, metadataKindOK := singleFactValue(observation.Facts, "metadata_kind")
		metadataMembers, metadataMembersOK := singleFactValue(observation.Facts, "metadata_members")
		_, physicalIndexOK := requiredDecimalFact(observation.Facts, "physical_header_index", 1, ^uint64(0)>>1)
		if (metadataKindOK || metadataMembers != "") != physicalIndexOK {
			return fmt.Errorf("physical tar header index does not match metadata binding")
		}
		rawName, rawNameOK := singleFactValue(observation.Facts, "raw_name_base64")
		if !rawNameOK || rawName != node.OriginalNameBase64 {
			return fmt.Errorf("raw tar name does not bind original-name evidence")
		}
		if _, err := base64.StdEncoding.DecodeString(rawName); err != nil {
			return fmt.Errorf("raw tar name is not valid base64")
		}
		if !metadataMembersOK {
			return fmt.Errorf("tar metadata-member binding is absent")
		}
		for _, key := range []string{
			"atime_present", "ctime_present", "gid_present", "gname_present",
			"gnu_long_link_present", "gnu_long_name_present", "mtime_present",
			"pax_charset_present", "pax_comment_present", "pax_linkpath_present",
			"pax_path_present", "uid_present", "uname_present", "xattr_present",
		} {
			if _, ok := booleanFact(observation.Facts, key); !ok {
				return fmt.Errorf("tar metadata presence %q is invalid", key)
			}
		}
		if metadataKindOK && !knownTarMetadataKind(metadataKind) {
			return fmt.Errorf("unknown tar metadata kind")
		}
	case "compression-gzip-v1":
		base = closedSet("comment", "extra_size", "header_name", "modtime_recorded", "os")
		if _, ok := requiredDecimalFact(observation.Facts, "extra_size", 0, ^uint64(0)>>1); !ok {
			return fmt.Errorf("invalid gzip extra size")
		}
		if _, ok := booleanFact(observation.Facts, "modtime_recorded"); !ok {
			return fmt.Errorf("invalid gzip modtime flag")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "os", 0, 255); !ok {
			return fmt.Errorf("invalid gzip OS identifier")
		}
		if _, ok := singleFactValue(observation.Facts, "comment"); !ok {
			return fmt.Errorf("gzip comment evidence is absent")
		}
		if _, ok := singleFactValue(observation.Facts, "header_name"); !ok {
			return fmt.Errorf("gzip header-name evidence is absent")
		}
	case "archive-ar-v1":
		base = closedSet(
			"declared_size", "member_size", "raw_name", "metadata_kind",
			"extended_name_size", "extended_name_sha256",
		)
		declared, declaredOK := requiredInt64Fact(observation.Facts, "declared_size", 0, math.MaxInt64)
		member, memberOK := requiredInt64Fact(observation.Facts, "member_size", 0, math.MaxInt64)
		memberText, memberTextOK := singleFactValue(observation.Facts, "member_size")
		if !declaredOK || !memberOK || !memberTextOK || node.Size < 0 ||
			memberText != strconv.FormatInt(node.Size, 10) {
			return fmt.Errorf("native-archive sizes do not bind the node")
		}
		rawName, rawNameOK := singleFactValue(observation.Facts, "raw_name")
		if !rawNameOK || rawName == "" {
			return fmt.Errorf("native-archive raw name is absent")
		}
		if strings.HasPrefix(rawName, "#1/") {
			nameLength, err := strconv.ParseInt(strings.TrimPrefix(rawName, "#1/"), 10, 64)
			recordedLength, lengthOK := requiredInt64Fact(
				observation.Facts, "extended_name_size", 1, math.MaxInt64,
			)
			recordedDigest, digestOK := singleFactValue(observation.Facts, "extended_name_sha256")
			originalName, decodeErr := base64.StdEncoding.DecodeString(node.OriginalNameBase64)
			expectedDeclared, addOK := checkedAdd(member, recordedLength)
			if err != nil || nameLength == 0 || !lengthOK || nameLength != recordedLength ||
				!digestOK || !sha256Identity.MatchString(recordedDigest) || decodeErr != nil ||
				int64(len(originalName)) != recordedLength || digestBytes(originalName) != recordedDigest ||
				!addOK || expectedDeclared != declared {
				return fmt.Errorf("BSD native-archive extended name evidence is not exact")
			}
			validated, pathErr := ValidateVirtualPath(string(originalName))
			containerPath := ""
			if len(node.ContainerChain) > 0 {
				containerPath = node.ContainerChain[len(node.ContainerChain)-1]
			}
			logicalPath := containerPath + "!/" + validated.Canonical
			occurrenceLogical, occurrenceOK := singleFactValue(observation.Facts, "logical_path")
			metadataKind, metadata := singleFactValue(observation.Facts, "metadata_kind")
			metadataPath := containerPath + "!/$ar-metadata/" + strings.ReplaceAll(metadataKind, "_", "-")
			logicalMatch := node.Path == logicalPath || occurrenceOK && occurrenceLogical == logicalPath
			metadataMatch := metadata && strings.HasPrefix(strings.ToUpper(string(originalName)), "__.SYMDEF") &&
				node.Path == metadataPath
			if containerPath == "" || pathErr != nil || (!logicalMatch && !metadataMatch) {
				return fmt.Errorf("BSD native-archive extended name does not bind the logical path")
			}
		} else {
			if declared != member {
				return fmt.Errorf("native-archive declared size does not equal member size")
			}
			if _, ok := singleFactValue(observation.Facts, "extended_name_size"); ok {
				return fmt.Errorf("non-BSD native-archive entry carries extended-name size")
			}
			if _, ok := singleFactValue(observation.Facts, "extended_name_sha256"); ok {
				return fmt.Errorf("non-BSD native-archive entry carries extended-name digest")
			}
		}
		if metadata, ok := singleFactValue(observation.Facts, "metadata_kind"); ok && !knownARMetadataKind(metadata) {
			return fmt.Errorf("unknown native-archive metadata kind")
		}
	}
	return validateEntryFactKeys(observation.Facts, base)
}

func validateEntryFactKeys(input []Fact, base map[string]struct{}) error {
	occurrence := closedSet("logical_path", "occurrence_count", "occurrence_identity", "occurrence_number")
	occurrenceCount := 0
	for _, fact := range input {
		if _, ok := base[fact.Key]; ok {
			continue
		}
		if _, ok := occurrence[fact.Key]; !ok {
			return fmt.Errorf("unsupported entry fact %q", fact.Key)
		}
		occurrenceCount++
	}
	if occurrenceCount == 0 {
		return nil
	}
	if occurrenceCount != len(occurrence) {
		return fmt.Errorf("partial duplicate-occurrence evidence")
	}
	logical, logicalOK := singleFactValue(input, "logical_path")
	count, countOK := requiredDecimalFact(input, "occurrence_count", 2, ^uint64(0)>>1)
	number, numberOK := requiredDecimalFact(input, "occurrence_number", 1, ^uint64(0)>>1)
	identity, identityOK := singleFactValue(input, "occurrence_identity")
	if !logicalOK || logical == "" || !countOK || !numberOK || number > count ||
		!identityOK || !sha256Identity.MatchString(identity) {
		return fmt.Errorf("invalid duplicate-occurrence evidence")
	}
	return nil
}

func deriveARMetadataClassification(node ManifestNode) (semanticClassification, bool, error) {
	if node.Kind != NodeRegularFile || node.SelectedDetectorID != "archive-ar-v1" {
		return semanticClassification{}, false, nil
	}
	for _, observation := range node.Observations {
		if observation.DetectorID != "archive-ar-v1" || observation.Result != "ENTRY" {
			continue
		}
		kind, ok := singleFactValue(observation.Facts, "metadata_kind")
		if !ok {
			continue
		}
		if !knownARMetadataKind(kind) {
			return semanticClassification{}, false, fmt.Errorf("unsupported native-archive metadata kind %q", kind)
		}
		return semanticClassification{
			class: ClassNativeLibraryStatic, variant: "ar.metadata." + kind, detectorID: "archive-ar-v1",
		}, true, nil
	}
	return semanticClassification{}, false, nil
}

func manifestedNodeEmittedSize(node ManifestNode) (int64, error) {
	for _, observation := range node.Observations {
		if observation.DetectorID != "archive-ar-v1" || observation.Result != "ENTRY" {
			continue
		}
		rawName, rawNameOK := singleFactValue(observation.Facts, "raw_name")
		if !rawNameOK {
			return 0, fmt.Errorf("native-archive raw name is absent for %q", node.Path)
		}
		if !strings.HasPrefix(rawName, "#1/") {
			return node.Size, nil
		}
		declared, ok := requiredInt64Fact(observation.Facts, "declared_size", 1, math.MaxInt64)
		if !ok {
			return 0, fmt.Errorf("BSD native-archive declared size is invalid for %q", node.Path)
		}
		return declared, nil
	}
	return node.Size, nil
}

func deriveTarMetadataClassification(node ManifestNode) (semanticClassification, bool, error) {
	if node.Kind != NodeRegularFile || node.SelectedDetectorID != "archive-tar-v1" {
		return semanticClassification{}, false, nil
	}
	for _, observation := range node.Observations {
		if observation.DetectorID != "archive-tar-v1" || observation.Result != "ENTRY" {
			continue
		}
		kind, ok := singleFactValue(observation.Facts, "metadata_kind")
		if !ok {
			continue
		}
		if !knownTarMetadataKind(kind) {
			return semanticClassification{}, false, fmt.Errorf("unsupported tar metadata kind %q", kind)
		}
		return semanticClassification{
			class: ClassTextMetadata, variant: "tar.metadata." + kind, detectorID: "archive-tar-v1",
		}, true, nil
	}
	return semanticClassification{}, false, nil
}

func knownTarMetadataKind(kind string) bool {
	switch kind {
	case "pax-local", "pax-global", "gnu-long-name", "gnu-long-link":
		return true
	default:
		return false
	}
}

func knownARMetadataKind(kind string) bool {
	for _, prefix := range []string{"symbol_table_", "symbol_table_64_", "string_table_", "bsd_symbol_table_"} {
		if !strings.HasPrefix(kind, prefix) {
			continue
		}
		index := strings.TrimPrefix(kind, prefix)
		if len(index) == 3 && index >= "001" && index <= "999" {
			return true
		}
	}
	return false
}

func classificationFromMatch(node ManifestNode, observation Observation) (semanticClassification, error) {
	switch observation.DetectorID {
	case "apple-bundle-path-v1":
		return semanticClassification{}, fmt.Errorf("bundle match appears on non-directory node")
	case "archive-zip-v1":
		if node.Kind != NodeArchive || !onlyOptionalFact(observation.Facts, "deny_indicating_class") {
			return semanticClassification{}, fmt.Errorf("ZIP match has incompatible node kind or facts")
		}
		return semanticClassification{class: ClassArchive, detectorID: observation.DetectorID}, nil
	case "archive-tar-v1":
		if node.Kind != NodeArchive || !onlyOptionalFact(observation.Facts, "deny_indicating_class") {
			return semanticClassification{}, fmt.Errorf("tar match has incompatible node kind or facts")
		}
		return semanticClassification{class: ClassArchive, detectorID: observation.DetectorID}, nil
	case "compression-gzip-v1":
		if node.Kind != NodeCompressedStream || !onlyOptionalFact(observation.Facts, "deny_indicating_class") {
			return semanticClassification{}, fmt.Errorf("gzip match has incompatible node kind or facts")
		}
		return semanticClassification{class: ClassCompressedStream, detectorID: observation.DetectorID}, nil
	case "archive-ar-v1":
		if node.Kind != NodeArchive || len(observation.Facts) != 0 {
			return semanticClassification{}, fmt.Errorf("native archive match has incompatible node kind or facts")
		}
		return semanticClassification{class: ClassNativeLibraryStatic, detectorID: observation.DetectorID}, nil
	case "unsupported-container-v1":
		format, ok := singleFactValue(observation.Facts, "format")
		if node.Kind != NodeRegularFile || !ok || len(observation.Facts) != 1 || !knownUnsupportedContainer(format) {
			return semanticClassification{}, fmt.Errorf("unsupported-container evidence is invalid")
		}
		return semanticClassification{class: ClassOpaqueUnknown, detectorID: observation.DetectorID}, nil
	case "elf-v1":
		return classificationFromELFFacts(node, observation)
	case "pe-coff-v1":
		return classificationFromPECOFFFacts(node, observation)
	case "macho-v1":
		return classificationFromMachOFacts(node, observation)
	case "jvm-class-v1":
		if _, ok := requiredDecimalFact(observation.Facts, "access_flags", 0, 1<<16-1); !ok {
			return semanticClassification{}, fmt.Errorf("JVM class evidence lacks valid access flags")
		}
		for _, key := range []string{"attributes", "fields", "interfaces", "methods"} {
			if _, ok := requiredDecimalFact(observation.Facts, key, 0, 1<<16-1); !ok {
				return semanticClassification{}, fmt.Errorf("JVM class evidence lacks valid %s", key)
			}
		}
		if _, ok := requiredDecimalFact(observation.Facts, "major", 45, 80); !ok {
			return semanticClassification{}, fmt.Errorf("JVM class evidence lacks a valid major version")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "minor", 0, 1<<16-1); !ok {
			return semanticClassification{}, fmt.Errorf("JVM class evidence lacks a valid minor version")
		}
		constantCount, constantOK := requiredDecimalFact(observation.Facts, "constant_pool_count", 1, 1<<16-1)
		thisClass, thisOK := requiredDecimalFact(observation.Facts, "this_class", 1, constantCount-1)
		superClass, superOK := requiredDecimalFact(observation.Facts, "super_class", 0, constantCount-1)
		thisName, nameOK := singleFactValue(observation.Facts, "this_class_name")
		if !constantOK || !thisOK || !superOK || thisClass == superClass || !nameOK || thisName == "" ||
			!exactFactKeys(observation.Facts, closedSet(
				"access_flags", "attributes", "constant_pool_count", "fields", "interfaces",
				"major", "methods", "minor", "super_class", "this_class", "this_class_name",
			), nil) {
			return semanticClassification{}, fmt.Errorf("JVM class evidence has an invalid field set")
		}
		return semanticClassification{class: ClassJVMBytecode, variant: "jvm.class", detectorID: observation.DetectorID}, nil
	case "dex-v1":
		version, ok := singleFactValue(observation.Facts, "version")
		if !ok || len(version) != 3 || !exactFactKeys(observation.Facts, closedSet("version"), nil) {
			return semanticClassification{}, fmt.Errorf("DEX evidence lacks a valid version")
		}
		if _, err := strconv.ParseUint(version, 10, 16); err != nil {
			return semanticClassification{}, fmt.Errorf("DEX evidence has an invalid version")
		}
		return semanticClassification{class: ClassJVMBytecode, variant: "android.dex", detectorID: observation.DetectorID}, nil
	case "wasm-v1":
		version, ok := singleFactValue(observation.Facts, "version")
		if !ok || version != "1" {
			return semanticClassification{}, fmt.Errorf("WebAssembly evidence lacks version 1")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "section_count", 0, ^uint64(0)>>1); !ok ||
			!exactFactKeys(observation.Facts, closedSet("section_count", "version"), nil) {
			return semanticClassification{}, fmt.Errorf("WebAssembly section evidence is invalid")
		}
		return semanticClassification{class: ClassWebAssembly, variant: "wasm.core", detectorID: observation.DetectorID}, nil
	case "llvm-bitcode-v1":
		variant := "llvm.bitcode.raw"
		if len(observation.Facts) != 0 {
			offset, offsetOK := requiredDecimalFact(observation.Facts, "offset", 20, ^uint64(0)>>1)
			size, sizeOK := requiredDecimalFact(observation.Facts, "size", 16, ^uint64(0)>>1)
			if !offsetOK || !sizeOK || len(observation.Facts) != 2 || offset%4 != 0 || size%4 != 0 {
				return semanticClassification{}, fmt.Errorf("LLVM wrapper evidence is invalid")
			}
			variant = "llvm.bitcode.wrapper"
		}
		return semanticClassification{class: ClassCompilerSerialized, variant: variant, detectorID: observation.DetectorID}, nil
	case "compiler-serialized-v1":
		claimed, claimedOK := singleFactValue(observation.Facts, "claimed_by_path")
		suffix, suffixOK := singleFactValue(observation.Facts, "suffix")
		if !claimedOK || claimed != "true" || !suffixOK || len(observation.Facts) != 2 ||
			!serializedSuffix(suffix) || !exactFactKeys(observation.Facts, closedSet("claimed_by_path", "suffix"), nil) {
			return semanticClassification{}, fmt.Errorf("serialized-compiler evidence is invalid")
		}
		return semanticClassification{class: ClassCompilerSerialized, variant: "compiler.serialized_by_role", detectorID: observation.DetectorID}, nil
	case "python-bytecode-v1":
		claimed, ok := singleFactValue(observation.Facts, "claimed_by_path")
		if !ok || (claimed != "true" && claimed != "false") {
			return semanticClassification{}, fmt.Errorf("python bytecode evidence lacks a valid path claim")
		}
		variant := "python.pyc.claimed"
		allowed := closedSet("claimed_by_path")
		if magic, magicPresent := singleFactValue(observation.Facts, "magic"); magicPresent {
			if len(magic) != 8 {
				return semanticClassification{}, fmt.Errorf("python bytecode evidence has invalid magic")
			}
			allowed["magic"] = struct{}{}
		}
		if _, flags := singleFactValue(observation.Facts, "flags"); flags {
			if _, ok := requiredDecimalFact(observation.Facts, "flags", 0, 3); !ok {
				return semanticClassification{}, fmt.Errorf("python bytecode evidence has invalid flags")
			}
			if magic, magicOK := singleFactValue(observation.Facts, "magic"); !magicOK || len(magic) != 8 {
				return semanticClassification{}, fmt.Errorf("python bytecode evidence lacks its magic")
			}
			variant = "python.pyc"
			allowed["flags"] = struct{}{}
		}
		if !exactFactKeys(observation.Facts, allowed, nil) {
			return semanticClassification{}, fmt.Errorf("python bytecode evidence has an invalid field set")
		}
		return semanticClassification{class: ClassPythonBytecode, variant: variant, detectorID: observation.DetectorID}, nil
	case "v8-cache-v1":
		claimed, claimedOK := booleanFact(observation.Facts, "claimed_by_path")
		magic, magicOK := booleanFact(observation.Facts, "recognized_magic")
		if !claimedOK || !magicOK || len(observation.Facts) != 2 || (!claimed && !magic) ||
			!exactFactKeys(observation.Facts, closedSet("claimed_by_path", "recognized_magic"), nil) {
			return semanticClassification{}, fmt.Errorf("JavaScript cache evidence is invalid")
		}
		return semanticClassification{class: ClassJavaScriptCodeCache, variant: "v8.serialized_code", detectorID: observation.DetectorID}, nil
	case "source-text-v1":
		return classificationFromSourceFacts(node, observation)
	default:
		return semanticClassification{}, fmt.Errorf("detector %q has no closed match semantics", observation.DetectorID)
	}
}

func classificationFromELFFacts(node ManifestNode, observation Observation) (semanticClassification, error) {
	if err := validateELFFactSchema(node, observation.Facts); err != nil {
		return semanticClassification{}, err
	}
	eType, ok := requiredDecimalFact(observation.Facts, "e_type", elfTypeRel, elfTypeDyn)
	if !ok {
		return semanticClassification{}, fmt.Errorf("ELF evidence lacks a supported e_type")
	}
	if err := validateResolvedUseFacts(node.DeclaredUses, observation.Facts); err != nil {
		return semanticClassification{}, fmt.Errorf("ELF use evidence: %w", err)
	}
	classification := semanticClassification{detectorID: observation.DetectorID}
	switch eType {
	case elfTypeRel:
		classification.class, classification.variant = ClassNativeObject, "elf.relocatable"
	case elfTypeExec:
		classification.class, classification.variant = ClassNativeExecutable, "elf.executable"
	case elfTypeDyn:
		pie, pieOK := booleanFact(observation.Facts, "df_1_pie")
		soname, sonameOK := booleanFact(observation.Facts, "dt_soname_present")
		interp, interpOK := requiredDecimalFact(observation.Facts, "pt_interp_count", 0, 1)
		if !pieOK || !sonameOK || !interpOK {
			return semanticClassification{}, fmt.Errorf("ELF ET_DYN evidence lacks its discriminators")
		}
		execute, link := countUses(node.DeclaredUses)
		switch {
		case pie:
			classification.class = ClassNativeExecutable
			classification.variant = "elf.pie.no_interpreter"
			if interp == 1 {
				classification.variant = "elf.pie.interpreter"
			}
		case execute > 0 && link > 0:
			classification.class, classification.variant = ClassELFETDYNAmbiguous, "use_conflict"
		case execute > 1 || link > 1:
			classification.class, classification.variant = ClassELFETDYNAmbiguous, "duplicate_use_edges"
		case execute == 1 && link == 0 && interp == 1 && !soname:
			classification.class, classification.variant = ClassNativeExecutable, "elf.et_dyn.executable_by_use"
		case execute == 0 && interp == 0 && (soname || link == 1):
			classification.class, classification.variant = ClassNativeLibraryDynamic, "elf.shared_object"
		default:
			classification.class, classification.variant = ClassELFETDYNAmbiguous, "insufficient_evidence"
			if interp > 0 && soname {
				classification.variant = "interp_soname_conflict"
			}
		}
	}
	return classification, nil
}

func classificationFromPECOFFFacts(node ManifestNode, observation Observation) (semanticClassification, error) {
	if machine, ok := requiredDecimalFact(observation.Facts, "machine", 0, 0xffff); !ok || !knownCOFFMachine(machine) {
		return semanticClassification{}, fmt.Errorf("PE/COFF evidence lacks a supported machine")
	}
	if _, ok := requiredDecimalFact(observation.Facts, "sections", 1, 96); !ok {
		return semanticClassification{}, fmt.Errorf("PE/COFF evidence lacks a valid section count")
	}
	if _, image := singleFactValue(observation.Facts, "optional_magic"); image {
		magic, magicOK := requiredDecimalFact(observation.Facts, "optional_magic", 0x10b, 0x20b)
		characteristics, characteristicsOK := requiredDecimalFact(observation.Facts, "characteristics", 0, 0xffff)
		if !magicOK || (magic != 0x10b && magic != 0x20b) || !characteristicsOK {
			return semanticClassification{}, fmt.Errorf("PE image evidence is invalid")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "data_directory_count", 0, 16); !ok {
			return semanticClassification{}, fmt.Errorf("PE image data-directory evidence is invalid")
		}
		if err := validatePECOFFSectionFacts(node.Size, observation.Facts, true); err != nil {
			return semanticClassification{}, err
		}
		if characteristics&0x2000 != 0 {
			return semanticClassification{class: ClassNativeLibraryDynamic, variant: "pe.dll", detectorID: observation.DetectorID}, nil
		}
		return semanticClassification{class: ClassNativeExecutable, variant: "pe.image", detectorID: observation.DetectorID}, nil
	}
	if err := validatePECOFFSectionFacts(node.Size, observation.Facts, false); err != nil {
		return semanticClassification{}, err
	}
	return semanticClassification{class: ClassNativeObject, variant: "coff.object", detectorID: observation.DetectorID}, nil
}

func classificationFromMachOFacts(node ManifestNode, observation Observation) (semanticClassification, error) {
	if fileType, thin := singleFactValue(observation.Facts, "file_type"); thin {
		value, err := strconv.ParseUint(fileType, 10, 32)
		if err != nil {
			return semanticClassification{}, fmt.Errorf("Mach-O file type is invalid")
		}
		if _, ok := booleanFact(observation.Facts, "is_64"); !ok {
			return semanticClassification{}, fmt.Errorf("Mach-O bitness evidence is invalid")
		}
		if _, ok := requiredDecimalFact(observation.Facts, "load_commands", 0, 1<<16-1); !ok {
			return semanticClassification{}, fmt.Errorf("Mach-O load-command count is invalid")
		}
		if _, ok := requiredInt64Fact(observation.Facts, "load_command_bytes", 0, node.Size); !ok ||
			!exactFactKeys(observation.Facts, closedSet("file_type", "is_64", "load_command_bytes", "load_commands"), nil) {
			return semanticClassification{}, fmt.Errorf("Mach-O thin evidence has an invalid field set")
		}
		classification := semanticClassification{detectorID: observation.DetectorID}
		switch value {
		case 1:
			classification.class, classification.variant = ClassNativeObject, "macho.object"
		case 2, 3, 4, 5, 7:
			classification.class, classification.variant = ClassNativeExecutable, "macho.executable"
		case 6:
			classification.class, classification.variant = ClassNativeLibraryDynamic, "macho.dylib"
		case 8:
			classification.class, classification.variant = ClassNativeLibraryDynamic, "macho.bundle"
		default:
			return semanticClassification{}, fmt.Errorf("unsupported Mach-O file type %d", value)
		}
		return classification, nil
	}
	count, ok := requiredDecimalFact(observation.Facts, "slice_count", 1, 64)
	if !ok {
		return semanticClassification{}, fmt.Errorf("Mach-O evidence lacks thin or universal discriminators")
	}
	if _, ok := booleanFact(observation.Facts, "fat_64"); !ok {
		return semanticClassification{}, fmt.Errorf("Mach-O universal bitness evidence is invalid")
	}
	variants, variantsOK := singleFactValue(observation.Facts, "slice_variants")
	if !variantsOK {
		return semanticClassification{}, fmt.Errorf("Mach-O universal variants are absent")
	}
	variantValues := strings.Split(variants, ",")
	if uint64(len(variantValues)) != count {
		return semanticClassification{}, fmt.Errorf("Mach-O universal variant count is inconsistent")
	}
	allowed := closedSet("fat_64", "slice_count", "slice_variants")
	selected := ArtifactClass("")
	for index := uint64(0); index < count; index++ {
		prefix := fmt.Sprintf("slice.%03d.", index)
		key := prefix + "class"
		classValue, present := singleFactValue(observation.Facts, key)
		class := ArtifactClass(classValue)
		if !present || (class != ClassNativeExecutable && class != ClassNativeObject && class != ClassNativeLibraryDynamic) {
			return semanticClassification{}, fmt.Errorf("Mach-O slice %d lacks a valid class", index)
		}
		if selected == "" || machoSliceClassPriority(class) < machoSliceClassPriority(selected) {
			selected = class
		}
		variant, variantOK := singleFactValue(observation.Facts, prefix+"variant")
		if !variantOK || variant != variantValues[index] || !machOVariantMatchesClass(variant, class) {
			return semanticClassification{}, fmt.Errorf("Mach-O slice %d variant is inconsistent", index)
		}
		offset, offsetOK := requiredInt64Fact(observation.Facts, prefix+"offset", 0, node.Size)
		size, sizeOK := requiredInt64Fact(observation.Facts, prefix+"size", 1, node.Size)
		if !offsetOK || !sizeOK || offset > node.Size-size {
			return semanticClassification{}, fmt.Errorf("Mach-O slice %d range is invalid", index)
		}
		if _, alignmentOK := requiredDecimalFact(observation.Facts, prefix+"alignment_power", 0, 63); !alignmentOK {
			return semanticClassification{}, fmt.Errorf("Mach-O slice %d alignment is invalid", index)
		}
		for _, suffix := range []string{"alignment_power", "class", "offset", "size", "variant"} {
			allowed[prefix+suffix] = struct{}{}
		}
	}
	if !exactFactKeys(observation.Facts, allowed, nil) {
		return semanticClassification{}, fmt.Errorf("Mach-O universal evidence has an invalid field set")
	}
	return semanticClassification{class: selected, variant: "macho.universal", detectorID: observation.DetectorID}, nil
}

func classificationFromSourceFacts(node ManifestNode, observation Observation) (semanticClassification, error) {
	grammarValue, grammarOK := singleFactValue(observation.Facts, "grammar")
	classValue, classOK := singleFactValue(observation.Facts, "selected_class")
	lineage, lineageOK := singleFactValue(observation.Facts, "generated_lineage")
	class := ArtifactClass(classValue)
	if !grammarOK || !classOK || !lineageOK || !textClass(class) {
		return semanticClassification{}, fmt.Errorf("source match lacks its closed grammar, class, or lineage evidence")
	}
	if _, ok := grammarIDs[GrammarID(grammarValue)]; !ok {
		return semanticClassification{}, fmt.Errorf("source match uses unsupported grammar %q", grammarValue)
	}
	if err := validateResolvedUseFacts(node.DeclaredUses, observation.Facts); err != nil {
		return semanticClassification{}, fmt.Errorf("source use evidence: %w", err)
	}
	if class != ClassSourceGeneratedText && lineage != "" {
		return semanticClassification{}, fmt.Errorf("non-generated text carries generator lineage")
	}
	if !exactFactKeys(
		observation.Facts,
		closedSet("generated_lineage", "grammar", "selected_class", "resolved_use"),
		closedSet("resolved_use"),
	) {
		return semanticClassification{}, fmt.Errorf("source match carries unsupported or duplicate facts")
	}
	if denySuffixClass(node.Path) != "" || hasResolvedUse(node.DeclaredUses, UseLinkOrLoad) {
		class = ClassOpaqueUnknown
	}
	return semanticClassification{class: class, detectorID: observation.DetectorID}, nil
}

func validateELFFactSchema(node ManifestNode, input []Fact) error {
	required := closedSet(
		"abi_version", "class", "data_encoding", "df_1_pie", "dt_flags_1",
		"dt_soname_present", "e_entry", "e_flags", "e_machine", "e_type",
		"os_abi", "program_header_types", "pt_dynamic", "pt_interp_count",
	)
	allowed := make(map[string]struct{}, len(required)+len(input))
	for key := range required {
		allowed[key] = struct{}{}
	}
	allowed["resolved_use"] = struct{}{}
	for _, field := range []struct {
		key          string
		minimum, max uint64
	}{
		{"abi_version", 0, 255}, {"class", 1, 2}, {"data_encoding", 1, 2},
		{"dt_flags_1", 0, ^uint64(0)}, {"e_entry", 0, ^uint64(0)},
		{"e_flags", 0, 1<<32 - 1}, {"e_machine", 0, 1<<16 - 1},
		{"e_type", elfTypeRel, elfTypeDyn}, {"os_abi", 0, 255},
		{"pt_interp_count", 0, 1},
	} {
		if _, ok := requiredDecimalFact(input, field.key, field.minimum, field.max); !ok {
			return fmt.Errorf("ELF fact %q is absent or invalid", field.key)
		}
	}
	pie, pieOK := booleanFact(input, "df_1_pie")
	soname, sonameOK := booleanFact(input, "dt_soname_present")
	ptDynamic, dynamicOK := booleanFact(input, "pt_dynamic")
	if !pieOK || !sonameOK || !dynamicOK {
		return fmt.Errorf("ELF boolean discriminators are invalid")
	}

	programTypesValue, typesOK := singleFactValue(input, "program_header_types")
	if !typesOK {
		return fmt.Errorf("ELF program-header type list is absent")
	}
	programTypes := []string{}
	if programTypesValue != "" {
		programTypes = strings.Split(programTypesValue, ",")
	}
	interpCount := uint64(0)
	dynamicCount := uint64(0)
	for index, expectedType := range programTypes {
		prefix := fmt.Sprintf("program_header.%05d.", index)
		programType, ok := requiredDecimalFact(input, prefix+"type", 0, 1<<32-1)
		if !ok || strconv.FormatUint(programType, 10) != expectedType {
			return fmt.Errorf("ELF program-header %d type is inconsistent", index)
		}
		fileOffset, offsetOK := requiredInt64Fact(input, prefix+"file_offset", 0, node.Size)
		fileSize, sizeOK := requiredInt64Fact(input, prefix+"file_size", 0, node.Size)
		if !offsetOK || !sizeOK || fileOffset > node.Size-fileSize {
			return fmt.Errorf("ELF program-header %d file range is invalid", index)
		}
		if _, memoryOK := requiredDecimalFact(input, prefix+"memory_size", 0, ^uint64(0)); !memoryOK {
			return fmt.Errorf("ELF program-header %d memory size is invalid", index)
		}
		for _, suffix := range []string{"file_offset", "file_size", "memory_size", "type"} {
			allowed[prefix+suffix] = struct{}{}
		}
		switch programType {
		case elfPTInterp:
			interpCount++
			interpreter, interpreterOK := singleFactValue(input, prefix+"interpreter")
			if !interpreterOK || interpreter == "" {
				return fmt.Errorf("ELF interpreter evidence is absent")
			}
			allowed[prefix+"interpreter"] = struct{}{}
		case elfPTDynamic:
			dynamicCount++
		}
	}
	declaredInterp, _ := requiredDecimalFact(input, "pt_interp_count", 0, 1)
	if interpCount != declaredInterp || (dynamicCount == 1) != ptDynamic || dynamicCount > 1 {
		return fmt.Errorf("ELF program-header summary is inconsistent")
	}

	dynamicGroups := make(map[uint64]map[string]uint64)
	for _, fact := range input {
		index, suffix, ok := indexedFactKey(fact.Key, "dynamic.", 5)
		if !ok {
			continue
		}
		if suffix != "tag" && suffix != "value" {
			return fmt.Errorf("unsupported ELF dynamic fact %q", fact.Key)
		}
		value, err := strconv.ParseUint(fact.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid ELF dynamic fact %q", fact.Key)
		}
		if dynamicGroups[index] == nil {
			dynamicGroups[index] = make(map[string]uint64)
		}
		if _, duplicate := dynamicGroups[index][suffix]; duplicate {
			return fmt.Errorf("duplicate ELF dynamic fact %q", fact.Key)
		}
		dynamicGroups[index][suffix] = value
		allowed[fact.Key] = struct{}{}
	}
	flags1 := uint64(0)
	flagsSeen := false
	sonameSeen := false
	for index := uint64(0); index < uint64(len(dynamicGroups)); index++ {
		group, ok := dynamicGroups[index]
		if !ok || len(group) != 2 {
			return fmt.Errorf("ELF dynamic-table indexes or pairs are incomplete")
		}
		switch group["tag"] {
		case elfDTFlags1:
			if flagsSeen {
				return fmt.Errorf("duplicate ELF DT_FLAGS_1 evidence")
			}
			flagsSeen, flags1 = true, group["value"]
		case elfDTSoname:
			if sonameSeen {
				return fmt.Errorf("duplicate ELF DT_SONAME evidence")
			}
			sonameSeen = true
		}
	}
	declaredFlags, _ := requiredDecimalFact(input, "dt_flags_1", 0, ^uint64(0))
	if declaredFlags != flags1 || soname != sonameSeen || pie != (flags1&elfDF1PIE != 0) {
		return fmt.Errorf("ELF dynamic-table summary is inconsistent")
	}
	if !exactFactKeys(input, allowed, closedSet("resolved_use")) {
		return fmt.Errorf("ELF evidence has unsupported or duplicate facts")
	}
	return nil
}

func validatePECOFFSectionFacts(nodeSize int64, input []Fact, image bool) error {
	sectionCount, _ := requiredDecimalFact(input, "sections", 1, 96)
	allowed := closedSet("machine", "sections")
	if image {
		for _, key := range []string{"characteristics", "data_directory_count", "optional_magic"} {
			allowed[key] = struct{}{}
		}
	}
	for index := uint64(0); index < sectionCount; index++ {
		prefix := fmt.Sprintf("section.%05d.", index)
		offset, offsetOK := requiredInt64Fact(input, prefix+"raw_offset", 0, nodeSize)
		size, sizeOK := requiredInt64Fact(input, prefix+"raw_size", 0, nodeSize)
		if !offsetOK || !sizeOK || (size > 0 && offset > nodeSize-size) {
			return fmt.Errorf("PE/COFF section %d range is invalid", index)
		}
		allowed[prefix+"raw_offset"] = struct{}{}
		allowed[prefix+"raw_size"] = struct{}{}
	}
	if !exactFactKeys(input, allowed, nil) {
		return fmt.Errorf("PE/COFF evidence has an invalid field set")
	}
	return nil
}

func machOVariantMatchesClass(variant string, class ArtifactClass) bool {
	switch class {
	case ClassNativeObject:
		return variant == "macho.object"
	case ClassNativeExecutable:
		return variant == "macho.executable"
	case ClassNativeLibraryDynamic:
		return variant == "macho.dylib" || variant == "macho.bundle"
	default:
		return false
	}
}

func indexedFactKey(key, prefix string, width int) (uint64, string, bool) {
	if !strings.HasPrefix(key, prefix) {
		return 0, "", false
	}
	remainder := strings.TrimPrefix(key, prefix)
	dot := strings.IndexByte(remainder, '.')
	if dot != width {
		return 0, "", false
	}
	index, err := strconv.ParseUint(remainder[:dot], 10, 64)
	if err != nil || fmt.Sprintf("%0*d", width, index) != remainder[:dot] {
		return 0, "", false
	}
	return index, remainder[dot+1:], true
}

func requiredInt64Fact(input []Fact, key string, minimum, maximum int64) (int64, bool) {
	value, ok := singleFactValue(input, key)
	if !ok || maximum < minimum {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < minimum || parsed > maximum || strconv.FormatInt(parsed, 10) != value {
		return 0, false
	}
	return parsed, true
}

func exactFactKeys(input []Fact, allowed, multiple map[string]struct{}) bool {
	seen := make(map[string]struct{}, len(input))
	for _, fact := range input {
		if _, ok := allowed[fact.Key]; !ok {
			return false
		}
		if _, repeatable := multiple[fact.Key]; repeatable {
			continue
		}
		if _, duplicate := seen[fact.Key]; duplicate {
			return false
		}
		seen[fact.Key] = struct{}{}
	}
	for key := range allowed {
		if _, repeatable := multiple[key]; repeatable {
			continue
		}
		if _, present := seen[key]; !present {
			return false
		}
	}
	return true
}

func validateSourceFallbackObservation(node ManifestNode, observation Observation) error {
	if _, ok := singleFactValue(observation.Facts, "reason"); !ok {
		return fmt.Errorf("source profile miss lacks its reason")
	}
	if err := validateResolvedUseFacts(node.DeclaredUses, observation.Facts); err != nil {
		return fmt.Errorf("source fallback use evidence: %w", err)
	}
	if len(observation.Facts) != 1+len(node.DeclaredUses) {
		return fmt.Errorf("source profile miss carries unsupported facts")
	}
	return nil
}

func validateResolvedUseFacts(uses []UseEdge, input []Fact) error {
	want := useFacts(uses)
	observed := make([]Fact, 0, len(want))
	for _, fact := range input {
		if fact.Key == "resolved_use" {
			observed = append(observed, fact)
		}
	}
	if !slices.Equal(observed, want) {
		return fmt.Errorf("resolved-use facts do not match declared uses")
	}
	return nil
}

func semanticClassPriority(classification semanticClassification) int {
	priority := compiledClassPriority(classification.class)
	if priority < 9 {
		return priority
	}
	switch classification.class {
	case ClassArchive, ClassCompressedStream:
		return 5
	case ClassSourceAuthoredText, ClassSourceGeneratedText, ClassTextMetadata:
		return 6
	case ClassOpaqueUnknown:
		return 7
	default:
		return 8
	}
}

func detectorSemanticPriority(detectorID string) int {
	order := []string{
		"elf-v1", "pe-coff-v1", "macho-v1", "jvm-class-v1", "dex-v1",
		"wasm-v1", "llvm-bitcode-v1", "compiler-serialized-v1",
		"python-bytecode-v1", "v8-cache-v1", "archive-ar-v1",
		"archive-tar-v1", "archive-zip-v1", "compression-gzip-v1",
		"source-text-v1", "unsupported-container-v1",
	}
	for index, candidate := range order {
		if candidate == detectorID {
			return index
		}
	}
	return len(order)
}

func validateResolvedNodeRule(manifest Manifest, node ManifestNode, expected semanticClassification) error {
	baseDecision := decisionForClass(manifest.TrustRole, expected.class)
	if expected.class == ClassArchive || expected.class == ClassCompressedStream || expected.class == ClassDirectory {
		baseDecision = DecisionDescend
	}
	if node.Decision != DecisionReject {
		if node.Decision != baseDecision {
			return fmt.Errorf("decision %q is invalid for role %q and class %q", node.Decision, manifest.TrustRole, expected.class)
		}
		if !validNonRejectRule(node) {
			return fmt.Errorf("rule %q is invalid for non-rejecting node", node.Rule)
		}
		return nil
	}

	valid := false
	switch node.Rule {
	case "artifact_finding_rejected":
		valid = node.Parent == "" && manifest.RawPayload.Kind == "file" && manifest.Findings.Total > 0
	case "tree_member_rejected":
		valid = node.Parent == "" && manifest.RawPayload.Kind == "canonical_tree" && manifest.Findings.Total > 0
	case "member_rejected":
		valid = node.Kind == NodeArchive || node.Kind == NodeCompressedStream
	case "descendant_rejected":
		valid = node.Parent != "" && containerCapableNode(node.Kind)
	case "detector_rejection":
		valid = node.SelectedDetectorID != "" &&
			(!node.InspectionComplete || expected.class == ClassOpaqueUnknown ||
				hasDenyIndicatingFact(node.Observations) || missingGeneratedLineage(node.Observations))
	case "class_role_decision":
		valid = baseDecision == DecisionReject
	case "compiled_native_archive":
		valid = manifest.TrustRole == RoleDependencyInput && node.Kind == NodeArchive && expected.class == ClassNativeLibraryStatic
	case "native_archive_metadata":
		valid = expected.detectorID == "archive-ar-v1" && strings.HasPrefix(expected.variant, "ar.metadata.") && baseDecision == DecisionReject
	case "tar_archive_metadata":
		valid = expected.detectorID == "archive-tar-v1" && strings.HasPrefix(expected.variant, "tar.metadata.")
	case "apple_bundle_forbidden":
		valid = expected.class == ClassAppleFramework || expected.class == ClassAppleXCFramework
	case "unsafe_entry", "unsafe_tree_entry":
		valid = expected.class == ClassLink || expected.class == ClassSpecial
	case "container_entry_rejected":
		valid = expected.class == ClassOpaqueUnknown || expected.class == ClassLink || expected.class == ClassSpecial ||
			expected.detectorID == "archive-tar-v1" && strings.HasPrefix(expected.variant, "tar.metadata.")
	case "capture_incomplete", "inspection_incomplete", "inspection_cancelled":
		valid = !node.InspectionComplete && expected.class == ClassOpaqueUnknown
	case "trust_role_not_authorized":
		valid = node.Parent == "" && manifest.Findings.Total > 0 &&
			(manifest.TrustRole == RoleExternalToolchain || manifest.TrustRole == RoleLocalBuildOutput ||
				manifest.TrustRole == RoleVerifiedBinaryCandidate)
	}
	if !valid && !node.InspectionComplete && hasStructuralRuleEvidence(manifest, node) {
		valid = true
	}
	if !valid {
		return fmt.Errorf("rule %q is not derived from node semantics", node.Rule)
	}
	return nil
}

func hasStructuralRuleEvidence(manifest Manifest, node ManifestNode) bool {
	for _, evidence := range manifest.Findings.Evidence {
		if evidence.Path != node.Path || evidence.Reason != node.Rule {
			continue
		}
		switch evidence.Code {
		case CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveEncrypted,
			CodeArchiveUnsafePath, CodeArchiveUnsafeEntry, CodeInspectionLimitExceeded,
			CodeInspectionUnavailable, CodePolicyInternalError:
			return true
		}
	}
	return false
}

func validNonRejectRule(node ManifestNode) bool {
	switch node.Rule {
	case "class_role_decision":
		return node.Kind == NodeRegularFile || (node.Kind == NodeArchive && node.Class == ClassNativeLibraryStatic)
	case "recursive_container_descent":
		return node.Kind == NodeArchive || node.Kind == NodeCompressedStream
	case "descend_directory":
		return node.Kind == NodeDirectory && node.Class == ClassDirectory
	case "native_archive_metadata":
		return node.Kind == NodeRegularFile && node.SelectedDetectorID == "archive-ar-v1"
	case "tar_archive_metadata":
		return node.Kind == NodeRegularFile && node.SelectedDetectorID == "archive-tar-v1"
	default:
		return false
	}
}

func hasDenyIndicatingFact(observations []Observation) bool {
	for _, observation := range observations {
		if _, ok := singleFactValue(observation.Facts, "deny_indicating_class"); ok {
			return true
		}
	}
	return false
}

func missingGeneratedLineage(observations []Observation) bool {
	for _, observation := range observations {
		if value, ok := singleFactValue(observation.Facts, "generated_lineage"); ok && value == "missing" {
			return true
		}
	}
	return false
}

func findObservation(observations []Observation, detectorID, result string) (Observation, bool) {
	for _, observation := range observations {
		if observation.DetectorID == detectorID && observation.Result == result {
			return observation, true
		}
	}
	return Observation{}, false
}

func factsExactly(input []Fact, expected map[string]string) bool {
	if len(input) != len(expected) {
		return false
	}
	for _, fact := range input {
		if expected[fact.Key] != fact.Value {
			return false
		}
	}
	return true
}

func onlyOptionalFact(input []Fact, key string) bool {
	if len(input) == 0 {
		return true
	}
	if len(input) != 1 || input[0].Key != key {
		return false
	}
	_, supported := artifactClasses[ArtifactClass(input[0].Value)]
	return supported && compiledClass(ArtifactClass(input[0].Value))
}

func singleFactValue(input []Fact, key string) (string, bool) {
	value := ""
	found := false
	for _, fact := range input {
		if fact.Key != key {
			continue
		}
		if found {
			return "", false
		}
		value, found = fact.Value, true
	}
	return value, found
}

func requiredDecimalFact(input []Fact, key string, minimum, maximum uint64) (uint64, bool) {
	value, ok := singleFactValue(input, key)
	if !ok || value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return parsed, err == nil && parsed >= minimum && parsed <= maximum
}

func booleanFact(input []Fact, key string) (bool, bool) {
	value, ok := singleFactValue(input, key)
	if !ok || (value != "true" && value != "false") {
		return false, false
	}
	return value == "true", true
}

func hasNonemptyFact(input []Fact) bool {
	for _, fact := range input {
		if fact.Key != "" && fact.Value != "" {
			return true
		}
	}
	return false
}

func knownUnsupportedContainer(format string) bool {
	switch format {
	case "7z", "xz", "bzip2", "zstd", "aix_big_archive", "iso9660", "dmg":
		return true
	default:
		return false
	}
}

func serializedSuffix(suffix string) bool {
	switch strings.ToLower(suffix) {
	case ".swiftmodule", ".swiftdoc", ".pcm", ".pch", ".rmeta", ".gch", ".ifc":
		return true
	default:
		return false
	}
}
