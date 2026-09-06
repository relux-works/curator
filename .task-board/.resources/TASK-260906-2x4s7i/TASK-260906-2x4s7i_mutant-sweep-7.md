# Cycle-7 mutant sweep (87a0d00)

40 mutants + harness falsification against the gate corpus (71 manager-config-v2 schema
cases + 42 v2 vectors). Unmutated corpus = 0 flips; every mutant asserted to change the
document and remain a legal Draft 2020-12 schema; intentional control must survive.

gate corpus: 71 schema cases + 42 v2 vectors
HARNESS CHECK 1  unmutated flips: 0 []

| mutant | verdict | first named case that flips | flips |
|---|---|---|---:|
| `M-control-reorder` | SURVIVED | `-` | 0 |
| `M-scheme-star` | KILLED | `manager-config-v2/valid-overlay-path-windows-double-slash.json` | 2 |
| `M-scheme-del` | KILLED | `manager-config-v2/invalid-overlay-file-url.json` | 4 |
| `M-scheme-add-file` | KILLED | `manager-config-v2/invalid-overlay-file-url-with-form.json` | 2 |
| `M-scheme-drop-git` | KILLED | `manager-config-v2/valid-overlay-git-git-uppercase.json` | 2 |
| `M-scheme-drop-ssh` | KILLED | `manager-config-v2/valid-overlay-git-ssh-uppercase.json` | 2 |
| `M-scheme-require-s` | KILLED | `manager-config-v2/valid-overlay-git-http-uppercase.json` | 2 |
| `M-scheme-case` | KILLED | `manager-config-v2/valid-overlay-git-git-uppercase.json` | 8 |
| `M-scheme-noanchor` | SURVIVED | `-` | 0 |
| `M-arm2-del` | KILLED | `manager-config-v2/invalid-overlay-bare-drive-letter.json` | 7 |
| `M-arm2-colonplus` | SURVIVED | `-` | 0 |
| `M-arm2-noanchor` | KILLED | `manager-config-v2/valid-overlay-path-colon-later-segment.json` | 2 |
| `M-drive-del2` | KILLED | `manager-config-v2/valid-overlay-path-windows-backslash.json` | 8 |
| `M-drive-wide` | KILLED | `manager-config-v2/invalid-overlay-bare-drive-letter.json` | 2 |
| `M-drive-multi` | SURVIVED | `-` | 0 |
| `M-drive-upper` | KILLED | `manager-config-v2/valid-overlay-path-windows-lowercase-drive.json` | 2 |
| `M-drive-lower` | KILLED | `manager-config-v2/valid-overlay-path-windows-backslash.json` | 6 |
| `M-drive-slashonly` | KILLED | `manager-config-v2/valid-overlay-path-windows-backslash.json` | 4 |
| `M-host-wide` | KILLED | `manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json` | 2 |
| `M-host-two` | KILLED | `manager-config-v2/valid-overlay-git-single-letter-host.json` | 2 |
| `M-host-underscore` | KILLED | `manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json` | 2 |
| `M-host-nodigit` | SURVIVED | `-` | 0 |
| `M-scp-bslash` | KILLED | `manager-config-v2/invalid-overlay-path-windows-requirement-form.json` | 6 |
| `M-scp-slash` | KILLED | `manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json` | 6 |
| `M-scp-nouser` | KILLED | `manager-config-v2/valid-overlay-git-scp.json` | 2 |
| `M-scp-useratsign` | SURVIVED | `-` | 0 |
| `M-arm3-del` | KILLED | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` | 17 |
| `M-arm3-anyof` | KILLED | `manager-config-v2/invalid-overlay-two-requirement-forms.json` | 2 |
| `M-arm3-drop-range` | KILLED | `manager-config-v2/invalid-overlay-two-requirement-forms.json` | 12 |
| `M-arm3-drop-tag` | KILLED | `manager-config-v2/invalid-overlay-two-requirement-forms.json` | 8 |
| `M-arm3-drop-rev` | KILLED | `manager-config-v2/valid-overlay-git-https-uppercase.json` | 2 |
| `M-else-drop-range` | KILLED | `manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json` | 2 |
| `M-else-drop-tag` | KILLED | `manager-config-v2/invalid-overlay-path-relative-requirement-form.json` | 2 |
| `M-else-drop-rev` | KILLED | `manager-config-v2/invalid-overlay-path-requirement-form.json` | 4 |
| `M-else-drop-dir` | KILLED | `manager-config-v2/invalid-overlay-path-directory.json` | 2 |
| `M-arm3-swap` | KILLED | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` | 31 |
| `M-arm3-scheme-only` | KILLED | `manager-config-v2/invalid-overlay-git-scp-no-requirement-form.json` | 8 |
| `M-arm3-scp-only` | KILLED | `manager-config-v2/invalid-overlay-git-uppercase-no-requirement-form.json` | 13 |
| `M-addprops-open` | KILLED | `manager-config-v2/invalid-unknown-overlay-field.json` | 1 |
| `M-source-optional` | SURVIVED | `-` | 0 |

killed 33/40   survived 7   harness-rejected 0
SURVIVORS: ['M-control-reorder', 'M-scheme-noanchor', 'M-arm2-colonplus', 'M-drive-multi', 'M-host-nodigit', 'M-scp-useratsign', 'M-source-optional']
HARNESS-REJECTED: []

## Survivor classification, measured over 2,264 spellings

corpus 2264 spellings

| survivor | classification flips | direction | example class |
|---|---:|---|---|
| `M-control-reorder` | 0 | — | **measured semantic no-op** |
| `M-scheme-noanchor` | 420 | path->REFUSED | '/x/file://-host/', '/x/file://-host/.git', '/x/file://-host//x', '/x/file://-host/a/b' |
| `M-arm2-colonplus` | 66 | REFUSED->path | ':', ':.git', '://-host/', '://-host/.git' |
| `M-drive-multi` | 2 | REFUSED->path | 'HOST:/x', 'host:/x' |
| `M-host-nodigit` | 12 | git->REFUSED | '9host:.git', '9host:a/b', '9host:a\\b', '9host:x' |
| `M-scp-useratsign` | 24 | REFUSED->git | '@9host:.git', '@9host:a/b', '@9host:a\\b', '@9host:x' |

overlay entry {}: valid=False  <- 'source' is a required property

overlay entry {'revision': '0000000000000000000000000000000000000000'}: valid=False  <- 'source' is a required property

overlay entry {'weight': 1}: valid=False  <- 'source' is a required property
  with required[] removed -> valid=False
  with required[] removed -> valid=True
  with required[] removed -> valid=False
