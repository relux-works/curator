package contextlock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func planFixture(t *testing.T, name string, out any) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata/profile-skill-plan-v1", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(payload, out); err != nil {
		t.Fatal(err)
	}
}

// PSP01/03/05: the Go reader accepts the published positive shapes and refuses
// each paired structural violation; legacy shapes remain frozen.
func TestProfileSkillPlanSchemaCorpus(t *testing.T) {
	var cases []struct {
		ID, Schema string
		Instance   json.RawMessage
		Valid      bool
	}
	planFixture(t, "schema-cases.json", &cases)
	if len(cases) != 42 {
		t.Fatalf("cases=%d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.ID, func(t *testing.T) {
			var err error
			if strings.HasSuffix(tc.Schema, "context-lock-v5.schema.json") || strings.Contains(tc.Schema, "context-lock-v") {
				_, err = Parse(tc.Instance)
			} else {
				_, err = ParseSkillPlan(tc.Instance)
			}
			if (err == nil) != tc.Valid {
				t.Fatalf("valid=%v: %v", tc.Valid, err)
			}
		})
	}
}

// PSP05: independent literal fixture digests bind every portable field, while
// a forged digest and an old external binding are refused without writes.
func TestProfileSkillPlanLiteralBindings(t *testing.T) {
	var fixture struct {
		Lock              json.RawMessage
		RuntimeLockSHA256 string `json:"runtime_lock_sha256"`
	}
	planFixture(t, "baseline.json", &fixture)
	lock, err := Parse(fixture.Lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.VerifySkillBinding(fixture.RuntimeLockSHA256); err != nil {
		t.Fatal(err)
	}
	if err := lock.VerifySkillBinding("sha256:" + strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), "profile_skill_plan_binding_invalid") {
		t.Fatalf("binding: %v", err)
	}
	before, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	round, err := Parse(before)
	if err != nil || !reflect.DeepEqual(round, lock) {
		t.Fatalf("roundtrip: %v", err)
	}
	lock.SkillPlanSHA256 = "sha256:" + strings.Repeat("0", 64)
	if _, err := lock.Canonical(); err == nil || !strings.Contains(err.Error(), "profile_skill_plan_binding_invalid") {
		t.Fatalf("forged digest: %v", err)
	}
}

// PSP02/03/05 subset: semantic portable-graph counterexamples are run through
// the Go reader, after the corpus's explicit rebind when requested. Package
// agreement and activation are exercised through shared closure tests.
func TestProfileSkillPlanPortableVectors(t *testing.T) {
	selected := map[string]bool{"runtime-narrow-transitive": true, "mode-full": true, "mode-context": true, "mode-runtime": true, "default-mode-full": true, "overlap-full-union": true, "overlap-alias-ref": true, "missing-node": true, "duplicate-node": true, "node-order": true, "missing-member": true, "member-directory": true, "wrong-required-by": true, "duplicate-edge": true, "wrong-consumer": true, "node-extra-field": true, "unreachable-extra-node": true, "selection-empty": true, "selection-duplicate": true, "selection-unsorted": true, "bad-plan-digest": true, "bad-runtime-binding": true, "tampered-selection-digest": true, "revision-pin-disagreement": true}
	var baseline map[string]any
	planFixture(t, "baseline.json", &baseline)
	var cases []struct {
		ID    string
		Patch []struct {
			Path   []any
			Value  any
			Delete bool
		}
		Rebind   bool
		Expected struct{ Diagnostic *string }
	}
	planFixture(t, "cases.json", &cases)
	count := 0
	for _, tc := range cases {
		if !selected[tc.ID] {
			continue
		}
		count++
		t.Run(tc.ID, func(t *testing.T) {
			payload, _ := json.Marshal(baseline)
			var fixture map[string]any
			_ = json.Unmarshal(payload, &fixture)
			for _, edit := range tc.Patch {
				applyPlanPatch(fixture, edit.Path, edit.Value, edit.Delete)
			}

			object := fixture["lock"].(map[string]any)
			if tc.Rebind {
				b, _ := json.Marshal(object["skill_plan"])
				var p SkillPlan
				_ = json.Unmarshal(b, &p)
				digest, err := p.Hash()
				if err != nil {
					t.Fatal(err)
				}
				object["skill_plan_sha256"] = digest
			}
			b, _ := json.Marshal(object)
			lock, err := Parse(b)
			if tc.ID == "bad-runtime-binding" && err == nil {
				err = lock.VerifySkillBinding(fixture["runtime_lock_sha256"].(string))
			}
			if (err == nil) != (tc.Expected.Diagnostic == nil) {
				t.Fatalf("expected=%v err=%v", tc.Expected.Diagnostic, err)
			}
			if err != nil && tc.Expected.Diagnostic != nil && !strings.Contains(err.Error(), *tc.Expected.Diagnostic) {
				t.Fatalf("want diagnostic %s, got %v", *tc.Expected.Diagnostic, err)
			}
		})
	}
	if count != len(selected) {
		t.Fatalf("portable vector selection: got %d, want %d", count, len(selected))
	}
}

func applyPlanPatch(parent any, path []any, value any, remove bool) any {
	key := path[0]
	switch typed := parent.(type) {
	case map[string]any:
		name := key.(string)
		if len(path) == 1 {
			if remove {
				delete(typed, name)
			} else {
				typed[name] = value
			}
		} else {
			typed[name] = applyPlanPatch(typed[name], path[1:], value, remove)
		}
	case []any:
		index := int(key.(float64))
		if len(path) == 1 {
			if remove {
				return append(typed[:index], typed[index+1:]...)
			}
			typed[index] = value
		} else {
			typed[index] = applyPlanPatch(typed[index], path[1:], value, remove)
		}
	}
	return parent
}

// Exact tag grammar is shared with the protocol's bounded Git ref scalar.
// Valid slash/Unicode/leading-dash names pair with length/control violations.
func TestProfileSkillPlanExactTagGrammar(t *testing.T) {
	for _, tc := range []struct {
		tag   string
		valid bool
	}{{"v1.0.0", true}, {"release/v1", true}, {"vå", true}, {strings.Repeat("a", 255), true}, {strings.Repeat("a", 256), false}, {"-option", true}, {"a>b", false}, {"a\u00a0b", false}, {"v\x1f", false}, {"v\x7f", false}, {"v..1", false}, {"a/.hidden", false}, {"a.lock/x", false}} {
		if got := validSkillRef(SkillRef{Kind: "tag", Value: tc.tag}, ""); got != tc.valid {
			t.Errorf("tag %q: %v", tc.tag, got)
		}
	}
}
