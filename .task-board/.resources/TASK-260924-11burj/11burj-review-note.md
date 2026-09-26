# Review note — TASK-260924-11burj playbook collection acceptance (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision against `11burj-brief.md`, `11burj-decision-1.md` and the binding operator memo
`skillfile-operator-memo-20260924.md`. Candidate changes 2 paths: internal/crossconformance/draftsources_playbook_acceptance_test.go (new)
and internal/install/draftsources.go (PRODUCTION change — name it, judge whether it is a real defect fix the acceptance exposed, whether it
is correct and minimal, and whether it has its own killing row).
1. The scenario drives the REAL production entry (built CLI binary or the cmd/curator entry), not internal calls; fixtures are local bare
   repos with the §8 layout (skills/orchestrator, skills/developer, …; a subfolder skill in a second repo).
2. Memo criteria: ONE collection entry {from, directory: "skills", include: ["*"]} installs every skill in skills/ with lock and audit;
   a manifest dependency with `directory` installs the subfolder skill of the other repo; new skill after a new tag appears via
   `project refresh` + `install` with lock/audit updated; fresh-home replay from the committed lock (lock byte-identical, tamper →
   source_snapshot_changed, unreachable → source_snapshot_unavailable); negative rows.
3. The Windows platform skip-class correction (results "Platform skip-class correction") — declared class, exact reason, not hiding a
   real Windows failure.
4. Release-note text for both capabilities in results; no CHANGELOG edit; no stray files; hosted gate green on all lanes.
Kill one mutant of your choice in the scenario (bounded, focused; host memory is tight). accept_cr or changes requested with file:line.
No LOGBOOK.md.
