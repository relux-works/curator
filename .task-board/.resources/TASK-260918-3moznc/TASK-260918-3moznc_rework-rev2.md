# Rework brief — TASK-260918-3moznc, revision 2

Revision 1 was rejected with three corrections
(`TASK-260918-3moznc_review-verdict-rev1.md`); the deliverable matrix and the
read-site inventory passed otherwise. Keep everything else byte-identical.

- **F1 — unreadable lock: no rebuild from unreadable evidence.** §4 (lines
  ~742–747) now lists a lock that "cannot be read or parsed" among the
  entry-class failures that "a real operation rebuilds from the revalidated
  snapshot", while the new §8.4.1 rule says the operation needing the file
  fails closed and forbids absence-shaped re-materialization; §10.1 (~2650)
  inherits the rebuild branch and the lock row (~2017) states no exception.
  Settle it in the fail-closed direction (the settled decision): an
  unreadable or malformed lock is `environment_store_untrusted` for that
  profile, resolve fails closed with no fragment, status non-current with
  currency unknown, and NO mutating operation (install, update, use, sync,
  repair, GC) rebuilds, re-materializes or replaces anything from that
  evidence — recovery is an explicit operator action (reinstall from the
  profile source, or the documented backup path) that never reads the
  unreadable file as input. State the precedence across §4, §8.4.1, §10.1
  and the manager mirrors once, referenced from each. Add a pinned vector:
  unreadable lock + a mutating operation (repair and update at least) →
  refused, nothing rebuilt, nothing written (rule 7: the validator pins the
  operation, the file class, the failure class, the outcome and "no
  mutation").
- **F2 — §9.7 admission row.** Add the `environment_passthrough_unreadable`
  row to the §9.7 diagnostics table (cross-reference the owning §7.7 row;
  no duplicated policy). No knob/lock changes.
- **F3 — backup-record row must name its diagnostic.** The closed table's
  backup row (~2019) and its owning paragraph (~1903–1912) carry behaviour
  only ("the backup inventory is unknown") and no diagnostic. Resolve it
  explicitly: reuse an existing code if one genuinely applies (state which
  and why), otherwise admit exactly ONE new code for the backup-record
  class (working name `environment_backup_record_unreadable`, warning or
  error — state which and why; it is authorised by this brief), spelled
  identically in the table, §8.3, §9.7 and any status/reporting mirror
  whose contract it changes, with a vector. No silent waiver.
- Also close the §9.4 migration note the reviewer flagged: cite §8.4.1
  explicitly and name the diagnostic the migration path reports for an
  unreadable install record (existing code or the table's row).

Validation as before: `make regenerate`, `make validate`, regeneration proof
(exit codes); evidence "Revision 2" section; `TASK-260918-3moznc_spec-patch_rev2.patch`
= `git diff HEAD` (base `5146c7b`) with new files via `git add -N`; EMPTY
curator delta; `task-board handoff TASK-260918-3moznc --role doc-writer`.
