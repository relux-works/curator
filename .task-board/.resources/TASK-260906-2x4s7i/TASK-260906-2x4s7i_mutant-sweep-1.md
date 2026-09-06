| mutant | gate exit | first named case that fails | outcome |
|---|---|---|---|
| M0-control  (no change) | 0 | `validated 60 schemas and 1037 vector files` | green (expected) |
| M1  path else drops `revision` (narrow) | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-path-requirement-form.json against manager-config-v2.schema.json: expected valid=Fals` | killed |
| M2  git then oneOf -> anyOf (narrow) | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-two-requirement-forms.json against manager-config-v2.schema.json: expected valid=Fals` | killed |
| M3  discriminator -> `^/` + arms swapped | 1 | `validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True,` | killed |
| M-old  cycle-1 pattern verbatim (token-preserving) | 1 | `validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True,` | killed |
| M-bs  drop only the backslash exclusion | 1 | `validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True,` | killed |
| M-host  drop the 6.1 host grammar, keep backslash exclusion | 0 | `validated 60 schemas and 1037 vector files` | **SURVIVOR** |
| M-swap  committed pattern, arms swapped | 1 | `validation failed: schema case manager-config-v2/valid.json against manager-config-v2.schema.json: expected valid=True, got {'range': '^1.2', 'source'` | killed |
| M-dir  path else drops `directory` (narrow) | 0 | `validated 60 schemas and 1037 vector files` | **SURVIVOR** |
| M-noarm2  drop the whole second allOf arm | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-unknown-scheme.json against manager-config-v2.schema.json: expected valid=False, got ` | killed |
| M-svn  widen scheme allowlist with svn | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-unknown-scheme.json against manager-config-v2.schema.json: expected valid=False, got ` | killed |
| M-nofile  narrow scheme allowlist: drop file | 1 | `validation failed: schema case manager-config-v2/valid-overlay-path-file-url.json against manager-config-v2.schema.json: expected valid=True, got Fals` | killed |
| M-lower  scheme alternation lowercase only | 1 | `validation failed: schema case manager-config-v2/valid-overlay-git-ssh-uppercase.json against manager-config-v2.schema.json: expected valid=True, got ` | killed |
| M-unanchored  second arm scheme pattern loses ^ | 0 | `validated 60 schemas and 1037 vector files` | **SURVIVOR** |
| M-noscheme  git arm keeps only the SCP branch | 1 | `validation failed: schema case manager-config-v2/valid.json against manager-config-v2.schema.json: expected valid=True, got {'range': '^1.2', 'source'` | killed |
| M-noscp  git arm keeps only the scheme branch | 1 | `validation failed: schema case manager-config-v2/valid-overlay-git-scp.json against manager-config-v2.schema.json: expected valid=True, got {'range': ` | killed |

schema restored byte-exact: True
