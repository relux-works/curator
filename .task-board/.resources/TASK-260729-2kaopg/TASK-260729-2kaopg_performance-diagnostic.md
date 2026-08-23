# TASK-260729-2kaopg performance diagnostic

Date: 2026-07-29
Worktree: `.temp/TASK-260729-2kaopg/worktree` (base `origin/main` 17804ce, nothing
staged or committed, **no product or test file changed by this run**)
Role: developer, performance rework directive after tester RUN-260729-051449

## Result in one line

The `cmd/curator` aggregate duration is **not** caused by anything inside the
`cmd/curator` test setup. It is caused by one shared production hot path —
`internal/godriver.fingerprintToolchain` — that re-hashes the whole GOROOT once
per plan and three more times per staged build. The prescribed test-side levers
(share immutable setup, task-owned fixtures, split integration packaging) cannot
reach the ≤480s target, and a digest-identical fix to that hot path can. That
fix is outside this task's declared surface, so the task is parked on that
decision rather than taken unilaterally.

## Host conditions

Every number below was measured on a host that could not be quieted:

- load average ~7.1 from unrelated processes (`suggestd` 97% CPU, a sync agent
  at 75%, `fileproviderd` at 74%);
- `/System/Volumes/Data` at 99% capacity, ~13.8 GiB free of 926 GiB.

A genuinely unloaded host would be faster; no run below was taken on one, and
none is reported as if it were.

## 1. Bounded diagnostic run

`go test -count=1 -timeout 30m -json ./cmd/curator`

- real exit code **0**
- package elapsed **533.463s**
- disk before 971350180 / used 922717112 / avail 14207832 KiB
- disk after&nbsp; 971350180 / used 922842400 / avail 14082548 KiB
- raw: `logs/diag-cmdcurator-01.json`

62 top-level tests. 15 take ≥1s and account for **531.1s**; the other 47 take
**1.66s** in total.

| test | seconds |
| --- | --- |
| TestStatusReportsCompiledCurrentnessAndFailsCheck | 184.57 |
| TestInstallAndUpgradeRepairCorruptCompiledState | 104.52 |
| TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck | 99.93 |
| TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall | 34.69 |
| TestStatusReportsATransitivelyResolvedCompiledCommand | 26.42 |
| TestGlobalStatusReportsATransitivelyResolvedCompiledCommand | 25.72 |
| TestGCRetainsAndReportsReferencedCompiledState | 16.49 |
| TestStatusReportsAnUnusableToolchainPerCompiledCommand | 13.26 |
| TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand | 12.72 |
| TestDryRunNeverClaimsACompletedCompilerCheck | 6.53 |
| TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands | 1.32 |
| TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands | 1.30 |
| TestCLIEndToEndInstallStatusAndTamperCheck | 1.30 |
| TestStatusAcceptsAnUnchangedLegacyMarkerSchema | 1.27 |
| TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands | 1.08 |

Every test above 6s drives a compiled installation. Nothing else in the package
is measurable.

## 2. Cost model derived from the run

Two atoms explain the whole package.

- **One read-only `status` / `global status` on a compiled scope: ~3.1s.**
  Six `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` subtests that
  only tamper state and then run `--json`, `--check`, and the human form take
  9.24–9.38s each — three CLI invocations at ~3.1s.
- **One real compiled install: ~12.5s.** The two subtests that additionally
  reinstall take 21.73s and 21.89s.

The same scope *without* a compiled command
(`TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands`: a full
global install plus four status invocations) takes **1.08s** in total, so all
non-build CLI machinery — manifest, closure, git archive, audit, registry, MCP,
adapters, transaction commit — is ~0.2–0.3s per operation. The compiled-command
path is the entire cost.

## 3. CPU attribution

`go test -count=1 -run '^TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand$' -cpuprofile ../logs/install.cpu ./cmd/curator`

- real exit code **0**, 13.570s; profile duration 12.88s, 13.18s samples
  (in-process CPU ≈ wall, so this is not I/O wait).

`go tool pprof -peek`:

```
6.99s 53.03%  godriver.(*Session).VerifyToolchain -> fingerprintToolchain
8.24s 62.52%  godriver.fingerprintToolchain
                3.61s 43.81%  os.(*Root).Open
                3.26s 39.56%  io/fs.WalkDir  (of which os.(*Root).Lstat 2.83s)
                1.34s 16.26%  godriver.copyWithContext   <- the only real work
```

**`fingerprintToolchain` is 62.5% of the package's CPU, and only 16% of that is
reading and hashing bytes.** The other 84% is `os.Root` path resolution.

Why it repeats: `fingerprintToolchain` walks and hashes the complete GOROOT
(here 14502 files / 203 MiB) on every call, and it is called

- once per read-only plan — `godriver.Probe` (`install/builddeps.go:154`),
  documented as "It does not memoize results";
- once per staging session — `godriver.Establish` (`install/builddeps.go:170`);
- three more times per staged build — `VerifyToolchain` at `godriver/build.go:148`,
  `:219`, `:253`, plus `install/plan.go:252`.

So a `status` pays one full GOROOT hash and an install pays four to five.
Dividing the profiled 8.24s by the measured ~1.65s per fingerprint gives ~5
fingerprints in that one test, which matches the call sites exactly.

## 4. Why the fingerprint is slow

`fingerprintToolchain` opens one `os.Root` over GOROOT and then, for every
entry, re-resolves the **full** root-relative path one `O_NOFOLLOW` component at
a time — once in `fs.WalkDir(root.FS(), …)`, once more in `root.Lstat(name)`,
and a third time in `root.Open(record.filesystem)`. At an average depth of ~5
that is ~15 `openat` calls per file instead of ~3.

Standalone measurement over the same GOROOT
(`.temp/TASK-260729-2kaopg/probe`, three rounds):

| walk strategy | seconds |
| --- | --- |
| plain `filepath.WalkDir` + `os.Open` (no rooted guarantee) | 0.516 / 0.541 / 0.522 |
| current strategy (`os.Root`, full path per access) | 1.600 / 1.636 / 1.678 |
| directory-scoped `os.Root` handles | 0.712 / 0.724 / 0.742 |

## 5. Quantified fix (prototyped, **not applied**)

The directory-scoped variant keeps the identical record set, the identical sort,
and the identical byte stream, and resolves each path component exactly once:
directories are descended through per-directory `*os.Root` handles, and both
phases reuse the handle of the directory they are inside. Sorted order groups a
directory's entries contiguously, so a single cached handle suffices.

- **`identical_digest=true` in every round** — the prototype computes the same
  SHA-256 as the current strategy over the same tree.
- **2.2x faster** (1.64s → 0.72s per full GOROOT fingerprint, −56%).
- The `O_NOFOLLOW`-per-component guarantee is preserved; resolving each
  component once rather than three times also narrows, not widens, the TOCTOU
  window between the walk, the `Lstat`, and the `Open`.

Projected effect: 62.5% of `cmd/curator` CPU × 56% ≈ **35% off the package**,
i.e. 533s → ~350s, comfortably under the 480s target and with headroom under the
default 10-minute timeout even when `./...` runs it concurrently with
`internal/install` (289s) and `internal/install/atomicity` (423s) — both of which
pay the same per-operation cost and would improve too.

The prototype lives in `.temp/TASK-260729-2kaopg/probe/main.go`. **No file in the
worktree was modified.**

## 6. Why the prescribed test-side levers cannot reach the target

- **Share immutable setup — already done.** The two largest tests each build one
  compiled fixture and one installation and then share it across 14 and 6
  subtests. The remaining per-subtest cost is three CLI invocations that *are*
  the assertions (`--json` row fields, `--check` exit code, human line). Removing
  any of them is the "weaken assertions" the directive forbids.
- **Task-owned caches/fixtures — no legal target.** The expensive state is (a)
  the GOROOT fingerprint, which is deliberately never memoized, and (b) the
  operation-private `GOCACHE` (`install/private.go`), which is a hermeticity
  property, not incidental setup. Copying a finished installed manager home
  between tests would require rewriting absolute paths baked into config,
  markers, receipts, and the consumer registry — a harness that hides real
  platform behavior rather than exercising it.
- **Split integration-test packaging — blocked by `package main`.** The tests
  live in `cmd/curator`, call `run()` in-process, and assert on unexported
  identifiers (`stateUpToDate`, `buildReport`, `exitOK`, `currentnessCodes()`).
  Go permits no second test binary in one directory, so a split means moving
  `run()` into an importable package and rewriting ~1500 lines of tests against
  string literals — strictly larger and riskier production surgery than the
  30-line fix above, and it loses the compile-time coupling to the constants, so
  "coverage remains equivalent" fails.
- **Trimming a task-owned GOROOT mirror — not viable.** GOROOT's file count is
  dominated by `src/` (11006 of 14502 files), which stdlib compilation needs.
  Removing `test/` and `api/` alone yields <25%, and a doctored toolchain is a
  toolchain no operator has.
- Masking routes (raising `-timeout`, sharding, `-run` exclusions, cached
  results) are forbidden and were not used.

## 7. Product deliverable re-verified (unchanged)

The global-status surface itself is untouched by this run and still green.

- `gofmt -l .` — exit **0**, no output.
- `git diff --check` — exit **0**.
- `go build ./...` — exit **0**.
- `go vet ./...` — exit **0**.
- `go test -count=1 -coverprofile=… ./cmd/curator -run '^TestGlobalStatus'`
  — exit **0**; 147.322s; broad package coverage 27.7%.
  - disk before 13784240 KiB avail; after 13799708 KiB avail.
- `go tool cover -func=…` — exit **0**:
  `cmdGlobalStatus` 96.9%, `globalStatusPlan` 100%, `globalStatusScope` 100%,
  `statusReport` 86.2%, `classifySkillBuilds` 81.8%, `checkFailed` 100%,
  `factsOf` 100%, `Describe` 100%, `plannedRows` 100%, `markerMoved` 100%.
- `command -v golangci-lint` — exit **1**, executable absent. Not installed:
  the directive forbids installing host software. The preserved implementer
  artifact `logs/lint-02.log` records `0 issues.` for the unchanged tree.

## 8. Not run, and why

- The two consecutive `go test -count=1 ./...` runs at the default timeout were
  **not** attempted. Nothing in the tree changed, so they would reproduce the
  tester's exit 1 at the `cmd/curator` 10-minute package timeout. Running them
  to re-record a known failure would burn ~20 minutes of a 99%-full disk for no
  new information.

## 9. Decision needed

Authorize **one** of:

1. **(recommended)** A separate, godriver-owned task to apply the
   digest-identical directory-scoped fingerprint walk in
   `internal/godriver/fingerprint.go`, re-attest the `go-v1` conformance
   vectors, and then re-run the default gate here. ~30 lines, no semantic
   change, ~35% off `cmd/curator` and material wins in `internal/install` and
   `internal/install/atomicity`.
2. Explicit authorization for **this** task to make that production change under
   its own review.
3. Accept the current `cmd/curator` duration and record an explicit CI timeout
   for the package — which the directive currently forbids.
