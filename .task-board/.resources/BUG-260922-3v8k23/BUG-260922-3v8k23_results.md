# BUG-260922-3v8k23 results

## Change

Raised the Go per-package test deadline to 60m for Ubuntu and macOS Test, Race, and candidate-suite lanes. Windows remains at 120m. The self-hosted macOS Test lane is explicitly 60m. Updated `.github/ci/test-gate.sh` and the Makefile local gate defaults to 60m so the Make targets do not override the script with the former 30m value.

`.github/ci/gate-selftest.sh` now pins each of the four workflow test-gate lanes, including the Windows value in the candidate-suite matrix expression, and rejects narrowed timeout mutants. The synthetic hanging-package case drives the real `test-gate.sh` with a 2s package timeout and confirms that Go reports a timeout and the gate exits 1. Its temporary module uses the Go version reported by the active runner, so it does not depend on the repository toolchain being installed in the gate-selftest job.

No product source or `CHANGELOG.md` was changed. The release-prep entry is below.

## Baseline and selected budget

Hosted gate run [35709048693](https://github.com/relux-works/curator/actions/runs/35709048693) had all four relevant Ubuntu/macOS Test and Race jobs green. The overall run was red because the separate Windows Test job failed. Its `internal/install` elapsed times, recalculated against the chosen 60m budget:

| Lane | Elapsed | Old budget | Old budget used | Chosen budget | Chosen budget used | Headroom |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Test (ubuntu-latest) | 926.106s | 1800s | 51.5% | 3600s | 25.7% | 3.89x |
| Test (macos-latest) | 558.623s | 1800s | 31.0% | 3600s | 15.5% | 6.44x |
| Race (ubuntu-latest) | 1114.886s | 1800s | 61.9% | 3600s | 31.0% | 3.23x |
| Race (macos-latest) | 537.885s | 1800s | 29.9% | 3600s | 14.9% | 6.69x |

The table shows at least 2x headroom for every measured lane. This leaf has one green baseline run with all four relevant jobs; it does not claim two consecutive green post-change hosted runs. Later runs hit the old 30m deadline during runner contention: Test/ubuntu reached 1800.165s on run 35713206841, and Race/ubuntu reached 1800.100s on run 35994653101.

## Validation

| Command/evidence | Exit | Result |
| --- | ---: | --- |
| `bash .github/ci/gate-selftest.sh` (latest full local rerun) | 0 | 212 passed, 0 failed. All four workflow timeout expressions and five drift mutants are covered. |
| Synthetic hanging package through `test-gate.sh` (self-test row) | 1 (expected red) | The gate forwarded `-timeout 2s`; Go emitted `panic: test timed out after 2s`; the gate exited 1 in 4s including compilation. |
| `bash .github/ci/ledger-consistency.sh /tmp/BUG-260922-3v8k23-ledger-evidence` | 0 | 355 rows checked across linux, darwin, and windows. Evidence stayed outside the worktree. |
| `golangci-lint run` | 0 | 0 issues. |
| `bash -n .github/ci/gate-selftest.sh` | 0 | Shell syntax clean. |
| `bash -n .github/ci/test-gate.sh` | 0 | Shell syntax clean. |
| `ruby -e 'require "yaml"; YAML.parse_file(".github/workflows/ci.yml"); puts "ci.yml YAML syntax ok"'` | 0 | Workflow YAML parsed. |
| `git diff --check` | 0 | No whitespace errors. |

The previous Change Request revision's remote validation log is attached as `BUG-260922-3v8k23_change-request_rev1-validation.log`. Gate run 36020770217 passed the Ubuntu/macOS Test and Race jobs, lint, and interop, but its Ubuntu and Windows gate-selftest jobs failed because the synthetic fixture produced no timeout event. The runner log also showed Go-dependent suite-plan checks skipping on those self-test runners. The likely cause was the fixture requiring the repository's Go 1.25.5 directive while those jobs use `GOTOOLCHAIN=local` and do not install the repository toolchain; the exact startup diagnostic was not preserved by that revision. The fixture now adopts `go env GOVERSION`, and its failure report includes the merged Go stream and gate output. The latest local full self-test passes with this change. The next task-board handoff runs the configured remote validation on the new candidate.

`actionlint` and `shellcheck` are not installed locally; YAML parsing and Bash syntax checks passed. The full landing suite was not run manually; task-board handoff runs the configured validation.

## CHANGELOG entry (for release prep)

Increase the Go per-package test timeout to 60 minutes on Ubuntu and macOS CI lanes while retaining 120 minutes on Windows.
