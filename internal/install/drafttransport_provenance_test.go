package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/gitcred"
)

// Revision-2 provenance sink end to end (repository-transport §7):
// the production sink construction (DraftTransportProvenanceTrace, the
// same function productionExternalDeps assigns) carries a
// declared-mirror acquisition's canonical identity with its listed and
// resolved mirror provenance into the manager-home operation
// diagnostics log, broker secrets never appear there, and portable
// content carries none of those properties.
func TestAcquireDraftNetworkRevision2ProvenanceSink(t *testing.T) {
	readSinkLines := func(t *testing.T, home string) []map[string]any {
		t.Helper()
		payload, err := os.ReadFile(DraftTransportProvenancePath(home))
		if err != nil {
			t.Fatalf("read provenance sink: %v", err)
		}
		var lines []map[string]any
		for _, raw := range strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n") {
			var line map[string]any
			if err := json.Unmarshal([]byte(raw), &line); err != nil {
				t.Fatalf("sink line is not JSON: %q: %v", raw, err)
			}
			lines = append(lines, line)
		}
		return lines
	}
	assertOnlySinkFile := func(t *testing.T, home string) {
		t.Helper()
		entries, err := os.ReadDir(home)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 || entries[0].Name() != DraftTransportProvenanceFileName {
			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				names = append(names, entry.Name())
			}
			t.Fatalf("manager home holds %q, want only the operation-diagnostics sink", names)
		}
	}

	t.Run("mirror provenance reaches the production sink", func(t *testing.T) {
		bare, commit := draftBareFixture(t)
		lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}
		home := t.TempDir()
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https",
			`"mirror_of":"fixture.test/repository"`), ""), ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("mirror-https"))
		deps := ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policy,
			DraftProvidersPath:  providers,
			DraftTransportTrace: DraftTransportProvenanceTrace(home),
			Audit:               func(context.Context, buildrepo.AuditSubject) error { return nil }}
		snapshot, err := v2acquire(t, deps, tool, lock)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1", len(fetches))
		}
		lines := readSinkLines(t, home)
		if len(lines) != 1 {
			t.Fatalf("sink lines = %v, want one provenance record", lines)
		}
		line := lines[0]
		for key, want := range map[string]any{
			"identity": "fixture.test/repository", "url": draftMirrorHTTPS,
			"resolved_host": "mirror.fixture.test", "mirror_of": "fixture.test/repository",
		} {
			if line[key] != want {
				t.Errorf("sink[%q] = %v, want %v (line %v)", key, line[key], want, line)
			}
		}
		if line["alias"] != nil {
			t.Errorf("sink carries alias = %v, want none", line["alias"])
		}
		// Portable content proves the locked commit, never the connection
		// address: the snapshot carries no endpoint properties.
		if dump := fmt.Sprintf("%+v", snapshot); strings.Contains(dump, "mirror.fixture.test") ||
			strings.Contains(dump, "mirror_of") || strings.Contains(dump, draftMirrorHTTPS) {
			t.Errorf("portable snapshot carries provenance: %+v", snapshot)
		}
		assertOnlySinkFile(t, home)
	})

	t.Run("secrets stay out of the sink while errors stay closed", func(t *testing.T) {
		lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}
		home := t.TempDir()
		tool, _ := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS:  {stderr: draftMirrorDNSStderr},
			draftHTTPSPrimary: {stderr: draftDNSStderr},
		})
		policy := draftPolicyFile(t, v2policy(`{"endpoints":[`+
			v2mirrorEndpoint(draftMirrorHTTPS, "prov-a", `"mirror_of":"fixture.test/repository"`)+`,`+
			v2mirrorEndpoint(draftHTTPSPrimary, "prov-b", "")+`],"fallback":"availability-auth"}`, ""))
		providers := draftProvidersFile(t, `{"schema_version":1,"providers":{`+
			`"prov-a":{"https":{"username":"a"}},"prov-b":{"https":{"username":"b"}}}}`)
		const secretMarker = "zz-secret-marker-260916-sink"
		reader := &draftProviderReader{secrets: map[string]gitcred.HostCredential{
			"prov-a": {Username: "a", Secret: secretMarker + "-a"},
			"prov-b": {Username: "b", Secret: secretMarker + "-b"},
		}}
		deps := ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policy,
			DraftProvidersPath: providers, DraftProviderReader: reader,
			DraftTransportTrace: DraftTransportProvenanceTrace(home),
			Audit:               func(context.Context, buildrepo.AuditSubject) error { return nil }}
		_, err := v2acquire(t, deps, tool, lock)
		if buildrepo.ErrorCode(err) != buildrepo.CodeRepositoryEndpointUnavailable {
			t.Fatalf("err = %v, want %s", err, buildrepo.CodeRepositoryEndpointUnavailable)
		}
		payload, readErr := os.ReadFile(DraftTransportProvenancePath(home))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if lines := readSinkLines(t, home); len(lines) != 2 {
			t.Fatalf("sink lines = %d, want two provenance records", len(lines))
		}
		if strings.Contains(string(payload), secretMarker) {
			t.Errorf("provenance sink contains broker secret material")
		}
		if message := err.Error(); strings.Contains(message, secretMarker) ||
			strings.Contains(message, "https://") || strings.Contains(message, "mirror.fixture.test") {
			t.Errorf("exhaustion %q leaks endpoint or secret detail", message)
		}
		assertOnlySinkFile(t, home)
	})

	t.Run("nil trace records nothing and changes no verdict", func(t *testing.T) {
		bare, commit := draftBareFixture(t)
		lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}
		home := t.TempDir()
		tool, _ := draftGitTool(t, map[string]draftGitArm{
			draftMirrorHTTPS: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, v2policy(v2entry(v2mirrorEndpoint(draftMirrorHTTPS, "mirror-https",
			`"mirror_of":"fixture.test/repository"`), ""), ""))
		providers := draftProvidersFile(t, v2anonymousProvidersDoc("mirror-https"))
		deps := ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policy,
			DraftProvidersPath: providers, DraftTransportTrace: nil,
			Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
		snapshot, err := v2acquire(t, deps, tool, lock)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		// The positive subtests above fail without a wired sink: the
		// sink file is their only provenance source.
		if _, statErr := os.Stat(DraftTransportProvenancePath(home)); !os.IsNotExist(statErr) {
			t.Fatalf("nil trace wrote %q", DraftTransportProvenancePath(home))
		}
	})

	t.Run("empty home records nothing", func(t *testing.T) {
		if trace := DraftTransportProvenanceTrace(""); trace != nil {
			t.Fatal("empty-home trace is non-nil, want fail-closed nil")
		}
	})

	t.Run("sink path stays under the manager home", func(t *testing.T) {
		home := t.TempDir()
		want := filepath.Join(home, DraftTransportProvenanceFileName)
		if got := DraftTransportProvenancePath(home); got != want {
			t.Fatalf("sink path = %q, want %q", got, want)
		}
	})
}
