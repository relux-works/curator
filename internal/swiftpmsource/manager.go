package swiftpmsource

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

// Manager is the production adapter entry point. It deliberately owns only
// resolution/source closure through C4; C-family validation and compilation
// are downstream boundaries.
type Manager struct{ config Config }

// NewManager binds the concrete evaluator, resolver, broker, verifier,
// toolchain, destination, and shared trust services once.
func NewManager(config Config) (*Manager, error) {
	if config.Store == nil || config.Policy == nil || config.Evaluator == nil || config.Broker == nil || config.MirrorVerifier == nil || config.OfflineRunner == nil || config.CausalHead == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM manager authority is incomplete")
	}
	if err := validateToolchain(config.Toolchain); err != nil {
		return nil, err
	}
	if err := validateDestination(config.Destination); err != nil {
		return nil, err
	}
	return &Manager{config: config}, nil
}

// CaptureAndClose invokes the complete production capture path.
func (manager *Manager) CaptureAndClose(ctx context.Context, request Request) (*Capture, error) {
	if manager == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM manager is absent")
	}
	return CaptureAndClose(ctx, manager.config, request)
}

// ProtectedRoot returns the admitted immutable package root after a fresh
// identity recheck. The downstream interop and build adapters read C-family
// sources, headers, and module maps only through this seam; the returned tree
// is the exact intake-admitted snapshot bound by IntakeReceiptID.
func (pkg PackageEvidence) ProtectedRoot() (string, error) {
	if pkg.input.Tree == nil {
		return "", fail(CodeDerivationUnauthorized, "package %s has no admitted protected tree", pkg.Identity)
	}
	protected, err := pkg.input.Tree.ProtectedPath()
	if err != nil {
		return "", err
	}
	if pkg.evaluationSubpath == "" || pkg.evaluationSubpath == "." {
		return protected, nil
	}
	return filepath.Join(protected, filepath.FromSlash(pkg.evaluationSubpath)), nil
}

// Destination returns the exact selection destination this capture was closed
// against. Downstream stages must bind the same destination; a different one
// requires its own capture and binding.
func (capture *Capture) Destination() Destination {
	if capture == nil {
		return Destination{}
	}
	destination := capture.config.Destination
	destination.Markers = cloneMap(capture.config.Destination.Markers)
	return destination
}

// SelectionToolchain returns the exact C0 toolchain identities bound by this
// capture so a downstream stage rechecks the same bytes rather than
// rediscovering a tool.
func (capture *Capture) SelectionToolchain() Toolchain {
	if capture == nil {
		return Toolchain{}
	}
	return capture.config.Toolchain
}

// OfflineMirror is one admitted kind-preserving mirror plus the exact isolated
// mount path an offline stage must reproduce. The downstream build adapter
// mounts these read-only instead of rediscovering an origin.
type OfflineMirror struct {
	Identity, Original, Mount string
	Kind                      SourceKind
	Input                     closureexec.AdmittedInput
	ReceiptID                 closuregraph.ID
}

// RootInput returns the admitted immutable root package tree and its intake
// receipt identity. A downstream offline stage replays from exactly this tree.
func (capture *Capture) RootInput() (closureexec.AdmittedInput, closuregraph.ID, error) {
	if capture == nil || capture.rootInput.Tree == nil {
		return closureexec.AdmittedInput{}, "", fail(CodeDerivationUnauthorized, "capture has no admitted root tree")
	}
	id, err := capture.rootInput.Receipt.ID()
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	return capture.rootInput, id, nil
}

// OfflineMirrors returns every admitted mirror with its exact isolated mount
// path, sorted by package identity. The set is bijective with the root lock.
func (capture *Capture) OfflineMirrors() ([]OfflineMirror, error) {
	if capture == nil {
		return nil, fail(CodeDerivationUnauthorized, "capture is absent")
	}
	mirrors := make([]OfflineMirror, 0, len(capture.Mirrors))
	for _, mirror := range capture.Mirrors {
		mount, err := mirrorMount(mirror.Identity)
		if err != nil {
			return nil, err
		}
		if mirror.input.Tree == nil || !mirror.MirrorIntakeReceiptID.Valid() {
			return nil, failFields(CodeDependencyMirrorMissing, map[string]string{"identity": mirror.Identity}, "mirror has no admitted tree")
		}
		mirrors = append(mirrors, OfflineMirror{Identity: mirror.Identity, Original: mirror.Original, Mount: mount, Kind: mirror.LocalKind, Input: mirror.input, ReceiptID: mirror.MirrorIntakeReceiptID})
	}
	sort.Slice(mirrors, func(i, j int) bool { return mirrors[i].Identity < mirrors[j].Identity })
	return mirrors, nil
}
