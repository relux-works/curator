## 2026-09-05

### 2225 — ErrModelMissing adopted as a core sentinel; composition negatives split per plugin
- DECISION: `agentic.BuildPlan` (pkg/agentic/plan.go) refuses an empty or whitespace `Model.ID` in every mode before any plugin surface; exec grammar previously admitted `--model ""`. pi keeps its own interactive refusal as a second line of defence.
- FIX: Invariant 7 added to docs/architecture.md.
- FIX: claude/codex/regress composition negatives split into prefix-only and servers-only; a `Prefix && Servers` narrowing mutant fails in all three packages plus core and pi.
- SCOPE: skill-agents-management head 93abeae on feat/interactive-mode-follow-ups, signed, not pushed.
- NOTE: Decision 0013 erratum for ErrParameterNotInteractive deferred to the spec side (environments 1.1 batch).
- STATUS: handed off to review (TASK-260905-1wi9j6).
