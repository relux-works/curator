# TASK-260922-18ex37 review verdict — revision 4: ACCEPTED

Candidate tree f1f7f19e on base c278af4f; reviewed against 18ex37-rework-3.md and the rev4 note.

- SPEC_PIN = dcc7f015e2d9… (.github/workflows/ci.yml); #89/v9 excluded as the operator decided.
- The `windows-exec-uncaptured-systemroot-hardlinks` row is gone. Ledger: 69 data rows, each with an owner (0 rows have an empty owner column).
- CHANGELOG.md blob equals trunk (sha256 97c55451…). Both entries are in results.md. The candidate has no root TASK-*/BUG-*, .temp/, test/ or ledger/ paths.
- Hosted gate run 36090141585: success on ubuntu, macOS and Windows for Test, Gate self-test and Race, plus Lint, Naming and Interop. rose-air was skipped. The gate commit tree d11fc174 differs from the candidate only in 33 .task-board paths (0 non-board paths).
- Diff against rev3: the only differences are these.
  1. conformance-gaps.tsv rows 65–66 move from owner STORY-260822-2h0v9j (done) to STORY-260925-1v7pvn (backlog, exists), with the reason text reworded.
  2. internal/scriptworker/exec_identity_conformance_test.go: the trunk exhaustive driver added by BUG-260924-5p8b0z is narrowed to TestUncapturedSystemRootHardlinkRegression. The exhaustive family driver in executable_identity_conformance_test.go still covers all cases.
  3. executable_identity_conformance_test.go:149-153 sets an ambient SYSTEMROOT for the `caller-or-package-value` case, so the uncaptured case passes at the production resolver.
  All other file hunks are identical to rev3.
- I reran the tests myself on a `git archive` copy of the candidate, with CURATOR_CONFORMANCE_ROOT set to the dcc7f015 conformance tree: `go test ./internal/scriptworker ./internal/scriptpolicy ./internal/conformancecoverage` exited 0.
- Mutant: I deleted the `windows-exec-unowned-file-hardlinks` gap row. TestExecutableIdentityCasesAtProductionEntry failed (exit 1), so the gate is killed.
- Bound: on darwin only 5 identity cases are counted locally. Windows counting relies on the hosted Windows lane, which is green.
