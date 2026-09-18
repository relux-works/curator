## Status
backlog

## Review
required

## Task Class
docs

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [ ] §9.5 closed per-manager × per-platform table (chezmoi, home-manager, yadm, stow, dotbot at least) with verified/docs-confidence labels and cited sources
- [ ] Precise per-platform path resolution (home, XDG, %LOCALAPPDATA%/%APPDATA%) and the implementation-reads-the-table rule; heuristic never blocks
- [ ] Vectors per platform (present → suspected, absent → none, none-on-platform → inert, XDG override) with a rule-7 validator gate; existing vectors byte-identical
- [ ] make regenerate + validate + regeneration proof exit 0; CHANGELOG entry
- [ ] TASK-260918-24eazm_spec-patch_rev1.patch = git diff HEAD (base recorded) with new files intent-to-added; EMPTY curator delta

## Notes

## Precondition Resources
- [TASK-260918-24eazm_brief.md](file://TASK-260918-24eazm/TASK-260918-24eazm_brief.md) — Producer brief
- [remediation-spec-producer-rules.md](file://TASK-260918-24eazm/remediation-spec-producer-rules.md) — Campaign rules for spec tasks (rule 7 pinning; patch = git diff HEAD)

## Outcome Resources
(none)

## Created
2026-09-18T04:44:47Z

## Last Update
2026-09-18T04:46:28Z
