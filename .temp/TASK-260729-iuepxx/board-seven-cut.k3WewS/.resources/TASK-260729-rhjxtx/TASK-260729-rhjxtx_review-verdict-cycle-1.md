# TASK-260729-rhjxtx review verdict — cycle 1

## Verdict

**CHANGES REQUESTED** → `to-dev`

The packaged Go probe is scoped and its local quality gates are green, but the
machine-readable evidence is not internally consistent and the Kotlin/Native
and Windows evidence do not yet satisfy the task's explicit acceptance
criteria.

## Review performed

- Inspected all seven producer outcome artifacts and both JSON fixtures.
- Extracted `TASK-260729-rhjxtx_probe.tar.gz` without modifying producer code.
- Re-ran from the packaged module:
  - `gofmt -l .` → exit 0, empty output
  - `go vet ./...` → exit 0
  - `go test -count=1 ./...` → exit 0; 34 test functions
  - `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...` → exit 0, `0 issues.`
- Re-ran the authoritative local positive probes:
  - Rust `rustc -vV` and `rustc --print host-tuple` → exit 0, output matches the macOS fixture.
  - Swift `swiftc -print-target-info` → exit 0, output matches the macOS fixture.
- Parsed both fixtures with `jq`; their format, counts, unique case IDs,
  consumers, and `auto_install: false` guidance invariants pass.
- Checked the claimed commands against primary sources. In particular, the
  current Kotlin compiler reference identifies `kotlinc-native` and
  `-list-targets`:
  https://kotlinlang.org/docs/compiler-reference.html

## Required changes

### 1. Cross-host case expectations drift for `.swift-version`

The same case ID has two different expected classes:

```text
macOS:   swift.negative.selector-swift-version-file  expected=none              verdict=match
Windows: swift.negative.selector-swift-version-file  expected=package_selector  verdict=not_run
```

Cause in the packaged source:

- `swiftCaseIDs()` assigns `GatePackageSelector`.
- `swiftSelectorCase()` assigns `GateNone`.

An expected classification is part of the case contract and cannot vary merely
because one host lacks Swift. Make the expectation identical in executed and
`not_run` fixtures and add a regression test that checks the case-table
expectation against the executed case.

### 2. Swift compilation-boundary evidence contradicts itself

For `swift.negative.target-stdlib-absent`, the JSON and rendered command log say:

```text
upstream_rejects_after_compilation = true
```

The handoff and results document say Rust is the only corpus case where
upstream rejection arrives after compilation starts, and the captured Swift
stderr contains only the standard-library load diagnostic—no measured
compilation-progress evidence.

Cause in `swiftTargetCases()`:

```go
BuildStartedCompilation: compile.ExitCode != 0 && targetPrint.ExitCode == 0
```

That field is inferred from failure, not measured from output, contrary to the
`Observation` contract. Record a defensible measured signal or set it false,
regenerate both fixtures/logs/results consistently, and add a regression test.

### 3. The supplied-root Kotlin/Native branch does not implement its promised probes

Even when a Kotlin/Native root is supplied and the backend/version probe
succeeds, `probeKotlin()` returns all four substantive cases as `not_run` with:

```text
the Kotlin/Native case corpus is not authored
```

There is no native-target probe, no `-list-targets` observation, no positive
version/target case, and no malformed/future metadata or unknown-target control.
The metadata source remains `(undecided)`.

Current host absence must remain honest and must not trigger an install, but it
does not satisfy the DoD item requiring reproducible Kotlin/Native version,
target, and metadata probes. Implement the runnable supplied-root branch and
keep current macOS/Windows Kotlin cases `not_run` until a qualified host can
measure them. If the accepted artifact boundary genuinely provides no legal
metadata field, record that as the explicit measured/design input instead of
claiming the metadata probe is implemented.

### 4. Exact Windows execution evidence is missing

The artifacts state that the probe was cross-compiled, copied to SSH alias
`win`, hash-verified, executed, and removed. They do not record the exact
cross-compile, transfer, SSH invocation, hash, cleanup commands, and their exit
codes. `TASK-260729-rhjxtx_windows-inventory.log` records results but not the
commands or per-command exits.

Attach those exact commands/argv, outputs, and exit codes in a task-scoped
artifact, while preserving the current read-only/no-install behavior.

## What already passes

- No normative curator-spec, release pin, dependency, or tracked project source
  was modified.
- Rust and Apple Swift positive/negative host measurements are substantial and
  reusable after the evidence inconsistencies above are corrected.
- Current Kotlin/Native and Windows absence is reported honestly rather than
  treated as a pass.
- Packaged Go formatting, vet, tests, and lint are green.
- The repository-wide `make check` failure is correctly attributed to unrelated
  `.temp` trees; task-scoped tracked Go files are not implicated.

