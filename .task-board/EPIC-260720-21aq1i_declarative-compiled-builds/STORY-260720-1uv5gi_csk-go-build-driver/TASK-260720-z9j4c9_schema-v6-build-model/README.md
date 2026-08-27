# Implement schema v6 build manifest model

## Description
Extend the Python manifest domain and skill validation with the accepted schema 6 surface: top-level build_roots and a closed build command containing only type build, driver go-v1, and source_dir. Preserve schema versions 1 through 5, canonical and legacy manifest parity, dependency closure activation, and the no-build legacy fallback.

## Scope
Target the cocoaskills repository from a clean, fast-forwarded local main and a task-scoped worktree based on all landed dependencies. Own src/csk/skillspec.py, the build-domain package initializer, validation-only changes in src/csk/skillcheck.py, and focused parser and skill-check tests. Validate static paths, root use and overlap, and module-root relationships without invoking Go. Do not implement hashing, toolchain probes, compiler execution, cache storage, or install mutation.

## Acceptance Criteria
Schema 6 accepts strict build commands alongside existing script and system commands and retains schema 5 capabilities and dependencies. build_roots are unique, pairwise disjoint, non-dot, real link-free directories that do not overlap runtime_roots; every root is used; every source_dir is below exactly one root whose direct go.mod is its nearest module file. Unknown drivers or fields, mixed shapes, schema 1 through 5 build surfaces, legacy runtime builds, root or escaped sources, missing or non-directory paths, unused or overlapping roots, runtime overlap, and nested modules fail before activation. Existing schema 1 through 5 behavior remains byte-semantically compatible. Focused pytest and strict mypy pass.
