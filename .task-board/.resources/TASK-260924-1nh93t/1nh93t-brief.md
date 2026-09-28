# TASK-260924-1nh93t — F-M1e: launcher capability closure (THE ONLY CURRENT INSTRUCTION)

Control root: /Users/administrator/Developer/ReluxWorks/skill-agents-management (shared with the operator — only your Story
worktree). Read `campaign-producer-rules.md`. Board validation at handoff.

F-L1b (launcher TASK-260922-2u5jzw) has stopped THREE times on missing module capabilities (conflict table → v0.5.20; vendor
admission + stored policy → v0.5.21; now: the headless detector needs to know whether a native-argument suffix selects a
non-interactive form — the parser is in `internal/nativeargs`, unreachable from the launcher). Stop the drip:
1. AUDIT first: read the launcher SPEC 0.5.0-draft §4 (file `SPEC.md` in
   /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher/.temp/STORY-260922-39hxog/worktree, read-only) and
   decision 0018 (curator-spec main). List EVERY behaviour the launcher must get from this module (spelling, grammar, admission,
   conflict, drift, stored policy, non-interactive detection, anything else) with the public API that satisfies it today or
   "MISSING". Put the table in results.
2. Close every MISSING row publicly. At minimum: a versioned public classifier (system, verified tool release, native-argument
   suffix) → non-interactive yes/no (+ which form), reusing `internal/nativeargs` and the same per-release grammar versioning;
   unknown system/unverified release fails closed with a typed error; prompt text after `--` is never a flag.
3. Rows per environment × non-interactive form × placement (`=`/separate, codex `exec`), fake inputs only; narrowing mutants in
   a disposable copy; argvguard/spelling guard green; `go test ./...`, `go vet ./...`.
4. README "Permission-mode" + one new CHANGELOG bullet (released entries untouched, heading `## Unreleased`); release v0.5.22 cut
   by the orchestrator.
Attach `TASK-260924-1nh93t_results.md`, check the DoD items citing it, `task-board handoff TASK-260924-1nh93t --role developer`. A
`run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
