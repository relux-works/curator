# Drafting report — TASK-260906-3x0w4y: path-overlay-cannot-be-declared

Commit `535fda6` on `task-board/story/STORY-260905-2z9pw4` (one signed commit
past the story base `c25d78e`; see lineage note at the end).

## 1. The contradiction (verified at 550579d, unchanged by this task)

- §6: an overlay is "an ordinary context package named by a `git` source with
  a range or exact form, **or by a `path` source under the section 1 rules**".
- §1, `path`: "A `path` declaration that carries `range`, `tag`, `branch`,
  `revision`, or `directory` is `profile_source_invalid`" and "A `path`
  package serves as a profile root **or as an overlay (section 6)**."
- §12.1 and `manager-config-v2` `$defs/overlay` required a form on *every*
  overlay, so no `path` overlay was declarable. §§1/6 untouched (authority).

## 2. §12.1 knob row — before / after

Before (`protocol/environments.md:2222`):

```
| `overlays.<profile>` | ordered list of `{ source, range \| tag \| revision, directory?, weight? }` | empty | 6 |
```

After (one-line table shape kept; `directory?` / `weight?` spellings kept;
nothing else in §12.1/§12.2 touched):

```
| `overlays.<profile>` | ordered list of `{ source, range \| tag \| revision, directory?, weight? }` for a `git` source, `{ source, weight? }` for a `path` source | empty | 6 |
```

## 3. `$defs/overlay` — before / after

Before (`schemas/v1/manager-config-v2.schema.json`): `required: [source]`
plus an unconditional
`oneOf: [{required:[range]}, {required:[tag]}, {required:[revision]}]`.

After: same `required`/`properties`/`additionalProperties: false`, with the
`oneOf` moved into a conditional (full text in the commit):

```json
"allOf": [{
  "if": { "properties": { "source": { "pattern": "^...git...$" } } },
  "then": { "oneOf": [{"required":["range"]}, {"required":["tag"]}, {"required":["revision"]}] },
  "else": { "not": { "anyOf": [{"required":["range"]}, {"required":["tag"]},
                               {"required":["revision"]}, {"required":["directory"]}] } }
}]
```

where the pattern is

```
^(?:(?:[Ss][Ss][Hh]|[Gg][Ii][Tt]|[Hh][Tt][Tt][Pp][Ss]?)://|(?:[^/:\s@]+@)?[^/:\s]+:[^/\s])
```

`branch` needs no entry in the `else`: it is not a known property, so the
closed object (`additionalProperties: false`) already rejects it on *any*
overlay — matching §1's "branch on any declaration". `weight` stays admitted
on both arms (§6: every overlay declaration carries a weight).

## 4. Discriminator decision (from §1's own words)

§1 gives the schema a spelling distinction, not a `kind` field: `git` is "a
**network git source under the core §6.1 canonical identity**", `path` is
"an operator-local package directory named by **an absolute path, or by a
project-relative path**". The pattern encodes exactly the core §6.1 network
spellings §1 incorporates by reference — the four schemes (`ssh`, `git`,
`http`, `https`, any letter case, ECMA/RE2-safe classes, no inline flags)
and SCP `[user@]host:path` — and treats every other spelling as a path.
Nothing invented: no new field, no new enum, no new grammar.

Verified classification (throwaway probe, `re.search`): all of
`https://…`, `http://…`, `ssh://…`, `git://…`, `HTTPS://…`,
`git@github.com:o/r`, `github.com:o/r` → git; all of
`/Users/operator/context`, `/x`, `packages/team`, `../x`, `./x`, `x`, `.`,
`file:///x` → path; 0 mismatches.

Stated bounds of the spelling rule (§1 resolution stays authoritative via
`profile_source_invalid`):
- `C:\foo` (drive-letter) and `a:b/c` (colon in first relative segment)
  classify as git. The spec's path spellings are POSIX (`/Users/…`; the
  `portablePath` grammar itself bans `:`), so these are pathological.
- A `file:` URL classifies as path (core §6.1: no network identity); the
  schema admits it form-free and resolution decides.
- No committed case pins the SCP arm or a project-relative spelling (M3
  survivor, §7).

## 5. Case inventory

Changed by the generator only (`tools/generate-vectors/manager_config.go`):
- `valid.json` + every `manager-config-v2` schema case embedding the
  every-knob fixture + `vectors/manager-config-v2.json` (`schema2-every-knob`
  and one more vector): third overlay `/Users/operator/context` drops its
  `revision` → form-free. That is the whole "stop giving a path source a
  revision" change; every other byte of those files is generator output.
- Added positive: `schema-cases/manager-config-v2/valid-overlay-path-source.json`
  (`{"source": "/Users/operator/context"}`) and vector
  `schema2-overlay-path-source` (valid, with `expected.environments`).
- Added negative: `schema-cases/manager-config-v2/invalid-overlay-path-requirement-form.json`
  (path source + the formerly published `revision`) and vector
  `schema2-overlay-path-requirement-form` (valid=false).
- `vectors/manager-config.json`: untouched (byte-frozen schema-1 family;
  `validate.py` enforces schema_version 1 only there).
- Generator also rolled `manifest.json`, `schema-cases/index.json` (+2
  entries), and the `release/1.0.0-rc.9.json` manifest pin, as designed
  (rc.5–rc.8 untouched).

## 6. Pin-consumption check (how verified)

`.github/ci/implementation-coverage.tsv` (59 lines) names **no**
manager-config or environments artefact (`grep` exit 1). The pin consumes
`schema-cases/{index.json, agent-skill-v8, csk-skill-v8, install-marker-v4}`
and `vectors/{module-roots, script-host-execution-policy}.json`. Changed
files in this commit: the four hand-edited sources (§2/§3/CHANGELOG +
generator) plus generator output (`manifest.json`, `schema-cases/index.json`,
31× `schema-cases/manager-config-v2/*`, `vectors/manager-config-v2.json`,
`release/1.0.0-rc.9.json`). Of the pin-consumed set only `index.json`
changed — two *added* entries, both `manager-config-v2.schema.json`, none
removed — and the v8 families the pin's `TestReleasedSchemaCases` consumes
"in full" are byte-identical. No file the pin consumes gained a case the pin
cannot pass.

## 7. Gates (each a standalone process, real exit codes)

- `make validate` → exit **0**: `validated 60 schemas and 1019 vector
  files`; `Ran 227 tests … OK`; `go test ./tools/... ok`.
- `make regenerate-check` on the committed tree → exit **0** (empty diff).
  (Pre-commit it exits 1 by construction — it diffs the worktree.)
- Env: repo-pinned `jsonschema==4.25.1` in project `.venv` (per python-env
  skill); Go toolchain as installed.

## 8. Mutant evidence (harness `/tmp/mutant_probe.py`, throwaway, full
`tools/validate.py` = the behavioral suite per mutant, files restored after)

| Mutant | What it narrows the gate to | Named test that fails | Outcome |
|---|---|---|---|
| M1: path `else` drops `revision` from the forbidden set | path+revision admitted, path+range/tag/directory still rejected | `manager-config-v2/invalid-overlay-path-requirement-form.json` (`expected valid=False, got valid`, gate exit 1) | killed |
| M2: git `then` `oneOf`→`anyOf` | git+two-forms admitted | `manager-config-v2/invalid-overlay-two-requirement-forms.json` (gate exit 1) | killed |
| M3: Shape-1 discriminator (`^/`-only path, arms swapped) | relative-path overlay requires a form again | none — suite green (exit 0) | **survivor**: bound = no committed case uses a project-relative spelling; that arm is pinned by pattern text + §1 prose only |
| M4: knob-row Values prose reverted, `` `overlays.<profile>` `` token kept | prose mandates universal form again | none — suite green (exit 0) | **survivor**: bound = the Values-cell wording is review-enforced; `validate.py` machine-checks knob names/defaults/enums, normative behavior is pinned by schema+cases+vectors |

Direct matrix through the committed schema (full-config instances,
`Draft202012Validator`): path bare→valid; path+range/tag/revision/
directory/branch→INVALID (all five §1 members); path+weight→valid;
git+range→valid; git bare→INVALID; relative bare→valid.

## 9. AC coverage — 7 of 7 rows driven through production entry points

1. Knob row git-only + form-free path — `protocol/environments.md:2222`
   (prose; machine-checked for knob presence by `validate.py`
   table↔schema comparison, wording bound per M4).
2. Schema requires form only for git — `manager-config-v2.schema.json`
   `$defs/overlay`, driven by `make validate` over 40 schema cases + 15
   v2 vectors (incl. `valid-overlay-path-source`,
   `schema2-overlay-path-source`, `invalid-overlay-no-requirement-form`).
3. Path+range/tag/branch/revision/directory stays rejected — same schema
   entry point; revision via committed negative, all five observed INVALID
   in the §8 matrix (`branch` via the closed object,
   `invalid-unknown-overlay-field`).
4. Published cases stop giving path a revision — corpus scan: the only
   remaining path+revision overlays are the two intentional negatives
   (`valid=False`).
5. +positive/−negative cases published — 2 schema cases + 2 vectors +
   index entries, all generator-written, all passing `make validate`.
6. `make validate` exit 0; `make regenerate-check` exit 0 (post-commit).
7. Pin-consumed files gain no unpassable case — §6 evidence.

## 10. Notes / follow-ups (not changed — out of scope or observed)

- `profiles/manager.md:2189-2191` restates the old universal shape
  ("`{ source, range | tag | revision, directory?, weight? }`"). Left
  untouched per the brief's file scope; needs a conforming follow-up.
- Brief says `tools/__pycache__` "is now ignored" — it is not
  (`git check-ignore` empty, status showed `?? tools/__pycache__/`). Removed
  before commit instead; no ignore rule added (out of scope).
- Lineage: story branch already carried `c25d78e` past `origin/main`
  (`550579d`); this task adds exactly one signed commit (`535fda6`,
  `Ivan Oparin <oparin@me.com>`, Good signature) past the story base.
  "Past current main" was not achievable without a forbidden rebase.
- `.venv/` (project env, pinned deps) is invisible to git here and was left
  in place, uncommitted.
