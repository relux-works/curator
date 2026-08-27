Fixes the two Curator CI failures that surface on the Linux lane once the
toolchain-identity gate stops failing first (PR #9 repairs that gate). Neither
has anything to do with a protocol vector.

Baseline on a native Linux runner, at this branch's base commit `bd6ba08`
(PR #9 head): run [30615765014](https://github.com/relux-works/curator/actions/runs/30615765014)
— `Lint` [job 91108467255](https://github.com/relux-works/curator/actions/runs/30615765014/job/91108467255)
red with 2 `unused` findings, `Test (ubuntu-latest)`
[job 91108467248](https://github.com/relux-works/curator/actions/runs/30615765014/job/91108467248)
red with exactly six `cmd/curator` cases failing.

## 1. Lint — two `unused` findings that exist only in the linux build

Both are genuinely dead code there, so both are **removed**. No `//nolint`, no
linter exclusion, no `_ = fn` reference trick — the `unused` check is not
weakened anywhere.

* `internal/godriver/controls_other.go` `(*controlDomain).destroy` had no
  caller on the `!darwin && !windows` build. The macOS and Windows domains each
  call their *own* `destroy` from inside their own `launch`; the shared worker
  client never calls it. Deleted, with a comment recording why no stub belongs
  there.
* `internal/transaction/namespace.go` `existingNamespaceAncestor` had exactly
  one consumer, `namespace_case_darwin.go`. Moved there. Windows answers the
  case- and normalization-sensitivity questions from the platform contract and
  the remaining unix builds from the POSIX one, so neither ever needs an
  ancestor to interrogate.

## 2. Test (ubuntu-latest) — six compiled-build cases on an uncovered host

`rc5-native-control-inventory-v1` defines native control records for exactly
macOS and Windows, so the go-v1 driver refuses a compiled build on any other
host *before a worker starts*:

```
go-v1 build_execution_control_unavailable: rc5-native-control-inventory-v1
defines no record for host linux (platform ""); the portable execution policy
is specified for macOS and Windows only
```

The six cases each need a **completed** compilation, so on such a host they have
nothing to assert. They are carved out by `requireNativeControlInventoryPlatform`,
which reads `godriver.InventoryPlatform` rather than a hand-written GOOS list —
the guard cannot drift from the inventory it stands for, and it stops skipping on
its own the moment the inventory gains a record for a host. No Linux execution
binding is fabricated and no unsupported execution is claimed.

### The carve-out is asserted, not obeyed

This mirrors what the repository already does one level up: `internal/godriver`
is excluded on linux by the root's own qualification vector
(`conformance-claim-v3-qualification.json`, `until_task: TASK-260728-1skseh` —
the open Linux qualification item), and `test-gate.sh` still runs
`TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` on that very runner to prove
the exclusion is fail-closed behaviour.

New `TestCompiledInstallFollowsTheNativeControlInventoryExactly` does the same
job at case granularity for `cmd/curator`. It runs on **every** runner and proves
the boundary from whichever side that runner is on:

* covered host → a real compiled install succeeds and publishes exactly one
  protected cache entry;
* uncovered host → the same invocation is refused with
  `build_execution_control_unavailable` naming the inventory and the host,
  publishes **nothing**, and leaves `status --check` non-zero rather than
  reporting a compiled command as current.

### Ledger and skip vocabulary

* `.github/ci/platform-cases.tsv` records all seven cases. The six carry
  `must_run_on=darwin,windows skip_allowed_on=linux class=platform-control`, so a
  future silent skip on macOS or Windows is fatal by name. The bidirectional
  assertion carries `must_run_on=linux,darwin,windows`.
* `.github/ci/skip-classes.tsv` gains one narrow `platform-control` pattern
  matching the inventory's own wording, `rc5-native-control-inventory-v1 defines
  no record for host `. The inventory carve-out itself is untouched.

## Verification

| Gate | Command | Exit |
| --- | --- | --- |
| lint, linux build | `GOOS=linux golangci-lint run ./...` (v2.12.2, CI's pin) | before `1` (the 2 findings), after `0` |
| lint, darwin build | `GOOS=darwin golangci-lint run ./...` | `0` |
| full gate, darwin | `test-gate.sh` against the committed `SPEC_PIN` root | `0` (go test `0`, platform-case gate `0`) |
| ledger | `ledger-consistency.sh` | `0` (56 rows across linux/darwin/windows) |
| gate self-test | `gate-selftest.sh` | `0` (75 passed, 0 failed) |
| suppression | `no-broad-suppression.sh` | `0` |
| format / vet | `gofmt -l cmd internal`, `go vet ./...` | clean / `0` |
| cross-build | `GOOS={linux,darwin,windows} go build ./...` | `0` |

## Out of scope — a known unrelated red on Windows

`Test (windows-latest)` fails at the `go vet` step on
`internal/runtimestore/targets_windows_test.go:97: undefined: decodeHelperOutput`,
before any test runs. That is BUG-260731-11bpa4, fixed in its own PR;
`internal/runtimestore` is deliberately untouched here.
