# Review brief — TASK-260910-1b1ens (curator-spec revision, R1/P1 records/log page boundary), review round 1

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-1b1ens` (story `STORY-260910-25yc0h`, wave 2 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-1b1ens_brief.md`, the producer's evidence
`TASK-260910-1b1ens_evidence.md` and patch `TASK-260910-1b1ens_spec-patch_rev1.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-25yc0h/worktree`
(branch `task-board/story/STORY-260910-25yc0h`, forked from curator-spec `main` `23dafa7`).
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
Record exactly one verdict resource `TASK-260910-1b1ens_review-verdict-rev1.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-1b1ens, revision=<N>, evidence=TASK-260910-1b1ens_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-1b1ens, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (R1/P1, revision 1)
Check in particular: the `boundary` field is REQUIRED in both v2 envelopes and
is the full `registry-snapshot-v1` object (incl. `sig`); v1 schema files are
byte-identical to `main`; the §9 endpoint table names the v2 schemas; the
client rule orders signature verification before the §5 rollback comparison
and names exactly `registry_page_boundary_stale`, `registry_page_boundary_mismatch`,
`registry_page_boundary_missing` with the exclusion / contributes-no-record
consequence; the missing-boundary case has no legacy-accept mode and no knob;
the cursor/boundary agreement rule is on the service profile with the `404
invalid_cursor` refusal; the new `registry-client.json` and
`registry-service.json` cases cover every branch listed in the brief and the
pre-existing cases are byte-identical (`git diff` on those files shows only
additions); the posture row exists; CHANGELOG names R1/P1 and both stories.
