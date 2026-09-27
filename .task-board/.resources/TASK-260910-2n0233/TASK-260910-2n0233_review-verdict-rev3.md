# TASK-260910-2n0233 review verdict — rev3: CHANGES REQUESTED

Reviewer: claude-opus-5-5 (low). Candidate tree 1cd3b6fa (worktree temp-index tree = 1cd3b6fa, verified), base 0ffe2e1d, 20 paths — all within this leaf, no CHANGELOG/LOGBOOK/build outputs.
Spec: rc.13 SPEC_PIN 23435129 (/tmp/spec-rc13 at that commit), registry §9.3 page boundary, §5 rollback.

## Independent runs (disposable clone, zsh, pipefail, CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1)
- go build ./... ; go vet registry+cmd/curator ; gofmt -l → exit 0, clean
- go test ./internal/registry → ok (exit 0); all 9 page_boundary_cases run (not skipped)
- TestManagerOwnedAbsenceReadsAreGuarded (internal/envprofile) → ok
- conformance-gaps.tsv: 0 rows owned by 25yc0h/2n0233 before and after (73 lines, unchanged)

## Mutants (internal/registry/boundary.go)
- M1 stale check removed (`if false && parsed.Version < …`, :108) → KILLED (4 tests fail)
- M4 chain mismatch removed (:97) → KILLED
- M3 ReadOnlyStateDir falls back on unreadable (:223) → KILLED
- **M2 unreadable high-water treated as absent in openPageChain (:66-68: on err set state={},exists=false and continue) → SURVIVES** internal/registry (ok 17.7s), cmd/curator and internal/install (-run Registry|Boundary|Rollback|Attest: ok/ok).

## Finding F1 (blocking) — review-note item 2 not proven at the production entry
The security read is `openPageChain` → `readSnapshotState` (boundary.go:66), reached from `newHTTPFetch` (http.go FetchFn) used by install and `status --attest`. The only unreadable rows cover `ReadOnlyStateDir` and `ReadBoundaryPosture`/`status --check` (boundary_test.go:124-172, main_test.go:848). No test drives an unreadable (or corrupt/invalid) `snapshot-<digest>.json` through the FetchFn: a mutant that treats it as first use lets a below-high-water page be accepted — and, in the persisting variant, overwrite the rollback state with the stale boundary — while every test stays green.

Required rework:
1. Add a row in internal/registry that builds the FetchFn (NewHTTPFetch / NewHTTPFetchWithPolicy persisting and the ReadOnly variant) against an httptest registry serving a below-high-water (or any) page, with the state file made unreadable (chmod 000 / directory in place of file, as stateread's rows do) and one with corrupt JSON; assert: error is the pageStateError (fail closed, no records), no records served, the state file bytes unchanged, record cache not written. M2 above must fail.
2. Optionally the same through a production caller (install resolve or `status --attest`) so the row is "driven".
3. Also consider the catalog-lists-it-but-missing branch (boundary.go:69) with a fetch-path row (not mutated here; stated bound).

Everything else (parsing, §9.3 order/precedence, cache non-write on rejection, posture rows, docs) looked correct in this bounded pass.
