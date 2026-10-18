# BUG-261018-bisu6n developer evidence

## Candidate and scope

- Base checkpoint: `b9caa20901ee859eec20e649940693033e6297d4` (0 commits behind local main at intake).
- Signed detached gate snapshot: `102a56dea6d80480bbd9c928138bd0fad66e5dee`; signature verified `G` before push. Story HEAD/index were not committed or switched.
- Hosted run: https://github.com/relux-works/curator/actions/runs/39153364696 (success).
- No surface table was supplied: this is a brief gap. The coverage table below maps the seven required behavior/validation rows; no unseen catalog coverage is claimed.
- No `repeat-of:` class was supplied in the brief; the leaf adds a named regression witness and narrowing mutant for the readonly-temp class. Scope includes the four manager test fixtures, CI workflow tests, the shared script and the runner setup note. No product behavior, root README, CHANGELOG, host hook, labels or repository variables changed.

## Required-item locations

| Required item | Source locations |
| --- | --- |
| Early restore and real pnpm call sites | `internal/pnpmsource/conformance_test.go:1127,1735,1839,1857` |
| Neighbor helpers and fixture regressions | `internal/npmsource/conformance_test.go:1728`, `internal/yarnclassicsource/conformance_test.go:1294`, `internal/yarnmodernsource/conformance_test.go:2120`; each neighboring `tmp_hygiene_test.go:11` |
| Every Linux-routable final step | `.github/workflows/ci.yml:141,303,551,618,651,695,723` |
| Shared chmod-only script and setup note | `scripts/restore-runner-tmp.sh:1`; `docs/self-hosted-runner-setup.md:167` |
| Pass/Fatal/capture-Fatal witness and cleanup plant control | `internal/pnpmsource/tmp_hygiene_test.go:15,57,108`; `internal/pnpmsource/tmp_controls_test.go:15` |
| Workflow guard/refusal plants and real script tests | `tools/ciworkflow/tmp_hygiene_test.go:45,52,123,178,226,259,274` |
| Full validation and runners | Complete command/exit records and signed-head Actions run in the sections below |

## Coverage map

| AC row | Production entry point / fixture call site | Named evidence | State |
| --- | --- | --- | --- |
| Restore before child creation and after success/Fatal | `TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization` → `writableTempDir` → `t.Cleanup`; `captureFixtureError`, `DerivePrivateStore`, `Materialize` | `TestPNPMTemporaryTreeRemovable/{pass,failure,capture-failure}` | Green |
| Sweep neighboring readonly creators | npm/Yarn/pnpm `writableTempDir`; capture/cache/execution/materialize/work roots | `TestWritableTempDirRestoresNestedTrees` in all four packages | Green |
| Every Linux self-hosted job has guarded final shared step | `.github/workflows/ci.yml` → `bash scripts/restore-runner-tmp.sh`; lint invokes `go test -count=1 ./tools/ciworkflow/` | `TestLinuxCIWorkflowTemporaryTreeRestore`, `TestLinuxCIWorkflowTemporaryTreeRestorePlants` | Green |
| Root refusal, chmod-only recovery, no deletion or external-link traversal, never fail job | `scripts/restore-runner-tmp.sh` | `TestRunnerTemporaryTreeSelfHeal`, `TestRunnerTemporaryTreeRefusals`, `TestRunnerTemporaryTreeChmodFailure`, `TestRunnerTemporaryTreeRootTrailingNewline` | Green |
| Regression control and narrowing mutants | Go overlay runs real conformance body; script override runs behavioral suite | `TestPNPMTemporaryTreeCleanupControls`, `TestRunnerTemporaryTreeNarrowingControls` | Green |
| Full local suite, lint and build | Go package suite, `go vet`, `go build`, `golangci-lint run` | Logs below | Green |
| Full hosted gate and runner identity | CI on exact signed snapshot, Actions jobs API and complete-runner logs | Run linked above | Green |

Measured coverage: **7 of 7 AC rows driven and green**. All required rows have persisted evidence.

## Mutant evidence

| Mutant | What it narrows | Named failing test | Result / survivor bound |
| --- | --- | --- | --- |
| `current-only` (retains chmod/root gate) | Restore only current/, admit stale read-only tree | `TestRunnerTemporaryTreeSelfHeal` (`not restored`) | Killed; child exit 1, expected red |
| `admit-root-alias` (retains root gate) | Refuse literal / but admit its alias | `TestRunnerTemporaryTreeRefusals/root-alias` | Killed; child exit 1, expected red; chmod spy prevents mutation outside fixture |
| `admit-double-slash-root` (retains literal / refusal and chmod) | Admit // only after physical resolution | `TestRunnerTemporaryTreeRefusals/double-slash-root` | Killed; behavioral child exit 1 expected red; harmless chmod spy |
| Per-job missing final step; echo preserving script token; success-only; hosted-only; later step; blocking error; wrong shell | One of seven jobs loses one required boundary | `TestLinuxCIWorkflowTemporaryTreeRestore` (real test binary on each planted workflow) | All 49 exit 1 expected red, harness exit 0; this is a structure contract, not a general expression interpreter |
| Strip only trailing pathname newline (retains all root gates and workflow invocation) | Admit a sibling chmod target for one legal pathname class | `TestRunnerTemporaryTreeRootTrailingNewline` | Killed; behavioral child exit 1, expected red |
| Remove cleanup restore | Plant control | `TestPNPMTemporaryTreeRemovable` | Killed; child exit 1, expected red (`non-writable entry after body cleanup`) |
| Skip only capture-store restoration, keep t.Cleanup token | Admit capture-store readonly leftovers only | `TestPNPMTemporaryTreeRemovable` | Killed; child exit 1, expected red (`non-writable entry after body cleanup`) |

## Sweep: fixes and exceptions

Searches covered pnpm/npm stores, module caches, explicit chmod/read-only modes, tar extraction modes and snapshot fixtures. Search output is saved locally as `sweep-01.log` and `readonly-files-01.txt`; the table below lists concrete code sites rather than asserting arbitrary dynamically computed mode coverage.

- Fixed: all temporary roots in the four manager conformance fixtures now register restoration before callers populate/seal them, covering capture stores, real-manager execution roots, private stores/mirrors/caches, materialized outputs and work trees. Previously cleanup was registered after a fallible producer and only for the store.
- Restoration uses a preorder walk, ORs owner write into existing permissions, and skips symlinks; it preserves executable bits and does not chmod an external target.
- Module caches: `internal/godriver/session.go:395` creates private cache directories; `session.go:445` closes the build environment (`GOPROXY=off`, `GOSUMDB=off`), builds use vendored modules rather than downloading readonly cache trees. `main_test.go:228` closes sessions. CI Go module caches live under `RUNNER_TEMP`, separate from the Go test `TMPDIR`; the new hook-root restore does not claim to rewrite those caches.
- Existing protected-tree owners: `internal/buildrepo/local.go:92` registers release before sealing; `local.go:858` restores during release. `internal/closureexec/closureexec_test.go:26,472,509,581` restores/removes owned captures. `internal/swiftpmbuild/fixture_test.go:58`, `internal/swiftpmsource/swiftpmsource_test.go:628`, `internal/swiftpmsource/swift_integration_test.go:64` register recursive restore before building child fixtures. Rust manager tests close managers (for example `internal/rustsource/build_test.go:33`), whose owned cleanup restores permissions (`manager.go:837`).
- Archive fixture headers are inert bytes until extracted by the tested adapter. The npm/Yarn/pnpm extraction roots are now covered by early cleanup. `internal/artifactpolicy/test_helpers_test.go:34` builds header fixtures; it does not extract them into a readonly filesystem tree.
- Existing mode-denial tests below restore explicitly; readonly regular-file fixtures under writable parents do not prevent POSIX unlink; empty readonly directories are removable from writable parents. These are listed exceptions to additional code edits, not untested claims about all Windows ACL behavior.

| Site | Disposition |
| --- | --- |
| `cmd/curator/builds_test.go:531` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `cmd/curator/draft_status_test.go:701` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `cmd/curator/env_unmanage_mode_unix_test.go:16` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `cmd/curator/hook_test.go:113` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `cmd/curator/status_test.go:1217` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/adapters/adapters_test.go:117` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/buildrepo/acquisition_conformance_test.go:717` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/buildrepo/admission_test.go:863` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/buildrepo/snapshot_budget_test.go:53,209` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/closure/refresh_atomicity_test.go:72` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/closureexec/closureexec_test.go:51,57,545,608,1123,1135,1150,1585` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/closureexec/portable_runner_unix_test.go:34` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/contextstore/state_read_test.go:31` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/crossconformance/draftsources_corpus_test.go:16` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/crossconformance/draftsources_transaction_e2e_test.go:67` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envprofile/import_test.go:245,780` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envprofile/materialize_capture_test.go:884` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envprofile/path_source_group_write_unix_test.go:30,42` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envprofile/pathkind_test.go:189,365` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envprofile/state_read_sites_test.go:22,36` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envsnapshot/destination_controls_test.go:38` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/envsnapshot/destination_unix_test.go:399` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envsnapshot/publication_unix_test.go:300` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/envsnapshot/verify_unix_test.go:197,203` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/gitops/deadlock_test.go:290` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/globalbins/globalbins_test.go:326` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/godriver/fingerprint_test.go:175` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/godriver/unreadabledir_other_test.go:12,19` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/godriver/unreadabledir_windows_test.go:15` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/godriver/worker_test.go:375` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/hookapproval/hookapproval_test.go:333,589` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/nodesource/conformance_test.go:31` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/npmsource/conformance_test.go:1036,1066,1082,1139` | Fixed: early writableTempDir restore across conformance fixture roots. |
| `internal/pnpmsource/conformance_test.go:1373,1389` | Fixed: early writableTempDir restore across conformance fixture roots. |
| `internal/rustsource/build_test.go:145,170,311` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/scopes/gc_conservative_test.go:186` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/scopes/gc_test.go:331` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/scopes/scopes_test.go:49` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/skillspec/fallback_test.go:123,175` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/swiftpmbuild/fixture_test.go:223` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/swiftpminterop/conformance_test.go:364` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/swiftpminterop/parser_test.go:210` | Readonly modes in file/archive/semantic fixtures; no nested readonly-directory creator identified. |
| `internal/swiftpmsource/swift_integration_test.go:135,355,375,406,409,436` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/transaction/root_metadata_unix_test.go:110` | Existing explicit restore / owned-tree release, or readonly regular file / empty directory under writable parent; no shared pnpm-store failure identified. |
| `internal/yarnclassicsource/conformance_test.go:676,693` | Fixed: early writableTempDir restore across conformance fixture roots. |
| `internal/yarnmodernsource/conformance_test.go:1562,1565,1582` | Fixed: early writableTempDir restore across conformance fixture roots. |

### Additional literal and helper sweep

A second sweep includes `os.Chmod`, method Chmod, shell chmod, mode-zero helpers and variable-mode fixtures (`chmod-sweep-03.log`). Hexadecimal strings, tar metadata counters, read-only API names and textual Git modes were false positives rather than filesystem tree creators.

| Site | Disposition |
| --- | --- |
| `cmd/curator/profile_path_permissions_test.go:8` (callers `profile_test.go:191,203`) | Mode-zero directories; caller restores through the returned cleanup; privileged skip restores before skipping. |
| `cmd/curator/muse_root_context_test.go:283`, `internal/envprofile/managed_dotfile_conformance_test.go:273`, `internal/transaction/state_read_test.go:24` | Existing explicit Cleanup restoration for denied file/parent/journal permissions. |
| `internal/envprofile/path_permissions_test.go:8` (caller `envprofile_f10f11f12_test.go:208`), `internal/registry/registry_test.go:666`, `internal/ui/state_read_test.go:30`, `internal/contextlock/read_state_test.go:27`, `internal/marker/read_state_test.go:30,56`, `internal/install/ownership_cases_test.go:117` | Regular-file read-denial fixtures; explicit returned/registered restoration, writable parents. |
| `internal/crossconformance/draftsources_semantic_transport_test.go:348` | Mode-zero regular policy file under writable parent; POSIX unlink does not require file write permission. |
| `internal/swiftpminterop/fixture_test.go:34`, `internal/crossconformance/fixtures_test.go:180` | Existing early recursive restore before child fixture setup. |
| `cmd/curator/global_status_test.go:251,281`, `cmd/curator/cacheprotection_other_test.go:14`, `internal/buildcache/cache_test.go:562`, `internal/buildcache/protection_unix_test.go:17` | Protected cache fixtures use owner-writable 0600/0700 boundaries, not nested readonly directories; existing snapshot clearProtection before replacement. |
| `internal/buildcache/collect_unix_test.go:383`, `cache_test.go:653` | Variable-mode helpers used by cache tests; boundary manipulation retains owner writes. |
| `internal/snapshot/capture_test.go:108`, `internal/crossconformance/draftsources_snapshot_test.go:76` | Snapshot fixture files are 0644/0755, directories 0755. |
| `internal/buildsource/builddriver_conformance_test.go:156` | Per-file executable-mode vectors; parent directories remain writable. |
| `internal/envsnapshot/destination_acl_darwin_test.go:20,48` | Existing ACL cleanup registered before applying each deny/permit fixture; ACL repair is outside the Linux chmod-only script contract. |
| Remaining chmod sites in `chmod-sweep-03.log` | Set writable/executable modes (0600/0644/0700/0755/0777), restore existing modes, or alter only regular-file special bits; they do not introduce a nested readonly tree. |

## Out-of-contract rows and bounds

- No required AC row is intentionally excluded. No validation row is waived.
- Host hook implementation is outside scope by the brief; it was observed through Actions logs only. No host access or mutations were used.
- Fork and hosted routing are preserved. The candidate-conformance lane stays hosted by its existing routing; rose-air is main-only and is skipped for gate branches.
- POSIX mode/symlink neighboring tests explicitly skip Windows. Real pnpm process witness runs on the pinned pnpm runtime where present. The Linux CI step is guarded out on hosted/macOS/Windows.
- Workflow route discovery is the current repository routing contract (`LINUX_SELF_HOSTED`/`relux-linux`), not arbitrary future Actions expressions. Exactly seven current routable jobs are required.
- No claim is made about ACL repair, unavailable roots, concurrent hostile pathname replacement or filesystem-root chmod. The script only restores owner write permissions within the admitted resolved directory.

## Validation commands and source identity

All commands ran as standalone processes with redirected complete logs, without tee/pipeline status masking. Named shard scripts, exact masks, command records, exit records and sanitized complete logs are in `BUG-261018-bisu6n_evidence.json`. Every selected parent runs all its subtests; only the independently rerun artifact live-tree test and the separately scheduled authoritative dry-run test use explicit skip filters in their neighboring shards.

| Scope / command | Real exit | Evidence |
| --- | ---: | --- |
| CLI A–C, D, E, F–L, M/N/O/R: `go test -count=1 -timeout=8m ./cmd/... -run <recorded mask>` | 0 each | `local-cmd-{ac,d,e,fl,mor}-03.*` |
| All 55 Profile declarations, six named CLI shards | 0 each | `cmd-prof-part{1..6}-04.*`, `prof-shards-04.txt` |
| Remaining 108 CLI declarations, nine named shards | 0 each except part 6 exit 1 (11 passed; hostile-config precondition failed); named retry below | `cmd-rest-part{1..9}-04.*`, `cmd-rest-shards-04.txt` |
| Signer hostile-config precondition retry with inherited GIT_CONFIG_* overrides removed | 0 | `signer-hostile-retry-05.*` |
| 70 other packages, A–F, G–L plus fuzz seeds, S–Z | 0 each | `local-other-{af,gl,sz}-03.*` |
| Same M–R aggregate | 1 (historical failure) | `local-other-mr-03.log`: 68 packages green; envprofile timeout and missing Rust PATH retried below |
| envprofile M/N/O: nine named shards; P: seven named shards; R: direct subset | 0 each | `env-mno-part{1..9}-04.*`, `env-p-part{1..7}-05.*`, `envprofile-r-03.*` |
| Full Rust package, with existing pinned rustup shims on PATH | 0 | `rust-full-03.*` |
| install A–C excluding separately run authoritative case, E–F, M–R, S–Z | 0 each | `install-{ac,ef,mr,sz}-04.*` |
| install D: eleven named shards; G–L: fifteen named shards | 0 each | `install-d-part{1..11}-05.*`, `install-gl-part{1..15}-05.*` |
| Authoritative dry-run case, after CLI released host GOROOT lock | 0 | `install-authoritative-04.*` |
| Four complete manager packages | 0 | `local-node-tools-03.*` (older tools/ciworkflow result superseded by current full suite below) |
| godriver/scriptworker A–F; both plus artifactpolicy G–Z | 0 each | `local-godriver-scriptworker-af-03.*`, `local-other-four-03.*` |
| Artifactpolicy A–F excluding independently rerun live-tree case; full 100,001-entry case | 0 each | `artifact-af-rest-03.*`, `artifact-live-tree-04.*` |
| Exact-pin snapshot acquisition retry | 0 | `snapshot-retry-03.log` |
| `go test -count=1 -tags tmphygienecontrols -v -timeout=8m ./internal/pnpmsource -run '^TestPNPMTemporaryTreeCleanupControls$'` | 0; both planted children 1 expected red | `pnpm-controls-03.log` |
| Current full `tools/ciworkflow` suite | 0 | `workflow-full-05.*` |
| Current real workflow-gate binary: positive control plus 49 plants | Harness 0; positive 0; each plant 1 expected red | `workflow-control-build-06.log`, `workflow-controls-06.log`, individual `workflow-<job>-<plant>.log` |
| `go vet ./...`; `golangci-lint run`; `go build -o <scratch>/curator ./cmd/curator` | 0 each | `vet-05.log`, `lint-06.log` (0 issues), `build-05.log` |
| `bash -n scripts/restore-runner-tmp.sh`; `git diff --check` | 0 each | `bash-syntax-05.log`, `diff-check-06.log` |

Local host: macOS arm64, Go 1.25.5, Node v25.6.1, pnpm 10.33.0, Rust/cargo 1.91.0. Go uses GOENV=off and GOTOOLCHAIN=local. The CI-style PATH excludes ambient operator providers; the Rust package additionally uses the existing rustup shim directory. Pinned pnpm was installed into ignored task scratch using `npm install --prefix .temp/BUG-261018-bisu6n/pnpm pnpm@10.33.0` (exit 0). No new test library was added. Default optional CURATOR_REAL_GO_TEST/CURATOR_REAL_GO_BUILD_TEST flags remain unset locally; hosted Go-family qualification executes the native driver lanes.

The full local suite is the union of bounded shards. The only fuzz entry, `FuzzNormalizeLocalSelectorIdempotent`, is included in the G–L package command. There are no Example declarations. `cli-mask-coverage-04.json` maps all 426 discovered CLI top-level test declarations: 263 in original range masks, 55 Profile and 108 remaining declarations, zero unmapped. Normal build constraints still apply locally; hosted matrix jobs cover supported platforms. Five packages have no test files and are compiled by `go vet ./...`/the build. All 109 previously committed test names in modified existing test files remain unchanged; all other reviewer files are untouched and their package suites remain in the coverage union.

Evidence reuse is scoped to unchanged source/test/configuration inputs. The delta between signed snapshot 3630472b... and current 102a56de... contains only `scripts/restore-runner-tmp.sh` and `tools/ciworkflow/tmp_hygiene_test.go` (`candidate-identity-delta-05.log`). The affected full tools suite, controls, vet/lint/build and hosted gate were rerun. Unchanged manager/CLI/internal results retain their recorded environment and immutable spec input identity; no claim is made that the older tools result covers the new pathname test.

The spec is the raw checkout of SPEC_PIN `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, manifest sha256 `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`. A temporary git index plus checkout-index preserved raw fixture bytes. Only snapshot acquisition reads the export-subst fixture transformed by the earlier archive attempt; its named retry and subsequent current-root suite are green. The live-tree artifact case was rerun with an explicit command record to bind that retained evidence precisely.

## Hosted runners

| Job | Runner | Conclusion |
| --- | --- | --- |
| Lint | relux-rise1-curator-1 | success |
| Race (ubuntu-latest) | relux-rise1-curator-4 | success |
| Race (macos-latest) | GitHub Actions 1000037615 | success |
| Test (windows-latest) | GitHub Actions 1000037614 | success |
| Gate self-test (ubuntu-latest) | relux-rise1-curator-2 | success |
| Go 1.27.1 driver (macos-latest) | GitHub Actions 1000037617 | success |
| Test (macos-latest) | GitHub Actions 1000037623 | success |
| Go 1.25.5 driver (windows-latest) | GitHub Actions 1000037616 | success |
| Test (ubuntu-latest) | relux-rise1-curator-3 | success |
| Naming gate | relux-rise1-curator-1 | success |
| Go 1.25.5 driver (macos-latest) | GitHub Actions 1000037622 | success |
| Interop conformance gate | relux-rise1-curator-2 | success |
| Go 1.26.0 driver (windows-latest) | GitHub Actions 1000037618 | success |
| Go 1.26.0 driver (macos-latest) | GitHub Actions 1000037621 | success |
| Go 1.25.5 driver (ubuntu-latest) | relux-rise1-curator-3 | success |
| Go 1.27.1 driver (ubuntu-latest) | relux-rise1-curator-2 | success |
| Go 1.27.1 driver (windows-latest) | GitHub Actions 1000037619 | success |
| Gate self-test (macos-latest) | GitHub Actions 1000037624 | success |
| Go 1.26.0 driver (ubuntu-latest) | relux-rise1-curator-1 | success |
| Gate self-test (windows-latest) | GitHub Actions 1000037620 | success |
| Test (rose-air) | Not scheduled | skipped |
| Candidate suite (${{ matrix.os }}) | Not scheduled | skipped |

All nine self-hosted Linux job executions reported both the restore step and Complete runner as success in the Actions steps API. No pre-heal inventory of the runner's stale tree was collected; passing hook execution is not presented as a measurement of its previous contents. The deterministic current/stale fixture and narrowing current-only mutant establish that behavior independently.

## Logbook / rejected evidence

- Initial status mutation refused for a missing estimate. Added Fibonacci 5, then entered development before implementation.
- Incident run 39119857144 currently points to successful attempt 3. Attempt 1 job 117653063370 on curator-4 contains the actual Complete runner Permission denied witness. Empty latest-attempt --log-failed output was not used as proof.
- Superseded gate runs 39147026994 and 39147383194 were cancelled after skip-vocabulary and mutant-temp-isolation corrections. Run 39148113242 was cancelled after the pathname bug below. Run 39151156127 was cancelled after the double-slash-root fix; only the linked current head is accepted for hosted validation. Their job successes are diagnostic history, not the current hosted gate.
- Initial local internal/tools run exited 1 with a mutable spec digest f1a078a8... and an 8-minute artifact timeout. CLI A–L exited 1 with ambient provider discoveries and timeout; simultaneous M–Z exited 1 before tests at its package GOROOT lock deadline. These are rejected evidence. No shared lock or another operator's process was removed.
- Git archive at the correct spec pin transformed an export-subst fixture. Replaced it with raw checkout-index materialization; the named snapshot acquisition retry exited 0. Correct manifest identity was preserved.
- Broad A–F/M–R, install, Profile and envprofile groups exceeded the 8-minute per-command budget, exit 1. Smaller named shards preserve every test and replace those partial group results. Two real Rust capture tests failed with rustc missing from the restricted PATH; the existing pinned Rust toolchain was selected and the full package rerun green. No product workaround, shortened fixture or weakened assertion was introduced.
- Regression/control children now use an outer cleanup-owned temporary root, so disabling cleanup inside a plant cannot leak readonly fixtures into the host tmp. The tag-controlled pnpm plants both fail by the named non-writable-entry witness.
- A chmod-spy control reproduced trailing-newline loss in command substitution without changing external permissions. Sentinel normalization preserves that legal pathname. `TestRunnerTemporaryTreeRootTrailingNewline` and its token-preserving narrowing mutant prove restoration of that root while keeping its trimmed sibling unchanged.
- CLI tail shard 6 exited 1 because inherited GIT_CONFIG_COUNT overrides replaced its generated hostile-config fixture key with the host signing key. Eleven other declarations passed. Removed only inherited GIT_CONFIG_* launch overrides for the named hostile-config retry; its own hostile fixture remains active. No signer, key, global config or product/test source changed.
- Read-only Bash probe showed pwd -P preserves //, so canonical string equality alone did not refuse every spelling of the filesystem root. Added same-file identity refusal (-ef /), double-slash-root fixture and its literal-only narrowing mutant. The chmod spy kept every refusal plant harmless.
- Initial current-head lint launch exited 127 because the restricted PATH excludes the linter install directory. Reran the existing linter via its absolute entry point; no installation or provider PATH broadening was needed.
- The board refused binary tar.gz resources with resource_binary_refused (NUL byte). The same complete sanitized evidence is published as a plain-text JSON container, which losslessly escapes control bytes. The verified tar.gz remains in ignored task scratch; the JSON records its neutral relative path, size and sha256.
- Every candidate used a detached author-signed commit-tree snapshot for hosted validation. Story HEAD remains at its recorded checkpoint, with the reviewed changes uncommitted for native handoff. No Story commit, switch, rebase, host hook, runner-label or repository-variable change was made.

## Changelog entry

Fix read-only pnpm/npm/Yarn test temporary-tree cleanup and add guarded, chmod-only permission restoration for self-hosted Linux runner temporary roots, preserving exact pathnames.

Current curator-4 evidence: Race (ubuntu-latest), job 117751414547, restore step success and Complete runner success. The exact-head restore/cleanup tail is retained as hosted-curator4-restore-complete-05.log.
