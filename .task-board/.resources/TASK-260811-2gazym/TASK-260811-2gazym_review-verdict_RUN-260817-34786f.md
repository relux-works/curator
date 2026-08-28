# Independent reviewer verdict for TASK-260811-2gazym

Verdict: **accepted -> done**

## Review authority and immutable inputs

- Reviewer run: `RUN-260817-34786f`
- Authoritative goal immediately before verdict: `GOAL-260817-6a3a20` revision 1
- Resolved scope: `TASK-260811-2gazym`
- Review policy: `required`
- Reviewed producer run: `RUN-260817-5cd5d5`
- Reviewed producer evidence: `TASK-260811-2gazym_rework-evidence_RUN-260817-5cd5d5.md`
- Producer evidence SHA-256: `4d61564e8bb320b2d8b616d2d7d790e7b55fc3bd856cd19fa246167effce76b3`
- Actionable prior verdict: `TASK-260811-2gazym_review-verdict_RUN-260811-5b088d.md`
- Reviewed Git HEAD: `8ff4a238f7725bada3cfb8aa7c9c135698483caa`
- Independently reproduced artifactpolicy fingerprint: `e9243e1d753e71427a920b15244f08d9476f54bcf154df8751a91839accc49f9`
- Independently reproduced corpus source SHA-256: `87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`

No product code was modified, staged, or committed by this reviewer.

## Acceptance findings

1. The package exposes closed artifact class, trust role, decision, diagnostic, profile, grammar, node-kind, and resolved-use vocabularies. Dependency admission is fail-closed, verified-binary admission remains unavailable, and authorization tokens are sealed to the service instance.
2. ZIP/ZIP64, tar/PAX/GNU, gzip, and native ar/COFF traversal recursively classify reachable content. Path normalization, portable collision handling, unsafe-entry rejection, depth/count/size/emitted-byte/ratio limits, and complete bounded findings are enforced during enumeration rather than after unbounded accumulation.
3. `artifact-manifest-v1` binds immutable origin/raw payload, policy, detector registry, full limit vector, role evidence, canonical nodes, traversal accounting, diagnostics/findings, and the final digest. Decoding independently derives semantic classifications and decisions and rejects self-rehashed node, role, path, metadata-association, and accounting forgeries.
4. The production external-toolchain path is manager-owned: the caller supplies only a closed selector and already admitted dependencies. The exact selected root and executable are fingerprinted, admitted, and rechecked before execution. Caller roots, package paths, caller evidence, and copied bytes cannot mint toolchain authority.
5. The external-package positive executes the exact absolute path returned by `AuthorizeSelectedAdapterExecution` through `exec.CommandContext`, with a 10-second context and 4096-byte combined-output limit. The process counter increments only after `Start` succeeds, no PATH lookup occurs, and the exact `go version` output is checked against the selected role-evidence version and GOOS/GOARCH.
6. Production negatives create a real escaping symlink and FIFO under a controlled selected root; both fail the real complete-tree recheck before any process start. A same-count dependency boundary with changed bytes/digest is rejected before launch. A real hard link from a pre-existing object receives no output admission or publication authority. Production A08 authority correctly remains unavailable until `TASK-260811-27xisf` owns protected execution/receipts.
7. The accepted bounded-metadata repair remains sound: PAX/GNU path metadata, GNU/COFF string-table names, and BSD extended names are checked against the complete virtual-path and leaf/emitted budgets before materialization. Native archive string-table references resolve only at exact record starts and unreferenced names are inspected order-independently.
8. BSD regular-member and `__.SYMDEF` evidence binds extended-name size/hash, original name, canonical logical/metadata path, declared/member size, and physical traversal accounting. Valid manifests round-trip; each required self-rehashed field and manifest-accounting mutation is rejected.
9. The shared corpus contains exactly 182 unique pinned A/C/F/T/V cases. Every case asserts class, node/final decision, primary diagnostic, authorization presence, canonical round-trip, and exact manifest digest. F14 preserves distinct exact raw-payload identities while its canonical logical evidence projection is order-independent. Compiled dependency bytes remain deny-dominant across adapter profiles.
10. Existing Go admission behavior is preserved by the focused baseline regression lane. Kotlin/Gradle/Maven are absent from `internal/artifactpolicy`; no deferred ecosystem or verified-binary capability was introduced.

## Independent validation

All commands ran against the current unchanged source and exited 0:

- targeted prior-blocker and exact-corpus tests: 4.600s;
- `go test -count=1 ./internal/artifactpolicy/...`: 35.191s;
- `go test -race -count=1 ./internal/artifactpolicy/...`: 334.746s;
- focused `go vet`, `go build`, `gofmt`, and pinned golangci-lint: clean, `0 issues.`;
- Go baseline regressions across `buildsource`, `buildmeta`, `buildcache`, `godriver`, `buildrepo`, `install`, and `install/atomicity`: all passed;
- `go test -count=1 ./...`: all packages passed; notable timings were `cmd/curator` 388.841s, `artifactpolicy` 164.133s, `godriver` 101.418s, `install` 134.568s, and `install/atomicity` 123.041s;
- `go vet ./...`, `go build ./...`, `gofmt -l cmd internal`, pinned `/Users/iv/go/1.25.5/bin/golangci-lint run ./...`, and `git diff --check`: clean;
- `task-board validate`: `Board is valid. No issues found.`;
- post-gate artifactpolicy fingerprint remained `e9243e1d753e71427a920b15244f08d9476f54bcf154df8751a91839accc49f9`.

## Route

The implementation satisfies the task description, acceptance criteria, all accumulated rework requirements, project architecture, test/lint/build gates, and the authoritative reviewer goal. Accept `TASK-260811-2gazym` and route it to `done`.

This reviewer supplies no `commit_ack`.
