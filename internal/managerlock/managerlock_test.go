package managerlock

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	helperModeEnv     = "CURATOR_MANAGERLOCK_HELPER_MODE"
	helperHomeEnv     = "CURATOR_MANAGERLOCK_HELPER_HOME"
	helperProjectEnv  = "CURATOR_MANAGERLOCK_HELPER_PROJECT"
	helperKeyEnv      = "CURATOR_MANAGERLOCK_HELPER_KEY"
	helperDeadlineEnv = "CURATOR_MANAGERLOCK_HELPER_DEADLINE"
)

const (
	blockedHelperDeadline  = 200 * time.Millisecond
	acquiredHelperDeadline = 30 * time.Second
)

func TestCanonicalProjectIdentitiesAndOrder(t *testing.T) {
	root := t.TempDir()
	projects := []string{
		filepath.Join(root, "文"),
		filepath.Join(root, "z"),
		filepath.Join(root, "é"),
	}
	for _, project := range projects {
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	identities, err := CanonicalProjects(projects...)
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index < len(identities); index++ {
		if bytes.Compare([]byte(identities[index-1]), []byte(identities[index])) >= 0 {
			t.Fatalf("identities are not in unsigned UTF-8 order: %#v", identities)
		}
	}
	manager, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	cleanupOperation(t, operation)
	held, err := operation.AcquireProjects(context.Background(), projects...)
	if err != nil {
		t.Fatal(err)
	}
	for index := range identities {
		if held[index] != identities[index] {
			t.Fatalf("acquisition order = %#v, want %#v", held, identities)
		}
	}

	alias := filepath.Join(root, "alias")
	if err := os.Symlink(projects[0], alias); err != nil {
		t.Logf("symlink identity check unavailable: %v", err)
		return
	}
	originalIdentity, err := CanonicalProject(projects[0])
	if err != nil {
		t.Fatal(err)
	}
	aliasIdentity, err := CanonicalProject(alias)
	if err != nil {
		t.Fatal(err)
	}
	if aliasIdentity != originalIdentity {
		t.Fatalf("symlink identity = %q, want %q", aliasIdentity, originalIdentity)
	}
	if _, err := CanonicalProjects(projects[0], alias); err == nil {
		t.Fatal("duplicate canonical project identities were accepted")
	}
}

func TestLockIdentitiesAreDeterministicAndManagerLocal(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	first, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := CanonicalProject(project)
	if err != nil {
		t.Fatal(err)
	}
	if first.projectLockPath(identity) != second.projectLockPath(identity) {
		t.Fatal("one canonical project produced different lock paths")
	}
	if first.buildKeyLockPath("sha256:logical") != second.buildKeyLockPath("sha256:logical") {
		t.Fatal("one logical build key produced different lock paths")
	}
	for _, lockPath := range []string{
		first.projectLockPath(identity),
		first.buildKeyLockPath("sha256:logical"),
		first.homeLockPath(),
	} {
		relative, relErr := filepath.Rel(first.Home(), lockPath)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("lock path %q is outside manager home %q", lockPath, first.Home())
		}
		if strings.HasPrefix(lockPath, project+string(filepath.Separator)) {
			t.Fatalf("lock path %q was placed inside project %q", lockPath, project)
		}
	}
}

// identityBelow spells the identity this platform assigns to a missing home
// under an already-resolved existing prefix.
//
// The suffix rule is not the same everywhere and must not be assumed here. On
// unix the identity is the path itself. On Windows a component whose containing
// directory performs case-insensitive lookup is folded into the identity, so the
// on-disk spelling is not the identity; identity_windows.go implements that rule
// and identity_windows_test.go asserts it directly. Spelling the expectation
// through the same rule keeps these portable cases about what they are for --
// which existing prefix was resolved, and whether the identity survives creation
// -- instead of re-asserting one platform's spelling on every host.
func identityBelow(t *testing.T, existing string, missing ...string) string {
	t.Helper()
	reversed := make([]string, 0, len(missing))
	for index := len(missing) - 1; index >= 0; index-- {
		reversed = append(reversed, missing[index])
	}
	identity, err := canonicalWithExistingPrefix(filepath.Clean(existing), reversed, "expected manager home")
	if err != nil {
		t.Fatalf("spell the expected identity below %q: %v", existing, err)
	}
	return identity
}

func TestMissingHomeUsesCanonicalExistingPrefix(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "missing", "curator-home")
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		t.Fatal(err)
	}

	before, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	want := identityBelow(t, resolvedBase, "missing", "curator-home")
	if before.Home() != want {
		t.Fatalf("missing home identity = %q, want canonical prefix identity %q", before.Home(), want)
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	after, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if after.Home() != before.Home() {
		t.Fatalf("home identity changed after creation: before %q, after %q", before.Home(), after.Home())
	}
	// The identity of the created home must also be the identity of the same
	// physical directory named without the alias the temporary root may carry,
	// which is the property the canonical existing prefix exists to give.
	resolved, err := New(filepath.Join(resolvedBase, "missing", "curator-home"))
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Home() != after.Home() {
		t.Fatalf("resolved twin identity = %q, want %q", resolved.Home(), after.Home())
	}
}

func TestMissingHomeBelowSymlinkKeepsIdentityAndContends(t *testing.T) {
	base := t.TempDir()
	realParent := filepath.Join(base, "real")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	aliasParent := filepath.Join(base, "alias")
	if err := os.Symlink(realParent, aliasParent); err != nil {
		t.Skipf("symlink canonical-prefix test unavailable: %v", err)
	}
	home := filepath.Join(aliasParent, "missing", "curator-home")

	before, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	resolvedRealParent, err := filepath.EvalSymlinks(realParent)
	if err != nil {
		t.Fatal(err)
	}
	wantHome := identityBelow(t, resolvedRealParent, "missing", "curator-home")
	if before.Home() != wantHome {
		t.Fatalf("missing aliased home identity = %q, want %q", before.Home(), wantHome)
	}
	homeLock, err := before.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}

	after, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	if after.Home() != before.Home() {
		t.Fatalf("aliased home identity changed after lock-state creation: before %q, after %q", before.Home(), after.Home())
	}
	if got := runHelper(t, "try-home", home, "", "blocked"); got != "blocked" {
		t.Fatalf("same physical home helper = %q, want blocked", got)
	}
	if err := homeLock.Close(); err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-home", home, "", "acquired"); got != "acquired" {
		t.Fatalf("released physical home helper = %q, want acquired", got)
	}
}

func TestDryRunCreatesNoLockState(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "missing", "curator-home")
	project := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(true)
	if _, err := operation.AcquireProjects(context.Background(), project); !errors.Is(err, ErrDryRun) {
		t.Fatalf("dry-run project acquisition error = %v", err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "key"); !errors.Is(err, ErrDryRun) {
		t.Fatalf("dry-run key acquisition error = %v", err)
	}
	if _, err := operation.AcquireHome(context.Background()); !errors.Is(err, ErrDryRun) {
		t.Fatalf("dry-run home acquisition error = %v", err)
	}
	if _, err := manager.AcquireHomeOnly(context.Background(), true); !errors.Is(err, ErrDryRun) {
		t.Fatalf("dry-run home-only acquisition error = %v", err)
	}
	if err := operation.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created manager lock state at %q: %v", home, err)
	}
}

func TestConcurrentFirstUsePreparesOneManagerIdentity(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "missing", "curator-home")
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 8
	projects := make([]string, workers)
	for index := range projects {
		projects[index] = t.TempDir()
	}
	start := make(chan struct{})
	errorsByWorker := make(chan error, workers)
	var wait sync.WaitGroup
	for _, project := range projects {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			operation := manager.NewOperation(false)
			if _, err := operation.AcquireProjects(context.Background(), project); err != nil {
				errorsByWorker <- err
				return
			}
			if err := operation.Close(); err != nil {
				errorsByWorker <- err
			}
		}()
	}
	close(start)
	wait.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Errorf("concurrent first-use acquisition: %v", err)
	}
	if _, err := os.Stat(manager.Home()); err != nil {
		t.Fatalf("prepared manager home: %v", err)
	}
}

func TestProjectOrderInversionFailsBeforeWaiting(t *testing.T) {
	home := t.TempDir()
	root := t.TempDir()
	paths := []string{filepath.Join(root, "a"), filepath.Join(root, "b")}
	for _, path := range paths {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	ordered, err := CanonicalProjects(paths...)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	high := manager.NewOperation(false)
	cleanupOperation(t, high)
	if _, err := high.AcquireProjects(context.Background(), string(ordered[1])); err != nil {
		t.Fatal(err)
	}
	low := manager.NewOperation(false)
	cleanupOperation(t, low)
	if _, err := low.AcquireProjects(context.Background(), string(ordered[0])); err != nil {
		t.Fatal(err)
	}

	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := high.AcquireProjects(ctx, string(ordered[0])); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("inversion error = %v, want ErrLockOrder", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("inversion waited on the contended lower lock for %v", elapsed)
	}
}

func TestBuildKeyMustBeReleasedBeforeHome(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	cleanupOperation(t, operation)
	if _, err := operation.AcquireProjects(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "sha256:key-one"); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.AcquireHome(context.Background()); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("home nested with key error = %v", err)
	}
	if _, err := manager.AcquireHomeOnly(context.Background(), false); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("home-only nested with in-process key error = %v", err)
	}
	if err := operation.ReleaseBuildKey(); err != nil {
		t.Fatal(err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "sha256:key-two"); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("second build key error = %v", err)
	}
	homeLock, err := operation.AcquireHome(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := homeLock.AssertHeld(); err != nil {
		t.Fatalf("held home witness failed: %v", err)
	}
	otherProject := manager.NewOperation(false)
	cleanupOperation(t, otherProject)
	if _, err := otherProject.AcquireProjects(context.Background(), t.TempDir()); !errors.Is(err, ErrLockOrder) {
		t.Fatalf("project acquired while process held home lock: %v", err)
	}
	if err := homeLock.Close(); err != nil {
		t.Fatal(err)
	}
	if err := homeLock.AssertHeld(); !errors.Is(err, ErrNotHeld) {
		t.Fatalf("released home witness error = %v", err)
	}
}

func TestCancellationReleasesPartialAcquisition(t *testing.T) {
	home := t.TempDir()
	root := t.TempDir()
	firstProject := filepath.Join(root, "a")
	secondProject := filepath.Join(root, "b")
	for _, project := range []string{firstProject, secondProject} {
		if err := os.Mkdir(project, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	blocker := manager.NewOperation(false)
	if _, err := blocker.AcquireProjects(context.Background(), secondProject); err != nil {
		t.Fatal(err)
	}
	waiter := manager.NewOperation(false)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := waiter.AcquireProjects(ctx, firstProject, secondProject); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contended batch error = %v", err)
	}
	if err := blocker.Close(); err != nil {
		t.Fatal(err)
	}
	probe := manager.NewOperation(false)
	cleanupOperation(t, probe)
	if _, err := probe.AcquireProjects(context.Background(), firstProject); err != nil {
		t.Fatalf("partial first lock leaked after cancellation: %v", err)
	}
}

func TestReleasedLockFileRemainsAndCanBeReacquired(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := CanonicalProject(project)
	if err != nil {
		t.Fatal(err)
	}
	path := manager.projectLockPath(identity)
	operation := manager.NewOperation(false)
	if _, err := operation.AcquireProjects(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	if err := operation.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stable lock file was removed on release: %v", err)
	}
	reacquired := manager.NewOperation(false)
	cleanupOperation(t, reacquired)
	if _, err := reacquired.AcquireProjects(context.Background(), project); err != nil {
		t.Fatalf("stable lock file could not be reacquired: %v", err)
	}
}

func TestSubprocessContentionAndIndependentProjects(t *testing.T) {
	home := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	if _, err := operation.AcquireProjects(context.Background(), projectA); err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-project", home, projectA, "blocked"); got != "blocked" {
		t.Fatalf("same project helper = %q, want blocked", got)
	}
	if got := runHelper(t, "try-project", home, projectB, "acquired"); got != "acquired" {
		t.Fatalf("independent project helper = %q, want acquired", got)
	}
	if err := operation.Close(); err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-project", home, projectA, "acquired"); got != "acquired" {
		t.Fatalf("released project helper = %q, want acquired", got)
	}

	homeLock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-home", home, "", "blocked"); got != "blocked" {
		t.Fatalf("same home helper = %q, want blocked", got)
	}
	if err := homeLock.Close(); err != nil {
		t.Fatal(err)
	}
	if got := runHelper(t, "try-home", home, "", "acquired"); got != "acquired" {
		t.Fatalf("released home helper = %q, want acquired", got)
	}
}

func TestSubprocessBuildKeyDeduplicationAcrossProjects(t *testing.T) {
	home := t.TempDir()
	projectA := t.TempDir()
	projectB := t.TempDir()
	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	cleanupOperation(t, operation)
	if _, err := operation.AcquireProjects(context.Background(), projectA); err != nil {
		t.Fatal(err)
	}
	if err := operation.AcquireBuildKey(context.Background(), "sha256:shared"); err != nil {
		t.Fatal(err)
	}
	if got := runKeyHelper(t, home, projectB, "sha256:shared", "blocked"); got != "blocked" {
		t.Fatalf("same build key helper = %q, want blocked", got)
	}
	if got := runKeyHelper(t, home, projectB, "sha256:independent", "acquired"); got != "acquired" {
		t.Fatalf("independent build key helper = %q, want acquired", got)
	}
	if err := operation.ReleaseBuildKey(); err != nil {
		t.Fatal(err)
	}
	if got := runKeyHelper(t, home, projectB, "sha256:shared", "acquired"); got != "acquired" {
		t.Fatalf("released build key helper = %q, want acquired", got)
	}
}

func TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked(t *testing.T) {
	if acquiredHelperDeadline <= blockedHelperDeadline {
		t.Fatalf("acquired helper deadline %v must exceed blocked helper deadline %v", acquiredHelperDeadline, blockedHelperDeadline)
	}
	home := t.TempDir()
	project := t.TempDir()
	if got := runHelperWithDeadline(t, "try-project", home, project, "", time.Nanosecond); got != "blocked" {
		t.Fatalf("uncontended helper with tiny deadline = %q, want blocked", got)
	}
}

func TestAbnormalChildExitReleasesOSLock(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	command := helperCommand("hold-project", home, project)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = os.Stderr
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
	})
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("child did not report held lock: %q (%v)", scanner.Text(), scanner.Err())
	}
	if err := command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Fatal("killed child exited successfully")
	}

	manager, err := New(home)
	if err != nil {
		t.Fatal(err)
	}
	operation := manager.NewOperation(false)
	cleanupOperation(t, operation)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := operation.AcquireProjects(ctx, project); err != nil {
		t.Fatalf("OS lock survived abnormal child exit: %v", err)
	}
}

func runHelper(t *testing.T, mode, home, project, expected string) string {
	t.Helper()
	return runHelperWithDeadline(t, mode, home, project, "", helperDeadline(t, expected))
}

func runKeyHelper(t *testing.T, home, project, key, expected string) string {
	t.Helper()
	return runHelperWithDeadline(t, "try-key", home, project, key, helperDeadline(t, expected))
}

func helperDeadline(t *testing.T, expected string) time.Duration {
	t.Helper()
	switch expected {
	case "blocked":
		return blockedHelperDeadline
	case "acquired":
		return acquiredHelperDeadline
	default:
		t.Fatalf("unknown expected helper outcome %q", expected)
		return 0
	}
}

func runHelperWithDeadline(t *testing.T, mode, home, project, key string, deadline time.Duration) string {
	t.Helper()
	command := helperCommand(mode, home, project)
	command.Env = append(command.Env, helperDeadlineEnv+"="+deadline.String())
	if key != "" {
		command.Env = append(command.Env, helperKeyEnv+"="+key)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("helper %s failed: %v\n%s", mode, err, output)
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(output)), "\n")
	return line
}

func helperCommand(mode, home, project string) *exec.Cmd {
	command := exec.Command(os.Args[0], "-test.run=^TestManagerLockHelper$")
	command.Env = append(os.Environ(),
		helperModeEnv+"="+mode,
		helperHomeEnv+"="+home,
		helperProjectEnv+"="+project,
	)
	return command
}

func TestManagerLockHelper(t *testing.T) {
	mode := os.Getenv(helperModeEnv)
	if mode == "" {
		return
	}
	manager, err := New(os.Getenv(helperHomeEnv))
	if err != nil {
		t.Fatal(err)
	}
	if mode == "hold-project" {
		operation := manager.NewOperation(false)
		cleanupOperation(t, operation)
		if _, err := operation.AcquireProjects(context.Background(), os.Getenv(helperProjectEnv)); err != nil {
			t.Fatalf("hold project lock: %v", err)
		}
		fmt.Println("ready")
		_ = os.Stdout.Sync()
		_, _ = io.Copy(io.Discard, os.Stdin)
		return
	}
	deadline, err := time.ParseDuration(os.Getenv(helperDeadlineEnv))
	if err != nil || deadline <= 0 {
		t.Fatalf("invalid helper deadline %q: %v", os.Getenv(helperDeadlineEnv), err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	var release func() error
	switch mode {
	case "try-project":
		operation := manager.NewOperation(false)
		release = operation.Close
		_, err = operation.AcquireProjects(ctx, os.Getenv(helperProjectEnv))
	case "try-home":
		var lock *HomeLock
		lock, err = manager.AcquireHomeOnly(ctx, false)
		if lock != nil {
			release = lock.Close
		}
	case "try-key":
		operation := manager.NewOperation(false)
		release = operation.Close
		if _, err = operation.AcquireProjects(ctx, os.Getenv(helperProjectEnv)); err == nil {
			err = operation.AcquireBuildKey(ctx, os.Getenv(helperKeyEnv))
		}
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}
	if release != nil {
		if releaseErr := release(); releaseErr != nil {
			t.Fatalf("release helper lock: %v", releaseErr)
		}
	}
	if err == nil {
		fmt.Println("acquired")
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("blocked")
		return
	}
	t.Fatalf("helper lock error: %v", err)
}

func cleanupOperation(t *testing.T, operation *Operation) {
	t.Helper()
	t.Cleanup(func() {
		if err := operation.Close(); err != nil {
			t.Errorf("release manager operation locks: %v", err)
		}
	})
}
