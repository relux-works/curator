# Review brief — TASK-260910-33j1hu (curator-spec revision, R3+P2 restore-checkpoint enforcement point), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-33j1hu` (story `STORY-260910-35tbgb`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-33j1hu_brief.md`, the producer's evidence
`TASK-260910-33j1hu_evidence.md` and patch `TASK-260910-33j1hu_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-35tbgb/worktree`
(branch `task-board/story/STORY-260910-35tbgb`, forked from curator-spec `main` `dced9b8`).
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
Record exactly one verdict resource `TASK-260910-33j1hu_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-33j1hu, revision=2, evidence=TASK-260910-33j1hu_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-33j1hu, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (R3+P2, revision 2)
The worktree base is `dced9b8`; `main` has since moved to `684c9f1` (E1 —
environments-side only). Compare the patch against the worktree with
`git diff HEAD` (the base), not `origin/main`. Check in particular: the
enforcement point is the service startup comparison AFTER the §5 integrity
verification and BEFORE binding/ready, when a checkpoint is configured;
the checkpoint object is a signed `registry-snapshot-v1` verified against the
service's accepted keys; the four closed diagnostics
(`restore_below_checkpoint`, `restore_inconsistent_with_checkpoint`,
`checkpoint_signature_invalid`, `checkpoint_not_configured`) are spelled
identically in prose, table, vectors, validator and CHANGELOG; the "above
the checkpoint" case requires the live log to reproduce the checkpoint
boundary at its `log_size`; refusal = non-ready + writes disabled, no
truncation; the no-checkpoint posture; `verify-backup` remains the offline
procedure and the text says which place is normative; the frozen
`health-response-v1` schema is untouched; the `checkpoint_cases` vector
block covers every branch the brief lists and the pre-existing
`recovery_cases` are byte-identical; the validator pins the new cases
(a narrowing probe of your own); scope (no client bootstrap interchange —
that is `TASK-260910-1tvf2t`).

## Round 2 specifics
Revision 1 was rejected with F1 only (`TASK-260910-33j1hu_review-verdict-rev1.md`):
the validator required the seven `checkpoint_cases` names but did not pin
their scenarios (0/4 replacement mutants rejected). Verify the closure with
your own replacement probe through `validate.main()` (each negative case
replaced by an internally consistent passing case under its name MUST be
refused), that the positive controls still pass, that every other file is
byte-identical to revision 1 (compare the two spec-patch resources), patch =
worktree (`git diff HEAD`, base `dced9b8`), `make validate` and the
regeneration proof.
