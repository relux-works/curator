# Review note — TASK-260928-36r9k5 posture carrier (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev2: base 213a53e5 (= trunk minus the 31gaka landing 7444178d), tree 438bacad, 34 paths, gate green. It carries three things; review each:
A. TASK-260927-4pv4au security_posture revision A — ACCEPTED earlier (verdict rev3, tree c7e73fa2 on base 6bd98d49). Check the re-applied
   code is that accepted content modulo trunk context (compare `git diff 6bd98d49 c7e73fa2` with `git diff 213a53e5 438bacad` restricted to
   the 4pv4au paths): permissive default, warning once on stderr only and never on the enforced launch path, hardened effective defaults,
   locked precedence, refusals, status rows. Nothing weakened during the re-apply.
B. TASK-260910-1sapuy — NEVER REVIEWED: unreachable-registry-permissive-warns / -hardened-refuses through `curator install`/`update`
   (internal/install, internal/registry, cmd/curator; results in TASK-260910-1sapuy_results.md). Full review against profiles/manager.md §7.1,
   registry evidence clauses and vectors/security-posture.json: the permissive gate notice names every artifact resolved without registry
   evidence; hardened refuses fail-closed before any write; the effective-posture API from part A is used (no re-derivation); mutants killed
   (notice removed; hardened downgraded to warn) — real exit codes.
C. Rev2 test/golden updates: global_adopt_test, global_lock_publication_test, profile_delta_confirmation_test and the two goldens now
   expect the spec-mandated permissive warning exactly once — confirm no test was weakened to ignore stderr wholesale; the 3 removed gap rows
   (hardened schema cases) now pass.
No trunk revert (rebase fidelity incl. rows trunk removed), no CHANGELOG/LOGBOOK, no stray files. Bounded runs. accept_cr or changes
requested with file:line. No LOGBOOK.md.
