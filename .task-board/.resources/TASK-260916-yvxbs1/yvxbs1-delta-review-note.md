# Delta review — TASK-260916-yvxbs1 rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev5 ACCEPTED (your verdict). Rev6 re-applies it on trunk 97e85642 (E1 landed): base = trunk, tree 8ac4bae1, gate green, 33 paths (rev5 had
32). Orchestrator check: 30 paths equal `git merge-tree --write-tree --merge-base 86552087 refs/campaign/wgt8vz-rev5-20260928 97e85642`.
Review ONLY the 3 that differ:
1. internal/envprofile/envprofile.go — the conflict resolution: E1's update/delta/confirmation path and E6's path-source preflight +
   default-update ordering both present; nothing of E1 dropped (compare with trunk 97e85642), nothing added beyond the combine.
2. .github/ci/platform-cases.tsv — rows from both sides, each case once, no weakened must/skip columns.
3. internal/envprofile/path_source_group_write_unix_test.go — a NEW file not in rev5: what does it test, is it correct per environments §4
   (group-writable path source → permissions boundary failure), and why was it added during a re-apply (the brief said add nothing)?
   If it is a legitimate strengthening, accept; if it masks something, request changes.
Run `go test ./internal/envprofile -run 'Path|Boundary|Update|Delta|Signer|Guarded'` with a real exit code. accept_cr or changes requested
with file:line. No LOGBOOK.md.
