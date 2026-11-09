package profilegraph_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/profilegraph"
	"github.com/relux-works/curator/internal/profilegraph/graphtest"
)

func writeFixture(t *testing.T, root, rel string, raw []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

// A heading at the wire bound is valid; a neighboring oversized heading must
// refuse atomically rather than emit a graph outside the published schema.
func TestProducerHeadingWireBound(t *testing.T) {
	for _, length := range []int{4096, 4097} {
		source := t.TempDir()
		writeFixture(t, source, "agent-context.json", []byte(`{"schema_version":2,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`))
		writeFixture(t, source, "context/base.md", []byte("# "+strings.Repeat("x", length)+"\n"))
		before := snapshot(t, source)
		g, err := profilegraph.Build(profilegraph.Request{Source: source, Profile: "sample", Environment: "codex", Version: "dev"})
		if length == 4096 {
			if err != nil || g == nil {
				t.Fatalf("valid bound: %v", err)
			}
		} else if g != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_invalid_input:") {
			t.Fatalf("oversized heading: %v", err)
		}
		if snapshot(t, source) != before {
			t.Fatal("wire validation wrote files")
		}
	}
}

// Preview runs install's resolver against only local Git objects: snapshots of
// the dependencies are deliberately absent. Shared DAG reuse is still legal,
// and both overlay declarations retain their exact directory and weight.
func TestSourcePreviewLocalObjectsDAG(t *testing.T) {
	home := t.TempDir()
	lock, err := graphtest.DAG(home)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := lock.RootMember()
	source := envprofile.GraphMemberEntry(home, root)
	for _, m := range lock.Members {
		if m.Name != lock.Root {
			if err := os.RemoveAll(envprofile.GraphMemberEntry(home, m)); err != nil {
				t.Fatal(err)
			}
		}
	}
	r := profilegraph.NewReader(profilegraph.V1Limits)
	projection, err := profilegraph.ReadProjection(r, filepath.Join(home, "config.json"), "sample")
	if err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, home)
	g, err := profilegraph.Build(profilegraph.Request{Profile: "sample", Source: source, Home: home, Environment: "codex", Reader: r, Projection: projection, Version: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot(t, home) != before {
		t.Fatal("preview wrote store, objects, refs or configuration")
	}
	leaf, overlays := "", 0
	for _, n := range g.Nodes {
		if n.Kind == "overlay" {
			overlays++
		}
		if n.Label == "leaf" {
			if leaf != "" {
				t.Fatal("duplicate shared dependency")
			}
			leaf = n.ID
		}
	}
	incoming := 0
	for _, e := range g.Edges {
		if e.To == leaf && e.Kind == "requires-context" {
			incoming++
		}
	}
	if overlays != 2 || leaf == "" || incoming != 2 {
		t.Fatalf("overlays=%d leaf=%q incoming=%d", overlays, leaf, incoming)
	}
}

// Output budgets include the LF, refuse atomically in both formats, and leave
// exact-boundary output byte-identical to encoding/json's independent encoder.
func TestProducerOutputBudget(t *testing.T) {
	home, _ := simpleFixture(t)
	g, err := profilegraph.Build(buildRequest(home))
	if err != nil {
		t.Fatal(err)
	}
	g.Generator.Version = "<unicode: \u2028>\n\""
	var expected bytes.Buffer
	enc := json.NewEncoder(&expected)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(g); err != nil {
		t.Fatal(err)
	}
	for _, limit := range []int{1, expected.Len() - 1} {
		raw, err := profilegraph.JSON(g, limit)
		if raw != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_limit_exceeded:") {
			t.Fatalf("limit=%d err=%v", limit, err)
		}
	}
	raw, err := profilegraph.JSON(g, expected.Len())
	if err != nil || !bytes.Equal(raw, expected.Bytes()) {
		t.Fatalf("exact-boundary JSON: %v", err)
	}
	text, err := profilegraph.Text(g, profilegraph.V1Limits.Output)
	if err != nil {
		t.Fatal(err)
	}
	if raw, err := profilegraph.Text(g, len(text)-1); raw != nil || err == nil {
		t.Fatal("partial/oversize text")
	}
}

// Both command captures satisfy the unchanged closed schema. The independent
// tuple, edge, load and ledger checks also inspect the generated records.
func TestProducerGoldenSchemaAndIdentity(t *testing.T) {
	for _, fixture := range []string{"small", "dag-overlays"} {
		value := readJSON(t, "testdata/"+fixture+"/graph.json")
		if err := contractSchema(t).Validate(value); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(value)
		var g profilegraph.Graph
		if err := json.Unmarshal(raw, &g); err != nil {
			t.Fatal(err)
		}
		nodes := map[string]*profilegraph.Node{}
		for _, n := range g.Nodes {
			if n.ID != tupleID(n.Identity) || nodes[n.ID] != nil {
				t.Fatal("invalid/duplicate identity")
			}
			nodes[n.ID] = n
		}
		for _, e := range g.Edges {
			if nodes[e.From] == nil || nodes[e.To] == nil {
				t.Fatal("dangling edge")
			}
		}
		for _, n := range g.Nodes {
			if n.Layer != nil {
				layer := nodes[*n.Layer]
				if layer == nil || layer.Kind != "overlay" && !strings.HasSuffix(layer.Kind, "package") {
					t.Fatal("invalid layer")
				}
			}
			for _, l := range n.Loads {
				if l.Class != "UNKNOWN" && l.Basis != "rendering-rule" {
					t.Fatal("unsupported load claim")
				}
			}
		}
		for _, d := range g.Documents {
			cursor := 0
			for _, c := range d.Coverage {
				if c.Range.Start != cursor || c.Range.End <= cursor || nodes[c.Node] == nil {
					t.Fatal("invalid ledger")
				}
				cursor = c.Range.End
			}
			if cursor != d.Bytes {
				t.Fatal("incomplete ledger")
			}
		}
	}
}

// An intermediate cache-directory link cannot redefine the admitted root.
// An internal link to a denied basename is refused before payload open too.
func TestProducerParentLinksAndDeniedTargets(t *testing.T) {
	for _, deniedTarget := range []bool{false, true} {
		home, entry := simpleFixture(t)
		if deniedTarget {
			writeFixture(t, entry, "context/auth.json", []byte("denied payload\n"))
			if err := os.Remove(filepath.Join(entry, "context/base.md")); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("auth.json", filepath.Join(entry, "context/base.md")); err != nil {
				t.Fatal(err)
			}
		} else {
			outside := filepath.Join(t.TempDir(), "snapshot")
			if err := os.Rename(entry, outside); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, entry); err != nil {
				t.Fatal(err)
			}
		}
		req := buildRequest(home)
		req.Reader = profilegraph.NewReader(profilegraph.V1Limits)
		opened := false
		req.Reader.OpenHook = func(_, rel string) {
			if strings.Contains(rel, "auth.json") {
				opened = true
			}
		}
		before := snapshot(t, home)
		g, err := profilegraph.Build(req)
		want := "graph_symlink_escape"
		if deniedTarget {
			want = "graph_path_denied"
		}
		if g != nil || err == nil || !strings.HasPrefix(err.Error(), want+":") || opened {
			t.Fatalf("wanted %s, got %v opened=%v", want, err, opened)
		}
		if snapshot(t, home) != before {
			t.Fatal("link refusal wrote files")
		}
	}
}

// Selected skill command admission matches install: a real declared command
// passes; a neighboring undeclared command refuses, without running scripts.
func TestSourcePreviewScopedSkillCommands(t *testing.T) {
	for _, command := range []string{"check", "absent"} {
		home, source := t.TempDir(), t.TempDir()
		files := map[string][]byte{
			"skills/tool/SKILL.md":         []byte("---\r\nname: tool\r\ndescription: Synthetic tool\r\n---\r\nBody.\r\n"),
			"skills/tool/agent-skill.json": []byte(`{"schema_version":2,"runtime_roots":["scripts"],"commands":{"check":{"type":"script","unix_path":"scripts/check.sh"}}}`),
			"skills/tool/scripts/check.sh": []byte("exit 97\n"),
		}
		pin, err := graphtest.Commit(home, "example.invalid/tools", files)
		if err != nil {
			t.Fatal(err)
		}
		root := map[string]any{"schema_version": 2, "name": "sample", "version": "1.0.0", "context": map[string]any{"modules": []any{map[string]any{"path": "base.md"}}}, "requires": map[string]any{"skills": map[string]any{"tool": map[string]any{"git": "https://example.invalid/tools", "revision": pin, "directory": "skills/tool", "mode": "runtime", "commands": []string{command}}}}}
		raw, _ := json.Marshal(root)
		writeFixture(t, source, "agent-context.json", raw)
		writeFixture(t, source, "context/base.md", []byte("# Sample\n"))
		before := snapshot(t, home) + snapshot(t, source)
		g, err := profilegraph.Build(profilegraph.Request{Home: home, Source: source, Profile: "sample", Environment: "codex", Version: "dev"})
		if command == "check" {
			if err != nil || g == nil {
				t.Fatalf("valid scoped command: %v", err)
			}
			// CRLF is retained in body bytes; listing text uses decoded YAML.
			seenListing, seenBody := false, false
			for _, n := range g.Nodes {
				if n.Kind == "skill-listing" {
					seenListing = n.Content.SHA256 != nil && *n.Content.SHA256 == fmt.Sprintf("%x", sha256.Sum256([]byte("tool\nSynthetic tool\n")))
				}
				if n.Kind == "skill-body" {
					seenBody = n.Content.SHA256 != nil && *n.Content.SHA256 == fmt.Sprintf("%x", sha256.Sum256([]byte("Body.\r\n")))
				}
			}
			if !seenListing || !seenBody {
				t.Fatal("CRLF listing/body accounting differs")
			}
		} else if err == nil || g != nil || !strings.HasPrefix(err.Error(), "graph_invalid_input:") {
			t.Fatalf("undeclared command: %v", err)
		}
		if snapshot(t, home)+snapshot(t, source) != before {
			t.Fatal("skill preview wrote files")
		}
	}
}

// Independently render both supported forms through the pre-existing API and
// compare every reported hash, total and inclusive byte projection with those
// bytes. This guards attribution even when a graph remains schema-valid.
func TestProducerAccountingAgainstRenderer(t *testing.T) {
	for _, form := range []string{"monolithic", "referenced"} {
		home := t.TempDir()
		lock, err := graphtest.DAG(home)
		if err != nil {
			t.Fatal(err)
		}
		r := profilegraph.NewReader(profilegraph.V1Limits)
		projection, err := profilegraph.ReadProjection(r, filepath.Join(home, "config.json"), "sample")
		if err != nil {
			t.Fatal(err)
		}
		projection.Forms = map[string]string{"claude_code": form}
		g, err := profilegraph.Build(profilegraph.Request{Profile: "sample", Home: home, Environment: "claude", Version: "dev", Reader: r, Projection: projection})
		if err != nil {
			t.Fatal(err)
		}
		packages := map[string]contextmaterialize.Package{}
		for _, m := range lock.Members {
			root := filepath.Join(envprofile.GraphMemberEntry(home, m), filepath.FromSlash(m.Directory))
			raw, err := os.ReadFile(filepath.Join(root, "agent-context.json"))
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := contextpkg.ParseManifest(raw)
			if err != nil {
				t.Fatal(err)
			}
			p := contextmaterialize.Package{HasContext: manifest.HasContext}
			for _, mod := range manifest.Modules {
				raw, err := os.ReadFile(filepath.Join(root, "context", mod.Path))
				if err != nil {
					t.Fatal(err)
				}
				p.Modules = append(p.Modules, contextmaterialize.Module{Module: mod, Bytes: raw})
			}
			packages[m.Name] = p
		}
		lockHash, err := lock.Hash()
		if err != nil {
			t.Fatal(err)
		}
		files := map[string][]byte{}
		if form == "referenced" {
			files, _, err = contextmaterialize.Referenced(lock, lockHash, contextmaterialize.DefaultPrecedence, "claude_code", packages)
		} else {
			var raw []byte
			raw, _, err = contextmaterialize.Monolithic(lock, lockHash, contextmaterialize.DefaultPrecedence, "claude_code", packages)
			files["CLAUDE.md"] = raw
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Documents) != len(files) {
			t.Fatal("missing rendered document")
		}
		docs := map[string][]byte{}
		for _, d := range g.Documents {
			raw := files[d.Path]
			if len(raw) != d.Bytes || hash(raw) != d.SHA256 || lineCount(raw) != d.Lines {
				t.Fatal("wrong document measurement")
			}
			docs[d.Node] = raw
			cursor := 0
			for _, c := range d.Coverage {
				if c.Range.Start != cursor || c.Range.End <= cursor || c.Range.End > len(raw) {
					t.Fatal("bad coverage")
				}
				cursor = c.Range.End
			}
			if cursor != len(raw) {
				t.Fatal("coverage total")
			}
		}
		for _, n := range g.Nodes {
			for _, e := range n.Emissions {
				raw := docs[e.Document]
				var union []byte
				previous := -1
				for _, r := range e.Ranges {
					if r.Start <= previous || r.Start < 0 || r.End <= r.Start || r.End > len(raw) {
						t.Fatal("bad union")
					}
					union = append(union, raw[r.Start:r.End]...)
					previous = r.End
				}
				if len(union) != e.Bytes || lineCount(union) != e.Lines || hash(union) != e.SHA256 {
					t.Fatal("wrong inclusive projection")
				}
			}
		}
	}
}

// A render-time node-cap failure remains an atomic refusal, not a panic or a
// partial document. The same fixture with the production budget succeeds above.
func TestProducerRenderLimitIsAtomic(t *testing.T) {
	home, _ := simpleFixture(t)
	req := buildRequest(home)
	req.Reader = profilegraph.NewReader(profilegraph.V1Limits)
	req.Reader.Limits.Nodes = 5
	before := snapshot(t, home)
	g, err := profilegraph.Build(req)
	if g != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_limit_exceeded:") {
		t.Fatalf("render limit: %v", err)
	}
	if snapshot(t, home) != before {
		t.Fatal("render refusal wrote inputs")
	}
}
func simpleFixture(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	source := t.TempDir()
	manifest := []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`)
	writeFixture(t, source, "agent-context.json", manifest)
	writeFixture(t, source, "context/base.md", []byte("# Baseline\n\nKeep the rule.\n"))
	lock := &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: "sample", Members: []contextlock.Member{{Kind: "context", Name: "sample", StateHash: strings.Repeat("a", 64), Version: "1.0.0", RequiredBy: []string{}}}}
	raw, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, home, "profiles/sample/lock.json", raw)
	writeFixture(t, home, "profiles/sample/source.json", []byte(`{"kind":"path","path":"source","requirement":{}}`))
	entry := envprofile.GraphMemberEntry(home, lock.Members[0])
	writeFixture(t, entry, "agent-context.json", manifest)
	writeFixture(t, entry, "context/base.md", []byte("# Baseline\n\nKeep the rule.\n"))
	repinFixture(t, home, entry)
	raw, err = os.ReadFile(filepath.Join(home, "profiles/sample/lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err = contextlock.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return home, envprofile.GraphMemberEntry(home, lock.Members[0])
}
func snapshot(t *testing.T, root string) string {
	t.Helper()
	var out strings.Builder
	err := filepath.WalkDir(root, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		fmt.Fprintf(&out, "%s %s %d\n", rel, info.Mode(), info.Size())
		fmt.Fprintf(&out, "mtime=%d\n", info.ModTime().UnixNano())
		if info.Mode().IsRegular() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(&out, "%s\n", hash(raw))
		} else if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			out.WriteString(target)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.String()
}
func buildRequest(home string) profilegraph.Request {
	return profilegraph.Request{Home: home, Profile: "sample", Environment: "claude", Version: "dev"}
}

// Each fatal fixture proves its named refusal, a nil graph and unchanged inputs.
// The positive control exercises the same installed read/render path.
func TestProducerFatalClassesAreAtomic(t *testing.T) {
	cases := []struct {
		name  string
		plant func(*testing.T, string, string, *profilegraph.Request)
	}{
		{"graph_cycle", func(t *testing.T, _ string, entry string, _ *profilegraph.Request) {
			writeFixture(t, entry, "agent-context.json", []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]},"requires":{"contexts":{"sample":{"git":"https://example.invalid/sample","revision":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}}}`))
		}},
		{"graph_missing_input", func(t *testing.T, _ string, entry string, _ *profilegraph.Request) {
			if err := os.Remove(filepath.Join(entry, "context/base.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"graph_path_denied", func(t *testing.T, _ string, entry string, _ *profilegraph.Request) {
			writeFixture(t, entry, "agent-context.json", []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"auth.json"}]}}`))
			writeFixture(t, entry, "context/auth.json", []byte("forbidden sentinel\n"))
		}},
		{"graph_symlink_escape", func(t *testing.T, _ string, entry string, _ *profilegraph.Request) {
			full := filepath.Join(entry, "context/base.md")
			if err := os.Remove(full); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(t.TempDir(), "outside.md")
			writeFixture(t, filepath.Dir(target), "outside.md", []byte("outside\n"))
			if err := os.Symlink(target, full); err != nil {
				t.Fatal(err)
			}
		}},
		{"graph_limit_exceeded", func(_ *testing.T, _, _ string, req *profilegraph.Request) { req.Reader.Limits.Nodes = 1 }},
		{"graph_document_drift", func(t *testing.T, home, _ string, _ *profilegraph.Request) {
			writeFixture(t, envprofile.ManagedHomeDir(home, "sample", "claude_code"), "CLAUDE.md", []byte("drift\n"))
		}},
		{"graph_invalid_input", func(t *testing.T, _ string, entry string, _ *profilegraph.Request) {
			writeFixture(t, entry, "agent-context.json", []byte("{"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home, entry := simpleFixture(t)
			req := buildRequest(home)
			req.Reader = profilegraph.NewReader(profilegraph.V1Limits)
			var opens []string
			req.Reader.OpenHook = func(_, p string) { opens = append(opens, p) }
			tc.plant(t, home, entry, &req)
			if tc.name == "graph_cycle" {
				repinFixture(t, home, entry)
			}
			before := snapshot(t, home)
			g, err := profilegraph.Build(req)
			if g != nil || err == nil || !strings.HasPrefix(err.Error(), tc.name+":") {
				t.Fatalf("graph=%v err=%v", g, err)
			}
			if snapshot(t, home) != before {
				t.Fatal("refusal changed input files")
			}
			if tc.name == "graph_path_denied" {
				for _, p := range opens {
					if strings.Contains(p, "auth.json") {
						t.Fatal("denied payload opened")
					}
				}
			}
		})
	}
	t.Run("positive", func(t *testing.T) {
		home, _ := simpleFixture(t)
		before := snapshot(t, home)
		g, err := profilegraph.Build(buildRequest(home))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := profilegraph.JSON(g, profilegraph.V1Limits.Output)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if json.Unmarshal(raw, &value) != nil {
			t.Fatal("invalid JSON")
		}
		if err := contractSchema(t).Validate(value); err != nil {
			t.Fatal(err)
		}
		if len(g.Documents) != 1 || g.Documents[0].Bytes == 0 {
			t.Fatal("no real document")
		}
		if snapshot(t, home) != before {
			t.Fatal("success wrote inputs")
		}
	})
}

// Source preview produces real renderer bytes without an installed profile.
// Fenced pseudo-headings do not become sections; a quote remains a source span.
func TestSourcePreviewMarkdownAndNoBootstrap(t *testing.T) {
	source := t.TempDir()
	writeFixture(t, source, "agent-context.json", []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`))
	writeFixture(t, source, "context/base.md", []byte("preamble\n\nTitle\n=====\n\n## Child\n> quote\n~~~go\n# fake\n~~~\n"))
	before := snapshot(t, source)
	g, err := profilegraph.Build(profilegraph.Request{Profile: "sample", Source: source, Environment: "codex", Version: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	titles := map[string]bool{}
	for _, n := range g.Nodes {
		if n.Kind == "section" {
			titles[n.Heading.Text] = true
		}
	}
	if len(titles) != 2 || !titles["Title"] || !titles["Child"] {
		t.Fatalf("headings=%v", titles)
	}
	if snapshot(t, source) != before {
		t.Fatal("source changed")
	}
	raw, err := profilegraph.JSON(g, profilegraph.V1Limits.Output)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(source)) {
		t.Fatal("absolute source path leaked")
	}
}

func repinFixture(t *testing.T, home, entry string) {
	t.Helper()
	files := map[string][]byte{}
	if err := filepath.WalkDir(entry, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(entry, p)
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = raw
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	h, err := hashing.ContentSHA256Files(files, hashing.VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(home, "profiles/sample/lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := contextlock.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	lock.Members[0].StateHash = strings.TrimPrefix(h, "sha256:")
	newEntry := envprofile.GraphMemberEntry(home, lock.Members[0])
	if entry != newEntry {
		if err := os.Rename(entry, newEntry); err != nil {
			t.Fatal(err)
		}
	}
	raw, err = lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, home, "profiles/sample/lock.json", raw)
}
