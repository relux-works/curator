# THE ONLY CURRENT INSTRUCTION — BUG-261018-uoisfe: round 2, native skills surface for claude_code, codex_cli and pi (developer)

Your Stop-The-Line packet was right, and the orchestrator takes option A.
- Muse is out of this leaf. It does not take part in in-place switching (`switch.go:82`), and its independent XDG roots do not fit the one-home path model. A native Muse multi-root ownership contract is a separate follow-up. Keep it in the gaps file as open spec question 1.
- Opencode is also out, as before.
- Everything else in `gw-NATSKILL-brief.md` holds for claude_code, codex_cli and pi. Where that brief and this one differ, this one wins.

## Scope

- The current in-place adapters claude_code, codex_cli and pi materialize the current profile's skills at `<native home>/skills/<name>` as recorded, manager-owned links, instead of `.agent-context/skills/<name>`.
- Operations: `profile install --use`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`.
- Out of scope:
  - Muse, opencode and managed homes keep today's behaviour;
  - spec text;
  - the root `README.md`. Put the docs note under `docs/` or in a package README; trunk changes the root README often, and every overlap forces another converge round.

## Attack surfaces (the review sweeps exactly these rows)

| # | Surface | Must hold |
| --- | --- | --- |
| 1 | Target entry: absent | Link created and recorded |
| 2 | Target entry: recorded manager link | Replaced or repaired as an entry; drift is detected and repaired |
| 3 | Target entry: unrecorded symlink outside the store | Without `--takeover`: `environment_foreign_manager_detected`, nothing written, no backup. With `--takeover`: backup of the link itself (identical link text), then replace |
| 4 | Target entry: unrecorded directory, file, or link into the store | Without takeover: `environment_surface_unmanaged_conflict`. With takeover: backup, then replace |
| 5 | Path components | A symlinked `skills/` directory, or any component below the native home, refuses with `environment_write_would_follow_link`; no flag authorizes traversal |
| 6 | Entries the profile does not name | Byte-identical before and after every operation, including takeover |
| 7 | Adapter-ledger entry at a target path | Left in place and reported non-current, with a clear detail |
| 8 | Old staging | Recorded `.agent-context/skills/<name>` paths are removed as recorded surfaces; the home converges |
| 9 | Profile switch | Recorded skill links not named by the new profile are removed; the new set is linked |
| 10 | Backup generation | An existing next generation gives `environment_backup_exists`, and nothing is half-written |
| 11 | Marker unreadable | `environment_marker_unreadable`; no writes |
| 12 | Per-operation flag acceptance | `--takeover` works on the five operations above and on nothing else |

Every row needs a named test driven through the production entry point, positive and negative where the row has both, and at least one narrowing mutant per gate row (3, 4, 5, 6, 7, 11). Report coverage as `n of 12 rows driven` with the production call site for each row.

## Validation and hand-off

1. Run the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
2. Update `BUG-261018-uoisfe_gaps.md` with a "Round 2" section: the coverage ratio, the mutant table and the open spec questions (Muse multi-root, adapter ledger, opencode).
3. Run `task-board handoff BUG-261018-uoisfe --role developer`.
4. END YOUR TURN.
