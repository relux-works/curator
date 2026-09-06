# Classification matrix — PR #47 at 2f2dfa4, driven against the committed schema

Driven through `Draft202012Validator` with `tools/validate.py`'s own `schema_registry()`, jsonschema 4.25.1.
`kind` is derived: bare VALID + form INVALID = path; bare INVALID + form VALID = git; both INVALID = refused outright.

| source | bare | range | tag | revision | directory | branch | weight | kind |
|---|---|---|---|---|---|---|---|---|
| `/Users/operator/context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C:\Users\operator\context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C:/Users/operator/context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C://Users/operator/context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C:///Users/operator/context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `D:\ctx` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `c:/x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `c:\x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C:foo` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `c:example/team-context` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `packages/team-context` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `./here` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `../sibling` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `team` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `.` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `..` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `\\server\share\x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `//host/x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `https://github.com/example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `HTTPS://github.com/example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `http://github.com/example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `ssh://git@host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `SSH://git@host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git://host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `GIT://host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `file:///x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `FILE:///x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `file://server/share/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `svn://host/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `SVN://host/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `ftp://host/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `gitfoo://host/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `g://host/x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `x://y` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `1://y` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `git@github.com:example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `github.com:example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@my_host:x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `my_host:x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `git@host:/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `host:/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `git@c:/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `a/b:c` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `:` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `a:b/c` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@github.com:/example/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |

## Edge sweep

| source | bare | range | tag | revision | directory | branch | weight | kind |
|---|---|---|---|---|---|---|---|---|
| `git@github.com:/example/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `github.com:/example/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `git@github.com:example/x.git` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@GitHub.com:example/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@host:x/` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `-host:x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `host-:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `host.:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@host:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `user.name@host:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `user_name@host:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `1:foo` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `1:/y` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `g:/x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `g://host/x` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `gg://host/x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `:x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `x:` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `C:` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `C:\` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `C:/` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `git@host:./x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@host:../x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `git@host: x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `https://user:pw@host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `https://host:8080/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `https://host/x?y` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `https://host/x#y` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `https://host/x%20y` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `https://host/x\y` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `ssh://host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `sSh://host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `HtTpS://host/x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `my_host:x` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `my-host:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `my.host:x` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `project:v2/ctx` | inval | VALID | VALID | VALID | inval | inval | inval | git |
| `proj_ect:v2/ctx` | inval | inval | inval | inval | inval | inval | inval | REFUSED |
| `../a:b/c` | VALID | inval | inval | inval | inval | inval | VALID | path |
| `./a:b` | VALID | inval | inval | inval | inval | inval | VALID | path |
