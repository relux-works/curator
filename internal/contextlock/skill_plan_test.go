package contextlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/relux-works/curator/internal/protocoljson"
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

// PSP01: successor fields are forbidden by presence on historical carriers,
// including JSON null. V5 preserves strict member unions and string shapes;
// state context accepts only absent network identity fields.
func TestProfileSkillPlanStrictMemberAndLegacyPresence(t *testing.T) {
	var fixture struct{ Lock json.RawMessage }
	planFixture(t, "baseline.json", &fixture)
	baseline, err := Parse(fixture.Lock)
	if err != nil {
		t.Fatal(err)
	}
	for _, schema := range []int{1, 2, 4} {
		legacy := *baseline
		legacy.SchemaVersion = schema
		legacy.HashVersion = 2
		if schema == 1 {
			legacy.HashVersion = 1
		}
		legacy.SkillPlan = nil
		legacy.SkillPlanSHA256 = ""
		legacy.Members = append([]Member(nil), baseline.Members...)
		for i := range legacy.Members {
			if legacy.Members[i].Kind == KindSkill {
				legacy.Members[i].Directory = ""
			}
		}
		payload, err := legacy.Canonical()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Parse(payload); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"skill_plan", "skill_plan_sha256"} {
			var object map[string]any
			_ = json.Unmarshal(payload, &object)
			object[key] = nil
			tampered, _ := json.Marshal(object)
			if _, err := Parse(tampered); err == nil || !strings.Contains(err.Error(), "profile_skill_plan_invalid") {
				t.Fatalf("legacy %d accepted null %s: %v", schema, key, err)
			}
		}
	}
	for _, mode := range []string{"state-valid", "state-source-empty", "state-source-null", "state-commit-null", "network-state-null", "skill-version-null", "skill-version-empty", "noncanonical-source"} {
		t.Run(mode, func(t *testing.T) {
			var object map[string]any
			_ = json.Unmarshal(fixture.Lock, &object)
			members := object["members"].([]any)
			context := members[0].(map[string]any)
			if strings.HasPrefix(mode, "state-") {
				delete(context, "source")
				delete(context, "commit")
				delete(context, "directory")
				context["state_sha256"] = strings.Repeat("a", 64)
			}
			switch mode {
			case "state-source-empty":
				context["source"] = ""
			case "state-source-null":
				context["source"] = nil
			case "state-commit-null":
				context["commit"] = nil
			case "network-state-null":
				context["state_sha256"] = nil
			case "skill-version-null":
				members[1].(map[string]any)["version"] = nil
			case "skill-version-empty":
				members[1].(map[string]any)["version"] = ""
			case "noncanonical-source":
				context["source"] = "github.com/example/profile.git"
			}
			payload, _ := json.Marshal(object)
			_, err := Parse(payload)
			if mode == "state-valid" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "profile_skill_plan_invalid") {
				t.Fatalf("member shape refusal: %v", err)
			}
		})
	}
}

// PSP01/05: lowercase carriers preserve the literal binding; each reviewer
// alias and its duplicate-case form must refuse rather than normalize to it.
func TestProfileSkillPlanCaseAliases(t *testing.T) {
	var fixture struct {
		Lock              json.RawMessage
		RuntimeLockSHA256 string `json:"runtime_lock_sha256"`
	}
	planFixture(t, "baseline.json", &fixture)
	lock, err := Parse(fixture.Lock)
	if err != nil {
		t.Fatal(err)
	}
	if digest, err := lock.Hash(); err != nil || digest != fixture.RuntimeLockSHA256 {
		t.Fatalf("lowercase binding: %s %v", digest, err)
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(fixture.Lock))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	plan := object["skill_plan"]
	for _, tc := range []struct {
		name  string
		input any
		path  []any
		plan  bool
	}{
		{"plan-schema-version", plan, []any{"schema_version"}, true},
		{"plan-nodes", plan, []any{"nodes"}, true},
		{"node-kind", plan, []any{"nodes", 0, "kind"}, true},
		{"edge-mode", plan, []any{"edges", 0, "mode"}, true},
		{"lock-member-kind", object, []any{"members", 0, "kind"}, false},
		{"lock-plan-digest", object, []any{"skill_plan_sha256"}, false},
	} {
		for _, duplicate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/duplicate=%v", tc.name, duplicate), func(t *testing.T) {
				assertCaseAliasRefused(t, tc.input, tc.path, duplicate, tc.plan, fixture.RuntimeLockSHA256)
			})
		}
	}
	// Walk the independent published fixture, covering every recursive object
	// shape including constraints, refs, consumers and enclosing precedence.
	for _, tc := range []struct {
		name  string
		input any
		plan  bool
	}{{"plan", plan, true}, {"lock", object, false}} {
		var paths [][]any
		var walk func(any, []any)
		walk = func(value any, path []any) {
			switch value := value.(type) {
			case map[string]any:
				keys := make([]string, 0, len(value))
				for key := range value {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					next := append(append([]any{}, path...), key)
					paths = append(paths, next)
					walk(value[key], next)
				}
			case []any:
				for i, child := range value {
					walk(child, append(append([]any{}, path...), i))
				}
			}
		}
		walk(tc.input, nil)
		for _, path := range paths {
			for _, duplicate := range []bool{false, true} {
				t.Run(fmt.Sprintf("recursive-%s/%v/duplicate=%v", tc.name, path, duplicate), func(t *testing.T) {
					assertCaseAliasRefused(t, tc.input, path, duplicate, tc.plan, fixture.RuntimeLockSHA256)
				})
			}
		}
	}
}

func assertCaseAliasRefused(t *testing.T, input any, path []any, duplicate, isPlan bool, baselineHash string) {
	t.Helper()
	original, err := protocoljson.MarshalCanonical(input)
	if err != nil {
		t.Fatal(err)
	}
	if isPlan {
		if _, err := ParseSkillPlan(original); err != nil {
			t.Fatalf("lowercase control: %v", err)
		}
	} else {
		if _, err := Parse(original); err != nil {
			t.Fatalf("lowercase control: %v", err)
		}
	}
	var mutated any
	decoder := json.NewDecoder(bytes.NewReader(original))
	decoder.UseNumber()
	if err := decoder.Decode(&mutated); err != nil {
		t.Fatal(err)
	}
	parent := mutated
	for _, step := range path[:len(path)-1] {
		switch step := step.(type) {
		case string:
			parent = parent.(map[string]any)[step]
		case int:
			parent = parent.([]any)[step]
		}
	}
	object := parent.(map[string]any)
	key := path[len(path)-1].(string)
	object[strings.ToUpper(key)] = object[key]
	if !duplicate {
		delete(object, key)
	}
	payload, err := protocoljson.MarshalCanonical(mutated)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(payload, original) {
		t.Fatal("alias was not planted")
	}
	if !isPlan && HashBytes(payload) == baselineHash {
		t.Fatal("mutated input retained baseline digest")
	}
	if isPlan {
		_, err = ParseSkillPlan(payload)
	} else {
		_, err = Parse(payload)
	}
	if err == nil || !strings.Contains(err.Error(), "exact JSON key") {
		t.Fatalf("want exact JSON key refusal, got %v", err)
	}
}
