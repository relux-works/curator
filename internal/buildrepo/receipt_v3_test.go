package buildrepo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/registry"
)

func testPackage() *buildmeta.Package {
	return &buildmeta.Package{Kind: buildmeta.PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("ab", 32)}
}

// packageBlindGo is a compiler session that binds the legacy context-only
// build input digest into the execution receipt even when asked for the
// receipt-3 wrapper: the negative row for "an implementation unable to bind
// that input MUST reject execution".
type packageBlindGo struct{ recordingGo }

func (g packageBlindGo) BuildInput(request CompileRequest) (buildmeta.Input, error) {
	request.Package = nil
	request.ExpectedDigest = ""
	return g.recordingGo.BuildInput(request)
}

func (g packageBlindGo) Compile(ctx context.Context, request CompileRequest) (CompileResult, error) {
	request.Package = nil
	request.ExpectedDigest = ""
	return g.recordingGo.Compile(ctx, request)
}

// compilerViewGo keeps the requested package but binds only the compiler
// go-v1 view digest instead of the exact receipt-3 wrapper digest: the
// negative row for "must not claim compatibility based on a context-only
// hash" when the package matches but the full external input does not.
type compilerViewGo struct{ recordingGo }

func (g compilerViewGo) BuildInput(request CompileRequest) (buildmeta.Input, error) {
	return g.recordingGo.BuildInput(request)
}

func (g compilerViewGo) Compile(ctx context.Context, request CompileRequest) (CompileResult, error) {
	request.ExpectedDigest = ""
	return g.recordingGo.Compile(ctx, request)
}

func sourceAwareRequest(t *testing.T, root string, events *[]string, goSession GoSession) (PipelineRequest, *Snapshot) {
	t.Helper()
	snapshot, declared, effective := pipelineFixture(t)
	if goSession == nil {
		goSession = recordingGo{events: events}
	}
	return PipelineRequest{Operation: OperationInstall, Assurance: closureexec.PortableAssuranceBinding(), AssuranceCheck: allowTestAssurance,
		Command: "tool", Target: "tool", Declared: declared, Effective: effective, Store: &DiskProtectedStore{Root: root}, Go: goSession,
		Acquire: func(context.Context) (*Snapshot, error) { return cloneSnapshot(snapshot), nil },
		Audit:   func(context.Context, AuditSubject) error { return nil }, Package: testPackage()}, snapshot
}

func decodeReceiptObject(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	if err := protocoljson.Validate(payload); err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	return object
}

// TestExternalReceipt3WrapsTheReceipt2InputOnTheExternalArm is the positive
// row of the external arm: the pipeline publishes under artifacts-receipt-3
// with input:{schema_version:3,package,build}, the build is the byte-identical
// receipt-2 driver input, the cache key is SHA-256 over the CCJ-1 wrapper, and
// the same request is a cache hit while a package-less request is not.
func TestExternalReceipt3WrapsTheReceipt2InputOnTheExternalArm(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	events := []string{}
	request, snapshot := sourceAwareRequest(t, root, &events, nil)
	result, err := RunPipeline(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != "would-preflight-and-build" || result.ReceiptSchemaVersion != SourceAwareReceiptSchemaVersion {
		t.Fatalf("result = %+v", result)
	}
	entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(result.CacheKey, "sha256:"))
	receipt, err := os.ReadFile(filepath.Join(entry, "receipt.json"))
	if err != nil {
		t.Fatalf("receipt-3 entry missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ArtifactsDir(LegacyReceiptSchemaVersion))); !os.IsNotExist(err) {
		t.Fatalf("legacy artifacts namespace was created: %v", err)
	}
	if !bytes.Equal(receipt, result.Receipt) {
		t.Fatal("result receipt differs from the protected receipt")
	}
	object := decodeReceiptObject(t, receipt)
	if len(object) != 4 || object["schema_version"] != json.Number("3") || object["cache_key"] != result.CacheKey {
		t.Fatalf("receipt shape = %v", object)
	}
	input := object["input"].(map[string]any)
	if len(input) != 3 || input["schema_version"] != json.Number("3") {
		t.Fatalf("wrapper shape = %v", input)
	}
	wantPackage, _ := registry.CanonicalBytesChecked(testPackage().Object())
	gotPackage, _ := registry.CanonicalBytesChecked(input["package"].(map[string]any))
	if !bytes.Equal(wantPackage, gotPackage) {
		t.Fatalf("package = %s, want %s", gotPackage, wantPackage)
	}
	legacyInput := legacyReceiptInput(request, Target{BuildRoot: "tools", SourceDir: "tools/cmd/tool"}, snapshot.Digest, request.Go.Identity())
	wantBuild, _ := registry.CanonicalBytesChecked(legacyInput)
	gotBuild, _ := registry.CanonicalBytesChecked(input["build"].(map[string]any))
	if !bytes.Equal(wantBuild, gotBuild) {
		t.Fatalf("wrapped build is not the unchanged receipt-2 input:\n%s\n%s", gotBuild, wantBuild)
	}
	wrapper, _ := registry.CanonicalBytesChecked(input)
	sum := sha256.Sum256(wrapper)
	if result.CacheKey != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatalf("cache key %s is not SHA-256 of the wrapped input", result.CacheKey)
	}
	legacyKey, _ := cacheKey(legacyInput)
	if legacyKey == result.CacheKey {
		t.Fatal("receipt-3 key aliases the receipt-2 key")
	}
	if !bytes.Contains(receipt, []byte(`"transport":"https"`)) || !bytes.Contains(receipt, []byte(`"locked_commit"`)) || !bytes.Contains(receipt, []byte(`"assurance"`)) {
		t.Fatalf("receipt lost declared/effective fields: %s", receipt)
	}

	events = events[:0]
	again, err := RunPipeline(context.Background(), request)
	if err != nil || again.State != "cache-hit" || again.CacheKey != result.CacheKey || again.ReceiptSchemaVersion != SourceAwareReceiptSchemaVersion {
		t.Fatalf("second run = %+v, %v", again, err)
	}
	if strings.Contains(strings.Join(events, ","), "compiler-call") {
		t.Fatalf("cache hit compiled again: %v", events)
	}

	legacyRequest := request
	legacyRequest.Package = nil
	legacy, err := RunPipeline(context.Background(), legacyRequest)
	if err != nil || legacy.State != "would-preflight-and-build" || legacy.ReceiptSchemaVersion != LegacyReceiptSchemaVersion || legacy.CacheKey != legacyKey {
		t.Fatalf("legacy request = %+v, %v (no receipt-3 hit may satisfy it)", legacy, err)
	}
	if _, err := os.Stat(filepath.Join(root, ArtifactsDir(LegacyReceiptSchemaVersion), strings.TrimPrefix(legacyKey, "sha256:"), "receipt.json")); err != nil {
		t.Fatalf("legacy entry not published under the legacy namespace: %v", err)
	}
	otherPackage := request
	otherPackage.Package = &buildmeta.Package{Kind: buildmeta.PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("cd", 32)}
	other, err := RunPipeline(context.Background(), otherPackage)
	if err != nil || other.State != "would-preflight-and-build" || other.CacheKey == result.CacheKey {
		t.Fatalf("package mutation reused the entry: %+v, %v", other, err)
	}
}

// TestExternalReceipt3RefusesEveryEvidenceFieldMismatch mutates every
// external-evidence field of a protected receipt-3 entry on disk — package,
// declared identity and locked commit, effective identity/commit/transport,
// substitution, build source, descriptor target, target, toolchain, policy,
// assurance, both schema versions and the cache key — and requires the next
// lookup to refuse the entry and rebuild instead of adopting it.
func TestExternalReceipt3RefusesEveryEvidenceFieldMismatch(t *testing.T) {
	rows := map[string]func(input map[string]any){
		"package snapshot": func(i map[string]any) {
			i["package"].(map[string]any)["snapshot"] = "sha256:" + strings.Repeat("cd", 32)
		},
		"package kind": func(i map[string]any) {
			i["package"] = map[string]any{"kind": "network-git", "repository": "git.example.com/a/b", "commit": map[string]any{"object_format": "sha1", "hex": strings.Repeat("1", 40)}, "directory": "."}
		},
		"package extra field": func(i map[string]any) { i["package"].(map[string]any)["ref"] = "v1" },
		"package removed":     func(i map[string]any) { delete(i, "package") },
		"wrapper schema 2":    func(i map[string]any) { i["schema_version"] = json.Number("2") },
		"build schema 3":      func(i map[string]any) { build(i)["schema_version"] = json.Number("3") },
		"declared identity": func(i map[string]any) {
			source(i)["declared"].(map[string]any)["identity"].(map[string]any)["value"] = "example.test/other"
		},
		"declared identity kind": func(i map[string]any) {
			source(i)["declared"].(map[string]any)["identity"].(map[string]any)["kind"] = "operator-local-git"
		},
		"declared transport": func(i map[string]any) { source(i)["declared"].(map[string]any)["transport"] = "ssh" },
		"declared locked commit hex": func(i map[string]any) {
			source(i)["declared"].(map[string]any)["locked_commit"].(map[string]any)["hex"] = strings.Repeat("2", 40)
		},
		"declared locked commit format": func(i map[string]any) {
			source(i)["declared"].(map[string]any)["locked_commit"].(map[string]any)["object_format"] = "sha256"
		},
		"declared tag added": func(i map[string]any) { source(i)["declared"].(map[string]any)["tag"] = "v1" },
		"effective identity": func(i map[string]any) {
			source(i)["effective"].(map[string]any)["identity"].(map[string]any)["value"] = "example.test/other"
		},
		"effective identity kind": func(i map[string]any) {
			source(i)["effective"].(map[string]any)["identity"].(map[string]any)["kind"] = "operator-local-git"
		},
		"effective commit":        func(i map[string]any) { source(i)["effective"].(map[string]any)["commit"] = strings.Repeat("2", 40) },
		"effective object format": func(i map[string]any) { source(i)["effective"].(map[string]any)["object_format"] = "sha256" },
		"effective transport":     func(i map[string]any) { source(i)["effective"].(map[string]any)["transport"] = "ssh" },
		"effective substituted":   func(i map[string]any) { source(i)["effective"].(map[string]any)["substituted"] = true },
		"effective substitution added": func(i map[string]any) {
			source(i)["effective"].(map[string]any)["substitution"] = map[string]any{"type": "local-path"}
		},
		"effective build source digest": func(i map[string]any) {
			source(i)["effective"].(map[string]any)["build_source"].(map[string]any)["content_sha256"] = "sha256:" + strings.Repeat("9", 64)
		},
		"effective build source algorithm": func(i map[string]any) {
			source(i)["effective"].(map[string]any)["build_source"].(map[string]any)["algorithm"] = "other"
		},
		"repository":        func(i map[string]any) { source(i)["repository"] = "other" },
		"descriptor target": func(i map[string]any) { source(i)["descriptor"].(map[string]any)["target"] = "other" },
		"descriptor path":   func(i map[string]any) { source(i)["descriptor"].(map[string]any)["path"] = "other.json" },
		"command":           func(i map[string]any) { build(i)["command"] = "other" },
		"build root":        func(i map[string]any) { build(i)["build_root"] = "other" },
		"source dir":        func(i map[string]any) { build(i)["source_dir"] = "tools/cmd/other" },
		"driver":            func(i map[string]any) { build(i)["driver"] = "go-v1" },
		"target goarch":     func(i map[string]any) { build(i)["target"].(map[string]any)["goarch"] = "other" },
		"target tuning": func(i map[string]any) {
			build(i)["target"].(map[string]any)["tuning"] = map[string]any{"GOAMD64": "v4"}
		},
		"toolchain digest": func(i map[string]any) {
			build(i)["toolchain"].(map[string]any)["content_sha256"] = "sha256:" + strings.Repeat("9", 64)
		},
		"toolchain version":  func(i map[string]any) { build(i)["toolchain"].(map[string]any)["go_version"] = "other" },
		"policy execution":   func(i map[string]any) { build(i)["policy"].(map[string]any)["execution_policy"] = "hardened-worker-v1" },
		"policy source kind": func(i map[string]any) { build(i)["policy"].(map[string]any)["source_kind"] = "other" },
		"policy network":     func(i map[string]any) { build(i)["policy"].(map[string]any)["network"] = "full" },
		"assurance mode":     func(i map[string]any) { build(i)["assurance"].(map[string]any)["assurance_mode"] = "verified" },
		"assurance removed":  func(i map[string]any) { delete(build(i), "assurance") },
	}
	for name, mutate := range rows {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "cache")
			events := []string{}
			compiles := 0
			request, _ := sourceAwareRequest(t, root, &events, recordingGo{events: &events, compile: func(CompileRequest) error { compiles++; return nil }})
			first, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(first.CacheKey, "sha256:"))
			receiptPath := filepath.Join(entry, "receipt.json")
			object := decodeReceiptObject(t, first.Receipt)
			mutate(object["input"].(map[string]any))
			forged, err := registry.CanonicalBytesChecked(object)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(forged, first.Receipt) {
				t.Fatal("mutation did not change the receipt")
			}
			if err := os.WriteFile(receiptPath, forged, 0o600); err != nil {
				t.Fatal(err)
			}
			hit, lookupErr := request.Store.LookupArtifact(first.CacheKey, mustSourceAwareInput(t, request, first), false)
			if hit != nil || lookupErr == nil || ErrorCode(lookupErr) != CodeReceiptInvalid {
				t.Fatalf("forged receipt adopted: hit=%v err=%v", hit != nil, lookupErr)
			}
			second, err := RunPipeline(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if second.State != "would-rebuild-untrusted-cache" || compiles != 2 || !bytes.Equal(second.Receipt, first.Receipt) {
				t.Fatalf("state=%s compiles=%d", second.State, compiles)
			}
		})
	}
	t.Run("receipt schema version", func(t *testing.T) {
		// Only receipt_schema_version changes to 3: a record that keeps the
		// exact wrapper input and key but claims schema 2 is refused as a
		// closed-shape mismatch, never adopted as either version.
		root := filepath.Join(t.TempDir(), "cache")
		events := []string{}
		request, _ := sourceAwareRequest(t, root, &events, nil)
		first, err := RunPipeline(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		object := decodeReceiptObject(t, first.Receipt)
		object["schema_version"] = json.Number("2")
		forged, _ := registry.CanonicalBytesChecked(object)
		entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(first.CacheKey, "sha256:"))
		if err := os.WriteFile(filepath.Join(entry, "receipt.json"), forged, 0o600); err != nil {
			t.Fatal(err)
		}
		if hit, err := request.Store.LookupArtifact(first.CacheKey, mustSourceAwareInput(t, request, first), false); hit != nil || err == nil || ErrorCode(err) != CodeReceiptInvalid {
			t.Fatalf("re-labelled receipt adopted: hit=%v err=%v", hit != nil, err)
		}
	})
	t.Run("cache key", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "cache")
		events := []string{}
		request, _ := sourceAwareRequest(t, root, &events, nil)
		first, err := RunPipeline(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		object := decodeReceiptObject(t, first.Receipt)
		object["cache_key"] = "sha256:" + strings.Repeat("0", 64)
		forged, _ := registry.CanonicalBytesChecked(object)
		entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(first.CacheKey, "sha256:"))
		if err := os.WriteFile(filepath.Join(entry, "receipt.json"), forged, 0o600); err != nil {
			t.Fatal(err)
		}
		if hit, err := request.Store.LookupArtifact(first.CacheKey, mustSourceAwareInput(t, request, first), false); hit != nil || err == nil {
			t.Fatalf("forged cache key adopted: %v", err)
		}
	})
}

func build(input map[string]any) map[string]any  { return input["build"].(map[string]any) }
func source(input map[string]any) map[string]any { return build(input)["source"].(map[string]any) }

func mustSourceAwareInput(t *testing.T, request PipelineRequest, result PipelineResult) map[string]any {
	t.Helper()
	object := decodeReceiptObject(t, result.Receipt)
	input := object["input"].(map[string]any)
	if ReceiptSchemaVersionOf(input) != SourceAwareReceiptSchemaVersion {
		t.Fatalf("published input is not receipt-3: %v", input)
	}
	_ = request
	return input
}

// TestExternalReceipt3EntriesNeverCrossNamespaces: a receipt-2 record planted
// at the receipt-3 key inside the receipt-3 namespace, and a receipt-3 record
// planted in the legacy namespace, are both refused; no legacy cache hit
// satisfies a source-aware lookup.
func TestExternalReceipt3EntriesNeverCrossNamespaces(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	events := []string{}
	request, _ := sourceAwareRequest(t, root, &events, nil)
	sourceAware, err := RunPipeline(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	legacyRequest := request
	legacyRequest.Package = nil
	legacy, err := RunPipeline(context.Background(), legacyRequest)
	if err != nil {
		t.Fatal(err)
	}
	sourceAwareEntry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(sourceAware.CacheKey, "sha256:"))
	legacyEntry := filepath.Join(root, ArtifactsDir(LegacyReceiptSchemaVersion), strings.TrimPrefix(legacy.CacheKey, "sha256:"))
	// Legacy receipt under the receipt-3 key: closed-shape refusal.
	if err := os.WriteFile(filepath.Join(sourceAwareEntry, "receipt.json"), legacy.Receipt, 0o600); err != nil {
		t.Fatal(err)
	}
	if hit, err := request.Store.LookupArtifact(sourceAware.CacheKey, mustSourceAwareInput(t, request, sourceAware), false); hit != nil || err == nil {
		t.Fatalf("legacy receipt adopted at the receipt-3 key: %v", err)
	}
	// Receipt-3 record under the legacy key: refused for the legacy input.
	if err := os.WriteFile(filepath.Join(legacyEntry, "receipt.json"), sourceAware.Receipt, 0o600); err != nil {
		t.Fatal(err)
	}
	legacyObject := decodeReceiptObject(t, legacy.Receipt)
	if hit, err := request.Store.LookupArtifact(legacy.CacheKey, legacyObject["input"].(map[string]any), false); hit != nil || err == nil {
		t.Fatalf("receipt-3 record adopted at the legacy key: %v", err)
	}
}

// TestExternalReceipt3ExecutionBindsExactWrapperDigest: the fresh publication
// binds sha256(CCJ-1(receipt.input)) == cache_key in the execution receipt's
// build_input_sha256, and a cache reuse is an exact hit on the same digest.
// A session that binds only the compiler go-v1 view digest (package matches)
// is refused before publication, even though its package is correct.
func TestExternalReceipt3ExecutionBindsExactWrapperDigest(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	events := []string{}
	request, _ := sourceAwareRequest(t, root, &events, nil)
	first, err := RunPipeline(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.State != "would-preflight-and-build" {
		t.Fatalf("first state = %s", first.State)
	}
	entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(first.CacheKey, "sha256:"))
	receiptPayload, err := os.ReadFile(filepath.Join(entry, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	executionPayload, err := os.ReadFile(filepath.Join(entry, "execution-receipt.ccj.json"))
	if err != nil {
		t.Fatal(err)
	}
	receiptObject := decodeReceiptObject(t, receiptPayload)
	executionObject := decodeReceiptObject(t, executionPayload)
	inputBytes, err := registry.CanonicalBytesChecked(receiptObject["input"].(map[string]any))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(inputBytes)
	want := "sha256:" + hex.EncodeToString(sum[:])
	if receiptObject["cache_key"] != want || first.CacheKey != want {
		t.Fatalf("cache key %s != wrapper digest %s", first.CacheKey, want)
	}
	if executionObject["build_input_sha256"] != want {
		t.Fatalf("execution binds %v, want exact wrapper digest %s", executionObject["build_input_sha256"], want)
	}
	// Cache reuse is an exact hit on the same digest without recompiling.
	events = events[:0]
	second, err := RunPipeline(context.Background(), request)
	if err != nil || second.State != "cache-hit" || second.CacheKey != want {
		t.Fatalf("reuse = %+v, %v", second, err)
	}
	if strings.Contains(strings.Join(events, ","), "compiler-call") {
		t.Fatalf("reuse compiled again: %v", events)
	}
	if string(second.ExecutionReceipt.BuildInputSHA256) != want {
		t.Fatalf("reused execution binds %s, want %s", second.ExecutionReceipt.BuildInputSHA256, want)
	}
	// A fresh store with a compiler-view session (package matches, digest is
	// the go-v1 view) is refused before publication.
	freshRoot := filepath.Join(t.TempDir(), "cache")
	freshEvents := []string{}
	fresh, _ := sourceAwareRequest(t, freshRoot, &freshEvents, compilerViewGo{recordingGo{events: &freshEvents}})
	if _, err := RunPipeline(context.Background(), fresh); err == nil || ErrorCode(err) != CodeReceiptInvalid {
		t.Fatalf("compiler-view session accepted: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(freshRoot, ArtifactsDir(SourceAwareReceiptSchemaVersion))); len(entries) != 0 {
		t.Fatalf("refused compiler-view compilation published %d entries", len(entries))
	}
}

// TestExternalReceipt3RequiresTheSessionToBindTheWrappedInput: a compiler
// session whose execution receipt binds the context-only legacy digest is
// refused at compile and never published; a protected entry whose execution
// receipt binds the legacy digest is refused at lookup.
func TestExternalReceipt3RequiresTheSessionToBindTheWrappedInput(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	events := []string{}
	request, _ := sourceAwareRequest(t, root, &events, packageBlindGo{recordingGo{events: &events}})
	_, err := RunPipeline(context.Background(), request)
	if err == nil || ErrorCode(err) != CodeReceiptInvalid {
		t.Fatalf("package-blind session accepted: %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion))); len(entries) != 0 {
		t.Fatalf("refused compilation published %d entries", len(entries))
	}
	bound, _ := sourceAwareRequest(t, root, &events, nil)
	first, err := RunPipeline(context.Background(), bound)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(first.CacheKey, "sha256:"))
	object := decodeReceiptObject(t, first.Receipt)
	legacyExecution := testExecutionReceiptBytes(t, driverInputOf(object["input"].(map[string]any)), first.Artifact)
	if err := os.WriteFile(filepath.Join(entry, "execution-receipt.ccj.json"), legacyExecution, 0o600); err != nil {
		t.Fatal(err)
	}
	compiles := 0
	bound.Go = recordingGo{events: &events, compile: func(CompileRequest) error { compiles++; return nil }}
	second, err := RunPipeline(context.Background(), bound)
	if err != nil {
		t.Fatal(err)
	}
	if second.State != "would-rebuild-untrusted-cache" || second.Code != CodeReceiptInvalid || compiles != 1 {
		t.Fatalf("legacy-bound execution receipt adopted: state=%s code=%s compiles=%d", second.State, second.Code, compiles)
	}
}

// TestCollectSweepsTheReceipt3Namespace: unreferenced receipt-3 artifacts are
// removed after grace, referenced ones keep their snapshot reachable.
func TestCollectSweepsTheReceipt3Namespace(t *testing.T) {
	root := filepath.Join(t.TempDir(), "cache")
	events := []string{}
	keep, _ := sourceAwareRequest(t, root, &events, nil)
	kept, err := RunPipeline(context.Background(), keep)
	if err != nil {
		t.Fatal(err)
	}
	drop := keep
	drop.Package = &buildmeta.Package{Kind: buildmeta.PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("cd", 32)}
	dropped, err := RunPipeline(context.Background(), drop)
	if err != nil {
		t.Fatal(err)
	}
	removed, err := Collect(root, []string{kept.CacheKey}, time.Now().Add(48*time.Hour), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(removed, "\n")
	if !strings.Contains(joined, strings.TrimPrefix(dropped.CacheKey, "sha256:")) || strings.Contains(joined, strings.TrimPrefix(kept.CacheKey, "sha256:")) || strings.Contains(joined, "external-snapshot") {
		t.Fatalf("removed = %v", removed)
	}
	if _, err := os.Stat(filepath.Join(root, ArtifactsDir(SourceAwareReceiptSchemaVersion), strings.TrimPrefix(kept.CacheKey, "sha256:"), "artifact")); err != nil {
		t.Fatalf("referenced receipt-3 entry swept: %v", err)
	}
}

// TestPrepareNamespacesCreatesEveryProtectedParentPrivately: the store creates
// its root, the snapshot namespace and one artifact namespace per requested
// receipt schema version through its own private creation, idempotently, so a
// transaction renaming entries into the final root never invents a parent.
func TestPrepareNamespacesCreatesEveryProtectedParentPrivately(t *testing.T) {
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	if err := store.PrepareNamespaces(LegacyReceiptSchemaVersion, SourceAwareReceiptSchemaVersion); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "snapshots", ArtifactsDir(LegacyReceiptSchemaVersion), ArtifactsDir(SourceAwareReceiptSchemaVersion)} {
		path := filepath.Join(store.Root, name)
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() {
			t.Fatalf("%s: %v", path, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
			t.Fatalf("%s mode = %o, want 0700", path, info.Mode().Perm())
		}
		// The prepared parent is what the store itself proves private.
		if err := store.protectedDir(path, false); err != nil {
			t.Fatalf("%s is not proved private after preparation: %v", path, err)
		}
	}
	if err := store.PrepareNamespaces(SourceAwareReceiptSchemaVersion); err != nil {
		t.Fatalf("second preparation is not idempotent: %v", err)
	}
	if err := store.PrepareNamespaces(4); err == nil {
		t.Fatal("an unknown receipt schema version prepared a namespace")
	}
	if entries, _ := os.ReadDir(store.Root); len(entries) != 3 {
		t.Fatalf("root entries = %d, want snapshots and two artifact namespaces", len(entries))
	}
}

// TestPrepareNamespacesRefusesAForeignParent: a namespace parent that the
// store did not create privately (here group/world-writable) is refused with
// the protected-boundary code and never repaired.
func TestPrepareNamespacesRefusesAForeignParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mode-bit foreign-parent fixture is exercised on the unix runners; the Windows DACL refusal is asserted by TestWindowsProtectedSecurityDescriptorRejectsWrongOwnerAndDACL")
	}
	store := &DiskProtectedStore{Root: filepath.Join(t.TempDir(), "cache")}
	foreign := filepath.Join(store.Root, ArtifactsDir(SourceAwareReceiptSchemaVersion))
	if err := os.MkdirAll(foreign, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(foreign, 0o777); err != nil {
		t.Fatal(err)
	}
	err := store.PrepareNamespaces(SourceAwareReceiptSchemaVersion)
	if err == nil || !strings.Contains(err.Error(), CodeProtectedBoundaryUntrusted) {
		t.Fatalf("foreign parent accepted: %v", err)
	}
	if info, statErr := os.Lstat(foreign); statErr != nil || info.Mode().Perm() != 0o777 {
		t.Fatalf("refusal repaired the foreign parent: %v %v", info, statErr)
	}
}
