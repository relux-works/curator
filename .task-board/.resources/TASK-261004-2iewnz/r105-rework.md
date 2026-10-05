# THE ONLY CURRENT INSTRUCTION — TASK-261004-2iewnz rework R1 (researcher)
Read `TASK-261004-2iewnz_review-verdict.md`. The design passed. The ONLY blocking issue is R1: reproducibility.
In `.research/261004_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md`, add the exact, sanitized script bodies for every claimed check:
- the Python assertion processes behind lines ~142 and ~176, including fixture construction and the configuration snapshots;
- the inline artifact checker behind lines ~196–198 (sections, source-link ranges, JSON examples, changed files, publication-pattern checks).
For each script, give the pinned inputs (curator commit, spec tag) and the expected counts. Rerun each script once and record the real exit code and date. Keep the historical results labelled as historical. Change no product code, never edit LOGBOOK.md, and print no secrets or personal paths.
Then update the results resource, run `task-board handoff TASK-261004-2iewnz --role researcher`, and END YOUR TURN.
