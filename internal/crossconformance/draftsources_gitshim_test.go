package crossconformance

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testcli"
)

// The helper is invoked by the production CLI as its Git executable. Keeping
// its source outside the Go file scan lets the integration guard continue to
// enforce that test code itself launches processes through internal/testcli.
//
//go:embed draftsources_gitshim_main.go.txt
var draftEndpointGitShimTemplate string

// draftEndpointGitShim maps a planned repository endpoint to a local bare
// repository. It records clone/fetch operations, refuses fetches by stored
// remote, and optionally fails endpoint fetches. The helper is compiled via
// testcli.Run and invoked only by the production CLI under test.
func draftEndpointGitShim(t *testing.T, endpoint, bare string, fail bool) ([]string, string) {
	t.Helper()
	realGit := testcli.RequireGit(t)
	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "git-operations.log")
	envLogPath := filepath.Join(shimDir, "git-environment.log")
	sourcePath := filepath.Join(shimDir, "git-shim.go")
	shimName := "git"
	if runtime.GOOS == "windows" {
		shimName += ".exe"
	}
	shimPath := filepath.Join(shimDir, shimName)
	source := strings.NewReplacer(
		"__DRAFT_ENDPOINT__", strconv.Quote(endpoint),
		"__DRAFT_REPOSITORY__", strconv.Quote(draftGitFileURL(bare)),
		"__DRAFT_REAL_GIT__", strconv.Quote(realGit),
		"__DRAFT_LOG_PATH__", strconv.Quote(logPath),
		"__DRAFT_ENV_LOG_PATH__", strconv.Quote(envLogPath),
		"__DRAFT_FAIL_FETCH__", strconv.FormatBool(fail),
	).Replace(draftEndpointGitShimTemplate)
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := testcli.Run(t, "", nil, "", "go", "build", "-o", shimPath, sourcePath); code != 0 {
		t.Fatalf("build Git test shim: exit %d\n%s%s", code, stdout, stderr)
	}
	if err := os.Chmod(shimPath, 0o700); err != nil {
		t.Fatal(err)
	}
	path := shimDir + string(os.PathListSeparator) + os.Getenv("PATH")
	return []string{"PATH=" + path}, logPath
}

func draftGitFileURL(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

func readDraftGitEnvironment(t *testing.T, operationsLog string) string {
	t.Helper()
	return readDraftFetchLog(t, filepath.Join(filepath.Dir(operationsLog), "git-environment.log"))
}

func draftRepoDirForTest(home, projectRoot, alias string) string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	sum := sha256Sum([]byte(abs + "\x00" + alias))
	return filepath.Join(home, "draft-git", sum[:16]+"-"+alias)
}

func sha256Sum(payload []byte) string {
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("%x", sum[:])
}

func readDraftFetchLog(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatal(err)
	}
	return strings.TrimSpace(string(payload))
}
