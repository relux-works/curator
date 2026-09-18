# TASK-260918-ryh3kw: manager-read-failure-discipline

## Description
curator: implement the landed environments §8.4.1 absence-versus-read-failure discipline (curator-spec 23be89e): inventory every read site that maps an error to absence with a verdict per site; one shared classification helper (absent vs present-but-unreadable/malformed) used at every site; a guard (analyzer, lint rule or test) that fails the build when an os.IsNotExist branch swallows another error class; the new diagnostics environment_passthrough_unreadable and environment_backup_record_unreadable; unreadable lock = environment_store_untrusted with no rebuild from it; vectors/environments-read-failure.json executed through the real code paths.

## Scope
(define task scope)

## Acceptance Criteria
1. Inventory resource lists every read site with file:line and a verdict; every site uses the shared helper. 2. Guard fails the build on an absence-shaped branch that swallows a non-ENOENT error (negative proof committed). 3. The two new diagnostics and the lock rule behave as the spec states, with status/resolve/repair/update tests. 4. environments-read-failure.json cases executed from CURATOR_CONFORMANCE_ROOT (a root at or after curator-spec 23be89e; skipped only when unset; ledger rows as needed). 5. Docs (cli/troubleshooting) and CHANGELOG entry; rule 8 hygiene.
