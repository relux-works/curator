# Evidence — TASK-260916-1hrx51 spec-system-module-admission-rule (E2) — revision 2 (rework)

Story `STORY-260916-2d9coh` (direct-only-system-modules), wave 1.
Role: doc-writer. Worktree branch `task-board/story/STORY-260916-2d9coh`
(`curator-spec/.temp/STORY-260916-2d9coh/worktree`, forked from `main` `07e2b41`).

This revision answers `TASK-260916-1hrx51_review-verdict-rev1.md`
(changes_requested, routed to-dev). Everything the verdict marks
present/passing stays exactly as it is: admission semantics, knobs,
vectors, generator, schema cases. Rev2 touches exactly two files —
`profiles/manager.md` (R1) and `protocol/environments.md` (R2) — plus this
evidence (count fix) and the refreshed spec patch.

## R1 — manager §1 closed lock set (closed)

`profiles/manager.md:56-58` — the closed enumeration now reads:

```text
   `environments.require_current_profile`, `environments.isolation`, and
   `environments.transitive_system_modules`.
   No other `environments` knob is lockable or carried by the system file.
```

`profiles/manager.md:61-65` — the direction restriction, stated where the
other directional key states its own:

```text
   a system file MUST NOT select or constrain it; `environments.isolation`
   is lockable only in the direction of `shared`, and schema 2 admits no
   other value there; `environments.transitive_system_modules` is lockable
   only in the direction of `error`, and schema 2 admits no other value
   there;
```

Spelling `environments.transitive_system_modules` is identical to
`protocol/environments.md` §12.2 (`:2344`) and
`schemas/v1/system-config-v2.schema.json` (`:23,51-53`). Nothing else in
the set changed; `system_module_waivers` stays excluded by the "No other
`environments` knob is lockable" closure, matching §12.2 and the schema
(which deliberately omits it).

## R2 — RFC 2119 implementer obligations (closed)

`protocol/environments.md:510-535` — the admission block now states:

- `:516-518`: "The section 5.5 system-prompt output and the section 10.2
  fragment `system_prompt` section MUST contain only admitted system
  modules."
- `:521-525` (drop): "a non-admitted applicable system module MUST be
  skipped at materialization, and the manager MUST emit the warning
  `context_system_module_dropped` naming the package and the module path;
  the materialized bytes MUST be exactly the admitted modules' bytes.
  Resolution, installation, and update MUST NOT fail for admission."
- `:526-530` (error): "resolution of the same module MUST fail with the
  resolution error `context_system_module_transitive` naming the package
  and the module path — the first such module in emitted order, manifest
  order within a package — and the manager MUST NOT write or change the
  lock: `profile install` MUST fail, and `profile update` MUST leave the
  old lock in place."
- `:532-535`: the `context-system-module-present` finding "MUST report
  every `class: system` module of every member at install and update,
  admitted or not".

`protocol/environments.md:2352-2356` — the direction rule now states: "a
system file MUST lock `transitive_system_modules` only to `error`, and
`system_module_waivers` MUST NOT be lockable — a lock MUST NOT admit a
transitive package's system modules." Settled behavior is unchanged; only
the modality changed from descriptive to normative.

## Evidence fix (schema-case count)

The rev1 text below said "eleven new schema-case files" in two places.
Correct count, verified from the worktree: **seven** new schema-case files
plus **four** new expected byte directories (the error case writes no
file, by design). The seven:

- `conformance/v1/schema-cases/manager-config-v2/valid-system-module-waiver.json`
- `conformance/v1/schema-cases/manager-config-v2/invalid-transitive-system-modules-value.json`
- `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-package-grammar.json`
- `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-missing-reason.json`
- `conformance/v1/schema-cases/manager-config-v2/invalid-system-module-waiver-unknown-field.json`
- `conformance/v1/schema-cases/system-config-v2/invalid-transitive-system-modules-drop-direction.json`
- `conformance/v1/schema-cases/system-config-v2/invalid-transitive-system-modules-value.json`

The four expected dirs (`system-module-direct`, `system-module-overlay-direct`,
`system-module-transitive-drop`, `system-module-transitive-waived`) are
unchanged from rev1. The carried-over rev1 text below is corrected to
"seven new schema-case files" in both places.

## Validation transcript (rev2, rerun by the producer)

Shell `/bin/zsh` with `set -o pipefail`, workdir the Story worktree above.
Python via the worktree venv (`$PWD/.venv/bin`, `jsonschema==4.25.1`).

1. `PATH="$PWD/.venv/bin:$PATH" make validate` → **EXIT=0**:

```text
python3 tools/validate.py
validated 60 schemas and 1058 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...........................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 235 tests in 327.264s

OK
go test ./tools/...
ok  github.com/relux-works/curator-spec/tools/generate-vectors 2.298s
EXIT=0
```

2. `git diff --check origin/main -- profiles/manager.md protocol/environments.md`
   → exit 0 (no whitespace errors in the rev2 delta).
3. `make regenerate-check` in the live worktree → **EXIT=2** (make wrapper;
   inner `git diff --exit-code` exits 1). Expected-red, reported honestly:
   the worktree carries the uncommitted candidate changes, so any
   working-tree-vs-HEAD diff is red by construction; this is not a
   generation-drift signal and is never presented as passing.
4. Disposable-copy `make regenerate-check` → **EXIT=0**. Method (as in the
   S6/review procedure): byte-copied all tracked plus untracked nonignored
   candidate files to `/tmp/regen-rev2`, established an isolated temporary
   git baseline there (no commit in either project worktree), then ran the
   unmodified gate:

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
EXIT=0
```

Zero generated drift from the md-only rev2 delta.

Patch identity: `git add -N . && git diff origin/main` →
`TASK-260916-1hrx51_spec-patch_rev2.patch`, 223600 bytes,
`git patch-id --stable` `b9bff40b930a859b540f85b6a02e2fcec4796805`
(file and live-diff patch-ids agree).

## Rev2 scope

Only `profiles/manager.md` and `protocol/environments.md` changed versus
rev1. `git status --short` is otherwise identical: 117 paths — the same
modified set plus the same 11 untracked paths (7 schema cases + 4 expected
dirs). No generator, schema, vector, manifest, or CHANGELOG changes; no
implementation code touched (the only `.go`/`.py` edits in the candidate
remain the rev1 spec-conformance tooling the Makefile prescribes).

## Checklist mapping (rev2)

- Item 1 (normative rule + RFC 2119 + closed sets): satisfied — R1 closes
  the manager §1 enumeration, R2 puts MUST/MUST NOT on every admission and
  direction obligation; spellings identical across prose, tables, schema,
  vectors, enforced by `make validate`.
- Items 2, 3, 4, 7, 11 (rollout + posture row; vectors + gate; CHANGELOG +
  attachments + no product code; outcome resource; tests green): remain
  satisfied, gate rerun green in rev2.
- Item 5 (docs consistent): satisfied — manager §1 mirror is now
  consistent with environments §12.2 and the system schema.
- Item 6 (no code/description gap): satisfied — md-only delta, rev1
  generator/validator byte agreement unchanged.
- Items 9, 10 (implementation matches AC / fits architecture): left
  unchecked — landing is the orchestrator's step and review acceptance is
  pending.
- Item 8 (logbook): the board note carries the durable record; campaign
  rules forbid `LOGBOOK.md` edits.
- Item 12 (verdict routing): satisfied by the rev1 cycle (verdict attached,
  routed to-dev, reworked here).

---

# Revision 1 record (carried over, schema-case counts corrected)

Story `STORY-260916-2d9coh` (direct-only-system-modules), wave 1.
Role: doc-writer. Worktree branch `task-board/story/STORY-260916-2d9coh`.

Finding E2 (`docs/security-audit-2026-09.md` §E2, confirmed by Appendix B):
every closure package's `class: system` modules materialize; the only
control is the always-warn finding `context-system-module-present`.

## What changed per file, and why

- `protocol/environments.md`
  - §3: new normative **System-module admission** rule. `direct` = the root
    itself, an active overlay, or a package named by the root's or an
    active overlay's `requires.contexts` entry; everything else is
    `transitive`. Only direct packages plus `system_module_waivers`-admitted
    packages are **admitted**. `transitive_system_modules` = exactly `drop`
    (default) or `error`: `drop` skips at materialization with
    `context_system_module_dropped` (bytes are exactly the admitted
    modules'; resolution/install/update never fail); `error` fails
    resolution with `context_system_module_transitive` (first module in
    emitted order, manifest order within a package; no lock written/changed;
    install fails, update keeps the old lock). `context-system-module-present`
    is explicitly unaffected (reports every member's modules, admitted or not).
  - §3.1: row for `context_system_module_transitive`.
  - §5.5: admitted-set wording; drop behavior; one-sentence fragment-flag
    interaction (the `system_prompt` presence keeps driving Decision 0013
    `works.relux.curator.system-modules`, so `ax` resume still refuses on
    drift); rollout paragraph (E2): default `drop` is non-breaking, so no
    warn-first split applies; `error` is opt-in strictness.
  - §5.7: rows for `context_system_module_dropped` (warning) and
    `context_system_module_transitive`.
  - §10.2: `system_prompt` presence now requires an **admitted** module.
  - §12: `env status` posture row (effective policy + every dropped module
    by package and path); `context_system_module_dropped` joins the
    never-non-current warnings.
  - §12.1: knob rows `transitive_system_modules` (`drop`, `error`, default
    `drop`) and `system_module_waivers` (list of `{ package, reason }`,
    default empty); waiver-entry prose (`package` = core §2 identifier
    naming a lock member; unknown package = no effect).
  - §12.2: `transitive_system_modules` joins the lockable keys; direction
    rule (`error` only); `system_module_waivers` is not lockable.
  - §13: admission vector family + knob schema cases listed.
  - §2 needs no change: it defines the package shape, not module admission
    (considered, deliberately untouched).
- `profiles/manager.md` (§12.7 mirror only): status-row list and warnings
  list gain the same drop row/warning. No other manager text changes: audit
  classes are unchanged and the new knobs are machine-global, not
  per-profile records. (Rev2 adds the §1 closed lock-set entry — see R1.)
- `schemas/v1/manager-config-v2.schema.json`: `transitive_system_modules`
  (enum `drop`/`error`, default `drop`), `system_module_waivers` (array of
  closed `$defs/systemModuleWaiver` `{ package, reason }`, default `[]`).
- `schemas/v1/system-config-v2.schema.json`: `transitive_system_modules`
  joins the lockable set and `locked` enum; admits `error` alone (mirrors
  the `isolation`/`shared` direction pattern); description updated.
  `system_module_waivers` is deliberately absent (not lockable).
- `tools/generate-vectors/environments.go`: `sysroot`/`sysmid`/`sysleaf`/
  `sysovl` fixtures (direct chain + overlay edge); `systemModulePolicy`,
  `environmentDirectPackages`, `environmentAdmitSystemModules`,
  `environmentSystemPromptFilesWithPolicy` (default-policy wrapper keeps the
  old signature); five `system-module-*` materialization cases emitting
  `machine_policy`, `admitted`, `dropped`, `warnings`, and refusal fields.
- `tools/generate-vectors/manager_config.go` + `system_config.go`:
  defaults, every-knob fixtures, and schema examples for both knobs
  (including the system-config drop-direction negative).
- `tools/validate.py`: independent Python re-implementation of the §3
  admission split; byte-recompute of admitted-only output; admitted/dropped/
  warnings/refusal record checks; `transitive_system_modules` default+enum
  cross-checks against §12.1; system-config error-alone check.
- `tools/generate-vectors/*_test.go`, `tools/test_validate.py`: knob lists
  extended (20 manager knobs, 7 lockable keys); direction/closed-set
  negatives; admission positives/negatives incl. a byte-exactness test
  (`system-module-direct` ≡ `system-module-transitive-drop` bytes).
- `conformance/v1/`: five new `system-module-*` materialization cases with
  four new `expected/environments/system-module-*/` byte files (the error
  case writes no file, by design), seven new schema-case files, refreshed
  embeddings, `manifest.json`, and the rc.9 release pin — all via
  `make regenerate`.
- `CHANGELOG.md`: Unreleased entry "E2 direct-only `class: system`
  modules" naming the finding, the rule, both diagnostics, both knobs, the
  no-warn-first-split rollout, the status row, and the vector coverage.
- `release/1.0.0-rc.9.json`: manifest pin refresh from regeneration.

## Deliberately out of scope

Implementation (`TASK-260916-55g9dg`); E1 signer rules; pi `SYSTEM.md`
channel semantics (only cited); marker-schema changes; CLI rows (no knob
rows exist under `cli/`); `path`-kind admission (E6 story).

Note on "no implementation code touched": the only `.go`/`.py` edits are
the spec repository's own conformance tooling (vector generator, schema
examples, and the independent validator) — the delivery mechanism the
Makefile prescribes for vectors (`make regenerate`) and the gate
(`make validate`). No manager/launcher product code exists in this
repository and none was touched. Precedent: `bd39adb`, `fcdb9ba`.

## Validation transcript (rev1)

Shell: `/bin/sh` via the agent harness, workdir
`curator-spec/.temp/STORY-260916-2d9coh/worktree`. Python deps were absent
system-wide (`python3 tools/validate.py` failed with
`ModuleNotFoundError: No module named 'jsonschema'`, exit 2 via make), so a
project venv was created per `requirements-dev.txt` and used for every
Python gate:

- `uv venv .venv && uv pip install -r requirements-dev.txt --python
  .venv/bin/python` → exit 0 (`jsonschema==4.25.1`,
  `jsonschema-specifications==2025.9.1`, `referencing==0.37.0`). The
  `.venv/` directory is git-ignored and left in the worktree so the
  reviewer can rerun the gates verbatim.
- `gofmt -l tools/` → empty (clean), exit 0.
- `go test ./tools/...` → `ok .../tools/generate-vectors`, exit 0
  (includes the new `TestEnvironmentSystemModuleAdmission`).
- `make regenerate` (`go run ./tools/generate-vectors -root .`) → exit 0;
  five `system-module-*` cases, four new expected byte files (the error
  case writes no file, by design), seven new schema-case files, refreshed
  `manifest.json`, `schema-cases/index.json`, and `release/1.0.0-rc.9.json`.
- `.venv/bin/python tools/validate.py` → `validated 60 schemas and 1058
  vector files`, exit 0.
- `.venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'` →
  `Ran 235 tests ... OK`, exit 0 (includes the new admission, direction,
  and byte-exactness tests).
- `PATH="$PWD/.venv/bin:$PATH" make validate` (the repository gate
  unmodified: `validate.py` + `unittest discover` + `go test ./tools/...`)
  → all three sections green, `EXIT=0`. Quoted tail:

  ```text
  python3 tools/validate.py
  validated 60 schemas and 1058 vector files
  python3 -B -m unittest discover -s tools -p 'test_*.py'
  Ran 235 tests in 327.531s

  OK
  go test ./tools/...
  ok  github.com/relux-works/curator-spec/tools/generate-vectors 1.956s
  EXIT=0
  ```

Independent byte-exactness proof (outside both generators): `cmp
conformance/v1/expected/environments/system-module-direct/.agent-context/system-prompt.md
conformance/v1/expected/environments/system-module-transitive-drop/.agent-context/system-prompt.md`
→ identical (exit 0); both record `surface_sha256`
`sha256:a48ba73f…` and file sha `sha256:537ebd5e…` in the vector.

Checklist mapping: item 1 (rule + closed sets) — §3/§3.1/§5.5/§5.7/§10.2/
§12/§12.1/§12.2/§13 + schemas + vectors share exact spellings, enforced by
`make validate`; item 2 (rollout) — §5.5 "Rollout (E2)" paragraph states
the default-`drop` non-breaking rule with no warn-first split and labels
`error` opt-in, plus the §12 posture row; item 3 (vectors + gate) —
above; item 4 (CHANGELOG + attachments + no product code) — entry present,
patch + this file attached, only spec conformance tooling edited (see
note); item 5 (docs consistent) — manager §12.7 mirror updated, §2
assessed as needing no change; item 6 (no code/description gap) —
generator and validator implement the same admission independently and
agree byte-for-byte; item 7 (outcome resource) — this file + the patch.
Item 8 is left unchecked: campaign rules forbid `LOGBOOK.md` edits, and
there is no finding, anomaly, or regression beyond what this file records.
