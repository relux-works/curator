package crossconformance

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Transport semantic rows (repository-transport §§2/6): the fallback
// gate runs on every OS as a pure decision table, and the fetch loop
// runs through the compiled CLI with a default-deny fake git on unix
// (POSIX-only wrapper; the declared platform-control reason on
// Windows). Every row asserts the token→class binding, the attempt
// count from the shim log, the exit code, and the lock state.

const (
	draftTransportPrimary   = "https://fixture.test/kit.git"
	draftTransportAlternate = "git@fixture.test:kit.git"
	draftTransportIdentity  = "fixture.test/kit"
)

// draftTransportStderr is the production-shaped fatal line per
// first_failure token: each classifies to its SemanticFailureClass
// after the CLI strips git's clone framing line.
var draftTransportStderr = map[string]string{
	"dns":           "fatal: unable to access 'https://fixture.test/kit.git/': Could not resolve host: fixture.test",
	"auth-rejected": "fatal: Authentication failed for 'https://fixture.test/kit.git/'",
	"tls":           "fatal: unable to access 'https://fixture.test/kit.git/': SSL certificate problem: self signed certificate",
	"host-key":      "Host key verification failed.",
	"integrity":     "fatal: fsck error: object abc is corrupted",
	"identity":      "fatal: unable to access 'https://fixture.test/kit.git/': Redirection from https://fixture.test/kit.git to https://evil.example.org/kit.git is forbidden",
	"ref-moved":     "fatal: couldn't find remote ref refs/heads/no-such-branch",
	"audit":         "remote: key is revoked: access denied",
	"unknown":       "fatal: something entirely new from a future git",
	"http-404":      "fatal: unable to access 'https://fixture.test/kit.git/': The requested URL returned error: 404",
}

func init() {
	for _, id := range []string{"fallback-dns", "fallback-auth-rejected", "fallback-tls", "fallback-host-key", "fallback-integrity", "fallback-identity", "fallback-ref-moved", "fallback-audit", "fallback-unknown", "fallback-http-404", "fallback-policy-unreadable", "pinned-auth", "endpoint-identity-mismatch"} {
		id := id
		registerDraftSemantic(id, func(t *testing.T, c draftSemanticCase) { driveTransportRow(t, c) })
	}
}

// TestDraftTransportFallbackGate pins the §2 gate decision for every
// semantic token on every OS: only dns and auth-rejected advance under
// availability-auth. The fetch-loop rows below prove the executor
// honors the gate.
func TestDraftTransportFallbackGate(t *testing.T) {
	for token, stderr := range draftTransportStderr {
		class, ok := buildrepo.SemanticFailureClass(token)
		if !ok {
			t.Fatalf("token %q has no class", token)
		}
		if got := buildrepo.ClassifyFetchOutput(stderr); got != class {
			t.Fatalf("token %q classifies %s, want %s", token, got, class)
		}
		wantSecond := token == "dns" || token == "auth-rejected"
		if got := buildrepo.AllowSecondAttempt("availability-auth", class); got != wantSecond {
			t.Fatalf("token %q AllowSecondAttempt = %v, want %v", token, got, wantSecond)
		}
		if class, ok := buildrepo.SemanticFailureClass("policy-unreadable"); !ok || class != buildrepo.FailurePolicy {
			t.Fatalf("policy-unreadable binds %v/%s", ok, class)
		}
		if buildrepo.AllowSecondAttempt("availability-auth", buildrepo.FailurePolicy) {
			t.Fatal("policy-unreadable must not advance")
		}
	}
}

func driveTransportRow(t *testing.T, c draftSemanticCase) {
	switch c.ID {
	case "fallback-policy-unreadable":
		driveTransportPolicyUnreadable(t, c)
	case "pinned-auth":
		driveTransportPinnedAuth(t, c)
	case "endpoint-identity-mismatch":
		driveTransportIdentityMismatch(t, c)
	default:
		token := strings.TrimPrefix(c.ID, "fallback-")
		stderr, ok := draftTransportStderr[token]
		if !ok {
			t.Fatalf("no stderr fixture for token %q", token)
		}
		driveTransportFallback(t, c, token, stderr, c.Expected == "attempt-second")
	}
}

// draftCLIKitBare builds a fixture repository with tag v1 over
// skills/review and returns its bare path and locked commit.
func draftCLIKitBare(t *testing.T, root string) (bare, commit string) {
	t.Helper()
	requireGit(t)
	work := filepath.Join(root, "kit-work")
	bare = filepath.Join(root, "kit.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(work, "skills", "review"), "review")
	runDraftGit(t, work, "init", "-q", "-b", "main")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "fixture")
	runDraftGit(t, work, "tag", "v1")
	runDraftGit(t, work, "tag", "v1.0.0")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	return bare, draftGitOutput(t, "", "--git-dir", bare, "rev-parse", "v1^{commit}")
}

// installDraftTransportShim writes a default-deny stand-in git: clones
// of failURL fail with the production-shaped framing plus failStderr
// (exit 128), clones of rewriteURL are served from the bare fixture,
// every clone URL is appended to the log file, and anything else fails
// closed. It returns the PATH directory and the log path.
func installDraftTransportShim(t *testing.T, failURL, failStderr, rewriteURL, bare, realGit string) (fakeDir, logPath string) {
	t.Helper()
	fakeDir = t.TempDir()
	logPath = filepath.Join(fakeDir, "clones.log")
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("log='" + logPath + "'\n")
	script.WriteString("fail=0; rewrite=0; url=''\n")
	script.WriteString("for arg in \"$@\"; do\n")
	script.WriteString("if [ \"$arg\" = '" + failURL + "' ]; then fail=1; url=\"$arg\"; fi\n")
	if rewriteURL != "" {
		script.WriteString("if [ \"$arg\" = '" + rewriteURL + "' ]; then rewrite=1; url=\"$arg\"; fi\n")
	}
	script.WriteString("done\n")
	script.WriteString("is_clone=0; for arg in \"$@\"; do if [ \"$arg\" = \"clone\" ]; then is_clone=1; fi; done\n")
	script.WriteString("if [ \"$is_clone\" = \"1\" ]; then clone_url=''; previous=''; for arg in \"$@\"; do clone_url=\"$previous\"; previous=\"$arg\"; done; printf '%s\\n' \"$clone_url\" >>\"$log\"\n")
	script.WriteString("dest=''; for arg in \"$@\"; do dest=\"$arg\"; done\n")
	script.WriteString("if [ \"$fail\" = \"1\" ]; then printf \"Cloning into '%s'...\\n\" \"$dest\" >&2\n")
	script.WriteString("printf '%s\\n' \"" + failStderr + "\" >&2\n")
	script.WriteString("exit 128; fi\n")
	script.WriteString("if [ \"$rewrite\" = \"1\" ]; then\n")
	script.WriteString("args=''; for arg in \"$@\"; do if [ \"$arg\" = '" + rewriteURL + "' ]; then arg='file://" + bare + "'; fi\n")
	script.WriteString("args=\"$args\n$arg\"; done\n")
	script.WriteString("oldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\n")
	script.WriteString("exec '" + realGit + "' \"$@\"; fi\n")
	script.WriteString("printf \"Cloning into '%s'...\\n\" \"$dest\" >&2\n")
	script.WriteString("echo 'fatal: unexpected clone in transport fixture' >&2\n")
	script.WriteString("exit 128; fi\n")
	script.WriteString("exec '" + realGit + "' \"$@\"\n")
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	return fakeDir, logPath
}

func draftTransportPATH(t *testing.T, fakeDir string) []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	return []string{"PATH=" + fakeDir + string(os.PathListSeparator) + os.Getenv("PATH")}
}

func draftCloneLog(t *testing.T, logPath string) []string {
	t.Helper()
	payload, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(payload), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

func draftTransportPolicy(endpoints, fallback, pin string) string {
	policy := `{"schema_version":1,"repositories":{"` + draftTransportIdentity + `":{"endpoints":[` + endpoints + `],"fallback":"` + fallback + `"`
	if pin != "" {
		policy += `,"pin":"` + pin + `"`
	}
	return policy + `}}}`
}

func draftTransportEndpoint(url, auth string) string {
	return `{"url":"` + url + `","authentication":"` + auth + `"}`
}

// driveTransportFallback runs one fallback matrix row through
// `project resolve`: the first endpoint fails with the token's class,
// the second is served from the bare fixture, and wantSecond decides
// whether the loop advances.
func driveTransportFallback(t *testing.T, _ draftSemanticCase, token, stderr string, wantSecond bool) {
	class, ok := buildrepo.SemanticFailureClass(token)
	if !ok {
		t.Fatalf("token %q has no class", token)
	}
	if got := buildrepo.ClassifyFetchOutput(stderr); got != class {
		t.Fatalf("stderr classifies %s, want %s", got, class)
	}
	realGit := requireGit(t)
	root := t.TempDir()
	bare, commit := draftCLIKitBare(t, root)
	fakeDir, logPath := installDraftTransportShim(t, draftTransportPrimary, stderr, draftTransportAlternate, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	policy := draftTransportPolicy(
		draftTransportEndpoint(draftTransportPrimary, "team-https")+","+draftTransportEndpoint(draftTransportAlternate, "team-alt"),
		"availability-auth", "")
	if err := os.WriteFile(filepath.Join(filepath.Dir(configPath), "source-policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"` + draftTransportIdentity + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, resolveStderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	clones := draftCloneLog(t, logPath)
	if wantSecond {
		if code != 0 {
			t.Fatalf("resolve = %d, want success:\n%s", code, resolveStderr)
		}
		if len(clones) != 2 || clones[0] != draftTransportPrimary || clones[1] != draftTransportAlternate {
			t.Fatalf("clones = %v, want both endpoints in order", clones)
		}
		if want := "endpoint 1 (https"; !strings.Contains(resolveStderr, want) || !strings.Contains(resolveStderr, string(class)) {
			t.Fatalf("stderr misses the endpoint-1 %s clause:\n%s", class, resolveStderr)
		}
		for _, leak := range []string{draftTransportPrimary, draftTransportAlternate, "endpoint 2: not attempted"} {
			if strings.Contains(resolveStderr, leak) {
				t.Fatalf("stderr leaks or misreports %q:\n%s", leak, resolveStderr)
			}
		}
		lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
		if err != nil {
			t.Fatal(err)
		}
		member, ok := lock.Find("review")
		if !ok || member.Package.Commit.Hex != commit {
			t.Fatalf("lock binds %+v, want alternate commit %s", member.Package, commit)
		}
		if member.Package.Repository != draftTransportIdentity {
			t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
		}
		return
	}
	if code == 0 {
		t.Fatalf("resolve succeeded, want the %s stop", class)
	}
	if len(clones) != 1 || clones[0] != draftTransportPrimary {
		t.Fatalf("clones = %v, want exactly the first endpoint", clones)
	}
	if !strings.Contains(resolveStderr, string(class)) || !strings.Contains(resolveStderr, class.Reason()) {
		t.Fatalf("stderr misses the %s clause:\n%s", class, resolveStderr)
	}
	if !strings.Contains(resolveStderr, "endpoint 2: not attempted") {
		t.Fatalf("stderr misses the not-attempted record:\n%s", resolveStderr)
	}
	for _, leak := range []string{draftTransportPrimary, draftTransportAlternate} {
		if strings.Contains(resolveStderr, leak) {
			t.Fatalf("stderr leaks %q:\n%s", leak, resolveStderr)
		}
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("stopped resolve published a lock")
	}
}

func driveTransportPinnedAuth(t *testing.T, _ draftSemanticCase) {
	stderr := draftTransportStderr["auth-rejected"]
	realGit := requireGit(t)
	root := t.TempDir()
	bare, _ := draftCLIKitBare(t, root)
	fakeDir, logPath := installDraftTransportShim(t, draftTransportPrimary, stderr, draftTransportAlternate, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	policy := draftTransportPolicy(
		draftTransportEndpoint(draftTransportPrimary, "team-https")+","+draftTransportEndpoint(draftTransportAlternate, "team-alt"),
		"availability-auth", draftTransportPrimary)
	if err := os.WriteFile(filepath.Join(filepath.Dir(configPath), "source-policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"` + draftTransportIdentity + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, resolveStderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("pinned resolve succeeded, want the stop:\n%s", resolveStderr)
	}
	if clones := draftCloneLog(t, logPath); len(clones) != 1 || clones[0] != draftTransportPrimary {
		t.Fatalf("clones = %v, want exactly the pinned endpoint", clones)
	}
	if !strings.Contains(resolveStderr, "auth") {
		t.Fatalf("stderr misses the auth clause:\n%s", resolveStderr)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("pinned stop published a lock")
	}
}

func driveTransportIdentityMismatch(t *testing.T, _ draftSemanticCase) {
	realGit := requireGit(t)
	root := t.TempDir()
	bare, _ := draftCLIKitBare(t, root)
	// Default-deny shim: any clone attempt fails the row through the log.
	fakeDir, logPath := installDraftTransportShim(t, "https://never.invalid/x.git", "fatal: unexpected", "", bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	policy := draftTransportPolicy(draftTransportEndpoint("https://evil.org/kit", "team-https"), "availability-auth", "")
	if err := os.WriteFile(filepath.Join(filepath.Dir(configPath), "source-policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"example.org/kit","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, resolveStderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded, want repository_policy_invalid:\n%s", resolveStderr)
	}
	if !strings.Contains(resolveStderr, "repository_policy_invalid") {
		t.Fatalf("stderr misses repository_policy_invalid:\n%s", resolveStderr)
	}
	if clones := draftCloneLog(t, logPath); len(clones) != 0 {
		t.Fatalf("clones = %v, want zero attempts", clones)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("invalid policy published a lock")
	}
}

func driveTransportPolicyUnreadable(t *testing.T, _ draftSemanticCase) {
	realGit := requireGit(t)
	root := t.TempDir()
	bare, _ := draftCLIKitBare(t, root)
	fakeDir, logPath := installDraftTransportShim(t, "https://never.invalid/x.git", "fatal: unexpected", "", bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	policyPath := filepath.Join(filepath.Dir(configPath), "source-policy.json")
	policy := draftTransportPolicy(draftTransportEndpoint(draftTransportPrimary, "team-https"), "availability-auth", "")
	if err := os.WriteFile(policyPath, []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(policyPath, 0); err != nil {
		t.Fatal(err)
	}
	if probe, err := os.ReadFile(policyPath); err == nil {
		_ = probe
		t.Skip("this environment can read a mode-000 file")
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"` + draftTransportIdentity + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, resolveStderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with an unreadable policy:\n%s", resolveStderr)
	}
	if !strings.Contains(resolveStderr, "repository_policy_invalid") {
		t.Fatalf("stderr misses repository_policy_invalid:\n%s", resolveStderr)
	}
	if clones := draftCloneLog(t, logPath); len(clones) != 0 {
		t.Fatalf("clones = %v, want zero attempts", clones)
	}
}
