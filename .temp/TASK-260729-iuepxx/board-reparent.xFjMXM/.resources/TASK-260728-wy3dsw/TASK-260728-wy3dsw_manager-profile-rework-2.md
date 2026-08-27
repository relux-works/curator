# TASK-260728-wy3dsw manager-profile rework 2

## Scope

Closed the single signer-boundary finding from independent review cycle 2 in
the pinned isolated worktree:

`/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-wy3dsw/curator-spec-worktree`

HEAD remains
`57c1f56846d221ecc55786bd3c2467ec32f11730`. No file was staged or
committed. This rework changes only `profiles/manager.md`; the accepted
cycle-1 `cli/curator.md` remains byte-identical.

## Signer boundary

Manager profile section 11.8 now requires all of the following:

- first-release `go-repository-v1` performs no manager post-build signing,
  timestamping, or notarization;
- a platform or operator policy requiring local signing fails installation
  closed before publication;
- that transition uses the already-defined stable tuple
  `signer-policy` / `unsupported` / `error` /
  `build_repository_signer_policy_unsupported`;
- the rejection remains in force until a separately versioned and reviewed
  signer profile defines the complete signer boundary; and
- operator or release-pipeline Apple Developer ID signing/notarization and
  Windows Authenticode signing/timestamping remain outside install-time build
  receipts, cache identity, manager publication, and this profile.

The text introduces no package-controlled signer input, does not claim
prebuilt provenance, and preserves Decision 0005 and Protocol Core signer
semantics.

## Validation evidence

Every validation command was run directly without `tee`.

- Focused signer-boundary contract assertion: exit 0. It checked the new
  section 11.8 requirements, the stable diagnostic row, the CLI's normative
  reference to manager-profile section 11.10, Decision 0005, Protocol Core,
  and the expected two manager-profile uses of the diagnostic code.
- `make validate` under the task-local pinned Python environment: exit 0. It
  validated 42 schemas and 400 vector files, ran 15 Python tests, and ran the
  Go tool tests.
- `python3 tools/validate.py`: exit 0; 42 schemas and 400 vector files
  validated, including local Markdown links.
- `python3 -B -m unittest discover -s tools -p 'test_*.py'`: exit 0; 15 tests
  passed.
- `go test ./tools/...`: exit 0.
- `git diff --check`: exit 0.
- `git diff --cached --exit-code`: exit 0; the index is clean.
- `git rev-parse HEAD`: exit 0 and returned the pinned HEAD above.
- Accepted-predecessor `rsync -anic --delete` comparison with both worktree
  `.git` file and directory exclusions: exit 0 and reported only
  `profiles/manager.md` and `cli/curator.md`.
- The first scope-comparison invocation also exited 0 but used only the
  directory-form `.git/` exclusion, so it additionally reported the worktree
  `.git` pointer file. The corrected invocation excluded both forms and
  produced the intended task-scope evidence.

Document SHA-256 values:

- `profiles/manager.md`:
  `25fe0c397f7cc4770279a2fe465ba7f60c3b3cd995644464b39c35873084b671`
- `cli/curator.md`:
  `6160f08ad4b433aacb772085949c8bb1f3361eddd6ca5cc52179bd5dbb7c1ba6`

## Handoff

The bounded documentation rework preserves every accepted cycle-1 fix and is
ready for a fresh independent review. This outcome does not self-accept the
task.
