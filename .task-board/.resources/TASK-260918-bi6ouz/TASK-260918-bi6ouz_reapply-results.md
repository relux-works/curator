TASK-260918-bi6ouz developer results: re-applied after converge (rev7 onto main 1720a0fd / v0.15.0-rc.5). Plain text, no archives.

WHAT WAS DONE
Re-applied the accepted rev7 managed.go delta onto today main. The six carried files (managed_dotfile_test.go, managed_dotfile_conformance_test.go, testdata/dotfile-manager-table-802caee.md, takeover_test.go addition, docs/troubleshooting.md section, 7 platform-cases.tsv rows) were kept unchanged; no test-file API adjustments were needed (pinHomes, pinOperatorHome, installIdleProfile, claudeHome, UseWithPolicy all still exist with the same signatures).

WHAT CHANGED VERSUS REV7 AND WHY
1. inventoryUnmanaged is now stateread-based (stateread.Lstat/Readlink), not the os.Lstat body rev7 patched. Kept the stateread body verbatim and added only the (bool, error) return: true solely on the plain exists-and-no-marker path, false for foreign-manager symlinks, unreadable metadata, and clean. Rationale: preserve the post-rev7 read-failure discipline while carrying the delta semantics.
2. New call site internal/envprofile/global.go preflightManagedProfile (did not exist in the rev7 base) adapted as if _, err := ... — error-only use, behavior unchanged.
3. applyPlan call site (also new since rev7; emits no suspect hint) adapted as if _, err := ... — behavior unchanged.
4. repairUnderLock call site carries the delta exactly: takeover notice on inventoryErr, environment_foreign_manager_suspected gated on plainUnmanaged.
5. Table, resolution, and lstat code (dotfileStateTable, cell, dotfilePlatformForGOOS, resolveDotfileStatePath, foreignManagerHintAt, foreignManagerHint) are the delta text; the seed/dotfile refactors from the v2 writer flip and Codex seed revision B are untouched. Fixture SHA-256 verified 6575912115cf5b87facc0a6981e2d17bbcc1716f6ca7db3e0da5ddbf98853670, matching the source comment.
6. switch.go foreignManagerHint caller (own inline inventory, already gated on len(taken)>0) needed no change: foreignManagerHint keeps its () string signature.
7. Docs section confirmed single at docs/troubleshooting.md:3 with accurate detection-semantics text; tsv carries exactly 7 new rows, no duplicates, neighbor rows intact.

VERDICT-ANSWER COVERAGE (carried, passing)
- Regression: TestUnreadableDotfileStateKeepsTakeoverQuiet (rev6 gate run 36205109346 failure).
- Narrowing mutant: TestTakeoverSymlinkedDotfileManagerStateDoesNotWarn drives the production takeover path with a directory symlink at the resolved chezmoi path — catches replacing production Lstat with Stat.

TESTS RUN (all via mini-build-lock run bi6ouz, R193; no cmd/curator locally, R194)
- go test ./internal/envprofile -run Dotfile|Takeover|Managed -count=1: exit 0, ok 30.653s.
- Verbose run of the 8 dotfile/takeover tests: 7 PASS, TestDotfileManagerVectors SKIP (CURATOR_CONFORMANCE_ROOT unset — the declared skip class).
- TestDotfileManagerVectors against local spec-main CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1 (43bf0a2, verified at/after 802caee): PASS, executed 6 of 22 vector cases on macos (13 linux / 3 windows run on their native CI lanes); vector table assertion green.
- gofmt clean; no dotfileStateDirs references remain.

NOT RUN (hosted gate owns the rest per the republish instruction)
Full package suite, go vet/lint, cmd/curator, Windows/Linux lanes.

NO CHANGELOG OR LOGBOOK EDITS per the republish instruction.
SUCCESSOR VERIFICATION (run RUN-261008-759b67, 2026-10-08; predecessor run ended without handoff)
- Worktree diff re-checked against the rev7 delta intent: table/resolution/lstat code is the delta text; inventoryUnmanaged keeps the stateread body with only the (bool, error) return added; global.go and applyPlan call sites take _, err; repairUnderLock gates the suspect hint on plainUnmanaged. global.go diff is exactly the one-line call-site adaptation.
- Mandated mask re-run by successor via mini-build-lock: exit 0, ok ~21s. Verbose: 10 PASS + 1 SKIP (TestDotfileManagerVectors skips, CURATOR_CONFORMANCE_ROOT unset). Count note: the "7 of 8" line above undercounts the mask — the mask also matches pre-existing Managed/Takeover tests (TestImportSkipsManagedRootContext, TestUseWithoutTakeoverRefusesUnmanaged, TestUseTakeoverBacksUpAndNotifies, TestRepairTakeoverProvisionsManagedHome); all green.
- Extra single-test run (same lock): TestForeignManagerHintLstatDiscipline — PASS, 5/5 subtests. Rationale: the mandated mask misses it ("Manager" does not contain "Managed"), and it is part of the carried suite.
- Fixture SHA-256 re-verified: 6575912115cf5b87facc0a6981e2d17bbcc1716f6ca7db3e0da5ddbf98853670. gofmt clean on all touched Go files. Tree holds only the 8 expected paths; no CHANGELOG/LOGBOOK/stray edits.
- Not re-run by successor (accepted from above + hosted gate): vectors against spec-main checkout, full suite, vet/lint, cmd/curator, Windows/Linux lanes.

GATE-FAILURE FIX (run RUN-261008-a76a02, 2026-10-08; attempt 2/3 after rev9 CR validation failed)
- Symptom: rev8 and rev9 remote-gate runs (37744231088, 37749874814) failed identically on all lanes that run go test (Test macos/ubuntu/windows, Race macos/ubuntu) with "go test exit=1, platform-case gate exit=0". Trunk 1720a0fd itself is green on main (run 37731698590), so the failure is caused by this task's patch.
- Root cause (from the gate's own go-test-served.json evidence artifact): exactly one failing test, internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded: "internal/envprofile/managed.go:foreignManagerHint tests not-exist after function-value.os.Lstat without a seam route or reviewed allowlist reason". The rev7 delta passed os.Lstat as an injected function value; trunk has since gained the manager-read guard audit, which flags any bare os-read selector used as a value. Local predecessors never saw it because they ran only the Dotfile|Takeover|Managed mask, which does not match the guard test. Confirmed identical single failure on the macos and ubuntu evidence artifacts (the guard is a platform-independent static AST scan).
- Fix (2 files, behavior-preserving): foreignManagerHintAt now takes statState func(string) (stateread.Metadata, error) instead of lstat func(string) (os.FileInfo, error), classifies via metadata.Kind (present vs absent/unreadable) exactly like inventoryUnmanaged, and production passes stateread.Lstat. This is seam-routing, not an allowlist: the absent-vs-unreadable distinction now flows through stateread.LstatWith semantics, which additionally guards the Windows ENOTDIR-style misclassification (missingPath ancestor walk) that the raw os.IsNotExist check lacked. Unit tests keep fault injection through the seam's own injectable boundary: new dotfileSeamLstat helper wraps the five existing fakes via stateread.LstatWith. No spec-visible behavior change: lstat (no symlink following), table order, continue-after-failure, and error preservation are unchanged; troubleshooting wording ("lstat semantics") still holds.
- Tests run by this successor (all via mini-build-lock run bi6ouz; no cmd/curator locally):
  - TestManagerOwnedAbsenceReadsAreGuarded (the gate-failing test): exit 0, PASS.
  - TestManagerReadScannerFindsAliasedNotExistCollapse (guard's mutant companion): exit 0, PASS.
  - Mandated mask Dotfile|Takeover|Managed with CURATOR_CONFORMANCE_ROOT unset: exit 0, 12 PASS + 1 SKIP (skip is TestDotfileManagerVectors, root unset).
  - Mandated mask with CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1 at 43bf0a2: exit 0 (vectors execute and pass).
  - TestDotfileManagerVectors against the same rc.14 root: exit 0, 6/6 macos subtests PASS.
  - TestForeignManagerHintLstatDiscipline after the seam refactor: exit 0, 5/5 subtests PASS.
  - go vet ./internal/envprofile: exit 0. gofmt clean on both touched Go files.
- Tree still holds only the 8 expected paths (5 modified + 2 new test files + testdata/); no CHANGELOG/LOGBOOK edits. No other callers of foreignManagerHintAt exist outside managed.go and its test.
- Not run (hosted gate owns the rest): full package suite, lint, cmd/curator, Windows/Linux lanes.
