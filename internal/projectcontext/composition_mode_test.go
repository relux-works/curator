package projectcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/relux-works/curator/internal/envregistry"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Claim: synthetic contracts advertise modes independently of surfaces; native
// mode needs permissions/memory-off but only strict needs native exclusion.
// This tests contract admission, not real provider certification.
func TestCompositionAdapterModes(t *testing.T) {
	a := AdapterContract{Kind: "project-adapter/1", Environment: "codex_cli", ProviderVersion: "1.2.3", Platform: "macos", Modes: []string{"native-project", "strict"}, Surfaces: []string{"context", "memory-off", "permissions"}}
	if err := a.validate("codex_cli", "native-project"); err != nil {
		t.Fatalf("native mode incorrectly requires exclusion: %v", err)
	}
	required(t, a.validate("codex_cli", "strict"), "native-source-exclusion")
	for _, surface := range []string{"permissions", "memory-off"} {
		bad := a
		bad.Surfaces = nil
		for _, s := range a.Surfaces {
			if s != surface {
				bad.Surfaces = append(bad.Surfaces, s)
			}
		}
		required(t, bad.validate("codex_cli", "native-project"), surface)
	}
	for _, mode := range []string{"", "off", "admitted", "future"} {
		required(t, a.validate("codex_cli", mode), "project_context_invalid")
	}
	for _, modes := range [][]string{nil, {"strict", "native-project"}, {"native-project", "native-project"}, {"future"}} {
		bad := a
		bad.Modes = modes
		required(t, bad.validate("codex_cli", "native-project"), "project_context_invalid")
	}
	a.Modes = []string{"strict"}
	required(t, a.validate("codex_cli", "native-project"), "project_context_unsupported")
	a.Modes = []string{}
	required(t, a.validate("codex_cli", "strict"), "project_context_unsupported")
	a.Modes = []string{"native-project", "strict"}
	a.Surfaces = []string{"context", "memory-off", "native-source-exclusion", "permissions"}
	if err := a.validate("codex_cli", "strict"); err != nil {
		t.Fatal(err)
	}
}

// Claim: both published mode values are accepted and hashed as part of the
// complete record; missing/null/unknown source modes fail closed. The pinned
// schema independently rejects those same plants, with a restored valid control.
func TestCompositionModeRecord(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "composition", "bytes", "composition-valid.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err = json.Unmarshal(b, &record); err != nil {
		t.Fatal(err)
	}
	schema := compositionSchemas(t)["composition"]
	var hashes []string
	for _, mode := range []string{"strict", "native-project"} {
		record["mode"] = mode
		b, _ := canonical(record)
		parsed, err := ParseCompositionRecord(b)
		if err != nil {
			t.Fatal(err)
		}
		if err = schema.Validate(record); err != nil {
			t.Fatal(err)
		}
		hashes = append(hashes, identity(parsed))
	}
	if hashes[0] == hashes[1] {
		t.Fatal("mode omitted from digest")
	}
	for _, mode := range []any{nil, "off", "admitted", "future"} {
		record["mode"] = mode
		b, _ := canonical(record)
		_, err = ParseCompositionRecord(b)
		required(t, err, "project_context_invalid")
		if err = schema.Validate(record); err == nil {
			t.Fatal("schema accepted invalid mode")
		}
	}
	delete(record, "mode")
	b, _ = canonical(record)
	_, err = ParseCompositionRecord(b)
	required(t, err, "project_context_invalid")
	if err = schema.Validate(record); err == nil {
		t.Fatal("schema accepted missing mode")
	}
	record["mode"] = "strict"
	if err = schema.Validate(record); err != nil {
		t.Fatal(err)
	}
}

// Claim: native managed root/system bytes and their separate inventories come
// solely from the protected profile, while approved whole skills remain delivered.
// Publication and held resolution preserve the mode-bound identity without a lease.
func TestCompositionNativeProfileOnly(t *testing.T) {
	for _, env := range []string{"codex_cli", "claude_code", "muse"} {
		t.Run(env, func(t *testing.T) {
			baseline := compositionFixture(t, env, nil, env != "muse")
			approveCompositionDeclaration(t, baseline, Declaration{Version: 1, Items: []Request{}})
			baselineReq := compositionRequest(baseline)
			baselineReq.Selections = nil
			strict, err := baseline.store.PreviewComposition(context.Background(), baselineReq)
			if err != nil {
				t.Fatal(err)
			}
			f := compositionModeFixture(t, env, nil, "native-project", env != "muse")
			write(t, filepath.Join(f.root, "AGENTS.md"), "Native repository sentinel.\n")
			src := compositionSkillSource(t, "project-skill", "")
			copyCompositionSkill(t, src, filepath.Join(f.root, "project-skill"))
			d := Declaration{Version: 1, Items: []Request{{ID: "project-skill", Type: "skill", Environment: env, Required: true, Source: "project-skill", Dependencies: map[string]string{}}}}
			approveCompositionDeclaration(t, f, d)
			req := compositionRequest(f)
			req.Selections = []Selection{{"project-skill", false}}
			var surfaces []string
			for _, s := range req.Adapter.Surfaces {
				if s != "native-source-exclusion" {
					surfaces = append(surfaces, s)
				}
			}
			req.Adapter.Surfaces = surfaces
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if p.Record.Mode != "native-project" || p.SHA256 != identity(p.Record) {
				t.Fatal("mode omitted from digest")
			}
			if !reflect.DeepEqual(strict.ProfileDocuments, p.ProfileDocuments) || !reflect.DeepEqual(strict.ProfileDocumentReceipt, p.ProfileDocumentReceipt) {
				t.Fatal("profile-only bytes/inventory changed by source mode")
			}
			adapter, _ := envregistry.ByID(env)
			if !bytes.Equal(p.Files[adapter.RootTarget], p.ProfileDocuments[adapter.RootTarget]) || len(p.Files[adapter.RootTarget]) == 0 {
				t.Fatal("native profile root differs")
			}
			for path, b := range p.ProfileDocuments {
				if !bytes.Equal(b, p.Files[path]) {
					t.Fatal("native profile document differs")
				}
			}
			receipt, err := compositionReceipt(p.ProfileDocuments, map[string]bool{}, p.Record.AdapterSHA256, p.Record.PolicySHA256, p.Record.WorktreeRef)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(receipt.Files, p.ProfileDocumentReceipt) {
				t.Fatal("profile inventory differs")
			}
			if !bytes.Equal(p.Files[filepath.ToSlash(filepath.Join(adapter.SkillsDir, "project-skill", "SKILL.md"))], []byte(compositionSkillDocument("project-skill", "# project-skill\n"))) {
				t.Fatal("approved whole-skill context bytes lost")
			}
			if len(p.Skills) != 1 || p.Skills[0].Layer != "project" {
				t.Fatal("whole skill delivery lost")
			}
			if env != "muse" && !bytes.Contains(p.ProfileDocuments[".agent-context/system-prompt.md"], []byte("Base system.")) {
				t.Fatal("profile system lost")
			}
			wire, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			var encoded map[string]json.RawMessage
			if err = json.Unmarshal(wire, &encoded); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"Files", "ProfileDocuments", "ProfileDocumentReceipt"} {
				if _, ok := encoded[field]; ok {
					t.Fatalf("preview serialized non-wire %s", field)
				}
			}
			assertCompositionSchemas(t, req, p)
			b, _ := canonical(f.registration)
			var value any
			_ = json.Unmarshal(b, &value)
			if err = compositionSchemas(t)["registration"].Validate(value); err != nil {
				t.Fatal(err)
			}
			a := Authority{Store: f.store, CWD: f.root}
			c, err := a.PublishComposition(context.Background(), req, p.SHA256, principal(t, f))
			if err != nil {
				t.Fatal(err)
			}
			explicit := req
			explicit.Mode = "native-project"
			repeated, err := a.PublishComposition(context.Background(), explicit, p.SHA256, principal(t, f))
			if err != nil || !reflect.DeepEqual(c, repeated) {
				t.Fatalf("equivalent explicit mode changed publication: %v", err)
			}
			resolved, err := resolveComposition(t, a, c.Ref, c.SHA256)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(c, resolved) {
				t.Fatal("mode-bound publication differs")
			}
		})
	}
}

// Claim: selecting required or optional project context in native mode refuses
// with no generated output or publication; an explicitly disabled optional item
// has only the existing disabled omission, and strict still renders additively.
func TestCompositionNativeContextRefusal(t *testing.T) {
	for _, optional := range []bool{false, true} {
		t.Run(fmt.Sprint(optional), func(t *testing.T) {
			f := compositionModeFixture(t, "codex_cli", nil, "native-project")
			packageApproved(t, f, optional)
			before, err := os.ReadFile(filepath.Join(f.store.Home, "project-context", "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			req := compositionRequest(f)
			p, err := f.store.PreviewComposition(context.Background(), req)
			required(t, err, "native-project cannot select context items")
			if len(p.Files) != 0 || p.SHA256 != "" {
				t.Fatal("failed selection returned output")
			}
			a := Authority{Store: f.store, CWD: f.root}
			_, err = a.PublishComposition(context.Background(), req, strings.Repeat("a", 64), principal(t, f))
			required(t, err, "native-project cannot select context items")
			after, err := os.ReadFile(filepath.Join(f.store.Home, "project-context", "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("refusal changed protected state")
			}
			if _, err = os.Stat(filepath.Join(f.store.Home, "project-context", "compositions")); !os.IsNotExist(err) {
				t.Fatal("refusal published composition")
			}
			if optional {
				req.Selections[0].Disabled = true
				p, err = f.store.PreviewComposition(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				if len(p.Generation.Items) != 0 || len(p.Record.Omitted) != 1 || p.Record.Omitted[0].Reason != "disabled" {
					t.Fatal("implicit native omission")
				}
			}
		})
	}
	f := compositionFixture(t, "codex_cli", nil)
	packageApproved(t, f, false)
	req := compositionRequest(f)
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(p.Files["AGENTS.md"], []byte("Selected project context.")) || bytes.Contains(p.ProfileDocuments["AGENTS.md"], []byte("Selected project context.")) {
		t.Fatal("strict additive/profile split differs")
	}
	req.Mode = "native-project"
	_, err = f.store.PreviewComposition(context.Background(), req)
	required(t, err, "project_context_invalid")
}

// Claim: registration defaults to strict, native mode needs the explicit
// operator API, unknown input causes no store write, and missing/null/unknown retained mode
// fails closed rather than enabling native discovery or upgrading old authority.
func TestCompositionRegistrationMode(t *testing.T) {
	f := newFixture(t)
	if f.registration.Mode != "strict" {
		t.Fatal("registration did not default strict")
	}
	state := filepath.Join(f.store.Home, "project-context", "state.json")
	before, err := os.ReadFile(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"", "off", "admitted", "future"} {
		_, err = f.store.RegisterMode("other", f.root, "base", "codex_cli", mode)
		required(t, err, "project_context_invalid")
	}
	after, err := os.ReadFile(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("unknown mode changed store")
	}
	native := compositionModeFixture(t, "codex_cli", nil, "native-project")
	if native.registration.Mode != "native-project" {
		t.Fatal("explicit mode was lost")
	}
	var value map[string]any
	if err = json.Unmarshal(before, &value); err != nil {
		t.Fatal(err)
	}
	registrations := value["registrations"].(map[string]any)
	registration := registrations[f.registration.CheckoutID].(map[string]any)["record"].(map[string]any)
	for _, mode := range []any{nil, "", "off", "admitted", "future"} {
		registration["mode"] = mode
		damaged, _ := canonical(value)
		if err = os.WriteFile(state, damaged, 0600); err != nil {
			t.Fatal(err)
		}
		_, err = f.store.Locate(f.root)
		if mode == nil {
			required(t, err, "project_context_invalid")
		} else {
			required(t, err, "project_context_drift")
		}
		if mode != nil {
			required(t, err, "invalid protected registration")
		}
	}
	delete(registration, "mode")
	damaged, _ := canonical(value)
	if err = os.WriteFile(state, damaged, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Locate(f.root)
	required(t, err, "project_context_invalid")
	if err = os.WriteFile(state, before, 0600); err != nil {
		t.Fatal(err)
	}
	restored, err := f.store.Locate(f.root)
	if err != nil || restored.Mode != "strict" {
		t.Fatalf("restored strict registration refused: %v", err)
	}
}
