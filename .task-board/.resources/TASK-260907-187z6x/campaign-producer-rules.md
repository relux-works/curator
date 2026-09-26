# Campaign producer/reviewer rules — unified Curator delivery, host e11-1 (2026-09-15)

These rules supersede stale path and policy wording in older briefs. Older briefs remain authoritative for product scope.

## Where things are on THIS machine
- Authoritative board: /Users/administrator/Developer/ReluxWorks/curator/curator/.task-board (TASK_BOARD_DIR is set for you; never run task-board against a worktree copy of .task-board, never edit board bytes by hand; use `task-board m`/`resource add`).
- Code repositories (control roots): curator = /Users/administrator/Developer/ReluxWorks/curator/curator; curator-agent-launcher = /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher; curator-spec = /Users/administrator/Developer/ReluxWorks/curator/curator-spec. Historical `/Users/iv/...` paths in older resources describe another machine; translate them to these roots. Old `.temp/STORY-*/worktree` candidates from that machine do not exist here; the accepted content survives as `<ID>_change-request_rev<N>.patch` outcome resources and, where landed, on each repository's main.
- The approved Skillfile source contract is LANDED on curator-spec main (`a4fcaf02` "Specify opt-in local and Git Skillfile sources", plan in `3535d63e`): protocol/skillfile-sources.md, protocol/repository-transport.md, schemas and conformance under draft-sources-v1, docs/skillfile-sources.md. Read it from the curator-spec checkout above, not from a worktree path.
- Work ONLY inside your assigned Story execution worktree (`<control-root>/.temp/<STORY-ID>/worktree`). No writes into the control root, no LOGBOOK.md edits, no writes to ~/.agents, ~/.claude, ~/.codex, ~/.curator, ~/.config/curator-run — EXCEPT when a task brief explicitly assigns host-state work (e.g. B5 onboarding TASK-260908-yl5x3k: `curator profile install/use` writes under ~/.curator and backups under ~/.curator/backups are the deliverable; the brief lists the exact commands). No installs of task-board/curator binaries, no daemon restarts, no tags/releases, no real ax calls (fake ax in tests only), no interactive permission bypass.

## Models, review, delivery
- **Since 2026-09-23 (skill-project-management#359): every run uses `--context-profile full`.** The `lite` profile drops the role body (Standing Orders, Evidence Honesty Contract, Status Transitions; the reviewer's review-round contract; the producer's handoff preconditions). If your prompt lacks your role's body, say so in your results before doing anything else.
- **Since 2026-09-23 (operator directive, supersedes every line below about models):** implementation/producers run on Codex `gpt-6-luna` at `max` (`--agent codex --model gpt-6-luna --reasoning-effort max`); independent reviews run on Claude `claude-opus-5-5` at `low` (`--agent claude --model claude-opus-5-5 --reasoning-effort low`). Bound runs unchanged (muse xhigh lite).
- Since 2026-09-22 (operator directive): independent reviews run on Codex `gpt-6-astra` at `low` reasoning effort (`--agent codex --model gpt-6-astra --reasoning-effort low`); bound runs (checkpoint/integrate/republish/complete) on Muse xhigh lite. Producers unchanged (next line).
- Since 2026-09-18 14:30Z (operator directive, supersedes 2026-09-16): coding producers run on Muse muse-spark-1.3-contributor at max reasoning effort (xhigh only as a stream-idle mitigation); independent reviewers run on Claude claude-opus-5 at max reasoning effort (Codex gpt-6-astra:low is retired for reviews because of provider limits; claude-fable-5-1:low remains admitted as a producer fallback). Do not change user-facing launcher defaults to mirror worker settings.
- Hand off through `task-board handoff <ID> --role <role>` after attaching a task-scoped outcome resource (`<ID>_results.md` or similar). The runtime publishes the Change Request and runs the configured landing suite EXACTLY ONCE; do not run the full landing suite manually in addition. Run narrow package tests with real exit codes (`set -o pipefail`, state the shell) and cite them.
- Coverage evidence: a row is "driven" only when a committed test reaches the production entry point named in the AC. Helper-direct tests, pin checks and inspection are bounds. Attack gates with narrowing mutants; report survivors with bounds.
- Reviewers: verify the exact candidate tree, rerun narrow tests independently, attack the gates, and record exactly one verdict through `accept_cr(<ID>, revision=<N>, evidence=<your own outcome resource>)` or a changes-requested verdict routed by `set_status`. Never accept on the producer's evidence alone.
- The parent (orchestrator) owns signed branch → PR → hosted checks (GitHub + rose-air self-hosted lane where configured) → exact-reviewed-head fast-forward to main → board closure. Commits are signed by the integration path with the configured human identity; do not commit on the managed Story branch yourself.

## Boundaries that stay in force
- Frozen v1 protocol schemas and release qualification unchanged unless the task says otherwise; opt-in draft support must be labelled as draft.
- Rules/knowledge types, generated/private instruction files, generalized MCP endpoint/plugin imports, prebuilt skill-CLI distribution, compiler bootstrap, new Git aliases/mirrors/ports remain deferred; they are not hidden acceptance criteria.
- Pi has no MCP channel (environments 1.1 §7.8); never invent one. Pi runtime preference for launcher defaults is an ordered convention (pi-anthropic, pi-openai, pi-google) recorded by the orchestrator.
- Cross-platform claims need real platform evidence (hosted ubuntu/macos/windows lanes, rose-air ARM64); report unavailable platforms as unverified, never as passing.

## Write-boundary warnings are NOT refusals (2026-09-23)
This project runs `spawn.worktree_isolation.run_write_boundary` at its default `warn`. When `task-board
handoff` prints `run_wrote_outside_worktree: … policy warn violated=true` with thousands of paths, that is a
loud WARNING caused by other runs and the orchestrator writing concurrently — not your refusal. Decide the
outcome from the board, not from that text: `task-board q 'get(<ID>) { id status }'` and the Change Request
revision it published (a new `<ID>_change-request_revN.patch` + `-validation.log` resource). If the status is
`to-review` and the revision exists, the handoff succeeded — say so in results and end. Only a status that did
not move is a refusal. Never attach thousands of warning lines; summarise them in one line.
