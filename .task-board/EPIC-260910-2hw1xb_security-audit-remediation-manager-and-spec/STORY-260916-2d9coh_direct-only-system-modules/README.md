# STORY-260916-2d9coh: direct-only-system-modules

## Description
Finding E2 (High): class: system modules (which replace or append the tool system prompt; pi SYSTEM.md replaces it wholesale) are admitted from any package anywhere in the closure, selected by a range; the only control is the always-warn finding context-system-module-present. Weights order chapters, they do not gate admission. Composed with E1 a transitive dependency update rewrites the operator system prompt behind a warning.

## Scope
curator-spec environments §2/§12/§13; curator contextresolve/contextaudit

## Acceptance Criteria
class: system modules are admitted only from packages the root or a machine-config allowlist names directly; a transitive system module is a resolution error context_system_module_transitive with an explicit scoped waiver; the fragment works.relux.curator.system-modules flag keeps carrying presence for ax resume refusal; conformance vectors cover direct, transitive and waived cases
