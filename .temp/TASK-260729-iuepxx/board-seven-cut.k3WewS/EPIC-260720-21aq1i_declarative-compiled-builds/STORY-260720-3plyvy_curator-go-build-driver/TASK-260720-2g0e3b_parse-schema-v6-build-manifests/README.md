# Parse and validate schema v6 build manifests

## Description
Extend Curator skill manifest parsing with schema version 6, top-level build_roots, and the single closed build command shape type=build, driver=go-v1, source_dir. Perform every compiler-free semantic check at parse time and retain byte-for-byte behavior for schemas 1 through 5 and the legacy runtime fallback.

## Scope
Own internal/skillspec/types.go, internal/skillspec/parse.go, their unit tests, and schema-resolution conformance consumers. Validate both agent-skill.json and the legacy csk-skill.json alias. Build roots must be real link-free non-dot directories, unique and pairwise disjoint, must not overlap runtime_roots, and every root must be used. Each source_dir must be a real directory below exactly one root, with that root containing the nearest go.mod. Do not implement closure activation, Go process execution, cache storage, or install lifecycle here. Base the work on Curator origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 and the landed rc.4 protocol artifacts.

## Acceptance Criteria
SupportedSchemaVersions accepts 6 while all schema 1 through 5 parser tests remain unchanged; canonical and legacy v6 manifests with mixed script, system, and build commands parse to explicit BuildRoots plus Driver and SourceDir fields; schema 5 build commands, legacy runtime build forms, unknown drivers, mixed command shapes, args, env, flags, output, toolchain, hooks, missing or unused roots, dot roots, overlapping roots, runtime overlap, escaped or linked paths, missing root go.mod, and intervening nested modules fail with stable field-scoped diagnostics; no Go executable is invoked by parsing; authoritative v6 schema cases consumed through CURATOR_CONFORMANCE_ROOT pass.
