# Reviewer instruction — TASK-261007-2uet3u: rc.4 release notes
Cross-provider review: the producer was muse max; the reviewer is codex gpt-6.1-sol medium (R187).
Verify:
1. Every rc.4 entry maps to a real commit in `git log v0.15.0-rc.3..origin/main`. Nothing landed after rc.3 is missing; check the v2 writer flip, the v1-only NUL gate, N1–N5 and Go 1.26/1.27 in particular.
2. No claims beyond the history.
3. The operator actions for the v2 cutover are correct against the code (re-pin, revocations).
4. Known issues are honest.
5. Only CHANGELOG.md changed; the format matches the earlier sections; there is an empty Unreleased section above.
accept_cr, or request changes with numbered findings. Do not edit files.
