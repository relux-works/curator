# TASK-260823-1l1p8q candidate-conformance rerun evidence

## Immutable inputs

- Curator workflow head: `95ca5ae837462463e84d27289c9ed6141f27c43d` (`main`)
- Candidate branch: `candidate/schema-8-rc.9`
- Candidate commit: `859727b103ed175ff214cbb64641f4686d8c6a68`
- Candidate manifest SHA-256: `782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f`
- Run: https://github.com/relux-works/curator/actions/runs/32638424105

Both prerequisite fixes were present on the dispatched `main`: `377d7a4` and
`95ca5ae` were each verified as ancestors of `origin/main` (exit 0). The
candidate commit resolved through GitHub, and a fresh download of
`conformance/v1/manifest.json` produced the expected digest (exit 0).

## Terminal matrix

| Job | Conclusion | Duration | Job ID |
| --- | --- | ---: | ---: |
| Candidate suite (macos-latest) | success | 7m17s | 97191572884 |
| Candidate suite (ubuntu-latest) | failure | 1m35s | 97191572950 |
| Candidate suite (windows-latest) | failure | 27m43s | 97191572958 |
| Test (macos-latest) | success | 10m38s | 97191572851 |
| Test (ubuntu-latest) | success | 1m37s | 97191572909 |
| Test (windows-latest) | success | 33m46s | 97191572905 |
| Race (macos-latest) | success | 9m57s | 97191572743 |
| Race (ubuntu-latest) | success | 3m16s | 97191572967 |
| Lint | success | 25s | 97191572912 |
| Interop conformance gate | success | 22s | 97191572867 |
| Naming gate | success | 9s | 97191572892 |
| Gate self-test (macos-latest) | success | 10s | 97191572902 |
| Gate self-test (ubuntu-latest) | success | 6s | 97191572906 |
| Gate self-test (windows-latest) | success | 25s | 97191572899 |

The workflow and `gh run watch --exit-status` exited 1. This is a red candidate
matrix and must not be used as green-candidate evidence.

## Candidate regressions

Ubuntu has one deterministic failure:

- `internal/buildsource/TestBuildSourceIdentityVectors/duplicate-build-source-path`
  rejects distinct encoded members `SAME` and `same` as a synthetic platform
  collision. The candidate requires distinct encoded paths to be admitted; an
  exactly repeated encoded path remains rejected.

Windows has five failing leaf tests across three ownership areas:

- `internal/buildsource/TestBuildSourceIdentityVectors/invalid-unicode-build-source-path`:
  an invalid Unicode path was accepted.
- `internal/godriver/TestFixedEnvironmentAndFiveDirectArgvFormsVector/fixed_environment`:
  observed `GOARCH=amd64`, candidate expected the closed `arm64` value.
- `internal/godriver/TestToolchainIdentityVectors/unsorted-directories-files-and-internal-link`:
  computed toolchain digest differs from the candidate digest.
- `internal/godriver/TestToolchainIdentityVectors/invalid-unicode-toolchain-path`:
  an invalid Unicode path was accepted.
- `internal/install/TestDryRunEffectBindingsSeeWhatARealOperationWrites`:
  real project install rejected staged script command `clonable-tool` as not
  executable.

The downloaded Actions evidence is retained under:

- `.temp/TASK-260823-1l1p8q/ubuntu-evidence-32638424105/`
- `.temp/TASK-260823-1l1p8q/windows-evidence-32638424105/`
- `.temp/TASK-260823-1l1p8q/ubuntu-candidate-32638424105.log`

## Scoped implementation produced

An isolated worktree at `.temp/TASK-260823-1l1p8q/worktree` contains a focused
fix for the Ubuntu build-source path regression. It removes synthetic
case-folding and Unicode-normalization collision rejection while retaining exact
encoded-path duplicate rejection. Unit tests now assert admission of case- and
normalization-distinct encoded paths.

Patch artifact:

- `TASK-260823-1l1p8q_buildsource-encoded-path-fix.patch`
- SHA-256: `4d62e862132f91d925258a3475375e9ba554d3301e66d4b50f5c2a40ae88752b`

The patch is intentionally unstaged and uncommitted under repository policy.

## Validation evidence and real exit codes

| Command/gate | Exit | Result |
| --- | ---: | --- |
| Pre-fix focused encoded-path unit test | 1 | Expected red: both over-rejection cases failed |
| Post-fix focused encoded-path unit test | 0 | Passed |
| rc.9 `go test ./internal/buildsource -count=1` | 0 | Passed |
| rc.9 `go test -cover ./internal/buildsource -count=1` | 0 | Passed, 81.8% statement coverage |
| `make build` | 0 | Passed |
| `gofmt` plus `git diff --check` | 0 | Passed |
| First full `test-gate.sh` call without evidence-dir | 2 | Failed invocation: required positional evidence-dir was missing |
| Full rc.9 gate before submodule initialization | 1 | Failed only because the isolated worktree lacked the declared `tuitestkit` replacement submodule |
| `git submodule update --init --recursive` | 0 | Initialized pinned submodule `21585d0e...` |
| Full rc.9 gate rerun, 41 served / 0 deferred / 0 excluded | 0 | Go test exit 0; platform-case gate exit 0 |
| GitHub Actions run 32638424105 | 1 | Candidate Ubuntu and Windows failed; all default jobs passed |

## Stop-the-line packet

The acceptance criterion cannot be satisfied on current `main`. Green rc.9
evidence now requires multiple implementation changes; weakening the workflow,
changing the immutable candidate SHA/digest, or treating macOS-only success as a
green matrix would be forced fits.

Recommended route:

1. Review and land the attached focused build-source encoded-path patch.
2. Route the independent Windows Unicode, fixed-environment/toolchain-identity,
   and staged-script-executable failures to focused owners with tests.
3. After all fixes land on `main`, dispatch the exact same candidate SHA and
   manifest digest again.
4. Only a three-OS green candidate matrix may be routed as release-unblocking
   evidence to `TASK-260822-c0rxj7`, `TASK-260822-f4qv7w`, and
   `TASK-260822-1so0ym`.

External input needed to resume: maintainer/reviewer landing of the scoped patch
and ownership/routing for the three independent Windows fix areas.
