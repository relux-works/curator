# TASK-260906-1hn93j: cli-rows-for-takeover-and-import

## Description
Close the takeover shape gap in protocol/environments.md 9.5 and profiles/manager.md, then publish the operator surface in cli/curator.md: [--takeover] on the five mutating operations 9.5 enumerates as onboarding triggers, no standalone takeover command, and the onboarding-import row with --as, the lossy consent flag, and the --use activation control 9.6 imports from 9.1. Review cycle 1 established that the standalone command shape is excluded by 9.5 rather than merely unsourced. Follow the attached rework brief.

## Scope
environments.md 9.5 requires an explicit takeover operation ("an explicit takeover", "the explicit takeover flag") and 9.6 requires the operator to request the onboarding import under a per-operation consent flag for a lossy import. cli/curator.md at f39f4a9 publishes neither: the profile and env rows cover install, list, use, update, remove, sync, resolve, status, unmanage, backups scrub, compose and config, and nothing names takeover, import, or the lossy-import consent flag. Stage (c) implements 9.6 import and cannot invent an operator surface. Verified by reading every curator profile and curator env row in cli/curator.md.

## Acceptance Criteria
cli/curator.md carries a row for the explicit takeover operation and a row for the onboarding import, each spelling its flags exactly as environments.md 9.5 and 9.6 describe them, including the lossy-import consent flag and the optional profile name of the reassembled package. The examples block gains one line per new row. No normative rule is added to cli/curator.md that environments.md does not already state; any genuine gap in environments.md is recorded rather than silently invented.
