# Republish unchanged — BUG-260921-3cgij4 revision 1 → 2 (bound developer run, curator-spec)

The revision-1 gate (`make validate`, local on this host) failed only by the host: `go test
./tools/...` → `signal: killed` at 87 s (a fresh test binary killed during a macOS exec-stall
window; the same package passes in 0.8 s now), and the unit suite took 1218 s under load. The
change is not at fault. Do NOT change any file. Verify `git status --short` in the Story
worktree (.temp/STORY-260921-atwkfi/worktree) lists the revision-1 paths only, add one line to
results.md ("revision 2 = revision 1 unchanged; gate rerun after a host exec-stall window"),
then `task-board handoff BUG-260921-3cgij4 --role developer`. If the rerun fails on the SAME
environmental symptom again, attach the failure and stop (the orchestrator will retry when the
host is healthy).
