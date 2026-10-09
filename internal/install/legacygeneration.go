package install

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
)

// legacyGeneration retains the §4 declaration binding outside the closed marker
// and the model-facing context. It is published in the same transaction as the
// marker. It is historical evidence only: schema-1 resolution never consumes it
// as a frozen lock. Repository locations are private acquisition hints; only
// the validated lock's commit objects can establish dependency declarations.
type legacyGeneration struct {
	Manifest     json.RawMessage   `json:"manifest"`
	Lock         json.RawMessage   `json:"lock"`
	Repositories map[string]string `json:"repositories"`
}

func legacyGenerationPath(store, name string) string {
	return filepath.Join(filepath.Dir(store), ".curator-generations", name+".json")
}

func encodeLegacyGeneration(payload []byte, lock *sourcelock.Lock, nodes []*closure.Node) ([]byte, error) {
	locked, err := json.Marshal(lock.Object())
	if err != nil {
		return nil, err
	}
	repos := make(map[string]string, len(nodes))
	for _, node := range nodes {
		repos[node.Name] = node.Repo
	}
	return json.Marshal(legacyGeneration{Manifest: payload, Lock: locked, Repositories: repos})
}

func stageLegacyGeneration(stageRoot, store, kind, name string, legacy *legacyLanePlan) (staging.Plan, error) {
	var plan staging.Plan
	if legacy == nil || legacy.Packages[name] == nil {
		return plan, nil
	}
	if len(legacy.Generation) == 0 {
		return plan, fmt.Errorf("source_selection_invalid: no installed-generation evidence for %s", name)
	}
	live := legacyGenerationPath(store, name)
	current, err := stateread.ReadRegularFile(live)
	if err != nil {
		return plan, err
	}
	if current.Kind == stateread.KindPresent && bytes.Equal(current.Bytes, legacy.Generation) {
		return plan, nil
	}
	staged := filepath.Join(stageRoot, "generations", kind, name+".json")
	if err := os.MkdirAll(filepath.Dir(staged), 0o700); err != nil {
		return plan, err
	}
	if err := os.WriteFile(staged, legacy.Generation, 0o600); err != nil {
		return plan, err
	}
	plan.Replace(staging.ClassContext, "generation/"+kind+"/"+name, live, staged)
	return plan, nil
}

// previousLegacyRef authenticates the historical selection against the marker,
// not against today's tags or today's declaring manifest. Missing, unreadable,
// or forged evidence cannot establish either tag movement or a declaration
// change and is reported as such, never converted into permission to publish.
func previousLegacyRef(store string, recorded *marker.Marker, observed *observations) (manifest.Ref, error) {
	path := legacyGenerationPath(store, recorded.Name)
	file, err := stateread.ReadRegularFile(path)
	if err != nil {
		return manifest.Ref{}, err
	}
	if file.Kind == stateread.KindAbsent {
		return manifest.Ref{}, stateread.AbsentError(path)
	}
	observed.observeDocument(documentKey("legacy-generation/"+path), path, digestDeclaration(file.Bytes))
	invalid := func(err error) (manifest.Ref, error) {
		return manifest.Ref{}, stateread.UnusableError(path, err)
	}
	if err := protocoljson.Validate(file.Bytes); err != nil {
		return invalid(err)
	}
	var generation legacyGeneration
	decoder := json.NewDecoder(bytes.NewReader(file.Bytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&generation); err != nil {
		return invalid(err)
	}
	lock, err := sourcelock.Parse(generation.Lock)
	if err != nil {
		return invalid(err)
	}
	if lock.LockSHA256 != recorded.LockSHA256 {
		return invalid(fmt.Errorf("installed-generation lock does not match marker lock_sha256"))
	}
	if err := lock.CheckStale(generation.Manifest); err != nil {
		return invalid(err)
	}
	declaring, err := manifest.ParseBytes(generation.Manifest, path)
	if err != nil {
		return invalid(err)
	}
	member, ok := lock.Find(recorded.Name)
	if !ok || !reflect.DeepEqual(draftMarkerPackage(member.Package), recorded.Package) {
		return invalid(fmt.Errorf("installed-generation member does not match marker package"))
	}
	if member.Selection != nil {
		index := *member.Selection
		if index < 0 || index >= len(declaring.Skills) || declaring.Skills[index].Name != recorded.Name {
			return invalid(fmt.Errorf("installed-generation selection does not name %s", recorded.Name))
		}
		return declaring.Skills[index].Ref, nil
	}
	ref, err := generation.dependencyRef(lock, declaring, recorded.Name)
	if err != nil {
		return invalid(err)
	}
	return ref, nil
}

// A transitive member has no root selection. Replay only declaration traversal
// in closure.Build's order, reading manifests from the prior locked Git objects.
// Never resolve their old refs: a moved/deleted tag must not rewrite history.
func (generation legacyGeneration) dependencyRef(lock *sourcelock.Lock, declaring *manifest.Manifest, name string) (manifest.Ref, error) {
	queue := append([]manifest.Decl(nil), declaring.Skills...)
	seen := map[string]bool{}
	for len(queue) > 0 {
		decl := queue[0]
		queue = queue[1:]
		if seen[decl.Name] {
			continue
		}
		seen[decl.Name] = true
		if decl.Name == name {
			return decl.Ref, nil
		}
		member, ok := lock.Find(decl.Name)
		if !ok {
			return manifest.Ref{}, fmt.Errorf("installed-generation member %s missing", decl.Name)
		}
		spec, err := generation.lockedSpec(member)
		if err != nil {
			return manifest.Ref{}, err
		}
		names := make([]string, 0, len(spec.Requirements))
		for dependency := range spec.Requirements {
			names = append(names, dependency)
		}
		sort.Strings(names)
		for _, dependency := range names {
			req := spec.Requirements[dependency]
			queue = append(queue, manifest.Decl{Name: req.Name, Ref: manifest.Ref{Kind: req.RefKind, Value: req.RefValue}})
		}
	}
	return manifest.Ref{}, fmt.Errorf("installed-generation declaration for %s missing", name)
}

func (generation legacyGeneration) lockedSpec(member sourcelock.Member) (*skillspec.Spec, error) {
	repo := generation.Repositories[member.Name]
	if repo == "" || !member.Package.IsGit() {
		return nil, fmt.Errorf("installed-generation repository for %s unavailable", member.Name)
	}
	scratch, err := os.MkdirTemp("", "curator-prior-declaration-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	if err := gitops.Extract(repo, member.Package.Commit.Hex, scratch); err != nil {
		return nil, err
	}
	selected := filepath.Join(scratch, filepath.FromSlash(member.Directory))
	spec, err := skillspec.Load(selected)
	if err != nil {
		return nil, err
	}
	content, err := closure.ContentHashFor(selected, spec)
	if err != nil {
		return nil, err
	}
	if content != member.ContentSHA256 {
		return nil, fmt.Errorf("installed-generation content for %s does not match lock", member.Name)
	}
	return spec, nil
}

func sameLegacyPackageSource(a, b *marker.Package) bool {
	return a != nil && b != nil && a.Kind == b.Kind && a.Repository == b.Repository &&
		a.Source == b.Source && a.Directory == b.Directory
}
