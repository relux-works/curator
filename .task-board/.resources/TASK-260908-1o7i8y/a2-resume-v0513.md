# Resume — TASK-260908-1o7i8y after the upstream API landed (2026-09-16)

skill-agents-management v0.5.13 (commit 63346f609614a9efef8e34f442ee9ca16e6ca697, signed tag) exposes the owned-environment snapshot of an admitted plan, computed by BuildPlan after preparation and alias projection (see the CHANGELOG entry and package docs in that tag; TASK-260915-23628n on this board holds the accepted evidence). The stop-the-line of your previous run (TASK-260908-1o7i8y_results.md option 1) is resolved.

Do:
1. Bump go.mod/go.sum to github.com/relux-works/skill-agents-management v0.5.13 (no replace, no go.work).
2. Adapt internal/composition so the composer consumes the snapshot instead of requiring the effective LaunchRequest (keep argv/env/stdin parity goldens; the SPEC §4.5 sentence about ChildEnv(nil, req) is satisfied by the upstream snapshot — record the erratum in SPEC 0.3.0-draft's changelog section: "the owned environment literals are supplied by the admitted plan's snapshot").
3. Continue the production main wiring per a2-main-wiring-brief.md and production-main-obligations.md, then hand off.
