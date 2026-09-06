# Review findings — TASK-260906-3x0w4y (cycle 1, CR revision 1)

**Verdict: CHANGES REQUESTED.** `repeat-of: none` (first review of revision 1).

Subject: `task-board/story/STORY-260905-2z9pw4` at `535fda6`, one signed commit past
the story base `c25d78e` (`origin/main` = `550579d`; the branch already carried
`c25d78e` from the prior leaf). Reviewed delta: `git diff c25d78e..535fda6`
(39 files, +364/−70).

## 0. The `repository_delta: empty` question — answered first

The Change Request records base OID `535fda6`, candidate tree `9c6bc7e`, and an
empty patch. That is a **snapshot artifact, not an empty deliverable.**
`9c6bc7e` is exactly `git rev-parse 535fda6^{tree}`, and `535fda6` is the
producer's own commit: the checkpoint was recorded *at* the commit, so the
working-tree diff the CR mechanism snapshots is necessarily zero.

The cause is a contract collision, not producer conduct: `producer-brief-path-overlay.md`
requires "exactly one signed commit past current main", while the generic spawn
template requires the work be left uncommitted. The leaf-specific brief won, and
`review-brief-path-overlay-1.md` itself directs the reviewer at
`git diff origin/main..535fda6`. Real repository changes exist and were reviewed.

This is worth surfacing to the orchestrator: on this leaf the CR patch resource is
empty and carries no reviewable content, so the patch artifact cannot be used as
the review subject for commit-based briefs.

## Findings

### F1 — BLOCKING. On Windows the discriminator reproduces the exact undeclarability this task exists to remove, and admits the §1-invalid form instead

The schema decides git-vs-path by a regex over the `source` spelling:

```
^(?:(?:[Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://|(?:[^/:\s@]+@)?[^/:\s]+:[^/\s])
```

The SCP arm's trailing class `[^/\s]` admits a backslash. Measured against the
**committed** schema through `Draft202012Validator` with the repo's own
`tools/validate.py` registry (`.temp/review-3x0w4y/matrix.py`):

| overlay `source` | bare | + `revision` |
|---|---|---|
| `/Users/operator/context` | VALID | INVALID |
| `C:\Users\operator\context` | **INVALID** | **VALID** |
| `D:\ctx` | **INVALID** | — |
| `C:/Users/operator/context` | VALID | INVALID |

Both directions are wrong for the native Windows spelling:

* **Refuses the legal declaration.** A Windows operator's `C:\Users\operator\context`
  overlay is classified `git`, required to carry a form, and then refused by §1 for
  carrying one. That is precisely the undeclarability this leaf was created to remove,
  reproduced on a supported platform.
* **Admits the illegal one.** `{"source": "C:\\Users\\operator\\context", "revision": …}`
  now **passes** schema validation and dies later at `profile_source_invalid`. The
  schema actively teaches the wrong shape.
* The same directory is declarable or not depending on slash direction
  (`C:/…` works, `C:\…` does not). No sentence in §1 or core §6.1 makes that so.

**The finding holds under either reading of the spec**, so it does not depend on
settling §1's ambiguity:

* If a Windows absolute path is a legal `path` source, the schema wrongly refuses it.
* If core §6.1's "Invalid network forms MUST be rejected, not treated as local"
  governs instead, the schema wrongly **accepts** it whenever a form is supplied.
  "Rejected" is not "accepted if you add a revision".

**My reading, with the sentences.** A Windows absolute path *is* a legal `path`
source. §1 says a `path` source is "an operator-local package directory named by an
**absolute path**, or by a project-relative path when the operation runs inside a
project" — unqualified; nothing restricts the spelling. Windows is a first-class
platform here: `ci.yml` and `.github/workflows/implementations.yml` both run
`windows-latest`, core.md names Windows throughout, and environments.md:2002-2004
explicitly legislates "automation on Windows or in PowerShell". Decisively, core §6.1 —
the authority the discriminator itself claims to encode — says network URLs "contain no
explicit port, password, query, fragment, percent escape, **or backslash**". A
backslash spelling therefore cannot be a core §6.1 network form at all, so classifying
it as `git` contradicts the pattern's own cited authority.

**Both justifications offered for the bound fail.**

1. "The spec's path spellings are POSIX (`/Users/…`)" — the review brief rules this
   out in advance, correctly: examples are not a restriction.
2. "the `portablePath` grammar itself bans `:`" — **misapplied authority.**
   `portablePath` is the `$def` for `directory`. `source` is
   `common.schema.json#/$defs/nonEmptyString` — `minLength: 1`, `maxLength: 8192`,
   no pattern. Verified in `schemas/v1/common.schema.json`. `portablePath` says
   nothing about `source`.

**The bound is not forced — the repair is free.** I patched the pattern to use core
§6.1's own host grammar and exclude a backslash after the colon:

```
^(?:(?:[Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://|(?:[^/:\s@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\s\\])
```

and re-ran the **entire committed corpus** (43 schema cases + 15 vectors):
**all green**, zero regressions. Under it `C:\Users\operator\context` bare becomes
VALID, `C:\Users\operator\context` + `revision` becomes INVALID, and every SCP-form
classification is preserved (`git@github.com:example/x` bare stays INVALID = form
still demanded). This shape is illustrative, not prescriptive — the producer owns the
fix — but it proves the defect is a one-character-class miss, not a constraint.

Evidence: `.temp/review-3x0w4y/matrix.py`, `.temp/review-3x0w4y/mutants.py`.

### F2 — MAJOR. The discriminator — the single load-bearing new gate — has zero killing evidence

I reproduced the producer's declared **M3 survivor** independently: replacing the whole
discriminator with `^/` and swapping the `then`/`else` arms leaves all 43 schema cases
and 15 vectors **green**. So does my F1 repair pattern. The committed corpus cannot
distinguish the shipped discriminator from either a materially different one or a
broken one.

The cause is visible in the corpus itself. Every overlay `source` ever published in
this family is one of four spellings:

```
  36  "/Users/operator/context"
  31  "https://github.com/example/team-context"
  31  "https://github.com/example/personal-context"
  11  "https://github.com/example/x"
   2  ""
```

Only the two most trivially separable points are pinned. No committed case exercises a
**project-relative** path (a first-class §1 spelling — "a project-relative path when the
operation runs inside a project"), the **SCP form**, a `file:` URL, or any non-POSIX
spelling. The producer disclosed M3 honestly and named the relative-path gap as its
bound; the bound is real but it is not acceptable for the change's central mechanism.
Any future narrowing of the discriminator lands green.

Requested: publish cases pinning at least a project-relative `path` overlay (form-free,
valid) and an SCP-form `git` overlay (bare → invalid, +form → valid), plus — once F1 is
fixed — the Windows spelling in both directions. Then re-run M3; it must die.

### F3 — MAJOR (scope decision for the orchestrator). `profiles/manager.md` still mandates a form on every overlay

`profiles/manager.md:2189-2190`:

> Overlays are machine configuration only. `overlays.<profile>` is an ordered
> list of `{ source, range | tag | revision, directory?, weight? }`;

That is verbatim the sentence this task identified in §12.1 as one of the two things
making a `path` overlay undeclarable — still normative, in a document environments.md's
own preamble incorporates ("It extends, and does not reinterpret … the manager behavior
in `../profiles/manager.md`"). An implementer reading the manager profile still cannot
declare a `path` overlay.

**The producer disclosed this explicitly** (drafting report §10) and left it because
`producer-brief-path-overlay.md` scoped the files and did not list `profiles/manager.md`.
That is correct conduct, not a miss. But the AC reads "A path overlay is declarable",
and it is not, in the manager profile. Note the preceding commit on this same branch
(`c25d78e`) already touched `profiles/manager.md`, so the file is not off-limits to the
story. Orchestrator's call: fold into this rework, or split a follow-up leaf. I am
recording it, not choosing.

### F4 — MINOR. `svn://host/x` classifies as `path` and is admitted form-free

Measured: `{"source": "svn://host/x"}` → **VALID**. Core §6.1 restricts network URLs to
`ssh`, `git`, `http`, `https`, so an unknown scheme is not a legal git source; the schema
admits it as a form-free "path" instead of refusing it. Failure direction: **toward
`path`** — a non-git URL skips the form requirement and passes schema, dying later at
resolution. Low impact (§1 resolution still refuses it, and `source` was never identity-
validated), but it is the same class as F1 and worth a sentence in the bounds if not fixed.

`file:///x` → `path` is defensible and correctly reasoned: core §6.1 states plainly that
"Local paths and `file:` URLs have no network identity". Accepting the producer's bound
here.

## What is correct, verified by me rather than read

* **POSIX matrix is exactly faithful to §1 and §6.** Driven through the committed schema:
  `path` bare → VALID; `path` + each of `range`, `tag`, `revision`, `directory` → INVALID
  (the conditional); `path` + `branch` → INVALID (`additionalProperties: false`) — all five
  §1-forbidden members refused. `weight` legal on both kinds. `git` bare → INVALID,
  `git` + one form → VALID, `git` + two forms → INVALID. Empty source → INVALID both ways.
* **§1 and §6 prose untouched.** environments.md changed exactly one line (2222, the §12.1
  knob row). §6:836-838 and the §1 `path` paragraph are byte-identical.
* **Mutant discipline is genuine.** M1 (drop `revision` from the path `else` set) and M2
  (`oneOf`→`anyOf` on the git arm) are real **narrowing** mutants that admit exactly one
  member of the class, not deletions, and both are killed by named committed cases. That is
  the right shape. The gap is that neither attacks the discriminator (F2).
* **Pin-consumption constraint: clean — verified from the pin's source, not the report.**
  At `relux-works/curator@a3abcf34` the interop suite reads exactly one manager-config
  artefact: `internal/interop/golden_test.go:290` → `vectors/manager-config.json`, which this
  diff does **not** touch. `internal/skillspec.TestReleasedSchemaCases` enumerates schema-case
  **directories by name** (`agent-skill-v7/v8`, `csk-skill-v7/v8`) and never opens
  `schema-cases/index.json`; `suiteRoot` only `os.Stat`s `manifest.json` and does not verify its
  hash. `.github/ci/implementation-coverage.tsv` names no manager-config artefact. Nothing the
  pin consumes gained a case it cannot pass.
* **Gates re-run by me, standalone processes:**
  * `PATH=$PWD/.venv/bin:$PATH make validate` → **exit 0** — "validated 60 schemas and 1019
    vector files"; "Ran 227 tests in 60.679s / OK"; `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
  * `make regenerate-check` → **exit 0**, empty diff.
* **Mechanics clean.** Exactly one commit past the story base `c25d78e`; author
  `Ivan Oparin <oparin@me.com>`, matching `origin/main`'s own history; signed (SSH,
  `%G?` = `U` — valid signature, unknown validity only because no local `allowed_signers`
  trust file, identical to every commit on main). No `LOGBOOK.md`, no stray file, no
  `tools/__pycache__` staged, `vectors/manager-config.json` untouched.
* **Honest reporting.** The producer corrected the brief's false claim that
  `tools/__pycache__` "is now ignored" (I confirmed: `git check-ignore` exits 1;
  `.gitignore` holds only `.temp/` and `.DS_Store`, and `make validate` does generate the
  directory — the trap is still live for the next producer), and disclosed both survivors
  and the `manager.md` gap rather than burying them.
* **M4 bound accepted.** Reverting the §12.1 Values-cell prose survives the suite. For a
  normative table whose *behavior* is pinned by schema + cases + vectors, and where
  `validate.py` already machine-checks knob names/defaults/enums, prose-by-review is the
  repository's existing convention. Not a finding.

## AC coverage: 5 of 7 rows confirmed, 2 of 7 fail on a supported platform

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | 12.1 knob row: form git-only, admits form-free path | PASS | `protocol/environments.md:2222` |
| 2 | schema requires a form only for a `git` source | **FAIL** | committed schema via `Draft202012Validator` — correct for POSIX/scheme spellings, wrong for `C:\…` (F1) |
| 3 | path + range/tag/branch/revision/directory stays invalid | **FAIL** | same — all five refused on `/Users/…`; `C:\…` + `revision` measured **VALID** (F1) |
| 4 | published cases stop giving a path source a revision | PASS | corpus scan — only the two intentional negatives remain |
| 5 | positive + negative path-overlay cases published | PASS | `valid-overlay-path-source.json`, `invalid-overlay-path-requirement-form.json`, 2 vectors, 2 index entries |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me, exit 0; pin consumption read from `curator@a3abcf34` source |
| 7 | one signed commit, human identity, no stray file | PASS | `git log`, `git diff --stat`, `git status` |

## What unblocks acceptance

1. **F1** — make the discriminator not classify a Windows absolute path as `git`, or
   state the sentence in §1 or core §6.1 that restricts a `path` source to POSIX
   spellings. "The examples are POSIX" and `portablePath` are not that sentence.
2. **F2** — publish cases that kill M3: a project-relative `path` overlay, an SCP-form
   `git` overlay, and the Windows spelling once F1 lands. Re-run M3 and report it dead.
3. **F3** — orchestrator decides: fold `profiles/manager.md:2189-2190` into this rework
   or split a follow-up leaf.
4. **F4** — fix or record as a stated bound.
5. Re-run both gates and requote the exit codes.

Reviewer artifacts: `.temp/review-3x0w4y/{matrix.py,mutants.py,validate.log,regen.log}`
inside the story workspace. Nothing outside `.temp/` was written.
