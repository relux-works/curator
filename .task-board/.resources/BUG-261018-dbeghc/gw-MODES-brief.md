# BUG-261018-dbeghc: managed-home seeds are world-readable (developer)

## Problem

Managed homes live under an environments root that is created with mode 0755, and the manager's managed writes use mode 0644. In `internal/envprofile/managed.go`, the `atomicManagedFile(..., 0o644)` calls include seed provisioning, near line 1280.

As a result, a native seed with mode 0600 becomes readable by group and others in the managed home. One example is the Codex `config.toml`, which carries `[projects]` trust levels and `[hooks.state]`. Under per-agent OS users this exposes one agent's managed home to the other agents.

## Required

1. **Seed permissions.** A seed keeps the native file's permission bits, and no managed seed is ever wider than its source. A 0600 native seed stays 0600.
2. **Directory modes.**
   - The environments root and every managed home directory are created with mode 0700.
   - Repair (`env resolve --repair` and the existing provisioning paths) tightens an existing wider directory without losing its contents.
3. **Mode table.** Every manager-written file in a managed home gets its mode from one documented table, in code and in a short docs note. Trust-bearing or credential-adjacent files are 0600.
4. **Seed refusals do not change.** Symlinked, unreadable or special-file seeds still refuse with `environment_seed_unreadable`, as today.
5. **Negative tests:**
   - a 0600 native seed never yields a copy that group or others can read;
   - a pre-existing 0755 home is tightened by repair;
   - a mutant that writes 0644 again fails for that reason.
6. **Validation.** Run the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.

## Scope

- In scope: code and tests, plus a docs note.
- Out of scope: the spec. Stricter modes are compatible with it, and a separate spec clarification follows.

## Hand-off

Run `task-board handoff BUG-261018-dbeghc --role developer`, then end your turn.
