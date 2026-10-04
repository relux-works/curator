# Review note — TASK-261004-31fcvu rc.3 release notes (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Release-critical docs: these lines become the public v0.15.0-rc.3 notes. Brief: rc3-notes-brief.md + rc3-notes-apply.md.

Verify, with real exit codes:
1. Only CHANGELOG.md changes. Sections from `## 0.15.0-rc.2 - 2026-09-26` down are byte-identical to the base. A fresh empty `## Unreleased` precedes `## 0.15.0-rc.3 - 2026-10-04`.
2. The producer claims 27 of 32 previous Unreleased entries are exact duplicates of already-released (rc.2 or older) lines and were dropped. Prove this independently for EVERY dropped line: find its identical text in a released section. A dropped line with no released twin and no rc.3 representation is a defect.
3. Completeness: for `git log v0.15.0-rc.2..origin/main`, every substantive (non-board, non-research) commit is represented by an rc.3 line. Spot-check at least 10 commits across the range, including: the muse environment (e4a6a8d5), the global-upgrade sweep fix, the rc.14 pin (6bb3c626), the Windows hard-link origin, marker cross-field validation, codex seed A / security posture A.
4. Accuracy: no claim beyond history. In particular: v2 hash writing is NOT enabled; rc.14 is pinned at 43bf0a25; nothing says N1–N4 or fp8vx7 are fixed.
5. Known issues state fp8vx7, the v2 writer deferral to rc.4, audit N1–N4 (#106) unfixed, and B3 excluded. Wording is public-safe: no internal hostnames, personal paths or employer names.

accept_cr if all hold; otherwise changes requested with the exact lines.
