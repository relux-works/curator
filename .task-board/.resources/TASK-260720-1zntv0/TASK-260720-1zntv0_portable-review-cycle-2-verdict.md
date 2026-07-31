# TASK-260720-1zntv0 independent review cycle 2

## Verdict

ACCEPTED. The cycle-2 rework closes R1 and R2 and matches the portable manager-worker-v1 acceptance criteria and the accepted rc.5 contract. No remaining implementation, architecture, or test finding requires rework.

## R1 closure

Native-control ownership is now in the manager parent. Per-operation probes use the normalized operation limits. The control domain is prepared before process creation; Darwin installs the private session at fork/exec and inherits the exact RLIMIT_FSIZE bound across the fork, while Windows creates the worker suspended, assigns the exact Job Object and restricted handle launch before resuming it. Capability evidence is derived only from domain.installedControls after installation, validated before request delivery, independently confirmed by the worker, and required to match byte-for-byte at the structured record level.

The three injected seams availability-probe, domain-preparation, and domain-installation all reject with build_execution_control_unavailable before worker session execution, with no compiler call, no applied evidence, no staged artifact, and no hardened-policy drift. The Windows native log exercises the suspended-launch installation failure path; the Darwin native focused log exercises the session, process-group teardown, and RLIMIT_FSIZE mechanisms. Normatively unavailable inventory controls and all six deferred hardened guarantees remain non-rejecting and absent from evidence.

## R2 closure

runGo now checks command.ProcessState before reading ExitCode. A refused creation produces StartFailed=true, ExitCode=-1, Started counting the single attempted creation site, bounded empty output, and no artifact. The parent maps the result to the existing go_list_failed or go_build_failed phase diagnostic and does not continue past a refused list. Real worker tests prove no panic, clean shutdown, complete domain teardown, and no publication. The Windows native gate additionally proves a real Job Object active-process refusal is handled through this path.

## Provenance and independent inspection

- Reviewed worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree.
- HEAD remains 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 with zero later commits and a clean index.
- internal/godriver sorted-file tree digest independently recomputed as 55e23e323a84ff76be4768ed664f08a61cb2975cbc4c5bb50bb77a9fdeccda19.
- Accepted conformance manifest SHA-256 independently recomputed as 58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9.
- Accepted host-execution vector SHA-256 independently recomputed as c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de.
- Cycle-2 native logs and their 143-file byte-match evidence were inspected. macOS amd64 full, vet, real-build, and focused gates pass. Windows amd64 build, vet, rc.5 task-package, focused R1/R2, and real-build gates pass. The reported Windows full-suite red remains confined to five unchanged predecessor packages and is not a task regression. Linux remains honestly compile/vet/test-link coverage only, as required by the inventory scope.

## Independent reviewer gates

All commands ran against the handed-off bytes and exited 0:

- go build ./...
- go vet ./...
- go test ./... -count=1
- go test -race ./... -count=1
- CURATOR_CONFORMANCE_ROOT=<accepted rc.5> CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver ./internal/buildmeta -count=1
- focused R1/R2 seam, evidence, refused-child, teardown, RLIMIT_FSIZE, real vendored build, policy-vector, and cache-identity tests
- scoped golangci-lint for internal/godriver, internal/buildmeta, and cmd/curator: 0 issues
- gofmt check and git diff --check
- Windows amd64 test-binary link; Linux amd64 test-binary link, build, and vet; Linux arm64 build

The focused real vendored build records only the fixed manager-selected Go probes, one go list, and one go build and never launches the artifact.

## Reviewer boundary

No implementation, test, candidate, release, ref, pin, or provenance file was modified, staged, committed, or published during review.