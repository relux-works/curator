package envmarker

import (
	"encoding/json"
	"errors"
	"github.com/dlclark/regexp2"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/schemaregexp"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type markerEngineControl struct{ cause error }

func (markerEngineControl) String() string { return "^fixture$" }
func (r markerEngineControl) MatchString(s string) bool {
	if s == "timeout" {
		panic(&schemaregexp.EngineError{Cause: r.cause})
	}
	return s == "valid"
}

// The production marker boundary preserves an injected engine fault as such;
// valid data passes and an ordinary invalid value remains a carrier mismatch.
func TestMarkerSchemaEngineError(t *testing.T) {
	cause := errors.New("injected engine timeout")
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(func(string) (jsonschema.Regexp, error) { return markerEngineControl{cause}, nil })
	if err := c.AddResource("urn:marker-engine", map[string]any{"type": "string", "pattern": "fixture"}); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile("urn:marker-engine")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateMarkerSchema(s, "valid"); err != nil {
		t.Fatal(err)
	}
	if err := validateMarkerSchema(s, "invalid"); err == nil || !strings.Contains(err.Error(), "carrier invalid") {
		t.Fatalf("invalid marker control: %v", err)
	}
	err = validateMarkerSchema(s, "timeout")
	var engine *schemaregexp.EngineError
	if !errors.Is(err, cause) || !errors.As(err, &engine) || strings.Contains(err.Error(), "carrier invalid") {
		t.Fatalf("marker engine error hidden: %v", err)
	}
}

func markerRevisionFixture(t *testing.T, revision string) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	p := doc["profile"].(map[string]any)
	p["kind"] = "git"
	p["source"] = "github.com/example/context-example"
	p["requirement"] = map[string]any{"revision": revision}
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// Actual bundled marker schemas accept 40/64-hex git revisions and reject
// short, uppercase, nonhex and newline variants, without changing the fixture.
func TestMarkerSchemaRevisionFormats(t *testing.T) {
	for _, revision := range []string{strings.Repeat("a", 40), strings.Repeat("b", 64), strings.Repeat("a", 39), strings.Repeat("A", 40), strings.Repeat("g", 40), strings.Repeat("a", 40) + "\n"} {
		err := validateMarker7(markerRevisionFixture(t, revision))
		valid := revision == strings.Repeat("a", 40) || revision == strings.Repeat("b", 64)
		if valid != (err == nil) {
			t.Fatalf("revision valid=%v: %v", valid, err)
		}
	}
}

// Two production schema callers restart the real dependency clock after idle;
// fixed revisions and the other ECMA fields remain valid. Injection in the
// shared adapter tests is the deterministic false-timeout regression; these
// three real-clock rounds are a bounded integration check, not exhaustive proof.
func TestMarker7ConcurrentIdleRevision(t *testing.T) {
	valid := markerRevisionFixture(t, "c78999d93c004a7c0becef85e6b8009daf76a097")
	invalid := markerRevisionFixture(t, strings.Repeat("g", 40))
	if err := validateMarker7(valid); err != nil {
		t.Fatal(err)
	}
	defer regexp2.StopTimeoutClock()
	for i := 0; i < 3; i++ {
		regexp2.StopTimeoutClock()
		time.Sleep(1300 * time.Millisecond)
		var wg sync.WaitGroup
		gate := make(chan struct{})
		results := make(chan error, 2)
		for j := 0; j < 2; j++ {
			wg.Add(1)
			go func() { defer wg.Done(); <-gate; results <- validateMarker7(valid) }()
		}
		close(gate)
		wg.Wait()
		close(results)
		for err := range results {
			if err != nil {
				t.Fatalf("valid production revision refused after concurrent idle: %v", err)
			}
		}
		if err := validateMarker7(invalid); err == nil {
			t.Fatal("invalid revision accepted after idle")
		}
	}
}
