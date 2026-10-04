# THE ONLY CURRENT INSTRUCTION — BUG-261004-16a407 (N5): trust pin overrides strict findings (developer, code)

Source: cocoaskills comparison (2026-10-04), confirmed by reading: internal/audit/audit.go decideWithPins returns allow for any pinned content before Decide runs, on both the cached-findings path and the fresh path. Spec manager §7 (rc.14): strict mode requires an operator pin for pre-capability schemas; a verifiable finding at or above fail_on blocks in strict mode; pins never override revocation. Acceptance criteria are on the element; satisfy every one.

Do:
1. RED FIRST through the production entry point (the audit decision as install/audit commands reach it): strict + pinned schema-3+ tree with a verifiable finding >= fail_on must block (red today); pinned pre-capability tree with no such finding -> allow; unpinned pre-capability -> require_pin; revocation blocks pinned content; advisory unchanged (warn). Exercise both the cache-hit and the fresh path.
2. Fix decideWithPins so a pin only satisfies the pre-capability require_pin rule and never waives a finding at or above fail_on in strict mode.
3. Mutants: restore the early `if pinned { allow }`, and apply the fix to only one of the two paths; both must be killed.
4. Spec: draft the one clarifying sentence for curator-spec profiles/manager.md §7 ("A pin satisfies only the pre-capability requirement; it never waives a finding at or above fail_on.") as an attached outcome resource; do NOT edit the spec repo.
5. Host rules: GOFLAGS=-work; syspolicyd counts; never edit LOGBOOK.md; no CHANGELOG edit.
Then `task-board handoff BUG-261004-16a407 --role developer` and END YOUR TURN.
