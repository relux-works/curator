# Review brief — TASK-260918-fjl6v2 (the accepted rc.12 union re-applied on the current trunk), review round 1 (story-final Change Request revision 1)

You are the independent reviewer of a curator change produced for
`TASK-260918-fjl6v2` (story `STORY-260918-2yvd86`, `EPIC-260910-2hw1xb`).
Context: the rc.12 conformance-pin promotion union — `SPEC_PIN` →
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (tag `v1.0.0-rc.12`, qualified by
`TASK-260917-2ecpjv`) plus the E2/E4/S4 manager implementations — was
independently reviewed three times and ACCEPTED at `TASK-260917-16l2md`
revision 5 (`TASK-260917-16l2md_review-verdict-rev5.md`; candidate tree
`559e88e475d813ebe6c28e3dcaa798af028c0c74` on base `3c45d4b`, checkpoint
`73fc8a4`). The curator trunk then moved (S6 squash `64cacfc`, rustup CI
fixes `d00fe7a`/`1c464c5`, board-state commits) and the sanctioned replay
tool failed on that story, so this task re-applies the accepted delta on a
fresh fork of the trunk with three additive conflicts resolved as unions and
a test-only fix-up (prepared and validated by `TASK-260918-11f9l1`). Your
job is NOT to re-review the union's design (accepted) but to prove that this
candidate is exactly "trunk + the accepted union + the stated combination
work", and that the combination is correct.

Read: `remediation-manager-producer-rules.md`, the brief
`TASK-260918-fjl6v2_brief.md`, the results `TASK-260918-fjl6v2_results.md`,
the union delta `rc12-union-73fc8a4-vs-3c45d4b.patch` (40 files, patch-id
`ccc9574a`), the prepared resolutions/fix-up (`TASK-260918-11f9l1_*`) and
their rationale (`TASK-260918-11f9l1_results.md` §2–§5), the published
`TASK-260918-fjl6v2_change-request_rev1.patch` and its validation log, and
the rev-5 verdict of `TASK-260917-16l2md` (resource of that task).

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260918-2yvd86/worktree`
(branch `task-board/story/STORY-260918-2yvd86`, HEAD = the trunk commit the
results name, uncommitted delta). Do not edit it and leave NO files in it;
build, test and probe in a disposable copy under `/tmp` with
`CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12-review/conformance/v1`
(`git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree add /tmp/spec-rc12-review dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
if absent). Builds go to `/tmp`, never into the worktree.

## What to verify
1. **Identity, file by file.** For each of the 40 union files:
   `git diff HEAD -- <file>` in the worktree vs the union patch's hunks for
   that file — `git patch-id --stable` identical, EXCEPT `CHANGELOG.md`,
   `cmd/curator/envstatus.go`, `internal/envprofile/status.go` (must be
   byte-identical to `TASK-260918-11f9l1_resolved-*`, and each of those must
   be the union of the S6 side (`git show 64cacfc:<file>`) and the union
   side (`git show 73fc8a4:<file>`) with nothing dropped, reordered beyond
   the stated order, or reworded — diff both directions) and
   `cmd/curator/main.go` (auto-merge: equals checkpoint content + the S6
   hunks; show it). Plus the two fix-up test files
   (`cmd/curator/hook_posture_test.go`, `cmd/curator/hook_test.go`): only
   the stated stub-provider blocks, no production code. NOTHING else may
   differ from HEAD (`git status --short --untracked-files=all`).
2. **The pin.** `SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b` in
   every suite checkout of `.github/workflows/ci.yml`; `release.yml`
   untouched; the ledger/skip-class state as accepted at rev 5 (no new
   skip rows).
3. **Combination correctness.** The closed output order of `curator status`
   and `env status` stated in the results (S6 shell-hook block first, then
   the union's §12 posture block) matches what the code prints and what
   BOTH test suites expect; the fix-up's rationale holds (a missing
   session provider is non-current under the union's §12 provider rows, so
   S6's `--check`-expects-OK tests need stub providers) — confirm by
   reverting the fix-up in your disposable copy and watching exactly those
   S6 tests fail, then restoring it; no production behaviour changed by the
   combination.
4. **Gates.** In the disposable copy: `go build ./... && go vet ./... && gofmt -l .`
   clean; the full suite at the rc.12 root green (quote the command and
   exit code; the hosted gate log for revision 1 read); at least one
   narrowing mutant per resolved Go file (e.g. drop the S6 rows from
   `envstatus.go`'s printer; drop a union field from `status.go`) caught by
   committed tests.
5. **Hygiene.** No build outputs, `.review/`, LOGBOOK or coverage files in
   the candidate; nothing left in the worktree by you.

## Verdict
Record `TASK-260918-fjl6v2_review-verdict-rev1.md` (task outcome) with the
per-file identity table, the union three-way evidence, transcripts, mutants
and findings; then exactly one of
`task-board m 'accept_cr(TASK-260918-fjl6v2, revision=1, evidence=TASK-260918-fjl6v2_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260918-fjl6v2, status=to-dev)`
naming the exact discrepancy. Never accept on the producer's or the
orchestrator's word alone.
