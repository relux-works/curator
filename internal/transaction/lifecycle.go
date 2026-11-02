package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/stateread"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectleaserecord"
)

func (engine *Engine) planLifecycle(plan Plan) (Plan, func(), error) {
	cleanup := func() {}
	if plan.Lifecycle == nil {
		return plan, cleanup, nil
	}
	plan.Unit = cloneEnvironmentUnit(plan.Unit)
	if plan.Unit != nil && plan.Unit.Action == envunit.GlobalExtend {
		path := envunit.LeaseReferencePath(engine.home)
		raw, present, err := envorigin.ReadPrivateAuthority(engine.home, path, projectleaserecord.MaxRecord)
		if err != nil {
			return plan, cleanup, err
		}
		capture := &envunit.ReferencePreimage{Image: envunit.RecordImage{Kind: "absent"}}
		if present {
			info, err := os.Lstat(path)
			if err != nil {
				return plan, cleanup, err
			}
			capture.Image = envunit.RecordImage{Kind: "file", Bytes: raw, Digest: envorigin.Digest(raw)}
			capture.Mode = uint32(info.Mode().Perm())
		}
		expected, err := capture.InputDigest()
		if err != nil {
			return plan, cleanup, err
		}
		observed, err := DigestPath(path)
		if err != nil || observed != expected {
			return plan, cleanup, fmt.Errorf("reference capture changed: %v", err)
		}
		if prior, ok := plan.ReadInputs[path]; ok && prior != expected {
			return plan, cleanup, fmt.Errorf("reference input changed")
		}
		plan.ReadInputs = cloneReadInputs(plan.ReadInputs)
		if plan.ReadInputs == nil {
			plan.ReadInputs = map[string]string{}
		}
		plan.ReadInputs[path] = expected
		plan.Unit.LeasePreimage = capture
	}
	if err := plan.Lifecycle.Bind(plan.Unit); err != nil {
		return plan, cleanup, err
	}
	path := envlifecycle.Path(engine.home)
	raw, present, err := envorigin.ReadPrivateAuthority(engine.home, path, envlifecycle.MaxBytes)
	if err != nil {
		return plan, cleanup, err
	}
	if present != (len(plan.Lifecycle.Prior) != 0) || !bytes.Equal(raw, plan.Lifecycle.Prior) {
		return plan, cleanup, fmt.Errorf("lifecycle CAS preimage changed")
	}
	for _, target := range plan.Targets {
		if target.LivePath == path {
			return plan, cleanup, fmt.Errorf("duplicate structured lifecycle publication")
		}
	}
	file, err := privatedir.CreateTemp("", "curator-lifecycle-*")
	if err != nil {
		return plan, cleanup, err
	}
	cleanup = func() { _ = os.Remove(file.Name()) }
	if _, err = file.Write([]byte("unpublished lifecycle staging\n")); err == nil {
		err = file.Close()
	} else {
		_ = file.Close()
	}
	if err != nil {
		cleanup()
		return plan, func() {}, err
	}
	digest, err := DigestPath(path)
	if err != nil {
		cleanup()
		return plan, func() {}, err
	}
	plan.Targets = append(append([]Target{}, plan.Targets...), Target{Class: "lifecycle", Identifier: path, Kind: KindBytes, LivePath: path, StagedSource: file.Name(), PreimageDigest: digest, PrivateFile: true})
	directories, err := engine.PublicationDirectories(path)
	if err != nil {
		cleanup()
		return plan, func() {}, err
	}
	byPath := map[string]bool{}
	for _, dir := range plan.Directories {
		byPath[dir.LivePath] = true
	}
	for _, dir := range directories {
		if !byPath[dir.LivePath] {
			plan.Directories = append(plan.Directories, dir)
		}
	}
	sort.Slice(plan.Directories, func(i, j int) bool { return plan.Directories[i].LivePath < plan.Directories[j].LivePath })
	return plan, cleanup, nil
}

func finishLifecycle(journal *Journal, plan Plan) error {
	if plan.Lifecycle == nil {
		return nil
	}
	if journal.Unit.Action == "global-extend" {
		if plan.Lifecycle.Conflict == nil {
			return fmt.Errorf("global conflict capture missing")
		}
		capture := *plan.Lifecycle.Conflict
		preimages, err := json.Marshal(journal.Unit.Targets)
		if err != nil {
			return err
		}
		capture.UnitPreimages = envorigin.Digest(preimages)
		journal.Global = &envjournal.GlobalRetention{Conflict: capture, State: "planned"}
	}
	// Read only exact noncredential source records, preferring the real
	// prepared target when first-use source is not yet published. The typed
	// receipt builder independently checks every byte identity against Unit.
	sources := map[string][]byte{}
	if journal.Unit.Action == "global-extend" {
		for _, p := range journal.Unit.Profiles {
			path := filepath.Join(journal.Unit.Manager.Path, "profiles", p.Name, "source.json")
			readPath := path
			for _, target := range journal.Targets {
				if target.LivePath == path && target.Class == "profile-record" {
					readPath = target.StagedSource
				}
			}
			state, err := stateread.ReadRegularFile(readPath)
			if err != nil {
				return err
			}
			if state.Kind != stateread.KindPresent {
				return fmt.Errorf("global prepared source is absent")
			}
			sources[path] = state.Bytes
		}
	}
	raw, err := plan.Lifecycle.Build(journal.Unit, sources)
	if err != nil {
		return err
	}
	for i := range journal.Targets {
		target := &journal.Targets[i]
		if target.Class != "lifecycle" {
			continue
		}
		if err := os.WriteFile(target.StagedSource, raw, 0600); err != nil {
			return err
		}
		target.DesiredDigest, err = target.Digest(target.StagedSource)
		if err != nil {
			return err
		}
		target.StagingEntries, err = captureRemovalEntries(target.Kind, target.StagedSource, target.DesiredDigest)
		return err
	}
	return fmt.Errorf("typed lifecycle target disappeared")
}
