# TASK-260729-osjeay — review verdict, cycle 5

**Run:** `RUN-260729-d52d30`  
**Role:** reviewer  
**Artifact reviewed:** `TASK-260729-osjeay_final-ci-execution-map.md`, revision 5,
SHA-256 `1948f03811c54f59b2a5c1a1d32e01b43609a0cea8b0ffcb5ae6213400ff0d96`  
**Verdict:** **changes requested → `analysis`**

Revision 5 closes the four cycle-4 findings in the areas its harness covers, but it is not yet an
executable final producer map. Four command-contract gaps remain. They are research/document
rework, not product implementation work, so the correct route is `analysis`.

## Independently verified

- The attached no-Go harness has SHA-256
  `65a02fbee0bffe0f5dfefbe64f89f3537b5a32185c024ce0ed2a25e90c774e5a`.
  I inspected it before execution and ran it from the board resource:
  `sh .task-board/.resources/TASK-260729-osjeay/TASK-260729-osjeay_verify-recipes.sh`.
  It produced `ALL 21 EXPECTATIONS MET` and real process exit `0`.
- `main` and `origin/main` are both
  `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`; current `HEAD` is divergent
  `c06aa1a15e4093410a686ff0ce4f579fba59dec1`. Neither tip is an ancestor of the other.
- `origin/main:.github/workflows/ci.yml` still has the pin
  `00b1688a9b2457ca397a0bb550acf47cad8ee967` twice, `checkout@v4`,
  `setup-go@v5`, `golangci-lint-action@v7`, `version: latest`, no race job and no timeout.
  Current `Makefile` still has only the six existing targets; `go.mod` requires `go 1.25.5`.
- The immutable candidate facts remeasure exactly: manifest
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`,
  tree `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`,
  448 files, 3 modified + 354 untracked conformance paths.
- Current upstream release facts were re-read from the official GitHub API:
  [checkout v7.0.1](https://github.com/actions/checkout/releases/tag/v7.0.1),
  [setup-go v7.0.0](https://github.com/actions/setup-go/releases/tag/v7.0.0),
  [golangci-lint-action v9.3.0](https://github.com/golangci/golangci-lint-action/releases/tag/v9.3.0),
  and [golangci-lint v2.12.2](https://github.com/golangci/golangci-lint/releases/tag/v2.12.2).
  The map's retain/freeze disposition is source-current.

No Go command, test, build, vet, format, lint, install, dependency fetch, product edit, CI edit,
Makefile edit, pin edit, or `TASK-260720-1pvfj5` mutation was performed.

## Findings

### F1 — hosted Go jobs do not execute the toolchain identity contract

The map says hosted steps relax path shape only and that **every identity assertion still runs**
(I11, lines 2500–2515). The exact YAML does not implement that statement:

- `test` runs direct `gofmt`, `go vet`, and `go test` after `setup-go@v5`
  (lines 1553–1593), with no `require-toolchain` or equivalent comparison step.
- `interop` remains a direct Go command (lines 1758–1762), also without the assertions.
- `lint` runs the action after setup-go but does not verify the Go toolchain that the action uses.
- Section 6.0 asks a producer merely to *print* `go env GOTOOLCHAIN` once (lines 1544–1550), while
  I11 requires comparisons of exact version, root, launcher/formatter binding, `GOTOOLCHAIN=local`
  and `GOENV=off`.

`test-linux` and `race` do reach `require-toolchain` through Make targets, but even `test-linux`'s
final direct godriver rejection command bypasses it. Therefore the exact YAML and the invariant are
not the same executable contract.

**Required rework:** add an exact fail-closed identity step after `setup-go` in every Go-consuming
job (`test`, `test-linux`, `race`, `lint`, `interop`). Give both Unix and Windows executable forms.
Each must compare, not just print: `go.mod` → `go1.25.5`, `go version`, absolute reported GOROOT,
launcher identity, formatter identity where used, `GOTOOLCHAIN=local`, and `GOENV=off`.
`naming-gate` consumes no Go and needs no such step. Update the target/job correspondence and add
no-Go negative cases proving the exact hosted step rejects wrong version, auto toolchain, user
GOENV, and cross-root formatter on every shell form.

### F2 — source-origin enumeration still masks a producer failure

The map claims “No producer→consumer pipe anywhere in the transport path” (I13, lines 2520–2527),
but source step 1 is:

```sh
find ... -print0 | LC_ALL=C sort -z | xargs -0 shasum -a 256
```

(lines 989–995). POSIX `/bin/sh` returns the last pipeline command's status. I reproduced the exact
failure class without Go:

```text
find: /tmp/TASK-260729-osjeay-definitely-missing: No such file or directory
pipeline_exit=0
output=
```

The 21-case harness covers archive-producer failure and post-extraction missing/extra files, but it
does not inject failure into `find`, `sort`, or the digest producer. Thus 21/21 is true but narrower
than the stated staging invariant.

**Required rework:** materialize and status-check the origin path stream before sorting, materialize
and status-check the sorted stream before hashing, and status-check digest generation separately.
Do not depend on `pipefail`, which `/bin/sh` does not provide portably. Extend the no-Go harness with
healthy enumeration plus partial-then-failing `find`, failing sort, and failing digest cases. Each
injected failure must produce a real non-zero exit before archive creation.

### F3 — W2 permits a stale Windows source tree

The lane's control-host script uses `set -u` (line 1163), not `set -e`. W2 runs an unguarded:

```sh
ssh win "mkdir C:\Users\admin\AppData\Local\Temp\TASK-260720-1pvfj5"
```

(lines 1232–1239). The prose acknowledges that an existing directory exits 1, but the command does
not stop or invoke W9. W4 then extracts both archives into that same base (lines 1279–1281).
Candidate extras are caught by W5, but the map explicitly says the source has no post-extraction
equivalent (lines 1288–1292). A stale extra source file can therefore survive archive extraction
and become part of the suite input even though the source archive digest is correct.

**Required rework:** make W2 an executable empty-root precondition: fail if the base exists, create
it, then compare that it exists and is empty before W3. Every W2/W3 command needs an explicit
status guard under the chosen `set -u` contract. Add a Windows negative check that precreates an
extra source file and proves the lane stops before transport/extraction. Cleanup must be
status-checked and absence-confirmed before a retry.

### F4 — the Linux confirmation command violates the map's own native toolchain rule

Section 9.1 step 6 requires an approved GOROOT but executes:

```sh
CURATOR_CONFORMANCE_ROOT=<root> go test ./internal/godriver/ -count=1 -timeout 30m
```

(lines 2253–2259). This uses ambient `go`, omits `GOROOT`, `GOTOOLCHAIN=local`, and `GOENV=off`, and
calls the prerequisite “Go 1.25.x” although the map elsewhere correctly binds gates to exact
`go1.25.5`. It contradicts I10/I11 and repeats the defect revision 5 otherwise removes.

**Required rework:** express the command using one operator-approved absolute Linux GOROOT, derive
`GO_EXE="$GOROOT/bin/go"`, run the full toolchain preflight, then invoke that absolute executable
with `GOROOT`, `GOTOOLCHAIN=local`, and `GOENV=off`. State exact `go1.25.5`, not the broader 1.25.x
family.

## Re-review gate

Re-review revision 6 only after:

1. F1–F4 are corrected in the map without changing product, CI, Makefile, pins, or target-task
   fields.
2. The extended no-Go harness covers the origin-enumeration failures and the exact hosted identity
   step(s), with expected/actual exits and real harness exit 0.
3. The Windows stale-root negative is specified as a producer gate if no reachable Windows host
   exists; it must not be claimed executed.
4. All commands remain future producer gates and the accepted manifest/tree/count plus committed
   pin invariants remain unchanged.

