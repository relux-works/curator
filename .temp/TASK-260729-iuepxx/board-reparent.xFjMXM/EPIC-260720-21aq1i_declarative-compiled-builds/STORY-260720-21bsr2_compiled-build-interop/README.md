# Compiled-build interoperability

## Description
Validate the protocol and two implementations as independent consumers rather than implementation copies. Add a minimal Go skill fixture and authoring guidance, and document how future compile-only drivers are standardized.

## Scope
Shared conformance fixture/vectors in curator-spec and implementation-side consumers/tests in Curator and csk. Include a language-driver matrix and explicit non-goals for unsafe build systems. Do not pin unreleased commits or fabricate release evidence.

## Acceptance Criteria
The same fixture and expected outcomes run in both implementations; both reject the same negative cases and launch the same built command behavior; documentation includes a complete schema v6 example, toolchain prerequisites, cache behavior, security limits, and the process for adding future drivers; no release pins are changed before real commits/releases exist.
