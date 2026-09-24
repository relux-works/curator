package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/locale"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/scriptworker"
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

// nodeSnapshots returns the resolved content snapshots of every closure
// node: the admitted inputs the publication boundary guard must protect.
// Staging already reads these directories (context projection and runtime
// trees consume snapshot bytes), so a node without one cannot stage and
// fails here with the member diagnostic instead of staging blind.
func nodeSnapshots(nodes []*closure.Node) ([]string, error) {
	admitted := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node == nil || node.Snapshot == "" {
			name := "<nil>"
			if node != nil {
				name = node.Name
			}
			return nil, fmt.Errorf("source_member_missing: node %s has no content snapshot", name)
		}
		admitted = append(admitted, node.Snapshot)
	}
	return admitted, nil
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
	// enforced names the published commands served by native enforced
	// launchers rather than ordinary shims.
	enforced map[string]bool
	// messages carries the per-command derivation records install reports.
	messages []string
	// builds records, per node, the published build identity of each compiled
	// command so the install marker can carry it.
	builds map[string]map[string]marker.Build
}

// stageRuntimeAndShims derives every runtime tree, canonical launcher, and
// stale launcher removal of one scope without touching a live path.
//
// runtimeKeys optionally overrides the commit-keyed runtime leaf per node:
// the draft lane passes frozen source-v1 keys for its local-snapshot
// members (which carry no commit) so they materialize under their package
// identity. A nil map keeps the resolved-commit behavior byte-identically.
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
	plannedInputs map[string]map[string]plannedBuildInput,
	external map[string]map[string]externalEntry,
	externalRoot string,
	runtimeKeys map[string]string,
	projectRoot string,
) (runtimeStaging, error) {
	result := runtimeStaging{
		commands: map[string]bool{},
		enforced: map[string]bool{},
		builds:   map[string]map[string]marker.Build{},
	}
	var specs []runtimestore.ShimSpec
	var enforcedSpecs []runtimestore.EnforcedShimSpec
	for _, node := range nodes {
		active := node.ActiveCommands()
		if len(active) == 0 {
			continue
		}
		pathEntries, err := runtimePathEntries(node, binDir, platform)
		if err != nil {
			return runtimeStaging{}, err
		}
		scriptCommands, err := activeScriptCommands(node, active)
		if err != nil {
			return runtimeStaging{}, err
		}
		enforcedCommands, err := activeEnforcedScriptCommands(node, active)
		if err != nil {
			return runtimeStaging{}, err
		}
		targets := map[string]runtimestore.RuntimeTarget{}
		scriptTargets := map[string]runtimestore.ScriptTarget{}
		allScriptCommands := append(append([]skillspec.Command(nil), scriptCommands...), enforcedCommands...)
		if len(allScriptCommands) > 0 {
			runtimeKey := draftRuntimeKey(node, runtimeKeys)
			runtimePlan, err := runtimestore.PrepareScriptRuntime(stageRoot, runtimestore.ScriptRuntimeSpec{
				Home: home, SkillName: node.Name, Commit: runtimeKey, Snapshot: node.Snapshot,
				RuntimeRoots: node.Spec.RuntimeRoots, Commands: allScriptCommands, Platform: platform,
			})
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s: %w", node.Name, err)
			}
			if runtimePlan.Desired != nil {
				result.plan.Replace(staging.ClassRuntime,
					node.Name+"/"+runtimeKey,
					runtimePlan.Desired.LivePath, runtimePlan.Desired.StagedPath)
			}
			for name, target := range runtimePlan.Commands {
				targets[name] = target
				scriptTargets[name] = target
			}
		}
		for _, name := range node.ActiveCommandNames() {
			command := node.Spec.Commands[name]
			if command.Type != "build" || !active[name] {
				continue
			}
			if command.Driver == "go-repository-v1" {
				entry, known := external[node.Name][name]
				if !known {
					return runtimeStaging{}, fmt.Errorf("%s.%s: the external build was not staged", node.Name, name)
				}
				keyName := strings.TrimPrefix(entry.result.CacheKey, "sha256:")
				finalArtifact := filepath.Join(externalRoot, buildrepo.ArtifactsDir(entry.result.ReceiptSchemaVersion), keyName, "artifact")
				compiled, err := runtimestore.ExternalCompiledTarget(entry.artifactPath, finalArtifact,
					entry.record.CacheKey, entry.record.ReceiptSHA256, entry.record.ArtifactSHA256, platform)
				if err != nil {
					return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, name, err)
				}
				targets[name] = compiled
				if result.builds[node.Name] == nil {
					result.builds[node.Name] = map[string]marker.Build{}
				}
				result.builds[node.Name][name] = entry.record
				continue
			}
			planned, known := plannedInputs[node.Name][name]
			if !known {
				return runtimeStaging{}, fmt.Errorf("%s.%s: the build was not planned", node.Name, name)
			}
			key := planned.key
			hit, published := commit.artifacts[key]
			if !published {
				return runtimeStaging{}, failedBuildBoundary(
					"%s.%s: no protected cache entry for %s", node.Name, name, key)
			}
			// The cache result carries the manager-private absolute artifact path,
			// so a refusal derived from it is a build-cache boundary failure and is
			// rendered through the same bounded, redacted path as every other one.
			compiled, err := runtimestore.CompiledTargetFromCache(hit, platform)
			if err != nil {
				return runtimeStaging{}, failedBuildBoundary("%s.%s: %w", node.Name, name, err)
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
		enforced := map[string]bool{}
		for _, command := range enforcedCommands {
			enforced[command.Name] = true
		}
		for _, name := range node.ActiveCommandNames() {
			if enforced[name] {
				continue
			}
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
		if len(enforcedCommands) != 0 {
			declaredRaw, err := rawDeclaredCapabilities(node.Snapshot)
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s: %w", node.Name, err)
			}
			declared, err := scriptworker.ParseDeclaredCapabilities(declaredRaw)
			if err != nil {
				return runtimeStaging{}, fmt.Errorf("%s: %w", node.Name, err)
			}
			for _, command := range enforcedCommands {
				target, selected := scriptTargets[command.Name]
				if !selected {
					return runtimeStaging{}, fmt.Errorf("%s.%s: the enforced runtime target was not staged", node.Name, command.Name)
				}
				destination, err := runtimestore.NewManagedEnforcedShim(role, binDir, command.Name, platform)
				if err != nil {
					return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, command.Name, err)
				}
				sidecar, err := scriptworker.NewShimSidecar(
					node.Name, command.Name, command.Interpreter,
					target.ExecutablePath(), target.RuntimeDir(),
					projectRoot, node.Spec.SchemaVersion, declaredRaw)
				if err != nil {
					return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, command.Name, err)
				}
				sidecarPayload, err := sidecar.Marshal()
				if err != nil {
					return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, command.Name, err)
				}
				enforcedSpecs = append(enforcedSpecs, runtimestore.EnforcedShimSpec{
					Destination: destination, Sidecar: sidecarPayload,
				})
				result.commands[command.Name] = true
				result.enforced[command.Name] = true
				reports, err := enforcedDerivationMessages(node.Name, command, declared, home, binDir, node.Snapshot)
				if err != nil {
					return runtimeStaging{}, fmt.Errorf("%s.%s: %w", node.Name, command.Name, err)
				}
				result.messages = append(result.messages, reports...)
			}
		}
	}

	currentlyManaged, err := runtimestore.ManagedShimsIn(binDir, role, platform)
	if err != nil {
		return runtimeStaging{}, err
	}
	currentlyEnforced, err := runtimestore.ManagedEnforcedShimsIn(binDir, role, platform)
	if err != nil {
		return runtimeStaging{}, err
	}
	// Sidecar-paired launchers and the sidecars themselves belong to the
	// enforced transition, so the ordinary transition must not also claim
	// or remove them. (A sidecar filename parses as an ordinary shim
	// name, which is exactly why the exclusion is explicit.)
	enforcedPaths := map[string]bool{}
	for _, shim := range currentlyEnforced {
		enforcedPaths[shim.Path()] = true
		enforcedPaths[shim.SidecarPath()] = true
	}
	ordinaryCurrent := currentlyManaged[:0]
	for _, shim := range currentlyManaged {
		if !enforcedPaths[shim.Path()] {
			ordinaryCurrent = append(ordinaryCurrent, shim)
		}
	}
	transition, err := runtimestore.StageShimTransition(stageRoot, specs, ordinaryCurrent)
	if err != nil {
		return runtimeStaging{}, err
	}
	enforcedTransition, err := runtimestore.StageEnforcedShimTransition(stageRoot, enforcedSpecs, currentlyEnforced)
	if err != nil {
		return runtimeStaging{}, err
	}
	// The two transitions are computed independently, so an enforcement
	// flip emits a replacement from one side and a removal of the same
	// live launcher from the other. The replacement always supersedes
	// the removal: the commit overwrites the live path, and only the
	// orphaned sidecar still needs its own removal.
	replaced := map[string]bool{}
	for _, desired := range transition.Desired {
		replaced[desired.LivePath] = true
	}
	for _, desired := range enforcedTransition.Desired {
		replaced[desired.LivePath] = true
	}
	for _, desired := range transition.Desired {
		result.plan.Replace(staging.ClassCanonicalShim, desired.Command, desired.LivePath, desired.StagedPath)
	}
	for _, removal := range transition.Removals {
		if replaced[removal.LivePath] {
			continue
		}
		result.plan.Remove("shim/"+removal.Command, removal.LivePath)
	}
	for _, desired := range enforcedTransition.Desired {
		identifier := enforcedPlanIdentifier(desired.Command, desired.LivePath)
		result.plan.Replace(staging.ClassCanonicalShim, identifier, desired.LivePath, desired.StagedPath)
	}
	for _, removal := range enforcedTransition.Removals {
		if replaced[removal.LivePath] {
			continue
		}
		result.plan.Remove(enforcedPlanIdentifier(removal.Command, removal.LivePath), removal.LivePath)
	}
	return result, nil
}

// enforcedPlanIdentifier names one enforced launcher or sidecar target in
// the transaction plan. The launcher and its sidecar share a command but
// are distinct live paths, so they need distinct identifiers: one shared
// identifier duplicates the removal target whenever both removals land,
// which is exactly the Windows enforcement flip — the `.exe` launcher and
// the `.cmd` shim are different paths, so no replacement suppresses
// either removal. Desired and removal sides share this spelling so one
// flip cannot collide with itself in either class.
func enforcedPlanIdentifier(command, livePath string) string {
	if strings.HasSuffix(livePath, ".curator-shim.json") {
		return "enforced-sidecar/" + command
	}
	return "enforced-shim/" + command
}

// activeScriptCommands lists the active declared-only script commands of
// one node in command-lexical order. Enforced commands are not returned
// here; they stage through activeEnforcedScriptCommands below.
func activeScriptCommands(node *closure.Node, active map[string]bool) ([]skillspec.Command, error) {
	var commands []skillspec.Command
	for _, name := range node.ActiveCommandNames() {
		command := node.Spec.Commands[name]
		if command.Type != "script" || !active[name] {
			continue
		}
		if scriptpolicy.Enforced(command) {
			continue
		}
		commands = append(commands, command)
	}
	return commands, nil
}

// activeEnforcedScriptCommands lists the active enforced script commands of
// one node in command-lexical order. Admission refuses an unknown policy
// and an incomplete control table; the host preflight then refuses with
// `script_execution_control_unavailable` when this host cannot provide a
// mandatory control, before anything is staged — rather than writing the
// ordinary uncontained shim §4.1.1 forbids or publishing a launcher that
// can never run. The host probe runs once per call, not once per command:
// the host cannot differ between two commands of one install. Once both
// gates pass the commands proceed to native launcher staging — the guard
// never formats a nil admission, so it cannot degrade to a wrapped nil.
func activeEnforcedScriptCommands(node *closure.Node, active map[string]bool) ([]skillspec.Command, error) {
	var commands []skillspec.Command
	for _, name := range node.ActiveCommandNames() {
		command := node.Spec.Commands[name]
		if command.Type != "script" || !active[name] {
			continue
		}
		if !scriptpolicy.Enforced(command) {
			continue
		}
		if err := scriptpolicy.Admit(map[string]skillspec.Command{name: command}); err != nil {
			return nil, fmt.Errorf("%s.%s: %w", node.Name, name, err)
		}
		commands = append(commands, command)
	}
	if len(commands) != 0 {
		if err := scriptworker.PreflightHostControls(); err != nil {
			return nil, fmt.Errorf("%s: %w", node.Name, err)
		}
	}
	return commands, nil
}

// rawDeclaredCapabilities re-reads the declared `capabilities` object of
// one node snapshot. Derivation reads the declared manifest bytes rather
// than the parsed schema defaults, so the sidecar stores the raw object
// and the launcher derives presence from it. The bytes are re-validated
// both here and at launch.
func rawDeclaredCapabilities(snapshot string) (json.RawMessage, error) {
	manifestPath := filepath.Join(snapshot, skillspec.ManifestSourcePath(snapshot))
	payload, err := os.ReadFile(manifestPath) // #nosec G304 -- admitted node snapshot manifest
	if err != nil {
		return nil, fmt.Errorf("cannot read the declared capabilities: %w", err)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, fmt.Errorf("cannot decode the declared capabilities: %w", err)
	}
	raw, present := envelope["capabilities"]
	if !present {
		return nil, nil
	}
	return raw, nil
}

// enforcedDerivationMessages records one enforced command's install-time
// derivation: the breaking-change notice plus every withheld env_read
// entry, recorded network host, and unresolvable exec name. The per-command
// profile itself derives fresh at every invocation; this record describes
// the install moment.
func enforcedDerivationMessages(skill string, command skillspec.Command, declared scriptworker.DeclaredCapabilities, home, binDir, snapshot string) ([]string, error) {
	report, err := scriptworker.DeriveStaticReport(declared, command.Interpreter, nil,
		[]string{snapshot, home, binDir})
	if err != nil {
		return nil, err
	}
	prefix := skill + ": command '" + command.Name + "' is enforced (script-worker-v1): "
	messages := []string{prefix +
		"undeclared environment variables are absent and undeclared executables do not resolve at run time"}
	if len(report.WithheldEnv) != 0 {
		messages = append(messages, prefix+"withholds manager-owned env_read entries: "+
			strings.Join(report.WithheldEnv, ", "))
	}
	if len(report.RecordedHosts) != 0 {
		messages = append(messages, prefix+"declares network hosts, recorded reporting-only without filtering: "+
			strings.Join(report.RecordedHosts, ", "))
	}
	if len(report.UnresolvedExec) != 0 {
		messages = append(messages, prefix+"omits unresolvable exec names from PATH: "+
			strings.Join(report.UnresolvedExec, ", "))
	}
	return messages, nil
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
type plannedBuildInput struct {
	key buildmeta.CacheKey
}

func (plan BuildPlan) plannedInputs() map[string]map[string]plannedBuildInput {
	inputs := map[string]map[string]plannedBuildInput{}
	for _, build := range plan.builds {
		if inputs[build.skill] == nil {
			inputs[build.skill] = map[string]plannedBuildInput{}
		}
		inputs[build.skill][build.command] = plannedBuildInput{key: build.key}
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
