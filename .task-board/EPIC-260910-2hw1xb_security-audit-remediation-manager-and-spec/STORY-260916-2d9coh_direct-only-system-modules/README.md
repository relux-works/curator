# STORY-260916-2d9coh: direct-only-system-modules

## Description
Finding E2 (High): class: system modules (which replace or append the tool system prompt; pi SYSTEM.md replaces it wholesale) are admitted from any package anywhere in the closure, selected by a range; the only control is the always-warn finding context-system-module-present. Weights order chapters, they do not gate admission. Composed with E1 a transitive dependency update rewrites the operator system prompt behind a warning.

Implementation verification (TASK-260916-dv7xv5 rev2, curator main 80483355, launcher main b34e1e27, static): confirmed as specified. SystemPrompt (contextmaterialize.go:248-265) appends the system modules of every EmittedOrder member with no admission check, called from managed.go:1785; contextaudit.go:22/:93 classify context-system-module-present and Report.Blocking (:109-115) blocks on Findings only. The admission rule slots in between EmittedOrder and Applicable.

## Scope
curator-spec environments §2/§12/§13; curator contextresolve/contextaudit

## Acceptance Criteria
Machine policy transitive_system_modules = drop | error, default drop: a class: system module carried by a package the root does not name directly is skipped with a warning naming the package and module while root modules still materialize, so an install or update never breaks; error makes it the resolution error context_system_module_transitive; naming the package directly in the root or a per-package waiver in machine configuration admits it; the fragment works.relux.curator.system-modules flag keeps carrying presence for ax resume refusal; conformance vectors cover direct, transitive-dropped, transitive-error and waived cases
