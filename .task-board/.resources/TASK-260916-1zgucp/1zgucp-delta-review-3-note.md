# Delta review 3 — TASK-260916-1zgucp rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev5 ACCEPTED (delta review 2). Rev6 re-applies it on trunk 86552087 (E4, E3, ryh3kw, 2n0233 landed): base 86552087 = trunk, tree 1c843a21,
gate green, same 31 paths. Orchestrator line check vs (trunk ∪ rev5): 28 paths clean. Review ONLY these 3:
1. .github/ci/root-artifacts.tsv — the cmd/curator row (vectors list) merged from both sides: every vector named by trunk and by rev5 present
   once, nothing dropped.
2. internal/envprofile/envprofile.go — `updateLocked(op, home, name, options.Policy, options.Surfacing…)` call(s): 3 lines in neither side, 4
   trunk lines replaced — confirm it is the combination of trunk's signature (ryh3kw/E3 arguments) and E1's delta/confirmation arguments,
   with no behaviour lost from either side.
3. internal/envprofile/status.go — 2 new lines `if lock, _, lockErr := readLock(home, entry.Name()); lockErr == nil {`: confirm this is
   E1's signer-posture status read and that it respects ryh3kw's §8.4.1 discipline (absent vs unreadable; an unreadable lock must not be
   silently treated as absent — if it is, that is a finding) and TestManagerOwnedAbsenceReadsAreGuarded passes.
Run `go test ./internal/envprofile -run 'Status|Signer|Delta|Update|Guarded|ReadFailure'` and `go test ./cmd/curator -run
'Profile|Signer|Delta|Status'` with real exit codes. accept_cr or changes requested with file:line. No LOGBOOK.md.
