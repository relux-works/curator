# TASK-260823-2erkhe implementation evidence

- Branch: `fix/windows-candidate-identity-digest`
- Commit: `f81d8b4` (`fix(ci): canonicalize candidate digests`)
- Pull request: https://github.com/relux-works/curator/pull/23

## Change

- Candidate file hashing now passes file bytes through stdin, preventing Git for Windows `shasum` filename escaping from prefixing the digest.
- Candidate digests must be exactly 64 lowercase hexadecimal characters; malformed output fails closed.
- The pin digest calculation and self-test fixture calculation use the same stdin-safe form.
- `gate-selftest.sh` simulates Git for Windows filename escaping and separately proves that a truly backslash-prefixed digest is rejected.

## Validation

| Command | Exit | Result |
| --- | ---: | --- |
| `bash -n .github/ci/candidate-suite.sh .github/ci/gate-selftest.sh` | 0 | pass |
| `bash .github/ci/gate-selftest.sh` | 0 | 78 passed, 0 failed |
| `shellcheck .github/ci/candidate-suite.sh` | 0 | pass |
| `shellcheck -e SC2016,SC2329 .github/ci/gate-selftest.sh` | 0 | pass; exclusions are two pre-existing unrelated informational findings |
| `actionlint .github/workflows/ci.yml` | 0 | pass |
| `golangci-lint run` | 0 | pass |
| `make build` | 0 | pass |
| `go test ./...` | 1 | environment failure: isolated worktree initially lacked the test-tool submodule, then shared system temp exhausted disk space |

GitHub Actions run `32635797688` has green gate-selftest jobs on macOS, Ubuntu, and Windows; the remaining PR checks were still running when this artifact was first written.
