package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/globalbins"
)

// These tests drive the production §11 lookup: resolveProvider directly
// for the verdict matrix, and run() end to end for the shipped
// revision-A dispatch behavior.

// TestActiveRevisionIsWarningRelease pins the rollout switch: this
// release ships the warning revision. The flip to B is a later release
// whose diff must touch this line deliberately.
func TestActiveRevisionIsWarningRelease(t *testing.T) {
	t.Parallel()
	if activeProviderRevision != providerRevisionA {
		t.Fatalf("activeProviderRevision = %d, want the warning revision A", activeProviderRevision)
	}
}

// writeProviderFile materializes one curator-<name> candidate: an
// executable regular file, or a present-but-skipped non-executable.
// On Windows an executable carries the .exe suffix the platform
// lookup requires and a non-executable keeps its bare name.
func writeProviderFile(t *testing.T, dir, name string, executable bool) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(dir, name)
	if runtime.GOOS == "windows" && executable {
		full += ".exe"
	}
	mode := os.FileMode(0o644)
	if executable {
		mode = 0o755
	}
	if err := os.WriteFile(full, []byte("#!/bin/sh\nexit 0\n"), mode); err != nil {
		t.Fatal(err)
	}
	return full
}

// umbrellaInputs builds lookup inputs over temp dirs with the host
// platform and the default executable extensions.
func umbrellaInputs(installDir string, providerDirs, pathEntries []string, revision providerRevision) providerInputs {
	return providerInputs{
		installDir:    installDir,
		providerDirs:  providerDirs,
		publishedDirs: []labeledDir{{path: filepath.Join(installDir, "..", "published"), label: "a manager-published directory"}},
		managedRoots:  []labeledDir{{path: filepath.Join(installDir, "..", "managed"), label: "a managed directory"}},
		pathEntries:   pathEntries,
		platform:      runtime.GOOS,
		pathExt:       append([]string{}, defaultWindowsPathExt...),
		revision:      revision,
	}
}

// TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd is the AC's hostile
// case through the shipped dispatch: a curator-run planted through a
// PATH entry outside the trust roots still runs under revision A, but
// warns with the resolved path, the roots consulted, and the
// provider_directories migration hint.
func TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no bash on this platform")
	}
	source, _ := profileHome(t)
	plant := t.TempDir()
	if err := os.WriteFile(filepath.Join(plant, "curator-run"), []byte("#!/bin/sh\necho planted\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", plant)
	code, stdout, stderr := runProfile(t, source, "run", "alpha")
	if code != exitOK {
		t.Fatalf("run = %d, want dispatch under revision A\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "planted") {
		t.Fatalf("the planted provider did not run:\n%s", stdout)
	}
	for _, want := range []string{
		"subcommand_provider_outside_trust_roots",
		filepath.Join(plant, "curator-run"),
		"trust roots consulted",
		"provider_directories",
		plant,
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr misses %q:\n%s", want, stderr)
		}
	}
}

// TestUmbrellaHostilePathPlantRefusedUnderB is the same plant under
// the enforcing revision: the PATH-only provider is refused with its
// path and the roots consulted, and nothing resolves — even when a
// trusted provider exists, the PATH selection never leaks through.
func TestUmbrellaHostilePathPlantRefusedUnderB(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	plant := filepath.Join(tmp, "plant")
	planted := writeProviderFile(t, plant, "curator-run", true)
	in := umbrellaInputs(install, nil, []string{plant}, providerRevisionB)
	if err := os.MkdirAll(install, 0o755); err != nil {
		t.Fatal(err)
	}
	outcome := resolveProvider("run", in)
	if outcome.resolved != "" || outcome.diagnostic != providerDiagnosticUntrusted {
		t.Fatalf("outcome = %+v, want an untrusted refusal with no selection", outcome)
	}
	if outcome.untrustedPath != planted {
		t.Fatalf("untrusted path = %q, want %q", outcome.untrustedPath, planted)
	}
	if len(outcome.roots) != 1 || outcome.roots[0] != install {
		t.Fatalf("roots consulted = %q, want [%q]", outcome.roots, install)
	}
	// A trusted provider alongside the plant resolves; the plant is
	// never selected and never named.
	trusted := writeProviderFile(t, install, "curator-run", true)
	outcome = resolveProvider("run", in)
	if outcome.resolved != trusted || outcome.diagnostic != "" {
		t.Fatalf("outcome = %+v, want the silent trusted selection", outcome)
	}
}

// TestUmbrellaTrustRootSelectionBothRevisions pins the selection
// matrix: install dir first, listed order, non-executables skipped,
// never descending, and the revision-A PATH-only semantics.
func TestUmbrellaTrustRootSelectionBothRevisions(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	first := filepath.Join(tmp, "first")
	second := filepath.Join(tmp, "second")
	pathDir := filepath.Join(tmp, "path")
	for _, dir := range []string{install, first, second, pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, revision := range []providerRevision{providerRevisionA, providerRevisionB} {
		in := umbrellaInputs(install, []string{first, second}, []string{pathDir}, revision)
		// Listed order, first match wins; install dir beats the list.
		wantSecond := writeProviderFile(t, second, "curator-run", true)
		if outcome := resolveProvider("run", in); revision == providerRevisionB && outcome.resolved != wantSecond {
			t.Fatalf("B: resolved %q, want %q", outcome.resolved, wantSecond)
		}
		wantInstall := writeProviderFile(t, install, "curator-run", true)
		if outcome := resolveProvider("run", in); revision == providerRevisionB && outcome.resolved != wantInstall {
			t.Fatalf("B: resolved %q, want the install-dir match %q", outcome.resolved, wantInstall)
		}
		// A non-executable file of the same name is skipped.
		writeProviderFile(t, first, "curator-run", false)
		if outcome := resolveProvider("run", in); revision == providerRevisionB && outcome.resolved != wantInstall {
			t.Fatalf("B: a non-executable shadowed the match: %+v", outcome)
		}
		// Search never descends into subdirectories.
		nested := filepath.Join(first, "nested")
		writeProviderFile(t, nested, "curator-deep", true)
		if outcome := resolveProvider("deep", in); outcome.diagnostic != providerDiagnosticMissing {
			t.Fatalf("resolved a nested provider: %+v", outcome)
		}
		if err := os.Remove(wantInstall); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(wantSecond); err != nil {
			t.Fatal(err)
		}
	}
	// Revision A never consults the trust roots for selection: a
	// trust-root-only provider is missing even though the root holds it.
	in := umbrellaInputs(install, []string{first}, []string{pathDir}, providerRevisionA)
	writeProviderFile(t, first, "curator-run", true)
	if outcome := resolveProvider("run", in); outcome.diagnostic != providerDiagnosticMissing || outcome.resolved != "" {
		t.Fatalf("A: trust-root-only provider outcome = %+v, want missing", outcome)
	}
	// A PATH selection directly inside a trust root resolves silently.
	in = umbrellaInputs(install, []string{pathDir}, []string{pathDir}, providerRevisionA)
	silent := writeProviderFile(t, pathDir, "curator-run", true)
	if outcome := resolveProvider("run", in); outcome.resolved != silent || outcome.warned || outcome.diagnostic != "" {
		t.Fatalf("A: trusted PATH selection outcome = %+v, want silent", outcome)
	}
}

// TestUmbrellaRefusedDirectoriesBothRevisions pins the
// ownership/writability gate: a provider inside a manager-published or
// managed directory is refused under both revisions, wherever else the
// path was found — including through a symlink, and including a trust
// root that lists a refused directory.
func TestUmbrellaRefusedDirectoriesBothRevisions(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	published := filepath.Join(tmp, "published")
	managed := filepath.Join(tmp, "managed")
	nested := filepath.Join(managed, "nested")
	plain := filepath.Join(tmp, "plain")
	for _, dir := range []string{install, published, managed, nested, plain} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	refusedPublished := writeProviderFile(t, published, "curator-run", true)
	refusedNested := writeProviderFile(t, nested, "curator-run", true)
	for _, revision := range []providerRevision{providerRevisionA, providerRevisionB} {
		newInputs := func(dirs, path []string) providerInputs {
			return providerInputs{
				installDir:    install,
				providerDirs:  dirs,
				publishedDirs: []labeledDir{{path: published, label: "a manager-published directory"}},
				managedRoots:  []labeledDir{{path: managed, label: "a managed directory"}},
				pathEntries:   path,
				platform:      runtime.GOOS,
				pathExt:       append([]string{}, defaultWindowsPathExt...),
				revision:      revision,
			}
		}
		// A PATH match inside a published directory is refused.
		if outcome := resolveProvider("run", newInputs(nil, []string{published})); outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != refusedPublished {
			t.Fatalf("revision %d: published PATH match outcome = %+v", revision, outcome)
		}
		// Below a managed root counts too.
		if outcome := resolveProvider("run", newInputs(nil, []string{nested})); outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != refusedNested {
			t.Fatalf("revision %d: managed PATH match outcome = %+v", revision, outcome)
		}
		// A trust root that lists a refused directory still refuses.
		if outcome := resolveProvider("run", newInputs([]string{published, plain}, []string{plain})); revision == providerRevisionB &&
			(outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != refusedPublished) {
			t.Fatalf("B: listed-but-published outcome = %+v", outcome)
		}
	}
}

// TestUmbrellaSymlinkIntoManagedRefused proves the refusal resolves
// through the link: a provider reached via a shim symlink into a
// managed directory is still a managed directory.
func TestUmbrellaSymlinkIntoManagedRefused(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	managed := filepath.Join(tmp, "managed")
	linkDir := filepath.Join(tmp, "links")
	for _, dir := range []string{install, managed, linkDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	target := writeProviderFile(t, managed, "curator-run", true)
	link := filepath.Join(linkDir, "curator-run")
	if runtime.GOOS == "windows" {
		link += ".exe"
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skip("creating Windows symlink requires host support")
	}
	for _, revision := range []providerRevision{providerRevisionA, providerRevisionB} {
		in := providerInputs{
			installDir:    install,
			publishedDirs: nil,
			managedRoots:  []labeledDir{{path: managed, label: "a managed directory"}},
			pathEntries:   []string{linkDir},
			platform:      runtime.GOOS,
			pathExt:       append([]string{}, defaultWindowsPathExt...),
			revision:      revision,
		}
		if outcome := resolveProvider("run", in); outcome.diagnostic != providerDiagnosticUntrusted || outcome.resolved != "" {
			t.Fatalf("revision %d: symlinked provider outcome = %+v, want a refusal", revision, outcome)
		}
	}
}

// TestUmbrellaUnreadableRootNeverAbsence pins §8.4 at the lookup: a
// trust root the manager cannot read — including one that does not
// exist — fails with the first unreadable root named, never falls
// through to a later root, to PATH, or to missing, and never resolves.
func TestUmbrellaUnreadableRootNeverAbsence(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	gone := filepath.Join(tmp, "gone")
	pathDir := filepath.Join(tmp, "path")
	for _, dir := range []string{install, pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A PATH provider that revision A would warn on still fails: the
	// unreadable root outranks selection, warning, and absence alike.
	writeProviderFile(t, pathDir, "curator-run", true)
	for _, revision := range []providerRevision{providerRevisionA, providerRevisionB} {
		in := umbrellaInputs(install, []string{gone}, []string{pathDir}, revision)
		outcome := resolveProvider("run", in)
		if outcome.diagnostic != providerDiagnosticRootUnreadable || outcome.resolved != "" {
			t.Fatalf("revision %d: outcome = %+v, want root-unreadable", revision, outcome)
		}
		if outcome.unreadableDir != gone {
			t.Fatalf("revision %d: unreadable dir = %q, want %q", revision, outcome.unreadableDir, gone)
		}
		// An unreadable install directory fails the same way.
		in.installDir = filepath.Join(tmp, "no-install-dir")
		outcome = resolveProvider("run", in)
		if outcome.diagnostic != providerDiagnosticRootUnreadable || outcome.unreadableDir != in.installDir {
			t.Fatalf("revision %d: install outcome = %+v", revision, outcome)
		}
	}
	// Under revision B an earlier trusted match wins before the scan
	// reaches a later unreadable root; nothing after the match fires.
	in := umbrellaInputs(install, []string{gone}, []string{pathDir}, providerRevisionB)
	trusted := writeProviderFile(t, install, "curator-run", true)
	if outcome := resolveProvider("run", in); outcome.resolved != trusted || outcome.diagnostic != "" {
		t.Fatalf("B: early match outcome = %+v, want the silent selection", outcome)
	}
}

// TestUmbrellaMissingNamesRoots pins the missing refusal: the exact
// executable name, the installation guidance with no implicit
// install, and the trust roots consulted.
func TestUmbrellaMissingNamesRoots(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	listed := filepath.Join(tmp, "listed")
	empty := filepath.Join(tmp, "empty")
	for _, dir := range []string{install, listed, empty} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, revision := range []providerRevision{providerRevisionA, providerRevisionB} {
		in := umbrellaInputs(install, []string{listed}, []string{empty}, revision)
		outcome := resolveProvider("frobnicate", in)
		if outcome.diagnostic != providerDiagnosticMissing || outcome.resolved != "" {
			t.Fatalf("revision %d: outcome = %+v, want missing", revision, outcome)
		}
		_, _, err := findProvider("frobnicate", in)
		if err == nil {
			t.Fatalf("revision %d: missing provider returned no error", revision)
		}
		for _, want := range []string{"subcommand_provider_missing", "curator-frobnicate", install, listed, "nothing is downloaded or installed implicitly"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("revision %d: error misses %q: %v", revision, want, err)
			}
		}
	}
}

// TestUmbrellaWarningText pins the revision-A warning contents: the
// resolved path, the roots consulted, and the migration hint naming
// the provider's directory.
func TestUmbrellaWarningText(t *testing.T) {
	tmp := t.TempDir()
	install := filepath.Join(tmp, "install")
	outside := filepath.Join(tmp, "outside")
	for _, dir := range []string{install, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	selected := writeProviderFile(t, outside, "curator-run", true)
	in := umbrellaInputs(install, nil, []string{outside}, providerRevisionA)
	path, warning, err := findProvider("run", in)
	if err != nil || path != selected {
		t.Fatalf("path = %q err = %v, want the selection", path, err)
	}
	for _, want := range []string{"subcommand_provider_outside_trust_roots", selected, install, "provider_directories", outside} {
		if !strings.Contains(warning, want) {
			t.Fatalf("warning misses %q: %q", want, warning)
		}
	}
}

// TestProviderPostureRows pins the §12 rows: every discovered provider
// plus always run and session, with the resolved path, the verdict,
// and the currency the spec assigns — warned rows stay current,
// refused, missing, and unreadable rows are non-current.
func TestProviderPostureRows(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	install := filepath.Join(tmp, "install")
	pathDir := filepath.Join(tmp, "path")
	for _, dir := range []string{install, pathDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	selected := writeProviderFile(t, pathDir, "curator-run", true)
	// A name outside the identifier grammar is not a provider.
	if err := os.WriteFile(filepath.Join(pathDir, "curator-!bad"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	in := providerInputs{
		installDir:    install,
		providerDirs:  nil,
		publishedDirs: nil,
		managedRoots:  nil,
		pathEntries:   []string{pathDir},
		platform:      runtime.GOOS,
		pathExt:       append([]string{}, defaultWindowsPathExt...),
		revision:      providerRevisionA,
	}
	rows := providerPostureFor(in)
	byName := map[string]envprofile.ProviderState{}
	for _, row := range rows {
		byName[row.Name] = row
	}
	if _, ok := byName["!bad"]; ok {
		t.Fatalf("a non-identifier name was discovered: %+v", rows)
	}
	run, ok := byName["run"]
	if !ok {
		t.Fatalf("no run row: %+v", rows)
	}
	if run.Verdict != envprofile.ProviderOutsideTrustRoots || !run.Current {
		t.Fatalf("run row = %+v, want a current warning", run)
	}
	if run.Resolved == nil || *run.Resolved != selected {
		t.Fatalf("run resolved = %+v, want %q", run.Resolved, selected)
	}
	session, ok := byName["session"]
	if !ok {
		t.Fatalf("session is always reported: %+v", rows)
	}
	if session.Verdict != envprofile.ProviderMissing || session.Current {
		t.Fatalf("session row = %+v, want a non-current missing row", session)
	}
	if session.Diagnostic == nil || *session.Diagnostic != providerDiagnosticMissing {
		t.Fatalf("session diagnostic = %+v", session.Diagnostic)
	}
	// An unreadable trust root fails every row with the directory.
	in.providerDirs = []string{filepath.Join(tmp, "gone")}
	rows = providerPostureFor(in)
	if len(rows) == 0 {
		t.Fatal("no rows under an unreadable root")
	}
	for _, row := range rows {
		if row.Verdict != envprofile.ProviderUnreadable || row.Current {
			t.Fatalf("row = %+v, want a non-current unreadable row", row)
		}
		if row.UnreadableDirectory == nil || *row.UnreadableDirectory != filepath.Join(tmp, "gone") {
			t.Fatalf("row unreadable dir = %+v", row.UnreadableDirectory)
		}
	}
}

// TestProviderSubcommandParsing pins provider discovery names: the
// curator- prefix with an identifier spelling, and the Windows
// executable extension stripped before validation.
func TestProviderSubcommandParsing(t *testing.T) {
	unix := providerInputs{platform: "linux", pathExt: append([]string{}, defaultWindowsPathExt...)}
	windows := providerInputs{platform: "windows", pathExt: append([]string{}, defaultWindowsPathExt...)}
	cases := []struct {
		entry string
		in    providerInputs
		name  string
		ok    bool
	}{
		{"curator-run", unix, "run", true},
		{"curator-session", unix, "session", true},
		{"curator-run.exe", windows, "run", true},
		{"curator-run.EXE", windows, "run", true},
		{"curator-run.bat", windows, "run", true},
		{"curator-", unix, "", false},
		{"curator-!bad", unix, "", false},
		{"curator-run", windows, "", false},
		{"run", unix, "", false},
		{"notcurator-run", unix, "", false},
	}
	for _, tc := range cases {
		if name, ok := providerSubcommand(tc.entry, tc.in); ok != tc.ok || name != tc.name {
			t.Errorf("providerSubcommand(%q, %s) = (%q, %v), want (%q, %v)",
				tc.entry, tc.in.platform, name, ok, tc.name, tc.ok)
		}
	}
}

// TestExecutableCandidates pins the Windows search order: the bare
// name never matches there, and the PATHEXT extensions apply in
// order. Unix searches the bare name alone.
func TestExecutableCandidates(t *testing.T) {
	unix := providerInputs{platform: "linux"}
	if got := executableCandidates("curator-run", unix); len(got) != 1 || got[0] != "curator-run" {
		t.Fatalf("unix candidates = %q", got)
	}
	windows := providerInputs{platform: "windows", pathExt: []string{".exe", ".bat"}}
	if got := executableCandidates("curator-run", windows); len(got) != 2 || got[0] != "curator-run.exe" || got[1] != "curator-run.bat" {
		t.Fatalf("windows candidates = %q", got)
	}
	empty := providerInputs{platform: "windows"}
	if got := executableCandidates("curator-run", empty); len(got) != len(defaultWindowsPathExt) {
		t.Fatalf("empty PATHEXT candidates = %q, want the platform default", got)
	}
}

// hostBinFixture stages a provider directory below a fake user home with
// PATH pointing at it, so the user-bin selector picks the provider
// directory from its scan — the Windows hosted-gate shape, where the
// test temp tree sits below the user profile. It returns the manager
// home, the fake user home, and the provider directory.
func hostBinFixture(t *testing.T) (managerHome, fakeHome, providerDir string) {
	t.Helper()
	root := t.TempDir()
	managerHome = filepath.Join(root, "manager")
	fakeHome = filepath.Join(root, "home")
	providerDir = filepath.Join(fakeHome, "providers")
	if err := os.MkdirAll(providerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)
	t.Setenv(globalbins.UserBinEnv, "")
	return managerHome, fakeHome, providerDir
}

// TestProviderInputsForHostScannedBinWarnsWithoutLedger is the Windows
// hosted-gate regression: the stub provider directory is selected from
// the PATH scan, but the manager never published there, so revision A
// warns instead of refusing. A scanned-but-unpublished directory holds
// no manager-written content.
func TestProviderInputsForHostScannedBinWarnsWithoutLedger(t *testing.T) {
	managerHome, fakeHome, providerDir := hostBinFixture(t)
	planted := writeProviderFile(t, providerDir, "curator-run", true)
	t.Setenv("PATH", providerDir)
	selection := globalbins.Select(managerHome, runtime.GOOS, map[string]string{"PATH": providerDir}, fakeHome)
	if selection.Path != providerDir || selection.Explicit {
		t.Fatalf("selection = %+v, want the scanned provider dir — else this test is vacuous", selection)
	}
	in, err := providerInputsForHost(managerHome, nil, nil, providerRevisionA)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range in.publishedDirs {
		if sameDir(filepath.Dir(planted), entry.path) {
			t.Fatalf("an unpublished scanned dir is refused: %+v", in.publishedDirs)
		}
	}
	outcome := resolveProvider("run", in)
	if outcome.resolved != planted || !outcome.warned || outcome.diagnostic != providerDiagnosticOutsideRoots {
		t.Fatalf("outcome = %+v, want the revision-A outside-roots warning", outcome)
	}
}

// TestProviderInputsForHostPublishedBinRefuses pins the other side of
// the refusal: once the manager publishes shims into the selected
// directory, providers there are refused under revision A. This must
// not flip to a warning when the unpublished case is repaired.
func TestProviderInputsForHostPublishedBinRefuses(t *testing.T) {
	managerHome, fakeHome, providerDir := hostBinFixture(t)
	planted := writeProviderFile(t, providerDir, "curator-run", true)
	platform := "unix"
	if runtime.GOOS == "windows" {
		platform = "windows"
	}
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalBin, "alpha"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	globalbins.Refresh(managerHome, map[string]bool{"alpha": true}, platform, map[string]string{"PATH": providerDir}, fakeHome)
	if !globalbins.PublishedShims(providerDir) {
		t.Fatal("the publish did not establish the ownership ledger")
	}
	t.Setenv("PATH", providerDir)
	in, err := providerInputsForHost(managerHome, nil, nil, providerRevisionA)
	if err != nil {
		t.Fatal(err)
	}
	outcome := resolveProvider("run", in)
	if outcome.resolved != "" || outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != planted {
		t.Fatalf("outcome = %+v, want an untrusted refusal of the published dir", outcome)
	}
}

// TestProviderInputsForHostExplicitBinRefuses pins the operator-declared
// publishing location: an explicitly configured user-bin directory is
// refused even before the first publish. The host glue must carry the
// explicit variable to the selector — ignoring it lets a declared shim
// directory silently degrade to whatever the scan finds first.
func TestProviderInputsForHostExplicitBinRefuses(t *testing.T) {
	managerHome, fakeHome, providerDir := hostBinFixture(t)
	planted := writeProviderFile(t, providerDir, "curator-run", true)
	preferred := filepath.Join(fakeHome, ".local", "bin")
	if err := os.MkdirAll(preferred, 0o755); err != nil {
		t.Fatal(err)
	}
	pathValue := preferred + string(os.PathListSeparator) + providerDir
	t.Setenv("PATH", pathValue)
	t.Setenv(globalbins.UserBinEnv, providerDir)
	selection := globalbins.Select(managerHome, runtime.GOOS, map[string]string{
		"PATH":                pathValue,
		globalbins.UserBinEnv: providerDir,
	}, fakeHome)
	if selection.Path != providerDir || !selection.Explicit {
		t.Fatalf("selection = %+v, want the explicit provider dir", selection)
	}
	in, err := providerInputsForHost(managerHome, nil, nil, providerRevisionA)
	if err != nil {
		t.Fatal(err)
	}
	outcome := resolveProvider("run", in)
	if outcome.resolved != "" || outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != planted {
		t.Fatalf("outcome = %+v, want an untrusted refusal of the explicit shim dir", outcome)
	}
}

// TestUmbrellaTrustRootCaseVariantIdentity pins spelling identity for
// the trust verdict: a trust root naming the provider directory in a
// different case still names the same root, and a managed root in a
// different case still refuses. Volumes that treat the variant as a
// missing path skip: there the two spellings name nothing in common.
func TestUmbrellaTrustRootCaseVariantIdentity(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	install := filepath.Join(tmp, "install")
	root := filepath.Join(tmp, "CaseRoot")
	managed := filepath.Join(tmp, "Managed")
	for _, dir := range []string{install, root, managed} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	lower := func(dir string) string {
		return filepath.Join(filepath.Dir(dir), strings.ToLower(filepath.Base(dir)))
	}
	for _, variant := range []string{lower(root), lower(managed)} {
		if _, err := os.Stat(variant); err != nil {
			t.Skipf("test filesystem is case-sensitive: %v", err)
		}
	}
	newInputs := func() providerInputs {
		return providerInputs{
			installDir:    install,
			publishedDirs: nil,
			pathExt:       append([]string{}, defaultWindowsPathExt...),
			platform:      runtime.GOOS,
			revision:      providerRevisionA,
		}
	}
	silent := writeProviderFile(t, root, "curator-run", true)
	in := newInputs()
	in.providerDirs = []string{lower(root)}
	in.pathEntries = []string{root}
	if outcome := resolveProvider("run", in); outcome.resolved != silent || outcome.warned || outcome.diagnostic != "" {
		t.Fatalf("case-variant trust root outcome = %+v, want a silent selection", outcome)
	}
	refused := writeProviderFile(t, managed, "curator-run", true)
	in = newInputs()
	in.managedRoots = []labeledDir{{path: lower(managed), label: "a managed directory"}}
	in.pathEntries = []string{managed}
	if outcome := resolveProvider("run", in); outcome.resolved != "" || outcome.diagnostic != providerDiagnosticUntrusted || outcome.untrustedPath != refused {
		t.Fatalf("case-variant managed root outcome = %+v, want an untrusted refusal", outcome)
	}
}
