//go:build curator_cold_repository

package godriver

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildsource"
)

// The opt-in witness drives the production frozen snapshot and worker with a
// real repository and an initially empty operation-private Go cache. Durations
// are observations, not portable speed guarantees. No source is downloaded.
func TestColdRepositoryMeasurement(t *testing.T) {
	source := os.Getenv("CURATOR_COLD_REPOSITORY")
	if source == "" {
		t.Fatal("CURATOR_COLD_REPOSITORY must name a real vendored repository")
	}
	command := os.Getenv("CURATOR_COLD_COMMAND")
	if command == "" {
		command = "tool"
	}
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	if err := os.CopyFS(filepath.Join(snapshot, "build"), os.DirFS(source)); err != nil {
		t.Fatal(err)
	}
	snapshot = mustPhysical(t, snapshot)
	token, err := buildsource.Validate(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer token.Close()
	config := ConfigFromEnvironment(t.TempDir(), snapshot)
	config.CuratorGo = filepath.Join(build.Default.GOROOT, "bin", platformGoName)
	config.observePhase = func(phase string, elapsed time.Duration) { t.Logf("cold phase=%s elapsed=%s", phase, elapsed) }
	t.Logf("measurement scheduling band=%s (unset means unclassified)", os.Getenv("CURATOR_COLD_SCHEDULING_BAND"))
	session, err := Establish(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	session.observePhase = func(phase string, elapsed time.Duration) { t.Logf("cold phase=%s elapsed=%s", phase, elapsed) }
	started := time.Now()
	result, err := Build(context.Background(), BuildRequest{Session: session, Source: token,
		CommandObject: BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/" + command},
		BuildRoot:     "build", SourceDir: "build/cmd/" + command, Command: command})
	t.Logf("cold whole-build elapsed=%s diagnostic=%s error=%v", time.Since(started), DiagnosticCode(err), err)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("artifact size=%d digest=%s", result.Artifact.Metadata.Size, result.Artifact.Metadata.SHA256)
}
