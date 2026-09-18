# Brief — TASK-260918-2mglq0 (revision 4 of the S1 + S3 specification, sibling of TASK-260910-2qtiho): rebase onto `1ca4b3d` and absorb the E3 gate

Revision 3 was ACCEPTED on base `e8b53a0` (`TASK-260910-2qtiho_review-verdict-rev3.md`).
It cannot land as is: curator-spec `main` moved to `4a2fa3e` (E3, codex-seed
revision) and `1ca4b3d` (E6, `path`-kind admission) after your base, and the
landing conflicts are not purely mechanical.

## What the orchestrator already did (do not redo)
The Story worktree `curator-spec/.temp/STORY-260910-2qmrb8/worktree` now sits
on `1ca4b3d` (= `origin/main`) with your revision-3 delta re-applied as an
uncommitted change; the five conflicting hunks were resolved as the union of
both sides (environments §2.2 diagnostic table — E6's
`mcp_declaration_path_source_refused` row kept next to your two
`mcp_package_allowlist_empty` rows; §12 status paragraph — E6's `path`
source-directory clause and E3's codex-seed / per-home rows kept, your
`security_posture` header and twelve rows appended; §12 warnings list —
E3's three native-server warnings kept; §13 vector enumeration — E3, E6 and
your `security-posture.json` families all listed; CHANGELOG — your S1/S3
entry first, then main's), and `manifest.json` / `release/1.0.0-rc.9.json`
regenerated. The rebased delta is attached as
`TASK-260910-2qtiho_spec-patch_rev3-rebased.patch` (patch = `git diff HEAD`
with new files intent-to-added); `make regenerate-check` and `make validate`
pass on it. Start from that worktree state; verify it first
(`git -C <worktree> diff HEAD | git patch-id --stable` equals the attached
rebased patch's id).

## The substantive correction
E3 landed a shipped-revision status row that your inventory predates:
environments §12 "The codex-seed row reports `A` when the manager ships
revision A of section 7.4 … and `B` when it ships revision B … The row is
informative and never makes a row non-current; no configuration knob
selects the revision." That is a gate of exactly the `update-confirmation`
kind. The inventory rule of your revision ("every landed gate is listed;
closed to exactly these N, in this order") therefore requires:
1. manager §10 posture table: add the `codex-seed` row — value `A` or `B`
   (environments §7.4), provenance `shipped` — at a stated fixed position
   (recommended: directly after `update-confirmation`, the other
   warn-first revision row); the count becomes thirteen everywhere the
   text says twelve (manager §10, environments §12, CHANGELOG, any vector
   or validator comment).
2. environments §12: list `codex-seed` in the enumerated rows and state,
   in the same parenthetical style you used for `source-signers` and
   `store-boundary`, that the per-home `codex_seed_record` native-server
   rows above are NOT posture rows and stay unchanged — so "No other
   `env status` posture row exists" is true in the union.
3. `vectors/security-posture.json`: every pinned `curator status` / `env
   status` output gains the `codex-seed` row in the fixed position (both
   revisions where the vectors pin a shipped revision — follow how you pin
   `update-confirmation`); validator expected rows and tests follow; a
   negative that omits or misplaces the row is refused (rule 7).
4. E6 check: E6 added no revision gate (always-on `path`-kind admission +
   the store-boundary extension to `path` source directories). Confirm the
   `store-boundary` row's reference (`enforced`, environments §4) still
   covers E6's §2.2/§4 extension; adjust the reference only if it does not.
   No new row for E6.
Nothing else changes: outside these edits the tree must stay byte-identical
to the rebased revision-3 state.

## Validation and handoff
`make regenerate`, `make validate` and the regeneration proof (exit codes);
evidence file TASK-260918-2mglq0_evidence.md ("Revision 4" of the S1+S3 spec; reference the rev-3 evidence of TASK-260910-2qtiho) (state the base move, the union resolution
you verified, and the thirteenth row); `TASK-260918-2mglq0_spec-patch_rev1.patch`
= `git diff HEAD` of the worktree (base `1ca4b3d`, NOT `e8b53a0`) with new
files via `git add -N`; EMPTY curator delta; leave no stray files;
`task-board handoff TASK-260918-2mglq0 --role doc-writer`.

## Task identity
You work on task `TASK-260918-2mglq0` (a sibling of the accepted `TASK-260910-2qtiho`
under `STORY-260910-2qmrb8`; the accepted revision cannot be reopened, so
its rebase and this correction are tracked here). Read `TASK-260910-2qtiho`'s
brief, rules (`remediation-spec-producer-rules.md`), evidence and rev-3
verdict for context. Outcome artifacts carry the `TASK-260918-2mglq0_` prefix:
`TASK-260918-2mglq0_spec-patch_rev1.patch` (= `git diff HEAD` of the worktree on base
`1ca4b3d`) and `TASK-260918-2mglq0_evidence.md`. Tick the checklist of TASK-260918-2mglq0, then
`task-board handoff TASK-260918-2mglq0 --role doc-writer`.
