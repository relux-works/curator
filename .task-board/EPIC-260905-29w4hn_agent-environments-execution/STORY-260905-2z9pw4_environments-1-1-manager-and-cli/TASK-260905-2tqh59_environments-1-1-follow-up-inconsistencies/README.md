# TASK-260905-2tqh59: environments-1-1-follow-up-inconsistencies

## Description
Follow-ups recorded by the batch-3 producer and reviewer (TASK-260905-369vye): (1) Decision 0012 impact row for manager §12.4 should read bytes-change (isolation knob, liveness row, seeds) — record as a 0012 erratum line; (2) environments §12.1 secret_material_waivers.pin spelling (bare lowercase hex 40/64 as the schema accepts) — state it in §12.1; (3) system-config-v2 with the environments lockable keys is its own batch — schema, cases, manager §1 text; (4) environments §7.7 lacks a row for environment_form_unavailable (only in §5.7) — add or cross-reference; (5) fold the cycle-4 nits N-a/N-b of environments 1.1 (§8.1 linked bullet parenthetical, §10.1 hash path for a copied surface); (6) reviewer F1: extend the tools/validate.py §12.1 cross-check to enum value sets (a widened precedence.winner enum must fail); (7) reviewer F2: negative schema cases for overlay.range, overlay.tag, and empty overlay.source grammars in manager_config.go; (8) reviewer F3: cite environments §9.2/§10 for the curator run --repair rule in manager §12.5 and cli/curator.md. One producer/reviewer cycle after the schemas batch lands. (9) batch-2 reviewer minors F1-F4 on the schemas/vectors (TASK-260905-1xkxe4 review-findings-schemas-1): lock member required_by self-reference not rejected; marker copies must be a subset of paths, per-surface form, unknown surface keys; fragment path segments with '..' not rejected; the mcp-pi-none marker case needs an env_names note.

## Scope
(define task scope)

## Acceptance Criteria
Each item applied as a recorded change in its owning document with a review ACCEPT; landed by fast-forward.
