# Review brief — TASK-260916-y4sa6s (curator-spec revision, E1 source signer allowlist and update-delta confirmation), review round 3

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260916-y4sa6s` (story `STORY-260916-ioemse`, wave 2 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260916-y4sa6s_brief.md`, the producer's evidence
`TASK-260916-y4sa6s_evidence.md` and patch `TASK-260916-y4sa6s_spec-patch_rev3.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`
(branch `task-board/story/STORY-260916-ioemse`, forked from curator-spec `main` `23dafa7`).
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
Record exactly one verdict resource `TASK-260916-y4sa6s_review-verdict-rev3.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260916-y4sa6s, revision=3, evidence=TASK-260916-y4sa6s_review-verdict-rev3.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-y4sa6s, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (E1, revision 3)
Check in particular: the knobs `source_signers.<source>` and
`require_source_signers` are spelled identically in §12.1, §12.2 (lockable,
`require_source_signers` only towards `true`, a locked signer list is fleet
policy), `manager-config-v2` and `system-config-v2` (closed entry shapes
`{type: ssh, key}` / `{type: gpg, fingerprint}`), schema-cases and vectors;
verification sits in §1.4 before a candidate enters the lock (tag OR commit
signature; fail-closed; `context_source_unsigned`,
`context_source_signer_rejected`, `context_source_signers_missing`; `path`
sources never verified; no silent downgrade); `context-lock-v1` is unchanged;
the posture rows are in §12; §9.2 prints the delta and carries BOTH rollout
revisions (`profile_update_system_delta` warning first,
`profile_update_confirmation_required` refusal later) with the per-run
`--confirm-system-delta` flag and no pre-confirmation knob; the Decision 0012
amendment is an added section, not a rewrite, and names the `latest`
residual; the generator (`tools/generate-vectors`) produces the vectors when
they are generated (`make regenerate-check` clean); pre-existing vectors are
byte-identical; CHANGELOG names E1 and the two revisions.



## Round 3 specifics
Revision 2 closed R2–R7; the rev-2 verdict (`TASK-260916-y4sa6s_review-verdict-rev2.md`)
kept R1 partially open: MCP snapshot fixtures could not express optional-field
absence, so a comparator padding absent `env_names`/`environments` with `[]`
survived the gate. The rev-3 rework (`TASK-260916-y4sa6s_rework-rev3.md`)
touches only the E1 vector file, its manifest/rc.9 pins, `tools/validate.py`
and `tools/test_validate.py`: snapshots are now real `agent-mcp-v1`-valid
declaration objects with presence preserved, eight absent↔present delta cases
were added, and a narrowing test normalising absent optional fields must fail.
Verify exactly that: replay your rev-2 reproduction (padding comparator) against
the published gate — it MUST now be rejected; the snapshots validate under the
real schema through the real entry; absent vs `[]` is distinguished in
`e1_mcp_changed`; the eight cases carry A hint / B refusal / flagged
acceptance; every other file is byte-identical to revision 2 (compare the
rev-2 and rev-3 spec-patch resources); patch = worktree; `make validate` and
the disposable-copy regeneration proof; the curator repository delta is EMPTY.
Do not leave files in the worktree.
