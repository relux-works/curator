# Integration instruction — TASK-260919-2cmg0y (bound developer run)

Revision 1 is accepted and it is the story_final revision of STORY-260919-37szes (the seven checkpoints 3ukdk4, 2wyzde, 3v7x6j, 2eg8nv, 1sbj7o, 3ccq6b, 2d9gfv plus this leaf, replayed onto trunk d4fe8347). This board and code are colocated (curator control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). If trunk moved by board-only commits since the review, let the integrate reparent and revalidate (the remote gate runs on GitHub, up to ~90 minutes; the shell call may stay open that long). Rules: no board writes (no resources, notes, checklist, set_status) before or during the integrate. Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260919-37szes --cr TASK-260919-2cmg0y --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2cmg0y-1.log

ONLY AFTER it exits: attach `.temp/integrate-2cmg0y-1.log` as outcome resource `TASK-260919-2cmg0y_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
