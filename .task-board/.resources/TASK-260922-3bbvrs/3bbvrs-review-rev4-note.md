# Review note — TASK-260922-3bbvrs revision 4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 3 was CHANGES_REQUESTED only for F1: three `cmd/` whole-family loops unrouted (umbrella-provider-resolution cases 14,
manager-lifecycle bootstrap-cases 3 and upgrade-cases 3). Verify in a disposable clone: all three routed through
`conformancecoverage.Run`/`RunOutcomes` with pins in `conformance-case-counts.tsv` (+ root-artifacts.tsv if required); the
producer's module-wide sweep (not only internal/) with a disposition per hit — repeat the sweep yourself over the whole module;
one narrowing mutant on a newly routed family re-applied by you → killed; rev3→rev4 patch diff shows nothing else; validation
log green. accept_cr on revision 4, or changes requested with file:line. No LOGBOOK.md.
