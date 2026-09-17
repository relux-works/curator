# Review brief — TASK-260916-1qfpu4 (curator-spec revision, E5 nofollow write discipline for managed surfaces), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260916-1qfpu4` (story `STORY-260916-73a5zg`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260916-1qfpu4_brief.md`, the producer's evidence
`TASK-260916-1qfpu4_evidence.md` and patch `TASK-260916-1qfpu4_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-73a5zg/worktree`
(branch `task-board/story/STORY-260916-73a5zg`, forked from curator-spec `main` `684c9f1`).
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
Record exactly one verdict resource `TASK-260916-1qfpu4_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260916-1qfpu4, revision=2, evidence=TASK-260916-1qfpu4_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-1qfpu4, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (E5, revision 2)
Compare the patch against the worktree with `git diff HEAD` (base `684c9f1`),
not `origin/main`. Check in particular: ONE normative rule (create under an
operation-private name in the same directory, then rename over the target;
`O_NOFOLLOW`-class open, `lstat`-class inspection; never follow a symlink at
the target or at any component below the managed root the manager did not
create in this operation) stated once and referenced from every place a
managed surface is written (materialization modes, takeover, repair, backups,
marker/ledger writes); a pre-existing link is either the §9.5 foreign-manager
stop or a manager-owned entry to be replaced — never opened for writing;
rollout direct; at most one new diagnostic and only for an uncovered case
(closed sets); the vectors express "target untouched" (byte-identical former
target after the operation), a symlinked parent component (refused, not
followed) and a repair case with a link planted after provisioning; the
validator pins each scenario to its inputs (rule 7 of the producer rules —
probe it with a narrowing replacement); pre-existing vectors byte-identical;
scope excludes S5/E7. Two other spec revisions (E3 `TASK-260916-2rnkei`, S5
`TASK-260910-39fzpq`) are in flight on the same base; not part of this
candidate.

## Round 2 specifics
Revision 1 was rejected with F1 (validator did not pin scenarios: 0/9
replacement rejections), F2 (backup/marker/ledger destination links were
routed through the foreign-manager logic instead of the §8.3.1 nofollow
refusal) and, added by the orchestrator, F3 (a `LOGBOOK.md` delta in the
curator Story worktree — the curator repository delta must be empty).
Verify: your own replacement probe through `validate.main()` for every
required name (internally consistent alternative case under the retained
name → refused); the new directly-symlinked backup-target case
(`environment_write_would_follow_link`, former target untouched) and the
negative test for the wrong `environment_foreign_manager_detected`
diagnostic; the existing parent-traversal case kept; the evidence states
what is vectorized vs text-only for marker/ledger destinations; the curator
CR revision 2 carries NO repository delta; everything else byte-identical to
revision 1 (compare the two spec-patch resources); standing items (patch =
`git diff HEAD` on base `684c9f1`, byte-identity of untouched vectors, make
validate, regeneration proof).
