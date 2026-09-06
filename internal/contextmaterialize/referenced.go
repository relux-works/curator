// Package contextmaterialize referenced-form materialization (environments §5.3): the applicable root
// modules materialize as individual files below
// .agent-context/modules/<package-name>/<module-path> carrying their exact
// bytes, and the root file references them through the tool's native
// mechanism. For claude_code the root file carries the generation header,
// chapter parts, and one @-reference part per module. For opencode the root
// file is the generation header part alone and the managed opencode.json —
// fully manager-authored CCJ-1 bytes — carries the ordered instructions
// array. codex_cli and pi support no referenced form (environments §7.2);
// requesting it for them is environment_form_unsupported, never a fallback
// case. Like every materialization here, output is a pure function of
// (lock, precedence policy, environment identifier, form).
package contextmaterialize

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// DiagFormUnsupported reports a configured form the adapter does not
// support (environments §7.2, §7.7).
const DiagFormUnsupported = "environment_form_unsupported"

// DiagFormUnavailable warns that the configured form is unavailable at
// materialization so monolithic was emitted instead (§5.3, §5.7).
const DiagFormUnavailable = "environment_form_unavailable"

// DiagPathCollision reports two protocol paths mapping to one platform path
// (environments §5).
const DiagPathCollision = "environment_path_collision"

// ModulesDir is the managed directory holding referenced module files,
// beside the root-context file.
const ModulesDir = ".agent-context/modules"

// OpenCodeConfigName is the managed opencode.json surface name.
const OpenCodeConfigName = "opencode.json"

// SupportsReferenced reports whether the adapter supports the referenced
// root-context form (environments §7.2).
func SupportsReferenced(environment string) bool {
	return environment == "claude_code" || environment == "opencode"
}

// RootTarget names the adapter's root-context target file (environments
// §7.1): CLAUDE.md for claude_code, AGENTS.md for every other revision-1
// adapter.
func RootTarget(environment string) string {
	if environment == "claude_code" {
		return "CLAUDE.md"
	}
	return "AGENTS.md"
}

// DetectPathCollision fails with environment_path_collision when two
// protocol paths to be written fold to one platform path (environments §5).
// Comparison is case-insensitive over the portable path: every platform
// this revision targets folds ASCII case on write, so a fold here fails
// closed before anything is written rather than letting one write clobber
// the other.
func DetectPathCollision(paths []string) error {
	seen := map[string]string{}
	for _, path := range paths {
		folded := strings.ToLower(path)
		if first, ok := seen[folded]; ok && first != path {
			return fmt.Errorf("%s: protocol paths %q and %q map to one platform path", DiagPathCollision, first, path)
		}
		seen[folded] = path
	}
	return nil
}

// referencePath is the home-relative portable path of one materialized
// module file.
func referencePath(packageName, modulePath string) string {
	return ModulesDir + "/" + packageName + "/" + modulePath
}

// Referenced assembles the referenced root-context file set for an
// environment, keyed by home-relative portable path. It returns
// written=false when the root declares no context (environments §2): no
// surface exists and no file is written. Module files carry the modules'
// exact bytes with no header, chapter, or reference line added.
func Referenced(lock *contextlock.Lock, lockHash string, precedence Precedence, environment string, packages map[string]Package) (files map[string][]byte, written bool, err error) {
	if !SupportsReferenced(environment) {
		return nil, false, fmt.Errorf("%s: the %s adapter supports no referenced form", DiagFormUnsupported, environment)
	}
	rootPkg, ok := packages[lock.Root]
	if !ok {
		return nil, false, fmt.Errorf("no package content for the root %s", lock.Root)
	}
	if !rootPkg.HasContext {
		return nil, false, nil
	}
	order, err := EmittedOrder(lock, precedence)
	if err != nil {
		return nil, false, err
	}
	header, err := Header(lock, lockHash, precedence, order)
	if err != nil {
		return nil, false, err
	}
	files = map[string][]byte{}
	var instructions []string
	instructions = []string{}
	parts := [][]byte{header}
	for _, member := range order {
		pkg, ok := packages[member.Name]
		if !ok {
			return nil, false, fmt.Errorf("no package content for member %s", member.Name)
		}
		modules := Applicable(pkg, "root", environment)
		if len(modules) == 0 {
			continue
		}
		if environment == "claude_code" {
			parts = append(parts, ChapterPart(member))
		}
		for _, module := range modules {
			if err := contextpkg.ValidateModuleBytes(module.Bytes); err != nil {
				return nil, false, fmt.Errorf("module %s of %s: %v", module.Path, member.Name, err)
			}
			path := referencePath(member.Name, module.Path)
			if !identifiers.PortablePath(path) {
				return nil, false, fmt.Errorf("module %s of %s is not a portable path", module.Path, member.Name)
			}
			if _, exists := files[path]; exists {
				return nil, false, fmt.Errorf("%s: duplicate module path %q", DiagPathCollision, path)
			}
			files[path] = module.Bytes
			if environment == "claude_code" {
				parts = append(parts, []byte("@"+path+"\n"))
			} else {
				instructions = append(instructions, path)
			}
		}
	}
	target := RootTarget(environment)
	files[target] = Join(parts)
	if environment == "opencode" {
		document, err := OpenCodeConfig(instructions)
		if err != nil {
			return nil, false, err
		}
		files[OpenCodeConfigName] = document
	}
	names := make([]string, 0, len(files))
	for path := range files {
		names = append(names, path)
	}
	sort.Strings(names)
	if err := DetectPathCollision(names); err != nil {
		return nil, false, err
	}
	return files, true, nil
}

// OpenCodeConfig renders the managed opencode.json bytes (environments
// §5.3): the CCJ-1 bytes of the object whose single member instructions is
// the ordered list of module paths — no other member — followed by exactly
// one trailing LF.
func OpenCodeConfig(instructions []string) ([]byte, error) {
	if instructions == nil {
		instructions = []string{}
	}
	document, err := protocoljson.MarshalCanonical(map[string]any{"instructions": instructions})
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}
