# Review brief — TASK-260918-2mglq0 (revision 4 of the S1+S3 hardened-defaults specification: rebase + codex-seed gate), review round 1

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260918-2mglq0` (story `STORY-260910-2qmrb8`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260918-2mglq0_brief.md`, the producer's evidence
`TASK-260918-2mglq0_evidence.md` and patch `TASK-260918-2mglq0_spec-patch_rev1.patch`
(all task resources), the accepted revision-3 verdict of the parent leaf
`TASK-260910-2qtiho_review-verdict-rev3.md` (a resource of `TASK-260910-2qtiho`;
this task is its revision 4 under a sibling id because an accepted CR cannot
be reopened), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2qmrb8/worktree`
(branch `task-board/story/STORY-260910-2qmrb8`, now on curator-spec `1e73c03` = current `main`).
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
Record exactly one verdict resource `TASK-260918-2mglq0_review-verdict-rev1.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260918-2mglq0, revision=1, evidence=TASK-260918-2mglq0_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260918-2mglq0, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.


## Round specifics (TASK-260918-2mglq0: S1+S3 revision 4)
Two layers, both to verify:

**A. What the producer did (base `1ca4b3d`).** The producer's patch
`TASK-260918-2mglq0_spec-patch_rev1.patch` is `git diff HEAD` on base `1ca4b3d` (E6). It
must equal the orchestrator-rebased revision-3 state
(`TASK-260910-2qtiho_spec-patch_rev3-rebased.patch`, patch-id `43f07dee`)
PLUS exactly the codex-seed absorption: manager §10 table gains the
`codex-seed` row (value `A` or `B`, environments §7.4, provenance
`shipped`) at one stated fixed position (the producer chose directly after
`update-confirmation`); "twelve" becomes "thirteen" everywhere (manager §10,
environments §12, CHANGELOG, vectors, validator, tests, comments);
environments §12 lists `codex-seed` and states in the existing
parenthetical style that the per-home `codex_seed_record` native-server
rows are NOT posture rows; `vectors/security-posture.json` pins the row in
every `curator status` / `env status` output; the validator's expected rows
and tests follow, with a rule-7 negative that omits or misplaces the row
refused (probe it yourself through `validate.main()`; also probe a
replacement of the row's value with a non-closed spelling); E6 needed no
row (confirm the `store-boundary` reference). Anything else that differs
from the rebased revision-3 state is a finding.

**B. The landing rebase (base `1e73c03`).** After the producer handed off,
curator-spec `main` moved again (S2, `1e73c03`). The orchestrator rebased
the Story worktree onto `1e73c03`; the worktree you review is that state,
and `TASK-260918-2mglq0_spec-patch_rev1-rebased.patch` (attached) is its `git diff HEAD`
(so "patch = worktree" in item 1 is checked against the REBASED patch and
base `1e73c03`). Verify the rebase is a faithful union of the producer's
patch and S2's landing: per-file patch-id identical for every file except
the five whose hunks S2's landing displaced — `profiles/manager.md` §1 (S2's registry-entry
paragraph followed by S1's `security_posture` paragraph),
`schemas/v1/manager-config-v2.schema.json` (one `description` sentence
carrying both extensions), `tools/generate-vectors/manager_config.go`
(S2's five registry cases followed by S1's four posture cases) and the
two regenerated files `manifest.json` and `release/1.0.0-rc.9.json`
(`schema-cases/index.json` regenerated to identical hunks). For those five, diff the rebased hunks against
both sources and confirm nothing was dropped, reordered beyond the
stated order, or reworded. `make regenerate-check` and `make validate`
must pass on the worktree as it stands.

Standing items: byte-identity of vectors untouched by this revision (S2's
and E6's vector files are `main` content, not candidate content); no
stray files; EMPTY curator delta; leave NO files in the worktree.
