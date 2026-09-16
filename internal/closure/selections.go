package closure

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/manifest"
)

// AcquireSelection is the acquisition boundary for a single expanded skill.
// It must freeze the selected package and return its validated Spec, Snapshot,
// and (for Git only) Repo, Resolved and canonical Identity. A local snapshot
// must leave Git fields empty, never put a snapshot digest in Commit. Lock,
// snapshot revalidation, audit and publication remain acquisition's obligations.
// The returned node is consumed by BuildExpanded and must not be reused.
type AcquireSelection func(manifest.Selection) (*Node, error)

// BuildExpanded resolves the full dependency closure after deterministic draft
// expansion. Acquisition is explicit so local bytes are never coerced through
// legacy Git resolution. Unchanged legacy declarations still use that lane.
// Callers must not publish these nodes using legacy marker/lock formats.
func BuildExpanded(opts Options, m *manifest.Manifest, expansion manifest.ExpansionOptions, acquire AcquireSelection, substitutions map[string]devsub.Substitution) ([]*Node, error) {
	members, err := manifest.Expand(m, expansion)
	if err != nil {
		return nil, err
	}
	queue := make([]pending, 0, len(members))
	commits := map[string]string{}
	for _, member := range members {
		decl := member.Decl
		item := pending{name: decl.Name, git: decl.Git, ref: decl.Ref, source: decl.Source, edge: Edge{Consumer: ProjectEdge, Mode: "full"}, chain: ProjectEdge + " -> " + decl.Name}
		if decl.Selector != nil {
			if acquire == nil {
				return nil, fmt.Errorf("source_selection_invalid: frozen package acquisition required for %s", decl.Name)
			}
			node, err := acquire(member)
			if err != nil {
				return nil, err
			}
			if node == nil || node.Name != decl.Name || node.Spec == nil || node.Snapshot == "" || node.Substituted != "" {
				return nil, fmt.Errorf("source_member_invalid: invalid acquired member %s", decl.Name)
			}
			source := m.Sources[decl.Selector.From]
			if source.Path != "" {
				if node.Resolved.Commit != "" || node.Identity != "" || node.Repo != "" {
					return nil, fmt.Errorf("source_member_invalid: local member %s has Git identity", decl.Name)
				}
			} else {
				commit := node.Resolved.Commit
				_, hexErr := hex.DecodeString(commit)
				if (len(commit) != 40 && len(commit) != 64) || hexErr != nil || strings.ToLower(commit) != commit || node.Identity != source.Identity || node.Repo == "" || node.Resolved.Kind != source.Ref.Kind || node.Resolved.Ref != source.Ref.Value {
					return nil, fmt.Errorf("source_member_invalid: acquired Git identity differs for %s", decl.Name)
				}
				if previous, ok := commits[decl.Selector.From]; ok && previous != commit {
					return nil, fmt.Errorf("source_name_conflict: alias %s resolved to multiple commits", decl.Selector.From)
				}
				commits[decl.Selector.From] = commit
			}
			node.Decl = decl
			node.Edges = nil
			node.Chains = nil
			item.selected = node
		}
		queue = append(queue, item)
	}
	nodes, err := buildQueue(opts, queue, substitutions, true)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
