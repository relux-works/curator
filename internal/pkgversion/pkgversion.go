// Package pkgversion implements the strict versions and the closed range
// grammar of the agent-environments capability (environments §1.4).
//
// Versions are Semantic Versioning 2.0 without build metadata; version tags
// carry a mandatory "v" prefix. Ranges follow node-semver 7.7.4 (caret, tilde,
// x-ranges, prerelease tags) restricted as the protocol states: no hyphen
// ranges, no "v" inside a range, and "latest" as a spelling of "*". Every
// exclusive upper bound is spelled "<X.Y.Z-0".
package pkgversion

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalidRange is the error class of a range that does not parse. Callers
// map it onto the protocol diagnostic profile_source_invalid.
var ErrInvalidRange = errors.New("profile_source_invalid")

// Version is one strict semantic version.
type Version struct {
	Major, Minor, Patch int64
	Prerelease          []string
}

var (
	numericRE      = regexp.MustCompile(`^(?:0|[1-9][0-9]*)$`)
	prereleaseIDRE = regexp.MustCompile(`^[0-9A-Za-z-]+$`)
)

// ParseVersion parses a version without the "v" prefix, as agent-context.json
// and agent-mcp.json spell it.
func ParseVersion(raw string) (Version, error) {
	core := raw
	prerelease := ""
	if index := strings.IndexByte(raw, '-'); index >= 0 {
		core = raw[:index]
		prerelease = raw[index+1:]
		if prerelease == "" {
			return Version{}, fmt.Errorf("invalid version %q: empty prerelease", raw)
		}
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: expected major.minor.patch", raw)
	}
	var numbers [3]int64
	for index, part := range parts {
		if !numericRE.MatchString(part) {
			return Version{}, fmt.Errorf("invalid version %q: component %q", raw, part)
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q: %v", raw, err)
		}
		numbers[index] = value
	}
	version := Version{Major: numbers[0], Minor: numbers[1], Patch: numbers[2]}
	if prerelease != "" {
		for _, id := range strings.Split(prerelease, ".") {
			if !validPrereleaseID(id) {
				return Version{}, fmt.Errorf("invalid version %q: prerelease identifier %q", raw, id)
			}
			version.Prerelease = append(version.Prerelease, id)
		}
	}
	return version, nil
}

func validPrereleaseID(id string) bool {
	if id == "" || !prereleaseIDRE.MatchString(id) {
		return false
	}
	if isNumeric(id) {
		return numericRE.MatchString(id)
	}
	return true
}

func isNumeric(id string) bool {
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return id != ""
}

// ParseTag reports whether a git tag name is a version candidate — a strict
// "v"-prefixed version without build metadata — and returns its version. A tag
// that does not parse is silently outside every range (environments §1.4).
func ParseTag(tag string) (Version, bool) {
	if !strings.HasPrefix(tag, "v") {
		return Version{}, false
	}
	version, err := ParseVersion(tag[1:])
	if err != nil {
		return Version{}, false
	}
	return version, true
}

// String renders the version without the "v" prefix.
func (version Version) String() string {
	text := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
	if len(version.Prerelease) > 0 {
		text += "-" + strings.Join(version.Prerelease, ".")
	}
	return text
}

// IsPrerelease reports whether the version carries a prerelease.
func (version Version) IsPrerelease() bool { return len(version.Prerelease) > 0 }

// Compare orders two versions by SemVer 2.0 precedence: -1, 0, or 1.
func Compare(a, b Version) int {
	switch {
	case a.Major != b.Major:
		return cmpInt(a.Major, b.Major)
	case a.Minor != b.Minor:
		return cmpInt(a.Minor, b.Minor)
	case a.Patch != b.Patch:
		return cmpInt(a.Patch, b.Patch)
	}
	switch {
	case len(a.Prerelease) == 0 && len(b.Prerelease) == 0:
		return 0
	case len(a.Prerelease) == 0:
		return 1
	case len(b.Prerelease) == 0:
		return -1
	}
	for index := 0; index < len(a.Prerelease) && index < len(b.Prerelease); index++ {
		left, right := a.Prerelease[index], b.Prerelease[index]
		if left == right {
			continue
		}
		leftNumeric, rightNumeric := isNumeric(left), isNumeric(right)
		switch {
		case leftNumeric && rightNumeric:
			l, _ := strconv.ParseInt(left, 10, 64)
			r, _ := strconv.ParseInt(right, 10, 64)
			return cmpInt(l, r)
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		default:
			if left < right {
				return -1
			}
			return 1
		}
	}
	return cmpInt(int64(len(a.Prerelease)), int64(len(b.Prerelease)))
}

func cmpInt(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// Sort orders versions ascending by precedence, in place.
func Sort(versions []Version) {
	sort.SliceStable(versions, func(i, j int) bool { return Compare(versions[i], versions[j]) < 0 })
}

// Comparator is one primitive after coercion: an operator and a version, or
// the any-comparator "*" (Any true).
type Comparator struct {
	Operator string // ">=", ">", "<=", "<", "="
	Version  Version
	Any      bool
}

// String renders the comparator in the vector spelling.
func (comparator Comparator) String() string {
	if comparator.Any {
		return "*"
	}
	return comparator.Operator + comparator.Version.String()
}

func (comparator Comparator) matches(version Version) bool {
	if comparator.Any {
		return true
	}
	order := Compare(version, comparator.Version)
	switch comparator.Operator {
	case "=":
		return order == 0
	case ">":
		return order > 0
	case ">=":
		return order >= 0
	case "<":
		return order < 0
	case "<=":
		return order <= 0
	}
	return false
}

// Range is a parsed range: one or more comparator sets joined by "||".
type Range struct {
	Text string
	Sets [][]Comparator
}

// ComparatorSets renders every set as the vector spelling of its comparators.
func (r Range) ComparatorSets() [][]string {
	out := make([][]string, 0, len(r.Sets))
	for _, set := range r.Sets {
		rendered := make([]string, 0, len(set))
		for _, comparator := range set {
			rendered = append(rendered, comparator.String())
		}
		out = append(out, rendered)
	}
	return out
}

// Satisfies reports whether version satisfies the range: some set is satisfied
// by every comparator, and — for a prerelease version — some comparator of
// that set names a prerelease on the same major.minor.patch.
func (r Range) Satisfies(version Version) bool {
	for _, set := range r.Sets {
		if setSatisfied(set, version) {
			return true
		}
	}
	return false
}

func setSatisfied(set []Comparator, version Version) bool {
	for _, comparator := range set {
		if !comparator.matches(version) {
			return false
		}
	}
	if !version.IsPrerelease() {
		return true
	}
	for _, comparator := range set {
		if comparator.Any || !comparator.Version.IsPrerelease() {
			continue
		}
		if comparator.Version.Major == version.Major && comparator.Version.Minor == version.Minor && comparator.Version.Patch == version.Patch {
			return true
		}
	}
	return false
}

// Highest returns the highest candidate satisfying the range and whether one
// exists.
func (r Range) Highest(candidates []Version) (Version, bool) {
	var best Version
	found := false
	for _, candidate := range candidates {
		if !r.Satisfies(candidate) {
			continue
		}
		if !found || Compare(candidate, best) > 0 {
			best = candidate
			found = true
		}
	}
	return best, found
}

var (
	primitiveRE   = regexp.MustCompile(`^(>=|<=|>|<|=|\^|~)?(\*|x|X|0|[1-9][0-9]*)(?:\.(\*|x|X|0|[1-9][0-9]*)(?:\.(\*|x|X|0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?)?)?$`)
	disjunctionRE = regexp.MustCompile(` *\|\| *`)
)

// ParseRange parses a range under the closed grammar. "latest" alone is "*".
func ParseRange(text string) (Range, error) {
	if text == "latest" {
		return Range{Text: text, Sets: [][]Comparator{{{Any: true}}}}, nil
	}
	if text == "" || strings.ContainsAny(text, "\t\r\n") {
		return Range{}, fmt.Errorf("%w: range %q does not parse", ErrInvalidRange, text)
	}
	if strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") {
		return Range{}, fmt.Errorf("%w: range %q does not parse", ErrInvalidRange, text)
	}
	parsed := Range{Text: text}
	for _, setText := range disjunctionRE.Split(text, -1) {
		if setText == "" {
			return Range{}, fmt.Errorf("%w: range %q has an empty comparator set", ErrInvalidRange, text)
		}
		var set []Comparator
		for _, primitiveText := range strings.Split(setText, " ") {
			if primitiveText == "" {
				return Range{}, fmt.Errorf("%w: range %q does not parse", ErrInvalidRange, text)
			}
			comparators, err := parsePrimitive(primitiveText)
			if err != nil {
				return Range{}, fmt.Errorf("%w: range %q: %v", ErrInvalidRange, text, err)
			}
			set = append(set, comparators...)
		}
		parsed.Sets = append(parsed.Sets, set)
	}
	return parsed, nil
}

type partial struct {
	major, minor, patch int64
	hasMinor, hasPatch  bool
	prerelease          []string
}

func isWild(component string) bool {
	return component == "" || component == "*" || component == "x" || component == "X"
}

func parsePrimitive(text string) ([]Comparator, error) {
	match := primitiveRE.FindStringSubmatch(text)
	if match == nil {
		return nil, fmt.Errorf("primitive %q does not parse", text)
	}
	operator := match[1]
	if isWild(match[2]) {
		// A bare "*", "x", or "X" with no operator — or with >=, <=, =,
		// ^, or ~ — is the any-comparator. A bare ">" or "<" on a
		// wildcard is rejected: node-semver reads those spellings as
		// match-nothing (<0.0.0-0), and environments §1.4 admits neither
		// reading, so neither meaning is silently substituted.
		if operator == ">" || operator == "<" {
			return nil, fmt.Errorf("primitive %q: > and < on a bare wildcard do not parse", text)
		}
		return []Comparator{{Any: true}}, nil
	}
	p := partial{}
	p.major, _ = strconv.ParseInt(match[2], 10, 64)
	if !isWild(match[3]) {
		p.hasMinor = true
		p.minor, _ = strconv.ParseInt(match[3], 10, 64)
		if !isWild(match[4]) {
			p.hasPatch = true
			p.patch, _ = strconv.ParseInt(match[4], 10, 64)
			if match[5] != "" {
				for _, id := range strings.Split(match[5], ".") {
					if !validPrereleaseID(id) {
						return nil, fmt.Errorf("primitive %q has an invalid prerelease identifier", text)
					}
					p.prerelease = append(p.prerelease, id)
				}
			}
		}
	}
	if !p.hasMinor && match[4] != "" && !isWild(match[4]) {
		// "1.x.3" reads as "1.x": a wildcard minor frees the patch too.
		p.hasPatch = false
	}
	full := Version{Major: p.major, Minor: p.minor, Patch: p.patch, Prerelease: p.prerelease}
	switch operator {
	case "", "=":
		switch {
		case p.hasPatch:
			return []Comparator{{Operator: "=", Version: full}}, nil
		case p.hasMinor:
			return []Comparator{{Operator: ">=", Version: full}, {Operator: "<", Version: bound(p.major, p.minor+1, 0)}}, nil
		default:
			return []Comparator{{Operator: ">=", Version: full}, {Operator: "<", Version: bound(p.major+1, 0, 0)}}, nil
		}
	case ">=":
		return []Comparator{{Operator: ">=", Version: full}}, nil
	case ">":
		switch {
		case p.hasPatch:
			return []Comparator{{Operator: ">", Version: full}}, nil
		case p.hasMinor:
			return []Comparator{{Operator: ">=", Version: Version{Major: p.major, Minor: p.minor + 1}}}, nil
		default:
			return []Comparator{{Operator: ">=", Version: Version{Major: p.major + 1}}}, nil
		}
	case "<":
		switch {
		case p.hasPatch:
			return []Comparator{{Operator: "<", Version: full}}, nil
		case p.hasMinor:
			return []Comparator{{Operator: "<", Version: bound(p.major, p.minor, 0)}}, nil
		default:
			return []Comparator{{Operator: "<", Version: bound(p.major, 0, 0)}}, nil
		}
	case "<=":
		switch {
		case p.hasPatch:
			return []Comparator{{Operator: "<=", Version: full}}, nil
		case p.hasMinor:
			return []Comparator{{Operator: "<", Version: bound(p.major, p.minor+1, 0)}}, nil
		default:
			return []Comparator{{Operator: "<", Version: bound(p.major+1, 0, 0)}}, nil
		}
	case "^":
		lower := Comparator{Operator: ">=", Version: full}
		var upper Version
		switch {
		case p.major != 0 || !p.hasMinor:
			upper = bound(p.major+1, 0, 0)
		case p.minor != 0 || !p.hasPatch:
			upper = bound(0, p.minor+1, 0)
		default:
			upper = bound(0, 0, p.patch+1)
		}
		return []Comparator{lower, {Operator: "<", Version: upper}}, nil
	case "~":
		lower := Comparator{Operator: ">=", Version: full}
		var upper Version
		if p.hasMinor {
			upper = bound(p.major, p.minor+1, 0)
		} else {
			upper = bound(p.major+1, 0, 0)
		}
		return []Comparator{lower, {Operator: "<", Version: upper}}, nil
	}
	return nil, fmt.Errorf("primitive %q has an unknown operator", text)
}

// bound is the lowest prerelease of a version: the "<X.Y.Z-0" upper bound.
func bound(major, minor, patch int64) Version {
	return Version{Major: major, Minor: minor, Patch: patch, Prerelease: []string{"0"}}
}
