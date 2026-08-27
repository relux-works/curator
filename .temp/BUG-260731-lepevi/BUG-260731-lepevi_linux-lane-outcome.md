# BUG-260731-lepevi — Curator CI Linux lane

**Branch:** `task/BUG-260731-lepevi-linux-lane`
**Base:** `bd6ba08acda3dc801512c408c759ac0ac6f79f26` (head of Curator PR #9, the
toolchain-identity repair). Branching off `main` (`cfffd7c`) would have hidden
the evidence: `main` still fails every Go job at the toolchain-identity gate
before Lint or Test findings are reached.
**Commit:** `b2ac7d7` — SSH-signed, `git log %G?` = `G`.
**PR:** https://github.com/relux-works/curator/pull/11 → `main`.
**Worktree:** `.temp/BUG-260731-lepevi/worktree` (primary checkout untouched).

---

## 1. Baseline reproduction on a native Linux runner

Reproduced at the exact base commit, on GitHub's `ubuntu-latest` runner —
run [30615765014](https://github.com/relux-works/curator/actions/runs/30615765014):

| Job | Id | Result |
| --- | --- | --- |
| `Lint` | 91108467255 | FAILURE — 2 `unused` findings |
| `Test (ubuntu-latest)` | 91108467248 | FAILURE — 6 `cmd/curator` cases |

Verbatim lint findings:

```
internal/godriver/controls_other.go:35:30: func (*controlDomain).destroy is unused (unused)
internal/transaction/namespace.go:310:6: func existingNamespaceAncestor is unused (unused)
```

The six failing cases, extracted from the uploaded `go-test.json` of that run
(`test-evidence-ubuntu-latest`), all in `cmd/curator`:

```
TestCompiledProjectStatusRepairRollbackRecovery
TestGCRetainsAndReportsReferencedCompiledState
TestGlobalStatusReportsATransitivelyResolvedCompiledCommand
TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck
TestStatusReportsATransitivelyResolvedCompiledCommand
TestStatusReportsAnUnusableToolchainPerCompiledCommand
```

Their common cause, verbatim from the same stream:

```
error: build-skill.build-tool: go-v1 build_execution_control_unavailable:
rc5-native-control-inventory-v1 defines no record for host linux (platform "");
the portable execution policy is specified for macOS and Windows only
```

The full failing-run evidence artifact is archived under
`.temp/BUG-260731-lepevi/evidence/baseline/`.

**A local native Linux runner was NOT used.** `lima` is installed on this host
but `limactl start` could not boot a `vz` VM in this environment (two attempts,
one timed out at 7 minutes with no console output, the instance stayed
`Stopped`). Docker/podman/colima are not installed. The lint findings were
therefore reproduced locally by cross-target analysis (`GOOS=linux
golangci-lint run`, which is faithful because both findings are build-tag
driven, not host driven — same file, line, column and rule as the native
runner), and the compiled-control failures were reproduced and re-verified on
GitHub's real `ubuntu-latest` runner, which is a native Linux runner.

---

## 2. Lint — two findings, both genuinely dead, both removed

No `//nolint`, no linter exclusion, no `_ = fn` reference trick. The `unused`
check is not weakened anywhere; `no-broad-suppression.sh` passes.

### `(*controlDomain).destroy` — `internal/godriver/controls_other.go`

Call-site analysis over the whole module: `domain.destroy(command)` appears at
`controls_darwin.go:188` and `controls_windows.go:161` only. Each of those files
defines its own `destroy` and calls it from inside its own `launch`. The shared
worker client (`workerclient.go`) calls `launch`, `installedControls` and
`close` — never `destroy`.

So on the `!darwin && !windows` build the method has no reachable caller in any
build configuration. Deleted, with a comment in its place recording why no stub
belongs there. The other stubs in that file stay: `close` is called from shared
code at `workerclient.go:391`, `launch` at `:131`, `installedControls` at `:136`.

### `existingNamespaceAncestor` — `internal/transaction/namespace.go`

Exactly one consumer in the module: `namespace_case_darwin.go`, twice. Moved
into that file. Windows answers `namespaceCaseInsensitive` /
`namespaceNormalizationInsensitive` from the platform contract and the remaining
unix builds answer from the POSIX one, so neither ever needs an existing
ancestor to interrogate — Darwin is the only platform that asks the filesystem.

---

## 3. Test — the compiled-build carve-out

### What was decided, and why it is not a forced fit

`rc5-native-control-inventory-v1` (`internal/godriver/controls.go`) defines
`nativeControlPlatforms` records for exactly `macos` and `windows`.
`InventoryPlatform(goos)` returns `""` for anything else, and
`probeNativeControlsFor` then rejects with `build_execution_control_unavailable`
**before a worker starts**. That is the spec's position, not an accident.

Each of the six cases needs a *completed* compilation to have anything to
assert. Two paths were rejected:

* fabricating a Linux execution binding — the spec does not grant it; that is
  the forced fit the task explicitly forbids;
* weakening the inventory or the qualification vector — the AC forbids it.

The remaining honest option is the one the repository already implements one
level up. `internal/godriver` as a whole package is excluded on linux by the
supplied root's own `vectors/conformance-claim-v3-qualification.json`
(`until_task: TASK-260728-1skseh` — **the open Linux qualification item** named
in the curator-spec conformance flow and quoted in
`.github/ci/platform-exclusions.tsv`), and `test-gate.sh` still runs
`TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` on that very runner so the
exclusion is *asserted*, not obeyed.

This change applies the same shape at case granularity.

### `requireNativeControlInventoryPlatform` (`cmd/curator/status_test.go`)

```go
if godriver.InventoryPlatform(runtime.GOOS) == "" {
    t.Skipf("rc5-native-control-inventory-v1 defines no record for host %s; "+
        "the portable execution policy is specified for macOS and Windows only", runtime.GOOS)
}
```

The predicate is read from the inventory rather than written as a GOOS list, so
it cannot drift from what it stands for, and it stops skipping on its own the
moment the inventory gains a record for a host. Applied to exactly the six
cases — nothing else was touched, and the 55 other `cmd/curator` cases that
already pass on linux still run there.

### The assertion that makes the skip a boundary, not an omission

New `TestCompiledInstallFollowsTheNativeControlInventoryExactly` runs on **every**
runner and proves the claim from whichever side that runner is on:

* covered host → a real compiled install succeeds and publishes exactly one
  protected cache entry;
* uncovered host → the same invocation is refused, `stderr` carries
  `build_execution_control_unavailable`, `rc5-native-control-inventory-v1` and
  `no record for host <goos>`, **nothing** is published to the protected build
  cache, and `status --check` stays non-zero rather than reporting a compiled
  command as current.

The uncovered branch could not be executed on this host, so its two
non-obvious assertions were verified on darwin through the equivalent
"driver refused before producing an artifact" shape (install with
`CURATOR_GO` pointing at a nonexistent toolchain), via a scratch probe that was
run and then deleted:

```
install code=1
published cache entries: 0 []
status --check code=1   (app: build-skill not-installed, state=not-installed)
```

The linux-specific strings the test matches are quoted verbatim from the
baseline native-runner stream above. **CI on `ubuntu-latest` then confirmed it
for real**: the case is recorded `pass` in that run's `observed-cases.tsv`, so
the uncovered branch and all four of its assertions executed on native Linux
(see §4).

### CI ledger and skip vocabulary

`.github/ci/platform-cases.tsv` — 7 new rows:

| test | must_run_on | skip_allowed_on | class |
| --- | --- | --- | --- |
| `TestCompiledInstallFollowsTheNativeControlInventoryExactly` | linux,darwin,windows | - | - |
| the six carved-out cases | darwin,windows | linux | platform-control |

A future silent skip of any of the six on macOS or Windows is now fatal by name,
and `ledger-consistency.sh` proves from `go list` on each target GOOS that every
row's case really is compiled into that platform's build.

`.github/ci/skip-classes.tsv` — one new narrow `platform-control` pattern,
`rc5-native-control-inventory-v1 defines no record for host `, matching the
inventory's own wording. That class is documented in the table itself as "the
AC's explicit carve-out and the ONLY class that may hide a whole platform's
behaviour", which is exactly what this is. The inventory and the exclusion
tables are otherwise untouched.

`README.md` gains a "The compiled-build platform carve-out" subsection under
*Gates and tooling*, documenting both granularities and what has to change when
the inventory gains a Linux record.

### The new skip class was tested for narrowness, not assumed to be narrow

A new entry in `skip-classes.tsv` is the one place this change could have
smuggled in a general escape hatch, so it was driven adversarially against
`platform-case-gate.sh` with synthetic `go test -json` streams
(`.temp/BUG-260731-lepevi/evidence/local/adversarial/`). Verdict recorded by the
gate for the same ledger case under four different skip reasons:

| stream | reason | GOOS | gate verdict |
| --- | --- | --- | --- |
| the real one | `rc5-native-control-inventory-v1 defines no record for host linux; …` | linux | `platform-control` / **`tolerated-by-ledger`** |
| the real one | same | darwin | `platform-control` / **`FATAL-not-tolerated`** |
| vague | `compiled builds are not supported here` | linux | `UNCLASSIFIED` / **`FATAL-wrong-class`** |
| wrong class | `symlinks unavailable` | linux | `host-capability` / **`FATAL-wrong-class`** |

So the pattern admits exactly the inventory's own statement, on exactly the
platform the ledger tolerates it on. It cannot be reused to hide an unrelated
skip, and none of the six cases can silently disappear from macOS or Windows.

---

## 4. Native-runner result on PR 11

Run [30620349565](https://github.com/relux-works/curator/actions/runs/30620349565):

| Job | Baseline at `bd6ba08` | PR 11 |
| --- | --- | --- |
| `Lint` | **FAILURE** (2 `unused`) | **pass** (51s) |
| `Test (ubuntu-latest)` | **FAILURE** (6 cases) | **pass** (1m36s) |
| `Race (ubuntu-latest)` | **FAILURE** | **pass** |
| `Interop conformance gate` | pass | pass |
| `Gate self-test` ×3 | pass | pass |
| `Naming gate` | pass | pass |
| `Test (windows-latest)` | FAILURE (`go vet`) | **fail** (`go vet`, unchanged — see §5) |
| `Test (macos-latest)` | pass | **pass** |
| `Race (macos-latest)` | pass | **pass** (10m41s) |

**The acceptance criteria — `Lint` and `Test (ubuntu-latest)` green — are met on the
native Linux runner.** Every other job in the run is green too; the only red is
`Test (windows-latest)`, at the sibling-owned `go vet` step (§6).

From that run's uploaded `test-evidence-ubuntu-latest`, verified rather than
assumed:

* `skips-observed.tsv` — all six carved-out cases recorded with class
  `platform-control`, verdict `tolerated-by-ledger`, and the verbatim reason
  `rc5-native-control-inventory-v1 defines no record for host linux; the
  portable execution policy is specified for macOS and Windows only`. Nothing
  is `UNCLASSIFIED` or `FATAL-*`;
* `observed-cases.tsv` — `pass cmd/curator
  TestCompiledInstallFollowsTheNativeControlInventoryExactly`. **The
  uncovered-host branch executed and passed on native Linux**, so the refusal,
  the zero published cache entries and the non-zero `status --check` are
  observed there, not inferred from the darwin analogue;
* `go-test-assert-godriver.json` — `pass
  TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`, i.e. the pre-existing
  whole-package exclusion is still asserted on the excluded runner;
* `ledger-consistency.txt` — `ok`, 56 rows across linux/darwin/windows;
* `platform-cases.txt` — `platform-case gate: ok`.

And from the same run's `test-evidence-macos-latest`, the covered side of the
same claim: all seven cases recorded `ok` in `platform-cases.txt`, and
**zero `cmd/curator` skips** in `skips-observed.tsv` — the guard does nothing on
a platform the inventory covers, and
`TestCompiledInstallFollowsTheNativeControlInventoryExactly` passed there too.
So both branches of the bidirectional assertion are proven on real runners, not
one branch proven and the other argued.

---

## 5. Verification — commands and real exit codes

Every gate was run as its own process, unpiped; the exit code below is the real
`$?` of that process.

| Gate | Command | Exit |
| --- | --- | --- |
| lint, linux build, **before** | `GOOS=linux golangci-lint run ./...` (v2.12.2, CI's pin) | `1` — the two `unused` findings |
| lint, linux build, **after** | same | `0` — `0 issues.` |
| lint, darwin build | `GOOS=darwin golangci-lint run ./...` | `0` |
| full gate, darwin | `test-gate.sh` against the committed `SPEC_PIN` root | `0` (go test `0`, platform-case gate `0`) |
| ledger consistency | `ledger-consistency.sh` | `0` — 56 rows across linux/darwin/windows |
| gate self-test | `gate-selftest.sh` | `0` — 75 passed, 0 failed |
| broad suppression | `no-broad-suppression.sh` | `0` |
| format | `gofmt -l cmd internal` | clean |
| vet | `go vet ./...` | `0` |
| cross-build | `GOOS={linux,darwin,windows} go build ./...` | `0` each |

Raw outputs under `.temp/BUG-260731-lepevi/evidence/local/`.

### `GOOS=windows golangci-lint` — expected red, out of scope

```
internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput (typecheck)
```

Exit is non-zero for the analysis. This is BUG-260731-11bpa4's territory —
`internal/runtimestore` is the sibling's exclusive ownership and was not touched.
CI's `Lint` job runs on `ubuntu-latest` only, so it does not gate this PR.

### One flake observed, not caused by this change

The first full darwin `test-gate.sh` run failed one case:
`internal/install TestStrictRegistryPolicyFailsUnknown`, on
`registry test-reg snapshot timestamp is too far in the future` — a clock-skew
tolerance check in `internal/registry/snapshot.go:159`, load-sensitive under the
whole-module parallel run. It passed in isolation on both the unmodified base
and the modified tree, and the full gate re-run was `0`. `Test (macos-latest)`
was already green at the baseline run. Recorded here as an observation; it is
outside this bug's scope and no code was changed for it.

---

## 6. Out of scope — the Windows lane stays red for a sibling's reason

`Test (windows-latest)` fails at the `go vet` step:

```
vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

That is before any test runs, and it is BUG-260731-11bpa4 (PR 10, stacked on the
same `bd6ba08` base). Per the orchestrator's ownership boundary this PR does not
touch `internal/runtimestore`. The consequence is stated plainly rather than
worked around: **this PR cannot show a green `Test (windows-latest)` until the
sibling lands**, so the six cases' `must_run_on=windows` claim is a correct
declaration of where the AC requires them, not an observed pass on that runner.
The ledger is the right place for that claim: if they then fail on Windows, the
gate reports them by name instead of quietly dropping them.

The AC targets `Lint` and `Test (ubuntu-latest)`, both of which this change
addresses directly.
