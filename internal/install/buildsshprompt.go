package install

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
)

// ErrBuildSSHAborted reports an operator who ended the credential prompt
// without selecting anything. The run then fails closed exactly as it would
// have without a terminal: an abort is a refusal to authenticate, not a
// licence to fall back on ambient SSH state.
var ErrBuildSSHAborted = errors.New(buildrepo.CodeSSHCredentialMissing +
	": credential selection was aborted")

// abortToken ends the prompt from any question it asks.
const abortToken = "q"

// BuildSSHRequest is one external build repository the precheck could not
// cover, as the prompt and the fail-closed diagnostic both describe it.
type BuildSSHRequest struct {
	// Skill and Command name where the repository was declared.
	Skill   string
	Command string
	// Identity is the canonical host/path the repository is locked to. It is
	// the only key credentials are ever matched by (Spec §12.2).
	Identity string
	// DefaultScope is the narrowest scope worth persisting for Identity: the
	// repository namespace, offered as the default of the scope question.
	DefaultScope string
}

// BuildSSHResolver is asked to cover the repositories no run-wide selection
// and no configured scope reached. It returns the scopes it persisted, keyed
// by scope, which the precheck folds into the run's own scope set and matches
// again by the ordinary longest-scope rule.
//
// A nil resolver means the run is non-interactive: it fails closed.
type BuildSSHResolver func(missing []BuildSSHRequest, candidates BuildSSHCandidates) (map[string]config.BuildSSHCredential, error)

// InteractiveBuildSSHResolver asks the operator to cover each unselected
// repository from the material discovery found, and persists only what they
// explicitly chose, under the scope they explicitly confirmed.
//
// The menu lists candidates and nothing else: no entry is pre-applied, the
// default entry still has to be accepted, and persist is reached only after
// both the credential and the scope question have been answered.
func InteractiveBuildSSHResolver(in io.Reader, out io.Writer, persist func(config.BuildSSHCredential) error) BuildSSHResolver {
	return func(missing []BuildSSHRequest, candidates BuildSSHCandidates) (map[string]config.BuildSSHCredential, error) {
		reader := bufio.NewReader(in)
		added := map[string]config.BuildSSHCredential{}
		for _, request := range missing {
			if _, covered := config.MatchBuildSSH(added, request.Identity); covered {
				// A scope the operator just chose already covers this
				// repository, so asking again would invite two answers to one
				// question.
				continue
			}
			credential, err := promptBuildSSHCredential(reader, out, request, candidates)
			if err != nil {
				return nil, err
			}
			if err := persist(credential); err != nil {
				return nil, err
			}
			say(out, "recorded build_ssh scope %s\n", credential.Scope)
			added[credential.Scope] = credential
		}
		return added, nil
	}
}

// say writes one line of the prompt. A terminal that stops accepting output is
// not a reason to stop here: the very next read reports it, and every question
// this prompt asks is followed by a read.
func say(out io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(out, format, args...)
}

// promptBuildSSHCredential runs the two questions one repository needs
// answered: what to authenticate with, and how widely that answer applies.
func promptBuildSSHCredential(reader *bufio.Reader, out io.Writer, request BuildSSHRequest, candidates BuildSSHCandidates) (config.BuildSSHCredential, error) {
	say(out, "\n%s: %s needs SSH credentials (command %q of skill %q)\n",
		buildrepo.CodeSSHCredentialMissing, request.Identity, request.Command, request.Skill)
	options := buildSSHOptions(candidates)
	writeBuildSSHMenu(out, options, candidates)
	credential, err := readBuildSSHChoice(reader, out, options)
	if err != nil {
		return config.BuildSSHCredential{}, err
	}
	scope, err := readBuildSSHScope(reader, out, request.DefaultScope)
	if err != nil {
		return config.BuildSSHCredential{}, err
	}
	credential.Scope = scope
	if err := config.ValidateBuildSSH(credential); err != nil {
		return config.BuildSSHCredential{}, err
	}
	return credential, nil
}

// buildSSHOption is one menu entry: a label the operator reads and the
// credential shape choosing it would record.
type buildSSHOption struct {
	label      string
	credential config.BuildSSHCredential
}

// buildSSHOptions turns discovered material into the offered choices. The
// first entry is the default — the agent pinned to the first discovered key,
// which is the only shape that both reuses a loaded key and stops the agent
// offering every other key to the destination.
func buildSSHOptions(candidates BuildSSHCandidates) []buildSSHOption {
	var options []buildSSHOption
	if candidates.AgentSocket != "" && len(candidates.Identities) > 0 {
		options = append(options, buildSSHOption{
			label:      "agent, pinned to " + candidates.Identities[0],
			credential: config.BuildSSHCredential{Agent: true, Identity: candidates.Identities[0]},
		})
	}
	if candidates.AgentSocket != "" {
		options = append(options, buildSSHOption{
			label:      candidates.AgentSummary() + ", any key it holds",
			credential: config.BuildSSHCredential{Agent: true},
		})
	}
	for _, identity := range candidates.Identities {
		options = append(options, buildSSHOption{
			label:      "identity " + identity,
			credential: config.BuildSSHCredential{Identity: identity},
		})
	}
	return options
}

// writeBuildSSHMenu prints the discovered material. Discovery only lists: the
// menu is the last point at which nothing has been selected.
func writeBuildSSHMenu(out io.Writer, options []buildSSHOption, candidates BuildSSHCandidates) {
	if len(options) == 0 {
		say(out, "no SSH agent and no ~/.ssh/*.pub identity were detected on this host\n")
	}
	for index, option := range options {
		suffix := ""
		if index == 0 {
			suffix = "  [default]"
		}
		say(out, "  %d) %s%s\n", index+1, option.label, suffix)
	}
	if candidates.MoreIdentities > 0 {
		say(out, "  (%d further ~/.ssh/*.pub file(s) are not listed; choose %s to name one)\n",
			candidates.MoreIdentities, manualToken)
	}
	say(out, "  %s) enter an identity path\n", manualToken)
	say(out, "  %s) abort\n", abortToken)
}

// manualToken selects the free-form identity path entry.
const manualToken = "m"

// readBuildSSHChoice reads one credential choice. An unparsable answer is
// asked again rather than resolved to a default, since the default authorizes
// a key just as much as any other entry does.
func readBuildSSHChoice(reader *bufio.Reader, out io.Writer, options []buildSSHOption) (config.BuildSSHCredential, error) {
	for {
		prompt := fmt.Sprintf("credential [%s, %s]", manualToken, abortToken)
		if len(options) > 0 {
			prompt = fmt.Sprintf("credential [1-%d, %s, %s] (default 1)", len(options), manualToken, abortToken)
		}
		answer, err := ask(reader, out, prompt)
		if err != nil {
			return config.BuildSSHCredential{}, err
		}
		switch {
		case answer == abortToken:
			return config.BuildSSHCredential{}, ErrBuildSSHAborted
		case answer == "" && len(options) > 0:
			return options[0].credential, nil
		case answer == manualToken:
			return readBuildSSHManualIdentity(reader, out)
		}
		index := 0
		if _, err := fmt.Sscanf(answer, "%d", &index); err == nil && index >= 1 && index <= len(options) {
			return options[index-1].credential, nil
		}
		say(out, "curator: %q is not one of the offered choices\n", answer)
	}
}

// readBuildSSHManualIdentity reads an identity path the operator types, for
// the key that lives outside ~/.ssh or was not listed.
func readBuildSSHManualIdentity(reader *bufio.Reader, out io.Writer) (config.BuildSSHCredential, error) {
	for {
		answer, err := ask(reader, out,
			fmt.Sprintf("identity path (%s, or %s to abort)", config.BuildSSHPathRule, abortToken))
		if err != nil {
			return config.BuildSSHCredential{}, err
		}
		if answer == abortToken {
			return config.BuildSSHCredential{}, ErrBuildSSHAborted
		}
		if config.ValidBuildSSHPath(answer) {
			return config.BuildSSHCredential{Identity: answer}, nil
		}
		say(out, "curator: identity path %s\n", config.BuildSSHPathRule)
	}
}

// readBuildSSHScope reads how widely the chosen credential applies. Nothing is
// persisted before this question is answered: the operator authorizes a scope,
// not just a key.
func readBuildSSHScope(reader *bufio.Reader, out io.Writer, fallback string) (string, error) {
	for {
		answer, err := ask(reader, out,
			fmt.Sprintf("scope [%s] (%s to abort)", fallback, abortToken))
		if err != nil {
			return "", err
		}
		if answer == abortToken {
			return "", ErrBuildSSHAborted
		}
		if answer == "" {
			answer = fallback
		}
		if config.ValidBuildSSHScope(answer) {
			return answer, nil
		}
		say(out, "curator: scope %s\n", config.BuildSSHScopeRule)
	}
}

// ask writes one prompt and reads one trimmed answer. End of input is an
// abort: a prompt nobody is there to answer must not resolve to a default.
func ask(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	say(out, "%s: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", ErrBuildSSHAborted
	}
	return strings.TrimSpace(line), nil
}
