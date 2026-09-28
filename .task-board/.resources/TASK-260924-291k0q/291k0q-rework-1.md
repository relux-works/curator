# TASK-260924-291k0q — rework 1: option B (THE ONLY CURRENT INSTRUCTION)

Review rev1 (verdict) found that the released rc.13 suite text normatively but CONDITIONALLY cites environments §9.4 ("Machine-global
Skillfiles follow environments §9.4 profile locks when the manager implements that capability") — conditional capability, so a partial client
on core rc.10 without profile locks is unaffected. Orchestrator decision: OPTION B.
1. Revert release/1.0.0-rc.13.json, protocol/skillfile-sources.md and conformance/skillfile-sources-v1/manifest.json to main's bytes (they
   must equal the v1.0.0-rc.13 tag; show `git diff v1.0.0-rc.13 -- <file>` empty for each).
2. The gate keeps rejecting post-rc.10 citations EXCEPT an explicit, tested allowlist of conditional-capability citations: exactly one entry,
   keyed on the exact file + clause text (environments §9.4 in protocol/skillfile-sources.md, the sentence above), with a reason string.
   Tests: the allowlisted clause passes; the same citation with the condition removed fails; any other post-rc.10 citation fails; $ref
   mutant still fails. The gate passes on main unchanged.
3. README/CHANGELOG (Unreleased) state the rc.10 baseline and the one conditional exception.
No LOGBOOK.md. `task-board m 'set_status(TASK-260924-291k0q, status=development)'` first; append "Revision 2 — option B", `resource update`,
`task-board handoff TASK-260924-291k0q --role developer`. A write-boundary `policy warn` block is a warning.
