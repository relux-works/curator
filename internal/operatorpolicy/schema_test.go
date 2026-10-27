package operatorpolicy

import (
	"encoding/json"
	"os"
	"testing"
)

// Claim: the exact pinned CIP-0020 event corpus's 49 structural/semantic
// witnesses are detected by the production event reader, beside its published
// positive. This validates private event bytes, not live approval or providers.
func TestPolicyPublishedEventSchema(t *testing.T) {
	b, e := os.ReadFile("testdata/positive.json")
	if e != nil {
		t.Fatal(e)
	}
	var base struct {
		Record json.RawMessage `json:"record"`
		Facts  struct {
			Policy Policy `json:"policy"`
		} `json:"facts"`
	}
	if e = json.Unmarshal(b, &base); e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Name, Op string
		Path     []string
		Value    any
		Expected *string
	}
	b, e = os.ReadFile("testdata/schema-cases.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	check := func(b []byte) error {
		var ev Event
		if e := decode(b, &ev); e != nil {
			return e
		}
		return validEvent(ev)
	}
	if e = check(base.Record); e != nil {
		t.Fatal("published positive", e)
	}
	var published Event
	if e = json.Unmarshal(base.Record, &published); e != nil || hash(base.Facts.Policy) != published.Policy.SHA256 {
		t.Fatal("complete policy CCJ-1 digest differs from the independent published witness", e)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var v map[string]any
			if e := json.Unmarshal(base.Record, &v); e != nil {
				t.Fatal(e)
			}
			target := v
			for _, p := range c.Path[:len(c.Path)-1] {
				target = target[p].(map[string]any)
			}
			key := c.Path[len(c.Path)-1]
			if c.Op == "remove" {
				delete(target, key)
			} else {
				target[key] = c.Value
			}
			b, e := json.Marshal(v)
			if e != nil {
				t.Fatal(e)
			}
			e = check(b)
			if c.Expected == nil {
				if e != nil {
					t.Fatal(e)
				}
			} else {
				require(t, e, *c.Expected)
			}
		})
	}
}
