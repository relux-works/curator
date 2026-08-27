# Specify protocol v6 core and security contract

## Description
Update the normative protocol model from the accepted TASK-260720-poa3ze contract. Define manifest schema 6 as a declarative compiled-artifact extension with build_roots and a closed go-v1 command. Build sources must be statically excluded from agent context and runtime copying on real installs, cache hits, and dry-runs. Record the no-hooks threat boundary, build-source identity, logical cache and receipt semantics, marker v2 compatibility, and the Go-only v1 decision. Preserve every schema 1 through 5 behavior.

## Scope
Work only in /Users/iv/Developer/ReluxWorks/curator-spec from origin/main commit 57c1f56846d221ecc55786bd3c2467ec32f11730. Own protocol/core.md, SECURITY.md, and new decisions/0004-compile-only-build-drivers.md. Use TASK-260720-poa3ze_compile-only-build-drivers.md at accepted SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681 as the contract. Do not edit schemas, vectors, generator, manager profile, CLI guide, or release metadata in this task.

## Acceptance Criteria
protocol/core.md normatively defines schema 6, build_roots, the strict build command, build-source exclusion, manager-derived artifact paths, curator-build-source-v1, logical cache and receipt identity, marker v2 implications, and schema 1 through 5 compatibility; SECURITY.md explicitly forbids package shell, argv, environment, output selection, hooks, plugins, generators, unsafe build systems, external-link fallback, and executing built output while documenting compiler-input and protected-cache trust boundaries; decision 0004 records the closed Go-only v1 choice, rejected alternatives, non-normative physical cache layout, and future-driver review rule; terminology matches the accepted contract without weakening any MUST or MUST NOT; python3 tools/validate.py passes.
