# Review note — TASK-261001-2yvag1 rev7 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev7 (base e87d488b, 28 paths, gate green, no overlap with trunk advances since) against your own rev6 verdict (`TASK-261001-2yvag1_review-verdict-rev6.md`) and `2yvag1-rework-r6.md`.
- **R1 (P1).** Rerun your symlinked config-parent probe: an outside parent holding a valid-looking auth link, under bare Resolve and `--repair`. Both must refuse non-zero, emit no fragment, and leave outside and native bytes unchanged. Confirm the new production-entry rows exist, and kill the boundary-drop mutant yourself.
- **R2 (P2).** Rerun your missing-`permissions` emission mutant: it must now FAIL. Confirm permanent schema assertions run on the actual Resolve/CLI output, and that reader-only rows are classified as such.
- **R3 (P2).** `internal/testcli/*` must be byte-identical to base.
- **Regression.** Everything you ruled OK in rev6 must still hold (spot-check 16/16, the status semantics, the seam guard).

accept_cr, or changes requested with file:line. Never spell any employer name.
