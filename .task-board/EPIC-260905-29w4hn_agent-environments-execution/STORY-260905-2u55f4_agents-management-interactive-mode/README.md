# STORY-260905-2u55f4: agents-management-interactive-mode

## Description
Step 1b: add LaunchModeInteractive to skill-agents-management per Decision 0013 Decision 5 (argv = model selection + effort transport only; no print mode, output format, permission bypass, goal machinery, budget, service tier; empty stdin unless stdin effort transport; Composition refused). Worktree ~/Developer/ReluxWorks/.worktrees/agents-management-interactive, branch feat/launch-mode-interactive, base 91bf945.

## Scope
(define story scope)

## Acceptance Criteria
Mode declared and implemented for claude, codex, pi with per-system argv tests and negative exec-marker tests; make build vet test regress green; docs updated; PR reviewed and landed on main by fast-forward of the exact reviewed head; no tag.
