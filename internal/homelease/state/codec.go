package state

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Exact public schemas from curator-spec 3f8ffb7f15213872c66ceec28501b9e51d56ac5c.
// All references resolve offline; no schema can introduce a network dependency.
//
//go:embed schemas/*.json
var schemas embed.FS

const schemaRoot = "https://relux-works.github.io/curator-spec/schemas/home-leases-v1/"
const MaxRecord = 16 << 20

var compiled = sync.OnceValues(func() (map[string]*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.UseLoader(offline{})
	c.UseRegexpEngine(pattern)
	c.AssertFormat()
	paths, err := fs.Glob(schemas, "schemas/*.json")
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		raw, err := schemas.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var document map[string]any
		if err = json.Unmarshal(raw, &document); err != nil {
			return nil, err
		}
		if err = c.AddResource(document["$id"].(string), document); err != nil {
			return nil, err
		}
	}
	result := map[string]*jsonschema.Schema{}
	for _, name := range []string{"holder", "record", "event", "result"} {
		s, err := c.Compile(schemaRoot + name + ".schema.json")
		if err != nil {
			return nil, err
		}
		result[name] = s
	}
	return result, nil
})

type offline struct{}

func (offline) Load(string) (any, error) { return nil, fmt.Errorf("unbundled home lease schema") }

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

// Validate rejects duplicate/case-aliased/missing fields before Go's permissive
// struct decoder can normalize them. Schemas also enforce safe integer bounds.
func Validate(name string, raw []byte) error {
	if len(raw) > MaxRecord {
		return fmt.Errorf("home lease object exceeds bound")
	}
	if err := protocoljson.Validate(raw); err != nil {
		return fmt.Errorf("invalid home lease JSON")
	}
	all, err := compiled()
	if err != nil {
		return err
	}
	s, ok := all[name]
	if !ok {
		return fmt.Errorf("unknown home lease schema")
	}
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err = d.Decode(&value); err != nil {
		return fmt.Errorf("invalid home lease JSON")
	}
	if err = s.Validate(value); err != nil {
		return fmt.Errorf("invalid home lease %s", name)
	}
	return nil
}

func Canonical(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var domain any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err = d.Decode(&domain); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(domain)
}

func ParseHolder(raw []byte) (Holder, error) {
	var h Holder
	if len(raw) > 8192 || Validate("holder", raw) != nil {
		return h, Refuse("usage", "invalid_holder")
	}
	if err := json.Unmarshal(raw, &h); err != nil {
		return h, Refuse("usage", "invalid_holder")
	}
	return h, nil
}
