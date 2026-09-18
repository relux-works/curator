# Review brief — TASK-260918-gudgu3: landing review of the rebased §9.5 dotfile-manager table spec tree

Round 2 (`TASK-260918-24eazm_review-verdict-rev2.md`) ACCEPTED the candidate
(`TASK-260918-24eazm_spec-patch_rev2.patch`, patch-id `55d33749`, base
curator-spec `5146c7b`). Since then curator-spec `main` moved to `23be89e`
(STORY-260916-1ll22r: the §8.4.1 absence-versus-read-failure discipline,
PR #73) on overlapping paths, so the orchestrator rebased the accepted
candidate for landing. The board Change Request of TASK-260918-24eazm stays
accepted (empty curator delta); this round reviews the EXACT tree that will
land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/spec-land-24eazm`
  at commit `802caee548ddc8b19408746d26c7972d39b39cc2` (tree `87234cd62356f44d290e420443aba675368d610d`) = head
  of https://github.com/relux-works/curator-spec/pull/74 (branch
  `land/24eazm`, parent `23be89e`). The rebased delta is also attached as
  `24eazm-rev2-rebased-vs-23be89e.patch`. Never edit the worktree; run
  anything in a disposable byte copy under `/tmp` (repo venv
  `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`
  on PATH).

## What to verify
1. **Merge fidelity.** `git diff 23be89e..802caee548ddc8b19408746d26c7972d39b39cc2` vs the accepted patch: for
   every file, either the hunks are patch-id-identical (`git patch-id
   --stable` per file — expected for the two files the 1ll22r landing did
   not touch: the new vector file and the manifest's new entry region if
   context-only), or the file is one of the merged sources:
   - `protocol/environments.md` — §9.5 step 1: the candidate's rewritten
     heuristic paragraph ("…from the closed table below…") followed by the
     1ll22r paragraph "Inventory reads follow the section 8.4.1 …"; confirm
     nothing of either side was dropped or reworded and that the two
     paragraphs are consistent (a probe the heuristic cannot complete yields
     no signal — both say so); the table and resolution rules unchanged from
     the accepted candidate.
   - `tools/validate.py` — the candidate's DOTFILE block (constants +
     `validate_environments_dotfile_managers_vectors` and helpers, 450 lines)
     and its `main()` call inserted as whole blocks next to the landed
     READ_FAILURE block and call; both families complete; compare each block
     byte-for-byte with the candidate's added lines and with `23be89e`.
   - `tools/test_validate.py` — the whole `DotfileManagersVectorTests` class
     (232 lines) inserted next to the landed `ReadFailureVectorTests` class;
     both classes complete and unmodified (diff each against its source).
   - `CHANGELOG.md` — both Unreleased entries, the 12lbww entry first,
     both verbatim.
   - `release/1.0.0-rc.9.json` and `conformance/v1/manifest.json` — equal to
     the regenerated output (item 2).
   Quote file:line and the commands.
2. **Regeneration exactness.** In the disposable copy: `make regenerate`
   changes nothing (`git status` clean afterwards against a temporary
   baseline commit of the copy) and `make regenerate-check` exits 0.
3. **Validation.** `make validate` (`set -o pipefail`) — quote the gate
   outputs and exit code; replay the round-2 rule-7 probes of BOTH families
   (a name-preserving replacement in `environments-dotfile-managers.json`
   and one in `environments-read-failure.json`) — both refused.
4. **No semantic drift.** The accepted 12lbww rules (closed table, labels,
   resolution rule with the scoped upstream note, lstat presence, never
   blocks, direct rollout wording) and the landed 1ll22r rules are each
   stated exactly once with consistent spellings.

## Verdict
Record `TASK-260918-gudgu3_review-verdict-rev1.md` (an outcome resource on THIS task)
with the per-file merge table, transcripts and the verdict `accept-landing`
or `changes_requested` (concrete corrections). This task carries no Change
Request: record the verdict, tick the checklist items you verified, then set
this task's status yourself — `task-board m 'set_status(TASK-260918-gudgu3, status=done)'`
on accept-landing, or `status=to-dev` with the corrections on
changes_requested. Do not touch TASK-260918-24eazm (accepted; the
orchestrator lands PR #74 and closes it with `integrate_external`). Leave
NO files anywhere but your disposable copy, and remove it.
