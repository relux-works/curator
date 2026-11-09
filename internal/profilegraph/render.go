package profilegraph

import (
	"bytes"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

func (b *build) render() error {
	for _, env := range b.g.Scope.Environments {
		adapter, err := envregistry.ByID(env)
		if err != nil {
			return err
		}
		form := b.req.Projection.Forms[env]
		if form == "" {
			form = adapter.DefaultForm
		}
		switch form {
		case contextmaterialize.FormReferenced:
			files, written, err := contextmaterialize.TraceReferenced(b.lock, b.lockHash, b.precedence(), env, b.material)
			if err != nil {
				return err
			}
			if written {
				paths := make([]string, 0, len(files))
				for p := range files {
					paths = append(paths, p)
				}
				sort.Strings(paths)
				for _, p := range paths {
					if err := b.document(env, p, form, files[p]); err != nil {
						return err
					}
				}
			}
		case contextmaterialize.FormMonolithic:
			rendered, written, err := contextmaterialize.TraceMonolithic(b.lock, b.lockHash, b.precedence(), env, b.material)
			if err != nil {
				return err
			}
			if written {
				if err := b.document(env, adapter.RootTarget, form, rendered); err != nil {
					return err
				}
			}
		default:
			return refusal("graph_invalid_input")
		}
		if env != envregistry.Muse && len(adapter.SystemPrompt) > 0 {
			rendered, written, err := contextmaterialize.TraceSystemPrompt(b.lock, b.precedence(), env, b.material, b.req.Projection.Admission)
			if err != nil {
				return err
			}
			if written {
				if err := b.document(env, ".agent-context/system-prompt.md", "system-prompt", rendered); err != nil {
					return err
				}
				if env == envregistry.Pi {
					rel := ""
					switch b.req.Projection.SystemPromptFiles {
					case "append":
						rel = "APPEND_SYSTEM.md"
					case "replace":
						rel = "SYSTEM.md"
					}
					if rel != "" {
						if err := b.document(env, rel, "system-prompt", rendered); err != nil {
							return err
						}
					}
				}

			}
		}
	}
	for _, n := range b.g.Nodes {
		switch n.Kind {
		case "instruction-module", "section", "quote-block", "generated":
			for i, load := range n.Loads {
				emitted := false
				for _, em := range n.Emissions {
					for _, d := range b.g.Documents {
						if em.Document == d.Node && load.Environment == d.Environment {
							emitted = true
						}
					}
				}
				if emitted {
					n.Loads[i].Class = "always"
					n.Loads[i].State = "emitted"
					n.Loads[i].Basis = "rendering-rule"
				} else {
					n.Loads[i].State = "excluded"
					n.Loads[i].Basis = "rendering-rule"
				}
			}
		}
	}
	return b.err
}
func (b *build) document(env, rel, form string, rendered contextmaterialize.Rendered) error {
	if len(b.g.Documents) >= b.r.Limits.Documents {
		return refusal("graph_limit_exceeded")
	}
	if !cleanRelative(rel) || denied(rel) {
		return refusal("graph_path_denied")
	}
	origin := "rendered"
	if b.req.Source == "" {
		full := filepath.Join(envprofile.ManagedHomeDir(b.req.Home, b.req.Profile, env), filepath.FromSlash(rel))
		p, err := filepath.Rel(b.req.Home, full)
		if err != nil {
			return err
		}
		raw, err := b.r.Read(b.req.Home, "managed", filepath.ToSlash(p))
		if err == nil {
			if !bytes.Equal(raw, rendered.Bytes) {
				return refusal("graph_document_drift")
			}
			origin = "installed"
		} else if !isMissing(err) {
			return err
		}
	}
	n := b.add("emitted-document", env+" document", nil, nil, env, rel)
	if b.err != nil {
		return b.err
	}
	n.Content = measured(rendered.Bytes)
	d := Document{Node: n.ID, Environment: env, Root: "environment/" + env, Path: rel, Form: form, Origin: origin, Bytes: len(rendered.Bytes), Lines: lines(rendered.Bytes), SHA256: hash(rendered.Bytes), Coverage: []Coverage{}}
	projections := map[string][]Range{}
	if len(rendered.Bytes) > 0 {
		projections[n.ID] = []Range{{0, len(rendered.Bytes)}}
		projections[b.profile.ID] = []Range{{0, len(rendered.Bytes)}}
	}
	ordinal := map[string]int{}
	offset := 0
	for _, span := range rendered.Spans {
		if span.Start != offset || span.End <= span.Start || span.End > len(rendered.Bytes) {
			return refusal("graph_invalid_input")
		}
		offset = span.End
		var leaf *Node
		if span.Kind == "module" {
			leaf = b.modules[span.Package+"/"+itoa(span.Slot)]
			if leaf == nil {
				return refusal("graph_invalid_input")
			}
			projections[leaf.ID] = append(projections[leaf.ID], Range{span.Start, span.End})
			if leaf.Layer != nil {
				projections[*leaf.Layer] = append(projections[*leaf.Layer], Range{span.Start, span.End})
			}
			for _, child := range b.g.Nodes {
				if (child.Kind == "section" || child.Kind == "quote-block") && len(child.Identity) > 2 && child.Identity[2] == leaf.ID {
					for _, r := range child.SourceRanges {
						projections[child.ID] = append(projections[child.ID], Range{span.Start + r.Start, span.Start + r.End})
					}
				}
			}
		} else {
			i := ordinal[span.Kind]
			ordinal[span.Kind]++
			leaf = b.add("generated", env+" "+span.Kind+" "+itoa(i), nil, nil, n.ID, span.Kind, itoa(i))
			if b.err != nil {
				return b.err
			}
			leaf.Content = measured(rendered.Bytes[span.Start:span.End])
			projections[leaf.ID] = []Range{{span.Start, span.End}}
		}
		d.Coverage = append(d.Coverage, Coverage{leaf.ID, Range{span.Start, span.End}})
		b.edge(leaf.ID, n.ID, "contributes", len(d.Coverage)-1, nil)
	}
	if offset != len(rendered.Bytes) {
		return refusal("graph_invalid_input")
	}
	for id, ranges := range projections {
		sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })
		merged := []Range{}
		for _, r := range ranges {
			if r.End > len(rendered.Bytes) || r.Start < 0 || r.Start >= r.End {
				return refusal("graph_invalid_input")
			}
			if len(merged) > 0 && r.Start <= merged[len(merged)-1].End {
				merged[len(merged)-1].End = max(merged[len(merged)-1].End, r.End)
			} else {
				merged = append(merged, r)
			}
		}
		var data []byte
		for _, r := range merged {
			data = append(data, rendered.Bytes[r.Start:r.End]...)
		}
		b.nodes[id].Emissions = append(b.nodes[id].Emissions, Emission{n.ID, merged, len(data), lines(data), hash(data)})
	}
	for i := range n.Loads {
		if n.Loads[i].Environment == env {
			n.Loads[i].Class = "always"
			n.Loads[i].State = "emitted"
			n.Loads[i].Basis = "rendering-rule"
		}
	}
	b.edge(b.profile.ID, n.ID, "emits", len(b.g.Documents), nil)
	b.g.Documents = append(b.g.Documents, d)
	return b.err
}

// Text is a deterministic view of the same graph and exact document totals.
func Text(g *Graph, limit int) ([]byte, error) {
	out := &boundedBuffer{limit: limit}
	_, _ = out.WriteString(SchemaVersion + " " + g.Scope.Profile + " (" + g.Scope.Mode + ")\n")
	record := func(label string, value any) {
		_, _ = out.WriteString(label + ": ")
		encodeValue(out, reflect.ValueOf(value), 0)
		_, _ = out.WriteString("\n")
	}
	record("generator", g.Generator)
	record("scope", g.Scope)
	record("inputs_digest", g.InputsDigest)
	for _, in := range g.Inputs {
		record("input", in)
	}
	for _, n := range g.Nodes {
		record("node", n)
	}
	for _, e := range g.Edges {
		record("edge", e)
	}
	for _, d := range g.Documents {
		record("document", d)
	}
	for _, d := range g.Diagnostics {
		record("diagnostic", d)
	}
	if out.err != nil {
		return nil, out.err
	}
	return out.Bytes(), nil
}
