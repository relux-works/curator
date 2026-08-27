package scopes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
)

// ambiguousRegistry renders a registry whose consumers member is stated twice:
// once with a live checkout, once empty. Struct decoding keeps the last one and
// calls the result trusted, which is the whole hazard.
func ambiguousRegistry(t *testing.T, project string) []byte {
	t.Helper()
	encoded, err := json.Marshal(project)
	if err != nil {
		t.Fatal(err)
	}
	return []byte(`{"schema_version": 1, "consumers": [` + string(encoded) + `], "consumers": []}`)
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

// failSafeCase is one way the live reference set becomes unprovable, together
// with what must still be true after maintenance ran twice.
type failSafeCase struct {
	// arrange makes the reference set unprovable. The project is already
	// registered as a consumer when it runs.
	arrange func(t *testing.T, home, project string)
	// warning is the substring the pass has to report every time.
	warning string
	// verify checks that the state which protects the reference set survived.
	verify func(t *testing.T, home, project string)
}

// runFailSafeAcrossTwoPasses proves an uncertainty is not consumed by the pass
// that reports it.
//
// One conservative pass is worthless if it destroys the very state that made it
// conservative: an emptied consumer registry, or a project quietly dropped from
// it, would let the next pass see a complete-looking reference set and sweep
// build artifacts that are still referenced. Every case therefore runs twice
// and asserts that the second pass is exactly as unwilling as the first.
func runFailSafeAcrossTwoPasses(t *testing.T, test failSafeCase) {
	t.Helper()
	home := t.TempDir()
	project := t.TempDir()
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	test.arrange(t, home, project)

	cache := &recordingCache{}
	for pass := 1; pass <= 2; pass++ {
		result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
		if err != nil {
			t.Fatalf("pass %d: %v", pass, err)
		}
		if cache.calls != 0 {
			t.Fatalf("pass %d swept the build cache with an unprovable reference set", pass)
		}
		if !warned(result, test.warning) {
			t.Fatalf("pass %d lost the warning %q: %v", pass, test.warning, result.Warnings)
		}
		if !warned(result, "build cache sweep skipped") {
			t.Fatalf("pass %d did not report the skipped sweep: %v", pass, result.Warnings)
		}
		test.verify(t, home, project)
	}
}

// registryKeeps asserts the checkout is still registered, which is what makes
// the next pass visit its markers again.
func registryKeeps(t *testing.T, home, project string) {
	t.Helper()
	consumers := LoadConsumers(home)
	if !slices.Contains(consumers, project) {
		t.Fatalf("an uncertain consumer was dropped from the registry: %v", consumers)
	}
}

// TestCollectStaysFailSafeAcrossConsecutivePasses covers the uncertainties a
// second pass could otherwise inherit as a clean reference set.
func TestCollectStaysFailSafeAcrossConsecutivePasses(t *testing.T) {
	tests := map[string]failSafeCase{
		"corrupt consumer registry": {
			arrange: func(t *testing.T, home, _ string) {
				if err := os.WriteFile(filepath.Join(home, ConsumersName), []byte("{"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "cannot be trusted",
			verify: func(t *testing.T, home, _ string) {
				payload, err := os.ReadFile(filepath.Join(home, ConsumersName)) // #nosec G304 -- test fixture
				if err != nil {
					t.Fatal(err)
				}
				if string(payload) != "{" {
					t.Fatalf("an untrusted registry was rewritten as %q", payload)
				}
			},
		},
		"registry of an unknown schema": {
			arrange: func(t *testing.T, home, project string) {
				payload := []byte(`{"schema_version": 99, "consumers": ["` + filepath.ToSlash(project) + `"]}`)
				if err := os.WriteFile(filepath.Join(home, ConsumersName), payload, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "unsupported registry schema_version",
			verify: func(t *testing.T, home, _ string) {
				payload, err := os.ReadFile(filepath.Join(home, ConsumersName)) // #nosec G304 -- test fixture
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(payload), `"schema_version": 99`) {
					t.Fatalf("an untrusted registry was rewritten as %q", payload)
				}
			},
		},
		"registry repeating a known member": {
			arrange: func(t *testing.T, home, project string) {
				if err := os.WriteFile(filepath.Join(home, ConsumersName), ambiguousRegistry(t, project), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: `repeats the "consumers" member`,
			verify: func(t *testing.T, home, project string) {
				if got := readTestFile(t, filepath.Join(home, ConsumersName)); got != string(ambiguousRegistry(t, project)) {
					t.Fatalf("an ambiguous registry was rewritten as %q", got)
				}
			},
		},
		"project with only an invalid marker": {
			arrange: func(t *testing.T, _, project string) {
				dir := filepath.Join(project, ".agents", "skills", "skill-broken")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, marker.Name), []byte("{not a marker}"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "is unreadable or invalid",
			verify:  registryKeeps,
		},
		"unreadable installed skill directory": {
			arrange: func(t *testing.T, _, project string) {
				dir := filepath.Join(project, ".agents", "skills", "skill-locked")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(dir, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
				if _, err := os.Lstat(filepath.Join(dir, marker.Name)); !os.IsPermission(err) {
					t.Skipf("this environment can inspect a mode-000 directory: %v", err)
				}
			},
			warning: "cannot be inspected",
			verify:  registryKeeps,
		},
		"installed skill replaced by a file": {
			arrange: func(t *testing.T, _, project string) {
				skills := filepath.Join(project, ".agents", "skills")
				if err := os.MkdirAll(skills, 0o755); err != nil {
					t.Fatal(err)
				}
				installMarker(t, skills, "skill-a", strings.Repeat("1", 40))
				if err := os.WriteFile(filepath.Join(skills, "notes.txt"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
				// A file cannot hide a marker, so it must not block the sweep on
				// its own; the invalid marker beside it is what does.
				broken := filepath.Join(skills, "skill-broken")
				if err := os.MkdirAll(broken, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(broken, marker.Name), []byte("{"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "ignored non-directory member",
			verify:  registryKeeps,
		},
		"skill root replaced by a file": {
			arrange: func(t *testing.T, _, project string) {
				agents := filepath.Join(project, ".agents")
				if err := os.MkdirAll(agents, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(agents, "skills"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "is not a directory",
			verify:  registryKeeps,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) { runFailSafeAcrossTwoPasses(t, test) })
	}
}

// TestNonDirectoryScopeMembersDoNotBlockMaintenance proves the one shape that
// cannot hide a marker is reported without making the sweep impossible — the
// dot-files a file browser leaves behind must not disable collection forever.
func TestNonDirectoryScopeMembersDoNotBlockMaintenance(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	skills := filepath.Join(project, ".agents", "skills")
	installBuildMarker(t, skills, "skill-p", strings.Repeat("a", 40), "tool", buildKey("1"))
	if err := os.WriteFile(filepath.Join(skills, ".DS_Store"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}

	cache := &recordingCache{}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
	if err != nil {
		t.Fatal(err)
	}
	if cache.calls != 1 {
		t.Fatalf("an ignorable member blocked the build sweep: %v", result.Warnings)
	}
	if !warned(result, "ignored non-directory member") {
		t.Fatalf("the ignored member was not reported: %v", result.Warnings)
	}
	if len(cache.referenced) != 1 || cache.referenced[0] != string(buildKey("1")) {
		t.Fatalf("referenced = %v", cache.referenced)
	}
}

// TestRecordingAConsumerNeverOverwritesAnUntrustedRegistry proves an install
// cannot silently unregister every other checkout, which is the other way a
// later pass would end up sweeping referenced artifacts.
func TestRecordingAConsumerNeverOverwritesAnUntrustedRegistry(t *testing.T) {
	home := t.TempDir()
	registry := filepath.Join(home, ConsumersName)
	if err := os.WriteFile(registry, []byte(`{"consumers": ["/a"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(registry) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, t.TempDir()); err == nil {
		t.Fatal("a checkout was recorded over a registry that could not be read")
	}
	if _, err := StageConsumer(t.TempDir(), home, t.TempDir()); err == nil {
		t.Fatal("a consumer was staged over a registry that could not be read")
	}
	after, err := os.ReadFile(registry) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("the registry changed from %q to %q", before, after)
	}
}

// TestParseConsumersAcceptsOnlyTheCanonicalRegistry pins the shape validation
// that keeps "unknown consumers" from reading as "no consumers".
func TestParseConsumersAcceptsOnlyTheCanonicalRegistry(t *testing.T) {
	canonical, err := ConsumersPayload(map[string]bool{absoluteTestPath("a"): true})
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := parseConsumers(canonical)
	if err != nil || len(consumers) != 1 || consumers[0] != absoluteTestPath("a") {
		t.Fatalf("parseConsumers(canonical) = %v, %v", consumers, err)
	}
	empty, err := ConsumersPayload(nil)
	if err != nil {
		t.Fatal(err)
	}
	if consumers, err := parseConsumers(empty); err != nil || len(consumers) != 0 {
		t.Fatalf("parseConsumers(empty) = %v, %v", consumers, err)
	}
	for name, payload := range map[string]string{
		"not json":               "{",
		"not an object":          `["/a"]`,
		"missing version":        `{"consumers": []}`,
		"unknown version":        `{"schema_version": 2, "consumers": []}`,
		"missing consumers":      `{"schema_version": 1}`,
		"null consumers":         `{"schema_version": 1, "consumers": null}`,
		"unknown field":          `{"schema_version": 1, "consumers": [], "extra": true}`,
		"wrong element type":     `{"schema_version": 1, "consumers": [7]}`,
		"empty checkout path":    `{"schema_version": 1, "consumers": [""]}`,
		"relative checkout":      `{"schema_version": 1, "consumers": ["relative/path"]}`,
		"trailing content":       `{"schema_version": 1, "consumers": []} {}`,
		"repeated consumers":     `{"schema_version": 1, "consumers": ["/a"], "consumers": []}`,
		"repeated empty first":   `{"schema_version": 1, "consumers": [], "consumers": ["/a"]}`,
		"repeated version":       `{"schema_version": 1, "schema_version": 1, "consumers": []}`,
		"repeated bad version":   `{"schema_version": 99, "schema_version": 1, "consumers": []}`,
		"fractional version":     `{"schema_version": 1.0, "consumers": []}`,
		"exponent version":       `{"schema_version": 1e0, "consumers": []}`,
		"string version":         `{"schema_version": "1", "consumers": []}`,
		"null version":           `{"schema_version": null, "consumers": []}`,
		"consumers is an object": `{"schema_version": 1, "consumers": {}}`,
		"nested consumer list":   `{"schema_version": 1, "consumers": [["/a"]]}`,
	} {
		if consumers, err := parseConsumers([]byte(payload)); err == nil {
			t.Fatalf("%s: parseConsumers accepted %q as %v", name, payload, consumers)
		}
	}
}

// TestARepeatedRegistryMemberNeverEmptiesTheRegistry is the two-pass regression
// for the ambiguity a struct decoder would resolve silently.
//
// A registry that states its consumer list twice does not say what it means.
// Reading it as "the last one wins" would make an empty list look trusted:
// maintenance would rewrite the registry empty, the next pass would no longer
// visit the live checkout, and the build artifact that checkout still runs
// would be collected. Both passes must therefore refuse, keep the bytes, and —
// once an operator repairs the registry — the live reference must still be
// there to protect its key.
func TestARepeatedRegistryMemberNeverEmptiesTheRegistry(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	key := buildKey("7")
	installBuildMarker(t, filepath.Join(project, ".agents", "skills"), "skill-live", strings.Repeat("a", 40), "live-tool", key)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	registry := filepath.Join(home, ConsumersName)
	ambiguous := ambiguousRegistry(t, project)
	if err := os.WriteFile(registry, ambiguous, 0o644); err != nil {
		t.Fatal(err)
	}

	cache := &recordingCache{}
	for pass := 1; pass <= 2; pass++ {
		result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
		if err != nil {
			t.Fatalf("pass %d: %v", pass, err)
		}
		if cache.calls != 0 {
			t.Fatalf("pass %d swept the build cache behind an ambiguous registry", pass)
		}
		if !warned(result, `repeats the "consumers" member`) {
			t.Fatalf("pass %d did not report the ambiguity: %v", pass, result.Warnings)
		}
		if got := readTestFile(t, registry); got != string(ambiguous) {
			t.Fatalf("pass %d rewrote the ambiguous registry as %q", pass, got)
		}
	}

	// Nothing was forgotten while the registry could not be trusted: repairing
	// it is enough for the live checkout to protect its build key again.
	if err := ReplaceConsumers(home, []string{project}); err != nil {
		t.Fatal(err)
	}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
	if err != nil {
		t.Fatal(err)
	}
	if cache.calls != 1 || len(cache.referenced) != 1 || cache.referenced[0] != string(key) {
		t.Fatalf("the live reference did not survive: calls=%d referenced=%v warnings=%v",
			cache.calls, cache.referenced, result.Warnings)
	}
}

// TestWritersRefuseARepeatedRegistryMember proves the ambiguity is refused
// before a writer can normalize it away. Merging into "the last one wins" would
// unregister every other checkout on the machine and leave a registry that now
// looks perfectly canonical to the next pass.
func TestWritersRefuseARepeatedRegistryMember(t *testing.T) {
	home := t.TempDir()
	registry := filepath.Join(home, ConsumersName)
	ambiguous := ambiguousRegistry(t, absoluteTestPath("live"))
	if err := os.WriteFile(registry, ambiguous, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, t.TempDir()); err == nil {
		t.Fatal("a checkout was recorded over an ambiguous registry")
	}
	if _, err := StageConsumer(t.TempDir(), home, t.TempDir()); err == nil {
		t.Fatal("a consumer was staged over an ambiguous registry")
	}
	if got := readTestFile(t, registry); got != string(ambiguous) {
		t.Fatalf("the ambiguous registry was rewritten as %q", got)
	}
	if consumers := LoadConsumers(home); len(consumers) != 0 {
		t.Fatalf("an ambiguous registry read as a consumer list: %v", consumers)
	}
}

// TestStructDecodingWouldTrustARepeatedRegistryMember is the negative control
// for the parser change. It runs the decoding strategy the previous cycle used
// and shows it accepting the ambiguous document as an empty, trusted registry —
// the exact reading the two-pass regressions above are written against.
func TestStructDecodingWouldTrustARepeatedRegistryMember(t *testing.T) {
	ambiguous := ambiguousRegistry(t, absoluteTestPath("live"))
	decoder := json.NewDecoder(strings.NewReader(string(ambiguous)))
	decoder.DisallowUnknownFields()
	var document struct {
		SchemaVersion *int      `json:"schema_version"`
		Consumers     *[]string `json:"consumers"`
	}
	if err := decoder.Decode(&document); err != nil {
		t.Fatalf("the control decoder rejected the document for an unrelated reason: %v", err)
	}
	if document.SchemaVersion == nil || *document.SchemaVersion != 1 {
		t.Fatalf("the control did not reproduce a trusted schema version: %v", document.SchemaVersion)
	}
	if document.Consumers == nil || len(*document.Consumers) != 0 {
		t.Fatalf("the control did not reproduce the emptied consumer list: %v", document.Consumers)
	}
	if consumers, err := parseConsumers(ambiguous); err == nil {
		t.Fatalf("parseConsumers accepted the ambiguous registry as %v", consumers)
	}
}

func absoluteTestPath(name string) string {
	resolved, err := filepath.Abs(filepath.Join(os.TempDir(), name))
	if err != nil {
		return filepath.Join(os.TempDir(), name)
	}
	return resolved
}
