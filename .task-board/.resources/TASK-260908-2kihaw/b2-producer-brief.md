# B2 producer brief: skill manifests in owning repos (TASK-260908-2kihaw)

Role: developer (producer). Work inline. No nested spawns.

## Objective
Ship Curator skill manifests (`agent-skill` manifests + git identity) for the
three skills, each in its OWNING repository (not in a monorepo):
- `pdf` (from `/Users/iv/Developer/IV/relux-agents-infra @ dee5403` `.skills/pdf`)
- `skill-creator` successor (from `.skills/skill-creator`; check whether a
  standalone successor repo already exists and the original plan permits it —
  reuse it if so, else manifest in place per mapping)
- `agents-attachments` CLI skill (the helper tooling shipping a CLI utility)

Follow the cocoaskills manifest pattern. The umbrella `relux-root-context-ivan`
(from B1, may land later) will reference these via `requires.skills` ranges —
so choose initial versions accordingly and record the exact version numbers in
the handoff. Keep skill bodies byte-faithful to the source; manifests are new.

## Authoritative inputs (read first)
1. This brief.
2. Epic precondition `agents-infra-to-curator-mapping.md` (skills row).
3. `goal-launcher-and-infra-migration.md` (B2 workstream only; do NOT start B1/B3-B7).
4. Curator Decision 0012/0013 as applicable; SPEC 0.3.0-draft skill contract.

## Operator overrides in force
- Producers run Muse Spark xhigh; independent review is Astra medium.
- No hosted CI: `[skip ci]` in every commit message. Local validation only.
- Agent-created signed tags AND GitHub Releases are AUTHORIZED within this
  task's scope (2026-09-09 override). Tags signed with the configured human
  identity/key, targeting the exact verified reviewed commit. Never retag.
- Repos stay PRIVATE until the operator flips visibility.

## Method
1. Inspect source skill bytes read-only
   (`git -C /Users/iv/Developer/IV/relux-agents-infra show dee5403:.skills/...`).
2. Work inside the managed Story worktree the runtime provides
   (control root `.temp/STORY-260908-sd6xkr/worktree`); never write source into
   the control root, never LOGBOOK.md.
3. Create manifests + versions per owning repo, validate manifests strict
   locally (parse + schema + range resolvability), record commands + exits.
4. Deliver each repo change through branch -> PR -> actual comment-review
   verdict for the exact head -> green LOCAL checks -> land exact reviewed head.
   Then signed v-tags on landed commits.
5. Do NOT touch `~/.agents`, `~/.claude`, `~/.codex`, real `ax`, or the
   agents-infra launchers.

## Handoff
Change Request via the runtime's normal completion flow + outcome resource
(repo URLs, landed SHAs, tags/versions, validation log, gaps). Do NOT mark the
task complete; orchestrator routes Astra-medium review and Story closure.
