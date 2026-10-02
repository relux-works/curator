# Integration preconditions — revision 1

Bound run: RUN-261002-a8a02c; researcher / analyst. No landing command, status mutation, repository edit, commit, tag, release, or generic handoff was performed. The latest integration assignment delegates synchronous landing to the runner after producer exit.

Read-only checks and real exits:
- task-board spawn status: exit 0; bound run running.
- task-board spawn directives: exit 0; no directives.
- task-board q get status and bounded change_request activity: exit 0; integrating, revision 1 accepted.
- task-board worktree status STORY-261002-2327ef --json: exit 0; CR-TASK-261002-1pif8m-1 revision 1 accepted, kind story_final, producer researcher/analyst, current lease RUN-261002-a8a02c.
- git status --short: exit 0; only two untracked .research files.
- git diff --exit-code and git diff --cached --exit-code: each exit 0; no tracked/index changes.
- Python byte comparison using git show candidate:path and exact candidate changed-path assertion: exit 0; 2/2 files match accepted tree fb7ec3494ab0379049c00e9716cf808f3ab618f7 exactly. Candidate changes only .research/261002_rc14_rc3_readiness.md and .research/TASK-261002-1pif8m_evidence.json.
- shasum -a 256: exit 0; readiness a2947df483d9412939b24e13a4b4f3167af574f5766449f7fcb78d37ac28a2b5; evidence 06b6c5624bce94e13912e1c359122e08fe89e10d4518405da711009d7583f765.
- git ls-remote --symref origin HEAD refs/heads/main: exit 0; current advertised main f40b77c19c01746bda8b9a610358d860f2ad20c5, whereas accepted base is 2247509a476423bfdd9dd57a5098056fdd2be8cb. Upstream has advanced. Runner must perform authoritative fetch/equality, overlap and revalidation gates before landing; freshness and landing success are NOT attested here.

Discovery commands: skill/reference reads and worktree help exit 0; schema(operation=change_request) exit 1 (unsupported query, recovered via activity and worktree status); rg --files -g AGENTS.md -g !vendor -g !.task-board exit 1 (no matches). No tests rerun; prior validation and review resources remain attached and are not represented as fresh test evidence.

Producer preconditions confirmed: acceptance, story_final kind, bound role, exact research candidate, unchanged workspace. Runner-owned delivery checks remain pending, particularly upstream drift. Status remains integrating.
