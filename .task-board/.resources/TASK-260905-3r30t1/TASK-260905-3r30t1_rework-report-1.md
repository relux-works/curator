# TASK-260905-3r30t1 rework report 1: hosted platform-case gate

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch
`feat/byte-exact-acquisition`, PR https://github.com/relux-works/curator/pull/58.
Host: macOS, `git version 2.50.1 (Apple Git-155)`, Go from go.mod (1.25.5).

## Commit

```
5abec244 Satisfy the hosted platform-case gate for byte-exact extraction   (on top of 5beced46, no rewrite)
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
```

Files: `internal/gitops/gitops_test.go`, `internal/interop/snapshot_acquisition_test.go`,
`.github/ci/platform-cases.tsv`. No production code touched; `skip-classes.tsv` untouched.

## What CI reported at 5beced46 and what changed

1. `required on linux/darwin but not compiled into that build: internal/gitops :: TestArchiveRejectsLinks`
   (ledger-consistency, all three Test jobs) and `required case never ran` (Race jobs).
   The rework had renamed the case to `TestExtractRejectsLinks`.
   Fix: the case is back under the exact name `TestArchiveRejectsLinks`. It no longer needs `ln`
   or host symlink support: the tree is written through `git mktree -z` with a `120000 blob <oid>`
   entry, `Extract` must refuse it with the "links" error class, and the test also asserts the
   refused path was not written to the destination. Because the case now runs everywhere, the
   ledger row was changed from `linux,darwin | skip windows host-capability` to
   `linux,darwin,windows | - | -`.
2. `skip with an unrecognised reason: internal/interop :: TestConformanceSnapshotAcquisition`
   (reason "conformance root … has no vectors/snapshot-acquisition.json").
   Fix chosen: change the skip text to `conformance root <root> publishes no
   vectors/snapshot-acquisition.json (pre-environments suite; root-content)`, which the existing
   `root-content` class (`publishes no `, policy `allow`) in `skip-classes.tsv` already matches.
   Why this shape and not a new class: the situation is exactly what `root-content` means ("the
   supplied root publishes no such vector group"), so adding a class would duplicate vocabulary
   and force a self-test change for no new meaning. Why not drop the skip: the test is registered
   in the ledger as required on linux,darwin,windows with `skip_allowed_on linux,darwin,windows`
   class `root-content`, i.e. the gate's documented "run it whenever this runner and this root
   can" shape: the pinned rc.9 root records a tolerated, named skip in `skips-observed.tsv`; a
   candidate root that publishes the vector must observe it passing, and any other skip reason is
   fatal (proved by the mutant below).
3. New ledger rows (ledger-consistency treats unlisted cases as merely unlisted, not a gap; rows
   were added so the byte-exact behaviour is required by name on every GOOS):
   `internal/gitops TestExtractReproducesByteExactVector`, `TestExtractIgnoresWorkingTreeConversion`,
   `TestExtractRefusesSubmodules`, all `linux,darwin,windows | - | -`. The Windows claim is
   proved by the hosted `Test (windows-latest)` job (see CI section); locally only darwin ran.

## Local gate reproduction (exact commands, real exit codes)

```
bash .github/ci/ledger-consistency.sh .temp/rework1/ledger        -> exit 0  (86 rows checked across linux darwin windows; ledger-consistency: ok)
bash .github/ci/gate-selftest.sh                                   -> exit 0  (gate-selftest: 81 passed, 0 failed)
gofmt -l cmd internal                                              -> exit 0, no output
go build ./...                                                     -> exit 0
go vet ./internal/gitops ./internal/interop                        -> exit 0
go test -count=1 -race -timeout 30m ./internal/gitops ./internal/snapshot ./internal/closure ./internal/interop -> exit 0
   ok internal/gitops 18.577s; internal/snapshot 8.557s; internal/closure 51.068s; internal/interop 8.207s
```

Pinned root materialised as a detached curator-spec worktree at SPEC_PIN
`0ed5c691e9208eea52f21db2fc05e226ce3516fd` (`.temp/rework1/spec-pinned`); it has no
`vectors/snapshot-acquisition.json`.

Platform-case gate per GOOS over a `go test -json` stream of `./internal/gitops ./internal/interop`
under the pinned root (`CI_GATE_GOOS=linux|darwin|windows bash .github/ci/platform-case-gate.sh
.temp/rework1/focused.json <evidence>`; overall exit 1 on each because the other 80 ledger rows
never ran in a two-package stream; the rows this change owns):

```
ok    internal/gitops :: TestArchiveRejectsLinks
ok    internal/gitops :: TestExtractReproducesByteExactVector
ok    internal/gitops :: TestExtractIgnoresWorkingTreeConversion
ok    internal/gitops :: TestExtractRefusesSubmodules
ok    internal/interop :: TestGoldenContextCopy
tol   internal/interop :: TestConformanceSnapshotAcquisition (tolerated skip: root-content)
skips-observed.tsv: internal/interop  TestConformanceSnapshotAcquisition  root-content  tolerated-by-ledger  conformance root …/spec-pinned/conformance/v1 publishes no vectors/snapshot-acquisition.json (pre-environments suite; root-content)
```
Identical on linux, darwin and windows GOOS overrides.

Negative check on the gate: the same stream with the skip text mutated back to the old
"has no vectors/…" wording fails by name:
```
FAIL  ledger case skipped for the wrong reason on linux: internal/interop :: TestConformanceSnapshotAcquisition
      the ledger tolerates a root-content skip here; this one classified as UNCLASSIFIED
```

Candidate root (`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
main ec695ba): `go test -count=1 -v -run TestConformanceSnapshotAcquisition ./internal/interop` ->
exit 0, subtests `byte-exact-snapshot/autocrlf=true` and `autocrlf=false` PASS.

Full CI-shaped gate, exactly as ci.yml runs it, on darwin with the pinned root:
`CURATOR_CONFORMANCE_ROOT=…/spec-pinned/conformance/v1 bash .github/ci/test-gate.sh .temp/rework1/test-evidence`
-> exit 0. suite-plan: served=61 deferred=0 excluded=0; stage served exit=0; go test overall exit=0;
platform-case gate: GOOS=darwin, 20 skips recorded (all classified), `platform-case gate: ok`;
`test-gate: go test exit=0, platform-case gate exit=0`. The gitops/interop rows in
`platform-cases.txt` read exactly as in the focused run above (four `ok`, one `tol root-content`).
This full run includes `cmd/curator`; it was re-executed here, not cited from the earlier 463 s run.

Not run locally: linux and windows builds (the GOOS overrides above exercise the gate's ledger
logic only, not the test binaries); the hosted jobs are the proof for those.

## Hosted CI on PR #58 (head 5abec244)

Pushed with a plain `git push origin feat/byte-exact-acquisition` (5beced46..5abec244, no force).
Run 33966680785 at 5abec244: every job green except `Test (windows-latest)`:

| Job | Result |
| --- | --- |
| Gate self-test (ubuntu/macos/windows) | pass |
| Interop conformance gate, Lint, Naming gate | pass |
| Test (ubuntu-latest), Test (macos-latest) | pass |
| Race (ubuntu-latest), Race (macos-latest) | pass |
| Test (windows-latest) | fail: `internal/gitops TestExtractPreservesExecutableBit` — `100755 entry lost the executable bit: -rw-rw-rw-` |

On that Windows job the ledger-consistency step and the platform-case gate itself both passed
(`platform-case gate: ok`, all four gitops rows `ok`, interop `tol root-content`); the only
failure was the go test stream (`test-gate: go test exit=1, platform-case gate exit=0`), i.e. a
real defect in the new test, not a gate finding. Windows synthesizes `-rw-rw-rw-` for every
regular file, so a 100755 blob can never show an execute bit there.

## Second commit: a46abc80

```
a46abc80 Assert the extraction executable bit on the unix runners only   (on top of 5abec244, no rewrite)
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
```

Files: `internal/gitops/byteexact_test.go` (windows skip with the reason
`Windows does not expose portable executable permission bits`, the `platform-control` pattern
`skip-classes.tsv` already carries and `internal/artifactpolicy` already uses),
`.github/ci/platform-cases.tsv` (new row `internal/gitops TestExtractPreservesExecutableBit
linux,darwin | windows | platform-control`). No production code touched.

Local reproduction (exact commands, real exit codes):

```
go build ./...                                                     -> exit 0
go vet ./internal/gitops ./internal/interop                        -> exit 0
gofmt -l internal/gitops                                           -> exit 0, no output
bash .github/ci/ledger-consistency.sh .temp/rework1/ledger2        -> exit 0 (87 rows checked; ledger-consistency: ok)
bash .github/ci/gate-selftest.sh                                   -> exit 0 (gate-selftest: 81 passed, 0 failed)
CURATOR_CONFORMANCE_ROOT=<pinned root> go test -count=1 -race -timeout 30m -json ./internal/gitops ./internal/interop -> exit 0
CI_GATE_GOOS=linux|darwin|windows bash .github/ci/platform-case-gate.sh .temp/rework1/focused2.json <evidence>
   -> exit 1 on each (the other ledger rows are absent from a two-package stream, as before);
      linux and darwin print `ok internal/gitops :: TestExtractPreservesExecutableBit`; windows prints
      nothing for it (not required there).
```

Skip-tolerance proof: the same stream with the case's `pass` event rewritten to a `skip` carrying
the Windows reason text (`.temp/rework1/skipmutant.json`):

```
CI_GATE_GOOS=windows -> skips-observed.tsv: internal/gitops  TestExtractPreservesExecutableBit  platform-control  tolerated-by-ledger  Windows does not expose portable executable permission bits
CI_GATE_GOOS=linux   -> FAIL  ledger case skipped where the ledger does not tolerate it (linux): internal/gitops :: TestExtractPreservesExecutableBit
                        FAIL  required case skipped on linux: internal/gitops :: TestExtractPreservesExecutableBit
```

## Hosted CI run 33968713699 at a46abc80

`Test (windows-latest)`: pass. `test-gate: go test exit=0, platform-case gate exit=0`; ledger
consistency prints `ok internal/gitops :: TestExtractPreservesExecutableBit [must=linux,darwin
skip=windows]`; the evidence artifact's `skips-observed.tsv` records
`internal/gitops TestExtractPreservesExecutableBit platform-control tolerated-by-ledger` and
the interop `root-content` skip; zero `fail` events in `go-test.json`.

`Race (macos-latest)`: fail on the first attempt — `internal/install
TestStrictRegistryPolicyFailsUnknown`: `registry test-reg snapshot timestamp is too far in the
future`. That check (`internal/registry/snapshot.go:158`, `parsed.CreatedAt.After(now.Add(clockSkew))`)
compares a fixture stamped with `time.Now()` against the wall clock; it touches nothing this
branch changes (the branch's non-test diff is `internal/gitops`, one call each in
`internal/snapshot` and `internal/closure`), and the same job passed at 5abec244. Classified as
a pre-existing timing flake; the failed job was rerun (`gh run rerun 33968713699 --failed`).
Every other job on the run passed on the first attempt. The adapter suites (Cargo/Swift/pnpm/yarn)
did not fail on this run.

Final state of run 33968713699 (after `gh run rerun --failed`; the platform re-executed every
job, new job ids): all twelve required jobs pass — Gate self-test ×3, Interop conformance gate,
Lint, Naming gate, Test (ubuntu 3m9s / macos 10m18s / windows 32m7s), Race (ubuntu 9m18s /
macos 17m55s). `Candidate suite` is `skipping` by design on a PR run. `gh pr checks 58`: no
failing check.

## Not verified

- linux and windows test binaries were not run locally; the hosted jobs are the proof.
- The `internal/install` timing flake was not root-caused beyond locating the clock-skew check;
  it is outside this task's scope and is recorded in the board logbook entry.
