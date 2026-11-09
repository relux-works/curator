package profilegraph

import (
	"bytes"
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/protocoljson"
)

func itoa(n int) string { return strconv.Itoa(n) }
func (b *build) add(kind, label string, source *Source, layer *string, coordinates ...string) *Node {
	tuple := append([]string{kind, b.req.Profile}, coordinates...)
	n := &Node{ID: tupleID(tuple), Identity: tuple, Kind: kind, Label: label, Source: source, Layer: layer, Content: Content{Status: "not-applicable"}, Emissions: []Emission{}, Loads: []Load{}}
	if _, exists := b.nodes[n.ID]; exists {
		b.err = refusal("graph_invalid_input")
		return n
	}
	if len(b.g.Nodes) >= b.r.Limits.Nodes {
		b.err = refusal("graph_limit_exceeded")
		return n
	}
	for _, env := range b.g.Scope.Environments {
		n.Loads = append(n.Loads, Load{Environment: env, Class: "UNKNOWN", State: "UNKNOWN", Basis: "UNKNOWN"})
	}
	b.nodes[n.ID] = n
	b.g.Nodes = append(b.g.Nodes, n)
	return n
}
func (b *build) edge(from, to, kind string, order int, req *Requirement) {
	if len(b.g.Edges) >= b.r.Limits.Edges {
		b.err = refusal("graph_limit_exceeded")
		return
	}
	b.g.Edges = append(b.g.Edges, Edge{from, to, kind, order, req})
}
func (b *build) provenance(p packageInput, rel *string) *Source {
	m := p.member
	locator := m.Source
	if locator == "" || !identity.ValidCanonical(locator) && identity.ClassifySource(locator) == identity.SourcePath {
		locator = p.label
	} else if identity.ClassifySource(locator) != identity.SourcePath {
		canonical, err := (offlineSource{b}).Identity("", "", locator)
		if err != nil {
			b.err = err
		} else {
			locator = canonical
		}
	}
	s := &Source{Locator: locator, Directory: m.Directory, Path: rel, Ownership: "UNKNOWN", Revision: Revision{Kind: "UNKNOWN"}}
	if m.Version != "" {
		s.Version = ptr(m.Version)
	}
	if m.Commit != "" {
		s.Revision = Revision{"commit", m.Commit}
	} else if m.StateHash != "" {
		s.Revision = Revision{"state-sha256", m.StateHash}
	}
	return s
}
func requirement(r contextpkg.Requirement) Requirement {
	if r.Range != "" {
		return Requirement{"range", r.Range}
	}
	if r.Tag != "" {
		return Requirement{"tag", r.Tag}
	}
	return Requirement{"revision", r.Revision}
}
func (b *build) composition(m contextlock.Member, order int) Composition {
	req := Requirement{"revision", m.Commit}
	declaration := "root"
	if m.StateHash != "" {
		req = Requirement{"path", b.packages[m.Name].label}
	}
	if m.Name == b.lock.Root && b.req.Source == "" {
		r := b.source.Req
		if r.Range != "" {
			req = Requirement{"range", r.Range}
		} else if r.Tag != "" {
			req = Requirement{"tag", r.Tag}
		} else if r.Revision != "" {
			req = Requirement{"revision", r.Revision}
		}
	}
	if m.Overlay {
		declaration = "overlays/UNKNOWN"
		for i, o := range b.req.Projection.Overlays {
			canonical := o.Source
			if identity.ClassifySource(o.Source) != identity.SourcePath {
				var err error
				canonical, err = (offlineSource{b}).Identity("", "", o.Source)
				if err != nil {
					b.err = err
					continue
				}
			}
			if canonical != m.Source || o.Directory != m.Directory || o.Weight == nil || *o.Weight != m.Weight {
				continue
			}
			if o.Revision != "" && o.Revision != m.Commit {
				continue
			}
			declaration = "overlays/" + itoa(i)
			if o.Range != "" {
				req = Requirement{"range", o.Range}
			} else if o.Tag != "" {
				req = Requirement{"tag", o.Tag}
			} else if o.Revision != "" {
				req = Requirement{"revision", o.Revision}
			}
			break
		}
		if declaration == "overlays/UNKNOWN" {
			b.err = refusal("graph_invalid_input")
		}
	} else if m.Name != b.lock.Root {
		for _, parent := range m.RequiredBy {
			if manifest := b.packages[parent].manifest; manifest != nil {
				if r, ok := manifest.Contexts[m.Name]; ok {
					req = requirement(r)
					declaration = "requires.contexts." + m.Name
					break
				}
			}
		}
	}
	return Composition{req, m.Weight, order, declaration}
}
func (b *build) graphNodes() error {
	b.profile = b.add("profile", b.req.Profile, nil, nil)
	precedence := b.precedence()
	order, err := contextmaterialize.EmittedOrder(b.lock, precedence)
	if err != nil {
		return refusal("graph_cycle")
	}
	for i, m := range order {
		p := b.packages[m.Name]
		kind := "context-package"
		if m.Name == b.lock.Root {
			kind = "root-package"
		} else if m.Overlay {
			kind = "overlay"
		} else if p.manifest != nil && len(p.manifest.Contexts) > 0 {
			kind = "profile-package"
		}
		s := b.provenance(p, nil)
		coords := []string{s.Locator, s.Directory, s.Revision.Kind, s.Revision.Value}
		composition := b.composition(m, i)
		if m.Overlay {
			coords = append(coords, composition.Declaration, strconv.FormatInt(m.Weight, 10))
		}
		n := b.add(kind, m.Name, s, nil, coords...)
		n.Layer = ptr(n.ID)
		n.Composition = &composition
		b.layers[m.Name] = n
		if m.Name == b.lock.Root {
			b.edge(b.profile.ID, n.ID, "root", 0, &composition.Requirement)
		} else if m.Overlay {
			b.edge(b.profile.ID, n.ID, "overlay", i, &composition.Requirement)
		}
	}
	for _, m := range b.lock.Members {
		if m.Kind == contextlock.KindSkill {
			p := b.packages[m.Name]
			s := b.provenance(p, nil)
			var layer *string
			if len(m.RequiredBy) == 1 && b.layers[m.RequiredBy[0]] != nil {
				layer = ptr(b.layers[m.RequiredBy[0]].ID)
			}
			n := b.add("skill", m.Name, s, layer, s.Locator, s.Directory, s.Revision.Kind, s.Revision.Value)
			n.Skill = &Skill{"lock", "UNKNOWN"}
			b.layers[m.Name] = n
		}
	}
	for _, m := range b.lock.Members {
		p := b.packages[m.Name]
		layer := b.layers[m.Name]
		if layer == nil {
			continue
		}
		if p.manifest == nil {
			if m.Kind == contextlock.KindSkill {
				if err := b.skill(p, layer); err != nil {
					return err
				}
			}
			continue
		}
		for slot, module := range b.material[m.Name].Modules {
			rel := path.Join(contextpkg.ContextDir, module.Path)
			n := b.add("instruction-module", module.Path, b.provenance(p, &rel), ptr(layer.ID), layer.ID, itoa(slot), rel)
			n.Content = measured(module.Bytes)
			if module.Environments != nil {
				condition := "environments=" + strings.Join(module.Environments, ",")
				for i := range n.Loads {
					n.Loads[i].Condition = ptr(condition)
				}
			}
			b.modules[m.Name+"/"+itoa(slot)] = n
			b.edge(layer.ID, n.ID, "contains", slot, nil)
			if err := b.markdown(n, module.Bytes); err != nil {
				return err
			}
		}
		reqs := []struct {
			kind   string
			values map[string]contextpkg.Requirement
		}{{"requires-context", p.manifest.Contexts}, {"requires-skill", p.manifest.Skills}}
		for _, family := range reqs {
			for slot, name := range contextpkg.SortedNames(family.values) {
				dest := b.layers[name]
				if dest == nil {
					return refusal("graph_invalid_input")
				}
				r := requirement(family.values[name])
				kind := family.kind
				if kind == "requires-context" && dest.Kind == "profile-package" {
					kind = "requires-profile"
				}
				b.edge(layer.ID, dest.ID, kind, slot, &r)
			}
		}
	}
	return b.err
}
func (b *build) precedence() contextmaterialize.Precedence {
	p := b.req.Projection.Precedence
	if b.lock.Precedence.Winner != "" {
		p.Winner = b.lock.Precedence.Winner
		p.Placement = b.lock.Precedence.Placement
	}
	if p.Winner == "" {
		p = contextmaterialize.DefaultPrecedence
	}
	return p
}

var referenceLink = regexp.MustCompile(`\]\(([^\s)]+)(?:\s+[^)]*)?\)`)

func (b *build) skill(p packageInput, n *Node) error {
	raw, err := b.packageRead(p, "SKILL.md")
	if err != nil {
		return err
	}
	name, description, bodyStart, err := manifest.ParseSkillListing(raw)
	if err != nil {
		return refusal("graph_invalid_input")
	}
	rel := "SKILL.md"
	l := b.add("skill-listing", "skill listing", b.provenance(p, &rel), n.Layer, n.ID, "skill-listing", rel)
	l.Content = measured([]byte(name + "\n" + description + "\n"))
	l.Skill = &Skill{"lock", "listed-only"}
	b.edge(n.ID, l.ID, "lists", 0, nil)
	body := b.add("skill-body", "skill body", b.provenance(p, &rel), n.Layer, n.ID, "skill-body", rel)
	body.Content = measured(raw[bodyStart:])
	body.SourceRanges = []Range{{bodyStart, len(raw)}}
	body.Skill = &Skill{"lock", "UNKNOWN"}
	b.edge(n.ID, body.ID, "body", 0, nil)
	for _, part := range []*Node{l, body} {
		for i := range part.Loads {
			part.Loads[i].State = "available"
		}
	}
	seen := map[string]bool{}
	for remaining := raw[bodyStart:]; len(remaining) > 0; {
		match := referenceLink.FindSubmatchIndex(remaining)
		if match == nil {
			break
		}
		rel := string(remaining[match[2]:match[3]])
		remaining = remaining[match[1]:]
		if len(seen) >= b.r.Limits.Nodes {
			return refusal("graph_limit_exceeded")
		}
		if strings.Contains(rel, ":") || strings.HasPrefix(rel, "#") {
			continue
		}
		rel = strings.Split(rel, "#")[0]
		if !cleanRelative(rel) || denied(rel) {
			return refusal("graph_path_denied")
		}
		if seen[rel] {
			continue
		}
		seen[rel] = true
		ref := b.add("skill-reference", "skill reference", b.provenance(p, &rel), n.Layer, n.ID, "skill-reference", rel)
		ref.Content = Content{Status: "unread"}
		ref.Skill = &Skill{"lock", "UNKNOWN"}
		for i := range ref.Loads {
			ref.Loads[i].State = "available"
		}
		b.edge(n.ID, ref.ID, "references", len(seen)-1, nil)
	}
	return b.err
}

func (b *build) finish() error {
	if b.err != nil {
		return b.err
	}
	for _, n := range b.g.Nodes {
		if b.nodes[n.ID] != n {
			return refusal("graph_invalid_input")
		}
		if n.Layer != nil {
			layer := b.nodes[*n.Layer]
			if layer == nil || layer.Kind != "overlay" && !strings.HasSuffix(layer.Kind, "package") {
				return refusal("graph_invalid_input")
			}
		}
	}
	for _, e := range b.g.Edges {
		if b.nodes[e.From] == nil || b.nodes[e.To] == nil {
			return refusal("graph_invalid_input")
		}
	}
	for _, p := range []struct {
		name  string
		value any
	}{{"composition.json", b.req.Projection}, {"environments.json", b.g.Scope.Environments}} {
		raw, err := json.Marshal(p.value)
		if err != nil {
			return err
		}
		var v any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&v); err != nil {
			return err
		}
		raw, err = protocoljson.MarshalCanonical(v)
		if err != nil {
			return err
		}
		if len(raw) > b.r.Limits.File || len(raw) > b.r.Limits.Input-b.r.total {
			return refusal("graph_limit_exceeded")
		}
		b.r.total += len(raw)
		b.r.inputs["projection\x00"+p.name] = Input{"projection", p.name, hash(raw), len(raw)}
	}
	for _, input := range b.r.inputs {
		b.g.Inputs = append(b.g.Inputs, input)
	}
	sort.Slice(b.g.Inputs, func(i, j int) bool {
		a, c := b.g.Inputs[i], b.g.Inputs[j]
		if a.Root != c.Root {
			return a.Root < c.Root
		}
		return a.Path < c.Path
	})
	var transcript strings.Builder
	transcript.WriteString(SchemaVersion + "\n" + b.g.Scope.Mode + "\n" + b.req.Profile + "\n")
	for _, env := range b.g.Scope.Environments {
		transcript.WriteString(env + "\n")
	}
	transcript.WriteByte('\n')
	for _, in := range b.g.Inputs {
		if strings.ContainsAny(in.Root+in.Path, "\x00\r\n") {
			return refusal("graph_invalid_input")
		}
		transcript.WriteString(in.Root + "\x00" + in.Path + "\x00" + in.SHA256 + "\n")
	}
	b.g.InputsDigest = hash([]byte(transcript.String()))
	sort.Slice(b.g.Nodes, func(i, j int) bool { return b.g.Nodes[i].ID < b.g.Nodes[j].ID })
	sort.Slice(b.g.Edges, func(i, j int) bool {
		a, c := b.g.Edges[i], b.g.Edges[j]
		if a.From != c.From {
			return a.From < c.From
		}
		if a.Kind != c.Kind {
			return a.Kind < c.Kind
		}
		if a.Order != c.Order {
			return a.Order < c.Order
		}
		return a.To < c.To
	})
	sort.Slice(b.g.Documents, func(i, j int) bool {
		a, c := b.g.Documents[i], b.g.Documents[j]
		if a.Environment != c.Environment {
			return a.Environment < c.Environment
		}
		return a.Path < c.Path
	})
	b.g.Diagnostics = append(b.g.Diagnostics, Diagnostic{Class: "graph_load_unknown", Message: "Curator rendering is not evidence of runtime injection."})
	return validateWireStrings(&b.g)
}
