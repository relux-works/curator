package install

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// maxListedIdentities bounds the identity files one diagnostic or menu names.
// A host with a large ~/.ssh should not bury the repositories the message is
// actually about; whatever is left over is reported as a count rather than
// dropped silently.
const maxListedIdentities = 8

// agentKeyCountTimeout bounds the one question discovery asks the operator's
// agent. An agent that does not answer is reported as "key count unknown", not
// waited on: the count is decoration on a menu entry, never a gate.
const agentKeyCountTimeout = 5 * time.Second

// BuildSSHCandidates is the authentication material discovery found on this
// host: the operator's live agent and the public identity files under
// ~/.ssh. It is a list, not a selection. Nothing in it is offered to any
// destination until the operator explicitly chooses it, and nothing in it is
// persisted until the operator explicitly names the scope it applies to.
type BuildSSHCandidates struct {
	// AgentSocket is the live SSH_AUTH_SOCK, empty when the operator's
	// environment advertises no agent.
	AgentSocket string
	// AgentKeys is how many identities the live agent reported holding.
	// Meaningful only with AgentKeysKnown; an agent that cannot be asked is
	// still a usable candidate, so the count degrades rather than the entry.
	AgentKeys      int
	AgentKeysKnown bool
	// Identities are the *.pub files under ~/.ssh in sorted order, spelled
	// with the leading `~/` the config grammar accepts, capped at
	// maxListedIdentities.
	Identities []string
	// MoreIdentities counts the *.pub files past the listed ones, so a capped
	// listing says so instead of reading as the whole set.
	MoreIdentities int
}

// Empty reports that discovery found no authentication material at all. The
// operator then has nothing to pick from and has to name a path by hand.
func (c BuildSSHCandidates) Empty() bool {
	return c.AgentSocket == "" && len(c.Identities) == 0
}

// AgentSummary describes the live agent for a menu entry or a diagnostic.
func (c BuildSSHCandidates) AgentSummary() string {
	if c.AgentSocket == "" {
		return ""
	}
	switch {
	case !c.AgentKeysKnown:
		return fmt.Sprintf("SSH agent at %s (key count unavailable)", c.AgentSocket)
	case c.AgentKeys == 1:
		return fmt.Sprintf("SSH agent at %s holding 1 key", c.AgentSocket)
	default:
		return fmt.Sprintf("SSH agent at %s holding %d keys", c.AgentSocket, c.AgentKeys)
	}
}

// discoverBuildSSHCandidates lists the operator's own authentication material.
//
// Discovery reads only state the operator already owns — the agent socket
// their environment advertises and the public keys in their own ~/.ssh — and
// it reads it after the run-wide selection has already failed to cover a
// repository. It never consults package data, and it never turns what it
// finds into a selection.
func discoverBuildSSHCandidates(selection BuildSSHSelection) BuildSSHCandidates {
	candidates := BuildSSHCandidates{AgentSocket: selection.AgentSocket}
	if candidates.AgentSocket != "" {
		count := selection.AgentKeys
		if count == nil {
			count = sshAgentKeyCount
		}
		candidates.AgentKeys, candidates.AgentKeysKnown = count(candidates.AgentSocket)
	}
	candidates.Identities, candidates.MoreIdentities = publicIdentities(selection.Home)
	return candidates
}

// publicIdentities lists the *.pub files directly under ~/.ssh. Only regular
// files are listed: a directory or a dangling link named `id.pub` is not
// material the operator can offer to anything.
func publicIdentities(home string) ([]string, int) {
	if home == "" {
		return nil, 0
	}
	entries, err := os.ReadDir(filepath.Join(home, ".ssh"))
	if err != nil {
		return nil, 0
	}
	var names []string
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		names = append(names, "~/.ssh/"+entry.Name())
	}
	sort.Strings(names)
	if len(names) > maxListedIdentities {
		return names[:maxListedIdentities], len(names) - maxListedIdentities
	}
	return names, 0
}

// sshAgentKeyCount asks the operator's own agent how many identities it holds.
//
// Every failure degrades to "unknown" rather than to an error: `ssh-add` may
// not be installed, the socket may be stale, and the agent may refuse to
// answer. None of that makes the agent an invalid choice — it only makes the
// count unavailable, and the count is decoration on a menu entry.
func sshAgentKeyCount(socket string) (int, bool) {
	toolPath, err := exec.LookPath("ssh-add")
	if err != nil {
		return 0, false
	}
	if toolPath, err = filepath.EvalSymlinks(toolPath); err != nil {
		return 0, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), agentKeyCountTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, toolPath, "-l") // #nosec G204 -- resolved operator tool, fixed argument.
	// The operator's whole environment is deliberately not inherited: the one
	// thing the tool needs is the socket this selection already resolved. The
	// two names kept alongside it are what a process needs to start at all on
	// Windows, and neither can redirect the question being asked.
	command.Env = []string{"SSH_AUTH_SOCK=" + socket}
	for _, name := range []string{"SYSTEMROOT", "PATH"} {
		if value := os.Getenv(name); value != "" {
			command.Env = append(command.Env, name+"="+value)
		}
	}
	output, err := command.Output()
	if err != nil {
		// `ssh-add -l` exits 1 for an agent that holds nothing, which is a
		// real answer and not a failure to reach it.
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return 0, true
		}
		return 0, false
	}
	return countKeyLines(string(output)), true
}

// countKeyLines counts the identities `ssh-add -l` listed. The no-identities
// notice is printed on stdout by some builds, so it is recognised rather than
// counted as a key.
func countKeyLines(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "has no identities") {
			continue
		}
		count++
	}
	return count
}

// buildSSHAddCommands renders the ready-to-run commands that would select each
// discovered candidate for one scope, most specific first. An operator can
// paste any line unchanged; nothing here has been applied.
func buildSSHAddCommands(scope string, candidates BuildSSHCandidates) []string {
	prefix := "  curator config build-ssh add " + scope
	var commands []string
	if candidates.AgentSocket != "" && len(candidates.Identities) > 0 {
		commands = append(commands, prefix+" --agent --identity "+candidates.Identities[0])
	}
	if candidates.AgentSocket != "" {
		commands = append(commands, prefix+" --agent")
	}
	for _, identity := range candidates.Identities {
		commands = append(commands, prefix+" --identity "+identity)
	}
	if len(commands) == 0 {
		// Nothing was discovered, so the command names the shape rather than
		// a path: a placeholder the operator must replace is honest, an
		// invented path is not.
		commands = append(commands,
			prefix+" --agent --identity ~/.ssh/<key>.pub",
			prefix+" --agent")
	}
	return commands
}
