# Rework 7 — TASK-260910-19w2aj: one P1 from verdict rev7 (TASK-260910-19w2aj_review-verdict-rev7.md; repro script/log attached)

P1 — the draft closure drops the machine acquisition allowlist: DraftResolveConfig (cmd/curator/project_resolve.go:113-125) has no AllowedSources and closure.Options (resolve.go:163-171) omits it, so ensureRepo/gateSource sees an empty allowlist = allow-all; alias roots are gated, but legacy named roots and transitive dependencies clone denied providers (reviewer fixture: denied.test/provider cloned and locked). The legacy install path already passes cfg.AllowedSources (install.go:396).

Required: carry cfg.AllowedSources from the production CLI through DraftResolveConfig into closure.Options for legacy and transitive acquisitions with the existing semantics; CLI negative tests with a restrictive temp config and a clone-attempt recorder — denied root and denied transitive dependency must fail BEFORE acquisition and leave lock/bindings/installed state unchanged; allowed positive controls; a narrowing mutant that propagates the policy only for alias roots must fail. Do not hand the omitted configuration to the helper under test.

Preserve all accepted fixes (rev2–rev7). Narrow tests (fast Git fixtures, -run filters), tool calls under 2 minutes, evidence, checklist, handoff rev8 (story_final).
