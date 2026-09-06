# Mutant sweep — cycle 8, at the landed commit `87a0d00`

## Gate corpus and harness validity, established before any result counts

The gate corpus is every case whose verdict the overlay definition can move: the **71**
`manager-config-v2` schema cases named by `conformance/v1/schema-cases/index.json` plus the **41**
schema-2 vectors in `conformance/v1/vectors/manager-config-v2.json` — **112** entries. A mutant is
*killed* when at least one named entry stops agreeing with its recorded `valid` flag.

Three controls, all passing, before the sweep is trusted:

1. **Unmutated corpus**: 112 entries, **0 disagreements**. The harness reproduces the recorded
   verdicts exactly.
2. **`C-noop`** (deep copy, no edit) **survives at 0 flips** — the harness does not manufacture kills.
3. **`C-reorder-allof`** permutes the three `allOf` arms. `allOf` is order-free by construction, so
   it must survive; it does, at **0 flips over the published corpus and 0 flips over a separate
   412-spelling population** (section 3). A control that only survives on the small corpus would
   prove nothing.

Every non-control mutant is additionally asserted to (a) change the document and (b) remain a legal
Draft 2020-12 schema. **0 harness-rejected.**

## Results

| mutant | verdict | flips | first named case that flips |
|---|---|---:|---|
| `C-noop` | SURVIVED | 0 | — |
| `C-reorder-allof` | SURVIVED | 0 | — |
| `M-colonish-drop-s` | SURVIVED | 0 | — |
| `M-scp-drop-s-arm2` | SURVIVED | 0 | — |
| `M-scp-drop-s-arm3` | SURVIVED | 0 | — |
| `M-colonish-widen-s` | SURVIVED | 0 | — |
| `M-scheme-del` | KILLED | 4 | `manager-config-v2/invalid-overlay-file-url.json` |
| `M-scheme-star` | KILLED | 4 | `manager-config-v2/invalid-overlay-file-url.json` |
| `M-scheme-add-file` | KILLED | 2 | `manager-config-v2/invalid-overlay-file-url-with-form.json` |
| `M-scheme-drop-git` | KILLED | 2 | `manager-config-v2/valid-overlay-git-git-uppercase.json` |
| `M-scheme-case` | KILLED | 8 | `manager-config-v2/valid-overlay-git-git-uppercase.json` |
| `M-scheme-require-s` | KILLED | 2 | `manager-config-v2/valid-overlay-git-http-uppercase.json` |
| `M-scheme-then-true` | KILLED | 4 | `manager-config-v2/invalid-overlay-file-url.json` |
| `M-arm2-del` | KILLED | 7 | `manager-config-v2/invalid-overlay-bare-drive-letter.json` |
| `M-arm2-then-true` | KILLED | 7 | `manager-config-v2/invalid-overlay-bare-drive-letter.json` |
| `M-drive-del2` | KILLED | 8 | `manager-config-v2/valid-overlay-path-windows-backslash.json` |
| `M-drive-wide` | KILLED | 2 | `manager-config-v2/invalid-overlay-bare-drive-letter.json` |
| `M-drive-upper` | KILLED | 2 | `manager-config-v2/valid-overlay-path-windows-lowercase-drive.json` |
| `M-drive-multi` | SURVIVED | 0 | — |
| `M-drive-slashonly` | KILLED | 4 | `manager-config-v2/valid-overlay-path-windows-backslash.json` |
| `M-host-wide` | KILLED | 3 | `manager-config-v2/invalid-overlay-scp-host-grammar-no-user.json` |
| `M-host-underscore` | KILLED | 3 | `manager-config-v2/invalid-overlay-scp-host-grammar-no-user.json` |
| `M-host-two` | KILLED | 2 | `manager-config-v2/valid-overlay-git-single-letter-host.json` |
| `M-host-nodigit` | SURVIVED | 0 | — |
| `M-scp-bslash` | KILLED | 2 | `manager-config-v2/invalid-overlay-scp-backslash-path.json` |
| `M-scp-slash` | SURVIVED | 0 | — |
| `M-scp-emptytail` | KILLED | 4 | `manager-config-v2/invalid-overlay-bare-drive-letter.json` |
| `M-scp-requser` | KILLED | 4 | `manager-config-v2/valid-overlay-git-scp-no-user.json` |
| `M-arm2-noanchor` | KILLED | 2 | `manager-config-v2/valid-overlay-path-colon-later-segment.json` |
| `M-arm3-anyof` | KILLED | 2 | `manager-config-v2/invalid-overlay-two-requirement-forms.json` |
| `M-else-drop-dir` | KILLED | 2 | `manager-config-v2/invalid-overlay-path-directory.json` |
| `M-else-drop-rev` | KILLED | 4 | `manager-config-v2/invalid-overlay-path-requirement-form.json` |
| `M-else-drop-tag` | KILLED | 2 | `manager-config-v2/invalid-overlay-path-relative-requirement-form.json` |
| `M-else-drop-range` | KILLED | 2 | `manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json` |
| `M-then-drop-range` | KILLED | 12 | `manager-config-v2/invalid-overlay-two-requirement-forms.json` |
| `M-then-drop-tag` | KILLED | 8 | `manager-config-v2/invalid-overlay-two-requirement-forms.json` |
| `M-arm3-swap` | KILLED | 43 | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` |
| `M-arm3-scp-only` | KILLED | 13 | `manager-config-v2/invalid-overlay-git-uppercase-no-requirement-form.json` |
| `M-arm3-scheme-only` | KILLED | 8 | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` |
| `M-arm3-del` | KILLED | 17 | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` |
| `M-addprops-open` | KILLED | 1 | `manager-config-v2/invalid-unknown-overlay-field.json` |
| `M-drop-required-source` | SURVIVED | 0 | — |

total 42 | killed 32 | survived 10 | harness-rejected 0
survivors: ['C-noop', 'C-reorder-allof', 'M-colonish-drop-s', 'M-scp-drop-s-arm2', 'M-scp-drop-s-arm3', 'M-colonish-widen-s', 'M-drive-multi', 'M-host-nodigit', 'M-scp-slash', 'M-drop-required-source']

**32 of 42 killed on a named case.** Only 6 of the 42 are deletions; the rest are boundary
narrowings and widenings. Two of the ten survivors are the intentional controls.

## Survivors, quantified rather than called no-ops

Each survivor was run over an independent 412-spelling population built from userinfo x host x tail
x scheme families plus drive, POSIX and whitespace-class spellings, and the classification diffed
against the committed schema.

population: 412
baseline kinds: {'path': 181, 'REFUSED': 197, 'git': 34}

| survivor | flips over 412 spellings | directions | example |
|---|---:|---|---|
| `C-reorder-allof` | 0 | — | `—` |
| `M-colonish-drop-s` | 160 | {'path->REFUSED': 160} | `ab\t:///x` |
| `M-scp-drop-s-arm2` | 64 | {'REFUSED->path': 64} | `1:\t///x` |
| `M-scp-drop-s-arm3` | 0 | — | `—` |
| `M-colonish-widen-s` | 0 | — | `—` |
| `M-drive-multi` | 2 | {'REFUSED->path': 2} | `host:/x` |
| `M-host-nodigit` | 6 | {'git->REFUSED': 6} | `9host:sub/dir/x` |
| `M-scp-slash` | 12 | {'REFUSED->path': 12} | `1:///x` |
| `M-drop-required-source` | 0 | — | `—` |

Reading:

- **`C-reorder-allof`** — 0 flips. The control behaves as a control must.
- **`M-colonish-drop-s`** (`^[^/\s]*:` -> `^[^/]*:`) — **0 flips on the published corpus, 160 over
  the population**, every one `path -> REFUSED`. That is the size of the whitespace bypass class:
  160 spellings that the committed schema admits as form-free paths and that an engine-independent
  colon test would refuse. No published case can see the difference.
- **`M-scp-drop-s-arm2`** — 64 flips `REFUSED -> path`, the same class from the other pattern.
- **`M-scp-drop-s-arm3`**, **`M-colonish-widen-s`** — 0 flips over the population; measured
  semantic no-ops *in this population*, not asserted ones.
- **`M-drive-multi`** (2 flips, `host:/x`), **`M-host-nodigit`** (6, `9host:x`), **`M-scp-slash`**
  (12, `1:///x` and kin) — three unpinned degenerate classes. The first two reproduce cycle 7's F27
  independently; `M-scp-slash` is new here.
- **`M-drop-required-source`** — 0 flips over the population because the population never omits
  `source`. Driven separately: with `required: ["source"]` removed, `{"revision": "a"*40}` and
  `{"tag": "v1"}` become valid, because all three arms condition on `properties.source.pattern`,
  which passes vacuously when `source` is absent. **This is not a regression**: the pre-branch
  definition at `550579d` had the same vacuity (its unconditional `oneOf` was likewise satisfied by
  the bare form). Reproduced from `git show 550579d:schemas/v1/manager-config-v2.schema.json`.

## What the kills prove

Every arm carries at least one killing case, and the kills are on *narrowings*, not only deletions:

| Class | Mutant | Named case that flips |
|---|---|---|
| arm 1 scheme allowlist, widened | `M-scheme-add-file` | `invalid-overlay-file-url-with-form.json` |
| arm 1 scheme allowlist, narrowed | `M-scheme-drop-git` | `valid-overlay-git-git-uppercase.json` |
| arm 1 case folding | `M-scheme-case` | `valid-overlay-git-git-uppercase.json` |
| arm 1 scheme length `+`->`*` | `M-scheme-star` | `invalid-overlay-file-url.json` |
| arm 2 drive carve-out, removed | `M-drive-del2` | `valid-overlay-path-windows-backslash.json` |
| arm 2 drive carve-out, widened | `M-drive-wide` | `invalid-overlay-bare-drive-letter.json` |
| arm 2 drive carve-out, narrowed | `M-drive-upper` | `valid-overlay-path-windows-lowercase-drive.json` |
| arm 2 host grammar, widened | `M-host-wide` | `invalid-overlay-scp-host-grammar-no-user.json` |
| arm 2 host grammar, narrowed | `M-host-two` | `valid-overlay-git-single-letter-host.json` |
| arm 2 SCP tail | `M-scp-bslash` | `invalid-overlay-scp-backslash-path.json` |
| arm 2 anchoring | `M-arm2-noanchor` | `valid-overlay-path-colon-later-segment.json` |
| arm 3 exclusivity | `M-arm3-anyof` | `invalid-overlay-two-requirement-forms.json` |
| arm 3 path refusal set | `M-else-drop-dir` | `invalid-overlay-path-directory.json` |
| closed object | `M-addprops-open` | `invalid-unknown-overlay-field.json` |

The **arm-2 drive carve-out is live, not dead**: `M-drive-del2` kills on a named case with 8 flips.

## The one class the corpus cannot see

Four mutants that materially change behaviour produce **zero** flips on the published corpus:
`M-colonish-drop-s`, `M-scp-drop-s-arm2`, `M-scp-drop-s-arm3`, `M-colonish-widen-s`. All four edit
the `\s` character class. Measured cause: of the 162 overlay `source` declarations across the cases
and vectors, **0** contain any whitespace character and **0** contain any non-ASCII character.
