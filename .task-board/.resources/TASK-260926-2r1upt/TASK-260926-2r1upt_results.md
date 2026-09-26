# TASK-260926-2r1upt results

## Change

Added `TestDraftFreshMachineTransitiveReplaySourceGuards` in `internal/install/draftsources_test.go`. Every scenario drives the production `Project` entry point, which reaches `replayMissingDraftSnapshots` and `declaredDependencyReplaySources` on a fresh home. The local Git fixture has a replayed `review` requirer and a transitive network-Git `qa` package selected from `roles/qa`.

Rows: 3/3 scenarios — directory mismatch, repository identity mismatch, and matching declaration replay. Both mismatch rows assert the specific `source_snapshot_changed` diagnostic from the declaration-to-lock guard. The matching row proves the declared source restores the locked package content and leaves `Skillfile.lock.json` byte-identical.

No production change was needed. No CHANGELOG.md or LOGBOOK.md edit was made; these findings are recorded here per the project rule.

## Per-guard mutant evidence

Each narrowed mutant was tested against the existing production-CLI acceptance test before running its new targeted row. The older acceptance row remained green; the dedicated row then failed under its corresponding mutant. Mutant source was restored from `/tmp/TASK-260926-2r1upt-draftsources.go.original` and byte-compared after each run (restore/cmp exit 0).

| Guard mutant | Existing CLI acceptance, before targeted row | Dedicated row after mutant | Result |
|---|---:|---:|---|
| Directory mismatch comparison narrowed to an unreachable consumer name | `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1` — exit 0 | `go test ./internal/install -run '^TestDraftFreshMachineTransitiveReplaySourceGuards/directory-mismatch$' -count=1 -v` — exit 1; wrong fallback diagnostic (`source_snapshot_unavailable`) instead of the expected directory guard diagnostic | Killed |
| Repository identity comparison narrowed to an unreachable consumer name | Same existing CLI acceptance command — exit 0 | `go test ./internal/install -run '^TestDraftFreshMachineTransitiveReplaySourceGuards/repository-identity-mismatch$' -count=1 -v` — exit 1; downstream identity refusal instead of the expected declaration guard diagnostic | Killed |

The pre-row CLI command exercises the valid fresh-machine transitive replay path and does not run the new `internal/install` guard subtests. This documents the previous coverage gap directly.

## Final validation

| Command | Exit | Result |
|---|---:|---|
| `go test ./internal/install -run '^TestDraftFreshMachineTransitiveReplaySourceGuards$' -count=1 -v` | 0 | All three production-entry rows pass |
| `go test ./internal/crossconformance -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' -count=1` | 0 | Existing production-CLI transitive replay acceptance passes unmutated |
| `go build -o /tmp/TASK-260926-2r1upt-curator ./cmd/curator` | 0 | CLI builds |
| `go vet ./internal/install` | 0 | No findings |
| `golangci-lint run ./internal/install` | 0 | 0 issues |
| `gofmt -l internal/install/draftsources_test.go` | 0 | No files listed |
| `git diff --check` | 0 | Clean |

Unfiltered `internal/install` and `internal/crossconformance` suites were not run for this focused task. The prior 11burj results record that their combined unfiltered run hit the host's 10-minute test timeout; this task's focused guard test and production-CLI acceptance both completed successfully.

During fixture development, three early runs exited 1 and were corrected: bare remote directories were not created before `git init`; the fixture struct omitted its dependency remote field; and the positive row attempted `AuthenticateGit` without supplying the replay repository. The final checks above were rerun after those corrections.
