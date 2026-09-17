# Rework 10 — TASK-260910-19w2aj: one P1 from verdict rev10 (TASK-260910-19w2aj_review-verdict-rev10.md; probe attached)

P1 — internal/gitops/gitops.go:126-142: HasRemote returns false on ANY `git remote` error and Fetch treats false as "no remote → no-op", so a failed remote enumeration makes `project refresh` succeed with a stale lock. Absence and failure must be distinct: HasRemote must return (bool, error); Fetch propagates the error; the origin-less skip happens only after a SUCCESSFUL empty enumeration. CLI negative test injecting a Git wrapper that fails only remote enumeration (exit 73) → refresh exits nonzero, lock/bindings/installed state unchanged; keep the genuine origin-less positive control; a mutant collapsing error into empty must fail the negative test.

Preserve the Windows fixture fix and every prior accepted fix. Narrow tests, tool calls under 2 minutes, evidence, checklist, handoff rev11 (story_final).
