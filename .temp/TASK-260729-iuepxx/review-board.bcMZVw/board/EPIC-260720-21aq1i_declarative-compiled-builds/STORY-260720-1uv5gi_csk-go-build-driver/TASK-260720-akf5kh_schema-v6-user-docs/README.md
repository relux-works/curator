# Document schema v6 compiled commands

## Description
Update English source documentation and maintained Russian mirrors so authors and operators can use csk schema 6 without inferring build, cache, security, status, repair, or activation behavior from code.

## Scope
Own README.md and README.ru.md, ARCHITECTURE.md and ARCHITECTURE.ru.md, SECURITY.md and SECURITY.ru.md, docs/skill-authoring.md, CHANGELOG.md, and directly related link or documentation tests. Include one complete mixed schema 6 example, fixed go-v1 prerequisites and limits, csk-specific manager-home build-cache layout, logical portability boundary, build-root context exclusion, lifecycle and dry-run ordering, status and reinstall repair, GC, and Unix and Windows command resolution. Maintain the README tool and validation command section. Do not claim releases, tags, pins, signatures, reviews, or interoperability evidence that do not yet exist.

## Acceptance Criteria
Documentation explains Go 1.23-plus accepted-family and vendor-only requirements, native target, private telemetry off, cgo, PGO, workspace, generator, external-link and network prohibitions, no arbitrary args or hooks, and why other drivers fail closed. It distinguishes installed content hashes from build-source identity and receipt consistency from protected-state provenance. It documents project, global, and hybrid transaction rollback, compiler-free dry-run outcomes, status --json and --check, reinstall repair, locked GC, and direct project/global shim resolution on Unix and Windows. English and maintained Russian files agree, local links resolve, examples parse, documentation tests pass, and strict mypy remains green.
