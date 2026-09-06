# Mutant sweep — cycle 4, PR #47 at `18dca85`

41 mutants plus a control, each run as a standalone `python3 tools/validate.py` process against the
committed corpus on an `rsync` copy of the clean `18dca85` worktree (`jsonschema==4.25.1`, the
`requirements-dev.txt` pin). The schema is restored byte-exactly after every run and its SHA-256
asserted (`d93a6673…f64d`). `git archive` is not usable for this repository — it breaks
`conformance/v1/fixtures/byte-exact/subst.txt` — so the tree was copied with `rsync`.

**32 killed on a named case. 9 survivors, of which 3 are measured semantic no-ops, leaving 4 distinct
unpinned decisions.**

Control: `M0-control` exit 0 — `validated 60 schemas and 1045 vector files`.

## Killed

| mutant | named case that fails |
|---|---|
| M-drive-del2  drop arm 2's Windows-drive carve-out | `valid-overlay-path-windows-backslash.json` |
| M-drive-noslash  carve-out `[\\/]` → `[/]` | `valid-overlay-path-windows-backslash.json` |
| M-drive-nobslash  carve-out `[\\/]` → `[\\]` | `valid-overlay-path-windows-slash.json` |
| M-drive-single  carve-out demands a non-slash after the separator | `valid-overlay-path-windows-double-slash.json` |
| **M-unanchored2  arm 2's colon test loses `^`** | **`valid-overlay-path-colon-later-segment.json`** (new) |
| **M-scp-bslash-arm2  backslash re-admitted in arm 2's `not` only** | **`invalid-overlay-scp-backslash-path.json`** (new) |
| **M-unanchored-scp  SCP pattern loses `^`, both sites** | **`valid-overlay-path-colon-later-segment.json`** (new) |
| **M-arm2-colon-anychar  arm 2's colon test admits `/` before the colon** | **`valid-overlay-path-colon-later-segment.json`** (new) |
| M-scp-bslash  SCP class re-admits backslash, both sites | `valid-overlay-path-windows-backslash.json` |
| M-scp-bslash-arm3  backslash re-admitted in arm 3 only | `valid-overlay-path-windows-backslash.json` |
| M-host  SCP host class → `[^/:\s]+` (the cycle-1 permissive class) | `invalid-overlay-scp-host-grammar-with-form.json` |
| M-host2  SCP host must be ≥ 2 characters | `valid-overlay-git-single-letter-host.json` |
| M-hostcls  SCP host class loses `.` and `-` | `valid-overlay-git-scp.json` |
| M-dir  drop `directory` from the path refusal | `invalid-overlay-path-directory.json` |
| M-range  drop `range` from the path refusal | `invalid-overlay-path-windows-slash-requirement-form.json` |
| M-tag  drop `tag` from the path refusal | `invalid-overlay-path-relative-requirement-form.json` |
| M-revision  drop `revision` from the path refusal | `invalid-overlay-path-requirement-form.json` |
| M-file  re-admit `file` to the scheme allowlist | `invalid-overlay-file-url-with-form.json` |
| M-schemelen  arm 1 scheme length `+` → `*` | `valid-overlay-path-windows-double-slash.json` |
| M-lower  allowlist lowercase only | `valid-overlay-git-ssh-uppercase.json` |
| M-arm1-del  drop the unknown-scheme arm | `invalid-overlay-unknown-scheme.json` |
| M-arm2-del  drop the invalid-SCP arm | `invalid-overlay-scp-host-grammar.json` |
| M-arm2-noscp  drop "not a valid SCP form" from arm 2 | `valid-overlay-git-single-letter-host.json` |
| M-arm2-noscheme  drop "not scheme-shaped" from arm 2 | `valid.json` |
| M-oneof  git `then` `oneOf` → `anyOf` | `invalid-overlay-two-requirement-forms.json` |
| M-notallof  path `not/anyOf` → `not/allOf` | `invalid-overlay-path-requirement-form.json` |
| M-arm3-noscp  arm 3's `anyOf` drops the SCP alternative | `valid-overlay-git-single-letter-host.json` |
| M-arm3-noallow  arm 3's `anyOf` drops the scheme alternative | `valid.json` |
| M-arm3-allof  arm 3's `anyOf` → `allOf` | `valid.json` |
| M-branch-prop  add `branch` to the overlay `properties` | `invalid-unknown-overlay-field.json` |
| M-addprops  `additionalProperties` `false` → `true` | `invalid-unknown-overlay-field.json` |
| M-branch-both  add `branch` **and** open the object | `invalid-unknown-overlay-field.json` |

The four bold rows are the load-bearing evidence for the two cases this commit adds: without
`valid-overlay-path-colon-later-segment.json` three of them go green, and without
`invalid-overlay-scp-backslash-path.json` the fourth does.

## Survivors

Three are measured semantic no-ops, not gaps:

| survivor | why it changes nothing |
|---|---|
| M-drive-restore3  re-add the carve-out F13 removed from arm 3 | 0 classification differences over 16,159 spellings; see the analytic proof in the verdict |
| M-reorder  arms permuted 3-2-1 | `allOf` is order-free and each arm's precondition is an explicit `not`, not a fall-through |
| M-unanchored-allow (arm 1's `not` only) | 0 flips over 16,171 probes — arm 1's `not` is only reached when the anchored scheme test already matched |

Four are real unpinned decisions. Flips measured by classifying 16,171 spellings under the committed
and the mutated schema and diffing:

| survivor | flips | direction | the class it leaves unpinned |
|---|---|---|---|
| **M-drive-wide**  arm 2's carve-out → `^[A-Za-z]:` | 42 | `refused` → `path` | a bare drive letter: `C:`, `C: ` — degenerate. **New this cycle**: this mutant died at `2f2dfa4` on `valid-overlay-git-single-letter-host.json` and survives at `18dca85` because F13 removed arm 3's copy of the carve-out |
| M-unanchored1  arm 1's scheme test loses `^` (arm-1-only and both-sites are identical) | 8 | `path` → `refused`, one `git` → `refused` | a project-relative spelling containing a literal `://` in a later segment: `ctx/http://y`, `context/http-mirror://x`. Every member of this class contains `//`, i.e. a redundant separator |
| M-unanchored-allow  arm 3's allowlist loses `^` (arm-3-only and both-sites are identical) | 3 | `path` → `git` | `a/b/https://c` — a project-relative path whose later segment is a scheme spelling |
| M-arm2-colonplus  arm 2's colon test `*` → `+` | 1175 | `refused` → `path` | leading-colon spellings: `:`, `:x`, `:/`, … |

None of the four admits a form-carrying `path` or lets a `git` source skip its form. Three fail toward
refusal (loud) and one toward a `path` that resolution then reports as `profile_source_path_missing`.
