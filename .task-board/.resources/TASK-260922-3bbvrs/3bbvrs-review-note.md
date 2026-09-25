# Review note — TASK-260922-3bbvrs gap ledger, CR revision 2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review revision 2 against `3bbvrs-brief.md`, the AC, `3bbvrs-rework-1.md` and the results. Read-only; disposable clone.
1. One committed ledger read by every consumer that renders EVERY published case (TestManagerConfigV2Vectors, the
   system-config-v2 sibling, and whatever other families the producer listed — check the list is complete with your
   own grep for loops over published case lists). Each case classified driven | known-gap | bound | skipped, counts
   tallied, total asserted against the published count.
2. The ratchet: a known-gap row that passes FAILS the gate; an unlisted failing case fails; a vanished case fails.
   The three narrowing mutants (drop tally, accept unlisted gap, let a passing gap stay listed) each killed by a named
   test — re-apply at least one yourself in the clone.
3. Rework 1: the 8 buildsource conformance rows no longer skip with an unregistered reason — either driven, or
   classified with a reason matching ONE narrow registered skip class plus a gate-selftest row proving an unregistered
   reason still fails; `no-broad-suppression.sh` green. No existing class broadened.
4. SPEC_PIN not moved. Docs say a gap row is an owed implementation, never an accepted deviation. CHANGELOG.
5. Runtime validation log green.
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md.
