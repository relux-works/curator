# TASK-260906-1xbrz6: takeover-closed-set-excludes-import-and-global

## Description


## Scope
The environments 9.5 amendment makes the takeover flag acceptable on exactly five mutating operations: profile install, profile use, profile sync, profile update, env resolve --repair. Review cycle 2 on TASK-260906-1hn93j observed two operations outside that closed set that can also meet unmanaged files. First, profile import installs through the 9.1 path pipeline and activation follows the 9.1 rules without magic, so a first-install activation materializes in-place surfaces; 9.6 says the import writes nothing into any native home by itself, which settles the import proper but arguably not its activation. Second, 9.4 global add and global install materialize in place under the current profile. Neither carries the flag. An operator blocked by environment_surface_unmanaged_conflict on either can still escape by running profile sync --takeover or profile use --takeover first and retrying, so revision 1 is usable as it stands. Widening the set was explicitly forbidden by the rework brief because the enumeration had to stay closed and sourced.

## Acceptance Criteria
Either environments 9.5 states why those two operations are outside the closed set (the escape path being sufficient), or the set is widened with the sentence that justifies each added member, and cli/curator.md follows. profiles/manager.md agrees either way. Decided before the stage (c) implementation consumes the enumeration.
