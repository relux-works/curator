package envprofile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

func museNativeHome() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "muse"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "muse"), nil
}

// Seeds come only from the locked root snapshot, never from native credentials
// or overlays. Store validation and the profile audit precede this read.
func (req *ResolveRequest) museSeedRoot() (string, error) {
	_, lock, _, err := loadResolveInputs(req.Home, req.Profile)
	if err != nil {
		return "", err
	}
	for _, member := range lock.Members {
		if member.Kind == "context" && member.Name == lock.Root && !member.Overlay {
			return packageRoot(newGitManager(req.Home).entryPath(req.Home, resolvedOf(member)), member.Directory), nil
		}
	}
	return "", fmt.Errorf("%s: root snapshot is missing", envregistry.DiagSeedUnreadable)
}

type museAuthState struct {
	live bool
	err  error
}

// inspectMuseAuth never reads credential bytes. A fork is a conflict and a
// failed observation is unreadable, while a missing entry can be re-linked only
// after establishing a readable regular native target. Repair preflights this
// before any managed-home writes; verification calls it on every resolve.
func (req *ResolveRequest) inspectMuseAuth(full, target string) museAuthState {
	unreadable := func(err error) museAuthState {
		return museAuthState{err: fmt.Errorf("%s: %s: %v", envregistry.DiagPassthroughUnreadable, full, err)}
	}
	conflict := func(reason string) museAuthState {
		return museAuthState{err: fmt.Errorf("%s: %s: %s", envregistry.DiagCredentialConflict, full, reason)}
	}
	// Check the managed route before inspecting auth liveness, including on
	// the current-home fast path. Only the final credential entry may link
	// outside the home. The E5 parent check allows absent provisioning paths
	// without creating them or following existing parent links.
	root := EnvRoot(req.Home)
	parentRel, err := managedRelative(root, filepath.Dir(full))
	if err != nil {
		return unreadable(err)
	}
	if err := checkManagedPrivateTarget(root, parentRel); err != nil {
		return unreadable(err)
	}
	var entry stateread.Metadata
	if req.passthroughLstat == nil {
		entry, err = stateread.Lstat(full)
	} else {
		entry, err = stateread.LstatWith(full, req.passthroughLstat)
	}
	if err != nil {
		return unreadable(err)
	}
	live := false
	if entry.Kind == stateread.KindPresent {
		if entry.Info == nil {
			return unreadable(fmt.Errorf("missing metadata"))
		}
		switch {
		case entry.Info.Mode()&os.ModeSymlink != 0:
			var link stateread.Link
			if req.passthroughReadlink == nil {
				link, err = stateread.Readlink(full)
			} else {
				link, err = stateread.ReadlinkWith(full, req.passthroughReadlink)
			}
			if err != nil {
				return unreadable(err)
			}
			if link.Kind != stateread.KindPresent {
				return unreadable(fmt.Errorf("link changed during inspection"))
			}
			got := link.Target
			if !filepath.IsAbs(got) {
				got = filepath.Join(filepath.Dir(full), got)
			}
			if filepath.Clean(got) != filepath.Clean(target) {
				return conflict("unexpected credential link target")
			}
			live = true
		case entry.Info.IsDir():
			directory, err := stateread.ReadDir(full)
			if err != nil {
				return unreadable(err)
			}
			if len(directory.Entries) != 0 {
				return conflict("non-empty credential directory")
			}
		default:
			return conflict("credential link replaced by a file; preserve both stores")
		}
	} else if entry.Kind != stateread.KindAbsent {
		return unreadable(fmt.Errorf("unknown entry state %q", entry.Kind))
	}
	metadata, err := stateread.StatWith(target, req.museTargetStat)
	if err != nil {
		return unreadable(err)
	}
	if metadata.Kind == stateread.KindAbsent {
		return conflict("native credential target is absent")
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
		return unreadable(fmt.Errorf("native target metadata unavailable"))
	}
	if !metadata.Info.Mode().IsRegular() {
		return conflict("native credential target is not a regular file")
	}
	if err := stateread.CheckReadableRegularFile(target); err != nil {
		return unreadable(err)
	}
	return museAuthState{live: live}
}

// managedAdapterParticipates keeps Muse's explicit provisioning boundary in
// global reconciliation. An unreadable marker is never treated as absent.
func managedAdapterParticipates(home, profile string, adapter envregistry.Adapter) (bool, error) {
	if adapter.ID != envregistry.Muse {
		return true, nil
	}
	metadata, err := stateread.Lstat(filepath.Join(ManagedHomeDir(home, profile, adapter.ID), envmarker.Name))
	if err != nil {
		return false, err
	}
	return metadata.Kind == stateread.KindPresent, nil
}
