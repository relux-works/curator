// Package interop consumes the authoritative external Curator Protocol suite.
// It contains no implementation-owned expected values.
package interop

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

func suiteRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(absolute, "manifest.json")); err != nil {
		t.Fatalf("invalid CURATOR_CONFORMANCE_ROOT: %v", err)
	}
	return absolute
}

func TestGoldenMarkerObject(t *testing.T) {
	dir := t.TempDir()
	wantContextHash := readGolden(t, "expected/context_sha256.txt")
	wantFiles := []string{".skill_triggers/en.md", "SKILL.md", "references/notes.md"}
	generated := &marker.Marker{
		Name: "golden-skill", Source: "golden-skill",
		RefKind: "revision", Ref: "0123456789abcdef0123456789abcdef01234567",
		Commit: "0123456789abcdef0123456789abcdef01234567", ContentSHA256: wantContextHash,
		Agents: []string{"codex_cli"}, Commands: []string{"golden-tool"},
		SkillSchemaVersion: 5, RuntimeRoots: []string{"scripts"},
		InstalledAt: "2000-01-01T00:00:00Z", Files: wantFiles,
		Activation: &marker.Activation{Context: true, Commands: []string{"golden-tool"}},
		Requirers:  []string{"<project>"},
	}
	if err := marker.Write(dir, generated); err != nil {
		t.Fatal(err)
	}
	actualPayload, err := os.ReadFile(filepath.Join(dir, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	wantPayload, err := os.ReadFile(golden(t, "expected/marker.json"))
	if err != nil {
		t.Fatal(err)
	}
	var actual, want any
	if err := json.Unmarshal(actualPayload, &actual); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantPayload, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("marker object diverges from golden:\n got %s\nwant %s", actualPayload, wantPayload)
	}
}

func golden(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join(suiteRoot(t), filepath.FromSlash(rel))
}

func readGolden(t *testing.T, rel string) string {
	t.Helper()
	payload, err := os.ReadFile(golden(t, rel))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(payload))
}

func TestGoldenSnapshotHash(t *testing.T) {
	got, err := hashing.ContentSHA256(golden(t, "fixtures/skill"), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := readGolden(t, "expected/snapshot_sha256.txt")
	if got != want {
		t.Fatalf("snapshot hash diverges from golden:\n got %s\nwant %s", got, want)
	}
}

func TestGoldenContextCopy(t *testing.T) {
	fixture := golden(t, "fixtures/skill")
	spec, err := skillspec.Load(fixture)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "ctx")
	includeScripts := len(spec.Commands) == 0
	files, err := whitelist.CopyContext(fixture, dest, includeScripts, spec.RuntimeRoots)
	if err != nil {
		t.Fatal(err)
	}

	var wantFiles []string
	if err := json.Unmarshal([]byte(readGolden(t, "expected/context_files.json")), &wantFiles); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := json.Marshal(files)
	wantJSON, _ := json.Marshal(wantFiles)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("context file list diverges:\n got %s\nwant %s", gotJSON, wantJSON)
	}

	gotHash, err := hashing.ContentSHA256(dest, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantHash := readGolden(t, "expected/context_sha256.txt")
	if gotHash != wantHash {
		t.Fatalf("context hash diverges from golden:\n got %s\nwant %s", gotHash, wantHash)
	}
}

func TestSkillManifestResolutionVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/skill-manifest-resolution.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name             string            `json:"name"`
		Files            map[string]string `json:"files"`
		ExpectedSource   *string           `json:"expected_source"`
		ExpectedCommands []string          `json:"expected_commands"`
		Error            string            `json:"error"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			dir := t.TempDir()
			for rel, content := range testCase.Files {
				path := filepath.Join(dir, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			spec, err := skillspec.Load(dir)
			if testCase.Error != "" {
				if err == nil {
					t.Fatalf("manifest accepted, want %s", testCase.Error)
				}
				if testCase.Error == "conflicting_skill_manifests" && !strings.Contains(err.Error(), testCase.Error) {
					t.Fatalf("error %q does not contain %q", err, testCase.Error)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantSource := ""
			if testCase.ExpectedSource != nil {
				wantSource = *testCase.ExpectedSource
			}
			if spec.SourceFile != wantSource {
				t.Fatalf("source = %q, want %q", spec.SourceFile, wantSource)
			}
			var commands []string
			for name := range spec.Commands {
				commands = append(commands, name)
			}
			sort.Strings(commands)
			if len(commands) == 0 {
				commands = []string{}
			}
			if !reflect.DeepEqual(commands, testCase.ExpectedCommands) {
				t.Fatalf("commands = %v, want %v", commands, testCase.ExpectedCommands)
			}
		})
	}
}

func readSigned(t *testing.T, rel string) map[string]any {
	t.Helper()
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader([]byte(readGolden(t, rel))))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestGoldenRegistryObjects(t *testing.T) {
	pinned := readGolden(t, "expected/registry/pinned_key.txt")
	keys := []string{pinned}

	audited := readSigned(t, "expected/registry/record_audited.json")
	if !registry.VerifySigned(audited, keys) {
		t.Fatal("golden audited record must verify against the pinned key")
	}
	revoked := readSigned(t, "expected/registry/record_revoked.json")
	if !registry.VerifySigned(revoked, keys) {
		t.Fatal("golden revoked record must verify")
	}
	forged := readSigned(t, "expected/registry/record_forged.json")
	if registry.VerifySigned(forged, keys) {
		t.Fatal("forged record must not verify")
	}
	wrongKeyID := readSigned(t, "expected/registry/record_wrong_key_id.json")
	if registry.VerifySigned(wrongKeyID, keys) {
		t.Fatal("record with a wrong key id must not verify")
	}
	snapshot := readSigned(t, "expected/registry/snapshot.json")
	if !registry.VerifySigned(snapshot, keys) {
		t.Fatal("golden snapshot must verify")
	}

	// wrong key rejects everything
	other := "ed25519:" + strings.Repeat("A", 43) + "="
	if registry.VerifySigned(audited, []string{other}) {
		t.Fatal("wrong key must not verify")
	}
}

func TestCanonicalJSONVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/canonical-valid.json"))
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var cases []struct {
		Name      string         `json:"name"`
		Input     map[string]any `json:"input"`
		Canonical string         `json:"canonical_utf8"`
	}
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		got, err := registry.CanonicalBytesChecked(testCase.Input)
		if err != nil {
			t.Fatalf("%s: %v", testCase.Name, err)
		}
		if string(got) != testCase.Canonical {
			t.Fatalf("%s canonical bytes:\n got %s\nwant %s", testCase.Name, got, testCase.Canonical)
		}
	}

	invalidPayload, err := os.ReadFile(golden(t, "vectors/canonical-invalid.json"))
	if err != nil {
		t.Fatal(err)
	}
	var invalidCases []struct {
		Name      string `json:"name"`
		InputText string `json:"input_text"`
	}
	if err := json.Unmarshal(invalidPayload, &invalidCases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range invalidCases {
		raw := []byte(testCase.InputText)
		if err := protocoljson.Validate(raw); err != nil {
			continue
		}
		var input map[string]any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&input); err != nil {
			continue
		}
		if _, err := registry.CanonicalBytesChecked(input); err == nil {
			t.Errorf("%s CCJ-1 input was accepted, want rejection", testCase.Name)
		}
	}
}

func TestSourceIdentityVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/source-identities.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input    string  `json:"input"`
		Identity *string `json:"identity"`
		Error    string  `json:"error"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		got, err := identity.Parse(testCase.Input)
		if testCase.Error != "" {
			if err == nil {
				t.Errorf("Parse(%q) = %q, want error %s", testCase.Input, got, testCase.Error)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q): %v", testCase.Input, err)
			continue
		}
		want := ""
		if testCase.Identity != nil {
			want = *testCase.Identity
		}
		if got != want {
			t.Errorf("Parse(%q) = %q, want %q", testCase.Input, got, want)
		}
	}
}

func TestIdentifierVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/identifiers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		if got := identifiers.Valid(testCase.Input); got != testCase.Valid {
			t.Errorf("identifier %q valid=%v, want %v", testCase.Input, got, testCase.Valid)
		}
	}
}

func TestLocaleSelectorVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/locale-selectors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		if got := identifiers.ValidLocale(testCase.Input); got != testCase.Valid {
			t.Errorf("locale selector %q valid=%v, want %v", testCase.Input, got, testCase.Valid)
		}
	}
}

func TestManagerConfigVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/manager-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string         `json:"name"`
		Input    map[string]any `json:"input"`
		Valid    bool           `json:"valid"`
		Expected struct {
			DefaultAgents            []string `json:"default_agents"`
			AdapterMode              string   `json:"adapter_mode"`
			ProjectAlias             string   `json:"project_alias"`
			CheckoutAlias            string   `json:"checkout_alias"`
			RegistryURLs             []string `json:"registry_urls"`
			SnapshotMaxAgeSeconds    int      `json:"snapshot_max_age_seconds"`
			SnapshotClockSkewSeconds int      `json:"snapshot_clock_skew_seconds"`
			CacheTTLSeconds          int      `json:"cache_ttl_seconds"`
			OfflineGraceSeconds      int      `json:"offline_grace_seconds"`
			MaxRequestBytes          int      `json:"max_request_bytes"`
		} `json:"expected"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			parsed, err := config.Parse(testCase.Input, "config.json")
			if !testCase.Valid {
				if err == nil {
					t.Fatal("manager config was accepted, want rejection")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(parsed.DefaultAgents, testCase.Expected.DefaultAgents) || parsed.AdapterMode != testCase.Expected.AdapterMode {
				t.Fatalf("defaults diverge: %+v", parsed)
			}
			var urls []string
			for _, item := range parsed.AuditRegistries {
				urls = append(urls, item.URL)
			}
			if urls == nil {
				urls = []string{}
			}
			if !reflect.DeepEqual(urls, testCase.Expected.RegistryURLs) {
				t.Fatalf("registry URLs = %v, want %v", urls, testCase.Expected.RegistryURLs)
			}
			if testCase.Expected.ProjectAlias != "" {
				project := parsed.Projects["app"]
				if project.ProjectAlias != testCase.Expected.ProjectAlias || project.CheckoutAlias != testCase.Expected.CheckoutAlias {
					t.Fatalf("project aliases = %+v", project)
				}
			}
			if parsed.Audit.SnapshotMaxAgeSeconds != testCase.Expected.SnapshotMaxAgeSeconds ||
				parsed.Audit.SnapshotClockSkewSeconds != testCase.Expected.SnapshotClockSkewSeconds ||
				parsed.Audit.CacheTTLSeconds != testCase.Expected.CacheTTLSeconds ||
				parsed.Audit.OfflineGraceSeconds != testCase.Expected.OfflineGraceSeconds ||
				parsed.Audit.MaxRequestBytes != testCase.Expected.MaxRequestBytes {
				t.Fatalf("audit defaults = %+v, want %+v", parsed.Audit, testCase.Expected)
			}
		})
	}
}

func TestManagerLifecycleVectors(t *testing.T) {
	payload, err := os.ReadFile(golden(t, "vectors/manager-lifecycle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		LauncherCases []struct {
			Name                  string   `json:"name"`
			Platforms             []string `json:"platforms"`
			RequiredPathRoles     []string `json:"required_path_roles"`
			PreserveInheritedPath bool     `json:"preserve_inherited_path"`
			ForwardArguments      bool     `json:"forward_arguments"`
			PreserveExitStatus    bool     `json:"preserve_exit_status"`
		} `json:"launcher_cases"`
		BootstrapCases []struct {
			Name    string `json:"name"`
			Outcome string `json:"outcome"`
		} `json:"bootstrap_cases"`
		UpgradeCases []struct {
			Name        string `json:"name"`
			Deduplicate bool   `json:"deduplicate"`
		} `json:"upgrade_cases"`
		DryRunCases []struct {
			Name                       string   `json:"name"`
			ForbiddenPersistentEffects []string `json:"forbidden_persistent_effects"`
		} `json:"dry_run_cases"`
	}
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.LauncherCases) != 2 || len(vector.BootstrapCases) != 3 || len(vector.UpgradeCases) != 3 || len(vector.DryRunCases) != 2 {
		t.Fatalf("manager lifecycle vector is incomplete: %+v", vector)
	}
	for _, testCase := range vector.LauncherCases {
		if !reflect.DeepEqual(testCase.Platforms, []string{"unix", "windows"}) ||
			len(testCase.RequiredPathRoles) != 3 || !testCase.PreserveInheritedPath ||
			!testCase.ForwardArguments || !testCase.PreserveExitStatus {
			t.Fatalf("launcher case %s is incomplete: %+v", testCase.Name, testCase)
		}
	}
	outcomes := map[string]string{}
	for _, testCase := range vector.BootstrapCases {
		outcomes[testCase.Name] = testCase.Outcome
	}
	if outcomes["existing-config-if-missing"] != "unchanged-success" || outcomes["if-missing-with-force"] != "usage-error" {
		t.Fatalf("bootstrap outcomes: %v", outcomes)
	}
	deduplicated := false
	for _, testCase := range vector.UpgradeCases {
		if testCase.Name == "all-projects-deduplicate" {
			deduplicated = testCase.Deduplicate
		}
	}
	if !deduplicated {
		t.Fatal("manager lifecycle vector does not require cross-project fetch deduplication")
	}
	for _, testCase := range vector.DryRunCases {
		if len(testCase.ForbiddenPersistentEffects) < 8 {
			t.Fatalf("dry-run case %s omits persistent surfaces: %v", testCase.Name, testCase.ForbiddenPersistentEffects)
		}
	}
}

func TestGoldenFederationSemantics(t *testing.T) {
	pinned := readGolden(t, "expected/registry/pinned_key.txt")
	audited := readSigned(t, "expected/registry/record_audited.json")
	revoked := readSigned(t, "expected/registry/record_revoked.json")
	snapHash := readGolden(t, "expected/snapshot_sha256.txt")

	registries := []registry.Registry{
		{Name: "golden-one", URL: "https://one", PublicKeys: []string{pinned}},
		{Name: "golden-two", URL: "https://two", PublicKeys: []string{pinned}},
	}
	fetch := func(url, _, _, _ string) ([]map[string]any, error) {
		if url == "https://one" {
			return []map[string]any{audited}, nil
		}
		return []map[string]any{revoked}, nil
	}
	resolution := registry.Resolve(registries,
		"git.example.com/skills/golden-skill",
		"0123456789abcdef0123456789abcdef01234567",
		snapHash, fetch)
	if resolution.Result != registry.ResultRevoked {
		t.Fatalf("deny-wins over golden records: %+v", resolution)
	}
}

type registryClientVector struct {
	StateKey               string `json:"state_key"`
	KeyRotationResetsState bool   `json:"key_rotation_resets_state"`
	RetryCases             []struct {
		Name           string `json:"name"`
		Method         string `json:"method"`
		Outcome        string `json:"outcome"`
		IdempotencyKey bool   `json:"idempotency_key"`
		RetryPermitted bool   `json:"retry_permitted"`
	} `json:"retry_cases"`
	RetryPolicy struct {
		MaxAttempts              int  `json:"max_attempts"`
		GetTotalDeadlineSeconds  int  `json:"get_total_deadline_seconds"`
		PostTotalDeadlineSeconds int  `json:"post_total_deadline_seconds"`
		FollowRedirects          bool `json:"follow_redirects"`
	} `json:"retry_policy"`
	SnapshotTransitions []struct {
		Name             string `json:"name"`
		StoredVersion    int    `json:"stored_version"`
		CandidateVersion int    `json:"candidate_version"`
		SameBody         bool   `json:"same_body"`
		Accepted         bool   `json:"accepted"`
	} `json:"snapshot_transitions"`
	PaginationRejections []struct {
		Name       string `json:"name"`
		Error      string `json:"error"`
		Characters int    `json:"characters"`
		Records    int    `json:"records"`
		Bytes      int    `json:"bytes"`
	} `json:"pagination_rejections"`
	RollbackStateCases []struct {
		Name     string `json:"name"`
		State    string `json:"state"`
		Accepted bool   `json:"accepted"`
	} `json:"rollback_state_cases"`
}

func loadRegistryClientVector(t *testing.T) registryClientVector {
	t.Helper()
	payload, err := os.ReadFile(golden(t, "vectors/registry-client.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vector registryClientVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	return vector
}

func TestRegistryClientRetryVectors(t *testing.T) {
	vector := loadRegistryClientVector(t)
	for _, testCase := range vector.RetryCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := registry.RetryPermitted(testCase.Method, testCase.Outcome, testCase.IdempotencyKey)
			if got != testCase.RetryPermitted {
				t.Fatalf("RetryPermitted(%q, %q, %v) = %v, want %v", testCase.Method, testCase.Outcome, testCase.IdempotencyKey, got, testCase.RetryPermitted)
			}
		})
	}
}

func TestRegistryClientRetryExecutionPolicy(t *testing.T) {
	vector := loadRegistryClientVector(t)
	policy := vector.RetryPolicy
	if registry.RegistryMaxAttempts != policy.MaxAttempts ||
		registry.RegistryGetTotalDeadline != time.Duration(policy.GetTotalDeadlineSeconds)*time.Second ||
		registry.RegistryPostTotalDeadline != time.Duration(policy.PostTotalDeadlineSeconds)*time.Second {
		t.Fatalf("registry retry policy does not match vectors: %+v", policy)
	}

	getCalls := 0
	getServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		getCalls++
		response.Header().Set("Content-Type", "application/json")
		response.Header().Set("Retry-After", "0")
		response.WriteHeader(http.StatusServiceUnavailable)
		_, _ = response.Write([]byte(`{"error":{"code":"unavailable","message":"retry"}}`))
	}))
	defer getServer.Close()
	if _, err := registry.HTTPGetSnapshot(getServer.URL); err == nil {
		t.Fatal("repeated 503 response was accepted")
	}
	if getCalls != policy.MaxAttempts {
		t.Fatalf("GET attempts=%d, want %d", getCalls, policy.MaxAttempts)
	}

	targetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetCalls++
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL+"/v1/snapshot", http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	if _, err := registry.HTTPGetSnapshot(redirect.URL); err == nil {
		t.Fatal("registry redirect was accepted")
	}
	if policy.FollowRedirects || targetCalls != 0 {
		t.Fatalf("redirect target calls=%d, follow_redirects=%v", targetCalls, policy.FollowRedirects)
	}

	record, err := os.ReadFile(golden(t, "expected/registry/record_audited.json"))
	if err != nil {
		t.Fatal(err)
	}
	var postBodies [][]byte
	var idempotencyKeys []string
	postServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, readErr := io.ReadAll(request.Body)
		if readErr != nil {
			t.Error(readErr)
		}
		postBodies = append(postBodies, body)
		idempotencyKeys = append(idempotencyKeys, request.Header.Get("Idempotency-Key"))
		response.Header().Set("Content-Type", "application/json")
		response.Header().Set("Retry-After", "0")
		response.WriteHeader(http.StatusServiceUnavailable)
		_, _ = response.Write([]byte(`{"error":{"code":"unavailable","message":"retry"}}`))
	}))
	defer postServer.Close()
	if _, err := registry.Publish(postServer.URL, "secret-token", record); err == nil {
		t.Fatal("repeated POST 503 response was accepted")
	}
	if len(postBodies) != policy.MaxAttempts {
		t.Fatalf("POST attempts=%d, want %d", len(postBodies), policy.MaxAttempts)
	}
	for index := range postBodies {
		if !bytes.Equal(postBodies[index], record) || idempotencyKeys[index] == "" ||
			idempotencyKeys[index] != idempotencyKeys[0] {
			t.Fatalf("POST attempt %d changed body or idempotency identity", index+1)
		}
	}
}

func TestRegistryClientSnapshotTransitionVectors(t *testing.T) {
	vector := loadRegistryClientVector(t)
	if vector.StateKey != "canonical_registry_url" || vector.KeyRotationResetsState {
		t.Fatalf("unsupported snapshot state policy: key=%q reset=%v", vector.StateKey, vector.KeyRotationResetsState)
	}
	oldPublic, oldPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	newPublic, newPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pins := []string{
		"ed25519:" + base64.StdEncoding.EncodeToString(oldPublic),
		"ed25519:" + base64.StdEncoding.EncodeToString(newPublic),
	}
	createdAt := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	now := createdAt.Add(time.Hour)
	for _, testCase := range vector.SnapshotTransitions {
		t.Run(testCase.Name, func(t *testing.T) {
			cacheDir := t.TempDir()
			registryURL := "https://registry.example/curator"
			storedBody := interopSnapshotBody(testCase.StoredVersion, strings.Repeat("a", 64), createdAt)
			stored := signInteropSnapshot(oldPrivate, oldPublic, storedBody)
			initial := registry.Registry{Name: "before-rotation", URL: registryURL, PublicKeys: pins}
			tampered, warnings := registry.CheckSnapshots([]registry.Registry{initial}, cacheDir, func(string) (map[string]any, error) {
				return stored, nil
			}, now, 0)
			if len(tampered) != 0 || len(warnings) != 0 {
				t.Fatalf("could not establish stored state: tampered=%v warnings=%v", tampered, warnings)
			}

			candidateBody := interopSnapshotBody(testCase.CandidateVersion, strings.Repeat("a", 64), createdAt)
			if !testCase.SameBody && testCase.CandidateVersion == testCase.StoredVersion {
				candidateBody["head"] = strings.Repeat("c", 64)
			}
			candidate := signInteropSnapshot(newPrivate, newPublic, candidateBody)
			rotated := registry.Registry{Name: "after-rotation", URL: registryURL, PublicKeys: pins}
			tampered, _ = registry.CheckSnapshots([]registry.Registry{rotated}, cacheDir, func(string) (map[string]any, error) {
				return candidate, nil
			}, now, 0)
			accepted := !tampered[registryURL]
			if accepted != testCase.Accepted {
				t.Fatalf("accepted=%v, want %v", accepted, testCase.Accepted)
			}
		})
	}
}

func interopSnapshotBody(version int, head string, createdAt time.Time) map[string]any {
	return map[string]any{
		"schema_version": 1,
		"merkle_root":    strings.Repeat("b", 64),
		"log_size":       version,
		"head":           head,
		"version":        version,
		"created_at":     createdAt.Format(time.RFC3339),
	}
}

func signInteropSnapshot(private ed25519.PrivateKey, public ed25519.PublicKey, body map[string]any) map[string]any {
	signed := make(map[string]any, len(body)+1)
	for key, value := range body {
		signed[key] = value
	}
	signature := ed25519.Sign(private, registry.CanonicalBytes(signed))
	signed["sig"] = map[string]any{
		"key_id":    registry.KeyID(public),
		"algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	return signed
}

func TestRegistryClientPaginationRejectionVectors(t *testing.T) {
	vector := loadRegistryClientVector(t)
	for _, testCase := range vector.PaginationRejections {
		t.Run(testCase.Name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				calls++
				response.Header().Set("Content-Type", "application/json")
				switch testCase.Error {
				case "pagination_cycle":
					_ = json.NewEncoder(response).Encode(map[string]any{"records": []any{}, "next_cursor": "same"})
				case "invalid_cursor":
					_ = json.NewEncoder(response).Encode(map[string]any{"records": []any{}, "next_cursor": strings.Repeat("x", testCase.Characters)})
				case "record_limit":
					remaining := testCase.Records - (calls-1)*1000
					pageSize := 1000
					if remaining < pageSize {
						pageSize = remaining
					}
					page := make([]any, pageSize)
					for index := range page {
						page[index] = map[string]any{}
					}
					var next any
					if remaining > pageSize {
						next = fmt.Sprintf("page-%d", calls+1)
					}
					_ = json.NewEncoder(response).Encode(map[string]any{"records": page, "next_cursor": next})
				case "body_limit":
					_, _ = response.Write(make([]byte, testCase.Bytes))
				default:
					t.Fatalf("unknown pagination vector error %q", testCase.Error)
				}
			}))
			defer server.Close()
			fetch := registry.NewHTTPFetchWithPolicy(t.TempDir(), 0, 0, time.Now)
			_, err := fetch(server.URL, "git.example/skill", strings.Repeat("a", 40), "sha256:"+strings.Repeat("b", 64))
			if err == nil {
				t.Fatal("response was accepted, want rejection")
			}
			want := map[string]string{
				"pagination_cycle": "repeated pagination cursor",
				"invalid_cursor":   "invalid 'next_cursor'",
				"record_limit":     "more than 10000 records",
				"body_limit":       "exceeds 16777216 bytes",
			}[testCase.Error]
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not contain %q", err, want)
			}
		})
	}
}

func TestRegistryClientRollbackStateVectors(t *testing.T) {
	vector := loadRegistryClientVector(t)
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pin := "ed25519:" + base64.StdEncoding.EncodeToString(public)
	createdAt := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	snapshot := signInteropSnapshot(private, public, interopSnapshotBody(8, strings.Repeat("a", 64), createdAt))
	reg := registry.Registry{Name: "state-case", URL: "https://registry.example/state", PublicKeys: []string{pin}}
	for _, testCase := range vector.RollbackStateCases {
		t.Run(testCase.Name, func(t *testing.T) {
			root := t.TempDir()
			stateDir := filepath.Join(root, "state")
			if testCase.State == "unavailable" {
				if err := os.WriteFile(stateDir, []byte("not a directory"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if testCase.State == "malformed" || testCase.State == "deleted" {
				tampered, warnings := registry.CheckSnapshots([]registry.Registry{reg}, stateDir, func(string) (map[string]any, error) {
					return snapshot, nil
				}, createdAt.Add(time.Hour), 0)
				if len(tampered) != 0 || len(warnings) != 0 {
					t.Fatalf("could not establish rollback state: %v %v", tampered, warnings)
				}
				paths, err := filepath.Glob(filepath.Join(stateDir, "snapshot-*.json"))
				if err != nil || len(paths) != 1 {
					t.Fatalf("snapshot state paths=%v err=%v", paths, err)
				}
				if testCase.State == "malformed" {
					if err := os.WriteFile(paths[0], []byte("{broken"), 0o600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Remove(paths[0]); err != nil {
					t.Fatal(err)
				}
			}
			tampered, _ := registry.CheckSnapshots([]registry.Registry{reg}, stateDir, func(string) (map[string]any, error) {
				return snapshot, nil
			}, createdAt.Add(time.Hour), 0)
			accepted := !tampered[reg.URL]
			if accepted != testCase.Accepted {
				t.Fatalf("accepted=%v, want %v", accepted, testCase.Accepted)
			}
		})
	}
}
