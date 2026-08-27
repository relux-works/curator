# TASK-260720-1nlmvv — cycle 2: answering the review verdict

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
(base `origin/main` 17804ce plus the composed accepted predecessor tree; nothing
staged, committed, or published). Platform: darwin/arm64, Go 1.25.5,
golangci-lint v2.4.0 (pinned).

Cycle 1 was returned with **CHANGES REQUESTED** and four blocking findings. This
document is the delta against cycle 1, finding by finding.

---

## 1. Every stable code is reachable through the real CLI

The verdict was that `corrupt`, `unsupported`, a missing or untested Go
toolchain, `unsupported-marker`, and `unsupported-build-driver` could not be
produced by any real invocation. Four separate causes, four separate fixes.

### 1a. A failed read-only plan no longer discards its compiled diagnostics, and reporting is no longer a verdict

`cmdStatus` skipped `statusReport` entirely when the dry-run plan failed, so any
refusal produced stderr text and an empty `status --json`.

The verdict also required keeping ordinary `status` reporting separate from
`--check` failure semantics, so the fix needed a truthful notion of "the report
is complete". `BuildPlan.Complete()` (surfaced as `Result.BuildsComplete`) is
true exactly when the plan derived a row for **every** command the closure
activates — which includes a plan that then failed, because a refusal *is* a
per-command verdict. It is false when planning died before describing them all,
so a partial report can never masquerade as a whole one.

```go
if result.Status == "failed" {
    if !result.BuildsComplete {          // legacy path, byte-for-byte unchanged
        printResult(result); exitCode = exitFail; continue
    }
    printStatusRefusal(result)           // same detail, as warning:, on stderr
}
```

So plain `status` now exits **zero** when it produced the report it was asked
for, and `--check` is the only surface that turns a non-current verdict into a
non-zero exit. `printResult` was split for the same reason: this path publishes
the stable rows on stdout, so `status --json` stays one document instead of plan
lines interleaved with JSON.

Pinned from both sides.
`TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands`: an
unresolvable declaration still prints `[]`, an error, and exit 1.
`TestIncompletePlansNeverClaimACompleteDiagnostic`: a script-only scope and a
plan that failed inside per-command planning both report incomplete.
`TestStatusReportsAnUnusableToolchainPerCompiledCommand`: exit 0, rows on
stdout, `warning:` and no `error:` on stderr, and `--check` exit 1.

### 1b. `corrupt` is no longer blocking (see finding 2)

A corrupt entry is now a rebuildable miss, so a dry run completes and
`status` reports `corrupt-build-cache` for it. Reached end to end by
`TestStatusReportsCompiledCurrentnessAndFailsCheck/protected cache entry cannot
be interpreted`, which rewrites the published artifact bytes of the real
protected entry.

### 1c. A refused toolchain now plans an inventory

`planBuilds` used to return before a single row existed. It now records one row
per active command with the new outcome `toolchain-unavailable`, the stable
`go-v1` boundary code, and **no** target, key, or cache verdict — the plan still
fails, so nothing is claimed that was never derived.

That gives the vocabulary a code it was missing: `unusable-build-toolchain`,
whose `cause` is the driver's own stable code (`go_toolchain_missing`,
`unsupported_go_family`, …).
`TestStatusReportsAnUnusableToolchainPerCompiledCommand` drives it through the
real CLI with an unusable `CURATOR_GO` and asserts the row, the cause, the
demoted skill, the human line, and that no invented identity is published.
`TestToolchainFailurePlansAnInventoryOfEveryActiveCommand` pins the same at the
plan boundary.

### 1d. A refused marker now says *why*

`marker.Read` is a strict validator that returns nothing but "refused", so
`invalid-marker` swallowed two distinguishable cases. `markerRefusal` is a
**presentation-only second look at the same bytes**: it never accepts a marker
the reader rejected, it only decides which stable code the operator is told.

| Refusal | Code |
|---|---|
| absent, unreadable, not JSON, no `schema_version` | `invalid-marker` |
| `schema_version` outside the readable set | `unsupported-marker` |
| a recorded build driver outside `go-v1` | `unsupported-build-driver` |
| readable schema, still not a valid marker | `invalid-marker` |

It is used by both surfaces — `scopeStatusDrift` and `classifySkillBuilds` — so
the declared-skill map and the compiled rows agree. The dead schema branch in
`scopeStatusDrift` (unreachable since `marker.Read` already rejects unknown
schemas) was removed rather than left as decoration.
`TestMarkerRefusalSeparatesUnsupportedFromInvalid` asserts, for every payload,
that the reader still refuses it *before* checking the code — a case that only
passes because the classification is real is worth nothing.
Three of them also run end to end through `curator status`.

### Reachability, stated honestly

Reached end to end through the CLI in `cmd/curator/status_test.go`:
`up-to-date`, `not-installed`, `needs-install`, `content-drift`,
`invalid-marker`, `unsupported-marker`, `unsupported-build-driver`, `current`,
`build-command-drift`, `build-context-exposed`, `build-source-drift`,
`build-input-drift` (all three causes), `unusable-build-toolchain`,
`missing-build-artifact`, `corrupt-build-receipt`, `build-artifact-drift`,
`corrupt-build-cache`, `untrusted-build-cache`.

Not reachable from a CLI test on this host, and why:

- **`unsupported-build-platform`** — `buildcache` protection support is a
  compile-time property of the host (`protection_unix.go` /
  `protection_windows.go` / `protection_unsupported.go`), and the CLI exposes no
  seam to override it. Proven as far as it can be:
  `TestUnsupportedCacheProtectionFailsClosed` now drives a real
  `install.Project` (dry run *and* real run) through the production
  `Options.Build` seam and asserts the refused command still travels out through
  `Result.Builds` with its reason — that is the exact data path `cmdStatus`
  consumes — and the outcome-to-code mapping is pinned separately.
- **`unresolvable`** — pre-existing declared-skill code, unchanged by this task.
- **`build-state-changed`** — a race by definition.
  `TestStatusReportMarksCompiledStateThatMovedDuringTheCheck` now really moves
  the marker between the two fingerprints instead of injecting a fake digest.
- **`unknown-build-state`** — the deliberate version-skew guard. It exists so a
  planner outcome from a newer manager fails closed instead of reading as
  current, and it is unreachable *by construction* while the planner vocabulary
  is closed. Documented as such rather than dressed up as reachable.

---

## 2. Install and upgrade repair corrupt compiled state

`BuildCorrupt.blocking()` returned before `stageBuilds`, so a corrupt protected
entry bricked the command permanently — while the README claimed it was rebuilt
and `buildcache.Publish` already had the quarantine-and-replace path for exactly
this case. The story decomposition also lists "corrupt or stale artifact
rebuild" as in-scope behaviour.

`buildable()` and `blocking()` were re-cut along the only line that matters:

- **buildable** — every state a rebuild resolves: cold miss, untrusted
  provenance, **corrupt**.
- **blocking** — only what a rebuild cannot resolve: `unsupported` (no provable
  boundary to publish into) and `toolchain-unavailable` (nothing can be built).

Nothing else moved: the rebuild still happens after every manifest, closure,
collision, requirement, audit, registry, and moved-tag gate, still stages
privately, and still publishes under the manager-home lock, where `Publish`
quarantines the unusable entry instead of adopting or permission-repairing it.

`TestInstallAndUpgradeRepairCorruptCompiledState` is the end-to-end proof, over
the full matrix of {`install`, `upgrade`} × {corrupt receipt bytes, corrupt
artifact bytes}. Each case:

1. corrupts real published cache bytes and asserts `status` reports
   `corrupt-build-cache`;
2. runs the command with an unusable toolchain and asserts the gate refuses
   first — the live cache is byte-identical (full-tree digest), the installed
   marker's build entry is unchanged, and `SKILL.md` is still there;
3. runs it for real and asserts the reported repair, an unchanged logical key,
   exactly one live protected entry, and `status --check` back to zero.

Two sibling tests changed with the behaviour, not around it:
`TestCorruptCacheEntryFailsBeforeAnyPersistentMutation` became
`TestCorruptCacheEntryIsRebuiltAndNeverReused` (same assertions as its untrusted
twin, including that a corrupt entry never exposes a reusable artifact path),
and the two `private_test.go` cases that used `Corrupt` merely as a way to force
a planning failure now use `Unsupported`, which still is one.

---

## 3. A logical-key mismatch no longer claims a cause it cannot prove

`build-target-mismatch` labelled *every* opaque key mismatch as target or
toolchain drift. The cache key is one digest over the whole `buildmeta.Input`,
and the marker records no prior input, so that was a guess.

The code is now **`build-input-drift`**, which names the fact, plus a stable
`cause` subcode that says exactly how much the marker can prove:

| Cause | Evidence |
|---|---|
| `build-root` | the marker's own `build_roots` do not contain the activated root |
| `target` | the marker's own `artifact_path` is not the path this target derives |
| `unattributed` | the marker records nothing that could attribute it further |

`cause` is a general refinement field on the row (`"cause"` in JSON,
`cause=` in the human line); `unusable-build-toolchain` carries the driver's
boundary code in the same field.

`TestBuildInputDriftIsAttributedOnlyAsFarAsTheMarkerProves` is the adversarial
pass: it starts from a real, valid `buildmeta.Input`, moves exactly one
component, derives the **real** key for the moved input, asserts the key
actually changed (so a case cannot pass vacuously), and asserts the cause.
Build root → `build-root`. Target GOOS → `target`. Source directory, target
architecture, target tuning, toolchain release, and toolchain content →
`unattributed`, because none of them leaves evidence in a marker. Build schema
version and the fixed build policy cannot be varied at all — this manager
refuses to derive a key for them — so they are covered as what they really look
like in production: an opaque key some other manager version recorded.

All three causes are also reached end to end through `curator status`.

---

## 4. Redaction covers embedded, URI, and error-path forms

`RedactDiagnostic` split on whitespace and only inspected token starts, so
`source=/private/cache/x`, `file:///private/cache/x`, and
`error=C:\Users\name\cache` survived. It is now a single left-to-right scan
that recognises a path start after any opener (`= : " ' \` ( [ { < > , ; |`) or
whitespace, in Unix, URI, Windows-drive, and UNC form, ends the run at a closer,
and drops trailing sentence punctuation. Relative declaration state
(`root=assets/build-tool`, `target=linux/amd64`) is untouched, because a path
character that follows an ordinary character never starts a path. The
invalid-UTF-8, control- and format-rune, single-line, and 240-rune guarantees
are unchanged.

The second half of the finding was the error path. `Result.failBuild` is the new
boundary for every build-phase failure: `planBuilds`, `stageBuilds`, the blocked
message that interpolates a raw cache reason, and the plan-release warning all
go through the same bounded, redacted rendering. `BuildDiagnostic` still carries
the stable code untouched — a code is not untrusted text.

`TestBuildFailuresAreRedactedInTheResult` pins the error surface, nine new
`TestRedactDiagnosticReplacesEveryAbsoluteLocation` cases pin the forms, and
`TestBuildReportsNeverPublishAnAbsolutePath` grew the embedded, URI, and
multi-location reasons — and now also forbids the leaked *user name*, not just
the path prefix.

---

## Additional scope adjudication

- **Transitive compiled status** — the reviewer asked for a real integration
  test rather than indirect coverage.
  `TestStatusReportsATransitivelyResolvedCompiledCommand` builds a consumer
  skill that reaches the compiled provider through a runtime-mode dependency,
  so the project declares only the consumer. The provider's build row is
  reported and classified, the provider never appears in the declared-skill
  map, and removing its cache entry makes `status --check` fail through the
  compiled surface alone.
- **`curator global status`** — recorded as an explicit contract-level
  exclusion, as the verdict permits. README now states plainly that compiled
  diagnostics are a project-scope surface and that `global status` reports
  declared-skill currentness only. Tracked as **`TASK-260729-2kaopg`**
  (report-compiled-currentness-in-global-status), which also has to decide
  whether that command may start running a read-only plan at all — today it
  only reads markers, and planning would pull the audit and registry read-only
  gates into it.
- **Compiled install idempotence** — tracked as **`TASK-260729-3jku56`**
  (make-repeated-compiled-install-idempotent). Still owned by the install
  staging layer, not by this presentation task.

---

## Evidence

Reported exit codes are the real status of each command, run standalone. Raw
logs are in the attached `TASK-260720-1nlmvv_gate-evidence-cycle-2.tar.gz`.

The three `-race` gates were completed by the cycle-2 continuation run after the
first sweep hit `ENOSPC` and the restarted run was killed before its result
existed. They were run sequentially, each as a standalone process, on the same
tree — no `.go` file is newer than the `go test ./...` log above. Details in
`TASK-260720-1nlmvv_gate-evidence-cycle-2-continuation.md`.

### macOS (darwin/arm64), primary

| Gate | Exit |
|---|---|
| `gofmt -l .` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1 -timeout 60m` (40 packages) | 0 |
| `golangci-lint run ./...` (pinned v2.4.0) | 0, **0 issues** |
| `go test ./cmd/curator -count=1 -timeout 40m -v` (54 top-level tests) | 0 (495.5s) |
| `go test -race -timeout 60m ./internal/install -count=1` | 0 (764.3s) |
| `go test -race -timeout 60m ./cmd/curator -count=1` | 0 (651.9s) |
| `go test -race -timeout 30m ./internal/godriver -count=1` | 0 (46.5s) |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./...` | 0 |

Slowest packages in the `./...` run: `internal/install/atomicity` 611.6s,
`cmd/curator` 609.3s, `internal/install` 432.7s.

### Expected-red gate, reported as failing

`GOOS=windows GOARCH=amd64 go vet ./...` exits **1**:

```
vet: internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput
```

Pre-existing and sibling-owned (`internal/runtimestore`, `TASK-260720-29hi1h`).
Measured, not assumed: the identical command on the untouched accepted base
(`.temp/TASK-260720-1ljev5/worktree`) also exits **1** with the identical
message. Excluding that one package, this tree's Windows vet exits **0**.

### Native Windows

Host `win`, Microsoft Windows NT 10.0.19045.0, `desktop-3pbo632\admin`. The host
has no Go toolchain and no `git`, so the test binaries were cross-compiled on
macOS and executed natively. Binary provenance was verified before the run
(`strings` finds this cycle's new strings in both binaries).

| Suite | Exit | Result |
|---|---|---|
| `install.test.exe` redaction suite | **0** | 6 tests, 14 subtests, including the Windows-drive, UNC, embedded, and URI redaction cases and the new error-path assertion |
| `curator.test.exe` platform-neutral classification suite | **0** | 19 tests pass; `TestInputCausesAreDistinctAndDocumented` skips with an explicit reason (README is not reachable from a standalone binary), by design |

The full `curator.test.exe` package is still unrunnable there: every CLI-level
test builds a git repository fixture, and the compiled ones also need a real Go
toolchain. That limit is pre-existing and already tracked for
`TASK-260720-1pvfj5`.

### Not run

Native Linux execution. The change is platform-neutral presentation and planning
code; `GOOS=linux` build and vet are both clean, and the only platform-specific
behaviour it touches — absolute-path redaction in Unix, Windows-drive, and UNC
form — passes on both macOS and native Windows.

---

## Findings worth carrying forward

1. **A corrupt protected cache entry used to be unrecoverable.** Not a
   presentation bug: `blocking()` refused before staging, and Curator has no
   repair command, so the only way out was deleting the entry by hand. Fixed
   here; worth a line in the rc.4 conformance task, because the planner
   vocabulary (`corrupt` is now a buildable miss) is part of the observable
   contract.
2. **The cache key cannot attribute drift, and it never could.** The install
   marker records no prior build input. Any future work that wants a truthful
   "the toolchain moved" verdict has to persist the input components, which is a
   marker-schema change, not a status change.
3. **`GOOS=windows go vet` is red on the accepted base**, in
   `internal/runtimestore` only. Relevant to `TASK-260720-1pvfj5`; measured
   again this cycle on the untouched base.
4. **The Windows validation host still has neither `git` nor a Go toolchain**,
   so every CLI-level test that builds a repository fixture is unrunnable there
   — including one that predates this story. Installing `git` there would
   recover that coverage. Relevant to `TASK-260720-1pvfj5`.
5. **PowerShell 5.1 mangles `-test.run` when invoking a native binary.** The
   argument arrives as `-test`, and the test binary exits 2 with a usage dump
   that reads like a product failure. Drive Go test binaries on that host
   through `cmd.exe /c` (or `--%`). Cost half an hour this cycle; worth knowing
   for any future native-Windows gate.
6. **`internal/install` and `internal/install/atomicity` both exceed the default
   `go test` timeout under load** — 427s and 577s respectively without `-race`.
   CI has to pin a generous timeout or the failure reads like a hang. Relevant
   to `TASK-260720-1pvfj5`.
7. **This suite is disk-hungry enough to fail as a product bug.** A `-race`
   sweep on a host with ~9 GiB free filled the disk, and three
   `internal/install` tests failed with `no space left on device` reported
   through `Result.Errors` as an ordinary installation failure. Freeing the Go
   build cache and re-running made all three green. CI needs a free-space floor
   (the Go build cache alone reached 8.9 GiB here), and the failure text is
   worth recognising: it reads exactly like a transaction-journal defect.
   Relevant to `TASK-260720-1pvfj5`.
