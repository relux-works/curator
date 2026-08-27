## CI evidence — run [30620349565](https://github.com/relux-works/curator/actions/runs/30620349565)

**The acceptance criteria are met on the native Linux runner.**

| Job | Baseline at `bd6ba08` ([30615765014](https://github.com/relux-works/curator/actions/runs/30615765014)) | This PR |
| --- | --- | --- |
| `Lint` | **FAILURE** — 2 `unused` | **pass** (51s) |
| `Test (ubuntu-latest)` | **FAILURE** — 6 cases | **pass** (1m36s) |
| `Race (ubuntu-latest)` | **FAILURE** | **pass** (4m0s) |
| `Interop conformance gate` | pass | pass |
| `Gate self-test` ×3 | pass | pass |
| `Naming gate` | pass | pass |
| `Test (windows-latest)` | FAILURE at `go vet` | **fail** at `go vet` — unchanged, see below |

### Focused evidence, from the run's own uploaded `test-evidence-ubuntu-latest`

Not asserted — read back out of the artifact:

* `skips-observed.tsv` — all six carved-out cases recorded as class
  `platform-control`, verdict `tolerated-by-ledger`, with the verbatim reason
  `rc5-native-control-inventory-v1 defines no record for host linux; the portable
  execution policy is specified for macOS and Windows only`. Nothing
  `UNCLASSIFIED`, nothing `FATAL-*`.
* `observed-cases.tsv` — `pass cmd/curator
  TestCompiledInstallFollowsTheNativeControlInventoryExactly`. **The
  uncovered-host branch executed on native Linux**, so the refusal carrying
  `build_execution_control_unavailable`, the zero published cache entries, and
  the non-zero `status --check` are all observed, not inferred.
* `go-test-assert-godriver.json` — `pass
  TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`: the pre-existing
  whole-package exclusion is still asserted on the excluded runner.
* `ledger-consistency.txt` — `ok`, 56 rows across linux/darwin/windows.
* `platform-cases.txt` — `platform-case gate: ok`.

### The new skip class was tested for narrowness, not assumed

The one entry added to `skip-classes.tsv` is the only place this change could
have introduced a general escape hatch, so it was driven adversarially against
`platform-case-gate.sh` with synthetic `go test -json` streams. Gate verdict for
the *same* ledger case under four reasons:

| reason | GOOS | verdict |
| --- | --- | --- |
| `rc5-native-control-inventory-v1 defines no record for host linux; …` | linux | `platform-control` / **`tolerated-by-ledger`** |
| same | darwin | `platform-control` / **`FATAL-not-tolerated`** |
| `compiled builds are not supported here` | linux | `UNCLASSIFIED` / **`FATAL-wrong-class`** |
| `symlinks unavailable` | linux | `host-capability` / **`FATAL-wrong-class`** |

The pattern admits exactly the inventory's own statement, on exactly the platform
the ledger tolerates it on. None of the six cases can silently disappear from
macOS or Windows.

### Local gates, real exit codes, each run unpiped as its own process

| Gate | Exit |
| --- | --- |
| `GOOS=linux golangci-lint run ./...` (v2.12.2, CI's pin) — **before** | `1` (the two findings) |
| `GOOS=linux golangci-lint run ./...` — **after** | `0` |
| `GOOS=darwin golangci-lint run ./...` | `0` |
| `test-gate.sh` on darwin against the committed `SPEC_PIN` root | `0` (go test `0`, gate `0`) |
| `ledger-consistency.sh` | `0` — 56 rows |
| `gate-selftest.sh` | `0` — 75 passed, 0 failed |
| `no-broad-suppression.sh` | `0` |
| `gofmt -l cmd internal` / `go vet ./...` | clean / `0` |
| `GOOS={linux,darwin,windows} go build ./...` | `0` each |

`Test (macos-latest)` and `Race (macos-latest)` were still queued on GitHub's
macOS pool when this was written; the full darwin gate is green locally and both
were green at the baseline run.

### Out of scope — the Windows red is not this PR's

`Test (windows-latest)` fails at **step 7, `go vet`**, before `Ledger consistency`
and `go test` ever run:

```
vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

That is BUG-260731-11bpa4 / #10, stacked on the same `bd6ba08` base.
`internal/runtimestore` is that PR's exclusive ownership and is untouched here.
Consequence stated plainly rather than worked around: this PR cannot show a green
`Test (windows-latest)` until #10 lands, so the six cases' `must_run_on=windows`
is a correct declaration of where the AC requires them — not an observed pass on
that runner. If they then fail there, the ledger names them instead of dropping
them.
