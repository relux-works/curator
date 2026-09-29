# Integration preconditions — TASK-260929-jup8re rev1

- Board: task and parent story both at integrating.
- Review verdict rev1: ACCEPTED (candidate tree = worktree, base fc499a96).
- Worktree HEAD: fc499a96; working tree holds exactly 2 uncommitted paths (.github/ci/gate-selftest.sh, .github/ci/naming-gate.sh), matching the accepted candidate. No commit past checkpoint.
- No file changed by this run. No integrate/checkpoint executed: per the bound-producer assignment the runner lands synchronously after this run exits.
- Spawn directives for this run: none.
- Residual from brief/verdict: a forged exact-signature line is exempt; prose never matches.