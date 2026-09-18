package buildmeta

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/sourcelock"
)

func localPackage() *Package {
	return &Package{Kind: PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("ab", 32)}
}

func networkPackage() *Package {
	return &Package{Kind: PackageKindNetworkGit, Repository: "git.example.com/skills/tools",
		Commit: &PackageCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: "skills/tool"}
}

func configuredPackage() *Package {
	return &Package{Kind: PackageKindConfiguredGit, Source: "vendor/tools",
		Commit: &PackageCommit{ObjectFormat: "sha256", Hex: strings.Repeat("2", 64)}, Directory: "."}
}

func sourceAwareInput(pkg *Package) Input {
	input := goldenInput()
	input.Package = pkg
	return input
}

func sha256Text(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// TestSourceAwareReceiptWrapsTheClosedDriverInput is the positive row of
// skillfile-sources §4 "Build receipt schema 3": the receipt is
// {schema_version:3, cache_key, input:{schema_version:3,package,build}, artifact},
// the build is the unchanged schema-1 driver input, and the cache key is the
// SHA-256 of the CCJ-1 bytes of the whole wrapped input.
func TestSourceAwareReceiptWrapsTheClosedDriverInput(t *testing.T) {
	for name, pkg := range map[string]*Package{"local-snapshot": localPackage(), "network-git": networkPackage(), "configured-git": configuredPackage()} {
		t.Run(name, func(t *testing.T) {
			input := sourceAwareInput(pkg)
			if input.ReceiptSchemaVersion() != SourceAwareSchemaVersion || !input.SourceAware() {
				t.Fatalf("input did not select receipt schema 3")
			}
			logical, err := input.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			var wrapper map[string]json.RawMessage
			if err := json.Unmarshal(logical, &wrapper); err != nil {
				t.Fatal(err)
			}
			if len(wrapper) != 3 || string(wrapper["schema_version"]) != "3" {
				t.Fatalf("wrapper shape = %s", logical)
			}
			legacy, err := goldenInput().CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wrapper["build"], legacy) {
				t.Fatalf("wrapped build is not the unchanged schema-1 driver input:\n%s\n%s", wrapper["build"], legacy)
			}
			lockPackage := sourcelock.Package{Kind: pkg.Kind, Snapshot: pkg.Snapshot, Repository: pkg.Repository, Source: pkg.Source, Directory: pkg.Directory}
			if pkg.Commit != nil {
				lockPackage.Commit = sourcelock.Commit{ObjectFormat: pkg.Commit.ObjectFormat, Hex: pkg.Commit.Hex}
			}
			lockBytes, err := lockPackage.Canonical()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wrapper["package"], lockBytes) {
				t.Fatalf("receipt package differs from the lock member identity:\n%s\n%s", wrapper["package"], lockBytes)
			}
			key, err := input.CacheKey()
			if err != nil {
				t.Fatal(err)
			}
			if string(key) != sha256Text(logical) {
				t.Fatalf("cache key %s is not SHA-256 of the wrapped input", key)
			}
			if legacyKey, _ := goldenInput().CacheKey(); legacyKey == key {
				t.Fatal("source-aware key aliases the legacy key")
			}
			receipt, err := NewReceipt(input, goldenArtifact())
			if err != nil {
				t.Fatal(err)
			}
			if receipt.SchemaVersion != SourceAwareSchemaVersion || receipt.CacheKey != key {
				t.Fatalf("receipt = %+v", receipt)
			}
			payload, err := receipt.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			if err := protocoljson.Validate(payload); err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeReceipt(payload)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, receipt) {
				t.Fatalf("round trip = %+v, want %+v", decoded, receipt)
			}
			if _, err := DecodeExpectedReceipt(payload, input); err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeExpectedReceipt(payload, goldenInput()); err == nil {
				t.Fatal("a legacy expectation accepted the source-aware receipt")
			}
		})
	}
}

// TestLegacyReceiptBytesAreUnchangedWithoutAPackage pins the switch-off
// path: an input without a package still produces the schema-1 golden key,
// receipt and hash byte for byte.
func TestLegacyReceiptBytesAreUnchangedWithoutAPackage(t *testing.T) {
	input := goldenInput()
	if input.SourceAware() || input.ReceiptSchemaVersion() != SchemaVersion {
		t.Fatal("legacy input selected the source-aware receipt")
	}
	key, err := input.CacheKey()
	if err != nil || string(key) != wantCacheKey {
		t.Fatalf("legacy key = %s, %v; want %s", key, err, wantCacheKey)
	}
	receipt, err := NewReceipt(input, goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	payload, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	rawInput := decodeRaw(t, payload)["input"].(map[string]any)
	if _, wrapped := rawInput["package"]; wrapped || rawInput["build"] != nil || len(rawInput) != 9 || receipt.SchemaVersion != SchemaVersion {
		t.Fatalf("legacy receipt gained wrapper fields: %s", payload)
	}
	hash, err := HashReceiptBytes(payload)
	if err != nil || string(hash) != wantReceiptHash {
		t.Fatalf("legacy receipt hash = %s, %v; want %s", hash, err, wantReceiptHash)
	}
}

// TestSourceAwareKeyFollowsEveryInputMutation: a change to the package
// identity or to any build input field changes the key, so the required
// cache, marker and assurance state is invalidated rather than reused.
func TestSourceAwareKeyFollowsEveryInputMutation(t *testing.T) {
	base := sourceAwareInput(localPackage())
	baseKey, err := base.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	mutants := map[string]func(*Input){
		"package snapshot": func(i *Input) {
			i.Package = &Package{Kind: PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("cd", 32)}
		},
		"package arm":     func(i *Input) { i.Package = networkPackage() },
		"build source":    func(i *Input) { i.BuildSource.ContentSHA256 = "sha256:" + strings.Repeat("e", 64) },
		"toolchain":       func(i *Input) { i.Toolchain.ContentSHA256 = "sha256:" + strings.Repeat("f", 64) },
		"target":          func(i *Input) { i.Target.GOARCH = "amd64"; i.Target.Tuning = map[string]string{"GOAMD64": "v1"} },
		"command":         func(i *Input) { i.Command = "other-tool"; i.SourceDir = "build/cmd/other-tool" },
		"package removed": func(i *Input) { i.Package = nil },
		"network-git commit": func(i *Input) {
			p := *networkPackage()
			p.Commit = &PackageCommit{ObjectFormat: "sha1", Hex: strings.Repeat("3", 40)}
			i.Package = &p
		},
		"network-git dir":       func(i *Input) { p := *networkPackage(); p.Directory = "skills/other"; i.Package = &p },
		"configured-git source": func(i *Input) { p := *configuredPackage(); p.Source = "vendor/other"; i.Package = &p },
	}
	seen := map[CacheKey]string{"": "base"}
	seen[baseKey] = "base"
	for name, mutate := range mutants {
		mutant := sourceAwareInput(localPackage())
		mutate(&mutant)
		key, err := mutant.CacheKey()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if previous, dup := seen[key]; dup {
			t.Fatalf("%s: key collides with %s", name, previous)
		}
		seen[key] = name
	}
}

func decodeRaw(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func encodeRaw(t *testing.T, raw map[string]any) []byte {
	t.Helper()
	payload, err := protocoljson.MarshalCanonical(raw)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestSourceAwareReceiptReaderRefusesEveryShapeConfusion drives the closed
// decoder with every way a receipt-3 document can be confused with a legacy
// one or carry a malformed package: each row must be refused, and the cache
// key of the accepted control must not be derivable from any of them.
func TestSourceAwareReceiptReaderRefusesEveryShapeConfusion(t *testing.T) {
	input := sourceAwareInput(networkPackage())
	receipt, err := NewReceipt(input, goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	control, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeReceipt(control); err != nil {
		t.Fatalf("control refused: %v", err)
	}
	legacyReceipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	legacyBytes, err := legacyReceipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]func(raw map[string]any){
		"receipt schema 1 over wrapper input":    func(raw map[string]any) { raw["schema_version"] = json.Number("1") },
		"receipt schema 2 is not a local schema": func(raw map[string]any) { raw["schema_version"] = json.Number("2") },
		"wrapper schema 1":                       func(raw map[string]any) { raw["input"].(map[string]any)["schema_version"] = json.Number("1") },
		"wrapper unknown field":                  func(raw map[string]any) { raw["input"].(map[string]any)["extra"] = "x" },
		"wrapper missing package":                func(raw map[string]any) { delete(raw["input"].(map[string]any), "package") },
		"wrapper missing build":                  func(raw map[string]any) { delete(raw["input"].(map[string]any), "build") },
		"package null":                           func(raw map[string]any) { raw["input"].(map[string]any)["package"] = nil },
		"package empty":                          func(raw map[string]any) { raw["input"].(map[string]any)["package"] = map[string]any{} },
		"package unknown kind": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["kind"] = "tarball"
		},
		"package foreign arm field": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["snapshot"] = "sha256:" + strings.Repeat("a", 64)
		},
		"package extra field": func(raw map[string]any) { raw["input"].(map[string]any)["package"].(map[string]any)["ref"] = "v1" },
		"package commit null": func(raw map[string]any) { raw["input"].(map[string]any)["package"].(map[string]any)["commit"] = nil },
		"package commit extra field": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["commit"].(map[string]any)["tag"] = "v1"
		},
		"package commit uppercase hex": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["commit"].(map[string]any)["hex"] = strings.Repeat("A", 40)
		},
		"package commit wrong length": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["commit"].(map[string]any)["hex"] = strings.Repeat("1", 64)
		},
		"package commit format": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["commit"].(map[string]any)["object_format"] = "md5"
		},
		"package repository with .git": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["repository"] = "git.example.com/skills/tools.git"
		},
		"package repository uppercase host": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["repository"] = "GIT.example.com/skills/tools"
		},
		"package directory escapes": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["directory"] = "../tool"
		},
		"package directory glob": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["directory"] = "skills/*"
		},
		"package snapshot into git arm": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"] = map[string]any{"kind": "local-snapshot", "snapshot": strings.Repeat("1", 40)}
		},
		"package non-string field": func(raw map[string]any) {
			raw["input"].(map[string]any)["package"].(map[string]any)["repository"] = json.Number("1")
		},
		"build schema 3": func(raw map[string]any) {
			raw["input"].(map[string]any)["build"].(map[string]any)["schema_version"] = json.Number("3")
		},
		"build carries package": func(raw map[string]any) {
			raw["input"].(map[string]any)["build"].(map[string]any)["package"] = raw["input"].(map[string]any)["package"]
		},
		"cache key of inner build":    func(raw map[string]any) { raw["cache_key"] = wantCacheKey },
		"cache key of legacy receipt": func(raw map[string]any) { raw["cache_key"] = string(legacyReceipt.CacheKey) },
	}
	for name, mutate := range rows {
		raw := decodeRaw(t, control)
		mutate(raw)
		payload := encodeRaw(t, raw)
		if _, err := DecodeReceipt(payload); err == nil {
			t.Fatalf("%s: accepted", name)
		}
		if _, err := DecodeExpectedReceipt(payload, input); err == nil {
			t.Fatalf("%s: accepted as the expected input", name)
		}
	}
	// A legacy receipt re-labelled as schema 3 is refused: the wrapper shape
	// is mandatory under version 3, the driver-input shape under version 1.
	raw := decodeRaw(t, legacyBytes)
	raw["schema_version"] = json.Number("3")
	if _, err := DecodeReceipt(encodeRaw(t, raw)); err == nil {
		t.Fatal("legacy driver input accepted under receipt schema 3")
	}
	raw = decodeRaw(t, legacyBytes)
	raw["input"].(map[string]any)["package"] = map[string]any{"kind": "local-snapshot", "snapshot": "sha256:" + strings.Repeat("ab", 32)}
	if _, err := DecodeReceipt(encodeRaw(t, raw)); err == nil {
		t.Fatal("legacy receipt with a smuggled package accepted")
	}
	// Duplicate keys in the wrapper are refused before any lossy decoding.
	duplicated := bytes.Replace(control, []byte(`"input":{`), []byte(`"input":{"build":null,`), 1)
	if _, err := DecodeReceipt(duplicated); err == nil {
		t.Fatal("duplicate wrapper key accepted")
	}
	if _, err := DecodeInput(bytes.Replace(control, []byte(`"schema_version":3}`), []byte(`"schema_version":3}}`), 1)); err == nil {
		t.Fatal("non-canonical input accepted")
	}
}

// TestSourceAwareInputValidationRefusesInvalidPackages: a present package is
// validated by every key derivation, so no receipt can be minted over an
// identity the lock would refuse.
func TestSourceAwareInputValidationRefusesInvalidPackages(t *testing.T) {
	rows := map[string]*Package{
		"empty":                      {},
		"snapshot with git fields":   {Kind: PackageKindLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("ab", 32), Directory: "."},
		"snapshot malformed digest":  {Kind: PackageKindLocalSnapshot, Snapshot: strings.Repeat("ab", 32)},
		"network without commit":     {Kind: PackageKindNetworkGit, Repository: "git.example.com/a/b", Directory: "."},
		"network with snapshot":      {Kind: PackageKindNetworkGit, Repository: "git.example.com/a/b", Commit: &PackageCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: ".", Snapshot: "sha256:" + strings.Repeat("ab", 32)},
		"configured directory":       {Kind: PackageKindConfiguredGit, Source: "vendor/x", Commit: &PackageCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: "sub"},
		"configured absolute source": {Kind: PackageKindConfiguredGit, Source: "/vendor/x", Commit: &PackageCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: "."},
	}
	for name, pkg := range rows {
		input := sourceAwareInput(pkg)
		if err := input.Validate(); err == nil {
			t.Fatalf("%s: validated", name)
		}
		if _, err := input.CacheKey(); err == nil {
			t.Fatalf("%s: derived a cache key", name)
		}
		if _, err := NewReceipt(input, goldenArtifact()); err == nil {
			t.Fatalf("%s: minted a receipt", name)
		}
	}
}
