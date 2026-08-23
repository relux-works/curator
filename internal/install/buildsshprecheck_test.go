package install

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/skillspec"
)

// discoveredSelection is a selection whose discovery is fully determined by
// the test: a home holding the named keys and an agent with a known count.
func discoveredSelection(t *testing.T, keys ...string) BuildSSHSelection {
	t.Helper()
	return BuildSSHSelection{
		Home:        sshHome(t, keys...),
		AgentSocket: "/run/agent.sock",
		AgentKeys:   countedAgent(2, true),
	}
}

func TestUnselectedRepositoriesFailClosedWithCommandsBuiltFromTheCandidates(t *testing.T) {
	rows := []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
		sshRow("portals", "build-agent", "git.example.test/portals/agent"),
		sshRow("infra", "build-kit", "other.example.test/tools/kit"),
		externalRow("public", "build", "network-git", "git.example.test/open/lib", "https"),
	}
	selection := discoveredSelection(t, "id_ed25519.pub", "work.pub")
	_, _, err := resolveBuildSSH(selection, rows)
	if err == nil {
		t.Fatal("unselected SSH repositories were admitted")
	}
	message := err.Error()
	for _, required := range []string{
		buildrepo.CodeSSHCredentialMissing,
		`git.example.test/portals/app (command "build-tool" of skill "portals")`,
		`other.example.test/tools/kit (command "build-kit" of skill "infra")`,
		// The material discovery found, named so the operator can tell which
		// of their own keys the commands below refer to.
		"detected on this host:",
		"SSH agent at /run/agent.sock holding 2 keys",
		"~/.ssh/id_ed25519.pub",
		"~/.ssh/work.pub",
		// Ready-to-run commands, one set per uncovered namespace, built from
		// exactly those candidates.
		"curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/id_ed25519.pub",
		"curator config build-ssh add git.example.test/portals --agent",
		"curator config build-ssh add git.example.test/portals --identity ~/.ssh/work.pub",
		"curator config build-ssh add other.example.test/tools --agent --identity ~/.ssh/id_ed25519.pub",
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
	// Two repositories of one namespace are one decision, so the commands for
	// it appear once.
	if count := strings.Count(message, "add git.example.test/portals --agent\n"); count != 1 {
		t.Errorf("the namespace command was repeated %d times:\n%s", count, message)
	}
}

func TestAHostWithoutCandidatesSaysSoRatherThanInventingAPath(t *testing.T) {
	selection := BuildSSHSelection{Home: sshHome(t)}
	_, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
	})
	if err == nil {
		t.Fatal("an unselected SSH repository was admitted")
	}
	message := err.Error()
	for _, required := range []string{
		"no SSH agent and no ~/.ssh/*.pub identity were detected on this host",
		"curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/<key>.pub",
	} {
		if !strings.Contains(message, required) {
			t.Errorf("diagnostic missing %q:\n%s", required, message)
		}
	}
}

func TestAPromptedScopeCoversTheRepositoryAndIsReportedAsItsSource(t *testing.T) {
	rows := []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
		sshRow("portals", "build-agent", "git.example.test/portals/agent"),
	}
	selection := discoveredSelection(t, "id_ed25519.pub")
	selection.DefaultKnownHosts = operatorFile(t, filepath.Join(canonicalPath(t, t.TempDir()), "known_hosts"))

	var asked []BuildSSHRequest
	var offered BuildSSHCandidates
	selection.Resolve = func(missing []BuildSSHRequest, candidates BuildSSHCandidates) (map[string]config.BuildSSHCredential, error) {
		asked, offered = missing, candidates
		return map[string]config.BuildSSHCredential{
			"git.example.test/portals": {
				Scope: "git.example.test/portals", Agent: true, AgentSocket: "/run/agent.sock",
			},
		}, nil
	}
	credentials, provenance, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	// Both repositories were named to the resolver, and the one scope it
	// returned covers both by the ordinary longest-scope rule.
	if len(asked) != 2 || asked[0].DefaultScope != "git.example.test/portals" {
		t.Fatalf("resolver was asked about %+v", asked)
	}
	if offered.AgentSocket != "/run/agent.sock" || len(offered.Identities) != 1 {
		t.Fatalf("resolver was offered %+v", offered)
	}
	for _, command := range []string{"build-tool", "build-agent"} {
		selected := resolvedFor(t, credentials, "portals", command)
		if selected.AgentSocket != "/run/agent.sock" {
			t.Fatalf("%s selected %+v", command, selected)
		}
	}
	want := []string{
		`external build ssh: git.example.test/portals/app (command "build-tool" of skill "portals") <- config scope "git.example.test/portals"`,
		`external build ssh: git.example.test/portals/agent (command "build-agent" of skill "portals") <- config scope "git.example.test/portals"`,
	}
	if strings.Join(provenance, "\n") != strings.Join(want, "\n") {
		t.Fatalf("provenance =\n%s\nwant\n%s", strings.Join(provenance, "\n"), strings.Join(want, "\n"))
	}
}

func TestProvenanceNamesEverySourceOnePerRepository(t *testing.T) {
	knownHosts := operatorFile(t, filepath.Join(canonicalPath(t, t.TempDir()), "known_hosts"))
	selection := BuildSSHSelection{
		Home:              sshHome(t),
		DefaultKnownHosts: knownHosts,
		Scopes: map[string]config.BuildSSHCredential{
			"git.example.test/portals": {Scope: "git.example.test/portals", Identity: "/operator/portals.pub"},
		},
	}
	rows := []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
		externalRow("public", "build", "network-git", "git.example.test/open/lib", "https"),
	}
	_, scoped, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(scoped) != 1 || !strings.Contains(scoped[0], `<- config scope "git.example.test/portals"`) {
		t.Fatalf("scoped provenance = %v", scoped)
	}

	selection.RunWide = BuildSSHFlags{Identity: "/operator/run.pub"}
	_, runWide, err := resolveBuildSSH(selection, rows)
	if err != nil {
		t.Fatal(err)
	}
	// A run-wide selection outranks the configured scope, and the report says
	// which one actually applied.
	if len(runWide) != 1 || !strings.Contains(runWide[0], "<- operator flags/env") {
		t.Fatalf("run-wide provenance = %v", runWide)
	}
	if !strings.Contains(runWide[0], `git.example.test/portals/app (command "build-tool" of skill "portals")`) {
		t.Fatalf("provenance does not name the repository it describes: %v", runWide)
	}
}

func TestAResolverThatCoversNothingStillFailsClosed(t *testing.T) {
	selection := discoveredSelection(t, "id_ed25519.pub")
	selection.Resolve = func([]BuildSSHRequest, BuildSSHCandidates) (map[string]config.BuildSSHCredential, error) {
		// A scope that covers a different host is not an answer to this
		// question, and must not be mistaken for one.
		return map[string]config.BuildSSHCredential{
			"other.example.test": {Scope: "other.example.test", Agent: true},
		}, nil
	}
	_, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
	})
	if err == nil || !strings.Contains(err.Error(), buildrepo.CodeSSHCredentialMissing) {
		t.Fatalf("err = %v, want the fail-closed diagnostic", err)
	}
}

func TestAnAbortedResolverFailsTheRunWithItsOwnError(t *testing.T) {
	selection := discoveredSelection(t, "id_ed25519.pub")
	selection.Resolve = func([]BuildSSHRequest, BuildSSHCandidates) (map[string]config.BuildSSHCredential, error) {
		return nil, ErrBuildSSHAborted
	}
	_, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
	})
	if !errors.Is(err, ErrBuildSSHAborted) {
		t.Fatalf("err = %v, want %v", err, ErrBuildSSHAborted)
	}
}

func TestPromptedScopesDoNotMutateTheLoadedConfiguration(t *testing.T) {
	scopes := map[string]config.BuildSSHCredential{}
	selection := discoveredSelection(t, "id_ed25519.pub")
	selection.Scopes = scopes
	selection.DefaultKnownHosts = operatorFile(t, filepath.Join(canonicalPath(t, t.TempDir()), "known_hosts"))
	selection.Resolve = func([]BuildSSHRequest, BuildSSHCandidates) (map[string]config.BuildSSHCredential, error) {
		return map[string]config.BuildSSHCredential{
			"git.example.test/portals": {Scope: "git.example.test/portals", Agent: true, AgentSocket: "/run/agent.sock"},
		}, nil
	}
	if _, _, err := resolveBuildSSH(selection, []plannedExternal{
		sshRow("portals", "build-tool", "git.example.test/portals/app"),
	}); err != nil {
		t.Fatal(err)
	}
	// The run matched against its own copy: what the process loaded from disk
	// is still what is on disk.
	if len(scopes) != 0 {
		t.Fatalf("the loaded scope set was mutated behind the run: %+v", scopes)
	}
}

// probeOnlyToolchain answers the dry-run toolchain probe and nothing else, so
// a plan reaches credential resolution without a real Go toolchain.
type probeOnlyToolchain struct{}

func (probeOnlyToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	return buildmeta.Target{GOOS: "darwin", GOARCH: "arm64"}, buildmeta.Toolchain{
		Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
		GoVersion: "go version go1.25.5 darwin/arm64", ContentSHA256: "sha256:" + strings.Repeat("a", 64),
	}, nil
}

func (probeOnlyToolchain) Establish(context.Context) (BuildSession, error) {
	return nil, errors.New("this plan never stages")
}

// sshBuildNode is one skill declaring a single external build command over SSH.
func sshBuildNode(skill, command, identity string) *closure.Node {
	return &closure.Node{
		Name: skill,
		Spec: &skillspec.Spec{
			SchemaVersion: 7,
			BuildRepositories: map[string]skillspec.BuildRepository{
				"tools": {
					Name: "tools", Git: "git@" + identity + ".git", Identity: identity,
					Transport:    "ssh",
					LockedCommit: skillspec.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("b", 40)},
				},
			},
			Commands: map[string]skillspec.Command{
				command: {
					Name: command, Type: "build", Driver: "go-repository-v1",
					Repository: "tools", Target: "tool",
				},
			},
		},
		Edges: []closure.Edge{{Consumer: "project", Mode: "full"}},
	}
}

func TestTheCredentialPrecheckRunsBeforeAnyFetch(t *testing.T) {
	fetched := 0
	deps := ExternalDeps{
		StoreRoot: t.TempDir(),
		Acquire: func(context.Context, ExternalSource) (*buildrepo.Snapshot, error) {
			fetched++
			return nil, errors.New("the fetch should never have been reached")
		},
		Audit:    func(context.Context, buildrepo.AuditSubject) error { return nil },
		BuildSSH: discoveredSelection(t, "id_ed25519.pub"),
	}
	nodes := []*closure.Node{
		sshBuildNode("portals", "build-tool", "git.example.test/portals/app"),
		sshBuildNode("infra", "build-kit", "other.example.test/tools/kit"),
	}
	_, err := planExternalBuilds(context.Background(), "project", "project/id", t.TempDir(),
		nodes, nil, probeOnlyToolchain{}, deps, true)
	if err == nil {
		t.Fatal("a plan with no selected credentials succeeded")
	}
	if !strings.Contains(err.Error(), buildrepo.CodeSSHCredentialMissing) {
		t.Fatalf("err = %v, want the fail-closed diagnostic", err)
	}
	// The whole point of a precheck: not one repository was reached over the
	// network before the run refused.
	if fetched != 0 {
		t.Fatalf("%d fetches ran before credentials were resolved", fetched)
	}
	// Both repositories are named at once, rather than the run stopping at the
	// first one part way through the closure.
	if !strings.Contains(err.Error(), "other.example.test/tools/kit") {
		t.Fatalf("the second repository was not reported:\n%v", err)
	}
}

func TestADryRunReportsThePerRepositoryCredentialSource(t *testing.T) {
	selection := BuildSSHSelection{
		Home:              sshHome(t),
		DefaultKnownHosts: operatorFile(t, filepath.Join(canonicalPath(t, t.TempDir()), "known_hosts")),
		Scopes: map[string]config.BuildSSHCredential{
			"git.example.test/portals": {Scope: "git.example.test/portals", Identity: "/operator/portals.pub"},
		},
	}
	// The fetch is deliberately made to fail: what is asserted is the report
	// the plan carries, which is written before any repository is reached.
	deps := ExternalDeps{
		StoreRoot: t.TempDir(),
		Acquire: func(context.Context, ExternalSource) (*buildrepo.Snapshot, error) {
			return nil, errors.New("no network in this test")
		},
		Audit:    func(context.Context, buildrepo.AuditSubject) error { return nil },
		BuildSSH: selection,
	}
	nodes := []*closure.Node{sshBuildNode("portals", "build-tool", "git.example.test/portals/app")}

	plan, err := planExternalBuilds(context.Background(), "project", "project/id", t.TempDir(),
		nodes, nil, probeOnlyToolchain{}, deps, true)
	if err == nil {
		t.Fatal("the stubbed fetch was expected to fail the plan")
	}
	want := `external build ssh: git.example.test/portals/app (command "build-tool" of skill "portals") <- config scope "git.example.test/portals"`
	if len(plan.messages) != 1 || plan.messages[0] != want {
		t.Fatalf("dry-run messages = %v, want %q", plan.messages, want)
	}

	// An install reports through its own build rows instead; the credential
	// report belongs to the mode whose whole output is the report.
	installPlan, err := planExternalBuilds(context.Background(), "project", "project/id", t.TempDir(),
		nodes, nil, probeOnlyToolchain{}, deps, false)
	if err == nil {
		t.Fatal("the stubbed fetch was expected to fail the plan")
	}
	if len(installPlan.messages) != 0 {
		t.Fatalf("an install carried the dry-run report: %v", installPlan.messages)
	}
}

// TestTheCredentialReportIsLabelledByTheScopeThatProducedIt covers the hop
// from the plan to the operator-visible result. The label is what tells a
// multi-project run which project a credential line belongs to, and a report
// that never reaches Result.Messages is a report the operator never sees.
func TestTheCredentialReportIsLabelledByTheScopeThatProducedIt(t *testing.T) {
	plan := externalPlan{messages: []string{
		`external build ssh: git.example.test/portals/app (command "build-tool" of skill "portals") <- config scope "git.example.test/portals"`,
		`external build ssh: other.example.test/tools/kit (command "build-kit" of skill "infra") <- operator flags/env`,
	}}
	for _, label := range []string{"global", "test"} {
		reported := plan.credentialReport(label)
		if len(reported) != len(plan.messages) {
			t.Fatalf("credentialReport(%q) = %v, want one line per repository", label, reported)
		}
		for index, line := range reported {
			if line != label+": "+plan.messages[index] {
				t.Fatalf("credentialReport(%q)[%d] = %q, want the labelled provenance line", label, index, line)
			}
		}
	}
	// An install carries no report, so appending it must not put an empty
	// line into an operator's output either.
	if reported := (externalPlan{}).credentialReport("test"); reported != nil {
		t.Fatalf("a plan with no report produced %v", reported)
	}
}
