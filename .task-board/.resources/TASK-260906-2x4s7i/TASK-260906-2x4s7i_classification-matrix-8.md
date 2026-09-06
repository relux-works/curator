# Classification matrix — cycle 8, at the landed commit `87a0d00`

Everything below was driven by me against the **committed** `schemas/v1/manager-config-v2.schema.json`
at `87a0d0060bad64ab883d007dcdf35df7485368bf` (= `origin/main`), through
`Draft202012Validator` with the registry `tools/validate.py` builds, in a clean `--shared` clone
(1188 tracked paths, `git status --porcelain` empty) against a venv from the repo's own pinned
`requirements-dev.txt` (`jsonschema==4.25.1`).

The instance is a **whole** `manager-config-v2` document (the `valid-overlay-path-source` case with
its single overlay entry replaced), never `$defs/overlay` in isolation. The kind is derived from
**behaviour**, never read off the pattern:

- `path` = valid bare, invalid with each of `range`/`tag`/`revision`
- `git` = invalid bare, valid with each of the three
- `REFUSED` = valid under no member combination
- `MIXED` = anything else

## 1. 80 named spellings x 8 member combinations

`V` = the document validates, `.` = it does not.

| # | source (repr) | kind | bare | range | tag | rev | dir | branch | weight | 2forms |
|---:|---|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| 1 | `/Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| 2 | `/tmp/app` | **path** | V | . | . | . | . | . | V | . |
| 3 | `relative/context` | **path** | V | . | . | . | . | . | V | . |
| 4 | `context` | **path** | V | . | . | . | . | . | V | . |
| 5 | `.` | **path** | V | . | . | . | . | . | V | . |
| 6 | `..` | **path** | V | . | . | . | . | . | V | . |
| 7 | `./context` | **path** | V | . | . | . | . | . | V | . |
| 8 | `../context` | **path** | V | . | . | . | . | . | V | . |
| 9 | `packages/team:context` | **path** | V | . | . | . | . | . | V | . |
| 10 | `a/b:c` | **path** | V | . | . | . | . | . | V | . |
| 11 | `/x/svn://host/y` | **path** | V | . | . | . | . | . | V | . |
| 12 | `//host/x` | **path** | V | . | . | . | . | . | V | . |
| 13 | `\\\\server\\share\\x` | **path** | V | . | . | . | . | . | V | . |
| 14 | `\\\\server\\share:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 15 | `C:\\Users\\operator\\context` | **path** | V | . | . | . | . | . | V | . |
| 16 | `c:\\Users\\operator\\context` | **path** | V | . | . | . | . | . | V | . |
| 17 | `C:/Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| 18 | `C://Users/operator/context` | **path** | V | . | . | . | . | . | V | . |
| 19 | `D:\\ctx` | **path** | V | . | . | . | . | . | V | . |
| 20 | `z:/ctx` | **path** | V | . | . | . | . | . | V | . |
| 21 | `C:` | **REFUSED** | . | . | . | . | . | . | . | . |
| 22 | `C:foo` | **git** | . | V | V | V | . | . | . | . |
| 23 | `c:example/team-context` | **git** | . | V | V | V | . | . | . | . |
| 24 | `C:\\Users\\x:y` | **path** | V | . | . | . | . | . | V | . |
| 25 | `https://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| 26 | `http://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| 27 | `ssh://git@host/x` | **git** | . | V | V | V | . | . | . | . |
| 28 | `git://host/x` | **git** | . | V | V | V | . | . | . | . |
| 29 | `HTTPS://github.com/example/x` | **git** | . | V | V | V | . | . | . | . |
| 30 | `HTTP://h/x` | **git** | . | V | V | V | . | . | . | . |
| 31 | `SSH://h/x` | **git** | . | V | V | V | . | . | . | . |
| 32 | `GIT://h/x` | **git** | . | V | V | V | . | . | . | . |
| 33 | `hTtPs://h/x` | **git** | . | V | V | V | . | . | . | . |
| 34 | `https://host:8080/x` | **git** | . | V | V | V | . | . | . | . |
| 35 | `https://user:pw@host/x` | **git** | . | V | V | V | . | . | . | . |
| 36 | `https://host/x?q=1` | **git** | . | V | V | V | . | . | . | . |
| 37 | `https://host/x#f` | **git** | . | V | V | V | . | . | . | . |
| 38 | `https://host/%41` | **git** | . | V | V | V | . | . | . | . |
| 39 | `https://host/x\\y` | **git** | . | V | V | V | . | . | . | . |
| 40 | `ssh:///x` | **git** | . | V | V | V | . | . | . | . |
| 41 | `https:///x` | **git** | . | V | V | V | . | . | . | . |
| 42 | `https://host` | **git** | . | V | V | V | . | . | . | . |
| 43 | `https://ho_st/x` | **git** | . | V | V | V | . | . | . | . |
| 44 | `file:///x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 45 | `file://host/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 46 | `svn://host/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 47 | `ftp://host/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 48 | `s3://b/k` | **REFUSED** | . | . | . | . | . | . | . | . |
| 49 | `git+ssh://h/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 50 | `HTTPX://h/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 51 | `1://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 52 | `+://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 53 | `-://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 54 | `.://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 55 | `a://x` | **path** | V | . | . | . | . | . | V | . |
| 56 | `ab://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 57 | `git@github.com:example/x` | **git** | . | V | V | V | . | . | . | . |
| 58 | `github.com:example/x` | **git** | . | V | V | V | . | . | . | . |
| 59 | `user@host:x` | **git** | . | V | V | V | . | . | . | . |
| 60 | `host:x` | **git** | . | V | V | V | . | . | . | . |
| 61 | `9host:x` | **git** | . | V | V | V | . | . | . | . |
| 62 | `my_host:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 63 | `git@my_host:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 64 | `github.com:\\example\\x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 65 | `@host:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 66 | `host:` | **REFUSED** | . | . | . | . | . | . | . | . |
| 67 | `host:/x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 68 | `:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 69 | `::` | **REFUSED** | . | . | . | . | . | . | . | . |
| 70 | `:` | **REFUSED** | . | . | . | . | . | . | . | . |
| 71 | `://x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 72 | `git@host:sub/dir/x` | **git** | . | V | V | V | . | . | . | . |
| 73 | `h.o-s.t:x` | **git** | . | V | V | V | . | . | . | . |
| 74 | `-host:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 75 | `.host:x` | **REFUSED** | . | . | . | . | . | . | . | . |
| 76 | ` ` | **path** | V | . | . | . | . | . | V | . |
| 77 | ` /Users/x` | **path** | V | . | . | . | . | . | V | . |
| 78 | `/Users/x ` | **path** | V | . | . | . | . | . | V | . |
| 79 | `x\ty` | **path** | V | . | . | . | . | . | V | . |
| 80 | `x:y z` | **git** | . | V | V | V | . | . | . | . |

counts: {'path': 25, 'REFUSED': 26, 'git': 29} total 80

## 2. Whitespace-class bypass of the arm-1 and arm-2 refusals

Each pinned refusal below has a twin that differs by one inserted character and flips to a
**form-free `path`**. Driven through the same committed schema.

| spelling | kind at head |
|---|---|
| `my_host:x` | **REFUSED** |
| `my_hos<NBSP>t:x` | **path** |
| `my_hos<VT>t:x` | **path** |
| `my_hos<IDSP>t:x` | **path** |
| `git@my_host:x` | **REFUSED** |
| `git@my_hos<NBSP>t:x` | **path** |
| `github.com:\example\x` | **REFUSED** |
| `github.com<NBSP>:\example\x` | **path** |
| `host:/x` | **REFUSED** |
| `host<NBSP>:/x` | **path** |
| `C:` | **REFUSED** |
| `C<NBSP>:` | **path** |
| `file:///x` | **REFUSED** |
| `fil<NBSP>e:///x` | **path** |
| `svn://host/x` | **REFUSED** |
| `sv<NBSP>n://host/x` | **path** |
| `@host:x` | **REFUSED** |
| `@hos<NBSP>t:x` | **path** |
| `:x` | **REFUSED** |
| `-host:x` | **REFUSED** |
| `-hos<NBSP>t:x` | **path** |

`<NBSP>` = U+00A0, `<VT>` = U+000B, `<IDSP>` = U+3000. The cause is `\s` inside
`^[^/\s]*:` (arm 2's "is this colon-shaped") and inside the SCP pattern: a character the engine
counts as whitespace makes the spelling invisible to the colon test, so arms 1 and 2 never fire and
arm 3 falls to its `else`. A plain ASCII space does the same in every engine (`my host:x` -> `path`
while `my_host:x` -> `REFUSED`).

## 3. The same document, three regex engines

JSON Schema specifies `pattern` as ECMA-262. The repository validates with CPython `re`; the pinned
reference manager is Go, whose RE2 `\s` is `[\t\n\f\r ]` only. I compiled the five overlay patterns
in all three engines and derived the arm decisions from the raw booleans.

corpus size: 154
kind-divergent spellings: 75

| source (control chars named) | CPython `re` | ECMA-262 (node) | Go RE2 |
|---|---|---|---|
| `<BOM>host:x` | REFUSED | path | REFUSED |
| `<ENQUAD>host:x` | path | path | REFUSED |
| `<FS>host:x` | path | REFUSED | REFUSED |
| `<GS>host:x` | path | REFUSED | REFUSED |
| `<IDSP>host:x` | path | path | REFUSED |
| `<LSEP>host:x` | path | path | REFUSED |
| `<MMSP>host:x` | path | path | REFUSED |
| `<NBSP>host:x` | path | path | REFUSED |
| `<NEL>host:x` | path | REFUSED | REFUSED |
| `<NNBSP>host:x` | path | path | REFUSED |
| `<OGHAM>host:x` | path | path | REFUSED |
| `<PSEP>host:x` | path | path | REFUSED |
| `<RS>host:x` | path | REFUSED | REFUSED |
| `<US>host:x` | path | REFUSED | REFUSED |
| `<VT>host:x` | path | path | REFUSED |
| `a<BOM>b:c` | REFUSED | path | REFUSED |
| `a<ENQUAD>b:c` | path | path | REFUSED |
| `a<FS>b:c` | path | REFUSED | REFUSED |
| `a<GS>b:c` | path | REFUSED | REFUSED |
| `a<IDSP>b:c` | path | path | REFUSED |
| `a<LSEP>b:c` | path | path | REFUSED |
| `a<MMSP>b:c` | path | path | REFUSED |
| `a<NBSP>b:c` | path | path | REFUSED |
| `a<NEL>b:c` | path | REFUSED | REFUSED |
| `a<NNBSP>b:c` | path | path | REFUSED |
| `a<OGHAM>b:c` | path | path | REFUSED |
| `a<PSEP>b:c` | path | path | REFUSED |
| `a<RS>b:c` | path | REFUSED | REFUSED |
| `a<US>b:c` | path | REFUSED | REFUSED |
| `a<VT>b:c` | path | path | REFUSED |
| `git@ho<BOM>st:x` | REFUSED | path | REFUSED |
| `git@ho<ENQUAD>st:x` | path | path | REFUSED |
| `git@ho<FS>st:x` | path | REFUSED | REFUSED |
| `git@ho<GS>st:x` | path | REFUSED | REFUSED |
| `git@ho<IDSP>st:x` | path | path | REFUSED |
| `git@ho<LSEP>st:x` | path | path | REFUSED |
| `git@ho<MMSP>st:x` | path | path | REFUSED |
| `git@ho<NBSP>st:x` | path | path | REFUSED |
| `git@ho<NEL>st:x` | path | REFUSED | REFUSED |
| `git@ho<NNBSP>st:x` | path | path | REFUSED |
| `git@ho<OGHAM>st:x` | path | path | REFUSED |
| `git@ho<PSEP>st:x` | path | path | REFUSED |
| `git@ho<RS>st:x` | path | REFUSED | REFUSED |
| `git@ho<US>st:x` | path | REFUSED | REFUSED |
| `git@ho<VT>st:x` | path | path | REFUSED |
| `host:<BOM>x` | git | REFUSED | git |
| `host:<ENQUAD>x` | REFUSED | REFUSED | git |
| `host:<FS>x` | REFUSED | git | git |
| `host:<GS>x` | REFUSED | git | git |
| `host:<IDSP>x` | REFUSED | REFUSED | git |
| `host:<LSEP>x` | REFUSED | REFUSED | git |
| `host:<MMSP>x` | REFUSED | REFUSED | git |
| `host:<NBSP>x` | REFUSED | REFUSED | git |
| `host:<NEL>x` | REFUSED | git | git |
| `host:<NNBSP>x` | REFUSED | REFUSED | git |
| `host:<OGHAM>x` | REFUSED | REFUSED | git |
| `host:<PSEP>x` | REFUSED | REFUSED | git |
| `host:<RS>x` | REFUSED | git | git |
| `host:<US>x` | REFUSED | git | git |
| `host:<VT>x` | REFUSED | REFUSED | git |
| `host<BOM>x:y` | REFUSED | path | REFUSED |
| `host<ENQUAD>x:y` | path | path | REFUSED |
| `host<FS>x:y` | path | REFUSED | REFUSED |
| `host<GS>x:y` | path | REFUSED | REFUSED |
| `host<IDSP>x:y` | path | path | REFUSED |
| `host<LSEP>x:y` | path | path | REFUSED |
| `host<MMSP>x:y` | path | path | REFUSED |
| `host<NBSP>x:y` | path | path | REFUSED |
| `host<NEL>x:y` | path | REFUSED | REFUSED |
| `host<NNBSP>x:y` | path | path | REFUSED |
| `host<OGHAM>x:y` | path | path | REFUSED |
| `host<PSEP>x:y` | path | path | REFUSED |
| `host<RS>x:y` | path | REFUSED | REFUSED |
| `host<US>x:y` | path | REFUSED | REFUSED |
| `host<VT>x:y` | path | path | REFUSED |

P1_scheme_any: py!=js 0, py!=go 0, js!=go 0
P2_scheme_ok: py!=js 0, py!=go 0, js!=go 0
P3_colonish: py!=js 24, py!=go 56, js!=go 40
P4_drive: py!=js 0, py!=go 0, js!=go 0
P5_scp: py!=js 6, py!=go 14, js!=go 10

`P1`/`P2` (the scheme patterns) and `P4` (the drive pattern) are engine-identical. All divergence
comes from `\s` in `P3_colonish` and `P5_scp`.

## 4. Corpus coverage of that class, measured

| Measure | Value |
|---|---:|
| Overlay `source` declarations across the 71 schema cases and 41 v2 vectors | 162 |
| Distinct `source` spellings among them | 24 |
| Declarations whose `source` contains **any** whitespace character | **0** |
| Declarations whose `source` contains any non-ASCII character | **0** |

## 5. Invariants

Over the 80 named spellings in section 1, checked a priori:

| Invariant | Violations |
|---|---:|
| Any `MIXED` classification | 0 |
| `branch` admitted on any source | 0 |
| Two requirement forms admitted on any source | 0 |
| A `path` admitting `directory` | 0 |
| A spelling valid both bare and carrying a form | 0 |
| A `path` refusing `weight` | 0 |
| A `git` admitting `weight` without a form | 0 |

## 6. Second declaration surface

`$defs/overlay` is the only overlay **declaration** shape in `schemas/v1`. `context-lock-v1` and
`agent-environment-marker-v1` both spell `overlay` as a `boolean` member flag, not a declaration;
`system-config-v2` carries only `overlays_allowed`. The one other unconditional
`range | tag | revision` `oneOf` in the schema set is `agent-context-v1#/$defs/requirementForm`,
which governs a package's `requires` edges — and environments section 1 states that
"a package's `requires` (section 2) never names a `path` source", so an unconditional form is
correct there. No bypass surface.
