package conformanceconsumer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RuntimeObservation is supplied by a platform observer while the manager is
// alive. Processes and network endpoints are logical evidence strings; the
// observer owns platform-specific collection and must not mutate the run.
type RuntimeObservation struct {
	Processes []string `json:"processes,omitempty"`
	Network   []string `json:"network,omitempty"`
}

// RuntimeObserver samples process and network activity until ctx is canceled.
// Native macOS and Windows harnesses can implement this without importing a
// manager. Tests can inject deterministic observations.
type RuntimeObserver interface {
	Observe(ctx context.Context, managerPID int) (RuntimeObservation, error)
}

// FileObservation is a stable logical snapshot entry. Absolute implementation
// paths never enter comparison or report bytes.
type FileObservation struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	SHA256 string `json:"sha256,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

// FileChange describes one observable filesystem effect.
type FileChange struct {
	Path   string `json:"path"`
	Change string `json:"change"`
	Before string `json:"before_sha256,omitempty"`
	After  string `json:"after_sha256,omitempty"`
}

// CommandSpec is one released-manager black-box invocation.
type CommandSpec struct {
	Manager          string
	Executable       string
	Args             []string
	Dir              string
	Env              []string
	WatchedRoots     map[string]string
	MutableRoots     []string
	AllowedWrites    []string
	AllowedProcesses []string
	NormalizePaths   map[string]string
	Observer         RuntimeObserver
	ObserverTimeout  time.Duration
}

// CommandObservation is deterministic evidence from one independent process.
type CommandObservation struct {
	Manager             string            `json:"manager"`
	ExitCode            int               `json:"exit_code"`
	Stdout              string            `json:"stdout"`
	Stderr              string            `json:"stderr"`
	Files               []FileObservation `json:"files"`
	Changes             []FileChange      `json:"changes"`
	Processes           []string          `json:"processes,omitempty"`
	Network             []string          `json:"network,omitempty"`
	UnexpectedProcesses []string          `json:"unexpected_processes,omitempty"`
	UnexpectedNetwork   []string          `json:"unexpected_network,omitempty"`
	UnexpectedWrites    []string          `json:"unexpected_writes,omitempty"`
	MutationOnFailure   bool              `json:"mutation_on_failure,omitempty"`
	ObservationSHA256   string            `json:"observation_sha256"`
}

// RunCommand executes one manager as an independent process with exactly the
// supplied environment, snapshots all declared roots, and records forbidden
// activity reported by the platform observer.
func RunCommand(ctx context.Context, spec CommandSpec) (CommandObservation, error) {
	if spec.Manager == "" || spec.Executable == "" {
		return CommandObservation{}, fmt.Errorf("manager and executable are required")
	}
	executable, err := exec.LookPath(spec.Executable)
	if err != nil {
		return CommandObservation{}, fmt.Errorf("resolve manager executable: %w", err)
	}
	before, err := snapshotRoots(spec.WatchedRoots)
	if err != nil {
		return CommandObservation{}, fmt.Errorf("snapshot before manager run: %w", err)
	}
	command := exec.CommandContext(ctx, executable, spec.Args...)
	command.Dir = spec.Dir
	command.Env = append([]string(nil), spec.Env...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return CommandObservation{}, fmt.Errorf("start manager process: %w", err)
	}

	observerCtx, cancelObserver := context.WithCancel(context.Background())
	type observerResult struct {
		observation RuntimeObservation
		err         error
	}
	observerDone := make(chan observerResult, 1)
	if spec.Observer == nil {
		observerDone <- observerResult{}
	} else {
		go func() {
			observation, observeErr := spec.Observer.Observe(observerCtx, command.Process.Pid)
			observerDone <- observerResult{observation: observation, err: observeErr}
		}()
	}
	waitErr := command.Wait()
	cancelObserver()
	timeout := spec.ObserverTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	var observed observerResult
	select {
	case observed = <-observerDone:
	case <-time.After(timeout):
		return CommandObservation{}, fmt.Errorf("runtime observer did not stop within %s", timeout)
	}
	if observed.err != nil {
		return CommandObservation{}, fmt.Errorf("runtime observer: %w", observed.err)
	}

	after, err := snapshotRoots(spec.WatchedRoots)
	if err != nil {
		return CommandObservation{}, fmt.Errorf("snapshot after manager run: %w", err)
	}
	exitCode := 0
	if waitErr != nil {
		if ctx.Err() != nil {
			return CommandObservation{}, ctx.Err()
		}
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			return CommandObservation{}, fmt.Errorf("wait for manager process: %w", waitErr)
		}
		exitCode = exitError.ExitCode()
	}

	changes := diffFiles(before, after)
	result := CommandObservation{
		Manager:   spec.Manager,
		ExitCode:  exitCode,
		Stdout:    normalizeOutput(stdout.String(), spec.NormalizePaths),
		Stderr:    normalizeOutput(stderr.String(), spec.NormalizePaths),
		Files:     after,
		Changes:   changes,
		Processes: sortedUnique(observed.observation.Processes),
		Network:   sortedUnique(observed.observation.Network),
	}
	result.UnexpectedProcesses = unexpectedProcesses(result.Processes, executable, spec.AllowedProcesses)
	result.UnexpectedNetwork = append([]string(nil), result.Network...)
	result.UnexpectedWrites = unexpectedWrites(changes, spec.AllowedWrites)
	if exitCode != 0 {
		result.MutationOnFailure = changedRoot(changes, spec.MutableRoots)
	}
	digest, err := observationDigest(result)
	if err != nil {
		return CommandObservation{}, err
	}
	result.ObservationSHA256 = digest
	return result, nil
}

func snapshotRoots(roots map[string]string) ([]FileObservation, error) {
	labels := make([]string, 0, len(roots))
	for label := range roots {
		if err := validateCorpusPath(label); err != nil {
			return nil, fmt.Errorf("watched root %q: %w", label, err)
		}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	var result []FileObservation
	for _, label := range labels {
		root := roots[label]
		info, err := os.Lstat(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, fmt.Errorf("watched root %q must be a real directory", label)
		}
		err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if name == root {
				return nil
			}
			relative, err := filepath.Rel(root, name)
			if err != nil {
				return err
			}
			logical := label + "/" + filepath.ToSlash(relative)
			entryInfo, err := entry.Info()
			if err != nil {
				return err
			}
			observation := FileObservation{Path: logical, Size: entryInfo.Size()}
			switch {
			case entry.Type()&os.ModeSymlink != 0:
				observation.Kind = "symlink"
				target, err := os.Readlink(name)
				if err != nil {
					return err
				}
				observation.SHA256 = digestBytes([]byte(filepath.ToSlash(target)))
			case entry.IsDir():
				observation.Kind = "directory"
				observation.Size = 0
			case entry.Type().IsRegular():
				observation.Kind = "file"
				payload, err := os.ReadFile(name)
				if err != nil {
					return err
				}
				observation.SHA256 = digestBytes(payload)
			default:
				observation.Kind = "special"
			}
			result = append(result, observation)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func diffFiles(before, after []FileObservation) []FileChange {
	old := make(map[string]FileObservation, len(before))
	current := make(map[string]FileObservation, len(after))
	for _, file := range before {
		old[file.Path] = file
	}
	for _, file := range after {
		current[file.Path] = file
	}
	paths := make([]string, 0, len(old)+len(current))
	seen := map[string]struct{}{}
	for path := range old {
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	for path := range current {
		if _, exists := seen[path]; !exists {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	changes := make([]FileChange, 0)
	for _, path := range paths {
		beforeFile, hadBefore := old[path]
		afterFile, hasAfter := current[path]
		change := FileChange{Path: path}
		switch {
		case !hadBefore:
			change.Change, change.After = "added", afterFile.SHA256
		case !hasAfter:
			change.Change, change.Before = "removed", beforeFile.SHA256
		case beforeFile != afterFile:
			change.Change, change.Before, change.After = "modified", beforeFile.SHA256, afterFile.SHA256
		default:
			continue
		}
		changes = append(changes, change)
	}
	return changes
}

func normalizeOutput(value string, replacements map[string]string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	type replacement struct{ from, to string }
	ordered := make([]replacement, 0, len(replacements))
	for from, to := range replacements {
		if from != "" {
			ordered = append(ordered, replacement{from, to})
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return len(ordered[i].from) > len(ordered[j].from) })
	for _, item := range ordered {
		value = strings.ReplaceAll(value, item.from, item.to)
		value = strings.ReplaceAll(value, filepath.ToSlash(item.from), item.to)
	}
	return value
}

func unexpectedProcesses(processes []string, executable string, allowed []string) []string {
	allow := map[string]struct{}{strings.ToLower(filepath.Base(executable)): {}}
	for _, name := range allowed {
		allow[strings.ToLower(filepath.Base(name))] = struct{}{}
	}
	var unexpected []string
	for _, process := range processes {
		if _, ok := allow[strings.ToLower(filepath.Base(process))]; !ok {
			unexpected = append(unexpected, process)
		}
	}
	return sortedUnique(unexpected)
}

func unexpectedWrites(changes []FileChange, allowed []string) []string {
	var unexpected []string
	for _, change := range changes {
		permitted := false
		for _, prefix := range allowed {
			prefix = strings.TrimSuffix(prefix, "/")
			if change.Path == prefix || strings.HasPrefix(change.Path, prefix+"/") {
				permitted = true
				break
			}
		}
		if !permitted {
			unexpected = append(unexpected, change.Path)
		}
	}
	return sortedUnique(unexpected)
}

func changedRoot(changes []FileChange, mutable []string) bool {
	for _, change := range changes {
		for _, root := range mutable {
			root = strings.TrimSuffix(root, "/")
			if change.Path == root || strings.HasPrefix(change.Path, root+"/") {
				return true
			}
		}
	}
	return false
}

func observationDigest(observation CommandObservation) (string, error) {
	copyObservation := observation
	copyObservation.ObservationSHA256 = ""
	payload, err := json.Marshal(copyObservation)
	if err != nil {
		return "", fmt.Errorf("encode command observation: %w", err)
	}
	return digestBytes(payload), nil
}

func digestBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func sortedUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

// InspectBinary pins a released binary by content and its black-box version
// output. Source revision remains explicit and cannot be guessed from a file.
func InspectBinary(ctx context.Context, manager, executable, revision, specRevision, toolchain, operatingSystem, architecture string, versionArgs, env []string) (RevisionEvidence, error) {
	if manager == "" || revision == "" || specRevision == "" || toolchain == "" || operatingSystem == "" || architecture == "" {
		return RevisionEvidence{}, fmt.Errorf("complete binary revision metadata is required")
	}
	resolved, err := exec.LookPath(executable)
	if err != nil {
		return RevisionEvidence{}, err
	}
	payload, err := os.ReadFile(resolved)
	if err != nil {
		return RevisionEvidence{}, err
	}
	command := exec.CommandContext(ctx, resolved, versionArgs...)
	command.Env = append([]string(nil), env...)
	output, err := command.CombinedOutput()
	if err != nil {
		return RevisionEvidence{}, fmt.Errorf("read %s version: %w: %s", manager, err, strings.TrimSpace(string(output)))
	}
	version := strings.TrimSpace(strings.ReplaceAll(string(output), "\r\n", "\n"))
	if version == "" {
		return RevisionEvidence{}, fmt.Errorf("%s returned an empty version", manager)
	}
	return RevisionEvidence{Manager: manager, Version: version, Revision: revision, BinarySHA256: digestBytes(payload), SpecRevision: specRevision, Toolchain: toolchain, OperatingSystem: operatingSystem, Architecture: architecture}, nil
}

// WriteFailureArtifacts preserves raw output and the complete normalized
// observation under a caller-owned task directory.
func WriteFailureArtifacts(root, caseID string, observation CommandObservation) (string, error) {
	if err := validateCorpusPath(caseID); err != nil || strings.Contains(caseID, "/") {
		return "", fmt.Errorf("failure case id must be one safe path component")
	}
	directory := filepath.Join(root, caseID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		return "", err
	}
	files := map[string][]byte{"observation.json": append(payload, '\n'), "stdout.log": []byte(observation.Stdout), "stderr.log": []byte(observation.Stderr)}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), content, 0o600); err != nil {
			return "", err
		}
	}
	return filepath.ToSlash(caseID), nil
}
