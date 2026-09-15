# TASK-260906-19gjyw — rework report 1

Repository `relux-works/curator`, worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume`, branch
`feat/consume-overlay-rule`. Base for this rework: `38702164` (the reviewed head).
New head: `e4ddca19` — two signed commits, author `Ivan Oparin <oparin@me.com>`,
both `%G? = G`. Not pushed; PR #62 untouched.
Authority: curator-spec `87a0d0060bad`.

Verdict addressed: review cycle 1, CHANGES REQUESTED — F1 major, F2 minor, F3 trivial.
**No product behaviour changed.** The two commits touch one test file, one ledger TSV, and
comment text plus a function reordering in `internal/envprofile/envprofile.go`.
`installOperandKind`'s body is byte-identical to the reviewed head.

---

## 1. F1 (major) — the property loop excluded its own subject

### 1.1 Reproduced first, by execution

I applied M10 to the reviewed head — deleted the widening, leaving
`return identity.ClassifySource(trimmed)` — and ran each named test as its own process,
counting `=== RUN` lines so a filter matching nothing could not read as a survivor:

| test | `=== RUN` | exit | verdict |
|---|---:|---:|---|
| `TestInstallNeverDemotesANetworkIdentityToAPath` | 1 | **0** | **SURVIVED**, logging `checked 5 network-identity operands` |
| `TestOperandShadowedByDirectoryResolvesAsGit` | 1 | 1 | killed (`shadowed operand err = <nil>, want profile_source_invalid`) |
| `TestInstallOperandKindIsSyntactic` | 1 | 1 | killed |

The reviewer's diagnosis is confirmed by direct measurement of the predicate. Driving the
whole matrix through `identity.Parse` and `identity.ValidCanonical`:

| operand | `ClassifySource` | `Parse` | `Parse` err | `ValidCanonical` | in the old loop? |
|---|---|---|---|---|---|
| `github.com/evil-org/pkg` | path | `""` | nil | **true** | **no — skipped** |
| `github.com/relux-works/pkg` | path | `""` | nil | **true** | **no — skipped** |
| `example.com/org/pkg` | path | `""` | nil | **true** | **no — skipped** |
| `https://example.com/org/pkg` | git | `example.com/org/pkg` | nil | false | yes |
| `ssh://git@example.com/org/pkg` | git | `example.com/org/pkg` | nil | false | yes |
| `git@example.com:org/pkg.git` | git | `example.com/org/pkg` | nil | false | yes |
| `c:example/pkg` | git | `c/example/pkg` | nil | false | yes |
| `github.com:/org/pkg` | invalid | `github.com/org/pkg` | nil | false | yes |

The three skipped rows are exactly and only the class the widening exists for; the five
checked rows are decided by `ClassifySource` alone and the widening cannot move any of them
in either direction. `checked 5` was a count of a class of size 0.

### 1.2 The repair

`internal/envprofile/envprofile_f13f14_test.go`. The class "carries a core §6.1 canonical
network identity" has **two** spellings, and membership must admit both: a spelling
`identity.Parse` canonicalizes, **or** one that is already `identity.ValidCanonical`.

```go
canonical, err := identity.Parse(trimmed)
switch {
case err == nil && canonical != "":
case identity.ValidCanonical(trimmed):
    canonical = trimmed
    bare++
default:
    continue
}
```

A second guard, `if bare == 0 { t.Fatal(...) }`, makes the coverage-ratio mechanism protect
the class it reports on: if the matrix ever loses its bare canonical identities, the test
fails rather than shrinking silently. The log line now reads
`checked 8 network-identity operands, 3 of them bare canonical`.

### 1.3 The kill, after the repair

M10 re-applied to `installOperandKind` at the corrected head:

| test | `=== RUN` | exit | verdict |
|---|---:|---:|---|
| `TestInstallNeverDemotesANetworkIdentityToAPath` | 1 | **1** | **killed** — `operand "github.com/evil-org/pkg" carries network identity "github.com/evil-org/pkg" but classified path` (and its two siblings) |
| `TestOperandShadowedByDirectoryResolvesAsGit` | 1 | 1 | killed |
| `TestInstallOperandKindIsSyntactic` | 1 | 1 | killed |

**3 of 3.** M10 was reverted; the tree was restored from a pre-mutation copy and
`git status` verified clean before committing.

### 1.4 The three artifacts that repeated the false claim

1. **The test's own comment** — "A mutant that drops the widening … fails here." Now true by
   measurement (§1.3), and the comment gains the reason the naive filter was wrong.
2. **The drafting report** — §4 row M10 said "`TestOperandShadowedByDirectoryResolvesAsGit`
   (+2) … killed". Corrected to name all three tests, record that the third survived as
   first written, and carry the repair. §3's closing paragraph and the total both corrected.
3. **The ledger row** (`.github/ci/platform-cases.tsv:312`) — reworded to state the property
   the test now actually asserts, naming both spellings of the class.

---

## 2. F2 (minor) — the missing refuse→path class

### 2.1 I re-derived the table rather than accept either the report's or the review's

Method: `git worktree add --detach` a throwaway checkout of `7c74a492` (removed and pruned
afterwards), drop a probe test into it that drives stage (c)'s classification
(`isPathOperand` → else `canonicalGit`), drive the same operand file through
`installOperandKind` at this head, and diff. **44 operands: 31 unchanged, 13 changed.**

| operand | before | after |
|---|---|---|
| `~/pkg` | git | path |
| `~pkg` | git | path |
| `pkg` | git | path |
| `MyOrg/pkg` | git | path |
| `GitHub.com/example/x` | git | path |
| `github.com/../../etc/passwd` | git | path |
| **`packages/team:context`** | **refused** | **path** |
| **`a/b:c`** | **refused** | **path** |
| `c:example/pkg` | path | git |
| `C:example/x` | path | git |
| `github.com:/org/pkg` | git | refused |
| `D:` | path | refused |
| `C:` | path | refused |

Every row of the report's table reproduces, and the two bold rows are the class it omitted.
That confirms the review's F2 independently.

### 2.2 The direction that matters is still clean, and I measured that too

The AC's word is "silently". The six git→path rows are the ones worth attacking, so I checked
what the pre-change `git` install actually did with each: `ensureRepo` clones
`https://<canonical>` **only** when `identity.ValidCanonical(canonical)`, otherwise it hands
the spelling to `git clone` as a local path.

| operand | `ValidCanonical` | pre-change clone target | now |
|---|---|---|---|
| `~/pkg`, `~pkg`, `pkg`, `MyOrg/pkg`, `GitHub.com/example/x`, `github.com/../../etc/passwd` | false | local path handed to git | path |
| `packages/team`, `packages/team-context`, `github.com/evil-org/pkg` | **true** | **`https://` network** | **git — unchanged** |

**No operand that reached the network before this change reaches a local path now.** Every
git→path row was already a local clone. That is the F14 invariant, and it is measured, not
argued.

### 2.3 The pin, and a narrowing mutant that proves it was missing

`packages/team:context → identity.SourcePath` added to `installOperandCases`.

**M19** narrows the install classifier to refuse a colon reached after a `/` — exactly
stage (c)'s verdict for that class and nothing else:

```go
if kind == identity.SourcePath {
    if i := strings.Index(trimmed, ":"); i > 0 && strings.Contains(trimmed[:i], "/") {
        return identity.SourceInvalid
    }
}
```

| matrix | `TestInstallOperandKindIsSyntactic` | exit | verdict |
|---|---|---:|---|
| as reviewed (no colon-in-later-segment operand) | ran, PASS | **0** | **SURVIVED** — the F2 gap, demonstrated |
| with the added operand | `installOperandKind("packages/team:context") = invalid, want path` | 1 | **killed** |

The gap and its closure are the same experiment run twice.

---

## 3. F3 (trivial) — the doc comment

`internal/envprofile/envprofile.go`. `sourceKindRefusal` moved above `installOperandKind`
so each function carries its own block. `installOperandKind` — including the widening's
justification, the one comment here a reader most needs — now documents
`installOperandKind`.

While there, one sentence in that block was too strong to leave: *"Nothing changes kind:
`packages/team` was a git install operand before this change and still is."* It now states
the measured direction — no operand that reached the network before reaches a local path
now — and names the one class that does move, refused → path.

---

## 4. Mutant table for this rework

Every mutant applied to a fresh copy, run as its own process, `=== RUN` counted so a
no-op `-run` filter reports NOT-RUN rather than "survived", then reverted.

| # | narrowed to admit / reject | named test | `=== RUN` | exit | verdict |
|---|---|---|---:|---:|---|
| M10 | install: drops the network-identity widening, so a bare canonical identity is a `path` | `TestInstallNeverDemotesANetworkIdentityToAPath` | 1 | 1 | **killed** (exit 0 / SURVIVED before the repair) |
| M10 | same | `TestOperandShadowedByDirectoryResolvesAsGit` | 1 | 1 | killed |
| M10 | same | `TestInstallOperandKindIsSyntactic` | 1 | 1 | killed |
| M18 | the property loop's membership filter narrowed back to `Parse`-only — the exact defect F1 found, reintroduced | `TestInstallNeverDemotesANetworkIdentityToAPath` | 1 | 1 | **killed** by the new `bare == 0` guard: *"the matrix carries no bare canonical identity: the widening's own class is unchecked"* |
| M19 | install refuses a colon in a later segment (stage (c)'s verdict for exactly that class) | `TestInstallOperandKindIsSyntactic` | 1 | 1 | **killed** — and **survived at exit 0** on the reviewed matrix |

M18 is the regression pin for F1: the defect cannot return silently, because reintroducing
it now fails the test by name. M19 is the regression pin for F2, and its survival on the
old matrix is the finding itself.

Standing total for the task after this rework: **19 mutants, 19 killed, 0 survivors**
(M1–M17 from the drafting report, re-run and confirmed by the reviewer; M18 and M19 added
here).

---

## 5. Gate table — head `e4ddca19`, darwin, each command a standalone process, run sequentially

| gate | command | exit |
|---|---|---:|
| build | `go build ./...` | 0 |
| vet | `go vet ./...` | 0 |
| gofmt | `gofmt -l cmd internal` (0 lines) | 0 |
| lint | `golangci-lint run ./...` — 0 issues | 0 |
| gate selftest | `.github/ci/gate-selftest.sh` — 94 passed, 0 failed | 0 |
| suppression | `.github/ci/no-broad-suppression.sh` — ok | 0 |
| ledger | `.github/ci/ledger-consistency.sh` — 225 rows across linux darwin windows, ok | 0 |
| ledger delta, by name | `platform-case-gate.sh` over a ledger filtered to the two edited rows + a real `go test -json` stream — both observed **PASSING by name**, 0 skips | 0 |
| race, touched packages | `go test -count=1 -race -timeout 20m ./internal/envprofile ./internal/identity` — 84.2s / 1.5s | 0 |
| CLI suite | `go test -count=1 -timeout 30m ./cmd/curator` — 295.4s | 0 |

No `-race` suite ran alongside anything else, and no two lanes ran concurrently.

**The two `test-gate.sh` lanes were not re-run, and here is why rather than a promise.**
The delta is one test file, comment text, a function reordering, and two ledger
descriptions. `grep -rl CURATOR_CONFORMANCE_ROOT --include='*.go' internal/envprofile`
returns **0 files**: this package's behaviour does not depend on which conformance root is
served, so neither root can change its verdict. `platform-cases.tsv` is consumed by
`ledger-consistency.sh` and `platform-case-gate.sh`, both re-run above. **The hosted
candidate lane does not need re-dispatching** for this change, and it was not
re-dispatched: the standing measurement is run
[34071813375](https://github.com/relux-works/curator/actions/runs/34071813375) at head
`38702164`, green on ubuntu / macos / windows with `CI_REQUIRE_FULL_ROOT=1`.

---

## 6. Acceptance-criteria coverage after this rework

**7 of 7 AC rows driven.** Only rows 4 and 6 changed status in this rework.

| # | AC row | production call site | status |
|---|---|---|---|
| 1 | path overlay reachable from all three surfaces, joins the closure with its weight, through `run()` | `config.Load` → `PolicyFromConfig` → `Install` → `resolveOverlay`; `cli.cmdComposeAdd` | driven — unchanged, reviewer-verified |
| 2 | one exported discriminator matches the landed schema, never probes the filesystem | `identity.ClassifySource` | driven — unchanged; reviewer's 121-spelling three-engine differential, 0 disagreements |
| 3 | a path overlay carrying range/tag/branch/revision/directory stays `profile_source_invalid` | `parseOverlay`, `resolveOverlay` | driven for four of five; `branch` is the accepted stated bound (no such property in `$defs/overlay`) |
| 4 | `profile install` uses the same helper, **no source silently changing kind** | `installOperandKind` → `installLocked` | **driven — F2 closed.** Table re-derived by measurement (44 operands, 31 unchanged, 13 changed); the omitted class added and pinned; M19 |
| 5 | candidate lane green on three runners, closing the 30 subcases | `.github/ci/test-gate.sh` | driven — run 34071813375, unaffected by this delta |
| 6 | the stale bound and ledger rows 303–305 **truthful** | `.github/ci/platform-cases.tsv` | **driven — F1 closed.** The one untruthful row now states what its test asserts, and both edited rows drive green through `platform-case-gate.sh` by name |
| 7 | every new refusal driven through `run()`, proven by a narrowing mutant | as above | driven — 19 of 19 killed |

---

## 7. Bounds, stated honestly

* **The two `test-gate.sh` lanes were not re-run in this rework.** Reasoned in §5 with a
  measured premise (`internal/envprofile` reads no conformance root), not asserted.
* **`platform-case-gate.sh` was driven over a two-row filtered ledger**, not the full
  225-row one, because a full-ledger run needs a complete suite `-json` stream. It proves
  the two edited rows are observed passing **by name**; the full-ledger property is carried
  by `ledger-consistency.sh` (exit 0) and by the hosted lanes at `38702164`, whose
  behaviour this delta does not change.
* **All measurement here is darwin.** The delta adds no platform-conditional code. Linux
  and windows execution of these two rows is carried by the hosted run at `38702164` plus
  `ledger-consistency.sh`'s cross-`GOOS` compile check; this rework does not re-measure it
  and does not claim to.
* **The two token-preserving text-gate mutants (T1, T2) were not re-applied in this rework.**
  `ledger-consistency.sh` is the source-text-inspecting gate this change touches and it is
  **unmodified**; my ledger edit changes only the free-text description column, not a case-name
  token and not `must_run_on`, which are what T1 and T2 attack. Both were reproduced by the
  reviewer at `38702164` and stand. The stronger check for *this* delta was run instead:
  `platform-case-gate.sh` over the two edited rows against a real `go test -json` stream,
  both observed passing **by name**.
* **The epic-wide false-survivor question is not re-audited here** — per the orchestrator's
  note, the `-run` split error is one-directional (a pattern matching nothing exits 0, which
  reads as *survived*; it cannot manufacture a kill), so every prior "killed" verdict stands
  and the only possible corruption is over-reporting survivors. Left as recorded.
