# Review findings — TASK-260906-2x4s7i (cycle 8)

**Verdict: ACCEPT. PR #47 is safe to land, and it has already landed at `87a0d00`.**

`repeat-of: F25 and F26 (cycle 7) both still live at head, unchanged — nothing has landed since
they were written. F27's class recurs and I extend it. F28 is new. F29 is the
F17/F20/F23/F25/F26 prose-lagging-behaviour class, on the PR description this time. F30 is new.`

## 0. Standing, and the subject named by an OID

No cycle-8 brief is attached to this element; the briefs stop at cycle 5 (`407424e`) while the
element already carries verdicts through cycle 7. I reviewed **the landed tree at
`87a0d0060bad64ab883d007dcdf35df7485368bf`**, which **is** `origin/main`. PR #47 is `MERGED` with
`mergeCommit.oid == headRefOid == 87a0d00`, a fast-forward of the exact reviewed head, so the five
signed commit objects were preserved rather than replaced. Nothing exists past it.

This element was in `done` when my run started; the runner's mandated first command returned it to
`reviewing`. That is the loop condition cycle 7 recorded, not a new fact about the change.

I did not re-derive cycles 1–4. I re-drove the classification from scratch, attacked the structure
my predecessors did not, and checked their reported facts.

Everything below was measured by me in a clean `--shared` clone of `87a0d00` (1188 tracked paths,
`git status --porcelain` empty) against a venv built from the repo's own pinned
`requirements-dev.txt` (`jsonschema==4.25.1`).

## 1. Findings

### F28 — MINOR, new. The arm-1 and arm-2 refusals are defeated by one whitespace character, and their meaning depends on which regex engine reads them

Every refusal the last four commits added has a twin that differs by a single inserted character
and flips to a **valid, form-free `path`**. Driven through the committed schema:

| pinned refusal | kind | twin | kind |
|---|---|---|---|
| `my_host:x` | REFUSED | `my_hos<NBSP>t:x` | **path** |
| `git@my_host:x` | REFUSED | `git@my_hos<NBSP>t:x` | **path** |
| `github.com:\example\x` | REFUSED | `github.com<NBSP>:\example\x` | **path** |
| `svn://host/x` | REFUSED | `sv<NBSP>n://host/x` | **path** |
| `file:///x` | REFUSED | `fil<NBSP>e:///x` | **path** |
| `C:` | REFUSED | `C<NBSP>:` | **path** |
| `host:/x` | REFUSED | `host<NBSP>:/x` | **path** |

`<NBSP>` is U+00A0; U+000B, U+3000, U+2028 and a plain ASCII space or tab do the same. The cause is
`\s` inside arm 2's "is this colon-shaped" test `^[^/\s]*:` and inside the SCP pattern: a character
the engine counts as whitespace hides the colon from the test, arms 1 and 2 never fire, and arm 3
falls to its `else` and admits the entry with no requirement form.

**The second half is the one that matters for a specification.** JSON Schema specifies `pattern` as
ECMA-262. This repository validates with CPython `re`; the pinned reference manager is Go, whose
RE2 `\s` is `[\t\n\f\r ]` and nothing else. I compiled the five overlay patterns in all three
engines and derived the arm decisions from the raw booleans: **75 of 154 tested spellings classify
differently depending on the engine**, and the divergence reaches the escape-to-`path` direction.
`my_hos<NBSP>t:x` is `path` under CPython **and** under normative ECMA-262, and `REFUSED` under
RE2. The scheme patterns and the drive pattern are engine-identical; all of it comes from `\s`.

**No published case can see any of this.** Four mutants that materially change behaviour produce
**zero** flips on the 112-entry gate corpus — `M-colonish-drop-s` alone moves 160 spellings of a
412-spelling population from `path` to `REFUSED`. Measured cause: of the **162** overlay `source`
declarations across the 71 cases and 41 vectors, **0** contain any whitespace character and **0**
contain any non-ASCII character.

**Why this is MINOR and not blocking.** The consequence is a config that validates and then fails
at resolution with `profile_source_path_missing` instead of being refused earlier — no network
fetch (section 1: nothing about a `path` source is ever fetched from a network), no privilege
change, no data consequence. And the engine divergence is **latent, not live**: I verified from the
repository rather than from a claim that no pinned implementation consumes the schema —
`.github/ci/implementation-coverage.tsv` names no `schemas/` path (0 occurrences), and
`conformance/v1/manifest.json` digests no path under `schemas/` (0 of 2097 string values). The
implementation lanes publish `CURATOR_CONFORMANCE_ROOT=conformance/v1`, which does not contain the
schema set. It becomes a real conformance defect the day an implementation validates a
manager-config against this file.

Two engine-independent notes so the finding is not overstated: a spelling with a space in the first
segment (`my host:x`) is arguably a legal section 1 relative directory, so refusing it is not
*forced* by the spec; and the schema was never the identity validator. What is not defensible is
that `my_host:x` and `my_hos<NBSP>t:x` — identical in kind under every sentence in section 1 and
core section 6.1 — get opposite verdicts, and that which verdict they get is a property of the
regex engine rather than of the specification.

Evidence: `TASK-260906-2x4s7i_classification-matrix-8.md` sections 2–4,
`TASK-260906-2x4s7i_mutant-sweep-8.md` section 3.

### F29 — MINOR, new. Cycle 7's invariant #4 and the PR's stated bounds are broader than the corpora behind them

Cycle 7 published *"A colon-shaped non-drive spelling escaping to `path`: **0** violations"*,
measured over 3,916 generated spellings. F28 exhibits that exact violation. The generator that
produced those 3,916 spellings emitted no whitespace-class character, so the sweep measured a
smaller class than the sentence claims. Same for the PR's stated-bounds paragraph, which names the
host-vs-directory ambiguity, `c:example/x`, the `file:` refusal, the `://`-prefix drive rule, the
backslash bound and the diagnostic bound — but not the whitespace class.

This is the "nothing is there is not nothing this checker sees" shape. The measurement was honest
and the number was real; the sentence around it was not bounded by the corpus that produced it.
Neither cycle-7 artefact needs withdrawing — the invariant holds over the population it names.

### F30 — MINOR, new. The PR description says eight mutants and lists nine

*"The eight mutants listed here all die against a named case — permissive host class, host ≥2
characters, dropping `directory`, re-admitting `file:`, scheme length, unanchored colon test,
backslash after the SCP colon, the widened drive pattern, and an uppercase-only drive carve-out."*
That is nine items. I reproduced all nine independently and **all nine do die on a named case**
(`M-host-wide`, `M-host-two`, `M-else-drop-dir`, `M-scheme-add-file`, `M-scheme-star`,
`M-arm2-noanchor`, `M-scp-bslash`, `M-drive-wide`, `M-drive-upper`) — only the count word is wrong.
The same paragraph says *"Three mutants reported by cycle 3 remain survivors"*; my own 42-mutant
sweep finds **eight** real survivors, four of them the F28 family that no cycle has named. Sixth
instance in this change of prose lagging the behaviour it describes.

### F25 (cycle 7) — still live at head, verified by me

`CHANGELOG.md`: *"a one-character prefix before `://` is a drive path, so `C://Users/…` stays a
`path`"*. Measured at head: `a://x` -> **path**, but `1://x`, `+://x`, `-://x`, `.://x` -> all
**REFUSED**. Only a one-**letter** prefix is a drive path. Prose-only, one word.

### F26 (cycle 7) — still live at head, verified by me

The changelog's enumeration of refused shapes ends *"an unknown scheme, a `file:` URL, an SCP host
outside the grammar, and a backslash after the SCP colon"* and never names the **bare drive letter
`C:` refused** (`407424e`) or the **lowercase drive `c:\…` valid** (`87a0d00`) — the two cases those
commits exist to add.

### F27 (cycle 7) — reproduced independently, and extended

`M-drive-multi` (2 flips, `host:/x` REFUSED->`path`) and `M-host-nodigit` (6 flips, `9host:x`
`git`->REFUSED) both survive my sweep at the same measured sizes. I add a third in the same family:
**`M-scp-slash`** — admitting `/` after the SCP colon — survives the corpus at 12 flips
(`1:///x` and kin, REFUSED->`path`). `M-drive-multi` remains the one worth pinning: it turns
`host:/x`, an invalid section 6.1 form, into a form-free path.

**The behaviour at head is correct in every one of these cases.** They are coverage gaps.

## 2. Verified correct, by me rather than read

* **Classification.** 80 named spellings x 8 member combinations, whole document, kind derived from
  behaviour: 25 `path`, 29 `git`, 26 `REFUSED`, **0 MIXED**, **0 violations** of seven a-priori
  invariants (no spelling valid both bare and with a form; `branch` never admitted; two forms never
  admitted; a `path` never admits `directory`; `weight` never changes a verdict).
* **The corpus is a real gate.** 42 mutants, **32 killed on a named case**, 0 harness-rejected, only
  6 deletions among them; the unmutated corpus disagrees on 0 of 112 entries and both controls
  survive at a *measured* 0 flips, on the corpus **and** on a separate 412-spelling population.
  Every arm carries a killing case, and the arm-2 drive carve-out is live (`M-drive-del2`, 8 flips).
* **No second declaration surface.** `$defs/overlay` is the only overlay *declaration* shape in
  `schemas/v1`. `context-lock-v1` and `agent-environment-marker-v1` spell `overlay` as a `boolean`
  member flag; `system-config-v2` carries only `overlays_allowed`. The one other unconditional
  `range|tag|revision` `oneOf` is `agent-context-v1#/$defs/requirementForm`, governing `requires`
  edges — where section 1 says a `requires` edge never names a `path` source, so it is correct.
* **Authority, read by me.** `protocol/core.md` is byte-untouched across all five commits.
  `protocol/environments.md` changed exactly one line — the section 12.1 knob row. Sections 1 and 6
  are byte-identical to `550579d`. `schemas/v1/manager-config-v1.schema.json` untouched;
  `vectors/manager-config.json` untouched.
* **The residual edges hold under both readings.** `file:///x` refused: core 6.1 says local paths
  and `file:` URLs have no network identity, so not `git`; section 1's `path` is "an absolute path,
  or … a project-relative path", and a URL is not a path spelling, so not `path`; neither kind means
  refusal. `c:example/x` git: 6.1's host grammar `[A-Za-z0-9][A-Za-z0-9.-]*` admits one character,
  and a drive-*relative* spelling is neither absolute nor project-relative, so section 1 cannot
  claim it. `C:` refused: not a directory under section 1, and a host with an empty repository path
  under 6.1, which requires a non-empty one. `packages/team:context` and `a/b:c` path: a colon after
  a slash cannot be a 6.1 host.
* **The manager sentence is in the manager's voice** — *"the manager requires the
  `range | tag | revision` form … on a `git` source and no requirement form on a `path` source"* —
  citing "the environments §12.1 shape" and "the section 1 rules" rather than restating them
  normatively. Cycle 1's F3 is closed.
* **Cases are driven, and pin the schema alone.** All 71 `manager-config-v2` case files on disk are
  named by `index.json` — 0 orphans in this family — and `validate_wire_semantics` dispatches only
  on skill manifests, never on manager-config. (Repo-wide there are 25 orphaned case files, all in
  `agent-environment-marker-v1` and `launch-env-fragment-v1`, all pre-dating this branch and none
  touched by it — cycle 7's observation, reproduced.)
* **Gates, re-run by me as standalone processes:**
  * `make validate` -> **exit 0** — "validated 60 schemas and 1047 vector files"; "Ran 227 tests in
    161.123s / OK"; `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
  * `make regenerate-check` -> **exit 0**, and `git status --porcelain` is empty over the *whole*
    worktree afterwards — stricter than the Makefile's own path list, and proof the generator drops
    no artefact.
* **CI, bound to the SHA not to the PR.** `repos/relux-works/curator-spec/commits/87a0d00/check-runs`
  -> every lane `completed/success`, including all three `windows-latest` lanes; the only non-success
  is one `Release target provenance` run reporting `skipped`, whose duplicate reports `success`.
* **Mechanics.** Five commits `bd39adb 2f2dfa4 18dca85 407424e 87a0d00`, each `%G?` = **G** against
  the repository's own tracked `maintainers.allowed_signers`, author **and** committer
  `Ivan Oparin <oparin@me.com>`. 1188 tracked paths; the single mode-755 entry
  (`conformance/v1/fixtures/go-build-skill/scripts/skill-helper`) and both `.bin` fixtures pre-exist
  on `550579d`; the delta contains no binary change at all. `.gitignore` carries `__pycache__/`,
  `*.py[cod]` and `/generate-vectors`. Cycle 3's F10 class stays closed.
* **Pin consumption.** Measured, not claimed: the coverage ledger names no `schemas/` path and no
  manager-config artefact, and the conformance manifest digests no path under `schemas/`. Nothing
  the pinned Go manager or the pinned Python manager consumes gained a case it cannot pass, and all
  three `Implementations (*)` lanes are green at the SHA.

## 3. Observation, not a finding

`required: ["source"]` is unpinned: remove it and `{"revision": "a"*40}` becomes valid, because all
three arms condition on `properties.source.pattern`, which passes vacuously when `source` is absent.
I checked whether this change created it and it did not — `git show 550579d:schemas/v1/manager-config-v2.schema.json`
has the same vacuity, its unconditional `oneOf` being likewise satisfied by the bare form. Cycle 7's
observation, independently confirmed and now with the pre-branch comparison behind it.

## 4. AC coverage — 7 of 7, each row driven

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | 12.1 knob row: form git-only, admits form-free path | PASS | `environments.md` diff — one line across five commits |
| 2 | Schema requires a form only for a `git` source | PASS | 80-spelling x 8-combination matrix, whole document |
| 3 | `path` + range/tag/branch/revision/directory stays invalid | PASS | all five refused; `M-else-drop-*` all die on named cases |
| 4 | Published cases stop giving a path source a revision | PASS | 24 distinct spellings scanned; only the intentional negatives |
| 5 | Positive + negative path-overlay cases published | PASS | 41 overlay cases, all named by `index.json` |
| 6 | Both gates green; pin gains no unpassable case | PASS | re-run by me, exit 0 / exit 0; ledger and manifest read |
| 7 | Signed commits, human identity, no stray file | PASS | five `G` signatures, 1188 tracked paths, no binary delta |

Checklist rows from the task, all satisfied: the matrix is driven against the committed schema in
both directions; the residual edges are decided from section 1 and core 6.1 with the sentences that
decide them; the cases are proven killing by mutation with named case flips; the second `allOf` arm
is checked against core 6.1 and section 1; the manager sentence agrees with the amended knob row in
the manager's voice; both gates re-run and `gh pr checks 47` read; the landing question answered
plainly below.

## 5. The landing question, stated plainly

**PR #47 is safe to land.** It has landed, at `87a0d00`, by fast-forward of the reviewed head.

The behaviour is settled and faithful to section 1 and core 6.1 in both directions for every
realistic spelling. The corpus is a genuine gate — 32 of 42 mutants die on a named case, mostly on
narrowings rather than deletions, with the harness falsified by three controls before its results
were counted. Both gates are green under my own hand, every CI lane is green against the landed
SHA, and all five commits are signed by a human maintainer.

Nothing in F25–F30 is a reason to revert or to hold anything. Three are single-sentence prose
corrections, two are coverage gaps with the behaviour at head correct, and F28 is a latent
robustness and portability bound on degenerate spellings that no implementation currently reads.
They are follow-up material.

**If exactly one thing is done next**, it should be F28: replace `\s` with an engine-independent
class in the two colon patterns and publish one case with a whitespace character in the `source` —
that closes a 160-spelling behavioural class and removes the only place in this change where the
normative artefact means different things to different implementations. That is a new leaf, not a
reopening of this one.

Cycles 1–4 each found a real defect, one of them blocking. Cycles 5, 6, 7 and now 8 have each found
nothing blocking. Four consecutive clean cycles on a change that has already landed is the signal to
stop reviewing it.

Reviewer artefacts, all under gitignored `.temp/`: `harness.py`, `matrix.py`, `bypass.py`,
`gate.py`, `sweep.py`, `survivors.py`, `xengine/` (Python, Node and Go differential), plus the
attached matrix, sweep, validate log and logbook. Nothing was written into the repository, nothing
was pushed, and the branch was not modified.
