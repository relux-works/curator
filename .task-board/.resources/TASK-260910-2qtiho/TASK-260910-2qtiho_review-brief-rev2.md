# Review brief — TASK-260910-2qtiho (curator-spec revision, S1+S3 hardened-defaults profile and advisory residual), review round 2

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260910-2qtiho` (story `STORY-260910-2qmrb8`, wave 3 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260910-2qtiho_brief.md`, the producer's evidence
`TASK-260910-2qtiho_evidence.md` and patch `TASK-260910-2qtiho_spec-patch_rev2.patch`
(all task resources), and the finding in
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/docs/security-audit-2026-09.md`.

## Where the candidate is
The candidate tree is the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2qmrb8/worktree`
(branch `task-board/story/STORY-260910-2qmrb8`, forked from curator-spec `main` `e8b53a0`).
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
Record exactly one verdict resource `TASK-260910-2qtiho_review-verdict-rev2.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260910-2qtiho, revision=2, evidence=TASK-260910-2qtiho_review-verdict-rev2.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-2qtiho, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (S1+S3, revision 2)
Compare the patch against the worktree with `git diff HEAD` (base `e8b53a0`),
not `origin/main` (which moved to `4a2fa3e` with E3). Check in particular:
one closed knob `security_posture` (`permissive` | `hardened`) in
`manager-config-v2` next to `audit`, lockable in `system-config-v2` only
towards `hardened`; the hardened effective defaults exactly as the brief
lists (audit strict, registry strict, non-empty `allowed_sources` →
`source_allowlist_empty`, non-empty MCP allowlist for MCP-carrying profiles →
`mcp_package_allowlist_empty` as an error, `passable_env_names: null` →
`passable_env_names_unbounded_refused`, `transitive_system_modules: error`,
`require_source_signers: true`, unreachable trusted registry → error) with
the explicit-value-wins rule and its three exceptions; warn-first as two
labelled revisions (A: knob admitted, default `permissive`, posture reported,
`security_posture_permissive` warning once per operation with the hint; B:
default flips — a later release); the `security_posture` status row with the
closed per-gate vocabulary (every landed gate listed) and `--check`
semantics; the S3 residual paragraph in registry §4 and SECURITY.md (named,
not changed) and the `registry_unreachable_during_install` gate notice
(warning under permissive, error under hardened) naming the artifacts
resolved without registry evidence; no existing gate default changed; the
new `security-posture.json` vector family with a scenario-pinning validator
gate (rule 7 — probe with a replacement); pre-existing vectors
byte-identical; EMPTY curator delta.

## Round 2 specifics
Revision 1 was rejected with F1 (the validator did not pin the effective
posture/precedence per scenario: the MCP refusal case rewritten under
`permissive` passed under its negative name; two more substitutions
accepted out of 272) and F2 (`curator status` closed to four rows, `env
status` to eight — the twelve-gate inventory missing on both). Verify: replay
the reviewer's three substitutions (its probe is a task outcome) through
`validate.main()` — all MUST be refused — and run your own exhaustive
same-name substitution sweep; every required scenario is pinned to schema
version, effective posture, precedence conditions and branch inputs; both
commands carry the complete twelve-gate inventory with one shared closed
vocabulary and provenance values when the environments capability exists,
the schema-1/no-environments case explicitly bounded, both outputs pinned in
vectors; everything else byte-identical to revision 1 (compare the
spec-patch resources); standing items (patch = `git diff HEAD` on base
`e8b53a0`; `main` is now `1ca4b3d` — not the comparison base; byte-identity
of untouched vectors; make validate; regeneration proof; EMPTY curator
delta).
