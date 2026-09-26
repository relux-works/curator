# CI Gates and Tooling

Every gate below is a script under `.github/ci/`, called directly by [../.github/workflows/ci.yml](../.github/workflows/ci.yml) and mirrored by a `make` target for local use. CI calls the scripts directly because `make` is not guaranteed on Windows runners. Each gate writes its raw stream and report under `EVIDENCE` (default `.temp/ci-evidence/`). CI uploads these artifacts per runner, allowing verification against the run that produced them.

| Tool | What it gates | How to run it | Where its output goes |
| --- | --- | --- | --- |
| `test-gate.sh` | plans the run from the supplied conformance root, executes it, then enforces the platform-case ledger; every status is fatal | `make ci-test` | `$EVIDENCE/test/` (including `go-test*.json`, `suite-plan.txt`, `platform-cases.txt`, `skips-observed.tsv`) |
| `test-gate.sh` (`-race`) | the same gate under the race detector | `make race` | `$EVIDENCE/race/` |
| `suite-plan.sh` | decides, from the root alone, which packages it serves, which it cannot, and which this platform does not qualify | called by `test-gate.sh` | `$EVIDENCE/*/suite-plan.txt`, `plan-*.txt` |
| `platform-case-gate.sh` | requires every case [`platform-cases.tsv`](../.github/ci/platform-cases.tsv) names on this runner, and classifies every skip against [`skip-classes.tsv`](../.github/ci/skip-classes.tsv) | called by `test-gate.sh` | `$EVIDENCE/*/platform-cases.txt`, `skips-observed.tsv` |
| `ledger-consistency.sh` | proves each ledger row against the real per-GOOS builds via `go list` without requiring a runner | `make ledger-check` | `$EVIDENCE/ledger/ledger-consistency.txt` |
| `excluded-packages.sh` | single implementation resolving which packages a platform does not execute and recording the authorizing task | called by `suite-plan.sh` and `ledger-consistency.sh` | stdout (TSV) |
| `candidate-suite.sh` | rejects non-immutable candidate revisions and records candidate root identities as candidate-only evidence | `make candidate-verify-ref CANDIDATE_REF=...` or `make candidate-record CANDIDATE_ROOT=...` | `$EVIDENCE/candidate/candidate-suite-identity.txt` |
| `toolchain-identity.sh` | asserts the resolved Go toolchain matches `go.mod` exactly, with `GOTOOLCHAIN=local` and `GOENV=off` read back | run by every Go-consuming job | job log |
| `no-broad-suppression.sh` | rejects bare `//nolint`, bare `//#nosec`, production-path lint exclusions, wholesale disabling, and unrecorded `gosec` exclusions | `make no-broad-suppression` | job log |
| `gate-selftest.sh` | drives every gate above against synthetic inputs and asserts real exit codes for negative cases | `make gate-selftest` | job log |
| `python_protocol_golden.py` | independently decodes the closed Node/Python fixture schema, derives canonical package/capture/binding/active/diagnostic records, and rejects cross-target reuse | `python3 internal/nodesource/testdata/python_protocol_golden.py` | stdout; redirect task evidence under `.temp/` when needed |
| `npm` profile harness | exercises exact-tarball private-cache derivation, poisoned-ambient replay, direct Node-launched npm, scripts-disabled offline `npm ci`, and the real verified launch boundary for `npm-source-v1` | `go test -count=1 ./internal/npmsource -run 'Test(N01RealNPMCIUsesOnlyDerivedPrivateCache|VerifiedProviderObservesRealNodeLaunchedNPMBoundary)$' -v` | Go test output; task evidence under `.temp/<TASK-ID>/` |
| `pnpm` 10.33.0 profile harness | validates lock/importer/snapshot/peer/patch/target closure, full installed-layout ownership, protected private-store derivation, and real frozen offline replay without ambient pnpm state | `npm install --prefix .temp/pnpm-10.33.0 --no-save --ignore-scripts pnpm@10.33.0`; then `PATH="$PWD/.temp/pnpm-10.33.0/node_modules/.bin:$PATH" go test -count=1 ./internal/pnpmsource` | Task-local pnpm under `.temp/pnpm-10.33.0/`; Go test output may be stored under `.temp/<TASK-ID>/` |
| Yarn Classic 1.22.22 profile harness | validates closed lock/workspace/config authority, private source mirror derivation, empty-cache frozen replay, exact staged Node/Yarn launch, lifecycle/native rejection, and poisoned ambient config/cache isolation | `go test -count=1 ./internal/yarnclassicsource` | Go test output and task evidence may be stored under `.temp/<TASK-ID>/` |
| Modern Yarn 4.9.2 profile harness | validates lock v8, exact built-in plugins, `.yarnrc.yml`, patches, cache key/compression/checksums, linker/conditions, peer-context virtualization and PnP aliases, deterministic private ZIP cache, preseed-state isolation, lifecycle/native rejection, and immutable network-disabled skip-build replay | `npm install --prefix .temp/yarn-4.9.2 --no-save --ignore-scripts @yarnpkg/cli-dist@4.9.2`; then `CURATOR_TEST_YARN_MODERN_JS="$PWD/.temp/yarn-4.9.2/node_modules/@yarnpkg/cli-dist/bin/yarn.js" go test -count=1 ./internal/yarnmodernsource` | Task-local Yarn under `.temp/yarn-4.9.2/`; Go test output may be stored under `.temp/<TASK-ID>/` |
| Swift / SwiftPM | validates `swiftpm-source-v1` lock, exact pin, tree-intake, manifest permit, kind-preserving mirror, selection-neutral capture, exact binding, extension/binary rejection, and offline replay contracts; runs real protected `swift package dump-package` and forced-lock `show-dependencies` mirror replay when Swift is installed | `go test -count=1 ./internal/swiftpmsource` | Go test output; task evidence may be stored under `.temp/<TASK-ID>/` |
| Cross-adapter source-closure conformance | runs one normative semantic suite across Rust, npm, pnpm, Yarn Classic, modern Yarn, and SwiftPM/C-family; independently canonicalizes and hashes all 53 accepted CGP05/CGP10 records; drives the published rejection matrix; refuses an incomplete coverage matrix; and emits the committed protocol export for independent implementations | `go test -count=1 ./internal/crossconformance` | Go test output and `internal/crossconformance/testdata/cross-adapter-protocol-export.json`; task evidence may be stored under `.temp/<TASK-ID>/` |

Executing `make ci-test`, `make race`, and `make check-ci` requires setting `CURATOR_CONFORMANCE_ROOT` to a materialized `<curator-spec>/conformance/v1` directory. The Makefile targets refuse execution without this variable set, preventing unconfigured runs.

## Per-package Go timeout budget

`test-gate.sh` passes `GO_TEST_TIMEOUT` to every `go test` invocation. The script and local Makefile gate targets default to `60m`; all Ubuntu/macOS Test, Race, and candidate-suite lanes use `60m`, while Windows keeps `120m`. These remain Go per-package timeouts, so a package that hangs still fails when its deadline expires. The self-hosted macOS Test lane also uses `60m`.

The green jobs in hosted gate run [35709048693](https://github.com/relux-works/curator/actions/runs/35709048693) recorded these `internal/install` package times:

| Lane | Elapsed | Original 30m budget | Chosen 60m budget |
| --- | ---: | ---: | ---: |
| Test (ubuntu-latest) | 926.106s | 51.5% | 25.7% |
| Test (macos-latest) | 558.623s | 31.0% | 15.5% |
| Race (ubuntu-latest) | 1114.886s | 61.9% | 31.0% |
| Race (macos-latest) | 537.885s | 29.9% | 14.9% |

Later contention events show why 30m was insufficient: Test (ubuntu-latest) reached 1800.165s on run 35713206841, and Race (ubuntu-latest) reached 1800.100s on run 35994653101. The larger timeout preserves the bounded, fatal hang behavior while providing at least 2x headroom over each measured green baseline.

## Protocol-suite pin and verification

The module's immutable release pin ([`internal/buildrepo/release_pin.go`](../internal/buildrepo/release_pin.go), verified by `curator-spec-pin`) remains curator-spec `v1.0.0-rc.8`. The CI conformance pin is independent and follows curator-spec v1.0.0-rc.13 tag commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` (conformance/v1 manifest SHA-256 `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`); see `SPEC_PIN` in [../.github/workflows/ci.yml](../.github/workflows/ci.yml). This is a conformance root pin, not a release qualification. The released skillfile-sources-v1 corpus is vendored from the same tag commit; its source manifest SHA-256 is `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`.

The module's separate release-check suite remains pinned to curator-spec `v1.0.0-rc.8` at commit `f8c405aa3ad0a39d260c2ed93684e55c5a346359`. `curator-spec-pin` verifies the suite manifest SHA-256 `d14e3a16bb4a01ff282791f08e3aefa269210234f41072beae6fe59b642595a1` and release metadata SHA-256 `293f101d10665061aa049efa72141f9e3c5d608bbde300e882f6e3e095e31ede`, including the empty published implementation, platform, and conformance claim sets. Run it locally with `make verify-spec-pin SPEC_PIN=f8c405aa3ad0a39d260c2ed93684e55c5a346359 CURATOR_CONFORMANCE_ROOT=/path/to/curator-spec/conformance/v1`.

## The compiled-build platform carve-out

The `rc5-native-control-inventory-v1` vector defines native control records exclusively for macOS and Windows. The `go-v1` driver refuses compiled builds on any other host before starting a worker process. This carve-out applies at two granularities:

- **Whole package**: `internal/godriver` is not executed where the root qualification vector marks the platform excluded (such as Linux, noted under `until_task: TASK-260728-1skseh`). `test-gate.sh` executes `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` on that runner to assert the exclusion boundary.
- **Individual case**: Commands in `cmd/curator` that require completed compilation use `requireNativeControlInventoryPlatform` to read `godriver.InventoryPlatform`. Skips log the inventory reason, classified under `platform-control` in [`skip-classes.tsv`](../.github/ci/skip-classes.tsv). `TestCompiledInstallFollowsTheNativeControlInventoryExactly` runs on every runner to verify the boundary state. Covered hosts execute and publish protected cache entries; uncovered hosts report `build_execution_control_unavailable`, publish nothing, and return non-zero exit codes for `curator status --check`.

When the inventory gains a record for a platform, the guard stops skipping automatically. Only `must_run_on` in [`platform-cases.tsv`](../.github/ci/platform-cases.tsv) requires updating.

## Suite consumption, not suite presence

Publishing a family in a conformance root does not prove that a build reads it. Analysis showed that pinned implementation jobs returned exit 0 against a schema-8 suite while consuming none of it. Two tables resolve both aspects:

- **Presence**: [`root-artifacts.tsv`](../.github/ci/root-artifacts.tsv) declares root artifacts read by a package conformance test without guards. If a root stops publishing an artifact, the package defers testing. Setting `CI_REQUIRE_FULL_ROOT=1` makes deferrals fatal in candidate verification lanes.
- **Consumption**: [`platform-cases.tsv`](../.github/ci/platform-cases.tsv) specifies cases that read each artifact per runner. Renames, deletions, or unmatched `-run` filters trigger named failures rather than silent test suite reductions.

For schema 8, artifacts include `agent-skill-v8`, `csk-skill-v8`, `install-marker-v4`, `vectors/module-roots.json`, and `vectors/script-host-execution-policy.json`. These are consumed by `internal/skillspec`, `internal/marker`, `internal/moduleroots`, `internal/godriver`, and `internal/scriptpolicy`.

The committed protocol-suite pin is declared as `SPEC_PIN` in the workflow `env:` block. Candidate suites enter via the `candidate-conformance` workflow on explicit `workflow_dispatch` calls supplying a full 40-character revision or materialized root. That job sets `CI_REQUIRE_FULL_ROOT=1` to enforce complete package coverage, and emitted artifacts are stamped as candidate-only evidence, proving neither a published release nor a conformance claim.

## Published case accounting

Consumers that iterate a complete published case family use the shared case-coverage harness. [`conformance-case-counts.tsv`](../.github/ci/conformance-case-counts.tsv) pins the expected case count for each family at its selected corpus revision; each run classifies every published case as driven, known-gap, bound, or skipped and requires the tally to equal that count. [`conformance-gaps.tsv`](../.github/ci/conformance-gaps.tsv) is the only place a known gap may be declared. A row is an owed implementation, never an accepted deviation: an unlisted failure fails the test, a listed case that starts passing fails until its row is removed, and a row whose case disappears fails the count or row-presence check.

The released v1.0.0-rc.13 root publishes the complete `skillfile-dev-v2` and `skill-build-v1` schema lists, external-repository fixtures, and the newly promoted config, environment marker, security-posture and launch-fragment cases. Measured family tallies and owned gaps were rechecked against v1.0.0-rc.13 tag commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`; the interim-to-release conformance diff changes protocol_version labels and manifest digests, not case IDs, so the pinned counts and gap rows remain unchanged. Tests that select one named case from a broader published vector set remain targeted checks; they do not claim to account for the complete family.

The 28 manager-config overlay gaps are attributed to the unsupported `permissions`, `source_signers`, and `require_source_signers` members added by 0017/0018 and E1. Those members are present in the updated overlay cases, so `Load` can report either new blocker first; it does not indicate an overlay path-kind failure. The config regression test checks all 14 schema inputs and 14 vectors against their production results and checks both owners in each ledger row. The rc.13 script-worker root also publishes eight Windows executable-identity cases: the manager resolver drives all eight, with two known gaps where the current System32 exception cannot establish component-store link targets or platform ownership.

## Self-hosted runner prerequisites

The rose-air lane needs exactly one thing installed on the runner by hand:
rustup. See [self-hosted-runner-setup.md](self-hosted-runner-setup.md).
