# Rebase brief — TASK-260916-1x0ogh, revision 3 (rebase the accepted rev2 candidate onto curator-spec main f544a01)

Review round 2 (`TASK-260916-1x0ogh_review-verdict-rev2.md`) ACCEPTED your rev2
candidate. Since then curator-spec `main` advanced to `f544a01` (S6 #60 and E2
#61 landed) and a three-way apply of the accepted patch conflicts in 30 paths,
including `protocol/environments.md`, `profiles/manager.md`,
`schemas/v1/system-config-v2.schema.json`, `tools/generate-vectors/*.go`,
`tools/validate.py`, `tools/test_validate.py`, the manifest, release pins and
~20 regenerated `system-config-v2` schema cases. Your job in this round is
ONLY to rebase; the product content of rev2 does not change. Rules in
`remediation-spec-producer-rules.md`, worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2otjbn/worktree`.

## Steps
1. In the worktree: `git add -A . && git commit -m "WIP: TASK-260916-1x0ogh rev2 accepted candidate"`
   (a scratch commit on the story branch `task-board/story/STORY-260916-2otjbn`;
   the orchestrator makes the signed landing commit later — do not push).
2. `git fetch origin && git rebase origin/main`. Resolve every conflict by
   KEEPING the landed S6/E2 content AND your E4 additions: table rows for all
   knobs (`transitive_system_modules`, `system_module_waivers`,
   `provider_directories`), both diagnostics sets, both lockable-set entries
   (`environments.transitive_system_modules` error-only, and
   `provider_directories` as your rev2 decided), both schema properties, both
   generator fixture sets and both validator consumers/tests. Never drop a
   landed line; never re-open the E2/S6 decisions.
3. Regenerate: `make regenerate`; then `make regenerate-check` MUST exit 0.
4. `make validate` (repo venv on PATH, `set -o pipefail`) MUST exit 0 — quote
   the three gate outputs; if a merged test or consumer fails, fix the merge
   (not the product rule) and say what was wrong.
5. Attach `TASK-260916-1x0ogh_spec-patch_rev3.patch` = `git diff origin/main`
   of the rebased worktree (include new files), and update
   `TASK-260916-1x0ogh_evidence.md` with a "Rev3 = rev2 rebased onto f544a01"
   section: the conflicted paths, how each was resolved, and the transcripts.
   Tick the checklist items you satisfy, then
   `task-board handoff TASK-260916-1x0ogh --role doc-writer`.
