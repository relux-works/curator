# Brief — TASK-260910-3ungjy: S6 approval commands and status posture (curator manager)

Story `STORY-260910-2awkzu` (S6 shell-hook trust gate), wave 1 of the 2026-09
security-audit remediation; second and final leaf of the story. The first leaf
`TASK-260910-1952mz` (accepted, checkpointed on the Story branch) implemented
the emitted hooks, the approval state under the manager home
(`internal/hookapproval`), the `approved_by: manager` recording at install and
the §8.7 vector execution test. Build on it; do not redo it.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-2awkzu/worktree`
(branch `task-board/story/STORY-260910-2awkzu`, already carrying the
checkpointed 1952mz commit). Rules: `remediation-manager-producer-rules.md`
(attached). Role: developer.

## Spec (read it first, it is the contract)
`profiles/manager.md` §8.3 "Approval commands", §8.4 diagnostics, §8.6 "Status
posture", §8.2 record shape, §8.7 conformance binding — in the read-only
curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
(main `23dafa7`). Read `TASK-260910-1952mz_results.md` (and its rev-2..rev-6
addenda) for the state layout, the canonical path identity rule (incl. the
Windows native/MSYS spelling mapping) and the atomic publication helper you
must reuse.

## Deliverable
1. `curator hook approve <path>`: resolves the operand to the canonical
   absolute identity the hook and `hookapproval` use (same resolution: absolute,
   symlinks resolved, Windows spelling folded); fails without recording when
   the file is absent or unreadable ("unreadable is never absence": distinct
   diagnostics/exit text for the two); records `{ path, sha256, approved_by:
   "operator", approved_at }` of the CURRENT bytes, re-recording after a
   change; atomic publication through the existing helper; idempotent when the
   record already matches.
2. `curator hook approvals`: read-only listing of every record (path, sha256,
   approved_by, approved_at), stable order, never mutating state — including
   never rewriting or "repairing" a malformed record; a malformed record is
   reported as such and skipped, matching how the hooks treat it.
3. `curator hook revoke <path>`: removes the record; on a path with no record
   leaves state byte-identical and reports there was nothing to revoke
   (exit 0 unless the repository convention says otherwise — state it).
4. Status posture (§8.6): `curator status` and `curator env status` gain the
   shell-hook trust rows: each known project env file (every recorded path,
   plus the `.agents/env.sh`/`.agents/env.ps1` of the current project when
   run inside one) with state approved / unapproved
   (`shell_hook_env_unapproved`) / changed (`shell_hook_env_changed`), the
   absolute path and, for recorded files, `approved_by`; `--check` treats a
   changed file as non-current and an unapproved file as a warning row. Use
   the existing status-row and `--check` machinery of `cmd/curator`.
5. Tests: command-level tests for approve (absent, unreadable, records,
   re-records after change, symlinked/alias operand), approvals (read-only,
   malformed record tolerated), revoke (present, absent → unchanged bytes),
   status rows in both commands incl. `--check` exit semantics; the §8.7
   vector test of 1952mz must keep passing; an end-to-end test: approve →
   generated hook sources without warning; revoke → warns again.
6. Docs: the CLI reference page(s) that document `curator hook` /
   `curator status` (find them under `docs/`), and `CHANGELOG.md` Unreleased
   entry under the existing S6 item naming the three commands and the posture
   rows (the shipped profile stays `A-warning`).

## Out of scope
The hook text and trust decision logic (1952mz; report a defect there as a
finding, do not silently change behaviour), the flip to `B-enforcing`, E2/E4/S4,
`SPEC_PIN`, PowerShell-specific approval UX beyond path folding.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-3ungjy_results.md`
(per-AC file:line, validation transcripts with `CURATOR_CONFORMANCE_ROOT`
set), then `task-board handoff TASK-260910-3ungjy --role developer`; the
runtime runs the hosted gate once at handoff.
