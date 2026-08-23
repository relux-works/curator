# Reviewer verdict for TASK-260811-1u42b9

Verdict: **accepted -> done**

## Scope and goal evidence

- Reviewer run: `RUN-260822-6f1d09`.
- `task-board spawn goal RUN-260822-6f1d09`: no active goal; the run is not goal-bound.
- No run directives were recorded at the review checkpoints.
- Reviewed rework outcome: `TASK-260811-1u42b9_rework-evidence_RUN-260822-3bb998.md`.
- Review was read-only; no product or test code was modified.

## Acceptance findings

1. Portable assurance remains the functional default. The npm adapter derives a deterministic private cache only from admitted tarballs, runs the shared manager process runner with frozen `npm ci --offline --ignore-scripts`, reconciles the lock and complete installed package contents, and invokes Node without a verified provider.
2. Portable evidence is honest. Its assurance binding advertises only immutable-input recheck, immediate toolchain recheck, and declared-output verification; the audit records `network=not-observed`, and lossless-only resolver/cache/lifecycle/process/read/write counters are absent rather than defaulted to observed zero.
3. Verified assurance uses the common nonce-bound provider negotiation and retained immutable `AssuredOperation` before cache derivation or process start. Missing, incomplete, incompatible, cross-mode, capability-drift, receipt-drift, and provider-identity-drift cases remain zero-start failures.
4. npm cache, install, and invocation operations are authorized through exact C0/C5 bindings and canonical `closureexec.DerivationPermit`/executor-issued receipt chains. Concrete argv, cwd, environment, admitted mounts, typed work copies, read/write roots, tool node, and process set are checked before start.
5. The prior ambient executable edge is removed. npm operations execute the exact C0-bound Node binary with the exact fingerprinted `npm-cli.js` entry point as argv[0]; the closed environment contains no `PATH`, and unbound shebang/PATH substitutions fail before process start. The positive portable and verified real-operation vectors observe the actual Node launch boundary and were not skipped.
6. Raw tarballs are SHA-512 SRI and Curator-digest bound, recursively admitted, and metadata-reconciled. Materialized external packages are re-admitted and compared file-for-file, including executable mode, against extraction evidence from the admitted tarballs. Lifecycle scripts, bundled trees, implicit `binding.gyp`/node-gyp, native addons, Wasm, V8 cache, opaque, renamed, nested, and substituted payloads fail closed.
7. Executable npm-wrapped S03, S04, and S08 coverage is present. S03/S08 perform two real offline replays with a poisoned inaccessible ambient cache and compare the complete materialized inventory and selected package graph while retaining honest portable `network=not-observed`; S04 reports `closure_network_attempted` from verified provider observation after one declared action start and issues neither receipt nor cache publication.
8. Supported package-lock and npm-shrinkwrap v2/v3 parsing, root/workspace reconciliation, selected/pruned edges, immutable locators, integrity, stale-lock handling, private-cache receipts, exact installed graph reconciliation, and N01-N13 coverage match the task acceptance criteria and project architecture.

## Fresh validation

| Gate | Exit | Evidence |
| --- | ---: | --- |
| Targeted real S03/S04/S08, direct Node-launched npm, verified real launch, and zero-start vectors | 0 | `targeted-verbose-01.log`; real npm and verified integration tests ran without skips |
| Affected focused suite | 0 | `focused-01.log`; artifactpolicy `41.874s`, closureexec `5.473s`, nodesource `2.023s`, npmsource `13.307s` |
| npm race | 0 | `race-01.log`; `35.742s` |
| npm coverage | 0 | `coverage-01.log`; `80.4%` statements |
| Vet | 0 | `vet-01.log` |
| Lint | 0 | `lint-01.log`; `0 issues` |
| Build | 0 | `build-01.log` |
| Diff whitespace | 0 | `diff-check-01.log` |
| Board validation | 0 | `board-validate-01.log`; board valid |
| Uncached repository suite | 0 | `repository-suite-01.log`; cmd/curator `386.789s`, artifactpolicy `156.474s`, closureexec `24.265s`, npmsource `100.417s`, rustsource `149.582s` |

The project has `version_control.confirm=false`, so this accepted Task transition does not require `commit_ack`. As a reviewer-archetype run, this verdict supplies no `commit_ack`.
