# TASK-260728-20ao7p — integration precheck (RUN-261001-27a48a, 2026-10-01T06:53:03Z)

Bound integration run for accepted CR `CR-TASK-260728-20ao7p-2` revision 2.
This run changed no file, ran no landing command, and leaves board writes to the runner's synchronous landing transaction.

## Preconditions — all hold

- Board: `TASK-260728-20ao7p` = `integrating`, `STORY-260728-1ojb1p` = `integrating`.
- Siblings `BUG-260901-393lbo` and `TASK-260728-pwbr32` are both `done`: this task is the final leaf.
- `integrationCheckpointed=false` (correct: checkpoint happens at integrate, not before).
- CR rev2 accepted (activity AEV-261001-45961b55…, `ready → accepted` 2026-10-01T04:11:48Z; reviewer verdict resource `TASK-260728-20ao7p_review-verdict-rev2.md`).
- Checklist 12/12 done.
- Worktree: branch `task-board/story/STORY-260728-1ojb1p`, HEAD `bab2433ba115a7eafb2298dba6ef16154e2cc62e` = review base; HEAD is an ancestor of `main` (pure uncommitted delta, no local commits past checkpoint — no `change_request_candidate_committed_past_checkpoint` risk).
- Uncommitted delta = 11 paths, matching the rev2 patch resource description (`repository_delta=present, 11 changed paths`):
  - modified (8): `.github/ci/platform-cases.tsv`, `README.md`, `internal/buildrepo/adopt_test.go`, `internal/buildrepo/protected.go`, `internal/buildrepo/protection_windows_test.go`, `internal/crossconformance/draftsources_semantic_external_test.go`, `internal/install/external.go`, `internal/install/targets.go`
  - new (3): `cmd/curator/native_blackbox_test.go` (254 lines), `docs/external-build-repositories.md` (75 lines), `internal/buildrepo/protected_artifact_test.go` (67 lines)
- `README.md:75` links `docs/external-build-repositories.md`; the guide links the spec's full rules.
- No Windows-reserved names in the delta; no CHANGELOG/LOGBOOK paths.
- No spawn directives pending for RUN-261001-27a48a.

## Evidence accepted (not rerun by this run)

Per the binding integrate instruction ("Change no file", runner lands synchronously): no tests, vet, or builds were rerun here. Relied on already-attached evidence: producer results (`TASK-260728-20ao7p_results_RUN-261001-32a6ee.md`), rev2 bounded validation log, and the reviewer verdict (gate green on every lane incl. Windows).

## Handoff to runner

Ready for the bound synchronous landing of CR-TASK-260728-20ao7p-2 rev2 onto trunk. Only the integration transaction may write `done`.
