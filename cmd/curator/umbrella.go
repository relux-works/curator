package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/install"
)

// Umbrella subcommand discovery (environments §11): a CLI subcommand the
// manager does not implement resolves to an executable named curator-<name>
// and is executed with the remaining arguments verbatim — the git,
// kubectl, and docker external-subcommand convention, closed to the search
// each rollout revision names — except that the run row rewrites a leading
// claude/codex alias to its canonical environment id first (manager
// profile §12.1). The manager carries no knowledge of any provider: no
// provider registry, no provider-specific flags, no version coupling. The
// first providers are curator-run (the launcher, its own specification)
// and curator-session (a shim to the agent session manager); neither
// ships here.
//
// The provider trust roots, in search order, are the install directory —
// the directory holding the running manager executable, resolved after
// symlinks — then the machine-configuration provider_directories list.

// Closed §11.1 diagnostics for umbrella provider lookup.
const (
	providerDiagnosticMissing        = "subcommand_provider_missing"
	providerDiagnosticUntrusted      = "subcommand_provider_untrusted"
	providerDiagnosticOutsideRoots   = "subcommand_provider_outside_trust_roots"
	providerDiagnosticRootUnreadable = "subcommand_provider_root_unreadable"
)

// providerRevision selects the §11 rollout profile.
type providerRevision int

const (
	// providerRevisionA is the warning release: the ambient PATH still
	// selects, and a selection outside the trust roots warns.
	providerRevisionA providerRevision = iota + 1
	// providerRevisionB is the flip release: the trust roots select, the
	// ambient PATH never does, and a PATH-only match is refused.
	providerRevisionB
)

// activeProviderRevision is the ONE rollout switch: the revision this
// release ships. Both revisions are implemented so the conformance
// vectors for both run; the flip to B is a later release, never this
// change.
const activeProviderRevision = providerRevisionA

// defaultWindowsPathExt is the PATHEXT fallback Go's own lookup uses
// when the variable is empty.
var defaultWindowsPathExt = []string{".com", ".exe", ".bat", ".cmd"}

// labeledDir is a refused directory with the human label the
// subcommand_provider_untrusted refusal names.
type labeledDir struct {
	path  string
	label string
}

// providerInputs drives resolveProvider: the trust roots, the refused
// directories, and the ambient search domain. Production builds it from
// the host; the conformance vectors build it from the case.
type providerInputs struct {
	installDir    string
	providerDirs  []string
	publishedDirs []labeledDir
	managedRoots  []labeledDir
	pathEntries   []string
	platform      string
	pathExt       []string
	revision      providerRevision
}

// providerOutcome is one §11 lookup. Resolved names the absolute
// provider path when the revision selects it — silently, or with the
// revision-A warning when warned is set. Otherwise diagnostic names the
// closed failure and resolved is empty; the five outcomes are disjoint.
type providerOutcome struct {
	resolved      string
	warned        bool
	diagnostic    string
	roots         []string
	revision      providerRevision
	untrustedPath string
	refusedLabel  string
	unreadableDir string
	unreadableErr error
	hintDir       string
}

// trustRootsOf returns the §11 search order: the install directory,
// then provider_directories in listed order.
func trustRootsOf(installDir string, providerDirs []string) []string {
	roots := make([]string, 0, len(providerDirs)+1)
	roots = append(roots, installDir)
	roots = append(roots, providerDirs...)
	return roots
}

// resolveProvider runs one §11 lookup for curator-<name>. Profile,
// marker, and fragment data never influence the dispatched name, the
// resolved path, or the argument vector: dispatch input is operator argv,
// the trust roots, and machine configuration alone.
func resolveProvider(name string, in providerInputs) providerOutcome {
	exe := "curator-" + name
	roots := trustRootsOf(in.installDir, in.providerDirs)
	if in.revision == providerRevisionB {
		return resolveRevisionB(exe, in, roots)
	}
	return resolveRevisionA(exe, in, roots)
}

// resolveRevisionA is exactly the pre-change behavior with the trust
// verdict: the ambient PATH selects, a PATH match inside a
// manager-published or managed directory is refused, an unreadable trust
// root fails instead of resolving, a PATH match inside a trust root
// resolves silently, a PATH match outside the trust roots resolves with
// the migration warning, and no PATH match is missing — even when a
// trust root holds the provider, which revision A never consults for
// selection.
func resolveRevisionA(exe string, in providerInputs, roots []string) providerOutcome {
	if dir, err := firstUnreadableRoot(roots); err != nil {
		return providerOutcome{diagnostic: providerDiagnosticRootUnreadable, roots: roots, revision: providerRevisionA, unreadableDir: dir, unreadableErr: err}
	}
	match, found := searchOrderedDirs(exe, in.pathEntries, in)
	if !found {
		return providerOutcome{diagnostic: providerDiagnosticMissing, roots: roots, revision: providerRevisionA}
	}
	absolute := mustAbsolute(match)
	canonical := canonicalPath(absolute)
	if refused, label := refuseDir(canonical, in.publishedDirs, in.managedRoots); refused {
		return providerOutcome{diagnostic: providerDiagnosticUntrusted, roots: roots, revision: providerRevisionA, untrustedPath: absolute, refusedLabel: label}
	}
	if insideTrustRoots(canonical, roots) {
		return providerOutcome{resolved: absolute, roots: roots, revision: providerRevisionA}
	}
	return providerOutcome{resolved: absolute, warned: true, diagnostic: providerDiagnosticOutsideRoots, roots: roots, revision: providerRevisionA, hintDir: filepath.Dir(absolute)}
}

// resolveRevisionB searches the trust roots in order for the first
// executable regular file directly inside a root, then performs the
// diagnostic-only PATH probe. The five outcomes are mutually exclusive:
// a trusted match resolves silently; a match inside a
// manager-published or managed directory is refused; a probe match is
// refused; neither holding the provider is missing; and an unreadable
// trust root fails with no later root, probe, or absence outcome firing.
func resolveRevisionB(exe string, in providerInputs, roots []string) providerOutcome {
	for _, root := range roots {
		if err := checkRootReadable(root); err != nil {
			return providerOutcome{diagnostic: providerDiagnosticRootUnreadable, roots: roots, revision: providerRevisionB, unreadableDir: root, unreadableErr: err}
		}
		match, found := searchOneDir(exe, root, in)
		if !found {
			continue
		}
		absolute := mustAbsolute(match)
		canonical := canonicalPath(absolute)
		if refused, label := refuseDir(canonical, in.publishedDirs, in.managedRoots); refused {
			return providerOutcome{diagnostic: providerDiagnosticUntrusted, roots: roots, revision: providerRevisionB, untrustedPath: absolute, refusedLabel: label}
		}
		return providerOutcome{resolved: absolute, roots: roots, revision: providerRevisionB}
	}
	if match, found := searchOrderedDirs(exe, in.pathEntries, in); found {
		absolute := mustAbsolute(match)
		canonical := canonicalPath(absolute)
		_, label := refuseDir(canonical, in.publishedDirs, in.managedRoots)
		return providerOutcome{diagnostic: providerDiagnosticUntrusted, roots: roots, revision: providerRevisionB, untrustedPath: absolute, refusedLabel: label}
	}
	return providerOutcome{diagnostic: providerDiagnosticMissing, roots: roots, revision: providerRevisionB}
}

// searchOrderedDirs searches directories in order for the first
// executable regular file named exe directly inside one, skipping
// non-executables and never descending. An empty entry means the
// working directory on unix (shell semantics) and is skipped on
// Windows, matching the platform lookup.
func searchOrderedDirs(exe string, dirs []string, in providerInputs) (string, bool) {
	for _, dir := range dirs {
		if dir == "" {
			if in.platform == "windows" {
				continue
			}
			dir = "."
		}
		if match, found := searchOneDir(exe, dir, in); found {
			return match, true
		}
	}
	return "", false
}

// searchOneDir names the executable directly inside dir: the bare name
// on unix, the PATHEXT candidates in order on Windows.
func searchOneDir(exe, dir string, in providerInputs) (string, bool) {
	for _, candidate := range executableCandidates(exe, in) {
		full := filepath.Join(dir, candidate)
		if isExecutableFile(full, in.platform) {
			return full, true
		}
	}
	return "", false
}

// executableCandidates lists the filenames a provider may carry: the
// bare executable name on unix, the name plus each PATHEXT extension in
// order on Windows — where an extensionless file is not executable and
// never matches, exactly as the platform lookup behaves.
func executableCandidates(exe string, in providerInputs) []string {
	if in.platform != "windows" {
		return []string{exe}
	}
	exts := in.pathExt
	if len(exts) == 0 {
		exts = defaultWindowsPathExt
	}
	out := make([]string, 0, len(exts))
	for _, ext := range exts {
		out = append(out, exe+ext)
	}
	return out
}

// isExecutableFile reports whether path is an executable regular file:
// a regular file with an execute bit on unix, a regular file with an
// executable extension on Windows (the caller supplies the extension,
// so existence as a regular file decides).
func isExecutableFile(path, platform string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	if platform == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

// firstUnreadableRoot names the first trust root in search order the
// manager cannot read. Any read failure — including a root that does
// not exist — is unreadable, never absence (§8.4): the caller fails
// instead of falling through to a later root, to PATH, or to missing.
func firstUnreadableRoot(roots []string) (string, error) {
	for _, root := range roots {
		if err := checkRootReadable(root); err != nil {
			return root, err
		}
	}
	return "", nil
}

func checkRootReadable(root string) error {
	_, err := os.ReadDir(root)
	return err
}

// refuseDir reports whether the canonical candidate path lives in a
// directory the manager itself publishes onto PATH — directly inside
// the user-bin shim directory or a managed skill bin directory — or at
// or below a manager-owned root, so profile-materialized files cannot
// poison the dispatch the manager trusts. Both sides resolve through
// symlinks: a provider reached via a shim symlink into a refused
// directory is still refused.
func refuseDir(canonical string, published []labeledDir, managed []labeledDir) (bool, string) {
	dir := filepath.Dir(canonical)
	for _, entry := range published {
		if entry.path == "" {
			continue
		}
		if sameDir(dir, canonicalPath(entry.path)) {
			return true, entry.label
		}
	}
	for _, entry := range managed {
		if entry.path == "" {
			continue
		}
		if underDir(canonical, canonicalPath(entry.path)) {
			return true, entry.label
		}
	}
	return false, ""
}

// insideTrustRoots reports whether the canonical candidate path lies
// directly inside one of the trust roots.
func insideTrustRoots(canonical string, roots []string) bool {
	dir := filepath.Dir(canonical)
	for _, root := range roots {
		if sameDir(dir, canonicalPath(root)) {
			return true
		}
	}
	return false
}

// canonicalPath resolves path for directory comparison: absolute,
// through symlinks when they resolve, cleaned otherwise.
func canonicalPath(path string) string {
	absolute := mustAbsolute(path)
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return resolved
	}
	return filepath.Clean(absolute)
}

func mustAbsolute(path string) string {
	if absolute, err := filepath.Abs(path); err == nil {
		return absolute
	}
	return path
}

// rootsConsulted renders the trust roots the outcome names.
func (o providerOutcome) rootsConsulted() string {
	return strings.Join(o.roots, ", ")
}

// failure renders the closed refusal for a lookup that selected
// nothing: missing, untrusted, or root-unreadable.
func (o providerOutcome) failure(name string) string {
	switch o.diagnostic {
	case providerDiagnosticUntrusted:
		if o.refusedLabel != "" {
			return fmt.Sprintf("%s: curator-%s at %s resolved inside %s (trust roots consulted: %s)",
				o.diagnostic, name, o.untrustedPath, o.refusedLabel, o.rootsConsulted())
		}
		return fmt.Sprintf("%s: curator-%s at %s is outside the trust roots consulted (%s)",
			o.diagnostic, name, o.untrustedPath, o.rootsConsulted())
	case providerDiagnosticRootUnreadable:
		return fmt.Sprintf("%s: trust root %s cannot be read: %v",
			o.diagnostic, o.unreadableDir, o.unreadableErr)
	default:
		domain := "on PATH"
		if o.revision == providerRevisionB {
			domain = "in the trust roots or the PATH probe"
		}
		return fmt.Sprintf("%s: no curator-%s %s (trust roots consulted: %s): install the %s provider; nothing is downloaded or installed implicitly",
			providerDiagnosticMissing, name, domain, o.rootsConsulted(), name)
	}
}

// warningText renders the revision-A outside-trust-roots warning: the
// resolved path, the roots consulted, and the provider_directories
// migration hint.
func (o providerOutcome) warningText(name string) string {
	return fmt.Sprintf("%s: curator-%s resolved to %s outside the trust roots consulted (%s); list %s in provider_directories to trust it",
		providerDiagnosticOutsideRoots, name, o.resolved, o.rootsConsulted(), o.hintDir)
}

// findProvider resolves curator-<name> under the §11 trust-root rule.
// It returns the absolute provider path, the revision-A warning when
// the selection lies outside the trust roots, or the closed refusal.
func findProvider(name string, in providerInputs) (path, warning string, err error) {
	outcome := resolveProvider(name, in)
	if outcome.resolved == "" {
		return "", "", fmt.Errorf("%s", outcome.failure(name))
	}
	if outcome.warned {
		return outcome.resolved, outcome.warningText(name), nil
	}
	return outcome.resolved, "", nil
}

// providerInputsForHost builds the lookup inputs from the running
// host: the install directory of this executable resolved through
// symlinks, the machine provider_directories, the refused directories
// (the user-bin shim directory, the managed skill bin directories, and
// the environments root), and the ambient PATH.
func providerInputsForHost(home string, providerDirs, projectBinDirs []string, revision providerRevision) (providerInputs, error) {
	executable, err := os.Executable()
	if err != nil {
		return providerInputs{}, fmt.Errorf("cannot determine the manager install directory: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	published := []labeledDir{}
	userHome, _ := os.UserHomeDir()
	selection := globalbins.Select(home, runtime.GOOS, pathEnvironment(), userHome)
	if selection.Path != "" && userBinShimRefused(selection) {
		published = append(published, labeledDir{path: selection.Path, label: "the user-bin shim directory"})
	}
	published = append(published, labeledDir{path: filepath.Join(install.GlobalRoot(home), "bin"), label: "a managed skill bin directory"})
	for _, dir := range projectBinDirs {
		published = append(published, labeledDir{path: dir, label: "a managed skill bin directory"})
	}
	return providerInputs{
		installDir:    filepath.Dir(executable),
		providerDirs:  providerDirs,
		publishedDirs: published,
		managedRoots:  []labeledDir{{path: filepath.Join(home, "environments"), label: "the environments root"}},
		pathEntries:   filepath.SplitList(os.Getenv("PATH")),
		platform:      runtime.GOOS,
		pathExt:       windowsPathExt(),
		revision:      revision,
	}, nil
}

// windowsPathExt reads the executable extensions in PATHEXT order,
// lowercased and dot-prefixed as the platform lookup does.
func windowsPathExt() []string {
	raw := os.Getenv("PATHEXT")
	if raw == "" {
		return append([]string{}, defaultWindowsPathExt...)
	}
	var exts []string
	for _, entry := range strings.Split(strings.ToLower(raw), ";") {
		if entry == "" {
			continue
		}
		if !strings.HasPrefix(entry, ".") {
			entry = "." + entry
		}
		exts = append(exts, entry)
	}
	if len(exts) == 0 {
		return append([]string{}, defaultWindowsPathExt...)
	}
	return exts
}

// userBinShimRefused reports whether the selected user-bin directory is
// one the manager itself publishes onto PATH (§11): an explicitly
// configured directory is the declared publishing location, and a
// scanned directory counts once the manager has published shims there
// (the ownership ledger exists). A merely selected directory — the
// first safe PATH entry, which may be any operator directory — holds
// no manager-written content, so providers there warn under revision A
// instead of refusing.
func userBinShimRefused(selection globalbins.Selection) bool {
	return selection.Explicit || globalbins.PublishedShims(selection.Path)
}

// pathEnvironment carries the process PATH and the explicit user-bin
// override to the selector, so an operator-declared publishing
// directory is honored as the refused shim directory.
func pathEnvironment() map[string]string {
	return map[string]string{"PATH": os.Getenv("PATH"), globalbins.UserBinEnv: os.Getenv(globalbins.UserBinEnv)}
}

func sameDir(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	// Filesystem identity covers spellings canonical text cannot see:
	// an 8.3 short name against its long name on Windows, or a case
	// variant on an insensitive volume. A side that does not stat is
	// never the same directory, so inspection failures resolve toward
	// the revision-A warning, never toward silent trust. Missing
	// refused entries are the common case; a refused directory that
	// exists but cannot be read resolves the same way — no closed
	// diagnostic covers that race, and the spec's fail-closed sites
	// stay the explicit trust-root readability checks.
	aInfo, aErr := os.Stat(a)
	bInfo, bErr := os.Stat(b)
	if aErr != nil || bErr != nil {
		return false
	}
	return os.SameFile(aInfo, bInfo)
}

func underDir(path, root string) bool {
	clean := filepath.Clean(root)
	if path == clean || strings.HasPrefix(path, clean+string(filepath.Separator)) {
		return true
	}
	// The SameFile ancestor walk covers spellings the string prefix
	// cannot see, as above. An unreadable root shelters nothing here —
	// the lookup fails closed on unreadable trust roots before any
	// comparison runs.
	rootInfo, err := os.Stat(root)
	if err != nil {
		return false
	}
	for current := path; ; {
		if info, statErr := os.Stat(current); statErr == nil && os.SameFile(rootInfo, info) {
			return true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return false
		}
		current = parent
	}
}

// providerSubcommand extracts the subcommand name from a directory
// entry: curator-<name>, with a Windows executable extension stripped.
// Anything else — including a name outside the identifier grammar — is
// not a provider.
func providerSubcommand(entry string, in providerInputs) (string, bool) {
	rest, found := strings.CutPrefix(entry, "curator-")
	if !found || rest == "" {
		return "", false
	}
	if in.platform == "windows" {
		// An extensionless file is not executable on Windows and the
		// search never discovers it, so discovery skips it too.
		lowered := strings.ToLower(rest)
		stripped := false
		for _, ext := range executableCandidates("", in) {
			if strings.HasSuffix(lowered, ext) {
				rest = rest[:len(rest)-len(ext)]
				stripped = true
				break
			}
		}
		if !stripped || rest == "" {
			return "", false
		}
	}
	if !identifiers.Valid(rest) {
		return "", false
	}
	return rest, true
}

// discoverProviderNames lists the subcommand names with a curator-<name>
// executable in the active revision's search domain — the PATH entries
// under revision A, the trust roots plus the diagnostic-only PATH probe
// domain under revision B — always including run and session, which §12
// reports even when absent. An unreadable directory during discovery is
// skipped here; unreadable trust roots are reported per provider by the
// lookup itself.
func discoverProviderNames(in providerInputs) []string {
	seen := map[string]bool{"run": true, "session": true}
	var dirs []string
	if in.revision == providerRevisionB {
		dirs = append(dirs, trustRootsOf(in.installDir, in.providerDirs)...)
	}
	dirs = append(dirs, in.pathEntries...)
	for _, dir := range dirs {
		if dir == "" {
			if in.platform == "windows" {
				continue
			}
			dir = "."
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if name, ok := providerSubcommand(entry.Name(), in); ok {
				seen[name] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// providerPosture resolves every discovered provider under the active
// revision and reports the resolved absolute path with its trust
// verdict (§12): trusted, the revision-A outside-roots warning, a
// refusal, missing, or an unreadable trust root. A refused or failed
// provider row is non-current; the warning row stays current.
func providerPosture(home string, providerDirs, projectBinDirs []string) []envprofile.ProviderState {
	in, err := providerInputsForHost(home, providerDirs, projectBinDirs, activeProviderRevision)
	if err != nil {
		return nil
	}
	return providerPostureFor(in)
}

// providerPostureFor is the test seam behind providerPosture.
func providerPostureFor(in providerInputs) []envprofile.ProviderState {
	var rows []envprofile.ProviderState
	for _, name := range discoverProviderNames(in) {
		outcome := resolveProvider(name, in)
		row := envprofile.ProviderState{
			Name:       name,
			Executable: "curator-" + name,
			Verdict:    envprofile.ProviderTrusted,
			Current:    true,
			TrustRoots: append([]string{}, outcome.roots...),
		}
		switch {
		case outcome.resolved != "" && !outcome.warned:
			resolved := outcome.resolved
			row.Resolved = &resolved
		case outcome.warned:
			resolved := outcome.resolved
			diagnostic := providerDiagnosticOutsideRoots
			row.Resolved = &resolved
			row.Verdict = envprofile.ProviderOutsideTrustRoots
			row.Diagnostic = &diagnostic
		case outcome.diagnostic == providerDiagnosticRootUnreadable:
			diagnostic := outcome.diagnostic
			unreadable := outcome.unreadableDir
			row.Verdict = envprofile.ProviderUnreadable
			row.Diagnostic = &diagnostic
			row.UnreadableDirectory = &unreadable
			row.Current = false
		case outcome.diagnostic == providerDiagnosticUntrusted:
			diagnostic := outcome.diagnostic
			refused := outcome.untrustedPath
			row.RefusedPath = &refused
			row.Verdict = envprofile.ProviderRefused
			row.Diagnostic = &diagnostic
			row.Current = false
		default:
			diagnostic := providerDiagnosticMissing
			row.Verdict = envprofile.ProviderMissing
			row.Diagnostic = &diagnostic
			row.Current = false
		}
		rows = append(rows, row)
	}
	return rows
}

// normalizeRunEnvOperand rewrites the launcher environment operand of
// `curator run <env-id> ...` from an alias to its canonical id. Only the
// first token after run is an operand position the manager may read: the
// launcher grammar puts <env-id> at the first non-flag token, and the
// manager knows no provider flags, so a later alias — a flag value such
// as a profile literally named claude — passes through untouched for the
// launcher to classify. Unknown spellings pass through as well: the
// launcher owns the refusal. The rewrite is a pure function of operator
// argv, so no profile, marker, fragment, or configuration data influences
// the dispatch.
func normalizeRunEnvOperand(args []string) []string {
	if len(args) < 2 || args[0] != "run" {
		return args
	}
	normalized := envregistry.NormalizeEnvID(args[1])
	if normalized == args[1] {
		return args
	}
	rewritten := append([]string{}, args...)
	rewritten[1] = normalized
	return rewritten
}

// cmdUmbrella executes the resolved provider with the remaining arguments
// verbatim and propagates its exit code. cfg is nil when the machine
// configuration did not load, in which case the lookup trusts the
// install directory alone.
func (c cli) cmdUmbrella(cfg *config.Config, home string, args []string) int {
	args = normalizeRunEnvOperand(args)
	var providerDirs []string
	var projectBins []string
	if cfg != nil {
		providerDirs = cfg.Env.ProviderDirectories
		for _, project := range cfg.Projects {
			projectBins = append(projectBins, filepath.Join(project.Path, ".agents", "bin"))
		}
		sort.Strings(projectBins)
	}
	in, err := providerInputsForHost(home, providerDirs, projectBins, activeProviderRevision)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	path, warning, err := findProvider(args[0], in)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	if warning != "" {
		_, _ = fmt.Fprintln(c.stderr, "warning:", warning)
	}
	command := exec.Command(path, args[1:]...) // #nosec G204,G702 -- §11 dispatches the trust-root-resolved provider with operator argv verbatim
	command.Stdin = os.Stdin
	command.Stdout = c.stdout
	command.Stderr = c.stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
