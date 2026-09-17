package sourcelock

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

// bindingsFixture carries a valid bindings record plus the platform-native
// fixture locations it binds. Locations come from t.TempDir, which is
// absolute and Clean-stable on every platform: Unix-only literals such as
// "/work/skills" are not absolute paths on Windows.
type bindingsFixture struct {
	bindings *Bindings
	project  string
	team     string
}

func mustBindingsFixture(t *testing.T) bindingsFixture {
	t.Helper()
	fixture := bindingsFixture{
		project: filepath.Join(t.TempDir(), "skills"),
		team:    filepath.Join(t.TempDir(), "kit"),
	}
	bindings, err := NewBindings(goldenLockSHA256, map[string]SourceBinding{
		"project": {Location: fixture.project, RootInputs: []string{"SKILL.md", "scripts", "references"}},
		"team":    {Location: fixture.team},
	})
	if err != nil {
		t.Fatalf("NewBindings: %v", err)
	}
	fixture.bindings = bindings
	return fixture
}

func mustBindings(t *testing.T) *Bindings {
	t.Helper()
	return mustBindingsFixture(t).bindings
}

func TestBindingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-bindings.json")
	fixture := mustBindingsFixture(t)
	want := fixture.bindings
	if err := WriteBindings(path, want); err != nil {
		t.Fatalf("WriteBindings: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat bindings: %v", err)
	}
	// Windows only honors the owner-write bit; exact Unix modes are
	// asserted off Windows, following the repository convention.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("bindings mode = %o, want 600", info.Mode().Perm())
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bindings: %v", err)
	}
	if err := protocoljson.RequireCanonical(payload); err != nil {
		t.Fatalf("bindings bytes are not canonical: %v", err)
	}
	// Machine locations live here and only here. The needle is encoded
	// with the same canonical string encoding the payload uses, so the
	// byte check holds on Windows backslash paths too.
	needle := mustCanonical(t, fixture.project)
	if !strings.Contains(string(payload), string(needle)) {
		t.Fatalf("bindings lost the machine location")
	}
	got, err := ReadBindings(path)
	if err != nil {
		t.Fatalf("ReadBindings: %v", err)
	}
	if got.LockSHA256 != want.LockSHA256 || len(got.Sources) != len(want.Sources) {
		t.Fatalf("bindings changed across the round trip")
	}
	entry := got.Sources["project"]
	if entry.Location != fixture.project || strings.Join(entry.RootInputs, ",") != "SKILL.md,scripts,references" {
		t.Fatalf("project binding = %+v", entry)
	}
	if got.Sources["team"].RootInputs != nil {
		t.Fatalf("absent root_inputs became %v", got.Sources["team"].RootInputs)
	}
	if _, err := ReadBindings(filepath.Join(t.TempDir(), "missing.json")); !os.IsNotExist(err) {
		t.Fatalf("missing bindings err = %v, want IsNotExist", err)
	}
}

func bindingsPayload(t *testing.T, mutate func(obj map[string]any)) []byte {
	t.Helper()
	raw, err := json.Marshal(mustBindings(t).object())
	if err != nil {
		t.Fatalf("marshal base bindings: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal base bindings: %v", err)
	}
	if mutate != nil {
		mutate(obj)
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal mutated bindings: %v", err)
	}
	return payload
}

func bindingAt(obj map[string]any, alias string) map[string]any {
	return obj["sources"].(map[string]any)[alias].(map[string]any)
}

func TestBindingsValidation(t *testing.T) {
	rawCases := map[string][]byte{
		"garbage":  []byte("{nope"),
		"array":    []byte("[]"),
		"dup-keys": []byte(`{"schema_version":1,"schema_version":1}`),
		"sources-missing": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { delete(obj, "sources") })
		}(),
		"sources-list": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { obj["sources"] = []any{} })
		}(),
		"unknown-top-level": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { obj["manifest_sha256"] = "x" })
		}(),
		"schema-version-2": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { obj["schema_version"] = float64(2) })
		}(),
		"lock-digest-bare": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { obj["lock_sha256"] = repeat("ab", 32) })
		}(),
		"bad-alias": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) {
				obj["sources"].(map[string]any)["Bad Name"] = obj["sources"].(map[string]any)["team"]
			})
		}(),
		"relative-location": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "team")["location"] = "work/kit" })
		}(),
		"unclean-location": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "team")["location"] = "/work//kit" })
		}(),
		"missing-location": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { delete(bindingAt(obj, "team"), "location") })
		}(),
		"unknown-entry-field": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "team")["snapshot"] = "x" })
		}(),
		"root-inputs-escape": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "project")["root_inputs"] = []any{"SKILL.md", "../x"} })
		}(),
		"root-inputs-absolute": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "project")["root_inputs"] = []any{"/etc/x"} })
		}(),
		"root-inputs-duplicate": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) {
				bindingAt(obj, "project")["root_inputs"] = []any{"SKILL.md", "SKILL.md"}
			})
		}(),
		"root-inputs-overlap": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) {
				bindingAt(obj, "project")["root_inputs"] = []any{"scripts", "scripts/worker"}
			})
		}(),
		"root-inputs-nonstring": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) { bindingAt(obj, "project")["root_inputs"] = []any{1} })
		}(),
		"root-inputs-object": func() []byte {
			return bindingsPayload(t, func(obj map[string]any) {
				bindingAt(obj, "project")["root_inputs"] = map[string]any{}
			})
		}(),
	}
	for name, payload := range rawCases {
		if _, err := ParseBindings(payload); err == nil {
			t.Fatalf("ParseBindings accepted %s", name)
		}
	}
	if _, err := NewBindings("bad", nil); err == nil {
		t.Fatalf("NewBindings accepted a bad digest")
	}
	if err := WriteBindings(filepath.Join(t.TempDir(), "b.json"), nil); err == nil {
		t.Fatalf("WriteBindings accepted nil")
	}
	empty, err := NewBindings(goldenLockSHA256, map[string]SourceBinding{})
	if err != nil {
		t.Fatalf("empty bindings rejected: %v", err)
	}
	if err := empty.Validate(); err != nil {
		t.Fatalf("empty Validate: %v", err)
	}
}

func TestBindingsCheckFresh(t *testing.T) {
	lock := mustGoldenLock(t)
	bindings := mustBindings(t)
	if err := bindings.CheckFresh(lock); err != nil {
		t.Fatalf("fresh bindings rejected: %v", err)
	}
	stale, err := NewBindings("sha256:"+repeat("00", 32), map[string]SourceBinding{
		"project": {Location: filepath.Join(t.TempDir(), "elsewhere")},
	})
	if err != nil {
		t.Fatalf("NewBindings: %v", err)
	}
	if err := stale.CheckFresh(lock); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("foreign generation err = %v, want source_lock_stale", err)
	}
	var nilBindings *Bindings
	if err := nilBindings.CheckFresh(lock); err == nil {
		t.Fatalf("nil bindings accepted")
	}
	if err := bindings.CheckFresh(nil); err == nil {
		t.Fatalf("nil lock accepted")
	}
}

func TestBindingsLocationIsPlatformAbsolute(t *testing.T) {
	// Regression: bindings locations are machine paths, so valid fixtures
	// must be absolute on the test platform. Unix-only literals such as
	// "/work/skills" are not absolute on Windows and failed CI there.
	absolute := filepath.Join(t.TempDir(), "skills")
	if !filepath.IsAbs(absolute) || filepath.Clean(absolute) != absolute {
		t.Fatalf("fixture location %q is not canonical absolute", absolute)
	}
	if _, err := NewBindings(goldenLockSHA256, map[string]SourceBinding{
		"project": {Location: absolute},
	}); err != nil {
		t.Fatalf("platform-absolute location rejected: %v", err)
	}
	if _, err := NewBindings(goldenLockSHA256, map[string]SourceBinding{
		"project": {Location: "relative/path"},
	}); err == nil {
		t.Fatalf("relative location accepted")
	}
}
