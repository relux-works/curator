# STORY-260916-ioemse: source-signer-allowlist-and-update-delta

## Description
Finding E1 (High): semver-range resolution (Decision 0012) trusts every future v-tag of a source; the lock is a record, not a signature, and no signer or provenance rule exists. profile update re-resolves ranges and re-materializes in-place surfaces (~/.claude/CLAUDE.md, ~/.codex/AGENTS.md) and managed homes, so whoever can push an in-range tag ships new system-prompt, root-context and MCP command bytes with only strict audit in the way. Strict-tag policy covers a moved tag, not a new one.

## Scope
curator-spec decisions/0012 + environments §4/§8/§12; curator contextresolve/contextlock/profile update

## Acceptance Criteria
A lockable per-source signer allowlist (SSH/GPG tag or commit signatures) is specified and enforced before a candidate enters the lock, reported as posture; profile update presents the resolved-version delta and refuses without explicit per-run confirmation when the delta introduces or changes a class: system module or an MCP declaration; the latest residual is named in the spec
