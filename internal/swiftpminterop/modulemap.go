package swiftpminterop

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// ModuleMapGrammarID is the versioned Clang module-map grammar this stage
// parses. SwiftPM's own source enumeration omits module maps and the headers
// they name, so parsing is mandatory rather than advisory.
const ModuleMapGrammarID = "clang-modulemap-lexer-v1"

// ReferenceKind is one closed module-map file or module reference.
type ReferenceKind string

// The closed module-map reference kinds.
const (
	// ReferenceHeader is a plain, textual, private, or excluded header.
	ReferenceHeader ReferenceKind = "header"
	// ReferenceUmbrellaHeader is an umbrella header declaration.
	ReferenceUmbrellaHeader ReferenceKind = "umbrella_header"
	// ReferenceUmbrellaDirectory is an umbrella directory declaration.
	ReferenceUmbrellaDirectory ReferenceKind = "umbrella_directory"
	// ReferenceExternModule names another module map file.
	ReferenceExternModule ReferenceKind = "extern_module"
)

// Reference is one resolved-by-path module-map reference.
type Reference struct {
	Kind   ReferenceKind
	Module string
	Path   string
	Line   int
}

// Link is one declared linker edge inside a module map.
type Link struct {
	Module    string
	Name      string
	Framework bool
	Line      int
}

// ModuleMap is the parsed, order-stable projection of one module map file.
type ModuleMap struct {
	Modules    []string
	References []Reference
	Links      []Link
	Requires   []string
	Frameworks []string
}

type modulemapToken struct {
	value  string
	line   int
	quoted bool
}

// ParseModuleMap lexes and parses a Clang module map. Every recognised file
// and module reference is retained with its owning module so containment can
// be proved independently of Clang. Any construct this grammar cannot resolve
// exactly is a rejection, never a silent skip.
func ParseModuleMap(logicalPath string, payload []byte) (ModuleMap, error) {
	tokens, err := lexModuleMap(logicalPath, string(payload))
	if err != nil {
		return ModuleMap{}, err
	}
	parser := &modulemapParser{tokens: tokens, path: logicalPath}
	result, err := parser.parse()
	if err != nil {
		return ModuleMap{}, err
	}
	sort.Strings(result.Modules)
	sort.Strings(result.Requires)
	sort.Strings(result.Frameworks)
	result.Modules = uniqueStrings(result.Modules)
	result.Requires = uniqueStrings(result.Requires)
	result.Frameworks = uniqueStrings(result.Frameworks)
	return result, nil
}

func lexModuleMap(logicalPath, source string) ([]modulemapToken, error) {
	tokens := []modulemapToken{}
	line := 1
	for offset := 0; offset < len(source); {
		character := source[offset]
		switch {
		case character == '\n':
			line++
			offset++
		case character == ' ' || character == '\t' || character == '\r':
			offset++
		case strings.HasPrefix(source[offset:], "//"):
			for offset < len(source) && source[offset] != '\n' {
				offset++
			}
		case strings.HasPrefix(source[offset:], "/*"):
			end := strings.Index(source[offset+2:], "*/")
			if end < 0 {
				return nil, modulemapSyntax(logicalPath, line, "unterminated block comment")
			}
			line += strings.Count(source[offset:offset+2+end+2], "\n")
			offset += 2 + end + 2
		case character == '"':
			end := strings.IndexByte(source[offset+1:], '"')
			if end < 0 {
				return nil, modulemapSyntax(logicalPath, line, "unterminated string literal")
			}
			literal := source[offset+1 : offset+1+end]
			if strings.ContainsAny(literal, "\n\x00") {
				return nil, modulemapSyntax(logicalPath, line, "string literal contains a control character")
			}
			tokens = append(tokens, modulemapToken{value: literal, line: line, quoted: true})
			offset += end + 2
		case character == '{' || character == '}' || character == '[' || character == ']' || character == ',' || character == '*' || character == '.' || character == '!':
			tokens = append(tokens, modulemapToken{value: string(character), line: line})
			offset++
		case isIdentifierByte(character):
			end := offset
			for end < len(source) && isIdentifierByte(source[end]) {
				end++
			}
			tokens = append(tokens, modulemapToken{value: source[offset:end], line: line})
			offset = end
		default:
			return nil, modulemapSyntax(logicalPath, line, fmt.Sprintf("unsupported character %q", string(character)))
		}
	}
	return tokens, nil
}

func isIdentifierByte(value byte) bool {
	return value == '_' || value == '+' || value == '-' ||
		unicode.IsLetter(rune(value)) || unicode.IsDigit(rune(value))
}

type modulemapParser struct {
	tokens []modulemapToken
	path   string
	index  int
}

func (parser *modulemapParser) peek() (modulemapToken, bool) {
	if parser.index >= len(parser.tokens) {
		return modulemapToken{}, false
	}
	return parser.tokens[parser.index], true
}

func (parser *modulemapParser) next() (modulemapToken, bool) {
	token, ok := parser.peek()
	if ok {
		parser.index++
	}
	return token, ok
}

func (parser *modulemapParser) expect(value string) (modulemapToken, error) {
	token, ok := parser.next()
	if !ok {
		return modulemapToken{}, modulemapSyntax(parser.path, parser.lastLine(), "expected "+value+" before end of file")
	}
	if token.quoted || token.value != value {
		return modulemapToken{}, modulemapSyntax(parser.path, token.line, "expected "+value)
	}
	return token, nil
}

func (parser *modulemapParser) expectString() (modulemapToken, error) {
	token, ok := parser.next()
	if !ok || !token.quoted {
		return modulemapToken{}, modulemapSyntax(parser.path, parser.lastLine(), "expected a quoted path literal")
	}
	return token, nil
}

func (parser *modulemapParser) lastLine() int {
	if parser.index > 0 && parser.index <= len(parser.tokens) {
		return parser.tokens[parser.index-1].line
	}
	return 1
}

func (parser *modulemapParser) parse() (ModuleMap, error) {
	result := ModuleMap{Modules: []string{}, References: []Reference{}, Links: []Link{}, Requires: []string{}, Frameworks: []string{}}
	for {
		token, ok := parser.peek()
		if !ok {
			return result, nil
		}
		if token.quoted {
			return ModuleMap{}, modulemapSyntax(parser.path, token.line, "unexpected top-level string literal")
		}
		switch token.value {
		case "explicit", "framework", "extern", "module":
			if err := parser.parseModuleDeclaration("", &result); err != nil {
				return ModuleMap{}, err
			}
		default:
			return ModuleMap{}, modulemapSyntax(parser.path, token.line, "unsupported top-level declaration "+token.value)
		}
	}
}

func (parser *modulemapParser) parseModuleDeclaration(parent string, result *ModuleMap) error {
	framework := false
	external := false
	for {
		token, ok := parser.peek()
		if !ok {
			return modulemapSyntax(parser.path, parser.lastLine(), "truncated module declaration")
		}
		if token.quoted {
			return modulemapSyntax(parser.path, token.line, "unexpected string literal in module declaration")
		}
		switch token.value {
		case "explicit":
			parser.index++
		case "framework":
			framework = true
			parser.index++
		case "extern":
			external = true
			parser.index++
		case "module":
			parser.index++
			name, err := parser.parseModuleName()
			if err != nil {
				return err
			}
			if err = parser.skipModuleAttributes(); err != nil {
				return err
			}
			qualified := name
			if parent != "" {
				qualified = parent + "." + name
			}
			if external {
				literal, stringErr := parser.expectString()
				if stringErr != nil {
					return stringErr
				}
				result.Modules = append(result.Modules, qualified)
				result.References = append(result.References, Reference{Kind: ReferenceExternModule, Module: qualified, Path: literal.value, Line: literal.line})
				return nil
			}
			result.Modules = append(result.Modules, qualified)
			if framework {
				result.Frameworks = append(result.Frameworks, qualified)
			}
			return parser.parseModuleBody(qualified, result)
		default:
			return modulemapSyntax(parser.path, token.line, "unsupported module qualifier "+token.value)
		}
	}
}

// skipModuleAttributes consumes the optional bracketed attribute list, for
// example the [system] and [extern_c] markers on a system module.
func (parser *modulemapParser) skipModuleAttributes() error {
	for {
		token, ok := parser.peek()
		if !ok || token.quoted || token.value != "[" {
			return nil
		}
		parser.index++
		for {
			inner, present := parser.next()
			if !present {
				return modulemapSyntax(parser.path, parser.lastLine(), "unterminated module attribute list")
			}
			if !inner.quoted && inner.value == "]" {
				break
			}
			if inner.quoted {
				return modulemapSyntax(parser.path, inner.line, "unsupported module attribute literal")
			}
		}
	}
}

func (parser *modulemapParser) parseModuleName() (string, error) {
	token, ok := parser.next()
	if !ok || token.quoted {
		return "", modulemapSyntax(parser.path, parser.lastLine(), "expected a module identifier")
	}
	name := token.value
	for {
		dot, present := parser.peek()
		if !present || dot.quoted || dot.value != "." {
			return name, nil
		}
		parser.index++
		part, present := parser.next()
		if !present || part.quoted {
			return "", modulemapSyntax(parser.path, parser.lastLine(), "expected a qualified module identifier")
		}
		name += "." + part.value
	}
}

func (parser *modulemapParser) parseModuleBody(name string, result *ModuleMap) error {
	if _, err := parser.expect("{"); err != nil {
		return err
	}
	for {
		token, ok := parser.next()
		if !ok {
			return modulemapSyntax(parser.path, parser.lastLine(), "unterminated module body")
		}
		if token.quoted {
			return modulemapSyntax(parser.path, token.line, "unexpected string literal in module body")
		}
		switch token.value {
		case "}":
			return nil
		case "header", "textual", "private", "exclude":
			if err := parser.parseHeaderDeclaration(name, token, result); err != nil {
				return err
			}
		case "umbrella":
			if err := parser.parseUmbrella(name, result); err != nil {
				return err
			}
		case "link":
			if err := parser.parseLink(name, result); err != nil {
				return err
			}
		case "export", "export_as", "use", "config_macros":
			if err := parser.skipIdentifierList(); err != nil {
				return err
			}
		case "requires":
			if err := parser.parseRequires(result); err != nil {
				return err
			}
		case "explicit", "framework", "extern", "module":
			parser.index--
			if err := parser.parseModuleDeclaration(name, result); err != nil {
				return err
			}
		default:
			return modulemapSyntax(parser.path, token.line, "unsupported module member "+token.value)
		}
	}
}

func (parser *modulemapParser) parseHeaderDeclaration(name string, first modulemapToken, result *ModuleMap) error {
	if first.value != "header" {
		token, ok := parser.next()
		if !ok || token.quoted {
			return modulemapSyntax(parser.path, parser.lastLine(), "expected header after "+first.value)
		}
		if token.value == "textual" {
			token, ok = parser.next()
			if !ok || token.quoted {
				return modulemapSyntax(parser.path, parser.lastLine(), "expected header after textual")
			}
		}
		if token.value != "header" {
			return modulemapSyntax(parser.path, token.line, "expected header after "+first.value)
		}
	}
	literal, err := parser.expectString()
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Kind: ReferenceHeader, Module: name, Path: literal.value, Line: literal.line})
	return parser.skipOptionalHeaderAttributes()
}

func (parser *modulemapParser) skipOptionalHeaderAttributes() error {
	token, ok := parser.peek()
	if !ok || token.quoted || token.value != "{" {
		return nil
	}
	depth := 0
	for {
		token, ok := parser.next()
		if !ok {
			return modulemapSyntax(parser.path, parser.lastLine(), "unterminated header attribute block")
		}
		if token.quoted {
			continue
		}
		switch token.value {
		case "{":
			depth++
		case "}":
			depth--
			if depth == 0 {
				return nil
			}
		}
	}
}

func (parser *modulemapParser) parseUmbrella(name string, result *ModuleMap) error {
	token, ok := parser.peek()
	if !ok {
		return modulemapSyntax(parser.path, parser.lastLine(), "truncated umbrella declaration")
	}
	if !token.quoted && token.value == "header" {
		parser.index++
		literal, err := parser.expectString()
		if err != nil {
			return err
		}
		result.References = append(result.References, Reference{Kind: ReferenceUmbrellaHeader, Module: name, Path: literal.value, Line: literal.line})
		return parser.skipOptionalHeaderAttributes()
	}
	literal, err := parser.expectString()
	if err != nil {
		return err
	}
	result.References = append(result.References, Reference{Kind: ReferenceUmbrellaDirectory, Module: name, Path: literal.value, Line: literal.line})
	return nil
}

func (parser *modulemapParser) parseLink(name string, result *ModuleMap) error {
	framework := false
	token, ok := parser.peek()
	if !ok {
		return modulemapSyntax(parser.path, parser.lastLine(), "truncated link declaration")
	}
	if !token.quoted && token.value == "framework" {
		framework = true
		parser.index++
	}
	literal, err := parser.expectString()
	if err != nil {
		return err
	}
	result.Links = append(result.Links, Link{Module: name, Name: literal.value, Framework: framework, Line: literal.line})
	return nil
}

func (parser *modulemapParser) parseRequires(result *ModuleMap) error {
	for {
		token, ok := parser.next()
		if !ok || token.quoted {
			return modulemapSyntax(parser.path, parser.lastLine(), "expected a requires feature")
		}
		feature := token.value
		if feature == "!" {
			token, ok = parser.next()
			if !ok || token.quoted {
				return modulemapSyntax(parser.path, parser.lastLine(), "expected a negated requires feature")
			}
			feature = "!" + token.value
		}
		result.Requires = append(result.Requires, feature)
		separator, present := parser.peek()
		if !present || separator.quoted || separator.value != "," {
			return nil
		}
		parser.index++
	}
}

// skipIdentifierList consumes the exact shape shared by export, export_as,
// use, and config_macros: an optional attribute list followed by a
// comma-separated list of wildcards or qualified identifiers.
func (parser *modulemapParser) skipIdentifierList() error {
	if err := parser.skipModuleAttributes(); err != nil {
		return err
	}
	for {
		token, ok := parser.peek()
		if !ok || token.quoted || token.value == "}" || token.value == "{" {
			return nil
		}
		switch {
		case token.value == "*":
			parser.index++
		case identifierToken(token.value):
			parser.index++
			for {
				dot, present := parser.peek()
				if !present || dot.quoted || dot.value != "." {
					break
				}
				parser.index++
				part, present := parser.next()
				if !present || part.quoted {
					return modulemapSyntax(parser.path, parser.lastLine(), "expected a qualified identifier")
				}
			}
		default:
			return nil
		}
		separator, present := parser.peek()
		if !present || separator.quoted || separator.value != "," {
			return nil
		}
		parser.index++
	}
}

func identifierToken(value string) bool {
	if value == "" {
		return false
	}
	first := value[0]
	return first == '_' || (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
}

func modulemapSyntax(logicalPath string, line int, detail string) error {
	return failFields(CodeModuleMapEscape, map[string]string{"module_map": logicalPath, "line": fmt.Sprintf("%d", line), "grammar": ModuleMapGrammarID}, "module map is not exactly parseable: %s", detail)
}

func uniqueStrings(values []string) []string {
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return append([]string{}, result...)
}
