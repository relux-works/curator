# Review note — TASK-260930-38fjt0 pin Test (rose-air) to the rose-air label (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 0e3169bb, tree 8311603e, 3 paths, gate green) against `38fjt0-brief.md`. Verify that:
- runs-on is exactly [self-hosted, macOS, ARM64, rose-air], and that no other self-hosted job exists or is also unpinned;
- the gate-selftest row asserts the label, and the label-removed mutant is killed (re-run it; give the real exit code);
- the comment and docs are accurate;
- nothing else changed.

accept_cr, or changes requested with file:line. Never spell any employer name.
