# Classification matrix — cycle 4, PR #47 at `18dca85`

Every row driven against the **committed** `manager-config-v2.schema.json` through
`Draft202012Validator` using `tools/validate.py`'s own `schema_registry()` (`jsonschema==4.25.1`),
inside a full `manager-config` instance built from the committed `valid.json`. `V` = the whole
instance validates; `.` = it does not.

`kind` is derived, not declared: **path** = form-free valid and a form refused; **git** = form-free
refused and one form valid; **refused** = both refused.

## Partition sweep

Over **16,159** spellings (every string of length ≤ 4 over `Aa0:/\.-@ +`, plus the probe list below):

| property | result |
|---|---|
| a source valid both bare **and** carrying a form | **0** |
| a `path`-classified source admitting `range`, `tag`, `revision`, `directory`, or `branch` | **0** |
| a `git`-classified source admitting two forms | **0** |
| a `git`-classified source admitting `branch` | **0** |

## Difference against `2f2dfa4` (the F13 arm-3 carve-out removal)

16,159 spellings classified under both schemas: **0 classification differences.**

## Rows

| source | kind | bare | range | tag | revision | directory | branch | weight | two-forms |
|---|---|---|---|---|---|---|---|---|---|
| `packages/team:context` | **path** | V | . | . | . | . | . | V | . |
| `github.com:\\example\\x` | **refused** | . | . | . | . | . | . | . | . |
| `a/b:c` | **path** | V | . | . | . | . | . | V | . |
| `./a:b` | **path** | V | . | . | . | . | . | V | . |
| `../a:b/c` | **path** | V | . | . | . | . | . | V | . |
| `docs/v1:draft/ctx` | **path** | V | . | . | . | . | . | V | . |
| `packages/team:context/sub` | **path** | V | . | . | . | . | . | V | . |
| `a/b:c\\d` | **path** | V | . | . | . | . | . | V | . |
| `github.com:\\example/x` | **refused** | . | . | . | . | . | . | . | . |
| `git@github.com:\\example\\x` | **refused** | . | . | . | . | . | . | . | . |
| `C:\\example\\x` | **path** | V | . | . | . | . | . | V | . |
| `my_host:\\x` | **refused** | . | . | . | . | . | . | . | . |
| `host:\\x` | **refused** | . | . | . | . | . | . | . | . |
| `1:\\x` | **refused** | . | . | . | . | . | . | . | . |
| `github.com:/example/x` | **refused** | . | . | . | . | . | . | . | . |
| `github.com:example/x` | **git** | . | V | V | V | . | . | . | . |
| `/Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| `packages/team-context` | **path** | V | . | . | . | . | . | V | . |
| `.` | **path** | V | . | . | . | . | . | V | . |
| `..` | **path** | V | . | . | . | . | . | V | . |
| `team` | **path** | V | . | . | . | . | . | V | . |
| `C:\\Users\\operator\\context` | **path** | V | . | . | . | . | . | V | . |
| `C:/Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| `C://Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| `C:foo` | **git** | . | V | V | V | . | . | . | . |
| `c:example/team-context` | **git** | . | V | V | V | . | . | . | . |
| `C:` | **refused** | . | . | . | . | . | . | . | . |
| `\\\\server\\share\\x` | **path** | V | . | . | . | . | . | V | . |
| `//host/x` | **path** | V | . | . | . | . | . | V | . |
| `https://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| `HTTPS://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| `ssh://git@github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| `git://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| `http://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| `svn://github.com/example/x` | **refused** | . | . | . | . | . | . | . | . |
| `file:///Users/operator/context` | **refused** | . | . | . | . | . | . | . | . |
| `FILE:///Users/operator/context` | **refused** | . | . | . | . | . | . | . | . |
| `g://host/x` | **path** | V | . | . | . | . | . | V | . |
| `x://y` | **path** | V | . | . | . | . | . | V | . |
| `1://y` | **refused** | . | . | . | . | . | . | . | . |
| `context/http-mirror://x` | **path** | V | . | . | . | . | . | V | . |
| `/x/svn://y` | **path** | V | . | . | . | . | . | V | . |
| `ctx/http://y` | **path** | V | . | . | . | . | . | V | . |
| `git@github.com:example/team-context` | **git** | . | V | V | V | . | . | . | . |
| `git@my_host:example/x` | **refused** | . | . | . | . | . | . | . | . |
| `my_host:example/x` | **refused** | . | . | . | . | . | . | . | . |
| `proj_ect:v2/ctx` | **refused** | . | . | . | . | . | . | . | . |
| `git@host: x` | **refused** | . | . | . | . | . | . | . | . |
| `git@github.com:/example/x` | **refused** | . | . | . | . | . | . | . | . |
| `:` | **refused** | . | . | . | . | . | . | . | . |
| `:x` | **refused** | . | . | . | . | . | . | . | . |
| `x:` | **refused** | . | . | . | . | . | . | . | . |
| `\\` | **path** | V | . | . | . | . | . | V | . |
| ` ` | **path** | V | . | . | . | . | . | V | . |
| `a:b` | **git** | . | V | V | V | . | . | . | . |
