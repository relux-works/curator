# Acceptance record and operational notes

TASK-261016-w0yx2w — retained-handle snapshot destination publication, CR revision 1.

The first accept_cr attempt refused with worktree_protected_authority_indeterminate: main changed between advertisement and exact-ref fetch. No local-ref fallback was used. A fresh retry succeeded and returned cr_state=accepted, status=integrating, evidence=TASK-261016-w0yx2w_review-verdict-rev1.md. The accepted revision remains not landed.

The successful mutation also reported run_wrote_outside_worktree under policy=warn, including unattributed control-root changes and foreign worktrees. This is an unresolved attribution warning, not a silent clean-boundary claim. This reviewer issued no direct control-root edits: scratch writes were inside the run worktree and board writes used task-board mutations/resource commands. No foreign changes were reset, removed, adopted or cleared. The board accepted the revision with that warning; operator inspection of its persisted boundary report remains available.

The verdict's origin/main equality describes the initial pre-review fetch. Main advanced during the full hosted run; the acceptance transaction used fresh authority. Manual PR landing must perform its own exact-head freshness and combined-tree requirements.
