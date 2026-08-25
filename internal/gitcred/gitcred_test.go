package gitcred

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

// The credential machinery is exercised against a stand-in Git that
// implements the credential protocol the way a helper-backed Git does: it
// answers `fill` from a store, writes on `approve`, removes on `reject`, and
// records the argv, environment and payload of every call. The stand-in is
// this test binary re-executed, so the same test runs on every platform
// without a shell script or a build step.
const (
	fakeGitDirEnv  = "CURATOR_GITCRED_FAKE_DIR"
	fakeGitModeEnv = "CURATOR_GITCRED_FAKE_MODE"

	// modeFail is a Git that cannot answer at all.
	modeFail = "fail"
	// modeSilentApprove is the failure this package exists to catch: a
	// helper that reports a stored credential and persists nothing.
	modeSilentApprove = "silent-approve"
	// modeIgnoreUsername is a helper that answers a username it was not
	// asked about, which is how an operator's own credential could be
	// mistaken for a manager entry.
	modeIgnoreUsername = "ignore-username"
	// modeStickyReject is a helper that accepts a rejection and keeps the
	// credential anyway.
	modeStickyReject = "sticky-reject"
	// modeHang is a helper waiting on something nobody will answer.
	modeHang = "hang"
	// modeHangGrandchild is the shape the real wedged helper has: Git is
	// killable, but the helper it started is a separate process holding the
	// output it inherited, and it outlives the kill.
	modeHangGrandchild = "hang-grandchild"
	// modeSleeper is that helper. It answers nothing and records nothing; it
	// exists only to still be there after Git is gone.
	modeSleeper = "sleeper"
)

func TestMain(m *testing.M) {
	if dir := os.Getenv(fakeGitDirEnv); dir != "" {
		os.Exit(fakeGitMain(dir))
	}
	os.Exit(m.Run())
}

func TestNamespacedEntryIsSeparateFromTheOperatorsOwn(t *testing.T) {
	if got, want := NamespaceUsername("gitlab.example.com/portals"), "curator-build-https:gitlab.example.com/portals"; got != want {
		t.Fatalf("NamespaceUsername = %q, want %q", got, want)
	}
	if got, want := ScopeHost("gitlab.example.com/portals/infra"), "gitlab.example.com"; got != want {
		t.Fatalf("ScopeHost = %q, want %q", got, want)
	}
	if !strings.HasPrefix(NamespaceUsername("h/x"), NamespacePrefix) {
		t.Fatal("a manager entry must be namespaced by its username")
	}
}

func TestReadHostGoesThroughGitCredentialFill(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "gitlab.example.com", Username: "oauth2", Password: "s3cret"})

	credential, ok := access.ReadHost(context.Background(), "gitlab.example.com")
	if !ok {
		t.Fatal("ReadHost found nothing")
	}
	if credential.Username != "oauth2" || credential.Secret != "s3cret" {
		t.Fatalf("ReadHost = %+v", credential)
	}
	call := singleCall(t, dir)
	if got := strings.Join(call.Args, " "); !strings.HasSuffix(got, "credential fill") {
		t.Fatalf("argv %q does not end in a credential fill", got)
	}
	if want := "protocol=https\nhost=gitlab.example.com\n\n"; call.Stdin != want {
		t.Fatalf("request payload = %q, want %q", call.Stdin, want)
	}
}

func TestEveryCallDisablesInteractivePrompting(t *testing.T) {
	access, dir := fakeAccess(t, "")
	access.Environ = append(access.Environ,
		"GIT_ASKPASS=/operator/askpass", "SSH_ASKPASS=/operator/ssh-askpass",
		"GIT_TERMINAL_PROMPT=1", "GCM_INTERACTIVE=always")
	seedStore(t, dir, credEntry{Host: "h", Username: NamespaceUsername("h/x"), Password: "tok"})

	if err := access.StoreScoped(context.Background(), "h/x", "h", "tok"); err != nil {
		t.Fatalf("StoreScoped: %v", err)
	}
	calls := recordedCalls(t, dir)
	if len(calls) == 0 {
		t.Fatal("no credential call was recorded")
	}
	for index, call := range calls {
		environment := environmentMap(call.Env)
		if environment["GIT_TERMINAL_PROMPT"] != "0" {
			t.Fatalf("call %d: GIT_TERMINAL_PROMPT = %q", index, environment["GIT_TERMINAL_PROMPT"])
		}
		if environment["GCM_INTERACTIVE"] != "never" {
			t.Fatalf("call %d: GCM_INTERACTIVE = %q", index, environment["GCM_INTERACTIVE"])
		}
		for _, askpass := range []string{"GIT_ASKPASS", "SSH_ASKPASS"} {
			if value, present := environment[askpass]; present {
				t.Fatalf("call %d: %s survived as %q; an askpass program is a prompt", index, askpass, value)
			}
		}
		argv := strings.Join(call.Args, " ")
		for _, disabled := range []string{"-c core.askPass=", "-c credential.interactive=false"} {
			if !strings.Contains(argv, disabled) {
				t.Fatalf("call %d: argv %q does not carry %q", index, argv, disabled)
			}
		}
	}
}

func TestOperatorHomeIsPinnedForTheHelperLookup(t *testing.T) {
	access, dir := fakeAccess(t, "")
	access.Home = filepath.Join(dir, "operator-home")
	access.Environ = append(access.Environ, "HOME=/private/fetch-home", "USERPROFILE=C:\\private\\fetch-home")

	access.ReadScoped(context.Background(), "h/x", "h")

	environment := environmentMap(singleCall(t, dir).Env)
	for _, name := range []string{"HOME", "USERPROFILE"} {
		if environment[name] != access.Home {
			t.Fatalf("%s = %q, want the pinned operator home %q", name, environment[name], access.Home)
		}
	}
}

func TestOperatorHomeDefaultsToThisAccountsHome(t *testing.T) {
	access, dir := fakeAccess(t, "")
	access.Home = ""
	home := OperatorHome()
	if home == "" {
		t.Skip("this platform names no home directory")
	}

	access.ReadHost(context.Background(), "h")

	if got := environmentMap(singleCall(t, dir).Env)["HOME"]; got != home {
		t.Fatalf("HOME = %q, want %q", got, home)
	}
}

func TestReadsReportAbsentMaterialRatherThanFailing(t *testing.T) {
	access, _ := fakeAccess(t, modeFail)
	if _, ok := access.ReadHost(context.Background(), "h"); ok {
		t.Fatal("ReadHost reported material from a Git that cannot answer")
	}
	if _, ok := access.ReadScoped(context.Background(), "h/x", "h"); ok {
		t.Fatal("ReadScoped reported material from a Git that cannot answer")
	}

	absent := Access{Git: filepath.Join(t.TempDir(), "no-such-git"), Home: t.TempDir(), Environ: []string{}}
	if _, ok := absent.ReadHost(context.Background(), "h"); ok {
		t.Fatal("ReadHost reported material without a Git to read it with")
	}
}

func TestReadHostRefusesAManagerNamespacedAnswer(t *testing.T) {
	// A helper asked for a host without a username answers with whatever it
	// holds for that host, and that can be the manager's own entry.
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "h", Username: NamespaceUsername("h/x"), Password: "tok"})

	if credential, ok := access.ReadHost(context.Background(), "h"); ok {
		t.Fatalf("ReadHost reported the manager's own entry as the operator's: %+v", credential)
	}
}

func TestReadScopedRefusesAnswerForAnotherUsername(t *testing.T) {
	access, dir := fakeAccess(t, modeIgnoreUsername)
	seedStore(t, dir, credEntry{Host: "h", Username: "operator", Password: "operator-secret"})

	if secret, ok := access.ReadScoped(context.Background(), "h/x", "h"); ok {
		t.Fatalf("ReadScoped accepted an answer for another username (secret %q)", secret)
	}
}

func TestScopedCredentialRoundTrip(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "h", Username: "operator", Password: "operator-secret"})
	ctx := context.Background()

	if err := access.StoreScoped(ctx, "h/x", "h", "manager-secret"); err != nil {
		t.Fatalf("StoreScoped: %v", err)
	}
	secret, ok := access.ReadScoped(ctx, "h/x", "h")
	if !ok || secret != "manager-secret" {
		t.Fatalf("ReadScoped = %q, %v", secret, ok)
	}
	own, ok := access.ReadHost(ctx, "h")
	if !ok || own.Username != "operator" || own.Secret != "operator-secret" {
		t.Fatalf("the manager entry disturbed the operator's own credential: %+v, %v", own, ok)
	}
	if !access.DeleteScoped(ctx, "h/x", "h") {
		t.Fatal("DeleteScoped reported the credential was not removed")
	}
	if _, ok := access.ReadScoped(ctx, "h/x", "h"); ok {
		t.Fatal("ReadScoped still finds a deleted credential")
	}
	if own, ok := access.ReadHost(ctx, "h"); !ok || own.Secret != "operator-secret" {
		t.Fatalf("deleting the manager entry disturbed the operator's own: %+v, %v", own, ok)
	}
}

func TestStoreRejectsAHelperThatPersistsNothing(t *testing.T) {
	access, _ := fakeAccess(t, modeSilentApprove)

	err := access.StoreScoped(context.Background(), "h/x", "h", "tok")
	if err == nil {
		t.Fatal("StoreScoped trusted a write it never read back")
	}
	assertPlatformGuidance(t, err, "did not persist")
}

func TestStoreReportsAHelperThatRefusesTheWrite(t *testing.T) {
	access, _ := fakeAccess(t, modeFail)

	err := access.StoreScoped(context.Background(), "h/x", "h", "tok")
	if err == nil {
		t.Fatal("StoreScoped accepted a refused write")
	}
	assertPlatformGuidance(t, err, "refused to store")
}

func TestStoreRefusesValuesTheProtocolCannotCarry(t *testing.T) {
	access, dir := fakeAccess(t, "")
	ctx := context.Background()
	for name, value := range map[string]string{
		"newline":        "tok\nprotocol=https",
		"carriage":       "tok\rhost=evil",
		"nul":            "tok\x00",
		"empty":          "",
		"line in a host": "h\nhost=evil",
	} {
		host, secret := "h", value
		if name == "line in a host" {
			host, secret = value, "tok"
		}
		if err := access.StoreScoped(ctx, "h/x", host, secret); err == nil {
			t.Fatalf("StoreScoped accepted a %s", name)
		}
	}
	if calls := recordedCalls(t, dir); len(calls) != 0 {
		t.Fatalf("a refused value still reached Git: %d call(s)", len(calls))
	}
}

func TestDeleteReportsAHelperThatKeepsTheCredential(t *testing.T) {
	access, dir := fakeAccess(t, modeStickyReject)
	seedStore(t, dir, credEntry{Host: "h", Username: NamespaceUsername("h/x"), Password: "tok"})

	if access.DeleteScoped(context.Background(), "h/x", "h") {
		t.Fatal("DeleteScoped reported a removal the helper did not perform")
	}
}

func TestDiscoverIsPresenceOnly(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir,
		credEntry{Host: "h", Username: "operator", Password: "operator-secret"},
		credEntry{Host: "h", Username: NamespaceUsername("h/x"), Password: "manager-secret"})

	material := access.Discover(context.Background(), "h",
		[]string{"h/x", "h/absent", "other.example.com/x"})

	if !material.HostCredential || material.HostUsername != "operator" {
		t.Fatalf("host material = %+v", material)
	}
	if len(material.Scopes) != 1 || material.Scopes[0] != "h/x" {
		t.Fatalf("discovered scopes = %v, want only the stored, host-addressed one", material.Scopes)
	}
	if material.Empty() {
		t.Fatal("material with a host credential reports empty")
	}
	if rendered := fmt.Sprintf("%+v", material); strings.Contains(rendered, "secret") {
		t.Fatalf("a presence-only view retained a secret: %s", rendered)
	}
	if (HostMaterial{}).Empty() != true {
		t.Fatal("a host with nothing on it must report empty")
	}
}

func TestACallIsBounded(t *testing.T) {
	access, _ := fakeAccess(t, modeHang)
	access.Timeout = 200 * time.Millisecond

	started := time.Now()
	if _, ok := access.ReadHost(context.Background(), "h"); ok {
		t.Fatal("a helper that never answers reported material")
	}
	if elapsed := time.Since(started); elapsed > 4*time.Second {
		t.Fatalf("a credential call was not bounded: %s", elapsed)
	}
}

// TestACallIsBoundedWhenTheHelperOutlivesGit is the bound that actually
// matters, and the one killing Git alone does not deliver. A credential
// helper is Git's child: it inherits the output Git was given, and a wedged
// one — a locked keychain, a session with no desktop behind it — is still
// holding that when the deadline kills Git. A call that waits on those pipes
// is bounded by the orphan's lifetime, not by its own timeout, which is why
// the drain delay and not the context is what ends this call.
//
// TestACallIsBounded cannot see this: it wedges the stand-in Git itself, and
// killing that closes its own pipes.
func TestACallIsBoundedWhenTheHelperOutlivesGit(t *testing.T) {
	access, _ := fakeAccess(t, modeHangGrandchild)
	access.Timeout = 200 * time.Millisecond

	started := time.Now()
	if _, ok := access.ReadHost(context.Background(), "h"); ok {
		t.Fatal("a helper that never answers reported material")
	}
	// The orphaned helper stays up for sleeperLifetime; anything under that
	// is proof the call stopped waiting on it rather than outliving it.
	if elapsed := time.Since(started); elapsed > sleeperLifetime/2 {
		t.Fatalf("a credential call waited on a helper that outlived Git: %s", elapsed)
	}
}

// --- the real article --------------------------------------------------------

// TestRealGitKeepsTheManagerEntrySeparate runs the whole surface against the
// operator's own Git and its built-in `store` helper. The helper writes to
// `<home>/.git-credentials`, so a store landing there is also proof the
// operator home was the one pinned for the lookup.
//
// It also pins down what a real helper does with two records for one host.
// `store` puts the newest record first, so once the manager has stored its
// own entry, a host-level read answers with it. That answer is refused rather
// than reported as the operator's own credential: the manager's material is
// never its own evidence that the operator configured anything.
func TestRealGitKeepsTheManagerEntrySeparate(t *testing.T) {
	access, home := realGitAccess(t, "[credential]\n\thelper = store\n")
	ctx := context.Background()
	credentials := filepath.Join(home, ".git-credentials")
	// The operator's own credential for the host, recorded before the
	// manager stores anything of its own.
	if err := os.WriteFile(credentials, []byte("https://operator:operator-secret@h\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	own, ok := access.ReadHost(ctx, "h")
	if !ok || own.Username != "operator" || own.Secret != "operator-secret" {
		t.Fatalf("ReadHost = %+v, %v; the operator's own credential must be readable", own, ok)
	}
	if err := access.StoreScoped(ctx, "h/x", "h", "manager-secret"); err != nil {
		t.Fatalf("StoreScoped: %v", err)
	}
	stored, err := os.ReadFile(credentials)
	if err != nil {
		t.Fatalf("the credential did not land in the pinned operator home: %v", err)
	}
	if !strings.Contains(string(stored), "curator-build-https") {
		t.Fatalf("the manager entry is not namespaced in the store: %s", stored)
	}
	if !strings.Contains(string(stored), "operator-secret") {
		t.Fatalf("the manager entry overwrote the operator's own record: %s", stored)
	}
	if secret, ok := access.ReadScoped(ctx, "h/x", "h"); !ok || secret != "manager-secret" {
		t.Fatalf("ReadScoped = %q, %v", secret, ok)
	}
	if own, ok := access.ReadHost(ctx, "h"); ok {
		t.Fatalf("ReadHost answered with the manager's own entry: %+v", own)
	}
	material := access.Discover(ctx, "h", []string{"h/x"})
	if len(material.Scopes) != 1 || material.Scopes[0] != "h/x" {
		t.Fatalf("Discover = %+v", material)
	}
	if !access.DeleteScoped(ctx, "h/x", "h") {
		t.Fatal("DeleteScoped reported the credential was not removed")
	}
	if own, ok := access.ReadHost(ctx, "h"); !ok || own.Secret != "operator-secret" {
		t.Fatalf("deleting the manager entry took the operator's with it: %+v, %v", own, ok)
	}
}

// TestRealGitWithoutAHelperIsCaughtByTheReadBack is the failure the read-back
// exists for, reproduced with the real Git: `approve` succeeds against a
// configuration with no store behind it and nothing is kept.
func TestRealGitWithoutAHelperIsCaughtByTheReadBack(t *testing.T) {
	access, _ := realGitAccess(t, "[user]\n\tname = operator\n")

	err := access.StoreScoped(context.Background(), "h/x", "h", "tok")
	if err == nil {
		t.Fatal("a write with no credential store behind it was reported as stored")
	}
	assertPlatformGuidance(t, err, "did not persist")
}

func realGitAccess(t *testing.T, gitconfig string) (Access, string) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("git is not available: %v", err)
	}
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(gitconfig), 0o600); err != nil {
		t.Fatal(err)
	}
	// The machine's own system configuration must not reach into the test:
	// it can name a real platform keychain, and no test writes to one.
	environment := []string{"GIT_CONFIG_NOSYSTEM=1"}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(key) {
		case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_NOSYSTEM",
			"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS":
			continue
		}
		environment = append(environment, entry)
	}
	return Access{Git: git, Home: home, Environ: environment, Timeout: 30 * time.Second}, home
}

func assertPlatformGuidance(t *testing.T, err error, lead string) {
	t.Helper()
	message := err.Error()
	if !strings.Contains(message, lead) {
		t.Fatalf("error %q does not report %q", message, lead)
	}
	for _, store := range []string{"osxkeychain", "libsecret", "dpapi"} {
		if !strings.Contains(message, store) {
			t.Fatalf("error %q carries no guidance for %s", message, store)
		}
	}
	if !strings.Contains(message, "environment-variable credential source") {
		t.Fatalf("error %q offers no alternative to a store", message)
	}
	if strings.Contains(message, "tok") && !strings.Contains(message, "token") {
		t.Fatalf("error %q leaked the credential", message)
	}
}

// --- the stand-in Git --------------------------------------------------------

// sleeperLifetime is how long the stand-in helper outlives the Git that
// started it. It only has to be long enough that a call which waited on the
// orphan would be plainly slower than one that did not.
const sleeperLifetime = 30 * time.Second

type credEntry struct{ Host, Username, Password string }

type fakeCall struct {
	Args  []string
	Env   []string
	Stdin string
}

func fakeAccess(t *testing.T, mode string) (Access, string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("locating the test binary: %v", err)
	}
	dir := t.TempDir()
	environ := []string{fakeGitDirEnv + "=" + dir, fakeGitModeEnv + "=" + mode}
	if runtime.GOOS == "windows" {
		// A Windows process needs its system root to start at all.
		environ = append(environ, "SYSTEMROOT="+os.Getenv("SYSTEMROOT"))
	}
	return Access{
		Git: self, Home: filepath.Join(dir, "home"), Environ: environ, Timeout: 30 * time.Second,
	}, dir
}

func seedStore(t *testing.T, dir string, entries ...credEntry) {
	t.Helper()
	payload, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "store.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func recordedCalls(t *testing.T, dir string) []fakeCall {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(dir, "call-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)
	calls := make([]fakeCall, 0, len(names))
	for _, name := range names {
		payload, err := os.ReadFile(name) // #nosec G304 -- test-owned path.
		if err != nil {
			t.Fatal(err)
		}
		var call fakeCall
		if err := json.Unmarshal(payload, &call); err != nil {
			t.Fatal(err)
		}
		calls = append(calls, call)
	}
	return calls
}

func singleCall(t *testing.T, dir string) fakeCall {
	t.Helper()
	calls := recordedCalls(t, dir)
	if len(calls) != 1 {
		t.Fatalf("expected exactly one credential call, got %d", len(calls))
	}
	return calls[0]
}

func environmentMap(entries []string) map[string]string {
	environment := map[string]string{}
	for _, entry := range entries {
		if key, value, found := strings.Cut(entry, "="); found {
			environment[strings.ToUpper(key)] = value
		}
	}
	return environment
}

// fakeGitMain is the stand-in Git: it records the call, then serves the
// credential protocol out of a JSON store, with one deliberate defect
// selected by mode.
func fakeGitMain(dir string) int {
	mode := os.Getenv(fakeGitModeEnv)
	if mode == modeSleeper {
		// Not a credential call and not recorded as one: this is the helper,
		// started by the stand-in Git, holding the output it inherited.
		time.Sleep(sleeperLifetime)
		return 1
	}
	input, _ := io.ReadAll(os.Stdin)
	recordCall(dir, fakeCall{Args: os.Args[1:], Env: os.Environ(), Stdin: string(input)})
	if mode == modeFail {
		return 1
	}
	if mode == modeHang {
		time.Sleep(sleeperLifetime)
		return 1
	}
	if mode == modeHangGrandchild {
		return waitOnASleepingHelper()
	}
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "credential" {
		return 1
	}
	request := map[string]string{}
	for _, line := range strings.Split(string(input), "\n") {
		if key, value, found := strings.Cut(line, "="); found {
			request[key] = value
		}
	}
	entries := fakeStore(dir)
	switch os.Args[len(os.Args)-1] {
	case "fill":
		wanted := request["username"]
		if mode == modeIgnoreUsername {
			wanted = ""
		}
		for _, entry := range entries {
			if entry.Host != request["host"] || (wanted != "" && entry.Username != wanted) {
				continue
			}
			fmt.Printf("protocol=https\nhost=%s\nusername=%s\npassword=%s\n",
				entry.Host, entry.Username, entry.Password)
			return 0
		}
		// What Git does when no helper answers and prompting is disabled.
		return 1
	case "approve":
		if mode == modeSilentApprove {
			return 0
		}
		kept := entries[:0]
		for _, entry := range entries {
			if entry.Host != request["host"] || entry.Username != request["username"] {
				kept = append(kept, entry)
			}
		}
		kept = append(kept, credEntry{request["host"], request["username"], request["password"]})
		writeFakeStore(dir, kept)
		return 0
	case "reject":
		if mode == modeStickyReject {
			return 0
		}
		kept := entries[:0]
		for _, entry := range entries {
			if entry.Host != request["host"] || entry.Username != request["username"] {
				kept = append(kept, entry)
			}
		}
		writeFakeStore(dir, kept)
		return 0
	}
	return 1
}

// waitOnASleepingHelper is the stand-in Git spawning a credential helper and
// blocking on it, the way Git blocks on a helper that never answers. The
// helper is handed this process's own stderr, which is the pipe the manager
// is reading — so killing this process leaves that pipe open in the helper,
// which is the whole point of the case.
func waitOnASleepingHelper() int {
	self, err := os.Executable()
	if err != nil {
		return 1
	}
	helper := exec.Command(self, "credential-helper") // #nosec G204 -- the test binary itself.
	helper.Stderr, helper.Stdout = os.Stderr, os.Stderr
	helper.Env = append(os.Environ(), fakeGitModeEnv+"="+modeSleeper)
	if err := helper.Start(); err != nil {
		return 1
	}
	_ = helper.Wait()
	return 1
}

func recordCall(dir string, call fakeCall) {
	payload, err := json.Marshal(call)
	if err != nil {
		return
	}
	existing, _ := filepath.Glob(filepath.Join(dir, "call-*.json"))
	name := filepath.Join(dir, fmt.Sprintf("call-%04d.json", len(existing)))
	_ = os.WriteFile(name, payload, 0o600)
}

func fakeStore(dir string) []credEntry {
	payload, err := os.ReadFile(filepath.Join(dir, "store.json")) // #nosec G304 -- test-owned path.
	if err != nil {
		return nil
	}
	var entries []credEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		return nil
	}
	return entries
}

func writeFakeStore(dir string, entries []credEntry) {
	payload, err := json.Marshal(entries)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, "store.json"), payload, 0o600)
}
