package projectmcpstate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/sourcelock"
)

func testLock(t *testing.T) *Lock {
	t.Helper()
	hash := "sha256:" + strings.Repeat("1", 64)
	skills, err := sourcelock.New(hash, []sourcelock.Member{})
	if err != nil {
		t.Fatal(err)
	}
	decl, err := projectmcp.ParseDeclaration(decode(t, `{"package":{"git":"https://packages.example/tool","tag":"v1.0.0"},"install":{"scope":"user","sig":"opaque"}}`))
	if err != nil {
		t.Fatal(err)
	}
	server, err := projectmcp.NormalizeServer(decode(t, `{"transport":"stdio","command":"npx","args":["tool@1.0.0"],"environments":["pi","codex_cli"]}`))
	if err != nil {
		t.Fatal(err)
	}
	dh, _ := projectmcp.Digest(decl.Model())
	sh, _ := projectmcp.Digest(server)
	lock := &Lock{SchemaVersion: 1, Capability: projectmcp.Capability, ProjectID: "project", ManifestSHA256: hash, SkillsLock: skills.Object(), Entries: []Entry{{Name: "tool", Declaration: decl.Model(), DeclarationSHA256: dh, Resolved: Resolved{Source: "packages.example/tool", Commit: strings.Repeat("a", 40), Version: "1.0.0", ContentSHA256: hash, Server: server, ServerSHA256: sh}}}}
	if err := lock.Seal(); err != nil {
		t.Fatal(err)
	}
	return lock
}

// Canonical replay binds the independent declaration/server/source identities;
// recomputing the outer hash cannot authorize altered inner records.
func TestLockBinding(t *testing.T) {
	lock := testLock(t)
	payload, err := Canonical(lock)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := DecodeLock(payload)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := Canonical(replay)
	if !bytes.Equal(payload, again) {
		t.Fatal("round trip changed bytes")
	}
	for _, tc := range []struct {
		name   string
		change func(*Lock)
	}{
		{"declaration", func(l *Lock) { l.Entries[0].Declaration["install"].(map[string]any)["scope"] = "different" }},
		{"server", func(l *Lock) { l.Entries[0].Resolved.Server["command"] = "other" }},
		{"content", func(l *Lock) { l.Entries[0].Resolved.ContentSHA256 = "" }},
		{"source", func(l *Lock) { l.Entries[0].Resolved.Source = "packages.example/other" }},
		{"directory", func(l *Lock) { d := "nested"; l.Entries[0].Resolved.Directory = &d }},
		{"version", func(l *Lock) { l.Entries[0].Resolved.Version = "latest" }},
		{"enclosed manifest", func(l *Lock) { l.ManifestSHA256 = "sha256:" + strings.Repeat("2", 64) }},
		{"duplicate", func(l *Lock) { l.Entries = append(l.Entries, l.Entries[0]) }},
		{"unnormalized declaration", func(l *Lock) {
			delete(l.Entries[0].Declaration, "install")
			d, _ := projectmcp.ParseDeclaration(l.Entries[0].Declaration)
			l.Entries[0].DeclarationSHA256, _ = projectmcp.Digest(d.Model())
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := testLock(t)
			tc.change(l)
			if err := l.Seal(); err == nil {
				t.Fatal("altered binding admitted")
			}
		})
	}
	if err := lock.CheckManifest("other", lock.ManifestSHA256); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatal(err)
	}
	if _, err := DecodeLock(append(payload, '\n')); err == nil {
		t.Fatal("noncanonical lock admitted")
	}
}

// Plans enumerate explicit targets independently of skill agents, expose only
// ignored option names, and require a lock for successful package observations.
func TestPlanTotalPairs(t *testing.T) {
	l := testLock(t)
	d, _ := projectmcp.ParseDeclaration(l.Entries[0].Declaration)
	decls := map[string]projectmcp.Declaration{"tool": d}
	p, err := NewPlan("project", l.ManifestSHA256, decls, l.Entries, &l.LockSHA256, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Rows) != 2 || p.Rows[0].Environment != "codex_cli" || p.Rows[1].Environment != "pi" {
		t.Fatalf("rows: %+v", p.Rows)
	}
	for _, row := range p.Rows {
		if row.Path != nil || row.Code != nil || row.Outcome != "skipped: unsupported" {
			t.Fatalf("row: %+v", row)
		}
	}
	lines := strings.Join(p.Lines(), "\n")
	if !strings.Contains(lines, `install option "scope": ignored`) || strings.Contains(lines, "opaque") || strings.Contains(lines, "user") || !strings.HasPrefix(lines, "would: ") {
		t.Fatal(lines)
	}
	if _, err := NewPlan("project", l.ManifestSHA256, decls, l.Entries, nil, true, nil); err == nil {
		t.Fatal("nonempty unlocked plan admitted")
	}
	if _, err := NewPlan("project", l.ManifestSHA256, decls, nil, nil, true, nil); err == nil {
		t.Fatal("nonempty unresolved plan admitted without lock")
	}
	if _, err := NewPlan("project", l.ManifestSHA256, nil, nil, nil, true, nil); err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"mcp_inline_approval_required", "mcp_install_option_invalid", "invented"} {
		if _, err := NewPlan("project", l.ManifestSHA256, decls, l.Entries, nil, true, map[string]string{"tool": code}); err == nil {
			t.Fatalf("reserved code %s admitted", code)
		}
	}
	refused, err := NewPlan("project", l.ManifestSHA256, decls, l.Entries, nil, true, map[string]string{"tool": "source_lock_stale"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range refused.Rows {
		if row.Outcome != "refused" || row.Code == nil || *row.Code != "source_lock_stale" {
			t.Fatal(row)
		}
	}
}

// A recognized inline arm always reports unsupported, even when a caller or an
// ignored option claims another approval verdict. No successful row is emitted.
func TestPlanInlineCannotBeApproved(t *testing.T) {
	d, err := projectmcp.ParseDeclaration(decode(t, `{"server":{"transport":"http","url":"https://service.example/mcp","environments":["muse"]},"install":{"approval":{"approved":true}}}`))
	if err != nil {
		t.Fatal(err)
	}
	hash := "sha256:" + strings.Repeat("1", 64)
	p, err := NewPlan("project", hash, map[string]projectmcp.Declaration{"inline": d}, nil, nil, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Rows) != 1 || p.Rows[0].Outcome != "refused" || p.Rows[0].Code == nil || *p.Rows[0].Code != projectmcp.InlineUnsupported {
		t.Fatalf("%+v", p.Rows)
	}
}

func decode(t *testing.T, text string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// Rehashing an observation cannot authorize an unsupported path, an unknown
// pair code, duplicate pairs or an honoured option. Each case starts with a
// valid locked plan; hashing is independent of these semantic refusals.
func TestPlanCarrierSemanticGates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Plan)
	}{
		{"valid", nil},
		{"reserved code", func(p *Plan) {
			code := "mcp_inline_approval_required"
			p.Rows[0].Outcome = "refused"
			p.Rows[0].Code = &code
		}},
		{"duplicate pair", func(p *Plan) { p.Rows = append(p.Rows, p.Rows[0]) }},
		{"unsupported path", func(p *Plan) { path := ".mcp.json"; p.Rows[0].Path = &path }},
		{"successful missing path", func(p *Plan) { p.Rows[0].Outcome = "added" }},
		{"honoured option", func(p *Plan) { p.InstallOptions[0].Handling = "honoured" }},
		{"duplicate option", func(p *Plan) { p.InstallOptions = append(p.InstallOptions, p.InstallOptions[0]) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := testLock(t)
			d, err := projectmcp.ParseDeclaration(l.Entries[0].Declaration)
			if err != nil {
				t.Fatal(err)
			}
			p, err := NewPlan("project", l.ManifestSHA256, map[string]projectmcp.Declaration{"tool": d}, l.Entries, &l.LockSHA256, false, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.change != nil {
				tc.change(p)
			}
			p.PlanSHA256, err = carrierHash(p, "plan_sha256")
			if err != nil {
				t.Fatal(err)
			}
			payload, err := Canonical(p)
			if err != nil {
				t.Fatal(err)
			}
			_, err = DecodePlan(payload)
			if tc.change == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.HasPrefix(err.Error(), projectmcp.Invalid+":") {
				t.Fatalf("semantic violation admitted: %v", err)
			}
		})
	}
}

// A valid inner record and digest cannot hide an extra outer field. This plant
// retains the original digest so accepting unknown fields would admit it.
func TestLockClosedCarrier(t *testing.T) {
	l := testLock(t)
	payload, err := Canonical(l)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeLock(payload); err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	raw["extra"] = true
	payload, err = Canonical(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeLock(payload); err == nil || !strings.HasPrefix(err.Error(), projectmcp.Invalid+":") {
		t.Fatalf("unknown carrier field admitted: %v", err)
	}
}
