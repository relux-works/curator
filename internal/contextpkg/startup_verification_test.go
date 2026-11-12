package contextpkg

import (
	"strconv"
	"testing"
)

// Manifest3 retains manifest2's empty settings field and skill-directory
// grammar while admitting only a closed integer minimum. Old readers stay
// strict about the added field; profile weakening is decided after resolution.
func TestStartupVerificationManifest(t *testing.T) {
	for _, level := range []int{1, 2, 3, 4} {
		payload := []byte(`{"schema_version":3,"name":"fixture","version":"1.0.0","codex_settings":{},"startup_verification":{"level":` + strconv.Itoa(level) + `}}`)
		m, err := ParseManifest(payload)
		if err != nil || m.StartupVerificationLevel == nil || *m.StartupVerificationLevel != level {
			t.Fatalf("level %d: %+v,%v", level, m, err)
		}
	}
	for _, value := range []string{`null`, `{}`, `{"level":0}`, `{"level":5}`, `{"level":3.0}`, `{"level":3e0}`, `{"level":3,"extra":true}`, `{"level":3,"level":4}`} {
		if _, err := ParseManifest([]byte(`{"schema_version":3,"name":"fixture","version":"1.0.0","startup_verification":` + value + `}`)); err == nil {
			t.Fatalf("accepted %s", value)
		}
	}
	for _, version := range []string{"1", "2"} {
		if _, err := ParseManifest([]byte(`{"schema_version":` + version + `,"name":"fixture","version":"1.0.0","startup_verification":{"level":3}}`)); err == nil {
			t.Fatal("old manifest accepted new field")
		}
	}
}
