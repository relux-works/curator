# TASK-261004-2asduq: launch-command-environment-fragment-design

## Description
DESIGN PENDING — NOT ACCEPTED FOR EXECUTION. Operator decision 2026-10-04: decide the design first, then prioritise. Do not spawn producers and do not schedule until the operator accepts a design. Launcher project-roots idea. The current launch-env fragment cannot carry an external project root or an append list (path_prepend is one environments-root path). Sketch: a new fragment revision with a typed command_environment object {version, project_root, path:{operation:append, roots:[{scope:project|global, path}]}}; closed enums, at most 2 absolute canonical roots; the manager computes roots from the registered project + launch CWD + its own home, never from profile bytes or repository content; curator-run appends project then global to the admitted child PATH, deduplicated, set in the child env and the tracked launch plan and hashed into plan identity; never source repository files. Open security decision: append guarantees precedence, not integrity — consider a manager-owned per-project shim dir generated from protected activation records. Native note: Claude settings.json env and Codex shell_environment_policy.set REPLACE PATH (no append operator).

## Scope
(define task scope)

## Acceptance Criteria
Research document (CIP draft per cip-template.md where the brief says so) under .research/ with options, tradeoffs, recommendation, evidence (file:line or measured probes) and decision-ready open questions; no product code changes; no secrets read or printed; LOGBOOK.md untouched.
