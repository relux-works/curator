# TASK-260720-2284br review verdict — cycle 3

## Verdict

Changes requested. Route to `to-dev`.

The cycle-2 hybrid-manifest observation closes the exact R3 reproduction, and
the previously accepted R1/R2 journal ordering remains intact. The required
closure-input audit is still incomplete, however: project and global manifests
can be changed by Curator commands that do not acquire the operation locks the
audit claims stabilize them. Both scopes then commit a stale closure.

## R4 — Project and global manifests are not stabilized or revalidated

Severity: high; acceptance-blocking.

Evidence:

- `observations` says project and global manifests are protected by canonical
  project operation locks (`internal/install/commit.go:237-241`).
- The public mutation paths do not take those locks:
  `curator add/remove` calls `manifest.AddDecl` / `manifest.RemoveDecl`
  directly (`cmd/curator/main.go:432,460`), and `curator global add/remove`
  does the same (`cmd/curator/main.go:872,883`).
- Project observations contain the hybrid activation manifest plus installed
  marker generations (`internal/install/install.go:236-239,430-439`), not the
  project manifest. Global observations contain installed marker generations
  only (`internal/install/global.go:234-240`), not the global manifest.
- Both manifests are therefore read before private staging, may move before the
  manager-home commit lock, and are never rechecked before their captured
  closure is staged and committed.

Independent read-only overlay reproduction:

```text
go test -overlay .../global-overlay.json ./internal/install \
  -run '^TestReviewerStale(Project|Global)ManifestRestartsClosure$' \
  -count=1 -v
```

Both cases fail the required restart assertion. In each case, `skill-b` is
removed from the manifest in `Options.OnStaged`, after private staging and
before the home lock. Actual behavior is `Status: ok`, no closure-resolution
restart message, and the stale `skill-b` context is committed. Scratch evidence:

- `.temp/TASK-260720-2284br/review/stale_global_manifest_test.go`
- `.temp/TASK-260720-2284br/review/global-overlay.json`

Impact:

- `curator add/remove` racing a project install can install a declaration that
  was removed or omit one that was added.
- `curator global add/remove` racing a global install has the same stale-closure
  behavior.
- This violates the acceptance criterion that stale closure or activation state
  restarts instead of applying an old plan.

Required rework:

1. Make the project and global manifest stability claim true. Either observe
   and recheck their exact consulted generations before publication, or make
   every supported writer acquire the same canonical operation lock; do not
   retain the current false lock invariant.
2. Route any detected mismatch to `restartClosure` before cache publication or
   target staging.
3. Add permanent project and global regressions for removal and at least one
   opposite-direction or retargeting change, proving no stale context, shim, or
   adapter state commits.
4. Re-audit every other pre-home-lock closure input against its actual writers,
   not its directory location, and correct the `observations` documentation.

## Cycle-2 R3 and prior findings

The narrow hybrid fix works. The prior reviewer overlay now passes:
`TestReviewerStaleHybridActivationRestarts` reports a closure restart and does
not commit the removed hybrid skill.

R1/R2 remain closed by inspection and focused checks:

- `KindEntry` still journals symlink mirrors as entries.
- Adapter entries and ledgers are class 60, mirror ledgers class 70, removals
  class 80, and the consumer ledger class 90.
- `go test ./internal/transaction ./internal/staging ./internal/adapters
  -count=1` passes.

## Verification

- `git diff --check` — passed.
- `gofmt -l .` — no output.
- `go build ./...` — passed.
- Hybrid stale-activation overlay — passed.
- Project/global stale-manifest overlays — failed as described above.
- Submitted final-tree archive records exit 0 for build, vet, all 40 packages
  in three chunks, three bounded race gates, and repo-wide lint with 0 issues.

No product code was modified during review. The only local files created are
the task-scoped reviewer scratch overlay and this verdict artifact.
