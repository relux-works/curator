# BUG-260731-3gm8kc — independent re-verification (third reviewer run)

**Role:** reviewer · **Run:** `RUN-260731-7d213b` · **Date:** 2026-07-31
**Conclusion:** re-verification **CORROBORATES ACCEPTED**.
**Run disposition:** this run was **cancelled by operator directive** at
`2026-07-31T09:25:11Z` (`RUN-260731-7d213b:cancel:4762a0`, requested by
`codex-orchestrator`: *"Duplicate managed reviewer superseded by bare Opus
RUN-260731-685cd6 because native-goal acknowledgement restarts were discarding
review work."*). Its successor `RUN-260731-685cd6` was itself cancelled at
`09:26:52Z`. The accepted verdict of record therefore remains
`RUN-260731-109b9b`; this document is **corroborating evidence, not a competing
verdict**, and claims no satisfaction of `GOAL-260731-de9afd` on its own behalf.

The board item was found in `done`, briefly moved to `reviewing` by this run's
mandated first command, and is back in `done` — the state this re-verification
independently agrees with. No board damage remains.

## 0. Why this run re-derived instead of reading

The orchestrator context claimed the two prior reviewer runs died at launch and
that this would be the first real review. That was already stale: two full
verdicts (`RUN-260731-4afbab`, `RUN-260731-109b9b`) existed. Everything below was
produced **by this run**, from scratch, in its own worktrees. §3 controls H, I
and J are new — no earlier review ran them.

Isolation honoured: primary checkout stayed on `agent/link-curator-skill-registry`
at `c06aa1a`, never switched, never committed to. Detached worktrees under
`.temp/BUG-260731-3gm8kc/rev-7d213b/`: `head`=`bd6ba08`, `old`=`cfffd7cd`
(Curator `main`), `spec-rc6`=curator-spec `2629aec`, `spec-rc3`=curator-spec
`00b1688`, plus mutated roots under `neg/`. Toolchain `go1.25.5 darwin/arm64`.

## 1. Local re-derivation — real exit codes

| # | Tree | Root | Command | Result |
| --- | --- | --- | --- | --- |
| A | `bd6ba08` | rc.6 | `go test -run 'TestManagerLifecycleVectors\|TestManagerCompiledCacheMissDryRunVector' ./internal/interop` | **PASS ×2**, `ok` |
| B | `cfffd7cd` | rc.6 | `go test -run TestManagerLifecycleVectors ./internal/interop` | **FAIL** `golden_test.go:488: manager lifecycle vector is incomplete` |
| C | `bd6ba08` | rc.3 | same as A | lifecycle **PASS**, focused **SKIP** |
| D | `bd6ba08` | rc.6 | `go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec` (verbatim Implementations step) | **exit 0**, all three `ok` |

**B is load-bearing:** the old gate genuinely rejects the rc.6 root, so the fix is
not decorative. C's skip reason is
`this conformance root publishes no compiled-cache-miss-is-read-only dry-run case`.

Root contents confirmed directly from curator-spec:
rc.6 `dry_run_cases` = `[project-upgrade, global-upgrade, compiled-cache-miss-is-read-only]`;
rc.3 = `[project-upgrade, global-upgrade]`.

## 2. Acceptance criterion — checked against real CI

- Curator PR 9, OPEN, base `main`, MERGEABLE, head
  `bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
- Both commits **GPG-signed** — `git log --format='%h %G?'` returns `G` for
  `fee35c8` and `bd6ba08` (key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`).
- `implementations.yml` read via `gh api` **at both PR heads**
  (PR 14 `b07ef1d5…`, PR 15 `2629aecf…`): `ref: bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
- `gh pr checks 14` → **8/8 pass**. `gh pr checks 15` → **8/8 pass**. Both include
  `Implementations (ubuntu-latest | macos-latest | windows-latest)`.
- Runs are for the current heads, not stale: run `30616098340` headSha
  `b07ef1d5…` conclusion `success`; run `30616107892` headSha `2629aecf…`
  conclusion `success`.
- **The tests really ran.** Job log for run `30616107892` on **all three** runners:
  `--- PASS: TestManagerLifecycleVectors` and
  `--- PASS: TestManagerCompiledCacheMissDryRunVector`. Not a root-unset skip.

**The acceptance criterion is met in full.**

## 3. Is the new gate genuinely stronger? — four controls, three of them new

Every mutation below **keeps case counts identical**, so the old length gate is
structurally blind to all of them. Roots mutated by this run from its own copies.

| # | Group | Mutation on rc.3 root | NEW gate (`bd6ba08`) | OLD gate (`cfffd7cd`) |
| --- | --- | --- | --- | --- |
| E | dry-run | `project-upgrade` renamed | **exit 1** — `dry-run cases omit [project-upgrade]` | **exit 0 — accepted** |
| H *(new)* | bootstrap | `missing-config-if-missing` renamed | **exit 1** — `bootstrap cases omit [missing-config-if-missing]` | **exit 0 — accepted** |
| I *(new)* | bootstrap | `missing-config-if-missing` outcome flipped to `silently-skipped`, **name kept** | **exit 1** — `bootstrap outcomes: …` | **exit 0 — accepted** |
| J *(new)* | upgrade | `global-closure` renamed | **exit 1** — `upgrade cases omit [global-closure]` | **exit 0 — accepted** |

**I is the sharpest new result.** It does not rename anything — it corrupts the
*outcome* of a case whose name is intact. The old gate never asserted
`missing-config-if-missing` at all (it only checked two of the three bootstrap
outcomes), so it waves a bootstrap case that silently skips config creation
straight through. The new gate catches it. That assertion is net-new teeth, not a
restatement.

Reading the diff, the replacement is stronger everywhere the old one asserted
anything, and the one thing it deliberately relaxes — tolerating *added* cases —
is exactly the bug being fixed:

- **launcher:** old `len != 2`; new requires both names **and** still runs the
  per-case shape loop over every published case, added ones included.
- **bootstrap:** old checked outcomes for 2 of 3 names; new requires all three
  names **and** adds the previously unasserted `missing-config-if-missing = created`
  (control I).
- **upgrade:** old checked only `all-projects-deduplicate`; new requires all three
  names (control J).
- **dry-run:** old `len != 2` — the bug; new requires `project-upgrade` +
  `global-upgrade` by name and keeps `len(ForbiddenPersistentEffects) >= 8` on
  every published case, added ones included.

## 4. Fit with the project

`git diff --name-only cfffd7cd bd6ba08` = exactly three files —
`internal/interop/golden_test.go`, `.github/ci/toolchain-identity.sh`,
`.github/ci/gate-selftest.sh`. **Zero product code.**

The rc.3 skip is not an ad-hoc escape hatch, and this run proved it by
provenance rather than by reading intent: `git diff cfffd7cd bd6ba08 --
.github/ci/skip-classes.tsv` is **empty** — the ledger is untouched by this
change — and it already contained, before it,
`root-content <TAB> publishes no  <TAB> allow`. The new skip reason matches that
pre-existing pattern verbatim, so `platform-case-gate.sh` classifies it instead
of flagging a newly-introduced silent skip.

## 5. The GOENV repair is not a weakening

`bd6ba08` relaxes `toolchain-identity.sh` from `[ "$ge" = off ]` to
`case "$ge" in off|'')`. Probed on the real toolchain (`go1.25.5 darwin/arm64`):

| GOENV in env | `go env GOENV` prints | gate |
| --- | --- | --- |
| `off` | *(empty)* | accept — correct, no per-user env file |
| unset | `/Users/…/Library/Application Support/go/env` | **reject** |
| `` (empty string) | `/Users/…/Library/Application Support/go/env` | **reject** |
| `/tmp/x.env` | `/tmp/x.env` | **reject** |

The accepted empty spelling is reachable **only** from `GOENV=off`; neither an
unset nor an empty-string `GOENV` can produce it. The relaxation admits exactly
the one true negative it was written for, and `main`'s assertion was genuinely
broken — which is why every Go job on `main` died at that step. `gate-selftest.sh`
gains a stub case for it, and `Gate self-test` passes on all three runners on PR 9.

Scope note: this CI-script repair is adjacent to the stated scope. Accepted as
folded in — without it PR 9 emits no Curator CI signal at all, it is 12 lines, it
is covered by a new self-test case, and it is proven above not to loosen the gate.

## 6. The two escalated checklist items — adjudicated independently

### Item 5 — "require Curator plus PR 15 CI green"

PR 15: **green, 8/8**. Curator PR 9: **red in four jobs** (`Lint`,
`Race (ubuntu-latest)`, `Test (ubuntu-latest)`, `Test (windows-latest)`). The
"pre-existing" claim was **verified here, not accepted**:

1. **Control composition.** `ci/goenv-control-BUG-260731-3gm8kc` head =
   `8bedf88402b87bc7da91565c59bfc644a72e6eff`.
   `git diff --stat cfffd7cd 8bedf884` = exactly `.github/ci/gate-selftest.sh`
   (+2) and `.github/ci/toolchain-identity.sh` (+12/−2).
   `git diff cfffd7cd 8bedf884 -- internal/interop/golden_test.go` is **empty** —
   the control provably does not contain the change under review.
2. **Same jobs fail.** Control run `30616027892` failed set = PR 9 failed set =
   {`Lint`, `Race (ubuntu-latest)`, `Test (ubuntu-latest)`, `Test (windows-latest)`}.
3. **Same signatures**, extracted from both `--log-failed` payloads and normalized:
   - `internal/godriver/controls_other.go:35:30: func (*controlDomain).destroy is unused`
   - `internal/transaction/namespace.go:310:6: func existingNamespaceAncestor is unused`
   - `vet: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput`

   Identical counts and identical texts on both runs.
4. **Owned and in flight.** `task-board q 'list(type=bug)'` confirms
   `BUG-260731-11bpa4` (curator-windows-test-vet-compile-break) = `development`
   and `BUG-260731-lepevi` (curator-main-ci-red-linux-lane) = `development`.

**Adjudication:** the sub-condition "Curator CI green" is literally false and is
**not this change's defect**. Every job this change can affect passes on PR 9
(`Interop conformance gate`, `Naming gate`, `Gate self-test` ×3, `Test (macos)`,
`Race (macos)`). The item correctly stays unchecked so the board reads true at a
glance; it does not block acceptance. The AC itself asks for **Implementations**
green on PR 14 and PR 15 — satisfied 3/3 on both.

### Item 8 — "Lint clean"

At `bd6ba08` (submodules initialized): `gofmt -l cmd internal` → no output;
`go build ./...` → exit 0; `go vet ./internal/interop/...` → exit 0;
`golangci-lint run ./...` → **`0 issues.`** The two CI `Lint` findings are
linux-build-only, live in files this change does not touch, and reproduce on the
control branch (§6.3).

## 7. Findings — none blocking

### 7a. An rc.6 root that *drops* the new case still passes

`requireLifecycleCases` cannot demand `compiled-cache-miss-is-read-only` while
Curator's own `SPEC_PIN` is rc.3, and the focused test skips. Not a regression —
the old gate exits 0 on the same root — and the skip is classified (§4).
Strengthenable later: `conformance/v1/manifest.json` carries `protocol_version`,
so the case could become mandatory once a root declares rc.6+. Out of scope; the
AC asks for name-based assertion, which is delivered.

### 7b. `internal/install` multi-project binding gap — still unowned

Re-run by this review against the rc.6 root:

```
go test -run TestAuthoritativeDryRunCasesMutateNothingPersistent ./internal/install
bd6ba08   → FAIL  dryrun_conformance_test.go:189:
            published dry-run scope "multi-project" has no executable binding
cfffd7cd  → FAIL  identical message
```

**Byte-identical on old main and on the head** → pre-existing, untouched by this
change. Invisible today because `internal/install` is not in the Implementations
package set and Curator CI pins rc.3. `task-board grep multi-project` still finds
no board item for it. It becomes a live Curator CI failure the moment `SPEC_PIN`
advances to rc.6. Raising it is an orchestrator action; this review does not
expand scope.

### 7c. The pin names an unmerged PR head

`conformance/README.md` §4 wants a full immutable commit ID (satisfied —
`bd6ba08acda3dc801512c408c759ac0ac6f79f26`) that passed its own required CI (not
satisfiable at any commit today, because `main` itself is red for §6.3's reasons).
The task scope explicitly says *"publish through their existing task branches/PRs;
do not tag or release"*, so pinning a PR head is what the scope permits, nothing
shippable is gated on it, and the alternative — leaving Implementations red on
both PRs — is strictly worse.
**Binding follow-up:** after PR 9 merges, re-pin `implementations.yml` to the
resulting `main` commit on PR 14 and PR 15, before rc.6 is released.

### 7d. Housekeeping

`ci/goenv-control-BUG-260731-3gm8kc` (`8bedf884`) has now served three
independent comparisons and can be deleted.

### 7e. Process anomaly — duplicate reviewers cancelled mid-flight

Five reviewer runs were spawned for this item. Three produced nothing
(`RUN-260731-0910bf`, `RUN-260731-7b9b43` and four zero-byte logs), two produced
full verdicts (`4afbab`, `109b9b`), and the last two (`7d213b`, `685cd6`) were
cancelled as duplicates within 110 seconds of each other. Worth recording: the
cancellation of `7d213b` fired *after* it had already moved the board item out of
`done`, so a cancel arriving mid-review can leave a board item in a transient
status if the cancelled run does not get to finish. It did not here.

## 8. Conclusion

The acceptance criterion is met and verified against real CI on both target PRs,
down to per-runner test results. The replacement gate is proven **strictly
stronger** than the length check it replaces by four count-preserving controls,
three of them new to this run — including one (control I) that corrupts a case
outcome without touching any name and that the old gate accepted. The change
touches no product code, lands inside the repo's pre-existing `root-content` skip
class, and the folded-in GOENV repair is proven non-weakening on the real
toolchain. The producer's two unchecked items were honest and both resolve in the
change's favour on evidence re-derived here.

**This re-verification corroborates ACCEPTED.** The verdict of record stays
`RUN-260731-109b9b`; this run was cancelled as a duplicate and asserts no
verdict of its own.

Open items for the orchestrator: **7a** (optional `protocol_version`-gated
strengthening), **7b** (raise a board item for the `internal/install`
multi-project binding before `SPEC_PIN` advances — still missing), **7c** (re-pin
both spec PRs after Curator PR 9 merges), **7d** (delete the control branch),
**7e** (reviewer-spawn duplication).
