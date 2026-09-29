package registry

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// MirrorViewObserver compares only accepted signed views. It is scoped to one
// operation so a stale view from an earlier invocation can never be paired
// with a current response from another registry.
type MirrorViewObserver struct {
	mu          sync.Mutex
	strict      bool
	stateDir    string
	persist     bool
	registries  []Registry
	views       map[string]map[int]map[string]mirrorView
	emitted     map[string]bool
	tofuSeen    map[string]bool
	outcomes    map[string]mirrorOutcome
	diagnostics []string
}

type mirrorView struct {
	registry string
	root     string
}

type mirrorOutcome struct {
	comparison string
	logSize    int
}

// NewMirrorViewObserver constructs the optional §5.1 comparison seam. A
// strict registry policy changes diagnostic severity only; divergence never
// changes resolution or excludes a registry.
func NewMirrorViewObserver(registryPolicy string) *MirrorViewObserver {
	return NewMirrorViewObserverWithState(registryPolicy, "", false, nil)
}

// NewMirrorViewObserverWithState compares accepted views and, for mutating
// operations, persists the last comparison posture alongside each registry's
// rollback state. A read-only operation still reports diagnostics without
// changing manager state.
func NewMirrorViewObserverWithState(registryPolicy, stateDir string, persist bool, registries []Registry) *MirrorViewObserver {
	return &MirrorViewObserver{
		strict:     registryPolicy == "strict",
		stateDir:   stateDir,
		persist:    persist,
		registries: append([]Registry(nil), registries...),
		views:      map[string]map[int]map[string]mirrorView{},
		emitted:    map[string]bool{},
		tofuSeen:   map[string]bool{},
		outcomes:   map[string]mirrorOutcome{},
	}
}

func bootstrapTOFUMessage(reg Registry) string {
	return fmt.Sprintf("registry_bootstrap_tofu: registry %s (%s) fixed its high-water from first use; configure bootstrap_checkpoint to pin its initial view", reg.Name, reg.URL)
}

// ReportBootstrapTOFU records the one-time warning for a registry whose
// accepted network snapshot or first page boundary fixed its initial state.
func (o *MirrorViewObserver) ReportBootstrapTOFU(reg Registry) {
	if o == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.tofuSeen[reg.URL] {
		return
	}
	o.tofuSeen[reg.URL] = true
	o.diagnostics = append(o.diagnostics, bootstrapTOFUMessage(reg))
}

// ReportCheckpointRegression records a refused rebootstrap checkpoint while
// preserving the established high-water and continuing to use it.
func (o *MirrorViewObserver) ReportCheckpointRegression(reg Registry) {
	if o == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	message := checkpointRegressionMessage(reg)
	for _, existing := range o.diagnostics {
		if existing == message {
			return
		}
	}
	o.diagnostics = append(o.diagnostics, message)
}

// Observe adds a verified snapshot or accepted page boundary to the
// comparison. Registries without a mirror group are intentionally ignored.
func (o *MirrorViewObserver) Observe(reg Registry, snapshot parsedSnapshot) {
	if o == nil || reg.MirrorGroup == "" {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()

	bySize := o.views[reg.MirrorGroup]
	if bySize == nil {
		bySize = map[int]map[string]mirrorView{}
		o.views[reg.MirrorGroup] = bySize
	}
	views := bySize[snapshot.LogSize]
	if views == nil {
		views = map[string]mirrorView{}
		bySize[snapshot.LogSize] = views
	}
	views[reg.URL] = mirrorView{registry: reg.Name, root: snapshot.MerkleRoot}
	o.outcomes[reg.URL] = mirrorOutcome{comparison: "not-compared", logSize: snapshot.LogSize}
	if len(views) < 2 {
		return
	}

	roots := map[string][]string{}
	for _, view := range views {
		roots[view.root] = append(roots[view.root], view.registry)
	}
	if len(roots) < 2 {
		for url := range views {
			o.outcomes[url] = mirrorOutcome{comparison: "agree", logSize: snapshot.LogSize}
		}
		return
	}
	for url := range views {
		o.outcomes[url] = mirrorOutcome{comparison: "diverged", logSize: snapshot.LogSize}
	}
	registryNames := make([]string, 0, len(views))
	for _, view := range views {
		registryNames = append(registryNames, view.registry)
	}
	sort.Strings(registryNames)
	key := fmt.Sprintf("%s\x00%d\x00%s", reg.MirrorGroup, snapshot.LogSize, strings.Join(registryNames, "\x00"))
	if o.emitted[key] {
		return
	}
	o.emitted[key] = true
	severity := "warning"
	if o.strict {
		severity = "error"
	}
	o.diagnostics = append(o.diagnostics, fmt.Sprintf(
		"registry_view_divergence (%s): mirror group %s has different Merkle roots at log_size %d for registries %s",
		severity, reg.MirrorGroup, snapshot.LogSize, strings.Join(registryNames, ", "),
	))
}

// Finalize stores the last observed comparison posture for configured mirror
// registries. Failure to save posture is returned to the caller while leaving
// the already-completed comparison diagnostic intact.
func (o *MirrorViewObserver) Finalize() []string {
	if o == nil || !o.persist || o.stateDir == "" {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	var failures []string
	for _, reg := range o.registries {
		if reg.MirrorGroup == "" || len(reg.PublicKeys) == 0 {
			continue
		}
		statePath := filepath.Join(o.stateDir, "snapshot-"+urlDigest(reg.URL)+".json")
		state, exists, err := readSnapshotState(statePath)
		if err != nil {
			failures = append(failures, fmt.Sprintf("registry %s mirror posture could not be read: %v", reg.Name, err))
			continue
		}
		if !exists {
			continue
		}
		outcome, observed := o.outcomes[reg.URL]
		if !observed {
			outcome.comparison = "not-compared"
		}
		state.MirrorGroup = reg.MirrorGroup
		state.LastMirrorComparison = outcome.comparison
		state.LastMirrorComparisonLogSize = outcome.logSize
		if err := writeSnapshotState(statePath, state); err != nil {
			failures = append(failures, fmt.Sprintf("registry %s mirror posture could not be saved: %v", reg.Name, err))
		}
	}
	return failures
}

// Diagnostics returns the report-only divergence diagnostics found so far.
func (o *MirrorViewObserver) Diagnostics() []string {
	if o == nil {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]string(nil), o.diagnostics...)
}
