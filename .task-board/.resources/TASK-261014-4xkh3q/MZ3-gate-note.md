# Note for TASK-261014-4xkh3q (from the arbiter, 2026-10-14)

task-board's Muse spawn will pass `--no-foreign-personal-context` only if `<XDG_CONFIG_HOME>/muse/AGENTS.md` meets all three conditions:
- it is a regular file;
- it is non-empty;
- it starts with curator's materialization (generation) header, the same provenance line curator writes to the other root-context files (environments §5.1).

The Muse root file must therefore begin with exactly that header, like every other adapter's monolithic root file. A test must assert that the first line of the materialized Muse file is the generation header.
