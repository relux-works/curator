# TASK-260907-2as5sx review verdict — CR rev7: ACCEPTED

Candidate tree 9cc4c09c (base 948ae7c9) materialised via git archive into disposable /tmp/r7c.

## Rev5 F1 (named-list guard) — FIXED
- `TestManagerOwnedAbsenceReadsAreGuarded` is deny-by-default over all non-test .go in internal/ + cmd/ (excl. internal/stateread):
  ratio line `175 guarded via seam / 117 allowlisted / 292 scanned (403 production files)` — reproduced.
- Rev5 mutant re-applied (internal/envprofile/zz_mut.go `readNewManagerState`: os.ReadFile; IsNotExist→"default"; err→"default"):
  named test FAILS, 292/293, names `zz_mut.go:readNewManagerState`. KILLED. Removed → pass.
- Allowlist: exact file:function keys with one-line reasons.

## Per-site narrowing mutant (fresh)
- import.go readRootSurface: stateread error → `return nil, nil` (collapse to absence) → `TestImportAbsentAndUnreadableRootSurfaceDiffer`
  FAILS ("unreadable root error = <nil>"). KILLED. Restored → pass. (Rev5 already killed status.go orphanHomes mutant.)

## macOS retention test (item 2)
- Producer: base 948ae7c9 -count=30 ok, pre-refresh candidate ok, refreshed candidate ok; test injects countingGeneration so the seam is
  not on its path. Test untouched in the patch. My rerun on the candidate: -count=15 ok. Accepted as unexplained pre-existing/transient,
  no seam cause.

## Refresh (item 3)
- R5 reader runtimestore/enforced.go:ManagedEnforcedShimsIn migrated to stateread.ReadDir; test + platform-cases row (windows platform-control).
- No CHANGELOG.md / LOGBOOK.md / .task-board in the delta; entry text in results.

## Items 3-4 of rev5 note / validation
- go build ./... + go vet (stateread, runtimestore) ok; go test stateread, runtimestore, marker, contextstore, ui, transaction, contextlock ok.
- gate-selftest 198/0. Windows ledger rows present for the mode-bit tests.
- Not rerun by me: full internal/envprofile and internal/install suites (exceed 10-min bound locally; hosted gate is arbiter).

## Residual (non-blocking, advisory)
- R1: the guard scans *absence-sensitive* reads (read + not-exist test). A reader that swallows every error without a not-exist test
  (`b, err := os.ReadFile(p); if err != nil { return "default" }`) in internal/scopes passes the guard (verified). This is outside the
  rev5-requested scope (which named IsNotExist/ErrNotExist patterns) but should be stated as the guard's bound in docs/results;
  candidate follow-up: extend scan to catch-all error→fallback on manager-owned paths.
