package rustsource

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// MetadataView distinguishes the selection-neutral and exact active derivations.
type MetadataView string

const (
	// MetadataUnfiltered derives the conservative all-feature metadata view.
	MetadataUnfiltered MetadataView = "unfiltered"
	// MetadataActive derives the exact target/feature metadata view.
	MetadataActive MetadataView = "active"
)

// metadataInvocation is an exact C4 evidence derivation permit payload.
type metadataInvocation struct {
	View, CaptureID, Executable, CWD, ManifestPath, Target, CargoHome, CargoHomeDigest, ConfigPath, ConfigSHA256 string
	Argv                                                                                                         []string
	ConfigBytes                                                                                                  []byte
	Environment                                                                                                  map[string]string
	Toolchain                                                                                                    cargoToolchain
	Network                                                                                                      string
}

func (invocation metadataInvocation) validate() error {
	if invocation.CaptureID == "" || invocation.Executable != invocation.Toolchain.CargoPath || invocation.Network != "none" || !filepath.IsAbs(invocation.ManifestPath) || !filepath.IsAbs(invocation.CargoHome) || !filepath.IsAbs(invocation.ConfigPath) || invocation.Environment["CARGO_HOME"] != invocation.CargoHome || invocation.Environment["CARGO_NET_OFFLINE"] != "true" || invocation.ConfigSHA256 != "sha256:"+digest(invocation.ConfigBytes) || invocation.CargoHomeDigest == "" {
		return fail(CodeGraphIncomplete, "metadata invocation authority is incomplete", nil)
	}
	prefix := []string{"metadata", "--config", invocation.ConfigPath, "--format-version", "1", "--locked", "--offline", "--manifest-path", invocation.ManifestPath}
	if len(invocation.Argv) < len(prefix) {
		return fail(CodeGraphIncomplete, "metadata argv is incomplete", nil)
	}
	for i := range prefix {
		if invocation.Argv[i] != prefix[i] {
			return fail(CodeGraphIncomplete, "metadata argv was widened", nil)
		}
	}
	if invocation.View != string(MetadataUnfiltered) && invocation.View != string(MetadataActive) {
		return fail(CodeGraphIncomplete, "metadata view is unsupported", nil)
	}
	if invocation.View == string(MetadataUnfiltered) {
		if len(invocation.Argv) != len(prefix)+1 || invocation.Argv[len(prefix)] != "--all-features" || invocation.Target != "" {
			return fail(CodeGraphIncomplete, "unfiltered metadata argv was widened", nil)
		}
	} else {
		if len(invocation.Argv) < len(prefix)+2 || invocation.Argv[len(prefix)] != "--filter-platform" || invocation.Argv[len(prefix)+1] != invocation.Target || invocation.Target == "" {
			return fail(CodeGraphIncomplete, "active metadata target binding differs", nil)
		}
		for index := len(prefix) + 2; index < len(invocation.Argv); {
			switch invocation.Argv[index] {
			case "--no-default-features":
				index++
			case "--features":
				if index+1 >= len(invocation.Argv) || invocation.Argv[index+1] == "" {
					return fail(CodeGraphIncomplete, "active feature argv is incomplete", nil)
				}
				index += 2
			default:
				return fail(CodeGraphIncomplete, "active metadata argv was widened", nil)
			}
		}
	}
	return nil
}

// metadataRunner commits and executes one exact metadata derivation. It is a
// separate method family so a vendor permit cannot authorize metadata.
type metadataRunner interface {
	CommitMetadata(context.Context, metadataInvocation) (permit, error)
	RunMetadata(context.Context, permit, metadataInvocation, func() error) ([]byte, string, error)
}

// metadataRequest binds both C4 derivations to the admitted capture and C0 toolchain.
type metadataRequest struct {
	CaptureID, WorkspaceRoot, ManifestPath, CargoHome, CargoHomeDigest, ConfigPath string
	ConfigBytes                                                                    []byte
	Selection                                                                      SelectionContext
	Toolchain                                                                      cargoToolchain
	RecheckToolchain                                                               func() (cargoToolchain, error)
	Runner                                                                         metadataRunner
	NormalizeRoots                                                                 map[string]string
}

// MetadataResult retains both receipts and parsed views.
type MetadataResult struct {
	Unfiltered, Active               Metadata
	UnfilteredReceipt, ActiveReceipt string
	owner                            *managerState
	capture                          *captureState
	selection                        SelectionContext
}

// ResolvedSelection returns the requested selection with Cargo's exact active
// per-node feature vectors bound for C4/build reconciliation.
func (result MetadataResult) ResolvedSelection() SelectionContext {
	selection := result.selection
	selection.Features = append([]string{}, selection.Features...)
	selection.TargetCFG = append([]string{}, selection.TargetCFG...)
	selection.ResolvedFeatures = map[string][]string{}
	for id, features := range result.selection.ResolvedFeatures {
		selection.ResolvedFeatures[id] = append([]string(nil), features...)
	}
	return selection
}

// runPermittedMetadata derives unfiltered then exact active metadata under two
// distinct committed permits and immediate toolchain rechecks.
func runPermittedMetadata(ctx context.Context, request metadataRequest) (MetadataResult, error) {
	if err := request.Toolchain.validate(); err != nil {
		return MetadataResult{}, err
	}
	if request.Runner == nil || request.RecheckToolchain == nil {
		return MetadataResult{}, fail(CodeGraphIncomplete, "metadata runner or recheck is missing", nil)
	}
	if !filepath.IsAbs(request.CargoHome) || !filepath.IsAbs(request.ConfigPath) {
		return MetadataResult{}, fail(CodeConfigUntrusted, "metadata Cargo home/config binding is missing", nil)
	}
	if filepath.Clean(request.ConfigPath) != filepath.Join(filepath.Clean(request.CargoHome), "config.toml") {
		return MetadataResult{}, fail(CodeConfigUntrusted, "metadata config must be private CARGO_HOME/config.toml", nil)
	}
	homeDigest, err := directoryDigest(request.CargoHome)
	if err != nil {
		return MetadataResult{}, err
	}
	if request.CargoHomeDigest != "" && request.CargoHomeDigest != homeDigest {
		return MetadataResult{}, fail(CodeConfigUntrusted, "metadata Cargo home digest differs", nil)
	}
	configBytes, err := os.ReadFile(request.ConfigPath) // #nosec G304 -- absolute manager-owned config path is an explicit request field.
	if err != nil || !bytes.Equal(configBytes, request.ConfigBytes) {
		return MetadataResult{}, fail(CodeConfigUntrusted, "metadata source config differs", nil)
	}
	base := []string{"metadata", "--config", request.ConfigPath, "--format-version", "1", "--locked", "--offline", "--manifest-path", request.ManifestPath}
	binding := func(invocation *metadataInvocation) {
		invocation.CargoHome = request.CargoHome
		invocation.CargoHomeDigest = homeDigest
		invocation.ConfigPath = request.ConfigPath
		invocation.ConfigBytes = append([]byte(nil), request.ConfigBytes...)
		invocation.ConfigSHA256 = "sha256:" + digest(request.ConfigBytes)
		invocation.Environment = map[string]string{"CARGO_HOME": request.CargoHome, "CARGO_NET_OFFLINE": "true"}
	}
	unfiltered := metadataInvocation{View: string(MetadataUnfiltered), CaptureID: request.CaptureID, Executable: request.Toolchain.CargoPath, CWD: request.WorkspaceRoot, ManifestPath: request.ManifestPath, Argv: append(append([]string(nil), base...), "--all-features"), Toolchain: request.Toolchain, Network: "none"}
	binding(&unfiltered)
	features, unique := sortedUnique(request.Selection.Features)
	if !unique {
		return MetadataResult{}, fail(CodeFeatureProfileMismatch, "duplicate requested feature", nil)
	}
	activeArgv := append(append([]string(nil), base...), "--filter-platform", request.Selection.Target)
	if !request.Selection.DefaultFeatures {
		activeArgv = append(activeArgv, "--no-default-features")
	}
	if len(features) > 0 {
		activeArgv = append(activeArgv, "--features", strings.Join(features, ","))
	}
	active := metadataInvocation{View: string(MetadataActive), CaptureID: request.CaptureID, Executable: request.Toolchain.CargoPath, CWD: request.WorkspaceRoot, ManifestPath: request.ManifestPath, Target: request.Selection.Target, Argv: activeArgv, Toolchain: request.Toolchain, Network: "none"}
	binding(&active)
	invocations := []metadataInvocation{unfiltered, active}
	payloads := make([][]byte, 2)
	receipts := make([]string, 2)
	for i, invocation := range invocations {
		if err := invocation.validate(); err != nil {
			return MetadataResult{}, err
		}
		permit, err := request.Runner.CommitMetadata(ctx, invocation)
		if err != nil {
			return MetadataResult{}, err
		}
		invocationID, idErr := invocation.ID()
		if idErr != nil {
			return MetadataResult{}, idErr
		}
		if permit.ID == "" || permit.InvocationID != invocationID {
			return MetadataResult{}, fail(CodeGraphIncomplete, "metadata permit is missing", map[string]string{"view": invocation.View})
		}
		recheck := func() error {
			if checkErr := recheckCargo(request.Toolchain, request.RecheckToolchain); checkErr != nil {
				return checkErr
			}
			observedHome, homeErr := directoryDigest(request.CargoHome)
			if homeErr != nil || observedHome != homeDigest {
				return fail(CodeConfigUntrusted, "metadata Cargo home changed before use", nil)
			}
			observedConfig, configErr := os.ReadFile(request.ConfigPath)
			if configErr != nil || !bytes.Equal(observedConfig, request.ConfigBytes) {
				return fail(CodeConfigUntrusted, "metadata source config changed before use", nil)
			}
			return nil
		}
		if err = recheck(); err != nil {
			return MetadataResult{}, err
		}
		payloads[i], receipts[i], err = request.Runner.RunMetadata(ctx, permit, invocation, recheck)
		if err != nil {
			return MetadataResult{}, err
		}
		if receipts[i] == "" {
			return MetadataResult{}, fail(CodeGraphIncomplete, "metadata receipt is missing", map[string]string{"view": invocation.View})
		}
	}
	unfilteredParsed, err := ParseMetadata(payloads[0])
	if err != nil {
		return MetadataResult{}, err
	}
	activeParsed, err := ParseMetadata(payloads[1])
	if err != nil {
		return MetadataResult{}, err
	}
	if len(request.NormalizeRoots) > 0 {
		if err = normalizeMetadataPaths(&unfilteredParsed, request.NormalizeRoots); err != nil {
			return MetadataResult{}, err
		}
		if err = normalizeMetadataPaths(&activeParsed, request.NormalizeRoots); err != nil {
			return MetadataResult{}, err
		}
	}
	// Every active node must occur in the unfiltered view; target filtering may only prune.
	unfilteredIDs := make([]string, len(unfilteredParsed.Resolve))
	for i, node := range unfilteredParsed.Resolve {
		unfilteredIDs[i] = node.ID
	}
	sort.Strings(unfilteredIDs)
	for _, node := range activeParsed.Resolve {
		index := sort.SearchStrings(unfilteredIDs, node.ID)
		if index == len(unfilteredIDs) || unfilteredIDs[index] != node.ID {
			return MetadataResult{}, fail(CodeGraphIncomplete, "active metadata is not a subset of unfiltered metadata", map[string]string{"package": node.ID})
		}
	}
	return MetadataResult{Unfiltered: unfilteredParsed, Active: activeParsed, UnfilteredReceipt: receipts[0], ActiveReceipt: receipts[1]}, nil
}

func normalizeMetadataPaths(metadata *Metadata, roots map[string]string) error {
	if metadata == nil || len(roots) == 0 {
		return fail(CodeGraphIncomplete, "metadata path normalization roots are missing", nil)
	}
	normalize := func(value string) (string, error) {
		for physical, logical := range roots {
			physical = filepath.Clean(physical)
			if value == physical || contained(physical, value) {
				relative, err := filepath.Rel(physical, value)
				if err != nil {
					return "", err
				}
				return filepath.ToSlash(filepath.Join(logical, relative)), nil
			}
		}
		return "", fail(CodeGraphIncomplete, "metadata path is outside admitted workspace/vendor roots", map[string]string{"path": value})
	}
	idMap := map[string]string{}
	for packageIndex := range metadata.Packages {
		var err error
		oldID := metadata.Packages[packageIndex].ID
		metadata.Packages[packageIndex].ManifestPath, err = normalize(metadata.Packages[packageIndex].ManifestPath)
		if err != nil {
			return err
		}
		for targetIndex := range metadata.Packages[packageIndex].Targets {
			metadata.Packages[packageIndex].Targets[targetIndex].SrcPath, err = normalize(metadata.Packages[packageIndex].Targets[targetIndex].SrcPath)
			if err != nil {
				return err
			}
		}
		if metadata.Packages[packageIndex].Source == "" {
			logicalRoot := filepath.ToSlash(filepath.Dir(metadata.Packages[packageIndex].ManifestPath))
			metadata.Packages[packageIndex].ID = "path+" + logicalRoot + "#" + metadata.Packages[packageIndex].Name + "@" + metadata.Packages[packageIndex].Version
		}
		idMap[oldID] = metadata.Packages[packageIndex].ID
	}
	for nodeIndex := range metadata.Resolve {
		if normalized, ok := idMap[metadata.Resolve[nodeIndex].ID]; ok {
			metadata.Resolve[nodeIndex].ID = normalized
		}
		for dependencyIndex := range metadata.Resolve[nodeIndex].Dependencies {
			if normalized, ok := idMap[metadata.Resolve[nodeIndex].Dependencies[dependencyIndex].ID]; ok {
				metadata.Resolve[nodeIndex].Dependencies[dependencyIndex].ID = normalized
			}
		}
		sort.Slice(metadata.Resolve[nodeIndex].Dependencies, func(i, j int) bool {
			return fmt.Sprint(metadata.Resolve[nodeIndex].Dependencies[i]) < fmt.Sprint(metadata.Resolve[nodeIndex].Dependencies[j])
		})
	}
	sort.Slice(metadata.Packages, func(i, j int) bool { return metadata.Packages[i].ID < metadata.Packages[j].ID })
	sort.Slice(metadata.Resolve, func(i, j int) bool { return metadata.Resolve[i].ID < metadata.Resolve[j].ID })
	return nil
}

// ID returns the exact portable metadata authority identity.
func (invocation metadataInvocation) ID() (string, error) {
	if err := invocation.validate(); err != nil {
		return "", err
	}
	encoded, err := protocoljson.MarshalCanonical(map[string]any{"argv": stringsAny(invocation.Argv), "capture_id": invocation.CaptureID, "cargo_home": invocation.CargoHome, "cargo_home_digest": invocation.CargoHomeDigest, "config_path": invocation.ConfigPath, "config_sha256": invocation.ConfigSHA256, "cwd": invocation.CWD, "environment": invocation.Environment, "executable": invocation.Executable, "manifest_path": invocation.ManifestPath, "network": invocation.Network, "target": invocation.Target, "toolchain": toolchainValue(invocation.Toolchain), "view": invocation.View})
	if err != nil {
		return "", err
	}
	return "sha256:" + digest(append([]byte("rust-metadata-invocation-v1\x00"), encoded...)), nil
}
