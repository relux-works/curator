package swiftpminterop

import (
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Language is one closed SwiftPM-representable source language class.
type Language string

// The closed language set. SwiftPM represents exactly these five source
// languages; anything else is an unsupported target shape.
const (
	// LanguageSwift identifies Swift source.
	LanguageSwift Language = "swift"
	// LanguageC identifies C source.
	LanguageC Language = "c"
	// LanguageCXX identifies C++ source.
	LanguageCXX Language = "c++"
	// LanguageObjC identifies Objective-C source.
	LanguageObjC Language = "objective-c"
	// LanguageObjCXX identifies Objective-C++ source.
	LanguageObjCXX Language = "objective-c++"
)

// TargetKind is the capture-target family implied by a target's declaration.
type TargetKind string

// The closed target families this stage models.
const (
	// KindSwift is a Swift compiler capture target.
	KindSwift TargetKind = "swift"
	// KindClang is a C-family compiler capture target.
	KindClang TargetKind = "clang"
	// KindSystem is a system-library declaration of external trust.
	KindSystem TargetKind = "system"
)

// sourceLanguage maps one admitted relative source path to its language, or
// reports false for headers, module maps, resources, and other non-sources.
//
// Two extensions are matched case-sensitively because the Clang driver is
// case-sensitive on exactly those two: verified with `clang -### -c` on the
// accepted Darwin profile, `.C` selects `-x c++` and `.M` selects
// `-x objective-c++`, while `.c` and `.m` select C and Objective-C. Lowercasing
// them reported `[c]` and `[objective-c]` for a translation unit the compiler
// compiles as C++ and Objective-C++, which bypassed both C++ gates — the
// restricted-profile check and the `closure_interop_undeclared` rejection that
// fires when a Swift consumer declares no `.interoperabilityMode(.Cxx)`. Every
// other extension is matched case-insensitively, which is what the driver does
// with them.
func sourceLanguage(relative string) (Language, bool) {
	switch path.Ext(relative) {
	case ".C":
		return LanguageCXX, true
	case ".M":
		return LanguageObjCXX, true
	}
	switch strings.ToLower(path.Ext(relative)) {
	case ".swift":
		return LanguageSwift, true
	case ".c":
		return LanguageC, true
	case ".cc", ".cpp", ".cxx", ".c++", ".cp":
		return LanguageCXX, true
	case ".m":
		return LanguageObjC, true
	case ".mm":
		return LanguageObjCXX, true
	default:
		return "", false
	}
}

// assemblySource reports whether one admitted relative source path is an
// assembly translation unit. SwiftPM admits `.s` and `.S` as target sources and
// the pinned Apple Clang runs them through the integrated assembler, whose
// grammar this stage does not model: `.incbin "/etc/passwd"` in a `.S` file
// embeds those bytes in the object with no preprocessing directive at all, and
// `-H` reports no read. Scanning them with the C preprocessor grammar is also
// unsound in the opposite direction — a lowercase `.s` file is not preprocessed
// at all, so an ordinary `# comment` line in it would be rejected as an
// unclassifiable directive.
//
// No admitted SwiftPM shape in the accepted profile needs an assembly source,
// so a C-family target that declares one is unsupported rather than partially
// inspected.
func assemblySource(relative string) bool {
	switch strings.ToLower(path.Ext(relative)) {
	case ".s", ".asm":
		return true
	default:
		return false
	}
}

func isHeader(relative string) bool {
	switch strings.ToLower(path.Ext(relative)) {
	case ".h", ".hh", ".hpp", ".hxx", ".h++", ".def", ".inc":
		return true
	default:
		return false
	}
}

func isModuleMap(relative string) bool {
	base := path.Base(relative)
	return base == "module.modulemap" || base == "module.private.modulemap"
}

func cFamily(language Language) bool { return language != LanguageSwift }

// classifyTarget derives the capture-target family and exact language set of
// one declared target. A target that mixes Swift and any C-family source is
// rejected before any downstream compiler, module, or header analysis runs.
func classifyTarget(pkg, name string, target swiftpmsource.Target) (TargetKind, []Language, error) {
	switch target.Type {
	case "plugin":
		return "", nil, failFields(CodePluginUnsupported, map[string]string{"target": pkg + ":" + name}, "interop closure reaches a plugin target")
	case "macro":
		return "", nil, failFields(CodeMacroUnsupported, map[string]string{"target": pkg + ":" + name}, "interop closure reaches a macro target")
	case "binary":
		return "", nil, failFields(CodeBinaryUnavailable, map[string]string{"target": pkg + ":" + name}, "interop closure reaches a binary target")
	case "system-target", "system":
		return KindSystem, nil, nil
	}
	seen := map[Language]bool{}
	for _, source := range target.Sources {
		if assemblySource(source) {
			return "", nil, failFields(CodeTargetPlatformUnsupported, map[string]string{"target": pkg + ":" + name, "source": source}, "target declares an assembly source whose integrated-assembler grammar this profile does not model")
		}
		language, ok := sourceLanguage(source)
		if !ok {
			continue
		}
		seen[language] = true
	}
	languages := make([]Language, 0, len(seen))
	for language := range seen {
		languages = append(languages, language)
	}
	sort.Slice(languages, func(i, j int) bool { return languages[i] < languages[j] })
	swift, native := seen[LanguageSwift], false
	for _, language := range languages {
		if cFamily(language) {
			native = true
		}
	}
	if swift && native {
		return "", nil, failFields(CodeMixedLanguageTarget, map[string]string{"target": pkg + ":" + name, "languages": joinLanguages(languages)}, "one target declares both Swift and C-family sources")
	}
	if len(languages) == 0 {
		return "", nil, failFields(CodeGraphIncomplete, map[string]string{"target": pkg + ":" + name}, "target declares no classifiable source language")
	}
	if swift {
		return KindSwift, languages, nil
	}
	return KindClang, languages, nil
}

func joinLanguages(values []Language) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = string(value)
	}
	return strings.Join(parts, ",")
}

func languageStrings(values []Language) []string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = string(value)
	}
	sort.Strings(parts)
	return parts
}

func containsLanguage(values []Language, wanted Language) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
