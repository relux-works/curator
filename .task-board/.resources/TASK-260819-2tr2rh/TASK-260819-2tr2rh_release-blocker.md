# TASK-260819-2tr2rh release publication evidence and blocker

## Published repository state

- Selected version: `1.0.0-rc.7`
- Accepted candidate: `993429eaf91d4950197eb0693bb2c416768da440`
- Pull request: https://github.com/relux-works/curator-spec/pull/18
- Protected squash merge commit: `99f70947d6f2447366d6c996127b73eca37a9159`
- Merge commit GitHub verification: valid; GitHub-created commit on `origin/main`
- Signed annotated tag: `v1.0.0-rc.7`
- Tag object: `de704f2951e683d52ae8e475cb690b918a94d4c5`
- Tag target: `99f70947d6f2447366d6c996127b73eca37a9159`
- Tag SSH signature verification: valid for `oparin@me.com`, key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
- Conformance manifest SHA-256: `7faa3cf95cb037e0afb5e2209895e3e89d993a7a706aa0e8770c1dad869e2c76`
- Rc.5 metadata SHA-256: `75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`
- Rc.6 metadata SHA-256: `c4ad58e76687bd563679773a60c6ce35c238d4117b7cbceb05d4f88b5300ed3f`
- Rc.7 metadata SHA-256: `e5872ee4dd207bf6b190d8c8be15a9366d9c1e3638047ea983620b97c9f84d5d`
- `verified_implementations`: `[]`
- `verified_platform_claims`: `[]`
- GitHub release: absent; packaging and publication did not run.

## Gate evidence

- First `python3 tools/validate.py`: exit 1, environment prerequisite failure (`jsonschema` absent).
- Isolated requirements installation: exit 0.
- `.venv/bin/python tools/validate.py`: exit 0; 49 schemas and 471 vector files validated.
- `.venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'`: exit 0; 84 tests passed.
- `go test ./tools/...`: exit 0.
- First regeneration and scoped diff: exit 0 / exit 0.
- Second regeneration and scoped diff: exit 0 / exit 0.
- First clean-candidate release-gate attempt: exit 1 because validation created untracked `tools/__pycache__`.
- Candidate release gate after removing the runtime cache and preventing bytecode: exit 0 at `993429eaf91d4950197eb0693bb2c416768da440`.
- `git verify-commit HEAD`: exit 0.
- Isolated exact-name signed-tag preflight: initial creation exit 128 because the isolated clone lacked `user.signingkey`; after configuring the existing maintainer key, tag creation and `git verify-tag` both exited 0.
- PR #18: all eight required checks passed across Linux, macOS, and Windows.
- Merged-commit release gate: exit 0 at `99f70947d6f2447366d6c996127b73eca37a9159`.
- Merged-commit provenance verifier: exit 0.
- Fresh `main` Specification CI run 32195700148: success.
- Fresh `main` Implementation conformance run 32195700170: success.
- Local and remote tag absence preflights returned non-zero as expected before tag creation (local 128, remote 2).
- Real annotated tag creation, verification, target check, and push: exit 0.
- Release workflow run 32195911143: exit 1 at `Validate release input`.

## Stop-the-line constraint

The release workflow runs `python tools/validate.py` before `python tools/release_gate.py`. Importing `tools/assurance.py` creates untracked `tools/__pycache__/assurance.cpython-312.pyc`. The release gate intentionally rejects any dirty candidate checkout, so the workflow fails after all schema, unit, Go, regeneration, signature, and provenance checks pass. Packaging, attestation, and GitHub prerelease publication are skipped.

The signed `v1.0.0-rc.7` tag has already been pushed. Governance states that release tags are immutable. Fixing `.github/workflows/release.yml` or ignoring Python bytecode requires a new reviewed commit, so it cannot be applied beneath the existing rc.7 tag without moving immutable release identity.

## Attempts and rejected workaround

- Confirmed the failure locally and in Actions.
- Confirmed the failure is only the untracked Python bytecode runtime artifact.
- Did not weaken the clean-checkout release gate.
- Did not delete, force-update, or repoint the signed remote tag.
- Did not manually fabricate release archives or bypass the reviewed workflow.

## Viable options

1. Recommended: retain rc.7 as an immutable failed publication tag, prepare a reviewed workflow fix, advance all release metadata and manifest identity to `1.0.0-rc.8`, and publish rc.8 through the corrected workflow. This preserves auditability and SemVer release immutability but requires downstream identity to become rc.8.
2. Exceptional: explicitly authorize deleting/replacing the remote rc.7 tag after a reviewed workflow fix. This satisfies the requested label but violates the repository's immutable-tag governance and is not recommended.
3. Leave rc.7 unpublished. This preserves policy but does not satisfy the task acceptance criteria.

## Required decision

Authorize option 1 (supersede with `v1.0.0-rc.8`) or explicitly approve the governance exception in option 2. No further safe autonomous publication step exists under the current rc.7 identity.
