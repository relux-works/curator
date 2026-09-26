# TASK-260907-2as5sx review verdict — CR rev5: CHANGES REQUESTED

Candidate tree f4a90004 materialised in a disposable clone (tree match verified).

## Blocking
F1 (review-note item 2; DoD "a shared helper or lint prevents new collapses"): the AST inventory does NOT fail on a new
IsNotExist-then-default pattern outside the seam.
- internal/envprofile/state_read_sites_test.go:518-640 `TestKnownAbsenceSensitiveReadersUseSharedSeam` checks a fixed list of
  37 named functions only; internal/marker/read_state_test.go:94 is likewise a named list.
- Reproduced mutant: added internal/envprofile/zz_mutant.go with `readNewManagerState` doing
  `os.ReadFile(CurrentFile(home)+".new")` → `if os.IsNotExist(err) { return "default" }; return "default"` (collapses both).
  `go test ./internal/envprofile ./internal/marker -run 'TestKnownAbsenceSensitiveReadersUseSharedSeam|TestProductionMarkerCallersUseReadState'` → ok/ok. SURVIVES.
- So nothing structural prevents a sixth site — the exact gap this task exists to close.
Fix: make the guard deny-by-default: scan all non-test .go under internal/ and cmd/ (outside internal/stateread) for
os.ReadFile/ReadDir/Stat/Lstat/Open + os.IsNotExist/errors.Is(fs.ErrNotExist) on manager-owned state, with an explicit reviewed
allowlist (file:function + one-line reason) for non-state reads; report the ratio (N guarded / M scanned). Kill the mutant above
with the named test. Alternatively, if a repo-wide scan is judged infeasible, state why with alternatives weighed (AC clause).

## Verified OK
- stateread seam: absent / unreadable (manager_state_unreadable) / unusable kept distinct (stateread.go).
- Per-site narrowing mutant re-applied: status.go orphanHomes ReadDir error → `return nil, nil` killed by
  TestStatusUnreadableOrphanInventoryIsNotEmpty (state_read_sites_test.go:389).
- gofmt -l clean on changed .go; internal/godriver byte-identical to base.
Not re-run by me: full suite, crossconformance execution_policy, surfacing-order, Windows ledger rows (items 3-4) — re-check next cycle.
