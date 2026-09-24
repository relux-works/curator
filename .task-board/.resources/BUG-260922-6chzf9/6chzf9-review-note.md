# Review note — 6chzf9 CI flake fix (orchestrator, binding)

Review CR revision 1 against the task's AC and `6chzf9` results. Judge independently, read-only (disposable copy if you run anything):
1. Root cause is evidenced (hosted artefact / log lines), not guessed; the fix addresses THAT cause.
2. The fix does not weaken a product guarantee: no blanket retry/sleep, no widened timeout that hides a real hang,
   no test skip; a persistent/genuine failure still fails closed (check the refusal rows exist and execute).
3. Rows: transient + persistent + genuine-error rows are real (executed on the platform they target in the hosted
   gate — read the gate's test evidence for the platform-specific rows; "compiles on macOS" is not execution).
4. Apply one narrowing mutant of your own in a disposable copy (e.g. retry any error / retry forever / drop the
   bound) and show which row kills it.
5. Hosted gate green on the candidate (runtime's validation log).
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md writes.
