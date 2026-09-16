# TASK-260915-4f44cs: admit-worker-policy-and-rose-air-ci-lane

## Description
Bootstrap for the unified delivery goal: rewrite the tracked task-board.config.json ceilings so the board admits exactly gpt-6-astra:low (codex producers) and claude-fable-5-1:low (claude reviewers) with role defaults seeded from them, and add a Test (rose-air) CI lane on the self-hosted macOS ARM64 runner. Delivered by the orchestrator inline as curator PR 70 (head 4d240bac6a30aea2585f566068b55a2b325cbef7), independently reviewed by a Claude Fable reviewer outside task-board because the pre-change ceilings could not admit the policy reviewer; the verdict is attached.

## Scope
(define task scope)

## Acceptance Criteria
PR 70 reviewed at its exact head and accepted; hosted checks and the rose-air lane green; exact reviewed head fast-forwarded to main; preflight on main resolves the two pairs.
