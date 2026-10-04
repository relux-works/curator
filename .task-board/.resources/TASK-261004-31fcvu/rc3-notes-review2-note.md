# Review note — TASK-261004-31fcvu rev2 delta review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Delta review of revision 2 against your rev1 verdict (`TASK-261004-31fcvu_review-verdict-rev1.md`); keep everything rev1 already passed in force.

1. P1: all 27 previous Unreleased entries are back, byte-for-byte as at the base, in their original order, under `### Shipped in 0.15.0-rc.2 but not recorded in its notes` at the end of the rc.3 section, with the attribution sentence. Prove it per entry. The rc.2 section and everything below are byte-identical to the base.
2. P2: no internal runner, host or machine name, user path or employer name anywhere in the rc.3 section (scan the whole section, not only the old lines 84–86).
3. The validator now fails rev1: rerun it against the rev1 candidate and confirm exit 1, then exit 0 on rev2.
4. Nothing else changed versus rev1 except these fixes; only CHANGELOG.md is in the delta; LOGBOOK.md untouched.

accept_cr if all hold; otherwise changes requested with exact lines.
