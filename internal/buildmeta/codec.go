package buildmeta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/protocoljson"
)

// CanonicalBytes returns the exact CCJ-1 logical input bytes used to derive the
// cache key.
func (input Input) CanonicalBytes() ([]byte, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(inputValue(input))
}

// CacheKey returns SHA-256 of the exact canonical logical input bytes.
func (input Input) CacheKey() (CacheKey, error) {
	payload, err := input.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return CacheKey(hashBytes(payload)), nil
}

// DecodeInput reads only exact canonical CCJ-1 and rejects incomplete,
// unknown, or unsupported logical input metadata.
func DecodeInput(payload []byte) (Input, error) {
	var raw map[string]any
	if err := protocoljson.UnmarshalCanonical(payload, &raw); err != nil {
		return Input{}, fmt.Errorf("decode build input: %w", err)
	}
	input, err := parseInput(raw)
	if err != nil {
		return Input{}, err
	}
	if err := input.Validate(); err != nil {
		return Input{}, err
	}
	return input, nil
}

// NewReceipt binds a validated logical input to its derived cache key and the
// one platform-derived artifact path.
func NewReceipt(input Input, artifact Artifact) (Receipt, error) {
	if err := input.Validate(); err != nil {
		return Receipt{}, err
	}
	if err := artifact.validate(input); err != nil {
		return Receipt{}, err
	}
	key, err := input.CacheKey()
	if err != nil {
		return Receipt{}, err
	}
	return Receipt{SchemaVersion: SchemaVersion, CacheKey: key, Input: input, Artifact: artifact}, nil
}

// Validate checks the receipt schema, its complete logical input, derived key,
// and platform-derived artifact metadata.
func (receipt Receipt) Validate() error {
	if receipt.SchemaVersion != SchemaVersion {
		return fmt.Errorf("build receipt schema_version must be %d", SchemaVersion)
	}
	if err := receipt.Input.Validate(); err != nil {
		return err
	}
	wantKey, err := receipt.Input.CacheKey()
	if err != nil {
		return err
	}
	if !validSHA256(string(receipt.CacheKey)) || receipt.CacheKey != wantKey {
		return fmt.Errorf("receipt cache_key does not match its complete input")
	}
	return receipt.Artifact.validate(receipt.Input)
}

// CanonicalBytes emits exact CCJ-1 receipt bytes with no BOM, whitespace, or
// terminal newline.
func (receipt Receipt) CanonicalBytes() ([]byte, error) {
	if err := receipt.Validate(); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(receiptValue(receipt))
}

// DecodeReceipt reads an exact canonical schema-1 receipt and validates all of
// its internally derived identities.
func DecodeReceipt(payload []byte) (Receipt, error) {
	var raw map[string]any
	if err := protocoljson.UnmarshalCanonical(payload, &raw); err != nil {
		return Receipt{}, fmt.Errorf("decode build receipt: %w", err)
	}
	if err := exactFields(raw, "receipt", "schema_version", "cache_key", "input", "artifact"); err != nil {
		return Receipt{}, err
	}
	version, err := integerField(raw, "schema_version", "receipt")
	if err != nil {
		return Receipt{}, err
	}
	cacheKey, err := stringField(raw, "cache_key", "receipt")
	if err != nil {
		return Receipt{}, err
	}
	inputObject, err := objectField(raw, "input", "receipt")
	if err != nil {
		return Receipt{}, err
	}
	input, err := parseInput(inputObject)
	if err != nil {
		return Receipt{}, err
	}
	artifactObject, err := objectField(raw, "artifact", "receipt")
	if err != nil {
		return Receipt{}, err
	}
	artifact, err := parseArtifact(artifactObject)
	if err != nil {
		return Receipt{}, err
	}
	receipt := Receipt{SchemaVersion: int(version), CacheKey: CacheKey(cacheKey), Input: input, Artifact: artifact}
	if err := receipt.Validate(); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

// DecodeExpectedReceipt additionally requires the complete receipt input to
// equal the independently derived expected input. No subset or cache key alone
// is accepted as an identity comparison.
func DecodeExpectedReceipt(payload []byte, expected Input) (Receipt, error) {
	if err := expected.Validate(); err != nil {
		return Receipt{}, fmt.Errorf("expected build input: %w", err)
	}
	receipt, err := DecodeReceipt(payload)
	if err != nil {
		return Receipt{}, err
	}
	if !reflect.DeepEqual(receipt.Input, expected) {
		return Receipt{}, fmt.Errorf("receipt input does not match the complete expected build input")
	}
	return receipt, nil
}

// HashReceiptBytes returns the deterministic receipt consistency identifier.
// It validates the receipt first; the result does not authenticate provenance.
func HashReceiptBytes(payload []byte) (ReceiptHash, error) {
	if _, err := DecodeReceipt(payload); err != nil {
		return "", err
	}
	return ReceiptHash(hashBytes(payload)), nil
}

// CheckReceiptHash validates consistency metadata only. A match is not proof
// of authorship, authorization, attestation, or protected-cache provenance.
func CheckReceiptHash(payload []byte, expected ReceiptHash) error {
	if !validSHA256(string(expected)) {
		return fmt.Errorf("expected receipt hash is malformed")
	}
	actual, err := HashReceiptBytes(payload)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("receipt hash mismatch")
	}
	return nil
}

func hashBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func parseInput(raw map[string]any) (Input, error) {
	if err := exactFields(raw, "build input", "schema_version", "driver", "build_source", "build_root", "command", "source_dir", "target", "toolchain", "policy"); err != nil {
		return Input{}, err
	}
	version, err := integerField(raw, "schema_version", "build input")
	if err != nil {
		return Input{}, err
	}
	driver, err := stringField(raw, "driver", "build input")
	if err != nil {
		return Input{}, err
	}
	buildSourceObject, err := objectField(raw, "build_source", "build input")
	if err != nil {
		return Input{}, err
	}
	if err := exactFields(buildSourceObject, "build_source", "algorithm", "content_sha256"); err != nil {
		return Input{}, err
	}
	buildSourceAlgorithm, err := stringField(buildSourceObject, "algorithm", "build_source")
	if err != nil {
		return Input{}, err
	}
	buildSourceSHA256, err := stringField(buildSourceObject, "content_sha256", "build_source")
	if err != nil {
		return Input{}, err
	}
	buildRoot, err := stringField(raw, "build_root", "build input")
	if err != nil {
		return Input{}, err
	}
	command, err := stringField(raw, "command", "build input")
	if err != nil {
		return Input{}, err
	}
	sourceDir, err := stringField(raw, "source_dir", "build input")
	if err != nil {
		return Input{}, err
	}
	targetObject, err := objectField(raw, "target", "build input")
	if err != nil {
		return Input{}, err
	}
	target, err := parseTarget(targetObject)
	if err != nil {
		return Input{}, err
	}
	toolchainObject, err := objectField(raw, "toolchain", "build input")
	if err != nil {
		return Input{}, err
	}
	toolchain, err := parseToolchain(toolchainObject)
	if err != nil {
		return Input{}, err
	}
	policyObject, err := objectField(raw, "policy", "build input")
	if err != nil {
		return Input{}, err
	}
	policy, err := parsePolicy(policyObject)
	if err != nil {
		return Input{}, err
	}
	return Input{
		SchemaVersion: int(version), Driver: driver,
		BuildSource: buildsource.Identity{Algorithm: buildSourceAlgorithm, ContentSHA256: buildSourceSHA256},
		BuildRoot:   buildRoot, Command: command, SourceDir: sourceDir,
		Target: target, Toolchain: toolchain, Policy: policy,
	}, nil
}

func parseTarget(raw map[string]any) (Target, error) {
	if err := exactFields(raw, "target", "goos", "goarch", "tuning"); err != nil {
		return Target{}, err
	}
	goos, err := stringField(raw, "goos", "target")
	if err != nil {
		return Target{}, err
	}
	goarch, err := stringField(raw, "goarch", "target")
	if err != nil {
		return Target{}, err
	}
	tuningObject, err := objectField(raw, "tuning", "target")
	if err != nil {
		return Target{}, err
	}
	tuning := make(map[string]string, len(tuningObject))
	for key, rawValue := range tuningObject {
		value, ok := rawValue.(string)
		if !ok {
			return Target{}, fmt.Errorf("target tuning field %q must be a string", key)
		}
		tuning[key] = value
	}
	return Target{GOOS: goos, GOARCH: goarch, Tuning: tuning}, nil
}

func parseToolchain(raw map[string]any) (Toolchain, error) {
	if err := exactFields(raw, "toolchain", "algorithm", "go_relpath", "go_version", "content_sha256"); err != nil {
		return Toolchain{}, err
	}
	algorithm, err := stringField(raw, "algorithm", "toolchain")
	if err != nil {
		return Toolchain{}, err
	}
	relpath, err := stringField(raw, "go_relpath", "toolchain")
	if err != nil {
		return Toolchain{}, err
	}
	version, err := stringField(raw, "go_version", "toolchain")
	if err != nil {
		return Toolchain{}, err
	}
	contentSHA256, err := stringField(raw, "content_sha256", "toolchain")
	if err != nil {
		return Toolchain{}, err
	}
	return Toolchain{Algorithm: algorithm, GoRelpath: relpath, GoVersion: version, ContentSHA256: contentSHA256}, nil
}

func parsePolicy(raw map[string]any) (Policy, error) {
	fields := []string{"module_mode", "network", "workspace", "cgo", "compiler_directives", "execution_policy", "target_mode", "link_mode", "libgcc", "package_assembly", "host_objects", "telemetry"}
	if err := exactFields(raw, "policy", fields...); err != nil {
		return Policy{}, err
	}
	moduleMode, err := stringField(raw, "module_mode", "policy")
	if err != nil {
		return Policy{}, err
	}
	network, err := stringField(raw, "network", "policy")
	if err != nil {
		return Policy{}, err
	}
	workspace, err := boolField(raw, "workspace", "policy")
	if err != nil {
		return Policy{}, err
	}
	cgo, err := boolField(raw, "cgo", "policy")
	if err != nil {
		return Policy{}, err
	}
	directives, err := stringField(raw, "compiler_directives", "policy")
	if err != nil {
		return Policy{}, err
	}
	executionPolicy, err := stringField(raw, "execution_policy", "policy")
	if err != nil {
		return Policy{}, err
	}
	targetMode, err := stringField(raw, "target_mode", "policy")
	if err != nil {
		return Policy{}, err
	}
	linkMode, err := stringField(raw, "link_mode", "policy")
	if err != nil {
		return Policy{}, err
	}
	libgcc, err := stringField(raw, "libgcc", "policy")
	if err != nil {
		return Policy{}, err
	}
	assembly, err := boolField(raw, "package_assembly", "policy")
	if err != nil {
		return Policy{}, err
	}
	hostObjects, err := boolField(raw, "host_objects", "policy")
	if err != nil {
		return Policy{}, err
	}
	telemetry, err := stringField(raw, "telemetry", "policy")
	if err != nil {
		return Policy{}, err
	}
	return Policy{
		ModuleMode: moduleMode, Network: network, Workspace: workspace, CGO: cgo,
		CompilerDirectives: directives, ExecutionPolicy: executionPolicy,
		TargetMode: targetMode, LinkMode: linkMode,
		Libgcc: libgcc, PackageAssembly: assembly, HostObjects: hostObjects, Telemetry: telemetry,
	}, nil
}

func parseArtifact(raw map[string]any) (Artifact, error) {
	if err := exactFields(raw, "artifact", "path", "sha256", "size"); err != nil {
		return Artifact{}, err
	}
	path, err := stringField(raw, "path", "artifact")
	if err != nil {
		return Artifact{}, err
	}
	contentSHA256, err := stringField(raw, "sha256", "artifact")
	if err != nil {
		return Artifact{}, err
	}
	size, err := integerField(raw, "size", "artifact")
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{Path: path, SHA256: contentSHA256, Size: size}, nil
}

func exactFields(raw map[string]any, label string, fields ...string) error {
	want := make(map[string]bool, len(fields))
	for _, field := range fields {
		want[field] = true
	}
	var missing, unknown []string
	for _, field := range fields {
		if _, present := raw[field]; !present {
			missing = append(missing, field)
		}
	}
	for field := range raw {
		if !want[field] {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(missing)
	sort.Strings(unknown)
	if len(missing) > 0 {
		return fmt.Errorf("%s is missing required fields: %s", label, strings.Join(missing, ", "))
	}
	if len(unknown) > 0 {
		return fmt.Errorf("%s has unknown fields: %s", label, strings.Join(unknown, ", "))
	}
	return nil
}

func objectField(raw map[string]any, field, label string) (map[string]any, error) {
	value, ok := raw[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s field %q must be an object", label, field)
	}
	return value, nil
}

func stringField(raw map[string]any, field, label string) (string, error) {
	value, ok := raw[field].(string)
	if !ok {
		return "", fmt.Errorf("%s field %q must be a string", label, field)
	}
	return value, nil
}

func boolField(raw map[string]any, field, label string) (bool, error) {
	value, ok := raw[field].(bool)
	if !ok {
		return false, fmt.Errorf("%s field %q must be a boolean", label, field)
	}
	return value, nil
}

func integerField(raw map[string]any, field, label string) (int64, error) {
	number, ok := raw[field].(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s field %q must be an integer", label, field)
	}
	value, err := number.Int64()
	if err != nil {
		return 0, fmt.Errorf("%s field %q must be an integer", label, field)
	}
	return value, nil
}
