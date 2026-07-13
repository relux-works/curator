# Changelog

All notable implementation changes are recorded here.

## Unreleased

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
