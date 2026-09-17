package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
)

func providersDoc(providers string) string {
	return `{"schema_version":1,"providers":{` + providers + `}}`
}

func mustParseProviders(t *testing.T, doc string) buildrepo.ProviderSet {
	t.Helper()
	set, err := ParseSourceProviders([]byte(doc), "source-providers.json")
	if err != nil {
		t.Fatalf("ParseSourceProviders: %v", err)
	}
	return set
}

func mustRejectProviders(t *testing.T, doc string) {
	t.Helper()
	_, err := ParseSourceProviders([]byte(doc), "source-providers.json")
	if err == nil {
		t.Fatalf("ParseSourceProviders accepted %s", doc)
	}
	if !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("error %q lacks class %q", err, CodeRepositoryPolicyInvalid)
	}
}

func TestParseSourceProvidersDocumentShape(t *testing.T) {
	valid := mustParseProviders(t, providersDoc(
		`"team-https":{"https":{"username":"ci"}},`+
			`"anon":{"https":{"anonymous":true}},`+
			`"team-ssh":{"ssh":{"identity":"/operator/id","known_hosts":"/operator/known_hosts"}},`+
			`"both":{"https":{},"ssh":{"agent_socket":"/operator/agent"}}`))
	if len(valid) != 4 {
		t.Fatalf("providers = %v", valid)
	}
	if valid["team-https"].HTTPS == nil || valid["team-https"].HTTPS.Username != "ci" || valid["team-https"].HTTPS.Anonymous {
		t.Fatalf("team-https = %+v", valid["team-https"].HTTPS)
	}
	if valid["anon"].HTTPS == nil || !valid["anon"].HTTPS.Anonymous {
		t.Fatalf("anon = %+v", valid["anon"].HTTPS)
	}
	if valid["team-ssh"].SSH == nil || valid["team-ssh"].SSH.Identity != "/operator/id" || valid["team-ssh"].SSH.KnownHosts == "" {
		t.Fatalf("team-ssh = %+v", valid["team-ssh"].SSH)
	}
	if valid["both"].HTTPS == nil || valid["both"].SSH == nil || valid["both"].SSH.AgentSocket == "" {
		t.Fatalf("both = %+v", valid["both"])
	}
	if valid["team-https"].SSH != nil || valid["team-ssh"].HTTPS != nil {
		t.Fatalf("transports leak across entries: %+v", valid)
	}
	empty := mustParseProviders(t, providersDoc(""))
	if len(empty) != 0 {
		t.Fatalf("providers = %v", empty)
	}
	for _, doc := range []string{
		`not json`,
		`{"schema_version":1,"providers":{}}trailing`,
		`["schema_version"]`,
		`{"schema_version":1,"providers":{},"unexpected":true}`,
		`{"schema_version":1}`,
		`{"schema_version":1,"providers":null}`,
		`{"schema_version":1,"providers":[]}`,
		`{"providers":{}}`,
		`{"schema_version":"1","providers":{}}`,
		`{"schema_version":1.5,"providers":{}}`,
		`{"schema_version":0,"providers":{}}`,
		`{"schema_version":2,"providers":{}}`,
		`{"schema_version":null,"providers":{}}`,
		providersDoc(`"not a provider!":{"https":{"anonymous":true}}`),
		providersDoc(`"team-https":[]`),
		providersDoc(`"team-https":{"https":{"anonymous":true},"token":"secret"}`),
		providersDoc(`"team-https":{}`),
		providersDoc(`"team-https":{"https":null,"ssh":null}`),
		providersDoc(`"team-https":{"https":[]}`),
		providersDoc(`"team-https":{"https":{"username":7}}`),
		providersDoc(`"team-https":{"https":{"anonymous":"yes"}}`),
		providersDoc(`"team-https":{"https":{"username":"ci","anonymous":true}}`),
		providersDoc(`"team-https":{"https":{"username":"ci","password":"secret"}}`),
		providersDoc(`"team-ssh":{"ssh":[]}`),
		providersDoc(`"team-ssh":{"ssh":{}}`),
		providersDoc(`"team-ssh":{"ssh":{"known_hosts":"/operator/known_hosts"}}`),
		providersDoc(`"team-ssh":{"ssh":{"identity":7}}`),
		providersDoc(`"team-ssh":{"ssh":{"identity":"/operator/id","command":"ssh"}}`),
	} {
		mustRejectProviders(t, doc)
	}
}

func TestLoadSourceProviders(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent-providers.json")
	set, err := LoadSourceProviders(missing)
	if err != nil || set != nil {
		t.Fatalf("missing = %v, %v; want nil, nil", set, err)
	}
	path := filepath.Join(t.TempDir(), "source-providers.json")
	if err := os.WriteFile(path, []byte(providersDoc(`"team-https":{"https":{"username":"ci"}}`)), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err = LoadSourceProviders(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 1 || set["team-https"].HTTPS == nil || set["team-https"].HTTPS.Username != "ci" {
		t.Fatalf("set = %+v", set)
	}
	if err := os.WriteFile(path, []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSourceProviders(path); err == nil || !strings.Contains(err.Error(), CodeRepositoryPolicyInvalid) {
		t.Fatalf("invalid err = %v, want %s", err, CodeRepositoryPolicyInvalid)
	}
}
