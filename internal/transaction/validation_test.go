package transaction

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildJournalRejectsInvalidPlansAndTargets(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	valid := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	tests := []struct {
		name   string
		mutate func(*Plan)
	}{
		{name: "transaction id", mutate: func(plan *Plan) { plan.TransactionID = "../escape" }},
		{name: "project", mutate: func(plan *Plan) { plan.ProjectIdentity = "" }},
		{name: "no targets", mutate: func(plan *Plan) { plan.Targets = nil }},
		{name: "empty build key", mutate: func(plan *Plan) { plan.ReferencedBuildKeys = []string{""} }},
		{name: "duplicate build key", mutate: func(plan *Plan) { plan.ReferencedBuildKeys = []string{"same", "same"} }},
		{name: "class", mutate: func(plan *Plan) { plan.Targets[0].Class = "" }},
		{name: "identifier", mutate: func(plan *Plan) { plan.Targets[0].Identifier = "" }},
		{name: "duplicate key", mutate: func(plan *Plan) { plan.Targets = append(plan.Targets, plan.Targets[0]) }},
		{name: "no expectation", mutate: func(plan *Plan) { plan.Targets[0].PreimageDigest = "" }},
		{name: "two expectations", mutate: func(plan *Plan) { plan.Targets[0].ExpectedGeneration = "generation" }},
		{name: "bad digest", mutate: func(plan *Plan) { plan.Targets[0].PreimageDigest = "sha256:no" }},
		{name: "empty live", mutate: func(plan *Plan) { plan.Targets[0].LivePath = "" }},
		{name: "generation escape", mutate: func(plan *Plan) { plan.Targets[0].GenerationPath = "../generation" }},
		{name: "source is live", mutate: func(plan *Plan) { plan.Targets[0].StagedSource = plan.Targets[0].LivePath }},
		{name: "source absent", mutate: func(plan *Plan) { plan.Targets[0].StagedSource = filepath.Join(root, "missing") }},
		{name: "duplicate live", mutate: func(plan *Plan) {
			second := plan.Targets[0]
			second.Identifier = "other"
			plan.Targets = append(plan.Targets, second)
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			plan := Plan{TransactionID: "txn-validation", ProjectIdentity: "/project", Targets: []Target{valid}, ReferencedBuildKeys: []string{"key"}}
			testCase.mutate(&plan)
			if _, _, err := engine.buildJournal(plan); err == nil {
				t.Fatal("invalid plan was accepted")
			}
		})
	}

	sidecar := sidecarPath(valid.LivePath, "txn-sidecar", 0, "backup")
	mustWrite(t, sidecar, "orphan")
	plan := Plan{TransactionID: "txn-sidecar", ProjectIdentity: "/project", Targets: []Target{valid}}
	if _, _, err := engine.buildJournal(plan); !errors.Is(err, ErrImplementationCorruption) {
		t.Fatalf("orphan sidecar error = %v", err)
	}
}

func TestValidateJournalRejectsEveryNoncanonicalFieldCluster(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-journal-validation", ProjectIdentity: "/project", Targets: []Target{target}, ReferencedBuildKeys: []string{"key"}})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*Journal)
	}{
		{name: "schema", mutate: func(value *Journal) { value.Schema = "other" }},
		{name: "project", mutate: func(value *Journal) { value.ProjectIdentity = "" }},
		{name: "phase", mutate: func(value *Journal) { value.Phase = "unknown" }},
		{name: "removal outside cleanup", mutate: func(value *Journal) { value.RemovalPath = value.Targets[0].BackupPath }},
		{name: "removal without entries", mutate: func(value *Journal) {
			value.Phase = PhaseCleanup
			value.RemovalPath = value.Targets[0].BackupPath
		}},
		{name: "entries without removal", mutate: func(value *Journal) {
			value.RemovalEntries = []RemovalEntry{{RelativePath: "", Kind: "file", Mode: 0o600, Digest: value.Targets[0].DesiredDigest}}
		}},
		{name: "invalid removal entries", mutate: func(value *Journal) {
			value.Phase = PhaseCleanup
			value.RemovalPath = value.Targets[0].BackupPath
			value.RemovalEntries = []RemovalEntry{{RelativePath: "", Kind: "file", Mode: 0o600, Digest: "bad"}}
		}},
		{name: "noncanonical removal", mutate: func(value *Journal) {
			value.Phase = PhaseCleanup
			value.RemovalPath = filepath.Join(root, "other-removal")
		}},
		{name: "keys", mutate: func(value *Journal) { value.ReferencedBuildKeys = []string{"z", "a"} }},
		{name: "targets", mutate: func(value *Journal) { value.Targets = nil }},
		{name: "target text", mutate: func(value *Journal) { value.Targets[0].Class = "" }},
		{name: "target order", mutate: func(value *Journal) { value.Targets = append(value.Targets, value.Targets[0]) }},
		{name: "relative live", mutate: func(value *Journal) { value.Targets[0].LivePath = "relative" }},
		{name: "sidecar parent", mutate: func(value *Journal) { value.Targets[0].BackupPath = filepath.Join(root, "other", "backup") }},
		{name: "noncanonical sidecar", mutate: func(value *Journal) {
			value.Targets[0].BackupPath = filepath.Join(filepath.Dir(value.Targets[0].LivePath), "other-backup")
		}},
		{name: "expectation", mutate: func(value *Journal) { value.Targets[0].ExpectedGeneration = "also" }},
		{name: "preimage digest", mutate: func(value *Journal) { value.Targets[0].PreimageDigest = "bad" }},
		{name: "generation text", mutate: func(value *Journal) {
			value.Targets[0].PreimageDigest = ""
			value.Targets[0].ExpectedGeneration = "bad\x00generation"
		}},
		{name: "generation escape", mutate: func(value *Journal) { value.Targets[0].GenerationPath = "../escape" }},
		{name: "desired digest", mutate: func(value *Journal) { value.Targets[0].DesiredDigest = "bad" }},
		{name: "desired absence with staging", mutate: func(value *Journal) { value.Targets[0].DesiredDigest = DigestAbsent }},
		{name: "backup digest", mutate: func(value *Journal) { value.Targets[0].BackupDigest = "bad" }},
		{name: "state", mutate: func(value *Journal) { value.Targets[0].State = "unknown" }},
		{name: "target namespace overlap", mutate: func(value *Journal) {
			second := value.Targets[0]
			second.Identifier = "nested"
			second.LivePath = filepath.Join(value.Targets[0].LivePath, "nested")
			second.StagedPath = sidecarPath(second.LivePath, value.TransactionID, 1, "desired")
			second.BackupPath = sidecarPath(second.LivePath, value.TransactionID, 1, "backup")
			second.RollbackPath = sidecarPath(second.LivePath, value.TransactionID, 1, "rollback")
			value.Targets = append(value.Targets, second)
		}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			value := cloneJournal(journal)
			testCase.mutate(value)
			if err := validateJournal(value); err == nil {
				t.Fatal("invalid journal was accepted")
			}
		})
	}
}

func TestValidateJournalRejectsNoncanonicalStagingPrefixProgress(t *testing.T) {
	root := t.TempDir()
	live := filepath.Join(root, "live", "target")
	source := filepath.Join(root, "stage", "target")
	target := fileTarget(t, "class", "id", live, source, "old", "new")
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, _, err := engine.buildJournal(Plan{TransactionID: "txn-staging-progress", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	journal.Targets[0].StagingActive = true
	journal.Targets[0].StagingCreated = true
	journal.Targets[0].StagingBytes = 1
	journal.Targets[0].StagingPrefixDigest = strings.Repeat("a", 64)
	journal.Targets[0].StagingPrefixDigest = "sha256:" + journal.Targets[0].StagingPrefixDigest
	journal.Targets[0].StagingWriteBytes = 2
	journal.Targets[0].StagingWriteDigest = "sha256:" + strings.Repeat("b", 64)
	if err := validateJournal(journal); err != nil {
		t.Fatalf("valid staging prefix progress rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*TargetRecord)
	}{
		{name: "bytes without digest", mutate: func(target *TargetRecord) { target.StagingPrefixDigest = "" }},
		{name: "digest without bytes", mutate: func(target *TargetRecord) { target.StagingBytes = 0 }},
		{name: "negative bytes", mutate: func(target *TargetRecord) { target.StagingBytes = -1 }},
		{name: "inactive progress", mutate: func(target *TargetRecord) { target.StagingActive = false }},
		{name: "uncreated progress", mutate: func(target *TargetRecord) { target.StagingCreated = false }},
		{name: "absent digest", mutate: func(target *TargetRecord) { target.StagingPrefixDigest = DigestAbsent }},
		{name: "invalid digest", mutate: func(target *TargetRecord) { target.StagingPrefixDigest = "sha256:no" }},
		{name: "write bytes without digest", mutate: func(target *TargetRecord) { target.StagingWriteDigest = "" }},
		{name: "write digest without bytes", mutate: func(target *TargetRecord) { target.StagingWriteBytes = 0 }},
		{name: "write not beyond acknowledged bytes", mutate: func(target *TargetRecord) { target.StagingWriteBytes = target.StagingBytes }},
		{name: "write exceeds one chunk", mutate: func(target *TargetRecord) { target.StagingWriteBytes = target.StagingBytes + stagingCopyChunkSize + 1 }},
		{name: "negative write bytes", mutate: func(target *TargetRecord) { target.StagingWriteBytes = -1 }},
		{name: "absent write digest", mutate: func(target *TargetRecord) { target.StagingWriteDigest = DigestAbsent }},
		{name: "invalid write digest", mutate: func(target *TargetRecord) { target.StagingWriteDigest = "sha256:no" }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			value := cloneJournal(journal)
			testCase.mutate(&value.Targets[0])
			if err := validateJournal(value); err == nil {
				t.Fatal("invalid staging prefix progress was accepted")
			}
		})
	}
}

func TestBuildJournalRejectsOverlappingAndAliasedTargetNamespaces(t *testing.T) {
	t.Run("nested directory target", func(t *testing.T) {
		root := t.TempDir()
		outerLive := filepath.Join(root, "live", "runtime")
		innerLive := filepath.Join(outerLive, "bin", "tool")
		outerSource := filepath.Join(root, "stage", "runtime")
		innerSource := filepath.Join(root, "stage", "tool")
		mustWrite(t, innerLive, "old-tool")
		mustWrite(t, filepath.Join(outerSource, "bin", "tool"), "new-runtime")
		mustWrite(t, innerSource, "new-tool")
		outerDigest, err := DigestPath(outerLive)
		if err != nil {
			t.Fatal(err)
		}
		innerDigest, err := DigestPath(innerLive)
		if err != nil {
			t.Fatal(err)
		}
		plan := Plan{TransactionID: "txn-nested", ProjectIdentity: "/project", Targets: []Target{
			{Class: "a", Identifier: "outer", LivePath: outerLive, StagedSource: outerSource, PreimageDigest: outerDigest},
			{Class: "b", Identifier: "inner", LivePath: innerLive, StagedSource: innerSource, PreimageDigest: innerDigest},
		}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("directory symlink alias", func(t *testing.T) {
		root := t.TempDir()
		realLive := filepath.Join(root, "live", "real")
		aliasLive := filepath.Join(root, "live", "alias")
		mustWrite(t, filepath.Join(realLive, "file"), "old")
		if err := os.Symlink(realLive, aliasLive); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		sourceA := filepath.Join(root, "stage", "a")
		sourceB := filepath.Join(root, "stage", "b")
		mustWrite(t, filepath.Join(sourceA, "file"), "new-a")
		mustWrite(t, filepath.Join(sourceB, "file"), "new-b")
		digest, err := DigestPath(realLive)
		if err != nil {
			t.Fatal(err)
		}
		plan := Plan{TransactionID: "txn-directory-alias", ProjectIdentity: "/project", Targets: []Target{
			{Class: "a", Identifier: "real", LivePath: realLive, StagedSource: sourceA, PreimageDigest: digest},
			{Class: "b", Identifier: "alias", LivePath: aliasLive, StagedSource: sourceB, PreimageDigest: digest},
		}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("file hard-link alias", func(t *testing.T) {
		root := t.TempDir()
		firstLive := filepath.Join(root, "live", "first")
		secondLive := filepath.Join(root, "live", "second")
		mustWrite(t, firstLive, "old")
		if err := os.Link(firstLive, secondLive); err != nil {
			t.Skipf("hard links unavailable: %v", err)
		}
		first := fileTarget(t, "a", "first", firstLive, filepath.Join(root, "stage", "first"), "old", "new-first")
		second := fileTarget(t, "b", "second", secondLive, filepath.Join(root, "stage", "second"), "old", "new-second")
		plan := Plan{TransactionID: "txn-hardlink-alias", ProjectIdentity: "/project", Targets: []Target{first, second}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("cross-target live and sidecar collision", func(t *testing.T) {
		root := t.TempDir()
		transactionID := "txn-cross-sidecar"
		first := fileTarget(t, "a", "first", filepath.Join(root, "live", "first"), filepath.Join(root, "stage", "first"), "old-first", "new-first")
		secondLive := sidecarPath(first.LivePath, transactionID, 0, "backup")
		secondSource := filepath.Join(root, "stage", "second")
		mustWrite(t, secondSource, "new-second")
		second := Target{Class: "b", Identifier: "second", LivePath: secondLive, StagedSource: secondSource, PreimageDigest: DigestAbsent}
		plan := Plan{TransactionID: transactionID, ProjectIdentity: "/project", Targets: []Target{first, second}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("live collides with own sidecar", func(t *testing.T) {
		root := t.TempDir()
		transactionID := "txn-own-sidecar"
		placeholder := filepath.Join(root, "live", "placeholder")
		live := sidecarPath(placeholder, transactionID, 0, "backup")
		source := filepath.Join(root, "stage", "source")
		mustWrite(t, source, "new")
		plan := Plan{TransactionID: transactionID, ProjectIdentity: "/project", Targets: []Target{{
			Class: "class", Identifier: "id", LivePath: live, StagedSource: source, PreimageDigest: DigestAbsent,
		}}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("live collides with cleanup tomb", func(t *testing.T) {
		root := t.TempDir()
		transactionID := "txn-cleanup-tomb"
		first := fileTarget(t, "a", "first", filepath.Join(root, "live", "first"), filepath.Join(root, "stage", "first"), "old-first", "new-first")
		secondLive := sidecarPath(first.LivePath, transactionID, 0, "backup") + ".delete"
		secondSource := filepath.Join(root, "stage", "second")
		mustWrite(t, secondSource, "new-second")
		second := Target{Class: "b", Identifier: "second", LivePath: secondLive, StagedSource: secondSource, PreimageDigest: DigestAbsent}
		plan := Plan{TransactionID: transactionID, ProjectIdentity: "/project", Targets: []Target{first, second}}
		assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
	})

	t.Run("live overlaps journal namespace", func(t *testing.T) {
		root := t.TempDir()
		home := filepath.Join(root, "home")
		engine := mustEngine(t, home)
		source := filepath.Join(root, "stage", "source")
		mustWrite(t, source, "new")
		plan := Plan{TransactionID: "txn-journal-overlap", ProjectIdentity: "/project", Targets: []Target{{
			Class: "class", Identifier: "id", LivePath: filepath.Join(engine.journalRoot, "consumer"), StagedSource: source, PreimageDigest: DigestAbsent,
		}}}
		assertPlanRejectedBeforeJournal(t, home, plan)
	})
}

func assertPlanRejectedBeforeJournal(t *testing.T, home string, plan Plan) {
	t.Helper()
	engine := mustEngine(t, home)
	if _, _, err := engine.buildJournal(plan); err == nil {
		t.Fatal("overlapping or aliased plan was accepted")
	}
	if _, err := os.Lstat(engine.journalPath(plan.TransactionID)); !os.IsNotExist(err) {
		t.Fatalf("invalid plan wrote a journal: %v", err)
	}
}

func TestPrepareFaultAndPreparingRecoveryCleanStagingWithoutLiveMutation(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	injected := errors.New("stop after prepare")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointPrepared {
			return injected
		}
		return nil
	}}))
	if _, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-prepare-fault", ProjectIdentity: "/project", Targets: []Target{target}}); !errors.Is(err, injected) {
		t.Fatalf("prepare error = %v", err)
	}
	if got := mustRead(t, target.LivePath); got != "old" {
		t.Fatalf("live target changed during aborted preparation: %q", got)
	}
	if _, err := os.Lstat(engine.journalPath("txn-prepare-fault")); !os.IsNotExist(err) {
		t.Fatalf("aborted preparation journal remains: %v", err)
	}

	restarted := mustEngine(t, filepath.Join(root, "home"))
	plan := Plan{TransactionID: "txn-preparing-restart", ProjectIdentity: "/different/project", Targets: []Target{target}}
	journal, sources, err := restarted.buildJournal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if err := restarted.saveJournal(journal); err != nil {
		t.Fatal(err)
	}
	if err := copyTarget(sources[0], journal.Targets[0].StagedPath); err != nil {
		t.Fatal(err)
	}
	journal.Targets[0].StagingIndex = len(journal.Targets[0].StagingEntries)
	if err := restarted.saveJournal(journal); err != nil {
		t.Fatal(err)
	}
	if err := restarted.Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, target.LivePath); got != "old" {
		t.Fatalf("preparing recovery changed live target: %q", got)
	}
	if _, err := os.Lstat(journal.Targets[0].StagedPath); !os.IsNotExist(err) {
		t.Fatalf("preparing recovery left staging: %v", err)
	}
}

func TestCleanupPhaseRecoveryRetainsDesiredAndRemovesBackups(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "live"))
	mustMkdirAll(t, filepath.Join(root, "stage"))
	target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
	injected := errors.New("crash before cleanup")
	engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
		if event.Point == PointBeforeCleanup {
			return injected
		}
		return nil
	}}))
	journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-cleanup-restart", ProjectIdentity: "/project", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, injected) {
		t.Fatalf("commit error = %v", err)
	}
	stored, err := engine.loadJournal(journal.TransactionID)
	if err != nil || stored.Phase != PhaseCleanup {
		t.Fatalf("cleanup journal = %+v, %v", stored, err)
	}
	if _, err := os.Lstat(stored.Targets[0].BackupPath); err != nil {
		t.Fatalf("backup removed before cleanup recovery: %v", err)
	}
	if err := mustEngine(t, filepath.Join(root, "home")).Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, target.LivePath); got != "new" {
		t.Fatalf("cleanup recovery target = %q", got)
	}
	if _, err := os.Lstat(stored.Targets[0].BackupPath); !os.IsNotExist(err) {
		t.Fatalf("cleanup recovery left backup: %v", err)
	}
}

func TestCleanupPreservesUnknownOrSimultaneousTombs(t *testing.T) {
	tests := []struct {
		name    string
		restart bool
	}{
		{name: "commit cleanup"},
		{name: "restart cleanup", restart: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			mustMkdirAll(t, filepath.Join(root, "live"))
			mustMkdirAll(t, filepath.Join(root, "stage"))
			target := fileTarget(t, "class", "id", filepath.Join(root, "live", "target"), filepath.Join(root, "stage", "target"), "old", "new")
			stop := errors.New("stop before cleanup")
			engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
				if testCase.restart && event.Point == PointBeforeCleanup {
					return stop
				}
				return nil
			}}))
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-unknown-tomb", ProjectIdentity: "/project", Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			unknownTomb := journal.Targets[0].BackupPath + ".delete"
			mustWrite(t, unknownTomb, "unknown")
			err = engine.Commit(testLock{}, journal.TransactionID)
			if testCase.restart && !errors.Is(err, stop) {
				t.Fatalf("paused commit error = %v", err)
			}
			if testCase.restart {
				err = mustEngine(t, filepath.Join(root, "home")).Recover(testLock{})
			}
			if !errors.Is(err, ErrImplementationCorruption) {
				t.Fatalf("cleanup error = %v, want implementation-corruption", err)
			}
			if got := mustRead(t, target.LivePath); got != "new" {
				t.Fatalf("desired target changed: %q", got)
			}
			if got := mustRead(t, unknownTomb); got != "unknown" {
				t.Fatalf("unknown tomb changed: %q", got)
			}
			if _, err := os.Lstat(journal.Targets[0].BackupPath); err != nil {
				t.Fatalf("recorded backup was removed: %v", err)
			}
			if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); err != nil {
				t.Fatalf("recovery journal was removed: %v", err)
			}
		})
	}
}

func TestCleanupRecoveryResumesOwnedDirectoryTombAtEveryRemovalBoundary(t *testing.T) {
	tests := []struct {
		name    string
		point   Point
		restart bool
	}{
		{name: "direct after tomb rename", point: PointAfterCleanupRename},
		{name: "restart after tomb rename", point: PointAfterCleanupRename, restart: true},
		{name: "direct during recursive removal", point: PointDuringCleanupRemoval},
		{name: "restart during recursive removal", point: PointDuringCleanupRemoval, restart: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			live := filepath.Join(root, "live", "target")
			stage := filepath.Join(root, "stage", "target")
			mustWrite(t, filepath.Join(live, "a"), "old-a")
			mustWrite(t, filepath.Join(live, "b"), "old-b")
			mustWrite(t, filepath.Join(stage, "a"), "new-a")
			mustWrite(t, filepath.Join(stage, "b"), "new-b")
			preimage, err := DigestPath(live)
			if err != nil {
				t.Fatal(err)
			}
			target := Target{Class: "runtime", Identifier: "target", LivePath: live, StagedSource: stage, PreimageDigest: preimage}
			injected := errors.New("cleanup removal fault")
			faulted := false
			var engine *Engine
			engine = mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{
				Fault: func(event Event) error {
					if !faulted && event.Point == testCase.point {
						faulted = true
						return injected
					}
					return nil
				},
				Observe: func(event Event) {
					if event.Point != PointAfterCleanupRename && event.Point != PointDuringCleanupRemoval {
						return
					}
					stored, err := engine.loadJournal(event.TransactionID)
					if err != nil {
						t.Fatalf("load cleanup boundary journal: %v", err)
					}
					if stored.RemovalPath == "" {
						t.Fatal("cleanup ownership was not durable before tomb mutation")
					}
				},
			}))
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-owned-tomb", ProjectIdentity: "/project", Targets: []Target{target}})
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, injected) {
				t.Fatalf("commit error = %v, want cleanup fault", err)
			}
			stored, err := engine.loadJournal(journal.TransactionID)
			if err != nil {
				t.Fatal(err)
			}
			backup := stored.Targets[0].BackupPath
			if stored.RemovalPath != backup {
				t.Fatalf("durable removal path = %q, want %q", stored.RemovalPath, backup)
			}
			if _, err := os.Lstat(backup + ".delete"); err != nil {
				t.Fatalf("owned cleanup tomb missing: %v", err)
			}
			if testCase.point == PointDuringCleanupRemoval {
				if digest, err := DigestPath(backup + ".delete"); err != nil || digest == stored.Targets[0].BackupDigest {
					t.Fatalf("directory tomb was not partially removed: digest=%q err=%v", digest, err)
				}
			}

			if testCase.restart {
				err = mustEngine(t, filepath.Join(root, "home")).Recover(testLock{})
			} else {
				engine.hooks = Hooks{}
				err = engine.Commit(testLock{}, journal.TransactionID)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := mustRead(t, filepath.Join(live, "a")); got != "new-a" {
				t.Fatalf("desired target changed during cleanup recovery: %q", got)
			}
			for _, path := range []string{backup, backup + ".delete", engine.journalPath(journal.TransactionID)} {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatalf("cleanup recovery left %s: %v", path, err)
				}
			}
		})
	}
}

func TestRollbackCleanupRecoveryResumesOwnedDirectoryTomb(t *testing.T) {
	for _, point := range []Point{PointAfterCleanupRename, PointDuringCleanupRemoval} {
		t.Run(string(point), func(t *testing.T) {
			root := t.TempDir()
			live := filepath.Join(root, "live", "target")
			stage := filepath.Join(root, "stage", "target")
			mustWrite(t, filepath.Join(live, "a"), "old-a")
			mustWrite(t, filepath.Join(live, "b"), "old-b")
			mustWrite(t, filepath.Join(stage, "a"), "new-a")
			mustWrite(t, filepath.Join(stage, "b"), "new-b")
			preimage, err := DigestPath(live)
			if err != nil {
				t.Fatal(err)
			}
			startRollback := errors.New("start rollback")
			cleanupFault := errors.New("rollback cleanup fault")
			rollbackStarted := false
			cleanupFaulted := false
			engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
				if !rollbackStarted && event.Point == PointTargetCommitted {
					rollbackStarted = true
					return startRollback
				}
				if rollbackStarted && !cleanupFaulted && event.Point == point {
					cleanupFaulted = true
					return cleanupFault
				}
				return nil
			}}))
			journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-rollback-owned-tomb", ProjectIdentity: "/project", Targets: []Target{{
				Class: "runtime", Identifier: "target", LivePath: live, StagedSource: stage, PreimageDigest: preimage,
			}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.Commit(testLock{}, journal.TransactionID); !errors.Is(err, startRollback) || !errors.Is(err, cleanupFault) {
				t.Fatalf("commit error = %v, want rollback and cleanup faults", err)
			}
			stored, err := engine.loadJournal(journal.TransactionID)
			if err != nil {
				t.Fatal(err)
			}
			if stored.Phase != PhaseRollbackCleanup || stored.RemovalPath != stored.Targets[0].RollbackPath {
				t.Fatalf("rollback cleanup progress = phase %s path %q", stored.Phase, stored.RemovalPath)
			}
			if err := mustEngine(t, filepath.Join(root, "home")).Recover(testLock{}); err != nil {
				t.Fatal(err)
			}
			if got := mustRead(t, filepath.Join(live, "a")); got != "old-a" {
				t.Fatalf("restored target changed during rollback cleanup recovery: %q", got)
			}
			if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
				t.Fatalf("rollback cleanup journal remains: %v", err)
			}
		})
	}
}

func TestCleanupRecoveryPreservesConcurrentOwnedTombState(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		for _, point := range []Point{PointAfterCleanupRename, PointDuringCleanupRemoval} {
			for _, restart := range []bool{false, true} {
				for _, mutation := range []string{"added", "replaced"} {
					name := fmt.Sprintf("rollback=%t/%s/restart=%t/%s", rollback, point, restart, mutation)
					t.Run(name, func(t *testing.T) {
						root := t.TempDir()
						live := filepath.Join(root, "live", "target")
						stage := filepath.Join(root, "stage", "target")
						mustWrite(t, filepath.Join(live, "a"), "old-a")
						mustWrite(t, filepath.Join(live, "b"), "old-b")
						mustWrite(t, filepath.Join(stage, "a"), "new-a")
						mustWrite(t, filepath.Join(stage, "b"), "new-b")
						preimage, err := DigestPath(live)
						if err != nil {
							t.Fatal(err)
						}
						startRollback := errors.New("start rollback")
						cleanupFault := errors.New("cleanup fault")
						rollbackStarted := false
						cleanupFaulted := false
						engine := mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{Fault: func(event Event) error {
							if rollback && !rollbackStarted && event.Point == PointTargetCommitted {
								rollbackStarted = true
								return startRollback
							}
							if (!rollback || rollbackStarted) && !cleanupFaulted && event.Point == point {
								cleanupFaulted = true
								return cleanupFault
							}
							return nil
						}}))
						journal, err := engine.Prepare(testLock{}, Plan{TransactionID: "txn-owned-tomb-concurrent", ProjectIdentity: "/project", Targets: []Target{{
							Class: "runtime", Identifier: "target", LivePath: live, StagedSource: stage, PreimageDigest: preimage,
						}}})
						if err != nil {
							t.Fatal(err)
						}
						err = engine.Commit(testLock{}, journal.TransactionID)
						if !errors.Is(err, cleanupFault) || (rollback && !errors.Is(err, startRollback)) {
							t.Fatalf("commit error = %v, want cleanup fault", err)
						}
						stored, err := engine.loadJournal(journal.TransactionID)
						if err != nil {
							t.Fatal(err)
						}
						if stored.RemovalPath == "" || len(stored.RemovalEntries) == 0 {
							t.Fatalf("durable removal provenance is missing: %+v", stored)
						}
						tomb := stored.RemovalPath + ".delete"
						changedPath := filepath.Join(tomb, "foreign-concurrent-bytes")
						changedContent := "foreign-added"
						if mutation == "replaced" {
							changedPath = filepath.Join(tomb, "a")
							changedContent = "foreign-replacement"
						}
						mustWrite(t, changedPath, changedContent)

						engine.hooks = Hooks{}
						if restart {
							err = mustEngine(t, filepath.Join(root, "home")).Recover(testLock{})
						} else {
							err = engine.Commit(testLock{}, journal.TransactionID)
						}
						if !errors.Is(err, ErrImplementationCorruption) {
							t.Fatalf("cleanup recovery error = %v, want implementation-corruption", err)
						}
						if got := mustRead(t, changedPath); got != changedContent {
							t.Fatalf("concurrent tomb bytes changed: %q", got)
						}
						wantLive := "new-a"
						if rollback {
							wantLive = "old-a"
						}
						if got := mustRead(t, filepath.Join(live, "a")); got != wantLive {
							t.Fatalf("consumer target changed: %q, want %q", got, wantLive)
						}
						if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); err != nil {
							t.Fatalf("recovery journal was removed: %v", err)
						}

						if mutation == "added" {
							if err := os.Remove(changedPath); err != nil {
								t.Fatal(err)
							}
						} else {
							recordedContent := "old-a"
							if rollback {
								recordedContent = "new-a"
							}
							mustWrite(t, changedPath, recordedContent)
						}
						if err := mustEngine(t, filepath.Join(root, "home")).Recover(testLock{}); err != nil {
							t.Fatalf("recovery after restoring recorded state: %v", err)
						}
						if _, err := os.Lstat(engine.journalPath(journal.TransactionID)); !os.IsNotExist(err) {
							t.Fatalf("recovered journal remains: %v", err)
						}
					})
				}
			}
		}
	}
}

func TestRestartDiscoversAndFinishesJournalCleanupTomb(t *testing.T) {
	engine, journal := committedCleanupFixture(t)
	target := &journal.Targets[0]
	if err := removeDurably(target.BackupPath, target.BackupDigest); err != nil {
		t.Fatal(err)
	}
	journalPath := engine.journalPath(journal.TransactionID)
	journalTomb := journalPath + ".delete"
	if err := durableRenameNoReplace(journalPath, journalTomb); err != nil {
		t.Fatal(err)
	}
	home := filepath.Dir(filepath.Dir(filepath.Dir(engine.journalRoot)))
	if err := mustEngine(t, home).Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(journalTomb); !os.IsNotExist(err) {
		t.Fatalf("journal cleanup tomb remains: %v", err)
	}
}

func TestDigestCopyAndNoReplaceRejectUnsafeOrUnknownPaths(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")
	if digest, err := DigestPath(missing); err != nil || digest != DigestAbsent {
		t.Fatalf("absent digest = %q, %v", digest, err)
	}
	file := filepath.Join(root, "file")
	mustWrite(t, file, "bytes")
	link := filepath.Join(root, "link")
	if err := os.Symlink(file, link); err == nil {
		if _, err := DigestPath(link); err == nil {
			t.Fatal("symlink digest was accepted")
		}
		if err := copyTarget(link, filepath.Join(root, "copied-link")); err == nil {
			t.Fatal("symlink staging was accepted")
		}
	}
	destination := filepath.Join(root, "destination")
	mustWrite(t, destination, "existing")
	if err := copyTarget(file, destination); err == nil {
		t.Fatal("copy overwrote existing destination")
	}
	from := filepath.Join(root, "from")
	mustWrite(t, from, "from")
	if err := durableRenameNoReplace(from, destination); err == nil {
		t.Fatal("exclusive rename overwrote destination")
	}
	if got := mustRead(t, destination); got != "existing" {
		t.Fatalf("exclusive rename changed destination: %q", got)
	}
}
