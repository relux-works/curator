# TASK-260916-2fu85y: amend-0018-config-driven-permission-default-yolo

## Description
Operator decision 2026-09-16: amend decision draft 0018 (curator run permission interface): the launch permission mode is configured in the launcher global config (curator-run-defaults, new member) or in the profile config, the CLI flag --permissions native|yolo only overrides it, and the DEFAULT is yolo. Replace the draft rule that defaults/config can never set yolo with: config can set it, precedence flag > profile config > global config > built-in default yolo; keep the per-environment native mapping, refusals for tools without an equivalent (pi, opencode), tracked-mode interplay and the provenance line (now reporting the source: flag|profile|global|default). Record the security implication (bypass by default on operator-owned machines) as an explicit operator choice with a lockable knob to force native. Update UNRESOLVED_QUESTIONS and CHANGELOG per convention.

## Scope
(define task scope)

## Acceptance Criteria
decisions/0018 updated (still proposed); make validate green; PR landed after independent review; issue #55 updated with a comment pointing at the amendment.
