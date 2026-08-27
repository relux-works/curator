# BUG-260731-3gm8kc — review verdict

**Role:** reviewer · **Run:** `RUN-260731-4afbab` · **Date:** 2026-07-31
**Goal:** `GOAL-260731-de9afd` revision 1 (parent `GOAL-260731-f6f304` revision 1)
**Verdict:** **ACCEPTED** → `done`

Two earlier reviewer runs (`RUN-260731-0910bf`, `RUN-260731-7b9b43`) died at launch
with a transport error and produced no verdict. This is the first real review.
Nothing below is taken from the producer's summary: every claim was re-derived in
an isolated worktree or read back from real CI.

---

## 1. Isolation

The primary checkout was neither switched nor committed to. Two fresh detached
worktrees and two fresh conformance roots were created under
`.temp/BUG-260731-3gm8kc/review/`:

| Path | Contents |
| --- | --- |
| `review/curator-head` | `bd6ba08` — the PR 9 head under review |
| `review/curator-old` | `cfffd7cd` — Curator `main`, the old length gate |
| `review/spec-rc6` | curator-spec `2629aec` (PR 15 head) tarball |
| `review/spec-rc3` | curator-spec `00b1688` — Curator's own `SPEC_PIN` |
| `review/neg/*` | seven independently mutated roots (§4) |

`go1.25.5 darwin/arm64`, matching `go.mod`'s `go 1.25.5`.

## 2. The acceptance criterion, item by item

| AC clause | Verified how | Result |
| --- | --- | --- |
| interop gate accepts the rc.6 vector incl. `compiled-cache-miss-is-read-only` | `CURATOR_CONFORMANCE_ROOT=<rc.6> go test -run 'TestManagerLifecycleVectors\|TestManagerCompiledCacheMissDryRunVector' ./internal/interop` at `bd6ba08` | **PASS / PASS**, exit 0 |
| …and the old gate really did reject it | same command at `cfffd7cd` | **FAIL** — `golden_test.go:488: manager lifecycle vector is incomplete` |
| …without breaking the pinned rc.3 root | same command, rc.3 root, at `bd6ba08` | **PASS** + `SKIP: …publishes no compiled-cache-miss-is-read-only dry-run case` |
| by required case **name**, not list length | read `internal/interop/golden_test.go` diff; `requireLifecycleCases` + seven negative controls (§4) | **met, and provably stronger** |
| published on a pinnable Curator commit | `gh pr view 9` → head `bd6ba08acda3dc801512c408c759ac0ac6f79f26`, base `main`, OPEN, MERGEABLE | **met** |
| `implementations.yml` advances the pin to that commit | `gh api contents/.github/workflows/implementations.yml` at **both** PR heads | `ref: bd6ba08acda3…` on both — **met** |
| Implementations green on ubuntu / macOS / windows for **PR 15 and PR 14** | `gh pr checks 14`, `gh pr checks 15` | **8 of 8 green on each** — **met** |

Verified PR heads match the evidence packet exactly: PR 14 `b07ef1d51dc6…`,
PR 15 `2629aecff19a…`.

Full Implementations step command, run verbatim locally against the rc.6 root:

```
go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec
ok  internal/interop 0.547s | ok internal/closure 3.819s | ok internal/skillspec 1.185s   exit=0
```

**The acceptance criterion is met in full.**

## 3. Does the change fit the architecture?

Yes. `git diff --name-only cfffd7c..bd6ba08` touches exactly three files —
`internal/interop/golden_test.go`, `.github/ci/toolchain-identity.sh`,
`.github/ci/gate-selftest.sh`. No product code, no `cmd/curator`, no
`internal/godriver`, `internal/transaction` or `internal/runtimestore`.

The one design decision worth scrutiny is the rc.3-compatible skip. It is not an
ad-hoc escape hatch: `.github/ci/skip-classes.tsv` defines `root-content` —
*"the supplied conformance root publishes no such vector group. Recorded against
the root's identity in the candidate evidence"* — with policy `allow` and regex
`publishes no `. The new skip reason matches that regex verbatim, so it is
classified, written to `skips-observed.tsv` with the root identity, and cannot
silently shrink the suite. That is the repo's own documented mechanism for
consuming two protocol revisions from one gate. **The solution fits.**

The `TestManagerCompiledCacheMissDryRunVector` binding to `install.BuildOutcome`
is the non-tautological part: it ties the spec's published vocabulary to the
planner's real one, so a suite that names an outcome Curator cannot produce fails
by name (proven in §4, row 3).

## 4. Negative controls — re-derived independently

Seven mutations, applied by this review to its own copies of the real roots, run
against **both** gates. Rows 1–3 and 7 reproduce the producer's controls; rows
4–6 are additional probes this review added.

| # | Root | Mutation | NEW gate (`bd6ba08`) | OLD gate (`cfffd7c`) |
| --- | --- | --- | --- | --- |
| 1 | rc.6 | `project-upgrade` renamed | **1** — `dry-run cases omit [project-upgrade]` | 1 (incidental — case count) |
| 2 | rc.6 | `compiled-artifact-cache` dropped from the rc.6 case | **1** — `does not forbid compiled-artifact-cache` | 1 (incidental) |
| 3 | rc.6 | bogus `would-silently-succeed` build outcome added | **1** — `which the planner cannot produce` | 1 (incidental) |
| 4 | rc.6 | `missing-config-if-missing` bootstrap case renamed | **1** — `bootstrap cases omit [missing-config-if-missing]` | 1 (incidental) |
| 5 | rc.6 | `artifact_executed` flipped to `true` | **1** — `executes the artifact a dry run only plans` | 1 (incidental) |
| 6 | rc.6 | `compiled-cache-miss-is-read-only` **deleted** | **0** — focused test skips (see §6a) | **0** — same blind spot |
| 7 | **rc.3** | `project-upgrade` renamed | **1** — `dry-run cases omit [project-upgrade]` | **0 — accepted the rename unnoticed** |

**Row 7 is the decisive one and it reproduces.** On a root whose case *count* is
unchanged, the old gate waved a renamed required case straight through; the new
gate rejects it by name. The "incidental" annotations matter: on rows 1–5 the old
gate only failed because the rc.6 root has three dry-run cases, not because it
detected the mutation — it would have accepted every one of those mutations on an
rc.3-shaped root. The replacement is genuinely stronger, not merely more
permissive. **The teeth are real.**

## 5. The two escalated checklist items — adjudicated

### Item 5 — "require Curator plus PR 15 CI green"

PR 15 CI: **green, 8/8**. Curator PR 9 CI: **red in four jobs** (`Lint`,
`Test (ubuntu)`, `Race (ubuntu)`, `Test (windows)`). The producer's pre-existing
claim was **verified independently, not accepted**:

1. **Control branch composition.** `git diff --stat cfffd7cd 8bedf884` (=
   `ci/goenv-control-BUG-260731-3gm8kc`) is exactly
   `.github/ci/gate-selftest.sh` + `.github/ci/toolchain-identity.sh`.
   `git diff cfffd7cd 8bedf884 -- internal/interop/golden_test.go` is **empty** —
   the control provably does not contain the change under review.
2. **Signature identity.** `gh run view --log-failed` for run `30615765014`
   (PR 9) and run `30616027892` (control) are both **35 491 bytes**. Extracted
   and set-compared, the failure signatures are identical:
   `internal/godriver/controls_other.go:35:30: func (*controlDomain).destroy is unused`,
   `internal/transaction/namespace.go:310:6: func existingNamespaceAncestor is unused`,
   `vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput`,
   plus the three `exit code 1` job terminators. The only textual difference
   between the two logs is the embedded timestamp.
3. **Ownership.** `BUG-260731-11bpa4` (windows vet) and `BUG-260731-lepevi`
   (linux lane) both exist on the board and are both in status `development` —
   actively owned, in flight, not orphaned by this acceptance.

**Adjudication: the sub-condition "Curator CI green" is literally unmet, and it
is not this change's defect.** The item is left unchecked so the board record
stays truthful at a glance. It does not block acceptance: every job this change
can affect passes on PR 9 (`Interop conformance gate`, `Naming gate`,
`Gate self-test` ×3, `Test (macos)`, `Race (macos)`), and the four red ones fail
identically without the change present.

Worth stating plainly: this change **improves** Curator CI honesty rather than
degrading it. `main`'s `toolchain-identity.sh` asserted `go env GOENV = off`.
Verified on the real toolchain — `GOTOOLCHAIN=local GOENV=off go env GOENV`
prints **empty** on go1.25.5 — so every Go job on `main` died at that step. The
four failures were always there; this PR is what makes them visible.

### Item 8 — "Lint clean"

`golangci-lint run` over the whole module in the review worktree at `bd6ba08`
(after `git submodule update --init` for the `tuitestkit` replace target):
**`0 issues.`** `gofmt -l cmd internal`: no output. `go vet ./internal/interop/...`:
exit 0. The two CI `Lint` findings are linux-build-only, in files this change does
not touch, and reproduce on the control branch.

**Adjudication: the change is lint clean. Item 8 checked.** The repo's CI `Lint`
job is red for pre-existing reasons owned by `BUG-260731-lepevi`; that is recorded
here rather than hidden behind the tick.

## 6. Findings recorded, not blocking

### 6a. An rc.6 root that *drops* the new case passes silently (control row 6)

If a future root publishes `dry_run_cases` without
`compiled-cache-miss-is-read-only`, `requireLifecycleCases` does not demand it
(by design — rc.3 must still pass) and the focused test skips. **This is not a
regression:** the old length gate exits 0 on exactly the same root. And the skip
is ledgered as `root-content`, so it appears in `skips-observed.tsv` against the
root's identity rather than vanishing.

It is, however, strengthenable later: `conformance/v1/manifest.json` carries
`protocol_version` (`1.0.0-rc.6` vs `1.0.0-rc.3`), so the gate could make the
case mandatory once the root declares rc.6 or later. Out of scope here — the AC
asks for name-based assertion, which is delivered. **Recommended as follow-up.**

### 6b. `internal/install` multi-project binding gap — needs a board item

Confirmed genuinely outside this bug, and confirmed pre-existing:

```
CURATOR_CONFORMANCE_ROOT=<rc.6> go test -run TestAuthoritativeDryRunCasesMutateNothingPersistent ./internal/install
bd6ba08 → FAIL  dryrun_conformance_test.go:189: published dry-run scope "multi-project" has no executable binding
cfffd7c → FAIL  (identical)
```

Identical on head and on `main`, so this change neither caused nor worsened it.
`internal/install` is not in the Implementations package set (confirmed from
`implementations.yml`: `go test -count=1 -v ./internal/interop ./internal/closure
./internal/skillspec`), and Curator CI pins rc.3 where the case does not exist —
so it is invisible today. The producer's refusal to weaken the `default:` branch
into an unknown-scope skip was the right call; that would have unasserted a
published case, the exact failure mode this task exists to fix.

**But it is a live landmine**, not a curiosity: the moment Curator's `SPEC_PIN`
advances to rc.6, `internal/install` runs against a root that publishes
`scope: multi-project` and Curator's own Test lane goes red. A board search finds
it only in this task's prose — **it has no board item of its own.** Orchestrator
should raise one before `SPEC_PIN` moves.

### 6c. The pin names an unmerged PR head — follow-up required, not a blocker

`conformance/README.md` §4 wants a full immutable commit ID (satisfied —
`bd6ba08acda3dc801512c408c759ac0ac6f79f26`) that has passed its own required CI
(not satisfiable at any commit today, because `main` itself has never passed).
The task's own scope says *"Publish Curator and spec changes through their
existing task branches/PRs; do not tag or release"* — so no release is gated on
this pin right now, and the alternative is leaving Implementations red on both
PRs, which is strictly worse.

**Binding follow-up:** after Curator PR 9 merges, re-pin `implementations.yml` to
the resulting `main` commit on both PR 14 and PR 15, and land
`BUG-260731-lepevi` + `BUG-260731-11bpa4` before rc.6 is actually released.

### 6d. Housekeeping

`ci/goenv-control-BUG-260731-3gm8kc` was kept for this review. The comparison has
now been re-run and recorded here; the branch can be deleted.

## 7. Verdict

**ACCEPTED.** The acceptance criterion is met and verified against real CI on
both target PRs. The replacement gate is independently proven strictly stronger
than the length check it replaces, and it fits the repo's documented
`root-content` skip architecture rather than working around it. The producer's
two unchecked items were honest, and both resolve in the change's favour on
evidence this review re-derived rather than inherited.

Per the reviewer constraint, **no `commit_ack` is supplied**; the commit-owning
mover commits and makes the final transition with `commit_ack=scope_committed`.

Open items handed to the orchestrator: **6a** (optional gate strengthening via
`protocol_version`), **6b** (raise a board item for the `internal/install`
multi-project binding before `SPEC_PIN` advances), **6c** (re-pin after PR 9
merges), **6d** (delete the control branch).
