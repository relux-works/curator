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
	"sort"
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
	draft := install.DraftSourcesEnabled(os.Getenv)
	var projectManifest *manifest.Manifest
	if draft {
		projectManifest, err = manifest.ParseBytesWithOptions(payload, manifest.PathIn(target.Root), manifest.ParseOptions{DraftSourcesV1: true})
	} else {
		projectManifest, err = manifest.ParseBytes(payload, manifest.PathIn(target.Root))
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	if projectManifest.SchemaVersion != 2 {
		// Frozen v1: read-only path report, exactly as before.
		_, _ = fmt.Fprintf(c.stdout, "alias: %s\npath: %s\nskillfile: %s\n", target.Alias, target.Root, filepath.Join(target.Root, manifest.Name))
		_, _ = fmt.Fprintf(c.stdout, "skills: %s\nbin: %s\n", filepath.Join(target.Root, ".agents", "skills"), filepath.Join(target.Root, ".agents", "bin"))
		return exitOK
	}
	if !draft {
		_, _ = fmt.Fprintf(c.stderr, "curator: source_selection_invalid: Skillfile schema 2 requires the draft lane; unset %s keeps frozen v1\n", install.EnvDraftSourcesV1)
		return exitFail
	}
	plan, lockPath, bindingsPath, err := c.resolveDraftPlan(cfg, target, projectManifest, payload)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
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
// existing tree. A positively classified availability/authentication
// failure advances to the next listed endpoint only under the
// availability-auth fallback; every other failure — and exhaustion —
// fails closed as repository_endpoint_unavailable with the sanitized
// per-attempt classes. The raw tool output is reported to the operator
// on stderr but never enters portable state.
func fetchDraftRepoAllowingFallback(repoDir string, resolution config.Resolution, alias string, stderr io.Writer) error {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		if err := gitops.Fetch(repoDir); err != nil {
			class := buildrepo.ClassifyFetchOutput(gitFailureDetail(err))
			if buildrepo.AllowSecondAttempt(resolution.Fallback, class) && len(resolution.Attempts) > 1 {
				_, _ = fmt.Fprintf(stderr, "warning: %s: fetch failed (%s); recloning from the alternate endpoint\n", alias, class)
				_ = os.RemoveAll(repoDir)
			} else {
				_, _ = fmt.Fprintf(stderr, "warning: %s: %v\n", alias, err)
				return fmt.Errorf("%s: %s: %s", config.CodeRepositoryEndpointUnavailable, alias, endpointExhaustion(resolution, []buildrepo.FailureClass{class}))
			}
		} else {
			return nil
		}
	}
	attempts := resolution.Attempts
	classes := make([]buildrepo.FailureClass, 0, len(attempts))
	for index, attempt := range attempts {
		cloneErr := gitops.Clone(attempt.URL, repoDir)
		if cloneErr == nil {
			return nil
		}
		class := buildrepo.ClassifyFetchOutput(gitFailureDetail(cloneErr))
		classes = append(classes, class)
		_, _ = fmt.Fprintf(stderr, "warning: %s: %v\n", alias, cloneErr)
		if index+1 < len(attempts) && !buildrepo.AllowSecondAttempt(resolution.Fallback, class) {
			break
		}
	}
	return fmt.Errorf("%s: %s: %s", config.CodeRepositoryEndpointUnavailable, alias, endpointExhaustion(resolution, classes))
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

// endpointExhaustion renders the sanitized per-attempt failure summary:
// one closed class reason per attempt, never raw remote output.
func endpointExhaustion(resolution config.Resolution, classes []buildrepo.FailureClass) string {
	reasons := make([]string, 0, len(classes))
	for _, class := range classes {
		reasons = append(reasons, string(class))
	}
	if len(reasons) == 0 {
		reasons = append(reasons, string(buildrepo.FailureUnknown))
	}
	return fmt.Sprintf("identity %s: %s", resolution.Identity, strings.Join(reasons, ", "))
}
