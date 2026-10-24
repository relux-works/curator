//go:build unix

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/godriver"
)

const nativeBuildPauseMarker = ".curator-test-build-pause"

// The test executable observes and forwards unchanged worker frames. A marked
// fixture stops its own worker before forwarding the authenticated build permit;
// the production worker still validates and executes the real Go build after
// release. A cached compiler cannot outrun this test-owned barrier.
func runNativeBuildWorker() int {
	reader, writer, err := os.Pipe()
	if err != nil {
		return 95
	}
	defer func() { _ = reader.Close() }()
	// The proxy also owns the original protocol input; preserve its close-on-exec
	// guarantee even though RunWorker receives the proxy's read descriptor.
	syscall.CloseOnExec(int(os.Stdin.Fd()))
	go func() {
		defer func() { _ = writer.Close() }()
		_ = forwardNativeBuildFrames(os.Stdin, writer, func() error {
			return syscall.Kill(os.Getpid(), syscall.SIGSTOP)
		})
	}()
	return godriver.RunWorker(reader, os.Stdout)
}

func forwardNativeBuildFrames(input io.Reader, output io.Writer, pause func() error) error {
	pauseOnPermit := false
	for {
		var header [4]byte
		if _, err := io.ReadFull(input, header[:]); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		length := binary.BigEndian.Uint32(header[:])
		if length == 0 || length > 64*1024*1024 {
			return fmt.Errorf("invalid test worker frame length")
		}
		payload := make([]byte, int(length))
		if _, err := io.ReadFull(input, payload); err != nil {
			return err
		}
		var frame struct {
			Kind    string `json:"kind"`
			Request *struct {
				Directory string `json:"directory"`
			} `json:"request"`
		}
		if err := json.Unmarshal(payload, &frame); err != nil {
			return err
		}
		if frame.Kind == "request" && frame.Request != nil {
			marker, err := os.ReadFile(filepath.Join(frame.Request.Directory, nativeBuildPauseMarker))
			pauseOnPermit = err == nil && string(marker) == "pause\n"
		}
		if pauseOnPermit && frame.Kind == "permit" {
			if err := pause(); err != nil {
				return err
			}
			pauseOnPermit = false
		}
		if _, err := output.Write(header[:]); err != nil {
			return err
		}
		if _, err := output.Write(payload); err != nil {
			return err
		}
	}
}

// Marked fixtures stop exactly at the permit, before any permit bytes are
// forwarded. Unmarked/wrong-marker streams pass unchanged without a pause;
// a failed barrier never forwards a permit. This tests instrumentation only.
func TestNativeBuildPermitBarrier(t *testing.T) {
	for _, marker := range []string{"pause\n", "", "other\n"} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("marked=%t/failure=%t/marker=%q", marker == "pause\n", fail, marker), func(t *testing.T) {
				dir := t.TempDir()
				if marker != "" {
					writeFile(t, filepath.Join(dir, nativeBuildPauseMarker), marker)
				}
				var input bytes.Buffer
				frame := func(value any) {
					payload, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					var header [4]byte
					binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
					input.Write(header[:])
					input.Write(payload)
				}
				frame(map[string]any{"kind": "request", "request": map[string]any{"directory": dir}})
				frame(map[string]any{"kind": "list"})
				boundary := input.Len()
				frame(map[string]any{"kind": "permit", "permit": "fixture-only"})
				frame(map[string]any{"kind": "shutdown"})
				original := append([]byte(nil), input.Bytes()...)
				var output bytes.Buffer
				calls := 0
				planted := errors.New("barrier refused")
				err := forwardNativeBuildFrames(&input, &output, func() error {
					calls++
					if output.Len() != boundary {
						t.Fatal("permit forwarded before barrier")
					}
					if fail {
						return planted
					}
					return nil
				})
				expected := 0
				if marker == "pause\n" {
					expected = 1
				}
				if calls != expected {
					t.Fatalf("build permit pause count = %d, want %d", calls, expected)
				}
				if marker == "pause\n" && fail {
					if !errors.Is(err, planted) || !bytes.Equal(output.Bytes(), original[:boundary]) {
						t.Fatalf("failed barrier forwarded permit: %v", err)
					}
				} else if err != nil || !bytes.Equal(output.Bytes(), original) {
					t.Fatalf("worker frames altered: %v", err)
				}
			})
		}
	}
}

// A test-owned barrier holds the real worker at its authenticated build permit.
// Profile sync stays available with no premature publication. After release,
// the real compiler publishes the skill; remove/add also succeeds normally.
// Uncovered hosts assert their refusal separately and declare this carve-out.
func TestGlobalNativeOwnerPausedBuild(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	fixture := setupNativeConcurrent(t, false)
	compiled := t.TempDir()
	writeFile(t, filepath.Join(compiled, "assets", "build-tool", "cmd", "tool", nativeBuildPauseMarker), "pause\n")
	writeCompiledSkillRepo(t, compiled)
	// Prepare both tags before the first clone. The later add is an unpaused
	// positive control using the same real repository without a cache refresh.
	if err := os.Remove(filepath.Join(compiled, "assets", "build-tool", "cmd", "tool", nativeBuildPauseMarker)); err != nil {
		t.Fatal(err)
	}
	runGit(t, compiled, "add", ".")
	runGit(t, compiled, "commit", "-qm", "remove test barrier")
	runGit(t, compiled, "tag", "unpaused")
	configureNativeCompiledSkill(t, fixture, compiled)
	child := startNativeCommand(t, fixture, "global", "install")
	pid := waitNativeBuildWorker(t, child)
	resumed := false
	t.Cleanup(func() {
		if !resumed {
			_ = syscall.Kill(pid, syscall.SIGCONT)
		}
	})
	code, stdout, stderr := runFile(t, fixture.config, "profile", "sync")
	if code != exitOK {
		t.Fatalf("profile sync blocked by private build: %d %s%s", code, stdout, stderr)
	}
	for _, path := range []string{".curator/global/skills/build-skill", ".codex/skills/build-skill", ".claude/skills/build-skill"} {
		if _, err := os.Lstat(filepath.Join(fixture.home, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Fatalf("paused build published state: %s: %v", path, err)
		}
	}
	if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
		t.Fatal(err)
	}
	resumed = true
	// Completion is not a latency assertion: the production worker permits a
	// two-minute real build, followed by toolchain verification/publication.
	// Lock-order assertions happen at the paused barrier, before this wait.
	released := time.Now()
	child.waitWithin(t, nativeCommandWatchdog)
	t.Logf("released real build completed in %s", time.Since(released))
	requireCompiledNativeOwner(t, fixture)
	fixture.command(t, exitOK, "global", "remove", "build-skill")
	fixture.command(t, exitOK, "global", "install")
	if _, err := os.Lstat(filepath.Join(fixture.home, ".curator", "global", "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("removed compiled skill remained canonical: %v", err)
	}
	fixture.command(t, exitOK, "global", "add", "build-skill", "--git", "https://example.com/build-skill", "--tag", "unpaused")
	requireCompiledNativeOwner(t, fixture)
}

func waitNativeBuildWorker(t *testing.T, child *nativeCommand) int {
	t.Helper()
	deadline := time.Now().Add(nativeCommandWatchdog)
	for time.Now().Before(deadline) {
		select {
		case err := <-child.done:
			child.finished = true
			t.Fatalf("CLI ended before its build permit barrier: %v\n%s", err, child.output.String())
		default:
		}
		output, err := exec.Command("ps", "-axo", "pid=,ppid=,stat=").Output()
		if err != nil {
			t.Fatal(err)
		}
		descendants := map[int]bool{child.cmd.Process.Pid: true}
		rows := strings.Split(string(output), "\n")
		for changed := true; changed; {
			changed = false
			for _, row := range rows {
				fields := strings.Fields(row)
				if len(fields) != 3 {
					continue
				}
				pid, _ := strconv.Atoi(fields[0])
				parent, _ := strconv.Atoi(fields[1])
				if !descendants[parent] || descendants[pid] {
					continue
				}
				descendants[pid] = true
				changed = true
				if !strings.Contains(fields[2], "T") {
					continue
				}
				args, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=").Output()
				if err != nil || !strings.Contains(string(args), godriver.WorkerMode) {
					continue
				}
				t.Log("owned worker stopped at authenticated build permit")
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("owned build worker did not reach its persistent permit barrier")
	return 0
}

// The fixture completion watchdog must cover the production two-minute build
// plus publication. Lock-order negatives remain at the unreleased barrier.
func TestNativeCommandWatchdogBudget(t *testing.T) {
	if nativeCommandWatchdog <= 2*time.Minute {
		t.Fatal("native fixture watchdog does not cover production build and publication")
	}
}
