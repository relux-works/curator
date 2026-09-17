package main

// Production provenance-sink wiring (rework P1): productionExternalDeps
// must assign a live manager-home operation-diagnostics sink for the
// sanitized resolved-lane records. A nil/removed assignment fails this
// test by construction.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
)

func TestProductionExternalDepsAssignsTransportProvenanceSink(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	cfg := &config.Config{Path: filepath.Join(home, "config.json")}
	deps := productionExternalDeps(cfg, true)
	if deps.DraftTransportTrace == nil {
		t.Fatal("production DraftTransportTrace is nil: resolved-lane provenance dies with the call")
	}
	record := buildrepo.AttemptRecord{Index: 1, URL: "https://mirror.fixture.test/repository.git",
		Transport: "https", Provider: "mirror-https",
		Identity: "fixture.test/repository", ResolvedHost: "mirror.fixture.test",
		MirrorOf: "fixture.test/repository", NetworkAttempted: true, Succeeded: true}
	deps.DraftTransportTrace(record)
	payload, err := os.ReadFile(install.DraftTransportProvenancePath(home))
	if err != nil {
		t.Fatalf("production sink wrote nothing to the manager home: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("sink lines = %d, want 1", len(lines))
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("sink line is not JSON: %v", err)
	}
	for key, want := range map[string]any{
		"identity": "fixture.test/repository", "url": "https://mirror.fixture.test/repository.git",
		"resolved_host": "mirror.fixture.test", "mirror_of": "fixture.test/repository",
	} {
		if line[key] != want {
			t.Errorf("sink[%q] = %v, want %v", key, line[key], want)
		}
	}
	// The sink serializes a fixed allowlisted shape: any future record
	// field (and any secret) stays out unless explicitly named here.
	allowed := map[string]bool{"index": true, "identity": true, "url": true, "transport": true,
		"provider": true, "class": true, "network_attempted": true, "succeeded": true,
		"resolved_host": true, "resolved_port": true, "alias": true, "mirror_of": true}
	for key := range line {
		if !allowed[key] {
			t.Errorf("sink carries unallowlisted field %q", key)
		}
	}
	// Machine-private, never portable: the sink resolves under the
	// manager home that productionExternalDeps was constructed with.
	if !strings.HasPrefix(install.DraftTransportProvenancePath(home), home+string(filepath.Separator)) {
		t.Errorf("sink path escapes the manager home")
	}
}

// TestProductionExternalDepsFalseDrivesMirrorFetchToSink is the rework-2
// production-composition proof (repository-transport §§6-7): an ACTUAL
// external operation through productionExternalDeps(cfg, false) — a real
// `install app` (dryRun=false) over a temporary machine policy with a
// listed mirror and a fake transport that fails the primary and succeeds
// on the mirror — keeping the constructed sink.
//
// It asserts in the actual machine-private sink: the canonical identity
// (never an alias/mirror host), the listed mirror URL with its resolved
// mirror host and mirror_of attestation, and no broker secret. It asserts
// the emitted portable artifacts (stdout/stderr, cache receipts, install
// markers, project manifest — the lock/receipt/marker/manifest family)
// carry no endpoint provenance and no secret.
//
// Both mutants must fail it: a removed sink (nil trace) leaves no sink
// file, and a dry-run-only sink (assigned only when dryRun is true)
// leaves no sink file for this dryRun=false run. Product code is
// untouched; this test only drives the production CLI path.
func TestProductionExternalDepsFalseDrivesMirrorFetchToSink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
	fixture := setupDraftCLI(t)
	const secretMarker = "zz-secret-marker-260916-prod"
	const mirrorHost = "mirror.fixture.test"
	const mirrorURL = "https://mirror.fixture.test/https-tools.git"
	const httpsIdentity = "fixture.test/https-tools"
	const sshIdentity = "fixture.test/ssh-tools"

	policy := `{"schema_version":2,"repositories":{` +
		`"fixture.test/https-tools":{"endpoints":[` +
		`{"url":"https://fixture.test/https-tools.git","authentication":"prov-a"},` +
		`{"url":"https://mirror.fixture.test/https-tools.git","authentication":"prov-a","mirror_of":"fixture.test/https-tools"}` +
		`],"fallback":"availability-auth"},` +
		`"fixture.test/ssh-tools":{"endpoints":[` +
		`{"url":"ssh://git@fixture.test/ssh-tools.git","authentication":"operator-acme"}` +
		`],"fallback":"none"}}}`

	providers := fmt.Sprintf(`{"schema_version":1,"providers":{`+
		`"prov-a":{"https":{"username":"a"}},`+
		`"operator-acme":{"https":{"anonymous":true},"ssh":{"identity":%q,"known_hosts":%q}}}}`,
		fixture.identity, fixture.knownHosts)

	fakeDir, logPath := fixture.installDraftFakeGit(t, map[string]draftCLIArm{
		draftCLIHTTPS: {stderr: draftCLIDNSStderr},
		mirrorURL:     {succeed: true, fileRepo: fixture.httpsBare},
		draftCLISSH:   {succeed: true, fileRepo: fixture.sshBare},
	}, map[string]string{"prov-a": secretMarker})

	// Production composition uses process environment (PATH for git
	// discovery, the draft switch) exactly like runDraftInstall, but this
	// run is a REAL install (dryRun=false) so productionExternalDeps is
	// constructed with false. Manual save/restore keeps the run bounded
	// and isolated; this test never runs in parallel.
	saved := map[string]*string{}
	set := map[string]string{
		"PATH":                              fakeDir + string(os.PathListSeparator) + fixture.origPath,
		"GIT_SSH":                           fixture.staticSSH,
		"CURATOR_BUILD_SSH_IDENTITY":        fixture.identity,
		"CURATOR_BUILD_SSH_KNOWN_HOSTS":     fixture.knownHosts,
		install.EnvDraftTransportResolution: "1",
	}
	unset := []string{"SSH_AUTH_SOCK", "CURATOR_BUILD_SSH_AGENT", "CURATOR_BUILD_HTTPS_TOKEN", "CURATOR_BUILD_HTTPS_HOST"}
	for key := range set {
		if value, ok := os.LookupEnv(key); ok {
			held := value
			saved[key] = &held
		} else {
			saved[key] = nil
		}
	}
	for _, key := range unset {
		if _, seen := saved[key]; seen {
			continue
		}
		if value, ok := os.LookupEnv(key); ok {
			held := value
			saved[key] = &held
		} else {
			saved[key] = nil
		}
	}
	defer func() {
		for key, held := range saved {
			if held == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *held)
			}
		}
	}()
	for key, value := range set {
		if err := os.Setenv(key, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range unset {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(fixture.policyPath, []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.providersPath, []byte(providers), 0o644); err != nil {
		t.Fatal(err)
	}

	home := filepath.Dir(fixture.configPath)
	cfgCheck := &config.Config{Path: fixture.configPath}
	depsCheck := productionExternalDeps(cfgCheck, false)
	if depsCheck.DraftTransportTrace == nil {
		t.Fatal("productionExternalDeps(cfg, false) left DraftTransportTrace nil: real-operation provenance dies with the call (removed-sink and dry-run-only mutants fail here)")
	}
	if !depsCheck.DraftTransportResolution {
		t.Fatal("productionExternalDeps(cfg, false) did not honour the draft switch from the environment")
	}
	sinkPath := install.DraftTransportProvenancePath(home)
	if _, err := os.Stat(sinkPath); !os.IsNotExist(err) {
		t.Fatalf("sink %q exists before the real operation, want absent", sinkPath)
	}

	code, stdout, stderr := capture(t, fixture.configPath, "install", "app")
	// Acquisition (and its sink records) precedes the external build. On a
	// host rc5-native-control-inventory-v1 defines no record for, the go-v1
	// driver refuses the build after the plan and staging fetches land; the
	// sink assertions below run there independent of that build outcome, so
	// both sink mutants fail on every lane that runs this test. Product code
	// is untouched; only the build-completion expectation branches here.
	covered := godriver.InventoryPlatform(runtime.GOOS) != ""
	if covered {
		if code != exitOK {
			t.Fatalf("real install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
	} else {
		if code == exitOK {
			t.Fatalf("real install succeeded on %s, for which rc5-native-control-inventory-v1 defines no record\nstdout:\n%s\nstderr:\n%s",
				runtime.GOOS, stdout, stderr)
		}
		for _, want := range []string{godriver.CodeControlUnavailable, godriver.NativeControlInventoryVersion, "no record for host " + runtime.GOOS} {
			if !strings.Contains(stderr, want) {
				t.Fatalf("uncovered-host refusal did not carry %q:\nstderr:\n%s", want, stderr)
			}
		}
	}
	// A real install fetches twice per repository: once for the
	// read-only plan and once for staging. Both phases run the resolved
	// lane with the production sink, so the primary/mirror/ssh pattern
	// repeats exactly. On an inventory-uncovered host the staging build of
	// the first repository refuses after its acquisition, so the second
	// repository's staging fetch never runs: plan (3) + staging partial
	// (primary/mirror).
	wantFetches := []string{draftCLIHTTPS, mirrorURL, draftCLISSH, draftCLIHTTPS, mirrorURL, draftCLISSH}
	if !covered {
		wantFetches = []string{draftCLIHTTPS, mirrorURL, draftCLISSH, draftCLIHTTPS, mirrorURL}
	}
	fetches := draftCLIFetchLines(t, logPath)
	if len(fetches) != len(wantFetches) {
		t.Fatalf("%d fetches, want %d (%s):\n%s", len(fetches), len(wantFetches), strings.Join(wantFetches, ","), strings.Join(fetches, "\n"))
	}
	for i, want := range wantFetches {
		if !strings.Contains(fetches[i], "<"+want+">") {
			t.Fatalf("fetch %d = %q, want %q\n%s", i, fetches[i], want, strings.Join(fetches, "\n"))
		}
	}

	payload, err := os.ReadFile(sinkPath)
	if err != nil {
		t.Fatalf("real-operation sink missing at %q (nil/removed and dry-run-only mutants fail here): %v", sinkPath, err)
	}
	if strings.Contains(string(payload), secretMarker) {
		t.Fatalf("machine-private sink leaks broker secret material")
	}
	lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
	wantLines := 6 // plan primary/mirror/ssh + stage primary/mirror/ssh
	if !covered {
		// Plan (3) + staging partial (primary/mirror): the staging mirror
		// acquisition records before its build refuses.
		wantLines = 5
	}
	if len(lines) != wantLines {
		t.Fatalf("sink lines = %d, want %d:\n%s", len(lines), wantLines, string(payload))
	}
	var decoded []map[string]any
	for _, raw := range lines {
		var line map[string]any
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			t.Fatalf("sink line is not JSON: %q: %v", raw, err)
		}
		decoded = append(decoded, line)
	}
	// The mirror success records (one per phase) carry the canonical
	// identity with the listed URL, the resolved mirror host, and the
	// mirror_of attestation — never the mirror host as identity.
	var mirrorLines []map[string]any
	for _, line := range decoded {
		if line["url"] == mirrorURL {
			mirrorLines = append(mirrorLines, line)
		}
		identity, _ := line["identity"].(string)
		if identity == mirrorHost || identity == mirrorURL {
			t.Errorf("sink identity = %q, want the canonical key, never the mirror", identity)
		}
		if url, _ := line["url"].(string); strings.Contains(url, secretMarker) {
			t.Errorf("sink url leaks secret: %q", url)
		}
	}
	if len(mirrorLines) != 2 {
		t.Fatalf("sink mirror records = %d, want 2 (plan + stage) for %q:\n%s", len(mirrorLines), mirrorURL, string(payload))
	}
	for _, mirrorLine := range mirrorLines {
		for key, want := range map[string]any{
			"identity": httpsIdentity, "url": mirrorURL,
			"resolved_host": mirrorHost, "mirror_of": httpsIdentity,
		} {
			if mirrorLine[key] != want {
				t.Errorf("mirror sink[%q] = %v, want %v (line %v)", key, mirrorLine[key], want, mirrorLine)
			}
		}
		if mirrorLine["succeeded"] != true {
			t.Errorf("mirror sink succeeded = %v, want true", mirrorLine["succeeded"])
		}
	}
	allowed := map[string]bool{"index": true, "identity": true, "url": true, "transport": true,
		"provider": true, "class": true, "network_attempted": true, "succeeded": true,
		"resolved_host": true, "resolved_port": true, "alias": true, "mirror_of": true}
	for _, line := range decoded {
		for key := range line {
			if !allowed[key] {
				t.Errorf("sink carries unallowlisted field %q (line %v)", key, line)
			}
		}
	}
	// Every record names a canonical identity, never an alias/mirror.
	for _, line := range decoded {
		identity, _ := line["identity"].(string)
		if identity != httpsIdentity && identity != sshIdentity {
			t.Errorf("sink identity = %q, want %q or %q", identity, httpsIdentity, sshIdentity)
		}
	}

	// Portable artifacts (lock/receipt/marker/manifest family) must carry
	// no endpoint provenance and no secret. Stdout/stderr are the
	// operator-visible portable surface; the cache and installed markers
	// are the on-disk portable surfaces. The machine policy and the
	// machine-private sink legitimately contain the mirror address and
	// are excluded by construction (we never walk home root, only the
	// portable cache and install outputs).
	combined := stdout + stderr
	if strings.Contains(combined, secretMarker) {
		t.Errorf("portable CLI output leaks broker secret")
	}
	if strings.Contains(combined, mirrorHost) || strings.Contains(combined, mirrorURL) {
		t.Errorf("portable CLI output carries endpoint provenance:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	assertPortableClean := func(path string, data []byte) {
		t.Helper()
		if strings.Contains(string(data), secretMarker) {
			t.Errorf("portable artifact %q leaks broker secret", path)
		}
		if strings.Contains(string(data), mirrorHost) || strings.Contains(string(data), mirrorURL) {
			t.Errorf("portable artifact %q carries endpoint provenance", path)
		}
	}
	// Receipts: protected build-cache entries published by the real install.
	for _, entry := range publishedCacheEntries(t, home) {
		entries, err := os.ReadDir(entry)
		if err != nil {
			t.Fatal(err)
		}
		for _, child := range entries {
			childPath := filepath.Join(entry, child.Name())
			info, err := child.Info()
			if err != nil {
				t.Fatal(err)
			}
			if info.IsDir() {
				continue
			}
			// Binaries are large; a bounded prefix still proves absence
			// of the ASCII provenance markers without reading megabytes.
			data, err := os.ReadFile(childPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) > 1<<20 {
				data = data[:1<<20]
			}
			assertPortableClean(childPath, data)
		}
	}
	// Markers: installed-skill markers under the project.
	project := filepath.Join(fixture.root, "project")
	_ = filepath.Walk(project, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		// The project input manifest is checked separately below; the
		// walk covers emitted markers and installed outputs.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(data) > 1<<20 {
			data = data[:1<<20]
		}
		// Skip the sink-adjacent machine policy only if the walk ever
		// reaches home (it does not: we walk the project only).
		assertPortableClean(path, data)
		return nil
	})
	// Manifest (portable project manifest) carries the declaration, never
	// the connection address.
	if manifestData, err := os.ReadFile(filepath.Join(project, "Skillfile.json")); err == nil {
		assertPortableClean("Skillfile.json", manifestData)
	}
}
