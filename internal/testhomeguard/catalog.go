package testhomeguard

import (
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
)

type parentSubdir struct {
	envVar string
	subdir string
}

func nativeParentEnvVar() string {
	for _, adapter := range envregistry.Registry {
		if adapter.ParentVar && adapter.EnvVar != "" {
			return adapter.EnvVar
		}
	}
	return ""
}

func nativeDirectEnvVars() []string {
	seen := map[string]bool{}
	var out []string
	for _, adapter := range envregistry.Registry {
		if adapter.ParentVar || len(adapter.HomeVariables) > 0 || adapter.EnvVar == "" {
			continue
		}
		if !seen[adapter.EnvVar] {
			seen[adapter.EnvVar] = true
			out = append(out, adapter.EnvVar)
		}
	}
	sort.Strings(out)
	return out
}

func nativeParentSubdirs() []parentSubdir {
	seen := map[parentSubdir]bool{}
	var out []parentSubdir
	add := func(envVar, subdir string) {
		if envVar == "" || subdir == "" {
			return
		}
		entry := parentSubdir{envVar: envVar, subdir: subdir}
		if !seen[entry] {
			seen[entry] = true
			out = append(out, entry)
		}
	}
	for _, adapter := range envregistry.Registry {
		if adapter.ParentVar {
			add(adapter.EnvVar, adapter.ID)
		}
		for _, pair := range adapter.HomeVariables {
			if pair[1] == "config" || pair[1] == "data" {
				add(pair[0], adapter.ID)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].envVar != out[j].envVar {
			return out[i].envVar < out[j].envVar
		}
		return out[i].subdir < out[j].subdir
	})
	return out
}

func nativeAuxiliaryEnvVars() []string {
	parent := nativeParentEnvVar()
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name == "" || name == parent || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	for _, adapter := range envregistry.Registry {
		add(adapter.EnvVar)
		for _, pair := range adapter.HomeVariables {
			add(pair[0])
		}
	}
	sort.Strings(out)
	return out
}

func nativeAllEnvVars() []string {
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	add(nativeParentEnvVar())
	for _, name := range nativeAuxiliaryEnvVars() {
		add(name)
	}
	sort.Strings(out)
	return out
}

func nativeHomeSuffixes() []string {
	seen := map[string]bool{}
	var out []string
	add := func(rel string) {
		if rel == "" {
			return
		}
		rel = filepath.ToSlash(filepath.Clean(rel))
		if !seen[rel] {
			seen[rel] = true
			out = append(out, rel)
		}
	}
	add(".agents")
	ctx := harnessapi.Context{}
	for _, adapter := range envregistry.Registry {
		if contract, err := harnessregistry.Lookup(harnessapi.ID(adapter.ID)); err == nil && contract.Capabilities().Has(harnessapi.InstructionOutput) {
			if layout, err := contract.Home(ctx); err == nil {
				for _, root := range layout.Roots().Paths() {
					add(root.String())
				}
				continue
			}
		}
		if skillDir, ok := adapters.AgentPaths[adapter.ID]; ok {
			if dir := filepath.ToSlash(filepath.Dir(filepath.FromSlash(skillDir))); dir != "." && dir != "" {
				add(dir)
				continue
			}
		}
		if adapter.ParentVar {
			add(".config/" + adapter.ID)
			continue
		}
		if len(adapter.HomeVariables) > 0 {
			for _, pair := range adapter.HomeVariables {
				switch pair[1] {
				case "config":
					add(".config/" + adapter.ID)
				case "data":
					add(".local/share/" + adapter.ID)
				}
			}
			continue
		}
		root := "." + adapter.ID
		add(root)
		for _, pt := range adapter.PassthroughFor("default") {
			if dir := filepath.ToSlash(filepath.Dir(filepath.FromSlash(pt.FileLinkTarget))); dir != "." && dir != "" {
				add(root + "/" + dir)
			}
		}
	}
	sort.Strings(out)
	return out
}

func watchedFiles() []string {
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}
	for _, name := range []string{
		"AGENTS.override.md",
		"settings.json", "settings.local.json", "config.json", "config.jsonc", "config.toml",
		"models.json", "trust.json", ".mcp.json",
		"curator-mcp.config.toml", "curator-mcp.json", ".agent-environment.json",
		"auth.json", ".credentials.json", "credentials.json", "oauth.json", "tokens.json",
	} {
		add(name)
	}
	for _, adapter := range envregistry.Registry {
		if adapter.RootTarget != "" {
			add(filepath.Base(filepath.FromSlash(adapter.RootTarget)))
		}
		for _, ch := range adapter.SystemPrompt {
			if ch.Filename != "" {
				add(ch.Filename)
			}
		}
		for _, seed := range adapter.Seeds {
			add(filepath.Base(filepath.FromSlash(seed)))
		}
		for _, shadow := range adapter.Shadows {
			add(filepath.Base(filepath.FromSlash(shadow.Path)))
		}
		if adapter.ParentVar {
			add(adapter.ID + ".json")
			add(adapter.ID + ".jsonc")
		}
	}
	sort.Strings(out)
	return out
}

func controlTarget() (rootDir, rootDoc string) {
	ctx := harnessapi.Context{}
	for _, id := range harnessregistry.IDs() {
		contract, err := harnessregistry.Lookup(id)
		if err != nil || !contract.Capabilities().Has(harnessapi.InstructionOutput) {
			continue
		}
		layout, err := contract.Home(ctx)
		if err != nil {
			continue
		}
		instructions, err := contract.Instructions(ctx, harnessapi.Monolithic)
		if err != nil {
			continue
		}
		roots := layout.Roots().Paths()
		targets := instructions.Paths().Paths()
		if len(roots) == 1 && len(targets) == 1 {
			return roots[0].String(), targets[0].String()
		}
	}
	panic("testhomeguard: no registered instruction-output adapter available for control target")
}
