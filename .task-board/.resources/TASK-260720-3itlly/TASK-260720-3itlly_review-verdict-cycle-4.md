# TASK-260720-3itlly reviewer verdict — stop the line, cycle 4

## Verdict

Route to `blocked`. Cycle 4 fixes the missing global MCP and registry gates and all independently run quality gates pass. The remaining issue is an explicit contract/architecture conflict, not ordinary rework.

## Constraint and evidence

The task scope says dry-run may create and remove only the toolchain probe root. The implementation instead creates a manager-owned `curator-install-private-*` root before closure resolution, creates `closure-*` beneath it, and later creates `go-probe-base-*` beneath the same root. See `internal/install/install.go`, `global.go`, `private.go`, and `builddeps.go`. The new isolated-TMPDIR tests prove exactly one top-level root, but do not prove that the only filesystem state is the toolchain probe root; they intentionally permit the closure subtree. The cycle-4 implementation note explicitly documents this deviation.

## Failed assumptions and attempts

The former separate `curator-dry-run-*` and `curator-global-dry-run-*` roots violated the requirement and were correctly rejected in cycle 3. Consolidating them under one top-level root removes the second top-level entry but does not make the closure snapshot toolchain-owned. A no-filesystem closure snapshot is incompatible with the current `skillspec`, `skillcheck`, audit, hashing, and especially `buildsource.Validate` path/token model. The toolchain probe currently runs after closure and is skipped for closures with no build commands, so its current lifecycle cannot simply host closure resolution.

## Viable alternatives

1. Amend the task contract to permit exactly one operation-private ephemeral root shared by read-only closure snapshots and toolchain probe/build state. This keeps the current implementation, has a small ownership change, and is fully covered by cleanup/persistence tests.
2. Keep the literal probe-only contract and authorize a wider redesign: move ephemeral-root ownership behind a toolchain/private-workspace interface available before closure, or refactor closure/buildsource to support an equivalent safe snapshot abstraction. This preserves the wording but expands cross-package ownership and risk.
3. Split that redesign into a prerequisite task and return this item to development only after the prerequisite contract is settled.

## Recommendation

Choose option 1 and amend the acceptance wording to: dry-run may create exactly one operation-private ephemeral root containing closure and toolchain probe state, must remove it on every exit, must run no `go list` or `go build`, and must leave persistent state unchanged. The current code can then be reviewed for acceptance without another implementation cycle.

## Exact human decision needed

Decide whether `toolchain probe root only` permits the current shared operation-private root. If yes, amend the task scope/AC accordingly. If no, authorize the cross-package ownership/interface redesign in option 2 or create the prerequisite in option 3.

## Independently verified

- `gofmt -l internal cmd` — pass, no output
- `git diff --check` — pass
- `go build ./...` — pass
- `go vet ./...` — pass
- `go test -count=1 ./internal/install/ ./internal/godriver/` — pass
- `go test -count=1 ./...` — pass across 36 packages
- `golangci-lint v2.1.6 run ./internal/install/... ./internal/godriver/...` — pass, 0 issues

No product code was modified by the reviewer.