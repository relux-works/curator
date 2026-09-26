# TASK-260915-s4qkmr — Managed publication: STOPPED with evidence (no write performed)

Verdict: the product's own command CANNOT adopt `~/.local/bin/task-board` — it
refuses unmanaged conflicts by design — and running it anyway would stage and
commit machine-wide global state (canonical shims under `~/.curator/global/bin`,
adapter mirrors, env files) under the manager-home mutation lock on a live
machine, for zero progress. Per brief §3 stop rule and DoD item 3 parenthetical,
this step stops here with evidence. No marker edit, no shim rewrite, no install
was performed. Rollback: not applicable (nothing written).

## 1. The product's own command (identified, not run for mutation)

- Surface enumerated from `curator --help` (exit 0) and `cmd/curator/main.go:73`:
  `global <subcommand>` = init | add | remove | list | status | install | update | upgrade.
- There is NO adopt/refresh-shims subcommand. The only path that writes the
  user-bin ledger (`.curator-managed.json`) is
  `curator global install` → `install.Global` → `stageGlobalTargets`
  (`internal/install/global.go:437`) → `globalbins.StageForwarding`
  (`internal/globalbins/stage.go:35`). (`global add` ends in the same
  `runGlobalInstall`; `update`/`upgrade` fetch then install.)
- Read-only probes used instead (all exit 0, zero writes):
  - `curator global status` — all three builds current, cache-hit (full output
    in run transcript / `/tmp/s4qkmr-global-status.txt` during the run):
    `project-management up-to-date`, task-board/task-board-tui/tb-sessiond
    `cache-hit state=current`, expected commands [task-board, task-board-tui, tb-sessiond].
  - `curator global install --dry-run` (exit 0, "dry-run; no files modified"):
    builds cache-hit, `project-management branch main ef6ee16 ... (planned)`.
    Caveat (verified in source, `internal/install/global.go:302-309`): dry-run
    returns BEFORE `stageGlobalTargets`, so it never derives the forwarding
    plan and cannot preview the shim refusal. The refusal below is proven by
    source + byte comparison instead, which is the stronger evidence.

## 2. Why the product refuses this shim (three independent proofs)

P1 — Source (`internal/globalbins/globalbins.go:330-345`, shared by `Refresh`
and `StageForwarding` via `stage.go:103`): a target that exists on disk and is
NOT (listed in the ledger AND byte-equal to the canonical template) is an
unmanaged conflict → skipped with message `global: command %q was not published
to %s; target exists and is not managed by Curator`, and NOT added to
`nextManaged`, so the rewritten ledger still excludes it. Design comment
(curator repo, `ownedTarget`): "A ledger entry is necessary but not
sufficient. This avoids adopting a matching shim that a user created."

P2 — Byte comparison (zsh, `cmp`, exit 0): `~/.local/bin/task-board-tui` and
`~/.local/bin/tb-sessiond` are BYTE-IDENTICAL to the canonical template
`UnixShimContent` (`internal/runtimestore/runtimestore.go:161-172`:
`#!/bin/sh\nexec '<canonical>' "$@"\n`). `~/.local/bin/task-board` DIFFERS —
hand-written unquoted form (`exec /Users/...` vs `exec '/Users/...'`). It is
additionally absent from the ledger. Both refusal preconditions hold.

P3 — Ledger semantics (`stage.go:98-108,139-155`): conflicts never enter
`nextManaged`; the staged ledger would be `{task-board-tui, tb-sessiond}` —
byte-identical to the current marker. A real install provably leaves the
marker unchanged while reporting the refusal.

## 3. Why the command was not run anyway (what it WOULD do)

A non-dry `curator global install` on this host would, under the manager-home
mutation lock (`StageForwarding` doc: "The caller runs it while holding the
manager-home mutation lock"):

1. `stageRuntimeAndShims` — stage canonical-shim replacements for all three
   commands under `~/.curator/global/bin` (`global.go:401-410`). `StageShimTransition`
   (`internal/runtimestore/targets.go:324`) stages EVERY desired shim as a
   `Desired` replacement with no identical-content skip at plan time; the
   transaction then commits them over the live paths. Brief hard rule 3
   forbids editing `~/.curator/global/bin/*` in place — unconditionally.
2. `stageStaleSkillRemovals`, `envfiles.StageGlobal`, `adapters.StageGlobal`
   (`global.go:431-463`) — stale-skill removals, env-file and agent-adapter
   mirror writes: machine-wide state beyond this task's scope, while 7
   task-board/tb-sessiond processes are live (2 tb-sessiond daemons, 1
   `task-board codex` host, 4 campaign `task-board-main-6cb09a23-curatorlike`
   processes including this run's own spawn chain).
3. The forwarding step would refuse exactly the file this task needs adopted
   (see §2) and report one refusal line. Net effect on the goal: zero.

All risk, zero progress: the correct outcome is to stop. A hand-edit of the
marker was also rejected: the product HAS an install command (so the brief's
"unless the product has no such command" escape does not apply), and a
hand-added entry for non-canonical bytes would be a FALSE ledger — Curator's
own `ownedTarget` check would keep refusing it on every future install.

## 4. Residual + quiet-window adoption plan (for the operator, not run live)

- Residual: `task-board` stays shimmed-but-unmanaged; the 10 legacy standalone
  binaries stay on disk (262 MB + manifest in backup dir; ~270 MB live).
- `task-board-main-6cb09a23-curatorlike`: UNTOUCHED (sha256 `18a02361…` verified
  identical before/after in the verification resource). Retirement condition:
  campaign finished AND no process has resolved it for N days (operator picks
  N; suggest 7), then re-backup and remove in a quiet window. Never `rm` while
  any `ps`/`lsof` hit exists.
- Adoption plan (quiet window only, each step with before/after hash):
  1. `printf '#!/bin/sh\nexec \'/Users/administrator/.curator/global/bin/task-board\' "$@"\n'`
     staged to temp, `cmp` against template, atomic `mv` over
     `~/.local/bin/task-board` (or delete it and let install re-create it);
  2. `curator global install` (real run) — now adopts without conflict;
  3. verify ledger lists all three + `zsh -lc 'command -v task-board'`.
