# THE ONLY CURRENT INSTRUCTION — BUG-261004-bknio5 (N1): GC sweeps live runtime on an uncertain reference set (developer, code)

Source: docs/security-audit-2026-10-inline.md §N1 (issue #106); independently confirmed by the cocoaskills comparison (manager §10, profiles/manager.md: uncertain entries are retained). Acceptance criteria are on the element; satisfy every one.

Defect: internal/scopes/gc.go Collect calls sweepRuntime before checking marked.uncertain; a failed consumers.json read only appends uncertainty, invalid markers are omitted, and unmarked runtime dirs are deleted. The uncertainty guard protects only the build cache.

Do:
1. RED FIRST through the production CLI entry (`run([gc], ...)` with real locks and a working shim from runtimestore.WriteBinShim): truncated consumers.json; invalid install marker; unreadable install marker; a repeated GC after each. Each must show the shim works before GC. Record the red exit codes on current main.
2. Fix: when the live reference set is not proven complete, remove NO runtime (and no build-cache) object; warn that the runtime sweep was skipped and why. Keep conservative registry/consumer handling. Do not weaken normal removal of genuinely unreferenced runtime (control row with a complete set).
3. Mutants: reorder the uncertainty check after the sweep again, and drop one uncertainty source; both must be killed by your tests.
4. Host rules: GOFLAGS=-work for every go command; record syspolicyd successive-crash counts before/after long runs; never edit LOGBOOK.md; no CHANGELOG edit (the release-notes step writes it).
Then `task-board handoff BUG-261004-bknio5 --role developer` and END YOUR TURN.
