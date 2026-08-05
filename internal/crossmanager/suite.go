package crossmanager

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
	"reflect"
	"sort"
	"strings"
)

// ManagerEvidence identifies the independently built process under test.
type ManagerEvidence struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Revision     string `json:"revision"`
	BinarySHA256 string `json:"binary_sha256"`
	Toolchain    string `json:"toolchain"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

// Workspace is an isolated case root. Adapters may prepare fixtures before
// returning an invocation, but the manager is always launched as a process.
type Workspace struct {
	Root string
	Home string
	Repo string
	Temp string
}

// Invocation is the complete black-box process boundary for one case.
type Invocation struct {
	Executable        string
	Args              []string
	Env               []string
	Dir               string
	Stdin             []byte
	WatchRoots        []string
	AllowedWriteRoots []string
}

// Adapter maps neutral cases to one manager CLI without importing it.
type Adapter struct {
	Evidence  ManagerEvidence
	Prepare   func(context.Context, Case, Workspace) (Invocation, error)
	Normalize func(Case, RunResult) (map[string]any, error)
}

// BoundaryEvents are observations from a platform sampler such as ps/lsof on
// macOS or CIM/NetTCPConnection on Windows.
type BoundaryEvents struct {
	UnexpectedProcesses []string `json:"unexpected_processes,omitempty"`
	NetworkConnections  []string `json:"network_connections,omitempty"`
}

// BoundaryProbe brackets a manager process. The engine treats all returned
// events as violations unless the adapter's normalized observation and case
// expectation explicitly allow them.
type BoundaryProbe interface {
	Begin(context.Context, Invocation) (any, error)
	End(context.Context, any) (BoundaryEvents, error)
}

// Suite executes the same authenticated cases for every adapter.
type Suite struct {
	Corpus       *Corpus
	SpecRevision string
	WorkRoot     string
	ArtifactRoot string
	Probe        BoundaryProbe
}

// RunAdapter executes all cases for one independent manager process.
func (s Suite) RunAdapter(ctx context.Context, adapter Adapter) (Report, error) {
	if s.Corpus == nil {
		return Report{}, fmt.Errorf("corpus is required")
	}
	if s.SpecRevision == "" {
		return Report{}, fmt.Errorf("spec revision is required")
	}
	if adapter.Evidence.Name == "" || adapter.Prepare == nil || adapter.Normalize == nil {
		return Report{}, fmt.Errorf("complete manager adapter is required")
	}
	if err := validateManagerEvidence(adapter.Evidence); err != nil {
		return Report{}, err
	}
	cases, err := s.Corpus.Cases()
	if err != nil {
		return Report{}, err
	}
	if s.Probe == nil {
		return Report{}, fmt.Errorf("boundary probe is required")
	}
	corpusEvidence := s.Corpus.Evidence()
	corpusEvidence.Revision = s.SpecRevision
	report := Report{Schema: ReportSchemaV1, Corpus: corpusEvidence, Adapter: adapter.Evidence.Name, Manager: adapter.Evidence}
	for _, item := range cases {
		result := s.runCase(ctx, s.Probe, adapter, item)
		report.Cases = append(report.Cases, result)
	}
	return report, nil
}

func (s Suite) runCase(ctx context.Context, probe BoundaryProbe, adapter Adapter, item Case) CaseResult {
	result := CaseResult{ID: item.ID, Status: CaseError, ExitCode: -1, Expected: append(json.RawMessage(nil), item.Expected...)}
	workspace, err := s.newWorkspace(adapter.Evidence.Name, item.ID)
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	invocation, err := adapter.Prepare(ctx, item, workspace)
	if err != nil {
		result.Detail = "prepare: " + err.Error()
		return result
	}
	if invocation.Executable == "" {
		result.Detail = "prepare: executable is required"
		return result
	}
	if invocation.Dir == "" {
		invocation.Dir = workspace.Repo
	}
	if len(invocation.WatchRoots) == 0 {
		invocation.WatchRoots = []string{workspace.Root}
	}
	if len(invocation.AllowedWriteRoots) == 0 {
		invocation.AllowedWriteRoots = []string{workspace.Root}
	}
	before, err := snapshotRoots(invocation.WatchRoots)
	if err != nil {
		result.Detail = "snapshot before: " + err.Error()
		return result
	}
	token, err := probe.Begin(ctx, invocation)
	if err != nil {
		result.Detail = "boundary begin: " + err.Error()
		return result
	}
	run, runErr := executeInvocation(ctx, invocation)
	events, probeErr := probe.End(ctx, token)
	after, snapshotErr := snapshotRoots(invocation.WatchRoots)
	result.ExitCode = run.ExitCode
	result.StdoutSHA256 = digestBytes(run.Stdout)
	result.StderrSHA256 = digestBytes(run.Stderr)
	result.Boundary = events
	result.ChangedPaths = changedPaths(before, after)
	result.OutsideWrites = outsideWrites(result.ChangedPaths, invocation.AllowedWriteRoots)
	result.ArtifactDir, _ = s.writeArtifacts(adapter.Evidence.Name, item.ID, run, before, after, events)
	if runErr != nil {
		result.Detail = "execute: " + runErr.Error()
		return result
	}
	if probeErr != nil {
		result.Detail = "boundary end: " + probeErr.Error()
		return result
	}
	if snapshotErr != nil {
		result.Detail = "snapshot after: " + snapshotErr.Error()
		return result
	}
	observed, err := adapter.Normalize(item, run)
	if err != nil {
		result.Detail = "normalize: " + err.Error()
		return result
	}
	observedBytes, _ := json.Marshal(observed)
	result.Observed = observedBytes
	var expected map[string]any
	if err := json.Unmarshal(item.Expected, &expected); err != nil {
		result.Detail = "expected: " + err.Error()
		return result
	}
	violations := compareExpected(expected, observed, "")
	violations = append(violations, boundaryViolations(expected, events)...)
	if mutation, present := expected["mutation"].(bool); present && !mutation && len(result.ChangedPaths) != 0 {
		violations = append(violations, "mutation=false but filesystem changed")
	}
	if len(result.OutsideWrites) != 0 {
		violations = append(violations, "write escaped allowed roots")
	}
	sort.Strings(violations)
	result.Violations = violations
	if len(violations) == 0 {
		result.Status = CaseObserved
	} else {
		result.Status = CaseMismatch
		result.Detail = strings.Join(violations, "; ")
	}
	return result
}

func (s Suite) newWorkspace(adapter, caseID string) (Workspace, error) {
	if err := validateSegment(adapter); err != nil {
		return Workspace{}, fmt.Errorf("adapter name: %w", err)
	}
	if err := validateSegment(caseID); err != nil {
		return Workspace{}, fmt.Errorf("case id: %w", err)
	}
	root := filepath.Join(s.WorkRoot, adapter, caseID)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return Workspace{}, err
	}
	workspace := Workspace{Root: root, Home: filepath.Join(root, "home"), Repo: filepath.Join(root, "repo"), Temp: filepath.Join(root, "tmp")}
	for _, directory := range []string{workspace.Home, workspace.Repo, workspace.Temp} {
		if err := os.Mkdir(directory, 0o700); err != nil && !os.IsExist(err) {
			return Workspace{}, err
		}
	}
	return workspace, nil
}

func validateSegment(value string) error {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\\`) {
		return fmt.Errorf("%q is not a path-safe segment", value)
	}
	return nil
}

func validateManagerEvidence(evidence ManagerEvidence) error {
	if evidence.Version == "" || evidence.Revision == "" || evidence.Toolchain == "" || evidence.OS == "" || evidence.Architecture == "" {
		return fmt.Errorf("manager %q revision metadata is incomplete", evidence.Name)
	}
	if _, err := parseDigest(evidence.BinarySHA256); err != nil {
		return fmt.Errorf("manager %q binary digest: %w", evidence.Name, err)
	}
	return validateSegment(evidence.Name)
}

func executeInvocation(ctx context.Context, invocation Invocation) (RunResult, error) {
	// #nosec G204 -- adapters deliberately select the independent manager binary;
	// arguments are passed directly without shell interpretation.
	command := exec.CommandContext(ctx, invocation.Executable, invocation.Args...)
	command.Env = append([]string(nil), invocation.Env...)
	command.Dir = invocation.Dir
	command.Stdin = bytes.NewReader(invocation.Stdin)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	result := RunResult{ExitCode: 0, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if err == nil {
		return result, nil
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return result, err
	}
	result.ExitCode = exitError.ExitCode()
	return result, nil
}

type fileState struct {
	Mode   fs.FileMode `json:"mode"`
	Size   int64       `json:"size"`
	SHA256 string      `json:"sha256,omitempty"`
}

func snapshotRoots(roots []string) (map[string]fileState, error) {
	result := make(map[string]fileState)
	for _, root := range roots {
		absolute, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		err = filepath.WalkDir(absolute, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				if os.IsNotExist(walkErr) {
					return nil
				}
				return walkErr
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			state := fileState{Mode: info.Mode(), Size: info.Size()}
			if info.Mode().IsRegular() {
				// #nosec G304 -- name is emitted by WalkDir beneath a validated watch root.
				payload, err := os.ReadFile(name)
				if err != nil {
					return err
				}
				state.SHA256 = digestBytes(payload)
			}
			result[name] = state
			if info.Mode()&os.ModeSymlink != 0 {
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return result, nil
}

func changedPaths(before, after map[string]fileState) []string {
	set := make(map[string]struct{}, len(before)+len(after))
	for name := range before {
		set[name] = struct{}{}
	}
	for name := range after {
		set[name] = struct{}{}
	}
	changed := make([]string, 0)
	for name := range set {
		if !reflect.DeepEqual(before[name], after[name]) {
			changed = append(changed, name)
		}
	}
	sort.Strings(changed)
	return changed
}

func outsideWrites(changed, allowed []string) []string {
	result := make([]string, 0)
	for _, name := range changed {
		inside := false
		for _, root := range allowed {
			absolute, err := filepath.Abs(root)
			if err != nil {
				continue
			}
			relative, err := filepath.Rel(absolute, name)
			if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
				inside = true
				break
			}
		}
		if !inside {
			result = append(result, name)
		}
	}
	return result
}

func compareExpected(expected, observed map[string]any, prefix string) []string {
	var violations []string
	keys := make([]string, 0, len(expected))
	for key := range expected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		got, present := observed[key]
		if !present {
			violations = append(violations, path+" missing")
			continue
		}
		wantMap, nested := expected[key].(map[string]any)
		gotMap, gotNested := got.(map[string]any)
		if nested && gotNested {
			violations = append(violations, compareExpected(wantMap, gotMap, path)...)
			continue
		}
		if !reflect.DeepEqual(expected[key], got) {
			violations = append(violations, path+" differs")
		}
	}
	return violations
}

func boundaryViolations(expected map[string]any, events BoundaryEvents) []string {
	var violations []string
	if value, present := expected["unexpected_children"].(bool); present && !value && len(events.UnexpectedProcesses) != 0 {
		violations = append(violations, "unexpected child process")
	}
	for _, key := range []string{"network_started", "shell_started", "compiler_started", "exec_started", "child_started"} {
		if value, present := expected[key].(bool); present && !value {
			if key == "network_started" && len(events.NetworkConnections) != 0 {
				violations = append(violations, "unexpected network connection")
			}
			if key != "network_started" && len(events.UnexpectedProcesses) != 0 {
				violations = append(violations, "unexpected process for "+key)
			}
		}
	}
	return violations
}

func digestBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func (s Suite) writeArtifacts(adapter, caseID string, run RunResult, before, after map[string]fileState, events BoundaryEvents) (string, error) {
	root := filepath.Join(s.ArtifactRoot, adapter, caseID)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", err
	}
	for name, payload := range map[string][]byte{"stdout.bin": run.Stdout, "stderr.bin": run.Stderr} {
		if err := os.WriteFile(filepath.Join(root, name), payload, 0o600); err != nil {
			return "", err
		}
	}
	metadata := struct {
		ExitCode int                  `json:"exit_code"`
		Before   map[string]fileState `json:"before"`
		After    map[string]fileState `json:"after"`
		Boundary BoundaryEvents       `json:"boundary"`
	}{run.ExitCode, before, after, events}
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, "observation.json"), append(payload, '\n'), 0o600); err != nil {
		return "", err
	}
	return root, nil
}

// BinarySHA256 records exact executable bytes without loading the manager.
func BinarySHA256(name string) (string, error) {
	// #nosec G304 -- name is the explicit independently built manager binary
	// selected by the qualification caller and is only read for exact evidence.
	payload, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}
