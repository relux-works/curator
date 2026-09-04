# Rework 1: LOGBOOK CR delta repaired (TASK-260901-2pho68)

## What changed (exactly one heading restored)

Review round 1 (`TASK-260901-2pho68_review-findings-env-manager-1.md`, finding 1, blocking)
found that the CR rev-1 delta on `LOGBOOK.md` deleted the heading

    ## 2026-08-28 — Authoring CLI commands documentation refresh (TASK-260827-21xw9d)

orphaning that entry's three bullets under the new 2026-09-01 entry.

Fix applied in the Story worktree `.temp/STORY-260901-2rrbff/worktree`, `LOGBOOK.md` only:
re-inserted that heading line plus one blank line directly above the
`- **Documentation created**:` bullet. The new 2026-09-01 entry is intact on top.
No other file and no other line touched. The curator-spec work at `6697c1e`
(worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-env-manager`) was not touched.

## Commit

- `3f9e9e9d` "Restore the 2026-08-28 authoring-CLI logbook heading" on
  `task-board/story/STORY-260901-2rrbff`, parent `0e060112`.
- SSH-signed; `git verify-commit HEAD` reports: Good "git" signature for
  oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM (exit 0).

## Candidate delta and digest (locally computed)

- Base `eb32105d` (= main tip), candidate HEAD `3f9e9e9d`, tree `2ddbfbbe`.
- Net delta vs base: `LOGBOOK.md` +6 lines, purely additive — the new 2026-09-01
  entry (5 lines + separator) with the 2026-08-28 heading now preserved below it.
- `git diff --binary main...HEAD | shasum -a 256` =
  `fcf3393215ee013a9bb710cfce772761751b6cf50b468bc810d40e63b1197411`
  (reference for comparing against the recorded rev-2 patch; the snapshotter may
  serialize the patch differently, so compare content, not only this hash).

## Revision 2 publication

Change Request revision 2 is published by the standard managed-workspace
publication machinery on this run's completion (handoff to `to-review` + this
outcome artifact), which also runs the configured story validation suite
(`git submodule update --init --recursive`, `go build ./...`, `go vet ./...`,
`go test -count=1 -timeout 30m ./...`) in the worktree and records the digests
keyed to the candidate tree. Not run manually in this session: the delta is two
markdown lines in `LOGBOOK.md`, and publication reruns the identical suite
fail-sticky before the revision can be `ready`.

Nothing pushed, nothing marked done. Per rework brief, curator-spec worktrees
untouched; no protocol/schemas/conformance/CHANGELOG changes.

## Validation run in this session (real exit codes)

Handoff fail-closed on DoD items 8-12, so the story validation suite was run
here in bounded chunks (each command standalone, output to a log file, exit
code captured directly — no pipes):

- `git submodule update --init --recursive && go build ./... && go vet ./...` — exit 0
- `go test -count=1 -timeout 30m ./cmd/...` — exit 0 (cmd/curator ok, 254.5s)
- `go test -count=1 -timeout 30m ./internal/install/...` — exit 0 (install ok 100.1s, atomicity ok 91.9s)
- `go test -count=1 -timeout 30m <all remaining packages>` — exit 0 (55 packages ok, zero failures)

Worktree clean after the suite (`git status --short` empty, HEAD still 3f9e9e9d),
so the validated tree is the candidate tree.

DoD item notes: item 11 (gates attacked, not read) — no gate/refusal/validation
code is in this delta (two markdown lines); the adversarial coverage for this
task is the round-1 review, which traced §12 claim-by-claim against
protocol/environments.md and judged the filed §9.1 gap real. Item 12 — review
round 1 did not accept; verdict evidence `TASK-260901-2pho68_review-findings-env-manager-1.md`
was attached and status routed to development per the changes-requested branch,
which is the rework this revision answers.

## Addendum (RUN-260901-455cb4): candidate squashed for CR provenance

The previous handoff's revision-2 publication failed with
`change_request_base_authority_mismatch`: the machinery requires the candidate
head to have exactly one parent equal to the selected authority `eb32105d`,
but the branch carried two commits (`0e060112` + `3f9e9e9d`).

Fix: `git reset --soft eb32105d` and one new signed commit `979fa36e`
("Record the environments manager-profile authoring in the logbook") that
combines the logbook entry and the restored heading.

Verified in this session (real exit codes, standalone commands):
- `git verify-commit HEAD` — exit 0, Good "git" signature for oparin@me.com,
  ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM.
- Parent check: `979fa36e` has the single parent `eb32105d` (= main tip).
- Tree identity: `979fa36e^{tree}` == `3f9e9e9d^{tree}` — the candidate tree is
  byte-identical to the tree the round-1 rework validated and the reviewer's
  requested content; only commit topology changed.
- `git diff eb32105d HEAD | shasum -a 256` =
  `fcf3393215ee013a9bb710cfce772761751b6cf50b468bc810d40e63b1197411` — identical
  to the digest recorded above for the rev-2 delta.
- `git status --short` empty.

Because the tree is identical to the one the full story validation suite ran
green on (see section above), that suite was not rerun in this session; the
publication machinery reruns it fail-sticky on this handoff regardless.
