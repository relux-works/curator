package crossconformance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/testcli"
)

// Transport-revision-2 semantic rows (repository-transport §§4-7): every
// row asserts the loader verdict at the config entry on every OS, and
// the fetch behavior through the compiled CLI on unix (POSIX-only
// wrapper; the declared platform-control reason on Windows). Refusals
// carry zero clone attempts; positives bind the canonical lock
// identity. Both user-configuration rows are driven: the resolved lane
// isolates user git/ssh configuration (proven in-package), and CLI
// resolve isolates user git configuration on the literal-URL lane.

const (
	v2Identity      = "fixture.test/kit"
	v2Primary       = "https://fixture.test/kit.git"
	v2Mirror        = "https://mirror.fixture.test/kit.git"
	v2PortHTTPS     = "https://fixture.test:8443/kit.git"
	v2AliasHost     = "mirror.corp.example"
	v2AliasURL      = "https://mirror.corp.example:8443/kit.git"
	v2AlternateSSH  = "git@fixture.test:kit.git"
	v2AliasEndpoint = "https://fixture.test/kit.git"
)

func v2Endpoint(url, auth, extra string) string {
	endpoint := `{"url":"` + url + `","authentication":"` + auth + `"`
	if extra != "" {
		endpoint += "," + extra
	}
	return endpoint + `}`
}

func v2PolicyDoc(entry, aliases string) string {
	doc := `{"schema_version":2,"repositories":{"` + v2Identity + `":` + entry + `}`
	if aliases != "" {
		doc += `,"aliases":` + aliases
	}
	return doc + `}`
}

func v2Entry(endpoints, fallback, extra string) string {
	entry := `"endpoints":[` + endpoints + `],"fallback":"` + fallback + `"`
	if extra != "" {
		entry += "," + extra
	}
	return `{` + entry + `}`
}

func init() {
	registerDraftSemantic("v2-port-endpoint", driveV2PortEndpoint)
	registerDraftSemantic("v2-declared-mirror", driveV2DeclaredMirror)
	registerDraftSemantic("v2-mirror-first", driveV2MirrorFirst)
	registerDraftSemantic("v2-alias-resolution", driveV2AliasResolution)
	registerDraftSemantic("v2-reader-accepts-v1-policy", driveV2ReaderAcceptsV1)
	registerDraftSemantic("v2-undeclared-mirror", driveV2UndeclaredMirror)
	registerDraftSemantic("v2-pin-port-mismatch", driveV2PinPortMismatch)
	registerDraftSemantic("v2-alias-unknown", driveV2AliasUnknown)
	registerDraftSemantic("v2-alias-mirror-undeclared", driveV2AliasMirrorUndeclared)
	registerDraftSemantic("v2-embedded-alias-host", driveV2EmbeddedAliasHost)
	registerDraftSemantic("v2-alias-auth-mismatch", driveV2AliasAuthMismatch)
	registerDraftSemantic("v2-alias-chain", driveV2AliasChain)
	registerDraftSemantic("v2-double-port", driveV2DoublePort)
	registerDraftSemantic("v2-spurious-mirror-of", driveV2SpuriousMirrorOf)
	registerDraftSemantic("v2-mirror-of-mismatch", driveV2MirrorOfMismatch)
	registerDraftSemantic("v2-v1-reader-rejects-v2-policy", driveV2V1ReaderRejectsV2)
	registerDraftSemantic("v2-user-ssh-alias-ignored", driveV2UserSSHAlias)
	registerDraftSemantic("v2-user-insteadof-ignored", driveV2UserInsteadOf)
	registerDraftSemantic("v2-external-build-mirror-admitted", driveV2ExternalMirrorAdmitted)
	registerDraftSemantic("v2-external-build-port-refused", driveV2ExternalPortRefused)
	registerDraftSemantic("v2-external-build-alias-refused", driveV2ExternalAliasRefused)
}

func v2Parse(t *testing.T, doc string) *config.SourcePolicy {
	t.Helper()
	policy, err := config.ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err != nil {
		t.Fatalf("ParseSourcePolicy: %v", err)
	}
	return policy
}

func v2Resolve(t *testing.T, policy *config.SourcePolicy) config.Resolution {
	t.Helper()
	resolved, err := config.ResolveRepositoryEndpoints(policy, "", v2Identity)
	if err != nil {
		t.Fatalf("ResolveRepositoryEndpoints: %v", err)
	}
	return resolved
}

func v2ResolveRefused(t *testing.T, doc, wantClass string) {
	t.Helper()
	policy, err := config.ParseSourcePolicy([]byte(doc), "source-policy.json")
	if err != nil {
		if !strings.Contains(err.Error(), wantClass) {
			t.Fatalf("parse err = %v, want %s", err, wantClass)
		}
		return
	}
	_, err = config.ResolveRepositoryEndpoints(policy, "", v2Identity)
	if err == nil || !strings.Contains(err.Error(), wantClass) {
		t.Fatalf("resolve err = %v, want %s", err, wantClass)
	}
}

// v2CLIResolve runs one CLI resolve with the given policy doc and
// fixture arms, returning the exit code, stderr, and clone log.
func v2CLIResolve(t *testing.T, policyDoc string, failURL, failStderr, rewriteURL string) (int, string, []string, string) {
	t.Helper()
	realGit := requireGit(t)
	root := t.TempDir()
	bare, commit := draftCLIKitBare(t, root)
	fakeDir, logPath := installDraftTransportShim(t, failURL, failStderr, rewriteURL, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	if policyDoc != "" {
		if err := os.WriteFile(filepath.Join(filepath.Dir(configPath), "source-policy.json"), []byte(policyDoc), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"` + v2Identity + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app")
	return code, stderr, draftCloneLog(t, logPath), commit + "|" + project
}

func v2LockIdentity(t *testing.T, project, wantCommit string) {
	t.Helper()
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	if member.Package.Repository != v2Identity {
		t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
	}
	if member.Package.Commit.Hex != wantCommit {
		t.Fatalf("lock commit = %s, want %s", member.Package.Commit.Hex, wantCommit)
	}
	if strings.Contains(member.Package.Repository, ":") {
		t.Fatalf("lock identity carries a port: %q", member.Package.Repository)
	}
}

func driveV2PortEndpoint(t *testing.T, _ draftSemanticCase) {
	doc := v2PolicyDoc(v2Entry(v2Endpoint(v2PortHTTPS, "team-https", ""), "none", ""), "")
	resolved := v2Resolve(t, v2Parse(t, doc))
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != v2PortHTTPS || !resolved.Attempts[0].HasExplicitPort || resolved.Attempts[0].ResolvedPort != 8443 {
		t.Fatalf("resolved = %+v, want the listed port endpoint once", resolved.Attempts)
	}
	if resolved.Identity != v2Identity {
		t.Fatalf("identity = %q, want the port-stripped key", resolved.Identity)
	}
	code, stderr, clones, rest := v2CLIResolve(t, doc, "https://never.invalid/x.git", "fatal: unexpected", v2PortHTTPS)
	parts := strings.SplitN(rest, "|", 2)
	if code != 0 {
		t.Fatalf("resolve = %d, want success:\n%s", code, stderr)
	}
	if len(clones) != 1 || clones[0] != v2PortHTTPS {
		t.Fatalf("clones = %v, want the listed endpoint once", clones)
	}
	v2LockIdentity(t, parts[1], parts[0])
}

func driveV2DeclaredMirror(t *testing.T, _ draftSemanticCase) {
	doc := v2PolicyDoc(v2Entry(v2Endpoint(v2Mirror, "mirror-https", `"mirror_of":"`+v2Identity+`"`), "none", ""), "")
	resolved := v2Resolve(t, v2Parse(t, doc))
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].URL != v2Mirror || resolved.Attempts[0].MirrorOf != v2Identity {
		t.Fatalf("resolved = %+v, want the declared mirror once", resolved.Attempts)
	}
	code, stderr, clones, rest := v2CLIResolve(t, doc, "https://never.invalid/x.git", "fatal: unexpected", v2Mirror)
	parts := strings.SplitN(rest, "|", 2)
	if code != 0 {
		t.Fatalf("resolve = %d, want success:\n%s", code, stderr)
	}
	if len(clones) != 1 || clones[0] != v2Mirror {
		t.Fatalf("clones = %v, want the listed mirror once", clones)
	}
	v2LockIdentity(t, parts[1], parts[0])
}

func driveV2MirrorFirst(t *testing.T, _ draftSemanticCase) {
	doc := v2PolicyDoc(v2Entry(
		v2Endpoint(v2Mirror, "mirror-https", `"mirror_of":"`+v2Identity+`"`)+`,`+
			v2Endpoint(v2Primary, "team-https", ""), "availability-auth", ""), "")
	resolved := v2Resolve(t, v2Parse(t, doc))
	if len(resolved.Attempts) != 2 || resolved.Attempts[0].MirrorOf != v2Identity {
		t.Fatalf("resolved = %+v, want mirror first", resolved.Attempts)
	}
	dns := "fatal: unable to access '" + v2Mirror + "': Could not resolve host: mirror.fixture.test"
	code, stderr, clones, rest := v2CLIResolve(t, doc, v2Mirror, dns, v2Primary)
	parts := strings.SplitN(rest, "|", 2)
	if code != 0 {
		t.Fatalf("resolve = %d, want the fallback success:\n%s", code, stderr)
	}
	if len(clones) != 2 || clones[0] != v2Mirror || clones[1] != v2Primary {
		t.Fatalf("clones = %v, want mirror then primary", clones)
	}
	v2LockIdentity(t, parts[1], parts[0])
}

// driveV2AliasResolution drives the §5 substitution end to end: the
// loader resolves the alias, CLI resolve connects to the substituted
// address exactly once, the lock keeps the canonical identity, and the
// substituted address never appears in user-facing diagnostics.
func driveV2AliasResolution(t *testing.T, _ draftSemanticCase) {
	aliases := `{"corp-mirror":{"host":"` + v2AliasHost + `","port":8443,"authentication":"team-https"}}`
	doc := v2PolicyDoc(v2Entry(v2Endpoint(v2AliasEndpoint, "team-https", `"alias":"corp-mirror","mirror_of":"`+v2Identity+`"`), "none", ""), aliases)
	resolved := v2Resolve(t, v2Parse(t, doc))
	if len(resolved.Attempts) != 1 {
		t.Fatalf("resolved = %+v, want one attempt", resolved.Attempts)
	}
	attempt := resolved.Attempts[0]
	if attempt.Alias != "corp-mirror" || attempt.ResolvedHost != v2AliasHost || attempt.ResolvedPort != 8443 {
		t.Fatalf("attempt = %+v, want the substituted address", attempt)
	}
	if resolved.Identity != v2Identity {
		t.Fatalf("identity = %q, want the canonical key", resolved.Identity)
	}
	if target, ok := attempt.ConnectionURL(); !ok || target != v2AliasURL {
		t.Fatalf("ConnectionURL = %q, %v, want %q", target, ok, v2AliasURL)
	}
	code, stderr, clones, rest := v2CLIResolve(t, doc, "https://never.invalid/x.git", "fatal: unexpected", v2AliasURL)
	parts := strings.SplitN(rest, "|", 2)
	if code != 0 {
		t.Fatalf("resolve = %d, want success:\n%s", code, stderr)
	}
	if len(clones) != 1 || clones[0] != v2AliasURL {
		t.Fatalf("clones = %v, want the substituted address once", clones)
	}
	if strings.Contains(stderr, v2AliasHost) || strings.Contains(stderr, "corp-mirror") {
		t.Fatalf("stderr leaks endpoint provenance:\n%s", stderr)
	}
	v2LockIdentity(t, parts[1], parts[0])
}

func driveV2ReaderAcceptsV1(t *testing.T, _ draftSemanticCase) {
	doc := `{"schema_version":1,"repositories":{"` + v2Identity + `":{"endpoints":[` + v2Endpoint(v2Primary, "team-https", "") + `],"fallback":"none"}}}`
	policy := v2Parse(t, doc)
	if policy.SchemaVersion != config.SourcePolicySchemaVersion || policy.Aliases != nil {
		t.Fatalf("policy = %+v, want a bare schema-1 load", policy)
	}
	resolved := v2Resolve(t, policy)
	if len(resolved.Attempts) != 1 || resolved.Attempts[0].MirrorOf != "" || resolved.Attempts[0].Alias != "" || resolved.Attempts[0].HasExplicitPort {
		t.Fatalf("resolved = %+v, want a bare revision-1 attempt", resolved.Attempts)
	}
	code, stderr, clones, rest := v2CLIResolve(t, doc, "https://never.invalid/x.git", "fatal: unexpected", v2Primary)
	parts := strings.SplitN(rest, "|", 2)
	if code != 0 {
		t.Fatalf("resolve = %d, want success:\n%s", code, stderr)
	}
	if len(clones) != 1 || clones[0] != v2Primary {
		t.Fatalf("clones = %v, want the v1 endpoint once", clones)
	}
	v2LockIdentity(t, parts[1], parts[0])
}

func driveV2Refusal(t *testing.T, doc, wantClass string) {
	v2ResolveRefused(t, doc, wantClass)
	code, stderr, clones, rest := v2CLIResolve(t, doc, "https://never.invalid/x.git", "fatal: unexpected", "")
	if code == 0 {
		t.Fatalf("resolve succeeded, want %s:\n%s", wantClass, stderr)
	}
	if !strings.Contains(stderr, wantClass) {
		t.Fatalf("stderr misses %s:\n%s", wantClass, stderr)
	}
	if len(clones) != 0 {
		t.Fatalf("clones = %v, want zero attempts", clones)
	}
	parts := strings.SplitN(rest, "|", 2)
	if _, err := os.Stat(filepath.Join(parts[1], "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

func driveV2UndeclaredMirror(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Mirror, "mirror-https", ""), "none", ""), ""), config.CodeRepositoryMirrorUndeclared)
}

func driveV2PinPortMismatch(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", ""), "none", `"pin":"`+v2PortHTTPS+`"`), ""), config.CodeRepositoryPolicyInvalid)
}

func driveV2AliasUnknown(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", `"alias":"absent-alias"`), "none", ""), `{}`), config.CodeRepositoryAliasUnknown)
}

func driveV2AliasMirrorUndeclared(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", `"alias":"corp-mirror"`), "none", ""),
		`{"corp-mirror":{"host":"`+v2AliasHost+`","authentication":"team-https"}}`), config.CodeRepositoryMirrorUndeclared)
}

func driveV2EmbeddedAliasHost(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint("ssh://git@corp-mirror/kit.git", "team-ssh", ""), "none", ""),
		`{"corp-mirror":{"host":"`+v2AliasHost+`","authentication":"team-ssh"}}`), config.CodeRepositoryPolicyInvalid)
}

func driveV2AliasAuthMismatch(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", `"alias":"corp-mirror"`), "none", ""),
		`{"corp-mirror":{"host":"fixture.test","authentication":"other-provider"}}`), config.CodeRepositoryPolicyInvalid)
}

func driveV2AliasChain(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", `"alias":"first","mirror_of":"`+v2Identity+`"`), "none", ""),
		`{"first":{"host":"second","authentication":"team-https"},"second":{"host":"fixture.test","authentication":"team-https"}}`), config.CodeRepositoryPolicyInvalid)
}

func driveV2DoublePort(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2PortHTTPS, "team-https", `"alias":"corp-mirror"`), "none", ""),
		`{"corp-mirror":{"host":"fixture.test","port":9443,"authentication":"team-https"}}`), config.CodeRepositoryPolicyInvalid)
}

func driveV2SpuriousMirrorOf(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Primary, "team-https", `"mirror_of":"`+v2Identity+`"`), "none", ""), ""), config.CodeRepositoryPolicyInvalid)
}

func driveV2MirrorOfMismatch(t *testing.T, _ draftSemanticCase) {
	driveV2Refusal(t, v2PolicyDoc(v2Entry(v2Endpoint(v2Mirror, "mirror-https", `"mirror_of":"other.test/kit"`), "none", ""), ""), config.CodeRepositoryPolicyInvalid)
}

func driveV2V1ReaderRejectsV2(t *testing.T, _ draftSemanticCase) {
	for _, doc := range []string{
		`{"schema_version":1,"repositories":{"` + v2Identity + `":{"endpoints":[` + v2Endpoint(v2Primary, "team-https", "") + `],"fallback":"none"}},"aliases":{"a":{"host":"example.org","authentication":"team-https"}}}`,
		`{"schema_version":1,"repositories":{"` + v2Identity + `":{"endpoints":[` + v2Endpoint(v2Primary, "team-https", `"mirror_of":"`+v2Identity+`"`) + `],"fallback":"none"}}}`,
		`{"schema_version":1,"repositories":{"` + v2Identity + `":{"endpoints":[` + v2Endpoint(v2PortHTTPS, "team-https", "") + `],"fallback":"none"}}}`,
		`{"schema_version":1,"repositories":{"` + v2Identity + `":{"endpoints":[` + v2Endpoint(v2Primary, "team-https", `"alias":"a"`) + `],"fallback":"none"}}}`,
	} {
		v2ResolveRefused(t, doc, config.CodeRepositoryPolicyInvalid)
	}
	// One representative through the CLI: no silent field ignoring,
	// zero attempts.
	driveV2Refusal(t, `{"schema_version":1,"repositories":{"`+v2Identity+`":{"endpoints":[`+v2Endpoint(v2Primary, "team-https", `"mirror_of":"`+v2Identity+`"`)+`],"fallback":"none"}}}`, config.CodeRepositoryPolicyInvalid)
}

// driveV2UserInsteadOf drives the spec case exactly: a literal URL
// declaration with no machine policy entry and a hostile user git
// configuration must attempt the declared URL once, literally, with the
// insteadOf never applied. The transport shim stands in for the network
// (the declared URL is served from the local bare fixture); the hostile
// insteadOf targets the URL real git receives, so any consultation
// redirects the clone to the evil fixture and the lock binds the evil
// commit instead of the declared one. Hostile configuration is planted
// through the GIT_CONFIG selectors and the child HOME file together, so
// the row also proves the isolation is HOME-independent.
func driveV2UserInsteadOf(t *testing.T, _ draftSemanticCase) {
	realGit := requireGit(t)
	root := t.TempDir()
	bare, commit := draftCLIKitBare(t, root)
	evilWork := filepath.Join(root, "evil-work")
	evilBare := filepath.Join(root, "evil.git")
	if err := os.MkdirAll(evilWork, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(evilWork, "skills", "review"), "review")
	runDraftGit(t, evilWork, "init", "-q", "-b", "main")
	runDraftGit(t, evilWork, "add", ".")
	runDraftGit(t, evilWork, "commit", "-qm", "evil")
	runDraftGit(t, evilWork, "tag", "v1")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", evilWork, evilBare)
	evilCommit := draftGitOutput(t, "", "--git-dir", evilBare, "rev-parse", "v1^{commit}")
	if evilCommit == commit {
		t.Fatal("declared and evil fixtures collide")
	}
	fakeDir, logPath := installDraftTransportShim(t, "https://never.invalid/x.git", "fatal: unexpected", v2Primary, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	hostileBody := "[url \"file://" + evilBare + "\"]\n" +
		"\tinsteadOf = file://" + bare + "\n" +
		"\tpushInsteadOf = file://" + bare + "\n" +
		"[core]\n\tsshCommand = /nonexistent-evil-ssh\n"
	hostile := filepath.Join(root, "hostile.gitconfig")
	if err := os.WriteFile(hostile, []byte(hostileBody), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(pathEnv, "GIT_CONFIG_GLOBAL="+hostile, "GIT_CONFIG_SYSTEM="+hostile, "GIT_CONFIG_NOSYSTEM=0")
	configPath, project, home := setupCLIProject(t, root)
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(hostileBody), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + v2Primary + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code != 0 {
		t.Fatalf("resolve = %d, want success:\n%s\n%s", code, stdout, stderr)
	}
	if clones := draftCloneLog(t, logPath); len(clones) != 1 || clones[0] != v2Primary {
		t.Fatalf("clones = %v, want the declared URL once, literally", clones)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	if member.Package.Commit.Hex == evilCommit {
		t.Fatalf("lock binds the evil commit %s: user insteadOf was applied", evilCommit)
	}
	if member.Package.Commit.Hex != commit {
		t.Fatalf("lock binds %s, want the declared commit %s", member.Package.Commit.Hex, commit)
	}
	if member.Package.Repository != v2Identity {
		t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
	}
}

// TestDraftLiteralRefreshIgnoresUserConfig pins the fetch half of the
// draft literal-URL isolation (BUG-260920-3ukdk4) at the production
// entry: after a clean resolve, a hostile user git configuration must
// not redirect the refresh fetch to another repository. It is
// intentionally NOT a corpus row (no registerDraftSemantic call): the
// semantic ratio stays 94 and TestDraftSourcesSemanticCoverage is
// unaffected.
func TestDraftLiteralRefreshIgnoresUserConfig(t *testing.T) {
	realGit := requireGit(t)
	root := t.TempDir()
	bare, commit := draftCLIKitBare(t, root)
	work := filepath.Join(root, "kit-work")
	evilWork := filepath.Join(root, "evil-work")
	evilBare := filepath.Join(root, "evil.git")
	if err := os.MkdirAll(evilWork, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(evilWork, "skills", "review"), "review")
	runDraftGit(t, evilWork, "init", "-q", "-b", "main")
	runDraftGit(t, evilWork, "add", ".")
	runDraftGit(t, evilWork, "commit", "-qm", "evil")
	runDraftGit(t, evilWork, "tag", "v1")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", evilWork, evilBare)
	evilMain := draftGitOutput(t, "", "--git-dir", evilBare, "rev-parse", "main^{commit}")
	if evilMain == commit {
		t.Fatal("declared and evil fixtures collide")
	}
	fakeDir, logPath := installDraftTransportShim(t, "https://never.invalid/x.git", "fatal: unexpected", v2Primary, bare, realGit)
	pathEnv := draftTransportPATH(t, fakeDir)
	configPath, project, home := setupCLIProject(t, root)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + v2Primary + `","branch":"main"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app"); code != 0 {
		t.Fatalf("initial resolve = %d:\n%s\n%s", code, stdout, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	if member.Package.Commit.Hex != commit {
		t.Fatalf("initial lock = %s, want %s", member.Package.Commit.Hex, commit)
	}
	// Advance the declared repository's main past the resolved commit.
	if err := os.WriteFile(filepath.Join(work, "skills", "review", "references", "info.md"), []byte("advanced"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "advance")
	runDraftGit(t, work, "push", "-q", "--", bare, "main:main")
	advanced := draftGitOutput(t, "", "--git-dir", bare, "rev-parse", "main^{commit}")
	if advanced == commit || advanced == evilMain {
		t.Fatal("fixture advance collided")
	}
	hostileBody := "[url \"file://" + evilBare + "\"]\n" +
		"\tinsteadOf = file://" + bare + "\n" +
		"\tpushInsteadOf = file://" + bare + "\n"
	hostile := filepath.Join(root, "hostile.gitconfig")
	if err := os.WriteFile(hostile, []byte(hostileBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(hostileBody), 0o644); err != nil {
		t.Fatal(err)
	}
	hostileEnv := []string{"GIT_CONFIG_GLOBAL=" + hostile, "GIT_CONFIG_SYSTEM=" + hostile, "GIT_CONFIG_NOSYSTEM=0"}
	// Positive control: a raw fetch of a clone of the declared tree
	// under the hostile selectors lands on evil's main, proving the
	// payload redirects fetches when it is consulted.
	control := filepath.Join(t.TempDir(), "control")
	if code, _, stderr := testcli.Run(t, "", nil, "", realGit, "clone", "--quiet", "--", "file://"+bare, control); code != 0 {
		t.Fatalf("control clone: %s", stderr)
	}
	if code, _, stderr := testcli.Run(t, control, hostileEnv, "", realGit, "fetch", "--quiet", "--all"); code != 0 {
		t.Fatalf("control fetch: %s", stderr)
	}
	if got := draftGitOutput(t, control, "rev-parse", "origin/main"); got != evilMain {
		t.Fatalf("control origin/main = %s, want evil %s (payload not live)", got, evilMain)
	}
	env := append(pathEnv, hostileEnv...)
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "refresh", "app")
	if code != 0 {
		t.Fatalf("refresh = %d:\n%s\n%s", code, stdout, stderr)
	}
	if clones := draftCloneLog(t, logPath); len(clones) != 1 {
		t.Fatalf("clones = %v, want exactly the initial clone (refresh must fetch, not reclone)", clones)
	}
	lock, err = sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok = lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	if member.Package.Commit.Hex == evilMain {
		t.Fatalf("refresh lock binds evil main %s: user configuration was applied on the fetch path", evilMain)
	}
	if member.Package.Commit.Hex != advanced {
		t.Fatalf("refresh lock = %s, want advanced declared %s", member.Package.Commit.Hex, advanced)
	}
	if member.Package.Repository != v2Identity {
		t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
	}
}

// driveV2UserSSHAlias drives the spec case exactly: a logical
// declaration with no machine policy entry and a hostile user ssh
// configuration must refuse with repository_endpoint_unavailable
// without consulting user configuration. The hostile GIT_SSH wrapper
// and user ssh config are planted so any consultation is observable;
// the refusal must happen at planning (no machine policy entry),
// before any fetch, with no lock published.
func driveV2UserSSHAlias(t *testing.T, _ draftSemanticCase) {
	realGit := requireGit(t)
	root := t.TempDir()
	wrapperDir := t.TempDir()
	invoked := filepath.Join(wrapperDir, "invoked.log")
	script := "#!/bin/sh\necho \"invoked $@\" >>'" + invoked + "'\necho 'host key verification failed for the hostile wrapper' >&2\nexit 255\n"
	wrapper := filepath.Join(wrapperDir, "evil-ssh")
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	fakeDir := t.TempDir()
	delegate := "#!/bin/sh\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(delegate), 0o700); err != nil {
		t.Fatal(err)
	}
	pathEnv := draftTransportPATH(t, fakeDir)
	env := append(pathEnv, "GIT_SSH="+wrapper, "GIT_SSH_COMMAND="+wrapper)
	configPath, project, home := setupCLIProject(t, root)
	// The child reads user ssh configuration from its own HOME.
	dotSSH := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dotSSH, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dotSSH, "config"), []byte("Host fixture.test\n\tHostName evil.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"fixture.test/kit","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	// No machine policy file: the spec case's policy_entry is null.
	code, stdout, stderr := runCurator(t, home, configPath, env, "project", "resolve", "app")
	if code == 0 {
		t.Fatalf("resolve succeeded, want the endpoint refusal:\n%s\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, config.CodeRepositoryEndpointUnavailable) {
		t.Fatalf("stderr misses %s:\n%s", config.CodeRepositoryEndpointUnavailable, stderr)
	}
	if !strings.Contains(stderr, "no machine policy entry") {
		t.Fatalf("stderr misses the pre-fetch planning refusal:\n%s", stderr)
	}
	if probe, err := os.ReadFile(invoked); err == nil && len(probe) != 0 {
		t.Fatalf("user ssh stack was consulted: %q", probe)
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatal("refused resolve published a lock")
	}
}

// v2FetchArm is one fetch-shim arm: fail with stderr, or serve from
// the bare fixture.
type v2FetchArm struct {
	stderr   string
	fileRepo string
}

// v2FetchTool builds a GitTool whose wrapper logs every invocation,
// fails fetch argv matching a fail arm, and rewrites success arms to
// local bare fixtures. It returns the tool and the argv log path.
func v2FetchTool(t *testing.T, arms map[string]v2FetchArm) (buildrepo.GitTool, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	git, ok := testcli.LookPath("git")
	if !ok {
		t.Fatal("git is not available")
	}
	git, err := filepath.EvalSymlinks(git)
	if err != nil {
		t.Fatal(err)
	}
	execPath := testcli.Output(t, "", nil, git, "--exec-path")
	version := testcli.Output(t, "", nil, git, "--version")
	root := t.TempDir()
	wrapper := filepath.Join(root, "git-wrapper")
	logPath := filepath.Join(root, "argv.log")
	var urls []string
	for url := range arms {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	fmt.Fprintf(&script, "{ printf 'argv:'; for arg in \"$@\"; do printf ' <%%s>' \"$arg\"; done; printf '\\n'; } >> %s\n", shellQuoteV2(logPath))
	script.WriteString("is_fetch=0; for arg in \"$@\"; do if [ \"$arg\" = \"fetch\" ]; then is_fetch=1; fi; done\n")
	script.WriteString("if [ \"$is_fetch\" = 1 ]; then\n:\n")
	for _, url := range urls {
		if arms[url].fileRepo == "" {
			fmt.Fprintf(&script, "for arg in \"$@\"; do if [ \"$arg\" = %s ]; then printf '%%s' %s >&2; exit 128; fi; done\n", shellQuoteV2(url), shellQuoteV2(arms[url].stderr))
		}
	}
	script.WriteString("fi\nargs=\"\"; for arg in \"$@\"; do\n")
	for _, url := range urls {
		if arms[url].fileRepo != "" {
			fmt.Fprintf(&script, "if [ \"$arg\" = %s ]; then arg=%s; fi\n", shellQuoteV2(url), shellQuoteV2("file://"+arms[url].fileRepo))
		}
	}
	script.WriteString("case \"$arg\" in protocol.https.allow=always|protocol.ssh.allow=always) arg=protocol.file.allow=always ;; esac\n")
	script.WriteString("args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\n")
	fmt.Fprintf(&script, "exec %s \"$@\"\n", shellQuoteV2(git))
	if err := os.WriteFile(wrapper, []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tool := buildrepo.GitTool{Executable: wrapper, ExecPath: execPath, AllowedVersions: []string{version}, AskPass: self, SSHWrapper: self}
	return tool, logPath
}

func shellQuoteV2(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// assertV2MirrorFetchesOncePerResolution groups fetch argv lines by
// their private repository and proves every resolution attempted the
// listed mirror exactly once, with no other fetch target.
func assertV2MirrorFetchesOncePerResolution(t *testing.T, fetches []string, mirror string) {
	t.Helper()
	if len(fetches) == 0 {
		t.Fatal("no fetches observed, want the listed mirror once per resolution")
	}
	byRepo := map[string]int{}
	for _, fetch := range fetches {
		if !strings.Contains(fetch, "<"+mirror+">") {
			t.Fatalf("fetch = %q, want the listed mirror %q", fetch, mirror)
		}
		repo := ""
		for _, field := range strings.Split(fetch, " ") {
			if rest, ok := strings.CutPrefix(field, "<--git-dir="); ok {
				repo = strings.TrimSuffix(rest, ">")
			}
		}
		if repo == "" {
			t.Fatalf("fetch carries no private repository: %q", fetch)
		}
		byRepo[repo]++
	}
	for repo, count := range byRepo {
		if count != 1 {
			t.Fatalf("repository %s attempted %d fetches, want exactly one", repo, count)
		}
	}
}

func v2FetchURLs(t *testing.T, logPath string) []string {
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
		if strings.Contains(line, "<fetch>") {
			out = append(out, line)
		}
	}
	return out
}

// v2ToolsBare builds a real tools repository carrying the proved
// external snapshot files, returning its bare path and commit.
func v2ToolsBare(t *testing.T) (bare, commit string) {
	t.Helper()
	requireGit(t)
	work := t.TempDir()
	files := map[string]string{
		"skill-build.json":       `{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`,
		"tools/cmd/tool/main.go": "package main\n\nfunc main() {}\n",
		"tools/go.mod":           "module example.test/tool\n",
	}
	for rel, content := range files {
		full := filepath.Join(work, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runDraftGit(t, work, "init", "-q", "-b", "main")
	runDraftGit(t, work, "add", ".")
	runDraftGit(t, work, "commit", "-qm", "tools")
	bare = filepath.Join(t.TempDir(), "tools.git")
	runDraftGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	return bare, draftGitOutput(t, "", "--git-dir", bare, "rev-parse", "HEAD")
}

func writeV2ExternalSkill(t *testing.T, dir, gitURL, commit string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: review\ndescription: Test\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := `{"schema_version":7,"build_roots":[],"capabilities":{},"commands":{"etool":{"type":"build","driver":"go-repository-v1","repository":"tools","target":"tool"}},"build_repositories":{"tools":{"git":"` + gitURL + `","locked_commit":{"object_format":"sha1","hex":"` + commit + `"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
}

func driveV2ExternalMirrorAdmitted(t *testing.T, _ draftSemanticCase) {
	const toolsIdentity = "git.example.com/skills/tools"
	const declaredGit = "https://git.example.com/skills/tools.git"
	const mirrorGit = "https://mirror.example.net/skills/tools.git"
	bare, commit := v2ToolsBare(t)
	tool, logPath := v2FetchTool(t, map[string]v2FetchArm{mirrorGit: {fileRepo: bare}})
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project := t.TempDir()
	home := t.TempDir()
	writeV2ExternalSkill(t, filepath.Join(project, "skills", "review"), declaredGit, commit)
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolveDraftPlan(t, project, home, payload)
	policyPath := filepath.Join(home, "source-policy.json")
	policyDoc := `{"schema_version":2,"repositories":{"` + toolsIdentity + `":{"endpoints":[{"url":"` + mirrorGit + `","authentication":"mirror-https","mirror_of":"` + toolsIdentity + `"}],"fallback":"none"}}}`
	if err := os.WriteFile(policyPath, []byte(policyDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	providersPath := filepath.Join(home, "source-providers.json")
	if err := os.WriteFile(providersPath, []byte(`{"schema_version":1,"providers":{"mirror-https":{"https":{"anonymous":true}}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	deps, _ := xbBuildDeps(t)
	external := install.ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policyPath, DraftProvidersPath: providersPath,
		Audit: func(_ context.Context, _ buildrepo.AuditSubject) error { return nil }}
	result := install.Project(draftInstallConfig(home), project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform(), Build: deps, External: external})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	fetches := v2FetchURLs(t, logPath)
	// Install resolves twice (plan phase plus stage phase, each in its
	// own private repository); each resolution must attempt the listed
	// mirror exactly once.
	assertV2MirrorFetchesOncePerResolution(t, fetches, mirrorGit)
	arm := readExternalArm(t, project)
	declared, _ := arm["declared_locked_commit"].(map[string]any)
	if declared["hex"] != commit || arm["commit"] != commit {
		t.Fatalf("arm commits = %v/%v, want the locked %s", declared["hex"], arm["commit"], commit)
	}
	if arm["repository"] != "tools" || arm["receipt_schema_version"] != float64(3) {
		t.Fatalf("arm = %+v, want the receipt-3 external record", arm)
	}
}

// driveV2ExternalRefused proves one strict-lane §7 refusal in two layers.
// Layer 1 runs on every OS through two pure production entries: the loader
// plans the spec case's port/alias endpoint (provenance asserted, never
// assumed), and the executor gate refuses the converted plan with
// build_repository_identity_invalid before any network I/O. Layer 2 runs
// on unix through install.Project: the run refuses with zero fetch
// attempts, preserves all state, and reports the true class end to end
// instead of masking it as build_repository_source_unavailable.
func driveV2ExternalRefused(t *testing.T, skillGit, policyDoc, providersDoc string) {
	policy := v2Parse(t, policyDoc)
	resolved, err := config.ResolveRepositoryEndpoints(policy, skillGit, "")
	if err != nil {
		t.Fatalf("loader refused the planned endpoint: %v", err)
	}
	if len(resolved.Attempts) != 1 {
		t.Fatalf("loader planned %d attempts, want exactly the refused endpoint", len(resolved.Attempts))
	}
	planned := resolved.Attempts[0]
	if !planned.HasExplicitPort && planned.Alias == "" {
		t.Fatalf("loader planned a bare attempt %+v, want port/alias provenance", planned)
	}
	if resolved.Fallback != config.FallbackNone {
		t.Fatalf("loader fallback = %q, want none", resolved.Fallback)
	}
	plan := buildrepo.TransportPlan{Identity: resolved.Identity, Fallback: buildrepo.TransportFallbackNone, Attempts: []buildrepo.TransportAttempt{{
		URL: planned.URL, Authentication: planned.Authentication, MirrorOf: planned.MirrorOf,
		Alias: planned.Alias, ResolvedHost: planned.ResolvedHost, ResolvedPort: planned.ResolvedPort,
		HasExplicitPort: planned.HasExplicitPort,
	}}}
	if err := buildrepo.ValidateTransportPlan(plan); err == nil || !strings.Contains(err.Error(), buildrepo.CodeIdentityInvalid) {
		t.Fatalf("ValidateTransportPlan err = %v, want %s", err, buildrepo.CodeIdentityInvalid)
	}
	tool, logPath := v2FetchTool(t, map[string]v2FetchArm{})
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project := t.TempDir()
	home := t.TempDir()
	writeV2ExternalSkill(t, filepath.Join(project, "skills", "review"), skillGit, strings.Repeat("0", 40))
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolveDraftPlan(t, project, home, payload)
	policyPath := filepath.Join(home, "source-policy.json")
	if err := os.WriteFile(policyPath, []byte(policyDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	providersPath := filepath.Join(home, "source-providers.json")
	if err := os.WriteFile(providersPath, []byte(providersDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	deps, _ := xbBuildDeps(t)
	external := install.ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policyPath, DraftProvidersPath: providersPath,
		Audit: func(_ context.Context, _ buildrepo.AuditSubject) error { return nil }}
	before := treeDigest(t, project) + treeDigest(t, home)
	result := install.Project(draftInstallConfig(home), project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform(), Build: deps, External: external})
	if result.Status != "failed" {
		t.Fatalf("install = %+v, want refusal", result)
	}
	combined := strings.Join(result.Errors, ";")
	if !strings.Contains(combined, buildrepo.CodeIdentityInvalid) {
		t.Fatalf("install diagnostic = %q, want the true class %s", combined, buildrepo.CodeIdentityInvalid)
	}
	if strings.Contains(combined, buildrepo.CodeSourceUnavailable) {
		t.Fatalf("install diagnostic masks the class as %s: %q", buildrepo.CodeSourceUnavailable, combined)
	}
	if fetches := v2FetchURLs(t, logPath); len(fetches) != 0 {
		t.Fatalf("fetches = %q, want zero attempts", fetches)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}

func driveV2ExternalPortRefused(t *testing.T, _ draftSemanticCase) {
	// The skill declares the canonical grammar URL (an explicit port in
	// the declaration itself is refused earlier, at member validation);
	// the machine policy plans the port endpoint the strict lane refuses.
	const toolsIdentity = "git.example.com/skills/tools"
	const declaredGit = "https://git.example.com/skills/tools.git"
	policyDoc := `{"schema_version":2,"repositories":{"` + toolsIdentity + `":{"endpoints":[{"url":"https://git.example.com:8443/skills/tools.git","authentication":"team-https"}],"fallback":"none"}}}`
	driveV2ExternalRefused(t, declaredGit, policyDoc,
		`{"schema_version":1,"providers":{"team-https":{"https":{"anonymous":true}}}}`)
}

func driveV2ExternalAliasRefused(t *testing.T, _ draftSemanticCase) {
	const toolsIdentity = "git.example.com/skills/tools"
	policyDoc := `{"schema_version":2,"repositories":{"` + toolsIdentity + `":{"endpoints":[{"url":"https://git.example.com/skills/tools.git","authentication":"team-https","alias":"corp-mirror","mirror_of":"` + toolsIdentity + `"}],"fallback":"none"}},"aliases":{"corp-mirror":{"host":"mirror.corp.example","authentication":"team-https"}}}`
	driveV2ExternalRefused(t, "https://git.example.com/skills/tools.git", policyDoc,
		`{"schema_version":1,"providers":{"team-https":{"https":{"anonymous":true}}}}`)
}
