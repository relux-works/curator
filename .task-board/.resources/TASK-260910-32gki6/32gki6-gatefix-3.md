# TASK-260910-32gki6 — Windows gate fix 3 (THE ONLY CURRENT INSTRUCTION, with 32gki6-sec-brief.md, 32gki6-decision-1.md, 32gki6-gatefix-2.md)

Rev3 (tree bafbeddb) is green on linux/macOS; Windows now fails 261 tests (run 36490110006). Representative:
TestResolveProvisionRepair — `environment_store_untrusted: profile store root failed permissions check: … boundary check failed at
C:\…\TestResolveProvisionRepair…\001: … DACL grants mutation rights to another identity` — and many `profile_use_partial` follow-ons.
`…\001` is the CURATOR HOME (the test's home dir), not a manager-created directory. environments.md §4 names exactly what must be
protected: "the environments root, the profile store root, the profile lock file, the environment marker of every home …, and every store
entry the lock names" — the enclosing-boundary check applies to the environments root and the profile store root THEMSELVES (and below),
not to the Curator home or its ancestors (those are the operator's directories, created by the OS/installer with inherited ACEs). Fix:
(1) resolve the checked roots to the exact manager-created directories (e.g. <home>/environments and <home>/profiles — verify the real paths
in code) and stop checking the home/ancestors; (2) keep creating those roots and everything below them with the owner-only DACL (your rev3
change) so fresh Windows homes pass; (3) investigate the profile_use_partial failures — they must disappear once (1) is right; if any remain,
find the write that now fails under the owner-only DACL (e.g. a child created by code that does not use the protected helper) and route it
through the helper. Do not relax the DACL check itself. `GOOS=windows go vet ./...`; unix `go test ./internal/envprofile` split with real exit
codes. Set status development; update results (`git diff bafbeddb` non-empty, listing the checked roots); handoff; END YOUR TURN.
No CHANGELOG/LOGBOOK edit.
