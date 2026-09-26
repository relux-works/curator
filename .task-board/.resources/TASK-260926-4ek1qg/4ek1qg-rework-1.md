# TASK-260926-4ek1qg — rework 1: partial-client lane for cocoaskills (THE ONLY CURRENT INSTRUCTION)

Revision 1 CHANGES_REQUESTED (verdict rev1): on PR #88 head b68a937 the Python manager lane (ivanopcode/cocoaskills) fails
tests/test_schema8_candidate_conformance.py::test_script_policy_sections_are_all_classified — it does not classify #88's
executable_identity_cases / hard_link_substitution_definition (run 36206606817).
OPERATOR DECISION 2026-09-26 (option 1): cocoaskills declares support for core 1.0.0-rc.10 and skillfile-sources-v1 and does NOT implement
rc.11–rc.12. The spec CI must check each implementation against what it claims:
1. Python/cocoaskills lane: run it against the core suite at tag v1.0.0-rc.10 (conformance/v1 from that tag — a second checkout or
   `git show`/worktree of the tag) PLUS the candidate's conformance/skillfile-sources-v1 (and schemas/skillfile-sources-v1); state exactly
   which env vars/paths the cocoaskills tests read and how each is fed. The Go/curator lane keeps checking the full candidate core.
   Record the per-implementation claim in the workflow comment and, if the coverage tooling has per-implementation declarations, there too.
2. Move the cocoaskills checkout from 3ecca1d (PR 43) to its current main 4a88aa0e (default-on schema 2, fresh-machine lock replay, pinned
   to the accepted skillfile-sources-v1 corpus at 5746367) — the cocoaskills orchestrator asked for it. If 4a88aa0e cannot pass the lane
   for a reason on their side, keep 3ecca1d and state why (they said it is not a blocker for them).
3. Reviewer's secondary note: state in results that TestPreflightRefusalCases covers the mandatory-control-unavailable refusal shapes
   (narrower than the old "every enforced shape refused before any worker surface") and either map the broader claim to its successor test
   at curator 0a628621 or record the narrowing as deliberate with the reason.
4. Prove it: in a disposable copy (candidate + b202b5d, i.e. the #88 content) run BOTH lanes' steps as the workflow would (bounded parts;
   host memory is tight) — exit 0; note what could only be verified by hosted CI (Windows sparse checkout).
Keep curator pin 0a628621, sparse checkout and the Go declarations from revision 1. CHANGELOG entry (unreleased) updated. No LOGBOOK.md.
Append "Revision 2 — partial-client lane", `resource update`, `task-board handoff TASK-260926-4ek1qg --role developer`. A write-boundary
`policy warn` block is a warning.
