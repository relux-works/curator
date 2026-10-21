package contextpkg_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/stateread"
)

// Read must reject non-number discriminators in canonical on-disk schema-4
// carriers as unusable state, alongside a passing numeric control per field.
func TestReviewSkillDirectoryNumericDiscriminants(t *testing.T) {
	testLockNumericDiscriminants(t, 4)
}

// Read must enforce numeric weights for both context and selected skill members;
// refusal returns no usable lock/hash and preserves the malformed file.
func TestReviewSkillDirectoryStringWeight(t *testing.T) {
	testLockNumericWeights(t, 4)
}

// Frozen schemas also require JSON numbers: valid numeric carriers remain
// readable while quoted numbers and other non-numbers cannot become usable state.
func TestLegacyLockNumericCarrierTypes(t *testing.T) {
	for _, schema := range []int{1, 2} {
		t.Run(fmt.Sprintf("schema-%d", schema), func(t *testing.T) {
			testLockNumericDiscriminants(t, schema)
			testLockNumericWeights(t, schema)
		})
	}
}

func testLockNumericDiscriminants(t *testing.T, schema int) {
	t.Helper()
	for _, field := range []string{"schema_version", "hash_version"} {
		if schema == 1 && field == "hash_version" {
			continue // Schema 1 has no hash_version field.
		}
		t.Run(field, func(t *testing.T) { testLockNumericField(t, schema, field, -1) })
	}
}

func testLockNumericWeights(t *testing.T, schema int) {
	t.Helper()
	for index, kind := range []string{"context", "skill"} {
		t.Run("weight-"+kind, func(t *testing.T) { testLockNumericField(t, schema, "weight", index) })
	}
}

func testLockNumericField(t *testing.T, schema int, field string, memberIndex int) {
	t.Helper()
	for _, shape := range []string{"numeric", "string", "null", "boolean", "array", "object", "missing"} {
		t.Run(shape, func(t *testing.T) {
			payload, err := os.ReadFile("testdata/profile-skill-directories-v1/fixtures.json")
			if err != nil {
				t.Fatal(err)
			}
			var fixture struct {
				Lock map[string]any `json:"candidate_lock"`
			}
			decoder := json.NewDecoder(bytes.NewReader(payload))
			decoder.UseNumber()
			if err := decoder.Decode(&fixture); err != nil {
				t.Fatal(err)
			}
			carrier := fixture.Lock
			carrier["schema_version"] = schema
			members := carrier["members"].([]any)
			if schema != 4 {
				delete(carrier, "codex_settings")
				delete(carrier, "codex_settings_contributions")
				delete(carrier, "precedence")
				delete(members[1].(map[string]any), "directory")
			}
			if schema == 1 {
				delete(carrier, "hash_version")
			}
			target := carrier
			if memberIndex >= 0 {
				target = members[memberIndex].(map[string]any)
			}
			numeric := target[field]
			switch shape {
			case "numeric":
				// Keep the normative fixture's number as the positive control.
			case "string":
				target[field] = fmt.Sprint(numeric)
			case "null":
				target[field] = nil
			case "boolean":
				target[field] = false
			case "array":
				target[field] = []any{numeric}
			case "object":
				target[field] = map[string]any{"number": numeric}
			case "missing":
				delete(target, field)
			}
			payload, err = protocoljson.MarshalCanonical(carrier)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "lock.json")
			if err := os.WriteFile(path, payload, 0600); err != nil {
				t.Fatal(err)
			}
			lock, hash, err := contextlock.Read(path)
			if shape == "numeric" {
				if err != nil || lock == nil || lock.SchemaVersion != schema || hash != contextlock.HashBytes(payload) {
					t.Fatalf("numeric positive control refused or changed: lock=%+v hash=%s error=%v", lock, hash, err)
				}
			} else {
				var state *stateread.Error
				if err == nil {
					t.Fatalf("lock Read admitted %s %s; numeric carrier required", shape, field)
				}
				if lock != nil || hash != "" || !errors.As(err, &state) || state.Kind != stateread.KindUnreadable || state.Path != path || !strings.Contains(err.Error(), "invalid lock:") || !strings.Contains(err.Error(), field) {
					t.Fatalf("wrong numeric carrier refusal: lock=%+v hash=%s error=%v", lock, hash, err)
				}
			}
			if retained, err := os.ReadFile(path); err != nil || !bytes.Equal(retained, payload) {
				t.Fatalf("lock Read changed the on-disk carrier: %v", err)
			}
		})
	}
}

// Neighboring closed/type obligations survive numeric-token repair through Read.
func TestReviewNumericRepairNeighbors(t *testing.T) {
 for _,tc:=range []string{"positive", "unknown-top", "unknown-member", "null-kind", "null-name", "null-source", "null-commit", "null-required-by", "null-overlay", "null-precedence", "unknown-precedence"} {
 t.Run(tc,func(t *testing.T){
 b,e:=os.ReadFile("testdata/profile-skill-directories-v1/fixtures.json");if e!=nil{t.Fatal(e)}
 var fixture struct{ Lock map[string]any `json:"candidate_lock"` };d:=json.NewDecoder(bytes.NewReader(b));d.UseNumber();if e=d.Decode(&fixture);e!=nil{t.Fatal(e)}
 c:=fixture.Lock;m:=c["members"].([]any)[1].(map[string]any)
 switch tc {
 case "unknown-top":c["extra"]=0
 case "unknown-member":m["extra"]=0
 case "null-kind":m["kind"]=nil
 case "null-name":m["name"]=nil
 case "null-source":m["source"]=nil
 case "null-commit":m["commit"]=nil
 case "null-required-by":m["required_by"]=nil
 case "null-overlay":m["overlay"]=nil
 case "null-precedence":c["precedence"]=nil
 case "unknown-precedence":c["precedence"].(map[string]any)["extra"]=0
 }
 b,e=protocoljson.MarshalCanonical(c);if e!=nil{t.Fatal(e)};path:=filepath.Join(t.TempDir(),"lock.json");if e=os.WriteFile(path,b,0600);e!=nil{t.Fatal(e)}
 lock,hash,e:=contextlock.Read(path)
 if tc=="positive" {if e!=nil||lock==nil||hash==""{t.Fatalf("valid control refused: %v",e)}} else {var state *stateread.Error;if e==nil||lock!=nil||hash!=""||!errors.As(e,&state)||state.Kind!=stateread.KindUnreadable {t.Fatalf("malformed carrier became usable: %v",e)}}
 after,e:=os.ReadFile(path);if e!=nil||!bytes.Equal(after,b){t.Fatal("reader changed input")}
 })
 }
}
