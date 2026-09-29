# TASK-260910-32gki6 — Windows gate fix 4 (THE ONLY CURRENT INSTRUCTION, with the earlier S5 briefs)

Rev4 (tree 61f5b184): Windows is down from 261 to 3 failures (run 36496984924). Fix exactly these, without relaxing the DACL check:
1. internal/envprofile TestLegacyPathInstallIgnoresStoreOverlap (boundaries_test.go:49): `profile install from a context-store path:
   permissions: …\001\contexts boundary check failed` — decide per environments §4 whether a legacy path install whose source overlaps the
   context store is checked as a path source (then the fixture must create that directory protected on Windows) or ignored (then the code
   must not run the path-source DACL check on it); follow the test's name/intent and cite the clause.
2. internal/envprofile TestStoreBoundaryCheckedRootsExcludeCuratorHome (store_boundary_conformance_test.go:514): the test helper fails with
   `cannot protect a tree containing a symbolic link: …\environments\acme\c…` — the managed home legitimately contains manager-created
   links; the helper must protect directories/files and leave (not follow, not refuse) the manager's own link entries, or the test must
   restore permissions only on the store entry it changed.
3. internal/envprofile TestRepairTakeoverProvisionsManagedHome (takeover_test.go:296): `environments root failed permissions check … at
   …\001` — in this test the environments root resolves to the test's temp dir itself; create it through the protected helper (fixture) or
   point the test's environments root at a manager-created subdirectory, matching what production does.
`GOOS=windows go vet ./internal/envprofile`; unix `go test ./internal/envprofile -run 'Legacy|Boundary|Takeover|Store'` with real exit codes.
Set status development; update results (`git diff 61f5b184` non-empty); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
