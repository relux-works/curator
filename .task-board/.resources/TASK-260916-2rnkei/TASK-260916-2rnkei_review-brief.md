# Review brief — TASK-260916-2rnkei (curator-spec revision, E3 codex seed mcp_servers, warn-first), review round 1

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260916-2rnkei` (story `STORY-260916-1i1gfo`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260916-2rnkei_brief.md`, the producer's evidence
`TASK-260916-2rnkei_evidence.md` and patch `TASK-260916-2rnkei_spec-patch_rev1.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1i1gfo/worktree`
(branch `task-board/story/STORY-260916-1i1gfo`, forked from curator-spec `main` `684c9f1`).
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
Record exactly one verdict resource `TASK-260916-2rnkei_review-verdict-rev1.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260916-2rnkei, revision=<N>, evidence=TASK-260916-2rnkei_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-2rnkei, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (E3, revision 1)
Compare the patch against the worktree with `git diff HEAD` (base `684c9f1`),
not `origin/main`. Check in particular: warn-first as two labelled revisions
(A: seed copied whole, `mcp_native_servers_ungoverned` names every inherited
native `mcp_servers` entry with the migration hint, `env status` lists them as
ungoverned; B: `mcp_servers` and every `mcp_servers.*` sub-table stripped,
`mcp_native_servers_not_inherited` names them once, posture lists them as
not inherited); the §7.4 seeds row honest about evidence confidence; the
§7.8 per-adapter asymmetry row/table (claude `--strict-mcp-config`, codex
`-p` over the seeded base, opencode merge-order residual, pi none); existing
homes keep their bytes and the marker seed record carries the seed rule
revision (`mcp_seed_unstripped` posture with the re-provision hint) — if the
marker schema had to change, the frozen-schema versioning rule was followed;
closed spellings identical in §7.7, §12, vectors and CHANGELOG; the seed
snapshot records server NAMES only; vectors cover both revisions, the
no-`mcp_servers` case and the pre-rule home; the validator pins scenarios
(producer rule 7 — probe with a replacement); pre-existing vectors
byte-identical; the curator repository delta is EMPTY (no LOGBOOK.md
change). Two other spec revisions (E5, S5) are in flight on the same base;
not part of this candidate.
