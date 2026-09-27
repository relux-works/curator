# TASK-260916-55g9dg — E2 on current trunk (THE ONLY CURRENT INSTRUCTION, with 55g9dg-sec-brief.md)

Your previous run worked on an ANCIENT workspace (base 1de6f8e1, SPEC_PIN rc.11 87a0d006, leftover deltas from 2026-09-17) — that is why
the gate contract looked impossible. The orchestrator snapshotted it (refs/campaign/2d9coh-full-20260927, parent 1de6f8e1) and DISCARDED it;
your new workspace starts on trunk (SPEC_PIN = curator-spec v1.0.0-rc.13 tag commit 23435129, attributed gap ledger .github/ci/conformance-gaps.tsv).
1. Re-implement E2 (direct-only system modules: context_system_module_transitive + its scoped waiver; context-system-module-present stays
   always-warn) on the current code; use the snapshot only as a reference (`git diff 1de6f8e1 refs/campaign/2d9coh-full-20260927`), never
   apply it blindly — trunk moved a lot.
2. Gap ledger contract (current trunk, 3bbvrs/18ex37): a pinned case whose FIRST production blocker is a DIFFERENT, not-yet-implemented
   surface stays a gap row owned by that surface's Story. The rc.13 E2 schema cases that also carry E4 provider-trust-root fields: if they
   still fail only because of E4, re-attribute them to STORY-260916-2otjbn (E4, next leaf) with the exact blocker; every case E2 alone makes
   pass LEAVES the ledger. Report before/after owner histogram.
3. Mutants per rule; `task-board m 'set_status(TASK-260916-55g9dg, status=development)'` first; bounded runs; handoff.
No CHANGELOG/LOGBOOK edit. A write-boundary `policy warn` block is a warning.
