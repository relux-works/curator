# THE ONLY CURRENT INSTRUCTION — TASK-260924-4mzun5 rework: keep trunk's draft-sources-v1 copies, add the v2 copies beside them (developer)
Your revision could not be published: `change_request_candidate_reverts_trunk` — the candidate restores 24 trunk-changed paths to older content, all under `internal/skillspec/testdata/draft-sources-v1/` and `internal/marker/testdata/draft-sources-v1/` (for example `install-marker-v5.schema.json` `skill_schema_version.maximum` 9 → 8, and the agent-skill-v9 / csk-skill-v9 schema cases). The board refuses any candidate that silently reverts trunk.
**Do, in the workspace as it is now (do NOT checkout, reset or converge):**
1. Restore every path under `internal/skillspec/testdata/draft-sources-v1/` and `internal/marker/testdata/draft-sources-v1/` to trunk content: `git checkout HEAD -- internal/skillspec/testdata/draft-sources-v1 internal/marker/testdata/draft-sources-v1` (this is the only checkout allowed, scoped to those two directories).
2. Put the copies of the spec's new draft suite under `.../testdata/draft-sources-v2/` (from curator-spec main 7eaeb73f) and point the new tests at v2. If an existing test needs the v1 copies to change, do not change them here; say so in the results and leave it for a separate task.
3. Keep the rest of your work (legacy-lane decision, marker/audit changes, production-entry rows, glob vectors).
4. R223: no local `go test`; compile-only (`go vet ./...`, `go build ./...`). The hosted gate is the arbiter.
5. Results resource (plain text): what changed versus your previous attempt.
Then `task-board handoff TASK-260924-4mzun5 --role developer` and END YOUR TURN.
