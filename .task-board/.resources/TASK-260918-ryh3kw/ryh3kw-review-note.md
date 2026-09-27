# Review note — TASK-260918-ryh3kw §8.4.1 read-failure discipline (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev4 (base 0ffe2e1d, tree 93f1102d, 18 paths, hosted gate GREEN incl. Windows) against `ryh3kw-sec-brief.md`, `ryh3kw-decision-1.md`,
`ryh3kw-gatefix-1.md`, `ryh3kw-gatefix-2.md` and curator-spec v1.0.0-rc.13 environments §8.4.1 (absence vs read failure; cite clauses).
Verify through the production entry:
1. internal/stateread classification: genuine absence (ENOENT; Windows ERROR_FILE_NOT_FOUND / ERROR_PATH_NOT_FOUND with every existing
   ancestor a directory) → absent; ENOTDIR / an existing non-directory ancestor / EACCES / other I/O → unreadable, fail closed. The
   ancestor walk must not follow links in a way that changes the verdict; check its cost on hot paths (bounded walk).
2. Manager-state reads at the §8.4.1 sites route through stateread (the 348-site inventory in results; spot-check call sites, and that
   TestManagerOwnedAbsenceReadsAreGuarded passes).
3. environments-read-failure published cases: 28/28 driven (incl. passthrough-parent-not-directory-unreadable); the 2 restore vectors
   parked for TASK-260927-1wc76r — 1wc76r LANDED on trunk (38c68570): these rows must be driven after the landing carry, or attributed to
   the real blocker; flag it if the ledger still names 1wc76r.
4. Mutants: ancestor walk removed; ENOTDIR folded into absence; one site reverted to os.IsNotExist — each killed (real exit codes).
5. No trunk revert vs base, no CHANGELOG/LOGBOOK, no stray files.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
