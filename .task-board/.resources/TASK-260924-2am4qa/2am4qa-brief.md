# TASK-260924-2am4qa — core §4.4 amendment: `directory` on manifest skill dependencies (THE ONLY CURRENT INSTRUCTION)

Control root: curator-spec; your Story worktree only. Read `campaign-producer-rules.md` and `skillfile-operator-memo-20260924.md`
(attached). Gate: board validation at handoff; locally the three `validate:` recipe lines in bounded parts + regenerate-check.
Today core §4.4 (`protocol/core.md` ~1113) lets a `dependencies.skills` entry carry only `git`, exact `ref`, `mode`, `commands`.
Add OPTIONAL `directory`: a portable relative subfolder of the dependency repository at the pinned ref selecting the skill there,
with EXACTLY the grammar/containment rules of the Skillfile schema 2 individual selector `directory` (`protocol/skillfile-sources.md`
§1) — reference them, don't restate differently. Absent `directory` = repository root (unchanged meaning). Define how the selected
package identity, closure unification (two dependencies selecting different folders of one repository at one ref; same folder twice),
lock and audit record the directory. Schema(s) for the skill manifest, conformance vectors (valid; invalid absolute/escape/empty/
backslash/.. ; missing folder; folder without SKILL.md; diamond), cross-references from skillfile-sources and profiles/manager.md,
CHANGELOG (unreleased). Motivation for the text: playbook repositories hold several skills under `skills/<role>` and depend on role
skills in subfolders of other repositories. Attach `TASK-260924-2am4qa_results.md` (before/after text, vector list, validation table), check the DoD
item, `task-board handoff TASK-260924-2am4qa --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
