# Integration preflight — curator v0.15.0-rc.5 release notes

Run: RUN-261008-6a9451. Accepted CR-TASK-261008-2wx7vm-1 revision 1 remains integrating. No repository files changed by this run.

Fresh read-only evidence:
- task-board worktree integrating --json: exit 0; revision 1 accepted, kind story_final, repository_delta present, awaiting_landing, not deferred. Candidate tree 5d22df33b0e662be8539b14beb0dc73dc348073b. Fresh protected main OID 9757765dca0dab09762d8de6f0bb616d6803a66c. Candidate not yet on protected trunk.
- Board task/children query: exit 0; this task is the only Story child and is integrating.
- git status --short and git diff --name-only HEAD: exit 0; only CHANGELOG.md modified; LOGBOOK.md untouched.
- git diff --check: exit 0.
- git diff --exit-code 5d22df33b0e662be8539b14beb0dc73dc348073b -- . ':(exclude).task-board': exit 0; tracked product working files match accepted candidate.
- task-board spawn directives: exit 0; no directives.
- task-board worktree status --json: terminated after remaining unresponsive without output; real exit 143. This broad diagnostic did not pass. The independent integrating query above returned successfully.
- Initial schema(operation=change_request) lookup: exit 1, unknown operation; recovered through schema discovery. No validation claim attached to this lookup.

Existing evidence inspected: TASK-261008-2wx7vm_change-request_rev1-validation.log (resource read exit 0) records sh scripts/remote-gate.sh exit 0 and successful hosted run 37725316422. Accepted as prior evidence, not rerun here. No build or test suite rerun because this integration preflight changes no code or documentation.

Latest Integration Assignment supersedes earlier FIRST/LAST and manual integrate instructions. No status mutation, generic handoff, checkpoint, or integrate invoked. Runner owns synchronous landing and its final freshness/validation gates after producer exit. This is preflight evidence, not a landing receipt. Unrelated board debt was reported; no unrelated files touched.
