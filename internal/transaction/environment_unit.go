package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/relux-works/curator/internal/envjournal"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
	"path/filepath"
)

func bindEnvironmentUnit(j *Journal, skeleton *envunit.Unit) error {
	if skeleton == nil {
		if len(j.Opaque) != 0 {
			return fmt.Errorf("opaque actions lack an environment unit")
		}
		for _, directory := range j.Directories {
			if directory.EmptyMuseHome != "" {
				return fmt.Errorf("empty Muse directory lacks a complete environment unit")
			}
		}
		return nil
	}
	// The plan is caller-owned mutable memory; the durable copy never aliases it.
	raw, err := json.Marshal(skeleton)
	if err != nil {
		return err
	}
	var unit envunit.Unit
	if err := json.Unmarshal(raw, &unit); err != nil {
		return err
	}
	if unit.ID != j.TransactionID {
		return fmt.Errorf("environment unit transaction binding differs")
	}
	unit.Targets = []envunit.Target{}
	var boundaryTargets []staging.Target
	for _, target := range j.Targets {
		if target.Class == "lifecycle" {
			continue
		}
		before, err := envunit.ObserveTarget(target.LivePath, target.Class)
		if err != nil {
			return err
		}
		after := envunit.Image{Kind: "absent"}
		if target.StagedSource != "" {
			after, err = envunit.ObserveTarget(target.StagedSource, target.Class)
			if err != nil {
				return err
			}
			// Source scratch and durable staging are different entries. Logical desired
			// file/link images bind bytes/text, owner and protection, never a fake inode.
			if after.Kind == "file" || after.Kind == "link" || after.Kind == "tree" {
				after.Identity = ""
				after.Protection = ""
			}
		}
		unit.Targets = append(unit.Targets, envunit.Target{Class: target.Class, Path: target.LivePath, Before: before, After: after})
		boundaryTargets = append(boundaryTargets, staging.Target{LivePath: target.LivePath, Kind: staging.KindEntry})
	}
	for _, directory := range j.Directories {
		class := "parent"
		for i := range unit.Homes {
			if unit.Homes[i].Root == directory.LivePath {
				class = "home-directory"
				unit.Homes[i].DirectoryAfter = directory.After
			}
		}
		for i := range unit.Native {
			if unit.Native[i].Root == directory.LivePath {
				class = "home-directory"
				unit.Native[i].DirectoryAfter = directory.After
			}
		}
		unit.Targets = append(unit.Targets, envunit.Target{Class: class, Path: directory.LivePath, Before: directory.Before, After: directory.After})
		boundaryTargets = append(boundaryTargets, staging.Target{LivePath: directory.LivePath, Kind: staging.KindEntry})
	}
	for _, action := range j.Opaque {
		unit.Targets = append(unit.Targets, envunit.Target{Class: envjournal.OpaqueClass(action), Path: action.LivePath, Before: action.Before, After: envunit.Image{Kind: "absent"}})
		boundaryTargets = append(boundaryTargets, staging.Target{LivePath: action.LivePath, Kind: staging.KindEntry})
	}
	snapshot, err := (staging.Plan{Targets: boundaryTargets}).Snapshot(nil)
	if err != nil {
		return err
	}
	unit.BoundaryProof, err = snapshot.Durable()
	if err != nil {
		return err
	}
	j.Unit = &unit
	return envjournal.ValidateUnit(j, filepath.Clean(unit.Manager.Path))
}
