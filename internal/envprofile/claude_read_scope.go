package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

// Claude read scope (BUG-261029-q1kxk5). In auto mode Claude Code asks once
// per home before the first Read outside the working directories, so an
// unattended member stalls when it reads its own declared skills (managed
// skills/<name> links into the contexts store) or its selected context
// modules. Curator projects a narrow Read allow set into the managed home's
// user settings at every resolve: the lexical skills/<name> path and the
// resolved store root of each declared skill, and the store root of each
// selected context member. Claude matches an allow rule only when both the
// requested and the resolved path match, so a sibling home's lexical path
// into the same immutable entry still prompts. Nothing wider is granted: no
// additionalDirectories, no Edit, no whole home or store, and deny and ask
// rules from any source keep their precedence.
const claudeSettingsName = "settings.json"

// errClaudeSettingsConflict marks operator settings Curator does not
// override: a non-object document, a non-object permissions node, a
// non-array allow list, or a settings path that is not a regular file. The
// projection is skipped and reported as a warning.
var errClaudeSettingsConflict = errors.New("claude settings conflict")

// claudeReadScope is a plan's projected Read rules and the rule prefixes
// Curator owns for retirement.
type claudeReadScope struct {
	rules    []string
	prefixes []string
}

// planClaudeReadScope derives the read scope from the declared skills and the
// lock's context members. Each path is rendered lexically and through the
// canonical location of its deepest existing ancestor, because Claude checks
// the requested path and the path it resolves to.
func planClaudeReadScope(home, homeDir, skillsDir string, skills []skillOf, lock *contextlock.Lock) claudeReadScope {
	manager := newGitManager(home)
	rules := map[string]bool{}
	add := func(path string) {
		for _, variant := range claudePathVariants(path) {
			rules[claudeReadRule(variant)] = true
		}
	}
	for _, skill := range skills {
		for _, base := range claudePathVariants(filepath.Join(homeDir, filepath.FromSlash(skillsDir))) {
			add(filepath.Join(base, skill.name))
		}
		add(skill.root)
	}
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindContext {
			continue
		}
		add(packageRoot(manager.entryPath(home, resolvedOf(member)), member.Directory))
	}
	scope := claudeReadScope{rules: sortedKeysBool(rules)}
	owned := map[string]bool{}
	for _, root := range []string{filepath.Join(homeDir, filepath.FromSlash(skillsDir)), contextstore.Root(home)} {
		for _, variant := range claudePathVariants(root) {
			owned["Read(/"+escapeClaudeRulePath(claudeRulePath(variant))+"/"] = true
		}
	}
	scope.prefixes = sortedKeysBool(owned)
	return scope
}

// claudePathVariants returns the clean path and, when it differs, the path
// below the canonical location of its deepest existing ancestor.
func claudePathVariants(path string) []string {
	clean := filepath.Clean(path)
	variants := []string{clean}
	rest := ""
	for dir := clean; ; dir = filepath.Dir(dir) {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			canonical := filepath.Join(resolved, rest)
			if canonical != clean {
				variants = append(variants, canonical)
			}
			return variants
		}
		if filepath.Dir(dir) == dir {
			return variants
		}
		rest = filepath.Join(filepath.Base(dir), rest)
	}
}

// claudeRulePath renders an absolute path in Claude's POSIX rule form:
// C:\Users\alice becomes /c/Users/alice on Windows.
func claudeRulePath(path string) string {
	slashed := filepath.ToSlash(path)
	if volume := filepath.VolumeName(path); len(volume) == 2 && volume[1] == ':' {
		slashed = "/" + strings.ToLower(volume[:1]) + slashed[2:]
	}
	return slashed
}

// escapeClaudeRulePath escapes gitignore pattern characters so the rule
// matches only the literal path, as Claude does for approved paths.
func escapeClaudeRulePath(path string) string {
	var out strings.Builder
	for _, r := range path {
		if strings.ContainsRune(`\*?[]`, r) {
			out.WriteByte('\\')
		}
		out.WriteRune(r)
	}
	return out.String()
}

// claudeReadRule is the absolute Read rule for everything below one root.
func claudeReadRule(root string) string {
	return "Read(/" + escapeClaudeRulePath(claudeRulePath(root)) + "/**)"
}

// owns reports whether a rule has the projected shape below an owned prefix:
// a literal path and the "/**" suffix. An operator's glob rule such as
// Read(//<store>/**) is not projected shape and stays untouched.
func (s claudeReadScope) owns(rule string) bool {
	if !strings.HasSuffix(rule, "/**)") {
		return false
	}
	for _, prefix := range s.prefixes {
		if !strings.HasPrefix(rule, prefix) || len(rule) <= len(prefix)+len("/**)") {
			continue
		}
		if !hasUnescapedGlob(rule[len(prefix) : len(rule)-len("/**)")]) {
			return true
		}
	}
	return false
}

func hasUnescapedGlob(text string) bool {
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case '*', '?', '[', ']':
			return true
		}
	}
	return false
}

// mergeClaudeReadScope returns the settings document carrying exactly the
// scope's rules among owned rules. Unowned allow entries keep their order,
// retired owned rules are removed, missing rules are appended, and every
// other key and number lexeme survives. changed is false when the input
// already carries the projection; a nil payload then preserves the file.
func mergeClaudeReadScope(input []byte, present bool, scope claudeReadScope) ([]byte, bool, error) {
	object := map[string]any{}
	if present {
		parsed, ok := decodeJSONObject(input)
		if !ok {
			return nil, false, fmt.Errorf("%w: managed %s is not a JSON object", errClaudeSettingsConflict, claudeSettingsName)
		}
		object = parsed
	}
	permissions := map[string]any{}
	if raw, ok := object["permissions"]; ok {
		if permissions, ok = raw.(map[string]any); !ok {
			return nil, false, fmt.Errorf("%w: managed %s permissions is not an object", errClaudeSettingsConflict, claudeSettingsName)
		}
	}
	var allow []any
	if raw, ok := permissions["allow"]; ok {
		if allow, ok = raw.([]any); !ok {
			return nil, false, fmt.Errorf("%w: managed %s permissions.allow is not an array", errClaudeSettingsConflict, claudeSettingsName)
		}
	}
	want := map[string]bool{}
	for _, rule := range scope.rules {
		want[rule] = true
	}
	kept := make([]any, 0, len(allow)+len(scope.rules))
	seen := map[string]bool{}
	for _, entry := range allow {
		if rule, ok := entry.(string); ok && (want[rule] || scope.owns(rule)) {
			if want[rule] && !seen[rule] {
				seen[rule] = true
				kept = append(kept, rule)
			}
			continue
		}
		kept = append(kept, entry)
	}
	for _, rule := range scope.rules {
		if !seen[rule] {
			kept = append(kept, rule)
		}
	}
	if sameAllowList(allow, kept) && (present || len(kept) == 0) {
		return nil, false, nil
	}
	permissions["allow"] = kept
	object["permissions"] = permissions
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(object); err != nil {
		return nil, false, err
	}
	return out.Bytes(), true, nil
}

func sameAllowList(before, after []any) bool {
	if len(before) != len(after) {
		return false
	}
	for i := range before {
		left, leftOK := before[i].(string)
		right, rightOK := after[i].(string)
		if leftOK != rightOK || (leftOK && left != right) {
			return false
		}
	}
	return true
}

// readClaudeSettings reads the managed settings without following links. A
// link or a non-regular file is an operator conflict, not absence.
func readClaudeSettings(root, homeRel string) ([]byte, bool, error) {
	rel := managedJoin(homeRel, claudeSettingsName)
	payload, present, err := readManagedRegular(root, rel)
	if err != nil {
		if strings.Contains(err.Error(), DiagWriteWouldFollowLink) {
			return nil, false, fmt.Errorf("%w: managed %s: %v", errClaudeSettingsConflict, claudeSettingsName, err)
		}
		return nil, false, fmt.Errorf("managed %s: %v", claudeSettingsName, err)
	}
	if !present {
		full, err := managedPath(root, rel, false)
		if err != nil {
			return nil, false, fmt.Errorf("managed %s: %v", claudeSettingsName, err)
		}
		state, err := stateread.Lstat(full)
		if err != nil {
			return nil, false, fmt.Errorf("managed %s: %v", claudeSettingsName, err)
		}
		if state.Kind == stateread.KindPresent && !state.Info.Mode().IsRegular() {
			return nil, false, fmt.Errorf("%w: managed %s is not a regular file", errClaudeSettingsConflict, claudeSettingsName)
		}
	}
	return payload, present, nil
}

// publishClaudeReadScope merges the scope into the managed settings. As for
// .claude.json, the tool does not share our lock: the exact input (including
// absence) is compared again immediately before rename, three times at most.
// A conflict is returned wrapped so the caller can skip it.
func publishClaudeReadScope(root, homeRel string, scope claudeReadScope) error {
	for attempt := 0; attempt < 3; attempt++ {
		input, present, err := readClaudeSettings(root, homeRel)
		if err != nil {
			return err
		}
		payload, changed, err := mergeClaudeReadScope(input, present, scope)
		if err != nil || !changed {
			return err
		}
		full, err := managedPath(root, managedJoin(homeRel, claudeSettingsName), true)
		if err != nil {
			return err
		}
		file, err := privatedir.CreateTemp(filepath.Dir(full), ".curator-write-*.tmp")
		if err != nil {
			return err
		}
		temp := file.Name()
		err = file.Chmod(managedAuthoredFileMode)
		if err == nil {
			_, err = file.Write(payload)
		}
		if err == nil {
			err = file.Sync()
		}
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(temp)
			return err
		}
		rename, release, err := prepareClaudeSeedRename(temp, full)
		if err != nil {
			_ = os.Remove(temp)
			return err
		}
		fresh, freshPresent, err := readClaudeSettings(root, homeRel)
		if err != nil || freshPresent != present || !bytes.Equal(fresh, input) {
			release()
			_ = os.Remove(temp)
			if err != nil {
				return err
			}
			continue
		}
		err = rename()
		release()
		_ = os.Remove(temp)
		return err
	}
	return fmt.Errorf("managed %s changed during all 3 publication checks; read scope not published", claudeSettingsName)
}
