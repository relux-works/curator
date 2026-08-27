package buildcache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCandidateCacheOutcomeVocabulary(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	buildDrivers := readJSONObject(t, filepath.Join(root, "vectors", "build-drivers.json"))
	rejections, ok := buildDrivers["rejection_cases"].([]any)
	if !ok {
		t.Fatal("candidate build-drivers vectors omit rejection_cases")
	}
	forged := namedObject(t, rejections, "self-consistent-forged-receipt-outside-protected-state")
	expected, ok := forged["expected"].(map[string]any)
	if !ok {
		t.Fatal("forged-cache vector omits expected outcome")
	}
	if got := (Result{Status: UntrustedProvenance}).DryRunOutcome(); expected["dry_run"] != got {
		t.Fatalf("untrusted dry-run outcome = %q, candidate wants %v", got, expected["dry_run"])
	}
	if expected["error"] != "untrusted_provenance" || expected["reuse"] != false {
		t.Fatalf("forged-cache candidate no longer fails closed: %+v", expected)
	}

	lifecycle := readJSONObject(t, filepath.Join(root, "vectors", "manager-lifecycle.json"))
	cases, ok := lifecycle["cache_publication_cases"].([]any)
	if !ok {
		t.Fatal("candidate manager lifecycle omits cache_publication_cases")
	}
	want := map[string]string{
		"publish-complete-immutable-entry-under-home-lock": "published",
		"concurrent-identical-winner":                      "reuse-winner",
		"concurrent-determinism-mismatch":                  "determinism-or-corruption-error",
		"corrupt-live-entry":                               "replace-from-verified-staging",
		"untrusted-cache-boundary":                         "rebuild-into-new-protected-state",
	}
	for name, result := range want {
		if got := namedObject(t, cases, name)["result"]; got != result {
			t.Fatalf("candidate publication %s = %v, want %s", name, got, result)
		}
	}
}

func readJSONObject(t *testing.T, path string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit candidate conformance input
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func namedObject(t *testing.T, values []any, name string) map[string]any {
	t.Helper()
	for _, value := range values {
		object, ok := value.(map[string]any)
		if ok && object["name"] == name {
			return object
		}
	}
	t.Fatalf("candidate vector %q is absent", name)
	return nil
}
