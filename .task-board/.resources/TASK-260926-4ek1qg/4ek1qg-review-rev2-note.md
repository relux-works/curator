# Review note — TASK-260926-4ek1qg revision 2 + exact head of curator-spec PR #88 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Operator decision 2026-09-26 (option 1): cocoaskills claims core 1.0.0-rc.10 + skillfile-sources-v1 only; the spec CI checks each
implementation against its claim. Review revision 2 against `4ek1qg-rework-1.md`:
1. Go/curator lane: pin 0a628621, full candidate core, sparse checkout without .task-board/ (the suite reads nothing under it).
2. cocoaskills lane: pin 4a88aa0e; core root = tag v1.0.0-rc.10 conformance/v1 (confirm the tag object/commit is the real rc.10);
   skillfile-sources suite + schemas from the candidate; CSK_* env wiring matches what cocoaskills 4a88aa0e reads (read their test code);
   the Windows-off condition for their draft step is justified (their own CI is POSIX-only for it) and stated.
3. implementation_coverage.py `--implementation` change is correct and tested; ledger rows honest; the refusal-claim narrowing is stated.
4. Exact head of PR #88 = 2c39c42: b202b5d (accepted erratum, verified before) + 2c39c42 (per-file patch-id identical to your rev2 CR).
5. `gh pr checks 88 --repo relux-works/curator-spec` on head 2c39c42: every required check green incl. Implementations on all three OSes.
   Wait (bounded, ≤45 min) and report the run ids.
accept_cr only if 1–5 hold; else changes requested with file:line / failing check. No LOGBOOK.md.
