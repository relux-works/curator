# Resume 4 — TASK-260910-3kvq02 (re-anchor onto current main c1aa0d2 before integration)

Revision 3 was ACCEPTED, but curator main advanced by a board-only commit (c1aa0d2) after acceptance and the runtime cannot revalidate during integration (task-board BUG-260916-20l47k), so the candidate must be re-anchored on current main and republished; the reviewer will re-accept and the orchestrator integrates immediately. The orchestrator captured your full candidate — six owned paths, now INCLUDING the Windows output-boundary fix (`checkOutputBoundary` ancestor walk in internal/manifest/expand.go and its test) — as the precondition resource `TASK-260910-3kvq02_collections-candidate-rev1.patch` (supersedes the rev0 patch), and cleaned the worktree so this spawn's final-leaf refresh replays the parser checkpoint onto current main.

Do:
1. `git log --oneline -3` must show the parser checkpoint (TASK-260910-24cuys) on top of current main (c1aa0d2 or later). If not, attach `task-board worktree status STORY-260910-197y84` output and stop.
2. Fetch and apply `TASK-260910-3kvq02_collections-candidate-rev1.patch` (`task-board resource get ...`; `git apply --check`, `git apply`). Zero conflicts expected.
3. Narrow verification: `go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers`, vet, golangci-lint, gofmt -l on the six files, `git diff --check`; confirm `TestExpandOutputReadFailure` passes and that `checkOutputBoundary` is present. Quote exit codes in TASK-260910-3kvq02_resume4-results.md.
4. Tick the checklist; `task-board handoff TASK-260910-3kvq02 --role developer` (foreground). The runtime runs the remote gate after the handoff; do not background anything; do not write board resources after the handoff.
