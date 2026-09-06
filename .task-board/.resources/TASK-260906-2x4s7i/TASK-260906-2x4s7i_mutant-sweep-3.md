# Mutant sweep — PR #47 at 2f2dfa4

30 mutants against `python3 tools/validate.py` on the committed corpus; schema restored byte-exactly (SHA-256 asserted) after each run.

M0-control             exit=0  
M-host                 exit=1  validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json against manager-config-v2.schema.json: expected valid
M-host2                exit=1  validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True,
M-dir                  exit=1  validation failed: schema case manager-config-v2/invalid-overlay-path-directory.json against manager-config-v2.schema.json: expected valid=False, got 
M-file                 exit=1  validation failed: schema case manager-config-v2/invalid-overlay-file-url-with-form.json against manager-config-v2.schema.json: expected valid=False, 
M-schemelen            exit=1  validation failed: schema case manager-config-v2/valid-overlay-path-windows-double-slash.json against manager-config-v2.schema.json: expected valid=Tr
M-drive-del2           exit=1  validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True,
M-drive-del3           exit=0  
M-drive-noslash        exit=1  validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True,
M-drive-nobslash       exit=1  validation failed: schema case manager-config-v2/valid-overlay-path-windows-slash.json against manager-config-v2.schema.json: expected valid=True, got
M-drive-single         exit=1  validation failed: schema case manager-config-v2/valid-overlay-path-windows-double-slash.json against manager-config-v2.schema.json: expected valid=Tr
M-drive-wide           exit=1  validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True,
M-arm1-del             exit=1  validation failed: schema case manager-config-v2/invalid-overlay-unknown-scheme.json against manager-config-v2.schema.json: expected valid=False, got 
M-arm2-del             exit=1  validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar.json against manager-config-v2.schema.json: expected valid=False, go
M-arm2-noscp           exit=1  validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True,
M-arm2-noscheme        exit=1  validation failed: schema case manager-config-v2/valid.json against manager-config-v2.schema.json: expected valid=True, got False schema does not allo
M-arm2-colonplus       exit=0  
M-reorder              exit=0  
M-oneof                exit=1  validation failed: schema case manager-config-v2/invalid-overlay-two-requirement-forms.json against manager-config-v2.schema.json: expected valid=Fals
M-notallof             exit=1  validation failed: schema case manager-config-v2/invalid-overlay-path-requirement-form.json against manager-config-v2.schema.json: expected valid=Fals
M-range                exit=1  validation failed: schema case manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json against manager-config-v2.schema.json: expec
M-tag                  exit=1  validation failed: schema case manager-config-v2/invalid-overlay-path-relative-requirement-form.json against manager-config-v2.schema.json: expected v
M-revision             exit=1  validation failed: schema case manager-config-v2/invalid-overlay-path-requirement-form.json against manager-config-v2.schema.json: expected valid=Fals
M-lower                exit=1  validation failed: schema case manager-config-v2/valid-overlay-git-ssh-uppercase.json against manager-config-v2.schema.json: expected valid=True, got 
M-hostcls              exit=1  validation failed: schema case manager-config-v2/valid-overlay-git-scp.json against manager-config-v2.schema.json: expected valid=True, got False sche
M-scp-bslash           exit=0  
M-unanchored1          exit=0  
M-unanchored-allow     exit=0  
M-unanchored2          exit=0  
M-unanchored-scp       exit=1  validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json against manager-config-v2.schema.json: expected valid

| mutant | exit | named case that fails | outcome |
|---|---|---|---|
| M0-control | 0 | — | green (expected) |
| M-host | 1 | validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-host2 | 1 | validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True, got False schema does not all | killed |
| M-dir | 1 | validation failed: schema case manager-config-v2/invalid-overlay-path-directory.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-file | 1 | validation failed: schema case manager-config-v2/invalid-overlay-file-url-with-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-schemelen | 1 | validation failed: schema case manager-config-v2/valid-overlay-path-windows-double-slash.json against manager-config-v2.schema.json: expected valid=True, got False schema does not  | killed |
| M-drive-del2 | 1 | validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True, got False schema does not all | killed |
| M-drive-del3 | 0 | — | **SURVIVOR** |
| M-drive-noslash | 1 | validation failed: schema case manager-config-v2/valid-overlay-path-windows-backslash.json against manager-config-v2.schema.json: expected valid=True, got False schema does not all | killed |
| M-drive-nobslash | 1 | validation failed: schema case manager-config-v2/valid-overlay-path-windows-slash.json against manager-config-v2.schema.json: expected valid=True, got False schema does not allow { | killed |
| M-drive-single | 1 | validation failed: schema case manager-config-v2/valid-overlay-path-windows-double-slash.json against manager-config-v2.schema.json: expected valid=True, got False schema does not  | killed |
| M-drive-wide | 1 | validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True, got {'range': '^1.2', 'source | killed |
| M-arm1-del | 1 | validation failed: schema case manager-config-v2/invalid-overlay-unknown-scheme.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-arm2-del | 1 | validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-arm2-noscp | 1 | validation failed: schema case manager-config-v2/valid-overlay-git-single-letter-host.json against manager-config-v2.schema.json: expected valid=True, got False schema does not all | killed |
| M-arm2-noscheme | 1 | validation failed: schema case manager-config-v2/valid.json against manager-config-v2.schema.json: expected valid=True, got False schema does not allow {'range': '^1.2', 'source':  | killed |
| M-arm2-colonplus | 0 | — | **SURVIVOR** |
| M-reorder | 0 | — | **SURVIVOR** |
| M-oneof | 1 | validation failed: schema case manager-config-v2/invalid-overlay-two-requirement-forms.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-notallof | 1 | validation failed: schema case manager-config-v2/invalid-overlay-path-requirement-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-range | 1 | validation failed: schema case manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-tag | 1 | validation failed: schema case manager-config-v2/invalid-overlay-path-relative-requirement-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-revision | 1 | validation failed: schema case manager-config-v2/invalid-overlay-path-requirement-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |
| M-lower | 1 | validation failed: schema case manager-config-v2/valid-overlay-git-ssh-uppercase.json against manager-config-v2.schema.json: expected valid=True, got False schema does not allow {' | killed |
| M-hostcls | 1 | validation failed: schema case manager-config-v2/valid-overlay-git-scp.json against manager-config-v2.schema.json: expected valid=True, got False schema does not allow {'range': '^ | killed |
| M-scp-bslash | 0 | — | **SURVIVOR** |
| M-unanchored1 | 0 | — | **SURVIVOR** |
| M-unanchored-allow | 0 | — | **SURVIVOR** |
| M-unanchored2 | 0 | — | **SURVIVOR** |
| M-unanchored-scp | 1 | validation failed: schema case manager-config-v2/invalid-overlay-scp-host-grammar-with-form.json against manager-config-v2.schema.json: expected valid=False, got valid | killed |

## Survivor behavioural diffs

A survivor with an empty diff is a semantic no-op; a survivor with a non-empty diff is unpinned behaviour.

### M-drive-del3 — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-reorder — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-arm2-colonplus — 2 classification change(s)
| source | committed | mutant |
|---|---|---|
| `:` | REFUSED | path |
| `:x` | REFUSED | path |

### M-scp-bslash — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-unanchored1 — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-unanchored-allow — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-unanchored2 — 3 classification change(s)
| source | committed | mutant |
|---|---|---|
| `a/b:c` | path | REFUSED |
| `../a:b/c` | path | REFUSED |
| `./a:b` | path | REFUSED |

### Targeted probes

### M-drive-del3 — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-reorder — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-arm2-colonplus — 0 classification change(s)
semantic no-op: identical classification on every probed source

### M-scp-bslash — 3 classification change(s)
| source | committed | mutant |
|---|---|---|
| `host:\x` | REFUSED | git |
| `git@host:\x` | REFUSED | git |
| `ab:\share\y` | REFUSED | git |

### M-unanchored1 — 3 classification change(s)
| source | committed | mutant |
|---|---|---|
| `/x/svn://y` | path | REFUSED |
| `ctx/http://y` | path | REFUSED |
| `x:svn://y` | git | REFUSED |

### M-unanchored-allow — 1 classification change(s)
| source | committed | mutant |
|---|---|---|
| `ctx/http://y` | path | git |

### M-unanchored2 — 3 classification change(s)
| source | committed | mutant |
|---|---|---|
| `/x/svn://y` | path | REFUSED |
| `ctx/http://y` | path | REFUSED |
| `a/b:c` | path | REFUSED |
