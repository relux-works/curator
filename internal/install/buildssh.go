package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
)

// The run-wide operator SSH selection an install accepts from the environment.
// Command-line flags carry the same three values and win over all of them.
const (
	// EnvBuildSSHIdentity names the identity file offered to every external
	// build repository of the run.
	EnvBuildSSHIdentity = "CURATOR_BUILD_SSH_IDENTITY"
	// EnvBuildSSHAgent names the agent socket, or the literal "auto" to adopt
	// the operator's own live agent.
	EnvBuildSSHAgent = "CURATOR_BUILD_SSH_AGENT"
	// EnvBuildSSHKnownHosts names the host keys the fetch verifies against.
	EnvBuildSSHKnownHosts = "CURATOR_BUILD_SSH_KNOWN_HOSTS"
	// BuildSSHAgentAuto adopts the operator's live agent socket instead of a
	// literal path. A macOS agent socket is per login session, so a path is
	// the one thing a persisted selection cannot keep current.
	BuildSSHAgentAuto = "auto"
	// envAgentSocket is where the operator's own agent advertises itself.
	envAgentSocket = "SSH_AUTH_SOCK"
)

var errBuildSSHAgentUnset = fmt.Errorf(
	"%s: an SSH agent was requested but %s is not set",
	buildrepo.CodeSSHCredentialMissing, envAgentSocket)

// BuildSSHFlags is one run-wide operator SSH selection as a command line
// spells it. An empty field means the operator named nothing there.
type BuildSSHFlags struct {
	Identity   string
	Agent      string
	KnownHosts string
}

// BuildSSHSelection is everything one install may select operator SSH
// credentials from. Package data never appears in it: the only key a
// repository is matched by is the canonical identity it is already locked to
// (Spec §12.2).
type BuildSSHSelection struct {
	// RunWide is the merged flags/env selection. Anything it names covers
	// every repository of the run, ahead of every configured scope.
	RunWide BuildSSHFlags
	// Scopes are the operator's configured build_ssh entries. The longest
	// scope matching a repository's canonical identity wins.
	Scopes map[string]config.BuildSSHCredential
	// AgentSocket is the operator's own live agent socket, read once at
	// process entry. A selection that asks for "the agent" resolves to it.
	AgentSocket string
	// DefaultKnownHosts is the operator's own host key file, used when no
	// selection names one. The fetch pins StrictHostKeyChecking=yes and has
	// no other source of truth for host keys.
	DefaultKnownHosts string
	// Home is the operator's own home directory, read once at process entry.
	// It is the only place candidate discovery looks for identity files, so a
	// project-owned environment cannot redirect it mid-run.
	Home string
	// Resolve covers the repositories no run-wide selection and no configured
	// scope reached. A nil resolver — every non-interactive surface — fails
	// closed instead.
	Resolve BuildSSHResolver
	// AgentKeys reports how many keys a live agent socket holds. nil asks the
	// operator's own `ssh-add`; a test injects an answer instead of an agent.
	AgentKeys func(socket string) (int, bool)
}

// CaptureBuildSSHSelection merges the run-wide selection from explicit flags
// and the CURATOR_BUILD_SSH_* environment, flags first, and binds the
// configured scopes it falls back to.
//
// Callers invoke this at process entry, before any project-owned state can
// influence the environment it reads.
func CaptureBuildSSHSelection(cfg *config.Config, flags BuildSSHFlags, environ func(string) string) BuildSSHSelection {
	if environ == nil {
		environ = func(string) string { return "" }
	}
	selection := BuildSSHSelection{RunWide: flags, AgentSocket: environ(envAgentSocket)}
	if cfg != nil {
		selection.Scopes = cfg.BuildSSH
	}
	for _, field := range []struct {
		name  string
		value *string
	}{
		{EnvBuildSSHIdentity, &selection.RunWide.Identity},
		{EnvBuildSSHAgent, &selection.RunWide.Agent},
		{EnvBuildSSHKnownHosts, &selection.RunWide.KnownHosts},
	} {
		if *field.value == "" {
			*field.value = environ(field.name)
		}
	}
	selection.Home = operatorHome(environ)
	selection.DefaultKnownHosts = operatorKnownHosts(environ)
	return selection
}

// operatorHome is the operator's own home directory as their environment
// reports it, read at process entry alongside everything else the selection
// captures.
func operatorHome(environ func(string) string) string {
	for _, name := range []string{"HOME", "USERPROFILE"} {
		if home := environ(name); home != "" {
			return home
		}
	}
	return ""
}

// operatorKnownHosts locates the operator's own host key file. Absent, the
// selection carries none and policy construction refuses rather than trusting
// an unverified host.
func operatorKnownHosts(environ func(string) string) string {
	for _, name := range []string{"HOME", "USERPROFILE"} {
		home := environ(name)
		if home == "" {
			continue
		}
		candidate := filepath.Join(home, ".ssh", "known_hosts")
		if info, err := os.Lstat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return ""
}

// buildSSHKey identifies one external build repository within a run.
type buildSSHKey struct {
	skill   string
	command string
}

// resolveBuildSSH selects operator SSH credentials for every planned external
// build repository.
//
// Precedence: an explicit run-wide selection covers every repository;
// otherwise the longest matching build_ssh scope covers its own. A repository
// reached over HTTPS, and one an operator substitution replaced with a local
// path, need no selection at all. Anything still unselected fails closed with
// the exact commands that would fix it — never with a fallback onto whatever
// the operator's ambient SSH state happens to offer.
func resolveBuildSSH(selection BuildSSHSelection, rows []plannedExternal) (map[buildSSHKey]buildrepo.OperatorSSHCredentials, []string, error) {
	credentials := map[buildSSHKey]buildrepo.OperatorSSHCredentials{}
	var provenance []string
	var missing []plannedExternal
	// The scopes this run matches against. It starts as the configured set and
	// only ever grows by an explicit operator choice, and it is a copy so a
	// prompted answer cannot mutate the loaded configuration behind the run.
	scopes := make(map[string]config.BuildSSHCredential, len(selection.Scopes))
	for scope, credential := range selection.Scopes {
		scopes[scope] = credential
	}
	runWide, runWideErr := selection.runWideCredentials()
	for _, row := range rows {
		if !needsBuildSSH(row) {
			continue
		}
		if runWideErr != nil {
			return nil, nil, runWideErr
		}
		if runWide.Selected() {
			credentials[buildSSHKeyFor(row)] = runWide
			provenance = append(provenance, buildSSHProvenance(row, "operator flags/env"))
			continue
		}
		selected, source, matched, err := selection.matchScope(scopes, row)
		if err != nil {
			return nil, nil, err
		}
		if !matched {
			missing = append(missing, row)
			continue
		}
		credentials[buildSSHKeyFor(row)] = selected
		provenance = append(provenance, buildSSHProvenance(row, source))
	}
	if len(missing) == 0 {
		return credentials, provenance, nil
	}

	// Everything below is the precheck's last chance to cover a repository,
	// and it runs before the first fetch of the run rather than part way
	// through it. Discovery happens here, not at process entry: it is only
	// ever needed once a repository is actually uncovered.
	candidates := discoverBuildSSHCandidates(selection)
	if selection.Resolve == nil {
		return nil, nil, missingBuildSSHError(missing, candidates)
	}
	chosen, err := selection.Resolve(buildSSHRequests(missing), candidates)
	if err != nil {
		return nil, nil, err
	}
	for scope, credential := range chosen {
		scopes[scope] = credential
	}
	var stillMissing []plannedExternal
	for _, row := range missing {
		selected, source, matched, err := selection.matchScope(scopes, row)
		if err != nil {
			return nil, nil, err
		}
		if !matched {
			stillMissing = append(stillMissing, row)
			continue
		}
		credentials[buildSSHKeyFor(row)] = selected
		provenance = append(provenance, buildSSHProvenance(row, source))
	}
	if len(stillMissing) > 0 {
		return nil, nil, missingBuildSSHError(stillMissing, candidates)
	}
	return credentials, provenance, nil
}

// matchScope selects the longest scope covering one repository and reports
// where it came from, for the per-repository provenance a dry run prints.
func (s BuildSSHSelection) matchScope(scopes map[string]config.BuildSSHCredential, row plannedExternal) (buildrepo.OperatorSSHCredentials, string, bool, error) {
	scope, matched := config.MatchBuildSSH(scopes, row.effective.Identity)
	if !matched {
		return buildrepo.OperatorSSHCredentials{}, "", false, nil
	}
	selected, err := s.scopeCredentials(scope)
	if err != nil {
		if errors.Is(err, errBuildSSHAgentUnset) {
			return buildrepo.OperatorSSHCredentials{}, "", false, err
		}
		return buildrepo.OperatorSSHCredentials{}, "", false, fmt.Errorf(
			"build_ssh scope %q selected for %s: %w", scope.Scope, row.effective.Identity, err)
	}
	return selected, fmt.Sprintf("config scope %q", scope.Scope), true, nil
}

func buildSSHKeyFor(row plannedExternal) buildSSHKey {
	return buildSSHKey{skill: row.node.Name, command: row.command.Name}
}

// buildSSHProvenance is the one line a dry run prints per repository that
// needs credentials, naming the repository and where its selection came from.
func buildSSHProvenance(row plannedExternal, source string) string {
	return fmt.Sprintf("external build ssh: %s (command %q of skill %q) <- %s",
		row.effective.Identity, row.command.Name, row.node.Name, source)
}

// buildSSHRequests describes the uncovered repositories for the resolver, so
// the prompt reads planned rows only through this narrow projection.
func buildSSHRequests(missing []plannedExternal) []BuildSSHRequest {
	requests := make([]BuildSSHRequest, 0, len(missing))
	for _, row := range missing {
		requests = append(requests, BuildSSHRequest{
			Skill:        row.node.Name,
			Command:      row.command.Name,
			Identity:     row.effective.Identity,
			DefaultScope: defaultBuildSSHScope(row.effective.Identity),
		})
	}
	return requests
}

// needsBuildSSH reports whether a planned repository is actually fetched over
// SSH. The effective state is what decides it: a substitution that redirects a
// declared SSH repository onto HTTPS, or onto a local path, moves the
// repository off the SSH transport with it.
func needsBuildSSH(row plannedExternal) bool {
	return row.effective.IdentityKind == "network-git" && row.effective.Transport == "ssh"
}

// runWideCredentials materializes the explicit flags/env selection.
func (s BuildSSHSelection) runWideCredentials() (buildrepo.OperatorSSHCredentials, error) {
	credentials := buildrepo.OperatorSSHCredentials{
		Identity:    s.RunWide.Identity,
		AgentSocket: s.RunWide.Agent,
		KnownHosts:  s.RunWide.KnownHosts,
	}
	if s.RunWide.Agent == BuildSSHAgentAuto {
		credentials.AgentSocket = s.AgentSocket
		if credentials.AgentSocket == "" {
			return buildrepo.OperatorSSHCredentials{}, errBuildSSHAgentUnset
		}
	}
	credentials.KnownHosts = s.knownHosts(credentials.KnownHosts)
	return credentials, nil
}

// knownHosts resolves the host keys one selection verifies against: the ones
// it names, else the run-wide ones, else the operator's own. A scope selects
// authentication material, not who the destination is allowed to be, so an
// explicit run-wide known-hosts file still applies to a scoped selection.
func (s BuildSSHSelection) knownHosts(named string) string {
	for _, candidate := range []string{named, s.RunWide.KnownHosts, s.DefaultKnownHosts} {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

// scopeCredentials materializes one configured scope. A leading `~/` resolves
// against the operator home, and an entry that records the agent without a
// socket resolves to the live one.
func (s BuildSSHSelection) scopeCredentials(scope config.BuildSSHCredential) (buildrepo.OperatorSSHCredentials, error) {
	expanded := scope.Expanded()
	credentials := buildrepo.OperatorSSHCredentials{
		Identity: expanded.Identity, KnownHosts: expanded.KnownHosts,
	}
	if expanded.Agent {
		credentials.AgentSocket = expanded.AgentSocket
		if credentials.AgentSocket == "" {
			credentials.AgentSocket = s.AgentSocket
		}
		if credentials.AgentSocket == "" {
			return buildrepo.OperatorSSHCredentials{}, errBuildSSHAgentUnset
		}
	}
	credentials.KnownHosts = s.knownHosts(credentials.KnownHosts)
	return credentials, nil
}

// missingBuildSSHError names every unselected repository at once, the material
// discovery found on this host, and the exact commands that would select that
// material for each repository's namespace.
//
// The commands are built from the same candidates the interactive menu offers,
// so a run that could not ask still tells the operator precisely what to run.
// Listing is not selecting: nothing here has been applied.
func missingBuildSSHError(missing []plannedExternal, candidates BuildSSHCandidates) error {
	lines := []string{buildrepo.CodeSSHCredentialMissing +
		": external build repositories need SSH credentials:"}
	for _, row := range missing {
		lines = append(lines, fmt.Sprintf("  %s (command %q of skill %q)",
			row.effective.Identity, row.command.Name, row.node.Name))
	}
	if candidates.Empty() {
		lines = append(lines, "no SSH agent and no ~/.ssh/*.pub identity were detected on this host")
	} else {
		lines = append(lines, "detected on this host:")
		if summary := candidates.AgentSummary(); summary != "" {
			lines = append(lines, "  "+summary)
		}
		for _, identity := range candidates.Identities {
			lines = append(lines, "  "+identity)
		}
		if candidates.MoreIdentities > 0 {
			lines = append(lines, fmt.Sprintf(
				"  (%d further ~/.ssh/*.pub file(s) are not listed)", candidates.MoreIdentities))
		}
	}
	lines = append(lines, "select credentials with one of:")
	for _, scope := range missingBuildSSHScopes(missing) {
		lines = append(lines, buildSSHAddCommands(scope, candidates)...)
	}
	lines = append(lines, fmt.Sprintf(
		"or pass --build-ssh-agent/--build-ssh-identity, or set %s/%s",
		EnvBuildSSHAgent, EnvBuildSSHIdentity))
	return errors.New(strings.Join(lines, "\n"))
}

// missingBuildSSHScopes is the deduplicated set of namespaces the uncovered
// repositories fall into, in first-seen order: two repositories of one group
// are one decision, not two.
func missingBuildSSHScopes(missing []plannedExternal) []string {
	seen := map[string]bool{}
	var scopes []string
	for _, row := range missing {
		scope := defaultBuildSSHScope(row.effective.Identity)
		if seen[scope] {
			continue
		}
		seen[scope] = true
		scopes = append(scopes, scope)
	}
	return scopes
}

// defaultBuildSSHScope is the narrowest scope worth persisting for a canonical
// identity: the repository's namespace, so a sibling repository of the same
// group is covered without the operator naming every repository by hand.
func defaultBuildSSHScope(canonical string) string {
	segments := strings.Split(canonical, "/")
	if len(segments) <= 2 {
		return canonical
	}
	return strings.Join(segments[:len(segments)-1], "/")
}

// credentialReport is the per-repository provenance one scope contributes to
// an operator-visible result, each line labelled by the scope it belongs to.
//
// A dry run is the mode whose whole output is the report, so it is the only
// mode the plan populates; every other mode reports through its own build
// rows and this returns nothing.
func (p externalPlan) credentialReport(label string) []string {
	if len(p.messages) == 0 {
		return nil
	}
	reported := make([]string, 0, len(p.messages))
	for _, message := range p.messages {
		reported = append(reported, label+": "+message)
	}
	return reported
}
