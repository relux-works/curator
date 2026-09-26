# TASK-260924-11burj — orchestrator decision on the update workflow (THE ONLY CURRENT INSTRUCTION, together with 11burj-brief.md)

Decision: option 1 — use the documented schema-2 project API. The memo's "install/update" means the schema-2 lifecycle, not the
manager-level `curator update` fetch. Acceptance drives: `curator install <project>` (resolve + install), and for the update step
`curator project refresh <project>` followed by `curator install <project> --audit advisory` (or the default audit mode — state which,
and assert the audit records either way). Do NOT change `curator update`. Everything else in `11burj-brief.md` stands: collection entry
{from: repo A, directory: "skills", include: ["*"]}, transitive manifest dependency with `directory` into repo B, new skill after a new tag
appears via refresh+install with lock/audit updated, fresh-home replay from the committed lock (tamper → source_snapshot_changed,
unreachable → source_snapshot_unavailable, lock byte-identical), negative rows, automated test with local fixtures only.
Keep the CHANGELOG entry text you already wrote in results. `task-board m 'set_status(TASK-260924-11burj, status=development)'` first;
handoff when done.
