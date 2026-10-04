# THE ONLY CURRENT INSTRUCTION — TASK-261004-31fcvu rework after review rev1 (developer, docs)

Read `TASK-261004-31fcvu_review-verdict-rev1.md`. Fix both findings in the same workspace; only CHANGELOG.md changes; never edit LOGBOOK.md.

1. **P1 — the 27 dropped entries.** They are NOT duplicates. They sat under `## Unreleased` in the v0.15.0-rc.2 tag: that code shipped in rc.2, but rc.2's notes never moved them into its section. Restore ALL 27, byte-for-byte as they read at the base, in a clearly labelled subsection at the END of the `## 0.15.0-rc.3 - 2026-10-04` section:
   `### Shipped in 0.15.0-rc.2 but not recorded in its notes`
   followed by one sentence: "These entries were listed under Unreleased when 0.15.0-rc.2 was tagged; the changes are part of 0.15.0-rc.2." Then the 27 entries in their original order. Do NOT edit the rc.2 section or anything below it (byte-identical to the base). If one of the 27 is also genuinely represented by an rc.3 line (post-rc.2 work on the same feature), keep both; do not merge them.
2. **P2 — internal runner name** (candidate lines ~84–86): replace with neutral wording, e.g. "select the explicit self-hosted runner label". Then scan the whole rc.3 section for any other hostname, user path or internal machine name.
3. Fix the validator so it would have failed rev1: every previous Unreleased entry must appear either in a released section (rc.2 or older, located by section, not anywhere in the file) or verbatim in the new rc.3 section. Run it; exit 0. Update the results/disposition ledger.

Then `task-board handoff TASK-261004-31fcvu --role developer` and END YOUR TURN.
