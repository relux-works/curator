# Review brief — TASK-260910-2ohnjo (curator-spec revision, S4 MCP passthrough bounds and declaration surfacing), review round 3

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-2ohnjo` (story `STORY-260910-1lf0m5`, wave 1 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-2ohnjo_brief.md`, the producer's evidence
`TASK-260910-2ohnjo_evidence.md` and patch `TASK-260910-2ohnjo_spec-patch_rev3.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`
(branch `task-board/story/STORY-260910-1lf0m5`, forked from curator-spec `main` `07e2b41`).
The Change Request revision the runtime published for this task carries an
EMPTY repository delta in the curator repository — that is expected: the spec
lives in curator-spec and the reviewed artifact is the worktree plus the
attached patch. Do not edit the worktree; read only.

## What to verify (all read-only)
1. **Patch = worktree.** `git -C <worktree> add -N . && git -C <worktree> diff origin/main`
   (or `git diff main` — same base) equals the attached patch resource
   (`git patch-id --stable` on both). Any difference is a finding.
2. **Brief conformance, item by item**: every deliverable line of the producer
   brief (sections named, rule content, settled decisions honoured, diagnostics
   in the tables, §12.1/§12.2 rows, vectors + manifest registration, CHANGELOG
   entry, warn-first two-step where marked user-visible, `env status` posture
   row). Quote the exact text for each.
3. **Closed sets stay closed**: every new diagnostic / knob / lock key is
   spelled identically in prose, tables, schema, vectors and CLI rows; no
   open-ended wording ("and similar", "etc."); no frozen v1 surface changed
   beyond what the brief authorises.
4. **Validation, independently**: from the worktree run
   `PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate`
   (`set -o pipefail`, quote the three gate outputs and exit code). If the
   repository has `make regenerate-check`, run it too. A red gate is a rework
   finding whatever the producer's evidence says.
5. **Vectors actually exercise the rule**: open each new vector; confirm the
   positive and negative cases correspond to the rule's branches (e.g. approved
   / unapproved / changed; install dir / listed dir / PATH-only / published
   dir; direct / transitive-drop / transitive-error / waived; both rollout
   profiles). A vector that only restates a default is a finding.
6. **Consistency with the settled operator decisions** in the brief; a
   deviation is a finding unless the evidence names a real spec gap and
   leaves the decision intact.
7. **Scope discipline**: no implementation code, no unrelated edits, no
   proposal 0014–0018 content; `git -C <worktree> status --short` lists only
   spec, schema, vector, manifest and CHANGELOG files.

## Verdict
Record exactly one verdict resource `TASK-260910-2ohnjo_review-verdict-rev3.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-2ohnjo, revision=<N>, evidence=TASK-260910-2ohnjo_review-verdict-rev3.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2ohnjo, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round 3 focus
Round 2 (`TASK-260910-2ohnjo_review-verdict-rev2.md`) closed everything except
one medium correction: the semantic consumer split surfacing rows on spaces
and rejected valid args containing a space. Verify rev3 parses the JSON-array
columns structurally (spaces/escapes preserved, missing/reordered/extra
columns still rejected), adds a positive surfacing vector with a space in an
argument, keeps both earlier mutants failing, and refreshed the pins; diff
rev2 vs rev3 patches to confirm nothing else changed; re-run `make validate`
and `make regenerate-check`. Note: curator-spec `main` is now `0da4020`
(S6 #60, E2 #61, E4 #62 landed, overlapping generator/validator/manifest/pins
and `environments.md`/`manager.md`); the candidate is still based on
`07e2b41` — review against its base; the orchestrator rebases at landing with
a separate landing review. Accept or list the remaining concrete corrections.
