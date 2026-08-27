# TASK-260823-3fnobk delivery evidence

## Landed change

- Reviewed patch SHA-256: `4d62e862132f91d925258a3475375e9ba554d3301e66d4b50f5c2a40ae88752b`.
- Applied delta SHA-256: `4d62e862132f91d925258a3475375e9ba554d3301e66d4b50f5c2a40ae88752b` (byte-identical).
- Source commit: `c73bc1339946b31a21cf9f784448ea3acac4fbb8`.
- PR: https://github.com/relux-works/curator/pull/26
- Merge commit on `main`: `c6092af0f7d01617832a1307832121e8853b11bc`.

The change removes synthetic case-folded/NFD path collision rejection from
`internal/buildsource`, while preserving portable-path validation and exact
encoded duplicate rejection. Unit tests now assert that case-distinct and
normalization-distinct encoded paths are admitted; the pre-existing exact
duplicate test remains.

## Local validation

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/buildsource -count=1` | 0 | Pass |
| `go test -cover ./internal/buildsource -count=1` | 0 | Pass; 81.8% statement coverage |
| `golangci-lint run ./internal/buildsource/...` | 0 | Pass; 0 issues |
| formatting check for the two changed Go files | 0 | Pass |
| `git diff --check` | 0 | Pass |
| `make build` | 0 | Pass |
| rc.9 focused candidate invocation on local macOS | 0 | Overall pass; leaf skipped because the filesystem folded the alias into one member |
| `go test ./...` | 143 | Not green: `cmd/curator` hit its 10-minute timeout in unrelated `TestCompiledProjectStatusRepairRollbackRecovery` under concurrent full-suite load; after its test child exited, the hung `go test` driver was terminated |

## Ubuntu rc.9 candidate evidence

- Candidate commit: `859727b103ed175ff214cbb64641f4686d8c6a68`.
- Candidate manifest SHA-256: `782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f`.
- Dispatch run: https://github.com/relux-works/curator/actions/runs/32641306472
- Ubuntu candidate job: https://github.com/relux-works/curator/actions/runs/32641306472/job/97198706059
- Job conclusion: success on Ubuntu 24.04.
- `TestBuildSourceIdentityVectors/duplicate-build-source-path`: non-skipped `pass` in the downloaded `go-test-served.json` and `observed-cases.tsv` evidence.

The remaining auxiliary dispatch jobs were cancelled after the successful
Ubuntu candidate artifact was downloaded; they cover independent known Windows
candidate regressions and were not merge gates for this focused fix.

## CI

- PR CI run: https://github.com/relux-works/curator/actions/runs/32641227472
- Post-merge `main` CI run: https://github.com/relux-works/curator/actions/runs/32641573792

