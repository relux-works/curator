# TASK-260720-1pvfj5 — final accepted-composite integration results

Date: 2026-07-30  
Role: developer  
Composite: `.temp/TASK-260720-1pvfj5/rework/composite`  
Result: ready for independent review

## Integrated inputs

The starting composite was proved byte-identical to the previously accepted
372-entry rework manifest before either patch was applied (`cmp` exit 0). That
manifest contains the accepted TASK-260720-jrrgw9 product plus the unchanged
CI/quality overlay.

The two independently accepted blocker patches were applied verbatim:

| Input | SHA-256 | Apply check | Apply | Reverse check |
| --- | --- | ---: | ---: | ---: |
| `BUG-260729-1o0m8f_lint-fix.patch` | `8a07c0b239548235aea7dfa05fdb1d1cb2926971d4444d3435a9e6f8da368062` | 0 | 0 | 0 |
| `BUG-260729-r0fe02_patch.diff` | `462f1ff0326f74540eeb2815cc80542c55f47b35c6b1baef17b80b8815709c28` | 0 | 0 | 0 |

All five lint-patch files compare byte-for-byte with the accepted
BUG-260729-1o0m8f worktree, and both cancellation-patch files compare
byte-for-byte with the accepted BUG-260729-r0fe02 worktree. Both comparison
commands exited 0.

No patch, workflow, vector, timeout, suppression policy, fixture, or unrelated
product byte was re-derived.

## Exact delta proof

The accepted composite has 372 total entries and 356 product entries. The
integrated composite has 374 total entries and 358 product entries. The two
additional entries are the accepted focused test files. A task-local manifest
verifier exited 0 and proved:

- exactly seven changed product paths;
- five modified paths, two added paths, zero removed paths;
- every other accepted product entry is byte-identical;
- every one of the 16 CI/quality paths is byte-identical to the accepted
  rework overlay.

The seven paths are:

1. `internal/godriver/builddriver_positive_conformance_test.go` (modified)
2. `internal/godriver/fingerprint.go` (modified)
3. `internal/godriver/fingerprint_equivalence_test.go` (modified)
4. `internal/protocoljson/ccj.go` (modified)
5. `internal/protocoljson/ccj_test.go` (added)
6. `internal/transaction/journal.go` (modified)
7. `internal/transaction/journal_order_test.go` (added)

Post-validation manifest generation again produced 374 entries. It compares
byte-for-byte with the immediate post-application manifest (`cmp` exit 0), so
the validation phase changed no composite byte.

Manifest SHA-256 values:

- accepted 372-entry composite:
  `527126c2186aabbfa9f917c9aca024111ad5ed9e377873e997a80a961f5955d6`
- integrated 374-entry composite:
  `aa9f8ece325dd4b81435657dd07b7a7c83e397822750974c9e0554f1d8778b6b`

## Immutable protocol inputs

The committed workflow contains exactly one `SPEC_PIN`, and the pin remains:

`00b1688a9b2457ca397a0bb550acf47cad8ee967`

The released-pin checkout HEAD matches that value and its `manifest.json`
digest remains:

`7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e`

The explicitly supplied non-default rc.5 candidate root is:

`.temp/TASK-260729-3nx97g/worktree/conformance/v1`

Its identity was recorded with caller-supplied expected digests and stamped
candidate-only, not a release and not a conformance claim:

- protocol: `1.0.0-rc.5`
- manifest:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- whole tree:
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`
- files: 448
- build-driver vector:
  `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
- manager lifecycle vector:
  `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`
- external repository lifecycle vector:
  `175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072`

The root is a pre-materialised accepted candidate and has no published revision
claim. Its evidence says that explicitly.

## Validation ledger

Each gate below ran as a standalone process with its real exit preserved. Heavy
Go suites ran sequentially with an empty-process barrier before each one.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| candidate identity record with expected manifest/tree digests | 0 | `candidate-record.log`, `evidence/candidate/` |
| gofmt check over `cmd internal` | 0 | `gofmt.log` |
| no broad suppression | 0 | `no-broad-suppression.log` |
| golangci-lint v2.12.2, fresh task-local cache | 0, 0 issues | `golangci-lint-v2.12.2.log` |
| `go vet ./...` | 0 | `go-vet.log` |
| `go build ./...` | 0 | `go-build.log` |
| deterministic godriver cancellation and precedence | 0 | `godriver-cancellation.log` |
| released-pin `.github/ci/test-gate.sh` | 0 | `default-test-gate.log`, `evidence/default/` |
| explicit rc.5 full-root `.github/ci/test-gate.sh` | 0 | `candidate-test-gate.log`, `evidence/candidate-test/` |
| one full `-race` `.github/ci/test-gate.sh` | 0 | `race-test-gate.log`, `evidence/race/` |
| `git diff --check` | 0 | no output |
| staged-index absence | 0 | `git diff --cached --quiet` |
| post-run manifest equality | 0 | `composite-final-manifest.txt` |
| exact seven-path manifest delta | 0 | `manifest-delta-proof.log` |
| one exact released pin assertion | 0 | count 1, expected value |

Released-pin non-race lane: 33 served, 7 explicitly deferred, 0 excluded;
served test 0, deferred test 0, platform-case gate 0.

Explicit rc.5 lane: 40 served, 0 deferred, 0 excluded under
`CI_REQUIRE_FULL_ROOT=1`; test 0, platform-case gate 0.

Full race lane: exactly one invocation, against the released pin; 33 served,
7 explicitly deferred, 0 excluded; both test stages 0 and platform-case gate 0.
An exact search for `WARNING: DATA RACE`, `DATA RACE`, or a `FAIL` token returned
exit 1 with no matches. That non-zero is the expected no-match result, not a
passing command disguised as exit 0.

An exact search for rc.4 wording in `README.md`, `.github/workflows/ci.yml`, and
`Makefile` also returned exit 1 with no matches. The stale board checklist item
requiring an rc.4 pin remains intentionally unchecked and is superseded by the
recorded released-pin/candidate boundary.

## Reused unchanged gate evidence

The gate scripts and tables are byte-identical to the accepted rework overlay;
the manifest proof reports zero changed overlay paths. Therefore the prior
task-scoped results are reused, as required:

- gate self-test: exit 0, 70 passed / 0 failed;
- ledger consistency: exit 0, 49 rows across linux/darwin/windows.

They were not rerun because the final integration instruction requires reuse
when script bytes are identical.

The workflow still runs ubuntu, macOS, and Windows test/candidate matrices;
Linux and macOS race jobs; Windows DACL/reparse/`.cmd` cases; and Unix
permission, no-follow, readonly-source, resource-policy, and executable cases.
This integration run was native Darwin only; hosted Linux and Windows execution
occurs in CI and is not misrepresented as local execution.

## Diagnostic command failures reported truthfully

- The initial aggregate tool-readiness command exited 127 because no unqualified
  `golangci-lint` exists on PATH. The measured lint gate used the preserved,
  CI-pinned v2.12.2 binary, whose version command and lint run both exited 0.
- Two first byte-compare helper loops each exited 127 because the loop variable
  was named `path`, which is a special zsh variable and replaced `PATH`. The
  corrected loops used `file_path` plus `/usr/bin/cmp`; both exited 0 and
  covered all seven files.
- `diff -rq` between the accepted jrrgw9 worktree and the accepted combined
  blocker worktree exited 1 as expected because exactly the seven paths above
  differ. Its only other output was symmetric directory-loop warnings for the
  existing skill symlinks.

No failed gate was retried. The race gate ran exactly once.

## Handoff to TASK-260720-38l1sy

Audit the unchanged released pin
`00b1688a9b2457ca397a0bb550acf47cad8ee967` against the default-lane evidence.
The candidate evidence is non-default and candidate-only:

- root: `.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- manifest:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- tree:
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`
- full-root plan: 40 served, 0 deferred, 0 excluded
- candidate test gate: exit 0

This is not evidence that rc.5 is published and must not be used to advance the
committed pin before the release owner qualifies it.

## Scope and repository state

No file was staged, committed, stashed, published, or pinned. No timeout,
fixture, suppression, workflow, Makefile, README, protocol vector, or unrelated
product byte was changed during final integration. The only composite delta
beyond the already accepted product and CI overlay is the exact seven-path
blocker patch set above.
