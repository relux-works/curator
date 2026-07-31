# TASK-260729-365r5r: prototype-savejournal-namespace-validation

## Description
Build an isolated production-side prototype that removes repeated O(P^2) target-namespace validation from saveJournal while preserving fail-closed validation before any transaction mutation and all journal integrity checks.

## Scope
Use a task-owned worktree copied from TASK-260729-rfrdfo prototype state or the pristine jrrgw9 candidate as explicitly recorded. Limit product edits to internal/transaction and tests needed to prove validation timing/coverage. Do not touch the main candidate, timeout values, CI, protocol/spec behavior, or weaken/remove assertions. Run only focused transaction/atomicity/install measurements sequentially; no ./... until independent tester stage.

## Acceptance Criteria
Literal file/function allowlist and pre/post manifests; static call-path proof that every externally supplied or recovered journal target graph is validated before mutation; every saveJournal call still revalidates current filesystem namespace facts but performs at most O(P) filesystem identity/resolution reads per pass instead of repeating them for O(P^2) pairs; malformed, overlapping, and between-save symlink/alias changes fail closed; focused non-race and race evidence demonstrates a defensible atomicity margin at or below 480 seconds or explicitly rejects the prototype; independent review required before integration.
