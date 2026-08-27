# TASK-260720-3itlly reviewer verdict — changes requested, cycle 3

## Verdict

Route to `to-dev`.

Cycle 3 closes the prior late toolchain-verdict defect: `stageBuilds` now
finalizes toolchain and source trust before `OnStaged`, while deferred release
is cleanup-only. Focused and full validation are green. Two acceptance gaps
remain in the project/global integration.

## R1 — dry-run creates a closure scratch root in addition to the probe root (high)

The task contract permits dry-run to create and remove only the toolchain probe
root. Both entry points still create a separate temporary closure workspace
before closure resolution:

- `internal/install/install.go:205-225` creates `curator-dry-run-*` and passes it
  as `closure.Options.ScratchRoot`.
- `internal/install/global.go:60-80` creates `curator-global-dry-run-*` and
  passes it as `closure.Options.ScratchRoot`.
- `internal/closure/closure.go:133-135,359-408` uses that root for ephemeral
  repositories and snapshots.

This is an observable filesystem mutation outside the allowed toolchain probe
root. The current tests check persistent manager/project paths and locks, but
do not assert the complete temporary-root creation set, so they pass without
covering this requirement.

Required rework:

1. Make project and global dry-run closure resolution read-only without
   creating the separate closure scratch workspace, so the only temporary
   filesystem state created by dry-run belongs to `Toolchain.Probe`.
2. Add project and global regressions that isolate `TMPDIR`, exercise dry-run,
   and prove no `curator-dry-run-*` or `curator-global-dry-run-*` root is
   created; the toolchain probe root must be removed on success and failure.
3. Include a missing-source or other path that previously needed
   `ScratchRoot`, so the test discriminates the old behavior.

## R2 — global builds bypass MCP and registry-attestation gates (high)

`Project` runs MCP verification and registry resolution before build planning
at `internal/install/install.go:258-347`. `Global` does not run either gate:
after closure, skill validation, collision, and system checks it runs audit,
moved-tag detection, and then enters `planBuilds` at
`internal/install/global.go:88-170`. Global markers also receive nil MCP and
attestation inputs at `internal/install/global.go:223`.

Therefore a global build can establish the toolchain, inspect cache state, and
compile a miss without first proving MCP requirements or registry attestation.
That contradicts the acceptance criterion that failing MCP and registry gates
occur before compiler work and persistent installation mutation in the project
and global refactor.

Required rework:

1. Run the applicable MCP verification and registry-attestation resolution for
   global scope before `BuildDeps.resolve`, moved-tag inspection, and
   `planBuilds`.
2. Carry the resulting MCP/attestation evidence into global markers with the
   same fail-closed semantics as project scope.
3. Add global regressions with failing injected `VerifyMcp` and
   `ResolveAttest` callbacks. Each must prove zero toolchain probes/sessions,
   cache inspections, builder calls, staged handoffs, and persistent-state
   changes.

## Cycle-3 fix review

The previous high finding is resolved:

- `godriver.Session.Release` removes private state without re-verifying the
  toolchain; `Close` retains the verify-and-remove behavior for probe callers.
- `BuildSession.Release`, `BuildPlan.Release`, and `releasePlan` preserve that
  cleanup-only contract.
- `stageBuilds` calls `BuildPlan.Verify` after the final build and before
  `OnStaged` or the first live mutation.
- The new project/global and godriver regressions discriminate the old
  post-mutation failure path.

## Independent validation

- `git diff --check` — pass
- `gofmt -l internal cmd` — pass, no output
- `go build ./...` — pass
- `go vet ./...` — pass
- `go test -count=1 ./internal/install/ ./internal/godriver/` — pass
- `go test -count=1 ./...` — pass, all 36 packages
- `golangci-lint v2.1.6 run ./internal/install/... ./internal/godriver/...` —
  pass, 0 issues

No product code was modified by the reviewer.
