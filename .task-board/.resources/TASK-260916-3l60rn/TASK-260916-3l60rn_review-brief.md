# Review brief — TASK-260916-3l60rn (curator-spec revision, E6 source-kind admission and path boundary), review round 1

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260916-3l60rn` (story `STORY-260916-wgt8vz`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260916-3l60rn_brief.md`, the producer's evidence
`TASK-260916-3l60rn_evidence.md` and patch `TASK-260916-3l60rn_spec-patch_rev1.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-wgt8vz/worktree`
(branch `task-board/story/STORY-260916-wgt8vz`, forked from curator-spec `main` `e8b53a0`).
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
Record exactly one verdict resource `TASK-260916-3l60rn_review-verdict-rev1.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260916-3l60rn, revision=<N>, evidence=TASK-260916-3l60rn_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-3l60rn, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (E6, revision 1)
Compare the patch against the worktree with `git diff HEAD` (base `e8b53a0`),
not `origin/main` (which moved to `4a2fa3e` with E3). Check in particular:
MCP declarations are admitted from `git` sources only — a `path`-kind
package carrying one is refused at resolution with
`mcp_declaration_path_source_refused` (never warned through); `class: system`
modules from a `path` source are admitted only for a directly named root or
overlay (E2 rule unchanged) AND only after the directory passes the §4
protected-boundary contract at every resolve and before materialization,
with the S5 outcomes (`environment_store_untrusted`, no fragment, posture)
and explicitly NO rebuild; every `path` overlay/onboarding import passes the
boundary contract regardless of content class; the E1 rationale (path
sources never verified) stated; Decision 0012 §5 amendment added as a dated
section, not a rewrite; rollout direct; closed spellings identical in
§1.1/§2.1/§3.1 tables, §12, §13, vectors and CHANGELOG; vectors cover the
refused MCP case, the admitted and the two untrusted (world-writable,
symlinked component) system-module cases, the transitive path case if
expressible; the validator pins scenarios (rule 7 — probe with a
replacement); pre-existing vectors byte-identical; EMPTY curator delta.
