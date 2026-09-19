package envprofile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envfragment"
)

// TestEnvironmentsEnvPassthroughVectors executes every case of the
// environments-env-passthrough vector family
// (vectors/environments-env-passthrough.json) through the production
// entry points: envfragment.ResolvePassthrough for the S4 default
// resolution under both profiles, config.Parse for the schema cases,
// contextresolve.Resolve for the outside-allowlist refusal, Install,
// UpdateWithOptions, and StatusOf for the allowlist-warning operations,
// surfacingRows plus contextmaterialize.FormatDeclarationRows for the
// surfacing bytes, and live Install/UpdateWithOptions emission —
// observed at the filesystem while the rows arrive — for the
// install/update order.
//
// The committed CI pin serves the family, so a root that publishes no
// such vector fails here instead of taking a root-content skip.
func TestEnvironmentsEnvPassthroughVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "environments-env-passthrough.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		DefaultResolution []passthroughVectorCase `json:"default_resolution_cases"`
		AllowlistEmpty    []allowlistVectorCase   `json:"allowlist_empty_cases"`
		Schema            []schemaVectorCase      `json:"schema_cases"`
		Surfacing         []surfacingVectorCase   `json:"surfacing_cases"`
		Order             []surfacingOrderCase    `json:"surfacing_order_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	// Every group is fail-closed: a root that publishes the family with
	// a group missing fails here instead of quietly narrowing the check.
	if len(vectors.DefaultResolution) == 0 {
		t.Fatal("vectors/environments-env-passthrough.json publishes no default_resolution_cases")
	}
	if len(vectors.AllowlistEmpty) == 0 {
		t.Fatal("vectors/environments-env-passthrough.json publishes no allowlist_empty_cases")
	}
	if len(vectors.Schema) == 0 {
		t.Fatal("vectors/environments-env-passthrough.json publishes no schema_cases")
	}
	if len(vectors.Surfacing) == 0 {
		t.Fatal("vectors/environments-env-passthrough.json publishes no surfacing_cases")
	}
	if len(vectors.Order) == 0 {
		t.Fatal("vectors/environments-env-passthrough.json publishes no surfacing_order_cases")
	}
	for _, tc := range vectors.DefaultResolution {
		tc := tc
		t.Run("resolution/"+tc.Name, func(t *testing.T) {
			runPassthroughVectorCase(t, tc)
		})
	}
	for _, tc := range vectors.AllowlistEmpty {
		tc := tc
		t.Run("allowlist/"+tc.Name, func(t *testing.T) {
			runAllowlistVectorCase(t, tc)
		})
	}
	for _, tc := range vectors.Schema {
		tc := tc
		t.Run("schema/"+tc.Name, func(t *testing.T) {
			runPassthroughSchemaCase(t, tc)
		})
	}
	for _, tc := range vectors.Surfacing {
		tc := tc
		t.Run("surfacing/"+tc.Name, func(t *testing.T) {
			runSurfacingVectorCase(t, tc)
		})
	}
	for _, tc := range vectors.Order {
		tc := tc
		t.Run("order/"+tc.Name, func(t *testing.T) {
			runSurfacingOrderCase(t, tc)
		})
	}
}

// passthroughVectorCase is one default_resolution_cases entry: the knob
// is "absent", null, a list, or missing (absent); the profile is one
// profile or a profiles list; Effective is "unbounded" or a list.
type passthroughVectorCase struct {
	Name       string          `json:"name"`
	Profile    string          `json:"profile"`
	Profiles   []string        `json:"profiles"`
	Knob       json.RawMessage `json:"knob"`
	Requested  []string        `json:"requested"`
	Passed     []string        `json:"passed"`
	Dropped    []string        `json:"dropped"`
	Effective  json.RawMessage `json:"effective"`
	Diagnostic *string         `json:"diagnostic"`
	Conforming *bool           `json:"conforming"`
	HintKnob   bool            `json:"migration_hint_names_knob"`
	HintVars   bool            `json:"migration_hint_names_variables"`
}

// decodeKnob maps the vector knob spelling onto the resolver inputs: a
// missing knob or "absent" is an absent knob, null is the explicit
// unbounded null, and an array is the configured list.
func decodeKnob(t *testing.T, raw json.RawMessage) (passable []string, knobSet bool) {
	t.Helper()
	if len(raw) == 0 {
		return nil, false
	}
	// Null first: unmarshalling null into a string succeeds and leaves
	// "", so the string branch below must never see it.
	if string(raw) == "null" {
		return nil, true
	}
	var absent string
	if err := json.Unmarshal(raw, &absent); err == nil {
		if absent != "absent" {
			t.Fatalf("knob %q is not absent, null, or a list", absent)
		}
		return nil, false
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("knob %s is not absent, null, or a list", raw)
	}
	return list, true
}

func equalOrdered(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// runPassthroughVectorCase drives one default-resolution case through
// envfragment.ResolvePassthrough, the production entry buildFragment
// resolves every fragment with. A conforming:false case names a
// forbidden passed set, and the run proves this manager does not
// produce it.
func runPassthroughVectorCase(t *testing.T, tc passthroughVectorCase) {
	t.Helper()
	passable, knobSet := decodeKnob(t, tc.Knob)
	profiles := tc.Profiles
	if len(profiles) == 0 {
		profiles = []string{tc.Profile}
	}
	for _, profile := range profiles {
		verdict := envfragment.ResolvePassthrough(tc.Requested, passable, knobSet, envfragment.S4Profile(profile))
		if tc.Conforming != nil && !*tc.Conforming {
			if equalOrdered(verdict.Passed, tc.Passed) {
				t.Fatalf("%s resolves %v to the forbidden passed set %v", profile, tc.Requested, tc.Passed)
			}
			continue
		}
		if !equalOrdered(verdict.Passed, tc.Passed) {
			t.Fatalf("%s passed = %v, want %v", profile, verdict.Passed, tc.Passed)
		}
		if !equalOrdered(verdict.Dropped, tc.Dropped) {
			t.Fatalf("%s dropped = %v, want %v", profile, verdict.Dropped, tc.Dropped)
		}
		if tc.Diagnostic == nil {
			if verdict.Warning != "" {
				t.Fatalf("%s warning = %q, want silent", profile, verdict.Warning)
			}
		} else {
			if !strings.HasPrefix(verdict.Warning, *tc.Diagnostic+":") {
				t.Fatalf("%s warning = %q, want the %s diagnostic", profile, verdict.Warning, *tc.Diagnostic)
			}
			named := tc.Dropped
			if *tc.Diagnostic == envfragment.DiagPassthroughUnlisted {
				named = tc.Passed
			}
			for _, name := range named {
				if !strings.Contains(verdict.Warning, name) {
					t.Fatalf("%s warning %q names no %q", profile, verdict.Warning, name)
				}
			}
			if strings.Contains(verdict.Warning, "passable_env_names") != tc.HintKnob && *tc.Diagnostic == envfragment.DiagPassthroughUnlisted {
				t.Fatalf("%s warning %q names-knob = %v, want %v", profile, verdict.Warning, !tc.HintKnob, tc.HintKnob)
			}
			if *tc.Diagnostic == envfragment.DiagPassthroughUnlisted {
				namesVariables := true
				for _, name := range tc.Passed {
					namesVariables = namesVariables && strings.Contains(verdict.Warning, name)
				}
				if namesVariables != tc.HintVars {
					t.Fatalf("%s warning %q names-variables = %v, want %v", profile, verdict.Warning, namesVariables, tc.HintVars)
				}
				if !strings.Contains(verdict.Warning, envfragment.MigrationHint) {
					t.Fatalf("%s warning %q carries no migration hint", profile, verdict.Warning)
				}
			}
		}
		if len(tc.Effective) > 0 {
			effective := envfragment.EffectivePassable(passable, knobSet, envfragment.S4Profile(profile))
			trimmed := strings.TrimSpace(string(tc.Effective))
			switch {
			case trimmed == `"unbounded"`, trimmed == "null":
				if effective != nil {
					t.Fatalf("%s effective = %v, want unbounded", profile, effective)
				}
			case strings.HasPrefix(trimmed, "["):
				var want []string
				if err := json.Unmarshal(tc.Effective, &want); err != nil {
					t.Fatal(err)
				}
				if !equalOrdered(effective, want) {
					t.Fatalf("%s effective = %v, want %v", profile, effective, want)
				}
			default:
				t.Fatalf("effective %s is not unbounded or a list", tc.Effective)
			}
		}
	}
}

// allowlistVectorCase is one allowlist_empty_cases entry.
type allowlistVectorCase struct {
	Name           string   `json:"name"`
	Operation      string   `json:"operation"`
	Diagnostic     *string  `json:"diagnostic"`
	Allowlist      []string `json:"mcp_package_allowlist"`
	Admitted       string   `json:"admitted"`
	FailsOperation *bool    `json:"fails_operation"`
	RowCurrent     *bool    `json:"row_current"`
	Package        string   `json:"package"`
	Conforming     *bool    `json:"conforming"`
}

// runAllowlistVectorCase drives one allowlist case through the operation
// it names: Install, UpdateWithPolicy, or StatusOf for the warning, and
// contextresolve.Resolve for the outside-allowlist refusal.
func runAllowlistVectorCase(t *testing.T, tc allowlistVectorCase) {
	t.Helper()
	if tc.Conforming != nil && !*tc.Conforming {
		if got := AllowlistEmptyWarning(tc.Allowlist); got != "" {
			t.Fatalf("non-empty allowlist warns %q, want silent", got)
		}
		return
	}
	if tc.Operation == "profile-install" && tc.Diagnostic != nil &&
		*tc.Diagnostic == contextresolve.DiagMCPPackageNotAllowed {
		runAllowlistRefusalCase(t, tc)
		return
	}
	if tc.Diagnostic == nil {
		if got := AllowlistEmptyWarning(tc.Allowlist); got != "" {
			t.Fatalf("non-empty allowlist warns %q, want silent", got)
		}
		return
	}
	if *tc.Diagnostic != DiagAllowlistEmpty {
		t.Fatalf("case diagnostic %q is not mcp_package_allowlist_empty", *tc.Diagnostic)
	}
	warning := AllowlistEmptyWarning(tc.Allowlist)
	if warning == "" {
		t.Fatal("empty allowlist warns nothing, want mcp_package_allowlist_empty")
	}
	if !strings.HasPrefix(warning, DiagAllowlistEmpty+":") {
		t.Fatalf("warning %q carries no diagnostic code", warning)
	}
	if tc.Admitted != "" && !strings.Contains(warning, tc.Admitted) {
		t.Fatalf("warning %q states no %q", warning, tc.Admitted)
	}
	switch tc.Operation {
	case "profile-install":
		home := t.TempDir()
		pinHomes(t)
		root := filepath.Join(t.TempDir(), "root")
		writePackage(t, root, "acme", "1.0.0", "hello\n")
		info, _, _, err := Install(home, InstallOptions{Operand: root})
		if tc.FailsOperation != nil && !*tc.FailsOperation && err != nil {
			t.Fatalf("the warning fails the install: %v", err)
		}
		if err != nil {
			t.Fatal(err)
		}
		assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
	case "profile-update":
		home := t.TempDir()
		pinHomes(t)
		root := filepath.Join(t.TempDir(), "root")
		writePackage(t, root, "acme", "1.0.0", "hello\n")
		if _, _, _, err := Install(home, InstallOptions{Operand: root}); err != nil {
			t.Fatal(err)
		}
		info, moved, err := UpdateWithPolicy(home, "acme", Policy{})
		if tc.FailsOperation != nil && !*tc.FailsOperation && err != nil {
			t.Fatalf("the warning fails the update: %v", err)
		}
		if err != nil {
			t.Fatal(err)
		}
		if moved {
			t.Fatal("a path root with no overlays must not move on update")
		}
		assertWarningPresent(t, info.Warnings, DiagAllowlistEmpty)
	case "env-status":
		fx := writeManagedFixture(t, "acme")
		req := statusRequest(fx)
		req.Policy = Policy{MCPAllowlist: append([]string{}, tc.Allowlist...)}
		status, err := StatusOf(req)
		if err != nil {
			t.Fatal(err)
		}
		assertWarningPresent(t, status.Warnings, DiagAllowlistEmpty)
		if tc.RowCurrent != nil && *tc.RowCurrent {
			// The warning never makes a row non-current: the same
			// machine with a non-empty allowlist reports identical
			// currency.
			req.Policy = Policy{MCPAllowlist: []string{"https://example.com/m"}}
			quiet, err := StatusOf(req)
			if err != nil {
				t.Fatal(err)
			}
			if status.NonCurrent != quiet.NonCurrent {
				t.Fatalf("NonCurrent = %v with the warning, %v without: the warning must not affect currency",
					status.NonCurrent, quiet.NonCurrent)
			}
		}
	default:
		t.Fatalf("operation %q is not profile-install, profile-update, or env-status", tc.Operation)
	}
}

func assertWarningPresent(t *testing.T, warnings []string, diagnostic string) {
	t.Helper()
	for _, warning := range warnings {
		if strings.HasPrefix(warning, diagnostic+":") || warning == diagnostic {
			return
		}
	}
	t.Fatalf("warnings %v carry no %s", warnings, diagnostic)
}

// stubRefusalSource serves one root manifest and refuses to fetch: the
// allowlist gate fires on Identity alone, before any candidate is
// listed, so Candidates, ResolveTag, and Manifest must never run.
type stubRefusalSource struct {
	t    *testing.T
	root *contextresolve.Package
}

func (s *stubRefusalSource) Identity(kind, name, declared string) (string, error) {
	if declared == "" {
		return "", fmt.Errorf("no declared source for %s %s", kind, name)
	}
	return declared, nil
}

func (s *stubRefusalSource) Candidates(kind, name, _ string) ([]contextresolve.Candidate, error) {
	s.t.Fatalf("Candidates(%s %s) must not run past the allowlist refusal", kind, name)
	return nil, nil
}

func (s *stubRefusalSource) ResolveTag(kind, name, _, _ string) (string, error) {
	s.t.Fatalf("ResolveTag(%s %s) must not run past the allowlist refusal", kind, name)
	return "", nil
}

func (s *stubRefusalSource) Manifest(kind, name, _, _, _ string) (*contextresolve.Package, error) {
	s.t.Fatalf("Manifest(%s %s) must not run past the allowlist refusal", kind, name)
	return nil, nil
}

// runAllowlistRefusalCase drives the outside-allowlist case through
// contextresolve.Resolve, the production entry Install resolves with: a
// declaration package outside a non-empty allowlist is refused with
// mcp_package_not_allowed and the operation fails.
func runAllowlistRefusalCase(t *testing.T, tc allowlistVectorCase) {
	t.Helper()
	source := &stubRefusalSource{
		t: t,
		root: &contextresolve.Package{
			Version: "1.0.0",
			Requires: []contextresolve.Requirement{
				{Kind: contextlock.KindMCP, Name: "hostile-mcp", Source: tc.Package, Range: "*"},
			},
		},
	}
	_, err := contextresolve.Resolve(source, contextresolve.Input{
		Root:         contextresolve.Requirement{Kind: contextlock.KindContext, Name: "acme"},
		RootState:    &contextresolve.StatePackage{StateHash: strings.Repeat("a", 64), Manifest: source.root},
		MCPAllowlist: append([]string{}, tc.Allowlist...),
	})
	if err == nil {
		t.Fatal("a package outside a non-empty allowlist resolves, want mcp_package_not_allowed")
	}
	if !strings.HasPrefix(err.Error(), contextresolve.DiagMCPPackageNotAllowed+":") {
		t.Fatalf("err = %v, want the mcp_package_not_allowed diagnostic", err)
	}
	resolveErr, ok := err.(*contextresolve.Error)
	if !ok || resolveErr.Diagnostic != contextresolve.DiagMCPPackageNotAllowed {
		t.Fatalf("err = %#v, want the refusal diagnostic", err)
	}
	if tc.FailsOperation != nil && !*tc.FailsOperation {
		t.Fatalf("the case records fails_operation = false for a refusal")
	}
}

// schemaVectorCase is one schema_cases entry: the knob spelling and the
// effective value the config reader renders.
type schemaVectorCase struct {
	Name      string          `json:"name"`
	Knob      json.RawMessage `json:"knob"`
	Effective json.RawMessage `json:"effective"`
	Valid     bool            `json:"valid"`
}

// runPassthroughSchemaCase drives one schema case through config.Parse,
// the production config entry point: valid knobs render the expected
// effective passable_env_names, invalid knobs are rejected.
func runPassthroughSchemaCase(t *testing.T, tc schemaVectorCase) {
	t.Helper()
	object := map[string]any{
		"schema_version": float64(2),
		"skills_root":    "./skills",
		"projects":       map[string]any{},
	}
	if len(tc.Knob) > 0 && string(tc.Knob) != `"absent"` {
		var knob any
		if err := json.Unmarshal(tc.Knob, &knob); err != nil {
			t.Fatal(err)
		}
		object["environments"] = map[string]any{"passable_env_names": knob}
	}
	cfg, err := config.Parse(object, "vector.json")
	if !tc.Valid {
		if err == nil {
			t.Fatal("invalid knob accepted, want rejection")
		}
		if !strings.Contains(err.Error(), "passable_env_names") {
			t.Fatalf("err = %v, want the passable_env_names knob named", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("valid knob rejected: %v", err)
	}
	rendered, err := json.Marshal(cfg.EffectiveJSON())
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(rendered, &got); err != nil {
		t.Fatal(err)
	}
	env, ok := got["environments"].(map[string]any)
	if !ok {
		t.Fatalf("no environments object: %v", got)
	}
	gotMember, err := json.Marshal(env["passable_env_names"])
	if err != nil {
		t.Fatal(err)
	}
	var unbounded string
	want := tc.Effective
	if err := json.Unmarshal(tc.Effective, &unbounded); err == nil {
		if unbounded != "unbounded" {
			t.Fatalf("effective %q is not unbounded or a list", unbounded)
		}
		want = json.RawMessage("null")
	}
	// Both sides normalize through unmarshal: the vector's raw bytes
	// carry file whitespace the rendering does not.
	var gotNorm, wantNorm any
	if err := json.Unmarshal(gotMember, &gotNorm); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(want, &wantNorm); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%#v", gotNorm) != fmt.Sprintf("%#v", wantNorm) {
		t.Fatalf("passable_env_names renders %s, want %s", gotMember, want)
	}
}

// surfacingVectorCase is one surfacing_cases entry: declarations with
// byte-exact expectations, or a non-conforming row this manager must
// never emit.
type surfacingVectorCase struct {
	Name         string `json:"name"`
	Declarations []struct {
		Package   string   `json:"package"`
		Version   string   `json:"version"`
		Transport string   `json:"transport"`
		Command   string   `json:"command"`
		Args      []string `json:"args"`
		EnvNames  []string `json:"env_names"`
	} `json:"declarations"`
	ExpectedBytes      string `json:"expected_bytes"`
	ExpectedByteLength int    `json:"expected_byte_length"`
	ExpectedSHA256     string `json:"expected_sha256"`
	Conforming         *bool  `json:"conforming"`
	Row                string `json:"row"`
}

// runSurfacingVectorCase drives one surfacing case through
// contextmaterialize.FormatDeclarationRows, the production renderer
// every surfacing path formats with: positive cases match byte for byte
// (bytes, length, sha256), and a non-conforming row is proven absent by
// rendering the declaration it purports to show and showing this
// manager's bytes differ.
func runSurfacingVectorCase(t *testing.T, tc surfacingVectorCase) {
	t.Helper()
	if tc.Conforming != nil && !*tc.Conforming {
		declaration := parseVectorRow(t, tc.Row)
		rendered := contextmaterialize.FormatDeclarationRows([]contextmaterialize.MCPDeclaration{declaration})
		if strings.TrimSuffix(rendered, "\n") == tc.Row {
			t.Fatalf("this manager emits the non-conforming row %q", tc.Row)
		}
		return
	}
	declarations := make([]contextmaterialize.MCPDeclaration, 0, len(tc.Declarations))
	for _, entry := range tc.Declarations {
		declarations = append(declarations, contextmaterialize.MCPDeclaration{
			Package: entry.Package, Version: entry.Version,
			Transport: entry.Transport, Command: entry.Command,
			Args: entry.Args, EnvNames: entry.EnvNames,
		})
	}
	rendered := contextmaterialize.FormatDeclarationRows(declarations)
	if rendered != tc.ExpectedBytes {
		t.Fatalf("bytes:\n got %q\nwant %q", rendered, tc.ExpectedBytes)
	}
	if len(rendered) != tc.ExpectedByteLength {
		t.Fatalf("byte length = %d, want %d", len(rendered), tc.ExpectedByteLength)
	}
	if got := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(rendered))); got != tc.ExpectedSHA256 {
		t.Fatalf("sha256 = %s, want %s", got, tc.ExpectedSHA256)
	}
}

// parseVectorRow parses a non-conforming vector row back into the
// declaration it purports to show, so the run can prove this manager
// renders different bytes for the same declaration. A missing env_names
// column reads as the empty list.
func parseVectorRow(t *testing.T, row string) contextmaterialize.MCPDeclaration {
	t.Helper()
	fields := strings.Split(row, " ")
	if len(fields) < 4 || fields[0] != "mcp-declaration" {
		t.Fatalf("row %q is not a surfacing row", row)
	}
	declaration := contextmaterialize.MCPDeclaration{
		Package: fields[1], Version: fields[2], Transport: fields[3],
		EnvNames: []string{},
	}
	for _, field := range fields[4:] {
		name, value, ok := strings.Cut(field, "=")
		if !ok {
			t.Fatalf("row %q carries a malformed %q column", row, field)
		}
		switch name {
		case "command":
			declaration.Command = value
		case "args":
			if err := json.Unmarshal([]byte(value), &declaration.Args); err != nil {
				t.Fatalf("row %q args: %v", row, err)
			}
		case "env_names":
			if err := json.Unmarshal([]byte(value), &declaration.EnvNames); err != nil {
				t.Fatalf("row %q env_names: %v", row, err)
			}
		default:
			t.Fatalf("row %q carries an unknown %q column", row, name)
		}
	}
	return declaration
}

// surfacingOrderCase is one surfacing_order_cases entry: the operation
// and the mandated audit-gate, surfacing, lock-published,
// materialization order.
type surfacingOrderCase struct {
	Name      string   `json:"name"`
	Operation string   `json:"operation"`
	Order     []string `json:"order"`
}

// runSurfacingOrderCase executes one order case by observing the runtime
// event order through the production entry points: a blocked audit emits
// no rows to the sink and publishes nothing (surfacing strictly follows
// the audit gate), and a live operation prints the exact rows before the
// lock is published (surfacing strictly precedes publication and
// materialization). The vector's own order is pinned first, so a vector
// that reorders the steps fails here instead of executing a stale
// reading.
func runSurfacingOrderCase(t *testing.T, tc surfacingOrderCase) {
	t.Helper()
	if !equalOrdered(tc.Order, []string{"audit-gate", "surfacing", "lock-published", "materialization"}) {
		t.Fatalf("order = %v, want audit-gate, surfacing, lock-published, materialization", tc.Order)
	}
	switch tc.Operation {
	case "profile-install":
		home := t.TempDir()
		pinHomes(t)
		root := filepath.Join(t.TempDir(), "root")
		writeManifestPackage(t, root,
			`{"schema_version": 1, "name": "acme", "version": "1.0.0",`+
				`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
			map[string]string{"a.md": directorySecret()})
		blocked := &lockObservingSink{lock: lockPath(home, "acme")}
		info, _, _, err := Install(home, InstallOptions{Operand: root, SurfacingSink: blocked})
		if err == nil || !strings.Contains(err.Error(), "blocking") {
			t.Fatalf("err = %v, want the blocking-finding refusal", err)
		}
		if len(info.Surfacing) != 0 || len(blocked.rows) != 0 {
			t.Fatalf("a blocked install surfaces %v/%v, want no rows", info.Surfacing, blocked.rows)
		}
		observeLiveInstallEmission(t)
	case "profile-update":
		home, _, oldHash := installBlockingOverlayRoot(t)
		overlay := filepath.Join(t.TempDir(), "overlay")
		writeManifestPackage(t, overlay,
			`{"schema_version": 1, "name": "personal", "version": "0.3.0",`+
				`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
			map[string]string{"a.md": directorySecret()})
		policy := Policy{
			OverlaysAllowed:      true,
			OverlayDefaultWeight: 1000,
			Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
		}
		blocked := &lockObservingSink{lock: lockPath(home, "acme")}
		info, _, err := UpdateWithOptions(home, "acme", UpdateOptions{Policy: policy, SurfacingSink: blocked})
		if err == nil || !strings.Contains(err.Error(), DiagUpdateBlocked) {
			t.Fatalf("err = %v, want %s", err, DiagUpdateBlocked)
		}
		if len(info.Surfacing) != 0 || len(blocked.rows) != 0 {
			t.Fatalf("a blocked update surfaces %v/%v, want no rows", info.Surfacing, blocked.rows)
		}
		if _, afterHash, err := readLock(home, "acme"); err != nil {
			t.Fatal(err)
		} else if afterHash != oldHash {
			t.Fatal("a blocked update moved the lock: the old lock must stand")
		}
		observeLiveUpdateEmission(t)
	default:
		t.Fatalf("operation %q is not profile-install or profile-update", tc.Operation)
	}
}
