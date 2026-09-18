# Changelog

All notable implementation changes are recorded here.

## Unreleased

### Added

- E4: umbrella provider lookup from trust roots (warning release,
  revision A). Unknown subcommands still resolve `curator-<name>` on the
  ambient `PATH`, but the manager now computes the trust verdict against
  the trust roots — the manager install directory, then the new
  machine-configuration `provider_directories` list (environments
  §11/§12.1, lockable under §12.2) — and warns
  `subcommand_provider_outside_trust_roots` with the resolved path, the
  roots consulted, and the migration hint when the selection lies
  outside them. Providers inside manager-published or managed
  directories stay refused (`subcommand_provider_untrusted`), an
  unreadable trust root fails (`subcommand_provider_root_unreadable`,
  never absence), and `env status` reports the resolved provider path
  with its trust verdict per discovered provider — always `curator-run`
  and `curator-session` — with refused, missing, and unreadable rows
  non-current for `--check`. Revision B (trust roots only, `PATH`
  never selects) follows in a later release; this story blocks proposal
  0016 / `path_prepend`.
- Scoped HTTPS credentials for external build repositories. A `build_https`
  configuration section maps a source scope to a token source — the operator's
  own Git credential for the host, a manager-namespaced keyring entry, or a
  named environment variable — and the `curator config build-https` command
  (`add`, `login`, `list`, `remove`) manages those selections. Selection is
  resolved per repository by canonical source identity before the first fetch
  and is never selectable by a package; a private HTTPS fetch is answered by a
  manager-owned, host-pinned askpass broker, and an uncovered repository stays
  anonymous (`Spec core §12.2`).
- Operator documentation for scoped HTTPS build-repository token sources,
  credential resolution, and the `curator config build-https` command.
  `CURATOR_BUILD_HTTPS_TOKEN` without `CURATOR_BUILD_HTTPS_HOST` is
  identity-unbound and may be offered to every HTTPS build-repository
  host reached by the run; bind it to one host or use a `build_https` scope
  (`Spec core §12.2`).
- Schema-8 first-party module roots for the `go-v1` driver: a build root may
  replace declared module directories elsewhere in the snapshot, and those
  directories join the directive, cgo, and assembly scan surface
  (Protocol Core §4.2.3).
- S6: shell-hook trust gate — warning release (`A-warning`). The POSIX and
  PowerShell hooks now verify a project `.agents/env.sh` / `.agents/env.ps1`
  against the manager-home approval state before sourcing it, and warn once
  per shell session with `shell_hook_env_unapproved` (no record) or
  `shell_hook_env_changed` (bytes differ; re-approval required) naming the
  absolute path and `curator hook approve <path>`. Files the manager itself
  writes are recorded as `approved_by: manager` at write time and source
  silently. This release only warns: unapproved and changed files are still
  sourced. Migration: run `curator hook approve <path>` for each project env
  file the warning names (a project-local approval record is ignored — only
  the manager-home record counts). The operator approval surface is
  `curator hook approve <path>` (records the current bytes as
  `approved_by: operator`, fails without recording when the file is absent
  or unreadable), `curator hook approvals` (read-only listing), and
  `curator hook revoke <path>` (removes the record; revoking a missing
  record leaves state unchanged and exits 0). `curator status` and
  `curator env status` report the shell-hook trust posture per known env
  file (`approved`, `shell_hook_env_unapproved`, `shell_hook_env_changed`);
  a recorded file whose bytes are missing or unreadable keeps its row
  with an explicit qualifier and is non-current under `--check`, as a
  changed file is, while an unapproved file whose bytes read stays a
  warning row. `status --json` gains `shell_hook_trust` (plus
  `shell_hook_trust_warnings` when approval-state warnings exist, so an
  unreadable approval state surfaces in JSON and fails `--check`). The
  enforcing revision (`B-enforcing`,
  refuse without sourcing) follows in a later release
  (Manager profile §8).
- E2: direct-only `class: system` modules. Only the system modules of direct
  packages — the root, the active overlays, and the packages their
  `requires.contexts` name — plus packages admitted by a
  `system_module_waivers` entry reach the system-prompt output and the
  launch fragment; transitive system modules are refused. The machine knob
  `transitive_system_modules` selects `drop` (default, non-breaking: the
  module is skipped at materialization with the
  `context_system_module_dropped` warning naming package and module) or
  `error` (opt-in strictness: resolution fails with
  `context_system_module_transitive` and the lock is left unchanged). The
  knob locks to `error` only; waivers are not lockable. `curator env
  status` reports the effective policy value with every dropped module by
  package and path (Protocol environments §3, §5.5, §12).
- S4 warning release (`s4-warn`, audit finding S4): `profile install`,
  `profile update`, and `env status` now surface every resolved MCP
  declaration package — package, version, transport, stdio command, args,
  and requested `env_names` — as one `mcp-declaration` row printed after
  the audit gate passes and before the lock is published; an empty MCP
  package allowlist warns `mcp_package_allowlist_empty`; and a resolution
  that passes an operator variable outside the configured
  `passable_env_names` list — or any passed variable when the knob is
  absent — warns `mcp_env_passthrough_unlisted` naming the variables and
  the knob with the migration hint ("list the named variables to keep
  passing them after the flip"). `env status` postures the active S4
  profile with the effective `passable_env_names`. An explicit
  `passable_env_names: null` stays unbounded with no warning. The
  enforcing profile (`s4-enforce`: absent knob is empty, unlisted names
  dropped with `mcp_env_passthrough_dropped`) is implemented behind the
  same option and follows in a later release; this release keeps the
  pre-S4 unbounded behaviour and only warns
  (Spec environments §2.2, §2.3, §10.3, §12).
- Conformance pin → v1.0.0-rc.12 (`dced9b8`): the hosted gate now runs
  the manager-config-v2, environments (with the E2 system-module
  admission cases), umbrella-provider-resolution, and
  environments-env-passthrough vectors at the rc.12 root, and the four
  `root-content` skip paths the E2/E4/S4 candidates carried for those
  families are removed — an absent family now fails instead of
  skipping.

### Changed

- A `go-v1` build root whose `vendor/modules.txt` carries a directory
  replacement the command does not declare is now refused with
  `build_module_root_directive_undeclared`. §4.2.3 requires a command with an
  absent or empty `modules` list to have an *empty* effective replace set, and
  `go mod vendor` materialises an annotation for an **unused** `replace`
  directive exactly as it does for a used one. A schema-6 or schema-7 skill
  that carried an unused directory `replace` therefore built before and now
  fails; declare the directory under `modules`, or drop the directive.

### Fixed

- E4: the user-bin shim directory counts as manager-published — and
  refuses providers under revision A — only once the manager has
  actually published shims there (the ownership ledger exists) or the
  operator declared it via `CURATOR_GLOBAL_USER_BIN` (now honored by
  the lookup; previously ignored). A merely selected PATH entry holds
  no manager-written content, so providers there warn
  `subcommand_provider_outside_trust_roots` and stay current instead
  of refusing; on Windows, where temp trees sit below the user
  profile, the selector otherwise claimed ordinary PATH directories.
  Trust-root and refused-directory comparisons now also match by
  filesystem identity, so an 8.3 short spelling and a case variant
  name the same directory (environments §11).
- Git snapshots are extracted from the object database (`git ls-tree -r -z`
  plus `git cat-file --batch`) instead of `git archive`, so every regular file
  carries exactly its committed blob bytes. `git archive` applied
  `core.autocrlf`, `text`/`eol`, and `export-subst` to its output, which made
  snapshot content hashes depend on the acquiring machine's git configuration
  and the repository's `.gitattributes` (Protocol environments §1.2, core §6.2,
  §6.5). The skills snapshot cache and closure scratch snapshots both use the
  new path; symlinks, gitlinks, path escapes, `.git` components,
  platform-path collisions, and oversize blobs are refused, and `100755`
  keeps its executable bit. Every refusal that needs no blob bytes is decided
  from the `ls-tree -l` listing before `cat-file` starts, so a refused
  extraction writes nothing and never leaves `cat-file` blocked on its
  output pipe; a failure while streaming terminates and drains the child
  before waiting on it and removes what the call wrote. Closure scratch
  snapshots are extracted into a sibling staging directory and renamed into
  place only on success.

- Status no longer reports a successfully installed schema-8 skill as
  `needs-install`. Every reader that decides whether a recorded compiled
  command is knowable now bands on the whole build-bearing marker schema set
  (2, 3 and 4) instead of only the schema the release writes, so a schema-7 or
  schema-8 installation is reported `current` on the marker it actually wrote.
  The old remedy was self-contradictory as well: it told an operator holding a
  marker v4 to reinstall so the manager would record marker schema 2, a schema
  it would never write for that band.
- Garbage collection no longer drops the live build references of a marker v4.
  A schema-8 installation's recorded cache keys went unmarked, so a
  maintenance pass could delete protected cache entries the installation was
  still running from.
- A marker document at a readable schema that is nonetheless invalid is now
  reported as an invalid document rather than as one from a newer manager.
  Schemas 3 and 4 are read by this release, so `upgrade the manager` was never
  the remedy for them.

## 0.12.5 - 2026-07-14

### Added

- Shared conformance coverage for canonical, legacy, dual-file, conflict,
  invalid-manifest, and runtime-fallback resolution.

### Changed

- `agent-skill.json` is now the implementation-neutral canonical skill
  manifest filename; `csk-skill.json` remains a protocol 1.x read alias.
- Diagnostics and authoring guidance now point new packages to the canonical
  filename.

### Fixed

- Dual manifests are validated independently and accepted only when their JSON
  values are equal; mismatches fail closed with
  `conflicting_skill_manifests`.
- An invalid modern manifest no longer falls through to another filename or
  `agents/runtime.json`.

## 0.12.4 - 2026-07-13

### Added

- Idempotent `bootstrap --if-missing` for repository-managed onboarding.
- Self-contained POSIX and Windows command launchers that carry skill and
  declared system dependency paths without shell profile setup.

### Changed

- `upgrade` fetches only the selected project or global dependency closure and
  deduplicates repositories shared by multi-project operations.
- Install and upgrade dry runs use temporary resolution state and no longer
  mutate persistent source checkouts, caches, audit or registry state, runtime
  state, configuration, or installation artifacts.
