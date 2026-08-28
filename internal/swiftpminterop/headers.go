package swiftpminterop

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/closuregraph"
)

// GeneratedModuleMapGrammarID identifies the reproduction of SwiftPM's own
// automatic module map for a conventional public-header layout.
const GeneratedModuleMapGrammarID = "swiftpm-generated-modulemap-v1"

// IncludeGrammarID identifies the C-family preprocessor reference scanner.
const IncludeGrammarID = "c-family-include-scanner-v10"

// inclusionDirectives are the directives portable mode positively admits as
// file-reading channels. Each one must carry an exact literal operand, which is
// then resolved, confined, and joined to the transitive include worklist. They
// are the whole of the admitted read surface: everything else a `#`-introduced
// line can spell is rejected by readDirective rather than dropped.
//
// `embed` is deliberately absent. It is the C23 resource-inclusion directive
// and it does paste the named file's bytes — verified on the accepted Darwin
// profile, where `#embed </etc/passwd>` and `#embed "../../../etc/passwd"` both
// read in the default GNU C mode SwiftPM selects — but its operand grammar
// carries parameters this stage does not model, no admitted SwiftPM C-family
// shape uses it, and admitting it would widen the read surface for nothing.
// Portable mode rejects it outright.
var inclusionDirectives = map[string]bool{"include": true, "include_next": true, "import": true}

// classifiableDirectives is the closed set of preprocessing directives this
// grammar recognizes that cannot themselves open a file. A `#`-introduced line
// outside this set and `inclusionDirectives` is a rejection, never a silently
// dropped line: the scanner is the only declared header-closure proof in
// portable mode, so an unrecognized directive spelling would otherwise be an
// admission hole rather than a diagnostic.
//
// The set is derived from the channel question "can this directive cause the
// compiler to read a file, or change where files are read from?", not from
// directive-name familiarity. `assert`, `unassert`, `define`, `undef`, `error`,
// `warning`, `ident`, `sccs`, and `line` name no file — `line` changes only the
// reported `__FILE__` and diagnostic position, never a search path. The
// conditional family carries `__has_include` and `__has_embed`, which are
// existence oracles that cannot introduce bytes, and whose operands this stage
// scans without evaluating by design. `pragma` stays listed because it is a
// recognized directive name, but it is never dropped here: every pragma body is
// classified by classifyPragmaBody before this set is consulted.
var classifiableDirectives = map[string]bool{
	"assert": true, "define": true, "elif": true, "elifdef": true, "elifndef": true,
	"else": true, "endif": true, "error": true, "ident": true,
	"if": true, "ifdef": true, "ifndef": true, "line": true, "pragma": true,
	"sccs": true, "unassert": true, "undef": true, "warning": true,
}

// trigraphCharacters is the complete nine-member trigraph set. Whether the
// compiler replaces them in translation phase 1 depends on the language mode:
// verified on the accepted Darwin profile, the pinned Apple Clang replaces them
// under `-std=c89/c99/c11/c17` and `-std=c++14`, and ignores them under the GNU
// modes SwiftPM selects when a target declares no language standard, under
// `-std=c++17`, and under the Objective-C++ default.
//
// Neither reading may be assumed. Without replacement `??=include </etc/passwd>`
// contains no `#` byte at all, so neither the directive grammar nor its
// unclassified-`#` backstop can see the inclusion an ISO-mode target performs.
// With replacement, `int a;??/` before a `#include` line splices the two lines
// and hides the inclusion a GNU-mode target performs. Both spellings were
// confirmed against the real compiler. A source that contains any trigraph is
// therefore rejected rather than classified under a mode this stage cannot bind
// per file: a target may compile C and Objective-C sources under one declared
// standard and headers shared between C and C++ translation units under two.
var trigraphCharacters = map[byte]bool{
	'=': true, '/': true, '\'': true, '(': true, ')': true,
	'!': true, '<': true, '>': true, '-': true,
}

// lineStartWhiteSpace is the non-ASCII white space the pinned Apple Clang skips
// without ending the white-space run that introduces a preprocessing directive,
// plus the UTF-8 BOM it ignores. Each code point was verified to keep
// `#include` a directive on the accepted Darwin profile.
func lineStartWhiteSpace(text string, cursor int) (int, bool) {
	if text[cursor] == 0 {
		// Clang diagnoses an embedded NUL and then skips it as white space.
		return 1, true
	}
	if text[cursor] < utf8.RuneSelf {
		return 0, false
	}
	character, width := utf8.DecodeRuneInString(text[cursor:])
	if character == utf8.RuneError && width <= 1 {
		return 0, false
	}
	switch {
	case character == '\uFEFF',
		character == '\u0085', character == '\u00A0', character == '\u1680',
		character >= '\u2000' && character <= '\u200A',
		character == '\u2028', character == '\u2029',
		character == '\u202F', character == '\u205F', character == '\u3000':
		return width, true
	}
	return 0, false
}

// HeaderFile is one admitted header retained with its exact content digest.
type HeaderFile struct {
	Relative string
	SHA256   closuregraph.ID
}

// ModuleMapEvidence is the complete parsed and confined evidence for one
// target's module map, whether custom or reproduced from SwiftPM's rules.
type ModuleMapEvidence struct {
	Package, Target   string
	Relative          string
	Generated         bool
	GrammarID         string
	SHA256            closuregraph.ID
	Parsed            ModuleMap
	ResolvedRefs      []Resolution
	PublicHeaderRoot  string
	PublicHeaderFiles []HeaderFile
}

// IncludeReference is one scanned preprocessor inclusion or Clang module
// import — the two reference shapes portable mode positively admits.
type IncludeReference struct {
	Package, Target string
	// SourcePackage names the package whose admitted root Source is relative
	// to. Once the include worklist crosses a package boundary the consuming
	// target's package no longer identifies the scanned file: two packages may
	// hold the same relative path, and the reference would not say which one
	// was opened.
	SourcePackage string
	Source        string
	Spelling      string
	Angled        bool
	ModuleImport  bool
	// ExpandedName records that the pinned compiler macro-expands the
	// identifier this reference names before resolving it, so Spelling is the
	// pre-expansion source token rather than the module the compiler imports.
	// It is true for `@import NAME` and for the Microsoft `__pragma(clang
	// module import NAME)` operator, and false for `#pragma clang module import
	// NAME` and for the destringized `_Pragma` operand — the asymmetry is
	// verified on the accepted Darwin profile, not assumed. See
	// rejectMacroDefinedModuleNames for what the flag gates.
	ExpandedName bool
}

// publicHeaderRoot derives the declared public header directory of one
// C-family target relative to the package root, applying SwiftPM's documented
// `include` default only when the target declared no publicHeadersPath. A
// declaration this profile cannot represent exactly is rejected rather than
// replaced by the default: substituting a default would inventory, generate a
// module map for, and confine the wrong directory while the directory that
// actually governs the target is never parsed.
func publicHeaderRoot(pkg, name, targetPath, declared string) (string, error) {
	reject := func() error {
		return failFields(CodeTargetPlatformUnsupported, map[string]string{"target": pkg + ":" + name, "public_headers_path": declared}, "declared publicHeadersPath is not a contained relative directory this profile can represent")
	}
	base := path.Clean(targetPath)
	if base == "" || base == "." {
		base = ""
	}
	value := declared
	if value == "" {
		value = "include"
	}
	if strings.ContainsAny(value, "\x00\r\n\\") || strings.HasPrefix(value, "/") || windowsAbsolute(value) {
		return "", reject()
	}
	cleaned := path.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", reject()
	}
	if base == "" {
		return cleaned, nil
	}
	joined := path.Join(base, cleaned)
	if joined != base && !strings.HasPrefix(joined, base+"/") {
		return "", reject()
	}
	return joined, nil
}

// confineModuleMapLayout rejects a C-family target whose admitted tree holds a
// module map outside the resolved public-header root. Such a map is one SwiftPM
// or Clang may honour while this stage would never parse, confine, or digest
// it, so the layout is unsupported rather than silently partially inspected.
func confineModuleMapLayout(pkg, name, packageRoot, treeRoot, publicHeaderRelative string) error {
	absolute := filepath.Join(packageRoot, filepath.FromSlash(treeRoot))
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() {
		return nil
	}
	expected := path.Clean(publicHeaderRelative)
	return filepath.WalkDir(absolute, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(packageRoot, current)
		if relErr != nil {
			return relErr
		}
		logical := filepath.ToSlash(relative)
		if !isModuleMap(logical) || path.Dir(logical) == expected {
			return nil
		}
		return failFields(CodeModuleMapEscape, map[string]string{"target": pkg + ":" + name, "module_map": logical, "public_header_root": expected}, "admitted target tree contains a module map outside the resolved public-header root")
	})
}

// inventoryHeaders lists every admitted header below root, sorted, with exact
// content digests. Symlinks and non-regular nodes reject: SwiftPM's own source
// enumeration omits headers entirely, so this inventory is the authority.
func inventoryHeaders(packageRoot, relativeRoot string) ([]HeaderFile, error) {
	absolute := filepath.Join(packageRoot, filepath.FromSlash(relativeRoot))
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil
	}
	files := []HeaderFile{}
	walkErr := filepath.WalkDir(absolute, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		entryInfo, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 || !entryInfo.Mode().IsRegular() {
			return failFields(CodeHeaderInputUndeclared, map[string]string{"path": current}, "public header tree contains a linked or special node")
		}
		relative, relErr := filepath.Rel(packageRoot, current)
		if relErr != nil {
			return relErr
		}
		logical := filepath.ToSlash(relative)
		if !isHeader(logical) && !isModuleMap(logical) {
			return nil
		}
		payload, readErr := os.ReadFile(current) // #nosec G304 -- admitted immutable header below a verified protected package root.
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(payload)
		files = append(files, HeaderFile{Relative: logical, SHA256: closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Relative < files[j].Relative })
	return files, nil
}

// GenerateModuleMap reproduces SwiftPM's documented automatic module map for a
// conventional C-family public-header layout: an umbrella header named after
// the module, a single directly contained header, or an umbrella directory.
func GenerateModuleMap(moduleName, publicHeaderRelative string, headers []HeaderFile) (string, error) {
	direct := []string{}
	for _, header := range headers {
		if isModuleMap(header.Relative) {
			continue
		}
		if path.Dir(header.Relative) == path.Clean(publicHeaderRelative) {
			direct = append(direct, header.Relative)
		}
	}
	sort.Strings(direct)
	umbrella := path.Join(publicHeaderRelative, moduleName+".h")
	for _, candidate := range direct {
		if candidate == umbrella {
			return fmt.Sprintf("module %s {\n    umbrella header \"%s\"\n    export *\n}\n", c99Identifier(moduleName), path.Base(candidate)), nil
		}
	}
	if len(direct) == 1 {
		return fmt.Sprintf("module %s {\n    umbrella header \"%s\"\n    export *\n}\n", c99Identifier(moduleName), path.Base(direct[0])), nil
	}
	if len(headers) == 0 {
		return "", failFields(CodeGraphIncomplete, map[string]string{"module": moduleName}, "C-family target has no admitted public headers and no custom module map")
	}
	return fmt.Sprintf("module %s {\n    umbrella \".\"\n    export *\n}\n", c99Identifier(moduleName)), nil
}

func c99Identifier(value string) string {
	var builder strings.Builder
	for index, character := range value {
		switch {
		case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z', character == '_':
			builder.WriteRune(character)
		case character >= '0' && character <= '9' && index > 0:
			builder.WriteRune(character)
		default:
			builder.WriteByte('_')
		}
	}
	return builder.String()
}

// scanIncludes reads one admitted C-family source, header, or transitively
// included file and returns every preprocessor and Clang module reference it
// names. Recognition runs on the translation the target compiler actually
// performs before directives execute — line-ending normalization, phase-2
// backslash-newline splicing, and phase-3 comment replacement — because the
// pinned Apple Clang honours `#inc\`<newline>`lude`, `/* */ #include`, a
// form-feed prefix, and the `%:` digraph exactly as it honours `#include`.
// Phase-1 trigraph replacement is mode-dependent, so a source that contains a
// trigraph is rejected instead of being translated under an assumed mode.
//
// Recognition is reject-by-default on the channel axis. Portable mode
// positively admits a literal `#include`/`#import`/`#include_next` operand, the
// `@import` and `#pragma clang module import` module spellings including their
// `_Pragma`/`__pragma` operator forms and the tokens of those hidden inside a
// macro definition, and the closed pragma allowlist. Every other `#`-introduced
// line, every pragma spelling outside that allowlist, `#embed`, any
// inline-assembly construct, and every C++ raw string reject the target. A
// file-read channel discovered after this is written therefore fails closed
// without new emulation, because it cannot be inside an allowlist nobody added
// it to.
//
// Phase 4 — macro expansion — is closed the same way rather than reproduced.
// The closure argument is stated over preprocessing tokens *including their
// alternative spellings*, because a spelling-level statement is not a statement
// about tokens: `%:%:` and `##` are one operator, and a rule naming only one of
// them leaves the other open.
//
// The scanner reads source tokens, so expansion can reach the compiler in
// exactly two ways the source does not spell.
//
//  1. A NEW token built from fragments. Adjacent tokens never merge — verified,
//     `#define A __as` + `#define B m__` + `A B(…)` is an unknown type name and
//     yields an object with no payload bytes — so the only builder is the paste
//     operator, in either spelling. macroPasteWidth reads both;
//     collapseMacroPastes performs a paste whose fragments are fixed, so the
//     result is scanned like any other token stream, and rejects a paste taking
//     a fragment from the call site, which no body-local analysis can bound.
//     Macro output is not re-scanned for directives — verified, `#define INC
//     #include "…"` invoked as `INC` is a syntax error and reads nothing — so a
//     built token can only reach a token-level channel, and every one of those
//     rejects.
//
//  2. An EXISTING identifier the compiler expands in a position this scanner
//     reads literally. That is an evidence-integrity question, not a keyword
//     one, and no rule about which keywords may follow `@` reaches it. It is
//     settled instead by enumerating every identifier position the scanner
//     records or gates on and pinning each one's disposition on the accepted
//     Darwin profile:
//
//     | position | expanded? | disposition |
//     | --- | --- | --- |
//     | `#include`/`#import`/`#include_next` operand | no | the literal `"…"`/`<…>` header name is used as written; any other operand is the computed-include form and rejects |
//     | directive name after `#` or `%:` | no | classified literally against a closed set |
//     | `@`-follower identifier | YES | rejected at the binding — a `#define` or a `.define` build setting: atPositionIdentifiers refuses any identifier valid in that position |
//     | `@import` module name | YES | rejected when macro-defined anywhere in the target's macro oracle — the scanned closure's `#define`s plus its selected `.define` settings — rejectMacroDefinedModuleNames |
//     | `#pragma` head and operands | no | recorded literally; verified for the head and for the module name |
//     | `_Pragma` string operand and the tokens it destringizes to | no | recorded literally; a non-literal operand rejects |
//     | `__pragma` tokens | YES | same rejection as `@import`; the one pragma spelling that expands |
//     | module-map module, header, and extern names | n/a | parsed from a module map, which no preprocessing phase touches |
//
//     Every other channel — inline assembly, `#embed`, a pragma outside the
//     allowlist, an unclassifiable `#`-line, a raw string — rejects outright, so
//     it exposes no identifier position to expand into.
//
//     The two rows that close by asking "is this identifier macro-defined?"
//     are only as complete as the ORACLE that answers it, so the oracle's
//     provenance is part of the argument rather than an implementation detail.
//     It is the union of two inputs: every source `#define` the target's
//     scanned closure binds, recorded by noteMacro at this site, and every
//     `.define` build setting the destination selects for the target, routed
//     by disposeBuildSettings.
//
//     The second input exists because a SwiftPM `.define` reaches the compiler as `-D`
//     without appearing in any admitted file: verified on the accepted Darwin
//     profile, `-Dprotocol=import` with `@ protocol SecretKit;` builds the
//     module and reads its header, and `-DNoSuchKitXYZ=SecretKit` with
//     `@import NoSuchKitXYZ;` does the same while the recorded evidence names
//     `NoSuchKitXYZ`. A source-only oracle reproduces both closed positions
//     one level down.
//
//     The build-setting axis behind input 2 is itself enumerated
//     reject-by-default the way the pragma axis is: settingKindDisposition
//     admits a kind only when it is provably macro-inert and resolution-inert,
//     `define` is routed through both rules above, `headerSearchPath` and
//     `unsafeFlags` reject as the only kinds that reach `-I` or an unbounded
//     flag, and an unknown kind rejects. The oracle therefore cannot be
//     bypassed by a channel that binds a macro without spelling `#define`.
//
//     Both inputs are checked on BOTH halves of what they bind — the macro's
//     NAME, against the two rules above, and the macro's BODY, against the
//     channel analyzer analyzeMacroBody, which every macro-binding input calls.
//     A name check alone leaves the replacement list free, and round 9 finding
//     M established that a replacement list is itself a channel: verified on
//     the accepted Darwin profile, `-D'A=__asm__'` with
//     `A(".incbin \"payload.bin\"");` reads the named file exactly as the
//     source `#define A __asm__` control does. The macro-INPUT surface for an
//     admitted C-family target is exactly these two inputs — the pinned
//     `clang -c` binds a macro only from a source directive or from `-D`, and
//     `-D` is reachable only through the `define` kind — so with the name and
//     the body of each routed through the same reject logic, the macro layer is
//     closed across every input surface, not only every spelling and position.
func scanIncludes(pkg, target, sourcePkg, packageRoot, relative string) (scanResult, error) {
	absolute := filepath.Join(packageRoot, filepath.FromSlash(relative))
	payload, err := os.ReadFile(absolute) // #nosec G304 -- admitted immutable source below a verified protected package root.
	if err != nil {
		return scanResult{}, failFields(CodeSourceInventoryDrift, map[string]string{"target": pkg + ":" + target, "source": relative}, "declared or transitively included input is absent or unreadable in the admitted tree")
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(payload), "\r\n", "\n"), "\r", "\n")
	if index, found := findTrigraph(text); found {
		return scanResult{}, failFields(CodeHeaderInputUndeclared, map[string]string{"target": pkg + ":" + target, "source": relative, "trigraph": text[index : index+3]}, "source contains a trigraph sequence whose phase-1 replacement depends on a language mode this stage cannot bind per file")
	}
	scanner := &directiveScanner{pkg: pkg, target: target, sourcePkg: sourcePkg, source: relative, text: spliceTranslationLines(text), atLineStart: true}
	if err = scanner.run(); err != nil {
		return scanResult{}, err
	}
	references := scanner.references
	sort.Slice(references, func(i, j int) bool {
		left := references[i].Source + "\x00" + references[i].Spelling
		right := references[j].Source + "\x00" + references[j].Spelling
		return left < right
	})
	return scanResult{references: references, macros: scanner.macros}, nil
}

// scanResult is one scanned file's contribution to the target's header proof:
// the references it names, and the macro names it binds. The macro set leaves
// this function because one identifier position the scanner records — the
// `@import` module name — is macro-expanded by the compiler before it resolves,
// so whether the recorded spelling is the module that is actually imported is
// not a per-file question. The definition may sit in any file of the target's
// scanned closure.
type scanResult struct {
	references []IncludeReference
	macros     map[string]bool
}

// spliceTranslationLines performs the translation phases that precede directive
// recognition: every line ending becomes a newline and every line splice is
// removed. A splice is a backslash followed by a run of horizontal white space
// and a newline, which is exactly what the pinned Apple Clang joins.
//
// The white-space form is spliced unconditionally rather than left as a
// residual to fail closed on. Leaving it unspliced was an admission hole, not a
// conservative divergence: a residual only exists when the split lands inside a
// construct recognized by line position — a `#`-introduced directive or a
// string literal, both of which still reject — whereas startsAsmStatement,
// startsPragmaOperator, and startsModuleImport all match their keyword by
// prefix at any column, so a split *inside the keyword* left no residual at
// all. Verified on the accepted Darwin profile: `__as\`+ws+newline+`m__(
// ".incbin \"payload.bin\"")` produces an object byte-identical to the direct
// `.incbin` control, and the `_Pragma`, `@import`, and `#include` splits behave
// the same. Unlike phase-1 trigraph replacement this is mode-independent —
// confirmed across six separator spellings and twelve `-std`/language modes —
// so there is nothing to bind per file and no reason to reject instead.
// Carriage returns need no separate case because the normalization above has
// already folded them.
func spliceTranslationLines(payload string) string {
	text := strings.ReplaceAll(payload, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if !strings.Contains(text, "\\") {
		return text
	}
	var builder strings.Builder
	builder.Grow(len(text))
	for cursor := 0; cursor < len(text); {
		if text[cursor] != '\\' {
			builder.WriteByte(text[cursor])
			cursor++
			continue
		}
		if end := skipHorizontal(text, cursor+1); end < len(text) && text[end] == '\n' {
			cursor = end + 1
			continue
		}
		builder.WriteByte('\\')
		cursor++
	}
	return builder.String()
}

// findTrigraph returns the index of the first trigraph sequence. The scan is
// leftmost so that `???=` reports the second `??` pair, exactly the pair the
// compiler would replace.
func findTrigraph(text string) (int, bool) {
	for cursor := 0; cursor+2 < len(text); cursor++ {
		if text[cursor] == '?' && text[cursor+1] == '?' && trigraphCharacters[text[cursor+2]] {
			return cursor, true
		}
	}
	return 0, false
}

// directiveScanner walks one translated input, tracking whether the cursor is
// at the start of a logical line so that a directive introduced after a
// comment — including a comment that spans lines — is still recognized.
type directiveScanner struct {
	pkg, target, sourcePkg, source string
	text                           string
	index                          int
	atLineStart                    bool
	rawString                      bool
	references                     []IncludeReference
	macros                         map[string]bool
}

// noteMacro records one macro name this file binds.
func (s *directiveScanner) noteMacro(name string) {
	if name == "" {
		return
	}
	if s.macros == nil {
		s.macros = map[string]bool{}
	}
	s.macros[name] = true
}

func (s *directiveScanner) run() error {
	for s.index < len(s.text) {
		character := s.text[s.index]
		switch {
		case character == '\n':
			s.index++
			s.atLineStart = true
		case horizontalSpace(character):
			s.index++
		case s.atLineStart && (character == 0 || character >= utf8.RuneSelf):
			if err := s.readLineStartPrefix(); err != nil {
				return err
			}
		case s.hasPrefix("/*"):
			s.skipBlockComment()
		case s.hasPrefix("//"):
			s.skipLineComment()
		case rawStringPrefixAt(s.text, s.index):
			return s.rejectRawString()
		case character == '"' || character == '\'':
			s.index = literalEnd(s.text, s.index)
			s.atLineStart = false
		case s.startsModuleImport():
			if err := s.readModuleImport(); err != nil {
				return err
			}
			s.atLineStart = false
		case character == '@':
			if err := s.readAtToken(); err != nil {
				return err
			}
			s.atLineStart = false
		case s.startsPragmaOperator():
			if err := s.readPragmaOperator(); err != nil {
				return err
			}
			s.atLineStart = false
		case s.startsAsmStatement():
			if err := s.rejectAsmStatement(); err != nil {
				return err
			}
			s.atLineStart = false
		case s.atLineStart && (character == '#' || s.hasPrefix("%:")):
			if err := s.readDirective(); err != nil {
				return err
			}
		default:
			s.index++
			s.atLineStart = false
		}
	}
	return nil
}

// readLineStartPrefix consumes the run of bytes that introduces a logical line
// when that run begins with a byte this grammar cannot read as ASCII white
// space. The pinned Apple Clang skips an embedded NUL, a UTF-8 BOM, and its
// Unicode white-space set without ending the white-space run that introduces a
// directive, so treating those bytes as content silently demoted the directive
// on that line to content and dropped the header it named. A byte sequence that
// is neither demotes the line only when no directive follows it; when one does,
// the line is rejected rather than classified on a guess.
func (s *directiveScanner) readLineStartPrefix() error {
	cursor := s.index
	unclassified := false
	for cursor < len(s.text) {
		if horizontalSpace(s.text[cursor]) {
			cursor++
			continue
		}
		if width, ok := lineStartWhiteSpace(s.text, cursor); ok {
			cursor += width
			continue
		}
		if s.text[cursor] < utf8.RuneSelf {
			break
		}
		_, width := utf8.DecodeRuneInString(s.text[cursor:])
		if width < 1 {
			width = 1
		}
		cursor += width
		unclassified = true
	}
	if unclassified && cursor < len(s.text) && (s.text[cursor] == '#' || strings.HasPrefix(s.text[cursor:], "%:")) {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": lineSnippet(s.text, s.index)}, "byte sequence before a preprocessing directive is not white space this scanner grammar can classify")
	}
	s.index = cursor
	if unclassified {
		s.atLineStart = false
	}
	return nil
}

func (s *directiveScanner) hasPrefix(value string) bool {
	return strings.HasPrefix(s.text[s.index:], value)
}

// startsModuleImport reports whether the cursor sits on the `@` of a Clang
// `@import`, at any column. Anchoring recognition to the line start dropped
// `int x = 1; @import Secret;` silently, and requiring the two tokens to be
// adjacent dropped `@ import Secret;` and `@/*c*/import Secret;` — both of
// which the pinned compiler imports, verified with `-fmodules` against a module
// whose header is an `#error` marker.
func (s *directiveScanner) startsModuleImport() bool {
	if s.text[s.index] != '@' {
		return false
	}
	cursor := s.skipTrivia(s.index + 1)
	end := cursor
	for end < len(s.text) && identifierByte(s.text[end], end > cursor) {
		end++
	}
	return s.text[cursor:end] == "import"
}

// skipTrivia advances over the white space and comments that may separate two
// preprocessing tokens, including line breaks and the non-ASCII white space the
// pinned compiler skips. It is only used between tokens whose separation the
// compiler ignores; directive recognition still runs on logical lines.
func (s *directiveScanner) skipTrivia(cursor int) int {
	for cursor < len(s.text) {
		character := s.text[cursor]
		if character == '\n' || horizontalSpace(character) {
			cursor++
			continue
		}
		if width, ok := lineStartWhiteSpace(s.text, cursor); ok {
			cursor += width
			continue
		}
		switch {
		case strings.HasPrefix(s.text[cursor:], "/*"):
			end := strings.Index(s.text[cursor+2:], "*/")
			if end < 0 {
				return len(s.text)
			}
			cursor += 2 + end + 2
		case strings.HasPrefix(s.text[cursor:], "//"):
			next := strings.IndexByte(s.text[cursor:], '\n')
			if next < 0 {
				return len(s.text)
			}
			cursor += next + 1
		default:
			return cursor
		}
	}
	return cursor
}

// skipBlockComment consumes one block comment. A comment is white space, so a
// directive may follow it on the same physical line; a comment that spans lines
// also starts a new logical line for the token that follows it. Both were
// verified against the pinned Apple Clang.
func (s *directiveScanner) skipBlockComment() {
	end := strings.Index(s.text[s.index+2:], "*/")
	if end < 0 {
		s.index = len(s.text)
		return
	}
	body := s.text[s.index : s.index+2+end+2]
	s.index += 2 + end + 2
	if strings.Contains(body, "\n") {
		s.atLineStart = true
	}
}

func (s *directiveScanner) skipLineComment() {
	next := strings.IndexByte(s.text[s.index:], '\n')
	if next < 0 {
		s.index = len(s.text)
		return
	}
	s.index += next
}

// readModuleImport resolves one `@import` to an exact module name. The keyword,
// the name, and the terminating semicolon may be separated by any white space
// or comment — each separator was verified to import on the pinned compiler. A
// spelling this grammar cannot resolve is a rejection: Clang would still import
// it.
func (s *directiveScanner) readModuleImport() error {
	start := s.index
	cursor := s.skipTrivia(s.skipTrivia(s.index+1) + len("import"))
	name, next, ok := readModuleName(s.text, cursor)
	if ok {
		next = s.skipTrivia(next)
		if next < len(s.text) && s.text[next] == ';' {
			s.references = append(s.references, IncludeReference{Package: s.pkg, Target: s.target, SourcePackage: s.sourcePkg, Source: s.source, Spelling: name, ModuleImport: true, ExpandedName: true})
			s.index = next + 1
			return nil
		}
	}
	return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": lineSnippet(s.text, start)}, "Clang module import is not an exact module name this scanner grammar can resolve")
}

// objcAtKeywords is the closed set of identifiers Objective-C allows directly
// after `@`. `import` is deliberately absent because startsModuleImport already
// owns that spelling, and so is `__experimental_modules_import`, Clang's older
// spelling of the same module import.
//
// The set exists so that `@` followed by any other identifier can reject. That
// identifier is not valid Objective-C on its own, so it is a macro, and a macro
// that expands to `import` performs an unconfined module import: verified on
// the accepted Darwin profile, `#define I import` followed by `@ I SecretKit;`
// builds the module and reads its header, while the scanner saw only an `@`
// and an ordinary identifier.
//
// The set is a recognizer, not the closure: every member is an ordinary
// identifier the preprocessor can rebind to `import`, so the rule that makes
// this position safe is atPositionIdentifiers, which refuses the `#define`.
var objcAtKeywords = map[string]bool{
	"autoreleasepool": true, "available": true, "catch": true, "class": true,
	"compatibility_alias": true, "defs": true, "dynamic": true, "encode": true,
	"end": true, "finally": true, "implementation": true, "interface": true,
	"optional": true, "package": true, "private": true, "property": true,
	"protected": true, "protocol": true, "public": true, "required": true,
	"selector": true, "synchronized": true, "synthesize": true, "throw": true,
	"try": true, "NO": true, "YES": true, "true": true, "false": true,
}

// atPositionIdentifiers is the closed set of identifiers whose macro definition
// portable mode refuses, because each one is an identifier the compiler expands
// in a position this scanner reads literally.
//
// It is `objcAtKeywords` plus the two module-import spellings. The `@`-follower
// is macro-expanded before the language sees it — verified on the accepted
// Darwin profile: `#define protocol import` with `@ protocol SecretKit;` builds
// the module and reads its header, and so do the `class`, `selector`, `end`,
// and `YES` spellings, and `#define class im##port` composes past the paste
// layer and the `@`-keyword layer alike. Rejecting `@ IDENT` for an identifier
// outside the keyword set therefore does not close the channel: every member of
// the set is an ordinary identifier the preprocessor will happily rebind.
//
// The rejection is placed on the `#define` rather than on the `@` because the
// definition is the only end of this shape that is decidable from one file. A
// macro bound in a header and used in a `.c` file is the realistic vector, and
// a rule that asks "is this identifier macro-defined?" at the `@` cannot answer
// it without the whole translation unit; a rule that asks "does this definition
// bind a name the compiler expands after `@`?" answers it wherever the
// definition sits, because every admitted file of the closure is scanned.
//
// The two import spellings are here for evidence integrity rather than for a
// file read: `#define import protocol` followed by `@import SecretKit;`
// compiles and imports nothing — verified — while the scanner would record a
// module import that never happened.
//
// The narrowing is real and deliberate: `#define interface struct`, the Windows
// COM idiom, and a package-local `#define true 1` C89 shim are rejected. No
// admitted SwiftPM C-family shape needs either, and both fail in the safe
// direction.
var atPositionIdentifiers = func() map[string]bool {
	names := map[string]bool{"import": true, "__experimental_modules_import": true}
	for keyword := range objcAtKeywords {
		names[keyword] = true
	}
	return names
}()

// readAtToken classifies one `@` that does not introduce a literal `@import`.
//
// Objective-C's `@`-keywords are a closed set and everything else `@` may
// introduce is a literal or a collection: `@"…"`, `@'c'`, `@[…]`, `@{…}`,
// `@(…)`, `@42`. An identifier outside that set can only be a macro, and phase
// 4 expands it before the language sees it, so portable mode rejects rather
// than consuming the `@` as content the way earlier rounds did.
func (s *directiveScanner) readAtToken() error {
	reject := func() error {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": lineSnippet(s.text, s.index)}, "`@` introduces a token this scanner grammar cannot classify, and a macro in that position expands to a Clang module import")
	}
	cursor := s.skipTrivia(s.index + 1)
	if cursor >= len(s.text) {
		return reject()
	}
	if identifier, _ := splitLeadingIdentifier(s.text[cursor:]); identifier != "" {
		if !objcAtKeywords[identifier] {
			return reject()
		}
		s.index++
		return nil
	}
	switch character := s.text[cursor]; {
	case character == '"', character == '\'', character == '[', character == '{', character == '(',
		character >= '0' && character <= '9':
		s.index++
		return nil
	}
	return reject()
}

// rawStringPrefixAt reports whether the delimiter at quote opens a C++ raw
// string literal. The five encoding prefixes are recognized at a token
// boundary, so an ordinary identifier that merely ends in `R` is not one.
func rawStringPrefixAt(text string, quote int) bool {
	if text[quote] != '"' {
		return false
	}
	start := quote
	for start > 0 && identifierByte(text[start-1], true) {
		start--
	}
	switch text[start:quote] {
	case "R", "LR", "uR", "UR", "u8R":
		return true
	}
	return false
}

// rejectRawString rejects a C++ raw string literal.
//
// A raw string suspends every escape rule this grammar relies on, and whether
// `R"` opens one at all depends on the language mode — the same mode-dependence
// that makes a trigraph a rejection rather than a translation. Left unmodeled
// it was not merely conservative: `R"x(" /* )x"` hands the scanner an unmatched
// `"` followed by `/*`, skipBlockComment then swallows the rest of the file,
// and the compiler sees no comment at all. Verified on the accepted Darwin
// profile — that prologue followed by `__asm__(".incbin \"payload.bin\"")`
// compiles and puts the named file's bytes in the object while the scanner
// reads nothing after the prologue.
//
// The shared artifact classifier independently rejects several raw-string
// spellings as opaque, but that defense is incidental to this grammar and
// cannot be relied on here, so the construct is rejected at this boundary too.
func (s *directiveScanner) rejectRawString() error {
	return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "construct": lineSnippet(s.text, s.index)}, "source contains a C++ raw string literal, whose delimiter grammar and language-mode dependence this scanner cannot resolve")
}

// startsPragmaOperator reports whether the cursor sits on a `_Pragma` or
// `__pragma` keyword at a token boundary. Neither is a `#`-introduced line, so
// the directive grammar never sees them, yet `_Pragma("clang module import
// SecretKit")` in a plain `.c` file builds and reads the module, and the
// Microsoft `__pragma(clang module import SecretKit)` spelling does the same
// under `-fms-extensions` — both verified against a module whose only header is
// an `#error` marker.
func (s *directiveScanner) startsPragmaOperator() bool {
	keyword := s.pragmaKeyword()
	if keyword == "" {
		return false
	}
	if s.index > 0 && identifierByte(s.text[s.index-1], true) {
		return false
	}
	next := s.index + len(keyword)
	return next >= len(s.text) || !identifierByte(s.text[next], true)
}

func (s *directiveScanner) pragmaKeyword() string {
	for _, keyword := range []string{"_Pragma", "__pragma"} {
		if strings.HasPrefix(s.text[s.index:], keyword) {
			return keyword
		}
	}
	return ""
}

// readPragmaOperator consumes one `_Pragma` operator and classifies its
// destringized operand. Only the exact `_Pragma ( "…" )` form is resolvable:
// an encoding-prefixed or raw literal, a macro operand, or a missing
// parenthesis names a pragma this grammar cannot read, and `_Pragma(M)` with
// `M` defined as the module-import string does import on the pinned compiler,
// so an unresolvable operand is a rejection rather than a dropped token.
func (s *directiveScanner) readPragmaOperator() error {
	start := s.index
	keyword := s.pragmaKeyword()
	reject := func() error {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": lineSnippet(s.text, start)}, "%s operand is not one this scanner grammar can classify", keyword)
	}
	cursor := s.skipTrivia(s.index + len(keyword))
	if cursor >= len(s.text) || s.text[cursor] != '(' {
		return reject()
	}
	if keyword == "__pragma" {
		// The Microsoft spelling takes raw preprocessing tokens rather than a
		// string literal, so the operand is the balanced parenthesis run.
		end, ok := balancedParenthesis(s.text, cursor)
		if !ok {
			return reject()
		}
		s.index = end
		// The Microsoft operator takes raw preprocessing tokens and those tokens
		// ARE macro-expanded: verified, `#define AliasKit SecretKit` +
		// `__pragma(clang module import AliasKit)` under `-fms-extensions` builds
		// SecretKit and reads its header. This is the one pragma spelling that
		// shares `@import`'s expansion disposition.
		return s.applyPragma(s.text[cursor+1:end-1], lineSnippet(s.text, start), true)
	}
	cursor = s.skipTrivia(cursor + 1)
	if cursor >= len(s.text) || s.text[cursor] != '"' {
		return reject()
	}
	end := literalEnd(s.text, cursor)
	if end <= cursor+1 || s.text[end-1] != '"' {
		return reject()
	}
	body := destringizePragma(s.text[cursor+1 : end-1])
	next := s.skipTrivia(end)
	if next >= len(s.text) || s.text[next] != ')' {
		return reject()
	}
	s.index = next + 1
	// `_Pragma`'s string operand is not macro-expanded and neither are the
	// tokens it destringizes to: verified, `#define X SecretKit` +
	// `_Pragma("clang module import X")` reports module `X` not found.
	return s.applyPragma(body, lineSnippet(s.text, start), false)
}

// balancedParenthesis returns the index just past the `)` that closes the `(`
// at open, ignoring parentheses inside string and character literals. An
// unbalanced run has no operand this grammar can read.
func balancedParenthesis(text string, open int) (int, bool) {
	depth := 0
	for cursor := open; cursor < len(text); {
		switch text[cursor] {
		case '(':
			depth++
			cursor++
		case ')':
			depth--
			cursor++
			if depth == 0 {
				return cursor, true
			}
		case '"', '\'':
			cursor = literalEnd(text, cursor)
		default:
			cursor++
		}
	}
	return 0, false
}

// destringizePragma reverses the `_Pragma` destringization the standard
// specifies: `\"` becomes `"` and `\\` becomes `\`; every other byte is literal.
func destringizePragma(value string) string {
	var builder strings.Builder
	for cursor := 0; cursor < len(value); cursor++ {
		if value[cursor] == '\\' && cursor+1 < len(value) && (value[cursor+1] == '"' || value[cursor+1] == '\\') {
			cursor++
		}
		builder.WriteByte(value[cursor])
	}
	return builder.String()
}

// startsAsmStatement reports whether the cursor sits on an `asm`, `__asm`, or
// `__asm__` keyword at a token boundary — the same treatment `_Pragma` and
// `__pragma` receive, and for the same reason. None of the three is a
// `#`-introduced line or a pragma, so the directive grammar never sees them,
// yet `__asm__(".incbin \"/etc/passwd\"");` at file scope in a plain `.c`
// embeds that file's bytes in the object.
func (s *directiveScanner) startsAsmStatement() bool {
	keyword := s.asmKeyword()
	if keyword == "" {
		return false
	}
	if s.index > 0 && identifierByte(s.text[s.index-1], true) {
		return false
	}
	next := s.index + len(keyword)
	return next >= len(s.text) || !identifierByte(s.text[next], true)
}

// asmKeyword returns the assembly keyword at the cursor, longest spelling
// first so `__asm__` is never read as `__asm` followed by content.
func (s *directiveScanner) asmKeyword() string {
	for _, keyword := range []string{"__asm__", "__asm", "asm"} {
		if strings.HasPrefix(s.text[s.index:], keyword) {
			return keyword
		}
	}
	return ""
}

// rejectAsmStatement rejects the target on any inline-assembly construct.
//
// This is the reject-by-default posture applied to the integrated assembler,
// and it replaces the per-spelling classifier earlier rounds carried. The
// assembler is a second file-reading stage inside the same `clang -c`
// invocation whose grammar shares no token with the preprocessor's: `.incbin`
// pastes a file's bytes, `.include` assembles a file's text, `.linker_option`
// makes the linker load an undeclared library, `.secure_log_unique` writes a
// file, and `.macro`/`.irp`/`.irpc` parameter substitution can build any of
// those names out of fragments before the assembler ever looks a directive up.
// Every one of those was verified on the accepted Darwin profile, and none of
// them is reported by `-H`, so an observed-read provider is no backstop either.
//
// Proving what an assembler reads means emulating it, spelling by spelling, and
// rounds 5 through 8 each found one more spelling that the previous round's
// grammar could not see. Portable mode cannot carry that proof, so it does not
// try: the construct itself is unsupported, and the adversarially complete
// acceptance of inline assembly is deferred to the observed-read provider. The
// cost is that a C-family target using any inline assembly is rejected, which
// no admitted SwiftPM source-only shape needs and which fails in the safe
// direction.
//
// Bare `asm` is rejected here rather than left as content even though it is an
// ordinary identifier outside the GNU modes: the disposition depends on a
// language mode this stage cannot bind per file, and the permissive reading is
// the unsound one. `#define K asm` followed by `K(".incbin \"/etc/passwd\"");`
// reads that file on the pinned compiler.
func (s *directiveScanner) rejectAsmStatement() error {
	return failFields(CodeTargetPlatformUnsupported, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "construct": lineSnippet(s.text, s.index)}, "source contains an inline-assembly construct: portable mode admits no assembler stage, whose file reads it cannot confine")
}

// pragmaVerdict is the channel disposition of one pragma body.
type pragmaVerdict struct {
	module   string
	rejected bool
	reason   string
}

// safePragmaHeads is the closed set of pragma spellings portable mode
// positively admits by their first token, whatever operand follows. Every one
// of them alters macro state, struct layout, symbol binding, diagnostics,
// floating-point evaluation, or editor presentation, and none can name a file
// the compiler opens or change where files are resolved from.
//
// It is an allowlist, not a denylist, because that is what makes the channel
// axis terminate: a pragma spelling nobody has enumerated yet — a new vendor
// extension, a spelling from another toolchain, an outright typo — lands
// outside it and is rejected, instead of falling through as content the way a
// deny-list grammar would drop it.
var safePragmaHeads = map[string]bool{
	"once": true, "mark": true, "pack": true, "push_macro": true, "pop_macro": true,
	"unused": true, "weak": true, "message": true, "warning": true,
	"region": true, "endregion": true, "STDC": true, "options": true,
}

// safeClangPragmas and safeGCCPragmas are the second tokens admitted under the
// two vendor namespaces a SwiftPM C-family target actually uses. `clang module`
// is handled separately: its `import` form is a module reference this stage
// resolves, and every other `clang module` spelling reaches module machinery
// that can name a file — `build`/`endbuild` parse an inline module map that may
// declare an absolute out-of-package header, `load` names a module file — so it
// is rejected at that branch rather than dropped here.
//
// Spellings deliberately outside both sets include `comment`, which names a
// library, `include_alias`, which really does substitute an aliased file under
// `-fms-extensions`, and `GCC dependency`, which names a file a conforming
// implementation opens. Under reject-by-default none of the three needs its own
// rule; each is listed here only because an earlier round proved it a channel.
var safeClangPragmas = map[string]bool{
	"diagnostic": true, "attribute": true, "assume_nonnull": true, "system_header": true,
	"arc_cf_code_audited": true, "deprecated": true, "fp": true, "loop": true,
	"unroll": true, "optimize": true, "section": true, "final": true,
	"max_tokens_here": true, "max_tokens_total": true, "restrict_expansion": true,
}

var safeGCCPragmas = map[string]bool{
	"diagnostic": true, "visibility": true, "poison": true, "system_header": true,
	"warning": true, "error": true, "push_options": true, "pop_options": true,
	"optimize": true, "target": true, "unroll": true, "ivdep": true, "novector": true,
}

// classifyPragmaBody answers the channel question for one pragma body: is this
// a spelling portable mode can prove reads no file?
//
// The answer is reject-by-default. A body whose head is in the allowlists above
// is content, `clang module import NAME` is a module reference routed through
// the same confinement `@import` uses, and everything else is a rejection. The
// pragma surface is the widest and least enumerable channel in the language —
// it is open by construction to every vendor — so an allowlist is the only form
// of this rule that stays closed as toolchains grow.
//
// `clang module import NAME` is admitted rather than rejected because it is the
// only module-import spelling available in a plain C translation unit, where
// `@import` is a syntax error, and rejecting it would reject the ordinary
// Objective-C-adjacent shapes this profile supports.
func classifyPragmaBody(body string) pragmaVerdict {
	reject := pragmaVerdict{rejected: true, reason: "pragma is outside the closed set of spellings portable mode can prove names no file"}
	trimmed := strings.TrimLeft(body, " \t\v\f")
	if trimmed == "" {
		return pragmaVerdict{}
	}
	first, rest := splitLeadingIdentifier(trimmed)
	switch first {
	case "clang":
		second, tail := splitLeadingIdentifier(strings.TrimLeft(rest, " \t\v\f"))
		if second != "module" {
			if safeClangPragmas[second] {
				return pragmaVerdict{}
			}
			return reject
		}
		third, nameTail := splitLeadingIdentifier(strings.TrimLeft(tail, " \t\v\f"))
		if third != "import" {
			return pragmaVerdict{rejected: true, reason: "Clang module pragma is a spelling this scanner grammar cannot confine"}
		}
		name, next, ok := readModuleName(nameTail, skipHorizontal(nameTail, 0))
		if !ok || strings.TrimSpace(nameTail[next:]) != "" {
			return pragmaVerdict{rejected: true, reason: "Clang module import pragma is not an exact module name this scanner grammar can resolve"}
		}
		return pragmaVerdict{module: name}
	case "GCC":
		second, _ := splitLeadingIdentifier(strings.TrimLeft(rest, " \t\v\f"))
		if safeGCCPragmas[second] {
			return pragmaVerdict{}
		}
		return reject
	}
	if safePragmaHeads[first] {
		return pragmaVerdict{}
	}
	return reject
}

func splitLeadingIdentifier(text string) (string, string) {
	cursor := 0
	for cursor < len(text) && identifierByte(text[cursor], cursor > 0) {
		cursor++
	}
	return text[:cursor], text[cursor:]
}

// readDirective consumes one preprocessing directive, including every block
// comment inside it, and classifies it. An inclusion directive must resolve to
// an exact literal operand; any other `#`-introduced line must be a directive
// this closed grammar names, or the input is rejected.
func (s *directiveScanner) readDirective() error {
	cursor := s.index + 1
	if s.text[s.index] == '%' {
		cursor = s.index + 2
	}
	body, end := s.readDirectiveBody(cursor)
	s.index = end
	s.atLineStart = true
	name, operand := splitDirectiveName(body)
	reject := func(message string) error {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": strings.TrimSpace("#" + body)}, "%s", message)
	}
	if s.rawString {
		s.rawString = false
		return s.rejectRawString()
	}
	switch {
	case name == "embed":
		return reject("resource-inclusion directive names translation-unit bytes portable mode does not admit")
	case inclusionDirectives[name]:
		spelling, angled, ok := literalIncludeOperand(operand)
		if !ok {
			return reject("inclusion directive operand is not an exact literal this scanner grammar can resolve")
		}
		s.references = append(s.references, IncludeReference{Package: s.pkg, Target: s.target, SourcePackage: s.sourcePkg, Source: s.source, Spelling: spelling, Angled: angled})
		return nil
	case name == "" && strings.TrimSpace(operand) == "":
		return nil
	case name == "pragma":
		// A `#pragma` line's tokens are not macro-expanded: verified, both for
		// the head (`#define CL clang` + `#pragma CL module import SecretKit`
		// imports nothing) and for the module name (`#define DeclaredKit
		// SecretKit` + `#pragma clang module import DeclaredKit` reports module
		// `DeclaredKit` not found). The recorded spelling is therefore the module
		// the compiler resolves, unlike `@import`.
		return s.applyPragma(operand, strings.TrimSpace("#"+body), false)
	case name == "define":
		return s.readMacroDefinition(body, operand)
	case classifiableDirectives[name]:
		return s.scanDirectiveChannels(body)
	}
	return reject("preprocessing directive is not one this scanner grammar can classify")
}

// applyPragma routes one pragma body — the text after `#pragma`, or the
// destringized `_Pragma` operand — through the channel disposition. A pragma
// that names a module is a module import with the same reach and the same
// `-fmodules` precondition as `@import`; a pragma that redirects file
// resolution, or a `clang module` spelling this grammar cannot resolve, is a
// rejection.
// The expanded argument says whether the compiler macro-expands the tokens of
// this pragma body before acting on them. It is false for a `#pragma` line and
// for a destringized `_Pragma` operand, and true for the Microsoft `__pragma`
// operator — all three verified on the accepted Darwin profile against a module
// whose only header is an `#error` marker, because the symmetry assumption is
// the wrong one here.
func (s *directiveScanner) applyPragma(body, snippet string, expanded bool) error {
	verdict := classifyPragmaBody(body)
	if verdict.rejected {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": snippet}, "%s", verdict.reason)
	}
	if verdict.module != "" {
		s.references = append(s.references, IncludeReference{Package: s.pkg, Target: s.target, SourcePackage: s.sourcePkg, Source: s.source, Spelling: verdict.module, ModuleImport: true, ExpandedName: expanded})
	}
	return nil
}

// readMacroDefinition classifies one `#define`, which is translation phase 4's
// only source-visible input and therefore the last layer of the channel proof.
//
// Phases 1-3 guarantee that a channel keyword cannot be reconstituted
// lexically, so the only remaining way to deliver one into a position the token
// scanner reads as ordinary content is the `##` paste operator, which joins two
// tokens into a single new token after the scanner has already read them apart.
// Verified on the accepted Darwin profile: `#define J(a,b) a##b` with
// `J(a,sm)(".incbin \"payload.bin\"")` produces an object byte-identical to the
// direct `__asm__` control, `J(__as,m__)` does the same, and
// `J(_Prag,ma)("clang module import SecretKit")` builds and reads a module the
// target never declared. None of `asm`, `__asm__`, or `_Pragma` appears in
// either the definition or the call.
//
// The disposition splits on where the pasted fragments come from:
//
//   - A paste that joins a parameter to another identifier-shaped token takes
//     at least one fragment from the call site, so no body-local analysis can
//     bound the result. Portable mode rejects the target. This narrows accepted
//     input — parameter pasting is an ordinary C idiom — in the same deliberate
//     way `#embed` was narrowed, and the GNU `, ## __VA_ARGS__` comma-deletion
//     idiom is unaffected because its left operand is a punctuator that cannot
//     contribute identifier characters.
//   - A paste that joins fixed fragments is resolvable: the fragments are in
//     the definition. The `##` and the white space around it are deleted and
//     the resulting token stream is scanned like any other, so
//     `#define A __as##m__` rejects on the `__asm__` it builds.
func (s *directiveScanner) readMacroDefinition(body, operand string) error {
	reject := func(message string) error {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": s.pkg + ":" + s.target, "source": s.source, "directive": strings.TrimSpace("#" + body)}, "%s", message)
	}
	name, rest := splitLeadingIdentifier(operand[skipHorizontal(operand, 0):])
	if atPositionIdentifiers[name] {
		return reject("macro binds an identifier the compiler expands in the `@`-keyword position, where a definition that reaches `import` performs an unconfined module import")
	}
	s.noteMacro(name)
	parameters, _, ok := readMacroParameters(rest)
	if !ok {
		return reject("function-like macro definition has no parameter list this scanner grammar can read")
	}
	references, err := analyzeMacroBody(s.pkg, s.target, s.sourcePkg, s.source, body, parameters)
	if errors.Is(err, errParameterPaste) {
		return reject(err.Error())
	}
	if err != nil {
		return err
	}
	s.references = append(s.references, references...)
	return nil
}

// readMacroParameters reads a function-like macro's parameter list and returns
// the parameter set plus the text that follows the closing parenthesis. A
// definition that is not function-like yields an empty set and its input
// unchanged.
//
// Both macro-binding inputs share this reader for the same reason they share
// analyzeMacroBody: the parameter set is what makes collapseMacroPastes able to
// tell a definition-resolvable paste from one that takes a fragment from the
// call site, and that distinction must not differ between a source `#define`
// and a `.define` build setting spelled `J(a,b)=a##b`.
func readMacroParameters(rest string) (map[string]bool, string, bool) {
	parameters := map[string]bool{}
	if !strings.HasPrefix(rest, "(") {
		return parameters, rest, true
	}
	end, ok := balancedParenthesis(rest, 0)
	if !ok {
		return nil, "", false
	}
	for _, part := range strings.Split(rest[1:end-1], ",") {
		if parameter, _ := splitLeadingIdentifier(strings.TrimLeft(part, " \t\v\f")); parameter != "" {
			parameters[parameter] = true
		}
	}
	parameters["__VA_ARGS__"] = true
	return parameters, rest[end:], true
}

// analyzeMacroBody is the single phase-4 replacement-list analyzer, shared by
// every input that binds a macro the pinned compiler honours.
//
// There are exactly two such inputs for an admitted C-family target: a source
// `#define`, and a SwiftPM `.define` build setting, which reaches the same
// `clang -c` as `-D` without appearing in any admitted file. Round 9 finding M
// established that a replacement list can deliver a channel keyword into a
// position the token scanner reads as ordinary content, and round 13 finding Q
// established that the build-setting spelling of the same body reaches the
// compiler identically: verified on the accepted Darwin profile,
// `-D'A=__asm__'` with `A(".incbin \"payload.bin\"");` produces an object
// byte-identical to the direct `__asm__(…)` control with the named file's bytes
// present, and `-D'A=_Pragma'` with `A("clang module import SecretKit")` builds
// an undeclared module and reads its header.
//
// So the body is analyzed in exactly one place and both routes call it: the
// paste layer first, because `##` and `%:%:` join tokens the scanner would
// otherwise read apart, then the channel scan over the collapsed stream. An
// unresolvable paste comes back as errParameterPaste so each caller can render
// it in its own input's terms; every other verdict is already a formed
// rejection carrying this stage's codes.
func analyzeMacroBody(pkg, target, sourcePkg, source, body string, parameters map[string]bool) ([]IncludeReference, error) {
	collapsed, err := collapseMacroPastes(body, parameters)
	if err != nil {
		return nil, err
	}
	scanner := &directiveScanner{pkg: pkg, target: target, sourcePkg: sourcePkg, source: source}
	if err := scanner.scanDirectiveChannels(collapsed); err != nil {
		return nil, err
	}
	return scanner.references, nil
}

// errParameterPaste is the body-local verdict collapseMacroPastes returns when
// a paste cannot be resolved from the definition alone.
var errParameterPaste = fmt.Errorf("macro pastes a call-site fragment into a new token portable mode cannot bound")

// collapseMacroPastes performs the `##` pastes a definition resolves by itself
// and reports the ones it does not. The result is the same token stream with
// every resolvable paste already joined, so an ordinary scan of it sees the
// tokens the compiler will see after phase 4.
func collapseMacroPastes(body string, parameters map[string]bool) (string, error) {
	if !strings.Contains(body, "##") && !strings.Contains(body, "%:%:") {
		return body, nil
	}
	var out []byte
	for cursor := 0; cursor < len(body); {
		switch {
		case body[cursor] == '"' || body[cursor] == '\'':
			end := literalEnd(body, cursor)
			out = append(out, body[cursor:end]...)
			cursor = end
		case macroPasteWidth(body, cursor) > 0:
			out = []byte(strings.TrimRight(string(out), " \t\v\f"))
			left := trailingIdentifier(string(out))
			next := skipHorizontal(body, cursor+macroPasteWidth(body, cursor))
			right, _ := splitLeadingIdentifier(body[next:])
			if (parameters[left] && right != "") || (parameters[right] && left != "") {
				return "", errParameterPaste
			}
			cursor = next
		default:
			out = append(out, body[cursor])
			cursor++
		}
	}
	return string(out), nil
}

// macroPasteWidth returns the source width of the token-paste operator at
// cursor, or zero when no paste starts there.
//
// `##` and `%:%:` are two spellings of one preprocessing token, so the paste
// layer has to see both or the digraph walks straight past it. `readDirective`
// already reads `%:` as the digraph for `#`, which is why the omission was a
// hole rather than an unmodeled feature: verified on the accepted Darwin
// profile, `#define A __as%:%:m__` and `#define J(a,b) a%:%:b` each produce an
// object byte-identical to the direct `__asm__(".incbin …")` control with the
// named file's bytes present, and `#define A _Prag%:%:ma` builds an undeclared
// module and reads its header. Digraphs, unlike trigraphs, are unconditional in
// every mode this profile admits, so there is nothing to bind per file.
func macroPasteWidth(body string, cursor int) int {
	switch {
	case strings.HasPrefix(body[cursor:], "##"):
		return 2
	case strings.HasPrefix(body[cursor:], "%:%:"):
		return 4
	}
	return 0
}

// trailingIdentifier returns the identifier token that ends text, or the empty
// string when text ends in a punctuator, a literal, or nothing.
func trailingIdentifier(text string) string {
	start := len(text)
	for start > 0 && identifierByte(text[start-1], true) {
		start--
	}
	if start == len(text) || !identifierByte(text[start], false) {
		return ""
	}
	return text[start:]
}

// scanDirectiveChannels re-scans one classified directive's body for the
// token-level channels that survive macro expansion. `#define IMP
// _Pragma("clang module import SecretKit")` followed by a bare `IMP` really does
// build and read that module on the pinned Apple Clang, and so does
// `#define DO(x) _Pragma(#x)` invoked as `DO(clang module import SecretKit)`.
// The expansion site is ordinary content no grammar short of a preprocessor can
// recognize, so the definition is where the channel has to be classified.
func (s *directiveScanner) scanDirectiveChannels(body string) error {
	nested := &directiveScanner{pkg: s.pkg, target: s.target, sourcePkg: s.sourcePkg, source: s.source, text: body}
	for nested.index < len(nested.text) {
		character := nested.text[nested.index]
		switch {
		case rawStringPrefixAt(nested.text, nested.index):
			return nested.rejectRawString()
		case character == '"' || character == '\'':
			nested.index = literalEnd(nested.text, nested.index)
		case nested.startsModuleImport():
			if err := nested.readModuleImport(); err != nil {
				return err
			}
		case character == '@':
			if err := nested.readAtToken(); err != nil {
				return err
			}
		case nested.startsPragmaOperator():
			if err := nested.readPragmaOperator(); err != nil {
				return err
			}
		case nested.startsAsmStatement():
			if err := nested.rejectAsmStatement(); err != nil {
				return err
			}
		default:
			nested.index++
		}
	}
	s.references = append(s.references, nested.references...)
	return nil
}

// readDirectiveBody returns the directive text with every comment replaced by
// one space, and the index just past the newline that ends it. A block comment
// inside a directive does not end it — the pinned Apple Clang continues the
// directive across the comment's newlines.
func (s *directiveScanner) readDirectiveBody(cursor int) (string, int) {
	var builder strings.Builder
	for cursor < len(s.text) {
		switch {
		case s.text[cursor] == '\n':
			return builder.String(), cursor + 1
		case strings.HasPrefix(s.text[cursor:], "/*"):
			builder.WriteByte(' ')
			end := strings.Index(s.text[cursor+2:], "*/")
			if end < 0 {
				return builder.String(), len(s.text)
			}
			cursor += 2 + end + 2
		case strings.HasPrefix(s.text[cursor:], "//"):
			builder.WriteByte(' ')
			next := strings.IndexByte(s.text[cursor:], '\n')
			if next < 0 {
				return builder.String(), len(s.text)
			}
			return builder.String(), cursor + next + 1
		case s.text[cursor] == '"':
			if rawStringPrefixAt(s.text, cursor) {
				s.rawString = true
			}
			end := literalEnd(s.text, cursor)
			builder.WriteString(s.text[cursor:end])
			cursor = end
		default:
			builder.WriteByte(s.text[cursor])
			cursor++
		}
	}
	return builder.String(), cursor
}

// splitDirectiveName separates the directive name from its operand. The full
// white-space set may precede the name: a form feed is white space, so
// `\f#include </etc/passwd>` is a directive the compiler honours.
func splitDirectiveName(body string) (string, string) {
	cursor := skipHorizontal(body, 0)
	start := cursor
	for cursor < len(body) && identifierByte(body[cursor], cursor > start) {
		cursor++
	}
	return body[start:cursor], body[cursor:]
}

func readModuleName(text string, cursor int) (string, int, bool) {
	start := cursor
	for cursor < len(text) {
		component := cursor
		for cursor < len(text) && identifierByte(text[cursor], cursor > component) {
			cursor++
		}
		if cursor == component {
			return "", cursor, false
		}
		if cursor < len(text) && text[cursor] == '.' {
			cursor++
			continue
		}
		break
	}
	if cursor == start {
		return "", cursor, false
	}
	return text[start:cursor], cursor, true
}

// literalEnd returns the index just past one string or character literal. A
// literal never crosses a newline; a delimiter with no partner on its logical
// line is therefore not a literal at all and is consumed as one ordinary byte,
// so the rest of that line is still scanned. Consuming to the end of the line
// instead let a C++ digit separator hide a module import — `int x = 1'0;
// @import Secret;` does import on the pinned compiler, verified with
// `-fmodules -fcxx-modules -std=c++14`.
func literalEnd(text string, cursor int) int {
	quote := text[cursor]
	for scan := cursor + 1; scan < len(text) && text[scan] != '\n'; {
		if text[scan] == '\\' {
			scan += 2
			continue
		}
		if text[scan] == quote {
			return scan + 1
		}
		scan++
	}
	return cursor + 1
}

func lineSnippet(text string, start int) string {
	end := strings.IndexByte(text[start:], '\n')
	if end < 0 {
		end = len(text) - start
	}
	return strings.TrimSpace(text[start : start+end])
}

func skipHorizontal(text string, cursor int) int {
	for cursor < len(text) && horizontalSpace(text[cursor]) {
		cursor++
	}
	return cursor
}

func horizontalSpace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\v' || character == '\f'
}

func identifierByte(character byte, allowDigit bool) bool {
	switch {
	case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z', character == '_':
		return true
	case allowDigit && character >= '0' && character <= '9':
		return true
	default:
		return false
	}
}

// literalIncludeOperand resolves one inclusion directive operand to an exact
// quoted or angled literal. Macro-computed forms such as `#include MACRO`, an
// empty spelling, an unterminated delimiter, or any trailing token other than a
// comment resolve to nothing this grammar can prove and are reported as
// unresolved rather than dropped.
func literalIncludeOperand(operand string) (string, bool, bool) {
	rest := strings.TrimLeft(operand, " \t")
	if rest == "" {
		return "", false, false
	}
	var closing byte
	angled := false
	switch rest[0] {
	case '"':
		closing = '"'
	case '<':
		closing, angled = '>', true
	default:
		return "", false, false
	}
	end := strings.IndexByte(rest[1:], closing)
	if end < 0 {
		return "", false, false
	}
	spelling := rest[1 : 1+end]
	if spelling == "" {
		return "", false, false
	}
	trailing := strings.TrimLeft(rest[2+end:], " \t")
	if trailing != "" && !strings.HasPrefix(trailing, "//") && !strings.HasPrefix(trailing, "/*") {
		return "", false, false
	}
	return spelling, angled, true
}
