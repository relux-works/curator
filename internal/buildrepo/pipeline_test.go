package buildrepo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

type recordingGo struct {
	events    *[]string
	compile   func(CompileRequest) error
	assurance closureexec.AssuranceBinding
}

func allowTestAssurance(context.Context) error { return nil }

func (g recordingGo) Identity() ToolchainIdentity {
	return ToolchainIdentity{ContentSHA256: "sha256:" + strings.Repeat("c", 64), GoVersion: "go version go1.26.1 test/arch", GoRelpath: "bin/go", GOOS: "test", GOARCH: "arch", Tuning: map[string]string{}}
}
func (g recordingGo) BuildInput(request CompileRequest) (buildmeta.Input, error) {
	token, err := buildsource.Validate(request.Root)
	if err != nil {
		return buildmeta.Input{}, err
	}
	defer func() { _ = token.Close() }()
	identity := g.Identity()
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: token.Identity(), BuildRoot: "build", Command: request.Command, SourceDir: request.SourceDir,
		Target:    buildmeta.Target{GOOS: identity.GOOS, GOARCH: identity.GOARCH, Tuning: identity.Tuning},
		Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, ContentSHA256: identity.ContentSHA256, GoVersion: identity.GoVersion, GoRelpath: identity.GoRelpath},
		Policy:    buildmeta.FixedPolicy(),
	}
	return input, input.Validate()
}
func (g recordingGo) Compile(_ context.Context, request CompileRequest) (CompileResult, error) {
	*g.events = append(*g.events, "compiler-call")
	if g.compile != nil {
		if err := g.compile(request); err != nil {
			return CompileResult{}, err
		}
	}
	artifact := []byte("deterministic-artifact")
	input, err := g.BuildInput(request)
	if err != nil {
		return CompileResult{}, err
	}
	metadata := artifactMetadata(map[string]any{"command": request.Command, "target": map[string]any{"goos": g.Identity().GOOS}}, artifact)
	binding := g.assurance
	if binding.AssuranceMode == "" {
		binding = closureexec.PortableAssuranceBinding()
	}
	var receipt closureexec.BuildSessionReceipt
	if binding.AssuranceMode == closureexec.AssuranceVerified {
		receipt, err = closureexec.NewVerifiedBuildSessionReceipt(binding, input, metadata, closuregraph.ID("sha256:"+strings.Repeat("e", 64)))
	} else {
		receipt, err = closureexec.NewPortableBuildSessionReceipt(input, metadata, nil)
	}
	return CompileResult{Artifact: artifact, ExecutionReceipt: receipt}, err
}

type recordingStore struct {
	inner  *DiskProtectedStore
	events *[]string
}

func (s recordingStore) LoadSnapshot(key string, mutate bool) (*Snapshot, error) {
	*s.events = append(*s.events, "snapshot-load")
	return s.inner.LoadSnapshot(key, mutate)
}
func (s recordingStore) StoreSnapshot(key string, snapshot *Snapshot) error {
	*s.events = append(*s.events, "snapshot-store")
	return s.inner.StoreSnapshot(key, snapshot)
}
func (s recordingStore) LookupArtifact(key string, input map[string]any, mutate bool) (*ArtifactHit, error) {
	*s.events = append(*s.events, "cache-call")
	return s.inner.LookupArtifact(key, input, mutate)
}
func (s recordingStore) StoreArtifact(key string, input map[string]any, command string, artifact, executionReceipt []byte) ([]byte, error) {
	*s.events = append(*s.events, "artifact-store")
	return s.inner.StoreArtifact(key, input, command, artifact, executionReceipt)
}

func pipelineFixture(t *testing.T) (*Snapshot, DeclaredState, EffectiveState) {
	t.Helper()
	files := []File{
		{Path: "outside-secret.txt", Content: []byte("must-not-be-visible")},
		{Path: "skill-build.json", Content: []byte(`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`)},
		{Path: "tools/cmd/tool/main.go", Content: []byte("package main\n")},
		{Path: "tools/go.mod", Content: []byte("module example.test/tool\n")},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	canonical := frameSnapshot(files)
	digest := sha256Digest(canonical)
	commit := strings.Repeat("1", 40)
	snapshot := &Snapshot{ObjectFormat: "sha1", Commit: commit, Files: files, CanonicalBytes: canonical, Digest: digest}
	declared := DeclaredState{Repository: "tools", Identity: "example.test/tools", Transport: "https", ObjectFormat: "sha1", Commit: commit}
	effective := EffectiveState{IdentityKind: "network-git", Identity: declared.Identity, Transport: declared.Transport, ObjectFormat: "sha1", Commit: commit}
	return snapshot, declared, effective
}

func sha256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestAuditWarningsRemainInPipelineResult(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	result, err := RunPipeline(context.Background(), PipelineRequest{
		Operation: OperationAudit, Command: "tool", Target: "tool", Declared: declared, Effective: effective,
		Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil },
		AuditWarnings: func(context.Context, AuditSubject) ([]string, error) {
			return []string{"advisory finding"}, nil
		},
	})
	if err != nil || len(result.Warnings) != 1 || result.Warnings[0] != "advisory finding" {
		t.Fatalf("audit warnings = %v, err=%v", result.Warnings, err)
	}
}

func TestReceiptArtifactPathMatchesTargetPlatform(t *testing.T) {
	for _, testCase := range []struct{ goos, want string }{
		{"linux", "bin/tool"},
		{"darwin", "bin/tool"},
		{"windows", "bin/tool.exe"},
	} {
		input := map[string]any{"command": "tool", "target": map[string]any{"goos": testCase.goos}}
		got, ok := artifactPathFromInput(input)
		if !ok || got != testCase.want {
			t.Errorf("%s artifact path = %q, %v; want %q", testCase.goos, got, ok, testCase.want)
		}
	}
}

func TestExternalPipelineOrderingAcrossOperations(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	for _, operation := range []Operation{OperationInstall, OperationDryRun, OperationRepair, OperationAudit} {
		t.Run(string(operation), func(t *testing.T) {
			events := []string{}
			store := recordingStore{inner: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, events: &events}
			request := PipelineRequest{Operation: operation, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: store, Go: recordingGo{events: &events}, Trace: func(phase string) { events = append(events, phase) }}
			request.Acquire = func(context.Context) (*Snapshot, error) {
				events = append(events, "acquire-call")
				return cloneSnapshot(snapshot), nil
			}
			request.Audit = func(_ context.Context, subject AuditSubject) error {
				events = append(events, "audit-call")
				if subject.BuildSource != snapshot.Digest || subject.Declared != declared || subject.Effective.Identity != effective.Identity {
					return errors.New("subject binding")
				}
				return nil
			}
			result, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			assertBefore(t, events, "whole-snapshot-validation", "audit-call")
			if operation == OperationAudit {
				if contains(events, "cache-call") || contains(events, "compiler-call") {
					t.Fatalf("audit-only crossed later boundary: %v", events)
				}
				return
			}
			assertBefore(t, events, "audit-call", "cache-call")
			if operation == OperationDryRun {
				if contains(events, "compiler-call") || contains(events, "snapshot-store") || contains(events, "artifact-store") {
					t.Fatalf("dry-run mutated or compiled: %v", events)
				}
				if result.State != "would-preflight-and-build" {
					t.Fatalf("state=%s", result.State)
				}
				return
			}
			assertBefore(t, events, "cache-call", "compiler-call")
			assertBefore(t, events, "compiler-call", "artifact-store")
		})
	}
}

func TestExternalPipelineAssuranceDriftPreventsCacheLookupAndCompile(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	binding := testVerifiedBinding("fixture.provider", 'a', 'b')
	assuranceChecks := 0
	request := PipelineRequest{
		Operation: OperationInstall, Assurance: binding,
		AssuranceCheck: func(context.Context) error {
			assuranceChecks++
			if assuranceChecks == 1 {
				return nil
			}
			return errors.New("verified_provider_unavailable: provider drifted")
		},
		Command: "tool", Target: "tool", Declared: declared, Effective: effective,
		Store:   recordingStore{inner: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, events: &events},
		Go:      recordingGo{events: &events, assurance: binding},
		Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil },
		Audit:   func(context.Context, AuditSubject) error { return nil },
	}
	if _, err := RunPipeline(t.Context(), request); err == nil || !strings.Contains(err.Error(), "verified_provider_unavailable") {
		t.Fatalf("assurance drift error = %v", err)
	}
	if assuranceChecks != 2 || !contains(events, "snapshot-store") {
		t.Fatalf("assurance checks=%d events=%v; want drift immediately before cache lookup", assuranceChecks, events)
	}
	for _, forbidden := range []string{"cache-call", "compiler-call", "artifact-store"} {
		if contains(events, forbidden) {
			t.Fatalf("assurance drift crossed %s: %v", forbidden, events)
		}
	}
}

func TestExternalPipelineCacheHitRepeatsAdmissionValidationAndAudit(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	store := recordingStore{inner: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, events: &events}
	compiles, audits, acquires := 0, 0, 0
	goSession := recordingGo{events: &events, compile: func(CompileRequest) error { compiles++; return nil }}
	run := func() PipelineResult {
		request := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: store, Go: goSession, Trace: func(p string) { events = append(events, p) }, Acquire: func(context.Context) (*Snapshot, error) { acquires++; return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { audits++; events = append(events, "audit-call"); return nil }}
		result, err := RunPipeline(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if first := run(); first.State != "would-preflight-and-build" {
		t.Fatalf("first=%s", first.State)
	}
	events = nil
	if second := run(); second.State != "cache-hit" {
		t.Fatalf("second=%s", second.State)
	}
	if acquires != 2 || audits != 2 || compiles != 1 {
		t.Fatalf("acquire=%d audit=%d compile=%d", acquires, audits, compiles)
	}
	assertBefore(t, events, "audit-call", "cache-call")
}

func TestExternalProtectedCacheSeparatesAssuranceModeProviderAndCapability(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	base := PipelineRequest{
		Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective,
		Store: store, Go: recordingGo{events: &events},
		Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil },
		Audit:   func(context.Context, AuditSubject) error { return nil },
	}
	portable, err := RunPipeline(t.Context(), base)
	if err != nil {
		t.Fatal(err)
	}
	verified := base
	verified.Assurance = testVerifiedBinding("fixture.provider", 'a', 'b')
	verified.Go = recordingGo{events: &events, assurance: verified.Assurance}
	verifiedResult, err := RunPipeline(t.Context(), verified)
	if err != nil {
		t.Fatal(err)
	}
	if verifiedResult.State == "cache-hit" || verifiedResult.CacheKey == portable.CacheKey {
		t.Fatalf("portable entry satisfied verified lookup: portable=%+v verified=%+v", portable, verifiedResult)
	}
	for name, change := range map[string]func(*closureexec.AssuranceBinding){
		"provider": func(value *closureexec.AssuranceBinding) {
			identity := *value.Provider
			identity.ProviderID = "other.provider"
			value.Provider = &identity
		},
		"capability": func(value *closureexec.AssuranceBinding) {
			id := closuregraph.ID("sha256:" + strings.Repeat("c", 64))
			value.CapabilityReceiptID = &id
		},
	} {
		t.Run(name, func(t *testing.T) {
			drifted := verified
			drifted.Operation = OperationDryRun
			drifted.Assurance = verified.Assurance
			change(&drifted.Assurance)
			drifted.Go = recordingGo{events: &events, assurance: drifted.Assurance}
			result, err := RunPipeline(t.Context(), drifted)
			if err != nil {
				t.Fatal(err)
			}
			if result.State == "cache-hit" || result.CacheKey == verifiedResult.CacheKey {
				t.Fatalf("%s drift adopted verified cache: %+v", name, result)
			}
		})
	}
}

func TestExternalProtectedCacheDoesNotAdoptLegacyAssuranceBlindEntry(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	goSession := recordingGo{events: &events}
	request := PipelineRequest{
		Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance,
		Command: "tool", Target: "tool", Declared: declared, Effective: effective,
		Store: store, Go: goSession,
		Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil },
		Audit:   func(context.Context, AuditSubject) error { return nil },
	}
	target := Target{BuildRoot: "tools", SourceDir: "tools/cmd/tool"}
	legacyInput := receiptInput(request, target, snapshot.Digest, goSession.Identity())
	delete(legacyInput, "assurance")
	legacyKey, err := cacheKey(legacyInput)
	if err != nil {
		t.Fatal(err)
	}
	legacyArtifact := []byte("legacy artifact")
	if _, err := store.StoreArtifact(legacyKey, legacyInput, "tool", legacyArtifact, testExecutionReceiptBytes(t, legacyInput, legacyArtifact)); err != nil {
		t.Fatal(err)
	}
	result, err := RunPipeline(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.State == "cache-hit" || result.CacheKey == legacyKey {
		t.Fatalf("legacy assurance-blind entry was adopted: %+v", result)
	}
}

func testVerifiedBinding(providerID string, providerSeed, capabilitySeed byte) closureexec.AssuranceBinding {
	contract := closureexec.VerifiedProviderContractID
	provider := closureexec.ProviderIdentity{
		Contract: contract, ProviderID: providerID, Version: "1.0.0",
		BinarySHA256: closuregraph.ID("sha256:" + strings.Repeat(string(providerSeed), 64)), TrustEvidence: "fixture-trust-v1",
	}
	capability := closuregraph.ID("sha256:" + strings.Repeat(string(capabilitySeed), 64))
	capabilities := []closureexec.CapabilityEvidence{
		{CapabilityID: "total-network-denial-v1", Status: "established"},
		{CapabilityID: "read-only-source-and-toolchain-v1", Status: "established"},
		{CapabilityID: "exact-executable-allowlisting-v1", Status: "established"},
		{CapabilityID: "private-build-root-only-writes-v1", Status: "established"},
		{CapabilityID: "hard-aggregate-descendant-resource-bounds-v1", Status: "established"},
		{CapabilityID: "fail-closed-capability-preflight-v1", Status: "established"},
	}
	return closureexec.AssuranceBinding{
		AssuranceMode: closureexec.AssuranceVerified, PolicyID: closureexec.VerifiedPolicyID,
		ExecutionPolicyID: closureexec.VerifiedExecutionPolicyID, ProviderContract: &contract,
		Provider: &provider, CapabilityReceiptID: &capability, ActualCapabilities: capabilities,
	}
}

func testExecutionReceiptBytes(t *testing.T, input map[string]any, artifact []byte) []byte {
	t.Helper()
	command, _ := input["command"].(string)
	target, _ := input["target"].(map[string]any)
	goos, _ := target["goos"].(string)
	buildInput := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Driver: buildmeta.DriverGoV1,
		BuildSource: buildsource.Identity{Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64)},
		BuildRoot:   "build", Command: command, SourceDir: "build",
		Target:    buildmeta.Target{GOOS: goos, GOARCH: "amd64", Tuning: map[string]string{"GOAMD64": "v1"}},
		Toolchain: buildmeta.Toolchain{Algorithm: buildmeta.ToolchainAlgorithm, ContentSHA256: "sha256:" + strings.Repeat("c", 64), GoVersion: "go version go1.26.1 test/arch", GoRelpath: "bin/go"},
		Policy:    buildmeta.FixedPolicy(),
	}
	receipt, err := closureexec.NewPortableBuildSessionReceipt(buildInput, artifactMetadata(input, artifact), nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestExternalPipelineOfflineProtectedSnapshotAndTagFailure(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	root := filepath.Join(t.TempDir(), "cache")
	disk := &DiskProtectedStore{Root: root}
	key, err := SnapshotKey(effective, snapshot.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if err = disk.StoreSnapshot(key, snapshot); err != nil {
		t.Fatal(err)
	}
	for _, tagged := range []bool{false, true} {
		t.Run(map[bool]string{false: "untagged-reuse", true: "tagged-blocked"}[tagged], func(t *testing.T) {
			d := declared
			if tagged {
				d.Tag = "v1.0.0"
			}
			events := []string{}
			request := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: d, Effective: effective, OfflineSnapshotKey: key, Store: recordingStore{inner: disk, events: &events}, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return nil, errors.New("offline") }, Audit: func(context.Context, AuditSubject) error { events = append(events, "audit-call"); return nil }, Trace: func(p string) { events = append(events, p) }}
			result, runErr := RunPipeline(context.Background(), request)
			if tagged {
				if ErrorCode(runErr) != CodeSourceUnavailable || contains(events, "audit-call") {
					t.Fatalf("tagged err=%v events=%v", runErr, events)
				}
				return
			}
			if runErr != nil {
				t.Fatal(runErr)
			}
			if result.State != "would-preflight-and-build" && result.State != "cache-hit" {
				t.Fatalf("state=%s", result.State)
			}
			assertBefore(t, events, "snapshot-load", "audit-call")
		})
	}
}

func TestExternalPipelineCompilerSeesOnlySelectedBuildRoot(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	session := recordingGo{events: &events, compile: func(request CompileRequest) error {
		if request.SourceDir != "build/cmd/tool" {
			return errors.New("wrong relative source dir")
		}
		if _, err := os.Stat(filepath.Join(request.Root, "outside-secret.txt")); !os.IsNotExist(err) {
			return errors.New("outside snapshot file is compiler-visible")
		}
		if _, err := os.Stat(filepath.Join(request.Root, "build", "go.mod")); err != nil {
			return err
		}
		return nil
	}}
	_, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, Go: session, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProtectedArtifactCorruptionQuarantinesAndRebuilds(t *testing.T) {
	for _, kind := range []string{"receipt", "execution-receipt", "artifact"} {
		t.Run(kind, func(t *testing.T) {
			snapshot, declared, effective := pipelineFixture(t)
			root := filepath.Join(t.TempDir(), "cache")
			store := &DiskProtectedStore{Root: root}
			events := []string{}
			compiles := 0
			request := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: store, Go: recordingGo{events: &events, compile: func(CompileRequest) error { compiles++; return nil }}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }}
			first, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			entry := filepath.Join(root, "artifacts", strings.TrimPrefix(first.CacheKey, "sha256:"))
			name := "artifact"
			switch kind {
			case "receipt":
				name = "receipt.json"
			case "execution-receipt":
				name = "execution-receipt.ccj.json"
			}
			if err = os.WriteFile(filepath.Join(entry, name), []byte("corrupt"), 0o600); err != nil {
				t.Fatal(err)
			}
			second, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if second.State != "would-rebuild-untrusted-cache" || compiles != 2 {
				t.Fatalf("state=%s compiles=%d", second.State, compiles)
			}
			entries, err := os.ReadDir(filepath.Join(root, "quarantine"))
			if err != nil || len(entries) != 1 {
				t.Fatalf("quarantine=%v err=%v", entries, err)
			}
		})
	}
}

func TestProtectedSnapshotCorruptionQuarantinesBeforeAudit(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	root := filepath.Join(t.TempDir(), "cache")
	store := &DiskProtectedStore{Root: root}
	key, _ := SnapshotKey(effective, snapshot.Digest)
	if err := store.StoreSnapshot(key, snapshot); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(root, "snapshots", strings.TrimPrefix(key, "sha256:"), "files", "tools", "go.mod")
	if err := os.WriteFile(entry, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	audited := false
	_, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, OfflineSnapshotKey: key, Store: store, Go: recordingGo{events: &[]string{}}, Acquire: func(context.Context) (*Snapshot, error) { return nil, errors.New("offline") }, Audit: func(context.Context, AuditSubject) error { audited = true; return nil }})
	if ErrorCode(err) != CodeSourceUnavailable || audited {
		t.Fatalf("err=%v audited=%v", err, audited)
	}
	entries, qerr := os.ReadDir(filepath.Join(root, "quarantine"))
	if qerr != nil || len(entries) != 1 {
		t.Fatalf("quarantine=%v err=%v", entries, qerr)
	}
}

func TestSubstitutionCannotAliasDeclaredCacheKey(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	events := []string{}
	base := PipelineRequest{Operation: OperationDryRun, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }}
	plain, err := RunPipeline(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}
	sub := base
	sub.Effective = effective
	sub.Effective.Substituted = true
	sub.Effective.IdentityKind = "operator-local-git"
	sub.Effective.Identity = "sha256:" + strings.Repeat("d", 64)
	sub.Effective.Transport = ""
	sub.Effective.Substitution = &SubstitutionState{Type: "local-path"}
	changed, err := RunPipeline(context.Background(), sub)
	if err != nil {
		t.Fatal(err)
	}
	if plain.CacheKey == changed.CacheKey || plain.SnapshotKey == changed.SnapshotKey {
		t.Fatal("substitution aliased declared/effective keys")
	}
}

func TestExternalReceiptV2CacheKeyVector(t *testing.T) {
	request := PipelineRequest{Assurance: closureexec.PortableAssuranceBinding(), Command: "golden-tool", Target: "golden-tool", Declared: DeclaredState{Repository: "golden-tools", Identity: "github.com/example/golden-tools", Transport: "https", ObjectFormat: "sha1", Commit: "0123456789abcdef0123456789abcdef01234567", Tag: "v1.4.0"}, Effective: EffectiveState{IdentityKind: "network-git", Identity: "github.com/example/golden-tools", Transport: "https", ObjectFormat: "sha1", Commit: "0123456789abcdef0123456789abcdef01234567"}}
	target := Target{BuildRoot: ".", SourceDir: "cmd/golden-tool"}
	tool := ToolchainIdentity{ContentSHA256: "sha256:" + strings.Repeat("c", 64), GoVersion: "go version go1.26.1 darwin/arm64", GoRelpath: "bin/go", GOOS: "darwin", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}}
	key, err := cacheKey(receiptInput(request, target, "sha256:"+strings.Repeat("b", 64), tool))
	if err != nil {
		t.Fatal(err)
	}
	if key != "sha256:6ca1b6b00f0ab343901daa1c90ee78ed417b3a258efeff706b41210dd702dbf2" {
		t.Fatalf("cache key=%s", key)
	}
}

func TestDryRunCorruptionIsReportedWithoutMutation(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	root := filepath.Join(t.TempDir(), "cache")
	store := &DiskProtectedStore{Root: root}
	events := []string{}
	base := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: store, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }}
	first, err := RunPipeline(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(root, "artifacts", strings.TrimPrefix(first.CacheKey, "sha256:"), "receipt.json")
	if err = os.WriteFile(entry, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	base.Operation = OperationDryRun
	result, err := RunPipeline(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "corrupt" || result.Code != CodeReceiptInvalid {
		t.Fatalf("result=%+v", result)
	}
	if _, err = os.Stat(filepath.Join(root, "quarantine")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created quarantine: %v", err)
	}
}

func TestDryRunDoesNotCreateProtectedRoot(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	parent := t.TempDir()
	root := filepath.Join(parent, "absent-cache")
	events := []string{}
	result, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationDryRun, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: &DiskProtectedStore{Root: root}, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "would-preflight-and-build" {
		t.Fatalf("state=%s", result.State)
	}
	if _, err = os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("dry-run created protected state: %v", err)
	}
}

func TestPipelineFailuresStopBeforeCacheAndCompiler(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	for _, testCase := range []struct {
		name  string
		alter func(*PipelineRequest, *Snapshot)
		code  string
	}{
		{name: "source", code: CodeSourceUnavailable, alter: func(request *PipelineRequest, _ *Snapshot) {
			request.Acquire = func(context.Context) (*Snapshot, error) { return nil, errors.New("offline") }
		}},
		{name: "descriptor", code: CodeDescriptorInvalid, alter: func(_ *PipelineRequest, value *Snapshot) {
			var kept []File
			for _, file := range value.Files {
				if file.Path != DescriptorName {
					kept = append(kept, file)
				}
			}
			value.Files = kept
			value.CanonicalBytes = frameSnapshot(value.Files)
			value.Digest = sha256Digest(value.CanonicalBytes)
		}},
		{name: "audit", code: CodeAuditBlocked, alter: func(request *PipelineRequest, _ *Snapshot) {
			request.Audit = func(context.Context, AuditSubject) error { return errors.New("blocked") }
		}},
		{name: "snapshot-race", code: CodeObjectSemanticsInvalid, alter: func(request *PipelineRequest, _ *Snapshot) {
			request.Audit = func(_ context.Context, subject AuditSubject) error {
				return os.WriteFile(filepath.Join(subject.SnapshotRoot, "tools", "go.mod"), []byte("changed"), 0o600)
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			events := []string{}
			value := cloneSnapshot(snapshot)
			request := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: recordingStore{inner: &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}, events: &events}, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return value, nil }, Audit: func(context.Context, AuditSubject) error { return nil }}
			testCase.alter(&request, value)
			_, err := RunPipeline(context.Background(), request)
			if ErrorCode(err) != testCase.code {
				t.Fatalf("error=%v, want %s", err, testCase.code)
			}
			if contains(events, "cache-call") || contains(events, "compiler-call") || contains(events, "artifact-store") {
				t.Fatalf("downstream work started: %v", events)
			}
		})
	}
}

func TestSigningPoliciesFailClosedBeforeSourceAndPackageInputs(t *testing.T) {
	for _, testCase := range []struct {
		name string
		req  PipelineRequest
		code string
	}{
		{name: "package", req: PipelineRequest{PackageSigningRequested: true}, code: CodePackageSigningForbidden},
		{name: "operator profile absent", req: PipelineRequest{SigningPolicy: "platform-required"}, code: CodeSignerPolicyUnsupported},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			acquired := false
			testCase.req.Acquire = func(context.Context) (*Snapshot, error) {
				acquired = true
				return nil, nil
			}
			_, err := RunPipeline(context.Background(), testCase.req)
			if ErrorCode(err) != testCase.code || acquired {
				t.Fatalf("code=%q acquired=%v err=%v", ErrorCode(err), acquired, err)
			}
		})
	}
}

func TestExternalGCUsesMarkerAndJournalKeysAsOnlyRoots(t *testing.T) {
	snapshot, declared, effective := pipelineFixture(t)
	root := filepath.Join(t.TempDir(), "cache")
	store := &DiskProtectedStore{Root: root}
	events := []string{}
	build := func(command string) PipelineResult {
		result, err := RunPipeline(context.Background(), PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: command, Target: "tool", Declared: declared, Effective: effective, Store: store, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	markerRoot, journalRoot := build("marker-tool"), build("journal-tool")
	removed, err := Collect(root, []string{markerRoot.CacheKey, journalRoot.CacheKey}, time.Now().Add(time.Hour), 0)
	if err != nil || len(removed) != 0 {
		t.Fatalf("rooted removed=%v err=%v", removed, err)
	}
	removed, err = Collect(root, []string{journalRoot.CacheKey}, time.Now().Add(time.Hour), 0)
	if err != nil || !contains(removed, "external:"+markerRoot.CacheKey) {
		t.Fatalf("marker release removed=%v err=%v", removed, err)
	}
	removed, err = Collect(root, nil, time.Now().Add(time.Hour), 0)
	if err != nil || !contains(removed, "external:"+journalRoot.CacheKey) || !contains(removed, "external-snapshot:"+journalRoot.SnapshotKey) {
		t.Fatalf("final removed=%v err=%v", removed, err)
	}
}

func TestProtectedArtifactHardLinkAndBoundaryFailClosed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("hard-link count is enforced by the native Windows protected-root backend")
	}
	snapshot, declared, effective := pipelineFixture(t)
	root := filepath.Join(t.TempDir(), "cache")
	store := &DiskProtectedStore{Root: root}
	events := []string{}
	request := PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance, Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: store, Go: recordingGo{events: &events}, Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil }, Audit: func(context.Context, AuditSubject) error { return nil }}
	first, err := RunPipeline(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(root, "artifacts", strings.TrimPrefix(first.CacheKey, "sha256:"))
	if err := os.Link(filepath.Join(entry, "artifact"), filepath.Join(entry, "artifact-link")); err != nil {
		t.Fatal(err)
	}
	second, err := RunPipeline(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if second.State != "would-rebuild-untrusted-cache" || second.Code != CodeArtifactInvalid {
		t.Fatalf("result=%+v", second)
	}
	if err := os.Chmod(root, 0o777); err != nil {
		t.Fatal(err)
	}
	_, err = store.LookupArtifact(first.CacheKey, map[string]any{}, true)
	if ErrorCode(err) != CodeProtectedBoundaryUntrusted {
		t.Fatalf("boundary error=%v", err)
	}
}

func TestProtectedStoreCannotReturnCacheHitWithoutIdentityProof(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	store := &DiskProtectedStore{
		Root: root,
		identityProof: func(os.FileInfo, bool) bool {
			return false
		},
	}
	hit, err := store.LookupArtifact("sha256:"+strings.Repeat("a", 64), map[string]any{}, false)
	if hit != nil || ErrorCode(err) != CodeProtectedBoundaryUntrusted {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
}

func cloneSnapshot(source *Snapshot) *Snapshot {
	copyValue := *source
	copyValue.CanonicalBytes = append([]byte(nil), source.CanonicalBytes...)
	copyValue.Files = make([]File, len(source.Files))
	for i, file := range source.Files {
		copyValue.Files[i] = file
		copyValue.Files[i].Content = append([]byte(nil), file.Content...)
	}
	return &copyValue
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func assertBefore(t *testing.T, values []string, left, right string) {
	t.Helper()
	li, ri := -1, -1
	for i, v := range values {
		if v == left && li < 0 {
			li = i
		}
		if v == right && ri < 0 {
			ri = i
		}
	}
	if li < 0 || ri < 0 || li >= ri {
		t.Fatalf("want %s before %s: %v", left, right, values)
	}
}
