# Mutant sweep — cycle 6, against the committed corpus at 87a0d00

Each mutant patches `schemas/v1/manager-config-v2.schema.json` **structurally** (parse,
mutate the named node, re-serialise) in a throwaway copy of the landed tree, then runs the
repository's own `tools/validate.py` over the entire committed corpus. A mutant is *killed*
only when validate.py names the case that failed.

## Harness validity, established before any result was believed

| control | result |
|---|---|
| unmutated schema over the corpus | **exit 0** (so the corpus is green to begin with) |
| `CTRL-noop-desc` — adds a `description` key, changes no behaviour | **SURVIVES** (so the harness is not killing everything) |
| every mutant asserted to actually change the serialised document | no `NOT-APPLIED` rows |
| `conformance/v1/manifest.json` digest coverage of `schemas/v1` | **none** — 0 manifest paths under `schemas/` |

That last row matters: if the manifest digested the schema file, every mutant would "die"
on a digest mismatch and the whole sweep would be a false positive. It does not.

## Result: 34 of 40 killed on a named case

Only 9 of the 34 kills are deletions. The other 25 move a boundary — a character class
narrowed, a quantifier changed, one member removed from a refusal set, an arm's host grammar
widened — which is the shape that says something about the class rather than about existence.

```
  M-arm1-del                 killed      manager-config-v2/invalid-overlay-unknown-scheme.json
  M-arm2-del                 killed      manager-config-v2/invalid-overlay-scp-host-grammar.json
  M-arm3-del                 killed      manager-config-v2/invalid-overlay-two-requirement-forms.json
  M-addprops                 killed      manager-config-v2/invalid-unknown-overlay-field.json
  M-range                    killed      manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json
  M-tag                      killed      manager-config-v2/invalid-overlay-path-relative-requirement-form.json
  M-revision                 killed      manager-config-v2/invalid-overlay-path-requirement-form.json
  M-dir                      killed      manager-config-v2/invalid-overlay-path-directory.json
  M-oneof-any                killed      manager-config-v2/invalid-overlay-two-requirement-forms.json
  M-then-true                killed      manager-config-v2/invalid-overlay-two-requirement-forms.json
  M-swap                     killed      manager-config-v2/valid.json
  M-file                     killed      manager-config-v2/invalid-overlay-file-url.json
  M-schemelen                killed      manager-config-v2/valid-overlay-path-windows-double-slash.json
  M-scheme-case              killed      manager-config-v2/valid-overlay-git-ssh-uppercase.json
  M-drop-git                 killed      manager-config-v2/valid-overlay-git-git-uppercase.json
  M-require-s                killed      manager-config-v2/valid-overlay-git-http-uppercase.json
  M-arm1-scheme-noalpha      SURVIVES    
  M-host                     killed      manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json
  M-host2                    killed      manager-config-v2/valid-overlay-git-single-letter-host.json
  M-drive-del2               killed      manager-config-v2/valid-overlay-path-windows-backslash.json
  M-drive-wide               killed      manager-config-v2/invalid-overlay-bare-drive-letter.json
  M-drive-anycase-off        killed      manager-config-v2/valid-overlay-path-windows-lowercase-drive.json
  M-drive-multiletter        SURVIVES    
  M-drive-slashonly          killed      manager-config-v2/valid-overlay-path-windows-backslash.json
  M-drive-bslashonly         killed      manager-config-v2/valid-overlay-path-windows-slash.json
  M-arm2-colonplus           SURVIVES    
  M-unanchored1              killed      manager-config-v2/valid-overlay-path-colon-later-segment.json
  M-scp-bslash-arm2          killed      manager-config-v2/invalid-overlay-scp-backslash-path.json
  M-scp-leadslash            killed      manager-config-v2/valid-overlay-path-windows-slash.json
  M-scp-user-required        killed      manager-config-v2/valid-overlay-git-single-letter-host.json
  M-scp-nouser               killed      manager-config-v2/valid-overlay-git-scp.json
  M-scp-bslash-anywhere      killed      manager-config-v2/valid-overlay-git-single-letter-host.json
  M-arm2-scp-only            killed      manager-config-v2/valid-overlay-git-single-letter-host.json
  M-arm2-schemeguard         killed      manager-config-v2/valid.json
  M-a3-url-only              killed      manager-config-v2/valid-overlay-git-single-letter-host.json
  M-a3-scp-only              killed      manager-config-v2/valid.json
  M-a3-slash                 killed      manager-config-v2/valid.json
  M-a3-host-permissive       SURVIVES    
  M-a3-drive-restore         SURVIVES    
  CTRL-noop-desc             SURVIVES    
```

## The six survivors, characterised by measurement rather than argument

Flip sets measured by classifying all 2,779 corpus spellings under the committed schema and
under the mutant, then diffing.

| mutant | flips | class | verdict |
|---|---:|---|---|
| `CTRL-noop-desc` | — | — | intentional control, must survive |
| `M-arm1-scheme-noalpha` (arm 1 scheme may start with a digit) | **0** | — | semantic no-op: a digit-leading `://` spelling is already refused by arm 2 |
| `M-a3-host-permissive` (arm 3 host class permissive, arm 2 untouched) | **0** | — | semantic no-op: arm 2 refuses everything arm 3 would newly admit |
| `M-a3-drive-restore` (re-add the arm-3 drive carve-out removed at 18dca85) | **0** | — | semantic no-op, confirming cycle 3's analytic proof by measurement |
| `M-drive-multiletter` (`^[A-Za-z]:[\\/]` → `^[A-Za-z]+:[\\/]`) | 20 | `refused` → `path`, an all-alpha multi-letter prefix before a separator (`HOST:/x`, `abc:/x`) | real unpinned decision, degenerate class |
| `M-arm2-colonplus` (`^[^/\s]*:` → `^[^/\s]+:`) | 2 | `refused` → `path`, a source whose first character is a colon (`:`, `:x`) | real unpinned decision, degenerate class |

Three of the five non-control survivors are 0-flip no-ops. The two real ones flip only
spellings no operator types.

## The kills that carry the most weight

| mutant | the case that catches it | what it proves |
|---|---|---|
| `M-drive-anycase-off` (drive carve-out uppercase only) | `valid-overlay-path-windows-lowercase-drive.json` | **cycle 5's F19 is closed.** This mutant SURVIVED at `407424e` with 81 flips, all `path` → `refused`, the class being `c:\users\operator\context` — i.e. it re-created the cycle-1 defect. It dies at the landed head. |
| `M-drive-wide` (`^[A-Za-z]:[\\/]` → `^[A-Za-z]:`) | `invalid-overlay-bare-drive-letter.json` | the bare drive letter stays refused |
| `M-host` (6.1 host grammar → permissive) | `invalid-overlay-scp-host-grammar-with-form.json` | the host grammar is pinned from above |
| `M-host2` (host must be ≥2 chars) | `valid-overlay-git-single-letter-host.json` | and from below — both directions |
| `M-scheme-case` (allowlist becomes case-sensitive) | `valid-overlay-git-ssh-uppercase.json` | the four uppercase positives are load-bearing, not decoration |
| `M-drop-git` / `M-require-s` | `valid-overlay-git-git-uppercase.json` / `valid-overlay-git-http-uppercase.json` | allowlist membership is pinned scheme by scheme |
| `M-range` / `M-tag` / `M-revision` / `M-dir` | four *distinct* named path cases | each of the four section-1-forbidden members is refused by its own case, not by one case standing in for all |
| `M-file` (re-admit `file:`) | `invalid-overlay-file-url.json` | the `file:` refusal is pinned |
| `M-scp-bslash-arm2` | `invalid-overlay-scp-backslash-path.json` | the backslash-after-colon refusal is pinned |
| `M-scp-user-required` / `M-scp-nouser` | `valid-overlay-git-single-letter-host.json` / `valid-overlay-git-scp.json` | the optional `[user@]` is pinned in both directions |
| `M-swap` (swap the then/else arms) | `valid.json` | the discriminator's orientation is pinned by the baseline document itself |
| `M-a3-slash` (cycle 1's surviving `^/` discriminator) | `valid.json` | the mutant that survived the cycle-1 corpus is dead |

`M-scheme-case`, `M-drop-git`, `M-require-s`, `M-drive-slashonly`, `M-drive-bslashonly`,
`M-scp-user-required`, `M-scp-nouser`, `M-scp-leadslash`, `M-arm1-scheme-noalpha`,
`M-a3-host-permissive` and `M-arm2-schemeguard` are mutants no earlier cycle ran.
All but two of them die.
