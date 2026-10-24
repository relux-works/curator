package envmarker

import (
	"encoding/json"
	"os"
	"testing"
)

// Full inherited node carriers accept A/B/C without branch substitution, and
// each invalid neighbor is refused by Parse rather than only a data checker.
func TestMarker7FullCarrier(t *testing.T) {
	raw, err := os.ReadFile("testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	var base map[string]any
	if err = json.Unmarshal(raw, &base); err != nil {
		t.Fatal(err)
	}
	clone := func() map[string]any {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	check := func(t *testing.T, m map[string]any, valid bool) {
		t.Helper()
		b, _ := json.Marshal(m)
		marker, err := Parse(b)
		if valid {
			if err != nil {
				t.Fatalf("full marker 7 positive: %v", err)
			}
			encoded, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var round map[string]any
			_ = json.Unmarshal(encoded, &round)
			if _, err := Parse(encoded); err != nil {
				t.Fatal(err)
			}
			if string(mustJSON(t, m["credential_source"])) != string(mustJSON(t, round["credential_source"])) {
				t.Fatal("credential source identity changed")
			}
		} else if err == nil {
			t.Fatal("invalid carrier accepted")
		}
	}
	for _, revision := range []string{"A", "B", "C"} {
		t.Run("node-"+revision, func(t *testing.T) {
			m := clone()
			m["codex_seed_record"].(map[string]any)["revision"] = revision
			check(t, m, true)
		})
	}
	for _, revision := range []string{"A", "B", "C"} {
		t.Run("per-home-"+revision, func(t *testing.T) {
			m := clone()
			m["credential_source"] = map[string]any{"mode": "per-home"}
			m["seeds"] = []any{"config.toml"}
			m["codex_seed_record"].(map[string]any)["revision"] = revision
			check(t, m, true)
		})
	}
	negatives := map[string]func(map[string]any){
		"missing-settings": func(m map[string]any) { delete(m, "codex_settings") },
		"null-settings":    func(m map[string]any) { m["codex_settings"] = nil },
		"declared-trust":   func(m map[string]any) { m["codex_settings"] = map[string]any{"projects": map[string]any{}} },
		"missing-source":   func(m map[string]any) { delete(m, "credential_source") },
		"null-source":      func(m map[string]any) { m["credential_source"] = nil },
		"invalid-ref": func(m map[string]any) {
			m["credential_source"].(map[string]any)["source"].(map[string]any)["source_ref"] = "local:invalid"
		},
		"ref-newline": func(m map[string]any) {
			m["credential_source"].(map[string]any)["source"].(map[string]any)["source_ref"] = "local:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"
		},
		"extra-source":    func(m map[string]any) { m["credential_source"].(map[string]any)["extra"] = true },
		"zero-generation": func(m map[string]any) { m["credential_source"].(map[string]any)["source_generation"] = 0 },
		"over-generation": func(m map[string]any) {
			m["credential_source"].(map[string]any)["policy_generation"] = 9007199254740992
		},
		"node-seeds": func(m map[string]any) { m["seeds"] = []any{"config.toml"} },
		"node-passthrough": func(m map[string]any) {
			m["passthrough"] = []any{map[string]any{"path": "auth.json", "isolation": "shared", "strategy": "file-link", "source_role": "native", "backend": "file", "backend_version": "1", "provenance": "provisioned"}}
		},
		"node-links": func(m map[string]any) { m["seed_links"] = []any{"config"} },
		"wrong-channel": func(m map[string]any) {
			m["credential_source"].(map[string]any)["source"].(map[string]any)["harness_channel"] = "OTHER"
		},
		"wrong-hash": func(m map[string]any) { m["hash_version"] = 1 },
		"extra-top":  func(m map[string]any) { m["extra"] = true },
	}
	for name, mutate := range negatives {
		t.Run(name, func(t *testing.T) { m := clone(); mutate(m); check(t, m, false) })
	}
	// Earlier adopted carriers keep their closed A/B revision set.
	for _, version := range []int{1, 2, 3} {
		m := clone()
		m["version"] = version
		delete(m, "codex_settings")
		delete(m, "credential_source")
		if version != 3 {
			delete(m, "hash_version")
		}
		check(t, m, false)
		m["codex_seed_record"].(map[string]any)["revision"] = "B"
		check(t, m, true)
		m["credential_source"] = nil
		check(t, m, false)
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
