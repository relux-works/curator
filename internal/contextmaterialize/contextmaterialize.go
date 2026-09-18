// Package contextmaterialize assembles the deterministic root-context bytes of
// environments §5: the curator-root-context-v2 generation header (§5.1), the
// emitted order under both precedence primitives (§5, §6), chapter parts, the
// monolithic form (§5.2), the zero-module case (§5.4), the system-prompt
// output (§5.5) with the §3 system-module admission, and the content-hash
// binding of a surface (§5.6).
//
// Materialized root context is a pure function of (lock, precedence policy,
// environment identifier, form); nothing here reads a clock, a path, or an
// operator identity.
package contextmaterialize

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// HeaderTypeLine is the generation-header type line.
const HeaderTypeLine = "curator-root-context-v2"

// Fixed header lines (environments §5.1).
const (
	GeneratedLine = "generated: Curator Protocol environments revision 1 (https://github.com/relux-works/curator-spec)"
	NoticeLine    = "notice: generated file; direct edits are unsupported and are detected as drift; update the source profile repository or its composed profiles instead"
)

// Precedence primitives (environments §6).
const (
	WinnerHigher   = "higher-weight"
	WinnerLower    = "lower-weight"
	PlacementLast  = "winner-last"
	PlacementFirst = "winner-first"
)

// Forms (environments §5.2, §5.3).
const (
	FormMonolithic = "monolithic"
	FormReferenced = "referenced"
)

// Precedence is the machine's precedence policy.
type Precedence struct {
	Winner    string
	Placement string
}

// DefaultPrecedence is the default pair.
var DefaultPrecedence = Precedence{Winner: WinnerHigher, Placement: PlacementLast}

// Validate checks both primitives.
func (p Precedence) Validate() error {
	if p.Winner != WinnerHigher && p.Winner != WinnerLower {
		return fmt.Errorf("precedence winner %q is not higher-weight or lower-weight", p.Winner)
	}
	if p.Placement != PlacementLast && p.Placement != PlacementFirst {
		return fmt.Errorf("precedence placement %q is not winner-last or winner-first", p.Placement)
	}
	return nil
}

// Module is one validated module of a context member with its exact bytes.
type Module struct {
	contextpkg.Module
	Bytes []byte
}

// Package is the module set of one context member.
type Package struct {
	// HasContext is false for a pure umbrella, which declares no
	// root-context surface at all.
	HasContext bool
	Modules    []Module
}

// EmittedOrder sorts the lock's context members: the core §7 Kahn order
// (dependencies first, smallest ready name first) stably sorted by effective
// weight under the precedence policy.
func EmittedOrder(lock *contextlock.Lock, precedence Precedence) ([]contextlock.Member, error) {
	if err := precedence.Validate(); err != nil {
		return nil, err
	}
	members := map[string]contextlock.Member{}
	for _, member := range lock.Contexts() {
		members[member.Name] = member
	}
	dependents := map[string]map[string]bool{}
	indegree := map[string]int{}
	for name := range members {
		dependents[name] = map[string]bool{}
		indegree[name] = 0
	}
	for name, member := range members {
		for _, requirer := range member.RequiredBy {
			if _, known := members[requirer]; !known || dependents[name][requirer] {
				continue
			}
			dependents[name][requirer] = true
			indegree[requirer]++
		}
	}
	var ready []string
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)
	var ordered []contextlock.Member
	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]
		ordered = append(ordered, members[name])
		consumers := make([]string, 0, len(dependents[name]))
		for consumer := range dependents[name] {
			consumers = append(consumers, consumer)
		}
		sort.Strings(consumers)
		for _, consumer := range consumers {
			indegree[consumer]--
			if indegree[consumer] == 0 {
				ready = append(ready, consumer)
			}
		}
		sort.Strings(ready)
	}
	if len(ordered) != len(members) {
		return nil, fmt.Errorf("the lock's context members form a cycle")
	}
	ascending := (precedence.Winner == WinnerHigher) == (precedence.Placement == PlacementLast)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ascending {
			return ordered[i].Weight < ordered[j].Weight
		}
		return ordered[i].Weight > ordered[j].Weight
	})
	return ordered, nil
}

// Header renders the generation header part for the emitted order.
func Header(lock *contextlock.Lock, lockHash string, precedence Precedence, order []contextlock.Member) ([]byte, error) {
	root, ok := lock.RootMember()
	if !ok {
		return nil, fmt.Errorf("lock root %s is not a context member", lock.Root)
	}
	var out bytes.Buffer
	out.WriteString("<!--\n")
	out.WriteString(HeaderTypeLine + "\n")
	fmt.Fprintf(&out, "root: %s %s %s\n", root.Name, root.Version, root.Pin())
	for _, member := range order {
		fmt.Fprintf(&out, "member: %s %s %s weight %s", member.Name, member.Version, member.Pin(), strconv.FormatInt(member.Weight, 10))
		if member.Overlay {
			out.WriteString(" overlay")
		}
		out.WriteString("\n")
	}
	fmt.Fprintf(&out, "precedence: winner=%s placement=%s\n", precedence.Winner, precedence.Placement)
	fmt.Fprintf(&out, "lock: %s\n", lockHash)
	out.WriteString(GeneratedLine + "\n")
	out.WriteString(NoticeLine + "\n")
	out.WriteString("-->\n")
	return out.Bytes(), nil
}

// Join assembles parts under the §5 rule: every part ends with exactly one
// LF, and the document is the parts joined with one additional LF between
// adjacent parts.
func Join(parts [][]byte) []byte {
	var out bytes.Buffer
	for index, part := range parts {
		if index > 0 {
			out.WriteByte('\n')
		}
		out.Write(part)
	}
	return out.Bytes()
}

// ChapterPart renders "---" LF LF "## Context: <name> <version>" LF.
func ChapterPart(member contextlock.Member) []byte {
	return []byte(fmt.Sprintf("---\n\n## Context: %s %s\n", member.Name, member.Version))
}

// Applicable returns the modules of a package of the given class that apply
// to the environment, in manifest order.
func Applicable(pkg Package, class, environment string) []Module {
	var out []Module
	for _, module := range pkg.Modules {
		moduleClass := module.Class
		if moduleClass == "" {
			moduleClass = "root"
		}
		if moduleClass == class && module.Applies(environment) {
			out = append(out, module)
		}
	}
	return out
}

// Monolithic assembles the monolithic root-context document for an
// environment. It returns written=false when the root declares no context
// (environments §2): no surface exists and no file is written.
func Monolithic(lock *contextlock.Lock, lockHash string, precedence Precedence, environment string, packages map[string]Package) (document []byte, written bool, err error) {
	rootPkg, ok := packages[lock.Root]
	if !ok {
		return nil, false, fmt.Errorf("no package content for the root %s", lock.Root)
	}
	if !rootPkg.HasContext {
		return nil, false, nil
	}
	order, err := EmittedOrder(lock, precedence)
	if err != nil {
		return nil, false, err
	}
	header, err := Header(lock, lockHash, precedence, order)
	if err != nil {
		return nil, false, err
	}
	parts := [][]byte{header}
	for _, member := range order {
		pkg, ok := packages[member.Name]
		if !ok {
			return nil, false, fmt.Errorf("no package content for member %s", member.Name)
		}
		modules := Applicable(pkg, "root", environment)
		if len(modules) == 0 {
			continue
		}
		parts = append(parts, ChapterPart(member))
		for _, module := range modules {
			if err := contextpkg.ValidateModuleBytes(module.Bytes); err != nil {
				return nil, false, fmt.Errorf("module %s of %s: %v", module.Path, member.Name, err)
			}
			parts = append(parts, module.Bytes)
		}
	}
	return Join(parts), true, nil
}

// SystemPrompt assembles the system-prompt output (environments §5.5): the
// admitted applicable system modules of every member in emitted order,
// joined with no header and no chapter parts. written is false when no
// admitted module applies. dropped names every non-admitted applicable
// module in the same order, so the caller warns
// context_system_module_dropped per module under drop. Under error the
// first dropped module refuses the materialization with
// context_system_module_transitive and no document is returned. Only
// admitted modules are byte-validated and joined: a skipped module
// contributes no bytes, and its validity was settled at snapshot
// validation.
func SystemPrompt(lock *contextlock.Lock, precedence Precedence, environment string, packages map[string]Package, admission Admission) (document []byte, written bool, dropped []DroppedModule, err error) {
	policy, err := admission.Policy()
	if err != nil {
		return nil, false, nil, err
	}
	order, err := EmittedOrder(lock, precedence)
	if err != nil {
		return nil, false, nil, err
	}
	for _, member := range order {
		if _, ok := packages[member.Name]; !ok {
			return nil, false, nil, fmt.Errorf("no package content for member %s", member.Name)
		}
	}
	admitted, dropped := ClassifySystemModules(lock, order, packages, environment, admission)
	if policy == TransitiveError && len(dropped) > 0 {
		first := dropped[0]
		return nil, false, dropped, &TransitiveSystemModuleError{Package: first.Package, Module: first.Path}
	}
	if len(admitted) == 0 {
		return nil, false, dropped, nil
	}
	parts := make([][]byte, 0, len(admitted))
	for _, module := range admitted {
		if err := contextpkg.ValidateModuleBytes(module.Module.Bytes); err != nil {
			return nil, false, dropped, fmt.Errorf("module %s of %s: %v", module.Module.Path, module.Package, err)
		}
		parts = append(parts, module.Module.Bytes)
	}
	return Join(parts), true, dropped, nil
}

// SurfaceHash is the core §8 content hash over a materialized file set keyed
// by home-relative portable path: records "path NUL content" joined by NUL,
// in bytewise path order, prefixed "sha256:".
func SurfaceHash(files map[string][]byte) string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	digest := sha256.New()
	for index, path := range paths {
		if index > 0 {
			digest.Write([]byte{0})
		}
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write(files[path])
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil))
}

// FileHash is the plain SHA-256 of one file's bytes, "sha256:" prefixed.
func FileHash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
