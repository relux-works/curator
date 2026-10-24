package projectcontext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

type compositionPattern struct{ re *regexp2.Regexp }

func (p compositionPattern) MatchString(s string) bool {
	ok, err := p.re.MatchString(s)
	return err == nil && ok
}
func (p compositionPattern) String() string { return p.re.String() }

type compositionSchemaLoader struct{}

func (compositionSchemaLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("schema outside pinned corpus: %s", url)
}
func compositionSchemas(t *testing.T) map[string]*jsonschema.Schema {
	t.Helper()
	c := jsonschema.NewCompiler()
	c.UseLoader(compositionSchemaLoader{})
	c.UseRegexpEngine(func(pattern string) (jsonschema.Regexp, error) {
		re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
		if err != nil {
			return nil, err
		}
		re.MatchTimeout = time.Second
		return compositionPattern{re}, nil
	})
	root := filepath.Join("testdata", "composition")
	b, err := os.ReadFile(filepath.Join(root, "schemas-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pins map[string]string
	if err = json.Unmarshal(b, &pins); err != nil {
		t.Fatal(err)
	}
	for path, hash := range pins {
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		if digest(b) != hash {
			t.Fatal("pinned schema bytes differ")
		}
		var doc map[string]any
		if err = json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if err = c.AddResource(doc["$id"].(string), doc); err != nil {
			t.Fatal(err)
		}
	}
	out := map[string]*jsonschema.Schema{}
	for _, name := range []string{"composition", "generation", "adapter", "control-receipt", "registration"} {
		schema, err := c.Compile("https://relux-works.github.io/curator-spec/schemas/project-context-leaf1-v1/" + name + ".schema.json")
		if err != nil {
			t.Fatal(err)
		}
		out[name] = schema
	}
	return out
}
func assertCompositionSchemas(t *testing.T, req ComposeRequest, p CompositionPreview) {
	t.Helper()
	schemas := compositionSchemas(t)
	for name, record := range map[string]any{"composition": p.Record, "generation": p.Generation, "adapter": req.Adapter, "control-receipt": p.Receipt} {
		b, err := canonical(record)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err = json.Unmarshal(b, &value); err != nil {
			t.Fatal(err)
		}
		if err = schemas[name].Validate(value); err != nil {
			t.Fatalf("generated %s violates pinned schema: %v", name, err)
		}
	}
}

// Claim: the pinned schema checker accepts its published valid control and
// rejects a planted memory-on value at the intended memory/const assertion.
// A schema compilation error or unrelated rejection does not satisfy this test.
func TestCompositionSchemaControl(t *testing.T) {
	schema := compositionSchemas(t)["composition"]
	b, err := os.ReadFile(filepath.Join("testdata", "composition", "bytes", "composition-valid.json"))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err = json.Unmarshal(b, &value); err != nil {
		t.Fatal(err)
	}
	if err = schema.Validate(value); err != nil {
		t.Fatal(err)
	}
	value["memory"] = "on"
	err = schema.Validate(value)
	failure, ok := err.(*jsonschema.ValidationError)
	if !ok || !compositionMemoryControl(failure) {
		t.Fatalf("known plant missed intended memory/const assertion: %v", err)
	}
	value["memory"] = "off"
	if err = schema.Validate(value); err != nil {
		t.Fatal(err)
	}
}
func compositionMemoryControl(err *jsonschema.ValidationError) bool {
	if _, ok := err.ErrorKind.(*kind.Const); ok && len(err.InstanceLocation) == 1 && err.InstanceLocation[0] == "memory" {
		return true
	}
	for _, cause := range err.Causes {
		if compositionMemoryControl(cause) {
			return true
		}
	}
	return false
}
