package buildmeta

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

// policyVectors is the execution-policy slice of the authoritative build-driver
// suite. Every expected value is read from the suite.
type policyVectors struct {
	CacheIdentity struct {
		Aliases  bool `json:"aliases"`
		Portable struct {
			CacheKey        string          `json:"cache_key"`
			ExecutionPolicy string          `json:"execution_policy"`
			SchemaValid     bool            `json:"schema_valid"`
			Input           json.RawMessage `json:"input"`
		} `json:"portable"`
		ReservedHardened struct {
			CacheKey    string          `json:"cache_key"`
			SchemaValid bool            `json:"schema_valid"`
			Input       json.RawMessage `json:"input"`
		} `json:"reserved_hardened"`
		Legacy struct {
			CacheKey    string          `json:"cache_key"`
			SchemaValid bool            `json:"schema_valid"`
			Input       json.RawMessage `json:"input"`
		} `json:"legacy_rc4_without_execution_policy"`
	} `json:"cache_identity"`
	PositiveCases []struct {
		Name              string `json:"name"`
		Result            string `json:"result"`
		CacheKey          string `json:"cache_key"`
		ExecutionPolicy   string `json:"execution_policy"`
		PolicyField       string `json:"policy_field"`
		PackageSelectable bool   `json:"package_selectable"`
	} `json:"positive_cases"`
	RejectionCases []struct {
		Name     string `json:"name"`
		Boundary string `json:"boundary"`
		Input    struct {
			BuildInput      json.RawMessage `json:"build_input"`
			DerivedCacheKey string          `json:"derived_cache_key"`
		} `json:"input"`
		Expected struct {
			Result                  string `json:"result"`
			Error                   string `json:"error"`
			Reuse                   bool   `json:"reuse"`
			SchemaValid             bool   `json:"schema_valid"`
			CacheLookupPerformed    bool   `json:"cache_lookup_performed"`
			AliasesPortableCacheKey bool   `json:"aliases_portable_cache_key"`
			ArtifactExecuted        bool   `json:"artifact_executed"`
		} `json:"expected"`
	} `json:"rejection_cases"`
}

func loadPolicyVectors(t *testing.T) policyVectors {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors policyVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.CacheIdentity.Portable.Input) == 0 {
		t.Skipf("%s publishes no portable execution-policy cache identity", root)
	}
	return vectors
}

// canonicalInputBytes re-encodes a published build input as exact CCJ-1, which
// is the only form Curator's strict decoder accepts. Only the framing is
// normalised; no field value is changed.
func canonicalInputBytes(t *testing.T, raw json.RawMessage) []byte {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	payload, err := protocoljson.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestPortableExecutionPolicyIsTheOnlyAdmittedBuildInput proves the published
// portable input is accepted and independently derives the published key, and
// that both published execution-policy rejection vectors are refused by the
// strict decoder before any cache lookup can be attempted.
func TestPortableExecutionPolicyIsTheOnlyAdmittedBuildInput(t *testing.T) {
	vectors := loadPolicyVectors(t)
	if vectors.CacheIdentity.Aliases {
		t.Fatal("the suite declares aliasing cache identities")
	}

	t.Run("portable-execution-policy-is-required-input", func(t *testing.T) {
		var positive struct {
			result, cacheKey, policy, field string
			selectable, found               bool
		}
		for _, testCase := range vectors.PositiveCases {
			if testCase.Name == "portable-execution-policy-is-required-input" {
				positive.result, positive.cacheKey = testCase.Result, testCase.CacheKey
				positive.policy, positive.field = testCase.ExecutionPolicy, testCase.PolicyField
				positive.selectable, positive.found = testCase.PackageSelectable, true
			}
		}
		if !positive.found {
			t.Skip("this root publishes no portable-execution-policy positive vector")
		}
		if positive.result != "accepted" || positive.selectable {
			t.Fatalf("positive vector = %+v", positive)
		}
		if positive.field != "policy.execution_policy" {
			t.Fatalf("policy field = %q", positive.field)
		}
		if FixedPolicy().ExecutionPolicy != positive.policy {
			t.Fatalf("Curator fixed policy = %q, vector = %q", FixedPolicy().ExecutionPolicy, positive.policy)
		}
		input, err := DecodeInput(canonicalInputBytes(t, vectors.CacheIdentity.Portable.Input))
		if err != nil {
			t.Fatalf("the portable build input was rejected: %v", err)
		}
		key, err := input.CacheKey()
		if err != nil {
			t.Fatal(err)
		}
		if string(key) != positive.cacheKey || string(key) != vectors.CacheIdentity.Portable.CacheKey {
			t.Fatalf("derived cache key = %s, vector = %s", key, positive.cacheKey)
		}
	})

	for _, testCase := range vectors.RejectionCases {
		if testCase.Boundary != "execution-policy" {
			continue
		}
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Expected.Result != "reject" || testCase.Expected.Reuse ||
				testCase.Expected.SchemaValid || testCase.Expected.CacheLookupPerformed ||
				testCase.Expected.AliasesPortableCacheKey || testCase.Expected.ArtifactExecuted {
				t.Fatalf("vector no longer fails closed: %+v", testCase.Expected)
			}
			if len(testCase.Input.BuildInput) == 0 {
				t.Fatal("vector publishes no build input")
			}
			if _, err := DecodeInput(canonicalInputBytes(t, testCase.Input.BuildInput)); err == nil {
				t.Fatalf("%s was admitted, want the %s rejection", testCase.Name, testCase.Expected.Error)
			}
			if testCase.Input.DerivedCacheKey == vectors.CacheIdentity.Portable.CacheKey {
				t.Fatalf("%s aliases the portable cache key", testCase.Name)
			}
		})
	}
}
