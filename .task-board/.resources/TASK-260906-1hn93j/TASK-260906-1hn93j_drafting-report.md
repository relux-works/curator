# Drafting report — TASK-260906-1hn93j: cli rows for takeover and import

Base: `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`. Files changed: `cli/curator.md`
(two table rows, two example lines), `CHANGELOG.md` (`## Unreleased` entry,
following the f61ee9a convention of documenting CLI-surface additions).
No normative rule added or changed; `protocol/environments.md`, schemas,
vectors, and the manager profile are untouched.

## Row → sourcing table

Every clause below quotes the `environments.md` sentence behind it.
Flag/subcommand spellings marked CHOICE are family-consistent spellings the
spec does not pin; each is also recorded as an open question below.

### Takeover row

Published: ``curator env takeover --takeover [--env <env-id>] [--target <target-id>]`` —
"Take over a specific unmanaged file outside onboarding under the explicit
takeover flag, with the same notice and section 8.3 backup generation as
onboarding; without the flag the operation fails with
`environment_surface_unmanaged_conflict` rather than overwrite"

| Row clause | Source sentence (environments.md) |
|---|---|
| Take over a specific unmanaged file outside onboarding | §9.5: "Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag and performs the same notice and backup; without the flag, section 8.3 applies and the operation fails rather than overwrite." |
| under the explicit takeover flag | Same §9.5 sentence ("requires the explicit takeover flag"). Flag spelling `--takeover`: CHOICE — see Q1. Bare-boolean shape follows the published `--purge`, `--repair`, `--use` family. |
| `curator env` placement, next to `env unmanage` | Producer-brief-directed (inverse of unmanage on the same surfaces). Trigger status sourced at §9.5: "Onboarding is triggered only by a **mutating** profile operation that meets unmanaged state — `profile install`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`, and an explicit takeover." Standalone-vs-modifier shape: no pinning sentence — see Q5. |
| with the same notice ... as onboarding | Same §9.5 sentence ("performs the same notice and backup"); notice content at §9.5 step 2: "before any write, the operator is told that native global context files are being replaced by managed ones and where the backup lands." |
| ... and section 8.3 backup generation ... | §9.5 step 3: "every file the operation will replace is copied into the next section 8.3 backup generation before the first write, whether or not any import was requested, subject to `environment_backup_exists`." Generations defined at §8.3: "Takeover and onboarding backups (section 9.5) land in **versioned** backup sets `.agent-environment-backup/<n>/` beside the marker..." |
| without the flag the operation fails with `environment_surface_unmanaged_conflict` rather than overwrite | §9.5 ("without the flag, section 8.3 applies and the operation fails rather than overwrite") plus §8.3: "a manager MUST remove or replace only files its preceding marker records and MUST fail with `environment_surface_unmanaged_conflict` rather than overwrite an unmanaged file." |
| `[--env <env-id>] [--target <target-id>]` scope flags | Spellings identical to the published `env unmanage` / `profile use` rows; scope-narrowing meaning at §9.3: "`profile use` accepts `--env <env-id>` and `--target <target-id>` to narrow the switch to a subset of registered adapters or to one secondary fixed-home target." That takeover scopes this way (rather than naming a path) is a CHOICE — see Q2. The row states no default scope (unmanage's "(default: every scope)" has no takeover counterpart in the spec). |
| Example line `curator env takeover --takeover --env claude_code` | Mirrors the existing `curator env unmanage --restore-backups --env pi` example; `claude_code` is a registry env id (§7.1/§7.9 tables). |

### Import row

Published: ``curator profile import [--as <name>] [--allow-lossy]`` —
"Reassemble the section 9.5 inventory into a context-package-shaped directory
and install it through the ordinary `path` pipeline; the profile is named
`imported` unless `--as` supplies a name; a lossy import stops with
`environment_import_lossy` and the loss list unless the per-operation consent
flag re-reports the list as warnings, and machine configuration never
pre-records consent"

| Row clause | Source sentence (environments.md) |
|---|---|
| Reassemble the section 9.5 inventory into a context-package-shaped directory | §9.6: "Its input is the section 9.5 inventory; its output is one installed, audited, locked profile whose environment markers record `imported_from_native`." plus "The manager assembles a context-package-shaped directory inside the machine home (physical location implementation-specific, manager §1):" |
| and install it through the ordinary `path` pipeline | §9.6: "The assembled directory then installs through section 9.1 exactly as an operator-supplied `path` source — snapshot copy, state-hash pin, resolution of the pinned skills, always-strict audit; a blocking finding, `context-secret-material` included, fails the import like any install." |
| `curator profile` placement | Output is "one installed, audited, locked profile" (§9.6, quoted above). Subcommand name `import`: CHOICE — see Q4. |
| the profile is named `imported` unless `--as` supplies a name | §9.6: "`agent-context.json`, `schema_version` 1, `name` `imported` unless the operator supplies a name under the core §2 grammar, `version` `1.0.0`, `weight` `0`, and no `weights`." The `--as <name>` carrier reuses the §9.1 spelling ("the profile name is the root package's `name` unless `--as <name>` is given"), already published on the install row: CHOICE — see Q4. |
| a lossy import stops with `environment_import_lossy` and the loss list | §9.6: "A lossy import stops with `environment_import_lossy` and the loss list; it proceeds only under an explicit per-operation consent flag, which re-reports the loss list as warnings under the same diagnostic." Loss-list content: "The **loss list** names each loss — adapter, platform path, and reason — ..." (§9.6). |
| unless the per-operation consent flag re-reports the list as warnings | Same §9.6 consent-gate sentence. Flag spelling `--allow-lossy`: CHOICE — see Q3. Verb-led shape follows the `--restore-backups` / `--allow <hash>` family. |
| machine configuration never pre-records consent | §9.6: "Machine configuration MUST NOT pre-record consent." |
| operator-requested operation (justifies a standalone row) | §9.5 step 4: "the import itself runs only on the operator's request and under the section 9.6 consent rules." |
| Example line `curator profile import --as legacy --allow-lossy` | Carries both new flags in one line, mirroring the existing `curator profile install ... --tag v1.2.0 --as companyA` example shape. |

### Deliberately not published (present in §9.5/§9.6, out of surface-index scope)

- Lossless-vs-lossy classification rules, the closed detected-surface list,
  reassembly normalization, `profile_import_name_taken`, and the
  `environment_import_skill_foreign` warning: normative protocol behavior,
  not operator-surface spelling; the rows point at the sections instead.
- The onboarding trigger list (`profile install/use/sync/update`,
  `env resolve --repair`): those rows already exist and already behave per
  §9.5; no edit needed.

## Gate output tails (exit codes observed directly, no pipes through tee)

`make validate` — exit 0:

```text
python3 tools/validate.py
validated 60 schemas and 1017 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
...................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 227 tests in 58.775s

OK
go test ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	0.938s
```

`make regenerate-check` — exit 0, no diff output (byte-clean).

## Verification bounds (honest statement)

- This batch touches prose only (`cli/curator.md`, `CHANGELOG.md`).
  `tools/validate.py` has no coverage of `cli/curator.md` (verified: no
  reference to `curator.md` or `CHANGELOG` anywhere under `tools/`), so no
  committed test asserts row text — such a test would restate the prose and
  prove nothing about the gate it narrows.
- No production code path changed; there is no executable entry point for
  these rows yet (the stage-(c) Go implementation is the consumer blocked on
  this task). Negative/narrowing-mutant evidence is therefore not applicable
  to this batch: there is no gating, refusing, or attesting code to attack.
  AC coverage is 4 of 4 rows present in the deliverable (takeover row,
  import row with consent flag and optional name, one example line each, no
  new normative rule), verified by reading the edited file and by the two
  green gates above.

## Open questions — spec sentences found missing

Q1. No sentence spells the explicit takeover flag. The full extent of the
spec text is: "Takeover of a specific unmanaged file outside onboarding
requires the explicit takeover flag and performs the same notice and backup;
without the flag, section 8.3 applies and the operation fails rather than
overwrite." (§9.5) Published `--takeover`: intended spelling?

Q2. No sentence states the standalone takeover's operand/scope grammar.
"Takeover of a specific unmanaged file" (§9.5) suggests a file, while the
inverse operation states a scope: "`env unmanage [--restore-backups]
[--env <env-id>] [--target <target-id>]` takes every in-place surface set of
the named scope (default: every scope)" (§9.2). Published scope flags mirror
unmanage; no default scope is stated in the row. Path operand or scope flags,
and which default?

Q3. No sentence spells the per-operation consent flag. The full extent is:
"it proceeds only under an explicit per-operation consent flag, which
re-reports the loss list as warnings under the same diagnostic." (§9.6)
Published `--allow-lossy`: intended spelling?

Q4. No sentence names the import subcommand or its name-supply flag. The
full extent is: "`name` `imported` unless the operator supplies a name under
the core §2 grammar" (§9.6). Published `curator profile import [--as
<name>]`: intended command and carrier flag?

Q5. No sentence states whether the explicit takeover is a standalone command
or a flag on the existing mutating operations. The trigger list names "an
explicit takeover" alongside "`profile install`, `profile use`,
`profile sync`, `profile update`, `env resolve --repair`" (§9.5) without
attaching it to a command. Standalone `env takeover`, or a `--takeover`
modifier on those commands?
