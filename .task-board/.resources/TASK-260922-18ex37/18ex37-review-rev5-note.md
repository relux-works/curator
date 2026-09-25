# Review note — TASK-260922-18ex37 revision 5 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 4 was ACCEPTED on content; revision 5 is the refresh onto ab34556e (`18ex37-refresh-5.md`). Verify:
0. BLOCKING by rule (orchestrator pre-check): the candidate adds `.github/workflows/ci.yml.merged.tmp` — a merge artefact. Record it as a
   finding (removal required) and finish the full review so one rework fixes everything.
1. Every other path equals revision 4's content (per-file patch-id) EXCEPT where trunk changed the same file since c278af4f — there
   both sides must be present: .github/ci/gate-selftest.sh must keep 3v8k23's per-lane timeout pins AND 306v4m-unrelated trunk rows AND this
   Story's gap-ledger/pin rows; name any other overlap and check it.
2. Spec pin still dcc7f015 everywhere; the 5p8b0z gap row absent; CHANGELOG.md equals trunk; no root TASK-*/BUG-*, .temp/, test/, ledger/.
3. Hosted gate green on all lanes (validation log). Focused checks only (host memory). accept_cr or changes requested. No LOGBOOK.md.
