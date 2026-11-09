package profilegraph

import (
	"reflect"
	"strings"
	"unicode/utf8"
)

// Generated shapes are typed; input-derived strings still have to fit the
// published wire bounds. Never let JSON's UTF-8 replacement hide invalid input.
func validateWireStrings(g *Graph) error {
	valid := func(s string, maximum int) bool { return utf8.ValidString(s) && utf8.RuneCountInString(s) <= maximum }
	var stringsValid func(reflect.Value) bool
	stringsValid = func(v reflect.Value) bool {
		switch v.Kind() {
		case reflect.Pointer:
			return v.IsNil() || stringsValid(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if !stringsValid(v.Field(i)) {
					return false
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				if !stringsValid(v.Index(i)) {
					return false
				}
			}
		case reflect.String:
			return valid(v.String(), 4096)
		}
		return true
	}
	short := func(s string) bool { return s != "" && valid(s, 256) }
	relative := func(s string) bool {
		if s == "" || strings.HasPrefix(s, "/") || strings.ContainsAny(s, "\\\x00\r\n") || len(s) >= 2 && s[1] == ':' {
			return false
		}
		for _, part := range strings.Split(s, "/") {
			if part == ".." {
				return false
			}
		}
		return true
	}
	if !stringsValid(reflect.ValueOf(g)) || !short(g.Generator.Version) || !short(g.Scope.Profile) {
		return refusal("graph_invalid_input")
	}
	for _, env := range g.Scope.Environments {
		if !short(env) {
			return refusal("graph_invalid_input")
		}
	}
	for _, in := range g.Inputs {
		if !short(in.Root) || !relative(in.Path) {
			return refusal("graph_invalid_input")
		}
	}
	for _, n := range g.Nodes {
		if len(n.Identity) < 2 || len(n.Identity) > 128 {
			return refusal("graph_invalid_input")
		}
		if s := n.Source; s != nil {
			if !short(s.Locator) || s.Version != nil && !short(*s.Version) || s.Directory != "" && !relative(s.Directory) || s.Path != nil && !relative(*s.Path) {
				return refusal("graph_invalid_input")
			}
		}
		if c := n.Composition; c != nil {
			if !short(c.Declaration) || c.Weight < -2147483648 || c.Weight > 2147483647 {
				return refusal("graph_invalid_input")
			}
		}
		for _, load := range n.Loads {
			if !short(load.Environment) {
				return refusal("graph_invalid_input")
			}
		}
	}
	for _, doc := range g.Documents {
		if !short(doc.Environment) || !short(doc.Root) || !relative(doc.Path) {
			return refusal("graph_invalid_input")
		}
	}
	return nil
}
