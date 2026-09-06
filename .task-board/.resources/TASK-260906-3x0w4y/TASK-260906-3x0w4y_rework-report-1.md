# Rework report 1 — TASK-260906-3x0w4y: path-overlay-cannot-be-declared

Rework of commit `535fda6` (kept: POSIX matrix, untouched §1/§6 prose,
pin-consumption check, disclosed survivors). Work is left **uncommitted**
in the story worktree per the spawn template, so the "exactly one signed
commit" item is for the orchestrator's integration step, not claimed here.
`origin/main` = `550579d`; story base = `c25d78e`.

## 1. Finding → resolution (F1–F4)

| Finding | Resolution |
|---|---|
| F1 BLOCKING — `C:\…` classified git: bare INVALID (undeclarable), +revision VALID (teaches the §1-invalid shape) | Fixed. SCP arm now uses the core §6.1 host grammar `[A-Za-z0-9][A-Za-z0-9.-]*` and the post-colon class excludes backslash (`[^/\s\\]`). `C:\…` bare → VALID, `C:\…`+form → INVALID, verified both directions (§4). Not taken on trust: reviewer's suggested pattern shape adopted, then the full matrix driven independently. |
| F2 MAJOR — discriminator has no killing evidence (corpus: 1 POSIX path + 3 https spellings) | Fixed. 18 schema cases + 18 vector twins published, one per arm (§5). Old M3 (`^/`-only + swapped arms) now dies; reviewer's repair-is-green observation now dies (old pattern dies on the Windows cases, §6). |
| F3 MAJOR — `profiles/manager.md:2189-2190` still mandates the universal form | Fixed, in scope. Sentence now states the manager-side obligation and cites environments §12.1 instead of restating the universal shape (§7). |
| F4 MINOR — `svn://host/x` classified path, admitted form-free | Refused, not bounded. New second `allOf` element rejects any `://` URL whose scheme is outside `ssh`/`git`/`http`/`https`/`file` (RE2-safe, no lookahead — `allOf` + `not` + `then: false`). `file:` stays a path (core §6.1: "Local paths and `file:` URLs have no network identity"). Two negative cases pin it (§5). |

## 2. §12.1 knob row — before / after

Unchanged by this rework (already correct since `535fda6`); quoted for the record.
`protocol/environments.md:2222`, one-line table shape kept:

Before (base `c25d78e`):

```
| `overlays.<profile>` | ordered list of `{ source, range \| tag \| revision, directory?, weight? }` | empty | 6 |
```

After (identical in this rework's tree):

```
| `overlays.<profile>` | ordered list of `{ source, range \| tag \| revision, directory?, weight? }` for a `git` source, `{ source, weight? }` for a `path` source | empty | 6 |
```

`git diff c25d78e` on `protocol/environments.md` is exactly this one line;
§1 and §6 prose are byte-identical to the base.

## 3. `$defs/overlay` — before / after

Before (`535fda6`, `schemas/v1/manager-config-v2.schema.json` `$defs/overlay`
`allOf[0].if.properties.source.pattern`):

```
^(?:(?:[Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://|(?:[^/:\s@]+@)?[^/:\s]+:[^/\s])
```

After (this rework):

```
^(?:(?:[Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://|(?:[^/:\s@]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\s\\])
```

plus a second `allOf` element (F4):

```json
{"if": {"allOf": [
  {"properties": {"source": {"pattern": "^[A-Za-z][A-Za-z0-9+.-]*://"}}},
  {"not": {"properties": {"source": {"pattern": "^([Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?|[Ff][Ii][Ll][Ee])://"}}}}
]}, "then": false}
```

`then`/`else` arms are unchanged: git ⇒ `oneOf` range/tag/revision;
path ⇒ `not anyOf` range/tag/revision/directory (`branch` refused by the
closed object on either arm; `weight` admitted on both).

## 4. Discriminator decision (from §1's own words)

§1: "`git` — a network git source under the core §6.1 canonical identity";
"`path` — an operator-local package directory named by an absolute path,
or by a project-relative path" (no spelling qualification). Core §6.1:
network URLs "use `ssh`, `git`, `http`, or `https`, or SCP form
`[user@]host:path`", "have an ASCII host matching
`[A-Za-z0-9][A-Za-z0-9.-]*`", and "contain no … backslash". Two consequences
decide the pattern: the SCP host class is the §6.1 host grammar (not the
permissive `[^/:\s]+`), and a backslash after the colon excludes the git
arm — a backslash spelling cannot be a §6.1 network form, so classifying it
git contradicts the pattern's own cited authority. A Windows absolute path
(`C:\…`) is a legal `path` source because §1's "absolute path" is
unqualified and `portablePath` governs `directory`, not `source`
(`source` is `nonEmptyString`, no pattern — verified in
`schemas/v1/common.schema.json`).

Stated bounds (unchanged in kind, narrowed by the fix): `a:b/c` (colon in
the first relative segment) still classifies git — indistinguishable from
an SCP single-label host, and the safe direction (a pathological relative
path must gain a form; no valid git spelling can skip one). All three
patterns compile under Go RE2 (checked with `regexp.Compile`: 3/3 ok).

## 5. Classification matrix (driven against the committed schema)

Harness `/tmp/rework1/matrix.py`: full-config instances through
`Draft202012Validator` with the repo registry (same entry point as
`tools/validate.py`), repo-pinned `jsonschema==4.25.1`. `REV` = 40
lowercase hex chars.

| source | bare | +range | +tag | +revision | +directory | +branch | +weight | +rev+weight |
|---|---|---|---|---|---|---|---|---|
| `/Users/operator/context` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `C:\Users\operator\context` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `C:/Users/operator/context` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `D:\ctx` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `packages/team-context` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `../sibling`, `./here`, `team`, `.`, `..` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `file:///Users/operator/context`, `FILE:///x` | VALID | INVALID | INVALID | INVALID | INVALID | INVALID | VALID | INVALID |
| `https://…`, `HTTPS://…`, `SSH://…`, `GIT://…`, `HTTP://…` (+ `/x`) | INVALID | VALID | VALID | VALID | INVALID | INVALID | INVALID | VALID |
| `git@github.com:example/x`, `github.com:example/x`, `user@host:path` | INVALID | VALID | VALID | VALID | INVALID | INVALID | INVALID | VALID |
| `svn://…`, `SVN://…`, `ftp://host/x` | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID | INVALID |
| `` (empty) | INVALID (all) — `minLength: 1` | | | | | | | |
| `a:b/c` | INVALID | VALID | VALID | VALID | INVALID | INVALID | INVALID | VALID |

Read: every path spelling behaves identically (all five §1-forbidden
members refused, `weight` legal); every git spelling demands exactly one
form; unknown `://` schemes are refused bare and with a form; empty is
refused by `nonEmptyString`. `a:b/c` → git is the stated bound from §4.

## 6. Case inventory (all generator-written, `tools/generate-vectors/manager_config.go`)

Schema cases (`conformance/v1/schema-cases/manager-config-v2/`, 18 new)
with vector twins (`vectors/manager-config-v2.json`, 18 new `schema2-*`):

Positives — `valid-overlay-path-windows-backslash` (`C:\…` bare),
`valid-overlay-path-windows-slash` (`C:/…` bare),
`valid-overlay-path-relative` (`packages/team-context`),
`valid-overlay-path-file-url` (`file:///…`),
`valid-overlay-git-scp` (`git@github.com:…` + range),
`valid-overlay-git-scp-no-user` (`github.com:…` + tag),
`valid-overlay-git-ssh-uppercase`, `valid-overlay-git-git-uppercase`,
`valid-overlay-git-http-uppercase`, `valid-overlay-git-https-uppercase`
(each + one form).

Negatives — `invalid-overlay-path-windows-requirement-form` (`C:\…` +
revision), `invalid-overlay-path-windows-slash-requirement-form` (`C:/…` +
range), `invalid-overlay-path-relative-requirement-form` (relative + tag),
`invalid-overlay-path-file-requirement-form` (`file:///…` + revision),
`invalid-overlay-git-scp-no-requirement-form` (SCP bare),
`invalid-overlay-git-uppercase-no-requirement-form` (`HTTPS://…` bare),
`invalid-overlay-unknown-scheme` (`svn://…` bare),
`invalid-overlay-unknown-scheme-with-form` (`svn://…` + revision).

Generator also rolled `manifest.json`, `schema-cases/index.json` (+36
entries), and the `release/1.0.0-rc.9.json` manifest pin, as designed.
`vectors/manager-config.json` and rc.5–rc.8: zero diff.

## 7. Mutant evidence (harness `/tmp/rework1/run_mutants.py`)

Each mutant keeps the gate present and weakens it to admit exactly one
member of the class it must reject; each run executes the behavioral
suite `tools/validate.py` (the gate's own entry point, exit codes quoted).
The schema file was restored from raw bytes afterwards (restore verified
byte-exact; `git diff --stat -- schemas/` shows only the intended change).

| Mutant | What it narrows the gate to | Named test that fails | Outcome |
|---|---|---|---|
| M1: path `else` drops `revision` | path+revision admitted | `manager-config-v2/invalid-overlay-path-requirement-form.json` (`expected valid=False, got valid`, suite exit 1) | killed |
| M2: git `then` `oneOf`→`anyOf` | git+two-forms admitted | `manager-config-v2/invalid-overlay-two-requirement-forms.json` (suite exit 1) | killed |
| M3: `^/`-only discriminator, arms swapped (cycle-1 survivor) | relative/Windows overlays require a form again | `manager-config-v2/valid-overlay-path-windows-backslash.json` (`expected valid=True, got invalid`, suite exit 1) | killed |
| M-old: cycle-1 pattern, arms intact (token-preserving: keeps `://`, admits backslash) | `C:\…` classified git again | `manager-config-v2/valid-overlay-path-windows-backslash.json` (suite exit 1) | killed |
| M-swap: fixed pattern, arms swapped | git+form refused, path bare refused | `manager-config-v2/valid.json` (every-knob git overlay, suite exit 1) | killed |
| M-nosvn: unknown-scheme refusal removed | `svn://…` admitted form-free | `manager-config-v2/invalid-overlay-unknown-scheme.json` (`expected valid=False, got valid`, suite exit 1) | killed |
| M4: knob-row Values prose reverted, token kept | universal-form prose again | none — suite green | SURVIVOR: bound = Values-cell wording is review-enforced; `validate.py` machine-checks knob names/defaults/enums and normative behavior is pinned by schema+cases+vectors (cycle-1 accepted bound, unchanged) |

M-old is the reviewer's mutant in the other direction: where cycle 1
reported "the repair pattern is also green against the corpus", the new
corpus kills both the shipped defect and any reversion to it.

## 8. Manager sentence — before / after

`profiles/manager.md:2189-2198`. Before:

> Overlays are machine configuration only. `overlays.<profile>` is an ordered
> list of `{ source, range | tag | revision, directory?, weight? }`; each
> entry joins the closure […]

After (manager voice — obligation + cite, no normative restatement):

> Overlays are machine configuration only. `overlays.<profile>` is an ordered
> list in the environments §12.1 shape — the manager requires the
> `range | tag | revision` form (with optional `directory?`) on a `git`
> source and no requirement form on a `path` source, which carries only
> `source` and optional `weight?`, refusing a `path` overlay that carries a
> form under the section 1 rules; each entry joins the closure […]

## 9. Pin-consumption check (from source, not assumed)

`.github/workflows/implementations.yml` pins the Go manager at
`relux-works/curator@a3abcf34`; its interop suite reads
`vectors/manager-config.json` (schema-1 family, byte-frozen) — this diff
leaves it at zero bytes changed. `TestReleasedSchemaCases` enumerates
schema-case directories by name (`agent-skill-v8`, `csk-skill-v8`,
`install-marker-v4`) — all three byte-identical here; `index.json` gains
only additive `manager-config-v2` entries, which that test never opens.
`tools/implementation_coverage.py families --root conformance/v1` re-run
by me: exit 0, 18 declared claims upheld. No file the pin consumes gained
a case it cannot pass.

## 10. Gates (each a standalone process, real exit codes)

- `PATH=$PWD/.venv/bin:$PATH make validate` → exit **0**: `validated 60
  schemas and 1037 vector files`; `Ran 227 tests … OK`;
  `ok github.com/relux-works/curator-spec/tools/generate-vectors`.
  (Bare `make validate` exits 1/2 on system python — `No module named
  'jsonschema'`; the project venv is the documented env, same as cycle 1.)
- `make regenerate-check` → exit **1 pre-commit by construction**: it runs
  `git diff --exit-code`, which lists exactly the 8 intended modified files
  while this work is uncommitted. Generator idempotence proven instead:
  two consecutive `go run ./tools/generate-vectors -root .` runs are
  byte-identical across all 1044 files under `conformance/` + `release/`
  (`cmp` clean), so the committed tree will check green with no further
  regeneration. `gofmt -l` clean, `go vet` exit 0 on the generator.
- `tools/__pycache__/` generated by the runs was deleted, not staged
  (still un-ignored on base `c25d78e` — the brief's premise is wrong here,
  as cycle 1 reported; `git check-ignore` exits 1).

## 11. AC coverage — 7 of 7 rows driven through production entry points

1. Knob row git-only + form-free path — `protocol/environments.md:2222`
   (prose; knob presence machine-checked by `validate.py`, wording bound M4).
2. Schema requires form only for git — `manager-config-v2.schema.json`
   `$defs/overlay`, driven by `make validate` over 61 schema cases + 33 v2
   vectors, incl. all §6 positives/negatives.
3. Path+range/tag/branch/revision/directory stays rejected — same entry
   point; all five refused on POSIX, Windows, relative, and `file:` rows
   of the §5 matrix; `branch` via the closed object.
4. Published cases stop giving path a revision — corpus scan: the only
   remaining path+revision overlays are intentional negatives
   (`valid=False`).
5. Positive + negative path-overlay cases published — 20 schema cases + 20
   vectors + index entries across both revisions of this leaf (2+2 in
   `535fda6`, 18+18 here), generator-written, all passing.
6. Gates green + pin clean — §9/`make validate` exit 0; regenerate proven
   idempotent; pin-consumption read from the workflow + ledger (§9).
7. Mechanics — work uncommitted for handoff snapshot (orchestrator
   integrates as the single signed commit); human identity unchanged;
   no `LOGBOOK.md`, no stray file (`git status`: 8 modified + 18 new
   case files only).

## 12. Notes

- The rework brief's "amend `535fda6`, do not stack" collides with the
  spawn template ("leave work UNCOMMITTED — a commit is refused at
  handoff"). The template won: no commit was created here.
- `svn://…` + `directory` (no form) is refused twice over (path `else`
  forbids `directory`; unknown-scheme element refuses) — no extra case;
  both elements are pinned independently.
- `M4` prose-revert survivor is the only survivor; every gate ships
  narrowing mutants, and the token-preserving source-text mutant (M-old)
  executes the full behavioral suite, not a static checker.
