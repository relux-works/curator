package envmarker

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Exact published carrier and its transitive schema closure. Canonical IDs
// resolve offline; an unknown reference fails closed, without a private alias.
//
//go:embed schemas/*/*.json
var markerSchemas embed.FS

var marker7Schema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.UseLoader(markerSchemaLoader{})
	c.UseRegexpEngine(markerPattern)
	paths, err := fs.Glob(markerSchemas, "schemas/*/*.json")
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		raw, err := markerSchemas.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err = json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		if err = c.AddResource(doc["$id"].(string), doc); err != nil {
			return nil, err
		}
	}
	return c.Compile("https://relux-works.github.io/curator-spec/schemas/codex-seed-c-v1/agent-environment-marker-v7.schema.json")
})

type markerSchemaLoader struct{}

func (markerSchemaLoader) Load(string) (any, error) {
	return nil, fmt.Errorf("unbundled marker schema")
}

type markerRegexp struct{ re *regexp2.Regexp }

func (r markerRegexp) String() string { return r.re.String() }
func (r markerRegexp) MatchString(s string) bool {
	ok, err := r.re.MatchString(s)
	return err == nil && ok
}
func markerPattern(pattern string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = time.Second
	return markerRegexp{re}, nil
}
func validateMarker7(raw []byte) error {
	schema, err := marker7Schema()
	if err != nil {
		return fmt.Errorf("marker-7 schema unavailable: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var doc any
	if err = decoder.Decode(&doc); err != nil {
		return err
	}
	if err = schema.Validate(doc); err != nil {
		return fmt.Errorf("marker-7 carrier invalid: %w", err)
	}
	return nil
}
