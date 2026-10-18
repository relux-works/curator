# BUG-261018-uoisfe: review B, native skills surface for claude_code, codex_cli and pi (deciding reviewer)

Review the published Change Request revision against `gw-NATSKILL-r2-brief.md`. Use `gw-NATSKILL-brief.md` only where round 2 does not override it. Muse, opencode and managed homes are out of scope by the orchestrator's decision. Their unchanged behaviour is itself a check.

You are the deciding reviewer: accept only when every row holds on the exact revision.

Sweep exactly the twelve attack-surface rows of the round 2 brief, in order, and record each in the verdict table with evidence: file and line, test name, command and exit code. Then check these points:
- every gate row (3, 4, 5, 6, 7, 11) has a narrowing mutant that fails for the stated reason;
- the coverage ratio the producer reports matches what you drove;
- no test or code path touches the real HOME;
- the root `README.md` is unchanged;
- the full local suite and the hosted gate pass on the exact candidate. An infrastructure failure is reported as a blocker, never a pass.

Before you start, compare the CR base with fresh protected trunk. A path overlap means the revision is stale: say so, and do not accept a stale revision.

Record the verdict with `accept_cr` or `request_changes` in the board's review flow, and attach `BUG-261018-uoisfe_review-verdict-rev<N>.md`. Then END YOUR TURN.
