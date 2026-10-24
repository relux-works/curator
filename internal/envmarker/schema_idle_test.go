package envmarker

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
)

// A full marker with a valid name remains valid when the regexp timeout clock
// restarts through a bounded post-idle window; malformed neighboring names
// still fail schema validation. The dependency deadline regression separately
// detects the stale-clock implementation, without relying on this window.
// This is deliberately serial because the dependency clock is process-global.
func TestMarker7AfterRegexpClockIdle(t *testing.T) {
	raw, err := os.ReadFile("testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	carrier := func(name string) []byte {
		t.Helper()
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		profile := doc["profile"].(map[string]any)
		profile["name"], profile["root"] = name, name
		doc["members"].([]any)[0].(map[string]any)["name"] = name
		return mustJSON(t, doc)
	}
	valid := carrier("capture")
	if _, err := Parse(valid); err != nil {
		t.Fatalf("valid warm marker: %v", err)
	}
	for _, name := range []string{"capture.", "../capture"} {
		if _, err := Parse(carrier(name)); err == nil || !strings.Contains(err.Error(), "pattern") || !strings.Contains(err.Error(), DiagMarkerInvalid) {
			t.Fatalf("invalid neighboring name %q: %v", name, err)
		}
	}
	regexp2.StopTimeoutClock()
	// The production pattern timeout is one second. An idle interval exceeding
	// it exposes stale deadlines when the dependency's periodic clock restarts.
	time.Sleep(1300 * time.Millisecond)
	defer regexp2.StopTimeoutClock()
	until := time.Now().Add(400 * time.Millisecond)
	for time.Now().Before(until) {
		if _, err := Parse(valid); err != nil {
			t.Fatalf("valid marker rejected after regexp clock idle: %v", err)
		}
	}
}
