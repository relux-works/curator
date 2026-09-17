# Review brief — TASK-260916-y4sa6s (curator-spec revision, E1 source signer allowlist and update-delta confirmation), review round 1

You are the independent reviewer of a curator-spec normative revision produced
for `TASK-260916-y4sa6s` (story `STORY-260916-ioemse`, wave 2 of the 2026-09 security-audit
remediation). Read, in this order: `remediation-spec-producer-rules.md`, the
producer brief `TASK-260916-y4sa6s_brief.md`, the producer's evidence
`TASK-260916-y4sa6s_evidence.md` and patch `TASK-260916-y4sa6s_spec-patch_rev1.patch`
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
Record exactly one verdict resource `TASK-260916-y4sa6s_review-verdict-rev1.md`
(per-item table with quotes and file:line, validation transcript, findings)
and then either
`task-board m 'accept_cr(TASK-260916-y4sa6s, revision=<N>, evidence=TASK-260916-y4sa6s_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-y4sa6s, status=to-dev)`,
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.

## Round specifics (E1, revision 1)
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

## Orchestrator note on the producer's reported gaps (evidence "Findings for the board")
The orchestrator has decided on them; fold them into your verdict as required
corrections in addition to anything you find yourself, so the producer gets one
consolidated rework:
1. **Widen the MCP trigger.** The confirmation trigger for a moved `mcp`
   member is any change of the declaration, not only `command`/`args`/
   `env_names`: specify it as a difference in the declaration's canonical
   bytes (CCJ-1 of the declaration object as the lock/materialization reads
   it — `url` of an `http` declaration and the `environments` selector
   included), with a vector case for an `http` `url` change and one for a
   selector change, and the gate recomputing that rule.
2. **`profile install` reinstall path takes the same flag.** Where reinstall
   re-resolves "exactly as `profile update`", `profile install` MUST accept
   `--confirm-system-delta` with identical per-run semantics; the CLI rows
   (`cli/curator.md`) and §9.2 say so; a vector or CLI row pins it.
3. Gap 2 (posture names the verified signer only when local source state
   reproduces the verification; `unknown` otherwise) is accepted as
   specified — no change requested.
Regenerated existing `manager-config-v2` / `system-config-v2` schema cases
(base config now carrying the new knobs) follow the E2/E4 precedent on
`main` and are not a byte-identity finding; check instead that every
regenerated case still tests what its name says.
