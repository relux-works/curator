package transaction

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/managerlock"
)

const (
	helperModeEnv    = "CURATOR_TRANSACTION_HELPER_MODE"
	helperHomeEnv    = "CURATOR_TRANSACTION_HELPER_HOME"
	helperPointEnv   = "CURATOR_TRANSACTION_HELPER_POINT"
	helperIDEnv      = "CURATOR_TRANSACTION_HELPER_ID"
	helperLiveEnv    = "CURATOR_TRANSACTION_HELPER_LIVE"
	helperSourceEnv  = "CURATOR_TRANSACTION_HELPER_SOURCE"
	helperOrdinalEnv = "CURATOR_TRANSACTION_HELPER_ORDINAL"
)

func TestSubprocessCrashRecoveryAtSwapBoundaries(t *testing.T) {
	for _, point := range []Point{PointAfterBackup, PointAfterInstall} {
		t.Run(string(point), func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			liveRoot := filepath.Join(root, "live")
			stageRoot := filepath.Join(root, "stage")
			mustMkdirAll(t, liveRoot)
			mustMkdirAll(t, stageRoot)
			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			targets := []Target{
				fileTarget(t, "class", "a", filepath.Join(liveRoot, "a"), filepath.Join(stageRoot, "a"), "old-a", "new-a"),
				fileTarget(t, "class", "b", filepath.Join(liveRoot, "b"), filepath.Join(stageRoot, "b"), "old-b", "new-b"),
			}
			engine := mustEngine(t, home)
			journal, err := engine.Prepare(lock, Plan{TransactionID: "txn-subprocess", ProjectIdentity: "/project-that-will-not-be-current", Targets: targets})
			if err != nil {
				t.Fatal(err)
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}

			command := exec.Command(os.Args[0], "-test.run=^TestTransactionCrashHelper$")
			command.Env = append(os.Environ(),
				helperModeEnv+"=1",
				helperHomeEnv+"="+home,
				helperPointEnv+"="+string(point),
				helperIDEnv+"="+journal.TransactionID,
			)
			err = command.Run()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 86 {
				t.Fatalf("helper exit = %v, want 86", err)
			}

			restartedManager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			recoveryLock, err := restartedManager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := recoveryLock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			if err := mustEngine(t, home).Recover(recoveryLock); err != nil {
				t.Fatal(err)
			}
			for index, target := range targets {
				if got, want := mustRead(t, target.LivePath), "new-"+string(rune('a'+index)); got != want {
					t.Fatalf("target %d after restart = %q, want %q", index, got, want)
				}
			}
			if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
				t.Fatalf("recovered journal remains: %v", err)
			}
		})
	}
}

func TestSubprocessCrashRecoveryDuringPreparation(t *testing.T) {
	for _, kind := range []string{"file", "directory"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, kind)
			preimage, err := DigestPath(live)
			if err != nil {
				t.Fatal(err)
			}
			journal := crashDuringPreparation(t, home, live, source, "txn-prepare-crash")
			if journal.Phase != PhasePreparing || !journal.Targets[0].StagingActive || !journal.Targets[0].StagingCreated {
				t.Fatalf("crash journal does not record active staging: %+v", journal.Targets[0])
			}
			stagedDigest, err := DigestPath(journal.Targets[0].StagedPath)
			if err != nil {
				t.Fatal(err)
			}
			if stagedDigest == DigestAbsent || stagedDigest == journal.Targets[0].DesiredDigest {
				t.Fatalf("crash staging digest = %q, want a present partial target", stagedDigest)
			}

			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			if err := mustEngine(t, home).Recover(lock); err != nil {
				t.Fatal(err)
			}
			if got, err := DigestPath(live); err != nil || got != preimage {
				t.Fatalf("preparing recovery changed live target: digest %q, error %v; want %q", got, err, preimage)
			}
			if _, err := os.Lstat(journal.Targets[0].StagedPath); !os.IsNotExist(err) {
				t.Fatalf("preparing recovery left staging: %v", err)
			}
			if _, err := os.Lstat(mustEngine(t, home).journalPath(journal.TransactionID)); !os.IsNotExist(err) {
				t.Fatalf("preparing recovery left journal: %v", err)
			}
		})
	}
}

func TestSubprocessCrashRecoveryAfterStagingChunkSync(t *testing.T) {
	for _, ordinal := range []int{1, 2} {
		t.Run(strconv.Itoa(ordinal), func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, "file")
			preimage, err := DigestPath(live)
			if err != nil {
				t.Fatal(err)
			}
			journal := crashDuringPreparationAt(t, home, live, source, "txn-prepare-write-ahead", PointAfterStagingChunkSync, ordinal)
			target := journal.Targets[0]
			wantAcknowledged := int64(ordinal-1) * stagingCopyChunkSize
			wantOwned := int64(ordinal) * stagingCopyChunkSize
			if target.StagingBytes != wantAcknowledged || target.StagingWriteBytes != wantOwned || target.StagingWriteDigest == "" {
				t.Fatalf("write-ahead staging progress = acknowledged %d, owned %d/%q; want %d and %d with digest", target.StagingBytes, target.StagingWriteBytes, target.StagingWriteDigest, wantAcknowledged, wantOwned)
			}
			info, err := os.Stat(target.StagedPath)
			if err != nil {
				t.Fatal(err)
			}
			if info.Size() != wantOwned {
				t.Fatalf("durable staged size = %d, want %d", info.Size(), wantOwned)
			}

			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			if err := mustEngine(t, home).Recover(lock); err != nil {
				t.Fatal(err)
			}
			if got, err := DigestPath(live); err != nil || got != preimage {
				t.Fatalf("write-ahead recovery changed live target: digest %q, error %v; want %q", got, err, preimage)
			}
			if _, err := os.Lstat(target.StagedPath); !os.IsNotExist(err) {
				t.Fatalf("write-ahead recovery left staging: %v", err)
			}
			if _, err := os.Lstat(mustEngine(t, home).journalPath(journal.TransactionID)); !os.IsNotExist(err) {
				t.Fatalf("write-ahead recovery left journal: %v", err)
			}
		})
	}
}

func TestPreparingRecoveryPreservesConcurrentStagingState(t *testing.T) {
	tests := []struct {
		name   string
		kind   string
		mutate func(*testing.T, string)
		check  func(*testing.T, string)
	}{
		{
			name: "replaced partial file",
			kind: "file",
			mutate: func(t *testing.T, path string) {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				mustWrite(t, path, "foreign replacement")
			},
			check: func(t *testing.T, path string) {
				if got := mustRead(t, path); got != "foreign replacement" {
					t.Fatalf("foreign replacement changed: %q", got)
				}
			},
		},
		{
			name: "added directory entry",
			kind: "directory",
			mutate: func(t *testing.T, path string) {
				mustWrite(t, filepath.Join(path, "foreign"), "concurrent bytes")
			},
			check: func(t *testing.T, path string) {
				if got := mustRead(t, filepath.Join(path, "foreign")); got != "concurrent bytes" {
					t.Fatalf("concurrent bytes changed: %q", got)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, testCase.kind)
			journal := crashDuringPreparation(t, home, live, source, "txn-prepare-concurrent")
			staged := journal.Targets[0].StagedPath
			testCase.mutate(t, staged)

			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			err = mustEngine(t, home).Recover(lock)
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("recover error = %v, want implementation-corruption", err)
			}
			testCase.check(t, staged)
			if _, err := os.Lstat(mustEngine(t, home).journalPath(journal.TransactionID)); err != nil {
				t.Fatalf("corrupt preparing recovery removed journal: %v", err)
			}
		})
	}
}

func TestPreparingRecoveryPreservesChangedDurableStagingPrefix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string, string, int64) []byte
	}{
		{name: "shorter prefix", mutate: func(t *testing.T, destination, source string, recorded int64) []byte {
			return rewriteDurableSourcePrefix(t, destination, source, recorded/2)
		}},
		{name: "longer prefix", mutate: func(t *testing.T, destination, source string, recorded int64) []byte {
			return rewriteDurableSourcePrefix(t, destination, source, recorded+recorded/2)
		}},
		{name: "same-size changed bytes", mutate: func(t *testing.T, destination, source string, recorded int64) []byte {
			prefix := rewriteDurableSourcePrefix(t, destination, source, recorded)
			prefix[len(prefix)/2] ^= 0xff
			rewriteDurableBytes(t, destination, prefix)
			return prefix
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, "file")
			journal := crashDuringPreparation(t, home, live, source, "txn-prepare-prefix-change")
			target := journal.Targets[0]
			if target.StagingBytes != stagingCopyChunkSize || target.StagingPrefixDigest == "" {
				t.Fatalf("durable staging progress = %d/%q, want one acknowledged chunk", target.StagingBytes, target.StagingPrefixDigest)
			}
			changedPrefix := testCase.mutate(t, target.StagedPath, source, target.StagingBytes)

			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			err = mustEngine(t, home).Recover(lock)
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("recover error = %v, want implementation-corruption", err)
			}
			if got, err := os.ReadFile(target.StagedPath); err != nil || string(got) != string(changedPrefix) {
				t.Fatalf("changed source prefix was not preserved: bytes=%d error=%v", len(got), err)
			}
			if _, err := os.Lstat(mustEngine(t, home).journalPath(journal.TransactionID)); err != nil {
				t.Fatalf("corrupt preparing recovery removed journal: %v", err)
			}
			if got := mustRead(t, live); got != "old" {
				t.Fatalf("corrupt preparing recovery changed live target: %q", got)
			}
		})
	}
}

func TestPreparingRecoveryPreservesStateOutsideStagingWriteAhead(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string, string, TargetRecord) []byte
	}{
		{name: "shorter than acknowledged", mutate: func(t *testing.T, destination, source string, target TargetRecord) []byte {
			return rewriteDurableSourcePrefix(t, destination, source, target.StagingBytes/2)
		}},
		{name: "longer than authorized", mutate: func(t *testing.T, destination, source string, target TargetRecord) []byte {
			return rewriteDurableSourcePrefix(t, destination, source, target.StagingWriteBytes+stagingCopyChunkSize/2)
		}},
		{name: "changed authorized size", mutate: func(t *testing.T, destination, source string, target TargetRecord) []byte {
			prefix := rewriteDurableSourcePrefix(t, destination, source, target.StagingWriteBytes)
			prefix[len(prefix)/2] ^= 0xff
			rewriteDurableBytes(t, destination, prefix)
			return prefix
		}},
		{name: "replacement", mutate: func(t *testing.T, destination, _ string, _ TargetRecord) []byte {
			if err := os.Remove(destination); err != nil {
				t.Fatal(err)
			}
			payload := []byte("foreign replacement")
			rewriteDurableBytes(t, destination, payload)
			return payload
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			live, source := preparationCrashFixture(t, root, "file")
			journal := crashDuringPreparationAt(t, home, live, source, "txn-prepare-write-ahead-corruption", PointAfterStagingChunkSync, 2)
			target := journal.Targets[0]
			if target.StagingBytes != stagingCopyChunkSize || target.StagingWriteBytes != 2*stagingCopyChunkSize {
				t.Fatalf("unexpected second-chunk write-ahead progress: %+v", target)
			}
			changed := testCase.mutate(t, target.StagedPath, source, target)

			manager, err := managerlock.New(home)
			if err != nil {
				t.Fatal(err)
			}
			lock, err := manager.AcquireHomeOnly(context.Background(), false)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := lock.Close(); err != nil {
					t.Errorf("release recovery lock: %v", err)
				}
			})
			err = mustEngine(t, home).Recover(lock)
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("recover error = %v, want implementation-corruption", err)
			}
			if got, err := os.ReadFile(target.StagedPath); err != nil || string(got) != string(changed) {
				t.Fatalf("state outside write-ahead was not preserved: bytes=%d error=%v", len(got), err)
			}
			if _, err := os.Lstat(mustEngine(t, home).journalPath(journal.TransactionID)); err != nil {
				t.Fatalf("corrupt write-ahead recovery removed journal: %v", err)
			}
			if got := mustRead(t, live); got != "old" {
				t.Fatalf("corrupt write-ahead recovery changed live target: %q", got)
			}
		})
	}
}

func rewriteDurableSourcePrefix(t *testing.T, destination, source string, size int64) []byte {
	t.Helper()
	payload, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if size < 0 || size > int64(len(payload)) {
		t.Fatalf("requested source prefix size %d outside payload size %d", size, len(payload))
	}
	prefix := payload[:size]
	rewriteDurableBytes(t, destination, prefix)
	return prefix
}

func rewriteDurableBytes(t *testing.T, destination string, payload []byte) {
	t.Helper()
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o600) // #nosec G304 -- test-owned transaction sidecar
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := syncDirectory(filepath.Dir(destination)); err != nil {
		t.Fatal(err)
	}
}

func preparationCrashFixture(t *testing.T, root, kind string) (string, string) {
	t.Helper()
	payload := strings.Repeat("desired-staging-payload-", stagingCopyChunkSize*3/24)
	if kind == "file" {
		live := filepath.Join(root, "live", "target")
		source := filepath.Join(root, "source", "target")
		mustWrite(t, live, "old")
		mustWrite(t, source, payload)
		return live, source
	}
	live := filepath.Join(root, "live", "target")
	source := filepath.Join(root, "source", "target")
	mustWrite(t, filepath.Join(live, "nested", "payload"), "old")
	mustWrite(t, filepath.Join(source, "nested", "payload"), payload)
	return live, source
}

func crashDuringPreparation(t *testing.T, home, live, source, transactionID string) *Journal {
	t.Helper()
	return crashDuringPreparationAt(t, home, live, source, transactionID, PointDuringStagingCopy, 1)
}

func crashDuringPreparationAt(t *testing.T, home, live, source, transactionID string, point Point, ordinal int) *Journal {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestTransactionCrashHelper$")
	command.Env = append(os.Environ(),
		helperModeEnv+"=prepare",
		helperHomeEnv+"="+home,
		helperLiveEnv+"="+live,
		helperSourceEnv+"="+source,
		helperIDEnv+"="+transactionID,
		helperPointEnv+"="+string(point),
		helperOrdinalEnv+"="+strconv.Itoa(ordinal),
	)
	err := command.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 86 {
		t.Fatalf("prepare helper exit = %v, want 86", err)
	}
	journal, err := mustEngine(t, home).loadJournal(transactionID)
	if err != nil {
		t.Fatal(err)
	}
	return journal
}

func TestTransactionCrashHelper(t *testing.T) {
	mode := os.Getenv(helperModeEnv)
	if mode == "" {
		return
	}
	home := os.Getenv(helperHomeEnv)
	manager, err := managerlock.New(home)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Errorf("release helper lock: %v", err)
		}
	}()
	point := Point(os.Getenv(helperPointEnv))
	if mode == "prepare" && point == "" {
		point = PointDuringStagingCopy
	}
	ordinal, err := strconv.Atoi(os.Getenv(helperOrdinalEnv))
	if err != nil || ordinal < 1 {
		ordinal = 1
	}
	observed := 0
	engine, err := New(home, WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == point && event.TargetIndex == 0 {
			observed++
			if observed == ordinal {
				os.Exit(86)
			}
		}
		return nil
	}}))
	if err != nil {
		t.Fatal(err)
	}
	if mode == "prepare" {
		live := os.Getenv(helperLiveEnv)
		preimage, err := DigestPath(live)
		if err != nil {
			t.Fatal(err)
		}
		_, err = engine.Prepare(lock, Plan{TransactionID: os.Getenv(helperIDEnv), ProjectIdentity: "/helper-project", Targets: []Target{{
			Class: "class", Identifier: "target", LivePath: live, StagedSource: os.Getenv(helperSourceEnv), PreimageDigest: preimage,
		}}})
		if err != nil {
			t.Fatal(err)
		}
	} else if err := engine.Commit(lock, os.Getenv(helperIDEnv)); err != nil {
		t.Fatal(err)
	}
	t.Fatal("helper transaction returned without crashing")
}
