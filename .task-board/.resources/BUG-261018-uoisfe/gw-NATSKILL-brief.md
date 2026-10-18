# BUG-261018-uoisfe: in-place profile skills never reach the native skills surface (developer)

## Problem

Environments §9.4 says that the resolved skills of a profile materialize into its managed homes and, for the current profile of each scope, into "the in-place adapter surfaces under the manager §5 discipline".

Shipped code writes the in-place skills of the current profile only to `<native home>/.agent-context/skills/<name>`:
- `internal/envprofile/switch.go`, the constant `profileSkillsSurfaceDir` at line 65, used at lines 584, 622 and 710;
- no agent harness discovers that directory.

The native discovery directories form the manager §5 table in `internal/adapters/adapters.go:22-35`. Only the legacy global scope writes them, through `internal/install/global.go:545` (`adapters.StageGlobal` from `<curator home>/global/skills`), and global operations carry no takeover flag (§9.4).

Consequences, reproduced by another operator on a scratch copy of real homes at the same code:
- direct native Claude Code and Codex sessions keep discovering foreign skill links, here those of a retired manager;
- `profile use --takeover`, `profile sync --takeover` and `env resolve --repair` never replace those links;
- the infra exit cannot complete without manual steps.

## Required

1. **Native surface.** For the current profile of each scope, these operations materialize the profile's skills at `<native home>/<SkillsDir>/<name>`, using `SkillsDir` from `internal/envregistry`:
   - `profile install --use`, `profile use`, `profile sync`, `profile update` and `env resolve --repair`;
   - for claude_code, codex_cli and pi the surface is `skills/`; for muse it is `data/muse/skills/`.

   Each entry:
   - is a manager-owned link to the immutable store entry, in the same link form the `.agent-context/skills` entries use today;
   - is recorded in the environment marker as a surface path;
   - follows drift and repair rules exactly like the root-context target.
2. **Takeover (§9.5, §8.3.1).** An unrecorded entry at a target path is handled by a closed rule:
   - **Symlink pointing outside the store.** Without `--takeover`, stop with `environment_foreign_manager_detected`. With `--takeover`, back up the link itself into the next backup generation (same link text, never dereferenced), then replace it.
   - **Any other unrecorded entry.** Without takeover this is `environment_surface_unmanaged_conflict`. With takeover, back it up and then replace it.
   - **Entries the profile does not name.** They are never modified, moved or deleted.
   - **Symlink component below the native home.** A symlink there, including a symlinked skills directory itself, refuses with `environment_write_would_follow_link`. No flag authorizes traversal.
3. **Adapter-ledger entries.** A path recorded in the core §11 adapter ledger belongs to the legacy global scope.
   - Leave it in place.
   - Report it as a non-current row with a clear detail.
   - Record the interplay as an open spec question in the gaps file.
   - Do not change adapter-ledger semantics.
4. **Convergence from the old staging.** `.agent-context/skills/<name>` paths recorded by earlier versions are removed as recorded surfaces by the next materialization, so a switched home converges without manual steps.
5. **Out of scope:**
   - opencode keeps today's behaviour; its skills target is the machine-global native surface (§7.1 note). Report it in the gaps file.
   - Managed homes are unchanged (`managed.go`, `<home>/skills/`).
6. **Tests.** Go tests in temporary homes only; never touch the real HOME. Each test carries a one-line comment naming the claim it proves.
   - Positive cases:
     - use, sync and install materialize native links for claude_code and codex_cli;
     - repair restores a drifted link;
     - a profile switch replaces the set and removes recorded entries that the new profile does not name;
     - old `.agent-context/skills` records converge.
   - Negative cases:
     - without takeover a foreign symlink stops the operation, and nothing is written or backed up;
     - with takeover the backup generation holds the link with identical link text and the entry is replaced;
     - an unrecorded plain directory gives `environment_surface_unmanaged_conflict`;
     - skills the profile does not name stay byte-identical before and after;
     - a symlinked skills directory refuses with `environment_write_would_follow_link`;
     - a ledgered global entry is untouched and reported;
     - an unreadable marker means no writes.
   - Control: a mutant that skips the native write fails at least one test for that reason.
7. **Validation.** Run the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.

## Scope

- In scope: `internal/envprofile` (switch, status and repair paths), the registry lookups it needs, tests, and a short docs note.
- Not in scope: spec text. A separate curator-spec clarification will pin the in-place skill surface; write the questions it must answer into the gaps file.

## Hand-off

1. Attach `BUG-261018-uoisfe_gaps.md`. It lists each required item with file and line, the tests that prove it, and the open spec questions (ledger interplay, opencode).
2. Run `task-board handoff BUG-261018-uoisfe --role developer`.
3. END YOUR TURN.
