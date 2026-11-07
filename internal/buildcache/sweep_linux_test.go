package buildcache

import (
	"bufio"
	"bytes"
	"errors"
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
// this user's running build (including an unlinked alias), and collects it once
// stopped. Non-dumpable same-user processes do not block unused builds.
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
	image, err := os.ReadFile(testExecutable(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name                                   string
		alias, unlink, unreadable, modeChanged bool
	}{
		{name: "running-build"},
		{name: "hard-link", alias: true},
		{name: "unlinked-hard-link", alias: true, unlink: true},
		{name: "unreadable-same-uid", unreadable: true},
		{name: "mode-changed-after-exec", unreadable: true, modeChanged: true},
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
			// Claim for mode-changed-after-exec: sweep-time 0700 does not
			// erase execution begun at 0755; the unreadable image still probes busy.
			if row.modeChanged {
				if err := os.Chmod(executable, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			stop := startLinuxImage(t, executable, row.unreadable)
			if row.modeChanged {
				if err := os.Chmod(executable, 0o700); err != nil {
					t.Fatal(err)
				}
			}
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
			// Kernel evidence is independent of proc visibility and does not
			// truncate the running inode. Check the exact ETXTBSY refusal.
			if probeErr := testLinuxProbe(t, hit.ArtifactPath); !errors.Is(probeErr, unix.ETXTBSY) {
				t.Fatalf("running image write probe did not report ETXTBSY: %v", probeErr)
			}
			t.Log("write-open execution probe: ETXTBSY")
			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			if err != nil || len(result.Removed) != 0 || !entryExists(t, store, hit.CacheKey) {
				t.Fatalf("running image removed: %+v %v", result, err)
			}
			if row.alias && !row.unlink {
				// A still-linked alias also violates the existing single-link
				// boundary. Its precise warning is expected; the direct probe
				// above independently proves ETXTBSY for the running inode.
				if len(result.Warnings) != 1 || !warningsContaining(result, "has multiple hard links") {
					t.Fatalf("linked-image boundary refusal missing: %+v", result)
				}
			} else if len(result.Warnings) != 0 {
				t.Fatalf("running-image retention mislabeled: %+v", result)
			}
			if row.unreadable {
				// An ssh-agent-like non-dumpable image is still running while
				// an unrelated unused build is collected through native proc.
				unused := publishTestEntry(t, store, "unused", "unused artifact")
				backdate(t, entryPathOf(store, unused), 30*24*time.Hour)
				result, err = store.Sweep(SweepRequest{}, testHomeLock{})
				if err != nil || entryExists(t, store, unused) || !entryExists(t, store, hit.CacheKey) || len(result.Removed) != 1 || len(result.Warnings) != 0 {
					t.Fatalf("ssh-agent-like process blocked unused build: %+v %v", result, err)
				}
			}
			stop.close()
			content, readErr := os.ReadFile(hit.ArtifactPath)
			if readErr != nil || !bytes.Equal(content, image) {
				t.Fatalf("busy probe changed executable bytes: %v", readErr)
			}
			if probeErr := testLinuxProbe(t, hit.ArtifactPath); probeErr != nil {
				t.Fatalf("stopped positive probe refused: %v", probeErr)
			}
			if row.alias && !row.unlink {
				if err := os.Remove(executable); err != nil {
					t.Fatal(err)
				}
			}
			// Nearby positive control: the identical, now-unused publication is swept.
			result, err = store.Sweep(SweepRequest{}, testHomeLock{})
			assertLinuxUnusedControl(t, store, hit.CacheKey, result, err)
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
			assertLinuxUnusedControl(t, store, hit.CacheKey, result, err)
		})
	}

}

// Nearby positive controls require collection, even with ambient unreadable
// same-user processes: the per-executable probe is now mandatory on Linux.
func assertLinuxUnusedControl(t *testing.T, store *Store, key buildmeta.CacheKey, result SweepResult, err error) {
	t.Helper()
	if err != nil || entryExists(t, store, key) || len(result.Removed) != 1 || len(result.Warnings) != 0 {
		t.Fatalf("unused positive control not swept: %+v %v", result, err)
	}
}

func testLinuxProbe(t *testing.T, path string) error {
	t.Helper()
	fd, err := unix.Open(filepath.Dir(path), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = unix.Close(fd) }()
	var stat unix.Stat_t
	if err := unix.Fstatat(fd, filepath.Base(path), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatal(err)
	}
	return probeLinuxExecutable(fd, filepath.Base(path), &stat)
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
				t.Fatalf("same-user UID misclassified: foreign=%v err=%v", foreign, err)
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
			if (err != nil) != row.fail || !row.fail && (!snapshot.privateOnly || len(snapshot.paths) != 0) {
				t.Fatalf("snapshot=%+v err=%v", snapshot, err)
			}
			if row.fail && !strings.Contains(err.Error(), "read executable of process 17") {
				t.Fatalf("wrong refusal: %v", err)
			}
		})
	}
}

// Claim: the non-root mode guard applies even if hidepid exposes no foreign
// image; it never authorizes removal of shared-executable,
// setuid, or setgid candidates. Owner-only ordinary files are the positive
// control, independent of whether a CI host restricts /proc/1/exe.
func TestLinuxForeignImagePermissions(t *testing.T) {
	// An empty visible table models hidepid=2; this is not a mount experiment.
	snapshot, err := linuxProcessSnapshot(t.TempDir(), 123, true)
	if err != nil {
		t.Fatal(err)
	}
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
			if got := snapshot.retains(info); got != row.keep {
				t.Fatalf("retain=%v mode=%v", got, info.Mode())
			}
		})
	}
}

// Claim: an actual non-truncating probe error retains the named build and its
// exact bytes; restoring write permission lets the same unused build collect.
func TestLinuxExecutionProbeErrors(t *testing.T) {
	store := newSweepStore(t)
	store.ExecutablePaths = nil
	key := publishTestEntry(t, store, "runner", "unchanged executable")
	dir := entryPathOf(store, key)
	artifact := filepath.Join(dir, "bin", "runner")
	backdate(t, dir, 30*24*time.Hour)
	if err := os.Chmod(artifact, 0o500); err != nil {
		t.Fatal(err)
	}
	probeErr := testLinuxProbe(t, artifact)
	if !errors.Is(probeErr, unix.EACCES) {
		t.Fatalf("real probe must fail EACCES under the unprivileged test UID: %v", probeErr)
	}
	result, err := store.Sweep(SweepRequest{}, testHomeLock{})
	if err != nil || len(result.Removed) != 0 || !entryExists(t, store, key) || !warningsContaining(result, `write-open execution probe "bin/runner": permission denied`) {
		t.Fatalf("probe error not retained and named: %+v %v", result, err)
	}
	bytes, err := os.ReadFile(artifact)
	if err != nil || string(bytes) != "unchanged executable" {
		t.Fatalf("probe modified bytes: %q %v", bytes, err)
	}
	if err := os.Chmod(artifact, 0o700); err != nil {
		t.Fatal(err)
	}
	result, err = store.Sweep(SweepRequest{}, testHomeLock{})
	assertLinuxUnusedControl(t, store, key, result, err)
}

// Claim: native Linux cleanup needs remaining executable evidence, not intact
// receipts. Partial retired trees finish; a running unreadable alias is retained
// until stopped. A probe error retains the named retired member until repaired.
func TestLinuxRetiredNativeSnapshot(t *testing.T) {
	image, err := os.ReadFile(testExecutable(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{"empty-bin", "missing-bin", "missing-receipt", "missing-artifact", "running-alias", "probe-error", "external-link"} {
		t.Run(row, func(t *testing.T) {
			store := newSweepStore(t)
			store.ExecutablePaths = nil
			publication, _ := testPublication(t, store.Home(), testInput("runner"), image)
			hit, err := store.Publish(publication, testHomeLock{})
			if err != nil {
				t.Fatal(err)
			}
			dir := entryPathOf(store, hit.CacheKey)
			retired := filepath.Join(filepath.Dir(dir), sweepPrefix+filepath.Base(dir)+"-partial")
			if err := os.Rename(dir, retired); err != nil {
				t.Fatal(err)
			}
			artifact := filepath.Join(retired, "bin", "runner")
			if err := os.Remove(filepath.Join(retired, ReceiptFilename)); err != nil {
				t.Fatal(err)
			}
			var stop linuxImageProcess
			switch row {
			case "empty-bin", "missing-artifact":
				if err := os.Remove(artifact); err != nil {
					t.Fatal(err)
				}
			case "missing-bin":
				if err := os.RemoveAll(filepath.Join(retired, "bin")); err != nil {
					t.Fatal(err)
				}
			case "running-alias":
				alias := filepath.Join(t.TempDir(), "alias")
				if err := os.Link(artifact, alias); err != nil {
					t.Fatal(err)
				}
				stop = startLinuxImage(t, alias, true)
				if err := os.Remove(alias); err != nil {
					t.Fatal(err)
				}
			case "probe-error":
				if err := os.Chmod(artifact, 0o500); err != nil {
					t.Fatal(err)
				}
			case "external-link":
				outside := filepath.Join(t.TempDir(), "outside")
				if err := os.WriteFile(outside, []byte("outside unchanged"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(retired, "external")); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					bytes, err := os.ReadFile(outside)
					if err != nil || string(bytes) != "outside unchanged" {
						t.Errorf("cleanup followed external link: %q %v", bytes, err)
					}
				})
			}
			result, err := store.Sweep(SweepRequest{}, testHomeLock{})
			_, statErr := os.Lstat(retired)
			keep := row == "running-alias" || row == "probe-error"
			if err != nil || keep && statErr != nil || !keep && !os.IsNotExist(statErr) {
				t.Fatalf("retired disposition wrong: %+v %v %v", result, err, statErr)
			}
			if row == "probe-error" {
				if !warningsContaining(result, filepath.Base(retired)) || !warningsContaining(result, `write-open execution probe "bin/runner": permission denied`) {
					t.Fatalf("retired probe error unnamed: %+v", result)
				}
				bytes, readErr := os.ReadFile(artifact)
				if readErr != nil || len(bytes) != len(image) || string(bytes) != string(image) {
					t.Fatal("probe error modified artifact")
				}
				if err := os.Chmod(artifact, 0o700); err != nil {
					t.Fatal(err)
				}
			} else if len(result.Warnings) != 0 {
				t.Fatalf("retired control warned: %+v", result)
			}
			if keep {
				if stop.close != nil {
					stop.close()
				}
				result, err = store.Sweep(SweepRequest{}, testHomeLock{})
				_, statErr = os.Lstat(retired)
				if err != nil || !os.IsNotExist(statErr) || len(result.Warnings) != 0 {
					t.Fatalf("retired positive control not finished: %+v %v %v", result, err, statErr)
				}
			}
		})
	}
}

// Contract row, not an HSM experiment: executables opened under fanotify
// pre-content/HSM watches are the accepted exception to the any-process
// guarantee. Kernel 0357ef03c94e exempts them from write denial. Path removal
// does not stop execution, but later exec fails until rebuilt. This declaration
// row records the accepted limit only; it supplies no HSM execution evidence.
func TestLinuxHSMExceptionContract(t *testing.T) {
	const exception = "fanotify pre-content/HSM watch: write denial exempt; later exec can fail until rebuilt"
	t.Log(exception)
}
