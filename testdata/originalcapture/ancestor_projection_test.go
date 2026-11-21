package harnessparity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	fixtureAncestorProjectionToken     = "${FIXTURE_ROOT_ANCESTORS}"
	fixtureAncestorManagedSettingsPath = "settings.json"
	fixtureAncestorSettingsJSONKey     = "claudeMdExcludes"
	fixtureAncestorCaseClaudeEmpty     = "claude_code-empty"
	fixtureAncestorCaseClaudeMCP       = "claude_code-mcp"
)

type fixtureAncestorProjectionBinding struct {
	table           fixtureAncestorCaseTable
	expectedEntries map[string]struct{}
}

type fixtureAncestorCaseTable struct {
	CaseID               string
	Home                 string
	Profile              string
	EnvID                string
	FixtureRootLexical   string
	FixtureRootCanonical string
	ProjectRoot          string
	CanonicalProjectRoot string
	LaunchDir            string
	ManagedSettingsPath  string
	ManagedRoot          string
}

type fixtureAncestorProjectionRefusal string

func (refusal fixtureAncestorProjectionRefusal) Error() string { return string(refusal) }

type fixtureAncestorJSONEntry struct {
	value string
	start int
	end   int
}

func bindFixtureAncestorProjection(table fixtureAncestorCaseTable) (fixtureAncestorProjectionBinding, error) {
	if table.CaseID != fixtureAncestorCaseClaudeEmpty && table.CaseID != fixtureAncestorCaseClaudeMCP {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/unknown-case")
	}
	paths := []string{table.Home, table.FixtureRootLexical, table.FixtureRootCanonical, table.ProjectRoot, table.CanonicalProjectRoot, table.LaunchDir, table.ManagedSettingsPath, table.ManagedRoot}
	for _, path := range paths {
		if path == "" || !filepath.IsAbs(path) {
			return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/case-table-invalid")
		}
	}
	if table.Profile != "sample" || table.EnvID != "claude_code" {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/case-table-invalid")
	}
	fixtureRoot := filepath.Clean(table.FixtureRootLexical)
	canonicalRoot := filepath.Clean(table.FixtureRootCanonical)
	homeRoot := filepath.Clean(table.Home)
	projectRoot := filepath.Clean(table.ProjectRoot)
	if (homeRoot != filepath.Join(fixtureRoot, "manager") && homeRoot != filepath.Join(canonicalRoot, "manager")) ||
		(projectRoot != filepath.Join(fixtureRoot, "project") && projectRoot != filepath.Join(canonicalRoot, "project")) ||
		filepath.Clean(table.LaunchDir) != filepath.Clean(table.ProjectRoot) ||
		filepath.Clean(table.ManagedSettingsPath) != filepath.Join(table.Home, "environments", table.Profile, table.EnvID, fixtureAncestorManagedSettingsPath) ||
		filepath.Clean(table.ManagedRoot) != filepath.Join(table.Home, "environments", table.Profile, table.EnvID, "CLAUDE.md") {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/case-table-invalid")
	}
	resolvedRoot, err := filepath.EvalSymlinks(fixtureRoot)
	if err != nil {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/root-resolution-failed")
	}
	rootInfo, err := os.Stat(fixtureRoot)
	if err != nil || !rootInfo.IsDir() {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/root-binding-invalid")
	}
	canonicalInfo, err := os.Stat(canonicalRoot)
	if err != nil || !canonicalInfo.IsDir() || !os.SameFile(rootInfo, canonicalInfo) || filepath.Clean(resolvedRoot) != canonicalRoot {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/root-binding-ambiguous")
	}
	resolvedProject, err := filepath.EvalSymlinks(table.ProjectRoot)
	if err != nil || filepath.Clean(resolvedProject) != filepath.Clean(table.CanonicalProjectRoot) {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/project-root-ambiguous")
	}
	projectInfo, err := os.Stat(table.ProjectRoot)
	if err != nil || !projectInfo.IsDir() {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/project-root-invalid")
	}
	canonicalProjectInfo, err := os.Stat(table.CanonicalProjectRoot)
	if err != nil || !canonicalProjectInfo.IsDir() || !os.SameFile(projectInfo, canonicalProjectInfo) {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/project-root-ambiguous")
	}
	expected, err := fixtureAncestorExpectedEntries(table)
	if err != nil {
		return fixtureAncestorProjectionBinding{}, err
	}
	if len(expected) == 0 {
		return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/oracle-failed")
	}
	set := make(map[string]struct{}, len(expected))
	for _, entry := range expected {
		if _, exists := set[entry]; exists {
			return fixtureAncestorProjectionBinding{}, fixtureAncestorProjectionRefusal("fixture-ancestor/oracle-duplicate")
		}
		set[entry] = struct{}{}
	}
	table.FixtureRootLexical = fixtureRoot
	table.FixtureRootCanonical = canonicalRoot
	return fixtureAncestorProjectionBinding{table: table, expectedEntries: set}, nil
}

// fixtureAncestorExpectedEntries is an independent test oracle. It consumes
// only the pre-run case table and filesystem metadata; it does not call
// claudememory, envprofile, the projector, or read product output.
func fixtureAncestorExpectedEntries(table fixtureAncestorCaseTable) ([]string, error) {
	roots := []string{table.ProjectRoot, table.CanonicalProjectRoot, table.LaunchDir}
	for _, root := range roots {
		if root == "" || !filepath.IsAbs(root) {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/root-binding-invalid")
		}
	}
	parents := map[string]struct{}{}
	for _, root := range roots {
		for dir := filepath.Dir(filepath.Clean(root)); ; dir = filepath.Dir(dir) {
			parents[dir] = struct{}{}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	allowedParents := map[string]struct{}{}
	for parent := range parents {
		allowedParents[filepath.Clean(parent)] = struct{}{}
		canonicalParent, err := fixtureAncestorCanonicalCandidate(parent)
		if err != nil {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/oracle-parent-unresolved")
		}
		allowedParents[canonicalParent] = struct{}{}
		allowedParents[filepath.Clean(filepath.Join(canonicalParent, ".claude"))] = struct{}{}
	}
	managedCanonical, err := fixtureAncestorCanonicalCandidate(table.ManagedRoot)
	if err != nil {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/managed-root-unresolved")
	}
	projectRoots := uniqueFixturePaths(table.ProjectRoot, table.CanonicalProjectRoot)
	names := []string{"CLAUDE.md", filepath.Join(".claude", "CLAUDE.md"), "CLAUDE.local.md"}
	entries := map[string]struct{}{}
	for parent := range parents {
		for _, name := range names {
			lexical := filepath.Clean(filepath.Join(parent, name))
			canonical, err := fixtureAncestorCanonicalCandidate(lexical)
			if err != nil {
				return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/canonical-candidate-unresolved")
			}
			protected := lexical == filepath.Clean(table.ManagedRoot) || canonical == managedCanonical
			for _, projectRoot := range projectRoots {
				within, err := fixtureAncestorWithinRoot(projectRoot, lexical)
				if err != nil {
					return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/protected-project-inspection-failed")
				}
				protected = protected || within
				within, err = fixtureAncestorWithinRoot(projectRoot, canonical)
				if err != nil {
					return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/protected-project-inspection-failed")
				}
				protected = protected || within
			}
			if protected {
				continue
			}
			if _, allowed := allowedParents[filepath.Dir(canonical)]; !allowed {
				return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/canonical-target-outside-chain")
			}
			entries[fixtureAncestorEmitPath(lexical)] = struct{}{}
			entries[fixtureAncestorEmitPath(canonical)] = struct{}{}
		}
	}
	result := make([]string, 0, len(entries))
	for entry := range entries {
		result = append(result, entry)
	}
	sort.Strings(result)
	return result, nil
}

func uniqueFixturePaths(paths ...string) []string {
	seen := make(map[string]bool, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if path != "" && !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	return result
}

func fixtureAncestorCanonicalCandidate(path string) (string, error) {
	prefix := filepath.Clean(path)
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(prefix)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if _, statErr := os.Lstat(prefix); statErr == nil {
			return "", fmt.Errorf("existing path cannot be resolved")
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return "", statErr
		}
		parent := filepath.Dir(prefix)
		if parent == prefix {
			return "", fmt.Errorf("path has no existing ancestor")
		}
		missing = append(missing, filepath.Base(prefix))
		prefix = parent
	}
}

func fixtureAncestorWithinRoot(root, candidate string) (bool, error) {
	rel, err := filepath.Rel(root, candidate)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		return true, nil
	}
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return false, err
	}
	for current := filepath.Clean(candidate); ; current = filepath.Dir(current) {
		info, statErr := os.Stat(current)
		if statErr == nil && os.SameFile(rootInfo, info) {
			return true, nil
		}
		if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
			return false, statErr
		}
		if filepath.Dir(current) == current {
			return false, nil
		}
	}
}

func fixtureAncestorEmitPath(path string) string {
	return fixtureAncestorLiteralPattern(filepath.ToSlash(filepath.Clean(path)))
}

// This duplicates the narrow path-literal encoding from the contract's output
// shape so the oracle remains independent of product packages.
func fixtureAncestorLiteralPattern(path string) string {
	var result strings.Builder
	for _, r := range path {
		switch r {
		case '*', '?', '[', ']', '{', '}', '(', ')', '!', '+', '@':
			result.WriteByte('[')
			result.WriteRune(r)
			result.WriteByte(']')
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// projectFixtureAncestorSettings is scoped to the managed Claude
// settings.json of the two named cases. It replaces only the exact contiguous
// matched string-token span; every byte outside that span is retained.
func projectFixtureAncestorSettings(caseID, managedPath string, payload []byte, binding fixtureAncestorProjectionBinding) ([]byte, error) {
	if binding.table.CaseID != caseID || (caseID != fixtureAncestorCaseClaudeEmpty && caseID != fixtureAncestorCaseClaudeMCP) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/case-binding-mismatch")
	}
	if !fixtureAncestorAbsolutePath(binding.table.Home) || !fixtureAncestorAbsolutePath(managedPath) ||
		binding.table.Profile != "sample" || binding.table.EnvID != "claude_code" ||
		filepath.Clean(binding.table.ManagedSettingsPath) != filepath.Join(binding.table.Home, "environments", binding.table.Profile, binding.table.EnvID, fixtureAncestorManagedSettingsPath) ||
		filepath.Clean(managedPath) != filepath.Clean(binding.table.ManagedSettingsPath) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/managed-output-mismatch")
	}
	if len(binding.expectedEntries) == 0 {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/expected-set-empty")
	}
	entries, err := fixtureAncestorSettingsEntries(payload)
	if err != nil {
		return nil, err
	}
	matchedIndexes := make([]int, 0, len(binding.expectedEntries))
	matched := make(map[string]bool, len(binding.expectedEntries))
	for index, entry := range entries {
		if _, expected := binding.expectedEntries[entry.value]; !expected {
			continue
		}
		if matched[entry.value] {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/duplicate-entry")
		}
		matched[entry.value] = true
		matchedIndexes = append(matchedIndexes, index)
	}
	if len(matched) != len(binding.expectedEntries) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/expected-set-mismatch")
	}
	first, last := matchedIndexes[0], matchedIndexes[len(matchedIndexes)-1]
	if last-first+1 != len(matchedIndexes) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/non-contiguous-span")
	}
	start, end := entries[first].start, entries[last].end
	token, err := json.Marshal(fixtureAncestorProjectionToken)
	if err != nil || start < 0 || end < start || end > len(payload) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/span-invalid")
	}
	projected := make([]byte, 0, len(payload)-(end-start)+len(token))
	projected = append(projected, payload[:start]...)
	projected = append(projected, token...)
	projected = append(projected, payload[end:]...)
	if !json.Valid(projected) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/projected-json-invalid")
	}
	return projected, nil
}

func fixtureAncestorAbsolutePath(path string) bool {
	return strings.HasPrefix(path, "/") || absoluteWindowsFixtureRoot(path)
}

func fixtureAncestorSettingsEntries(payload []byte) ([]fixtureAncestorJSONEntry, error) {
	if !json.Valid(payload) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-json-invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	first, err := decoder.Token()
	delim, ok := first.(json.Delim)
	if err != nil || !ok || delim != '{' {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-object-invalid")
	}
	var entries []fixtureAncestorJSONEntry
	found := false
	for decoder.More() {
		keyToken, err := decoder.Token()
		key, ok := keyToken.(string)
		if err != nil || !ok {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-object-invalid")
		}
		valueToken, err := decoder.Token()
		if err != nil {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-object-invalid")
		}
		if key != fixtureAncestorSettingsJSONKey {
			if err := skipFixtureJSONValue(decoder, valueToken); err != nil {
				return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-object-invalid")
			}
			continue
		}
		if found {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-key-ambiguous")
		}
		found = true
		array, ok := valueToken.(json.Delim)
		if !ok || array != '[' {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-array-invalid")
		}
		for decoder.More() {
			before := int(decoder.InputOffset())
			valueToken, err := decoder.Token()
			value, ok := valueToken.(string)
			if err != nil || !ok {
				return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-array-invalid")
			}
			end := int(decoder.InputOffset())
			start := fixtureJSONArrayValueStart(payload, before, end)
			if start < 0 || end < start || end > len(payload) {
				return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-array-invalid")
			}
			entries = append(entries, fixtureAncestorJSONEntry{value: value, start: start, end: end})
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-array-invalid")
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-object-invalid")
	}
	if !found {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-key-missing")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, fixtureAncestorProjectionRefusal("fixture-ancestor/settings-trailing-data")
	}
	return entries, nil
}

func fixtureJSONArrayValueStart(payload []byte, from, end int) int {
	for index := from; index < end; index++ {
		switch payload[index] {
		case ' ', '\t', '\r', '\n', ',':
			continue
		default:
			return index
		}
	}
	return -1
}

func skipFixtureJSONValue(decoder *json.Decoder, token json.Token) error {
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		for decoder.More() {
			key, err := decoder.Token()
			if _, ok := key.(string); err != nil || !ok {
				return fmt.Errorf("invalid object key")
			}
			value, err := decoder.Token()
			if err != nil || skipFixtureJSONValue(decoder, value) != nil {
				return fmt.Errorf("invalid object value")
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return fmt.Errorf("invalid object close")
		}
		return nil
	case '[':
		for decoder.More() {
			value, err := decoder.Token()
			if err != nil || skipFixtureJSONValue(decoder, value) != nil {
				return fmt.Errorf("invalid array value")
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return fmt.Errorf("invalid array close")
		}
		return nil
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
}

type fixtureGitRunner func(context.Context, string, []string, []string) (string, string, error)

type fixtureGitOps struct {
	lookPath func() (string, error)
	run      fixtureGitRunner
	scan     func(...string) (bool, error)
}

func fixtureGitSystemOps() fixtureGitOps {
	return fixtureGitOps{lookPath: func() (string, error) { return exec.LookPath("git") }, run: fixtureGitRun, scan: fixtureGitMetadataOnPathChains}
}

func fixtureGitRun(ctx context.Context, git string, args, environment []string) (string, string, error) {
	cmd := exec.CommandContext(ctx, git, args...)
	cmd.Env = environment
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func requireFixtureProjectOutsideGitWorktree(projectDir string) error {
	return requireFixtureProjectOutsideGitWorktreeWith(projectDir, 5*time.Second, fixtureGitSystemOps())
}

func requireFixtureProjectOutsideGitWorktreeWith(projectDir string, timeout time.Duration, ops fixtureGitOps) error {
	if !filepath.IsAbs(projectDir) {
		return fixtureAncestorProjectionRefusal("fixture-root/project-path-invalid")
	}
	canonicalProject, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		return fixtureAncestorProjectionRefusal("fixture-root/project-path-invalid")
	}
	info, err := os.Stat(canonicalProject)
	if err != nil || !info.IsDir() {
		return fixtureAncestorProjectionRefusal("fixture-root/project-path-invalid")
	}
	git, err := ops.lookPath()
	if err != nil || git == "" {
		return fixtureAncestorProjectionRefusal("fixture-root/git-check-unavailable")
	}
	environment := fixtureGitCheckEnvironment()
	if err := verifyFixtureGitPositiveControl(git, environment, timeout, ops.run); err != nil {
		return fixtureAncestorProjectionRefusal("fixture-root/git-positive-control-failed")
	}
	args := []string{"--no-optional-locks", "--no-replace-objects", "-c", "core.fsmonitor=false", "-c", "core.hooksPath=" + os.DevNull, "-C", canonicalProject, "rev-parse", "--is-inside-work-tree"}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	stdout, stderr, queryErr := ops.run(ctx, git, args, environment)
	queryContextErr := ctx.Err()
	cancel()
	if errors.Is(queryContextErr, context.DeadlineExceeded) {
		return fixtureAncestorProjectionRefusal("fixture-root/git-query-timeout")
	}
	if queryErr == nil {
		switch strings.TrimSpace(stdout) {
		case "true":
			return fixtureAncestorProjectionRefusal("fixture-root/inside-git-worktree")
		case "false":
			if strings.TrimSpace(stderr) != "" {
				return fixtureAncestorProjectionRefusal("fixture-root/git-query-malformed")
			}
		default:
			return fixtureAncestorProjectionRefusal("fixture-root/git-query-malformed")
		}
	} else {
		if ctx.Err() == context.DeadlineExceeded {
			return fixtureAncestorProjectionRefusal("fixture-root/git-query-timeout")
		}
		var exit *exec.ExitError
		if !errors.As(queryErr, &exit) || exit.ExitCode() != 128 || strings.TrimSpace(stdout) != "" || !strings.Contains(strings.ToLower(stderr), "not a git repository") {
			return fixtureAncestorProjectionRefusal("fixture-root/git-query-failed")
		}
		// This expected no-repository error is only a query observation. The
		// independent metadata walk below and the known-worktree positive
		// control establish the outside-worktree result.
	}
	metadataFound, err := ops.scan(projectDir, canonicalProject)
	if err != nil {
		return fixtureAncestorProjectionRefusal("fixture-root/git-metadata-scan-incomplete")
	}
	if metadataFound {
		return fixtureAncestorProjectionRefusal("fixture-root/git-metadata-found")
	}
	return nil
}

func verifyFixtureGitPositiveControl(git string, environment []string, timeout time.Duration, run fixtureGitRunner) error {
	base, err := os.MkdirTemp("", "capture-fixture-git-positive-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(base)
	repo := filepath.Join(base, "repo")
	project := filepath.Join(repo, "project")
	if err := os.MkdirAll(project, 0o700); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if stdout, stderr, err := run(ctx, git, []string{"init", "--quiet", repo}, environment); err != nil || stdout != "" || stderr != "" {
		return fmt.Errorf("known work-tree setup failed")
	}
	args := []string{"--no-optional-locks", "--no-replace-objects", "-c", "core.fsmonitor=false", "-c", "core.hooksPath=" + os.DevNull, "-C", project, "rev-parse", "--is-inside-work-tree"}
	stdout, stderr, err := run(ctx, git, args, environment)
	if err != nil || strings.TrimSpace(stdout) != "true" || strings.TrimSpace(stderr) != "" || ctx.Err() != nil {
		return fmt.Errorf("known work-tree query failed")
	}
	return nil
}

func fixtureGitCheckEnvironment() []string {
	environment := make([]string, 0, len(os.Environ())+6)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			environment = append(environment, entry)
		}
	}
	return append(environment,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_NO_LAZY_FETCH=1",
		"LC_ALL=C",
	)
}

func fixtureGitMetadataOnPathChains(paths ...string) (bool, error) {
	seen := map[string]bool{}
	for _, path := range paths {
		for dir := filepath.Clean(path); ; dir = filepath.Dir(dir) {
			if seen[dir] {
				break
			}
			seen[dir] = true
			gitMetadata, err := os.Lstat(filepath.Join(dir, ".git"))
			if err == nil {
				_ = gitMetadata
				return true, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return false, err
			}
			var markerCount int
			for _, marker := range []string{"HEAD", "objects", "refs", "config"} {
				if _, err := os.Lstat(filepath.Join(dir, marker)); err == nil {
					markerCount++
				} else if !errors.Is(err, fs.ErrNotExist) {
					return false, err
				}
			}
			if markerCount >= 3 {
				return true, nil
			}
			if markerCount != 0 {
				return false, fmt.Errorf("partial bare-repository markers")
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return false, nil
}

func TestFixtureAncestorProjectionControls(t *testing.T) {
	t.Run("macos-var-private-var-aliases", func(t *testing.T) {
		expected := []string{
			"/var/folders/unit/tmp/CLAUDE.md",
			"/var/folders/unit/tmp/.claude/CLAUDE.md",
			"/var/folders/unit/tmp/CLAUDE.local.md",
			"/private/var/folders/unit/tmp/CLAUDE.md",
			"/private/var/folders/unit/tmp/.claude/CLAUDE.md",
			"/private/var/folders/unit/tmp/CLAUDE.local.md",
		}
		input := []byte(`{"keep":{"n":1e2},"claudeMdExcludes":["before","/var/folders/unit/tmp/CLAUDE.md","/private/var/folders/unit/tmp/.claude/CLAUDE.md","/var/folders/unit/tmp/.claude/CLAUDE.md","/private/var/folders/unit/tmp/CLAUDE.local.md","/var/folders/unit/tmp/CLAUDE.local.md","/private/var/folders/unit/tmp/CLAUDE.md","after"],"tail":[2,1]}`)
		want := []byte(`{"keep":{"n":1e2},"claudeMdExcludes":["before","${FIXTURE_ROOT_ANCESTORS}","after"],"tail":[2,1]}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
	t.Run("linux-temp-root", func(t *testing.T) {
		expected := []string{
			"/tmp/cap07-fixture/CLAUDE.md",
			"/tmp/cap07-fixture/.claude/CLAUDE.md",
			"/tmp/cap07-fixture/CLAUDE.local.md",
		}
		input := []byte(`{"keep":"outside","claudeMdExcludes":["before", "/tmp/cap07-fixture/CLAUDE.local.md","/tmp/cap07-fixture/CLAUDE.md","/tmp/cap07-fixture/.claude/CLAUDE.md", "after"],"tail":true}`)
		want := []byte(`{"keep":"outside","claudeMdExcludes":["before", "${FIXTURE_ROOT_ANCESTORS}", "after"],"tail":true}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
	t.Run("windows-drive-slash-normalized-json", func(t *testing.T) {
		expected := []string{
			"C:/Users/runner/AppData/Local/Temp/fixture/CLAUDE.md",
			"C:/Users/runner/AppData/Local/Temp/fixture/.claude/CLAUDE.md",
			"C:/Users/runner/AppData/Local/Temp/fixture/CLAUDE.local.md",
		}
		input := []byte("{\"claudeMdExcludes\":[\"C:/Users/runner/AppData/Local/Temp/fixture/CLAUDE.local.md\",\"C:/Users/runner/AppData/Local/Temp/fixture/CLAUDE.md\",\"C:/Users/runner/AppData/Local/Temp/fixture/.claude/CLAUDE.md\"],\"keep\":17}")
		want := []byte(`{"claudeMdExcludes":["${FIXTURE_ROOT_ANCESTORS}"],"keep":17}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
	t.Run("adjacent-unrelated-private-home-stays-unprojected", func(t *testing.T) {
		expected := []string{`/tmp/cap07/CLAUDE.md`, `/tmp/cap07/.claude/CLAUDE.md`, `/tmp/cap07/CLAUDE.local.md`}
		privatePath := `/Users/runner/.ssh/id_ed25519`
		input := []byte(`{"claudeMdExcludes":["` + privatePath + `","/tmp/cap07/CLAUDE.md","/tmp/cap07/.claude/CLAUDE.md","/tmp/cap07/CLAUDE.local.md"]}`)
		want := []byte(`{"claudeMdExcludes":["` + privatePath + `","${FIXTURE_ROOT_ANCESTORS}"]}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
	t.Run("reordering-is-the-only-normalized-sequence-property", func(t *testing.T) {
		expected := []string{`/tmp/order/CLAUDE.md`, `/tmp/order/.claude/CLAUDE.md`, `/tmp/order/CLAUDE.local.md`}
		input := []byte(`{"claudeMdExcludes":["/tmp/order/CLAUDE.local.md","/tmp/order/CLAUDE.md","/tmp/order/.claude/CLAUDE.md"]}`)
		want := []byte(`{"claudeMdExcludes":["${FIXTURE_ROOT_ANCESTORS}"]}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
}

func TestFixtureAncestorProjectionRefusals(t *testing.T) {
	expected := []string{`/x/abc/CLAUDE.md`, `/x/abc/.claude/CLAUDE.md`, `/x/abc/CLAUDE.local.md`}
	base := append([]string{}, expected...)
	tests := []struct {
		name string
		rows []string
		want string
	}{
		{"one-expected-entry-missing", base[:2], "fixture-ancestor/expected-set-mismatch"},
		{"expected-entry-duplicated", []string{base[0], base[0], base[1], base[2]}, "fixture-ancestor/duplicate-entry"},
		{"unexpected-parent-breaks-span", []string{base[0], `/x/other/CLAUDE.md`, base[1], base[2]}, "fixture-ancestor/non-contiguous-span"},
		{"near-prefix-sibling-breaks-span", []string{base[0], `/x/ab/CLAUDE.md`, base[1], base[2]}, "fixture-ancestor/non-contiguous-span"},
		{"traversal-spelling-breaks-span", []string{base[0], `/x/abc/../elsewhere/CLAUDE.md`, base[1], base[2]}, "fixture-ancestor/non-contiguous-span"},
		{"fourth-filename-breaks-span", []string{base[0], `/x/abc/CLAUDE.override.md`, base[1], base[2]}, "fixture-ancestor/non-contiguous-span"},
		{"matched-entries-under-another-key", nil, "fixture-ancestor/expected-set-mismatch"},
		{"matched-entries-in-another-file", base, "fixture-ancestor/managed-output-mismatch"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := fixtureAncestorControlSettings(test.rows)
			if test.name == "matched-entries-under-another-key" {
				payload = []byte(`{"claudeMdExcludes":["native"],"other":["/x/abc/CLAUDE.md","/x/abc/.claude/CLAUDE.md","/x/abc/CLAUDE.local.md"]}`)
			}
			binding := fixtureAncestorControlBinding(fixtureAncestorCaseClaudeEmpty, expected)
			path := binding.table.ManagedSettingsPath
			if test.name == "matched-entries-in-another-file" {
				path = filepath.Join(filepath.Dir(path), "other.json")
			}
			artifactDir := filepath.Join(t.TempDir(), "artifact")
			err := writeFixtureAncestorControlExport(artifactDir, "case.json", fixtureAncestorCaseClaudeEmpty, path, payload, binding)
			if err == nil || err.Error() != test.want {
				t.Fatalf("wanted refusal %s with no projected output", test.want)
			}
			if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
				t.Fatal("refused control wrote an artifact directory")
			}
		})
	}
	t.Run("permitted-order-string-escaping-and-separator-variants", func(t *testing.T) {
		input := []byte("{\"claudeMdExcludes\":[\"/x\\/abc/CLAUDE.local.md\",  \"/x/abc/.claude/CLAUDE.md\" , \"/x/abc/CLAUDE.md\"],\"tail\":true}")
		want := []byte(`{"claudeMdExcludes":["${FIXTURE_ROOT_ANCESTORS}"],"tail":true}`)
		assertFixtureAncestorProjection(t, input, expected, want)
	})
}

func TestFixturePlacementMustBeOutsideGitWorktree(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal("fixture-root/git-check-failed")
	}
	t.Run("outside-worktree", func(t *testing.T) {
		root := t.TempDir()
		project := filepath.Join(root, "project")
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal("fixture-root/test-setup-failed")
		}
		artifactDir := filepath.Join(t.TempDir(), "artifact")
		if err := writeFixturePlacementControlExport(artifactDir, project, 5*time.Second, fixtureGitSystemOps()); err != nil {
			t.Fatal("plain temporary project was not proven outside a work tree")
		}
		if _, err := os.Stat(filepath.Join(artifactDir, "case.json")); err != nil {
			t.Fatal("accepted outside-worktree control did not write its export")
		}
	})
	t.Run("inside-worktree-refused", func(t *testing.T) {
		root := t.TempDir()
		cmd := exec.Command(git, "init", "--quiet", root)
		cmd.Env = fixtureGitCheckEnvironment()
		if err := cmd.Run(); err != nil {
			t.Fatal("fixture-root/test-repository-setup-failed")
		}
		project := filepath.Join(root, "project")
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal("fixture-root/test-setup-failed")
		}
		artifactDir := filepath.Join(t.TempDir(), "artifact")
		err := writeFixturePlacementControlExport(artifactDir, project, 5*time.Second, fixtureGitSystemOps())
		if err == nil || err.Error() != "fixture-root/inside-git-worktree" {
			t.Fatal("fixture root inside a Git work tree was not refused")
		}
		if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
			t.Fatal("work-tree refusal wrote an artifact directory")
		}
	})
	t.Run("uncertain-git-metadata-refused", func(t *testing.T) {
		root := t.TempDir()
		project := filepath.Join(root, "project")
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal("fixture-root/test-setup-failed")
		}
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("not-a-gitdir\n"), 0o600); err != nil {
			t.Fatal("fixture-root/test-setup-failed")
		}
		artifactDir := filepath.Join(t.TempDir(), "artifact")
		err := writeFixturePlacementControlExport(artifactDir, project, 5*time.Second, fixtureGitSystemOps())
		if err == nil {
			t.Fatal("uncertain Git metadata was treated as proof of outside-worktree")
		}
		switch err.Error() {
		case "fixture-root/git-metadata-found", "fixture-root/git-query-failed", "fixture-root/git-query-malformed":
		default:
			t.Fatalf("uncertain Git metadata did not produce a named fail-closed refusal: %v", err)
		}
		if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
			t.Fatal("uncertain metadata refusal wrote an artifact directory")
		}
	})
	t.Run("bare-repository-markers-refused", func(t *testing.T) {
		root := t.TempDir()
		cmd := exec.Command(git, "init", "--bare", "--quiet", root)
		cmd.Env = fixtureGitCheckEnvironment()
		if err := cmd.Run(); err != nil {
			t.Fatal("fixture-root/test-bare-repository-setup-failed")
		}
		artifactDir := filepath.Join(t.TempDir(), "artifact")
		err := writeFixturePlacementControlExport(artifactDir, root, 5*time.Second, fixtureGitSystemOps())
		if err == nil || err.Error() != "fixture-root/git-metadata-found" {
			t.Fatalf("bare repository markers were not refused: %v", err)
		}
		if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
			t.Fatal("bare-repository refusal wrote an artifact directory")
		}
	})
}

func assertFixtureAncestorProjection(t *testing.T, input []byte, expected []string, want []byte) {
	t.Helper()
	binding := fixtureAncestorControlBinding(fixtureAncestorCaseClaudeEmpty, expected)
	dir := t.TempDir()
	if err := writeFixtureAncestorControlExport(dir, "case.json", fixtureAncestorCaseClaudeEmpty, binding.table.ManagedSettingsPath, input, binding); err != nil {
		t.Fatal("fixture ancestor projection differed from its independent full-JSON expectation")
	}
	got, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("fixture ancestor projection differed from its independent full-JSON expectation")
	}
}

func assertFixtureAncestorRefusal(t *testing.T, input []byte, expected []string, want string) {
	t.Helper()
	binding := fixtureAncestorControlBinding(fixtureAncestorCaseClaudeEmpty, expected)
	dir := filepath.Join(t.TempDir(), "artifact")
	err := writeFixtureAncestorControlExport(dir, "case.json", fixtureAncestorCaseClaudeEmpty, binding.table.ManagedSettingsPath, input, binding)
	if err == nil || err.Error() != want {
		t.Fatalf("wanted refusal %s with no projected output", want)
	}
	if _, statErr := os.Lstat(dir); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatal("refused control wrote an artifact directory")
	}
}

func fixtureAncestorControlBinding(caseID string, expected []string) fixtureAncestorProjectionBinding {
	managedPath := filepath.Join(os.TempDir(), "cap07-control", "manager", "environments", "sample", "claude_code", fixtureAncestorManagedSettingsPath)
	return fixtureAncestorControlBindingAtPath(caseID, managedPath, expected)
}

func fixtureAncestorControlBindingAtPath(caseID, managedPath string, expected []string) fixtureAncestorProjectionBinding {
	set := make(map[string]struct{}, len(expected))
	for _, entry := range expected {
		set[entry] = struct{}{}
	}
	home := managedPath
	for index := 0; index < 4; index++ {
		home = fixtureAncestorPathDir(home)
	}
	managedHome := filepath.Join(home, "environments", "sample", "claude_code")
	return fixtureAncestorProjectionBinding{
		table: fixtureAncestorCaseTable{
			CaseID: caseID, Home: home, Profile: "sample", EnvID: "claude_code",
			ManagedSettingsPath: managedPath, ManagedRoot: filepath.Join(managedHome, "CLAUDE.md"),
		},
		expectedEntries: set,
	}
}

func fixtureAncestorPathDir(path string) string {
	index := strings.LastIndexAny(path, "/\\")
	if index < 0 {
		return "."
	}
	if index == 0 {
		return path[:1]
	}
	if index == 2 && len(path) >= 3 && path[1] == ':' {
		return path[:3]
	}
	return path[:index]
}

func writeFixtureAncestorControlExport(dir, name, caseID, managedPath string, payload []byte, binding fixtureAncestorProjectionBinding) error {
	projected, err := projectFixtureAncestorSettings(caseID, managedPath, payload, binding)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), projected, 0o600)
}

func fixtureAncestorControlSettings(entries []string) []byte {
	array, err := json.Marshal(entries)
	if err != nil {
		panic("control JSON encoding failed")
	}
	return append(append([]byte(`{"claudeMdExcludes":`), array...), []byte(`,"keep":17}`)...)
}
func TestFixtureAncestorContractAuthoredPositiveJSON(t *testing.T) {
	fixtures := []struct {
		name         string
		settingsPath string
		expected     []string
	}{
		{
			name:         "linux-temp-root",
			settingsPath: "/tmp/cap07-root/manager/environments/sample/claude_code/settings.json",
			expected: []string{
				"/tmp/cap07-root/CLAUDE.md",
				"/tmp/cap07-root/.claude/CLAUDE.md",
				"/tmp/cap07-root/CLAUDE.local.md",
				"/tmp/CLAUDE.md",
				"/tmp/.claude/CLAUDE.md",
				"/tmp/CLAUDE.local.md",
				"/CLAUDE.md",
				"/.claude/CLAUDE.md",
				"/CLAUDE.local.md",
			},
		},
		{
			name:         "macos-var-and-private-var",
			settingsPath: "/var/folders/unit/cap07-root/manager/environments/sample/claude_code/settings.json",
			expected: []string{
				"/var/folders/unit/cap07-root/CLAUDE.md",
				"/var/folders/unit/cap07-root/.claude/CLAUDE.md",
				"/var/folders/unit/cap07-root/CLAUDE.local.md",
				"/var/folders/unit/CLAUDE.md",
				"/var/folders/unit/.claude/CLAUDE.md",
				"/var/folders/unit/CLAUDE.local.md",
				"/var/folders/CLAUDE.md",
				"/var/folders/.claude/CLAUDE.md",
				"/var/folders/CLAUDE.local.md",
				"/var/CLAUDE.md",
				"/var/.claude/CLAUDE.md",
				"/var/CLAUDE.local.md",
				"/CLAUDE.md",
				"/.claude/CLAUDE.md",
				"/CLAUDE.local.md",
				"/private/var/folders/unit/cap07-root/CLAUDE.md",
				"/private/var/folders/unit/cap07-root/.claude/CLAUDE.md",
				"/private/var/folders/unit/cap07-root/CLAUDE.local.md",
				"/private/var/folders/unit/CLAUDE.md",
				"/private/var/folders/unit/.claude/CLAUDE.md",
				"/private/var/folders/unit/CLAUDE.local.md",
				"/private/var/folders/CLAUDE.md",
				"/private/var/folders/.claude/CLAUDE.md",
				"/private/var/folders/CLAUDE.local.md",
				"/private/var/CLAUDE.md",
				"/private/var/.claude/CLAUDE.md",
				"/private/var/CLAUDE.local.md",
				"/private/CLAUDE.md",
				"/private/.claude/CLAUDE.md",
				"/private/CLAUDE.local.md",
			},
		},
		{
			name:         "windows-drive-slash-normalized",
			settingsPath: "C:/Users/runner/AppData/Local/Temp/cap07-root/manager/environments/sample/claude_code/settings.json",
			expected: []string{
				"C:/Users/runner/AppData/Local/Temp/cap07-root/CLAUDE.md",
				"C:/Users/runner/AppData/Local/Temp/cap07-root/.claude/CLAUDE.md",
				"C:/Users/runner/AppData/Local/Temp/cap07-root/CLAUDE.local.md",
				"C:/Users/runner/AppData/Local/Temp/CLAUDE.md",
				"C:/Users/runner/AppData/Local/Temp/.claude/CLAUDE.md",
				"C:/Users/runner/AppData/Local/Temp/CLAUDE.local.md",
				"C:/Users/runner/AppData/Local/CLAUDE.md",
				"C:/Users/runner/AppData/Local/.claude/CLAUDE.md",
				"C:/Users/runner/AppData/Local/CLAUDE.local.md",
				"C:/Users/runner/AppData/CLAUDE.md",
				"C:/Users/runner/AppData/.claude/CLAUDE.md",
				"C:/Users/runner/AppData/CLAUDE.local.md",
				"C:/Users/runner/CLAUDE.md",
				"C:/Users/runner/.claude/CLAUDE.md",
				"C:/Users/runner/CLAUDE.local.md",
				"C:/Users/CLAUDE.md",
				"C:/Users/.claude/CLAUDE.md",
				"C:/Users/CLAUDE.local.md",
				"C:/CLAUDE.md",
				"C:/.claude/CLAUDE.md",
				"C:/CLAUDE.local.md",
			},
		},
	}
	for _, fixture := range fixtures {
		for _, caseID := range []string{fixtureAncestorCaseClaudeEmpty, fixtureAncestorCaseClaudeMCP} {
			t.Run(fixture.name+"/"+caseID, func(t *testing.T) {
				input, want := fixtureAncestorFullJSON(caseID, fixture.expected)
				binding := fixtureAncestorControlBindingAtPath(caseID, fixture.settingsPath, fixture.expected)
				dir := t.TempDir()
				if err := writeFixtureAncestorControlExport(dir, "settings.json", caseID, fixture.settingsPath, input, binding); err != nil {
					t.Fatalf("contract-authored positive fixture refused: %v", err)
				}
				got, err := os.ReadFile(filepath.Join(dir, "settings.json"))
				if err != nil || !bytes.Equal(got, want) {
					t.Fatal("contract-authored full-JSON projection changed bytes outside the matched span")
				}
			})
		}
	}
}

func TestFixtureAncestorFullJSONUsesFlatStringEntries(t *testing.T) {
	expected := []string{"/tmp/cap07-root/CLAUDE.md", "/tmp/cap07-root/.claude/CLAUDE.md"}
	input, _ := fixtureAncestorFullJSON(fixtureAncestorCaseClaudeEmpty, expected)
	var decoded struct {
		Excludes []string `json:"claudeMdExcludes"`
	}
	if err := json.Unmarshal(input, &decoded); err != nil {
		t.Fatalf("full-JSON fixture did not encode excludes as a string array: %v", err)
	}
	want := append([]string{"before"}, expected...)
	want = append(want, "after")
	if len(decoded.Excludes) != len(want) {
		t.Fatalf("full-JSON fixture has %d top-level excludes; want %d", len(decoded.Excludes), len(want))
	}
	for index, value := range want {
		if decoded.Excludes[index] != value {
			t.Fatalf("full-JSON fixture exclude %d = %q; want %q", index, decoded.Excludes[index], value)
		}
	}
}

func TestNativeCaptureUploadRequiresSuccessfulCaptureAndSanitizer(t *testing.T) {
	workflow, err := readBoundNativeCaptureWorkflow()
	if err != nil {
		t.Fatalf("read native capture workflow: %v", err)
	}
	condition, err := nativeCaptureUploadCondition(workflow)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"steps.capture.outcome == 'success'",
		"steps.sanitize.outcome == 'success'",
	} {
		if !strings.Contains(condition, required) {
			t.Errorf("artifact upload condition %q does not require %s", condition, required)
		}
	}
}

func TestNativeCaptureFailureNameSummaryWithholdsNames(t *testing.T) {
	workflow, err := readBoundNativeCaptureWorkflow()
	if err != nil {
		t.Fatalf("read native capture workflow: %v", err)
	}
	script, err := nativeCaptureRunScript(workflow)
	if err != nil {
		t.Fatalf("extract native capture run script: %v", err)
	}
	if err := validateNativeCaptureFailureNameSummary(script); err != nil {
		t.Fatalf("native capture failure-name summary: %v", err)
	}
}

func TestNativeCaptureFailureNameSummaryRejectsBoundaryPlants(t *testing.T) {
	workflow, err := readBoundNativeCaptureWorkflow()
	if err != nil {
		t.Fatalf("read native capture workflow: %v", err)
	}
	script, err := nativeCaptureRunScript(workflow)
	if err != nil {
		t.Fatalf("extract native capture run script: %v", err)
	}
	_, realEnd, err := nativeCaptureFailureNameSummaryRange(script)
	if err != nil {
		t.Fatalf("locate native capture failure-name summary: %v", err)
	}

	plants := []struct {
		name  string
		lines []string
	}{
		{name: "duplicate allowed statement", lines: []string{"failure_name_count=0"}},
		{name: "unknown reporter command", lines: []string{"unapproved_reporter"}},
		{name: "quoted here-document marker before real marker", lines: []string{
			"cat <<'CAPTURE_FAKE_END'",
			"fi",
			"exit \"$audit_exit\"",
			"CAPTURE_FAKE_END",
			"unapproved_reporter",
		}},
		{name: "comment marker before real marker", lines: []string{
			"# exit \"$audit_exit\"",
			"unapproved_reporter",
		}},
	}
	for _, plant := range plants {
		t.Run(plant.name, func(t *testing.T) {
			mutated, err := insertNativeCaptureFailureNamePlant(script, plant.lines)
			if err != nil {
				t.Fatal(err)
			}
			_, end, err := nativeCaptureFailureNameSummaryRange(mutated)
			if err != nil {
				t.Fatalf("locate real marker after plant: %v", err)
			}
			if end != realEnd+len(plant.lines) {
				t.Fatalf("selected end marker line %d, want real end marker at %d", end, realEnd+len(plant.lines))
			}
			if err := validateNativeCaptureFailureNameSummary(mutated); err == nil {
				t.Fatal("boundary plant was accepted")
			}
		})
	}

	block, err := nativeCaptureFailureNameSummaryBlock(script)
	if err != nil {
		t.Fatalf("extract native capture failure-name summary: %v", err)
	}
	hereDocOnlyAttack := append([]string{"cat <<'CAPTURE_ATTACK'"}, block...)
	hereDocOnlyAttack = append(hereDocOnlyAttack, "fi", "exit \"$audit_exit\"", "CAPTURE_ATTACK", "unapproved_reporter", "fi")
	if err := validateNativeCaptureFailureNameSummary(strings.Join(hereDocOnlyAttack, "\n")); err == nil {
		t.Fatal("a complete fake block inside a quoted here-document was accepted without executable markers")
	}
}

func TestNativeCaptureRunScriptRequiresRunWithinNamedStep(t *testing.T) {
	const workflow = "      - name: Run the complete capture suite\n" +
		"        shell: bash\n" +
		"      - name: Later step\n" +
		"        run: |\n" +
		"          echo later\n" +
		"      - name: Final step\n" +
		"        run: echo done\n"
	if _, err := nativeCaptureRunScript([]byte(workflow)); err == nil {
		t.Fatal("run script from a later step satisfied the capture-step lookup")
	}
}

func nativeCaptureRunScript(workflow []byte) (string, error) {
	text := strings.ReplaceAll(string(workflow), "\r\n", "\n")
	const stepHeader = "      - name: Run the complete capture suite\n"
	if strings.Count(text, stepHeader) != 1 {
		return "", errors.New("native capture run step is missing or duplicated")
	}
	stepStart := strings.Index(text, stepHeader)
	stepBodyStart := stepStart + len(stepHeader)
	nextStep := strings.Index(text[stepBodyStart:], "\n      - name: ")
	if nextStep < 0 {
		return "", errors.New("native capture run step is unterminated")
	}
	stepEnd := stepBodyStart + nextStep
	step := text[stepStart:stepEnd]
	const runHeader = "        run: |\n"
	runOffset := strings.Index(step, runHeader)
	if runOffset < 0 {
		return "", errors.New("native capture run script block is missing")
	}
	bodyStart := stepStart + runOffset + len(runHeader)
	body := text[bodyStart:stepEnd]
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[i] = ""
			continue
		}
		if !strings.HasPrefix(line, "          ") {
			return "", errors.New("native capture run script has unexpected YAML indentation")
		}
		lines[i] = line[len("          "):]
	}
	return strings.Join(lines, "\n"), nil
}

const (
	nativeCaptureFailureNameBlockStart = "if [[ \"$test_exit\" -ne 0 && -f \"$failed_names_file\" ]]; then"
	nativeCaptureFailureNameBlockEnd   = "exit \"$audit_exit\""
)

type nativeCaptureHereDoc struct {
	delimiter string
	stripTabs bool
}

func nativeCaptureFailureNameSummaryRange(script string) (int, int, error) {
	script = strings.ReplaceAll(script, "\r\n", "\n")
	lines := strings.Split(script, "\n")
	starts, ends, err := nativeCaptureExecutableMarkers(lines)
	if err != nil {
		return 0, 0, err
	}
	if len(starts) != 1 {
		return 0, 0, errors.New("native capture failure-name summary must have exactly one executable start marker")
	}
	if len(ends) != 1 {
		return 0, 0, errors.New("native capture failure-name summary must have exactly one executable end marker")
	}
	if starts[0] >= ends[0] {
		return 0, 0, errors.New("native capture failure-name summary markers are out of order")
	}
	if ends[0] == 0 || strings.TrimSpace(lines[ends[0]-1]) != "fi" {
		return 0, 0, errors.New("native capture failure-name summary enclosing wrapper is missing")
	}
	return starts[0], ends[0], nil
}

func nativeCaptureExecutableMarkers(lines []string) ([]int, []int, error) {
	var starts, ends []int
	var pending, active []nativeCaptureHereDoc
	var quote byte
	continued := false
	for lineNumber, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if len(active) != 0 {
			delimiterLine := line
			if active[0].stripTabs {
				delimiterLine = strings.TrimLeft(delimiterLine, "\t")
			}
			if delimiterLine == active[0].delimiter {
				active = active[1:]
			}
			continue
		}

		if quote == 0 && !continued && len(pending) == 0 {
			switch strings.TrimSpace(line) {
			case nativeCaptureFailureNameBlockStart:
				starts = append(starts, lineNumber)
			case nativeCaptureFailureNameBlockEnd:
				ends = append(ends, lineNumber)
			}
		}

		hereDocs, nextQuote, lineContinued, err := nativeCaptureShellLine(line, quote)
		if err != nil {
			return nil, nil, err
		}
		quote = nextQuote
		pending = append(pending, hereDocs...)
		continued = lineContinued
		if !continued && len(pending) != 0 {
			active = pending
			pending = nil
		}
	}
	if len(active) != 0 || len(pending) != 0 {
		return nil, nil, errors.New("native capture run script has an unterminated here-document")
	}
	if quote != 0 || continued {
		return nil, nil, errors.New("native capture run script has an unterminated shell quote or continuation")
	}
	return starts, ends, nil
}

func nativeCaptureShellLine(line string, quote byte) ([]nativeCaptureHereDoc, byte, bool, error) {
	var hereDocs []nativeCaptureHereDoc
	continued := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			}
			continue
		}
		if quote == '"' {
			if ch == '\\' {
				if i+1 == len(line) {
					continued = true
					break
				}
				i++
				continue
			}
			if ch == '"' {
				quote = 0
			}
			continue
		}
		switch {
		case ch == '\\':
			if i+1 == len(line) {
				continued = true
				break
			}
			i++
		case ch == '\'' || ch == '"':
			quote = ch
		case ch == '#' && (i == 0 || line[i-1] == ' ' || line[i-1] == '\t' || strings.ContainsRune(";&|()", rune(line[i-1]))):
			return hereDocs, quote, continued, nil
		case ch == '<' && i+1 < len(line) && line[i+1] == '<' && (i+2 == len(line) || line[i+2] != '<'):
			hereDoc, next, err := nativeCaptureParseHereDoc(line, i+2)
			if err != nil {
				return nil, quote, false, err
			}
			hereDocs = append(hereDocs, hereDoc)
			i = next - 1
		case ch == '<' && i+2 < len(line) && line[i+1] == '<' && line[i+2] == '<':
			i += 2
		}
	}
	return hereDocs, quote, continued, nil
}

func nativeCaptureParseHereDoc(line string, index int) (nativeCaptureHereDoc, int, error) {
	hereDoc := nativeCaptureHereDoc{}
	if index < len(line) && line[index] == '-' {
		hereDoc.stripTabs = true
		index++
	}
	for index < len(line) && (line[index] == ' ' || line[index] == '\t') {
		index++
	}
	var delimiter strings.Builder
	var quote byte
	for index < len(line) {
		ch := line[index]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			} else {
				delimiter.WriteByte(ch)
			}
			index++
			continue
		}
		if quote == '"' {
			switch ch {
			case '"':
				quote = 0
				index++
			case '\\':
				if index+1 == len(line) {
					return nativeCaptureHereDoc{}, 0, errors.New("native capture run script has an invalid here-document delimiter")
				}
				delimiter.WriteByte(line[index+1])
				index += 2
			default:
				delimiter.WriteByte(ch)
				index++
			}
			continue
		}
		switch {
		case ch == '\'' || ch == '"':
			quote = ch
			index++
		case ch == '\\':
			if index+1 == len(line) {
				return nativeCaptureHereDoc{}, 0, errors.New("native capture run script has an invalid here-document delimiter")
			}
			delimiter.WriteByte(line[index+1])
			index += 2
		case ch == ' ' || ch == '\t' || strings.ContainsRune(";|&<>()", rune(ch)):
			if delimiter.Len() == 0 {
				return nativeCaptureHereDoc{}, 0, errors.New("native capture run script has an empty here-document delimiter")
			}
			hereDoc.delimiter = delimiter.String()
			return hereDoc, index, nil
		default:
			delimiter.WriteByte(ch)
			index++
		}
	}
	if quote != 0 || delimiter.Len() == 0 {
		return nativeCaptureHereDoc{}, 0, errors.New("native capture run script has an invalid here-document delimiter")
	}
	hereDoc.delimiter = delimiter.String()
	return hereDoc, index, nil
}

func nativeCaptureFailureNameSummaryBlock(script string) ([]string, error) {
	lines := strings.Split(strings.ReplaceAll(script, "\r\n", "\n"), "\n")
	start, end, err := nativeCaptureFailureNameSummaryRange(script)
	if err != nil {
		return nil, err
	}
	return lines[start : end-1], nil
}

func validateNativeCaptureFailureNameSummary(script string) error {
	block, err := nativeCaptureFailureNameSummaryBlock(script)
	if err != nil {
		return err
	}
	allowed := map[string]int{
		nativeCaptureFailureNameBlockStart:                                         1,
		"# Names and sanitized details are each capped at 200 lines per platform.": 1,
		"failure_name_count=0":                                                     1,
		"failure_names_truncated=false":                                            1,
		"while IFS= read -r line; do":                                              1,
		"if [[ \"$line\" =~ ^CAPTURE_FAIL\\ Package=([A-Za-z0-9_./-]+)\\ Test=([A-Za-z0-9_./-]*)$ ]]; then": 1,
		"if [[ \"$failure_name_count\" -ge 199 ]]; then":                                                    1,
		"failure_names_truncated=true":                                                                      1,
		"break":                                                                                             1,
		"printf '%s\\n' 'CAPTURE_FAIL Failure detail withheld by the sanitizer.' >&2":                       1,
		"failure_name_count=$((failure_name_count + 1))":                                                    1,
		"fi":                            4,
		"done < \"$failed_names_file\"": 1,
		"if [[ \"$failure_names_truncated\" == \"true\" ]]; then":                                         1,
		"printf '%s\\n' 'CAPTURE_FAIL [remaining failing test names withheld at the 200-line limit]' >&2": 1,
	}
	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := allowed[trimmed]; !ok {
			return errors.New("failure-name summary has an unapproved shell statement")
		}
		allowed[trimmed]--
	}
	for _, remaining := range allowed {
		if remaining != 0 {
			return errors.New("failure-name summary statement is missing or duplicated")
		}
	}
	return nil
}

func insertNativeCaptureFailureNamePlant(script string, plant []string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(script, "\r\n", "\n"), "\n")
	_, end, err := nativeCaptureFailureNameSummaryRange(script)
	if err != nil {
		return "", err
	}
	insertAt := end - 1
	mutated := make([]string, 0, len(lines)+len(plant))
	mutated = append(mutated, lines[:insertAt]...)
	mutated = append(mutated, plant...)
	mutated = append(mutated, lines[insertAt:]...)
	return strings.Join(mutated, "\n"), nil
}

func TestNativeCaptureUploadConditionAcceptsLFAndCRLF(t *testing.T) {
	const workflow = "      - name: Upload platform artifact\n        if: ${{ steps.capture.outcome == 'success' && steps.sanitize.outcome == 'success' }}\n      - name: Next step\n"
	for _, newline := range []string{"\n", "\r\n"} {
		input := []byte(strings.ReplaceAll(workflow, "\n", newline))
		condition, err := nativeCaptureUploadCondition(input)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range []string{
			"steps.capture.outcome == 'success'",
			"steps.sanitize.outcome == 'success'",
		} {
			if !strings.Contains(condition, required) {
				t.Errorf("condition %q does not require %s", condition, required)
			}
		}
	}
}

func TestNativeCaptureWorkflowPathNormalizesMixedSeparators(t *testing.T) {
	const mixed = `D:\a\curator\curator/.github/workflows/original-capture-native.yml`
	want := filepath.Clean(strings.ReplaceAll(mixed, `\`, string(os.PathSeparator)))
	if got := normalizeNativeCaptureWorkflowPath(mixed); got != want {
		t.Fatalf("normalized workflow path = %q, want %q", got, want)
	}
}

func TestNativeCaptureUploadMissingWorkflowBindingFailsClosed(t *testing.T) {
	t.Setenv("CAPTURE_WORKFLOW_PATH", "")
	if _, err := readBoundNativeCaptureWorkflow(); err == nil {
		t.Fatal("missing CAPTURE_WORKFLOW_PATH binding was accepted")
	}
}

func readBoundNativeCaptureWorkflow() ([]byte, error) {
	path := os.Getenv("CAPTURE_WORKFLOW_PATH")
	if path == "" {
		return nil, errors.New("CAPTURE_WORKFLOW_PATH is required")
	}
	return os.ReadFile(normalizeNativeCaptureWorkflowPath(path))
}

func normalizeNativeCaptureWorkflowPath(path string) string {
	path = strings.ReplaceAll(path, `\`, string(os.PathSeparator))
	return filepath.Clean(filepath.FromSlash(path))
}

func nativeCaptureUploadCondition(workflow []byte) (string, error) {
	text := strings.ReplaceAll(string(workflow), "\r\n", "\n")
	const stepStart = "      - name: Upload platform artifact\n"
	start := strings.Index(text, stepStart)
	if start < 0 {
		return "", errors.New("native capture artifact upload step is missing")
	}
	step := text[start+len(stepStart):]
	if next := strings.Index(step, "\n      - name:"); next >= 0 {
		step = step[:next]
	}
	for _, line := range strings.Split(step, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "if:") {
			return trimmed, nil
		}
	}
	return "", errors.New("native capture artifact upload condition is missing")
}

func fixtureAncestorFullJSON(caseID string, entries []string) ([]byte, []byte) {
	encoded, err := json.Marshal(entries)
	if err != nil {
		panic("contract fixture JSON encoding failed")
	}
	input := append([]byte(fmt.Sprintf("{\"case\":%q,\"keep\":{\"number\":1e2},\"claudeMdExcludes\":[\"before\",", caseID)), encoded[1:len(encoded)-1]...)
	input = append(input, []byte(",\"after\"],\"tail\":17}")...)
	want := []byte(fmt.Sprintf("{\"case\":%q,\"keep\":{\"number\":1e2},\"claudeMdExcludes\":[\"before\",\"%s\",\"after\"],\"tail\":17}", caseID, fixtureAncestorProjectionToken))
	return input, want
}

func TestFixtureAncestorPreRunCaseTable(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	home := filepath.Join(root, "manager")
	for _, dir := range []string{project, home} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal("fixture-ancestor/test-table-setup-failed")
		}
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal("fixture-ancestor/test-root-resolution-failed")
	}
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal("fixture-ancestor/test-project-resolution-failed")
	}
	for _, caseID := range []string{fixtureAncestorCaseClaudeEmpty, fixtureAncestorCaseClaudeMCP} {
		managedHome := filepath.Join(home, "environments", "sample", "claude_code")
		table := fixtureAncestorCaseTable{
			CaseID:               caseID,
			Home:                 home,
			Profile:              "sample",
			EnvID:                "claude_code",
			FixtureRootLexical:   root,
			FixtureRootCanonical: canonicalRoot,
			ProjectRoot:          project,
			CanonicalProjectRoot: canonicalProject,
			LaunchDir:            project,
			ManagedSettingsPath:  filepath.Join(managedHome, "settings.json"),
			ManagedRoot:          filepath.Join(managedHome, "CLAUDE.md"),
		}
		binding, err := bindFixtureAncestorProjection(table)
		if err != nil {
			t.Fatalf("pre-run case table refused: %v", err)
		}
		if binding.table.ManagedSettingsPath != filepath.Join(table.Home, "environments", table.Profile, table.EnvID, "settings.json") {
			t.Fatal("full managed settings path was not fixed in the pre-run case table")
		}
		for _, parent := range []string{root, canonicalRoot} {
			for _, name := range []string{"CLAUDE.md", filepath.Join(".claude", "CLAUDE.md"), "CLAUDE.local.md"} {
				expected := filepath.ToSlash(filepath.Join(parent, name))
				if _, ok := binding.expectedEntries[expected]; !ok {
					t.Fatalf("independently bound ancestor set omitted the fixture-root form %q", expected)
				}
			}
		}
	}
}

func TestFixtureAncestorSymlinkCanonicalBindings(t *testing.T) {
	base := t.TempDir()
	canonicalRoot := filepath.Join(base, "canonical-root")
	lexicalRoot := filepath.Join(base, "root-alias")
	if err := os.MkdirAll(canonicalRoot, 0o700); err != nil {
		t.Fatal("fixture-ancestor/symlink-test-setup-failed")
	}
	if err := os.Symlink(canonicalRoot, lexicalRoot); err != nil {
		t.Fatalf("fixture-ancestor/symlink-test-unavailable: %v", err)
	}
	canonicalTarget := filepath.Join(base, "CLAUDE.local.md")
	if err := os.WriteFile(canonicalTarget, []byte("fixture\n"), 0o600); err != nil {
		t.Fatal("fixture-ancestor/symlink-target-setup-failed")
	}
	if err := os.Symlink(canonicalTarget, filepath.Join(canonicalRoot, "CLAUDE.md")); err != nil {
		t.Fatalf("fixture-ancestor/candidate-symlink-test-unavailable: %v", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(canonicalRoot)
	if err != nil {
		t.Fatal("fixture-ancestor/canonical-root-resolution-failed")
	}
	project := filepath.Join(lexicalRoot, "project")
	home := filepath.Join(lexicalRoot, "manager")
	if err := os.MkdirAll(project, 0o700); err != nil {
		t.Fatal("fixture-ancestor/symlink-project-setup-failed")
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal("fixture-ancestor/symlink-home-setup-failed")
	}
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal("fixture-ancestor/symlink-project-resolution-failed")
	}
	managedHome := filepath.Join(home, "environments", "sample", "claude_code")
	table := fixtureAncestorCaseTable{
		CaseID:               fixtureAncestorCaseClaudeEmpty,
		Home:                 home,
		Profile:              "sample",
		EnvID:                "claude_code",
		FixtureRootLexical:   lexicalRoot,
		FixtureRootCanonical: canonicalRoot,
		ProjectRoot:          project,
		CanonicalProjectRoot: canonicalProject,
		LaunchDir:            project,
		ManagedSettingsPath:  filepath.Join(managedHome, "settings.json"),
		ManagedRoot:          filepath.Join(managedHome, "CLAUDE.md"),
	}
	binding, err := bindFixtureAncestorProjection(table)
	if err != nil {
		t.Fatalf("symlink case table refused: %v", err)
	}
	for _, expected := range []string{
		filepath.ToSlash(filepath.Join(lexicalRoot, "CLAUDE.md")),
		filepath.ToSlash(filepath.Join(canonicalRoot, "CLAUDE.md")),
		filepath.ToSlash(canonicalTarget),
	} {
		if _, ok := binding.expectedEntries[expected]; !ok {
			t.Fatalf("canonical or lexical fixture form missing from bound set: %q", expected)
		}
	}
}

func TestFixtureAncestorUnresolvedLinkRefusesWithoutExport(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	home := filepath.Join(root, "manager")
	for _, dir := range []string{project, home} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal("fixture-ancestor/unresolved-link-test-setup-failed")
		}
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "missing", "CLAUDE.md"), filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatalf("fixture-ancestor/unresolved-link-control-unavailable: %v", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal("fixture-ancestor/unresolved-link-root-resolution-failed")
	}
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal("fixture-ancestor/unresolved-link-project-resolution-failed")
	}
	managedHome := filepath.Join(home, "environments", "sample", "claude_code")
	table := fixtureAncestorCaseTable{
		CaseID:               fixtureAncestorCaseClaudeEmpty,
		Home:                 home,
		Profile:              "sample",
		EnvID:                "claude_code",
		FixtureRootLexical:   root,
		FixtureRootCanonical: canonicalRoot,
		ProjectRoot:          project,
		CanonicalProjectRoot: canonicalProject,
		LaunchDir:            project,
		ManagedSettingsPath:  filepath.Join(managedHome, "settings.json"),
		ManagedRoot:          filepath.Join(managedHome, "CLAUDE.md"),
	}
	artifactDir := filepath.Join(t.TempDir(), "artifact")
	_, err = bindFixtureAncestorProjection(table)
	if err == nil || err.Error() != "fixture-ancestor/canonical-candidate-unresolved" {
		t.Fatalf("expected unresolved-link refusal, got %v", err)
	}
	if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatal("unresolved-link refusal wrote an artifact directory")
	}
}

func TestFixtureAncestorGlobLiteralContractFixture(t *testing.T) {
	expected := []string{
		"/tmp/cap07-[[]glob[]][?]/CLAUDE.md",
		"/tmp/cap07-[[]glob[]][?]/.claude/CLAUDE.md",
		"/tmp/cap07-[[]glob[]][?]/CLAUDE.local.md",
		"/tmp/CLAUDE.md",
		"/tmp/.claude/CLAUDE.md",
		"/tmp/CLAUDE.local.md",
		"/CLAUDE.md",
		"/.claude/CLAUDE.md",
		"/CLAUDE.local.md",
	}
	input, want := fixtureAncestorFullJSON(fixtureAncestorCaseClaudeEmpty, expected)
	settingsPath := "/tmp/cap07-[glob]?/manager/environments/sample/claude_code/settings.json"
	binding := fixtureAncestorControlBindingAtPath(fixtureAncestorCaseClaudeEmpty, settingsPath, expected)
	dir := t.TempDir()
	if err := writeFixtureAncestorControlExport(dir, "settings.json", fixtureAncestorCaseClaudeEmpty, settingsPath, input, binding); err != nil {
		t.Fatalf("literal glob contract fixture refused: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("literal glob contract fixture did not preserve its full JSON bytes")
	}
}

func TestFixtureAncestorAdjacentPrivatePathRemainsUnprojected(t *testing.T) {
	expected := []string{
		"/tmp/cap07/CLAUDE.md",
		"/tmp/cap07/.claude/CLAUDE.md",
		"/tmp/cap07/CLAUDE.local.md",
	}
	input := []byte("{\"claudeMdExcludes\":[\"/tmp/cap07/CLAUDE.md\",\"/tmp/cap07/.claude/CLAUDE.md\",\"/tmp/cap07/CLAUDE.local.md\",\"/Users/example/private/state\"]}")
	want := []byte(fmt.Sprintf("{\"claudeMdExcludes\":[\"%s\",\"/Users/example/private/state\"]}", fixtureAncestorProjectionToken))
	assertFixtureAncestorProjection(t, input, expected, want)
}

func TestFixtureAncestorCanonicalTargetOutsideChainRefusesWithoutExport(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	home := filepath.Join(root, "manager")
	outside := filepath.Join(t.TempDir(), "outside")
	for _, dir := range []string{project, home, outside} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal("fixture-ancestor/outside-chain-test-setup-failed")
		}
	}
	target := filepath.Join(outside, "CLAUDE.md")
	if err := os.WriteFile(target, []byte("fixture\n"), 0o600); err != nil {
		t.Fatal("fixture-ancestor/outside-chain-target-setup-failed")
	}
	if err := os.Symlink(target, filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatalf("fixture-ancestor/outside-chain-symlink-unavailable: %v", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal("fixture-ancestor/outside-chain-root-resolution-failed")
	}
	canonicalProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal("fixture-ancestor/outside-chain-project-resolution-failed")
	}
	managedHome := filepath.Join(home, "environments", "sample", "claude_code")
	table := fixtureAncestorCaseTable{
		CaseID:               fixtureAncestorCaseClaudeEmpty,
		Home:                 home,
		Profile:              "sample",
		EnvID:                "claude_code",
		FixtureRootLexical:   root,
		FixtureRootCanonical: canonicalRoot,
		ProjectRoot:          project,
		CanonicalProjectRoot: canonicalProject,
		LaunchDir:            project,
		ManagedSettingsPath:  filepath.Join(managedHome, "settings.json"),
		ManagedRoot:          filepath.Join(managedHome, "CLAUDE.md"),
	}
	artifactDir := filepath.Join(t.TempDir(), "artifact")
	binding, err := bindFixtureAncestorProjection(table)
	if err == nil || err.Error() != "fixture-ancestor/canonical-target-outside-chain" {
		t.Fatalf("expected named outside-chain refusal, got %v", err)
	}
	if binding.expectedEntries != nil {
		t.Fatal("outside-chain target produced a partial expected set")
	}
	if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatal("outside-chain refusal wrote an artifact directory")
	}
}

type fixtureGitTestOutcome struct {
	lookupError          error
	queryStdout          string
	queryStderr          string
	queryError           error
	queryTimeout         bool
	queryReturnsAfterTimeout bool
	scanFound            bool
	scanError            error
	positiveControlFails bool
}

func fixtureGitControlledOps(outcome fixtureGitTestOutcome, observedEnvironments *[]string) fixtureGitOps {
	calls := 0
	return fixtureGitOps{
		lookPath: func() (string, error) {
			if outcome.lookupError != nil {
				return "", outcome.lookupError
			}
			return "/neutral/git", nil
		},
		run: func(ctx context.Context, _ string, _ []string, environment []string) (string, string, error) {
			calls++
			if observedEnvironments != nil {
				*observedEnvironments = append(*observedEnvironments, strings.Join(environment, "\n"))
			}
			if calls == 1 {
				return "", "", nil
			}
			if calls == 2 {
				if outcome.positiveControlFails {
					return "false\n", "", nil
				}
				return "true\n", "", nil
			}
			if outcome.queryTimeout {
				<-ctx.Done()
				if outcome.queryReturnsAfterTimeout {
					return "false\n", "", nil
				}
				return "", "", ctx.Err()
			}
			return outcome.queryStdout, outcome.queryStderr, outcome.queryError
		},
		scan: func(...string) (bool, error) {
			return outcome.scanFound, outcome.scanError
		},
	}
}

func TestFixtureGitQueryUncertaintyRefusesWithoutExport(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o700); err != nil {
		t.Fatal("fixture-root/test-setup-failed")
	}
	t.Setenv("GIT_DIR", filepath.Join(root, "ambient-git-dir"))
	t.Setenv("GIT_WORK_TREE", filepath.Join(root, "ambient-work-tree"))
	t.Setenv("git_index_file", filepath.Join(root, "ambient-index"))
	tests := []struct {
		name    string
		outcome fixtureGitTestOutcome
		timeout time.Duration
		want    string
	}{
		{"work-tree-result", fixtureGitTestOutcome{queryStdout: "true\n"}, time.Second, "fixture-root/inside-git-worktree"},
		{"missing-git", fixtureGitTestOutcome{lookupError: errors.New("missing")}, time.Second, "fixture-root/git-check-unavailable"},
		{"query-timeout", fixtureGitTestOutcome{queryTimeout: true}, 100 * time.Millisecond, "fixture-root/git-query-timeout"},
		{"timed-out-command-with-apparent-result", fixtureGitTestOutcome{queryTimeout: true, queryReturnsAfterTimeout: true}, 100 * time.Millisecond, "fixture-root/git-query-timeout"},
		{"command-error", fixtureGitTestOutcome{queryError: errors.New("command failed")}, time.Second, "fixture-root/git-query-failed"},
		{"ambiguous-result", fixtureGitTestOutcome{queryStdout: "false\nmaybe\n"}, time.Second, "fixture-root/git-query-malformed"},
		{"query-stderr-with-false", fixtureGitTestOutcome{queryStdout: "false\n", queryStderr: "unexpected warning"}, time.Second, "fixture-root/git-query-malformed"},
		{"positive-control-fails", fixtureGitTestOutcome{positiveControlFails: true}, time.Second, "fixture-root/git-positive-control-failed"},
		{"metadata-scan-fails", fixtureGitTestOutcome{queryStdout: "false\n", scanError: errors.New("scan incomplete")}, time.Second, "fixture-root/git-metadata-scan-incomplete"},
		{"metadata-found", fixtureGitTestOutcome{queryStdout: "false\n", scanFound: true}, time.Second, "fixture-root/git-metadata-found"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var environments []string
			ops := fixtureGitControlledOps(test.outcome, &environments)
			artifactDir := filepath.Join(t.TempDir(), "artifact")
			err := writeFixturePlacementControlExport(artifactDir, project, test.timeout, ops)
			if err == nil || err.Error() != test.want {
				t.Fatalf("wanted named placement refusal %s, got %v", test.want, err)
			}
			if _, statErr := os.Lstat(artifactDir); !errors.Is(statErr, fs.ErrNotExist) {
				t.Fatal("uncertain placement control wrote an artifact directory")
			}
			allowedGitEnvironment := map[string]bool{
				"GIT_CONFIG_NOSYSTEM": true,
				"GIT_CONFIG_GLOBAL":   true,
				"GIT_TERMINAL_PROMPT": true,
				"GIT_NO_LAZY_FETCH":   true,
			}
			for _, environment := range environments {
				for _, entry := range strings.Split(environment, "\n") {
					key, _, _ := strings.Cut(entry, "=")
					if strings.HasPrefix(strings.ToUpper(key), "GIT_") && !allowedGitEnvironment[strings.ToUpper(key)] {
						t.Fatalf("ambient Git variable %q reached the query", key)
					}
				}
			}
		})
	}
}

func writeFixturePlacementControlExport(artifactDir, project string, timeout time.Duration, ops fixtureGitOps) error {
	if err := requireFixtureProjectOutsideGitWorktreeWith(project, timeout, ops); err != nil {
		return err
	}
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(artifactDir, "case.json"), []byte("{}\n"), 0o600)
}
