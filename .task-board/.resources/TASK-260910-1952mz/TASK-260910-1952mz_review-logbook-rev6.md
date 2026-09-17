# TASK-260910-1952mz review logbook — revision 6

Accepted for integration. Candidate production bytes are unchanged from revision 5; revision 6 fixes only the changed-A sourced-marker assertion. Independent narrow tests passed, including 14/14 external vectors with real pwsh; 2/2 narrowing mutants killed. Previous malformed-record, POSIX parse, symlink and atomic-publication corrections remain covered.

Non-blocking observation: Windows interpreter dedupe keys are case-sensitive resolved strings, so duplicate spellings may cause redundant runs. Unicode case identity beyond the tested native/MSYS ASCII examples is not proven. An independent same-file case probe through the actual PowerShell hook passed both profiles locally; no defect inferred from a code-only suspicion.

Operational anomalies: a broad test invocation was intentionally terminated (143), replaced by bounded checks; its incomplete transcript is retained and never counted as a pass. Several shell launches temporarily stalled, then recovered. Board attachment and verdict are completed only after recovery. Repository LOGBOOK.md and candidate code were not edited.
