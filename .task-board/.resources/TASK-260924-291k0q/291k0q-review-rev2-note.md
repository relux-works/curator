# Review note — TASK-260924-291k0q revision (option B, rc.13 bytes preserved) (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Orchestrator decision: option B + preserve rc.13 identity. The candidate changes only ci.yml, CHANGELOG.md, COMPATIBILITY.md and the gate +
its tests. Verify: (1) `git diff v1.0.0-rc.13 -- release/1.0.0-rc.13.json protocol/skillfile-sources.md schemas/skillfile-sources-v1
conformance/skillfile-sources-v1` is EMPTY on the candidate; (2) the gate passes on main and the candidate, with exactly ONE allow-listed
conditional-capability citation (environments §9.4 sentence in protocol/skillfile-sources.md) keyed on exact file + text with a reason;
(3) tests: allow-listed clause passes, the same citation without its condition fails, any other post-rc.10 citation fails, the $ref mutant
fails — reproduce at least two; (4) COMPATIBILITY.md states the rc.10 baseline and the single exception; CHANGELOG Unreleased;
(5) the Specification CI step runs the gate. accept_cr or changes requested with file:line. No LOGBOOK.md.
