# TASK-260910-16k7xy revision 6 — ACCEPT

Candidate tree b03d0985a177ac55c3f11ba4c5ddfcccf8fb3ab8; base 1de6f8e12f33212843820be756923b1179063d2c. HEAD c79d0fc is the accepted 14hsti checkpoint. The eight leaf paths match the candidate byte-for-byte; the remaining story delta is the checkpoint. No production/test source changed during review; mutants used Go overlays in ignored .temp.

Reviewed against skillfile-sources §3 and the three conformance snapshot vectors. Production path: envprofile.stateForPath -> admitPathSource -> PrepareLocalAcquisition -> Capture -> PublishLocal -> OpenLocal -> contextstore.EnsureState. DraftSourcesV1 bounds the new behavior. All admitted inputs are captured, including dirty/untracked Git files and runtime/build/dependency files; no Git commit synthesis. Frozen bytes feed installation. Missing/tampered snapshots fail closed.

Prior findings resolved: capture.go:223,276,483-509 retain and compare eager physical identity tokens; identity_windows.go uses staging.FileIdentity (single handle volume/index), identity_unix.go uses device/inode. capture.go:372-379 rehashes actual publication bytes before rename. draft_capture_test.go:44-71 mutates live bytes between freeze and consumption within one Install. capture_test.go exercises an actual in-capture content race.

Independent commands (zsh; observed process exit codes):
- go test -p 1 ./internal/snapshot -run 'Capture|Vector|Publish|Open|ReviewIdentity' -count=1 -timeout=90s: 0, 2.283s; includes 3/3 normative vectors.
- go test -p 1 ./internal/envprofile -run 'Draft|AdmitPath' -count=1 -timeout=90s: 0, 13.264s.
- go vet ./internal/snapshot ./internal/envprofile: 0.
- gofmt -l internal/snapshot internal/envprofile: no output; git diff --check: 0.

Independent narrowing mutants, all -count=1 -p 1 -timeout=90s with Go -overlay; 4/4 killed (each exit 1):
1. Identity mismatch refused only when token length also differs: TestReviewIdentityReplacementDuringCapture fails (Capture returns nil).
2. Live content digest comparison narrowed to file-count comparison: TestCaptureDetectsLiveContentChange fails (Capture returns nil).
3. Publication digest comparison narrowed to file-count comparison: TestPublishRefusesCopyMutatedAfterAudit fails (PublishLocal returns nil).
4. Consume stored source only when its path is empty: TestDraftInstallFreezesLiveMutations fails (installed mutated bytes).
These are targeted bounds, not exhaustive mutation coverage. Original source files were never edited.

Hosted evidence independently inspected: https://github.com/relux-works/curator/actions/runs/35169169704 ; commit 09b3afb00181d910a6c6e38ff368ed6ad78f4e2c has exactly the reviewed tree (GitHub commit API). Lint, Ubuntu/macOS/Windows tests, Ubuntu/macOS race and conformance lanes succeed. Full suite was not rerun locally. Windows test-evidence-windows-latest/go-test-served.json explicitly records PASS for identity replacement, post-audit publication mutation and frozen install consumption. The three executable normative subcases SKIP on Windows with the accepted platform-control exception; 3/3 run on local Darwin. rose-air and Candidate suite lanes skipped; no claim of rose-air verification.

Evidence caution: results_rev6.md is stale (describes an earlier SameFile implementation and six paths). This acceptance relies on actual candidate bytes, independent checks and exact-tree hosted artifacts, not that description.

Checklist: contract and production wiring reviewed; three vectors verified; refusal regressions independently run; four narrowing mutants killed; exact hosted tree verified; candidate unchanged; no blocking findings. ACCEPT revision 6, route through accept_cr to integrating, never reviewer done. spawn goal reports not goal-bound; no directives.
