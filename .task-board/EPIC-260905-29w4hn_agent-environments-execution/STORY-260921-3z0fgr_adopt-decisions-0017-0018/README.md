# STORY-260921-3z0fgr: adopt-decisions-0017-0018

## Description
Operator decision 2026-09-21: adopt curator-spec decisions 0017 (environment credential modes) and 0018 (curator run permission interface, already amended per the 2026-09-16 decision: config-driven permission mode in the launcher global config or profile config, flag overrides, default yolo). Work in the curator-spec repository (control root curator-spec). Adoption = the decision documents move from proposed to adopted with every open question resolved to a recorded choice, and the normative text they amend (environments, launcher/curator run interface, cli rows, conformance vectors if any) is updated in the same revision or explicitly deferred to named follow-ups.

## Scope
curator-spec decisions/0017-*.md, decisions/0018-*.md, the normative sections and CLI rows they amend, decision index/README

## Acceptance Criteria
both decisions read Status: adopted with the selected options and resolved open questions; amended normative text consistent (make validate green, cross-references updated); follow-up implementation leaves named per repository (curator, curator-agent-launcher)
