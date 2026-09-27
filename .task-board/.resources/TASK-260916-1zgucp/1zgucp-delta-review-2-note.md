# Delta review 2 — TASK-260916-1zgucp rev5 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Your rev4 verdict flagged F1 (RawStdEncoding) and F2 (projected comparator) as content in neither side. The producer showed both were needed
by the pinned rc.13 manager-config-v2 vectors, and the orchestrator decided (`1zgucp-decision-1.md`, read it): §12.1 is a grammar
(`<key-type> <base64> [<comment>]`) plus string identity — accept padded/unpadded base64, refuse non-base64/unknown key type (negative row);
for vectors whose `expected` is a projection ({"environments": …} — 24 of 56), compare exactly the declared top-level keys strictly, keep
trunk's full-object comparator and TestManagerEffectiveJSONComparisonRejectsExtraKnobs verbatim, add a row proving an extra knob inside
`environments` fails a projected vector.
Rev5: tree 1b69fd0c, base d41da0fb, gate green; vs rev4 (3a680a1e) it changes 5 files (environments.go, environments_conformance_test.go,
environments_test.go, contextresolve.go, contextresolve_test.go). Review ONLY those rev4→rev5 changes against the decision: grammar gate not
widened beyond §12.1 (mutant: accept any string → negative row must fail), identity normalisation consistent in contextresolve (padded vs
unpadded same key matches; different material never), comparator per decision (mutant: projection also used for full-object vectors →
TestManagerEffectiveJSONComparisonRejectsExtraKnobs must fail). The rest of rev4 was already checked in your rev4 verdict. Real exit codes.
accept_cr or changes requested with file:line. No LOGBOOK.md.
