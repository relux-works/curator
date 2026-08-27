package scopes

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/marker"
)

// TestCollectSweepsOnlyUnreferencedProtectedEntries drives the whole chain —
// mark from real markers, sweep the real protected store — instead of a stub,
// so the logical key a marker records and the key the cache stores must agree.
func TestCollectSweepsOnlyUnreferencedProtectedEntries(t *testing.T) {
	home := protectedTestHome(t)
	store, err := buildcache.New(home)
	if err != nil {
		t.Fatal(err)
	}
	referenced := publishRealEntry(t, store, "kept-tool", "kept artifact")
	orphan := publishRealEntry(t, store, "orphan-tool", "orphan artifact")

	project := t.TempDir()
	installBuildMarker(t, filepath.Join(project, ".agents", "skills"), "skill-p",
		strings.Repeat("a", 40), "kept-tool", referenced)
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}
	backdateEntry(t, home, referenced, 30*24*time.Hour)
	backdateEntry(t, home, orphan, 30*24*time.Hour)

	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("a clean pass warned: %v", result.Warnings)
	}
	if len(result.RemovedBuilds) != 1 || result.RemovedBuilds[0] != string(orphan) {
		t.Fatalf("removed builds = %v, want only %s", result.RemovedBuilds, orphan)
	}
	if _, err := os.Lstat(realEntryPath(home, referenced)); err != nil {
		t.Fatalf("a marker-referenced protected entry was swept: %v", err)
	}
	if _, err := os.Lstat(realEntryPath(home, orphan)); err == nil {
		t.Fatal("an unreferenced protected entry older than grace survived")
	}
	// The removal is atomic: no partially deleted tree is left behind.
	names, err := os.ReadDir(filepath.Join(home, "cache", "build", buildmeta.DriverGoV1))
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 {
		t.Fatalf("cache root after the sweep = %v", names)
	}
}

// TestCollectRetainsAJournalOwnedEntry proves an artifact an in-flight
// transaction still depends on survives even with no marker naming it.
func TestCollectRetainsAJournalOwnedEntry(t *testing.T) {
	home := protectedTestHome(t)
	store, err := buildcache.New(home)
	if err != nil {
		t.Fatal(err)
	}
	inFlight := publishRealEntry(t, store, "in-flight-tool", "in-flight artifact")
	backdateEntry(t, home, inFlight, 30*24*time.Hour)

	result, err := Collect(MaintenanceRequest{
		Home: home, Lock: testHomeLock{}, JournalKeys: []string{string(inFlight)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedBuilds) != 0 {
		t.Fatalf("a journal-owned entry was swept: %v", result.RemovedBuilds)
	}
	if _, err := os.Lstat(realEntryPath(home, inFlight)); err != nil {
		t.Fatalf("a journal-owned entry was removed: %v", err)
	}
}

// TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead proves the fail-safe
// path end to end against the real store: one unreadable marker retains every
// cache entry rather than sweeping artifacts it can no longer account for.
func TestCollectRetainsBuildEntriesWhenAMarkerCannotBeRead(t *testing.T) {
	home := protectedTestHome(t)
	store, err := buildcache.New(home)
	if err != nil {
		t.Fatal(err)
	}
	orphan := publishRealEntry(t, store, "orphan-tool", "orphan artifact")
	backdateEntry(t, home, orphan, 30*24*time.Hour)

	project := t.TempDir()
	broken := filepath.Join(project, ".agents", "skills", "skill-broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, marker.Name), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RecordConsumer(home, project); err != nil {
		t.Fatal(err)
	}

	result, err := Collect(MaintenanceRequest{Home: home, Lock: testHomeLock{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RemovedBuilds) != 0 {
		t.Fatalf("an unprovable reference set still swept: %v", result.RemovedBuilds)
	}
	if _, err := os.Lstat(realEntryPath(home, orphan)); err != nil {
		t.Fatalf("an entry was removed despite an unreadable marker: %v", err)
	}
	if !warned(result, "is unreadable or invalid") {
		t.Fatalf("the unreadable marker was not reported: %v", result.Warnings)
	}
}

func realEntryPath(home string, key buildmeta.CacheKey) string {
	return filepath.Join(home, "cache", "build", buildmeta.DriverGoV1,
		strings.TrimPrefix(string(key), "sha256:"))
}

func backdateEntry(t *testing.T, home string, key buildmeta.CacheKey, age time.Duration) {
	t.Helper()
	when := time.Now().Add(-age)
	if err := os.Chtimes(realEntryPath(home, key), when, when); err != nil {
		t.Fatal(err)
	}
}

// publishRealEntry publishes one protected winner through the real store.
func publishRealEntry(t *testing.T, store *buildcache.Store, command, artifact string) buildmeta.CacheKey {
	t.Helper()
	input := realBuildInput(command)
	key, err := input.CacheKey()
	if err != nil {
		t.Fatal(err)
	}
	artifactPath, err := buildmeta.ArtifactPath(command, input.Target.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(artifact))
	receipt, err := buildmeta.NewReceipt(input, buildmeta.Artifact{
		Path: artifactPath, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(artifact)),
	})
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	sourceDir := filepath.Join(store.Home(), "private-builds")
	if err := os.MkdirAll(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(sourceDir, command)
	if err := os.WriteFile(source, []byte(artifact), 0o600); err != nil {
		t.Fatal(err)
	}
	published, err := store.Publish(buildcache.Publication{
		Input: input, ReceiptBytes: receiptBytes, ArtifactSource: source,
	}, testHomeLock{})
	if err != nil {
		t.Skipf("this platform cannot publish a protected cache entry: %v", err)
	}
	if published.ArtifactPath == "" {
		t.Fatalf("publication = %+v", published)
	}
	return key
}

func realBuildInput(command string) buildmeta.Input {
	tuning := map[string]string{}
	switch runtime.GOARCH {
	case "amd64":
		tuning["GOAMD64"] = "v1"
	case "arm64":
		tuning["GOARM64"] = "v8.0"
	case "386":
		tuning["GO386"] = "sse2"
	}
	return buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource: buildsource.Identity{
			Algorithm: buildsource.Algorithm, ContentSHA256: "sha256:" + strings.Repeat("b", 64),
		},
		BuildRoot: "build",
		Command:   command,
		SourceDir: "build/cmd/" + command,
		Target:    buildmeta.Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Tuning: tuning},
		Toolchain: buildmeta.Toolchain{
			Algorithm: buildmeta.ToolchainAlgorithm, GoRelpath: buildmeta.ToolchainGoRelpath,
			GoVersion:     "go version go1.26.1 " + runtime.GOOS + "/" + runtime.GOARCH,
			ContentSHA256: "sha256:" + strings.Repeat("c", 64),
		},
		Policy: buildmeta.FixedPolicy(),
	}
}
