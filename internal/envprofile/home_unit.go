package envprofile

import (
	"encoding/json"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"os"
	"path/filepath"
)

// homeUnit captures immutable authority under the operation lock. The engine
// completes the actual target/parent inventory and binds its allocated directory
// entries before the first durable journal, never from the caller's selector.
func (req *ResolveRequest) homeUnit(id string, plan *homePlan, prior *envmarker.Marker, marker []byte) (*envunit.Unit, error) {
	manager, err := envunit.ObserveManager(req.Home)
	if err != nil {
		return nil, err
	}
	lock, _, err := readLock(req.Home, req.Profile)
	if err != nil {
		return nil, err
	}
	oldCanonical, err := lock.Canonical()
	if err != nil {
		return nil, err
	}
	if req.rebindLock != nil {
		lock = req.rebindLock
	}
	canonical, err := lock.Canonical()
	if err != nil {
		return nil, err
	}
	source, err := os.ReadFile(sourcePath(req.Home, req.Profile))
	if err != nil {
		return nil, err
	}
	var closure []string
	for _, member := range lock.Members {
		raw, err := json.Marshal(member)
		if err != nil {
			return nil, err
		}
		closure = append(closure, envorigin.Digest(raw))
	}
	image := envunit.RecordImage{Kind: "file", Bytes: canonical, Digest: envorigin.Digest(canonical)}
	before := envunit.HomeImage{Kind: "absent"}
	action := envunit.Provision
	if prior != nil {
		action = envunit.Maintain
		before.Kind = "recorded"
		before.Marker, err = os.ReadFile(filepath.Join(plan.homeDir, envmarker.Name))
		if err != nil {
			return nil, err
		}
		before.Origin = append([]byte(nil), req.origin.raw...)
		if req.rebindLock != nil {
			action = envunit.Rebind
			before.Origin = append([]byte{}, req.originBefore...)
		}
	}
	root := envorigin.ManagedRoot(req.Home, req.Profile, plan.adapter.ID)
	directory, err := envunit.ObserveDirectory(root)
	if err != nil {
		return nil, err
	}
	return &envunit.Unit{Record: envunit.Version, ID: id, Manager: manager, Action: action,
		Profiles: []envunit.ProfileTransition{{Name: req.Profile, Source: envorigin.Digest(source), Closure: closure, OldLock: envunit.RecordImage{Kind: "file", Bytes: oldCanonical, Digest: envorigin.Digest(oldCanonical)}, NewLock: image}},
		Homes:    []envunit.Home{{Profile: req.Profile, Adapter: plan.adapter.ID, Mode: "managed-home", Target: plan.homeDir, Root: root, DirectoryBefore: directory, DirectoryAfter: directory, Before: before, After: envunit.HomeImage{Kind: "recorded", Marker: marker, Origin: append([]byte(nil), req.origin.raw...)}}}}, nil
}
