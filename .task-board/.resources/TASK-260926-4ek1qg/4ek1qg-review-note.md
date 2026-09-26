# Review note — TASK-260926-4ek1qg + exact head of curator-spec PR #88 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

1. Review the 4ek1qg CR against `implq-brief.md`: curator pin 0a628621 with accurate comment; sparse checkout (`/*`, `!/.task-board/`,
   non-cone) keeps everything the Go suite needs (sources, declared submodule) — check the suite reads nothing under .task-board; the
   old → new Go declaration table is honest (TestARefusalPrecedesEveryWorkerSurface → TestPreflightRefusalCases proves the same claim —
   read both; the added ProductionConsumersCoverAllCases row is real); CHANGELOG entry fine.
2. Exact head of PR #88 = b68a937: b202b5d (accepted erratum dcc7f01 carried onto 5746367; non-CHANGELOG paths patch-id identical,
   CHANGELOG union, stray root results file dropped) + b68a937 (the 4ek1qg patch; per-file patch-id identical to your CR). Confirm both.
3. `gh pr checks 88 --repo relux-works/curator-spec` on head b68a937: every required check green, INCLUDING Implementations
   (windows-latest) — the Windows sparse checkout was unverified locally. Wait (bounded, ≤40 min) for completion and report.
accept_cr on 4ek1qg only if 1–3 hold; else changes requested with file:line / failing check. No LOGBOOK.md.
