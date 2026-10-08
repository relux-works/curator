# THE ONLY CURRENT INSTRUCTION — TASK-260927-1e5qqm rework 3: re-apply one doc hunk on the post-posture-B main and republish (developer)
**What happened.** Your revision 2 (the re-application on 75ab9a71) passed the hosted gate. Before it could be reviewed, posture-B (TASK-260927-25hk87) landed on main (35cac659), so the orchestrator converged your workspace onto the new main. Five of your six files were carried unchanged (the test files merged cleanly). ONE file conflicted and was reverted: `docs/ci-gates.md`; your original delta for it is attached as `seedB-cigates-delta.patch`.
**Do, in the workspace as it is now (do NOT checkout, reset or converge):**
1. `docs/ci-gates.md`: re-apply your intent from the patch onto the NEW file (posture-B edited the same doc). Keep posture-B's text; add yours next to it.
2. Check that `cmd/curator/env_credential_marker_test.go` and `cmd/curator/envstatus_test.go` still carry your seed-B intent after the merge with posture-B's edits (read them; do not rewrite).
3. Targeted tests ONLY through `~/.local/bin/mini-build-lock run seedB -- env GOFLAGS=-work go test ./internal/envprofile ./internal/envregistry -run 'Seed|Codex' -count=1 -timeout=6m`. No cmd/curator locally (R194).
4. In your results resource: "re-applied after converge (2)", the file and what changed versus revision 2 (expected: nothing semantically).
Then `task-board handoff TASK-260927-1e5qqm --role developer` and END YOUR TURN. No LOGBOOK, CHANGELOG or scripts/remote-gate.sh edits.
