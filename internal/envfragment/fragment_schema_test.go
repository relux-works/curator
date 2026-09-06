package envfragment

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/protocoljson"
)

// The fragment schema driver validates the published
// launch-env-fragment-v1 family against the closed adapter registry and
// the §10.3 boundary — the same model production emission obeys. Grammar
// alone cannot decide cases like wrong-adapter-variable or
// channel-not-registry, so the driver strict-decodes each instance and
// then checks registry membership, descriptor identity, sorted and
// unreserved env_names, and path discipline. The indexed set alone
// decides; published but unindexed files are logged and skipped.

// fragmentCase is the strict schema model: every member the schema
// allows, nothing it forbids.
type fragmentCase struct {
	Fragment     string            `json:"fragment"`
	Environment  string            `json:"environment"`
	Profile      profileCase       `json:"profile"`
	Precedence   precedenceCase    `json:"precedence"`
	Env          map[string]string `json:"env"`
	SystemPrompt *sectionCase      `json:"system_prompt"`
	MCP          *mcpCase          `json:"mcp"`
	PathPrepend  string            `json:"path_prepend"`
}

type profileCase struct {
	Name        string `json:"name"`
	LockSHA256  string `json:"lock_sha256"`
	Commit      string `json:"commit"`
	StateSHA256 string `json:"state_sha256"`
}

type precedenceCase struct {
	Winner    string `json:"winner"`
	Placement string `json:"placement"`
}

type sectionCase struct {
	Path     string            `json:"path"`
	Channels []json.RawMessage `json:"channels"`
}

type mcpCase struct {
	Path     string            `json:"path"`
	EnvNames []string          `json:"env_names"`
	Channels []json.RawMessage `json:"channels"`
}

// channelFields is one descriptor in decoded form.
type channelFields map[string]any

func decodeChannel(raw json.RawMessage) (channelFields, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var fields channelFields
	if err := decoder.Decode(&fields); err != nil {
		return nil, err
	}
	kind, _ := fields["kind"].(string)
	allowed := map[string]bool{"kind": true}
	switch kind {
	case "flag":
		allowed["flag"] = true
		allowed["argument"] = true
		allowed["name"] = true
		allowed["with"] = true
		allowed["semantics"] = true
	case "config-key":
		allowed["key"] = true
		allowed["semantics"] = true
	case "variable":
		allowed["variable"] = true
		allowed["semantics"] = true
	case "file":
		allowed["filename"] = true
		allowed["semantics"] = true
	default:
		return nil, &caseError{"unknown channel kind"}
	}
	for field := range fields {
		if !allowed[field] {
			return nil, &caseError{"unknown channel field"}
		}
	}
	argument, _ := fields["argument"].(string)
	switch argument {
	case "", "path", "contents", "name":
	default:
		return nil, &caseError{"unknown flag argument"}
	}
	if argument == "name" {
		if _, ok := fields["name"]; !ok {
			return nil, &caseError{"name argument without name"}
		}
	} else if _, ok := fields["name"]; ok {
		return nil, &caseError{"name without name argument"}
	}
	if semantics, ok := fields["semantics"].(string); ok {
		if semantics != "append" && semantics != "replace" {
			return nil, &caseError{"unknown semantics"}
		}
	} else if _, ok := fields["semantics"]; ok {
		return nil, &caseError{"bad semantics"}
	}
	return fields, nil
}

type caseError struct{ message string }

func (e *caseError) Error() string { return e.message }

// validHex reports 64 (or width) lowercase hex without any prefix.
func validHex(value string, width int) bool {
	if len(value) != width {
		return false
	}
	for i := 0; i < len(value); i++ {
		char := value[i]
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

// checkPath enforces absolute, ..-free paths for the evident test root.
func checkPath(value, root string) error {
	if value == "" {
		return &caseError{"empty path"}
	}
	if !filepath.IsAbs(value) {
		return &caseError{"relative path"}
	}
	for _, segment := range strings.Split(filepath.ToSlash(value), "/") {
		if segment == ".." {
			return &caseError{"dotdot path"}
		}
	}
	if value != root && !strings.HasPrefix(value, root+"/") {
		return &caseError{"out of root"}
	}
	return nil
}

// validateFragmentCase strict-decodes one instance and checks it against
// the registry. It returns nil exactly for the indexed-valid set.
func validateFragmentCase(raw []byte, root string) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var instance fragmentCase
	if err := decoder.Decode(&instance); err != nil {
		return err
	}
	if instance.Fragment != Version {
		return &caseError{"fragment identity"}
	}
	adapter, err := envregistry.ByID(instance.Environment)
	if err != nil {
		return err
	}
	if len(instance.Env) != 1 {
		return &caseError{"env arity"}
	}
	value, ok := instance.Env[adapter.EnvVar]
	if !ok {
		return &caseError{"wrong adapter variable"}
	}
	if err := checkPath(value, root); err != nil {
		return err
	}
	// The profile pin is the lock hash alone (§10.2): commit and state
	// pins belong to an older shape and are invalid here.
	if instance.Profile.Commit != "" || instance.Profile.StateSHA256 != "" {
		return &caseError{"profile pin form"}
	}
	if !validHex(instance.Profile.LockSHA256, 64) {
		return &caseError{"profile lock pin"}
	}
	if instance.Profile.Name == "" {
		return &caseError{"profile name"}
	}
	switch instance.Precedence.Winner {
	case "higher-weight", "lower-weight":
	default:
		return &caseError{"precedence winner"}
	}
	switch instance.Precedence.Placement {
	case "winner-last", "winner-first":
	default:
		return &caseError{"precedence placement"}
	}
	if instance.SystemPrompt != nil {
		if err := checkPath(instance.SystemPrompt.Path, root); err != nil {
			return err
		}
		if err := checkChannels(instance.SystemPrompt.Channels, adapter.SystemPrompt, true); err != nil {
			return err
		}
	}
	if instance.MCP != nil {
		if adapter.MCP == nil {
			return &caseError{"adapter declares no channel"}
		}
		if instance.MCP.EnvNames == nil {
			return &caseError{"mcp env_names required"}
		}
		if err := checkPath(instance.MCP.Path, root); err != nil {
			return err
		}
		if !sort.StringsAreSorted(instance.MCP.EnvNames) {
			return &caseError{"env_names unsorted"}
		}
		for _, name := range instance.MCP.EnvNames {
			if contextpkg.ReservedEnvName(name) {
				return &caseError{"env_names reserved"}
			}
		}
		if err := checkChannels(instance.MCP.Channels, []envregistry.Channel{*adapter.MCP}, false); err != nil {
			return err
		}
	}
	if instance.PathPrepend != "" {
		if err := checkPath(instance.PathPrepend, root); err != nil {
			return err
		}
	}
	return nil
}

// checkChannels decodes each descriptor and requires the sequence to equal
// the registered descriptors exactly: no subset, no extension, no
// reordering, and — for MCP — no semantics member.
func checkChannels(raw []json.RawMessage, registered []envregistry.Channel, semantics bool) error {
	if len(raw) != len(registered) {
		return &caseError{"channel arity"}
	}
	for i, item := range raw {
		fields, err := decodeChannel(item)
		if err != nil {
			return err
		}
		want, err := protocoljson.MarshalCanonical(channelObject(registered[i], semantics))
		if err != nil {
			return err
		}
		got, err := protocoljson.MarshalCanonical(map[string]any(fields))
		if err != nil {
			return err
		}
		if !bytes.Equal(got, want) {
			return &caseError{"channel not registry"}
		}
	}
	return nil
}

// TestFragmentAuthoritativeSchemaCases runs the published
// launch-env-fragment-v1 family through the registry model. The family is
// named explicitly, so a root that stops publishing it fails here instead
// of quietly narrowing the check.
func TestFragmentAuthoritativeSchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []struct {
		Instance string `json:"instance"`
		Valid    bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Instance, "launch-env-fragment-v1/") {
			continue
		}
		name := strings.TrimPrefix(entry.Instance, "launch-env-fragment-v1/")
		seen++
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, "schema-cases", entry.Instance)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			err = validateFragmentCase(raw, "/manager/environments")
			if entry.Valid && err != nil {
				t.Fatalf("valid case rejected: %v", err)
			}
			if !entry.Valid && err == nil {
				t.Fatal("invalid case accepted")
			}
		})
	}
	if seen == 0 {
		t.Fatal("the root publishes no launch-env-fragment-v1 cases")
	}
}

// TestFragmentEmissionMatchesReference proves the production emitter
// agrees with the published reference: the canonical claude fragment
// re-encodes the valid.json reference bytes exactly once the profile pin
// is normalized.
func TestFragmentEmissionMatchesReference(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	raw, err := os.ReadFile(filepath.Join(root, "schema-cases", "launch-env-fragment-v1", "valid.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]any
	if err := json.Unmarshal(raw, &reference); err != nil {
		t.Fatal(err)
	}
	profile := reference["profile"].(map[string]any)
	fragment := testFragment()
	profile["lock_sha256"] = fragment.LockSHA256
	want, err := protocoljson.MarshalCanonical(reference)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fragment.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("emission differs from the reference:\n got %q\nwant %q", got, want)
	}
}
