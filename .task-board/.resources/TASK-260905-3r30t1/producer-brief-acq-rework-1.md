# Producer brief: byte-exact acquisition — rework 1 (hosted CI platform-case gate)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch
`feat/byte-exact-acquisition`, head `5beced46`, PR https://github.com/relux-works/curator/pull/58.
Hosted CI (`gh pr checks 58`; logs via `gh run view <id> --log-failed`) fails `Test` and `Race` on
every OS with exactly these platform-case-gate findings — the local gates were green because the
ledger gate runs only in CI:

- `FAIL required on linux but not compiled into that build: internal/gitops :: TestArchiveRejectsLinks`
  (and darwin) — `.github/ci/platform-cases.tsv` line 82 requires the case
  `internal/gitops TestArchiveRejectsLinks` on linux,darwin (windows excluded: needs `ln`). The rework
  renamed or removed it.
- `FAIL skip with an unrecognised reason on linux/darwin: internal/interop :: TestConformanceSnapshotAcquisition`
  — `.github/ci/skip-classes.tsv` is the closed vocabulary of legitimate skip reasons; the new test's
  skip text ("root has no vectors/snapshot-acquisition.json" or similar) matches none.

Read `.github/ci/platform-case-gate.sh`, `platform-cases.tsv`, `skip-classes.tsv`,
`ledger-consistency.sh`, and `gate-selftest.sh` (the gate has its own self-test — any TSV change must
keep it green) before editing. Then:

1. Restore a test named exactly `TestArchiveRejectsLinks` in `internal/gitops` that proves the
   object-database extraction refuses a tree containing a symlink (the ledger's row describes it as
   "archiving refuses a tree containing links"; keep the row's platform claim: it needs `ln`, so it
   stays required on linux,darwin and excluded on windows — or, if your refusal test no longer needs
   `ln` because you build the tree with `git update-index --cacheinfo 120000`, update the row's
   platforms/reason accordingly and prove the case runs on all three GOOS).
2. Register the new cases the ledger should know: `internal/gitops` byte-exact tests and
   `internal/interop TestConformanceSnapshotAcquisition` — read `ledger-consistency.sh` to see whether
   unmentioned cases are a failure ("coverage gap") or merely unlisted, and add rows only where the
   gate requires them.
3. Give the interop test a legitimate skip: either add a skip class to `skip-classes.tsv` for
   "conformance root predates vector <name>" with a precise pattern and a policy line (mirror an
   existing class; check `gate-selftest.sh` asserts), or avoid the skip entirely on the pinned rc.9
   root by asserting the vector's *absence* is the expected state there and running the case only
   through `CURATOR_CONFORMANCE_ROOT` candidate lanes — choose the shape the gate scripts accept, and
   say why in the report.
4. Reproduce the gate locally before pushing: run the platform-case gate the way `ci.yml` does
   (`bash .github/ci/test-gate.sh` / `platform-case-gate.sh` with `CI_GATE_GOOS=linux|darwin|windows`
   over your `go test -json` output — read `ci.yml` lines 100–130 for the exact invocation) and paste
   its output; run `bash .github/ci/gate-selftest.sh` if you touched a TSV.
5. Gates as before: `go build ./...`, `go vet ./...`, `gofmt -l`, focused `-race` packages; do not rerun
   `./cmd/curator` unless you touched it. Signed commits on top of `5beced46` (no rewrite); plain push
   (no force) so PR #58 updates; then watch `gh pr checks 58 --watch` until Test/Race are green (the
   adapter suites' known x86 redness is a different failure class — if only those remain red, say so
   with the job names). Attach `TASK-260905-3r30t1_rework-report-1.md`; then
   `task-board handoff TASK-260905-3r30t1 --role developer`. Never write into the control root.
