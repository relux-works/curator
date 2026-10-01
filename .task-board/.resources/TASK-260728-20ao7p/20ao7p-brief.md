# TASK-260728-20ao7p — native black-box test + author guide, fastest path (THE ONLY CURRENT INSTRUCTION)

Keep this small. Ivan asked for the fastest path.
1. Add cmd/curator/native_blackbox_test.go. Build the real curator binary once, via the existing testcli helper if there is one. Then,
   against a local file:// external build repository fixture (a tiny Go command package in a temp git repo; no network) and a schema-7
   skill that declares it:
   - install: build, cache, shim;
   - run the shim and check its output;
   - install again (cache hit, no rebuild);
   - remove it, checking that the shim and the build root are gone.

   Assert exit codes and key outputs. It must run on the hosted macOS and Windows lanes (and Linux). Reuse existing fixture helpers
   from internal/install / internal/buildrepo tests. Do not add a skip unless the platform-case ledger already permits it, with the
   exact reason.
2. Add docs/external-build-repositories.md in curator: a short author guide. It covers declaring build_repositories in schema 7, the
   locked commit and tag, local development substitution, and what curator verifies (admission → audit → cache → compiler). Link it
   to the spec's docs/external-build-repositories.md for the full rules, and add a link from README.md.
3. Run `go test ./cmd/curator -run NativeBlackbox` with the real exit code; the hosted gate is the arbiter for Windows. No
   Windows-reserved file names. No CHANGELOG/LOGBOOK: put the entry text in the results. Never spell any employer name.
Update the results, then run `task-board handoff TASK-260728-20ao7p --role developer`, then END YOUR TURN.

## First step (binding)
Before anything else run: `task-board m 'set_status(TASK-260728-20ao7p, status=development)'`. Two earlier runs ended without setting it, so the runner refused their handoff. Build the test and the doc, then hand off.

## Update (binding)
The hosted-lane checklist item was removed. The runner gate produces that evidence after handoff. The work is already in the worktree: verify it, set status development, update the results, run `task-board handoff TASK-260728-20ao7p --role developer`, then END YOUR TURN.
