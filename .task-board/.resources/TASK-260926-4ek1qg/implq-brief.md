# TASK-260926-4ek1qg — qualify the Implementations workflow against curator main (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Lockstep partner: PR #88 branch land/TASK-260924-mcmova — commit b202b5d
(accepted erratum on main 5746367) + 94e8c78 (orchestrator's pin-only attempt; your change REPLACES it). Evidence: run 36202537423.
1. .github/workflows/implementations.yml: curator ref → 0a62862130fd6cbd9db8b246da63fdb5aecdfaaf, accurate comment (what it classifies, why
   lockstep with #88). Windows: make the curator checkout succeed (prefer a sparse checkout that excludes .task-board/, or core.longpaths —
   state the choice and why); keep cocoaskills and registry pins unchanged.
2. tools/implementation_coverage.py (and any declaration data it reads): bring the Go declared consumption cases in line with curator
   0a628621 — for each declared test, verify it exists (`go test -list` in a disposable curator clone at 0a628621) and maps to the same
   claim; replace renamed ones with their successors (e.g. whatever now proves "every enforced shape is refused before any worker surface"),
   never just delete a claim; list old → new.
3. Prove it: in a disposable copy of your candidate with b202b5d cherry-picked (the #88 content), run the Implementations job's Go steps
   against curator 0a628621 (bounded parts; host memory is tight) and implementation_coverage.py on the resulting stream — exit 0.
4. CHANGELOG entry under unreleased (this repo keeps its CHANGELOG). No LOGBOOK.md.
Attach results (old → new case table, commands with real exit codes), check DoD, `task-board handoff TASK-260926-4ek1qg --role developer`. A write-boundary
`policy warn` block is a warning.
