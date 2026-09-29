# Review note — BUG-260928-uyak0e path boundary walk vs vanishing entry (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 3f60f7f0, tree 7096aa0a, 3 paths, gate green incl. Windows) against `pathwalk-bug-brief.md` and `pathwalk-gatefix-1.md`
and environments.md §4 "path source directories". Verify: (1) an entry that no longer exists when examined (fs.ErrNotExist from ANY per-entry
probe — Lstat, Windows reparse/DACL/owner probes, directory open for recursion) is skipped; every other error stays a failure (fail
closed); (2) an entry replaced by a symlink/special file in that window is still refused; (3) git fixtures disable auto-maintenance
(gc.auto=0, maintenance.auto=false); (4) mutants: "treat every error as skip" and "skip vanished only in Lstat" killed with real exit codes;
the original flake (TestPathInstallCapturesDirtyUntrackedInsideGit) repeated ≥10 times locally green. Security judgement: skipping a vanished
entry cannot admit an unchecked file into the store (the store is built separately and pinned) — confirm against the code. accept_cr or
changes requested with file:line. No LOGBOOK.md.
