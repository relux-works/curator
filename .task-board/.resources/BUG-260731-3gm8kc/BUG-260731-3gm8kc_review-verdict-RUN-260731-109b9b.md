# BUG-260731-3gm8kc — review verdict (second independent review)

**Role:** reviewer · **Run:** `RUN-260731-109b9b` · **Date:** 2026-07-31
**Goal:** `GOAL-260731-de9afd` revision 1 (parent `GOAL-260731-f6f304` revision 1),
re-read at start and again at the pre-verdict checkpoint — unchanged, scope
`BUG-260731-3gm8kc`, no directives recorded.
**Verdict:** **ACCEPTED** → `done`

## 0. Standing of this review

The orchestrator context said two reviewer runs died at launch and this would be
the first real review. That is now stale: `RUN-260731-4afbab` completed a full
review and recorded ACCEPTED before this run started. This run therefore re-ran
the decisive checks **from scratch** in its own worktrees rather than reading the
earlier verdict as evidence, and added checks the earlier review did not make
(§2 CI test-level proof, §5 skip-ledger delta, §6 GOENV spelling matrix).
Everything below was produced by this run.

Isolation honoured: the primary checkout stayed on `agent/link-curator-skill-registry`
at `c06aa1a`, was not switched and was not committed to. Fresh detached worktrees
under `.temp/BUG-260731-3gm8kc/rev-109b9b/`:
`head` = `bd6ba08`, `old` = `cfffd7cd` (Curator `main`),
`spec-rc6` = curator-spec `2629aec`, `spec-rc3` = curator-spec `00b1688`
(Curator's own `SPEC_PIN`), plus mutated copies under `neg/`.
Toolchain `go1.25.5 darwin/arm64`, matching `go.mod`.

## 1. Local re-derivation — real exit codes

| # | Tree | Conformance root | Command | exit |
| --- | --- | --- | --- | --- |
| A | `bd6ba08` | rc.6 (`2629aec`) | `go test -run 'TestManagerLifecycleVectors\|TestManagerCompiledCacheMissDryRunVector' ./internal/interop` | **0** — both PASS |
| B | `cfffd7cd` | rc.6 | `go test -run TestManagerLifecycleVectors ./internal/interop` | **1** — `golden_test.go:488: manager lifecycle vector is incomplete` |
| C | `bd6ba08` | rc.3 (`00b1688`) | same as A | **0** — lifecycle PASS, focused test SKIP |
| D | `bd6ba08` | rc.6 | `go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec` (verbatim Implementations step) | **0** — all three `ok` |

B is the load-bearing control: the old gate really does reject the rc.6 root, so
the fix is not decorative. C's skip reason is
`this conformance root publishes no compiled-cache-miss-is-read-only dry-run case`.

## 2. Acceptance criterion — checked against real CI, at test granularity

- Curator PR 9 `relux-works/curator#9`, OPEN, base `main`, MERGEABLE,
  head `bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
- `implementations.yml` read back with `gh api` **at both PR heads**
  (`b07ef1d51dc6…` PR 14, `2629aecff19a…` PR 15): `ref: bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
- `gh pr checks 14` and `gh pr checks 15`: **8 of 8 green on each**, including
  `Implementations (ubuntu-latest | macos-latest | windows-latest)`.
- Those Implementations runs are for the current heads, not stale:
  run `30616098340` head_sha `b07ef1d5…` conclusion success;
  run `30616107892` head_sha `2629aecf…` conclusion success.
- **New in this review — the tests actually ran, they did not skip.**
  `CURATOR_CONFORMANCE_ROOT` is a job-level env in `implementations.yml`
  (`${{ github.workspace }}/conformance/v1`), and the job log for run `30616107892`
  shows on **all three** runners:

  ```
  === RUN   TestManagerLifecycleVectors
  --- PASS: TestManagerLifecycleVectors (0.00s)
  === RUN   TestManagerCompiledCacheMissDryRunVector
  --- PASS: TestManagerCompiledCacheMissDryRunVector (0.00s)
  ```

  So the green Implementations lane is a real pass of the new gate, not a
  root-unset skip.

**The acceptance criterion is met in full.**

## 3. Is the new gate genuinely stronger? — controls re-derived

Roots mutated by this run from its own copies, run against **both** gates.

| # | Root | Mutation | NEW gate (`bd6ba08`) | OLD gate (`cfffd7cd`) |
| --- | --- | --- | --- | --- |
| E | **rc.3** | `project-upgrade` renamed | **exit 1** — `dry-run cases omit [project-upgrade]; this root publishes [project-upgrade-renamed global-upgrade]` | **exit 0 — accepted the rename** |
| F | rc.6 | `compiled-cache-miss-is-read-only` deleted | exit 0 — focused test SKIPs (see §7a) | exit 0 — same blind spot |
| G | rc.6 | bogus outcome `would-silently-succeed` added to `reported_build_outcomes` | **exit 1** — `reports build outcome "would-silently-succeed", which the planner cannot produce` | n/a (case rejected on length) |

**E is decisive and it reproduces.** On a root whose case *count* is unchanged the
old length check waved a renamed required case straight through; the new gate
rejects it by name. G shows the `install.BuildOutcome` binding is not decorative:
the suite's published vocabulary is tied to the planner's real one.

Reading the diff confirms the replacement is strictly stronger everywhere the old
one asserted anything:

- launcher: old `len != 2`; new requires both names **and** still runs the
  per-case shape loop over every published case, so an added case is tolerated
  but not unchecked.
- bootstrap: old checked outcomes for 2 of 3 names; new requires all three names
  **and** adds the previously unasserted `missing-config-if-missing = created`.
- upgrade: old checked only `all-projects-deduplicate`; new requires all three names.
- dry-run: old `len != 2` (the bug); new requires `project-upgrade` +
  `global-upgrade` by name and keeps `len(ForbiddenPersistentEffects) >= 8` on
  every published case, including added ones.

## 4. Does it fit the architecture?

`git diff --name-only cfffd7c..bd6ba08` touches exactly three files:
`internal/interop/golden_test.go`, `.github/ci/toolchain-identity.sh`,
`.github/ci/gate-selftest.sh`. No product code.

The rc.3 skip is not an ad-hoc escape hatch. `.github/ci/skip-classes.tsv` already
defines, before this change:

```
root-content	publishes no 	allow	the supplied root does not publish this vector group
```

and the new skip reason matches that pattern verbatim, so the platform-case gate
classifies it instead of treating it as a newly-introduced silent skip.

## 5. Proof the skip is ledgered, not swallowed (new in this review)

Set-comparing the normalized `--log-failed` output of PR 9 run `30615765014`
against control run `30616027892`, the **only** non-noise difference is:

```
- Test (ubuntu-latest)  platform-case gate: 33 skips recorded in .temp/ci-evidence/test/skips-observed.tsv   (PR 9)
+ Test (ubuntu-latest)  platform-case gate: 32 skips recorded in .temp/ci-evidence/test/skips-observed.tsv   (control)
```

i.e. exactly **+1 skip**, the new focused test on Curator's rc.3 pin, and the same
job reports `platform-case gate: ok` / `test-gate: go test exit=1, platform-case gate exit=0`.
The skip is written to the evidence ledger against the root's identity and the
gate accepts it under an existing class. Everything else differed only in PIDs,
timings and temp-file UUIDs.

## 6. The GOENV repair is not a weakening (new in this review)

`bd6ba08` relaxes `toolchain-identity.sh` from `[ "$ge" = off ]` to `case "$ge" in off|'')`.
Probed on the real toolchain (`go1.25.5`), `go env GOENV` prints:

| GOENV in env | `go env GOENV` prints | gate |
| --- | --- | --- |
| `off` | *(empty)* | accept — correct, no per-user file |
| unset | `/Users/…/Library/Application Support/go/env` | **reject** |
| `` (empty string) | `/Users/…/Library/Application Support/go/env` | **reject** |
| `/tmp/x.env` | `/tmp/x.env` | **reject** |

The accepted empty spelling is reachable **only** from `GOENV=off`; neither an
unset nor an empty-string `GOENV` can produce it. So the relaxation admits exactly
the one true-negative it was written for, and `main`'s assertion was genuinely
broken — which is why every Go job on `main` died at that step. `gate-selftest.sh`
gains a stub case for it (74 → 75), and `Gate self-test` passes on all three
runners on PR 9.

Scope note: this CI-script repair is adjacent to the stated scope. It is accepted
as folded in rather than split out because without it PR 9 produces no Curator CI
signal at all, it is 12 lines, it is covered by a new self-test case, and it is
proven above not to loosen the gate. Recorded, not held against the change.

## 7. Findings recorded — none blocking

### 7a. An rc.6 root that *drops* the new case still passes (control F)

`requireLifecycleCases` does not demand `compiled-cache-miss-is-read-only` (it
cannot, while rc.3 is the pin) and the focused test skips. Not a regression — the
old gate exits 0 on the same root — and the skip is ledgered (§5). Strengthenable
later: `conformance/v1/manifest.json` carries `protocol_version`
(`1.0.0-rc.6` vs `1.0.0-rc.3`), so the case could be made mandatory once the root
declares rc.6+. Out of scope here; the AC asks for name-based assertion, delivered.

### 7b. `internal/install` multi-project binding gap — still has no board item

Re-run by this review against the rc.6 root:

```
go test -run TestAuthoritativeDryRunCasesMutateNothingPersistent ./internal/install
bd6ba08   → exit 1
cfffd7cd  → exit 1   (identical)
```

Pre-existing and untouched by this change, invisible today because
`internal/install` is not in the Implementations package set and Curator CI pins
rc.3. A `task-board grep multi-project` finds it only in this task's artifacts —
**it still has no board item of its own**, and it becomes a live Curator CI
failure the moment `SPEC_PIN` advances to rc.6. Raising it stays an orchestrator
action; this review does not expand scope to create it.

### 7c. The pin names an unmerged PR head

`conformance/README.md` §4 wants a full immutable commit ID (satisfied —
`bd6ba08acda3dc801512c408c759ac0ac6f79f26`) that passed its own required CI (not
satisfiable at any commit today, because `main` itself is red). The task scope
forbids tagging or releasing, so nothing shippable is gated on this pin right now,
and the alternative — leaving Implementations red on both PRs — is strictly worse.
**Binding follow-up:** after PR 9 merges, re-pin `implementations.yml` to the
resulting `main` commit on PR 14 and PR 15, before rc.6 is released.

### 7d. Housekeeping

`ci/goenv-control-BUG-260731-3gm8kc` (`8bedf884`) has now served two independent
comparisons and can be deleted.

## 8. The two escalated checklist items — adjudicated

### Item 5 — "…and require Curator plus PR 15 CI green"

PR 15: **green, 8/8**. Curator PR 9: **red in four jobs** (`Lint`, `Test (ubuntu)`,
`Race (ubuntu)`, `Test (windows)`). The "pre-existing" claim was verified by this
run, not accepted:

1. **Control composition.** `gh api` confirms `ci/goenv-control-BUG-260731-3gm8kc`
   head is `8bedf88402b87bc7da91565c59bfc644a72e6eff`.
   `git diff --stat cfffd7cd 8bedf884` = exactly `.github/ci/gate-selftest.sh`
   (+2) and `.github/ci/toolchain-identity.sh` (+12/−2).
   `git diff cfffd7cd 8bedf884 -- internal/interop/golden_test.go` is **empty** —
   the control provably does not contain the change under review.
2. **Same jobs fail.** PR 9 failed set = control failed set =
   {`Lint`, `Race (ubuntu-latest)`, `Test (ubuntu-latest)`, `Test (windows-latest)`}.
3. **Same signatures.** Both `--log-failed` payloads are 35 491 bytes; normalized
   set-comparison differs only in PIDs/timings/UUIDs and the +1 skip of §5. Shared
   findings: `internal/godriver/controls_other.go:35:30: func (*controlDomain).destroy is unused`,
   `internal/transaction/namespace.go:310:6: func existingNamespaceAncestor is unused`,
   `vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput`.
4. **Owned.** `BUG-260731-11bpa4` (windows vet) and `BUG-260731-lepevi` (linux lane)
   both exist and are both in `development` — in flight, not orphaned by this acceptance.

**Adjudication:** the sub-condition "Curator CI green" is literally false and is
not this change's defect. Every job this change can affect passes on PR 9
(`Interop conformance gate`, `Naming gate`, `Gate self-test` ×3, `Test (macos)`,
`Race (macos)`). The item stays **unchecked** so the board record is true at a
glance, and because 7c's re-pin residue is real; it does **not** block acceptance.

### Item 8 — "Lint clean" — checked, and re-verified here

`golangci-lint v2.12.2 run ./...` over the whole module at `bd6ba08`
(after `git submodule update --init` for the `tuitestkit` replace target):
**`0 issues.` exit 0**. `gofmt -l cmd internal`: no output. `go vet ./internal/interop/...`:
exit 0. `go build ./...`: exit 0. The two CI `Lint` findings are linux-build-only,
in files this change does not touch, and reproduce on the control.

## 9. Verdict

**ACCEPTED → `done`.** The acceptance criterion is met and verified against real
CI on both target PRs, down to the individual test results on all three runners.
The replacement gate is independently proven strictly stronger than the length
check it replaces — including the rc.3 rename the old gate accepted — and it
lands inside the repo's documented `root-content` skip class rather than working
around it. The folded-in GOENV repair is proven non-weakening on the real
toolchain. The producer's two unchecked items were honest and both resolve in the
change's favour on evidence re-derived here.

Per the reviewer constraint **no `commit_ack` is supplied**
(`version_control.confirm` is `configured=false` on this board, so the `done`
transition does not require one); the commit-owning mover remains responsible for
`commit_ack=scope_committed` if that policy is ever enabled.

Open items handed to the orchestrator: **7a** (optional `protocol_version`-gated
strengthening), **7b** (raise a board item for the `internal/install`
multi-project binding before `SPEC_PIN` advances — still missing), **7c** (re-pin
both spec PRs after Curator PR 9 merges), **7d** (delete the control branch).
