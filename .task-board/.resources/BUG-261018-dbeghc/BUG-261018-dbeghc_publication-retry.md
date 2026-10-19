# BUG-261018-dbeghc — Publication retry

## Candidate identity

The six task files still match the last signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`, tree `088c305b5f9cfcd4c1bc913163cb10f6361e7551`, signature `G`. The four tracked task files compare byte-identically with that candidate (`git diff --exit-code <candidate> -- README.md docs/environment-config.md internal/envprofile/managed.go internal/envprofile/nofollow.go`, exit 0). Both untracked tests match its blobs: `managed_modes_mutants_unix_test.go` = `f7dc7896879f3ab282b17098650a351fbcf3b15d`, `managed_modes_unix_test.go` = `b6ffb58f18fd8c905111412a134c3b7fed3bbada`.

No source, test, README, CHANGELOG or other repository documentation file was edited. Story HEAD remains `9c1605a74436d2448f759bd4b66ae6685237d83b`; task changes remain uncommitted. `git diff --cached --exit-code` and `git diff --check` each exit 0. Incoming README changes are disjoint from the existing permission note: audit cache ruleset 4 and network-contact test documentation. No combination was installed because refresh refused before installation.

## Blocking refresh refusal

`task-board worktree status STORY-261018-ogw2f1` exits 0 and reports revision 2 stale. The assignment explicitly requires refresh because protected trunk advanced on README. Two direct invocations of `task-board worktree refresh-candidate BUG-261018-dbeghc`, with the mandated agent Git environment, each exit **1**:

`change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying`

Prepared checkpoint tree: `09b478fc86f54e1596c3fb07efb9857c50b6c186`. Workspace candidate tree: `088c305b5f9cfcd4c1bc913163cb10f6361e7551`. Refusal names eight tracked board paths belonging to the incoming reserved-TLD task: three activity streams and five element/progress files. The unchanged retry confirms this is persistent, not a successful refresh. Exact refusal log is attached separately.

## Required recovery

The failed assumption is that the producer refresh can install incoming tracked board changes. CLI help and project-management documentation expose no worker board synchronization route; `worktree converge` explicitly requires an orchestrator session. Direct board-file editing and clearing the tracked-run identity to impersonate an operator are prohibited. No such bypass was attempted.

Recommended external input: orchestrator/CLI owner restores the board/checkpoint projection through supported recovery, preserving this exact candidate, then resumes this producer for refresh and handoff. A CLI repair at its source is an alternative if supported operator recovery also refuses; that is outside this run's write boundary. Neither weakening the drift guard nor manually copying activity history is acceptable.

## Validation and bounds

No behavior changed in this retry. Previous full local evidence remains historical evidence for the old candidate; no validation of the combined new trunk tree is claimed. Full local suite, build, lint, permission mutants and hosted gate were not rerun in this publication-only retry: the required candidate refresh failed, and r4 delegates configured validation to handoff. No new signed combined candidate exists, and AC5 remains unchecked. No claim is made that runner repair alone establishes a green full gate. Prior AC coverage remains 4 of 5 rows driven; AC5 requires external validation. Prior test names, mutant evidence and contract bounds remain in the gaps file.

## Logbook

Publication retry found a new blocker before the formerly failing hosted cleanup could be exercised. Candidate, index and Story checkpoint preserved; only task-scoped scratch and board resources/notes changed. No new validation checklist item was added. Status is blocked pending supported board/checkpoint synchronization; review and integration are not claimed.
