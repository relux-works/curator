# TASK-261001-1mdlah — recovery evidence and blocker

The existing uncommitted v0.5.37 WIP is preserved. No product-code workaround was added.

## Concrete constraint

The binding brief requires Muse 1.4.1 refused as unlisted. The pinned upstream Muse `policy.go` explicitly lists BOTH 1.4.1 and 1.4.2. Its `probe_test.go` explicitly proves admission for both; upstream negative tests use 1.4.0 and 1.5.0. Root-entry tests reproduced admission of 1.4.1 in native and yolo modes.

Decision needed: authorize unlisted 1.4.0 as the refusal row (recommended), or supply an upstream tag excluding 1.4.1. The former preserves upstream policy ownership with no product-code exception; the latter needs upstream work and a changed pin. An asynchronous clarification was requested; no answer arrived. No launcher-owned release exception or test substitution was made.

## Commands run directly in this recovery

| Command | Observed exit | Evidence |
| --- | --- | --- |
| `go test -p 1 ./cmd/curator-run -run '^TestMuseV3' -count=1 -v` | 1 | Admission 3/3 passes; direct interactive plan modes 2/2 pass; duplicate-yolo refusals 3/3 pass; unknown/absent refusal rows 4/4 pass; required 1.4.1 refusal rows fail 2/2 because both admit. |
| `go build ./...` | 0 | Candidate compiles. |
| `make fmt-check` | 0 | Formatting clean. |
| `go vet ./...` | 0 | Native vet clean. |
| `GOOS=windows go vet ./...` | 1 | Existing POSIX syscall errors in execution and Mkfifo test packages. This is failing, never a Windows pass. |
| `go test -p 1 ./...` | 1 | Interrupted after command execution stalled; no output or suite summary. No passing full-suite claim. |

The fake binary prints exactly `Muse Code 1.4.2 (1.4.2-R4684.1)`, matching the upstream parser and its tests. The root rows assert exact argv, env, inherited HOME, four XDG overrides and stdin bytes. Muse native/yolo/alias goldens remain present; existing Claude goldens add only `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false`, documented under Unreleased for curator#102.

## Inherited evidence and limits

Earlier saved evidence in `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TASK-261001-1mdlah-2y_u7709` reports Windows baseline/candidate exit 1 with identical diagnostic line multisets, and HOME/double-yolo mutants killed 2/2. These observations were NOT rerun or independently revalidated in this recovery. Fresh baseline extraction stalled and was interrupted with exit 130; no fresh baseline comparison is claimed.

Command execution became unresponsive even for an echo probe. Pending diagnostic, baseline, directive, and first outcome attachment commands were interrupted (exit 130). The failed attachment is not claimed persisted. No required evidence from a command without a terminal result is claimed. Board persistence will be attempted separately; report its actual outcome.

The subsequent board mutation requesting a task-scoped outcome, blocker notes, and `status=blocked` also stalled, produced no response, and was interrupted with exit 130. Persistence is unknown; the last verified board status remains `development`. No successful board attachment or status change is claimed. All pending commands from this recovery were interrupted; none is intentionally left running.

No LOGBOOK per binding brief. Review handoff is withheld for the release-policy conflict and missing green full-suite evidence. Strict failing WIP refusal rows remain for the owner decision; work is not ready for review.
