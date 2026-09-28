# TASK-260924-291k0q — gate skillfile-sources independence from post-rc.10 core (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description (updated 2026-09-26: paths are now schemas/skillfile-sources-v1 and
conformance/skillfile-sources-v1; v9 is not in the accepted suite). Implement the gate in tools/ (Python, like the existing validators) and
wire it into the Specification CI; unit tests incl. two mutants: (a) repoint one skillfile-sources $ref to a v1 definition absent/different
at v1.0.0-rc.10 → gate fails; (b) add a prose citation to a post-rc.10 clause → gate fails. Show the gate passing on main (real exit
codes). CHANGELOG entry under Unreleased (this repo keeps its CHANGELOG). No LOGBOOK.md. Attach results, check DoD,
`task-board handoff TASK-260924-291k0q --role developer`. A write-boundary `policy warn` block is a warning.
