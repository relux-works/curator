package schemaregexp

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Injecting a timed-engine failure is deterministic: fixed-format revisions
// must never invoke that engine, while invalid neighboring values still fail.
func TestRevisionBypassesTimedEngine(t *testing.T) {
	const revision = "c78999d93c004a7c0becef85e6b8009daf76a097"
	for _, pattern := range []string{`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`, `^(?:[0-9a-f]{40}|[0-9a-f]{64})$`} {
		re, err := compile(pattern, func(s string) (jsonschema.Regexp, error) {
			return timedPattern{s, func(string) (bool, error) { return false, errors.New("injected engine timeout") }}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		schema := fixtureSchema(t, re)
		for _, valid := range []string{revision, strings.Repeat("a", 64)} {
			if err := Validate(schema, valid); err != nil {
				t.Fatalf("valid revision reached timed engine: %q", valid)
			}
		}
		for _, invalid := range []string{"", revision[:39], revision + "a", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.ToUpper(revision), revision + "\n", " " + revision, revision[:39] + "g"} {
			if Validate(schema, invalid) == nil {
				t.Fatalf("invalid revision accepted: %q", invalid)
			}
		}
	}
}

// Fixed digest/commit formats retain their exact prefix, length and lowercase
// alphabet; an invalid suffix must never be normalized away.
func TestFixedHexFormats(t *testing.T) {
	for _, tc := range []struct {
		pattern, prefix string
		size            int
	}{
		{`^[0-9a-f]{40}$`, "", 40}, {`^[0-9a-f]{64}$`, "", 64},
		{`^[0-9a-f]{16}$`, "", 16}, {`^sha256:[0-9a-f]{64}$`, "sha256:", 64},
	} {
		re, err := Compile(tc.pattern)
		if err != nil {
			t.Fatal(err)
		}
		valid := tc.prefix + strings.Repeat("a", tc.size)
		if !re.MatchString(valid) {
			t.Fatalf("valid fixed hex refused: %s", tc.pattern)
		}
		for _, bad := range []string{valid + "\n", valid + "a", valid[:len(valid)-1], strings.ToUpper(valid)} {
			if re.MatchString(bad) {
				t.Fatalf("invalid fixed hex accepted: %q", bad)
			}
		}
	}
}

func fixtureSchema(t *testing.T, re jsonschema.Regexp) *jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(func(string) (jsonschema.Regexp, error) { return re, nil })
	if err := c.AddResource("urn:engine-control", map[string]any{"type": "string", "pattern": "^fixture$"}); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile("urn:engine-control")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// A backend timeout is an engine error, never a schema pattern mismatch.
// Subsequent valid/invalid controls and concurrent calls retain independent results.
func TestEngineErrorIsNotPatternMismatch(t *testing.T) {
	timeout := errors.New("injected engine timeout with private-input")
	s := fixtureSchema(t, timedPattern{"^fixture$", func(v string) (bool, error) {
		if v == "timeout" {
			return false, timeout
		}
		return v == "valid", nil
	}})
	check := func(v string) {
		err := Validate(s, v)
		if v == "timeout" {
			var engine *EngineError
			var mismatch *jsonschema.ValidationError
			if !errors.As(err, &engine) || !errors.Is(err, timeout) || errors.As(err, &mismatch) {
				t.Errorf("engine error became pattern mismatch: %T %v", err, err)
			}
			if err != nil && strings.Contains(err.Error(), "private-input") {
				t.Error("engine diagnostic exposed backend input")
			}
		} else if v == "valid" {
			if err != nil {
				t.Errorf("valid control refused: %v", err)
			}
		} else {
			var mismatch *jsonschema.ValidationError
			if !errors.As(err, &mismatch) {
				t.Errorf("invalid control lost schema mismatch: %T %v", err, err)
			}
		}
	}
	for _, v := range []string{"timeout", "valid", "invalid"} {
		check(v)
	}
	var wg sync.WaitGroup
	for _, v := range []string{"timeout", "valid"} {
		wg.Add(1)
		go func() { defer wg.Done(); check(v) }()
	}
	wg.Wait()
}

// Both true+error and false+error remain errors; unrelated engine panics must
// escape unchanged, while deferred serialization unlocks on both fault paths.
func TestEngineFaultBoundaries(t *testing.T) {
	timeout := errors.New("injected timeout")
	for _, matched := range []bool{false, true} {
		s := fixtureSchema(t, timedPattern{"fixture", func(string) (bool, error) { return matched, timeout }})
		if err := Validate(s, "valid"); !errors.Is(err, timeout) {
			t.Fatalf("engine fault ignored: %v", err)
		}
		if !engineClock.TryLock() {
			t.Fatal("engine clock lock leaked after error")
		}
		engineClock.Unlock()
	}
	panicValue := new(int)
	func() {
		defer func() {
			if recover() != panicValue {
				t.Error("unrelated panic swallowed")
			}
		}()
		_ = Validate(fixtureSchema(t, timedPattern{"fixture", func(string) (bool, error) { panic(panicValue) }}), "valid")
	}()
	if !engineClock.TryLock() {
		t.Fatal("engine clock lock leaked after panic")
	}
	engineClock.Unlock()
}

// Every timed backend executes under the shared clock lock, even when schemas
// use different compiled patterns. This deterministically detects its removal.
func TestTimedEngineClockSerialized(t *testing.T) {
	for _, name := range []string{"first", "second"} {
		re := timedPattern{name, func(string) (bool, error) {
			if engineClock.TryLock() {
				engineClock.Unlock()
				t.Error("timed engine entered without shared clock serialization")
			}
			return true, nil
		}}
		if !re.MatchString("valid") {
			t.Fatal("valid backend control refused")
		}
	}
}

// Bad patterns remain compiler errors; ECMA-only lookaheads still use the
// published grammar rather than being silently reinterpreted as RE2 syntax.
func TestECMAPatternControls(t *testing.T) {
	if _, err := Compile("["); err == nil {
		t.Fatal("invalid pattern compiled")
	}
	re, err := Compile(`^(?!\.\.$)[a-z.]+$`)
	if err != nil {
		t.Fatal(err)
	}
	if !re.MatchString("valid") || re.MatchString("..") {
		t.Fatal("ECMA lookahead contract changed")
	}
}
