# Producer brief: author protocol/environments.md (normative draft 1)

## Setup
- Base: fresh `origin/main` of `~/Developer/ReluxWorks/curator-spec` — MUST be commit `2a861e5d3ab23f63e16cd9cb3b8d9bd87517ed3c` or later; fetch and verify before branching.
- Create worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative` on new branch `draft/environments-protocol` from that base. Work only there. Shell tooling only; verify the tree exists before writing.
- Do not push, do not open PRs, do not tag. One or more signed commits (`git commit -S`, verify signature) on the branch.

## Task
Author `protocol/environments.md` — the normative document for Decision 0010 (`decisions/0010-agent-environment-profiles.md` on main, rev 2a861e5). Model: `protocol/assurance.md` (separately versioned surface that adds identities and widens nothing) and the prose discipline of `protocol/core.md` (MUST/MUST NOT, closed sets, exact byte rules, diagnostics tables).

Scope: revision-1 surfaces only, normative:
1. Profile object: names (§2 grammar), source kinds `git` (full §6.1–6.3 reuse) and `local`; `path` named as reserved/deferred with its story reference. Pin/branch semantics.
2. `Profilefile.json` schema-1 semantics and strictness; profile directory shape; `PROFILE.md` informative.
3. `context.json` schema-1: module entries, `environments` selector, `class: root|system`; validation rules (UTF-8, LF-only, exactly one trailing LF — reject diagnostics with stable codes).
4. Deterministic materialization: EXACT byte rules for the generation header (fields, ordering, no timestamps), chapter separators under composition (resolve the exact bytes — carry-forward), monolithic join, zero-applicable-modules output shape (resolve — carry-forward: recommend empty output = header only, state it normatively), referenced-form layout naming (resolve — carry-forward: managed subdirectory, per-source-profile grouping, exact naming rule), §8 content-hash binding.
5. Composition: machine-declared overlays, precedence (later-overrides-earlier default), skill-closure precedence rule, marker/fragment recording.
6. Environment adapter registry: closed rev-1 set (claude_code, codex_cli, opencode, pi) with per-adapter normative declarations (home mechanism/shape, surfaces, forms + defaults, system-prompt channels incl. pi APPEND_SYSTEM.md append + SYSTEM.md replacement, shadowing paths, credential passthrough per platform); secondary fixed-home targets (xcode-coding-assistant probe/home/surfaces, auto|off|explicit participation).
7. Materialization modes (managed-home/linked/copied), mode defaults, `.agent-environment.json` marker schema-1 semantics, ledger discipline extension, drift diagnostics (stable codes, both-halves wording).
8. Current profile, scoped switching, install activation rules, onboarding steps 1 + 3-notice + 4 (takeover flag), always-strict profile audit rule and the authorized secret-detection detector class hook into manager §7.
9. `launch-env-fragment-v1`: closed object, fields, system-prompt section, the profile-influence boundary.
10. Umbrella subcommand discovery convention (curator-<name>), env resolve behavior (verify-and-repair), read-only status extensions.
11. Diagnostics tables per section, house style. English. Cross-reference core/manager precisely (verify every § you cite).

Out of scope: launcher internals (its own spec), ax changes, MCP write, settings, path-kind import mechanics, schemas/vectors files themselves (a sibling story), CHANGELOG.

## Deliverables
1. `protocol/environments.md` in the worktree, signed commit(s).
2. Board resource `environments-doc-draft-notes.md` on the task: resolved carry-forwards (exact rules chosen), any deliberate deviation from Decision 0010 with rationale, open items for the reviewer.
3. Handoff: status to-review with summary.

## Do not
- Touch any other file (except CHANGELOG explicitly excluded — do not touch it).
- Mark done; orchestrator owns closure.
