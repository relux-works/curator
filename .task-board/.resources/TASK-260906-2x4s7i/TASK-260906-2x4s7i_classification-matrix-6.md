# Classification matrix — cycle 6, driven against the committed schema at 87a0d00

Driven through `Draft202012Validator` with the repository's own schema registry
(`tools/validate.py:schema_registry` shape), against the **whole document** —
`valid-overlay-path-source.json` used as the carrier and its single overlay member
replaced — not against `$defs/overlay` in isolation. jsonschema 4.25.1
(the `requirements-dev.txt` pin).

Kind is *derived from behaviour*, never read from the pattern: `path` = valid bare and
invalid with `range`; `git` = the converse; `refused` = invalid both ways;
`both!` = a partition violation.

`V` = the document validates. `.` = it does not.

```
source                         | kind    | bare | range | tag | revision | directory | branch | weight | range+dir | range+tag
-------------------------------+---------+------+-------+-----+----------+-----------+--------+--------+-----------+----------
'/Users/operator/context'      | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'packages/team-context'        | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'packages/team:context'        | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'.'                            | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'..'                           | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'./x'                          | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'a/b:c'                        | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'~/ctx'                        | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'C:\\Users\\operator\\context' | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'C:/Users/operator/context'    | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'C://Users/operator/context'   | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'c:\\users\\operator\\context' | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'c:/users/operator/context'    | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'D:\\ctx'                      | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'z:/x'                         | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'\\\\server\\share\\x'         | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'//host/x'                     | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'https://github.com/example/x' | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'http://h/x'                   | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'ssh://h/x'                    | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'git://h/x'                    | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'HTTPS://H/x'                  | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'SSH://h/x'                    | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'GIT://h/x'                    | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'HTTP://h/x'                   | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'git@github.com:example/x'     | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'github.com:example/x'         | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'c:example/team-context'       | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'notes:2026'                   | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'a:b'                          | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'svn://host/x'                 | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'file:///x'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'FILE:///x'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'ftp://h/x'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'git@my_host:x'                | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'my_host:x'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'github.com:\\example\\x'      | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'C:'                           | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
':'                            | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
':x'                           | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'git@:x'                       | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'g://host/x'                   | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'C:foo'                        | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'c:'                           | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'http://'                      | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'https://h/x?q=1'              | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'https://h:22/x'               | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'https://h/x\\y'               | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'github.com:example\\x'        | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'git@github.com:example/x\\y'  | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'HtTpS://h/x'                  | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'http://h/x'                   | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'xxxxxxxxxxxxxxxxxxxx://h/x'   | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'1a://h/x'                     | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'a+b://h/x'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'-a://h/x'                     | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'C:\\'                         | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'C:/'                          | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'\\x'                          | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'x:'                           | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'host:'                        | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'git@host:'                    | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'git@host:/x'                  | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'host:/x'                      | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'host:.'                       | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'.:x'                          | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'-h:x'                         | refused | .    | .     | .   | .        | .         | .      | .      | .         | .        
'h-:x'                         | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
'h.:x'                         | git     | .    | V     | V   | V        | .         | .      | .      | V         | .        
' /x'                          | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        
'/x y'                         | path    | V    | .     | .   | .        | .         | .      | V      | .         | .        

expectation mismatches: NONE
partition violations (valid both bare and with a form): NONE
```

## Identical at both heads

The same 74 spellings driven against the `407424e` schema produce a byte-identical
table (`diff` of the two runs is empty). The landed commit moved no classification.

## Partition sweep

2,779 generated spellings (user×host×separator×tail cross-product plus 45 hand-written
edges): **0** valid both bare and form-carrying, **0** outside the three-way partition.
Split: 120 `path` / 676 `git` / 1,983 `refused`.

## Bypass hunt — core section 6.1 "not treated as local"

The one failure that would matter is a section 6.1 network-shaped spelling escaping to
`path`, since that skips the form requirement *and* contradicts 6.1's mandate. Measured
over the same 2,779 spellings:

| assertion | result |
|---|---|
| `path`-classified spellings that are colon-shaped (`^[^/\s]*:`) but not a Windows drive | **0** of 120 |
| `git`-classified spellings with neither an allowed scheme nor a 6.1-grammar host | **0** of 676 |

So every colon-shaped spelling is a drive letter (path), a 6.1-legal network form (git),
or refused. There is no third road to `path`.
