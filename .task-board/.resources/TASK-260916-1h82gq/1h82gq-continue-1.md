# TASK-260916-1h82gq continuation 1 (orchestrator, binding)

Your previous run (RUN-260921-2fb4fe) hit the 150-minute launcher timeout before any handoff. The
Story workspace holds your uncommitted work (28 paths: `internal/scriptworker` capabilities.go,
exec*.go, launcher.go, derive_test.go, forgeworker/stubinterp fixtures, install/scriptpolicy/
runtimestore edits, docs, CHANGELOG, ledger) — do NOT recreate, checkout, clean or stash anything;
continue from that state (`git status` first).

Order of work (budget discipline — this run also has 150 minutes):
1. In the first 10 minutes write `results.md` (attach as outcome resource
   `TASK-260916-1h82gq_results.md` NOW, update later): what is implemented, what is not, which
   rows exist and pass locally, which mutants were run. This protects the work if the budget runs
   out again.
2. Finish the smallest complete slice of the brief 1h82gq-brief.md that makes the gate green:
   `go build ./... && go vet ./internal/scriptworker/ ./internal/scriptpolicy/ ./internal/install/`,
   then `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1` and the install rows
   you touched. Anything the brief asks for that is not finished goes into results.md as an explicit
   "undone" list with the reason — an honest partial is acceptable (brief R5); a silent partial is not.
3. Publish the Change Request (handoff) as soon as the tree is coherent and the gate can run; do not
   start new scope in this run.
Rulings R1–R5 of 1h82gq-brief.md unchanged. If a muse stream idles, the runtime spawns a successor —
keep files consistent at every step (no half-written edits) so a successor can continue.
