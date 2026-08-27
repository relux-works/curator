package scopes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/marker"
)

type testHomeLock struct{ err error }

func (lock testHomeLock) AssertHeld() error { return lock.err }

// recordingCache stands in for the protected store so a scopes test can assert
// exactly which references the mark phase produced without publishing real
// protected artifacts.
type recordingCache struct {
	calls      int
	referenced []string
	result     buildcache.SweepResult
	err        error
}

func (cache *recordingCache) Sweep(request buildcache.SweepRequest, lock buildcache.HomeLock) (buildcache.SweepResult, error) {
	if lock == nil {
		return buildcache.SweepResult{}, fmt.Errorf("sweep was called without a lock witness")
	}
	if err := lock.AssertHeld(); err != nil {
		return buildcache.SweepResult{}, err
	}
	cache.calls++
	cache.referenced = append([]string(nil), request.Referenced...)
	sort.Strings(cache.referenced)
	return cache.result, cache.err
}

func buildKey(seed string) buildmeta.CacheKey {
	return buildmeta.CacheKey("sha256:" + strings.Repeat(seed, 64))
}

// installBuildMarker writes a valid marker v2 that references one build key.
func installBuildMarker(t *testing.T, skillsDir, name, commit, command string, key buildmeta.CacheKey) {
	t.Helper()
	installBuildMarkerForBand(t, skillsDir, name, commit, command, key, 6)
}

// installBuildMarkerForBand writes, through the real marker writer, the marker
// an installation of a skill in the given manifest band leaves behind. The
// schema is chosen by the writer, never by the test.
func installBuildMarkerForBand(t *testing.T, skillsDir, name, commit, command string,
	key buildmeta.CacheKey, skillSchema int) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	hash, _ := hashing.ContentSHA256(dir, nil)
	artifactPath, err := buildmeta.ArtifactPath(command, "linux")
	if err != nil {
		t.Fatal(err)
	}
	if err := marker.Write(dir, &marker.Marker{
		Name: name, Source: name, RefKind: "tag", Ref: "v1",
		Commit: commit, ContentSHA256: hash,
		Agents: []string{}, Commands: []string{command}, Dependencies: []string{},
		SkillSchemaVersion: skillSchema, BuildRoots: []string{"build"},
		BuildSource: &buildsource.Identity{
			Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64),
		},
		InstalledAt: "2026-07-20T00:00:00Z",
		Builds:      map[string]marker.Build{command: buildRecordForBand(key, artifactPath, skillSchema)},
	}); err != nil {
		t.Fatal(err)
	}
}

// buildRecordForBand carries the per-schema fields marker v3 and v4 require of
// a local go-v1 command. Nothing the mark phase reads changes with the band, so
// a band difference can only be observed through the schema banding under test.
func buildRecordForBand(key buildmeta.CacheKey, artifactPath string, skillSchema int) marker.Build {
	build := marker.Build{
		Driver: buildmeta.DriverGoV1, CacheKey: key,
		ReceiptSHA256:  buildmeta.ReceiptHash("sha256:" + strings.Repeat("d", 64)),
		ArtifactSHA256: "sha256:" + strings.Repeat("e", 64),
		ArtifactPath:   artifactPath,
	}
	if skillSchema >= 7 {
		build.ExecutionPolicy = buildmeta.ExecutionPolicy
		build.ReceiptSchemaVersion = 1
	}
	return build
}

// TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema proves the mark
// phase keeps a live build reference from every marker schema that can record
// one, including the marker v4 a schema-8 installation writes.
//
// A schema missing from that band is not "this installation has no builds": its
// recorded keys go unmarked, the sweep sees them as unreferenced, and the
// collector deletes cache entries the installation is still running from. The
// schema-1 case is the other side of the bound -- it genuinely records no build
// and must contribute nothing -- so the band cannot be widened to accept
// anything either.
func TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	skillsDir := filepath.Join(project, ".agents", "skills")

	bands := []struct {
		skillSchema  int
		markerSchema int
		key          buildmeta.CacheKey
	}{
		{skillSchema: 6, markerSchema: marker.SchemaVersion, key: buildKey("1")},
		{skillSchema: 7, markerSchema: marker.ExternalSchemaVersion, key: buildKey("2")},
		{skillSchema: 8, markerSchema: marker.PolicySchemaVersion, key: buildKey("3")},
	}
	want := make([]string, 0, len(bands))
	for index, band := range bands {
		name := fmt.Sprintf("skill-band-%d", band.skillSchema)
		commit := strings.Repeat(fmt.Sprintf("%d", index+4), 40)
		installBuildMarkerForBand(t, skillsDir, name, commit, fmt.Sprintf("tool-%d", band.skillSchema),
			band.key, band.skillSchema)

		// The writer really did produce the schema this band is about, so a
		// marked key cannot come from silently falling back to schema 2.
		written := marker.Read(filepath.Join(skillsDir, name))
		if written == nil {
			t.Fatalf("skill schema %d: the marker writer produced a marker its own reader refuses", band.skillSchema)
		}
		if written.SchemaVersion != band.markerSchema {
			t.Fatalf("skill schema %d wrote marker schema %d, want %d",
				band.skillSchema, written.SchemaVersion, band.markerSchema)
		}
		want = append(want, string(band.key))
	}

	// A schema-1 installation records no build at all and must add no key.
	legacyCommit := strings.Repeat("9", 40)
	installMarker(t, skillsDir, "skill-legacy", legacyCommit)
	downgradeToLegacyMarker(t, filepath.Join(skillsDir, "skill-legacy"))

	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	cache := &recordingCache{}
	if _, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(want)
	if strings.Join(cache.referenced, ",") != strings.Join(want, ",") {
		t.Fatalf("referenced = %v, want %v", cache.referenced, want)
	}
}

// TestAbsorbKeysBuildLivenessOnTheSchemaBandNotOnAPopulatedBuildsMap pins the
// delete direction of the mark band, which no marker on disk can exercise: the
// reader already refuses a schema-1 document that carries build fields, so a
// mark phase with the band removed still behaves correctly through Collect.
//
// The band is therefore asserted against the struct directly. Liveness is keyed
// on the schema that gives `builds` its meaning, not on the map merely being
// non-empty, so a document from outside the build-bearing band can never
// contribute a reference that protects a cache entry.
func TestAbsorbKeysBuildLivenessOnTheSchemaBandNotOnAPopulatedBuildsMap(t *testing.T) {
	builds := map[string]marker.Build{"tool": {
		Driver: buildmeta.DriverGoV1, CacheKey: buildKey("1"),
	}}
	for _, testCase := range []struct {
		name   string
		schema int
		want   int
	}{
		{name: "schema 1 predates the build record", schema: marker.LegacySchemaVersion, want: 0},
		{name: "schema 0 is not a readable schema", schema: 0, want: 0},
		{name: "a schema past the readable band", schema: marker.NewestSchemaVersion + 1, want: 0},
		{name: "schema 2 records builds", schema: marker.SchemaVersion, want: 1},
		{name: "schema 3 records builds", schema: marker.ExternalSchemaVersion, want: 1},
		{name: "schema 4 records builds", schema: marker.PolicySchemaVersion, want: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			marked := &marks{runtime: map[string]bool{}}
			marked.absorb(scopeMarks{markers: []*marker.Marker{{
				SchemaVersion: testCase.schema, Name: "skill",
				Commit: strings.Repeat("a", 40), Builds: builds,
			}}})
			if len(marked.builds) != testCase.want {
				t.Fatalf("schema %d marked %v, want %d reference(s)",
					testCase.schema, marked.builds, testCase.want)
			}
			// Runtime liveness never depends on the band: an installation of any
			// schema keeps its runtime tree.
			if !marked.runtime["skill/"+strings.Repeat("a", 40)] {
				t.Fatalf("schema %d lost its runtime reference", testCase.schema)
			}
		})
	}
}

// TestCollectMarksBuildKeysFromEveryLiveScope proves the mark phase covers
// project, global, and hybrid installs plus in-flight journals.
func TestCollectMarksBuildKeysFromEveryLiveScope(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	projectKey := buildKey("1")
	globalKey := buildKey("2")
	hybridKey := buildKey("3")
	journalKey := buildKey("4")

	installBuildMarker(t, filepath.Join(project, ".agents", "skills"), "skill-p", strings.Repeat("a", 40), "p-tool", projectKey)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	installBuildMarker(t, filepath.Join(home, "global", "skills"), "skill-g", strings.Repeat("b", 40), "g-tool", globalKey)
	installBuildMarker(t, HybridSkillsRoot(home), "skill-h", strings.Repeat("c", 40), "h-tool", hybridKey)

	cache := &recordingCache{result: buildcache.SweepResult{Removed: []string{string(buildKey("9"))}}}
	result, err := Collect(MaintenanceRequest{
		Home: home, Lock: testHomeLock{}, Cache: cache,
		JournalKeys: []string{string(journalKey)},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{string(globalKey), string(hybridKey), string(journalKey), string(projectKey)}
	sort.Strings(want)
	if strings.Join(cache.referenced, ",") != strings.Join(want, ",") {
		t.Fatalf("referenced = %v, want %v", cache.referenced, want)
	}
	if len(result.RemovedBuilds) != 1 || result.RemovedBuilds[0] != string(buildKey("9")) {
		t.Fatalf("removed builds = %v", result.RemovedBuilds)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("a clean pass warned: %v", result.Warnings)
	}
}

// TestCollectMarksRuntimeFromEverySupportedMarkerSchema proves a still-current
// schema-1 installation keeps its runtime entry while contributing no build
// reference, which is what the manager profile requires.
func TestCollectMarksRuntimeFromEverySupportedMarkerSchema(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	legacyCommit := strings.Repeat("1", 40)
	buildCommit := strings.Repeat("2", 40)
	skillsDir := filepath.Join(project, ".agents", "skills")

	installMarker(t, skillsDir, "skill-legacy", legacyCommit)
	downgradeToLegacyMarker(t, filepath.Join(skillsDir, "skill-legacy"))
	installBuildMarker(t, skillsDir, "skill-build", buildCommit, "tool", buildKey("7"))
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{"skill-legacy/" + legacyCommit, "skill-build/" + buildCommit, "skill-x/" + strings.Repeat("3", 40)} {
		if err := os.MkdirAll(filepath.Join(home, "runtime", filepath.FromSlash(entry)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	cache := &recordingCache{}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-legacy", legacyCommit)); err != nil {
		t.Fatal("a valid marker-v1 installation lost its runtime entry")
	}
	if _, err := os.Stat(filepath.Join(home, "runtime", "skill-build", buildCommit)); err != nil {
		t.Fatal("a valid marker-v2 installation lost its runtime entry")
	}
	if len(result.RemovedRuntime) != 1 {
		t.Fatalf("removed runtime = %v", result.RemovedRuntime)
	}
	if len(cache.referenced) != 1 || cache.referenced[0] != string(buildKey("7")) {
		t.Fatalf("marker v1 contributed a build reference: %v", cache.referenced)
	}
}

// TestCollectSkipsTheBuildSweepOnUnprovableReferences proves an unreadable
// marker, skill directory, or consumer registry retains every cache entry and
// reports the uncertainty.
func TestCollectSkipsTheBuildSweepOnUnprovableReferences(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T, home, project string)
		warning string
	}{
		{
			name: "invalid marker",
			arrange: func(t *testing.T, _, project string) {
				dir := filepath.Join(project, ".agents", "skills", "skill-broken")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, marker.Name), []byte("{not a marker}"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "is unreadable or invalid",
		},
		{
			name: "corrupt consumer registry",
			arrange: func(t *testing.T, home, _ string) {
				if err := os.WriteFile(filepath.Join(home, ConsumersName), []byte("{"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			warning: "consumer registry",
		},
		{
			name: "unreadable global skills directory",
			arrange: func(t *testing.T, home, _ string) {
				dir := filepath.Join(home, "global", "skills")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(dir, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
				if _, err := os.ReadDir(dir); err == nil {
					t.Skip("this environment can read a mode-000 directory")
				}
			},
			warning: "is unreadable",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			project := t.TempDir()
			installBuildMarker(t, filepath.Join(project, ".agents", "skills"), "skill-p", strings.Repeat("a", 40), "tool", buildKey("1"))
			if err := RecordConsumer(home, project); err != nil {
				t.Fatal(err)
			}
			test.arrange(t, home, project)

			cache := &recordingCache{}
			result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
			if err != nil {
				t.Fatal(err)
			}
			if cache.calls != 0 {
				t.Fatal("an unprovable reference set still drove a build sweep")
			}
			if !warned(result, test.warning) {
				t.Fatalf("missing warning %q: %v", test.warning, result.Warnings)
			}
			if !warned(result, "build cache sweep skipped") {
				t.Fatalf("the skipped sweep was not reported: %v", result.Warnings)
			}
		})
	}
}

// TestCollectRequiresTheHomeLock proves maintenance never prunes consumers or
// sweeps anything without a held lock witness.
func TestCollectRequiresTheHomeLock(t *testing.T) {
	home := t.TempDir()
	dead := filepath.Join(t.TempDir(), "gone")
	if err := RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	cache := &recordingCache{}
	for _, lock := range []HomeLock{nil, testHomeLock{err: os.ErrClosed}} {
		if _, err := Collect(MaintenanceRequest{Home: home, Lock: lock, Cache: cache}); err == nil {
			t.Fatalf("collect accepted lock witness %v", lock)
		}
	}
	if _, err := Collect(MaintenanceRequest{Lock: testHomeLock{}, Cache: cache}); err == nil {
		t.Fatal("collect accepted an empty manager home")
	}
	if cache.calls != 0 {
		t.Fatal("an unlocked pass swept the build cache")
	}
	if consumers := LoadConsumers(home); len(consumers) != 1 {
		t.Fatalf("an unlocked pass pruned consumers: %v", consumers)
	}
}

// TestCollectPrunesConsumersInsideTheSamePass proves consumer pruning is part
// of the serialized maintenance transaction, not a separate unlocked write.
func TestCollectPrunesConsumersInsideTheSamePass(t *testing.T) {
	home := t.TempDir()
	dead := filepath.Join(t.TempDir(), "gone")
	if err := RecordConsumer(home, dead); err != nil {
		t.Fatal(err)
	}
	live := t.TempDir()
	installMarker(t, filepath.Join(live, ".agents", "skills"), "skill-a", strings.Repeat("1", 40))
	if err := RecordConsumer(home, live); err != nil {
		t.Fatal(err)
	}

	cache := &recordingCache{}
	if _, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache}); err != nil {
		t.Fatal(err)
	}
	consumers := LoadConsumers(home)
	if len(consumers) != 1 || consumers[0] != live {
		t.Fatalf("consumers = %v, want only %s", consumers, live)
	}
	if cache.calls != 1 {
		t.Fatalf("the build sweep ran %d times in one pass", cache.calls)
	}
}

// TestCollectReportsSweepFailuresWithoutLosingRuntimeWork proves a cache-side
// failure is surfaced while the runtime and consumer work of the same pass
// still stands.
func TestCollectReportsSweepFailuresWithoutLosingRuntimeWork(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	installMarker(t, filepath.Join(project, ".agents", "skills"), "skill-a", strings.Repeat("1", 40))
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "runtime", "skill-x", strings.Repeat("3", 40)), 0o755); err != nil {
		t.Fatal(err)
	}

	cache := &recordingCache{
		err:    fmt.Errorf("injected sweep failure"),
		result: buildcache.SweepResult{Warnings: []string{"retained something"}},
	}
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}, Cache: cache})
	if err == nil {
		t.Fatal("a sweep failure was swallowed")
	}
	if len(result.RemovedRuntime) != 1 {
		t.Fatalf("runtime work was lost: %v", result.RemovedRuntime)
	}
	if !warned(result, "retained something") {
		t.Fatalf("sweep warnings were dropped: %v", result.Warnings)
	}
}

// TestCollectForwardsTheSweepClockAndGrace proves the caller-controlled
// retention window reaches the protected store unchanged.
func TestCollectForwardsTheSweepClockAndGrace(t *testing.T) {
	home := t.TempDir()
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	var seen buildcache.SweepRequest
	cache := &forwardingCache{observe: func(request buildcache.SweepRequest) { seen = request }}
	if _, err := Collect(MaintenanceRequest{
		Home: home, Lock: testHomeLock{}, Cache: cache, Now: now, Grace: 3 * time.Hour,
	}); err != nil {
		t.Fatal(err)
	}
	if !seen.Now.Equal(now) || seen.Grace != 3*time.Hour {
		t.Fatalf("sweep request = %+v", seen)
	}
}

// TestCollectResolvesTheRealStoreByDefault proves the zero request reaches a
// real protected store instead of silently doing nothing.
func TestCollectResolvesTheRealStoreByDefault(t *testing.T) {
	home := t.TempDir()
	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedBuilds) != 0 {
		t.Fatalf("an empty home removed build entries: %v", result.RemovedBuilds)
	}
	if _, err := os.Stat(filepath.Join(home, "cache")); err == nil {
		t.Fatal("maintenance created cache state in an empty home")
	}
}

type forwardingCache struct {
	observe func(buildcache.SweepRequest)
}

func (cache *forwardingCache) Sweep(request buildcache.SweepRequest, _ buildcache.HomeLock) (buildcache.SweepResult, error) {
	cache.observe(request)
	return buildcache.SweepResult{}, nil
}

func warned(result MaintenanceResult, needle string) bool {
	for _, warning := range result.Warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
}

// downgradeToLegacyMarker rewrites a written marker as the schema-1 shape the
// manager must still read: no build state, and the legacy schema version.
func downgradeToLegacyMarker(t *testing.T, installedDir string) {
	t.Helper()
	path := filepath.Join(installedDir, marker.Name)
	payload, err := os.ReadFile(path) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"build_roots", "build_source", "builds"} {
		delete(fields, field)
	}
	fields["schema_version"] = json.RawMessage(fmt.Sprintf("%d", marker.LegacySchemaVersion))
	legacy, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, legacy, 0o644); err != nil {
		t.Fatal(err)
	}
	if recorded := marker.Read(installedDir); recorded == nil || recorded.SchemaVersion != marker.LegacySchemaVersion {
		t.Fatalf("the downgraded marker is not a valid schema-1 marker: %+v", recorded)
	}
}
