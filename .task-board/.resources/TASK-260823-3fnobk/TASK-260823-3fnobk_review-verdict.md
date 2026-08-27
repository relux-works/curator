# TASK-260823-3fnobk reviewer verdict

Verdict: **accepted**.

## Scope and implementation

- Reviewed merge commit `c6092af0f7d01617832a1307832121e8853b11bc` and its single implementation commit `c73bc1339946b31a21cf9f784448ea3acac4fbb8`.
- The landed diff is byte-for-byte identical to `TASK-260823-1l1p8q_buildsource-encoded-path-fix.patch`, SHA-256 `4d62e862132f91d925258a3475375e9ba554d3301e66d4b50f5c2a40ae88752b`.
- `pathSet` now keys only on the exact encoded path. It retains invalid UTF-8 rejection and exact duplicate rejection while removing synthetic case-folded/NFD collision rejection.
- Regression coverage explicitly proves exact duplicates still fail and case- and normalization-distinct encoded paths pass.

## Landing and CI evidence

- GitHub PR #26, `Fix encoded build-source path admission`, is merged to `main` at `c6092af`.
- All executed PR checks passed: Lint; Test on Ubuntu, macOS, and Windows; Race on Ubuntu and macOS; Gate self-test on Ubuntu, macOS, and Windows; Interop conformance gate; Naming gate.
- The candidate-suite matrix was skipped by design on the pull-request event because no candidate input was supplied.

## Independent local validation

- Fresh detached worktree created from `origin/main` at exactly `c6092af`.
- Checked out `relux-works/curator-spec` rc.9 at `859727b103ed175ff214cbb64641f4686d8c6a68`.
- Verified rc.9 `conformance/v1/manifest.json` SHA-256: `782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f`.
- `go test ./internal/buildsource -run '^TestRejectsInvalidPathsLinksAndCollisions$' -count=1 -v`: PASS, including exact duplicate rejection and both distinct-path admission cases.
- `CURATOR_CONFORMANCE_ROOT=<rc.9-root> go test ./internal/buildsource -count=1`: PASS.
- On the host's default case-insensitive filesystem, the candidate subtest correctly skipped because `same` and `SAME` collapse to one member; this result was not used as acceptance evidence.
- Re-ran the exact rc.9 candidate subtest with `TMPDIR` on a temporary case-sensitive HFS+ volume: `TestBuildSourceIdentityVectors/duplicate-build-source-path` PASS and did not skip.
- `git diff --check c6092af^ c6092af --`: PASS.

## Review notes

- No correctness, architecture-fit, test, lint, or landing issues found.
- Two discarded validation invocations are retained in scratch logs: one used a wrong relative candidate root and caused a top-level skip; one used invalid `git diff-tree` revision syntax. Both were rerun correctly and did not contribute to acceptance.

