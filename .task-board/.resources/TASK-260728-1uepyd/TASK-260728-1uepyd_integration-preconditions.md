# TASK-260728-1uepyd — integration preconditions (bound run RUN-261001-470cf8)

Role binding: developer / implementer. Accepted change request CR-TASK-260728-1uepyd-3, revision 3.
This run confirms preconditions and attaches evidence only. It invokes no landing,
sets no status, calls no handoff, and changes no repository file.

## Preconditions — all hold (exit 0 on every check)

1. Board state: task `integrating`, story `integrating`
   - `task-board q 'get(TASK-260728-1uepyd) { overview }'` → status `integrating`, parent STORY-260930-3feoy5.
   - `task-board q 'get(STORY-260930-3feoy5) { full }'` → status `integrating`, children `[TASK-260728-1uepyd]` (sole leaf).
2. Revision 3 accepted with green gate
   - `task-board worktree obligations` → `TASK-260728-1uepyd 3 accepted checkpoint 3h50m integrating STORY-260930-3feoy5`.
   - `TASK-260728-1uepyd_review-verdict-rev3.md` head: `review verdict — CR rev3 — ACCEPTED`, candidate base 5432c85f tree 5a1f7895 (5 paths).
   - `TASK-260728-1uepyd_change-request_rev3-validation.log` tail: remote gate run 36804719262 `success`, every lane success incl. Test (windows-latest); `[exit 0]`.
3. Worktree holds exactly the accepted candidate, uncommitted, no commit past base
   - Branch `task-board/story/STORY-260930-3feoy5`, HEAD `5432c85f` (= rev3 base).
   - `git status --porcelain=v1`: 3 modified (`.github/ci/conformance-case-counts.tsv`, `internal/buildrepo/admission.go`, `internal/buildrepo/admission_test.go`) + 2 untracked (`internal/buildrepo/acquisition_conformance_test.go`, `internal/buildrepo/testdata/` containing only `acquisitiongitshim/main.go`) — the same 5 paths as the rev3 patch `diff --git` list.
   - Exact-tree proof (read-only temp index seeded from HEAD, removed after): `read-tree HEAD` + `add -A -- .` + `write-tree` = `5a1f7895c9a17d181f021adbbbaae9c13c798fde`, matching the accepted tree. Real index untouched: status identical before and after.
4. No run directives: `task-board spawn directives RUN-261001-470cf8` → none recorded.
5. No writes by this run: no repository file created/modified, no board mutation issued (all checks read-only; temp index files lived in /tmp and were removed).

## Routing note for the runner (no action taken here)

- Obligations lists this revision as `accepted / needs checkpoint`, but the story has this task as its sole child leaf, and the run binding states a sole leaf integrates rather than checkpoints. Routing is left to the synchronous runner landing.

## Evidence commands (all exit 0)

- spawn status / directives; task + story overview/full queries; obligations; verdict head; rev3 validation-log tail; rev3 patch `diff --git` list; `git status`, `git log`, `find testdata -type f`; temp-index `write-tree` reproduction.
