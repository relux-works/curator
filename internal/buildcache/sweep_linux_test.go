package buildcache

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/relux-works/curator/internal/buildmeta"
)

// Claim: a real non-root sweep ignores inaccessible foreign credentials, keeps
// this user's running build (including an unlinked alias), and can collect it
// once stopped. A non-dumpable same-user image refuses the entire sweep.
func TestLinuxUnprivilegedSweep(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("this qualification must run as a non-root user")
	}
	status, err := os.ReadFile("/proc/1/status")
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := linuxForeignUID(string(status), os.Geteuid())
	if err != nil || !foreign {
		t.Fatalf("PID 1 must provide a real other-UID control: foreign=%v err=%v", foreign, err)
	}
	_, foreignErr := os.Readlink("/proc/1/exe")
	t.Logf("non-root uid=%d, other-UID PID 1 exe readable=%v", os.Geteuid(), foreignErr == nil)
	_, baselineErr := nativeExecutableSnapshot()
	if baselineErr != nil {
		if os.Getenv("CURATOR_SWEEP_REQUIRE_COMPLETE") == "1" {
			t.Fatalf("isolated qualification has incomplete baseline: %v", baselineErr)
		}
		t.Logf("pre-existing same-user uncertainty must keep every build: %v", baselineErr)
	}
	image, err := os.ReadFile(testExecutable(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name                      string
		alias, unlink, unreadable bool
	}{
		{name: "running-build"},
		{name: "hard-link", alias: true},
		{name: "unlinked-hard-link", alias: true, unlink: true},
		{name: "unreadable-same-uid", unreadable: true},
	} {
		t.Run(row.name, func(t *testing.T) {
			store := newSweepStore(t)
			store.ExecutablePaths = nil // Exercise native /proc evidence, not the test store default.
			input := testInput("runner")
			publication, _ := testPublication(t, store.Home(), input, image)
			hit, err := store.Publish(publication, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			dir, _, err := store.paths(hit.CacheKey)
			if err != nil {
				t.Fatal(err)
			}
			backdate(t, dir, 30*24*time.Hour)
			executable := hit.ArtifactPath
			if row.alias {
				executable = filepath.Join(t.TempDir(), "alias")
				if err := os.Link(hit.ArtifactPath, executable); err != nil {
					t.Fatal(err)
				}
			}
			stop := startLinuxImage(t, executable, row.unreadable)
			if row.unlink {
				if err := os.Remove(executable); err != nil {
					t.Fatal(err)
				}
			}
			if row.unreadable {
				// The helper really changes dumpability, not just a fake error seam.
				_, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(stop.pid), "exe"))
				if !os.IsPermission(err) {
					t.Fatalf("same-user image not actually unreadable: %v", err)
				}
			}
			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			if err != nil || len(result.Removed) != 0 || !entryExists(t, store, hit.CacheKey) {
				t.Fatalf("running/unknown image removed: %+v %v", result, err)
			}
			if baselineErr == nil && row.unreadable && !warningsContaining(result, "read executable of process "+strconv.Itoa(stop.pid)) {
				t.Fatalf("same-user refusal missing: %+v", result)
			}
			if baselineErr == nil && !row.unreadable && warningsContaining(result, "build cache sweep skipped") {
				t.Fatalf("foreign process blocked sweep: %+v", result)
			}
			stop.close()
			if row.alias && !row.unlink {
				if err := os.Remove(executable); err != nil {
					t.Fatal(err)
				}
			}
			// Nearby positive control: the identical, now-unused publication is swept.
			result, err = store.Sweep(SweepRequest{}, testHomeLock{})
			assertLinuxUnusedControl(t, store, hit.CacheKey, result, err, baselineErr)
		})
	}
	// An otherwise unused candidate must not borrow the foreign-UID exception
	// when its inode is executable by another user or has a special identity mode.
	for _, row := range []struct {
		name string
		mode os.FileMode
	}{
		{"shared-group", 0o710}, {"shared-other", 0o701},
		{"special-setuid", 0o700 | os.ModeSetuid}, {"special-setgid", 0o700 | os.ModeSetgid},
	} {
		t.Run(row.name, func(t *testing.T) {
			store := newSweepStore(t)
			store.ExecutablePaths = nil
			publication, _ := testPublication(t, store.Home(), testInput("runner"), []byte("fixture"))
			hit, err := store.Publish(publication, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			dir, _, err := store.paths(hit.CacheKey)
			if err != nil {
				t.Fatal(err)
			}
			backdate(t, dir, 30*24*time.Hour)
			if err := os.Chmod(hit.ArtifactPath, row.mode); err != nil {
				t.Fatal(err)
			}
			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			info, statErr := os.Stat(hit.ArtifactPath)
			if err != nil || statErr != nil || len(result.Removed) != 0 || info.Mode() != row.mode {
				t.Fatalf("shared/special candidate changed: %+v %v %v", result, err, statErr)
			}
			if err := os.Chmod(hit.ArtifactPath, 0o700); err != nil {
				t.Fatal(err)
			}
			result, err = store.Sweep(SweepRequest{}, testHomeLock{})
			assertLinuxUnusedControl(t, store, hit.CacheKey, result, err, baselineErr)
		})
	}

}

// A known ambient unreadable same-user process must also block the unused
// control. The isolated hosted UID requires a complete baseline and therefore
// proves actual collection; ordinary suites prove safe retention when unknown.
func assertLinuxUnusedControl(t *testing.T, store *Store, key buildmeta.CacheKey, result SweepResult, err, baselineErr error) {
	t.Helper()
	if baselineErr != nil {
		if err != nil || !entryExists(t, store, key) || len(result.Removed) != 0 || !warningsContaining(result, "build cache sweep skipped") {
			t.Fatalf("ambient uncertainty did not retain unused entry: %+v %v", result, err)
		}
		return
	}
	if err != nil || entryExists(t, store, key) || len(result.Removed) != 1 || len(result.Warnings) != 0 {
		t.Fatalf("unused positive control not swept: %+v %v", result, err)
	}
}

func testExecutable(t *testing.T) string {
	t.Helper()
	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type linuxImageProcess struct {
	pid   int
	close func()
}

func startLinuxImage(t *testing.T, path string, unreadable bool) linuxImageProcess {
	t.Helper()
	command := exec.Command(path, "-test.run=^TestLinuxSweepImageHelper$") // #nosec G204 -- isolated copy of this test binary, no installed command.
	command.Env = append(os.Environ(), "CURATOR_SWEEP_IMAGE_HELPER=1", "CURATOR_SWEEP_IMAGE_UNREADABLE="+strconv.FormatBool(unreadable))
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	closed := false
	closeChild := func() {
		if !closed {
			closed = true
			_ = stdin.Close()
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	}
	t.Cleanup(closeChild)
	ready := make(chan string, 1)
	go func() { line, _ := bufio.NewReader(stdout).ReadString('\n'); ready <- line }()
	select {
	case line := <-ready:
		if line != "ready\n" {
			t.Fatalf("helper did not become ready: %q", line)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("helper readiness deadline")
	}
	return linuxImageProcess{pid: command.Process.Pid, close: closeChild}
}

// Claim: helper re-execution has an explicit ready handshake after making its
// own image unreadable; the parent owns and stops this process by PID.
func TestLinuxSweepImageHelper(t *testing.T) {
	if os.Getenv("CURATOR_SWEEP_IMAGE_HELPER") != "1" {
		return
	}
	if os.Getenv("CURATOR_SWEEP_IMAGE_UNREADABLE") == "true" {
		if err := unix.Prctl(unix.PR_SET_DUMPABLE, 0, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println("ready")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

// Claim: all four UID fields count, so setuid/saved/fs identities cannot be
// mistaken for an unrelated process just because /proc directory ownership is
// root. Unknown/malformed credentials refuse; only a fully foreign row omits.
func TestLinuxProcessUIDControls(t *testing.T) {
	for _, row := range []struct {
		name, status  string
		foreign, fail bool
	}{
		{"same", "Uid:\t123 123 123 123", false, false},
		{"foreign", "Uid:\t0 0 0 0", true, false},
		{"real", "Uid:\t123 0 0 0", false, false},
		{"effective", "Uid:\t0 123 0 0", false, false},
		{"saved", "Uid:\t0 0 123 0", false, false},
		{"filesystem", "Uid:\t0 0 0 123", false, false},
		{"missing", "Name:\thelper", false, true},
		{"truncated", "Uid:\t0 0", false, true},
		{"invalid", "Uid:\t0 0 bad 0", false, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			foreign, err := linuxForeignUID(row.status, 123)
			if foreign != row.foreign || (err != nil) != row.fail {
				t.Fatalf("foreign=%v err=%v", foreign, err)
			}
		})
	}
	// A regular exe fixture gives a real readlink EINVAL, with no fake error API.
	for _, row := range []struct {
		name, uids string
		fail       bool
	}{
		{"foreign", "0 0 0 0", false}, {"same", "123 123 123 123", true}, {"saved-owner", "0 0 123 0", true},
	} {
		t.Run("unreadable-"+row.name, func(t *testing.T) {
			proc := t.TempDir()
			dir := filepath.Join(proc, "17")
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			for name, body := range map[string]string{"exe": "not a process link", "status": "Uid:\t" + row.uids + "\n"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			snapshot, err := linuxProcessSnapshot(proc, 123, false)
			if (err != nil) != row.fail || !row.fail && (!snapshot.foreignExcluded || len(snapshot.paths) != 0) {
				t.Fatalf("snapshot=%+v err=%v", snapshot, err)
			}
			if row.fail && !strings.Contains(err.Error(), "read executable of process 17") {
				t.Fatalf("wrong refusal: %v", err)
			}
		})
	}
}

// Claim: omitted foreign images never authorize removal of shared-executable,
// setuid, or setgid candidates. Owner-only ordinary files are the positive
// control, independent of whether a CI host restricts /proc/1/exe.
func TestLinuxForeignImagePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact")
	if err := os.WriteFile(path, []byte("fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name string
		mode os.FileMode
		keep bool
	}{
		{"private", 0o700, false}, {"group", 0o710, true}, {"other", 0o701, true},
		{"setuid", 0o700 | os.ModeSetuid, true}, {"setgid", 0o700 | os.ModeSetgid, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			if err := os.Chmod(path, row.mode); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := (executableSnapshot{foreignExcluded: true}).retains(info); got != row.keep {
				t.Fatalf("retain=%v mode=%v", got, info.Mode())
			}
		})
	}
}
