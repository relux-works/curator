package install

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/marker"
)

// BuildSession is the trusted, operation-private Go session that one staging
// pass runs under. Only the frozen identities that enter a cache key are
// exposed; selection, probing, and fingerprinting stay inside the driver.
//
// VerifyToolchain re-establishes trust through the last build child and is the
// only toolchain verdict a plan takes. Release is its counterpart: a pure
// operation-private teardown that verifies nothing. The split is what keeps the
// verdict ahead of the handoff — teardown runs once the installation has either
// failed or committed, so it must not be able to report drift against live
// state it can no longer protect.
type BuildSession interface {
	Target() buildmeta.Target
	Toolchain() buildmeta.Toolchain
	VerifyToolchain(ctx context.Context) error
	Release() error
}

// Toolchain is the narrow trusted-toolchain boundary of a plan.
//
// Probe is the dry-run form: it creates and removes an operation-private probe
// root and runs no source-aware Go command. Establish is the staging form and
// keeps the private root alive until the returned session is closed.
type Toolchain interface {
	Probe(ctx context.Context) (buildmeta.Target, buildmeta.Toolchain, error)
	Establish(ctx context.Context) (BuildSession, error)
}

// CacheInspector reports reusable protected build-cache state. Every
// implementation is strictly read-only: planning never repairs, quarantines,
// publishes, or otherwise touches the live cache.
type CacheInspector interface {
	Inspect(expect buildcache.Expectation) buildcache.Result
}

// StageRequest is one planned build miss handed to the builder. The command
// surface is reproduced exactly as the package declared it so the driver can
// reject any package attempt to reach the execution boundary.
type StageRequest struct {
	Session       BuildSession
	Source        *buildsource.Token
	CommandObject map[string]any
	BuildRoot     string
	SourceDir     string
	Command       string
}

// StagedArtifact is one verified operation-private output. The path is
// manager-private and is never executed, published, or installed here.
type StagedArtifact struct {
	Path     string
	Metadata buildmeta.Artifact
}

// Builder compiles one planned miss into operation-private staging.
type Builder interface {
	Stage(ctx context.Context, request StageRequest) (StagedArtifact, error)
}

// Clock supplies the installation timestamp recorded in install markers.
type Clock interface{ Now() time.Time }

// GenerationReader reads the persistent installation generation recorded for
// one installed directory. Implementations must never write.
type GenerationReader interface {
	InstalledMarker(installedDir string) *marker.Marker
}

// BuildDeps injects the narrow boundaries the plan and staging phases use. The
// zero value resolves to the real trusted toolchain, the protected build
// cache, the go-v1 builder, the system clock, and on-disk marker reads.
type BuildDeps struct {
	Toolchain  Toolchain
	Cache      CacheInspector
	Builder    Builder
	Clock      Clock
	Generation GenerationReader
}

// resolve fills every unset boundary with its real implementation. private is
// the operation-private root the toolchain allocates its bases inside, and
// forbidden lists roots that must never host operation-private state.
func (deps BuildDeps) resolve(home string, private *privateRoot, forbidden []string) (BuildDeps, error) {
	if deps.Clock == nil {
		deps.Clock = systemClock{}
	}
	if deps.Generation == nil {
		deps.Generation = markerGeneration{}
	}
	if deps.Toolchain == nil {
		deps.Toolchain = &goToolchain{private: private, forbidden: append([]string(nil), forbidden...)}
	}
	if deps.Builder == nil {
		deps.Builder = goBuilder{}
	}
	if deps.Cache == nil {
		store, err := buildcache.New(home)
		if err != nil {
			return BuildDeps{}, fmt.Errorf("open protected build cache: %w", err)
		}
		deps.Cache = store
	}
	return deps, nil
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// installedAt renders the marker timestamp for one installation.
func installedAt(clock Clock) string {
	return clock.Now().UTC().Format("2006-01-02T15:04:05Z")
}

type markerGeneration struct{}

func (markerGeneration) InstalledMarker(installedDir string) *marker.Marker {
	return marker.Read(installedDir)
}

// goToolchain is the real trusted-toolchain boundary. Every operation gets its
// own private base inside the operation-private root — outside the manager
// home, the repository, and the runtime store — so neither a probe nor a
// staging session can write persistent state.
type goToolchain struct {
	private   *privateRoot
	forbidden []string
}

func (toolchain *goToolchain) config(base string) godriver.Config {
	return godriver.ConfigFromEnvironment(base, toolchain.forbidden...)
}

func (toolchain *goToolchain) Probe(ctx context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	base, err := toolchain.private.dir("go-probe-base-")
	if err != nil {
		return buildmeta.Target{}, buildmeta.Toolchain{}, fmt.Errorf("create toolchain probe root: %w", err)
	}
	snapshot, probeErr := godriver.Probe(ctx, toolchain.config(base))
	removeErr := os.RemoveAll(base)
	if probeErr != nil {
		return buildmeta.Target{}, buildmeta.Toolchain{}, errors.Join(probeErr, removeErr)
	}
	if removeErr != nil {
		return buildmeta.Target{}, buildmeta.Toolchain{}, fmt.Errorf("remove toolchain probe root: %w", removeErr)
	}
	return snapshot.Target, snapshot.Toolchain, nil
}

func (toolchain *goToolchain) Establish(ctx context.Context) (BuildSession, error) {
	base, err := toolchain.private.dir("go-build-base-")
	if err != nil {
		return nil, fmt.Errorf("create private build root: %w", err)
	}
	session, err := godriver.Establish(ctx, toolchain.config(base))
	if err != nil {
		return nil, errors.Join(err, os.RemoveAll(base))
	}
	return &goSession{Session: session, base: base}, nil
}

// goSession keeps the private base alive for exactly as long as the driver
// session that stages inside it.
type goSession struct {
	*godriver.Session
	base string
}

// Release drops the whole operation-private root. It goes through the driver's
// cleanup-only release rather than Close, so the deepest failure it can report
// is a leftover private path, never a toolchain verdict.
func (session *goSession) Release() error {
	return errors.Join(session.Session.Release(), os.RemoveAll(session.base))
}

// goBuilder stages one miss through the go-v1 driver. The driver places the
// artifact inside its own operation-private root, so the staged output is
// never reachable from an installation target.
type goBuilder struct{}

func (goBuilder) Stage(ctx context.Context, request StageRequest) (StagedArtifact, error) {
	session, ok := request.Session.(*goSession)
	if !ok {
		return StagedArtifact{}, fmt.Errorf("the go-v1 builder requires a trusted Go session")
	}
	result, err := godriver.Build(ctx, godriver.BuildRequest{
		Session:       session.Session,
		Source:        request.Source,
		CommandObject: godriver.BuildCommand(request.CommandObject),
		BuildRoot:     request.BuildRoot,
		SourceDir:     request.SourceDir,
		Command:       request.Command,
	})
	if err != nil {
		return StagedArtifact{}, err
	}
	return StagedArtifact{Path: result.Artifact.StagedPath, Metadata: result.Artifact.Metadata}, nil
}
