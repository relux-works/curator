# Integration preconditions — research-pi-opencode-tool-disable

Bound run: RUN-261010-1cc446. Read-only precondition inspection; landing is delegated to the synchronous runner under the latest Integration Assignment. No integrate, checkpoint, handoff or status mutation was invoked.

## Confirmed evidence

- Task query: integrating (exit 0).
- Worktree status: CR-TASK-261010-aqpf2a-2, revision 2, accepted, story_final; producer researcher/analyst; reviewer RUN-261010-4ba650; active lease RUN-261010-1cc446 (exit 0).
- Base: 5ec599f0b7eddeb3403fe89104b7b800fd0a55db. Candidate tree: 73547d1b1b4175fc697c8d4bf42fa3316b8ac579.
- git diff --name-status BASE CANDIDATE: exactly one added path, .research/261010_pi-opencode-tool-lockdown.md (exit 0).
- git diff --exit-code BASE -- LOGBOOK.md: empty (exit 0).
- git hash-object of working research: 1a285a1a66973de80ab622c7060db3e7bf5cb22b (exit 0).
- Standalone Python byte comparison of rev1 tree c3285cf15b77bedfabbd7a458ad533ce688c9114, rev2 candidate and working research: identical, 506 lines (exit 0).
- git status --short: only the research document is untracked (exit 0), consistent with the uncommitted candidate.
- Retrieved TASK-261010-aqpf2a_review-verdict-rev2.md (exit 0): accepted; content acceptance retained from rev1.
- Retrieved TASK-261010-aqpf2a_change-request_rev2-validation.log (exit 0): hosted run https://github.com/relux-works/curator/actions/runs/38013607830 reports success and gate exit 0; required command shards 1/1 green, failed 0, missing 0; test-case coverage unknown; two optional jobs skipped. Accepted from existing evidence, not rerun.

No builds, tests, harness installs, harness runs or login performed. No repository files edited. Fresh protected-authority checks and landing remain the runner transaction responsibility; this evidence does not assert a landing.

Discovery diagnostics: unsupported task query, change-request help command and resources field each exited 1; corrected to get, worktree status and resource get, all exit 0. These were CLI discovery errors, not validation passes.
