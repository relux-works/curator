# TASK-260905-369vye drafting report — environments 1.1 batch 3

Worktree `.temp/STORY-260905-2z9pw4/worktree`, branch
`task-board/story/STORY-260905-2z9pw4`, base `a68559b`. Two signed commits
(SSH, `git log --format='%G?'` = `U`, same validity as the base commits):

| Commit | Scope |
|---|---|
| `5855e98` | manager-config schema 2, schema cases, vectors, validator gate + tests, generator, rc.9 pin, manifest, COMPATIBILITY, CHANGELOG, READMEs |
| `ffbf803` | manager.md §1 note and §12 rewrite; cli/curator.md rows and examples |

No push, tag, or PR. Nothing written into the control root.

## Item → file/section

| Brief item | Where |
|---|---|
| manager §12 intro: schema-2 knob carriage, §1 discipline, §12.2 lockable extension | `profiles/manager.md` §12 (intro paragraphs 2), §1 paragraph 2 |
| §12.1 registry: MCP channel rows, reserved `curator-mcp` layer, `forms`, `xdg_seed_allowlist`, targets participation/consent, shadowing + `shadow_acknowledged`, tool releases, `mcp_package_allowlist` pointer | `profiles/manager.md` §12.1 |
| §12.2 modes/marker: marker contents per env §8.2, `in_place_mode`, claude_code always-copied, fresh homes + first-resolve notice, two doors, versioned backups + `backup_retention`, `env backups scrub`, drift | §12.2 |
| §12.3 lifecycle: install grammar (one root, `--directory`, `--as`, name-less `--use`, `--range latest`), overlays/weights/precedence knobs, `profile use` (shims, partial, `require_current_profile`), scoped `--clear`, `profile update` order, `profile remove [--purge]` + orphans, `env unmanage [--restore-backups]`, lock-writing skill scope, shim singleton, `default` migration, onboarding + import | §12.3 |
| §12.4 credential passthrough: `isolation` knob, unsupported directions, liveness row, seeds | §12.4 (extended — see inconsistencies) |
| §12.5 `env resolve`: tuple, `mcp` section, lock-free read-only, `--repair` + mutation lock + bounded wait, formats and digest, `passable_env_names`, `system_prompt_files.<profile>.pi`, `curator run` pointer | §12.5 |
| §12.6 audit: every member, detector scope (`agent-context.json`, `agent-mcp.json` args/url, `CONTEXT.md`), unpinnable, `secret_material_waivers`, `context-system-module-present`, `path` snapshot | §12.6 |
| §12.7 status/GC: `profile list` columns, `env status` rows per env §12, currency rule, warnings, GC roots incl. retained previous lock | §12.7 |
| CLI: install `--range\|--tag\|--revision` one root, `--use` no name; list columns; `use --clear`; `update`; `remove [--purge]`; `sync`; informative `profile compose` and `env config`; `env resolve [--repair]`; `env status` rows; `env unmanage [--restore-backups]`; `env backups scrub` | `cli/curator.md` Commands table |
| `curator run` provider pointer to Decision 0013 D6.4; examples (`--range '^1.2'`, `profile update`, `profile remove --purge`, `env resolve claude_code --repair --format json`, `--use`, `--clear`, `env unmanage`) | `cli/curator.md` § Environment profiles |
| `manager-config-v2.schema.json` | `schemas/v1/manager-config-v2.schema.json` (schema-1 members by `$ref` to the frozen v1 file; `schema_version` const 2; closed `environments`) |
| schema cases | `conformance/v1/schema-cases/manager-config-v2/`: `valid.json` (every knob non-default), `valid-minimal`, `valid-empty-environments`, `invalid.json`, `invalid-schema-version-1`, 7 closed-object negatives, 26 value-grammar negatives (38 files) via `tools/generate-vectors/manager_config.go` |
| vectors | `conformance/v1/vectors/manager-config.json`: 4 valid schema-2 cases with `expected.environments` = defaults ⊕ input, `schema1-rejects-environments`, 5 schema-2 negatives |
| validator + tests | `tools/validate.py` `validate_manager_config_vectors` (schema per `schema_version`, `valid` agreement, defaults ⊕ input, knob-name and literal-default cross-check against env §12.1 table, closed object, every knob defaulted, both versions covered) + `manager_config_semantic_error`; `tools/test_validate.py` `ManagerConfigVectorTests` (14 negatives); `tools/generate-vectors/manager_config_test.go` (2 tests) |
| README bullet, manifest, rc.9 pin | `conformance/README.md`, `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json` (regenerated), `schemas/v1/README.md` |
| COMPATIBILITY / CHANGELOG | `COMPATIBILITY.md` § Manager configuration schema 2; `CHANGELOG.md` Unreleased Added + Changed |

## Knob → schema property

All under `$defs.environments.properties` (`additionalProperties: false`).

| §12.1 knob | Property | Grammar | Default |
|---|---|---|---|
| `current_profile` | `current_profile` | identifier \| null | `null` |
| `scoped_current` | `scoped_current` | object identifier → identifier | `{}` |
| `overlays.<profile>` | `overlays` | object identifier → array of `$defs.overlay` (`source` required; exactly one of `range` (closed range chars) \| `tag` (gitRefName) \| `revision` (commit); `directory` portablePath; `weight` ≥ 0; closed) | `{}` |
| `overlay_default_weight` | `overlay_default_weight` | integer ≥ 0 | `1000` |
| `overlays_allowed` | `overlays_allowed` | boolean | `true` |
| `precedence.winner` | `precedence.winner` (`$defs.precedence`, closed) | enum `higher-weight`, `lower-weight` | `higher-weight` |
| `precedence.placement` | `precedence.placement` | enum `winner-last`, `winner-first` | `winner-last` |
| `forms.<env-id>` | `forms` | object identifier → enum `monolithic`, `referenced` | `{}` (adapter default) |
| `system_prompt_files.<profile>.pi` | `system_prompt_files` → `$defs.systemPromptFiles` (only `pi`, closed) | enum `off`, `append`, `replace` | `off` |
| `targets.<target-id>.participation` | `targets` → `$defs.target.participation` | enum `auto`, `off`, `enabled` | `auto` |
| `targets.<target-id>.consented` | `$defs.target.consented` | boolean | `false` |
| `isolation.<profile>.<env-id>` | `isolation` | object identifier → object identifier → enum `shared`, `isolated` | `{}` (`shared`; adapter-fixed for claude_code/macOS) |
| `xdg_seed_allowlist` | `xdg_seed_allowlist` | unique array of `$defs.xdgEntryName` (no separators, not `.`/`..`/`opencode`) | `["git", "gh", "ssh"]` |
| `passable_env_names` | `passable_env_names` | identifierSet \| null | `null` |
| `mcp_package_allowlist` | `mcp_package_allowlist` | unique array of non-empty strings | `[]` |
| `shadow_acknowledged` | `shadow_acknowledged` | array of `$defs.shadowAcknowledgement` `{ env, path }` (closed) | `[]` |
| `secret_material_waivers` | `secret_material_waivers` | array of `$defs.secretMaterialWaiver` `{ pin, file, span: [start, end], reason }` (pin = lowercase hex 40/64; closed) | `[]` |
| `backup_retention` | `backup_retention` | integer ≥ 0 | `5` |
| `require_current_profile` | `require_current_profile` | identifier \| null | `null` |
| `in_place_mode.<env-id>` | `in_place_mode` | object identifier → enum `linked`, `copied` | `{}` (adapter default) |

The validator cross-checks the property set against the table's first
segments and the twelve literal defaults against the schema; renaming a knob
or moving a default on either side fails `make validate` (tests
`test_schema_property_missing_from_the_table_fails`,
`test_table_knob_missing_from_the_schema_fails`,
`test_schema_default_drifting_from_the_table_fails`).

## Gate tails

`make validate` (`.temp/TASK-260905-369vye/validate-03.log`, exit 0):

```text
----------------------------------------------------------------------
Ran 166 tests in 27.813s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.592s
validate exit=0
```

`make regenerate-check` after the commits (`.temp/TASK-260905-369vye/regenerate-check-02.log`, exit 0):

```text
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
regenerate-check exit=0
```

Earlier runs: `validate-02.log` exit 2 (the validator's local-link scan found
the scratch copy of §12 under `.temp/`; scratch removed), and
`regenerate-check-01.log` exit 2 (uncommitted regenerated files, expected
before the commit; the diff was exactly the four generated files).

## Inconsistencies found, for follow-up (not edited)

1. Decision 0012 compatibility table row "manager §12.4 | unchanged | —" —
   but environments §12.1 places `isolation.<profile>.<env-id>` under §7.4
   and §12.2 makes it lockable, and §7.4 adds the liveness row and seeds; the
   brief's knob rule required §12.4 to state those obligations, so §12.4 was
   extended (one paragraph and a three-row table). The impact row should read
   *bytes change*.
2. environments §12.1: "`secret_material_waivers` | list of `{ pin, file,
   span: [start, end], reason }`" — the `pin` spelling (commit vs. state
   hash, prefixed or bare) is not stated; the schema accepts bare lowercase hex
   of 40 or 64 characters, matching the marker's `commit`/`state_sha256`
   values. §12.1 should say which.
3. environments §12.2: "The manager §1 `locked` set is extended … by exactly
   these keys under `environments`" — `system-config-v1.schema.json` closes
   `locked` to four schema-1 keys and has no `environments` member, so a
   system file cannot lock any §12.2 key today; a `system-config-v2` is a
   separate batch (noted in COMPATIBILITY).
4. Schema-1 vectors `insecure-registry` and `duplicate-canonical-registry`
   are `valid: false` by the manager §1 registry semantics, not by
   `manager-config-v1.schema.json` (whose URL pattern admits `http`); they
   were never validated before. `validate.py` now encodes those two rules in
   `manager_config_semantic_error` so the vector file is checked both ways.
5. environments §7.7 has no row for `environment_form_unavailable` (it lives
   only in §5.7); manager §12.1 cites §5.7 for it.

## Not run

Nothing was skipped. The venv was created fresh at `.temp/venv`
(`jsonschema==4.25.1`); Go 1.25.5, Python 3.14.7.
