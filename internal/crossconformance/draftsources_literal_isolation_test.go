package crossconformance

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testcli"
)

// Draft literal-lane environment and ssh isolation (TASK-260920-3ccq6b).
//
// These rows are intentionally NOT corpus rows (no registerDraftSemantic
// calls): the v2 semantic matrix stays 94 and
// TestDraftSourcesSemanticCoverage is unaffected. The spec's
// v2-user-ssh-alias-ignored case covers only the logical declaration
// (planning refusal before any fetch); the literal ssh:// fetch isolation
// below has no v2 semantic case, so local rows carry it.
//
// Every row drives the production entry (the compiled CLI's project
// resolve/refresh, which runs CloneIsolated/FetchIsolated) with real git
// on POSIX (shell shims; Windows skips like the sibling transport rows)
// and asserts the fail-closed class, the evil fixture untouched, and
// sanitized diagnostics. Mutants: remove the allow-list filter, drop -F,
// drop ProxyCommand=none, pass the ambient GIT_SSH_COMMAND through — each
// is killed by the named row below.

const literalIsolationSSH = "ssh://git@fixture.test/kit.git"

const literalIsolationHTTPS = "https://fixture.test/kit.git"

// literalIsolationGitShim writes a fake git that logs every clone URL plus
// the selected environment channels, then either rewrites rewriteURL to the
// bare fixture (when set) or execs real git unchanged (direct network
// attempt, for the fail-closed rows). It returns the PATH directory, the
// clone log, and the env log.
func literalIsolationGitShim(t *testing.T, rewriteURL, bare, realGit string) (fakeDir, cloneLog, envLog string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	fakeDir = t.TempDir()
	cloneLog = filepath.Join(fakeDir, "clones.log")
	envLog = filepath.Join(fakeDir, "env.log")
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("cloneLog='" + cloneLog + "'\n")
	script.WriteString("envLog='" + envLog + "'\n")
	script.WriteString("is_clone=0; for arg in \"$@\"; do if [ \"$arg\" = \"clone\" ]; then is_clone=1; fi; done\n")
	script.WriteString("if [ \"$is_clone\" = \"1\" ]; then\n")
	script.WriteString("url=''; prev=''; for arg in \"$@\"; do if [ \"$prev\" = \"--\" ]; then url=\"$arg\"; break; fi; prev=\"$arg\"; done\n")
	script.WriteString("printf '%s\\n' \"$url\" >>\"$cloneLog\"\n")
	script.WriteString("{ printf 'GIT_SSH_COMMAND=%s\\n' \"${GIT_SSH_COMMAND-<absent>}\"; printf 'GIT_SSH=%s\\n' \"${GIT_SSH-<absent>}\"; printf 'GIT_PROXY_COMMAND=%s\\n' \"${GIT_PROXY_COMMAND-<absent>}\"; printf 'GIT_EXEC_PATH=%s\\n' \"${GIT_EXEC_PATH-<absent>}\"; printf 'SSH_AUTH_SOCK=%s\\n' \"${SSH_AUTH_SOCK-<absent>}\"; printf 'GIT_ASKPASS=%s\\n' \"${GIT_ASKPASS-<absent>}\"; printf 'https_proxy=%s\\n' \"${https_proxy-<absent>}\"; printf 'HTTPS_PROXY=%s\\n' \"${HTTPS_PROXY-<absent>}\"; } >>\"$envLog\"\n")
	if rewriteURL != "" {
		script.WriteString("if [ \"$url\" = '" + rewriteURL + "' ]; then\n")
		script.WriteString("args=''; for arg in \"$@\"; do if [ \"$arg\" = '" + rewriteURL + "' ]; then arg='file://" + bare + "'; fi\n")
		script.WriteString("args=\"$args\n$arg\"; done\n")
		script.WriteString("oldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\n")
		script.WriteString("exec '" + realGit + "' \"$@\"; fi\n")
	}
	script.WriteString("fi\n")
	script.WriteString("exec '" + realGit + "' \"$@\"\n")
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	return fakeDir, cloneLog, envLog
}

func literalIsolationEnvLog(t *testing.T, envLog string) string {
	t.Helper()
	payload, err := os.ReadFile(envLog)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatal(err)
	}
	return string(payload)
}

// writeEvilSSHCommand writes an evil GIT_SSH_COMMAND target: it logs every
// invocation, then serves the evil bare repository over the ssh transport
// by exec'ing upload-pack locally, ignoring the requested host. When
// honoured, an ssh:// clone succeeds with the evil commit; when ignored
// (the isolated lane), it is never invoked.
func writeEvilSSHCommand(t *testing.T, evilBare, realGit string) (evilSSH, invokedLog string) {
	t.Helper()
	dir := t.TempDir()
	invokedLog = filepath.Join(dir, "evil-ssh-invoked.log")
	evilSSH = filepath.Join(dir, "evil-ssh")
	script := "#!/bin/sh\n" +
		"echo \"invoked $@\" >>'" + invokedLog + "'\n" +
		"exec '" + realGit + "' upload-pack '" + evilBare + "'\n"
	if err := os.WriteFile(evilSSH, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return evilSSH, invokedLog
}

// writeEvilBare builds a second kit fixture whose tag v1 binds a different
// commit from the declared one.
func writeEvilBare(t *testing.T, root string, declaredCommit string) (evilBare, evilCommit string) {
	t.Helper()
	evilWork := filepath.Join(root, "evil-work")
	if err := os.MkdirAll(evilWork, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(evilWork, "skills", "review"), "review")
	if err := os.WriteFile(filepath.Join(evilWork, "skills", "review", "references", "evil.md"), []byte("evil"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, evilWork, "init", "-q", "-b", "main")
	runDraftGit(t, evilWork, "add", ".")
	runDraftGit(t, evilWork, "commit", "-qm", "evil")
	runDraftGit(t, evilWork, "tag", "v1")
	evilBare = filepath.Join(root, "evil.git")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", evilWork, evilBare)
	evilCommit = draftGitOutput(t, "", "--git-dir", evilBare, "rev-parse", "v1^{commit}")
	if evilCommit == declaredCommit {
		t.Fatal("declared and evil fixtures collide")
	}
	return evilBare, evilCommit
}

func assertSanitized(t *testing.T, stderr string, forbidden ...string) {
	t.Helper()
	for _, leak := range forbidden {
		if leak != "" && strings.Contains(stderr, leak) {
			t.Fatalf("stderr leaks %q:\n%s", leak, stderr)
		}
	}
}

// TestDraftLiteralIgnoresGitSSHCommand is row (a): an ambient
// GIT_SSH_COMMAND pointing at a script that would serve the evil
// repository must not redirect the literal ssh:// clone. The lane fails
// closed on the declared endpoint (no network to fixture.test) with the
// evil script never invoked. Mutant "pass the ambient GIT_SSH_COMMAND
// through" succeeds with the evil commit and is KILLED by the exit code,
// the missing lock, and the invocation log.
func TestDraftLiteralIgnoresGitSSHCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	_, declaredCommit := draftCLIKitBare(t, root)
	evilBare, evilCommit := writeEvilBare(t, root, declaredCommit)
	evilSSH, invokedLog := writeEvilSSHCommand(t, evilBare, realGit)
	// Positive control: raw git with the ambient evil command honoured
	// binds the evil commit, proving the payload is live.
	control := filepath.Join(t.TempDir(), "control")
	if code, _, stderr := testcli.Run(t, "", []string{"GIT_SSH_COMMAND=" + evilSSH}, "", realGit, "clone", "--quiet", "--", literalIsolationSSH, control); code != 0 {
		t.Fatalf("control clone with honoured GIT_SSH_COMMAND = %d, want success (payload not live):\n%s", code, stderr)
	}
	if got := draftGitOutput(t, control, "rev-parse", "HEAD"); got != evilCommit {
		t.Fatalf("control HEAD = %s, want evil %s (payload not live)", got, evilCommit)
	}
	if probe, err := os.ReadFile(invokedLog); err != nil || len(probe) == 0 {
		t.Fatal("control left no evil-ssh invocation (payload not live)")
	}
	_ = os.Remove(invokedLog)

	fakeDir, cloneLog, envLog := literalIsolationGitShim(t, "", "", realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationSSH + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "GIT_SSH_COMMAND="+evilSSH, "GIT_SSH="+evilSSH)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with a hostile GIT_SSH_COMMAND, want the endpoint refusal:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	assertSanitized(t, stderr, literalIsolationSSH, evilSSH, evilBare, "evil-ssh")
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationSSH {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	// The pinned curator-owned command reaches git, never the ambient evil.
	envSeen := literalIsolationEnvLog(t, envLog)
	if strings.Contains(envSeen, evilSSH) {
		t.Fatalf("git saw the ambient GIT_SSH_COMMAND:\n%s", envSeen)
	}
	if !strings.Contains(envSeen, "ssh -F ") || !strings.Contains(envSeen, "BatchMode=yes") {
		t.Fatalf("git missed the pinned curator-owned command:\n%s", envSeen)
	}
	if probe, err := os.ReadFile(invokedLog); err == nil && len(probe) != 0 {
		t.Fatalf("evil GIT_SSH_COMMAND was invoked: %q", probe)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

// TestDraftLiteralIgnoresGitSSH is the GIT_SSH half of row (a): git prefers
// GIT_SSH_COMMAND, but the allow-list drops both, so an ambient GIT_SSH
// alone must also never redirect the clone.
func TestDraftLiteralIgnoresGitSSH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	_, declaredCommit := draftCLIKitBare(t, root)
	evilBare, _ := writeEvilBare(t, root, declaredCommit)
	evilSSH, invokedLog := writeEvilSSHCommand(t, evilBare, realGit)
	fakeDir, cloneLog, envLog := literalIsolationGitShim(t, "", "", realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationSSH + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "GIT_SSH="+evilSSH)
	code, _, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with a hostile GIT_SSH, want the endpoint refusal:\n%s", stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationSSH {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	if envSeen := literalIsolationEnvLog(t, envLog); strings.Contains(envSeen, evilSSH) {
		t.Fatalf("git saw the ambient GIT_SSH:\n%s", envSeen)
	}
	if probe, err := os.ReadFile(invokedLog); err == nil && len(probe) != 0 {
		t.Fatalf("evil GIT_SSH was invoked: %q", probe)
	}
}

// TestDraftLiteralDropsGitProxyCommand is row (b). GIT_PROXY_COMMAND
// (core.gitProxy) applies only to the git:// transport, which neither the
// Skillfile git: grammar nor the policy endpoint grammar admits, so a
// git:// production entry through project resolve is infeasible; the
// behavioral git:// proof lives at the gitops layer
// (TestCloneIsolatedIgnoresGitProxyCommand). This production-entry row
// proves the allow-list drops the ambient value on the admitted https://
// lane: the resolve succeeds with the declared bind and git never sees
// the hostile value. Mutant "remove the allow-list filter" exposes the
// evil value in the env log and is KILLED.
func TestDraftLiteralDropsGitProxyCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	bare, _ := draftCLIKitBare(t, root)
	evilProxy := filepath.Join(t.TempDir(), "evil-proxy")
	if err := os.WriteFile(evilProxy, []byte("#!/bin/sh\necho evil-proxy-invoked >&2\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	fakeDir, cloneLog, envLog := literalIsolationGitShim(t, literalIsolationHTTPS, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationHTTPS + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "GIT_PROXY_COMMAND="+evilProxy)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code != 0 {
		t.Fatalf("resolve = %d, want success with the proxy override dropped:\n%s\n%s", code, stdout, stderr)
	}
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationHTTPS {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	if envSeen := literalIsolationEnvLog(t, envLog); strings.Contains(envSeen, evilProxy) {
		t.Fatalf("git saw the ambient GIT_PROXY_COMMAND:\n%s", envSeen)
	}
}

// TestDraftLiteralIgnoresGitExecPath is row (c): an ambient GIT_EXEC_PATH
// pointing at a directory with a poisoned git-remote-https shim must not
// hijack the literal https:// clone. The lane uses the real helper, fails
// closed on the declared endpoint (no network), and never invokes the
// poisoned shim. Mutant "remove the allow-list filter" invokes it and is
// KILLED by the invocation log.
func TestDraftLiteralIgnoresGitExecPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	evilDir := t.TempDir()
	invokedLog := filepath.Join(evilDir, "remote-https-invoked.log")
	poisoned := "#!/bin/sh\necho \"invoked $@\" >>'" + invokedLog + "'\necho 'evil remote helper' >&2\nexit 128\n"
	if err := os.WriteFile(filepath.Join(evilDir, "git-remote-https"), []byte(poisoned), 0o700); err != nil {
		t.Fatal(err)
	}
	// Positive control: raw git with the ambient exec-path honoured
	// invokes the poisoned helper, proving the payload is live.
	if code, _, _ := testcli.Run(t, "", []string{"GIT_EXEC_PATH=" + evilDir}, "", realGit, "ls-remote", literalIsolationHTTPS); code == 0 {
		t.Fatal("control ls-remote succeeded, want failure (payload not live)")
	}
	if probe, err := os.ReadFile(invokedLog); err != nil || len(probe) == 0 {
		t.Fatal("control left no poisoned-helper invocation (payload not live)")
	}
	_ = os.Remove(invokedLog)

	fakeDir, cloneLog, envLog := literalIsolationGitShim(t, "", "", realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationHTTPS + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "GIT_EXEC_PATH="+evilDir)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with a hostile GIT_EXEC_PATH, want the endpoint refusal:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	assertSanitized(t, stderr, literalIsolationHTTPS, evilDir, "evil remote helper")
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationHTTPS {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	if envSeen := literalIsolationEnvLog(t, envLog); strings.Contains(envSeen, evilDir) {
		t.Fatalf("git saw the ambient GIT_EXEC_PATH:\n%s", envSeen)
	}
	if probe, err := os.ReadFile(invokedLog); err == nil && len(probe) != 0 {
		t.Fatalf("poisoned git-remote-https was invoked: %q", probe)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

// writeFakeSSH writes a stand-in ssh that logs its argv (one <arg> per word
// per line) and fails closed like an unknown host, so the lane's ssh
// argv is observable at the production entry without network.
func writeFakeSSH(t *testing.T, fakeDir string) (argvLog string) {
	t.Helper()
	argvLog = filepath.Join(fakeDir, "ssh-argv.log")
	script := "#!/bin/sh\n" +
		"{ for arg in \"$@\"; do printf '<%s>' \"$arg\"; done; printf '\\n'; } >>'" + argvLog + "'\n" +
		"echo 'ssh: Could not resolve hostname fixture.test: Name or service not known' >&2\n" +
		"exit 255\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "ssh"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return argvLog
}

// TestDraftLiteralIgnoresUserSSHConfig is row (d): a hostile ~/.ssh/config
// (Host alias to an evil HostName plus a ProxyCommand) under a temp HOME
// must not redirect the literal ssh:// clone. The lane runs the
// curator-owned command (-F empty, ProxyCommand=none, host-key validation
// never disabled) resolved through PATH, fails closed on the declared
// endpoint, and never invokes the evil ProxyCommand. Mutants "drop -F"
// and "drop ProxyCommand=none" lose their argv markers and are KILLED;
// "pass the ambient GIT_SSH_COMMAND through" bypasses the fake ssh and is
// KILLED by its silence plus the evil invocation.
func TestDraftLiteralIgnoresUserSSHConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	fakeDir, cloneLog, _ := literalIsolationGitShim(t, "", "", realGit)
	sshArgvLog := writeFakeSSH(t, fakeDir)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	evilProxy := filepath.Join(root, "evil-proxy")
	proxyInvoked := filepath.Join(root, "evil-proxy-invoked.log")
	if err := os.WriteFile(evilProxy, []byte("#!/bin/sh\necho \"invoked $@\" >>'"+proxyInvoked+"'\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	dotSSH := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dotSSH, 0o700); err != nil {
		t.Fatal(err)
	}
	hostileSSH := "Host fixture.test\n\tHostName evil.test\n\tProxyCommand " + evilProxy + " %h %p\n"
	if err := os.WriteFile(filepath.Join(dotSSH, "config"), []byte(hostileSSH), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationSSH + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with a hostile ~/.ssh/config, want the endpoint refusal:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	assertSanitized(t, stderr, literalIsolationSSH, "evil.test", evilProxy, "evil-proxy")
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationSSH {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	argv, err := os.ReadFile(sshArgvLog)
	if err != nil || len(argv) == 0 {
		t.Fatalf("fake ssh was never invoked (argv log missing): %v", err)
	}
	for _, want := range []string{"<-F>", "<BatchMode=yes>", "<StrictHostKeyChecking=yes>", "<ProxyCommand=none>", "<ProxyJump=none>"} {
		if !strings.Contains(string(argv), want) {
			t.Fatalf("ssh argv misses %s:\n%s", want, argv)
		}
	}
	if strings.Contains(string(argv), "StrictHostKeyChecking=no") || strings.Contains(string(argv), "accept-new") {
		t.Fatalf("ssh argv disables host-key validation:\n%s", argv)
	}
	if probe, err := os.ReadFile(proxyInvoked); err == nil && len(probe) != 0 {
		t.Fatalf("evil ProxyCommand was invoked: %q", probe)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

// TestDraftSSHEmptyConfigIgnoresHostileFile proves the -F empty mechanism
// with real ssh (no HOME dependence: OpenSSH resolves ~/.ssh/config from
// the passwd entry, not $HOME, on macOS and Linux): a hostile file that
// ssh honours via -F is ignored via -F empty, and ProxyCommand=none
// overrides a hostile ProxyCommand. This is the behavioral leg behind row
// (d)'s temp-HOME production entry.
func TestDraftSSHEmptyConfigIgnoresHostileFile(t *testing.T) {
	ssh, ok := testcli.LookPath("ssh")
	if !ok {
		t.Fatal("ssh is not available")
	}
	hostile := filepath.Join(t.TempDir(), "hostile-ssh-config")
	if err := os.WriteFile(hostile, []byte("Host fixture.test\n\tHostName evil.test\n\tProxyCommand /tmp/evil-proxy %h %p\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(t.TempDir(), "empty-ssh-config")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	honoured := testcli.Output(t, "", nil, ssh, "-F", hostile, "-G", "fixture.test")
	if !strings.Contains(honoured, "hostname evil.test") || !strings.Contains(strings.ToLower(honoured), "proxycommand") {
		t.Fatalf("control ssh -F hostile misses the evil mapping (payload not live):\n%s", honoured)
	}
	isolated := testcli.Output(t, "", nil, ssh, "-F", empty, "-G", "fixture.test")
	if strings.Contains(isolated, "evil.test") || strings.Contains(isolated, "evil-proxy") {
		t.Fatalf("ssh -F empty consulted the hostile file:\n%s", isolated)
	}
	overridden := testcli.Output(t, "", nil, ssh, "-F", hostile, "-o", "ProxyCommand=none", "-G", "fixture.test")
	if strings.Contains(overridden, "evil-proxy") {
		t.Fatalf("ProxyCommand=none did not override the hostile file:\n%s", overridden)
	}
}

// proxyListener starts a TCP listener that counts accepted connections.
// It runs on every OS (no shell), so row (e) needs no Windows skip.
func proxyListener(t *testing.T) (addr string, count *atomic.Int64, closeFn func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("this host cannot create a loopback listener the proxy row needs: %v", err)
	}
	count = &atomic.Int64{}
	go func() {
		for {
			if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			count.Add(1)
			_ = conn.Close()
		}
	}()
	return listener.Addr().String(), count, func() {
		_ = listener.Close()
	}
}

// TestDraftLiteralIgnoresProxyEnvironment is row (e): ambient proxy
// variables pointing at a recording listener must produce zero connections
// from the draft lane, which fails closed with a direct-DNS diagnostic
// (not a proxy diagnostic). Mutant "remove the allow-list filter" connects
// to the proxy and is KILLED by the connection count.
func TestDraftLiteralIgnoresProxyEnvironment(t *testing.T) {
	realGit := requireGit(t)
	root := t.TempDir()
	// Positive control on its own listener: raw git honours the proxy
	// variables, proving the payload is live.
	controlAddr, controlCount, closeControl := proxyListener(t)
	controlProxy := "http://" + controlAddr
	controlEnv := []string{
		"http_proxy=" + controlProxy, "https_proxy=" + controlProxy,
		"HTTP_PROXY=" + controlProxy, "HTTPS_PROXY=" + controlProxy,
		"all_proxy=" + controlProxy, "ALL_PROXY=" + controlProxy,
	}
	code, _, controlStderr := testcli.Run(t, "", controlEnv, "", realGit, "ls-remote", literalIsolationHTTPS)
	if code == 0 {
		closeControl()
		t.Fatalf("control ls-remote succeeded, want failure:\n%s", controlStderr)
	}
	deadline := time.Now().Add(5 * time.Second)
	for controlCount.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	closeControl()
	if controlCount.Load() == 0 {
		t.Fatalf("control left zero proxy connections (payload not live):\n%s", controlStderr)
	}
	if !strings.Contains(controlStderr, "Proxy") && !strings.Contains(controlStderr, "proxy") {
		// curl-version phrasing (observed macOS: "Recv failure:
		// Connection reset by peer" with no proxy word); the accepted
		// connection above is the verdict that git honoured the proxy,
		// since only git learned this listener's ephemeral port.
		t.Logf("control stderr lacks proxy wording; connection count %d is the verdict:\n%s", controlCount.Load(), controlStderr)
	}

	addr, count, closeListener := proxyListener(t)
	defer closeListener()
	proxy := "http://" + addr
	// On Windows the POSIX git shim is unavailable; the lane still runs
	// real git directly and the proxy count is the verdict either way.
	var pathEnv []string
	var cloneLog, envLog string
	if runtime.GOOS != "windows" {
		var fakeDir string
		fakeDir, cloneLog, envLog = literalIsolationGitShim(t, "", "", realGit)
		pathEnv = draftTransportPATH(t, fakeDir)
	}
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationHTTPS + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv,
		"http_proxy="+proxy, "https_proxy="+proxy,
		"HTTP_PROXY="+proxy, "HTTPS_PROXY="+proxy,
		"all_proxy="+proxy, "ALL_PROXY="+proxy,
		"no_proxy=", "NO_PROXY=",
	)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded with a hostile proxy, want the endpoint refusal:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	assertSanitized(t, stderr, literalIsolationHTTPS, proxy, addr)
	if runtime.GOOS != "windows" {
		if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationHTTPS {
			t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
		}
		if envSeen := literalIsolationEnvLog(t, envLog); strings.Contains(envSeen, proxy) {
			t.Fatalf("git saw the ambient proxy environment:\n%s", envSeen)
		}
	}
	// Give a leaked proxy dial time to arrive; the lane must stay silent.
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if count.Load() != 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if got := count.Load(); got != 0 {
		t.Fatalf("proxy connections = %d, want zero: the draft lane honoured the proxy environment", got)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

// TestDraftLiteralKeepsAgentAndAskpass is row (g): SSH_AUTH_SOCK and
// GIT_ASKPASS still pass through on the draft lane. The resolve succeeds
// with both set (file:// rewrite, so neither credential is consumed) and
// git observes both values. Dropping either (narrowing the allow-list)
// fails the env assertions.
func TestDraftLiteralKeepsAgentAndAskpass(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit := requireGit(t)
	root := t.TempDir()
	bare, commit := draftCLIKitBare(t, root)
	fakeDir, cloneLog, envLog := literalIsolationGitShim(t, literalIsolationHTTPS, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + literalIsolationHTTPS + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	agentSock := filepath.Join(t.TempDir(), "agent.sock")
	askpass := filepath.Join(t.TempDir(), "askpass")
	if err := os.WriteFile(askpass, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "SSH_AUTH_SOCK="+agentSock, "GIT_ASKPASS="+askpass)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code != 0 {
		t.Fatalf("resolve = %d, want success with agent/askpass set:\n%s\n%s", code, stdout, stderr)
	}
	if clones := draftCloneLog(t, cloneLog); len(clones) != 1 || clones[0] != literalIsolationHTTPS {
		t.Fatalf("clones = %v, want the declared literal endpoint once", clones)
	}
	envSeen := literalIsolationEnvLog(t, envLog)
	if !strings.Contains(envSeen, "SSH_AUTH_SOCK="+agentSock) {
		t.Fatalf("git missed SSH_AUTH_SOCK:\n%s", envSeen)
	}
	if !strings.Contains(envSeen, "GIT_ASKPASS="+askpass) {
		t.Fatalf("git missed GIT_ASKPASS:\n%s", envSeen)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok || member.Package.Commit.Hex != commit {
		t.Fatalf("lock binds %+v, want the declared commit %s", member.Package, commit)
	}
	// Refresh (the fetch half) succeeds with the same credentials.
	code, stdout, stderr = runCurator(t, home, configPath, env, "project", "refresh", "app")
	if code != 0 {
		t.Fatalf("refresh = %d, want success with agent/askpass set:\n%s\n%s", code, stdout, stderr)
	}
}
