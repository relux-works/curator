# Review note — TASK-260910-32gki6 rev8: fidelity + merge review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

You ACCEPTED rev5 (base 213a53e5, tree 0fd98b72, 31 paths; saved as `refs/campaign/148pj1-rev5-20260929`). Trunk moved, and the delta was
re-applied. Rev8 (base 0a638288, tree ad18cefb, 31 paths) is green on every lane. It has the same +/- line multiset as rev7, which merged
`internal/pathboundary/pathboundary.go` with BUG-260928-uyak0e (vanishing-entry skip).
1. The path sets are equal:
   `git diff --name-only 213a53e5 refs/campaign/148pj1-rev5-20260929 -- . ':!.task-board'` vs
   `git diff --name-only 0a638288 ad18cefb -- . ':!.task-board'`.
2. For every path except the conflict paths (pathboundary.go, conformance-case-counts.tsv, envprofile/status.go), the +/- line multisets
   are identical. Report any difference.
3. On the conflict paths, check three buckets and nothing else:
   (a) lines in neither side;
   (b) trunk lines dropped;
   (c) ledger rows trunk removed but rev8 re-added.
   Each must be empty or explained by a both-sides resolution.
4. pathboundary.go merge, the substantive part:
   - uyak0e's rule is kept: a transient child that vanishes during a tree walk (ENOENT between readdir and lstat) is skipped;
   - S5's rule is kept: a missing root, lock, marker or lock-named store entry still fails closed with the S5 failure class.

   Run `go test ./internal/pathboundary` and `go test ./internal/envprofile -run 'Store|Boundary|Pin|Resolve|Status|Legacy|PathInstall'`
   and give the real exit codes. Mutant: make the vanishing-entry skip also apply to a named store entry, then show that an S5 row fails.
accept_cr, or changes requested with file:line. No LOGBOOK.md. Never spell any employer name.
