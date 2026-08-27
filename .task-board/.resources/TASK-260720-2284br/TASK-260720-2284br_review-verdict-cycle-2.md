# TASK-260720-2284br review verdict — cycle 2

## Verdict

Changes requested. Route to `to-dev`.

The cycle-1 symlink and stale-removal defects are fixed, and the submitted
atomicity/race evidence is credible. A separate acceptance blocker remains:
the machine-wide hybrid activation input can change after planning without
triggering revalidation, so the commit may apply a stale closure.

## R3 — Hybrid activation is not observed or revalidated

Severity: high; acceptance-blocking.

Evidence:

- Project planning reads `hybrid/Skillfile.json` and merges applicable hybrid
  declarations into the effective closure before the commit lock
  (`internal/install/install.go:220-255`).
- The optimistic observation set records only installed marker paths and cache
  outcomes (`internal/install/install.go:411-424,449-450`;
  `internal/install/commit.go:213-283`).
- `runCommit` claims it revalidates stale activation, but its call receives only
  those marker paths and cache outcomes (`internal/install/commit.go:381-384`).
  `scopes.HybridManifestPath(home)` is never observed.
- The stale `nodes` and `hybridNames` captured before private staging are then
  passed unchanged into home-locked target staging
  (`internal/install/install.go:483-499`).

Independent reproduction:

1. A scratch Go overlay test declares `skill-h` for the project through the
   hybrid manifest.
2. `Options.OnStaged` removes that declaration after closure/build staging and
   before `runCommit`.
3. The expected contract is a closure restart followed by an install without
   `skill-h`.
4. Actual result: the install returns `ok`, reports no restart, and commits the
   stale hybrid context.

The failing log is
`.temp/TASK-260720-2284br/review/stale-activation.log`; the read-only overlay
test is
`.temp/TASK-260720-2284br/review/stale_activation_test.go`. The implementation
worktree was not modified.

Impact:

- A hybrid declaration removed or retargeted during private staging can still
  install its old context and adapter mirror.
- A newly applicable hybrid declaration can be omitted from the committed
  closure.
- This violates the explicit criterion that stale closure/activation state
  restarts instead of applying the old plan.

Required rework:

1. Add every activation/closure input consulted outside the home lock to the
   optimistic observation set, at minimum
   `scopes.HybridManifestPath(cfg.Home())`.
2. Recheck it under the home lock before cache publication or target staging;
   a mismatch must return `restartClosure`.
3. Add a mutation-injection regression at the pre-home-lock boundary proving a
   removed/retargeted hybrid declaration causes a restart and that no stale
   hybrid context or adapter mirror commits.
4. Audit the remaining closure inputs for the same omission and either observe
   them or document the lock/frozen-snapshot invariant that makes them stable.

## Cycle-1 findings

R1 is closed. `transaction.KindEntry` owns a symlink as a directory entry,
digests its exact destination string, and uses the same kind for staging,
backup, rollback, cleanup, and namespace validation. Adapter entries are now
journaled class-60 targets and stale adapter entries are class-80 removals;
there is no direct mutation between `Prepare` and `Commit`.

R2 is closed. The acceptance package sweeps project+hybrid and global
scenarios, enforces that each named class is actually produced, faults at
`PointAfterBackup`, and covers auto/explicit-symlink link restoration and
stale-removal ordering. The submitted full and broad race logs exit 0.

## Independent verification

Current implementation tree:

- `git diff --check` — passed.
- `gofmt -l .` — no output.
- `go build ./...` — passed.
- `go vet ./...` — passed.
- `go test ./internal/transaction ./internal/staging ./internal/adapters -count=1` — passed.
- Focused exact-link and stale-removal atomicity tests — passed.
- Focused install lifecycle/concurrency/recovery tests — passed.
- Focused race run over transaction/staging/adapters — passed.
- Tests for the six lint-expanded packages — passed.
- `golangci-lint` v2.4.0, repo-wide — `0 issues`.
- Stale hybrid activation overlay regression — failed as described above.

Submitted archive:

- Full `go test ./... -count=1` — exit 0.
- Broad race run over install/atomicity/transaction/managerlock/staging/adapters
  — exit 0.
- Rework task-scope lint — 0 issues.

No product code was modified during review.
