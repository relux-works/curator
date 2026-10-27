package skillruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// Export is one actual exposed command and its admitted provider.
type Export struct{ Name, Skill, Profile string }

// Binding connects a runtime generation to the complete lock and portable plan.
type Binding struct{ Profile, LockHash, PlanHash string }

// NameKey implements #196's portable key, without broadening export grammar.
func NameKey(name string) string {
	folded := []byte(name)
	for i, c := range folded {
		if c >= 'A' && c <= 'Z' {
			folded[i] = c + ('a' - 'A')
		}
	}
	name = string(folded)
	for {
		old := name
		for _, suffix := range []string{".exe", ".com", ".bat", ".cmd", ".ps1"} {
			if strings.HasSuffix(name, suffix) {
				name = strings.TrimSuffix(name, suffix)
				break
			}
		}
		if old == name {
			return name
		}
	}
}

// CheckNames checks the complete selected/retained set before any stager can
// replace a path. Retained owners are independently verified by Preflight.
func CheckNames(exports []Export) error {
	owners := map[string]Export{}
	for _, export := range exports {
		key := NameKey(export.Name)
		if !identifiers.Valid(export.Name) || key == "" {
			return fmt.Errorf("profile_skill_command_invalid: %q", export.Name)
		}
		if key == "curator" || strings.HasPrefix(key, "curator-") {
			return fmt.Errorf("environment_reserved_command_name: %q", export.Name)
		}
		if old, exists := owners[key]; exists {
			return fmt.Errorf("profile_skill_command_conflict: %s/%s and %s/%s claim %q", old.Profile, old.Skill, export.Profile, export.Skill, export.Name)
		}
		owners[key] = export
	}
	return nil
}

// OwnedFile is a protected routing preimage. The journal publishes this record
// with the shims. A filename or a parseable launcher alone never proves ownership.
type OwnedFile struct {
	Path    string  `json:"path"`
	Kind    string  `json:"kind"`
	Digest  string  `json:"digest"`
	Export  Export  `json:"export"`
	Binding Binding `json:"binding"`
}
type routingState struct {
	Schema  int         `json:"schema"`
	Binding Binding     `json:"binding"`
	Files   []OwnedFile `json:"files"`
}

// Inventory is private, validated state captured under admission/home exclusion.
// Retained exports are opaque: never resolve, build or rewrite them here.
type Inventory struct {
	state    routingState
	digest   string
	retained []Export
}

func statePath(home string) string { return filepath.Join(DomainDir(home), "routing.json") }

// Preflight checks the complete namespace read-only before live staging.
func (s *Session) Preflight(binding Binding, selected []Export) (*Inventory, error) {
	if err := s.Revalidate(); err != nil {
		return nil, err
	}
	if binding.Profile != s.request.Profile || binding.LockHash == "" || binding.PlanHash == "" {
		return nil, fmt.Errorf("profile_skill_plan_binding_invalid: incomplete runtime binding")
	}
	inventory := &Inventory{}
	file, err := stateread.ReadRegularFile(statePath(s.domain.Home))
	if err != nil {
		return nil, err
	}
	if file.Kind == stateread.KindPresent {
		if err := pathboundary.ValidateWithin(DomainDir(s.domain.Home), statePath(s.domain.Home)); err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(file.Bytes))
		decoder.DisallowUnknownFields()
		if err := protocoljson.Validate(file.Bytes); err != nil {
			return nil, err
		}
		if err := decoder.Decode(&inventory.state); err != nil {
			return nil, err
		}
		// Round-trip enforces exact recursive key spelling, required fields and no
		// redundant fields in this manager-owned record (encoding/json folds keys).
		canonical, err := json.Marshal(inventory.state)
		if err != nil || !bytes.Equal(canonical, file.Bytes) || inventory.state.Schema != 1 {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: invalid routing record")
		}
	}
	inventory.digest, err = transaction.DigestTarget(transaction.KindBytes, statePath(s.domain.Home))
	if err != nil {
		return nil, err
	}
	if inventory.state.Schema != 0 {
		if !identifiers.Valid(inventory.state.Binding.Profile) {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: invalid routing profile")
		}
		lock, hash, err := contextlock.Read(filepath.Join(s.domain.Home, "profiles", inventory.state.Binding.Profile, "lock.json"))
		if err != nil || lock.SchemaVersion != contextlock.SchemaVersion5 || hash != inventory.state.Binding.LockHash || lock.SkillPlanSHA256 != inventory.state.Binding.PlanHash {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: routing lock differs")
		}
	}
	for _, export := range selected {
		if export.Profile != binding.Profile {
			return nil, fmt.Errorf("profile_skill_command_conflict: another profile claims selected command")
		}
	}
	owned := map[string]OwnedFile{}
	retainedNames := map[string]bool{}
	for _, file := range inventory.state.Files {
		if _, exists := owned[file.Path]; exists {
			return nil, fmt.Errorf("profile_skill_command_conflict: duplicate owned path")
		}
		if !s.commandPath(file.Path) {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: owned path outside routing roots")
		}
		metadata, err := stateread.Lstat(file.Path)
		if err != nil || metadata.Kind != stateread.KindPresent || !metadata.Info.Mode().IsRegular() || file.Kind != staging.KindBytes {
			return nil, fmt.Errorf("environment_surface_unmanaged_conflict: invalid owned entry type")
		}
		if err := pathboundary.ValidateRouteWithOwner(filepath.Dir(file.Path), file.Path, pathboundary.DefaultOwnerLookup()); err != nil {
			return nil, err
		}
		if file.Export.Profile != "" && file.Export.Profile != inventory.state.Binding.Profile {
			return nil, fmt.Errorf("profile_skill_command_conflict: another profile claims an owned command")
		}
		if !identifiers.Valid(file.Binding.Profile) {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: owned generation missing")
		}
		source, hash, err := contextlock.Read(filepath.Join(s.domain.Home, "profiles", file.Binding.Profile, "lock.json"))
		if err != nil || hash != file.Binding.LockHash || source.SkillPlanSHA256 != file.Binding.PlanHash {
			return nil, fmt.Errorf("profile_skill_plan_binding_invalid: owned generation binding differs")
		}
		digest, err := transaction.DigestTarget(transaction.TargetKind(file.Kind), file.Path)
		if err != nil || digest != file.Digest {
			return nil, fmt.Errorf("environment_surface_unmanaged_conflict: owned preimage changed: %s", file.Path)
		}
		owned[file.Path] = file
		if file.Export.Profile == "" && file.Export.Name != "" && !retainedNames[file.Export.Name] {
			inventory.retained = append(inventory.retained, file.Export)
			retainedNames[file.Export.Name] = true
		}
	}
	// A registered private root has a closed command inventory. Unknown files,
	// directories, links and case aliases all refuse, including unselected names.
	for _, root := range []string{filepath.Join(s.domain.Home, "global", "bin"), s.domain.UserBin} {
		entries, err := stateread.ReadDir(root)
		if err != nil {
			return nil, err
		}
		if entries.Kind != stateread.KindPresent {
			return nil, fmt.Errorf("environment_surface_unmanaged_conflict: command root absent")
		}
		for _, entry := range entries.Entries {
			path := filepath.Join(root, entry.Name())
			if _, exists := owned[path]; !exists {
				return nil, fmt.Errorf("environment_surface_unmanaged_conflict: %s", path)
			}
		}
	}
	all := append(append([]Export(nil), selected...), inventory.retained...)
	if err := CheckNames(all); err != nil {
		return nil, err
	}
	if err := s.checks.Inventory(binding, all); err != nil {
		return nil, err
	}
	if err := s.checks.BasePATH(s.domain, binding, all); err != nil {
		return nil, err
	}
	return inventory, nil
}

func (s *Session) commandPath(path string) bool {
	return filepath.IsAbs(path) && filepath.Clean(path) == path && (filepath.Dir(path) == filepath.Join(s.domain.Home, "global", "bin") || filepath.Dir(path) == s.domain.UserBin)
}

// Retained returns the independently checked unrelated direct exports.
func (i *Inventory) Retained() []Export { return append([]Export(nil), i.retained...) }

// Publication contains staged targets for the *outer* profile transaction.
// No method here commits a second transaction or publishes a current pointer.
type Publication struct {
	Admitted            []string
	Plan                staging.Plan
	ReferencedBuildKeys []string
	Diagnostic          string
	Verify              func() error
}

// Bind joins the complete command transition with a lock/plan ownership record.
// Stale owned commands disappear through the same journal. Retained direct
// exports are removed from the stager's deletion set and remain byte-identical.
func (s *Session) Bind(stageRoot string, binding Binding, selected []Export, inventory *Inventory, plan staging.Plan, keys []string) (Publication, error) {
	if inventory == nil {
		return Publication{}, fmt.Errorf("profile_skill_plan_binding_invalid: command preflight required")
	}
	if err := s.Revalidate(); err != nil {
		return Publication{}, err
	}
	next := routingState{Schema: 1, Binding: binding, Files: []OwnedFile{}}
	retainedPaths := map[string]bool{}
	for _, file := range inventory.state.Files {
		if file.Export.Profile == "" && file.Export.Name != "" {
			next.Files = append(next.Files, file)
			retainedPaths[file.Path] = true
		}
	}
	byName := map[string]Export{}
	for _, export := range selected {
		byName[export.Name] = export
	}
	var filtered staging.Plan
	for _, target := range plan.Targets {
		if retainedPaths[target.LivePath] {
			if !target.Removal() {
				digest, err := transaction.DigestTarget(transaction.TargetKind(target.Kind), target.StagedPath)
				matched := false
				for _, old := range inventory.state.Files {
					if old.Path == target.LivePath && digest == old.Digest {
						matched = true
					}
				}
				if err != nil || !matched {
					return Publication{}, fmt.Errorf("profile_skill_command_conflict: retained path replacement")
				}
			}
			continue
		}
		filtered.Targets = append(filtered.Targets, target)
		if !s.commandPath(target.LivePath) || target.Removal() {
			continue
		}
		name := filepath.Base(target.LivePath)
		if name == ".curator-managed.json" {
			digest, err := transaction.DigestTarget(transaction.TargetKind(target.Kind), target.StagedPath)
			if err != nil {
				return Publication{}, err
			}
			next.Files = append(next.Files, OwnedFile{Path: target.LivePath, Kind: target.Kind, Digest: digest, Binding: binding})
			continue
		}
		export, exists := byName[name]
		if !exists { // Windows wrappers and enforced sidecars retain the source export.
			for candidate, owner := range byName {
				if name == candidate+".cmd" || name == candidate+".exe" || strings.HasPrefix(name, candidate+".") {
					export = owner
					exists = true
					break
				}
			}
		}
		if !exists {
			return Publication{}, fmt.Errorf("profile_skill_plan_binding_invalid: staged command has no owner")
		}
		digest, err := transaction.DigestTarget(transaction.TargetKind(target.Kind), target.StagedPath)
		if err != nil {
			return Publication{}, err
		}
		next.Files = append(next.Files, OwnedFile{Path: target.LivePath, Kind: target.Kind, Digest: digest, Export: export, Binding: binding})
	}
	for name := range byName {
		found := false
		for _, file := range next.Files {
			if filepath.Dir(file.Path) == filepath.Join(s.domain.Home, "global", "bin") && file.Export.Name == name {
				found = true
			}
		}
		if !found {
			return Publication{}, fmt.Errorf("profile_skill_plan_binding_invalid: selected command missing: %s", name)
		}
	}
	for _, old := range inventory.state.Files {
		if retainedPaths[old.Path] {
			continue
		}
		found := false
		for _, target := range filtered.Targets {
			if target.LivePath == old.Path {
				found = true
			}
		}
		if !found {
			return Publication{}, fmt.Errorf("profile_skill_plan_binding_invalid: stale command omitted: %s", old.Path)
		}
	}
	sort.Slice(next.Files, func(i, j int) bool { return next.Files[i].Path < next.Files[j].Path })
	payload, err := json.Marshal(next)
	if err != nil {
		return Publication{}, err
	}
	staged := filepath.Join(stageRoot, "profile-routing.json")
	if err := os.WriteFile(staged, payload, 0600); err != nil {
		return Publication{}, err
	}
	filtered.Replace("85-profile-routing", "routing", statePath(s.domain.Home), staged)
	verify := func() error {
		if err := s.Revalidate(); err != nil {
			return err
		}
		digest, err := transaction.DigestTarget(transaction.KindBytes, statePath(s.domain.Home))
		if err != nil || digest != inventory.digest {
			return fmt.Errorf("profile_skill_plan_binding_invalid: routing changed during compilation")
		}
		_, err = s.Preflight(binding, selected)
		return err
	}
	return Publication{Plan: filtered, ReferencedBuildKeys: keys, Verify: verify}, nil
}

// RetainedAt returns only already-owned forwarders at the admitted root; an
// enforced canonical launcher cannot acquire a new shell wrapper by retention.
func (i *Inventory) RetainedAt(root string) []Export {
	out := []Export{}
	for _, file := range i.state.Files {
		if file.Export.Profile == "" && file.Export.Name != "" && filepath.Dir(file.Path) == root {
			out = append(out, file.Export)
		}
	}
	return out
}
