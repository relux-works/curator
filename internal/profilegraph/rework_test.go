package profilegraph_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/profilegraph"
	"github.com/relux-works/curator/internal/profilegraph/graphtest"
)

// Each family refuses a declared payload before open, and skips an undeclared
// sibling before open/hash/ledger insertion, in source and installed state trees.
// A nearby ordinary module must still render; broad rejection cannot pass.
func TestProducerDenyFamilies(t *testing.T) {
	names := []string{"token.json", "oauth.json", ".env.local", "secrets.json", "credentials.json", ".env", ".EnV.LoCaL", "access-ToKeN.json", "my-SeCrEt.json", "my-CrEdEnTiAl.json", "my-OAuTh.json", "AuTh.json", "cert.PEM", "private.KEY", "ID_ed25519", "cache.SQLITE-wal", "cache.DB", "HiStOrY.backup", "chat-SeSsIoN.json", "trace.LOG", "login.KEYCHAIN-db", "login.KEYRING", "token-budget.md", "bash_history", "cache-wal", "cache-shm", "auth/cache.json", "logs/trace.txt", "log/trace.txt", "shell-snapshots/capture.txt"}
	for i, name := range names {
		for _, installed := range []bool{false, true} {
			for _, declared := range []bool{false, true} {
				t.Run(fmt.Sprintf("%02d/installed=%t/declared=%t", i, installed, declared), func(t *testing.T) {
					home, source := simpleFixture(t)
					req := buildRequest(home)
					if !installed {
						req.Source = source
					}
					rel := "context/" + name
					if declared {
						manifest := fmt.Sprintf(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":%q},{"path":"base.md"}]}}`, name)
						writeFixture(t, source, "agent-context.json", []byte(manifest))
					}
					writeFixture(t, source, rel, []byte("synthetic denied payload\n"))
					req.Reader = profilegraph.NewReader(profilegraph.V1Limits)
					opened := false
					ordinaryOpened := false
					req.Reader.OpenHook = func(_, p string) {
						if p == rel {
							opened = true
						}
						if p == "context/base.md" {
							ordinaryOpened = true
						}
					}
					before := snapshot(t, home)
					g, err := profilegraph.Build(req)
					if opened {
						t.Fatal("denied payload opened")
					}
					if declared {
						if g != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_path_denied:") {
							t.Fatalf("declared: graph=%v err=%v", g, err)
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						if !ordinaryOpened || len(g.Documents) != 1 || g.Documents[0].Bytes == 0 {
							t.Fatal("ordinary module did not render")
						}
						for _, input := range g.Inputs {
							if input.Path == rel {
								t.Fatal("denied payload hashed/listed")
							}
						}
					}
					if snapshot(t, home) != before {
						t.Fatal("graph changed files")
					}
				})
			}
		}
	}
}

// The fixture builds real resolver state seeds and lock members, then installs
// their exact bytes without provisioning an app or writing to a real home.
func localOverlayFixture(t *testing.T, weights ...int64) (profilegraph.Request, string, []byte) {
	t.Helper()
	home, source := t.TempDir(), t.TempDir()
	packages := map[string]map[string][]byte{}
	makePackage := func(name, root string) *contextresolve.StatePackage {
		files := map[string][]byte{
			"agent-context.json": []byte(fmt.Sprintf(`{"schema_version":2,"name":%q,"version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`, name)),
			"context/base.md":    []byte("# " + name + "\n\nLocal rule.\n"),
		}
		for p, raw := range files {
			writeFixture(t, root, p, raw)
		}
		packages[name] = files
		h, err := hashing.ContentSHA256Files(files, hashing.WriteVersion())
		if err != nil {
			t.Fatal(err)
		}
		m, err := contextpkg.ParseManifest(files["agent-context.json"])
		if err != nil {
			t.Fatal(err)
		}
		return &contextresolve.StatePackage{StateHash: strings.TrimPrefix(h, "sha256:"), Manifest: envprofile.GraphPackage(m)}
	}
	input := contextresolve.Input{Root: contextresolve.Requirement{Kind: contextlock.KindContext, Name: "sample"}, RootState: makePackage("sample", source)}
	projection := profilegraph.Projection{}
	for i, weight := range weights {
		root := filepath.Join(t.TempDir(), "package")
		name := fmt.Sprintf("extra%d", i)
		state := makePackage(name, root)
		input.Overlays = append(input.Overlays, contextresolve.Overlay{Name: name, Source: root, State: state, Weight: &weight})
		projection.Overlays = append(projection.Overlays, envprofile.OverlaySpec{Source: root, Weight: &weight})
	}
	resolved, err := contextresolve.Resolve(nil, input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := resolved.Lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, home, "profiles/sample/lock.json", raw)
	writeFixture(t, home, "profiles/sample/source.json", []byte(`{"kind":"path","path":"source","requirement":{}}`))
	for _, member := range resolved.Lock.Members {
		for p, raw := range packages[member.Name] {
			writeFixture(t, envprofile.GraphMemberEntry(home, member), p, raw)
		}
	}
	req := buildRequest(home)
	req.Projection = projection
	material := map[string]contextmaterialize.Package{}
	for _, member := range resolved.Lock.Members {
		files := packages[member.Name]
		manifest, err := contextpkg.ParseManifest(files["agent-context.json"])
		if err != nil {
			t.Fatal(err)
		}
		material[member.Name] = contextmaterialize.Package{HasContext: manifest.HasContext, Modules: []contextmaterialize.Module{{Module: manifest.Modules[0], Bytes: files["context/base.md"]}}}
	}
	emitted, _, err := contextmaterialize.Monolithic(resolved.Lock, resolved.LockHash, contextmaterialize.DefaultPrecedence, "claude_code", material)
	if err != nil {
		t.Fatal(err)
	}
	return req, source, emitted
}

// Installed and source graphs preserve path-overlay slots, state provenance,
// weights and actual rendered bytes. Changing only weight changes occurrence id.
func TestProducerLocalOverlays(t *testing.T) {
	for _, preview := range []bool{false, true} {
		t.Run(fmt.Sprintf("source=%t", preview), func(t *testing.T) {
			ids := map[int64]string{}
			for _, weight := range []int64{5, 9} {
				t.Run(fmt.Sprintf("weight=%d", weight), func(t *testing.T) {
					req, source, emitted := localOverlayFixture(t, weight, 17)
					if preview {
						req.Source = source
					}
					before := snapshot(t, req.Home)
					g, err := profilegraph.Build(req)
					if err != nil {
						t.Fatal(err)
					}
					count := 0
					for _, n := range g.Nodes {
						if n.Kind != "overlay" {
							continue
						}
						count++
						i := 0
						switch n.Label {
						case "extra0":
						case "extra1":
							i = 1
						default:
							t.Fatal("unexpected overlay")
						}
						wantWeight := weight
						if i == 1 {
							wantWeight = 17
						}
						if n.Composition.Declaration != fmt.Sprintf("overlays/%d", i) || n.Composition.Weight != wantWeight || n.Composition.Requirement.Form != "path" || n.Source.Revision.Kind != "state-sha256" || len(n.Source.Revision.Value) != 64 {
							t.Fatalf("bad overlay: %+v", n)
						}
						if i == 0 {
							ids[weight] = n.ID
						}
					}
					if count != 2 || len(g.Documents) != 1 {
						t.Fatal("missing overlays/document")
					}
					if g.Documents[0].Bytes != len(emitted) || g.Documents[0].SHA256 != hash(emitted) || g.Documents[0].Lines != lineCount(emitted) {
						t.Fatal("graph differs from the untraced renderer")
					}
					// All three modules must contribute actual output bytes.
					modules := 0
					for _, n := range g.Nodes {
						if n.Kind == "instruction-module" && len(n.Emissions) > 0 {
							modules++
						}
					}
					if modules != 3 {
						t.Fatalf("rendered modules=%d", modules)
					}
					if snapshot(t, req.Home) != before {
						t.Fatal("graph wrote installed state")
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
				})
			}
			if ids[5] == "" || ids[5] == ids[9] {
				t.Fatal("weight-only change collapsed overlay identity")
			}
		})
	}
}

// A valid installed overlay is the positive control above. Mismatching name,
// state bytes or weight must refuse with no partial graph or writes.
func TestProducerLocalOverlayMismatch(t *testing.T) {
	for _, change := range []string{"name", "state", "weight", "git-form"} {
		t.Run(change, func(t *testing.T) {
			req, _, _ := localOverlayFixture(t, 5)
			o := &req.Projection.Overlays[0]
			switch change {
			case "name":
				writeFixture(t, o.Source, "agent-context.json", []byte(`{"schema_version":2,"name":"other","version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`))
			case "state":
				writeFixture(t, o.Source, "context/base.md", []byte("# Changed\n"))
			case "weight":
				weight := int64(6)
				o.Weight = &weight
			case "git-form":
				o.Revision = strings.Repeat("a", 40)
			}
			before := snapshot(t, req.Home)
			g, err := profilegraph.Build(req)
			if g != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_invalid_input:") {
				t.Fatalf("graph=%v err=%v", g, err)
			}
			if snapshot(t, req.Home) != before {
				t.Fatal("refusal wrote state")
			}
		})
	}
}

// Declared Git-object paths use the same family guard before opening the blob;
// an ordinary neighbor proves the offline object route can resolve and render.
func TestProducerObjectFamilies(t *testing.T) {
	names := []string{"token.json", "oauth.json", ".env.local", "secrets.json", "credentials.json", ".env", ".EnV.LoCaL", "access-ToKeN.json", "my-SeCrEt.json", "my-CrEdEnTiAl.json", "my-OAuTh.json", "AuTh.json", "cert.PEM", "private.KEY", "ID_ed25519", "cache.SQLITE-wal", "cache.DB", "HiStOrY.backup", "chat-SeSsIoN.json", "trace.LOG", "login.KEYCHAIN-db", "login.KEYRING", "token-budget.md", "bash_history", "cache-wal", "cache-shm", "auth/cache.json", "logs/trace.txt", "log/trace.txt", "shell-snapshots/capture.txt"}
	names = append(names, "guide.md")
	for i, name := range names {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			home, source := t.TempDir(), t.TempDir()
			repo := "example.invalid/extra"
			files := map[string][]byte{
				"agent-context.json": []byte(fmt.Sprintf(`{"schema_version":2,"name":"extra","version":"1.0.0","context":{"modules":[{"path":%q}]}}`, name)),
				"context/" + name:    []byte("# Synthetic rule\n"),
			}
			commit, err := graphtest.Commit(home, repo, files)
			if err != nil {
				t.Fatal(err)
			}
			blob, err := graphtest.Object(filepath.Join(envprofile.GraphRepository(home, repo), ".git"), "blob", files["context/"+name])
			if err != nil {
				t.Fatal(err)
			}
			writeFixture(t, source, "agent-context.json", []byte(fmt.Sprintf(`{"schema_version":2,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]},"requires":{"contexts":{"extra":{"git":"https://example.invalid/extra","revision":%q}}}}`, commit)))
			writeFixture(t, source, "context/base.md", []byte("# Base\n"))
			req := buildRequest(home)
			req.Source = source
			req.Reader = profilegraph.NewReader(profilegraph.V1Limits)
			opened := false
			req.Reader.OpenHook = func(_, rel string) {
				if rel == "objects/"+blob[:2]+"/"+blob[2:] {
					opened = true
				}
			}
			g, err := profilegraph.Build(req)
			if name == "guide.md" {
				if err != nil || g == nil || !opened {
					t.Fatalf("ordinary object failed: %v opened=%t", err, opened)
				}
			} else if g != nil || err == nil || !strings.HasPrefix(err.Error(), "graph_path_denied:") || opened {
				t.Fatalf("denied object: graph=%v err=%v opened=%t", g, err, opened)
			}
		})
	}
}
