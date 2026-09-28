# TASK-260906-3o75d6 review verdict rev2 — CHANGES_REQUESTED

Candidate tree 522ff20bd vs base eadb1c06 reviewed.

## Verified OK
- Rework item 1: note (cli/curator.md:54-63) now carries notice+backup and `environment_surface_unmanaged_conflict` refusal, verbatim with environments.md:2628-2632; `env resolve` "--takeover applies only with --repair" kept. No rule added; `env resolve --takeover` without `--repair` still unstated.
- Rows reduced to "see the takeover note below."; example moved into the `profile use` group; import/global rows carry no `--takeover`.
- 1xbrz6 pins still present (validate_cli_takeover_rows import/global check, TAKEOVER_CARRYING_OPERATIONS); pin strengthened (exact note equality, carrier counts, pointer-only rows).
- `uv run --with-requirements requirements-dev.txt python tools/validate.py` → exit 0 ("validated 64 schemas and 1166 vector files").

## Blocking
- Rework item 3 (binding, 3o75d6-rework-1.md) NOT done: cli/curator.md:57-58 still reads "named above as onboarding triggers", which points at nothing in the CLI guide. Reproduce: `grep -n "named above as onboarding" cli/curator.md` → hit. Required: "named in environments section 9.5 as onboarding triggers", and adapt `takeover_cli_clause`/the note equality in tools/validate.py (e.g. substitute that phrase in the expected clause) with a negative test that the unsubstituted "named above" form fails.

## Minor
- cli/curator.md example block: removing the old example left a blank line before the closing fence (~line 232); drop it.
