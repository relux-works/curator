package harnessparity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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
	{
		alias := t
		alias.Run("shadowed-static-probe", func(t *testing.T) {})
	}
	{
		alias := 1
		_ = alias
	}
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
	text := string(workflow)
	controlInvocation := strings.Index(text, `"$native_parent_env_binary" --native-parent-env-control`)
	captureInvocation := strings.Index(text, `go test -p 1 -count=1 -v -json ./.temp/original-capture/harnessparity`)
	if controlInvocation < 0 || captureInvocation < 0 || controlInvocation > captureInvocation {
		t.Fatal("native-parent exact-environment control is missing or runs after the capture suite")
	}
	if !strings.Contains(text, `CAPTURE_REASON=capture/native-parent-exact-environment-control-failed`) {
		t.Fatal("native-parent exact-environment control failure does not use its fixed safe reason")
	}
	if !strings.Contains(text, `CAPTURE_NATIVE_PARENT_UNRELATED_CONTROL=native-parent-control`) {
		t.Fatal("native-parent exact-environment control does not plant an unrelated parent variable")
	}
	program, err := staticCaptureWorkflowGoProgram(text, `cat > "$CAPTURE_ROOT/native-parent-env.go" <<'GO'`)
	if err != nil {
		t.Fatalf("extract native-parent environment program: %v", err)
	}
	programText := string(program)
	for _, required := range []string{
		`os.StartProcess(path, arguments, &os.ProcAttr{`,
		`Env:   environment,`,
		`Files: []*os.File{stdin, stdoutWrite, stderrWrite},`,
		`func runNativeParentEnvironmentControl() int`,
		`func nativeParentEnvironmentProbe() int`,
		`environment := []string{"HOME=" + home, "PATH=" + trustedPath}`,
		`nativeParentEnvUnrelatedControl`,
		`nativeParentEnvControlStdoutMarker`,
		`nativeParentEnvControlStderrMarker`,
		`nativeParentTimeoutControlLimit`,
		`missingProgram := executable + ".missing-control"`,
		`nativeParentChildTimeout = 5 * time.Minute`,
		`childStdout.truncated`,
		`childStderr.truncated`,
	} {
		if !strings.Contains(programText, required) {
			t.Errorf("native-parent exact-environment control is missing %q", required)
		}
	}
	if strings.Contains(programText, `command := exec.Command(program, programArgs...)`) {
		t.Fatal("native-parent exact-environment child still uses os/exec")
	}
}

func TestNativeCaptureFailureNamesPassThroughSanitizer(t *testing.T) {
	workflow, err := readBoundNativeCaptureWorkflow()
	if err != nil {
		t.Fatalf("read native capture workflow: %v", err)
	}
	text := string(workflow)
	roots, subtests, err := staticCaptureSuiteIdentifiers()
	if err != nil {
		t.Fatalf("derive static capture test identifiers: %v", err)
	}
	programs := []struct {
		name   string
		marker string
	}{
		{"auditor", `cat > "$CAPTURE_ROOT/audit.go" <<'GO'`},
		{"sanitizer", `cat > "$sanitizer_source_dir/main.go" <<'GO'`},
	}
	wantRoots := staticCaptureNameSet(roots)
	wantSubtests := staticCaptureNameSet(subtests)
	verdictPositiveNames := []string{
		"TestFixtureAncestorProjectionControls/macos-var-private-var-aliases",
		"TestFixtureAncestorProjectionRefusals/expected-entry-duplicated",
		"TestFixtureAncestorFullJSONUsesFlatStringEntries/shadowed-static-probe",
	}
	for _, program := range programs {
		source, err := staticCaptureWorkflowGoProgram(text, program.marker)
		if err != nil {
			t.Fatalf("extract %s failure-name program: %v", program.name, err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), program.name+".go", source, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s failure-name program: %v", program.name, err)
		}
		gotRoots, err := staticCaptureWorkflowAllowlist(parsed, "publicTestRoots")
		if err != nil {
			t.Fatalf("read %s root allowlist: %v", program.name, err)
		}
		gotSubtests, err := staticCaptureWorkflowAllowlist(parsed, "publicFailureSubtests")
		if err != nil {
			t.Fatalf("read %s subtest allowlist: %v", program.name, err)
		}
		assertSet := func(label string, got, want map[string]struct{}) {
			for name := range want {
				if _, ok := got[name]; !ok {
					t.Errorf("static %s identifier %q is missing from the %s allowlist", label, name, program.name)
				}
			}
			for name := range got {
				if _, ok := want[name]; !ok {
					t.Errorf("%s allowlist accepts non-emitted static %s identifier %q", program.name, label, name)
				}
			}
		}
		assertSet("root", gotRoots, wantRoots)
		assertSet("subtest", gotSubtests, wantSubtests)
		for _, name := range verdictPositiveNames {
			if _, ok := gotSubtests[name]; !ok {
				t.Errorf("verdict-positive static name %q is missing from the %s allowlist", name, program.name)
			}
		}
		if program.name == "auditor" {
			for constantName, want := range map[string]string{
				"staticFailureTest":        verdictPositiveNames[0],
				"staticRefusalFailureTest": verdictPositiveNames[1],
			} {
				got, err := staticCaptureWorkflowStringConstant(parsed, constantName)
				if err != nil || got != want {
					t.Errorf("auditor positive control %s = %q (%v), want %q", constantName, got, err, want)
				}
			}
		} else {
			positiveNames, err := staticCaptureWorkflowStringSlice(parsed, "checkFailureDiagnosticControls", "staticSlashNames")
			if err != nil {
				t.Fatalf("read sanitizer verdict-positive control: %v", err)
			}
			positiveSet := staticCaptureNameSet(positiveNames)
			for _, name := range verdictPositiveNames {
				if _, ok := positiveSet[name]; !ok {
					t.Errorf("sanitizer positive control does not exercise verdict name %q", name)
				}
			}
		}
	}
	required := []struct {
		name, fragment string
	}{
		{"failure events only", `if event.Action == "fail" && event.Test == "" {`},
		{"name-only serialization", `fmt.Fprintf(&names, "CAPTURE_FAIL Package=%s Test=%s\n", failure.Package, failure.Test)`},
		{"failing event carries withheld output control", "CAPTURE_PRIVATE_OUTPUT_CONTROL"},
		{"static package allowlist in both programs", `const publicCapturePackage = "github.com/relux-works/curator/.temp/original-capture/harnessparity"`},
		{"static test root allowlist in both programs", `var publicTestRoots = map[string]struct{}{`},
		{"unknown subtests reduce to an allowed root", `if _, ok := publicTestRoots[root]; ok {`},
		{"auditor accepts every mapped root and subtest", `for publicName := range publicFailureSubtests {`},
		{"sanitizer accepts every mapped name", `if allowFailureDiagnosticLine(line) != line {`},
		{"auditor positive control is the verdict alias identifier", `const staticFailureTest = "TestFixtureAncestorProjectionControls/macos-var-private-var-aliases"`},
		{"auditor positive control is the verdict refusal identifier", `const staticRefusalFailureTest = "TestFixtureAncestorProjectionRefusals/expected-entry-duplicated"`},
		{"sanitizer positive controls are static slash tests", `staticSlashNames := []string{`},
		{"failure name controls reject a host-shaped value", `"TestOriginalComparatorControls/example-host.local"`},
		{"failure name controls reject an account-shaped value", `"TestOriginalComparatorControls/example-user"`},
		{"malformed name with output is refused", `TestOriginalComparatorControls/claude_code-mcp Output=UNAPPROVED_OUTPUT_CONTROL", "UNAPPROVED_OUTPUT_CONTROL"`},
		{"sanitizer receives names file", `"$CAPTURE_ROOT/test-failures.names.txt" "$safe_failure_names"`},
		{"only sanitized file is printed", `cat "$safe_failure_names"`},
		{"sanitizer receives private detail file", `"$CAPTURE_ROOT/failure-details.txt" "$safe_failure_details"`},
		{"only sanitized detail file is printed", `cat "$safe_failure_details"`},
		{"native-parent failure has a fixed safe reason", `CAPTURE_REASON=capture/native-parent-row-failed`},
		{"existing reason allowlist remains", `if _, ok := failureDiagnosticReasonCodes[reason]; ok {`},
		{"existing refusal fallback remains", `return "Failure detail withheld by the sanitizer."`},
	}
	for _, check := range required {
		if !strings.Contains(text, check.fragment) {
			t.Errorf("native capture failure-name control %q is missing", check.name)
		}
	}
	if got := strings.Count(text, `"test-output.raw.txt"`); got != 1 ||
		!strings.Contains(text, `filepath.Join(paths.root, "test-output.raw.txt")`) {
		t.Errorf("raw-output sentinel fixture count = %d, want exactly one self-check fixture", got)
	}
	if got := strings.Count(text, `const publicCapturePackage = "github.com/relux-works/curator/.temp/original-capture/harnessparity"`); got != 2 {
		t.Errorf("public failure package allowlist appears %d times, want one auditor and one sanitizer copy", got)
	}
	if got := strings.Count(text, `var publicFailureSubtests = func() map[string]struct{} {`); got != 2 {
		t.Errorf("public failure subtest allowlist appears %d times, want one auditor and one sanitizer copy", got)
	}
	for _, check := range []string{
		`for publicName := range publicTestRoots {`,
		`for publicName := range publicFailureSubtests {`,
	} {
		if got := strings.Count(text, check); got != 2 {
			t.Errorf("public failure positive control %q appears %d times, want one auditor and one sanitizer copy", check, got)
		}
	}
	for _, forbidden := range []string{
		`if [[ "$test_exit" -ne 0 && -f "$failed_names_file" ]]; then`,
		`cat "$CAPTURE_ROOT/test-failures.names.txt"`,
		`cat "$CAPTURE_ROOT/raw-output"`,
		`failureDiagnosticNameLinePattern`,
		`[A-Za-z0-9_./-]+`,
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("native capture workflow bypasses the failure-name sanitizer with %q", forbidden)
		}
	}
}

func staticCaptureNameSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

func staticCaptureWorkflowGoProgram(workflow, marker string) ([]byte, error) {
	if strings.Count(workflow, marker) != 1 {
		return nil, fmt.Errorf("heredoc marker %q is not unique", marker)
	}
	markerStart := strings.Index(workflow, marker)
	lineEnd := strings.IndexByte(workflow[markerStart:], '\n')
	if lineEnd < 0 {
		return nil, fmt.Errorf("heredoc marker %q has no body", marker)
	}
	lineStart := markerStart + lineEnd + 1
	var lines []string
	for lineStart < len(workflow) {
		next := strings.IndexByte(workflow[lineStart:], '\n')
		lineEnd := len(workflow)
		if next >= 0 {
			lineEnd = lineStart + next
		}
		line := workflow[lineStart:lineEnd]
		if line == "          GO" {
			return []byte(strings.Join(lines, "\n")), nil
		}
		if line == "" {
			lines = append(lines, "")
			if next < 0 {
				break
			}
			lineStart = lineEnd + 1
			continue
		}
		if !strings.HasPrefix(line, "          ") {
			return nil, fmt.Errorf("heredoc line lacks workflow indentation")
		}
		lines = append(lines, strings.TrimPrefix(line, "          "))
		if next < 0 {
			break
		}
		lineStart = lineEnd + 1
	}
	return nil, fmt.Errorf("heredoc marker %q has no terminator", marker)
}

func staticCaptureWorkflowAllowlist(file *ast.File, name string) (map[string]struct{}, error) {
	var initializer ast.Expr
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.VAR {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, identifier := range value.Names {
				if identifier.Name != name {
					continue
				}
				if initializer != nil || index >= len(value.Values) {
					return nil, fmt.Errorf("allowlist %s has ambiguous initializer", name)
				}
				initializer = value.Values[index]
			}
		}
	}
	if initializer == nil {
		return nil, fmt.Errorf("allowlist %s is missing", name)
	}
	var literal *ast.CompositeLit
	switch name {
	case "publicTestRoots":
		var ok bool
		literal, ok = initializer.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("allowlist %s is not a direct map literal", name)
		}
	case "publicFailureSubtests":
		call, ok := initializer.(*ast.CallExpr)
		if !ok || len(call.Args) != 0 {
			return nil, fmt.Errorf("allowlist %s is not a zero-argument map factory", name)
		}
		function, ok := call.Fun.(*ast.FuncLit)
		if !ok || function.Body == nil || len(function.Body.List) != 2 {
			return nil, fmt.Errorf("allowlist %s has an unsupported map factory", name)
		}
		assignment, ok := function.Body.List[0].(*ast.AssignStmt)
		if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			return nil, fmt.Errorf("allowlist %s does not initialize its returned map directly", name)
		}
		mapName, ok := assignment.Lhs[0].(*ast.Ident)
		if !ok || mapName.Name != "names" {
			return nil, fmt.Errorf("allowlist %s uses an unexpected map variable", name)
		}
		literal, ok = assignment.Rhs[0].(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("allowlist %s does not initialize a literal map", name)
		}
		returned, ok := function.Body.List[1].(*ast.ReturnStmt)
		if !ok || len(returned.Results) != 1 {
			return nil, fmt.Errorf("allowlist %s does not return its literal map", name)
		}
		returnedName, ok := returned.Results[0].(*ast.Ident)
		if !ok || returnedName.Name != mapName.Name {
			return nil, fmt.Errorf("allowlist %s returns a value other than its literal map", name)
		}
	default:
		return nil, fmt.Errorf("allowlist %s is unsupported", name)
	}
	if !staticCaptureStringStructMap(literal.Type) {
		return nil, fmt.Errorf("allowlist %s is not map[string]struct{}", name)
	}
	entries := make(map[string]struct{}, len(literal.Elts))
	for _, element := range literal.Elts {
		entry, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("allowlist %s has a non-keyed entry", name)
		}
		key, ok := entry.Key.(*ast.BasicLit)
		if !ok || key.Kind != token.STRING {
			return nil, fmt.Errorf("allowlist %s has a nonliteral key", name)
		}
		value, err := strconv.Unquote(key.Value)
		if err != nil {
			return nil, fmt.Errorf("allowlist %s has an invalid string key", name)
		}
		if _, exists := entries[value]; exists {
			return nil, fmt.Errorf("allowlist %s duplicates key %q", name, value)
		}
		emptyValue, ok := entry.Value.(*ast.CompositeLit)
		if !ok || emptyValue.Type != nil || len(emptyValue.Elts) != 0 {
			return nil, fmt.Errorf("allowlist %s key %q has a nonempty value", name, value)
		}
		entries[value] = struct{}{}
	}
	return entries, nil
}

func staticCaptureStringStructMap(expression ast.Expr) bool {
	mapping, ok := expression.(*ast.MapType)
	if !ok {
		return false
	}
	key, ok := mapping.Key.(*ast.Ident)
	if !ok || key.Name != "string" {
		return false
	}
	value, ok := mapping.Value.(*ast.StructType)
	return ok && len(value.Fields.List) == 0
}

func staticCaptureWorkflowStringConstant(file *ast.File, name string) (string, error) {
	var result string
	found := 0
	var lookupErr error
	ast.Inspect(file, func(node ast.Node) bool {
		value, ok := node.(*ast.ValueSpec)
		if !ok || len(value.Names) != len(value.Values) {
			return true
		}
		for index, identifier := range value.Names {
			if identifier.Name != name {
				continue
			}
			found++
			literal, ok := value.Values[index].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				lookupErr = fmt.Errorf("constant %s is not a string literal", name)
				return false
			}
			parsed, err := strconv.Unquote(literal.Value)
			if err != nil {
				lookupErr = fmt.Errorf("constant %s is invalid", name)
				return false
			}
			result = parsed
		}
		return true
	})
	if lookupErr != nil {
		return "", lookupErr
	}
	if found != 1 {
		return "", fmt.Errorf("constant %s is missing", name)
	}
	return result, nil
}

func staticCaptureWorkflowStringSlice(file *ast.File, functionName, variableName string) ([]string, error) {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != functionName || function.Body == nil {
			continue
		}
		var result []string
		found := false
		var lookupErr error
		ast.Inspect(function.Body, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 1 {
				return true
			}
			identifier, ok := assignment.Lhs[0].(*ast.Ident)
			if !ok || identifier.Name != variableName {
				return true
			}
			if found {
				lookupErr = fmt.Errorf("slice %s is assigned more than once", variableName)
				return false
			}
			if assignment.Tok != token.DEFINE || len(assignment.Rhs) != 1 {
				lookupErr = fmt.Errorf("slice %s is not initialized directly", variableName)
				return true
			}
			literal, ok := assignment.Rhs[0].(*ast.CompositeLit)
			if !ok {
				lookupErr = fmt.Errorf("slice %s is not a literal", variableName)
				return true
			}
			slice, ok := literal.Type.(*ast.ArrayType)
			if !ok {
				lookupErr = fmt.Errorf("slice %s has an unexpected type", variableName)
				return true
			}
			element, elementOK := slice.Elt.(*ast.Ident)
			if !elementOK || element.Name != "string" {
				lookupErr = fmt.Errorf("slice %s is not []string", variableName)
				return true
			}
			for _, item := range literal.Elts {
				value, ok := item.(*ast.BasicLit)
				if !ok || value.Kind != token.STRING {
					lookupErr = fmt.Errorf("slice %s has a nonliteral entry", variableName)
					return true
				}
				parsed, err := strconv.Unquote(value.Value)
				if err != nil {
					lookupErr = fmt.Errorf("slice %s has an invalid string entry", variableName)
					return true
				}
				result = append(result, parsed)
			}
			found = true
			return true
		})
		if !found {
			return nil, fmt.Errorf("slice %s is missing or dynamic", variableName)
		}
		if lookupErr != nil {
			return nil, lookupErr
		}
		return result, nil
	}
	return nil, fmt.Errorf("function %s is missing", functionName)
}

type staticCaptureNameRow map[string]string

func staticCaptureSuiteIdentifiers() ([]string, []string, error) {
	workflowPath := normalizeNativeCaptureWorkflowPath(os.Getenv("CAPTURE_WORKFLOW_PATH"))
	if workflowPath == "" {
		return nil, nil, errors.New("CAPTURE_WORKFLOW_PATH is required")
	}
	repositoryRoot := filepath.Dir(filepath.Dir(filepath.Dir(workflowPath)))
	sourceDir := filepath.Join(repositoryRoot, "testdata", "originalcapture")
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read capture test sources: %w", err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		path := filepath.Join(sourceDir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, nil, fmt.Errorf("parse capture test source: %w", err)
		}
		files = append(files, file)
	}
	typeFields := map[string][]string{}
	constants := map[string]string{}
	for _, file := range files {
		for _, declaration := range file.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range group.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if fields, ok := staticCaptureStructFields(spec.Type, typeFields); ok {
						typeFields[spec.Name.Name] = fields
					}
				case *ast.ValueSpec:
					if group.Tok != token.CONST || len(spec.Values) != len(spec.Names) {
						continue
					}
					for index, name := range spec.Names {
						if value, ok := staticCaptureString(spec.Values[index], constants); ok {
							constants[name.Name] = value
						}
					}
				}
			}
		}
	}
	var roots, subtests []string
	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			testReceivers := staticCaptureTestingTNames(function)
			if !staticCaptureTestFunction(function) {
				if staticCaptureHasTestRun(function.Body, testReceivers) {
					return nil, nil, fmt.Errorf("%s: t.Run outside a Test function needs explicit source inventory", fset.Position(function.Pos()))
				}
				continue
			}
			root := function.Name.Name
			roots = append(roots, root)
			locals := staticCaptureLocalExpressions(function.Body)
			var ranges []*ast.RangeStmt
			var runs []*ast.CallExpr
			runSelectors := map[*ast.SelectorExpr]struct{}{}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.RangeStmt:
					ranges = append(ranges, node)
				case *ast.CallExpr:
					selector, ok := node.Fun.(*ast.SelectorExpr)
					if !ok || selector.Sel.Name != "Run" {
						break
					}
					receiver, ok := selector.X.(*ast.Ident)
					if ok && hasStaticCaptureObject(testReceivers, receiver.Obj) {
						runSelectors[selector] = struct{}{}
					}
					if ok && hasStaticCaptureObject(testReceivers, receiver.Obj) && len(node.Args) >= 1 {
						runs = append(runs, node)
					}
				}
				return true
			})
			unhandledRun := false
			ast.Inspect(function.Body, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Run" {
					return true
				}
				receiver, ok := selector.X.(*ast.Ident)
				if !ok || !hasStaticCaptureObject(testReceivers, receiver.Obj) {
					return true
				}
				if _, called := runSelectors[selector]; !called {
					unhandledRun = true
					return false
				}
				return true
			})
			if unhandledRun {
				return nil, nil, fmt.Errorf("%s: t.Run is not a direct call with an inventoried static name", fset.Position(function.Pos()))
			}
			for _, run := range runs {
				environment := map[*ast.Object][]staticCaptureNameRow{}
				for _, loop := range ranges {
					if loop.Body.Pos() > run.Pos() || run.End() > loop.Body.End() {
						continue
					}
					variable, ok := loop.Value.(*ast.Ident)
					if !ok || variable.Obj == nil {
						continue
					}
					rows, ok := staticCaptureRows(loop.X, locals, constants, typeFields, 0)
					if ok {
						environment[variable.Obj] = rows
					}
				}
				names, ok := staticCaptureNameValues(run.Args[0], environment, constants)
				if !ok || len(names) == 0 {
					return nil, nil, fmt.Errorf("%s: cannot derive static t.Run name", fset.Position(run.Pos()))
				}
				for _, name := range names {
					if name == "" || strings.ContainsRune(name, '\x00') {
						return nil, nil, fmt.Errorf("%s: invalid static t.Run name", fset.Position(run.Pos()))
					}
					subtests = append(subtests, root+"/"+name)
				}
			}
		}
	}
	return staticCaptureUnique(roots), staticCaptureUnique(subtests), nil
}

func staticCaptureTestingTNames(function *ast.FuncDecl) map[*ast.Object]struct{} {
	names := map[*ast.Object]struct{}{}
	addParameters := func(fields *ast.FieldList) {
		if fields == nil {
			return
		}
		for _, field := range fields.List {
			star, ok := field.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			selector, ok := star.X.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "T" {
				continue
			}
			packageName, ok := selector.X.(*ast.Ident)
			if !ok || packageName.Name != "testing" {
				continue
			}
			for _, parameter := range field.Names {
				if parameter.Name != "_" && parameter.Obj != nil {
					names[parameter.Obj] = struct{}{}
				}
			}
		}
	}
	addParameters(function.Type.Params)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if literal, ok := node.(*ast.FuncLit); ok {
			addParameters(literal.Type.Params)
		}
		return true
	})
	locals := staticCaptureLocalExpressions(function.Body)
	for changed := true; changed; {
		changed = false
		for object, expression := range locals {
			identifier, ok := expression.(*ast.Ident)
			if !ok || !hasStaticCaptureObject(names, identifier.Obj) || hasStaticCaptureObject(names, object) {
				continue
			}
			names[object] = struct{}{}
			changed = true
		}
	}
	return names
}

func staticCaptureHasTestRun(body ast.Node, receivers map[*ast.Object]struct{}) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Run" {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if ok && hasStaticCaptureObject(receivers, receiver.Obj) {
			found = true
			return false
		}
		return true
	})
	return found
}

func hasStaticCaptureObject(objects map[*ast.Object]struct{}, object *ast.Object) bool {
	if object == nil {
		return false
	}
	_, ok := objects[object]
	return ok
}

func staticCaptureTestFunction(function *ast.FuncDecl) bool {
	if !strings.HasPrefix(function.Name.Name, "Test") || function.Name.Name == "TestMain" || function.Type.Params == nil {
		return false
	}
	for _, field := range function.Type.Params.List {
		pointer, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		selector, ok := pointer.X.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "T" {
			continue
		}
		packageName, ok := selector.X.(*ast.Ident)
		return ok && packageName.Name == "testing"
	}
	return false
}

func staticCaptureLocalExpressions(body ast.Node) map[*ast.Object]ast.Expr {
	locals := map[*ast.Object]ast.Expr{}
	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for index, left := range node.Lhs {
				name, ok := left.(*ast.Ident)
				if ok && name.Obj != nil && index < len(node.Rhs) {
					locals[name.Obj] = node.Rhs[index]
				}
			}
		case *ast.ValueSpec:
			for index, name := range node.Names {
				if name.Obj != nil && index < len(node.Values) {
					locals[name.Obj] = node.Values[index]
				}
			}
		}
		return true
	})
	return locals
}

func staticCaptureRows(expression ast.Expr, locals map[*ast.Object]ast.Expr, constants map[string]string, typeFields map[string][]string, depth int) ([]staticCaptureNameRow, bool) {
	if depth > 8 {
		return nil, false
	}
	switch expression := expression.(type) {
	case *ast.Ident:
		initializer, ok := locals[expression.Obj]
		if !ok || initializer == expression {
			return nil, false
		}
		return staticCaptureRows(initializer, locals, constants, typeFields, depth+1)
	case *ast.CallExpr:
		function, ok := expression.Fun.(*ast.Ident)
		if !ok {
			return nil, false
		}
		var cases []fixtureCase
		switch function.Name {
		case "fixtureCases":
			cases = fixtureCases()
		case "launcherCases":
			cases = launcherCases()
		case "allCases":
			cases = allCases()
		default:
			return nil, false
		}
		rows := make([]staticCaptureNameRow, 0, len(cases))
		for _, testCase := range cases {
			rows = append(rows, staticCaptureNameRow{"Name": testCase.Name})
		}
		return rows, true
	case *ast.CompositeLit:
		slice, ok := expression.Type.(*ast.ArrayType)
		if !ok {
			return nil, false
		}
		if element, ok := slice.Elt.(*ast.Ident); ok && element.Name == "string" {
			rows := make([]staticCaptureNameRow, 0, len(expression.Elts))
			for _, item := range expression.Elts {
				value, ok := staticCaptureString(item, constants)
				if !ok {
					return nil, false
				}
				rows = append(rows, staticCaptureNameRow{"": value})
			}
			return rows, true
		}
		fields, ok := staticCaptureStructFields(slice.Elt, typeFields)
		if !ok {
			return nil, false
		}
		rows := make([]staticCaptureNameRow, 0, len(expression.Elts))
		for _, item := range expression.Elts {
			literal, ok := item.(*ast.CompositeLit)
			if !ok {
				return nil, false
			}
			row := staticCaptureNameRow{}
			for index, item := range literal.Elts {
				fieldName := ""
				value := item
				if keyed, ok := item.(*ast.KeyValueExpr); ok {
					key, ok := keyed.Key.(*ast.Ident)
					if !ok {
						continue
					}
					fieldName, value = key.Name, keyed.Value
				} else if index < len(fields) {
					fieldName = fields[index]
				}
				if fieldName == "" {
					continue
				}
				if text, ok := staticCaptureString(value, constants); ok {
					row[fieldName] = text
				}
			}
			rows = append(rows, row)
		}
		return rows, true
	}
	return nil, false
}

func staticCaptureStructFields(expression ast.Expr, typeFields map[string][]string) ([]string, bool) {
	switch expression := expression.(type) {
	case *ast.Ident:
		fields, ok := typeFields[expression.Name]
		return fields, ok
	case *ast.StructType:
		var fields []string
		for _, field := range expression.Fields.List {
			for _, name := range field.Names {
				fields = append(fields, name.Name)
			}
		}
		return fields, len(fields) > 0
	case *ast.StarExpr:
		return staticCaptureStructFields(expression.X, typeFields)
	}
	return nil, false
}

func staticCaptureString(expression ast.Expr, constants map[string]string) (string, bool) {
	switch expression := expression.(type) {
	case *ast.BasicLit:
		if expression.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(expression.Value)
		return value, err == nil
	case *ast.Ident:
		value, ok := constants[expression.Name]
		return value, ok
	case *ast.ParenExpr:
		return staticCaptureString(expression.X, constants)
	case *ast.BinaryExpr:
		if expression.Op != token.ADD {
			return "", false
		}
		left, leftOK := staticCaptureString(expression.X, constants)
		right, rightOK := staticCaptureString(expression.Y, constants)
		return left + right, leftOK && rightOK
	}
	return "", false
}

func staticCaptureNameValues(expression ast.Expr, environment map[*ast.Object][]staticCaptureNameRow, constants map[string]string) ([]string, bool) {
	switch expression := expression.(type) {
	case *ast.BasicLit, *ast.Ident, *ast.ParenExpr:
		if value, ok := staticCaptureString(expression, constants); ok {
			return []string{value}, true
		}
		if identifier, ok := expression.(*ast.Ident); ok {
			rows, ok := environment[identifier.Obj]
			if !ok {
				return nil, false
			}
			var values []string
			for _, row := range rows {
				if value, ok := row[""]; ok {
					values = append(values, value)
				}
			}
			return values, len(values) > 0
		}
		if parenthesized, ok := expression.(*ast.ParenExpr); ok {
			return staticCaptureNameValues(parenthesized.X, environment, constants)
		}
	case *ast.SelectorExpr:
		variable, ok := expression.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		rows, ok := environment[variable.Obj]
		if !ok {
			return nil, false
		}
		var values []string
		for _, row := range rows {
			if value, ok := row[expression.Sel.Name]; ok {
				values = append(values, value)
			}
		}
		return values, len(values) > 0
	case *ast.BinaryExpr:
		if expression.Op != token.ADD {
			return nil, false
		}
		left, leftOK := staticCaptureNameValues(expression.X, environment, constants)
		right, rightOK := staticCaptureNameValues(expression.Y, environment, constants)
		if !leftOK || !rightOK {
			return nil, false
		}
		var values []string
		for _, prefix := range left {
			for _, suffix := range right {
				values = append(values, prefix+suffix)
			}
		}
		return values, true
	}
	return nil, false
}

func staticCaptureUnique(values []string) []string {
	sort.Strings(values)
	unique := values[:0]
	for _, value := range values {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}

func TestNativeCapturePrivateLogCannotReachOutput(t *testing.T) {
	if fixture := os.Getenv("CAPTURE_PRIVATE_LOG_PRINT_FIXTURE"); fixture != "" {
		data, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal("private-log/output-control-read-failed")
		}
		writer := io.Writer(os.Stdout)
		if os.Getenv("CAPTURE_PRIVATE_LOG_PRINT_STREAM") == "stderr" {
			writer = os.Stderr
		}
		if _, err := writer.Write(data); err != nil {
			t.Fatal("private-log/output-control-write-failed")
		}
		return
	}

	workflow, err := readBoundNativeCaptureWorkflow()
	if err != nil {
		t.Fatalf("read native capture workflow: %v", err)
	}
	script, err := nativeCaptureRunScript(workflow)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkNativeCapturePrivateLogRoutes(script); err != nil {
		t.Fatalf("private test log has an unapproved output route: %v", err)
	}
	sanitizerScript, err := nativeCaptureSanitizeScript(workflow)
	if err != nil {
		t.Fatal(err)
	}
	failureDetailWorkflow := script + "\n" + sanitizerScript
	if err := checkNativeCaptureFailureDetailRoutes(failureDetailWorkflow); err != nil {
		t.Fatalf("failure details have an unapproved output route: %v", err)
	}
	for _, plantedRoute := range []string{
		`cat "$CAPTURE_ROOT/failure-details.txt"`,
		`cat "$CAPTURE_ROOT/failure-details.txt" >&2`,
		`cp "$CAPTURE_ROOT/failure-details.txt" "$GITHUB_OUTPUT"`,
		`raw_failure_details="$CAPTURE_ROOT/failure-details.txt"` + "\n" + `cp "$raw_failure_details" "$GITHUB_OUTPUT"`,
		`tee "$GITHUB_OUTPUT" < "$CAPTURE_ROOT/failure-details.txt"`,
		`raw_failure_details="$CAPTURE_ROOT/failure-details.txt"` + "\n" + `cat "$raw_failure_details"`,
		`cat "$CAPTURE_ROOT"/*`,
	} {
		plant := strings.Replace(sanitizerScript, `cat "$safe_failure_details"`+"\n", `cat "$safe_failure_details"`+"\n"+plantedRoute+"\n", 1)
		if plant == sanitizerScript {
			t.Fatal("failure-detail/output-control-plant-point-missing")
		}
		if err := checkNativeCaptureFailureDetailRoutes(script + "\n" + plant); err == nil {
			t.Fatalf("failure-detail/output-control-plant-accepted route=%q", plantedRoute)
		}
	}

	const assignment = `private_test_log="$CAPTURE_ROOT/test-output.private.raw.txt"`
	for _, plantedRoute := range []string{
		`cat "$private_test_log"`,
		`cat "$CAPTURE_ROOT/test-output.private.raw.txt" >&2`,
		`tee "$CAPTURE_ROOT/output.log" < "$private_test_log"`,
		`cp "$private_test_log" "$CAPTURE_ROOT/test-output.raw.txt"`,
		`cat "$CAPTURE_ROOT"/*.raw.txt`,
		`cat "$CAPTURE_ROOT"/*`,
		`find "$CAPTURE_ROOT" -type f -name '*.raw.txt' -exec cat {} \;`,
	} {
		plant := strings.Replace(script, assignment+"\n", assignment+"\n"+plantedRoute+"\n", 1)
		if plant == script {
			t.Fatal("private-log/output-control-plant-point-missing")
		}
		if err := checkNativeCapturePrivateLogRoutes(plant); err == nil {
			t.Fatalf("private-log/output-control-plant-accepted route=%q", plantedRoute)
		}
	}

	fixturePath := filepath.Join(t.TempDir(), "private-test-log.raw.txt")
	const sentinel = "CAPTURE_PRIVATE_LOG_OUTPUT_CONTROL"
	if err := os.WriteFile(fixturePath, []byte(sentinel+"\n"), 0o600); err != nil {
		t.Fatal("private-log/output-control-fixture-write-failed")
	}
	for _, stream := range []string{"stdout", "stderr"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeCapturePrivateLogCannotReachOutput$")
		cmd.Env = withEnvironment(os.Environ(), map[string]string{
			"CAPTURE_PRIVATE_LOG_PRINT_FIXTURE": fixturePath,
			"CAPTURE_PRIVATE_LOG_PRINT_STREAM":  stream,
		})
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("private-log/output-control-child-failed stream=%s", stream)
		}
		observed := stdout.Bytes()
		if stream == "stderr" {
			observed = stderr.Bytes()
		}
		if !bytes.Contains(observed, []byte(sentinel)) {
			t.Fatalf("private-log/output-control-plant-not-observed stream=%s", stream)
		}
	}
}

func checkNativeCapturePrivateLogRoutes(script string) error {
	const assignment = `private_test_log="$CAPTURE_ROOT/test-output.private.raw.txt"`
	const testLogRedirection = `> "$private_test_log" 2>&1; then`
	const auditLogArgument = `go run "$CAPTURE_ROOT/audit.go" "$private_test_log" `
	seen := map[string]int{"assignment": 0, "test-log-redirection": 0, "audit-log-argument": 0}
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		privateLogMarkers := []string{
			"private_test_log", "test-output.private.raw.txt", "test-output.raw.txt",
			"*.raw.txt", `"$CAPTURE_ROOT"/*`,
		}
		containsPrivateRoute := false
		for _, marker := range privateLogMarkers {
			if strings.Contains(trimmed, marker) {
				containsPrivateRoute = true
				break
			}
		}
		if strings.Contains(trimmed, `find "$CAPTURE_ROOT" -type f`) ||
			strings.Contains(trimmed, `find "$CAPTURE_ROOT" -name`) {
			return fmt.Errorf("unapproved private-log enumeration")
		}
		if !containsPrivateRoute {
			continue
		}
		switch {
		case trimmed == assignment:
			seen["assignment"]++
		case trimmed == testLogRedirection:
			seen["test-log-redirection"]++
		case strings.HasPrefix(trimmed, auditLogArgument) && len(trimmed) == len(auditLogArgument)+1 && trimmed[len(trimmed)-1] == '\\':
			seen["audit-log-argument"]++
		default:
			return fmt.Errorf("unapproved private-log reference")
		}
	}
	for _, reference := range []string{"assignment", "test-log-redirection", "audit-log-argument"} {
		if seen[reference] != 1 {
			return fmt.Errorf("expected one approved private-log reference")
		}
	}
	for _, required := range []string{
		`> "$CAPTURE_ROOT/audit-metrics.raw.txt" 2> "$CAPTURE_ROOT/audit-error.raw.txt"; then`,
		`cat "$CAPTURE_ROOT/audit-metrics.raw.txt" > "$CAPTURE_ROOT/audit-output.raw.txt"`,
	} {
		if !strings.Contains(script, required) {
			return fmt.Errorf("approved sanitized output path missing")
		}
	}
	return nil
}

func checkNativeCaptureFailureDetailRoutes(script string) error {
	allowedContentOutputRoutes := map[string]struct{}{
		`cat "$CAPTURE_ROOT/audit-metrics.raw.txt" > "$CAPTURE_ROOT/audit-output.raw.txt"`: {},
		`cat "$safe_failure_names"`:   {},
		`cat "$safe_failure_details"`: {},
	}
	allowedRawReferenceLines := map[string]struct{}{
		`failure_details_file="$CAPTURE_ROOT/failure-details.txt"`:                                                                        {},
		`"$CAPTURE_ROOT/raw-output" "$test_exit" "$failed_names_file" "$failure_details_file" \`:                                          {},
		`if [[ -s "$CAPTURE_ROOT/failure-details.txt" && -f "$sanitizer_binary" ]] && env -i "${clean_env[@]}" GOPROXY=off GOSUMDB=off \`: {},
		`"$CAPTURE_ROOT/failure-details.txt" "$safe_failure_details" \`:                                                                   {},
	}
	seenRawReferences := map[string]int{}
	inGoHereDoc := false
	for _, line := range strings.Split(script, "\n") {
		trimmed := strings.TrimSpace(line)
		if inGoHereDoc {
			if trimmed == "GO" {
				inGoHereDoc = false
			}
			continue
		}
		if strings.Contains(line, "<<'GO'") {
			inGoHereDoc = true
			continue
		}
		if strings.Contains(trimmed, "cat ") {
			if _, ok := allowedContentOutputRoutes[trimmed]; !ok {
				return fmt.Errorf("content output has an unapproved workflow route")
			}
		}
		for _, command := range []string{"cp ", "tee ", "mv ", "install "} {
			if strings.Contains(trimmed, command) {
				return fmt.Errorf("file copy has an unapproved workflow route")
			}
		}
		if strings.Contains(trimmed, `"$CAPTURE_ROOT"`) && strings.ContainsAny(trimmed, "*?") {
			return fmt.Errorf("capture-root glob has an unapproved failure-detail route")
		}
		if strings.Contains(trimmed, `find "$CAPTURE_ROOT"`) && strings.Contains(trimmed, "-exec") {
			return fmt.Errorf("capture-root enumeration has an unapproved output route")
		}
		if !strings.Contains(line, "failure-details.txt") && !strings.Contains(line, "$failure_details_file") {
			continue
		}
		if _, ok := allowedRawReferenceLines[trimmed]; !ok {
			return fmt.Errorf("raw failure details have an unapproved workflow reference")
		}
		seenRawReferences[trimmed]++
	}
	for line := range allowedRawReferenceLines {
		if seenRawReferences[line] != 1 {
			return fmt.Errorf("expected one approved raw failure-detail workflow reference")
		}
	}

	const startMarker = `safe_failure_details="$CAPTURE_ROOT/failure-details.log-safe.txt"`
	if strings.Count(script, startMarker) != 1 {
		return fmt.Errorf("expected one safe failure-detail output section")
	}
	start := strings.Index(script, startMarker)
	const endMarker = `exit "$sanitize_exit"`
	endRelative := strings.Index(script[start:], endMarker)
	if endRelative < 0 {
		return fmt.Errorf("safe failure-detail output section is unterminated")
	}
	section := script[start : start+endRelative+len(endMarker)]
	var lines []string
	for _, line := range strings.Split(section, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	const approved = `safe_failure_details="$CAPTURE_ROOT/failure-details.log-safe.txt"
if [[ -s "$CAPTURE_ROOT/failure-details.txt" && -f "$sanitizer_binary" ]] && env -i "${clean_env[@]}" GOPROXY=off GOSUMDB=off \
"$sanitizer_binary" --failure-diagnostics "$CAPTURE_ROOT/raw-output" \
"$CAPTURE_ROOT/failure-details.txt" "$safe_failure_details" \
> "$CAPTURE_ROOT/failure-details.stdout" 2> "$CAPTURE_ROOT/failure-details.stderr"; then
if [[ -f "$safe_failure_details" ]]; then
cat "$safe_failure_details"
else
printf '%s\n' 'CAPTURE_DETAIL Failure details withheld; sanitizer produced no safe diagnostic.'
fi
else
printf '%s\n' 'CAPTURE_DETAIL Failure details withheld; sanitizer did not produce a safe diagnostic.'
fi
fi
exit "$sanitize_exit"`
	if strings.Join(lines, "\n") != approved {
		return fmt.Errorf("failure-detail section differs from the sanitizer-only output contract")
	}
	return nil
}

func withEnvironment(environment []string, values map[string]string) []string {
	result := make([]string, 0, len(environment)+len(values))
	for _, entry := range environment {
		key, _, ok := strings.Cut(entry, "=")
		if _, replaced := values[key]; !ok || !replaced {
			result = append(result, entry)
		}
	}
	for key, value := range values {
		result = append(result, key+"="+value)
	}
	return result
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
	return nativeCaptureStepScript(workflow, "Run the complete capture suite")
}

func nativeCaptureSanitizeScript(workflow []byte) (string, error) {
	return nativeCaptureStepScript(workflow, "Sanitize and verify artifact contents")
}

func nativeCaptureStepScript(workflow []byte, stepName string) (string, error) {
	text := strings.ReplaceAll(string(workflow), "\r\n", "\n")
	stepHeader := "      - name: " + stepName + "\n"
	if strings.Count(text, stepHeader) != 1 {
		return "", fmt.Errorf("native capture step %q is missing or duplicated", stepName)
	}
	stepStart := strings.Index(text, stepHeader)
	stepBodyStart := stepStart + len(stepHeader)
	nextStep := strings.Index(text[stepBodyStart:], "\n      - name: ")
	if nextStep < 0 {
		return "", fmt.Errorf("native capture step %q is unterminated", stepName)
	}
	stepEnd := stepBodyStart + nextStep
	step := text[stepStart:stepEnd]
	const runHeader = "        run: |\n"
	runOffset := strings.Index(step, runHeader)
	if runOffset < 0 {
		return "", fmt.Errorf("native capture step %q script block is missing", stepName)
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
			return "", fmt.Errorf("native capture step %q script has unexpected YAML indentation", stepName)
		}
		lines[i] = line[len("          "):]
	}
	return strings.Join(lines, "\n"), nil
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
	lookupError              error
	queryStdout              string
	queryStderr              string
	queryError               error
	queryTimeout             bool
	queryReturnsAfterTimeout bool
	scanFound                bool
	scanError                error
	positiveControlFails     bool
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
