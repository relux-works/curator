package install

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/skillspec"
)

// externalRow is one planned external build repository, reduced to the state
// credential resolution actually reads.
func externalRow(skill, command, kind, identity, transport string) plannedExternal {
	return plannedExternal{
		node:      &closure.Node{Name: skill},
		command:   skillspec.Command{Name: command},
		effective: buildrepo.EffectiveState{IdentityKind: kind, Identity: identity, Transport: transport},
	}
}

func sshRow(skill, command, identity string) plannedExternal {
	return externalRow(skill, command, "network-git", identity, "ssh")
}

func resolvedFor(t *testing.T, credentials map[buildSSHKey]buildrepo.OperatorSSHCredentials, skill, command string) buildrepo.OperatorSSHCredentials {
	t.Helper()
	selected, ok := credentials[buildSSHKey{skill: skill, command: command}]
	if !ok {
		t.Fatalf("%s.%s selected no credentials", skill, command)
	}
	return selected
}

// operatorFile creates a real operator credential file, since every selection
// is proved against the filesystem before it reaches the wrapper policy.
func operatorFile(t *testing.T, path string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte("operator\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func operatorAgentSocket(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("an operator agent socket is a unix-domain rendezvous point")
	}
	// A unix-domain socket address is bounded near 104 bytes, which the
	// per-test temporary directory of macOS does not fit inside.
	root, err := os.MkdirTemp("/tmp", "curator-install")
	if err != nil {
		t.Skipf("this host cannot create a short scratch directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	path := filepath.Join(root, "agent.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Skipf("this host cannot create an agent socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	return path
}

// canonicalPath is how validation reports an operator path back: every
// symbolic link resolved, which on macOS moves /tmp and /var under /private.
func canonicalPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// An explicit run-wide selection is the operator saying "use this, here, now",
// so it has to cover repositories a configured scope would otherwise claim and
// repositories no scope covers at all.
func TestRunWideSelectionCoversEveryRepositoryAheadOfConfiguredScopes(t *testing.T) {
	identity := operatorFile(t, filepath.Join(t.TempDir(), "run-wide.pub"))
	selection := BuildSSHSelection{
		RunWide: BuildSSHFlags{Identity: identity},
		Scopes: map[string]config.BuildSSHCredential{
			"git.example.test/portals": {Scope: "git.example.test/portals", Identity: "/operator/scoped"},
		},
	}
	rows := []plannedExternal{
		sshRow("covered", "build", "git.example.test/portals/app"),
		sshRow("uncovered", "build", "other.example.test/tools/kit"),
	}
	credentials, provenance, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, skill := range []string{"covered", "uncovered"} {
		if selected := resolvedFor(t, credentials, skill, "build"); selected.Identity != identity {
			t.Fatalf("%s identity = %q, want the run-wide %q", skill, selected.Identity, identity)
		}
	}
	for _, line := range provenance {
		if !strings.Contains(line, "operator flags/env") {
			t.Fatalf("provenance = %q, want the run-wide origin", line)
		}
	}
}

// Without a run-wide selection each repository is matched on its own canonical
// identity, so one closure spanning two hosts cannot offer either host the
// other's key.
func TestConfiguredScopesAreSelectedPerRepository(t *testing.T) {
	root := t.TempDir()
	host := operatorFile(t, filepath.Join(root, "host.pub"))
	group := operatorFile(t, filepath.Join(root, "group.pub"))
	selection := BuildSSHSelection{Scopes: map[string]config.BuildSSHCredential{
		"git.example.test":                {Scope: "git.example.test", Identity: host},
		"git.example.test/portals":        {Scope: "git.example.test/portals", Identity: group},
		"git.example.test/portals-evil":   {Scope: "git.example.test/portals-evil", Identity: "/operator/never"},
		"other.example.test/tools":        {Scope: "other.example.test/tools", Identity: host},
		"git.example.test/portals/nested": {Scope: "git.example.test/portals/nested", Identity: group},
	}}
	rows := []plannedExternal{
		sshRow("group", "build", "git.example.test/portals/app"),
		sshRow("host", "build", "git.example.test/infra/base"),
		sshRow("boundary", "build", "git.example.test/portals-evil/app"),
	}
	credentials, provenance, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if selected := resolvedFor(t, credentials, "group", "build"); selected.Identity != group {
		t.Fatalf("longest scope identity = %q, want %q", selected.Identity, group)
	}
	if selected := resolvedFor(t, credentials, "host", "build"); selected.Identity != host {
		t.Fatalf("host scope identity = %q, want %q", selected.Identity, host)
	}
	if selected := resolvedFor(t, credentials, "boundary", "build"); selected.Identity != "/operator/never" {
		t.Fatalf("segment boundary identity = %q, want the portals-evil scope", selected.Identity)
	}
	joined := strings.Join(provenance, "\n")
	if !strings.Contains(joined, `config scope "git.example.test/portals"`) {
		t.Fatalf("provenance does not name the selected scope: %s", joined)
	}
}

// Only a repository actually fetched over SSH consumes a credential. HTTPS
// authenticates through the manager credential broker, and a repository an
// operator substitution replaced with a local path is never fetched at all.
func TestRepositoriesThatNeedNoCredentialSkipSelection(t *testing.T) {
	rows := []plannedExternal{
		externalRow("https", "build", "network-git", "git.example.test/portals/app", "https"),
		externalRow("local", "build", "operator-local-git", "local/portals/app", ""),
	}
	credentials, provenance, err := resolveBuildSSH(BuildSSHSelection{}, rows)
	if err != nil {
		t.Fatalf("a closure needing no SSH credential failed: %v", err)
	}
	if len(credentials) != 0 || len(provenance) != 0 {
		t.Fatalf("credentials = %v, provenance = %v, want neither", credentials, provenance)
	}
}

// A substitution that redirects a declared SSH repository decides the
// transport with it: the declared spelling is not what gets fetched.
func TestSubstitutionMovesARepositoryOffAndOntoTheSSHTransport(t *testing.T) {
	identity := operatorFile(t, filepath.Join(t.TempDir(), "id.pub"))
	selection := BuildSSHSelection{RunWide: BuildSSHFlags{Identity: identity}}
	redirected := externalRow("mirrored", "build", "network-git", "mirror.example.test/portals/app", "https")
	redirected.declared = buildrepo.DeclaredState{Identity: "git.example.test/portals/app", Transport: "ssh"}
	promoted := externalRow("promoted", "build", "network-git", "git.example.test/portals/app", "ssh")
	promoted.declared = buildrepo.DeclaredState{Identity: "git.example.test/portals/app", Transport: "https"}
	credentials, _, err := resolveBuildSSH(selection, []plannedExternal{redirected, promoted})
	if err != nil {
		t.Fatal(err)
	}
	if _, selected := credentials[buildSSHKey{skill: "mirrored", command: "build"}]; selected {
		t.Fatal("a repository redirected onto HTTPS still consumed an SSH credential")
	}
	if selected := resolvedFor(t, credentials, "promoted", "build"); selected.Identity != identity {
		t.Fatalf("promoted identity = %q, want %q", selected.Identity, identity)
	}
}

// Nothing may fall back onto whatever the operator's ambient SSH state happens
// to offer, so an unselected repository stops the run before the first fetch.
func TestUnselectedSSHRepositoriesFailClosedWithTheProtocolCode(t *testing.T) {
	rows := []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
		sshRow("infra", "build-agent", "other.example.test/tools/kit"),
		externalRow("public", "build", "network-git", "git.example.test/open/lib", "https"),
	}
	_, _, err := resolveBuildSSH(BuildSSHSelection{}, rows)
	if err == nil {
		t.Fatal("unselected SSH repositories were admitted")
	}
	message := err.Error()
	for _, required := range []string{
		buildrepo.CodeSSHCredentialMissing,
		`git.example.test/portals/app (command "build-tool" of skill "portals")`,
		`other.example.test/tools/kit (command "build-agent" of skill "infra")`,
		"curator config build-ssh add git.example.test/portals",
		EnvBuildSSHAgent,
		EnvBuildSSHIdentity,
	} {
		if !strings.Contains(message, required) {
			t.Errorf("diagnostic missing %q:\n%s", required, message)
		}
	}
	if strings.Contains(message, "git.example.test/open/lib") {
		t.Errorf("an HTTPS repository was reported as missing credentials:\n%s", message)
	}
}

func TestDefaultBuildSSHScopeIsTheRepositoryNamespace(t *testing.T) {
	for identity, scope := range map[string]string{
		"git.example.test/portals/infra/app": "git.example.test/portals/infra",
		"git.example.test/portals/app":       "git.example.test/portals",
		"git.example.test/app":               "git.example.test/app",
	} {
		if got := defaultBuildSSHScope(identity); got != scope {
			t.Errorf("defaultBuildSSHScope(%q) = %q, want %q", identity, got, scope)
		}
	}
}

func TestCaptureBuildSSHSelectionPrefersFlagsOverTheEnvironment(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	knownHosts := operatorFile(t, filepath.Join(home, ".ssh", "known_hosts"))
	environment := map[string]string{
		"HOME":                "" + home,
		EnvBuildSSHIdentity:   "/operator/env.pub",
		EnvBuildSSHAgent:      "/operator/env.sock",
		EnvBuildSSHKnownHosts: "/operator/env_known_hosts",
		envAgentSocket:        "/operator/live.sock",
	}
	environ := func(name string) string { return environment[name] }

	flagged := CaptureBuildSSHSelection(nil, BuildSSHFlags{Identity: "/operator/flag.pub"}, environ)
	if flagged.RunWide.Identity != "/operator/flag.pub" {
		t.Fatalf("identity = %q, want the flag value", flagged.RunWide.Identity)
	}
	if flagged.RunWide.Agent != "/operator/env.sock" || flagged.RunWide.KnownHosts != "/operator/env_known_hosts" {
		t.Fatalf("unflagged fields = %+v, want the environment values", flagged.RunWide)
	}
	if flagged.AgentSocket != "/operator/live.sock" || flagged.DefaultKnownHosts != knownHosts {
		t.Fatalf("operator state = %q %q", flagged.AgentSocket, flagged.DefaultKnownHosts)
	}

	cfg := &config.Config{BuildSSH: map[string]config.BuildSSHCredential{
		"git.example.test": {Scope: "git.example.test", Agent: true},
	}}
	bound := CaptureBuildSSHSelection(cfg, BuildSSHFlags{}, environ)
	if len(bound.Scopes) != 1 {
		t.Fatalf("scopes = %v, want the configured one", bound.Scopes)
	}
	if bound.RunWide.Identity != "/operator/env.pub" {
		t.Fatalf("identity = %q, want the environment value", bound.RunWide.Identity)
	}
}

// A machine-global config records the operator's spelling, so `~/` has to
// resolve at selection time rather than reaching SSH verbatim.
func TestConfiguredPathsExpandTheOperatorHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	identity := operatorFile(t, filepath.Join(home, "id_ed25519.pub"))
	knownHosts := operatorFile(t, filepath.Join(home, "known_hosts"))
	selection := BuildSSHSelection{Scopes: map[string]config.BuildSSHCredential{
		"git.example.test": {
			Scope: "git.example.test", Identity: "~/id_ed25519.pub", KnownHosts: "~/known_hosts",
		},
	}}
	credentials, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build", "git.example.test/portals/app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	selected := resolvedFor(t, credentials, "portals", "build")
	if selected.Identity != identity || selected.KnownHosts != knownHosts {
		t.Fatalf("expanded selection = %+v, want %q and %q", selected, identity, knownHosts)
	}
}

// A scope selects authentication material, not who the destination is allowed
// to be, so an explicitly named host key file still governs a scoped selection.
func TestKnownHostsFallBackFromScopeToRunWideToOperator(t *testing.T) {
	scoped := config.BuildSSHCredential{
		Scope: "git.example.test", Identity: "/operator/id.pub", KnownHosts: "/operator/scoped_known_hosts",
	}
	unscoped := config.BuildSSHCredential{Scope: "git.example.test", Identity: "/operator/id.pub"}
	for name, expectation := range map[string]struct {
		scope config.BuildSSHCredential
		want  string
	}{
		"scope wins":    {scoped, "/operator/scoped_known_hosts"},
		"run-wide next": {unscoped, "/operator/run_wide_known_hosts"},
	} {
		t.Run(name, func(t *testing.T) {
			selection := BuildSSHSelection{
				RunWide:           BuildSSHFlags{KnownHosts: "/operator/run_wide_known_hosts"},
				DefaultKnownHosts: "/operator/home_known_hosts",
				Scopes:            map[string]config.BuildSSHCredential{expectation.scope.Scope: expectation.scope},
			}
			credentials, _, err := resolveBuildSSH(selection, []plannedExternal{
				sshRow("portals", "build", "git.example.test/portals/app"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if selected := resolvedFor(t, credentials, "portals", "build"); selected.KnownHosts != expectation.want {
				t.Fatalf("known hosts = %q, want %q", selected.KnownHosts, expectation.want)
			}
		})
	}
	// Named nowhere, the operator's own host keys are what remains.
	selection := BuildSSHSelection{
		DefaultKnownHosts: "/operator/home_known_hosts",
		Scopes:            map[string]config.BuildSSHCredential{unscoped.Scope: unscoped},
	}
	credentials, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build", "git.example.test/portals/app"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if selected := resolvedFor(t, credentials, "portals", "build"); selected.KnownHosts != "/operator/home_known_hosts" {
		t.Fatalf("known hosts = %q, want the operator default", selected.KnownHosts)
	}
}

// "agent": true records the operator's choice of an agent, not a socket path:
// a macOS agent socket is per login session and a persisted path goes stale.
func TestConfiguredAgentResolvesToTheLiveSocket(t *testing.T) {
	selection := BuildSSHSelection{
		AgentSocket: "/operator/live.sock",
		Scopes: map[string]config.BuildSSHCredential{
			"git.example.test": {Scope: "git.example.test", Agent: true},
		},
	}
	rows := []plannedExternal{sshRow("portals", "build", "git.example.test/portals/app")}
	credentials, _, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if selected := resolvedFor(t, credentials, "portals", "build"); selected.AgentSocket != "/operator/live.sock" {
		t.Fatalf("agent socket = %q, want the live one", selected.AgentSocket)
	}

	selection.AgentSocket = ""
	if _, _, err := resolveBuildSSH(selection, rows); err == nil ||
		!strings.Contains(err.Error(), envAgentSocket) {
		t.Fatalf("agentless resolution error = %v, want one naming %s", err, envAgentSocket)
	}
}

func TestRunWideAgentAutoRequiresALiveSocket(t *testing.T) {
	selection := BuildSSHSelection{RunWide: BuildSSHFlags{Agent: BuildSSHAgentAuto}}
	rows := []plannedExternal{sshRow("portals", "build", "git.example.test/portals/app")}
	_, _, err := resolveBuildSSH(selection, rows)
	if err == nil || !strings.Contains(err.Error(), buildrepo.CodeSSHCredentialMissing) {
		t.Fatalf("error = %v, want %s", err, buildrepo.CodeSSHCredentialMissing)
	}
	// A closure with nothing to fetch over SSH is unaffected by an agent the
	// operator asked for and the machine does not offer.
	if _, _, err := resolveBuildSSH(selection, []plannedExternal{
		externalRow("https", "build", "network-git", "git.example.test/open/lib", "https"),
	}); err != nil {
		t.Fatalf("an SSH-free closure failed on an absent agent: %v", err)
	}

	selection.AgentSocket = "/operator/live.sock"
	credentials, _, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if selected := resolvedFor(t, credentials, "portals", "build"); selected.AgentSocket != "/operator/live.sock" {
		t.Fatalf("agent socket = %q, want the live one", selected.AgentSocket)
	}
}

// End to end for the shapes the operator can actually write down: a configured
// scope has to reach the wrapper policy as the exact OpenSSH tail it means.
func TestEveryConfiguredShapeReachesTheWrapperPolicy(t *testing.T) {
	root := t.TempDir()
	identity := operatorFile(t, filepath.Join(root, "id_ed25519.pub"))
	knownHosts := operatorFile(t, filepath.Join(root, "known_hosts"))
	agentSocket := operatorAgentSocket(t)
	source, err := buildrepo.ParseSource("git@git.example.test:portals/app.git")
	if err != nil {
		t.Fatal(err)
	}
	base := buildrepo.SSHPolicy{
		Wrapper:         filepath.Join(root, "ssh-wrapper"),
		SSH:             filepath.Join(root, "ssh"),
		EmptyConfig:     filepath.Join(root, "ssh.config"),
		EmptyKnownHosts: filepath.Join(root, "empty_known_hosts"),
		ConnectTimeout:  15,
	}
	// The policy reports every operator path canonicalized, so the expected
	// tail is spelled the same way.
	admittedIdentity := canonicalPath(t, identity)
	admittedAgent := canonicalPath(t, agentSocket)
	admittedKnownHosts := canonicalPath(t, knownHosts)
	for _, testCase := range []struct {
		name   string
		scope  config.BuildSSHCredential
		wanted []string
	}{
		{
			name:   "identity only",
			scope:  config.BuildSSHCredential{Scope: "git.example.test", Identity: identity},
			wanted: []string{"IdentitiesOnly=yes", "IdentityAgent=none", "-i " + admittedIdentity},
		},
		{
			name:   "agent only",
			scope:  config.BuildSSHCredential{Scope: "git.example.test", Agent: true},
			wanted: []string{"IdentitiesOnly=no", "IdentityFile=none", "IdentityAgent=" + admittedAgent},
		},
		{
			name:   "agent pinned to one identity",
			scope:  config.BuildSSHCredential{Scope: "git.example.test", Agent: true, Identity: identity},
			wanted: []string{"IdentitiesOnly=yes", "IdentityAgent=" + admittedAgent, "-i " + admittedIdentity},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := config.ValidateBuildSSH(testCase.scope); err != nil {
				t.Fatalf("the scope this case selects is not one an operator may write: %v", err)
			}
			selection := BuildSSHSelection{
				AgentSocket:       agentSocket,
				DefaultKnownHosts: knownHosts,
				Scopes:            map[string]config.BuildSSHCredential{testCase.scope.Scope: testCase.scope},
			}
			credentials, _, err := resolveBuildSSH(selection, []plannedExternal{
				sshRow("portals", "build", "git.example.test/portals/app"),
			})
			if err != nil {
				t.Fatal(err)
			}
			policy, err := buildrepo.SSHPolicyFor(base, source, resolvedFor(t, credentials, "portals", "build"))
			if err != nil {
				t.Fatal(err)
			}
			command, err := buildrepo.ExactSSHCommand(policy,
				[]string{policy.Wrapper, policy.ExpectedHost, "git-upload-pack 'portals/app.git'"})
			if err != nil {
				t.Fatal(err)
			}
			joined := strings.Join(command, " ")
			for _, required := range append(testCase.wanted, "UserKnownHostsFile="+admittedKnownHosts) {
				if !strings.Contains(joined, required) {
					t.Errorf("SSH command missing %q in %s", required, joined)
				}
			}
		})
	}
}
