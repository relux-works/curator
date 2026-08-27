# TASK-260823-czs1cx — reviewer verdict, cycle 2 (RUN-260823-ee0acd)

**VERDICT: ACCEPTED.** Both owned Windows cases are green against candidate
`edd0721`, the fix is on the guilty side, and it is load-bearing — all of which
this review reproduced on a native Windows host rather than taking on trust.

Supersedes the cycle-1 verdict (RUN-260823-3e7189, changes requested). Every
item that verdict blocked on is closed.

## What was reviewed

`main` @ `2671743` (merge of PR #31 / signed commit `695c041`) against candidate
`curator-spec@edd07210d4f3db34fd60238cb14b90f837de03cb`.

The merge introduces exactly the reviewed content: `git diff d76fe4d 2671743`
differs from `git diff 695c041^ 695c041` only in hunk offsets and blob hashes
caused by PR #30 landing in between. No drift.

## AC part 1 — both cases green on Windows

Reproduced by the reviewer on `DESKTOP-3PBO632`, go1.25.5 windows/amd64, against
a locally materialized `edd0721` conformance root:

```
--- PASS: TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment (2.52s)
--- PASS: TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link (0.00s)
--- PASS: TestFingerprintImplementationMatchesRC4ToolchainVector (0.01s)
ok  github.com/relux-works/curator/internal/godriver  4.555s   FOCUSED_EXIT=0
```

Full package on the same host, same root: `ok ... 123.016s`, `FULL_GODRIVER_EXIT=0`.
Nothing else in `internal/godriver` regressed on the platform the change targets.

CI agrees for the right reason. Every lane on `695c041` is green, and the
**Candidate suite (windows-latest)** job (run 32645833559) verifies in-job that
it checked out `CANDIDATE_REF: edd07210d4f3db34fd60238cb14b90f837de03cb` and that
`candidate-suite: manifest digest matches the supplied expectation` for
`803918bf...`, with `suite-plan: served=42 deferred=0 excluded=0` under
`CI_REQUIRE_FULL_ROOT=1`. This is the lane that was red on this case last cycle.

## AC part 2 — candidate identity recorded

No regeneration was needed this cycle and none was done. The reviewer
recomputed `shasum -a 256` of `conformance/v1/manifest.json` at `edd0721`
independently: `803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`.
TASK-260822-c0rxj7 still carries that identity (`edd0721` / manifest `803918bf`
/ tree `9d5a10b6` / 692 files) with the byte-identical double-regeneration
proof, and the superseded `859727b` identity is still recorded unrewritten.

## The fix is on the guilty side, and it is load-bearing

Verified, not accepted on the producer's word. Neutralizing **only**
`protocolLinkTarget` on Windows — no other edit — and rerunning on the native
host:

| Windows run vs `edd0721` | `unsorted-…-internal-link` | pinned-digest test |
| --- | --- | --- |
| normalizer neutralized | FAIL — `..\bin\go` vs `../bin/go` | FAIL — `sha256:e9b9a60b…` |
| restored | PASS | PASS (`RESTORED_EXIT=0`) |

The forked digest is exactly the `e9b9a60b…` the producer reported, against the
published `baf7c5f3…`. The vector was right; the implementation was hashing a
host property into a digest that is supposed to name a tree. Normalizing in
`fingerprint.go` before validation and hashing puts the fix where the bug was.

## Adversarial checks the reviewer ran on the normalization point

The security checks downstream of `Readlink` now see the slash form, so each was
re-examined and each still holds on Windows:

- `filepath.VolumeName` still catches `C:/…` after `ToSlash`.
- `//?/C:/Windows` and `/foo` are caught by the existing `HasPrefix(target, "/")`;
  the `\` -prefix arm becomes redundant on Windows but is still correct for the
  unix build where `protocolLinkTarget` is the identity.
- `utf8.ValidString` cannot be weakened by `ToSlash`: `0x5C` never occurs as a
  UTF-8 continuation byte, so the substitution cannot repair invalid input.
- Confirmed empirically on Windows — `escaping-toolchain-link`,
  `absolute-toolchain-link`, `duplicate-toolchain-path` and
  `invalid-unicode-toolchain-path` all PASS.

Build tags rather than a blanket `ToSlash` is the right call and the comments say
why: a backslash is an ordinary filename character on unix.

The equivalence gate's reference traversal normalizes at the same point, which is
required — normalizing only one of the two traversals would have turned that gate
red on Windows.

`record.link` has exactly two uses (`fingerprint.go:85` hashing, `:226` assignment),
so nothing downstream consumes the host-native form.

## The `Fatalf` → `Skipf` nit is closed, and it is safe

`fixedEnvironmentForHost` now skips a host the suite is silent about. That cannot
hide a dropped case: the candidate's `validate_fixed_environment_cases` pins the
host set to exactly `{darwin-arm64, linux-amd64, windows-amd64}` and raises
`fixed environment host coverage changed` otherwise — and those three are exactly
the CI runners. `TestFixedEnvironmentForHostSelectsNativeCase` pins the selection
logic itself.

## Gates

| Gate | Host | Exit |
| --- | --- | ---: |
| focused godriver vs `edd0721` | macOS | 0 |
| focused godriver vs `edd0721` | native Windows | 0 |
| full `internal/godriver` vs `edd0721` | native Windows | 0 |
| `go vet ./...` | GOOS=windows | 0 |
| `golangci-lint run ./...` v2.12.2 | darwin | 0 — "0 issues." |
| all CI lanes on `695c041` incl. Candidate suite ×3 | GitHub | success |

## Findings — non-blocking, do not reopen this task for them

1. **The claimed regression guard is real but unenforced.** The producer's note
   says the newly cross-platform `TestFingerprintImplementationMatchesRC4ToolchainVector`
   closes the gap on the ordinary `Test (windows-latest)` lane. On the reviewer's
   native Windows host it does run and pass with no conformance root
   (`DEFAULT_LANE_EXIT=0`), so the claim is true there. It is not *guaranteed*
   on the GitHub runner: the test is absent from `.github/ci/platform-cases.tsv`,
   and its `symlink unavailable: %v` reason matches `skip-classes.tsv:62`
   (`host-capability`, `allow` in every lane), so a runner that forbids
   unprivileged symlink creation would skip it green. The repo's own ledger
   (`platform-cases.tsv:124-125`) states that the Windows runner may do exactly
   that. Recommended follow-up, cheap: add a Tier-1 row for this case with
   `must_run_on` including `windows` and the `host-capability` tolerance, so a
   rename or deletion fails by name instead of shrinking the suite.

2. **Pre-existing `GOOS=windows` lint debt is unguarded.** `GOOS=windows
   golangci-lint run ./...` reports 10 findings (5 errcheck, 4 gosec, 1 revive)
   in `internal/buildcache/protection_windows.go`,
   `internal/buildrepo/protection_windows.go`,
   `internal/managerlock/identity_windows.go`(+test),
   `internal/transaction/durability_windows.go` and
   `internal/godriver/controls_windows.go`. The reviewer reproduced all 10 —
   the producer's count and classification were accurate, and none are in a file
   this change touches. Worth noting that the CI `Lint` job runs on
   `ubuntu-latest` only, so no gate covers Windows-tagged code at all.

## Artifacts

- `TASK-260823-czs1cx_reviewer-verification-cycle2.log` — the reviewer's own runs
  (macOS, native Windows, neutralized/restored, full package, vet, lint).
- `TASK-260823-czs1cx_review-verdict-cycle2.md` — this verdict.
