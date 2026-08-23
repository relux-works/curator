package artifactpolicy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"text/scanner"
	"unicode"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

func (inspector *inspector) detectText(item blob, virtualPath, descriptorPath string) detection {
	declaration, declared := inspector.descriptor.DeclaredText[descriptorPath]
	if !declared {
		declaration, declared = inferTextDeclaration(inspector.descriptor.ProfileID, virtualPath, item)
	}
	if !declared {
		valid, reason := validateTextStream(item.reader(), GrammarPlain)
		observation := Observation{
			DetectorID: "source-text-v1", Result: "NO_PROFILE_MATCH",
			Facts: []Fact{{Key: "reason", Value: reason}},
		}
		if !valid {
			observation.Result = "NO_MATCH"
		}
		return detection{
			class: ClassOpaqueUnknown, detectorID: "source-text-v1",
			diagnostic: CodeOpaqueDependency, reason: "no_declared_source_grammar",
			observations: []Observation{observation},
		}
	}
	if !profileAllowsGrammar(inspector.descriptor.ProfileID, declaration.Grammar) ||
		!textClass(declaration.Class) {
		return detection{
			class: ClassOpaqueUnknown, detectorID: "source-text-v1",
			diagnostic: CodeOpaqueDependency, reason: "grammar_not_allowed_by_profile",
			observations: []Observation{{
				DetectorID: "source-text-v1", Result: "ERROR",
				Facts: facts(map[string]any{"grammar": declaration.Grammar, "profile": inspector.descriptor.ProfileID}),
			}},
		}
	}
	if declaration.Class == ClassSourceGeneratedText &&
		inspector.descriptor.RequireGeneratedLineage && declaration.GeneratedLineage == "" {
		return detection{
			class: declaration.Class, detectorID: "source-text-v1",
			diagnostic: CodeGeneratedInputUndeclared, reason: "missing_generated_input_lineage",
			observations: []Observation{{
				DetectorID: "source-text-v1", Result: "MATCH",
				Facts: facts(map[string]any{
					"generated_lineage": "missing",
					"grammar":           declaration.Grammar,
					"selected_class":    declaration.Class,
				}),
			}},
		}
	}
	valid, reason := validateDeclaredText(item, declaration.Grammar)
	if !valid {
		return detection{
			class: ClassOpaqueUnknown, detectorID: "source-text-v1",
			diagnostic: CodeOpaqueDependency, reason: reason,
			observations: []Observation{{
				DetectorID: "source-text-v1", Result: "ERROR",
				Facts: facts(map[string]any{"grammar": declaration.Grammar, "reason": reason}),
			}},
		}
	}
	return detection{
		class: declaration.Class, detectorID: "source-text-v1",
		observations: []Observation{{
			DetectorID: "source-text-v1", Result: "MATCH",
			Facts: facts(map[string]any{
				"grammar":           declaration.Grammar,
				"generated_lineage": declaration.GeneratedLineage,
				"selected_class":    declaration.Class,
			}),
		}},
	}
}

func textClass(class ArtifactClass) bool {
	return class == ClassSourceAuthoredText || class == ClassSourceGeneratedText || class == ClassTextMetadata
}

func inferTextDeclaration(profile ProfileID, virtualPath string, item blob) (TextDeclaration, bool) {
	name := strings.ToLower(path.Base(leafPath(virtualPath)))
	if name == "" {
		return TextDeclaration{}, false
	}
	if strings.HasSuffix(name, ".swiftinterface") {
		return TextDeclaration{Grammar: GrammarSwiftInterface, Class: ClassSourceGeneratedText}, true
	}
	if strings.HasSuffix(name, ".min.js") || strings.HasSuffix(name, ".min.mjs") || strings.HasSuffix(name, ".min.cjs") {
		return TextDeclaration{Grammar: GrammarJavaScript, Class: ClassSourceGeneratedText}, true
	}
	if strings.HasSuffix(name, ".map") {
		return TextDeclaration{Grammar: GrammarSourceMap, Class: ClassTextMetadata}, true
	}
	if declaration, ok := metadataDeclaration(name); ok {
		return declaration, true
	}
	extension := strings.ToLower(path.Ext(name))
	declaration, ok := sourceExtensions[extension]
	if ok && profileAllowsGrammar(profile, declaration.Grammar) {
		return declaration, true
	}
	if extension == "" {
		prefix, err := item.prefix(256)
		if err == nil {
			firstLine := string(prefix)
			if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
				firstLine = firstLine[:newline]
			}
			switch {
			case strings.HasPrefix(firstLine, "#!") && strings.Contains(firstLine, "python"):
				declaration = TextDeclaration{Grammar: GrammarPython, Class: ClassSourceAuthoredText}
			case strings.HasPrefix(firstLine, "#!") && (strings.Contains(firstLine, "node") || strings.Contains(firstLine, "deno")):
				declaration = TextDeclaration{Grammar: GrammarJavaScript, Class: ClassSourceAuthoredText}
			case strings.HasPrefix(firstLine, "#!"):
				declaration = TextDeclaration{Grammar: GrammarShell, Class: ClassSourceAuthoredText}
			default:
				return TextDeclaration{}, false
			}
			if profileAllowsGrammar(profile, declaration.Grammar) {
				return declaration, true
			}
		}
	}
	return TextDeclaration{}, false
}

var sourceExtensions = map[string]TextDeclaration{
	".go":        {Grammar: GrammarGo, Class: ClassSourceAuthoredText},
	".rs":        {Grammar: GrammarRust, Class: ClassSourceAuthoredText},
	".swift":     {Grammar: GrammarSwift, Class: ClassSourceAuthoredText},
	".c":         {Grammar: GrammarC, Class: ClassSourceAuthoredText},
	".h":         {Grammar: GrammarC, Class: ClassSourceAuthoredText},
	".cc":        {Grammar: GrammarCXX, Class: ClassSourceAuthoredText},
	".cpp":       {Grammar: GrammarCXX, Class: ClassSourceAuthoredText},
	".cxx":       {Grammar: GrammarCXX, Class: ClassSourceAuthoredText},
	".hpp":       {Grammar: GrammarCXX, Class: ClassSourceAuthoredText},
	".hh":        {Grammar: GrammarCXX, Class: ClassSourceAuthoredText},
	".m":         {Grammar: GrammarObjectiveC, Class: ClassSourceAuthoredText},
	".mm":        {Grammar: GrammarObjectiveC, Class: ClassSourceAuthoredText},
	".js":        {Grammar: GrammarJavaScript, Class: ClassSourceAuthoredText},
	".mjs":       {Grammar: GrammarJavaScript, Class: ClassSourceAuthoredText},
	".cjs":       {Grammar: GrammarJavaScript, Class: ClassSourceAuthoredText},
	".jsx":       {Grammar: GrammarJavaScript, Class: ClassSourceAuthoredText},
	".ts":        {Grammar: GrammarTypeScript, Class: ClassSourceAuthoredText},
	".mts":       {Grammar: GrammarTypeScript, Class: ClassSourceAuthoredText},
	".cts":       {Grammar: GrammarTypeScript, Class: ClassSourceAuthoredText},
	".tsx":       {Grammar: GrammarTypeScript, Class: ClassSourceAuthoredText},
	".py":        {Grammar: GrammarPython, Class: ClassSourceAuthoredText},
	".sh":        {Grammar: GrammarShell, Class: ClassSourceAuthoredText},
	".bash":      {Grammar: GrammarShell, Class: ClassSourceAuthoredText},
	".zsh":       {Grammar: GrammarShell, Class: ClassSourceAuthoredText},
	".s":         {Grammar: GrammarAssembly, Class: ClassSourceAuthoredText},
	".asm":       {Grammar: GrammarAssembly, Class: ClassSourceAuthoredText},
	".modulemap": {Grammar: GrammarModuleMap, Class: ClassSourceAuthoredText},
}

func metadataDeclaration(name string) (TextDeclaration, bool) {
	base := name
	if dot := strings.IndexByte(base, '.'); dot >= 0 {
		base = base[:dot]
	}
	switch base {
	case "readme", "license", "licence", "notice", "copying", "authors", "contributors", "changelog", "changes", "security",
		"wheel", "metadata", "record", "entry_points", "top_level":
		return TextDeclaration{Grammar: GrammarMarkdown, Class: ClassTextMetadata}, true
	}
	switch path.Ext(name) {
	case ".md", ".markdown", ".rst", ".txt":
		return TextDeclaration{Grammar: GrammarMarkdown, Class: ClassTextMetadata}, true
	case ".json":
		return TextDeclaration{Grammar: GrammarJSON, Class: ClassTextMetadata}, true
	case ".toml":
		return TextDeclaration{Grammar: GrammarTOML, Class: ClassTextMetadata}, true
	case ".yaml", ".yml":
		return TextDeclaration{Grammar: GrammarYAML, Class: ClassTextMetadata}, true
	case ".lock", ".sum", ".patch", ".diff":
		return TextDeclaration{Grammar: GrammarPlain, Class: ClassTextMetadata}, true
	}
	switch name {
	case "go.mod", "go.sum", "cargo.lock", "package-lock.json", "npm-shrinkwrap.json", "binding.gyp",
		"yarn.lock", "pnpm-lock.yaml", "package.resolved", "manifest", "makefile", ".npmrc", ".yarnrc", ".gitmodules", ".gitignore", ".gitattributes":
		return TextDeclaration{Grammar: GrammarPlain, Class: ClassTextMetadata}, true
	default:
		return TextDeclaration{}, false
	}
}

func profileAllowsGrammar(profile ProfileID, grammar GrammarID) bool {
	if _, ok := grammarIDs[grammar]; !ok {
		return false
	}
	if metadataGrammar(grammar) {
		return true
	}
	switch profile {
	case ProfileCommonV1:
		return true
	case ProfileGoV1:
		return grammar == GrammarGo || grammar == GrammarAssembly
	case ProfileRustV1:
		return grammar == GrammarRust || grammar == GrammarC || grammar == GrammarAssembly
	case ProfileNodeV1:
		return grammar == GrammarJavaScript || grammar == GrammarTypeScript || grammar == GrammarShell
	case ProfilePythonSourceV1:
		return grammar == GrammarPython || grammar == GrammarShell
	case ProfileSwiftPMV1:
		return grammar == GrammarSwift || grammar == GrammarSwiftInterface || grammar == GrammarC ||
			grammar == GrammarCXX || grammar == GrammarObjectiveC || grammar == GrammarAssembly ||
			grammar == GrammarModuleMap
	default:
		return false
	}
}

func metadataGrammar(grammar GrammarID) bool {
	switch grammar {
	case GrammarPlain, GrammarMarkdown, GrammarJSON, GrammarTOML, GrammarYAML, GrammarSourceMap:
		return true
	default:
		return false
	}
}

func validateDeclaredText(item blob, grammar GrammarID) (bool, string) {
	switch grammar {
	case GrammarJSON, GrammarSourceMap:
		if err := validateJSONStream(item.reader()); err != nil {
			return false, "invalid_json:" + err.Error()
		}
		return true, ""
	case GrammarTOML:
		var value map[string]any
		if _, err := toml.NewDecoder(item.reader()).Decode(&value); err != nil {
			return false, "invalid_toml:" + err.Error()
		}
		return true, ""
	case GrammarGo:
		if err := validateGoLexemes(item.reader()); err != nil {
			return false, "invalid_go_lexeme:" + err.Error()
		}
		return validateTextStream(item.reader(), grammar)
	default:
		return validateTextStream(item.reader(), grammar)
	}
}

func validateGoLexemes(reader io.Reader) error {
	var lexer scanner.Scanner
	lexer.Init(reader)
	lexer.Mode = scanner.GoTokens
	firstError := ""
	lexer.Error = func(_ *scanner.Scanner, message string) {
		if firstError == "" {
			firstError = message
		}
	}
	for tokenValue := lexer.Scan(); tokenValue != scanner.EOF; tokenValue = lexer.Scan() {
		if tokenValue >= 0 && !strings.ContainsRune("+-*/%&|^<>=!()[]{}.,;:~", tokenValue) && firstError == "" {
			firstError = fmt.Sprintf("invalid Go character %q", tokenValue)
		}
	}
	if firstError != "" {
		return fmt.Errorf("%s", firstError)
	}
	return nil
}

func validateTextStream(reader io.Reader, grammar GrammarID) (bool, string) {
	buffered := bufio.NewReader(reader)
	state := textLexer{grammar: grammar}
	for {
		character, size, err := buffered.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, "read_error"
		}
		if character == utf8.RuneError && size == 1 {
			return false, "invalid_utf8"
		}
		if character == 0 || (unicode.IsControl(character) && character != '\n' && character != '\r' && character != '\t' && character != '\f') {
			return false, "forbidden_control_character"
		}
		if !state.consume(character) {
			return false, state.reason
		}
	}
	if !state.finish() {
		return false, state.reason
	}
	return true, ""
}

type lexerMode int

const (
	lexerNormal lexerMode = iota
	lexerSingleQuote
	lexerDoubleQuote
	lexerBacktick
	lexerLineComment
	lexerBlockComment
)

type textLexer struct {
	grammar  GrammarID
	mode     lexerMode
	previous rune
	escaped  bool
	stack    []rune
	reason   string
}

func (lexer *textLexer) consume(character rune) bool {
	switch lexer.mode {
	case lexerLineComment:
		if character == '\n' {
			lexer.mode = lexerNormal
		}
		lexer.previous = character
		return true
	case lexerBlockComment:
		if lexer.previous == '*' && character == '/' {
			lexer.mode = lexerNormal
			lexer.previous = 0
			return true
		}
		lexer.previous = character
		return true
	case lexerSingleQuote, lexerDoubleQuote, lexerBacktick:
		if lexer.escaped {
			lexer.escaped = false
			lexer.previous = character
			return true
		}
		if character == '\\' {
			lexer.escaped = true
			lexer.previous = character
			return true
		}
		terminator := rune('\'')
		switch lexer.mode {
		case lexerDoubleQuote:
			terminator = '"'
		case lexerBacktick:
			terminator = '`'
		}
		if character == terminator {
			lexer.mode = lexerNormal
		}
		lexer.previous = character
		return true
	}

	if lexer.previous == '/' && character == '/' && slashComments(lexer.grammar) {
		lexer.mode = lexerLineComment
		lexer.previous = character
		return true
	}
	if lexer.previous == '/' && character == '*' && slashComments(lexer.grammar) {
		lexer.mode = lexerBlockComment
		lexer.previous = character
		return true
	}
	if character == '#' && hashComments(lexer.grammar) {
		lexer.mode = lexerLineComment
		lexer.previous = character
		return true
	}
	switch character {
	case '\'':
		lexer.mode = lexerSingleQuote
	case '"':
		lexer.mode = lexerDoubleQuote
	case '`':
		if lexer.grammar == GrammarJavaScript || lexer.grammar == GrammarTypeScript || lexer.grammar == GrammarGo || lexer.grammar == GrammarShell {
			lexer.mode = lexerBacktick
		}
	case '(', '[', '{':
		if structuralGrammar(lexer.grammar) {
			lexer.stack = append(lexer.stack, character)
		}
	case ')', ']', '}':
		if structuralGrammar(lexer.grammar) {
			if len(lexer.stack) == 0 || !matchingDelimiter(lexer.stack[len(lexer.stack)-1], character) {
				lexer.reason = "unbalanced_delimiter"
				return false
			}
			lexer.stack = lexer.stack[:len(lexer.stack)-1]
		}
	}
	lexer.previous = character
	return true
}

func (lexer *textLexer) finish() bool {
	if lexer.mode == lexerBlockComment || lexer.mode == lexerSingleQuote || lexer.mode == lexerDoubleQuote || lexer.mode == lexerBacktick || lexer.escaped {
		lexer.reason = "unterminated_lexical_construct"
		return false
	}
	if len(lexer.stack) != 0 {
		lexer.reason = "unbalanced_delimiter"
		return false
	}
	return true
}

func slashComments(grammar GrammarID) bool {
	switch grammar {
	case GrammarGo, GrammarRust, GrammarSwift, GrammarC, GrammarCXX,
		GrammarObjectiveC, GrammarJavaScript, GrammarTypeScript, GrammarModuleMap:
		return true
	default:
		return false
	}
}

func hashComments(grammar GrammarID) bool {
	return grammar == GrammarPython || grammar == GrammarShell || grammar == GrammarYAML || grammar == GrammarAssembly
}

func structuralGrammar(grammar GrammarID) bool {
	switch grammar {
	case GrammarGo, GrammarRust, GrammarSwift, GrammarC, GrammarCXX,
		GrammarObjectiveC, GrammarJavaScript, GrammarTypeScript, GrammarModuleMap:
		return true
	default:
		return false
	}
}

func matchingDelimiter(open, closing rune) bool {
	return (open == '(' && closing == ')') || (open == '[' && closing == ']') || (open == '{' && closing == '}')
}

func validateJSONStream(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := consumeJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON data")
		}
		return err
	}
	return nil
}

func consumeJSONValue(decoder *json.Decoder) error {
	tokenValue, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := tokenValue.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, tokenErr := decoder.Token()
			if tokenErr != nil {
				return tokenErr
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("malformed JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("malformed JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	return nil
}
