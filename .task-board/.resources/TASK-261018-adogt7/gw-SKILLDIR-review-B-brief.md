# TASK-261018-adogt7: review B, CIP-0021 implementation (deciding reviewer)

Review the published Change Request revision against `gw-SKILLDIR-brief.md` and the curator-spec text of CIP-0021 (PR #162, decision 0027). Accept only when items 1–9 hold on the exact revision. Check in particular:
- the CIP-0021 vectors are driven through the production entry points;
- manifest schema 1 still refuses a skill `directory`;
- audit and materialization act below the directory;
- escape and symlink are refused with the spec's diagnostics;
- the mutants fail for the stated reasons;
- the real HOME and the root README are untouched.

Before you start, compare the CR base with fresh trunk. A path overlap means the revision is stale: say so, and do not accept it.

Record the verdict with `accept_cr` or `request_changes` and attach `TASK-261018-adogt7_review-verdict-rev<N>.md`. Then END YOUR TURN.
