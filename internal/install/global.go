package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/envfiles"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
)

// GlobalRoot returns the global scope directory under the machine home.
func GlobalRoot(home string) string { return filepath.Join(home, "global") }

// Global installs the machine-wide scope: the global Skillfile into
// global/skills with shims in global/bin and home-level adapters
// (Spec §9.2). userHome receives the adapter mirrors.
func Global(cfg *config.Config, userHome string, opts Options) Result {
	home := cfg.Home()
	result := Result{Alias: "global", Path: GlobalRoot(home), Status: "ok"}
	platform := opts.Platform
	if platform == "" {
		platform = runtimestore.Platform()
	}

	globalManifest, err := manifest.Load(GlobalRoot(home))
	if err != nil {
		result.failf("%v", err)
		return result
	}
	if globalManifest == nil {
		result.Status = "skipped"
		result.Messages = append(result.Messages, "global: Skillfile.json not found; run 'curator global init' first")
		return result
	}

	agents := globalManifest.Agents
	if len(agents) == 0 {
		agents = cfg.DefaultAgents
	}
	if unknown := adapters.UnknownAgents(agents); len(unknown) > 0 {
		result.Messages = append(result.Messages, fmt.Sprintf(
			"global: warning: unknown agent(s) ignored: %s", strings.Join(unknown, ", ")))
	}
	effectiveLocale := globalManifest.Locale
	if effectiveLocale == "" {
		effectiveLocale = cfg.PreferredLocale
	}

	nodes, err := closure.Build(closure.Options{
		SkillsRoot:     cfg.SkillsRoot,
		Home:           home,
		AllowedSources: cfg.AllowedSources,
	}, globalManifest, map[string]devsub.Substitution{})
	if err != nil {
		result.failf("%v", err)
		return result
	}
	if !validateNodes(nodes, effectiveLocale, "global", &result) {
		return result
	}
	if err := closure.DetectActiveCommandCollisions(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := checkSystemCommands(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := checkLegacySkillDependencies(nodes); err != nil {
		result.failf("%v", err)
		return result
	}
	for _, node := range nodes {
		for _, dependency := range node.Spec.Dependencies {
			if dependency.Type == "skill" {
				result.Messages = append(result.Messages, fmt.Sprintf(
					"global: %s uses dependencies.commands with type 'skill'; migrate to csk-skill.json schema v4 dependencies.skills",
					node.Name))
				break
			}
		}
	}

	auditGate := opts.AuditGate
	if auditGate == nil {
		auditGate = func(nodes []*closure.Node) ([]string, []string) {
			subjects := make([]audit.Subject, 0, len(nodes))
			for _, node := range nodes {
				subjects = append(subjects, audit.Subject{
					Name: node.Name, Source: node.Decl.Source, Git: node.Decl.Git,
					Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
					SchemaVersion: node.Spec.SchemaVersion, Capabilities: node.Spec.Capabilities,
				})
			}
			warnings, errs := audit.Gate(cfg, subjects)
			for index := range warnings {
				warnings[index] = "global: " + warnings[index]
			}
			return warnings, errs
		}
	}
	warnings, auditErrors := auditGate(nodes)
	result.Messages = append(result.Messages, warnings...)
	if len(auditErrors) > 0 {
		result.failf("%s", strings.Join(auditErrors, "; "))
		return result
	}

	movedTags := detectMovedTagsIn(filepath.Join(GlobalRoot(home), "skills"), nodes)
	if len(movedTags) > 0 {
		if opts.StrictTags {
			result.failf("%s", strings.Join(movedTags, "; "))
			return result
		}
		for _, warning := range movedTags {
			result.Messages = append(result.Messages, "global: "+warning)
		}
	}

	if opts.DryRun {
		for _, node := range nodes {
			result.Messages = append(result.Messages, fmt.Sprintf("global: %s (planned)", nodeSummary(node)))
		}
		result.Messages = append(result.Messages, "global: dry-run; no files modified")
		return result
	}

	skillsDir := filepath.Join(GlobalRoot(home), "skills")
	binDir := filepath.Join(GlobalRoot(home), "bin")
	expectedCommands := map[string]bool{}
	var contextNames []string
	expectedSkills := map[string]bool{}
	for _, node := range nodes {
		expectedSkills[node.Name] = true
		active := node.ActiveCommands()
		var activeSorted []string
		for name := range active {
			activeSorted = append(activeSorted, name)
		}
		sort.Strings(activeSorted)
		if len(active) > 0 {
			commandNames, err := installRuntime(home, binDir, node, active, platform)
			if err != nil {
				result.failf("%s: %v", node.Name, err)
				return result
			}
			for name := range commandNames {
				expectedCommands[name] = true
			}
		}
		expected := buildMarker(node, effectiveLocale, agents, activeSorted, nil, nil)
		var status string
		var installErr error
		if node.ContextActive() {
			status, installErr = installContext(skillsDir, node, effectiveLocale, expected)
			contextNames = append(contextNames, node.Name)
		} else {
			status, installErr = installMarkerOnly(skillsDir, node, expected)
		}
		if installErr != nil {
			result.failf("%s: %v", node.Name, installErr)
			return result
		}
		result.Messages = append(result.Messages, fmt.Sprintf("global: %s %s", nodeSummary(node), status))
	}

	if err := cleanupRemoved(skillsDir, expectedSkills); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := runtimestore.RemoveStaleShims(binDir, expectedCommands, platform); err != nil {
		result.failf("%v", err)
		return result
	}
	if err := envfiles.WriteGlobal(home); err != nil {
		result.failf("%v", err)
		return result
	}
	sort.Strings(contextNames)
	if err := adapters.RefreshGlobal(home, userHome, agents, contextNames, cfg.AdapterMode); err != nil {
		result.failf("%v", err)
		return result
	}
	if _, err := scopes.CollectRuntime(home); err != nil {
		result.failf("%v", err)
		return result
	}
	return result
}

// GlobalInit creates an empty global Skillfile.
func GlobalInit(home string) (string, error) {
	root := GlobalRoot(home)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return manifest.EnsureEmpty(root)
}
