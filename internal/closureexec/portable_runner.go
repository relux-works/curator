package closureexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/relux-works/curator/internal/closuregraph"
)

// ManagerProcessRunner is the portable CLI process implementation. It reports
// only exit status and the manager-owned output root; Executor independently
// hashes and reconciles every declared output after the process exits.
type ManagerProcessRunner struct {
	ExecutionRoot  string
	ExecutableRoot string
	OutputRoot     string
	// ProcessStartObserver is an instrumentation-only notification immediately
	// before the sole portable process-start seam.
	ProcessStartObserver ProcessStartObserver
	// ProcessLaunchObserver exposes the exact constructed command at that seam.
	ProcessLaunchObserver ProcessLaunchObserver
}

// NewManagerProcessRunner validates manager-owned absolute roots. The caller
// prepares the trusted executable below executionRoot. Run creates immutable
// replay copies and reserves outputRoot for declared process output.
func NewManagerProcessRunner(executionRoot, outputRoot string) (*ManagerProcessRunner, error) {
	return NewManagerProcessRunnerWithExecutableRoot(executionRoot, executionRoot, outputRoot)
}

// NewManagerProcessRunnerWithExecutableRoot separates the read-only trusted
// tool root from the writable task execution root.
func NewManagerProcessRunnerWithExecutableRoot(executionRoot, executableRoot, outputRoot string) (*ManagerProcessRunner, error) {
	executionAbs, err := filepath.Abs(executionRoot)
	if err != nil {
		return nil, err
	}
	outputAbs, err := filepath.Abs(outputRoot)
	if err != nil {
		return nil, err
	}
	executableAbs, err := filepath.Abs(executableRoot)
	if err != nil || executableRoot == "" {
		return nil, fmt.Errorf("portable executable root is invalid")
	}
	if executionRoot == "" || outputRoot == "" || executionAbs == outputAbs || !pathWithin(executionAbs, outputAbs) {
		return nil, fmt.Errorf("portable output root must be a distinct child of the execution root")
	}
	return &ManagerProcessRunner{ExecutionRoot: executionAbs, ExecutableRoot: executableAbs, OutputRoot: outputAbs}, nil
}

// Run executes one manager-owned portable operation. Portable mode proves the
// immutable replay copies, direct executable identity, deadline, combined
// output bound, and complete contents of OutputRoot. It deliberately makes no
// claim that the host observed every read, write, process, or network attempt.
func (runner *ManagerProcessRunner) Run(ctx context.Context, request ExecutionRequest) (PortableRunResult, error) {
	if runner == nil {
		return PortableRunResult{}, failure("portable_runner_missing", "portable runner is nil")
	}
	if err := ensureEmptyDirectory(runner.OutputRoot); err != nil {
		return PortableRunResult{}, err
	}
	boundOutput, outputErr := resolveManagerPath(runner.ExecutionRoot, request.Permit.Environment["CURATOR_OUTPUT_ROOT"])
	if outputErr != nil || boundOutput != runner.OutputRoot {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "portable output root is not bound in the permit environment")
	}
	executable, err := safeExecutionPath(runner.ExecutableRoot, request.Permit.Executable)
	if err != nil {
		return PortableRunResult{}, err
	}
	cwd, err := safeExecutionPath(runner.ExecutionRoot, request.Permit.CWD)
	if err != nil {
		return PortableRunResult{}, err
	}
	if err = verifyPortableExecutable(executable, request.Permit.ExecutableSHA256); err != nil {
		return PortableRunResult{}, err
	}
	verifyReplays := make([]func() error, len(request.Inputs))
	preparedTargets := make([]string, 0, len(request.Inputs))
	retainedTargets := map[string]bool{}
	for index, input := range request.Inputs {
		verifyReplays[index], err = runner.prepareReplay(input)
		if err != nil {
			return PortableRunResult{}, err
		}
		target, pathErr := safeExecutionPath(runner.ExecutionRoot, input.MountPath)
		if pathErr != nil {
			return PortableRunResult{}, pathErr
		}
		preparedTargets = append(preparedTargets, target)
	}
	inputsByReceipt := make(map[closuregraph.ID]ReplayInput, len(request.Inputs))
	for _, input := range request.Inputs {
		inputsByReceipt[input.ReceiptID] = input
	}
	for _, work := range request.Permit.WorkCopies {
		input, ok := inputsByReceipt[work.ReceiptID]
		if !ok {
			return PortableRunResult{}, failure("closure_derivation_unauthorized", "work copy names an unavailable admitted input")
		}
		if err = runner.prepareWorkCopy(input, work.Path); err != nil {
			return PortableRunResult{}, err
		}
		target, pathErr := safeExecutionPath(runner.ExecutionRoot, work.Path)
		if pathErr != nil {
			return PortableRunResult{}, pathErr
		}
		preparedTargets = append(preparedTargets, target)
		if work.Retain {
			retainedTargets[target] = true
		}
	}
	cleanup := func(includeRetained bool) {
		for _, target := range preparedTargets {
			if !includeRetained && retainedTargets[target] {
				continue
			}
			_ = filepath.WalkDir(target, func(current string, entry fs.DirEntry, walkErr error) error {
				if walkErr == nil {
					if entry.IsDir() {
						_ = os.Chmod(current, 0o700) // #nosec G302 -- cleanup must restore owner traversal before removing private replay trees.
					} else {
						_ = os.Chmod(current, 0o600)
					}
				}
				return nil
			})
			_ = os.RemoveAll(target)
		}
	}
	cleanupOnReturn := true
	defer func() {
		if cleanupOnReturn {
			cleanup(true)
		}
	}()
	if info, statErr := os.Lstat(cwd); statErr != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return PortableRunResult{}, failure("closure_derivation_unauthorized", "portable cwd is not a real directory")
	}

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(request.Permit.ResourceLimits.WallTimeMillis)*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(runCtx, executable, request.Permit.Argv...) // #nosec G204 -- executable and argv are committed manager permit authority.
	configurePortableProcess(command)
	command.Dir = cwd
	keys := make([]string, 0, len(request.Permit.Environment))
	for key := range request.Permit.Environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	command.Env = make([]string, 0, len(keys))
	for _, key := range keys {
		command.Env = append(command.Env, key+"="+request.Permit.Environment[key])
	}
	budget := &portableOutputBudget{remaining: request.Permit.ResourceLimits.OutputBytes, cancel: cancel}
	stdout := &portableBoundedBuffer{budget: budget}
	stderr := &portableBoundedBuffer{budget: budget}
	command.Stdout = stdout
	command.Stderr = stderr
	if runner.ProcessStartObserver != nil {
		runner.ProcessStartObserver(request.Permit)
	}
	if runner.ProcessLaunchObserver != nil {
		runner.ProcessLaunchObserver(ProcessLaunch{Executable: command.Path, CWD: command.Dir, Argv: append([]string(nil), command.Args[1:]...), Environment: append([]string(nil), command.Env...)})
	}
	if err = command.Start(); err != nil {
		return PortableRunResult{}, err
	}
	processDone := make(chan struct{})
	go func() {
		select {
		case <-runCtx.Done():
			terminatePortableProcess(command)
		case <-processDone:
		}
	}()
	waitErr := command.Wait()
	close(processDone)
	terminatePortableProcess(command)
	if budget.exceeded() {
		return PortableRunResult{}, failure("portable_output_limit", "portable process exceeded its combined output bound")
	}
	if runCtx.Err() != nil {
		return PortableRunResult{}, failure("portable_process_timeout", "portable process exceeded its deadline")
	}
	evidenceRoot := runner.OutputRoot
	if selected := request.Permit.Environment["CURATOR_EVIDENCE_ROOT"]; selected != "" {
		selected, pathErr := filepath.Abs(selected)
		if pathErr != nil || selected != runner.ExecutionRoot {
			return PortableRunResult{}, failure("closure_derivation_unauthorized", "portable evidence root is not the manager execution root")
		}
		evidenceRoot = selected
	}
	if waitErr == nil && request.Permit.StdoutEvidencePath != "" {
		stdoutPath, pathErr := safeExecutionPath(evidenceRoot, request.Permit.StdoutEvidencePath)
		if pathErr != nil {
			return PortableRunResult{}, pathErr
		}
		if pathErr = os.MkdirAll(filepath.Dir(stdoutPath), 0o700); pathErr != nil {
			return PortableRunResult{}, pathErr
		}
		if pathErr = os.WriteFile(stdoutPath, stdout.buffer.Bytes(), 0o600); pathErr != nil {
			return PortableRunResult{}, pathErr
		}
	}
	if waitErr == nil && evidenceRoot != runner.OutputRoot {
		for _, expected := range request.Permit.ExpectedEvidence {
			source, pathErr := safeExecutionPath(evidenceRoot, expected.Path)
			if pathErr != nil {
				return PortableRunResult{}, pathErr
			}
			info, pathErr := os.Lstat(source)
			if pathErr != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
				return PortableRunResult{}, failure("closure_derivation_drift", "declared portable evidence %s is absent or non-regular", expected.Path)
			}
			payload, pathErr := os.ReadFile(source) // #nosec G304 -- exact permit evidence below the manager execution root.
			if pathErr != nil {
				return PortableRunResult{}, pathErr
			}
			destination, pathErr := safeExecutionPath(runner.OutputRoot, expected.Path)
			if pathErr != nil {
				return PortableRunResult{}, pathErr
			}
			if pathErr = os.MkdirAll(filepath.Dir(destination), 0o700); pathErr != nil {
				return PortableRunResult{}, pathErr
			}
			mode := fs.FileMode(0o600)
			if info.Mode().Perm()&0o100 != 0 {
				mode = 0o500
			}
			if pathErr = os.WriteFile(destination, payload, mode); pathErr != nil {
				return PortableRunResult{}, pathErr
			}
		}
		evidenceRoot = runner.OutputRoot
	}
	for _, verify := range verifyReplays {
		if err = verify(); err != nil {
			return PortableRunResult{}, err
		}
	}
	if waitErr == nil {
		cleanupOnReturn = false
		return PortableRunResult{ExitCode: 0, OutputRoot: runner.OutputRoot, EvidenceRoot: evidenceRoot, cleanup: func() { cleanup(false) }}, nil
	}
	if exit, ok := waitErr.(*exec.ExitError); ok {
		return PortableRunResult{}, failure("closure_derivation_drift", "portable process exited %d: %s", exit.ExitCode(), strings.TrimSpace(stderr.buffer.String()))
	}
	return PortableRunResult{}, waitErr
}

func resolveManagerPath(root, value string) (string, error) {
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return safeExecutionPath(root, value)
}

func (runner *ManagerProcessRunner) prepareReplay(input ReplayInput) (func() error, error) {
	source, err := input.ProtectedPath()
	if err != nil {
		return nil, err
	}
	target, err := safeExecutionPath(runner.ExecutionRoot, input.MountPath)
	if err != nil {
		return nil, err
	}
	insideOutput := target == runner.OutputRoot || pathWithin(runner.OutputRoot, target)
	if insideOutput || pathWithin(target, runner.OutputRoot) {
		return nil, failure("closure_derivation_unauthorized", "replay overlaps the portable output root")
	}
	if _, err = os.Lstat(target); !errors.Is(err, fs.ErrNotExist) {
		return nil, failure("closure_input_undeclared", "portable replay target is not absent")
	}
	if input.IsTree() {
		if err = copyReplayTree(source, target); err != nil {
			return nil, err
		}
		return func() error { return verifyReplayTree(target, input.input.Tree.files) }, nil
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return nil, err
	}
	payload, err := os.ReadFile(source) // #nosec G304 -- source is an opaque manager-protected intake path.
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(target, payload, 0o400); err != nil {
		return nil, err
	}
	digest, size := input.input.Handle.digest, input.input.Handle.size
	return func() error { return verifyReplayFile(target, digest, size) }, nil
}

func copyReplayTree(source, target string) error {
	if err := os.MkdirAll(target, 0o700); err != nil {
		return err
	}
	err := filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, current)
		if err != nil || rel == "." {
			return err
		}
		destination := filepath.Join(target, rel)
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("closure_input_undeclared", "protected replay contains a link")
		}
		if entry.IsDir() {
			return os.Mkdir(destination, 0o700)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return failure("closure_input_undeclared", "protected replay contains a special node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the protected replay root.
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o400)
		if info.Mode().Perm()&0o100 != 0 {
			mode = 0o500
		}
		return os.WriteFile(destination, payload, mode)
	})
	if err != nil {
		return err
	}
	return filepath.WalkDir(target, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return os.Chmod(current, 0o500) // #nosec G302 -- immutable replay directories remain traversable.
		}
		return nil
	})
}

func (runner *ManagerProcessRunner) prepareWorkCopy(input ReplayInput, logical string) error {
	if !input.IsTree() {
		return failure("closure_derivation_unauthorized", "work copies require an admitted tree")
	}
	source, err := safeExecutionPath(runner.ExecutionRoot, input.MountPath)
	if err != nil {
		return err
	}
	target, err := safeExecutionPath(runner.ExecutionRoot, logical)
	if err != nil {
		return err
	}
	if target == runner.OutputRoot || pathWithin(runner.OutputRoot, target) || pathWithin(target, runner.OutputRoot) {
		return failure("closure_derivation_unauthorized", "work copy overlaps the portable output root")
	}
	if _, err = os.Lstat(target); !errors.Is(err, fs.ErrNotExist) {
		return failure("closure_write_undeclared", "work-copy target is not absent")
	}
	if err = os.MkdirAll(target, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, current)
		if err != nil || rel == "." {
			return err
		}
		destination := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.Mkdir(destination, 0o700)
		}
		if !entry.Type().IsRegular() {
			return failure("closure_input_undeclared", "work-copy seed contains a non-regular member")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- immutable replay seed below execution root.
		if err != nil {
			return err
		}
		return os.WriteFile(destination, payload, 0o600)
	})
}

func verifyReplayTree(root string, expected []SnapshotFile) error {
	actual := make([]SnapshotFile, 0, len(expected))
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			info, infoErr := entry.Info()
			if infoErr != nil || info.Mode().Perm()&0o222 != 0 {
				return failure("closure_input_mutated", "portable replay directory became writable")
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("closure_input_mutated", "portable replay contains a link after execution")
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o222 != 0 {
			return failure("closure_input_mutated", "portable replay member %s changed type", current)
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the replay root.
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, current)
		actual = append(actual, SnapshotFile{Path: filepath.ToSlash(rel), SHA256: digestBytes(payload), Size: int64(len(payload)), Executable: info.Mode().Perm()&0o100 != 0})
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(actual, func(i, j int) bool { return actual[i].Path < actual[j].Path })
	if !equalSnapshotFiles(actual, expected) {
		return failure("closure_input_mutated", "portable replay changed during execution: expected=%v actual=%v", expected, actual)
	}
	return nil
}

func equalSnapshotFiles(left, right []SnapshotFile) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func verifyReplayFile(path string, digest closuregraph.ID, size int64) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o222 != 0 || info.Size() != size {
		return failure("closure_input_mutated", "portable replay changed during execution")
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- path is a manager-created replay target.
	if err != nil || digestBytes(payload) != digest {
		return failure("closure_input_mutated", "portable replay bytes changed during execution")
	}
	return nil
}

func verifyPortableExecutable(path string, expected closuregraph.ID) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return failure("artifact_toolchain_identity_changed", "portable executable is not a regular file")
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- path is below the manager execution root and committed by the permit.
	if err != nil || digestBytes(payload) != expected {
		return failure("artifact_toolchain_identity_changed", "portable executable bytes differ from the permit")
	}
	return nil
}

func safeExecutionPath(root, logical string) (string, error) {
	if err := portablePath(logical); err != nil {
		return "", failure("closure_derivation_unauthorized", "portable process path is invalid")
	}
	path := filepath.Join(root, filepath.FromSlash(logical))
	if !pathWithin(root, path) {
		return "", failure("closure_derivation_unauthorized", "portable process path escapes execution root")
	}
	return path, nil
}

func ensureEmptyDirectory(root string) error {
	info, err := os.Lstat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(root, 0o700)
		}
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return failure("closure_input_undeclared", "portable output root is not a private real directory")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return failure("closure_input_undeclared", "portable output root is not empty")
	}
	return nil
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && relative[:min(len(relative), 3)] != ".."+string(filepath.Separator)
}

type portableOutputBudget struct {
	mu        sync.Mutex
	remaining int64
	cancel    context.CancelFunc
	over      bool
}

func (budget *portableOutputBudget) exceeded() bool {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	return budget.over
}

type portableBoundedBuffer struct {
	buffer bytes.Buffer
	budget *portableOutputBudget
}

func (buffer *portableBoundedBuffer) Write(payload []byte) (int, error) {
	buffer.budget.mu.Lock()
	defer buffer.budget.mu.Unlock()
	if int64(len(payload)) > buffer.budget.remaining {
		buffer.budget.over = true
		buffer.budget.cancel()
		return 0, errors.New("portable output limit exceeded")
	}
	buffer.budget.remaining -= int64(len(payload))
	return buffer.buffer.Write(payload)
}
