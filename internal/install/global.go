package install

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/adapters"
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

	nodes, err := closure.Build(closure.Options{
		SkillsRoot:     cfg.SkillsRoot,
		Home:           home,
		AllowedSources: cfg.AllowedSources,
	}, globalManifest, map[string]devsub.Substitution{})
	if err != nil {
		result.failf("%v", err)
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
		expected := buildMarker(node, cfg.PreferredLocale, agents, activeSorted, nil, nil)
		var status string
		var installErr error
		if node.ContextActive() {
			status, installErr = installContext(skillsDir, node, cfg.PreferredLocale, expected, &result, "global")
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
