package envsnapshot

import (
	"errors"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/schemaregexp"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type snapshotEngineControl struct{ cause error }

func (snapshotEngineControl) String() string { return "^fixture$" }
func (r snapshotEngineControl) MatchString(s string) bool {
	if s == "timeout" {
		panic(&schemaregexp.EngineError{Cause: r.cause})
	}
	return s == "valid"
}

// The production snapshot boundary retains its bounded refusal plus typed
// engine cause, while ordinary invalid data remains only a schema mismatch.
func TestSnapshotSchemaEngineError(t *testing.T) {
	cause := errors.New("injected engine timeout with private input")
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(func(string) (jsonschema.Regexp, error) { return snapshotEngineControl{cause}, nil })
	if err := c.AddResource("urn:snapshot-engine", map[string]any{"type": "string", "pattern": "fixture"}); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile("urn:snapshot-engine")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSnapshotSchema(s, "valid", "manifest-schema"); err != nil {
		t.Fatal(err)
	}
	if err := validateSnapshotSchema(s, "invalid", "manifest-schema"); err == nil || err.Error() != Mismatch+": manifest-schema" {
		t.Fatalf("invalid snapshot control: %v", err)
	}
	err = validateSnapshotSchema(s, "timeout", "manifest-schema")
	var engine *schemaregexp.EngineError
	var refusal *Error
	if !errors.Is(err, cause) || !errors.As(err, &engine) || !errors.As(err, &refusal) || refusal.Reason != "schema-engine" {
		t.Fatalf("snapshot engine error hidden: %v", err)
	}
	if strings.Contains(err.Error(), "private input") {
		t.Fatal("engine cause leaked input into bounded refusal")
	}
}
