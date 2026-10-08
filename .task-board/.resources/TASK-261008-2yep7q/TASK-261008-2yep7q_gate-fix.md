# TASK-261008-2yep7q recovery (run 2): rc.14 release-gate fixture fix

## Why this run exists

CR construction for revision 1 failed validation: `spec-gate.sh`
(`make validate`) exited 1 with 3 red tests, all in
`ProtocolRC14ReleaseGateTests` —
`test_accepts_complete_rc14_artifact_set` (ERROR) plus the two
lifecycle/compiled-fixture rejection tests (FAIL), every one reporting
`ReleaseFailure: 1.0.0-rc.14 metadata does not pin the accepted source
suite separately`. Run 1 had reported these as a known red with no
in-scope fix. The autonomous-recovery directive asked for the failure
to be fixed and the producer completed again.

## Root cause

`ProtocolRC14ReleaseGateTests.setUp` materialized its fixture with
`shutil.copytree` of the live worktree, then ran the rc.14 release gate
against it. The gate requires the live skillfile-sources manifest sha
to equal the frozen `release/1.0.0-rc.14.json` pin — true only while the
tree still IS rc.14. This task is the first post-tag suite change (new
v9 schemas, marker v6, vectors; manifest 138 -> 167 files), so the
fixture could no longer validate as rc.14, and the pin check tripped
before the lifecycle assertions the negative tests target. Every future
post-tag change to the suite would have failed the same way. Precedent
for the fix: the rc.13 gate test restored its corpus from the tag with
`git show`, and BUG-261003-mjzrnv made the freeze tests tag-aware the
same week.

## Fix (test-only; no production, schema, vector, or release change)

- `tools/test_release_gate.py`: `ProtocolRC14ReleaseGateTests` now
  fixtures the published `v1.0.0-rc.14` tree via `git archive` instead
  of copying the live worktree, and restores raw blob bytes for
  `export-subst` paths (`conformance/v1/fixtures/byte-exact/subst.txt`
  — archive expansion would otherwise break the pinned manifest hash).
  Discovery is generic (`ls-tree` + batched `check-attr --stdin -z`;
  one file matched). Missing tag / empty archive / malformed output
  fail the test loudly — never a silent skip.
- `CHANGELOG.md`: one `Unreleased/Changed` line for the fixture fix
  (same shape as the BUG-261003 entry).

Gate logic is untouched and unnarrowed: the same 35 tests, same
mutations, same assertions — now against the true rc.14 corpus. The
30+ negative tests still prove each rejection class on that corpus.
Live-tree protection stays where it belongs: `validate.py`'s
released-record/schema immutability gates, which already require tags
locally and run green on the live tree (this run, exit 0).

## Validation (this run; dev venv)

- `python -B -m unittest
  test_release_gate.ProtocolRC14ReleaseGateTests.test_accepts_complete_rc14_artifact_set
  ...test_rejects_removed_lifecycle_case_with_updated_manifest_hash
  ...test_rejects_stale_compiled_fixture_with_updated_manifest_hash`
  -> exit 0, `Ran 3 tests / OK` (the exact 3 from the CR failure log).
- `python -B -m unittest test_release_gate` -> exit 0,
  `Ran 35 tests in 62.5s / OK` (full file, no regressions).
- `python -B tools/validate.py` -> exit 0,
  `validated 73 schemas and 1294 vector files`.
- `python tools/verify_skillfile_sources_independence.py` -> exit 0,
  gate passed (152/152, 31/31).
- `git diff --check` -> exit 0, clean; `gofmt -l tools` -> clean.
- Accepted from run-1 attached evidence without rerun:
  `test_validate.py` 567/567 (verified this run that
  `test_validate.py` imports neither `test_release_gate` nor
  CHANGELOG, so run 1's tree state is identical for everything it
  reads), `go test ./tools/...`, regenerate idempotence, embedded
  suite command. Full `unittest discover` (677 tests, ~13 min)
  exceeds one bounded shell call; it reruns automatically as the CR
  suite after handoff.

## Supersedes

Section 4's "known red / no in-scope fix / sequence rc.15 prep with
landing" in `TASK-261008-2yep7q_results.md` is withdrawn: the red had
an in-scope test-isolation fix, applied here. No rc.15 prep is needed
to land this task. Work left uncommitted in the story worktree.
