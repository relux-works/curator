// Package observation reads the closed curator-run registry/status contract.
// It never invokes manager plugins or turns session status into process proof.
package observation

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Schemas and vectors copied byte-for-byte from curator-agent-launcher
// 870e62350a008f8b71d368a054140221cbc97925. References resolve offline.
//
//go:embed schemas/*.json
var documents embed.FS

const MaxBytes = 64 << 10

type Record struct {
	Manager struct {
		ID           string `json:"id"`
		RegisteredAt string `json:"registered_at"`
	} `json:"manager"`
	Session struct {
		Session    string `json:"session"`
		Manager    string `json:"manager"`
		Readable   bool   `json:"readable"`
		State      string `json:"state"`
		Generation *int64 `json:"generation"`
	} `json:"session"`
}

type offline struct{}

func (offline) Load(string) (any, error) { return nil, fmt.Errorf("unbundled observation schema") }

type expression struct{ re *regexp2.Regexp }

func (e expression) String() string { return e.re.String() }
func (e expression) MatchString(s string) bool {
	ok, err := e.re.MatchString(s)
	return err == nil && ok
}
func pattern(s string) (jsonschema.Regexp, error) {
	r, err := regexp2.Compile(s, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	r.MatchTimeout = time.Second
	return expression{r}, nil
}

var compiled = sync.OnceValues(func() (*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.UseLoader(offline{})
	c.UseRegexpEngine(pattern)
	c.AssertFormat()
	for _, name := range []string{"session-record", "session-observation"} {
		raw, err := documents.ReadFile("schemas/" + name + ".schema.json")
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		if err := c.AddResource(doc["$id"].(string), doc); err != nil {
			return nil, err
		}
	}
	return c.Compile("urn:relux:curator-run:session-observation")
})

// Parse validates all required/unknown fields before extracting the small
// projection used for identity checks. The projection grants no lease authority.
func Parse(raw []byte) (Record, error) {
	var r Record
	invalid := fmt.Errorf("invalid curator-run observation")
	if len(raw) > MaxBytes || protocoljson.Validate(raw) != nil {
		return r, invalid
	}
	schema, err := compiled()
	if err != nil {
		return r, invalid
	}
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if d.Decode(&value) != nil || schema.Validate(value) != nil {
		return r, invalid
	}
	if json.Unmarshal(raw, &r) != nil {
		return Record{}, invalid
	}
	return r, nil
}
