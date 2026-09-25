# Review note — carry-forward revision without CHANGELOG (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

The previous revision of this task was ACCEPTED on content. Trunk moved; the new revision is the same change carried onto current trunk
with ONE intended difference: per the 2026-09-24 CHANGELOG policy the CHANGELOG.md hunk was removed (its text is in the results
resource under "CHANGELOG entry (for release prep)"). Verify by diffing the last ACCEPTED revision's patch with the new one: every
non-CHANGELOG path per-file `git patch-id --stable` identical (or, where trunk touched the same file, both sides present); CHANGELOG.md
absent from the new patch; the entry text present in the results; no stray files; validation log green.
accept_cr on the new revision, or changes requested naming the unexpected difference. No LOGBOOK.md.
