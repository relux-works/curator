# TASK-260924-1ytz0u — finish the release-prep handoff (THE ONLY CURRENT INSTRUCTION)

The orchestrator replaced checklist item 2 (the @v0.1.0 install needs a tag that only exists after landing): it now requires
`make check` exit 0 and a clean-cache `go install …/cmd/curator-run@<your candidate commit as a pseudo-version or the pushed branch
commit>` — if the candidate commit is not fetchable by the Go proxy, install from the local worktree module with `GOFLAGS=-mod=mod
GOPROXY=off go install ./cmd/curator-run` in a clean GOPATH/GOMODCACHE and say so. The signed tag + `@v0.1.0` install are verified by
the orchestrator after landing.
1. `task-board m 'set_status(TASK-260924-1ytz0u, status=development)'` if needed. 2. Run the checks, record them in the results resource
(`resource update`), check the items citing it. 3. `task-board handoff TASK-260924-1ytz0u --role developer`; stay in the turn during the
gate. Keep `git status --short` to product/docs/CHANGELOG/.github paths only; no artifacts or task documents in the worktree.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
