# Changelog

All notable implementation changes are recorded here.

## Unreleased

### Added

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
