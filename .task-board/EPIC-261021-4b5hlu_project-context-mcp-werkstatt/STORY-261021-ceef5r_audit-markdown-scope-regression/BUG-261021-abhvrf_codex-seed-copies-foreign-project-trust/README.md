# BUG-261021-abhvrf: codex-seed-copies-foreign-project-trust

## Description
The managed codex_cli home config.toml seed carries [projects] trust entries copied verbatim from the native Codex config, including entries for home paths that do not exist on this host (from another machine). Project trust is per-checkout state and should not be seeded into a managed home wholesale; at minimum entries outside the host must be dropped. Evidence kept privately on ivmbp; no paths recorded here (public board).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
