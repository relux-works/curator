package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type claudeEnvSchemaLoader struct{}

func (claudeEnvSchemaLoader) Load(string) (any, error) {
	return nil, fmt.Errorf("unadmitted schema resource")
}

// The unmodified member schemas are an independent structural oracle; they
// do not stand in for a whole-carrier or real launch/resume acceptance check.
func TestLauncherClaudeEnvPublishedSchemas(t *testing.T) {
	root := filepath.Join("testdata", "machine-claude-env-v1")
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(claudeEnvSchemaLoader{})
	base := "https://relux-works.github.io/curator-spec/schemas/machine-claude-env-v1/"
	schemas := map[string]*jsonschema.Schema{}
	names := map[string]string{"declaration": "claude-env.schema.json", "projection": "projection.schema.json"}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		var doc any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal("invalid pinned schema")
		}
		if err := compiler.AddResource(base+name, doc); err != nil {
			t.Fatal("cannot register pinned schema")
		}
	}
	for entry, name := range names {
		s, err := compiler.Compile(base + name)
		if err != nil {
			t.Fatal("cannot compile pinned schema")
		}
		schemas[entry] = s
	}
	var vectors []claudeEnvVector
	raw, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatal(err)
	}
	var plants []claudeEnvPlant
	raw, err = os.ReadFile(filepath.Join(root, "plants.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &plants); err != nil {
		t.Fatal(err)
	}
	for _, p := range plants {
		vectors = append(vectors, claudeEnvVector{ID: p.ID + "-control", Entry: p.Entry, Input: p.Valid, SchemaValid: true}, claudeEnvVector{ID: p.ID + "-plant", Entry: p.Entry, Input: p.Invalid, SchemaValid: false})
	}
	for _, v := range vectors {
		t.Run(v.ID, func(t *testing.T) {
			var instance any
			accepted := json.Unmarshal([]byte(v.Input), &instance) == nil
			if accepted {
				accepted = schemas[v.Entry].Validate(instance) == nil
			}
			if accepted != v.SchemaValid {
				t.Fatal("pinned JSON schema verdict mismatch")
			}
		})
	}
}
