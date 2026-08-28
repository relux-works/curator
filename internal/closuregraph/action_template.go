package closuregraph

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var actionSlotNamePattern = regexp.MustCompile(`^[a-z][a-z0-9._-]*$`)
var actionToolInvocationPattern = regexp.MustCompile(`^\$TOOL\([a-z][a-z0-9._-]*\)$`)

type actionTemplateReferences struct {
	tools  map[string]int
	reads  map[string]int
	writes map[string]int
}

func newActionTemplateReferences() actionTemplateReferences {
	return actionTemplateReferences{tools: map[string]int{}, reads: map[string]int{}, writes: map[string]int{}}
}

func validateActionTemplate(payload ActionPayload) error {
	if !actionToolInvocationPattern.MatchString(payload.ArgvTemplate[0]) {
		return fmt.Errorf("argv_template[0] must be exactly one $TOOL(slot) placeholder")
	}
	fields := []struct {
		name   string
		values []string
	}{
		{name: "tool_slot_names", values: payload.ToolSlotNames},
		{name: "read_slot_names", values: payload.ReadSlotNames},
		{name: "write_slot_names", values: payload.WriteSlotNames},
	}
	for _, field := range fields {
		for index, value := range field.values {
			if !actionSlotNamePattern.MatchString(value) {
				return fmt.Errorf("%s[%d] must match the closed action slot grammar", field.name, index)
			}
		}
	}

	references := newActionTemplateReferences()
	for index, value := range payload.ArgvTemplate {
		if _, err := scanActionTemplate(value, fmt.Sprintf("argv_template[%d]", index), &references); err != nil {
			return err
		}
	}
	if payload.WorkingDirectoryTemplate != "" {
		expanded, err := scanActionTemplate(payload.WorkingDirectoryTemplate, "working_directory_template", &references)
		if err != nil {
			return err
		}
		if err := validatePortablePath(expanded, "working_directory_template"); err != nil {
			return err
		}
	}
	if err := validateExactTemplateSlots("tool", payload.ToolSlotNames, references.tools); err != nil {
		return err
	}
	if err := validateExactTemplateSlots("read", payload.ReadSlotNames, references.reads); err != nil {
		return err
	}
	return validateExactTemplateSlots("write", payload.WriteSlotNames, references.writes)
}

func scanActionTemplate(value, field string, references *actionTemplateReferences) (string, error) {
	var expanded strings.Builder
	for offset := 0; offset < len(value); {
		relative := strings.IndexByte(value[offset:], '$')
		if relative < 0 {
			expanded.WriteString(value[offset:])
			break
		}
		start := offset + relative
		expanded.WriteString(value[offset:start])
		var kind string
		var slots map[string]int
		var prefixLength int
		switch {
		case strings.HasPrefix(value[start:], "$TOOL("):
			kind, slots, prefixLength = "tool", references.tools, len("$TOOL(")
		case strings.HasPrefix(value[start:], "$READ("):
			kind, slots, prefixLength = "read", references.reads, len("$READ(")
		case strings.HasPrefix(value[start:], "$WRITE("):
			kind, slots, prefixLength = "write", references.writes, len("$WRITE(")
		default:
			return "", fmt.Errorf("%s contains an unsupported action placeholder at byte %d", field, start)
		}
		nameStart := start + prefixLength
		endRelative := strings.IndexByte(value[nameStart:], ')')
		if endRelative < 0 {
			return "", fmt.Errorf("%s contains an unterminated %s placeholder", field, kind)
		}
		end := nameStart + endRelative
		name := value[nameStart:end]
		if !actionSlotNamePattern.MatchString(name) {
			return "", fmt.Errorf("%s contains a %s placeholder with an invalid slot name %q", field, kind, name)
		}
		slots[name]++
		expanded.WriteString(name)
		offset = end + 1
	}
	return expanded.String(), nil
}

func validateExactTemplateSlots(kind string, declared []string, referenced map[string]int) error {
	declaredSet := make(map[string]bool, len(declared))
	for _, slot := range declared {
		declaredSet[slot] = true
	}
	referencedSlots := make([]string, 0, len(referenced))
	for slot := range referenced {
		referencedSlots = append(referencedSlots, slot)
	}
	sort.Strings(referencedSlots)
	for _, slot := range referencedSlots {
		if !declaredSet[slot] {
			return fmt.Errorf("action template references undeclared %s slot %q", kind, slot)
		}
	}
	for _, slot := range declared {
		if referenced[slot] == 0 {
			return fmt.Errorf("action %s slot %q is declared but absent from its templates", kind, slot)
		}
	}
	return nil
}
