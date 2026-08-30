# TASK-260827-2gmk4c Outcome Report: prose-style-en (Rework Rev2)

## Summary of Changes

Rework rev2 addresses both blocking findings and non-blocking nits from reviewer run RUN-260827-701ce4:

1. **Finding 1 (blocking - non-existent command)**: Replaced `curator repair` on line 19 of `docs/prose-style.md` with `curator install`:
   - Old: `A compiled command links a skill binary to an executable shim in .agents/bin/. curator repair restores broken links.`
   - New: `A compiled command links a skill binary to an executable shim in .agents/bin/. curator install rebuilds a missing or drifted shim.`
   - Verified that Curator has no separate `repair` command (`cmd/curator/main.go`, `cmd/curator/builds.go:680`, `docs/compiled-commands.md:57`).

2. **Finding 2 (blocking - non-existent document cross-reference)**: Replaced `ARCHITECTURE.md` cross-reference example on line 67 of `docs/prose-style.md` with a target existing in this repository:
   - Old: `Cross-reference with a precise target: "see Security model in ARCHITECTURE.md", never "as we will see later".`
   - New: `Cross-reference with a precise target: "see Reconciliation and repair in docs/compiled-commands.md", never "as we will see later".` 

3. **Nit (non-blocking - em-dash typography in Bad example)**: Updated Bad rhetorical-glue example on line 77 to display an em-dash (`—`), and ensured consistent rule naming.

## Verification Evidence

- **File line count**: `docs/prose-style.md` is 107 lines (well within the 200-line ceiling).
- **Cyrillic check**: 0 Cyrillic characters found.
- **Validation Gates**:
  - `make gate-selftest`: Exit code 0 (81 passed, 0 failed).
  - `make no-broad-suppression`: Exit code 0 (`no-broad-suppression: ok`).
- **Grep verification**:
  - Line 19: `curator install rebuilds a missing or drifted shim.`
  - Line 67: `see Reconciliation and repair in docs/compiled-commands.md`
