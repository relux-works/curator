package crossconformance

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
)

// External-build fixtures for the conformance suites: a fake toolchain,
// session, and builder implementing the production BuildDeps interfaces
// (mirroring the in-package fakes), a proved external snapshot, and a
// schema-7 skill declaring one external go-repository-v1 command.

const xbExternalLockedCommit = "0123456789abcdef0123456789abcdef01234567"

type xbSession struct {
	target    buildmeta.Target
	toolchain buildmeta.Toolchain
	root      string
}

func (s *xbSession) Target() buildmeta.Target       { return s.target }
func (s *xbSession) Toolchain() buildmeta.Toolchain { return s.toolchain }

func (s *xbSession) VerifyToolchain(context.Context) error { return nil }

func (s *xbSession) Release() error { return os.RemoveAll(s.root) }

type xbToolchain struct {
	t         *testing.T
	target    buildmeta.Target
	toolchain buildmeta.Toolchain
}

func (tc *xbToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	return tc.target, tc.toolchain, nil
}

func (tc *xbToolchain) Establish(context.Context) (install.BuildSession, error) {
	return &xbSession{target: tc.target, toolchain: tc.toolchain, root: tc.t.TempDir()}, nil
}

type xbBuilder struct {
	t     *testing.T
	calls []string
}

func (b *xbBuilder) Stage(_ context.Context, request install.StageRequest) (install.StagedArtifact, error) {
	b.t.Helper()
	b.calls = append(b.calls, request.Command)
	session, ok := request.Session.(*xbSession)
	if !ok {
		return install.StagedArtifact{}, errStagingOutsideTrustedSession
	}
	relative, err := buildmeta.ArtifactPath(request.Command, session.target.GOOS)
	if err != nil {
		return install.StagedArtifact{}, err
	}
	path := filepath.Join(session.root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		b.t.Fatal(err)
	}
	payload := []byte("artifact:" + request.Command)
	if err := os.WriteFile(path, payload, 0o700); err != nil {
		b.t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	metadata := buildmeta.Artifact{
		Path: relative, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(payload)),
	}
	if request.ExpectedBuildInputSHA256 != "" {
		executionReceipt, err := closureexec.NewPortableBuildSessionReceiptForDigest(closuregraph.ID(request.ExpectedBuildInputSHA256), session.toolchain, metadata, nil)
		if err != nil {
			return install.StagedArtifact{}, err
		}
		return install.StagedArtifact{Path: path, Metadata: metadata, ExecutionReceipt: executionReceipt}, nil
	}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion, Package: request.Package, Driver: buildmeta.DriverGoV1,
		BuildSource: request.Source.Identity(), BuildRoot: request.BuildRoot,
		Command: request.Command, SourceDir: request.SourceDir,
		Target: session.target, Toolchain: session.toolchain, Policy: buildmeta.FixedPolicy(),
	}
	executionReceipt, err := closureexec.NewPortableBuildSessionReceipt(input, metadata, nil)
	if err != nil {
		return install.StagedArtifact{}, err
	}
	return install.StagedArtifact{Path: path, Metadata: metadata, ExecutionReceipt: executionReceipt}, nil
}

type xbStaticError string

func (e xbStaticError) Error() string { return string(e) }

const errStagingOutsideTrustedSession = xbStaticError("staging outside a trusted session")

type xbClock struct{ at time.Time }

func (c xbClock) Now() time.Time { return c.at }

type xbGeneration struct{}

func (g *xbGeneration) InstalledMarker(installedDir string) (*marker.Marker, error) {
	recorded, _, err := marker.ReadState(installedDir)
	return recorded, err
}

func xbTestTarget() buildmeta.Target {
	target := buildmeta.Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Tuning: map[string]string{}}
	var key string
	switch runtime.GOARCH {
	case "386":
		key = "GO386"
	case "amd64":
		key = "GOAMD64"
	case "arm":
		key = "GOARM"
	case "arm64":
		key = "GOARM64"
	case "ppc64", "ppc64le":
		key = "GOPPC64"
	case "riscv64":
		key = "GORISCV64"
	}
	if key != "" {
		target.Tuning[key] = "v1"
	}
	return target
}

func xbTestToolchain() buildmeta.Toolchain {
	return buildmeta.Toolchain{
		Algorithm:     buildmeta.ToolchainAlgorithm,
		GoRelpath:     buildmeta.ToolchainGoRelpath,
		GoVersion:     "1.25.0",
		ContentSHA256: "sha256:" + strings.Repeat("a", 64),
	}
}

// xbBuildDeps returns BuildDeps with the fake toolchain/builder and the
// real protected cache under the manager home.
func xbBuildDeps(t *testing.T) (install.BuildDeps, *xbBuilder) {
	t.Helper()
	toolchain := &xbToolchain{t: t, target: xbTestTarget(), toolchain: xbTestToolchain()}
	builder := &xbBuilder{t: t}
	deps := install.BuildDeps{
		Toolchain: toolchain, Builder: builder,
		Clock:      xbClock{at: time.Unix(1_700_000_000, 0).UTC()},
		Generation: &xbGeneration{},
	}
	return deps, builder
}

// xbExternalSnapshot frames one proved external repository snapshot the
// way the raw-object admission does.
func xbExternalSnapshot() *buildrepo.Snapshot {
	files := []buildrepo.File{
		{Path: "skill-build.json", Content: []byte(`{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`)},
		{Path: "tools/cmd/tool/main.go", Content: []byte("package main\n\nfunc main() {}\n")},
		{Path: "tools/go.mod", Content: []byte("module example.test/tool\n")},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	canonical := []byte("curator-build-source-v1\x00")
	for _, file := range files {
		canonical = append(canonical, 'F')
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Path)))
		canonical = append(canonical, file.Path...)
		canonical = binary.BigEndian.AppendUint64(canonical, uint64(len(file.Content)))
		canonical = append(canonical, file.Content...)
	}
	sum := sha256.Sum256(canonical)
	return &buildrepo.Snapshot{ObjectFormat: "sha1", Commit: xbExternalLockedCommit, Files: files, CanonicalBytes: canonical, Digest: "sha256:" + hex.EncodeToString(sum[:])}
}

// xbExternalDeps serves the proved snapshot and approves every audit.
func xbExternalDeps() install.ExternalDeps {
	return install.ExternalDeps{
		Acquire: func(_ context.Context, _ install.ExternalSource) (*buildrepo.Snapshot, error) {
			return xbExternalSnapshot(), nil
		},
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil },
	}
}

// writeExternalSkill writes one local draft package declaring an
// external go-repository-v1 build command ("etool") over a declared
// build repository.
func writeExternalSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"SKILL.md":           "---\nname: " + name + "\ndescription: Test\n---\n# " + name + "\n",
		"references/info.md": "context",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	spec := `{"schema_version":7,"build_roots":[],"capabilities":{},"commands":{"etool":{"type":"build","driver":"go-repository-v1","repository":"tools","target":"tool"}},"build_repositories":{"tools":{"git":"https://git.example.com/skills/tools.git","locked_commit":{"object_format":"sha1","hex":"` + xbExternalLockedCommit + `"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
}
