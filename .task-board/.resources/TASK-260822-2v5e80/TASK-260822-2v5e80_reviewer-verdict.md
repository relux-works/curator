# TASK-260822-2v5e80 — reviewer verdict: ACCEPTED

**Run:** RUN-260822-ec3c84 (reviewer, claude) — not goal-bound (`spawn goal` reports "run is not goal-bound").
**Reviewed:** branch `task/TASK-260822-2v5e80-toolchain-shim-remedy`, worktree `.temp/TASK-260822-2v5e80/worktree`, base `origin/main` = `6a9b201`. Uncommitted, as repo policy requires.

## Verdict

Accepted. Every AC clause holds, verified by this reviewer running the gates as its own processes rather than reading the implementer's log.

## AC, clause by clause

| AC clause | Verified how | Result |
| --- | --- | --- |
| Remedy text asserted in tests | 3 new/extended cases: `internal/godriver/toolchain_remedy_test.go` (both mismatch sites, byte-exact detail + remedy + rendered line), `internal/install` `TestGoToolchainRemedyReachesTheOperatorIntact`, `cmd/curator` `TestInstallPrintsTheRemedyAVersionManagerSelectionEarns` | met |
| Protocol string byte-identical | `git show 6a9b201:internal/godriver/session.go` vs worktree — both literals unchanged character for character; the remedy rides in a separate `Diagnostic.Remedy` field, never folded into `Detail` | met |
| go test green | see gates below | met |

Baseline vs. head, both sites:

- `selected Go executable is not the regular executable under the derived GOROOT` — unchanged (`session.go:503` → `session.go:522`)
- `go env GOROOT does not match the selected toolchain` — unchanged (`session.go:650` → `session.go:670`)

## Gates re-run by the reviewer

Toolchain `GOROOT=/Users/iv/.goenv/versions/1.25.5`, `GOTOOLCHAIN=local`, `GOENV=off`, real GOROOT/bin fronting PATH. go1.25.5 darwin/arm64.

| Gate | Exit | Notes |
| --- | ---: | --- |
| `gofmt -l cmd internal` | 0 | no output |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `golangci-lint run` | 0 | 0 issues |
| `go test -count=1 ./internal/godriver/ ./internal/install/` | 0 | 41.1s / 50.0s, both ok |
| `go test -count=1 -run TestInstallPrintsTheRemedyAVersionManagerSelectionEarns ./cmd/curator/` | 0 | 0.8s |

Not re-run by the reviewer: the full `go test -count=1 ./...` (`cmd/curator` alone is ~527s). The implementer attached its log — `TASK-260822-2v5e80_go-test-all.log`, 41 packages ok, 0 FAIL, exit 0 — and this review re-ran all three touched packages plus build/vet/lint independently. Proportionate to a two-site diagnostic change.

## Architecture fit

Sound, and the constraint that forced the design is real:

- **The CLI guidance layer is closed to PATH advice.** `goToolchainGuidance` (`cmd/curator/builds.go:751`) documents that it "deliberately never suggests a PATH lookup", and `cmd/curator/builds_test.go:678` pins that it may not contain the substring `path` outside its closed-rule sentence. A PATH remedy therefore cannot live there; carrying it on the driver diagnostic is the only clean route. Confirmed by reading both.
- **The remedy does not displace the closed selection rule.** `TestInstallPrintsTheRemedyAVersionManagerSelectionEarns` asserts stderr still carries "Curator never searches PATH and never downloads a toolchain" alongside the remedy. Good defensive test for exactly the tension this change introduces.
- **Conformance is unaffected.** `builddriver_rejection_conformance_test.go:513` compares `DiagnosticCode` only, never detail text. Read and confirmed. No `.md`/`.json` in the repo carries either protocol string.
- **`Error()` rewrite is behaviour-preserving for every remedy-free boundary** (code-only and code+detail), and `TestDiagnosticRenderingIsUnchangedWithoutARemedy` pins that.
- **No unkeyed `Diagnostic` literals** anywhere in the repo, so adding the field breaks no construction site (`internal/install/stage_test.go:1306` is keyed).
- The `diagnosticRemedy` / `diagnosticErrRemedy` pair mirrors the existing `diagnostic` / `diagnosticErr` vocabulary. Consistent with the file.

## Confirmed limitation — disclosed, and correctly analysed

The implementer's report-row truncation finding is accurate; this reviewer traced it rather than taking it on trust.

`report()` (`cmd/curator/builds.go`) applies `install.RedactDiagnostic` to the **whole combined** detail, not to the reason alone:

- operator failure line (`Result.Errors`): 193 runes — under the 240-rune `maxDiagnosticRunes` bound, so the remedy survives whole. Asserted by the install test.
- build report row: 101-rune prefix + the same 193 = 294 runes, passed through `install.RedactDiagnostic(detail)` in `report()` → truncated at 240, and the remedy is exactly what the ellipsis eats.

This is out of the task's scope — making it visible there needs `maxDiagnosticRunes` or the prefix changed — and it is already recorded in `LOGBOOK.md` §2051. The primary operator surface (`curator install` stderr) does carry it, which is what the task asked for. Not grounds for rework; worth a follow-up task if the report row matters.

## For the commit-owning mover

1. **Canonical branch: `task/TASK-260822-2v5e80-toolchain-shim-remedy`** (worktree `.temp/TASK-260822-2v5e80/worktree`). Commit this one.
2. **Discard the duplicate.** `task/TASK-260822-2v5e80-toolchain-shim-remedy-r88df86` and worktree `.temp/TASK-260822-2v5e80/worktree-r88df86` came from the concurrent RUN-260822-88df86 spawn. This reviewer diffed the two: same design, same operator text byte for byte, only comment wording differs, and the duplicate's godriver test file is `diagnostic_remedy_test.go` with fewer cases. Nothing in it is worth keeping. `git worktree remove` + `git branch -D`.
3. Also clean `.temp/TASK-260822-2v5e80/baseline` (detached scratch worktree).
4. Nothing is committed or pushed. Per the reviewer archetype constraint this run supplied no `commit_ack`; the mover makes the final `done` transition for any enforced Bug/Story scope after committing.

## Files in scope

| File | Change |
| --- | --- |
| `internal/godriver/errors.go` | `Diagnostic.Remedy` field; `Error()` renders `go-v1 <code>: <detail>; <remedy>`; two new constructors |
| `internal/godriver/session.go` | `toolchainSelectionRemedy` const; both `toolchain_executable_mismatch` sites carry it |
| `internal/godriver/toolchain_remedy_test.go` (new) | both sites, byte-exact strings; remedy-free rendering unchanged |
| `cmd/curator/toolchain_remedy_test.go` (new) | end-to-end through `install`, plus the closed-rule non-displacement assertion |
| `internal/install/diagnostics_test.go` | remedy survives `RedactDiagnostic` on the operator failure surface |
