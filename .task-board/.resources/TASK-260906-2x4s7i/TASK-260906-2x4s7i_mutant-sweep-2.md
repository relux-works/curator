| mutant | exit | first failing case | outcome |
|---|---|---|---|
| M-branch  overlay admits `branch` again (no path-side ban) | 1 | `validation failed: schema case manager-config-v2/invalid-unknown-overlay-field.json against manager-config-v2.schema.json: expected valid=Fa` | killed |
| M-branch2 overlay admits `branch`, banned only on the path arm | 1 | `validation failed: schema case manager-config-v2/invalid-unknown-overlay-field.json against manager-config-v2.schema.json: expected valid=Fa` | killed |
| M-range   path else drops `range` (narrow) | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-path-windows-slash-requirement-form.json against manager-config-v2.schema.j` | killed |
| M-tag     path else drops `tag` (narrow) | 1 | `validation failed: schema case manager-config-v2/invalid-overlay-path-relative-requirement-form.json against manager-config-v2.schema.json: ` | killed |
| M-hostcls host class -> [A-Za-z0-9]+ (no dot/hyphen) | 1 | `validation failed: schema case manager-config-v2/valid-overlay-git-scp.json against manager-config-v2.schema.json: expected valid=True, got ` | killed |
| M-host2   host must be >=2 chars (excludes drive letters) | 0 | `validated 60 schemas and 1037 vector files` | **SURVIVOR** |

restored: True
