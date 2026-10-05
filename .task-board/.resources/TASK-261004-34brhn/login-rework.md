# THE ONLY CURRENT INSTRUCTION — TASK-261004-34brhn rework R1 (researcher)
Read `TASK-261004-34brhn_review-verdict.md`. The design passed. The ONLY blocking issue is R1: reproducibility.
In `.research/261004_CIP-0003-claude-managed-home-credential-modes_evidence.md`, add the exact, sanitized script bodies for every claimed check:
- the Python assertion processes behind each "verified" or "measured" claim, including fixture construction and the configuration snapshots;
- the inline artifact checker for document structure checks.
For each script, give the pinned inputs (curator commit, spec tag) and the expected counts. Rerun each script once and record the real exit code and date. Keep the historical results labelled as historical. Change no product code, never edit LOGBOOK.md, and print no secrets or personal paths.
Then update the results resource, run `task-board handoff TASK-261004-34brhn --role researcher`, and END YOUR TURN.
