package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fileURL renders an absolute directory as a file:// URL in slash form so
// insteadOf rules match it on every platform.
func fileURL(t *testing.T, dir string) string {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return "file://" + filepath.ToSlash(abs)
}

// hostileGitConfig writes a user-level gitconfig that redirects every
// fetch of declared to evil, and returns its path.
func hostileGitConfig(t *testing.T, declared, evil string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hostile.gitconfig")
	body := "[url \"" + evil + "\"]\n" +
		"\tinsteadOf = " + declared + "\n" +
		"\tpushInsteadOf = " + declared + "\n" +
		"[core]\n" +
		"\tsshCommand = /nonexistent-evil-ssh\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// cloneHead clones url into a fresh destination with env and returns the
// cloned HEAD commit. It is the positive control: under hostile config
// selection the HEAD must be evil's, proving the payload is live.
func cloneHead(t *testing.T, env []string, url string) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "control")
	cmd := exec.Command("git", "clone", "--quiet", "--", url, dest)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("control clone: %v\n%s", err, out)
	}
	out, err := exec.Command("git", "-C", dest, "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("control rev-parse: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestCloneIsolatedIgnoresUserConfig(t *testing.T) {
	declared := makeRepo(t)
	declaredURL := fileURL(t, declared)
	declaredHead := gitRun(t, declared, "rev-parse", "HEAD")
	evil := makeRepo(t)
	if err := os.WriteFile(filepath.Join(evil, "SKILL.md"), []byte("evil"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, evil, "commit", "-qam", "evil")
	evilURL := fileURL(t, evil)
	evilHead := gitRun(t, evil, "rev-parse", "HEAD")
	if evilHead == declaredHead {
		t.Fatal("fixtures collide")
	}
	hostile := hostileGitConfig(t, declaredURL, evilURL)

	t.Run("global and system selectors", func(t *testing.T) {
		t.Setenv("GIT_CONFIG_GLOBAL", hostile)
		t.Setenv("GIT_CONFIG_SYSTEM", hostile)
		t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
		if got := cloneHead(t, gitEnvHostile(t), declaredURL); got != evilHead {
			t.Fatalf("control HEAD = %s, want evil %s (payload not live)", got, evilHead)
		}
		dest := filepath.Join(t.TempDir(), "clone")
		if err := CloneIsolated(declaredURL, dest); err != nil {
			t.Fatal(err)
		}
		if got := gitRun(t, dest, "rev-parse", "HEAD"); got != declaredHead {
			t.Fatalf("HEAD = %s, want declared %s (evil %s)", got, declaredHead, evilHead)
		}
	})

	t.Run("home file without selectors", func(t *testing.T) {
		for _, key := range []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_NOSYSTEM", "GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS"} {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}
		}
		home := t.TempDir()
		payload, err := os.ReadFile(hostile)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, ".gitconfig"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("HOME", home)
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		if got := cloneHead(t, gitEnvHome(t, home), declaredURL); got != evilHead {
			t.Fatalf("control HEAD = %s, want evil %s (payload not live)", got, evilHead)
		}
		dest := filepath.Join(t.TempDir(), "clone")
		if err := CloneIsolated(declaredURL, dest); err != nil {
			t.Fatal(err)
		}
		if got := gitRun(t, dest, "rev-parse", "HEAD"); got != declaredHead {
			t.Fatalf("HEAD = %s, want declared %s (evil %s)", got, declaredHead, evilHead)
		}
	})

	t.Run("environment injected config", func(t *testing.T) {
		t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "empty"))
		t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(t.TempDir(), "empty"))
		t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
		t.Setenv("GIT_CONFIG_COUNT", "1")
		t.Setenv("GIT_CONFIG_KEY_0", "url."+evilURL+".insteadOf")
		t.Setenv("GIT_CONFIG_VALUE_0", declaredURL)
		if got := cloneHead(t, gitEnvParams(t), declaredURL); got != evilHead {
			t.Fatalf("control HEAD = %s, want evil %s (payload not live)", got, evilHead)
		}
		dest := filepath.Join(t.TempDir(), "clone")
		if err := CloneIsolated(declaredURL, dest); err != nil {
			t.Fatal(err)
		}
		if got := gitRun(t, dest, "rev-parse", "HEAD"); got != declaredHead {
			t.Fatalf("HEAD = %s, want declared %s (evil %s)", got, declaredHead, evilHead)
		}
	})
}

// gitEnvHostile is gitEnv with the hostile user/system selectors applied,
// for the positive control only.
func gitEnvHostile(t *testing.T) []string {
	t.Helper()
	return append(gitEnv(),
		"GIT_CONFIG_GLOBAL="+os.Getenv("GIT_CONFIG_GLOBAL"),
		"GIT_CONFIG_SYSTEM="+os.Getenv("GIT_CONFIG_SYSTEM"),
		"GIT_CONFIG_NOSYSTEM=0",
	)
}

// gitEnvHome is committer identity with HOME pointed at home and no
// config selectors, for the positive control only.
func gitEnvHome(t *testing.T, home string) []string {
	t.Helper()
	var env []string
	for _, entry := range os.Environ() {
		name := entry
		if index := strings.IndexByte(entry, '='); index >= 0 {
			name = entry[:index]
		}
		switch name {
		case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_NOSYSTEM",
			"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS":
			continue
		}
		env = append(env, entry)
	}
	return append(env,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		"HOME="+home, "XDG_CONFIG_HOME="+os.Getenv("XDG_CONFIG_HOME"),
	)
}

// gitEnvParams is committer identity with the environment-injected
// insteadOf applied, for the positive control only.
func gitEnvParams(t *testing.T) []string {
	t.Helper()
	return append(gitEnv(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_COUNT="+os.Getenv("GIT_CONFIG_COUNT"),
		"GIT_CONFIG_KEY_0="+os.Getenv("GIT_CONFIG_KEY_0"),
		"GIT_CONFIG_VALUE_0="+os.Getenv("GIT_CONFIG_VALUE_0"),
	)
}

func TestFetchIsolatedIgnoresUserConfig(t *testing.T) {
	declared := makeRepo(t)
	declaredURL := fileURL(t, declared)
	evil := makeRepo(t)
	if err := os.WriteFile(filepath.Join(evil, "SKILL.md"), []byte("evil"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, evil, "commit", "-qam", "evil")
	evilURL := fileURL(t, evil)
	evilHead := gitRun(t, evil, "rev-parse", "HEAD")
	hostile := hostileGitConfig(t, declaredURL, evilURL)
	t.Setenv("GIT_CONFIG_GLOBAL", hostile)
	t.Setenv("GIT_CONFIG_SYSTEM", hostile)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")

	dest := filepath.Join(t.TempDir(), "clone")
	if err := CloneIsolated(declaredURL, dest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(declared, "SKILL.md"), []byte("v3"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, declared, "commit", "-qam", "three")
	declaredHead := gitRun(t, declared, "rev-parse", "HEAD")
	if err := FetchIsolated(dest); err != nil {
		t.Fatal(err)
	}
	if got := gitRun(t, dest, "rev-parse", "origin/main"); got != declaredHead {
		t.Fatalf("origin/main = %s, want declared %s (evil %s)", got, declaredHead, evilHead)
	}
}

func TestIsolatedGitEnvPinsEmptyConfig(t *testing.T) {
	hostile := filepath.Join(t.TempDir(), "hostile.gitconfig")
	if err := os.WriteFile(hostile, []byte("[url \"https://evil.test/\"]\n\tinsteadOf = https://example.org/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", hostile)
	t.Setenv("GIT_CONFIG_SYSTEM", hostile)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_PARAMETERS", "'url.evil.insteadOf=x'")
	t.Setenv("GIT_CONFIG_KEY_0", "url.evil.insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "x")
	t.Setenv("GIT_ALLOW_PROTOCOL", "all")
	t.Setenv("GIT_TERMINAL_PROMPT", "1")
	t.Setenv("SSH_AUTH_SOCK", "/tmp/kept-agent.sock")
	t.Setenv("GIT_ASKPASS", "/operator/askpass")

	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	values := map[string]string{}
	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		values[strings.ToUpper(name)] = value
	}
	global, ok := values["GIT_CONFIG_GLOBAL"]
	if !ok {
		t.Fatal("GIT_CONFIG_GLOBAL missing")
	}
	if global == hostile {
		t.Fatal("GIT_CONFIG_GLOBAL still hostile")
	}
	if values["GIT_CONFIG_SYSTEM"] != global {
		t.Fatalf("GIT_CONFIG_SYSTEM = %q, want %q", values["GIT_CONFIG_SYSTEM"], global)
	}
	info, err := os.Stat(global)
	if err != nil {
		t.Fatalf("pinned config unreadable: %v", err)
	}
	if info.Size() != 0 || !info.Mode().IsRegular() {
		t.Fatalf("pinned config = size %d mode %s, want an empty regular file", info.Size(), info.Mode())
	}
	if values["GIT_CONFIG_NOSYSTEM"] != "1" {
		t.Fatalf("GIT_CONFIG_NOSYSTEM = %q, want 1", values["GIT_CONFIG_NOSYSTEM"])
	}
	for _, forbidden := range []string{"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0"} {
		if value, ok := values[forbidden]; ok {
			t.Fatalf("%s = %q, want it dropped", forbidden, value)
		}
	}
	if values["GIT_ALLOW_PROTOCOL"] != AllowedProtocols {
		t.Fatalf("GIT_ALLOW_PROTOCOL = %q, want %q", values["GIT_ALLOW_PROTOCOL"], AllowedProtocols)
	}
	if values["GIT_TERMINAL_PROMPT"] != "0" {
		t.Fatalf("GIT_TERMINAL_PROMPT = %q, want 0", values["GIT_TERMINAL_PROMPT"])
	}
	if values["GIT_PROTOCOL_FROM_USER"] != "0" {
		t.Fatalf("GIT_PROTOCOL_FROM_USER = %q, want 0", values["GIT_PROTOCOL_FROM_USER"])
	}
	if values["LANG"] != "C" || values["LC_ALL"] != "C" {
		t.Fatalf("locale = %q/%q, want C/C", values["LANG"], values["LC_ALL"])
	}
	if values["SSH_AUTH_SOCK"] != "/tmp/kept-agent.sock" {
		t.Fatal("SSH_AUTH_SOCK was not preserved")
	}
	if values["GIT_ASKPASS"] != "/operator/askpass" {
		t.Fatal("GIT_ASKPASS was not preserved")
	}
	if values["PATH"] != os.Getenv("PATH") {
		t.Fatal("PATH was not preserved")
	}
	sshCommand, ok := values["GIT_SSH_COMMAND"]
	if !ok || !strings.Contains(sshCommand, "-F ") {
		t.Fatalf("GIT_SSH_COMMAND = %q, want the curator-owned command with -F", sshCommand)
	}
	cleanup()
	if _, err := os.Stat(global); !os.IsNotExist(err) {
		t.Fatalf("cleanup left %s: %v", global, err)
	}
}

// envValues returns every value in env whose name matches want
// case-insensitively (the Windows spelling).
func envValues(env []string, want string) []string {
	var out []string
	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		if strings.EqualFold(name, want) {
			out = append(out, value)
		}
	}
	return out
}

// TestIsolatedGitEnvAllowList proves the allow-list by construction: every
// ambient name outside the list is absent (including every GIT_* override
// and proxy name, in hostile and Windows-mixed-case spellings), and every
// listed name is present. Removing the filter (appending the ambient
// environment) must fail this test.
func TestIsolatedGitEnvAllowList(t *testing.T) {
	hostileGit := []string{
		"GIT_SSH", "GIT_SSH_COMMAND", "GIT_SSH_VARIANT",
		"GIT_PROXY_COMMAND", "GIT_EXEC_PATH",
		"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR",
		"GIT_OBJECT_DIRECTORY", "GIT_INDEX_FILE",
		"GIT_SSL_CAINFO", "GIT_SSL_NO_VERIFY",
		"GIT_CURL_VERBOSE", "GIT_TRACE", "GIT_TRACE_PACKET",
		"GIT_HTTP_USER_AGENT", "GIT_CONFIG_COUNT",
		"GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0",
		"GIT_CONFIG_PARAMETERS", "GIT_TEMPLATE_DIR",
		"GIT_NAMESPACE", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
		"GIT_PAGER", "GIT_EDITOR", "GIT_AUTHOR_NAME",
	}
	for _, name := range hostileGit {
		t.Setenv(name, "evil-"+name)
	}
	// Windows-mixed-case spellings of the same overrides.
	t.Setenv("git_ssh_command", "evil-lower")
	t.Setenv("Git_Proxy_Command", "evil-mixed")
	t.Setenv("gIt_eXeC_pAtH", "evil-mixed-exec")
	for _, name := range []string{
		"http_proxy", "https_proxy", "all_proxy", "no_proxy",
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY",
	} {
		t.Setenv(name, "http://127.0.0.1:9/")
	}
	t.Setenv("CURATOR_TEST_KEEP", "kept")
	t.Setenv("SSH_ASKPASS", "/tmp/evil-ssh-askpass")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/evil-xdg")
	t.Setenv("LANG", "evil.UTF-8")
	t.Setenv("LC_ALL", "evil.UTF-8")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("TMP", t.TempDir())
	t.Setenv("TEMP", t.TempDir())
	t.Setenv("TZ", "UTC")
	t.Setenv("SYSTEMROOT", `C:\Windows`)
	t.Setenv("WINDIR", `C:\Windows`)
	t.Setenv("COMSPEC", `C:\Windows\System32\cmd.exe`)
	t.Setenv("PATHEXT", ".EXE")
	t.Setenv("SYSTEMDRIVE", "C:")
	t.Setenv("LOCALAPPDATA", `C:\evil`)
	t.Setenv("APPDATA", `C:\evil-roam`)
	t.Setenv("SSH_AUTH_SOCK", "/tmp/kept-agent.sock")
	t.Setenv("GIT_ASKPASS", "/operator/askpass")

	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, name := range hostileGit {
		if name == "GIT_SSH_COMMAND" {
			continue
		}
		if got := envValues(env, name); len(got) != 0 {
			t.Fatalf("%s = %q, want it absent by construction", name, got)
		}
	}
	// GIT_SSH_COMMAND ambient values (every spelling) are replaced by the
	// single curator-owned command: hostile payloads never survive.
	sshValues := envValues(env, "GIT_SSH_COMMAND")
	if len(sshValues) != 1 {
		t.Fatalf("GIT_SSH_COMMAND entries = %q, want exactly the pinned command", sshValues)
	}
	for _, hostile := range []string{"evil-GIT_SSH_COMMAND", "evil-lower"} {
		if strings.Contains(sshValues[0], hostile) {
			t.Fatalf("GIT_SSH_COMMAND carries the ambient payload %q: %q", hostile, sshValues[0])
		}
	}
	for _, name := range []string{
		"http_proxy", "https_proxy", "all_proxy", "no_proxy",
		"CURATOR_TEST_KEEP", "SSH_ASKPASS", "XDG_CONFIG_HOME",
		"LOCALAPPDATA", "APPDATA",
	} {
		if got := envValues(env, name); len(got) != 0 {
			t.Fatalf("%s = %q, want it absent by construction", name, got)
		}
	}
	for name, want := range map[string]string{
		"PATH": os.Getenv("PATH"), "HOME": home, "USERPROFILE": home,
		"TZ": "UTC", "SYSTEMROOT": `C:\Windows`, "WINDIR": `C:\Windows`,
		"COMSPEC": `C:\Windows\System32\cmd.exe`, "PATHEXT": ".EXE",
		"SYSTEMDRIVE": "C:", "SSH_AUTH_SOCK": "/tmp/kept-agent.sock",
		"GIT_ASKPASS": "/operator/askpass",
	} {
		got := envValues(env, name)
		if len(got) != 1 || got[0] != want {
			t.Fatalf("%s = %q, want [%q]", name, got, want)
		}
	}
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		if got := envValues(env, name); len(got) != 1 || got[0] == "" {
			t.Fatalf("%s = %q, want it preserved", name, got)
		}
	}
	if got := envValues(env, "LANG"); len(got) != 1 || got[0] != "C" {
		t.Fatalf("LANG = %q, want [C] (pinned, not honoured)", got)
	}
	if got := envValues(env, "LC_ALL"); len(got) != 1 || got[0] != "C" {
		t.Fatalf("LC_ALL = %q, want [C] (pinned, not honoured)", got)
	}
}

// TestIsolatedGitEnvWindowsCaseHonoursAllowed proves the allow-list matches
// case-insensitively in the honour direction too: a Windows-cased allowed
// spelling is preserved. On POSIX this coexists with the canonical entry;
// on Windows there is only one entry and it is the same file.
func TestIsolatedGitEnvWindowsCaseHonoursAllowed(t *testing.T) {
	t.Setenv("sSh_AuTh_SoCk", "/tmp/mixed-agent.sock")
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	got := envValues(env, "SSH_AUTH_SOCK")
	found := false
	for _, value := range got {
		if value == "/tmp/mixed-agent.sock" {
			found = true
		}
	}
	if !found {
		t.Fatalf("SSH_AUTH_SOCK entries = %q, want the mixed-case ambient value preserved", got)
	}
}

// TestIsolatedGitEnvSSHCommand pins the curator-owned ssh command: -F with
// a fresh empty config, BatchMode, host-key validation never disabled, and
// ProxyCommand/Jump none. Dropping -F or ProxyCommand=none must fail.
func TestIsolatedGitEnvSSHCommand(t *testing.T) {
	t.Setenv("GIT_SSH", "/tmp/evil-ssh")
	t.Setenv("GIT_SSH_COMMAND", "/tmp/evil-ssh-command")
	env, cleanup, err := isolatedGitEnv()
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, entry := range env {
		name, value, _ := strings.Cut(entry, "=")
		values[strings.ToUpper(name)] = value
	}
	command, ok := values["GIT_SSH_COMMAND"]
	if !ok {
		t.Fatal("GIT_SSH_COMMAND missing")
	}
	for _, evil := range []string{"/tmp/evil-ssh", "/tmp/evil-ssh-command"} {
		if strings.Contains(command, evil) {
			t.Fatalf("GIT_SSH_COMMAND carries the ambient payload: %q", command)
		}
	}
	for _, want := range []string{
		"ssh -F ", "BatchMode=yes", "StrictHostKeyChecking=yes",
		"ProxyCommand=none", "ProxyJump=none", "PermitLocalCommand=no",
		"ForwardAgent=no", "ClearAllForwardings=yes", "RequestTTY=no",
		"CanonicalizeHostname=no", "UpdateHostKeys=no", "ConnectionAttempts=1",
	} {
		if !strings.Contains(command, want) {
			t.Fatalf("GIT_SSH_COMMAND misses %q: %q", want, command)
		}
	}
	for _, forbidden := range []string{
		"StrictHostKeyChecking=no", "StrictHostKeyChecking=accept-new",
		"StrictHostKeyChecking=off",
	} {
		if strings.Contains(command, forbidden) {
			t.Fatalf("GIT_SSH_COMMAND disables host-key validation: %q", command)
		}
	}
	if strings.Contains(command, "IdentityFile") || strings.Contains(command, "UserKnownHostsFile") {
		t.Fatalf("GIT_SSH_COMMAND pins identities/known_hosts, want the compiled-in defaults: %q", command)
	}
	// The -F path names a fresh empty regular file, distinct from the
	// pinned git config, and cleanup removes both.
	gitConfig := values["GIT_CONFIG_GLOBAL"]
	quoted := strings.TrimPrefix(strings.SplitN(command[strings.Index(command, "-F ")+3:], " -o ", 2)[0], "")
	sshConfig := strings.Trim(quoted, "'")
	sshConfig = strings.ReplaceAll(sshConfig, "'\\''", "'")
	if sshConfig == "" || sshConfig == gitConfig {
		t.Fatalf("ssh config %q collides with the git config %q", sshConfig, gitConfig)
	}
	info, err := os.Stat(sshConfig)
	if err != nil {
		t.Fatalf("pinned ssh config unreadable: %v", err)
	}
	if info.Size() != 0 || !info.Mode().IsRegular() {
		t.Fatalf("pinned ssh config = size %d mode %s, want an empty regular file", info.Size(), info.Mode())
	}
	cleanup()
	if _, err := os.Stat(sshConfig); !os.IsNotExist(err) {
		t.Fatalf("cleanup left %s: %v", sshConfig, err)
	}
	if _, err := os.Stat(gitConfig); !os.IsNotExist(err) {
		t.Fatalf("cleanup left %s: %v", gitConfig, err)
	}
}

// TestIsolatedGitConfigArgs pins the per-invocation transport flags before
// the subcommand.
func TestIsolatedGitConfigArgs(t *testing.T) {
	got := withIsolatedConfigArgs([]string{"clone", "--", "https://example.org/kit.git", "dest"})
	want := []string{
		"-c", "http.sslVerify=true",
		"-c", "http.followRedirects=false",
		"-c", "credential.helper=",
		"clone", "--", "https://example.org/kit.git", "dest",
	}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %q, want %q", got, want)
	}
}

// TestShellQuoteEscapesSingleQuote proves the temp ssh-config path cannot
// break out of the GIT_SSH_COMMAND quoting.
func TestShellQuoteEscapesSingleQuote(t *testing.T) {
	if got := shellQuote("/tmp/a'b"); got != "'/tmp/a'\\''b'" {
		t.Fatalf("shellQuote = %q", got)
	}
	if got := draftSSHCommand("/tmp/a'b"); !strings.Contains(got, "'/tmp/a'\\''b'") {
		t.Fatalf("ssh command = %q", got)
	}
}

// TestCloneIsolatedIgnoresGitProxyCommand is the behavioral git:// leg of
// row (b): GIT_PROXY_COMMAND applies only to the git:// transport, which
// the Skillfile and policy grammars never admit, so no project-resolve
// production entry can reach it. At the gitops layer with real git, an
// ambient proxy command that would intercept the git:// dial must never be
// invoked: the lane dials directly and fails closed on fixture.test.
func TestCloneIsolatedIgnoresGitProxyCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	evilDir := t.TempDir()
	invokedLog := filepath.Join(evilDir, "proxy-invoked.log")
	evilProxy := filepath.Join(evilDir, "evil-proxy")
	if err := os.WriteFile(evilProxy, []byte("#!/bin/sh\necho \"invoked $@\" >>'"+invokedLog+"'\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	// Positive control: raw git with the ambient proxy honoured invokes
	// it, proving the payload is live.
	controlEnv := append(gitEnv(), "GIT_PROXY_COMMAND="+evilProxy)
	controlDest := filepath.Join(t.TempDir(), "control")
	cmd := exec.Command("git", "clone", "--quiet", "--", "git://fixture.test/kit.git", controlDest)
	cmd.Env = controlEnv
	if err := cmd.Run(); err == nil {
		t.Fatal("control clone succeeded, want failure (payload not live)")
	}
	if probe, err := os.ReadFile(invokedLog); err != nil || len(probe) == 0 {
		t.Fatal("control left no proxy invocation (payload not live)")
	}
	_ = os.Remove(invokedLog)

	t.Setenv("GIT_PROXY_COMMAND", evilProxy)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := CloneIsolated("git://fixture.test/kit.git", dest); err == nil {
		t.Fatal("CloneIsolated with a hostile GIT_PROXY_COMMAND succeeded, want the fail-closed error")
	}
	if probe, err := os.ReadFile(invokedLog); err == nil && len(probe) != 0 {
		t.Fatalf("evil GIT_PROXY_COMMAND was invoked: %q", probe)
	}
}
