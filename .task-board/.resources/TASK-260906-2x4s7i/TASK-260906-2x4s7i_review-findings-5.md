# Review findings — TASK-260906-2x4s7i (cycle 5, PR #47 at `407424e`)

**Verdict: ACCEPT.** `repeat-of: F17 / F11` (F20), `repeat-of: F16 / F6` (F21), `repeat-of: F2` (F19).
Three minor findings, no blocking and no major.

**PR #47 is safe to land.** The one thing cycle 4 said had to change — `CHANGELOG.md` telling
implementers `file:` was an accepted scheme while the schema refuses it — is fixed, and I verified the
fix against the committed schema and against both published `file:` cases rather than against the
diff. The one coverage regression cycle 4 measured is closed: `M-drive-wide` survived at `18dca85`
and dies at this head on the case this commit adds. Nothing in the delta could have moved behaviour —
the schema, `protocol/environments.md` and `profiles/manager.md` blobs are byte-identical to
`18dca85`, and the schema's SHA-256 is the same `d93a6673…f64d` cycle 4 recorded. Both gates are green
re-run by me, all eight PR lanes pass, the tracked file set is clean, and all four commits verify `G`
under the repository's own `maintainers.allowed_signers`.

The three findings are one unpinned classification decision and two prose imprecisions. None changes
what the schema does; each is a one-line follow-up or a sentence in the bounds. Holding a settled,
five-times-driven schema for them would be a defect in the review loop, not diligence.

Subject: `feat/path-overlay-declarable` at `407424e`, four signed commits past `origin/main` =
`550579d`. Cycle-4 delta `git diff 18dca85..407424e`. No Change Request is captured for this element;
the reviewed artefact is the PR.

Everything below was measured by me on an `rsync` copy of the clean `407424e` worktree, verified
byte-identical to the committed tree: 1,187 tracked paths, **0** content mismatches against
`git ls-tree -r 407424e` (`git hash-object --no-filters`), **0** missing, **0** extra.
`jsonschema==4.25.1`, the `requirements-dev.txt` pin. Nothing outside my scratch was written and the
review branch was not modified.

---

## 0. The delta cannot have changed behaviour, and that is provable rather than argued

The brief calls the delta three files; it is seven, of which five are generator output. The two
authored files are `CHANGELOG.md` and `tools/generate-vectors/manager_config.go` (+2 lines: one
schema example, one vector).

| artefact | `18dca85` | `407424e` |
|---|---|---|
| `schemas/v1/manager-config-v2.schema.json` | `569ca2cb…` | **`569ca2cb…`** |
| `protocol/environments.md` | `3edea3d5…` | **`3edea3d5…`** |
| `profiles/manager.md` | `da60d07c…` | **`da60d07c…`** |

Identical blobs, so no classification verdict can have moved. The corpus delta is purely additive and
I checked it member by member: one index entry appended (`valid: false`), one vector appended
(`valid: false`), one manifest entry plus two changed digests, and `release/1.0.0-rc.9.json`'s two
pin hashes. **No existing entry's `valid` flag was touched.** `release/1.0.0-rc.5.json` through
`rc.8.json` are untouched across the whole branch, and rc.5's `manifest_sha256` is still
`b6f56aac…`, the digest the pinned Go manager hardcodes.

Regeneration is the generator's own output and nothing else's: `go run ./tools/generate-vectors
-root .` exits 0, and re-hashing **all 1,187 tracked paths** afterwards against the committed blobs
gives **0 mismatches and 0 extra files**. That is a stronger statement than `git diff --exit-code`,
which the scratch copy cannot run.

## 1. F16 — the changelog now, clause by clause

Fixed. The entry at head reads `ssh`/`git`/`http`/`https`, with `file:` moved into the refused list
and sourced: "core section 6.1 gives it no network identity and the section 1 path operand names a
directory, not a URL". Both halves check out against the text — core §6.1's first sentence is "Local
paths and `file:` URLs have no network identity", and §1's `path` kind is "an operator-local package
directory named by an absolute path, or by a project-relative path". Driven through the committed
schema, `file:///Users/operator/context` and `FILE:///x` are refused in every column, and
`invalid-overlay-file-url.json` and `invalid-overlay-file-url-with-form.json` pin both. `M-file`
(re-admitting `file` to the allowlist) dies on the second.

Every classification claim in the entry's enumeration resolves against a published case — I checked
all twelve: `C:\…` → `valid-overlay-path-windows-backslash`, `C:/…` → `…-windows-slash`, `C://…` →
`…-windows-double-slash`, project-relative → `valid-overlay-path-relative`, colon in a later segment
→ `valid-overlay-path-colon-later-segment`, SCP with and without a user → `valid-overlay-git-scp`
and `…-scp-no-user`, single-character host → `valid-overlay-git-single-letter-host`, four uppercase
schemes → four positives, unknown scheme → `invalid-overlay-unknown-scheme`, `file:` URL →
`invalid-overlay-file-url`, SCP host outside the grammar → `invalid-overlay-scp-host-grammar`,
backslash after the SCP colon → `invalid-overlay-scp-backslash-path`.

Two clauses in it are still loose. They are F21 below, and they are not the F16 shape: F16 named a
scheme the schema refuses and no case contradicted it, whereas both of these are contradicted by a
published case in the same paragraph.

The entry's line-length profile is unchanged — 3 lines over 80 characters at head, 3 on `origin/main`.

## 2. F18 — the bare drive letter, and whether refusing it is right

**The pin works.** `invalid-overlay-bare-drive-letter.json` and vector
`schema2-overlay-bare-drive-letter` publish `{"source": "C:"}` as `valid: false`, and

```
M-drive-wide  (arm 2's carve-out ^[A-Za-z]:[\\/] widened to ^[A-Za-z]:)
  against 18dca85 (cycle 4) : exit 0, SURVIVES
  against 407424e (me)      : exit 1, invalid-overlay-bare-drive-letter.json fails
```

The case pins arm 2 and not something incidental: `C:` fails at schema path `…/allOf/1`
(`False schema does not allow`), and it turns **valid** under both `M-arm2-del` and `M-drive-wide`.
A negative case that is invalid for the wrong reason would not do that.

**Both readings agree, and I checked the sentences rather than the summary.**

* *As a network form.* core §6.1 requires a protocol-1.0 network URL to "contain a non-empty portable
  repository path". `C:` is `host` with an empty path, so it is an invalid network form, and
  "Invalid network forms MUST be rejected, not treated as local" makes refusal — not reclassification
  to `path` — the conforming outcome. This half is exact.
* *As a §1 path.* On Windows `C:` is **drive-relative**, not absolute: it denotes the current
  directory of drive C, which is per-process state, not a stable operator-local directory. §1 admits
  "an absolute path, or … a project-relative path when the operation runs inside a project"; a
  drive-relative reference is neither, since it is relative to a drive's working directory rather
  than to the project. So refusing it is substantively right on the platform it is a spelling of,
  not merely convenient.

**Where it is only nearly right, said plainly.** On POSIX, `C:` is a perfectly legal relative
filename — a colon is a legal POSIX path character, which is exactly why cycle 4 was right to admit
`packages/team:context` as a `path`. A top-level directory literally named `C:` is therefore a §1-legal
project-relative operand that this schema refuses. It is not a new bound: `notes:2026` and `a:b` are
classified `git` for the same reason, and the PR's own stated bounds say the schema "cannot
distinguish a directory spelled like a host from a real host". `C:` is the sub-case where the
host-shaped spelling has no path at all, so it falls to refusal rather than to `git`. The bounds
paragraph would be more honest naming that sub-case, but the behaviour is the same decision §1's
missing kind marker forces everywhere else, and the operator sees a refusal at config time rather
than a silent misresolution.

So: **the behaviour an operator on Windows should get.** Not merely what was easy to pin.

## 3. F17 — the description at head

Every number reproduces. **Forty** overlay schema cases (I counted the index: 910 entries, 70
manager-config-v2, 40 overlay-named, 0 orphans, 0 missing files, 23 distinct published `source`
spellings). `make validate` exit 0 with "60 schemas and 1046 vector files" and "Ran 227 tests … OK" —
the body's three numbers, verbatim. `git ls-files` carries no build artefact: 1,187 tracked paths,
two binary blobs and one mode-`100755` file, all three pre-existing on `origin/main`; 29 files added
by the branch, every one a `schema-cases/manager-config-v2/*.json`.

The enumerated spellings all resolve to cases, including the four cycle-4 asked for: "a `path`
carrying `range`, `tag`, `revision` and `directory`" is now the accurate four rather than the wrong
five, and `branch` is correctly attributed to `additionalProperties: false` in the bounds. All eight
mutants the body names as dying do die on the case it implies.

One claim still does not hold and one bound is still missing — F20 below.

## 4. Regression surface of the delta

Covered in §0 and re-measured here: gates re-run by me as standalone processes on the clean copy.

| gate | result |
|---|---|
| `make validate` | **exit 0** — `validated 60 schemas and 1046 vector files`; `Ran 227 tests in 45.378s / OK`; `ok github.com/relux-works/curator-spec/tools/generate-vectors` |
| regeneration | `go run ./tools/generate-vectors -root .` **exit 0**; all 1,187 tracked paths re-hashed against the committed blobs — **0 mismatches, 0 extra files** |
| `gh pr checks 47` | 8 lanes **pass** (Formatting, Links, Specification ×3, Implementations ×3); "Release target provenance" *skipping* on a pull request by design. PR head OID `407424e` = the reviewed commit; `mergeable: MERGEABLE` |

`__pycache__/` and `/generate-vectors` are both ignored (`git check-ignore -v` names `.gitignore:8`
and `.gitignore:12`), and `make validate` did regenerate `tools/__pycache__` in my copy — the ignore
holds.

---

## Findings

### F19 — MINOR. `repeat-of: F2`. A lowercase Windows drive letter is unpinned, and it is the one class this change exists to make declarable

```
M-drive-anycase-off  (arm 2's carve-out ^[A-Za-z]:[\\/] narrowed to ^[A-Z]:[\\/])
  → SURVIVES the whole committed corpus, exit 0
  → 64 flips over 1,694 spellings, every one path → refused
```

The flip class is `c:\users\operator\context`, `c:/users/operator/context`, `d:/ctx`, `z:\x` — an
entirely ordinary Windows spelling, and Windows drive letters are case-insensitive. All three
published Windows positives spell the drive `C`, so a narrowing of the carve-out to uppercase lands
green and re-creates the **F1 class**: a legal `path` overlay undeclarable on a supported platform.

The committed behaviour is correct — I measured `c:\users\operator\context` as a form-free `path`.
This is a coverage finding, not a defect, and it is the cheapest one left: one line in
`managerConfigV2SchemaExamples`. It is a stronger case for pinning than the four survivors cycle 4
left as bounds, because their flip classes are degenerate (a leading colon, a redundant `//`) and
this one is a spelling operators type.

Two more survivors, both new and both weaker, belong in the same paragraph rather than in a rework:

* `M-drive-multiletter` (carve-out `^[A-Za-z]+:[\\/]`) — 109 flips, `refused` → `path`, the class
  being an alphabetic multi-letter prefix before a separator (`abc:/x`, `Users:\x`).
* `M-scp-bslash-anywhere` (SCP path forbids a backslash anywhere, not only at the first character) —
  35 flips, `git` → `refused`, the class being `github.com:example\x` and
  `git@github.com:example/x\y`. See F21 for why this one also touches the prose.

Full table, flip sets and directions: `TASK-260906-2x4s7i_mutant-sweep-5.md`.

### F20 — MINOR. `repeat-of: F17 / F11`. The PR body contradicts itself on the mutants, and F17's one-letter-scheme bound is still missing

The body says:

> **Every mutant the four reviews named as unpinned now dies against a named case** — permissive host
> class, host ≥2 characters, dropping `directory`, re-admitting `file:`, scheme length, unanchored
> colon test, backslash after the SCP colon, and the widened drive pattern.

and, four sentences later, in the stated bounds:

> Three mutants reported by cycle 3 remain survivors and are recorded there rather than claimed dead
> here.

Both cannot be true, and the second is the accurate one — I reproduced all three (`M-arm2-colonplus`,
`M-unanchored1`, `M-unanchored-allow`) surviving at this head. Cycle 4 asked for this sentence to be
narrowed to "the mutants cycle 3 asked to be killed"; it was instead re-qualified as "named as
unpinned", which is the same claim. The eight mutants the em-dash actually lists **do** all die — I
verified each — so narrowing the sentence to its own list makes it true.

The bound cycle 4 named as missing is still missing: a **one-character prefix before `://` is a drive
path**, so `g://host/x` classifies as `path` (measured). That is the direct consequence of the F5
repair and exactly the kind of thing a bounds paragraph exists for. The bounds do state the
one-character *host* (`c:example/x` → `git`), which is a different decision.

### F21 — MINOR. `repeat-of: F16 / F6`. Two clauses of the changelog entry overstate what the schema encodes

Both are contradicted by a published case in the same paragraph, which is what keeps them out of
F16's class and off the blocking list.

| clause at head | measured against the committed schema |
|---|---|
| "A `://` URL outside the `ssh`/`git`/`http`/`https` schemes is refused outright" | false for a one-character scheme. `g://host/x` is a form-free **`path`**, because arm 1's scheme test is `^[A-Za-z][A-Za-z0-9+.-]+://` — two characters minimum, deliberately, so that `C://Users/operator/context` stays a Windows path. The same sentence's own enumeration lists `C://…` among the pinned positives, and `valid-overlay-path-windows-double-slash.json` settles it |
| "The git spelling follows the core section 6.1 canonical identity (… and no backslash)" | the encoded rule is *no backslash immediately after the SCP colon*. `github.com:example\x`, `git@github.com:example/x\y` and `https://h/x\y` are all admitted as `git`. The entry's own refused-shapes list says it precisely — "a backslash after the SCP colon" — so the document contains both spellings of the rule |

The second is not a schema defect, and I want to be exact about why, because the asymmetry looks
arbitrary until you name the sentence. core §6.1's operative mandate is "Invalid network forms MUST be
rejected, **not treated as local**". The schema refuses an invalid network form precisely when the
alternative would be `path` — `git@my_host:x` and `github.com:\example\x` would otherwise fall
through to local, so arm 2 refuses them. A `://` URL with a port, password, query, fragment, percent
escape or a later backslash is already classified `git`, so it is not treated as local and the mandate
is satisfied; §1.1 then reports `profile_source_invalid` at resolution. I measured that whole family
(`https://h:22/x`, `https://user:pw@h/x`, `https://h/x?q=1`, `https://h/x#frag`, `https://h/x%20y`,
`https://h/x\y`) as `git`-with-a-form, and it is consistent with the PR's stated bound that "the
schema layer decides admissibility, not which §1.1 diagnostic an implementation reports". The finding
is the wording, not the rule.

---

## What I verified as correct, by measuring rather than reading

* **The partition holds.** Over **2,258** generated spellings: **0** sources valid both bare and
  form-carrying; **0** `path`-classified sources admitting `range`/`tag`/`revision`/`directory`/`branch`;
  **0** `git`-classified sources admitting two forms, `branch`, or skipping their form; **0**
  spellings falling outside the three-way partition. Split 1,238 `path` / 377 `git` / 643 `refused`.
* **§1's legal combinations, both directions.** `git` + `range` + `directory` valid (§1: "A `git`
  declaration MAY carry `directory`"); `git` + `range` + `weight` valid; `path` + `weight` valid;
  `path` + `directory` invalid; `path` + each of `range`/`tag`/`revision` invalid; `path` + `branch`
  invalid by the closed object. Full grid in `TASK-260906-2x4s7i_classification-matrix-5.md`.
* **37 of 45 mutants die on a named case**, only 8 of them deletions; two survivors measured as
  0-flip semantic no-ops, six as real unpinned decisions with their flip classes enumerated.
* **The manager sentence, read myself.** `profiles/manager.md:2189-2197` states the split as a
  manager-side obligation — "the manager requires the `range | tag | revision` form (with optional
  `directory?`) on a `git` source and no requirement form on a `path` source, which carries only
  `source` and optional `weight?`" — citing the environments §12.1 shape and the section 1 rules,
  member for member with the amended knob row at `protocol/environments.md:2223`. Manager's voice,
  not a restatement of the norm.
* **`protocol/environments.md` changed exactly one line** across the whole branch — a single
  `@@ -2223 +2223 @@` hunk — so §1's `path` paragraph and §6 are byte-identical to `origin/main`.
* **Pin consumption, read from the pin's source.** `relux-works/curator@a3abcf34` contains **zero**
  occurrences of `manager-config-v2` anywhere in the tree, and every `schema-cases` reader in it names
  an explicit directory (`build-receipt-v1`, `skill-build-v1`, `skillfile-dev-v2`, `install-marker-v2`,
  `install-marker-v4`, `agent-skill-v6`, and `conformance_test.go`'s suite list) — none is
  manager-config-v2. Its `ConformanceManifestSHA256 = b6f56aac…` pins the rc.5 manifest, and
  `release/1.0.0-rc.5.json` is untouched with that exact digest. The three Implementations lanes are
  green at this head, which is the direct empirical confirmation.
* **Mechanics.** Four commits past `origin/main`; author *and* committer `Ivan Oparin
  <oparin@me.com>` on each; all four verify **`G`** under the repository's own
  `maintainers.allowed_signers`. No stray file, no `tools/__pycache__` tracked, no root binary.

## AC coverage — 7 of 7

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | §12.1 knob row: form git-only, form-free path admitted | PASS | `protocol/environments.md:2223`; `profiles/manager.md:2189-2197`, both read by me |
| 2 | schema requires a form only for a `git` source | PASS | committed schema via `Draft202012Validator`, 2,258 spellings, 0 partition violations |
| 3 | path + `range`/`tag`/`branch`/`revision`/`directory` stays invalid | PASS, 5 of 5 | `M-range`, `M-tag`, `M-revision`, `M-dir` die on named `path` cases; `branch` via `M-branch-prop`, `M-addprops`, `M-branch-both` on `invalid-unknown-overlay-field.json` |
| 4 | published cases stop giving a path source a form | PASS | 40 overlay cases, 23 distinct spellings; no `valid=true` case gives a non-git source a form |
| 5 | positive + negative path-overlay cases published, and the new ones kill | PASS | `M-drive-wide` dies on `invalid-overlay-bare-drive-letter.json`; the case fails at `allOf/1` and goes valid under `M-arm2-del` |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me, exit 0/0, regeneration byte-identical over 1,187 paths; pin consumption read from source; three Implementations lanes green |
| 7 | signed commits, human identity, no stray file, change documented | PASS | four `G` signatures, tracked set clean, `CHANGELOG.md` no longer contradicts the schema |

## The landing question

**PR #47 is safe to land, without qualification.** The behaviour is settled: five cycles have attacked
this discriminator, the last two found no behavioural defect at all, and this cycle found none. The
three findings above are one unpinned mutant class and two sentences; each is a follow-up leaf or a
line in the bounds paragraph, and none of them can change a single classification verdict. If they are
worth doing at all — and F19 is, at one line — they are worth doing after this lands, on a base that
is already correct.

Reviewer artifacts, all in `.temp/review-c5/` in the story worktree and attached to this element:
`TASK-260906-2x4s7i_classification-matrix-5.md`, `TASK-260906-2x4s7i_mutant-sweep-5.md`,
`TASK-260906-2x4s7i_validate-5.log`, plus `matrix.py`, `mutants.py`, `flips.py`. Nothing outside
`.temp/` was written and the review branch was not modified.
