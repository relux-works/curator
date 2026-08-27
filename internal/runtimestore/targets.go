package runtimestore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/identifiers"
)

// TargetKind distinguishes copied script runtimes from immutable compiled
// cache artifacts. Callers must not overload a snapshot source path as an
// executable target.
type TargetKind string

const (
	ScriptRuntime TargetKind = "script-runtime"
	BuildArtifact TargetKind = "build-artifact"
)

// RuntimeTarget is the closed set of executable sources accepted by shim
// staging. Implementations are constructed only by this package.
type RuntimeTarget interface {
	Kind() TargetKind
	ExecutablePath() string
	runtimeTarget()
}

// ScriptTarget selects one command in a validated commit-keyed runtime tree.
type ScriptTarget struct {
	runtimeDir string
	executable string
}

func (target ScriptTarget) Kind() TargetKind       { return ScriptRuntime }
func (target ScriptTarget) ExecutablePath() string { return target.executable }
func (target ScriptTarget) RuntimeDir() string     { return target.runtimeDir }
func (target ScriptTarget) runtimeTarget()         {}

// CompiledTarget selects one immutable artifact returned by protected cache
// inspection. It never copies the artifact or its build root into runtime/.
type CompiledTarget struct {
	artifactPath string
	cacheKey     buildmeta.CacheKey
	receiptHash  buildmeta.ReceiptHash
}

func (target CompiledTarget) Kind() TargetKind                   { return BuildArtifact }
func (target CompiledTarget) ExecutablePath() string             { return target.artifactPath }
func (target CompiledTarget) CacheKey() buildmeta.CacheKey       { return target.cacheKey }
func (target CompiledTarget) ReceiptHash() buildmeta.ReceiptHash { return target.receiptHash }
func (target CompiledTarget) runtimeTarget()                     {}

// CompiledTargetFromCache converts only an exact protected-cache hit into a
// shim target. Marker selection belongs to the caller through the expectation
// used for Store.Inspect; this constructor preserves that validated identity.
func CompiledTargetFromCache(hit buildcache.Result, platform string) (CompiledTarget, error) {
	if hit.Status != buildcache.Hit {
		return CompiledTarget{}, fmt.Errorf("compiled runtime target requires a protected cache hit")
	}
	if err := hit.Receipt.Validate(); err != nil {
		return CompiledTarget{}, fmt.Errorf("compiled runtime receipt: %w", err)
	}
	if hit.ReceiptHash == "" || hit.ArtifactPath == "" {
		return CompiledTarget{}, fmt.Errorf("compiled runtime target is missing cache identity")
	}
	if platform != "unix" && platform != "windows" {
		return CompiledTarget{}, fmt.Errorf("unsupported shim platform %q", platform)
	}
	if hit.Receipt.Input.Target.GOOS != runtime.GOOS || hit.Receipt.Input.Target.GOARCH != runtime.GOARCH {
		return CompiledTarget{}, fmt.Errorf("compiled artifact target is not native to this manager")
	}
	if (hit.Receipt.Input.Target.GOOS == "windows") != (platform == "windows") {
		return CompiledTarget{}, fmt.Errorf("compiled artifact target OS does not match shim platform")
	}
	if _, err := cleanAbsolute(hit.ArtifactPath); err != nil {
		return CompiledTarget{}, fmt.Errorf("compiled artifact path: %w", err)
	}
	info, err := os.Lstat(hit.ArtifactPath)
	if err != nil || !info.Mode().IsRegular() {
		return CompiledTarget{}, fmt.Errorf("compiled artifact is not a regular file")
	}
	if platform == "windows" && !strings.EqualFold(filepath.Ext(hit.ArtifactPath), ".exe") {
		return CompiledTarget{}, fmt.Errorf("Windows compiled artifact must have an .exe suffix")
	}
	if platform == "unix" && info.Mode().Perm()&0o111 == 0 {
		return CompiledTarget{}, fmt.Errorf("compiled artifact is not executable")
	}
	return CompiledTarget{
		artifactPath: hit.ArtifactPath,
		cacheKey:     hit.Receipt.CacheKey,
		receiptHash:  hit.ReceiptHash,
	}, nil
}

// ShimRole classifies the only live launcher paths managed by this planner.
type ShimRole string

const (
	ProjectShim         ShimRole = "project"
	GlobalCanonicalShim ShimRole = "global-canonical"
	SafeForwardingShim  ShimRole = "safe-forwarding"
)

// ManagedShim is a typed manager-owned launcher destination. Its path is
// derived from a validated bin directory and command name, never supplied as
// an arbitrary removal path.
type ManagedShim struct {
	role     ShimRole
	binDir   string
	command  string
	platform string
	path     string
}

func NewManagedShim(role ShimRole, binDir, command, platform string) (ManagedShim, error) {
	if role != ProjectShim && role != GlobalCanonicalShim && role != SafeForwardingShim {
		return ManagedShim{}, fmt.Errorf("unsupported managed shim role %q", role)
	}
	if !identifiers.Valid(command) {
		return ManagedShim{}, fmt.Errorf("managed shim command is not a portable identifier")
	}
	if platform != "unix" && platform != "windows" {
		return ManagedShim{}, fmt.Errorf("unsupported shim platform %q", platform)
	}
	abs, err := cleanAbsolute(binDir)
	if err != nil {
		return ManagedShim{}, fmt.Errorf("managed shim bin: %w", err)
	}
	name := command
	if platform == "windows" {
		name += ".cmd"
	}
	return ManagedShim{role: role, binDir: abs, command: command, platform: platform, path: filepath.Join(abs, name)}, nil
}

func (shim ManagedShim) Role() ShimRole   { return shim.role }
func (shim ManagedShim) BinDir() string   { return shim.binDir }
func (shim ManagedShim) Command() string  { return shim.command }
func (shim ManagedShim) Platform() string { return shim.platform }
func (shim ManagedShim) Path() string     { return shim.path }

// ShimSpec describes one desired launcher and its direct executable target.
type ShimSpec struct {
	Destination ManagedShim
	Target      RuntimeTarget
	PathEntries []string
}

// ManagedTargetKind identifies staged replacement classes for the transaction
// layer which owns live swaps.
type ManagedTargetKind string

const (
	RuntimeTreeTarget ManagedTargetKind = "runtime-tree"
	ShimTarget        ManagedTargetKind = "shim"
)

// DesiredTarget maps operation-private staged bytes to an eventual live path.
// This package writes StagedPath only.
type DesiredTarget struct {
	Kind        ManagedTargetKind
	Role        ShimRole
	Command     string
	RuntimeKind TargetKind
	LivePath    string
	StagedPath  string
}

// RemovalTarget is an explicitly manager-owned live path absent from the next
// desired set. Planning does not remove it.
type RemovalTarget struct {
	Kind     ManagedTargetKind
	Role     ShimRole
	Command  string
	LivePath string
}

// TransitionPlan contains deterministic desired replacements and removals.
type TransitionPlan struct {
	Desired  []DesiredTarget
	Removals []RemovalTarget
}

// StageShimTransition materializes only operation-private launcher files.
// Live project, canonical-global, and safe-forwarding paths are returned as
// transaction targets and are never created, replaced, or removed here.
func StageShimTransition(stageRoot string, desired []ShimSpec, currentlyManaged []ManagedShim) (TransitionPlan, error) {
	stageRoot, err := cleanAbsolute(stageRoot)
	if err != nil {
		return TransitionPlan{}, fmt.Errorf("shim staging root: %w", err)
	}
	desired = append([]ShimSpec(nil), desired...)
	currentlyManaged = append([]ManagedShim(nil), currentlyManaged...)
	sort.Slice(desired, func(i, j int) bool {
		return shimSortKey(desired[i].Destination) < shimSortKey(desired[j].Destination)
	})
	plan := TransitionPlan{}
	desiredPaths := make(map[string]bool, len(desired))
	for _, spec := range desired {
		shim := spec.Destination
		if err := validateManagedShim(shim); err != nil {
			return TransitionPlan{}, err
		}
		if pathsOverlap(stageRoot, shim.binDir) {
			return TransitionPlan{}, fmt.Errorf("shim staging root overlaps live managed bin %s", shim.binDir)
		}
		if spec.Target == nil {
			return TransitionPlan{}, fmt.Errorf("managed shim %s has no runtime target", shim.path)
		}
		if err := validatePathEntries(spec.PathEntries, shim.platform); err != nil {
			return TransitionPlan{}, fmt.Errorf("managed shim %s: %w", shim.path, err)
		}
		key := platformPathKey(shim.path, shim.platform)
		if desiredPaths[key] {
			return TransitionPlan{}, fmt.Errorf("duplicate desired managed shim %s", shim.path)
		}
		desiredPaths[key] = true
		staged := stagedShimPath(stageRoot, shim)
		if err := writeStagedShim(staged, spec.Target.ExecutablePath(), shim.platform, spec.PathEntries); err != nil {
			return TransitionPlan{}, fmt.Errorf("stage managed shim %s: %w", shim.path, err)
		}
		plan.Desired = append(plan.Desired, DesiredTarget{
			Kind: ShimTarget, Role: shim.role, Command: shim.command,
			RuntimeKind: spec.Target.Kind(), LivePath: shim.path, StagedPath: staged,
		})
	}

	sort.Slice(currentlyManaged, func(i, j int) bool { return shimSortKey(currentlyManaged[i]) < shimSortKey(currentlyManaged[j]) })
	seenCurrent := map[string]bool{}
	for _, shim := range currentlyManaged {
		if err := validateManagedShim(shim); err != nil {
			return TransitionPlan{}, err
		}
		key := platformPathKey(shim.path, shim.platform)
		if seenCurrent[key] {
			continue
		}
		seenCurrent[key] = true
		if !desiredPaths[key] {
			plan.Removals = append(plan.Removals, RemovalTarget{
				Kind: ShimTarget, Role: shim.role, Command: shim.command, LivePath: shim.path,
			})
		}
	}
	return plan, nil
}

func validatePathEntries(entries []string, platform string) error {
	for _, entry := range entries {
		if _, err := cleanAbsolute(entry); err != nil {
			return fmt.Errorf("PATH entry: %w", err)
		}
		if platform == "unix" && strings.Contains(entry, ":") {
			return fmt.Errorf("Unix PATH entry contains a separator")
		}
		if platform == "windows" && strings.Contains(entry, ";") {
			return fmt.Errorf("Windows PATH entry contains a separator")
		}
	}
	return nil
}

func validateManagedShim(shim ManagedShim) error {
	want, err := NewManagedShim(shim.role, shim.binDir, shim.command, shim.platform)
	if err != nil {
		return err
	}
	if want.path != shim.path {
		return fmt.Errorf("managed shim path is not manager-derived")
	}
	return nil
}

func writeStagedShim(path, executable, platform string, pathEntries []string) error {
	if executable == "" || !filepath.IsAbs(executable) {
		return fmt.Errorf("runtime executable must be an absolute path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := UnixShimContent(executable, pathEntries)
	if platform == "windows" {
		content = WindowsShimContent(executable, pathEntries)
	}
	return os.WriteFile(path, []byte(content), 0o700)
}

func stagedShimPath(stageRoot string, shim ManagedShim) string {
	digest := sha256.Sum256([]byte(shim.path))
	return filepath.Join(stageRoot, "shims", string(shim.role), hex.EncodeToString(digest[:]), filepath.Base(shim.path))
}

func shimSortKey(shim ManagedShim) string {
	return string(shim.role) + "\x00" + shim.command + "\x00" + platformPathKey(shim.path, shim.platform)
}

func platformPathKey(path, platform string) string {
	if platform == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func cleanAbsolute(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", fmt.Errorf("path must be clean and absolute")
	}
	return path, nil
}

func pathsOverlap(one, two string) bool {
	return pathWithin(one, two) || pathWithin(two, one)
}

func pathWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
