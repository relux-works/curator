# Binding rework from independent review RUN-260818-f6b128

Reviewed commit 704060526560a36e540bb27678e58edb381482da. Preserve the accepted architecture and historical release bytes.

Close every blocker:
1. Add semantic validation and generated invalid vectors for capability receipt observed_at before expires_at, execution receipt started_at at or before completed_at, and phase-specific checkpoint predecessor rules including null predecessor for the first permit-issued checkpoint.
2. Add executable relational vectors and validator checks for provider identity, provider contract and binary, capability set and receipt, nonce, operation, permit, build input, artifact, freshness and expiry, checkpoint chain, and no portable fallback. Every normative mismatch needs a stable named rejection and required pre-execution failure.
3. Add mutation tests proving validate and release gates reject each violation and generation is deterministic.
4. Run make validate, make regenerate-check, and make release-check VERSION=1.0.0-rc.7 against the new committed clean candidate. Preserve rc.5 and rc.6 hashes.
5. Remove only generated tools/__pycache__ from the spec worktree; keep local task-board.config.json untracked and out of commits.

Review evidence is stored as TASK-260819-3kwd8g_review-verdict.md on the task.
