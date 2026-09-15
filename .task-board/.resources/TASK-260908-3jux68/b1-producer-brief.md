# B1 producer brief: relux-root-context packages (TASK-260908-3jux68)

Role: developer (producer). Work inline. No nested spawns.

## Objective
Create a NEW PRIVATE GitHub repository `relux-root-context` (private until the
operator flips visibility) with Curator context packages built from the exact
source bytes of `/Users/iv/Developer/IV/relux-agents-infra @ dee5403eaf1ba7dbeddd8dc33d45e3b0ad0f5e59`
(`.instructions/*.md` + `INSTRUCTIONS.md`/`AGENTS.md` index), per the package
split in `agents-infra-to-curator-mapping.md` (epic precondition resource):

- `relux-root-context-core` (STRUCTURE, TOOLS, SKILLS, SKILL_TRIGGERS, DOCS, DIAGRAMS, PLATFORM)
- `relux-root-context-workflow` (WORKFLOW, TESTING)
- `relux-root-context-style` (STYLE, lowest weight)
- `relux-root-context-claude` with `environments: ["claude_code"]` (EXTERNAL_RESOURCES, REMOTE_AGENTS)
- `relux-root-context-attachments` (ATTACHMENTS + requires.skills on the attachments CLI skill)
- umbrella `relux-root-context-ivan` requiring the above with module weights per Decision 0012.

Each package ships a valid `agent-context.json` manifest; module bytes must
match the source exactly (no rewording, no invented content).

## Authoritative inputs (read first)
1. This brief.
2. Epic precondition `agents-infra-to-curator-mapping.md` (package split, weights).
3. `goal-launcher-and-infra-migration.md` (B1 workstream only; do NOT start B2-B7).
4. Curator Decision 0012 (weights/locks) and Decision 0013 (ownership), SPEC 0.3.0-draft
   §4 context-package contract, environments.md 1.1.

## Operator overrides in force
- Producers run Muse Spark xhigh; independent review is Astra medium (do not
  self-review).
- No hosted CI: put `[skip ci]` in every commit message. Run local validation only.
- Agent-created signed tags AND GitHub Releases are AUTHORIZED within this task's
  scope (2026-09-09 override, supersedes the old "no agent tags" line in the story
  text). Every tag must be signed with the configured human identity/key, must
  target the exact verified reviewed commit, and must follow the strict v-tag
  contract (e.g. `core/vX.Y.Z`, one tag per package + umbrella). Never retag or
  move a published tag.
- New repository stays PRIVATE until the operator flips visibility.

## Method
1. Inspect the source tree read-only (`git -C /Users/iv/Developer/IV/relux-agents-infra show dee5403:...`
   or a detached archive; never mutate the source checkout).
2. Work inside the managed Story worktree the runtime provides
   (control root `.temp/STORY-260908-l5nerr/worktree`). If absent, create your
   scratch under `.temp/TASK-260908-3jux68/` in the worktree — never write
   source files into the control root itself, never write LOGBOOK.md.
3. Scaffold the new repo layout (one directory per package + umbrella), copy
   module bytes exactly, write manifests, add a minimal README per package.
4. Validate locally: JSON schemas parse strict, `environments` selectors exact,
   weights match 0012, umbrella `requires` ranges resolve to the tagged versions
   you will publish. Record exact commands + exit codes.
5. Publish through the canonical flow: `gh repo create relux-root-context --private`,
   signed commits (`git commit -S`, verify with `git log --show-signature -3`),
   push branch, open PR, answer review, land the exact reviewed head only after
   an actual comment-review verdict for that head plus green LOCAL checks.
   Then create the strict signed v-tags on the landed commit.
6. Do NOT touch `~/.agents`, `~/.claude`, `~/.codex`, real `ax`, or the
   agents-infra launchers. No hand edits of managed homes.

## Handoff
Publish a Change Request through the runtime's normal completion flow and write
findings as a task outcome resource (what was created: repo URL, landed SHAs,
tags; validation log; known gaps). Leave the candidate state clean and
described — do NOT mark the task complete yourself; the orchestrator routes
review (Astra medium) and Story closure.
