# TASK-260917-2tx81l — orchestrator decision 1 (THE ONLY CURRENT INSTRUCTION, with 2tx81l-brief.md)

Your recommendation is right: honour the published spec. Curator-spec main b1a2efb §8 defers skillfile-lock-v2, install-marker-v6 and
source-audit-v2 to TASK-260930-3ny11n. Those schemas are not part of this task. For the seven skillfile-sources-v1/schema-cases rows:
- If the referenced cases EXIST in the suite whose digest selects them, re-own those rows to TASK-260930-3ny11n. They stay known gaps,
  and 3ny11n owns them.
- If the cases do NOT exist at b1a2efb, because they were only in the pre-split candidate, remove the rows and remove or retarget any
  count table keyed to that obsolete candidate digest. No dead rows and no dead digest tables. Prove that the ledger ratchet and
  conformance-case-counts are exact for the rc.13 digest and for the b1a2efb suite digest (compute both).

Keep everything else you implemented. Finish the envprofile coverage that was interrupted: run the relevant envprofile cases, split with
-run so each run finishes, with the real exit codes.

Set status development, update the results (before/after table: 80 driven, 7 re-owned or removed), run
`task-board handoff TASK-260917-2tx81l --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate. No
CHANGELOG/LOGBOOK edit. Never spell any employer name.
