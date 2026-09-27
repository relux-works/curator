# TASK-260910-2n0233 review verdict — rev4: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate: base d41da0fb, tree c05b7e41. The worktree temp-index write-tree equals c05b7e41. Review ran in a disposable rsync copy (zsh, pipefail).
CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc13/conformance/v1.

## Rev3 F1 (M2 survivor) — fixed
- boundary_test.go:170 TestFetchFnRejectsUnreadableOrCorruptHighWaterBeforeNetwork. It covers all 4 FetchFn builders (NewHTTPFetch, WithPolicy persistent, WithPolicy read-only, WithPolicyReadOnly; builders at :129) × {directory-in-place, corrupt JSON}. The catalog is deliberately left without the entry, so only the unreadable check can reject. The test asserts pageStateError, no records, 0 network calls, state/catalog bytes unchanged and no record-cache entries.
- boundary_test.go:246 covers the catalog-listed-but-missing branch (boundary.go:69) for all 4 builders.
- chmod-000 variant: not added. The directory-in-place variant runs on every platform with no skip, so no ledger row is needed. This is accepted as the portable form of the unreadable row.

## Validation (real exit codes)
- go test ./internal/registry/ → exit 0
- go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded → exit 0
- go vet ./internal/registry/ → 0; gofmt -l internal cmd → empty

## Mutants (boundary.go; each restored afterwards)
| id | mutation | result |
|---|---|---|
| M1 | below-high-water check `<` → false (:112) | killed (6 FAIL) |
| M2 | unreadable state → treated as absent (:68) | killed (13 FAIL) |
| M3 | catalog-listed missing ignored (:69) | killed (TestFetchFnRejectsCatalogListedMissingHighWater, all 4 builders) |
| M4 | equal version: head comparison dropped (:120) | killed |
| M5 | chain byte-identity check disabled (:100) | killed (9 FAIL) |
| M6 | VerifySigned dropped from boundary verification | killed (7 FAIL) |

## Rebase fidelity / scope
- Paths touched both by trunk (0ffe2e1d..d41da0fb) and by the candidate: env.go, env_test.go, main.go, cli.md, envprofile/status.go, registry_test.go.
- Candidate removals on those paths: 0 lines, except main.go (2) and registry_test.go (10). Every one of those is the intended rewrite from the NewHTTPFetch signature/state-dir change and the v2 page envelope. No trunk revert.
- The diff contains no CHANGELOG or LOGBOOK changes and no stray files.

## Residuals (non-blocking)
- No chmod-000 row (the directory variant covers it portably).
