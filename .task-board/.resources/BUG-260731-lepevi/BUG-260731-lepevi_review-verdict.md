# BUG-260731-lepevi — reviewer verdict

**Verdict: ACCEPTED.**

**Reviewed:** PR [#11](https://github.com/relux-works/curator/pull/11), commit
`b2ac7d7` on `task/BUG-260731-lepevi-linux-lane`, base `main`.
**Reviewer run:** `RUN-260731-a087ac` (reviewer, Opus 5). Read-only; no code,
no board file and no PR state was modified by this run.
**Method:** every claim below was re-derived from the repository, the GitHub API
and re-run gates. Nothing is carried over from
`BUG-260731-lepevi_linux-lane-outcome.md` on trust.

---

## 1. Acceptance criteria

> Curator CI `Lint` and `Test (ubuntu-latest)` pass on main without weakening
> the unused check or the native control inventory carve-out.

| AC clause | Verdict | Evidence I obtained myself |
| --- | --- | --- |
| `Lint` passes | **met** | `gh pr checks 11` → `Lint pass 51s` (job 91123182081) |
| `Test (ubuntu-latest)` passes | **met** | same → `pass 1m36s` (job 91123182110) |
| unused check not weakened | **met** | §2 |
| inventory carve-out not weakened | **met** | §3 |

Both lanes were `FAILURE` at the same base in the baseline run
[30615765014](https://github.com/relux-works/curator/actions/runs/30615765014),
so the flip is attributable to this change and not to the PR 9 gate repair.

`Race (ubuntu-latest)`, `Test (macos-latest)`, `Race (macos-latest)`, all three
`Gate self-test` jobs, `Interop conformance gate` and `Naming gate` are green in
run [30620349565](https://github.com/relux-works/curator/actions/runs/30620349565)
as well. The sole red is `Test (windows-latest)` — see §5.

**Base/rebase, independently checked.** PR 9 merged at `09:40:53Z` into merge
commit `2b6ef21`. I compared trees via the GitHub API:
`2b6ef21`'s tree is `4f3d788…`, and `git rev-parse bd6ba08^{tree}` is the same
`4f3d788…`. The trees are identical, `gh pr view` reports `MERGEABLE`, and
`compare/main...b2ac7d7` lists exactly the 8 files of this change (ahead 1,
behind 1 — the merge commit only). So the green run at `bd6ba08 + change` is
content-equivalent to `main + change`; no rebase is required. The
`RUN-260731-dbbe43` checkpoint note already records this on the board.

---

## 2. Lint — the two `unused` findings

Both were removed as genuinely dead code. **No `//nolint`, no linter exclusion,
no `_ = fn` reference trick anywhere in the diff.** I ran
`.github/ci/no-broad-suppression.sh` myself → `ok`.

### `(*controlDomain).destroy` — `internal/godriver/controls_other.go`

I re-derived the call graph rather than accepting the claim. `grep -rn destroy
internal/godriver/` gives exactly two call sites: `controls_darwin.go:188` and
`controls_windows.go:161`. Each of those files defines its own `destroy`
(`controls_darwin.go:206`, `controls_windows.go:235`) and calls it from inside
its own `launch`. `workerclient.go` has `destroyBeforeExecution`, which is a
different function and never calls `domain.destroy`. `controls_other.go` is
`//go:build !darwin && !windows`, so on that build the method had no reachable
caller in any configuration. Deletion is correct; the remaining stubs in that
file (`close`, `launch`, `installedControls`) do have shared-code callers and
are untouched.

### `existingNamespaceAncestor` — `internal/transaction`

`grep -rn existingNamespaceAncestor --include=*.go` after the move: definition
plus two call sites, all three inside `namespace_case_darwin.go`. I read the
sibling implementations to confirm the stated reason rather than take it:
`namespace_case_windows.go` returns `true`/`false` from the platform contract
and `namespace_case_other.go` (`//go:build !darwin && !windows`) returns
`false`/`false` from the POSIX one — neither interrogates the filesystem, so
neither can need an ancestor. Moving the helper beside its only consumer is the
right call, not deleting it.

I also confirmed `namespace.go` still legitimately uses every import left behind
(`fmt`, `os`, `filepath` all still referenced), which the cross-builds
corroborate.

---

## 3. The compiled-build carve-out — the part that had to be judged, not measured

This is where the task could have gone wrong, so I checked the design, not just
the result.

**The predicate is bound to the authority.** `requireNativeControlInventoryPlatform`
tests `godriver.InventoryPlatform(runtime.GOOS) == ""`.
`InventoryPlatform` is at `internal/godriver/controls.go:200` and is the same
function the driver's own refusal path consults; `NativeControlInventoryVersion`
is the literal `rc5-native-control-inventory-v1` at `controls.go:25`. The guard
is therefore not a GOOS allow-list that can drift from the inventory — it *is*
the inventory. When a Linux record is added, the guard stops skipping with no
edit. This is the single most important property of the change and it holds.

**The carve-out is asserted, not obeyed.** New
`TestCompiledInstallFollowsTheNativeControlInventoryExactly` runs on all three
runners with `skip_allowed_on = -`. I verified both branches actually executed
on real hardware, from artifacts I downloaded myself:

* `test-evidence-ubuntu-latest/test/observed-cases.tsv` → `pass cmd/curator
  TestCompiledInstallFollowsTheNativeControlInventoryExactly`. The **uncovered**
  branch — refusal carrying `build_execution_control_unavailable`, zero
  published protected cache entries, non-zero `status --check` — ran on native
  Linux;
* `test-evidence-macos-latest/test/platform-cases.txt` → all seven `cmd/curator`
  ledger rows `ok`, and `skips-observed.tsv` contains **zero** `cmd/curator`
  rows. The **covered** branch ran on macOS and the guard is inert there.

Both sides are observed. Neither is argued from an analogue.

**Nothing is silently lost on Linux.** From the same ubuntu artifact:
`cmd/curator` = **56 pass, 6 skip, 0 fail**. I read all six carved-out bodies:
each begins with a real compiled install (`compiledProject` /
`compiledGlobalScope` / `newInstalledCompiledProject` + `install`, asserted
`exitOK`) and every assertion after it is about compiled state. There is no
non-compiled behaviour hiding inside them. General `gc` coverage — lock,
pruning, serialization — still runs on Linux from `cmd/curator/gc_test.go`; only
the compiled-specific `TestGCRetainsAndReportsReferencedCompiledState` is carved
out.

**The gate mechanics are what the outcome doc claims.** I read
`platform-case-gate.sh` rather than trusting the adversarial table. For the six
rows (`must_run_on=darwin,windows`, `skip_allowed_on=linux`,
`class=platform-control`): a skip on darwin/windows fails `listed(tol, goos)` →
`FATAL-not-tolerated` (line 216-220); a skip whose reason classifies as anything
other than `platform-control` → `FATAL-wrong-class` (line 221-225). The
narrowness is a property of the gate's code, not of a one-off experiment.

**The new skip-class row is narrow.** `rc5-native-control-inventory-v1 defines
no record for host ` is the inventory's own wording, and sits in a
`platform-control` block that already carried
`available only on (Windows|macOS) in rc5-native-control-inventory-v1`. It is
consistent with the established shape and cannot match a vague or unrelated
reason. In run 30620349565's ubuntu evidence, of **39** recorded skips, **zero**
are `UNCLASSIFIED` or `FATAL-*`.

**The open Linux qualification item is correctly related.** `TASK-260728-1skseh`
is the pre-existing item already cited in `.github/ci/platform-exclusions.tsv:17`;
the new ledger block and the README subsection both name it and state what has
to change when the inventory gains a Linux record. That is the relation the task
scope asked for, and it did not invent a new one.

**This is not a forced fit.** The repository already carves `internal/godriver`
out on Linux at whole-package granularity while still asserting
`TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` on that very runner — which
I confirmed still passes there, `pass` in
`test-evidence-ubuntu-latest/test/go-test-assert-godriver.json`. This change
applies the identical shape one level finer. No Linux execution binding was
fabricated and no unsupported execution is claimed.

---

## 4. Gates re-run by me, in the worktree

| Gate | Result |
| --- | --- |
| `gofmt -l cmd internal` | clean |
| `go vet ./...` (darwin) | `0` |
| `GOOS=linux go build ./...` | `0` |
| `GOOS=darwin go build ./...` | `0` |
| `GOOS=windows go build ./...` | `0` |
| `.github/ci/gate-selftest.sh` | `75 passed, 0 failed` |
| `.github/ci/ledger-consistency.sh` | `ok`, 56 rows across linux/darwin/windows |
| `.github/ci/no-broad-suppression.sh` | `ok` |

`golangci-lint` is not installed on this reviewer host, so I could not re-run the
linter locally. It is not needed: `Lint` on the real `ubuntu-latest` runner is
green on the PR and was `FAILURE` with exactly the two findings at the base, and
that is stronger evidence than a cross-target local run.

Commit signature verified independently: `git log -1 --format='%G?'` → `G`,
`Good "git" signature for oparin@me.com`.

---

## 5. Landing precondition — stated, not a defect in this work

`Test (windows-latest)` fails at step 7 `go vet`:
`vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined:
decodeHelperOutput` — read from job 91123182213's log, not from the outcome doc.
That is BUG-260731-11bpa4 / PR #10, which is still **OPEN**. This PR's diff is 8
files and touches nothing under `internal/runtimestore`, so the failure is not
attributable to it and the ownership boundary was respected exactly.

`Test (windows-latest)` was last green on **2026-07-14** (run 29298518437), i.e.
before the compiled-builds feature landed in `cfffd7c`, so the six cases'
`must_run_on=windows` claim has never been observed passing.

### Reviewer finding: what actually happens to those six on Windows

The implementer said the Windows claim was "a correct declaration, not an
observed pass" and left it there. That is honest but incomplete, and the answer
is obtainable: PR 10's run
[30620739038](https://github.com/relux-works/curator/actions/runs/30620739038)
got the Windows lane past `go vet`, so its
`test-evidence-windows-latest/test/observed-cases.tsv` records what those cases
really do. I downloaded and parsed it.

**All six FAIL on Windows today.** Not skip — fail. The reason, identical for
every one of them, from that run's `go-test.json`:

```
status_test.go:373: install = 1
error: go-v1 go_toolchain_missing: trusted GOROOT is not a real directory
error: go-v1 go_toolchain_missing: the trusted Go installation could not be
       resolved or verified; select a trusted Go installation with
       CURATOR_GO=<GOROOT>/bin/go (bin/go.exe on Windows) or GOROOT=<GOROOT>
```

This **strengthens** the accepted design rather than undermining it, and the
distinction matters:

* the failure is `go_toolchain_missing` — GOROOT resolution on the Windows
  runner — and has **nothing to do with `rc5-native-control-inventory-v1`**.
  `InventoryPlatform("windows")` returns `"windows"`, so
  `requireNativeControlInventoryPlatform` is correctly inert there and this PR's
  guard is not implicated;
* it is pre-existing, from `cfffd7c`, and observed on a branch that does not
  contain this change;
* it confirms narrowing those rows to `must_run_on=darwin` would have been the
  wrong call: it would have hidden a real Windows defect behind a ledger edit.
  Declaring the requirement is what keeps the gate naming them.

Ownership: those six are 6 of the 8 unowned `cmd/curator` Windows failures that
`BUG-260731-33v6zz` was created for (per the 1400 logbook entry). The other two
are `TestAuthoritativeUpgradeCasesAreExecutable` and
`TestDryRunNeverClaimsACompletedCompilerCheck` — the latter is a compiled case
this PR deliberately did **not** carve out (it does not need a completed
compilation) and is correctly absent from the ledger.

**Derived, not observed — flagged for BUG-260731-33v6zz:** the new
`TestCompiledInstallFollowsTheNativeControlInventoryExactly` takes its
*covered-host* branch on Windows and requires `install` to exit `exitOK`. Under
the same `go_toolchain_missing` defect it will fail there too, making it a 7th
red `cmd/curator` case on the Windows lane once PR 10 lands. That is a correct
test exposing an existing defect, not a new defect — but `BUG-260731-33v6zz`'s
scope should count it. This is inferred from the shared failure mode above; it
has not been executed on Windows.

**Consequence:** PR 11 cannot show an all-green required-CI set until PR 10
lands, and `Test (windows-latest)` will stay red past that until
`BUG-260731-33v6zz` fixes GOROOT resolution. Neither is in this bug's AC.
Landing order: PR 10 → PR 11 → confirm `Lint` and `Test (ubuntu-latest)` green
on `main`.

---

## 6. Non-blocking observations

None of these affects the verdict.

1. **`publishedCacheEntries` duplicates `cacheEntries`** —
   `cmd/curator/status_test.go:346` and `:368` are ~15 near-identical lines
   differing only in tolerating a missing cache root. `cacheEntries` could be a
   two-line wrapper that fatals when the new helper's root is absent. Test-helper
   tidiness only.
2. **The new skip-class pattern's narrowness is not regression-protected.** It
   was proven adversarially with synthetic streams that were then deleted, rather
   than added to `gate-selftest.sh`. Mitigating, and why this is not a finding:
   `gate-selftest.sh` covers skip-class *mechanics* per class rather than per
   pattern (only one of the four existing `platform-control` patterns has a
   dedicated case), and every CI run re-proves the property empirically in
   `skips-observed.tsv`.
3. **The outcome artifact never states the rebase decision** the orchestrator
   context asked for after PR 9 merged. The decision itself is correct (§1) and
   is recorded in the board notes by the `RUN-260731-dbbe43` checkpoint, so this
   is a documentation gap in one artifact, not a missing decision.

---

## 7. Checklist disposition

Checked by this review: *Implementation matches AC*, *Solution fits project
architecture*, *Tests green*.

**Left unchecked, deliberately:** *"Obtain independent Opus review and land only
after required CI is green."* The review half is now satisfied — this is that
independent Opus review. The **land** half is not, and cannot be until PR 10
clears the Windows lane. Ticking it would misreport the state.

*"If review does not accept the work — verdict evidence added and status routed
by the explicit verdict branches"* is left unchecked as not-applicable: the
review accepts.

---

## 8. Verdict

**ACCEPTED.** The acceptance criteria are met on the native Linux runner, both
lint findings were removed as genuinely dead code with no suppression of any
kind, and the compiled-build expectation was aligned with the authoritative
inventory through a guard bound to that inventory and asserted from both sides
on real runners. The solution matches the shape the repository already uses for
its whole-package exclusion, and no unsupported execution is claimed anywhere.

Acceptance evidence for the commit-owning mover: commit `b2ac7d7` is signed
(`%G?` = `G`) and `MERGEABLE` against `main` with no rebase required. Landing is
gated on BUG-260731-11bpa4 / PR #10 first. This reviewer run supplied no
`commit_ack` and did not merge.
