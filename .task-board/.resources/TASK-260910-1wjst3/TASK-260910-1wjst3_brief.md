# Brief — TASK-260910-1wjst3: spec rule for the shell hook trust gate (S6 / I1)

Story `STORY-260910-2awkzu` (shell-hook-project-env-approval-gate), wave 1.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2awkzu/worktree`.
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer (technical writer) of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` S6 (spec) and `curator/docs/security-audit-2026-09.md`
S6 / I1 (manager): the manager profile's shell integration (`profiles/manager.md`,
the shell-hook section around "`.agents/env.sh` … upward search", lines ≈1055–1090)
caches a hook that sources a project-controlled `.agents/env.sh` walking up from
`$PWD` on every directory change, with no approval or digest gate anywhere in the
protocol. The implementation confirms it is live behavior.

## Settled decisions (do not reopen)
- The hook sources ONLY env files whose bytes the manager recorded (a digest
  recorded when the manager itself generated/installed the project env file) or
  the operator approved once through the manager's approval command
  (`hook-approval-command`, manager task `TASK-260910-3ungjy`). An unknown file,
  or a recorded file whose digest changed, is NOT sourced: the hook warns once
  per shell session naming the path and the approval command, and continues.
- Re-approval is required after the file changes (digest mismatch).
- Warn-first rollout (user-visible, impact table row "S6 hook gate"): revision A
  ("warning release") keeps sourcing but emits the warning and migration hint
  for unapproved/changed files; revision B ("flip release") stops sourcing them.
  Specify both explicitly as two conformance profiles / labelled revisions of
  the rule; the manager ships A first.

## Deliverable (normative text, `profiles/manager.md` + closed sets)
1. In the shell-integration section: the trust rule above with RFC 2119
   keywords; the definition of the **approval record**: a closed shape stored
   in manager state (not in the project), one record per absolute project env
   file path: `{ path, sha256, approved_by: "manager" | "operator", approved_at }`
   — name the file/location the way manager §1 names other state, keep it
   outside every profile/package/project surface, and state that the record is
   never read from package or project data. State how the digest is computed
   (sha256 over the exact bytes sourced, hex lowercase).
2. The approval command row in the manager CLI section (`hook-approval-command`
   is the working name; align the spelling with existing CLI rows, e.g.
   `curator hook approve <path>` / `curator hook approvals`): approve one file
   (records path + current digest), list approvals, revoke. Its diagnostics.
3. Diagnostics (closed, snake_case like the rest of the document), added to the
   relevant diagnostics table: unapproved file seen (`shell_hook_env_unapproved`
   or the spelling the tables already favour), digest changed
   (`shell_hook_env_changed`), plus posture reporting: `env status` /
   doctor-class output lists each known project env file with
   approved / changed / unapproved.
4. Conformance: the hook code the manager emits is a vector surface
   (§ conformance surfaces): add positive/negative vectors for approved,
   unapproved and changed files under both rollout profiles, registered in the
   conformance manifest like existing cases; a schema for the approval record
   if the repository keeps state records under `schemas/v1/` (follow the
   existing pattern; otherwise define the shape in text only and say so).
5. `CHANGELOG.md` Unreleased entry "S6: …" naming the two rollout steps.
6. Reference: cross-link the rule from the profile's security/trust summary
   if one exists; do not touch `.agents/env.ps1` semantics beyond applying the
   same rule to it (both files are sourced by the hooks — state it).

## Out of scope
Implementation (`TASK-260910-1952mz`, `TASK-260910-3ungjy`), the S4 / E4 / E2
rules (parallel stories), proposals 0014–0018.

## Checklist and handoff
Tick the task checklist items you satisfy; attach
`TASK-260910-1wjst3_change-request_rev1.patch` and `TASK-260910-1wjst3_evidence.md`
(with the `make validate` transcript), then
`task-board handoff TASK-260910-1wjst3 --role doc-writer`.
