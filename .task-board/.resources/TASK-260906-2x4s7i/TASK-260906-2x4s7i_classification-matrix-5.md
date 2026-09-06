# Overlay classification matrix — cycle 5, PR #47 at `407424e`

Every row driven through the **committed** `manager-config-v2.schema.json`
(SHA-256 `d93a6673a989bc0dadc3029db4cbfa5c83c5e8037f402c51c459ebc5b6bef64d`) with
`Draft202012Validator` and the repository's own `schemas/v1/*.json` registry, exactly as
`tools/validate.py` builds it. `jsonschema==4.25.1`. `validate_wire_semantics` does not reach
`manager-config-v2`, so the schema alone decides.

Each column is the member added alongside `source` in one overlay of an otherwise valid
`manager-config-v2` instance. `V` = the instance validates.

- `path`  = valid bare, invalid with any requirement form or `directory`
- `git`   = invalid bare, valid with exactly one form, invalid with two
- `refused` = invalid in every column

```
Overlay classification matrix — committed manager-config-v2 schema at 407424e
V = the instance validates; - = it does not.  Columns are the overlay member added to `source`.


## POSIX and project-relative paths

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'/Users/operator/context'                path     V         -         -         -         -         -         V         -        
'packages/team-context'                  path     V         -         -         -         -         -         V         -        
'team'                                   path     V         -         -         -         -         -         V         -        
'.'                                      path     V         -         -         -         -         -         V         -        
'..'                                     path     V         -         -         -         -         -         V         -        
'./ctx'                                  path     V         -         -         -         -         -         V         -        
'../ctx'                                 path     V         -         -         -         -         -         V         -        
'packages/team:context'                  path     V         -         -         -         -         -         V         -        
'a/b:c'                                  path     V         -         -         -         -         -         V         -        
'a/b:c\\d'                               path     V         -         -         -         -         -         V         -        
'ctx/http://y'                           path     V         -         -         -         -         -         V         -        
'a/b/https://c'                          path     V         -         -         -         -         -         V         -        
'./notes:2026'                           path     V         -         -         -         -         -         V         -        
'x/notes:2026'                           path     V         -         -         -         -         -         V         -        

## Windows spellings

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'C:\\Users\\operator\\context'           path     V         -         -         -         -         -         V         -        
'C:/Users/operator/context'              path     V         -         -         -         -         -         V         -        
'C://Users/operator/context'             path     V         -         -         -         -         -         V         -        
'c:\\users\\operator\\context'           path     V         -         -         -         -         -         V         -        
'c:/users/operator/context'              path     V         -         -         -         -         -         V         -        
'D:\\ctx'                                path     V         -         -         -         -         -         V         -        
'd:/ctx'                                 path     V         -         -         -         -         -         V         -        
'z:/x'                                   path     V         -         -         -         -         -         V         -        
'\\\\server\\share\\x'                   path     V         -         -         -         -         -         V         -        
'C:foo'                                  git      -         V         V         V         -         -         -         -        
'C:'                                     refused  -         -         -         -         -         -         -         -        
'C: '                                    refused  -         -         -         -         -         -         -         -        
'c:'                                     refused  -         -         -         -         -         -         -         -        

## scheme URLs

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'https://github.com/example/x'           git      -         V         V         V         -         -         -         -        
'http://h/x'                             git      -         V         V         V         -         -         -         -        
'git://h/x'                              git      -         V         V         V         -         -         -         -        
'ssh://h/x'                              git      -         V         V         V         -         -         -         -        
'HTTPS://github.com/example/x'           git      -         V         V         V         -         -         -         -        
'SSH://h/x'                              git      -         V         V         V         -         -         -         -        
'GIT://h/x'                              git      -         V         V         V         -         -         -         -        
'HTTP://h/x'                             git      -         V         V         V         -         -         -         -        
'svn://h/x'                              refused  -         -         -         -         -         -         -         -        
'ftp://h/x'                              refused  -         -         -         -         -         -         -         -        
'file:///Users/operator/context'         refused  -         -         -         -         -         -         -         -        
'FILE:///x'                              refused  -         -         -         -         -         -         -         -        
'g://host/x'                             path     V         -         -         -         -         -         V         -        
'//host/x'                               path     V         -         -         -         -         -         V         -        

## core 6.1-invalid URL decorations

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'https://h:22/x'                         git      -         V         V         V         -         -         -         -        
'https://user:pw@h/x'                    git      -         V         V         V         -         -         -         -        
'https://h/x?q=1'                        git      -         V         V         V         -         -         -         -        
'https://h/x#frag'                       git      -         V         V         V         -         -         -         -        
'https://h/x%20y'                        git      -         V         V         V         -         -         -         -        
'https://h/x\\y'                         git      -         V         V         V         -         -         -         -        

## SCP forms

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'git@github.com:example/x'               git      -         V         V         V         -         -         -         -        
'github.com:example/x'                   git      -         V         V         V         -         -         -         -        
'c:example/team-context'                 git      -         V         V         V         -         -         -         -        
'a:b'                                    git      -         V         V         V         -         -         -         -        
'notes:2026'                             git      -         V         V         V         -         -         -         -        
'1host:x'                                git      -         V         V         V         -         -         -         -        
'host-:x'                                git      -         V         V         V         -         -         -         -        
'-host:x'                                refused  -         -         -         -         -         -         -         -        
'git@my_host:example/x'                  refused  -         -         -         -         -         -         -         -        
'my_host:example/x'                      refused  -         -         -         -         -         -         -         -        
'proj_ect:v2/ctx'                        refused  -         -         -         -         -         -         -         -        
'github.com:\\example\\x'                refused  -         -         -         -         -         -         -         -        
'github.com:example\\x'                  git      -         V         V         V         -         -         -         -        
'git@github.com:example/x\\y'            git      -         V         V         V         -         -         -         -        
'git@github.com:/example/x'              refused  -         -         -         -         -         -         -         -        
'github.com:/example/x'                  refused  -         -         -         -         -         -         -         -        
'abc:/x'                                 refused  -         -         -         -         -         -         -         -        
'git@host: x'                            refused  -         -         -         -         -         -         -         -        

## degenerate

source                                   class    bare      range     tag       revision  directory branch    weight    two-forms
---------------------------------------------------------------------------------------------------------------------------------
'x:'                                     refused  -         -         -         -         -         -         -         -        
':'                                      refused  -         -         -         -         -         -         -         -        
':x'                                     refused  -         -         -         -         -         -         -         -        
' '                                      path     V         -         -         -         -         -         V         -        
'  a'                                    path     V         -         -         -         -         -         V         -        


## Partition sweep over 2258 generated spellings

classified: {'path': 1238, 'git': 377, 'refused': 643}
valid both bare and form-carrying : 0 []
path admitting a forbidden member : 0 []
git skipping its form / 2 forms / branch / directory: 0 []
anomalous (neither of the three)  : 0 []
```

## Legal combinations checked separately

| instance | expected by §1 | measured |
|---|---|---|
| `git` source + `range` + `directory` | valid — "A `git` declaration MAY carry `directory`" | valid |
| `git` source + `range` + `weight` | valid | valid |
| `path` source + `weight` | valid — `{ source, weight? }` | valid |
| `path` source + `directory` | invalid — §1 forbids `directory` on a `path` declaration | invalid |
