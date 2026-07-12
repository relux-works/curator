# EPIC-260712-b62557: install-engine

## Description
Phase 4 of docs/implementation-plan.md. The install pipeline: context whitelist and atomic staging (Spec 4.2), localization (Spec 4.3), markers with tamper detection (Spec 8.5), runtime store and shims (Spec 8.6), managed adapters with unmanaged-conflict refusal and native-discovery agents (Spec 10), gitignore gate, env files, dry run, and the normative phase order (Spec 8.1).

## Scope
See docs/implementation-plan.md, Phase 4. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Full install cycle on a temp project with real git fixtures
- Second install reports up-to-date; local edits are detected via re-hash
- Windows shims and adapter copy fallback covered in CI
