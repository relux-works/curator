# Review findings — TASK-260906-3x0w4y (cycle 2, rework 1, PR #47)

**Verdict: CHANGES REQUESTED.** `repeat-of: F1` (F5), `repeat-of: F4` (F6, F7b),
`repeat-of: F2` (F7a, F8). One blocking, three major, one minor.

**PR 47 is not safe to land as it stands.** The F1 repair is real and correct on the two
canonical Windows spellings, but the arm added for F4 hard-refuses a third one that
`origin/main` accepted, and the half of the F1 repair that the producer's own rationale is
built on — the core §6.1 host grammar — still has no case that can fail it in either
direction.

Subject: `feat/path-overlay-declarable` at `bd39adb`, one signed commit past
`origin/main` = `550579d`. Reviewed delta: `git diff origin/main..bd39adb`
(58 files, +3027/−75). No Change Request captured for this element; the artefact
under review is the PR, per the cycle-2 brief.

Everything below was measured by me against the **committed** schema through
`Draft202012Validator` with the repo's own registry (`tools/validate.py`'s entry point,
`jsonschema==4.25.1` from `requirements-dev.txt`), on a byte-exact rsync copy of the
`bd39adb` checkout (1178 files, identical aggregate SHA-256 to the PR worktree). The
classification was independently reproduced in Go under RE2 — all three patterns compile
and agree cell for cell with the Python run.

---

## 0. What the fix got right — verified by me, not read

* **F1 is genuinely fixed on the two canonical Windows spellings.** Driven through the
  committed schema, both directions:

  | source | bare | +range | +tag | +revision | +directory | +branch | +weight |
  |---|---|---|---|---|---|---|---|
  | `/Users/operator/context` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `C:\Users\operator\context` | **VALID** | INVALID | INVALID | **INVALID** | INVALID | INVALID | VALID |
  | `C:/Users/operator/context` | **VALID** | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `D:\ctx` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `packages/team-context`, `./here`, `../sibling`, `team`, `.`, `..` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `\\server\share\ctx` (UNC) | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `//host/x` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID |
  | `https://…`, `HTTPS://…`, `SSH://…`, `GIT://…`, `HTTP://…` | INVALID | VALID | VALID | VALID | INVALID | INVALID | INVALID |
  | `git@github.com:example/x`, `github.com:example/x`, `user@host:path` | INVALID | VALID | VALID | VALID | INVALID | INVALID | INVALID |
  | `svn://…`, `SVN://…`, `ftp://…`, `gitfoo://…` | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID |
  | `` (empty) | INVALID everywhere (`nonEmptyString`) | | | | | | |

  The cycle-1 blocking shape is dead in both directions, UNC and scheme-relative absolute
  paths land on the correct arm, and F4's `svn://` is refused bare **and** with a form.
* **The published corpus stopped teaching the wrong shape.** The overlay in
  `managerConfigV2EveryKnob` lost its `revision`; the only remaining path-source-with-a-form
  instances in `schema-cases/` and `vectors/manager-config-v2.json` are the seven intentional
  `valid=false` negatives. AC row 4 holds.
* **Corpus breadth is a real improvement.** Distinct overlay `source` spellings went from
  4 to 16.
* **17 of my 21 mutants die on named cases** (§3). `branch`, `range`, `tag`, `revision` on the
  path arm, the two-form `oneOf`, the scheme branch, the SCP branch, the case-insensitivity,
  the backslash exclusion and the unknown-scheme refusal are all pinned by a case that
  names itself in the failure.
* **Gates re-run by me as standalone processes on the committed tree, not trusted from the PR body:**
  * `make validate` → **exit 0** — `validated 60 schemas and 1037 vector files`;
    `Ran 227 tests in 45.271s / OK`; `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
  * `make regenerate-check` → **exit 0**, empty diff. (The producer reported exit 1 "by
    construction" from an uncommitted workspace; on the committed tree it is genuinely green.)
  * `gh pr checks 47` → all lanes **pass**: Formatting, Links, Specification and
    Implementations on `ubuntu-latest`, `macos-latest` **and `windows-latest`**;
    "Release target provenance" skipping. PR head OID = `bd39adb`.
* **Pin consumption clean — read from the pin's source, not from the report.** At
  `relux-works/curator@a3abcf34`: `internal/skillspec/conformance_test.go` enumerates only
  `agent-skill-v7/v8` and `csk-skill-v7/v8` by name and never opens
  `schema-cases/index.json`; `internal/interop/golden_test.go:290` reads
  `vectors/manager-config.json`, which this diff leaves at **0 changed files**;
  `internal/buildrepo/buildrepo.go:32 ConformanceManifestSHA256 = b6f56aac…` is the
  **rc.5** manifest and rc.5 is untouched — only the rc.9 *candidate* pin advanced, which
  the generator writes. The three green Implementations lanes are that check executed.
* **Mechanics clean, better than cycle 1 could establish.** Exactly one commit past
  `origin/main`; author and committer `Ivan Oparin <oparin@me.com>`; the signature verifies
  **`G` (good)** against the repository's own `maintainers.allowed_signers`, not merely `U`.
  No `LOGBOOK.md`, no `.DS_Store`, no `tools/__pycache__` — and `__pycache__/` really is
  ignored on this base (`git check-ignore` exit 0, `.gitignore:8`), so the cycle-1 trap is
  gone.
* **Manager sentence (F3) is correct.** It agrees with the amended §12.1 row member for
  member, stays in the manager's voice (obligation + cite, not a normative restatement), and
  uses `environments §12.1`, which is that file's dominant convention (4 prior uses). Bare
  "section N" for environments sections matches the same paragraph's pre-existing
  "the section 1 warning when the key is locked". F3 closed.
* **The predecessor's repair pattern is not merely indistinguishable from the committed
  one — it *is* the committed one**, byte for byte. The producer disclosed adopting it. The
  brief's question is therefore moot, but worth stating: the first arm was never
  independently re-derived, which is part of why F7 slipped through.

---

## Findings

### F5 — BLOCKING. `repeat-of: F1`. The arm added for F4 hard-refuses a Windows absolute path that `origin/main` accepted

The new second `allOf` element decides "this is a URI with a scheme" with

```
^[A-Za-z][A-Za-z0-9+.-]*://
```

`*` admits a **one-character** scheme, and a Windows drive letter is exactly one character.
So `C://Users/operator/context` matches, `C` is not in the `{ssh,git,http,https,file}`
allowlist, and `then: false` fires. Measured against the committed schema:

| source | bare | +range | +tag | +revision | +directory | +branch | +weight |
|---|---|---|---|---|---|---|---|
| `C://Users/operator/context` | **INVALID** | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID |
| `D://ctx` | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID |
| `c://x`, `x://y` | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID |

**Eight of eight INVALID: the declaration has no accepted shape at all.** Reproduced in Go
under RE2 (`C://Users/operator/context -> REFUSED`).

**This is a regression against `origin/main`, and that argument needs no Windows trivia.**
I drove the same spellings through `origin/main`'s `$defs/overlay`: under the universal-form
rule `{"source": "C://Users/operator/context", "revision": …}` was **VALID**. This diff
removes the only shape that spelling ever had. A change whose stated purpose is to *make a
declaration possible* must not leave a previously-accepted one with nowhere to go.

It is also a §1-permitted declaration. §1: a `path` source is "an operator-local package
directory named by an **absolute path**" — unqualified. Win32 path canonicalisation collapses
redundant separators after the drive root, so `C://Users/operator/context` names exactly the
directory `C:/Users/operator/context` names — and that spelling the producer thought
important enough to publish a positive case for (`valid-overlay-path-windows-slash`). The
doubled separator is not exotic: it is what naive path joining produces from a root ending
in `/`.

**The two arms of this change disagree about what `C:` is.** The first arm goes to
deliberate lengths — the §6.1 host grammar plus a backslash exclusion — to keep a drive
letter *out* of the git classification. The second arm then readmits it as a "scheme" and
refuses it outright. The arm is not implementing any principled notion of a scheme either:
`1://y` is **VALID** (a path) and `x://y` is **INVALID**, because the scheme test starts with
`[A-Za-z]`. Nothing in §1 or core §6.1 draws that line.

Answering the brief's dimension-4 question directly: **no, the second arm cannot be shown to
refuse nothing §1 permits — it refuses this.**

**The repair is free — one character.** I changed `*` to `+` in that arm's scheme pattern
(a scheme is at least two characters; a drive letter is exactly one) and re-ran the entire
committed corpus:

```
validate.py exit: 0   validated 60 schemas and 1037 vector files
  'C://Users/operator/context'   bare=VALID    +revision=INVALID
  'D://ctx'                      bare=VALID    +revision=INVALID
  'svn://host/x'                 bare=INVALID  +revision=INVALID
  'ftp://host/x'                 bare=INVALID  +revision=INVALID
  'https://github.com/example/x' bare=INVALID  +revision=VALID
```

Zero regressions, F4's refusal intact, the drive-letter class restored in both directions.
Illustrative, not prescriptive — the producer owns the fix — but it proves the bound is a
one-character miss, not a constraint. **And note the corpus stayed green under the change,
which is F8's point: no case pins this arm anywhere near a drive letter.**

### F6 — MAJOR. `repeat-of: F4`. A `file:` URL is admitted as a form-free `path`, and the claim is now pinned into the conformance suite

Measured: `{"source": "file:///Users/operator/context"}` → **VALID**, form-free;
`FILE:///x` and `file://server/share/x` likewise. The producer added `[Ff][Ii][Ll][Ee]` to
the second arm's allowlist deliberately and published
`valid-overlay-path-file-url.json` plus `schema2-overlay-path-file-url` to pin it.

Deciding it from §1, as the brief asks:

* **Not a `git` source.** §1: "`git` — a **network** git source under the core §6.1
  canonical identity". core §6.1: "Local paths and `file:` URLs have no network identity."
  The producer's half of the argument is right.
* **Not a `path` source.** §1: "an operator-local package directory named by an **absolute
  path**, or by a project-relative path"; the very next sentence: "**The operand names a
  directory** whose root contains `agent-context.json`", and §1.1 gives that operand
  `profile_source_path_missing` / `profile_source_path_unreadable`. A `file:` URL is a URL,
  not a path operand. `file:` appears **nowhere** in `protocol/environments.md`. The CLI
  operand is `curator profile install <git-url|path>` — the same two kinds, no third.
* §1's kind set is closed. Neither kind ⇒ §1.1 `profile_source_kind_unsupported`.

The rationale in the rework report uses a sentence about what a `file:` URL **is not** (no
network identity) to conclude what it **is** (a path). That inference does not hold, and it
leaves the change internally inconsistent: `svn://host/x` and `file:///x` are both URLs
naming no §1 source kind, and one is refused outright while the other is admitted with no
requirement form at all.

This is worse than the cycle-1 `svn://` gap it mirrors, which is why it is major rather than
minor: an unexercised gap teaches nothing, but `valid-overlay-path-file-url.json` is a
**published positive conformance case** that every implementation must now reproduce.

Cost: drop `[Ff][Ii][Ll][Ee]` from the allowlist and delete or invert one positive case —
`invalid-overlay-path-file-requirement-form` stays `valid=false` either way (it moves from
the path arm to the refusal arm). My M-nofile mutant confirms the positive case is the only
thing holding it: `validation failed: schema case
manager-config-v2/valid-overlay-path-file-url.json … expected valid=True, got False`.

If the project *wants* `file:` URLs admitted, that is a §1 amendment, not an allowlist entry.
Either way the schema must stop deciding it alone.

### F7 — MAJOR. `repeat-of: F2` and `repeat-of: F4`. The §6.1 host grammar — half of the F1 repair — is unpinned in **both** directions, and it silently sends invalid network forms to `path`

**(a) No case can fail it.** Two mutants that change the host class survive the entire
committed corpus:

* **M-host**: `[A-Za-z0-9][A-Za-z0-9.-]*` → `[^/:\s]+` (the cycle-1 permissive class), backslash
  exclusion kept. `validate.py` **exit 0**.
* **M-host2**: `[A-Za-z0-9][A-Za-z0-9.-]*` → `[A-Za-z0-9][A-Za-z0-9.-]+` (host must be ≥2
  characters, i.e. drive letters excluded). `validate.py` **exit 0**.

This is F2's exact shape recurring on the mechanism F2 was about. The rework report §1
presents the host grammar and the backslash exclusion together as the F1 repair; only the
backslash exclusion is load-bearing (M-bs and M-old both die on
`valid-overlay-path-windows-backslash`). The host grammar carries no evidence at all, and
the corpus cannot distinguish the shipped discriminator from the pre-rework one on that axis
or from a stricter one.

**(b) And the change it silently made is the wrong direction.** Because every host outside
`[A-Za-z0-9][A-Za-z0-9.-]*` now falls through to the path arm:

| source | main | revision 1 | `bd39adb` | what §1/§6.1 says |
|---|---|---|---|---|
| `git@my_host:x` (underscore host) | git, form required | git, form required | **path, admitted form-free** | not a §6.1 host ⇒ invalid network form |
| `my_host:x` | git, form required | git, form required | **path, admitted form-free** | same |
| `git@host:/x` (leading-slash path) | git, form required | path | **path, admitted form-free** | the repo's own `validate_repository_path` rejects a leading `/` |

core §6.1, the authority this discriminator claims to encode: **"Invalid network forms MUST
be rejected, not treated as local."** `git@my_host:x` is now treated as local — silently, and
form-free. `git@my_host:x` + `revision` is refused, so it is also undeclarable as a git
overlay. The operator gets `profile_source_path_missing` for a source they wrote as a git
URL.

For the `my_host` row this is a **direction change introduced by this diff**, from the safe
one (demand a form, fail loudly at identity validation) to the unsafe one (accept as a
form-free local directory).

Requested: decide the class from core §6.1 and publish a case that pins the decision, in
whichever direction. Whatever you choose, M-host and M-host2 must both die.

### F8 — MAJOR. `repeat-of: F2`. `directory` on a `path` source is refused by no case — and it is a named member of the AC

**M-dir**: drop `{"required": ["directory"]}` from the path arm's `not/anyOf` — a narrowing
mutant that admits exactly one member of the class. `validate.py` **exit 0, corpus green.**

So `{"source": "/Users/operator/context", "directory": "packages/team"}` would be admitted
and nothing in the suite notices. §1 is explicit: "A `path` declaration that carries `range`,
`tag`, `branch`, `revision`, or `directory` is `profile_source_invalid`", and the AC repeats
all five by name.

The other four are pinned and I killed a mutant for each:

| member | mutant | named case that fails |
|---|---|---|
| `range` | M-range | `invalid-overlay-path-windows-slash-requirement-form.json` |
| `tag` | M-tag | `invalid-overlay-path-relative-requirement-form.json` |
| `revision` | M1 | `invalid-overlay-path-requirement-form.json` |
| `branch` | M-branch, M-branch2 | `invalid-unknown-overlay-field.json` |
| `directory` | **M-dir** | **none — survivor** |

**4 of 5.** `directory` is the likeliest of the five to be written by mistake on a path
overlay, because it is the one member that sounds path-shaped. One negative case closes it.

### F9 — MINOR (three items, no rework blocked on its own)

* **`C:foo` — the residual edge, decided.** It classifies `git` and therefore demands a form.
  On Windows `C:foo` is *drive-relative* — resolved against the process's current directory
  on drive C — so it is neither "an absolute path" nor "a project-relative path", and §1
  admits it as neither kind. Misclassified toward `git` it is refused loudly at identity
  validation and can never be silently accepted as a form-free local directory: the safe
  direction. **Not a defect — but it belongs in the stated bounds.** The rework report §4
  states only `a:b/c`; add `C:foo` (and, once F5 is fixed, whatever the drive-letter rule
  becomes).
* **M-unanchored survivor.** Removing `^` from the second arm's scheme pattern leaves the
  corpus green. Reachable only through contrived sources such as `/x/svn://y`; low value, but
  say so as a bound rather than leaving it silent.
* **`profiles/manager.md:2197` is 107 characters** in a paragraph that wraps at 66–75. The
  edit did not reflow the tail. Cosmetic; the file's long-line count goes 106 → 107.

---

## 3. Mutant table — 21 mutants, 17 killed, 4 survivors

Each runs the repo's own gate (`python3 tools/validate.py`) as a standalone process against
the committed corpus; the schema was restored byte-exact afterwards (asserted in the
harness). Every mutant either narrows the gate to admit exactly one member of the class it
must reject, or preserves the searched-for tokens while changing behaviour — no delete-only
mutants.

| mutant | exit | named case that fails | outcome |
|---|---|---|---|
| M0 control | 0 | `validated 60 schemas and 1037 vector files` | green (expected) |
| M1 path `else` drops `revision` | 1 | `invalid-overlay-path-requirement-form.json` | killed |
| M2 git `then` `oneOf`→`anyOf` | 1 | `invalid-overlay-two-requirement-forms.json` | killed |
| M3 discriminator → `^/`, arms swapped (cycle-1 survivor) | 1 | `valid-overlay-path-windows-backslash.json` | killed |
| M-old cycle-1 pattern verbatim (token-preserving) | 1 | `valid-overlay-path-windows-backslash.json` | killed |
| M-bs drop only the backslash exclusion | 1 | `valid-overlay-path-windows-backslash.json` | killed |
| M-swap committed pattern, arms swapped | 1 | `valid.json` | killed |
| M-noarm2 drop the whole second `allOf` arm | 1 | `invalid-overlay-unknown-scheme.json` | killed |
| M-svn widen allowlist with `svn` | 1 | `invalid-overlay-unknown-scheme.json` | killed |
| M-nofile narrow allowlist: drop `file` | 1 | `valid-overlay-path-file-url.json` | killed |
| M-lower scheme alternation lowercase only | 1 | `valid-overlay-git-ssh-uppercase.json` | killed |
| M-noscheme git arm keeps only the SCP branch | 1 | `valid.json` | killed |
| M-noscp git arm keeps only the scheme branch | 1 | `valid-overlay-git-scp.json` | killed |
| M-branch overlay admits `branch` again | 1 | `invalid-unknown-overlay-field.json` | killed |
| M-branch2 `branch` admitted, banned only on the path arm | 1 | `invalid-unknown-overlay-field.json` | killed |
| M-range path `else` drops `range` | 1 | `invalid-overlay-path-windows-slash-requirement-form.json` | killed |
| M-tag path `else` drops `tag` | 1 | `invalid-overlay-path-relative-requirement-form.json` | killed |
| M-hostcls host class → `[A-Za-z0-9]+` (no dot/hyphen) | 1 | `valid-overlay-git-scp.json` | killed |
| **M-host** host class → `[^/:\s]+` (cycle-1 permissive) | 0 | — | **SURVIVOR (F7a)** |
| **M-host2** host must be ≥2 chars | 0 | — | **SURVIVOR (F7a)** |
| **M-dir** path `else` drops `directory` | 0 | — | **SURVIVOR (F8)** |
| **M-unanchored** second-arm scheme pattern loses `^` | 0 | — | **SURVIVOR (F9)** |

M4 (the §12.1 Values-cell prose revert) is unchanged from cycle 1 and remains an accepted
bound: prose in a normative table whose behaviour is pinned by schema + cases + vectors, with
`validate.py` machine-checking knob names, defaults and enums.

---

## 4. AC coverage — 6 of 7 rows fully driven, 1 row 4 of 5

| # | AC row | Result | Driven through |
|---|---|---|---|
| 1 | 12.1 knob row: form git-only, form-free path admitted | PASS | `protocol/environments.md:2222`; wording bound M4 |
| 2 | schema requires a form only for a `git` source | **PARTIAL** | committed schema via `Draft202012Validator` + Go RE2 — correct for POSIX, relative, UNC, `C:\`, `C:/`, all four schemes in both cases, SCP with and without a user; **wrong for `C://…` (F5), `file:` (F6), invalid SCP hosts (F7)** |
| 3 | path + `range`/`tag`/`branch`/`revision`/`directory` stays invalid | **4 of 5** | M1, M-range, M-tag, M-branch each die on a named case; **`directory` has no case (M-dir survives, F8)** |
| 4 | published cases stop giving a path source a `revision` | PASS | corpus scan: the seven remaining path+form instances are all `valid=false` negatives |
| 5 | positive + negative path-overlay cases published | PASS | 20 schema cases + 20 vector twins + index entries |
| 6 | `make validate` + `make regenerate-check` green; pin gains no unpassable case | PASS | both re-run by me, exit 0/0; pin consumption read from `curator@a3abcf34` source; three green Implementations lanes incl. `windows-latest` |
| 7 | one signed commit, human identity, no stray file | PASS | signature `G` against `maintainers.allowed_signers`; `__pycache__/` ignored on this base |

---

## What unblocks acceptance

1. **F5** — stop the second `allOf` arm refusing a drive-letter path. Requiring a scheme of
   at least two characters does it with the corpus green; anything demonstrably better is
   fine. Publish a case for the drive-letter boundary — the arm currently has none, which is
   why the corpus stayed green under my repair.
2. **F6** — decide `file:` from §1 and act on it. Either drop it from the allowlist and
   delete/invert `valid-overlay-path-file-url`, or say which sentence of §1 makes a URL a
   `path` operand. The schema may not settle it by allowlist.
3. **F7** — decide from core §6.1 what an invalid SCP form does, and publish the case that
   kills **both** M-host and M-host2. "Invalid network forms MUST be rejected, not treated as
   local" is the sentence to answer.
4. **F8** — one negative case for `directory` on a `path` source; re-run M-dir and report it
   dead.
5. **F9** — state `C:foo` and M-unanchored as bounds; reflow `profiles/manager.md:2197`.
6. Re-run both gates and requote the exit codes; re-read `gh pr checks 47` after the push.

Reviewer artifacts (in the story workspace, nothing written outside `.temp/`):
`.temp/review-2x4s7i/{matrix.py,edges.py,mutants.py,mutants2.py,repair.py,re2/main.go,
matrix-committed.md,edges-committed.md,mutants.md,mutants2.md,repair.md,validate.log,regen.log}`.
