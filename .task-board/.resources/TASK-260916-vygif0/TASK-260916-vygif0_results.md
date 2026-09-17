# TASK-260916-vygif0 results \u2014 pin the Rust toolchain in every CI lane (rev2)

Producer (developer) handoff evidence. Shell for all commands below: bash.
Worktree: `.temp/STORY-260915-3w11un/worktree`, branch
`task-board/story/STORY-260915-3w11un`, uncommitted. rev1
(`TASK-260916-vygif0_change-request_rev1.patch`, 11 paths) is the base; rev2
adds the two Windows-portability fixes below and nothing else.

## Decision implemented (unchanged from rev1)

Operator decision 2026-09-17: pin, do not install by hand \u2014 the pnpm pattern.
Supported release is Rust 1.91.0, read from the code: the closed Cargo
registry pins `Version 1.91.0` / commit ea2d97820c16195b0ca3fadb4319fe512c199a43,
the toolchain-root selector joins `1.91.0-<target>`, and `cargoToolchain.validate`
admits only `1.91.0`.

## Why rev2 exists: rev1 remote-gate failures

CR revision 1 validation (remote gate run 35144488083) failed on exactly two
Windows jobs; ubuntu/macos/race/lint/interop/naming were green:

1. `Gate self-test (windows-latest)`: 2 FAIL rows \u2014 `a runner without rustup
   fails` (exit 127, want 1) and `the failure names the runner-setup note`.
   Root cause: the negative case ran the installer with `PATH` replaced by a
   bare single directory containing only an `awk` symlink. On Git Bash for
   Windows, awk then died loading its DLLs (`cannot open shared object
   file`), so the row failed for the wrong reason instead of exercising the
   missing-rustup branch.
2. `Test (windows-latest)`: `go test exit=1`, platform-case gate exit 0. Root
   cause, from the `test-evidence-windows-latest` artifact stream
   (`go-test.json`): the single failing test was the rev1-new
   `internal/rustsource :: TestCargoToolchainValidateAdmitsOnlyTheSupportedVersion`.
   Its fixture used a Unix-literal `CargoPath: "/pinned/toolchains/..."`, and
   `validate` gates on `filepath.IsAbs` \u2014 which on Windows is false for a
   path with no volume (`volumeNameLen == 0` \u2192
   `src/internal/filepathlite/path_windows.go: IsAbs` returns false). So the
   "supported admitted" leg failed on Windows only. Recent main runs
   (35144897219, 35143517379, 35141193640, 35139633418) all show
   `Test (windows-latest): success`, confirming the regression is mine, not
   pre-existing.

## rev2 changes (2 files, on top of rev1)

- `internal/rustsource/toolchain_registry_test.go`:
  `TestCargoToolchainValidateAdmitsOnlyTheSupportedVersion` now builds
  `CargoPath` from `filepath.Join(t.TempDir(), "toolchains",
  SupportedRustToolchainVersion+"-test", "bin", "cargo")` \u2014 absolute on every
  GOOS by construction, mirroring the sibling
  `TestCargoHostCapabilityReasonClassifiesOnlyAbsence` cases that passed on
  Windows in the rev1 evidence. Test expectation unchanged (supported
  admitted, `9.9.9-drift-probe` rejected); only the non-portable fixture
  path is fixed.
- `.github/ci/gate-selftest.sh`: the rustup-absent negative case no longer
  replaces `PATH`; it filters it, dropping only entries shipping a `rustup`
  / `rustup.exe` executable and keeping every other entry (with `BASH_ABS`
  invocation as before). awk/bash keep working on Git Bash while
  `command -v rustup` genuinely fails, so the row exercises the real
  missing-rustup branch and its `docs/self-hosted-runner-setup.md` message.

No production-code change in rev2 (the `validate` rejection of a
non-absolute path on Windows is correct behavior; the fixture was wrong).

## Verification (all rerun by me in this session, real exit codes)

- `bash .github/ci/gate-selftest.sh` \u2192 exit 0, 180 passed, 0 failed
  (includes the two fixed rustup-absent rows, green).
- `bash -n` on gate-selftest \u2192 ok.
- PATH-filter probe with a simulated rustup on PATH (`/tmp/probe-norustpath.sh`,
  throwaway): filter drops the rustup dir \u2192 yes; `command -v rustup` hidden
  under filtered PATH \u2192 yes; awk works \u2192 yes; bash resolves \u2192 yes. This is the
  hosted-image condition (rustup present) this rustup-less dev host otherwise
  cannot reproduce; the remote gate reruns the real row on all three OSes.
- `go test ./internal/rustsource/ -run
  'TestCargoToolchainValidateAdmitsOnlyTheSupportedVersion|TestSupportedRustToolchainVersionIsTheRegistrySource|TestCargoHostCapabilityReasonClassifiesOnlyAbsence'
  -count=1 -v` \u2192 PASS (all three).
- `go test ./internal/rustsource/ -count=1 -timeout 8m` \u2192 ok (7.657s).
- `GOOS=windows go build ./internal/rustsource/` \u2192 ok;
  `GOOS=windows go vet ./internal/rustsource/` \u2192 ok.
- `go test ./internal/rustsource/ -run 'TestProductionManager' -count=1 -v`
  \u2192 both SKIP with `no operator-approved Cargo descriptor for native target
  x86_64-apple-darwin`: this dev host is darwin/amd64 with no rustc, so the
  production cases cannot execute here (brief allows stating this).
- `bash .github/ci/rust-pin-guard.sh` \u2192 exit 0,
  `toolchain file and Go agree on Rust 1.91.0`.
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev` \u2192 exit 0,
  239 rows checked, both new rust rows
  `ok [must=darwin skip=linux,windows]`.
- `gofmt -l internal/rustsource/` \u2192 clean; `go vet` \u2192 exit 0;
  `golangci-lint run ./internal/rustsource/` \u2192 0 issues;
  `ruby -ryaml` load of ci.yml \u2192 ok.
- Full landing suite NOT run manually: it runs exactly once via handoff,
  per campaign rules. Remote gate (GitHub + rose-air) is the handoff gate.

## Operator action (required before rose-air goes green; unchanged)

The rose-air lane stays red until the operator installs rustup once on
macbook-iv per `docs/self-hosted-runner-setup.md`. Until then the lane fails
in `Install pinned Rust toolchain via rustup` with that note named \u2014 by
design, not with a generic rustc-not-found. No toolchain needs installing by
hand: the lane installs 1.91.0 itself on every run. (Not done by me: host
state is outside the worktree brief.)

## Findings / notes for the reviewer

- The two failures were independent, both Windows-only, both mine: a
  Unix-path-literal fixture in a new Go test, and a PATH-replacement trick in
  a new self-test row that Git Bash cannot honor. The durable collateral is
  the fixed rows themselves: the Go test now uses the same `t.TempDir()`
  pattern as its passing siblings, and the self-test row now filters PATH
  the way the ci.yml comment already requires ("must behave identically
  under Git Bash on Windows").
- Negative-evidence bounds (rev2 deltas): the Go narrowing still proves
  admit-supported vs reject-`9.9.9-drift-probe` with a now-portable absolute
  path; the installer narrowing still proves filed-channel install plus
  rustup-absent failure naming the setup note, with the absence now produced
  by entry filtering (probe above) rather than PATH replacement.
  Production call sites: the guard/install steps in the four `test-gate.sh`
  lanes of ci.yml, enforced by platform-case-gate.sh.
- `rust-toolchain.toml` remains inert against the rustsource manager's own
  cargo runs (direct `<toolchain-root>/bin/cargo` resolution, never the
  rustup shim); fixture workspaces live in `t.TempDir()`.
- No LOGBOOK.md edit: campaign producer rules forbid worktree LOGBOOK edits;
  findings are recorded here instead.

## Checklist (task DoD)

- [x] rust-toolchain.toml pinned exactly; every rustsource lane installs it via rustup before go test
- [x] rust-pin-guard.sh catches drift vs internal/rustsource supported version; self-test rows pass/fail
- [x] Runner setup note: rustup once on macbook-iv; rose-air fails with a clear message if rustup is absent
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] Outcome artifact attached (this file)
- [x] Findings recorded here (no LOGBOOK.md edit per campaign rules)

revision 3 = revision 2 unchanged; republished after the amended review
