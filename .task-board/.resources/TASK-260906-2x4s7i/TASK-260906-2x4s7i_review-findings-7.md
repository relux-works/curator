# Review findings — TASK-260906-2x4s7i (cycle 7)

**Verdict: ACCEPT. PR #47 is safe to land — and it already has landed.**

`repeat-of: F24 (as F25), F17/F20/F23 class (as F26). F27 is new.`

## 0. The subject, named by OID

`origin/main` **is** `87a0d0060bad64ab883d007dcdf35df7485368bf`. PR #47 merged at
2026-09-06T16:50:24Z with `mergeCommit == headRefOid == 87a0d00`, i.e. a fast-forward of the
exact reviewed head, so the five signed commit objects were preserved rather than replaced.
Nothing exists past it on the branch or on trunk.

My brief (cycle 5) names `407424e` and the delta `18dca85..407424e`. That commit is an
ancestor of `87a0d00`. I reviewed **the landed tree at `87a0d00`**, because reviewing a
superseded commit would produce a verdict about a tree nobody runs. Cycle 6 made the same
call for the same reason.

Everything below was driven by me at `87a0d00` in a `--shared` clone with a clean worktree
(1188 tracked paths, `git status --porcelain` empty), against a venv built from the repo's
own pinned `requirements-dev.txt` (`jsonschema==4.25.1`).

## 1. The discriminator, driven against the committed schema

The committed `$defs/overlay` is three `allOf` arms: refuse a ≥2-character URI scheme outside
core §6.1's four; refuse a colon-shaped spelling that is neither a Windows drive nor a valid
§6.1 SCP form; then require exactly one of `range`/`tag`/`revision` for a git spelling and
forbid all four of `range`/`tag`/`revision`/`directory` otherwise. `branch` is refused by
`additionalProperties: false`.

**Matrix: 74 named spellings × 8 member combinations, whole document through
`Draft202012Validator` with the repo's own registry.** Classification is derived from
*behaviour* (bare valid + forms invalid = `path`; bare invalid + each form valid = `git`;
nothing valid = `REFUSED`), never read off the pattern. Result: 22 `path`, 20 `git`,
32 `REFUSED`, **0 MIXED**, and **0 partition violations** against six a-priori invariants
(nothing valid both bare and form-carrying; `branch` never admitted; two forms never
admitted; a `path` never admits `directory`; no MIXED; `weight` never changes validity).

## 2. The bypass hunt — the one failure that would matter

The failure mode that matters is a valid core §6.1 network form escaping to `path`, because
that skips the form requirement and directly contradicts §6.1's *"Invalid network forms MUST
be rejected, not treated as local."*

I wrote an **independent core §6.1 checker from the prose alone** (four schemes; host
`[A-Za-z0-9][A-Za-z0-9.-]*`; no port, password, query, fragment, percent escape or backslash;
non-empty portable repository path) and cross-checked it against the schema's classification
over **3,916 generated spellings** built from 7 userinfo forms × 11 hosts × 6 path tails ×
13 schemes plus drive and POSIX families.

| Invariant | Violations |
|---|---:|
| A valid core §6.1 network form classified as anything but `git` | **0** |
| An unambiguous §1 absolute/project-relative path not classified `path` | **0** |
| Any MIXED classification | **0** |
| A colon-shaped non-drive spelling escaping to `path` | **0** |
| A `git`-classified spelling with neither an allowed scheme nor a §6.1 host | **0** |

## 3. Is the corpus a gate? Mutant sweep

Harness falsified before it was trusted: the unmutated schema produces **0 flips** over the
gate corpus (71 manager-config-v2 schema cases + 42 v2 vectors); every mutant is asserted to
change the document and to remain a legal Draft 2020-12 schema (0 harness-rejected); and an
intentional control that reorders `allOf` — order-free by construction — **survives**, as it
must.

**33 of 40 mutants die on a named case.** Only 9 are deletions; the rest are boundary
narrowings and widenings. Selected kills:

| Mutant | Named case that flips |
|---|---|
| `M-scheme-star` (`+`→`*`, 1-char scheme) | `valid-overlay-path-windows-double-slash.json` |
| `M-scheme-add-file` (re-admit `file:`) | `invalid-overlay-file-url-with-form.json` |
| `M-scheme-drop-git` / `-ssh` / `-require-s` | the matching uppercase-scheme positive |
| `M-scheme-case` (drop case folding) | `valid-overlay-git-git-uppercase.json` |
| `M-host-wide` / `M-host-underscore` | `invalid-overlay-scp-host-grammar-with-form.json` |
| `M-host-two` (host ≥2 chars) | `valid-overlay-git-single-letter-host.json` |
| `M-scp-bslash` (admit `\` after the colon) | `invalid-overlay-path-windows-requirement-form.json` |
| `M-drive-del2` (remove the arm-2 carve-out) | `valid-overlay-path-windows-backslash.json` |
| `M-drive-wide` (drop the required separator) | `invalid-overlay-bare-drive-letter.json` |
| `M-drive-upper` (uppercase-only drive) | `valid-overlay-path-windows-lowercase-drive.json` |
| `M-else-drop-dir` | `invalid-overlay-path-directory.json` |
| `M-arm3-anyof` (`oneOf`→`anyOf`) | `invalid-overlay-two-requirement-forms.json` |
| `M-addprops-open` | `invalid-unknown-overlay-field.json` |

The arm-2 drive carve-out is **live, not dead**: `M-drive-del2` kills on a named case with
8 flips. Cycle 3's F13 removed the *arm-3* carve-out as provably dead; the arm-2 one is
load-bearing and is pinned.

## 4. Faithfulness to the authority, read by me

* **core §6.1 is untouched across all five commits** (`git diff --stat -- protocol/core.md`
  is empty). The discriminator cites an authority it does not edit.
* **`protocol/environments.md` changed exactly one line in five commits** — the §12.1 knob
  row, now `… for a `git` source, `{ source, weight? }` for a `path` source`. §1 and §6 are
  byte-identical to pre-branch main.
* **§1 decides the residual edges**, and I read the sentences rather than accepting them:
  * `file:///x` **refused** is right. §1's kind set is closed to `git`/`local`/`path`; core
    §6.1 says plainly *"Local paths and `file:` URLs have no network identity"* so it is not
    `git`, and §1's `path` is *"an absolute path, or … a project-relative path"* — a URL is
    not a path spelling. Neither kind ⇒ refusal. Two published cases pin it.
  * `c:example/x` **git** is right. §6.1's host grammar admits one character, and a
    drive-*relative* spelling is neither absolute nor project-relative, so §1 cannot claim it
    as `path`. No §1/§6.1 conflict; no spec amendment needed.
  * `C:` **refused** holds under both readings: not an absolute or project-relative directory
    under §1, and a host with an empty repository path under §6.1, which requires a non-empty
    one.
  * `packages/team:context`, `a/b:c` **path**: a colon after a slash cannot be a §6.1 host, so
    "reject, don't treat as local" never reaches them.
  * `\\server\share\x` and `//host/x` **path**: neither is a §6.1 network form (§6.1 forbids
    backslash outright), and both are absolute spellings on their platform.
* **The manager sentence is in the manager's voice.** `profiles/manager.md` states the split
  as a manager-side obligation — *"the manager requires the `range | tag | revision` form …
  on a `git` source and no requirement form on a `path` source"* — citing "the environments
  §12.1 shape" and "the section 1 rules" rather than restating them normatively. Cycle 1's F3
  is closed.

## 5. Gates and mechanics, re-run by me

| Gate | Result |
|---|---|
| `make validate` | **exit 0** — validated 60 schemas and 1047 vector files; Ran 227 tests, OK; `ok …/tools/generate-vectors` |
| `make regenerate-check` | **exit 0**, and the *whole* worktree is clean afterwards — a stricter check than the Makefile's own path list, and proof the generator drops no artefact |
| `gh api …/commits/87a0d00/check-runs` | **all lanes success**, including all three `windows-latest` lanes. Bound to the SHA, not read off `gh pr checks`, which does not tie a lane to a commit |
| Signatures | all five commits `%G?` = **G** against the repo's own `maintainers.allowed_signers`, human author **and** committer `Ivan Oparin <oparin@me.com>` |
| Tracked file set | 1188 paths; the only non-plain-text or mode-755 entries are 4 conformance fixtures, each verified to **pre-exist on `550579d`**. No build artefact. Cycle 3's F10 class stays closed |
| Pin consumption | CI pins and the coverage ledger are **untouched** by the branch, and no lane names a manager-config artefact. All three `Implementations (*)` lanes are green at `87a0d00` |
| Bypass surface | `$defs/overlay` is `$ref`-ed **exactly once**; no sibling schema defines a competing overlay declaration shape (`context-lock-v1`'s `overlay` is a boolean flag) |
| Cases actually driven | all 41 manager-config-v2 overlay cases are named by `index.json`, and `validate_wire_semantics` never dispatches on manager-config — so the cases pin the schema **and only the schema** |

## 6. Findings — three, all MINOR, none blocking

### F25 — MINOR. `repeat-of: F24 (cycle 6).` The changelog clause is over-broad, in the commit that set out to fix it

`CHANGELOG.md` at head: *"a one-character prefix before `://` is a drive path, so `C://Users/…`
stays a `path`"*. Measured: only a one-**letter** prefix is. `1://x`, `+://x`, `-://x` and
`.://x` are all **REFUSED**, not paths, because the drive carve-out is `^[A-Za-z]:[\\/]`.

`87a0d00` — subject *"say what the scheme rule actually encodes"* — is the commit that
introduced this clause. Cycle 6 named the same defect and said "one-letter" is the accurate
word. No published case contradicts it: I checked all 45 pinned overlay entries and none uses
a digit or punctuation prefix before `://`. Prose-only, one word.

### F26 — MINOR. `repeat-of: the F17/F20/F23 stale-enumeration class.` The changelog's list of pinned arms omits the two arms the last two commits added

The entry enumerates what the new cases pin and ends *"… an SCP host outside the grammar, and
a backslash after the SCP colon."* It never mentions the **bare drive letter `C:` refused**
(added at `407424e`) or the **lowercase drive letter `c:\…` valid** (added at `87a0d00`) —
the two cases those commits exist to add. Fifth instance of prose lagging the corpus in this
change.

### F27 — MINOR, new. Four classification classes are unpinned, each with a measured named flip

Of the 7 mutants that survive, one is the intentional control (0 flips, measured). The other
six are real, and I measured every one over 2,264 spellings rather than calling them no-ops:

| Survivor | Flips | Direction | Class |
|---|---:|---|---|
| `M-drive-multi` | 2 | REFUSED→**path** | `host:/x` |
| `M-scheme-noanchor` | 420 | path→REFUSED | `/x/svn://host/y` |
| `M-scp-useratsign` | 24 | REFUSED→git | `@host:x` |
| `M-host-nodigit` | 12 | git→REFUSED | `9host:x` |
| `M-arm2-colonplus` | 66 | REFUSED→**path** | `:x`, `://…` |
| `M-control-reorder` | 0 | — | intentional control |

`M-drive-multi` is the one worth closing. Widening the drive carve-out to multiple letters
turns `host:/x` — an *invalid* §6.1 network form, since its repository path is not portable-
relative — into a **form-free path**. That is precisely the "rejected, not treated as local"
violation cycle 2's F7b existed to remove, and nothing would catch the regression. One
invalid case pinning `host:/x` closes it. `M-arm2-colonplus` is the same direction on
degenerate colon-leading spellings and is now the third cycle to survive.

`M-host-nodigit` is the opposite direction and the F1 class in miniature: §6.1's host grammar
explicitly admits a leading digit, so `9host:x` is a legal git source, and nothing pins it.

**The behaviour at head is correct in all six cases.** These are coverage gaps, not defects.

## 7. Observations, not findings

* The overlay's `required: ["source"]` is **unpinned**: remove it and
  `{"revision": "…"}` with no `source` becomes valid, because all three arms condition on
  `properties.source.pattern`, which passes vacuously when `source` is absent. It
  **pre-exists on `550579d`** (verified against the pre-branch schema) and is not a
  regression — but this change turned that vacuity into a live path, where the old schema
  had an unconditional `oneOf`. No published case has a source-less overlay entry.
* **25 schema cases on disk are named by no `index.json` entry** and are therefore never
  driven by `validate.py`, while still sitting inside the digest-covered conformance
  manifest (12 `agent-environment-marker-v1`, 13 `launch-env-fragment-v1`). All 25 pre-date
  this branch and none is touched by it. Same class as cycle 4's "no lane inspects the
  tracked file set" — a checker whose corpus and whose manifest disagree about what exists.

## 8. AC coverage — 7 of 7

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | §12.1 knob row: form git-only, admits form-free path | PASS | `environments.md` diff, one line across five commits |
| 2 | Schema requires a form only for a `git` source | PASS | 74-spelling matrix + 3,916-spelling invariant sweep |
| 3 | `path` + range/tag/branch/revision/directory stays invalid | PASS | all five refused; `M-else-drop-dir` dies on a named case |
| 4 | Published cases stop giving a path source a revision | PASS | only the two intentional negatives remain |
| 5 | Positive + negative path-overlay cases published | PASS | 45 pinned overlay entries, all named by `index.json` |
| 6 | Both gates green; pin gains no unpassable case | PASS | re-run by me, exit 0/0; pins untouched, lanes green |
| 7 | Signed commits, human identity, no stray file | PASS | five `G` signatures, 1188 tracked paths, no artefact |

## 9. The landing question, stated plainly

**PR #47 is safe to land.** Its behaviour is settled, faithful to §1 and core §6.1 in both
directions, and pinned by a corpus that is a real gate — 33 of 40 mutants die on a named
case, and the classification survives an independent §6.1 cross-check over 3,916 spellings
with zero violations. Both gates are green under my own hand, all CI lanes are green against
the landed SHA, and every commit is signed by a human maintainer.

It has already landed, at `87a0d00`, by fast-forward. Nothing in F25–F27 is a reason to
revert or to hold anything: two are single-sentence prose corrections and one is a coverage
suggestion. They are follow-up material.

Four cycles of "not yet" on this change were justified — cycles 1–4 each found a real defect,
including one blocking. Cycles 5, 6 and 7 have now each found nothing blocking. That is the
signal to stop reviewing this change.
