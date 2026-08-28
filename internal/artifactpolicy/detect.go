package artifactpolicy

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

type containerFormat string

const (
	formatNone        containerFormat = ""
	formatZIP         containerFormat = "zip"
	formatTar         containerFormat = "tar"
	formatGZIP        containerFormat = "gzip"
	formatAR          containerFormat = "ar"
	formatUnsupported containerFormat = "unsupported"
)

type detection struct {
	class        ArtifactClass
	variant      string
	detectorID   string
	format       containerFormat
	observations []Observation
	diagnostic   DiagnosticCode
	reason       string
}

type detectorCandidate struct {
	detection *detection
	observed  *Observation
}

func detectorIdentities() []DetectorIdentity {
	return []DetectorIdentity{
		{ID: "apple-bundle-path-v1", Version: "1"},
		{ID: "archive-ar-v1", Version: "1"},
		{ID: "archive-tar-v1", Version: "1"},
		{ID: "archive-zip-v1", Version: "1"},
		{ID: "compiler-serialized-v1", Version: "1"},
		{ID: "compression-gzip-v1", Version: "1"},
		{ID: "dex-v1", Version: "1"},
		{ID: "elf-v1", Version: "1"},
		{ID: "jvm-class-v1", Version: "1"},
		{ID: "llvm-bitcode-v1", Version: "1"},
		{ID: "macho-v1", Version: "1"},
		{ID: "pe-coff-v1", Version: "1"},
		{ID: "python-bytecode-v1", Version: "1"},
		{ID: "source-text-v1", Version: "1"},
		{ID: "unsupported-container-v1", Version: "1"},
		{ID: "v8-cache-v1", Version: "1"},
		{ID: "wasm-v1", Version: "1"},
	}
}

func (inspector *inspector) detect(item blob, virtualPath, descriptorPath string, uses []UseEdge) detection {
	if class, ok := appleBundleClass(virtualPath); ok {
		return detection{
			class: class, detectorID: "apple-bundle-path-v1",
			observations: []Observation{{
				DetectorID: "apple-bundle-path-v1", Result: "MATCH",
				Facts: []Fact{{Key: "bundle_path", Value: virtualPath}},
			}},
		}
	}

	prefix, err := item.prefix(64 << 10)
	if err != nil {
		return detection{
			class: ClassOpaqueUnknown, detectorID: "source-text-v1",
			diagnostic: CodeInspectionUnavailable, reason: "read_detector_prefix",
			observations: []Observation{{DetectorID: "source-text-v1", Result: "ERROR"}},
		}
	}
	if unsupported, reason := unsupportedContainer(prefix, item); unsupported {
		return detection{
			class: ClassOpaqueUnknown, detectorID: "unsupported-container-v1",
			format: formatUnsupported, diagnostic: CodeArchiveUnsupported, reason: reason,
			observations: []Observation{{
				DetectorID: "unsupported-container-v1", Result: "MATCH",
				Facts: []Fact{{Key: "format", Value: reason}},
			}},
		}
	}
	if isZIPPrefix(prefix) {
		result := detection{
			class: ClassArchive, detectorID: "archive-zip-v1", format: formatZIP,
			observations: []Observation{{DetectorID: "archive-zip-v1", Result: "MATCH"}},
		}
		return applyContainerNameAmbiguity(result, virtualPath)
	}
	if bytes.HasPrefix(prefix, []byte{0x1f, 0x8b}) {
		result := detection{
			class: ClassCompressedStream, detectorID: "compression-gzip-v1", format: formatGZIP,
			observations: []Observation{{DetectorID: "compression-gzip-v1", Result: "MATCH"}},
		}
		return applyContainerNameAmbiguity(result, virtualPath)
	}
	if bytes.HasPrefix(prefix, []byte("!<arch>\n")) || bytes.HasPrefix(prefix, []byte("!<thin>\n")) {
		return detection{
			class: ClassNativeLibraryStatic, detectorID: "archive-ar-v1", format: formatAR,
			observations: []Observation{{DetectorID: "archive-ar-v1", Result: "MATCH"}},
		}
	}
	if looksLikeTar(prefix, item.size) {
		result := detection{
			class: ClassArchive, detectorID: "archive-tar-v1", format: formatTar,
			observations: []Observation{{DetectorID: "archive-tar-v1", Result: "MATCH"}},
		}
		return applyContainerNameAmbiguity(result, virtualPath)
	}

	candidates := []detectorCandidate{
		detectELF(item, uses),
		detectPECOFF(item),
		detectMachO(item),
		detectJVMClass(item),
		detectDEX(item),
		detectWebAssembly(item),
		detectLLVMBitcode(item),
		detectCompilerSerialized(item, virtualPath),
		detectPythonBytecode(item, virtualPath),
		detectJavaScriptCache(item, virtualPath),
	}
	candidates = resolveCAFEBABECandidates(prefix, candidates)
	observations := make([]Observation, 0, len(candidates))
	valid := make([]detection, 0, 2)
	detectorFailed := false
	failedDetector := ""
	for _, candidate := range candidates {
		if candidate.observed != nil {
			observations = append(observations, *candidate.observed)
			if candidate.observed.Result == "ERROR" {
				detectorFailed = true
				if failedDetector == "" || candidate.observed.DetectorID < failedDetector {
					failedDetector = candidate.observed.DetectorID
				}
			}
		}
		if candidate.detection != nil {
			valid = append(valid, *candidate.detection)
		}
	}
	if len(valid) > 0 {
		sort.SliceStable(valid, func(left, right int) bool {
			return compiledClassPriority(valid[left].class) < compiledClassPriority(valid[right].class)
		})
		selected := valid[0]
		selected.observations = append(selected.observations, observations...)
		selected.class = specializeNativeExtension(selected.class, virtualPath)
		if detectorFailed {
			selected.detectorID = failedDetector
			selected.diagnostic = CodeOpaqueDependency
			selected.reason = "detector_error_prevented_complete_classification"
			if inspector.role == RoleDependencyInput && compiledClass(selected.class) {
				// A sound compiled match still owns the primary dependency
				// diagnostic, but the ERROR observation makes inspection
				// incomplete and prevents any role from receiving authority.
				selected.diagnostic = CodeCompiledDependency
				selected.reason = "compiled_match_with_detector_error"
			}
		}
		return selected
	}
	if detectorFailed {
		diagnostic := CodeOpaqueDependency
		reason := "recognized_format_failed_structural_validation"
		if denySuffixClass(virtualPath) != "" {
			diagnostic = CodeTypeAmbiguous
			reason = "deny_indicating_name_with_invalid_structural_candidate"
		}
		return detection{
			class: ClassOpaqueUnknown, detectorID: failedDetector,
			diagnostic: diagnostic, reason: reason,
			observations: observations,
		}
	}

	text := inspector.detectText(item, virtualPath, descriptorPath)
	text.observations = append(text.observations, observations...)
	if len(text.observations) > 0 && len(uses) > 0 {
		text.observations[0].Facts = append(text.observations[0].Facts, useFacts(uses)...)
		sortFacts(text.observations[0].Facts)
	}
	if denySuffixClass(virtualPath) != "" && (text.class != ClassOpaqueUnknown || textProfileMismatch(text.observations)) {
		text.class = ClassOpaqueUnknown
		text.diagnostic = CodeTypeAmbiguous
		text.reason = "deny_indicating_name_with_noncompiled_bytes"
	}
	if hasResolvedUse(uses, UseLinkOrLoad) {
		text.class = ClassOpaqueUnknown
		text.diagnostic = CodeTypeAmbiguous
		text.reason = "resolved_link_or_load_with_noncompiled_bytes"
	}
	if text.class == ClassOpaqueUnknown && text.diagnostic == "" && len(observations) > 0 {
		text.reason = "recognized_format_failed_structural_validation"
	}
	return text
}

// CAFEBABE is both the JVM class-file magic and the big-endian fat Mach-O
// magic. A structural failure from one grammar must not make a fully validated
// match from the other grammar incomplete. If both grammars validate, both
// observations remain and the normal deny-dominant ambiguity handling applies;
// if neither validates, both ERROR observations remain fail-closed evidence.
func resolveCAFEBABECandidates(prefix []byte, candidates []detectorCandidate) []detectorCandidate {
	const (
		machoCandidate = 2
		jvmCandidate   = 3
	)
	if len(prefix) < 4 || !bytes.Equal(prefix[:4], []byte{0xca, 0xfe, 0xba, 0xbe}) ||
		len(candidates) <= jvmCandidate {
		return candidates
	}
	macho := candidates[machoCandidate]
	jvm := candidates[jvmCandidate]
	if jvm.detection != nil && macho.observed != nil && macho.observed.Result == "ERROR" {
		candidates[machoCandidate] = detectorCandidate{}
	}
	if macho.detection != nil && jvm.observed != nil && jvm.observed.Result == "ERROR" {
		candidates[jvmCandidate] = detectorCandidate{}
	}
	return candidates
}

func hasResolvedUse(uses []UseEdge, kind UseKind) bool {
	for _, use := range uses {
		if use.Kind == kind {
			return true
		}
	}
	return false
}

func applyContainerNameAmbiguity(result detection, virtualPath string) detection {
	if claimed := denySuffixClass(virtualPath); claimed != "" {
		result.diagnostic = CodeTypeAmbiguous
		result.reason = "deny_indicating_name_with_container_bytes"
		result.observations[0].Facts = append(result.observations[0].Facts,
			Fact{Key: "deny_indicating_class", Value: string(claimed)})
		sortFacts(result.observations[0].Facts)
	}
	return result
}

func textProfileMismatch(observations []Observation) bool {
	for _, observation := range observations {
		if observation.DetectorID == "source-text-v1" && observation.Result == "NO_PROFILE_MATCH" {
			return true
		}
	}
	return false
}

func isZIPPrefix(prefix []byte) bool {
	return bytes.HasPrefix(prefix, []byte{'P', 'K', 3, 4}) ||
		bytes.HasPrefix(prefix, []byte{'P', 'K', 5, 6}) ||
		bytes.HasPrefix(prefix, []byte{'P', 'K', 7, 8})
}

func unsupportedContainer(prefix []byte, item blob) (bool, string) {
	signatures := []struct {
		magic []byte
		name  string
	}{
		{[]byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}, "7z"},
		{[]byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, "xz"},
		{[]byte{0x42, 0x5a, 0x68}, "bzip2"},
		{[]byte{0x28, 0xb5, 0x2f, 0xfd}, "zstd"},
		{[]byte("<bigaf>\n"), "aix_big_archive"},
	}
	for _, signature := range signatures {
		if bytes.HasPrefix(prefix, signature.magic) {
			return true, signature.name
		}
	}
	if len(prefix) >= 0x8006 && string(prefix[0x8001:0x8006]) == "CD001" {
		return true, "iso9660"
	}
	if item.size >= 512 {
		trailer := make([]byte, 4)
		if _, err := item.readAt(trailer, item.size-512); err == nil && string(trailer) == "koly" {
			return true, "dmg"
		}
	}
	return false, ""
}

func looksLikeTar(prefix []byte, size int64) bool {
	if size < 512 || len(prefix) < 512 {
		return false
	}
	if string(prefix[257:263]) == "ustar\x00" || string(prefix[257:263]) == "ustar " {
		return true
	}
	return validTarHeaderChecksum(prefix[:512])
}

func validTarHeaderChecksum(header []byte) bool {
	if len(header) != 512 || allZero(header) {
		return false
	}
	want, ok := parseTarNumber(header[148:156])
	if !ok {
		return false
	}
	var unsigned int64
	var signed int64
	for index, value := range header {
		if index >= 148 && index < 156 {
			value = ' '
		}
		unsigned += int64(value)
		signed += int64(int8(value))
	}
	return want == unsigned || want == signed
}

func allZero(payload []byte) bool {
	for _, value := range payload {
		if value != 0 {
			return false
		}
	}
	return true
}

func compiledClassPriority(class ArtifactClass) int {
	switch class {
	case ClassNodeExtension, ClassPythonExtension:
		return 1
	case ClassNativeExecutable, ClassNativeObject, ClassNativeLibraryStatic,
		ClassNativeLibraryDynamic, ClassELFETDYNAmbiguous:
		return 2
	case ClassJVMBytecode, ClassPythonBytecode, ClassJavaScriptCodeCache:
		return 3
	case ClassWebAssembly, ClassCompilerSerialized:
		return 4
	default:
		return 9
	}
}

func specializeNativeExtension(class ArtifactClass, virtualPath string) ArtifactClass {
	if class != ClassNativeExecutable && class != ClassNativeObject &&
		class != ClassNativeLibraryDynamic && class != ClassELFETDYNAmbiguous {
		return class
	}
	lower := strings.ToLower(path.Base(leafPath(virtualPath)))
	if strings.HasSuffix(lower, ".node") {
		return ClassNodeExtension
	}
	if strings.HasSuffix(lower, ".pyd") ||
		(strings.HasSuffix(lower, ".so") && (strings.Contains(lower, ".cpython-") || strings.Contains(lower, ".abi3"))) {
		return ClassPythonExtension
	}
	return class
}

func denySuffixClass(virtualPath string) ArtifactClass {
	lower := strings.ToLower(path.Base(leafPath(virtualPath)))
	extensions := []struct {
		suffix string
		class  ArtifactClass
	}{
		{".node", ClassNodeExtension}, {".pyd", ClassPythonExtension},
		{".so", ClassNativeLibraryDynamic}, {".dylib", ClassNativeLibraryDynamic},
		{".dll", ClassNativeLibraryDynamic}, {".a", ClassNativeLibraryStatic},
		{".lib", ClassNativeLibraryStatic}, {".rlib", ClassNativeLibraryStatic},
		{".o", ClassNativeObject}, {".obj", ClassNativeObject}, {".syso", ClassNativeObject},
		{".exe", ClassNativeExecutable}, {".class", ClassJVMBytecode},
		{".dex", ClassJVMBytecode}, {".wasm", ClassWebAssembly},
		{".bc", ClassCompilerSerialized}, {".swiftmodule", ClassCompilerSerialized},
		{".swiftdoc", ClassCompilerSerialized}, {".pcm", ClassCompilerSerialized},
		{".pch", ClassCompilerSerialized}, {".gch", ClassCompilerSerialized},
		{".ifc", ClassCompilerSerialized}, {".rmeta", ClassCompilerSerialized},
	}
	for _, extension := range extensions {
		if strings.HasSuffix(lower, extension.suffix) {
			return extension.class
		}
	}
	if strings.HasSuffix(lower, ".pyc") {
		return ClassPythonBytecode
	}
	return ""
}

func leafPath(virtualPath string) string {
	if separator := strings.LastIndex(virtualPath, "!/"); separator >= 0 {
		return virtualPath[separator+2:]
	}
	return virtualPath
}

func appleBundleClass(virtualPath string) (ArtifactClass, bool) {
	for _, component := range strings.FieldsFunc(virtualPath, func(character rune) bool {
		return character == '/' || character == '!'
	}) {
		lower := strings.ToLower(component)
		if strings.HasSuffix(lower, ".xcframework") {
			return ClassAppleXCFramework, true
		}
		if strings.HasSuffix(lower, ".framework") {
			return ClassAppleFramework, true
		}
	}
	return "", false
}

func facts(values map[string]any) []Fact {
	result := make([]Fact, 0, len(values))
	for key, value := range values {
		result = append(result, Fact{Key: key, Value: fmt.Sprint(value)})
	}
	sortFacts(result)
	return result
}

func useFacts(uses []UseEdge) []Fact {
	result := make([]Fact, 0, len(uses))
	for _, use := range uses {
		result = append(result, Fact{Key: "resolved_use", Value: string(use.Kind) + ":" + use.Origin})
	}
	sortFacts(result)
	return result
}

func parseDecimal(field []byte) (int64, bool) {
	text := strings.TrimSpace(string(field))
	if text == "" {
		return 0, false
	}
	value, err := strconv.ParseInt(text, 10, 64)
	return value, err == nil && value >= 0
}

func readExactAt(item blob, offset, size int64) ([]byte, error) {
	if offset < 0 || size < 0 || offset > item.size || size > item.size-offset {
		return nil, io.ErrUnexpectedEOF
	}
	payload := make([]byte, size)
	if size == 0 {
		return payload, nil
	}
	_, err := item.reader().ReadAt(payload, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return payload, nil
}
