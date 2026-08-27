package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/locale"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/whitelist"
)

// nodeInstall is one node's desired installed state inside one context store.
type nodeInstall struct {
	node *closure.Node
	// store is the live directory that holds the node's context directory.
	store string
	// kind labels the store in target identifiers: project, global, or hybrid.
	kind     string
	locale   string
	agents   []string
	expected *marker.Marker
}

// stageNode renders the complete desired context directory of one node below
// the stage root and returns the target it replaces.
//
// An already current node produces no target at all: a marker match is the
// installation's idempotence, and journaling an identical replacement would
// make every re-run rewrite every skill. Its status is still reported.
func stageNode(stageRoot string, install nodeInstall, clock Clock) (staging.Plan, string, error) {
	var plan staging.Plan
	live := filepath.Join(install.store, install.node.Name)
	current, err := marker.Current(live, install.expected)
	if err != nil {
		return staging.Plan{}, "", err
	}
	if current {
		return plan, "up-to-date", nil
	}
	staged := filepath.Join(stageRoot, "context", install.kind, install.node.Name)
	if err := os.RemoveAll(staged); err != nil {
		return staging.Plan{}, "", err
	}
	if err := os.MkdirAll(staged, 0o755); err != nil {
		return staging.Plan{}, "", err
	}

	var files []string
	if install.node.ContextActive() {
		includeScripts := len(install.node.Spec.Commands) == 0
		if includeScripts {
			if _, err := os.Stat(filepath.Join(install.node.Snapshot, "scripts")); err != nil {
				includeScripts = false
			}
		}
		excludeRoots := whitelist.ContextExcludedRoots(install.node.Spec.RuntimeRoots, install.node.Spec.BuildRoots)
		files, err = whitelist.CopyContext(install.node.Snapshot, staged, includeScripts, excludeRoots)
		if err != nil {
			return staging.Plan{}, "", err
		}
		if _, err := locale.Render(install.node.Snapshot, staged, install.locale); err != nil {
			return staging.Plan{}, "", err
		}
	} else {
		files = []string{}
		install.expected.Locale = ""
		install.expected.Agents = []string{}
	}
	contentHash, err := hashing.ContentSHA256(staged, nil)
	if err != nil {
		return staging.Plan{}, "", err
	}
	install.expected.ContentSHA256 = contentHash
	install.expected.Files = files
	install.expected.InstalledAt = installedAt(clock)
	if err := marker.Write(staged, install.expected); err != nil {
		return staging.Plan{}, "", err
	}
	plan.Replace(staging.ClassContext, install.kind+"/"+install.node.Name, live, staged)
	return plan, "installed", nil
}

// contextSources maps each installed skill name to the directory its content
// currently lives in, so an adapter mirror can be staged from the replacement
// this run is about to commit rather than from a canonical path that does not
// hold it yet. A skill that is already current keeps its live directory.
func contextSources(plan staging.Plan, store string, names []string) map[string]string {
	staged := map[string]string{}
	for _, target := range plan.Targets {
		if target.Class == staging.ClassContext && filepath.Dir(target.LivePath) == store {
			staged[filepath.Base(target.LivePath)] = target.StagedPath
		}
	}
	sources := make(map[string]string, len(names))
	for _, name := range names {
		if replacement, found := staged[name]; found {
			sources[name] = replacement
			continue
		}
		sources[name] = filepath.Join(store, name)
	}
	return sources
}

// runtimeStaging is the staged executable state of one scope: commit-keyed
// script runtime trees plus the canonical launchers that reach them.
type runtimeStaging struct {
	plan staging.Plan
	// commands is the complete set of command names the scope publishes.
	commands map[string]bool
	// builds records, per node, the published build identity of each compiled
	// command so the install marker can carry it.
	builds map[string]map[string]marker.Build
}

// stageRuntimeAndShims derives every runtime tree, canonical launcher, and
// stale launcher removal of one scope without touching a live path.
//
// A compiled command resolves through the protected cache entry that the
// commit phase already published or verified, so a launcher can only ever point
// at an immutable protected artifact — never at a snapshot or private path.
func stageRuntimeAndShims(
	stageRoot, home, binDir string,
	nodes []*closure.Node,
	role runtimestore.ShimRole,
	platform string,
	commit scopeCommit,
	plannedInputs map[string]map[string]buildmeta.Input,
) (runtimeStaging, error) {
	result := runtimeStaging{commands: map[string]bool{}, builds: map[string]map[string]marker.Build{}}
	var specs []runtimestore.ShimSpec
	for _, node := range nodes {
		active := node.ActiveCommands()
		if len(active) == 0 {
			continue
		}
		pathEntries, err := runtimePathEntries(node, binDir, platform)
		if err != nil {
			return runtimeStaging{}, err
		}
		scriptCommands := activeScriptCommands(node, active)
		targets := map[string]runtimestore.RuntimeTarget{}
		if len(scriptCommands) > 0 {
			runtimePlan, err := runtimestore.PrepareScriptRuntime(stageRoot, runtimestore.ScriptRuntimeSpec{
				Home: home, SkillName: node.Name, Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
				RuntimeRoots: node.Spec.RuntimeRoots, Commands: scriptCommands, Platform: platform,
			})
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s: %w", node.Name, err)
			}
			if runtimePlan.Desired != nil {
				result.plan.Replace(staging.ClassRuntime,
					node.Name+"/"+node.Resolved.Commit,
					runtimePlan.Desired.LivePath, runtimePlan.Desired.StagedPath)
			}
			for name, target := range runtimePlan.Commands {
				targets[name] = target
			}
		}
		for _, name := range node.ActiveCommandNames() {
			command := node.Spec.Commands[name]
			if command.Type != "build" || !active[name] {
				continue
			}
			input, known := plannedInputs[node.Name][name]
			if !known {
				return runtimeStaging{}, fmt.Errorf("%s.%s: the build was not planned", node.Name, name)
			}
			key, err := input.CacheKey()
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, name, err)
			}
			hit, published := commit.artifacts[key]
			if !published {
				return runtimeStaging{}, fmt.Errorf("%s.%s: no protected cache entry for %s", node.Name, name, key)
			}
			compiled, err := runtimestore.CompiledTargetFromCache(hit, platform)
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, name, err)
			}
			targets[name] = compiled
			if result.builds[node.Name] == nil {
				result.builds[node.Name] = map[string]marker.Build{}
			}
			result.builds[node.Name][name] = marker.Build{
				Driver:         buildmeta.DriverGoV1,
				CacheKey:       key,
				ReceiptSHA256:  hit.ReceiptHash,
				ArtifactSHA256: hit.Receipt.Artifact.SHA256,
				ArtifactPath:   hit.Receipt.Artifact.Path,
			}
		}
		for _, name := range node.ActiveCommandNames() {
			target, selected := targets[name]
			if !selected {
				continue
			}
			destination, err := runtimestore.NewManagedShim(role, binDir, name, platform)
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, name, err)
			}
			specs = append(specs, runtimestore.ShimSpec{
				Destination: destination, Target: target, PathEntries: pathEntries,
			})
			result.commands[name] = true
		}
	}

	currentlyManaged, err := runtimestore.ManagedShimsIn(binDir, role, platform)
	if err != nil {
		return runtimeStaging{}, err
	}
	transition, err := runtimestore.StageShimTransition(stageRoot, specs, currentlyManaged)
	if err != nil {
		return runtimeStaging{}, err
	}
	for _, desired := range transition.Desired {
		result.plan.Replace(staging.ClassCanonicalShim, desired.Command, desired.LivePath, desired.StagedPath)
	}
	for _, removal := range transition.Removals {
		result.plan.Remove("shim/"+removal.Command, removal.LivePath)
	}
	return result, nil
}

// activeScriptCommands lists the active script commands of one node in
// command-lexical order.
func activeScriptCommands(node *closure.Node, active map[string]bool) []skillspec.Command {
	var commands []skillspec.Command
	for _, name := range node.ActiveCommandNames() {
		command := node.Spec.Commands[name]
		if command.Type == "script" && active[name] {
			commands = append(commands, command)
		}
	}
	return commands
}

// stageStaleSkillRemovals turns installed skills that the next closure does not
// keep into removal targets. The store is manager-owned and holds only context
// directories, so every visible entry is claimed — except a link or special
// file, which this store never produces and which is therefore reported as
// unexpected state rather than silently deleted.
func stageStaleSkillRemovals(store, kind string, expected map[string]bool) (staging.Plan, error) {
	var plan staging.Plan
	entries, err := os.ReadDir(store)
	if os.IsNotExist(err) {
		return plan, nil
	}
	if err != nil {
		return staging.Plan{}, err
	}
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || expected[name] {
			continue
		}
		info, err := os.Lstat(filepath.Join(store, name))
		if err != nil {
			return staging.Plan{}, err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return staging.Plan{}, fmt.Errorf(
				"stale entry %s in %s is a link or special file, which this store never installs", name, store)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		plan.Remove("skill/"+kind+"/"+name, filepath.Join(store, name))
	}
	return plan, nil
}

// plannedInputs indexes the derived build inputs of a plan by node and command.
func (plan BuildPlan) plannedInputs() map[string]map[string]buildmeta.Input {
	inputs := map[string]map[string]buildmeta.Input{}
	for _, build := range plan.builds {
		if inputs[build.skill] == nil {
			inputs[build.skill] = map[string]buildmeta.Input{}
		}
		inputs[build.skill][build.command] = build.input
	}
	return inputs
}

// referencedKeys lists every protected cache key the committed markers depend
// on, so garbage collection retains them while the transaction is in flight.
func (staged runtimeStaging) referencedKeys() []string {
	var keys []string
	for _, commands := range staged.builds {
		for _, build := range commands {
			keys = append(keys, string(build.CacheKey))
		}
	}
	return keys
}
