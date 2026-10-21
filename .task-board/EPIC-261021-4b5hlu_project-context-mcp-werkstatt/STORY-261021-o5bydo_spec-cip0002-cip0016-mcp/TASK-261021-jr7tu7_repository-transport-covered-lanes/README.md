# TASK-261021-jr7tu7: repository-transport-covered-lanes

## Description
Editorial clarification: enumerate the Git acquisition lanes covered by repository-transport, including profile closure members, without behaviour change.

## Scope
curator-spec: protocol/repository-transport.md covered-lanes list; editorial only.

## Acceptance Criteria
1. Every Git acquisition lane is listed with its section, or explicitly excluded with a reason.
2. No lane grammar is widened and no behaviour changes.
3. Contradictions found in lane sections are reported, not silently fixed.
4. Validator, prose and links hold.
