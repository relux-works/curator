// Package closure resolves the transitive dependency closure of a project
// manifest (Spec §8.3, §8.4).
//
// Direct declarations enter as full-mode requirements rooted in the
// synthetic consumer "<project>". Within one closure a skill name resolves
// to exactly one commit and one canonical source identity; providers precede
// consumers in the returned order, deterministically.
package closure

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/snapshot"
)

// ProjectEdge is the synthetic consumer name for direct Skillfile entries.
const ProjectEdge = "<project>"

// Edge is one activation edge: who requires the node and how.
type Edge struct {
	Consumer string
	Mode     string // full | runtime | context
	Commands []string
}

// Node is one resolved closure member.
type Node struct {
	Name        string
	Decl        manifest.Decl
	Resolved    gitops.ResolvedRef
	Repo        string
	Snapshot    string
	Spec        *skillspec.Spec
	Identity    string // canonical identity, "" for local sources
	Chains      []string
	Substituted string // devsub description, "" when not substituted
	Edges       []Edge
}

// ContextActive reports whether any edge activates the node's context.
func (n *Node) ContextActive() bool {
	for _, edge := range n.Edges {
		if edge.Mode == "full" || edge.Mode == "context" {
			return true
		}
	}
	return false
}

// ActiveCommands returns the active command set under the node's edges
// (Spec §8.3): full activates every exported script command; runtime edges
// activate their narrowing, or everything when unnarrowed.
func (n *Node) ActiveCommands() map[string]bool {
	exported := map[string]bool{}
	for name, command := range n.Spec.Commands {
		if command.Type == "script" {
			exported[name] = true
		}
	}
	for _, edge := range n.Edges {
		if edge.Mode == "full" {
			return exported
		}
	}
	active := map[string]bool{}
	for _, edge := range n.Edges {
		if edge.Mode != "runtime" {
			continue
		}
		if len(edge.Commands) == 0 {
			for name := range exported {
				active[name] = true
			}
			continue
		}
		for _, name := range edge.Commands {
			active[name] = true
		}
	}
	return active
}

// Consumers returns the distinct consumers in first-seen order.
func (n *Node) Consumers() []string {
	var seen []string
	has := map[string]bool{}
	for _, edge := range n.Edges {
		if !has[edge.Consumer] {
			has[edge.Consumer] = true
			seen = append(seen, edge.Consumer)
		}
	}
	return seen
}

// Options carries the environment of a resolution run.
type Options struct {
	SkillsRoot     string
	Home           string // machine home for caches
	AllowedSources []string
	// DevRoot is where git substitutions clone (outside SkillsRoot so a
	// substitution never shadows the declared source repository).
	DevRoot string
}

type pending struct {
	name   string
	git    string
	ref    manifest.Ref
	source string
	edge   Edge
	chain  string
}

// Build expands the manifest into an ordered closure.
func Build(opts Options, projectManifest *manifest.Manifest, substitutions map[string]devsub.Substitution) ([]*Node, error) {
	nodes := map[string]*Node{}
	var queue []pending
	for _, decl := range projectManifest.Skills {
		queue = append(queue, pending{
			name:   decl.Name,
			git:    decl.Git,
			ref:    decl.Ref,
			source: decl.Source,
			edge:   Edge{Consumer: ProjectEdge, Mode: "full"},
			chain:  ProjectEdge + " -> " + decl.Name,
		})
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		node, exists := nodes[item.name]
		if !exists {
			resolved, err := resolveNode(opts, item, substitutions)
			if err != nil {
				return nil, err
			}
			nodes[item.name] = resolved
			node = resolved
			names := make([]string, 0, len(node.Spec.Requirements))
			for name := range node.Spec.Requirements {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				requirement := node.Spec.Requirements[name]
				queue = append(queue, pending{
					name:   requirement.Name,
					git:    requirement.Git,
					ref:    manifest.Ref{Kind: requirement.RefKind, Value: requirement.RefValue},
					source: requirement.Name,
					edge:   Edge{Consumer: item.name, Mode: requirement.Mode, Commands: requirement.Commands},
					chain:  item.chain + " -> " + requirement.Name,
				})
			}
		} else if err := unify(node, item); err != nil {
			return nil, err
		}
		node.Edges = append(node.Edges, item.edge)
		node.Chains = append(node.Chains, item.chain)
	}

	if err := validateRequirementCommands(nodes); err != nil {
		return nil, err
	}
	return topologicalOrder(nodes)
}

// DetectActiveCommandCollisions enforces one owner per active command name
// across the closure (Spec §8.4).
func DetectActiveCommandCollisions(nodes []*Node) error {
	owners := map[string]string{}
	for _, node := range nodes {
		var commands []string
		for name := range node.ActiveCommands() {
			commands = append(commands, name)
		}
		sort.Strings(commands)
		for _, command := range commands {
			if previous, taken := owners[command]; taken {
				return fmt.Errorf("command collision for %q: exported by %s and %s", command, previous, node.Name)
			}
			owners[command] = node.Name
		}
	}
	return nil
}

// unify checks that a later requirement of an existing node names the same
// repository and the same commit (Spec §8.3). A substituted node skips
// unification: the substitution replaces every requirement of that name.
func unify(node *Node, item pending) error {
	if node.Substituted != "" {
		return nil
	}
	if item.git != "" {
		id, err := identity.Parse(item.git)
		if err != nil {
			return fmt.Errorf("invalid source for %s (via %s): %w", item.name, item.chain, err)
		}
		if id != "" {
			if node.Identity == "" {
				node.Identity = id
			} else if node.Identity != id {
				return fmt.Errorf(
					"source conflict for %s: %s (via %s) and %s (via %s) name different repositories",
					node.Name, node.Identity, node.Chains[0], id, item.chain)
			}
		}
	}
	if item.ref.Kind == node.Resolved.Kind && item.ref.Value == node.Resolved.Ref {
		return nil
	}
	other, err := gitops.Resolve(node.Repo, item.ref.Kind, item.ref.Value)
	if err != nil {
		return fmt.Errorf("cannot resolve %s %q for %s (via %s): %w", item.ref.Kind, item.ref.Value, node.Name, item.chain, err)
	}
	if other.Commit != node.Resolved.Commit {
		return fmt.Errorf(
			"version conflict for %s: %s %s -> %s (via %s) and %s %s -> %s (via %s); align the requirement refs at their declarations",
			node.Name, node.Resolved.Kind, node.Resolved.Ref, short(node.Resolved.Commit), node.Chains[0],
			item.ref.Kind, item.ref.Value, short(other.Commit), item.chain)
	}
	return nil
}

func resolveNode(opts Options, item pending, substitutions map[string]devsub.Substitution) (*Node, error) {
	var repo string
	var resolved gitops.ResolvedRef
	substituted := ""

	if substitution, present := substitutions[item.name]; present {
		if substitution.Path != "" {
			repo = substitution.Path
			if err := gitops.EnsureRepo(repo); err != nil {
				return nil, fmt.Errorf("substitution for %s points to %s, which is not a git repository", item.name, repo)
			}
			ref, err := gitops.Resolve(repo, "revision", "HEAD")
			if err != nil {
				return nil, err
			}
			resolved = ref
		} else {
			if err := gateSource(opts, item.name, substitution.Git, item.chain); err != nil {
				return nil, err
			}
			devRepo, err := ensureDevRepo(opts, item.name, substitution.Git)
			if err != nil {
				return nil, err
			}
			repo = devRepo
			ref, err := gitops.Resolve(repo, substitution.RefKind, substitution.RefValue)
			if err != nil {
				return nil, err
			}
			resolved = ref
		}
		substituted = substitution.Describe()
	} else {
		localRepo, err := ensureRepo(opts, item)
		if err != nil {
			return nil, err
		}
		repo = localRepo
		ref, err := gitops.Resolve(repo, item.ref.Kind, item.ref.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve %s %q for %s (via %s): %w", item.ref.Kind, item.ref.Value, item.name, item.chain, err)
		}
		resolved = ref
	}

	snap, err := snapshot.Get(opts.Home, item.source, repo, resolved.Commit)
	if err != nil {
		return nil, err
	}
	if gitops.HasSubmodules(snap) {
		return nil, fmt.Errorf("submodules are unsupported: %s", item.source)
	}
	spec, err := skillspec.Load(snap)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", item.name, err)
	}

	id := ""
	if item.git != "" {
		id, err = identity.Parse(item.git)
		if err != nil {
			return nil, fmt.Errorf("invalid source for %s (via %s): %w", item.name, item.chain, err)
		}
	}
	return &Node{
		Name:        item.name,
		Decl:        manifest.Decl{Name: item.name, Source: item.source, Ref: manifest.Ref{Kind: resolved.Kind, Value: resolved.Ref}, Git: item.git},
		Resolved:    resolved,
		Repo:        repo,
		Snapshot:    snap,
		Spec:        spec,
		Identity:    id,
		Substituted: substituted,
	}, nil
}

func ensureRepo(opts Options, item pending) (string, error) {
	repo := filepath.Join(opts.SkillsRoot, filepath.FromSlash(item.source))
	if _, err := os.Stat(repo); err == nil {
		if err := gitops.EnsureRepo(repo); err != nil {
			return "", fmt.Errorf("local skill path exists but is not a git repository: %s", repo)
		}
		return repo, nil
	}
	if item.git == "" {
		return "", fmt.Errorf("skill repository not found for %s: %s (via %s)", item.name, repo, item.chain)
	}
	if err := gateSource(opts, item.name, item.git, item.chain); err != nil {
		return "", err
	}
	if err := gitops.Clone(item.git, repo); err != nil {
		return "", fmt.Errorf("failed to clone %s from %s: %w", item.name, item.git, err)
	}
	return repo, nil
}

func ensureDevRepo(opts Options, name, gitURL string) (string, error) {
	root := opts.DevRoot
	if root == "" {
		root = filepath.Join(opts.Home, "dev")
	}
	repo := filepath.Join(root, name)
	if _, err := os.Stat(filepath.Join(repo, ".git")); err == nil {
		if err := gitops.Fetch(repo); err != nil {
			return "", err
		}
		return repo, nil
	}
	if err := gitops.Clone(gitURL, repo); err != nil {
		return "", err
	}
	return repo, nil
}

// gateSource applies the machine allowlist before any network clone
// (Spec §8.2).
func gateSource(opts Options, name, gitURL, chain string) error {
	id, err := identity.Parse(gitURL)
	if err != nil {
		return fmt.Errorf("invalid source for %s: %s; %v; required via %s", name, gitURL, err, chain)
	}
	if identity.Allowed(id, opts.AllowedSources) {
		return nil
	}
	label := id
	if label == "" {
		label = "unknown"
	}
	return fmt.Errorf(
		"source not allowed for %s: %s (identity %s); allowed prefixes: %s; required via %s",
		name, gitURL, label, joinComma(opts.AllowedSources), chain)
}

// validateRequirementCommands checks that command narrowing names script
// commands the provider actually exports (Spec §8.3).
func validateRequirementCommands(nodes map[string]*Node) error {
	var problems []string
	names := make([]string, 0, len(nodes))
	for name := range nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		node := nodes[name]
		reqNames := make([]string, 0, len(node.Spec.Requirements))
		for reqName := range node.Spec.Requirements {
			reqNames = append(reqNames, reqName)
		}
		sort.Strings(reqNames)
		for _, reqName := range reqNames {
			requirement := node.Spec.Requirements[reqName]
			provider, present := nodes[requirement.Name]
			if !present {
				continue
			}
			for _, command := range requirement.Commands {
				provided, exported := provider.Spec.Commands[command]
				if !exported || provided.Type != "script" {
					problems = append(problems, fmt.Sprintf(
						"requirement %s -> %s names command %q, but %s does not export a script command named %q",
						node.Name, requirement.Name, command, requirement.Name, command))
				}
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s", joinSemicolon(problems))
	}
	return nil
}

// topologicalOrder returns providers before consumers, deterministically
// (Spec §8.3). A cycle is an error naming its members.
func topologicalOrder(nodes map[string]*Node) ([]*Node, error) {
	dependents := map[string]map[string]bool{}
	indegree := map[string]int{}
	for name := range nodes {
		dependents[name] = map[string]bool{}
		indegree[name] = 0
	}
	for name, node := range nodes {
		for _, edge := range node.Edges {
			if edge.Consumer == ProjectEdge {
				continue
			}
			if _, known := nodes[edge.Consumer]; !known {
				continue
			}
			if !dependents[name][edge.Consumer] {
				dependents[name][edge.Consumer] = true
				indegree[edge.Consumer]++
			}
		}
	}
	var ready []string
	for name, degree := range indegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)
	var ordered []*Node
	for len(ready) > 0 {
		name := ready[0]
		ready = ready[1:]
		ordered = append(ordered, nodes[name])
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
	if len(ordered) != len(nodes) {
		var remaining []string
		placed := map[string]bool{}
		for _, node := range ordered {
			placed[node.Name] = true
		}
		for name := range nodes {
			if !placed[name] {
				remaining = append(remaining, name)
			}
		}
		sort.Strings(remaining)
		return nil, fmt.Errorf("dependency cycle between skills: %s", joinComma(remaining))
	}
	return ordered, nil
}

func short(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

func joinComma(values []string) string {
	out := ""
	for index, value := range values {
		if index > 0 {
			out += ", "
		}
		out += value
	}
	return out
}

func joinSemicolon(values []string) string {
	out := ""
	for index, value := range values {
		if index > 0 {
			out += "; "
		}
		out += value
	}
	return out
}
