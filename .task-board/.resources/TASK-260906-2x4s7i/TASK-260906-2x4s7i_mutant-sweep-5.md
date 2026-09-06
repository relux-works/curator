# Mutant sweep — cycle 5, PR #47 at `407424e`

45 mutants plus a control, each run as a standalone `python3 tools/validate.py` process against the
committed corpus on an `rsync` copy of the clean `407424e` worktree (`jsonschema==4.25.1`, the
`requirements-dev.txt` pin). The schema is restored byte-exactly after every run and its SHA-256
re-asserted: `d93a6673a989bc0dadc3029db4cbfa5c83c5e8037f402c51c459ebc5b6bef64d` — **the same digest
cycle 4 recorded at `18dca85`**, which is the first proof that this delta changed no behaviour.

Control `M0-control`: exit 0 — `validated 60 schemas and 1046 vector files`.

**37 killed on a named case, 8 survivors, of which 2 are measured 0-flip semantic no-ops.**

## The mutant this cycle exists to kill

| mutant | at `18dca85` (cycle 4) | at `407424e` |
|---|---|---|
| `M-drive-wide` — arm 2's carve-out `^[A-Za-z]:[\\/]` widened to `^[A-Za-z]:` | **SURVIVED** | **KILLED** on `invalid-overlay-bare-drive-letter.json` |

The new case pins arm 2 specifically, not some incidental refusal. `C:` fails at schema path
`…/allOf/1` (`False schema does not allow {'source': 'C:'}`), and it becomes **valid** under both
`M-arm2-del` and `M-drive-wide` — so the case is load-bearing for exactly the clause it names.

## Killed (37)

| mutant | named case that fails |
|---|---|
| M-drive-wide  arm 2 carve-out → `^[A-Za-z]:` | **`invalid-overlay-bare-drive-letter.json`** (new) |
| M-drive-del2  drop arm 2's carve-out | `valid-overlay-path-windows-backslash.json` |
| M-drive-noslash  carve-out `[\\/]` → `[/]` | `valid-overlay-path-windows-backslash.json` |
| M-drive-nobslash  carve-out `[\\/]` → `[\\]` | `valid-overlay-path-windows-slash.json` |
| M-drive-single  carve-out demands a non-slash after the separator | `valid-overlay-path-windows-double-slash.json` |
| M-unanchored2  arm 2's colon test loses `^` | `valid-overlay-path-colon-later-segment.json` |
| M-arm2-colon-anychar  arm 2's colon test admits `/` before the colon | `valid-overlay-path-colon-later-segment.json` |
| M-arm2-del  drop the invalid-SCP arm | `invalid-overlay-scp-host-grammar.json` |
| M-arm2-noscp  drop "not a valid SCP form" from arm 2 | `valid-overlay-git-single-letter-host.json` |
| M-arm2-noscheme  drop "not scheme-shaped" from arm 2 | `valid.json` |
| M-scp-bslash-arm2  backslash re-admitted in arm 2's `not` only | `invalid-overlay-scp-backslash-path.json` |
| M-scp-bslash-arm3  backslash re-admitted in arm 3 only | `valid-overlay-path-windows-backslash.json` |
| M-scp-bslash  backslash re-admitted at both sites | `valid-overlay-path-windows-backslash.json` |
| M-host  SCP host class → `[^/:\s]+` (the cycle-1 permissive class) | `invalid-overlay-scp-host-grammar-with-form.json` |
| M-host2  SCP host must be ≥ 2 characters | `valid-overlay-git-single-letter-host.json` |
| M-hostcls  SCP host class loses `.` and `-` | `valid-overlay-git-scp.json` |
| M-unanchored-scp  SCP pattern loses `^`, both sites | `valid-overlay-path-colon-later-segment.json` |
| M-scp-nouser  SCP pattern drops the optional user part | `valid-overlay-git-scp.json` |
| M-scp-slashpath  SCP path admits a leading `/` | `valid-overlay-path-windows-slash.json` |
| M-file  re-admit `file` to the scheme allowlist | `invalid-overlay-file-url-with-form.json` |
| M-schemelen  scheme length `+` → `*`, both sites | `valid-overlay-path-windows-double-slash.json` |
| M-schemelen-arm1  scheme length `+` → `*`, arm 1 only | `valid-overlay-path-windows-double-slash.json` |
| M-lower  allowlist lowercase only | `valid-overlay-git-ssh-uppercase.json` |
| M-arm1-del  drop the unknown-scheme arm | `invalid-overlay-unknown-scheme.json` |
| M-oneof  git `then` `oneOf` → `anyOf` | `invalid-overlay-two-requirement-forms.json` |
| M-notallof  path `not/anyOf` → `not/allOf` | `invalid-overlay-path-requirement-form.json` |
| M-dir  drop `directory` from the path refusal | `invalid-overlay-path-directory.json` |
| M-range  drop `range` from the path refusal | `invalid-overlay-path-windows-slash-requirement-form.json` |
| M-tag  drop `tag` from the path refusal | `invalid-overlay-path-relative-requirement-form.json` |
| M-revision  drop `revision` from the path refusal | `invalid-overlay-path-requirement-form.json` |
| M-arm3-noscp  arm 3's `anyOf` drops the SCP alternative | `valid-overlay-git-single-letter-host.json` |
| M-arm3-noallow  arm 3's `anyOf` drops the scheme alternative | `valid.json` |
| M-arm3-allof  arm 3's `anyOf` → `allOf` | `valid.json` |
| M-arm3-del  drop the kind arm entirely | `invalid-overlay-two-requirement-forms.json` |
| M-branch-prop  add `branch` to the overlay `properties` | `invalid-unknown-overlay-field.json` |
| M-addprops  `additionalProperties` `false` → `true` | `invalid-unknown-overlay-field.json` |
| M-branch-both  add `branch` **and** open the object | `invalid-unknown-overlay-field.json` |

Only 8 of the 37 are deletions; the rest narrow, widen, unanchor, or re-spell a character class.

## Survivors (8)

Flip sets computed by classifying 1,694 spellings under the committed and the mutated schema and
diffing the three-way classification (`path` / `git` / `refused`).

### Measured semantic no-ops (2)

| survivor | flips |
|---|---|
| M-drive-restore3  re-add the carve-out F13 removed from arm 3 | **0** — confirms cycle 4's analytic disjointness proof empirically at this head |
| M-reorder  arms permuted 3-2-1 | **0** — `allOf` is order-free and each arm's precondition is an explicit `not` |

### Real unpinned decisions (6)

| survivor | flips | direction | the class it leaves unpinned |
|---|---|---|---|
| **M-drive-anycase-off**  carve-out `^[A-Za-z]:` → `^[A-Z]:` | 64 | `path` → `refused` | **a lowercase Windows drive letter**: `c:\users\operator\context`, `c:/users/x`, `d:/ctx`, `z:\x`. All three published Windows positives spell the drive `C`. **New this cycle.** |
| M-drive-multiletter  carve-out `^[A-Za-z]+:[\\/]` | 109 | `refused` → `path` | an alphabetic multi-letter prefix followed by a separator: `abc:/x`, `Users:\x`, `file:/x`. **New this cycle.** |
| M-scp-bslash-anywhere  SCP path forbids a backslash anywhere, not only at the first character | 35 | `git` → `refused` | a backslash later in an SCP repository path: `github.com:example\x`, `git@github.com:example/x\y`. **New this cycle.** |
| M-arm2-colonplus  arm 2's colon test `*` → `+` | 44 | `refused` → `path` | leading-colon spellings: `:`, `:x`, `://x` |
| M-unanchored1  arm 1's scheme test loses `^` | 2 | `path` → `refused` | a project-relative spelling with a literal `://` in a later segment: `ctx/http://y`, `a/b/https://c` |
| M-unanchored-allow  arm 3's allowlist loses `^` | 2 | `path` → `git` | the same two spellings, toward `git` |

The last three are cycle 3's survivors, unchanged. None of the six admits a form-carrying `path` or
lets a `git` source skip its form.
