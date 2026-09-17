# Brief — TASK-260916-y4sa6s: spec revision for the source signer allowlist and the update-delta confirmation (E1)

Story `STORY-260916-ioemse` (source-signer-allowlist-and-update-delta), wave 2
of the 2026-09 security-audit remediation. The manager task
`TASK-260916-1zgucp` implements what THIS revision specifies.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-ioemse/worktree`
(branch `task-board/story/STORY-260916-ioemse`, forked from curator-spec `main` `23dafa7`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E1 (High) and Appendix B (E1 confirmed: no
signer verification anywhere; `profile update` prints only
`updated profile <name> (lock <hash>)`; `latest` is `*`). Decision 0012 and
`protocol/environments.md` §1.3/§1.4 resolve semver ranges to the highest
`v`-tag with no signature step; the lock "is a record, not a signature";
`profile update` (§9.2) re-resolves and re-materializes in-place surfaces and
managed homes, so whoever pushes an in-range tag ships new system-prompt,
root-context and MCP bytes. Text to revise: `decisions/0012-context-packages-and-semver-locks.md`
(amendment in the style of its existing Erratum section — do not rewrite the
decision history), `protocol/environments.md` §1.3, §1.4, §9.2, §9.7
(diagnostics), §12 (status posture), §12.1/§12.2 (knobs), §13 (conformance
surfaces); `schemas/v1/manager-config-v2.schema.json` and
`system-config-v2.schema.json` for the knobs (closed, like the existing
`environments` keys).

## Settled decisions (do not reopen)
- **Signer allowlist per source, optional but lockable machine policy.** New
  §12.1 knob `source_signers.<source>` (source = canonical source identity
  exactly as the lock spells it): list of entries `{ type: "ssh", key }` (an
  OpenSSH public key line) or `{ type: "gpg", fingerprint }` (40 uppercase
  hex), default empty for every source. A second knob
  `require_source_signers` (boolean, default `false`). Both enter the §12.2
  lockable set; a locked `source_signers` list is fleet policy (a machine file
  may add signers only when the lock says nothing for that source — state the
  rule plainly), and `require_source_signers` is lockable only in the
  direction of `true`.
- **Verification happens in resolution, before a candidate enters the lock**
  (§1.4): for a `git` source that has a signer allowlist, the selected
  candidate's annotated tag signature OR the signature of the commit it peels
  to MUST verify against an allowed signer; either suffices. Failure is a
  resolution error, the operation fails closed and the old lock stays:
  unsigned → `context_source_unsigned`; signed by a key outside the allowlist
  or invalid → `context_source_signer_rejected`; both name the source, the
  tag/commit and the signer that was seen. With `require_source_signers` true,
  a `git` source without an allowlist is `context_source_signers_missing`
  (error). A `path` source is never verified (say so). Verification never
  moves a selection to a lower candidate silently.
- **Posture**: `env status` reports per lock member's source whether signer
  verification is `enforced` (allowlist present, verified signer named),
  `unconfigured`, or `required-missing`, and the machine-level
  `require_source_signers` value; §12 status text gets the row.
- **`profile update` delta (E1 recommendation 2), warn-first** (impact row
  "E1 update confirmation"): §9.2 step 3 is extended so `profile update` MUST
  print the resolved-version delta of the candidate lock against the old lock
  — per member: added / removed / moved `<from> → <to>` with pins — before the
  lock is published. When the delta introduces or changes a `class: system`
  module (a member new to the lock that carries one, or a moved member whose
  system-module set or bytes changed) or an MCP declaration (an `mcp` member
  new to the lock, or a moved one whose `command`, `args` or `env_names`
  changed): revision A warns with `profile_update_system_delta` (naming the
  members) and proceeds; revision B refuses with
  `profile_update_confirmation_required` unless the per-run flag
  `--confirm-system-delta` is given. The flag is per invocation only — no
  configuration knob may pre-confirm; `--all` needs the same flag and confirms
  every profile of the run. Specify both revisions explicitly.
- **`latest` residual named**: `latest` stays `*`; with an allowlist it follows
  every signed in-range tag of that source, without one it follows any tag —
  state this in Decision 0012's amendment and in §1.4, and cross-reference the
  strict-tag policy (moved tag ≠ new tag).
- The lock schema (`context-lock-v1`) does not change: verification results
  are posture, not lock content; "the lock is a record, not a signature"
  stays true and now says where the signature check lives.

## Deliverable
1. Decision 0012 amendment section (dated 2026-09-17, E1) with the two rules
   and the `latest` residual; §8 of the decision cross-referenced.
2. `protocol/environments.md`: §1.3 pointer, §1.4 verification step with the
   three diagnostics, §9.2 delta printing + confirmation with both rollout
   revisions and the `--confirm-system-delta` flag, §9.7 and §1.1 diagnostics
   tables as the tables are organised, §12 status rows, §12.1 knob rows,
   §12.2 lockable set, §13 conformance surface names.
3. Schemas: `manager-config-v2` / `system-config-v2` `environments` objects
   gain `source_signers` and `require_source_signers` (closed entry shapes);
   schema-cases valid/invalid registered in the manifest; the generator under
   `tools/generate-vectors/` extended if these vectors are generated (check
   the Makefile and how `manager-config-v2.json` is produced) so `make
   regenerate` is clean.
4. Vectors: a new `conformance/v1/vectors/environments-source-signers.json`
   (allowlisted ssh tag signature accepted; allowlisted gpg commit signature
   accepted; unsigned refused; wrong signer refused; no allowlist →
   unconfigured posture, accepted; `require_source_signers` + no allowlist →
   refused; locked list vs machine addition) and update-delta cases
   (no system/MCP change → no confirmation; new system module → A warning /
   B refusal; changed MCP `args` → same; with `--confirm-system-delta` →
   proceeds; `--all`), registered in `conformance/v1/manifest.json` like
   `environments-env-passthrough.json`. Existing vectors byte-identical.
5. `CHANGELOG.md` Unreleased entry "E1: …" naming the allowlist, the two
   rollout revisions of the confirmation, and the `latest` residual.

## Out of scope
Implementation (`TASK-260916-1zgucp`), registry-side provenance, proposals
0014–0018, E2 (already landed: `transitive_system_modules`), tags/releases.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-y4sa6s_spec-patch_rev1.patch` (git diff against origin/main with
new files via `git add -N`) and `TASK-260916-y4sa6s_evidence.md` (with the
`make validate` transcript), then
`task-board handoff TASK-260916-y4sa6s --role doc-writer`.
