# TASK-260720-2284br review verdict — cycle 1

## Verdict

Changes requested. Route to `to-dev`.

The journaled copy-mode path passes its focused tests, but the production-default
adapter path is not atomic and the rollback sweep does not exercise all required
scope/target classes.

## R1 — Default adapter symlinks bypass the transaction

Severity: high; acceptance-blocking.

Evidence:

- Configuration defaults `adapter_mode` to `auto`
  (`internal/config/config.go:313-319`).
- `auto` selects a symlink whenever a probe in the private staging filesystem
  succeeds (`internal/adapters/stage.go:209-221`).
- A live mirror update then deletes the current adapter entry and creates the
  new symlink directly (`internal/adapters/stage.go:53-61`).
- `runCommit` performs that direct mutation after `Journal.Prepare` but before
  `Journal.Commit` (`internal/install/commit.go:444-453`).
- A link failure returns immediately without rolling the prepared journal back.
  Transaction recovery treats `PhasePrepared` as a transaction to commit
  (`internal/transaction/engine.go:300-306`), so a later recovery can apply an
  operation that already returned `failed`, without rerunning its revalidation.
- If a later journal target fails, journal rollback cannot restore the adapter
  symlink/copy that `Mirror.Link` already replaced because the mirror is absent
  from the journal.
- Stale symlink entries are pruned only after `Journal.Commit`
  (`internal/install/commit.go:455-460`,
  `internal/adapters/stage.go:65-78`). For a project install the consumer ledger
  has therefore committed before this shared target changes. This violates both
  consumer-last ordering and the requirement that stale managed removals be part
  of the atomic target set.

Impact:

- A failed install can remove or replace a pre-existing adapter entry.
- A rollback can leave a new or dangling mirror behind.
- A failed link can leave a prepared journal that a subsequent recovery commits.
- Consumer state is not the last shared install mutation when stale adapter
  entries exist.

Required rework:

1. Put symlink adapter replacements and stale removals under durable journal
   ownership, including exact preimage restoration, or use another design that
   preserves symlink behavior while providing the same crash/reverse-rollback
   guarantees.
2. Do not leave a `prepared` journal on any ordinary post-prepare failure; such
   failures must durably enter rollback while the home lock remains held.
3. Keep every adapter/mirror mutation in deterministic classes before the
   consumer ledger.

## R2 — The “every target class” rollback test covers one copy-mode project

Severity: high; acceptance-evidence gap.

Evidence:

- Both shared install fixtures force `AdapterMode: "copy"`
  (`internal/install/install_test.go:26-35`,
  `internal/install/commit_test.go:687-696`), so the production-default
  `auto`/symlink branch is never exercised by the atomicity tests.
- `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` only adds a
  second directly declared project skill (`commit_test.go:192-227`). It does not
  create a global forwarding shim/mirror ledger, a hybrid target, or a stale
  managed removal.
- The ordering assertion explicitly requires only context, runtime, canonical
  shim, env, adapter-ledger, and consumer classes
  (`commit_test.go:141-168`). It does not cover forwarding-shim,
  mirror-ledger, or removal classes.
- No other fault-injection sweep references `ClassForwardingShim`,
  `ClassMirrorLedger`, or `ClassRemoval`.

Required rework:

1. Add failure injection over project, global, and hybrid operations so every
   deterministic class is actually present and failed in turn, including
   forwarding shims, mirror ledgers, and managed removals.
2. Add default `auto` and explicit `symlink` adapter rollback/recovery tests.
   Capture symlink state with `Lstat`/`Readlink`; `transaction.DigestPath`
   intentionally rejects symlink trees and is not sufficient evidence here.
3. Add a link-boundary failure test proving the prior mirror is restored and no
   prepared journal later commits the failed operation.
4. Add a stale-adapter-removal case proving the consumer ledger is the final
   shared commit target.

## Verification

Independent reviewer commands in the implementation worktree:

- `git diff --check` — passed.
- `gofmt -l internal cmd` — no output.
- `go build ./...` — passed.
- `go vet ./...` — passed.
- Focused atomicity/concurrency/recovery/install tests — passed in 163.781s.
- `go test ./internal/adapters ./internal/globalbins ./internal/envfiles ./internal/scopes ./internal/staging -count=1` — passed.

Producer evidence archive:

- Full `go test ./... -count=1` — passed.
- Race run over install/transaction/managerlock — passed.
- Task-scoped golangci-lint — 0 issues.
- Repo-wide golangci-lint — 45 inherited issues, matching the supplied baseline
  counts.

The reviewer could not independently rerun lint because `golangci-lint` is not
on this runtime's `PATH`; this does not affect the verdict, which is based on
the atomicity defects above.

No product code was modified during review.
