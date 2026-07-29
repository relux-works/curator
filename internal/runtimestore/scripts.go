package runtimestore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/skillspec"
)

// ScriptRuntimeSpec is the complete active script runtime needed at one skill
// commit. Build roots and build commands are intentionally not representable.
type ScriptRuntimeSpec struct {
	Home         string
	SkillName    string
	Commit       string
	Snapshot     string
	RuntimeRoots []string
	Commands     []skillspec.Command
	Platform     string
}

// ScriptRuntimePlan selects validated live command targets and optionally a
// complete staged replacement for a missing or incomplete runtime directory.
type ScriptRuntimePlan struct {
	Commands    map[string]ScriptTarget
	Desired     *DesiredTarget
	Replacement bool
}

// PrepareScriptRuntime validates every required runtime root and active script
// path before reuse. Missing or incomplete live state is left untouched and a
// complete operation-private replacement is staged for the transaction layer.
func PrepareScriptRuntime(stageRoot string, spec ScriptRuntimeSpec) (ScriptRuntimePlan, error) {
	stageRoot, err := cleanAbsolute(stageRoot)
	if err != nil {
		return ScriptRuntimePlan{}, fmt.Errorf("runtime staging root: %w", err)
	}
	if _, err := cleanAbsolute(spec.Home); err != nil {
		return ScriptRuntimePlan{}, fmt.Errorf("manager home: %w", err)
	}
	if _, err := cleanAbsolute(spec.Snapshot); err != nil {
		return ScriptRuntimePlan{}, fmt.Errorf("snapshot: %w", err)
	}
	if !identifiers.Valid(spec.SkillName) || !identifiers.Valid(spec.Commit) {
		return ScriptRuntimePlan{}, fmt.Errorf("script runtime skill and commit must be portable identifiers")
	}
	if spec.Platform != "unix" && spec.Platform != "windows" {
		return ScriptRuntimePlan{}, fmt.Errorf("unsupported runtime platform %q", spec.Platform)
	}
	roots, commands, err := validateScriptSpec(spec)
	if err != nil {
		return ScriptRuntimePlan{}, err
	}
	live := Dir(spec.Home, spec.SkillName, spec.Commit)
	if pathsOverlap(stageRoot, live) {
		return ScriptRuntimePlan{}, fmt.Errorf("runtime staging root overlaps live runtime %s", live)
	}
	plan := ScriptRuntimePlan{Commands: scriptTargets(live, roots, commands, spec.Platform)}
	if err := validateRuntimeTree(live, roots, commands, spec.Platform); err == nil {
		return plan, nil
	} else if _, statErr := os.Lstat(live); statErr == nil {
		plan.Replacement = true
	} else if !os.IsNotExist(statErr) {
		return ScriptRuntimePlan{}, fmt.Errorf("inspect script runtime: %w", statErr)
	}

	staged := filepath.Join(stageRoot, "runtime", spec.SkillName, spec.Commit)
	if err := os.RemoveAll(staged); err != nil {
		return ScriptRuntimePlan{}, fmt.Errorf("reset private runtime staging: %w", err)
	}
	if err := stageScriptTree(staged, spec.Snapshot, roots, commands, spec.Platform); err != nil {
		_ = os.RemoveAll(staged)
		return ScriptRuntimePlan{}, err
	}
	if err := validateRuntimeTree(staged, roots, commands, spec.Platform); err != nil {
		_ = os.RemoveAll(staged)
		return ScriptRuntimePlan{}, fmt.Errorf("validate staged script runtime: %w", err)
	}
	plan.Desired = &DesiredTarget{Kind: RuntimeTreeTarget, RuntimeKind: ScriptRuntime, LivePath: live, StagedPath: staged}
	return plan, nil
}

func validateScriptSpec(spec ScriptRuntimeSpec) ([]string, []skillspec.Command, error) {
	roots := append([]string(nil), spec.RuntimeRoots...)
	sort.Strings(roots)
	for index, root := range roots {
		if !identifiers.PortablePath(root) {
			return nil, nil, fmt.Errorf("runtime root %q is not a portable relative path", root)
		}
		if index > 0 && roots[index-1] == root {
			return nil, nil, fmt.Errorf("runtime roots must be unique")
		}
	}
	commands := append([]skillspec.Command(nil), spec.Commands...)
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	for index, command := range commands {
		if command.Type != "script" || !identifiers.Valid(command.Name) {
			return nil, nil, fmt.Errorf("active runtime command %q is not a script command", command.Name)
		}
		if index > 0 && commands[index-1].Name == command.Name {
			return nil, nil, fmt.Errorf("active script commands must be unique")
		}
		rel := commandRel(command, spec.Platform)
		if !identifiers.PortablePath(rel) {
			return nil, nil, fmt.Errorf("script command %q path is not portable", command.Name)
		}
		if len(roots) > 0 && !insideAnyRoot(rel, roots) {
			return nil, nil, fmt.Errorf("script command %q is outside every runtime root", command.Name)
		}
	}
	return roots, commands, nil
}

func scriptTargets(runtimeDir string, roots []string, commands []skillspec.Command, platform string) map[string]ScriptTarget {
	targets := make(map[string]ScriptTarget, len(commands))
	for _, command := range commands {
		path := filepath.Join(runtimeDir, filepath.FromSlash(commandRel(command, platform)))
		if len(roots) == 0 {
			name := command.Name
			if platform == "windows" {
				name += ".cmd"
			}
			path = filepath.Join(runtimeDir, "bin", name)
		}
		targets[command.Name] = ScriptTarget{runtimeDir: runtimeDir, executable: path}
	}
	return targets
}

func validateRuntimeTree(root string, runtimeRoots []string, commands []skillspec.Command, platform string) error {
	rootInfo, err := os.Lstat(root)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("runtime directory is absent or unsafe")
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !allowedRuntimePath(rel, info.IsDir(), runtimeRoots, commands, platform) {
			return fmt.Errorf("runtime contains unmanaged path: %s", rel)
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return fmt.Errorf("runtime contains unsafe path: %s", rel)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, rel := range runtimeRoots {
		info, err := os.Lstat(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("runtime root is absent or unsafe: %s", rel)
		}
	}
	for name, target := range scriptTargets(root, runtimeRoots, commands, platform) {
		info, err := os.Lstat(target.executable)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("script command is absent or unsafe: %s", name)
		}
		if platform == "unix" && info.Mode().Perm()&0o111 == 0 {
			return fmt.Errorf("script command is not executable: %s", name)
		}
	}
	return nil
}

func allowedRuntimePath(path string, directory bool, roots []string, commands []skillspec.Command, platform string) bool {
	if len(roots) > 0 {
		for _, root := range roots {
			if path == root || strings.HasPrefix(path, root+"/") ||
				(directory && strings.HasPrefix(root, path+"/")) {
				return true
			}
		}
		return false
	}
	if directory && path == "bin" {
		return true
	}
	for _, target := range scriptTargets("", roots, commands, platform) {
		if path == filepath.ToSlash(strings.TrimPrefix(target.executable, string(filepath.Separator))) {
			return true
		}
	}
	return false
}

func stageScriptTree(staged, snapshot string, roots []string, commands []skillspec.Command, platform string) error {
	if err := os.MkdirAll(staged, 0o700); err != nil {
		return err
	}
	if len(roots) > 0 {
		for _, root := range roots {
			source := filepath.Join(snapshot, filepath.FromSlash(root))
			info, err := os.Lstat(source)
			if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("snapshot runtime root is absent or unsafe: %s", root)
			}
			if err := copyTree(source, filepath.Join(staged, filepath.FromSlash(root))); err != nil {
				return fmt.Errorf("stage runtime root %s: %w", root, err)
			}
		}
	} else {
		for _, command := range commands {
			source := filepath.Join(snapshot, filepath.FromSlash(commandRel(command, platform)))
			info, err := os.Lstat(source)
			if err != nil || !info.Mode().IsRegular() {
				return fmt.Errorf("snapshot script command is absent or unsafe: %s", command.Name)
			}
			name := command.Name
			if platform == "windows" {
				name += ".cmd"
			}
			if err := copyFile(source, filepath.Join(staged, "bin", name)); err != nil {
				return fmt.Errorf("stage script command %s: %w", command.Name, err)
			}
		}
	}
	if platform == "unix" {
		for _, target := range scriptTargets(staged, roots, commands, platform) {
			if err := makeExecutable(target.executable); err != nil {
				return err
			}
		}
	}
	return nil
}

func insideAnyRoot(path string, roots []string) bool {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}
