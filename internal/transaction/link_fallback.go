package transaction

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/pathboundary"
	"io"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/staging"
)

// WithSymlink replaces only the raw destination staging syscall. It does not
// decide authority, fallback admission, error classification or recovery policy.
func WithSymlink(create func(string, string) error) Option {
	return func(engine *Engine) {
		if create != nil {
			engine.symlink = create
		}
	}
}

type linkUnavailable struct {
	path  string
	cause error
}

func (e *linkUnavailable) Error() string {
	return fmt.Sprintf("symlink unavailable for %s: %v", e.path, e.cause)
}
func (e *linkUnavailable) Unwrap() error { return e.cause }

// Prepare chooses only declared noncredential projection alternatives. A
// capability refusal is returned directly only after the preparing journal and
// every sidecar were compensated. An interrupted/failed cleanup, permission
// error, nested source link or undeclared alternative never authorizes retry.
// Each retry revalidates the original live preimages, origin, whole unit and
// boundary. The final chosen bytes and copy reasons are durable before Commit;
// recovery never needs these planning alternatives or a process-local callback.
func (engine *Engine) Prepare(lock HomeLock, plan Plan) (*Journal, error) {
	if err := requireHomeLock(lock); err != nil {
		return nil, err
	}
	if err := envauthority.ValidateForUnit(engine.home, plan.Unit); err != nil {
		return nil, err
	}
	plan.Targets = append([]Target(nil), plan.Targets...)
	plan.Unit = cloneEnvironmentUnit(plan.Unit)
	pinned := map[string]string{}
	for _, target := range plan.Targets {
		fallback := target.LinkFallback
		if fallback == nil {
			continue
		}
		if target.Kind != KindEntry || target.PrivateFile || target.PrivateTree || target.StagedSource == "" || target.Class != "owned-bytes" && target.Class != "owned-tree" && target.Class != staging.ClassAdapterLedger {
			return nil, journalf("link fallback outside declared projection")
		}
		if (target.Class == "owned-bytes" || target.Class == "owned-tree") && (plan.Unit == nil || fallback.MarkerPath == "") || target.Class == staging.ClassAdapterLedger && fallback.MarkerPath != "" {
			return nil, journalf("link fallback marker binding is missing or unexpected")
		}
		if plan.Unit != nil {
			if err := plan.Unit.ValidateReadScope(fallback.Class, target.LivePath); err != nil {
				return nil, err
			}
		}
		if _, err := os.Readlink(target.StagedSource); err != nil {
			return nil, fmt.Errorf("fallback source is not a staged link: %w", err)
		}
		digest, err := DigestTarget(KindBytes, fallback.Source)
		if err != nil || digest == DigestAbsent {
			return nil, fmt.Errorf("invalid prepared copy alternative: %v", err)
		}
		pinned[target.LivePath] = digest
	}
	var scratch []string
	defer func() {
		for _, path := range scratch {
			_ = os.Remove(path)
		}
	}()
	for {
		journal, err := engine.prepare(lock, plan)
		refusal, exact := err.(*linkUnavailable)
		if !exact {
			return journal, err
		}
		selected := -1
		for i, target := range plan.Targets {
			if target.LivePath == refusal.path && target.LinkFallback != nil {
				selected = i
				break
			}
		}
		if selected == -1 {
			return nil, err
		}
		target := &plan.Targets[selected]
		fallback := target.LinkFallback
		digest, readErr := DigestTarget(KindBytes, fallback.Source)
		if readErr != nil || digest != pinned[target.LivePath] {
			return nil, journalf("prepared copy alternative changed")
		}
		info, readErr := os.Lstat(fallback.Source)
		if readErr != nil {
			return nil, readErr
		}
		expectedClass := target.Class
		if target.Class == "owned-bytes" && info.IsDir() {
			expectedClass = "owned-tree"
		}
		if fallback.Class != expectedClass {
			return nil, journalf("copy alternative class differs")
		}
		if fallback.MarkerPath != "" {
			temp, markerErr := selectCopyMarker(&plan, target.LivePath, *fallback)
			if markerErr != nil {
				return nil, markerErr
			}
			scratch = append(scratch, temp)
		}
		target.Class = fallback.Class
		target.StagedSource = fallback.Source
		target.LinkFallback = nil // at most one retry per declared entry
	}
}

func selectCopyMarker(plan *Plan, live string, fallback LinkFallback) (string, error) {
	markerIndex := -1
	for i, target := range plan.Targets {
		if target.LivePath == fallback.MarkerPath && target.PrivateFile && (target.Class == "marker" || target.Class == "native-marker") {
			markerIndex = i
			break
		}
	}
	if markerIndex == -1 {
		return "", journalf("copy alternative lacks private marker target")
	}
	target := &plan.Targets[markerIndex]
	raw, err := readFallbackMarker(target.StagedSource)
	if err != nil {
		return "", err
	}
	marker, err := envmarker.Parse(raw)
	if err != nil {
		return "", err
	}
	if !projectionSurface(fallback.Surface) || filepath.Join(filepath.Dir(fallback.MarkerPath), filepath.FromSlash(fallback.RelativePath)) != live {
		return "", journalf("copy alternative surface binding differs")
	}
	surface, ok := marker.Surfaces[fallback.Surface]
	if !ok {
		return "", journalf("copy alternative surface absent")
	}
	declared := false
	for _, path := range surface.Paths {
		if path == fallback.RelativePath {
			declared = true
		}
	}
	if !declared {
		return "", journalf("copy alternative entry not declared")
	}
	copies := []envmarker.Copy{}
	if surface.Copies != nil {
		copies = append(copies, *surface.Copies...)
	}
	for _, copy := range copies {
		if copy.Path == fallback.RelativePath {
			return "", journalf("copy alternative entry already copied")
		}
	}
	copies = append(copies, envmarker.Copy{Path: fallback.RelativePath, Reason: envmarker.ReasonSymlinkFallback})
	surface.Copies = &copies
	marker.Surfaces[fallback.Surface] = surface
	updated, err := marker.Marshal()
	if err != nil {
		return "", err
	}
	bound := 0
	for i := range plan.Unit.Homes {
		home := &plan.Unit.Homes[i]
		if filepath.Join(home.Target, envmarker.Name) == fallback.MarkerPath && bytes.Equal(home.After.Marker, raw) {
			home.After.Marker = updated
			bound++
		}
	}
	for i := range plan.Unit.Native {
		native := &plan.Unit.Native[i]
		if filepath.Join(native.Root, envmarker.Name) == fallback.MarkerPath && bytes.Equal(native.MarkerAfter, raw) {
			native.MarkerAfter = updated
			bound++
		}
	}
	if bound != 1 {
		return "", journalf("copy alternative whole-unit marker binding differs")
	}
	file, err := privatedir.CreateTemp("", "curator-copy-marker-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if _, err := file.Write(updated); err != nil {
		_ = file.Close()
		_ = os.Remove(name)
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return "", err
	}
	target.StagedSource = name
	return name, nil
}

// Read only the exact enumerated private marker: bounded, no-follow, and pinned
// before/after the read. No native credential or unrelated home tree is read.
func readFallbackMarker(path string) (data []byte, resultErr error) {
	if err := privatedir.ValidateFile(path); err != nil {
		return nil, err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if before.Size() > 16<<20 {
		return nil, journalf("fallback marker too large")
	}
	file, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			data = nil
			resultErr = errors.Join(resultErr, err)
		}
	}()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) {
		return nil, journalf("fallback marker entry changed")
	}
	raw, err := io.ReadAll(io.LimitReader(file, (16<<20)+1))
	if err != nil {
		return nil, err
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(before, after) || before.Size() != int64(len(raw)) || before.ModTime() != after.ModTime() {
		return nil, journalf("fallback marker changed during read")
	}
	if err := privatedir.ValidateFile(path); err != nil {
		return nil, err
	}
	return raw, nil
}

// These are the existing marker's closed, manager-authored projection surfaces.
// Credentials, seed inputs, tool settings and sessions have no such alternative.
func projectionSurface(surface string) bool {
	switch surface {
	case envmarker.SurfaceRootContext, envmarker.SurfaceSkills, envmarker.SurfaceSystemPrompt, envmarker.SurfaceMCP:
		return true
	}
	return false
}
