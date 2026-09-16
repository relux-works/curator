# Review brief — CLI aliases claude / codex (TASK-260916-11lwua curator side, TASK-260916-rkrphg launcher side)

Spec rule (curator-spec main 2d6d497, profiles/manager.md CLI section): the environment operand accepts `claude` and `codex` as aliases of `claude_code` and `codex_cli`; aliases are normalized before validation; every output, marker, fragment, config value and lock keeps the canonical id; aliases are never persisted; wire ids are unchanged.

Verify with evidence (quote commands and exit codes; goldens/tests you ran):
1. Both spellings reach the same production path at EVERY CLI entry point that takes an environment operand (list them from `--help`); an alias given anywhere is normalized once, before validation and lookup; unknown ids are still refused with the existing diagnostic.
2. Outputs print the canonical id (status lines, provenance/origin lines, errors); nothing under the managed home, config, markers, fragments or locks ever contains the alias (test asserting on persisted bytes).
3. Help text / README (launcher: SPEC flag table) lists the aliases; no wire-id, schema or defaults.json change slipped in; no spec edits.
4. Tests cover both spellings and the refusal of unknown ids at the production entry points; narrow package tests green; hosted gate green (validation log attached).
Verdict ACCEPT or CHANGES_REQUESTED with numbered findings and file:line; accept_cr on ACCEPT.
