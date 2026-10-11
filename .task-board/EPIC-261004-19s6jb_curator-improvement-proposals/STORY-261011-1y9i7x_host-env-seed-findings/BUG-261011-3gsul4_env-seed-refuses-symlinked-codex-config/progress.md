## Status
backlog

## Review
required

## Task Class
code

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Confirmed against v0.15.0-rc.4 source: internal/stateread/stateread.go ReadRegularFile uses Lstat and refuses any non-regular seed (environments spec §8.4.1, path must be a regular file), so the rc.4 binary refuses the symlinked ~/.codex/config.toml by design. A full run in an isolated HOME also needs an installed profile; the source read was used instead, because the owner real ~/.curator is off-limits. Tier 2 per tb-keeper: Gate A primary composition would refuse at cutover on machines with the agents-management symlink layout. Preferred fix (tb-keeper): admit a same-owner symlink that resolves inside HOME, opened without following further links, with the target recorded. This needs a spec amendment of environments §8.4.1, an owner-level security boundary change.
2026-10-11 tb-keeper: the main regression is fixed forward in task-board as option (c), tier 1 for critical-fixes (primary path resolves without --repair). This bug drops to tier 3: the Curator §8.4.1 amendment (b) versus host migration (a) is the owner decision, now non-urgent.
2026-10-11 closed without a Curator change (owner decision, Ivan): the strict regular-file seed rule of environments section 8.4.1 stays. The symlink is created by agents-infra, which is being retired; at its retirement its symlinks are converted once into regular files with the same bytes on every machine where it was installed (runbook: infra-exit, Gate A; execution on the mini: macmini-infra; laptops: the dispatcher side). Until then nobody re-runs the agents-infra setup. Recorded in a2a #skill-project-management orchestrator-channel, tb-keeper ack. The symlinked-seed amendment is recorded as a provisionally declined CIP so it can be revisited.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-10-11T03:05:01Z

## Last Update
2026-10-11T11:42:06Z
