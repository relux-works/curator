package globalbins

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/staging"
)

// Forwarding is the staged user-bin mirror of one global installation.
//
// Every published launcher, every stale removal, and the ownership ledger are
// journaled transaction targets, so the mirror commits and rolls back with the
// rest of the installation. Nothing here touches a live path except creating
// the selected user bin directory itself, which is the user's own PATH entry
// and never a managed target.
type Forwarding struct {
	plan staging.Plan
	// Messages carries the same operator-facing warnings Refresh reports.
	Messages []string
	// Path is the selected user bin, empty when none is safe.
	Path string
	// Published counts launchers the commit would publish.
	Published int
}

// Plan returns the journaled forwarding targets.
func (forwarding Forwarding) Plan() staging.Plan { return forwarding.plan }

// StageForwarding derives the complete user-bin mirror without publishing it.
// The caller runs it while holding the manager-home mutation lock, because the
// ownership ledger it merges is shared machine-wide state.
func StageForwarding(
	stageRoot, home string,
	expected map[string]bool,
	platform string,
	environment map[string]string,
	userHome string,
) (Forwarding, error) {
	if platform == "" {
		platform = runtimestore.Platform()
	}
	if userHome == "" {
		userHome, _ = os.UserHomeDir()
	}
	canonicalBin := filepath.Join(home, "global", "bin")
	selection := Select(home, platform, environment, userHome)
	var forwarding Forwarding
	if selection.Path == "" {
		if countExpected(expected) == 0 {
			return forwarding, nil
		}
		if selection.Warning != "" {
			forwarding.Messages = []string{selection.Warning}
		} else {
			forwarding.Messages = []string{noSafeBinWarning(canonicalBin)}
		}
		return forwarding, nil
	}
	target := selection.Path
	// The selected directory is the user's own PATH entry, not managed state.
	// Creating it is idempotent and is not journaled.
	if err := os.MkdirAll(target, 0o755); err != nil {
		if countExpected(expected) == 0 {
			return Forwarding{}, nil
		}
		return Forwarding{Messages: []string{fmt.Sprintf(
			"global: command shims were installed in %s, but %s could not be created: %v; set %s to a writable PATH directory or use curator shell-init --install",
			canonicalBin, target, err, UserBinEnv,
		)}}, nil
	}
	forwarding.Path = target

	managed := readLedger(target)
	var currentlyManaged []runtimestore.ManagedShim
	for _, name := range expectedNames(managed) {
		published := shimPath(target, name, platform)
		canonical := shimPath(canonicalBin, name, platform)
		if expected[name] {
			continue
		}
		if !ownedTarget(published, canonical, platform) {
			forwarding.Messages = append(forwarding.Messages, fmt.Sprintf(
				"global: stale command %q was not removed from %s because the target no longer matches Curator ownership",
				name, target,
			))
			continue
		}
		stale, err := runtimestore.NewManagedShim(runtimestore.SafeForwardingShim, target, name, platform)
		if err != nil {
			return Forwarding{}, err
		}
		currentlyManaged = append(currentlyManaged, stale)
	}

	nextManaged := map[string]bool{}
	var specs []runtimestore.ShimSpec
	for _, name := range expectedNames(expected) {
		canonicalPath := shimPath(canonicalBin, name, platform)
		published := shimPath(target, name, platform)
		if unmanagedConflict(published, name, managed, canonicalPath, platform) {
			forwarding.Messages = append(forwarding.Messages, fmt.Sprintf(
				"global: command %q was not published to %s; target exists and is not managed by Curator: %s",
				name, target, published,
			))
			continue
		}
		canonicalShim, err := runtimestore.NewManagedShim(runtimestore.GlobalCanonicalShim, canonicalBin, name, platform)
		if err != nil {
			return Forwarding{}, err
		}
		launcher, err := runtimestore.ForwardingTarget(canonicalShim)
		if err != nil {
			return Forwarding{}, err
		}
		destination, err := runtimestore.NewManagedShim(runtimestore.SafeForwardingShim, target, name, platform)
		if err != nil {
			return Forwarding{}, err
		}
		specs = append(specs, runtimestore.ShimSpec{Destination: destination, Target: launcher})
		nextManaged[name] = true
	}

	transition, err := runtimestore.StageShimTransition(filepath.Join(stageRoot, "forwarding"), specs, currentlyManaged)
	if err != nil {
		return Forwarding{}, err
	}
	for _, desired := range transition.Desired {
		forwarding.plan.Replace(staging.ClassForwardingShim, desired.Command, desired.LivePath, desired.StagedPath)
	}
	for _, removal := range transition.Removals {
		forwarding.plan.Remove("forwarding-shim/"+removal.Command, removal.LivePath)
	}

	payload, err := ledgerPayload(nextManaged)
	if err != nil {
		return Forwarding{}, err
	}
	stagedLedger := filepath.Join(stageRoot, "forwarding-ledger", managedFile)
	if err := os.MkdirAll(filepath.Dir(stagedLedger), 0o700); err != nil {
		return Forwarding{}, fmt.Errorf("stage user-bin ownership ledger: %w", err)
	}
	if err := os.WriteFile(stagedLedger, payload, 0o600); err != nil {
		return Forwarding{}, fmt.Errorf("stage user-bin ownership ledger: %w", err)
	}
	forwarding.plan.Replace(staging.ClassMirrorLedger, managedFile, filepath.Join(target, managedFile), stagedLedger)

	forwarding.Published = len(nextManaged)
	if forwarding.Published > 0 {
		forwarding.Messages = append(forwarding.Messages, "global: command shims published to "+target)
	}
	return forwarding, nil
}
