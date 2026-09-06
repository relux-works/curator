// Package contextresolve implements the joint resolution of environments
// §1.4: the core §7 closure with its admission rule generalized from exact
// refs to version constraints, over context packages, skills, and MCP
// declaration packages together, producing the context-lock-v1 object of
// environments §1.3 and the effective weights of environments §6.
//
// The algorithm is fixed so that two managers lock identically: seed, select
// the lexicographically smallest pending name and expand it, re-select
// downward when an added constraint excludes a selection, never increase a
// selection, and check every constraint at the end. No backtracking across
// names is performed.
package contextresolve

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/pkgversion"
)

// Diagnostic codes (environments §1.1, §2.1, §6.1).
const (
	DiagRangeConflict        = "context_range_conflict"
	DiagVersionMismatch      = "context_version_mismatch"
	DiagWeightConflict       = "context_weight_conflict"
	DiagWeightsNotRoot       = "context_weights_not_root"
	DiagWeightsDuplicate     = "context_weights_duplicate"
	DiagWeightUnknown        = "context_weight_unknown"
	DiagCompositionInvalid   = "environment_composition_invalid"
	DiagMCPPackageNotAllowed = "mcp_package_not_allowed"
	DiagSourceMismatch       = "context_source_mismatch"
)

// MachineRequirer names the machine as the requirer of the install
// declaration, every overlay declaration, and every direct declaration.
const MachineRequirer = "machine"

// Requirement is one constraint-contributing declaration.
type Requirement struct {
	Kind      string
	Name      string
	Source    string // canonical network identity, or "" to take the source's own
	Range     string
	Tag       string
	Revision  string
	Directory string
	Weight    *int64
}

// Form renders the requirement's constraint for diagnostics: "range ^3",
// "tag v3.2.1", or "revision <hex>".
func (r Requirement) Form() string {
	switch {
	case r.Range != "":
		return "range " + r.Range
	case r.Tag != "":
		return "tag " + r.Tag
	case r.Revision != "":
		return "revision " + r.Revision
	default:
		return "state"
	}
}

func (r Requirement) exact() bool { return r.Range == "" }

// Package is the manifest of one package at one commit as resolution reads
// it. Version is empty for a skill.
type Package struct {
	Version  string
	Weight   int64
	Weights  map[string]int64
	Requires []Requirement
}

// Candidate is one version tag of a source peeled to its commit.
type Candidate struct {
	Tag     string
	Version pkgversion.Version
	Commit  string
}

// Source supplies candidates and manifests. Every method receives the
// canonical identity the requirement declared ("" when the requirement gave
// none) and returns the identity resolution records.
type Source interface {
	// Identity returns the canonical source identity of a name.
	Identity(kind, name, declared string) (string, error)
	// Candidates returns every version tag of the source peeled to a commit.
	Candidates(kind, name, source string) ([]Candidate, error)
	// ResolveTag peels one exact tag to its commit.
	ResolveTag(kind, name, source, tag string) (string, error)
	// Manifest reads the package at a commit (below directory when set).
	Manifest(kind, name, source, directory, commit string) (*Package, error)
}

// StatePackage is a path or local package: a fixed manifest under a state
// hash, never fetched.
type StatePackage struct {
	StateHash string
	Manifest  *Package
}

// Overlay is one machine overlay declaration (environments §6).
type Overlay struct {
	Name     string
	Source   string
	Range    string
	Tag      string
	Revision string
	// Directory selects a directory within a git snapshot.
	Directory string
	// Weight is the declaration's weight; nil takes the machine default.
	Weight *int64
	// State is set for a path overlay.
	State *StatePackage
}

// Input is one resolution request.
type Input struct {
	// Root is the install declaration: kind context.
	Root Requirement
	// RootState is set for a path or local root; Root then carries only Name.
	RootState *StatePackage
	Overlays  []Overlay
	// OverlayDefaultWeight is the machine's overlay_default_weight knob.
	OverlayDefaultWeight int64
	// Direct declarations are the machine's exact skill declarations
	// (environments §9.4), attributed to the machine.
	Direct []Requirement
	// MCPAllowlist bounds MCP declaration packages; empty permits every identity.
	MCPAllowlist []string
}

// RequirerConstraint is one requirer's constraint in a conflict detail.
type RequirerConstraint struct {
	Requirer   string `json:"requirer"`
	Constraint string `json:"constraint"`
}

// RequirerWeight is one requirer's edge weight in a weight-conflict detail.
type RequirerWeight struct {
	Requirer string `json:"requirer"`
	Weight   int64  `json:"weight"`
}

// Error is a resolution failure carrying its protocol diagnostic and detail.
type Error struct {
	Diagnostic string
	Name       string
	Requirers  []RequirerConstraint
	Candidates []string
	// Tag and ManifestVersion are set for context_version_mismatch.
	Tag             string
	ManifestVersion string
	// WeightRequirers is set for context_weight_conflict.
	WeightRequirers []RequirerWeight
	Detail          string
}

func (e *Error) Error() string {
	var parts []string
	if e.Name != "" {
		parts = append(parts, e.Name)
	}
	for _, requirer := range e.Requirers {
		parts = append(parts, fmt.Sprintf("%s requires %s", requirer.Requirer, requirer.Constraint))
	}
	for _, requirer := range e.WeightRequirers {
		parts = append(parts, fmt.Sprintf("%s declares weight %d", requirer.Requirer, requirer.Weight))
	}
	if len(e.Candidates) > 0 {
		parts = append(parts, "candidates "+strings.Join(e.Candidates, ", "))
	}
	if e.Tag != "" {
		parts = append(parts, fmt.Sprintf("tag %s, manifest version %s", e.Tag, e.ManifestVersion))
	}
	if e.Detail != "" {
		parts = append(parts, e.Detail)
	}
	return e.Diagnostic + ": " + strings.Join(parts, "; ")
}

// Warning is one non-blocking resolution finding.
type Warning struct {
	Diagnostic string
	Name       string
	Requirers  []RequirerWeight
}

// Result is a successful resolution.
type Result struct {
	Lock     *contextlock.Lock
	LockHash string
	Warnings []Warning
	// Members carries the resolved commit or state of every member, keyed by
	// kind and name, so a caller can install store entries.
	Members map[string]Resolved
}

// Resolved is where a member was resolved from.
type Resolved struct {
	Kind      string
	Name      string
	Source    string
	Directory string
	Commit    string
	StateHash string
	Version   string
	Package   *Package
}

// Key joins kind and name.
func Key(kind, name string) string { return kind + ":" + name }

// canonicalForAgreement maps a declared source onto its comparison key: the
// core §6.1 canonical identity for network URLs, the trimmed raw URL for
// file:// remotes (which carry no network identity) and for malformed
// inputs (which the Identity boundary rejects with profile_source_invalid).
// An already-canonical host/path passes through unchanged.
func canonicalForAgreement(declared string) string {
	trimmed := strings.TrimSpace(declared)
	if trimmed == "" || identity.ValidCanonical(trimmed) {
		return trimmed
	}
	canonical, err := identity.Parse(trimmed)
	if err != nil || canonical == "" {
		return trimmed
	}
	return canonical
}

type constraint struct {
	requirement Requirement
	requirer    string // "machine" or "<name>@<version>"
	attributed  string // the selection key that contributed it ("" for machine)
	order       int
}

type selection struct {
	kind      string
	name      string
	source    string
	directory string
	commit    string
	state     *StatePackage
	version   pkgversion.Version
	hasVer    bool
	pkg       *Package
	tag       string // version tag the selection came from, "" otherwise
	overlay   *Overlay
}

func (s *selection) key() string {
	if s.state != nil {
		return s.name + "@state"
	}
	return s.name + "@" + s.commit
}

func (s *selection) requirerName() string {
	if s.hasVer {
		return s.name + "@" + s.version.String()
	}
	return s.name
}

type resolver struct {
	source      Source
	input       Input
	constraints map[string][]constraint // by name
	kinds       map[string]string       // name → kind
	selected    map[string]*selection
	ceiling     map[string]pkgversion.Version
	hasCeiling  map[string]bool
	pending     map[string]bool
	overlays    map[string]*Overlay
	order       int
}

// Resolve runs the algorithm over the input.
func Resolve(source Source, input Input) (*Result, error) {
	r := &resolver{
		source:      source,
		input:       input,
		constraints: map[string][]constraint{},
		kinds:       map[string]string{},
		selected:    map[string]*selection{},
		ceiling:     map[string]pkgversion.Version{},
		hasCeiling:  map[string]bool{},
		pending:     map[string]bool{},
		overlays:    map[string]*Overlay{},
	}
	return r.run()
}

func (r *resolver) run() (*Result, error) {
	// 1. Seed.
	root := r.input.Root
	root.Kind = contextlock.KindContext
	if r.input.RootState == nil && root.Range == "" && root.Tag == "" && root.Revision == "" {
		root.Range = "latest"
	}
	r.add(root, MachineRequirer, "")
	for index := range r.input.Overlays {
		overlay := &r.input.Overlays[index]
		if _, declared := r.constraints[overlay.Name]; declared || overlay.Name == root.Name {
			return nil, &Error{Diagnostic: DiagCompositionInvalid, Name: overlay.Name,
				Detail: "overlay declaration repeats a name already declared"}
		}
		requirement := Requirement{Kind: contextlock.KindContext, Name: overlay.Name, Source: overlay.Source,
			Range: overlay.Range, Tag: overlay.Tag, Revision: overlay.Revision, Directory: overlay.Directory}
		if overlay.State == nil && requirement.Range == "" && requirement.Tag == "" && requirement.Revision == "" {
			requirement.Range = "latest"
		}
		r.overlays[overlay.Name] = overlay
		r.add(requirement, MachineRequirer, "")
	}
	for _, direct := range r.input.Direct {
		direct.Kind = contextlock.KindSkill
		r.add(direct, MachineRequirer, "")
	}

	// 2–3. Select, expand, and re-select downward until nothing is pending.
	for {
		name, ok := r.nextPending()
		if !ok {
			break
		}
		if err := r.selectName(name); err != nil {
			return nil, err
		}
	}

	// 4. Check.
	for name, constraints := range r.constraints {
		sel := r.selected[name]
		if sel == nil {
			return nil, r.conflict(name, nil)
		}
		for _, c := range constraints {
			if !r.admits(c.requirement, sel) {
				return nil, r.conflict(name, r.candidateVersions(sel))
			}
		}
	}
	return r.build()
}

func (r *resolver) nextPending() (string, bool) {
	best := ""
	for name := range r.pending {
		if best == "" || name < best {
			best = name
		}
	}
	if best == "" {
		return "", false
	}
	delete(r.pending, best)
	return best, true
}

func (r *resolver) add(requirement Requirement, requirer, attributed string) {
	r.order++
	if existing, known := r.kinds[requirement.Name]; known && existing != requirement.Kind {
		// One name, one kind: the closure is keyed by name across kinds.
		requirement.Kind = existing
	}
	r.kinds[requirement.Name] = requirement.Kind
	r.constraints[requirement.Name] = append(r.constraints[requirement.Name], constraint{
		requirement: requirement, requirer: requirer, attributed: attributed, order: r.order,
	})
	r.pending[requirement.Name] = true
}

// drop removes every constraint attributed to one selection key and returns
// the names whose constraint set changed.
func (r *resolver) drop(attributed string) []string {
	var changed []string
	for name, constraints := range r.constraints {
		kept := constraints[:0]
		dropped := false
		for _, c := range constraints {
			if c.attributed == attributed {
				dropped = true
				continue
			}
			kept = append(kept, c)
		}
		if dropped {
			changed = append(changed, name)
			if len(kept) == 0 {
				delete(r.constraints, name)
			} else {
				r.constraints[name] = kept
			}
		}
	}
	sort.Strings(changed)
	return changed
}

// leave removes a member with no constraint from the closure together with
// everything it contributed.
func (r *resolver) leave(name string) {
	sel := r.selected[name]
	if sel == nil {
		return
	}
	delete(r.selected, name)
	delete(r.pending, name)
	for _, changed := range r.drop(sel.key()) {
		if _, still := r.constraints[changed]; !still {
			r.leave(changed)
			continue
		}
		r.pending[changed] = true
	}
}

func (r *resolver) selectName(name string) error {
	constraints := r.constraints[name]
	if len(constraints) == 0 {
		r.leave(name)
		return nil
	}
	kind := r.kinds[name]
	current := r.selected[name]

	// Source identity and directory must agree across every requirer.
	// Comparison is over the core §6.1 canonical identity (environments
	// §1.4): two spellings of one repository are one identity and never a
	// spurious context_source_mismatch. File:// remotes (no network
	// identity) compare as their raw URLs so distinct test remotes stay
	// distinct.
	declaredSource, directory := "", ""
	declaredCanonical := ""
	for _, c := range constraints {
		if c.requirement.Source != "" {
			canonical := canonicalForAgreement(c.requirement.Source)
			if declaredSource != "" && declaredCanonical != canonical {
				return &Error{Diagnostic: DiagSourceMismatch, Name: name,
					Detail: fmt.Sprintf("requirers disagree on the source identity (%s vs %s)", declaredSource, c.requirement.Source)}
			}
			declaredSource = c.requirement.Source
			declaredCanonical = canonical
		}
		if c.requirement.Directory != "" {
			if directory != "" && directory != c.requirement.Directory {
				return &Error{Diagnostic: DiagSourceMismatch, Name: name,
					Detail: fmt.Sprintf("requirers disagree on the directory (%s vs %s)", directory, c.requirement.Directory)}
			}
			directory = c.requirement.Directory
		}
	}

	// A path or local package has exactly one candidate: its state.
	if state := r.stateFor(name); state != nil {
		if current != nil && current.state == state {
			return nil
		}
		sel := &selection{kind: kind, name: name, state: state, pkg: state.Manifest, overlay: r.overlays[name]}
		if state.Manifest.Version != "" {
			version, err := pkgversion.ParseVersion(state.Manifest.Version)
			if err != nil {
				return fmt.Errorf("%s: %v", name, err)
			}
			sel.version, sel.hasVer = version, true
		}
		return r.install(name, sel)
	}

	source, err := r.source.Identity(kind, name, declaredSource)
	if err != nil {
		return err
	}
	if kind == contextlock.KindMCP && len(r.input.MCPAllowlist) > 0 && !identity.Allowed(source, r.input.MCPAllowlist) {
		return &Error{Diagnostic: DiagMCPPackageNotAllowed, Name: name,
			Detail: fmt.Sprintf("source %s is outside the machine's MCP package allowlist", source)}
	}
	candidates, err := r.source.Candidates(kind, name, source)
	if err != nil {
		return err
	}
	sort.SliceStable(candidates, func(i, j int) bool { return pkgversion.Compare(candidates[i].Version, candidates[j].Version) < 0 })

	var chosen *selection
	exactCommit := ""
	for _, c := range constraints {
		if !c.requirement.exact() {
			continue
		}
		commit := c.requirement.Revision
		if c.requirement.Tag != "" {
			commit, err = r.source.ResolveTag(kind, name, source, c.requirement.Tag)
			if err != nil {
				return err
			}
		}
		if exactCommit != "" && exactCommit != commit {
			return r.conflict(name, []string{})
		}
		exactCommit = commit
	}
	if exactCommit != "" {
		chosen = &selection{kind: kind, name: name, source: source, directory: directory, commit: exactCommit, overlay: r.overlays[name]}
		// The version a fixed candidate carries: the highest version tag
		// peeling to the commit for a skill; the manifest version otherwise.
		for _, candidate := range candidates {
			if candidate.Commit == exactCommit {
				chosen.version, chosen.hasVer, chosen.tag = candidate.Version, true, candidate.Tag
			}
		}
		pkg, err := r.source.Manifest(kind, name, source, directory, exactCommit)
		if err != nil {
			return err
		}
		chosen.pkg = pkg
		if kind != contextlock.KindSkill {
			version, err := pkgversion.ParseVersion(pkg.Version)
			if err != nil {
				return fmt.Errorf("%s: manifest version %q: %v", name, pkg.Version, err)
			}
			if chosen.hasVer && r.exactVersionTagPinned(constraints, chosen.tag) && pkgversion.Compare(version, chosen.version) != 0 {
				return &Error{Diagnostic: DiagVersionMismatch, Name: name, Tag: chosen.tag, ManifestVersion: pkg.Version}
			}
			chosen.version, chosen.hasVer = version, true
			if !r.exactVersionTagPinned(constraints, chosen.tag) {
				chosen.tag = ""
			}
		}
		for _, c := range constraints {
			if !r.admits(c.requirement, chosen) {
				var considered []string
				if chosen.hasVer {
					considered = []string{chosen.version.String()}
				}
				return r.conflict(name, considered)
			}
		}
		if current != nil && current.commit == exactCommit {
			return nil
		}
		if r.hasCeiling[name] && chosen.hasVer && pkgversion.Compare(chosen.version, r.ceiling[name]) > 0 {
			return r.conflict(name, []string{chosen.version.String()})
		}
		return r.install(name, chosen)
	}

	// Ranges only: the highest satisfying candidate, never above a previous
	// selection.
	var best *Candidate
	for index := range candidates {
		candidate := candidates[index]
		if r.hasCeiling[name] && pkgversion.Compare(candidate.Version, r.ceiling[name]) > 0 {
			continue
		}
		satisfied := true
		for _, c := range constraints {
			parsed, err := pkgversion.ParseRange(c.requirement.Range)
			if err != nil {
				return fmt.Errorf("%s: %v", name, err)
			}
			if !parsed.Satisfies(candidate.Version) {
				satisfied = false
				break
			}
		}
		if satisfied && (best == nil || pkgversion.Compare(candidate.Version, best.Version) > 0) {
			best = &candidates[index]
		}
	}
	if best == nil {
		var considered []string
		for _, candidate := range candidates {
			considered = append(considered, candidate.Version.String())
		}
		if considered == nil {
			considered = []string{}
		}
		return r.conflict(name, considered)
	}
	if current != nil && current.commit == best.Commit {
		return nil
	}
	if current != nil && pkgversion.Compare(best.Version, current.version) > 0 {
		// Rule 3: a selection never increases; keep the current one while it
		// still satisfies every constraint.
		return nil
	}
	chosen = &selection{kind: kind, name: name, source: source, directory: directory, commit: best.Commit,
		version: best.Version, hasVer: true, tag: best.Tag, overlay: r.overlays[name]}
	pkg, err := r.source.Manifest(kind, name, source, directory, best.Commit)
	if err != nil {
		return err
	}
	chosen.pkg = pkg
	if kind != contextlock.KindSkill {
		version, err := pkgversion.ParseVersion(pkg.Version)
		if err != nil {
			return fmt.Errorf("%s: manifest version %q: %v", name, pkg.Version, err)
		}
		if pkgversion.Compare(version, best.Version) != 0 {
			return &Error{Diagnostic: DiagVersionMismatch, Name: name, Tag: best.Tag, ManifestVersion: pkg.Version}
		}
	}
	return r.install(name, chosen)
}

// exactVersionTagPinned reports whether some exact constraint names the
// version tag the selection came from, so the manifest version must match it.
func (r *resolver) exactVersionTagPinned(constraints []constraint, tag string) bool {
	if tag == "" {
		return false
	}
	for _, c := range constraints {
		if c.requirement.Tag == tag {
			return true
		}
	}
	return false
}

func (r *resolver) stateFor(name string) *StatePackage {
	if name == r.input.Root.Name && r.input.RootState != nil {
		return r.input.RootState
	}
	if overlay := r.overlays[name]; overlay != nil && overlay.State != nil {
		return overlay.State
	}
	return nil
}

func (r *resolver) admits(requirement Requirement, sel *selection) bool {
	if sel.state != nil {
		// A path or local package is admitted only by the formless machine
		// declaration that named it; no requirement names a path source.
		return requirement.Range == "" && requirement.Tag == "" && requirement.Revision == ""
	}
	if requirement.exact() {
		if requirement.Revision != "" {
			return requirement.Revision == sel.commit
		}
		commit, err := r.source.ResolveTag(sel.kind, sel.name, sel.source, requirement.Tag)
		return err == nil && commit == sel.commit
	}
	parsed, err := pkgversion.ParseRange(requirement.Range)
	if err != nil {
		return false
	}
	if !sel.hasVer {
		return false
	}
	return parsed.Satisfies(sel.version)
}

// install records a selection, dropping and re-expanding as rule 3 requires.
func (r *resolver) install(name string, sel *selection) error {
	if previous := r.selected[name]; previous != nil {
		for _, changed := range r.drop(previous.key()) {
			if _, still := r.constraints[changed]; !still {
				r.leave(changed)
				continue
			}
			r.pending[changed] = true
		}
	}
	r.selected[name] = sel
	if sel.hasVer {
		r.ceiling[name] = sel.version
		r.hasCeiling[name] = true
	}
	if sel.pkg != nil {
		if name != r.input.Root.Name && len(sel.pkg.Weights) > 0 {
			return &Error{Diagnostic: DiagWeightsNotRoot, Name: name}
		}
		for _, requirement := range sel.pkg.Requires {
			r.add(requirement, sel.requirerName(), sel.key())
		}
	}
	delete(r.pending, name)
	return nil
}

func (r *resolver) candidateVersions(sel *selection) []string {
	if sel.state != nil || !sel.hasVer {
		return []string{}
	}
	return []string{sel.version.String()}
}

func (r *resolver) conflict(name string, candidates []string) *Error {
	constraints := append([]constraint(nil), r.constraints[name]...)
	sort.SliceStable(constraints, func(i, j int) bool { return constraints[i].order < constraints[j].order })
	err := &Error{Diagnostic: DiagRangeConflict, Name: name, Candidates: candidates}
	if err.Candidates == nil {
		err.Candidates = []string{}
	}
	for _, c := range constraints {
		err.Requirers = append(err.Requirers, RequirerConstraint{Requirer: c.requirer, Constraint: c.requirement.Form()})
	}
	return err
}

// build computes effective weights and required_by and assembles the lock.
func (r *resolver) build() (*Result, error) {
	rootName := r.input.Root.Name
	rootSel := r.selected[rootName]
	if rootSel == nil {
		return nil, fmt.Errorf("resolution lost the root %s", rootName)
	}
	rootEdges := map[string]int64{}
	if rootSel.pkg != nil {
		for _, requirement := range rootSel.pkg.Requires {
			if requirement.Weight != nil {
				rootEdges[requirement.Name] = *requirement.Weight
			}
		}
	}
	rootMap := map[string]int64{}
	if rootSel.pkg != nil {
		for name, weight := range rootSel.pkg.Weights {
			if _, onEdge := rootEdges[name]; onEdge {
				return nil, &Error{Diagnostic: DiagWeightsDuplicate, Name: name}
			}
			if _, inClosure := r.selected[name]; !inClosure {
				return nil, &Error{Diagnostic: DiagWeightUnknown, Name: name}
			}
			rootMap[name] = weight
		}
	}
	for name, weight := range rootEdges {
		rootMap[name] = weight
	}

	var warnings []Warning
	result := &Result{Members: map[string]Resolved{}}
	lock := &contextlock.Lock{Root: rootName}
	names := make([]string, 0, len(r.selected))
	for name := range r.selected {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sel := r.selected[name]
		member := contextlock.Member{Kind: sel.kind, Name: name, Source: sel.source, Directory: sel.directory,
			Commit: sel.commit, RequiredBy: []string{}, Overlay: sel.overlay != nil}
		if sel.state != nil {
			member.StateHash = sel.state.StateHash
		}
		if sel.hasVer {
			member.Version = sel.version.String()
		}
		requirers := map[string]bool{}
		var edgeWeights []RequirerWeight
		for _, c := range r.constraints[name] {
			if c.requirer == MachineRequirer {
				continue
			}
			requirerName := strings.SplitN(c.requirer, "@", 2)[0]
			requirers[requirerName] = true
			if c.requirement.Weight != nil {
				edgeWeights = append(edgeWeights, RequirerWeight{Requirer: c.requirer, Weight: *c.requirement.Weight})
			}
		}
		for requirer := range requirers {
			member.RequiredBy = append(member.RequiredBy, requirer)
		}
		sort.Strings(member.RequiredBy)
		// Environments §6: every closure member has one effective weight by
		// rules 1–3 in order — the manifest weight, the agreeing
		// direct-requirer edge weights, then the root weights map — whatever
		// the member kind. Overlays name context members only; the override
		// below is a no-op for any other kind.
		weight := int64(0)
		if sel.pkg != nil {
			weight = sel.pkg.Weight
		}
		sort.SliceStable(edgeWeights, func(i, j int) bool { return edgeWeights[i].Requirer < edgeWeights[j].Requirer })
		if len(edgeWeights) > 0 {
			agreed := true
			for _, edge := range edgeWeights[1:] {
				if edge.Weight != edgeWeights[0].Weight {
					agreed = false
				}
			}
			if !agreed {
				if _, named := rootMap[name]; !named {
					return nil, &Error{Diagnostic: DiagWeightConflict, Name: name, WeightRequirers: edgeWeights}
				}
				warnings = append(warnings, Warning{Diagnostic: DiagWeightConflict, Name: name, Requirers: edgeWeights})
			} else {
				weight = edgeWeights[0].Weight
			}
		}
		if mapped, named := rootMap[name]; named {
			weight = mapped
		}
		if sel.overlay != nil {
			if sel.overlay.Weight != nil {
				weight = *sel.overlay.Weight
			} else {
				weight = r.input.OverlayDefaultWeight
			}
		}
		member.Weight = weight
		lock.Members = append(lock.Members, member)
		result.Members[Key(sel.kind, name)] = Resolved{Kind: sel.kind, Name: name, Source: sel.source, Directory: sel.directory,
			Commit: sel.commit, StateHash: member.StateHash, Version: member.Version, Package: sel.pkg}
	}
	lock.Sort()
	hash, err := lock.Hash()
	if err != nil {
		return nil, err
	}
	result.Lock, result.LockHash, result.Warnings = lock, hash, warnings
	return result, nil
}
