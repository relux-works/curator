# TASK-260901-2tdoy5 results: protocol/environments.md draft 1

## Deliverable
- `protocol/environments.md` (947 lines, 13 sections, per-section diagnostics tables), authored from accepted Decision 0010 at rev `2a861e5`.
- Location: worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`, branch `draft/environments-protocol`.
- Signed commit `eddd509e88194f914ca0473a7453d4568e649c7f`, base exactly `2a861e5d3ab23f63e16cd9cb3b8d9bd87517ed3c` (= origin/main at authoring; ancestor check ran green before branching).
- Signature: `Good "git" signature for oparin@me.com` (ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM), verified against the repo `maintainers.allowed_signers`.
- Not pushed, no PR, no tag; CHANGELOG and every other file untouched, per brief.

## Scope coverage (brief items 1-11)
1. Profile object: names (core sec. 2), source kinds git + local, `path` reserved with STORY-260901-2hkq49 reference, pin/branch/strict-tag semantics — doc sec. 1.
2. Profilefile.json schema-1 strictness, profile directory shape, PROFILE.md informative — sec. 2.
3. context.json schema-1: module entries, environments selector, class root|system, UTF-8/LF-only/one-trailing-LF rejection with stable codes — sec. 3.
4. Deterministic materialization with exact byte rules: generation header (closed 7-line grammar, no timestamps), chapter separator bytes, monolithic join, zero-applicable-modules output (header-only, resolved normatively), referenced-form layout (`.agent-context/modules/<profile>/<module-path>`, resolved normatively), core-sec.-8 hash binding — sec. 5.
5. Composition: machine-declared overlays, later-overrides-earlier default, skill-closure precedence + divergence warning, marker/fragment recording — sec. 6.
6. Adapter registry: closed rev-1 set (claude_code, codex_cli, opencode, pi) with home mechanism/shape, surfaces, forms + defaults, system-prompt channels (incl. pi APPEND_SYSTEM.md append + SYSTEM.md replacement), shadowing paths, per-platform credential passthrough; xcode-coding-assistant secondary targets with probe/home/surfaces and auto|off|explicit participation — sec. 7.
7. Modes (managed-home/linked/copied) + defaults, `.agent-environment.json` schema-1 semantics, ledger-discipline extension, drift diagnostics with both-halves wording — sec. 8.
8. Current profile, scoped switching, install activation rules, onboarding steps (inventory/notice/backup + takeover flag), always-strict profile audit + `context-secret-material` detector class hooked into manager sec. 7 — sec. 9.
9. `launch-env-fragment-v1` closed object, fields, system-prompt section, profile-influence boundary — sec. 10.
10. Umbrella `curator-<name>` discovery; `env resolve` verify-and-repair; read-only status extensions — secs. 10-12.
11. Per-section diagnostics tables, house style, English throughout.

## Verification
- Every cited core/manager section read and matched in the worktree at 2a861e5: core secs. 1, 1.1, 2, 5, 6.1-6.3, 7, 8, 9.4, 10, 11, 12.3; manager secs. 1, 2, 2.5, 5, 6, 7, 10.
- `make validate` (venv from requirements-dev.txt): exit 0 — 53 schemas + 691 vector files validated, 134 tool unit tests OK, `go test ./tools/...` OK. Docs-only change; run as tree-health evidence. The initial bare `make validate` exited non-zero on a missing `jsonschema` module — environment setup, not the change; resolved via venv.
- Not run: nothing else applicable; the task changes one markdown file in curator-spec.

## Companion artifacts
- Board resource `environments-doc-draft-notes.md` (same task): resolved carry-forwards with exact rules, deliberate deviations with rationale (notably opencode skills surface kept per Decision 0010 open question 3 recommendation), and reviewer open items.
- Logbook entry committed on the story branch (`389b6ff4`).
