# Review findings — TASK-260906-2x4s7i (cycle 6)

**Verdict: ACCEPT.** `repeat-of: F17 / F11 / F20` (F23), `repeat-of: F21` (F24),
`repeat-of: none` (F22). Three minor findings, no blocking and no major.

**The landing question, answered against the fact rather than the brief.** PR #47 **merged
while this review was running** — at 16:50:24Z, into `origin/main`, as commit
`87a0d00`, which is **not** the commit any brief named. My subject was `407424e`; the
orchestrator authored `87a0d00` in response to cycle 5's F19/F21, pushed it, and landed it
nine minutes after cycle 5's ACCEPT. So no cycle's verdict ever covered the commit that
actually landed.

I reviewed `87a0d00` for that reason, and the answer is: **it was safe to land, and I can
say so from measurement rather than from inheritance.** The schema, `protocol/core.md`,
`protocol/environments.md` and `profiles/manager.md` blobs at `87a0d00` are byte-identical
to `407424e`; the corpus delta is purely additive with no existing `valid` flag moved; and
the classification table over 74 named spellings is *identical* at the two heads. The one
thing `87a0d00` changes is that it closes cycle 5's F19 — `M-drive-anycase-off` survived at
`407424e` and dies at the landed head on the case this commit adds.

Everything below was measured by me on two `--shared` clones checked out at `407424e` and
`87a0d00`, each verified byte-exact against its own committed blobs before anything was run
(1,187 and 1,188 paths, **0** content mismatches, **0** missing, **0** extra — hashed with
raw `sha1("blob N\0"+bytes)`, not through a filter). `jsonschema==4.25.1`, the
`requirements-dev.txt` pin. Nothing outside my scratch was written and no branch was modified.

---

## 0. Why this cycle is not cycle 5 repeated

The run that produced cycle 5's ACCEPT was flagged unsatisfied by the board — *"reviewer run
has no verdict branch while TASK-260906-2x4s7i is to-review"* — and I am its recovery
successor. Re-attesting a predecessor's conclusion would be worth nothing, so I re-derived
every dimension independently, with my own harness and my own mutants, and then extended the
subject to the commit that had landed in the meantime. Where I reproduce cycle 5's numbers I
say so; where I get different ones I say that too.

## 1. What I drove, and what it says

### The classification matrix — 74 spellings, both directions, against the committed schema

Driven through `Draft202012Validator` with the repository's own registry, against the **whole
document** (a published valid case used as the carrier, its overlay member replaced) rather
than against `$defs/overlay` in isolation — the production entry point, not the piece.

Kind is derived from behaviour, never read off the pattern. All **41** of my a-priori
expectations held: every POSIX, project-relative, Windows (four slash spellings, both letter
cases), UNC and scheme-relative spelling classifies `path` and is refused every one of
`range`, `tag`, `revision`, `directory` and `branch`; every allowed scheme in every letter
case and both SCP forms classify `git`, are refused bare, admit exactly one form, and are
refused two; `svn:`, `file:`, `ftp:`, an invalid 6.1 host with or without a user, a backslash
after the SCP colon and a bare drive letter are refused in every column. `weight` is legal on
both kinds, `directory` on `git` only. **0 partition violations.**

### The bypass hunt — the one failure that would matter

Core §6.1 mandates that invalid network forms be *"rejected, not treated as local"*. The
defect that would break this change is therefore a network-shaped spelling escaping to
`path`, because that skips the form requirement **and** contradicts 6.1. I hunted it over
2,779 generated spellings:

| assertion | result |
|---|---|
| `path`-classified spellings that are colon-shaped (`^[^/\s]*:`) but not a Windows drive | **0** of 120 |
| `git`-classified spellings with neither an allowed scheme nor a 6.1-grammar host | **0** of 676 |
| spellings valid both bare and form-carrying | **0** of 2,779 |

Every colon-shaped spelling is a drive letter, a 6.1-legal network form, or refused. There is
no third road to `path`. That is the strongest statement the schema layer admits, and it is a
measurement, not the argument I could have written instead.

### The mutants — 34 of 40 killed on a named case, 25 of them not deletions

Full table, flip sets and directions: `TASK-260906-2x4s7i_mutant-sweep-6.md`.

Harness validity first, because a sweep that cannot fail proves nothing: the unmutated corpus
exits 0; a no-op mutant survives; every mutant is asserted to actually change the serialised
schema; and `conformance/v1/manifest.json` digests **no** path under `schemas/` — so no kill
is a digest artefact.

Eleven of my mutants were run by no earlier cycle. Nine of those die, including
`M-scheme-case` (making the allowlist case-sensitive kills on
`valid-overlay-git-ssh-uppercase.json`, so the four uppercase positives are load-bearing) and
`M-drop-git`/`M-require-s` (allowlist membership pinned scheme by scheme).

Six survive: one is my intentional control, three are **measured 0-flip semantic no-ops**
(including cycle 3's `M-drive-restore3`, whose analytic proof I confirm by measurement), and
two are real unpinned decisions with degenerate classes — `M-drive-multiletter` (20 flips,
`HOST:/x`) and `M-arm2-colonplus` (2 flips, a source starting with a colon).

### The spec authority, read rather than cited

* §1's `path` kind is *"an operator-local package directory named by an absolute path, or by
  a project-relative path when the operation runs inside a project"* — unqualified as to
  spelling, which is what makes the Windows spellings legal `path` operands.
* Core §6.1 says network URLs *"contain no explicit port, password, query, fragment, percent
  escape, or backslash"*, so a backslash spelling cannot be a network form at all; its host
  grammar `[A-Za-z0-9][A-Za-z0-9.-]*` admits one character, which is what makes
  `c:example/x` a legal SCP host; *"Local paths and `file:` URLs have no network identity"*
  plus §1's *"directory"* operand is what makes a `file:` URL neither kind.
* §6's overlay paragraph and §1's `path` paragraph are **byte-identical to `origin/main`**;
  `protocol/core.md` is untouched across the whole branch; `protocol/environments.md` changed
  **exactly one line** in five commits — the §12.1 knob row.
* `profiles/manager.md:2189-2194` states the split as a manager-side obligation — *"the
  manager requires the `range | tag | revision` form (with optional `directory?`) on a `git`
  source and no requirement form on a `path` source"* — citing §12.1 and section 1. Manager's
  voice, member-for-member with the amended row. Cycle 1's F3 is properly closed.

### Gates and CI, re-run and re-read by me

| gate | at `407424e` | at `87a0d00` (landed) |
|---|---|---|
| `make validate` | **exit 0** — 60 schemas, 1046 vector files, 227 tests OK | **exit 0** — 60 schemas, **1047** vector files, 227 tests OK |
| `make regenerate-check` | **exit 0**, real `git diff --exit-code`, tree clean after regeneration | **exit 0**, same |

Both run as standalone processes on byte-verified clean clones. Because I cloned rather than
copied, `regenerate-check`'s `git diff --exit-code` is the repository's actual assertion, not
a re-hash standing in for it.

**CI ran against the landed commit, not an earlier one.** All 18 check runs on
`87a0d00` report `success` (one duplicate "Release target provenance" `skipped` by design on
the pull-request trigger), including all three `windows-latest` lanes — the platform whose
spelling this change exists to make declarable.

### Mechanics

Five commits past `550579d`, every one verifying **`G`** under the repository's own
`maintainers.allowed_signers`, author *and* committer `Ivan Oparin <oparin@me.com>`.
`mergeCommit` is `87a0d00` itself, so the PR landed by fast-forward of the exact head and the
signed objects were preserved rather than replaced. No tracked build artefact: 1,188 paths,
three binary blobs and one mode-`100755` file, all four byte-identical to `origin/main`; no
`__pycache__`, no root binary; `.gitignore:8` and `.gitignore:12` hold both traps.

**Pin consumption, read from the pin rather than from a claim.** The change touches
`conformance/v1/vectors/manager-config-v2.json`. The pinned Go manager
(`relux-works/curator@a3abcf34`) reads `vectors/manager-config.json` — a *different* file,
untouched across all five commits. Historical releases `rc.5`–`rc.8` are untouched; only the
rolling `rc.9` candidate digest moved. The three green Implementations lanes at the landed
head are the empirical confirmation.

---

## Findings

### F22 — MINOR. `repeat-of: none`. The commit that landed carried no reviewer verdict at merge time

Cycle 5 accepted `407424e`. `87a0d00` was authored, pushed and merged nine minutes later, and
five cycles of adversarial review all stopped one commit short of what is now `main`. The
merge itself is well-formed — fast-forward, exact head, signed, CI green on that SHA — so
this is not a defect in the change. It is a gap in the loop: the review verdict and the
landed artefact were allowed to name different commits.

I have closed it after the fact rather than merely noting it, which is why this finding is
minor instead of major: the delta is provably behaviour-free (identical schema and prose
blobs, purely additive corpus, no `valid` flag moved, identical classification at both heads)
and it strictly *improves* coverage. Had `87a0d00` touched the schema, nobody would have been
in a position to say so.

**What would fix the class:** treat "the verdict names an OID" as binding, and require a
re-verdict — even a one-line one — when the head moves after acceptance.

### F23 — MINOR. `repeat-of: F17 / F11 / F20`. The PR body says eight mutants and lists nine

> **The eight mutants listed here all die against a named case** — permissive host class,
> host ≥2 characters, dropping `directory`, re-admitting `file:`, scheme length, unanchored
> colon test, backslash after the SCP colon, the widened drive pattern, and an uppercase-only
> drive carve-out.

That is nine items. All **nine** do die, each on a named case — I verified every one
independently (`M-host`, `M-host2`, `M-dir`, `M-file`, `M-schemelen`, `M-unanchored1`,
`M-scp-bslash-arm2`, `M-drive-wide`, `M-drive-anycase-off`). So the claim is *true* and the
numeral is stale: the ninth was added for F19 without updating the count.

F20's substance is genuinely fixed — the sentence now says "the eight mutants listed here"
rather than "every mutant the four reviews named", so it no longer contradicts the bounds
paragraph four sentences later. This is the fourth stale count in this change's prose
(cycle 3's F11, cycle 4's F17, cycle 5's F20), which is why it is worth a line rather than a
shrug. The related nit: the body's "Four review cycles" heading now describes five.

Separately, the bounds sentence *"Three mutants reported by cycle 3 remain survivors"* is not
independently checkable from the body, because the mutants are named nowhere in it — the
definitions live only in the cycle-3 board resource. My own sweep finds two real unpinned
classes plus three 0-flip no-ops at this head, which is compatible with it but is not a
confirmation of it.

### F24 — MINOR. `repeat-of: F21`. "One-character prefix" overstates the rule by a character class

Both `CHANGELOG.md` and the PR body now state the F20 bound as:

> a one-character prefix before `://` is a drive path, so `C://Users/…` stays a `path`

Measured against the committed schema:

| spelling | classified |
|---|---|
| `g://host/x`, `C://x`, `z://x` — one **letter** | `path` |
| `1://x`, `9://h/x` — one **digit** | **refused** |
| `+://x`, `.://x`, `-://x` — one punctuation character | **refused** |

"One-**letter**" is the accurate word. The operative claim — that `C://Users/…` stays a
Windows path, which is the whole reason the bound exists — is exact, and the word "drive"
implies a letter to a careful reader, so this is a wording fix and not a behaviour question.
Unlike cycle 4's F16 and cycle 5's F21, no published case contradicts it.

---

## What I verified as correct, by measuring rather than reading

* **No bypass surface.** `$defs/overlay` is `$ref`-ed exactly once in the entire schema set.
  No sibling schema defines a competing overlay object: `context-lock-v1` and
  `agent-environment-marker-v1` use `overlay` only as a boolean, and `system-config-v2` only
  `$ref`s `overlays_allowed`. The lock's local-context member branch refuses `directory`
  alongside `source` and `commit`, so declarability composes without contradicting itself.
* **The corpus is a gate, not decoration.** 41 overlay cases at the landed head — **27
  negative** to 14 positive — over 24 distinct published `source` spellings covering every
  arm in both directions. Cycle 1's F2 (four spellings, nothing killable) is thoroughly
  closed.
* **No existing case was quietly re-verdicted.** Across all five commits: 881 → 911 index
  entries, 30 added, **0 removed**, and **0** pre-existing entry's `valid` flag changed. No
  index entry lacks a file; regeneration is byte-identical to the committed blobs.
* **Every number in the PR body's evidence paragraph reproduces**: 41 overlay cases, 60
  schemas, 1047 vector files, 227 tests, both gates exit 0, no tracked build artefact.

## AC coverage — 7 of 7

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | §12.1 knob row: form git-only, form-free `path` admitted | PASS | `protocol/environments.md:2223` and `profiles/manager.md:2189-2194`, both read by me; one-line diff in five commits |
| 2 | schema requires a form only for a `git` source | PASS | committed schema via `Draft202012Validator`, 74 named + 2,779 generated spellings, 0 partition violations |
| 3 | `path` + `range`/`tag`/`branch`/`revision`/`directory` stays invalid | PASS, 5 of 5 | `M-range`/`M-tag`/`M-revision`/`M-dir` each die on a *distinct* named path case; `branch` via `M-addprops` on `invalid-unknown-overlay-field.json` |
| 4 | published cases stop giving a path source a form | PASS | 41 overlay cases, 24 spellings; no `valid=true` case gives a non-git source a form |
| 5 | new cases proven killing by mutating the discriminator | PASS | 34 of 40 mutants killed on a named case, 25 of them boundary narrowings; `M-drive-anycase-off` flips `407424e` SURVIVES → `87a0d00` killed |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me at both heads, exit 0; pin reads a different vector file; 18 CI runs green on the landed SHA |
| 7 | signed commits, human identity, no stray file, change documented | PASS | five `G` signatures, fast-forward landing of the exact head, tracked set clean |

## Observation for a separate leaf, not a finding against this change

**25 schema-case files ship in the conformance manifest but are referenced by no `index.json`
entry**, so `tools/validate.py` never drives them against their schema — 12 under
`agent-environment-marker-v1/`, 13 under `launch-env-fragment-v1/`. All 25 pre-date this
branch (127 such files already at `550579d`, none touched here), so this is inherited, not
introduced. It is the same class as cycle 4's observation that no lane inspects the tracked
file set: the manifest and the index are two inventories that nothing reconciles. Worth one
gate asserting `manifest ∩ schema-cases == index`, on its own leaf.

Note for the record: cycle 5 reported "0 orphans" here. That measurement was index → disk
only; disk → index finds these 25. Nothing in this change depends on it.

## The landing question

**PR #47 landed at `87a0d00`, and it was safe to land.** I say that about the commit that is
in `main`, not about the one my brief named. The behaviour is settled — six cycles have now
attacked this discriminator, the last three found no behavioural defect at all, and my own
independent harness, eleven mutants nobody had run, and a targeted hunt for the one bypass
that would matter all failed to find one.

The three findings above are one process gap I have already closed by verifying the landed
commit, one stale numeral, and one word. None of them can change a classification verdict,
and none of them is a reason to reopen a merged, green, correctly signed change.

Reviewer artifacts, all under `.temp/review-c6/` in the story worktree and attached to this
element: `TASK-260906-2x4s7i_classification-matrix-6.md`,
`TASK-260906-2x4s7i_mutant-sweep-6.md`, `TASK-260906-2x4s7i_validate-6.log`, plus
`matrix.py`, `mutants.py`, `flips.py`, `run_matrix.py`. Nothing outside `.temp/` was written
and no branch was modified.
