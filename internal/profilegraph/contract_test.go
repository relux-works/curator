// Package profilegraph_test checks the proposed wire contract only. There is
// no graph command, filesystem admission implementation or runtime probe here.
package profilegraph_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaID = "https://relux-works.github.io/curator/schemas/curator-profile-graph-v1.schema.json"

type offlineLoader struct{}

func (offlineLoader) Load(string) (any, error) {
	return nil, errors.New("unbundled graph schema")
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func contractSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(offlineLoader{})
	if err := compiler.AddResource(schemaID, readJSON(t, "schemas/curator-profile-graph-v1.schema.json")); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(schemaID)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func object(v any) map[string]any { return v.(map[string]any) }

func nodeOf(g map[string]any, kind string) map[string]any {
	for _, value := range g["nodes"].([]any) {
		n := object(value)
		if n["kind"] == kind {
			return n
		}
	}
	panic("fixture has no " + kind)
}

// The golden satisfies the closed local schema. Each planted shape violation
// must fail schema validation at its affected property; no command is exercised.
func TestGraphSchemaGoldenAndRejections(t *testing.T) {
	schema := contractSchema(t)
	if err := schema.Validate(readJSON(t, "testdata/small/graph.json")); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, property string
		change         func(map[string]any)
	}{
		{"version", "schema_version", func(g map[string]any) { g["schema_version"] = "curator-profile-graph/2" }},
		{"closed-object", "additional properties 'surprise'", func(g map[string]any) { g["surprise"] = true }},
		{"negative-size", "bytes", func(g map[string]any) { object(g["documents"].([]any)[0])["bytes"] = -1 }},
		{"absolute-path", "path", func(g map[string]any) { object(g["inputs"].([]any)[0])["path"] = "/outside/file" }},
		{"parent-path", "path", func(g map[string]any) { object(g["inputs"].([]any)[0])["path"] = "root/../file" }},
		{"windows-path", "path", func(g map[string]any) { object(g["inputs"].([]any)[0])["path"] = `C:\outside\file` }},
		{"backslash-path", "path", func(g map[string]any) { object(g["inputs"].([]any)[0])["path"] = `root\file` }},
		{"missing-overlay-coordinates", "composition", func(g map[string]any) { delete(nodeOf(g, "overlay"), "composition") }},
		{"missing-heading", "heading", func(g map[string]any) { delete(nodeOf(g, "section"), "heading") }},
		{"short-commit", "value", func(g map[string]any) { object(object(nodeOf(g, "overlay")["source"])["revision"])["value"] = "abc" }},
		{"load-claim", "class", func(g map[string]any) {
			object(nodeOf(g, "instruction-module")["loads"].([]any)[0])["class"] = "observed"
		}},
		{"missing-skill-state", "materialization", func(g map[string]any) { delete(object(nodeOf(g, "skill")["skill"]), "materialization") }},
		{"binary-lines", "lines", func(g map[string]any) { object(nodeOf(g, "skill-reference")["content"])["status"] = "binary" }},
		{"unread-hash", "sha256", func(g map[string]any) { object(nodeOf(g, "skill-reference")["content"])["status"] = "unread" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := readJSON(t, "testdata/small/graph.json")
			tc.change(g)
			err := schema.Validate(g)
			var validation *jsonschema.ValidationError
			if !errors.As(err, &validation) || !strings.Contains(err.Error(), tc.property) {
				t.Fatalf("wanted schema rejection of %s, got %v", tc.property, err)
			}
		})
	}
}

type span struct{ Start, End int }
type emission struct {
	Document string
	Ranges   []span
	Bytes    int
	Lines    int
	SHA256   string
}
type node struct {
	ID, Kind, Layer string
	Identity        []string
	Emissions       []emission
	Heading         struct{ Text string }
}
type document struct {
	Node, Environment, Root, Path, SHA256 string
	Bytes, Lines                          int
	Coverage                              []struct {
		Node  string
		Range span
	}
}
type graph struct {
	InputsDigest string `json:"inputs_digest"`
	Scope        struct {
		Profile, Mode string
		Environments  []string
	}
	Inputs []struct {
		Root, Path, SHA256 string
		Bytes              int
	}
	Nodes []node
	Edges []struct {
		From, To, Kind string
		Order          int
	}
	Documents []document
}

func fixtureGraph(t *testing.T) graph {
	t.Helper()
	raw, err := os.ReadFile("testdata/small/graph.json")
	if err != nil {
		t.Fatal(err)
	}
	var g graph
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func lineCount(raw []byte) int {
	count := bytes.Count(raw, []byte{'\n'})
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		count++
	}
	return count
}

func tupleID(parts []string) string {
	var raw bytes.Buffer
	for _, part := range parts {
		raw.WriteString(strconv.Itoa(len([]byte(part))))
		raw.WriteByte(':')
		raw.WriteString(part)
	}
	return "n:" + hash(raw.Bytes())
}

// This fixture checker supplements JSON Schema with cross-record invariants.
// It is deliberately test-only and is not the future command's safety gate.
func verifyFixture(g graph, files fs.FS) error {
	nodes := map[string]node{}
	for _, n := range g.Nodes {
		if _, exists := nodes[n.ID]; exists || n.ID != tupleID(n.Identity) {
			return errors.New("identity: duplicate or inconsistent tuple")
		}
		nodes[n.ID] = n
	}
	for _, n := range g.Nodes {
		if n.Layer != "" {
			kind := nodes[n.Layer].Kind
			if kind != "root-package" && kind != "profile-package" && kind != "context-package" && kind != "overlay" {
				return errors.New("layer: missing package or overlay")
			}
		}
	}
	for _, e := range g.Edges {
		if _, ok := nodes[e.From]; !ok {
			return errors.New("edge: missing from")
		}
		if _, ok := nodes[e.To]; !ok {
			return errors.New("edge: missing to")
		}
	}
	var transcript strings.Builder
	transcript.WriteString("curator-profile-graph/1\n" + g.Scope.Mode + "\n" + g.Scope.Profile + "\n")
	for _, env := range g.Scope.Environments {
		transcript.WriteString(env + "\n")
	}
	transcript.WriteByte('\n')
	previous := ""
	for _, in := range g.Inputs {
		key := in.Root + "\x00" + in.Path
		if key <= previous {
			return errors.New("inputs: duplicate or unsorted key")
		}
		previous = key
		raw, err := fs.ReadFile(files, in.Path)
		if err != nil {
			return err
		}
		if in.Root != "fixture" || hash(raw) != in.SHA256 || len(raw) != in.Bytes {
			return errors.New("inputs: hash or size mismatch")
		}
		transcript.WriteString(key + "\x00" + in.SHA256 + "\n")
	}
	if hash([]byte(transcript.String())) != g.InputsDigest {
		return errors.New("inputs: digest mismatch")
	}
	docs := map[string][]byte{}
	for _, d := range g.Documents {
		raw, err := fs.ReadFile(files, d.Path)
		if err != nil {
			return err
		}
		if nodes[d.Node].Kind != "emitted-document" || len(raw) != d.Bytes || lineCount(raw) != d.Lines || hash(raw) != d.SHA256 {
			return errors.New("document: node, size, lines or hash mismatch")
		}
		docs[d.Node] = raw
		cursor := 0
		for _, c := range d.Coverage {
			n := nodes[c.Node]
			if c.Range.Start != cursor || c.Range.End <= cursor || c.Range.End > len(raw) {
				return errors.New("coverage: gap, overlap or invalid range")
			}
			if n.Kind != "generated" && n.Kind != "instruction-module" {
				return errors.New("coverage: contributor is not a leaf")
			}
			found := false
			for _, emission := range n.Emissions {
				if emission.Document == d.Node {
					for _, r := range emission.Ranges {
						found = found || (r.Start <= c.Range.Start && r.End >= c.Range.End)
					}
				}
			}
			if !found {
				return errors.New("coverage: contributor lacks projection")
			}
			cursor = c.Range.End
		}
		if cursor != len(raw) {
			return errors.New("coverage: incomplete total")
		}
	}
	for _, n := range g.Nodes {
		for _, e := range n.Emissions {
			raw, ok := docs[e.Document]
			if !ok {
				return errors.New("projection: missing document")
			}
			var union []byte
			previousEnd := -1
			for _, r := range e.Ranges {
				if r.Start <= previousEnd || r.Start < 0 || r.End <= r.Start || r.End > len(raw) {
					return errors.New("projection: overlapping or noncanonical ranges")
				}
				union = append(union, raw[r.Start:r.End]...)
				previousEnd = r.End
			}
			if len(union) != e.Bytes || lineCount(union) != e.Lines || hash(union) != e.SHA256 {
				return errors.New("projection: size, lines or hash mismatch")
			}
		}
	}
	return nil
}

// Both document ledgers exactly cover checked-in bytes. Deliberate corruptions
// remain schema-valid but must fail the named fixture invariant.
func TestGraphGoldenAccountingAndControls(t *testing.T) {
	files := os.DirFS("testdata/small")
	if err := verifyFixture(fixtureGraph(t), files); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, want string
		change     func(*graph)
	}{
		{"gap", "coverage:", func(g *graph) { g.Documents[0].Coverage[1].Range.Start++ }},
		{"overlap", "coverage:", func(g *graph) { g.Documents[0].Coverage[1].Range.Start-- }},
		{"missing-tail", "coverage: incomplete", func(g *graph) { g.Documents[0].Coverage = g.Documents[0].Coverage[:len(g.Documents[0].Coverage)-1] }},
		{"wrong-document-total", "document:", func(g *graph) { g.Documents[0].Bytes++ }},
		{"wrong-projection-hash", "projection:", func(g *graph) {
			for i := range g.Nodes {
				if len(g.Nodes[i].Emissions) > 0 {
					g.Nodes[i].Emissions[0].SHA256 = strings.Repeat("0", 64)
					return
				}
			}
		}},
		{"duplicate-id", "identity:", func(g *graph) { g.Nodes = append(g.Nodes, g.Nodes[0]) }},
		{"missing-edge-target", "edge:", func(g *graph) { g.Edges[0].To = "n:" + strings.Repeat("0", 64) }},
		{"wrong-input-digest", "inputs:", func(g *graph) { g.InputsDigest = strings.Repeat("0", 64) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := fixtureGraph(t)
			tc.change(&g)
			if err := verifyFixture(g, files); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
				t.Fatalf("wanted %s, got %v", tc.want, err)
			}
		})
	}
}

// The synthetic lock and source modules reproduce both golden outputs through
// the existing materializer; the environment-excluded module contributes zero
// bytes to Codex, and the fake heading inside a fence is never a section node.
func TestGraphGoldenMatchesMaterializer(t *testing.T) {
	base := os.DirFS("testdata/small")
	raw, err := fs.ReadFile(base, "context-lock.json")
	if err != nil {
		t.Fatal(err)
	}
	lock, err := contextlock.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	lockHash, err := lock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	packages := map[string]contextmaterialize.Package{}
	for _, directory := range []string{"root", "nested", "overlay"} {
		raw, err := fs.ReadFile(base, directory+"/agent-context.json")
		if err != nil {
			t.Fatal(err)
		}
		manifest, err := contextpkg.ParseManifest(raw)
		if err != nil {
			t.Fatal(err)
		}
		pkg := contextmaterialize.Package{HasContext: manifest.HasContext}
		for _, module := range manifest.Modules {
			raw, err := fs.ReadFile(base, directory+"/context/"+module.Path)
			if err != nil {
				t.Fatal(err)
			}
			pkg.Modules = append(pkg.Modules, contextmaterialize.Module{Module: module, Bytes: raw})
		}
		packages[manifest.Name] = pkg
	}
	g := fixtureGraph(t)
	for _, d := range g.Documents {
		actual, written, err := contextmaterialize.Monolithic(lock, lockHash, contextmaterialize.DefaultPrecedence, d.Environment, packages)
		if err != nil || !written {
			t.Fatalf("%s: written=%v err=%v", d.Environment, written, err)
		}
		expected, err := fs.ReadFile(base, d.Path)
		if err != nil || !bytes.Equal(actual, expected) {
			t.Fatalf("%s: materializer differs from golden: %v", d.Environment, err)
		}
		if d.Environment == "codex_cli" && bytes.Contains(actual, []byte("# Guide")) {
			t.Fatal("excluded module emitted")
		}
	}
	counts := map[string]int{}
	for _, n := range g.Nodes {
		counts[n.Kind]++
		if n.Kind == "section" && n.Heading.Text != "Baseline" && n.Heading.Text != "Guide" {
			t.Fatal("fenced text became a section")
		}
	}
	for kind, want := range map[string]int{"root-package": 1, "overlay": 1, "profile-package": 1, "instruction-module": 2, "section": 2, "skill": 1, "quote-block": 1} {
		if counts[kind] != want {
			t.Errorf("%s: got %d want %d", kind, counts[kind], want)
		}
	}
}

// Physical line metrics handle UTF-8, empty and unterminated slices. Hash
// tuple framing distinguishes ambiguous concatenations without ASCII offsets.
func TestGraphMetricConventions(t *testing.T) {
	for _, tc := range []struct {
		text         string
		bytes, lines int
	}{
		{"", 0, 0}, {"a", 1, 1}, {"a\n", 2, 1}, {"a\n\n", 3, 2},
		{"a\r\n", 3, 1}, {"\u03bb\n", 3, 1},
	} {
		if len([]byte(tc.text)) != tc.bytes || lineCount([]byte(tc.text)) != tc.lines {
			t.Errorf("metrics for %q", tc.text)
		}
	}
	if tupleID([]string{"ab", "c"}) == tupleID([]string{"a", "bc"}) {
		t.Fatal("tuple framing collision")
	}
	if tupleID([]string{"\u03bb"}) != "n:"+hash([]byte("2:\u03bb")) {
		t.Fatal("tuple length must count bytes")
	}
}

// The golden stores nodes and edges in contract order. The offline loader
// refuses an unbundled schema reference rather than consulting the network.
func TestGraphDeterminismAndOfflineSchema(t *testing.T) {
	g := fixtureGraph(t)
	if !sort.SliceIsSorted(g.Nodes, func(i, j int) bool { return g.Nodes[i].ID < g.Nodes[j].ID }) {
		t.Fatal("nodes not sorted")
	}
	if !sort.SliceIsSorted(g.Edges, func(i, j int) bool {
		a, b := g.Edges[i], g.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.To < b.To
	}) {
		t.Fatal("edges not sorted")
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(offlineLoader{})
	if _, err := compiler.Compile("https://example.invalid/missing-schema"); err == nil || !strings.Contains(err.Error(), "unbundled graph schema") {
		t.Fatalf("expected offline reference refusal, got %v", err)
	}
	if len(g.Nodes) > 20000 || len(g.Edges) > 50000 {
		t.Fatal("fixture exceeds graph budget")
	}
}
