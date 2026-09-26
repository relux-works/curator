# TASK-260922-cww1ov results — revision 5 (gate-flake analysis, no content change)

Status: ready for review. No file changed in this revision; candidate tree is
byte-identical to accepted revision 3 / republished revision 4
(`768bacfa2c52a0906b80e233e8af14a275137ede`, verified via temp index in this run).

## Why revision 5 exists

CR-TASK-260922-cww1ov-4 (revision 4 = revision 3 unchanged) went to
`changes_requested` SOLELY because the landing-suite gate
(run 35852658095, `sh scripts/remote-gate.sh`) failed the Windows lane:
`go test` exit=1 while the platform-case gate exited 0
(`test-gate: go test exit=1, platform-case gate exit=0`).
There is no reviewer prose verdict on revision 4 and no content finding to answer:
`TASK-260922-cww1ov_review-verdict-rev3.md` remains ACCEPT, and the Review Round
Brief's "rejection" premise does not match that verdict file (same template
mismatch already recorded in `TASK-260922-cww1ov_republish-rev4.md`).
Per brief R1, no new regression test or narrowing mutant is due for a finding
that does not exist; adding one would be scope expansion.

## The single Windows failure (evidence, not inference)

From the run's `test-evidence-windows-latest` artifact
(`test/go-test-served.json`), the ONLY failure:

- `internal/managerlock :: TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
  - `managerlock_test.go:532: uncontended helper with tiny deadline = "acquired", want blocked`
  - Package result: `FAIL github.com/relux-works/curator/internal/managerlock 2.647s`
  - Every other package green; all ledger platform-case rows `ok`, including all
    46 F-C3 rows (`TestSharedToIsolatedRemovesStaleLink`,
    `TestCredentialLinkRegularFileRefuses`, `TestMigrateNoSecretCopies`,
    `TestMigrateInterruptedApplyRecovers`, …) and the `cmd/curator` env rows.

`internal/managerlock` is outside this task's scope (owned by closed
BUG-260922-1vx45a area work; this leaf touches `internal/envprofile`, the env CLI
surface, `docs/troubleshooting`, `CHANGELOG`, and the platform-cases ledger only).

## It is a timing flake: same-tree pass vs fail

1. Revision 3 gate (run 35735733311) on the IDENTICAL tree 768bacfa: Windows
   green — `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked` passed.
2. Revision 4 gate (run 35852658095) on tree 768bacfa: the test failed.
3. Trunk CI on `48da2690` (no F-C3 content): Windows green.
4. A previous trunk Windows failure (run 35588481406, Sep 21) failed on a
   DIFFERENT timing-sensitive test
   (`internal/snapshot TestConcurrentGetAcceptsOneImmutablePublisher`), showing
   the Windows lane's background timing-flake rate in unrelated packages.

Mechanism (`internal/managerlock/filelock.go:22`, `managerlock.go:122`,
`managerlock_test.go:525-534,604-690`): the test spawns a helper subprocess with
a `time.Nanosecond` context deadline on an UNCONTENDED lock and requires
"blocked". `acquireFileLock` checks `ctx.Err()` before `MkdirAll`/`OpenFile` and
again after `tryFileLock`. Between `context.WithTimeout(1ns)` creation and the
first check, the helper runs `New` → `CanonicalProjects` → `prepare()` (two
Windows path canonicalizations plus `MkdirAll`). Whether the 1ns timer has fired
by the check depends on Windows timer granularity vs host speed: a fast/idle
host acquires first ("acquired" → FAIL); a loaded host observes the expired
deadline ("blocked" → PASS). The outcome is host-load dependent for identical
content — the definition of a flake. Deterministic alternatives (pre-expired
context, helper-contract change, product change in `managerlock`, ledger skip)
all violate this leaf's scope rules (R1, no new skip class, no test weakening),
so the fix belongs to the `managerlock` area, not this leaf.

Recommendation to orchestrator: file a follow-up bug task against
`internal/managerlock` for the `TestSubprocessExpectedAcquiredWithTinyDeadline`
intermittency (repro: fast Windows host, identical tree passes/fails across
runs 35735733311/35852658095); do NOT hold this leaf for it.

## Local re-verification in this run (zsh, `set -o pipefail`, `-count=1`)

- `go test -count=1 -timeout 9m -run 'Migrat|Recover|NoSecret' ./internal/envprofile/` → ok 38.2s, exit 0 (observed this run)
- `go test -count=1 -timeout 25m -v -skip 'Migrat|Recover|NoSecret' ./internal/envprofile/` → ok 1129.5s, exit 0 (247 PASS, 0 FAIL; an earlier identical run with `-timeout 9m` tripped the go-test timeout at 541s with zero test failures — slow-host artifact, same as the rev3 reviewer observed; the 25m rerun passes clean)
- `go test -count=1 -timeout 25m -run 'TestEnv' ./cmd/curator/` → ok 613.5s, exit 0

## Tree identity (this run, no file changed)

- `git status --short`: the same 7 modified + 1 untracked set as revision 4
  (`.github/ci/platform-cases.tsv`, `CHANGELOG.md`,
  `cmd/curator/env_migrate_test.go`, `docs/troubleshooting.md`,
  `internal/envprofile/credential_link_test.go`,
  `internal/envprofile/migrate_test.go`,
  `internal/envprofile/reviewer_recovery_test.go`,
  `?? internal/envprofile/credential_production_test.go`).
- Temp-index tree (`git read-tree HEAD` + `git add -A` + `git write-tree`):
  `768bacfa2c52a0906b80e233e8af14a275137ede` — matches the reviewer-recorded
  CR candidate tree. Trunk is still `48da2690`; no refresh needed.
- Directives for RUN-260923-3e20f9: none recorded.

## Coverage / mutant standing (unchanged, carried from accepted revisions)

Complete per revision-2 acceptance: coverage table with narrowing mutant per row
(conflict-admitted ⇒ test fails), dangling-target driven rows, migration
end-to-end incl. journal/recovery and whole-home no-copy scan, every refusal
class incl. the codex admission table, `resolve --repair` never migrates
silently; 46 ledger rows (37 all-platform w/ Windows host-capability tolerance +
5 all-platform/no-tolerance + 4 Unix-only); no new skip class; no test weakened;
CHANGELOG `### Added` (tests-only) + docs cross-references. See
`TASK-260922-cww1ov_results.md` / `TASK-260922-cww1ov_results-rev3.md`.
