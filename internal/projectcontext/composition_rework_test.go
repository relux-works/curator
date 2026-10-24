package projectcontext

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Claim: equivalent empty and reordered selection requests preserve content-addressed publication.
func TestCompositionPublicationSelections(t *testing.T) {
	for _, scenario := range []string{"empty", "reordered"} {
		t.Run(scenario, func(t *testing.T) {
			f := compositionModeFixture(t, "codex_cli", nil, "native-project")
			d := Declaration{Version: 1, Items: []Request{}}
			req := compositionRequest(f)
			req.Selections = nil
			if scenario == "reordered" {
				for _, name := range []string{"beta", "alpha"} {
					src := compositionSkillSource(t, name, "")
					copyCompositionSkill(t, src, filepath.Join(f.root, name))
					d.Items = append(d.Items, Request{ID: name, Type: "skill", Environment: "codex_cli", Required: false, Source: name, Dependencies: map[string]string{}})
					req.Selections = append(req.Selections, Selection{name, false})
				}
			}
			approveCompositionDeclaration(t, f, d)
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			a := Authority{Store: f.store, CWD: f.root}
			owner := principal(t, f)
			first, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
			if err != nil {
				t.Fatal(err)
			}
			// An unchanged request is a positive control for immutable publication.
			if same, err := a.PublishComposition(context.Background(), req, p.SHA256, owner); err != nil || !reflect.DeepEqual(first, same) {
				t.Fatalf("unchanged publication failed: %v", err)
			}
			other := req
			if scenario == "empty" {
				other.Selections = []Selection{}
			} else {
				other.Selections = []Selection{req.Selections[1], req.Selections[0]}
			}
			q, err := f.store.PreviewComposition(context.Background(), other)
			if err != nil {
				t.Fatal(err)
			}
			if p.SHA256 != q.SHA256 {
				t.Fatal("equivalent selections unexpectedly changed composition")
			}
			repeated, err := a.PublishComposition(context.Background(), other, q.SHA256, owner)
			if err != nil {
				t.Errorf("same composition identity rejected during repeat publication: %v", err)
			} else if !reflect.DeepEqual(first, repeated) {
				t.Error("repeat result differs")
			}
			if scenario == "reordered" {
				// Disabling an optional item is a real selection change: the old
				// preview cannot authorize it, and refusal preserves stored bytes.
				path := filepath.Join(f.store.Home, "project-context/compositions", first.SHA256+".json")
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				different := req
				different.Selections = append([]Selection{}, req.Selections...)
				different.Selections[0].Disabled = true
				changed, err := f.store.PreviewComposition(context.Background(), different)
				if err != nil {
					t.Fatal(err)
				}
				if changed.SHA256 == p.SHA256 {
					t.Fatal("different selection retained old digest")
				}
				_, err = a.PublishComposition(context.Background(), different, p.SHA256, owner)
				required(t, err, "project_context_drift")
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("selection refusal changed retained metadata")
				}
			}
			got, err := resolveComposition(t, a, first.Ref, first.SHA256)
			if err != nil || !reflect.DeepEqual(first, got) {
				t.Errorf("original publication stopped resolving: %v", err)
			}
		})
	}
}

// Claim: companion-only and manifest-only changes alter the complete command
// artifact identity; unchanged skills retain it, and the entry receipt keeps
// its independent byte digest when only a companion changes.
func TestCompositionCommandArtifact(t *testing.T) {
	f := compositionModeFixture(t, "codex_cli", nil, "native-project")
	src := compositionSkillSource(t, "runner", "run")
	copyCompositionSkill(t, src, filepath.Join(f.root, "runner"))
	helper := filepath.Join(f.root, "runner/scripts/helper.py")
	write(t, helper, "VALUE = 1\n")
	write(t, filepath.Join(f.root, "runner/scripts/tool.py"), "from helper import VALUE\nprint(VALUE)\n")
	d := Declaration{Version: 1, Items: []Request{{ID: "runner", Type: "skill", Environment: "codex_cli", Required: true, Source: "runner", Dependencies: map[string]string{}}}}
	approveCompositionDeclaration(t, f, d)
	req := compositionRequest(f)
	req.Selections = []Selection{{"runner", false}}
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	same, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Record.CommandOwners, same.Record.CommandOwners) {
		t.Fatal("positive control: unchanged skill changed command owners")
	}
	write(t, helper, "VALUE = 2\n")
	approveCompositionDeclaration(t, f, d)
	q, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if p.Record.GenerationSHA256 == q.Record.GenerationSHA256 {
		t.Fatal("positive control: changed skill generation was not observed")
	}
	if p.Record.CommandOwners[0].ArtifactSHA256 == q.Record.CommandOwners[0].ArtifactSHA256 {
		t.Error("changed audited command artifact has unchanged command-owner artifact hash")
	}
	entryPath := ".agent-context/commands/run"
	var entryBefore, entryAfter File
	for _, file := range p.Receipt.Files {
		if file.Path == entryPath {
			entryBefore = file
		}
	}
	for _, file := range q.Receipt.Files {
		if file.Path == entryPath {
			entryAfter = file
		}
	}
	if entryBefore.Path == "" || entryBefore != entryAfter || entryAfter.SHA256 != digest(q.Files[entryPath]) {
		t.Fatal("unchanged entry materialization hash was lost")
	}
	if q.Record.CommandOwners[0].ArtifactSHA256 != q.Skills[0].ArtifactSHA256 {
		t.Error("command owner does not bind complete audited skill")
	}
	// A metadata-only change must bind the same complete artifact too.
	manifest := filepath.Join(f.root, "runner/agent-skill.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	write(t, manifest, string(raw)+"\n")
	approveCompositionDeclaration(t, f, d)
	r, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Record.CommandOwners[0].ArtifactSHA256 == q.Record.CommandOwners[0].ArtifactSHA256 {
		t.Error("changed audited manifest has unchanged command-owner artifact hash")
	}
	t.Logf("full receipts equal=%t; command hash equal=%t", p.Record.MaterializationSHA256 == q.Record.MaterializationSHA256, p.Record.CommandOwners[0].ArtifactSHA256 == q.Record.CommandOwners[0].ArtifactSHA256)
}

// Claim: inspection separates the registered source mode from admission off;
// only native-project includes the mandatory isolation notice, before and after
// an empty approved review. Inspection does not mutate protected state.
func TestCompositionInspectionStatus(t *testing.T) {
	for _, mode := range []string{"strict", "native-project"} {
		t.Run(mode, func(t *testing.T) {
			f := compositionModeFixture(t, "codex_cli", nil, mode)
			for _, reviewed := range []bool{false, true} {
				if reviewed {
					approveCompositionDeclaration(t, f, Declaration{Version: 1, Items: []Request{}})
				}
				path := filepath.Join(f.store.Home, "project-context/state.json")
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				out, err := f.store.Inspect(f.registration.CheckoutID)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(out)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]any
				if err = json.Unmarshal(raw, &fields); err != nil {
					t.Fatal(err)
				}
				if fields["mode"] != mode || fields["admission"] != "off" || fields["memory"] != "off" {
					t.Errorf("inspection conflates source mode with admission: %s", raw)
				}
				notice := strings.Contains(string(raw), "project sources not isolated")
				if notice != (mode == "native-project") || (mode == "native-project" && fields["notice"] != "project sources not isolated") {
					t.Errorf("inspection isolation notice differs: %s", raw)
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("inspection changed protected state")
				}
			}
		})
	}
}

// Claim: a legitimate repeat publication passes, but mismatched retained
// supporting metadata refuses without an overwrite; restored valid bytes
// publish and resolve again. Canonicalization must preserve this boundary.
func TestCompositionPublicationMetadata(t *testing.T) {
	f := compositionModeFixture(t, "codex_cli", nil, "native-project")
	approveCompositionDeclaration(t, f, Declaration{Version: 1, Items: []Request{}})
	req := compositionRequest(f)
	req.Selections = nil
	p, err := f.store.PreviewComposition(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	a := Authority{Store: f.store, CWD: f.root}
	owner := principal(t, f)
	c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.store.Home, "project-context/compositions", c.SHA256+".json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.PublishComposition(context.Background(), req, p.SHA256, owner); err != nil {
		t.Fatal(err)
	}
	var saved compositionEnvelope
	if err = json.Unmarshal(original, &saved); err != nil {
		t.Fatal(err)
	}
	saved.BuildKeys = append(saved.BuildKeys, "unexpected-build")
	changed, err := canonical(saved)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	_, err = a.PublishComposition(context.Background(), req, p.SHA256, owner)
	required(t, err, "immutable composition metadata changed")
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, changed) {
		t.Fatal("publication overwrote mismatched immutable metadata")
	}
	if err = os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = a.PublishComposition(context.Background(), req, p.SHA256, owner); err != nil {
		t.Fatal(err)
	}
	if _, err = resolveComposition(t, a, c.Ref, c.SHA256); err != nil {
		t.Fatal(err)
	}
}
