package envprofile
import ("os"; "path/filepath"; "strings"; "testing")
func TestReviewerRecordedTempFile(t *testing.T) {
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 {
		t.Fatalf("one relink op: %+v", report.Ops())
	}
	owned := filepath.Join(filepath.Dir(piLink), ".migrate-owned-test.tmp")
	if err := os.WriteFile(owned, []byte("foreign backup"), 0o600); err != nil {
		t.Fatal(err)
	}
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    report.Hash,
		Ops: []migrationJournalOp{{
			Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
			Path: "auth.json", From: oldTarget, To: agentAuth,
			Temp: owned, TempTarget: agentAuth,
		}},
		Markers: []migrationJournalMarker{},
	}
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	retry := fx.migrateRequest()
	retry.Expect = report.Hash
	before := snapshotCredentialScope(t, fx)
 _, err = ApplyMigration(retry)
 if err == nil || !strings.Contains(err.Error(), "temporary path") { t.Fatalf("foreign recorded temp must refuse: %v", err) }
 if _, err := os.Lstat(owned); err != nil { t.Fatalf("foreign temp lost: %v",err) }
 if _, err := os.Lstat(migrationJournalPath(fx.home)); err != nil { t.Fatalf("journal lost: %v",err) }
 added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
 if len(added)+len(removed)+len(changed) != 0 { t.Fatalf("refusal mutated state: %v %v %v", added,removed,changed) }
}

func TestReviewerRecordedTempDirectory(t *testing.T) {
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 {
		t.Fatalf("one relink op: %+v", report.Ops())
	}
	owned := filepath.Join(filepath.Dir(piLink), ".migrate-owned-test.tmp")
	if err := os.Mkdir(owned, 0o700); err != nil {
		t.Fatal(err)
	}
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    report.Hash,
		Ops: []migrationJournalOp{{
			Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
			Path: "auth.json", From: oldTarget, To: agentAuth,
			Temp: owned, TempTarget: agentAuth,
		}},
		Markers: []migrationJournalMarker{},
	}
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	retry := fx.migrateRequest()
	retry.Expect = report.Hash
	before := snapshotCredentialScope(t, fx)
 _, err = ApplyMigration(retry)
 if err == nil || !strings.Contains(err.Error(), "temporary path") { t.Fatalf("foreign recorded temp must refuse: %v", err) }
 if _, err := os.Lstat(owned); err != nil { t.Fatalf("foreign temp lost: %v",err) }
 if _, err := os.Lstat(migrationJournalPath(fx.home)); err != nil { t.Fatalf("journal lost: %v",err) }
 added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
 if len(added)+len(removed)+len(changed) != 0 { t.Fatalf("refusal mutated state: %v %v %v", added,removed,changed) }
}

func TestReviewerRecordedTempSymlink(t *testing.T) {
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 {
		t.Fatalf("one relink op: %+v", report.Ops())
	}
	owned := filepath.Join(filepath.Dir(piLink), ".migrate-owned-test.tmp")
	if err := os.Symlink(oldTarget, owned); err != nil {
		t.Fatal(err)
	}
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    report.Hash,
		Ops: []migrationJournalOp{{
			Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
			Path: "auth.json", From: oldTarget, To: agentAuth,
			Temp: owned, TempTarget: agentAuth,
		}},
		Markers: []migrationJournalMarker{},
	}
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	retry := fx.migrateRequest()
	retry.Expect = report.Hash
	before := snapshotCredentialScope(t, fx)
 _, err = ApplyMigration(retry)
 if err == nil || !strings.Contains(err.Error(), "temporary path") { t.Fatalf("foreign recorded temp must refuse: %v", err) }
 if _, err := os.Lstat(owned); err != nil { t.Fatalf("foreign temp lost: %v",err) }
 if _, err := os.Lstat(migrationJournalPath(fx.home)); err != nil { t.Fatalf("journal lost: %v",err) }
 added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
 if len(added)+len(removed)+len(changed) != 0 { t.Fatalf("refusal mutated state: %v %v %v", added,removed,changed) }
}
