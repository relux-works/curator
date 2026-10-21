# Brief: CLI grammar and UX review, curator + curator-run + werkstatt (proposal only)

**Repositories, read-only:** `curator` (your clone) and `curator-agent-launcher` (the curator-run launcher, in your Werkstatt). Do not change code. The deliverable is a proposal document.

## Why

The owner wants the command-line surface to read as one well-designed system: clear, consistent and pleasant to use, with no surplus grammar. It covers the Curator CLI, the curator-run front door, Werkstatt commands and the task-board session-manager registration that ties them together.

## Inputs

- `curator help` and every subcommand's `--help`; the CLI docs in the curator repository (`docs/cli.md` and the others it links).
- curator-agent-launcher `FRONT-DOOR.md`, `SPEC.md`, `README.md` and `cmd/curator-run`: §2 Surface, §3 host mode, §4 session-manager registry, §6 Werkstätten (`werkstatt up/down/recycle/status/list`), the refusal tables.
- task-board `session-manager register --for curator-run`, as FRONT-DOOR §4.3 describes it.
- Known inconsistency to start from: FRONT-DOOR §4.3 spells `curator-run session-manager register`, but rule 6 prints `curator run session-manager select`.

## Produce

`<TASK-ID>_cli-review.md`, with these parts:

1. **Inventory:** every user-facing command, with its subcommands, flags and exit codes.
2. **Findings**, each with evidence (file and section, or help text):
   - inconsistent naming (for example `curator run` against `curator-run`, verbs, nouns, singular or plural);
   - redundant or overlapping commands and flags;
   - surplus grammar;
   - unclear or inconsistent errors and remediation hints;
   - discoverability gaps;
   - hidden coupling between the tools (who registers what, and when);
   - places where an agent and a human need different defaults.
3. **Proposed grammar**, as a table "now -> proposed", with a one-line reason per row, grouped by noun: profile, env, project, run, werkstatt, session-manager, global. Prefer a small, regular verb set. Say where back-compat aliases are needed, and for how long.
4. **The setup.sh removal case:** how a session manager gets registered once `setup.sh` is gone. Curator runs no package code at install, so there is no install hook. Evaluate lazy registration by curator-run on first hosted use.
5. **Top five changes by value**, and which repository and spec each touches: curator, curator-spec, curator-agent-launcher, skill-project-management.

Each claim cites its source. No code changes, no PRs.
