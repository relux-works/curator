# BUG-261018-b2jnyg: review B, floats in a managed .claude.json (deciding reviewer)

Review the published Change Request revision against `gw-FLOATS-brief.md`. You are the deciding reviewer: accept only when every point holds on the exact revision.

Sweep and record in the verdict table with evidence (file and line, test name, command and exit code):
1. spec reading: the producer's conclusion on whether `.claude.json` after first run is tool-owned matches the spec text;
2. regression witness with real-world float keys, exponent and large-integer forms; values preserved as JSON values and number texts;
3. no-op when already satisfied (bytes and mtime);
4. neighbouring variants and unchanged refusals (non-object, symlink, non-regular, unreadable) with no writes;
5. absent-file seed bytes unchanged (golden);
6. verification reader accepts floats;
7. sweep list complete: grep yourself for other CCJ-1 or full re-encodes of tool-owned harness files and compare;
8. plant control fails for the stated reason;
9. no test touches the real HOME; root `README.md` unchanged;
10. full local suite and hosted gate pass on the exact candidate. An infrastructure failure is a blocker, never a pass.

Before you start, compare the CR base with fresh protected trunk. A path overlap means the revision is stale: say so, and do not accept a stale revision.

Record the verdict with `accept_cr` or `request_changes`, attach `BUG-261018-b2jnyg_review-verdict-rev<N>.md`, then END YOUR TURN.
