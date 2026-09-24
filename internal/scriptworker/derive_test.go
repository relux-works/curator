package scriptworker

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/skillspec"
)

func runDerivedSession(t *testing.T, fixture *launchFixture) (Result, stubReport) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := runSession(ctx, fixture.request, nil)
	if err != nil {
		t.Fatalf("runSession refused: %v", err)
	}
	return result, decodeStubReport(t, result.Stdout)
}

// TestCapabilityDerivationAllFieldsAbsentDenyByDefault drives vector case
// `all-fields-absent-deny-by-default` at the production entry: an absent
// (or empty) declaration derives the offline environment, exactly the
// manager-owned PATH farm, the private runtime area with no project root,
// and no secrets — and the inherited environment and PATH are discarded.
func TestCapabilityDerivationAllFieldsAbsentDenyByDefault(t *testing.T) {
	for _, testCase := range []struct {
		name string
		raw  string
	}{
		{"nil-declaration", ""},
		{"empty-object", `{}`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newLaunchFixture(t, os.Args[0])
			if testCase.raw != "" {
				fixture.declareCapabilities(t, testCase.raw)
			} else {
				fixture.request.CapabilitiesRaw = nil
			}
			decoyDir := mustPhysical(t, t.TempDir())
			fixture.request.HostEnvironment = []string{
				"PATH=" + decoyDir,
				"HOME=/host/home",
				"TMPDIR=/host/tmp",
				"HTTP_PROXY=http://host-proxy:8080",
				"LD_PRELOAD=/host/evil.so",
				"CURATOR_UNDECLARED=leaked",
			}
			result, report := runDerivedSession(t, fixture)

			if result.Report.NetworkMode != networkModeOffline {
				t.Fatalf("network mode = %q, want %q", result.Report.NetworkMode, networkModeOffline)
			}
			if len(result.Report.RecordedHosts) != 0 || len(result.Report.WithheldEnv) != 0 ||
				len(result.Report.ResolvedExec) != 0 || len(result.Report.UnresolvedExec) != 0 ||
				len(result.Report.SecretIDs) != 0 {
				t.Fatalf("absent report = %+v, want empty", result.Report)
			}
			// The interpreter environment is exactly the manager set:
			// nothing inherited, nothing declared, nothing extra.
			wantKeys := []string{"PATH", "HOME", "TMPDIR",
				"XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME"}
			if runtime.GOOS == "windows" {
				// SYSTEMROOT and WINDIR are manager-set loader essentials
				// on Windows (never env_read): the builder guarantees them
				// so Go's os/exec never injects an unmanaged SYSTEMROOT
				// past the derived set.
				wantKeys = append(wantKeys, "TEMP", "TMP", "USERPROFILE",
					"APPDATA", "LOCALAPPDATA", "PATHEXT",
					"SYSTEMROOT", "WINDIR")
			}
			var gotKeys []string
			for key := range report.Env {
				gotKeys = append(gotKeys, key)
			}
			sort.Strings(gotKeys)
			sort.Strings(wantKeys)
			if strings.Join(gotKeys, ",") != strings.Join(wantKeys, ",") {
				t.Fatalf("interpreter env keys = %q, want %q", gotKeys, wantKeys)
			}
			farm := report.Env["PATH"]
			if farm == "" || !filepath.IsAbs(farm) {
				t.Fatalf("PATH = %q, want the absolute farm directory", farm)
			}
			// The farm lives and dies with the invocation, so the
			// post-session proof is the string form plus the report:
			// the farm sits under the operation-private base the
			// session owned, never under the inherited entry.
			if !strings.HasPrefix(farm, mustPhysical(t, fixture.request.PrivateBase)+string(filepath.Separator)) {
				t.Fatalf("PATH %q is not under the operation-private base", farm)
			}
			if strings.Contains(farm, decoyDir) {
				t.Fatalf("PATH %q reaches the inherited entry", farm)
			}
			if strings.Join(result.Report.FarmEntries, ",") != filepath.Base(fixture.stub) {
				t.Fatalf("farm entries = %q, want exactly the resolved interpreter", result.Report.FarmEntries)
			}
			// Absent filesystem derives no project root: the invocation
			// works in its private temporary root.
			if report.Cwd != report.Env["TMPDIR"] {
				t.Fatalf("cwd = %q, want the private temporary root %q", report.Cwd, report.Env["TMPDIR"])
			}
			if _, present := report.Env["CSK_PROJECT_ROOT"]; present {
				t.Fatalf("CSK_PROJECT_ROOT = %q, want absent without a derived root", report.Env["CSK_PROJECT_ROOT"])
			}
			if report.Env["HOME"] != filepath.Dir(report.Env["TMPDIR"]) {
				t.Fatalf("HOME = %q, want the private operation base", report.Env["HOME"])
			}
		})
	}
}

// TestCapabilityDerivationNetworkHostsReportingOnly drives vector case
// `declared-network-hosts-are-reporting-only`: a host list is accepted,
// recorded, and never applied as a filter.
func TestCapabilityDerivationNetworkHostsReportingOnly(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.declareCapabilities(t, `{"network": ["api.example.com"], "env_read": ["HTTP_PROXY", "CURATOR_TEST_OK"]}`)
	fixture.request.HostEnvironment = []string{
		"HTTP_PROXY=http://host-proxy:8080",
		"CURATOR_TEST_OK=yes",
	}
	result, report := runDerivedSession(t, fixture)

	if result.Report.NetworkMode != networkModeRecorded {
		t.Fatalf("network mode = %q, want %q", result.Report.NetworkMode, networkModeRecorded)
	}
	if strings.Join(result.Report.RecordedHosts, ",") != "api.example.com" {
		t.Fatalf("recorded hosts = %q", result.Report.RecordedHosts)
	}
	// Recorded, not filtered: no filter state exists anywhere. The proxy
	// value stays out because the name is manager-owned, not because the
	// host list filtered it — the same withholding applies offline.
	if _, present := report.Env["HTTP_PROXY"]; present {
		t.Fatal("a manager-owned proxy value reached the interpreter")
	}
	if strings.Join(result.Report.WithheldEnv, ",") != "HTTP_PROXY" {
		t.Fatalf("withheld = %q, want the reserved proxy name reported", result.Report.WithheldEnv)
	}
	if report.Env["CURATOR_TEST_OK"] != "yes" {
		t.Fatal("the declared non-reserved variable did not pass through")
	}
}

// TestCapabilityDerivationExecIsManagerResolved drives vector case
// `declared-exec-is-manager-resolved`: the built PATH exposes exactly the
// resolved interpreter and the manager-resolved declared names, the
// inherited PATH is discarded, and the vector's declared git grant is
// resolved by the manager before the worker starts.
func TestCapabilityDerivationExecIsManagerResolved(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	execDir := fixture.request.ExecSearchDirs[0]
	gitName := "git"
	if runtime.GOOS == "windows" {
		gitName = "git.exe"
	}
	copyTestFile(t, fixture.stub, filepath.Join(execDir, gitName))
	// A caller-PATH decoy and a repository-planted impostor must both
	// lose to the manager mechanism.
	decoyDir := mustPhysical(t, t.TempDir())
	copyTestFile(t, fixture.stub, filepath.Join(decoyDir, gitName))
	copyTestFile(t, fixture.stub, filepath.Join(decoyDir, "decoyonly"+exeSuffix()))
	repoDir := fixture.request.ForbiddenRoots[0]
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	copyTestFile(t, fixture.stub, filepath.Join(repoDir, gitName))
	separator := string(os.PathListSeparator)
	interpBase := filepath.Base(fixture.stub)
	fixture.declareCapabilities(t, `{"exec": ["git"], "env_read": ["STUB_LOOKPATH", "STUB_LOOKPATH_MISSING"]}`)
	fixture.request.HostEnvironment = []string{
		"PATH=" + decoyDir,
		"STUB_LOOKPATH=" + interpBase + separator + "git",
		"STUB_LOOKPATH_MISSING=no-such-tool-zzz" + separator + "decoyonly",
	}
	result, report := runDerivedSession(t, fixture)

	farm := report.Env["PATH"]
	if strings.Contains(farm, separator) {
		t.Fatalf("PATH = %q, want exactly the farm directory", farm)
	}
	gitPath := report.LookPath["git"]
	if strings.HasPrefix(gitPath, "unresolved:") || gitPath == "" {
		t.Fatalf("bare git did not resolve: %q", gitPath)
	}
	if filepath.Dir(gitPath) != farm {
		t.Fatalf("bare git resolved to %q, want the farm %q", gitPath, farm)
	}
	if strings.HasPrefix(gitPath, decoyDir) || strings.HasPrefix(gitPath, repoDir) {
		t.Fatalf("bare git resolved outside the manager mechanism: %q", gitPath)
	}
	interpPath := report.LookPath[interpBase]
	if strings.HasPrefix(interpPath, "unresolved:") || interpPath == "" {
		t.Fatalf("bare interpreter did not resolve: %q", interpPath)
	}
	if filepath.Dir(interpPath) != farm {
		t.Fatalf("bare interpreter resolved to %q, want the farm", interpPath)
	}
	if report.LookPath["no-such-tool-zzz"] != "missing" {
		t.Fatalf("unresolvable exec lookpath = %q, want missing", report.LookPath["no-such-tool-zzz"])
	}
	if report.LookPath["decoyonly"] != "missing" {
		t.Fatalf("caller-PATH-only tool lookpath = %q, want missing (inherited PATH discarded)", report.LookPath["decoyonly"])
	}
	if resolved := result.Report.ResolvedExec["git"]; !strings.HasPrefix(mustPhysical(t, resolved), mustPhysical(t, execDir)) {
		t.Fatalf("resolved git = %q, want the manager search directory", resolved)
	}
	if len(result.Report.UnresolvedExec) != 0 {
		t.Fatalf("unresolved exec = %q, want no unresolved grants", result.Report.UnresolvedExec)
	}
	wantFarm := []string{interpBase, "git"}
	if runtime.GOOS == "windows" {
		wantFarm = []string{"git.exe", interpBase}
	}
	sort.Strings(wantFarm)
	gotFarm := append([]string(nil), result.Report.FarmEntries...)
	sort.Strings(gotFarm)
	if strings.Join(gotFarm, ",") != strings.Join(wantFarm, ",") {
		t.Fatalf("farm entries = %q, want %q", gotFarm, wantFarm)
	}
}

// TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied proves Launch
// proceeds with unresolved declarations, reports them, and keeps a same-name
// caller-PATH executable unreachable through the script's PATH.
func TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	const missing = "curator-r5-missing-manager-exec"
	const alsoMissing = "curator-r5-second-missing-manager-exec"
	decoyDir := mustPhysical(t, t.TempDir())
	copyTestFile(t, fixture.stub, filepath.Join(decoyDir, missing+exeSuffix()))
	fixture.declareCapabilities(t, `{"exec":["`+missing+`","`+alsoMissing+`"],"env_read":["STUB_LOOKPATH_MISSING"]}`)
	fixture.request.HostEnvironment = []string{
		"PATH=" + decoyDir,
		"STUB_LOOKPATH_MISSING=" + missing,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := Launch(ctx, fixture.request)
	if err != nil {
		t.Fatalf("Launch refused unresolved declarations: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Launch exit = %d, stderr %q; want the interpreter to run", result.ExitCode, result.Stderr)
	}
	observed := decodeStubReport(t, result.Stdout)
	if observed.LookPath[missing] != "missing" {
		t.Fatalf("caller-PATH decoy lookup = %q, want missing", observed.LookPath[missing])
	}
	farm := observed.Env["PATH"]
	if farm == "" || farm == decoyDir || strings.Contains(farm, decoyDir) {
		t.Fatalf("interpreter PATH = %q, want only the manager-built farm, excluding caller PATH %q", farm, decoyDir)
	}
	wantUnresolved := []string{missing, alsoMissing}
	if strings.Join(result.Report.UnresolvedExec, ",") != strings.Join(wantUnresolved, ",") {
		t.Fatalf("unresolved exec report = %q, want %q", result.Report.UnresolvedExec, wantUnresolved)
	}
	if len(result.Report.ResolvedExec) != 0 {
		t.Fatalf("resolved exec report = %q, want neither unresolved name resolved", result.Report.ResolvedExec)
	}
	wantFarmEntries := []string{filepath.Base(fixture.stub)}
	if strings.Join(result.Report.FarmEntries, ",") != strings.Join(wantFarmEntries, ",") {
		t.Fatalf("PATH farm entries = %q, want only the interpreter %q", result.Report.FarmEntries, wantFarmEntries)
	}
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// TestPathFarmPreservesSingleLinkIdentity proves farm construction never
// adds a filesystem link to a verified executable. A farm entry that
// hard-linked the interpreter would raise its link count and make the
// unchanged launch-boundary check refuse the session the manager just
// built — the Windows gate failure of run 35627712332, where every session
// refused with "multiple filesystem links or is a reparse point". The farm
// is built through the production constructor, then the interpreter
// re-verifies exactly as the launch boundary re-verifies it.
func TestPathFarmPreservesSingleLinkIdentity(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	interpreter, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": stub}, nil)
	if err != nil {
		t.Fatal(err)
	}
	farm, entries, err := buildPathFarm(farmRequest{
		parent:      mustPhysical(t, t.TempDir()),
		interpreter: interpreter,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0] != filepath.Base(interpreter.Path) {
		t.Fatalf("farm entries = %q, want exactly the interpreter", entries)
	}
	info, err := os.Lstat(interpreter.Path)
	if err != nil {
		t.Fatal(err)
	}
	multiple, err := godriver.HasMultipleLinks(interpreter.Path, info)
	if err != nil || multiple {
		t.Fatalf("the farm build left the interpreter multi-linked (multiple=%v, err=%v)", multiple, err)
	}
	if err := VerifyInterpreter(interpreter); err != nil {
		t.Fatalf("the interpreter no longer verifies after the farm build: %v", err)
	}
	if !farmEntryResolves(filepath.Join(farm, entries[0]), interpreter.Path) {
		t.Fatal("the farm entry does not resolve to the verified interpreter")
	}
}

// TestCapabilityDerivationSecretsRemainIdentifiers drives vector case
// `declared-secrets-remain-identifiers`: secret identifiers never resolve
// to values and never widen the environment.
func TestCapabilityDerivationSecretsRemainIdentifiers(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.declareCapabilities(t, `{"secrets": ["release-token"]}`)
	fixture.request.HostEnvironment = []string{"release-token=SUPERSECRET-VALUE"}
	result, report := runDerivedSession(t, fixture)

	if strings.Join(result.Report.SecretIDs, ",") != "release-token" {
		t.Fatalf("secret identifiers = %q", result.Report.SecretIDs)
	}
	for key, value := range report.Env {
		if strings.Contains(value, "SUPERSECRET-VALUE") {
			t.Fatalf("secret material reached the interpreter environment via %s", key)
		}
	}
	if _, present := report.Env["release-token"]; present {
		t.Fatal("the secret identifier widened the environment")
	}

	// The normalized-name shape: the host carries the upper-cased,
	// underscore-normalised name of a declared secret with no env_read
	// entry, and nothing containing the value may reach the interpreter.
	normalized := newLaunchFixture(t, os.Args[0])
	normalized.declareCapabilities(t, `{"secrets": ["release-token"]}`)
	normalized.request.HostEnvironment = []string{"RELEASE_TOKEN=SUPERSECRET-NORMALISED"}
	_, normalizedReport := runDerivedSession(t, normalized)
	for key, value := range normalizedReport.Env {
		if strings.Contains(value, "SUPERSECRET-NORMALISED") {
			t.Fatalf("secret material reached the interpreter environment via %s", key)
		}
	}

	// The control: naming the same variable through env_read passes the
	// host value by env_read semantics — that is passthrough, not secret
	// injection, and the report still carries only the identifier.
	second := newLaunchFixture(t, os.Args[0])
	second.declareCapabilities(t, `{"secrets": ["release-token"], "env_read": ["RELEASE_TOKEN"]}`)
	second.request.HostEnvironment = []string{"RELEASE_TOKEN=operator-value"}
	secondResult, secondReport := runDerivedSession(t, second)
	if secondReport.Env["RELEASE_TOKEN"] != "operator-value" {
		t.Fatal("the declared env_read variable did not pass through")
	}
	if strings.Join(secondResult.Report.SecretIDs, ",") != "release-token" {
		t.Fatalf("secret identifiers = %q", secondResult.Report.SecretIDs)
	}
}

// TestDerivedPathSetShapes pins the carried path set Protocol Core §4.1.1
// promises: exactly the declared paths beneath the canonical project root
// — the whole root for a `repo` declaration — and nothing when no path
// set is derivable. A member that escapes the root refuses fail-closed.
// The Linux Landlock row proves the worker grants exactly this set.
func TestDerivedPathSetShapes(t *testing.T) {
	stub := mustPhysical(t, stubInterpreterBinary(t))
	interpreter, err := ResolveInterpreter("python3-v1", map[string]string{"python3-v1": stub}, nil)
	if err != nil {
		t.Fatal(err)
	}
	derive := func(t *testing.T, capabilities, projectRoot string) (DerivedProfile, error) {
		t.Helper()
		declared, err := ParseDeclaredCapabilities(json.RawMessage(capabilities))
		if err != nil {
			return DerivedProfile{}, err
		}
		base := mustPhysical(t, t.TempDir())
		private := privateArea{base: base}
		for _, leaf := range []struct {
			name   string
			target *string
		}{{"tmp", &private.tmp}, {"config", &private.config}, {"cache", &private.cache}} {
			path := filepath.Join(base, leaf.name)
			if err := os.MkdirAll(path, 0o700); err != nil {
				t.Fatal(err)
			}
			*leaf.target = path
		}
		return DeriveProfile(DerivationInput{
			Declared:      declared,
			InterpreterID: "python3-v1",
			Interpreter:   interpreter,
			HostEnv:       []string{},
			Private:       private,
			FarmParent:    base,
			ProjectRoot:   projectRoot,
		})
	}
	root := mustPhysical(t, t.TempDir())
	sub := filepath.Join(root, "sub")
	nested := filepath.Join(sub, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	filesystemCaps := func(filesystem any) string {
		payload, err := json.Marshal(map[string]any{"filesystem": filesystem})
		if err != nil {
			t.Fatal(err)
		}
		return string(payload)
	}
	for _, testCase := range []struct {
		name         string
		capabilities string
		projectRoot  string
		want         []string
	}{
		{"repo-derives-root", `{"filesystem": "repo"}`, root, []string{root}},
		{"paths-derive-members", `{"filesystem": ["sub/nested", "sub"]}`, root, []string{sub, nested}},
		{"absolute-inside-derives", filesystemCaps([]string{sub}), root, []string{sub}},
		{"home-config-derives-nothing", `{"filesystem": "home-config"}`, root, nil},
		{"absent-derives-nothing", `{}`, root, nil},
		{"none-derives-nothing", `{"filesystem": "none"}`, root, nil},
		{"global-derives-nothing", `{"filesystem": "repo"}`, "", nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			profile, err := derive(t, testCase.capabilities, testCase.projectRoot)
			if err != nil {
				t.Fatalf("DeriveProfile refused: %v", err)
			}
			if strings.Join(profile.WritePaths, ",") != strings.Join(testCase.want, ",") {
				t.Fatalf("WritePaths = %q, want %q", profile.WritePaths, testCase.want)
			}
		})
	}
	outside := filepath.Join(mustPhysical(t, t.TempDir()), "curator-escape-probe")
	for _, testCase := range []struct {
		name         string
		capabilities string
	}{
		{"relative-escape-refuses", `{"filesystem": ["../evil"]}`},
		{"absolute-outside-refuses", filesystemCaps([]string{outside})},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := derive(t, testCase.capabilities, root); DiagnosticCode(err) != CodePackageInfluenceForbidden {
				t.Fatalf("DeriveProfile error = %v, want %s", err, CodePackageInfluenceForbidden)
			}
		})
	}
}

// TestManagerBuiltEnvironmentWithholdsReserved proves the reserved
// enumeration at the production entry: every reserved name stays out of
// the interpreter environment even when env_read names it and the host
// sets it, every withheld entry is reported, and the scoping rules hold —
// per-interpreter prefixes, macOS-only DYLD, exact POSIX matching, and
// case-insensitive Windows matching.
func TestManagerBuiltEnvironmentWithholdsReserved(t *testing.T) {
	// Reserved everywhere and never manager-set: withheld means absent.
	absent := []string{
		"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "FTP_PROXY", "NO_PROXY",
		"http_proxy", "https_proxy", "all_proxy", "ftp_proxy", "no_proxy",
		"RES_OPTIONS", "HOSTALIASES", "LOCALDOMAIN",
		"LD_PRELOAD", "LD_LIBRARY_PATH", "LD_AUDIT", "LD_SOMETHING_NEW",
		"IFS",
	}
	python := []string{"PYTHONPATH", "PYTHONHOME", "PYTHONSAFEPATH", "PYTHONNOUSERSITE", "__PYVENV_LAUNCHER__", "PYTHONFUTURE"}
	node := []string{"NODE_PATH", "NODE_OPTIONS", "NODE_EXTRA_CA_CERTS", "NPM_CONFIG_REGISTRY", "NPM_CONFIG_FUTURE"}
	// Reserved everywhere and manager-set: withheld means the
	// manager-set value stands, never the host value. CSK_PROJECT_ROOT
	// has no derived value in this test (no project root), so it is
	// absent here and bound in the runtime-area rows.
	stands := []string{"PATH", "HOME", "TMPDIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "CSK_PROJECT_ROOT"}
	// Windows-only reserved names: withheld there, ordinary elsewhere.
	windowsOnly := []string{"USERPROFILE", "APPDATA", "LOCALAPPDATA", "COMSPEC", "PATHEXT", "SYSTEMROOT"}
	// Reserved everywhere; manager-set on Windows only.
	tempNames := []string{"TEMP", "TMP"}
	// Platform-shaped spellings: withheld on their platform, passed
	// elsewhere (exact POSIX matching, case-insensitive Windows).
	folded := []string{"path", "home", "WINDIR", "Pathext", "lD_pReLoAd", "uSeRpRoFiLe"}
	dyld := []string{"DYLD_INSERT_LIBRARIES", "DYLD_FUTURE"}
	legit := []string{"CURATOR_TEST_LEGIT_ONE", "CURATOR_TEST_LEGIT_TWO", "SHELL"}

	for _, interpreter := range []string{"python3-v1", "node-v1"} {
		t.Run(interpreter, func(t *testing.T) {
			own := python
			foreign := node
			if interpreter == "node-v1" {
				own, foreign = node, python
			}
			var names []string
			names = append(names, absent...)
			names = append(names, own...)
			names = append(names, foreign...)
			names = append(names, stands...)
			names = append(names, windowsOnly...)
			names = append(names, tempNames...)
			names = append(names, folded...)
			names = append(names, dyld...)
			names = append(names, legit...)
			payload, err := json.Marshal(map[string]any{"env_read": names})
			if err != nil {
				t.Fatal(err)
			}
			fixture := newLaunchFixture(t, os.Args[0])
			fixture.request.InterpreterID = interpreter
			fixture.request.CapabilitiesRaw = payload
			host := make([]string, 0, len(names))
			for _, name := range names {
				host = append(host, name+"=host-value-for-"+name)
			}
			fixture.request.HostEnvironment = host
			result, report := runDerivedSession(t, fixture)

			withheld := map[string]bool{}
			for _, name := range result.Report.WithheldEnv {
				withheld[name] = true
			}
			mustWithholdAbsent := func(name string) {
				t.Helper()
				if _, present := report.Env[name]; present {
					t.Errorf("reserved %s reached the interpreter: %q", name, report.Env[name])
				}
				if !withheld[name] {
					t.Errorf("reserved %s was not reported withheld", name)
				}
			}
			mustWithholdStands := func(declared, canonical, want string) {
				t.Helper()
				if report.Env[canonical] != want {
					t.Errorf("%s: manager-set %s = %q, want %q (host value must not stand)",
						declared, canonical, report.Env[canonical], want)
				}
				if !withheld[declared] {
					t.Errorf("reserved %s was not reported withheld", declared)
				}
			}
			mustPass := func(name string) {
				t.Helper()
				if report.Env[name] != "host-value-for-"+name {
					t.Errorf("%s = %q, want the host value passed through", name, report.Env[name])
				}
				if withheld[name] {
					t.Errorf("%s was reported withheld", name)
				}
			}
			for _, name := range absent {
				mustWithholdAbsent(name)
			}
			for _, name := range own {
				mustWithholdAbsent(name)
			}
			// The other interpreter's prefixes are ordinary names here.
			for _, name := range foreign {
				mustPass(name)
			}
			farm, tmp, base := report.Env["PATH"], report.Env["TMPDIR"], report.Env["HOME"]
			mustWithholdStands("PATH", "PATH", farm)
			if farm == "host-value-for-PATH" || farm == "" || !filepath.IsAbs(farm) {
				t.Errorf("PATH = %q, want the farm directory", farm)
			}
			mustWithholdStands("HOME", "HOME", base)
			if base != filepath.Dir(tmp) {
				t.Errorf("HOME = %q, want the operation base", base)
			}
			mustWithholdStands("TMPDIR", "TMPDIR", tmp)
			mustWithholdStands("XDG_CONFIG_HOME", "XDG_CONFIG_HOME", filepath.Join(base, "config"))
			mustWithholdStands("XDG_CACHE_HOME", "XDG_CACHE_HOME", filepath.Join(base, "cache"))
			mustWithholdAbsent("CSK_PROJECT_ROOT")
			for _, name := range windowsOnly {
				if runtime.GOOS == "windows" {
					switch name {
					case "USERPROFILE":
						mustWithholdStands(name, name, base)
					case "APPDATA":
						mustWithholdStands(name, name, filepath.Join(base, "config"))
					case "LOCALAPPDATA":
						mustWithholdStands(name, name, filepath.Join(base, "cache"))
					case "PATHEXT":
						mustWithholdStands(name, name, windowsPathExtensions)
					default: // SYSTEMROOT, COMSPEC: indispensable host values the manager sets itself.
						mustWithholdStands(name, name, "host-value-for-"+name)
					}
				} else {
					mustPass(name)
				}
			}
			for _, name := range tempNames {
				if runtime.GOOS == "windows" {
					mustWithholdStands(name, name, tmp)
				} else {
					mustWithholdAbsent(name)
				}
			}
			for _, name := range folded {
				if runtime.GOOS == "windows" {
					switch name {
					case "path":
						mustWithholdStands(name, "PATH", farm)
					case "home":
						mustWithholdStands(name, "HOME", base)
					case "WINDIR":
						mustWithholdStands(name, "WINDIR", "host-value-for-WINDIR")
					case "Pathext":
						mustWithholdStands(name, "PATHEXT", windowsPathExtensions)
					case "uSeRpRoFiLe":
						mustWithholdStands(name, "USERPROFILE", base)
					default: // lD_pReLoAd: reserved, never manager-set.
						mustWithholdAbsent(name)
					}
				} else {
					mustPass(name)
				}
			}
			for _, name := range dyld {
				if runtime.GOOS == "darwin" {
					mustWithholdAbsent(name)
				} else {
					mustPass(name)
				}
			}
			for _, name := range legit {
				mustPass(name)
			}
		})
	}
}

// TestReservedEnvironmentNameTable pins the reserved enumeration per
// platform and interpreter as a unit bound: every listed name and prefix
// is reserved where the profile says, and the near misses are not.
func TestReservedEnvironmentNameTable(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		platform    string
		interpreter string
		reserved    bool
	}{
		{"PATH", "linux", "python3-v1", true},
		{"PATH", "darwin", "node-v1", true},
		{"PATH", "windows", "python3-v1", true},
		{"path", "windows", "python3-v1", true},
		{"path", "linux", "python3-v1", false},
		{"SystemRoot", "windows", "node-v1", true},
		{"systemroot", "linux", "node-v1", false},
		{"LD_PRELOAD", "linux", "python3-v1", true},
		{"LD_ANYTHING", "windows", "node-v1", true},
		{"ld_preload", "windows", "node-v1", true},
		{"ld_preload", "linux", "node-v1", false},
		{"LDX_FOO", "linux", "python3-v1", false},
		{"DYLD_INSERT_LIBRARIES", "darwin", "python3-v1", true},
		{"DYLD_INSERT_LIBRARIES", "linux", "python3-v1", false},
		{"DYLD_INSERT_LIBRARIES", "windows", "python3-v1", false},
		{"PYTHONPATH", "linux", "python3-v1", true},
		{"PYTHONPATH", "linux", "node-v1", false},
		{"pythonpath", "windows", "python3-v1", true},
		{"__PYVENV_LAUNCHER__", "darwin", "python3-v1", true},
		{"__PYVENV_LAUNCHER__", "darwin", "node-v1", false},
		{"NODE_OPTIONS", "linux", "node-v1", true},
		{"NODE_OPTIONS", "linux", "python3-v1", false},
		{"NPM_CONFIG_REGISTRY", "windows", "node-v1", true},
		{"NODE", "linux", "node-v1", false},
		{"http_proxy", "linux", "python3-v1", true},
		{"Http_Proxy", "linux", "python3-v1", false},
		{"Http_Proxy", "windows", "python3-v1", true},
		{"LOCALDOMAIN", "darwin", "node-v1", true},
		{"SHELL", "linux", "python3-v1", false},
		{"EDITOR", "windows", "node-v1", false},
		{"MYPYTHON", "linux", "python3-v1", false},
		{"", "linux", "python3-v1", false},
	} {
		t.Run(testCase.platform+"/"+testCase.interpreter+"/"+testCase.name, func(t *testing.T) {
			if got := reservedEnvironmentName(testCase.name, testCase.platform, testCase.interpreter); got != testCase.reserved {
				t.Fatalf("reserved = %v, want %v", got, testCase.reserved)
			}
		})
	}
}

// TestOfflineScrubWhenNetworkNone proves offline configuration at the
// production entry: when the derived network is none — whether declared
// "none" or absent — no proxy or resolver configuration survives in the
// interpreter environment even when env_read names it.
func TestOfflineScrubWhenNetworkNone(t *testing.T) {
	for _, testCase := range []struct {
		name string
		raw  string
	}{
		{"declared-none", `{"network": "none", "env_read": ["HTTP_PROXY", "http_proxy", "NO_PROXY", "RES_OPTIONS", "HOSTALIASES", "LOCALDOMAIN", "CURATOR_TEST_OK"]}`},
		{"absent-derives-none", `{"env_read": ["HTTP_PROXY", "http_proxy", "NO_PROXY", "RES_OPTIONS", "HOSTALIASES", "LOCALDOMAIN", "CURATOR_TEST_OK"]}`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newLaunchFixture(t, os.Args[0])
			fixture.declareCapabilities(t, testCase.raw)
			fixture.request.HostEnvironment = []string{
				"HTTP_PROXY=http://proxy:8080", "http_proxy=http://proxy:8080",
				"NO_PROXY=example.com", "RES_OPTIONS=ndots:1",
				"HOSTALIASES=/host/aliases", "LOCALDOMAIN=host.example",
				"CURATOR_TEST_OK=yes",
			}
			result, report := runDerivedSession(t, fixture)
			if result.Report.NetworkMode != networkModeOffline {
				t.Fatalf("network mode = %q, want offline", result.Report.NetworkMode)
			}
			for _, name := range []string{"HTTP_PROXY", "http_proxy", "NO_PROXY", "RES_OPTIONS", "HOSTALIASES", "LOCALDOMAIN"} {
				if value, present := report.Env[name]; present {
					t.Errorf("offline env carries %s=%q", name, value)
				}
			}
			if report.Env["CURATOR_TEST_OK"] != "yes" {
				t.Error("the declared non-reserved variable did not pass through")
			}
		})
	}
}

// TestInterpreterWithoutReservedSetIsRefused proves an interpreter
// identifier the manager has no reserved set for never launches: the
// production entry refuses rather than running with an unfiltered
// environment.
func TestInterpreterWithoutReservedSetIsRefused(t *testing.T) {
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.request.InterpreterID = "ruby-v1"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := runSession(ctx, fixture.request, nil)
	if DiagnosticCode(err) != CodePackageInfluenceForbidden {
		t.Fatalf("unknown-interpreter error = %v, want %s", err, CodePackageInfluenceForbidden)
	}
	// The derivation layer refuses the same identifier on its own: the
	// reserved set is the gate, not just resolution order.
	declared, err := ParseDeclaredCapabilities(nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = DeriveProfile(DerivationInput{Declared: declared, InterpreterID: "ruby-v1"})
	if DiagnosticCode(err) != CodePackageInfluenceForbidden {
		t.Fatalf("derivation error = %v, want %s", err, CodePackageInfluenceForbidden)
	}
}

// TestOperationPrivateRuntimeAreaApplied proves the runtime area is bound,
// not just created: the manager-selected working directory follows the
// filesystem derivation, and the temporary, configuration, and cache
// roots reach the interpreter through the platform environment.
func TestOperationPrivateRuntimeAreaApplied(t *testing.T) {
	project := mustPhysical(t, t.TempDir())
	for _, testCase := range []struct {
		name        string
		raw         string
		projectRoot string
		wantCwd     string
		wantCSK     string
	}{
		{"absent-derives-no-root", `{}`, project, "tmp", ""},
		{"repo-derives-project-root", `{"filesystem": "repo"}`, project, "project", project},
		{"repo-without-project-denies", `{"filesystem": "repo"}`, "", "tmp", ""},
		{"home-config-derives-nothing", `{"filesystem": "home-config"}`, project, "tmp", ""},
		{"paths-derive-project-root", `{"filesystem": ["rel/path"]}`, project, "project", project},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newLaunchFixture(t, os.Args[0])
			fixture.declareCapabilities(t, testCase.raw)
			fixture.request.ProjectRoot = testCase.projectRoot
			if testCase.name == "paths-derive-project-root" {
				// The row declares the path set ["rel/path"]: the
				// derived members must exist, because the Linux write
				// confinement rules exactly the derived set and a member
				// that cannot be ruled refuses the invocation.
				if err := os.MkdirAll(filepath.Join(testCase.projectRoot, "rel", "path"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			_, report := runDerivedSession(t, fixture)

			wantCwd := report.Env["TMPDIR"]
			if testCase.wantCwd == "project" {
				wantCwd = testCase.projectRoot
			}
			if report.Cwd != wantCwd {
				t.Fatalf("cwd = %q, want %q", report.Cwd, wantCwd)
			}
			if got := report.Env["CSK_PROJECT_ROOT"]; got != testCase.wantCSK {
				t.Fatalf("CSK_PROJECT_ROOT = %q, want %q", got, testCase.wantCSK)
			}
			base := filepath.Dir(report.Env["TMPDIR"])
			if report.Env["HOME"] != base {
				t.Errorf("HOME = %q, want the operation base %q", report.Env["HOME"], base)
			}
			if report.Env["XDG_CONFIG_HOME"] != filepath.Join(base, "config") {
				t.Errorf("XDG_CONFIG_HOME = %q, want the private configuration root", report.Env["XDG_CONFIG_HOME"])
			}
			for _, key := range []string{"XDG_CACHE_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME"} {
				if report.Env[key] != filepath.Join(base, "cache") {
					t.Errorf("%s = %q, want the private cache root", key, report.Env[key])
				}
			}
			if runtime.GOOS == "windows" {
				if report.Env["TEMP"] != report.Env["TMPDIR"] || report.Env["TMP"] != report.Env["TMPDIR"] {
					t.Error("TEMP/TMP are not bound to the private temporary root")
				}
				if report.Env["APPDATA"] != report.Env["XDG_CONFIG_HOME"] {
					t.Error("APPDATA is not bound to the private configuration root")
				}
				if report.Env["LOCALAPPDATA"] != report.Env["XDG_CACHE_HOME"] {
					t.Error("LOCALAPPDATA is not bound to the private cache root")
				}
				if report.Env["USERPROFILE"] != base {
					t.Error("USERPROFILE is not bound to the operation base")
				}
			}
		})
	}
}

// TestWorkerRevalidatesDerivedEnvironment proves the worker applies the
// same derived-environment check the manager ran as a self-check: a
// mis-derived request refuses at the worker boundary before the
// interpreter starts.
func TestWorkerRevalidatesDerivedEnvironment(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	rebind := func(environment []string, key, value string) []string {
		rebound := make([]string, 0, len(environment))
		for _, item := range environment {
			if itemKey, _, _ := strings.Cut(item, "="); itemKey == key {
				continue
			}
			rebound = append(rebound, item)
		}
		return append(rebound, key+"="+value)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(t *testing.T, fixture *workerFixture, wire *workerRequest)
	}{
		{"offline-carries-proxy", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.Environment = append(wire.Environment, "HTTP_PROXY=http://proxy:8080")
		}},
		{"tmpdir-unbound", func(_ *testing.T, fixture *workerFixture, wire *workerRequest) {
			wire.Environment = rebind(wire.Environment, "TMPDIR", fixture.workDir)
		}},
		{"path-doubled", func(_ *testing.T, fixture *workerFixture, wire *workerRequest) {
			wire.Environment = rebind(wire.Environment, "PATH", fixture.farm+string(os.PathListSeparator)+fixture.farm)
		}},
		{"path-elsewhere", func(_ *testing.T, fixture *workerFixture, wire *workerRequest) {
			wire.Environment = rebind(wire.Environment, "PATH", fixture.workDir)
		}},
		{"farm-outside-base", func(t *testing.T, _ *workerFixture, wire *workerRequest) {
			outside := mustPhysical(t, t.TempDir())
			wire.FarmDir = outside
			wire.Environment = rebind(wire.Environment, "PATH", outside)
		}},
		{"csk-without-root", func(_ *testing.T, _ *workerFixture, wire *workerRequest) {
			wire.Environment = append(wire.Environment, "CSK_PROJECT_ROOT=/elsewhere")
		}},
		{"csk-mismatch", func(_ *testing.T, fixture *workerFixture, wire *workerRequest) {
			wire.ProjectRoot = fixture.workDir
			wire.Environment = append(wire.Environment, "CSK_PROJECT_ROOT=/elsewhere")
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker := startRawWorker(t, os.Args[0])
			fixture := newWorkerFixture(t, os.Args[0])
			wire := fixture.request()
			wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
			testCase.mutate(t, fixture, wire)
			worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
			worker.expectFailure(CodeWorkerProtocolInvalid)
		})
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a mis-derived request")
	}
}

// TestWorkerWaitsForPermit proves the interpreter never starts before the
// parent's permit: a worker that hears ready and then silence runs
// nothing, and a shutdown instead of the permit ends the session cleanly.
func TestWorkerWaitsForPermit(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	marker := filepath.Join(fixture.workDir, "ran")
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	time.Sleep(300 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran before the permit")
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
	// The declined worker exits on its own; wait for the clean exit
	// rather than tearing the domain down first.
	waited := make(chan error, 1)
	go func() { waited <- worker.command.Wait() }()
	select {
	case err := <-waited:
		if err != nil {
			t.Fatalf("declined worker exit = %v, want a clean exit", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the declined worker did not exit")
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran after the declined permit")
	}
}

// TestWorkerRejectsPermitWithForeignNonce proves the permit is bound to
// the session nonce like every other frame.
func TestWorkerRejectsPermitWithForeignNonce(t *testing.T) {
	worker := startRawWorker(t, os.Args[0])
	fixture := newWorkerFixture(t, os.Args[0])
	marker := filepath.Join(fixture.workDir, "ran")
	wire := fixture.request()
	wire.Environment = append(wire.Environment, "STUB_MARKER="+marker)
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: wire})
	if ready := worker.receive(); ready.Kind != kindReady {
		t.Fatalf("worker sent %q, want ready", ready.Kind)
	}
	worker.send(workerMessage{Kind: kindPermit, Nonce: strings.Repeat("00", 32)})
	worker.expectFailure(CodeWorkerProtocolInvalid)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the interpreter ran for a foreign permit nonce")
	}
}

// TestParentWithholdsPermitOnProofMismatch proves the parent validates the
// worker's identity proof before permitting the run: a worker that lies
// about the manager identity, or about the interpreter identity, earns no
// permit and the run refuses.
func TestParentWithholdsPermitOnProofMismatch(t *testing.T) {
	forge := mustPhysical(t, forgeWorkerBinary(t))
	for _, testCase := range []struct {
		name   string
		suffix string
	}{
		{"forged-manager-proof", ""},
		{"forged-interpreter-proof", ".forge-interp"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newLaunchFixture(t, forge)
			entry := fixture.entry + testCase.suffix
			if testCase.suffix != "" {
				writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
			}
			fixture.request.RuntimeEntry = entry
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, err := runSession(ctx, fixture.request, nil)
			if DiagnosticCode(err) != CodeWorkerIdentityInvalid {
				t.Fatalf("forged-proof error = %v, want %s", err, CodeWorkerIdentityInvalid)
			}
			// Fail fast on a pre-launch refusal: the forged proof is the
			// only refusal this row may see, so anything else (for
			// example a launch-boundary identity refusal before the
			// worker starts) fails here with its detail instead of
			// waiting 10 s for a record a never-started worker cannot
			// write.
			if err == nil || !strings.Contains(err.Error(), "identity proof") {
				t.Fatalf("forged-proof error = %v, want the proof-mismatch refusal", err)
			}
			waitForForgeRecord(t, entry+".forge-shutdown")
			if _, statErr := os.Stat(entry + ".forge-permit"); statErr == nil {
				t.Fatal("the parent permitted a forged proof")
			}
		})
	}
}

// TestParentRejectsWrongStartedCount proves the parent requires exactly
// one program below the worker: a worker that reports two started refuses
// even with a truthful identity proof.
func TestParentRejectsWrongStartedCount(t *testing.T) {
	forge := mustPhysical(t, forgeWorkerBinary(t))
	fixture := newLaunchFixture(t, forge)
	entry := fixture.entry + ".forge-started"
	writeTestFile(t, entry, []byte("print('tool')\n"), 0o644)
	fixture.request.RuntimeEntry = entry
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := runSession(ctx, fixture.request, nil)
	if DiagnosticCode(err) != CodeWorkerIdentityInvalid {
		t.Fatalf("started-count error = %v, want %s", err, CodeWorkerIdentityInvalid)
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("started-count error = %v, want the count diagnostic", err)
	}
	waitForForgeRecord(t, entry+".forge-permit")
}

func waitForForgeRecord(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the forge worker never recorded %s", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestLaunchDrivesDerivation proves the production entry runs the derived
// profile now that the preflight table is complete: admission proceeds,
// the launch runs, and the report records the derivation.
func TestLaunchDrivesDerivation(t *testing.T) {
	if err := scriptpolicy.Admit(map[string]skillspec.Command{
		"tool": {Name: "tool", Type: "script", ExecutionPolicy: skillspec.ScriptExecutionPolicy, Interpreter: "python3-v1"},
	}); err != nil {
		t.Fatalf("Admit refused with the complete table: %v", err)
	}
	fixture := newLaunchFixture(t, os.Args[0])
	fixture.declareCapabilities(t, `{"env_read": ["STUB_EXIT"], "exec": "none", "network": "none"}`)
	fixture.request.HostEnvironment = []string{"STUB_EXIT=0"}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := Launch(ctx, fixture.request)
	if err != nil {
		t.Fatalf("Launch refused with the complete table: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d (stderr %q)", result.ExitCode, result.Stderr)
	}
	if result.Report.NetworkMode != networkModeOffline {
		t.Fatalf("network mode = %q, want offline", result.Report.NetworkMode)
	}
	report := decodeStubReport(t, result.Stdout)
	if report.Env["STUB_EXIT"] != "0" {
		t.Fatal("the declared passthrough did not reach the interpreter")
	}
}

// TestNativeLauncherWorkerModeWinsOverShimDispatch proves a sidecar-paired
// launcher re-executed in the fixed hidden worker mode speaks the worker
// protocol instead of redispatching into a second launcher: the worker is
// the launcher's own re-execution, so dispatch order is load-bearing.
func TestNativeLauncherWorkerModeWinsOverShimDispatch(t *testing.T) {
	curator := mustPhysical(t, builtCuratorBinary(t))
	binDir := mustPhysical(t, t.TempDir())
	launcher := filepath.Join(binDir, "tool"+exeSuffix())
	copyTestFile(t, curator, launcher)
	sidecar, err := NewShimSidecar("skill", "tool", "python3-v1",
		filepath.Join(binDir, "entry"), binDir, "", 8, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := sidecar.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, launcher+ShimSidecarSuffix, payload, 0o644)

	worker := startRawWorker(t, launcher)
	fixture := newWorkerFixture(t, launcher)
	fixture.nonce = strings.Repeat("cd", 32)
	worker.send(workerMessage{Kind: kindRequest, Nonce: fixture.nonce, Request: fixture.request()})
	ready := worker.receive()
	if ready.Kind != kindReady || ready.Ready == nil {
		t.Fatalf("launcher in worker mode sent %q, want the ready proof (not a nested launcher session)", ready.Kind)
	}
	if ready.Ready.ExecutableSHA256 != fixture.manager.SHA256 {
		t.Fatal("the launcher proved a different identity than the session resolved")
	}
	worker.permit(fixture.nonce)
	result := worker.receive()
	if result.Kind != kindResult || result.Result == nil {
		t.Fatalf("launcher in worker mode sent %q, want one invocation result", result.Kind)
	}
	worker.send(workerMessage{Kind: kindShutdown, Nonce: fixture.nonce})
}

// TestProductionBinaryLaunchesWhenHostProvides proves admission honesty on
// the production image: a fresh production process admits an enforced
// command through `skill check`, an executed native launcher with an
// operator-bound interpreter runs the script end to end, and a launcher
// with no binding refuses `script_execution_control_unavailable` because
// the interpreter cannot be resolved on this host.
func TestProductionBinaryLaunchesWhenHostProvides(t *testing.T) {
	curator := mustPhysical(t, builtCuratorBinary(t))
	stub := mustPhysical(t, stubInterpreterBinary(t))

	skillDir := t.TempDir()
	writeTestFile(t, filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: s\ndescription: d\n---\n# s\n"), 0o644)
	writeTestFile(t, filepath.Join(skillDir, "scripts", "tool"), []byte("print(1)\n"), 0o755)
	writeTestFile(t, filepath.Join(skillDir, "agent-skill.json"), []byte(`{
  "schema_version": 8,
  "capabilities": {},
  "commands": {"tool": {"type": "script", "unix_path": "scripts/tool",
    "execution_policy": "script-worker-v1", "interpreter": "python3-v1"}}
}`), 0o644)

	// The `skill check` surface admits the enforced command: the control
	// table is complete.
	check := exec.Command(curator, "skill", "check", skillDir)
	check.Env = append(os.Environ(),
		"HOME="+t.TempDir(),
		"CURATOR_CONFIG="+filepath.Join(t.TempDir(), "config.json"))
	if output, err := check.CombinedOutput(); err != nil {
		t.Fatalf("skill check refused an enforced command: %v: %s", err, output)
	}

	binDir := mustPhysical(t, t.TempDir())
	launcher := filepath.Join(binDir, "tool"+exeSuffix())
	copyTestFile(t, curator, launcher)
	sidecar, err := NewShimSidecar("skill", "tool", "python3-v1",
		filepath.Join(skillDir, "scripts", "tool"), skillDir, "", 8, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := sidecar.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, launcher+ShimSidecarSuffix, payload, 0o644)

	// With an operator-bound interpreter the executed launcher runs the
	// script: shim dispatch works and the worker session completes.
	boundConfig, err := json.Marshal(map[string]any{
		"schema_version": 1, "skills_root": t.TempDir(), "projects": map[string]any{},
		"script_interpreters": map[string]string{"python3-v1": stub},
	})
	if err != nil {
		t.Fatal(err)
	}
	bound := filepath.Join(t.TempDir(), "config.json")
	writeTestFile(t, bound, boundConfig, 0o644)
	launched := exec.Command(launcher, "--help")
	launched.Env = append(os.Environ(), "CURATOR_CONFIG="+bound)
	launched.Dir = t.TempDir()
	launched.Stdin = strings.NewReader("")
	shimOutput, shimErr := launched.CombinedOutput()
	if shimErr != nil {
		t.Fatalf("the native launcher refused with a bound interpreter: %v: %s", shimErr, shimOutput)
	}
	report := decodeStubReport(t, shimOutput)
	if len(report.Argv) == 0 || report.Argv[len(report.Argv)-1] != "--help" {
		t.Fatalf("stub argv = %q, want the forwarded argument", report.Argv)
	}

	// With no binding the same launcher refuses control-unavailable: the
	// interpreter cannot be resolved on this host, so the
	// interpreter-resolution control is unavailable here.
	unbound := exec.Command(launcher, "--help")
	unbound.Env = append(os.Environ(), "CURATOR_CONFIG="+filepath.Join(t.TempDir(), "config.json"))
	unbound.Dir = t.TempDir()
	unbound.Stdin = strings.NewReader("")
	deniedOutput, deniedErr := unbound.CombinedOutput()
	if deniedErr == nil {
		t.Fatalf("the native launcher ran with no interpreter binding: %s", deniedOutput)
	}
	if !strings.Contains(string(deniedOutput), scriptpolicy.ControlUnavailable) {
		t.Fatalf("launcher output does not name %s: %s", scriptpolicy.ControlUnavailable, deniedOutput)
	}
	if exitErr, ok := deniedErr.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("launcher error = %v, want exit 1", deniedErr)
	}
}
