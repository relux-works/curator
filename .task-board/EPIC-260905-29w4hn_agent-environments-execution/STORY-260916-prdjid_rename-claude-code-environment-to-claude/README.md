# rename environment ids: claude_code → claude, codex_cli → codex

## Description
Operator decisions 2026-09-16: environment identifiers claude_code → claude and codex_cli → codex (pi and opencode unchanged). Spec-first: environments spec registers claude and codex as canonical ids with claude_code and codex_cli as deprecated aliases for one release (manifests, machine config, CLI, launcher, defaults.json, markers of existing managed homes migrate transparently); the frozen v1 launch-env-fragment schema enumerates env ids, so the spec task must decide the compatible path (additive enum values plus alias normalization rule, or a versioned schema) without breaking v1 consumers. Then Curator (envregistry, markers, CLI), launcher (curator run claude|codex), relux-root-context packages (validate.sh registered envs, per-env targets/forms) and docs.

## Scope
(define story scope)

## Acceptance Criteria
Spec amendment accepted and landed; curator resolves both ids with claude canonical and claude_code deprecated (warning), existing managed homes keep working without re-provisioning; curator run claude works; packages validate with the new id; conformance vectors updated; all landed through signed PRs with independent review.
