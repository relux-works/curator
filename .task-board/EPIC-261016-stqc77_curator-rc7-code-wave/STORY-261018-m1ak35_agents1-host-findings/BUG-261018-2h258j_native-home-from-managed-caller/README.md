# BUG-261018-2h258j: native-home-from-managed-caller

## Description
NativeHome (internal/envprofile/switch.go:99-123) returns the adapter mechanism variable (CODEX_HOME, CLAUDE_CONFIG_DIR, XDG for opencode) whenever it is set. A caller that itself runs in a Curator-managed home therefore passes its managed home as the native seed origin: provisioning links auth.json to the managed home link, a two-hop chain, instead of directly to the native file. Contract: a Curator-managed environment home is never a native home. NativeHome ignores a mechanism value that resolves (after symlink evaluation) inside the Curator environments root and falls back to the HOME default, with a notice. Resolve and repair of an already provisioned profile verify against the origin recorded at provisioning, never re-derived from the caller environment. Reported by the Alexis-side migration keeper with a CM20b witness.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
