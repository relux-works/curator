# TASK-260906-2t2t6w: launcher-spec-0-2-2-minors

## Description
Four non-blocking minors from the 0.2.1 review (TASK-260905-2czqqy_review-findings-launcher-0.2.1.md): (1) §4.1/§6 — --repair widened resolve failure modes (environment_marker_invalid, environment_surface_unmanaged_conflict, environment_backup_exists, environment_seed_unreadable) that all collapse into resolve_invocation_failed whose gloss says curator was not startable; add a pass-through clause so Curator diagnostic code and message print verbatim, as ax_handoff_failed does; (2) §4.6 three simultaneous pre-launch checks with one required diagnostic line and no stated order; (3) and (4) as recorded in the findings. Fold into SPEC 0.2.2-draft with the next touch.

## Scope
(define task scope)

## Acceptance Criteria
SPEC 0.2.2-draft applies all four minors; make check green; independent review ACCEPT; landed by fast-forward.
