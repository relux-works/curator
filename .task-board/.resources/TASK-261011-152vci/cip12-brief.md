# THE ONLY CURRENT INSTRUCTION — TASK-261011-152vci: CIP-0012 "Symlinked seed files" (Rejected, provisional) (developer)

Repository: curator-spec. Docs only. One new CIP file plus one README index row.

## What happened (facts you may rely on; verify the spec citations yourself)

- `protocol/environments.md` §8.4.1 (seed row) and the seed rule near the isolation matrix (search "A seed path that is not a regular file"): a seed path that is a symlink, directory or special file is unreadable-side and provisioning stops with `environment_seed_unreadable`; Curator never copies through a link and never skips the seed as absent. Only `codex_cli` `config.toml` is a copied seed among the files below (`claude_code` `.claude.json` is written, not copied; a Claude `settings.json` seed is not declared in revision 1).
- On a shared host, a separate installer of the agent tooling (the agents-infra tool, now being retired) installs `~/.codex/config.toml` as a symlink into a central configuration directory (`~/.agents/.configs/codex-config.toml`); it also links `~/.claude/settings.json` and the Codex rules files.
- Effect: a managed Codex home that needs first provisioning or a stale-home repair refuses with `environment_seed_unreadable`; a later Curator release candidate reports the home as `environment_home_stale` (passthrough: environment_seed) even on a plain resolve, which forces a repair that then refuses. A board tool released a harm-reduction change (its primary path no longer asks for `--repair`), which helps only while the home is current.

## Options considered (record all; keep the analysis honest)

- (a) Host migration: convert the installer's symlinks once into regular files with the same bytes; no Curator change. Cost: the central file no longer propagates automatically; a re-run of the installer recreates the link unless the installer changes or is retired.
- (b) Amend §8.4.1 to admit a symlinked seed under constraints: the link and its final target are owned by the home's user; the target resolves inside that user's HOME; the target is a regular file not writable by group or others; it is opened without following further links (one hop, then O_NOFOLLOW-class open of the resolved path, or a descriptor-relative walk); the target identity (path, device/inode, digest) is recorded in the marker so a later change is reported as drift, never silently re-read. Benefits: no host migration; helps users of dotfile managers (stow, chezmoi in symlink mode). Limits: Nix home-manager links into a root-owned store, so a same-owner rule would not help there; a wider rule weakens the "read failure is not absence / never copy through a link" boundary and adds a TOCTOU surface that needs its own tests.
- (c) Harm reduction outside Curator (the board tool avoids `--repair` on its primary path). Necessary but not sufficient.
- (d) Copy through the link silently. Rejected: violates the current seed rule and the security boundary.
- (e) Treat the link as an absent seed. Rejected: "read failure is not absence".

## Decision (operator, 2026-10-11) — write it in neutral words, no names

The installer that creates the links is being retired within days. Decision: no Curator specification change; the strict regular-file rule of §8.4.1 stays. At the installer's retirement its symlinks are converted once into regular files with the same bytes on every machine where it was installed; until then the installer's setup is not re-run. The proposal (b) is recorded as **Rejected (provisional)**: it may be revisited by a new revision of this CIP if real users of Curator need symlinked seeds (for example dotfile managers), or if a successor installer must keep links.

## Do this

1. Create `cips/CIP-0012-symlinked-seed-files.md` following the CIP-0001 template headings (Summary; Motivation and user stories; Current state; Design → Options considered, Recommendation; Security considerations; Compatibility and migration; Specification changes; Implementation plan; Test plan; Open questions for the operator). Header: Status `Rejected (provisional; operator, 2026-10-11)`, Owner `ivan-curator (orchestrator); decision: operator`, Created 2026-10-11, Affects `protocol/environments.md §8.4.1 (no change)`. Start with a short operative summary table like CIP-0002…0011 have (purpose, boundary, authority owner, inputs and outputs, normative links).
2. "Specification changes": state explicitly that none are made. "Implementation plan" and "Test plan": what a future revision of (b) would need (constraints above, marker fields, negative cases: link to another user's file, link outside HOME, chain of links, link to a directory or FIFO, group-writable target, target swapped between check and open, target changed after provisioning → drift), clearly labelled as conditions for a revisit, not planned work.
3. Add the CIP-0012 row with status `Rejected` to the index table in `cips/README.md`.
4. Cite spec sections by their numbers; do not invent new refusal codes or marker fields beyond naming them as "would need".

## Rules

- Public repository: no personal first names, host names, local absolute paths, session links or board-resource links in repo files. `~/.codex/config.toml`-style home-relative paths are fine.
- Do not change `protocol/environments.md` or any other CIP.
- Docs only. Do not run tests or builds yourself; the configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.

## Hand-off

Attach `TASK-261011-152vci_cip0012.md` (what you wrote, where, and the options table), then run `task-board handoff TASK-261011-152vci --role developer` and END YOUR TURN.
