# TASK-260926-1ddneb results — trust the Relux Bot release signer

## Change
- `maintainers.allowed_signers`: kept the `oparin@me.com` line byte-identical, appended exactly one line:
  `bot@relux.works ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPG7xTX05HL1XaD4XLUk0/TTeqRNHbMj5HdnqNQdDTID`
  File ends with a trailing newline (verified at byte level).
- `GOVERNANCE.md` (release-process paragraph): states the allowlist trusts `oparin@me.com` (maintainer) and `bot@relux.works` (automation signer authorized by the operator on 2026-09-26).
- `RELEASE.md` (stable-1.0.0 checklist): same trusted-principal statement where the tag/commit verification is required.
- `tools/test_allowed_signers.py` (new): pins the two-line file (byte-identical maintainer line, exact bot line, trailing newline, known principals) plus negative parser tests (malformed line, unknown key type, non-email principal rejected).
- `CHANGELOG.md` intentionally NOT edited per the 2026-09-24 curator-repo policy (leaf entries go to the release-prep leaf). Entry text for release prep is below.

## CHANGELOG entry (for release prep)

```md
## Unreleased

### Added

- Trust the Relux Bot release signer: `maintainers.allowed_signers` now also trusts `bot@relux.works` (automation signer authorized by the operator on 2026-09-26) alongside the existing maintainer key. `GOVERNANCE.md` and `RELEASE.md` name both trusted principals.
```

## Verification (real exit codes, `set -o pipefail` shell)
- `git -c gpg.format=ssh -c gpg.ssh.allowedSignersFile=$PWD/maintainers.allowed_signers verify-commit a21905d` (Relux Bot-signed commit) -> exit 0 (`Good signature for bot@relux.works`). Baseline with the old file failed exit 1 (`No principal matched`), proving the new line is what authorizes it.
- Same invocation `verify-tag v1.0.0-rc.9` (oparin@me.com-signed tag) -> exit 0, so the existing maintainer trust still works.
- New test: `python -B -m unittest test_allowed_signers` -> 7 tests OK, exit 0.
- Quick tools tests (`test_allowed_signers`, `test_verify_release_commit`, `test_verify_release_merge_policy` from `tools/`) -> 17 tests OK, exit 0.
- `test_release_gate` + `test_implementation_coverage` -> 74 tests OK, exit 0.
- `test_validate.WriteNofollowVectorTests` + `ReadFailureVectorTests` + `SourceSignersVectorTests` -> 77 tests OK, exit 0.
- `python tools/validate.py` -> `validated 64 schemas and 1169 vector files`, exit 0 (with our changes applied).
- `go test ./tools/...` -> ok, exit 0. `gofmt -l tools` empty, `git diff --check` clean.
- Python suite note: the system python lacks `jsonschema`, so the repo gates were run with an isolated venv in $TMPDIR (requirements-dev.txt installed there); the worktree contains no venv residue. A full `test_validate` run (535 tests) observed 1 failure while sibling runs were concurrently rewriting `conformance/` vectors in this shared story worktree; the three vector suites above were re-run green on the clean tree afterwards. `tools/validate.py` red runs seen mid-task also coincided with dirty vector files and pass on the clean tree.

## Findings
- The Story worktree is shared: `conformance/v1/*.json` and `release/1.0.0-rc.13.json` were dirtied more than once by something other than this task (vector content churned between runs). Unrelated files were reverted; final `git status` is only the four intended paths. Reviewers should expect vector-tree flakiness if other leaves run generators concurrently.
- No `Unreleased` section exists in CHANGELOG.md head; the entry above assumes release prep creates it.
- Skills: none of swiftui / core-data / go-testing-tools / architecture-diagrams apply to this trust-file + docs scope; project-management (board ops) was followed for status/evidence/handoff.

## DoD mapping
- allowed_signers = existing line + bot line: yes, byte-verified.
- verify-commit (bot) and verify-tag (maintainer) exit 0 with candidate file: yes.
- docs consistent: GOVERNANCE.md + RELEASE.md name both principals with the 2026-09-26 authorization.
- CHANGELOG: entry text supplied above for the release-prep leaf instead of a direct edit (2026-09-24 policy).
- tests + build gates: listed above with exit codes; worktree left with only product/test/docs changes.
