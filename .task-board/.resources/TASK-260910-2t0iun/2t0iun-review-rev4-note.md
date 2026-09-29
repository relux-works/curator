# Review note — TASK-260910-2t0iun rev4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev4 (base 3f60f7f0, tree 3db884ae, 6 paths, gate green) against the rev3 verdict F1 and `2t0iun-rework-1.md`. Verify:
1. The gh branch passes exactly one signer pin. Run the real gh on this host (`gh attestation verify --help` plus the exact argv
   install.sh builds, against a dummy file) and show that flag parsing accepts it: no "if any flags in the group" error, real exit code.
   The --signer-workflow value format matches gh's help.
2. The fake gh rejects mutually exclusive signer flags the way real gh does, and the rows assert the exact argv for each branch. Re-run
   the both-flags mutant yourself and confirm it is killed (real exit code).
3. Everything rev3 had OK is still OK: order, exact checksum match, nothing written before verification, refusal without a verifier,
   loud opt-out, the anchored identity regexp, and the Windows ledger rows. Diff rev3→rev4 touches only what F1 needs.
accept_cr or changes requested with file:line. No LOGBOOK.md.
