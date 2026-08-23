# Publish the portable compiled-skill authoring guide

## Description
Publish a practical implementation-neutral guide for authors of schema v6 Go command skills, building on the normative schema and CLI documentation without asking authors to infer portability rules from manager source.

## Scope
Work in curator-spec after TASK-260720-3lo9jc and TASK-260720-2g7avf. Own docs/compiled-skill-authoring.md plus its README and conformance navigation links. Reuse canonical terms and link the exact schema, fixture, build-driver vectors, and conformance guide. Cover both manifest spellings only where compatibility requires it. Do not promise manager support before release, document physical cache paths as portable, add package-selected build options, or duplicate implementation-specific troubleshooting.

## Acceptance Criteria
The guide contains a complete valid schema v6 manifest and matching repository tree for the shared minimal Go fixture. It states trusted Go 1.23-or-newer allowlisted toolchain requirements, native target only, vendor consistency, main-package and build-root rules, disabled cgo, PGO, workspaces, downloads, toolchain switching, external linking, generators, plugins, and output execution during install. It explains prompt-context and runtime exclusion, cache miss, verified hit, corrupt or untrusted rebuild, logical portable key and receipt fields versus implementation-specific storage, compiler-free dry-run, status and repair behavior, command launch and argument or exit propagation, and the protected-state trust boundary. It lists resource and compiler-input denial-of-service limits honestly, links the shared cases, labels unreleased support accurately, passes link and documentation validation, and contains no release claim or pin change.
