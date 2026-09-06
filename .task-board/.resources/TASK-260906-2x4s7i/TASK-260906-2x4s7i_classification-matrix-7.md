# Cycle-7 classification matrix and core 6.1 cross-check (87a0d00)

Driven by the cycle-7 reviewer against the committed manager-config-v2 schema through
Draft202012Validator with tools/validate.py's own registry, in a clean --shared clone.

## Matrix: 74 named spellings x 8 member combinations

base overlay profile: companyA template entry: {"source": "/Users/operator/context"}

| source | bare | range | tag | revision | directory | branch | weight | two-forms | kind |
|---|---|---|---|---|---|---|---|---|---|
| `/Users/operator/context` | Y | n | n | n | n | n | Y | n | **path** |
| `/` | Y | n | n | n | n | n | Y | n | **path** |
| `./ctx` | Y | n | n | n | n | n | Y | n | **path** |
| `ctx` | Y | n | n | n | n | n | Y | n | **path** |
| `packages/team/context` | Y | n | n | n | n | n | Y | n | **path** |
| `.` | Y | n | n | n | n | n | Y | n | **path** |
| `..` | Y | n | n | n | n | n | Y | n | **path** |
| `a/b:c` | Y | n | n | n | n | n | Y | n | **path** |
| `packages/team:context` | Y | n | n | n | n | n | Y | n | **path** |
| `./a:b` | Y | n | n | n | n | n | Y | n | **path** |
| `C:\Users\operator\context` | Y | n | n | n | n | n | Y | n | **path** |
| `C:/Users/operator/context` | Y | n | n | n | n | n | Y | n | **path** |
| `C://Users/operator/context` | Y | n | n | n | n | n | Y | n | **path** |
| `C:///x` | Y | n | n | n | n | n | Y | n | **path** |
| `c:\users\operator\context` | Y | n | n | n | n | n | Y | n | **path** |
| `c:/users/operator/context` | Y | n | n | n | n | n | Y | n | **path** |
| `Z:\x` | Y | n | n | n | n | n | Y | n | **path** |
| `C:` | n | n | n | n | n | n | n | n | **REFUSED** |
| `C:foo` | n | Y | Y | Y | n | n | n | n | **git** |
| `c:example/team-context` | n | Y | Y | Y | n | n | n | n | **git** |
| `\\server\share\x` | Y | n | n | n | n | n | Y | n | **path** |
| `//host/x` | Y | n | n | n | n | n | Y | n | **path** |
| `//server/share/x` | Y | n | n | n | n | n | Y | n | **path** |
| `https://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `http://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `ssh://git@github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `git://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `HTTPS://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `SSH://git@github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `GIT://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `HTTP://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `HtTpS://github.com/example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `svn://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `file:///x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `file://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `ftp://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `s3://b/k` | n | n | n | n | n | n | n | n | **REFUSED** |
| `git+ssh://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `httpx://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `http2://host/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `1://x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `+://x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `-://x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `.://x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `git@github.com:example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `github.com:example/x` | n | Y | Y | Y | n | n | n | n | **git** |
| `host:x` | n | Y | Y | Y | n | n | n | n | **git** |
| `git@my_host:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `my_host:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `git@github.com:/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `github.com:/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `github.com:\example\x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `git@github.com:\x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `HOST:/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `HOST:x` | n | Y | Y | Y | n | n | n | n | **git** |
| `git@:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `@host:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `user@host:x` | n | Y | Y | Y | n | n | n | n | **git** |
| `git@github.com:` | n | n | n | n | n | n | n | n | **REFUSED** |
| `github.com:` | n | n | n | n | n | n | n | n | **REFUSED** |
| `:` | n | n | n | n | n | n | n | n | **REFUSED** |
| `:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `:/x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `:/:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `::x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `a:` | n | n | n | n | n | n | n | n | **REFUSED** |
| `x:y:z` | n | Y | Y | Y | n | n | n | n | **git** |
| ` :x` | Y | n | n | n | n | n | Y | n | **path** |
| `	:x` | Y | n | n | n | n | n | Y | n | **path** |
| `-host:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `.host:x` | n | n | n | n | n | n | n | n | **REFUSED** |
| `host-.:x` | n | Y | Y | Y | n | n | n | n | **git** |
| `9host:x` | n | Y | Y | Y | n | n | n | n | **git** |
| `git@host.example.com:a/b` | n | Y | Y | Y | n | n | n | n | **git** |

kind counts: {'path': 22, 'REFUSED': 32, 'git': 20}
partition violations: 0

## Independent core 6.1 cross-check over 3,916 generated spellings

corpus size: 3916
kinds: Counter({'REFUSED': 1991, 'git': 1647, 'path': 278})

INVARIANT A  valid 6.1 network form not classified git: 0

INVARIANT B  unambiguous section-1 path not classified path: 0

INVARIANT C  MIXED classifications: 0

INVARIANT D  colon-shaped non-drive spellings classified path: 0

INVARIANT E  git-classified with neither allowed scheme nor 6.1 host: 0
