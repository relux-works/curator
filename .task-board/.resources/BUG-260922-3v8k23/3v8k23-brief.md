# BUG-260922-3v8k23 — ubuntu/macOS internal/install timeout budget (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. Priority: this flake now costs a full republish cycle on most curator Stories (latest: gate run
35994653101, Race(ubuntu) `FAIL internal/install 1800.100s`; earlier 35713206841 Test(ubuntu)). Implement the fix in the description:
choose between a larger budget class for ubuntu/macOS lanes (same class as Windows, e.g. 60m) and/or reducing internal/install wall time;
state the measured baseline and headroom. Keep per-package timeout semantics (a synthetic hanging package must still fail within budget —
negative row). Pin every lane's timeout expression in `.github/ci/gate-selftest.sh` (all lanes, including the Windows candidate-suite line
that is currently unpinned — M4 survivor in the Windows budget memory). AC 4 is superseded by the CHANGELOG POLICY: do not edit
CHANGELOG.md; put the entry text in results under "## CHANGELOG entry (for release prep)". Evidence for AC 1 = the handoff gate run's
per-lane internal/install elapsed/budget table (≤50 %); if one gate run is all you can show, say so. Bounded local runs of gate-selftest
and ledger scripts. Attach results, check DoD, `task-board handoff BUG-260922-3v8k23 --role developer`. A `run_wrote_outside_worktree …
policy warn` block is a warning.
