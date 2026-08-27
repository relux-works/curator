# TASK-260823-czs1cx — Windows environment and toolchain vector reconciliation

## Root causes and changes

- The candidate generator published a Darwin/arm64 closed Go environment as a universal value. The superseding candidate publishes explicit Darwin/arm64, Linux/amd64, and Windows/amd64 cases. The Windows case includes the required private process variables and permits only the indispensable optional `SYSTEMROOT` and `WINDIR` variables.
- The authoritative toolchain digest and preimage were platform-neutral. Curator's conformance materializer changed the published symlink target from `../bin/go` to `..\bin\go` on Windows before hashing it. Curator now materializes and verifies the exact protocol target bytes.

## Delivery identities

- Curator branch: `fix/TASK-260823-czs1cx-windows-vectors`
- Curator commit: `fbca88617c3765cfa40c1284035429962bf81bda`
- Draft PR: https://github.com/relux-works/curator/pull/29
- Superseding `curator-spec` candidate branch: `candidate/schema-8-rc.9`
- Candidate commit: `edd07210d4f3db34fd60238cb14b90f837de03cb`
- Manifest SHA-256: `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`
- Tree SHA-256: `sha256:9d5a10b6ef1bd867f4d055d830d10a240620d759ff245fed9ccdb40b888ab769`
- File count: `692`
- Previous candidate `859727b103ed175ff214cbb64641f4686d8c6a68` was not rewritten.

## Local validation (real exit codes)

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test ./internal/godriver -run '^TestFixedEnvironmentForHostSelectsNativeCase$' -count=1` | 0 | Host-case selector regression passed. |
| Candidate-root focused environment/toolchain Go tests | 0 | Both target conformance tests passed. |
| `go test ./tools/generate-vectors -count=1` (first attempt) | 1 | Compile failure exposed an incorrect test-helper name; corrected before the green rerun. |
| `go test ./tools/generate-vectors -count=1` | 0 | Generator tests passed after correction. |
| Python unittest direct module invocation | 1 | Expected setup failure: wrong import mode, then missing pinned `jsonschema`; no product assertion ran. |
| Task-local venv install from `requirements-dev.txt` | 0 | Installed exactly the repository-pinned validation dependency. |
| Post-regeneration `tools/validate.py` | 0 | Validated 53 schemas and 691 vector files. |
| Post-regeneration Python unittest discovery | 0 | 98 tests passed. |
| Post-regeneration `go test ./tools/...` | 0 | Generator package passed. |
| `gofmt -l tools` | 0 | No output. |
| Regeneration pass 1 | 0 | Generated suite and rc.9 metadata. |
| Regeneration pass 2 | 0 | Generated the same bytes. |
| Pass-1/pass-2 checksum `cmp` | 0 | Byte-identical inventories. |
| Curator `go test ./...` | 1 | Resource failure: unrelated `cmd/curator` test hit its 10m timeout under concurrent host-GOROOT hashing and atomicity tests reported ENOSPC. Changed `internal/godriver` package passed. |
| `golangci-lint run` | 0 | 0 issues. |
| `make build` | 0 | Curator binary built. |
| Full local candidate gate | 0 | 41 served, 0 deferred, 0 excluded; Go tests and platform-case gate passed. |

The later green full candidate gate reran the package set that encountered the contention failure; the formerly timed-out rollback-recovery test passed in 276.89s.

## Remote validation

- `curator-spec` Specification CI: https://github.com/relux-works/curator-spec/actions/runs/32642316308 (pending terminal result at initial artifact write).
- Curator PR CI: https://github.com/relux-works/curator/actions/runs/32642306296 (pending terminal result at initial artifact write).
- Curator candidate-conformance dispatch: https://github.com/relux-works/curator/actions/runs/32642340559 (pending terminal result at initial artifact write).

`SPEC_PIN` remains unchanged. The candidate evidence is candidate-only, not a release or conformance claim.
