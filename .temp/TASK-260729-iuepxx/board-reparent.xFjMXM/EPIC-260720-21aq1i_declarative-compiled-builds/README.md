# Declarative compiled skill builds

## Description
Define and implement a first-class build phase for compiled skill commands without reintroducing arbitrary install hooks. The protocol change is authored in curator-spec first, then implemented in the Go Curator manager, then independently in the Python csk manager. The initial executable driver targets Go; the design must include an evidence-backed language/toolchain applicability matrix and an extension path for future drivers.

## Scope
In scope: threat model and language assessment; a new versioned skill-manifest schema; normative manager lifecycle, caching, environment isolation, and artifact rules; conformance vectors; Go reference implementation; Python implementation; authoring documentation and end-to-end fixtures. Out of scope: generic shell hooks, package-provided argv, automatic installation of system compilers, remote build services, and immediate support for build systems that execute package code.

## Acceptance Criteria
A reviewed protocol schema and decision record define compile-only build drivers and a normative Go driver; both Curator and csk independently parse, validate, build, cache, install, and launch a fixture command; unsupported or unsafe declarations fail closed; dry-run executes no build; existing schemas 1-5 remain compatible; relevant test suites and cross-implementation conformance pass; language candidates are classified with security rationale.
