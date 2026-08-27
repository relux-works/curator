package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sshHome creates an operator home holding the named ~/.ssh entries.
func sshHome(t *testing.T, names ...string) string {
	t.Helper()
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		operatorFile(t, filepath.Join(home, ".ssh", name))
	}
	return home
}

// countedAgent injects an answer to the one question discovery asks the agent,
// so a test never depends on a live one.
func countedAgent(count int, known bool) func(string) (int, bool) {
	return func(string) (int, bool) { return count, known }
}

func TestDiscoveryListsTheLiveAgentAndPublicIdentitiesOnly(t *testing.T) {
	home := sshHome(t, "id_ed25519", "id_ed25519.pub", "work.pub", "known_hosts", "config")
	// A directory named like a key is material nobody can offer to anything.
	if err := os.Mkdir(filepath.Join(home, ".ssh", "stale.pub"), 0o700); err != nil {
		t.Fatal(err)
	}
	candidates := discoverBuildSSHCandidates(BuildSSHSelection{
		Home: home, AgentSocket: "/run/agent.sock", AgentKeys: countedAgent(2, true),
	})
	want := []string{"~/.ssh/id_ed25519.pub", "~/.ssh/work.pub"}
	if strings.Join(candidates.Identities, ",") != strings.Join(want, ",") {
		t.Fatalf("identities = %v, want %v", candidates.Identities, want)
	}
	if candidates.MoreIdentities != 0 {
		t.Fatalf("MoreIdentities = %d, want 0", candidates.MoreIdentities)
	}
	if candidates.Empty() {
		t.Fatal("discovery with an agent and two identities reported nothing")
	}
	if got := candidates.AgentSummary(); got != "SSH agent at /run/agent.sock holding 2 keys" {
		t.Fatalf("agent summary = %q", got)
	}
}

func TestDiscoveryDegradesWhenTheAgentCannotBeAsked(t *testing.T) {
	home := sshHome(t, "id.pub")
	unknown := discoverBuildSSHCandidates(BuildSSHSelection{
		Home: home, AgentSocket: "/run/agent.sock", AgentKeys: countedAgent(0, false),
	})
	// An unanswerable key count does not disqualify the agent: the count is
	// decoration on the entry, not a gate on it.
	if unknown.AgentSocket != "/run/agent.sock" {
		t.Fatalf("agent dropped when its key count was unavailable: %+v", unknown)
	}
	if summary := unknown.AgentSummary(); !strings.Contains(summary, "key count unavailable") {
		t.Fatalf("agent summary = %q", summary)
	}
	empty := discoverBuildSSHCandidates(BuildSSHSelection{
		Home: home, AgentSocket: "/run/agent.sock", AgentKeys: countedAgent(0, true),
	})
	if summary := empty.AgentSummary(); summary != "SSH agent at /run/agent.sock holding 0 keys" {
		t.Fatalf("empty agent summary = %q", summary)
	}
	single := discoverBuildSSHCandidates(BuildSSHSelection{
		Home: home, AgentSocket: "/run/agent.sock", AgentKeys: countedAgent(1, true),
	})
	if summary := single.AgentSummary(); summary != "SSH agent at /run/agent.sock holding 1 key" {
		t.Fatalf("single-key agent summary = %q", summary)
	}
}

func TestDiscoveryWithoutAnAgentOrAHomeFindsNothing(t *testing.T) {
	noAgent := discoverBuildSSHCandidates(BuildSSHSelection{Home: sshHome(t)})
	if !noAgent.Empty() || noAgent.AgentSummary() != "" {
		t.Fatalf("empty ~/.ssh and no agent reported %+v", noAgent)
	}
	// A home the environment never named, and a home without ~/.ssh, are both
	// simply nothing to list rather than a failure.
	if got := discoverBuildSSHCandidates(BuildSSHSelection{}); !got.Empty() {
		t.Fatalf("discovery without a home reported %+v", got)
	}
	if got := discoverBuildSSHCandidates(BuildSSHSelection{Home: t.TempDir()}); !got.Empty() {
		t.Fatalf("discovery without a ~/.ssh reported %+v", got)
	}
}

func TestDiscoveryCapsTheIdentityListingAndReportsTheRemainder(t *testing.T) {
	var names []string
	for index := 0; index < maxListedIdentities+3; index++ {
		names = append(names, string(rune('a'+index))+".pub")
	}
	candidates := discoverBuildSSHCandidates(BuildSSHSelection{Home: sshHome(t, names...)})
	if len(candidates.Identities) != maxListedIdentities {
		t.Fatalf("listed %d identities, want %d", len(candidates.Identities), maxListedIdentities)
	}
	// A capped listing must say how much it left out; a silently truncated one
	// reads as the whole set.
	if candidates.MoreIdentities != 3 {
		t.Fatalf("MoreIdentities = %d, want 3", candidates.MoreIdentities)
	}
	if candidates.Identities[0] != "~/.ssh/a.pub" {
		t.Fatalf("identities are not sorted: %v", candidates.Identities)
	}
}

func TestAgentKeyCountReadsTheToolOutputRatherThanItsNoise(t *testing.T) {
	for output, want := range map[string]int{
		"":                                    0,
		"The agent has no identities.\n":      0,
		"256 SHA256:aaa operator (ED25519)\n": 1,
		"256 SHA256:aaa a (ED25519)\n256 SHA256:bbb b (ED25519)\n\n": 2,
	} {
		if got := countKeyLines(output); got != want {
			t.Errorf("countKeyLines(%q) = %d, want %d", output, got, want)
		}
	}
}

func TestSSHAddKeyCountDegradesWhenTheSocketIsNotAnAgent(t *testing.T) {
	// Both degrade paths land on the same answer, so this needs no skip: a
	// runner without `ssh-add` and a runner whose socket reaches no agent both
	// report the count as unavailable rather than failing.
	if _, known := sshAgentKeyCount(filepath.Join(t.TempDir(), "absent.sock")); known {
		t.Fatal("a socket that reaches no agent reported a known key count")
	}
}

func TestAddCommandsAreBuiltFromTheDiscoveredCandidates(t *testing.T) {
	candidates := BuildSSHCandidates{
		AgentSocket: "/run/agent.sock", AgentKeys: 1, AgentKeysKnown: true,
		Identities: []string{"~/.ssh/id_ed25519.pub", "~/.ssh/work.pub"},
	}
	commands := buildSSHAddCommands("git.example.test/portals", candidates)
	want := []string{
		"  curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/id_ed25519.pub",
		"  curator config build-ssh add git.example.test/portals --agent",
		"  curator config build-ssh add git.example.test/portals --identity ~/.ssh/id_ed25519.pub",
		"  curator config build-ssh add git.example.test/portals --identity ~/.ssh/work.pub",
	}
	if strings.Join(commands, "\n") != strings.Join(want, "\n") {
		t.Fatalf("commands =\n%s\nwant\n%s", strings.Join(commands, "\n"), strings.Join(want, "\n"))
	}
	// With nothing discovered the command names the shape and marks the path
	// as the operator's to supply, rather than inventing one.
	bare := buildSSHAddCommands("git.example.test/portals", BuildSSHCandidates{})
	if len(bare) != 2 || !strings.Contains(bare[0], "~/.ssh/<key>.pub") {
		t.Fatalf("commands without candidates = %v", bare)
	}
	// An agent-only host offers no pinned entry, because there is no key to
	// pin it to.
	agentOnly := buildSSHAddCommands("git.example.test/portals", BuildSSHCandidates{AgentSocket: "/run/agent.sock"})
	if len(agentOnly) != 1 || !strings.HasSuffix(agentOnly[0], "--agent") {
		t.Fatalf("agent-only commands = %v", agentOnly)
	}
}
