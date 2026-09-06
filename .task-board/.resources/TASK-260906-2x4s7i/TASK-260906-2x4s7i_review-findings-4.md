# Review findings — TASK-260906-2x4s7i (cycle 4, PR #47 at `18dca85`)

**Verdict: CHANGES REQUESTED.** `repeat-of: F6 / F11` (F16), `repeat-of: F11` (F17), `repeat-of: none` (F18).
One major, two minor.

**PR #47 is not safe to land as it stands, and one thing must change: `CHANGELOG.md` tells implementers
that `file:` is an accepted scheme, and the schema this PR ships refuses it.** That is the whole
blocking set. The discriminator itself is sound — I could not construct a §1-legal or §6.1-legal
spelling it misclassifies, the partition is clean over 16,159 spellings, 32 of 41 mutants die on a
named case, both gates are green, all eight PR checks pass, and every cycle-3 finding is genuinely
repaired. This is a one-line fix to a prose file, not a rework of the change.

Subject: `feat/path-overlay-declarable` at `18dca85`, three signed commits past `origin/main` =
`550579d`. Cycle-3 delta `git diff 2f2dfa4..18dca85` (11 files). No Change Request captured for this
element; the artefact under review is the PR.

Everything below was measured by me on an `rsync` copy of the clean `18dca85` worktree
(`jsonschema==4.25.1`, the `requirements-dev.txt` pin). Nothing outside my scratch was written; the
review worktree was read only.

---

## 0. The cycle-3 findings, each re-driven

| cycle-3 finding | claim | my measurement |
|---|---|---|
| **F10** 5.4 MB `generate-vectors` tracked at the root | untracked and ignored | `git ls-tree -r 18dca85` — 1186 tracked paths, no `generate-vectors`. Largest tracked blob is `tools/generate-vectors/main.go` at 249 KB, text. `git check-ignore -v generate-vectors` → `.gitignore:12:/generate-vectors`, root-anchored, and `tools/generate-vectors/` stays tracked. **Fixed** — see §1 for whether the class is closed |
| **F13** arm 3's drive carve-out is dead code | removed | Removal is a **provable and measured no-op**: 0 classification differences over 16,159 spellings, plus the proof below. Arm 2's copy is still load-bearing — `M-drive-del2` dies on `valid-overlay-path-windows-backslash.json`. **Fixed**, with one measured consequence the commit message does not report (F18) |
| **F12** two of five flips pinned | `packages/team:context` and `github.com:\example\x` | Both pinned and both **load-bearing**: four mutants die only on the new cases. Both decisions are sourced (§2). Three flips remain unpinned (§3) |
| **F11** stale PR body | rewritten | Rewritten, and the vector count is now right. Three claims in it still do not reproduce (F17), and the same error survives untouched in `CHANGELOG.md` (F16) |
| **F14** `profiles/manager.md:2197` at 107 characters | reflowed | Paragraph now runs 62–75 characters. File's >90-character line count is **92** at head and **92** on `origin/main` — the regression is fully reversed. **Fixed** |
| **F15** bounds stated nowhere | a stated-bounds paragraph | Present in the PR body; four bounds stated, and each of the four is accurate against the head. Incomplete — see F17 |

### The F13 proof, checked exactly rather than nearly

Arm 3's `anyOf[1]` is now the bare SCP pattern
`^(?:[^/:\s@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\s\\]`, with the `not ^[A-Za-z]:[\\/]` removed. The two
languages are disjoint, and the argument closes:

* To match the drive pattern a string has a letter at index 0, `:` at index 1, and `/` or `\` at index 2.
* The SCP pattern's optional user part is `[^/:\s@]+@` — it needs an `@` at index ≥ 1 with no `:` before
  it. Index 1 is `:`, so the user alternative can never apply to a drive-matching string.
* Without a user part the host is `[A-Za-z0-9][A-Za-z0-9.-]*`, which cannot consume `:`. So the SCP
  colon must be the one at index 1, and the next character must satisfy `[^/\s\\]` — which excludes
  exactly `/` and `\`, the only two characters the drive pattern allows there.

Intersection empty. Confirmed empirically: re-adding the clause (`M-drive-restore3`) leaves the corpus
green and changes no classification across 16,159 spellings.

## 1. F10's class, and whether it deserves a gate

The artefact is gone and the ignore is correct. Two further checks:

* **No other build artefact rides along.** `go list ./tools/...` reports exactly one `main` package,
  `tools/generate-vectors`, so `/generate-vectors` is the complete set of root binaries the current
  package set can produce. `git ls-tree -r 18dca85` has 3 non-text blobs — two `expected/build-driver/*.preimage.bin`
  conformance fixtures and an empty `.txt` — and exactly one mode-`100755` file,
  `conformance/v1/fixtures/go-build-skill/scripts/skill-helper`. All four are on `origin/main` and
  untouched by this branch. 28 files added, 39 modified, 0 deleted, 0 renamed; every addition is a
  `schema-cases/manager-config-v2/*.json`.
* **The gap is real and I would gate it.** I read all three workflows. `ci.yml` runs
  `git diff --check` (whitespace) and `git diff --exit-code` scoped to `conformance/v1` and the five
  release JSONs; `implementations.yml` reads the conformance root; `release.yml` verifies signatures.
  Nothing looks at the tracked file set, which is why nine checks passed with a Mach-O binary present.
  Note also that CI can never *produce* the artefact — every lane uses `go run`, never `go build` — so
  a gate cannot be a "working tree is clean" check; it has to assert something about the committed set.

  A concrete assertion that would have failed on both accidents and passes at head: enumerate
  `git ls-files -z`, and fail on any tracked path that is (a) detected binary by `git grep -I`, or
  (b) mode `100755`, unless it is in a four-entry allowlist under `conformance/v1/`. At head that
  gate is green; against `2f2dfa4` it fails on `generate-vectors`; against a stray `tools/__pycache__`
  it fails on the `.pyc`. `.gitignore` now names two artefact classes discovered by two separate
  accidents and will keep growing one accident at a time until something asserts the invariant instead
  of the exclusions.

  **This belongs in its own leaf, not in this rework.** It is a CI change, outside this leaf's
  authority (`environments.md` §1/§6/§12.1, `core.md` §6.1, `profiles/manager.md`, the schema), and
  holding a correct schema for it would be the wrong trade.

## 2. The two newly pinned decisions, sourced

**`packages/team:context` is a valid form-free `path` overlay — correct.**

* It is not a §6.1 network form *at all*, not even an invalid one. §6.1's SCP form is
  `[user@]host:path` with host `[A-Za-z0-9][A-Za-z0-9.-]*`, a grammar that cannot contain `/`. The
  colon here comes after a `/`, so nothing before it can be a host. §6.1's "Invalid network forms MUST
  be rejected, not treated as local" therefore does not reach it — that sentence governs strings that
  *are* network forms and are malformed, which is exactly the `my_host:example/x` class arm 2 refuses.
* §1's `path` kind is "an operator-local package directory named by an absolute path, or by a
  project-relative path when the operation runs inside a project". `packages/team:context` is a legal
  POSIX project-relative directory name, so it is a `path` source.
* **Core §2's portable-relative-path grammar does not govern the operand**, and the change is right not
  to apply it. Rule 4 of §2 forbids a component containing `:`, and rule 3 forbids an absolute path —
  but §1 explicitly permits an absolute `path` operand, so §2 cannot be the operand's grammar without
  contradicting §1's own sentence. §1 says "portable relative path" where it means §2 — for `git`'s
  `directory`, which names a path *inside a protocol snapshot*. It deliberately does not say it here.
  A `path` operand names a platform directory, not a protocol object.
* Bound worth knowing, not a defect: this spelling is unrepresentable on Windows, where `:` is not a
  legal filename character. It fails there at resolution as `profile_source_path_missing` — loudly,
  and by the §1.1 rule that distinguishes missing from unreadable. The schema layer is not where a
  platform-specific operand rule belongs.

**`github.com:\example\x` is refused — correct, and §6.1 requires exactly this.** It *is* SCP-shaped
(`github.com` is a well-formed host), so §6.1 applies, and §6.1's third bullet says a protocol-1.0
network URL contains "no explicit port, password, query, fragment, percent escape, **or backslash**".
It is therefore an invalid network form, and "Invalid network forms MUST be rejected, not treated as
local" makes refusal — not reclassification to `path` — the only conforming outcome. Falling through
to `path` is what cycle 2 raised as F7b.

## 3. F12's other three flips

Measured, not reasoned. Each mutant's flip set was computed by classifying 16,171 spellings under the
committed and the mutated schema:

| still unpinned | flips | direction | class |
|---|---|---|---|
| `M-unanchored1` — arm 1's scheme test loses `^` | 8 | `path` → `refused` | a project-relative spelling with a literal `://` in a later segment (`ctx/http://y`). Every member contains `//`, i.e. a redundant separator |
| `M-unanchored-allow` — arm 3's allowlist loses `^` | 3 | `path` → `git` | `a/b/https://c` |
| `M-arm2-colonplus` — arm 2's colon test `*` → `+` | 1175 | `refused` → `path` | leading-colon spellings: `:`, `:x` |

I also measured that the arm-1-only and both-sites variants of `M-unanchored1` are identical, and that
`M-unanchored-allow` mutated at arm 1's `not` alone is a **0-flip semantic no-op** — arm 1's `not` is
only reached when the anchored scheme test already matched, so cycle 3's naming of it as a survivor
overstated it slightly.

**Leaving these three unpinned is defensible.** None admits a form-carrying `path` and none lets a
`git` source skip its form — the partition sweep found 0 of both across 16,159 spellings. Two fail
toward refusal, which an operator sees immediately; the third admits `:` as a path, which dies at
resolution. Each requires a spelling with a redundant `//` or a leading colon. They belong in the
stated bounds (F17), not in another rework cycle.

---

## Findings

### F16 — MAJOR. `repeat-of: F6 / F11`. `CHANGELOG.md` states that `file:` is an accepted scheme; the schema this PR ships refuses it

`CHANGELOG.md` at head, in the entry this branch adds:

> The git spelling follows the core section 6.1 canonical identity … and a `://` URL
> outside the `ssh`/`git`/`http`/`https`/**`file`** schemes is refused outright.

and, listing what the new cases pin:

> Windows spellings (`C:\…`, `C:/…`), project-relative paths, **`file:` URLs**, SCP forms with and
> without a user, each scheme in uppercase, and the refused unknown-scheme shape.

The committed allowlist at `18dca85` is `^([Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://` —
**no `file`**. Driven through the committed schema, `file:///Users/operator/context` and
`FILE:///Users/operator/context` are **refused in every column**, and two published cases pin it:
`invalid-overlay-file-url.json` and `invalid-overlay-file-url-with-form.json`, both `valid=false`.

`git log --oneline origin/main..18dca85 -- CHANGELOG.md` returns exactly one commit: **`bd39adb`**.
The entry was written against the cycle-1 head, where `file` *was* in the allowlist, and neither
rework touched it. Cycle 2 raised the `file:` classification as F6; cycle 3 raised the artefact
claiming `file:` among the positives as F11 and the PR body was rewritten for it — the in-repo copy of
the same two errors was not.

This is worse than the stale PR body cycle 3 rated MAJOR. The PR body lives on GitHub; `CHANGELOG.md`
lands in the specification repository permanently and is the document whose entire job is to tell an
implementer what changed. An implementer who reads it builds the wrong allowlist, and the repository
then contains a changelog contradicting its own schema and two of its own conformance cases.

**Fix:** drop `file` from the scheme list and move `file:` URLs from the pinned-positives list to
the refused list. One line and one clause. Nothing else in the entry is wrong — notably it states the
`branch` rule more accurately than the PR body does (see F17).

### F17 — MINOR. `repeat-of: F11`. Three claims in the rewritten PR body still do not reproduce, and the bounds paragraph is short

| PR body claims | measured at `18dca85` |
|---|---|
| "**Thirty-seven** overlay schema cases" | **39**. Thirty-seven was the count at `2f2dfa4`; this commit adds the two cases the body's own paragraph describes |
| "a `path` carrying **each of the five** §1-forbidden members" | four, not five. `range`, `tag`, `revision` and `directory` each have a `path`-source negative case; **`branch` has none on a `path` source** — `invalid-unknown-overlay-field.json` carries `branch` on a `git` source and the refusal comes from `additionalProperties: false`. The behaviour is genuinely gated (I killed `M-branch-prop`, `M-addprops` and `M-branch-both` on that case), so this is a wording defect, not a coverage gap. `CHANGELOG.md` states it correctly — "with `branch` rejected by the closed object as before" |
| "**Every mutant the reviews named** — permissive host class, host ≥2 characters, dropping `directory`, re-admitting `file:`, scheme length, unanchored colon test, backslash after the SCP colon — dies against a named case" | all seven named do die, verified. But cycle 3 named **thirty** mutants, three of which it recorded as surviving and which still survive (§3). Read as written the sentence is false; narrow it to "the mutants cycle 3 asked to be killed" and it is true |
| Stated bounds: four bounds | each of the four is accurate. Missing: the three surviving mutants of §3, and cycle 3's F15 bound that a one-letter prefix before `://` is a **drive path**, so `g://host/x` classifies as `path` (measured; harmless because no §6.1 scheme is one letter, but it is the direct consequence of the F5 repair and is the kind of thing a bounds paragraph exists for) |

### F18 — MINOR. `repeat-of: none`. The F13 simplification silently reduced mutation coverage, and the commit message reports only the half that is true

The commit message says arm 3's carve-out "was dead … Arm 2 keeps its own, where it is load-bearing."
Both halves are true of the *committed* schema, and I verified both. What it does not say is that the
removal took a case out of the kill path of a mutant that used to die:

```
M-drive-wide  (arm 2's carve-out ^[A-Za-z]:[\\/] widened to ^[A-Za-z]:)
  against the 2f2dfa4 schema : exit 1, valid-overlay-git-single-letter-host.json fails
  against the 18dca85 schema : exit 0, SURVIVES
```

At `2f2dfa4` the mutation hit both copies of the clause, and arm 3's widened copy pushed
`c:example/team-context` out of the `git` arm, failing the positive case. With arm 3's copy gone the
mutation is arm-2-local and no case notices. The class it now leaves unpinned is a bare drive letter —
`C:`, `C: ` — 42 flips, all `refused` → `path`, all degenerate. Cycle 3 counted this mutant as killed;
at head it is not.

The behaviour is unchanged and the simplification is correct. The finding is that "provably dead" was
established for the committed pattern only, and the evidence set it was holding up went unmeasured. A
`valid`/`invalid` case for `C:` would close it, or it joins the stated bounds.

---

## What I verified as correct, by measuring rather than reading

* **The partition holds.** Over 16,159 spellings: **0** sources valid both bare and form-carrying;
  **0** `path`-classified sources admitting `range`/`tag`/`revision`/`directory`/`branch`; **0**
  `git`-classified sources admitting two forms or `branch`. Full matrix in
  `TASK-260906-2x4s7i_classification-matrix-4.md`.
* **Every edge the brief named, and four of my own.** `svn://` refused; `file:///x` refused bare and
  with a form; `git@my_host:x`, `my_host:x`, `proj_ect:v2/ctx` refused; `github.com:/example/x` and
  `git@github.com:/example/x` refused (§6.1 wants a portable *relative* repository path); `git@host: x`
  refused; `C:foo`, `c:example/x`, `a:b` git-with-form; all four schemes in uppercase git; both SCP
  forms git; `.`, `..`, `team`, `packages/team-context` form-free paths; UNC `\\server\share\x` and
  scheme-relative `//host/x` form-free paths; `a/b:c` and `a/b:c\d` form-free paths; `:` , `:x`, `x:`,
  `C:` refused. A lone space is a form-free `path` — pre-existing `nonEmptyString` behaviour, resolves
  to `profile_source_path_missing`, outside this delta.
* **No bypass surface.** `#/$defs/overlay` is referenced exactly once, at line 361, by the `overlays`
  map. `system-config-v2` re-refs only `overlays_allowed`. There is no second schema through which a
  form-carrying `path` overlay could enter.
* **Gates, re-run by me as standalone processes.**
  `PATH=<venv>/bin:$PATH make validate` → **exit 0** — "validated 60 schemas and 1045 vector files";
  "Ran 227 tests in 44.379s / OK"; `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
  Regeneration: `go run ./tools/generate-vectors -root .` → **exit 0**, and I checksummed all 1051
  files under `conformance/v1` and `release/` before and after — **byte-identical** — then compared
  every one against the committed blob at `18dca85` — **0 mismatches**. The generator is the only
  writer, and `tools/generate-vectors/manager_config.go` is the only source of the two new cases.
* **`gh pr checks 47`:** 8 lanes pass (Formatting, Links, Specification ×3, Implementations ×3),
  "Release target provenance" skipping on a pull request by design. PR head OID `18dca85` = the
  reviewed commit. `mergeable: MERGEABLE`.
* **Signatures and identity.** All three commits verify **`G`** against the repository's own
  `maintainers.allowed_signers`; author *and* committer are `Ivan Oparin <oparin@me.com>` on each.
* **Pin consumption — read from both pins' source, not from a claim.** The Go pin
  `relux-works/curator@a3abcf34` contains **zero** occurrences of `manager-config-v2` in any `.go`,
  `.json` or `.tsv`; its `schema-cases` reads name directories explicitly (`build-receipt-v1`,
  `skill-build-v1`, `skillfile-dev-v2`, `install-marker-v2/v4`, `agent-skill-v6`, and
  `conformance_test.go`'s suite list) and none is manager-config-v2. `internal/interop/golden_test.go:290`
  reads `vectors/manager-config.json`; the diff touches only `vectors/manager-config-v2.json`.
  `ConformanceManifestSHA256 = b6f56aac…` pins the **rc.5** manifest, and `release/1.0.0-rc.5.json`
  is untouched — only rc.9 advanced, generator-written. The Python pin
  `ivanopcode/cocoaskills@3ecca1db` likewise has **zero** `manager-config-v2` references; its one
  generic index-walking consumer (`tests/protocol_conformance_adapters.py:279`, driven by
  `SCHEMA_CASES` filtered to `IN_SCOPE_SCHEMA_NAMES`) is deliberately **not** run by this job — the
  workflow carries an explicit note that the module authenticates one immutable rc.6 suite and returns
  when its pin advances to rc.9. Nothing either pin consumes gained a case it cannot pass.
* **The rename is still clean.** 909 index entries, **0** missing files. 69 manager-config-v2 index
  entries and 69 files, **0** orphans. The 25 orphans under `agent-environment-marker-v1` and
  `launch-env-fragment-v1` are exactly the 25 on `origin/main`. The old stem `overlay-path-file-url`
  appears nowhere. Both new names appear in `manifest.json`, `schema-cases/index.json` and
  `vectors/manager-config-v2.json`.
* **F14 measured against `main`, not against `2f2dfa4`.** `profiles/manager.md` has 92 lines over 90
  characters at head and 92 on `origin/main`. The paragraph reads 62–75.
* **No surface still asserts a universal form.** I swept every `.md` for the old sentence shape. The
  three remaining hits — `CHANGELOG.md:11`, `profiles/manager.md:2191`, `protocol/environments.md:2223`
  — are all correctly kind-conditional. `cli/curator.md:38` still writes
  `[<source> --range|--tag|--revision <ref>]` as one bracketed group, which under this file's own
  notation (compare `profile install`, which brackets the ref separately) means a source implies a
  ref — but `environments.md:2252` says in terms that "Informative CLI rows (`profile compose`,
  `env config`) that edit these knobs are the next batch's `cli/curator.md` work". The repository has
  already scheduled it by name. Not a finding; noted so the next batch does not lose it.
* **§1, §6 and §12.1 prose.** Across the whole branch `protocol/environments.md` changed **exactly one line** — 2223, the
  §12.1 knob row (`git diff -U0 origin/main..18dca85` reports a single `@@ -2223 +2223 @@` hunk), so
  §1's `path` paragraph and §6:833-838 are byte-identical to `origin/main`; the §12.1 row
  reads `{ source, range | tag | revision, directory?, weight? }` for `git` and `{ source, weight? }`
  for `path`, and `profiles/manager.md:2189-2203` agrees with it member for member, in the manager's
  voice, citing §12.1 and section 1.

---

## AC coverage — 6 of 7 rows pass, 1 fails on documentation

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | §12.1 knob row: form git-only, form-free path admitted | PASS | `protocol/environments.md:2223`; `profiles/manager.md:2189-2203` |
| 2 | schema requires a form only for a `git` source | PASS | committed schema via `Draft202012Validator`, 16,159 spellings, 0 partition violations |
| 3 | path + `range`/`tag`/`branch`/`revision`/`directory` stays invalid | PASS, 5 of 5 | `M-range`, `M-tag`, `M-revision`, `M-dir` die on named `path` cases; `branch` by `additionalProperties: false` — `M-branch-prop`, `M-addprops`, `M-branch-both` all die on `invalid-unknown-overlay-field.json` |
| 4 | published cases stop giving a path source a form | PASS | 39 overlay cases; no `valid=true` case gives a non-git source a form |
| 5 | positive + negative path-overlay cases published | PASS | 39 overlay schema cases, 22 distinct published `source` spellings (4 → 16 → 20 → 22 across the cycles) |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me, exit 0/0, regeneration byte-identical to the committed blobs; pin consumption read from both pins' source |
| 7 | signed commits, human identity, no stray file, **and the change documented** | **FAIL** | signatures `G`, identity correct, tracked set clean — but `CHANGELOG.md` contradicts the shipped schema (**F16**) |

## What unblocks acceptance

1. **F16** — one line in `CHANGELOG.md`: drop `file` from the accepted-scheme list, and move `file:`
   URLs out of the pinned-positives sentence. This is the only blocking item.
2. **F17** — 37 → 39; "each of the five" → four plus `branch` by the closed object; narrow "every
   mutant the reviews named"; add the three surviving mutants and the one-letter-scheme bound.
3. **F18** — pin `C:` with a case, or record it as a bound, and note that `M-drive-wide` is no longer
   killed.
4. Re-run both gates and re-read `gh pr checks 47` after the push.

Items 2–4 touch no schema, case, vector or normative sentence. **With F16 fixed, PR #47 is safe to
land.**

Reviewer artifacts, all inside `.temp/review-c4/` in the story worktree and attached to this element:
`TASK-260906-2x4s7i_classification-matrix-4.md`, `TASK-260906-2x4s7i_mutant-sweep-4.md`,
plus `matrix.py`, `classify.py`, `compare.py`, `mutants.py`, `survivors.py`, `partition.py`,
`validate.log`. Nothing outside `.temp/` was written and the review branch was not modified.
