# TASK-260905-2czqqy: launcher-spec-0-2-1-follow-ups

## Description
Launcher SPEC 0.2.1-draft follow-ups: (1) §4.3 invokes `curator env resolve <env-id> --repair --format json` (environments.md 1.1 §10.1 makes resolve read-only without --repair and fail-closed on a stale home; the launcher is the caller that repairs); (2) cycle-2 residual minors: §6 defaults row wording, §9 docs-confidence item; (3) codex layer file must be stat-ed before launch (verification sprint item 9: a missing -p layer is silently ignored) and `-p` accepts exactly one value; (4) the ax.json/defaults.json file family once environments 1.1 §12 names the config knobs.

## Scope
(define task scope)

## Acceptance Criteria
SPEC.md 0.2.1-draft with the four items, stub/README version bumped, make check green, independent review ACCEPT, landed by fast-forward.
