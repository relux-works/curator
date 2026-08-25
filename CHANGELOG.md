# Changelog

All notable implementation changes are recorded here.

## Unreleased

### Added

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
