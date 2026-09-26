package main

// Explicit draft source closure (skillfile-sources §3, repository-transport
// revision 1): `curator project resolve` and `curator project refresh`.
//
// Both verbs run the same explicit attempt: acquire every Git source alias
// through the machine endpoint policy (declared URL once when no entry
// applies, the machine list with pin/fallback ordering otherwise; a
// logical `repository` without an entry fails
// repository_endpoint_unavailable), resolve refs anew, freeze local
// dirty/untracked bytes and Git commits into a deterministic locked plan,
// and publish the portable lock plus the machine bindings transactionally
// after every gate succeeds. Install and launch never reselect: they
// consume the published lock only. Frozen v1 manifests keep the legacy
// read-only resolve output byte-identically.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

// draftGitReposDir is the managed parent for acquired draft Git trees.
func draftGitReposDir(home string) string { return filepath.Join(home, "draft-git") }

// draftRepoDir maps one project and source alias to a stable acquired-tree
// path. The alias is a portable identifier, so it is a safe final
// component; the project hash keeps same-named aliases of different
// projects apart.
func draftRepoDir(home, projectRoot, alias string) string {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		abs = projectRoot
	}
	sum := sha256.Sum256([]byte(abs + "\x00" + alias))
	return filepath.Join(draftGitReposDir(home), hex.EncodeToString(sum[:])[:16]+"-"+alias)
}

// cmdProjectResolve runs the explicit draft resolve/refresh attempt for one
// project target. verb names the invoked subcommand for diagnostics only.
func (c cli) cmdProjectResolve(cfg *config.Config, target projectTarget, verb string) int {
	payload, err := os.ReadFile(manifest.PathIn(target.Root)) // #nosec G304 -- operator-selected project Skillfile
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(c.stderr, "curator: Skillfile.json not found at or above", target.Root)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	projectManifest, err := manifest.ParseBytes(payload, manifest.PathIn(target.Root))
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", withDraftRemediation(err.Error()))
		return exitFail
	}
	if projectManifest.SchemaVersion != 2 {
		// Frozen v1: read-only path report, exactly as before.
		_, _ = fmt.Fprintf(c.stdout, "alias: %s\npath: %s\nskillfile: %s\n", target.Alias, target.Root, filepath.Join(target.Root, manifest.Name))
		_, _ = fmt.Fprintf(c.stdout, "skills: %s\nbin: %s\n", filepath.Join(target.Root, ".agents", "skills"), filepath.Join(target.Root, ".agents", "bin"))
		return exitOK
	}
	plan, lockPath, bindingsPath, err := c.resolveDraftPlan(cfg, target, projectManifest, payload)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", withDraftRemediation(err.Error()))
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "%s %s: %d skills\nlock: %s\nbindings: %s\nlock_sha256: %s\n",
		verb, target.Alias, len(plan.Lock.Members), lockPath, bindingsPath, plan.Lock.LockSHA256)
	return exitOK
}

// resolveDraftPlan acquires every Git alias, runs the explicit closure,
// and publishes the lock plus bindings transactionally. Resolution never
// reads the previous lock and never reuses a previous snapshot.
func (c cli) resolveDraftPlan(cfg *config.Config, target projectTarget, projectManifest *manifest.Manifest, payload []byte) (*closure.DraftPlan, string, string, error) {
	home := cfg.Home()
	policy, err := config.LoadSourcePolicy("")
	if err != nil {
		return nil, "", "", err
	}
	roots, err := acquireDraftGitRoots(home, target.Root, projectManifest, policy, cfg.AllowedSources, c.stderr)
	if err != nil {
		return nil, "", "", err
	}
	var rootInputs map[string][]string
	if policy != nil {
		rootInputs = policy.RootInputs
	}
	// Explicit resolve/refresh reselects: every applicable repository is
	// refreshed through the admitted acquisition path. Alias trees were
	// just cloned or fetched above, so they are pre-marked in the shared
	// dedup set and pinGitAliases never fetches them twice; legacy and
	// transitive checkouts refresh in ensureRepo (remote-less checkouts
	// stay a no-op, a failed fetch fails the attempt before publication).
	fetched := map[string]bool{}
	for _, root := range roots {
		closure.MarkRepoFetched(fetched, root)
	}
	resolveCfg := closure.DraftResolveConfig{
		ProjectRoot: target.Root,
		Home:        home,
		// The configured dependency repository root serves the legacy
		// and transitive lanes: without it ensureRepo resolves provider
		// checkouts relative to the process working directory.
		SkillsRoot: cfg.SkillsRoot,
		// The machine allowlist gates legacy and transitive acquisitions
		// with the existing closure semantics (empty allows all). Alias
		// roots are already gated in acquireDraftGitRoots above.
		AllowedSources:  cfg.AllowedSources,
		Manifest:        projectManifest,
		ManifestPayload: payload,
		Expansion:       manifest.ExpansionOptions{GitRoots: roots},
		RootInputs:      rootInputs,
		Fetch:           true,
		FetchedRepos:    fetched,
	}
	lockPath := sourcelock.PathIn(target.Root)
	bindingsPath := install.DraftBindingsPath(home, target.Root)
	plan, err := closure.RefreshDraft(resolveCfg, lockPath, bindingsPath)
	if err != nil {
		return nil, "", "", err
	}
	return plan, lockPath, bindingsPath, nil
}

// acquireDraftGitRoots clones or fetches one working tree per Git source
// alias under the manager home and returns the alias-to-tree map for
// expansion. Endpoint selection reuses the landed machine policy planning
// (config.ResolveRepositoryEndpoints); every attempt runs the same git
// binary the legacy lanes use, and fallback across the listed endpoints
// follows the §2 gate (buildrepo.AllowSecondAttempt over the classified
// fetch output). Allowlist checks match the stable canonical identity
// before any network I/O, exactly like the legacy clone gate.
func acquireDraftGitRoots(home, projectRoot string, projectManifest *manifest.Manifest, policy *config.SourcePolicy, allowed []string, stderr io.Writer) (map[string]string, error) {
	aliases := make([]string, 0, len(projectManifest.Sources))
	for alias, source := range projectManifest.Sources {
		if source.Path == "" {
			aliases = append(aliases, alias)
		}
	}
	sort.Strings(aliases)
	roots := make(map[string]string, len(aliases))
	for _, alias := range aliases {
		source := projectManifest.Sources[alias]
		resolution, err := config.ResolveRepositoryEndpoints(policy, source.Git, source.Repository)
		if err != nil {
			return nil, err
		}
		if !identity.Allowed(resolution.Identity, allowed) {
			return nil, fmt.Errorf("source not allowed for %s (identity %s)", alias, resolution.Identity)
		}
		repoDir := draftRepoDir(home, projectRoot, alias)
		if err := fetchDraftRepoAllowingFallback(repoDir, resolution, alias, stderr); err != nil {
			return nil, err
		}
		roots[alias] = repoDir
	}
	return roots, nil
}

// fetchDraftRepoAllowingFallback brings repoDir onto the resolved
// endpoints: a fresh clone when no tree exists, otherwise a fetch of the
// existing tree. Clone and fetch run user-configuration-isolated
// (repository-transport: user configuration never consulted).
// A positively classified availability/authentication
// failure advances to the next listed endpoint only under the
// availability-auth fallback; every other failure — and exhaustion —
// fails closed as repository_endpoint_unavailable with the sanitized
// per-attempt classes and operator remediation. Raw tool output feeds
// classification only: warnings and errors carry the closed class
// vocabulary, never URLs, stderr, or secrets, and nothing here enters
// portable state.
func fetchDraftRepoAllowingFallback(repoDir string, resolution config.Resolution, alias string, stderr io.Writer) error {
	attempts := resolution.Attempts
	// Alias-substituted endpoints connect to the resolved host:port,
	// not the listed URL (repository-transport §5). Every connection
	// target is derived before any clone runs: a mistranslated attempt
	// fails closed with no network traffic.
	targets := make([]string, len(attempts))
	for index, attempt := range attempts {
		target, ok := attempt.ConnectionURL()
		if !ok {
			return fmt.Errorf("%s: %s: endpoint %d connection target is malformed", config.CodeRepositoryPolicyInvalid, alias, index+1)
		}
		targets[index] = target
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		classes := make([]buildrepo.FailureClass, 0, len(attempts))
		for index, target := range targets {
			err := gitops.FetchURLIsolated(repoDir, target)
			if err == nil {
				return nil
			}
			class := buildrepo.ClassifyFetchOutput(gitFailureDetail(err))
			classes = append(classes, class)
			_, _ = fmt.Fprintf(stderr, "warning: %s: %s\n", alias, draftAttemptClause(index+1, attempts[index], class))
			if index+1 < len(attempts) && buildrepo.AllowSecondAttempt(resolution.Fallback, class) {
				continue
			}
			break
		}
		return fmt.Errorf("%s: %s: %s", config.CodeRepositoryEndpointUnavailable, alias, endpointExhaustion(resolution, classes))
	}
	classes := make([]buildrepo.FailureClass, 0, len(attempts))
	for index, attempt := range attempts {
		cloneErr := gitops.CloneIsolated(targets[index], repoDir)
		if cloneErr == nil {
			return nil
		}
		class := buildrepo.ClassifyFetchOutput(stripCloneFraming(gitFailureDetail(cloneErr)))
		classes = append(classes, class)
		_, _ = fmt.Fprintf(stderr, "warning: %s: %s\n", alias, draftAttemptClause(index+1, attempt, class))
		if index+1 < len(attempts) && !buildrepo.AllowSecondAttempt(resolution.Fallback, class) {
			break
		}
	}
	return fmt.Errorf("%s: %s: %s", config.CodeRepositoryEndpointUnavailable, alias, endpointExhaustion(resolution, classes))
}

// cloneFramingPattern matches git's fixed clone progress line. git prints
// it to stderr even when stderr is not a terminal; it names the local
// destination directory only and carries no failure evidence.
var cloneFramingPattern = regexp.MustCompile(`^Cloning into '[^']+'\.\.\.$`)

// stripCloneFraming drops git's fixed clone progress line from clone
// stderr before classification. Without this every clone failure
// classifies unknown (the closed table matches no progress line) and
// the availability-auth fallback never advances past a DNS or
// authentication failure on a fresh clone. Only exact matches are
// dropped; every other line — including a near-match — is kept, so
// classification stays fail-closed.
func stripCloneFraming(detail string) string {
	lines := strings.Split(detail, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if cloneFramingPattern.MatchString(strings.TrimSpace(line)) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// gitFailureDetail extracts the inner git tool stderr from a gitops error
// ("git <args> failed: <detail>", possibly wrapped once by clone). Text
// without that marker is not git output and classifies as unknown.
func gitFailureDetail(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if index := strings.Index(message, " failed: "); index >= 0 {
		return message[index+len(" failed: "):]
	}
	return message
}

// endpointExhaustion renders the sanitized clone-attempt failure summary
// in the revision-2 exhaustion shape: the portable identity, one closed
// clause per planned endpoint, and operator remediation. classes[i]
// describes attempts[i]; endpoints the fallback gate never reached are
// reported as not attempted. URLs and raw remote output never appear.
func endpointExhaustion(resolution config.Resolution, classes []buildrepo.FailureClass) string {
	attempts := resolution.Attempts
	clauses := make([]string, 0, len(attempts))
	for index, attempt := range attempts {
		if index < len(classes) {
			clauses = append(clauses, draftAttemptClause(index+1, attempt, classes[index]))
			continue
		}
		clauses = append(clauses, "endpoint "+strconv.Itoa(index+1)+": not attempted")
	}
	if len(clauses) == 0 {
		clauses = append(clauses, "endpoint 1 ("+string(buildrepo.FailureUnknown)+"): "+buildrepo.FailureUnknown.Reason())
	}
	return fmt.Sprintf("identity %s: %s; %s", resolution.Identity, strings.Join(clauses, "; "), draftEndpointRemediation)
}

// fetchExhaustion renders the sanitized failure of a fetch against the
// existing checkout: the portable identity, the closed class of the
// fetch, and operator remediation. The checkout's origin is not one of
// the planned attempts, so it is never attributed to an endpoint number.
func fetchExhaustion(resolution config.Resolution, class buildrepo.FailureClass) string {
	return fmt.Sprintf("identity %s: fetch of the existing checkout failed (%s: %s); %s",
		resolution.Identity, class, class.Reason(), draftEndpointRemediation)
}
