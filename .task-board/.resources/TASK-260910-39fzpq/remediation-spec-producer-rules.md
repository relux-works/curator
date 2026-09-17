# Security-remediation campaign — rules for curator-spec producers and reviewers (2026-09-16)

These rules apply to every `spec-*` task of `EPIC-260910-2hw1xb`. The task brief
carries the product scope; this file carries where and how.

## Where things are on THIS machine
- Authoritative board: `/Users/administrator/Developer/ReluxWorks/curator/curator/.task-board`
  (`TASK_BOARD_DIR` is set for you). Never run task-board against a copy of the
  board, never edit board bytes by hand; use `task-board m …` / `task-board resource add …`.
- curator-spec control root: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
  (read-only for you). Your **Story worktree** is
  `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/<STORY-ID>/worktree`
  on branch `task-board/story/<STORY-ID>`, already checked out at curator-spec
  `main` when you start. Work ONLY there. No writes into any control root, no
  LOGBOOK.md edits, no writes under `~/.agents`, `~/.claude`, `~/.codex`,
  `~/.curator`. Do not commit, do not push, do not create branches or PRs: the
  orchestrator owns signed commits, the PR and the landing.
- The audit that motivates the task: `curator-spec/docs/security-audit-2026-09.md`
  (findings S1–S6, E1–E7, Appendix B) and `curator/docs/security-audit-2026-09.md`
  (manager side). Read the finding named in the brief before editing.

## What a spec revision must contain
1. Normative text in the section(s) the brief names (RFC 2119 keywords, closed
   lists), consistent with the surrounding style of `protocol/environments.md`
   and `profiles/manager.md`; every new diagnostic goes into the section's
   diagnostics table and, where one exists, the `§12.1` knob table with its
   default and the `§12.2` lockable set when the brief says it is lockable.
2. **Closed sets stay closed**: a new diagnostic, config key, lock key or
   marker field is admitted explicitly, with its exact spelling in every place
   the document names it (text, tables, schema, vectors, CLI rows) — never an
   open-ended "and similar".
3. Schema changes in `schemas/v1/*.schema.json` only where the brief says so;
   conformance vectors under `conformance/v1/` (positive and negative cases)
   for every new rule, registered in `conformance/v1/manifest.json` the way the
   existing cases are; `make regenerate` if the repository generates derived
   files (check the Makefile) and commit-ready output.
4. `CHANGELOG.md` → Unreleased entry naming the finding id (e.g. "S4"), the
   rule and the warn-first rollout note from the brief.
5. Warn-first: where the brief marks a change as user-visible, the text
   specifies BOTH steps — the warning release (diagnostic + migration hint,
   old behavior kept) and the flip release (new default/refusal) — as two
   explicitly labelled revisions or conformance profiles; never one release
   that flips a default and adds a refusal at once.
6. Posture: every gate added gains its `env status` (or `curator doctor`-class)
   reporting row in the text.

## Validation and evidence
- Run `make validate` from the worktree (the repository's Python venv; state
  the exact command, shell and exit code). Run any vector/schema checks the
  Makefile exposes. Quote the outputs in your evidence.
- Attach, on YOUR task: `<TASK-ID>_spec-patch_rev<N>.patch` (`git diff HEAD` of the
  worktree — its own base commit, NOT `origin/main`, which other landings move —
  including new files via `git add -N`; NOT the
  name `<TASK-ID>_change-request_rev<N>.patch`, which the runtime uses for its
  own curator-repository Change Request artifact), and
  `<TASK-ID>_evidence.md` (what changed per file, why, the validation
  transcript, what is deliberately out of scope). Tick the task checklist
  items you satisfied with `task-board m 'check_item(<TASK-ID>, item=N)'`.
- Hand off with `task-board handoff <TASK-ID> --role <your role>`; the task
  reaches `to-review`. Do not set `done`.

## Models and review
- Producers: Muse `muse-spark-1.3-contributor`; reviewers: Codex `gpt-6-astra` at
  `low`. Reviewers verify the exact worktree tree, re-run `make validate`
  independently, check every closed-set spelling, and record exactly one
  verdict resource `<TASK-ID>_review-verdict-rev<N>.md` before routing.
- Out of scope for every spec task: implementation code, tags/releases, ax,
  proposals 0014–0018, anything the brief does not name. Report a real spec
  gap you cannot resolve as a finding in the evidence, not as an edit.
