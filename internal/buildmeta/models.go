// Package buildmeta defines the portable logical go-v1 build input and strict
// receipt metadata. It deliberately owns no filesystem cache behavior.
package buildmeta

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	SchemaVersion = 1
	DriverGoV1    = "go-v1"

	ToolchainAlgorithm = "curator-go-toolchain-v1"
	ToolchainGoRelpath = "bin/go"

	MaxSafeInteger = protocoljson.MaxSafeInteger
)

const (
	policyModuleMode         = "vendor"
	policyNetwork            = "none"
	policyCompilerDirectives = "reject-nonstandard-cgo-import-dynamic-v1"
	policyTargetMode         = "native"
	policyLinkMode           = "internal"
	policyLibgcc             = "none"
	policyTelemetry          = "off-private"
)

var sha256Pattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Target is the native Go target and its single architecture-specific tuning
// input. Architectures without a Go tuning variable use an empty map.
type Target struct {
	GOOS   string
	GOARCH string
	Tuning map[string]string
}

// Toolchain is the location-independent curator-go-toolchain-v1 identity.
type Toolchain struct {
	Algorithm     string
	GoRelpath     string
	GoVersion     string
	ContentSHA256 string
}

// Policy is the closed go-v1 directive and build policy. FixedPolicy returns
// the only valid value.
type Policy struct {
	ModuleMode         string
	Network            string
	Workspace          bool
	CGO                bool
	CompilerDirectives string
	TargetMode         string
	LinkMode           string
	Libgcc             string
	PackageAssembly    bool
	HostObjects        bool
	Telemetry          string
}

// Input is the complete logical go-v1 cache input. It contains portable
// identities and paths only; absolute paths and timestamps are not represented.
type Input struct {
	SchemaVersion int
	Driver        string
	BuildSource   buildsource.Identity
	BuildRoot     string
	Command       string
	SourceDir     string
	Target        Target
	Toolchain     Toolchain
	Policy        Policy
}

// Artifact is the single manager-derived executable described by a receipt.
type Artifact struct {
	Path   string
	SHA256 string
	Size   int64
}

// Receipt is the strict schema-1 record for one logical cache entry.
type Receipt struct {
	SchemaVersion int
	CacheKey      CacheKey
	Input         Input
	Artifact      Artifact
}

// CacheKey is SHA-256 of the exact canonical logical input bytes.
type CacheKey string

// ReceiptHash is a deterministic consistency/corruption identifier over exact
// canonical receipt bytes. It is not a signature, attestation, authorization
// token, or proof of provenance.
type ReceiptHash string

// FixedPolicy returns the only policy admitted by go-v1.
func FixedPolicy() Policy {
	return Policy{
		ModuleMode:         policyModuleMode,
		Network:            policyNetwork,
		Workspace:          false,
		CGO:                false,
		CompilerDirectives: policyCompilerDirectives,
		TargetMode:         policyTargetMode,
		LinkMode:           policyLinkMode,
		Libgcc:             policyLibgcc,
		PackageAssembly:    false,
		HostObjects:        false,
		Telemetry:          policyTelemetry,
	}
}

// ArtifactPath derives the sole artifact path from the command and target OS.
func ArtifactPath(command, goos string) (string, error) {
	if !identifiers.Valid(command) {
		return "", fmt.Errorf("artifact command is not a portable identifier")
	}
	if !identifiers.Valid(goos) {
		return "", fmt.Errorf("artifact target GOOS is not a portable identifier")
	}
	suffix := ""
	if goos == "windows" {
		suffix = ".exe"
	}
	return "bin/" + command + suffix, nil
}

// Validate checks the complete logical input independently of filesystem state.
func (input Input) Validate() error {
	if input.SchemaVersion != SchemaVersion {
		return fmt.Errorf("build input schema_version must be %d", SchemaVersion)
	}
	if input.Driver != DriverGoV1 {
		return fmt.Errorf("unsupported build driver %q", input.Driver)
	}
	if input.BuildSource.Algorithm != buildsource.Algorithm || !validSHA256(input.BuildSource.ContentSHA256) {
		return fmt.Errorf("build_source must be a valid %s identity", buildsource.Algorithm)
	}
	if !identifiers.PortablePath(input.BuildRoot) {
		return fmt.Errorf("build_root must be a portable relative path")
	}
	if !identifiers.Valid(input.Command) {
		return fmt.Errorf("command must be a portable identifier")
	}
	if !identifiers.PortablePath(input.SourceDir) {
		return fmt.Errorf("source_dir must be a portable relative path")
	}
	if input.SourceDir != input.BuildRoot && !strings.HasPrefix(input.SourceDir, input.BuildRoot+"/") {
		return fmt.Errorf("source_dir must be contained by build_root")
	}
	if err := input.Target.validate(); err != nil {
		return err
	}
	if err := input.Toolchain.validate(); err != nil {
		return err
	}
	if input.Policy != FixedPolicy() {
		return fmt.Errorf("policy is not the fixed go-v1 policy")
	}
	return nil
}

func (target Target) validate() error {
	if !identifiers.Valid(target.GOOS) || !identifiers.Valid(target.GOARCH) {
		return fmt.Errorf("target GOOS and GOARCH must be portable identifiers")
	}
	if target.Tuning == nil {
		return fmt.Errorf("target tuning must be an object")
	}
	expectedKey := tuningKey(target.GOARCH)
	if expectedKey == "" {
		if len(target.Tuning) != 0 {
			return fmt.Errorf("target GOARCH %q has no go-v1 tuning input", target.GOARCH)
		}
		return nil
	}
	if len(target.Tuning) != 1 {
		return fmt.Errorf("target GOARCH %q requires exactly tuning key %s", target.GOARCH, expectedKey)
	}
	value, present := target.Tuning[expectedKey]
	if !present || value == "" || utf8.RuneCountInString(value) > 8192 || !utf8.ValidString(value) {
		return fmt.Errorf("target GOARCH %q requires a non-empty %s tuning value", target.GOARCH, expectedKey)
	}
	return nil
}

func tuningKey(goarch string) string {
	switch goarch {
	case "386":
		return "GO386"
	case "amd64":
		return "GOAMD64"
	case "arm":
		return "GOARM"
	case "arm64":
		return "GOARM64"
	case "mips", "mipsle":
		return "GOMIPS"
	case "mips64", "mips64le":
		return "GOMIPS64"
	case "ppc64", "ppc64le":
		return "GOPPC64"
	case "riscv64":
		return "GORISCV64"
	case "wasm":
		return "GOWASM"
	default:
		return ""
	}
}

func (toolchain Toolchain) validate() error {
	if toolchain.Algorithm != ToolchainAlgorithm || toolchain.GoRelpath != ToolchainGoRelpath {
		return fmt.Errorf("toolchain has an unsupported identity algorithm or go_relpath")
	}
	if toolchain.GoVersion == "" || utf8.RuneCountInString(toolchain.GoVersion) > 4096 ||
		strings.ContainsAny(toolchain.GoVersion, "\r\n\x00") || !utf8.ValidString(toolchain.GoVersion) {
		return fmt.Errorf("toolchain go_version is malformed")
	}
	if !validSHA256(toolchain.ContentSHA256) {
		return fmt.Errorf("toolchain content_sha256 is malformed")
	}
	return nil
}

func (artifact Artifact) validate(input Input) error {
	wantPath, err := ArtifactPath(input.Command, input.Target.GOOS)
	if err != nil {
		return err
	}
	if artifact.Path != wantPath || !identifiers.PortablePath(artifact.Path) {
		return fmt.Errorf("artifact path %q does not match manager-derived path %q", artifact.Path, wantPath)
	}
	if !validSHA256(artifact.SHA256) {
		return fmt.Errorf("artifact sha256 is malformed")
	}
	if artifact.Size < 0 || artifact.Size > MaxSafeInteger {
		return fmt.Errorf("artifact size is outside the non-negative safe integer range")
	}
	return nil
}

func validSHA256(value string) bool {
	return sha256Pattern.MatchString(value)
}

func inputValue(input Input) map[string]any {
	tuning := make(map[string]any, len(input.Target.Tuning))
	for key, value := range input.Target.Tuning {
		tuning[key] = value
	}
	return map[string]any{
		"schema_version": input.SchemaVersion,
		"driver":         input.Driver,
		"build_source": map[string]any{
			"algorithm":      input.BuildSource.Algorithm,
			"content_sha256": input.BuildSource.ContentSHA256,
		},
		"build_root": input.BuildRoot,
		"command":    input.Command,
		"source_dir": input.SourceDir,
		"target": map[string]any{
			"goos": input.Target.GOOS, "goarch": input.Target.GOARCH, "tuning": tuning,
		},
		"toolchain": map[string]any{
			"algorithm": input.Toolchain.Algorithm, "go_relpath": input.Toolchain.GoRelpath,
			"go_version": input.Toolchain.GoVersion, "content_sha256": input.Toolchain.ContentSHA256,
		},
		"policy": map[string]any{
			"module_mode": input.Policy.ModuleMode, "network": input.Policy.Network,
			"workspace": input.Policy.Workspace, "cgo": input.Policy.CGO,
			"compiler_directives": input.Policy.CompilerDirectives, "target_mode": input.Policy.TargetMode,
			"link_mode": input.Policy.LinkMode, "libgcc": input.Policy.Libgcc,
			"package_assembly": input.Policy.PackageAssembly, "host_objects": input.Policy.HostObjects,
			"telemetry": input.Policy.Telemetry,
		},
	}
}

func receiptValue(receipt Receipt) map[string]any {
	return map[string]any{
		"schema_version": receipt.SchemaVersion,
		"cache_key":      string(receipt.CacheKey),
		"input":          inputValue(receipt.Input),
		"artifact": map[string]any{
			"path": receipt.Artifact.Path, "sha256": receipt.Artifact.SHA256, "size": receipt.Artifact.Size,
		},
	}
}
