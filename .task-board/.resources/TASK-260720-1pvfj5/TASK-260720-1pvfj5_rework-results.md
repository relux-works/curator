# TASK-260720-1pvfj5 — cross-platform compiled-build CI gates (rework composite)

**Role:** developer · **Run:** `RUN-260729-703e35` · **Date:** 2026-07-29
**Supersedes:** the `RUN-260729-9ead43` outcome, which validated `origin/main` and
therefore treated the accepted compiled-build packages as absent.

Nothing here is a release, a release claim or a conformance claim. No commit,
stage, tag, publish or pin promotion was performed.

---

## 1. Composite provenance — the accepted product, proven byte for byte

| Item | Value |
| --- | --- |
| Composite worktree | `.temp/TASK-260720-1pvfj5/rework/composite` |
| Base | `git worktree add --detach 17804ce` (origin/main, "Pin landed rc.3 protocol") |
| Product bytes | `rsync -a --delete` from `.temp/TASK-260720-jrrgw9/worktree` (the accepted candidate) |
| Manifest generator | `.temp/TASK-260720-jrrgw9/integration/manifest.py`, unmodified |

Verified against the accepted evidence before a single CI file was written:

```
live candidate manifest   sha256 83b5df8b81ec21f472a90ad082d1ab3464b807968418a7d3d82c8672ff6a2819  (359 entries)
composite base manifest   sha256 83b5df8b81ec21f472a90ad082d1ab3464b807968418a7d3d82c8672ff6a2819  (359 entries)
cmp vs verifier-4 post-run-manifest.txt   exit 0
```

That digest is the one the accepted `TASK-260720-jrrgw9` final review verdict
records, and `verifier4/post-run-manifest.txt`, `verifier4/current-manifest.txt`,
`integration/manifest-post.txt` and `integration/manifest-final.txt` all carry it.
The composite is therefore the exact accepted composite: `TASK-260729-2kaopg` plus
the accepted `TASK-260720-2qqq0w` implementation plus the accepted `jrrgw9`
integration (Patch A + the accepted namespace optimisation).

Package inventory confirms the point the superseded run got wrong: the composite
has **40 packages**, including `internal/godriver`, `internal/buildcache`,
`internal/buildsource`, `internal/buildmeta`, `internal/transaction`,
`internal/install/atomicity` and `internal/interop`. `origin/main` has 31 and none
of those.

### 1.1 The overlay is exactly the CI and quality surface

After the overlay the manifest is 372 entries. The delta from 359 is:

* **13 new files**, all under `.github/ci/`;
* **3 modified**: `.github/workflows/ci.yml`, `Makefile`, `README.md`.

Restricting both manifests to everything *outside* that surface:

```
base  product entries  356
final product entries  356
cmp                    exit 0     <- every accepted product byte preserved
```

`.golangci.yml` was deliberately **not** touched. No Go file was touched. No
fixture, timeout, pin, module or conformance vector was touched.

| File | sha256 |
| --- | --- |
| `.github/ci/candidate-suite.sh` | `e2e874a699db5d45abd81d9c0a74f2c53981de9dac6cfdcf08a560a9949e149f` |
| `.github/ci/excluded-packages.sh` | `faafba5d5c361324b0759a4ca066e0d5f61a5c697ffa3c83cdb79afaebc9618a` |
| `.github/ci/gate-selftest.sh` | `bd7eafe9b8a51d8b4292ec3330762fa1e2af3879009debb050edab1af9bd7a6a` |
| `.github/ci/ledger-consistency.sh` | `21ab8fd4d7f614ca571337dbf80a8df294484a89330881662b36bd331b023a84` |
| `.github/ci/no-broad-suppression.sh` | `ded21636302653e4b9e2205a7e8fa44192d7c7178017dcf259b759c4c77831a5` |
| `.github/ci/platform-case-gate.sh` | `c358d136f42f3c4bfed9625c4ae6117fb0b52742a8656b1bd494784a0a79bf4b` |
| `.github/ci/platform-cases.tsv` | `168f48603ae0dc56fbbc4392e05cf6d95bd360be3a6d56113b4de03ac8cbbc3d` |
| `.github/ci/platform-exclusions.tsv` | `e63416207b89419f09475ecb3fb541181fa8d8baf55c19f96c5624ffec109528` |
| `.github/ci/root-artifacts.tsv` | `15c96c9ea79f6f3a9ff49b1cf87c9e1842ac83af1f01ab5937c6860c21bd715a` |
| `.github/ci/skip-classes.tsv` | `bd605c5dd9bb4922b47a354bd628ff4579a9b20f85d2b629de5ef8ca86051c28` |
| `.github/ci/suite-plan.sh` | `bb76ca726ce4cbe08cfdce2b2d44b614974d493a7e9b538961914629dc321955` |
| `.github/ci/test-gate.sh` | `2ca6c98dbb0e5789e8aaecde2718be712799b1b0feacee68d0975b268d97725f` |
| `.github/ci/toolchain-identity.sh` | `7834ede17dff20efe43184760c2c955c532563324c0559027767807a22f59570` |
| `.github/workflows/ci.yml` | `0626efe3818add42fc5cb9b8ee4c24829755d7b2baa5eb3b87e701feee794630` |
| `Makefile` | `23ed81458d63b4e6663b116e964c030e86a729384913d23b4f411de92ba3e9ba` |
| `README.md` | `b777bdb5ae36fbb23ca6098b667d47da10d6725cabb4feacc154bd147dc99f03` |

Patch: `TASK-260720-1pvfj5_ci-overlay.patch`, sha256
`8002063b9f3643652b71c3745c9f6010b1b47c83d3f172c3573e061f392201b7`, 16 paths.

---

## 2. The pin did not move

`SPEC_PIN = 00b1688a9b2457ca397a0bb550acf47cad8ee967` — the value the composite
already commits — is now declared **once**, in the workflow `env:` block, and read
from there by `test`, `race`, `interop` and (for comparison only)
`candidate-conformance`. Previously it was repeated literally in three jobs.

* No promotion. Promotion remains owned by `TASK-260720-38l1sy`, after
  `TASK-260720-25d05o` qualifies the release.
* The pin publishes `1.0.0-rc.3`. No committed file claims rc.4 or rc.5 as a
  release or as a pin: every rc.5 mention in the overlay is either a synthetic
  self-test fixture, a description of the *candidate*, or a citation of the
  protocol's own qualification vector.
* `gate-selftest.sh` reads the pin out of `ci.yml` and asserts it satisfies the
  same immutability shape a candidate must — so a branch or mutable tag can never
  be committed as the default pin either.

---

## 3. Two real conflicts, measured rather than assumed

The superseded run resolved these by declaring the behaviour absent. On the real
composite both are measurable, and both are resolved inside `ci.yml` + `.github/ci/`
without touching a Go file.

### 3.1 The released pin cannot serve the schema v6 conformance reads

Measured, full suite, committed pin root exported to every package:

```
go test -count=1 -json ./...      exit 1
1577 pass, 33 skip, 10 FAIL across 7 packages
```

The ten failures are unguarded reads of artefacts `00b1688a` does not publish —
`vectors/build-drivers.json`, `vectors/external-repository-lifecycle.json`,
`schema-cases/install-marker-v2`, `expected/build-driver/{marker,context_files}.json`,
`schema-cases/agent-skill-v6`, `fixtures/go-build-skill`. This is a property of
the accepted composite, not of this task.

**Resolution — a partition derived from the root, never from the lane.**
`.github/ci/root-artifacts.tsv` declares which root artefacts each package's
conformance tests read without a guard. `suite-plan.sh` checks the *supplied* root
and splits the package set:

| Root | served | deferred | excluded |
| --- | ---: | ---: | ---: |
| committed pin (rc.3), darwin/windows | 33 | 7 | 0 |
| committed pin (rc.3), linux | 32 | 7 | 1 |
| rc.5 candidate, darwin/windows | **40** | **0** | 0 |
| rc.5 candidate, linux | 39 | 0 | 1 |

A deferred package runs with `CURATOR_CONFORMANCE_ROOT` **unset**, taking the
`root is not set` path its own tests already implement, and the exact missing
artefact is named in the report. Against the rc.5 candidate the deferred set is
empty and the whole module runs with the root exported — which is what the
candidate lane's `CI_REQUIRE_FULL_ROOT=1` turns into a hard assertion. A candidate
run therefore cannot be green while any package quietly ran without its vectors.

Drift is fail-loud, not silent: a package that later gains an unguarded read and
is not listed goes red with the `open()` error naming the missing file.

### 3.2 Linux is excluded by the protocol, and the exclusion is asserted

`internal/godriver` implements `rc5-native-control-inventory-v1`, which defines
control records for macOS and Windows only (`controls.go:200`,
`InventoryPlatform` returns `""` for every other GOOS). The rc.5 root says the
same thing normatively:

```json
"platforms": [ { "name": "linux", "status": "excluded", "until_task": "TASK-260728-1skseh" }, … ]
```

`excluded-packages.sh` reads that vector from the **supplied root**; the recorded
`default_excluded_on` applies only to a root that predates the vector (the
committed pin is such a root, and publishes none). A root that publishes the
vector always wins, so when `TASK-260728-1skseh` flips linux, the package returns
to the Linux run automatically with no edit here.

The exclusion is never taken on trust. `test-gate.sh` runs, on the excluded runner
itself:

```
go test -count=1 -run '^TestProbeRejectsAnUncoveredPlatformBeforeTheWorker$' …/internal/godriver
```

which drives `probeNativeControlsFor("")` and `probeNativeControlsFor("linux")`
and requires `build_execution_control_unavailable` from both — the manager refuses
*before any worker starts*. Measured on this host: **exit 0, `--- PASS`**. A ledger
row requires that case on linux, darwin **and** windows.

---

## 4. How "no case silently skipped" is enforced

### Tier 1 — `.github/ci/platform-cases.tsv` (49 rows)

Required-passing counts: **darwin 37, windows 32, linux 31**. Coverage against the
acceptance criteria, every name verified to exist in the composite:

| AC behaviour | Cases | Required on |
| --- | --- | --- |
| Windows DACL | `buildcache::TestWindowsProtectedStateMatrix`, `TestValidateWindowsSecurityPolicy` | windows |
| Windows reparse | `buildcache::TestSweepDoesNotFollowAReparsed{Entry,CacheRoot}`, `scopes::TestCollectStaysFailSafeOnRedirectedWindowsScopes` | windows |
| `.cmd` launchers | `runtimestore::TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode` (drives real `cmd.exe`), `TestInstallSingleCommandAndShims`, `TestRuntimeCommandPath` | windows / all three |
| Unix ownership & permission | `buildcache::TestUnixProtectedStateMatrix`, `TestUnixProtectionHelperFailuresFailClosed`, `config::TestBootstrapAndAddProject` | linux, darwin |
| No-follow | 4 × `buildcache` sweep/retire cases, `scopes::TestCollectStaysFailSafeOnRedirectedUnixScopes`, `buildsource::TestRejectsSpecialFile`, `gitops::TestArchiveRejectsLinks` | linux, darwin |
| Read-only source | `buildsource::TestFrozenTokenRejects{Mutation,RootReplacement}`, `TestWithValidatedOrdersCallbackAndRejectsMutation`, `install::TestDryRunTouchesNothing`, `whitelist::TestRuntimeRootsExcludedFromContext` | all three |
| Resource policy | `godriver::TestResourceLimitsStayInsideManagerBounds`, `TestProbeSetGuardMatchesTheInventoryExactly`, `TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes`, `TestPerFileSizeLimitIsReallyApplied` (macOS control), `TestBuildFailsClosedWhenTheGoChildCannotStart` (Windows control) | darwin, windows |
| Executable | `godriver::TestExecutableIdentityRejectsSubstitutionAndMutation`, `runtimestore::TestUnixPostInstallShimPropagatesSignal`, `shell::TestPosixHook…` / `TestPowerShellHook…`, `envfiles::TestWriteProjectShapesAndSourcing`, `install::TestEndToEndInstall`, `adapters::TestSymlinkModeStagesRelativeLinkTargets` | per platform |
| Transaction & atomicity | 3 × `transaction` namespace-alias cases, 3 × `install/atomicity` rollback/mirror cases | all three |
| Interop / conformance | `interop::TestGoldenContextCopy`, `godriver::TestCandidateGoV1SourceAwareContract` | all three / darwin,windows |
| Protocol exclusion | `godriver::TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` | all three |

A skip of a listed case — or of any of its subtests — is fatal unless the row
tolerates it *and* the reason the test actually printed carries the class the row
declared.

### Tier 2 — `.github/ci/skip-classes.tsv` (35 rows)

Every **other** skip must have a reason matching a declared class. An unrecognised
reason is fatal, so a newly-introduced skip is caught the first time it runs. The
class vocabulary was derived from the complete set of skip literals in the
composite, not guessed. `root-unset` is `deferred-only`: it is legal only for a
package `suite-plan.sh` deferred, so it cannot appear at all in a lane that
requires a fully-serving root.

Every skip, tolerated or not, is written to `skips-observed.tsv` with package,
case, class, verdict and verbatim reason. Measured on darwin:

| Lane | root-unset | root-content | host-capability | platform-control | opt-in | helper-process | total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| candidate (rc.5) | 0 | 2 | 4 | 2 | 3 | 1 | **12** |
| default (pin rc.3) | 19 | 19 | 1 | 2 | 3 | 1 | **45** |

### Tier 0 — `ledger-consistency.sh`

`go list` reports the exact test file set for each target GOOS, so every ledger
claim is checked against the **real linux, darwin and windows builds** from any
host. It fails a row that requires a case the target build does not compile, and a
case compiled into a build the ledger never mentions. It caught two invented names
during authoring (`adapters::TestSymlinkModeUsesRelativeLinks` and an
`install/atomicity` case that exists only on `origin/main`), which is exactly the
class of error the superseded run shipped.

---

## 5. Gates measured — real commands, real exit codes

Host: Darwin arm64, `go1.25.5` (`GOROOT=/Users/iv/.goenv/versions/1.25.5`),
task-owned `GOTMPDIR` under `.temp/TASK-260720-1pvfj5/rework/gotmp`. Each command
ran as its own unpiped process; the status shown is that process's own.

| # | Command | Exit |
| --- | --- | ---: |
| 1 | `bash .github/ci/test-gate.sh` — **default lane**, committed pin root | **0** |
| 2 | `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` — **candidate lane**, rc.5 root | **0** |
| 3a | `make race` (`GO_TEST_FLAGS=-race`), committed pin root — attempt 1 | **2** (`go test` 1, gate 0) — see §5.1 |
| 3b | `make race` (`GO_TEST_FLAGS=-race`), committed pin root — attempt 2 | **0** |
| 4 | `make gate-selftest` | **0** — 70 passed, 0 failed |
| 5 | `make ledger-check` | **0** — 49 rows across linux, darwin, windows |
| 6 | `make no-broad-suppression` | **0** |
| 7 | `gofmt -l cmd internal` | **0**, empty listing |
| 8 | `go vet ./...` | **0** |
| 9 | `make build` | **0** |
| 10 | `go test -run '^TestProbeRejectsAnUncoveredPlatformBeforeTheWorker$' …/godriver` | **0** — `--- PASS` |
| 11 | workflow YAML shape (`GOENV` is the string `off`, `GOTOOLCHAIN` is `local`, `SPEC_PIN` is 40-hex) | **0** |
| 12 | naming gate | **0** — one README line, nothing elsewhere |
| 13 | `golangci-lint run` @ **v2.12.2** (the version CI pins) | **1** — see §6 |

Lane detail:

* **Default lane** — `served exit=0`, `deferred exit=0`, merged `go test exit=0`,
  platform-case gate `ok`. 40/40 packages pass; 45 skips, all classified.
* **Candidate lane** — `served=40 deferred=0 excluded=0`, `go test exit=0`,
  platform-case gate `ok`, **37 required cases observed passing**, 12 skips, none
  in the `root-unset` class.

### 5.1 Race — and an intermittent failure in accepted test code

The race gate was run end to end through `make race` twice, because the first
attempt was red.

| | `go test` | gate | atomicity | transaction | godriver | cmd/curator | cases |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| attempt 1 | **1** | 0 | 86.68s | 34.56s | **fail** 63.17s | 414.23s | 1571 pass / 2 fail / 45 skip |
| attempt 2 | **0** | 0 | 87.49s | 35.26s | 64.64s | 415.97s | 1573 pass / 0 fail / 45 skip |

The single failure, attempt 1 only:

```
internal/godriver
  TestFingerprintCancellationStaysFailClosed/cancelled_between_the_walk_and_the_digest
  fingerprint_equivalence_test.go:519:
    error = go-v1 toolchain_mutated: toolchain file "a/b/c/d/leaf" changed while reading,
    want nil or toolchain_timeout
```

**This is an intermittent flake in accepted test code, not a defect this task
introduced and not a gate defect.** Characterised rather than assumed:

* the subtest deliberately races cancellation (`go cancel()`) against a 256 KiB
  read and asserts the outcome is `nil` **or** `toolchain_timeout`. The third
  outcome the implementation can legitimately produce in that window —
  `toolchain_mutated` — is not admitted by the assertion;
* **185 targeted executions under `-race` produced 0 failures**: `-count=25` once,
  then 8 × `-count=20`, then `-count=8 -cpu 1,2,8`. It does not reproduce in
  isolation, only under whole-suite scheduling;
* the accepted `jrrgw9` verifier-4 full race run over byte-identical product code
  exited **0** in 441.11s, so it did not reproduce there either;
* the platform-case gate reported **ok** on both attempts — the ledger and the skip
  classification are unaffected either way. The failure surfaced through `go test`'s
  own status, which `test-gate.sh` propagates and refuses to mask.

It is left unfixed here for the same reason as §6: the fix is in an accepted Go
test file. Recommended for its owner: admit `toolchain_mutated` in that subtest's
assertion, or remove the racy `go cancel()` in favour of the deterministic
phase-boundary sibling that already exists two subtests below.

**Timing headroom is comfortable.** `internal/install/atomicity` runs at 86.68 /
87.49s against the established 480s acceptance bar (verifier 4 measured 115.687s).
The long pole is `cmd/curator` at ~415s, well inside the 30m per-package timeout;
no timeout was inflated.

---

## 6. Open finding: the accepted composite is not lint-clean — **not fixed here**

`golangci-lint run` at the pinned **v2.12.2** exits **1** with four findings, all in
**accepted product code**:

| # | Finding | Assessment |
| --- | --- | --- |
| 1–2 | `G115: integer overflow conversion rune -> byte` — `internal/protocoljson/ccj.go:211,212` | **False positive.** Both conversions are inside `if character < 0x20`, so the rune is provably in `[0,0x1f]`. gosec does not track the branch guard. |
| 3 | `G602: slice index out of range` — `internal/transaction/journal.go:398` | **False positive.** `entries[index-1]` is guarded by `index > 0 &&`, which short-circuits. The identical pattern 13 lines earlier (`values[index-1]`, journal.go:385) is *not* flagged — the rule is reacting to the `[]byte(...)` conversion form, not to a real reachability. |
| 4 | `ineffectual assignment to environment` — `internal/godriver/builddriver_positive_conformance_test.go:178` | **True positive**, dead code in a test file: `environment` is assigned twice and never read; `values` is used instead. Behaviourally harmless. |

Cross-checked at **v2.4.0**: exit **1**, finding 4 only. So finding 4 is
pre-existing under any recent golangci-lint (including the `version: latest` the
composite previously used, which is the mutable input this task replaced), and
findings 1–3 are surfaced by pinning the newer gosec.

**Why this task did not fix it.** Both available fixes are outside what this task
may do:

* editing the three Go files would break the byte-identity with the accepted
  candidate that this rework exists to preserve (rework instructions 2 and 6);
* a `.golangci.yml` exclusion for `G115`/`G602` on production paths is precisely
  the "broad suppression for new security code" the acceptance criteria forbid —
  and would be rejected by this task's own `no-broad-suppression.sh`.

**Recommendation for the owner of those files** (three lines, no behaviour change):
add line-scoped `//nolint:gosec // G115: guarded by character < 0x20` and
`//nolint:gosec // G602: guarded by index > 0` with reasons, and delete the dead
`environment` variable. Both `//nolint` forms name the linter and carry a reason,
so they pass the suppression gate. Until then the `lint` job is red on this
composite, and that is reported rather than papered over.

---

## 7. Native Linux and Windows — honestly unqualified

Every Linux and Windows claim in this run is derived from the real per-GOOS builds
(`go list` file sets, build constraints) and from the protocol's own qualification
vector. **No native Linux or Windows execution was performed**, because no runner
was reachable:

| Host | Result |
| --- | --- |
| `ssh relux` | reachable, but **Darwin x86_64 with no Go** — not a Linux runner; installing one is forbidden |
| `ssh win` | `exit 255`, connection timed out (matches the jrrgw9 verifier-4/5 finding) |
| `ssh lev` | `exit 255`, connection timed out |
| `ssh chip`, `ssh reluxts` | hostname does not resolve |

Producer confirmations owed on the first hosted CI run, stated rather than assumed:

1. **Windows DACL and `.cmd`.** `TestWindowsProtectedStateMatrix`,
   `TestValidateWindowsSecurityPolicy` and
   `TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode` are required
   with **no skip tolerance**. If the hosted image cannot satisfy them the gate
   fails loudly by name — the correct outcome, and a real finding, not something to
   relax by editing the ledger.
2. **Windows reparse points.** Those three cases tolerate a `host-capability` skip,
   because the source guards them with "this host cannot create a directory reparse
   point". DACL is required strictly, so the AC's "DACL **or** reparse" clause holds
   either way.
3. **Undeclared Windows/Linux subtest skips.** Tier 2 will fail loudly on any skip
   reason not in `skip-classes.tsv`. The vocabulary was derived from every skip
   literal in the composite, so this should hold; if it does not, add the row with
   the measured reason — do not widen a class.
4. **`pwsh` on the Windows runner** for `TestPowerShellHookRunsOnEveryPrompt`.
5. **Race on `ubuntu-latest`** is unmeasured on Linux (macOS-only measurement).
6. **`golangci-lint v2.12.2` under `golangci-lint-action@v7`** ran clean *as a
   standalone binary* here; the action's own install path is unverified.

---

## 8. Handoff to `TASK-260720-38l1sy` (released-pin audit)

* **Pin under audit:** `SPEC_PIN = 00b1688a9b2457ca397a0bb550acf47cad8ee967`,
  declared once at `.github/workflows/ci.yml` `env:`. Protocol `1.0.0-rc.3`.
  Described by no release tag (`v1.0.0-rc.2-1-g00b1688`) — immutable but untagged.
* **This task did not move it.**
* **Candidate evidence format** a `candidate-conformance` run produces, per runner,
  as `candidate-evidence-<os>`: `candidate-suite-identity.txt` (revision, root,
  protocol version, `manifest_sha256`, `tree_sha256`, `file_count`, the committed
  pin it was compared against, `evidence_class candidate-only`,
  `release_claim none`, `conformance_claim none`), plus `suite-plan.txt`,
  `plan-{served,deferred,excluded,assert}.txt`, `go-test*.json`,
  `platform-cases.txt`, `skips-observed.tsv` and `observed-cases.tsv`.
* **Reference candidate identity, measured** (rc.5 root
  `.temp/TASK-260729-3nx97g/worktree/conformance/v1`):
  `manifest sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`,
  `tree sha256:e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`,
  448 files, `protocol_version 1.0.0-rc.5`. Those reproduce the independently
  accepted `TASK-260729-3nx97g` values exactly.
* **Promotion pre-flight the gates now give for free.** Point `SPEC_PIN` at the
  proposed revision: `test` and `race` must stay exit 0 on all runners with the
  platform-case gate green; `suite-plan.sh` must report **`deferred=0`** against it
  (a released pin that still defers packages is not yet serving the implementation);
  `candidate-suite.sh verify-ref` must accept its shape and will refuse a revision
  equal to the current pin, so a no-op promotion cannot be recorded as one.

---

## 9. Hygiene

No commit, stage, stash, checkout, tag, push, pin change, broad suppression,
fixture weakening or timeout inflation. `.golangci.yml` and every Go file are
byte-identical to the accepted candidate. The outer checkout was not modified. All
scratch state is under `.temp/TASK-260720-1pvfj5/rework/` (composite worktree,
task-owned `GOTMPDIR`, logs, packet); no shared Go cache was cleared and no
software was installed.
