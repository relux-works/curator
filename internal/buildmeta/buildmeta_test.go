package buildmeta

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	wantCacheKey    = "sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b"
	wantReceiptHash = "sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd"
)

func goldenInput() Input {
	return Input{
		SchemaVersion: SchemaVersion,
		Driver:        DriverGoV1,
		BuildSource: buildsource.Identity{
			Algorithm:     buildsource.Algorithm,
			ContentSHA256: "sha256:" + strings.Repeat("b", 64),
		},
		BuildRoot: "build",
		Command:   "golden-tool",
		SourceDir: "build/cmd/golden-tool",
		Target: Target{
			GOOS: "darwin", GOARCH: "arm64",
			Tuning: map[string]string{"GOARM64": "v8.0"},
		},
		Toolchain: Toolchain{
			Algorithm:     ToolchainAlgorithm,
			GoRelpath:     ToolchainGoRelpath,
			GoVersion:     "go version go1.26.1 darwin/arm64",
			ContentSHA256: "sha256:" + strings.Repeat("c", 64),
		},
		Policy: FixedPolicy(),
	}
}

func goldenArtifact() Artifact {
	return Artifact{
		Path:   "bin/golden-tool",
		SHA256: "sha256:" + strings.Repeat("d", 64),
		Size:   1234567,
	}
}

func TestGoV1GoldenInputKeyReceiptAndHash(t *testing.T) {
	input := goldenInput()
	inputBytes, err := input.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	wantInput := `{"build_root":"build","build_source":{"algorithm":"curator-build-source-v1","content_sha256":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},"command":"golden-tool","driver":"go-v1","policy":{"cgo":false,"compiler_directives":"reject-nonstandard-cgo-import-dynamic-v1","execution_policy":"manager-worker-v1","host_objects":false,"libgcc":"none","link_mode":"internal","module_mode":"vendor","network":"none","package_assembly":false,"target_mode":"native","telemetry":"off-private","workspace":false},"schema_version":1,"source_dir":"build/cmd/golden-tool","target":{"goarch":"arm64","goos":"darwin","tuning":{"GOARM64":"v8.0"}},"toolchain":{"algorithm":"curator-go-toolchain-v1","content_sha256":"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","go_relpath":"bin/go","go_version":"go version go1.26.1 darwin/arm64"}}`
	if string(inputBytes) != wantInput {
		t.Fatalf("input bytes:\n got %s\nwant %s", inputBytes, wantInput)
	}
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	if string(key) != wantCacheKey {
		t.Fatalf("cache key = %s, want %s", key, wantCacheKey)
	}

	receipt, err := NewReceipt(input, goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	wantReceipt := `{"artifact":{"path":"bin/golden-tool","sha256":"sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","size":1234567},"cache_key":"` + wantCacheKey + `","input":` + wantInput + `,"schema_version":1}`
	if string(receiptBytes) != wantReceipt {
		t.Fatalf("receipt bytes:\n got %s\nwant %s", receiptBytes, wantReceipt)
	}
	hash, err := HashReceiptBytes(receiptBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(hash) != wantReceiptHash {
		t.Fatalf("receipt hash = %s, want %s", hash, wantReceiptHash)
	}

	decodedInput, err := DecodeInput(inputBytes)
	if err != nil || !reflect.DeepEqual(decodedInput, input) {
		t.Fatalf("DecodeInput() = %+v, %v", decodedInput, err)
	}
	decodedReceipt, err := DecodeExpectedReceipt(receiptBytes, input)
	if err != nil || !reflect.DeepEqual(decodedReceipt, receipt) {
		t.Fatalf("DecodeExpectedReceipt() = %+v, %v", decodedReceipt, err)
	}
}

func TestCandidateBuildMetadataArtifacts(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	expected := filepath.Join(root, "expected", "build-driver")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Skipf("%s publishes no expected/build-driver artifacts", root)
	}
	inputBytes := readFile(t, filepath.Join(expected, "build-input.ccj.json"))
	if !bytes.Contains(inputBytes, []byte(`"execution_policy":"`+ExecutionPolicy+`"`)) {
		t.Skipf("%s is a pre-revision root without the portable execution policy", root)
	}
	receiptBytes := readFile(t, filepath.Join(expected, "receipt.ccj.json"))
	if got, err := goldenInput().CanonicalBytes(); err != nil || !bytes.Equal(got, inputBytes) {
		t.Fatalf("candidate input mismatch: %v\n got %s\nwant %s", err, got, inputBytes)
	}
	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := receipt.CanonicalBytes(); err != nil || !bytes.Equal(got, receiptBytes) {
		t.Fatalf("candidate receipt mismatch: %v\n got %s\nwant %s", err, got, receiptBytes)
	}
	if got := strings.TrimSpace(string(readFile(t, filepath.Join(expected, "cache-key.txt")))); got != wantCacheKey {
		t.Fatalf("candidate cache key = %s", got)
	}
	if got := strings.TrimSpace(string(readFile(t, filepath.Join(expected, "receipt-sha256.txt")))); got != wantReceiptHash {
		t.Fatalf("candidate receipt hash = %s", got)
	}
}

// TestCandidateBuildReceiptSchemaCase proves the implementation reproduces the
// accepted rc.5 build-receipt-v1 case, including the normative portable
// execution-policy identity and its derived cache key.
func TestCandidateBuildReceiptSchemaCase(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "schema-cases", "build-receipt-v1", "valid.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-receipt-v1 schema case", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var candidate map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&candidate); err != nil {
		t.Fatal(err)
	}
	policy, _ := candidate["input"].(map[string]any)["policy"].(map[string]any)
	if policy["execution_policy"] == nil {
		t.Skipf("%s is a pre-revision root without the portable execution policy", root)
	}
	if policy["execution_policy"] != ExecutionPolicy {
		t.Fatalf("candidate policy execution_policy = %v, want %q", policy["execution_policy"], ExecutionPolicy)
	}
	if candidate["cache_key"] != wantCacheKey {
		t.Fatalf("candidate cache_key = %v, want %s", candidate["cache_key"], wantCacheKey)
	}
	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	got, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	want, err := protocoljson.MarshalCanonical(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("canonical receipt mismatch\n got %s\nwant %s", got, want)
	}
}

func TestStrictReadersRejectNoncanonicalOrIncompleteMetadata(t *testing.T) {
	inputBytes, err := goldenInput().CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, receiptBytes, "", "  "); err != nil {
		t.Fatal(err)
	}
	receiptCases := map[string][]byte{
		"pretty printed":   pretty.Bytes(),
		"terminal newline": append(append([]byte{}, receiptBytes...), '\n'),
		"byte order mark":  append([]byte{0xef, 0xbb, 0xbf}, receiptBytes...),
		"duplicate key":    []byte(strings.Replace(string(receiptBytes), `{`, `{"schema_version":1,`, 1)),
		"noninteger size":  []byte(strings.Replace(string(receiptBytes), `"size":1234567`, `"size":1.5`, 1)),
		"unknown field": mutateObject(t, receiptBytes, func(value map[string]any) {
			value["created_at"] = "2026-07-21T00:00:00Z"
		}),
		"unsupported version": mutateObject(t, receiptBytes, func(value map[string]any) {
			value["schema_version"] = json.Number("2")
		}),
		"missing artifact": mutateObject(t, receiptBytes, func(value map[string]any) {
			delete(value, "artifact")
		}),
	}
	for name, payload := range receiptCases {
		t.Run("receipt "+name, func(t *testing.T) {
			if _, err := DecodeReceipt(payload); err == nil {
				t.Fatal("receipt was accepted")
			}
		})
	}

	inputCases := map[string][]byte{
		"unknown field": mutateObject(t, inputBytes, func(value map[string]any) {
			value["timestamp"] = "2026-07-21T00:00:00Z"
		}),
		"unsupported version": mutateObject(t, inputBytes, func(value map[string]any) {
			value["schema_version"] = json.Number("2")
		}),
		"missing policy": mutateObject(t, inputBytes, func(value map[string]any) {
			delete(value, "policy")
		}),
		"missing false policy field": mutateObject(t, inputBytes, func(value map[string]any) {
			delete(value["policy"].(map[string]any), "workspace")
		}),
		"absolute build root": mutateObject(t, inputBytes, func(value map[string]any) {
			value["build_root"] = "/build"
		}),
	}
	for name, payload := range inputCases {
		t.Run("input "+name, func(t *testing.T) {
			if _, err := DecodeInput(payload); err == nil {
				t.Fatal("input was accepted")
			}
		})
	}
}

func TestStrictReadersRejectEveryMissingOrUnknownField(t *testing.T) {
	inputBytes, err := goldenInput().CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	inputFields := map[string][]string{
		"":             {"schema_version", "driver", "build_source", "build_root", "command", "source_dir", "target", "toolchain", "policy"},
		"build_source": {"algorithm", "content_sha256"},
		"target":       {"goos", "goarch", "tuning"},
		"toolchain":    {"algorithm", "go_relpath", "go_version", "content_sha256"},
		"policy":       {"module_mode", "network", "workspace", "cgo", "compiler_directives", "target_mode", "link_mode", "libgcc", "package_assembly", "host_objects", "telemetry"},
	}
	for objectName, fields := range inputFields {
		for _, field := range fields {
			name := objectName + "." + field
			t.Run("missing "+name, func(t *testing.T) {
				payload := mutateObject(t, inputBytes, func(value map[string]any) {
					object := value
					if objectName != "" {
						object = value[objectName].(map[string]any)
					}
					delete(object, field)
				})
				if _, err := DecodeInput(payload); err == nil {
					t.Fatal("incomplete input was accepted")
				}
			})
		}
	}
	for _, objectName := range []string{"build_source", "target", "toolchain", "policy"} {
		t.Run("unknown "+objectName, func(t *testing.T) {
			payload := mutateObject(t, inputBytes, func(value map[string]any) {
				value[objectName].(map[string]any)["extension"] = true
			})
			if _, err := DecodeInput(payload); err == nil {
				t.Fatal("unknown nested input field was accepted")
			}
		})
	}

	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes := mustReceiptBytes(t, receipt)
	for objectName, fields := range map[string][]string{
		"":         {"schema_version", "cache_key", "input", "artifact"},
		"artifact": {"path", "sha256", "size"},
	} {
		for _, field := range fields {
			t.Run("receipt missing "+objectName+"."+field, func(t *testing.T) {
				payload := mutateObject(t, receiptBytes, func(value map[string]any) {
					object := value
					if objectName != "" {
						object = value[objectName].(map[string]any)
					}
					delete(object, field)
				})
				if _, err := DecodeReceipt(payload); err == nil {
					t.Fatal("incomplete receipt was accepted")
				}
			})
		}
	}
	unknownArtifact := mutateObject(t, receiptBytes, func(value map[string]any) {
		value["artifact"].(map[string]any)["mode"] = "0755"
	})
	if _, err := DecodeReceipt(unknownArtifact); err == nil {
		t.Fatal("unknown artifact field was accepted")
	}
}

func TestReceiptReaderRejectsDerivedIdentityAndArtifactMismatches(t *testing.T) {
	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	payload := mustReceiptBytes(t, receipt)
	cases := map[string]func(map[string]any){
		"cache key": func(value map[string]any) {
			value["cache_key"] = "sha256:" + strings.Repeat("0", 64)
		},
		"artifact path": func(value map[string]any) {
			value["artifact"].(map[string]any)["path"] = "bin/other"
		},
		"artifact hash": func(value map[string]any) {
			value["artifact"].(map[string]any)["sha256"] = "sha256:deadbeef"
		},
		"artifact size": func(value map[string]any) {
			value["artifact"].(map[string]any)["size"] = json.Number("-1")
		},
		"build source algorithm": func(value map[string]any) {
			value["input"].(map[string]any)["build_source"].(map[string]any)["algorithm"] = "curator-build-source-v2"
		},
		"toolchain algorithm": func(value map[string]any) {
			value["input"].(map[string]any)["toolchain"].(map[string]any)["algorithm"] = "curator-go-toolchain-v2"
		},
		"target tuning": func(value map[string]any) {
			value["input"].(map[string]any)["target"].(map[string]any)["tuning"] = map[string]any{"GOAMD64": "v3"}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeReceipt(mutateObject(t, payload, mutate)); err == nil {
				t.Fatal("mismatched receipt was accepted")
			}
		})
	}
}

func TestExpectedInputComparisonCoversEveryLogicalComponent(t *testing.T) {
	expected := goldenInput()
	receipt, err := NewReceipt(expected, goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}

	mutations := map[string]func(*Input){
		"build root": func(value *Input) {
			value.BuildRoot = "other"
			value.SourceDir = "other/cmd/golden-tool"
		},
		"source dir": func(value *Input) { value.SourceDir = "build/cmd/other" },
		"source identity": func(value *Input) {
			value.BuildSource.ContentSHA256 = "sha256:" + strings.Repeat("a", 64)
		},
		"target":            func(value *Input) { value.Target.GOOS = "linux" },
		"target tuning":     func(value *Input) { value.Target.Tuning["GOARM64"] = "v8.1" },
		"toolchain version": func(value *Input) { value.Toolchain.GoVersion = "go version go1.26.2 darwin/arm64" },
		"toolchain identity": func(value *Input) {
			value.Toolchain.ContentSHA256 = "sha256:" + strings.Repeat("e", 64)
		},
		"command": func(value *Input) { value.Command = "other-tool" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			changed := cloneInput(receipt.Input)
			mutate(&changed)
			changedReceipt, err := NewReceipt(changed, Artifact{
				Path: artifactPathForTest(t, changed), SHA256: receipt.Artifact.SHA256, Size: receipt.Artifact.Size,
			})
			if err != nil {
				t.Fatal(err)
			}
			payload, err := changedReceipt.CanonicalBytes()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeExpectedReceipt(payload, expected); err == nil {
				t.Fatal("mismatched expected input was accepted")
			}
		})
	}

	badPolicy := mutateObject(t, mustReceiptBytes(t, receipt), func(value map[string]any) {
		value["input"].(map[string]any)["policy"].(map[string]any)["network"] = "host"
	})
	if _, err := DecodeExpectedReceipt(badPolicy, expected); err == nil {
		t.Fatal("wrong directive/network policy was accepted")
	}
}

func TestArtifactConstraintsArePlatformDerived(t *testing.T) {
	for _, testCase := range []struct {
		goos string
		want string
	}{
		{goos: "darwin", want: "bin/golden-tool"},
		{goos: "linux", want: "bin/golden-tool"},
		{goos: "windows", want: "bin/golden-tool.exe"},
	} {
		t.Run(testCase.goos, func(t *testing.T) {
			got, err := ArtifactPath("golden-tool", testCase.goos)
			if err != nil || got != testCase.want {
				t.Fatalf("ArtifactPath() = %q, %v; want %q", got, err, testCase.want)
			}
		})
	}

	input := goldenInput()
	for name, artifact := range map[string]Artifact{
		"wrong path":    {Path: "bin/other", SHA256: goldenArtifact().SHA256, Size: 1},
		"absolute path": {Path: "/bin/golden-tool", SHA256: goldenArtifact().SHA256, Size: 1},
		"bad hash":      {Path: "bin/golden-tool", SHA256: "sha256:ABC", Size: 1},
		"negative size": {Path: "bin/golden-tool", SHA256: goldenArtifact().SHA256, Size: -1},
		"unsafe size":   {Path: "bin/golden-tool", SHA256: goldenArtifact().SHA256, Size: MaxSafeInteger + 1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewReceipt(input, artifact); err == nil {
				t.Fatal("invalid artifact was accepted")
			}
		})
	}
}

func TestTargetTuningIsArchitectureDerived(t *testing.T) {
	for architecture, key := range map[string]string{
		"386": "GO386", "amd64": "GOAMD64", "arm": "GOARM", "arm64": "GOARM64",
		"mips": "GOMIPS", "mipsle": "GOMIPS", "mips64": "GOMIPS64", "mips64le": "GOMIPS64",
		"ppc64": "GOPPC64", "ppc64le": "GOPPC64", "riscv64": "GORISCV64", "wasm": "GOWASM",
	} {
		t.Run(architecture, func(t *testing.T) {
			target := Target{GOOS: "linux", GOARCH: architecture, Tuning: map[string]string{key: "default"}}
			if err := target.validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
	if err := (Target{GOOS: "linux", GOARCH: "loong64", Tuning: map[string]string{}}).validate(); err != nil {
		t.Fatal(err)
	}
	invalid := []Target{
		{GOOS: "linux", GOARCH: "arm64", Tuning: nil},
		{GOOS: "linux", GOARCH: "arm64", Tuning: map[string]string{}},
		{GOOS: "linux", GOARCH: "arm64", Tuning: map[string]string{"GOAMD64": "v3"}},
		{GOOS: "linux", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": ""}},
		{GOOS: "linux", GOARCH: "loong64", Tuning: map[string]string{"GOARM64": "v8.0"}},
		{GOOS: "/linux", GOARCH: "arm64", Tuning: map[string]string{"GOARM64": "v8.0"}},
	}
	for index, target := range invalid {
		if err := target.validate(); err == nil {
			t.Errorf("invalid target %d was accepted", index)
		}
	}
}

func TestInputAndToolchainValidationRejectMalformedModels(t *testing.T) {
	inputMutations := []func(*Input){
		func(value *Input) { value.SchemaVersion = 2 },
		func(value *Input) { value.Driver = "go-v2" },
		func(value *Input) { value.BuildSource.ContentSHA256 = "sha256:bad" },
		func(value *Input) { value.BuildRoot = "/build" },
		func(value *Input) { value.Command = "-tool" },
		func(value *Input) { value.SourceDir = "/build/cmd" },
		func(value *Input) { value.SourceDir = "other/cmd" },
		func(value *Input) { value.Policy.Telemetry = "on" },
	}
	for index, mutate := range inputMutations {
		input := goldenInput()
		mutate(&input)
		if err := input.Validate(); err == nil {
			t.Errorf("invalid input mutation %d was accepted", index)
		}
		if _, err := input.CanonicalBytes(); err == nil {
			t.Errorf("invalid input mutation %d was encoded", index)
		}
	}

	toolchains := []Toolchain{
		{Algorithm: "v2", GoRelpath: ToolchainGoRelpath, GoVersion: "go version go1.26", ContentSHA256: "sha256:" + strings.Repeat("a", 64)},
		{Algorithm: ToolchainAlgorithm, GoRelpath: "go", GoVersion: "go version go1.26", ContentSHA256: "sha256:" + strings.Repeat("a", 64)},
		{Algorithm: ToolchainAlgorithm, GoRelpath: ToolchainGoRelpath, GoVersion: "", ContentSHA256: "sha256:" + strings.Repeat("a", 64)},
		{Algorithm: ToolchainAlgorithm, GoRelpath: ToolchainGoRelpath, GoVersion: "go version\n", ContentSHA256: "sha256:" + strings.Repeat("a", 64)},
		{Algorithm: ToolchainAlgorithm, GoRelpath: ToolchainGoRelpath, GoVersion: "go version", ContentSHA256: "sha256:bad"},
	}
	for index, toolchain := range toolchains {
		if err := toolchain.validate(); err == nil {
			t.Errorf("invalid toolchain %d was accepted", index)
		}
	}
	if _, err := ArtifactPath("-tool", "linux"); err == nil {
		t.Fatal("invalid artifact command was accepted")
	}
	if _, err := ArtifactPath("tool", "/linux"); err == nil {
		t.Fatal("invalid artifact GOOS was accepted")
	}
}

func TestReadersRejectWrongJSONTypes(t *testing.T) {
	inputBytes, err := goldenInput().CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	cases := []func(map[string]any){
		func(value map[string]any) { value["schema_version"] = "1" },
		func(value map[string]any) { value["driver"] = true },
		func(value map[string]any) { value["target"] = "darwin/arm64" },
		func(value map[string]any) { value["target"].(map[string]any)["tuning"] = []any{} },
		func(value map[string]any) {
			value["target"].(map[string]any)["tuning"].(map[string]any)["GOARM64"] = true
		},
		func(value map[string]any) { value["policy"].(map[string]any)["workspace"] = "false" },
		func(value map[string]any) { value["toolchain"].(map[string]any)["go_version"] = false },
	}
	for index, mutate := range cases {
		if _, err := DecodeInput(mutateObject(t, inputBytes, mutate)); err == nil {
			t.Errorf("wrong JSON type mutation %d was accepted", index)
		}
	}

	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes := mustReceiptBytes(t, receipt)
	for index, mutate := range []func(map[string]any){
		func(value map[string]any) { value["input"] = true },
		func(value map[string]any) { value["artifact"] = true },
		func(value map[string]any) { value["artifact"].(map[string]any)["path"] = false },
		func(value map[string]any) { value["artifact"].(map[string]any)["size"] = "1" },
	} {
		if _, err := DecodeReceipt(mutateObject(t, receiptBytes, mutate)); err == nil {
			t.Errorf("wrong receipt JSON type mutation %d was accepted", index)
		}
	}
}

func TestReceiptHashIsConsistencyMetadata(t *testing.T) {
	receipt, err := NewReceipt(goldenInput(), goldenArtifact())
	if err != nil {
		t.Fatal(err)
	}
	payload := mustReceiptBytes(t, receipt)
	if err := CheckReceiptHash(payload, ReceiptHash(wantReceiptHash)); err != nil {
		t.Fatal(err)
	}
	if err := CheckReceiptHash(payload, ReceiptHash("sha256:"+strings.Repeat("0", 64))); err == nil {
		t.Fatal("mismatched consistency hash was accepted")
	}
	if err := CheckReceiptHash(append(payload, '\n'), ReceiptHash(wantReceiptHash)); err == nil {
		t.Fatal("hash check accepted noncanonical receipt bytes")
	}
}

func mutateObject(t *testing.T, payload []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	mutate(value)
	result, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func cloneInput(value Input) Input {
	cloned := value
	cloned.Target.Tuning = make(map[string]string, len(value.Target.Tuning))
	for key, item := range value.Target.Tuning {
		cloned.Target.Tuning[key] = item
	}
	return cloned
}

func artifactPathForTest(t *testing.T, input Input) string {
	t.Helper()
	path, err := ArtifactPath(input.Command, input.Target.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func mustReceiptBytes(t *testing.T, receipt Receipt) []byte {
	t.Helper()
	payload, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
