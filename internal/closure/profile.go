package closure

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/protocoljson"
)

// ProfileRequirement is an admitted context-origin requirement. Context
// snapshots are checked by the profile manager before constructing these roots.
type ProfileRequirement struct {
	Consumer  string
	Name      string
	Git       string
	Directory string
	Ref       manifest.Ref
	Mode      string
	Commands  []string
}

// ProfileAcquire freezes one exact package through the caller's existing
// source/signer admission. It is used for transitive dependencies too; it must
// return a private immutable snapshot and validated skill metadata.
type ProfileAcquire func(name, source, directory string, ref manifest.Ref) (*Node, error)

// ProfilePlan owns no live surfaces. Nodes retain the shared closure's
// provider-first order and activation semantics for ordinary install staging.
type ProfilePlan struct {
	Lock  *contextlock.Lock
	Nodes []*Node
}

// BuildProfile bridges context roots to the ordinary shared closure. Direct
// declarations enter only after reachability is known; unrelated roots are
// never acquired, expanded, upgraded or removed (PSR01/03).
func BuildProfile(old *contextlock.Lock, roots []ProfileRequirement, direct []manifest.Decl, declarations map[string]any, acquire ProfileAcquire) (*ProfilePlan, error) {
	if old == nil || acquire == nil {
		return nil, fmt.Errorf("profile_skill_plan_invalid: missing lock or acquisition")
	}
	if err := old.Validate(); err != nil {
		return nil, err
	}
	opts := Options{}
	opts.profileAcquire = func(item pending) (*Node, error) {
		if item.ref.Kind != "tag" && item.ref.Kind != "revision" {
			return nil, fmt.Errorf("profile_skill_plan_invalid: skill ref must be exact")
		}
		directory := item.directory
		if directory == "" {
			directory = "."
		}
		node, err := acquire(item.name, item.git, directory, item.ref)
		if err != nil {
			return nil, err
		}
		if node == nil || node.Spec == nil || node.Snapshot == "" || node.Name != item.name || node.Directory != directory || node.Substituted != "" || node.Identity == "" {
			return nil, fmt.Errorf("profile_skill_plan_invalid: acquired package disagrees")
		}
		canonical, err := identity.Parse(item.git)
		if err != nil || canonical != node.Identity {
			return nil, fmt.Errorf("profile_skill_plan_invalid: source disagreement")
		}
		if item.ref.Kind == "revision" && node.Resolved.Commit != item.ref.Value {
			return nil, fmt.Errorf("profile_skill_pin_changed: revision differs")
		}
		if node.Resolved.Kind != item.ref.Kind || node.Resolved.Ref != item.ref.Value {
			return nil, fmt.Errorf("profile_skill_plan_invalid: missing admitted ref evidence")
		}
		copyNode := *node
		copyNode.Edges = nil
		copyNode.Chains = nil
		return &copyNode, nil
	}
	queue := []pending{}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Consumer != roots[j].Consumer {
			return roots[i].Consumer < roots[j].Consumer
		}
		return roots[i].Name < roots[j].Name
	})
	for _, root := range roots {
		if _, ok := old.Find(contextlock.KindContext, root.Consumer); !ok {
			return nil, fmt.Errorf("profile_skill_plan_invalid: unknown context consumer")
		}
		mode := root.Mode
		if mode == "" {
			mode = "full"
		}
		queue = append(queue, pending{name: root.Name, git: root.Git, ref: root.Ref, directory: root.Directory, edge: Edge{Consumer: root.Consumer, ConsumerKind: "context", Mode: mode, Commands: root.Commands, Ref: root.Ref}, chain: root.Consumer + " -> " + root.Name})
	}
	nodes, err := buildQueue(opts, queue, nil, false)
	if err != nil {
		if strings.HasPrefix(err.Error(), "profile_skill_") {
			return nil, err
		}
		return nil, fmt.Errorf("profile_skill_plan_invalid: %w", err)
	}
	byName := map[string]*Node{}
	for _, node := range nodes {
		byName[node.Name] = node
	}
	plan := &contextlock.SkillPlan{SchemaVersion: 1, Nodes: []contextlock.SkillPin{}, Edges: []contextlock.SkillEdge{}, DirectConstraints: []contextlock.DirectConstraint{}}
	for _, decl := range direct {
		node, overlap := byName[decl.Name]
		if !overlap {
			continue
		}
		item := pending{name: decl.Name, git: decl.Git, ref: decl.Ref, directory: "."}
		if decl.Selector != nil {
			return nil, fmt.Errorf("profile_skill_pin_conflict: overlapping direct selector is unsupported")
		}
		other, err := opts.profileAcquire(item)
		if err != nil {
			return nil, fmt.Errorf("profile_skill_pin_conflict: %w", err)
		}
		if other.Identity != node.Identity || other.Resolved.Commit != node.Resolved.Commit || other.Directory != node.Directory {
			return nil, fmt.Errorf("profile_skill_pin_conflict: profile consumers %v request %s %s -> %s@%s directory %s; machine consumer %s requests %s %s -> %s@%s directory %s", node.Consumers(), node.Resolved.Kind, node.Resolved.Ref, node.Identity, node.Resolved.Commit, node.Directory, decl.Name, decl.Ref.Kind, decl.Ref.Value, other.Identity, other.Resolved.Commit, other.Directory)
		}
		object, ok := declarations[decl.Name]
		if !ok {
			return nil, fmt.Errorf("profile_skill_plan_invalid: missing original declaration")
		}
		payload, err := protocoljson.MarshalCanonical(object)
		if err != nil {
			return nil, err
		}
		plan.DirectConstraints = append(plan.DirectConstraints, contextlock.DirectConstraint{Name: node.Name, Source: node.Identity, Commit: node.Resolved.Commit, Directory: node.Directory, Ref: contextlock.SkillRef{Kind: decl.Ref.Kind, Value: decl.Ref.Value}, DeclarationSHA256: contextlock.HashBytes(payload)})
		node.Edges = append(node.Edges, Edge{Consumer: node.Name, ConsumerKind: "machine", Mode: "full", Ref: decl.Ref})
	}
	for _, node := range nodes {
		digest, err := hashing.ContentSHA256WithVersion(node.Snapshot, map[string]bool{}, hashing.VersionV2)
		if err != nil {
			return nil, err
		}
		plan.Nodes = append(plan.Nodes, contextlock.SkillPin{Kind: "skill", Name: node.Name, Source: node.Identity, Commit: node.Resolved.Commit, Directory: node.Directory, HashVersion: 2, ContentSHA256: strings.TrimPrefix(digest, "sha256:")})
		for _, edge := range node.Edges {
			kind := edge.ConsumerKind
			if kind == "" {
				kind = "skill"
			}
			mode := edge.Mode
			if mode == "" {
				mode = "full"
			}
			for _, command := range edge.Commands {
				exported, ok := node.Spec.Commands[command]
				if !ok || (exported.Type != "script" && exported.Type != "build") {
					return nil, fmt.Errorf("profile_skill_command_invalid: %s does not export %s", node.Name, command)
				}
			}
			plan.Edges = append(plan.Edges, contextlock.SkillEdge{Consumer: contextlock.SkillConsumer{Kind: kind, Name: edge.Consumer}, Provider: node.Name, Ref: contextlock.SkillRef{Kind: edge.Ref.Kind, Value: edge.Ref.Value}, Mode: mode, Commands: edge.Commands})
		}
	}
	sort.Slice(plan.Nodes, func(i, j int) bool { return plan.Nodes[i].Name < plan.Nodes[j].Name })
	sort.Slice(plan.Edges, func(i, j int) bool {
		return contextlock.SkillEdgeKey(plan.Edges[i]) < contextlock.SkillEdgeKey(plan.Edges[j])
	})
	sort.Slice(plan.DirectConstraints, func(i, j int) bool { return plan.DirectConstraints[i].Name < plan.DirectConstraints[j].Name })
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	next := *old
	next.Members = append([]contextlock.Member(nil), old.Members...)
	next.SchemaVersion = 5
	next.HashVersion = 2
	if next.Precedence.Winner == "" {
		next.Precedence = contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}
	}
	// Retained members are protected old-lock entries, never reconstructed from
	// today's direct declarations. Only an absent legacy root directory changes.
	members := []contextlock.Member{}
	for _, member := range next.Members {
		if _, selected := byName[member.Name]; member.Kind == "skill" && selected {
			continue
		}
		if member.Kind == "skill" && member.Directory == "" {
			member.Directory = "."
		}
		members = append(members, member)
	}
	for _, pin := range plan.Nodes {
		required := []string{}
		seen := map[string]bool{}
		for _, edge := range plan.Edges {
			if edge.Provider == pin.Name && edge.Consumer.Kind != "machine" && !seen[edge.Consumer.Name] {
				required = append(required, edge.Consumer.Name)
				seen[edge.Consumer.Name] = true
			}
		}
		sort.Strings(required)
		member := contextlock.Member{Kind: "skill", Name: pin.Name, Source: pin.Source, Commit: pin.Commit, Directory: pin.Directory, RequiredBy: required}
		if previous, ok := old.Find("skill", pin.Name); ok {
			member.Weight = previous.Weight
			if previous.Commit == pin.Commit && previous.Source == pin.Source && (previous.Directory == pin.Directory || previous.Directory == "" && pin.Directory == ".") {
				member.Version = previous.Version
			}
		}
		members = append(members, member)
	}
	next.Members = members
	next.Sort()
	next.SkillPlan = plan
	next.SkillPlanSHA256, err = plan.Hash()
	if err != nil {
		return nil, err
	}
	if err := next.Validate(); err != nil {
		return nil, err
	}
	return &ProfilePlan{Lock: &next, Nodes: nodes}, nil
}

// VerifyProfilePlan compares a pinned recomputation with the stored portable
// plan. Rehashing changed requirements or declarations does not authorize them.
func VerifyProfilePlan(stored, recomputed *contextlock.Lock) error {
	if stored.SchemaVersion != 5 {
		return fmt.Errorf("profile_skill_plan_required: explicit update required")
	}
	if err := stored.Validate(); err != nil {
		return err
	}
	if !reflect.DeepEqual(stored.SkillPlan, recomputed.SkillPlan) {
		return fmt.Errorf("profile_skill_plan_invalid: pinned graph or direct provenance changed")
	}
	return nil
}

func unifyProfile(node, other *Node) error {
	if node.Identity != other.Identity || node.Directory != other.Directory || node.Resolved.Commit != other.Resolved.Commit {
		return fmt.Errorf("profile_skill_pin_conflict: incompatible pins for %s", node.Name)
	}
	if !strings.EqualFold(node.Name, other.Name) || node.Name != other.Name {
		return fmt.Errorf("profile_skill_plan_invalid: name disagreement")
	}
	return nil
}
