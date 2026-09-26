# TASK-260915-s4qkmr: migrate-global-cli-ownership

## Description
Back up legacy standalone task-board commands, publish managed shims through Curator and verify ordinary shell resolution without restarting live sessions.

## Scope
Host state only: ~/.local/bin, ~/.curator/global/bin, ~/.curator/backups and the Curator-managed marker. No repository change, no daemon restart, no session interruption, no removal of any binary a live process resolves.

## Acceptance Criteria
Existing work is preserved; exact scoped state and validation evidence recorded; no new feature implementation.
