# Brief — TASK-260916-1irwfr (story_final publication, no code)

Context: STORY-260908-2a4936 (umbrella relux-root-context-ivan + MCP declarations). Both code leaves are already LANDED and checkpointed:
- relux-mcp main 027f55b7 (TASK-260908-1bpra2), tags figma/v1.0.0, safari/v1.0.0, v1.0.0
- relux-root-context main abaadf43772341d0196e72a4ca9914017dc8f512 (TASK-260916-bn5kvb, PR #2, tag v1.0.0), tree 9eaad1ee3a823004b020f998ec2cfdf2e5caefad

The Story branch task-board/story/STORY-260908-2a4936 (tip 910cd9ac) has exactly that tree. Your job is ONLY to publish the Story-final Change Request so the Story can close:

1. Work in the managed Story worktree the spawn gives you (control root /Users/administrator/Developer/ReluxWorks/relux-root-context). Do not edit any file. Do not commit manually on the Story branch.
2. Verify: `git status --porcelain` empty; `git diff origin/main --stat` empty (or only the identical tree); `bash scripts/validate.sh` exits 0.
3. Attach a short outcome resource `TASK-260916-1irwfr_evidence.md` with those command outputs (exit codes).
4. Tick the checklist items (`check_item`) and hand off: `task-board handoff TASK-260916-1irwfr --role developer`. The handoff publishes the Change Request (kind derived story_final). If the handoff refuses with `change_request_base_authority_mismatch` or a refresh error, attach the exact refusal as the evidence resource and stop; do not work around it.

Never run worktree integrate/complete/checkpoint or set_status yourself.
