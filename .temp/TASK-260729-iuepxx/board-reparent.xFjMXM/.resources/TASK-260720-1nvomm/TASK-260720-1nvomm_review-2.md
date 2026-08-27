# TASK-260720-1nvomm final review

## Verdict

Accepted. Route to done.

## Evidence

- Repository HEAD and origin/main remain at required base 57c1f56846d221ecc55786bd3c2467ec32f11730.
- Accepted contract SHA-256 independently verified as 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681.
- Product changes are limited to SECURITY.md, protocol/core.md, and decisions/0004-compile-only-build-drivers.md. Task-local .temp material is not part of the product diff.
- Core normatively defines schema 6, strict build_roots and go-v1 semantics, static exclusion for real builds, cache hits, and dry-runs, manager-derived artifacts, build-source and toolchain identity, logical cache and receipt semantics, marker v2, and schema 1 through 5 compatibility.
- SECURITY.md explicitly closes package shell, argv, environment, output selection, hook, plugin, generator, unsafe build-system, external-link/libgcc, native-input, and built-output execution surfaces, and records compiler-input and protected-cache trust boundaries.
- Decision 0004 records the closed Go-only v1 decision, rejected alternatives, non-normative physical layout, compatibility, lifecycle isolation, and future-driver review rule.
- Prior review gaps are corrected: bootstrap probes use the required manager-owned empty CWD and clean private environment; trusted go-list results require Standard == true and Goroot == true; schema 1 through 5 dependency narrowing remains script-only.

## Validation

- python3 tools/validate.py: passed; validated 30 schemas and 93 vector files.
- make validate: passed; validator, 8 Python tests, and go test ./tools/... all green.
- git diff --check: passed with no output.
