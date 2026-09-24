# TASK-260924-1kpw4w — implement `dependencies.skills[].directory` (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree only. Read `campaign-producer-rules.md` (artifacts only in $TMPDIR; results are board
resources, never repo files) and `skillfile-operator-memo-20260924.md`. Landing gate = hosted CI at handoff.
Source of truth: the ACCEPTED curator-spec amendment TASK-260924-2am4qa (core §4.4; candidate tree b3663bb3 on spec 3d2c611, not yet
on spec main): read it from the curator-spec control root's Story branch for STORY-260924-15f1yi or from the patch resource
`TASK-260924-2am4qa_change-request_rev2.patch` on the curator board. It adds optional `directory` on skill-manifest
`dependencies.skills` entries (manifest schema revisions agent-skill-v9 / csk-skill-v9 in the draft-sources namespace), same
grammar/containment as the Skillfile schema 2 individual selector `directory`, identity/closure/lock/audit recording the directory,
diamond and same-folder rules, and vectors (`manifest-dependency-directories.json` + agent/csk-skill-v9 schema cases).
Curator's conformance pin does not contain these yet: copy the needed vectors byte-for-byte from the accepted candidate into test
fixtures (cite path + sha256), exactly as TASK-260918-bi6ouz did; the pin move happens later in lockstep with the spec merge.
Implement exactly the task description/AC: install/update through the real CLI with lock and audit; invalid grammar/containment,
missing folder, no SKILL.md refused with the spec's diagnostics; diamond; absent `directory` byte-identical to today; narrowing
mutants in a disposable copy. TASK-260924-1aa9wb is removing the CURATOR_DRAFT_SOURCES_V1 switch in parallel — enable the lane in
tests the way existing draft tests do; do not touch the switch. Bounded runs; CHANGELOG. Attach results (board resource), check DoD,
`task-board handoff TASK-260924-1kpw4w --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
