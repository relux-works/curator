# TASK-260729-osjeay review verdict — cycle 3

**Verdict:** changes requested  
**Route:** `analysis` (the deliverable is a read-only execution-map/research artifact)

Revision 3 fixes cycle 2 findings F1–F4: the candidate archive operand is corrected and asserted,
the fixed manifest/tree/count identity is enforced, the Make/CI root plumbing is explicit, and the
Windows candidate root and transport are no longer an option set. Independent read-only checks also
reproduced the fixed artifact identity
`d93e155edf2a4ddf2b23b353f5c411bb40223a9c4be8295973a03bc2990c7d93`, the accepted manifest
`b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`, the whole-tree identity
`e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae` over 448 files, and the
3-modified/354-untracked conformance status. No Go command, build, test, lint, fetch, install,
product edit, target-task edit, or pin mutation was performed by this review.

## Blocking findings

### F1 — native gates still accept ambient Go despite the explicit toolchain invariant

The target task's current directive says never to accept ambient `PATH` on remote hosts. Revision 3
also states that the local host resolves a goenv shim before two native launchers and therefore
requires `GOENV=off`, `GOTOOLCHAIN=local`, and an explicit launcher.

The executable plan does not enforce either statement:

- W1 uses `where go`, then bare `go env GOROOT` and `go version`.
- W6's transferred batch file runs bare `go test`.
- C3 runs bare `go mod vendor`.
- The primary local gate set in §9.2 runs bare `go`, `gofmt`, and `make`; the proposed Makefile
  defaults `GO ?= go`.

This is not theoretical on the review host: `command -v go` is
`/Users/iv/.goenv/shims/go`, `GOROOT=/Users/iv/.goenv/versions/1.25.5`, while separate launchers
also exist at `/opt/homebrew/bin/go` and `/usr/local/go/bin/go`. The map diagnoses this selection
hazard in §5.3 and then reintroduces it in every primary command.

Required correction: define one absolute, operator-approved Go executable and GOROOT per native
host. Verify that exact executable with `version` and `env GOROOT`, set `GOENV=off` and
`GOTOOLCHAIN=local`, and use the same absolute executable in vendoring, Make (`GO=<absolute>`), and
the Windows batch runner. Use the matching `$GOROOT/bin/gofmt` for the local format gate. On Windows,
transfer the approved absolute `WIN_GO_EXE`/`WIN_GOROOT` into the runner instead of resolving
`where go`; fail closed if the fixed path or identity differs.

### F2 — Linux package discovery and its drift guard can mask `go list` failure

The proposed `LINUX_PKGS` is produced by:

```make
$(shell $(GO) list ./... | grep -v -x '$(GODRIVER_PKG)')
```

and both `linux-package-guard` probes also pipe `go list` directly into `grep`/`awk` without
`pipefail`. Make's `$(shell ...)` does not propagate the command status. A partial `go list ./...`
result followed by a non-zero exit can therefore become a successful filtered list, and the test
lane may exercise only that partial set. The first guard is even more vulnerable because
`grep -q` may exit after seeing `internal/godriver`, before `go list` reaches a later error.

That violates the rework requirement for an executable new-package drift guard and the map's own
claim that all 39 safe packages are derived fail-closed.

Required correction: materialize the complete `go list` or formatted importer output in a checked
assignment first (`rows="$$(...)" || exit 1`), then filter only the successfully materialized
output. Apply that pattern to both the safe package set and importer-set guard; assert the excluded
set is exactly `internal/godriver`, the safe set is non-empty, and every new package is included
automatically.

### F3 — the race evidence is stale relative to existing first-hand logs

Revision 3 repeatedly says the 30-minute race lane rests only on 918 s / 1121 s projections and
that the only executed numbers are 10-minute timeouts. `LOGBOOK.md` entry 1732 and the existing raw
log `.temp/TASK-260720-2284br/gates-rework1/gate-race.log` contradict that:

- `internal/install` passed under race in **609.117 s**;
- `internal/install/atomicity` passed under race in **1422.407 s**;
- the recorded gate used a 45-minute alarm and exited 0.

Entry 1732 also corrects atomicity's race factor to 4.02× and its projection to 1284–1494 s; the
map's 2.75× / 1121 s model is superseded. The exact full `go test -race ./... -timeout 30m` gate is
still unproven, so the implementation must still measure it, but the current evidence packet must
distinguish that full-suite gap from the already measured focused-package success.

Required correction: incorporate entry 1732 and the raw gate log into the fact ledger, update D6,
§6.3, §8, §9.1, and I7, and retain an exact future full-suite 30-minute gate without claiming it
green.

### F4 — the stored review contract is still incomplete

Checklist item 13, `Tests green`, remains unchecked. The cycle-2 gate-status artifact correctly
explains that this generic item conflicts with the no-Go/read-only audit scope and must not be
checked without a real command. An accepted reviewer verdict cannot silently treat an unchecked
definition-of-done item as satisfied.

Required correction: after F1–F3 rework, the board owner must explicitly reconcile item 13 as
not-applicable for this read-only audit or replace it with an executable task-scoped validation
item. Do not manufacture a green test claim and do not check the current item without evidence.

## Re-review gate

Revise only the task-scoped execution-map/gate-status evidence. Preserve the verified pin and
candidate identities, the 3/354 status count, Linux native non-gating prerequisite, and the explicit
future-gate posture. Do not edit product, CI, Makefile, target-task fields, spec, or pins. Re-review
can accept when native tool selection is absolute and identity-bound, Linux discovery fails closed,
the race ledger reflects entry 1732, and the stored checklist contract is reconciled.
